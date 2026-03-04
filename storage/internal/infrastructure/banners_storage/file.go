package bannersstorage

import (
	"os"
	"path"

	"github.com/hurtki/github-banners/storage/internal/domain"
	"github.com/hurtki/github-banners/storage/internal/logger"
)

type WriteFileFunc func(string, []byte, os.FileMode) error

type FileStorage struct {
	basePath      string
	logger        logger.Logger
	writeFileFunc WriteFileFunc
}

func NewFileStorage(basePath string, logger logger.Logger, wWriteFileFunc WriteFileFunc) *FileStorage {
	return &FileStorage{
		basePath:      basePath,
		logger:        logger.With("service", "banners-file-storage"),
		writeFileFunc: wWriteFileFunc,
	}
}

// Save is idempotent, saves banner's content with given extenstion
// returns nil or domain.ErrUnavailable
func (s *FileStorage) Save(name string, extension domain.BannerExtension, content []byte) error {
	switch extension {
	case domain.SvgBannerExtension:
		path := path.Join(s.basePath, name+".svg")
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
