package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"ap-controller-go/internal/audit"
	"ap-controller-go/internal/config"
	httpapi "ap-controller-go/internal/http"
	"ap-controller-go/internal/security"
	"ap-controller-go/internal/store"
)

func main() {
	cfgPath := os.Getenv("CONTROLLER_CONFIG")
	if cfgPath == "" {
		cfgPath = "/app/config/controller.yaml"
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatalf("load config failed: %v", err)
	}

	// audit secret
	secret := ""
	if cfg.Controller.Audit.Enabled {
		secret, err = config.ResolveSecret(cfg.Controller.Audit.SecretRef)
		if err != nil {
			log.Fatalf("resolve audit secret failed: %v", err)
		}
	}
	aud := audit.New(cfg.Controller.Audit.Enabled, secret)

	// redis password
	redisPwd := ""
	if cfg.Redis.AuthRef != "" {
		redisPwd, _ = config.ResolveSecret(cfg.Redis.AuthRef)
	}

	st := store.New(cfg, redisPwd)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := st.Ping(ctx); err != nil {
		log.Printf("redis ping failed: %v", err)
	}

	// --------------------------------------------------
	// 🔐 init portal hmac security (NEW)
	// --------------------------------------------------
	ks, err := security.LoadPortalHMACKeySet()
	if err != nil {
		log.Fatalf("load portal hmac keyset failed: %v", err)
	}
	security.InitPortalHMAC(ks)

	// --------------------------------------------------
	// init JWT issuer (NEW)
	// --------------------------------------------------
	jwtSecret := []byte("dev-jwt-secret-change-me") // ⚠️ 先用固定值
	jwtTTL := 15 * time.Minute
	jwtIssuer := security.NewJWTIssuer(jwtSecret, jwtTTL)

	pv := fmt.Sprintf("%v", cfg.Dataplane.PolicyVersion)

	srv := httpapi.New(cfg, st, aud, pv, jwtIssuer)

	addr := fmt.Sprintf("%s:%d", cfg.Controller.Bind.Host, cfg.Controller.Bind.Port)
	log.Printf("starting %s on %s", cfg.Controller.Name, addr)
	if err := http.ListenAndServe(addr, srv.Router()); err != nil {
		log.Fatal(err)
	}
}
