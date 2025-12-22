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
		Password: req.Password,
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
	// ============================
	// Resolve Redirect URL（认证成功后的跳转地址）
	//
	// 优先级说明：
	// 1. 请求显式指定（Portal UI 传入，最高优先级）
	// 2. 策略配置指定（policy.yaml 中的 redirectURL）
	// 3. 系统默认兜底（PortalServer.DefaultRedirectURL）
	//
	// 设计原则：
	// - UI 可以临时覆盖跳转行为
	// - 业务策略可以统一控制落地页
	// - 系统必须始终有安全兜底
	// ============================
	redirectURL := req.RedirectURL

	// a 若请求未指定跳转地址，则尝试使用策略中配置的 redirectURL
	if redirectURL == "" && policy != nil {
		redirectURL = policy.RedirectURL
	}

	// b 若策略也未指定，则使用系统级默认跳转地址
	if redirectURL == "" {
		redirectURL = p.DefaultRedirectURL
	}

	// c 记录最终跳转决策，便于排查策略 / UI / 默认配置问题
	log.Printf(
		"[AUTH SUCCESS] user=%s redirect=%s (req=%s policy=%s default=%s)",
		username,
		redirectURL,
		req.RedirectURL,
		policy.RedirectURL,
		p.DefaultRedirectURL,
	)

	// 构造响应
	resp := AuthResponse{
		Success:     true,
		Message:     "ok",
		RedirectURL: redirectURL,
		Policy:      policy.Name,
		Strategy:    strategy.Name(),
	}

	// 打印日志：包含用户、策略信息
	log.Printf("[AUTH SUCCESS] User: %s | Policy: %s | Strategy: %s | RedirectURL: %s",
		username, resp.Policy, resp.Strategy, resp.RedirectURL)

	// 发送响应
	WriteJSON(w, resp)
}
