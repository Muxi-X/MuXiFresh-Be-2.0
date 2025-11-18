package logic

import (
	"MuXiFresh-Be-2.0/app/userauth/cmd/rpc/accountCenter/internal/svc"
	"MuXiFresh-Be-2.0/app/userauth/cmd/rpc/accountCenter/pb"
	"MuXiFresh-Be-2.0/app/userauth/model"
	"MuXiFresh-Be-2.0/common/xerr"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

type ExistEmailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewExistEmailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ExistEmailLogic {
	return &ExistEmailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ExistEmailLogic) ExistEmail(in *pb.ExistEmailReq) (*pb.ExistEmailResp, error) {

	_, err := l.svcCtx.UserAuthClient.FindOneByEmail(l.ctx, in.Email)
	if err != nil {
		if err == model.ErrNotFound {
			return &pb.ExistEmailResp{
				Exist: false,
			}, nil
		}
		return nil, xerr.NewErrCode(xerr.DB_ERROR).Status()
	}
	return &pb.ExistEmailResp{
		Exist: true,
	}, nil
}
