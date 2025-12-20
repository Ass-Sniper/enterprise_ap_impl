package store

import (
	"ap-controller-go/internal/config"

	"github.com/redis/go-redis/v9"
)

type Store struct {
	cfg    *config.Config
	rdb    *redis.Client
	prefix string
}

type SessionV2 struct {
	Schema int    `json:"schema"`
	MAC    string `json:"mac"`

	Role    string `json:"role"`
	Profile string `json:"profile"`

	PolicyVersion string `json:"policy_version"`

	Rule struct {
		Name     string `json:"name,omitempty"`
		Priority int    `json:"priority,omitempty"`
	} `json:"rule"`

	AP struct {
		APID    string `json:"ap_id,omitempty"`
		SSID    string `json:"ssid,omitempty"`
		RadioID string `json:"radio_id,omitempty"`
	} `json:"ap"`

	Attrs struct {
		VLAN          int    `json:"vlan,omitempty"`
		FirewallGroup string `json:"firewall_group,omitempty"`
	} `json:"attrs"`

	Auth struct {
		Method string `json:"method,omitempty"`
		Source string `json:"source,omitempty"`
	} `json:"auth"`

	TS struct {
		Created int64 `json:"created"`
		Updated int64 `json:"updated"`
	} `json:"ts"`
}
