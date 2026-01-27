package req

type GetServerPublicKeyRequest struct {
	ClientId string `form:"client_id" binding:"required"`
	Type     string `form:"type" binding:"required,oneof=jsencrypt"`
}

var GetServerPublicKeyRequestMessage = map[string]string{
	"client_id.required": "client_id不能为空",
}

type LianLianCallbackRequest struct {
	CompanyUuid     string `json:"shop_supplier_id"`  // 商户ID
	MerchantOrderNo string `json:"merchant_order_no"` // 商户订单号
	MerchantUserId  string `json:"merchant_user_id"`  // 商户用户ID
	PayTypeDesc     string `json:"pay_type_desc"`     // 支付类型描述
	PayStatus       int    `json:"pay_status"`        // 支付状态
	PaymentId       string `json:"payment_id"`        // 支付ID
	OrderAmount     string `json:"order_amount"`      // 订单金额
	OrderCurrency   string `json:"order_currency"`    // 订单币种
	PayAt           string `json:"pay_at"`            // 支付时间
}

type LianLianRefundCallbackRequest struct {
	CompanyUuid      string `json:"shop_supplier_id"`
	RefundStatus     string `json:"refund_status"` // 'RS' 等于已经支付
	RefundOrderId    string `json:"refund_order_id"`
	PaymentOrderId   string `json:"payment_order_id"`
	MerchantRefundId string `json:"merchant_refund_id"`
}

// CacheHitRateStatsReq 缓存命中率统计查询请求
type CacheHitRateStatsReq struct {
	Top  int    `form:"top,default=10" binding:"omitempty,min=1,max=100"`                // Top N，默认10，最大100
	Sort string `form:"sort,default=total" binding:"omitempty,oneof=total hits hitrate"` // 排序方式：total（总访问次数）、hits（命中次数）、hitrate（命中率），默认total
	Key  string `form:"key"`                                                             // 指定 key，如果提供则只返回该 key 的统计信息
}

var CacheHitRateStatsReqMessage = map[string]string{
	"top.min":    "Top N 不能小于1",
	"top.max":    "Top N 不能大于100",
	"sort.oneof": "排序方式必须是 total、hits 或 hitrate",
}
