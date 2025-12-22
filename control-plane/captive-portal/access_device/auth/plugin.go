package auth

type Plugin interface {
	Name() string
	Install(store *StrategyStore, deps *Dependencies)
}
