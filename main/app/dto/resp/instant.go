package resp

import (
	"ttpos-server-go/app/dto"
)

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

type ProductMustPlanList struct {
	List []InstantProductMustPlan `json:"list"` // 必点方案列表
}

// InstantProductMustPlan 必点方案
type InstantProductMustPlan struct {
	Uuid         uint64             `json:"uuid"`           // 方案uuid
	Name         string             `json:"name"`           // 方案名称
	MustType     int                `json:"must_type"`      // 必点类型.0-每笔订单必点1份 1-每人必点1份
	MustRule     int                `json:"must_rule"`      // 必点规则.0-固定商品 1-可选商品
	CanChangeNum bool               `json:"can_change_num"` // 顾客可修改必点数量
	MealNum      uint               `json:"meal_num"`       // 就餐人数
	SelectedNum  float64            `json:"selected_num"`   // 已选数量。已选择xx份
	NeedNum      float64            `json:"need_num"`       // 这个商品还需要点的数量。还差xx份
	Products     ProductPackageList `json:"products"`       // 商品列表
}
type ProductPackageList struct {
	List []InstantMustPlanProductStat `json:"list"`
}

// ProductAutoAddReq 自动加购商品请求参数
type ProductAutoAddReq struct {
	FlavorUuid        uint64   `json:"flavor_uuid"`    // 某个规格商品ID
	SauceUuidList     []uint64 `json:"sauce_uuid"`     // 小料ID列表
	AttributeUuidList []uint64 `json:"attribute_uuid"` // 属性ID列表
	Num               float64  `json:"num"`            // 加购数量
}

type InstantMustPlanProductStat struct {
	Product           InstantMustPlanProduct `json:"product"`
	IsAutoAdd         bool                   `json:"is_auto_add"`          // 是否是自动加购的商品。是则自动加入购物车
	ProductAutoAddReq ProductAutoAddReq      `json:"product_auto_add_req"` // 自动加购商品请求参数
	SelectedNum       float64                `json:"selected_num"`         // 已选数量
	MustNum           float64                `json:"must_num"`             // 这个商品必选点的数量。还需点数量=must_num-selected_num
	NeedNum           float64                `json:"need_num"`             // 这个商品还需要点的数量。还需点数量=must_num-selected_num
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
	MinSelect  uint               `json:"min_select"`  // 最小可选的属性数量。0时不限制选择数量
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
	MemberInfo     *MemberInfo               `json:"member_info,omitempty"`     // 会员信息。如果订单选择了会员，则返回会员信息
	CouponList     *CouponList               `json:"coupon_list"`               // 优惠券列表
	PaymentOrders  PaymentInfoList           `json:"payment_orders"`            // 支付订单列表
	PaymentMethods PaymentMethodList         `json:"payment_methods"`           // 支付方式列表
	Amounts        PaymentMethodAmountList   `json:"amounts"`                   // 支付方式列表及订单金额信息
	PointsExchange PointsExchangeInfo        `json:"points_exchange,omitempty"` // 积分抵扣信息
	ActivityList   FullReductionActivityList `json:"activity_list"`             // 满减活动列表
}

type CouponList struct {
	List []Coupon `json:"list"`
}

type Coupon struct {
	Uuid           uint64                `json:"uuid"`             // 优惠券UUID
	Name           string                `json:"name"`             // 优惠券名称
	Requirement    string                `json:"type"`             // 优惠券类型 none-无门槛（任何人可以使用） marketing-会员账户里的优惠券，营销活动获得
	Amount         float64               `json:"amount"`           // 优惠券金额
	Count          int                   `json:"count"`            // 优惠券数量
	IsSelected     bool                  `json:"is_selected"`      // 是否选中
	IsAvailable    bool                  `json:"is_available"`     // 是否可用
	DayStartTime   string                `json:"day_start_time"`   // 每日适用时段开始时间, hh:mm 格式。00:00-23:59表示全天可用
	DayEndTime     string                `json:"day_end_time"`     // 每日适用时段结束时间, hh:mm 格式。00:00-23:59表示全天可用
	ValidStartTime string                `json:"valid_start_time"` // 优惠券有效期，开始时间 格式：YYYY-MM-DD
	ValidEndTime   string                `json:"valid_end_time"`   // 优惠券有效期，结束时间 格式：YYYY-MM-DD
	CouponList     []*SampleMemberCoupon `json:"coupon_list"`      // 会员优惠券列表。列表的长度即为优惠券数量Count
	Sort           SortParam             `json:"-"`                // 排序参数,仅用于排序
}

type SortParam struct {
	IsAvailable       bool   `json:"is_available"`        // 优惠券是否可用。true-可用，false-不可用。在有效期内，且在使用时段内。none在前，marketing在后
	Type              string `json:"type"`                // 优惠券类型。none-通用，marketing-会员。none在前，marketing在后
	ValidEndTimestamp int64  `json:"valid_end_timestamp"` // 优惠券有效期结束时间戳，先到期的排前面
	Sort              int    `json:"sort"`                // 优惠券排序，小的在前
	CouponUuid        uint64 `json:"coupon_uuid"`         // 优惠券UUID。小的在前，即先创建的优惠券在前
}

type SampleMemberCoupon struct {
	Uuid         uint64 `json:"uuid"`        // 会员优惠券UUID
	Name         string `json:"name"`        // 会员优惠券名称
	CouponUuid   uint64 `json:"coupon_uuid"` // 优惠券UUID
	StartTime    int64  `gorm:"column:start_time;type:bigint(20);not null;comment:优惠券开始时间" json:"start_time"`
	EndTime      int64  `gorm:"column:end_time;type:bigint(20);not null;comment:优惠券结束时间" json:"end_time"`
	DayStartTime string `json:"day_start_time"` // 每日适用时段开始时间, hh:mm 格式。00:00-23:59表示全天可用
	DayEndTime   string `json:"day_end_time"`   // 每日适用时段结束时间, hh:mm 格式。00:00-23:59表示全天可用
}

type PointsExchangeInfo struct {
	MaxPoints          float64 `json:"max_points"`           // 最大可抵扣积分。不能大于（订单应收/积分抵扣比例）。不能大于会员余额
	PointsExchangeRate float64 `json:"points_exchange_rate"` // 积分抵扣比例。1个积分抵扣多少金额
	PayPoints          float64 `json:"pay_points"`           // 已抵扣积分
	PayPointsAmount    float64 `json:"pay_points_amount"`    // 已抵扣金额
	OpenPointsExchange bool    `json:"open_points_exchange"` // 是否开启积分抵扣。用于前端显示积分抵扣信息。
	CanChangePoints    bool    `json:"can_change_points"`    // 是否可以修改抵扣积分的数量。用于前端置灰修改按钮。
}

// GetZeroAmount 获取结账抹零金额
func (obj InstantOrderPaymentInfoResp) GetZeroAmount() float64 {
	// 遍历支付方式列表，找到结账抹零金额
	// 如果支付方式为有手续费时，结账抹零金额为0
	// 如果支付方式为无手续费时，结账抹零金额都相同，返回第一个的结账抹零金额
	for _, amount := range obj.Amounts.List {
		if amount.ZeroAmount > 0 {
			return amount.ZeroAmount
		}
	}
	return 0
}

// GetCommissionAmount 获取已经支付的手续费金额。 已经支付的手续费金额= 所有付款单的手续费之和
func (obj InstantOrderPaymentInfoResp) GetCommissionAmount() float64 {
	// Amounts.List列表中的每一个元素的CommissionFee都相同，所以直接返回第一个元素的CommissionFee
	if len(obj.Amounts.List) != 0 {
		return obj.Amounts.List[0].CommissionFee
	}
	// 如果支付方式列表为空，则返回0
	return 0
}

// GetUnpaidAmount 获取未收金额。未收金额=应收金额-实收金额
func (obj InstantOrderPaymentInfoResp) GetUnpaidAmount(paymentMethodUuid ...uint64) float64 {
	//  Amounts.List列表中的每一个元素的UnpaidAmount都相同，所以直接返回第一个元素的UnpaidAmount
	if len(paymentMethodUuid) > 0 {
		for _, amount := range obj.Amounts.List {
			if amount.PaymentMethodUuid == paymentMethodUuid[0] {
				return amount.UnpaidAmount
			}
		}
	}
	if len(obj.Amounts.List) != 0 {
		return obj.Amounts.List[0].UnpaidAmount
	}
	// 如果Amounts.List列表为空，则返回0
	return 0
}

type PaymentMethodAmountList struct {
	List []PaymentMethodAmount `json:"list"`
}
type PaymentMethodAmount struct {
	SaleOrderOriginAmount float64 `json:"sale_order_origin_amount"` // 订单原价。订单原价=商品总价（折前价）+服务费（折前价）+消费税（折前价）
	SaleOrderCartAmount   float64 `json:"sale_order_cart_amount"`   // 购物车应收金额。
	SaleOrderAmount       float64 `json:"sale_order_amount"`        // 应收金额。
	UnpaidAmount          float64 `json:"unpaid_amount"`            // 未收金额。用于显示在金额输入框，默认显示未收金额。未收金额=应收金额-实付金额。实付金额指去掉手续费为这笔订单支付的金额
	ZeroAmount            float64 `json:"zero_amount"`              // 抹零金额。当支付方式为有手续费时，结账抹零金额为0。
	ZeroRule              uint8   `json:"zero_rule"`                // 结账抹零规格。0-实款实收 1-抹分 2-抹角 3-抹元. 当支付方式为有手续费时，值为0 实款实收
	IsAutoZero            bool    `json:"is_auto_zero"`             // 是否是自动抹零
	PaymentMethodUuid     uint64  `json:"payment_method_uuid"`      // 支付方式uuid。表示这个amount信息是当前端选择这个支付方式时显示的
	Code                  int     `json:"code"`                     // 支付方式代码。
	CouponExchangeAmount  float64 `json:"coupon_exchange_amount"`   // 优惠券抵扣金额
	ActivityAmount        float64 `json:"activity_amount"`          // 满减活动抵扣金额
	CommissionFee         float64 `json:"commission_fee"`           // 已付款的手续费。用于显示最终应收，前端显示的最终应收=应收金额+已付款的手续费+（当前支付方式的手续费费率*当前支付方式的金额输入框的值）
}

// 找零金额，由前端自己计算后显示。 找零金额=金额输入框的值 - 未收金额
// 手续费金额，由前端自己计算后显示。手续费金额= 金额输入框的值 * 支付方式的手续费费率
// 实际应付金额，由前端自己计算后显示，实际应付金额=金额输入框的值*（1+手续费率）=金额输入框的值 + 手续费金额

type PaymentInfoList struct {
	List []PaymentOrder `json:"list"`
}

type MemberInfo struct {
	Uuid          uint64    `json:"uuid"`           // 会员UUID
	Nickname      string    `json:"nickname"`       // 会员名称
	Card          CardInfo  `json:"card"`           // 会员卡信息
	Level         LevelInfo `json:"level"`          // 会员等级
	Balance       float64   `json:"balance"`        // 会员余额
	Points        float64   `json:"points"`         // 会员积分
	RechargeMoney float64   `json:"recharge_money"` // 会员累计充值金额
}

type LevelInfo struct {
	Name string `json:"name"` // 会员等级名称
}

type CardInfo struct {
	Name string `json:"name"` // 卡名称
}

type InstantHideOrderListResp struct {
	List []InstantHideSaleBill `json:"list"` // 点餐订单列表
	Meta dto.PageResponse      `json:"meta"` // 分页信息
}

type InstantHideSaleBill struct {
	SaleBillUuid uint64                     `json:"sale_bill_uuid"` // 销售账单UUID
	SerialNo     string                     `json:"serial_no"`      // 订单编号
	Amount       float64                    `json:"amount"`         // 订单总价。订单总价=销售订单的应收金额之和
	HideBillTime int64                      `json:"hide_bill_time"` // 挂单时间
	Products     InstantHideSaleProductList `json:"products"`       // 商品列表
}

type InstantHideSaleProductList struct {
	List []Product `json:"list"`
}

// 销售订单的支付页的二维码信息
type InstantOrderPaymentQrcodeInfoResp struct {
	PaymentOrderUuid uint64  `json:"payment_order_uuid"` // 支付单uuid (当/cashier/desk/order/payment/info接口的payment_orders中的存在相同的uuid时证明已经支付)
	QrCode           string  `json:"qr_code"`            // 支付单二维码 (永远都是返回base64图片给前端直接显示)
	QrCodeExpireSec  int64   `json:"qr_code_expire_sec"` // 支付单二维码剩余时间（秒）(少于等于0的时候 需要重新生成二维码, 再次请求当前接口就行)
	Status           int     `json:"status"`             // 支付单状态 支付状态, 0-未支付 1-已支付 (可选择轮询当前接口，获取支付状态)
	PaymentAmount    float64 `json:"payment_amount"`     // 支付金额
}

type InstantOrderMemberList struct {
	List  []InstantOrderMember    `json:"list"`
	Extra InstantOrderMemberExtra `json:"extra"` // 会员额外信息
}

type InstantOrderMember struct {
	Uuid     uint64 `json:"uuid"`     // 会员UUID
	Nickname string `json:"nickname"` // 会员名称
	Phone    string `json:"phone"`    // 会员手机号
}

type InstantOrderMemberExtra struct {
	IsCheckout        bool `json:"is_checkout"`         // 是否结账
	IsPartialCheckout bool `json:"is_partial_checkout"` // 是否部分结账
}

// FullReductionActivityList 满减活动列表
type FullReductionActivityList struct {
	List []FullReductionActivityItem `json:"list"`
}

// FullReductionActivityItem 满减活动项
type FullReductionActivityItem struct {
	Uuid           uint64             `json:"uuid"`            // 活动UUID
	LocaleName     dto.LocaleResponse `json:"locale_name"`     // 活动多语言名称
	ActivityType   uint               `json:"activity_type"`   // 活动类型：1-阶梯满减，2-循环满减
	StartDate      string             `json:"start_date"`      // 开始日期
	EndDate        string             `json:"end_date"`        // 结束日期
	StartTime      string             `json:"start_time"`      // 开始时间（HH:mm）
	EndTime        string             `json:"end_time"`        // 结束时间（HH:mm）
	IsAllDay       bool               `json:"is_all_day"`      // 是否全天
	Rules          []ActivityRule     `json:"rules"`           // 活动规则列表
	IsAvailable    bool               `json:"is_available"`    // 是否可用
	IsSelected     bool               `json:"is_selected"`     // 是否已选中
	DiscountAmount float64            `json:"discount_amount"` // 抵扣金额（如果已选中）
}

// ActivityRule 活动规则
type ActivityRule struct {
	Threshold float64 `json:"threshold"` // 满减阈值
	Discount  float64 `json:"discount"`  // 减价金额
}
