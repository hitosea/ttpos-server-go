// Package template 提供打印模板相关功能
package template

import (
	"fmt"
	"strings"

	"ttpos-server-go/app/constant"
	settingResp "ttpos-server-go/app/dto/resp/setting"
	"ttpos-server-go/app/modules/printer/pkg"
	"ttpos-server-go/app/modules/printer/template_struct"
	takeoutModel "ttpos-server-go/app/modules/takeout/domain/model"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/language"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"go.uber.org/zap"
)

// platformTakeoutImgTemplate 外卖平台订单图片打印模板
type platformTakeoutImgTemplate struct {
	base *printerTemplate
}

// NewPlatformTakeoutImgTemplate 创建新的外卖平台订单图片打印模板
func NewPlatformTakeoutImgTemplate(
	base *printerTemplate,
) *platformTakeoutImgTemplate {
	return &platformTakeoutImgTemplate{
		base: base,
	}
}

// GetPrintContent 获取打印内容
func (t *platformTakeoutImgTemplate) GetPrintContent(
	settingPrinterInfo settingResp.PrinterInfo,
	tmpData string,
	order *takeoutModel.TakeoutOrder,
	is58mmPrinter bool,
) string {
	// 构建打印数据结构
	printData := t.buildPrintData(settingPrinterInfo, order)

	// 将结构体转换为map
	dataMap, err := utils.StrToMap(utils.ToJsonString(printData))
	if err != nil {
		logger.Logger.Error("外卖平台订单打印数据转换为map失败", zap.Error(err))
		return ""
	}

	// 创建解析器
	parser, err := pkg.NewImgTemplateParser(pkg.ImgBaseData{
		Language:             t.base.Lang,
		CurrencyUnit:         t.base.CurrencyUnit,
		CurrencyUnitPosition: t.base.CurrencyUnitPosition,
		Is58mmPrinter:        is58mmPrinter,
	}, tmpData, dataMap)
	if err != nil {
		logger.Logger.Error("创建外卖平台订单打印解析器失败", zap.Error(err))
		return ""
	}

	// 验证模板
	err = parser.ValidateTemplate()
	if err != nil {
		logger.Logger.Error("外卖平台订单打印模板验证失败", zap.Error(err))
		return ""
	}

	// 解析模板
	img, err := parser.Parse()
	if err != nil {
		logger.Logger.Error("外卖平台订单打印模板解析失败", zap.Error(err))
		return ""
	}

	// TODO: 临时处理，后续需要去掉
	if config.Server.Mode == constant.ServerModeDebug {
		img.SegmentationHeight = 200000
	}

	// 保存图片
	return img.Save("", !t.base.IsSunMi && settingPrinterInfo.IsEnableSound(), 0)
}

// buildPrintData 构建打印数据
func (t *platformTakeoutImgTemplate) buildPrintData(
	settingPrinterInfo settingResp.PrinterInfo,
	order *takeoutModel.TakeoutOrder,
) *template_struct.StatementOrderData {
	// 店铺信息
	storeData := template_struct.StatementStoreData{
		Name:             t.base.StoreSetting.Name,
		Logo:             t.base.GetLogoAddr(),
		Company:          t.base.StoreSetting.Company,
		Address:          t.base.StoreSetting.Address,
		Phone:            t.base.StoreSetting.Phone,
		CompanyAddr:      t.base.StoreSetting.Address,
		CompanyPhone:     t.base.StoreSetting.Phone,
		CompanyTaxNumber: t.base.StoreSetting.TaxNumber,
		CashierSn:        t.base.GetCashierSn(settingPrinterInfo.PrinterCashierDeviceSn),
		PrinterSn:        settingPrinterInfo.PrinterSn,
	}

	// 订单信息
	orderData := t.buildOrderData(order)

	return &template_struct.StatementOrderData{
		BrandName: config.Server.BrandName,
		Store:     storeData,
		Order:     orderData,
	}
}

// buildOrderData 构建订单数据
func (t *platformTakeoutImgTemplate) buildOrderData(order *takeoutModel.TakeoutOrder) template_struct.StatementOrderInfoData {
	orderData := template_struct.StatementOrderInfoData{
		Platform:   order.Platform,
		OrderNo:    order.ShortOrderNumber,
		OrderType:  order.OrderType,
		CreateTime: t.base.FormatUnixTimeDefault(order.OrderTime),     // 下单时间
		FinishTime: t.base.FormatUnixTimeDefault(order.CompletedTime), // 完成时间
		PayTime:    t.base.FormatUnixTimeDefault(order.SubmitTime),    // 支付时间
	}

	// 支付时间
	if order.SubmitTime > 0 && order.SubmitTime != order.OrderTime {
		orderData.PayTime = t.base.FormatUnixTimeDefault(order.SubmitTime)
	}

	// 商品列表
	products := make([]template_struct.StatementProductData, 0, len(order.TakeoutOrderItems))
	productNum := 0.0

	for _, item := range order.TakeoutOrderItems {
		product := template_struct.StatementProductData{
			Name:     language.JsonToLocaleResponse(item.ItemName).GetLocale(t.base.Lang),
			Price:    t.base.Amount(item.Price), // 使用 base.Amount 添加千分位
			Num:      float64(item.Quantity),
			PriceNum: fmt.Sprintf("%s*%d", t.base.Amount(item.Price), item.Quantity),
			Subtotal: t.base.Amount(item.Price * float64(item.Quantity)),
			Remark:   item.Specifications, // 添加备注字段
		}

		// 修饰符信息
		var subProducts []template_struct.StatementProductData // 子商品列表
		if len(item.TakeoutOrderItemModifiers) > 0 {
			// 分类存储不同类型的修饰符
			flavorNames := make([]string, 0) // 规格
			sauceNames := make([]string, 0)  // 加料
			attrNames := make([]string, 0)   // 属性

			for _, modifier := range item.TakeoutOrderItemModifiers {
				if modifier.ModifierName == "" {
					continue
				}

				modifierName := language.JsonToLocaleResponse(modifier.ModifierName).GetLocale(t.base.Lang)

				// 根据修饰符类型分类
				switch modifier.TtposModifierType {
				case "commodity": // 商品类型，作为套餐子商品显示
					subProduct := template_struct.StatementProductData{
						Name:         modifierName,
						Price:        t.base.Amount(modifier.Price),
						Num:          float64(modifier.Quantity),
						PriceNum:     fmt.Sprintf("%d", modifier.Quantity),
						Subtotal:     t.base.Amount(modifier.Price * float64(modifier.Quantity)),
						IsSubProduct: true, // 标记为子商品
					}
					subProducts = append(subProducts, subProduct)
				case "flavor": // 规格
					flavorNames = append(flavorNames, modifierName)
				case "sauce": // 加料
					sauceNames = append(sauceNames, modifierName)
				case "attr": // 属性
					attrNames = append(attrNames, modifierName)
				default:
					// 其他类型归到加料中显示
					sauceNames = append(sauceNames, modifierName)
				}
			}

			// 设置规格、加料、属性
			if len(flavorNames) > 0 {
				product.FlavorName = strings.Join(flavorNames, ", ") // 规格
			}
			if len(sauceNames) > 0 {
				product.SauceNames = strings.Join(sauceNames, ", ") // 加料
			}
			if len(attrNames) > 0 {
				product.Attrs = strings.Join(attrNames, ", ") // 属性
			}
		}

		// 先添加主商品
		products = append(products, product)
		productNum += float64(item.Quantity)

		// 再添加子商品（紧跟在主商品后面）
		for _, subProduct := range subProducts {
			products = append(products, subProduct)
			productNum += subProduct.Num
		}
	}

	orderData.Products = products
	orderData.ProductNum = productNum

	// 价格信息
	orderData.ProductAmount = t.base.Amount(order.Subtotal)
	orderData.DeliveryFee = t.base.Amount(order.DeliveryFee)

	// 计算优惠金额
	totalDiscount := order.PlatformDiscount + order.MerchantDiscount + order.BasketPromo
	orderData.DiscountFee = t.base.Amount(totalDiscount)

	orderData.ActualReceivePrice = t.base.Amount(order.EaterPayment)

	// 支付信息
	orderData.PaymentName = order.PaymentType
	orderData.PaidAmount = t.base.Amount(order.EaterPayment)

	// 收货人信息
	if order.TakeoutOrderReceiver != nil {
		orderData.CustomerName = order.TakeoutOrderReceiver.ReceiverName
		orderData.CustomerPhone = order.TakeoutOrderReceiver.ReceiverPhones
		orderData.CustomerAddress = order.TakeoutOrderReceiver.Address
	}

	// 异常提示
	if order.IsAbnormal == 1 {
		orderData.WarningMessage = order.AbnormalDetail
	}

	return orderData
}
