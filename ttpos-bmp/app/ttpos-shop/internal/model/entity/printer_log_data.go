// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// PrinterLogData is the golang structure for table printer_log_data.
type PrinterLogData struct {
	Id         uint   `json:"id"         orm:"id"          description:""`         //
	Uuid       int64  `json:"uuid"       orm:"uuid"        description:"唯一ID"`     // 唯一ID
	LogUuid    int64  `json:"logUuid"    orm:"log_uuid"    description:"打印日志UUID"` // 打印日志UUID
	Data       string `json:"data"       orm:"data"        description:"打印数据"`     // 打印数据
	CreateTime int    `json:"createTime" orm:"create_time" description:"创建时间"`     // 创建时间
	UpdateTime int    `json:"updateTime" orm:"update_time" description:"更新时间"`     // 更新时间
	DeleteTime int    `json:"deleteTime" orm:"delete_time" description:"删除时间"`     // 删除时间
}
