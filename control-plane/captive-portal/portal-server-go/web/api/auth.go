package api

// AuthRequest 是 Portal → Access Device 的认证请求
type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	UserIP   string `json:"user_ip"`
}

// AuthResponse 是 Access Device → Portal 的认证响应
type AuthResponse struct {
	Success     bool   `json:"success"`
	Message     string `json:"message,omitempty"`
	RedirectURL string `json:"redirect_url,omitempty"`
}
