// Package template 提供打印模板相关功能
package template

import (
	"fmt"
	"strconv"
	"ttpos-server-go/app/constant"
	settingResp "ttpos-server-go/app/dto/resp/setting"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/printer/pkg"
	"ttpos-server-go/app/printer/template_struct"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// invoiceImgTemplateCustom 图片发票打印模板
type invoiceImgTemplateCustom struct {
	base *printerTemplate
}

// NewInvoiceImgTemplateCustom 创建新的图片发票打印模板
func NewInvoiceImgTemplateCustom(
	base *printerTemplate,
) *invoiceImgTemplateCustom {
	return &invoiceImgTemplateCustom{
		base: base,
	}
}

// GetPrintContent 图片打印
func (t *invoiceImgTemplateCustom) GetPrintContent(
	settingPrinterInfo settingResp.PrinterInfo,
	tmpData string,
	saleBill *model.SaleBill,
	saleOrder *model.SaleOrder,
	is58mmPrinter bool,
) string {
	// 订单名称
	orderName := saleOrder.GetOrderName()
	// 就餐人数
	mealNumStr := utils.IfString(saleBill.MealNum > 0, fmt.Sprintf(" (%d%s)", saleBill.MealNum, t.base.Translate("人")), "")
	// 商品数量
	productNum := decimal.NewFromFloat(0)

	// 商品列表
	products := []template_struct.StatementProductData{}

	// 自助餐顾客类型
	for _, orderBuffetCustomer := range saleOrder.SaleOrderBuffetCustomerTypes {
		if orderBuffetCustomer.IsDelete() {
			continue
		}
		productNum = productNum.Add(decimal.NewFromFloat(float64(orderBuffetCustomer.Num)).Round(3))
		originPrice := orderBuffetCustomer.GetOriginPrice()
		products = append(products, template_struct.StatementProductData{
			Name:            orderBuffetCustomer.BuffetPackage.MultiLanguageName.GetNameByLang(t.base.Lang),
			PriceNum:        fmt.Sprintf("%s*%d", t.base.Amount(orderBuffetCustomer.SalePrice), orderBuffetCustomer.Num),
			Price:           t.base.Amount(originPrice),
			Num:             float64(orderBuffetCustomer.Num),
			Subtotal:        t.base.Amount(orderBuffetCustomer.TotalPrice),
			Attrs:           orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name,
			Attr:            orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name,
			SauceNames:      "",
			IsBuffet:        true,
			IsBuffetProduct: false,
			IsGift:          false,
			IsPackage:       false,
			IsSubProduct:    false,
		})
	}

	// 添加加钟商品
	buffetDelayProducts, num := t.base.MergeSaleOrderBuffetDelayProducts(saleOrder)
	productNum = productNum.Add(decimal.NewFromFloat(num).Round(3))
	for _, delay := range buffetDelayProducts {
		products = append(products, template_struct.StatementProductData{
			Name:            delay.DelayName,
			PriceNum:        fmt.Sprintf("%s*%d", t.base.Amount(delay.DelayPrice), delay.DelayNum),
			Price:           t.base.Amount(delay.DelayPrice),
			Num:             float64(delay.DelayNum),
			Subtotal:        t.base.Amount(delay.DelayTotalPrice),
			Attrs:           delay.DelayName,
			Attr:            delay.DelayName,
			SauceNames:      "",
			IsDelay:         true,
			IsBuffet:        false,
			IsBuffetProduct: false,
			IsGift:          false,
			IsPackage:       false,
			IsSubProduct:    false,
		})
	}

	// 商品列表
	mergeProducts, num := t.base.MergeSaleOrderProduct(MergeSaleOrderProductOptions{
		saleBill:   saleBill,
		saleOrder:  saleOrder,
		IsShowSku:  true,
		IsShowWrap: true,
	})
	productNum = productNum.Add(decimal.NewFromFloat(num).Round(3))
	for _, product := range mergeProducts {
		products = append(products, template_struct.StatementProductData{
			Name:     product.Name,
			PriceNum: fmt.Sprintf("%s*%v", t.base.Amount(product.ProductPrice), product.ProductNum),
			Price:    t.base.Amount(product.ProductPrice),
			Num:      product.ProductNum,
			Subtotal: t.base.Amount(product.ProductTotalPrice),
			Attrs:    product.Attrs,
			Attr:     product.Attr,
			AttrList: func() []template_struct.StatementProductDataAttrList {
				attrs := []template_struct.StatementProductDataAttrList{}
				for _, attr := range product.AttrList {
					attrs = append(attrs, template_struct.StatementProductDataAttrList{
						Name: attr,
						Text: attr,
					})
				}
				return attrs
			}(),
			FlavorName: product.FlavorName,
			SauceList: func() []template_struct.StatementProductDataSauceList {
				sauces := []template_struct.StatementProductDataSauceList{}
				for _, sauce := range product.SauceList {
					sauces = append(sauces, template_struct.StatementProductDataSauceList{
						Name: sauce,
						Text: sauce,
					})
				}
				return sauces
			}(),
			SauceNames:      product.SauceNames,
			IsGift:          product.IsGift,
			IsPackage:       product.IsWrap,
			IsSubProduct:    product.IsSub,
			IsBuffetProduct: product.IsBuffetProduct,
			Remark:          product.Remark,
		})
		// 套餐子商品
		for _, subProduct := range product.SubProducts {
			products = append(products, template_struct.StatementProductData{
				Name:     subProduct.Name,
				PriceNum: fmt.Sprintf("%v", subProduct.ProductNum),
				Price:    t.base.Amount(subProduct.ProductPrice),
				Num:      subProduct.ProductNum,
				Subtotal: t.base.Amount(subProduct.ProductTotalPrice),
				Attrs:    subProduct.Attrs,
				Attr:     subProduct.Attr,
				AttrList: func() []template_struct.StatementProductDataAttrList {
					attrs := []template_struct.StatementProductDataAttrList{}
					for _, attr := range subProduct.AttrList {
						attrs = append(attrs, template_struct.StatementProductDataAttrList{
							Name: attr,
							Text: attr,
						})
					}
					return attrs
				}(),
				SauceList: func() []template_struct.StatementProductDataSauceList {
					sauces := []template_struct.StatementProductDataSauceList{}
					for _, sauce := range subProduct.SauceList {
						sauces = append(sauces, template_struct.StatementProductDataSauceList{
							Name: sauce,
							Text: sauce,
						})
					}
					return sauces
				}(),
				FlavorName:      subProduct.FlavorName,
				SauceNames:      subProduct.SauceNames,
				IsGift:          false,
				IsPackage:       false,
				IsSubProduct:    true,
				IsBuffetProduct: false,
			})
		}
	}

	// 支付方式
	paymentMethods := []template_struct.StatementPaymentMethod{}
	if saleOrder.IsFreeSaleOrder() {
		paymentMethods = append(paymentMethods, template_struct.StatementPaymentMethod{
			Name: t.base.Translate("支付方式"),
			Text: t.base.Translate("免单"),
		})
		paymentMethods = append(paymentMethods, template_struct.StatementPaymentMethod{
			Name: t.base.Translate("实收金额"),
			Text: t.base.Amount(0),
		})
	}
	if len(saleOrder.PaymentOrders) > 0 {
		for _, paymentOrder := range saleOrder.PaymentOrders {
			paymentMethods = append(paymentMethods, template_struct.StatementPaymentMethod{
				Name: t.base.Translate("支付方式"),
				Text: paymentOrder.PaymentMethod.GetName(),
			})
			paymentMethods = append(paymentMethods, template_struct.StatementPaymentMethod{
				Name: t.base.Translate("实收金额"),
				Text: t.base.Amount(paymentOrder.Amount),
			})
			if saleOrder.ChangeAmount > 0 && paymentOrder.PaymentMethod.Code == constant.PaymentMethodCodeCash {
				paymentMethods = append(paymentMethods, template_struct.StatementPaymentMethod{
					Name: t.base.Translate("找零"),
					Text: t.base.Amount(saleOrder.ChangeAmount),
				})
			}
			if paymentOrder.TransactionNumber != "" {
				paymentMethods = append(paymentMethods, template_struct.StatementPaymentMethod{
					Name: t.base.Translate("交易单号"),
					Text: paymentOrder.TransactionNumber,
				})
			}
		}
	}

	// 百分比列表
	percentageLists := []template_struct.StatementPercentageData{}
	for _, percentage := range saleOrder.GetPercentageList() {
		taxRate := percentage["TaxRate"]
		taxFee, _ := strconv.ParseFloat(percentage["TaxFee"], 64)
		totalPrice, _ := strconv.ParseFloat(percentage["TotalPrice"], 64)
		percentageLists = append(percentageLists, template_struct.StatementPercentageData{
			TaxRate:    taxRate,
			TaxFee:     t.base.Amount(taxFee),
			TotalPrice: t.base.Amount(totalPrice),
		})
	}

	// 构建订单数据结构体（包含发票数据）
	statementData := &template_struct.StatementOrderData{
		BrandName: config.Server.BrandName,
		Store: template_struct.StatementStoreData{
			Name:             t.base.StoreSetting.Name,
			StoreCode:        t.base.StoreSetting.StoreCode,
			Address:          t.base.StoreSetting.Address,
			Phone:            t.base.StoreSetting.Phone,
			Logo:             t.base.GetLogoAddr(),
			Company:          t.base.StoreSetting.Company,
			CompanyAddr:      t.base.StoreSetting.Address,
			CompanyPhone:     t.base.StoreSetting.Phone,
			CompanyTaxNumber: t.base.StoreSetting.TaxNumber,
			CashierSn:        t.base.GetCashierSn(settingPrinterInfo.PrinterCashierDeviceSn),
			PrinterSn:        settingPrinterInfo.PrinterSn,
		},
		Order: template_struct.StatementOrderInfoData{
			Status:               saleOrder.Status,
			SerialNo:             saleBill.SerialNo,
			IsOrderSourceTakeout: saleBill.IsOrderSourceTakeout(),
			SerialNos: func() string {
				if saleBill.DeskUuid > 0 {
					return fmt.Sprintf("%s: %s%s%s", t.base.Translate("桌号"), saleBill.SerialNo, orderName, mealNumStr)
				} else {
					return fmt.Sprintf("%s: %s%s", t.base.Translate("取单号"), saleBill.SerialNo, orderName)
				}
			}(),
			OrderNo:     saleOrder.OrderNo,
			MealNum:     saleBill.MealNum,
			Remark:      saleBill.Remark,
			CashierName: saleOrder.CashierName,
			FinishTime:  t.base.FormatUnixTimeDefault(saleOrder.FinishTime),
			CreateTime:  t.base.FormatUnixTimeDefault(saleOrder.CreateTime),
			PayTime:     t.base.FormatUnixTimeDefault(saleOrder.FinishTime),
			UpdateTime:  t.base.FormatUnixTimeDefault(saleOrder.UpdateTime),
			// 商品
			Products: products,
			//
			ProductNum:    productNum.InexactFloat64(),
			ProductAmount: t.base.Amount(saleOrder.ProductAmount),
			ServiceFee:    t.base.Amount(saleOrder.ServiceFee),
			TaxFeeType:    saleBill.SaleBillSetting.TaxFeeType,
			IsContainTax: func() uint {
				if saleOrder.TaxFee > 0 && saleBill.SaleBillSetting.TaxFeeType == 1 && (t.base.ConsumptionTax == 1 || t.base.ConsumptionTax == 3) {
					return 1
				}
				if saleOrder.TaxFee > 0 && saleBill.SaleBillSetting.TaxFeeType == 2 && (t.base.ConsumptionTax == 1 || t.base.ConsumptionTax == 2) {
					return 2
				}
				return 0
			}(),
			DiscountFee: t.base.Amount(saleOrder.CustomDiscountFee),
			DiscountRate: func() string {
				// 计算折扣率：折扣金额 / 原始金额 * 100
				discountRate := decimal.NewFromFloat(saleOrder.CustomDiscountFee).Div(decimal.NewFromFloat(saleOrder.ProductOriginalAmount)).Mul(decimal.NewFromInt(100))
				return t.base.Number(discountRate.InexactFloat64())
			}(),
			// 会员
			MemberDiscountFee: t.base.Amount(saleOrder.MemberDiscountFee),
			MemberDiscountRate: func() float64 {
				oldGradeEquity := float64(100)
				gradeEquity := float64(100)
				if saleOrder.MemberDiscountRate != 0 {
					gradeEquity = saleOrder.MemberDiscountRate * 100
					oldGradeEquity = gradeEquity
				}
				if t.base.Lang == "zh" || t.base.Lang == "zhtw" {
					gradeEquity /= 10
				}
				if oldGradeEquity != 100 && gradeEquity > 0 {
					return gradeEquity
				}
				return 0
			}(),
			MemberCardDiscountRate: func() float64 {
				oldCardDiscount := float64(100)
				cardDiscount := float64(100)
				if saleOrder.MemberCardDiscountRate != 0 {
					cardDiscount = saleOrder.MemberCardDiscountRate * 100
					oldCardDiscount = cardDiscount
				}
				if t.base.Lang == "zh" || t.base.Lang == "zhtw" {
					cardDiscount /= 10
				}
				if oldCardDiscount != 100 && cardDiscount > 0 {
					return oldCardDiscount
				}
				return 0
			}(),
			MemberPointsDiscount: func() float64 {
				if saleOrder.PayPointsAmount > 0 && saleOrder.PayPoints > 0 {
					return saleOrder.PayPointsAmount
				}
				return 0
			}(),
			//
			CouponExchangeAmount: saleOrder.CalcCouponExchangeAmount(),
			ActivityAmount:       saleOrder.ActivityAmount,
			CheckOutZeroFee:      t.base.Amount(saleOrder.GetCheckOutZeroFee()),
			ReturnAmount:         t.base.Amount(saleOrder.GetReturnAmount()),
			PaymentCommissionFee: saleOrder.PaymentCommissionFee,
			FreeAmount: func() string {
				if saleOrder.IsFreeSaleOrder() {
					return t.base.Amount(saleOrder.GetAmount())
				}
				return "0"
			}(),
			ActualReceivePrice: t.base.Amount(saleOrder.GetPrintReceivablePrice()),
			PaymentMethods:     paymentMethods,
			PercentageLists:    percentageLists,
			IsFree:             saleOrder.IsFreeSaleOrder(),
			// 会员
			IsMember: saleOrder.Member != nil,
			MemberRemainingBalance: func() string {
				if saleOrder.Member == nil {
					return "0"
				}
				return t.base.Amount(saleOrder.GetMemberSurplusBalance())
			}(),
			MemberPoints: func() float64 {
				if saleOrder.Member == nil {
					return 0
				}
				var rule settingResp.PointsRule
				if !saleOrder.IsPaid() {
					pointsSetting, err := t.base.Setting.GetPointsSetting(t.base.Ctx)
					if err == nil {
						rule = pointsSetting.GetPointsGiftRule(saleBill.IsBuffetSaleBill(), saleOrder.Member.MemberLevelUuid)
					}
				}
				// 计算本单获取的积分
				return saleOrder.GetMemberSurplusPoints(int(saleBill.MealNum), rule)
			}(),
			//
			PaymentName:   "",
			PaymentQrcode: "",
			//
			Barcode: saleOrder.OrderNo,
		},
		// 发票数据
		Invoice: template_struct.StatementInvoiceData{
			InvoiceNumber:     "",
			UnifiedCreditCode: "",
			CompanyName:       saleOrder.InvoiceInfo.CompanyName,
			CompanyAddress:    saleOrder.InvoiceInfo.CompanyAddr,
			CompanyPhone:      saleOrder.InvoiceInfo.CompanyPhone,
			CompanyTaxNumber:  saleOrder.InvoiceInfo.CompanyTaxNumber,
			PrintNum:          saleOrder.InvoiceInfo.PrintNum,
		},
	}

	// 将结构体转换为map
	dataMap, _ := utils.StrToMap(utils.ToJsonString(statementData))

	// 创建解析器
	parser, err := pkg.NewImgTemplateParser(pkg.ImgBaseData{
		Language:             t.base.Lang,
		CurrencyUnit:         t.base.CurrencyUnit,
		CurrencyUnitPosition: t.base.CurrencyUnitPosition,
		Is58mmPrinter:        is58mmPrinter,
	}, tmpData, dataMap)
	if err != nil {
		fmt.Println("发票模板创建解析器失败", err)
		logger.Logger.Error("发票模板创建解析器失败", zap.Error(err))
		return ""
	}

	// 验证模板
	err = parser.ValidateTemplate()
	if err != nil {
		fmt.Println("发票模板验证失败", err)
		logger.Logger.Error("发票模板验证失败", zap.Error(err))
		return ""
	}

	// 解析模板
	img, err := parser.Parse()
	if err != nil {
		fmt.Println("发票模板解析失败", err)
		logger.Logger.Error("发票模板解析失败", zap.Error(err))
		return ""
	}

	// 保存图片
	return img.Save("", !t.base.IsSunMi && settingPrinterInfo.IsEnableSound(), 0)
}
