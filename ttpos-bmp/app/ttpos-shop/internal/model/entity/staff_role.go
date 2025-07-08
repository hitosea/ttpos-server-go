// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// StaffRole is the golang structure for table staff_role.
type StaffRole struct {
	Id         uint   `json:"id"         orm:"id"          description:"自增ID"`      // 自增ID
	Uuid       uint64 `json:"uuid"       orm:"uuid"        description:"员工角色关系ID"`  // 员工角色关系ID
	StaffUuid  uint64 `json:"staffUuid"  orm:"staff_uuid"  description:"超管用户ID"`    // 超管用户ID
	RoleUuid   uint64 `json:"roleUuid"   orm:"role_uuid"   description:"角色ID"`      // 角色ID
	CreateTime uint   `json:"createTime" orm:"create_time" description:"创建时间(时间戳)"` // 创建时间(时间戳)
	UpdateTime uint   `json:"updateTime" orm:"update_time" description:"更新时间(时间戳)"` // 更新时间(时间戳)
	DeleteTime uint   `json:"deleteTime" orm:"delete_time" description:"删除时间(时间戳)"` // 删除时间(时间戳)
}
