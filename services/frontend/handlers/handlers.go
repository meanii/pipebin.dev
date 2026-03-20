package handlers

import (
	"io/fs"
)

type FrontendHandler struct {
	templateFS fs.FS
	apiBaseURL string
}

func NewFrontendHandler(template fs.FS, apiBaseURL string) *FrontendHandler {
	return &FrontendHandler{
		templateFS: template,
		apiBaseURL: apiBaseURL,
	}
}

func (fh *FrontendHandler) render(page string) {

}
