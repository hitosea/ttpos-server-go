// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// Printer is the golang structure of table ttpos_printer for DAO operations like Where/Data.
type Printer struct {
	g.Meta          `orm:"table:ttpos_printer, do:true"`
	Id              interface{} // 自增ID
	Uuid            interface{} // 打印机ID
	Name            interface{} // 打印机名称
	PrinterTypeUuid interface{} // 打印机类型ID
	ConfigJson      interface{} // 打印机json配置
	Copies          interface{} // 打印份数
	Sort            interface{} // 排序
	CreateTime      interface{} // 创建时间(时间戳)
	UpdateTime      interface{} // 更新时间(时间戳)
	DeleteTime      interface{} // 删除时间(时间戳)
}
