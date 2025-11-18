package logic

import (
	"MuXiFresh-Be-2.0/app/task/model"
	"MuXiFresh-Be-2.0/common/globalKey"
	"context"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"

	"MuXiFresh-Be-2.0/app/task/cmd/rpc/comment/internal/svc"
	"MuXiFresh-Be-2.0/app/task/cmd/rpc/comment/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetSubmissionCommentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetSubmissionCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetSubmissionCommentLogic {
	return &SetSubmissionCommentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SetSubmissionCommentLogic) SetSubmissionComment(in *pb.SetSubmissionCommentReq) (*pb.SetSubmissionCommentResp, error) {

	userId, err := primitive.ObjectIDFromHex(in.UserId)
	if err != nil {
		return nil, err
	}
	submissionId, err := primitive.ObjectIDFromHex(in.SubmissionID)
	if err != nil {
		return nil, err
	}
	comment := &model.Comment{
		UserId:       userId,
		SubmissionID: submissionId,
		Content:      in.Content,
		UpdateAt:     time.Now(),
		CreateAt:     time.Now(),
	}

	if err = l.svcCtx.CommentModel.Insert(l.ctx, comment); err != nil {
		return nil, err
	}

	//如果是管理员权限才更新审阅状态
	userInfo, err := l.svcCtx.UserInfoModel.FindOne(l.ctx, in.UserId)
	if userInfo.UserType == globalKey.Admin || userInfo.UserType == globalKey.SuperAdmin {
		submission := model.Submission{
			ID:     submissionId,
			Status: globalKey.Reviewed,
		}

		if _, err = l.svcCtx.SubmissionModel.Update(l.ctx, &submission); err != nil {
			return nil, err
		}
	}

	return &pb.SetSubmissionCommentResp{
		Flag: true,
	}, nil
}
