package setting

type SysAdminConfig struct {
	BrandName          string `json:"brand_name"`
	BrandLogo          string `json:"brand_logo"`
	BrandLogoLong      string `json:"brand_logo_long"`
	BrowserLogo        string `json:"browser_logo"`
	BrowserTitle       string `json:"browser_title"`
	ExpirationReminder int    `json:"expiration_reminder"`
}
