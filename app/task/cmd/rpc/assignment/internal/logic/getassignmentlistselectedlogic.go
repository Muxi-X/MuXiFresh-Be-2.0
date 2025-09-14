package logic

import (
	"context"

	"MuXiFresh-Be-2.0/app/task/cmd/rpc/assignment/internal/svc"
	"MuXiFresh-Be-2.0/app/task/cmd/rpc/assignment/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAssignmentListSelectedLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetAssignmentListSelectedLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAssignmentListSelectedLogic {
	return &GetAssignmentListSelectedLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetAssignmentListSelectedLogic) GetAssignmentListSelected(in *pb.GetAssignmentListSelectedReq) (*pb.GetAssignmentListSelectedResp, error) {
	assignments, err := l.svcCtx.AssignmentModelClient.FindByGroupYearandSemester(l.ctx, in.Group, in.Year, in.Semester)
	if err != nil {
		return nil, err
	}
	var titles []*pb.Title
	for _, assignment := range assignments {
		titles = append(titles, &pb.Title{
			ID:   assignment.ID.String()[10:34],
			Text: assignment.TitleText,
		})
	}

	return &pb.GetAssignmentListSelectedResp{
		Titles: titles,
	}, nil
}
