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
	Products     ProductPackageList `json:"products"`       // 商品列表
}
type ProductPackageList struct {
	List []*InstantMustPlanProduct `json:"list"`
}
type InstantMustPlanProduct struct {
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
