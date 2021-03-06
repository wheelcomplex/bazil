// Code generated by protoc-gen-go.
// source: bazil.org/bazil/server/control/wire/control.proto
// DO NOT EDIT!

/*
Package wire is a generated protocol buffer package.

It is generated from these files:
	bazil.org/bazil/server/control/wire/control.proto
	bazil.org/bazil/server/control/wire/peer.proto
	bazil.org/bazil/server/control/wire/sharing.proto
	bazil.org/bazil/server/control/wire/volume.proto

It has these top-level messages:
	PingRequest
	PingResponse
*/
package wire

import proto "github.com/golang/protobuf/proto"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal

type PingRequest struct {
}

func (m *PingRequest) Reset()         { *m = PingRequest{} }
func (m *PingRequest) String() string { return proto.CompactTextString(m) }
func (*PingRequest) ProtoMessage()    {}

type PingResponse struct {
}

func (m *PingResponse) Reset()         { *m = PingResponse{} }
func (m *PingResponse) String() string { return proto.CompactTextString(m) }
func (*PingResponse) ProtoMessage()    {}

func init() {
}

// Client API for Control service

type ControlClient interface {
	Ping(ctx context.Context, in *PingRequest, opts ...grpc.CallOption) (*PingResponse, error)
	VolumeCreate(ctx context.Context, in *VolumeCreateRequest, opts ...grpc.CallOption) (*VolumeCreateResponse, error)
	VolumeMount(ctx context.Context, in *VolumeMountRequest, opts ...grpc.CallOption) (*VolumeMountResponse, error)
	VolumeStorageAdd(ctx context.Context, in *VolumeStorageAddRequest, opts ...grpc.CallOption) (*VolumeStorageAddResponse, error)
	SharingKeyAdd(ctx context.Context, in *SharingKeyAddRequest, opts ...grpc.CallOption) (*SharingKeyAddResponse, error)
	PeerAdd(ctx context.Context, in *PeerAddRequest, opts ...grpc.CallOption) (*PeerAddResponse, error)
	PeerLocationSet(ctx context.Context, in *PeerLocationSetRequest, opts ...grpc.CallOption) (*PeerLocationSetResponse, error)
	PeerStorageAllow(ctx context.Context, in *PeerStorageAllowRequest, opts ...grpc.CallOption) (*PeerStorageAllowResponse, error)
}

type controlClient struct {
	cc *grpc.ClientConn
}

func NewControlClient(cc *grpc.ClientConn) ControlClient {
	return &controlClient{cc}
}

func (c *controlClient) Ping(ctx context.Context, in *PingRequest, opts ...grpc.CallOption) (*PingResponse, error) {
	out := new(PingResponse)
	err := grpc.Invoke(ctx, "/bazil.control.Control/Ping", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *controlClient) VolumeCreate(ctx context.Context, in *VolumeCreateRequest, opts ...grpc.CallOption) (*VolumeCreateResponse, error) {
	out := new(VolumeCreateResponse)
	err := grpc.Invoke(ctx, "/bazil.control.Control/VolumeCreate", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *controlClient) VolumeMount(ctx context.Context, in *VolumeMountRequest, opts ...grpc.CallOption) (*VolumeMountResponse, error) {
	out := new(VolumeMountResponse)
	err := grpc.Invoke(ctx, "/bazil.control.Control/VolumeMount", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *controlClient) VolumeStorageAdd(ctx context.Context, in *VolumeStorageAddRequest, opts ...grpc.CallOption) (*VolumeStorageAddResponse, error) {
	out := new(VolumeStorageAddResponse)
	err := grpc.Invoke(ctx, "/bazil.control.Control/VolumeStorageAdd", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *controlClient) SharingKeyAdd(ctx context.Context, in *SharingKeyAddRequest, opts ...grpc.CallOption) (*SharingKeyAddResponse, error) {
	out := new(SharingKeyAddResponse)
	err := grpc.Invoke(ctx, "/bazil.control.Control/SharingKeyAdd", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *controlClient) PeerAdd(ctx context.Context, in *PeerAddRequest, opts ...grpc.CallOption) (*PeerAddResponse, error) {
	out := new(PeerAddResponse)
	err := grpc.Invoke(ctx, "/bazil.control.Control/PeerAdd", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *controlClient) PeerLocationSet(ctx context.Context, in *PeerLocationSetRequest, opts ...grpc.CallOption) (*PeerLocationSetResponse, error) {
	out := new(PeerLocationSetResponse)
	err := grpc.Invoke(ctx, "/bazil.control.Control/PeerLocationSet", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *controlClient) PeerStorageAllow(ctx context.Context, in *PeerStorageAllowRequest, opts ...grpc.CallOption) (*PeerStorageAllowResponse, error) {
	out := new(PeerStorageAllowResponse)
	err := grpc.Invoke(ctx, "/bazil.control.Control/PeerStorageAllow", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Control service

type ControlServer interface {
	Ping(context.Context, *PingRequest) (*PingResponse, error)
	VolumeCreate(context.Context, *VolumeCreateRequest) (*VolumeCreateResponse, error)
	VolumeMount(context.Context, *VolumeMountRequest) (*VolumeMountResponse, error)
	VolumeStorageAdd(context.Context, *VolumeStorageAddRequest) (*VolumeStorageAddResponse, error)
	SharingKeyAdd(context.Context, *SharingKeyAddRequest) (*SharingKeyAddResponse, error)
	PeerAdd(context.Context, *PeerAddRequest) (*PeerAddResponse, error)
	PeerLocationSet(context.Context, *PeerLocationSetRequest) (*PeerLocationSetResponse, error)
	PeerStorageAllow(context.Context, *PeerStorageAllowRequest) (*PeerStorageAllowResponse, error)
}

func RegisterControlServer(s *grpc.Server, srv ControlServer) {
	s.RegisterService(&_Control_serviceDesc, srv)
}

func _Control_Ping_Handler(srv interface{}, ctx context.Context, codec grpc.Codec, buf []byte) (interface{}, error) {
	in := new(PingRequest)
	if err := codec.Unmarshal(buf, in); err != nil {
		return nil, err
	}
	out, err := srv.(ControlServer).Ping(ctx, in)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func _Control_VolumeCreate_Handler(srv interface{}, ctx context.Context, codec grpc.Codec, buf []byte) (interface{}, error) {
	in := new(VolumeCreateRequest)
	if err := codec.Unmarshal(buf, in); err != nil {
		return nil, err
	}
	out, err := srv.(ControlServer).VolumeCreate(ctx, in)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func _Control_VolumeMount_Handler(srv interface{}, ctx context.Context, codec grpc.Codec, buf []byte) (interface{}, error) {
	in := new(VolumeMountRequest)
	if err := codec.Unmarshal(buf, in); err != nil {
		return nil, err
	}
	out, err := srv.(ControlServer).VolumeMount(ctx, in)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func _Control_VolumeStorageAdd_Handler(srv interface{}, ctx context.Context, codec grpc.Codec, buf []byte) (interface{}, error) {
	in := new(VolumeStorageAddRequest)
	if err := codec.Unmarshal(buf, in); err != nil {
		return nil, err
	}
	out, err := srv.(ControlServer).VolumeStorageAdd(ctx, in)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func _Control_SharingKeyAdd_Handler(srv interface{}, ctx context.Context, codec grpc.Codec, buf []byte) (interface{}, error) {
	in := new(SharingKeyAddRequest)
	if err := codec.Unmarshal(buf, in); err != nil {
		return nil, err
	}
	out, err := srv.(ControlServer).SharingKeyAdd(ctx, in)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func _Control_PeerAdd_Handler(srv interface{}, ctx context.Context, codec grpc.Codec, buf []byte) (interface{}, error) {
	in := new(PeerAddRequest)
	if err := codec.Unmarshal(buf, in); err != nil {
		return nil, err
	}
	out, err := srv.(ControlServer).PeerAdd(ctx, in)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func _Control_PeerLocationSet_Handler(srv interface{}, ctx context.Context, codec grpc.Codec, buf []byte) (interface{}, error) {
	in := new(PeerLocationSetRequest)
	if err := codec.Unmarshal(buf, in); err != nil {
		return nil, err
	}
	out, err := srv.(ControlServer).PeerLocationSet(ctx, in)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func _Control_PeerStorageAllow_Handler(srv interface{}, ctx context.Context, codec grpc.Codec, buf []byte) (interface{}, error) {
	in := new(PeerStorageAllowRequest)
	if err := codec.Unmarshal(buf, in); err != nil {
		return nil, err
	}
	out, err := srv.(ControlServer).PeerStorageAllow(ctx, in)
	if err != nil {
		return nil, err
	}
	return out, nil
}

var _Control_serviceDesc = grpc.ServiceDesc{
	ServiceName: "bazil.control.Control",
	HandlerType: (*ControlServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Ping",
			Handler:    _Control_Ping_Handler,
		},
		{
			MethodName: "VolumeCreate",
			Handler:    _Control_VolumeCreate_Handler,
		},
		{
			MethodName: "VolumeMount",
			Handler:    _Control_VolumeMount_Handler,
		},
		{
			MethodName: "VolumeStorageAdd",
			Handler:    _Control_VolumeStorageAdd_Handler,
		},
		{
			MethodName: "SharingKeyAdd",
			Handler:    _Control_SharingKeyAdd_Handler,
		},
		{
			MethodName: "PeerAdd",
			Handler:    _Control_PeerAdd_Handler,
		},
		{
			MethodName: "PeerLocationSet",
			Handler:    _Control_PeerLocationSet_Handler,
		},
		{
			MethodName: "PeerStorageAllow",
			Handler:    _Control_PeerStorageAllow_Handler,
		},
	},
	Streams: []grpc.StreamDesc{},
}
