package logic

import (
	"MuXiFresh-Be-2.0/app/userauth/model"
	"MuXiFresh-Be-2.0/common/tool"
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"

	"MuXiFresh-Be-2.0/app/userauth/cmd/rpc/accountCenter/internal/svc"
	"MuXiFresh-Be-2.0/app/userauth/cmd/rpc/accountCenter/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetStudentIDLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetStudentIDLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetStudentIDLogic {
	return &SetStudentIDLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SetStudentIDLogic) SetStudentID(in *pb.SetStudentIDReq) (*pb.SetStudentIDResp, error) {

	//是否绑定
	_, err := l.svcCtx.UserInfoClient.FindByStudentID(l.ctx, in.StudentID)
	if err == nil {
		return nil, errors.New("the student_id  already bind")
	}
	//一站式登录
	if !tool.CCNULogin(in.StudentID, in.Password) {
		return nil, errors.New("student_id or password is wrong")
	}
	//存userinfo
	uid, err := primitive.ObjectIDFromHex(in.UserId)
	if err != nil {
		return nil, err
	}
	_, err = l.svcCtx.UserInfoClient.Update(l.ctx, &model.UserInfo{
		ID:        uid,
		StudentID: in.StudentID,
		UpdateAt:  time.Now(),
	})
	if err != nil {
		return nil, err
	}
	return &pb.SetStudentIDResp{
		Flag: true,
	}, nil
}
