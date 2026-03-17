package service

import (
	"fmt"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/language"
	"ttpos-server-go/i18n"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// IAutoReceiptSrv 自动收货规则服务接口
type IAutoReceiptSrv interface {
	CreateRule(ctx context.Context, r req.CreateAutoReceiptRuleReq) error
	UpdateRule(ctx context.Context, r req.UpdateAutoReceiptRuleReq) error
	DeleteRule(ctx context.Context, r req.DeleteAutoReceiptRuleReq) error
	GetRuleList(ctx context.Context, r req.AutoReceiptRuleListReq) (resp.AutoReceiptRuleListResp, error)
	GetShopList(ctx context.Context, r req.AutoReceiptShopListReq) (resp.AutoReceiptShopListResp, error)
	GetLogList(ctx context.Context, r req.AutoReceiptLogListReq) (resp.AutoReceiptLogListResp, error)
	GetLogDetail(ctx context.Context, r req.AutoReceiptLogDetailReq) (shopCtx context.Context, receiptOrderUuid uint64, err error)
	GetWarehouseList(ctx context.Context) (resp.AutoReceiptWarehouseListResp, error)
}

// autoReceiptSrv 自动收货规则服务实现
type autoReceiptSrv struct {
	dbm   *database.DBManager
	cache cache.Cache
}

// NewAutoReceiptSrv 创建自动收货规则服务
func NewAutoReceiptSrv(dbm *database.DBManager, cache cache.Cache) IAutoReceiptSrv {
	return &autoReceiptSrv{dbm: dbm, cache: cache}
}

// getSaasDB 获取 saas 主库连接
func (s *autoReceiptSrv) getSaasDB() *gorm.DB {
	return s.dbm.GetDB(constant.DefaultDB)
}

// checkHeadquarter 校验是否为总店，非总店不允许操作
func (s *autoReceiptSrv) checkHeadquarter(ctx context.Context) error {
	companySetting := ctx.GetCompanySetting()
	if !companySetting.IsHeadquarter() {
		return errors.New("只有总店可以操作自动收货配置")
	}
	return nil
}

// getHeadquarterUuid 获取总部UUID（仅总店调用，直接返回当前公司UUID）
func (s *autoReceiptSrv) getHeadquarterUuid(ctx context.Context) uint64 {
	return ctx.GetCompanyUuid()
}

// CreateRule 创建自动收货规则
func (s *autoReceiptSrv) CreateRule(ctx context.Context, r req.CreateAutoReceiptRuleReq) error {
	if err := s.checkHeadquarter(ctx); err != nil {
		return err
	}
	db := s.getSaasDB()
	headquarterUuid := s.getHeadquarterUuid(ctx)

	// 事务内校验 + 创建（避免 TOCTOU 竞态）
	err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		shopRepo := repository.NewAutoReceiptRuleShopRepo(tx)

		// 校验门店是否已在同仓库其他规则中配置
		configuredUuids, err := shopRepo.GetConfiguredShopUuids(headquarterUuid, r.WarehouseErpCode, 0)
		if err != nil {
			return errors.WithMessage(err, "查询已配置门店失败")
		}
		configuredMap := make(map[uint64]bool, len(configuredUuids))
		for _, uid := range configuredUuids {
			configuredMap[uid] = true
		}
		for _, shopUuid := range r.ShopUuids {
			if configuredMap[shopUuid] {
				return errors.New(fmt.Sprintf(i18n.Translate(ctx.GetLanguage(), "门店 %d 已在该仓库的其他规则中配置"), shopUuid))
			}
		}

		ruleRepo := repository.NewAutoReceiptRuleRepo(tx)
		rule, err := ruleRepo.Create(model.AutoReceiptRule{
			HeadquarterCompanyUuid: headquarterUuid,
			Name:                   r.LocaleName.ToJson(),
			WarehouseErpCode:       r.WarehouseErpCode,
			DelayDays:              r.DelayDays,
			Status:                 r.Status,
		})
		if err != nil {
			return errors.WithMessage(err, "创建规则失败")
		}

		shops := make([]model.AutoReceiptRuleShop, 0, len(r.ShopUuids))
		for _, shopUuid := range r.ShopUuids {
			shops = append(shops, model.AutoReceiptRuleShop{
				RuleUuid:        rule.Uuid,
				ShopCompanyUuid: shopUuid,
			})
		}
		if err := shopRepo.BatchCreate(shops); err != nil {
			return errors.WithMessage(err, "创建规则门店失败")
		}
		return nil
	})

	return err
}

// UpdateRule 更新自动收货规则（含门店全量替换）
func (s *autoReceiptSrv) UpdateRule(ctx context.Context, r req.UpdateAutoReceiptRuleReq) error {
	if err := s.checkHeadquarter(ctx); err != nil {
		return err
	}
	db := s.getSaasDB()
	headquarterUuid := s.getHeadquarterUuid(ctx)

	err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		ruleRepo := repository.NewAutoReceiptRuleRepo(tx)
		_, err := ruleRepo.GetByUuid(r.Uuid, headquarterUuid)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return errors.New("规则不存在")
			}
			return errors.WithMessage(err, "查询规则失败")
		}

		// 全量更新规则主表字段
		vars := map[string]any{
			"name":               r.LocaleName.ToJson(),
			"warehouse_erp_code": r.WarehouseErpCode,
			"delay_days":         r.DelayDays,
			"status":             r.Status,
		}
		if err := ruleRepo.Update(r.Uuid, headquarterUuid, vars); err != nil {
			return errors.WithMessage(err, "更新规则失败")
		}

		shopRepo := repository.NewAutoReceiptRuleShopRepo(tx)

		// 获取当前门店列表
		currentShops, err := shopRepo.GetByRuleUuids([]uint64{r.Uuid})
		if err != nil {
			return errors.WithMessage(err, "查询当前门店失败")
		}
		currentMap := make(map[uint64]uint64, len(currentShops)) // shopCompanyUuid → subTableUuid
		for _, shop := range currentShops {
			currentMap[shop.ShopCompanyUuid] = shop.Uuid
		}

		// 计算 diff
		newSet := make(map[uint64]bool, len(r.ShopUuids))
		for _, uid := range r.ShopUuids {
			newSet[uid] = true
		}

		// 需要删除的门店（在 current 不在 new）
		var toDeleteUuids []uint64
		for shopUuid, subUuid := range currentMap {
			if !newSet[shopUuid] {
				toDeleteUuids = append(toDeleteUuids, subUuid)
			}
		}

		// 需要新增的门店（在 new 不在 current）
		var toAdd []uint64
		for _, shopUuid := range r.ShopUuids {
			if _, exists := currentMap[shopUuid]; !exists {
				toAdd = append(toAdd, shopUuid)
			}
		}

		// 校验新增门店是否已在其他规则中配置（排除当前规则自身）
		if len(toAdd) > 0 {
			configuredUuids, err := shopRepo.GetConfiguredShopUuids(headquarterUuid, r.WarehouseErpCode, r.Uuid)
			if err != nil {
				return errors.WithMessage(err, "查询已配置门店失败")
			}
			configuredMap := make(map[uint64]bool, len(configuredUuids))
			for _, uid := range configuredUuids {
				configuredMap[uid] = true
			}
			for _, shopUuid := range toAdd {
				if configuredMap[shopUuid] {
					return errors.New(fmt.Sprintf(i18n.Translate(ctx.GetLanguage(), "门店 %d 已在该仓库的其他规则中配置"), shopUuid))
				}
			}
		}

		// 执行删除
		if len(toDeleteUuids) > 0 {
			if err := shopRepo.SoftDeleteByUuids(toDeleteUuids, headquarterUuid); err != nil {
				return errors.WithMessage(err, "删除门店失败")
			}
		}

		// 执行新增
		if len(toAdd) > 0 {
			shops := make([]model.AutoReceiptRuleShop, 0, len(toAdd))
			for _, shopUuid := range toAdd {
				shops = append(shops, model.AutoReceiptRuleShop{
					RuleUuid:        r.Uuid,
					ShopCompanyUuid: shopUuid,
				})
			}
			if err := shopRepo.BatchCreate(shops); err != nil {
				return errors.WithMessage(err, "新增门店失败")
			}
		}

		return nil
	})

	if err != nil {
		logger.Logger.Error("更新自动收货规则失败",
			zap.Uint64("company_uuid", ctx.GetCompanyUuid()),
			zap.Uint64("rule_uuid", r.Uuid),
			zap.Error(err),
		)
	}
	return err
}

// DeleteRule 删除自动收货规则（级联删除门店）
func (s *autoReceiptSrv) DeleteRule(ctx context.Context, r req.DeleteAutoReceiptRuleReq) error {
	if err := s.checkHeadquarter(ctx); err != nil {
		return err
	}
	db := s.getSaasDB()
	headquarterUuid := s.getHeadquarterUuid(ctx)

	err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		ruleRepo := repository.NewAutoReceiptRuleRepo(tx)
		if err := ruleRepo.SoftDelete(r.Uuids, headquarterUuid); err != nil {
			return errors.WithMessage(err, "删除规则失败")
		}
		shopRepo := repository.NewAutoReceiptRuleShopRepo(tx)
		if err := shopRepo.SoftDeleteByRuleUuids(r.Uuids, headquarterUuid); err != nil {
			return errors.WithMessage(err, "删除规则门店失败")
		}
		return nil
	})

	if err != nil {
		logger.Logger.Error("删除自动收货规则失败",
			zap.Uint64("company_uuid", ctx.GetCompanyUuid()),
			zap.Error(err),
		)
	}
	return err
}

// GetRuleList 获取规则列表
func (s *autoReceiptSrv) GetRuleList(ctx context.Context, r req.AutoReceiptRuleListReq) (resp.AutoReceiptRuleListResp, error) {
	if err := s.checkHeadquarter(ctx); err != nil {
		return resp.AutoReceiptRuleListResp{}, err
	}
	db := s.getSaasDB()
	headquarterUuid := s.getHeadquarterUuid(ctx)

	ruleRepo := repository.NewAutoReceiptRuleRepo(db)
	rules, err := ruleRepo.GetListByHeadquarter(headquarterUuid, "")
	if err != nil {
		return resp.AutoReceiptRuleListResp{}, errors.WithMessage(err, "查询规则列表失败")
	}

	// 获取总部下所有门店（用于汇总统计）
	companyRepo := repository.NewCompanyRepo(db)
	allCompanies, err := companyRepo.GetNoDeleteListByHeadquarterUuid(headquarterUuid)
	if err != nil {
		return resp.AutoReceiptRuleListResp{}, errors.WithMessage(err, "获取门店列表失败")
	}
	// 排除总部自身
	currentCompanyUuid := ctx.GetCompanyUuid()
	allShops := make([]model.Company, 0, len(allCompanies))
	for _, c := range allCompanies {
		if c.Uuid != currentCompanyUuid {
			allShops = append(allShops, c)
		}
	}

	if len(rules) == 0 {
		unconfigured := make([]resp.UnconfiguredShopItem, 0, len(allShops))
		for _, shop := range allShops {
			unconfigured = append(unconfigured, resp.UnconfiguredShopItem{Uuid: shop.Uuid, Name: shop.Name})
		}
		return resp.AutoReceiptRuleListResp{
			List:              make([]resp.AutoReceiptRuleGroup, 0),
			UnconfiguredCount: len(allShops),
			UnconfiguredShops: unconfigured,
		}, nil
	}

	// 查仓库多语言名称（仓库在总店库）
	warehouseNameMap := s.getWarehouseNameMap(ctx.GetDB())

	// 收集规则UUID查门店
	ruleUuids := make([]uint64, 0, len(rules))
	for _, rule := range rules {
		ruleUuids = append(ruleUuids, rule.Uuid)
	}

	shopRepo := repository.NewAutoReceiptRuleShopRepo(db)
	ruleShops, err := shopRepo.GetByRuleUuids(ruleUuids)
	if err != nil {
		return resp.AutoReceiptRuleListResp{}, errors.WithMessage(err, "查询规则门店失败")
	}

	// 查门店名称
	shopCompanyUuids := make([]uint64, 0, len(ruleShops))
	for _, shop := range ruleShops {
		shopCompanyUuids = append(shopCompanyUuids, shop.ShopCompanyUuid)
	}
	shopNameMap := s.getCompanyNameMap(db, shopCompanyUuids)

	// 查门店编码
	shopCodeMap := s.getShopStoreCodeMap(ctx, shopCompanyUuids)

	// 按规则UUID分组门店
	ruleShopMap := make(map[uint64][]resp.AutoReceiptRuleShop)
	for _, shop := range ruleShops {
		ruleShopMap[shop.RuleUuid] = append(ruleShopMap[shop.RuleUuid], resp.AutoReceiptRuleShop{
			Uuid:     shop.Uuid,
			ShopUuid: shop.ShopCompanyUuid,
			ShopCode: shopCodeMap[shop.ShopCompanyUuid],
			ShopName: shopNameMap[shop.ShopCompanyUuid],
		})
	}

	// 组装响应
	list := make([]resp.AutoReceiptRuleGroup, 0, len(rules))
	for _, rule := range rules {
		shops := ruleShopMap[rule.Uuid]
		if shops == nil {
			shops = make([]resp.AutoReceiptRuleShop, 0)
		}
		list = append(list, resp.AutoReceiptRuleGroup{
			Uuid:                rule.Uuid,
			LocaleName:          *language.JsonToLocaleResponse(rule.Name),
			WarehouseErpCode:    rule.WarehouseErpCode,
			WarehouseLocaleName: warehouseNameMap[rule.WarehouseErpCode],
			DelayDays:           rule.DelayDays,
			Status:              rule.Status,
			ShopCount:           len(shops),
			Shops:               shops,
		})
	}

	// 汇总统计：已配置门店（去重）和未配置门店
	configuredSet := make(map[uint64]bool, len(ruleShops))
	for _, shop := range ruleShops {
		configuredSet[shop.ShopCompanyUuid] = true
	}
	unconfigured := make([]resp.UnconfiguredShopItem, 0)
	for _, shop := range allShops {
		if !configuredSet[shop.Uuid] {
			unconfigured = append(unconfigured, resp.UnconfiguredShopItem{Uuid: shop.Uuid, Name: shop.Name})
		}
	}

	return resp.AutoReceiptRuleListResp{
		List:              list,
		ConfiguredCount:   len(configuredSet),
		UnconfiguredCount: len(unconfigured),
		UnconfiguredShops: unconfigured,
	}, nil
}

// GetShopList 获取可选门店列表（标记已配置的门店为 disabled）
func (s *autoReceiptSrv) GetShopList(ctx context.Context, r req.AutoReceiptShopListReq) (resp.AutoReceiptShopListResp, error) {
	if err := s.checkHeadquarter(ctx); err != nil {
		return resp.AutoReceiptShopListResp{}, err
	}
	db := s.getSaasDB()
	headquarterUuid := s.getHeadquarterUuid(ctx)

	// 获取总部下所有门店
	companyRepo := repository.NewCompanyRepo(db)
	companies, err := companyRepo.GetNoDeleteListByHeadquarterUuid(headquarterUuid)
	if err != nil {
		return resp.AutoReceiptShopListResp{}, errors.WithMessage(err, "获取门店列表失败")
	}

	// 获取同仓库下已配置的门店
	shopRepo := repository.NewAutoReceiptRuleShopRepo(db)
	configuredUuids, err := shopRepo.GetConfiguredShopUuids(headquarterUuid, r.WarehouseErpCode, 0)
	if err != nil {
		return resp.AutoReceiptShopListResp{}, errors.WithMessage(err, "查询已配置门店失败")
	}
	configuredMap := make(map[uint64]bool, len(configuredUuids))
	for _, uid := range configuredUuids {
		configuredMap[uid] = true
	}

	// 获取门店编码
	companyUuids := make([]uint64, 0, len(companies))
	for _, company := range companies {
		companyUuids = append(companyUuids, company.Uuid)
	}
	shopCodeMap := s.getShopStoreCodeMap(ctx, companyUuids)

	// 组装响应（排除总部自身）
	currentCompanyUuid := ctx.GetCompanyUuid()
	list := make([]resp.AutoReceiptShopItem, 0, len(companies))
	for _, company := range companies {
		if company.Uuid == currentCompanyUuid {
			continue
		}
		item := resp.AutoReceiptShopItem{
			Uuid:      company.Uuid,
			Name:      company.Name,
			StoreCode: shopCodeMap[company.Uuid],
			Status:    company.Status,
			Disabled:  configuredMap[company.Uuid],
		}
		list = append(list, item)
	}

	utils.SortByStoreCode(list)

	return resp.AutoReceiptShopListResp{List: list}, nil
}

// GetLogList 获取自动收货记录列表
func (s *autoReceiptSrv) GetLogList(ctx context.Context, r req.AutoReceiptLogListReq) (resp.AutoReceiptLogListResp, error) {
	if err := s.checkHeadquarter(ctx); err != nil {
		return resp.AutoReceiptLogListResp{}, err
	}
	db := s.getSaasDB()
	headquarterUuid := s.getHeadquarterUuid(ctx)

	// 根据总店时区将日期时间字符串解析为 Unix 时间戳
	var startTime, endTime int64
	if r.StartTime != "" || r.EndTime != "" {
		companySetting := ctx.GetCompanySetting()
		tz := utils.SetTimezone(companySetting.GetTimezone())
		if r.StartTime != "" {
			startTime, _ = tz.FormatDateTimeToUnix(r.StartTime)
		}
		if r.EndTime != "" {
			endTime, _ = tz.FormatDateTimeToUnix(r.EndTime)
		}
	}

	logRepo := repository.NewAutoReceiptLogRepo(db)
	logs, total, err := logRepo.GetList(headquarterUuid, r.ShopCompanyUuid, startTime, endTime, r.PageNo, r.PageSize)
	if err != nil {
		return resp.AutoReceiptLogListResp{}, errors.WithMessage(err, "查询自动收货记录失败")
	}

	// 获取门店名称
	shopUuids := make([]uint64, 0, len(logs))
	for _, log := range logs {
		shopUuids = append(shopUuids, log.ShopCompanyUuid)
	}
	shopNameMap := s.getCompanyNameMap(db, shopUuids)

	// 组装响应
	items := make([]resp.AutoReceiptLogItem, 0, len(logs))
	for _, log := range logs {
		items = append(items, resp.AutoReceiptLogItem{
			Uuid:              log.Uuid,
			ShopCompanyUuid:   log.ShopCompanyUuid,
			ShopName:          shopNameMap[log.ShopCompanyUuid],
			ReceiptOrderUuid:  log.ReceiptOrderUuid,
			ReceiptOrderNo:    log.ReceiptOrderNo,
			ReceiptErpOrderNo: log.ReceiptErpOrderNo,
			ReceiptTime:       log.ReceiptTime,
		})
	}

	return resp.AutoReceiptLogListResp{
		List: items,
		Meta: dto.PageResponse{
			PageNo:   r.PageNo,
			PageSize: r.PageSize,
			Total:    total,
		},
	}, nil
}

// GetLogDetail 获取自动收货日志详情，返回门店上下文和收货单UUID，供 handler 层直接调用收货单详情接口
func (s *autoReceiptSrv) GetLogDetail(ctx context.Context, r req.AutoReceiptLogDetailReq) (context.Context, uint64, error) {
	if err := s.checkHeadquarter(ctx); err != nil {
		return nil, 0, err
	}
	db := s.getSaasDB()
	headquarterUuid := s.getHeadquarterUuid(ctx)

	// 查日志记录
	logRepo := repository.NewAutoReceiptLogRepo(db)
	log, err := logRepo.GetByUuid(r.Uuid, headquarterUuid)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, 0, errors.New("记录不存在")
		}
		return nil, 0, errors.WithMessage(err, "查询自动收货记录失败")
	}

	// 构建门店上下文（DB + Company + CompanySetting）
	companyRepo := repository.NewCompanyRepo(db)
	shopCompany, err := companyRepo.GetCompanyInfoByUuid(log.ShopCompanyUuid)
	if err != nil {
		return nil, 0, errors.WithMessage(err, "查询门店信息失败")
	}

	shopCtx := ctx.Copy()
	shopCtx.SetDB(s.dbm.GetDB(log.ShopCompanyUuid))
	shopCtx.SetCompany(*shopCompany)
	if shopCompany.CompanySetting != nil {
		shopCtx.SetCompanySetting(*shopCompany.CompanySetting)
	}

	return shopCtx, log.ReceiptOrderUuid, nil
}

// GetWarehouseList 获取发货仓库列表
func (s *autoReceiptSrv) GetWarehouseList(ctx context.Context) (resp.AutoReceiptWarehouseListResp, error) {
	if err := s.checkHeadquarter(ctx); err != nil {
		return resp.AutoReceiptWarehouseListResp{}, err
	}

	warehouseRepo := repository.NewWarehouseRepo(ctx.GetDB())
	warehouses, err := warehouseRepo.Get(warehouseRepo.WhereErpCodeNotEmpty())
	if err != nil {
		return resp.AutoReceiptWarehouseListResp{}, errors.WithMessage(err, "查询仓库列表失败")
	}

	list := make([]resp.AutoReceiptWarehouseItem, 0, len(warehouses))
	for _, w := range warehouses {
		list = append(list, resp.AutoReceiptWarehouseItem{
			ErpCode:    w.ErpCode,
			LocaleName: *language.JsonToLocaleResponse(w.Name),
			Type:       w.Type,
		})
	}

	return resp.AutoReceiptWarehouseListResp{List: list}, nil
}

// getCompanyNameMap 批量获取公司名称映射
func (s *autoReceiptSrv) getCompanyNameMap(db *gorm.DB, companyUuids []uint64) map[uint64]string {
	nameMap := make(map[uint64]string)
	if len(companyUuids) == 0 {
		return nameMap
	}
	companyRepo := repository.NewCompanyRepo(db)
	companies, err := companyRepo.GetByUuids(companyUuids)
	if err != nil {
		logger.Logger.Error("批量获取公司名称失败", zap.Error(err))
		return nameMap
	}
	for _, c := range companies {
		nameMap[c.Uuid] = c.Name
	}
	return nameMap
}

// getWarehouseNameMap 批量获取仓库多语言名称映射（按ERP编码）
func (s *autoReceiptSrv) getWarehouseNameMap(shopDB *gorm.DB) map[string]dto.LocaleResponse {
	nameMap := make(map[string]dto.LocaleResponse)
	warehouseRepo := repository.NewWarehouseRepo(shopDB)
	warehouses, err := warehouseRepo.Get(warehouseRepo.WhereErpCodeNotEmpty())
	if err != nil {
		logger.Logger.Error("批量获取仓库名称失败", zap.Error(err))
		return nameMap
	}
	for _, w := range warehouses {
		nameMap[w.ErpCode] = *language.JsonToLocaleResponse(w.Name)
	}
	return nameMap
}

// getShopStoreCodeMap 批量获取门店编码映射
// 注: 每个门店的 store_setting 存储在独立门店库中，无法跨库批量查询，故逐个获取
func (s *autoReceiptSrv) getShopStoreCodeMap(ctx context.Context, companyUuids []uint64) map[uint64]string {
	codeMap := make(map[uint64]string)
	if len(companyUuids) == 0 {
		return codeMap
	}
	settingSrv := setting.NewSrvImpl(s.dbm, s.cache)
	for _, uuid := range companyUuids {
		ctxCopy := ctx.Copy()
		ctxCopy.SetCompanyUuid(uuid)
		storeSetting, err := settingSrv.GetStoreSetting(ctxCopy)
		if err != nil {
			continue
		}
		codeMap[uuid] = storeSetting.StoreCode
	}
	return codeMap
}
