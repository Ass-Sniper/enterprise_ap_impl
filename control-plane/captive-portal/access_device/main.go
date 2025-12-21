package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc2866" // 导入计费协议相关的包
)

const (
	radiusServer  = "127.0.0.1:1812" // 认证端口
	acctServer    = "127.0.0.1:1813" // 计费端口
	sharedSecret  = "testing123"     // 与 FreeRADIUS 配置一致
	nasIdentifier = "enterprise-ap-01"
)

// 步骤 9 & 10: 发起 RADIUS 认证请求
func authenticateUser(username, password string) (bool, error) {
	packet := radius.New(radius.CodeAccessRequest, []byte(sharedSecret))
	rfc2865.UserName_SetString(packet, username)
	rfc2865.UserPassword_SetString(packet, password)
	rfc2865.NASIdentifier_SetString(packet, nasIdentifier)

	fmt.Printf("[RADIUS] 正在为用户 %s 发起认证请求...\n", username)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	response, err := radius.Exchange(ctx, packet, radiusServer)
	if err != nil {
		return false, err
	}

	if response.Code == radius.CodeAccessAccept {
		fmt.Printf("[RADIUS] 用户 %s 认证通过 (Access-Accept)\n", username)
		return true, nil
	}

	fmt.Printf("[RADIUS] 用户 %s 认证失败 (Code: %v)\n", username, response.Code)
	return false, nil
}

// 步骤 11 & 12: 发起 RADIUS 计费请求 (Start)
func startAccounting(username string) error {
	packet := radius.New(radius.CodeAccountingRequest, []byte(sharedSecret))

	rfc2865.UserName_SetString(packet, username)
	rfc2865.NASIdentifier_SetString(packet, nasIdentifier)

	// 修正点：正确的常量名称是 AcctStatusType_Value_Start
	rfc2866.AcctStatusType_Set(packet, rfc2866.AcctStatusType_Value_Start)

	fmt.Printf("[RADIUS] 正在为用户 %s 发起计费开始请求...\n", username)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := radius.Exchange(ctx, packet, acctServer)
	return err
}

func portalAuthHandler(w http.ResponseWriter, r *http.Request) {
	// 步骤 8: 接收浏览器提交的参数 (支持 POST 或 302 跳转后的 GET)
	username := r.FormValue("username")
	token := r.FormValue("token") // 实际场景中 token 应映射为密码或临时凭证

	if username == "" {
		http.Error(w, "Missing username", http.StatusBadRequest)
		return
	}

	// 模拟将 Token 作为密码进行 RADIUS 认证
	authSuccess, err := authenticateUser(username, token)
	if err != nil {
		fmt.Printf("RADIUS 错误: %v\n", err)
		http.Error(w, "RADIUS Server Unreachable", http.StatusServiceUnavailable)
		return
	}

	if !authSuccess {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, "<h1>认证失败</h1><p>用户名或 Token 错误。</p>")
		return
	}

	// 步骤 11: 认证成功后，启动计费
	err = startAccounting(username)
	if err != nil {
		fmt.Printf("计费失败: %v\n", err)
		// 计费失败通常也应视为准入失败
		http.Error(w, "Accounting Failed", http.StatusInternalServerError)
		return
	}

	// 步骤 13: 告知用户成功
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, "<h1>Portal 认证成功</h1><p>欢迎, %s！您的网络连接已激活并开始计费。</p>", username)
	fmt.Printf("用户 %s 已完成 AAA 全流程。\n", username)
}

func main() {
	http.HandleFunc("/portal_auth", portalAuthHandler)

	// 模拟提供一个简单的登录页入口
	http.HandleFunc("/index.html", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "<html><body><h1>Access Device</h1><p>Please wait for Portal redirect...</p></body></html>")
	})

	fmt.Println("接入设备 (NAS) 模拟器运行在 :8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
