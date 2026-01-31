package github

import (
	"context"
	"net/http"
	"sync"
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

// GithubClient is a wrap of GithubClient that contains rate limit info about client
type GithubClient struct {
	Client *github.Client
	// Requests, that client has remaining, until limit
	Remaining int
	// time, when limit will reset
	ResetsAt time.Time

	// mutex for concurrent changes of Remaining and ResetsAt fields
	mu sync.Mutex
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

		timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		clLimit, _, err := client.RateLimit.Get(timeoutCtx)
		cancel()
		if err != nil {
			initLogger.Error("unexpected error, when getting client's rate limit, skipping it", "err", err)
			continue
		}

		clients = append(clients, &GithubClient{
			Client:    client,
			Remaining: clLimit.Core.Remaining,
			ResetsAt:  clLimit.Core.Reset.Time,
			mu:        sync.Mutex{},
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

// FetchUser fetches the user data from GitHub
func (f *Fetcher) fetchUser(ctx context.Context, username string) (*github.User, error) {
	cl := f.acquireClient(ctx)
	if cl == nil {
		f.logger.Warn("can't find available client for github api request")
		return nil, domain.ErrUnavailable
	}

	ctx, cancel := context.WithTimeout(ctx, f.config.RequestTimeout)
	defer cancel()

	user, res, err := cl.Client.Users.Get(ctx, username)
	f.updateClientWithDoneResponse(cl, res.Response)
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

	var allRepos []*github.Repository
	opts := &github.RepositoryListByUserOptions{
		Type:        "owner",
		Sort:        "updated",
		ListOptions: github.ListOptions{PerPage: 100},
	}

	for {
		// every page aquire a new client for one request
		cl := f.acquireClient(ctx)
		if cl == nil {
			f.logger.Warn("can't find available client for github api request")
			if allRepos == nil {
				return nil, domain.ErrUnavailable
			} else {
				// even if we already collected couple repositories it shouldn't be returned
				// because Fetcher is used as source of truth
				return nil, domain.ErrUnavailable
			}
		}
		timeoutCtx, cancel := context.WithTimeout(ctx, f.config.RequestTimeout)
		repos, resp, err := cl.Client.Repositories.ListByUser(timeoutCtx, username, opts)
		cancel()

		if err != nil {
			if er, ok := err.(*github.ErrorResponse); ok {
				if er.Response.StatusCode == http.StatusNotFound {
					return nil, domain.ErrNotFound
				}
			}
			return nil, domain.ErrUnavailable
		}
		f.updateClientWithDoneResponse(cl, resp.Response)

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

		var owner *github.User = repos[i].GetOwner()
		if owner == nil || owner.GetLogin() == "" {
			continue
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
