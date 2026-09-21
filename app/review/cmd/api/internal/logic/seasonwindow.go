package logic

import "time"

// seasonWindow 返回某届次的查询时间窗口 [start, end]。
//
// 与 app/form/model 的 CycleOf 分界严格对齐：秋招 7/1 00:00 起，春招至 6/30 止。
// 注意 6 月只有 30 天，写 June 31 会被 Go 归一为 7 月 1 日，与秋招窗口重叠，
// 跨届多表时同一用户会在相邻两届各出现一次。GetReview 与导出共用此函数，
// 避免两处各自计算而漂移。
func seasonWindow(year int, season string) (time.Time, time.Time) {
	const lastNano = 999999999

	switch season {
	case "spring":
		return time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC),
			time.Date(year, time.June, 30, 23, 59, 59, lastNano, time.UTC)
	case "":
		return time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC),
			time.Date(year, time.December, 31, 23, 59, 59, lastNano, time.UTC)
	default: // autumn
		return time.Date(year, time.July, 1, 0, 0, 0, 0, time.UTC),
			time.Date(year, time.December, 31, 23, 59, 59, lastNano, time.UTC)
	}
}
