package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hurtki/github-banners/renderer/internal/domain"
	"github.com/hurtki/github-banners/renderer/internal/logger"
)

type BannerUpdateUsecase interface {
	Update(context.Context, domain.BannerInfo) error
}

type BannerUpdateHandler struct {
	logger logger.Logger
	// usecase BannerUpdateUsecase
}

func NewBannerUpdateHandler(logger logger.Logger) *BannerUpdateHandler {
	return &BannerUpdateHandler{
		logger: logger.With("service", "banner-update-handler"),
	}
}

func (h *BannerUpdateHandler) Handle(ctx context.Context, msg Message) error {
	var event BannerUpdateEvent
	err := json.Unmarshal(msg.Value, &event)
	if err != nil {
		return fmt.Errorf("can't unmarshal msg's value as BannerUpdateEvent, %w", err)
	}

	h.logger.Info("Handling new event, key: %s value: $s", msg.Key, event.String())

	// return h.usecase.Update(ctx, domain.BannerInfo{})
	return nil
}
