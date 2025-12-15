// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// OrderStatusLog is the golang structure of table takeout_order_status_log for DAO operations like Where/Data.
type OrderStatusLog struct {
	g.Meta       `orm:"table:takeout_order_status_log, do:true"`
	Id           any         // 主键
	Uuid         any         // 唯一ID
	OrderUuid    any         // 关联订单UUID
	ProviderName any         // 渠道: grab
	StatusBefore any         // 变更前状态
	StatusAfter  any         // 变更后状态
	ChangeSource any         // 变更来源: WEBHOOK, API, SYSTEM
	DriverEta    any         // 骑手预计到达时间(分钟)
	Remark       any         // 备注或原因
	RawData      any         // 原始JSON数据
	CreatedAt    *gtime.Time // 创建时间
}
