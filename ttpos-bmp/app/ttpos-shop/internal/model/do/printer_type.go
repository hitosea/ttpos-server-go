// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// PrinterType is the golang structure of table ttpos_printer_type for DAO operations like Where/Data.
type PrinterType struct {
	g.Meta                `orm:"table:ttpos_printer_type, do:true"`
	Id                    interface{} // 自增ID
	Uuid                  interface{} // 打印机类型ID
	Name                  interface{} // 打印机类型名称
	MultiLanguageNameUuid interface{} // 多语言名称ID
	Key                   interface{} // 打印机类型key
	ConfigJson            interface{} // 打印机类型json配置,描述需要填写的字段
	CreateTime            interface{} // 创建时间(时间戳)
	UpdateTime            interface{} // 更新时间(时间戳)
	DeleteTime            interface{} // 删除时间(时间戳)
}
