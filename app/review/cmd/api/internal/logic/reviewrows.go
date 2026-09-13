package logic

import (
	"context"
	"time"

	"MuXiFresh-Be-2.0/app/review/cmd/api/internal/svc"
	"MuXiFresh-Be-2.0/app/review/cmd/api/internal/types"
	scheduleModel "MuXiFresh-Be-2.0/app/schedule/model"
	userauthModel "MuXiFresh-Be-2.0/app/userauth/model"

	"github.com/zeromicro/go-zero/core/logx"
)

// buildReviewRows 查询报名表，并批量补充进度(schedule)与个人信息(userinfo)，
// 避免逐条查询导致的 N+1 问题。缺少 schedule/userinfo 的记录会被跳过并记录日志。
func buildReviewRows(ctx context.Context, svcCtx *svc.ServiceContext, group, school, grade, status string, startTime, endTime time.Time) ([]types.Row, error) {
	entryForms, err := svcCtx.EntryFormModel.FindByGroup(ctx, group, school, grade, startTime, endTime)
	if err != nil {
		return nil, err
	}
	if len(entryForms) == 0 {
		return []types.Row{}, nil
	}

	userIds := make([]string, 0, len(entryForms))
	for _, entryForm := range entryForms {
		userIds = append(userIds, entryForm.UserId.Hex())
	}

	schedules, err := svcCtx.ScheduleClient.FindByUserIds(ctx, userIds)
	if err != nil {
		return nil, err
	}
	userInfos, err := svcCtx.UserInfoModel.FindByUserIds(ctx, userIds)
	if err != nil {
		return nil, err
	}

	scheduleMap := make(map[string]*scheduleModel.Schedule, len(schedules))
	for _, schedule := range schedules {
		scheduleMap[schedule.UserID.Hex()] = schedule
	}
	userInfoMap := make(map[string]*userauthModel.UserInfo, len(userInfos))
	for _, userInfo := range userInfos {
		userInfoMap[userInfo.ID.Hex()] = userInfo
	}

	rows := make([]types.Row, 0, len(entryForms))
	for _, entryForm := range entryForms {
		userId := entryForm.UserId.Hex()

		schedule := scheduleMap[userId]
		if schedule == nil {
			logx.WithContext(ctx).Infof("buildReviewRows: missing schedule for user %s", userId)
			continue
		}
		if status != "" && schedule.AdmissionStatus != status {
			continue
		}

		userInfo := userInfoMap[userId]
		if userInfo == nil {
			logx.WithContext(ctx).Infof("buildReviewRows: missing userinfo for user %s", userId)
			continue
		}

		rows = append(rows, types.Row{
			Name:            userInfo.Name,
			Grade:           entryForm.Grade,
			School:          userInfo.School,
			Group:           entryForm.Group,
			Gender:          entryForm.Gender,
			Major:           entryForm.Major,
			Phone:           entryForm.Phone,
			QQ:              userInfo.QQ,
			FormID:          entryForm.ID.Hex(),
			UserId:          userId,
			AdmissionStatus: schedule.AdmissionStatus,
			ScheduleID:      schedule.ID.Hex(),
			Understanding:   entryForm.Knowledge,
			Reason:          entryForm.Reason,
			SelfIntro:       entryForm.SelfIntro,
			ExtraQuestion:   entryForm.ExtraQuestion,
		})
	}

	return rows, nil
}
