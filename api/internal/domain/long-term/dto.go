package longterm

type CreateBannerIn struct {
	Username   string
	BannerType string
}

type CreateBannerOut struct {
	BannerUrlPath string
}
