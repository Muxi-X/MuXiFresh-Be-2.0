package logic

import (
	"MuXiFresh-Be-2.0/app/userauth/cmd/rpc/accountCenter/internal/svc"
	"MuXiFresh-Be-2.0/app/userauth/cmd/rpc/accountCenter/pb"
	"MuXiFresh-Be-2.0/app/userauth/model"
	"MuXiFresh-Be-2.0/common/xerr"
	"context"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetEmailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetEmailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetEmailLogic {
	return &SetEmailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SetEmailLogic) SetEmail(in *pb.SetEmailReq) (*pb.SetEmailResp, error) {

	uid, err := primitive.ObjectIDFromHex(in.UserId)
	if err != nil {
		return nil, xerr.ErrExistInvalidId.Status()
	}
	userAuth := &model.UserAuth{
		Email:      in.Email,
		UserInfoID: uid,
	}
	_, err = l.svcCtx.UserAuthClient.UpdateByUserId(l.ctx, userAuth)
	if err != nil {
		return nil, xerr.NewErrCode(xerr.DB_ERROR).Status()
	}
	userInfo := &model.UserInfo{
		ID:    uid,
		Email: in.Email,
	}
	_, err = l.svcCtx.UserInfoClient.Update(l.ctx, userInfo)
	if err != nil {
		return nil, xerr.NewErrCode(xerr.DB_ERROR).Status()
	}
	return &pb.SetEmailResp{
		Flag: true,
	}, nil
}
