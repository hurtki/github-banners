package github

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/google/go-github/v81/github"
	"github.com/hurtki/github-banners/api/internal/domain"
	"github.com/hurtki/github-banners/api/internal/logger"
	"golang.org/x/oauth2"
)

type Fetcher struct {
	clients []*GithubClient
	config  *domain.ServiceConfig
	logger  logger.Logger
}

type GithubClient struct {
	Client    *github.Client
	Remaining int
	ResetsAt  time.Time
}

func NewFetcher(tokens []string, config *domain.ServiceConfig, logger logger.Logger) *Fetcher {
	clients := []*GithubClient{}
	initLogger := logger.With("service", "fetcher initialization function")
	for _, token := range tokens {
		var client *github.Client

		if token == "" {
			continue
		}
		tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
		tokenClient := oauth2.NewClient(context.Background(), tokenSource)
		client = github.NewClient(tokenClient)

		clLimit, _, err := client.RateLimit.Get(context.TODO())
		if err != nil {
			initLogger.Error("unexpected error, when getting client's rate limit, skipping it", "err", err)
			continue
		}

		clients = append(clients, &GithubClient{
			Client:    client,
			Remaining: clLimit.Core.Remaining,
			ResetsAt:  clLimit.Core.Reset.Time,
		})
	}
	if len(clients) == 0 {
		initLogger.Warn("initialized Fetcher without clients")
	} else {
		initLogger.Info("initialized Fetcher", "clients count", len(clients))
	}

	return &Fetcher{
		clients: clients,
		config:  config,
		logger:  logger.With("service", "github-fetcher"),
	}
}

// findBestClient finds client with at least one request available
// if all clients are in limit, returns nil
func (f *Fetcher) findClient(ctx context.Context) *GithubClient {
	fn := "internal.infrastructure.github.Fetcher.findClient"
	for _, cl := range f.clients {
		if cl.Remaining > 0 {
			return cl
		}
		if cl.ResetsAt.Before(time.Now()) {
			ctx, cancel := context.WithTimeout(ctx, f.config.RequestTimeout)
			rl, _, err := cl.Client.RateLimit.Get(ctx)
			cancel()
			if err != nil {
				f.logger.Error("found client, that its Reset time is before Now(), error occured when getting its rate limit, skipping", "err", err, "source", fn)
				continue
			}
			cl.ResetsAt = rl.Core.Reset.Time
			cl.Remaining = rl.Core.Remaining
			if cl.Remaining > 0 {
				return cl
			}
		}
	}
	return nil
}

// updateClientWithResponse tries to get rate limit headers from response
// and updates client's fields using them
func (f *Fetcher) updateClientWithResponse(cl *GithubClient, res *http.Response) {
	if res == nil {
		return
	}
	resetUnix, err := strconv.ParseInt(
		res.Header.Get("X-RateLimit-Reset"), 10, 64,
	)
	if err != nil {
		f.logger.Error("can't parse X-RateLimit-Reset github api response header into int64", "err", err)
	} else {
		cl.ResetsAt = time.Unix(resetUnix, 0)
	}

	remaining, err := strconv.ParseInt(
		res.Header.Get("X-RateLimit-Remaining"), 10, 64,
	)
	if err != nil {
		f.logger.Error("can't parse X-RateLimit-Remaining github api response header into int64", "err", err)
	} else {
		cl.Remaining = int(remaining)
	}

}

// FetchUser fetches the user data from GitHub
func (f *Fetcher) fetchUser(ctx context.Context, username string) (*github.User, error) {
	cl := f.findClient(ctx)
	if cl == nil {
		f.logger.Warn("can't find available client for github api request")
		return nil, domain.ErrUnavailable
	}

	ctx, cancel := context.WithTimeout(ctx, f.config.RequestTimeout)
	defer cancel()

	user, res, err := cl.Client.Users.Get(ctx, username)
	f.updateClientWithResponse(cl, res.Response)
	if err != nil {
		if er, ok := err.(*github.ErrorResponse); ok {
			if er.Response.StatusCode == http.StatusNotFound {
				return nil, domain.ErrNotFound
			}
		}
		return nil, domain.ErrUnavailable
	}

	return user, err
}

// FetchRepositories fetches all repositories for a user (paginated)
func (f *Fetcher) fetchRepositories(ctx context.Context, username string) ([]*github.Repository, error) {
	cl := f.findClient(ctx)
	if cl == nil {
		f.logger.Warn("can't find available client for github api request")
		return nil, domain.ErrUnavailable
	}

	var allRepos []*github.Repository
	opts := &github.RepositoryListByUserOptions{
		Type:        "owner",
		Sort:        "updated",
		ListOptions: github.ListOptions{PerPage: 100},
	}

	for {
		repos, resp, err := cl.Client.Repositories.ListByUser(ctx, username, opts)
		if err != nil {
			if er, ok := err.(*github.ErrorResponse); ok {
				if er.Response.StatusCode == http.StatusNotFound {
					return nil, domain.ErrNotFound
				}
			}
			return nil, domain.ErrUnavailable
		}
		f.updateClientWithResponse(cl, resp.Response)

		allRepos = append(allRepos, repos...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return allRepos, nil
}

// FetchUserStats fetches user and repositories together
func (f *Fetcher) FetchUserData(ctx context.Context, username string) (*domain.GithubUserData, error) {
	user, err := f.fetchUser(ctx, username)
	if err != nil {
		return nil, err
	}

	repos, err := f.fetchRepositories(ctx, username)
	if err != nil {
		return nil, err
	}

	domainRepos := make([]domain.GithubRepository, len(repos))
	for i := range len(repos) {
		if repos[i] == nil {
			continue
		}
		var pushedAt *time.Time = nil
		var updatedAt *time.Time = nil
		if repos[i].PushedAt != nil {
			pushedAt = repos[i].PushedAt.GetTime()
		}
		if repos[i].UpdatedAt != nil {
			updatedAt = repos[i].UpdatedAt.GetTime()
		}

		domainRepos[i] = domain.GithubRepository{
			ID:            repos[i].GetID(),
			OwnerUsername: repos[i].GetOwner().GetLogin(),
			PushedAt:      pushedAt,
			UpdatedAt:     updatedAt,
			Language:      repos[i].Language,
			StarsCount:    repos[i].GetStargazersCount(),
			Fork:          repos[i].GetFork(),
			ForksCount:    repos[i].GetForksCount(),
		}
	}

	return &domain.GithubUserData{
		Username:     user.GetLogin(),
		Name:         user.Name,
		Company:      user.Company,
		Location:     user.Location,
		PublicRepos:  user.GetPublicRepos(),
		Followers:    user.GetFollowers(),
		Following:    user.GetFollowing(),
		Repositories: domainRepos,
		// sets the FetchedAt field to time when it was fetched
		FetchedAt: time.Now(),
	}, nil
}
