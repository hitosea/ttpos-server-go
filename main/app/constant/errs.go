package constant

const (
	CodeSuccess   = 1   // 成功
	CodeFail      = 500 // 失败
	CodeForceFail = 501 // 强调失败

	CodeBadRequest   = 400 // 参数错误
	CodeUnauthorized = 401 // 登录失败
	CodeBindLimit    = 404 // 登录判断设备绑定超过上限
)
