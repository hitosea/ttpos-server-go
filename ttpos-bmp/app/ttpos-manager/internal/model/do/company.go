// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// Company is the golang structure of table ttpos_company for DAO operations like Where/Data.
type Company struct {
	g.Meta        `orm:"table:ttpos_company, do:true"`
	Id            interface{} // 自增ID
	Uuid          interface{} // 集团ID
	Name          interface{} // 集团名称
	Logo          interface{} // logo
	ExpireTime    interface{} // 过期时间;not null
	AuthDay       interface{} // 授权时间(天) 0为永不过期
	Status        interface{} // 状态 1-启用 0-禁用;not null
	AuthStartTime interface{} // 授权开始时间（时间戳）
	OldCompanyId  interface{} // 原商家ID
	CreateTime    interface{} // 创建时间（时间戳）
	UpdateTime    interface{} // 更新时间（时间戳）
	DeleteTime    interface{} // 删除时间（时间戳）
}
