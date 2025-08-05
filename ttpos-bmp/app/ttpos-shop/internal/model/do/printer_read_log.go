// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// PrinterReadLog is the golang structure of table ttpos_printer_read_log for DAO operations like Where/Data.
type PrinterReadLog struct {
	g.Meta     `orm:"table:ttpos_printer_read_log, do:true"`
	Id         interface{} // 自增ID
	Uuid       interface{} // 打印读取日志ID
	LogUuid    interface{} // 打印uuid
	DeviceId   interface{} // 设备id
	CreateTime interface{} // 创建时间(时间戳)
	UpdateTime interface{} // 更新时间(时间戳)
	DeleteTime interface{} // 删除时间(时间戳)
}
