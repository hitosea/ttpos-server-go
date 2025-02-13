package constant

// TODO 部分错误码直接拿重构前的，是否需要更换
const (
	CodeSuccess    = 1   // 成功
	CodeFail       = 500 // 失败
	CodeForceFail  = 501 // 强调失败
	CodeDialogFail = 502 // 模态框异常

	CodeBadRequest   = 400 // 参数错误
	CodeUnauthorized = 401 // 登录失败
	CodeBindLimit    = 404 // 登录判断设备绑定超过上限

	CodeUnbindError = -201 // 设备已解绑，请重新绑定
	CodeTableError  = -5   // 桌台用餐已关闭, 对应业务被关闭，需要自行处理

	TokenError = -2 // TOKEN异常

	TokenErrorNotLogin = -202 // token异常，不需要重新登录，定格当前页面
)
