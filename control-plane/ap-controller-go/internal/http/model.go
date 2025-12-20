package httpapi

import (
	"ap-controller-go/internal/audit"
	"ap-controller-go/internal/config"
	"ap-controller-go/internal/security"
	"ap-controller-go/internal/store"
)

// Server controller http server
type Server struct {
	cfg   *config.Config
	st    *store.Store
	audit *audit.Logger
	// policy version can be hot-reloaded later; for now read from cfg
	policyVersion string
	//
	jwtIssuer *security.JWTIssuer // NEW
}

// -------------------------------------------------------------------
// Trusted Portal Context (v2)
// -------------------------------------------------------------------

// PortalContextReq is a trusted context envelope sent from Portal Server.
// All fields are derived from trusted headers injected by nginx.
type PortalContextReq struct {
	Client struct {
		MAC string `json:"mac" example:"aa:bb:cc:dd:ee:ff"`
		IP  string `json:"ip,omitempty" example:"192.168.1.23"`
		OS  string `json:"os,omitempty" example:"iOS"`
	} `json:"client"`

	Wireless struct {
		SSID    string `json:"ssid,omitempty" example:"GuestWiFi"`
		RadioID string `json:"radio_id,omitempty" example:"radio-1"`
	} `json:"wireless"`

	Access struct {
		APID   string `json:"ap_id,omitempty" example:"ap-123"`
		VLANID string `json:"vlan_id,omitempty" example:"100"`
	} `json:"access"`

	Security struct {
		Timestamp string `json:"timestamp" example:"1690000000.123"`
		Nonce     string `json:"nonce" example:"req-uuid"`
		Signature string `json:"signature" example:"base64(hmac)"`
	} `json:"security"`

	Meta struct {
		Source string `json:"source,omitempty" example:"portal"`
	} `json:"meta"`
}

// -------------------------------------------------------------------
// Legacy / Ops APIs (unchanged)
// -------------------------------------------------------------------

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
