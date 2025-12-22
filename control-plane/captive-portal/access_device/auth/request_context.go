package auth

import "net/http"

// RequestContext 是 Strategy Builder 的统一输入
// 用于解耦 HTTP 与认证逻辑
type RequestContext struct {
	// 原始请求（可选，用于高级策略）
	Request *http.Request

	// 通用字段
	Username string
	Password string
	Token    string
	Phone    string
	Code     string

	// 客户端信息
	ClientIP  string
	UserAgent string
}
