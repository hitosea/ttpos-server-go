package constant

const (
	ServerModeRelease = "release"
	ServerModeDebug   = "debug"
	ServerModeTest    = "test"
	ServerModeStop    = "stop" // 停止服务, 不用关闭某些特性功能
)

const (
	//
	CodeSuccessOpenCashBox = 200 // 成功-并打开钱箱
	CodeSuccess            = 0   // 成功

	CodeFail                = -1   // 失败
	CodeSystemError         = -2   // 系统错误
	CodeParamError          = -3   // 参数错误
	CodeNetworkError        = -4   // 网络错误
	CodeTimeoutError        = -5   // 超时错误
	CodeVersionError        = -6   // 版本错误
	CodeErrorConfirmClose   = -11  // 需要弹窗提示，点击确定 关闭弹窗
	CodeErrorConfirmRequest = -12  // 需要弹窗提示，点击确定之后传 is_confirm 为 true 到当前接口再此请求接口
	CodeErrorConfirmRefresh = -13  // 需要弹窗提示，点击确定后刷新页面
	UnknownError            = -99  // 未知错误
	CodeUnauthorized        = -100 // 未授权
	CodeTokenExpired        = -101 // Token过期
	CodeTokenInvalid        = -102 // Token无效
	CodeAccessDenied        = -103 // 拒绝访问
	CodeAccountDisabled     = -104 // 账号已禁用
	CodeNeedLogin           = -105 // 需要登录
	CodeLoginFailed         = -106 // 登录失败
	CodeCashierNotLogin     = -107 // 点餐助手绑定的收银机未登录
	CodeCashierHandedOver   = -108 // 已交班
	CodeNeedCompanyUuid     = -109 // 需要绑定门店UUID

	CodeCashierOrderMethodNotOpen = -110 //
	CodeCashierLoginLimit         = -111 // 收银机登录已达上限
	CodeAssistantLoginLimit       = -112 // 点餐助手登录已达上限
	CodeKitchenLoginLimit         = -113 // 厨显登录已达上限
	CodeTabletLoginLimit          = -114 // 平板登录已达上限

	CodeFunctionDisabled = -119 // 助手端、平板端、厨显端、会员、自助餐功能

	CodeCompanyLicenceExpired = -191 // 商家过期
)

// 送厨检查的业务错误码
const (
	CodeOrderCheckProductDown         = -200 // 商品已下架
	CodeOrderCheckProductFlavorDown   = -201 // 商品某规格已下架
	CodeOrderCheckProductStockZero    = -202 // 商品库存不足
	CodeOrderCheckProductMust         = -203 // 未选择必点商品
	CodeOrderCheckProductPriceChanged = -204 // 商品价格已变动
	CodeOrderCheckProductLimitOut     = -205 // 商品超出限购
	CodeOrderCheckProductSauceDown    = -208 // 商品某小料已下架
	// 结账检查
	CodeOrderCheckProductUnCooking = -206 // 有商品未送厨
	// 关闭桌台检查
	CodeOrderCheckProductCooking = -207 // 有商品已送厨
	CodeOrderCheckProductBatch   = -209 // 有商品是分批商品，请前去送厨。弹出分批送厨弹窗
	CodeOrderCheckSplit          = -210 // 订单已经拆单，请前去收银机操作

	CodeH5OrderTimeLimit = -231 // h5订单时间限制
	CodeH5OrderNumLimit  = -232 // h5订单数量限制

	// 优惠券已经失效
	CodeCouponInvalid = -233
)

// 桌台业务错误码
const (
	CodeDeskOrderEnd      = -300 // 桌台订单已经结束。前端收到这个业务码后，退出当前桌台，返回首页
	CodeTabletNotBindDesk = -301 // 平板未绑定桌台
)

// 退款业务错误码
const (
	CodeReturnOrderBank = -401 // 请选择银行
)

// 营销活动业务错误码
const (
	CodeMarketingActivityInvalid = -501 // 营销活动已失效
)

// 支付业务错误码
const (
	CodeOrderAmountLessThan1 = -601 // 订单金额小于1，无法支付
	CodeOrderPayError        = -602 // 支付服务异常
)

// 外送业务错误码
const (
	CodeDistanceError                  = -701 // 计算距离错误
	CodeTakeoutCreateOrderError        = -702 // 外送订单创建失败
	CodeOrderAddressNotVerifiedPhone   = -703 // 订单地址未认证手机号
	CodeOrderAddressNotInDeliveryRange = -704 // 订单地址不在配送范围内
	CodeOrderPrepareDataError          = -705 // 准备数据发起外送订单创建时失败
)

// 商家端业务错误码
const (
	CodeProductDeleteCanNotDeletePackage = -801 // 商品已关联套餐，无法删除，请先修改套餐
	CodeProductEditCanNotDeletePackage   = -802 // 商品规格已关联套餐，无法删除，请先修改套餐
	CodeMaterialDisabled                 = -803 // 物品已停用，您可启用物品后再进行收货
	CodeWarehouseStockNotEnough          = -804 // 物品库存不足
)

// 采购订单业务错误码
const (
	CodePurchaseOrderSupplierDisabled = -901 // 供应商已禁用
)

// 调拨订单业务错误码
const (
	CodeTransferOrderItemUnitNumZero = -805 // 调拨单明细单位数量为0，提交后将移除对应物品
)

// 盘点单业务错误码
const (
	CodeWarehouseDisabled = -806 // 仓库状态已关闭，请修改仓库状态
	CodeItemDisabled      = -807 // 物品状态已关闭，请修改物品状态
)

type ParseCodeOrderCheckOption struct {
	IsH5     bool // 是否是h5端的文案
	IsTablet bool // 是否是平板端的文案
}

func WithIsH5() func(option *ParseCodeOrderCheckOption) {
	return func(option *ParseCodeOrderCheckOption) {
		option.IsH5 = true
	}
}

func WithIsTablet() func(option *ParseCodeOrderCheckOption) {
	return func(option *ParseCodeOrderCheckOption) {
		option.IsTablet = true
	}
}

func ParseCodeOrderCheck(code int, options ...func(option *ParseCodeOrderCheckOption)) string {
	option := &ParseCodeOrderCheckOption{}
	for _, opt := range options {
		opt(option)
	}
	switch code {
	case CodeOrderCheckProductDown:
		return "商品已下架"
	case CodeOrderCheckProductFlavorDown:
		return "规格已下架,请选择其他规格"
	case CodeOrderCheckProductStockZero:
		return "以下商品库存不足，请删除后再下单"
	case CodeOrderCheckProductMust:
		tips := "已下单和本次要下单的商品未选择必点商品，确定要继续下单吗？"
		if option.IsH5 {
			return tips
		}
		if option.IsTablet {
			return tips
		}
		return "已送厨和本次要送厨的商品未选择必点商品，确定要继续送厨吗？"
	case CodeOrderCheckProductPriceChanged:
		return "订单商品数据有变动，请重新查看订单"
	case CodeOrderCheckProductLimitOut:
		return "以下商品超出限购数量，请在限购数量内下单"
	default:
		return "success"
	}
}
