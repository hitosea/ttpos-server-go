package req

import "ttpos-server-go/app/dto"

var OrderReqMessage = map[string]string{
	"uuid.required":               "桌台uuid不能为空",
	"sale_bill_uuid.required":     "销售账单UUID不能为空",
	"sale_order_uuid.required":    "销售订单UUID不能为空",
	"order_product_uuid.required": "订单商品UUID不能为空",
	"price.required":              "商品价格不能为空",
	"population.required":         "人数不能为空",
	"remark.required":             "备注不能为空",
}

// 订单列表查询
type OrderListReq struct {
	dto.PageReq             // 分页参数
	OrderNo          string `form:"order_no"`             // 订单编号
	DateType         int    `form:"date_type,default=-1"` // 日期类型 -1=全都、 0=今天、 1=昨天、 2=本周
	EnableCreateTime bool   `form:"enable_create_time"`   // 启用开台时间 false-不启用，true-启用
	EnablePayTime    bool   `form:"enable_pay_time"`      // 启用支付时间 false-不启用，true-启用
	QueryStartTime   uint   `form:"query_start_time"`     // 查询开始时间戳
	QueryEndTime     uint   `form:"query_end_time"`       // 查询结束时间戳
	Status           int    `form:"status,default=-1"`    // 账单状态, -1=全都、 0=待付款、1=已完成、2=已取消
	BillType         int    `form:"bill_type,default=-1"` // 账单类型, -1=全都、 0=Desk桌台订单、1=OrderingFood点餐订单
}

// 订单信息查询
type OrderInfoReq struct {
	SaleBillUuid  uint64 `form:"sale_bill_uuid"`  // 销售账单UUID
	SaleOrderUuid uint64 `form:"sale_order_uuid"` // 销售订单UUID 当查看子订单信息的时候才需要传
}

// 订单取消
type OrderCancelReq struct {
	SaleBillUuid uint64 `json:"sale_bill_uuid"` // 销售账单UUID
	CancelReason string `json:"cancel_reason"`  // 取消原因
	Password     string `form:"password"`       // 高级密码 后台开启的时候才传
}

// 订单删除
type OrderDeleteReq struct {
	SaleBillUuid  uint64 `json:"sale_bill_uuid"`  // 销售账单UUID
	SaleOrderUuid uint64 `json:"sale_order_uuid"` // 销售订单UUID 传0的时候默认删除主单以及所有子单，不然只删除子单
}

// 是否可关闭订单
type OrderIsCellCloseReq struct {
	DeskUuid     uint64 `form:"desk_uuid"`      // 桌台UUID	   二选一, 销售账单UUID权重最大
	SaleBillUuid uint64 `form:"sale_bill_uuid"` // 销售账单UUID	二选一，销售账单UUID权重最大
}

// 删除订单商品
type OrderProductDeleteReq struct {
	SaleBillUuid     uint64 `json:"sale_bill_uuid" binding:"required"`     // 销售账单UUID
	SaleOrderUuid    uint64 `json:"sale_order_uuid" binding:"required"`    // 销售订单UUID
	OrderProductUuid uint64 `json:"order_product_uuid" binding:"required"` // 订单商品UUID
}

// 订单商品改价
type OrderProductChangePriceReq struct {
	SaleBillUuid     uint64  `json:"sale_bill_uuid" binding:"required"`     // 销售账单UUID
	SaleOrderUuid    uint64  `json:"sale_order_uuid" binding:"required"`    // 销售订单UUID
	OrderProductUuid uint64  `json:"order_product_uuid" binding:"required"` // 订单商品UUID
	Price            float64 `json:"price" binding:"required"`              // 改价
}

// 订单人数
type OrderChangePopulationReq struct {
	SaleBillUuid uint64 `json:"sale_bill_uuid" binding:"required"` // 销售账单UUID
	Population   int    `json:"population" binding:"required"`     // 人数
}

// 订单商品remark
type OrderProductRemarkReq struct {
	SaleBillUuid     uint64 `json:"sale_bill_uuid" binding:"required"`     // 销售账单UUID
	SaleOrderUuid    uint64 `json:"sale_order_uuid" binding:"required"`    // 销售订单UUID
	OrderProductUuid uint64 `json:"order_product_uuid" binding:"required"` // 订单商品UUID
	Remark           string `json:"remark"`                                // remark
}
