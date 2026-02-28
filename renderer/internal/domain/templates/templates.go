package templates

import (
	"embed"
	"html/template"
)

//go:embed *.svg
var bannerFiles embed.FS

func Load() (*template.Template, error) {
	return template.ParseFS(bannerFiles, "*.svg")
}
