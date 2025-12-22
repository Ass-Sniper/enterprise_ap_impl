package web

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

type Renderer struct {
	tpls *template.Template
}

// NewRenderer 解析所有模板文件
func NewRenderer(dir string) (*Renderer, error) {
	tpls, err := template.ParseGlob(filepath.Join(dir, "*.html"))
	if err != nil {
		return nil, err
	}

	log.Printf("[RENDER] templates loaded from %s", dir)
	for _, t := range tpls.Templates() {
		log.Printf("[RENDER] template registered: %s", t.Name())
	}

	return &Renderer{tpls: tpls}, nil
}

// render 执行指定的入口模板（login / result）
func (r *Renderer) render(w http.ResponseWriter, tpl string, data any) {
	log.Printf("[RENDER] execute template=%s data=%T %+v", tpl, data, data)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if err := r.tpls.ExecuteTemplate(w, tpl, data); err != nil {
		log.Printf("[RENDER][ERROR] template=%s err=%v", tpl, err)
	}
}

// RenderLogin 渲染登录页
func (r *Renderer) RenderLogin(w http.ResponseWriter, data LoginData) {
	log.Printf("[RENDER] RenderLogin")
	r.render(w, "login", data)
}

// RenderResult 渲染结果页（成功 / 失败）
func (r *Renderer) RenderResult(w http.ResponseWriter, data ResultData) {
	log.Printf("[RENDER] RenderResult success=%v", data.Success)
	r.render(w, "result", data)
}
