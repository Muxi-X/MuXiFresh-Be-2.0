// Code generated by goctl. DO NOT EDIT.
// Source: user.proto

package userclient

import (
	"context"

	"MuXiFresh-Be-2.0/app/user/cmd/rpc/user/pb"

	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

type (
	GetAdminListReq  = pb.GetAdminListReq
	GetAdminListResp = pb.GetAdminListResp
	GetUserInfoReq   = pb.GetUserInfoReq
	GetUserInfoResp  = pb.GetUserInfoResp
	GetUserTypeReq   = pb.GetUserTypeReq
	GetUserTypeResp  = pb.GetUserTypeResp
	One              = pb.One
	SetUserInfoReq   = pb.SetUserInfoReq
	SetUserInfoResp  = pb.SetUserInfoResp
	SetUserTypeReq   = pb.SetUserTypeReq
	SetUserTypeResp  = pb.SetUserTypeResp

	UserClient interface {
		GetUserInfo(ctx context.Context, in *GetUserInfoReq, opts ...grpc.CallOption) (*GetUserInfoResp, error)
		SetUserInfo(ctx context.Context, in *SetUserInfoReq, opts ...grpc.CallOption) (*SetUserInfoResp, error)
		SetUserType(ctx context.Context, in *SetUserTypeReq, opts ...grpc.CallOption) (*SetUserTypeResp, error)
		GetAdminList(ctx context.Context, in *GetAdminListReq, opts ...grpc.CallOption) (*GetAdminListResp, error)
		GetUserType(ctx context.Context, in *GetUserTypeReq, opts ...grpc.CallOption) (*GetUserTypeResp, error)
	}

	defaultUserClient struct {
		cli zrpc.Client
	}
)

func NewUserClient(cli zrpc.Client) UserClient {
	return &defaultUserClient{
		cli: cli,
	}
}

func (m *defaultUserClient) GetUserInfo(ctx context.Context, in *GetUserInfoReq, opts ...grpc.CallOption) (*GetUserInfoResp, error) {
	client := pb.NewUserClientClient(m.cli.Conn())
	return client.GetUserInfo(ctx, in, opts...)
}

func (m *defaultUserClient) SetUserInfo(ctx context.Context, in *SetUserInfoReq, opts ...grpc.CallOption) (*SetUserInfoResp, error) {
	client := pb.NewUserClientClient(m.cli.Conn())
	return client.SetUserInfo(ctx, in, opts...)
}

func (m *defaultUserClient) SetUserType(ctx context.Context, in *SetUserTypeReq, opts ...grpc.CallOption) (*SetUserTypeResp, error) {
	client := pb.NewUserClientClient(m.cli.Conn())
	return client.SetUserType(ctx, in, opts...)
}

func (m *defaultUserClient) GetAdminList(ctx context.Context, in *GetAdminListReq, opts ...grpc.CallOption) (*GetAdminListResp, error) {
	client := pb.NewUserClientClient(m.cli.Conn())
	return client.GetAdminList(ctx, in, opts...)
}

func (m *defaultUserClient) GetUserType(ctx context.Context, in *GetUserTypeReq, opts ...grpc.CallOption) (*GetUserTypeResp, error) {
	client := pb.NewUserClientClient(m.cli.Conn())
	return client.GetUserType(ctx, in, opts...)
}