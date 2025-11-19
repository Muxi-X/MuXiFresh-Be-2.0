package logic

import (
	"MuXiFresh-Be-2.0/common/tool"
	"context"
	"errors"

	"MuXiFresh-Be-2.0/app/userauth/cmd/rpc/accountCenter/internal/svc"
	"MuXiFresh-Be-2.0/app/userauth/cmd/rpc/accountCenter/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type CcnuLoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCcnuLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CcnuLoginLogic {
	return &CcnuLoginLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CcnuLoginLogic) CcnuLogin(in *pb.CcnuLoginReq) (*pb.CcnuLoginResp, error) {

	//查看是否绑定
	userInfo, err := l.svcCtx.UserInfoClient.FindByStudentID(l.ctx, in.StudentID)
	if err != nil {
		return nil, errors.New("not bind to email")
	}
	//一站式登录
	if !tool.CCNULogin(in.StudentID, in.Password) {
		return nil, errors.New("student_id or password is wrong")
	}
	//返回userinfoID
	return &pb.CcnuLoginResp{
		UserinfoID: userInfo.ID.String()[10:34],
	}, nil
}
