package github

import (
	"context"
	"time"

	"github.com/google/go-github/v81/github"
	"golang.org/x/oauth2"
)

type Service struct {
	client *github.Client
	cache Cache
	config *ServiceConfig
}

func NewService(token string, cache Cache, config *ServiceConfig) *Service {
	var client *github.Client

	if token != "" {
		tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
		tokenClient := oauth2.NewClient(context.Background(), tokenSource)
		client = github.NewClient(tokenClient)
	} else {
		client = github.NewClient(nil)
	}

	return &Service {
		client: client,
		cache: cache,
		config: config,
	}
}


func (s *Service) GetUserStats(ctx context.Context, username string) (*UserStats, error) {
	if cachedStats, found := s.cache.Get(username); found {
		return cachedStats, nil
	}
	userStats, err := s.fetchFromGithub(ctx, username)
	if err != nil {
		return nil, err
	}
	go s.cache.Set(username, userStats)
	return userStats, nil
}

func (s *Service) fetchFromGithub(ctx context.Context, username string) (*UserStats, error) {
	ctx, cancel := context.WithTimeout(ctx, s.config.RequestTimeout)
	defer cancel()

	user, _, err := s.client.Users.Get(ctx, username)
	if err != nil {
		return nil, err
	}

	repos, err := s.getAllRepositories(ctx, username)
	if err != nil {
		return nil, err
	}

	stats := s.calculateStats(repos)

	return &UserStats{
		User: user,
		Repositories: repos,
		Stats: stats,
		FetchedAt: time.Now(),
		Cached: false,
	}, nil
}

func (s *Service) getAllRepositories(ctx context.Context, username string) ([]*github.Repository, error) {
	var allRepos []*github.Repository
	opts := &github.RepositoryListByUserOptions{
		Type: "owner",
		Sort: "updated",
		ListOptions: github.ListOptions{PerPage: 100},
	}
	for {
		repos, resp, err := s.client.Repositories.ListByUser(ctx, username, opts)
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

func (s *Service) calculateStats(repos []*github.Repository) Stats {
	var stats Stats
	stats.Languages = make(map[string]int)

	for _, repo := range repos {
		if repo.GetFork() {
			stats.ForkedRepos++
		} else {
			stats.OriginalRepos++
			stats.TotalStars += repo.GetStargazersCount()
			stats.TotalForks += repo.GetForksCount()

			if lang := repo.GetLanguage(); lang != "" {
				stats.Languages[lang] = stats.Languages[lang] + 1
			}
		}
	}
	stats.TotalRepos = len(repos)
	return stats
}

//getmultipleusers fetches stats for multiple users concurrently
func (s *Service) GetMultipleUsers(ctx context.Context, usernames []string) (map[string]*UserStats, error) {
	result := make(map[string]*UserStats)
	errChan := make(chan error, len(usernames))
	statsChan := make(chan struct {
		username string
		stats *UserStats
	}, len(usernames))

	for _, username := range usernames {
		go func(user string) {
			stats, err := s.GetUserStats(ctx, user)
			if err != nil {
				errChan <- err
				return
			}
			statsChan <- struct {
				username string 
				stats *UserStats
			}{username: user, stats: stats}
		}(username)
	}

	for i:= 0; i< len(usernames); i++ {
		select {
		case err := <-errChan:
			_ = err
			//log the error but continue with other users
			continue
		case resultData := <-statsChan:
			result[resultData.username] = resultData.stats
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	return result, nil
} 
