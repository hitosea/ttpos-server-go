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
	var getPosProfileListResp resp.GetPosProfileListResp
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
			getPosProfileListResp.ProfileList = append(getPosProfileListResp.ProfileList, resp.PosProfileInfo{
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
