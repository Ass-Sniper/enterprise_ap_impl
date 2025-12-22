package i18n

type Key string

const (
	// ===== 页面标题 =====
	TitleLogin Key = "title.login"

	// ===== 通用结果 =====
	MsgAuthSuccessRedirect Key = "msg.auth_success_redirect"

	// ===== 错误提示 =====
	ErrInvalidParams       Key = "error.invalid_params"
	ErrAuthServiceDown     Key = "error.auth_service_down"
	ErrAuthRespParseFailed Key = "error.auth_resp_parse_failed"
	ErrAuthFailed          Key = "error.auth_failed"
)

type Lang string

const (
	ZH_CN Lang = "zh-CN"
	EN_US Lang = "en-US"
)

var catalogs = map[Lang]map[Key]string{
	ZH_CN: zhCN,
	EN_US: enUS,
}

// T 翻译函数
func T(lang Lang, key Key) string {
	if cat, ok := catalogs[lang]; ok {
		if v, ok := cat[key]; ok {
			return v
		}
	}
	// fallback：key 本身（便于发现缺失翻译）
	return string(key)
}
