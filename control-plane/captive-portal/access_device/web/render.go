package web

import (
	"html/template"
	"net/http"
	"path/filepath"
)

type Renderer struct {
	templates *template.Template
}

// NewRenderer 初始化模板渲染器
func NewRenderer(tplDir string) (*Renderer, error) {
	tpls, err := template.ParseGlob(filepath.Join(tplDir, "*.html"))
	if err != nil {
		return nil, err
	}

	return &Renderer{
		templates: tpls,
	}, nil
}

// RenderResult 渲染认证结果页（成功 / 失败统一出口）
func (r *Renderer) RenderResult(w http.ResponseWriter, data ResultData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if err := r.templates.ExecuteTemplate(w, "layout.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// RenderLogin 渲染登录页
func (r *Renderer) RenderLogin(w http.ResponseWriter, data LoginData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if err := r.templates.ExecuteTemplate(w, "layout.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
