package logic

import (
	"MuXiFresh-Be-2.0/app/intro/rpc/introclient"
	"context"

	"MuXiFresh-Be-2.0/app/intro/api/internal/svc"
	"MuXiFresh-Be-2.0/app/intro/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GroupIntroLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGroupIntroLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupIntroLogic {
	return &GroupIntroLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GroupIntroLogic) GroupIntro(req *types.GroupIntroReq) (resp *types.GroupIntroResp, err error) {
	intro, err := l.svcCtx.IntroClient.GetGroupIntro(l.ctx, &introclient.GroupIntroReq{
		GroupName: req.GroupName,
	})
	if err != nil {
		return nil, err
	}
	return &types.GroupIntroResp{
		Intro: intro.Intro,
	}, nil
}
