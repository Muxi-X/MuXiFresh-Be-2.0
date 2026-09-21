package model

import (
	"testing"
	"time"
)

func TestCycleOf(t *testing.T) {
	cases := []struct {
		name string
		in   time.Time
		want string
	}{
		{"7月1日0时起为秋招", time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC), "2026autumn"},
		{"6月30日为春招", time.Date(2026, 6, 30, 23, 59, 59, 0, time.UTC), "2026spring"},
		{"报名高峰9月为秋招", time.Date(2026, 9, 20, 0, 0, 0, 0, time.UTC), "2026autumn"},
		{"1月为春招", time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), "2026spring"},
	}
	for _, c := range cases {
		if got := CycleOf(c.in); got != c.want {
			t.Errorf("%s: CycleOf(%v)=%q, want %q", c.name, c.in, got, c.want)
		}
	}
}

// 非 UTC 时间应按 UTC 归一到届次，避免服务器时区影响分界
func TestCycleOf_NormalizesToUTC(t *testing.T) {
	in := time.Date(2026, 7, 1, 3, 0, 0, 0, time.FixedZone("CST", 8*3600))
	if got := CycleOf(in); got != "2026spring" {
		t.Errorf("CycleOf(%v)=%q, want 2026spring (UTC 为 6/30)", in, got)
	}
}

func TestEffectiveCycle(t *testing.T) {
	// 显式 cycle 优先
	f := &EntryForm{Cycle: "2025autumn", CreateAt: time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)}
	if got := f.EffectiveCycle(); got != "2025autumn" {
		t.Errorf("explicit cycle should win, got %q", got)
	}

	// 存量数据无 cycle：按 createAt 推导
	legacy := &EntryForm{CreateAt: time.Date(2024, 9, 22, 0, 0, 0, 0, time.UTC)}
	if got := legacy.EffectiveCycle(); got != "2024autumn" {
		t.Errorf("legacy form should derive cycle from createAt, got %q", got)
	}

	// createAt 缺失的脏数据落到极早届次，等价于往届，不会与当届混淆
	dirty := &EntryForm{}
	if got := dirty.EffectiveCycle(); got == CycleOf(time.Now()) {
		t.Errorf("form without cycle/createAt must not be treated as current cycle, got %q", got)
	}
}
