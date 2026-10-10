package logic

import (
	"testing"

	"MuXiFresh-Be-2.0/app/review/cmd/api/internal/types"
)

func TestBuildExportSheets(t *testing.T) {
	rows := []types.Row{
		{Name: "a", Group: "Product"},
		{Name: "b", Group: "Design"},
		{Name: "c", Group: "Product"},
	}

	t.Run("empty group exports all groups with leading 全部 sheet", func(t *testing.T) {
		sheets := buildExportSheets("", rows)
		if len(sheets) != len(groupNames)+1 {
			t.Fatalf("got %d sheets, want %d", len(sheets), len(groupNames)+1)
		}
		if sheets[0].name != "全部" {
			t.Fatalf("first sheet = %q, want 全部", sheets[0].name)
		}
		if len(sheets[0].rows) != len(rows) {
			t.Fatalf("全部 sheet has %d rows, want %d", len(sheets[0].rows), len(rows))
		}
		for i, g := range groupNames {
			if sheets[i+1].name != g.cn {
				t.Fatalf("sheet[%d] = %q, want %q", i+1, sheets[i+1].name, g.cn)
			}
		}
	})

	t.Run("All group behaves like empty", func(t *testing.T) {
		sheets := buildExportSheets("All", rows)
		if len(sheets) != len(groupNames)+1 || sheets[0].name != "全部" {
			t.Fatalf("All group should export all groups with leading 全部, got %+v", sheets)
		}
	})

	t.Run("specific group has no 全部 sheet", func(t *testing.T) {
		sheets := buildExportSheets("Product", rows)
		if len(sheets) != 1 || sheets[0].name != "产品组" {
			t.Fatalf("got %+v, want single 产品组 sheet", sheets)
		}
		if len(sheets[0].rows) != 2 {
			t.Fatalf("产品组 sheet has %d rows, want 2", len(sheets[0].rows))
		}
	})

	t.Run("unknown group falls back to per-group sheets without 全部", func(t *testing.T) {
		sheets := buildExportSheets("Nope", rows)
		if len(sheets) != len(groupNames) {
			t.Fatalf("got %d sheets, want %d", len(sheets), len(groupNames))
		}
		for i, g := range groupNames {
			if sheets[i].name != g.cn {
				t.Fatalf("sheet[%d] = %q, want %q", i, sheets[i].name, g.cn)
			}
		}
	})

	t.Run("empty rows still yields 全部 plus empty group sheets", func(t *testing.T) {
		sheets := buildExportSheets("", nil)
		if len(sheets) != len(groupNames)+1 || sheets[0].name != "全部" {
			t.Fatalf("got %+v", sheets)
		}
		for i, s := range sheets {
			if len(s.rows) != 0 {
				t.Fatalf("sheet[%d] %q has %d rows, want 0", i, s.name, len(s.rows))
			}
		}
	})
}
