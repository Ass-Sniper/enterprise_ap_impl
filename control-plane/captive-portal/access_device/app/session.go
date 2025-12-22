package app

import "time"

// Session 表示一个已认证用户会话（对齐 NAC / 华为 Portal 语义）
type Session struct {
	// 基本身份
	Username string `json:"username"`
	IP       string `json:"ip"`

	// 认证与策略
	Strategy string `json:"strategy"` // pap / token / sms
	Policy   string `json:"policy"`   // policy name

	// 时间信息
	LoginAt time.Time `json:"login_at"`
	TTL     int       `json:"ttl"` // 秒

	// 状态（预留给状态机）
	State string `json:"state,omitempty"`
}
