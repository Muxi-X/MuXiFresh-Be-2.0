package logic

import (
	"context"

	"MuXiFresh-Be-2.0/app/task/cmd/rpc/submission/internal/svc"
	"MuXiFresh-Be-2.0/app/task/cmd/rpc/submission/pb"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetAllSubmissionStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetAllSubmissionStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAllSubmissionStatusLogic {
	return &GetAllSubmissionStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetAllSubmissionStatusLogic) GetAllSubmissionStatus(in *pb.GetAllSubmissionStatusReq) (*pb.GetAllSubmissionStatusResp, error) {

	submissions, err := l.svcCtx.SubmissionModel.FindByAssignmentID(l.ctx, in.AssignmentID)
	if err != nil {
		return nil, err
	}
	var completions []*pb.Completion
	for _, submission := range submissions {
		userId := submission.UserId.String()[10:34]

		entryForm, err := l.svcCtx.EntryFormModel.FindOneByUserId(l.ctx, userId)
		if err != nil {
			return nil, err
		}
		userInfo, err := l.svcCtx.UserInfoModel.FindOne(l.ctx, userId)
		if err != nil {
			return nil, err
		}

		completions = append(completions, &pb.Completion{
			UserId:  entryForm.UserId.String()[10:34],
			Name:    userInfo.Name,
			Avatar:  entryForm.Avatar,
			Email:   userInfo.Email,
			Grade:   entryForm.Grade,
			School:  userInfo.School,
			Status:  submission.Status,
			Version: submission.VersionNum,
		})
	}
	return &pb.GetAllSubmissionStatusResp{
		Completions: completions,
	}, nil
}
