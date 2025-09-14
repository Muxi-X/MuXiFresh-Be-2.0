package logic

import (
	"MuXiFresh-Be-2.0/app/review/cmd/api/internal/svc"
	"MuXiFresh-Be-2.0/app/review/cmd/api/internal/types"
	"MuXiFresh-Be-2.0/app/user/cmd/rpc/user/userclient"
	"MuXiFresh-Be-2.0/common/ctxData"
	"MuXiFresh-Be-2.0/common/globalKey"
	"context"
	"errors"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetReviewLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetReviewLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetReviewLogic {
	return &GetReviewLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetReviewLogic) GetReview(req *types.GetReviewReq) (resp *types.GetReviewResp, err error) {
	//管理员认证
	getUserTypeResp, err := l.svcCtx.UserClient.GetUserType(l.ctx, &userclient.GetUserTypeReq{
		UserId: ctxData.GetUserIdFromCtx(l.ctx),
	})
	if err != nil {
		return nil, err
	}
	if getUserTypeResp.UserType != globalKey.Admin && getUserTypeResp.UserType != globalKey.SuperAdmin {
		return nil, errors.New("permission denied")
	}
	//秋招
	startTime := time.Date(req.Year, time.July, 1, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(req.Year, time.December, 31, 23, 59, 59, 999999999, time.UTC)

	if req.Season == "spring" {
		startTime = time.Date(req.Year, time.January, 1, 0, 0, 0, 0, time.UTC)
		endTime = time.Date(req.Year, time.June, 31, 23, 59, 59, 999999999, time.UTC)
	}

	entryForms, err := l.svcCtx.EntryFormModel.FindByGroup(l.ctx, req.Group, req.School, req.Grade, startTime, endTime)

	if err != nil {
		return nil, err
	}
	var rows []types.Row
	for _, entryForm := range entryForms {

		userId := entryForm.UserId.String()[10:34]
		//录取进度
		schedule, err := l.svcCtx.ScheduleClient.FindOneByUserId(l.ctx, userId)
		if err != nil {
			return nil, err
		}

		if req.Status != "" && schedule.AdmissionStatus != req.Status {
			continue
		}
		//测验情况
		userInfo, err := l.svcCtx.UserInfoModel.FindOne(l.ctx, userId)
		if err != nil {
			return nil, err
		}

		examStatus := "已提交"

		rows = append(rows, types.Row{
			Name:            userInfo.Name,
			Grade:           entryForm.Grade,
			School:          userInfo.School,
			Group:           entryForm.Group,
			Gender:          entryForm.Gender,
			FormID:          entryForm.ID.String()[10:34],
			ExamStuatus:     examStatus,
			UserId:          userId,
			AdmissionStatus: schedule.AdmissionStatus,
			ScheduleID:      schedule.ID.String()[10:34],
		})
	}
	return &types.GetReviewResp{
		Rows: rows,
	}, nil
}
