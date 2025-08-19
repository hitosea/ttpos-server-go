// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// CallbackMsg is the golang structure of table takeout_callback_msg for DAO operations like Where/Data.
type CallbackMsg struct {
	g.Meta         `orm:"table:takeout_callback_msg, do:true"`
	Id             interface{} // 主键
	CreatedAt      *gtime.Time // 创建时间
	UpdatedAt      *gtime.Time // 修改时间
	DeletedAt      *gtime.Time // 软删除
	Uuid           interface{} // 全局唯一ID
	TakeoutRefNo   interface{} // 外送系统订单号，如skootar.jobId
	Content        interface{} // 消息内容
	StatusDatetime *gtime.Time // 状态变更时间
}
