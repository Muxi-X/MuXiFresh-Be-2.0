package model

import "testing"

func TestReverseLogs(t *testing.T) {
	revs := func(logs []*InterviewCommentLog) []int64 {
		out := make([]int64, len(logs))
		for i, l := range logs {
			out[i] = l.Rev
		}
		return out
	}

	cases := []struct {
		name string
		in   []int64
		want []int64
	}{
		{"空", nil, nil},
		{"单元素", []int64{1}, []int64{1}},
		{"偶数", []int64{1, 2, 3, 4}, []int64{4, 3, 2, 1}},
		{"奇数", []int64{1, 2, 3}, []int64{3, 2, 1}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			logs := make([]*InterviewCommentLog, len(c.in))
			for i, r := range c.in {
				logs[i] = &InterviewCommentLog{Rev: r}
			}
			reverseLogs(logs)
			got := revs(logs)
			if len(got) != len(c.want) {
				t.Fatalf("got %v, want %v", got, c.want)
			}
			for i := range got {
				if got[i] != c.want[i] {
					t.Fatalf("got %v, want %v", got, c.want)
				}
			}
		})
	}
}
