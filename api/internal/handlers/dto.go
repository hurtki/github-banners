package handlers

type CreateBannerRequest struct {
	Username   string `json:"username"`
	BannerType string `json:"type"`
}

type CreateBannerResponse struct {
	BannerUrlPath string `json:"url"`
}
