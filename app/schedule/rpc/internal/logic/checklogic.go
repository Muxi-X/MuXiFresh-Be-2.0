package logic

import (
	formmodel "MuXiFresh-Be-2.0/app/form/model"
	schedulemodel "MuXiFresh-Be-2.0/app/schedule/model"
	"MuXiFresh-Be-2.0/app/schedule/rpc/internal/svc"
	"MuXiFresh-Be-2.0/app/schedule/rpc/pb"
	"MuXiFresh-Be-2.0/common/xerr"
	"context"

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

	f, err := l.svcCtx.EntryFormClient.FindOneByUserId(l.ctx, in.UserId)

	switch err {
	case formmodel.ErrNotFound:
		f = new(formmodel.EntryForm)
	case nil:
	default:
		return nil, xerr.NewErrCode(xerr.DB_ERROR).Status()
	}

	s, err := l.svcCtx.ScheduleClient.FindOne(l.ctx, in.ScheduleID)
	if err != nil {
		switch err {
		case schedulemodel.ErrNotFound:
			return nil, xerr.ErrNotFind.Status()
		case schedulemodel.ErrInvalidObjectId:
			return nil, xerr.ErrExistInvalidId.Status()
		default:
			return nil, xerr.NewErrCode(xerr.DB_ERROR).Status()
		}
	}

	userInfo, err := l.svcCtx.UserInfoClient.FindOne(l.ctx, in.UserId)
	if err != nil {
		return nil, xerr.NewErrCode(xerr.DB_ERROR).Status()
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
