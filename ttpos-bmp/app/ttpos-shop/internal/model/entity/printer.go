// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// Printer is the golang structure for table printer.
type Printer struct {
	Id              uint   `json:"id"              orm:"id"                description:"自增ID"`      // 自增ID
	Uuid            uint64 `json:"uuid"            orm:"uuid"              description:"打印机ID"`     // 打印机ID
	Name            string `json:"name"            orm:"name"              description:"打印机名称"`     // 打印机名称
	PrinterTypeUuid uint64 `json:"printerTypeUuid" orm:"printer_type_uuid" description:"打印机类型ID"`   // 打印机类型ID
	ConfigJson      string `json:"configJson"      orm:"config_json"       description:"打印机json配置"` // 打印机json配置
	Copies          uint   `json:"copies"          orm:"copies"            description:"打印份数"`      // 打印份数
	Sort            uint   `json:"sort"            orm:"sort"              description:"排序"`        // 排序
	CreateTime      uint   `json:"createTime"      orm:"create_time"       description:"创建时间(时间戳)"` // 创建时间(时间戳)
	UpdateTime      uint   `json:"updateTime"      orm:"update_time"       description:"更新时间(时间戳)"` // 更新时间(时间戳)
	DeleteTime      uint   `json:"deleteTime"      orm:"delete_time"       description:"删除时间(时间戳)"` // 删除时间(时间戳)
}
