package logic

import (
	"MuXiFresh-Be-2.0/app/schedule/api/internal/svc"
	"MuXiFresh-Be-2.0/app/schedule/api/internal/types"
	"MuXiFresh-Be-2.0/app/schedule/rpc/scheduleclient"
	"MuXiFresh-Be-2.0/common/ctxData"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/metadata"
)

type CreateScheduleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateScheduleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateScheduleLogic {
	return &CreateScheduleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateScheduleLogic) CreateSchedule(req *types.CreateReq) (resp *types.CreateResp, err error) {
	userId := ctxData.GetUserIdFromCtx(l.ctx)
	_, err = l.svcCtx.ScheduleClient.Create(metadata.AppendToOutgoingContext(l.ctx, ctxData.CallerIDKey, userId), &scheduleclient.CreateReq{
		UserId: userId,
	})
	if err != nil {
		return nil, err
	}
	return &types.CreateResp{
		Flag: true,
	}, nil
}
