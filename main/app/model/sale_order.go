package model

import (
	"fmt"
	"slices"
	"sort"
	"strings"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/i18n"
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
}

// 获取销售订单的每个付款单的可退款金额。
// 要求排好序：退款顺序优先退会员、不够退则到现金、再到记录支付（多个时，哪个先后都行）、再到lianlian（多个时，哪个先后都行）
func (model *SaleOrder) GetPaymentOrderCanReturnAmount() ([]resp.OrderReturnPaymentRecord, string) {
	paymentRecords := make([]resp.OrderReturnPaymentRecord, 0)
	currencyUnit := ""
	for _, paymentOrder := range model.PaymentOrders {
		paymentRecords = append(paymentRecords, resp.OrderReturnPaymentRecord{
			PaymentOrderUuid:  paymentOrder.Uuid,
			PaymentMethodName: paymentOrder.PaymentMethodName,
			PaymentMethodUuid: paymentOrder.PaymentMethodUuid,
			CurrencyUnit:      paymentOrder.CurrencyUnit,
			PaymentAmount:     paymentOrder.Amount,
			CanReturnAmount:   paymentOrder.GetCanReturnAmount(), // 可退款金额=支付金额-已退款金额
			PaymentMethodCode: paymentOrder.PaymentMethod.Code,
		})
		currencyUnit = paymentOrder.CurrencyUnit
	}
	// 排序。code越小，越靠前
	sort.Slice(paymentRecords, func(i, j int) bool {
		return paymentRecords[i].PaymentMethodCode < paymentRecords[j].PaymentMethodCode
	})
	return paymentRecords, currencyUnit
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
			ReturnOrderUuid:   returnOrderUuid,
			PaymentMethodUuid: paymentOrder.PaymentMethodUuid,
			PaymentOrderUuid:  paymentOrder.PaymentOrderUuid,
			Amount:            amount.InexactFloat64(),
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

// GetReturnAmount 获取销售订单的退款金额. 退款金额=所有退货单的退款金额之和
func (model *SaleOrder) GetReturnAmount() float64 {
	amount := decimal.NewFromFloat(0)
	for _, returnOrder := range model.ReturnOrders {
		amount = amount.Add(decimal.NewFromFloat(returnOrder.RefundAmount))
	}
	return amount.InexactFloat64()
}

// GetCanReturnAmount 获取销售订单的可退款金额. 可退款金额=订单最终应收金额-已退款金额
func (model *SaleOrder) GetCanReturnAmount() float64 {
	return decimal.NewFromFloat(model.PaymentAmount).Sub(decimal.NewFromFloat(model.GetReturnAmount())).Round(2).InexactFloat64()
}

// GetOriginAmount 获取订单没打折之前的订单应收金额。原订单应收金额=现应收金额+会员折扣金额+优惠折扣金额
func (model *SaleOrder) GetOriginAmount() float64 {
	//原订单应收金额=现应收金额+会员折扣金额+优惠折扣金额
	return decimal.NewFromFloat(model.Amount).Add(decimal.NewFromFloat(model.MemberDiscountFee)).Add(decimal.NewFromFloat(model.CustomDiscountFee)).Round(2).InexactFloat64()
}

// GetMemberDiscountAmount 获取订单的会员折扣后应收金额。 会员折扣后应收金额=原订单应收金额-会员折扣金额
func (model *SaleOrder) GetMemberDiscountAmount() float64 {
	//会员折扣后应收金额=现应收金额+会员折扣金额
	return decimal.NewFromFloat(model.GetOriginAmount()).Sub(decimal.NewFromFloat(model.MemberDiscountFee)).Round(2).InexactFloat64()
}

// GetMemberName 获取订单的会员名称
func (model *SaleOrder) GetMemberName() string {
	if model.Member == nil {
		return ""
	}
	return model.Member.Nickname
}

// CalcGiftAmount 计算赠菜金额. 赠菜金额=销售订单赠菜商品.总最终单价之和
func (model *SaleOrder) CalcGiftAmount(options ...func(option *CalcOption)) float64 {
	option := &CalcOption{}
	for _, optionFunc := range options {
		optionFunc(option)
	}
	amount := float64(0)
	for _, saleOrderProduct := range model.SaleOrderProducts {
		// 不是赠菜的商品不计入
		if !saleOrderProduct.IsGiftBool() {
			continue
		}
		// 删除的商品不计、退菜的商品不计入。 未送厨的商品也要计入
		if saleOrderProduct.IsDelete() || saleOrderProduct.IsCancelProduct() {
			continue
		}
		if option.IsCooking {
			// 未送厨的商品不计入
			if !saleOrderProduct.IsCookingProduct() {
				continue
			}
		}
		// 商品的最终金额
		giftFee := saleOrderProduct.GetPrice()
		// 累计各个赠品的最终金额
		amount = decimal.NewFromFloat(amount).Add(decimal.NewFromFloat(giftFee)).InexactFloat64()
	}
	return amount
}

// SaleOrderBuffetDelayTimeTotal 总加钟时间
func (model *SaleOrder) SaleOrderBuffetDelayTimeTotal() int64 {
	delayTime := int64(0)
	for _, saleOrderProduct := range model.SaleOrderBuffetDelayProducts {
		// 商品的加钟时间
		delayTime += saleOrderProduct.DelayTime
	}
	return delayTime
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

func (model *SaleOrder) SetCheckoutZeroingMethod(zeroRule int) {
	model.ZeroCheckoutRule = uint8(zeroRule)
}

// 设置初始时销售订单的服务费。
// 当关闭服务费费时，订单服务费=0
// 当开启服务费按固定服务费收费时， 订单服务费=固定金额
// 当开启服务费按比例收取服务费时，订单服务费=各个订单商品的服务费之和。初始时订单服务费=0，在添加商品后再重建计算
func (model *SaleOrder) SetInitServiceFee(setting SaleBillSetting) float64 {
	// 当开启服务费按固定服务费收费时， 订单服务费=固定金额
	if setting.GetServiceFeeType() == constant.SaleBillSettingServiceFeeTypeFixed {
		return setting.ServiceFeeValue
	}
	// 其他情况，初始化时服务费都是0
	return 0
}

// 判断销售订单是否部分支付
func (model *SaleOrder) IsPartialPay() bool {
	return len(model.PaymentOrders) > 1
}

// 判断销售订单是否已支付
func (model *SaleOrder) IsPaid() bool {
	return model.Status == constant.SaleOrderStatusFinish
}

type FinalAmount struct {
	PaymentAmount        float64 // 已支付的金额
	ChangeAmount         float64 // 找零金额
	ZeroCheckoutFee      float64 // 结账抹零金额
	FinalPrice           float64 // 最终应收金额
	PaymentCommissionFee float64 // 支付手续费
	GiftAmount           float64 // 赠菜金额
}

func (model *SaleOrder) SetFinishStatus(final FinalAmount) {
	// 修改状态
	model.Status = constant.SaleOrderStatusFinish
	model.FinishTime = time.Now().Unix()
	// 更新订单结算后要计算的金额字段
	model.PaymentAmount = final.PaymentAmount
	model.ChangeAmount = final.ChangeAmount
	model.ZeroCheckoutFee = final.ZeroCheckoutFee
	model.FinalPrice = final.FinalPrice
	model.PaymentCommissionFee = final.PaymentCommissionFee
	model.GiftAmount = final.GiftAmount

}

// IsFreeSaleOrder 判断销售订单是否免单
func (model *SaleOrder) IsFreeSaleOrder() bool {
	return model.IsFree == constant.SaleOrderIsFreeYes
}

// TableName 指定表名
func (model *SaleOrder) TableName() string {
	return "ttpos_sale_order"
}

// SetFreeOrder 设置免单
func (model *SaleOrder) SetFreeOrder(reason string, freeReasons []*SaleOrderProductReason) {
	defer model.SetUpdate() // 标记更新
	model.IsFree = constant.SaleOrderIsFreeYes
	model.FreeReason = reason
	model.FreeReasons = freeReasons
	// 订单状态
	model.Status = constant.SaleOrderStatusFinish
	model.FinishTime = time.Now().Unix()
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

// GetBuffetUuidMap 获取处理包
type BuffetUuidMapBuffetCustomerTypes struct {
	Uuid    uint64
	MealNum *uint
}

func (b *SaleOrder) GetSaleOrderBuffetCustomerTypes(
	buffetList []*BuffetPackage,
	buffetUuids []uint64,
	buffetCustomerTypes []BuffetUuidMapBuffetCustomerTypes,
	saleBillSetting *SaleBillSetting,
) ([]*SaleOrderBuffetCustomerType, []uint64, uint, int) {
	buffetUuidMap := make(map[uint64]map[uint64]*struct {
		BaseModel
		BuffetPackageUuid  uint64
		CustomerTypeUuid   uint64
		Price              float64
		BuffetCustomerType struct{}
	})
	buffetMap := make(map[uint64]*BuffetPackage)
	//
	for _, buffet := range buffetList {
		for index, _ := range buffet.BuffetCustomerTypePrices {
			customerTypePrice := buffet.BuffetCustomerTypePrices[index]
			if buffetUuidMap[buffet.Uuid] == nil {
				buffetUuidMap[buffet.Uuid] = make(map[uint64]*struct {
					BaseModel
					BuffetPackageUuid  uint64
					CustomerTypeUuid   uint64
					Price              float64
					BuffetCustomerType struct{}
				})
			}
			// 使用匿名结构体
			priceStruct := &struct {
				BaseModel
				BuffetPackageUuid  uint64
				CustomerTypeUuid   uint64
				Price              float64
				BuffetCustomerType struct{}
			}{
				BaseModel:         customerTypePrice.BaseModel,
				BuffetPackageUuid: customerTypePrice.BuffetPackageUuid,
				CustomerTypeUuid:  customerTypePrice.CustomerTypeUuid,
				Price:             customerTypePrice.Price,
			}
			buffetUuidMap[buffet.Uuid][customerTypePrice.CustomerTypeUuid] = priceStruct
		}
		buffetMap[buffet.Uuid] = buffet
	}
	// 使用map来跟踪已经添加的buffetUuid，实现去重
	newBuffetUuidMap2 := make(map[uint64]bool)
	newBuffetUuids := make([]uint64, 0)
	mealNum := uint(0)
	maxTimeLimit := int(0)
	saleOrderBuffetCustomerTypes := make([]*SaleOrderBuffetCustomerType, 0)
	// 创建一个map来跟踪已处理的CustomerType
	processedCustomerTypes := make(map[uint64]bool)
	//
	for _, buffetUuid := range buffetUuids {
		buffetPackage := buffetMap[buffetUuid]
		for _, CustomerType := range buffetCustomerTypes {
			num := *CustomerType.MealNum
			if num == 0 {
				continue
			}
			m := buffetUuidMap[buffetUuid]
			if m[CustomerType.Uuid] == nil {
				continue
			}

			customerTypePrice := m[CustomerType.Uuid]
			// 使用匿名结构体的字段
			buffetCustomerTypePriceUuid := customerTypePrice.BaseModel.Uuid
			taxRate := buffetPackage.GeTaxRate()
			saleOrderBuffetCustomerType := NewSaleOrderBuffetCustomerType(b.Uuid, buffetUuid, buffetCustomerTypePriceUuid, num, customerTypePrice.Price, taxRate, *saleBillSetting)
			saleOrderBuffetCustomerTypes = append(saleOrderBuffetCustomerTypes, saleOrderBuffetCustomerType)
			// 只有当buffetUuid不在map中时，才添加到_buffetUuids
			if !newBuffetUuidMap2[buffetUuid] {
				newBuffetUuids = append(newBuffetUuids, buffetUuid)
				newBuffetUuidMap2[buffetUuid] = true
				// 取得最大的可用餐时长
				if maxTimeLimit != -1 {
					if buffetPackage.IsLimitTime == 0 {
						maxTimeLimit = -1
					} else {
						maxTimeLimit = max(maxTimeLimit, int(buffetPackage.LimitTime)*60)
					}
				}
			}
			//
			// 只有当这个CustomerType未被处理过时，才累加mealNum
			if !processedCustomerTypes[CustomerType.Uuid] {
				mealNum += num
				processedCustomerTypes[CustomerType.Uuid] = true
			}
		}
	}
	//
	return saleOrderBuffetCustomerTypes, newBuffetUuids, mealNum, maxTimeLimit
}

// GetPercentageList 获取当前订单的百分比对象列表
func (model *SaleOrder) GetPercentageList() []map[string]string {
	// 创建 map 来存储不同税率的税费和商品总价
	taxRateMap := make(map[string]float64)
	totalPriceMap := make(map[string]float64)

	// 自助餐顾客类型
	for _, orderBuffetCustomer := range model.SaleOrderBuffetCustomerTypes {
		if orderBuffetCustomer.IsDelete() {
			continue
		}
		// 获取税率
		taxRate := fmt.Sprintf("%.0f", orderBuffetCustomer.TaxRate*100)
		// 累加相同税率的税费和总价
		taxRateMap[taxRate] += orderBuffetCustomer.TaxFee
		totalPriceMap[taxRate] += orderBuffetCustomer.Price
	}

	// 商品列表
	for _, item := range model.SaleOrderProducts {
		if item.IsDelete() || item.IsUnCookingProduct() || item.IsCancelProduct() {
			continue
		}
		// 获取税率
		taxRate := fmt.Sprintf("%.0f", item.TaxRate*100)
		// 累加相同税率的税费和总价
		taxRateMap[taxRate] += item.TaxFee
		totalPriceMap[taxRate] += item.GetPrice()
	}

	// 将 map 转换为数组
	result := make([]map[string]string, 0, len(taxRateMap))
	for taxRate, taxFee := range taxRateMap {
		if taxFee > 0 {
			result = append(result, map[string]string{
				"TaxRate":    taxRate,
				"TaxFee":     fmt.Sprintf("%.2f", taxFee),
				"TotalPrice": fmt.Sprintf("%.2f", totalPriceMap[taxRate]),
			})
		}
	}

	return result
}

// 获取会员余额
func (model *SaleOrder) GetMemberSurplusBalance() float64 {
	if model.Member == nil {
		return 0
	}
	if model.Status == constant.SaleOrderStatusFinish {
		// todo 未完成
		return 0
	}
	return model.Member.GetBalanceAll()
}

// 获取会员积分
func (model *SaleOrder) GetMemberSurplusPoints() float64 {
	if model.IsFree != 0 {
		return 0
	} else {
		// todo 未完成
		return model.Member.Point
	}
}

// 获取支付方式
func (model *SaleOrder) GetPayTypeNames(language string) string {
	payTypeNames := []string{}
	if model.IsFree == 1 {
		// 免单处理
		payTypeName := i18n.Translate(language, "免单")
		if !slices.Contains(payTypeNames, payTypeName) {
			payTypeNames = append(payTypeNames, payTypeName)
		}
	} else {
		// 正常支付方式处理
		for _, payment := range model.PaymentOrders {
			if !slices.Contains(payTypeNames, payment.PaymentMethodName) {
				payTypeNames = append(payTypeNames, payment.PaymentMethodName)
			}
		}
	}
	return strings.Join(payTypeNames, ",")
}
