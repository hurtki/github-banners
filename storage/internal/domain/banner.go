package domain

type BannerExtension uint8

const (
	SvgBannerExtension = iota
)

var BannerExtensions = map[string]BannerExtension{
	"svg": SvgBannerExtension,
}
