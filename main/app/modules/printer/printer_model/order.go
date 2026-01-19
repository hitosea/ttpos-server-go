package printer_model

import (
	"ttpos-server-go/app/dto/resp"
)

type Order struct {
	Uuid                   uint64                   `json:"uuid"`                      // 订单UUID-销售账单UUID/外卖订单UUID
	SaleOrderUuid          uint64                   `json:"sale_order_uuid"`           // 销售订单UUID
	OrderNo                string                   `json:"order_no"`                  // 订单号
	MealNum                uint                     `json:"meal_num"`                  // 就餐人数
	IsTakeoutBill          bool                     `json:"is_takeout_bill"`           // 是否是外送订单（传统店内外送）
	OrderSourceTakeoutText string                   `json:"order_source_takeout_text"` // 订单来源为外卖的文本
	SerialNo               string                   `json:"serial_no"`                 // 序列号
	OrderRemark            *resp.OrderRemarkResItem `json:"order_remark"`              // 整单备注
	DeskUuid               uint64                   `json:"desk_uuid"`                 // 桌台UUID
	Desk                   *OrderDesk               `json:"desk"`                      // 桌台
	UpdateTime             int64                    `json:"update_time"`               // 更新时间
	FinishTime             int64                    `json:"finish_time"`               // 完成时间
	IsTakeout              bool                     `json:"is_takeout"`                // 是否是外送订单（第三方外卖平台）
	// 订单商品列表
	Products []OrderProduct `json:"products"` // 订单商品列表
}

// Desk 桌台信息表,定义桌台的相关信息 ttpos_desk
type OrderDesk struct {
	DeskNo     string `json:"desk_no"`     // 桌位编号
	RegionUuid uint64 `json:"region_uuid"` // 桌台区域ID
}
