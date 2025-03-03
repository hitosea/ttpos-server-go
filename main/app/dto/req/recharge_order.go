package req

import "ttpos-server-go/app/dto"

// RechargeOrderListReq 充值订单列表查询
type RechargeOrderListReq struct {
	dto.PageReq              // 分页参数
	OrderNo           string `form:"order_no"`             // 订单编号
	DateType          int    `form:"date_type,default=-1"` // 日期类型 -1=全都、 0=今天、 1=昨天、 2=本周
	EnableCreateTime  bool   `form:"enable_create_time"`   // 启用添加时间 false-不启用，true-启用
	EnablePaymentTime bool   `form:"enable_payment_time"`  // 启用支付时间 false-不启用，true-启用
	QueryStartTime    int64  `form:"query_start_time"`     // 查询开始时间戳
	QueryEndTime      int64  `form:"query_end_time"`       // 查询结束时间戳
	Status            int    `form:"status,default=-1"`    // 充值订单状态, -1=全都、 0=待付款、1=已完成、2=已取消
}

type RechargeOrderUuidReq struct {
	Uuid uint64 `json:"uuid"` // 充值订单uuid
}
