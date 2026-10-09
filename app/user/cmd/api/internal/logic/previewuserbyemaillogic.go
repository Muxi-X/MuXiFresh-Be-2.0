package logic

import (
	"MuXiFresh-Be-2.0/app/user/cmd/api/internal/svc"
	"MuXiFresh-Be-2.0/app/user/cmd/api/internal/types"
	"MuXiFresh-Be-2.0/app/user/cmd/rpc/user/userclient"
	"MuXiFresh-Be-2.0/common/ctxData"
	"MuXiFresh-Be-2.0/common/globalKey"
	"MuXiFresh-Be-2.0/common/tool"
	"context"
	"errors"

	"github.com/zeromicro/go-zero/core/logx"
)

type PreviewUserByEmailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPreviewUserByEmailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PreviewUserByEmailLogic {
	return &PreviewUserByEmailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PreviewUserByEmailLogic) PreviewUserByEmail(req *types.PreviewUserByEmailReq) (resp *types.PreviewUserByEmailResp, err error) {
	//super_admin authorization
	getUserTypeResp, err := l.svcCtx.UserClient.GetUserType(l.ctx, &userclient.GetUserTypeReq{
		UserId: ctxData.GetUserIdFromCtx(l.ctx),
	})
	if err != nil {
		return nil, err
	}
	if getUserTypeResp.UserType != globalKey.SuperAdmin {
		return nil, errors.New("permission denied")
	}

	email := tool.NormalizeEmail(req.Email)
	if email == "" {
		return nil, errors.New("email is empty")
	}

	previewResp, err := l.svcCtx.UserClient.GetUserInfoByEmail(l.ctx, &userclient.GetUserInfoByEmailReq{
		Email: email,
	})
	if err != nil {
		return nil, err
	}

	return &types.PreviewUserByEmailResp{
		UserId:   previewResp.UserId,
		Avatar:   previewResp.Avatar,
		NickName: previewResp.NickName,
		Name:     previewResp.Name,
		Email:    previewResp.Email,
		UserType: previewResp.UserType,
	}, nil
}
