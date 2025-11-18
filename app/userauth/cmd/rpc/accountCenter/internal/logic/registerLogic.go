package logic

import (
	"MuXiFresh-Be-2.0/app/userauth/model"
	"MuXiFresh-Be-2.0/common/globalKey"
	"MuXiFresh-Be-2.0/common/tool"
	"MuXiFresh-Be-2.0/common/xerr"
	"context"
	"time"

	"MuXiFresh-Be-2.0/app/userauth/cmd/rpc/accountCenter/internal/svc"
	"MuXiFresh-Be-2.0/app/userauth/cmd/rpc/accountCenter/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type RegisterLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RegisterLogic) Register(in *pb.RegisterDataReq) (*pb.RegisterDataResp, error) {
	userInfo := &model.UserInfo{
		Avatar:     l.svcCtx.Config.DefaultUserInfo.Avatar,
		Nickname:   l.svcCtx.Config.DefaultUserInfo.NickName + "_" + tool.RandStringBytes(6),
		Email:      in.Email,
		TestResult: &model.ExamResult{},
		StudentID:  globalKey.NULL,
		UserType:   globalKey.Freshman,
		UpdateAt:   time.Now(),
		CreateAt:   time.Now(),
	}
	if err := l.svcCtx.UserInfoClient.Insert(l.ctx, userInfo); err != nil {
		return nil, xerr.NewErrCode(xerr.DB_ERROR).Status()
	}
	if err := l.svcCtx.UserAuthClient.Insert(l.ctx, &model.UserAuth{
		Email:      in.Email,
		Password:   in.Password,
		UserInfoID: userInfo.ID,
		UpdateAt:   time.Now(),
		CreateAt:   time.Now(),
	}); err != nil {
		return nil, xerr.NewErrCode(xerr.DB_ERROR).Status()
	}
	return &pb.RegisterDataResp{
		ID: userInfo.ID.String()[10:34],
	}, nil
}
