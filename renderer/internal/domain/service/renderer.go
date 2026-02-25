package service

import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/hurtki/github-banners/renderer/internal/domain"
	"github.com/hurtki/github-banners/renderer/internal/logger"
)

type RenderService struct {
	templates *template.Template
	log       logger.Logger
}

func NewRenderService(tmpl *template.Template, log logger.Logger) *RenderService {
	return &RenderService{
		templates: tmpl,
		log:       log.With("service", "render-service"),
	}
}

func (s *RenderService) Render(info domain.BannerInfo) (domain.RenderedBanner, error) {
	fn := "service.RenderService.Render"

	s.log.Debug("starting banner rendering", "source", fn, "username", info.Username, "banner_type", info.BannerType)

	var buf bytes.Buffer
	templateName := fmt.Sprintf("%s.svg", info.BannerType)

	if err := s.templates.ExecuteTemplate(&buf, templateName, info); err != nil {
		s.log.Error("failed to execute template", "source", fn, "err", err, "template_name", templateName)
		return domain.RenderedBanner{}, domain.ErrRenderFailure
	}

	filename := fmt.Sprintf("%s_%s.svg", info.Username, info.BannerType)
	return domain.RenderedBanner{
		Filename: filename,
		Data:     buf.Bytes(),
	}, nil
}
