package auth

import "context"

// 永驻策略覆盖的实际意义：docs/feature-flag-and-policy-override.md
// 主决策流程图：docs/auth-decision-pipeline.md
type PolicyOverride interface {
	Get(ctx context.Context, username string) (*Policy, error)
}
