package logic

import (
	"MuXiFresh-Be-2.0/app/form/rpc/entryformclient"
	schedulemodel "MuXiFresh-Be-2.0/app/schedule/model"
	externalModel "MuXiFresh-Be-2.0/app/userauth/model"
	"MuXiFresh-Be-2.0/common/ctxData"
	"MuXiFresh-Be-2.0/common/xerr"
	"context"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"

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

func (l *CreateFormLogic) CreateForm(req *types.CreateReq) (resp *types.CreateResp, err error) {
	userId := ctxData.GetUserIdFromCtx(l.ctx)
	if req.FormId != "" {
		_, err = l.svcCtx.FormClient.UpdateForm(l.ctx, &entryformclient.CreateReq{
			FormId:        req.FormId,
			UserId:        userId,
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
	CtResp, err := l.svcCtx.FormClient.CreateForm(l.ctx, &entryformclient.CreateReq{
		UserId:        userId,
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
	u, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		return nil, xerr.ErrExistInvalidId
	}
	f, err := primitive.ObjectIDFromHex(CtResp.FormID)
	if err != nil {
		return nil, xerr.ErrExistInvalidId
	}

	_, err = l.svcCtx.UserInfoModelClient.Update(l.ctx, &externalModel.UserInfo{
		ID:          u,
		EntryFormID: f,
		UpdateAt:    time.Now(),
	})

	if err != nil {
		switch err {
		case externalModel.ErrNotFound:
			return nil, xerr.ErrNotFind
		default:
			return nil, xerr.NewErrCode(xerr.DB_ERROR)
		}
	}

	_, err = l.svcCtx.ScheduleModel.UpdateByUserId(l.ctx, &schedulemodel.Schedule{
		UserID:          u,
		EntryFormStatus: "已提交",
		AdmissionStatus: "已报名",
	})

	if err != nil {
		return nil, xerr.NewErrCode(xerr.DB_ERROR)
	}
	return &types.CreateResp{
		Flag: true,
	}, nil
}
