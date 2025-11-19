package logic

import (
	"MuXiFresh-Be-2.0/app/intro/rpc/introclient"
	"context"

	"MuXiFresh-Be-2.0/app/intro/api/internal/svc"
	"MuXiFresh-Be-2.0/app/intro/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RecruitInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRecruitInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RecruitInfoLogic {
	return &RecruitInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RecruitInfoLogic) RecruitInfo(req *types.RecruitInfoReq) (resp *types.RecruitInfoResp, err error) {
	info, err := l.svcCtx.IntroClient.GetRecruitInfo(l.ctx, &introclient.RecruitInfoReq{
		URL: req.URL,
	})
	if err != nil {
		return nil, err
	}
	return &types.RecruitInfoResp{
		URL: info.URL,
	}, nil
}
