package model

import (
	"fmt"
	"math"
	"strings"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/resp"
	settingResp "ttpos-server-go/app/dto/resp/setting"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/i18n"
	"ttpos-server-go/pkg/utils"

	"github.com/duke-git/lancet/cryptor"
	"github.com/shopspring/decimal"
)

// SaleOrder 销售订单 `ttpos_sale_order`
type SaleOrder struct {
	BaseModel
	// 基础标识字段
	OrderNo     string `gorm:"column:order_no;comment:订单编号" json:"order_no"`
	Status      uint   `gorm:"column:status;comment:订单状态, 0-未结账 1-已结账" json:"status"`
	IsFree      uint   `gorm:"column:is_free;comment:是否免单, 0-否 1-是" json:"is_free"`
	FreeReason  string `gorm:"column:free_reason;comment:免单原因" json:"free_reason"`
	CashierName string `gorm:"column:cashier_name;type:varchar(255);default:'';comment:收银员名称" json:"cashier_name"`
	DeviceId    string `gorm:"column:device_id;type:varchar(255);default:'';comment:设备ID,用于标识订单来源设备.来源h5时，device_id为h5" json:"device_id"`

	// 关联ID字段
	ConsumerUuid uint64 `gorm:"column:consumer_uuid;type:bigint(20);default:0;comment:消费者ID" json:"consumer_uuid"`
	CashierUuid  uint64 `gorm:"column:cashier_uuid;type:bigint(20);default:0;comment:收银员ID" json:"cashier_uuid"`
	SaleBillUuid uint64 `gorm:"column:sale_bill_uuid;type:bigint(20);default:0;comment:销售账单ID" json:"sale_bill_uuid"`

	// 商品金额相关字段
	ProductAmount         float64 `gorm:"column:product_amount;type:decimal(12,2);default:0;comment:商品金额" json:"product_amount"`
	ProductOriginalAmount float64 `gorm:"column:product_original_amount;type:decimal(12,2);default:0;comment:商品原始金额" json:"product_original_amount"`

	// 费用相关字段
	ServiceFee        float64 `gorm:"column:service_fee;type:decimal(12,2);default:0;comment:服务费" json:"service_fee"`
	TaxFee            float64 `gorm:"column:tax_fee;type:decimal(12,2);default:0;comment:税费" json:"tax_fee"`
	CustomDiscountFee float64 `gorm:"column:custom_discount_fee;type:decimal(12,2);default:0;comment:自定义折扣金额" json:"custom_discount_fee"`
	MemberDiscountFee float64 `gorm:"column:member_discount_fee;type:decimal(12,2);default:0;comment:会员折扣金额" json:"member_discount_fee"`

	// 订单总额相关字段
	OriginAmount float64 `gorm:"column:origin_amount;type:decimal(12,2);default:0;comment:原始应收金额(折前价)。商品未含税时，总金额=商品金额+服务费+税费。商品已含税时，总金额=商品金额（含商品消费税）+服务费+税费（只有服务费税）" json:"origin_amount"`
	Amount       float64 `gorm:"column:amount;type:decimal(12,2);default:0;comment:应收金额(折后价)。商品未含税时，总金额=商品金额+服务费+税费。商品已含税时，总金额=商品金额（含商品消费税）+服务费+税费（只有服务费税）" json:"amount"`
	CustomAmount float64 `gorm:"column:custom_amount;type:decimal(12,2);default:-1;comment:整单改价金额。改价后，应收金额=整单改价金额，前端优先显示改价后的金额，改价金额不能为负数。当为-1时，表示不改价，显示amount改收金额" json:"custom_amount"`

	// 时间相关字段
	FinishTime int64 `gorm:"column:finish_time;type:int(10);default:0;comment:完成时间（时间戳）" json:"finish_time"`

	MemberDiscountRate     float64 `gorm:"column:member_discount_rate;type:decimal(12,2);default:1;comment:会员折扣率(0-100%)，默认100%，取值范围0-1，如折扣率为10%，则取值为0.1	" json:"member_discount_rate"`
	MemberCardDiscountRate float64 `gorm:"column:member_card_discount_rate;type:decimal(12,2);default:1;comment:会员卡折扣率(0-100%)，默认100%，取值范围0-1，如折扣率为10%，则取值为0.1" json:"member_card_discount_rate"`
	CustomDiscountRate     float64 `gorm:"column:custom_discount_rate;type:decimal(12,2);default:1;comment:自定义折扣率(0-100%)，默认100%，取值范围0-1，如折扣率为10%，则取值为0.1" json:"custom_discount_rate"`

	// 抹零相关
	ZeroRule         uint8   `gorm:"column:zero_rule;type:tinyint(1);default:0;comment:优惠折扣抹零, 0-实款实收 1-抹分 2-抹角 3-四舍五入保留一位小数 4-四舍五入保留整数" json:"zero_rule"`
	ZeroFee          float64 `gorm:"column:zero_fee;type:decimal(12,2);default:0;comment:优惠折扣抹零金额" json:"zero_fee"`
	ZeroCheckoutRule uint8   `gorm:"column:zero_checkout_rule;type:tinyint(1);default:0;comment:结账抹零, 0-实款实收 1-抹分 2-抹角 5-抹元（为了整体无歧义，抹元使用5）" json:"zero_checkout_rule"`

	// 积分抵扣相关
	PayPoints          float64 `gorm:"column:pay_points;type:decimal(12,2);default:0;comment:抵扣积分,用了多少积分进行抵扣" json:"pay_points"`
	PayPointsAmount    float64 `gorm:"column:pay_points_amount;type:decimal(12,2);default:0;comment:抵扣金额,积分 抵扣了多少金额" json:"pay_points_amount"`
	PointsExchangeRate float64 `gorm:"column:points_exchange_rate;type:decimal(12,4);default:0;comment:积分抵扣汇率,1积分抵扣多少元" json:"points_exchange_rate"`
	AutoPointsExchange uint    `gorm:"column:auto_points_exchange;type:tinyint(1);default:0;comment:积分抵扣类型,0-手动抵扣 1-自动抵扣" json:"auto_points_exchange"`

	// 结账完成后才记录的字段
	CouponAmount         float64 `gorm:"column:coupon_amount;type:decimal(12,2);default:0;comment:优惠券抵扣金额，实际抵扣金额" json:"coupon_amount"`
	PaymentAmount        float64 `gorm:"column:payment_amount;type:decimal(12,2);default:0;comment:支付金额,支付金额=订单总金额+支付手续费" json:"payment_amount"`
	ChangeAmount         float64 `gorm:"column:change_amount;type:decimal(12,2);default:0;comment:找零金额,结账完成后才记录" json:"change_amount"`
	ZeroCheckoutFee      float64 `gorm:"column:zero_checkout_fee;type:decimal(12,2);default:0;comment:结账抹零金额" json:"zero_checkout_fee"`
	FinalPrice           float64 `gorm:"column:final_price;type:decimal(12,2);default:0;comment:最终应收金额。最终应收金额=应收金额+手续费-结账抹零金额" json:"final_price"`
	PaymentCommissionFee float64 `gorm:"column:payment_commission_fee;type:decimal(12,2);default:0;comment:支付手续费,关联付款单的支付手续费之和" json:"payment_commission_fee"`
	GiftAmount           float64 `gorm:"column:gift_amount;type:decimal(12,2);default:0;comment:赠菜金额,(销售订单赠菜商品.总最终单价)之和" json:"gift_amount"`
	GiftPoints           float64 `gorm:"column:gift_points;type:decimal(12,2);default:0;comment:赠送积分,应收金额amount*积分赠送比例" json:"gift_points"`
	GiftPointsRate       float64 `gorm:"column:gift_points_rate;type:decimal(12,4);default:0;comment:赠送积分比例或每人赠送积分数量。赠送积分比例,取值范围0-1。结账后记录，不受后台改变" json:"gift_points_rate"`
	GiftPointsType       uint8   `gorm:"column:gift_points_type;type:tinyint(1);default:0;comment:赠送积分类型, 0-按比例赠送 1-按人数固定金额赠送" json:"gift_points_type"`
	MemberLevelName      string  `gorm:"column:member_level_name;type:varchar(255);default:'';comment:会员等级名称" json:"member_level_name"`
	MemberBalance        float64 `gorm:"column:member_balance;type:decimal(12,2);default:0;comment:会员余额,会员消费本单后剩余的余额" json:"member_balance"`
	Unit                 string  `gorm:"column:unit;type:varchar(255);default:0;comment:金额的单位,$-美元 ￥-人民币,用于显示订单金额价值" json:"unit"`

	// erp相关
	ErpProductsInvoiceName string `gorm:"column:erp_products_invoice_name;type:varchar(255);comment:商品发票名称;NOT NULL" json:"erp_products_invoice_name"`
	ErpMaterialInvoiceName string `gorm:"column:erp_material_invoice_name;type:varchar(255);comment:原材料发票名称;NOT NULL" json:"erp_material_invoice_name"`

	// 关联对象
	PaymentOrders                []*PaymentOrder                `gorm:"foreignKey:RelatedUuid;references:uuid"` // 支付订单，也叫付款单
	Member                       *Member                        `gorm:"foreignKey:ConsumerUuid;references:uuid"`
	SaleOrderProducts            []*SaleOrderProduct            `gorm:"foreignKey:SaleOrderUuid;references:uuid"`
	ReturnOrders                 []ReturnOrder                  `gorm:"foreignKey:RelatedOrderUuid;references:uuid"`
	SaleOrderBuffetCustomerTypes []*SaleOrderBuffetCustomerType `gorm:"foreignKey:SaleOrderUuid;references:uuid"`
	SaleOrderBuffetDelayProducts []*SaleOrderBuffetDelayProduct `gorm:"foreignKey:SaleOrderUuid;references:uuid"`
	FreeReasons                  []*SaleOrderProductReason      `gorm:"foreignKey:SaleOrderUuid;references:uuid"`
	InvoiceInfo                  *SaleOrderInvoiceInfo          `gorm:"foreignKey:SaleOrderUuid;references:uuid"`
	SaleBill                     *SaleBill                      `gorm:"foreignKey:SaleBillUuid;references:uuid"`
	MemberPointLogs              []*MemberPointLog              `gorm:"foreignKey:RelatedUuid;references:uuid"`   // 关联积分变动记录.赠送积分、退款积分、反结账积分
	Coupons                      []*SaleOrderCoupon             `gorm:"foreignKey:SaleOrderUuid;references:uuid"` // 订单使用的优惠券
	// 虚拟字段，用于标记当前子单是第几个
	index int `gorm:"-"`
}

// 获取订单的erp原材料列表
func (model *SaleOrder) GetErpProductBomMaterials() []*ErpProductBomMaterials {
	materials := make([]*ErpProductBomMaterials, 0)
	for _, saleOrderProduct := range model.SaleOrderProducts {
		saleOrderProductMaterials := saleOrderProduct.GetErpProductBomMaterials()
		for index, _ := range saleOrderProductMaterials {
			material := saleOrderProductMaterials[index]
			material.Num = decimal.NewFromFloat(material.Num).Mul(decimal.NewFromFloat(saleOrderProduct.Num)).Round(4).InexactFloat64()
		}
		materials = append(materials, saleOrderProductMaterials...)
	}
	// 去重
	splitKey := "--@--"
	materialMap := make(map[string]float64) // key: erp_code和uom, value: 原材料数量
	for _, material := range materials {
		key := fmt.Sprintf("%s%s%s", material.ErpCode, splitKey, material.Uom)
		materialMap[key] += material.Num
	}
	// 转换为列表
	materials = make([]*ErpProductBomMaterials, 0)
	for key, num := range materialMap {
		erpCode, uom := strings.Split(key, splitKey)[0], strings.Split(key, splitKey)[1]
		materials = append(materials, &ErpProductBomMaterials{ErpCode: erpCode, Uom: uom, Num: num})
	}
	return materials
}

// 获取套餐商品的子商品列表
func (model *SaleOrder) GetPackageSubProductList(saleOrderProductUuid uint64) []*SaleOrderProduct {
	subProducts := make([]*SaleOrderProduct, 0)
	for _, saleOrderProduct := range model.SaleOrderProducts {
		if saleOrderProduct.PackageUuid == saleOrderProductUuid {
			subProducts = append(subProducts, saleOrderProduct)
		}
	}
	return subProducts
}

// 获取订单付款信息
func (model *SaleOrder) GetPaymentInfoList() []resp.PaymentOrder {
	paymentOrders := make([]resp.PaymentOrder, 0)
	for _, paymentOrder := range model.PaymentOrders {
		order := resp.PaymentOrder{
			Uuid:                 paymentOrder.Uuid,
			PaymentMethodUuid:    paymentOrder.PaymentMethodUuid,
			PaymentMethodName:    paymentOrder.PaymentMethodName,
			PaymentMethodCode:    paymentOrder.PaymentMethod.Code,
			PaymentAmount:        paymentOrder.PaymentAmount,
			PaymentCommissionFee: paymentOrder.PaymentCommissionFee,
			Amount:               paymentOrder.Amount,
			DisabledCancel:       paymentOrder.PaymentMethod.IsDisabledCancel(),
		}
		paymentOrders = append(paymentOrders, order)
	}
	return paymentOrders
}

// 获取商品数量. 用于会员端订单
func (model *SaleOrder) GetProductNum() float64 {
	num := decimal.NewFromFloat(0)
	for _, saleOrderProduct := range model.SaleOrderProducts {
		if saleOrderProduct.IsDelete() {
			continue
		}
		num = num.Add(decimal.NewFromFloat(saleOrderProduct.Num))
	}
	return num.Round(2).InexactFloat64()
}

// 获取商品金额. 用于会员端订单
func (model *SaleOrder) GetProductAmount() float64 {
	amount := decimal.NewFromFloat(0)
	for _, saleOrderProduct := range model.SaleOrderProducts {
		amount = amount.Add(decimal.NewFromFloat(saleOrderProduct.GetOriginTotalPriceWithTax()))
	}
	return amount.Round(2).InexactFloat64()
}

// 获取商品原价. 用于会员端订单
func (model *SaleOrder) GetOriginProductAmount() float64 {
	amount := decimal.NewFromFloat(0)
	for _, saleOrderProduct := range model.SaleOrderProducts {
		amount = amount.Add(decimal.NewFromFloat(saleOrderProduct.GetTotalPriceOrigin()))
	}
	return amount.Round(2).InexactFloat64()
}

// 获取已选择的优惠券uuid
func (model *SaleOrder) GetSelectedCouponUuid() uint64 {
	if model.HasCoupon() {
		if model.Coupons[0].MarketingCouponUuid != 0 {
			return model.Coupons[0].MarketingCouponUuid
		}
		if model.Coupons[0].MemberCouponUuid != 0 {
			return model.Coupons[0].MemberCouponUuid
		}
	}
	return 0
}

// 销售订单是否使用了优惠券
func (model *SaleOrder) HasCoupon() bool {
	for _, coupon := range model.Coupons {
		if !coupon.IsDelete() {
			return true
		}
	}
	return false
}

// 判断这个优惠券是否已经应用到该销售订单
func (model *SaleOrder) HasCouponByUuid(couponUuid uint64, couponRequirement string) bool {
	for _, coupon := range model.Coupons {
		if !coupon.IsDelete() {
			if couponRequirement == constant.CouponRequirementNone {
				if coupon.MarketingCouponUuid == couponUuid && coupon.CouponRequirement == constant.CouponRequirementNone {
					return true
				}
			}
			if couponRequirement == constant.CouponRequirementMember {
				if coupon.MemberCouponUuid == couponUuid && coupon.CouponRequirement == constant.CouponRequirementMember {
					return true
				}
			}
		}
	}
	return false
}

// 获取优惠券
func (model *SaleOrder) GetCouponByUuid(couponUuid uint64, couponRequirement string) *SaleOrderCoupon {
	for _, coupon := range model.Coupons {
		if !coupon.IsDelete() {
			if couponRequirement == constant.CouponRequirementNone {
				if coupon.MarketingCouponUuid == couponUuid && coupon.CouponRequirement == constant.CouponRequirementNone {
					return coupon
				}
			}
			if couponRequirement == constant.CouponRequirementMember {
				if coupon.MemberCouponUuid == couponUuid && coupon.CouponRequirement == constant.CouponRequirementMember {
					return coupon
				}
			}
		}
	}
	return nil
}

// 新增一个优惠券
func (model *SaleOrder) AddCoupon(couponUuid uint64, couponRequirement string, couponAmount float64) {
	model.Coupons = append(model.Coupons, NewSaleOrderCoupon(model.Uuid, couponUuid, couponRequirement, couponAmount))
}

// 订单是否使用了会员余额支付
func (model *SaleOrder) HasBalancePayment() bool {
	for _, paymentOrder := range model.PaymentOrders {
		if paymentOrder.IsDelete() {
			continue
		}
		if paymentOrder.PaymentMethod.Code == constant.PaymentMethodCodeBalance {
			return true
		}
	}
	return false
}

// 订单是否使用了现金支付
func (model *SaleOrder) HasCashPayment() bool {
	for _, paymentOrder := range model.PaymentOrders {
		if paymentOrder.IsDelete() {
			continue
		}
		if paymentOrder.PaymentMethod.Code == constant.PaymentMethodCodeCash {
			return true
		}
	}
	return false
}

// 判断订单退款能否手动退积分（显示手动退积分输入框）
func (model *SaleOrder) CanManualReturnPoints() bool {
	// 订单使用了会员、按比例赠送积分且未抵扣积分时，手动退积分输入框。
	if model.ConsumerUuid != 0 && model.GiftPointsType == 0 && model.PayPoints == 0 {
		return false
	}
	// 订单没有会员，不显示
	if model.ConsumerUuid == 0 {
		return false
	}
	return true
}

// 获取销售订单已经退款的积分数量
func (model *SaleOrder) GetReturnedPoints() float64 {
	points := 0.0
	for _, memberPointLog := range model.MemberPointLogs {
		if memberPointLog.IsDelete() {
			continue
		}
		if memberPointLog.Scene == constant.MemberPointLogSceneRefund {
			points += memberPointLog.Value // 负数
		}
	}
	return math.Abs(points) // 绝对值,将负数变为正数
}

// 获取手动退款积分时，可退积分数量=订单赠送的积分-已经退回的积分. 如果会员积分余额不足时，可退积分数量等于会员积分余额
func (model *SaleOrder) GetManualReturnPoints() float64 {
	// 订单没有会员，可退积分为0
	if model.ConsumerUuid == 0 {
		return 0
	}
	// 可退积分数量=订单赠送的积分-已经退回的积分
	returnedPoints := model.GiftPoints - model.GetReturnedPoints()
	if returnedPoints < 0 {
		returnedPoints = 0
	}
	// 如果会员积分余额不足时，可退积分数量等于会员积分余额
	if returnedPoints > model.Member.GetPoints() {
		returnedPoints = model.Member.GetPoints()
	}
	return returnedPoints
}

// 订单金额。积分抵扣后、优惠券抵扣后的金额
func (model *SaleOrder) GetAmountValue() float64 {
	// 积分抵扣后的金额-优惠券抵扣金额
	return decimal.NewFromFloat(model.GetPointsExchangeAmount()).Sub(decimal.NewFromFloat(model.CalcCouponExchangeAmount())).Round(2).InexactFloat64()
}

// 获取积分抵扣后的应收金额。等于Amount-PayPointsAmount
func (model *SaleOrder) GetPointsExchangeAmount() float64 {
	return decimal.NewFromFloat(model.GetAmount()).Sub(decimal.NewFromFloat(model.PayPointsAmount)).Round(2).InexactFloat64()
}

// 获取最终应收金额（不含手续费）。等于Amount-积分抵扣金额-优惠券抵扣金额-结账抹零
func (model *SaleOrder) GetFinalNoFeeAmount() float64 {
	amount := decimal.NewFromFloat(model.GetAmountValue()).Sub(decimal.NewFromFloat(model.CalcCheckOutZeroFee())).Round(2).InexactFloat64() // 计算积分的基数，值为本订单的应收金额(已减积分抵扣金额)
	return amount
}

// 获取销售订单的序号
func (model *SaleOrder) GetIndex() int {
	return model.index
}

// 判断订单是不是优惠折扣自动抹零。如果SaleBillSetting中的自动抹零规格与订单的自动抹零规格不一致，则返回false
func (model *SaleOrder) IsAutoZeroDiscount(setting SaleBillSetting) bool {
	return model.ZeroRule == uint8(setting.ZeroRule)
}

// 判断订单是不是结账自动抹零。如果SaleBillSetting中的结账抹零规格与订单的结账抹零规格不一致，则返回false
func (model *SaleOrder) IsAutoCheckoutZeroDiscount(setting SaleBillSetting) bool {
	return model.ZeroCheckoutRule == uint8(setting.ZeroCheckoutRule)
}

// 获取自动抹零信息
func (model *SaleOrder) GetAutoDiscountMessage(setting SaleBillSetting, lang string) string {
	// 如果抹零金额为0，则不显示自动抹零信息
	if model.ZeroFee == 0 {
		return ""
	}

	if model.IsAutoZeroDiscount(setting) {
		return ParseAutoDiscountMessage(uint8(setting.ZeroRule), model.ZeroFee, lang)
	}
	return ""
}

// 解析自动抹零信息.  0-实款实收 1-抹分 2-抹角 3-四舍五入保留一位小数 4-四舍五入保留整数
func ParseAutoDiscountMessage(zeroRule uint8, zeroFee float64, lang string) string {
	switch zeroRule {
	case 1:
		return translate(zeroRule, lang) + "：" + fmt.Sprintf("%.2f", zeroFee)
	case 2:
		return translate(zeroRule, lang) + "：" + fmt.Sprintf("%.2f", zeroFee)
	case 3:
		return translate(zeroRule, lang) + "：" + fmt.Sprintf("%.2f", zeroFee)
	case 4:
		return translate(zeroRule, lang) + "：" + fmt.Sprintf("%.2f", zeroFee)
	}
	return ""
}

// "优惠折扣自动抹零-抹分"
func translate(zeroRule uint8, lang string) string {
	switch zeroRule {
	case 1:
		return i18n.Translate(lang, "优惠折扣自动抹零-抹分")
	case 2:
		return i18n.Translate(lang, "优惠折扣自动抹零-抹角")
	case 3:
		return i18n.Translate(lang, "优惠折扣自动抹零-四舍五入保留一位小数")
	case 4:
		return i18n.Translate(lang, "优惠折扣自动抹零-四舍五入保留整数")
	}
	return ""
}

// 判断销售订单是否已经结账
func (model *SaleOrder) IsSettled() bool {
	return model.Status == constant.SaleOrderStatusFinish
}

// 清除结账信息
func (model *SaleOrder) ClearSettleInfo() {
	model.PaymentAmount = 0
	model.ChangeAmount = 0
	model.ZeroCheckoutFee = 0
	model.FinalPrice = 0
	model.PaymentCommissionFee = 0
	model.GiftAmount = 0
	model.GiftPoints = 0
	model.GiftPointsRate = 0
	model.MemberLevelName = ""
	model.MemberBalance = 0
	model.Unit = ""
}

// 插入销售订单商品，如果商品已存在，则更新
func (model *SaleOrder) InsertSaleOrderProduct(saleOrderProducts []*SaleOrderProduct) {
	saleOrderProductMap := make(map[uint64]*SaleOrderProduct)
	for _, saleOrderProduct := range model.SaleOrderProducts {
		saleOrderProductMap[saleOrderProduct.Uuid] = saleOrderProduct
	}
	for i, _ := range saleOrderProducts {
		if _, ok := saleOrderProductMap[saleOrderProducts[i].Uuid]; !ok {
			// 如果商品不存在，则添加
			model.SaleOrderProducts = append(model.SaleOrderProducts, saleOrderProducts[i])
		} else {
			// 如果商品已存在，则更新
			saleOrderProductMap[saleOrderProducts[i].Uuid] = saleOrderProducts[i]
		}
	}

	// saleOrderProductMap中保存的是最新的商品，所以需要更新销售订单中的商品
	for i, saleOrderProduct := range model.SaleOrderProducts {
		model.SaleOrderProducts[i] = saleOrderProductMap[saleOrderProduct.Uuid]
	}
}

// 设置销售订单的序号
func (model *SaleOrder) SetIndex(index int) {
	model.index = index
}

// 设置销售订单的序号
func (model *SaleOrder) GetOrderName() string {
	if index := model.GetIndex(); index > 0 {
		return "-" + fmt.Sprintf("%d", index)
	}
	return ""
}

// 获取销售订单的顾客列表
func (model *SaleOrder) GetCustomerList() []resp.Product {
	productList := make([]resp.Product, 0)
	for _, orderBuffetCustomer := range model.SaleOrderBuffetCustomerTypes {
		if orderBuffetCustomer.IsDelete() {
			continue
		}
		// 自助餐顾客价格收费列表
		product := resp.Product{
			Uuid:       orderBuffetCustomer.Uuid,
			LocaleName: orderBuffetCustomer.BuffetPackage.MultiLanguageName.GetNames(),
			LocaleAttributeName: dto.LocaleResponse{
				ZH:   orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name,
				TH:   orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name,
				EN:   orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name,
				ZHTW: orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name,
				JA:   orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name,
				KO:   orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name,
				MY:   orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name,
				TR:   orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name,
				SV:   orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name,
			},
			Num:           float64(orderBuffetCustomer.Num), // 这种类型顾客多少个，如老人这个类型2人
			FinishedNum:   float64(orderBuffetCustomer.Num),
			SalePrice:     orderBuffetCustomer.GetOriginPrice(),
			DiscountPrice: orderBuffetCustomer.GetDiscountPrice(),
			TotalPrice:    orderBuffetCustomer.TotalPrice,
			Status:        1,
			Remark:        "",
			IsMust:        false,
			IsGift:        false,
			IsCancel:      false,
			IsBuffet:      false,
			AboutBuffet: resp.AboutBuffet{
				IsCustomer:       true,
				IsDelay:          false,
				CustomerTypeUuid: orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Uuid,
				BuffetUuid:       orderBuffetCustomer.BuffetPackageUuid,
			},
			SendKitchenTime: orderBuffetCustomer.CreateTime,
			Sign:            cryptor.Md5String(orderBuffetCustomer.GetSign()),
			UnitPrice:       orderBuffetCustomer.SalePrice,
			CanReturnNum:    orderBuffetCustomer.GetCanReturnNum(),
			CanReturnAmount: orderBuffetCustomer.GetCanReturnPrice(),
			CreateTime:      orderBuffetCustomer.CreateTime,
		}
		productList = append(productList, product)
	}
	return productList
}

// 获取销售订单的加钟商品列表
func (model *SaleOrder) GetDelayProductList() []resp.Product {
	productList := make([]resp.Product, 0)
	for _, delayProduct := range model.SaleOrderBuffetDelayProducts {
		if delayProduct.IsDelete() {
			continue
		}
		product := resp.Product{
			Uuid: delayProduct.Uuid,
			LocaleName: dto.LocaleResponse{
				ZH:   delayProduct.Name,
				TH:   delayProduct.Name,
				EN:   delayProduct.Name,
				ZHTW: delayProduct.Name,
				JA:   delayProduct.Name,
				KO:   delayProduct.Name,
				MY:   delayProduct.Name,
				TR:   delayProduct.Name,
				SV:   delayProduct.Name,
			},
			LocaleAttributeName: dto.LocaleResponse{},
			Num:                 float64(delayProduct.Num), // 拆单后不等于桌台人数，但同一个加钟商品的总数等于桌台人数
			FinishedNum:         float64(delayProduct.Num),
			SalePrice:           delayProduct.Price,
			DiscountPrice:       delayProduct.Price, // 加钟商品没有优惠价
			Status:              1,                  // 添加后标记送厨状态，不可修改
			Remark:              "",                 // 加钟商品没有备注
			IsMust:              false,
			IsGift:              false,
			IsCancel:            false,
			IsBuffet:            false,
			AboutBuffet: resp.AboutBuffet{
				IsCustomer: false,
				IsDelay:    true, // 标记该商品是加钟商品
			},
			SendKitchenTime: delayProduct.CreateTime,
			Sign:            cryptor.Md5String(delayProduct.GetSign()),
			UnitPrice:       delayProduct.Price,
			CanReturnNum:    delayProduct.GetCanReturnNum(),
			CanReturnAmount: delayProduct.GetCanReturnPrice(),
			CreateTime:      delayProduct.CreateTime,
		}
		productList = append(productList, product)
	}
	return productList
}

// 获取销售订单的商品列表
func (model *SaleOrder) GetProductList(hasOrderedH5ProductWithReject bool) []resp.Product {
	productList := make([]resp.Product, 0)
	for _, saleOrderProduct := range model.SaleOrderProducts {
		// 套餐子商品不返回
		if saleOrderProduct.ProductType == constant.ProductTypePackageSubProduct {
			continue
		}
		// 如果查询的是H5已下单的商品和被拒单的商品，则不跳过被删除的商品
		if !hasOrderedH5ProductWithReject {
			if saleOrderProduct.IsDelete() {
				continue
			}
		}

		sendKitchenTime := saleOrderProduct.SendKitchenTime
		if sendKitchenTime == 0 {
			sendKitchenTime = saleOrderProduct.CreateTime
		}
		var h5OrderTime int64
		var isH5NeedAudit bool
		if saleOrderProduct.H5Order != nil {
			h5OrderTime = saleOrderProduct.H5Order.CreateTime
			isH5NeedAudit = saleOrderProduct.H5Order.IsNeedAudit == 1
		}
		canChangeNum := true
		if saleOrderProduct.MustPlanUuid != 0 {
			canChangeNum = saleOrderProduct.ProductMustPlan.GetCanChangeNum()
		}
		// 套餐商品列表
		packageProductList := make([]resp.PackageProduct, 0)
		if saleOrderProduct.ProductType == constant.ProductTypePackage {
			subProductList := model.GetPackageSubProductList(saleOrderProduct.Uuid) // 获取套餐的子商品列表
			for _, subProduct := range subProductList {
				packageProductList = append(packageProductList, resp.PackageProduct{
					Uuid:                subProduct.Uuid,
					LocaleName:          subProduct.MultiLanguageName.GetNames(),
					LocaleAttributeName: subProduct.GetAttributeName(),
					Num:                 subProduct.Num,
					UnitNum:             subProduct.UnitNum,
				})
			}
		}

		product := resp.Product{
			Uuid:                saleOrderProduct.Uuid,
			LocaleName:          saleOrderProduct.MultiLanguageName.GetNames(),
			LocaleAttributeName: saleOrderProduct.GetAttributeName(),
			Num:                 saleOrderProduct.Num,
			NumType:             saleOrderProduct.NumType,
			SalePrice:           saleOrderProduct.GetSalePrice(),
			DiscountPrice:       saleOrderProduct.GetProductFinalSalePrice(),
			Status:              saleOrderProduct.StatusValue(),
			Remark:              saleOrderProduct.Remark,
			IsMust:              saleOrderProduct.IsMustProduct(),
			IsGift:              saleOrderProduct.IsGiftProduct(),
			IsWrap:              saleOrderProduct.IsWrapProduct(),
			IsBuffet:            saleOrderProduct.IsBuffetProduct(),
			IsCancel:            saleOrderProduct.IsCancelProduct(),
			CanChangeNum:        canChangeNum,
			SendKitchenTime:     sendKitchenTime,
			H5OrderTime:         h5OrderTime,
			IsH5OrderNeedAudit:  isH5NeedAudit,
			Sign:                cryptor.Md5String(saleOrderProduct.Sign),
			ProductPackageUuid:  saleOrderProduct.ProductPackageUuid,
			MustPlanUuid:        saleOrderProduct.MustPlanUuid,
			AcceptTime:          saleOrderProduct.GetAcceptTime(),
			IsAccept:            saleOrderProduct.IsAcceptOrderProduct(),
			UnitPrice:           saleOrderProduct.SalePrice,
			IsShowKitchen:       saleOrderProduct.ProductPackage.IsShowKitchen,
			CreateTime:          saleOrderProduct.CreateTime,
			ProductType:         saleOrderProduct.ProductPackage.ProductType,
			PackageProductList: resp.PackageProductList{
				List: packageProductList,
			},
			CanEdit: saleOrderProduct.IsCanEdit(),
		}
		if saleOrderProduct.ProductionOrderProduct != nil {
			if saleOrderProduct.ProductionOrderProduct.Status == constant.ProductionOrderProductStatusFinished {
				product.FinishedNum = saleOrderProduct.ProductionOrderProduct.Num
			}
		}
		productList = append(productList, product)
	}
	return productList
}

// 设置会员余额。如果订单设置了会员，则记录会员消费这笔订单后的余额。
func (model *SaleOrder) SetMemberBalance() {
	if model.ConsumerUuid == 0 {
		return
	}
	model.MemberBalance = model.Member.GetBalanceAll()
	model.MemberLevelName = model.Member.MemberLevel.Name
}

// 设置积分赠送比例
func (model *SaleOrder) SetGiftPointsRate(mealNum int, rule settingResp.PointsRule) {
	model.GiftPointsRate = rule.Value
	model.GiftPointsType = rule.GiftPointsType
	model.GiftPoints = model.CalcMemberPoint(mealNum, rule)
}

// CalcMemberPoint 计算会员积分. 会员积分=订单最终应收金额*积分赠送比例
func (model *SaleOrder) CalcMemberPoint(mealNum int, rule settingResp.PointsRule, finalPrice ...float64) float64 {
	// 如果订单使用了现金支付，且积分赠送规则中使用会员余额支付不赠送积分时，则不发放积分
	if model.HasBalancePayment() && !rule.BalancePaymentGetPoints {
		return 0
	}

	// 如果积分是按人数赠送的话
	if model.GiftPointsType == 1 {
		// AC17:如果订单不是主单，则不发放积分.
		if model.GetIndex() > 1 {
			return 0
		}
		return decimal.NewFromInt(int64(mealNum)).Mul(decimal.NewFromFloat(model.GiftPointsRate)).Round(2).InexactFloat64()
	}

	// 如果积分是按比例赠送的话
	baseNum := model.GetFinalNoFeeAmount()
	if len(finalPrice) > 0 {
		baseNum = finalPrice[0]
	}
	// 如果订单时按比例赠送，要检查订单最终应收金额是否达到阀值，达到才送积分。未达到，不送积分
	if rule.GiftPointsType == 0 && baseNum < rule.PaymentAmountRequirement {
		return 0
	}
	return decimal.NewFromFloat(baseNum).Mul(decimal.NewFromFloat(model.GiftPointsRate)).Round(2).InexactFloat64()
}

// 发放消费积分
func (model *SaleOrder) HandleMemberPoints(member *Member) {
	model.Member = member // 使用最新的会员信息。避免该会员的积分信息已经被更新过
	// 如果开启积分赠送且赠送比例大于0，则发放积分
	if model.GiftPointsRate > 0 {
		// 发放积分
		member.ChangePoint(model.GiftPoints) // 增加积分
	}
}

// 创建积分变动记录
func (model *SaleOrder) NewMemberPointLog() *MemberPointLog {
	memberPointLog := &MemberPointLog{
		MemberUuid:  model.ConsumerUuid,
		Scene:       constant.MemberPointLogSceneConsume,
		Value:       model.GiftPoints,
		Describe:    fmt.Sprintf("订单赠送：%s", model.OrderNo),
		RelatedUuid: model.Uuid,
	}
	return memberPointLog
}

// 创建退款积分变动记录
func (model *SaleOrder) NewRefundMemberPointLog(points float64) *MemberPointLog {
	memberPointLog := &MemberPointLog{
		MemberUuid:  model.ConsumerUuid,
		Scene:       constant.MemberPointLogSceneRefund,
		Value:       points,
		Describe:    fmt.Sprintf("退款扣除：%s", model.OrderNo),
		RelatedUuid: model.Uuid,
	}
	return memberPointLog
}

// 创建订单反结账积分变动记录
func (model *SaleOrder) NewReverseSettleMemberPointLog(points float64) *MemberPointLog {
	memberPointLog := &MemberPointLog{
		MemberUuid:  model.ConsumerUuid,
		Scene:       constant.MemberPointLogSceneReverse,
		Value:       points,
		Describe:    fmt.Sprintf("订单反结账：%s", model.OrderNo),
		RelatedUuid: model.Uuid,
	}
	return memberPointLog
}

// 创建订单反结账退回已抵扣积分变动记录
func (model *SaleOrder) NewReverseSettleExchangeMemberPointLog(points float64) *MemberPointLog {
	memberPointLog := &MemberPointLog{
		MemberUuid:  model.ConsumerUuid,
		Scene:       constant.MemberPointLogScenePointsExchangeReverse,
		Value:       points,
		Describe:    fmt.Sprintf("抵扣反结账：%s", model.OrderNo),
		RelatedUuid: model.Uuid,
	}
	return memberPointLog
}

// 创建退货单
func (model *SaleOrder) NewReturnOrder(scene string, deliveryFee float64, dutyNo string, lang string, saleOrderProducts []*SaleOrderProduct, buffetCustomers []*SaleOrderBuffetCustomerType, buffetDelays []*SaleOrderBuffetDelayProduct, numMap map[uint64]float64, returnType int, canReturnAmount float64) (*ReturnOrder, error) {
	returnOrderUuid, _ := utils.GetID()

	// 如果退款类型为整单退款，则退款金额=订单最终应收金额-已退款金额
	// 如果退款类型为部分退款，则退款金额=退货单商品总金额之和
	returnAmount := decimal.NewFromFloat(0) // 本次退款操作的退款金额。退款金额=退货单商品总金额之和
	returnOrderProducts := make([]*ReturnOrderProduct, 0)
	for _, saleOrderProduct := range saleOrderProducts {
		// 退货数量
		num := numMap[saleOrderProduct.Uuid]
		// 如果退货数量为0，则跳过
		if num == 0 {
			continue
		}
		// 商品总金额=退货商品数量*商品最终单价
		productTotalAmount := decimal.NewFromFloat(saleOrderProduct.TotalPrice).Mul(decimal.NewFromFloat(num))
		returnOrderProducts = append(returnOrderProducts, &ReturnOrderProduct{
			SaleOrderUuid:        model.Uuid,
			SaleOrderProductUuid: saleOrderProduct.Uuid,
			ReturnOrderUuid:      returnOrderUuid,
			ProductType:          constant.ReturnOrderProductTypeSaleOrderProduct,
			ProductPackageUuid:   saleOrderProduct.ProductPackageUuid,
			ProductName:          saleOrderProduct.Name,
			ProductPrice:         saleOrderProduct.Price,
			TaxRate:              saleOrderProduct.TaxRate,
			Num:                  num,
			ProductTotalAmount:   productTotalAmount.Round(2).InexactFloat64(), // 商品总金额=退货商品数量*商品最终单价
			ErpCode:              saleOrderProduct.ErpCode,
		})
		returnAmount = returnAmount.Add(productTotalAmount)
	}
	for _, saleOrderProduct := range buffetCustomers {
		// 退货数量
		num := numMap[saleOrderProduct.Uuid]
		// 如果退货数量为0，则跳过
		if num == 0 {
			continue
		}
		// 商品总金额=退货商品数量*商品最终单价
		productTotalAmount := decimal.NewFromFloat(saleOrderProduct.TotalPrice).Mul(decimal.NewFromFloat(num))
		returnOrderProducts = append(returnOrderProducts, &ReturnOrderProduct{
			SaleOrderUuid:        model.Uuid,
			SaleOrderProductUuid: saleOrderProduct.Uuid,
			ReturnOrderUuid:      returnOrderUuid,
			ProductType:          constant.ReturnOrderProductTypeSaleOrderBuffetCustomer,
			ProductPackageUuid:   saleOrderProduct.BuffetPackageUuid,
			ProductName:          saleOrderProduct.Name,
			ProductPrice:         saleOrderProduct.Price,
			TaxRate:              saleOrderProduct.TaxRate,
			Num:                  num,
			ProductTotalAmount:   productTotalAmount.Round(2).InexactFloat64(), // 商品总金额=退货商品数量*商品最终单价
		})
		returnAmount = returnAmount.Add(productTotalAmount)
	}
	for _, saleOrderProduct := range buffetDelays {
		// 退货数量
		num := numMap[saleOrderProduct.Uuid]
		// 如果退货数量为0，则跳过
		if num == 0 {
			continue
		}
		// 商品总金额=退货商品数量*商品最终单价
		productTotalAmount := decimal.NewFromFloat(saleOrderProduct.Price).Mul(decimal.NewFromFloat(num))
		returnOrderProducts = append(returnOrderProducts, &ReturnOrderProduct{
			SaleOrderUuid:        model.Uuid,
			SaleOrderProductUuid: saleOrderProduct.Uuid,
			ReturnOrderUuid:      returnOrderUuid,
			ProductType:          constant.ReturnOrderProductTypeBuffetAddTimeProduct,
			ProductPackageUuid:   saleOrderProduct.BuffetDelayUuid,
			ProductName:          saleOrderProduct.Name,
			ProductPrice:         saleOrderProduct.Price,
			Num:                  num,
			ProductTotalAmount:   productTotalAmount.Round(2).InexactFloat64(), // 商品总金额=退货商品数量*商品最终单价
		})
		returnAmount = returnAmount.Add(productTotalAmount)
	}

	// 退货金额=退货商品总金额之和
	refundAmount := returnAmount.Round(2).InexactFloat64()
	if returnType == constant.ReturnOrderRefundTypeTotal {
		// 整单退款，退款金额=订单最终应收金额-已退款金额
		refundAmount = decimal.NewFromFloat(model.FinalPrice).Sub(decimal.NewFromFloat(model.GetReturnAmount())).Round(2).InexactFloat64()
		// 如果是会员端订单，则退款金额=订单最终应收金额-已退款金额-配送费
		if scene == constant.SceneMemberOrder {
			refundAmount = decimal.NewFromFloat(refundAmount).Sub(decimal.NewFromFloat(deliveryFee)).Round(2).InexactFloat64()
		}
	}
	totalRefundAmount := refundAmount

	// 退款金额不能大于可退金额
	if totalRefundAmount > canReturnAmount {
		return nil, errors.WithMessage(errors.New(i18n.Translate(lang, "退款金额不能大于可退金额") + fmt.Sprintf(" %v", canReturnAmount)))
	}

	// 获取销售订单的每个付款单的可退款金额
	paymentRecords, currencyUnit := model.GetPaymentOrderCanReturnAmount()

	// 退款金额. 按照产品规格的顺序依次原路退款
	// 退款顺序优先退会员、不够退则到现金、再到记录支付（多个时，哪个先后都行）、再到lianlian（多个时，哪个先后都行）
	returnOrderAmounts := make([]ReturnOrderAmount, 0)
	for _, paymentOrder := range paymentRecords {
		amount := float64(0)
		// 如果退款金额大于付款单的可退款金额，则该退款单的退款金额=付款单的可退款金额
		if refundAmount > paymentOrder.CanReturnAmount {
			amount = paymentOrder.CanReturnAmount
		}
		// 如果退款金额小于或等于付款单的可退款金额，则该退款单的退款金额=退款金额
		if refundAmount <= paymentOrder.CanReturnAmount {
			amount = refundAmount
		}
		// 创建退款金额记录
		returnOrderAmountUuid, _ := utils.GetID()
		returnOrderAmount := ReturnOrderAmount{
			BaseModel: BaseModel{
				Uuid: returnOrderAmountUuid,
			},
			ReturnOrderUuid:       returnOrderUuid,
			PaymentMethodUuid:     paymentOrder.PaymentMethodUuid,
			PaymentOrderUuid:      paymentOrder.PaymentOrderUuid,
			Amount:                amount,
			MerchantRefundOrderNo: utils.GenerateMerchantOrderNo("PS"),
			PaymentMethod: &PaymentMethod{
				PaymentName: paymentOrder.PaymentMethodName,
				Code:        paymentOrder.PaymentMethodCode,
			},
		}
		// 如果退款金额为余额，则创建余额变动记录
		if returnOrderAmount.PaymentMethod.Code == constant.PaymentMethodCodeBalance {
			returnOrderAmount.MemberBalanceLog = &MemberBalanceLog{
				MemberUuid:  model.ConsumerUuid,
				Scene:       constant.MemberPointLogSceneRefund,
				Money:       0,
				GiftMoney:   returnOrderAmount.Amount,
				Describe:    fmt.Sprintf("订单退款：%s", model.OrderNo),
				RelatedUuid: returnOrderAmountUuid,
			}
		}

		// 如果退款金额为现金，则更新钱箱
		if returnOrderAmount.PaymentMethod.Code == constant.PaymentMethodCodeCash {
			returnOrderAmount.CashBoxLog = &CashBoxLog{
				Scene:                 constant.CashBoxLogSceneRefund,
				Amount:                returnOrderAmount.Amount,
				Remark:                fmt.Sprintf("订单退款：%s", model.OrderNo),
				RelatedUuid:           model.Uuid,            // 关联销售订单
				ReturnOrderUuid:       returnOrderUuid,       // 关联退货单
				RefundOrderAmountUuid: returnOrderAmountUuid, // 关联退货单退款金额
			}
		}

		returnOrderAmounts = append(returnOrderAmounts, returnOrderAmount)
		refundAmount = refundAmount - amount
		// 如果退款金额为0，则退出. 表示本次要退款的金额已经全都分到不同的退款渠道中
		if refundAmount <= 0 {
			break
		}
	}
	return &ReturnOrder{
		BaseModel: BaseModel{
			Uuid:       returnOrderUuid,
			CreateTime: time.Now().Unix(),
		},
		RelatedOrderType:    constant.ReturnOrderRelatedOrderTypeSaleOrder,
		RelatedOrderUuid:    model.Uuid,
		RelatedOrderNo:      model.OrderNo,
		ReturnType:          uint(returnType),
		RefundAmount:        totalRefundAmount,
		Unit:                currencyUnit,
		RefundTaxAmount:     model.TaxFee,
		RefundReason:        "退款",
		ReturnOrderAmounts:  returnOrderAmounts,
		ReturnOrderProducts: returnOrderProducts,
		DutyNo:              dutyNo,
	}, nil
}

func (model *SaleOrder) NewSaleOrderBuffetCustomerType(buffetPackageUuid, buffetCustomerTypePriceUuid uint64, customerNum uint, buffetCustomerTypePricePrice float64, buffetPackageTaxRate float64, setting SaleBillSetting) *SaleOrderBuffetCustomerType {
	saleOrderBuffetCustomerType := &SaleOrderBuffetCustomerType{
		SaleOrderUuid:               model.Uuid,
		SaleBillUuid:                model.SaleBillUuid,
		BuffetPackageUuid:           buffetPackageUuid,
		BuffetCustomerTypePriceUuid: buffetCustomerTypePriceUuid,
		Num:                         customerNum,
		SalePrice:                   buffetCustomerTypePricePrice,
		TaxRate:                     buffetPackageTaxRate,
		CustomDiscountRate:          model.CustomDiscountRate, // 跟随订单的自定义折扣
	}
	// 计算金额
	saleOrderBuffetCustomerType.CalcSaleOrderBuffetCustomerType(setting)
	//
	return saleOrderBuffetCustomerType
}

func (model *SaleOrder) NewFreeOrderReason(freeReasons []*FreeReason) []*SaleOrderProductReason {
	list := make([]*SaleOrderProductReason, 0)
	for _, reason := range freeReasons {
		reasonUuid, _ := utils.GetID()
		list = append(list, &SaleOrderProductReason{
			BaseModel: BaseModel{
				Uuid: reasonUuid,
			},
			SaleOrderUuid:         model.Uuid,
			MultiLanguageNameUuid: reason.MultiLanguageNameUuid,
			FreeReasonUuid:        reason.Uuid,
		})
	}
	return list
}

// 判断销售订单是否部分支付
func (model *SaleOrder) IsPartialPay() bool {
	num := 0
	for _, paymentOrder := range model.PaymentOrders {
		if !paymentOrder.IsDelete() {
			num++
		}
	}
	return num > 0
}

// 判断销售订单是否已支付
func (model *SaleOrder) IsPaid() bool {
	return model.Status == constant.SaleOrderStatusFinish
}

// IsFreeSaleOrder 判断销售订单是否免单
func (model *SaleOrder) IsFreeSaleOrder() bool {
	return model.IsFree == constant.SaleOrderIsFreeYes
}

// erp是否进行过反结账
func (model *SaleOrder) IsErpReverseSettle() bool {
	// 结账后saleOrder中就会记录下这两个发票名称
	return model.ErpProductsInvoiceName != "" && model.ErpMaterialInvoiceName != ""
}

// TableName 指定表名
func (model *SaleOrder) TableName() string {
	return "ttpos_sale_order"
}

// ValidateOrderStatus 判断订单是否可操作
func (model *SaleOrder) ValidateOrderStatus() error {
	if model.Status == constant.SaleBillStatusCanceled {
		return errors.New("订单已取消")
	}
	if model.Status == constant.SaleBillStatusComplete {
		return errors.New("订单已结账")
	}
	return nil
}

func NewSaleOrder(deviceId string, saleBillUuid uint64, saleBillOrderNo string, setting SaleBillSetting) *SaleOrder {
	uuid, _ := utils.GetID()
	saleOrder := &SaleOrder{
		BaseModel:    BaseModel{Uuid: uuid},
		SaleBillUuid: saleBillUuid,
		OrderNo:      saleBillOrderNo,
		DeviceId:     deviceId,
	}
	// 设置服务费初始值
	saleOrder.SetInitServiceFee(setting)
	// 设置默认的订单抹零规格、结账抹零规则
	saleOrder.ZeroRule = uint8(setting.ZeroRule)
	saleOrder.ZeroCheckoutRule = uint8(setting.ZeroCheckoutRule)
	// 积分抵扣类型
	if setting.IsOpenPointsExchange() {
		saleOrder.AutoPointsExchange = setting.AutoPointsExchange
		saleOrder.PointsExchangeRate = setting.PointsExchangeRate
	}

	return saleOrder
}

func NewSaleOrderBuffetCustomerType(customerName string, saleOrderUuid, saleBillUuid, buffetPackageUuid, buffetCustomerTypePriceUuid uint64, customerNum uint, buffetCustomerTypePricePrice float64, buffetPackageTaxRate float64, setting SaleBillSetting, openOverallDiscount uint) *SaleOrderBuffetCustomerType {
	saleOrderBuffetCustomerType := &SaleOrderBuffetCustomerType{
		Name:                        customerName,
		SaleOrderUuid:               saleOrderUuid,
		SaleBillUuid:                saleBillUuid,
		BuffetPackageUuid:           buffetPackageUuid,
		BuffetCustomerTypePriceUuid: buffetCustomerTypePriceUuid,
		Num:                         customerNum,
		SalePrice:                   buffetCustomerTypePricePrice,
		Price:                       buffetCustomerTypePricePrice,
		TaxRate:                     buffetPackageTaxRate,
		CustomDiscountRate:          1,                   // 默认自定义折扣率为1，即不打折。刚开始创建时是没有折扣的
		OpenOverallDiscount:         openOverallDiscount, // 默认开启整单折扣
	}
	// 计算金额
	saleOrderBuffetCustomerType.CalcSaleOrderBuffetCustomerType(setting)
	//
	return saleOrderBuffetCustomerType
}

// GetBuffetUuidMap 获取处理包
type BuffetUuidMapBuffetCustomerTypes struct {
	Uuid    uint64
	MealNum *uint
}

type FinalAmount struct {
	PaymentAmount        float64 // 已支付的金额
	ChangeAmount         float64 // 找零金额
	ZeroCheckoutFee      float64 // 结账抹零金额
	FinalPrice           float64 // 最终应收金额
	PaymentCommissionFee float64 // 支付手续费
	GiftAmount           float64 // 赠菜金额
	Unit                 string  // 金额的单位,$-美元 ￥-人民币,用于显示订单金额价值
}
