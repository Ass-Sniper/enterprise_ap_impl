package auth

import (
	"encoding/json"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// YAML 文件结构
type policyYAML struct {
	Policies []Policy `yaml:"policies"`
}

// PolicyStoreFromYAML 从 YAML 文件加载 PolicyStore
func LoadPolicyStoreFromYAML(path string) (*PolicyStore, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg policyYAML
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// 将 store 或 cfg 转换为 JSON
	// 使用 MarshalIndent 可以让输出格式化，便于人类阅读
	jsonData, err := json.MarshalIndent(cfg, "", "    ")
	if err == nil {
		fmt.Println(string(jsonData))
	}

	store := NewPolicyStore()
	for _, p := range cfg.Policies {
		pol := p // copy
		store.Add(&pol)
	}

	return store, nil
}
