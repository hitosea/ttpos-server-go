// Package template 提供打印模板相关功能
package template

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	respSetting "ttpos-server-go/app/dto/resp/setting"
	"ttpos-server-go/app/printer/pkg"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/i18n"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/utils"
)

// 打印类型
const (
	PrinterTypeFeiEYunTag    = "FEI_E_YUN_TAG"  // 飞鹅标签打印机
	PrinterTypePrintCenter   = "PRINT_CENTER"   // 365云打印
	PrinterTypeSunmiLan      = "SUNMI_LAN"      // 商米 局域网内打印
	PrinterTypeSunmiCloud    = "SUNMI_CLOUD"    // 商米 云打印
	PrinterTypeXPrinterLan   = "XPRINTER_LAN"   // 芯烨-有线
	PrinterTypeXPrinterWifi  = "XPRINTER_WIFI"  // 芯烨-WIFI
	PrinterTypeCashierCompax = "CASHIER_COMPAX" // Compax 收银打印机 80mm 自带
	PrinterTypeCashierSunmi  = "CASHIER_SUNMI"  // SUNMI 商米 收银打印机 80mm 自带
	PrinterTypeCodesoftLan   = "CODESOFT_LAN"   // Codesoft（网口）80mm
	PrinterTypeCodesoftWifi  = "CODESOFT_WIFI"  //Codesoft（WIFI）80mm
)

// printerTemplate 模板基类
type printerTemplate struct {
	Ctx context.Context
	// 系统设置
	Lang                 string // 语言
	DefaultCalendar      int    // 日历：1公历，3佛历
	CurrencyUnit         string // 金额单位
	CurrencyUnitPosition int    // 金额单位位置
	ConsumptionTax       int    // 消费税类型
	IsSunMi              bool   // 是否商米打印
	// 系统设置
	Setting         *setting.Srv          // 系统设置
	StoreSetting    *respSetting.Store    // 商城设置
	PrinterSetting  *respSetting.Printer  // 打印设置
	CurrencySetting *respSetting.Currency // 货币设置
}

// NewPrinterTemplate 创建新的基础模板
func NewPrinterTemplate(
	ctx context.Context,
	setting *setting.Srv,
	storeSetting *respSetting.Store,
	printerSetting *respSetting.Printer,
	currencySetting *respSetting.Currency,
	isSunMi bool,
) *printerTemplate {
	// 初始化默认值
	template := &printerTemplate{
		Ctx:             ctx,
		Setting:         setting,
		StoreSetting:    storeSetting,
		PrinterSetting:  printerSetting,
		CurrencySetting: currencySetting,
		//
		IsSunMi:              isSunMi,
		DefaultCalendar:      1,
		CurrencyUnit:         "",
		CurrencyUnitPosition: 0,
		ConsumptionTax:       0,
	}
	// 设置系统设置
	if template.CurrencySetting != nil && template.PrinterSetting != nil {
		// 将string类型转为int类型
		template.DefaultCalendar, _ = strconv.Atoi(template.PrinterSetting.DefaultCalendar)
		template.ConsumptionTax, _ = strconv.Atoi(template.PrinterSetting.ConsumptionTax)
		template.CurrencyUnitPosition, _ = strconv.Atoi(template.CurrencySetting.UnitPosition)
		// 货币单位
		template.CurrencyUnit = ""
		if template.CurrencySetting != nil && template.PrinterSetting.MonetaryUnitOpen != "0" {
			template.CurrencyUnit = template.CurrencySetting.PrintUnit
		}
	}
	// 设置语言 - 使用简化逻辑，后续可接入i18n包
	template.Lang = ctx.GetLanguage()
	// todo - wfs
	// template.Lang = printerSetting.DefaultLanguage
	// template.Lang = printerSetting.KitchenLanguage
	//
	return template
}

// GetPriceAndUnit 获取价格和单位
func (t *printerTemplate) GetPriceAndUnit(price float64) string {
	// 格式化金额为字符串，保留两位小数
	priceStr := t.Amount(price)
	// priceStr := strconv.FormatFloat(price, 'f', 2, 64)
	if t.CurrencyUnitPosition == 1 {
		return priceStr + t.CurrencyUnit
	}
	return t.CurrencyUnit + priceStr
}

// IsMy 判断是否缅甸语
func (t *printerTemplate) IsMy(text string) bool {
	// 缅甸语Unicode范围：\u1000-\u109F
	myanmarRegex := regexp.MustCompile(`[\x{1000}-\x{109F}]`)
	return t.Lang == "my" || myanmarRegex.MatchString(text)
}

// IsMyText 判断文本是否包含缅甸语
func (t *printerTemplate) IsMyText(text string) bool {
	// 缅甸语Unicode范围：\u1000-\u109F
	myanmarRegex := regexp.MustCompile(`[\x{1000}-\x{109F}]`)
	return myanmarRegex.MatchString(text)
}

// IsThText 判断文本是否包含泰语
func (t *printerTemplate) IsThText(text string) bool {
	// 泰语Unicode范围：\u0E00-\u0E7F
	thaiRegex := regexp.MustCompile(`[\x{0E00}-\x{0E7F}]`)
	return thaiRegex.MatchString(text) || text == "฿"
}

func (t *printerTemplate) Translate(message string) string {
	return i18n.Translate(t.Lang, message)
}

// 打印文本（支持可选参数和默认值）
func (t *printerTemplate) PrintText(
	leftText string,
	options ...interface{},
) string {
	// 默认值
	centerText := ""
	rightText := ""
	total := 32
	leftNum := 0
	centerNum := 0
	rightNum := 0
	intervalNum := 4

	// 处理可选参数
	optionsLen := len(options)
	if optionsLen > 0 {
		if v, ok := options[0].(string); ok {
			centerText = v
		}
	}
	if optionsLen > 1 {
		if v, ok := options[1].(string); ok {
			rightText = v
		}
	}
	if optionsLen > 2 {
		if v, ok := options[2].(int); ok {
			total = v
		}
	}
	if optionsLen > 3 {
		if v, ok := options[3].(int); ok {
			leftNum = v
		}
	}
	if optionsLen > 4 {
		if v, ok := options[4].(int); ok {
			centerNum = v
		}
	}
	if optionsLen > 5 {
		if v, ok := options[5].(int); ok {
			rightNum = v
		}
	}
	if optionsLen > 6 {
		if v, ok := options[6].(int); ok {
			intervalNum = v
		}
	}

	return pkg.NewPrintTextHelper().PrintText(leftText, centerText, rightText, total, leftNum, centerNum, rightNum, intervalNum)
}

/*
* 格式化时间戳
 */
func (p *printerTemplate) FormatUnixTimeDefault(datetime int64) string {
	if p.DefaultCalendar == 3 {
		return p.ChangeBuddhistCalendar(datetime)
	}
	return utils.SetTimezone(p.StoreSetting.TimeZone).FormatUnixTimeWithSlash(datetime)
}

/*
* 公历转佛历
* 将公历日期转换为佛历日期
* 参数可以是时间戳（int64）或时间字符串（string）
* 返回同类型的佛历日期
 */
func (p *printerTemplate) ChangeBuddhistCalendar(datetime int64) string {
	// 泰国佛历偏移量（公元543年对应佛历1年）
	thaiOffset := 543
	// 获取时区
	loc, err := time.LoadLocation(p.StoreSetting.TimeZone)
	if err != nil {
		loc = time.Local
	}
	// 将时间戳转换为time.Time
	t := time.Unix(datetime, 0).In(loc)
	// 计算佛历年份
	thaiYear := t.Year() + thaiOffset
	// 格式化为斜杠格式
	result := fmt.Sprintf("%d/%02d/%02d %02d:%02d:%02d", thaiYear, t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
	return result
}

// Amount 计算金额的千分位
func (p *printerTemplate) Amount(amount float64) string {
	if amount == 0 {
		return "0"
	}
	// 先格式化为2位小数
	formattedAmount := strconv.FormatFloat(amount, 'f', 2, 64)

	// 分割整数和小数部分
	parts := strings.Split(formattedAmount, ".")
	integerPart := parts[0]
	decimalPart := ""
	if len(parts) > 1 {
		decimalPart = parts[1]
	}

	// 处理整数部分，添加千分位
	var result strings.Builder
	length := len(integerPart)
	for i := length - 1; i >= 0; i-- {
		result.WriteByte(integerPart[i])
		if i > 0 && (length-i)%3 == 0 {
			result.WriteByte(',')
		}
	}

	// 反转字符串
	reversed := result.String()
	runes := []rune(reversed)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	integerPart = string(runes)

	// 处理小数部分，去除尾部的0
	decimalPart = strings.TrimRight(decimalPart, "0")

	// 组合结果
	if decimalPart != "" {
		return integerPart + "." + decimalPart
	}
	return integerPart
}
