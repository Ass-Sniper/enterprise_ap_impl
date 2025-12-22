package api

// AuthRequest 是 Portal → Access Device 的认证请求
type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	UserIP   string `json:"user_ip"`
}

// AuthResponse 是 Access Device → Portal 的认证响应
// 与下面的文件中的一致
// /home/kay/codebase/enterprise_ap_impl/control-plane/captive-portal/access_device/app/api_types.go
type AuthResponse struct {
	Success     bool              `json:"success"`
	Message     string            `json:"message,omitempty"`
	RedirectURL string            `json:"redirect_url,omitempty"`
	Policy      string            `json:"policy,omitempty"`
	Strategy    string            `json:"strategy,omitempty"`
	ReplyAttrs  map[string]string `json:"reply_attrs,omitempty"`
}
