package service

import (
	stderrors "errors"
	"fmt"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/i18n"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/language"
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

// validateWarehouseEnabled 校验发货仓库是否存在且启用
func (s *autoReceiptSrv) validateWarehouseEnabled(ctx context.Context, warehouseErpCode string) error {
	warehouseRepo := repository.NewWarehouseRepo(ctx.GetDB())
	_, err := warehouseRepo.GetByErpCode(warehouseErpCode, warehouseRepo.WhereStatus(1))
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return errors.NewWithCode(constant.CodeWarehouseDisabled, i18n.Translate(ctx.GetLanguage(), "发货仓库已禁用或已删除，请重新选择"))
		}
		return errors.WithMessage(err, "查询发货仓库失败")
	}
	return nil
}

// CreateRule 创建自动收货规则
func (s *autoReceiptSrv) CreateRule(ctx context.Context, r req.CreateAutoReceiptRuleReq) error {
	if err := s.checkHeadquarter(ctx); err != nil {
		return err
	}
	db := s.getSaasDB()
	headquarterUuid := s.getHeadquarterUuid(ctx)

	if !r.LocaleName.CheckLen(100) {
		return errors.New("规则名称长度不能超过100个字符")
	}

	if err := s.validateWarehouseEnabled(ctx, r.WarehouseErpCode); err != nil {
		return err
	}

	// 事务内校验 + 创建（避免 TOCTOU 竞态）
	err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		shopRepo := repository.NewAutoReceiptRuleShopRepo(tx)

		// 校验门店是否已在同仓库其他规则中配置
		configuredMap, err := s.getConfiguredShopMap(shopRepo, headquarterUuid, r.WarehouseErpCode, 0)
		if err != nil {
			return err
		}
		if err := s.checkShopsNotConfigured(ctx, tx, configuredMap, r.ShopUuids); err != nil {
			return err
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

	if !r.LocaleName.CheckLen(100) {
		return errors.New("规则名称长度不能超过100个字符")
	}

	if err := s.validateWarehouseEnabled(ctx, r.WarehouseErpCode); err != nil {
		return err
	}

	err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		ruleRepo := repository.NewAutoReceiptRuleRepo(tx)
		if err := s.verifyRuleExists(ruleRepo, r.Uuid, headquarterUuid); err != nil {
			return err
		}

		// 全量更新规则主表字段
		vars := map[string]any{
			"name":               r.LocaleName.ToJson(),
			"warehouse_erp_code": r.WarehouseErpCode,
			"delay_days":         r.DelayDays,
			"status":             *r.Status,
		}
		if err := ruleRepo.Update(r.Uuid, headquarterUuid, vars); err != nil {
			return errors.WithMessage(err, "更新规则失败")
		}

		shopRepo := repository.NewAutoReceiptRuleShopRepo(tx)
		return s.syncRuleShops(ctx, tx, shopRepo, r.Uuid, headquarterUuid, r.WarehouseErpCode, r.ShopUuids)
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
		unconfigured := buildUnconfiguredList(allShops, nil)
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

	// 查门店名称和编码
	ruleShopMap := s.buildRuleShopMap(ctx, db, ruleShops)

	// 组装响应
	list := buildRuleGroupList(rules, ruleShopMap, warehouseNameMap)

	// 汇总统计
	configuredSet := buildConfiguredSet(ruleShops)
	unconfigured := buildUnconfiguredList(allShops, configuredSet)

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
	configuredMap, err := s.getConfiguredShopMap(shopRepo, headquarterUuid, r.WarehouseErpCode, 0)
	if err != nil {
		return resp.AutoReceiptShopListResp{}, err
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
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
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
	warehouses, err := warehouseRepo.Get(warehouseRepo.WhereErpCodeNotEmpty(), warehouseRepo.WhereStatus(1))
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

// buildRuleShopMap 按规则UUID分组门店并填充名称/编码
func (s *autoReceiptSrv) buildRuleShopMap(ctx context.Context, db *gorm.DB, ruleShops []model.AutoReceiptRuleShop) map[uint64][]resp.AutoReceiptRuleShop {
	shopCompanyUuids := make([]uint64, 0, len(ruleShops))
	for _, shop := range ruleShops {
		shopCompanyUuids = append(shopCompanyUuids, shop.ShopCompanyUuid)
	}
	shopNameMap := s.getCompanyNameMap(db, shopCompanyUuids)
	shopCodeMap := s.getShopStoreCodeMap(ctx, shopCompanyUuids)

	ruleShopMap := make(map[uint64][]resp.AutoReceiptRuleShop)
	for _, shop := range ruleShops {
		ruleShopMap[shop.RuleUuid] = append(ruleShopMap[shop.RuleUuid], resp.AutoReceiptRuleShop{
			Uuid:     shop.Uuid,
			ShopUuid: shop.ShopCompanyUuid,
			ShopCode: shopCodeMap[shop.ShopCompanyUuid],
			ShopName: shopNameMap[shop.ShopCompanyUuid],
		})
	}
	return ruleShopMap
}

// buildRuleGroupList 构建规则列表响应
func buildRuleGroupList(rules []model.AutoReceiptRule, ruleShopMap map[uint64][]resp.AutoReceiptRuleShop, warehouseNameMap map[string]dto.LocaleResponse) []resp.AutoReceiptRuleGroup {
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
			Status:              derefIntOr(rule.Status, 1),
			ShopCount:           len(shops),
			Shops:               shops,
		})
	}
	return list
}

// buildConfiguredSet 从规则门店列表构建已配置门店集合
func buildConfiguredSet(ruleShops []model.AutoReceiptRuleShop) map[uint64]bool {
	configuredSet := make(map[uint64]bool, len(ruleShops))
	for _, shop := range ruleShops {
		configuredSet[shop.ShopCompanyUuid] = true
	}
	return configuredSet
}

// buildUnconfiguredList 计算未配置规则的门店列表
func buildUnconfiguredList(allShops []model.Company, configuredSet map[uint64]bool) []resp.UnconfiguredShopItem {
	unconfigured := make([]resp.UnconfiguredShopItem, 0)
	for _, shop := range allShops {
		if !configuredSet[shop.Uuid] {
			unconfigured = append(unconfigured, resp.UnconfiguredShopItem{Uuid: shop.Uuid, Name: shop.Name})
		}
	}
	return unconfigured
}

// verifyRuleExists 验证规则是否存在
func (s *autoReceiptSrv) verifyRuleExists(ruleRepo repository.IAutoReceiptRuleRepo, uuid uint64, headquarterUuid uint64) error {
	_, err := ruleRepo.GetByUuid(uuid, headquarterUuid)
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("规则不存在")
		}
		return errors.WithMessage(err, "查询规则失败")
	}
	return nil
}

// syncRuleShops 全量同步规则门店（diff 增删）
func (s *autoReceiptSrv) syncRuleShops(ctx context.Context, db *gorm.DB, shopRepo repository.IAutoReceiptRuleShopRepo, ruleUuid uint64, headquarterUuid uint64, warehouseErpCode string, newShopUuids []uint64) error {
	currentShops, err := shopRepo.GetByRuleUuids([]uint64{ruleUuid})
	if err != nil {
		return errors.WithMessage(err, "查询当前门店失败")
	}

	toDelete, toAdd := s.diffShops(currentShops, newShopUuids)

	// 校验新增门店是否已在其他规则中配置
	if len(toAdd) > 0 {
		configuredMap, err := s.getConfiguredShopMap(shopRepo, headquarterUuid, warehouseErpCode, ruleUuid)
		if err != nil {
			return err
		}
		if err := s.checkShopsNotConfigured(ctx, db, configuredMap, toAdd); err != nil {
			return err
		}
	}

	// 执行删除
	if len(toDelete) > 0 {
		if err := shopRepo.SoftDeleteByUuids(toDelete, headquarterUuid); err != nil {
			return errors.WithMessage(err, "删除门店失败")
		}
	}

	// 执行新增
	if len(toAdd) > 0 {
		shops := make([]model.AutoReceiptRuleShop, 0, len(toAdd))
		for _, shopUuid := range toAdd {
			shops = append(shops, model.AutoReceiptRuleShop{
				RuleUuid:        ruleUuid,
				ShopCompanyUuid: shopUuid,
			})
		}
		if err := shopRepo.BatchCreate(shops); err != nil {
			return errors.WithMessage(err, "新增门店失败")
		}
	}
	return nil
}

// diffShops 计算门店变更（返回需删除的子表UUID列表和需新增的门店UUID列表）
func (s *autoReceiptSrv) diffShops(currentShops []model.AutoReceiptRuleShop, newShopUuids []uint64) (toDeleteUuids []uint64, toAddUuids []uint64) {
	currentMap := make(map[uint64]uint64, len(currentShops))
	for _, shop := range currentShops {
		currentMap[shop.ShopCompanyUuid] = shop.Uuid
	}
	newSet := make(map[uint64]bool, len(newShopUuids))
	for _, uid := range newShopUuids {
		newSet[uid] = true
	}
	for shopUuid, subUuid := range currentMap {
		if !newSet[shopUuid] {
			toDeleteUuids = append(toDeleteUuids, subUuid)
		}
	}
	for _, shopUuid := range newShopUuids {
		if _, exists := currentMap[shopUuid]; !exists {
			toAddUuids = append(toAddUuids, shopUuid)
		}
	}
	return
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
		logger.Logger.Error("批量获取公司名称失败", zap.Uint64s("company_uuids", companyUuids), zap.Error(err))
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

// getConfiguredShopMap 查询同仓库下已配置的门店UUID集合
func (s *autoReceiptSrv) getConfiguredShopMap(shopRepo repository.IAutoReceiptRuleShopRepo, headquarterUuid uint64, warehouseErpCode string, excludeRuleUuid uint64) (map[uint64]bool, error) {
	configuredUuids, err := shopRepo.GetConfiguredShopUuids(headquarterUuid, warehouseErpCode, excludeRuleUuid)
	if err != nil {
		return nil, errors.WithMessage(err, "查询已配置门店失败")
	}
	configuredMap := make(map[uint64]bool, len(configuredUuids))
	for _, uid := range configuredUuids {
		configuredMap[uid] = true
	}
	return configuredMap, nil
}

// checkShopsNotConfigured 校验门店是否已在同仓库其他规则中配置
func (s *autoReceiptSrv) checkShopsNotConfigured(ctx context.Context, db *gorm.DB, configuredMap map[uint64]bool, shopUuids []uint64) error {
	for _, shopUuid := range shopUuids {
		if configuredMap[shopUuid] {
			nameMap := s.getCompanyNameMap(db, []uint64{shopUuid})
			name := fmt.Sprintf("%d", shopUuid)
			if n, ok := nameMap[shopUuid]; ok {
				name = n
			}
			msg := fmt.Sprintf(i18n.Translate(ctx.GetLanguage(), "门店 %s 已在该仓库的其他规则中配置"), name)
			return errors.New(msg)
		}
	}
	return nil
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

// derefIntOr 安全解引用 *int 指针，nil 时返回默认值
func derefIntOr(p *int, def int) int {
	if p != nil {
		return *p
	}
	return def
}
