package model

import (
	"strconv"
	"strings"
	"time"
)

const (
	cycleSeasonSpring = "spring"
	cycleSeasonAutumn = "autumn"

	// DeadlineLayout 是报名截止时间对外的字符串格式，固定按东八区解析与格式化。
	DeadlineLayout = "2006-01-02 15:04:05"
)

// deadlineLocation 固定为东八区。用 FixedZone 而非 LoadLocation，避免运行环境
// 缺少 tzdata 时时区解析失败。
var deadlineLocation = time.FixedZone("CST", 8*3600)

// ParseCycle 解析届次字符串（YYYYautumn / YYYYspring）；非法格式返回 ok=false。
func ParseCycle(cycle string) (year int, season string, ok bool) {
	for _, s := range []string{cycleSeasonAutumn, cycleSeasonSpring} {
		if !strings.HasSuffix(cycle, s) {
			continue
		}
		num := strings.TrimSuffix(cycle, s)
		if len(num) != 4 {
			return 0, "", false
		}
		y, err := strconv.Atoi(num)
		if err != nil || y <= 0 {
			return 0, "", false
		}
		return y, s, true
	}
	return 0, "", false
}

// DefaultDeadlineForCycle 返回届次的默认报名截止时间。
// 秋招默认该年 10-06 23:59:00（东八区）；春招无默认值，返回 ok=false。
func DefaultDeadlineForCycle(cycle string) (time.Time, bool) {
	year, season, ok := ParseCycle(cycle)
	if !ok || season != cycleSeasonAutumn {
		return time.Time{}, false
	}
	return time.Date(year, time.October, 6, 23, 59, 0, 0, deadlineLocation), true
}

// ParseDeadline 按 DeadlineLayout（东八区）解析截止时间字符串。
func ParseDeadline(s string) (time.Time, error) {
	return time.ParseInLocation(DeadlineLayout, s, deadlineLocation)
}

// FormatDeadline 按 DeadlineLayout（东八区）格式化截止时间。
func FormatDeadline(t time.Time) string {
	return t.In(deadlineLocation).Format(DeadlineLayout)
}
