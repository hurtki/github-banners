package cache

import (
	"encoding/binary"
	"fmt"
	"slices"
	"sync"

	"github.com/cespare/xxhash/v2"
	"github.com/hurtki/github-banners/api/internal/domain"
)

type bannerInfoHashCounter struct {
	xxhashPool         sync.Pool
	languagesSlicePool sync.Pool
}

func newBannerInfoHashCounter() bannerInfoHashCounter {
	return bannerInfoHashCounter{
		xxhashPool: sync.Pool{New: func() any {
			return xxhash.New()
		}},
		languagesSlicePool: sync.Pool{
			New: func() any {
				return make([]string, 0)
			},
		},
	}
}

// getHashKey counts hash for domain.bannerInfo
// uses xxhashPool to get xxhash Digest
// uses languagesSlicePool to get blank slice for languages bannerInfo field sorting
// doesn't count FetchedAt field
func (c *bannerInfoHashCounter) Hash(b domain.BannerInfo) string {
	h := c.xxhashPool.Get().(*xxhash.Digest)

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
		// taking blank slice from the pool ( should be len=0 )
		keys := c.languagesSlicePool.Get().([]string)

		for k := range b.Stats.Languages {
			keys = append(keys, k)
		}
		slices.Sort(keys)

		for _, k := range keys {
			h.WriteString(k)
			writeInt(h, b.Stats.Languages[k])
		}

		// resetting len on the slice, so other gorutine could reuse it
		keys = keys[:0]
		c.languagesSlicePool.Put(keys)
	}

	res := fmt.Sprintf("%x", h.Sum64())

	// resetting the xxhash, so other goruite could reuse it
	h.Reset()
	c.xxhashPool.Put(h)

	return res
}

func writeInt(h *xxhash.Digest, v int) {
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], uint64(v))
	h.Write(buf[:])
}
