package model

import (
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/pkg/utils"
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
	Amount        float64 `gorm:"column:amount;type:decimal(12,2);default:0;comment:应收金额。商品未含税时，总金额=商品金额+服务费+税费。商品已含税时，总金额=商品金额（含商品消费税）+服务费+税费（只有服务费税）" json:"amount"`
	CustomAmount  float64 `gorm:"column:custom_amount;type:decimal(12,2);default:-1;comment:整单改价金额。改价后，应收金额=整单改价金额，前端优先显示改价后的金额，改价金额不能为负数。当为-1时，表示不改价，显示amount改收金额" json:"custom_amount"`
	PaymentAmount float64 `gorm:"column:payment_amount;type:decimal(12,2);default:0;comment:支付金额,支付金额-订单总金额=支付手续费" json:"payment_amount"`
	ChangeAmount  float64 `gorm:"column:change_amount;type:decimal(12,2);default:0;comment:找零金额,结账完成后才记录" json:"change_amount"`

	// 时间相关字段
	FinishTime int64 `gorm:"column:finish_time;type:int(10);default:0;comment:完成时间（时间戳）" json:"finish_time"`

	MemberDiscountRate     float64 `gorm:"column:member_discount_rate;type:decimal(12,2);default:1;comment:会员折扣率(0-100%)，默认100%，取值范围0-1，如折扣率为10%，则取值为0.1	" json:"member_discount_rate"`
	MemberCardDiscountRate float64 `gorm:"column:member_card_discount_rate;type:decimal(12,2);default:1;comment:会员卡折扣率(0-100%)，默认100%，取值范围0-1，如折扣率为10%，则取值为0.1" json:"member_card_discount_rate"`
	CustomDiscountRate     float64 `gorm:"column:custom_discount_rate;type:decimal(12,2);default:1;comment:自定义折扣率(0-100%)，默认100%，取值范围0-1，如折扣率为10%，则取值为0.1" json:"custom_discount_rate"`

	// 抹零相关
	ZeroRule         uint8   `gorm:"column:zero_rule;type:tinyint(1);default:0;comment:优惠折扣抹零, 0-实款实收 1-抹分 2-抹角 3-四舍五入保留一位小数 4-四舍五入保留整数" json:"zero_rule"`
	ZeroFee          float64 `gorm:"column:zero_fee;type:decimal(12,2);default:0;comment:优惠折扣抹零金额" json:"zero_fee"`
	ZeroCheckoutRule uint8   `gorm:"column:zero_checkout_rule;type:tinyint(1);default:0;comment:结账抹零, 0-实款实收 1-抹分 2-抹角 3-抹元" json:"zero_checkout_rule"`
	ZeroCheckoutFee  float64 `gorm:"column:zero_checkout_fee;type:decimal(12,2);default:0;comment:结账抹零金额" json:"zero_checkout_fee"`

	// 关联对象
	PaymentOrders                []PaymentOrder                 `gorm:"foreignKey:RelatedUuid;references:uuid"`
	Member                       Member                         `gorm:"foreignKey:ConsumerUuid;references:uuid"`
	SaleOrderProducts            []*SaleOrderProduct            `gorm:"foreignKey:SaleOrderUuid;references:uuid"`
	ReturnOrders                 []ReturnOrder                  `gorm:"foreignKey:RelatedOrderUuid;references:uuid"`
	SaleOrderBuffetCustomerTypes []*SaleOrderBuffetCustomerType `gorm:"foreignKey:SaleOrderUuid;references:uuid"`
	SaleOrderBuffetDelayProducts []SaleOrderBuffetDelayProduct  `gorm:"foreignKey:SaleOrderUuid;references:uuid"`
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

func (model *SaleOrder) SetFinishStatus(changeAmount float64) {
	model.Status = constant.SaleOrderStatusFinish
	model.FinishTime = time.Now().Unix()
	model.ChangeAmount = changeAmount
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
type GetBuffetUuidMapBuffetCustomerTypes []struct {
	Uuid    uint64
	MealNum *uint
}

func (b *SaleOrder) GetBuffetUuidMap(
	buffetList []*BuffetPackage,
	BuffetUuids []uint64,
	BuffetCustomerTypes GetBuffetUuidMapBuffetCustomerTypes,
	SaleBillSetting *SaleBillSetting,
) []*SaleOrderBuffetCustomerType {
	buffetUuidMap := make(map[uint64]map[uint64]*struct {
		BaseModel
		BuffetPackageUuid  uint64
		CustomerTypeUuid   uint64
		Price              float64
		BuffetCustomerType struct{}
	})
	buffetMap := make(map[uint64]*BuffetPackage)
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
	saleOrderBuffetCustomerTypes := make([]*SaleOrderBuffetCustomerType, 0)
	for _, buffetUuid := range BuffetUuids {
		buffetPackage := buffetMap[buffetUuid]
		for _, CustomerType := range BuffetCustomerTypes {
			num := *CustomerType.MealNum
			if num == 0 {
				continue
			}
			m := buffetUuidMap[buffetUuid]
			customerTypePrice := m[CustomerType.Uuid]
			// 使用匿名结构体的字段
			buffetCustomerTypePriceUuid := customerTypePrice.BaseModel.Uuid
			taxRate := buffetPackage.GeTaxRate()
			saleOrderBuffetCustomerType := NewSaleOrderBuffetCustomerType(b.Uuid, buffetUuid, buffetCustomerTypePriceUuid, num, customerTypePrice.Price, taxRate, *SaleBillSetting)
			saleOrderBuffetCustomerTypes = append(saleOrderBuffetCustomerTypes, saleOrderBuffetCustomerType)
		}
	}
	return saleOrderBuffetCustomerTypes
}
