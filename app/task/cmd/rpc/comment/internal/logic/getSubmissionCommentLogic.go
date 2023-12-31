package logic

import (
	"MuXiFresh-Be-2.0/app/task/cmd/rpc/comment/internal/svc"
	"MuXiFresh-Be-2.0/app/task/cmd/rpc/comment/pb"
	"MuXiFresh-Be-2.0/common/convert"
	"context"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetSubmissionCommentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetSubmissionCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSubmissionCommentLogic {
	return &GetSubmissionCommentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetSubmissionCommentLogic) GetSubmissionComment(in *pb.GetSubmissionCommentReq) (*pb.GetSubmissionCommentResp, error) {

	comments, err := l.svcCtx.CommentModel.FindBySubmissionID(l.ctx, in.SubmissionID)
	if err != nil {
		return nil, err
	}
	var cmtsWithUserInfo []*pb.Comment
	for _, comment := range comments {
		userInfo, err := l.svcCtx.UserInfoModel.FindOne(l.ctx, comment.UserId.String()[10:34])
		if err != nil {
			return nil, err
		}
		entryForm, err := l.svcCtx.EntryFormModel.FindOne(l.ctx, userInfo.EntryFormID.String()[10:34])
		if err != nil {
			return nil, err
		}
		group := entryForm.Grade + "级" + convert.GroupCvtChinese(entryForm.Group) + "成员"
		cmtsWithUserInfo = append(cmtsWithUserInfo, &pb.Comment{
			CommentID: comment.ID.String()[10:34],
			Avatar:    userInfo.Avatar,
			NickName:  userInfo.NickName,
			Group:     group,
			Content:   comment.Content,
		})
	}
	return &pb.GetSubmissionCommentResp{
		Comments: cmtsWithUserInfo,
	}, nil
}
