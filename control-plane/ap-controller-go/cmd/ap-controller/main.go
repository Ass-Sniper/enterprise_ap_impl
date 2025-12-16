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

	// policy version: recommended to pass via env or hot reload later
	pv := httpapi.ParsePolicyVersion(os.Getenv("POLICY_VERSION"))

	srv := httpapi.New(cfg, st, aud, pv)

	addr := fmt.Sprintf("%s:%d", cfg.Controller.Bind.Host, cfg.Controller.Bind.Port)
	log.Printf("starting %s on %s", cfg.Controller.Name, addr)
	if err := http.ListenAndServe(addr, srv.Router()); err != nil {
		log.Fatal(err)
	}
}
