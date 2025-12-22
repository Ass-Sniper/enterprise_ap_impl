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

	log.Printf(
		"[PORTAL_UI] handleLogin start method=%s remote=%s",
		r.Method,
		r.RemoteAddr,
	)

	// ============================
	// 1️⃣ Parse Form
	// ============================
	if err := r.ParseForm(); err != nil {
		log.Printf(
			"[PORTAL_UI][ERROR] parse form failed err=%v",
			err,
		)

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

	log.Printf(
		"[PORTAL_UI] login submit user=%q password.len=%d",
		username,
		len(password),
	)

	// ============================
	// 2️⃣ Build AuthRequest
	// ============================
	req := api.AuthRequest{
		Username: username,
		Password: password,
		UserIP:   r.RemoteAddr,
	}

	body, err := json.Marshal(req)
	if err != nil {
		log.Printf(
			"[PORTAL_UI][ERROR] marshal AuthRequest failed user=%q err=%v",
			username,
			err,
		)

		renderer.RenderLogin(w, web.LoginData{
			BasePage: web.BasePage{
				Title: i18n.T(lang, i18n.TitleLogin),
			},
			Username: username,
			Error:    i18n.T(lang, i18n.ErrAuthRespParseFailed),
		})
		return
	}

	log.Printf(
		"[PORTAL_UI] POST /portal_auth -> %s payload=%s",
		accessDeviceURL,
		string(body),
	)

	// ============================
	// 3️⃣ Call AccessDevice (NAC)
	// ============================
	resp, err := http.Post(
		accessDeviceURL,
		"application/json",
		bytes.NewReader(body),
	)

	if err != nil {
		log.Printf(
			"[PORTAL_UI][ERROR] POST access device failed user=%q err=%v",
			username,
			err,
		)

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

	log.Printf(
		"[PORTAL_UI] access device response status=%d",
		resp.StatusCode,
	)

	// ============================
	// 4️⃣ Decode NAC Response
	// ============================
	var ar api.AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&ar); err != nil {
		log.Printf(
			"[PORTAL_UI][ERROR] decode AuthResponse failed user=%q err=%v",
			username,
			err,
		)

		renderer.RenderLogin(w, web.LoginData{
			BasePage: web.BasePage{
				Title: i18n.T(lang, i18n.TitleLogin),
			},
			Username: username,
			Error:    i18n.T(lang, i18n.ErrAuthRespParseFailed),
		})
		return
	}

	log.Printf(
		"[PORTAL_UI] auth response user=%q success=%v policy=%s strategy=%s redirect=%s",
		username,
		ar.Success,
		ar.Policy,
		ar.Strategy,
		ar.RedirectURL,
	)

	// ============================
	// 5️⃣ Auth Failed
	// ============================
	if !ar.Success {
		log.Printf(
			"[PORTAL_UI][DENY] user=%q msg=%q",
			username,
			ar.Message,
		)

		renderer.RenderLogin(w, web.LoginData{
			BasePage: web.BasePage{
				Title: i18n.T(lang, i18n.TitleLogin),
			},
			Username: username,
			Error:    ar.Message,
		})
		return
	}

	// ============================
	// 6️⃣ Auth Success → Result Page
	// ============================
	log.Printf(
		"[PORTAL_UI][SUCCESS] user=%q redirect=%s delay=2s",
		username,
		ar.RedirectURL,
	)

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
