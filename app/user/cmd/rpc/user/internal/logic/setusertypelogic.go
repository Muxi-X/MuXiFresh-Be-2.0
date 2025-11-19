package logic

import (
	"MuXiFresh-Be-2.0/app/user/cmd/rpc/user/internal/svc"
	"MuXiFresh-Be-2.0/app/user/cmd/rpc/user/pb"
	"MuXiFresh-Be-2.0/app/userauth/model"
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetUserTypeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetUserTypeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetUserTypeLogic {
	return &SetUserTypeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SetUserTypeLogic) SetUserType(in *pb.SetUserTypeReq) (*pb.SetUserTypeResp, error) {

	update, err := l.svcCtx.UserInfoModel.UpdateByEmail(l.ctx, &model.UserInfo{
		Email:    in.Email,
		UserType: in.UserType,
	})

	if update.MatchedCount == 0 {
		return nil, fmt.Errorf("email is valid")
	}
	if err != nil {
		return nil, err
	}
	return &pb.SetUserTypeResp{
		Flag: true,
	}, nil
}
