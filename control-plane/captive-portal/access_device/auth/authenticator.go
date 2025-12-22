package auth

import "context"

// Authenticator 是 auth 层对“外部认证系统”的抽象
// pap / token / sms Strategy 只依赖这个接口
//
// radius.go / ldap.go / http.go 都可以实现它
type Authenticator interface {
	// RadiusPAP: 用户名 + 密码 → 是否通过 + ReplyAttrs + SessionTimeout
	RadiusPAP(
		ctx context.Context,
		username string,
		password string,
	) (
		ok bool,
		replyAttrs map[string]string,
		sessionTimeout int64,
		err error,
	)
}
