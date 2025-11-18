package logic

import (
	formmodel "MuXiFresh-Be-2.0/app/form/model"
	"MuXiFresh-Be-2.0/app/test/api/internal/svc"
	"MuXiFresh-Be-2.0/app/test/api/internal/types"
	"MuXiFresh-Be-2.0/app/user/cmd/rpc/user/userclient"
	"MuXiFresh-Be-2.0/app/userauth/model"
	"MuXiFresh-Be-2.0/common/ctxData"
	"MuXiFresh-Be-2.0/common/globalKey"
	"MuXiFresh-Be-2.0/common/xerr"
	"context"
	"github.com/zeromicro/go-zero/core/logx"
)

type CheckTestResultLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCheckTestResultLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CheckTestResultLogic {
	return &CheckTestResultLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CheckTestResultLogic) CheckTestResult(req *types.TestInfoReq) (resp *types.TestInfoResp, err error) {

	uid := ctxData.GetUserIdFromCtx(l.ctx)
	if req.UserID != "myself" {
		//管理员认证
		getUserTypeResp, err := l.svcCtx.UserClient.GetUserType(l.ctx, &userclient.GetUserTypeReq{
			UserId: uid,
		})
		if err != nil {
			return nil, err
		}
		if getUserTypeResp.UserType != globalKey.Admin && getUserTypeResp.UserType != globalKey.SuperAdmin {
			return nil, xerr.ErrPermissionDenied
		}
		uid = req.UserID
	}

	userInfo, err := l.svcCtx.UserInfoClient.FindOne(l.ctx, uid)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, xerr.ErrNotFind
		}
		return nil, xerr.NewErrCode(xerr.DB_ERROR)
	}

	formid := userInfo.EntryFormID

	form, err := l.svcCtx.FormClient.FindOne(l.ctx, formid.String()[10:34])
	if err != nil && err != formmodel.ErrNotFound {
		return nil, xerr.NewErrCode(xerr.DB_ERROR)
	}
	if form == nil {
		form = new(formmodel.EntryForm)
	}
	typesChoice := make([]types.ChoiceItem, len(userInfo.TestChoice))
	for i, item := range userInfo.TestChoice {
		typesChoice[i] = types.ChoiceItem(item)
	}

	return &types.TestInfoResp{
		UserId:      userInfo.ID.String()[10:34],
		Name:        userInfo.Name,
		Gender:      form.Gender,
		Major:       form.Major,
		Grade:       form.Grade,
		LeQunXing:   userInfo.TestResult.LeQunXing,
		YouHengXing: userInfo.TestResult.YouHengXing,
		XingFenXing: userInfo.TestResult.XingFenXing,
		CongHuiXing: userInfo.TestResult.CongHuiXing,
		JiaoJiXing:  userInfo.TestResult.JiaoJiXing,
		HuaiYiXing:  userInfo.TestResult.HuaiYiXing,
		WenDingXing: userInfo.TestResult.WenDingXing,
		Choice:      typesChoice,
	}, nil
}
