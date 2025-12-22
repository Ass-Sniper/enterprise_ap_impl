package config

type AuthConfig struct {
	Auth struct {
		Strategies map[string]AuthStrategyConfig `yaml:"strategies" json:"strategies"`
	} `yaml:"auth" json:"auth"`
}

type AuthStrategyConfig struct {
	Enabled bool `yaml:"enabled" json:"enabled"`
}
