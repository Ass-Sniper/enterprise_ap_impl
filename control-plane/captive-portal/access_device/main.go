package main

import (
	"fmt"
	"net/http"
)

const PORTAL_SVR = "http://localhost:8081/portal"
const NAS_IP = "localhost"

func main() {
	// 步骤3：HTTP重定向 (模拟拦截未认证流量)
	http.HandleFunc("/index.html", func(w http.ResponseWriter, r *http.Request) {
		clientIP := r.RemoteAddr
		redirectURL := fmt.Sprintf("%s?client_ip=%s&nas_ip=%s", PORTAL_SVR, clientIP, NAS_IP)
		fmt.Printf("拦截到未认证请求，重定向至: %s\n", redirectURL)
		http.Redirect(w, r, redirectURL, http.StatusFound)
	})

	// 步骤8 & 13：接收客户端发起的Portal认证请求
	http.HandleFunc("/portal_auth", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			return
		}

		username := r.FormValue("username")
		token := r.FormValue("token")

		fmt.Printf("收到客户端认证请求: 用户=%s, Token=%s\n", username, token)

		// 步骤9-12：模拟与 RADIUS 服务器的交互
		fmt.Println("步骤9: 发送 RADIUS 认证请求 (Access-Request)...")
		fmt.Println("步骤10: 收到 RADIUS 认证成功 (Access-Accept)")
		fmt.Println("步骤11: 发送 RADIUS 计费开始 (Accounting-Request)...")
		fmt.Println("步骤12: 收到 RADIUS 计费响应")

		// 步骤13：告知用户认证结果
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, "<h1>认证成功！</h1><p>欢迎, %s。你现在可以访问互联网了。</p>", username)
	})

	fmt.Println("接入设备(NAS) 模拟运行在 :8080...")
	http.ListenAndServe(":8080", nil)
}
