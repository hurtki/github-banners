package handlers

import (
	"net/http"

	"github.com/hurtki/github-banners/api/internal/domain/services"
	log "github.com/hurtki/github-banners/api/internal/logger"
)

type BannersHandler struct {
	logger       log.Logger
	statsService *services.StatsService
}

func NewBannersHandler(logger log.Logger, statsService *services.StatsService) *BannersHandler {
	return &BannersHandler{
		logger:       logger.With("service", "banners handler"),
		statsService: statsService,
	}
}

func (h *BannersHandler) Create(rw http.ResponseWriter, req *http.Request) {
	fn := "api.internal.handlers.BannersHandler.Create"
	h.logger.Info("Not filled handler", "source", fn)
	rw.WriteHeader(http.StatusNotImplemented)
}

func (h *BannersHandler) Preview(rw http.ResponseWriter, req *http.Request) {
	fn := "api.internal.handlers.BannersHandler.Preview"
	h.logger.Info("Not filled handler", "source", fn)
	rw.WriteHeader(http.StatusNotImplemented)
}
