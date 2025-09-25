// Package template 提供打印模板相关功能
package template

import (
	"fmt"
	settingResp "ttpos-server-go/app/dto/resp/setting"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/printer/pkg"
	"ttpos-server-go/app/printer/printer_model"
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

// ImgPrint 图片打印
func (t *dishesImgTemplateCustom) GetCompleteOrderPrintContent(
	settingPrinterInfo settingResp.PrinterInfo,
	tmpData string,
	saleBill *model.SaleBill,
	products printer_model.Products,
	payMethodUuid uint64,
) string {

	// 将结构体转换为map
	dataMap, _ := utils.StrToMap(utils.ToJsonString(products))

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
	err = parser.ValidateTemplate()
	if err != nil {
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
	return img.Save("", !t.base.IsSunMi && settingPrinterInfo.IsEnableSound(), 0)
}
