// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// PrinterTemplate is the golang structure for table printer_template.
type PrinterTemplate struct {
	Id         uint   `json:"id"         orm:"id"          description:"自增ID"`      // 自增ID
	Uuid       uint64 `json:"uuid"       orm:"uuid"        description:"打印机模板ID"`   // 打印机模板ID
	Name       string `json:"name"       orm:"name"        description:"打印名称"`      // 打印名称
	Template   int    `json:"template"   orm:"template"    description:"模板选择"`      // 模板选择
	IsShowSku  int    `json:"isShowSku"  orm:"is_show_sku" description:"是否显示SKU"`   // 是否显示SKU：0=不显示，1=显示
	CreateTime uint   `json:"createTime" orm:"create_time" description:"创建时间(时间戳)"` // 创建时间(时间戳)
	UpdateTime uint   `json:"updateTime" orm:"update_time" description:"更新时间(时间戳)"` // 更新时间(时间戳)
	DeleteTime uint   `json:"deleteTime" orm:"delete_time" description:"删除时间(时间戳)"` // 删除时间(时间戳)
}
