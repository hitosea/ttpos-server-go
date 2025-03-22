package model

import (
	"fmt"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/pkg/utils"

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
	Amount       float64 `gorm:"column:amount;type:decimal(12,2);default:0;comment:应收金额。商品未含税时，总金额=商品金额+服务费+税费。商品已含税时，总金额=商品金额（含商品消费税）+服务费+税费（只有服务费税）" json:"amount"`
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

	// 结账完成后才记录的字段
	PaymentAmount        float64 `gorm:"column:payment_amount;type:decimal(12,2);default:0;comment:支付金额,支付金额-订单总金额=支付手续费" json:"payment_amount"`
	ChangeAmount         float64 `gorm:"column:change_amount;type:decimal(12,2);default:0;comment:找零金额,结账完成后才记录" json:"change_amount"`
	ZeroCheckoutFee      float64 `gorm:"column:zero_checkout_fee;type:decimal(12,2);default:0;comment:结账抹零金额" json:"zero_checkout_fee"`
	FinalPrice           float64 `gorm:"column:final_price;type:decimal(12,2);default:0;comment:最终应收金额。最终应收金额=应收金额+手续费-结账抹零金额" json:"final_price"`
	PaymentCommissionFee float64 `gorm:"column:payment_commission_fee;type:decimal(12,2);default:0;comment:支付手续费,关联付款单的支付手续费之和" json:"payment_commission_fee"`
	GiftAmount           float64 `gorm:"column:gift_amount;type:decimal(12,2);default:0;comment:赠菜金额,(销售订单赠菜商品.总最终单价)之和" json:"gift_amount"`
	GiftPoints           float64 `gorm:"column:gift_points;type:decimal(12,2);default:0;comment:赠送积分,应收金额amount*积分赠送比例" json:"gift_points"`
	GiftPointsRate       float64 `gorm:"column:gift_points_rate;type:decimal(12,4);default:0;comment:赠送积分比例,取值范围0-1。结账后记录，不受后台改变" json:"gift_points_rate"`
	MemberBalance        float64 `gorm:"column:member_balance;type:decimal(12,2);default:0;comment:会员余额,会员消费本单后剩余的余额" json:"member_balance"`

	// 虚拟字段，用于标记当前子单是第几个
	Index int `gorm:"-" json:"index,omitempty"`

	// 关联对象
	PaymentOrders                []*PaymentOrder                `gorm:"foreignKey:RelatedUuid;references:uuid"`
	Member                       *Member                        `gorm:"foreignKey:ConsumerUuid;references:uuid"`
	SaleOrderProducts            []*SaleOrderProduct            `gorm:"foreignKey:SaleOrderUuid;references:uuid"`
	ReturnOrders                 []ReturnOrder                  `gorm:"foreignKey:RelatedOrderUuid;references:uuid"`
	SaleOrderBuffetCustomerTypes []*SaleOrderBuffetCustomerType `gorm:"foreignKey:SaleOrderUuid;references:uuid"`
	SaleOrderBuffetDelayProducts []*SaleOrderBuffetDelayProduct `gorm:"foreignKey:SaleOrderUuid;references:uuid"`
	FreeReasons                  []*SaleOrderProductReason      `gorm:"foreignKey:SaleOrderUuid;references:uuid"`
	InvoiceInfo                  *SaleOrderInvoiceInfo          `gorm:"foreignKey:SaleOrderUuid;references:uuid"`
	SaleBill                     *SaleBill                      `gorm:"foreignKey:SaleBillUuid;references:uuid"`
	MemberPointLog               *MemberPointLog                `gorm:"foreignKey:RelatedUuid;references:uuid"` // 关联积分变动记录.赠送积分
}

// 设置会员余额。如果订单设置了会员，则记录会员消费这笔订单后的余额。
func (model *SaleOrder) SetMemberBalance() {
	if model.ConsumerUuid == 0 {
		return
	}
	model.MemberBalance = model.Member.GetBalanceAll()
}

// 设置积分赠送比例
func (model *SaleOrder) SetGiftPointsRate(giftPointsRate float64) {
	model.GiftPointsRate = giftPointsRate
	model.GiftPoints = model.CalcMemberPoint()
}

// CalcMemberPoint 计算会员积分. 会员积分=订单应收金额*积分赠送比例
func (model *SaleOrder) CalcMemberPoint() float64 {
	return decimal.NewFromFloat(model.Amount).Mul(decimal.NewFromFloat(model.GiftPointsRate)).InexactFloat64()
}

// 发放消费积分
func (model *SaleOrder) HandleMemberPoints(member *Member) {
	model.Member = member // 使用最新的会员信息。避免该会员的积分信息已经被更新过
	// 如果开启积分赠送且赠送比例大于0，则发放积分
	if model.GiftPointsRate > 0 {
		// 发放积分
		member.ChangePoint(model.GiftPoints) // 增加积分
	}
	// 创建积分变动记录
	model.MemberPointLog = model.NewMemberPointLog()
}

// 累计会员的消费金额、消费次数
func (model *SaleOrder) AccumulateMemberConsumeAmountAndTimes(member *Member) {
	model.Member = member // 使用最新的会员信息。避免该会员的积分信息已经被更新过
	model.Member.AccumulateConsumeAmount(model.Amount)
}

func (model *SaleOrder) NewMemberPointLog() *MemberPointLog {
	memberPointLog := &MemberPointLog{
		MemberUuid:  model.Member.Uuid,
		Scene:       constant.MemberPointLogSceneConsume,
		Value:       model.GiftPoints,
		Describe:    fmt.Sprintf("订单赠送：%s", model.OrderNo),
		RelatedUuid: model.Uuid,
	}
	return memberPointLog
}

// 创建退货单
func (model *SaleOrder) NewReturnOrder(saleOrderProducts []*SaleOrderProduct, numMap map[uint64]uint, returnType int) *ReturnOrder {
	returnOrderUuid, _ := utils.GetID()

	// 如果退款类型为整单退款，则退款金额=订单最终应收金额-已退款金额
	// 如果退款类型为部分退款，则退款金额=退货单商品总金额之和
	returnAmount := decimal.NewFromFloat(0) // 本次退款操作的退款金额。退款金额=退货单商品总金额之和
	returnOrderProducts := make([]*ReturnOrderProduct, 0)
	for _, saleOrderProduct := range saleOrderProducts {
		// 退货数量
		num := numMap[saleOrderProduct.Uuid]
		// 商品总金额=退货商品数量*商品最终单价
		productTotalAmount := decimal.NewFromFloat(saleOrderProduct.Price).Mul(decimal.NewFromInt(int64(num)))
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
			ProductTotalAmount:   productTotalAmount.InexactFloat64(), // 商品总金额=退货商品数量*商品最终单价
		})
		returnAmount = returnAmount.Add(productTotalAmount)
	}
	refundAmount := returnAmount.Round(2).InexactFloat64()

	if returnType == constant.ReturnOrderRefundTypeTotal {
		// 整单退款，退款金额=订单最终应收金额-已退款金额
		refundAmount = decimal.NewFromFloat(model.FinalPrice).Sub(decimal.NewFromFloat(model.GetReturnAmount())).Round(2).InexactFloat64()
	}

	// 获取销售订单的每个付款单的可退款金额
	paymentRecords, currencyUnit := model.GetPaymentOrderCanReturnAmount()

	// 退款金额. 按照产品规格的顺序依次原路退款
	// 退款顺序优先退会员、不够退则到现金、再到记录支付（多个时，哪个先后都行）、再到lianlian（多个时，哪个先后都行）
	returnOrderAmounts := make([]ReturnOrderAmount, 0)
	for _, paymentOrder := range paymentRecords {
		amount := decimal.NewFromFloat(0)
		// 如果退款金额大于付款单的可退款金额，则该退款单的退款金额=付款单的可退款金额
		if returnAmount.InexactFloat64() > paymentOrder.CanReturnAmount {
			amount = decimal.NewFromFloat(paymentOrder.CanReturnAmount)
		}
		// 如果退款金额小于或等于付款单的可退款金额，则该退款单的退款金额=退款金额
		if returnAmount.InexactFloat64() <= paymentOrder.CanReturnAmount {
			amount = returnAmount
		}
		returnOrderAmounts = append(returnOrderAmounts, ReturnOrderAmount{
			ReturnOrderUuid:       returnOrderUuid,
			PaymentMethodUuid:     paymentOrder.PaymentMethodUuid,
			PaymentOrderUuid:      paymentOrder.PaymentOrderUuid,
			Amount:                amount.InexactFloat64(),
			MerchantRefundOrderNo: utils.GenerateMerchantOrderNo("PS"),
			PaymentMethod:         &PaymentMethod{Code: paymentOrder.PaymentMethodCode},
		})
		returnAmount = returnAmount.Sub(amount)
		// 如果退款金额为0，则退出
		if returnAmount.InexactFloat64() <= 0 {
			break
		}
	}
	return &ReturnOrder{
		BaseModel: BaseModel{
			Uuid: returnOrderUuid,
		},
		RelatedOrderType:    constant.ReturnOrderRelatedOrderTypeSaleOrder,
		RelatedOrderUuid:    model.Uuid,
		RelatedOrderNo:      model.OrderNo,
		ReturnType:          uint(returnType),
		RefundAmount:        refundAmount,
		Unit:                currencyUnit,
		RefundTaxAmount:     model.TaxFee,
		RefundReason:        "退款",
		ReturnOrderAmounts:  returnOrderAmounts,
		ReturnOrderProducts: returnOrderProducts,
	}
}

func (model *SaleOrder) NewSaleOrderBuffetCustomerType(buffetPackageUuid, buffetCustomerTypePriceUuid uint64, customerNum uint, buffetCustomerTypePricePrice float64, buffetPackageTaxRate float64, setting SaleBillSetting) *SaleOrderBuffetCustomerType {
	saleOrderBuffetCustomerType := &SaleOrderBuffetCustomerType{
		SaleOrderUuid:               model.Uuid,
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
	return len(model.PaymentOrders) > 1
}

// 判断销售订单是否已支付
func (model *SaleOrder) IsPaid() bool {
	return model.Status == constant.SaleOrderStatusFinish
}

// IsFreeSaleOrder 判断销售订单是否免单
func (model *SaleOrder) IsFreeSaleOrder() bool {
	return model.IsFree == constant.SaleOrderIsFreeYes
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

func NewSaleOrder(saleBillUuid uint64, saleBillOrderNo string, setting SaleBillSetting) *SaleOrder {
	uuid, _ := utils.GetID()
	saleOrder := &SaleOrder{
		BaseModel:    BaseModel{Uuid: uuid},
		SaleBillUuid: saleBillUuid,
		OrderNo:      saleBillOrderNo,
	}
	// 设置服务费初始值
	saleOrder.SetInitServiceFee(setting)
	return saleOrder
}

func NewSaleOrderBuffetCustomerType(saleOrderUuid, buffetPackageUuid, buffetCustomerTypePriceUuid uint64, customerNum uint, buffetCustomerTypePricePrice float64, buffetPackageTaxRate float64, setting SaleBillSetting) *SaleOrderBuffetCustomerType {
	saleOrderBuffetCustomerType := &SaleOrderBuffetCustomerType{
		SaleOrderUuid:               saleOrderUuid,
		BuffetPackageUuid:           buffetPackageUuid,
		BuffetCustomerTypePriceUuid: buffetCustomerTypePriceUuid,
		Num:                         customerNum,
		SalePrice:                   buffetCustomerTypePricePrice,
		Price:                       buffetCustomerTypePricePrice,
		TaxRate:                     buffetPackageTaxRate,
		CustomDiscountRate:          1, // 默认自定义折扣率为1，即不打折。刚开始创建时是没有折扣的
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
}
