package domain

type BannerType int

const (
	TypeWide BannerType = iota
)

var BannerTypes = map[string]BannerType{
	"wide": TypeWide,
}

var BannerTypesBackward = map[BannerType]string{
	TypeWide: "wide",
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

type BannerMetadata struct {
	Username string 
	BannerType BannerType
	UrlPath string
}

// Rendered banner
type Banner struct {
	Username   string
	BannerType BannerType
	Banner     []byte
}
