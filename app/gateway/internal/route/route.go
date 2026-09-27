// Package route holds the gateway's path-prefix table and its matching logic.
//
// The table mirrors the prefix rules that used to live in the production
// Caddyfile, so routing decisions are versioned, reviewed and unit tested
// instead of being hand-edited on the server.
package route

import (
	"fmt"
	"strings"
)

// Rule maps a URL path prefix to a configured upstream name.
type Rule struct {
	Prefix   string
	Upstream string
}

// Default returns the production prefix table. Prefixes follow Caddy's `*`
// semantics (raw prefix match, not path-segment match), so "/api/v2/formula"
// still resolves to the form upstream.
func Default() []Rule {
	return []Rule{
		{Prefix: "/api/v2/auth", Upstream: "auth"},
		{Prefix: "/api/v2/users", Upstream: "user"},
		{Prefix: "/api/v2/task", Upstream: "task"},
		{Prefix: "/api/v2/review", Upstream: "review"},
		{Prefix: "/api/v2/user/test", Upstream: "exam"},
		{Prefix: "/api/v2/form", Upstream: "form"},
		{Prefix: "/api/v2/recruit", Upstream: "form"},
		{Prefix: "/api/v2/schedule", Upstream: "schedule"},
	}
}

// Match returns the upstream for the longest prefix matching path. Ties fall
// back to declaration order; two distinct prefixes of the same length cannot
// both match one path, so this only matters for duplicate rules.
func Match(rules []Rule, path string) (string, bool) {
	best := -1
	upstream := ""
	for _, rule := range rules {
		if !strings.HasPrefix(path, rule.Prefix) {
			continue
		}
		if len(rule.Prefix) > best {
			best = len(rule.Prefix)
			upstream = rule.Upstream
		}
	}
	if best < 0 {
		return "", false
	}
	return upstream, true
}

// Validate fails fast when a rule points at an upstream that is not configured
// or when two rules declare the same prefix.
func Validate(rules []Rule, upstreams []string) error {
	known := make(map[string]struct{}, len(upstreams))
	for _, name := range upstreams {
		if _, ok := known[name]; ok {
			return fmt.Errorf("duplicate upstream %q", name)
		}
		known[name] = struct{}{}
	}

	seen := make(map[string]struct{}, len(rules))
	for _, rule := range rules {
		if _, ok := known[rule.Upstream]; !ok {
			return fmt.Errorf("route prefix %q references unknown upstream %q", rule.Prefix, rule.Upstream)
		}
		if _, ok := seen[rule.Prefix]; ok {
			return fmt.Errorf("duplicate route prefix %q", rule.Prefix)
		}
		seen[rule.Prefix] = struct{}{}
	}

	return nil
}
