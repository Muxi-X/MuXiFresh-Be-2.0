package logic

import (
	"MuXiFresh-Be-2.0/common/ctxData"
	"context"

	"MuXiFresh-Be-2.0/app/form/api/internal/svc"
	"MuXiFresh-Be-2.0/app/form/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type JudgeUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewJudgeUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *JudgeUserLogic {
	return &JudgeUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *JudgeUserLogic) JudgeUser(req *types.ClickReq) (resp *types.ClickResp, err error) {
	userId := ctxData.GetUserIdFromCtx(l.ctx)
	userInfo, err := l.svcCtx.UserInfoModelClient.FindOne(l.ctx, userId)
	if err != nil {
		return nil, err
	}
	status := "已交表"
	if userInfo.EntryFormID.IsZero() {
		status = "未交表"
	}
	return &types.ClickResp{
		UserType:   userInfo.UserType,
		FormStatus: status,
	}, nil
}
