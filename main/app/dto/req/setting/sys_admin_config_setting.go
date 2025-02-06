package setting

// SysAdminConfig 系统设置
type SysAdminConfig struct {
	BrandName          string `json:"brand_name"`          // 商城名称
	BrandLogo          string `json:"brand_logo"`          // 商城背景图
	BrandLogoLong      string `json:"brand_logo_long"`     // 商城logo
	BrowserLogo        string `json:"browser_logo"`        // 浏览器LOGO
	BrowserTitle       string `json:"browser_title"`       // 浏览器标题
	ExpirationReminder int    `json:"expiration_reminder"` // 剩余多少天到期提示
}
