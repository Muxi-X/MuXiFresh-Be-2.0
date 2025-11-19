package logic

import (
	"context"

	"MuXiFresh-Be-2.0/app/task/cmd/rpc/comment/internal/svc"
	"MuXiFresh-Be-2.0/app/task/cmd/rpc/comment/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type IsMyCommentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewIsMyCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *IsMyCommentLogic {
	return &IsMyCommentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *IsMyCommentLogic) IsMyComment(in *pb.IsMyCommentReq) (*pb.IsMyCommentResp, error) {

	comment, err := l.svcCtx.CommentModel.FindOne(l.ctx, in.CommentID)
	if err != nil {
		return nil, err
	}
	return &pb.IsMyCommentResp{
		Flag: comment.UserId.String()[10:34] == in.UserId,
	}, nil
}
