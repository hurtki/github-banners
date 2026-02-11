package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/hurtki/github-banners/api/internal/domain"
	"github.com/hurtki/github-banners/api/internal/domain/preview"
	log "github.com/hurtki/github-banners/api/internal/logger"
)

type PreviewUsecase interface {
	GetPreview(ctx context.Context, username string, bannerType string) (*domain.Banner, error)
}

type BannersHandler struct {
	logger  log.Logger
	preview PreviewUsecase
}

func NewBannersHandler(logger log.Logger, previewUsecase PreviewUsecase) *BannersHandler {
	return &BannersHandler{
		logger:  logger.With("service", "banners-handler"),
		preview: previewUsecase,
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
		return
	}
}

func (h *BannersHandler) Create(rw http.ResponseWriter, req *http.Request) {
	fn := "internal.handlers.BannersHandler.Create"
	h.logger.Info("Not filled handler", "source", fn)
	rw.WriteHeader(http.StatusNotImplemented)
}

// errror is used to write error in json
// if error, when marshaling appears, handles and logs it
func (h *BannersHandler) error(rw http.ResponseWriter, statusCode int, message string) {
	resJson := make(map[string]string)
	resJson["error"] = message
	res, err := json.Marshal(resJson)
	if err != nil {
		h.logger.Error("canl't marshal error to response")
		rw.WriteHeader(http.StatusInternalServerError)
		_, err := rw.Write([]byte("{\"error\": \"server error occured\"}"))
		if err != nil {
			h.logger.Warn("can't write error response")
			return
		}
	}
	rw.WriteHeader(statusCode)
	_, err = rw.Write(res)
	if err != nil {
		h.logger.Warn("can't write error response")
		return
	}
}
