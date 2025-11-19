package logic

import (
	"MuXiFresh-Be-2.0/app/intro/model"
	"context"
	"time"

	"MuXiFresh-Be-2.0/app/intro/rpc/internal/svc"
	"MuXiFresh-Be-2.0/app/intro/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetRecruitInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetRecruitInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetRecruitInfoLogic {
	return &GetRecruitInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetRecruitInfoLogic) GetRecruitInfo(in *pb.RecruitInfoReq) (*pb.RecruitInfoResp, error) {
	l.svcCtx.IntroClient.Insert(l.ctx, &model.Intro{
		URL:      in.URL,
		UpdateAt: time.Now(),
		CreateAt: time.Now(),
	})
	return &pb.RecruitInfoResp{
		URL: in.URL,
	}, nil
}
