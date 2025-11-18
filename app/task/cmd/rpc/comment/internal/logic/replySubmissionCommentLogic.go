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

type ReplySubmissionCommentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewReplySubmissionCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReplySubmissionCommentLogic {
	return &ReplySubmissionCommentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ReplySubmissionCommentLogic) ReplySubmissionComment(in *pb.ReplySubmissionCommentReq) (*pb.ReplySubmissionCommentResp, error) {
	// 校验 userId
	userId, err := primitive.ObjectIDFromHex(in.UserId)
	if err != nil {
		return nil, err
	}

	// 校验 submissionId
	submissionId, err := primitive.ObjectIDFromHex(in.SubmissionID)
	if err != nil {
		return nil, err
	}

	// 校验 fatherId（必须存在，回复就是基于父评论）
	fatherId, err := primitive.ObjectIDFromHex(in.FatherID)
	if err != nil {
		return nil, err
	}

	// 构造新的回复评论
	reply := &model.Comment{
		UserId:       userId,
		SubmissionID: submissionId,
		Content:      in.Content,
		FatherId:     fatherId,
		UpdateAt:     time.Now(),
		CreateAt:     time.Now(),
	}

	// 插入数据库
	if err = l.svcCtx.CommentModel.Insert(l.ctx, reply); err != nil {
		return nil, err
	}

	// 更新 submission 状态为已评论/已复核
	submission := model.Submission{
		ID:     submissionId,
		Status: globalKey.Reviewed,
	}
	if _, err = l.svcCtx.SubmissionModel.Update(l.ctx, &submission); err != nil {
		return nil, err
	}

	return &pb.ReplySubmissionCommentResp{
		Flag: true,
	}, nil
}
