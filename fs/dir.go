package fs

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"syscall"

	"bazil.org/bazil/cas/blobs"
	wirecas "bazil.org/bazil/cas/wire"
	"bazil.org/bazil/db"
	"bazil.org/bazil/fs/inodes"
	"bazil.org/bazil/fs/snap"
	wiresnap "bazil.org/bazil/fs/snap/wire"
	"bazil.org/bazil/fs/wire"
	"bazil.org/bazil/util/env"
	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/golang/protobuf/proto"
	"golang.org/x/net/context"
)

type dir struct {
	fs.NodeRef

	inode  uint64
	parent *dir
	fs     *Volume

	// mu protects the fields below.
	//
	// If multiple dir.mu instances need to be locked at the same
	// time, the locks must be taken in topologically sorted
	// order, parent first.
	//
	// As there can be only one db.Update at a time, those calls
	// must be considered as lock operations too. To avoid lock
	// ordering related deadlocks, never hold mu while calling
	// db.Update.
	mu sync.Mutex

	name string

	// each in-memory child, so we can return the same node on
	// multiple Lookups and know what to do on .save()
	//
	// each child also stores its own name; if the value in the child
	// is an empty string, that means the child has been unlinked
	active map[string]node
}

func newDir(filesys *Volume, inode uint64, parent *dir, name string) *dir {
	d := &dir{
		inode:  inode,
		name:   name,
		parent: parent,
		fs:     filesys,
		active: make(map[string]node),
	}
	return d
}

var _ = node(&dir{})
var _ = fs.Node(&dir{})
var _ = fs.NodeCreater(&dir{})
var _ = fs.NodeForgetter(&dir{})
var _ = fs.NodeMkdirer(&dir{})
var _ = fs.NodeRemover(&dir{})
var _ = fs.NodeRenamer(&dir{})
var _ = fs.NodeStringLookuper(&dir{})
var _ = fs.HandleReadDirAller(&dir{})

func (d *dir) setName(name string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.name = name
}

func (d *dir) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = d.inode
	a.Mode = os.ModeDir | 0755
	a.Uid = env.MyUID
	a.Gid = env.MyGID
	return nil
}

func (d *dir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	if d.inode == 1 && name == ".snap" {
		return &listSnaps{
			fs:      d.fs,
			rootDir: d,
		}, nil
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	if child, ok := d.active[name]; ok {
		return child, nil
	}

	var de *wire.Dirent
	key := pathToKey(d.inode, name)
	lookup := func(tx *db.Tx) error {
		bucket := d.fs.bucket(tx)
		buf := bucket.DirBucket().Get(key)
		if buf == nil {
			return fuse.ENOENT
		}
		var err error
		de, err = unmarshalDirent(buf)
		if err != nil {
			return fmt.Errorf("dirent unmarshal problem: %v", err)
		}
		return nil
	}
	if err := d.fs.db.View(lookup); err != nil {
		return nil, err
	}
	child, err := d.reviveNode(de, name)
	if err != nil {
		return nil, fmt.Errorf("dirent node unmarshal problem: %v", err)
	}
	d.active[name] = child
	return child, nil
}

func unmarshalDirent(buf []byte) (*wire.Dirent, error) {
	var de wire.Dirent
	if err := proto.Unmarshal(buf, &de); err != nil {
		return nil, err
	}
	return &de, nil
}

func (d *dir) reviveDir(de *wire.Dirent, name string) (*dir, error) {
	if de.Dir == nil {
		return nil, fmt.Errorf("tried to revive non-directory as directory: %v", de)
	}
	child := newDir(d.fs, de.Inode, d, name)
	return child, nil
}

func (d *dir) reviveNode(de *wire.Dirent, name string) (node, error) {
	switch {
	case de.Dir != nil:
		return d.reviveDir(de, name)

	case de.File != nil:
		manifest, err := de.File.Manifest.ToBlob("file")
		if err != nil {
			return nil, err
		}
		blob, err := blobs.Open(d.fs.chunkStore, manifest)
		if err != nil {
			return nil, err
		}
		child := &file{
			inode:  de.Inode,
			name:   name,
			parent: d,
			blob:   blob,
		}
		return child, nil
	}

	return nil, fmt.Errorf("dirent unknown type: %v", de)
}

func (d *dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	var entries []fuse.Dirent
	readDirAll := func(tx *db.Tx) error {
		bucket := d.fs.bucket(tx)
		c := bucket.DirBucket().Cursor()
		prefix := pathToKey(d.inode, "")
		for k, v := c.Seek(prefix); k != nil; k, v = c.Next() {
			if !bytes.HasPrefix(k, prefix) {
				// past the end of the directory
				break
			}

			name := string(k[len(prefix):])
			de, err := unmarshalDirent(v)
			if err != nil {
				return fmt.Errorf("readdir error: %v", err)
			}
			fde := de.GetFUSEDirent(name)
			entries = append(entries, fde)
		}
		return nil
	}
	err := d.fs.db.View(readDirAll)
	return entries, err
}

// saveInternal persists entry name in dir to the database.
//
// uses no mutable state of d, and hence does not need to lock d.mu.
func (d *dir) saveInternal(tx *db.Tx, name string, n node) error {
	de, err := n.marshal()
	if err != nil {
		return fmt.Errorf("node save error: %v", err)
	}

	buf, err := proto.Marshal(de)
	if err != nil {
		return fmt.Errorf("Dirent marshal error: %v", err)
	}

	key := pathToKey(d.inode, name)
	bucket := d.fs.bucket(tx)
	if err := bucket.DirBucket().Put(key, buf); err != nil {
		return fmt.Errorf("db write error: %v", err)
	}
	return nil
}

func (d *dir) marshal() (*wire.Dirent, error) {
	de := &wire.Dirent{
		Inode: d.inode,
	}
	de.Dir = &wire.Dir{}
	return de, nil
}

func (d *dir) save(tx *db.Tx, name string, n node) error {
	if name == "" {
		// unlinked
		return nil
	}

	d.mu.Lock()
	defer d.mu.Unlock()
	return d.saveInternal(tx, name, n)
}

const debugCreateExisting = true

func (d *dir) Create(ctx context.Context, req *fuse.CreateRequest, resp *fuse.CreateResponse) (fs.Node, fs.Handle, error) {
	// TODO check for duplicate name

	switch req.Mode & os.ModeType {
	case 0:
		var child node
		createFile := func(tx *db.Tx) error {
			bucket := d.fs.bucket(tx).InodeBucket()
			if bucket == nil {
				return errors.New("inode bucket is missing")
			}
			inode, err := inodes.Allocate(bucket)
			if err != nil {
				return err
			}

			manifest := blobs.EmptyManifest("file")
			blob, err := blobs.Open(d.fs.chunkStore, manifest)
			if err != nil {
				return fmt.Errorf("blob open problem: %v", err)
			}
			child = &file{
				inode:  inode,
				name:   req.Name,
				parent: d,
				blob:   blob,
			}
			if err := d.saveInternal(tx, req.Name, child); err != nil {
				return err
			}
			return nil
		}
		if err := d.fs.db.Update(createFile); err != nil {
			return nil, nil, err
		}

		d.mu.Lock()
		defer d.mu.Unlock()
		if debugCreateExisting {
			if n, ok := d.active[req.Name]; ok {
				log.Printf("asked to create with existing node: %q %#v", req.Name, n)
				n.setName("")
			}
		}
		d.active[req.Name] = child
		return child, child, nil
	default:
		return nil, nil, fuse.EPERM
	}
}

const debugActiveChildren = true

func (d *dir) forgetChild(name string, child node) {
	if name == "" {
		// unlinked
		return
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	if debugActiveChildren {
		have, ok := d.active[name]
		switch {
		case !ok:
			log.Printf("asked to forget non-active child: %q %#v", name, child)
		case have != child:
			log.Printf("asked to forget wrong child: %q %#v", name, child)
		}
	}
	delete(d.active, name)
}

func (d *dir) Forget() {
	if d.parent == nil {
		// root dir, don't keep track
		return
	}

	d.mu.Lock()
	name := d.name
	d.mu.Unlock()

	d.parent.forgetChild(name, d)
}

const debugMkdirExisting = true

func (d *dir) Mkdir(ctx context.Context, req *fuse.MkdirRequest) (fs.Node, error) {
	// TODO handle req.Mode

	var child node
	mkdir := func(tx *db.Tx) error {
		bucket := d.fs.bucket(tx).InodeBucket()
		if bucket == nil {
			return errors.New("inode bucket is missing")
		}
		inode, err := inodes.Allocate(bucket)
		if err != nil {
			return err
		}
		child = newDir(d.fs, inode, d, req.Name)
		if err := d.saveInternal(tx, req.Name, child); err != nil {
			return err
		}
		return nil
	}
	if err := d.fs.db.Update(mkdir); err != nil {
		if err == inodes.ErrOutOfInodes {
			return nil, fuse.Errno(syscall.ENOSPC)
		}
		return nil, err
	}

	d.mu.Lock()
	defer d.mu.Unlock()
	if debugMkdirExisting {
		if n, ok := d.active[req.Name]; ok {
			log.Printf("asked to mkdir with existing node: %q %#v", req.Name, n)
			n.setName("")
		}
	}
	d.active[req.Name] = child
	return child, nil
}

func (d *dir) Remove(ctx context.Context, req *fuse.RemoveRequest) error {
	key := pathToKey(d.inode, req.Name)
	remove := func(tx *db.Tx) error {
		bucket := d.fs.bucket(tx)
		dirBucket := bucket.DirBucket()

		if dirBucket.Get(key) == nil {
			return fuse.ENOENT
		}
		err := dirBucket.Delete(key)
		if err != nil {
			return err
		}

		// TODO free inode
		return nil
	}
	if err := d.fs.db.Update(remove); err != nil {
		return err
	}

	d.mu.Lock()
	defer d.mu.Unlock()
	if node, ok := d.active[req.Name]; ok {
		delete(d.active, req.Name)
		node.setName("")
	}
	return nil
}

func (d *dir) Rename(ctx context.Context, req *fuse.RenameRequest, newDir fs.Node) error {
	// if you ever change this, also guard against renaming into
	// special directories like .snap; check type of newDir is *dir
	//
	// also worry about deadlocks
	if newDir != d {
		return fuse.Errno(syscall.EXDEV)
	}

	kOld := pathToKey(d.inode, req.OldName)
	kNew := pathToKey(d.inode, req.NewName)

	rename := func(tx *db.Tx) error {
		bucket := d.fs.bucket(tx)
		dirBucket := bucket.DirBucket()

		// TODO don't need to load from db if req.OldName is in active.
		// instead, save active state if we have it; call .save() not this
		// kludge
		bufOld := dirBucket.Get(kOld)
		if bufOld == nil {
			return fuse.ENOENT
		}

		// the file getting overwritten
		var loserInode uint64
		{
			// TODO don't need to load from db if req.NewName is in active
			bufLoser := dirBucket.Get(kNew)
			if bufLoser != nil {
				// overwriting
				deLoser, err := unmarshalDirent(bufLoser)
				if err != nil {
					return fmt.Errorf("dirent unmarshal problem: %v", err)
				}
				loserInode = deLoser.Inode
			}
		}

		if err := dirBucket.Put(kNew, bufOld); err != nil {
			return err
		}
		if err := dirBucket.Delete(kOld); err != nil {
			return err
		}

		if loserInode > 0 {
			// TODO free loser inode
		}

		return nil
	}
	if err := d.fs.db.Update(rename); err != nil {
		return err
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	// tell overwritten node it's unlinked
	if n, ok := d.active[req.NewName]; ok {
		n.setName("")
	}

	// if the source inode is active, record its new name
	if nodeOld, ok := d.active[req.OldName]; ok {
		nodeOld.setName(req.NewName)
		delete(d.active, req.OldName)
		d.active[req.NewName] = nodeOld
	}

	return nil
}

// snapshot records a snapshot of the directory and stores it in wde
func (d *dir) snapshot(ctx context.Context, tx *db.Tx) (*wiresnap.Dir, error) {
	// NOT HOLDING THE LOCK, accessing database snapshot ONLY

	// TODO move bucket lookup to caller?
	bucket := d.fs.bucket(tx)

	manifest := blobs.EmptyManifest("dir")
	blob, err := blobs.Open(d.fs.chunkStore, manifest)
	if err != nil {
		return nil, err
	}
	w := snap.NewWriter(blob)

	c := bucket.DirBucket().Cursor()
	prefix := pathToKey(d.inode, "")
	for k, v := c.Seek(prefix); k != nil; k, v = c.Next() {
		if !bytes.HasPrefix(k, prefix) {
			// past the end of the directory
			break
		}

		name := string(k[len(prefix):])
		de, err := unmarshalDirent(v)
		if err != nil {
			return nil, err
		}
		sde := wiresnap.Dirent{
			Name: name,
		}
		switch {
		case de.File != nil:
			// TODO d.reviveNode would do blobs.Open and that's a bit
			// too much work; rework the apis
			sde.File = &wiresnap.File{
				Manifest: de.File.Manifest,
			}
		case de.Dir != nil:
			child, err := d.reviveDir(de, name)
			if err != nil {
				return nil, err
			}
			sde.Dir, err = child.snapshot(ctx, tx)
			if err != nil {
				return nil, err
			}
		default:
			return nil, errors.New("TODO")
		}
		err = w.Add(&sde)
		if err != nil {
			return nil, err
		}
	}

	manifest, err = blob.Save()
	if err != nil {
		return nil, err
	}
	msg := wiresnap.Dir{
		Manifest: wirecas.FromBlob(manifest),
		Align:    w.Align(),
	}
	return &msg, nil
}
