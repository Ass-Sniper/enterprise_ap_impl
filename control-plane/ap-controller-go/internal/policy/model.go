package policy

// =========================
// Runtime Policy Model
// =========================

// RuntimePolicy is the top-level response returned to AP / Router
type RuntimePolicy struct {
	Controller ControllerInfo            `json:"controller"`
	Version    PolicyVersion             `json:"version"`
	Roles      map[string]RuntimeRole    `json:"roles"`
	Profiles   map[string]RuntimeProfile `json:"profiles"`
	Bypass     RuntimeBypass             `json:"bypass"`
}

// ControllerInfo identifies controller instance
type ControllerInfo struct {
	ID   string `json:"id"`
	Site string `json:"site"`
	Name string `json:"name"`
}

// =========================
// Versioning
// =========================

type PolicyVersion struct {
	Version   string `json:"version"`   // semantic or incremental
	Checksum  string `json:"checksum"`  // sha256 of runtime payload
	Generated int64  `json:"generated"` // unix timestamp
}

// =========================
// Role / Profile
// =========================

type RuntimeRole struct {
	Profile string `json:"profile"`
}

type RuntimeProfile struct {
	VLAN          int    `json:"vlan"`
	FirewallGroup string `json:"firewall_group"`
	SessionTTL    int    `json:"session_ttl"`
}

// =========================
// Bypass
// =========================

type RuntimeBypass struct {
	Enabled      bool     `json:"enabled"`
	EnforceOrder []string `json:"enforce_order"`
	MacWhitelist []string `json:"mac_whitelist"`
	IPWhitelist  []string `json:"ip_whitelist"`
	Domains      []string `json:"domains"`
}
