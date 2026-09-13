package logic

import (
	"context"
	"errors"

	"MuXiFresh-Be-2.0/app/schedule/rpc/scheduleclient"
	"MuXiFresh-Be-2.0/common/ctxData"
	"MuXiFresh-Be-2.0/common/globalKey"

	"MuXiFresh-Be-2.0/app/schedule/api/internal/svc"
	"MuXiFresh-Be-2.0/app/schedule/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/metadata"
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
	//确定scheduleID，并做归属校验：只能查询当前用户自己的进度，防止按 schedule_id 越权读取录取状态
	userid := ctxData.GetUserIdFromCtx(l.ctx)
	if userid == "" {
		return nil, errors.New("身份缺失")
	}
	u, err := l.svcCtx.UserInfoClient.FindOne(l.ctx, userid)
	if err != nil {
		return nil, err
	}
	if req.ScheduleID == globalKey.Myself {
		if u.ScheduleID.IsZero() {
			return nil, errors.New("尚未创建进度")
		}
		req.ScheduleID = u.ScheduleID.Hex()
	} else if u.ScheduleID.IsZero() || req.ScheduleID != u.ScheduleID.Hex() {
		return nil, errors.New("无权查看该进度")
	}

	c, err := l.svcCtx.ScheduleClient.Check(metadata.AppendToOutgoingContext(l.ctx, ctxData.CallerIDKey, userid), &scheduleclient.CheckReq{
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
