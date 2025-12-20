package config

type Config struct {
	Controller Controller         `yaml:"controller"`
	Redis      Redis              `yaml:"redis"`
	Roles      map[string]RoleDef `yaml:"roles"`
	Profiles   map[string]Profile `yaml:"profiles"`
	RoleRules  []RoleRule         `yaml:"role_rules"`
	Bypass     Bypass             `yaml:"bypass"`
	Dataplane  Dataplane          `yaml:"dataplane"`
}

type Controller struct {
	ID      string `yaml:"id"`
	Site    string `yaml:"site"`
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
	Bind    struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
	} `yaml:"bind"`
	Audit struct {
		Enabled   bool   `yaml:"enabled"`
		Level     string `yaml:"level"`
		SecretRef string `yaml:"secret_ref"`
		Algo      string `yaml:"algo"`
	} `yaml:"audit"`
	HMACSecret string `yaml:"hmac_secret"`
}

type Redis struct {
	Host    string `yaml:"host"`
	Port    int    `yaml:"port"`
	DB      int    `yaml:"db"`
	Prefix  string `yaml:"prefix"`
	TLS     bool   `yaml:"tls"`
	AuthRef string `yaml:"auth_ref"`
}

type RoleDef struct {
	Profile string `yaml:"profile"`
}

type Profile struct {
	VLAN          int    `yaml:"vlan"`
	FirewallGroup string `yaml:"firewall_group"`
	SessionTTL    int    `yaml:"session_ttl"`
}

type RoleRule struct {
	Name     string         `yaml:"name"`
	Priority int            `yaml:"priority"`
	When     map[string]any `yaml:"when"`
	Assign   string         `yaml:"assign"`
}

type Bypass struct {
	Enabled      bool     `yaml:"enabled"`
	EnforceOrder []string `yaml:"enforce_order"`
	MacWhitelist []string `yaml:"mac_whitelist"`
	IPWhitelist  []string `yaml:"ip_whitelist"`
	Domains      []string `yaml:"domains"`
}

type Dataplane struct {
	PolicyVersion int               `yaml:"policy_version"`
	PortalIP      string            `yaml:"portal_ip"`
	LanIF         string            `yaml:"lan_if"`
	IPSets        map[string]string `yaml:"ipsets"`
}
