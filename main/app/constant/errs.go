package constant

const (
	ServerModeRelease = "release"
	ServerModeDebug   = "debug"
	ServerModeTest    = "test"
)

// TODO 部分错误码直接拿重构前的，是否需要更换
const (
	CodeSuccess = 0 // 成功

	CodeFail         = -1  // 失败
	CodeSystemError  = -2  // 系统错误
	CodeParamError   = -3  // 参数错误
	CodeNetworkError = -4  // 网络错误
	CodeTimeoutError = -5  // 超时错误
	UnknownError     = -99 // 未知错误

	CodeUnauthorized    = -100 // 未授权
	CodeTokenExpired    = -101 // Token过期
	CodeTokenInvalid    = -102 // Token无效
	CodeAccessDenied    = -103 // 拒绝访问
	CodeAccountDisabled = -104 // 账号已禁用
	CodeNeedLogin       = -105 // 需要登录
	CodeLoginFailed     = -106 // 登录失败
)
