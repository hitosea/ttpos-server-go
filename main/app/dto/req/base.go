package req

import "ttpos-server-go/app/dto"

type VerifyPasswordReq struct {
	Password string `json:"password" binding:"required"` // 密码
}

type UpdateAcceptOrderSetting struct {
	IsAutoOrder    string `json:"is_auto_order" binding:"required,oneof=0 1"`  // 是否自动接单：0-否；1-是
	AutoOrderLimit string `json:"auto_order_limit" binding:"auto_order_limit"` // 自动接单金额上限，0.01-100000000
}

type UpdateAcceptMemberOrderSetting struct {
	IsAutoMemberOrder    string `json:"is_auto_member_order" binding:"required,oneof=0 1"`  // 是否自动接单会员订单：0-否；1-是
	AutoMemberOrderLimit string `json:"auto_member_order_limit" binding:"auto_order_limit"` // 自动接单会员订单金额上限，0.01-100000000
}

var UpdateAcceptOrderSettingMessage = map[string]string{
	"is_auto_order.required":            "是否自动接单参数错误",
	"is_auto_order.oneof":               "是否自动接单参数错误",
	"auto_order_limit.auto_order_limit": "自动接单金额上限，0.01-100000000",
}

var UpdateAcceptMemberOrderSettingMessage = map[string]string{
	"is_auto_member_order.required":            "是否自动接单外送订单参数错误",
	"is_auto_member_order.oneof":               "是否自动接单外送订单参数错误",
	"auto_member_order_limit.auto_order_limit": "自动接单外送订单金额上限，0.01-100000000",
}

type UpdateSystemSetting struct {
	IsShowAssistantSoldOut *int   `json:"is_show_assistant_sold_out" binding:"required,oneof=0 1"` // 助手端点餐助手是否显示售罄商品 0-不显示 1-显示
	IsShowSoldOut          *int   `json:"is_show_sold_out" binding:"required,oneof=0 1"`           // 平板端是否显示售罄商品 0-关闭（不显示售罄） 1-开启（显示售罄）
	IsShowScanSoldOut      *int   `json:"is_show_scan_sold_out" binding:"required,oneof=0 1"`      // 扫码点餐端是否显示售罄商品 0-关闭（不显示售罄） 1-开启（显示售罄）
	MenuShowSoldOut        *int   `json:"menu_show_sold_out" binding:"required,oneof=0 1"`         // 电子菜单是否显示售罄商品 0-关闭（不显示售罄） 1-开启（显示售罄）
	MemberShowSoldOut      *int   `json:"member_show_sold_out" binding:"required,oneof=0 1"`       // 会员端是否显示售罄商品 0-关闭（不显示售罄） 1-开启（显示售罄）
	DishCardStyle          string `json:"dish_card_style" binding:"required,oneof=0 1"`            // 菜品卡片样式 0-无图模式 1-图片模式
	DeviceRemark           string `json:"device_remark"`                                           // 机器备注
}

type PaymentMethodListReq struct {
	Type string `form:"type,default=all" binding:"omitempty,oneof=all checkout recharge" ` // 支付方式列表
}

type KitchenBindReq struct {
	Brand              string `json:"brand"`                // 品牌
	ProductPrinterUuid uint64 `json:"product_printer_uuid"` // 商品打印Uuid
	Mode               *uint  `json:"mode"`                 // 模式 0-默认，传菜模式; 1-制作模式; 2-制作+传菜模式
	Remark             string `json:"remark"`               // 备注
}

// SubmitShiftReq 提交交班
type SubmitShiftReq struct {
	WithdrawCash float64 `json:"withdraw_cash"` // 取出金额: 0 - 当前钱箱现金总计
	LeaveCash    float64 `json:"leave_cash"`    // 遗留现金: 0 - 当前钱箱现金总计
	IsBackground bool    `json:"is_background"` // 是否后台交班: false-否，true-是
	StaffUuid    uint64  `json:"staff_uuid"`    // 员工uuid: 后台交班时，传入员工uuid
}

type CashierReportReq struct {
	ExceptionRemark string `json:"exception_remark"` // 异常备注
}

// ShiftWithdrawReq 交班取钱
type ShiftWithdrawReq struct {
	WithdrawCash string `json:"withdraw_cash"` // 取出金额, 最多小数点后两位
}

// ShiftDepositReq 交班存钱
type ShiftDepositReq struct {
	DepositCash string `json:"deposit_cash"` // 存入金额, 最多小数点后两位
}

type ShiftPrinterReq struct {
	WithdrawCash float64 `json:"withdraw_cash"` // 取出金额: 0 - 当前钱箱现金总计
	LeaveCash    float64 `json:"leave_cash"`    // 遗留现金: 0 - 当前钱箱现金总计
	DutyNo       string  `json:"duty_no"`       // 班次编号 (交班时，返回的班次编号)
}

// EditDeviceRemarkReq 助手端修改设置请求参数
type EditDeviceRemarkReq struct {
	Remark string `json:"remark"` // 机器备注
}

// SendMemberRechargeSMS
type SendMemberRechargeSMS struct {
	Company       string  `json:"company"`
	Phone         string  `json:"phone"`
	Recharge      float64 `json:"recharge"`
	BonusMoney    float64 `json:"bonus_money"`
	BonusPoints   float64 `json:"bonus_points"`
	Balance       float64 `json:"balance"`
	PointsBalance float64 `json:"points_balance"`
}

type UpdateStoreSetting struct {
	Name        string             `json:"name" binding:"required,max=100"`   // 店铺名称
	LogoUrl     string             `json:"logo_url" binding:"required"`       // 店铺logo，上传后保存url，必填
	TimeZone    string             `json:"time_zone" binding:"required"`      // 时区，必填
	CompanyName string             `json:"company_name" binding:"max=500"`    // 公司名称，区别于店铺名称，最大500个字符
	Address     string             `json:"address" binding:"max=500"`         // 地址，必填，最大500个字符
	Phone       string             `json:"phone" binding:"required,max=20"`   // 联系电话，必填，最大20个字符
	TaxNumber   string             `json:"tax_number"`                        // 税号
	Language    []dto.LanguageItem `json:"language" binding:"required,min=1"` // 系统语言，必填，至少一个
	Coordinates string             `json:"coordinates"`                       // 经纬度
}

type UpdateBusinessSetting struct {
	ZeroingMethod         string `json:"zeroing_method" binding:"required,oneof=0 1 2 3 4"`        // 优惠折扣自动抹零方式: 0-实款实收 1-抹分 2-抹角 3-四舍五入保留一位小数 4-四舍五入到整数
	CheckoutZeroingMethod string `json:"checkout_zeroing_method" binding:"required,oneof=0 1 2 5"` // 结账自动抹零方式: 0-实款实收 1-抹分 2-抹角 5-抹元. // 5-抹元为5是为了全局唯一，各个数字有不重复的抹零定义
	GiftMethod            string `json:"gift_method" binding:"required,oneof=10 20"`               // 赠菜计算方式: 10-计入总销售额、优惠折扣 20-不计入总销售额、优惠折扣
	FreeMethod            string `json:"free_method" binding:"required,oneof=10 20"`               // 免单计算方式: 10-计入总销售额、优惠折扣、服务费、税费 20-不计入总销售额、优惠折扣、服务费、税费
	IsInvoice             string `json:"is_invoice" binding:"required,oneof=0 1"`                  // 开票信息: 0-不需要填写 1-需要填写
	DiscountMethod        string `json:"discount_method" binding:"required,oneof=10 20"`           // 折扣计算方式: 10-按百分比 20-直接减免
	NoClearTable          string `json:"no_clear_table" binding:"required,oneof=0 1"`              // 结账后不清台: 0-清台 1-不清台
	IsNeedPassword        string `json:"is_need_password" binding:"required,oneof=0 1"`            // 取消订单/退菜 0-无需密码 1-需要密码
	DishCardStyle         string `json:"dish_card_style" binding:"required,oneof=0 1"`             // 菜品卡片样式 0-无图模式 1-图片模式
	OpeningHours          string `json:"opening_hours" binding:"required,opening_hours"`           // 营业时间格式正则验证：HH:MM-HH:MM
	DeliveryPriceRatio    uint   `json:"delivery_price_ratio" binding:"required,gte=1,lte=300"`    // 外送商品价格和商品原价比例. 取值范围1-300， 表示原价的1%到300%
	StartSerialNo         string `json:"start_serial_no" binding:"required"`                       // 开始序列号
	IsBatch               string `json:"is_batch" binding:"required,oneof=0 1"`                    // 是否是分批商品 0-否 1-是
	BatchCookingMode      string `json:"batch_cooking_mode" binding:"required,oneof=pre post"`     // 分批模式 pre-前置模式 post-后置模式，默认为post
	SafetyStockType       string `json:"safety_stock_type" binding:"required,oneof=1 2"`           // 安全库存类型 1-门店纬度 2-仓库纬度，默认为1

	// 调拨规则
	RequiredParentCompanyApproval string `json:"required_parent_company_approval" binding:"omitempty,oneof=0 1"` // 调拨规则-经过上级门店审批 "0"-否 "1"-是，总部和上级支持此选项
	ViaParentCompanyWarehouse     string `json:"via_parent_company_warehouse" binding:"omitempty,oneof=0 1"`     // 调拨规则-经过上级门店仓库 "0"-否 "1"-是，总部和上级支持此选项
}
