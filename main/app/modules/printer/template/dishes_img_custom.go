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
	order printer_model.Order,
	mealNumStr string,
	products []template_struct.StatementProductData,
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
			SerialNo: order.SerialNo,
			OrderNo:  order.OrderNo,
			PayTime:  t.base.FormatUnixTimeDefault(order.FinishTime),
			// 商品
			Products: products,
		},
	}
}

// ImgPrint 图片打印
func (t *dishesImgTemplateCustom) GetCompleteOrderPrintContent(
	printer *model.Printer,
	tmpData string,
	order printer_model.Order,
) string {
	// 就餐人数
	mealNumStr := utils.IfString(order.MealNum > 0, fmt.Sprintf(" (%d%s)", order.MealNum, t.base.Translate("人")), "")

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
	processProducts(order.Products, false)

	// 构建订单数据结构体
	statementData := t.getData(printer, order, mealNumStr, products)

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
	order printer_model.Order,
) string {
	// 就餐人数
	mealNumStr := utils.IfString(order.MealNum > 0, fmt.Sprintf(" (%d%s)", order.MealNum, t.base.Translate("人")), "")

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
	processProducts(order.Products, false)

	// 构建订单数据结构体
	statementData := t.getData(printer, order, mealNumStr, products)

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
