package i18n

import (
	"net/http"
	"strings"
)

// DetectLang 从 HTTP 请求中推断语言
// 优先级：
// 1. Accept-Language
// 2. 默认 zh-CN
func DetectLang(r *http.Request) Lang {

	// if c, err := r.Cookie("lang"); err == nil {
	// 	return Lang(c.Value)
	// }

	al := r.Header.Get("Accept-Language")
	if al == "" {
		return ZH_CN
	}

	al = strings.ToLower(al)

	if strings.HasPrefix(al, "en") {
		return EN_US
	}

	return ZH_CN
}
