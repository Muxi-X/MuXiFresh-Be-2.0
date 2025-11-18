package logic

import (
	"MuXiFresh-Be-2.0/app/form/rpc/entryformclient"
	schedulemodel "MuXiFresh-Be-2.0/app/schedule/model"
	"MuXiFresh-Be-2.0/common/ctxData"
	"MuXiFresh-Be-2.0/common/globalKey"
	"MuXiFresh-Be-2.0/common/xerr"
	"context"

	"MuXiFresh-Be-2.0/app/form/api/internal/svc"
	"MuXiFresh-Be-2.0/app/form/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateFormLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateFormLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateFormLogic {
	return &UpdateFormLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateFormLogic) UpdateForm(req *types.CreateReq) (resp *types.CreateResp, err error) {
	userId := ctxData.GetUserIdFromCtx(l.ctx)
	//查录取状态，转正后不修改group
	schedule, err := l.svcCtx.ScheduleModel.FindOneByUserId(l.ctx, userId)
	if err != nil {
		switch err {
		case schedulemodel.ErrNotFound:
			return nil, xerr.ErrNotFind
		case schedulemodel.ErrInvalidObjectId:
			return nil, xerr.ErrExistInvalidId
		default:
			return nil, xerr.NewErrCode(xerr.DB_ERROR)
		}
	}
	if schedule.AdmissionStatus == globalKey.Formal {
		req.Group = globalKey.NULL
	}
	_, err = l.svcCtx.FormClient.UpdateForm(l.ctx, &entryformclient.CreateReq{
		UserId:        userId,
		FormId:        req.FormId,
		Avatar:        req.Avatar,
		Major:         req.Major,
		Grade:         req.Grade,
		Gender:        req.Gender,
		Phone:         req.Phone,
		Group:         req.Group,
		Reason:        req.Reason,
		Knowledge:     req.Knowledge,
		SelfIntro:     req.SelfIntro,
		ExtraQuestion: req.ExtraQuestion,
	})

	if err != nil {
		return nil, err
	}
	return &types.CreateResp{
		Flag: true,
	}, nil
}
