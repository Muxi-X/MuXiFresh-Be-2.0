package assigned

import (
	"MuXiFresh-Be-2.0/app/task/cmd/rpc/assignment/assignmentclient"
	"context"
	"github.com/jinzhu/copier"

	"MuXiFresh-Be-2.0/app/task/cmd/api/internal/svc"
	"MuXiFresh-Be-2.0/app/task/cmd/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAssignmentListSelectedLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取作业列表（组别 年份 学期）
func NewGetAssignmentListSelectedLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAssignmentListSelectedLogic {
	return &GetAssignmentListSelectedLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAssignmentListSelectedLogic) GetAssignmentListSelected(req *types.GetAssignmentListSelectedReq) (resp *types.GetAssignmentListSelectedResp, err error) {
	getListSelectedResp, err := l.svcCtx.AssignmentClient.GetAssignmentListSelected(l.ctx, &assignmentclient.GetAssignmentListSelectedReq{
		Group:    req.Group,
		Year:     int64(req.Year),
		Semester: req.Semester,
	})
	if err != nil {
		return nil, err
	}
	var titles []types.Title
	copier.Copy(&titles, &getListSelectedResp.Titles)
	return &types.GetAssignmentListSelectedResp{
		Titles: titles,
	}, nil
}
