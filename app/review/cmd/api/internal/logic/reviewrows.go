package logic

import (
	"context"
	"strings"
	"time"

	"MuXiFresh-Be-2.0/app/review/cmd/api/internal/svc"
	"MuXiFresh-Be-2.0/app/review/cmd/api/internal/types"
	scheduleModel "MuXiFresh-Be-2.0/app/schedule/model"
	userauthModel "MuXiFresh-Be-2.0/app/userauth/model"

	"github.com/zeromicro/go-zero/core/logx"
)

// groupAll 是请求中表示"不按组过滤（全量）"的显式枚举值。
// 依赖 form 模块 FindByGroup 的空串分支（group == "" 时不追加过滤条件）；
// 若该分支语义变化，这里必须同步，否则查询会静默返回空结果。
const groupAll = "All"

// isGroupAll 报告 group 是否为全量哨兵。查询过滤与导出拆表共用此判断，
// 避免同一哨兵在多处各自比较而漂移。
func isGroupAll(group string) bool {
	return group == groupAll
}

// groupFilter 把全量哨兵归一为空串，交给 FindByGroup 走全量分支。
func groupFilter(group string) string {
	if isGroupAll(group) {
		return ""
	}
	return group
}

// filterByName 按姓名做大小写不敏感的模糊匹配（rune 对齐子串包含）。
// keyword 为空时原样返回，供"未搜索"场景直接复用。
func filterByName(rows []types.Row, keyword string) []types.Row {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return rows
	}

	filtered := make([]types.Row, 0, len(rows))
	for _, row := range rows {
		if containsFold(row.Name, keyword) {
			filtered = append(filtered, row)
		}
	}

	return filtered
}

// containsFold 报告 name 是否包含 keyword，按 rune 对齐并做大小写折叠比较。
// 用 EqualFold 而非 ToLower：后者对部分 Unicode 并不等价（如 Σ 与 ς），
// 会把本应命中的姓名漏掉。
func containsFold(name, keyword string) bool {
	keywordRunes := []rune(keyword)
	nameRunes := []rune(name)
	for i := 0; i+len(keywordRunes) <= len(nameRunes); i++ {
		if strings.EqualFold(string(nameRunes[i:i+len(keywordRunes)]), keyword) {
			return true
		}
	}
	return false
}

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
			logx.WithContext(ctx).Errorf("buildReviewRows: missing schedule for user %s, skip", userId)
			continue
		}
		if status != "" && schedule.AdmissionStatus != status {
			continue
		}

		userInfo := userInfoMap[userId]
		if userInfo == nil {
			logx.WithContext(ctx).Errorf("buildReviewRows: missing userinfo for user %s, skip", userId)
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
