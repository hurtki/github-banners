package cache

import (
	"encoding/binary"
	"fmt"
	"sort"
	"time"

	"github.com/cespare/xxhash/v2"
	"github.com/hurtki/github-banners/api/internal/domain"
	"github.com/patrickmn/go-cache"
)

type PreviewMemoryCache struct {
	cache *cache.Cache
	ttl   time.Duration
}

func NewPreviewMemoryCache(ttl time.Duration) *PreviewMemoryCache {
	return &PreviewMemoryCache{
		cache: cache.New(ttl, time.Minute*10),
		ttl:   ttl,
	}
}

func (c *PreviewMemoryCache) Get(bf domain.BannerInfo) (*domain.Banner, bool) {
	if item, found := c.cache.Get(getHashKey(bf)); found {
		if banner, ok := item.(*domain.Banner); ok {
			return banner, true
		}
	}
	return nil, false
}

func (c *PreviewMemoryCache) Set(bf domain.BannerInfo, banner *domain.Banner) {
	c.cache.Set(getHashKey(bf), banner, c.ttl)
}

func getHashKey(b domain.BannerInfo) string {
	h := xxhash.New()

	// Username
	h.WriteString(b.Username)

	// BannerType
	writeInt(h, int(b.BannerType))

	// Stats
	writeInt(h, b.Stats.TotalRepos)
	writeInt(h, b.Stats.OriginalRepos)
	writeInt(h, b.Stats.ForkedRepos)
	writeInt(h, b.Stats.TotalStars)
	writeInt(h, b.Stats.TotalForks)

	// Languages (sorted for determinism)
	if len(b.Stats.Languages) > 0 {
		keys := make([]string, 0, len(b.Stats.Languages))
		for k := range b.Stats.Languages {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			h.WriteString(k)
			writeInt(h, b.Stats.Languages[k])
		}
	}

	// FetchedAt (use unix nano for determinism)
	writeInt64(h, b.Stats.FetchedAt.UnixNano())

	return fmt.Sprintf("%x", h.Sum64())
}

func writeInt(h *xxhash.Digest, v int) {
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], uint64(v))
	h.Write(buf[:])
}

func writeInt64(h *xxhash.Digest, v int64) {
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], uint64(v))
	h.Write(buf[:])
}
