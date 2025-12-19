// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// OrderSkootar is the golang structure of table takeout_order_skootar for DAO operations like Where/Data.
type OrderSkootar struct {
	g.Meta          `orm:"table:takeout_order_skootar, do:true"`
	Id              any // 主键ID
	Uuid            any // 唯一标识
	OrderUuid       any // 关联主订单UUID (takeout_order.uuid)
	SkootarId       any // 骑手ID
	SkootarName     any // 骑手名称
	SkootarPhone    any // 骑手电话
	SkootarRating   any // 骑手评分
	SkootarImageUrl any // 骑手头像
	CreatedAt       any // 创建时间
	UpdatedAt       any // 更新时间
	DeletedAt       any // 软删除
}
