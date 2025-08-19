// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// PrinterTemplate is the golang structure of table ttpos_printer_template for DAO operations like Where/Data.
type PrinterTemplate struct {
	g.Meta     `orm:"table:ttpos_printer_template, do:true"`
	Id         interface{} // 自增ID
	Uuid       interface{} // 打印机模板ID
	Name       interface{} // 打印名称
	Template   interface{} // 模板选择
	IsShowSku  interface{} // 是否显示SKU：0=不显示，1=显示
	CreateTime interface{} // 创建时间(时间戳)
	UpdateTime interface{} // 更新时间(时间戳)
	DeleteTime interface{} // 删除时间(时间戳)
}
