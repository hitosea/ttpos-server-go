package model

import (
	"fmt"
	"slices"
	"strings"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/i18n"
	"ttpos-server-go/pkg/utils"

	"github.com/shopspring/decimal"
)

func (model *SaleBill) GetBuffetProductList() resp.BuffetProductList {
	buffetProductList := resp.BuffetProductList{}
	buffetProductList.List = make([]resp.BuffetProduct, 0)
	// 去重的列表
	list := make([]resp.BuffetProduct, 0)
	if model.BuffetPackage1 != nil {
		buffetProductList.List = append(buffetProductList.List, model.BuffetPackage1.GetBuffetProductList().List...)
	}
	if model.BuffetPackage2 != nil {
		buffetProductList.List = append(buffetProductList.List, model.BuffetPackage2.GetBuffetProductList().List...)
	}
	// 去重, 合并显示设置（优先显示）, 并重新计算限制数量
	// 如果桌台选了A、B两个自助餐，同一商品在A设置显示、在B设置不显示，最终该商品显示
	buffetProductMap := make(map[uint64]int) // uuid -> list中的索引
	for _, buffetProduct := range buffetProductList.List {
		if idx, exists := buffetProductMap[buffetProduct.Uuid]; exists {
			// 已存在，合并显示设置（优先显示：只要有一个设置为显示，最终就显示）
			if buffetProduct.IsShowCashier == 1 {
				list[idx].IsShowCashier = 1
			}
			if buffetProduct.IsShowTablet == 1 {
				list[idx].IsShowTablet = 1
			}
			if buffetProduct.IsShowKitchen == 1 {
				list[idx].IsShowKitchen = 1
			}
			if buffetProduct.IsShowAssistant == 1 {
				list[idx].IsShowAssistant = 1
			}
		} else {
			// 不存在，添加到列表
			buffetProduct.Limit = buffetProduct.Limit * model.MealNum // 限制数量=单人限购数量*用餐人数
			buffetProductMap[buffetProduct.Uuid] = len(list)          // 记录该商品在list中的索引
			list = append(list, buffetProduct)
		}
	}
	buffetProductList.List = list
	return buffetProductList
}

// 获取支付方式名称列表. 获取所有子单的付款单用到的支付方式，不重复
func (model *SaleBill) GetPaymentMethodNameList(lang string) []string {
	payMethods := make(map[string]bool)
	for _, saleOrder := range model.SaleOrders {
		if saleOrder.IsFreeSaleOrder() {
			payMethods[i18n.Translate(lang, "免单")] = true
			continue
		}
		for _, paymentOrder := range saleOrder.PaymentOrders {
			if paymentOrder.IsDelete() {
				continue
			}
			payMethods[paymentOrder.PaymentMethodName] = true
		}
	}
	payMethodList := make([]string, 0)
	for payMethod := range payMethods {
		payMethodList = append(payMethodList, payMethod)
	}
	return payMethodList
}

// 获取已送厨的销售订单商品
func (model *SaleBill) GetSaleOrderProductCooking() []*SaleOrderProduct {
	cookingSaleOrderProducts := make([]*SaleOrderProduct, 0)
	for _, saleOrder := range model.SaleOrders {
		for i := range saleOrder.SaleOrderProducts {
			orderProduct := saleOrder.SaleOrderProducts[i]
			if !orderProduct.IsAcceptOrderBool() || orderProduct.IsDelete() || orderProduct.IsCancelProduct() {
				continue
			}
			if orderProduct.Status == constant.SaleOrderProductStatusCooking {
				cookingSaleOrderProducts = append(cookingSaleOrderProducts, saleOrder.SaleOrderProducts[i])
			}
		}
	}
	return cookingSaleOrderProducts
}

// 获取未送厨的销售订单商品
func (model *SaleBill) GetSaleOrderProductUnCooking() []*SaleOrderProduct {
	unCookingSaleOrderProducts := make([]*SaleOrderProduct, 0)
	for _, saleOrder := range model.SaleOrders {
		packageUuidRemarkMap := make(map[uint64]string)
		for _, orderProduct := range saleOrder.SaleOrderProducts {
			if orderProduct.IsPackageProduct() {
				packageUuidRemarkMap[orderProduct.Uuid] = orderProduct.Remark
			}
		}
		for i := range saleOrder.SaleOrderProducts {
			orderProduct := saleOrder.SaleOrderProducts[i]
			if !orderProduct.IsAcceptOrderBool() || orderProduct.IsDelete() {
				continue
			}
			if orderProduct.Status == constant.SaleOrderProductStatusNormal {
				if saleOrder.SaleOrderProducts[i].IsPackageSubProduct() {
					saleOrder.SaleOrderProducts[i].Remark = packageUuidRemarkMap[saleOrder.SaleOrderProducts[i].PackageUuid]
				}
				unCookingSaleOrderProducts = append(unCookingSaleOrderProducts, saleOrder.SaleOrderProducts[i])
			}
		}
	}
	return unCookingSaleOrderProducts
}

// 判断是否需要弹出分批送厨弹窗。有分批商品且未标记分批类型时，需要弹出分批送厨弹窗
func (model *SaleBill) IsNeedBatchSendCooking() bool {
	saleOrderProducts := make([]*SaleOrderProduct, 0)
	// 获取未标记分批类型的分批商品
	for _, saleOrder := range model.SaleOrders {
		for _, saleOrderProduct := range saleOrder.SaleOrderProducts {
			if saleOrderProduct.IsDelete() || saleOrderProduct.IsCancelProduct() || saleOrderProduct.IsUnAcceptOrderBool() {
				// 删除、取消、未接单的商品不参与判断
				continue
			}
			if saleOrderProduct.IsBatchBool() && saleOrderProduct.IsPreCooking() {
				saleOrderProducts = append(saleOrderProducts, saleOrderProduct)
			}
		}
	}
	return len(saleOrderProducts) > 0
}

func FilterPaymentDeductStockProduct(products []*SaleOrderProduct) []*SaleOrderProduct {
	list := make([]*SaleOrderProduct, 0)
	for _, product := range products {
		if product.DeductStockType == constant.ProductPackageDeductStockTypePay {
			continue
		}
		list = append(list, product)
	}
	return list
}

// 获取未送厨的销售订单商品
func (model *SaleBill) SetSaleOrderProductCooking() {
	for _, saleOrder := range model.SaleOrders {
		for _, saleOrderProduct := range saleOrder.SaleOrderProducts {
			if !saleOrderProduct.IsAcceptOrderBool() || saleOrderProduct.IsDelete() {
				continue
			}
			if saleOrderProduct.Status == constant.SaleOrderProductStatusNormal {
				saleOrderProduct.SetCooking(0)
			}
		}
	}
}

// 获取未送厨的销售订单商品.获取出刚刚被送厨的商品
func (model *SaleBill) GetSaleOrderProductUnCookingByUuids(uuids map[uint64]bool) []*SaleOrderProduct {
	unCookingSaleOrderProducts := make([]*SaleOrderProduct, 0)
	for _, saleOrder := range model.SaleOrders {
		for i := range saleOrder.SaleOrderProducts {
			orderProduct := saleOrder.SaleOrderProducts[i]
			if uuids[orderProduct.Uuid] {
				unCookingSaleOrderProducts = append(unCookingSaleOrderProducts, orderProduct)
			}
		}
	}
	return unCookingSaleOrderProducts
}

// 获取销售账单中某个H5订单的已下单但未接单的商品
func (model *SaleBill) GetH5OrderProductUnAccept(h5OrderUuid uint64) []*SaleOrderProduct {
	products := make([]*SaleOrderProduct, 0)
	for _, saleOrder := range model.SaleOrders {
		for _, saleOrderProduct := range saleOrder.SaleOrderProducts {
			if saleOrderProduct.H5OrderUuid == h5OrderUuid && !saleOrderProduct.IsAcceptOrderBool() && !saleOrderProduct.IsDelete() {
				products = append(products, saleOrderProduct)
			}
		}
	}
	return products
}

// 获取订单中下单减库存的商品
func (model *SaleBill) GetSaleOrderProductCookingSubStock() []*SaleOrderProduct {
	products := make([]*SaleOrderProduct, 0)
	cookingSaleOrderProducts := model.GetSaleOrderProductCooking()
	for _, saleOrderProduct := range cookingSaleOrderProducts {
		if saleOrderProduct.DeductStockType == constant.ProductPackageDeductStockTypeCooking {
			products = append(products, saleOrderProduct)
		}
	}
	return products
}

// 获取已送厨和未送厨的销售订单商品
func (model *SaleBill) GetSaleOrderProductAll(options ...func(option *CalcOption)) []*SaleOrderProduct {
	option := &CalcOption{}
	for _, optionFunc := range options {
		optionFunc(option)
	}

	// 获取SaleOrderProductUuids商品的ProductPackageUuid
	productPackageUuids := make([]uint64, 0)
	for _, saleOrder := range model.SaleOrders {
		for _, saleOrderProduct := range saleOrder.SaleOrderProducts {
			if slices.Contains(option.SaleOrderProductUuids, saleOrderProduct.Uuid) {
				productPackageUuids = append(productPackageUuids, saleOrderProduct.ProductPackageUuid)
			}
		}
	}

	saleOrderProducts := make([]*SaleOrderProduct, 0)
	for _, saleOrder := range model.SaleOrders {
		for i := range saleOrder.SaleOrderProducts {
			orderProduct := saleOrder.SaleOrderProducts[i]
			if orderProduct == nil {
				continue
			}

			// 在检查限购的时候，只检查SaleOrderProductUuids中相关的商品包是否超过限购
			if option.SaleOrderProductUuids != nil && !slices.Contains(productPackageUuids, orderProduct.ProductPackageUuid) {
				continue
			}

			if option.H5CheckLimit {
				if orderProduct.IsDelete() {
					continue
				}
				// 如果是在接单场景下检查限购，则只检查当前h5订单的商品。不是本h5订单的商品跳过
				if orderProduct.H5OrderUuid != 0 && option.H5OrderUuid != 0 && orderProduct.H5OrderUuid != option.H5OrderUuid {
					continue
				}
			} else {
				// 非h5端检查限购的场景下，未接单的商品不占限购数
				if !orderProduct.IsAcceptOrderBool() || orderProduct.IsDelete() {
					continue
				}
			}
			saleOrderProducts = append(saleOrderProducts, saleOrder.SaleOrderProducts[i])
		}
	}
	return saleOrderProducts
}

// 返回新的销售账单
func (model *SaleBill) GetSaleOrder(saleOrderUuid uint64) *SaleOrder {
	for i, saleOrder := range model.SaleOrders {
		if saleOrderUuid == saleOrder.Uuid {
			if len(model.SaleOrders) > 1 {
				saleOrder.SetIndex(i + 1)
			}
			return saleOrder
		}
	}
	return nil
}

// saleBill的商品不能超过1000项。
func (model *SaleBill) CheckSaleOrderProductNum() error {
	num := 0
	for _, saleOrder := range model.SaleOrders {
		num += len(saleOrder.SaleOrderProducts)
	}
	if num > 1000 {
		return errors.WithMessage(errors.NewWithCode(constant.CodeOrderCheckProductLimitOut, "最多可下单1000种商品"))
	}
	return nil
}

// 通过签名获取销售订单商品
func (model *SaleBill) GetSaleOrderProductBySign(sign string) []*SaleOrderProduct {
	products := make([]*SaleOrderProduct, 0)
	for _, saleOrder := range model.SaleOrders {
		for _, saleOrderProduct := range saleOrder.SaleOrderProducts {
			if saleOrderProduct.IsDelete() {
				continue
			}
			if saleOrderProduct.Sign == sign {
				products = append(products, saleOrderProduct)
			}
		}
	}
	return products
}

// 通过签名获取销售订单商品
func (model *SaleBill) GetSaleOrderProductByUuid(uuid uint64) *SaleOrderProduct {
	for _, saleOrder := range model.SaleOrders {
		for _, saleOrderProduct := range saleOrder.SaleOrderProducts {
			if saleOrderProduct.IsDelete() {
				continue
			}
			if saleOrderProduct.Uuid == uuid {
				return saleOrderProduct
			}
		}
	}
	return nil
}

// 返回第一个销售订单
func (model *SaleBill) GetFirstSaleOrder() *SaleOrder {
	return model.SaleOrders[0]
}

// 重新设置外送订单的会员端折扣率，并重新计算商品价格
func (model *SaleBill) SetMemberOrderDiscountRate(rate float64) {
	defer model.SetUpdate()
	for _, saleOrder := range model.SaleOrders {
		saleOrder.SetUpdate()
		for _, saleOrderProduct := range saleOrder.SaleOrderProducts {
			saleOrderProduct.MemberOrderDiscountRate = rate
			saleOrderProduct.SetUpdate()
		}
	}
}

// 获取销售账单的销售订单和销售订单商品
func (model *SaleBill) GetSaleOrderAndProduct(saleOrderUuid uint64, saleOrderProductUuid uint64) (*SaleOrder, *SaleOrderProduct) {
	// 获取销售订单
	saleOrder := model.GetSaleOrder(saleOrderUuid)
	if saleOrder == nil {
		return nil, nil
	}
	// 获取销售订单商品
	saleOrderProduct, _, err := saleOrder.GetSaleOrderProduct(saleOrderProductUuid)
	if err != nil {
		return saleOrder, nil
	}
	return saleOrder, saleOrderProduct
}

// 获取自助餐名称
// 优先使用快照字段，降级使用关联表数据，支持多语言
// Requirement: story-main-buffet-package-name-snapshot-fix
func (model *SaleBill) GetBuffetName() (name dto.LocaleResponse) {
	// 优先使用快照字段
	name1 := model.GetLocaleBuffetPackage1Name()
	name2 := model.GetLocaleBuffetPackage2Name()

	// 如果两个套餐都有数据，合并显示
	if !name1.IsNull() && !name2.IsNull() {
		name = dto.LocaleResponse{
			ZH:   fmt.Sprintf("%s+%s", name1.ZH, name2.ZH),
			TH:   fmt.Sprintf("%s+%s", name1.TH, name2.TH),
			EN:   fmt.Sprintf("%s+%s", name1.EN, name2.EN),
			ZHTW: fmt.Sprintf("%s+%s", name1.ZHTW, name2.ZHTW),
			JA:   fmt.Sprintf("%s+%s", name1.JA, name2.JA),
			KO:   fmt.Sprintf("%s+%s", name1.KO, name2.KO),
			MY:   fmt.Sprintf("%s+%s", name1.MY, name2.MY),
			TR:   fmt.Sprintf("%s+%s", name1.TR, name2.TR),
			SV:   fmt.Sprintf("%s+%s", name1.SV, name2.SV),
		}
		return
	}
	// 只有一个自助餐时都是只填在BuffetPackage1
	if !name1.IsNull() {
		name = name1
		return
	}
	// 如果套餐1为空，尝试使用套餐2（兼容异常情况）
	if !name2.IsNull() {
		name = name2
		return
	}
	return name
}

// IsBuffetTabletH5TimeSet 助手、平板、h5时间是否已设置
func (model *SaleBill) IsBuffetTabletH5TimeSet() bool {
	if model.BuffetPackage1 != nil && model.BuffetPackage2 != nil {
		return model.BuffetPackage1.NonOrderingTime != 0 || model.BuffetPackage1.ReminderOrderTime != 0 ||
			model.BuffetPackage2.NonOrderingTime != 0 || model.BuffetPackage2.ReminderOrderTime != 0
	} else if model.BuffetPackage1 != nil {
		return model.BuffetPackage1.NonOrderingTime != 0 || model.BuffetPackage1.ReminderOrderTime != 0
	} else if model.BuffetPackage2 != nil {
		return model.BuffetPackage2.NonOrderingTime != 0 || model.BuffetPackage2.ReminderOrderTime != 0
	}
	return false
}

// 自助餐还剩余多少秒
func (model *SaleBill) GetBuffetRemainingSeconds() int64 {
	if model.BuffetDuration == 0 {
		return -1
	}
	remainingTime := model.GetBuffetEndTime() - time.Now().Unix()
	if remainingTime < 0 {
		return 0
	}
	return remainingTime
}

// 获取加钟剩余时长
func (model *SaleBill) GetRemainingDelayDuration() int64 {
	useDuration := time.Now().Unix() - model.DelayStartTime
	if useDuration <= 0 {
		useDuration = 0
	}
	duration := int64(model.DelayDuration) - useDuration
	if duration < 0 {
		return 0
	}
	return duration
}

// 获取总的剩余时长. -1表示自助餐不限时
func (model *SaleBill) GetTotalRemainingSeconds() int64 {
	if model.BuffetDuration == 0 {
		return -1
	}
	return model.GetRemainingDelayDuration() + model.GetBuffetRemainingSeconds()
}

// 获取剩余可点自助餐商品时长
func (model *SaleBill) GetRemainingOrderingSeconds() int64 {
	if model.GetTotalRemainingSeconds() == -1 {
		return -1
	}
	remainingOrderingSeconds := model.GetTotalRemainingSeconds() - int64(model.NonOrderingTime)*60
	if remainingOrderingSeconds > 0 {
		return remainingOrderingSeconds
	}
	return 0
}

// 获取自助餐结束时间
func (model *SaleBill) GetBuffetEndTime() int64 {
	endTime := model.BuffetStartTime + int64(model.BuffetDuration)
	return endTime
}

// 获取总的退款金额
func (model *SaleBill) GetTotalRefundAmount() float64 {
	refundAmount := decimal.NewFromFloat(0)
	for _, saleOrder := range model.SaleOrders {
		for _, refundOrder := range saleOrder.ReturnOrders {
			refundAmount = refundAmount.Add(decimal.NewFromFloat(refundOrder.RefundAmount))
		}
	}
	return refundAmount.InexactFloat64()
}

// GetPaymentAmount 获取账单的付款金额 = 订单金额 - 退款金额
func (model *SaleBill) GetPaymentAmount() float64 {
	// 退款金额
	refundAmount := model.GetTotalRefundAmount()
	// 所有销售订单的付款金额
	paymentAmount := decimal.NewFromFloat(0)
	for _, saleOrder := range model.SaleOrders {
		paymentAmount = paymentAmount.Add(decimal.NewFromFloat(saleOrder.PaymentAmount))
	}
	//订单金额=PaymentAmount-退款金额
	return paymentAmount.Sub(decimal.NewFromFloat(refundAmount)).InexactFloat64()
}

// 获取所有自助餐名称
// 优先使用快照字段，降级使用关联表数据，支持多语言
// Requirement: story-main-buffet-package-name-snapshot-fix
func (model *SaleBill) GetBuffetNames(language string) string {
	buffets := make([]string, 0)
	buffetUuidMap := make(map[uint64]bool) // 用于去重

	// 优先使用快照字段
	if model.BuffetPackage1Uuid != 0 {
		name1 := model.GetLocaleBuffetPackage1Name()
		if !name1.IsNull() {
			name := name1.GetLocale(language)
			if name != "" && !slices.Contains(buffets, name) {
				buffets = append(buffets, name)
				buffetUuidMap[model.BuffetPackage1Uuid] = true
			}
		}
	}
	if model.BuffetPackage2Uuid != 0 {
		name2 := model.GetLocaleBuffetPackage2Name()
		if !name2.IsNull() {
			name := name2.GetLocale(language)
			if name != "" && !slices.Contains(buffets, name) {
				buffets = append(buffets, name)
				buffetUuidMap[model.BuffetPackage2Uuid] = true
			}
		}
	}

	// 降级：如果快照字段为空，使用关联表（兼容历史数据）
	// 遍历 SaleOrderBuffetCustomerTypes，但跳过已经在快照中找到的套餐
	for _, order := range model.SaleOrders {
		for _, buffet := range order.SaleOrderBuffetCustomerTypes {
			// 如果这个套餐已经在快照中找到，跳过
			if buffetUuidMap[buffet.BuffetPackageUuid] {
				continue
			}
			// 使用关联表数据
			name := buffet.BuffetPackage.MultiLanguageName.GetNameByLang(language)
			if name != "" && !slices.Contains(buffets, name) {
				buffets = append(buffets, name)
			}
		}
	}
	return strings.Join(buffets, "+")
}

// GetPayTypes 获取支付方式
func (model *SaleBill) GetPayTypes(language string, saleOrderUuid uint64) []resp.OrderInfoPayTypes {
	payTypeNames := []string{}
	payTypeMap := make(map[string]*resp.OrderInfoPayTypes)
	for _, saleOrder := range model.SaleOrders {
		if saleOrderUuid > 0 && saleOrderUuid != saleOrder.Uuid {
			continue
		}
		if saleOrder.IsFree == 1 {
			// 免单处理
			payTypeName := i18n.Translate(language, "免单")
			key := fmt.Sprintf("%s_%d", payTypeName, 2)
			if existingPayType, ok := payTypeMap[key]; ok {
				existingPayType.PaymentAmount += saleOrder.PaymentAmount
			} else {
				if !slices.Contains(payTypeNames, payTypeName) {
					payTypeNames = append(payTypeNames, payTypeName)
				}
				payType := resp.OrderInfoPayTypes{
					Uuid:            0,
					PaymentTypeName: payTypeName,
					CurrencyUnit:    "",
					PaymentAmount:   saleOrder.PaymentAmount,
					Status:          2,
					Source:          0,
					SourceText:      "",
				}
				payTypeMap[key] = &payType
			}
		} else {
			// 正常支付方式处理
			for _, payment := range saleOrder.PaymentOrders {
				if payment.Status != constant.PaymentOrderStatusPaid || payment.IsDelete() {
					continue
				}
				key := fmt.Sprintf("%s_%d", payment.PaymentMethodName, payment.Status)
				if existingPayType, ok := payTypeMap[key]; ok {
					existingPayType.PaymentAmount += payment.PaymentAmount
				} else {
					payType := resp.OrderInfoPayTypes{
						Uuid:            payment.Uuid,
						PaymentTypeName: payment.PaymentMethodName,
						CurrencyUnit:    payment.CurrencyUnit,
						PaymentAmount:   payment.PaymentAmount,
						Status:          uint(payment.Status),
						StatusReason:    payment.StatusReason,
						Source:          uint(payment.GetSource()),
						SourceText:      payment.GetSourceText(language),
					}
					payTypeMap[key] = &payType
					if !slices.Contains(payTypeNames, payment.PaymentMethodName) {
						payTypeNames = append(payTypeNames, payment.PaymentMethodName)
					}
				}
			}
		}
	}
	payTypes := make([]resp.OrderInfoPayTypes, 0)
	for _, payType := range payTypeMap {
		payTypes = append(payTypes, *payType)
	}
	return payTypes
}

// 获取未下单的h5订单商品
func (model *SaleBill) GetUnOrderH5OrderProduct() []*SaleOrderProduct {
	// 获取第一个销售订单
	saleOrder := model.SaleOrders[0]
	// 获取未下单的h5订单商品
	h5OrderProducts := saleOrder.GetH5CartProductList() // FIXME 为什么会有已下单的h5订单商品？
	// 过滤掉已下单的h5订单商品
	newH5OrderProducts := make([]*SaleOrderProduct, 0)
	for _, h5OrderProduct := range h5OrderProducts {
		if h5OrderProduct.IsPackageSubProduct() {
			continue // 套餐子商品跳过
		}
		if !h5OrderProduct.IsH5CartProduct() {
			continue
		}
		newH5OrderProducts = append(newH5OrderProducts, h5OrderProduct)
	}
	return newH5OrderProducts
}

// 获取套餐子商品列表
func (model *SaleBill) GetSubProducts(saleOrderProductUuid uint64) []*SaleOrderProduct {
	// 获取第一个销售订单
	saleOrder := model.SaleOrders[0]
	// 获取未下单的h5订单商品
	h5OrderProducts := saleOrder.GetH5CartProductList() // FIXME 为什么会有已下单的h5订单商品？
	// 过滤掉已下单的h5订单商品
	subProducts := make([]*SaleOrderProduct, 0)
	for _, h5OrderProduct := range h5OrderProducts {
		if h5OrderProduct.IsPackageSubProduct() {
			if h5OrderProduct.PackageUuid == saleOrderProductUuid {
				subProducts = append(subProducts, h5OrderProduct)
			}
		}
	}
	return subProducts
}

// 获取未下单的h5订单商品数量
func (model *SaleBill) GetUnOrderH5OrderProductNum() float64 {
	num := decimal.NewFromFloat(0)
	products := model.GetUnOrderH5OrderProduct()
	for _, product := range products {
		num = num.Add(product.GetNumDecimal())
	}
	return num.Truncate(2).InexactFloat64()
}

// 获取未接单的h5订单商品的商品金额之和
func (model *SaleBill) GetUnAcceptH5OrderProductTotalPrice(h5OrderProducts []*SaleOrderProduct) float64 {
	totalPrice := decimal.NewFromFloat(0)
	for _, h5OrderProduct := range h5OrderProducts {
		totalPrice = totalPrice.Add(decimal.NewFromFloat(h5OrderProduct.Price).Mul(h5OrderProduct.GetNumDecimal()))
	}
	return totalPrice.Truncate(2).InexactFloat64()
}

// 获取超过限购的商品
func (model *SaleBill) GetSaleOrderProductOverLimit(limitProducts map[uint64]uint, options ...func(option *CalcOption)) []*SaleOrderProduct {
	products := make([]*SaleOrderProduct, 0)

	saleOrderProductAll := model.GetSaleOrderProductAll(options...)
	// product_package_uuid => num
	numMap := make(map[uint64]float64) // key为商品包uuid value为已购买数量
	// product_package_uuid => ProductPackage
	productPackageMap := make(map[uint64]*ProductPackage) // key为商品包uuid value为已购买数量
	// product_package_uuid => SaleOrderProduct
	saleOrderProductMap := make(map[uint64]*SaleOrderProduct) // key为商品包uuid value为订单商品
	for _, saleOrderProduct := range saleOrderProductAll {
		// 限购检查只检查本台的商品，并台过来的商品不记.
		// 跳过非本台的商品
		if !saleOrderProduct.IsCurrentDeskProduct() {
			continue
		}
		// 跳过退菜商品
		if saleOrderProduct.IsCancelProduct() {
			continue
		}
		productPackageUuid := saleOrderProduct.ProductPackageUuid
		productPackageMap[productPackageUuid] = saleOrderProduct.ProductPackage
		numMap[productPackageUuid] = decimal.NewFromFloat(numMap[productPackageUuid]).Add(saleOrderProduct.GetNumDecimal()).Truncate(2).InexactFloat64()
		saleOrderProductMap[productPackageUuid] = saleOrderProduct
	}

	for productPackageUuid, num := range numMap {
		limitNum := productPackageMap[productPackageUuid].LimitNum
		// 如果商品是自助餐商品的话，使用自助餐商品的限购规则
		if buffetLimitNum, ok := limitProducts[productPackageUuid]; ok {
			limitNum = buffetLimitNum
		}
		// 0表示不限购
		if limitNum == 0 {
			// 不限购，但是商品数量不能超过999个
			if num > 999 {
				products = append(products, saleOrderProductMap[productPackageUuid])
			}
			continue
		}
		if num > float64(limitNum) {
			products = append(products, saleOrderProductMap[productPackageUuid])
		}
	}
	return products
}

func (model *SaleBill) SetNil() {
	model.SaleOrders = nil
	model.H5OrderProducts = nil
	model.SaleBillSetting = nil
	model.Cashier = Staff{}
	model.Desk = nil
	model.BuffetPackage1 = nil
	model.BuffetPackage2 = nil
}

// 设置收银员信息
func (model *SaleBill) SetCashier(dutyNo string, cashierUuid uint64, cashierName string) {
	model.DutyNo = dutyNo
	model.CashierUuid = cashierUuid
	model.CashierName = cashierName
	for _, saleOrder := range model.SaleOrders {
		saleOrder.SetCashier(cashierUuid, cashierName)
	}
}

// 设置打包销售账单。并更新订单的税率，更新订单商品的打包状态
func (model *SaleBill) SetTakeoutSaleBill(diningMethod uint) {
	// 默认堂食
	method := constant.SaleBillDiningMethodDineIn
	// 严谨判断。拒绝非法的值
	if diningMethod == constant.SaleBillDiningMethodTakeout {
		method = constant.SaleBillDiningMethodTakeout
	}
	// 严谨判断。拒绝非法的值
	if diningMethod == constant.SaleBillDiningMethodDineIn {
		method = constant.SaleBillDiningMethodDineIn
	}
	model.DiningMethod = uint(method)
	model.SetUpdate() // 标记要更新model

	// 由于就餐方式改变，导致税率改变，要重新计算账单金额
	// 从ProductPackage中获取税率
	for _, saleOrder := range model.SaleOrders {
		for _, saleOrderProduct := range saleOrder.SaleOrderProducts {
			// 如果商品的就餐方式与账单的就餐方式相同，则不更新商品
			if saleOrderProduct.GetDiningMethod() == uint(method) {
				continue
			}
			if saleOrderProduct.ProductPackage == nil {
				// 这种情况不应该出现，因为商品包是必填的.panic是为了提示没有预加载这个表
				panic("saleOrderProduct.ProductPackage is nil")
			}
			taxRate := saleOrderProduct.ProductPackage.TaxRate(model.DiningMethod)
			saleOrderProduct.SetTaxRate(taxRate)
			// 设置打包状态
			if method == constant.SaleBillDiningMethodTakeout {
				saleOrderProduct.SetWrap()
				saleOrderProduct.UpdateSign() // 重新生成签名
			} else {
				saleOrderProduct.SetUnwrap()
				saleOrderProduct.UpdateSign() // 重新生成签名
			}
		}
	}
}

// 设置反结账
func (model *SaleBill) SetReverseSettle() {
	// 销售账单状态变为待付款
	// 销售订单状态变为未结账状态
	// 销售订单的所有付款单都退款，并生成退款单
	model.Status = constant.SaleBillStatusPending
	model.FinishTime = 0                                    // 反结账后支付时间finish_time置0
	model.ReverseSettleCount = model.ReverseSettleCount + 1 // 反结账次数+1
	for _, saleOrder := range model.SaleOrders {
		saleOrder.FinishTime = 0 // 反结账后支付时间finish_time置0
		saleOrder.Status = constant.SaleOrderStatusPending
		for _, paymentOrder := range saleOrder.PaymentOrders {
			if !paymentOrder.PaymentMethod.IsLianLianPay() {
				paymentOrder.DeleteTime = time.Now().Unix()
				paymentOrder.Status = constant.PaymentOrderStatusRefund
				paymentOrder.RefundOrder = paymentOrder.NewRefundOrder() // 生成退款单
			}
		}
	}
}

// 设置销售账单完成
func (model *SaleBill) SetFinishSaleBill(dutyNo string, cashierUuid uint64, cashierName string) {
	model.Status = constant.SaleBillStatusComplete
	model.FinishTime = time.Now().Unix()
	model.SetCashier(dutyNo, cashierUuid, cashierName)
}

// 设置自助餐套餐
func (model *SaleBill) SetBuffetPackage(buffetPackageUuids []uint64) {
	if len(buffetPackageUuids) == 1 {
		model.BuffetPackage1Uuid = buffetPackageUuids[0]
		model.BuffetPackage2Uuid = 0
	}
	if len(buffetPackageUuids) == 2 {
		model.BuffetPackage1Uuid = buffetPackageUuids[0]
		model.BuffetPackage2Uuid = buffetPackageUuids[1]
	}
}

// 设置人数
func (model *SaleBill) SetMealNum(mealNum uint) {
	model.MealNum = mealNum
}

// 设置加钟人数
func (model *SaleBill) SetDelayProductMealNum(mealNum uint) {
	for _, saleOrder := range model.SaleOrders {
		for _, delayProduct := range saleOrder.SaleOrderBuffetDelayProducts {
			delayProduct.Num = mealNum
		}
	}
}

// 设置隐藏必点方案
func (model *SaleBill) SetHideMustPlan() {
	model.ShowMustPlan = constant.SaleBillShowMustPlanNo
}

// 设置账单为已送厨状态。如果状态已经是送厨，则不修改
func (model *SaleBill) SetCookingStatus() {
	if model.IsCookingStatus() {
		return
	}
	model.SetUpdate()
	model.ProductionTime = time.Now().Unix()
}

// 设置显示销售账单(取单)
func (model *SaleBill) SetShowSaleBill(deviceUuid uint64) {
	model.HideBillTime = 0
	model.DeviceUuid = deviceUuid
}

// 设置隐藏销售账单(挂单)
func (model *SaleBill) SetHideSaleBill() {
	model.HideBillTime = time.Now().Unix()
}

// SetProductFields 动态更新商品多个字段值
func (model *SaleBill) SetProductFields(saleOrderProductUuid uint64, updateProduct SaleOrderProduct, specialFields ...map[string]bool) {
	for i, order := range model.SaleOrders {
		for j, product := range order.SaleOrderProducts {
			if product.Uuid == saleOrderProductUuid {
				product := model.SaleOrders[i].SaleOrderProducts[j]
				var fields map[string]bool
				if len(specialFields) > 0 {
					fields = specialFields[0]
				}
				product.SetFields(updateProduct, fields)
				product.SetUpdate()
				break
			}
		}
	}
}

// SetAllDiscountCancel 设置整单折扣取消
func (model *SaleBill) SetAllDiscountCancel() bool {
	isChange := false
	for _, saleOrder := range model.SaleOrders {
		isChange = saleOrder.SetAllDiscountCancel() || isChange
	}
	return isChange
}

// 将销售订单商品列表中的“未下单h5商品”变为“h5已下单商品”
func (model *SaleBill) SetH5OrderProduct(h5OrderUuid uint64) {
	// 获取未下单的h5订单商品
	h5OrderProducts := model.GetUnOrderH5OrderProduct()
	// 遍历销售订单商品列表
	for _, saleOrderProduct := range h5OrderProducts {
		saleOrderProduct.SetH5OrderProduct(h5OrderUuid) // 将未下单的h5订单商品变为已下单的h5订单商品
	}
}

// 设置自助餐开启时间
func (model *SaleBill) SetBuffetStartTimeAndDuration(maxTimeLimit int) {
	newStartTime := time.Now().Unix()
	// 维护套餐时长和加钟时长
	if maxTimeLimit == -1 {
		model.BuffetDuration = 0
		model.BuffetStartTime = newStartTime
		model.DelayDuration = uint(model.GetRemainingDelayDuration())
		model.DelayStartTime = newStartTime
	} else {
		// 重新计算开启时间 - 如果已经超时了，则开启时间为当前时间减去上一个套餐的限制时长（已用时长）
		if model.BuffetIsTimeOut() || model.BuffetDuration == 0 {
			model.BuffetStartTime = newStartTime - int64(model.BuffetDuration)
		}
		// 永远等于最大时间限制 + 总加钟时间
		model.BuffetDuration = uint(int64(maxTimeLimit))
		// 重新计算加钟的启动时长
		model.DelayDuration = uint(model.GetRemainingDelayDuration())
		if model.BuffetIsTimeOut() || model.BuffetDuration == 0 {
			model.DelayStartTime = newStartTime
		} else {
			model.DelayStartTime = model.GetBuffetEndTime()
		}
	}
}

// SetIsSplitOrder 设置拆单
func (model *SaleBill) SetIsSplitOrder(isSplitOrder bool) {
	model.IsSplitOrder = uint(utils.IfInt(isSplitOrder, constant.SaleBillIsSplitOrderYes, constant.SaleBillIsSplitOrderNo))
}

// GetBuffetSaleNum 获取自助餐销售数量
func (model *SaleBill) GetBuffetSaleNum(buffetPackageUuid uint64) uint {
	num := uint(0)
	for _, saleOrder := range model.SaleOrders {
		for _, customerType := range saleOrder.SaleOrderBuffetCustomerTypes {
			if customerType.BuffetPackageUuid == buffetPackageUuid {
				num += customerType.Num
			}
		}
	}
	return num
}
