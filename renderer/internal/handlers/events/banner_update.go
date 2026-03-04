package events

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hurtki/github-banners/renderer/internal/domain"
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
		return fmt.Errorf("can't unmarshal msg's value as BannerUpdateEvent, %w", err)
	}

	h.logger.Debug("Handling new event", "key", msg.Key)

	payload := event.Payload
	updateIn := render.UpdateBannerIn{
		Username:   payload.Username,
		BannerType: payload.BannerType,
		URLPath:    payload.StoragePath,
		Stats: domain.GithubUserStats{
			TotalRepos:    payload.Stats.TotalRepos,
			TotalStars:    payload.Stats.TotalStars,
			TotalForks:    payload.Stats.TotalForks,
			ForkedRepos:   payload.Stats.ForkedRepos,
			OriginalRepos: payload.Stats.OriginalRepos,
			Languages:     payload.Stats.Languages,
			FetchedAt:     payload.FetchedAt,
		},
	}

	err = h.usecase.ProcessBanner(ctx, updateIn)
	if err != nil {
		if errors.Is(err, render.ErrInvalidBannerType) || errors.Is(err, render.ErrInvalidUsername) || errors.Is(err, render.ErrInvalidUrlPath) {
			return fmt.Errorf("%w:%w", err, ErrValidation)
		}
		return fmt.Errorf("%w:%w", err, ErrTransient)
	}

	return nil
}
