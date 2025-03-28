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
	CompanyUuid           string `json:"shop_supplier_id"`
	RefundStatus          string `json:"refund_status"` // 'RS' 等于已经支付
	RefundOrderId         string `json:"refund_order_id"`
	PaymentOrderId        string `json:"payment_order_id"`
	MerchantRefundOrderNo string `json:"merchant_refund_id"`
}
