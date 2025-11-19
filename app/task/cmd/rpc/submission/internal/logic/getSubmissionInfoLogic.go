package logic

import (
	"context"

	"MuXiFresh-Be-2.0/app/task/cmd/rpc/submission/internal/svc"
	"MuXiFresh-Be-2.0/app/task/cmd/rpc/submission/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetSubmissionInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetSubmissionInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSubmissionInfoLogic {
	return &GetSubmissionInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetSubmissionInfoLogic) GetSubmissionInfo(in *pb.GetSubmissionInfoReq) (*pb.GetSubmissionInfoResp, error) {

	submissions, err := l.svcCtx.SubmissionModel.FindByUserIdAndAssignmentID(l.ctx, in.UserId, in.AssignmentID)
	if err != nil {
		return nil, err
	}
	var submissionInfos []*pb.SubmissionInfo
	for _, submission := range submissions {
		submissionInfos = append(submissionInfos, &pb.SubmissionInfo{
			SubmissionID: submission.ID.String()[10:34],
			Urls:         submission.Urls,
		})
	}

	return &pb.GetSubmissionInfoResp{
		SubmissionInfos: submissionInfos,
	}, nil
}
