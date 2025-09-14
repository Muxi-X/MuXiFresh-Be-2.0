package assigned

import (
	"MuXiFresh-Be-2.0/app/task/cmd/rpc/assignment/assignmentclient"
	"context"

	"MuXiFresh-Be-2.0/app/task/cmd/api/internal/svc"
	"MuXiFresh-Be-2.0/app/task/cmd/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAssignmentInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAssignmentInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAssignmentInfoLogic {
	return &GetAssignmentInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAssignmentInfoLogic) GetAssignmentInfo(req *types.GetAssignmentInfoReq) (resp *types.GetAssignmentInfoResp, err error) {
	getInfoResp, err := l.svcCtx.AssignmentClient.GetAssignmentInfo(l.ctx, &assignmentclient.GetAssignmentInfoReq{AssignmentID: req.AssignmentID})
	if err != nil {
		return nil, err
	}
	return &types.GetAssignmentInfoResp{
		TitleText: getInfoResp.TitleText,
		Content:   getInfoResp.Content,
		Urls:      getInfoResp.Urls,
		Deadline:  getInfoResp.Deadline,
		Semester:  getInfoResp.Semester,
	}, nil
}
