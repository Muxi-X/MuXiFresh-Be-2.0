package logic

import (
	"MuXiFresh-Be-2.0/app/user/cmd/rpc/user/internal/svc"
	"MuXiFresh-Be-2.0/app/user/cmd/rpc/user/pb"
	"MuXiFresh-Be-2.0/app/userauth/model"
	"MuXiFresh-Be-2.0/common/globalKey"
	"MuXiFresh-Be-2.0/common/xerr"
	"context"
	"github.com/zeromicro/go-zero/core/logx"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SetUserInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetUserInfoLogic {
	return &SetUserInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SetUserInfoLogic) SetUserInfo(in *pb.SetUserInfoReq) (*pb.SetUserInfoResp, error) {

	uid, err := primitive.ObjectIDFromHex(in.UserId)
	if err != nil {
		return nil, xerr.ErrExistInvalidId.Status()
	}
	userInfo := &model.UserInfo{
		ID:     uid,
		Avatar: in.Avatar,
		Name:   in.Name,
		School: in.School,
		QQ:     in.QQ,
	}
	if in.NickName != globalKey.NULL {
		userInfo.Nickname = in.NickName
	}
	_, err = l.svcCtx.UserInfoModel.Update(l.ctx, userInfo)
	if err != nil {
		return nil, xerr.NewErrCode(xerr.DB_ERROR).Status()
	}
	return &pb.SetUserInfoResp{
		Flag: true,
	}, nil
}
