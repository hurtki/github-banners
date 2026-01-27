package handlers

import (
	"net/http"

	userstats "github.com/hurtki/github-banners/api/internal/domain/user_stats"
	log "github.com/hurtki/github-banners/api/internal/logger"
)

type BannersHandler struct {
	logger       log.Logger
	statsService *userstats.UserStatsService
}

func NewBannersHandler(logger log.Logger, statsService *userstats.UserStatsService) *BannersHandler {
	return &BannersHandler{
		logger:       logger.With("service", "banners-handler"),
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
