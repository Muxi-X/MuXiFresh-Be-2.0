package logic

import (
	"MuXiFresh-Be-2.0/app/form/model"
	"MuXiFresh-Be-2.0/app/form/rpc/internal/svc"
	"MuXiFresh-Be-2.0/app/form/rpc/pb"
	"MuXiFresh-Be-2.0/common/tool"
	"context"
	"errors"
	"fmt"
	"github.com/zeromicro/go-zero/core/logx"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type CreateFormLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateFormLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateFormLogic {
	return &CreateFormLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateFormLogic) CreateForm(in *pb.CreateReq) (*pb.CreateResp, error) {
	callerID, err := callerIDFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}
	// 报名表归属只允许本人：拒绝内部调用者代他人建表（ID 由客户端传入不可信）
	if callerID != in.UserId {
		return nil, errors.New("无权创建该报名表")
	}
	avatar, err := tool.ValidateAvatarURL(in.Avatar)
	if err != nil {
		return nil, err
	}
	userId, err := primitive.ObjectIDFromHex(in.UserId)
	if err != nil {
		return nil, err
	}
	formID, err := l.svcCtx.FormClient.InsertReturnID(l.ctx, &model.EntryForm{
		UserId:        userId,
		Avatar:        avatar,
		Major:         in.Major,
		Grade:         in.Grade,
		Gender:        in.Gender,
		Phone:         in.Phone,
		Group:         in.Group,
		Reason:        in.Reason,
		Knowledge:     in.Knowledge,
		SelfIntro:     in.SelfIntro,
		ExtraQuestion: in.ExtraQuestion,
		UpdateAt:      time.Now(),
		CreateAt:      time.Now(),
	})

	if err != nil {
		return nil, err
	}
	return &pb.CreateResp{
		FormID: fmt.Sprint(formID)[10:34],
	}, nil
}
