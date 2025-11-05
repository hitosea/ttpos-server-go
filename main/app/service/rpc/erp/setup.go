package erp

import (
	"context"
	"errors"
	"slices"
	"strconv"
	"strings"
	"ttpos-bmp/app/ttpos-erp/api/setup"
	"ttpos-bmp/app/ttpos-erp/api/warehouse"
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
	companySetting, _ := companySettingRepo.GetOne(companySettingRepo.WhereErpnextCompanyAbbr(initShopReq.CompanyAbbr), companySettingRepo.WhereSiteCode(initShopReq.SiteCode), repository.NotDeleted)
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
	initReq := &setup.InitShopReq{
		ShopName:    company.Name,
		CompanyAbbr: initShopReq.CompanyAbbr,
		ShopUuid:    strconv.FormatUint(initShopReq.CompanyUuid, 10),
		AdminUuid:   strconv.FormatUint(staff.Uuid, 10),
	}
	result, err := client.InitShop(WithSiteCode(context.Background(), initShopReq.SiteCode), initReq)
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

	// 获取初始化创建的仓库
	warehouseClient, conn, err := NewErpWarehouseClient()
	if err != nil {
		return resp.InitShopResp{}, err
	}
	defer conn.Close()
	warehouseListReq := &warehouse.GetWarehouseListReq{
		CompanyAbbr: initShopReq.CompanyAbbr,
		Branch:      response.BranchName,
	}
	warehouseResult, err := warehouseClient.GetWarehouseList(WithSiteCode(context.Background(), initShopReq.SiteCode), warehouseListReq)
	if err != nil {
		return resp.InitShopResp{}, err
	}
	if warehouseResult.GetCode() != "0" || warehouseResult.Data == nil {
		logger.Logger.Error("InitShop-GetWarehouseList", zap.Any("err", err), zap.String("code", warehouseResult.GetCode()), zap.String("result_message", warehouseResult.GetMessage()))
		return resp.InitShopResp{}, errors.New("初始化失败")
	}
	warehouseList := &warehouse.GetWarehouseListResp{}
	if err := warehouseResult.Data.UnmarshalTo(warehouseList); err != nil {
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

	// 修改ttpos仓库erp_code
	var warehouses []model.Warehouse
	s.dbm.GetDB(company.Uuid).Model(&model.Warehouse{}).Scopes(repository.NotDeleted).Find(&warehouses)
	erpWarehouseNameMap := make(map[string]string)
	for _, erpWarehouse := range warehouseList.WarehouseList {
		if strings.Contains(erpWarehouse.Name, constant.NormalWarehouseCodeContains) {
			erpWarehouseNameMap[constant.NormalWarehouseCodeContains] = erpWarehouse.Name
		} else if strings.Contains(erpWarehouse.Name, constant.TransitWarehouseCodeContains) {
			erpWarehouseNameMap[constant.TransitWarehouseCodeContains] = erpWarehouse.Name
		}
	}
	for _, warehouse := range warehouses {
		if warehouse.IsDefault == 1 {
			s.dbm.GetDB(company.Uuid).Model(&model.Warehouse{}).Where("uuid = ?", warehouse.Uuid).Update("erp_code", erpWarehouseNameMap[constant.NormalWarehouseCodeContains])
		} else if warehouse.Type == "transit" {
			s.dbm.GetDB(company.Uuid).Model(&model.Warehouse{}).Where("uuid = ?", warehouse.Uuid).Update("erp_code", erpWarehouseNameMap[constant.TransitWarehouseCodeContains])
		}
	}

	// 自动同步了支付方式，cash, balance, lianlian(wechat, alipay, qr_promptpay)
	repository.NewPaymentMethodRepo(s.dbm.GetDB(company.Uuid)).InitErpnextPayment(map[int]string{
		constant.PaymentMethodCodeBalance:             "Balance",
		constant.PaymentMethodCodeCash:                "Cash",
		constant.PaymentMethodCodeLianLianWechatPay:   "LianlianPay-WeChat Pay",
		constant.PaymentMethodCodeLianLianAliPay:      "LianlianPay-AliPay",
		constant.PaymentMethodCodeLianLianQRPromptPay: "LianlianPay-QR PromptPay",
	})

	// ##### 更新父级公司UUID树 #####
	companyResp, err := NewIErpSrv(s.dbm).GetCompanyList(ctx, req.ErpnextSiteCompanyReq{
		SiteCode: initShopReq.SiteCode,
	})
	if err != nil {
		logger.Logger.Error("InitShop-GetCompanyList", zap.Any("err", err), zap.String("site_code", initShopReq.SiteCode))
	}
	// 遍历所有节点，如果IsUsed为true，则获取所有父级树的parent_company_uuid，以map[uuid][]parent_company_uuid形式存储，用于构建公司树
	companyUuidMap := NewIErpSrv(s.dbm).BuildCompanyUuidMap(companyResp.List)
	for uuid, parentUuids := range companyUuidMap {
		if uuid != company.Uuid {
			continue
		}
		hasChildren := 0
		for _, parentUuids2 := range companyUuidMap {
			if slices.Contains(parentUuids2, uuid) {
				hasChildren = 1
				break
			}
		}
		parentCompanyUuids := make([]string, len(parentUuids))
		for i, parentUuid := range parentUuids {
			parentCompanyUuids[i] = strconv.FormatUint(parentUuid, 10)
		}
		slices.Reverse(parentCompanyUuids)
		parentCompanyUuidsStr := strings.Join(parentCompanyUuids, ",")
		s.dbm.GetDB(0).Model(&model.CompanySetting{}).Where("company_uuid = ?", uuid).Updates(map[string]any{
			"parent_company_uuids": parentCompanyUuidsStr,
			"has_children":         hasChildren,
		})
		s.dbm.GetDB(uuid).Model(&model.CompanySetting{}).Where("company_uuid = ?", uuid).Updates(map[string]any{
			"parent_company_uuids": parentCompanyUuidsStr,
			"has_children":         hasChildren,
		})
	}

	return resp.InitShopResp{
		BranchName: response.BranchName,
		AdminEmail: response.AdminEmail,
	}, nil
}
