package logic

import (
	"MuXiFresh-Be-2.0/app/userauth/cmd/rpc/accountCenter/internal/svc"
	"MuXiFresh-Be-2.0/app/userauth/cmd/rpc/accountCenter/pb"
	"MuXiFresh-Be-2.0/app/userauth/model"
	"MuXiFresh-Be-2.0/common/xerr"
	"context"
	"fmt"

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
		if err == model.ErrNotFound {
			return nil, xerr.ErrStudentIdHasNotBingToEmail.Status()
		}
		return nil, xerr.NewErrCode(xerr.DB_ERROR).Status()
	}
	//一站式登录
	//if !tool.CCNULogin(in.StudentID, in.Password) {
	//	return nil, xerr.ErrStudentIdOrPasswordIsWrong.Status()
	//}
	ok, err := l.svcCtx.CCNUSvc.Login(l.ctx, in.GetStudentID(), in.GetPassword())
	fmt.Println("DEBUG: -------------------------")
	if err != nil || !ok {
		fmt.Println(err.Error(), ok)
		return nil, xerr.NewErrMsg(err.Error())
	}
	fmt.Println("DEBUG: -------------------------")
	//返回userinfoID
	return &pb.CcnuLoginResp{
		UserinfoID: userInfo.ID.String()[10:34],
	}, nil
}
