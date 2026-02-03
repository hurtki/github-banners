package domain

type BannerType int

const (
	TypeWide BannerType = iota
)

var BannerTypes = map[string]BannerType{
	"wide" : TypeWide,
}

var BannerTypesBackward = map[BannerType]string {
	TypeWide: "wide",
}

type GithubUserBannerInfo struct {
	Username string
	BannerType BannerType
	Stats	GithubUserStats
	StoragePath string
}

type GithubBanner struct {
	Username	string 
	BannerType	BannerType
	Banner		[]byte
}

type GithubUserBannerDescriptor struct {
	Username	string
	URLPath		string
	BannerType BannerType
	Stats		GithubUserStats
}
