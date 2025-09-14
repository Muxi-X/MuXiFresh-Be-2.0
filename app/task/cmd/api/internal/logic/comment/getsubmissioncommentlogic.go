package comment

import (
	"MuXiFresh-Be-2.0/app/task/cmd/api/internal/svc"
	"MuXiFresh-Be-2.0/app/task/cmd/api/internal/types"
	"MuXiFresh-Be-2.0/app/task/cmd/rpc/comment/commentclient"
	"context"
	"github.com/jinzhu/copier"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetSubmissionCommentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetSubmissionCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSubmissionCommentLogic {
	return &GetSubmissionCommentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetSubmissionCommentLogic) GetSubmissionComment(req *types.GetSubmissionCommentReq) (resp *types.GetSubmissionCommentResp, err error) {
	getCommentResp, err := l.svcCtx.CommentClient.GetSubmissionComment(l.ctx, &commentclient.GetSubmissionCommentReq{
		SubmissionID: req.SubmissionID,
	})
	if err != nil {
		return nil, err
	}
	var comments []types.Comment
	err = copier.Copy(&comments, &getCommentResp.Comments)
	if err != nil {
		return nil, err
	}
	commentMap := make(map[string]int)
	var index = 0
	var tree []types.Comment
	for _, comment := range comments {
		if comment.FatherID == "000000000000000000000000" {
			tree = append(tree, comment)
			commentMap[comment.CommentID] = index
			index++
		}
	}
	for _, comment := range comments {
		if _, ok := commentMap[comment.CommentID]; !ok {
			tree[commentMap[comment.FatherID]].Replies = append(tree[commentMap[comment.FatherID]].Replies, comment)
		}
	}
	return &types.GetSubmissionCommentResp{
		Comments: tree,
	}, nil
}
