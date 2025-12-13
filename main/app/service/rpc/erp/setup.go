package erp

import (
	"context"
	"errors"
	"slices"
	"strconv"
	"strings"
	"ttpos-bmp/app/ttpos-erp/api/selling"
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
	"ttpos-server-go/pkg/utils"

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
	companySettingRepo := repository.NewCompanySettingRepo(s.dbm.GetDB(constant.DefaultDB))
	// 判断是否已经有店铺
	companySetting, _ := companySettingRepo.GetOne(companySettingRepo.WhereErpnextCompanyAbbr(initShopReq.CompanyAbbr), companySettingRepo.WhereSiteCode(initShopReq.SiteCode), repository.NotDeleted)
	if companySetting.Uuid != 0 {
		return resp.InitShopResp{}, errors.New("所属erpnext公司已授权其他商家")
	}

	// 判断商家是否存在
	companyRepo := repository.NewCompanyRepo(s.dbm.GetDB(constant.DefaultDB))
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
		s.dbm.GetDB(constant.DefaultDB).Model(&model.CompanySetting{}).Where("erpnext_site_code = ? AND erpnext_company_abbr = ?", initShopReq.SiteCode, initShopReq.HeadquarterAbbr).Scopes(repository.NotDeleted).First(&headquarter)
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
	s.dbm.GetDB(constant.DefaultDB).Model(&model.Company{}).Where("uuid = ?", company.Uuid).Updates(map[string]any{
		"is_enable_erp": 1,
	})
	s.dbm.GetDB(constant.DefaultDB).Model(&model.CompanySetting{}).Where("company_uuid = ?", company.Uuid).Updates(map[string]any{
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

	paymentClient, conn, err := NewErpSellingClient()
	if err != nil {
		return resp.InitShopResp{}, err
	}
	defer conn.Close()
	erpnextPaymentMap := make(map[int]string)
	paymentList := repository.NewPaymentMethodRepo(s.dbm.GetDB(company.Uuid)).GetPaymentMethodList()
	for _, payment := range paymentList {
		// 如果已存在，则跳过
		if payment.ErpnextPayment != "" {
			continue
		}
		// 如果支付方式是总部同步的，且是自行添加，且没有二维码图片，则跳过
		if payment.IsHeadquarterPayment() {
			if payment.Source == constant.PaymentMethodSourceDefault && payment.QrcodeFileUuid == 0 {
				continue
			}
		}

		saveModeOfPaymentResp, err := paymentClient.SaveModeOfPayment(WithSiteCode(context.Background(), initShopReq.SiteCode), &selling.SaveModeOfPaymentReq{
			CompanyAbbr: initShopReq.CompanyAbbr,
			Branch:      response.BranchName,
			Channel:     GetChannelBySource(payment.Source),
			PayType:     payment.PaymentName,
		})
		if err != nil {
			logger.Logger.Error("InitShop-SaveModeOfPayment-1", zap.Any("err", err), zap.String("payment_name", payment.PaymentName), zap.String("company_abbr", initShopReq.CompanyAbbr))
			continue
		}
		if saveModeOfPaymentResp.GetCode() != "0" || saveModeOfPaymentResp.Data == nil {
			logger.Logger.Error("InitShop-SaveModeOfPayment-2", zap.Any("code", saveModeOfPaymentResp.GetCode()), zap.String("payment_name", payment.PaymentName), zap.String("company_abbr", initShopReq.CompanyAbbr))
			continue
		}
		result := &selling.SaveModeOfPaymentResp{}
		if err := saveModeOfPaymentResp.Data.UnmarshalTo(result); err != nil {
			logger.Logger.Error("InitShop-SaveModeOfPayment-3", zap.Any("err", err), zap.String("payment_name", payment.PaymentName), zap.String("company_abbr", initShopReq.CompanyAbbr))
			continue
		}
		erpnextPaymentMap[payment.Code] = result.Name
	}

	// 自动同步了支付方式到erpnext
	repository.NewPaymentMethodRepo(s.dbm.GetDB(company.Uuid)).InitErpnextPayment(erpnextPaymentMap)

	utils.SafeGo(func() {
		err := s.UpdateTtposCompanyParentUuids(initShopReq.SiteCode)
		if err != nil {
			logger.Logger.Error("InitShop-UpdateParentUuids", zap.Any("err", err), zap.String("site_code", initShopReq.SiteCode))
		}
	})

	return resp.InitShopResp{
		BranchName: response.BranchName,
		AdminEmail: response.AdminEmail,
	}, nil
}

func (s *erpSrv) UpdateTtposCompanyParentUuids(siteCode string) error {
	// 获取erp company
	erpnextSiteCompanyReq := req.ErpnextSiteCompanyReq{
		SiteCode: siteCode,
	}
	ctx := cc.NewContext()
	// 调用erpnext服务，获取公司名称
	companyResp, err := s.GetCompanyList(ctx, erpnextSiteCompanyReq)
	if err != nil {
		return err
	}
	// 遍历所有节点，如果IsUsed为true，则获取所有父级树的parent_company_uuid，以map[uuid][]CompanyInfo形式存储
	companyAbbrMap, companyAbbrUuidMap := s.buildCompanyAbbrMap(companyResp.List)
	for companyAbbr, parentCompanyInfos := range companyAbbrMap {
		hasChildren := 0
		// 检查当前UUID是否在其他公司的父级路径中（判断是否有子公司）
		for _, parentCompanyInfos2 := range companyAbbrMap {
			for _, info := range parentCompanyInfos2 {
				if info.CompanyAbbr == companyAbbr {
					hasChildren = 1
					break
				}
			}
			if hasChildren == 1 {
				break
			}
		}
		// 从 CompanyInfo 中提取 UUID 并转换为字符串
		parentCompanyUuids := make([]string, 0)
		for _, info := range parentCompanyInfos {
			if info.CompanyUuid == 0 {
				continue
			}
			parentCompanyUuids = append(parentCompanyUuids, strconv.FormatUint(info.CompanyUuid, 10))
		}

		slices.Reverse(parentCompanyUuids)
		parentCompanyUuidsStr := strings.Join(parentCompanyUuids, ",")

		s.dbm.GetDB(0).Model(&model.CompanySetting{}).Where("erpnext_site_code = ? AND erpnext_company_abbr = ?", siteCode, companyAbbr).Updates(map[string]any{
			"parent_company_uuids": parentCompanyUuidsStr,
			"has_children":         hasChildren,
		})
		uuid := companyAbbrUuidMap[companyAbbr]
		if uuid > 0 {
			s.dbm.GetDB(uuid).Model(&model.CompanySetting{}).Where("erpnext_site_code = ? AND erpnext_company_abbr = ?", siteCode, companyAbbr).Updates(map[string]any{
				"parent_company_uuids": parentCompanyUuidsStr,
				"has_children":         hasChildren,
			})
		}
	}
	return nil
}
