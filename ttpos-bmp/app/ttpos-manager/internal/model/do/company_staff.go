// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// CompanyStaff is the golang structure of table ttpos_company_staff for DAO operations like Where/Data.
type CompanyStaff struct {
	g.Meta      `orm:"table:ttpos_company_staff, do:true"`
	Id          interface{} // 自增ID
	Uuid        interface{} // 员工ID
	CompanyUuid interface{} // 集团ID
	Username    interface{} // 员工账号
	Phone       interface{} // 员工手机号
	IsSuper     interface{} // 是否超级管理员
	CreateTime  interface{} // 创建时间（时间戳）
	UpdateTime  interface{} // 更新时间（时间戳）
	DeleteTime  interface{} // 删除时间（时间戳）
}
