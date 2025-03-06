package model

import (
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/pkg/utils"

	"github.com/shopspring/decimal"
)

// SaleOrder 销售订单 `ttpos_sale_order`
type SaleOrder struct {
	BaseModel
	// 基础标识字段
	OrderNo    string `gorm:"column:order_no;comment:订单编号" json:"order_no"`
	Status     uint   `gorm:"column:status;comment:订单状态, 0-未结账 1-已结账" json:"status"`
	IsFree     uint   `gorm:"column:is_free;comment:是否免单, 0-否 1-是" json:"is_free"`
	FreeReason string `gorm:"column:free_reason;comment:免单原因" json:"free_reason"`

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
	CustomDiscountFee float64 `gorm:"column:custom_discount_fee;type:decimal(12,2);default:0;comment:自定义折扣金额" json:"discount_fee"`
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
	ZeroCheckoutRule uint8   `gorm:"column:zero_checkout_rule;type:tinyint(1);default:0;comment:结账抹零, 0-实款实收 1-抹分 2-抹角 3-抹元" json:"zero_checkout_rule"`

	// 结账完成后才记录的字段
	PaymentAmount        float64 `gorm:"column:payment_amount;type:decimal(12,2);default:0;comment:支付金额,支付金额-订单总金额=支付手续费" json:"payment_amount"`
	ChangeAmount         float64 `gorm:"column:change_amount;type:decimal(12,2);default:0;comment:找零金额,结账完成后才记录" json:"change_amount"`
	ZeroCheckoutFee      float64 `gorm:"column:zero_checkout_fee;type:decimal(12,2);default:0;comment:结账抹零金额" json:"zero_checkout_fee"`
	FinalPrice           float64 `gorm:"column:final_price;type:decimal(12,2);default:0;comment:最终应收金额。最终应收金额=应收金额+手续费-结账抹零金额" json:"final_price"`
	PaymentCommissionFee float64 `gorm:"column:payment_commission_fee;type:decimal(12,2);default:0;comment:支付手续费,关联付款单的支付手续费之和" json:"payment_commission_fee"`
	GiftAmount           float64 `gorm:"column:gift_amount;type:decimal(12,2);default:0;comment:赠菜金额,(销售订单赠菜商品.总最终单价)之和" json:"gift_amount"`

	// 关联对象
	PaymentOrders                []PaymentOrder                 `gorm:"foreignKey:RelatedUuid;references:uuid"`
	Member                       *Member                        `gorm:"foreignKey:ConsumerUuid;references:uuid"`
	SaleOrderProducts            []*SaleOrderProduct            `gorm:"foreignKey:SaleOrderUuid;references:uuid"`
	ReturnOrders                 []ReturnOrder                  `gorm:"foreignKey:RelatedOrderUuid;references:uuid"`
	SaleOrderBuffetCustomerTypes []*SaleOrderBuffetCustomerType `gorm:"foreignKey:SaleOrderUuid;references:uuid"`
	SaleOrderBuffetDelayProducts []SaleOrderBuffetDelayProduct  `gorm:"foreignKey:SaleOrderUuid;references:uuid"`
}

func (model *SaleOrder) GetMemberName() string {
	if model.Member == nil {
		return ""
	}
	return model.Member.Nickname
}

// CalcGiftAmount 计算赠菜金额. 赠菜金额=销售订单赠菜商品.总最终单价之和
func (model *SaleOrder) CalcGiftAmount() float64 {
	amount := float64(0)
	for _, saleOrderProduct := range model.SaleOrderProducts {
		// 删除的商品不计、退菜的商品不计入、未送厨的商品不计
		if saleOrderProduct.IsDelete() || saleOrderProduct.IsCancelProduct() || saleOrderProduct.IsUnCookingProduct() {
			continue
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
			mealNum += num
		}
	}
	//
	return saleOrderBuffetCustomerTypes, newBuffetUuids, mealNum, maxTimeLimit
}
