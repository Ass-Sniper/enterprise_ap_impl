package audit

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"time"
)

func New(enabled bool, secret string) *Logger {
	return &Logger{
		Enabled: enabled,
		Secret:  []byte(secret),
	}
}

func (l *Logger) Sign(payload []byte) string {
	m := hmac.New(sha256.New, l.Secret)
	m.Write(payload)
	return hex.EncodeToString(m.Sum(nil))
}

func (l *Logger) Write(event map[string]any) {
	if !l.Enabled {
		return
	}
	if _, ok := event["ts"]; !ok {
		event["ts"] = time.Now().Unix()
	}
	// 先不带 sig 序列化
	tmp := make(map[string]any, len(event))
	for k, v := range event {
		if k == "sig" {
			continue
		}
		tmp[k] = v
	}
	b, _ := json.Marshal(tmp)
	sig := l.Sign(b)

	tmp["sig"] = sig
	out, _ := json.Marshal(tmp)
	// 写 stdout，docker logs 里就是 JSON（你现在就是这么看的）
	_, _ = os.Stdout.Write(append(out, '\n'))
}
