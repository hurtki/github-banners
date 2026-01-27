package userstats

import "github.com/hurtki/github-banners/api/internal/domain"

// CalculateStats aggregates repository statistics without additional API calls.
func CalculateStats(repos []domain.GithubRepository) domain.GithubUserStats {
	var stats domain.GithubUserStats
	stats.Languages = make(map[string]int)

	for _, repo := range repos {
		if repo.Fork {
			stats.ForkedRepos++
		} else {
			stats.OriginalRepos++
			stats.TotalStars += repo.StarsCount
			stats.TotalForks += repo.ForksCount

			if lang := repo.Language; lang != nil {
				stats.Languages[*lang] += 1
			}
		}
	}
	stats.TotalRepos = len(repos)
	return stats
}
