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
	radiusServer  = "172.19.0.4:1812" // 认证端口
	acctServer    = "172.19.0.4:1813" // 计费端口
	sharedSecret  = "testing123"      // 与 FreeRADIUS 配置一致
	nasIdentifier = "docker-nas-01"   // NAS 标识符
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

// stopRadiusAccounting 发起计费停止请求 (Logout)
func stopRadiusAccounting(username string) error {
	packet := radius.New(radius.CodeAccountingRequest, []byte(sharedSecret))
	rfc2865.UserName_SetString(packet, username)
	rfc2865.NASIdentifier_SetString(packet, nasIdentifier)

	// 设置 Acct-Status-Type 为 Stop (2)
	rfc2866.AcctStatusType_Set(packet, rfc2866.AcctStatusType_Value_Stop)

	log.Printf("[RADIUS-Acct] 发送 Accounting-Request (Stop): 用户=%s", username)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
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

// portalDeauthHandler 处理用户下线请求 (Step: Deauthentication)
func portalDeauthHandler(w http.ResponseWriter, r *http.Request) {
	// 获取要下线的用户名
	username := r.FormValue("username")
	if username == "" {
		http.Error(w, "Deauth failed: username is required", http.StatusBadRequest)
		return
	}

	log.Printf("[NAS] 收到下线请求 (Deauth)，用户: %s", username)

	// 1. 发起 RADIUS Accounting-Stop (停止计费)
	// 在 RADIUS 协议中，下线必须伴随计费停止报文
	if err := stopRadiusAccounting(username); err != nil {
		log.Printf("[Error] RADIUS Accounting-Stop 失败: %v", err)
		// 即使计费报文失败，通常也要继续执行本地下线逻辑
	}

	// 2. 执行本地准入控制清理 (如：iptables -D ...)
	// 这里模拟清理该 IP 的放行规则
	log.Printf("[NAS] 已从内核中清除用户 %s 的放行规则", username)

	// 3. 告知 Portal 执行成功或重定向
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `
		<html>
		<head><meta http-equiv="refresh" content="2;url=http://127.0.0.1:8081/portal"></head>
		<body>
			<div style="text-align:center; margin-top:50px;">
				<h2 style="color: orange;">您已成功退出登录</h2>
				<p>正在断开网络连接并返回登录页...</p>
			</div>
		</body>
		</html>
	`)
}

// 在 main() 中注册:
// http.HandleFunc("/logout", logoutHandler)

func redirectToPortal(w http.ResponseWriter, r *http.Request) {
	// 路径必须是 /portal 而不是 /login，因为 /login 是处理 POST 的
	portalServerURL := "http://127.0.0.1:8081/portal"
	redirectURL := fmt.Sprintf("%s?nas_id=%s&user_ip=%s", portalServerURL, nasIdentifier, "127.0.0.1")
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

func main() {
	// 处理认证请求 (Step 8)
	http.HandleFunc("/portal_auth", portalAuthHandler)
	http.HandleFunc("/portal_deauth", portalDeauthHandler)

	// 模拟受保护的资源入口
	// 任何访问 /index.html 或根路径的请求都将被重定向
	http.HandleFunc("/index.html", redirectToPortal)
	http.HandleFunc("/", redirectToPortal)

	fmt.Println("接入设备 (NAS) 模拟器运行在 :8080...")
	log.Printf("认证重定向功能已开启 -> 目标: http://172.19.0.4:8081/login")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
