package logic

import (
	"MuXiFresh-Be-2.0/app/form/rpc/entryformclient"
	schedulemodel "MuXiFresh-Be-2.0/app/schedule/model"
	externalModel "MuXiFresh-Be-2.0/app/userauth/model"
	"MuXiFresh-Be-2.0/common/ctxData"
	"context"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"time"

	"MuXiFresh-Be-2.0/app/form/api/internal/svc"
	"MuXiFresh-Be-2.0/app/form/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/metadata"
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
	CtResp, err := l.svcCtx.FormClient.CreateForm(metadata.AppendToOutgoingContext(l.ctx, ctxData.CallerIDKey, userId), &entryformclient.CreateReq{
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
		return nil, err
	}
	f, err := primitive.ObjectIDFromHex(CtResp.FormID)
	if err != nil {
		return nil, err
	}

	// 原子 upsert schedule：不存在则创建、存在则更新（配合 user_id 唯一索引防并发双写）。
	// 唯一索引冲突（并发双击）时 DuplicateKey 视为已创建，继续走后续关联。
	_, err = l.svcCtx.ScheduleModel.UpsertByUserId(l.ctx, &schedulemodel.Schedule{
		UserID:          u,
		EntryFormStatus: "已提交",
		AdmissionStatus: "已报名",
	})
	if err != nil && !mongo.IsDuplicateKeyError(err) {
		return nil, err
	}

	// upsert 后查一次拿 scheduleID，写入 userinfo 关联
	schedule, err := l.svcCtx.ScheduleModel.FindOneByUserId(l.ctx, userId)
	if err != nil {
		return nil, err
	}
	sid := schedule.ID
	_, err = l.svcCtx.UserInfoModelClient.Update(l.ctx, &externalModel.UserInfo{
		ID:          u,
		EntryFormID: f,
		ScheduleID:  sid,
		UpdateAt:    time.Now(),
	})
	if err != nil {
		return nil, err
	}
	return &types.CreateResp{
		Flag: true,
	}, nil
}
