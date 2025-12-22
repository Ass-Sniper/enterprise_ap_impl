package app

// import (
// 	"context"
// 	"encoding/json"
// 	"log"
// 	"net/http"
// 	"time"

// 	"access_device/auth"

// 	"github.com/redis/go-redis/v9"
// )

// const (
// 	policyConfigPath                 = "./config/policy.yaml"
// 	webRootDir                       = "./web/templates"
// 	portalServerPort                 = 8080
// 	portalServerAddress              = "172.19.0.1"
// 	nasAuthSuccessDefaultRedirectURL = "www.bing.com"
// )

// // PortalServer: NAC API Server（不渲染页面）
// type PortalServer struct {
// 	AuthF      *auth.Factory
// 	SessionMgr *SessionManager
// 	// Renderer   *web.Renderer
// 	// 默认成功跳转（Portal UI 可以不传 redirect_url 时使用）
// 	DefaultRedirectURL string
// }

// // BuildPortalServer 负责装配 PortalServer（composition root）
// func BuildPortalServer(rdb *redis.Client) *PortalServer {
// 	log.Println("[BOOT] build PortalServer")

// 	// ============================
// 	// Session Manager
// 	// ============================
// 	log.Println("[BOOT] init SessionManager")
// 	sessMgr := NewSessionManager(rdb)

// 	// ============================
// 	// Strategy Store
// 	// ============================
// 	log.Println("[BOOT] init StrategyStore")
// 	store := auth.NewStrategyStore()

// 	// ============================
// 	// Auth Dependencies
// 	// ============================
// 	log.Println("[BOOT] init Auth Dependencies")
// 	deps := &auth.Dependencies{
// 		RadiusAuth: &RadiusPAPAuth{},
// 	}

// 	// ============================
// 	// Install Auth Plugins
// 	// ============================
// 	plugins := auth.AllPlugins()
// 	log.Printf("[BOOT] install auth plugins: %d\n", len(plugins))
// 	for _, p := range plugins {
// 		log.Printf("[BOOT] plugin install: %s\n", p.Name())
// 		p.Install(store, deps)
// 	}
// 	log.Printf("[BOOT] available strategies: %v\n", store.Names())

// 	// ============================
// 	// Policy Store (YAML)
// 	// ============================
// 	log.Println("[BOOT] load policy store")
// 	policyStore, err := auth.LoadPolicyStoreFromYAML(
// 		policyConfigPath,
// 	)
// 	if err != nil {
// 		log.Fatalf("[BOOT][FATAL] load policy failed: %v", err)
// 	}

// 	// ============================
// 	// Policy Engine
// 	// ============================
// 	log.Println("[BOOT] init PolicyEngine")
// 	policyEngine := auth.NewPolicyEngine(
// 		policyStore,
// 		auth.NewRedisPolicyOverride(rdb),
// 	)

// 	// ============================
// 	// Feature Flag
// 	// ============================
// 	var ff auth.FeatureFlag
// 	if rdb != nil {
// 		log.Println("[BOOT] use RedisFeatureFlag")
// 		ff = &auth.RedisFeatureFlag{RDB: rdb}
// 	} else {
// 		log.Println("[BOOT] use NoopFeatureFlag")
// 		ff = auth.NoopFeatureFlag{}
// 	}

// 	// ============================
// 	// Auth Factory
// 	// ============================
// 	log.Println("[BOOT] init Auth Factory")
// 	authFactory := auth.NewFactory(
// 		store,
// 		policyEngine,
// 		ff,
// 	)

// 	// // ============================
// 	// // Renderer
// 	// // ============================
// 	// log.Println("[BOOT] init Renderer")
// 	// renderer, err := web.NewRenderer(webRootDir)
// 	// if err != nil {
// 	// 	log.Fatalf("[BOOT][FATAL] init renderer failed: %v", err)
// 	// }

// 	log.Println("[BOOT] PortalServer ready")

// 	return &PortalServer{
// 		SessionMgr: sessMgr,
// 		AuthF:      authFactory,
// 		// Renderer:   renderer,
// 	}
// }

// // ============================
// // HTTP Handlers
// // ============================

// // PortalAuthHandler 处理 Portal(UI) 发起的认证请求（JSON API）
// func (p *PortalServer) PortalAuthHandler(w http.ResponseWriter, r *http.Request) {
// 	if r.Method != http.MethodPost {
// 		w.WriteHeader(http.StatusMethodNotAllowed)
// 		WriteJSON(w, AuthResponse{
// 			Success: false,
// 			Message: "method not allowed",
// 		})
// 		return
// 	}

// 	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
// 	defer cancel()

// 	var req AuthRequest
// 	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
// 		WriteJSON(w, AuthResponse{
// 			Success: false,
// 			Message: "invalid request body",
// 		})
// 		return
// 	}

// 	username := req.Username
// 	clientIP := req.UserIP
// 	if clientIP == "" {
// 		clientIP = r.RemoteAddr // 兜底，不推荐长期依赖
// 	}

// 	log.Printf("[PORTAL_AUTH] request user=%s ip=%s", username, clientIP)

// 	// 初始 hint（用户名 / RADIUS Reply）
// 	hint := &auth.Result{
// 		Username: username,
// 	}

// 	// ============================
// 	// Strategy + Policy 选择
// 	// ============================
// 	strategy, policy, err := p.AuthF.SelectAndBuild(ctx, r, hint)
// 	if err != nil {
// 		log.Printf("[AUTH][DENY] user=%s err=%v", username, err)
// 		WriteJSON(w, AuthResponse{
// 			Success: false,
// 			Message: err.Error(),
// 		})
// 		return
// 	}

// 	log.Printf("[AUTH] user=%s strategy=%s policy=%s",
// 		username, strategy.Name(), policy.Name,
// 	)

// 	// ============================
// 	// Authenticate
// 	// ============================
// 	ok, result, err := strategy.Authenticate(ctx)
// 	if result != nil {
// 		log.Printf("[AUTH] extra result: %+v", result)
// 	}
// 	if err != nil {
// 		log.Printf("[AUTH][ERROR] user=%s err=%v", username, err)
// 		WriteJSON(w, AuthResponse{
// 			Success: false,
// 			Message: "auth service error",
// 		})
// 		return
// 	}
// 	if !ok {
// 		log.Printf("[AUTH][REJECT] user=%s", username)
// 		WriteJSON(w, AuthResponse{
// 			Success: false,
// 			Message: "认证失败",
// 		})
// 		return
// 	}

// 	// ============================
// 	// Session Create
// 	// ============================
// 	sess := &Session{
// 		Username: username,
// 		IP:       clientIP,
// 		Policy:   policy.Name,
// 		Strategy: strategy.Name(),
// 		LoginAt:  time.Now(),
// 		TTL:      int(policy.SessionTimeout),
// 	}

// 	if err := p.SessionMgr.Save(ctx, sess); err != nil {
// 		log.Printf("[SESSION][ERROR] user=%s err=%v", username, err)
// 		WriteJSON(w, AuthResponse{
// 			Success: false,
// 			Message: "session create failed",
// 		})
// 		return
// 	}

// 	log.Printf("[SESSION] create user=%s policy=%s strategy=%s ttl=%d",
// 		username, policy.Name, strategy.Name(), policy.SessionTimeout,
// 	)

// 	// ============================
// 	// Success (JSON)
// 	// ============================
// 	redirectURL := req.RedirectURL
// 	if redirectURL == "" {
// 		redirectURL = p.DefaultRedirectURL
// 	}

// 	WriteJSON(w, AuthResponse{
// 		Success:     true,
// 		Message:     "ok",
// 		RedirectURL: redirectURL,
// 		Policy:      policy.Name,
// 		Strategy:    strategy.Name(),
// 	})
// }

// // PortalAuthHandler 处理 Portal 登录请求
// func (p *PortalServer) PortalAuthHandler(w http.ResponseWriter, r *http.Request) {
// 	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
// 	defer cancel()

// 	if err := r.ParseForm(); err != nil {
// 		p.Renderer.RenderResult(w, web.ResultData{
// 			Success: false,
// 			Message: "参数解析失败",
// 		})
// 		return
// 	}

// 	username := r.FormValue("username")
// 	clientIP := r.RemoteAddr

// 	log.Printf("[PORTAL] auth request user=%s ip=%s\n", username, clientIP)

// 	// 初始 hint（RADIUS Reply / username）
// 	hint := &auth.Result{
// 		Username: username,
// 	}

// 	// ============================
// 	// Strategy + Policy 选择
// 	// ============================
// 	strategy, policy, err := p.AuthF.SelectAndBuild(ctx, r, hint)
// 	if err != nil {
// 		log.Printf("[AUTH][DENY] user=%s err=%v\n", username, err)
// 		p.Renderer.RenderResult(w, web.ResultData{
// 			Success: false,
// 			Message: err.Error(),
// 		})
// 		return
// 	}

// 	log.Printf(
// 		"[AUTH] user=%s strategy=%s policy=%s\n",
// 		username, strategy.Name(), policy.Name,
// 	)

// 	// ============================
// 	// Authenticate
// 	// ============================
// 	ok, result, err := strategy.Authenticate(ctx)
// 	if result != nil {
// 		log.Printf("[AUTH] extra result: %+v\n", result)
// 	}
// 	if err != nil {
// 		log.Printf("[AUTH][ERROR] user=%s err=%v\n", username, err)
// 		p.Renderer.RenderResult(w, web.ResultData{
// 			Success: false,
// 			Message: "认证服务异常",
// 		})
// 		return
// 	}

// 	if !ok {
// 		log.Printf("[AUTH][REJECT] user=%s\n", username)
// 		p.Renderer.RenderResult(w, web.ResultData{
// 			Success: false,
// 			Message: "认证失败",
// 		})
// 		return
// 	}

// 	// ============================
// 	// Session Create
// 	// ============================
// 	sess := &Session{
// 		Username: username,
// 		IP:       clientIP,
// 		Policy:   policy.Name,
// 		Strategy: strategy.Name(),
// 		LoginAt:  time.Now(),
// 		TTL:      int(policy.SessionTimeout),
// 	}

// 	if err := p.SessionMgr.Save(ctx, sess); err != nil {
// 		log.Printf("[SESSION][ERROR] user=%s err=%v\n", username, err)
// 		p.Renderer.RenderResult(w, web.ResultData{
// 			Success: false,
// 			Message: "会话创建失败",
// 		})
// 		return
// 	}

// 	log.Printf(
// 		"[SESSION] create user=%s policy=%s strategy=%s ttl=%d\n",
// 		username, policy.Name, strategy.Name(), policy.SessionTimeout,
// 	)

// 	// ============================
// 	// Success
// 	// ============================
// 	redirectURL := r.FormValue("redirect_url")
// 	if redirectURL == "" {
// 		redirectURL = nasAuthSuccessDefaultRedirectURL
// 	}
// 	p.Renderer.RenderResult(w, web.ResultData{
// 		Success:  true,
// 		Message:  "认证成功",
// 		Redirect: redirectURL,
// 	})
// }

// RedirectToPortal 捕获未认证访问
// func (p *PortalServer) RedirectToPortal(w http.ResponseWriter, r *http.Request) {
// 	log.Printf("[PORTAL] redirect from %s\n", r.RemoteAddr)
// 	http.Redirect(
// 		w,
// 		r,
// 		fmt.Sprintf("http://%s:%d/", portalServerAddress, portalServerPort),
// 		http.StatusFound,
// 	)
// }

// func (p *PortalServer) ShowLoginPage(w http.ResponseWriter, r *http.Request) {
// 	p.Renderer.RenderLogin(w, web.LoginData{})
// }
