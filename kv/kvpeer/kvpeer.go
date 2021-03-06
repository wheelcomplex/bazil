package kvpeer

import (
	"io"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"bazil.org/bazil/kv"
	"bazil.org/bazil/peer/wire"
)

type KVPeer struct {
	peer wire.PeerClient
}

var _ = kv.KV(&KVPeer{})

func (k *KVPeer) Put(key, value []byte) error {
	ctx := context.TODO()
	stream, err := k.peer.ObjectPut(ctx)
	if err != nil {
		return err
	}

	first := true

	const chunkSize = 4 * 1024 * 1024
	var chunk []byte
	buf := value
	for len(buf) > 0 {
		size := chunkSize
		if size > len(buf) {
			size = len(buf)
		}
		chunk, buf = buf[:size], buf[size:]

		req := &wire.ObjectPutRequest{Data: chunk}
		if first {
			req.Key = key
			first = false
		}
		if err := stream.Send(req); err != nil {
			return err
		}
	}

	if _, err := stream.CloseAndRecv(); err != nil {
		return err
	}
	return nil
}

func (k *KVPeer) Get(key []byte) ([]byte, error) {
	ctx := context.TODO()
	stream, err := k.peer.ObjectGet(ctx, &wire.ObjectGetRequest{
		Key: key,
	})
	if err != nil {
		if grpc.Code(err) == codes.NotFound {
			return nil, kv.NotFoundError{Key: key}
		}
		return nil, err
	}

	var data []byte
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		data = append(data, resp.Data...)
	}
	return data, nil
}

func Open(peer wire.PeerClient) (*KVPeer, error) {
	return &KVPeer{
		peer: peer,
	}, nil
}
