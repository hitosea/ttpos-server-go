// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// PrinterTag is the golang structure of table ttpos_printer_tag for DAO operations like Where/Data.
type PrinterTag struct {
	g.Meta     `orm:"table:ttpos_printer_tag, do:true"`
	Id         interface{} // 自增ID
	Uuid       interface{} // 打印机标签ID
	Name       interface{} // 名称
	CreateTime interface{} // 创建时间(时间戳)
	UpdateTime interface{} // 更新时间(时间戳)
	DeleteTime interface{} // 删除时间(时间戳)
}
