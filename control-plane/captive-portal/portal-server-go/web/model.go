package web

// LoginData 用于登录页面
type LoginData struct {
	BasePage

	Username string
	Error    string
}

// ResultData 用于结果页面（成功 / 失败）
type ResultData struct {
	BasePage

	Success bool
	Message string
}
