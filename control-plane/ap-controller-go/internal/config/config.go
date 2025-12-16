package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

func Load(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return nil, err
	}
	if cfg.Redis.Prefix == "" {
		cfg.Redis.Prefix = "session:"
	}
	if cfg.Dataplane.LanIF == "" {
		return nil, fmt.Errorf("dataplane.lan_if must be set")
	}
	if cfg.Dataplane.PortalIP == "" {
		return nil, fmt.Errorf("dataplane.portal_ip must be set")
	}
	return &cfg, nil
}

// Resolve "env:XXX" to actual secret.
func ResolveSecret(ref string) (string, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return "", errors.New("empty secret_ref")
	}
	if strings.HasPrefix(ref, "env:") {
		key := strings.TrimPrefix(ref, "env:")
		v := os.Getenv(key)
		if v == "" {
			return "", fmt.Errorf("env %s is empty", key)
		}
		return v, nil
	}
	// future extension: file:/path, vault:..., kms:...
	return ref, nil
}
