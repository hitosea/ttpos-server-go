package model

import (
	"fmt"
	"slices"
	"sort"
	"strings"
	"time"

	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/resp"
	settingResp "ttpos-server-go/app/dto/resp/setting"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/i18n"

	"github.com/shopspring/decimal"
)

func (model *SaleOrder) GetDiscountInfo() DiscountInfo {
	return DiscountInfo{
		MemberDiscountRate:     model.MemberDiscountRate,
		MemberCardDiscountRate: model.MemberCardDiscountRate,
		CustomDiscountRate:     model.CustomDiscountRate,
	}
}

// 获取免单原因
func (model *SaleOrder) GetFreeReason() dto.LocaleResponse {
	// 获取免单原因
	zhNames := make([]string, 0)
	thNames := make([]string, 0)
	enNames := make([]string, 0)
	zhtwNames := make([]string, 0)
	jaNames := make([]string, 0)
	koNames := make([]string, 0)
	myNames := make([]string, 0)
	trNames := make([]string, 0)
	svNames := make([]string, 0)
	// 遍历选择的免单原因
	for _, reason := range model.FreeReasons {
		if !reason.IsFreeReason() || reason.IsDelete() {
			continue
		}
		zhNames = append(zhNames, reason.MultiLanguageName.ZhName)
		thNames = append(thNames, reason.MultiLanguageName.ThName)
		enNames = append(enNames, reason.MultiLanguageName.EnName)
		zhtwNames = append(zhtwNames, reason.MultiLanguageName.ZhTwName)
		jaNames = append(jaNames, reason.MultiLanguageName.JaName)
		koNames = append(koNames, reason.MultiLanguageName.KoName)
		myNames = append(myNames, reason.MultiLanguageName.MyName)
		trNames = append(trNames, reason.MultiLanguageName.TrName)
		svNames = append(svNames, reason.MultiLanguageName.SvName)
	}
	// 添加自定义的免单原因
	if model.FreeReason != "" {
		zhNames = append(zhNames, model.FreeReason)
		thNames = append(thNames, model.FreeReason)
		enNames = append(enNames, model.FreeReason)
		zhtwNames = append(zhtwNames, model.FreeReason)
		jaNames = append(jaNames, model.FreeReason)
		koNames = append(koNames, model.FreeReason)
		myNames = append(myNames, model.FreeReason)
		trNames = append(trNames, model.FreeReason)
		svNames = append(svNames, model.FreeReason)
	}
	reasonDto := dto.LocaleResponse{
		ZH:   strings.Join(zhNames, "、"),
		TH:   strings.Join(thNames, "、"),
		EN:   strings.Join(enNames, "、"),
		ZHTW: strings.Join(zhtwNames, "、"),
		JA:   strings.Join(jaNames, "、"),
		KO:   strings.Join(koNames, "、"),
		MY:   strings.Join(myNames, "、"),
		TR:   strings.Join(trNames, "、"),
		SV:   strings.Join(svNames, "、"),
	}
	return reasonDto
}

// 获取所有自助餐名称
func (model *SaleOrder) GetBuffetNames(language string) string {
	buffets := make([]string, 0)
	for _, buffet := range model.SaleOrderBuffetCustomerTypes {
		buffets = append(buffets, buffet.BuffetPackage.MultiLanguageName.GetNameByLang(language))
	}
	return strings.Join(buffets, "+")
}

// 获取实际支付金额 - 结账之后才有值
func (model *SaleOrder) GetActualPaymentAmount() float64 {
	return model.PaymentAmount - model.GetTotalRefundAmount()
}

// 获取汇总支付订单的支付金额 - 未结账也有值
func (model *SaleOrder) GetSummaryPaymentAmount() float64 {
	amount := 0.0
	for _, paymentOrder := range model.PaymentOrders {
		if paymentOrder.Status != constant.PaymentOrderStatusRefund {
			amount += paymentOrder.PaymentAmount
		}
	}
	return amount
}

// 获取待支付的金额
func (model *SaleOrder) GetUnpaidAmount() float64 {
	unpaidAmount, _ := model.calcFinallyAmount()
	if unpaidAmount < 0 {
		return 0
	}
	return unpaidAmount - model.GetSummaryPaymentAmount()
}

// 获取总的退款金额
func (model *SaleOrder) GetTotalRefundAmount() float64 {
	refundAmount := decimal.NewFromFloat(0)
	for _, refundOrder := range model.ReturnOrders {
		refundAmount = refundAmount.Add(decimal.NewFromFloat(refundOrder.RefundAmount))
	}
	return refundAmount.InexactFloat64()
}

// 返回销售订单商品
func (model *SaleOrder) GetSaleOrderProduct(saleOrderProductUuid uint64) (*SaleOrderProduct, int, error) {
	for i, saleOrderProduct := range model.SaleOrderProducts {
		if saleOrderProductUuid == saleOrderProduct.Uuid {
			return saleOrderProduct, i, nil
		}
	}
	return nil, 0, errors.New("销售订单商品不存在")
}

// 获取退款商品列表
func (model *SaleOrder) GetSaleOrderProductList(saleOrderProductUuids []uint64) []*SaleOrderProduct {
	saleOrderProducts := make([]*SaleOrderProduct, 0)
	productMap := make(map[uint64]*SaleOrderProduct)
	for _, saleOrderProduct := range model.SaleOrderProducts {
		productMap[saleOrderProduct.Uuid] = saleOrderProduct
	}
	for _, saleOrderProductUuid := range saleOrderProductUuids {
		saleOrderProduct, ok := productMap[saleOrderProductUuid]
		if !ok {
			continue
		}
		saleOrderProducts = append(saleOrderProducts, saleOrderProduct)
	}
	return saleOrderProducts
}

// 获取退款的自助餐顾客列表
func (model *SaleOrder) GetSaleOrderBuffetComstomerTypeList(productUuids []uint64) []*SaleOrderBuffetCustomerType {
	targetList := make([]*SaleOrderBuffetCustomerType, 0)
	productMap := make(map[uint64]*SaleOrderBuffetCustomerType)
	for _, customerType := range model.SaleOrderBuffetCustomerTypes {
		productMap[customerType.Uuid] = customerType
	}
	for _, productUuid := range productUuids {
		saleOrderBuffetCustomerType, ok := productMap[productUuid]
		if !ok {
			continue
		}
		targetList = append(targetList, saleOrderBuffetCustomerType)
	}
	return targetList
}

// 获取退款的自助餐加钟列表
func (model *SaleOrder) GetSaleOrderBuffetDelayList(productUuids []uint64) []*SaleOrderBuffetDelayProduct {
	targetList := make([]*SaleOrderBuffetDelayProduct, 0)
	productMap := make(map[uint64]*SaleOrderBuffetDelayProduct)
	for _, delayProduct := range model.SaleOrderBuffetDelayProducts {
		productMap[delayProduct.Uuid] = delayProduct
	}
	for _, productUuid := range productUuids {
		saleOrderBuffetDelay, ok := productMap[productUuid]
		if !ok {
			continue
		}
		targetList = append(targetList, saleOrderBuffetDelay)
	}
	return targetList
}

// 获取销售订单应收金额
func (model *SaleOrder) GetAmount() float64 {
	// 整单改价金额大于等于0时，返回整单改价金额
	if model.CustomAmount >= 0 {
		return model.CustomAmount
	}
	// 默认返回订单总金额
	return model.Amount
}

// 打印 - 获取打印单的最终应收金额
func (model *SaleOrder) GetPrintReceivablePrice() float64 {
	finalPrice := model.FinalPrice
	// 未结账时，需要计算最终应收金额
	if model.Status != constant.SaleOrderStatusFinish {
		finalPrice, _ = model.calcFinallyAmount()
	}
	// 如果是免单，返回0
	if model.IsFreeSaleOrder() {
		return 0
	}
	return finalPrice
}

// 根据sign获取销售订单商品
func (model *SaleOrder) GetSaleOrderProductBySign(sign string) *SaleOrderProduct {
	for _, saleOrderProduct := range model.SaleOrderProducts {
		if saleOrderProduct.Sign == sign {
			return saleOrderProduct
		}
	}
	return nil
}

// 获取未送厨的订单商品金额（折后价）
func (model *SaleOrder) GetUnCookingProductAmount() float64 {
	return model.calcSumOrderProductPrice(model.GetUnCookingOrderProductList())
}

// 获取已送厨的订单商品金额（折后价）
func (model *SaleOrder) GetCookingProductAmount() float64 {
	return model.calcProductAmount(model.GetCookingOrderProductList())
}

// 获取已送厨商品的订单服务费
func (model *SaleOrder) GetCookingProductServiceFee(serviceFeeType int, serviceFeeValue float64) float64 {
	return model.calcServiceFee(model.GetCookingOrderProductList(), serviceFeeType, serviceFeeValue)
}

// 获取已送厨商品的订单消费税
func (model *SaleOrder) GetCookingProductTaxFee() float64 {
	return model.calcTaxFee(model.GetCookingOrderProductList())
}

// 获取已送厨的订单商品列表
func (model *SaleOrder) GetCookingOrderProductList() []*SaleOrderProduct {
	return model.GetAllOrderProductList(WithCooking())
}

// 获取未送厨的订单商品列表，不包含未接单的商品
func (model *SaleOrder) GetUnCookingOrderProductList() []*SaleOrderProduct {
	return model.GetAllOrderProductList(WithUnCooking())
}

// 获取全部商品，包括已送厨和未送厨
func (model *SaleOrder) GetUnCookingAndCookingOrderProductList(h5OrderUuid uint64, isCanceled bool) []*SaleOrderProduct {
	var products []*SaleOrderProduct
	if h5OrderUuid == 0 {
		products = model.GetAllOrderProductList(WithAll())
	} else {
		products = model.GetAllOrderProductList(WithAllAndOneH5Order(h5OrderUuid))
	}
	products = FilterPackageSubProduct(products)
	if isCanceled {
		products = FilterUnCookingOrderProduct(products)
	}
	return FilterUnAcceptOrderProduct(products, h5OrderUuid)
}

// 过滤未送厨的商品
func FilterUnCookingOrderProduct(products []*SaleOrderProduct) []*SaleOrderProduct {
	list := make([]*SaleOrderProduct, 0)
	for _, product := range products {
		if product.IsUnCookingProduct() {
			continue
		}
		list = append(list, product)
	}
	return list
}

// 过滤套餐子商品
func FilterPackageSubProduct(products []*SaleOrderProduct) []*SaleOrderProduct {
	list := make([]*SaleOrderProduct, 0)
	for _, product := range products {
		if product.IsPackageSubProduct() {
			continue
		}
		list = append(list, product)
	}
	return list
}

// 过滤未接单的商品
func FilterUnAcceptOrderProduct(products []*SaleOrderProduct, h5OrderUuid uint64) []*SaleOrderProduct {
	list := make([]*SaleOrderProduct, 0)
	for _, product := range products {
		if product.IsAcceptOrder == constant.OrderProductIsAcceptOrderUnAccept {
			// 如果从待接单进入桌台时，不过滤该h5订单的商品
			if product.H5OrderUuid != 0 && product.H5OrderUuid == h5OrderUuid {
				list = append(list, product)
			}
			continue
		}
		list = append(list, product)
	}
	return list
}

// 获取h5已下单的商品
func (model *SaleOrder) GetH5OrderProductList() []*SaleOrderProduct {
	return model.GetAllOrderProductList(WithH5Order())
}

// 获取h5购物车的商品(未下单的商品)
func (model *SaleOrder) GetH5CartProductList() []*SaleOrderProduct {
	return model.GetAllOrderProductList(WithH5Cart())
}

// 获取全部商品，包括已送厨和未送厨
func (model *SaleOrder) GetAllOrderProductList(options ...func(option *CalcOption)) []*SaleOrderProduct {
	option := &CalcOption{}
	for _, optionFunc := range options {
		optionFunc(option)
	}

	products := make([]*SaleOrderProduct, 0)
	for _, orderProduct := range model.SaleOrderProducts {
		if orderProduct == nil {
			continue
		}
		// 已经移动到其他订单的商品不计
		if orderProduct.SaleOrderUuid != model.Uuid {
			continue
		}

		// 删除的商品不计入
		if orderProduct.IsDelete() {
			continue
		}

		// 赠菜？计入
		// 退菜？退了不计入
		if orderProduct.IsCancelProduct() {
			continue
		}

		if option.H5OrderStatus == H5OrderStatusAccepted {
			// H5下单的商品计入
			if orderProduct.IsH5OrderProductBool() {
				products = append(products, orderProduct)
				continue
			}
			continue
		}

		if option.H5OrderStatus == H5OrderStatusUnAccepted {
			// H5购物车的商品计入
			if orderProduct.IsH5CartProduct() {
				products = append(products, orderProduct)
				continue
			}
			continue
		}

		if option.CookingStatus == CookingStatusCooking {
			// 已送厨的商品计入
			if orderProduct.IsCookingProduct() {
				products = append(products, orderProduct)
				continue
			}
			continue
		}

		if option.CookingStatus == CookingStatusUnCooking {
			// 未送厨的商品计入
			if orderProduct.IsUnCookingProduct() {
				products = append(products, orderProduct)
				continue
			}
			continue
		}

		if option.CookingStatus == CookingStatusAllAndOneH5Order {
			// 已送厨和未送厨的商品和某个h5订单的商品计入，排除未接单的商品
			if orderProduct.IsUnOrderH5OrderProduct() /*未下单*/ {
				continue
			}
			// 如果是已下单未接单商品且是某个h5订单的商品时记入
			if orderProduct.IsH5OrderProductBool() /*未接单*/ {
				if orderProduct.H5OrderUuid != option.H5OrderUuid {
					continue
				}
			}
			products = append(products, orderProduct)
			continue
		}

		if option.CookingStatus == CookingStatusAll {
			// 已送厨和未送厨的商品计入，排除未接单的商品
			if orderProduct.IsH5OrderProductBool() /*未接单*/ || orderProduct.IsUnOrderH5OrderProduct() /*未下单*/ {
				continue
			}
			products = append(products, orderProduct)
			continue
		}
		// 不做任何过滤，所有商品
		products = append(products, orderProduct)
	}
	return products
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

// 外送订单的可退款金额。可退款金额=订单最终应收金额-配送费-已退款金额
func (model *SaleOrder) GetCanReturnAmountWithDeliveryFee(deliveryFee float64) float64 {
	return decimal.NewFromFloat(model.GetCanReturnAmount()).Sub(decimal.NewFromFloat(deliveryFee)).Round(2).InexactFloat64()
}

// 本单最多可退的会员累计消费金额。=销售订单应收-结账抹零金额。 销售订单应收=购物车应收-积分抵扣金额-优惠券抵扣金额
func (model *SaleOrder) GetCanReturnMemberConsumptionAmountMax() float64 {
	return decimal.NewFromFloat(model.GetAmountValue()).Sub(decimal.NewFromFloat(model.ZeroCheckoutFee)).Round(2).InexactFloat64()
}

// 剩余可退的会员累计消费金额，不能小于0。=本单最多可退的会员累计消费金额-已退的会员累计消费金额。 已退的会员累计消费金额=本单的退款商品表中商品的金额之和
func (model *SaleOrder) GetCanReturnMemberConsumptionAmount() float64 {
	// 已经退的累计消费金额=本单的退款商品表中商品的金额之和
	returnAmount := model.GetReturnProductAmount()
	// 本单最多可退的会员累计消费金额
	maxReturnAmount := model.GetCanReturnMemberConsumptionAmountMax()
	// 不能小于0
	if returnAmount > maxReturnAmount {
		returnAmount = maxReturnAmount
	}
	// 剩余可退的会员累计消费金额
	return decimal.NewFromFloat(maxReturnAmount).Sub(decimal.NewFromFloat(returnAmount)).Round(2).InexactFloat64()
}

// 订单的退款商品累计的金额
func (model *SaleOrder) GetReturnProductAmount() float64 {
	amount := decimal.NewFromFloat(0)
	for _, returnOrder := range model.ReturnOrders {
		amount = amount.Add(decimal.NewFromFloat(returnOrder.RefundAmount))
	}
	return amount.Round(2).InexactFloat64()
}

// GetOriginAmount 获取订单没打折之前的订单应收金额。原订单应收金额=现应收金额+会员折扣金额+优惠折扣金额
func (model *SaleOrder) GetOriginAmount() float64 {
	//原订单应收金额=现应收金额+会员折扣金额+优惠折扣金额
	return decimal.NewFromFloat(model.Amount).Add(decimal.NewFromFloat(model.MemberDiscountFee)).Add(decimal.NewFromFloat(model.CustomDiscountFee)).Round(2).InexactFloat64()
}

// GetOriginAmount 获取订单没打折之前的订单应收金额。原订单应收金额=现应收金额+会员折扣金额+优惠折扣金额
func (model *SaleOrder) GetOriginAmountValue() float64 {
	return model.OriginAmount
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

// 获取该销售订单使用会员余额支付的金额
func (model *SaleOrder) GetMemberBalanceAmount() float64 {
	memberBalanceAmount := decimal.NewFromFloat(0)
	for _, paymentOrder := range model.PaymentOrders {
		if paymentOrder.PaymentMethod.Code == constant.PaymentMethodCodeBalance {
			memberBalanceAmount = memberBalanceAmount.Add(decimal.NewFromFloat(paymentOrder.Amount))
		}
	}
	return memberBalanceAmount.InexactFloat64()
}

// 获取该销售订单使用现金支付的金额
func (model *SaleOrder) GetCashAmount() float64 {
	cashAmount := decimal.NewFromFloat(0)
	for _, paymentOrder := range model.PaymentOrders {
		if paymentOrder.PaymentMethod.Code == constant.PaymentMethodCodeCash {
			cashAmount = cashAmount.Add(decimal.NewFromFloat(paymentOrder.Amount))
		}
	}
	return cashAmount.InexactFloat64()
}

func (b *SaleOrder) GetSaleOrderBuffetCustomerTypes(
	buffetList []*BuffetPackage,
	buffetUuids []uint64,
	buffetCustomerTypes []BuffetUuidMapBuffetCustomerTypes,
	saleBillSetting *SaleBillSetting,
) ([]*SaleOrderBuffetCustomerType, []uint64, uint, int, uint, uint) {
	buffetUuidMap := make(map[uint64]map[uint64]*struct {
		BaseModel
		Name               string
		BuffetPackageUuid  uint64
		CustomerTypeUuid   uint64
		Price              float64
		BuffetCustomerType struct{}
	})
	buffetMap := make(map[uint64]*BuffetPackage)
	//
	var nonOrderingTimes, reminderOrderTimes []uint
	for _, buffet := range buffetList {
		nonOrderingTimes = append(nonOrderingTimes, buffet.NonOrderingTime)
		reminderOrderTimes = append(reminderOrderTimes, buffet.ReminderOrderTime)
		for index, _ := range buffet.BuffetCustomerTypePrices {
			customerTypePrice := buffet.BuffetCustomerTypePrices[index]
			if buffetUuidMap[buffet.Uuid] == nil {
				buffetUuidMap[buffet.Uuid] = make(map[uint64]*struct {
					BaseModel
					Name               string
					BuffetPackageUuid  uint64
					CustomerTypeUuid   uint64
					Price              float64
					BuffetCustomerType struct{}
				})
			}
			// 使用匿名结构体
			priceStruct := &struct {
				BaseModel
				Name               string
				BuffetPackageUuid  uint64
				CustomerTypeUuid   uint64
				Price              float64
				BuffetCustomerType struct{}
			}{
				BaseModel:         customerTypePrice.BaseModel,
				Name:              customerTypePrice.BuffetCustomerType.Name,
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
			saleOrderBuffetCustomerType := NewSaleOrderBuffetCustomerType(customerTypePrice.Name, b.Uuid, b.SaleBillUuid, buffetUuid, buffetCustomerTypePriceUuid, num, customerTypePrice.Price, taxRate, *saleBillSetting, buffetPackage.OpenOverallDiscount)
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
	var minNonOrderingTime, minReminderOrderTime uint
	if len(nonOrderingTimes) > 0 {
		minNonOrderingTime = slices.Min(nonOrderingTimes)
	}
	if len(reminderOrderTimes) > 0 {
		minReminderOrderTime = slices.Min(reminderOrderTimes)
	}
	return saleOrderBuffetCustomerTypes, newBuffetUuids, mealNum, maxTimeLimit, minNonOrderingTime, minReminderOrderTime
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
		taxRateMap[taxRate] += orderBuffetCustomer.GetTotalTaxFee()
		totalPriceMap[taxRate] += orderBuffetCustomer.GetDiscountPrice()
	}

	// 商品列表
	for _, item := range model.SaleOrderProducts {
		if item.IsDelete() || item.IsUnCookingProduct() || item.IsUnAcceptOrderBool() || item.IsCancelProduct() {
			continue
		}
		// 获取税率
		taxRate := fmt.Sprintf("%.0f", item.TaxRate*100)
		// 累加相同税率的税费和总价
		taxRateMap[taxRate] += item.GetTotalTaxFee()
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
		return model.MemberBalance
	}
	return model.Member.GetBalanceAll()
}

// 获取会员积分
func (model *SaleOrder) GetMemberSurplusPoints(mealNum int, rule settingResp.PointsRule) float64 {
	if model.Member == nil || model.IsFree != 0 {
		return 0
	} else {
		if model.Status == constant.SaleOrderStatusFinish {
			return model.GiftPoints
		}
		if rule.Value == 0 {
			return 0
		}
		// 计算本单获取的积分
		model.SetGiftPointsRate(mealNum, rule)
		// 如果是免单，返回0积分
		if model.IsFreeSaleOrder() {
			return 0
		}
		// 计算本单应收金额
		baseNum := model.GetFinalNoFeeAmount() // 计算积分的基数，值为本订单的应收金额(已减积分抵扣金额、已减结账抹零金额)
		return model.CalcMemberPoint(mealNum, rule, baseNum)

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

// GetSaleOrderBuffetDelayTimeTotal 总加钟时间
func (model *SaleOrder) GetSaleOrderBuffetDelayTimeTotal() int64 {
	delayTime := int64(0)
	for _, saleOrderProduct := range model.SaleOrderBuffetDelayProducts {
		// 商品的加钟时间
		delayTime += saleOrderProduct.DelayTime
	}
	return delayTime
}

// 设置为空。为了更新数据库数据时，不更新关联对象
func (model *SaleOrder) SetNil() {
	model.PaymentOrders = nil
	model.Member = nil
	model.SaleOrderProducts = nil
	model.ReturnOrders = nil
	model.SaleOrderBuffetCustomerTypes = nil
	model.SaleOrderBuffetDelayProducts = nil
}

func (model *SaleOrder) SetCashier(cashierUuid uint64, cashierName string) {
	model.CashierUuid = cashierUuid
	model.CashierName = cashierName
}

// 设置会员折扣，并修改订单商品的折扣
func (model *SaleOrder) setMemberDiscount(memberUuid uint64, memberDiscount, cardDiscount float64) {
	// 修改订单的会员信息
	model.MemberDiscountRate = memberDiscount
	model.MemberCardDiscountRate = cardDiscount
	model.ConsumerUuid = memberUuid
	// 对商品进行打折
	for _, saleOrderProduct := range model.SaleOrderProducts {
		// 如果订单商品已删除，则不修改折扣. 已退菜、赠菜的商品也要修改折扣，表示退菜的金额也打折了
		if saleOrderProduct.IsDelete() {
			continue
		}
		saleOrderProduct.SetMemberDiscountInfo(model.MemberDiscountRate, model.MemberCardDiscountRate)
		saleOrderProduct.SetUpdate()
	}
	// 对自助餐顾客进行打折. 顾客没有会员折扣
}

// 设置会员折扣
func (model *SaleOrder) SetMemberDiscount(member Member) {
	defer model.SetZeroRuleCancel()     // 设置会员折扣后，要取消订单抹零
	defer model.SetCustomAmountCancel() // 设置会员折扣后，要取消整单改价
	defer model.SetPayPointsCancel()    // 设置会员折扣后，要取消积分抵扣

	// 修改订单的会员信息
	model.setMemberDiscount(member.Uuid, member.GetMemberDiscountRate(), member.GetMemberCardDiscountRate())
}

// 判断订单是否进行了整单改价或抹零
func (model *SaleOrder) IsCustomAmount() bool {
	return model.CustomAmount != -1
}

// 判断订单是否进行了订单抹零
func (model *SaleOrder) IsZeroRule() bool {
	return model.ZeroRule != 0
}

// 设置会员折扣取消
func (model *SaleOrder) SetMemberDiscountCancel() {
	// 修改订单的会员信息
	discountRate := float64(1)                  // 无折扣，1乘任何价格都等于原价
	model.MemberDiscountRate = discountRate     // 会员折扣，无折扣
	model.MemberCardDiscountRate = discountRate // 会员卡折扣，无折扣
	model.ConsumerUuid = 0                      // 会员ID置空
	defer model.SetPayPointsCancel()            // 设置会员折扣后，要取消积分抵扣
	// 对商品进行打折
	for _, saleOrderProduct := range model.SaleOrderProducts {
		// 如果订单商品已删除，则不修改折扣. 已退菜、赠菜的商品也要修改折扣，表示退菜的金额也打折了
		if saleOrderProduct.IsDelete() {
			continue
		}
		saleOrderProduct.SetMemberDiscountInfo(discountRate, discountRate)
		saleOrderProduct.SetUpdate()
	}
}

// 删除销售订单中所有优惠券
func (model *SaleOrder) SetAllCouponCancel() {
	for _, coupon := range model.Coupons {
		if coupon.IsDelete() {
			continue
		}
		coupon.SetDelete()
	}
}

// 当订单取消会员时，删除销售订单中已经选中的会员优惠券
func (model *SaleOrder) SetMemberCouponCancel() {
	for _, coupon := range model.Coupons {
		if coupon.IsDelete() {
			continue
		}
		if coupon.IsMemberCoupon() {
			coupon.SetDelete()
		}
	}
}

// 当订单取消积分时，删除销售订单中已经选中的积分优惠券
func (model *SaleOrder) SetPointsCouponCancel() {
	for _, coupon := range model.Coupons {
		if coupon.IsDelete() {
			continue
		}
		coupon.SetDelete()
	}
}

// 设置整单折扣，并修改订单商品的折扣
// 参数discount，表示给订单设置的打折率，统一使用百分比打折。比如八折，discount值为0.8；比如30% off，discount值为0.7。
// 注意：请在调用该方法时，就做好discount值的转化
func (model *SaleOrder) SetCustomDiscount(discount float64) {
	defer model.SetCustomAmountCancel() // 取消整单改价金额
	defer model.SetZeroRuleCancel()     // 取消订单抹零

	model.CustomDiscountRate = discount
	// 对商品进行打折
	for _, saleOrderProduct := range model.SaleOrderProducts {
		// 如果订单商品已删除，则不修改折扣. 已退菜、赠菜的商品也要修改折扣，表示退菜的金额也打折了
		if saleOrderProduct.IsDelete() {
			continue
		}
		saleOrderProduct.CustomDiscountRate = discount
		saleOrderProduct.SetUpdate()
	}
	// 对自助餐顾客进行打折
	for _, buffetCustomer := range model.SaleOrderBuffetCustomerTypes {
		if buffetCustomer.IsDelete() {
			continue
		}
		buffetCustomer.CustomDiscountRate = discount
		buffetCustomer.SetUpdate()
	}
}

// 取消整单折扣
func (model *SaleOrder) SetCustomDiscountCancel() bool {
	isChange := false
	model.CustomDiscountRate = constant.NoDiscount
	for _, saleOrderProduct := range model.SaleOrderProducts {
		// 如果订单商品已删除，则不修改折扣. 已退菜、赠菜的商品也要修改折扣，表示退菜的金额也打折了
		if saleOrderProduct.IsDelete() {
			continue
		}
		// 如果订单商品折扣不为100%，则修改折扣。确保如果原本就没有自定义折扣就不用更新数据库
		if saleOrderProduct.CustomDiscountRate != constant.NoDiscount {
			saleOrderProduct.CustomDiscountRate = constant.NoDiscount
			saleOrderProduct.SetUpdate()
			isChange = true
		}
	}
	// 取消自助餐顾客折扣
	for _, buffetCustomer := range model.SaleOrderBuffetCustomerTypes {
		if buffetCustomer.IsDelete() {
			continue
		}
		// 如果自助餐顾客折扣不为100%，则修改折扣。确保如果原本就没有自定义折扣就不用更新数据库
		if buffetCustomer.CustomDiscountRate != constant.NoDiscount {
			buffetCustomer.CustomDiscountRate = constant.NoDiscount
			buffetCustomer.SetUpdate()
			isChange = true
		}
	}
	return isChange
}

// 设置整单改价金额
func (model *SaleOrder) SetCustomAmount(amount float64) {
	defer model.SetZeroRuleCancel()       // 取消订单抹零
	defer model.SetCustomDiscountCancel() // 取消整单折扣
	model.CustomAmount = amount
}

// 取消整单改价金额
func (model *SaleOrder) SetCustomAmountCancel() bool {
	isChange := false
	model.CustomAmount = constant.SaleOrderCustomAmountCancel
	return isChange
}

// 取消积分抵扣金额
func (model *SaleOrder) SetPayPointsCancel() {
	model.PayPoints = 0       // 积分抵扣金额置空
	model.PayPointsAmount = 0 // 积分抵扣金额置空
}

// 设置订单抹零规则
func (model *SaleOrder) SetZeroRule(zeroRule int) {
	model.ZeroRule = uint8(zeroRule)
}

// 取消订单抹零
func (model *SaleOrder) SetZeroRuleCancel() bool {
	isChange := false
	// 将订单的抹零规则设置为实款实收
	if model.ZeroRule != constant.DiscountZeroRuleNone {
		model.ZeroRule = constant.DiscountZeroRuleNone
		isChange = true
	}
	return isChange
}

// SetCheckoutZeroRuleCancel 取消结账抹零
func (model *SaleOrder) SetCheckoutZeroRuleCancel() {
	// 将订单的结账抹零规则设置为实款实收
	if model.ZeroCheckoutRule != constant.SaleBillSettingCheckoutZeroingMethodNone {
		model.ZeroCheckoutRule = constant.SaleBillSettingCheckoutZeroingMethodNone
		model.ZeroCheckoutFee = 0
		model.SetUpdate()
	}
}

// 取消整单折扣
func (model *SaleOrder) SetAllDiscountCancel() bool {
	isChange := false
	isChange = model.SetZeroRuleCancel() || isChange
	isChange = model.SetCustomDiscountCancel() || isChange
	isChange = model.SetCustomAmountCancel() || isChange
	return isChange
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

// SetCancelFreeOrder 设置免单
func (model *SaleOrder) SetCancelFreeOrder() {
	defer model.SetUpdate() // 标记更新
	model.IsFree = constant.SaleOrderIsFreeNo
	model.FreeReason = ""
}

// 是否存在手动折扣
func (model *SaleOrder) IsManualDiscount(ZeroRule uint8) bool {
	// custom_amount == -1 是没有进行订单改价
	// custom_discount_rate = 1 是没有折扣
	// zero_rule = 0 是没有去零
	return model.CustomAmount != -1 || model.CustomDiscountRate != 1 || (model.ZeroRule != 0 && model.ZeroRule != ZeroRule)
}

// IsMemberDiscount 判断是否存在会员优惠折扣
func (model *SaleOrder) IsMemberDiscount() bool {
	return model.ConsumerUuid != 0
}

// HasCheckoutZeroRule 判断订单是否存在结账抹零
func (model *SaleOrder) HasCheckoutZeroRule() bool {
	// 如果订单的结账抹零规则不为实款实收，则表示订单存在结账抹零
	return model.ZeroCheckoutRule != constant.SaleBillSettingCheckoutZeroingMethodNone
}

type DiscountInfo struct {
	MemberDiscountRate     float64 `json:"member_discount_rate"`
	MemberCardDiscountRate float64 `json:"member_card_discount_rate"`
	CustomDiscountRate     float64 `json:"custom_discount_rate"`
}

// 计算销售订单结账抹零金额
func (model *SaleOrder) GetCheckOutZeroFee() float64 {
	if model.Status != constant.SaleOrderStatusFinish {
		return model.CalcCheckOutZeroFee()
	}
	return model.ZeroCheckoutFee
}
