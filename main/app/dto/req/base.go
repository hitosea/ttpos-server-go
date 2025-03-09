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
	IsShowScanSoldOut      *int   `json:"is_show_scan_sold_out" binding:"required,oneof=0 1"`      // 扫码点餐端是否显示售罄商品 0-关闭（不显示售罄） 1-开启（显示售罄）
	IsShowAssistantSoldOut *int   `json:"is_show_assistant_sold_out" binding:"required,oneof=0 1"` // 助手端点餐助手是否显示售罄商品 0-不显示 1-显示
	MenuShowSoldOut        *int   `json:"menu_show_sold_out" binding:"required,oneof=0 1"`         // 电子菜单是否显示售罄商品 0-关闭（不显示售罄） 1-开启（显示售罄）
	DishCardStyle          string `json:"dish_card_style" binding:"required,oneof=0 1"`            // 菜品卡片样式 0-无图模式 1-图片模式
	IsShowSoldOut          *int   `json:"is_show_sold_out" binding:"required,oneof=0 1"`           // 平板端是否显示售罄商品 0-关闭（不显示售罄） 1-开启（显示售罄）
}

type PaymentMethodListReq struct {
	Type string `form:"type,default=all" binding:"omitempty,oneof=all checkout recharge" ` // 支付方式列表
}

type KitchenBindReq struct {
	Brand              string `json:"brand"`                // 品牌
	ProductPrinterUuid uint64 `json:"product_printer_uuid"` // 商品打印Uuid
	Remark             string `json:"remark"`               // 备注
}
