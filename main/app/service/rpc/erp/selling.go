package erp

import (
	"context"
	"regexp"
	"slices"
	"strings"
	"ttpos-bmp/app/ttpos-erp/api/selling"
	"ttpos-server-go/app/cloud"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/model"
	pkgCtx "ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/utils"

	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/repository"
	cc "ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/logger"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func NewErpSellingClient() (selling.SellingServiceClient, *grpc.ClientConn, error) {
	conn, err := cloud.GetRpcConnWithName(cloud.ErpServiceName)
	if err != nil {
		return nil, nil, err
	}
	return selling.NewSellingServiceClient(conn), conn, nil
}

func (s *erpSrv) GetPosProfileList(ctx context.Context, getPosProfileListReq req.GetPosProfileListReq) (resp.GetPosProfileListResp, error) {
	getPosProfileListResp := resp.GetPosProfileListResp{
		List: make([]resp.PosProfileInfo, 0),
	}
	client, conn, err := NewErpSellingClient()
	if err != nil {
		return getPosProfileListResp, err
	}
	defer conn.Close()
	req := &selling.PosProfileReq{
		Name:        getPosProfileListReq.PosProfileName,
		Company:     getPosProfileListReq.Company,
		CompanyAbbr: getPosProfileListReq.CompanyAbbr,
	}
	result, err := client.GetPosProfileList(WithSiteCode(ctx, getPosProfileListReq.SiteCode), req)
	if err != nil {
		return getPosProfileListResp, err
	}
	if result.GetCode() != "0" || result.Data == nil {
		logger.Logger.Error("GetPosProfileList", zap.String("result_message", result.GetMessage()), zap.String("code", result.GetCode()))
		return getPosProfileListResp, errors.New(result.GetMessage())
	}
	var posProfileListResp selling.PosProfileListResp
	if err := result.Data.UnmarshalTo(&posProfileListResp); err != nil {
		logger.Logger.Error("GetPosProfileList-UnmarshalTo", zap.Any("err", err))
		return getPosProfileListResp, err
	}
	for _, profile := range posProfileListResp.ProfileList {
		getPosProfileListResp.List = append(getPosProfileListResp.List, resp.PosProfileInfo{
			Name:      profile.Name,
			Company:   profile.Company,
			Branch:    profile.Branch,
			Warehouse: profile.Warehouse,
		})
	}
	return getPosProfileListResp, nil
}

// AddLianPayment 添加连连支付
func (s *erpSrv) AddLianPayment(ctx cc.Context, req req.ErpnextSiteAddLianLianPaymentReq) error {
	client, conn, err := NewErpSellingClient()
	if err != nil {
		return err
	}
	defer conn.Close()

	company, err := repository.NewCompanyRepo(s.dbm.GetDB(constant.DefaultDB)).GetCompanyInfoByUuid(req.CompanyUuid)
	if err != nil {
		logger.Logger.Error("AddLianPayment-获取当前公司数据失败", zap.Error(err))
		return err
	}

	// 创建连连支付账号
	paymentTypes := []string{"Wechat Pay", "Alipay", "QR PromptPay"}
	for _, paymentType := range paymentTypes {
		createPaymentAccountReq := &selling.CreatePaymentAccountReq{
			CompanyAbbr:   company.CompanySetting.ErpnextCompanyAbbr,
			PaymentType:   paymentType,
			Branch:        company.CompanySetting.ErpnextBranchName,
			PaymentSource: "2",
		}
		_, err = client.CreatePaymentAccount(WithSiteCode(context.Background(), company.CompanySetting.ErpnextSiteCode), createPaymentAccountReq)
		if err != nil {
			logger.Logger.Error("AddLianPayment-创建支付账号失败", zap.Error(err))
			return err
		}
	}

	return nil
}

// OpenPosEntry 开账
func (s *erpSrv) OpenPosEntry(ctx context.Context, openEntryReq req.OpenPosEntryReq) (string, error) {
	client, conn, err := NewErpSellingClient()
	if err != nil {
		return "", err
	}
	defer conn.Close()

	openPosEntryDetail := make([]*selling.OpenPosEntryDetail, 0)
	if len(openEntryReq.OpenPosEntryDetail) != 0 {
		for _, detail := range openEntryReq.OpenPosEntryDetail {
			openPosEntryDetailErp := &selling.OpenPosEntryDetail{
				OpeningAmount: detail.OpeningAmount,
			}
			if detail.PaymentId != nil && *detail.PaymentId != "" {
				openPosEntryDetailErp.PaymentId = detail.PaymentId
			} else if detail.ModeOfPayment != "" {
				openPosEntryDetailErp.ModeOfPayment = &detail.ModeOfPayment
			}

			openPosEntryDetail = append(openPosEntryDetail, openPosEntryDetailErp)
		}
	}
	openPosEntryReq := &selling.OpenPosEntryReq{
		PosProfileName:     openEntryReq.PosProfileName,
		CashierEmail:       openEntryReq.CashierEmail,
		CompanyAbbr:        openEntryReq.CompanyAbbr,
		PeriodStartDate:    openEntryReq.PeriodStartDate,
		OpenPosEntryDetail: openPosEntryDetail,
		Branch:             openEntryReq.Branch,
	}
	result, err := client.OpenPosEntry(WithSiteCode(ctx, openEntryReq.SiteCode), openPosEntryReq)
	if err != nil {
		return "", err
	}
	if result.GetCode() != "0" || result.Data == nil {
		logger.Logger.Error("OpenPosEntry", zap.String("result_message", result.GetMessage()), zap.String("code", result.GetCode()))
		return "", errors.New(result.GetMessage())
	}
	var openPosEntryResp selling.OpenPosEntryResp
	if err := result.Data.UnmarshalTo(&openPosEntryResp); err != nil {
		logger.Logger.Error("OpenPosEntry-UnmarshalTo", zap.Any("err", err))
		return "", err
	}
	if openPosEntryResp.OpenPosEntryInfo == nil {
		return "", errors.New(result.GetMessage())
	}
	return openPosEntryResp.OpenPosEntryInfo.OpenPosEntryName, nil
}

// ClosePosEntry 关账
func (s *erpSrv) ClosePosEntry(ctx context.Context, closeEntryReq req.ClosePosEntryReq) (string, error) {
	client, conn, err := NewErpSellingClient()
	if err != nil {
		return "", err
	}
	defer conn.Close()

	closePosEntryDetail := make([]*selling.ClosePosEntryDetail, 0)
	if len(closeEntryReq.ClosePosEntryDetail) != 0 {
		for _, detail := range closeEntryReq.ClosePosEntryDetail {
			entryDetail := &selling.ClosePosEntryDetail{
				OpeningAmount: detail.OpeningAmount,
				ClosingAmount: detail.ClosingAmount,
			}
			if detail.PaymentId != nil && *detail.PaymentId != "" {
				entryDetail.PaymentId = detail.PaymentId
			} else if detail.ModeOfPayment != "" {
				entryDetail.ModeOfPayment = &detail.ModeOfPayment
			}
			closePosEntryDetail = append(closePosEntryDetail, entryDetail)
		}
	}
	closePosEntryReq := &selling.ClosePosEntryReq{
		PosOpenEntryName:    closeEntryReq.PosOpenEntryName,
		PeriodEndDate:       closeEntryReq.PeriodEndDate,
		ClosePosEntryDetail: closePosEntryDetail,
		InvoiceCount:        closeEntryReq.InvoiceCount,
	}
	result, err := client.ClosePosEntry(WithSiteCode(ctx, closeEntryReq.SiteCode), closePosEntryReq)
	if err != nil {
		return "", err
	}
	if result.GetCode() != "0" || result.Data == nil {
		resultMessage := result.GetMessage()
		logger.Logger.Error("ClosePosEntry", zap.String("result_message", resultMessage), zap.String("code", result.GetCode()))
		return "", errors.New(resultMessage)
	}
	var closePosEntryResp selling.ClosePosEntryResp
	if err := result.Data.UnmarshalTo(&closePosEntryResp); err != nil {
		logger.Logger.Error("ClosePosEntry-UnmarshalTo", zap.Any("err", err))
		return "", err
	}
	return closePosEntryResp.AsyncRecordId, nil
}

// SavePosInvoice 保存 Pos Invoice
func (s *erpSrv) SavePosInvoice(ctx pkgCtx.Context, savePosInvoiceReq req.SavePosInvoiceReq) (*selling.SavePosInvoiceResp, error) {
	companySetting := ctx.GetCompanySetting()

	client, conn, err := NewErpSellingClient()
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	defer conn.Close()

	params := &selling.SavePosInvoiceReq{
		OrderNo:           savePosInvoiceReq.OrderNo,
		OpenPosEntryName:  savePosInvoiceReq.OpenPosEntryName,
		CompanyAbbr:       companySetting.ErpnextCompanyAbbr,
		PostingDatetime:   savePosInvoiceReq.PostingDatetime,
		UpdateStock:       1,
		Currency:          "THB",
		PriceListCurrency: "THB",
		Branch:            companySetting.ErpnextBranchName,
		CustomerUuid:      savePosInvoiceReq.CustomerUuid,
		Items:             savePosInvoiceReq.Items,
		MaterialItems:     savePosInvoiceReq.MaterialItems,
		Taxes:             savePosInvoiceReq.Taxes,
		Payments:          savePosInvoiceReq.Payments,
		// 反结账后，重新结账时填写原商品发票名称
		AmendedProductsInvoiceName: savePosInvoiceReq.AmendedProductsInvoiceName,
		AmendedMaterialInvoiceName: savePosInvoiceReq.AmendedMaterialInvoiceName,
		Remark:                     savePosInvoiceReq.Remark,
		TakeoutOrderNo:             savePosInvoiceReq.TakeoutOrderNo,
		TakeoutProvider:            savePosInvoiceReq.TakeoutProvider,
	}
	res, err := client.SavePosInvoice(WithSiteCode(ctx.GetContext(), savePosInvoiceReq.SiteCode), params)
	if err != nil {
		ctx.Log().Info("SavePosInvoice", zap.String("msg", err.Error()))
		msg := utils.ShortenErpnextError(err.Error())
		return nil, errors.WithMessage(errors.New(msg))
	}
	if res.GetCode() != "0" {
		ctx.Log().Info("SavePosInvoice", zap.String("msg", res.Message), zap.String("code", res.GetCode()))
		msg := utils.ShortenErpnextError(res.Message)
		return nil, errors.WithMessage(errors.New(msg))
	}
	if res.Data != nil {
		var savePosInvoiceResp selling.SavePosInvoiceResp
		if err := res.Data.UnmarshalTo(&savePosInvoiceResp); err != nil {
			logger.Logger.Error("SavePosInvoice-UnmarshalTo", zap.Any("err", err))
			return nil, errors.WithMessage(err)
		}
		return &savePosInvoiceResp, nil
	}
	return nil, errors.WithMessage(errors.New("保存POS发票异常, data为空"))
}

// 解析SavePosInvoice接口返回到错误信息
func (s *erpSrv) ParseSavePosInvoiceError(err error) (*resp.GetPosInvoiceErrorResp, error) {
	// 解析 “调用erpnext接口返回异常,error:Row #1: Item Code: <strong>WPR3685375438618625</strong> is not available under warehouse <strong>散户商家-Normal-Default - xs1</strong>.”
	// 提取出Item Code
	if strings.Contains(err.Error(), "Item Code: <strong>") && strings.Contains(err.Error(), "is not available under warehouse") {
		itemCode := ""
		re := regexp.MustCompile(`Item Code: <strong>(WPR\d+)</strong>`)
		matches := re.FindStringSubmatch(err.Error())
		if len(matches) > 1 {
			itemCode = matches[1]
		}
		return &resp.GetPosInvoiceErrorResp{ErrorScene: constant.ErpItemStockNotEnough, ItemCode: itemCode}, nil
	}

	// 解析“Stock quantity not enough for Item Code: <strong>WPR3685375438618625</strong> under warehouse <strong>散户商家-Normal-Default - xs1</strong>.”
	if strings.Contains(err.Error(), "Stock quantity not enough for Item Code") {
		itemCode := ""
		re := regexp.MustCompile(`Item Code: <strong>(WPR\d+)</strong>`)
		matches := re.FindStringSubmatch(err.Error())
		if len(matches) > 1 {
			itemCode = matches[1]
		}
		return &resp.GetPosInvoiceErrorResp{ErrorScene: constant.ErpItemStockNotEnough, ItemCode: itemCode}, nil
	}

	return nil, errors.WithMessage(err)
}

func (s *erpSrv) CancelPosInvoice(ctx pkgCtx.Context, cancelPosInvoiceReq req.CancelPosInvoiceReq) error {
	companySetting := ctx.GetCompanySetting()

	client, conn, err := NewErpSellingClient()
	if err != nil {
		return errors.WithMessage(err)
	}
	defer conn.Close()

	params := &selling.CancelPosInvoiceReq{
		ProductsInvoiceName: cancelPosInvoiceReq.ProductsInvoiceName,
		MaterialInvoiceName: cancelPosInvoiceReq.MaterialInvoiceName,
		OpenPosEntryName:    cancelPosInvoiceReq.OpenPosEntryName,
		OrderNo:             cancelPosInvoiceReq.OrderNo,
		Remark:              cancelPosInvoiceReq.Remark,
	}
	res, err := client.CancelPosInvoice(WithSiteCode(ctx.GetContext(), companySetting.ErpnextSiteCode), params)
	if err != nil {
		return errors.WithMessage(err)
	}
	if res.GetCode() != "0" {
		return errors.WithMessage(errors.New(res.Message))
	}
	return nil
}

func (s *erpSrv) ReturnPosInvoice(ctx pkgCtx.Context, returnPosInvoiceReq req.ReturnPosInvoiceReq) (*selling.ReturnPosInvoiceResp, error) {
	companySetting := ctx.GetCompanySetting()

	client, conn, err := NewErpSellingClient()
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	defer conn.Close()

	params := &selling.ReturnPosInvoiceReq{
		OrderNo:          returnPosInvoiceReq.OrderNo,
		OpenPosEntryName: returnPosInvoiceReq.OpenPosEntryName,
		PostingDatetime:  returnPosInvoiceReq.PostingDatetime,
		CompanyAbbr:      companySetting.ErpnextCompanyAbbr,
		InvoiceName:      returnPosInvoiceReq.InvoiceName,
		Items:            returnPosInvoiceReq.Items,
		Taxes:            returnPosInvoiceReq.Taxes,
		Payments:         returnPosInvoiceReq.Payments,
		InvoiceType:      1,
	}
	res, err := client.ReturnPosInvoice(WithSiteCode(ctx.GetContext(), returnPosInvoiceReq.SiteCode), params)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	if res.GetCode() != "0" {
		return nil, errors.WithMessage(errors.New(res.Message))
	}
	if res.Data != nil {
		var returnPosInvoiceResp selling.ReturnPosInvoiceResp
		if err := res.Data.UnmarshalTo(&returnPosInvoiceResp); err != nil {
			logger.Logger.Error("ReturnPosInvoice-UnmarshalTo", zap.Any("err", err))
			return nil, errors.WithMessage(err)
		}
		return &returnPosInvoiceResp, nil
	}
	return nil, errors.WithMessage(errors.New("退款POS发票异常, data为空"))
}

// SaveSalesInvoice 保存 Sales Invoice（替代 SavePosInvoice）
func (s *erpSrv) SaveSalesInvoice(ctx pkgCtx.Context, saveSalesInvoiceReq req.SaveSalesInvoiceReq) (*selling.SaveSalesInvoiceResp, error) {
	client, conn, err := NewErpSellingClient()
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	defer conn.Close()

	params := &selling.SaveSalesInvoiceReq{
		OrderNo:           saveSalesInvoiceReq.OrderNo,
		SaleOrderUuid:     saveSalesInvoiceReq.SaleOrderUuid,
		SaleBillUuid:      saveSalesInvoiceReq.SaleBillUuid,
		PosProfile:        saveSalesInvoiceReq.PosProfile,
		Company:           saveSalesInvoiceReq.Company,
		Customer:          saveSalesInvoiceReq.Customer,
		Currency:          saveSalesInvoiceReq.Currency,
		PriceListCurrency: saveSalesInvoiceReq.PriceListCurrency,
		PostingDatetime:   saveSalesInvoiceReq.PostingDatetime,
		Branch:            saveSalesInvoiceReq.Branch,
		UpdateStock:       saveSalesInvoiceReq.UpdateStock,
		Items:             saveSalesInvoiceReq.Items,
		MaterialItems:     saveSalesInvoiceReq.MaterialItems,
		Taxes:             saveSalesInvoiceReq.Taxes,
		Payments:          saveSalesInvoiceReq.Payments,
		DiscountAmount:    saveSalesInvoiceReq.DiscountAmount,
		AmendedFrom:       saveSalesInvoiceReq.AmendedFrom,
		Remark:            saveSalesInvoiceReq.Remark,
		TakeoutOrderNo:    saveSalesInvoiceReq.TakeoutOrderNo,
		TakeoutProvider:   saveSalesInvoiceReq.TakeoutProvider,
		CompanyUuid:       saveSalesInvoiceReq.CompanyUuid,
	}
	res, err := client.SaveSalesInvoice(WithSiteCode(ctx.GetContext(), saveSalesInvoiceReq.SiteCode), params)
	if err != nil {
		ctx.Log().Info("SaveSalesInvoice", zap.String("msg", err.Error()))
		msg := utils.ShortenErpnextError(err.Error())
		return nil, errors.WithMessage(errors.New(msg))
	}
	if res.GetCode() != "0" {
		ctx.Log().Info("SaveSalesInvoice", zap.String("msg", res.Message), zap.String("code", res.GetCode()))
		msg := utils.ShortenErpnextError(res.Message)
		return nil, errors.WithMessage(errors.New(msg))
	}
	if res.Data != nil {
		var saveSalesInvoiceResp selling.SaveSalesInvoiceResp
		if err := res.Data.UnmarshalTo(&saveSalesInvoiceResp); err != nil {
			logger.Logger.Error("SaveSalesInvoice-UnmarshalTo", zap.Any("err", err))
			return nil, errors.WithMessage(err)
		}
		return &saveSalesInvoiceResp, nil
	}
	return nil, errors.WithMessage(errors.New("保存Sales Invoice异常, data为空"))
}

// CancelSalesInvoice 取消 Sales Invoice
func (s *erpSrv) CancelSalesInvoice(ctx pkgCtx.Context, cancelSalesInvoiceReq req.CancelSalesInvoiceReq) (*selling.CancelSalesInvoiceResp, error) {
	client, conn, err := NewErpSellingClient()
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	defer conn.Close()

	params := &selling.CancelSalesInvoiceReq{
		OrderNo:           cancelSalesInvoiceReq.OrderNo,
		SaleOrderUuid:     cancelSalesInvoiceReq.SaleOrderUuid,
		SalesInvoiceName:  cancelSalesInvoiceReq.SalesInvoiceName,
		PaymentEntryNames: cancelSalesInvoiceReq.PaymentEntryNames,
		StockDeducted:     cancelSalesInvoiceReq.StockDeducted,
		Remark:            cancelSalesInvoiceReq.Remark,
	}
	res, err := client.CancelSalesInvoice(WithSiteCode(ctx.GetContext(), cancelSalesInvoiceReq.SiteCode), params)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	if res.GetCode() != "0" {
		return nil, errors.WithMessage(errors.New(res.Message))
	}
	if res.Data != nil {
		var cancelSalesInvoiceResp selling.CancelSalesInvoiceResp
		if err := res.Data.UnmarshalTo(&cancelSalesInvoiceResp); err != nil {
			logger.Logger.Error("CancelSalesInvoice-UnmarshalTo", zap.Any("err", err))
			return nil, errors.WithMessage(err)
		}
		return &cancelSalesInvoiceResp, nil
	}
	return nil, errors.WithMessage(errors.New("取消Sales Invoice异常, data为空"))
}

// ReturnSalesInvoice 退款 Sales Invoice (Credit Note)
func (s *erpSrv) ReturnSalesInvoice(ctx pkgCtx.Context, returnSalesInvoiceReq req.ReturnSalesInvoiceReq) (*selling.ReturnSalesInvoiceResp, error) {
	client, conn, err := NewErpSellingClient()
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	defer conn.Close()

	params := &selling.ReturnSalesInvoiceReq{
		OrderNo:          returnSalesInvoiceReq.OrderNo,
		SaleOrderUuid:    returnSalesInvoiceReq.SaleOrderUuid,
		SalesInvoiceName: returnSalesInvoiceReq.SalesInvoiceName,
		PostingDatetime:  returnSalesInvoiceReq.PostingDatetime,
		Company:          returnSalesInvoiceReq.Company,
		Customer:         returnSalesInvoiceReq.Customer,
		Items:            returnSalesInvoiceReq.Items,
		MaterialItems:    returnSalesInvoiceReq.MaterialItems,
		Taxes:            returnSalesInvoiceReq.Taxes,
		Payments:         returnSalesInvoiceReq.Payments,
		RefundType:       returnSalesInvoiceReq.RefundType,
		Remark:           returnSalesInvoiceReq.Remark,
	}
	res, err := client.ReturnSalesInvoice(WithSiteCode(ctx.GetContext(), returnSalesInvoiceReq.SiteCode), params)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	if res.GetCode() != "0" {
		return nil, errors.WithMessage(errors.New(res.Message))
	}
	if res.Data != nil {
		var returnSalesInvoiceResp selling.ReturnSalesInvoiceResp
		if err := res.Data.UnmarshalTo(&returnSalesInvoiceResp); err != nil {
			logger.Logger.Error("ReturnSalesInvoice-UnmarshalTo", zap.Any("err", err))
			return nil, errors.WithMessage(err)
		}
		return &returnSalesInvoiceResp, nil
	}
	return nil, errors.WithMessage(errors.New("退款Sales Invoice异常, data为空"))
}

func (s *erpSrv) GetPaymentMethodList(ctx pkgCtx.Context, getPaymentReq req.GetPaymentMethodListReq) (*resp.GetPaymentMethodListResp, error) {
	company, err := repository.NewCompanyRepo(s.dbm.GetDB(constant.DefaultDB)).GetCompanyInfoByUuid(getPaymentReq.CompanyUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	if company.Uuid == 0 || company.CompanySetting == nil || company.CompanySetting.Uuid == 0 {
		return nil, errors.WithMessage(errors.New("商家信息异常"))
	}
	if !company.IsOpenErp() || company.CompanySetting.ErpnextSiteCode == "" {
		return nil, errors.WithMessage(errors.New("商家erp未开启"))
	}
	client, conn, err := NewErpSellingClient()
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	defer conn.Close()

	params := &selling.GetModeOfPaymentListReq{}
	result, err := client.GetModeOfPaymentList(WithSiteCode(ctx.GetContext(), company.CompanySetting.ErpnextSiteCode), params)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	if result.GetCode() != "0" || result.Data == nil {
		logger.Logger.Error("GetPaymentMethodList-GetModeOfPaymentList", zap.String("result_message", result.GetMessage()), zap.String("code", result.GetCode()))
		return nil, errors.WithMessage(errors.New(result.GetMessage()))
	}
	var getModeOfPaymentListResp selling.GetModeOfPaymentListResp
	if err := result.Data.UnmarshalTo(&getModeOfPaymentListResp); err != nil {
		logger.Logger.Error("GetModeOfPaymentList-UnmarshalTo", zap.Any("err", err))
		return nil, errors.WithMessage(err)
	}
	// 获取商家支付方式列表
	paymentMethodList := repository.NewPaymentMethodRepo(s.dbm.GetDB(getPaymentReq.CompanyUuid)).GetPaymentMethodList()
	// “免单”默认置灰
	var erpnextPayments []string
	for _, paymentMethod := range paymentMethodList {
		erpnextPayments = append(erpnextPayments, paymentMethod.ErpnextPayment)
	}
	// 获取支付方式列表
	var paymentMethodListResp []resp.PaymentMethodInfo
	for _, modeOfPayment := range getModeOfPaymentListResp.ModeOfPaymentList {
		if modeOfPayment.Name == "Free Meal" {
			continue
		}
		paymentMethodListResp = append(paymentMethodListResp, resp.PaymentMethodInfo{
			Name:      modeOfPayment.Name,
			IsAddable: !slices.Contains(erpnextPayments, modeOfPayment.Name),
		})
	}
	return &resp.GetPaymentMethodListResp{List: paymentMethodListResp}, nil
}

func (s *erpSrv) AddPaymentMethod(ctx pkgCtx.Context, addPaymentMethodReq req.AddPaymentMethodReq) error {
	db := s.dbm.GetDB(addPaymentMethodReq.CompanyUuid)
	company, err := repository.NewCompanyRepo(db).GetCompanyInfoByUuid(addPaymentMethodReq.CompanyUuid)
	if err != nil {
		return errors.WithMessage(err)
	}
	if company.Uuid == 0 || company.CompanySetting == nil || company.CompanySetting.Uuid == 0 {
		return errors.WithMessage(errors.New("商家信息异常"))
	}
	if !company.IsOpenErp() || company.CompanySetting.ErpnextSiteCode == "" {
		return errors.WithMessage(errors.New("商家erp未开启"))
	}
	// 判断支付方式名称是否已经存在
	paymentMethodRepo := repository.NewPaymentMethodRepo(db)
	exists, _ := paymentMethodRepo.Count(
		paymentMethodRepo.WhereNameSourceOrErpnextPayment(addPaymentMethodReq.Name, constant.PaymentMethodSourceDefault, addPaymentMethodReq.ErpnextPayment),
		repository.NewCommonRepo().WhereBySoftDelete(),
	)
	if exists > 0 {
		return errors.WithMessage(errors.New("支付方式已存在"))
	}
	client, conn, err := NewErpSellingClient()
	if err != nil {
		return errors.WithMessage(err)
	}
	defer conn.Close()

	params := &selling.CreatePaymentAccountReq{
		CompanyAbbr: company.CompanySetting.ErpnextCompanyAbbr,
		Branch:      company.CompanySetting.ErpnextBranchName,
		PaymentType: addPaymentMethodReq.ErpnextPayment,
	}
	result, err := client.CreatePaymentAccount(WithSiteCode(ctx.GetContext(), company.CompanySetting.ErpnextSiteCode), params)
	if err != nil {
		return errors.WithMessage(err)
	}
	if result.GetCode() != "0" {
		logger.Logger.Error("AddPaymentMethod-CreatePaymentAccount", zap.String("result_message", result.GetMessage()), zap.String("code", result.GetCode()))
		return errors.WithMessage(errors.New(result.GetMessage()))
	}
	// 获取来源是1的支付方式code最大值
	maxCode := paymentMethodRepo.GetMaxCodeBySource(constant.PaymentMethodSourceDefault, 20000)

	isShowCashier := 0
	isShowAssistant := 0
	isShowMemberRecharge := 0
	if slices.Contains(addPaymentMethodReq.CheckoutShow, "cashier") {
		isShowCashier = 1
	}
	if slices.Contains(addPaymentMethodReq.CheckoutShow, "assistant") {
		isShowAssistant = 1
	}
	if slices.Contains(addPaymentMethodReq.MemberRechargeShow, "cashier") {
		isShowMemberRecharge = 1
	}
	feePercent := decimal.NewFromFloat(*addPaymentMethodReq.Fee).Div(decimal.NewFromInt(100)).Round(4).InexactFloat64()
	paymentMethod := model.PaymentMethod{
		Name:                 addPaymentMethodReq.ErpnextPayment,
		Code:                 maxCode + 100,
		PaymentName:          addPaymentMethodReq.Name,
		Source:               constant.PaymentMethodSourceDefault,
		FeePercent:           feePercent,
		IsShowCashier:        isShowCashier,
		IsShowAssistant:      isShowAssistant,
		IsShowMemberRecharge: isShowMemberRecharge,
		Status:               *addPaymentMethodReq.Status,
		Sort:                 *addPaymentMethodReq.Sort,
		DefaultImg:           "/image/pay/ja_pay.png",
		ErpnextPayment:       addPaymentMethodReq.ErpnextPayment,
	}
	// 添加支付方式到商家数据库
	repository.NewPaymentMethodRepo(s.dbm.GetDB(addPaymentMethodReq.CompanyUuid)).CreatePaymentMethod(paymentMethod)
	return nil
}

// SaveModeOfPayment 保存/同步支付方式
func (s *erpSrv) SaveModeOfPayment(ctx pkgCtx.Context, saveModeOfPaymentReq req.SaveModeOfPaymentReq) (*selling.SaveModeOfPaymentResp, error) {
	db := s.dbm.GetDB(saveModeOfPaymentReq.CompanyUuid)
	company, err := repository.NewCompanyRepo(db).GetCompanyInfoByUuid(saveModeOfPaymentReq.CompanyUuid)
	if err != nil {
		logger.Logger.Error("SaveModeOfPayment-GetCompanyInfoByUuid", zap.Error(err))
		return nil, errors.WithMessage(err)
	}
	if company.Uuid == 0 || company.CompanySetting == nil || company.CompanySetting.Uuid == 0 {
		logger.Logger.Error("SaveModeOfPayment-GetCompanyInfoByUuid", zap.Error(err))
		return nil, errors.WithMessage(errors.New("商家信息异常"))
	}
	if !company.IsOpenErp() || company.CompanySetting.ErpnextSiteCode == "" {
		logger.Logger.Error("SaveModeOfPayment-IsOpenErp", zap.String("company_abbr", company.CompanySetting.ErpnextCompanyAbbr), zap.String("branch", company.CompanySetting.ErpnextBranchName))
		return nil, errors.WithMessage(errors.New("商家erp未开启"))
	}
	companySetting := ctx.GetCompanySetting()

	client, conn, err := NewErpSellingClient()
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	defer conn.Close()

	params := &selling.SaveModeOfPaymentReq{
		CompanyAbbr: companySetting.ErpnextCompanyAbbr,
		Branch:      companySetting.ErpnextBranchName,
		PayType:     saveModeOfPaymentReq.PayType,
	}
	if saveModeOfPaymentReq.Channel != "" {
		params.Channel = saveModeOfPaymentReq.Channel
	}
	if saveModeOfPaymentReq.Enabled != nil {
		params.Enabled = saveModeOfPaymentReq.Enabled
	}
	// 更新时优先使用 PaymentId
	if saveModeOfPaymentReq.PaymentId != nil {
		params.PaymentId = *saveModeOfPaymentReq.PaymentId
	}
	// 更新时同时传递 Name
	if saveModeOfPaymentReq.Name != nil {
		params.Name = saveModeOfPaymentReq.Name
	}
	if saveModeOfPaymentReq.AddedBy != nil && *saveModeOfPaymentReq.AddedBy != "" {
		params.AddedBy = saveModeOfPaymentReq.AddedBy
	}
	logger.Logger.Info("保存/同步支付方式", zap.Any("params", params))
	result, err := client.SaveModeOfPayment(WithSiteCode(ctx.GetContext(), companySetting.ErpnextSiteCode), params)
	if err != nil {
		logger.Logger.Error("SaveModeOfPayment-SaveModeOfPayment", zap.Error(err))
		return nil, errors.WithMessage(err)
	}
	if result.GetCode() != "0" {
		logger.Logger.Error("SaveModeOfPayment-SaveModeOfPayment", zap.String("result_message", result.GetMessage()), zap.String("code", result.GetCode()))
		return nil, errors.WithMessage(errors.New(result.GetMessage()))
	}
	if result.Data != nil {
		var saveModeOfPaymentResp selling.SaveModeOfPaymentResp
		if err := result.Data.UnmarshalTo(&saveModeOfPaymentResp); err != nil {
			logger.Logger.Error("SaveModeOfPayment-UnmarshalTo", zap.Any("err", err))
			return nil, errors.WithMessage(err)
		}
		return &saveModeOfPaymentResp, nil
	}
	return nil, errors.WithMessage(errors.New("保存支付方式异常, data为空"))
}
