package domain

import "time"

// UserStats is the aggregated GitHub statistics for a user, populated from a
// single GraphQL query. Replaces the legacy GithubUserData + GithubUserStats
// pair as part of GIT-33. The legacy types remain in this package while the
// migration is in progress and will be removed once the new banner template
// fully consumes UserStats.
type UserStats struct {
	Username  string
	FetchedAt time.Time

	TotalRepos int
	TotalStars int
	Languages  map[string]int

	OpenPRs   int
	ClosedPRs int
	MergedPRs int

	OpenIssues   int
	ClosedIssues int

	TotalCommits       int
	TotalPRsCreated    int
	TotalIssuesCreated int
	TotalPRReviews     int

	ContributionCalendar ContributionCalendar
	CurrentStreak        int
}

type ContributionCalendar struct {
	TotalContributions int
	Weeks              []ContributionWeek
}

type ContributionWeek struct {
	FirstDay time.Time
	Days     []ContributionDay
}

type ContributionDay struct {
	Date  time.Time
	Count int
}

// CurrentStreak counts consecutive trailing days with at least one contribution.
// The most recent day is allowed to be zero (treated as "in progress today")
// without breaking the streak; any earlier zero day terminates the count.
func (c ContributionCalendar) CurrentStreak() int {
	if len(c.Weeks) == 0 {
		return 0
	}

	days := make([]ContributionDay, 0, len(c.Weeks)*7)
	for _, w := range c.Weeks {
		days = append(days, w.Days...)
	}
	if len(days) == 0 {
		return 0
	}

	i := len(days) - 1
	if days[i].Count == 0 {
		i--
	}

	streak := 0
	for ; i >= 0; i-- {
		if days[i].Count == 0 {
			break
		}
		streak++
	}
	return streak
}
