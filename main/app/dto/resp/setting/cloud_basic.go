package setting

// CloudBasic 云端基础信息
type CloudBasic struct {
	BrandLogo          string `json:"brand_logo"`          // 品牌LOGO： (正方型)
	BrandLogoLong      string `json:"brand_logo_long"`     // 品牌LOGO： (长方型)
	BrandName          string `json:"brand_name"`          // 品牌名称
	BrowserLogo        string `json:"browser_logo"`        // 浏览器title： （LOGO）
	BrowserTitle       string `json:"browser_title"`       // 浏览器title： (品牌名称)
	ExpirationReminder uint   `json:"expiration_reminder"` // 准备到期提醒，根据填写的时间提醒商家剩余可用时间
}
