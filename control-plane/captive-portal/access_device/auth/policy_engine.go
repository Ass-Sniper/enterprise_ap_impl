package auth

import (
	"context"
	"fmt"
	"log"
)

// PolicyEngine 负责综合决策用户策略
type PolicyEngine struct {
	store    *PolicyStore
	override PolicyOverride
}

// NewPolicyEngine 创建一个新的 PolicyEngine
func NewPolicyEngine(
	store *PolicyStore,
	override PolicyOverride,
) *PolicyEngine {
	return &PolicyEngine{
		store:    store,
		override: override,
	}
}

// Resolve 根据用户名 + RADIUS Reply 决定最终 Policy
func (policyEngine *PolicyEngine) Resolve(
	ctx context.Context,
	username string,
	replyAttrs map[string]string,
) (*Policy, error) {

	// 1. Redis Override（最高优先级）
	if policyEngine.override != nil {
		if p, err := policyEngine.override.Get(ctx, username); err == nil && p != nil {
			p.Source = PolicyFromOverride
			logPolicy(username, p)
			return p, nil
		}
	}

	// 2. RADIUS Reply Hint（如 Filter-Id）
	if v := replyAttrs["Filter-Id"]; v != "" {
		if p := policyEngine.store.Get(v); p != nil {
			p.Source = PolicyFromRadius
			logPolicy(username, p)
			return p, nil
		}
	}

	// 3. 默认策略
	if p := policyEngine.store.Get("default"); p != nil {
		p.Source = PolicyFromDefault
		logPolicy(username, p)
		return p, nil
	}

	// 4. No Policy
	log.Printf("[POLICY] user=%s source=none policy=none", username)
	return nil, fmt.Errorf("no policy found for user %s", username)
}
