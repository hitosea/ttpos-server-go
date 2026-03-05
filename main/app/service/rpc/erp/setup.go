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
	commonRepo := repository.NewCommonRepo()

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
		saasCompanySettingRepo := repository.NewCompanySettingRepo(s.dbm.GetDB(constant.DefaultDB))
		headquarter, _ := saasCompanySettingRepo.GetOne(saasCompanySettingRepo.WhereSiteCode(initShopReq.SiteCode), saasCompanySettingRepo.WhereErpnextCompanyAbbr(initShopReq.HeadquarterAbbr), repository.CommonRepo.WhereBySoftDelete())
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
	erpCompanyData := map[string]any{"is_enable_erp": 1}
	erpSettingData := map[string]any{
		"erpnext_site_code":        initShopReq.SiteCode,
		"erpnext_company_abbr":     initShopReq.CompanyAbbr,
		"erpnext_branch_name":      response.BranchName,
		"erpnext_pos_profile_name": response.PosProfile,
		"erpnext_admin_email":      response.AdminEmail,
		"erpnext_headquarter_abbr": headquarterAbbr,
		"headquarter_uuid":         headquarterUuid,
	}
	repository.NewCompanyRepo(s.dbm.GetDB(constant.DefaultDB)).UpdateCompany(company.Uuid, erpCompanyData)
	repository.NewCompanySettingRepo(s.dbm.GetDB(constant.DefaultDB)).UpdateByCompanyUuid(company.Uuid, erpSettingData)

	// 更新商家库
	repository.NewCompanyRepo(s.dbm.GetDB(company.Uuid)).UpdateCompany(company.Uuid, erpCompanyData)
	repository.NewCompanySettingRepo(s.dbm.GetDB(company.Uuid)).UpdateByCompanyUuid(company.Uuid, erpSettingData)

	// 修改ttpos仓库erp_code
	warehouseRepo := repository.NewWarehouseRepo(s.dbm.GetDB(company.Uuid))
	warehouses, _ := warehouseRepo.Get(repository.CommonRepo.WhereBySoftDelete())
	erpWarehouseNameMap := make(map[string]string)
	for _, erpWarehouse := range warehouseList.WarehouseList {
		if strings.Contains(erpWarehouse.Name, constant.NormalWarehouseCodeContains) {
			erpWarehouseNameMap[constant.NormalWarehouseCodeContains] = erpWarehouse.Name
		} else if strings.Contains(erpWarehouse.Name, constant.TransitWarehouseCodeContains) {
			erpWarehouseNameMap[constant.TransitWarehouseCodeContains] = erpWarehouse.Name
		}
	}
	for _, wh := range warehouses {
		if wh.IsDefault == 1 {
			warehouseRepo.UpdateByUuid(wh.Uuid, map[string]any{"erp_code": erpWarehouseNameMap[constant.NormalWarehouseCodeContains]})
		} else if wh.Type == "transit" {
			warehouseRepo.UpdateByUuid(wh.Uuid, map[string]any{"erp_code": erpWarehouseNameMap[constant.TransitWarehouseCodeContains]})
		}
	}

	// 同步支付方式到ERP
	sellingClient, conn, err := NewErpSellingClient()
	if err != nil {
		return resp.InitShopResp{}, err
	}
	defer conn.Close()

	paymentMethodRepo := repository.NewPaymentMethodRepo(s.dbm.GetDB(company.Uuid))

	// 定义基础支付方式列表（仅同步 Cash、Balance、Free Meal for ERP）
	basePaymentCodes := []int{
		constant.PaymentMethodCodeCash,           // 40
		constant.PaymentMethodCodeBalance,        // 10
		constant.PaymentMethodCodeFreeMealForErp, // 92000（用于ERP同步的Free Meal）
	}

	// 确保 Free Meal for ERP 存在（code=92000，不改变原有的Free Meal code=-1）
	freeMealForErpPayment := paymentMethodRepo.GetPaymentMethod(
		paymentMethodRepo.WhereCode(constant.PaymentMethodCodeFreeMealForErp),
		commonRepo.WhereBySoftDelete(),
	)
	if freeMealForErpPayment.Uuid == 0 {
		// Free Meal for ERP 不存在，创建它
		maxSort, err := paymentMethodRepo.GetMaxSort()
		if err != nil {
			logger.Logger.Error("InitShop-GetMaxSort", zap.Error(err))
		} else {
			freeMealForErpPayment = model.PaymentMethod{
				Code:                 constant.PaymentMethodCodeFreeMealForErp,
				Name:                 "Free Meal",
				PaymentName:          "Free Meal",
				Source:               constant.PaymentMethodSourceSystem,
				Status:               constant.PaymentMethodStatusEnable,
				Sort:                 maxSort + 1,
				DefaultImg:           "/image/pay/ja_pay.png",
				IsShowCashier:        0, // 不在收银机显示
				IsShowAssistant:      0, // 不在助手端显示
				IsShowKiosk:          0,
				IsShowMemberRecharge: 0,
				FeePercent:           0.0000,
			}
			if err := paymentMethodRepo.CreatePaymentMethod(freeMealForErpPayment); err != nil {
				logger.Logger.Error("InitShop-CreateFreeMealForErp", zap.Error(err))
			} else {
				logger.Logger.Info("InitShop-CreateFreeMealForErp", zap.String("name", "Free Meal"), zap.Int("code", constant.PaymentMethodCodeFreeMealForErp))
			}
		}
	}

	saveModeOfPaymentRespMap := make(map[int]repository.ErpnextPaymentInfo)
	// 只同步基础支付方式
	for _, code := range basePaymentCodes {
		paymentMethod := paymentMethodRepo.GetPaymentMethod(
			paymentMethodRepo.WhereCode(code),
			commonRepo.WhereBySoftDelete(),
		)
		if paymentMethod.Uuid == 0 {
			continue // 支付方式不存在，跳过
		}

		// 如果已有 erpnext_payment 或 erpnext_payment_id，跳过
		if paymentMethod.ErpnextPayment != "" || paymentMethod.ErpnextPaymentId != "" {
			continue
		}

		// 根据 source 获取 channel
		channel := GetChannelBySource(paymentMethod.Source)
		addedBy := ""
		if paymentMethod.Source == constant.PaymentMethodSourceSystem {
			addedBy = "sys"
		}
		params := &selling.SaveModeOfPaymentReq{
			CompanyAbbr: initShopReq.CompanyAbbr,
			Branch:      response.BranchName,
			Channel:     channel,
			PayType:     paymentMethod.PaymentName,
			AddedBy:     &addedBy,
		}
		saveResp, err := sellingClient.SaveModeOfPayment(WithSiteCode(context.Background(), initShopReq.SiteCode), params)
		if err != nil {
			logger.Logger.Error(
				"InitShop-SaveModeOfPayment",
				zap.Any("err", err),
				zap.Any("params", params),
				zap.Any("company_uuid", company.Uuid),
				zap.Any("payment_method_uuid", paymentMethod.Uuid),
				zap.Int("code", code),
			)
			continue
		}
		if saveResp.GetCode() != "0" || saveResp.Data == nil {
			logger.Logger.Error(
				"InitShop-SaveModeOfPayment",
				zap.Any("code", saveResp.GetCode()),
				zap.Any("message", saveResp.GetMessage()),
				zap.Any("params", params),
				zap.Any("company_uuid", company.Uuid),
				zap.Any("payment_method_uuid", paymentMethod.Uuid),
				zap.Int("code", code),
			)
			continue
		}
		var saveModeOfPaymentResp selling.SaveModeOfPaymentResp
		if err := saveResp.Data.UnmarshalTo(&saveModeOfPaymentResp); err != nil {
			logger.Logger.Error(
				"InitShop-SaveModeOfPayment-UnmarshalTo",
				zap.Any("err", err),
				zap.Any("params", params),
				zap.Any("company_uuid", company.Uuid),
				zap.Any("payment_method_uuid", paymentMethod.Uuid),
				zap.Int("code", code),
			)
			continue
		}
		// 保存 Name 和 PaymentId
		saveModeOfPaymentRespMap[paymentMethod.Code] = repository.ErpnextPaymentInfo{
			Name:      saveModeOfPaymentResp.Name,
			PaymentId: saveModeOfPaymentResp.PaymentId,
		}
	}

	if len(saveModeOfPaymentRespMap) > 0 {
		paymentMethodRepo.InitErpnextPaymentWithId(saveModeOfPaymentRespMap)
	}

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

		treeUpdateData := map[string]any{
			"parent_company_uuids": parentCompanyUuidsStr,
			"has_children":         hasChildren,
		}
		repository.NewCompanySettingRepo(s.dbm.GetDB(0)).UpdateBySiteCodeAndCompanyAbbr(siteCode, companyAbbr, treeUpdateData)
		uuid := companyAbbrUuidMap[companyAbbr]
		if uuid > 0 {
			repository.NewCompanySettingRepo(s.dbm.GetDB(uuid)).UpdateBySiteCodeAndCompanyAbbr(siteCode, companyAbbr, treeUpdateData)
		}
	}
	return nil
}
