package erp

import (
	"context"
	"errors"
	"strconv"
	"ttpos-bmp/app/ttpos-erp/api/setup"
	"ttpos-server-go/app/cloud"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	cc "ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/logger"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func NewErpSetupClient() (setup.SetupServiceClient, *grpc.ClientConn, error) {
	conn, err := cloud.GetRpcConnWithName(cloud.ErpServiceName)
	if err != nil {
		return nil, nil, err
	}
	return setup.NewSetupServiceClient(conn), conn, nil
}

func (s *erpSrv) InitShop(ctx cc.Context, initShopReq req.InitShopReq) (resp.InitShopResp, error) {
	companySettingRepo := repository.NewCompanySettingRepo(s.dbm.GetDB(0))
	// 判断是否已经有店铺
	companySetting, _ := companySettingRepo.GetOne(companySettingRepo.WhereErpnextCompanyAbbr(initShopReq.CompanyAbbr))
	if companySetting.Uuid != 0 {
		return resp.InitShopResp{}, errors.New("所属erpnext公司已授权其他商家")
	}

	// 判断商家是否存在
	companyRepo := repository.NewCompanyRepo(s.dbm.GetDB(0))
	company, _ := companyRepo.GetCompanyInfoByUuid(initShopReq.CompanyUuid)
	if company.Uuid == 0 {
		return resp.InitShopResp{}, errors.New("商家不存在")
	}

	// 判断商家超管是否存在
	staffRepo := repository.NewStaffRepo(s.dbm.GetDB(initShopReq.CompanyUuid))
	staff, _ := staffRepo.GetStaff(staffRepo.WhereIsSuper(1))
	if staff.Uuid == 0 {
		return resp.InitShopResp{}, errors.New("商家超管不存在")
	}

	var headquarterUuid uint64
	headquarterAbbr := initShopReq.HeadquarterAbbr
	if initShopReq.SiteCode != "1" && headquarterAbbr == "" {
		return resp.InitShopResp{}, errors.New("连锁店总部未设置")
	}
	// 连锁店-总部关联处理
	if headquarterAbbr != "" {
		var headquarter model.CompanySetting
		s.dbm.GetDB(0).Model(&model.CompanySetting{}).Where("erpnext_site_code = ? AND erpnext_company_abbr = ?", initShopReq.SiteCode, initShopReq.HeadquarterAbbr).Scopes(repository.NotDeleted).First(&headquarter)
		headquarterUuid = headquarter.Uuid
	}

	client, conn, err := NewErpSetupClient()
	if err != nil {
		return resp.InitShopResp{}, err
	}
	defer conn.Close()
	req := &setup.InitShopReq{
		ShopName:    company.Name,
		CompanyAbbr: initShopReq.CompanyAbbr,
		ShopUuid:    strconv.FormatUint(initShopReq.CompanyUuid, 10),
		AdminUuid:   strconv.FormatUint(staff.Uuid, 10),
	}
	result, err := client.InitShop(WithSiteCode(context.Background(), initShopReq.SiteCode), req)
	if err != nil {
		return resp.InitShopResp{}, err
	}
	if result.GetCode() != "0" || result.Data == nil {
		logger.Logger.Error("InitShop-InitShop", zap.Any("err", err), zap.String("code", result.GetCode()), zap.String("result_message", result.GetMessage()))
		return resp.InitShopResp{}, errors.New("初始化失败")
	}
	response := &setup.InitShopResp{}
	if err := result.Data.UnmarshalTo(response); err != nil {
		return resp.InitShopResp{}, err
	}

	// 更新saas库
	s.dbm.GetDB(0).Model(&model.Company{}).Where("uuid = ?", company.Uuid).Updates(map[string]any{
		"is_enable_erp": 1,
	})
	s.dbm.GetDB(0).Model(&model.CompanySetting{}).Where("company_uuid = ?", company.Uuid).Updates(map[string]any{
		"erpnext_site_code":        initShopReq.SiteCode,
		"erpnext_company_abbr":     initShopReq.CompanyAbbr,
		"erpnext_branch_name":      response.BranchName,
		"erpnext_pos_profile_name": response.PosProfile,
		"erpnext_admin_email":      response.AdminEmail,
		"erpnext_headquarter_abbr": headquarterAbbr,
		"headquarter_uuid":         headquarterUuid,
	})

	// 更新商家库
	s.dbm.GetDB(company.Uuid).Model(&model.Company{}).Where("uuid = ?", company.Uuid).Updates(map[string]any{
		"is_enable_erp": 1,
	})
	s.dbm.GetDB(company.Uuid).Model(&model.CompanySetting{}).Where("company_uuid = ?", company.Uuid).Updates(map[string]any{
		"erpnext_site_code":        initShopReq.SiteCode,
		"erpnext_company_abbr":     initShopReq.CompanyAbbr,
		"erpnext_branch_name":      response.BranchName,
		"erpnext_pos_profile_name": response.PosProfile,
		"erpnext_admin_email":      response.AdminEmail,
		"erpnext_headquarter_abbr": headquarterAbbr,
		"headquarter_uuid":         headquarterUuid,
	})

	// 自动同步了支付方式，cash, balance, lianlian(wechat, alipay, qr_promptpay)
	repository.NewPaymentMethodRepo(s.dbm.GetDB(company.Uuid)).InitErpnextPayment(map[int]string{
		constant.PaymentMethodCodeBalance:             "Balance",
		constant.PaymentMethodCodeCash:                "Cash",
		constant.PaymentMethodCodeLianLianWechatPay:   "LianlianPay-WeChat Pay",
		constant.PaymentMethodCodeLianLianAliPay:      "LianlianPay-AliPay",
		constant.PaymentMethodCodeLianLianQRPromptPay: "LianlianPay-QR PromptPay",
	})

	return resp.InitShopResp{
		BranchName: response.BranchName,
		AdminEmail: response.AdminEmail,
	}, nil
}
