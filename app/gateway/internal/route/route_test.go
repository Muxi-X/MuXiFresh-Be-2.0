package route

import (
	"reflect"
	"testing"
)

func TestDefaultMatchesProductionPrefixes(t *testing.T) {
	want := []Rule{
		{Prefix: "/api/v2/auth", Upstream: "auth"},
		{Prefix: "/api/v2/users", Upstream: "user"},
		{Prefix: "/api/v2/task", Upstream: "task"},
		{Prefix: "/api/v2/review", Upstream: "review"},
		{Prefix: "/api/v2/user/test", Upstream: "exam"},
		{Prefix: "/api/v2/form", Upstream: "form"},
		{Prefix: "/api/v2/recruit", Upstream: "form"},
		{Prefix: "/api/v2/schedule", Upstream: "schedule"},
	}

	if got := Default(); !reflect.DeepEqual(got, want) {
		t.Fatalf("Default() = %#v, want %#v", got, want)
	}
}

func TestMatch(t *testing.T) {
	rules := Default()

	tests := []struct {
		path     string
		upstream string
		ok       bool
	}{
		{path: "/api/v2/auth/login", upstream: "auth", ok: true},
		{path: "/api/v2/auth", upstream: "auth", ok: true},
		{path: "/api/v2/users/info/abc", upstream: "user", ok: true},
		{path: "/api/v2/user/test", upstream: "exam", ok: true},
		{path: "/api/v2/user/test/result", upstream: "exam", ok: true},
		{path: "/api/v2/task/assigned/list/selected", upstream: "task", ok: true},
		{path: "/api/v2/review/export", upstream: "review", ok: true},
		{path: "/api/v2/form/view", upstream: "form", ok: true},
		{path: "/api/v2/recruit/deadline", upstream: "form", ok: true},
		{path: "/api/v2/schedule/create", upstream: "schedule", ok: true},
		// Caddy's `*` is a raw prefix match, not a segment match.
		{path: "/api/v2/formula", upstream: "form", ok: true},
		// intro is deliberately not exposed.
		{path: "/api/v2/intro", upstream: "", ok: false},
		{path: "/api/v2", upstream: "", ok: false},
		{path: "/health", upstream: "", ok: false},
		{path: "/", upstream: "", ok: false},
	}

	for _, tt := range tests {
		upstream, ok := Match(rules, tt.path)
		if ok != tt.ok || upstream != tt.upstream {
			t.Errorf("Match(%q) = (%q, %v), want (%q, %v)", tt.path, upstream, ok, tt.upstream, tt.ok)
		}
	}
}

func TestMatchPicksLongestPrefix(t *testing.T) {
	rules := []Rule{
		{Prefix: "/api/v2/user", Upstream: "user"},
		{Prefix: "/api/v2/user/test", Upstream: "exam"},
	}

	upstream, ok := Match(rules, "/api/v2/user/test/result")
	if !ok || upstream != "exam" {
		t.Fatalf("Match() = (%q, %v), want (exam, true)", upstream, ok)
	}
}

func TestValidate(t *testing.T) {
	rules := []Rule{{Prefix: "/api/v2/auth", Upstream: "auth"}}

	if err := Validate(rules, []string{"auth"}); err != nil {
		t.Fatalf("Validate(known upstream) = %v, want nil", err)
	}
	if err := Validate(rules, []string{"user"}); err == nil {
		t.Fatal("Validate(unknown upstream) = nil, want error")
	}
}

func TestValidateRejectsDuplicatePrefix(t *testing.T) {
	rules := []Rule{
		{Prefix: "/api/v2/auth", Upstream: "auth"},
		{Prefix: "/api/v2/auth", Upstream: "user"},
	}

	if err := Validate(rules, []string{"auth", "user"}); err == nil {
		t.Fatal("Validate(duplicate prefix) = nil, want error")
	}
}

func TestValidateRejectsDuplicateUpstream(t *testing.T) {
	rules := []Rule{{Prefix: "/api/v2/auth", Upstream: "auth"}}

	if err := Validate(rules, []string{"auth", "auth"}); err == nil {
		t.Fatal("Validate(duplicate upstream) = nil, want error")
	}
}
