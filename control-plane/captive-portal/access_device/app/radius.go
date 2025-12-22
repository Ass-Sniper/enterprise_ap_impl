package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc2866"
)

const (
	radiusAuthServer   = "172.19.0.4:1812"
	radiusAcctServer   = "172.19.0.4:1813"
	radiusSharedSecret = "testing123"
	nasIdentifier      = "docker-nas-01"
)

// RadiusPAPAuth implements auth.Authenticator
// - 返回 ok
// - replyAttrs: 从 Access-Accept / Reject 里提取的字段（Filter-Id, Session-Timeout...）
// - sessionTimeout: 秒，0=无
type RadiusPAPAuth struct{}

func (a *RadiusPAPAuth) RadiusPAP(ctx context.Context, username, password string) (bool, map[string]string, int64, error) {
	secret := []byte(radiusSharedSecret)
	packet := radius.New(radius.CodeAccessRequest, secret)

	rfc2865.UserName_SetString(packet, username)
	rfc2865.UserPassword_SetString(packet, password)
	rfc2865.NASIdentifier_SetString(packet, nasIdentifier)

	// // Message-Authenticator：如果你已关闭 require_message_authenticator，可不加
	// // 如果开启要求，此处应正确设置（否则报 malformed / insecure）
	// _ = rfc2869.MessageAuthenticator_Add(packet, secret)

	log.Printf("[RADIUS] request start user=%s server=%s\n", username, radiusAuthServer)

	// 使用传入 ctx（portal handler 的 ctx），但确保有超时
	rctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	start := time.Now()
	response, err := radius.Exchange(rctx, packet, radiusAuthServer)
	elapsed := time.Since(start)

	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			log.Printf("[RADIUS][TIMEOUT] user=%s server=%s elapsed=%s\n", username, radiusAuthServer, elapsed)
		} else {
			log.Printf("[RADIUS][ERROR] user=%s server=%s elapsed=%s err=%v\n", username, radiusAuthServer, elapsed, err)
		}
		return false, nil, 0, err
	}

	reply := extractReplyAttrs(response)
	st := extractSessionTimeout(response)

	switch response.Code {
	case radius.CodeAccessAccept:
		log.Printf("[RADIUS][ACCEPT] user=%s server=%s elapsed=%s\n", username, radiusAuthServer, elapsed)
		return true, reply, st, nil
	case radius.CodeAccessReject:
		log.Printf("[RADIUS][REJECT] user=%s server=%s elapsed=%s\n", username, radiusAuthServer, elapsed)
		return false, reply, st, nil
	default:
		log.Printf("[RADIUS][UNKNOWN] user=%s server=%s elapsed=%s code=%v\n", username, radiusAuthServer, elapsed, response.Code)
		return false, reply, st, fmt.Errorf("unexpected RADIUS response code: %v", response.Code)
	}
}

func extractReplyAttrs(p *radius.Packet) map[string]string {
	out := map[string]string{}

	// Filter-Id
	if s, err := rfc2865.FilterID_LookupString(p); err == nil && s != "" {
		out["Filter-Id"] = s
	}
	// Session-Timeout
	if v, err := rfc2865.SessionTimeout_Lookup(p); err == nil && v > 0 {
		out["Session-Timeout"] = fmt.Sprintf("%d", v)
	}

	return out
}

func extractSessionTimeout(p *radius.Packet) int64 {
	if v, err := rfc2865.SessionTimeout_Lookup(p); err == nil && v > 0 {
		return int64(v)
	}
	return 0
}

// Accounting Start（保留你原日志）
func startAccounting(username string) error {
	packet := radius.New(radius.CodeAccountingRequest, []byte(radiusSharedSecret))
	rfc2865.UserName_SetString(packet, username)
	rfc2865.NASIdentifier_SetString(packet, nasIdentifier)
	rfc2866.AcctStatusType_Set(packet, rfc2866.AcctStatusType_Value_Start)

	log.Printf("[RADIUS] 正在为用户 %s 发起计费开始请求...\n", username)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := radius.Exchange(ctx, packet, radiusAcctServer)
	return err
}
