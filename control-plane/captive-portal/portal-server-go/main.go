package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

const (
	/*
	   系统架构示意:
	   [Client] -> [NAS:8080] -> [Portal:8081]
	      ^           |             |
	      +-----------+-------------+
	*/
	nasAuthPort = "8080"
	serverPort  = ":8081"
	nasIP       = "172.19.0.1"
)

func handlePortal(w http.ResponseWriter, r *http.Request) {
	clientIP := r.URL.Query().Get("user_ip")
	nasID := r.URL.Query().Get("nas_id")

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `
		<html>
		<head><title>Portal Login</title><script src="https://cdn.tailwindcss.com"></script></head>
		<body class="bg-gray-100 flex items-center justify-center min-h-screen">
			<form action="/login" method="post" class="bg-white p-8 rounded-lg shadow-md w-80">
				<h2 class="text-xl font-bold mb-4 text-center">网络准入登录</h2>
				<input type="hidden" name="nas_id" value="%s">
				<input type="hidden" name="client_ip" value="%s">
				<input type="text" name="username" placeholder="用户名" class="w-full mb-3 p-2 border rounded">
				<input type="password" name="password" placeholder="密码" class="w-full mb-4 p-2 border rounded">
				<button type="submit" class="w-full bg-blue-600 text-white py-2 rounded hover:bg-blue-700">登录</button>
			</form>
		</body>
		</html>
	`, nasID, clientIP)
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/portal", http.StatusSeeOther)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// 1. 模拟验证逻辑
	if username == "" || password == "" {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, FailurePageTemplate, CommonCSS, "用户名和密码不能为空", FailureJS)
		return
	}

	// 2. 验证成功：生成 Token 并组装页面
	token := fmt.Sprintf("auth_token_%d", time.Now().Unix())
	authURL := fmt.Sprintf("http://%s:%s/portal_auth", nasIP, nasAuthPort)
	deauthURL := fmt.Sprintf("http://%s:%s/portal_deauth", nasIP, nasAuthPort)

	log.Printf("[Portal] 用户 %s 登录成功, 下发 Token: %s", username, token)

	fmt.Fprintf(w, SuccessPageTemplate,
		CommonCSS, // %1
		username,  // %2
		authURL,   // %3
		username,  // %4
		token,     // %5
		deauthURL, // %6
		username,  // %7
		SuccessJS, // %8
	)
}

func main() {
	http.HandleFunc("/portal", handlePortal)
	http.HandleFunc("/login", handleLogin)

	log.Printf("Portal Server 运行在 %s", serverPort)
	if err := http.ListenAndServe(serverPort, nil); err != nil {
		log.Fatal(err)
	}
}
