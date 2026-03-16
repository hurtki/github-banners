package http_handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/hurtki/github-banners/renderer/internal/domain/render"
	"github.com/hurtki/github-banners/renderer/internal/logger"
)

type PreviewUsecase interface {
	Render(ctx context.Context, in render.RenderIn) ([]byte, error)
}

type PreviewHandler struct {
	logger  logger.Logger
	usecase PreviewUsecase
}

func NewPreviewHandler(logger logger.Logger, uc PreviewUsecase) *PreviewHandler {
	return &PreviewHandler{
		logger:  logger.With("service", "preview-http-handler"),
		usecase: uc,
	}
}

func (h *PreviewHandler) Preview(rw http.ResponseWriter, r *http.Request) {
	fn := "internal.handlers.http.PreviewHandler.Preview"
	var req PreviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.error(rw, http.StatusBadRequest, "Invalid request body format")
		return
	}
	defer r.Body.Close()
	h.logger.Debug("Received preview payload", "username", req.Username, "banner_type", req.BannerType)

	renderIn := req.ToDomainRenderIn()

	svgBytes, err := h.usecase.Render(r.Context(), renderIn)
	if err != nil {
		if errors.Is(err, render.ErrInvalidUsername) || errors.Is(err, render.ErrInvalidBannerType) {
			h.error(rw, http.StatusBadRequest, err.Error())
			return
		}
		h.logger.Error("failed to render preview banner", "err", err, "username", req.Username)
		h.error(rw, http.StatusInternalServerError, "Internal server error")
		return
	}

	rw.Header().Set("Content-Type", "image/svg+xml")
	rw.WriteHeader(http.StatusOK)
	_, err = rw.Write(svgBytes)
	if err != nil {
		h.logger.Warn("can't write svg response", "err", err, "source", fn)
	}
}

func (h *PreviewHandler) error(rw http.ResponseWriter, statusCode int, message string) {
	fn := "internal.handlers.http.PreviewHandler.error"
	rw.Header().Set("Content-Type", "application/json")

	res, err := json.Marshal(map[string]string{"error": message})
	if err != nil {
		h.logger.Error("can't marshal error response", "err", err, "source", fn)
		rw.WriteHeader(http.StatusInternalServerError)
		_, err := rw.Write([]byte("{\"error\": \"server error occurred\"}"))
		if err != nil {
			h.logger.Warn("can't write error response", "err", err, "source", fn)
		}
		return
	}

	rw.WriteHeader(statusCode)
	_, err = rw.Write(res)
	if err != nil {
		h.logger.Warn("can't write error response", "err", err, "source", fn)
	}
}
