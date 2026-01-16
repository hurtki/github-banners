package domain

import "github.com/google/go-github/v81/github"

// Calculating stats aggregates repository statistics without api calls
func CalculateStats(repos []*github.Repository) Stats {
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
