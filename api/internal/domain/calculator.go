package domain

// Calculating stats aggregates repository statistics without api calls
func CalculateStats(repos []GithubRepository) GithubUserStats {
	var stats GithubUserStats
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
