// Package consts 定义消息中心服务的常量
package consts

// 消息类型常量
const (
	MessageTypeEmail = "email" // 邮件消息
	MessageTypeSMS   = "sms"   // 短信消息
)

// 消息状态常量
const (
	MessageStatusPending = 0 // 待发送
	MessageStatusSending = 1 // 发送中
	MessageStatusSuccess = 2 // 发送成功
	MessageStatusFailed  = 3 // 发送失败
)

// 模板状态常量
const (
	TemplateStatusDisabled = 0 // 禁用
	TemplateStatusEnabled  = 1 // 启用
)

// 发送结果常量
const (
	SendResultFailed  = 0 // 发送失败
	SendResultSuccess = 1 // 发送成功
)

// RocketMQ 配置常量
const (
	TopicMessageSend = "ttpos-message-send" // 消息发送主题
	TagEmail         = "email"              // 邮件标签
	TagSMS           = "sms"                // 短信标签
)

// 错误消息常量
const (
	ErrMsgInvalidParameter   = "参数验证失败"
	ErrMsgMessageNotFound    = "消息不存在"
	ErrMsgTemplateNotFound   = "模板不存在"
	ErrMsgTemplateDisabled   = "模板已禁用"
	ErrMsgDuplicateMessage   = "消息已存在"
	ErrMsgSendFailed         = "消息发送失败"
	ErrMsgInvalidStatus      = "消息状态不允许该操作"
	ErrMsgMailgunError       = "Mailgun 服务错误"
	ErrMsgRocketMQError      = "RocketMQ 队列错误"
	ErrMsgTemplateRender     = "模板渲染失败"
	ErrMsgInvalidRecipient   = "接收人格式不正确"
	ErrMsgInvalidMessageType = "不支持的消息类型"
)

// 默认配置常量
const (
	DefaultRetryCount     = 3   // 默认重试次数
	DefaultTimeout        = 30  // 默认超时时间（秒）
	DefaultPoolSize       = 20  // 默认线程池大小
	MaxRetryCount         = 5   // 最大重试次数
	MaxSubjectLength      = 200 // 最大主题长度
	MaxRecipientLength    = 100 // 最大接收人长度
	MaxErrorMessageLength = 500 // 最大错误信息长度
)
