package selling

import (
	"context"
	"fmt"
	"time"
	"ttpos-bmp/app/ttpos-erp/api/selling"
	"ttpos-bmp/app/ttpos-erp/internal/consts"
	"ttpos-bmp/app/ttpos-erp/internal/dao"
	"ttpos-bmp/app/ttpos-erp/internal/model/do"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto/erp"
	"ttpos-bmp/app/ttpos-erp/internal/model/entity"
	"ttpos-bmp/app/ttpos-erp/internal/service"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// SaveSalesInvoice 同步创建 Sales Invoice + Payment Entry
func (s *sSelling) SaveSalesInvoice(ctx context.Context, req *selling.SaveSalesInvoiceReq) (*selling.SaveSalesInvoiceResp, error) {
	// 通过公司缩写查询 ERPNext 公司全名
	companyName, err := service.Company().GetCompanyNameWithAbbr(ctx, req.CompanyAbbr)
	if err != nil {
		return nil, gerror.Wrapf(err, "根据公司缩写[%s]查询公司名称失败", req.CompanyAbbr)
	}

	// 解析支付项 payment_id
	if err := s.resolvePaymentIDs(ctx, req.Payments); err != nil {
		return nil, err
	}

	postingDate, postingTime := formatPostingDatetime(req.PostingDatetime)

	// 构建 Sales Invoice 请求体
	si := s.buildSalesInvoice(ctx, req, companyName, postingDate, postingTime)

	// 如果有物品明细，合并到 items（价格已为0）
	isTakeout := req.TakeoutOrderNo != nil && *req.TakeoutOrderNo != ""
	if len(req.MaterialItems) > 0 {
		for _, item := range req.MaterialItems {
			itemCode := item.ItemCode
			description := item.Description
			// Spec 5: 外卖物品同样支持 BY001 降级
			if itemCode == "" && isTakeout {
				itemCode = erp.SpareGoodsItemCode
				description = fmt.Sprintf("[BY001] %s", item.Description)
			}
			si.Items = append(si.Items, erp.SalesInvoiceItem{
				ItemCode:    itemCode,
				Qty:         item.Qty,
				Rate:        0,
				Amount:      0,
				Uom:         item.Uom,
				Description: description,
			})
		}
	}

	// 调用 ERPNext API 创建 Sales Invoice
	resp, err := service.Document().Create(ctx, erp.DocTypeSalesInvoice, si)
	if err != nil {
		return nil, gerror.Wrapf(err, "创建Sales Invoice失败")
	}
	siName := resp.Get("data.name").String()
	if siName == "" {
		return nil, gerror.New("创建Sales Invoice失败，未返回单据名称")
	}

	// 提交 SI（docstatus=1）
	_, err = service.Document().ChangeDocStatus(ctx, erp.DocTypeSalesInvoice, siName, erp.DocstatusSubmitted)
	if err != nil {
		return nil, gerror.Wrapf(err, "提交Sales Invoice[%s]失败", siName)
	}

	// 创建 Payment Entry（每个支付方式一张）
	peNames := make([]string, 0, len(req.Payments))
	for _, payment := range req.Payments {
		if payment.Amount <= 0 {
			continue
		}
		peName, peErr := s.createPaymentEntry(ctx, req, companyName, siName, payment, postingDate)
		if peErr != nil {
			g.Log().Errorf(ctx, "创建Payment Entry失败: mode=%s, amount=%.2f, err=%v",
				payment.ModeOfPayment, payment.Amount, peErr)
			// 记录错误但继续处理其他支付方式
			continue
		}
		peNames = append(peNames, peName)
	}

	return &selling.SaveSalesInvoiceResp{
		SalesInvoiceName:  siName,
		PaymentEntryNames: peNames,
	}, nil
}

// CancelSalesInvoice 取消 Sales Invoice（反结账时调用）
func (s *sSelling) CancelSalesInvoice(ctx context.Context, req *selling.CancelSalesInvoiceReq) error {
	// 先取消关联的 Payment Entry（使用 frappe.client.cancel 标准流程）
	for _, peName := range req.PaymentEntryNames {
		if peName == "" {
			continue
		}
		_, err := service.Rpc().Execute(ctx, &erp.ErpReq{
			Method: erp.ApiSaveCancel,
		}, g.MapStrStr{
			"doctype": erp.DocTypePaymentEntry,
			"name":    peName,
		})
		if err != nil {
			g.Log().Errorf(ctx, "取消Payment Entry[%s]失败: %v", peName, err)
		}
	}

	// 取消 Sales Invoice（使用 frappe.client.cancel 标准流程）
	if req.SalesInvoiceName != "" {
		_, err := service.Rpc().Execute(ctx, &erp.ErpReq{
			Method: erp.ApiSaveCancel,
		}, g.MapStrStr{
			"doctype": erp.DocTypeSalesInvoice,
			"name":    req.SalesInvoiceName,
		})
		if err != nil {
			return gerror.Wrapf(err, "取消Sales Invoice[%s]失败", req.SalesInvoiceName)
		}
	}

	return nil
}

// ReturnSalesInvoice 退款 Credit Note（退款时调用）
func (s *sSelling) ReturnSalesInvoice(ctx context.Context, req *selling.ReturnSalesInvoiceReq) (*selling.ReturnSalesInvoiceResp, error) {
	// 通过公司缩写查询 ERPNext 公司全名
	companyName, err := service.Company().GetCompanyNameWithAbbr(ctx, req.CompanyAbbr)
	if err != nil {
		return nil, gerror.Wrapf(err, "根据公司缩写[%s]查询公司名称失败", req.CompanyAbbr)
	}

	// 解析支付项 payment_id
	if err := s.resolvePaymentIDs(ctx, req.Payments); err != nil {
		return nil, err
	}

	postingDate, _ := formatPostingDatetime(req.PostingDatetime)

	// 构建 Credit Note（is_return=1 的 Sales Invoice）
	cn := s.buildCreditNote(ctx, req, companyName, postingDate)

	// 创建 Credit Note
	resp, err := service.Document().Create(ctx, erp.DocTypeSalesInvoice, cn)
	if err != nil {
		return nil, gerror.Wrapf(err, "创建Credit Note失败")
	}
	cnName := resp.Get("data.name").String()
	if cnName == "" {
		return nil, gerror.New("创建Credit Note失败，未返回单据名称")
	}

	// 提交 Credit Note
	_, err = service.Document().ChangeDocStatus(ctx, erp.DocTypeSalesInvoice, cnName, erp.DocstatusSubmitted)
	if err != nil {
		return nil, gerror.Wrapf(err, "提交Credit Note[%s]失败", cnName)
	}

	// 创建退款 Payment Entry（payment_type=Pay）
	peNames := make([]string, 0, len(req.Payments))
	for _, payment := range req.Payments {
		if payment.Amount <= 0 {
			continue
		}
		peName, peErr := s.createRefundPaymentEntry(ctx, req, companyName, cnName, payment, postingDate)
		if peErr != nil {
			g.Log().Errorf(ctx, "创建退款Payment Entry失败: mode=%s, amount=%.2f, err=%v",
				payment.ModeOfPayment, payment.Amount, peErr)
			continue
		}
		peNames = append(peNames, peName)
	}

	return &selling.ReturnSalesInvoiceResp{
		CreditNoteName:    cnName,
		PaymentEntryNames: peNames,
	}, nil
}

// buildSalesInvoice 构建 Sales Invoice 请求体
// companyName: ERPNext 公司全名（如 "华莱士泰国"），由调用方通过 GetCompanyNameWithAbbr 查询得到
func (s *sSelling) buildSalesInvoice(ctx context.Context, req *selling.SaveSalesInvoiceReq, companyName, postingDate, postingTime string) *erp.SalesInvoice {
	isTakeout := req.TakeoutOrderNo != nil && *req.TakeoutOrderNo != ""

	items := make([]erp.SalesInvoiceItem, 0, len(req.Items))
	for _, item := range req.Items {
		itemCode := item.ItemCode
		description := item.Description
		// Spec 5: 外卖订单商品无法映射 ERP Item 时使用 BY001 备用商品
		if itemCode == "" && isTakeout {
			itemCode = erp.SpareGoodsItemCode
			description = fmt.Sprintf("[BY001] %s", item.Description)
			g.Log().Infof(ctx, "外卖订单使用BY001备用商品: order=%s, orig_desc=%s", req.OrderNo, item.Description)
		}
		siItem := erp.SalesInvoiceItem{
			ItemCode:    itemCode,
			Qty:         item.Qty,
			Rate:        item.Rate,
			Amount:      item.Amount,
			Uom:         item.Uom,
			Description: description,
		}
		if item.IsFreeItem {
			siItem.IsFreeItem = 1
			siItem.DiscountPercentage = 100 // ERPNext SI 不自动处理 is_free_item，需显式打折
		}
		items = append(items, siItem)
	}

	taxes := make([]erp.SalesInvoiceTax, 0, len(req.Taxes))
	for _, tax := range req.Taxes {
		taxes = append(taxes, erp.SalesInvoiceTax{
			ChargeType:  DefaultTaxChargeType,
			AccountHead: DefaultTaxAccountHeadPrefix + req.CompanyAbbr,
			Description: tax.Description,
			Rate:        tax.Rate,
			TaxAmount:   tax.TaxAmount,
		})
	}

	si := &erp.SalesInvoice{
		PosProfile:         req.PosProfile,
		Company:            companyName,
		Customer:           req.Customer,
		IsPos:              0,
		UpdateStock:        int(req.UpdateStock),
		PostingDate:        postingDate,
		PostingTime:        postingTime,
		SetPostingTime:     1,
		Currency:           req.Currency,
		ConversionRate:     1.0,
		PriceListCurrency:  req.PriceListCurrency,
		PlcConversionRate:  1.0,
		SellingPriceList:   consts.DefaultSellingPriceList,
		Branch:             req.Branch,
		Items:              items,
		Taxes:              taxes,
		DiscountAmount:     req.DiscountAmount,
		Docstatus:          0,
		PoNo: req.OrderNo,
	}

	if req.OrderSourceUuid != nil && *req.OrderSourceUuid != "" {
		si.TtposOrderSourceUuid = *req.OrderSourceUuid
	}
	if req.OrderSourceName != nil && *req.OrderSourceName != "" {
		si.TtposOrderSourceName = *req.OrderSourceName
	}
	if req.TakeoutOrderNo != nil {
		si.TtposTakeoutOrderNo = *req.TakeoutOrderNo
	}
	if req.TakeoutProvider != nil {
		si.TtposTakeoutProvider = *req.TakeoutProvider
	}
	if req.AmendedFrom != "" {
		si.AmendedFrom = req.AmendedFrom
	}

	return si
}

// buildCreditNote 构建 Credit Note 请求体
// companyName: ERPNext 公司全名（如 "华莱士泰国"）
func (s *sSelling) buildCreditNote(ctx context.Context, req *selling.ReturnSalesInvoiceReq, companyName, postingDate string) *erp.CreditNote {
	items := make([]erp.SalesInvoiceItem, 0, len(req.Items))
	for _, item := range req.Items {
		siItem := erp.SalesInvoiceItem{
			ItemCode:    item.ItemCode,
			Qty:         item.Qty, // Main 端已传负数
			Rate:        item.Rate,
			Amount:      item.Amount,
			Uom:         item.Uom,
			Description: item.Description,
		}
		if item.IsFreeItem {
			siItem.IsFreeItem = 1
			siItem.DiscountPercentage = 100
		}
		items = append(items, siItem)
	}
	// 物品明细
	for _, item := range req.MaterialItems {
		items = append(items, erp.SalesInvoiceItem{
			ItemCode:    item.ItemCode,
			Qty:         item.Qty,
			Rate:        0,
			Amount:      0,
			Uom:         item.Uom,
			Description: item.Description,
		})
	}

	taxes := make([]erp.SalesInvoiceTax, 0, len(req.Taxes))
	for _, tax := range req.Taxes {
		taxes = append(taxes, erp.SalesInvoiceTax{
			ChargeType:  DefaultTaxChargeType,
			AccountHead: DefaultTaxAccountHeadPrefix + req.CompanyAbbr,
			Description: tax.Description,
			TaxAmount:   tax.TaxAmount, // Main 端已传负数
		})
	}

	return &erp.CreditNote{
		IsReturn:           1,
		ReturnAgainst:      req.SalesInvoiceName,
		Customer:           req.Customer,
		Company:            companyName,
		PostingDate:        postingDate,
		UpdateStock:        0,
		Items:              items,
		Taxes:              taxes,
		TtposRefundType: req.RefundType,
	}
}

// createPaymentEntry 创建收款 Payment Entry
// companyName: ERPNext 公司全名（如 "华莱士泰国"）
func (s *sSelling) createPaymentEntry(ctx context.Context, req *selling.SaveSalesInvoiceReq, companyName, siName string, payment *selling.PosInvoicePayment, postingDate string) (string, error) {
	pe := &erp.PaymentEntry{
		PaymentType:             "Receive",
		PartyType:               "Customer",
		Party:                   req.Customer,
		Company:                 companyName,
		ModeOfPayment:           payment.ModeOfPayment,
		PaidAmount:              payment.Amount,
		ReceivedAmount:          payment.Amount,
		PaidFrom:                "Debtors - " + req.CompanyAbbr,
		PaidTo:                  "Cash - " + req.CompanyAbbr,
		PaidFromAccountCurrency: req.Currency,
		PaidToAccountCurrency:   req.Currency,
		SourceExchangeRate:      1.0,
		TargetExchangeRate:      1.0,
		PostingDate:             postingDate,
		Docstatus:               0,
		References: []erp.PaymentReference{
			{
				ReferenceDoctype: erp.DocTypeSalesInvoice,
				ReferenceName:    siName,
				AllocatedAmount:  payment.Amount,
			},
		},
	}

	// 记录到 receive_payment_entry
	siteCode := service.Rpc().GetSiteCode(ctx)
	peBody, _ := gjson.Encode(pe)
	dao.ReceivePaymentEntry.Ctx(ctx).InsertAndGetId(&entity.ReceivePaymentEntry{
		SaleOrderUuid: req.SaleOrderUuid,
		ModeOfPayment: payment.ModeOfPayment,
		Docstatus:     erp.DocstatusDraft,
		PaidAmount:    payment.Amount,
		SiteCode:      siteCode,
		ReqBody:       string(peBody),
		CreatedAt:     int(gtime.Now().Timestamp()),
	})

	resp, err := service.Document().Create(ctx, erp.DocTypePaymentEntry, pe)
	if err != nil {
		return "", gerror.Wrapf(err, "创建Payment Entry失败")
	}
	peName := resp.Get("data.name").String()
	if peName == "" {
		return "", gerror.New("创建Payment Entry失败，未返回单据名称")
	}

	// 提交 PE
	_, err = service.Document().ChangeDocStatus(ctx, erp.DocTypePaymentEntry, peName, erp.DocstatusSubmitted)
	if err != nil {
		return peName, gerror.Wrapf(err, "提交Payment Entry[%s]失败", peName)
	}

	// 更新 receive_payment_entry 记录
	dao.ReceivePaymentEntry.Ctx(ctx).Where(do.ReceivePaymentEntry{
		SaleOrderUuid: req.SaleOrderUuid,
		ModeOfPayment: payment.ModeOfPayment,
	}).Data(do.ReceivePaymentEntry{
		PaymentEntryName: peName,
		Docstatus:        erp.DocstatusSubmitted,
		RespBody:         fmt.Sprintf("PE created: %s", peName),
		UpdatedAt:        int(gtime.Now().Timestamp()),
	}).Update()

	return peName, nil
}

// createRefundPaymentEntry 创建退款 Payment Entry（payment_type=Pay）
// companyName: ERPNext 公司全名（如 "华莱士泰国"）
func (s *sSelling) createRefundPaymentEntry(ctx context.Context, req *selling.ReturnSalesInvoiceReq, companyName, cnName string, payment *selling.PosInvoicePayment, postingDate string) (string, error) {
	pe := &erp.PaymentEntry{
		PaymentType:             "Pay",
		PartyType:               "Customer",
		Party:                   req.Customer,
		Company:                 companyName,
		ModeOfPayment:           payment.ModeOfPayment,
		PaidAmount:              payment.Amount,
		ReceivedAmount:          payment.Amount,
		PaidFrom:                "Cash - " + req.CompanyAbbr,
		PaidTo:                  "Debtors - " + req.CompanyAbbr,
		PaidFromAccountCurrency: DefaultAccountCurrency,
		PaidToAccountCurrency:   DefaultAccountCurrency,
		SourceExchangeRate:      1.0,
		TargetExchangeRate:      1.0,
		PostingDate:             postingDate,
		Docstatus:               0,
		References: []erp.PaymentReference{
			{
				ReferenceDoctype: erp.DocTypeSalesInvoice,
				ReferenceName:    cnName,
				AllocatedAmount:  -payment.Amount, // Credit Note outstanding 为负数，allocated 需匹配
			},
		},
	}

	resp, err := service.Document().Create(ctx, erp.DocTypePaymentEntry, pe)
	if err != nil {
		return "", gerror.Wrapf(err, "创建退款Payment Entry失败")
	}
	peName := resp.Get("data.name").String()
	if peName == "" {
		return "", gerror.New("创建退款Payment Entry失败，未返回单据名称")
	}

	_, err = service.Document().ChangeDocStatus(ctx, erp.DocTypePaymentEntry, peName, erp.DocstatusSubmitted)
	if err != nil {
		return peName, gerror.Wrapf(err, "提交退款Payment Entry[%s]失败", peName)
	}

	return peName, nil
}

// formatPostingDatetime 将时间戳格式化为 ERPNext 日期和时间
func formatPostingDatetime(timestamp int64) (date, timeStr string) {
	if timestamp <= 0 {
		timestamp = time.Now().Unix()
	}
	t := time.Unix(timestamp, 0)
	return t.Format("2006-01-02"), t.Format("15:04:05")
}
