package auth

// Dependencies 是所有 Strategy Plugin 共享的外部依赖
type Dependencies struct {
	RadiusAuth Authenticator
}
