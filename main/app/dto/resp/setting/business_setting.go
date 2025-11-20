package setting

import (
	"ttpos-server-go/app/constant"

	"github.com/shopspring/decimal"
)

// Business 门店业务设置
type Business struct {
	ZeroingMethodList         []ZeroingMethodItem         `json:"zeroing_method_list"`          // 优惠折扣自动抹零方式列表
	CheckoutZeroingMethodList []CheckoutZeroingMethodItem `json:"checkout_zeroing_method_list"` // 结账自动抹零方式列表
	ZeroingMethod             string                      `json:"zeroing_method"`               // 优惠折扣自动抹零方式: 0-实款实收 1-抹分 2-抹角 3-四舍五入保留一位小数 4-四舍五入到整数
	CheckoutZeroingMethod     string                      `json:"checkout_zeroing_method"`      // 结账自动抹零方式: 0-实款实收 1-抹分 2-抹角 5-抹元. // 5-抹元为5是为了全局唯一，各个数字有不重复的抹零定义
	GiftMethodList            []GiftMethodItem            `json:"gift_method_list"`             // 赠菜计算方式列表
	GiftMethod                string                      `json:"gift_method"`                  // 赠菜计算方式
	FreeMethodList            []FreeMethodItem            `json:"free_method_list"`             // 免单计算方式列表
	FreeMethod                string                      `json:"free_method"`                  // 免单计算方式
	DiscountMethod            string                      `json:"discount_method"`              // 折扣计算方式 10-按百分比 20-直接减免
	QrCode                    string                      `json:"qr_code"`                      // 电子菜单二维码校验失效值，6位数数字
	NoClearTable              string                      `json:"no_clear_table"`               // 结账后不清台 0-清台 1-不清台
	IsNeedPassword            string                      `json:"is_need_password"`             // 取消订单/退菜 0-无需密码 1-需要密码
	DishCardStyle             string                      `json:"dish_card_style"`              // 菜品卡片样式 0-无图模式 1-图片模式
	DishCardStyleTime         string                      `json:"dish_card_style_time"`         // 菜品卡片样式最后更新时间
	IsInvoice                 string                      `json:"is_invoice"`                   // 开票信息 0-不需要填写 1-需要填写
	OpeningHours              string                      `json:"opening_hours"`                // 营业时间 18:00-02:00
	DeliveryPriceRatio        uint                        `json:"delivery_price_ratio"`         // 外送商品价格和商品原价比例. 取值范围1-300， 表示原价的1%到300%
	StartSerialNo             string                      `json:"start_serial_no"`              // 开始序列号
	// 分批商品相关
	IsBatch           string   `json:"is_batch"`            // 是否是分批商品 0-否 1-是
	BatchProductUuids []uint64 `json:"batch_product_uuids"` // 分批商品UUID列表
	BatchTagNum       uint     `json:"batch_tag_num"`       // 分批类型数量
	BatchCookingMode  string   `json:"batch_cooking_mode"`  // 分批送厨模式: "pre" 前置 / "post" 后置，默认 "post"
	SafetyStockType   string   `json:"safety_stock_type"`   // 安全库存类型 1-门店维度 2-仓库维度，默认为1

	RequiredParentCompanyApproval string `json:"required_parent_company_approval"` // 调拨规则-经过上级门店审批 "0"-否 "1"-是，总部和上级支持此选项
	ViaParentCompanyWarehouse     string `json:"via_parent_company_warehouse"`     // 调拨规则-经过上级门店仓库 "0"-否 "1"-是，总部和上级支持此选项

	// 敏感操作设置
	DiscountNeedPassword       string   `json:"discount_need_password"`       // 折扣操作是否需要密码 0-否 1-是
	DiscountAuthorizedStaffIds []uint64 `json:"discount_authorized_staff_ids"` // 折扣操作授权员工ID列表
	RefundNeedPassword         string   `json:"refund_need_password"`          // 退款操作是否需要密码 0-否 1-是
	RefundAuthorizedStaffIds   []uint64 `json:"refund_authorized_staff_ids"`  // 退款操作授权员工ID列表
}

type ShopBusiness struct {
	Business
	FreeReasonCount       int `json:"free_reason_count"`        // 免单原因数量
	ReturnFoodReasonCount int `json:"return_food_reason_count"` // 退菜原因数量
	OrderRemarkCount      int `json:"order_remark_count"`       // 整单备注原因数量

	HeadquarterRequiredParentCompanyApproval string `json:"headquarter_required_parent_company_approval"` // 总部调拨规则-经过上级门店审批 "0"-否 "1"-是
	HeadquarterViaParentCompanyWarehouse     string `json:"headquarter_via_parent_company_warehouse"`     // 总部调拨规则-经过上级门店仓库 "0"-否 "1"-是
}

// 是否开启了分批送厨商品
func (resp *Business) OpenIsBatch() bool {
	return resp.IsBatch == "1"
}

func (resp *Business) IsAutoClearDesk() bool {
	return resp.NoClearTable == constant.AutoClearTable
}

// 外送折扣率。100%应表示为1，90%应表示为0.9
func (resp *Business) GetDeliveryPriceRatio() float64 {
	value := float64(resp.DeliveryPriceRatio)
	return decimal.NewFromFloat(value).Div(decimal.NewFromInt(100)).Round(4).InexactFloat64() // 取值范围为0.01-3
}

// 是否需要经过上级门店审批
func (resp *Business) IsRequiredParentCompanyApproval() bool {
	return resp.RequiredParentCompanyApproval == "1"
}

// 是否经过上级门店仓库
func (resp *Business) IsViaParentCompanyWarehouse() bool {
	return resp.ViaParentCompanyWarehouse == "1"
}

type ZeroingMethodItem MethodItem

type CheckoutZeroingMethodItem MethodItem

type GiftMethodItem MethodItem

type FreeMethodItem MethodItem

type MethodItem struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}
