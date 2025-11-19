package logic

import (
	"MuXiFresh-Be-2.0/app/review/cmd/api/internal/svc"
	"MuXiFresh-Be-2.0/app/review/cmd/api/internal/types"
	"MuXiFresh-Be-2.0/app/schedule/model"
	pb2 "MuXiFresh-Be-2.0/app/user/cmd/rpc/user/pb"
	externalModel "MuXiFresh-Be-2.0/app/userauth/model"
	"MuXiFresh-Be-2.0/common/ctxData"
	"MuXiFresh-Be-2.0/common/globalKey"
	"context"
	"errors"
	"github.com/zeromicro/go-zero/core/logx"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SetAdmissionStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSetAdmissionStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetAdmissionStatusLogic {
	return &SetAdmissionStatusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SetAdmissionStatusLogic) SetAdmissionStatus(req *types.SetAdmissionStatusReq) (resp *types.SetAdmissionStatusResp, err error) {

	getUserTypeResp, err := l.svcCtx.UserClient.GetUserType(l.ctx, &pb2.GetUserTypeReq{
		UserId: ctxData.GetUserIdFromCtx(l.ctx),
	})
	if err != nil {
		return nil, err
	}
	if getUserTypeResp.UserType != globalKey.Admin && getUserTypeResp.UserType != globalKey.SuperAdmin {
		return nil, errors.New("permission denied")
	}

	sid, err := primitive.ObjectIDFromHex(req.ScheduleID)
	if err != nil {
		return nil, err
	}
	_, err = l.svcCtx.ScheduleClient.Update(l.ctx, &model.Schedule{
		ID:              sid,
		AdmissionStatus: req.NewStatus,
	})
	if err != nil {
		return nil, err
	}
	//转正 usertype to normal
	if req.NewStatus == globalKey.Formal {
		schedule, err := l.svcCtx.ScheduleClient.FindOne(l.ctx, req.ScheduleID)
		if err != nil {
			return nil, err
		}
		userInfo, err := l.svcCtx.UserInfoModel.FindOne(l.ctx, schedule.UserID.String()[10:34])
		if err != nil {
			return nil, err
		}
		if userInfo.UserType != globalKey.SuperAdmin && userInfo.UserType != globalKey.Admin {
			_, err = l.svcCtx.UserInfoModel.Update(l.ctx, &externalModel.UserInfo{
				ID:       schedule.UserID,
				UserType: globalKey.Normal,
			})
			if err != nil {
				return nil, err
			}
		}
	}

	return &types.SetAdmissionStatusResp{
		Flag: true,
	}, nil
}
