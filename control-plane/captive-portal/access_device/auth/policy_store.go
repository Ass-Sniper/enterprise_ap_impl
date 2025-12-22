package auth

import "sync"

type PolicyStore struct {
	mu       sync.RWMutex
	policies map[string]*Policy
}

func NewPolicyStore() *PolicyStore {
	return &PolicyStore{
		policies: make(map[string]*Policy),
	}
}

func (ps *PolicyStore) Add(p *Policy) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.policies[p.Name] = p
}

func (ps *PolicyStore) Get(name string) *Policy {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	return ps.policies[name]
}
