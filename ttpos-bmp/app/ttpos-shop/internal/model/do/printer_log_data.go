// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// PrinterLogData is the golang structure of table ttpos_printer_log_data for DAO operations like Where/Data.
type PrinterLogData struct {
	g.Meta     `orm:"table:ttpos_printer_log_data, do:true"`
	Id         interface{} //
	Uuid       interface{} // 唯一ID
	LogUuid    interface{} // 打印日志UUID
	Data       interface{} // 打印数据
	CreateTime interface{} // 创建时间
	UpdateTime interface{} // 更新时间
	DeleteTime interface{} // 删除时间
}
