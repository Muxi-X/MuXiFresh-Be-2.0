package logic

import (
	"context"
	"errors"
	"time"

	"MuXiFresh-Be-2.0/app/form/api/internal/svc"
	"MuXiFresh-Be-2.0/app/form/api/internal/types"
	"MuXiFresh-Be-2.0/app/form/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetRecruitDeadlineLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetRecruitDeadlineLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetRecruitDeadlineLogic {
	return &GetRecruitDeadlineLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// GetRecruitDeadline 返回指定届次的报名截止时间；cycle 为空时取当前届次。
// 未配置时：秋招回退到默认值（当年 10-06 23:59，rev=0），春招返回空串（放行）。
func (l *GetRecruitDeadlineLogic) GetRecruitDeadline(req *types.GetRecruitDeadlineReq) (resp *types.GetRecruitDeadlineResp, err error) {
	cycle := req.Cycle
	if cycle == "" {
		cycle = model.CycleOf(time.Now())
	} else if _, _, ok := model.ParseCycle(cycle); !ok {
		return nil, errors.New("invalid cycle")
	}

	setting, err := l.svcCtx.RecruitSettingModel.GetByCycle(l.ctx, cycle)
	switch {
	case err == nil:
		return &types.GetRecruitDeadlineResp{
			Cycle:    cycle,
			Deadline: model.FormatDeadline(setting.Deadline),
			Rev:      setting.Rev,
		}, nil
	case errors.Is(err, model.ErrNotFound):
		if d, ok := model.DefaultDeadlineForCycle(cycle); ok {
			return &types.GetRecruitDeadlineResp{
				Cycle:    cycle,
				Deadline: model.FormatDeadline(d),
				Rev:      0,
			}, nil
		}
		return &types.GetRecruitDeadlineResp{Cycle: cycle}, nil
	default:
		return nil, err
	}
}
