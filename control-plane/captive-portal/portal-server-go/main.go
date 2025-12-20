package main

import (
	"fmt"
	"net/http"
)

func main() {
	// 步骤5：返回登录页面
	http.HandleFunc("/portal", func(w http.ResponseWriter, r *http.Request) {
		clientIP := r.URL.Query().Get("client_ip")
		nasIP := r.URL.Query().Get("nas_ip")

		fmt.Fprintf(w, `
			<html>
			<head><title>Portal Login</title></head>
			<body>
				<h2>欢迎登录网络</h2>
				<form action="/login" method="post">
					<input type="hidden" name="nas_ip" value="%s">
					<input type="hidden" name="client_ip" value="%s">
					用户名: <input type="text" name="username"><br>
					密码: <input type="password" name="password"><br>
					<button type="submit">登录</button>
				</form>
			</body>
			</html>
		`, nasIP, clientIP)
	})

	// 步骤6 & 7：验证并通知客户端向接入设备发起认证
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			return
		}

		username := r.FormValue("username")
		nasIP := r.FormValue("nas_ip")
		// 模拟Portal服务器内部逻辑：验证用户合法性
		fmt.Printf("Portal验证用户: %s, 准备下发Token到接入设备: %s\n", username, nasIP)

		// 【关键步骤 7】：通过JS自动提交表单，通知客户端请求接入设备
		// 这里的 action 指向接入设备的认证地址
		authURL := fmt.Sprintf("http://%s:8080/portal_auth", nasIP)

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `
			<html>
			<body onload="document.forms[0].submit()">
				<p>正在跳转至接入设备进行准入认证...</p>
				<form action="%s" method="post">
					<input type="hidden" name="username" value="%s">
					<input type="hidden" name="token" value="mock_token_123456">
				</form>
			</body>
			</html>
		`, authURL, username)
	})

	fmt.Println("Portal Server 运行在 :8081...")
	http.ListenAndServe(":8081", nil)
}
