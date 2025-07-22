package sms

// SMSClient 短信客户端接口
type SMSClient interface {
	// SendMemberConsumptionSMS 发送会员消费短信
	SendMemberConsumptionSMS(phone, language string, params *MemberConsumptionRequest) (*SMSResponse, error)
	// SendMemberRechargeSMS 发送会员充值短信
	SendMemberRechargeSMS(phone, language string, params *MemberRechargeRequest) (*SMSResponse, error)
	// SendMemberRechargeRefundSMS 发送会员充值退款短信
	SendMemberRechargeRefundSMS(phone, language string, params *MemberRechargeRefundRequest) (*SMSResponse, error)
	// SendMemberOrderRefundSMS 发送会员用餐订单退款短信
	SendMemberOrderRefundSMS(phone, language string, params *MemberOrderRefundRequest) (*SMSResponse, error)
	// SendMemberCodeSMS 发送会员验证码短信
	SendMemberCodeSMS(phone, language string, params *MemberSendCodeRequest) (*SMSResponse, error)
	// SendMemberPointsSMS 发送会员积分短信
	SendMemberPointsSMS(phone, language string, params *MemberPointsRequest) (*SMSResponse, error)
	// SendMemberCouponSMS 发送会员优惠券短信
	SendMemberCouponSMS(phone, language string, params *MemberCouponRequest) (*SMSResponse, error)
	// SendDeliveryOrderBySelfCancelSMS 发送外送订单取消短信
	SendDeliveryOrderBySelfCancelSMS(phone, language string, params *DeliveryOrderCancelBySelfRequest) (*SMSResponse, error)
	// CheckConfig 检查客户端配置是否正确
	CheckConfig() error
}
