package logic

import (
	"MuXiFresh-Be-2.0/app/schedule/rpc/scheduleclient"
	"MuXiFresh-Be-2.0/common/ctxData"
	"MuXiFresh-Be-2.0/common/globalKey"
	"MuXiFresh-Be-2.0/common/xerr"
	"context"

	"MuXiFresh-Be-2.0/app/schedule/api/internal/svc"
	"MuXiFresh-Be-2.0/app/schedule/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CheckLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCheckLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CheckLogic {
	return &CheckLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CheckLogic) Check(req *types.CheckReq) (resp *types.CheckResp, err error) {
	//确定scheduleID
	userid := ctxData.GetUserIdFromCtx(l.ctx)
	if req.ScheduleID == globalKey.Myself {
		u, err := l.svcCtx.UserInfoClient.FindOne(l.ctx, userid)
		if err != nil {
			return nil, xerr.NewErrCode(xerr.DB_ERROR).Status()

		}
		req.ScheduleID = u.ScheduleID.String()[10:34]
	}
	c, err := l.svcCtx.ScheduleClient.Check(l.ctx, &scheduleclient.CheckReq{
		UserId:     userid,
		ScheduleID: req.ScheduleID,
	})
	if err != nil {
		return nil, err
	}
	return &types.CheckResp{
		Name:            c.Name,
		School:          c.School,
		Major:           c.Major,
		Group:           c.Group,
		EntryFormStatus: c.EntryFormStatus,
		AdmissionStatus: c.AdmissionStatus,
	}, nil
}
