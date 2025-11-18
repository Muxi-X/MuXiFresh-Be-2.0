package logic

import (
	"MuXiFresh-Be-2.0/app/form/rpc/entryformclient"
	"context"

	"MuXiFresh-Be-2.0/app/form/api/internal/svc"
	"MuXiFresh-Be-2.0/app/form/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetApplicantNumberLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetApplicantNumberLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetApplicantNumberLogic {
	return &GetApplicantNumberLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetApplicantNumberLogic) GetApplicantNumber(req *types.GetApplicantNumberReq) (resp *types.GetApplicantNumberResp, err error) {

	numberResp, err := l.svcCtx.FormClient.GetApplicantNumber(l.ctx, &entryformclient.GetApplicantNumberReq{
		Group:  req.Group,
		Year:   req.Year,
		Season: req.Season,
	})

	if err != nil {
		return nil, err
	}

	return &types.GetApplicantNumberResp{
		Number: numberResp.Number,
	}, nil
}
