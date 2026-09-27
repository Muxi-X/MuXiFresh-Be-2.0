package logic

import (
	"context"
	"errors"
	"time"

	"MuXiFresh-Be-2.0/app/form/model"
	"MuXiFresh-Be-2.0/app/form/rpc/internal/svc"
)

// ensureWithinDeadline 校验 now 是否在 cycle 届次的报名截止时间之前，超时返回错误。
// 未配置时：秋招回退默认值（当年 10-06 23:59），春招无截止（放行）。
func ensureWithinDeadline(ctx context.Context, svcCtx *svc.ServiceContext, cycle string, now time.Time) error {
	setting, err := svcCtx.RecruitSettingModel.GetByCycle(ctx, cycle)
	var limit time.Time
	hasLimit := false
	switch {
	case err == nil:
		limit, hasLimit = setting.Deadline, true
	case errors.Is(err, model.ErrNotFound):
		if d, ok := model.DefaultDeadlineForCycle(cycle); ok {
			limit, hasLimit = d, true
		}
	default:
		return err
	}

	if hasLimit && now.After(limit) {
		return errors.New("未在报名时间内")
	}
	return nil
}
