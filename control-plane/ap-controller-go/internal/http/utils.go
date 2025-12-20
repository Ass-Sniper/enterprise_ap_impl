package httpapi

import (
	"ap-controller-go/internal/audit"
	"ap-controller-go/internal/config"
	"ap-controller-go/internal/security"
	"ap-controller-go/internal/store"
	"encoding/json"
	"net/http"
	"strings"
)

func New(
	cfg *config.Config,
	st *store.Store,
	aud *audit.Logger,
	pv string,
	jwtIssuer *security.JWTIssuer,
) *Server {
	return &Server{
		cfg:           cfg,
		st:            st,
		audit:         aud,
		policyVersion: pv,
		jwtIssuer:     jwtIssuer,
	}
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func macNorm(m string) string {
	return strings.ToLower(strings.TrimSpace(m))
}

// -------------------------------------------------------------------
// Response Helpers
// -------------------------------------------------------------------

func (s *Server) buildSessionResp(sess *store.SessionV2, ttl int) map[string]any {
	role := sess.Role
	roleDef, ok := s.cfg.Roles[role]

	profileName := sess.Profile
	if ok && profileName == "" {
		profileName = roleDef.Profile
	}
	profile := s.cfg.Profiles[profileName]

	return map[string]any{
		"authorized":     true,
		"role":           role,
		"ttl":            ttl,
		"policy_version": sess.PolicyVersion,
		"profile": map[string]any{
			"name":           profileName,
			"vlan":           profile.VLAN,
			"firewall_group": profile.FirewallGroup,
		},
	}
}
