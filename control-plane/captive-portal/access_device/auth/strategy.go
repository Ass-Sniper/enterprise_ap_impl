package auth

import "context"

// Result：认证输出（用于后续 Policy/Session）
type Result struct {
	Username       string
	Password       string
	Token          string
	Phone          string
	Code           string
	AuthMethod     string // pap/token/sms
	Policy         string // 例如 portal-users / staff / guest
	SessionTimeout int64  // 秒，0=永不过期
	ReplyAttrs     map[string]string
}

type Strategy interface {
	Name() string
	Authenticate(ctx context.Context) (ok bool, res *Result, err error)
}
