package logic

import (
	"MuXiFresh-Be-2.0/app/task/cmd/rpc/assignment/internal/svc"
	"MuXiFresh-Be-2.0/app/task/cmd/rpc/assignment/pb"
	"context"
	"github.com/zeromicro/go-zero/core/logx"
)

type DelAssignmentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDelAssignmentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DelAssignmentLogic {
	return &DelAssignmentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DelAssignmentLogic) DelAssignment(in *pb.DelAssignmentReq) (*pb.DelAssignmentResp, error) {
	_, err := l.svcCtx.AssignmentModelClient.Delete(l.ctx, in.AssignmentID)
	if err != nil {
		return nil, err
	}
	return &pb.DelAssignmentResp{
		Flag: true,
	}, nil
}
