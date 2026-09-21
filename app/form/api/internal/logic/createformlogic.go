package logic

import (
	"context"
	"errors"

	"MuXiFresh-Be-2.0/common/ctxData"

	"MuXiFresh-Be-2.0/app/form/api/internal/svc"
	"MuXiFresh-Be-2.0/app/form/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateFormLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateFormLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateFormLogic {
	return &CreateFormLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// CreateForm 提交报名表。本届已有表时覆盖该表（幂等），不再报重复键错误；
// 这正是"往届报名者今年重新报名"的入口。
func (l *CreateFormLogic) CreateForm(req *types.CreateReq) (resp *types.CreateResp, err error) {
	userId := ctxData.GetUserIdFromCtx(l.ctx)
	if userId == "" {
		return nil, errors.New("身份缺失")
	}

	// 已录取成员（实习期/已转正）不得重复报名
	if err := ensureNotAdmittedMember(l.ctx, l.svcCtx.ScheduleModel, userId); err != nil {
		return nil, err
	}

	if err := submitEntryForm(l.ctx, l.svcCtx, userId, req); err != nil {
		return nil, err
	}
	return &types.CreateResp{
		Flag: true,
	}, nil
}
