package logic

import (
	formmodel "MuXiFresh-Be-2.0/app/form/model"
	"MuXiFresh-Be-2.0/app/task/model"
	"MuXiFresh-Be-2.0/common/globalKey"
	"MuXiFresh-Be-2.0/common/tool"
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"strings"
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
	// 组别/身份校验（N-L3）：白名单制——
	// freshman 仅可提交本组作业；normal（管理员录取设置的老成员，可能无报名表）
	// /admin/super_admin 豁免；其余未知类型一律拒绝（fail closed）
	userInfo, err := l.svcCtx.UserInfoModel.FindOne(l.ctx, in.UserId)
	if err != nil {
		return nil, err
	}
	switch userInfo.UserType {
	case globalKey.Normal, globalKey.Admin, globalKey.SuperAdmin:
		// 豁免
	case globalKey.Freshman:
		entryForm, err := l.svcCtx.EntryFormModel.FindOneByUserId(l.ctx, in.UserId)
		if err != nil {
			if errors.Is(err, formmodel.ErrNotFound) {
				return nil, errors.New("尚未报名，无权提交作业")
			}
			return nil, err
		}
		if !strings.EqualFold(entryForm.Group, assignment.Group) {
			return nil, errors.New("无权提交该组的作业")
		}
	default:
		return nil, errors.New("无权提交作业")
	}

	dl, err := time.ParseInLocation("2006-01-02 15:04:05", assignment.Deadline, time.Local)
	if err != nil {
		return nil, err
	}
	if !dl.IsZero() && time.Now().After(dl) {
		return &pb.SetSubmissionResp{
			Flag: false,
		}, nil
	}

	count, err := l.svcCtx.SubmissionModel.CountByUserAndAssignment(l.ctx, in.UserId, in.AssignmentID)
	if err != nil {
		return nil, err
	}

	// 拒绝前端上传失败拼出的坏值（如 "undefined"），并清洗空项
	urls, err := tool.ValidateResourceURLs(in.Urls)
	if err != nil {
		return nil, err
	}

	newSubmission := &model.Submission{
		UserId:       userId,
		AssignmentID: assignmentID,
		Urls:         urls,
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
