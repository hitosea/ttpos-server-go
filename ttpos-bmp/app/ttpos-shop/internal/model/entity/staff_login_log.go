// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// StaffLoginLog is the golang structure for table staff_login_log.
type StaffLoginLog struct {
	Id         uint   `json:"id"         orm:"id"          description:"自增ID"`   // 自增ID
	Uuid       uint64 `json:"uuid"       orm:"uuid"        description:"UUID"`   // UUID
	StaffUuid  uint64 `json:"staffUuid"  orm:"staff_uuid"  description:"员工UUID"` // 员工UUID
	Username   string `json:"username"   orm:"username"    description:"用户名"`    // 用户名
	Ip         string `json:"ip"         orm:"ip"          description:"登录ip"`   // 登录ip
	Result     string `json:"result"     orm:"result"      description:"登录结果"`   // 登录结果
	CreateTime uint   `json:"createTime" orm:"create_time" description:"创建时间"`   // 创建时间
	UpdateTime uint   `json:"updateTime" orm:"update_time" description:"更新时间"`   // 更新时间
	DeleteTime uint   `json:"deleteTime" orm:"delete_time" description:"删除时间"`   // 删除时间
}
