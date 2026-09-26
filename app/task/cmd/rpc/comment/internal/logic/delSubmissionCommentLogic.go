package logic

import (
	"context"
	"errors"

	"MuXiFresh-Be-2.0/app/task/cmd/rpc/comment/internal/svc"
	"MuXiFresh-Be-2.0/app/task/cmd/rpc/comment/pb"
	usermodel "MuXiFresh-Be-2.0/app/userauth/model"
	"MuXiFresh-Be-2.0/common/globalKey"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/mon"
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
	// 鉴权：仅评论作者本人或 Admin/SuperAdmin 可删除（身份经 grpc metadata 注入），
	// 防止 RPC 直连删任意评论
	callerID, err := callerIDFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}
	comment, err := l.svcCtx.CommentModel.FindOne(l.ctx, in.CommentID)
	if err != nil {
		if errors.Is(err, mon.ErrNotFound) {
			return nil, errors.New("评论不存在")
		}
		return nil, err
	}
	if comment.UserId.Hex() != callerID {
		caller, err := l.svcCtx.UserInfoModel.FindOne(l.ctx, callerID)
		if err != nil {
			// caller 不存在（not-found / 非法 hex）与无权同语义，归一不泄露内部错误串
			if errors.Is(err, mon.ErrNotFound) || errors.Is(err, usermodel.ErrInvalidObjectId) {
				return nil, errors.New("无权删除该评论")
			}
			return nil, err
		}
		if caller.UserType != globalKey.Admin && caller.UserType != globalKey.SuperAdmin {
			return nil, errors.New("无权删除该评论")
		}
	}

	if _, err = l.svcCtx.CommentModel.Delete(l.ctx, in.CommentID); err != nil {
		return nil, err
	}
	return &pb.DelSubmissionCommentResp{
		Flag: true,
	}, nil
}
