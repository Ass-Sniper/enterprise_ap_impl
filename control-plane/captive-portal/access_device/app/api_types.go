package app

// AuthRequest: Portal(UI) -> AccessDevice(NAC) 的请求
type AuthRequest struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	UserIP      string `json:"user_ip"`
	NASID       string `json:"nas_id,omitempty"`
	RedirectURL string `json:"redirect_url,omitempty"`
	AuthType    string `json:"auth_type,omitempty"` // e.g. pap/sms/token，可选
}

// AuthResponse: AccessDevice(NAC) -> Portal(UI) 的响应
type AuthResponse struct {
	Success     bool              `json:"success"`
	Message     string            `json:"message,omitempty"`
	RedirectURL string            `json:"redirect_url,omitempty"`
	Policy      string            `json:"policy,omitempty"`
	Strategy    string            `json:"strategy,omitempty"`
	ReplyAttrs  map[string]string `json:"reply_attrs,omitempty"`
}

// APIResponse: 通用 JSON 响应（health 等）
type APIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}
