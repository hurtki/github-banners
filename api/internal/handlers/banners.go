package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/hurtki/github-banners/api/internal/domain"
	longterm "github.com/hurtki/github-banners/api/internal/domain/long-term"
	"github.com/hurtki/github-banners/api/internal/domain/preview"
	"github.com/hurtki/github-banners/api/internal/logger"
)

type PreviewUsecase interface {
	GetPreview(ctx context.Context, username string, bannerType string) (*domain.Banner, error)
}

type BannersHandler struct {
	logger    logger.Logger
	preview   PreviewUsecase
	ltBanners LTBannersUsecase
}

func NewBannersHandler(logger logger.Logger, previewUsecase PreviewUsecase, ltBannersUsecase LTBannersUsecase) *BannersHandler {
	return &BannersHandler{
		logger:    logger.With("service", "banners-handler"),
		preview:   previewUsecase,
		ltBanners: ltBannersUsecase,
	}
}

func (h *BannersHandler) Preview(rw http.ResponseWriter, req *http.Request) {
	fn := "internal.handlers.BannersHandler.Preview"
	username := req.URL.Query().Get("username")
	bannerType := req.URL.Query().Get("type")

	banner, err := h.preview.GetPreview(req.Context(), username, bannerType)
	if err != nil {
		switch {
		case errors.Is(err, preview.ErrInvalidBannerType):
			h.error(rw, http.StatusBadRequest, err.Error())
		case errors.Is(err, preview.ErrUserDoesntExist):
			h.error(rw, http.StatusNotFound, err.Error())
		case errors.Is(err, preview.ErrInvalidInputs):
			h.error(rw, http.StatusBadRequest, err.Error())
		case errors.Is(err, preview.ErrCantGetPreview):
			h.error(rw, http.StatusInternalServerError, err.Error())
		default:
			h.logger.Warn("unhandled error from usecase", "source", fn, "err", err)
			h.error(rw, http.StatusInternalServerError, "can't get preview")
		}
		return
	}

	rw.Header().Add("Content-Type", "image/svg+xml")
	rw.WriteHeader(http.StatusOK)
	_, err = rw.Write(banner.Banner)
	if err != nil {
		h.logger.Error("can't write response's body", "source", fn, "err", err, "body_length", len(banner.Banner))
	}
}

type LTBannersUsecase interface {
	CreateBanner(ctx context.Context, in longterm.CreateBannerIn) (longterm.CreateBannerOut, error)
}

func (h *BannersHandler) Create(rw http.ResponseWriter, req *http.Request) {
	fn := "internal.handlers.BannersHandler.Create"
	reqDto := CreateBannerRequest{}
	defer req.Body.Close()
	if err := json.NewDecoder(req.Body).Decode(&reqDto); err != nil {
		h.error(rw, http.StatusBadRequest, "invalid json")
		return
	}
	out, err := h.ltBanners.CreateBanner(req.Context(), longterm.CreateBannerIn{
		Username:   reqDto.Username,
		BannerType: reqDto.BannerType,
	})
	if err != nil {
		switch {
		case errors.Is(err, longterm.ErrUserDoesntExist):
			h.error(rw, http.StatusNotFound, err.Error())
		case errors.Is(err, longterm.ErrInvalidBannerType):
			h.error(rw, http.StatusBadRequest, err.Error())
		case errors.Is(err, longterm.ErrCantCreateBanner):
			h.logger.Error("failed to create long-term banner", "source", fn, "err", err)
			h.error(rw, http.StatusInternalServerError, "Failed to process banner creation request")
		default:
			h.logger.Warn("unhandled error from usecase", "source", fn, "err", err)
			h.error(rw, http.StatusInternalServerError, "Failed to process banner creation request")
		}
		return
	}

	rw.Header().Add("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	resDto := CreateBannerResponse{BannerUrlPath: out.BannerUrlPath}
	err = json.NewEncoder(rw).Encode(resDto)
	if err != nil {
		h.logger.Error("can't encode response", "dto", resDto, "err", err, "source", fn)
		h.error(rw, http.StatusInternalServerError, "can't create banner")
	}
}
