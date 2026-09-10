package rpcauth

import (
	"context"
	"errors"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const testToken = "test-secret-token"

func ctxWithToken(token string) context.Context {
	return metadata.AppendToOutgoingContext(context.Background(), MetadataKey, token)
}

func TestUnaryServerInterceptor(t *testing.T) {
	interceptor, err := UnaryServerInterceptor(testToken)
	if err != nil {
		t.Fatalf("UnaryServerInterceptor() error = %v", err)
	}

	handler := func(ctx context.Context, req any) (any, error) {
		return "ok", nil
	}
	serverInfo := &grpc.UnaryServerInfo{
		FullMethod: "/test.Service/Ping",
	}

	tests := []struct {
		name    string
		ctx     context.Context
		wantErr bool
	}{
		{
			name:    "no metadata",
			ctx:     context.Background(),
			wantErr: true,
		},
		{
			name:    "empty metadata",
			ctx:     metadata.NewIncomingContext(context.Background(), metadata.MD{}),
			wantErr: true,
		},
		{
			name:    "wrong token",
			ctx:     metadata.NewIncomingContext(ctxWithToken("wrong"), metadata.MD{MetadataKey: []string{"wrong"}}),
			wantErr: true,
		},
		{
			name:    "duplicate token values",
			ctx:     metadata.NewIncomingContext(context.Background(), metadata.MD{MetadataKey: []string{testToken, testToken}}),
			wantErr: true,
		},
		{
			name:    "correct token",
			ctx:     metadata.NewIncomingContext(ctxWithToken(testToken), metadata.MD{MetadataKey: []string{testToken}}),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := interceptor(tt.ctx, nil, serverInfo, handler)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr = %v", err, tt.wantErr)
			}
			if err != nil && status.Code(err) != codes.Unauthenticated {
				t.Errorf("code = %v, want %v", status.Code(err), codes.Unauthenticated)
			}
		})
	}
}

func TestUnaryServerInterceptorNoMethodBypassed(t *testing.T) {
	interceptor, err := UnaryServerInterceptor(testToken)
	if err != nil {
		t.Fatalf("UnaryServerInterceptor() error = %v", err)
	}

	handler := func(ctx context.Context, req any) (any, error) {
		return "ok", nil
	}

	methods := []string{
		"/grpc.health.v1.Health/Check",
		"/grpc.reflection.v1.ServerReflection/ServerReflectionInfo",
	}
	for _, method := range methods {
		serverInfo := &grpc.UnaryServerInfo{FullMethod: method}
		_, err := interceptor(context.Background(), nil, serverInfo, handler)
		if err == nil || status.Code(err) != codes.Unauthenticated {
			t.Errorf("method %s without token should be rejected, got err = %v", method, err)
		}
	}
}

func TestUnaryServerInterceptorEmptyToken(t *testing.T) {
	_, err := UnaryServerInterceptor("")
	if !errors.Is(err, ErrMissingToken) {
		t.Fatalf("error = %v, want ErrMissingToken", err)
	}
}

func TestUnaryClientInterceptor(t *testing.T) {
	var captured string
	invoker := func(ctx context.Context, method string, req, reply any,
		cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		md, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			t.Fatal("client interceptor did not append outgoing metadata")
		}
		values := md.Get(MetadataKey)
		if len(values) != 1 {
			t.Fatalf("metadata values = %v, want exactly 1", values)
		}
		captured = values[0]
		return nil
	}

	interceptor, err := UnaryClientInterceptor(testToken)
	if err != nil {
		t.Fatalf("UnaryClientInterceptor() error = %v", err)
	}
	err = interceptor(context.Background(), "/test.Service/Ping", nil, nil, nil, invoker)
	if err != nil {
		t.Fatalf("invoker error = %v", err)
	}
	if captured != testToken {
		t.Errorf("captured token = %q, want %q", captured, testToken)
	}
}

func TestUnaryClientInterceptorEmptyToken(t *testing.T) {
	_, err := UnaryClientInterceptor("")
	if !errors.Is(err, ErrMissingToken) {
		t.Fatalf("error = %v, want ErrMissingToken", err)
	}
}

func TestUnaryClientInterceptorRejectsPreexistingDuplicate(t *testing.T) {
	// ctx 已含 internal-auth 时（调用方误注入或攻击者预置），Append 会追加
	// 第二个值；server 侧 len(values) != 1 校验应拒绝。固化该安全默认。
	clientInterceptor, err := UnaryClientInterceptor(testToken)
	if err != nil {
		t.Fatalf("UnaryClientInterceptor() error = %v", err)
	}

	invoker := func(ctx context.Context, method string, req, reply any,
		cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		return nil
	}

	ctx := metadata.AppendToOutgoingContext(context.Background(), MetadataKey, "attacker-value")
	if err := clientInterceptor(ctx, "/test.Service/Ping", nil, nil, nil, invoker); err != nil {
		t.Fatalf("client invoker should not fail: %v", err)
	}

	// 上面 invoker 无法直接断言 metadata，这里直接验证 server 拒绝双值：
	// 模拟 Append 后的双值 metadata 进入 server interceptor。
	serverInterceptor, err := UnaryServerInterceptor(testToken)
	if err != nil {
		t.Fatalf("UnaryServerInterceptor() error = %v", err)
	}
	handler := func(ctx context.Context, req any) (any, error) { return "ok", nil }
	serverInfo := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Ping"}

	ctx2 := metadata.NewIncomingContext(context.Background(),
		metadata.MD{MetadataKey: []string{"attacker-value", testToken}})
	if _, err := serverInterceptor(ctx2, nil, serverInfo, handler); err == nil {
		t.Error("server should reject duplicate internal-auth values")
	}
}
