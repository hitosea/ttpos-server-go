package sms

// 模板ID常量
const (
	// TemplateMemberConsumption 会员消费模板ID
	TemplateMemberConsumption = "member_consumption"
	// TemplateMemberRecharge 会员充值模板ID
	TemplateMemberRecharge = "member_recharge"
	// TemplateMemberRechargeRefund 会员充值退款模板ID
	TemplateMemberRechargeRefund = "member_recharge_refund"
	// TemplateMemberOrderRefund 会员用餐订单退款模板ID
	TemplateMemberOrderRefund = "member_order_refund"
	// TemplateMemberSendCode 会员发送验证码模板ID
	TemplateMemberSendCode = "member_send_code"
	// TemplateMemberPoints 会员积分模板ID
	TemplateMemberPoints = "referral_consumption_got_points"
	// TemplateMemberCoupon 会员优惠券模板ID
	TemplateMemberCoupon = "referral_consumption_got_coupons"
	// TemplateDeliveryOrderCanceledBySelf 外送订单取消模板ID
	TemplateDeliveryOrderCanceledBySelf = "delivery_order_canceled_by_self"
)

// 语言常量
const (
	// LanguageChinese 中文
	LanguageChinese = "zh"
	// LanguageEnglish 英文
	LanguageEnglish = "en"
)

// API路径常量
const (
	// APIPathSend 发送短信API路径
	APIPathSend = "/api/sms/send"
	// APIPathQuery 查询短信状态API路径
	APIPathQuery = "/api/sms/query"
	// APIPathHealth 健康检查API路径
	APIPathHealth = "/api/health"
	// APIPathCheckKey API密钥检查路径
	APIPathCheckKey = "/api/sms/check_api_key"
)

// 响应码常量
const (
	// ResponseCodeSuccess 成功响应码
	ResponseCodeSuccess = 0
	// ResponseCodeInvalidAPIKey 无效的API密钥响应码
	ResponseCodeInvalidAPIKey = 401
)
