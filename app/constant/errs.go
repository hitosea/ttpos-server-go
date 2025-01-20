package constant

const (
	CodeSuccess = 0   // 成功
	CodeFail    = 500 // 失败

	CodeBadRequest            = 400 // 参数错误
	CodeUnauthorized          = 401 // 登录失败
	CodeAccountDeleted        = 402 // 账号被删除
	CodeUnhandShiftUserExists = 403 // 登录当前设备有未交班用户

	CodeBindLimit = 404 // 登录判断设备绑定超过上限
)
