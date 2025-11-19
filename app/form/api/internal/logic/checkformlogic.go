package logic

import (
	"MuXiFresh-Be-2.0/app/form/rpc/entryformclient"
	"MuXiFresh-Be-2.0/common/ctxData"
	"MuXiFresh-Be-2.0/common/globalKey"
	"context"

	"MuXiFresh-Be-2.0/app/form/api/internal/svc"
	"MuXiFresh-Be-2.0/app/form/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
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
	//确定formid
	if req.EntryFormID == globalKey.Myself {
		userid := ctxData.GetUserIdFromCtx(l.ctx)
		userInfo, err := l.svcCtx.UserInfoModelClient.FindOne(l.ctx, userid)
		if err != nil {
			return nil, err
		}
		req.EntryFormID = userInfo.EntryFormID.String()[10:34]
	}

	r, err := l.svcCtx.FormClient.CheckForm(l.ctx, &entryformclient.CheckReq{
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
