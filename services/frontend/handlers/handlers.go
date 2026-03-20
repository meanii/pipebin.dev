package handlers

import (
	"html/template"
	"io/fs"
	"net/http"
)

type FrontendHandler struct {
	templateFS fs.FS
	apiBaseURL string
}

type pageData struct {
	Title  string
	Error  string
	Sucess string
}

func NewFrontendHandler(template fs.FS, apiBaseURL string) *FrontendHandler {
	return &FrontendHandler{
		templateFS: template,
		apiBaseURL: apiBaseURL,
	}
}

func (fh *FrontendHandler) render(w http.ResponseWriter, page string, data pageData) {
	tmpl, err := template.ParseFS(fh.templateFS, "layout.html", page+".html")
	if err != nil {
		http.Error(w, "failed to process the page.", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err = tmpl.ExecuteTemplate(w, "layout.html", data); err != nil {
		http.Error(w, "failed to process the page.", http.StatusInternalServerError)
		return
	}
}

func (fh *FrontendHandler) Home(w http.ResponseWriter, r *http.Request) {
	fh.render(w, "home", pageData{Title: "pipebin.dev - developer friendly"})
}
