package logic

import (
	"context"
	"errors"

	"MuXiFresh-Be-2.0/app/form/rpc/entryformclient"
	"MuXiFresh-Be-2.0/common/ctxData"

	"MuXiFresh-Be-2.0/app/form/api/internal/svc"
	"MuXiFresh-Be-2.0/app/form/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/metadata"
)

type UpdateFormLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateFormLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateFormLogic {
	return &UpdateFormLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateFormLogic) UpdateForm(req *types.CreateReq) (resp *types.CreateResp, err error) {
	userId := ctxData.GetUserIdFromCtx(l.ctx)
	if userId == "" {
		return nil, errors.New("身份缺失")
	}
	userInfo, err := l.svcCtx.UserInfoModelClient.FindOne(l.ctx, userId)
	if err != nil {
		return nil, err
	}
	// 归属校验：只能修改当前用户自己的报名表，防止按 form_id 越权修改
	if userInfo.EntryFormID.IsZero() || req.FormId != userInfo.EntryFormID.Hex() {
		return nil, errors.New("无权修改该报名表")
	}

	_, err = l.svcCtx.FormClient.UpdateForm(metadata.AppendToOutgoingContext(l.ctx, ctxData.CallerIDKey, userId), &entryformclient.CreateReq{
		FormId:        req.FormId,
		UserId:        userId,
		Avatar:        req.Avatar,
		Major:         req.Major,
		Grade:         req.Grade,
		Gender:        req.Gender,
		Phone:         req.Phone,
		Group:         req.Group,
		Reason:        req.Reason,
		Knowledge:     req.Knowledge,
		SelfIntro:     req.SelfIntro,
		ExtraQuestion: req.ExtraQuestion,
	})
	if err != nil {
		return nil, err
	}
	return &types.CreateResp{
		Flag: true,
	}, nil
}
