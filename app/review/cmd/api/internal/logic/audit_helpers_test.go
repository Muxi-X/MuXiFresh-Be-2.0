package logic

import (
	"testing"
	"time"
)

func TestFormatAuditTime(t *testing.T) {
	if got := formatAuditTime(time.Time{}); got != "" {
		t.Fatalf("zero time should format empty, got %q", got)
	}

	cst := time.Date(2026, time.October, 5, 14, 3, 7, 0, time.FixedZone("CST", 8*3600))
	if got := formatAuditTime(cst); got != "2026-10-05 14:03:07" {
		t.Fatalf("formatAuditTime = %q, want 2026-10-05 14:03:07", got)
	}

	// 存储为 UTC 的时间应换算到东八区展示
	utc := time.Date(2026, time.October, 5, 6, 3, 7, 0, time.UTC)
	if got := formatAuditTime(utc); got != "2026-10-05 14:03:07" {
		t.Fatalf("UTC time should render in UTC+8, got %q", got)
	}
}
