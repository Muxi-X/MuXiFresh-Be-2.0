package logic

import (
	"MuXiFresh-Be-2.0/app/test/rpc/internal/svc"
	"MuXiFresh-Be-2.0/app/test/rpc/pb"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserTestLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUserTestLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserTestLogic {
	return &UserTestLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UserTestLogic) UserTest(in *pb.TestReq) (*pb.TestResp, error) {
	return &pb.TestResp{
		Flag: true,
	}, nil
}
