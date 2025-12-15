// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// OrderItem is the golang structure of table takeout_order_item for DAO operations like Where/Data.
type OrderItem struct {
	g.Meta                `orm:"table:takeout_order_item, do:true"`
	Id                    any         // 主键
	OrderUuid             any         // 关联订单UUID
	PartnerItemId         any         // 平台商品ID (Grab Item ID)
	ItemName              any         // 商品名称
	Quantity              any         // 数量
	Price                 any         // 单价
	TotalPrice            any         // 总价
	Specifications        any         // 规格描述(JSON格式)
	Modifiers             any         // 修改项/配料 (JSON)
	Note                  any         // 商品备注
	OutOfStockInstruction any         // 缺货处理方式
	CreatedAt             *gtime.Time // 创建时间
}
