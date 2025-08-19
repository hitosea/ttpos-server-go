// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// StaffOperationLog is the golang structure of table ttpos_staff_operation_log for DAO operations like Where/Data.
type StaffOperationLog struct {
	g.Meta       `orm:"table:ttpos_staff_operation_log, do:true"`
	Id           interface{} // 自增ID
	Uuid         interface{} // 操作日志ID
	StaffUuid    interface{} // 员工ID
	Title        interface{} // 标题
	Url          interface{} // 操作URL
	RequestData  interface{} // 请求数据
	ResponseData interface{} // 响应数据
	Type         interface{} // 操作类型
	Ip           interface{} // 操作IP
	Source       interface{} // 操作来源
	Agent        interface{} // 操作用户代理
	CreateTime   interface{} // 创建时间(时间戳)
	UpdateTime   interface{} // 更新时间(时间戳)
	DeleteTime   interface{} // 删除时间(时间戳)
}
