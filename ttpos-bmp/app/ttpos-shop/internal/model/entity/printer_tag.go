// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// PrinterTag is the golang structure for table printer_tag.
type PrinterTag struct {
	Id         uint   `json:"id"         orm:"id"          description:"自增ID"`      // 自增ID
	Uuid       uint64 `json:"uuid"       orm:"uuid"        description:"打印机标签ID"`   // 打印机标签ID
	Name       string `json:"name"       orm:"name"        description:"名称"`        // 名称
	CreateTime uint   `json:"createTime" orm:"create_time" description:"创建时间(时间戳)"` // 创建时间(时间戳)
	UpdateTime uint   `json:"updateTime" orm:"update_time" description:"更新时间(时间戳)"` // 更新时间(时间戳)
	DeleteTime uint   `json:"deleteTime" orm:"delete_time" description:"删除时间(时间戳)"` // 删除时间(时间戳)
}
