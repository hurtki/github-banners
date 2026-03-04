package events

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hurtki/github-banners/renderer/internal/domain/render"
	"github.com/hurtki/github-banners/renderer/internal/logger"
)

type BannerUpdateUsecase interface {
	ProcessBanner(ctx context.Context, req render.UpdateBannerIn) error
}

type UpdateBannerHandler struct {
	logger  logger.Logger
	usecase BannerUpdateUsecase
}

func NewBannerUpdateHandler(logger logger.Logger, usecase BannerUpdateUsecase) *UpdateBannerHandler {
	return &UpdateBannerHandler{
		logger:  logger.With("service", "update-banner-handler"),
		usecase: usecase,
	}
}

func (h *UpdateBannerHandler) Handle(ctx context.Context, msg Message) error {
	var event BannerUpdateEvent
	err := json.Unmarshal(msg.Value, &event)
	if err != nil {
		return fmt.Errorf("can't unmarshal msg's value as BannerUpdateEvent, %w: %w", err, ErrValidation)
	}

	h.logger.Debug("Handling new event", "key", string(msg.Key))

	updateIn := event.Payload.ToDomainInUpdateBannerIn()

	err = h.usecase.ProcessBanner(ctx, updateIn)
	if err != nil {
		switch {
		case errors.Is(err, render.ErrInvalidBannerType),
			errors.Is(err, render.ErrInvalidUsername),
			errors.Is(err, render.ErrInvalidUrlPath):
			return fmt.Errorf("%w:%w", err, ErrValidation)
		case errors.Is(err, render.ErrRenderFailure),
			errors.Is(err, render.ErrStorageFailure):
			return fmt.Errorf("%w:%w", err, ErrTransient)
		default:
			return fmt.Errorf("%w:%w", err, ErrBusiness)
		}
	}

	return nil
}
