package resp

import "ttpos-server-go/app/dto"

// 桌台购物车
type ShopCart struct {
	SaleBillUuid  uint64               `json:"sale_bill_uuid"`       // 销售账单ID
	Takeout       *bool                `json:"takeout,omitempty"`    // 是否是打包订单，false:堂食订单 true:打包订单。只有点餐订单才有这个字段
	IsDeskOrder   bool                 `json:"is_desk_order"`        // 购物车类型 true:桌台购物车 false:点餐购物车
	IsLock        bool                 `json:"is_lock"`              // 购物车是否锁定 true:锁定 false:未锁定
	Desk          *DeskInfo            `json:"desk,omitempty"`       // 桌台信息
	Buffet        *BuffetInfo          `json:"buffet,omitempty"`     // 自助餐信息
	MustPlans     *ProductMustPlanList `json:"must_plans,omitempty"` // 必点方案列表信息
	DiningMethod  uint                 `json:"dining_method"`        // 用餐方式 0:堂食 1:打包。与Takeout重复，废弃
	SaleOrderList []SaleOrder          `json:"sale_order_list"`      // 销售订单列表
}

// UnSendKitchen 未送厨商品
type UnSendKitchen struct {
	List          []Product `json:"product_list"`   // 商品列表
	ProductAmount float64   `json:"product_amount"` // 商品金额(折后价)
}

// SendKitchen 未送厨商品
type SendKitchen struct {
	List       []SendKitchenProductGroup `json:"list"`
	AmountInfo AmountInfo                `json:"amount_info"`
}

type SendKitchenProductGroup struct {
	SendKitchenTime int64     `json:"send_kitchen_time"` // 送厨时间
	List            []Product `json:"list"`              // 商品列表
}

// 送厨接口响应：商品 XXX 已下架，请选择其他商品
// 送厨接口响应：规格 商品名称 已下架，请选择其他规格
// 送厨接口响应：以下商品库存不足，请删除后再下单 -商品名稱1（规格大份）-商品名稱2（规格小份）
// 送厨接口响应：已送厨和本次要送厨的商品未选择必点商品，确定要继续送厨吗？· 方案名称1 少点1份 · 方案名称2 少点2份
// 送厨接口响应：以下商品价格有变动，请核对后再下单 -商品名稱1（规格大份）-商品名稱2（规格小份）
// 送厨接口响应：以下商品超出限购数量，请在限购数量内下单 -商品名稱1（规格大份）-商品名稱2（规格小份）
type OrderCheckRes struct {
	Products            *CartProductList     `json:"products,omitempty"`
	ProductMustPlanList *ProductMustPlanList `json:"product_must_plans,omitempty"`
}

type OrderCheckServiceRes struct {
	Code int `json:"code"`
	OrderCheckRes
}

type CartProductList struct {
	List []Product `json:"list"`
}

type InstantShopCart struct {
	SaleBillUuid  uint64      `json:"sale_bill_uuid"`  // 销售账单ID
	DiningMethod  uint        `json:"dining_method"`   // 用餐方式 0:堂食 1:打包
	SaleOrderList []SaleOrder `json:"sale_order_list"` // 销售订单列表
}

type AmountInfo struct {
	ProductOriginalAmount float64 `json:"product_origin_amount"`  // 商品金额(原价)
	ProductAmount         float64 `json:"product_amount"`         // 商品金额(折后价)
	ServiceAmount         float64 `json:"service_amount"`         // 服务费
	TaxAmount             float64 `json:"tax_amount"`             // 税费（商品税费+服务费税费）
	DiscountAmount        float64 `json:"discount_amount"`        // 优惠折扣金额(整单打折优惠金额+订单抹零金额)
	MemberDiscountAmount  float64 `json:"member_discount_amount"` // 会员优惠折扣金额
	Amount                float64 `json:"amount"`                 // 总金额。商品未含税时，总金额=商品金额(折后)+服务费+税费。商品已含税时，总金额=商品金额（折后，含商品消费税）+服务费+税费（只有服务费税）
	ProductNum            uint    `json:"product_num"`            // 总数量，用于点餐助手、平板端、h5
}

// BuffetInfo 自助餐信息
type BuffetInfo struct {
	RemainingSeconds int64              `json:"remaining_seconds"` // 自助餐还剩余多少秒。可以为负数，表示自助餐已经结束了多少秒
	IsTimeLimited    bool               `json:"is_time_limited"`   // 是否限时
	LocaleName       dto.LocaleResponse `json:"locale_name"`       // 自助餐名称
}

// DeskInfo 桌台信息
type DeskInfo struct {
	Uuid      uint64 `json:"uuid"`       // 桌台ID
	DeskNo    string `json:"desk_no"`    // 桌台编号
	MealNum   uint   `json:"meal_num"`   // 就餐人数
	StartTime int64  `json:"start_time"` // 开台时间
	Duration  int64  `json:"duration"`   // 用餐时长，单位：秒，表示从开台时间到当前时间的时间差。避免前端计算时间不准
}

// SaleOrder 购物车销售订单信息
type SaleOrder struct {
	Uuid               uint64     `json:"uuid"`
	OrderNo            string     `json:"order_no"`
	Status             uint       `json:"status"`               // 订单状态, 0-未结账 1-已结账
	IsDiscount         bool       `json:"is_discount"`          // 是否存在折扣 true:存在 false:不存在
	CustomDiscountRate float64    `json:"custom_discount_rate"` // 订单改价折扣率
	ZeroRule           uint8      `json:"zero_rule"`            // 订单抹零规则
	ProductList        []Product  `json:"product_list"`         // 商品列表
	ProductNum         int        `json:"product_num"`          // 商品数量
	AmountInfo         AmountInfo `json:"amount_info"`
}

// Product 购物车商品
type Product struct {
	Uuid                uint64             `json:"uuid"`                  // 商品uuid
	LocaleName          dto.LocaleResponse `json:"locale_name"`           // 商品名称。商品名称、自助餐名称、自助餐加钟名称
	LocaleAttributeName dto.LocaleResponse `json:"locale_attribute_name"` // 商品属性
	Num                 uint               `json:"num"`                   // 数量
	SalePrice           float64            `json:"price"`                 // 原价
	DiscountPrice       float64            `json:"discount_price"`        // 折扣价,折后。折扣价不等于原价时，前端要显示出折扣价。
	Status              int                `json:"status"`                // 0: 未送厨 1:已送厨
	Remark              string             `json:"remark"`                // 备注
	IsMust              bool               `json:"is_must"`               // 是否必点
	IsGift              bool               `json:"is_gift"`               // 是否是赠菜
	IsBuffet            bool               `json:"is_buffet"`             // 是否是自助餐
	IsCancel            bool               `json:"is_cancel"`             // 是否退菜
	AboutBuffet         AboutBuffet        `json:"about_buffet"`          // 自助餐信息
	SendKitchenTime     int64              `json:"send_kitchen_time"`     // 送厨时间
	Sign                string             `json:"sign"`                  // 签名，用于合并商品
}

type AboutBuffet struct {
	IsCustomer       bool   `json:"is_customer"`        // 是否是自助餐顾客
	IsDelay          bool   `json:"is_delay"`           // 是否是加钟
	BuffetUuid       uint64 `json:"buffet_uuid"`        // 自助餐Id
	CustomerTypeUuid uint64 `json:"customer_type_uuid"` // 自助餐顾客类型uuid
}

type OrderFinishResp struct {
	SaleBillUuid  uint64        `json:"sale_bill_uuid"`  // 销售账单uuid
	SaleOrderUuid uint64        `json:"sale_order_uuid"` // 销售订单uuid
	AmountInfo    PayAmountInfo `json:"amount_info"`     // 金额信息
	PayMethodList PayMethodList `json:"pay_methods"`     // 支付方式列表
}

type PayAmountInfo struct {
	OrderAmount  float64 `json:"order_amount"`  // 订单应收金额
	PayAmount    float64 `json:"pay_amount"`    // 订单实收金额
	ChangeAmount float64 `json:"change_amount"` // 找零金额
}
type PayMethodList struct {
	List []PayMethod `json:"list"`
}

type PayMethod struct {
	Uuid uint64 `json:"uuid"` // 支付方式uuid
	Name string `json:"name"` // 支付方式名称
}

// BuffetProductList 自助餐商品列表
type BuffetProductList struct {
	List []BuffetProduct `json:"list"` // 商品列表
}

type BuffetProduct struct {
	Uuid uint64 `json:"uuid"` // 商品uuid
	Name string `json:"name"` // 商品名称. 不用于前端展示，仅用于开发核对接口数据是否正确
}
