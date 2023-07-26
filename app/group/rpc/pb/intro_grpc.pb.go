// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v4.22.2
// source: intro.proto

package pb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// IntroClientClient is the client API for IntroClient service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type IntroClientClient interface {
	GetGroupIntro(ctx context.Context, in *GroupIntroReq, opts ...grpc.CallOption) (*GroupIntroResp, error)
	GetRecruitInfo(ctx context.Context, in *RecruitInfoReq, opts ...grpc.CallOption) (*RecruitInfoResp, error)
}

type introClientClient struct {
	cc grpc.ClientConnInterface
}

func NewIntroClientClient(cc grpc.ClientConnInterface) IntroClientClient {
	return &introClientClient{cc}
}

func (c *introClientClient) GetGroupIntro(ctx context.Context, in *GroupIntroReq, opts ...grpc.CallOption) (*GroupIntroResp, error) {
	out := new(GroupIntroResp)
	err := c.cc.Invoke(ctx, "/intro.IntroClient/GetGroupIntro", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *introClientClient) GetRecruitInfo(ctx context.Context, in *RecruitInfoReq, opts ...grpc.CallOption) (*RecruitInfoResp, error) {
	out := new(RecruitInfoResp)
	err := c.cc.Invoke(ctx, "/intro.IntroClient/GetRecruitInfo", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// IntroClientServer is the server API for IntroClient service.
// All implementations must embed UnimplementedIntroClientServer
// for forward compatibility
type IntroClientServer interface {
	GetGroupIntro(context.Context, *GroupIntroReq) (*GroupIntroResp, error)
	GetRecruitInfo(context.Context, *RecruitInfoReq) (*RecruitInfoResp, error)
	mustEmbedUnimplementedIntroClientServer()
}

// UnimplementedIntroClientServer must be embedded to have forward compatible implementations.
type UnimplementedIntroClientServer struct {
}

func (UnimplementedIntroClientServer) GetGroupIntro(context.Context, *GroupIntroReq) (*GroupIntroResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetGroupIntro not implemented")
}
func (UnimplementedIntroClientServer) GetRecruitInfo(context.Context, *RecruitInfoReq) (*RecruitInfoResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetRecruitInfo not implemented")
}
func (UnimplementedIntroClientServer) mustEmbedUnimplementedIntroClientServer() {}

// UnsafeIntroClientServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to IntroClientServer will
// result in compilation errors.
type UnsafeIntroClientServer interface {
	mustEmbedUnimplementedIntroClientServer()
}

func RegisterIntroClientServer(s grpc.ServiceRegistrar, srv IntroClientServer) {
	s.RegisterService(&IntroClient_ServiceDesc, srv)
}

func _IntroClient_GetGroupIntro_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GroupIntroReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IntroClientServer).GetGroupIntro(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/intro.IntroClient/GetGroupIntro",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IntroClientServer).GetGroupIntro(ctx, req.(*GroupIntroReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _IntroClient_GetRecruitInfo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RecruitInfoReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IntroClientServer).GetRecruitInfo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/intro.IntroClient/GetRecruitInfo",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IntroClientServer).GetRecruitInfo(ctx, req.(*RecruitInfoReq))
	}
	return interceptor(ctx, in, info, handler)
}

// IntroClient_ServiceDesc is the grpc.ServiceDesc for IntroClient service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var IntroClient_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "intro.IntroClient",
	HandlerType: (*IntroClientServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetGroupIntro",
			Handler:    _IntroClient_GetGroupIntro_Handler,
		},
		{
			MethodName: "GetRecruitInfo",
			Handler:    _IntroClient_GetRecruitInfo_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "intro.proto",
}
