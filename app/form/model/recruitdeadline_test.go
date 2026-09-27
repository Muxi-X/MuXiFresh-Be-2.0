package model

import (
	"testing"
	"time"
)

func TestParseCycle(t *testing.T) {
	cases := []struct {
		name   string
		in     string
		year   int
		season string
		ok     bool
	}{
		{"秋招届次", "2026autumn", 2026, "autumn", true},
		{"春招届次", "2026spring", 2026, "spring", true},
		{"缺年份", "autumn", 0, "", false},
		{"缺届次", "2026", 0, "", false},
		{"未知届次", "2026summer", 0, "", false},
		{"年份非数字", "abcdautumn", 0, "", false},
		{"年份非四位", "26autumn", 0, "", false},
		{"空串", "", 0, "", false},
	}
	for _, c := range cases {
		year, season, ok := ParseCycle(c.in)
		if year != c.year || season != c.season || ok != c.ok {
			t.Errorf("%s: ParseCycle(%q)=(%d,%q,%v), want (%d,%q,%v)",
				c.name, c.in, year, season, ok, c.year, c.season, c.ok)
		}
	}
}

func TestDefaultDeadlineForCycle(t *testing.T) {
	got, ok := DefaultDeadlineForCycle("2026autumn")
	if !ok {
		t.Fatal("autumn cycle should have a default deadline")
	}
	want := time.Date(2026, time.October, 6, 23, 59, 0, 0, deadlineLocation)
	if !got.Equal(want) {
		t.Fatalf("autumn default = %v, want %v", got, want)
	}

	if _, ok := DefaultDeadlineForCycle("2027spring"); ok {
		t.Fatal("spring cycle must have no default deadline")
	}
	if _, ok := DefaultDeadlineForCycle("bad"); ok {
		t.Fatal("invalid cycle must have no default deadline")
	}
}

func TestDeadlineRoundTrip(t *testing.T) {
	in := "2026-10-06 23:59:00"
	parsed, err := ParseDeadline(in)
	if err != nil {
		t.Fatalf("parse valid deadline failed: %v", err)
	}
	if got := FormatDeadline(parsed); got != in {
		t.Fatalf("round trip = %q, want %q", got, in)
	}
	if _, err := ParseDeadline("2026/10/06 23:59"); err == nil {
		t.Fatal("invalid format must be rejected")
	}
}

// 解析与格式化固定按东八区，避免服务器时区影响
func TestDeadlineLocationIsFixedCST(t *testing.T) {
	parsed, err := ParseDeadline("2026-10-06 23:59:00")
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if _, offset := parsed.Zone(); offset != 8*3600 {
		t.Fatalf("deadline must be in UTC+8, got offset %d", offset)
	}
}
