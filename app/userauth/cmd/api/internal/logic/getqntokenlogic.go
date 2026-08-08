package logic

import (
	"context"

	"MuXiFresh-Be-2.0/app/userauth/cmd/api/internal/common/tube"
	"MuXiFresh-Be-2.0/app/userauth/cmd/api/internal/svc"
	"MuXiFresh-Be-2.0/app/userauth/cmd/api/internal/types"
	"MuXiFresh-Be-2.0/common/ctxData"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetQNTokenLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetQNTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetQNTokenLogic {
	return &GetQNTokenLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetQNTokenLogic) GetQNToken(req *types.GetQNTokenReq) (resp *types.GetQNTokenResp, err error) {
	qnToken, err := tube.GetQNToken(ctxData.GetUserIdFromCtx(l.ctx))
	if err != nil {
		return nil, err
	}

	return &types.GetQNTokenResp{
		QNToken: qnToken,
	}, nil
}
