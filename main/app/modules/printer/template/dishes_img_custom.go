// Package template 提供打印模板相关功能
package template

import (
	"fmt"
	"strings"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/modules/printer/pkg"
	"ttpos-server-go/app/modules/printer/printer_model"
	"ttpos-server-go/app/modules/printer/template_struct"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// dishesImgTemplateCustom 图片打印模板
type dishesImgTemplateCustom struct {
	base *printerTemplate
}

// NewDishesImgTemplateCustom 创建新的图片订单打印模板
func NewDishesImgTemplateCustom(
	base *printerTemplate,
) *dishesImgTemplateCustom {
	return &dishesImgTemplateCustom{
		base: base,
	}
}

// getData 构建订单数据结构体
func (t *dishesImgTemplateCustom) getData(
	printer *model.Printer,
	saleBill model.SaleBill,
	saleOrder model.SaleOrder,
	orderName string,
	mealNumStr string,
	products []template_struct.StatementProductData,
	productNum decimal.Decimal,
) *template_struct.StatementOrderData {
	// 构建订单数据结构体
	return &template_struct.StatementOrderData{
		BrandName: config.Server.BrandName,
		Store: template_struct.StatementStoreData{
			Name:             t.base.StoreSetting.Name,
			Address:          t.base.StoreSetting.Address,
			Phone:            t.base.StoreSetting.Phone,
			Logo:             "",
			Company:          t.base.StoreSetting.Company,
			CompanyAddr:      t.base.StoreSetting.Address,
			CompanyPhone:     t.base.StoreSetting.Phone,
			CompanyTaxNumber: t.base.StoreSetting.TaxNumber,
			CashierSn:        t.base.GetCashierSn(printer.SourceDeviceSn),
			PrinterSn:        printer.Sn,
		},
		Order: template_struct.StatementOrderInfoData{
			Status: saleOrder.Status,
			SerialNo: func() string {
				if saleBill.DeskUuid > 0 {
					return fmt.Sprintf("%s: %s%s%s", t.base.Translate("桌号"), saleBill.SerialNo, orderName, mealNumStr)
				} else if saleBill.IsTakeoutBill() {
					return t.base.Translate("外送") + ": " + saleBill.SerialNo
				} else {
					return fmt.Sprintf("%s: %s%s", t.base.Translate("取单号"), saleBill.SerialNo, orderName)
				}
			}(),
			OrderNo:     saleOrder.OrderNo,
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
			PaymentMethods:     []template_struct.StatementPaymentMethod{},
			PercentageLists:    []template_struct.StatementPercentageData{},
			IsFree:             saleOrder.IsFreeSaleOrder(),
			// 会员
			IsMember:               saleOrder.Member != nil,
			MemberRemainingBalance: "0",
			MemberPoints:           0,
			//
			PaymentName:   "",
			PaymentQrcode: "",
			//
			Barcode: saleOrder.OrderNo,
		},
	}
}

// ImgPrint 图片打印
func (t *dishesImgTemplateCustom) GetCompleteOrderPrintContent(
	printer *model.Printer,
	tmpData string,
	saleBill model.SaleBill,
	saleOrder model.SaleOrder,
	printer_products printer_model.Products,
) string {
	// 订单名称
	orderName := saleOrder.GetOrderName()
	// 就餐人数
	mealNumStr := utils.IfString(saleBill.MealNum > 0, fmt.Sprintf(" (%d%s)", saleBill.MealNum, t.base.Translate("人")), "")
	// 商品数量
	productNum := decimal.NewFromFloat(0)

	// 商品列表
	products := []template_struct.StatementProductData{}
	var processProducts func(printer_products printer_model.Products, isSubProduct bool)
	processProducts = func(printer_products printer_model.Products, isSubProduct bool) {
		for _, product := range printer_products {
			products = append(products, template_struct.StatementProductData{
				Name:     product.ProductName.GetLocale(t.base.Lang),
				PriceNum: fmt.Sprintf("%s*%v", t.base.Amount(product.TotalNum), product.ProductPrice),
				Price:    t.base.Amount(product.ProductPrice),
				Num:      product.TotalNum,
				Subtotal: t.base.Amount(product.TotalPrice),
				Remark:   product.Remark,
				Attrs:    product.ProductAttr.GetLocale(t.base.Lang),
				Attr:     product.Attr.GetLocale(t.base.Lang),
				AttrList: func() []template_struct.StatementProductDataAttrList {
					attrs := []template_struct.StatementProductDataAttrList{}
					for _, attr := range product.AttrList {
						attrs = append(attrs, template_struct.StatementProductDataAttrList{
							Name: attr.GetLocale(t.base.Lang),
							Text: attr.GetLocale(t.base.Lang),
						})
					}
					return attrs
				}(),
				FlavorName: product.FlavorName.GetLocale(t.base.Lang),
				SauceNames: func() string {
					sauces := []string{}
					for _, sauce := range product.ProductSauceNamesList {
						sauces = append(sauces, sauce.GetLocale(t.base.Lang))
					}
					return strings.Join(sauces, ";")
				}(),
				SauceList: func() []template_struct.StatementProductDataSauceList {
					sauces := []template_struct.StatementProductDataSauceList{}
					for _, sauce := range product.ProductSauceNamesList {
						sauces = append(sauces, template_struct.StatementProductDataSauceList{
							Name: sauce.GetLocale(t.base.Lang),
							Text: sauce.GetLocale(t.base.Lang),
						})
					}
					return sauces
				}(),
				IsGift:          product.IsGift,
				IsPackage:       product.IsWrap,
				IsSubProduct:    isSubProduct,
				IsBuffetProduct: product.IsBuffet,
			})

			// 打印套餐子商品
			if len(product.SubProducts) > 0 {
				processProducts(product.SubProducts, true)
			}
		}
	}
	processProducts(printer_products, false)

	// 构建订单数据结构体
	statementData := t.getData(printer, saleBill, saleOrder, orderName, mealNumStr, products, productNum)

	// 将结构体转换为map
	dataMap, _ := utils.StrToMap(utils.ToJsonString(statementData))

	// 创建解析器
	parser, err := pkg.NewImgTemplateParser(pkg.ImgBaseData{
		Language:             t.base.Lang,
		CurrencyUnit:         t.base.CurrencyUnit,
		CurrencyUnitPosition: t.base.CurrencyUnitPosition,
	}, tmpData, dataMap)
	if err != nil {
		fmt.Println("复杂模板创建解析器失败", err)
		logger.Logger.Error("复杂模板创建解析器失败", zap.Error(err))
		return ""
	}

	// 验证模板
	if err = parser.ValidateTemplate(); err != nil {
		fmt.Println("复杂模板验证失败", err)
		logger.Logger.Error("复杂模板验证失败", zap.Error(err))
		return ""
	}

	// 解析模板 - 开始计时
	img, err := parser.Parse()
	if err != nil {
		fmt.Println("复杂模板解析失败", err)
		logger.Logger.Error("复杂模板解析失败", zap.Error(err))
		return ""
	}

	// 保存图片
	return img.Save("", !t.base.IsSunMi && printer.IsEnableSound(), 0)
}

// ImgPrint 图片打印
func (t *dishesImgTemplateCustom) GetOneDishOneOrderPrintContent(
	productPrinter model.ProductPrinter,
	printer *model.Printer,
	tmpData string,
	saleBill model.SaleBill,
	saleOrder model.SaleOrder,
	printer_products printer_model.Products,
) string {
	// 订单名称
	orderName := saleOrder.GetOrderName()
	// 就餐人数
	mealNumStr := utils.IfString(saleBill.MealNum > 0, fmt.Sprintf(" (%d%s)", saleBill.MealNum, t.base.Translate("人")), "")
	// 商品数量
	productNum := decimal.NewFromFloat(0)

	// 商品列表
	products := []template_struct.StatementProductData{}
	var processProducts func(printer_products printer_model.Products, isSubProduct bool)
	processProducts = func(printer_products printer_model.Products, isSubProduct bool) {
		for _, product := range printer_products {
			exportation := func(num float64) {
				products = append(products, template_struct.StatementProductData{
					Name:     product.ProductName.GetLocale(t.base.Lang),
					PriceNum: fmt.Sprintf("%s*%v", t.base.Amount(num), product.ProductPrice),
					Price:    t.base.Amount(product.ProductPrice),
					Num:      num,
					Subtotal: t.base.Amount(product.TotalPrice),
					Remark:   product.Remark,
					Attrs:    product.ProductAttr.GetLocale(t.base.Lang),
					Attr:     product.Attr.GetLocale(t.base.Lang),
					AttrList: func() []template_struct.StatementProductDataAttrList {
						attrs := []template_struct.StatementProductDataAttrList{}
						for _, attr := range product.AttrList {
							attrs = append(attrs, template_struct.StatementProductDataAttrList{
								Name: attr.GetLocale(t.base.Lang),
								Text: attr.GetLocale(t.base.Lang),
							})
						}
						return attrs
					}(),
					FlavorName: product.FlavorName.GetLocale(t.base.Lang),
					SauceNames: func() string {
						sauces := []string{}
						for _, sauce := range product.ProductSauceNamesList {
							sauces = append(sauces, sauce.GetLocale(t.base.Lang))
						}
						return strings.Join(sauces, ";")
					}(),
					SauceList: func() []template_struct.StatementProductDataSauceList {
						sauces := []template_struct.StatementProductDataSauceList{}
						for _, sauce := range product.ProductSauceNamesList {
							sauces = append(sauces, template_struct.StatementProductDataSauceList{
								Name: sauce.GetLocale(t.base.Lang),
								Text: sauce.GetLocale(t.base.Lang),
							})
						}
						return sauces
					}(),
					IsGift:          product.IsGift,
					IsPackage:       product.IsWrap,
					IsSubProduct:    isSubProduct,
					IsBuffetProduct: product.IsBuffet,
				})
			}
			// 根据打印选择执行打印
			if productPrinter.PrintModeScene == 1 && product.NumType == 0 {
				for i := 0; i < int(product.TotalNum); i++ {
					exportation(1.0)
				}
			} else {
				exportation(product.TotalNum)
			}
		}
	}
	processProducts(printer_products, false)

	// 构建订单数据结构体
	statementData := t.getData(printer, saleBill, saleOrder, orderName, mealNumStr, products, productNum)

	// 将结构体转换为map
	dataMap, _ := utils.StrToMap(utils.ToJsonString(statementData))

	// 创建解析器
	parser, err := pkg.NewImgTemplateParser(pkg.ImgBaseData{
		Language:             t.base.Lang,
		CurrencyUnit:         t.base.CurrencyUnit,
		CurrencyUnitPosition: t.base.CurrencyUnitPosition,
	}, tmpData, dataMap)
	if err != nil {
		fmt.Println("复杂模板创建解析器失败", err)
		logger.Logger.Error("复杂模板创建解析器失败", zap.Error(err))
		return ""
	}

	// 验证模板
	if err = parser.ValidateTemplate(); err != nil {
		fmt.Println("复杂模板验证失败", err)
		logger.Logger.Error("复杂模板验证失败", zap.Error(err))
		return ""
	}

	// 解析模板 - 开始计时
	img, err := parser.Parse()
	if err != nil {
		fmt.Println("复杂模板解析失败", err)
		logger.Logger.Error("复杂模板解析失败", zap.Error(err))
		return ""
	}

	// 保存图片
	return img.Save("", !t.base.IsSunMi && printer.IsEnableSound(), 0)
}
