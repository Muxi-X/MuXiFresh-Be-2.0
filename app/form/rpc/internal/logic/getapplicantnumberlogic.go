package logic

import (
	"MuXiFresh-Be-2.0/common/xerr"
	"context"
	"time"

	"MuXiFresh-Be-2.0/app/form/rpc/internal/svc"
	"MuXiFresh-Be-2.0/app/form/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetApplicantNumberLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetApplicantNumberLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetApplicantNumberLogic {
	return &GetApplicantNumberLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetApplicantNumberLogic) GetApplicantNumber(in *pb.GetApplicantNumberReq) (*pb.GetApplicantNumberResp, error) {
	//秋招
	startTime := time.Date(int(in.Year), time.July, 1, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(int(in.Year), time.December, 31, 23, 59, 59, 999999999, time.UTC)

	if in.Season == "spring" {
		startTime = time.Date(int(in.Year), time.January, 1, 0, 0, 0, 0, time.UTC)
		endTime = time.Date(int(in.Year), time.June, 31, 23, 59, 59, 999999999, time.UTC)
	}

	number, err := l.svcCtx.FormClient.CountByGroup(l.ctx, in.Group, startTime, endTime)
	if err != nil {
		return nil, xerr.NewErrCode(xerr.DB_ERROR).Status()
	}

	return &pb.GetApplicantNumberResp{
		Number: number,
	}, nil
}
