package service

import (
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/context"
)

// OrderProductQuotation 订单商品报价 - 只计算价格，不写数据库
func (s *orderSrv) OrderProductQuotation(ctx context.Context, request req.OrderProductQuotationReq) (*resp.OrderProductQuotationResp, error) {
	db := ctx.GetDB()

	// 1. 加载完整的 SaleBill 对象树（含 SaleOrders、SaleOrderProducts 等）
	orderRepo := repository.NewOrderRepo(db)
	shopCart, err := orderRepo.GetOrderCartInfo(request.SaleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err, "加载销售账单失败")
	}
	saleBill := shopCart.SaleBill

	// 2. 获取目标 SaleOrder
	saleOrder := saleBill.GetSaleOrder(request.SaleOrderUuid)
	if saleOrder == nil {
		return nil, errors.New("销售订单不存在")
	}

	// 3. 构建新商品对象（跳过所有业务校验：库存、上下架、限购等）
	newProducts, err := s.newSaleOrderProduct(ctx, CreateSaleOrderProductParams{
		Setting:   *saleBill.SaleBillSetting,
		SaleBill:  saleBill,
		SaleOrder: saleOrder,
		Products:  request.Products,
	}, WithSkipValidation())
	if err != nil {
		return nil, errors.WithMessage(err, "构建商品失败")
	}

	// 4. 重新计算所有层级金额（商品级 → 订单级 → 账单级）
	// newSaleOrderProduct 已经把新商品 append 到 saleOrder.SaleOrderProducts
	// CalcAll 会重算整个 SaleBill 的所有金额
	saleBill.CalcAll()

	// 5. 构建响应
	result := buildQuotationResponse(newProducts, saleOrder, saleBill)

	return result, nil
}

// buildQuotationResponse 从计算后的模型对象构建报价响应
func buildQuotationResponse(newProducts []*model.SaleOrderProduct, saleOrder *model.SaleOrder, saleBill *model.SaleBill) *resp.OrderProductQuotationResp {
	// 商品明细（只返回主商品，套餐子商品不单独返回）
	products := make([]resp.QuotationProduct, 0, len(newProducts))
	for _, p := range newProducts {
		if p.IsPackageSubProduct() {
			continue
		}
		qp := resp.QuotationProduct{
			ProductPackageUuid: p.ProductPackageUuid,
			LocaleName:         p.GetLocaleName(),
			Num:                p.Num,
			FlavorPrice:        p.FlavorPrice,
			SaucePrice:         p.SaucePrice,
			AddPrice:           p.AddPrice,
			ProductPrice:       p.ProductPrice,
			SalePrice:          p.SalePrice,
			SalePriceNoTax:     p.SalePriceNoTax,
			Price:              p.Price,
			MemberDiscountFee:  p.MemberDiscountFee,
			CustomDiscountFee:  p.CustomDiscountFee,
			DiscountFee:        p.DiscountFee,
			TaxFee:             p.TaxFee,
			ServiceFee:         p.ServiceFee,
			ServiceTaxFee:      p.ServiceTaxFee,
			TotalPrice:         p.TotalPrice,
			OriginTotalPrice:   p.OriginTotalPrice,
		}
		products = append(products, qp)
	}

	// 订单级金额
	orderCalc := resp.QuotationOrderCalc{
		ProductOriginalAmount: saleOrder.ProductOriginalAmount,
		ProductAmount:         saleOrder.ProductAmount,
		ServiceFee:            saleOrder.ServiceFee,
		TaxFee:                saleOrder.TaxFee,
		CustomDiscountFee:     saleOrder.CustomDiscountFee,
		MemberDiscountFee:     saleOrder.MemberDiscountFee,
		Amount:                saleOrder.Amount,
		OriginAmount:          saleOrder.OriginAmount,
	}

	// 账单级金额
	billCalc := resp.QuotationBillCalc{
		Amount:                saleBill.Amount,
		OriginAmount:          saleBill.OriginAmount,
		ProductAmount:         saleBill.ProductAmount,
		ProductOriginalAmount: saleBill.ProductOriginalAmount,
		ServiceFee:            saleBill.ServiceFee,
		TaxFee:                saleBill.TaxFee,
		DiscountFee:           saleBill.CustomDiscountFee,
		MemberDiscountFee:     saleBill.MemberDiscountFee,
	}

	return &resp.OrderProductQuotationResp{
		Products:  products,
		OrderCalc: orderCalc,
		BillCalc:  billCalc,
	}
}
