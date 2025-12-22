package config

// import (
// 	"os"

// 	"gopkg.in/yaml.v3"
// )

// func LoadAuthConfig(path string) (*AuthConfig, error) {
// 	data, err := os.ReadFile(path)
// 	if err != nil {
// 		return nil, err
// 	}

// 	var cfg AuthConfig
// 	if err := yaml.Unmarshal(data, &cfg); err != nil {
// 		return nil, err
// 	}
// 	return &cfg, nil
// }
