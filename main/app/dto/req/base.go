package req

type VerifyPasswordReq struct {
	Password string `json:"password" binding:"required"` // 密码
}

type UpdateAcceptOrderSetting struct {
	IsAutoOrder    string `json:"is_auto_order" binding:"required,oneof=0 1"`  // 是否自动接单：0-否；1-是
	AutoOrderLimit string `json:"auto_order_limit" binding:"auto_order_limit"` // 自动接单金额上限，0.01-100000000
}

var UpdateAcceptOrderSettingMessage = map[string]string{
	"is_auto_order.required":            "是否自动接单参数错误",
	"is_auto_order.oneof":               "是否自动接单参数错误",
	"auto_order_limit.auto_order_limit": "自动接单金额上限，0.01-100000000",
}

type UpdateSystemSetting struct {
	IsShowAssistantSoldOut *int   `json:"is_show_assistant_sold_out" binding:"required,oneof=0 1"` // 助手端点餐助手是否显示售罄商品 0-不显示 1-显示
	IsShowSoldOut          *int   `json:"is_show_sold_out" binding:"required,oneof=0 1"`           // 平板端是否显示售罄商品 0-关闭（不显示售罄） 1-开启（显示售罄）
	IsShowScanSoldOut      *int   `json:"is_show_scan_sold_out" binding:"required,oneof=0 1"`      // 扫码点餐端是否显示售罄商品 0-关闭（不显示售罄） 1-开启（显示售罄）
	MenuShowSoldOut        *int   `json:"menu_show_sold_out" binding:"required,oneof=0 1"`         // 电子菜单是否显示售罄商品 0-关闭（不显示售罄） 1-开启（显示售罄）
	DishCardStyle          string `json:"dish_card_style" binding:"required,oneof=0 1"`            // 菜品卡片样式 0-无图模式 1-图片模式
	DeviceRemark           string `json:"device_remark"`                                           // 机器备注
}

type PaymentMethodListReq struct {
	Type string `form:"type,default=all" binding:"omitempty,oneof=all checkout recharge" ` // 支付方式列表
}

type KitchenBindReq struct {
	Brand              string `json:"brand"`                // 品牌
	ProductPrinterUuid uint64 `json:"product_printer_uuid"` // 商品打印Uuid
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
	WithdrawCash float64 `json:"withdraw_cash"` // 取出金额, 最多小数点后两位
}

// ShiftDepositReq 交班存钱
type ShiftDepositReq struct {
	DepositCash float64 `json:"deposit_cash"` // 存入金额, 最多小数点后两位
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
