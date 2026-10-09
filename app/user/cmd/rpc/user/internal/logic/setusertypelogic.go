package logic

import (
	"MuXiFresh-Be-2.0/app/user/cmd/rpc/user/internal/svc"
	"MuXiFresh-Be-2.0/app/user/cmd/rpc/user/pb"
	"MuXiFresh-Be-2.0/app/userauth/model"
	"MuXiFresh-Be-2.0/common/tool"
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
		Email:    tool.NormalizeEmail(in.Email),
		UserType: in.UserType,
	})

	if err != nil {
		return nil, err
	}
	if update.MatchedCount == 0 {
		return nil, fmt.Errorf("email is valid")
	}
	return &pb.SetUserTypeResp{
		Flag: true,
	}, nil
}
