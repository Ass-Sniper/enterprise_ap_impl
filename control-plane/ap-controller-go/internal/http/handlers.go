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

	_ "ap-controller-go/docs/openapi"

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

	// Swagger MUST be registered first or anywhere on the SAME router
	registerSwagger(r)

	// @Summary 服务根状态
	// @Description 返回服务运行状态
	// @Tags System
	// @Produce json
	// @Success 200 {object} map[string]interface{} "status=ok"
	// @Router / [get]
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]any{
			"status": "ok",
		})
	})

	// @Summary 健康检查
	// @Description 检查与 Redis 的连接状态
	// @Tags System
	// @Produce json
	// @Success 200 {object} map[string]interface{} "status、redis_ping"
	// @Router /healthz [get]
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

	// @Summary 获取策略运行时信息
	// @Description 返回当前策略运行时配置快照
	// @Tags Policy
	// @Produce json
	// @Success 200 {object} map[string]interface{}
	// @Router /api/v1/policy/runtime [get]
	r.Get("/api/v1/policy/runtime", policy.RuntimeHandler(s.cfg))
	return r
}

// portalLogin handles portal login.
// @Summary 门户登录授权
// @Description 根据 MAC / SSID / Auth 等信息创建或更新会话
// @Tags Portal
// @Accept json
// @Produce json
// @Param body body LoginReq true "登录请求体"
// @Success 200 {object} map[string]interface{} "authorized=true 时返回会话信息"
// @Failure 400 {object} ErrorResponse "bad_json"
// @Failure 422 {object} ErrorResponse "mac_required"
// @Router /portal/login [post]
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

// portalHeartbeat refreshes session TTL.
// @Summary 门户心跳
// @Description 刷新会话 TTL，返回最新会话信息
// @Tags Portal
// @Accept json
// @Produce json
// @Param body body HeartbeatReq true "心跳请求体"
// @Success 200 {object} map[string]interface{} "会话信息"
// @Failure 400 {object} ErrorResponse "bad_json"
// @Failure 422 {object} ErrorResponse "mac_required"
// @Router /portal/heartbeat [post]
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

// @Summary 门户登出
// @Description 删除指定 MAC 的会话
// @Tags Portal
// @Accept json
// @Produce json
// @Param body body LogoutReq true "登出请求体"
// @Success 200 {object} map[string]interface{} "authorized=false"
// @Failure 400 {object} ErrorResponse "bad_json"
// @Failure 422 {object} ErrorResponse "mac_required"
// @Router /portal/logout [post]
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

// @Summary 门户状态查询
// @Description 查询单个 MAC 的授权状态
// @Tags Portal
// @Produce json
// @Param mac path string true "客户端 MAC 地址" example(aa:bb:cc:dd:ee:ff)
// @Success 200 {object} map[string]interface{} "会话状态"
// @Router /portal/status/{mac} [get]
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

// @Summary 批量门户状态查询
// @Description 批量查询多个 MAC 的会话状态
// @Tags Portal
// @Accept json
// @Produce json
// @Param body body BatchReq true "批量状态请求体"
// @Success 200 {object} map[string]interface{} "results 数组"
// @Failure 400 {object} ErrorResponse "bad_json"
// @Router /portal/batch_status [post]
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
