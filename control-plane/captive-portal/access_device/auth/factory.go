package auth

import (
	"context"
	"errors"
	"log"
	"net/http"
)

// Factory 是认证裁决核心
type Factory struct {
	store        *StrategyStore
	PolicyEngine *PolicyEngine
	FeatureFlag  FeatureFlag
}

// NewFactory 创建认证工厂（最终版）
func NewFactory(
	store *StrategyStore,
	policyEngine *PolicyEngine,
	ff FeatureFlag,
) *Factory {
	return &Factory{
		store:        store,
		PolicyEngine: policyEngine,
		FeatureFlag:  ff,
	}
}

func (f *Factory) SelectAndBuild(
	ctx context.Context,
	r *http.Request,
	hint *Result,
) (Strategy, *Policy, error) {

	// ================================
	// 1️⃣ Resolve Policy（YAML + Redis + RADIUS hint）
	// ================================
	var pol *Policy
	if f.PolicyEngine != nil {
		var err error

		username := ""
		replyAttrs := map[string]string{}

		if hint != nil {
			username = hint.Username
			replyAttrs = hint.ReplyAttrs
		}

		pol, err = f.PolicyEngine.Resolve(
			ctx,
			username,
			replyAttrs,
		)
		if err != nil {
			return nil, nil, err
		}
	}

	// fallback policy（兜底）
	if pol == nil {
		pol = &Policy{
			Name:            "default",
			Allowed:         []string{"pap"},
			DefaultStrategy: "pap",
		}
	}

	// ================================
	// 2️⃣ Determine requested strategy
	// ================================
	reqType := r.FormValue("auth_type")

	chosen := reqType
	if chosen == "" {
		if pol.DefaultStrategy != "" {
			chosen = pol.DefaultStrategy
		} else if len(pol.Allowed) > 0 {
			chosen = pol.Allowed[0]
		}
	}

	if chosen == "" {
		return nil, pol, errors.New("no authentication strategy selected")
	}

	// ================================
	// 3️⃣ Feature Flag check（灰度裁决）
	// ================================
	if f.FeatureFlag != nil {
		user := ""
		if hint != nil {
			user = hint.Username
		}
		if !f.FeatureFlag.Enabled(ctx, chosen, user) {
			return nil, pol, errors.New("该认证方式尚未开放")
		}
	}

	// ================================
	// 4️⃣ Policy allow check（用户策略）
	// ================================
	if !contains(pol.Allowed, chosen) {
		return nil, pol, errors.New("该用户策略不允许使用此认证方式")
	}

	// ================================
	// 5️⃣ Build Strategy（plugin / store）
	// ================================
	reqCtx := RequestContext{
		Request:   r,
		Username:  "",
		Password:  "",
		Token:     "",
		Phone:     "",
		Code:      "",
		ClientIP:  r.RemoteAddr,
		UserAgent: r.UserAgent(),
	}

	if hint != nil {
		reqCtx.Username = hint.Username
		reqCtx.Password = hint.Password
		reqCtx.Token = hint.Token
		reqCtx.Phone = hint.Phone
		reqCtx.Code = hint.Code
	}

	log.Printf(
		"[FACTORY] rc.Username=%q rc.Password.len=%d",
		reqCtx.Username,
		len(reqCtx.Password),
	)

	st, err := f.store.Build(chosen, reqCtx)
	if err != nil {
		return nil, pol, err
	}

	return st, pol, nil
}

func contains(list []string, v string) bool {
	for _, s := range list {
		if s == v {
			return true
		}
	}
	return false
}
