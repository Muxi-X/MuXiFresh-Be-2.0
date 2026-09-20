package logic

import (
	"testing"

	"MuXiFresh-Be-2.0/app/review/cmd/api/internal/types"
)

func TestGroupFilter(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{groupAll, ""},
		{"", ""},
		{"Product", "Product"},
		{"Backend", "Backend"},
		{"Nope", "Nope"},
	}
	for _, c := range cases {
		if got := groupFilter(c.in); got != c.want {
			t.Errorf("groupFilter(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestIsGroupAll(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{groupAll, true},
		{"", false},
		{"Product", false},
		{"Nope", false},
	}
	for _, c := range cases {
		if got := isGroupAll(c.in); got != c.want {
			t.Errorf("isGroupAll(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestContainsFold(t *testing.T) {
	cases := []struct {
		name    string
		keyword string
		want    bool
	}{
		{"any", "", true},
		{"", "", true},
		{"", "x", false},
		{"ab", "abc", false},
		{"abc", "abc", true},
		{"张三丰", "三", true},
		{"ΟΔΥΣΣΕΑΣ", "ςεα", true},
		{"ΟΔΥΣΣΕΑΣ", "O", false},
	}
	for _, c := range cases {
		if got := containsFold(c.name, c.keyword); got != c.want {
			t.Errorf("containsFold(%q, %q) = %v, want %v", c.name, c.keyword, got, c.want)
		}
	}
}

func TestFilterByName(t *testing.T) {
	rows := []types.Row{
		{Name: "张三"},
		{Name: "李四"},
		{Name: "Alice"},
		{Name: "alicechen"},
		{Name: "ΟΔΥΣΣΕΑΣ"},
		{Name: ""},
	}

	cases := []struct {
		keyword string
		want    []string
	}{
		{"", []string{"张三", "李四", "Alice", "alicechen", "ΟΔΥΣΣΕΑΣ", ""}},
		{"  ", []string{"张三", "李四", "Alice", "alicechen", "ΟΔΥΣΣΕΑΣ", ""}},
		{"张", []string{"张三"}},
		{"Alice", []string{"Alice", "alicechen"}},
		{"alice", []string{"Alice", "alicechen"}},
		{"ALICE", []string{"Alice", "alicechen"}},
		{"ςεα", []string{"ΟΔΥΣΣΕΑΣ"}},
		{"无此人", nil},
	}
	for _, c := range cases {
		got := filterByName(rows, c.keyword)
		if len(got) != len(c.want) {
			t.Errorf("filterByName(%q) returned %d rows, want %d", c.keyword, len(got), len(c.want))
			continue
		}
		for i := range got {
			if got[i].Name != c.want[i] {
				t.Errorf("filterByName(%q)[%d] = %q, want %q", c.keyword, i, got[i].Name, c.want[i])
			}
		}
	}
}
