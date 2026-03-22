package handlers

import (
	"html/template"
	"io/fs"
	"log"
	"net/http"
)

type FrontendHandler struct {
	templateFS fs.FS
	apiBaseURL string
}

type pageData struct {
	// shared
	Title   string
	Page    string
	Error   string
	Success string

	// paste view
	PasteID    string
	PasteTitle string
	Content    string
	Language   string
	CreatedAt  string
	ExpiresAt  string

	// upload result
	PasteURL string
	RawURL   string
}

func NewFrontendHandler(templateFS fs.FS, apiBaseURL string) *FrontendHandler {
	return &FrontendHandler{
		templateFS: templateFS,
		apiBaseURL: apiBaseURL,
	}
}

// render parses all *.html templates together so that every cross-reference
// inside layout.html (the if/else page dispatch) is always resolvable.
// html/template performs static security analysis across all branches of an
// if/else block — not just the executed one — so every referenced template
// must be present in the set at execution time.
func (fh *FrontendHandler) render(w http.ResponseWriter, page string, data pageData) {
	tmpl, err := template.ParseFS(fh.templateFS, "*.html")
	if err != nil {
		log.Printf("template parse error (page=%s): %v", page, err)
		http.Error(w, "failed to process the page.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if err = tmpl.ExecuteTemplate(w, "layout.html", data); err != nil {
		log.Printf("template execute error (page=%s): %v", page, err)
		http.Error(w, "failed to process the page.", http.StatusInternalServerError)
		return
	}
}

func (fh *FrontendHandler) Home(w http.ResponseWriter, r *http.Request) {
	fh.render(w, "home", pageData{
		Title: "pipebin.dev — minimal pastebin for developers",
		Page:  "home",
	})
}

func (fh *FrontendHandler) Paste(w http.ResponseWriter, r *http.Request) {
	// TODO: fetch paste from API and populate fields.
	fh.render(w, "paste", pageData{
		Title: "paste — pipebin.dev",
		Page:  "paste",
	})
}

func (fh *FrontendHandler) UploadResult(w http.ResponseWriter, r *http.Request) {
	// TODO: populate PasteURL and RawURL from the API response.
	fh.render(w, "upload_result", pageData{
		Title: "created — pipebin.dev",
		Page:  "upload_result",
	})
}

func (fh *FrontendHandler) NotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	fh.render(w, "404", pageData{
		Title: "not found — pipebin.dev",
		Page:  "404",
	})
}

func (fh *FrontendHandler) Error(w http.ResponseWriter, r *http.Request, message string) {
	w.WriteHeader(http.StatusInternalServerError)
	fh.render(w, "error", pageData{
		Title: "error — pipebin.dev",
		Page:  "error",
		Error: message,
	})
}
