package rpcauth

import (
	"context"
	"crypto/subtle"
	"errors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// MetadataKey 是服务间调用携带共享密钥的 metadata 键。
// 跨服务协议键，只在此定义一处，防止 client/server 两侧拼写漂移。
const MetadataKey = "internal-auth"

// ErrMissingToken 配置错误：server 未配置密钥时拒绝启动，防止静默裸奔（fail closed）。
var ErrMissingToken = errors.New("rpcauth: token must not be empty")

// UnaryServerInterceptor 返回校验共享密钥的 unary server interceptor。
// 每个 unary 请求必须携带 metadata MetadataKey 且值与 token 常量时间相等，否则拒绝。
// 注意：本 interceptor 只覆盖 unary 调用，新增 stream RPC 方法需另配 stream interceptor。
func UnaryServerInterceptor(token string) (grpc.UnaryServerInterceptor, error) {
	if token == "" {
		return nil, ErrMissingToken
	}

	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (any, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "unauthenticated")
		}
		values := md.Get(MetadataKey)
		if len(values) != 1 || subtle.ConstantTimeCompare([]byte(values[0]), []byte(token)) != 1 {
			return nil, status.Error(codes.Unauthenticated, "unauthenticated")
		}

		return handler(ctx, req)
	}, nil
}

// UnaryClientInterceptor 返回向每个 unary 请求注入共享密钥的 client interceptor。
// token 为空时返回 ErrMissingToken，调用方应像 server 侧一样 fail fast，
// 避免服务带病启动、运行期每次调用被拒却无从排查。
func UnaryClientInterceptor(token string) (grpc.UnaryClientInterceptor, error) {
	if token == "" {
		return nil, ErrMissingToken
	}

	return func(ctx context.Context, method string, req, reply any,
		cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ctx = metadata.AppendToOutgoingContext(ctx, MetadataKey, token)
		return invoker(ctx, method, req, reply, cc, opts...)
	}, nil
}
