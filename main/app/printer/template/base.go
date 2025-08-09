// Package template 提供打印模板相关功能
package template

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
	"ttpos-server-go/app/dto/resp/business_data_resp"
	respSetting "ttpos-server-go/app/dto/resp/setting"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/printer/pkg"
	"ttpos-server-go/app/printer/printer_model"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/i18n"
	image "ttpos-server-go/pkg/Image"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/utils"

	"github.com/disintegration/imaging"
	"github.com/shopspring/decimal"
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

// 打印业务数据的
type PrintingBusinessData struct {
	All             *business_data_resp.BusinessDataAll             `json:"business_data_all"`              // 全部数据
	PaymentMethod   *business_data_resp.BusinessDataPaymentMethod   `json:"business_data_payment_method"`   // 支付方式数据
	ProductCategory *business_data_resp.BusinessDataProductCategory `json:"business_data_product_category"` // 产品类别数据
	Product         *business_data_resp.BusinessDataProduct         `json:"business_data_product"`          // 产品数据
}

// MergeSaleOrderBuffetDelayProducts 合并加钟商品数据
type MergeSaleOrderBuffetDelayProducts struct {
	DelayName       string  `json:"delay_name"`
	DelayPrice      float64 `json:"delay_price"`
	DelayNum        uint    `json:"delay_num"`
	DelayTotalPrice float64 `json:"delay_total_price"`
}

// MergeSaleOrderProduct 合并销售订单商品数据
type MergeSaleOrderProduct struct {
	ProductName       string                  `json:"product_name"`
	ProductPrice      float64                 `json:"product_price"`
	ProductNum        float64                 `json:"product_num"`
	ProductTotalPrice float64                 `json:"product_total_price"`
	SubProducts       []MergeSaleOrderProduct `json:"sub_products"`
}

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
	opts ...string,
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
		if template.CurrencyUnit == "￥" {
			template.CurrencyUnit = "\xC2\xA5"
		}
	}
	// 设置语言 - 使用简化逻辑，后续可接入i18n包
	if len(opts) > 0 && opts[0] != "" {
		template.Lang = opts[0]
	} else if printerSetting != nil {
		template.Lang = printerSetting.DefaultLanguage
	}
	return template
}

// GetPriceAndUnit 获取价格和单位
func (t *printerTemplate) GetPriceAndUnit(price float64) string {
	// 格式化金额为字符串，保留两位小数
	priceStr := t.Amount(price)
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
	if datetime == 0 {
		return ""
	}
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

// Number 计算数字
func (p *printerTemplate) Number(amount float64) string {
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
	// 处理小数部分，去除尾部的0
	decimalPart = strings.TrimRight(decimalPart, "0")

	// 组合结果
	if decimalPart != "" {
		return integerPart + "." + decimalPart
	}
	return integerPart
}

// FloatToString 浮点数转字符串
func (p *printerTemplate) FloatToString(num float64) string {
	return fmt.Sprintf("%v", num)
}

// GetLogoAddr 获取logo地址
func (p *printerTemplate) GetLogoAddr() string {
	if p.StoreSetting == nil || p.StoreSetting.LogoURL == "" {
		return ""
	}

	logoURL := p.StoreSetting.LogoURL

	// 处理storage.googleapis.com的URL
	if strings.Contains(logoURL, "storage_googleapis") {
		// 解析URL以获取域名和路径
		parsedURL, err := url.Parse(logoURL)
		if err == nil {
			// 替换为新的域名
			parsedURL.Host = "storage.googleapis.com"
			// 修复路径中的双斜杠问题
			parsedURL.Path = strings.Replace(parsedURL.Path, "//storage_googleapis", "", 1)
			parsedURL.Path = strings.Replace(parsedURL.Path, "/storage_googleapis", "", 1)
			logoURL = parsedURL.String()
		}
	}

	// 从 URL 中提取文件名
	urlParts := strings.Split(logoURL, "/")
	fileName := urlParts[len(urlParts)-1]

	// 如果有扩展名，去掉扩展名
	fileNameWithoutExt := strings.Split(fileName, ".")[0]

	// 创建目录
	outputDir := "./tmp/printer/shop/logo"
	os.MkdirAll(outputDir, 0777)

	// 输出文件路径
	outputPath := fmt.Sprintf("%s/%s.png", outputDir, fileNameWithoutExt)

	// 检查文件是否已经存在
	if _, err := os.Stat(outputPath); err == nil {
		// 文件已存在，直接返回
		return outputPath
	}

	// 文件不存在，处理图像，确保是白底黑字
	err := image.NewImageHelp().WhiteBackgroundWithBlackText(logoURL, outputPath)
	if err != nil {
		fmt.Println("处理图像失败:", err)
		return ""
	}

	return outputPath
}

// GetQrcodeAddr 获取二维码地址
func (p *printerTemplate) GetQrcodeAddr(qrcodeURL string) string {
	// 检查是否是URL
	if !strings.HasPrefix(qrcodeURL, "http://") && !strings.HasPrefix(qrcodeURL, "https://") {
		return qrcodeURL
	}

	// 从 URL 中提取文件名
	urlParts := strings.Split(qrcodeURL, "/")
	fileName := urlParts[len(urlParts)-1]

	// 如果有扩展名，去掉扩展名
	fileNameWithoutExt := strings.Split(fileName, ".")[0]

	// 创建目录
	outputDir := "./tmp/printer/shop/qrcode"
	os.MkdirAll(outputDir, 0777)

	// 输出文件路径
	outputPath := fmt.Sprintf("%s/%s.png", outputDir, fileNameWithoutExt)

	// 检查文件是否已经存在
	if _, err := os.Stat(outputPath); err == nil {
		// 文件已存在，直接返回
		return outputPath
	}

	// 从URL下载图像
	src, err := image.NewImageHelp().DownloadImage(qrcodeURL)
	if err != nil {
		fmt.Println("下载图像失败:", err)
		return ""
	}

	// 保存图像
	err = imaging.Save(src, outputPath)
	if err != nil {
		fmt.Println("保存图像失败:", err)
		return ""
	}

	return outputPath
}

// 合并加钟商品数据
func (p *printerTemplate) MergeSaleOrderBuffetDelayProducts(saleOrder *model.SaleOrder) ([]MergeSaleOrderBuffetDelayProducts, float64) {
	num := decimal.NewFromFloat(0)
	delayMap := make(map[string]*MergeSaleOrderBuffetDelayProducts)
	delays := make([]MergeSaleOrderBuffetDelayProducts, 0)
	keyOrder := make([]string, 0)
	for _, delay := range saleOrder.SaleOrderBuffetDelayProducts {
		if delay.IsDelete() {
			continue
		}
		num = num.Add(decimal.NewFromFloat(float64(delay.Num)).Round(3))
		originPrice := delay.GetAmount()
		key := fmt.Sprintf("%s(%v)", delay.Name, originPrice)
		// 按产品名称分组累加
		if _, exists := delayMap[key]; exists {
			// 如果产品名称已存在，则累加数量和总价
			delayMap[key].DelayNum += delay.Num
			delayMap[key].DelayTotalPrice += originPrice
		} else {
			// 如果产品名称不存在，则创建新记录
			delayMap[key] = &MergeSaleOrderBuffetDelayProducts{
				DelayName:       delay.Name,
				DelayPrice:      delay.Price,
				DelayNum:        delay.Num,
				DelayTotalPrice: originPrice,
			}
			// 记录key首次出现的顺序
			keyOrder = append(keyOrder, key)
		}
	}
	// 将map转换为slice
	for _, key := range keyOrder {
		delays = append(delays, *delayMap[key])
	}
	return delays, num.Round(3).InexactFloat64()
}

// ColumnConfig 列配置结构体
type MergeSaleOrderProductOptions struct {
	saleBill             *model.SaleBill
	saleOrder            *model.SaleOrder
	saleOrderProductUuid uint64 // 提取套餐子商品
	IsShowSku            bool   // 是否显示sku
	IsShowWrap           bool   // 是否显示打包
}

// 合并销售订单商品数据
func (p *printerTemplate) MergeSaleOrderProduct(options MergeSaleOrderProductOptions) ([]MergeSaleOrderProduct, float64) {
	saleBill := options.saleBill
	saleOrder := options.saleOrder
	productNum := decimal.NewFromFloat(0)
	productMap := make(map[string]*MergeSaleOrderProduct)
	products := make([]MergeSaleOrderProduct, 0)
	keyOrder := make([]string, 0)
	//
	saleOrderProducts := saleOrder.SaleOrderProducts
	if options.saleOrderProductUuid != 0 {
		saleOrderProducts = saleOrder.GetPackageSubProductList(options.saleOrderProductUuid)
	}
	for _, item := range saleOrderProducts {
		if item.IsDelete() || item.IsUnCookingProduct() || item.IsUnAcceptOrderBool() || item.IsCancelProduct() {
			continue
		}
		if item.IsBuffetProduct() && item.GetTotalSaucePrice() <= 0 {
			continue
		}

		// 提取套餐子商品
		subProducts := make([]MergeSaleOrderProduct, 0)
		md5SubProductsStr := ""
		if options.saleOrderProductUuid == 0 {
			if item.IsPackageSubProduct() {
				continue
			}
			if item.IsPackageProduct() {
				subProducts, _ = p.MergeSaleOrderProduct(MergeSaleOrderProductOptions{
					saleBill:             saleBill,
					saleOrder:            saleOrder,
					saleOrderProductUuid: item.Uuid,
					IsShowSku:            options.IsShowSku,
					IsShowWrap:           options.IsShowWrap,
				})
				//
				for _, subProduct := range subProducts {
					md5SubProductsStr += subProduct.ProductName + item.GetAttributeNamesByLang(p.Lang, options.IsShowSku)
				}
				hexStr := md5.Sum([]byte(md5SubProductsStr))
				md5SubProductsStr = hex.EncodeToString(hexStr[:])
			}
		}

		// 商品数量
		productNum = productNum.Add(decimal.NewFromFloat(item.Num).Round(3))
		// 商品价格
		productPrice := utils.IfFloat64(item.IsBuffetProduct(), item.SaucePrice, item.SalePrice)
		productTotalPrice := utils.IfFloat64(item.IsBuffetProduct(), item.GetTotalSaucePrice(), item.GetSalePrice()) // 商品原价
		// 赠品
		var gift string
		if item.IsGiftProduct() {
			gift = "(" + p.Translate("赠") + ") "
			productTotalPrice = 0
		}
		// 打包商品
		var wrap string
		if options.IsShowWrap {
			if item.IsWrapProduct() || (saleBill.IsTakeout() && saleBill.MemberSaleOrderUuid == 0) {
				wrap = "(" + p.Translate("打包") + ") "
				productTotalPrice = 0
			}
		}

		// 商品名称
		productAttr := item.GetAttributeNamesByLang(p.Lang, options.IsShowSku)
		productName := utils.IfString(item.IsPackageSubProduct(), "--", "") + wrap + gift + item.MultiLanguageName.GetNameByLang(p.Lang)
		if productAttr != "" {
			productName = productName + "\n" + utils.IfString(item.IsPackageSubProduct(), "  ", "") + "(" + productAttr + ")"
		}
		// 按产品名称分组累加
		key := fmt.Sprintf("%s(%v)(%v)(%v)(%s)", productName, productPrice, item.ProductPackageUuid, item.PackageUuid, md5SubProductsStr)
		if _, exists := productMap[key]; exists {
			// 如果产品名称已存在，则累加数量和总价
			productMap[key].ProductNum = decimal.NewFromFloat(productMap[key].ProductNum).Add(decimal.NewFromFloat(item.Num).Round(3)).InexactFloat64()
			productMap[key].ProductTotalPrice = decimal.NewFromFloat(productMap[key].ProductTotalPrice).Add(decimal.NewFromFloat(productTotalPrice).Round(2)).InexactFloat64()
			// 累加套餐子商品
			for i := 0; i < len(subProducts); i++ {
				productMap[key].SubProducts[i].ProductNum = decimal.NewFromFloat(productMap[key].SubProducts[i].ProductNum).Add(decimal.NewFromFloat(subProducts[i].ProductNum).Round(3)).InexactFloat64()
				productMap[key].SubProducts[i].ProductTotalPrice = decimal.NewFromFloat(productMap[key].SubProducts[i].ProductTotalPrice).Add(decimal.NewFromFloat(subProducts[i].ProductTotalPrice).Round(2)).InexactFloat64()
			}
		} else {
			// 如果产品名称不存在，则创建新记录
			productMap[key] = &MergeSaleOrderProduct{
				ProductName:       productName,
				ProductPrice:      productPrice,
				ProductNum:        item.Num,
				ProductTotalPrice: productTotalPrice,
				SubProducts:       subProducts,
			}
			// 记录key首次出现的顺序
			keyOrder = append(keyOrder, key)
		}
	}
	// 按照首次出现的顺序将map转换为slice
	for _, key := range keyOrder {
		products = append(products, *productMap[key])
	}
	return products, productNum.Round(3).InexactFloat64()
}

// 获取收银机SN
func (p *printerTemplate) GetCashierSn(printerSn string) string {
	// 是否存在
	isExist := false
	for _, item := range p.PrinterSetting.CashierPrinter {
		if item.Key == p.Ctx.GetDeviceSn() {
			isExist = true
			if item.Sn != "" {
				return item.Sn
			}
		}
	}
	if !isExist {
		for _, item := range p.PrinterSetting.CashierPrinter {
			if item.Key == printerSn {
				if item.Sn != "" {
					return item.Sn
				}
			}
		}
	}
	//
	return ""
}

// 打印图片商品
func (p *printerTemplate) PrintCompleteOrderImgProducts(
	img *pkg.ImgFont,
	tmpInfo model.PrinterTemplate,
	products printer_model.Products,
	opts ...interface{},
) {
	tmp := tmpInfo.Template
	isShowSku := tmpInfo.IsShowSku
	//
	for _, product := range products {
		// 处理自助餐文本
		buffetText := ""
		if p.PrinterSetting.BuffetSignOpen == "1" {
			if product.IsBuffet {
				buffetText = p.Translate("自助餐") + "-"
			}
		}

		if tmp == 2 || tmp == 3 {
			if p.Lang == "my" {
				img.SetTextLineHeight(75)
			} else {
				img.SetTextLineHeight(60)
			}
			img.SetTextLineHeight(60)
		} else {
			img.SetTextLineHeight(50)
		}
		img.LineFeed(1, 12)

		// 打包商品
		wrapText := ""
		if product.IsWrap {
			wrapText = "(" + p.Translate("打包") + ") "
		}

		// 打印产品名称和数量
		productName := utils.IfString(len(opts) > 0, "--", "") + wrapText + buffetText + product.ProductName.GetLocale(p.Lang)
		totalNum := "x" + p.FloatToString(product.TotalNum)
		productNameWidth := utils.IfInt(len(totalNum) >= 3, 470-(len(totalNum)*7), 480)
		if tmp == 2 || tmp == 3 {
			img.PrintInColumns(
				pkg.ColumnConfig{Text: productName, Width: productNameWidth, Align: pkg.AlignLeft, FontWeight: 2, FontSize: 30},
				pkg.ColumnConfig{Text: totalNum, Width: 0, Align: pkg.AlignRight, FontWeight: 2, FontSize: 30, LineHeight: utils.IfInt(tmp == 3, 90, 42)},
			)
		} else {
			img.PrintInColumns(
				pkg.ColumnConfig{Text: productName, Width: productNameWidth, Align: pkg.AlignLeft, FontWeight: 2, FontSize: 20},
				pkg.ColumnConfig{Text: totalNum, Width: 0, Align: pkg.AlignRight, FontWeight: 2, FontSize: 20, LineHeight: 42},
			)
		}
		if p.Lang == "my" {
			img.LineFeed(1, 12)
		}

		// 分割处理属性
		for _, attr := range utils.IfSlice(isShowSku == 0, product.ProductSauceNamesList, product.ProductAttrList) {
			if attr.GetLocale(p.Lang) == "" {
				continue
			}
			if tmp == 3 {
				img.SetTextLineHeight(utils.IfInt(p.Lang == "my", 110, 100))
				img.SetFontSize(28)
			} else {
				img.SetTextLineHeight(utils.IfInt(p.Lang == "my", 50, 40))
			}
			img.AppendText(utils.IfString(len(opts) > 0, "  ", "") + attr.GetLocale(p.Lang))
			img.LineFeed(1, utils.IfInt(tmp == 3, 100, 40))
			img.SetFontSize(20)
		}

		// 打印备注
		if product.Remark != "" && len(opts) == 0 {
			img.SetTextLineHeight(utils.IfInt(p.IsMyText(product.Remark), 68, 50))
			img.LineFeed(1, 12)
			img.SetFontSize(28)
			img.AppendText(product.Remark)
			img.LineFeed(1, 50)
			img.SetFontSize(20)
		}

		// 换行
		img.LineFeed(1, 12)

		// 打印套餐子商品
		if len(product.SubProducts) > 0 {
			p.PrintCompleteOrderImgProducts(img, tmpInfo, product.SubProducts, 1)
		}

		// 设置文本行高
		img.SetTextLineHeight(50)
	}
}
