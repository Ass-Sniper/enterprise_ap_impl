package auth

import "context"

// FeatureFlag 的实际意义：docs/feature-flag-and-policy-override.md
// 主决策流程图：docs/auth-decision-pipeline.md
// FeatureFlag 决定某个认证方式是否“对某用户开放”
// FeatureFlag（是否开放）
//
//	↓
//
// Policy（是否允许）
//
//	↓
//
// Strategy（如何认证）
//
//	↓
//
// Authentication(eg: RADIUS)
type FeatureFlag interface {
	Enabled(ctx context.Context, feature string, user string) bool
}

// NoopFeatureFlag：默认实现（全部放行）
// - 本地开发
// - 未开启灰度
// - 单测
type NoopFeatureFlag struct{}

func (NoopFeatureFlag) Enabled(ctx context.Context, feature, user string) bool {
	return true
}
