package layout

import "github.com/hurtki/github-banners/renderer/internal/domain"

type Theme struct {
	Background string
	Foreground string
	Muted      string
}

type LanguageSegment struct {
	X     int
	Width int
	Color string
}

type LegendItem struct {
	DotX  int
	DotY  int
	TextX int
	TextY int
	Color string
	Label string
}

type BannerView struct {
	Width         int
	Height        int
	Username      string
	BannerType    string
	Stats         domain.GithubUserStats
	Theme         Theme
	BarWidth      int
	Languages     []LanguageSegment
	Legend        []LegendItem
	FormattedTime string
}
