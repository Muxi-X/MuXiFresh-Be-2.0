package comment

import (
	"MuXiFresh-Be-2.0/app/task/cmd/api/internal/svc"
	"MuXiFresh-Be-2.0/app/task/cmd/api/internal/types"
	"MuXiFresh-Be-2.0/app/task/cmd/rpc/comment/commentclient"
	"MuXiFresh-Be-2.0/common/ctxData"
	"MuXiFresh-Be-2.0/common/xerr"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

type DelSubmissionCommentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDelSubmissionCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DelSubmissionCommentLogic {
	return &DelSubmissionCommentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DelSubmissionCommentLogic) DelSubmissionComment(req *types.DelSubmissionCommentReq) (resp *types.DelSubmissionCommentResp, err error) {

	isMyCmtResp, err := l.svcCtx.CommentClient.IsMyComment(l.ctx, &commentclient.IsMyCommentReq{
		UserId:    ctxData.GetUserIdFromCtx(l.ctx),
		CommentID: req.CommentID,
	})
	if err != nil {
		return nil, err
	}
	if !isMyCmtResp.Flag {
		return nil, xerr.ErrPermissionDenied
	}
	delCommentResp, err := l.svcCtx.CommentClient.DelSubmissionComment(l.ctx, &commentclient.DelSubmissionCommentReq{
		CommentID: req.CommentID,
	})
	if err != nil {
		return nil, err
	}
	return &types.DelSubmissionCommentResp{
		Flag: delCommentResp.Flag,
	}, nil
}
