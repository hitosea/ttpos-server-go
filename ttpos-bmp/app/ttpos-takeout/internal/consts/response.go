package consts

// ResponseCode gRPC 响应码常量
// 用于 takeout.ApiResponse 的 Code 字段
type ResponseCode string

const (
	// CodeSuccess 成功
	CodeSuccess ResponseCode = "0"

	// CodeInvalidParam 参数校验失败 (4xxx 系列)
	CodeInvalidParam ResponseCode = "4001"

	// CodeServiceError 服务内部错误 (5xxx 系列)
	CodeServiceError ResponseCode = "5001"

	// CodeSerializeError 序列化错误
	CodeSerializeError ResponseCode = "5002"

	// CodeExternalAPIError 外部 API 调用错误
	CodeExternalAPIError ResponseCode = "5003"
)

// ResponseMessage 响应消息常量
const (
	MsgSuccess           = "success"
	MsgSerializeFailed   = "序列化响应数据失败"
	MsgMerchantIDEmpty   = "merchant_id 不能为空"
	MsgItemIDEmpty       = "item_id 不能为空"
	MsgModifierIDEmpty   = "modifier_id 不能为空"
	MsgModifierNameEmpty = "modifier_name 不能为空"
)
