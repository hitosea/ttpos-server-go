package service

// 订单操作类型
const (
	OrderOpenTable           = "OPEN"                // 开台
	OrderSendKitchen         = "SEND"                // 送厨
	OrderRefundProduct       = "REFUND"              // 退菜
	OrderCancelRefundProduct = "CANCEL_REFUND"       // 取消退菜
	OrderChangeTable         = "CHANGE_TABLE"        // 转台
	OrderChangePrice         = "CHANGE_PRICE"        // 改价
	OrderRemark              = "CHANGE_PRICE"        // 备注
	OrderUpdateMealNum       = "UPDATE_MEAL_NUM"     // 修改桌台就餐人数
	OrderStayOrder           = "STAY_ORDER"          // 挂单
	OrderPickOrder           = "PICK_ORDER"          // 取单
	OrderProductFree         = "PRODUCT_FREE"        // 赠菜
	OrderCancelProductFree   = "CANCEL_PRODUCT_FREE" // 取消赠菜
	OrderProductMove         = "PRODUCT_MOVE"        // 转菜
	OrderDiscount            = "DISCOUNT"            // 优惠折扣
	OrderCancelDiscount      = "CANCEL_DISCOUNT"     // 撤销优惠折扣
	OrderSettle              = "SETTLE"              // 结账
	OrderReverseSettle       = "REVERSE_SETTLE"      // 反结账
	OrderRefund              = "REFUND"              // 退款
	OrderOrderTaking         = "ORDER_TAKING"        // 接单
	OrderOrderReject         = "ORDER_REJECT"        // 拒单
	OrderMergeTable          = "MERGE_TABLE"         // 并台
	OrderOrderCancel         = "ORDER_CANCEL"        // 整单取消
	OrderCheckoutDiscount    = "CHECKOUT_DISCOUNT"   // 结账手动抹零
	OrderSplitOrder          = "SPLIT_ORDER"         // 拆单
	OrderCancelSplitOrder    = "CANCEL_SPLIT_ORDER"  // 撤销拆单
	OrderAddProduct          = "ADD_PRODUCT"         // 增加菜品
	OrderDeleteProduct       = "DELETE_PRODUCT"      // 删除菜品
	OrderClock               = "CLOCK"               // 加钟
)

// // 开台操作
// type OptOpenDesk struct {
// 	MealNum  int    `json:"meal_num"`  // 就餐人数
// 	IsBuffet int    `json:"is_buffet"` // 是否自助餐
// 	TableId  int    `json:"table_id"`  // 桌台ID
// 	TableNo  string `json:"table_no"`  // 桌台号
// }

// // 撤销优惠折扣
// // {"parent_id":1993,"order_name":"1"}
// type OrderOperationLog struct {
// 	ParentId  int    `json:"parent_id"`  // 父级ID
// 	OrderName string `json:"order_name"` // 订单名称
// }

// // 取消退菜
// // {"order_product_id":9879,"product_id":980,"product_name":"{\"zh\":\"拌饭（单点）\",\"en\":\"Bibimbap (A La Carte)\",\"th\":\"บิบิมบับ (สั่งเดี่ยว)\",\"ja\":\"ビビンバ (単品)\",\"ko\":\"비빔밥 (단품)\",\"my\":\"ဘီဘီမွတ် (တစ်ခါသုံး)\",\"tr\":\"Bibimbap (Tek Kişilik)\",\"zhtw\":\"拌飯（單點）\"}","product_attr":"{\"th\":\"จาน\",\"en\":\"portion\",\"my\":\"ပိုင်း\",\"zh\":\"份\"}","num":1,"parent_id":1998,"order_name":"3"}
// type OrderProductOperationLog struct {
// 	OrderProductId int    `json:"order_product_id"` // 订单商品ID
// 	ProductId      int    `json:"product_id"`       // 商品ID
// 	ProductName    string `json:"product_name"`     // 商品名称
// 	ProductAttr    string `json:"product_attr"`     // 商品属性
// 	Num            int    `json:"num"`              // 数量
// 	ParentId       int    `json:"parent_id"`        // 父级ID
// 	OrderName      string `json:"order_name"`       // 订单名称
// }

// // 订单取消赠菜
// // {"order_product_id":10560,"product_id":995,"product_name":"{\"zh\":\"伯牙绝弦(商品属性测试)\",\"en\":\"Boya Breaks the Strings\",\"th\":\"ป๋อหยาคลายสายพิณ\",\"ja\":\"伯牙絶弦\",\"ko\":\"백아절현\",\"my\":\"Boya သည် ကြိုးပြိုထွက်သည်\",\"tr\":\"Boya'nın Teli Koparması\",\"zhtw\":\"伯牙絕弦\"}","product_attr":"{\"zh\":\"小杯\",\"en\":\"Small Cup\",\"th\":\"ถ้วยเล็ก\",\"ja\":\"小カップ\",\"ko\":\"작은 컵\",\"my\":\"ခွက်ငယ်\",\"tr\":\"Küçük Bardak\",\"zhtw\":\"小杯\"};{\"zh\":\"标准冰\",\"en\":\"Standard Ice\",\"th\":\"น้ำแข็งมาตรฐาน\",\"ja\":\"標準の氷\",\"ko\":\"표준 얼음\",\"my\":\"စံနှင်း\",\"tr\":\"Standart Buz\",\"zhtw\":\"標準冰\"};{\"zh\":\"标准糖\",\"en\":\"Standard sugar\",\"th\":\"น้ำตาลมาตรฐาน\",\"ja\":\"標準の砂糖\",\"ko\":\"표준 설탕\",\"my\":\"စံသတ်မှတ်ထားသော သကြား\",\"tr\":\"Standart şeker\",\"zhtw\":\"標準糖\"}","product_price":15,"total_num":1,"total_price":"15.00","parent_id":0,"order_name":""}
// type OrderProductOperationLogExt struct {
// 	OrderProductId int     `json:"order_product_id"` // 订单商品ID
// 	ProductId      int     `json:"product_id"`       // 商品ID
// 	ProductName    string  `json:"product_name"`     // 商品名称(多语言)
// 	ProductAttr    string  `json:"product_attr"`     // 商品属性(多语言)
// 	ProductPrice   float64 `json:"product_price"`    // 商品单价
// 	TotalNum       int     `json:"total_num"`        // 总数量
// 	TotalPrice     string  `json:"total_price"`      // 总价格
// 	ParentId       int     `json:"parent_id"`        // 父级ID
// 	OrderName      string  `json:"order_name"`       // 订单名称
// }

// // 撤销拆单

// // 改价
// // {"order_product_id":4353,"product_id":995,"product_name":"{\"zh\":\"伯牙绝弦(商品属性测试)\",\"en\":\"Boya Breaks the Strings\",\"th\":\"ป๋อหยาคลายสายพิณ\",\"ja\":\"伯牙絶弦\",\"ko\":\"백아절현\",\"my\":\"Boya သည် ကြိုးပြိုထွက်သည်\",\"tr\":\"Boya'nın Teli Koparması\",\"zhtw\":\"伯牙絕弦\"}","product_attr":"小杯;标准冰;标准糖","total_num":1,"price":"50"}
// type OrderProductPriceChangeLog struct {
// 	OrderProductId int    `json:"order_product_id"` // 订单商品ID
// 	ProductId      int    `json:"product_id"`       // 商品ID
// 	ProductName    string `json:"product_name"`     // 商品名称(多语言)
// 	ProductAttr    string `json:"product_attr"`     // 商品属性
// 	TotalNum       int    `json:"total_num"`        // 总数量
// 	Price          string `json:"price"`            // 改价后的价格
// }

// // 换台记录
// // {"old":{"table_id":189,"table_no":"A07"},"new":{"table_id":219,"table_no":"301"}}
// type TableChangeLog struct {
// 	Old struct {
// 		TableId int    `json:"table_id"` // 原桌台ID
// 		TableNo string `json:"table_no"` // 原桌台号
// 	} `json:"old"`
// 	New struct {
// 		TableId int    `json:"table_id"` // 新桌台ID
// 		TableNo string `json:"table_no"` // 新桌台号
// 	} `json:"new"`
// }

// // 结账抹零
// // {"operation":"add","rounding_type":1,"special_discount":0.02,"parent_id":2014,"order_name":"1"}
// type RoundingLog struct {
// 	Operation       string  `json:"operation"`        // 操作类型
// 	RoundingType    int     `json:"rounding_type"`    // 抹零类型
// 	SpecialDiscount float64 `json:"special_discount"` // 特殊折扣
// 	ParentId        int     `json:"parent_id"`        // 父级ID
// 	OrderName       string  `json:"order_name"`       // 订单名称
// }

// 优惠折扣
type OptDiscountLog struct {
	Price        string `json:"price"`         // 折扣金额
	DiscountType int    `json:"discount_type"` // 折扣类型
}

// 并台
type OptMergeTableLog []string // 存储桌台号数组

// 退款记录
type OptRefundLog struct {
	PayType []struct {
		OrderId        int     `json:"order_id"`         // 订单ID
		RefundId       int     `json:"refund_id"`        // 退款ID
		Price          float64 `json:"price"`            // 退款金额
		Value          int     `json:"value"`            // 支付方式值
		Name           string  `json:"name"`             // 支付方式名称
		Remark         string  `json:"remark"`           // 备注
		Source         int     `json:"source"`           // 来源
		PaymentOrderId int     `json:"payment_order_id"` // 支付订单ID
		RefundMoney    string  `json:"refund_money"`     // 退款金额
		ShopSupplierId int64   `json:"shop_supplier_id"` // 商户ID
		AppId          int64   `json:"app_id"`           // 应用ID
		Status         int     `json:"status"`           // 状态
	} `json:"pay_type"`
	RefundType    int           `json:"refund_type"`    // 退款类型
	RefundMethod  int           `json:"refund_method"`  // 退款方式
	RefundProduct []interface{} `json:"refund_product"` // 退款商品
	RefundBuffet  []interface{} `json:"refund_buffet"`  // 退款自助餐
	RefundDelay   []interface{} `json:"refund_delay"`   // 延迟退款
}

// 退菜记录
type OptRefundProductLog struct {
	OrderProductId int    `json:"order_product_id"` // 订单商品ID
	ProductId      int    `json:"product_id"`       // 商品ID
	ProductName    string `json:"product_name"`     // 商品名称
	ProductAttr    string `json:"product_attr"`     // 商品属性
	Num            int    `json:"num"`              // 退菜数量
	Reason         string `json:"reason"`           // 退菜原因
}

// 反结账记录
type OptReverseSettleLog struct {
	PayPrice string `json:"pay_price"` // 支付金额
	PayType  []struct {
		Name           string  `json:"name"`             // 支付方式名称
		SourceText     string  `json:"source_text"`      // 来源描述
		OrderId        int     `json:"order_id"`         // 订单ID
		Value          int     `json:"value"`            // 支付方式值
		Price          float64 `json:"price"`            // 金额
		DisabledCancel int     `json:"disabled_cancel"`  // 是否禁用取消
		PayStatus      int     `json:"pay_status"`       // 支付状态
		FeeMoney       float64 `json:"fee_money"`        // 手续费
		PaymentOrderId int     `json:"payment_order_id"` // 支付订单ID
		Source         int     `json:"source"`           // 来源
		Remark         string  `json:"remark"`           // 备注
	} `json:"pay_type"`
	ChangeDue     string `json:"change_due"`     // 找零
	IsFree        int    `json:"is_free"`        // 是否免单
	DiscountMoney int    `json:"discount_money"` // 折扣金额
}

// 送厨记录
type OptSendKitchenLog []struct {
	OrderProductId int    `json:"order_product_id"` // 订单商品ID
	ProductId      int    `json:"product_id"`       // 商品ID
	ProductName    string `json:"product_name"`     // 商品名称
	ProductAttr    string `json:"product_attr"`     // 商品属性
	TotalNum       int    `json:"total_num"`        // 总数量
}

// 结账记录
type OptSettleLog struct {
	PayPrice string `json:"pay_price"` // 支付金额
	PayType  []struct {
		Name           string  `json:"name"`             // 支付方式名称
		SourceText     string  `json:"source_text"`      // 来源描述
		OrderId        int     `json:"order_id"`         // 订单ID
		Value          int     `json:"value"`            // 支付方式值
		Price          float64 `json:"price"`            // 金额
		DisabledCancel int     `json:"disabled_cancel"`  // 是否禁用取消
		PayStatus      int     `json:"pay_status"`       // 支付状态
		FeeMoney       float64 `json:"fee_money"`        // 手续费
		PaymentOrderId int     `json:"payment_order_id"` // 支付订单ID
		Source         int     `json:"source"`           // 来源
		Remark         string  `json:"remark"`           // 备注
	} `json:"pay_type"`
	ActualPrice   string `json:"actual_price"`   // 实际金额
	ChangeDue     string `json:"change_due"`     // 找零
	IsFree        int    `json:"is_free"`        // 是否免单
	DiscountMoney string `json:"discount_money"` // 折扣金额
}

// 拆单记录
type OptSplitOrderLog struct {
	SplitOrder []struct {
		OrderId   int     `json:"order_id"`   // 订单ID
		OrderName string  `json:"order_name"` // 订单名称
		PayPrice  float64 `json:"pay_price"`  // 支付金额
	} `json:"split_order"`
}

// 修改就餐人数记录
type OptUpdateMealNumLog struct {
	OldMealNum int    `json:"old_meal_num"` // 原就餐人数
	NewMealNum string `json:"new_meal_num"` // 新就餐人数
}

// 赠菜记录
type OptProductFreeLog struct {
	OrderProductId int     `json:"order_product_id"` // 订单商品ID
	ProductId      int     `json:"product_id"`       // 商品ID
	ProductName    string  `json:"product_name"`     // 商品名称
	ProductAttr    string  `json:"product_attr"`     // 商品属性
	ProductPrice   float64 `json:"product_price"`    // 商品价格
	TotalNum       int     `json:"total_num"`        // 总数量
	TotalPrice     string  `json:"total_price"`      // 总价格
	FreeTagIds     []int   `json:"free_tag_ids"`     // 赠菜标签ID
	FreeRemark     string  `json:"free_remark"`      // 赠菜备注
}

// 转菜记录
type OptProductMoveLog struct {
	OrderProductId int    `json:"order_product_id"` // 订单商品ID
	ProductId      int    `json:"product_id"`       // 商品ID
	ProductName    string `json:"product_name"`     // 商品名称
	ProductAttr    string `json:"product_attr"`     // 商品属性
	TotalNum       int    `json:"total_num"`        // 总数量
	ToOrderId      int    `json:"to_order_id"`      // 目标订单ID
	ToTableId      int    `json:"to_table_id"`      // 目标桌台ID
	ToTableNo      string `json:"to_table_no"`      // 目标桌台号
}
