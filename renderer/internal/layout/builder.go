package layout

import (
	"fmt"
	"math"
	"sort"

	"github.com/hurtki/github-banners/renderer/internal/domain"
)

func BuildView(info domain.BannerInfo) *BannerView {
	const (
		W        = 460
		H        = 210
		pad      = 20
		barWidth = W - pad*2
		maxLangs = 5
	)

	var theme Theme

	if string(info.BannerType) == "dark" || string(info.BannerType) == string(domain.BannerTypeDark) {
		theme = Theme{
			Background:                 "#0d1117",
			Foreground:                 "#e6edf3",
			Muted:                      "#8b949e",
			BackgroundColorGradientOne: "#161b22",
		}
	} else {
		theme = Theme{
			Background:                 "#f6f8fa",
			Foreground:                 "#24292f",
			Muted:                      "#57606a",
			BackgroundColorGradientOne: "#ffffff",
		}
	}

	total := 0
	for _, v := range info.Stats.Languages {
		total += v
	}

	type kv struct {
		Name  string
		Value int
	}

	var sorted []kv
	for k, v := range info.Stats.Languages {
		sorted = append(sorted, kv{k, v})
	}

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Value > sorted[j].Value
	})

	if len(sorted) > maxLangs {
		other := 0
		for _, l := range sorted[maxLangs:] {
			other += l.Value
		}
		sorted = append(sorted[:maxLangs], kv{
			Name:  "Other",
			Value: other,
		})
	}

	var segments []LanguageSegment
	var legend []LegendItem

	cursor := pad
	legendY := 170
	colW := 145

	for i, l := range sorted {
		pct := 0.0
		if total > 0 {
			pct = math.Round(float64(l.Value)/float64(total)*1000) / 10
		}

		w := max(int(pct/100*float64(barWidth)), 1)
		color := langColorHash(l.Name)

		segments = append(segments, LanguageSegment{
			X:     cursor,
			Width: w,
			Color: color,
		})

		col := i % 3
		row := i / 3

		legend = append(legend, LegendItem{
			DotX:  pad + col*colW + 4,
			DotY:  legendY + row*15,
			TextX: pad + col*colW + 14,
			TextY: legendY + row*15 + 4,
			Color: color,
			Label: fmt.Sprintf("%s %.1f%%", l.Name, pct),
		})

		cursor += w
	}
	return &BannerView{
		Width:         W,
		Height:        H,
		Username:      info.Username,
		BannerType:    string(info.BannerType),
		Stats:         info.Stats,
		Theme:         theme,
		BarWidth:      barWidth,
		Languages:     segments,
		Legend:        legend,
		FormattedTime: info.Stats.FetchedAt.Format("02 Jan 2006 · 15:04"),
	}
}
