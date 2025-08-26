package erp

import (
	"context"
	"slices"
	"ttpos-bmp/app/ttpos-erp/api/selling"
	"ttpos-server-go/app/cloud"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/model"
	pkgCtx "ttpos-server-go/pkg/context"

	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/repository"
	cc "ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/logger"

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
	if result.Data != nil {
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

	company, err := repository.NewCompanyRepo(s.dbm.GetDB(0)).GetCompanyInfoByUuid(req.CompanyUuid)
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
			openPosEntryDetail = append(openPosEntryDetail, &selling.OpenPosEntryDetail{
				ModeOfPayment: detail.ModeOfPayment,
				OpeningAmount: detail.OpeningAmount,
			})
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
	res, err := client.OpenPosEntry(WithSiteCode(ctx, openEntryReq.SiteCode), openPosEntryReq)
	if err != nil {
		return "", err
	}
	if res.Data == nil {
		return "", errors.New(res.Message)
	}
	var openPosEntryResp selling.OpenPosEntryResp
	if err := res.Data.UnmarshalTo(&openPosEntryResp); err != nil {
		logger.Logger.Error("OpenPosEntry-UnmarshalTo", zap.Any("err", err))
		return "", err
	}
	if openPosEntryResp.OpenPosEntryInfo != nil {
		return openPosEntryResp.OpenPosEntryInfo.OpenPosEntryName, nil
	}
	return "", errors.New(res.Message)
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
			closePosEntryDetail = append(closePosEntryDetail, &selling.ClosePosEntryDetail{
				ModeOfPayment: detail.ModeOfPayment,
				OpeningAmount: detail.OpeningAmount,
				ClosingAmount: detail.ClosingAmount,
			})
		}
	}
	closePosEntryReq := &selling.ClosePosEntryReq{
		PosOpenEntryName:    closeEntryReq.PosOpenEntryName,
		PeriodEndDate:       closeEntryReq.PeriodEndDate,
		ClosePosEntryDetail: closePosEntryDetail,
	}
	res, err := client.ClosePosEntry(WithSiteCode(ctx, closeEntryReq.SiteCode), closePosEntryReq)
	if err != nil {
		return "", err
	}
	if res.Data != nil {
		var closePosEntryResp selling.ClosePosEntryResp
		if err := res.Data.UnmarshalTo(&closePosEntryResp); err != nil {
			logger.Logger.Error("ClosePosEntry-UnmarshalTo", zap.Any("err", err))
			return "", err
		}
		if closePosEntryResp.ClosePosEntryInfo != nil {
			return closePosEntryResp.ClosePosEntryInfo.ClosePosEntryName, nil
		}
	}
	return "", nil
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
	}
	res, err := client.SavePosInvoice(WithSiteCode(ctx.GetContext(), savePosInvoiceReq.SiteCode), params)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	if res.GetCode() != "0" {
		return nil, errors.WithMessage(errors.New(res.Message))
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

func (s *erpSrv) GetPaymentMethodList(ctx pkgCtx.Context, getPaymentReq req.GetPaymentMethodListReq) (*resp.GetPaymentMethodListResp, error) {
	company, err := repository.NewCompanyRepo(s.dbm.GetDB(0)).GetCompanyInfoByUuid(getPaymentReq.CompanyUuid)
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

	params := &selling.GetModeOfPaymentListReq{
		CompanyAbbr: company.CompanySetting.ErpnextCompanyAbbr,
		Branch:      company.CompanySetting.ErpnextBranchName,
	}
	res, err := client.GetModeOfPaymentList(WithSiteCode(ctx.GetContext(), company.CompanySetting.ErpnextSiteCode), params)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	if res.GetCode() != "0" {
		return nil, errors.WithMessage(errors.New(res.Message))
	}
	if res.Data != nil {
		var getModeOfPaymentListResp selling.GetModeOfPaymentListResp
		if err := res.Data.UnmarshalTo(&getModeOfPaymentListResp); err != nil {
			logger.Logger.Error("GetModeOfPaymentList-UnmarshalTo", zap.Any("err", err))
			return nil, errors.WithMessage(err)
		}
		// 获取商家支付方式列表
		paymentMethodList := repository.NewPaymentMethodRepo(s.dbm.GetDB(0)).GetPaymentMethodList()
		var erpnextPayments []string
		for _, paymentMethod := range paymentMethodList {
			erpnextPayments = append(erpnextPayments, paymentMethod.ErpnextPayment)
		}
		// 获取支付方式列表
		var paymentMethodListResp []resp.PaymentMethodInfo
		for _, modeOfPayment := range getModeOfPaymentListResp.ModeOfPaymentList {
			paymentMethodListResp = append(paymentMethodListResp, resp.PaymentMethodInfo{
				Name:   modeOfPayment.Name,
				IsUsed: slices.Contains(erpnextPayments, modeOfPayment.Name),
			})
		}
		return &resp.GetPaymentMethodListResp{List: paymentMethodListResp}, nil
	}
	return nil, errors.WithMessage(errors.New(res.Message))
}

func (s *erpSrv) AddPaymentMethod(ctx pkgCtx.Context, addPaymentMethodReq req.AddPaymentMethodReq) error {
	company, err := repository.NewCompanyRepo(s.dbm.GetDB(0)).GetCompanyInfoByUuid(addPaymentMethodReq.CompanyUuid)
	if err != nil {
		return errors.WithMessage(err)
	}
	if company.Uuid == 0 || company.CompanySetting == nil || company.CompanySetting.Uuid == 0 {
		return errors.WithMessage(errors.New("商家信息异常"))
	}
	if !company.IsOpenErp() || company.CompanySetting.ErpnextSiteCode == "" {
		return errors.WithMessage(errors.New("商家erp未开启"))
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
	res, err := client.CreatePaymentAccount(WithSiteCode(ctx.GetContext(), company.CompanySetting.ErpnextSiteCode), params)
	if err != nil {
		return errors.WithMessage(err)
	}
	if res.GetCode() != "0" {
		return errors.WithMessage(errors.New(res.Message))
	}
	// 获取来源是1的支付方式code最大值

	db := s.dbm.GetDB(addPaymentMethodReq.CompanyUuid)
	var maxCode = 20000
	db.Model(&model.PaymentMethod{}).Where("source = ?", constant.PaymentMethodSourceDefault).Scopes(repository.NotDeleted).Select("ifnull(max(code), 20000) as code").Scan(&maxCode)

	isShowCashier := 0
	isShowAssistant := 0
	isShowMemberRecharge := 0
	if slices.Contains(addPaymentMethodReq.CheckoutShow, "cashier") {
		isShowCashier = 1
	}
	if slices.Contains(addPaymentMethodReq.CheckoutShow, "assistant") {
		isShowAssistant = 1
	}
	if slices.Contains(addPaymentMethodReq.CheckoutShow, "member_recharge") {
		isShowMemberRecharge = 1
	}
	paymentMethod := model.PaymentMethod{
		Name:                 addPaymentMethodReq.Name,
		Code:                 maxCode + 100,
		PaymentName:          addPaymentMethodReq.Name,
		Source:               constant.PaymentMethodSourceDefault,
		FeePercent:           *addPaymentMethodReq.Fee,
		IsShowCashier:        isShowCashier,
		IsShowAssistant:      isShowAssistant,
		IsShowMemberRecharge: isShowMemberRecharge,
		Status:               *addPaymentMethodReq.Status,
		Sort:                 *addPaymentMethodReq.Sort,
		DefaultImg:           "/image/pay/ja_pay.png",
	}
	// 添加支付方式到商家数据库
	repository.NewPaymentMethodRepo(s.dbm.GetDB(addPaymentMethodReq.CompanyUuid)).CreatePaymentMethod(paymentMethod)
	return nil
}
