package logic

import (
	"testing"
	"time"
)

// 春招与秋招窗口不得重叠（历史上 June 31 归一为 7/1 造成重叠）
func TestSeasonWindow_NoOverlapBetweenSpringAndAutumn(t *testing.T) {
	_, springEnd := seasonWindow(2026, "spring")
	autumnStart, _ := seasonWindow(2026, "autumn")

	if !springEnd.Before(autumnStart) {
		t.Fatalf("spring end %v must be before autumn start %v", springEnd, autumnStart)
	}

	// 春招窗口最后一天必须仍是 6 月
	if springEnd.Month() != time.June {
		t.Fatalf("spring window must end in June, got %v", springEnd)
	}
}

func TestSeasonWindow_AutumnStartsJuly(t *testing.T) {
	start, _ := seasonWindow(2026, "autumn")
	if start.Month() != time.July || start.Day() != 1 {
		t.Fatalf("autumn should start Jul 1, got %v", start)
	}
}

// 全年窗口覆盖 1/1–12/31
func TestSeasonWindow_FullYear(t *testing.T) {
	start, end := seasonWindow(2026, "")
	if start.Month() != time.January || start.Day() != 1 {
		t.Fatalf("full-year start should be Jan 1, got %v", start)
	}
	if end.Month() != time.December || end.Day() != 31 {
		t.Fatalf("full-year end should be Dec 31, got %v", end)
	}
}

// 与 form.CycleOf 的分界一致：6/30 属春招，7/1 属秋招
func TestSeasonWindow_BoundaryMatchesCycle(t *testing.T) {
	_, springEnd := seasonWindow(2026, "spring")
	autumnStart, _ := seasonWindow(2026, "autumn")

	// 6/30 23:59:59.999 落在春招窗口内
	if !(springEnd.Month() == time.June && springEnd.Day() == 30) {
		t.Fatalf("expected spring end at Jun 30, got %v", springEnd)
	}
	// 7/1 00:00 落在秋招窗口内，且不在春招窗口内
	if !(autumnStart.Month() == time.July && autumnStart.Day() == 1) {
		t.Fatalf("expected autumn start at Jul 1, got %v", autumnStart)
	}
}
