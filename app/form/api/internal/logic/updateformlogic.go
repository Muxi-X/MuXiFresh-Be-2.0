package logic

import (
	"context"
	"errors"

	"MuXiFresh-Be-2.0/app/form/model"
	"MuXiFresh-Be-2.0/common/ctxData"

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

// UpdateForm 修改报名表。当 form_id 指向往届表（往届报名者在新一届重新提交）时，
// 走与提交一致的幂等流程：确保本届表存在并更新，旧表保留供历届审阅。
// 前端仍按 form_id 走修改流程，接口契约不变。
func (l *UpdateFormLogic) UpdateForm(req *types.CreateReq) (resp *types.CreateResp, err error) {
	userId := ctxData.GetUserIdFromCtx(l.ctx)
	if userId == "" {
		return nil, errors.New("身份缺失")
	}

	// 归属校验：目标表必须属于当前用户。这里只要求"是我自己的表"而非"必须是最新
	// 一届的表"——跨届重报后 userinfo 已指向新表，客户端缓存的旧 form_id 也应能继续
	// 提交（否则移动端重报成功后再次点击会因缓存旧 id 被拒），且旧表仍是本人所有，
	// 不存在越权。实际写入的目标由 submitEntryForm 解析为本届表。
	target, err := l.svcCtx.EntryFormModel.FindOne(l.ctx, req.FormId)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, errors.New("无权修改该报名表")
		}
		return nil, err
	}
	if target.UserId.Hex() != userId {
		return nil, errors.New("无权修改该报名表")
	}

	// 已录取成员（实习期/已转正）不得重复报名
	if err := ensureNotAdmittedMember(l.ctx, l.svcCtx.ScheduleModel, userId); err != nil {
		return nil, err
	}

	if err := submitEntryForm(l.ctx, l.svcCtx, userId, req); err != nil {
		return nil, err
	}
	return &types.CreateResp{
		Flag: true,
	}, nil
}
