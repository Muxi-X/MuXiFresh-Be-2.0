package svc

import (
	"MuXiFresh-Be-2.0/app/task/cmd/api/internal/config"
	"MuXiFresh-Be-2.0/app/task/cmd/rpc/assignment/assignmentclient"
	"MuXiFresh-Be-2.0/app/task/cmd/rpc/comment/commentclient"
	"MuXiFresh-Be-2.0/app/task/cmd/rpc/submission/submissionclient"
	"MuXiFresh-Be-2.0/app/user/cmd/rpc/user/userclient"
	"MuXiFresh-Be-2.0/common/rpcauth"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

type ServiceContext struct {
	Config           config.Config
	AssignmentClient assignmentclient.AssignmentClient
	SubmissionClient submissionclient.SubmissionClient
	CommentClient    commentclient.CommentClient
	UserClient       userclient.UserClient
}

func NewServiceContext(c config.Config) *ServiceContext {
	clientInterceptor, err := rpcauth.UnaryClientInterceptor(c.Infra.RpcAuth.Token)
	if err != nil {
		panic(err)
	}
	rpcOpts := []zrpc.ClientOption{
		zrpc.WithDialOption(grpc.WithChainUnaryInterceptor(clientInterceptor)),
	}

	return &ServiceContext{
		Config:           c,
		AssignmentClient: assignmentclient.NewAssignmentClient(zrpc.MustNewClient(c.AssignmentConf, rpcOpts...)),
		SubmissionClient: submissionclient.NewSubmissionClient(zrpc.MustNewClient(c.SubmissionConf, rpcOpts...)),
		CommentClient:    commentclient.NewCommentClient(zrpc.MustNewClient(c.CommentConf, rpcOpts...)),
		UserClient:       userclient.NewUserClient(zrpc.MustNewClient(c.UserConf, rpcOpts...)),
	}
}
