package preview

import (
	"context"
	"testing"

	"github.com/hurtki/github-banners/api/internal/domain"
	"github.com/hurtki/github-banners/api/internal/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestServiceGetPreviewSuccessFromCache(t *testing.T) {
	ctrl := gomock.NewController(t)

	cache := mocks.NewMockCache(ctrl)
	renderer := mocks.NewMockPreviewRenderer(ctrl)

	service := NewPreviewService(renderer, cache)

	bnrInfo := domain.BannerInfo{Username: "hurtki", BannerType: domain.TypeWide}
	expectedBanner := domain.Banner{Username: bnrInfo.Username, BannerType: bnrInfo.BannerType, Banner: []byte("some")}

	cache.EXPECT().Get(bnrInfo).Return(&expectedBanner, true)
	bnr, err := service.GetPreview(context.Background(), bnrInfo)
	require.Nil(t, err)
	require.Equal(t, &expectedBanner, bnr)
}

func TestServiceGetPreviewSuccessNoCache(t *testing.T) {
	ctrl := gomock.NewController(t)

	cache := mocks.NewMockCache(ctrl)
	renderer := mocks.NewMockPreviewRenderer(ctrl)

	service := NewPreviewService(renderer, cache)

	bnrInfo := domain.BannerInfo{Username: "hurtki", BannerType: domain.TypeWide}
	expectedBanner := domain.Banner{Username: bnrInfo.Username, BannerType: bnrInfo.BannerType, Banner: []byte("some")}
	ctx := t.Context()

	cache.EXPECT().Get(bnrInfo).Return(nil, false)
	renderer.EXPECT().RenderPreview(ctx, bnrInfo).Return(&expectedBanner, nil)
	cache.EXPECT().Set(bnrInfo, &expectedBanner)

	bnr, err := service.GetPreview(ctx, bnrInfo)
	require.Nil(t, err)
	require.Equal(t, &expectedBanner, bnr)
}

func TestServiceGetPreviewErrorsHandling(t *testing.T) {
	ctrl := gomock.NewController(t)

	cache := mocks.NewMockCache(ctrl)
	renderer := mocks.NewMockPreviewRenderer(ctrl)

	service := NewPreviewService(renderer, cache)

	bnrInfo := domain.BannerInfo{Username: "hurtki", BannerType: domain.TypeWide}
	ctx := t.Context()

	errorsFromRenderer := []error{
		domain.ErrNotFound,
		domain.ErrUnavailable,
		&domain.ConflictError{},
	}

	for _, er := range errorsFromRenderer {
		cache.EXPECT().Get(bnrInfo).Return(nil, false)
		renderer.EXPECT().RenderPreview(ctx, bnrInfo).Return(nil, er)

		bnr, err := service.GetPreview(ctx, bnrInfo)
		require.Nil(t, bnr)
		require.Equal(t, er, err)
	}
}
