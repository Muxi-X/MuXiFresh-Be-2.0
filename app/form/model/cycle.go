package model

import (
	"fmt"
	"time"
)

// CycleOf 返回给定时间所属的招生届次，格式为 YYYYautumn / YYYYspring。
//
// 分界为 7 月 1 日 00:00 UTC：与 review 的查询窗口一致（秋招 7/1 起，春招至
// 6 月 30 日止），否则同一份报名表会同时落入相邻两届的 review 列表。改动此处
// 必须同步 app/review/cmd/api/internal/logic/getreviewlogic.go 的窗口计算。
func CycleOf(t time.Time) string {
	t = t.UTC()
	year := t.Year()
	season := "spring"
	if t.Month() >= time.July {
		season = "autumn"
	}
	return fmt.Sprintf("%d%s", year, season)
}

// EffectiveCycle 返回报名表的实际届次。
//
// 优先使用显式 cycle；存量数据（本次改动上线前创建）没有该字段，按 createAt
// 推导，使老数据无需迁移也能被正确判为当届或往届。createAt 缺失的脏数据会落到
// 极早年份，等价于往届，不会与当届误判为同一届。
func (f *EntryForm) EffectiveCycle() string {
	if f.Cycle != "" {
		return f.Cycle
	}
	return CycleOf(f.CreateAt)
}
