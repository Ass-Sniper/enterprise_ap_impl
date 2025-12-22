//go:build pap
// +build pap

package pap

import "access_device/auth"

// Plugin 实现 auth.Plugin 接口
type Plugin struct{}

func (p *Plugin) Name() string {
	return "pap"
}

// Install 把 PAP Strategy 注册进 StrategyStore
func (p *Plugin) Install(store *auth.StrategyStore, deps *auth.Dependencies) {
	store.Add("pap", func(r auth.RequestContext) (auth.Strategy, error) {
		return &Strategy{
			Username: r.Username,
			Password: r.Password,
			Auth:     deps.RadiusAuth,
		}, nil
	})
}

// init 是插件真正生效的关键
func init() {
	auth.RegisterPlugin(&Plugin{})
}
