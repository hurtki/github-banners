package domain

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func calendarFromCounts(counts ...int) ContributionCalendar {
	weeks := make([]ContributionWeek, 0)
	var current ContributionWeek
	for i, c := range counts {
		current.Days = append(current.Days, ContributionDay{Count: c})
		if len(current.Days) == 7 || i == len(counts)-1 {
			weeks = append(weeks, current)
			current = ContributionWeek{}
		}
	}
	return ContributionCalendar{Weeks: weeks}
}

func TestCurrentStreakEmpty(t *testing.T) {
	require.Equal(t, 0, ContributionCalendar{}.CurrentStreak())
	require.Equal(t, 0, calendarFromCounts().CurrentStreak())
}

func TestCurrentStreakAllZero(t *testing.T) {
	cal := calendarFromCounts(0, 0, 0, 0, 0, 0, 0)
	require.Equal(t, 0, cal.CurrentStreak())
}

func TestCurrentStreakAllActive(t *testing.T) {
	cal := calendarFromCounts(1, 2, 3, 4, 5, 6, 7)
	require.Equal(t, 7, cal.CurrentStreak())
}

func TestCurrentStreakInterruptedByZero(t *testing.T) {
	cal := calendarFromCounts(5, 5, 5, 0, 2, 3, 4)
	require.Equal(t, 3, cal.CurrentStreak())
}

func TestCurrentStreakTodayZeroIgnored(t *testing.T) {
	cal := calendarFromCounts(1, 2, 3, 0)
	require.Equal(t, 3, cal.CurrentStreak())
}

func TestCurrentStreakTodayZeroAndYesterdayZero(t *testing.T) {
	cal := calendarFromCounts(5, 5, 5, 0, 0)
	require.Equal(t, 0, cal.CurrentStreak())
}

func TestCurrentStreakSpansMultipleWeeks(t *testing.T) {
	cal := calendarFromCounts(
		0, 0, 0, 0, 0, 0, 0,
		2, 2, 2, 2, 2, 2, 2,
		3, 3, 3, 3, 3, 3, 3,
	)
	require.Equal(t, 14, cal.CurrentStreak())
}

func TestCurrentStreakSingleDayActive(t *testing.T) {
	cal := calendarFromCounts(7)
	require.Equal(t, 1, cal.CurrentStreak())
}
