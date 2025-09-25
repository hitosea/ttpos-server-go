// Package template 提供打印模板相关功能
package template

import (
	"fmt"
	"strconv"
	"ttpos-server-go/app/constant"
	settingResp "ttpos-server-go/app/dto/resp/setting"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/printer/pkg"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// StatementOrderData 订单数据结构体
type StatementOrderData struct {
	BrandName string                 `json:"brand_name"`
	Store     StatementStoreData     `json:"store"`
	Order     StatementOrderInfoData `json:"order"`
}

// StatementStoreData 店铺数据结构体
type StatementStoreData struct {
	Name             string `json:"name"`               // 店铺名称
	Address          string `json:"address"`            // 店铺地址
	Phone            string `json:"phone"`              // 店铺电话
	Logo             string `json:"logo"`               // 店铺logo
	Company          string `json:"company"`            // 公司名称
	CompanyAddr      string `json:"company_addr"`       // 公司地址
	CompanyPhone     string `json:"company_phone"`      // 公司电话
	CompanyTaxNumber string `json:"company_tax_number"` // 公司税号
	CashierSn        string `json:"cashier_sn"`         // 收银员编号
	PrinterSn        string `json:"printer_sn"`         // 打印机编号
}

// StatementOrderInfoData 订单信息数据结构体
type StatementOrderInfoData struct {
	Status                 uint                      `json:"status"`                    // 订单状态
	SerialNo               string                    `json:"serial_no"`                 // 桌位编号
	OrderNo                string                    `json:"order_no"`                  // 订单号
	Remark                 string                    `json:"remark"`                    // 订单备注
	CashierName            string                    `json:"cashier_name"`              // 收银员名称
	FinishTime             string                    `json:"finish_time"`               // 完成时间
	CreateTime             string                    `json:"create_time"`               // 创建时间
	PayTime                string                    `json:"pay_time"`                  // 支付时间
	Buffets                []StatementBuffetData     `json:"buffets"`                   // 自助餐列表
	Delays                 []StatementDelayData      `json:"delays"`                    // 加钟列表
	Products               []StatementProductData    `json:"products"`                  // 商品列表
	ProductNum             float64                   `json:"product_num"`               // 商品数量
	ProductAmount          string                    `json:"product_amount"`            // 商品金额
	ServiceFee             string                    `json:"service_fee"`               // 服务费
	TaxRate                float64                   `json:"tax_rate"`                  // 税费率
	TaxFeeType             uint                      `json:"tax_fee_type"`              // 税费类型	0-关闭消费税 1-商品未含税 2-商品已含税
	IsContainTax           uint                      `json:"is_contain_tax"`            // 是否包含税
	DiscountFee            string                    `json:"discount_fee"`              // 折扣金额
	DiscountRate           string                    `json:"discount_rate"`             // 折扣率
	MemberDiscountFee      string                    `json:"member_discount_fee"`       // 会员折扣金额
	MemberDiscountRate     float64                   `json:"member_discount_rate"`      // 会员折扣率
	MemberCardDiscountRate float64                   `json:"member_card_discount_rate"` // 会员卡折扣率
	MemberPointsDiscount   float64                   `json:"member_points_discount"`    // 会员积分抵扣金额
	CouponExchangeAmount   float64                   `json:"coupon_exchange_amount"`    // 优惠券抵扣金额
	CheckOutZeroFee        string                    `json:"check_out_zero_fee"`        // 结账抹零金额
	ReturnAmount           string                    `json:"return_amount"`             // 退款金额
	PaymentCommissionFee   float64                   `json:"payment_commission_fee"`    // 支付手续费
	FreeAmount             string                    `json:"free_amount"`               // 免单金额
	ActualReceivePrice     string                    `json:"actual_receive_price"`      // 实际收款金额
	PaymentMethods         []StatementPaymentMethod  `json:"payment_methods"`           // 支付方式列表
	PercentageLists        []StatementPercentageData `json:"percentage_lists"`          // 百分比列表
	IsFree                 bool                      `json:"is_free"`                   // 是否免单
	IsMember               bool                      `json:"is_member"`                 // 是否会员
	MemberRemainingBalance string                    `json:"member_remaining_balance"`  // 会员剩余余额
	MemberPoints           float64                   `json:"member_points"`             // 会员积分
	PaymentName            string                    `json:"payment_name"`              // 支付方式名称
	PaymentQrcode          string                    `json:"payment_qrcode"`            // 支付方式二维码
	Barcode                string                    `json:"barcode"`                   // 条形码
}

// StatementBuffetData 自助餐数据结构体
type StatementBuffetData struct {
	Name     string `json:"name"`
	PriceNum string `json:"price_num"`
	Price    string `json:"price"`
	Num      uint   `json:"num"`
	Subtotal string `json:"subtotal"`
	Attrs    string `json:"attrs"`
	Attr     string `json:"attr"`
}

// StatementDelayData 加钟数据结构体
type StatementDelayData struct {
	Name     string `json:"name"`
	PriceNum string `json:"price_num"`
	Price    string `json:"price"`
	Num      uint   `json:"num"`
	Attrs    string `json:"attrs"`
	Attr     string `json:"Attr"`
	Subtotal string `json:"subtotal"`
}

// StatementProductData 商品数据结构体
type StatementProductData struct {
	Name         string  `json:"name"`
	PriceNum     string  `json:"price_num"`
	Price        string  `json:"price"`
	Num          float64 `json:"num"`
	Subtotal     string  `json:"subtotal"`
	Attrs        string  `json:"attrs"`
	Attr         string  `json:"attr"`
	SauceNames   string  `json:"sauce_names"`
	IsDelay      bool    `json:"is_delay"`
	IsBuffet     bool    `json:"is_buffet"`
	IsGift       bool    `json:"is_gift"`
	IsPackage    bool    `json:"is_package"`
	IsSubProduct bool    `json:"is_sub_product"`
}

// StatementPaymentMethod 支付方式数据结构体
type StatementPaymentMethod struct {
	Name string `json:"name"`
	Text string `json:"text"`
}

// StatementPercentageData 百分比数据结构体
type StatementPercentageData struct {
	TaxRate    string `json:"tax_rate"`
	TaxFee     string `json:"tax_fee"`
	TotalPrice string `json:"total_price"`
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
) string {
	// 订单名称
	orderName := saleOrder.GetOrderName()
	// 就餐人数
	mealNumStr := utils.IfString(saleBill.MealNum > 0, fmt.Sprintf(" (%d%s)", saleBill.MealNum, t.base.Translate("人")), "")
	// 商品数量
	productNum := decimal.NewFromFloat(0)

	// 支付方式
	var paymentMethodName string
	var qrCodeUrl string
	if payMethodUuid != 0 {
		db := t.base.Ctx.GetDB()
		paymentMethodRepo := repository.NewPaymentMethodRepo(db)
		paymentMethod := paymentMethodRepo.GetPaymentMethod(
			paymentMethodRepo.WhereUuid(payMethodUuid),
			paymentMethodRepo.WithQrcodeFile(),
		)
		if paymentMethod.Uuid != 0 {
			paymentMethodName = paymentMethod.Name
			if paymentMethod.IsLianLianPay() {
				llPaymentOrder, err := repository.NewLlPaymentOrderRepo(db).GetPaymentOrder(
					repository.CommonRepo.WhereBySoftDelete(),
					func(db *gorm.DB) *gorm.DB {
						db = db.Where("related_uuid = ?", saleOrder.Uuid)
						db = db.Where("order_type = ?", constant.PaymentOrderRelatedTypeSaleOrder)
						db = db.Where("payment_method_uuid = ?", paymentMethod.Uuid)
						return db.Order("id desc")
					},
				)
				if err == nil && llPaymentOrder.Uuid > 0 {
					qrCodeUrl = llPaymentOrder.LinkUrl
				} else {
					qrCodeUrl = t.base.Translate("获取二维码错误") + ":" + qrCodeUrl
				}
			} else {
				qrCodeUrl = paymentMethod.QrcodeFile.GetUrl(utils.GetBaseURL(t.base.Ctx.GetGin().Request))
				if url := t.base.GetQrcodeAddr(qrCodeUrl); url != "" {
					qrCodeUrl = url
				} else {
					qrCodeUrl = t.base.Translate("获取二维码错误") + ":" + qrCodeUrl
				}
			}
		}
	}

	// 商品列表
	products := []StatementProductData{}

	// 自助餐顾客类型
	for _, orderBuffetCustomer := range saleOrder.SaleOrderBuffetCustomerTypes {
		if orderBuffetCustomer.IsDelete() {
			continue
		}
		productNum = productNum.Add(decimal.NewFromFloat(float64(orderBuffetCustomer.Num)).Round(3))
		originPrice := orderBuffetCustomer.GetOriginPrice()
		products = append(products, StatementProductData{
			Name:         orderBuffetCustomer.BuffetPackage.MultiLanguageName.GetNameByLang(t.base.Lang),
			PriceNum:     fmt.Sprintf("%s*%d", t.base.Amount(orderBuffetCustomer.SalePrice), orderBuffetCustomer.Num),
			Price:        t.base.Amount(originPrice),
			Num:          float64(orderBuffetCustomer.Num),
			Subtotal:     t.base.Amount(orderBuffetCustomer.TotalPrice),
			Attrs:        orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name,
			Attr:         orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name,
			SauceNames:   "",
			IsBuffet:     true,
			IsGift:       false,
			IsPackage:    false,
			IsSubProduct: false,
		})
	}

	// 添加加钟商品
	buffetDelayProducts, num := t.base.MergeSaleOrderBuffetDelayProducts(saleOrder)
	productNum = productNum.Add(decimal.NewFromFloat(num).Round(3))
	for _, delay := range buffetDelayProducts {
		products = append(products, StatementProductData{
			Name:         delay.DelayName,
			PriceNum:     fmt.Sprintf("%s*%d", t.base.Amount(delay.DelayPrice), delay.DelayNum),
			Price:        t.base.Amount(delay.DelayPrice),
			Num:          float64(delay.DelayNum),
			Subtotal:     t.base.Amount(delay.DelayTotalPrice),
			Attrs:        delay.DelayName,
			Attr:         delay.DelayName,
			SauceNames:   "",
			IsDelay:      true,
			IsBuffet:     false,
			IsGift:       false,
			IsPackage:    false,
			IsSubProduct: false,
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
		products = append(products, StatementProductData{
			Name:         product.Name,
			PriceNum:     fmt.Sprintf("%s*%v", t.base.Amount(product.ProductPrice), product.ProductNum),
			Price:        t.base.Amount(product.ProductPrice),
			Num:          product.ProductNum,
			Subtotal:     t.base.Amount(product.ProductTotalPrice),
			Attrs:        product.Attrs,
			Attr:         product.Attr,
			SauceNames:   product.SauceNames,
			IsGift:       product.IsGift,
			IsPackage:    product.IsWrap,
			IsSubProduct: product.IsSub,
		})
		// 套餐子商品
		for _, subProduct := range product.SubProducts {
			products = append(products, StatementProductData{
				Name:         subProduct.Name,
				PriceNum:     fmt.Sprintf("%v", subProduct.ProductNum),
				Price:        t.base.Amount(subProduct.ProductPrice),
				Num:          subProduct.ProductNum,
				Subtotal:     t.base.Amount(subProduct.ProductTotalPrice),
				Attrs:        subProduct.Attrs,
				Attr:         subProduct.Attr,
				SauceNames:   subProduct.SauceNames,
				IsGift:       false,
				IsPackage:    false,
				IsSubProduct: true,
			})
		}
	}

	// 支付方式
	paymentMethods := []StatementPaymentMethod{}
	if saleOrder.IsFreeSaleOrder() {
		paymentMethods = append(paymentMethods, StatementPaymentMethod{
			Name: t.base.Translate("支付方式"),
			Text: t.base.Translate("免单"),
		})
		paymentMethods = append(paymentMethods, StatementPaymentMethod{
			Name: t.base.Translate("实收金额"),
			Text: t.base.Amount(0),
		})
	}
	if len(saleOrder.PaymentOrders) > 0 {
		for _, paymentOrder := range saleOrder.PaymentOrders {
			paymentMethods = append(paymentMethods, StatementPaymentMethod{
				Name: t.base.Translate("支付方式"),
				Text: paymentOrder.PaymentMethod.GetName(),
			})
			paymentMethods = append(paymentMethods, StatementPaymentMethod{
				Name: t.base.Translate("实收金额"),
				Text: t.base.Amount(paymentOrder.Amount),
			})
			if saleOrder.ChangeAmount > 0 && paymentOrder.PaymentMethod.Code == constant.PaymentMethodCodeCash {
				paymentMethods = append(paymentMethods, StatementPaymentMethod{
					Name: t.base.Translate("找零"),
					Text: t.base.Amount(saleOrder.ChangeAmount),
				})
			}
		}
	}

	// 百分比列表
	percentageLists := []StatementPercentageData{}
	for _, percentage := range saleOrder.GetPercentageList() {
		taxRate := percentage["TaxRate"]
		taxFee, _ := strconv.ParseFloat(percentage["TaxFee"], 64)
		totalPrice, _ := strconv.ParseFloat(percentage["TotalPrice"], 64)
		percentageLists = append(percentageLists, StatementPercentageData{
			TaxRate:    taxRate,
			TaxFee:     t.base.Amount(taxFee),
			TotalPrice: t.base.Amount(totalPrice),
		})
	}

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
			OrderNo:     saleOrder.OrderNo,
			Remark:      saleBill.Remark,
			CashierName: saleOrder.CashierName,
			FinishTime:  t.base.FormatUnixTimeDefault(saleOrder.FinishTime),
			CreateTime:  t.base.FormatUnixTimeDefault(saleOrder.CreateTime),
			PayTime:     t.base.FormatUnixTimeDefault(saleOrder.FinishTime),
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
			PaymentName:   paymentMethodName,
			PaymentQrcode: qrCodeUrl,
			//
			Barcode: saleOrder.OrderNo,
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
