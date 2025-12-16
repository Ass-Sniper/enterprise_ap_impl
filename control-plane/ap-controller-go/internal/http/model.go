package httpapi

import (
	"ap-controller-go/internal/audit"
	"ap-controller-go/internal/config"
	"ap-controller-go/internal/store"
)

type Server struct {
	cfg   *config.Config
	st    *store.Store
	audit *audit.Logger
	// policy version can be hot-reloaded later; for now read from cfg
	policyVersion int
}

type LoginReq struct {
	MAC     string `json:"mac"`
	SSID    string `json:"ssid,omitempty"`
	Auth    string `json:"auth,omitempty"`
	APID    string `json:"ap_id,omitempty"`
	RadioID string `json:"radio_id,omitempty"`
	IP      string `json:"ip,omitempty"`
	Source  string `json:"source,omitempty"` // dhcp|arp|fdb|portal|sync
}

type HeartbeatReq struct {
	MAC     string `json:"mac"`
	Source  string `json:"source,omitempty"`
	APID    string `json:"ap_id,omitempty"`
	SSID    string `json:"ssid,omitempty"`
	RadioID string `json:"radio_id,omitempty"`
}

type LogoutReq struct {
	MAC string `json:"mac"`
}

type BatchReq struct {
	Entries []struct {
		MAC    string `json:"mac"`
		Source string `json:"source,omitempty"`
	} `json:"entries"`
}
