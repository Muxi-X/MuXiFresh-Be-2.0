package logic

import (
	"MuXiFresh-Be-2.0/app/user/cmd/rpc/user/internal/svc"
	"MuXiFresh-Be-2.0/app/user/cmd/rpc/user/pb"
	"MuXiFresh-Be-2.0/app/userauth/model"
	"MuXiFresh-Be-2.0/common/convert"
	"MuXiFresh-Be-2.0/common/globalKey"
	"MuXiFresh-Be-2.0/common/xerr"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserInfoLogic {
	return &GetUserInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUserInfoLogic) GetUserInfo(in *pb.GetUserInfoReq) (*pb.GetUserInfoResp, error) {

	userInfo, err := l.svcCtx.UserInfoModel.FindOne(l.ctx, in.UserId)
	if err != nil {
		switch err {
		case model.ErrNotFound:
			return nil, xerr.ErrNotFind.Status()
		case model.ErrInvalidObjectId:
			return nil, xerr.ErrExistInvalidId.Status()
		default:
			return nil, xerr.NewErrCode(xerr.DB_ERROR).Status()
		}
	}
	group := "尚未报名"
	if !userInfo.EntryFormID.IsZero() {
		group = "简历审阅中"
		schedule, err := l.svcCtx.ScheduleModel.FindOne(l.ctx, userInfo.ScheduleID.String()[10:34])
		if err != nil {
			return nil, xerr.NewErrCode(xerr.DB_ERROR).Status()
		}
		if schedule.AdmissionStatus != globalKey.Registered {
			entryForm, err := l.svcCtx.EntryFormModel.FindOneByUserId(l.ctx, in.UserId)
			if err != nil {
				return nil, xerr.NewErrCode(xerr.DB_ERROR).Status()
			}
			identity := "实习生"
			if schedule.AdmissionStatus != globalKey.Internship {
				identity = "成员"
			}
			group = entryForm.Grade + "级" + convert.GroupCvtChinese(entryForm.Group) + identity
		}
	}
	return &pb.GetUserInfoResp{
		Avatar:    userInfo.Avatar,
		NickName:  userInfo.Nickname,
		Name:      userInfo.Name,
		School:    userInfo.School,
		Group:     group,
		Email:     userInfo.Email,
		StudentID: userInfo.StudentID,
		QQ:        userInfo.QQ,
	}, nil
}
