package resp

import "ttpos-server-go/app/dto"

// 创建点餐订单响应
type CreateInstantOrderResp struct {
	SaleBillUuid  uint64 `json:"sale_bill_uuid"`  // 销售账单UUID
	SaleOrderUuid uint64 `json:"sale_order_uuid"` // 销售订单UUID
}

// InstantOrder 点餐订单
type InstantOrder struct {
	SaleOrderUuid  uint64  `json:"sale_order_uuid"` // 销售订单UUID
	ProductCount   uint    `json:"product_count"`   // 产品数量
	ProductAmount  float64 `json:"product_amount"`  // 产品总金额
	ServiceFee     float64 `json:"service_fee"`     // 服务费
	TaxFee         float64 `json:"tax_fee"`         // 消费税
	DiscountAmount float64 `json:"discount_amount"` // 优惠折扣
	MemberDiscount float64 `json:"member_discount"` // 会员折扣
	TotalAmount    float64 `json:"total_amount"`    // 总金额
}

// InstantOrderList 点餐订单列表
type InstantOrderList struct {
	List []InstantOrder `json:"list"` // 点餐订单列表
}

// GetInstantOrderInfoResp 获取点餐订单详情响应
type GetInstantOrderInfoResp struct {
	SaleBillUuid  uint64             `json:"sale_bill_uuid"`  // 销售账单UUID
	IsKitchen     bool               `json:"is_kitchen"`      // 是否送厨: false-否 true-是
	SaleOrderList []InstantOrderList `json:"sale_order_list"` // 销售订单列表
}

type InstantProductMustPlanResp struct {
	List         []InstantProductMustPlan `json:"list"`                     // 必点方案列表
	ShopCartInfo *InstantShopCart         `json:"shop_cart_info,omitempty"` // 购物车信息。当必点方案中有自动加购商品时，返回购物车信息。后台会自动加购商品到购物车中，前端用这个购物车信息更新界面
}

// InstantProductMustPlan 必点方案
type InstantProductMustPlan struct {
	Name         string             `json:"name"`           // 方案名称
	MustType     int                `json:"must_type"`      // 必点类型.0-每笔订单必点1份 1-每人必点1份
	MustRule     int                `json:"must_rule"`      // 必点规则.1-固定商品 2-可选商品
	IsAutoCart   bool               `json:"is_auto_cart"`   // 自动加入购物车
	CanChangeNum bool               `json:"can_change_num"` // 顾客可修改必点数量
	SelectedNum  uint               `json:"selected_num"`   // 已选数量。各个商品的selected_num之和
	NeedNum      uint               `json:"need_num"`       // 这个商品还需要点的数量。各个商品的need_num之和
	Products     ProductPackageList `json:"products"`       // 商品列表
}
type ProductPackageList struct {
	List []InstantMustPlanProductStat `json:"list"`
}

type InstantMustPlanProductStat struct {
	Product     InstantMustPlanProduct `json:"product"`
	IsAutoAdd   bool                   `json:"is_auto_add"`  // 是否是自动加购的商品。是则自动加入购物车，并且不显示在“必选方案”的弹框中
	SelectedNum uint                   `json:"selected_num"` // 已选数量
	MustNum     uint                   `json:"must_num"`     // 这个商品必选点的数量。还需点数量=must_num-selected_num
	NeedNum     uint                   `json:"need_num"`     // 这个商品还需要点的数量。还需点数量=must_num-selected_num
}
type InstantMustPlanProduct struct {
	Uuid            uint64             `json:"uuid"`             // 商品product_package的uuid
	LocaleName      dto.LocaleResponse `json:"locale_name"`      // 商品名称
	Image           string             `json:"image"`            // 商品图片url
	Unit            dto.LocaleResponse `json:"unit"`             // 商品单位
	LimitNum        uint               `json:"limit_num"`        // 商品限购数量,0-不限购
	Price           float64            `json:"price"`            // 商品价格.所有规格中价格最小的价格
	Flavors         Flavors            `json:"flavors"`          // 商品规格列表
	Sauces          ProductSauces      `json:"sauces"`           // 商品小料信息
	AttributeGroups AttributeGroups    `json:"attribute_groups"` // 商品属性组列表
}

type Flavors struct {
	List []ProductFlavor `json:"list"`
}

type AttributeGroups struct {
	List []ProductAttributeGroup `json:"list"`
}

func (obj InstantMustPlanProduct) CanAutoJoinCart() bool {
	// 能自动加购的商品:单规格、无属性、无加料
	if len(obj.Flavors.List) == 1 {
		if len(obj.AttributeGroups.List) == 0 && len(obj.Sauces.List) == 0 {
			return true
		}
	}
	return false
	// 能自动加购的商品:单规格、属性不要求选、无加料
	//if len(obj.FlavorList) == 1 {
	//	if len(obj.AttributeGroups) >= 0 && len(obj.Sauces.List) == 0 {
	//		isMust := false
	//		for _, group := range obj.AttributeGroups {
	//			if group.IsMust == true {
	//				isMust = true
	//			}
	//		}
	//		if isMust == false {
	//			return true
	//		}
	//	}
	//}
	//// 能自动加购的商品:单规格、无属性、加料不要求选
	//if len(obj.FlavorList) == 1 {
	//	if len(obj.AttributeGroups) == 0 && len(obj.Sauces.List) >= 0 {
	//			if obj.Sauces.IsMust == false {
	//				return true
	//			}
	//		}
	//	}
	//}
}

type ProductFlavor struct {
	Uuid       uint64             `json:"uuid"`        // 商品规格UUID
	LocaleName dto.LocaleResponse `json:"locale_name"` // 商品规格名称
	Price      float64            `json:"price"`       // 商品规格价格
	StockNum   int                `json:"stock_num"`   // 库存数量
}

type ProductSauces struct {
	List      []ProductSauce `json:"list"`       // 小料列表
	IsMust    bool           `json:"is_must"`    // 是否必选小料
	MaxSelect int            `json:"max_select"` // 小料最大可选数量
}

type ProductSauce struct {
	Uuid              uint64             `json:"uuid"`                // 商品小料UUID
	LocaleName        dto.LocaleResponse `json:"locale_name"`         // 商品小料名称
	Price             float64            `json:"price"`               // 商品小料价格
	IsDefaultSelected bool               `json:"is_default_selected"` // 是否默认选中
	StockNum          int                `json:"stock_num"`           // 库存数量
}

type ProductAttributeGroup struct {
	LocaleName dto.LocaleResponse `json:"locale_name"` // 商品属性组名称
	Attributes Attributes         `json:"attributes"`  // 商品属性列表
	IsMust     bool               `json:"is_must"`     // 是否必选
	MaxSelect  uint               `json:"max_select"`  // 最大可选的属性数量。0时不限制选择数量
}

type Attributes struct {
	List []Attribute `json:"list"`
}
type Attribute struct {
	Uuid              uint64             `json:"uuid"`                // 商品属性UUID
	LocaleName        dto.LocaleResponse `json:"locale_name"`         // 商品属性名称
	IsDefaultSelected bool               `json:"is_default_selected"` // 是否默认选中
}

// 销售订单的支付页信息
type InstantOrderPaymentInfoResp struct {
	MemberInfo     MemberInfo              `json:"member_info"`     // 会员信息。如果订单选择了会员，则返回会员信息
	PaymentOrders  PaymentInfoList         `json:"payment_orders"`  // 支付订单列表
	PaymentMethods PaymentMethodList       `json:"payment_methods"` // 支付方式列表
	Amounts        PaymentMethodAmountList `json:"amounts"`         // 支付方式列表及订单金额信息
}

type PaymentMethodAmountList struct {
	List []PaymentMethodAmount `json:"list"`
}
type PaymentMethodAmount struct {
	SaleOrderOriginAmount float64 `json:"sale_order_origin_amount"` // 订单原价。订单原价=商品总价（折前价）+服务费（折前价）+消费税（折前价）
	SaleOrderAmount       float64 `json:"sale_order_amount"`        // 应收金额。
	UnpaidAmount          float64 `json:"unpaid_amount"`            // 未收金额。用于显示在金额输入框，默认显示未收金额。未收金额=应收金额-实付金额。实付金额指去掉手续费为这笔订单支付的金额
	ZeroAmount            float64 `json:"zero_amount"`              // 抹零金额。当支付方式为有手续费时，结账抹零金额为0。
	ZeroRule              uint8   `json:"zero_rule"`                // 结账抹零规格。0-实款实收 1-抹分 2-抹角 3-抹元. 当支付方式为有手续费时，值为0 实款实收
	PaymentMethodUuid     uint64  `json:"payment_method_uuid"`      // 支付方式uuid。表示这个amount信息是当前端选择这个支付方式时显示的
	CommissionFee         float64 `json:"commission_fee"`           // 已付款的手续费。用于显示最终应收，前端显示的最终应收=应收金额+已付款的手续费+（当前支付方式的手续费费率*当前支付方式的金额输入框的值）
}

// 找零金额，由前端自己计算后显示。 找零金额=金额输入框的值 - 未收金额
// 手续费金额，由前端自己计算后显示。手续费金额= 金额输入框的值 * 支付方式的手续费费率
// 实际应付金额，由前端自己计算后显示，实际应付金额=金额输入框的值*（1+手续费率）=金额输入框的值 + 手续费金额

type PaymentInfoList struct {
	List []PaymentOrder `json:"list"`
}

type MemberInfo struct {
	Uuid    uint64    `json:"uuid"`    // 会员UUID
	Name    string    `json:"name"`    // 会员名称
	Card    CardInfo  `json:"card"`    // 会员卡信息
	Level   LevelInfo `json:"level"`   // 会员等级
	Balance float64   `json:"balance"` // 会员余额
	Points  float64   `json:"points"`  // 会员积分
}

type LevelInfo struct {
	Name string `json:"name"` // 会员等级名称
}

type CardInfo struct {
	Name string `json:"name"` // 卡名称
}
