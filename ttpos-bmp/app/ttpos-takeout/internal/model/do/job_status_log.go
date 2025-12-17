// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// JobStatusLog is the golang structure of table takeout_job_status_log for DAO operations like Where/Data.
type JobStatusLog struct {
	g.Meta       `orm:"table:takeout_job_status_log, do:true"`
	Id           any         // 主键
	Uuid         any         // 全局唯一ID
	JobUuid      any         // 外送订单uuid
	StatusBefore any         // 变更前状态
	StatusAfter  any         // 变更后状态
	CreatedAt    *gtime.Time // 创建时间
	UpdatedAt    *gtime.Time // 更新时间
	DeletedAt    *gtime.Time // 软删除
}
