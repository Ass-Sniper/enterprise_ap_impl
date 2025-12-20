package httpapi

import (
	"encoding/json"
	"net/http"

	"ap-controller-go/internal/policy"
	"ap-controller-go/internal/roles"
	"ap-controller-go/internal/security"
	"ap-controller-go/internal/store"

	"github.com/go-chi/chi/v5"
)

func (s *Server) Router() http.Handler {
	r := chi.NewRouter()

	registerSwagger(r)

	// ========================
	// Public endpoints
	// ========================
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]any{"status": "ok"})
	})

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		err := s.st.Ping(r.Context())
		writeJSON(w, 200, map[string]any{
			"status":     "ok",
			"redis_ping": err == nil,
		})
	})

	// ========================
	// Portal login (NO HMAC)
	// ========================
	r.Post("/portal/login", s.portalLogin)

	// ========================
	// Protected APIs (HMAC required)
	// ========================
	r.Route("/", func(pr chi.Router) {
		// 🔐 强制 HMAC 校验
		pr.Use(security.PortalAuthMiddleware(s.st))

		// 🔑 新增：auth_request 专用 verify
		pr.Post("/portal/context/verify", s.portalContextVerify)

		// Portal APIs (post-login)
		pr.Post("/portal/heartbeat", s.portalHeartbeat)
		pr.Post("/portal/logout", s.portalLogout)

		// Ops APIs
		pr.Get("/portal/status/{mac}", s.portalStatus)
		pr.Post("/portal/batch_status", s.portalBatchStatus)

		// Policy runtime
		pr.Get("/api/v1/policy/runtime", policy.RuntimeHandler(s.cfg))
	})

	return r
}

// -------------------------------------------------------------------
// Portal Handlers (Context-based)
// -------------------------------------------------------------------

// portalLogin handles portal login with trusted context.
func (s *Server) portalLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req PortalContextReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, 400, map[string]any{"authorized": false, "error": "bad_json"})
		return
	}

	mac := macNorm(req.Client.MAC)
	if mac == "" {
		writeJSON(w, 422, map[string]any{"authorized": false, "error": "mac_required"})
		return
	}

	decision := roles.DecideRole(s.cfg, map[string]string{
		"mac":      mac,
		"ssid":     req.Wireless.SSID,
		"ap_id":    req.Access.APID,
		"radio_id": req.Wireless.RadioID,
		"ip":       req.Client.IP,
		"os":       req.Client.OS,
	}, "guest")

	role := decision.Role
	roleDef := s.cfg.Roles[role]
	profile := s.cfg.Profiles[roleDef.Profile]
	ttl := profile.SessionTTL

	sess := store.SessionV2{
		Schema:        2,
		MAC:           mac,
		Role:          role,
		Profile:       roleDef.Profile,
		PolicyVersion: s.policyVersion,
	}

	sess.Rule.Name = decision.MatchedRule
	sess.Rule.Priority = decision.Priority
	sess.AP.APID = req.Access.APID
	sess.AP.SSID = req.Wireless.SSID
	sess.AP.RadioID = req.Wireless.RadioID
	sess.Attrs.VLAN = profile.VLAN
	sess.Attrs.FirewallGroup = profile.FirewallGroup
	sess.Auth.Method = "portal"
	sess.Auth.Source = req.Meta.Source

	_ = s.st.SetSession(ctx, sess, ttl)

	s.audit.Write(map[string]any{
		"event":      "portal.login",
		"mac":        mac,
		"role":       role,
		"ttl":        ttl,
		"rule":       decision.MatchedRule,
		"ap_id":      req.Access.APID,
		"ssid":       req.Wireless.SSID,
		"radio_id":   req.Wireless.RadioID,
		"source":     req.Meta.Source,
		"policy_ver": s.policyVersion,
		"result":     "ok",
	})

	sess2, ttl2, _ := s.st.GetSessionFull(ctx, mac)
	if sess2 == nil {
		writeJSON(w, 200, map[string]any{"authorized": false})
		return
	}

	// issue JWT (NEW)
	token, exp, err := s.jwtIssuer.Issue(ctx, mac)
	if err != nil {
		writeJSON(w, 500, map[string]any{
			"authorized": false,
			"error":      "issue_token_failed",
		})
		return
	}

	writeJSON(w, 200, map[string]any{
		"authorized": true,
		"session":    s.buildSessionResp(sess2, ttl2),
		"token": map[string]any{
			"access_token": token,
			"expires_in":   exp,
			"token_type":   "Bearer",
		},
	})

}

// portalHeartbeat refreshes session TTL using context.
func (s *Server) portalHeartbeat(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req PortalContextReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, 400, map[string]any{"authorized": false, "error": "bad_json"})
		return
	}

	mac := macNorm(req.Client.MAC)
	if mac == "" {
		writeJSON(w, 422, map[string]any{"authorized": false, "error": "mac_required"})
		return
	}

	sess, _, err := s.st.GetSessionFull(ctx, mac)
	if err != nil || sess == nil {
		writeJSON(w, 200, map[string]any{"authorized": false})
		return
	}

	roleDef := s.cfg.Roles[sess.Role]
	profile := s.cfg.Profiles[roleDef.Profile]

	ok, _ := s.st.Refresh(ctx, mac, profile.SessionTTL)
	if !ok {
		writeJSON(w, 200, map[string]any{"authorized": false})
		return
	}

	sess2, ttl2, _ := s.st.GetSessionFull(ctx, mac)
	writeJSON(w, 200, s.buildSessionResp(sess2, ttl2))
}

// portalLogout deletes session.
func (s *Server) portalLogout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req PortalContextReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, 400, map[string]any{"authorized": false, "error": "bad_json"})
		return
	}

	mac := macNorm(req.Client.MAC)
	if mac == "" {
		writeJSON(w, 422, map[string]any{"authorized": false, "error": "mac_required"})
		return
	}

	existed, _ := s.st.Delete(ctx, mac)

	s.audit.Write(map[string]any{
		"event":   "portal.logout",
		"mac":     mac,
		"existed": existed,
		"source":  req.Meta.Source,
		"result":  map[bool]string{true: "ok", false: "not_found"}[existed],
	})

	writeJSON(w, 200, map[string]any{"authorized": false})
}

// -------------------------------------------------------------------
// Ops / Status APIs (unchanged)
// -------------------------------------------------------------------

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
		PolicyVersion *string        `json:"policy_version,omitempty"`
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

// ---------------------------------------------------
// Portal Context Verify (auth_request backend)
// ---------------------------------------------------
//
// This endpoint is called by portal-signer via nginx auth_request.
// It must be:
//   - side-effect free
//   - header driven
//   - status-code only
func (s *Server) portalContextVerify(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// --------------------------------------------------
	// 1. Internal-only guard
	// --------------------------------------------------
	if r.Header.Get("X-Portal-Internal") != "1" {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	// --------------------------------------------------
	// 2. Extract original request info
	// --------------------------------------------------
	method := r.Header.Get("X-Original-Method")
	uri := r.Header.Get("X-Original-URI")

	if method == "" || uri == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	// --------------------------------------------------
	// 3. Get client MAC from context (set by middleware)
	// --------------------------------------------------
	macAny := ctx.Value(security.CtxKeyClientMAC)
	mac, ok := macAny.(string)
	if !ok || mac == "" {
		// Not logged in
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// --------------------------------------------------
	// 4. Session existence check (v0 semantics)
	// --------------------------------------------------
	sess, _, err := s.st.GetSessionFull(ctx, mac)
	if err != nil || sess == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// --------------------------------------------------
	// 5. Allow (NO BODY)
	// --------------------------------------------------
	w.WriteHeader(http.StatusNoContent)
}
