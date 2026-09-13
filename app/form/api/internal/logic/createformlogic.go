package logic

import (
	"MuXiFresh-Be-2.0/app/form/rpc/entryformclient"
	schedulemodel "MuXiFresh-Be-2.0/app/schedule/model"
	externalModel "MuXiFresh-Be-2.0/app/userauth/model"
	"MuXiFresh-Be-2.0/common/ctxData"
	"context"
	"errors"
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
		l.rollbackEntryForm(f)
		return nil, err
	}

	// upsert 后查一次拿 scheduleID，写入 userinfo 关联
	schedule, err := l.svcCtx.ScheduleModel.FindOneByUserId(l.ctx, userId)
	if err != nil {
		l.rollbackEntryForm(f)
		return nil, err
	}
	sid := schedule.ID
	res, err := l.svcCtx.UserInfoModelClient.Update(l.ctx, &externalModel.UserInfo{
		ID:          u,
		EntryFormID: f,
		ScheduleID:  sid,
		UpdateAt:    time.Now(),
	})
	if err != nil {
		l.rollbackEntryForm(f)
		return nil, err
	}
	// userinfo 不存在时 Update 静默 no-op（MatchedCount=0），此时表未被关联，
	// 必须回滚，否则留下"有表无 userinfo"的孤儿。
	if res.MatchedCount == 0 {
		l.rollbackEntryForm(f)
		return nil, errors.New("用户信息缺失，报名失败")
	}
	return &types.CreateResp{
		Flag: true,
	}, nil
}

// rollbackEntryForm 在报名后续步骤失败时删除第 1 步已插入的 entry_form，
// 避免留下无 schedule/userinfo 关联的孤儿报名表（会被审阅列表静默跳过）。
// 补偿失败仅记日志，不改变对外返回的原始错误。
func (l *CreateFormLogic) rollbackEntryForm(formID primitive.ObjectID) {
	if _, err := l.svcCtx.EntryFormModel.Delete(l.ctx, formID.Hex()); err != nil {
		logx.WithContext(l.ctx).Errorf("createform rollback entry_form %s failed: %v", formID.Hex(), err)
	}
}
