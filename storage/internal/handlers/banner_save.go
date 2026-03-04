package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/hurtki/github-banners/storage/internal/domain/usecase"
	"github.com/hurtki/github-banners/storage/internal/logger"
)

type BannerSaveUsecase interface {
	Save(in usecase.SaveIn) (usecase.SaveOut, error)
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

	out, err := h.usecase.Save(in)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrInvalidUrlPath):
			h.error(rw, http.StatusBadRequest, "invalid url path")
		case errors.Is(err, usecase.ErrInvalidBannerFormat):
			h.error(rw, http.StatusBadRequest, "invalid banner format")
		case errors.Is(err, usecase.ErrCantSaveBanner):
			h.logger.Warn("can't save banner", "err", err)
			h.error(rw, http.StatusInternalServerError, "can't save banner")
			return
		}
	}
	resDto := SaveResponse{URL: out.BannerUrl}
	err = json.NewEncoder(rw).Encode(resDto)
	if err != nil {
		h.logger.Error("can't encode response", "dto", resDto, "err", err)
		h.error(rw, http.StatusInternalServerError, "can't create banner")
	}
}
