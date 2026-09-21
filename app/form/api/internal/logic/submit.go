package logic

import (
	"context"
	"errors"
	"time"

	"MuXiFresh-Be-2.0/app/form/api/internal/svc"
	"MuXiFresh-Be-2.0/app/form/api/internal/types"
	"MuXiFresh-Be-2.0/app/form/model"
	"MuXiFresh-Be-2.0/app/form/rpc/entryformclient"
	scheduleModel "MuXiFresh-Be-2.0/app/schedule/model"
	externalModel "MuXiFresh-Be-2.0/app/userauth/model"
	"MuXiFresh-Be-2.0/common/ctxData"

	"github.com/zeromicro/go-zero/core/logx"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc/metadata"
)

// submitEntryForm 是报名与重报的公共落库流程：幂等地确保用户拥有本届报名表，
// 再建立 schedule 与 userinfo 关联。POST /form 与 PUT /form（往届表重报）共用。
func submitEntryForm(ctx context.Context, svcCtx *svc.ServiceContext, userId string, req *types.CreateReq) error {
	formID, created, err := ensureCurrentCycleForm(ctx, svcCtx, userId, req)
	if err != nil {
		return err
	}
	if err := upsertScheduleAndAssociate(ctx, svcCtx, userId, formID); err != nil {
		// 仅回滚本次新建的表；复用的既有表是用户已提交内容，不能删
		if created {
			rollbackUnassociatedForm(ctx, svcCtx, userId, formID)
		}
		return err
	}
	return nil
}

// ensureCurrentCycleForm 确保用户拥有一份本届报名表并返回其 ID。
//
// 先查本届是否已有表：已存在则用本次提交内容覆盖（重复提交/并发双击/重报后再次
// 提交都幂等），不存在才新建。先查后建而非"建失败再回查"，避免把非重复键的失败
// （mongo 抖动、avatar 校验失败）误判为并发。
// 并发下两个请求都判定"不存在"时，后到者会撞唯一索引，此时回查复用先到者建的表。
func ensureCurrentCycleForm(ctx context.Context, svcCtx *svc.ServiceContext, userId string, req *types.CreateReq) (primitive.ObjectID, bool, error) {
	existing, err := findCurrentCycleForm(ctx, svcCtx, userId)
	switch {
	case err == nil:
		rpcReq := toRPCReq(userId, req)
		rpcReq.FormId = existing.ID.Hex()
		if _, err := svcCtx.FormClient.UpdateForm(withCaller(ctx, userId), rpcReq); err != nil {
			return primitive.NilObjectID, false, err
		}
		return existing.ID, false, nil
	case !errors.Is(err, model.ErrNotFound):
		return primitive.NilObjectID, false, err
	}

	created, err := svcCtx.FormClient.CreateForm(withCaller(ctx, userId), toRPCReq(userId, req))
	if err == nil {
		id, err := primitive.ObjectIDFromHex(created.FormID)
		if err != nil {
			return primitive.NilObjectID, false, err
		}
		return id, true, nil
	}
	// 并发：另一请求可能已建好本届表。复用前用本次内容覆盖，避免本次提交被静默丢弃。
	if existing, lookErr := findCurrentCycleForm(ctx, svcCtx, userId); lookErr == nil {
		rpcReq := toRPCReq(userId, req)
		rpcReq.FormId = existing.ID.Hex()
		if _, err := svcCtx.FormClient.UpdateForm(withCaller(ctx, userId), rpcReq); err != nil {
			return primitive.NilObjectID, false, err
		}
		return existing.ID, false, nil
	}
	return primitive.NilObjectID, false, err
}

// findCurrentCycleForm 返回用户本届的报名表。
//
// 除按显式 cycle 精确匹配外，也把"无 cycle 但 createAt 落在本届"的存量表（本次
// 改动上线前创建）视为本届，否则会对老数据重复建表，导致同届出现两份表。
func findCurrentCycleForm(ctx context.Context, svcCtx *svc.ServiceContext, userId string) (*model.EntryForm, error) {
	cycle := model.CycleOf(time.Now())

	form, err := svcCtx.EntryFormModel.FindByUserIdAndCycle(ctx, userId, cycle)
	if err == nil {
		return form, nil
	}
	if !errors.Is(err, model.ErrNotFound) {
		return nil, err
	}

	latest, err := svcCtx.EntryFormModel.FindOneByUserId(ctx, userId)
	if err != nil {
		return nil, err
	}
	if latest.EffectiveCycle() == cycle {
		return latest, nil
	}
	return nil, model.ErrNotFound
}

// upsertScheduleAndAssociate 原子 upsert 进度并把它与本届报名表写回 userinfo 关联。
// schedule 不存在时创建，兼容 PR_3 之前报名、没有 schedule 的老用户重报。
func upsertScheduleAndAssociate(ctx context.Context, svcCtx *svc.ServiceContext, userId string, formID primitive.ObjectID) error {
	u, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		return err
	}

	// 入口的录取守卫已拦截已录取成员，故此处重置为"已报名"不会误伤其录取状态
	if _, err := svcCtx.ScheduleModel.UpsertByUserId(ctx, &scheduleModel.Schedule{
		UserID:          u,
		EntryFormStatus: "已提交",
		AdmissionStatus: "已报名",
	}); err != nil && !mongo.IsDuplicateKeyError(err) {
		return err
	}

	schedule, err := svcCtx.ScheduleModel.FindOneByUserId(ctx, userId)
	if err != nil {
		return err
	}
	ret, err := svcCtx.UserInfoModelClient.Update(ctx, &externalModel.UserInfo{
		ID:          u,
		EntryFormID: formID,
		ScheduleID:  schedule.ID,
		UpdateAt:    time.Now(),
	})
	if err != nil {
		return err
	}
	// userinfo 不存在时 Update 静默 no-op（MatchedCount=0），必须显式判断
	if ret.MatchedCount == 0 {
		return errors.New("用户信息缺失，报名失败")
	}
	return nil
}

func toRPCReq(userId string, req *types.CreateReq) *entryformclient.CreateReq {
	return &entryformclient.CreateReq{
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
	}
}

func withCaller(ctx context.Context, userId string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, ctxData.CallerIDKey, userId)
}

// rollbackUnassociatedForm 回滚本次新建的报名表，但仅当它尚未被任何请求关联为
// 用户的 entry_form_id 时才删除。
//
// 并发场景：请求 A 新建表 F 后关联失败；与此同时请求 B 已复用了 F 并把
// userinfo.entry_form_id 指向 F。若 A 无条件删除 F，B 的关联就会指向不存在的表。
// 故删除前重新核对：只有当用户的 entry_form_id 仍不是 F（说明无人关联）才删。
// 补偿脱离请求生命周期：请求超时/客户端断连时 ctx 已取消，用它删除会立即失败。
func rollbackUnassociatedForm(ctx context.Context, svcCtx *svc.ServiceContext, userId string, formID primitive.ObjectID) {
	rctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	userInfo, err := svcCtx.UserInfoModelClient.FindOne(rctx, userId)
	if err != nil {
		logx.WithContext(ctx).Errorf("form rollback check userinfo %s failed, skip delete: %v", userId, err)
		return
	}
	if !userInfo.EntryFormID.IsZero() && userInfo.EntryFormID == formID {
		logx.WithContext(ctx).Infof("entry_form %s already associated, skip rollback", formID.Hex())
		return
	}
	if _, err := svcCtx.EntryFormModel.Delete(rctx, formID.Hex()); err != nil {
		logx.WithContext(ctx).Errorf("form submit rollback entry_form %s failed: %v", formID.Hex(), err)
	}
}
