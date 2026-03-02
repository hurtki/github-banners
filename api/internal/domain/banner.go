package domain

type BannerType int

const (
	TypeDefault BannerType = iota
	TypeDark    BannerType = iota
)

var BannerTypes = map[string]BannerType{
	"default": TypeDefault,
	"dark":    TypeDark,
}

var BannerTypesBackward = map[BannerType]string{
	TypeDefault: "default",
	TypeDark:    "dark",
}

// BannerInfo is all data that banner contains
// used to render banner
type BannerInfo struct {
	Username   string
	BannerType BannerType
	Stats      GithubUserStats
}

// Long term banner info, embedded GithubBannerInfo with UrlPath
// used for updates of long term banners
type LTBannerInfo struct {
	BannerInfo
	UrlPath string
}

type LTBannerMetadata struct {
	Username   string
	BannerType BannerType
	UrlPath    string
	Active     bool
}

// Rendered banner
type Banner struct {
	Username   string
	BannerType BannerType
	Banner     []byte
}
