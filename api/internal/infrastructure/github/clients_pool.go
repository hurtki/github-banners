package github

import (
	"context"
	"net/http"
	"strconv"
	"time"
)

// acquireClient finds client with at least one request available.
// Acquires, that one request will be sent using it.
// If all clients are out of requests, returns nil.
func (f *Fetcher) acquireClient(ctx context.Context) *GithubClient {
	fn := "internal.infrastructure.github.Fetcher.findClient"
	for _, cl := range f.clients {
		cl.mu.Lock()

		if cl.Remaining > 0 {
			cl.Remaining-- // acquiring one request
			cl.mu.Unlock()

			return cl
		}
		// if the ResetTime we store is already in past
		// then updating Remaining field
		if cl.ResetsAt.Before(time.Now().UTC()) { // github returns ResetsAt header at UTC
			// here unlock to do net call
			cl.mu.Unlock()

			ctx, cancel := context.WithTimeout(ctx, f.config.RequestTimeout)
			rl, _, err := cl.Client.RateLimit.Get(ctx)
			cancel()
			if err != nil {
				f.logger.Error("found client, that its Reset time is before Now(), error occured when getting its rate limit, skipping", "err", err, "source", fn)
				continue
			}
			// after net call, we are having new source of truth
			// locking mutex for changes
			cl.mu.Lock()

			cl.ResetsAt = rl.Core.Reset.Time
			cl.Remaining = rl.Core.Remaining
			if cl.Remaining > 0 {
				cl.Remaining-- // acquires one request for client
				cl.mu.Unlock()
				return cl
			}
			// if new Remaining is still 0, then continuing with next client
			cl.mu.Unlock()
			continue
		}
		// if ResetTime is in future, unlock mutex to continue with next client
		cl.mu.Unlock()
	}
	return nil
}

// UpdateClientWithResponse tries to get rate limit headers from response.
// Updates client's fields using this reponse's headers.
func (f *Fetcher) updateClientWithDoneResponse(cl *GithubClient, res *http.Response) {
	if res == nil {
		return
	}
	cl.mu.Lock()
	defer cl.mu.Unlock()

	resetUnix, err := strconv.ParseInt(
		res.Header.Get("X-RateLimit-Reset"), 10, 64,
	)
	if err != nil {
		f.logger.Warn("can't parse X-RateLimit-Reset github api response header into int64", "err", err)
	} else {
		cl.ResetsAt = time.Unix(resetUnix, 0)
	}

	remaining, err := strconv.ParseInt(
		res.Header.Get("X-RateLimit-Remaining"), 10, 64,
	)
	if err != nil {
		f.logger.Warn("can't parse X-RateLimit-Remaining github api response header into int64", "err", err)
	} else {
		cl.Remaining = int(remaining)
	}

}
