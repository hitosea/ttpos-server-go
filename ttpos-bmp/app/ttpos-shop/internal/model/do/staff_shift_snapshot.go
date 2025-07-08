// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// StaffShiftSnapshot is the golang structure of table ttpos_staff_shift_snapshot for DAO operations like Where/Data.
type StaffShiftSnapshot struct {
	g.Meta       `orm:"table:ttpos_staff_shift_snapshot, do:true"`
	Id           interface{} //
	Uuid         interface{} // 交班快照ID
	ShiftLogUuid interface{} // 交班记录ID
	Content      interface{} // 快照json
	CreateTime   interface{} // 创建时间
	UpdateTime   interface{} // 更新时间
	DeleteTime   interface{} // 删除时间
}
