package policy

import (
	"encoding/json"
	"net/http"

	"ap-controller-go/internal/config"
)

// =========================
// Builder
// =========================

func BuildRuntimePolicy(cfg *config.Config) RuntimePolicy {
	rp := RuntimePolicy{
		Controller: ControllerInfo{
			ID:   cfg.Controller.ID,
			Site: cfg.Controller.Site,
			Name: cfg.Controller.Name,
		},
		Roles:    map[string]RuntimeRole{},
		Profiles: map[string]RuntimeProfile{},
		Bypass: RuntimeBypass{
			Enabled:      cfg.Bypass.Enabled,
			EnforceOrder: cfg.Bypass.EnforceOrder,
			MacWhitelist: cfg.Bypass.MacWhitelist,
			IPWhitelist:  cfg.Bypass.IPWhitelist,
			Domains:      cfg.Bypass.Domains,
		},
	}

	// Roles
	for role, r := range cfg.Roles {
		rp.Roles[role] = RuntimeRole{
			Profile: r.Profile,
		}
	}

	// Profiles
	for name, p := range cfg.Profiles {
		rp.Profiles[name] = RuntimeProfile{
			VLAN:          p.VLAN,
			FirewallGroup: p.FirewallGroup,
			SessionTTL:    p.SessionTTL,
		}
	}

	// Version (filled later)
	rp.Version = BuildPolicyVersion(
		struct {
			Roles    any
			Profiles any
			Bypass   any
		}{
			rp.Roles,
			rp.Profiles,
			rp.Bypass,
		},
		cfg.Controller.Version,
	)

	return rp
}

// =========================
// HTTP Handler
// =========================

func RuntimeHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		policy := BuildRuntimePolicy(cfg)

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Policy-Version", policy.Version.Version)
		w.Header().Set("X-Policy-Checksum", policy.Version.Checksum)

		_ = json.NewEncoder(w).Encode(policy)
	}
}
