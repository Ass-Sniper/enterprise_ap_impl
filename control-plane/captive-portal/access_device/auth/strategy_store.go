package auth

import (
	"fmt"
	"sync"
)

// StrategyBuilder 根据 RequestContext 构建一个 Strategy
type StrategyBuilder func(RequestContext) (Strategy, error)

// StrategyStore 保存所有可用的 Strategy Builder
type StrategyStore struct {
	mu       sync.RWMutex
	builders map[string]StrategyBuilder
}

// NewStrategyStore 创建一个新的 StrategyStore
func NewStrategyStore() *StrategyStore {
	return &StrategyStore{
		builders: make(map[string]StrategyBuilder),
	}
}

// Add 注册一个 Strategy Builder
func (s *StrategyStore) Add(name string, b StrategyBuilder) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.builders[name] = b
}

// Build 根据名字和 RequestContext 构建 Strategy
func (s *StrategyStore) Build(name string, ctx RequestContext) (Strategy, error) {
	s.mu.RLock()
	b, ok := s.builders[name]
	s.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("unknown auth strategy: %s", name)
	}
	return b(ctx)
}

// Has 判断 Strategy 是否存在
func (s *StrategyStore) Has(name string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.builders[name]
	return ok
}

// Names 返回所有已注册的 Strategy 名称
func (s *StrategyStore) Names() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	names := make([]string, 0, len(s.builders))
	for k := range s.builders {
		names = append(names, k)
	}
	return names
}
