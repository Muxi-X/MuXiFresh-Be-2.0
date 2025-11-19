package logic

import (
	"MuXiFresh-Be-2.0/app/test/rpc/internal/svc"
	"MuXiFresh-Be-2.0/app/test/rpc/pb"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

type CheckTestResultLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCheckTestResultLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CheckTestResultLogic {
	return &CheckTestResultLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CheckTestResultLogic) CheckTestResult(in *pb.TestInfoReq) (*pb.TestInfoResp, error) {
	return &pb.TestInfoResp{}, nil
}
