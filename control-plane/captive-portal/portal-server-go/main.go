package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"

	"portal-server-go/web"
	"portal-server-go/web/api"
	"portal-server-go/web/i18n"
)

const (
	serverAddr      = ":8080"
	accessDeviceURL = "http://172.19.0.1:9000/portal_auth"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)

	renderer, err := web.NewRenderer("web/templates")
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			// 展示登录页
			lang := i18n.DetectLang(r)
			renderer.RenderLogin(w, web.LoginData{
				BasePage: web.BasePage{
					Title: i18n.T(lang, i18n.TitleLogin),
				},
			})

		case http.MethodPost:
			// 处理登录提交
			handleLogin(w, r, renderer)

		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})

	log.Printf("[PORTAL] UI listening on %s", serverAddr)
	log.Fatal(http.ListenAndServe(serverAddr, nil))
}

// ========================
// Handlers
// ========================

func handleLogin(w http.ResponseWriter, r *http.Request, renderer *web.Renderer) {
	lang := i18n.DetectLang(r)
	if err := r.ParseForm(); err != nil {
		renderer.RenderLogin(w, web.LoginData{
			BasePage: web.BasePage{
				Title: i18n.T(lang, i18n.TitleLogin),
			},
			Error: i18n.T(lang, i18n.ErrInvalidParams),
		})
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	req := api.AuthRequest{
		Username: username,
		Password: password,
		UserIP:   r.RemoteAddr,
	}

	body, _ := json.Marshal(req)

	log.Printf("handleLogin req: %v", req)
	log.Printf("handleLogin body: %v", body)
	log.Printf("handleLogin POST to accessDeviceURL: %s", accessDeviceURL)
	resp, err := http.Post(
		accessDeviceURL,
		"application/json",
		bytes.NewReader(body),
	)
	log.Printf("handleLogin POST complete err: %v", err)
	if err != nil {
		renderer.RenderLogin(w, web.LoginData{
			BasePage: web.BasePage{
				Title: i18n.T(lang, i18n.TitleLogin),
			},
			Username: username,
			Error:    i18n.T(lang, i18n.ErrAuthServiceDown),
		})
		return
	}
	defer resp.Body.Close()

	var ar api.AuthResponse
	log.Printf("handleLogin Decode response")
	if err := json.NewDecoder(resp.Body).Decode(&ar); err != nil {
		log.Printf("handleLogin Decoded response: %v", ar)
		renderer.RenderLogin(w, web.LoginData{
			BasePage: web.BasePage{
				Title: i18n.T(lang, i18n.TitleLogin),
			},
			Username: username,
			Error:    i18n.T(lang, i18n.ErrAuthRespParseFailed),
		})
		return
	}

	if !ar.Success {
		renderer.RenderLogin(w, web.LoginData{
			BasePage: web.BasePage{
				Title: i18n.T(lang, i18n.TitleLogin),
			},
			Username: username,
			Error:    ar.Message,
		})
		return
	}

	renderer.RenderResult(w, web.ResultData{
		BasePage: web.BasePage{
			Title:         i18n.T(lang, i18n.TitleLogin),
			RedirectURL:   ar.RedirectURL,
			RedirectDelay: 2,
		},
		Success: true,
		Message: i18n.T(lang, i18n.MsgAuthSuccessRedirect),
	})
}
