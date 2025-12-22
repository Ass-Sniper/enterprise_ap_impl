package web

// BasePage 是所有页面共享的最小公共字段集合
// layout.html 只能访问这里的字段，保证类型安全
type BasePage struct {
	// 页面标题
	Title string

	// 成功页可选：自动跳转 URL
	RedirectURL string

	// 成功页可选：自动跳转延迟（秒）
	RedirectDelay int
}
