//go:build sms
// +build sms

package sms

import (
	"access_device/auth"
)

type Plugin struct{}

func (p *Plugin) Name() string { return "sms" }

func (p *Plugin) Install(store *auth.StrategyStore, deps *auth.Dependencies) {
	store.Add("sms", func(r auth.RequestContext) (auth.Strategy, error) {
		return &Strategy{
			Phone: r.Phone,
			Code:  r.Code,
			Auth:  deps.RadiusAuth, // 如果后续有 SMSProvider，可替换
		}, nil
	})
}

func init() {
	auth.RegisterPlugin(&Plugin{})
}
