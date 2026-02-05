package preview

import (
	"context"
	"testing"

	"github.com/hurtki/github-banners/api/internal/domain"
	"github.com/hurtki/github-banners/api/internal/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestGetPreviewSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockRenderer := mocks.NewMockPreviewRenderer(ctrl)
	mockStatsService := mocks.NewMockStatsService(ctrl)
	usecase := NewPreviewUsecase(mockStatsService, mockRenderer)
	username := "hurtki"
	bnrType := "wide"
	userStats := domain.GithubUserStats{
		TotalRepos: 4,
		TotalStars: 3,
	}
	bannerExpect := &domain.Banner{Username: username, BannerType: domain.BannerTypes[bnrType], Banner: []byte("<svg></svg>")}

	ctx := context.Background()

	mockStatsService.EXPECT().GetStats(ctx, username).Return(userStats, nil)

	mockRenderer.EXPECT().RenderPreview(ctx, domain.BannerInfo{
		Username:   username,
		BannerType: domain.BannerTypes[bnrType],
		Stats:      userStats,
	}).Return(bannerExpect, nil)

	bnr, err := usecase.GetPreview(context.Background(), username, bnrType)
	require.NoError(t, err)
	require.Equal(t, bannerExpect, bnr)
}

func TestGetPreviewUserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockRenderer := mocks.NewMockPreviewRenderer(ctrl)
	mockStatsService := mocks.NewMockStatsService(ctrl)
	usecase := NewPreviewUsecase(mockStatsService, mockRenderer)
	username := "hurtki324"
	bnrType := "wide"
	ctx := context.Background()

	mockStatsService.EXPECT().GetStats(ctx, username).Return(domain.GithubUserStats{}, domain.ErrNotFound)

	bnr, err := usecase.GetPreview(context.Background(), username, bnrType)
	require.Nil(t, bnr)
	require.Equal(t, ErrUserDoesntExist, err)
}

func TestGetPreviewNotValidBanner(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockRenderer := mocks.NewMockPreviewRenderer(ctrl)
	mockStatsService := mocks.NewMockStatsService(ctrl)
	usecase := NewPreviewUsecase(mockStatsService, mockRenderer)
	username := "hurtki324"
	bnrType := "frej"
	bnr, err := usecase.GetPreview(context.Background(), username, bnrType)
	require.Nil(t, bnr)
	require.Equal(t, ErrInvalidBannerType, err)
}

func TestGetPreviewNotAvailableStatsService(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockRenderer := mocks.NewMockPreviewRenderer(ctrl)
	mockStatsService := mocks.NewMockStatsService(ctrl)
	usecase := NewPreviewUsecase(mockStatsService, mockRenderer)
	username := "hurtki"
	bnrType := "wide"

	ctx := context.Background()

	mockStatsService.EXPECT().GetStats(ctx, username).Return(domain.GithubUserStats{}, domain.ErrUnavailable)

	bnr, err := usecase.GetPreview(context.Background(), username, bnrType)
	require.Equal(t, ErrCantGetPreview, err)
	require.Nil(t, bnr)
}
