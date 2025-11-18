package logic

import (
	"MuXiFresh-Be-2.0/app/task/cmd/rpc/comment/internal/svc"
	"MuXiFresh-Be-2.0/app/task/cmd/rpc/comment/pb"
	"MuXiFresh-Be-2.0/common/xerr"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

type DelSubmissionCommentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDelSubmissionCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DelSubmissionCommentLogic {
	return &DelSubmissionCommentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DelSubmissionCommentLogic) DelSubmissionComment(in *pb.DelSubmissionCommentReq) (*pb.DelSubmissionCommentResp, error) {

	_, err := l.svcCtx.CommentModel.Delete(l.ctx, in.CommentID)
	if err != nil {
		return nil, xerr.NewErrCode(xerr.DB_ERROR).Status()
	}
	return &pb.DelSubmissionCommentResp{
		Flag: true,
	}, nil
}
