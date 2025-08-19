// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// CompanyStaff is the golang structure for table company_staff.
type CompanyStaff struct {
	Id          int    `json:"id"          orm:"id"           description:"自增ID"`      // 自增ID
	Uuid        uint64 `json:"uuid"        orm:"uuid"         description:"员工ID"`      // 员工ID
	CompanyUuid uint64 `json:"companyUuid" orm:"company_uuid" description:"集团ID"`      // 集团ID
	Username    string `json:"username"    orm:"username"     description:"员工账号"`      // 员工账号
	Phone       string `json:"phone"       orm:"phone"        description:"员工手机号"`     // 员工手机号
	IsSuper     int    `json:"isSuper"     orm:"is_super"     description:"是否超级管理员"`   // 是否超级管理员
	CreateTime  int    `json:"createTime"  orm:"create_time"  description:"创建时间（时间戳）"` // 创建时间（时间戳）
	UpdateTime  int    `json:"updateTime"  orm:"update_time"  description:"更新时间（时间戳）"` // 更新时间（时间戳）
	DeleteTime  int    `json:"deleteTime"  orm:"delete_time"  description:"删除时间（时间戳）"` // 删除时间（时间戳）
}
