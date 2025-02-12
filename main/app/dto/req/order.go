package req

import "ttpos-server-go/app/dto"

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
	SaleBillUuid  uint64 `form:"sale_bill_uuid" json:"sale_bill_uuid"`   // 销售账单UUID
	SaleOrderUuid uint64 `form:"sale_order_uuid" json:"sale_order_uuid"` // 销售订单UUID 当查看子订单信息的时候才需要传
}
