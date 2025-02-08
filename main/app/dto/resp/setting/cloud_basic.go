package setting

// CloudBasic 云端基础信息
type CloudBasic struct {
	BrandLogo          string `json:"brand_logo"`
	BrandLogoLong      string `json:"brand_logo_long"`
	BrandName          string `json:"brand_name"`
	BrowserLogo        string `json:"browser_logo"`
	BrowserTitle       string `json:"browser_title"`
	ExpirationReminder uint   `json:"expiration_reminder"`
}
