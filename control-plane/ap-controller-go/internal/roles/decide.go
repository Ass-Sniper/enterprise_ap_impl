package roles

import (
	"path"
	"sort"
	"strings"

	"ap-controller-go/internal/config"
)

func norm(s string) string { return strings.TrimSpace(s) }

// match supports:
// - exact string
// - wildcard "*" / "?" via path.Match
// - list in YAML (decoded as []any)
func match(pattern any, value string) bool {
	if pattern == nil {
		return true
	}
	value = norm(value)
	if value == "" {
		return false
	}

	switch p := pattern.(type) {
	case string:
		ps := norm(p)
		if ps == "" {
			return true
		}
		if strings.ContainsAny(ps, "*?") {
			ok, _ := path.Match(ps, value)
			return ok
		}
		return ps == value
	case []any:
		for _, it := range p {
			if match(it, value) {
				return true
			}
		}
		return false
	default:
		// unknown type: treat as mismatch
		return false
	}
}

func matchWhen(when map[string]any, ctx map[string]string) bool {
	for _, f := range []string{"ssid", "auth", "ap_id", "radio_id", "mac", "ip"} {
		if _, ok := when[f]; ok {
			if !match(when[f], ctx[f]) {
				return false
			}
		}
	}
	return true
}

func DecideRole(cfg *config.Config, ctx map[string]string, defaultRole string) Decision {
	rules := append([]config.RoleRule{}, cfg.RoleRules...)
	sort.SliceStable(rules, func(i, j int) bool {
		if rules[i].Priority == rules[j].Priority {
			return rules[i].Name < rules[j].Name
		}
		return rules[i].Priority < rules[j].Priority // smaller = higher priority
	})

	for _, r := range rules {
		if matchWhen(r.When, ctx) {
			role := r.Assign
			if role == "" {
				role = defaultRole
			}
			return Decision{Role: role, MatchedRule: r.Name, Priority: r.Priority}
		}
	}
	return Decision{Role: defaultRole}
}
