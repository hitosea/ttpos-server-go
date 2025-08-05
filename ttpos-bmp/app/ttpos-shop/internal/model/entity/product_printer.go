// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// ProductPrinter is the golang structure for table product_printer.
type ProductPrinter struct {
	Id                 uint   `json:"id"                 orm:"id"                   description:"自增ID"`                              // 自增ID
	Uuid               uint64 `json:"uuid"               orm:"uuid"                 description:"商品打印机ID"`                           // 商品打印机ID
	Name               string `json:"name"               orm:"name"                 description:"名称.厨显上叫档口"`                         // 名称.厨显上叫档口
	Status             int    `json:"status"             orm:"status"               description:"状态,1-open开启 1、0-close关闭"`           // 状态,1-open开启 1、0-close关闭
	PrintMode          int    `json:"printMode"          orm:"print_mode"           description:"打印模式,0-payment付款打印 1-kitchen送厨打印"`  // 打印模式,0-payment付款打印 1-kitchen送厨打印
	PrintMethod        int    `json:"printMethod"        orm:"print_method"         description:"打印方式,0-order整单打印 1-item按一菜一单打印"`    // 打印方式,0-order整单打印 1-item按一菜一单打印
	PrintProductSelect int    `json:"printProductSelect" orm:"print_product_select" description:"打印商品选择,0-category按商品分类 1-tag按打印标签"` // 打印商品选择,0-category按商品分类 1-tag按打印标签
	PrintModeScene     int    `json:"printModeScene"     orm:"print_mode_scene"     description:"打印模式场景,0-merge合并 1-separate分开"`     // 打印模式场景,0-merge合并 1-separate分开
	CreateTime         uint   `json:"createTime"         orm:"create_time"          description:"创建时间(时间戳)"`                         // 创建时间(时间戳)
	UpdateTime         uint   `json:"updateTime"         orm:"update_time"          description:"更新时间(时间戳)"`                         // 更新时间(时间戳)
	DeleteTime         uint   `json:"deleteTime"         orm:"delete_time"          description:"删除时间(时间戳)"`                         // 删除时间(时间戳)
}
