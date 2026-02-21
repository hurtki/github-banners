package longterm

import (
	"fmt"

	"github.com/hurtki/github-banners/api/internal/domain"
)

func GenerateUrlPath(username string, bt domain.BannerType) string {
	return fmt.Sprintf("%s-%s", username, domain.BannerTypesBackward[bt])
}
