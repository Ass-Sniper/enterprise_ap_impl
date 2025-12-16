package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"ap-controller-go/internal/audit"
	"ap-controller-go/internal/config"
	"ap-controller-go/internal/policy"
	"ap-controller-go/internal/roles"
	"ap-controller-go/internal/store"

	"github.com/go-chi/chi/v5"
)

func New(cfg *config.Config, st *store.Store, audit *audit.Logger, policyVersion int) *Server {
	return &Server{cfg: cfg, st: st, audit: audit, policyVersion: policyVersion}
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func macNorm(m string) string {
	return strings.ToLower(strings.TrimSpace(m))
}

func (s *Server) buildSessionResp(sess *store.SessionV2, ttl int) map[string]any {
	// role -> profile -> attrs
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

func (s *Server) Router() http.Handler {
	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]any{
			"status": "ok",
		})
	})

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		err := s.st.Ping(ctx)
		writeJSON(w, 200, map[string]any{
			"status":     "ok",
			"redis_ping": err == nil,
		})
	})

	r.Post("/portal/login", s.portalLogin)
	r.Post("/portal/heartbeat", s.portalHeartbeat)
	r.Post("/portal/logout", s.portalLogout)

	r.Get("/portal/status/{mac}", s.portalStatus)
	r.Post("/portal/batch_status", s.portalBatchStatus)

	r.Get("/api/v1/policy/runtime", policy.RuntimeHandler(s.cfg))

	return r
}

func (s *Server) portalLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req LoginReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, 400, map[string]any{"authorized": false, "error": "bad_json"})
		return
	}
	req.MAC = macNorm(req.MAC)
	if req.MAC == "" {
		writeJSON(w, 422, map[string]any{"authorized": false, "error": "mac_required"})
		return
	}

	// decide role by rules
	decision := roles.DecideRole(s.cfg, map[string]string{
		"mac":      req.MAC,
		"ssid":     req.SSID,
		"auth":     req.Auth,
		"ap_id":    req.APID,
		"radio_id": req.RadioID,
		"ip":       req.IP,
	}, "guest")

	role := decision.Role
	roleDef := s.cfg.Roles[role]
	profile := s.cfg.Profiles[roleDef.Profile]
	ttl := profile.SessionTTL

	sess := store.SessionV2{
		Schema:        2,
		MAC:           req.MAC,
		Role:          role,
		Profile:       roleDef.Profile,
		PolicyVersion: s.policyVersion,
	}
	sess.Rule.Name = decision.MatchedRule
	sess.Rule.Priority = decision.Priority
	sess.AP.APID = req.APID
	sess.AP.SSID = req.SSID
	sess.AP.RadioID = req.RadioID
	sess.Attrs.VLAN = profile.VLAN
	sess.Attrs.FirewallGroup = profile.FirewallGroup
	sess.Auth.Method = req.Auth
	sess.Auth.Source = req.Source

	_ = s.st.SetSession(ctx, sess, ttl)

	s.audit.Write(map[string]any{
		"event":          "portal.login",
		"mac":            req.MAC,
		"authorized":     true,
		"role":           role,
		"ttl":            ttl,
		"policy_version": s.policyVersion,
		"profile":        roleDef.Profile,
		"vlan":           profile.VLAN,
		"firewall_group": profile.FirewallGroup,
		"rule":           decision.MatchedRule,
		"ap_id":          req.APID,
		"ssid":           req.SSID,
		"radio_id":       req.RadioID,
		"source":         req.Source,
		"result":         "ok",
	})

	// return session resp
	sess2, ttl2, _ := s.st.GetSessionFull(ctx, req.MAC)
	if sess2 == nil {
		writeJSON(w, 200, map[string]any{"authorized": false})
		return
	}
	writeJSON(w, 200, s.buildSessionResp(sess2, ttl2))
}

func (s *Server) portalHeartbeat(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req HeartbeatReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, 400, map[string]any{"authorized": false, "error": "bad_json"})
		return
	}
	req.MAC = macNorm(req.MAC)
	if req.MAC == "" {
		writeJSON(w, 422, map[string]any{"authorized": false, "error": "mac_required"})
		return
	}

	sess, _, err := s.st.GetSessionFull(ctx, req.MAC)
	if err != nil || sess == nil {
		s.audit.Write(map[string]any{
			"event":      "portal.heartbeat",
			"mac":        req.MAC,
			"authorized": false,
			"source":     req.Source,
			"result":     "not_found",
		})
		writeJSON(w, 200, map[string]any{"authorized": false})
		return
	}

	// refresh with the profile TTL (not current ttl)
	roleDef := s.cfg.Roles[sess.Role]
	profile := s.cfg.Profiles[roleDef.Profile]
	ok, _ := s.st.Refresh(ctx, req.MAC, profile.SessionTTL)
	if !ok {
		s.audit.Write(map[string]any{
			"event":      "portal.heartbeat",
			"mac":        req.MAC,
			"authorized": false,
			"source":     req.Source,
			"result":     "expired_after_refresh",
		})
		writeJSON(w, 200, map[string]any{"authorized": false})
		return
	}

	// return refreshed ttl
	sess2, ttl2, _ := s.st.GetSessionFull(ctx, req.MAC)
	s.audit.Write(map[string]any{
		"event":          "portal.heartbeat",
		"mac":            req.MAC,
		"authorized":     true,
		"role":           sess.Role,
		"ttl":            ttl2,
		"policy_version": sess.PolicyVersion,
		"source":         req.Source,
		"result":         "ok",
	})
	writeJSON(w, 200, s.buildSessionResp(sess2, ttl2))
}

func (s *Server) portalLogout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req LogoutReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, 400, map[string]any{"authorized": false, "error": "bad_json"})
		return
	}
	req.MAC = macNorm(req.MAC)
	if req.MAC == "" {
		writeJSON(w, 422, map[string]any{"authorized": false, "error": "mac_required"})
		return
	}

	existed, _ := s.st.Delete(ctx, req.MAC)

	s.audit.Write(map[string]any{
		"event":      "portal.logout",
		"mac":        req.MAC,
		"authorized": false,
		"result":     map[bool]string{true: "ok", false: "not_found"}[existed],
	})

	writeJSON(w, 200, map[string]any{"authorized": false})
}

func (s *Server) portalStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	mac := macNorm(chi.URLParam(r, "mac"))
	sess, ttl, _ := s.st.GetSessionFull(ctx, mac)
	if sess == nil {
		writeJSON(w, 200, map[string]any{
			"authorized": false,
			"role":       nil,
			"ttl":        nil,
		})
		return
	}
	writeJSON(w, 200, s.buildSessionResp(sess, ttl))
}

func (s *Server) portalBatchStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req BatchReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, 400, map[string]any{"error": "bad_json"})
		return
	}

	type Item struct {
		MAC           string         `json:"mac"`
		Authorized    bool           `json:"authorized"`
		Role          *string        `json:"role,omitempty"`
		TTL           *int           `json:"ttl,omitempty"`
		PolicyVersion *int           `json:"policy_version,omitempty"`
		Profile       map[string]any `json:"profile,omitempty"`
	}

	out := make([]Item, 0, len(req.Entries))

	for _, e := range req.Entries {
		m := macNorm(e.MAC)
		sess, ttl, _ := s.st.GetSessionFull(ctx, m)
		if sess == nil {
			out = append(out, Item{MAC: m, Authorized: false})
			continue
		}
		role := sess.Role
		pv := sess.PolicyVersion
		ttl2 := ttl

		resp := s.buildSessionResp(sess, ttl2)
		// flatten into expected structure
		profile := resp["profile"].(map[string]any)

		out = append(out, Item{
			MAC:           m,
			Authorized:    true,
			Role:          &role,
			TTL:           &ttl2,
			PolicyVersion: &pv,
			Profile:       profile,
		})
	}

	writeJSON(w, 200, map[string]any{"results": out})
}

// helper: parse policy version string env, optional
func ParsePolicyVersion(s string) int {
	i, _ := strconv.Atoi(strings.TrimSpace(s))
	if i < 0 {
		return 0
	}
	return i
}
