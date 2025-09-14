package logic

import (
	"MuXiFresh-Be-2.0/app/task/model"
	"MuXiFresh-Be-2.0/common/globalKey"
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"

	"MuXiFresh-Be-2.0/app/task/cmd/rpc/submission/internal/svc"
	"MuXiFresh-Be-2.0/app/task/cmd/rpc/submission/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetSubmissionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetSubmissionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetSubmissionLogic {
	return &SetSubmissionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SetSubmissionLogic) SetSubmission(in *pb.SetSubmissionReq) (*pb.SetSubmissionResp, error) {

	userId, err := primitive.ObjectIDFromHex(in.UserId)
	if err != nil {
		return nil, err
	}

	assignmentID, err := primitive.ObjectIDFromHex(in.AssignmentID)
	if err != nil {
		return nil, err
	}

	assignment, err := l.svcCtx.AssignmentModel.FindOne(l.ctx, in.AssignmentID)
	if err != nil {
		return nil, err
	}
	dl, err := time.ParseInLocation("2006-01-02 15:04:05", assignment.Deadline, time.Local)
	if err != nil {
		return nil, err
	}
	fmt.Println("time", dl)
	if !dl.IsZero() && time.Now().After(dl) {
		return &pb.SetSubmissionResp{
			Flag: false,
		}, nil
	}

	count, err := l.svcCtx.SubmissionModel.CountByUserAndAssignment(l.ctx, in.UserId, in.AssignmentID)
	if err != nil {
		return nil, err
	}

	newSubmission := &model.Submission{
		UserId:       userId,
		AssignmentID: assignmentID,
		Urls:         in.Urls,
		Status:       globalKey.WaitComment,
		Version:      count + 1, // 新版本号 = 历史提交数量 + 1
		CreateAt:     time.Now(),
		UpdateAt:     time.Now(),
	}
	if err := l.svcCtx.SubmissionModel.Insert(l.ctx, newSubmission); err != nil {
		return nil, err
	}
	return &pb.SetSubmissionResp{
		Flag: true,
	}, nil
}
