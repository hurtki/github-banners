package bannersstorage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hurtki/github-banners/storage/internal/domain"
	"github.com/hurtki/github-banners/storage/internal/logger"
)

type WriteFileFunc func(string, []byte, os.FileMode) error

type FileStorage struct {
	basePath      string
	logger        logger.Logger
	writeFileFunc WriteFileFunc

	kmu *keyedMutex
}

func NewFileStorage(basePath string, logger logger.Logger, wWriteFileFunc WriteFileFunc) *FileStorage {
	return &FileStorage{
		basePath:      basePath,
		logger:        logger.With("service", "banners-file-storage"),
		writeFileFunc: wWriteFileFunc,
		kmu:           newKeyedMutex(),
	}
}

// Save is idempotent, saves banner's content with given extenstion
// returns nil or domain.ErrUnavailable
func (s *FileStorage) Save(ctx context.Context, name string, extension domain.BannerExtension, content []byte) error {
	switch extension {
	case domain.SvgBannerExtension:
		path := filepath.Join(s.basePath, name+".svg")
		muKey := fmt.Sprintf("%s%d", name, uint8(extension))

		// usage of keyed mutex
		// in order to block on same file write operations
		// and don't block on different file write operations
		s.kmu.Lock(muKey)
		defer s.kmu.Unlock(muKey)

		err := s.writeFileFunc(path, content, 0644)

		if err != nil {
			s.logger.Error("error from os", "err", err, "path", path)
			return domain.ErrUnavailable
		}
		return nil
	default:
		s.logger.Warn("unexpected extension", "extension", extension)
		return domain.ErrUnavailable
	}
}
