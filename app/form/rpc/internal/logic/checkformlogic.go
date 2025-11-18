package logic

import (
	"MuXiFresh-Be-2.0/app/form/model"
	"MuXiFresh-Be-2.0/common/xerr"
	"context"

	"MuXiFresh-Be-2.0/app/form/rpc/internal/svc"
	"MuXiFresh-Be-2.0/app/form/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type CheckFormLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCheckFormLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CheckFormLogic {
	return &CheckFormLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CheckFormLogic) CheckForm(in *pb.CheckReq) (*pb.CheckResp, error) {
	r, err := l.svcCtx.FormClient.FindOne(l.ctx, in.EntryFormID)
	if err != nil {
		switch err {
		case model.ErrNotFound:
			return nil, xerr.ErrNotFind.Status()
		case model.ErrInvalidObjectId:
			return nil, xerr.ErrExistInvalidId.Status()
		default:
			return nil, xerr.NewErrCode(xerr.DB_ERROR).Status()
		}
	}
	return &pb.CheckResp{
		Avatar:        r.Avatar,
		Major:         r.Major,
		Grade:         r.Grade,
		Gender:        r.Gender,
		Phone:         r.Phone,
		Group:         r.Group,
		Reason:        r.Reason,
		Knowledge:     r.Knowledge,
		SelfIntro:     r.SelfIntro,
		ExtraQuestion: r.ExtraQuestion,
	}, nil
}
