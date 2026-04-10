package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

type FrontendHandler struct {
	templateFS fs.FS
	apiBaseURL string
	apiClient  *http.Client
}

type pageData struct {
	// shared
	Title        string
	Page         string
	Error        string
	Success      string
	CanonicalURL template.HTML // full canonical URL, pre-built; safe to use in href
	BaseURL      string

	// paste view
	PasteID            string
	PasteTitle         string
	Content            string
	HighlightedContent template.HTML // pre-rendered HTML from Chroma; empty falls back to Content
	Language           string
	CreatedAt          string
	ExpiresAt          string
	LineCount          int
	SizeStr            string
	DownloadName       string // suggested filename for <a download="...">
	BurnAfterReading   bool   // true = paste was deleted after this view

	// upload result
	PasteURL string
	RawURL   string
}

func formatSize(n int) string {
	switch {
	case n < 1024:
		return fmt.Sprintf("%d B", n)
	case n < 1024*1024:
		return fmt.Sprintf("%.1f KB", float64(n)/1024)
	default:
		return fmt.Sprintf("%.1f MB", float64(n)/(1024*1024))
	}
}

// downloadFilename returns a sensible filename for the browser's Save dialog.
func downloadFilename(title, language string) string {
	extMap := map[string]string{
		"go": ".go", "python": ".py", "javascript": ".js", "typescript": ".ts",
		"bash": ".sh", "sh": ".sh", "json": ".json", "yaml": ".yml", "toml": ".toml",
		"sql": ".sql", "rust": ".rs", "c": ".c", "cpp": ".cpp", "c++": ".cpp",
		"ruby": ".rb", "php": ".php", "java": ".java", "kotlin": ".kt",
		"html": ".html", "css": ".css", "xml": ".xml", "markdown": ".md",
		"dockerfile": "", "text": ".txt",
	}
	title = strings.TrimSpace(title)
	if title == "" || title == "paste" {
		title = "paste"
	}
	lang := strings.ToLower(language)
	if lang == "dockerfile" {
		return "Dockerfile"
	}
	ext, ok := extMap[lang]
	if !ok {
		ext = ".txt"
	}
	return title + ext
}

// formatTimestamp tries to parse an RFC3339 timestamp and return a human date.
func formatTimestamp(s string) string {
	if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
		return t.UTC().Format("Jan 2, 2006")
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t.UTC().Format("Jan 2, 2006")
	}
	return s
}

// highlight returns Chroma-generated HTML for the given code and language.
// Falls back to the plain text if lexer or formatting fails.
func highlight(code, language string) template.HTML {
	lexer := lexers.Get(language)
	if lexer == nil {
		lexer = lexers.Fallback
	}

	style := styles.Get("github-dark")
	if style == nil {
		style = styles.Fallback
	}

	formatter := chromahtml.New(
		chromahtml.WithClasses(false),
		chromahtml.WithLineNumbers(true),
		chromahtml.LineNumbersInTable(true), // line numbers in separate column — copy-paste skips them
	)

	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		return template.HTML(template.HTMLEscapeString(code))
	}

	var buf strings.Builder
	if err = formatter.Format(&buf, style, iterator); err != nil {
		return template.HTML(template.HTMLEscapeString(code))
	}
	return template.HTML(buf.String())
}

func NewFrontendHandler(templateFS fs.FS, apiBaseURL string) *FrontendHandler {
	return &FrontendHandler{
		templateFS: templateFS,
		apiBaseURL: strings.TrimRight(apiBaseURL, "/"),
		apiClient:  &http.Client{Timeout: 10 * time.Second},
	}
}

// render parses all *.html templates together so that every cross-reference
// inside layout.html (the if/else page dispatch) is always resolvable.
// html/template performs static security analysis across all branches of an
// if/else block — not just the executed one — so every referenced template
// must be present in the set at execution time.
func (fh *FrontendHandler) render(w http.ResponseWriter, r *http.Request, data pageData) {
	tmpl, err := template.ParseFS(fh.templateFS, "*.html")
	if err != nil {
		slog.ErrorContext(r.Context(), "template parse error",
			slog.String("page", data.Page),
			slog.String("error", err.Error()),
		)
		http.Error(w, "failed to process the page.", http.StatusInternalServerError)
		return
	}

	// Use a buffer so a template error doesn't leave the browser with a blank
	// half-written page. If execution succeeds, flush the buffer to the writer.
	var buf bytes.Buffer
	if err = tmpl.ExecuteTemplate(&buf, "layout.html", data); err != nil {
		slog.ErrorContext(r.Context(), "template execute error",
			slog.String("page", data.Page),
			slog.String("error", err.Error()),
		)
		http.Error(w, "internal template error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	buf.WriteTo(w)
}

// requestBaseURL returns the scheme+host of the incoming request,
// respecting X-Forwarded-Proto for reverse-proxy deployments.
func requestBaseURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
		scheme = proto
	}
	return scheme + "://" + r.Host
}

func (fh *FrontendHandler) Home(w http.ResponseWriter, r *http.Request) {
	fh.render(w, r, pageData{
		Title:        "pipebin.dev — minimal pastebin for developers",
		Page:         "home",
		BaseURL:      requestBaseURL(r),
		CanonicalURL: template.HTML(requestBaseURL(r) + "/"),
	})
}

// CreatePaste handles POST /.
// Supports three input modes:
//  1. Browser HTML form (application/x-www-form-urlencoded with a "content" field)
//  2. JSON body (application/json)
//  3. Raw pipe — echo "hello" | curl -sd @- https://pipebin.dev/
//     `curl -d` sends x-www-form-urlencoded but the body is raw text (no "content=" key),
//     so we detect that and treat the whole body as content.
//     Metadata via query params: ?t=title  ?lang=go  ?expires=24h
func (fh *FrontendHandler) CreatePaste(w http.ResponseWriter, r *http.Request) {
	ct := r.Header.Get("Content-Type")
	isCurl := strings.HasPrefix(r.UserAgent(), "curl/")

	var title, language, content, expiresAt string
	var burn bool

	switch {
	case strings.HasPrefix(ct, "application/json"):
		var req struct {
			Title            string `json:"title"`
			Content          string `json:"content"`
			Language         string `json:"language"`
			ExpiresAt        string `json:"expires_at"`
			BurnAfterReading bool   `json:"burn"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			slog.WarnContext(r.Context(), "create paste: JSON decode error", slog.String("error", err.Error()))
			fh.renderError(w, r, "Invalid JSON: "+err.Error())
			return
		}
		title, content, language, expiresAt, burn = req.Title, req.Content, req.Language, req.ExpiresAt, req.BurnAfterReading

	case strings.HasPrefix(ct, "application/x-www-form-urlencoded"):
		raw, err := io.ReadAll(r.Body)
		if err != nil {
			slog.WarnContext(r.Context(), "create paste: read body error", slog.String("error", err.Error()))
			fh.renderError(w, r, "Failed to read body.")
			return
		}
		if vals, err := url.ParseQuery(string(raw)); err == nil && vals.Get("content") != "" {
			// Genuine HTML form submission.
			title = vals.Get("title")
			content = vals.Get("content")
			language = vals.Get("language")
			expiresAt = vals.Get("expires_at")
			burn = vals.Get("burn") == "1"
		} else {
			// curl -d @- raw pipe: body IS the content.
			content = string(raw)
			q := r.URL.Query()
			title = q.Get("t")
			language = q.Get("lang")
			expiresAt = q.Get("expires")
			if expiresAt == "" {
				expiresAt = q.Get("e")
			}
			burn = q.Get("once") == "1" || q.Get("burn") == "1"
		}

	default:
		// No / unknown Content-Type — raw body (curl -T - or bare pipe).
		raw, err := io.ReadAll(r.Body)
		if err != nil {
			slog.WarnContext(r.Context(), "create paste: read body error", slog.String("error", err.Error()))
			fh.renderError(w, r, "Failed to read body.")
			return
		}
		content = string(raw)
		q := r.URL.Query()
		title = q.Get("t")
		language = q.Get("lang")
		expiresAt = q.Get("expires")
		if expiresAt == "" {
			expiresAt = q.Get("e")
		}
		burn = q.Get("once") == "1" || q.Get("burn") == "1"
	}

	if title == "" {
		title = "paste"
	}
	if language == "" {
		language = "text"
	}

	body, _ := json.Marshal(map[string]interface{}{
		"title":      title,
		"content":    content,
		"language":   language,
		"expires_at": expiresAt,
		"burn":       burn,
	})

	resp, err := fh.apiClient.Post(fh.apiBaseURL+"/", "application/json", bytes.NewReader(body))
	if err != nil {
		slog.ErrorContext(r.Context(), "create paste: API unreachable", slog.String("error", err.Error()))
		if isCurl {
			http.Error(w, "could not reach API\n", http.StatusBadGateway)
			return
		}
		fh.renderError(w, r, "Could not reach API: "+err.Error())
		return
	}
	defer resp.Body.Close()

	var apiResp struct {
		Data struct {
			URL  string `json:"url"`
			Burn bool   `json:"burn"`
		} `json:"data"`
		Error string `json:"error"`
	}
	if err = json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		slog.ErrorContext(r.Context(), "create paste: unexpected API response", slog.String("error", err.Error()))
		if isCurl {
			http.Error(w, "unexpected API response\n", http.StatusBadGateway)
			return
		}
		fh.renderError(w, r, "Unexpected API response.")
		return
	}
	if resp.StatusCode != http.StatusCreated || apiResp.Data.URL == "" {
		msg := apiResp.Error
		if msg == "" {
			msg = "Failed to create paste."
		}
		slog.WarnContext(r.Context(), "create paste: API error",
			slog.Int("status", resp.StatusCode),
			slog.String("error", msg),
		)
		if isCurl {
			http.Error(w, msg+"\n", resp.StatusCode)
			return
		}
		fh.renderError(w, r, msg)
		return
	}

	pasteURL := apiResp.Data.URL
	rawURL := strings.Replace(pasteURL, "/p/", "/raw/", 1)

	slog.InfoContext(r.Context(), "paste created",
		slog.String("title", title),
		slog.String("language", language),
		slog.String("url", pasteURL),
		slog.Bool("burn", burn),
		slog.Bool("curl", isCurl),
	)

	if isCurl {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, "%s\n", pasteURL)
		if burn {
			fmt.Fprintf(w, "⚠  burn after reading — deleted on first view\n")
		}
		return
	}

	fh.render(w, r, pageData{
		Title:            "created — pipebin.dev",
		Page:             "upload_result",
		BaseURL:          requestBaseURL(r),
		PasteTitle:       title,
		Language:         language,
		BurnAfterReading: burn,
		PasteURL:         pasteURL,
		RawURL:           rawURL,
	})
}

// Paste handles GET /p/{id} — fetches the paste from the API and renders it.
func (fh *FrontendHandler) Paste(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		fh.NotFound(w, r)
		return
	}

	resp, err := fh.apiClient.Get(fmt.Sprintf("%s/p/%s", fh.apiBaseURL, id))
	if err != nil {
		slog.ErrorContext(r.Context(), "paste view: API unreachable",
			slog.String("id", id),
			slog.String("error", err.Error()),
		)
		fh.renderError(w, r, "Could not reach API: "+err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		slog.DebugContext(r.Context(), "paste view: not found", slog.String("id", id))
		fh.NotFound(w, r)
		return
	}
	if resp.StatusCode == http.StatusGone {
		slog.InfoContext(r.Context(), "paste view: expired", slog.String("id", id))
		w.WriteHeader(http.StatusGone)
		fh.render(w, r, pageData{
			Title: "expired — pipebin.dev",
			Page:  "error",
			Error: "This paste has expired.",
		})
		return
	}
	if resp.StatusCode != http.StatusOK {
		slog.WarnContext(r.Context(), "paste view: unexpected API status",
			slog.String("id", id),
			slog.Int("status", resp.StatusCode),
		)
		fh.renderError(w, r, "Failed to retrieve paste.")
		return
	}

	var apiResp struct {
		Data struct {
			PublicID         string  `json:"public_id"`
			Title            string  `json:"title"`
			Content          string  `json:"content"`
			Language         string  `json:"language"`
			CreatedAt        string  `json:"created_at"`
			ExpiresAt        *string `json:"expires_at"`
			BurnAfterReading bool    `json:"burn_after_reading"`
		} `json:"data"`
	}
	if err = json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		slog.ErrorContext(r.Context(), "paste view: decode error",
			slog.String("id", id),
			slog.String("error", err.Error()),
		)
		fh.renderError(w, r, "Unexpected API response.")
		return
	}

	d := apiResp.Data
	expiresAt := ""
	if d.ExpiresAt != nil {
		expiresAt = formatTimestamp(*d.ExpiresAt)
	}

	lineCount := strings.Count(d.Content, "\n") + 1
	if strings.TrimSpace(d.Content) == "" {
		lineCount = 0
	}

	slog.DebugContext(r.Context(), "paste view: ok",
		slog.String("id", id),
		slog.Bool("burn", d.BurnAfterReading),
	)

	fh.render(w, r, pageData{
		Title:              d.Title + " — pipebin.dev",
		Page:               "paste",
		BaseURL:            requestBaseURL(r),
		CanonicalURL:       template.HTML(requestBaseURL(r) + "/p/" + id),
		PasteID:            d.PublicID,
		PasteTitle:         d.Title,
		Content:            d.Content,
		HighlightedContent: highlight(d.Content, d.Language),
		Language:           d.Language,
		CreatedAt:          formatTimestamp(d.CreatedAt),
		ExpiresAt:          expiresAt,
		LineCount:          lineCount,
		SizeStr:            formatSize(len(d.Content)),
		DownloadName:       downloadFilename(d.Title, d.Language),
		BurnAfterReading:   d.BurnAfterReading,
	})
}

// RawPaste handles GET /raw/{id} — returns paste content as plain text.
func (fh *FrontendHandler) RawPaste(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.NotFound(w, r)
		return
	}

	resp, err := fh.apiClient.Get(fmt.Sprintf("%s/p/%s", fh.apiBaseURL, id))
	if err != nil {
		slog.ErrorContext(r.Context(), "raw paste: API unreachable",
			slog.String("id", id),
			slog.String("error", err.Error()),
		)
		http.Error(w, "Could not reach API.", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusNotFound:
		http.NotFound(w, r)
		return
	case http.StatusGone:
		http.Error(w, "This paste has expired.", http.StatusGone)
		return
	}
	if resp.StatusCode != http.StatusOK {
		slog.WarnContext(r.Context(), "raw paste: unexpected API status",
			slog.String("id", id),
			slog.Int("status", resp.StatusCode),
		)
		http.Error(w, "Failed to retrieve paste.", http.StatusBadGateway)
		return
	}

	var apiResp struct {
		Data struct {
			Content string `json:"content"`
		} `json:"data"`
	}
	if err = json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		slog.ErrorContext(r.Context(), "raw paste: decode error",
			slog.String("id", id),
			slog.String("error", err.Error()),
		)
		http.Error(w, "Unexpected API response.", http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	fmt.Fprint(w, apiResp.Data.Content)
}

func (fh *FrontendHandler) NotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	fh.render(w, r, pageData{
		Title: "not found — pipebin.dev",
		Page:  "404",
	})
}

func (fh *FrontendHandler) renderError(w http.ResponseWriter, r *http.Request, message string) {
	w.WriteHeader(http.StatusInternalServerError)
	fh.render(w, r, pageData{
		Title: "error — pipebin.dev",
		Page:  "error",
		Error: message,
	})
}
