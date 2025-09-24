// Package template 提供打印模板相关功能
package template

import (
	"fmt"
	"os"
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
	FinishTime             int64                     `json:"finish_time"`               // 完成时间
	CreateTime             int64                     `json:"create_time"`               // 创建时间
	PayTime                string                    `json:"pay_time"`                  // 支付时间
	Buffets                []StatementBuffetData     `json:"buffets"`                   // 自助餐列表
	Delays                 []StatementDelayData      `json:"delays"`                    // 加钟列表
	Products               []StatementProductData    `json:"products"`                  // 商品列表
	ProductNum             float64                   `json:"product_num"`               // 商品数量
	ProductAmount          float64                   `json:"product_amount"`            // 商品金额
	ServiceFee             float64                   `json:"service_fee"`               // 服务费
	TaxRate                float64                   `json:"tax_rate"`                  // 税费率
	TaxFeeType             uint                      `json:"tax_fee_type"`              // 税费类型	0-关闭消费税 1-商品未含税 2-商品已含税
	IsContainTax           uint                      `json:"is_contain_tax"`            // 是否包含税
	DiscountFee            float64                   `json:"discount_fee"`              // 折扣金额
	DiscountRate           float64                   `json:"discount_rate"`             // 折扣率
	MemberDiscountFee      float64                   `json:"member_discount_fee"`       // 会员折扣金额
	MemberDiscountRate     float64                   `json:"member_discount_rate"`      // 会员折扣率
	MemberCardDiscountRate float64                   `json:"member_card_discount_rate"` // 会员卡折扣率
	MemberPointsDiscount   float64                   `json:"member_points_discount"`    // 会员积分抵扣金额
	CouponExchangeAmount   float64                   `json:"coupon_exchange_amount"`    // 优惠券抵扣金额
	CheckOutZeroFee        float64                   `json:"check_out_zero_fee"`        // 结账抹零金额
	ReturnAmount           float64                   `json:"return_amount"`             // 退款金额
	PaymentCommissionFee   float64                   `json:"payment_commission_fee"`    // 支付手续费
	FreeAmount             float64                   `json:"free_amount"`               // 免单金额
	ActualReceivePrice     float64                   `json:"actual_receive_price"`      // 实际收款金额
	PaymentMethods         []StatementPaymentMethod  `json:"payment_methods"`           // 支付方式列表
	PercentageLists        []StatementPercentageData `json:"percentage_lists"`          // 百分比列表
	IsFree                 bool                      `json:"is_free"`                   // 是否免单
	MemberRemainingBalance float64                   `json:"member_remaining_balance"`  // 会员剩余余额
	MemberPoints           float64                   `json:"member_points"`             // 会员积分
	PaymentName            string                    `json:"payment_name"`              // 支付方式名称
	PaymentQrcode          string                    `json:"payment_qrcode"`            // 支付方式二维码
}

// StatementBuffetData 自助餐数据结构体
type StatementBuffetData struct {
	Name     string  `json:"name"`
	PriceNum string  `json:"price_num"`
	Price    float64 `json:"price"`
	Num      uint    `json:"num"`
	Subtotal float64 `json:"subtotal"`
	Info     string  `json:"info"`
}

// StatementDelayData 加钟数据结构体
type StatementDelayData struct {
	Name     string  `json:"name"`
	PriceNum string  `json:"price_num"`
	Price    float64 `json:"price"`
	Num      uint    `json:"num"`
	Info     string  `json:"info"`
	Subtotal float64 `json:"subtotal"`
}

// StatementProductData 商品数据结构体
type StatementProductData struct {
	Name         string  `json:"name"`
	PriceNum     string  `json:"price_num"`
	Price        float64 `json:"price"`
	Num          float64 `json:"num"`
	Subtotal     float64 `json:"subtotal"`
	Info         string  `json:"info"`
	IsGift       uint    `json:"is_gift"`
	IsPackage    uint    `json:"is_package"`
	IsSubProduct uint    `json:"is_sub_product"`
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
	// 订单名称
	orderName := saleOrder.GetOrderName()
	// 就餐人数
	mealNumStr := utils.IfString(saleBill.MealNum > 0, fmt.Sprintf(" (%d%s)", saleBill.MealNum, t.base.Translate("人")), "")

	//
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

	// 商品数量
	productNum := decimal.NewFromFloat(0)

	// 自助餐顾客类型
	buffets := []StatementBuffetData{}
	for _, orderBuffetCustomer := range saleOrder.SaleOrderBuffetCustomerTypes {
		if orderBuffetCustomer.IsDelete() {
			continue
		}
		productNum = productNum.Add(decimal.NewFromFloat(float64(orderBuffetCustomer.Num)).Round(3))
		originPrice := orderBuffetCustomer.GetOriginPrice()
		buffets = append(buffets, StatementBuffetData{
			Name:     orderBuffetCustomer.BuffetPackage.MultiLanguageName.GetNameByLang(t.base.Lang),
			PriceNum: fmt.Sprintf("%s*%d", t.base.Amount(orderBuffetCustomer.SalePrice), orderBuffetCustomer.Num),
			Price:    originPrice,
			Num:      orderBuffetCustomer.Num,
			Subtotal: orderBuffetCustomer.TotalPrice,
			Info:     orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name,
		})
	}

	// 添加加钟商品
	delays := []StatementDelayData{}
	buffetDelayProducts, num := t.base.MergeSaleOrderBuffetDelayProducts(saleOrder)
	productNum = productNum.Add(decimal.NewFromFloat(num).Round(3))
	for _, delay := range buffetDelayProducts {
		delays = append(delays, StatementDelayData{
			Name:     delay.DelayName,
			PriceNum: fmt.Sprintf("%s*%d", t.base.Amount(delay.DelayPrice), delay.DelayNum),
			Price:    delay.DelayPrice,
			Num:      delay.DelayNum,
			Subtotal: delay.DelayTotalPrice,
			Info:     delay.DelayName,
		})
	}

	// 商品列表
	products := []StatementProductData{}
	mergeProducts, num := t.base.MergeSaleOrderProduct(MergeSaleOrderProductOptions{
		saleBill:   saleBill,
		saleOrder:  saleOrder,
		IsShowSku:  false,
		IsShowWrap: true,
	})
	productNum = productNum.Add(decimal.NewFromFloat(num).Round(3))
	for _, product := range mergeProducts {
		products = append(products, StatementProductData{
			Name:         product.ProductName,
			PriceNum:     fmt.Sprintf("%s*%v", t.base.Amount(product.ProductPrice), product.ProductNum),
			Price:        product.ProductPrice,
			Num:          product.ProductNum,
			Subtotal:     product.ProductTotalPrice,
			Info:         product.ProductName,
			IsGift:       0,
			IsPackage:    0,
			IsSubProduct: 0,
		})
		// 套餐子商品
		for _, subProduct := range product.SubProducts {
			products = append(products, StatementProductData{
				Name:         subProduct.ProductName,
				PriceNum:     fmt.Sprintf("%v", subProduct.ProductNum),
				Price:        subProduct.ProductPrice,
				Num:          subProduct.ProductNum,
				Subtotal:     subProduct.ProductTotalPrice,
				Info:         subProduct.ProductName,
				IsGift:       0,
				IsPackage:    0,
				IsSubProduct: 0,
			})
		}
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
			OrderNo:       saleOrder.OrderNo,
			Remark:        saleBill.Remark,
			CashierName:   saleOrder.CashierName,
			FinishTime:    saleOrder.FinishTime,
			CreateTime:    saleOrder.CreateTime,
			PayTime:       t.base.FormatUnixTimeDefault(saleOrder.FinishTime),
			Buffets:       buffets,  // 需要根据实际业务逻辑填充
			Delays:        delays,   // 需要根据实际业务逻辑填充
			Products:      products, // 需要根据实际业务逻辑填充
			ProductNum:    productNum.InexactFloat64(),
			ProductAmount: saleOrder.ProductAmount,
			ServiceFee:    saleOrder.ServiceFee,
			TaxFeeType:    saleBill.SaleBillSetting.TaxFeeType,
			IsContainTax: func() uint {
				if saleOrder.TaxFee > 0 && saleBill.SaleBillSetting.TaxFeeType == 2 && (t.base.ConsumptionTax == 1 || t.base.ConsumptionTax == 2) {
					return 1
				}
				if saleOrder.TaxFee > 0 && saleBill.SaleBillSetting.TaxFeeType == 1 && (t.base.ConsumptionTax == 1 || t.base.ConsumptionTax == 3) {
					return 2
				}
				return 0
			}(),
			DiscountFee:            saleOrder.CustomDiscountFee,
			DiscountRate:           saleOrder.CustomDiscountRate,
			MemberDiscountFee:      saleOrder.MemberDiscountFee,
			MemberDiscountRate:     saleOrder.MemberDiscountRate,
			MemberCardDiscountRate: saleOrder.MemberCardDiscountRate,
			MemberPointsDiscount: func() float64 {
				if saleOrder.PayPointsAmount > 0 && saleOrder.PayPoints > 0 {
					return saleOrder.PayPointsAmount
				}
				return 0
			}(),
			CouponExchangeAmount: saleOrder.CalcCouponExchangeAmount(),
			CheckOutZeroFee:      saleOrder.GetCheckOutZeroFee(),
			ReturnAmount:         saleOrder.GetReturnAmount(),
			PaymentCommissionFee: saleOrder.PaymentCommissionFee,
			FreeAmount: func() float64 {
				if saleOrder.IsFreeSaleOrder() {
					return saleOrder.GetAmount()
				}
				return 0
			}(),
			ActualReceivePrice: saleOrder.GetPrintReceivablePrice(),
			PaymentMethods:     []StatementPaymentMethod{},  // 需要根据实际业务逻辑填充
			PercentageLists:    []StatementPercentageData{}, // 需要根据实际业务逻辑填充
			IsFree:             saleOrder.IsFreeSaleOrder(),
			MemberRemainingBalance: func() float64 {
				if saleOrder.Member == nil {
					return 0
				}
				return saleOrder.GetMemberSurplusBalance()
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
			PaymentName:   paymentMethodName,
			PaymentQrcode: qrCodeUrl,
		},
	}

	// 将结构体转换为map
	dataMap, _ := utils.StrToMap(utils.ToJsonString(statementData))

	// 创建复杂的测试模板
	templateJSON, err := os.ReadFile("../template_json/statement_order_tmp.json")
	if err != nil {
		logger.Logger.Error("读取 tmp.json 文件失败", zap.Error(err))
		return ""
	}

	// 将 templateJSON 转换为字符串
	tmpData = string(templateJSON)

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

	// TODO: 临时测试
	img.DebugSetSegmentationHeight(22200)

	//
	return img.Save("", !t.base.IsSunMi && settingPrinterInfo.IsEnableSound(), 0)
}
