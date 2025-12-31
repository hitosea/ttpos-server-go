package entity

import (
	"strconv"
	"time"
)

// BusinessSetting 门店业务设置实体（字段顺序与旧服务 Business 保持一致）
type BusinessSetting struct {
	// 以下字段顺序严格按照旧服务 setting.Business 的顺序定义，确保 JSON 输出一致
	ZeroingMethodList             []ZeroingMethodItem         `json:"zeroing_method_list"`              // 优惠折扣自动抹零方式列表
	CheckoutZeroingMethodList     []CheckoutZeroingMethodItem `json:"checkout_zeroing_method_list"`     // 结账自动抹零方式列表
	ZeroingMethod                 string                      `json:"zeroing_method"`                   // 抹零方式: 0-不抹零 1-抹分 2-抹角 3-四舍五入到角 4-四舍五入到元
	CheckoutZeroingMethod         string                      `json:"checkout_zeroing_method"`          // 结账抹零方式
	GiftMethodList                []GiftMethodItem            `json:"gift_method_list"`                 // 赠送方式列表
	GiftMethod                    string                      `json:"gift_method"`                      // 赠送方式 0-不赠送 1-按商品 2-按金额
	FreeMethodList                []FreeMethodItem            `json:"free_method_list"`                 // 免单方式列表
	FreeMethod                    string                      `json:"free_method"`                      // 免单方式 0-不免单 1-按商品 2-按金额
	DiscountMethod                string                      `json:"discount_method"`                  // 折扣计算方式 10-按百分比 20-直接减免
	QrCode                        string                      `json:"qr_code"`                          // 二维码
	NoClearTable                  string                      `json:"no_clear_table"`                   // 结账后不清台 0-清台 1-不清台
	IsNeedPassword                string                      `json:"is_need_password"`                 // 取消订单/退菜 0-无需密码 1-需要密码
	DishCardStyle                 string                      `json:"dish_card_style"`                  // 菜品卡片样式 0-默认 1-样式一
	DishCardStyleTime             string                      `json:"dish_card_style_time"`             // 菜品卡片样式最后更新时间
	IsInvoice                     string                      `json:"is_invoice"`                       // 开票信息 0-不需要填写 1-需要填写
	OpeningHours                  string                      `json:"opening_hours"`                    // 营业时间 00:00-23:59
	DeliveryPriceRatio            uint                        `json:"delivery_price_ratio"`             // 外送加价比例
	StartSerialNo                 string                      `json:"start_serial_no"`                  // 起始流水号
	IsBatch                       string                      `json:"is_batch"`                         // 是否开启分批送厨商品 0-关闭 1-开启
	BatchProductUuids             []uint64                    `json:"batch_product_uuids"`              // 分批商品UUID列表
	BatchTagNum                   uint                        `json:"batch_tag_num"`                    // 分批类型数量
	BatchCookingMode              string                      `json:"batch_cooking_mode"`               // 分批烹饪模式 post/pre
	SafetyStockType               string                      `json:"safety_stock_type"`                // 安全库存类型
	RequiredParentCompanyApproval string                      `json:"required_parent_company_approval"` // 需要总部审批
	ViaParentCompanyWarehouse     string                      `json:"via_parent_company_warehouse"`     // 通过总部仓库
	DiscountNeedPassword          string                      `json:"discount_need_password"`           // 折扣操作是否需要密码 0-否 1-是
	DiscountAuthorizedStaffIds    []uint64                    `json:"discount_authorized_staff_ids"`    // 折扣授权员工ID
	RefundNeedPassword            string                      `json:"refund_need_password"`             // 退款操作是否需要密码 0-否 1-是
	RefundAuthorizedStaffIds      []uint64                    `json:"refund_authorized_staff_ids"`      // 退款授权员工ID
	EnableOrderSource             string                      `json:"enable_order_source"`              // 外卖功能开关 0-关闭 1-开启
	EnableNationality             string                      `json:"enable_nationality"`               // 国籍功能开关 0-关闭 1-开启

	// 以下字段在旧服务 Business 中不存在（新增字段），可以放在最后
	IsAutoOrder            string            `json:"is_auto_order,omitempty"`              // 是否自动接单 0-关闭 1-开启
	AutoOrderLimit         string            `json:"auto_order_limit,omitempty"`           // 自动接单金额上限
	IsAutoVoice            string            `json:"is_auto_voice,omitempty"`              // 是否开启自动接单语音播报
	IsAutoMemberOrder      string            `json:"is_auto_member_order,omitempty"`       // 是否自动接单会员订单 0-关闭 1-开启
	AutoMemberOrderLimit   string            `json:"auto_member_order_limit,omitempty"`    // 自动接单会员订单金额上限
	IsAutoVoiceMemberOrder string            `json:"is_auto_voice_member_order,omitempty"` // 是否开启自动接单会员订单语音播报
	IsDiscountAuth         string            `json:"is_discount_auth,omitempty"`           // 是否开启折扣授权 0-关闭 1-开启
	IsRefundAuth           string            `json:"is_refund_auth,omitempty"`             // 是否开启退款授权 0-关闭 1-开启
	IsTakeout              string            `json:"is_takeout,omitempty"`                 // 是否开启外送 0-关闭 1-开启
	IsScanOrder            string            `json:"is_scan_order,omitempty"`              // 是否开启扫码点餐 0-关闭 1-开启
	IsQueue                string            `json:"is_queue,omitempty"`                   // 是否开启排队 0-关闭 1-开启
	IsPreOrder             string            `json:"is_pre_order,omitempty"`               // 是否开启预点餐 0-关闭 1-开启
	IsCallService          string            `json:"is_call_service,omitempty"`            // 是否开启呼叫服务 0-关闭 1-开启
	IsCustomerOrder        string            `json:"is_customer_order,omitempty"`          // 是否开启顾客点餐 0-关闭 1-开启
	IsVoiceRemind          string            `json:"is_voice_remind,omitempty"`            // 是否开启语音提醒 0-关闭 1-开启
	IsShowSoldOut          string            `json:"is_show_sold_out,omitempty"`           // 是否显示售罄商品 0-不显示 1-显示
	IsBuffetOrderLimit     string            `json:"is_buffet_order_limit,omitempty"`      // 是否开启自助餐点餐限制 0-关闭 1-开启
	BuffetOrderLimit       *BuffetOrderLimit `json:"buffet_order_limit,omitempty"`         // 自助餐点餐限制
	IsOrderLimit           string            `json:"is_order_limit,omitempty"`             // 是否开启点餐限制 0-关闭 1-开启
	OrderLimit             *OrderLimit       `json:"order_limit,omitempty"`                // 点餐限制
	IsShowScanSoldOut      int               `json:"is_show_scan_sold_out,omitempty"`      // 扫码点餐是否显示售罄商品 0-不显示 1-显示
	IsShowAssistantSoldOut int               `json:"is_show_assistant_sold_out,omitempty"` // 点餐助手是否显示售罄商品 0-不显示 1-显示
	MenuShowSoldOut        string            `json:"menu_show_sold_out,omitempty"`         // 是否显示售罄商品 0-关闭（不显示售罄） 1-开启（显示售罄）
	MemberShowSoldOut      string            `json:"member_show_sold_out,omitempty"`       // 是否显示会员端售罄商品 0-关闭（不显示售罄） 1-开启（显示售罄）
}

// ZeroingMethodItem 抹零方式项（字段顺序与旧服务保持一致）
type ZeroingMethodItem struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

// CheckoutZeroingMethodItem 结账抹零方式项（字段顺序与旧服务保持一致）
type CheckoutZeroingMethodItem struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

// GiftMethodItem 赠送方式项（字段顺序与旧服务保持一致）
type GiftMethodItem struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

// FreeMethodItem 免单方式项（字段顺序与旧服务保持一致）
type FreeMethodItem struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

// NewBusinessSetting 创建新的业务设置
func NewBusinessSetting() *BusinessSetting {
	return &BusinessSetting{
		ZeroingMethodList:          make([]ZeroingMethodItem, 0),
		CheckoutZeroingMethodList:  make([]CheckoutZeroingMethodItem, 0),
		GiftMethodList:             make([]GiftMethodItem, 0),
		FreeMethodList:             make([]FreeMethodItem, 0),
		BatchProductUuids:          make([]uint64, 0),
		DiscountAuthorizedStaffIds: make([]uint64, 0),
		RefundAuthorizedStaffIds:   make([]uint64, 0),
		DishCardStyle:              "1", // 默认样式
		DishCardStyleTime:          strconv.Itoa(int(time.Now().Unix())),
		StartSerialNo:              "1",    // 默认起始流水号
		IsBatch:                    "0",    // 默认关闭分批送厨
		BatchCookingMode:           "post", // 默认后厨分批
		SafetyStockType:            "1",    // 默认安全库存类型
	}
}

// UpdateDishCardStyle 更新菜品卡片样式
func (b *BusinessSetting) UpdateDishCardStyle(style string) {
	b.DishCardStyle = style
	b.DishCardStyleTime = strconv.Itoa(int(time.Now().Unix()))
}

// UpdateBatchSettings 更新分批设置
func (b *BusinessSetting) UpdateBatchSettings(isBatch string, batchCookingMode string) {
	b.IsBatch = isBatch
	b.BatchCookingMode = batchCookingMode
}

// AddDiscountAuthorizedStaff 添加折扣授权员工
func (b *BusinessSetting) AddDiscountAuthorizedStaff(staffId uint64) {
	if !b.containsStaff(b.DiscountAuthorizedStaffIds, staffId) {
		b.DiscountAuthorizedStaffIds = append(b.DiscountAuthorizedStaffIds, staffId)
	}
}

// RemoveDiscountAuthorizedStaff 移除折扣授权员工
func (b *BusinessSetting) RemoveDiscountAuthorizedStaff(staffId uint64) {
	b.DiscountAuthorizedStaffIds = b.removeStaff(b.DiscountAuthorizedStaffIds, staffId)
}

// AddRefundAuthorizedStaff 添加退款授权员工
func (b *BusinessSetting) AddRefundAuthorizedStaff(staffId uint64) {
	if !b.containsStaff(b.RefundAuthorizedStaffIds, staffId) {
		b.RefundAuthorizedStaffIds = append(b.RefundAuthorizedStaffIds, staffId)
	}
}

// RemoveRefundAuthorizedStaff 移除退款授权员工
func (b *BusinessSetting) RemoveRefundAuthorizedStaff(staffId uint64) {
	b.RefundAuthorizedStaffIds = b.removeStaff(b.RefundAuthorizedStaffIds, staffId)
}

// UpdateBatchCookingMode 更新分批烹饪模式
func (b *BusinessSetting) UpdateBatchCookingMode(cookingMode string) {
	b.BatchCookingMode = cookingMode
}

func (b *BusinessSetting) containsStaff(staffIds []uint64, staffId uint64) bool {
	for _, id := range staffIds {
		if id == staffId {
			return true
		}
	}
	return false
}

func (b *BusinessSetting) removeStaff(staffIds []uint64, staffId uint64) []uint64 {
	for i, id := range staffIds {
		if id == staffId {
			return append(staffIds[:i], staffIds[i+1:]...)
		}
	}
	return staffIds
}

// DefaultBusinessSetting 获取默认业务设置
func DefaultBusinessSetting() *BusinessSetting {
	return &BusinessSetting{
		ZeroingMethodList: []ZeroingMethodItem{
			{Key: "0", Name: "实款实收"},
			{Key: "1", Name: "抹分"},
			{Key: "2", Name: "抹角"},
			{Key: "3", Name: "四舍五入保留一位小数"},
			{Key: "4", Name: "四舍五入到整数"},
		}, // 优惠折扣自动抹零方式列表
		CheckoutZeroingMethodList: []CheckoutZeroingMethodItem{
			{Key: "0", Name: "实款实收"},
			{Key: "1", Name: "抹分"},
			{Key: "2", Name: "抹角"},
			{Key: "5", Name: "抹元"},
		}, // 结账自动抹零方式列表
		ZeroingMethod:         "0", // 优惠折扣自动抹零方式
		CheckoutZeroingMethod: "0", // 结账自动抹零方式
		GiftMethodList: []GiftMethodItem{
			{Key: "10", Name: "计入总销售额、优惠折扣"},
			{Key: "20", Name: "不计入总销售额、优惠折扣"},
		}, // 赠菜计算方式列表
		GiftMethod: "10", // 赠菜计算方式，10-计入总销售额、优惠折扣 20-不计入总销售额、优惠折扣
		FreeMethodList: []FreeMethodItem{
			{Key: "10", Name: "计入总销售额、优惠折扣、服务费、税费"},
			{Key: "20", Name: "不计入总销售额、优惠折扣、服务费、税费"},
		}, // 免单计算方式列表
		FreeMethod:                    "10",              // 免单计算方式，10-计入总销售额、优惠折扣、服务费、税费 20-不计入总销售额、优惠折扣、服务费、税费
		DiscountMethod:                "10",              // 折扣计算方式 10-按百分比 20-直接减免
		QrCode:                        "123456",          // 电子菜单二维码校验失效值，6位数数字
		NoClearTable:                  "0",               // 结账后不清台 0-清台 1-不清台
		IsNeedPassword:                "1",               // 取消订单/退菜 0-无需密码 1-需要密码
		DishCardStyle:                 "0",               // 菜品卡片样式 0-无图模式 1-图片模式
		DishCardStyleTime:             "0",               // 菜品卡片样式最后更新时间
		IsInvoice:                     "0",               // 开票信息 0-不需要填写 1-需要填写
		OpeningHours:                  "00:00-23:59",     // 营业时间 00:00-23:59
		DeliveryPriceRatio:            100,               // 外送商品价格和商品原价比例
		StartSerialNo:                 "0001",            // 开始序列号
		IsBatch:                       "0",               // 是否是分批商品 0-否 1-是
		BatchProductUuids:             make([]uint64, 0), // 分批商品UUID列表
		BatchTagNum:                   0,                 // 分批类型数量
		SafetyStockType:               "1",               // 安全库存类型 1-门店纬度 2-仓库纬度，默认为1
		RequiredParentCompanyApproval: "0",               // 调拨规则-经过上级门店审批 "0"-否 "1"-是, 总部和上级(有下级门店)支持此选项
		ViaParentCompanyWarehouse:     "0",               // 调拨规则-经过上级门店仓库 "0"-否 "1"-是, 总部和上级(有下级门店)支持此选项
		BatchCookingMode:              "post",            // 分批送厨模式: "pre" 前置 / "post" 后置，默认 "post"
		EnableOrderSource:             "0",               // 外卖功能开关 0-关闭 1-开启
		EnableNationality:             "0",               // 国籍功能开关 0-关闭 1-开启
		DiscountNeedPassword:          "0",               // 折扣操作是否需要密码 0-否 1-是
		DiscountAuthorizedStaffIds:    make([]uint64, 0), // 折扣操作授权员工ID列表
		RefundNeedPassword:            "0",               // 退款操作是否需要密码 0-否 1-是
		RefundAuthorizedStaffIds:      make([]uint64, 0), // 退款操作授权员工ID列表
		// 以下字段在旧服务 Business 中不存在，设为空使 omitempty 生效
		// IsAutoOrder, AutoOrderLimit, IsAutoVoice 等字段在新服务内部使用，但不在 GetBusinessSetting 返回
		// 这些字段的默认值保持为空，确保与旧服务输出一致
	}
}

// BuffetOrderLimit 自助餐点餐限制
type BuffetOrderLimit struct {
	IsLimitTime string `json:"is_limit_time"` // 是否开启时间限制 0-关闭 1-开启
	LimitTime   string `json:"limit_time"`    // 时间限制（分钟）
	IsLimitNum  string `json:"is_limit_num"`  // 是否开启数量限制 0-关闭 1-开启
	LimitNum    string `json:"limit_num"`     // 数量限制
}

// OrderLimit 点餐限制
type OrderLimit struct {
	IsLimitTime string `json:"is_limit_time"` // 是否开启时间限制 0-关闭 1-开启
	LimitTime   string `json:"limit_time"`    // 时间限制（分钟）
	IsLimitNum  string `json:"is_limit_num"`  // 是否开启数量限制 0-关闭 1-开启
	LimitNum    string `json:"limit_num"`     // 数量限制
}
