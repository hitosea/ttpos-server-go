package resp

import "ttpos-server-go/app/dto"

// 桌台购物车
type ShopCart struct {
	IsDeskOrder   bool        `json:"is_desk_order"`    // 购物车类型 true:桌台购物车 false:点餐购物车
	Desk          *DeskInfo   `json:"desk,omitempty"`   // 桌台信息
	Buffet        *BuffetInfo `json:"buffet,omitempty"` // 自助餐信息
	DiningMethod  uint        `json:"dining_method"`    // 用餐方式 0:堂食 1:打包
	SaleOrderList []SaleOrder `json:"sale_order_list"`  // 销售订单列表
}

type AmountInfo struct {
	ProductOriginalAmount float64 `json:"product_origin_amount"`  // 商品金额(原价)
	ProductAmount         float64 `json:"product_amount"`         // 商品金额(折后价)
	ServiceAmount         float64 `json:"service_amount"`         // 服务费
	TaxAmount             float64 `json:"tax_amount"`             // 税费（商品税费+服务费税费）
	DiscountAmount        float64 `json:"discount_amount"`        // 优惠折扣金额(整单打折优惠金额+订单抹零金额)
	MemberDiscountAmount  float64 `json:"member_discount_amount"` // 会员优惠折扣金额
	Amount                float64 `json:"amount"`                 // 总金额。商品未含税时，总金额=商品金额(折后)+服务费+税费。商品已含税时，总金额=商品金额（折后，含商品消费税）+服务费+税费（只有服务费税）
}

// 自助餐信息
type BuffetInfo struct {
	EndTime    int64              `json:"end_time"`    // 自助餐结束时间
	LocaleName dto.LocaleResponse `json:"locale_name"` // 自助餐名称
}

// 桌台信息
type DeskInfo struct {
	Uuid      uint64 `json:"uuid"`       // 桌台ID
	DeskNo    string `json:"desk_no"`    // 桌台编号
	MealNum   uint   `json:"meal_num"`   // 就餐人数
	StartTime int64  `json:"start_time"` // 开台时间
}

// 购物车销售订单信息
type SaleOrder struct {
	Uuid        uint64     `json:"uuid"`
	OrderNo     string     `json:"order_no"`
	ProductList []Product  `json:"product_list"` // 商品列表
	AmountInfo  AmountInfo `json:"amount_info"`
}

// 购物车商品
type Product struct {
	Uuid                uint64             `json:"uuid"`                  // 商品uuid
	LocaleName          dto.LocaleResponse `json:"locale_name"`           // 自助餐名称
	LocaleAttributeName dto.LocaleResponse `json:"locale_attribute_name"` // 商品属性
	Num                 uint               `json:"num"`                   // 数量
	SalePrice           float64            `json:"price"`                 // 原价
	DiscountPrice       float64            `json:"discount_price"`        // 折扣价。折扣加为0的话表示没有对商品进行折扣，则显示原价
	Status              int                `json:"status"`                // 0: 未送厨 1:已送厨
	Remark              string             `json:"remark"`                // 备注
	IsMust              bool               `json:"is_must"`               // 是否必点
	IsGift              bool               `json:"is_gift"`               // 是否是赠菜
	IsCancel            bool               `json:"is_cancel"`             // 是否退菜
	AboutBuffet         AboutBuffet        `json:"about_buffet"`          // 自助餐信息
}

type AboutBuffet struct {
	IsCustomer bool `json:"is_customer"` // 是否是自助餐顾客
	IsDelay    bool `json:"is_delay"`    // 是否是加钟
}
