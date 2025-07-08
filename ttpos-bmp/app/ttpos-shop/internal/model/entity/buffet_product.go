// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// BuffetProduct is the golang structure for table buffet_product.
type BuffetProduct struct {
	Id                 uint   `json:"id"                 orm:"id"                   description:"自增ID"`              // 自增ID
	Uuid               uint64 `json:"uuid"               orm:"uuid"                 description:"自助餐商品ID"`           // 自助餐商品ID
	BuffetPackageUuid  uint64 `json:"buffetPackageUuid"  orm:"buffet_package_uuid"  description:"自助餐套餐ID"`           // 自助餐套餐ID
	ProductPackageUuid uint64 `json:"productPackageUuid" orm:"product_package_uuid" description:"商品包ID"`             // 商品包ID
	IsShowCashier      int    `json:"isShowCashier"      orm:"is_show_cashier"      description:"是否在收银台显示, 0-否 1-是"` // 是否在收银台显示, 0-否 1-是
	IsShowTablet       int    `json:"isShowTablet"       orm:"is_show_tablet"       description:"是否在平板显示, 0-否 1-是"`  // 是否在平板显示, 0-否 1-是
	IsShowKitchen      int    `json:"isShowKitchen"      orm:"is_show_kitchen"      description:"是否在厨房显示, 0-否 1-是"`  // 是否在厨房显示, 0-否 1-是
	IsShowAssistant    int    `json:"isShowAssistant"    orm:"is_show_assistant"    description:"是否在助手显示, 0-否 1-是"`  // 是否在助手显示, 0-否 1-是
	Limit              int    `json:"limit"              orm:"limit"                description:"限购数量"`              // 限购数量
	CreateTime         uint   `json:"createTime"         orm:"create_time"          description:"创建时间(时间戳)"`         // 创建时间(时间戳)
	UpdateTime         uint   `json:"updateTime"         orm:"update_time"          description:"更新时间(时间戳)"`         // 更新时间(时间戳)
	DeleteTime         uint   `json:"deleteTime"         orm:"delete_time"          description:"删除时间(时间戳)"`         // 删除时间(时间戳)
}
