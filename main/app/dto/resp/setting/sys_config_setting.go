package setting

// SysConfig 系统设置
type SysConfig struct {
	ShopName     string `json:"shop_name"`      // 商城名称
	ShopBgImg    string `json:"shop_bg_img"`    // 商城背景图
	ShopLogoImg  string `json:"shop_logo_img"`  // 商城logo
	CashierName  string `json:"cashier_name"`   // 收银台名称
	CashierBgImg string `json:"cashier_bg_img"` // 收银台背景图
}
