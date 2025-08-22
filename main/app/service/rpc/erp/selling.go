package erp

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/api/selling"
	"ttpos-server-go/app/cloud"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
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
	if res.Data != nil {
		var openPosEntryResp selling.OpenPosEntryResp
		if err := res.Data.UnmarshalTo(&openPosEntryResp); err != nil {
			logger.Logger.Error("OpenPosEntry-UnmarshalTo", zap.Any("err", err))
			return "", err
		}
		if openPosEntryResp.OpenPosEntryInfo != nil {
			return openPosEntryResp.OpenPosEntryInfo.OpenPosEntryName, nil
		}
	}
	return "", nil
}

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
