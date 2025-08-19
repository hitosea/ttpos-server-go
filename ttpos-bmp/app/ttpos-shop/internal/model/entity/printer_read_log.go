// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// PrinterReadLog is the golang structure for table printer_read_log.
type PrinterReadLog struct {
	Id         uint   `json:"id"         orm:"id"          description:"自增ID"`      // 自增ID
	Uuid       uint64 `json:"uuid"       orm:"uuid"        description:"打印读取日志ID"`  // 打印读取日志ID
	LogUuid    int    `json:"logUuid"    orm:"log_uuid"    description:"打印uuid"`    // 打印uuid
	DeviceId   string `json:"deviceId"   orm:"device_id"   description:"设备id"`      // 设备id
	CreateTime uint   `json:"createTime" orm:"create_time" description:"创建时间(时间戳)"` // 创建时间(时间戳)
	UpdateTime uint   `json:"updateTime" orm:"update_time" description:"更新时间(时间戳)"` // 更新时间(时间戳)
	DeleteTime uint   `json:"deleteTime" orm:"delete_time" description:"删除时间(时间戳)"` // 删除时间(时间戳)
}
