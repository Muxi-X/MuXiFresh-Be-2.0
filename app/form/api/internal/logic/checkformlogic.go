package logic

import (
	"context"
	"errors"

	"MuXiFresh-Be-2.0/app/form/rpc/entryformclient"
	"MuXiFresh-Be-2.0/common/ctxData"
	"MuXiFresh-Be-2.0/common/globalKey"

	"MuXiFresh-Be-2.0/app/form/api/internal/svc"
	"MuXiFresh-Be-2.0/app/form/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/metadata"
)

type CheckFormLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCheckFormLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CheckFormLogic {
	return &CheckFormLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CheckFormLogic) CheckForm(req *types.CheckReq) (resp *types.CheckResp, err error) {
	//确定formid，并做归属校验：只能查看当前用户自己的报名表
	userid := ctxData.GetUserIdFromCtx(l.ctx)
	if userid == "" {
		return nil, errors.New("身份缺失")
	}
	userInfo, err := l.svcCtx.UserInfoModelClient.FindOne(l.ctx, userid)
	if err != nil {
		return nil, err
	}
	// 管理员/超管可查看任意报名表（审阅场景）；其余用户仅能查看本人报名表
	isAdmin := userInfo.UserType == globalKey.Admin || userInfo.UserType == globalKey.SuperAdmin
	if req.EntryFormID == globalKey.Myself {
		if userInfo.EntryFormID.IsZero() {
			return nil, errors.New("尚未提交报名表")
		}
		req.EntryFormID = userInfo.EntryFormID.Hex()
	} else if !isAdmin && (userInfo.EntryFormID.IsZero() || req.EntryFormID != userInfo.EntryFormID.Hex()) {
		return nil, errors.New("无权查看该报名表")
	}

	r, err := l.svcCtx.FormClient.CheckForm(metadata.AppendToOutgoingContext(l.ctx, ctxData.CallerIDKey, userid), &entryformclient.CheckReq{
		EntryFormID: req.EntryFormID,
	})
	if err != nil {
		return nil, err
	}
	return &types.CheckResp{
		FormId:        req.EntryFormID,
		Avatar:        r.Avatar,
		Major:         r.Major,
		Grade:         r.Grade,
		Gender:        r.Gender,
		Phone:         r.Phone,
		Group:         r.Group,
		Reason:        r.Reason,
		Knowledge:     r.Knowledge,
		SelfIntro:     r.SelfIntro,
		ExtraQuestion: r.ExtraQuestion,
	}, nil
}
