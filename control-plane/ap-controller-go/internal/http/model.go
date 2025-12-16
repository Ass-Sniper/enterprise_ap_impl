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

// LoginReq portal login request
type LoginReq struct {
	MAC     string `json:"mac" example:"aa:bb:cc:dd:ee:ff"`
	SSID    string `json:"ssid,omitempty" example:"GuestWiFi"`
	Auth    string `json:"auth,omitempty" example:"portal"`
	APID    string `json:"ap_id,omitempty" example:"ap-123"`
	RadioID string `json:"radio_id,omitempty" example:"radio-1"`
	IP      string `json:"ip,omitempty" example:"192.168.1.23"`
	Source  string `json:"source,omitempty" example:"portal"`
}

// HeartbeatReq portal heartbeat request
type HeartbeatReq struct {
	MAC     string `json:"mac" example:"aa:bb:cc:dd:ee:ff"`
	Source  string `json:"source,omitempty" example:"portal"`
	APID    string `json:"ap_id,omitempty" example:"ap-123"`
	SSID    string `json:"ssid,omitempty" example:"GuestWiFi"`
	RadioID string `json:"radio_id,omitempty" example:"radio-1"`
}

// LogoutReq portal logout request
type LogoutReq struct {
	MAC    string `json:"mac" example:"aa:bb:cc:dd:ee:ff"`
	Source string `json:"source,omitempty" example:"portal"`
}

// BatchEntry single batch query entry
type BatchEntry struct {
	MAC    string `json:"mac" example:"aa:bb:cc:dd:ee:ff"`
	Source string `json:"source,omitempty" example:"portal"`
}

// BatchReq batch portal status request
type BatchReq struct {
	Entries []BatchEntry `json:"entries"`
}

// ErrorResponse standard error response
type ErrorResponse struct {
	Code    string `json:"code" example:"bad_request"`
	Message string `json:"message" example:"invalid mac"`
}
