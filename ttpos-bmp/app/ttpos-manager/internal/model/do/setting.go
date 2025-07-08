// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// Setting is the golang structure of table ttpos_setting for DAO operations like Where/Data.
type Setting struct {
	g.Meta     `orm:"table:ttpos_setting, do:true"`
	Key        interface{} // 设置项标示
	Describe   interface{} // 设置项描述
	Values     interface{} // 设置内容（json格式）
	CreateTime interface{} // 创建时间（时间戳）
	UpdateTime interface{} // 更新时间（时间戳）
	DeleteTime interface{} // 删除时间（时间戳）
}
