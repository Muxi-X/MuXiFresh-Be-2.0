package logic

import (
	"MuXiFresh-Be-2.0/app/form/model"
	"MuXiFresh-Be-2.0/app/form/rpc/internal/svc"
	"MuXiFresh-Be-2.0/app/form/rpc/pb"
	"MuXiFresh-Be-2.0/common/tool"
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateFormLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateFormLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateFormLogic {
	return &UpdateFormLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateFormLogic) UpdateForm(in *pb.CreateReq) (*pb.CreateResp, error) {
	callerID, err := callerIDFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}
	if err := checkEntryFormWriteAccess(l.ctx, l.svcCtx, callerID, in.FormId); err != nil {
		return nil, err
	}
	avatar, err := tool.ValidateAvatarURL(in.Avatar)
	if err != nil {
		return nil, err
	}

	// 归属只信任 metadata 中的 callerID，忽略入参 in.UserId，防止直连篡改表单归属
	u, err := primitive.ObjectIDFromHex(callerID)
	if err != nil {
		return nil, err
	}
	f, err := primitive.ObjectIDFromHex(in.FormId)
	if err != nil {
		return nil, err
	}
	//form := model.EntryForm{}
	//copier.Copy(&form,in)
	updateRet, err := l.svcCtx.FormClient.Update(l.ctx, &model.EntryForm{
		UserId:        u,
		ID:            f,
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
	})
	if err != nil {
		return nil, err
	}
	if updateRet.MatchedCount == 0 {
		return nil, model.ErrNotFound
	}
	return &pb.CreateResp{
		FormID: fmt.Sprint(updateRet.UpsertedID),
	}, nil
}
