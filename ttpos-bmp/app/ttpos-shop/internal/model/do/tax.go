// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// Tax is the golang structure of table ttpos_tax for DAO operations like Where/Data.
type Tax struct {
	g.Meta     `orm:"table:ttpos_tax, do:true"`
	Id         interface{} // 自增ID
	Uuid       interface{} // 税率ID
	Name       interface{} // 名称
	TaxRate    interface{} // 税率
	CreateTime interface{} // 创建时间(时间戳)
	UpdateTime interface{} // 更新时间(时间戳)
	DeleteTime interface{} // 删除时间(时间戳)
}
