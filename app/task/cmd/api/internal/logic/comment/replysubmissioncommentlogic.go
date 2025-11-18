package comment

import (
	"MuXiFresh-Be-2.0/app/task/cmd/api/internal/svc"
	"MuXiFresh-Be-2.0/app/task/cmd/api/internal/types"
	"MuXiFresh-Be-2.0/app/task/cmd/rpc/comment/commentclient"
	"MuXiFresh-Be-2.0/common/ctxData"
	"context"
	"github.com/zeromicro/go-zero/core/logx"
)

type ReplySubmissionCommentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewReplySubmissionCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReplySubmissionCommentLogic {
	return &ReplySubmissionCommentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ReplySubmissionCommentLogic) ReplySubmissionComment(req *types.ReplySubmissionCommentReq) (resp *types.ReplySubmissionCommentResp, err error) {
	replyCommentResp, err := l.svcCtx.CommentClient.ReplySubmissionComment(l.ctx, &commentclient.ReplySubmissionCommentReq{
		UserId:       ctxData.GetUserIdFromCtx(l.ctx),
		SubmissionID: req.SubmissionID,
		FatherID:     req.FatherID,
		Content:      req.Content,
	})
	if err != nil {
		return nil, err
	}
	return &types.ReplySubmissionCommentResp{
		Flag: replyCommentResp.Flag,
	}, nil
}
