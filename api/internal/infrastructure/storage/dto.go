package storage

type SaveRequest struct {
	URLPath      string `json:"url_path"`
	BannerData   string `json:"banner_info"`
	BannerFormat string `json:"banner_format"`
}
