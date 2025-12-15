// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// OrderItem is the golang structure for table order_item.
type OrderItem struct {
	Id                    int64       `json:"id"                    orm:"id"                       description:"主键"`                    // 主键
	OrderUuid             string      `json:"orderUuid"             orm:"order_uuid"               description:"关联订单UUID"`              // 关联订单UUID
	PartnerItemId         string      `json:"partnerItemId"         orm:"partner_item_id"          description:"平台商品ID (Grab Item ID)"` // 平台商品ID (Grab Item ID)
	ItemName              string      `json:"itemName"              orm:"item_name"                description:"商品名称"`                  // 商品名称
	Quantity              int         `json:"quantity"              orm:"quantity"                 description:"数量"`                    // 数量
	Price                 float64     `json:"price"                 orm:"price"                    description:"单价"`                    // 单价
	TotalPrice            float64     `json:"totalPrice"            orm:"total_price"              description:"总价"`                    // 总价
	Specifications        string      `json:"specifications"        orm:"specifications"           description:"规格描述(JSON格式)"`          // 规格描述(JSON格式)
	Modifiers             string      `json:"modifiers"             orm:"modifiers"                description:"修改项/配料 (JSON)"`         // 修改项/配料 (JSON)
	Note                  string      `json:"note"                  orm:"note"                     description:"商品备注"`                  // 商品备注
	OutOfStockInstruction string      `json:"outOfStockInstruction" orm:"out_of_stock_instruction" description:"缺货处理方式"`                // 缺货处理方式
	CreatedAt             *gtime.Time `json:"createdAt"             orm:"created_at"               description:"创建时间"`                  // 创建时间
}
