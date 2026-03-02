package templates

import (
	"bytes"
	"embed"
	"html/template"

	"github.com/hurtki/github-banners/renderer/internal/domain/render"
	"github.com/hurtki/github-banners/renderer/internal/layout"
)

//go:embed assets/*.svg
var bannerAssets embed.FS

type Renderer struct {
	tmpl *template.Template
}

func NewRenderer() (*Renderer, error) {
	t, err := template.ParseFS(bannerAssets, "assets/*.svg")
	if err != nil {
		return nil, err
	}
	return &Renderer{tmpl: t}, nil
}

func (r *Renderer) RenderBanner(view *layout.BannerView) ([]byte, error) {
	var buf bytes.Buffer
	templateName := "banner.svg"

	if err := r.tmpl.ExecuteTemplate(&buf, templateName, view); err != nil {
		return nil, render.ErrRenderFailure
	}
	return buf.Bytes(), nil
}
