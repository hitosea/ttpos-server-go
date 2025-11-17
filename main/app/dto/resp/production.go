package resp

import (
	"ttpos-server-go/app/dto"
)

type ProductionGroup struct {
	LocaleName        *dto.LocaleResponse `json:"locale_name"`          // 序列号
	DiningMethod      uint                `json:"dining_method"`        // 用餐方式,0-堂食(店内就餐) 1-打包
	IsTakeoutBill     bool                `json:"is_takeout_bill"`      // 是否是外送订单
	ProductionList    ProductionList      `json:"product_list"`         // 送厨商品列表
	SaleBillUuid      uint64              `json:"sale_bill_uuid"`       // 销售账单Uuid
	IsSaleBillDeleted bool                `json:"is_sale_bill_deleted"` // 销售账单是否已删除
	OrderRemark       OrderRemarkRes      `json:"order_remark"`         // 整单备注
}

type ProductionList struct {
	List []ProductionItem `json:"list"` // 送厨商品
}

type ProductionItem struct {
	DiningMethod          uint               `json:"dining_method"`           // 用餐方式,0-堂食(店内就餐) 1-打包
	SerialNo              string             `json:"serial_no"`               // 序列号
	Uuid                  uint64             `json:"uuid"`                    // 送厨商品Uuid
	LocaleName            dto.LocaleResponse `json:"locale_name"`             // 送厨商品名称
	Num                   float64            `json:"num"`                     // 送厨商品数量
	InitNum               float64            `json:"init_num"`                // 送厨商品初始数量
	CreateTime            int64              `json:"create_time"`             // 送厨时间
	FinishedTime          int64              `json:"finished_time"`           // 完成时间
	MakeDuration          int64              `json:"make_duration"`           // 制作时长(秒)
	SendDuration          int64              `json:"send_duration"`           // 传菜时长(秒)
	ProductAttributeNames dto.LocaleResponse `json:"product_attribute_names"` // 商品属性
	Remark                string             `json:"remark"`                  // 备注
	IsSaleBillDeleted     bool               `json:"is_sale_bill_deleted"`    // 销售账单是否已删除
	MakeStatus            uint               `json:"make_status"`             // 制作状态，0-默认，未制作完成，1-已制作完成，2-已恢复到制作中
	MadeTime              int64              `json:"made_time"`               // 制作完成时间
	BatchTag              BatchTagInfo       `json:"batch_tag"`               // 分批类型
	OrderRemark           OrderRemarkRes     `json:"order_remark"`            // 整单备注
}

type BatchTagInfo struct {
	Uuid       uint64             `json:"uuid"`        // 分批类型UUID
	LocaleName dto.LocaleResponse `json:"locale_name"` // 分批类型名称，多语言
	Color      string             `json:"color"`       // 颜色值，如#FF0000
}

// ProductionListWithPagination 商品列表响应
type ProductionListWithPagination struct {
	SendKitchenNum float64           `json:"send_kitchen_num"` // 送厨商品数量，用于厨显来菜提醒
	List           []ProductionGroup `json:"list"`             // 分组列表
	FinishedList   ProductionList    `json:"finished_list"`    // 最近三个上菜历史
	Meta           dto.PageResponse  `json:"meta"`             // 分页信息
}

// ProductionHistory 上菜历史
type ProductionHistory struct {
	List []ProductionGroup `json:"list"`
}

// SendKitchenProduction 平板端和助手端已送厨商品列表
type SendKitchenProduction struct {
	List                []SendKitchenProductionItem `json:"list"`                  // 送厨商品列表
	IsSplitOrder        bool                        `json:"is_split_order"`        // 是否已拆单
	TotalPrice          string                      `json:"total_price"`           // 小计
	TotalProductNum     int                         `json:"total_product_num"`     // 总数量
	TotalProductPrice   string                      `json:"total_product_price"`   // 商品金额
	ServiceMoney        string                      `json:"service_money"`         // 服务费
	ConsumptionTaxMoney string                      `json:"consumption_tax_money"` // 消费税
	DiscountMoney       string                      `json:"discount_money"`        // 优惠折扣
	UserDiscountMoney   string                      `json:"user_discount_money"`   //
}

// SendKitchenProductionItem 平板端和助手端已送厨商品列表
type SendKitchenProductionItem struct {
	Products   []SendKitchenProduct `json:"products"`    // 商品列表
	CreateTime int64                `json:"create_time"` // 送厨时间
}

type SendKitchenProduct struct {
	LocaleName dto.LocaleResponse `json:"locale_name"` // 送厨商品名称
	Num        uint               `json:"num"`         // 送厨商品数量

	ProductAttributeNames string `json:"product_attribute_names"` // 商品属性
}
