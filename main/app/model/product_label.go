package model

// ProductLabel 商品标签表 ttpos_product_label
type ProductLabel struct {
	BaseModel
	Name            string `gorm:"default:'';column:name;comment:'标签名称'"`
	Style           string `gorm:"default:'default';column:style;comment:'标签样式'"`
	IsShowCashier   uint   `gorm:"default:0;column:is_show_cashier;comment:'是否在收银机显示, 0-否 1-是'"`
	IsShowTablet    uint   `gorm:"default:0;column:is_show_tablet;comment:'是否在平板显示, 0-否 1-是'"`
	IsShowAssistant uint   `gorm:"default:0;column:is_show_assistant;comment:'是否在助手显示, 0-否 1-是'"`
	IsShowH5        uint   `gorm:"default:0;column:is_show_h5;comment:'是否在H5显示, 0-否 1-是'"`
	IsShowDelivery  uint   `gorm:"default:0;column:is_show_delivery;comment:'是否在外送显示, 0-否 1-是'"`
	IsShowMenu      uint   `gorm:"default:0;column:is_show_menu;comment:'是否在电子菜单显示, 0-否 1-是'"`

	// 关联的商品列表
	ProductPackages []ProductPackage `gorm:"foreignKey:product_label_uuid;references:uuid"`
}
