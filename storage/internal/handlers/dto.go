package handlers

import (
	"encoding/base64"

	"github.com/hurtki/github-banners/storage/internal/domain/usecase"
)

type SaveRequest struct {
	UrlPath       string `json:"url_path"`
	B64BannerData string `json:"banner_info"`
	BannerFormat  string `json:"banner_format"`
}

func (r SaveRequest) ToDomainSaveBannerIn() (usecase.SaveIn, error) {
	decoded, err := base64.StdEncoding.DecodeString(r.B64BannerData)
	if err != nil {
		return usecase.SaveIn{}, err
	}

	return usecase.SaveIn{
		UrlPath:    r.UrlPath,
		BannerData: decoded,
		Format:     r.BannerFormat,
	}, nil
}

type SaveResponse struct {
	URL string `json:"url"`
}
