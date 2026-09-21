package logic

import (
	"context"
	"errors"

	scheduleModel "MuXiFresh-Be-2.0/app/schedule/model"
	"MuXiFresh-Be-2.0/common/globalKey"
)

// isAdmittedStatus 报告录取状态是否属于"已录取成员"（实习期/已转正）。
// 报名表是面向新生的入口，这两种状态不得被报名流程改写。
func isAdmittedStatus(status string) bool {
	return status == globalKey.Internship || status == globalKey.Formal
}

// ensureNotAdmittedMember 拒绝已录取成员（实习期/已转正）再次提交或重报报名表。
// 无进度记录视为尚未报名的用户，放行。
//
// 这是入口预检，用于给用户明确提示；写库时由 UpsertByUserId 的过滤条件兜底并发窗口。
func ensureNotAdmittedMember(ctx context.Context, schedules scheduleModel.ScheduleModel, userId string) error {
	schedule, err := schedules.FindOneByUserId(ctx, userId)
	if err != nil {
		if errors.Is(err, scheduleModel.ErrNotFound) {
			return nil
		}
		return err
	}
	if isAdmittedStatus(schedule.AdmissionStatus) {
		return errors.New("已是正式成员，无需重复报名")
	}
	return nil
}
