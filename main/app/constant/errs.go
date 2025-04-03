package constant

const (
	ServerModeRelease = "release"
	ServerModeDebug   = "debug"
	ServerModeTest    = "test"
)

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
	CodeCashierNotLogin = -107 // 点餐助手绑定的收银机未登录

	CashierLoginLimit   = -111 // 收银机登录已达上限
	AssistantLoginLimit = -112 // 点餐助手登录已达上限
	KitchenLoginLimit   = -113 // 厨显登录已达上限
	TabletLoginLimit    = -114 // 平板登录已达上限

	CompanyLicenceExpired = -191 // 商家过期
)

// 送厨检查的业务错误码
const (
	CodeOrderCheckProductDown         = -200 // 商品已下架
	CodeOrderCheckProductFlavorDown   = -201 // 商品某规格已下架
	CodeOrderCheckProductStockZero    = -202 // 商品库存不足
	CodeOrderCheckProductMust         = -203 // 未选择必点商品
	CodeOrderCheckProductPriceChanged = -204 // 商品价格已变动
	CodeOrderCheckProductLimitOut     = -205 // 商品超出限购
	// 结账检查
	CodeOrderCheckProductUnCooking = -206 // 有商品未送厨
	// 关闭桌台检查
	CodeOrderCheckProductCooking = -207 // 有商品已送厨
)

func ParseCodeOrderCheck(code int) string {
	switch code {
	case CodeOrderCheckProductDown:
		return "商品已下架"
	case CodeOrderCheckProductFlavorDown:
		return "规格已下架,请选择其他规格"
	case CodeOrderCheckProductStockZero:
		return "以下商品库存不足，请删除后再下单"
	case CodeOrderCheckProductMust:
		return "已送厨和本次要送厨的商品未选择必点商品，确定要继续送厨吗？" // todo 扫码端要这样显示？已下单和本次要下单的商品未选择必点商品，确定要继续下单吗？
	case CodeOrderCheckProductPriceChanged:
		return "订单商品数据有变动，请重新查看订单"
	case CodeOrderCheckProductLimitOut:
		return "超过限购数量"
	default:
		return "success"
	}
}

// 桌台业务错误码
const (
	CodeDeskOrderEnd = -300 // 桌台订单已经结束。前端收到这个业务码后，退出当前桌台，返回首页
)

// 退款业务错误码
const (
	CodeReturnOrderBank = -401 // 请选择银行
)
