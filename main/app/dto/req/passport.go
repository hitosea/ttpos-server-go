package req

type GetServerPublicKeyRequest struct {
	ClientId string `form:"client_id" binding:"required"`
	Type     string `form:"type" binding:"required,oneof=jsencrypt"`
}

var GetServerPublicKeyRequestMessage = map[string]string{
	"client_id.required": "client_id不能为空",
}

type LianLianCallbackRequest struct {
	CompanyUuid     string `json:"shop_supplier_id"`
	MerchantOrderNo string `json:"merchant_order_no"`
	MerchantUserId  string `json:"merchant_user_id"`
	PayTypeDesc     string `json:"pay_type_desc"`
	PayStatus       int    `json:"pay_status"`
	PaymentId       string `json:"payment_id"`
	OrderAmount     string `json:"order_amount"`
	OrderCurrency   string `json:"order_currency"`
	PayAt           string `json:"pay_at"`
}
