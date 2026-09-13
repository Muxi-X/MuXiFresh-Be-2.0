package logic

import (
	"MuXiFresh-Be-2.0/app/schedule/model"
	userauthModel "MuXiFresh-Be-2.0/app/userauth/model"
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"

	"MuXiFresh-Be-2.0/app/schedule/rpc/internal/svc"
	"MuXiFresh-Be-2.0/app/schedule/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateLogic {
	return &CreateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateLogic) Create(in *pb.CreateReq) (*pb.CreateResp, error) {
	// 归属校验：只允许为调用者本人创建进度，RPC 直连不可信（fail closed）
	callerID, err := callerIDFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}
	if callerID != in.UserId {
		return nil, errors.New("无权创建该进度")
	}
	uid, err := primitive.ObjectIDFromHex(in.UserId)
	if err != nil {
		return nil, err
	}
	scheduleID, err := l.svcCtx.ScheduleClient.InsertGetID(l.ctx, &model.Schedule{
		UserID:          uid,
		EntryFormStatus: "未提交",
		AdmissionStatus: "未报名",
		UpdateAt:        time.Now(),
		CreateAt:        time.Now(),
	})
	if err != nil {
		return nil, err
	}
	sid, err := primitive.ObjectIDFromHex(scheduleID[10:34])
	if err != nil {
		return nil, err
	}
	_, err = l.svcCtx.UserInfoClient.Update(l.ctx, &userauthModel.UserInfo{
		ID:         uid,
		ScheduleID: sid,
	})
	if err != nil {
		return nil, err
	}
	return &pb.CreateResp{
		Flag: true,
	}, nil
}
