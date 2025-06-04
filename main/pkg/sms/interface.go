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
	// CheckConfig 检查客户端配置是否正确
	CheckConfig() error
}
