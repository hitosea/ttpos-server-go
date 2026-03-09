package service

import (
	"fmt"
	"ttpos-bmp/app/ttpos-erp/api/selling"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service/rpc/erp"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/language"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// ReturnSalesInvoice 退款发票到 ERP（Sales Invoice 模式 - 生成 Credit Note）
// 与 ReturnPosInvoice 功能一致，但使用 Sales Invoice + Credit Note 替代 POS Invoice
func (s *orderSrv) ReturnSalesInvoice(ctx context.Context, saleOrder *model.SaleOrder, returnOrder *model.ReturnOrder, saleBill *model.SaleBill, db *gorm.DB, returnType int, isPartReturn bool) (*selling.ReturnSalesInvoiceResp, error) {
	companySetting := ctx.GetCompanySetting()

	// 构建退款商品列表（与 ReturnPosInvoice 逻辑一致）
	items := make([]*selling.PosInvoiceItem, 0)
	totalTaxFee := decimal.NewFromFloat(0)
	totalServiceFee := decimal.NewFromFloat(0)

	for _, product := range returnOrder.ReturnOrderProducts {
		if product.ProductType == constant.ReturnOrderProductTypeSaleOrderProduct {
			saleOrderProduct, _, _ := saleOrder.GetSaleOrderProduct(product.SaleOrderProductUuid)
			if saleOrderProduct.IsPackageSubProduct() {
				continue
			}
			if isPartReturn && saleOrderProduct.IsGiftProduct() {
				continue
			}
			if saleOrderProduct.IsPackageProduct() {
				subProducts := saleOrder.GetPackageSubProductList(saleOrderProduct.Uuid)
				for _, subProduct := range subProducts {
					packageName := language.JsonToLocaleResponse(saleOrderProduct.Name)
					num := decimal.NewFromFloat(product.Num).Mul(decimal.NewFromFloat(subProduct.GetUnitNum())).Round(3).InexactFloat64()
					items = append(items, &selling.PosInvoiceItem{
						ItemCode:    subProduct.ErpCode,
						Qty:         -num,
						Rate:        0,
						Amount:      0,
						Description: fmt.Sprintf("Sales in package:%s", packageName.EN),
						IsFreeItem:  true,
					})
				}
			}

			taxFee := saleOrderProduct.TaxFee
			{
				tax := decimal.NewFromFloat(taxFee).Mul(decimal.NewFromFloat(product.Num)).Round(2).InexactFloat64()
				serviceTaxFee := decimal.NewFromFloat(saleOrderProduct.ServiceTaxFee).Mul(decimal.NewFromFloat(product.Num)).Round(2).InexactFloat64()
				totalTaxFee = totalTaxFee.Add(decimal.NewFromFloat(tax)).Add(decimal.NewFromFloat(serviceTaxFee))
				serviceFee := decimal.NewFromFloat(saleOrderProduct.ServiceFee).Mul(decimal.NewFromFloat(product.Num))
				totalServiceFee = totalServiceFee.Add(serviceFee)
			}

			if saleOrderProduct.SalePrice == 0 {
				if saleOrderProduct.IsPackageProduct() {
					product.ErpCode = "TC001"
				}
				packageName := language.JsonToLocaleResponse(product.ProductName)
				items = append(items, &selling.PosInvoiceItem{
					ItemCode:    product.ErpCode,
					Qty:         -product.Num,
					Rate:        0,
					Amount:      0,
					IsFreeItem:  true,
					Description: packageName.EN,
				})
			} else {
				packageName := language.JsonToLocaleResponse(product.ProductName)
				item := &selling.PosInvoiceItem{
					ItemCode:    product.ErpCode,
					Qty:         -product.Num,
					Rate:        product.GetProductPriceNoneTax(taxFee, saleOrderProduct.HasTax()),
					Amount:      -product.GetProductTotalAmountNoneTax(taxFee, saleOrderProduct.HasTax()),
					Description: packageName.EN,
				}
				if saleOrderProduct.IsGiftProduct() {
					item.IsFreeItem = true
				}
				if saleOrderProduct.IsPackageProduct() {
					item.ItemCode = "TC001"
				}
				items = append(items, item)
			}
		} else if product.ProductType == constant.ReturnOrderProductTypeSaleOrderBuffetCustomer {
			buffetCustomer, _, _ := saleOrder.GetSaleOrderBuffetCustomerType(product.SaleOrderProductUuid)
			{
				taxFee := buffetCustomer.TaxFee
				tax := decimal.NewFromFloat(taxFee).Mul(decimal.NewFromFloat(product.Num)).Round(2).InexactFloat64()
				serviceTaxFee := decimal.NewFromFloat(buffetCustomer.ServiceTaxFee).Mul(decimal.NewFromFloat(product.Num)).Round(2).InexactFloat64()
				totalTaxFee = totalTaxFee.Add(decimal.NewFromFloat(tax)).Add(decimal.NewFromFloat(serviceTaxFee))
				serviceFee := decimal.NewFromFloat(buffetCustomer.ServiceFee).Mul(decimal.NewFromFloat(product.Num))
				totalServiceFee = totalServiceFee.Add(serviceFee)
			}
			buffetLocaleName := buffetCustomer.GetLocaleBuffetPackageName()
			buffetName := buffetLocaleName.EN
			if buffetCustomer.SalePrice == 0 {
				items = append(items, &selling.PosInvoiceItem{
					ItemCode:    "ZZC001",
					Qty:         -product.Num,
					Rate:        0,
					Amount:      0,
					IsFreeItem:  true,
					Description: fmt.Sprintf("%s-%s", buffetName, buffetCustomer.Name),
				})
			} else {
				items = append(items, &selling.PosInvoiceItem{
					ItemCode:    "ZZC001",
					Qty:         -product.Num,
					Rate:        buffetCustomer.GetFinalSalePriceNoneTax(),
					Amount:      -decimal.NewFromFloat(buffetCustomer.GetFinalSalePriceNoneTax()).Mul(decimal.NewFromFloat(product.Num)).Truncate(3).Round(2).InexactFloat64(),
					Description: fmt.Sprintf("%s-%s", buffetName, buffetCustomer.Name),
				})
			}
		} else if product.ProductType == constant.ReturnOrderProductTypeBuffetAddTimeProduct {
			buffetPackageName := saleBill.GetBuffetPackageName()
			buffetDelayProduct, _, _ := saleOrder.GetSaleOrderBuffetDelayProduct(product.SaleOrderProductUuid)
			if buffetDelayProduct.Price == 0 {
				items = append(items, &selling.PosInvoiceItem{
					ItemCode:    "ZZCJZ001",
					Qty:         -product.Num,
					Rate:        0,
					Amount:      0,
					IsFreeItem:  true,
					Description: fmt.Sprintf("Delay:%s %s", buffetPackageName.EnName, buffetDelayProduct.Name),
				})
			} else {
				items = append(items, &selling.PosInvoiceItem{
					ItemCode:    "ZZCJZ001",
					Qty:         -product.Num,
					Rate:        buffetDelayProduct.Price,
					Amount:      -decimal.NewFromFloat(buffetDelayProduct.Price).Mul(decimal.NewFromFloat(product.Num)).Truncate(3).Round(2).InexactFloat64(),
					Description: fmt.Sprintf("Delay:%s %s", buffetPackageName.EnName, buffetDelayProduct.Name),
				})
			}
		}
	}

	// 构建税费列表
	taxes := make([]*selling.PosInvoiceTax, 0)
	if totalTaxFee.GreaterThan(decimal.NewFromFloat(0)) {
		taxes = append(taxes, &selling.PosInvoiceTax{
			TaxAmount:   -totalTaxFee.InexactFloat64(),
			Description: "Tax",
		})
	}
	if returnType == constant.ReturnOrderRefundTypePart {
		if totalServiceFee.GreaterThan(decimal.NewFromFloat(0)) {
			items = append(items, &selling.PosInvoiceItem{
				ItemCode: constant.PosInvoiceItemCodeServiceFee,
				Qty:      -totalServiceFee.InexactFloat64(),
				Rate:     1,
				Amount:   -totalServiceFee.InexactFloat64(),
			})
		}
	}
	if returnType == constant.ReturnOrderRefundTypeTotal {
		if saleOrder.PaymentCommissionFee > 0 {
			items = append(items, &selling.PosInvoiceItem{
				ItemCode: constant.PosInvoiceItemCodePaymentProcessingFee,
				Qty:      -saleOrder.PaymentCommissionFee,
				Rate:     1,
				Amount:   -saleOrder.PaymentCommissionFee,
			})
		}
		if saleOrder.IsFixedServiceFee() {
			if saleOrder.ServiceFee > 0 {
				items = append(items, &selling.PosInvoiceItem{
					ItemCode: constant.PosInvoiceItemCodeServiceFee,
					Qty:      -saleOrder.ServiceFee,
					Rate:     1,
					Amount:   -saleOrder.ServiceFee,
				})
			}
		} else {
			if totalServiceFee.GreaterThan(decimal.NewFromFloat(0)) {
				items = append(items, &selling.PosInvoiceItem{
					ItemCode: constant.PosInvoiceItemCodeServiceFee,
					Qty:      -totalServiceFee.InexactFloat64(),
					Rate:     1,
					Amount:   -totalServiceFee.InexactFloat64(),
				})
			}
		}

		if saleOrder.ErpDiscountAmount != 0 {
			taxes = append(taxes, &selling.PosInvoiceTax{
				TaxAmount:   saleOrder.ErpDiscountAmount,
				Description: "Whole Order Price Adjustment",
			})
		}
		if saleOrder.ZeroFee != 0 {
			taxes = append(taxes, &selling.PosInvoiceTax{
				TaxAmount:   saleOrder.ZeroFee,
				Description: "Discount Rounding Off",
			})
		}
		if saleOrder.ZeroCheckoutFee != 0 {
			taxes = append(taxes, &selling.PosInvoiceTax{
				TaxAmount:   saleOrder.ZeroCheckoutFee,
				Description: "Checkout Rounding Off",
			})
		}
		if saleOrder.CouponAmount != 0 {
			taxes = append(taxes, &selling.PosInvoiceTax{
				TaxAmount:   saleOrder.CouponAmount,
				Description: "Coupon Deduction",
			})
		}
		if saleOrder.PayPointsAmount != 0 {
			taxes = append(taxes, &selling.PosInvoiceTax{
				TaxAmount:   saleOrder.PayPointsAmount,
				Description: "Points Deduction",
			})
		}
	}

	// 获取所有支付方式
	paymentMethodRepo := repository.NewPaymentMethodRepo(db)
	paymentMethods := paymentMethodRepo.GetPaymentMethodList(paymentMethodRepo.WhereStatus(constant.PaymentMethodStatusEnable))
	methodMap := make(map[int]string)
	for _, paymentMethod := range paymentMethods {
		if paymentMethod.ErpnextPayment != "" {
			methodMap[paymentMethod.Code] = paymentMethod.ErpnextPayment
		}
	}
	payments := make([]*selling.PosInvoicePayment, 0)
	for _, payment := range returnOrder.ReturnOrderAmounts {
		var modeOfPayment string
		if method, ok := methodMap[payment.PaymentMethod.Code]; ok {
			modeOfPayment = method
		} else {
			return nil, errors.WithMessage(errors.New("不支持的支付方式"))
		}
		payments = append(payments, &selling.PosInvoicePayment{
			ModeOfPayment: modeOfPayment,
			Amount:        payment.Amount,
		})
	}

	// 退款类型
	refundType := "full_refund"
	if isPartReturn {
		refundType = "partial_refund"
	}

	erpSrv := erp.NewIErpSrv(s.dbm)
	param := req.ReturnSalesInvoiceReq{
		SiteCode:         companySetting.ErpnextSiteCode,
		OrderNo:          saleOrder.OrderNo,
		SaleOrderUuid:    fmt.Sprintf("%d", saleOrder.Uuid),
		SalesInvoiceName: saleOrder.ErpSalesInvoiceName,
		PostingDatetime:  int64(returnOrder.CreateTime),
		Company:          companySetting.ErpnextCompanyAbbr,
		Customer:         "Default",
		Items:            items,
		Taxes:            taxes,
		Payments:         payments,
		RefundType:       refundType,
	}
	response, err := erpSrv.ReturnSalesInvoice(ctx, param)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return response, nil
}

// SaveSalesInvoice 保存 Sales Invoice 到 ERP（替代 SavePosInvoice）
// 与 SavePosInvoice 相同的商品/税费/支付构建逻辑，但：
// - 不需要 shiftLog/OpenPosEntry
// - update_stock=0（延迟扣库存）
// - 返回 SalesInvoiceName + PaymentEntryNames
func (s *orderSrv) SaveSalesInvoice(ctx context.Context, saleOrder *model.SaleOrder, saleBill *model.SaleBill, db *gorm.DB) (*selling.SaveSalesInvoiceResp, error) {
	companySetting := ctx.GetCompanySetting()

	// 订单商品列表
	items := make([]*selling.PosInvoiceItem, 0)
	isFreeOrder := saleOrder.IsFreeSaleOrder()
	for _, product := range saleOrder.SaleOrderBuffetCustomerTypes {
		buffetLocaleName := product.GetLocaleBuffetPackageName()
		buffetName := buffetLocaleName.EN
		items = append(items, &selling.PosInvoiceItem{
			ItemCode:    "ZZC001",
			Qty:         float64(product.Num),
			Rate:        product.GetFinalSalePriceNoneTax(),
			Amount:      decimal.NewFromFloat(product.GetFinalSalePriceNoneTax()).Mul(decimal.NewFromFloat(float64(product.Num))).Truncate(3).Round(2).InexactFloat64(),
			Description: fmt.Sprintf("%s-%s", buffetName, product.Name),
		})
	}
	buffetPackageName := saleBill.GetBuffetPackageName()
	for _, product := range saleOrder.SaleOrderBuffetDelayProducts {
		items = append(items, &selling.PosInvoiceItem{
			ItemCode:    "ZZCJZ001",
			Qty:         float64(product.Num),
			Rate:        product.Price,
			Amount:      decimal.NewFromFloat(product.Price).Mul(decimal.NewFromFloat(float64(product.Num))).Truncate(3).Round(2).InexactFloat64(),
			Description: fmt.Sprintf("Delay:%s %s", buffetPackageName.EnName, product.Name),
		})
	}
	for _, product := range saleOrder.SaleOrderProducts {
		if product.IsDelete() || product.IsCancelProduct() {
			continue
		}
		if product.IsPackageSubProduct() {
			continue
		}
		if product.IsPackageProduct() {
			subProducts := saleOrder.GetPackageSubProductList(product.Uuid)
			for _, subProduct := range subProducts {
				productBom := subProduct.GetFlavorSaleOrderProductBom()
				erpCode := productBom.ProductBom.ErpCode
				if erpCode == "" {
					erpCode = constant.PosInvoiceItemCodeSpareGoods // BY001 降级
				}
				packageName := language.JsonToLocaleResponse(product.Name)
				items = append(items, &selling.PosInvoiceItem{
					ItemCode:    erpCode,
					Qty:         subProduct.Num,
					Rate:        0,
					Amount:      0,
					Description: fmt.Sprintf("Sales in package:%s", packageName.EN),
					IsFreeItem:  true,
				})
			}
		}
		productBom := product.GetFlavorSaleOrderProductBom()
		erpCode := productBom.ProductBom.ErpCode
		if erpCode == "" {
			erpCode = constant.PosInvoiceItemCodeSpareGoods // BY001 降级
		}
		if product.IsGiftProduct() {
			if product.IsPackageProduct() {
				packageName := language.JsonToLocaleResponse(product.Name)
				items = append(items, &selling.PosInvoiceItem{
					ItemCode:    "TC001",
					Qty:         product.Num,
					Rate:        0,
					Amount:      0,
					Description: packageName.EN,
					IsFreeItem:  true,
				})
			} else {
				items = append(items, &selling.PosInvoiceItem{
					ItemCode:   erpCode,
					Qty:        product.Num,
					Rate:       0,
					Amount:     0,
					IsFreeItem: true,
				})
			}
		} else if product.SalePrice == 0 {
			item := &selling.PosInvoiceItem{
				ItemCode:   erpCode,
				Qty:        product.Num,
				Rate:       0,
				Amount:     0,
				IsFreeItem: true,
			}
			if product.IsPackageProduct() {
				packageName := language.JsonToLocaleResponse(product.Name)
				item.ItemCode = "TC001"
				item.Description = packageName.EN
			}
			items = append(items, item)
		} else {
			if product.IsPackageProduct() {
				packageName := language.JsonToLocaleResponse(product.Name)
				items = append(items, &selling.PosInvoiceItem{
					ItemCode:    "TC001",
					Qty:         product.Num,
					Rate:        product.GetFinalSalePriceNoneTax(),
					Amount:      product.GetProductFinalSalePriceNoneTax(),
					Description: packageName.EN,
					IsFreeItem:  false,
				})
			} else {
				item := &selling.PosInvoiceItem{
					ItemCode:   erpCode,
					Qty:        product.Num,
					Rate:       product.GetFinalSalePriceNoneTax(),
					Amount:     product.GetProductFinalSalePriceNoneTax(),
					IsFreeItem: isFreeOrder,
				}
				if isFreeOrder {
					item.Rate = 0
					item.Amount = 0
				}
				items = append(items, item)
			}
		}
		// 小料
		sauceBoms := product.GetSauceSaleOrderProductBom()
		for _, sauceBom := range sauceBoms {
			erpCode := sauceBom.ProductBom.ProductSauce.ErpCode
			if erpCode == "" {
				erpCode = constant.PosInvoiceItemCodeSpareGoods // BY001 降级
			}
			items = append(items, &selling.PosInvoiceItem{
				ItemCode:   erpCode,
				Qty:        product.Num,
				Rate:       0,
				Amount:     0,
				IsFreeItem: true,
			})
		}
	}
	materialItems := make([]*selling.PosInvoiceItem, 0)
	erpProductBomMaterials := saleOrder.GetErpProductBomMaterials()
	for _, material := range erpProductBomMaterials {
		materialItems = append(materialItems, &selling.PosInvoiceItem{
			ItemCode: material.ErpCode,
			Qty:      material.Num,
			Uom:      material.Uom,
			Rate:     0,
			Amount:   0,
		})
	}

	// 税费
	taxes := make([]*selling.PosInvoiceTax, 0)
	if saleOrder.TaxFee > 0 {
		taxes = append(taxes, &selling.PosInvoiceTax{
			TaxAmount:   saleOrder.TaxFee,
			Description: "Tax",
		})
	}
	if saleOrder.ServiceFee > 0 {
		serviceFeeItem := &selling.PosInvoiceItem{
			ItemCode: constant.PosInvoiceItemCodeServiceFee,
			Qty:      saleOrder.ServiceFee,
			Rate:     1,
			Amount:   saleOrder.ServiceFee,
		}
		if isFreeOrder {
			serviceFeeItem.IsFreeItem = true
		}
		items = append(items, serviceFeeItem)
	}
	if saleOrder.PaymentCommissionFee > 0 {
		items = append(items, &selling.PosInvoiceItem{
			ItemCode: constant.PosInvoiceItemCodePaymentProcessingFee,
			Qty:      saleOrder.PaymentCommissionFee,
			Rate:     1,
			Amount:   saleOrder.PaymentCommissionFee,
		})
	}

	erpDiscountAmount := saleOrder.GetErpCustomAmount()
	if erpDiscountAmount != 0 {
		taxes = append(taxes, &selling.PosInvoiceTax{
			TaxAmount:   -erpDiscountAmount,
			Description: "Whole Order Price Adjustment",
		})
		if err := repository.NewOrderRepo(db).UpdateErpDiscountAmount(saleOrder.Uuid, erpDiscountAmount); err != nil {
			return nil, errors.WithMessage(err)
		}
	}
	if saleOrder.ZeroFee != 0 {
		taxes = append(taxes, &selling.PosInvoiceTax{
			TaxAmount:   -saleOrder.ZeroFee,
			Description: "Discount Rounding Off",
		})
	}
	if saleOrder.ZeroCheckoutFee != 0 {
		taxes = append(taxes, &selling.PosInvoiceTax{
			TaxAmount:   -saleOrder.ZeroCheckoutFee,
			Description: "Checkout Rounding Off",
		})
	}
	if saleOrder.CouponAmount != 0 {
		taxes = append(taxes, &selling.PosInvoiceTax{
			TaxAmount:   -saleOrder.CouponAmount,
			Description: "Coupon Deduction",
		})
	}
	if saleOrder.PayPointsAmount != 0 {
		taxes = append(taxes, &selling.PosInvoiceTax{
			TaxAmount:   -saleOrder.PayPointsAmount,
			Description: "Points Deduction",
		})
	}

	if isFreeOrder {
		taxes = make([]*selling.PosInvoiceTax, 0)
	}

	// 支付方式
	paymentMethodRepo := repository.NewPaymentMethodRepo(db)
	paymentMethods := paymentMethodRepo.GetPaymentMethodList(paymentMethodRepo.WhereStatus(constant.PaymentMethodStatusEnable))
	methodMap := make(map[int]string)
	for _, paymentMethod := range paymentMethods {
		if paymentMethod.ErpnextPayment != "" {
			methodMap[paymentMethod.Code] = paymentMethod.ErpnextPayment
		}
	}

	payments := make([]*selling.PosInvoicePayment, 0)
	if saleOrder.IsFreeSaleOrder() {
		payments = append(payments, &selling.PosInvoicePayment{
			ModeOfPayment: "Free Meal",
			Amount:        0,
		})
	} else if saleOrder.GetAmountValue() == 0 {
		payments = append(payments, &selling.PosInvoicePayment{
			ModeOfPayment: "Cash",
			Amount:        0,
		})
	} else {
		for _, payment := range saleOrder.PaymentOrders {
			if payment.IsDelete() {
				continue
			}
			var modeOfPayment string
			if method, ok := methodMap[payment.PaymentMethod.Code]; ok {
				modeOfPayment = method
			} else {
				return nil, errors.WithMessage(errors.New("不支持的支付方式"))
			}
			payments = append(payments, &selling.PosInvoicePayment{
				ModeOfPayment: modeOfPayment,
				Amount:        payment.Amount,
			})
		}
	}

	// 根据反结账次数生成 OrderNo
	orderNo := saleOrder.OrderNo
	if saleBill.ReverseSettleCount > 0 {
		orderNo = fmt.Sprintf("%s-%d", saleOrder.OrderNo, saleBill.ReverseSettleCount)
	}

	erpSrv := erp.NewIErpSrv(s.dbm)
	param := req.SaveSalesInvoiceReq{
		SiteCode:          companySetting.ErpnextSiteCode,
		OrderNo:           orderNo,
		SaleOrderUuid:     fmt.Sprintf("%d", saleOrder.Uuid),
		SaleBillUuid:      fmt.Sprintf("%d", saleOrder.SaleBillUuid),
		PosProfile:        companySetting.ErpnextPosProfileName,
		Company:           companySetting.ErpnextCompanyAbbr,
		Customer:          "Default",
		Currency:          "THB",
		PriceListCurrency: "THB",
		PostingDatetime:   saleOrder.FinishTime,
		Branch:            companySetting.ErpnextBranchName,
		UpdateStock:       0, // 延迟扣库存
		Items:             items,
		MaterialItems:     materialItems,
		Taxes:             taxes,
		Payments:          payments,
		CompanyUuid:       fmt.Sprintf("%d", ctx.GetCompanyUuid()),
		OrderType:         "sale_order",
	}
	// 反结账后重新结账时，设置 AmendedFrom
	if saleOrder.ErpSalesInvoiceName != "" {
		param.AmendedFrom = saleOrder.ErpSalesInvoiceName
	}

	response, err := erpSrv.SaveSalesInvoice(ctx, param)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	// 异步模式：BMP MQ consumer 会在创建 SI+PE 成功后回写 shop 数据库
	// 此处设置 erp_sync_status=1 表示已入队
	saleOrder.ErpSyncStatus = 1

	return response, nil
}
