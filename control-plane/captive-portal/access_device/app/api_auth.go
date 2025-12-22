package app

import (
	"access_device/auth"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	policyConfigPath = "./config/policy.yaml"
)

// PortalServer: NAC API Server（不渲染页面）
type PortalServer struct {
	AuthF      *auth.Factory
	SessionMgr *SessionManager
	// 默认成功跳转（Portal UI 可以不传 redirect_url 时使用）
	DefaultRedirectURL string
}

// BuildPortalServer 负责装配 PortalServer（composition root）
func BuildPortalServer(rdb *redis.Client) *PortalServer {
	log.Println("[BOOT] build PortalServer")

	// ============================
	// Session Manager
	// ============================
	log.Println("[BOOT] init SessionManager")
	sessMgr := NewSessionManager(rdb)

	// ============================
	// Strategy Store
	// ============================
	log.Println("[BOOT] init StrategyStore")
	store := auth.NewStrategyStore()

	// ============================
	// Auth Dependencies
	// ============================
	log.Println("[BOOT] init Auth Dependencies")
	deps := &auth.Dependencies{
		RadiusAuth: &RadiusPAPAuth{},
	}

	// ============================
	// Install Auth Plugins
	// ============================
	plugins := auth.AllPlugins()
	log.Printf("[BOOT] install auth plugins: %d\n", len(plugins))
	for _, p := range plugins {
		log.Printf("[BOOT] plugin install: %s\n", p.Name())
		p.Install(store, deps)
	}
	log.Printf("[BOOT] available strategies: %v\n", store.Names())

	// ============================
	// Policy Store (YAML)
	// ============================
	log.Println("[BOOT] load policy store")
	policyStore, err := auth.LoadPolicyStoreFromYAML(
		policyConfigPath,
	)
	if err != nil {
		log.Fatalf("[BOOT][FATAL] load policy failed: %v", err)
	}

	// ============================
	// Policy Engine
	// ============================
	log.Println("[BOOT] init PolicyEngine")
	policyEngine := auth.NewPolicyEngine(
		policyStore,
		auth.NewRedisPolicyOverride(rdb),
	)

	// ============================
	// Feature Flag
	// ============================
	var ff auth.FeatureFlag
	if rdb != nil {
		log.Println("[BOOT] use RedisFeatureFlag")
		ff = &auth.RedisFeatureFlag{RDB: rdb}
	} else {
		log.Println("[BOOT] use NoopFeatureFlag")
		ff = auth.NoopFeatureFlag{}
	}

	// ============================
	// Auth Factory
	// ============================
	log.Println("[BOOT] init Auth Factory")
	authFactory := auth.NewFactory(
		store,
		policyEngine,
		ff,
	)

	log.Println("[BOOT] PortalServer ready")

	return &PortalServer{
		SessionMgr: sessMgr,
		AuthF:      authFactory,
	}
}

// ============================
// HTTP Handlers
// ============================

// PortalAuthHandler 处理 Portal(UI) 发起的认证请求（JSON API）
func (p *PortalServer) PortalAuthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		WriteJSON(w, AuthResponse{
			Success: false,
			Message: "method not allowed",
		})
		return
	}

	if ct := r.Header.Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		WriteJSON(w, AuthResponse{
			Success: false,
			Message: "content-type must be application/json",
		})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req AuthRequest
	log.Printf("[PORTAL_AUTH] decode request")

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSON(w, AuthResponse{
			Success: false,
			Message: "invalid request body",
		})
		return
	}

	username := req.Username
	if username == "" {
		WriteJSON(w, AuthResponse{
			Success: false,
			Message: "missing username",
		})
		return
	}

	clientIP := req.UserIP
	if clientIP == "" {
		clientIP = r.RemoteAddr // 兜底
	}

	log.Printf(
		"[PORTAL_AUTH] request user=%s ip=%s auth_type=%s nas_id=%s",
		username, clientIP, req.AuthType, req.NASID,
	)

	// 初始 hint
	hint := &auth.Result{
		Username: username,
	}

	// ============================
	// Strategy + Policy
	// ============================
	strategy, policy, err := p.AuthF.SelectAndBuild(ctx, r, hint)
	if err != nil {
		log.Printf("[AUTH][DENY] user=%s err=%v", username, err)
		WriteJSON(w, AuthResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	log.Printf(
		"[AUTH] user=%s strategy=%s policy=%s",
		username, strategy.Name(), policy.Name,
	)

	// ============================
	// Authenticate
	// ============================
	ok, result, err := strategy.Authenticate(ctx)
	if result != nil {
		log.Printf("[AUTH] extra result: %+v", result)
	}

	if err != nil {
		log.Printf("[AUTH][ERROR] user=%s err=%v", username, err)
		WriteJSON(w, AuthResponse{
			Success: false,
			Message: "auth service error",
		})
		return
	}

	if !ok {
		log.Printf("[AUTH][REJECT] user=%s", username)
		WriteJSON(w, AuthResponse{
			Success: false,
			Message: "authentication failed",
		})
		return
	}

	// ============================
	// Session Create
	// ============================
	sess := &Session{
		Username: username,
		IP:       clientIP,
		Policy:   policy.Name,
		Strategy: strategy.Name(),
		LoginAt:  time.Now(),
		TTL:      int(policy.SessionTimeout),
	}

	if err := p.SessionMgr.Save(ctx, sess); err != nil {
		log.Printf("[SESSION][ERROR] user=%s err=%v", username, err)
		WriteJSON(w, AuthResponse{
			Success: false,
			Message: "session create failed",
		})
		return
	}

	log.Printf(
		"[SESSION] create user=%s policy=%s strategy=%s ttl=%d",
		username, policy.Name, strategy.Name(), policy.SessionTimeout,
	)

	// ============================
	// Success
	// ============================
	redirectURL := req.RedirectURL
	if redirectURL == "" {
		redirectURL = p.DefaultRedirectURL
	}

	WriteJSON(w, AuthResponse{
		Success:     true,
		Message:     "ok",
		RedirectURL: redirectURL,
		Policy:      policy.Name,
		Strategy:    strategy.Name(),
	})
}
