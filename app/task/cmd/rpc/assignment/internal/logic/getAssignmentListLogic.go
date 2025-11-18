package logic

import (
	"MuXiFresh-Be-2.0/app/task/cmd/rpc/assignment/internal/svc"
	"MuXiFresh-Be-2.0/app/task/cmd/rpc/assignment/pb"
	"MuXiFresh-Be-2.0/common/xerr"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAssignmentListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetAssignmentListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAssignmentListLogic {
	return &GetAssignmentListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetAssignmentListLogic) GetAssignmentList(in *pb.GetAssignmentListReq) (*pb.GetAssignmentListResp, error) {
	assignments, err := l.svcCtx.AssignmentModelClient.FindByGroup(l.ctx, in.Group)
	if err != nil {
		return nil, xerr.NewErrCode(xerr.DB_ERROR).Status()
	}
	var titles []*pb.Title
	for _, assignment := range assignments {
		titles = append(titles, &pb.Title{
			ID:   assignment.ID.String()[10:34],
			Text: assignment.TitleText,
		})
	}
	return &pb.GetAssignmentListResp{
		Titles: titles,
	}, nil
}
