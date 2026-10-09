package logic

import (
	"MuXiFresh-Be-2.0/app/userauth/model"
	"MuXiFresh-Be-2.0/common/globalKey"
	"MuXiFresh-Be-2.0/common/tool"
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
	email := tool.NormalizeEmail(in.Email)
	userInfo := &model.UserInfo{
		Avatar:    l.svcCtx.Config.DefaultUserInfo.Avatar,
		NickName:  l.svcCtx.Config.DefaultUserInfo.NickName + "_" + tool.RandStringBytes(6),
		Email:     email,
		StudentID: globalKey.NULL,
		UserType:  globalKey.Freshman,
		UpdateAt:  time.Now(),
		CreateAt:  time.Now(),
	}
	if err := l.svcCtx.UserInfoClient.Insert(l.ctx, userInfo); err != nil {
		return nil, err
	}
	if err := l.svcCtx.UserAuthClient.Insert(l.ctx, &model.UserAuth{
		Email:      email,
		Password:   in.Password,
		UserInfoID: userInfo.ID,
		UpdateAt:   time.Now(),
		CreateAt:   time.Now(),
	}); err != nil {
		// 补偿：UserAuth 写入失败时回滚已插入的 userinfo，避免孤儿 userInfo
		// （注册两步写无事务，DB 抖动时 UserAuth 失败会留下无 auth 的 userinfo）
		if _, delErr := l.svcCtx.UserInfoClient.Delete(l.ctx, userInfo.ID.Hex()); delErr != nil {
			logx.WithContext(l.ctx).Errorf("register rollback userinfo %s failed: %v", userInfo.ID.Hex(), delErr)
		}
		return nil, err
	}
	return &pb.RegisterDataResp{
		ID: userInfo.ID.String()[10:34],
	}, nil
}
