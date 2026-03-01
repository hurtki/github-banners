package layout

import "fmt"

func langColorHash(name string) string {
	if c, ok := knownLanguageColors[name]; ok {
		return c
	}
	h := uint32(5381)
	for _, c := range name {
		h = h*33 + uint32(c)
	}
	return fmt.Sprintf("hsl(%d,60%%,52%%)", int(h%360))
}

var knownLanguageColors = map[string]string{
	"Go":         "#00ADD8",
	"Python":     "#3572A5",
	"JavaScript": "#f1e05a",
	"TypeScript": "#2b7489",
	"Java":       "#b07219",
	"C++":        "#f34b7d",
	"C":          "#555555",
	"C#":         "#178600",
	"Rust":       "#dea584",
	"Shell":      "#89e051",
	"Ruby":       "#701516",
	"PHP":        "#4F5D95",
	"Swift":      "#F05138",
	"Kotlin":     "#A97BFF",
	"Scala":      "#c22d40",
	"Dart":       "#00B4AB",
	"Elixir":     "#6e4a7e",
	"Haskell":    "#5e5086",
	"Lua":        "#000080",
	"R":          "#198CE7",
	"MATLAB":     "#e16737",
	"HTML":       "#e34c26",
	"CSS":        "#563d7c",
	"Vue":        "#41b883",
	"Dockerfile": "#384d54",
	"Makefile":   "#427819",
	"Brainfuck":  "#59491c",
}
