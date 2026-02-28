package render

import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/hurtki/github-banners/renderer/internal/domain"
)

type Service struct {
	templates *template.Template
}

func NewRenderSerivce(tmpl *template.Template) *Service {
	return &Service{
		templates: tmpl,
	}
}

func (s *Service) Render(info domain.BannerInfo) ([]byte, error) {
	var buf bytes.Buffer
	templateName := fmt.Sprintf("%s.svg", info.BannerType)

	if err := s.templates.ExecuteTemplate(&buf, templateName, info); err != nil {
		return nil, ErrRenderFailure
	}

	return buf.Bytes(), nil
}
