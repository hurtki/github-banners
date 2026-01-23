package github

import (
	"context"
	"time"

	"github.com/google/go-github/v81/github"
	"github.com/hurtki/github-banners/api/internal/domain"
	"golang.org/x/oauth2"
)

type Fetcher struct {
	client *github.Client
	config *domain.ServiceConfig
}

func NewFetcher(token string, config *domain.ServiceConfig) *Fetcher {
	var client *github.Client

	if token != "" {
		tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
		tokenClient := oauth2.NewClient(context.Background(), tokenSource)
		client = github.NewClient(tokenClient)
	} else {
		client = github.NewClient(nil)
	}

	return &Fetcher{
		client: client,
		config: config,
	}
}

// FetchUser fetches the user data from GitHub
func (f *Fetcher) fetchUser(ctx context.Context, username string) (*github.User, error) {
	ctx, cancel := context.WithTimeout(ctx, f.config.RequestTimeout)
	defer cancel()

	user, _, err := f.client.Users.Get(ctx, username)
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
		repos, resp, err := f.client.Repositories.ListByUser(ctx, username, opts)
		if err != nil {
			return nil, err
		}
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
