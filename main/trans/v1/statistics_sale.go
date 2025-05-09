package v1

import (
	"fmt"
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/utils"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// 订单
type JOrder struct {
	OrderID                           uint64  `gorm:"column:order_id;primaryKey;comment:订单id"`
	ParentID                          uint64  `gorm:"column:parent_id;comment:父订单id"`
	IsBuffet                          int     `gorm:"column:is_buffet;comment:是否是自助餐"`
	TotalPrice                        float64 `gorm:"column:total_price;comment:商品总金额"`
	TotalProductPrice                 float64 `gorm:"column:total_product_price;comment:订单商品总价(原价)"`
	OrderPrice                        float64 `gorm:"column:order_price;comment:订单总金额"`
	PayPrice                          float64 `gorm:"column:pay_price;comment:实际应收金额（不包含退款）"`
	FreePayPrice                      float64 `gorm:"column:free_pay_price;comment:免赠前商品消费税"`
	ActualPrice                       float64 `gorm:"column:actual_price;comment:客户实付金额"`
	OriginalPrice                     float64 `gorm:"column:original_price;comment:订单原始价格"`
	PayStatus                         int     `gorm:"column:pay_status;comment:付款状态(10未付款 20已付款)"`
	IsFree                            int     `gorm:"column:is_free;comment:是否免单"`
	PayTime                           int64   `gorm:"column:pay_time;comment:付款时间"`
	OrderStatus                       int     `gorm:"column:order_status;comment:订单状态10=>进行中，20=>已经取消，30=>已完成"`
	IsSettled                         int     `gorm:"column:is_settled;comment:订单是否已结算(0未结算 1已结算)"`
	UserID                            int     `gorm:"column:user_id;comment:用户id"`
	RefundMoney                       float64 `gorm:"column:refund_money;comment:退款金额"`
	RefundConsumptionTax              float64 `gorm:"column:refund_consumption_tax;comment:退款消费税"`
	DiscountMoney                     float64 `gorm:"column:discount_money;comment:优惠金额"`
	DiscountChangePrice               float64 `gorm:"column:discount_change_price;comment:改价金额"`
	IsChangPrice                      int     `gorm:"column:is_chang_price;comment:是否优惠折扣 0-否 1-是"`
	TableID                           uint64  `gorm:"column:table_id;comment:桌台id"`
	MealNum                           int     `gorm:"column:meal_num;comment:用餐人数"`
	ServiceMoney                      float64 `gorm:"column:service_money;comment:服务费"`
	Balance                           float64 `gorm:"column:balance;comment:余额抵扣金额"`
	OnlineMoney                       float64 `gorm:"column:online_money;comment:在线支付金额"`
	SettingServiceMoney               float64 `gorm:"column:setting_service_money;comment:门店设定服务费"`
	ConsumptionTaxMoney               float64 `gorm:"column:consumption_tax_money;comment:门店设定消费税"`
	OriginalConsumptionTaxMoney       float64 `gorm:"column:original_consumption_tax_money;comment:原价消费税"`
	TotalProductConsumptionTax        float64 `gorm:"column:total_product_consumption_tax;comment:商品消费税(折后)"`
	TotalProductServiceConsumptionTax float64 `gorm:"column:total_product_service_consumption_tax;comment:商品服务费消费税(折后)"`
	TotalProductServiceFee            float64 `gorm:"column:total_product_service_fee;comment:商品服务费(折后)"`
	ConsumptionTaxType                int     `gorm:"column:consumption_tax_type;comment:消费税类型：0关闭, 1商品已含税, 2商品未含税"`
	UserDiscountMoney                 float64 `gorm:"column:user_discount_money;comment:会员折扣金额"`
	PayFeeMoney                       float64 `gorm:"column:pay_fee_money;comment:支付手续费"`
	SmallDiscountType                 int     `gorm:"column:small_discount_type;comment:抹零(优惠折扣)：1-抹分 2-抹角 3-四舍五入到角 4-四舍五入到元"`
	DiscountRate                      float64 `gorm:"column:discount_rate;comment:优惠折扣比例 如：50-百分之五十"`
	ChangeDue                         float64 `gorm:"column:change_due;comment:找零"`
	CheckoutDiffMoney                 float64 `gorm:"checkout_diff_money;comment:结账抹零后与pay_price差值"`
	MergeParentID                     uint64  `gorm:"column:merge_parent_id;comment:合并订单id"`
	IsMerge                           int     `gorm:"column:is_merge;comment:是否合并订单 0-否 1-是"`

	OrderProducts []JOrderProduct           `gorm:"foreignKey:OrderID;references:OrderID"` // 订单商品
	OrderPayTypes []JOrderPayType           `gorm:"foreignKey:OrderID;references:OrderID"` // 订单支付类型
	OrderRefunds  []JOrderRefundDestination `gorm:"foreignKey:OrderID;references:OrderID"` // 订单退款
}

// 订单商品
type JOrderProduct struct {
	OrderProductID                uint64  `gorm:"column:order_product_id;primaryKey;comment:订单id"`
	OrderID                       uint64  `gorm:"column:order_id;comment:订单id"`
	IsReturn                      int     `gorm:"column:is_return;comment:是否退菜 0-否 1-是"`
	TotalNum                      int     `gorm:"column:total_num;comment:购买数量"`
	ProductPrice                  float64 `gorm:"column:product_price;comment:商品价格"`
	ConsumptionTax                float64 `gorm:"column:consumption_tax;comment:商品消费税"`
	IsFree                        int     `gorm:"column:is_free;comment:是否赠菜 0-否 1-赠菜，计入总销售额、优惠折扣 2-免单，不计入总销售额、优惠折扣"`
	ProductOriginalConsumptionTax float64 `gorm:"column:product_original_consumption_tax;comment:商品原价消费税"`
	ProductServiceConsumptionTax  float64 `gorm:"column:product_service_consumption_tax;comment:商品服务费消费税"`
	ProductDiscountMoney          float64 `gorm:"column:product_discount_money;comment:优惠折扣后与原价总差额(包含数量)"`
	ProductServiceFee             float64 `gorm:"column:product_service_fee;comment:商品服务费"`
	TaxCalcType                   int     `gorm:"column:tax_calc_type;comment:是否含税 0-关闭 1-已含税 2-未含税"`

	OrderProductReturns []JOrderProductReturn `gorm:"foreignKey:OrderProductID;references:OrderProductID"` // 订单商品退款
}

// 订单支付类型
type JOrderPayType struct {
	ID        uint    `gorm:"column:id;primaryKey;comment:id"`
	SubID     uint    `gorm:"column:sub_id;comment:子订单id"`
	OrderID   uint    `gorm:"column:order_id;comment:订单id"`
	PayStatus int     `gorm:"column:pay_status;comment:支付状态"`
	Value     int     `gorm:"column:value;comment:支付方式"`
	Price     float64 `gorm:"column:price;comment:支付金额"`
	FeeMoney  float64 `gorm:"column:fee_money;comment:支付手续费"`
}

// 订单退款
type JOrderRefundDestination struct {
	ID                    uint    `gorm:"column:id;primaryKey;comment:id"`
	OrderID               uint    `gorm:"column:order_id;comment:订单id"`
	RefundID              uint    `gorm:"column:refund_id;comment:退款id"`
	PaymentOrderID        uint    `gorm:"column:payment_order_id;comment:支付订单id"`
	MerchantRefundOrderNo string  `gorm:"column:merchant_refund_order_no;comment:商户退款订单号"`
	Status                int     `gorm:"column:status;comment:退款状态 -1-失败 0-处理中 1-完成"`
	Value                 int     `gorm:"column:value;comment:支付方式"`
	Price                 float64 `gorm:"column:price;comment:支付金额"`
	RefundMoney           float64 `gorm:"column:refund_money;comment:退款金额"`
}

// 订单商品退款
type JOrderProductReturn struct {
	ID             uint `gorm:"column:id;primaryKey;comment:id"`
	OrderID        uint `gorm:"column:order_id;comment:订单id"`
	OrderProductID uint `gorm:"column:order_product_id;comment:订单商品id"`
	ProductID      uint `gorm:"column:product_id;comment:商品id"`
	Num            int  `gorm:"column:num;comment:退款数量"`
}

// 订单表名
func (JOrder) TableName() string {
	return "jjjfood_order"
}

// 订单商品表名
func (JOrderProduct) TableName() string {
	return "jjjfood_order_product"
}

// 订单支付类型表名
func (JOrderPayType) TableName() string {
	return "jjjfood_order_pay_type"
}

// 订单退款表名
func (JOrderRefundDestination) TableName() string {
	return "jjjfood_order_refund_destination"
}

// 订单商品退款表名
func (JOrderProductReturn) TableName() string {
	return "jjjfood_order_product_return"
}

// 库存服务
type StatisticsSaleService struct {
	db       *gorm.DB
	targetDB *gorm.DB
}

// 实例化库存服务
func NewStatisticsSaleService(db *gorm.DB, targetDB *gorm.DB) *StatisticsSaleService {
	return &StatisticsSaleService{db: db, targetDB: targetDB}
}

// 迁移销售统计
func (s *StatisticsSaleService) ConvertStatisticsSale() error {
	s1 := s.convertStatisticsSale1()
	if s1 != nil {
		return s1
	}
	s2 := s.convertStatisticsSale2()
	if s2 != nil {
		return s2
	}
	return nil
}

// 迁移销售统计1
func (s *StatisticsSaleService) convertStatisticsSale1() error {
	var isContinue bool = true
	var page int = 1
	var pageSize int = 500

	for isContinue {
		var oldOrders []JOrder
		var newOrders []model.StatisticsSale
		db1 := s.db
		db1.Preload("OrderProducts").
			Preload("OrderPayTypes").
			Preload("OrderRefunds").
			Preload("OrderProducts.OrderProductReturns").
			Where("pay_status = 20").
			Where("order_status = 30").
			Where("parent_id = 0").
			Where("is_merge = 0").
			Where("is_delete = 0").
			Where("delete_time = 0").
			Limit(pageSize).Offset((page - 1) * pageSize).Find(&oldOrders)
		if len(oldOrders) < pageSize {
			isContinue = false
		}

		page++
		for _, order := range oldOrders {
			var orderProductNum int
			var orderProductPrice decimal.Decimal
			var orderProductOriginPrice decimal.Decimal
			var orderProductSalePrice decimal.Decimal
			var orderProductTax decimal.Decimal
			var orderServiceFee decimal.Decimal
			var orderServiceTax decimal.Decimal
			var orderDiscount decimal.Decimal
			var orderGiftNum int
			var orderGiftAmount decimal.Decimal
			var orderFreeNum int
			var orderFreeAmount decimal.Decimal
			var orderPaymentAmount decimal.Decimal
			var orderPaymentBalance decimal.Decimal
			var orderRefundAmount decimal.Decimal
			var orderRefundPaymentBalance decimal.Decimal
			var orderRefundTax decimal.Decimal
			var noOrderRefundTax decimal.Decimal
			var orderRefundServiceFee decimal.Decimal
			var orderRefundDiscount decimal.Decimal
			var orderRefundDiscountMember decimal.Decimal
			var orderRefundFee decimal.Decimal

			var isFree bool = order.IsFree > 0
			var isStatFree bool = order.IsFree == 1
			var isSateGive int
			var isFeeType bool = order.ConsumptionTaxType == 1
			var isFixServiceFee bool = order.SettingServiceMoney > 0

			if isFree {
				// if isStatFree {
				// 	isSateGive = 1
				// } else {
				// 	isSateGive = 2
				// }
				orderFreeNum = 1
			}

			orderFreeAmount = decimal.NewFromFloat(order.FreePayPrice)

			// 统计订单免单
			if isFree {
				if isStatFree {
					orderServiceFee = decimal.NewFromFloat(order.ServiceMoney)
					orderDiscount = decimal.NewFromFloat(order.DiscountMoney)
				}
				// 赠菜计入优惠折扣
				if isSateGive == 1 {
					orderDiscount = decimal.NewFromFloat(order.DiscountMoney)
				}
			} else {
				// 统计自定义优惠折扣
				orderServiceFee = decimal.NewFromFloat(order.ServiceMoney)
				orderDiscount = decimal.NewFromFloat(order.DiscountMoney).Add(decimal.NewFromFloat(order.CheckoutDiffMoney))
			}

			orderProductOriginPrice = orderProductOriginPrice.Add(decimal.NewFromFloat(order.TotalProductPrice))

			if order.IsFree == 2 {
				// 等于子订单的 orderProductTax， orderRefundTax
			} else {
				orderProductTax = orderProductTax.Add(decimal.NewFromFloat(order.ConsumptionTaxMoney))
				orderRefundTax = decimal.NewFromFloat(order.RefundConsumptionTax)
			}

			for _, orderProduct := range order.OrderProducts {
				if orderProduct.IsReturn == 0 {
					totalNum := orderProduct.TotalNum
					totalNumDec := decimal.NewFromFloat(float64(totalNum))
					orderProductNum += totalNum

					productPrice := decimal.NewFromFloat(orderProduct.ProductPrice)
					if isFeeType || (orderProduct.TaxCalcType == 2 && orderProduct.ConsumptionTax > 0 && orderProduct.ProductOriginalConsumptionTax == 0) {
						productPrice = productPrice.Sub(decimal.NewFromFloat(orderProduct.ConsumptionTax))
					}

					orderProductSalePrice = orderProductSalePrice.Add(decimal.NewFromFloat(orderProduct.ProductPrice).Mul(totalNumDec))

					// productTax := decimal.NewFromFloat(orderProduct.ConsumptionTax)
					productServiceTax := decimal.NewFromFloat(orderProduct.ProductServiceConsumptionTax)
					if orderProduct.IsFree > 0 {
						// productTax = decimal.NewFromFloat(0)
						productServiceTax = decimal.NewFromFloat(0)
						orderGiftNum += totalNum
						orderGiftAmount = orderGiftAmount.Add(decimal.NewFromFloat(orderProduct.ProductDiscountMoney))
					}

					// 减去商品服务费消费税
					orderProductTax = orderProductTax.Sub(productServiceTax)

					if isFree {
						if isStatFree {
							orderProductPrice = orderProductPrice.Add(productPrice.Mul(totalNumDec))
							// orderProductTax = orderProductTax.Add(productTax)
							orderServiceTax = orderServiceTax.Add(productServiceTax)
						}
					} else {
						if orderProduct.IsFree > 0 {
							if isSateGive == 0 {
								isSateGive = orderProduct.IsFree
							}
							if isSateGive == 1 {
								orderProductPrice = orderProductPrice.Add(productPrice.Mul(totalNumDec))
								// orderProductTax = orderProductTax.Add(productTax)
								orderServiceTax = orderServiceTax.Add(productServiceTax)
							}
							if isSateGive == 2 {
								orderDiscount = orderDiscount.Sub(decimal.NewFromFloat(orderProduct.ProductDiscountMoney))
							}
						} else {
							orderProductPrice = orderProductPrice.Add(productPrice.Mul(totalNumDec))
							// orderProductTax = orderProductTax.Add(productTax)
							orderServiceTax = orderServiceTax.Add(productServiceTax)
						}
					}

					for _, orderProductReturn := range orderProduct.OrderProductReturns {
						// orderRefundTax = orderRefundTax.Add(decimal.NewFromFloat(orderProduct.ConsumptionTax)).
						// 	Add(decimal.NewFromFloat(orderProduct.ProductServiceConsumptionTax)).
						// 	Mul(decimal.NewFromFloat(float64(orderProductReturn.Num)))
						if isFeeType {
							noOrderRefundTax = noOrderRefundTax.Add(decimal.NewFromFloat(orderProduct.ConsumptionTax))
						}
						if !isFixServiceFee {
							orderRefundServiceFee = orderRefundServiceFee.Add(decimal.NewFromFloat(orderProduct.ProductServiceFee).Mul(decimal.NewFromFloat(float64(orderProductReturn.Num))))
						}
					}

				}
			}

			for _, orderPayType := range order.OrderPayTypes {
				if orderPayType.PayStatus == 1 {
					if orderPayType.Value == 10 {
						orderPaymentBalance = orderPaymentBalance.Add(decimal.NewFromFloat(orderPayType.Price))
					}
				}
			}

			// for _, orderRefund := range order.OrderRefunds {
			// 	if orderRefund.Status == 1 {
			// 		if orderRefund.Value == 10 {
			// 			orderRefundPaymentBalance = orderRefundPaymentBalance.Add(decimal.NewFromFloat(orderRefund.RefundMoney))
			// 		} else {
			// 			orderRefundAmount = orderRefundAmount.Add(decimal.NewFromFloat(orderRefund.RefundMoney))
			// 		}
			// 	}
			// }

			orderRefundAmount = decimal.NewFromFloat(order.RefundMoney)
			if order.MergeParentID != 0 && order.IsMerge != 1 {
				orderPaymentAmount = decimal.NewFromFloat(0)
			} else {
				orderPaymentAmount = decimal.NewFromFloat(order.PayPrice)
			}

			if !isFree && order.PayPrice == order.RefundMoney {
				orderRefundFee = decimal.NewFromFloat(order.PayFeeMoney)
				// orderRefundDiscount = decimal.NewFromFloat(order.DiscountMoney)
				orderRefundDiscountMember = decimal.NewFromFloat(order.UserDiscountMoney)
				if isFixServiceFee {
					orderRefundServiceFee = decimal.NewFromFloat(order.ServiceMoney)
				}
			}

			newOrders = append(newOrders, model.StatisticsSale{
				SaleBillUuid:         order.OrderID,
				SaleOrderUuid:        order.OrderID,
				DeskUuid:             order.TableID,
				MealNum:              order.MealNum,
				ProductPrice:         orderProductPrice.InexactFloat64(),
				ProductOriginPrice:   orderProductOriginPrice.InexactFloat64(),
				ProductSalePrice:     orderProductSalePrice.InexactFloat64(),
				ProductNum:           orderProductNum,
				ProductTax:           orderProductTax.InexactFloat64(),
				ServiceFee:           orderServiceFee.InexactFloat64(),
				ServiceTax:           orderServiceTax.InexactFloat64(),
				Discount:             orderDiscount.InexactFloat64(),
				DiscountMember:       order.UserDiscountMoney,
				GiftAmount:           orderGiftAmount.InexactFloat64(),
				GiftNum:              orderGiftNum,
				FreeAmount:           orderFreeAmount.InexactFloat64(),
				FreeNum:              orderFreeNum,
				PaymentAmount:        orderPaymentAmount.InexactFloat64(),
				PaymentFee:           order.PayFeeMoney,
				PaymentBalance:       orderPaymentBalance.InexactFloat64(),
				RefundAmount:         orderRefundAmount.InexactFloat64(),
				RefundPaymentBalance: orderRefundPaymentBalance.InexactFloat64(),
				RefundTax:            orderRefundTax.InexactFloat64(),
				NoRefundTax:          noOrderRefundTax.InexactFloat64(),
				RefundServiceFee:     orderRefundServiceFee.InexactFloat64(),
				RefundDiscount:       orderRefundDiscount.InexactFloat64(),
				RefundDiscountMember: orderRefundDiscountMember.InexactFloat64(),
				RefundFee:            orderRefundFee.InexactFloat64(),
				CompleteTime:         order.PayTime,
				IsSpecial:            utils.IfInt(order.MergeParentID != 0 && order.IsMerge != 1, 1, 0),
			})
		}

		if len(newOrders) > 0 {
			fmt.Println(fmt.Sprintf("statistics-sale - page: %d - num: %d", page, len(newOrders)))
			if err := s.targetDB.Create(&newOrders).Error; err != nil {
				return err
			}
		}
	}

	return nil
}

// 迁移销售统计
func (s *StatisticsSaleService) convertStatisticsSale2() error {
	var isContinue bool = true
	var page int = 1
	var pageSize int = 500

	for isContinue {
		var oldOrders []JOrder
		var newOrders []model.StatisticsSale
		db1 := s.db
		db1.Preload("OrderProducts").
			Preload("OrderPayTypes").
			Preload("OrderRefunds").
			Preload("OrderProducts.OrderProductReturns").
			Where("pay_status = 20").
			Where("order_status = 30").
			Where("parent_id = 0").
			Where("is_merge = 1").
			Where("is_delete = 0").
			Where("delete_time = 0").
			Limit(pageSize).Offset((page - 1) * pageSize).Find(&oldOrders)
		if len(oldOrders) < pageSize {
			isContinue = false
		}

		page++
		for _, order := range oldOrders {
			var orderServiceFee decimal.Decimal
			var orderPaymentAmount decimal.Decimal
			var orderPaymentBalance decimal.Decimal

			for _, orderPayType := range order.OrderPayTypes {
				if orderPayType.PayStatus == 1 {
					if orderPayType.Value == 10 {
						orderPaymentBalance = orderPaymentBalance.Add(decimal.NewFromFloat(orderPayType.Price))
					}
				}
			}

			orderPaymentAmount = decimal.NewFromFloat(order.PayPrice)

			newOrders = append(newOrders, model.StatisticsSale{
				SaleBillUuid:   order.OrderID,
				SaleOrderUuid:  order.OrderID,
				DeskUuid:       order.TableID,
				ServiceFee:     orderServiceFee.InexactFloat64(),
				PaymentAmount:  orderPaymentAmount.InexactFloat64(),
				PaymentFee:     order.PayFeeMoney,
				PaymentBalance: orderPaymentBalance.InexactFloat64(),
				CompleteTime:   order.PayTime,
				IsMeger:        1,
			})
		}

		if len(newOrders) > 0 {
			fmt.Println(fmt.Sprintf("statistics-sale - page: %d - num: %d", page, len(newOrders)))
			if err := s.targetDB.Create(&newOrders).Error; err != nil {
				return err
			}
		}
	}

	return nil
}
