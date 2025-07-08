// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// StaffShiftSnapshot is the golang structure for table staff_shift_snapshot.
type StaffShiftSnapshot struct {
	Id           uint   `json:"id"           orm:"id"             description:""`       //
	Uuid         uint64 `json:"uuid"         orm:"uuid"           description:"交班快照ID"` // 交班快照ID
	ShiftLogUuid uint64 `json:"shiftLogUuid" orm:"shift_log_uuid" description:"交班记录ID"` // 交班记录ID
	Content      string `json:"content"      orm:"content"        description:"快照json"` // 快照json
	CreateTime   int    `json:"createTime"   orm:"create_time"    description:"创建时间"`   // 创建时间
	UpdateTime   int    `json:"updateTime"   orm:"update_time"    description:"更新时间"`   // 更新时间
	DeleteTime   int    `json:"deleteTime"   orm:"delete_time"    description:"删除时间"`   // 删除时间
}
