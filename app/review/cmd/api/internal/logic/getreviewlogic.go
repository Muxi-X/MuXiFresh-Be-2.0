package logic

import (
	"MuXiFresh-Be-2.0/app/review/cmd/api/internal/svc"
	"MuXiFresh-Be-2.0/app/review/cmd/api/internal/types"
	"MuXiFresh-Be-2.0/app/user/cmd/rpc/user/userclient"
	"MuXiFresh-Be-2.0/common/ctxData"
	"MuXiFresh-Be-2.0/common/globalKey"
	"context"
	"errors"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetReviewLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetReviewLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetReviewLogic {
	return &GetReviewLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetReviewLogic) GetReview(req *types.GetReviewReq) (resp *types.GetReviewResp, err error) {
	//管理员认证
	getUserTypeResp, err := l.svcCtx.UserClient.GetUserType(l.ctx, &userclient.GetUserTypeReq{
		UserId: ctxData.GetUserIdFromCtx(l.ctx),
	})
	if err != nil {
		return nil, err
	}
	if getUserTypeResp.UserType != globalKey.Admin && getUserTypeResp.UserType != globalKey.SuperAdmin {
		return nil, errors.New("permission denied")
	}
	// 届次窗口与 form 的 CycleOf 分界对齐，避免同一份表落入相邻两届
	startTime, endTime := seasonWindow(req.Year, req.Season)

	rows, err := buildReviewRows(l.ctx, l.svcCtx, groupFilter(req.Group), req.School, req.Grade, req.Status, startTime, endTime)
	if err != nil {
		return nil, err
	}
	// 姓名筛选在分页之前，保证 total 为筛选后的总数
	rows = filterByName(rows, req.Name)

	total := int64(len(rows))
	// 传了 page_size 才分页；不传则全量返回，兼容旧前端
	if req.PageSize > 0 {
		page := req.Page
		if page <= 0 {
			page = 1
		}
		start := (page - 1) * req.PageSize
		if start < 0 || start > total {
			start = total
		}
		end := start + req.PageSize
		if end < start || end > total {
			end = total
		}
		rows = rows[start:end]
	}

	return &types.GetReviewResp{
		Rows:  rows,
		Total: total,
	}, nil
}
