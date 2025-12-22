package auth

import "context"

// FeatureFlag 决定某个认证方式是否“对某用户开放”
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
