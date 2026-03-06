package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/hurtki/github-banners/storage/internal/domain/banner"
	"github.com/hurtki/github-banners/storage/internal/logger"
)

type BannerSaveUsecase interface {
	Save(ctx context.Context, in banner.SaveIn) (banner.SaveOut, error)
}

type BannerSaveHandler struct {
	logger  logger.Logger
	usecase BannerSaveUsecase
}

func NewBannerSaveHandler(logger logger.Logger, usecase BannerSaveUsecase) *BannerSaveHandler {
	return &BannerSaveHandler{
		logger:  logger.With("service", "banner-save-handler"),
		usecase: usecase,
	}
}

func (h *BannerSaveHandler) Save(rw http.ResponseWriter, req *http.Request) {
	fn := "internal.handlers.BannerSaveHandler.Save"
	reqDto := SaveRequest{}
	err := json.NewDecoder(req.Body).Decode(&reqDto)
	if err != nil {
		h.error(rw, http.StatusBadRequest, "invalid json")
		return
	}
	in, err := reqDto.ToDomainSaveBannerIn()
	if err != nil {
		h.error(rw, http.StatusBadRequest, "invalid base64 encoding")
		return
	}

	out, err := h.usecase.Save(req.Context(), in)
	if err != nil {
		switch {
		case errors.Is(err, banner.ErrInvalidUrlPath):
			h.error(rw, http.StatusBadRequest, "invalid url path")
		case errors.Is(err, banner.ErrInvalidBannerFormat):
			h.error(rw, http.StatusBadRequest, "invalid banner format")
		case errors.Is(err, banner.ErrCantSaveBanner):
			h.logger.Warn("can't save banner", "err", err)
			h.error(rw, http.StatusInternalServerError, "can't save banner")
		default:
			h.logger.Warn("unhandled error from usecase", "source", fn, "err", err)
			h.error(rw, http.StatusInternalServerError, "can't save banner")
		}
		return
	}
	resDto := SaveResponse{URL: out.BannerUrl}
	err = json.NewEncoder(rw).Encode(resDto)
	if err != nil {
		h.logger.Error("can't encode response", "dto", resDto, "err", err)
		h.error(rw, http.StatusInternalServerError, "can't create banner")
	}
}
