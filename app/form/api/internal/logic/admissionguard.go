package logic

import (
	"context"
	"errors"

	scheduleModel "MuXiFresh-Be-2.0/app/schedule/model"
	"MuXiFresh-Be-2.0/common/globalKey"
)

// ensureNotAdmittedMember 拒绝已录取成员（实习期/已转正）再次提交或重报报名表。
//
// 报名表是面向新生的入口：已录取成员重报会混入当届审阅名单，其录取状态也不应
// 被重新报名覆盖。无进度记录视为尚未报名的用户，放行。
func ensureNotAdmittedMember(ctx context.Context, schedules scheduleModel.ScheduleModel, userId string) error {
	schedule, err := schedules.FindOneByUserId(ctx, userId)
	if err != nil {
		if errors.Is(err, scheduleModel.ErrNotFound) {
			return nil
		}
		return err
	}
	switch schedule.AdmissionStatus {
	case globalKey.Internship, globalKey.Formal:
		return errors.New("已是正式成员，无需重复报名")
	default:
		return nil
	}
}
