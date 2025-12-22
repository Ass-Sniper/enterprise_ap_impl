//go:build token
// +build token

package token

import "access_device/auth"

// Plugin 必须先声明
type Plugin struct{}

func (p *Plugin) Name() string {
	return "token"
}

func (p *Plugin) Install(store *auth.StrategyStore, deps *auth.Dependencies) {
	store.Add("token", func(r auth.RequestContext) (auth.Strategy, error) {
		return &Strategy{
			Username: r.Username,
			Token:    r.Token,
		}, nil
	})
}

func init() {
	auth.RegisterPlugin(&Plugin{})
}
