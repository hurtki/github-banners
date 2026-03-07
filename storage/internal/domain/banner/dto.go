package banner

type SaveIn struct {
	UrlPath    string
	BannerData []byte
	Format     string
}

type SaveOut struct {
	BannerUrl string
}
