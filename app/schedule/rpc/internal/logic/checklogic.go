package logic

import (
	formmodel "MuXiFresh-Be-2.0/app/form/model"
	"context"

	"MuXiFresh-Be-2.0/app/schedule/rpc/internal/svc"
	"MuXiFresh-Be-2.0/app/schedule/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type CheckLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCheckLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CheckLogic {
	return &CheckLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CheckLogic) Check(in *pb.CheckReq) (*pb.CheckResp, error) {

	f := &formmodel.EntryForm{}
	f, err := l.svcCtx.EntryFormClient.FindOneByUserId(l.ctx, in.UserId)
	if err != nil && err != formmodel.ErrNotFound {
		return nil, err
	}

	s, err := l.svcCtx.ScheduleClient.FindOne(l.ctx, in.ScheduleID)
	if err != nil {
		return nil, err
	}

	userInfo, err := l.svcCtx.UserInfoClient.FindOne(l.ctx, in.UserId)
	if err != nil {
		return nil, err
	}

	return &pb.CheckResp{
		Name:            userInfo.Name,
		School:          userInfo.School,
		Major:           f.Major,
		Group:           f.Group,
		EntryFormStatus: s.EntryFormStatus,
		AdmissionStatus: s.AdmissionStatus,
	}, nil
}
