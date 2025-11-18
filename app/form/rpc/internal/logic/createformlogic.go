package logic

import (
	"MuXiFresh-Be-2.0/app/form/model"
	"MuXiFresh-Be-2.0/app/form/rpc/internal/svc"
	"MuXiFresh-Be-2.0/app/form/rpc/pb"
	"MuXiFresh-Be-2.0/common/xerr"
	"context"
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
	userId, err := primitive.ObjectIDFromHex(in.UserId)
	if err != nil {
		return nil, xerr.ErrExistInvalidId.Status()
	}
	formID, err := l.svcCtx.FormClient.InsertReturnID(l.ctx, &model.EntryForm{
		UserId:        userId,
		Avatar:        in.Avatar,
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
		return nil, xerr.NewErrCode(xerr.DB_ERROR).Status()
	}
	return &pb.CreateResp{
		FormID: fmt.Sprint(formID)[10:34],
	}, nil
}
