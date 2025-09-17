// Package template 提供打印模板相关功能
package template

import (
	"fmt"
	settingResp "ttpos-server-go/app/dto/resp/setting"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/printer/pkg"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"go.uber.org/zap"
)

// StatementOrderData 订单数据结构体
type StatementOrderData struct {
	BrandName string                 `json:"brand_name"`
	Store     StatementStoreData     `json:"store"`
	Order     StatementOrderInfoData `json:"order"`
}

// StatementStoreData 店铺数据结构体
type StatementStoreData struct {
	Name             string `json:"name"`
	Address          string `json:"address"`
	Phone            string `json:"phone"`
	Logo             string `json:"logo"`
	Company          string `json:"company"`
	CompanyAddr      string `json:"company_addr"`
	CompanyPhone     string `json:"company_phone"`
	CompanyTaxNumber string `json:"company_tax_number"`
	CashierSn        string `json:"cashier_sn"`
	PrinterSn        string `json:"printer_sn"`
}

// StatementOrderInfoData 订单信息数据结构体
type StatementOrderInfoData struct {
	Status                 uint                      `json:"status"`
	SerialNo               string                    `json:"serial_no"`
	OrderNo                string                    `json:"order_no"`
	Remark                 string                    `json:"remark"`
	CashierName            string                    `json:"cashier_name"`
	Time                   string                    `json:"time"`
	Buffets                []StatementBuffetData     `json:"buffets"`
	Delays                 []StatementDelayData      `json:"delays"`
	Products               []StatementProductData    `json:"products"`
	ProductNum             float64                   `json:"product_num"`
	ProductAmount          float64                   `json:"product_amount"`
	ServiceFee             float64                   `json:"service_fee"`
	TaxRate                float64                   `json:"tax_rate"`
	TaxFee                 float64                   `json:"tax_fee"`
	IsContainTax           uint                      `json:"is_contain_tax"`
	DiscountFee            float64                   `json:"discount_fee"`
	DiscountRate           float64                   `json:"discount_rate"`
	MemberDiscountFee      float64                   `json:"member_discount_fee"`
	MemberDiscountRate     float64                   `json:"member_discount_rate"`
	MemberCardDiscountRate float64                   `json:"member_card_discount_rate"`
	MemberPointsDiscount   float64                   `json:"member_points_discount"`
	CouponExchangeAmount   float64                   `json:"coupon_exchange_amount"`
	CheckOutZeroFee        float64                   `json:"check_out_zero_fee"`
	ReturnAmount           float64                   `json:"return_amount"`
	PaymentCommissionFee   float64                   `json:"payment_commission_fee"`
	FreeAmount             float64                   `json:"free_amount"`
	ActualReceivePrice     float64                   `json:"actual_receive_price"`
	PaymentMethods         []StatementPaymentMethod  `json:"payment_methods"`
	PercentageLists        []StatementPercentageData `json:"percentage_lists"`
	IsFree                 uint                      `json:"is_free"`
	MemberRemainingBalance float64                   `json:"member_remaining_balance"`
	MemberPoints           float64                   `json:"member_points"`
	PaymentName            string                    `json:"payment_name"`
	PaymentQrcode          string                    `json:"payment_qrcode"`
}

// StatementBuffetData 自助餐数据结构体
type StatementBuffetData struct {
	Name     string  `json:"name"`
	PriceNum string  `json:"price_num"`
	Price    float64 `json:"price"`
	Num      int     `json:"num"`
	Subtotal float64 `json:"subtotal"`
	Info     string  `json:"info"`
}

// StatementDelayData 加钟数据结构体
type StatementDelayData struct {
	Name     string  `json:"name"`
	PriceNum string  `json:"price_num"`
	Price    float64 `json:"price"`
	Num      int     `json:"num"`
	Subtotal float64 `json:"subtotal"`
}

// StatementProductData 商品数据结构体
type StatementProductData struct {
	Name         string  `json:"name"`
	PriceNum     string  `json:"price_num"`
	Price        float64 `json:"price"`
	Num          int     `json:"num"`
	Subtotal     float64 `json:"subtotal"`
	Info         string  `json:"info"`
	IsGift       int     `json:"is_gift"`
	IsPackage    int     `json:"is_package"`
	IsSubProduct int     `json:"is_sub_product"`
}

// StatementPaymentMethod 支付方式数据结构体
type StatementPaymentMethod struct {
	Name string `json:"name"`
	Text string `json:"text"`
}

// StatementPercentageData 百分比数据结构体
type StatementPercentageData struct {
	Name string `json:"name"`
	Text string `json:"text"`
}

// statementOrderImgTemplateCustom 图片订单打印模板
type statementOrderImgTemplateCustom struct {
	base *printerTemplate
}

// NewStatementOrderImgTemplate 创建新的图片订单打印模板
func NewStatementOrderImgTemplateCustom(
	base *printerTemplate,
) *statementOrderImgTemplateCustom {
	return &statementOrderImgTemplateCustom{
		base: base,
	}
}

// ImgPrint 图片打印
func (t *statementOrderImgTemplateCustom) GetPrintContent(
	settingPrinterInfo settingResp.PrinterInfo,
	tmpData string,
	saleBill *model.SaleBill,
	saleOrder *model.SaleOrder,
	payMethodUuid uint64,
) string { // 就餐人数

	mealNumStr := ""
	if saleBill.MealNum > 0 {
		mealNumStr = fmt.Sprintf(" (%d%s)", saleBill.MealNum, t.base.Translate("人"))
	}

	// 订单名称
	orderName := saleOrder.GetOrderName()

	// 构建订单数据结构体
	statementData := &StatementOrderData{
		BrandName: config.Server.BrandName,
		Store: StatementStoreData{
			Name:             t.base.StoreSetting.Name,
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
		Order: StatementOrderInfoData{
			Status: saleOrder.Status,
			SerialNo: func() string {
				if saleBill.DeskUuid > 0 {
					return fmt.Sprintf("%s: %s%s%s", t.base.Translate("桌号"), saleBill.SerialNo, orderName, mealNumStr)
				} else {
					return fmt.Sprintf("%s: %s%s", t.base.Translate("取单号"), saleBill.SerialNo, orderName)
				}
			}(),
			OrderNo:       saleOrder.OrderNo,
			Remark:        saleBill.Remark,
			CashierName:   saleOrder.CashierName,
			Time:          t.base.FormatUnixTimeDefault(saleOrder.FinishTime),
			Buffets:       []StatementBuffetData{},  // 需要根据实际业务逻辑填充
			Delays:        []StatementDelayData{},   // 需要根据实际业务逻辑填充
			Products:      []StatementProductData{}, // 需要根据实际业务逻辑填充
			ProductNum:    float64(saleOrder.GetProductNum()),
			ProductAmount: saleOrder.ProductAmount,
			ServiceFee:    saleOrder.ServiceFee,
			// TaxRate:                saleOrder.TaxRate,
			// TaxFee:                 saleOrder.TaxFee,
			// IsContainTax:           saleOrder.IsContainTax,
			// DiscountFee:            saleOrder.DiscountFee,
			// DiscountRate:           saleOrder.DiscountRate,
			// MemberDiscountFee:      saleOrder.MemberDiscountFee,
			// MemberDiscountRate:     saleOrder.MemberDiscountRate,
			// MemberCardDiscountRate: saleOrder.MemberCardDiscountRate,
			// MemberPointsDiscount:   saleOrder.MemberPointsDiscount,
			// CouponExchangeAmount:   saleOrder.CouponExchangeAmount,
			// CheckOutZeroFee:        saleOrder.CheckOutZeroFee,
			// ReturnAmount:           saleOrder.ReturnAmount,
			// PaymentCommissionFee:   saleOrder.PaymentCommissionFee,
			// FreeAmount:             saleOrder.FreeAmount,
			// ActualReceivePrice:     saleOrder.ActualReceivePrice,
			// PaymentMethods:         []StatementPaymentMethod{},  // 需要根据实际业务逻辑填充
			// PercentageLists:        []StatementPercentageData{}, // 需要根据实际业务逻辑填充
			// IsFree:                 saleOrder.IsFree,
			// MemberRemainingBalance: saleOrder.MemberRemainingBalance,
			// MemberPoints:           saleOrder.MemberPoints,
			// PaymentName:            saleOrder.PaymentName,
			// PaymentQrcode:          saleOrder.PaymentQrcode,
		},
	}

	// 将结构体转换为map
	dataMap, _ := utils.StrToMap(utils.ToJsonString(statementData))

	// 创建解析器
	parser, err := pkg.NewImgTemplateParser(pkg.ImgBaseData{
		Language:             t.base.Lang,
		CurrencyUnit:         t.base.CurrencyUnit,
		CurrencyUnitPosition: t.base.CurrencyUnitPosition,
	}, tmpData, dataMap)
	if err != nil {
		logger.Logger.Error("复杂模板创建解析器失败", zap.Error(err))
		return ""
	}

	// 验证模板
	err = parser.ValidateTemplate()
	if err != nil {
		logger.Logger.Error("复杂模板验证失败", zap.Error(err))
		return ""
	}

	// 解析模板
	img, err := parser.Parse()
	if err != nil {
		logger.Logger.Error("复杂模板解析失败", zap.Error(err))
		return ""
	}

	//
	return img.Save("", false, 0)
}
