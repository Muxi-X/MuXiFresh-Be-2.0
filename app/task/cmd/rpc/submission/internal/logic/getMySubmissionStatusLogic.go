package logic

import (
	"MuXiFresh-Be-2.0/common/globalKey"
	"context"

	"MuXiFresh-Be-2.0/app/task/cmd/rpc/submission/internal/svc"
	"MuXiFresh-Be-2.0/app/task/cmd/rpc/submission/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMySubmissionStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetMySubmissionStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMySubmissionStatusLogic {
	return &GetMySubmissionStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetMySubmissionStatusLogic) GetMySubmissionStatus(in *pb.GetMySubmissionStatusReq) (*pb.GetMySubmissionStatusResp, error) {

	submissions, err := l.svcCtx.SubmissionModel.FindByUserIdAndAssignmentID(l.ctx, in.UserId, in.AssignmentID)
	if err != nil {
		// 查询出错，返回错误
		return nil, err
	}
	if len(submissions) == 0 {
		// 没有任何提交
		return &pb.GetMySubmissionStatusResp{
			Status: globalKey.NotSubmitted,
		}, nil
	}
	// 默认状态
	status := globalKey.Submitted
	for _, s := range submissions {
		if s.Status == globalKey.Reviewed {
			status = globalKey.Reviewed
			break // 有已批改的就优先返回
		}
	}
	return &pb.GetMySubmissionStatusResp{
		Status: status,
	}, nil
}
