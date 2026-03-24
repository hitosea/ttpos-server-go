package service

import (
	"fmt"

	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// IHqPushSrv 总部推送服务接口
type IHqPushSrv interface {
	// API 方法
	GetControlSetting(ctx context.Context) (resp.HqControlSettingResp, error)
	UpdateControlSetting(ctx context.Context, updateReq req.HqControlSettingUpdateReq) error
	GetBatchPushStoreList(ctx context.Context) (resp.HqBatchPushStoreListResp, error)
	BatchPush(ctx context.Context, pushReq req.HqBatchPushReq) (resp.HqBatchPushResp, error)

	// 内部方法：HQ 修改后触发自动推送（通过 utils.Go 异步调用）
	OnHqProductChanged(ctx context.Context, productUuid uint64)
	OnHqProductTakeoutChanged(ctx context.Context, takeoutUuid uint64)
	OnHqMaterialChanged(ctx context.Context, materialUuid uint64)

	// 内部方法：子店修改时标记 override
	MarkFieldOverridden(ctx context.Context, entityType string, entityUuid uint64, fieldType string) error

	// 内部方法：检查字段是否可编辑（子店查询用，hqUuid 为总部 UUID）
	IsFieldEditable(hqUuid uint64, fieldType string) bool
}

// hqPushSrv 总部推送服务实现
type hqPushSrv struct {
	dbm *database.DBManager
}

// NewHqPushSrv 创建总部推送服务
func NewHqPushSrv(dbm *database.DBManager) IHqPushSrv {
	return &hqPushSrv{dbm: dbm}
}

// GetControlSetting 获取总部控制设置
func (s *hqPushSrv) GetControlSetting(ctx context.Context) (resp.HqControlSettingResp, error) {
	companySetting := ctx.GetCompanySetting()
	if !companySetting.IsHeadquarter() {
		return resp.HqControlSettingResp{}, errors.New("仅总部可查看控制设置")
	}

	controlRepo := s.getHqControlRepo(ctx.GetCompanyUuid())
	controlMap := controlRepo.GetControlMap(ctx.GetCompanyUuid())

	return resp.HqControlSettingResp{
		HqControlDineShelf:     getControlModeWithDefault(controlMap, constant.HqFieldDineShelf),
		HqControlTakeoutShelf:  getControlModeWithDefault(controlMap, constant.HqFieldTakeoutShelf),
		HqControlTakeoutPrice:  getControlModeWithDefault(controlMap, constant.HqFieldTakeoutPrice),
		HqControlNegativeStock: getControlModeWithDefault(controlMap, constant.HqFieldNegativeStock),
	}, nil
}

// UpdateControlSetting 更新总部控制设置
func (s *hqPushSrv) UpdateControlSetting(ctx context.Context, updateReq req.HqControlSettingUpdateReq) error {
	companySetting := ctx.GetCompanySetting()
	if !companySetting.IsHeadquarter() {
		return errors.New("仅总部可修改控制设置")
	}

	companyUuid := ctx.GetCompanyUuid()
	controlRepo := s.getHqControlRepo(companyUuid)
	currentMap := controlRepo.GetControlMap(companyUuid)

	// 检测从"分开控制"切换为"统一控制"的字段，需要触发强制推送
	forcePushFields := make([]string, 0)
	hasChange := false

	type fieldUpdate struct {
		fieldType   string
		controlMode int
	}
	updates := make([]fieldUpdate, 0, 4)

	if updateReq.HqControlDineShelf != nil {
		updates = append(updates, fieldUpdate{constant.HqFieldDineShelf, *updateReq.HqControlDineShelf})
		if getControlModeWithDefault(currentMap, constant.HqFieldDineShelf) == constant.HqControlSeparate && *updateReq.HqControlDineShelf == constant.HqControlUnified {
			forcePushFields = append(forcePushFields, constant.HqFieldDineShelf)
		}
	}
	if updateReq.HqControlTakeoutShelf != nil {
		updates = append(updates, fieldUpdate{constant.HqFieldTakeoutShelf, *updateReq.HqControlTakeoutShelf})
		if getControlModeWithDefault(currentMap, constant.HqFieldTakeoutShelf) == constant.HqControlSeparate && *updateReq.HqControlTakeoutShelf == constant.HqControlUnified {
			forcePushFields = append(forcePushFields, constant.HqFieldTakeoutShelf)
		}
	}
	if updateReq.HqControlTakeoutPrice != nil {
		updates = append(updates, fieldUpdate{constant.HqFieldTakeoutPrice, *updateReq.HqControlTakeoutPrice})
		if getControlModeWithDefault(currentMap, constant.HqFieldTakeoutPrice) == constant.HqControlSeparate && *updateReq.HqControlTakeoutPrice == constant.HqControlUnified {
			forcePushFields = append(forcePushFields, constant.HqFieldTakeoutPrice)
		}
	}
	if updateReq.HqControlNegativeStock != nil {
		updates = append(updates, fieldUpdate{constant.HqFieldNegativeStock, *updateReq.HqControlNegativeStock})
		if getControlModeWithDefault(currentMap, constant.HqFieldNegativeStock) == constant.HqControlSeparate && *updateReq.HqControlNegativeStock == constant.HqControlUnified {
			forcePushFields = append(forcePushFields, constant.HqFieldNegativeStock)
		}
	}

	// 写入 KV 表（总部 DB）
	for _, u := range updates {
		if err := controlRepo.Upsert(u.fieldType, u.controlMode); err != nil {
			return errors.WithMessage(err, "更新控制设置失败")
		}
		hasChange = true
	}

	if !hasChange {
		return nil
	}

	// 失效缓存并立即重建（子店读取时 db=nil 无法回填）
	controlRepo.InvalidateCache(companyUuid)
	controlRepo.GetControlMap(companyUuid)

	// 分开→统一：触发强制推送到所有子店
	if len(forcePushFields) > 0 {
		utils.Go(func() {
			s.forcePushToAllSubStores(companyUuid, forcePushFields)
		})
	}

	return nil
}

// GetBatchPushStoreList 获取可推送门店列表
func (s *hqPushSrv) GetBatchPushStoreList(ctx context.Context) (resp.HqBatchPushStoreListResp, error) {
	companySetting := ctx.GetCompanySetting()
	if !companySetting.IsHeadquarter() {
		return resp.HqBatchPushStoreListResp{}, errors.New("仅总部可查看门店列表")
	}

	hqUuid := ctx.GetCompanyUuid()
	saasDB := s.dbm.GetDB(constant.DefaultDB)
	companies, err := repository.NewCompanyRepo(saasDB).GetNoDeleteListByHeadquarterUuid(hqUuid)
	if err != nil {
		return resp.HqBatchPushStoreListResp{}, errors.WithMessage(err, "获取门店列表失败")
	}

	list := make([]resp.HqBatchPushStoreItem, 0, len(companies))
	for _, company := range companies {
		// 排除总部自身
		if company.Uuid == hqUuid {
			continue
		}
		list = append(list, resp.HqBatchPushStoreItem{
			CompanyUuid: company.Uuid,
			Name:        company.Name,
			Status:      company.Status,
		})
	}

	return resp.HqBatchPushStoreListResp{List: list}, nil
}

// BatchPush 批量推送给子店
func (s *hqPushSrv) BatchPush(ctx context.Context, pushReq req.HqBatchPushReq) (resp.HqBatchPushResp, error) {
	companySetting := ctx.GetCompanySetting()
	if !companySetting.IsHeadquarter() {
		return resp.HqBatchPushResp{}, errors.New("仅总部可执行批量推送")
	}

	// 验证字段类型
	for _, ft := range pushReq.FieldTypes {
		if !constant.IsValidBatchPushFieldType(ft) {
			return resp.HqBatchPushResp{}, errors.New("无效的推送类型: " + ft)
		}
	}

	hqUuid := ctx.GetCompanyUuid()

	// 获取目标门店
	targetStoreUuids, err := s.resolveTargetStores(hqUuid, pushReq.IsAllStores, pushReq.StoreUuids)
	if err != nil {
		return resp.HqBatchPushResp{}, err
	}
	if len(targetStoreUuids) == 0 {
		return resp.HqBatchPushResp{}, errors.New("未选择任何门店")
	}

	// 异步执行推送（isForce=true 强制覆盖，忽略 override）
	utils.Go(func() {
		for _, fieldType := range pushReq.FieldTypes {
			s.pushFieldToStores(hqUuid, targetStoreUuids, fieldType, true)
		}
	})

	return resp.HqBatchPushResp{Message: "推送任务已启动"}, nil
}

// OnHqProductChanged HQ 商品变更后触发推送（所有字段）
func (s *hqPushSrv) OnHqProductChanged(ctx context.Context, productUuid uint64) {
	companySetting := ctx.GetCompanySetting()
	if !companySetting.IsHeadquarter() {
		return
	}

	hqUuid := ctx.GetCompanyUuid()
	targetStoreUuids, err := s.getSubStoreUuids(hqUuid)
	if err != nil || len(targetStoreUuids) == 0 {
		return
	}

	s.pushProductToStores(hqUuid, targetStoreUuids, productUuid)
}

// OnHqProductTakeoutChanged HQ 外卖商品变更后触发推送（所有字段）
func (s *hqPushSrv) OnHqProductTakeoutChanged(ctx context.Context, takeoutUuid uint64) {
	companySetting := ctx.GetCompanySetting()
	if !companySetting.IsHeadquarter() {
		return
	}

	hqUuid := ctx.GetCompanyUuid()
	targetStoreUuids, err := s.getSubStoreUuids(hqUuid)
	if err != nil || len(targetStoreUuids) == 0 {
		return
	}

	s.pushProductTakeoutToStores(hqUuid, targetStoreUuids, takeoutUuid)
}

// OnHqMaterialChanged HQ 物品变更后触发推送（所有字段）
func (s *hqPushSrv) OnHqMaterialChanged(ctx context.Context, materialUuid uint64) {
	companySetting := ctx.GetCompanySetting()
	if !companySetting.IsHeadquarter() {
		return
	}

	hqUuid := ctx.GetCompanyUuid()
	targetStoreUuids, err := s.getSubStoreUuids(hqUuid)
	if err != nil || len(targetStoreUuids) == 0 {
		return
	}

	s.pushMaterialToStores(hqUuid, targetStoreUuids, materialUuid)
}

// MarkFieldOverridden 子店修改时标记字段已覆盖
func (s *hqPushSrv) MarkFieldOverridden(ctx context.Context, entityType string, entityUuid uint64, fieldType string) error {
	db := ctx.GetDB()
	return repository.NewHqFieldOverrideRepo(db).MarkOverridden(entityUuid, entityType, fieldType)
}

// IsFieldEditable 检查子店是否可编辑指定字段
func (s *hqPushSrv) IsFieldEditable(hqUuid uint64, fieldType string) bool {
	// 安全库存始终可编辑（无控制模式）
	if fieldType == constant.HqFieldSafetyStock {
		return true
	}
	controlRepo := s.getHqControlRepo(hqUuid)
	return !controlRepo.IsUnifiedControl(hqUuid, fieldType)
}

// ========== 内部方法 ==========

// getSubStoreUuids 获取所有子店的 UUID
func (s *hqPushSrv) getSubStoreUuids(hqUuid uint64) ([]uint64, error) {
	saasDB := s.dbm.GetDB(constant.DefaultDB)
	companies, err := repository.NewCompanyRepo(saasDB).GetNoDeleteListByHeadquarterUuid(hqUuid)
	if err != nil {
		logger.Logger.Error("获取子店列表失败", zap.Uint64("company_uuid", hqUuid), zap.Error(err))
		return nil, err
	}

	uuids := make([]uint64, 0, len(companies))
	for _, c := range companies {
		if c.Uuid != hqUuid {
			uuids = append(uuids, c.Uuid)
		}
	}
	return uuids, nil
}

// resolveTargetStores 解析目标门店
func (s *hqPushSrv) resolveTargetStores(hqUuid uint64, isAllStores bool, storeUuids []uint64) ([]uint64, error) {
	if isAllStores {
		return s.getSubStoreUuids(hqUuid)
	}
	// 过滤掉 HQ 自身
	filtered := make([]uint64, 0, len(storeUuids))
	for _, uuid := range storeUuids {
		if uuid != hqUuid {
			filtered = append(filtered, uuid)
		}
	}
	return filtered, nil
}

// getHqControlRepo 获取 HQ 控制设置 repo（指向总部 DB）
func (s *hqPushSrv) getHqControlRepo(hqUuid uint64) repository.IHqControlSettingRepo {
	hqDB := s.dbm.GetDB(hqUuid)
	return repository.NewHqControlSettingRepo(hqDB)
}

// getControlModeWithDefault 从 controlMap 获取控制模式，不存在时返回默认值
func getControlModeWithDefault(controlMap map[string]int, fieldType string) int {
	if mode, ok := controlMap[fieldType]; ok {
		return mode
	}
	return defaultControlMode(fieldType)
}

// defaultControlMode 返回字段类型的默认控制模式（与 repository 层保持一致）
func defaultControlMode(fieldType string) int {
	switch fieldType {
	case constant.HqFieldNegativeStock:
		return constant.HqControlUnified // 负库存默认统一控制
	default:
		return constant.HqControlSeparate // 其他默认分开控制
	}
}

// forcePushToAllSubStores 强制推送到所有子店（控制模式切换时触发）
func (s *hqPushSrv) forcePushToAllSubStores(hqUuid uint64, fieldTypes []string) {
	storeUuids, err := s.getSubStoreUuids(hqUuid)
	if err != nil || len(storeUuids) == 0 {
		return
	}

	for _, ft := range fieldTypes {
		fieldType := ft
		s.pushFieldToStores(hqUuid, storeUuids, fieldType, true)
	}
}

// pushFieldToStores 推送指定字段类型到多个子店（并行，最大并发 10）
func (s *hqPushSrv) pushFieldToStores(hqUuid uint64, storeUuids []uint64, fieldType string, isForce bool) {
	logger.Logger.Info("开始批量推送字段到子店",
		zap.Uint64("company_uuid", hqUuid),
		zap.String("field_type", fieldType),
		zap.Int("store_count", len(storeUuids)),
		zap.Bool("is_force", isForce),
	)
	sem := make(chan struct{}, 10)
	for _, storeUuid := range storeUuids {
		su := storeUuid
		utils.Go(func() {
			sem <- struct{}{}
			defer func() { <-sem }()
			s.pushFieldToSingleStore(hqUuid, su, fieldType, isForce)
		})
	}
}

// pushFieldToSingleStore 推送指定字段类型到单个子店
func (s *hqPushSrv) pushFieldToSingleStore(hqUuid, storeUuid uint64, fieldType string, isForce bool) {
	controlRepo := s.getHqControlRepo(hqUuid)

	// 非强制模式下，分开控制的字段由 override 逻辑处理
	isUnified := controlRepo.IsUnifiedControl(hqUuid, fieldType)

	switch fieldType {
	case constant.HqFieldDineShelf:
		s.pushDineShelfToStore(hqUuid, storeUuid, isForce || isUnified)
	case constant.HqFieldTakeoutShelf:
		s.pushTakeoutShelfToStore(hqUuid, storeUuid, isForce || isUnified)
	case constant.HqFieldTakeoutPrice:
		s.pushTakeoutPriceToStore(hqUuid, storeUuid, isForce || isUnified)
	case constant.HqFieldNegativeStock:
		s.pushNegativeStockToStore(hqUuid, storeUuid, isForce || isUnified)
	default:
		logger.Logger.Warn("未知的推送字段类型",
			zap.Uint64("company_uuid", hqUuid),
			zap.Uint64("store_uuid", storeUuid),
			zap.String("field_type", fieldType),
		)
	}
}

// pushDineShelfToStore 推送店内上下架到单个子店
func (s *hqPushSrv) pushDineShelfToStore(hqUuid, storeUuid uint64, forceOverwrite bool) {
	hqDB := s.dbm.GetDB(hqUuid)
	storeDB := s.dbm.GetDB(storeUuid)
	commonRepo := repository.NewCommonRepo()

	// 获取 HQ 的商品列表（仅总部原生商品）
	hqProducts, err := repository.NewProductPackageRepo(hqDB).GetProductPackageList(
		commonRepo.WhereByHeadquarterUuid(0),
	)
	if err != nil {
		logger.Logger.Error("获取总部商品列表失败", zap.Uint64("company_uuid", hqUuid), zap.Error(err))
		return
	}
	if len(hqProducts) == 0 {
		return
	}

	// 获取子店的 override 状态
	overrideRepo := repository.NewHqFieldOverrideRepo(storeDB)
	productUuids := make([]uint64, 0, len(hqProducts))
	hqStatusMap := make(map[uint64]uint)
	for _, p := range hqProducts {
		productUuids = append(productUuids, p.Uuid)
		hqStatusMap[p.Uuid] = p.Status
	}

	overriddenMap := make(map[uint64]bool)
	if !forceOverwrite {
		overriddenMap, _ = overrideRepo.BatchCheckOverridden(productUuids, constant.HqFieldDineShelf)
	}

	// 获取子店商品 UUID 集合（用于存在性检查）
	storeProductRepo := repository.NewProductPackageRepo(storeDB)
	storeProducts, _ := storeProductRepo.GetProductPackageList(
		commonRepo.WhereByHeadquarterUuid(hqUuid),
	)
	storeProductSet := make(map[uint64]bool, len(storeProducts))
	for _, sp := range storeProducts {
		storeProductSet[sp.Uuid] = true
	}

	// 逐个商品更新（每个商品一个事务，互不影响）
	for _, productUuid := range productUuids {
		hqStatus := hqStatusMap[productUuid]

		if !storeProductSet[productUuid] {
			continue
		}
		if !forceOverwrite && overriddenMap[productUuid] {
			continue
		}

		if err := repository.CommonRepo.Transaction(storeDB, func(tx *gorm.DB) error {
			txProductRepo := repository.NewProductPackageRepo(tx)
			txBomRepo := repository.NewProductBomRepo(tx)

			if err := txProductRepo.UpdateProductPackage(
				map[string]any{"status": hqStatus},
				commonRepo.WhereByUuid(productUuid),
			); err != nil {
				return err
			}
			if err := txBomRepo.UpdateProductBom(
				map[string]any{"status": hqStatus},
				commonRepo.WhereByProductPackageUuid(productUuid),
			); err != nil {
				return err
			}
			if forceOverwrite {
				txOverrideRepo := repository.NewHqFieldOverrideRepo(tx)
				txOverrideRepo.ClearOverride(productUuid, constant.HqFieldDineShelf)
			}
			return nil
		}); err != nil {
			logger.Logger.Error("推送店内上下架失败",
				zap.Uint64("company_uuid", hqUuid),
				zap.Uint64("store_uuid", storeUuid),
				zap.Uint64("product_uuid", productUuid),
				zap.Error(err),
			)
		}
	}
}

// pushTakeoutShelfToStore 推送外卖上下架到单个子店
// 注意：子店外卖商品 UUID 与 HQ 不同，需通过 product_package_uuid + takeout_type 匹配
func (s *hqPushSrv) pushTakeoutShelfToStore(hqUuid, storeUuid uint64, forceOverwrite bool) {
	hqDB := s.dbm.GetDB(hqUuid)
	storeDB := s.dbm.GetDB(storeUuid)
	commonRepo := repository.NewCommonRepo()

	// 获取 HQ 外卖商品，按 product_package_uuid + takeout_type 建索引
	hqTakeouts, err := repository.NewProductPackageTakeoutRepo(hqDB).GetProductPackageTakeoutList(
		commonRepo.WhereByHeadquarterUuid(0),
	)
	if err != nil || len(hqTakeouts) == 0 {
		return
	}
	type takeoutKey struct {
		ProductPackageUuid uint64
		TakeoutType        uint
	}
	hqMap := make(map[takeoutKey]uint, len(hqTakeouts))
	for _, t := range hqTakeouts {
		hqMap[takeoutKey{t.ProductPackageUuid, t.TakeoutType}] = t.Status
	}

	// 获取子店外卖商品（UUID 与 HQ 不同）
	storeTakeoutRepo := repository.NewProductPackageTakeoutRepo(storeDB)
	storeTakeouts, _ := storeTakeoutRepo.GetProductPackageTakeoutList(
		commonRepo.WhereByHeadquarterUuid(hqUuid),
	)
	if len(storeTakeouts) == 0 {
		return
	}

	// 获取子店 override 状态（使用子店 UUID）
	overrideRepo := repository.NewHqFieldOverrideRepo(storeDB)
	storeUuids := make([]uint64, 0, len(storeTakeouts))
	for _, st := range storeTakeouts {
		storeUuids = append(storeUuids, st.Uuid)
	}
	overriddenMap := make(map[uint64]bool)
	if !forceOverwrite {
		overriddenMap, _ = overrideRepo.BatchCheckOverridden(storeUuids, constant.HqFieldTakeoutShelf)
	}

	for _, storeTakeout := range storeTakeouts {
		hqStatus, ok := hqMap[takeoutKey{storeTakeout.ProductPackageUuid, storeTakeout.TakeoutType}]
		if !ok {
			continue
		}
		if !forceOverwrite && overriddenMap[storeTakeout.Uuid] {
			continue
		}

		if err := repository.CommonRepo.Transaction(storeDB, func(tx *gorm.DB) error {
			txTakeoutRepo := repository.NewProductPackageTakeoutRepo(tx)
			if err := txTakeoutRepo.UpdateProductPackageTakeout(
				map[string]any{"status": hqStatus},
				commonRepo.WhereByUuid(storeTakeout.Uuid),
			); err != nil {
				return err
			}
			if forceOverwrite {
				txOverrideRepo := repository.NewHqFieldOverrideRepo(tx)
				txOverrideRepo.ClearOverride(storeTakeout.Uuid, constant.HqFieldTakeoutShelf)
			}
			return nil
		}); err != nil {
			logger.Logger.Error("推送外卖上下架失败",
				zap.Uint64("company_uuid", hqUuid),
				zap.Uint64("store_uuid", storeUuid),
				zap.Uint64("takeout_uuid", storeTakeout.Uuid),
				zap.Error(err),
			)
		}
	}
}

// pushTakeoutPriceToStore 推送外卖价格到单个子店
// 注意：子店外卖商品/规格 UUID 与 HQ 不同，需通过 product_package_uuid + takeout_type 匹配
func (s *hqPushSrv) pushTakeoutPriceToStore(hqUuid, storeUuid uint64, forceOverwrite bool) {
	hqDB := s.dbm.GetDB(hqUuid)
	storeDB := s.dbm.GetDB(storeUuid)
	commonRepo := repository.NewCommonRepo()

	// 获取 HQ 外卖商品，按 product_package_uuid + takeout_type 建索引
	hqTakeouts, err := repository.NewProductPackageTakeoutRepo(hqDB).GetProductPackageTakeoutList(
		commonRepo.WhereByHeadquarterUuid(0),
	)
	if err != nil || len(hqTakeouts) == 0 {
		return
	}
	type takeoutKey struct {
		ProductPackageUuid uint64
		TakeoutType        uint
	}
	hqTakeoutMap := make(map[takeoutKey]*model.ProductPackageTakeout, len(hqTakeouts))
	hqTakeoutUuids := make([]uint64, 0, len(hqTakeouts))
	for i := range hqTakeouts {
		hqTakeoutMap[takeoutKey{hqTakeouts[i].ProductPackageUuid, hqTakeouts[i].TakeoutType}] = hqTakeouts[i]
		hqTakeoutUuids = append(hqTakeoutUuids, hqTakeouts[i].Uuid)
	}

	// 获取 HQ 外卖规格价格，按 product_bom_uuid 建索引（product_bom_uuid 在 HQ 和子店间共享）
	hqBomTakeouts, _ := repository.NewProductBomTakeoutRepo(hqDB).GetProductBomTakeoutList(
		commonRepo.WhereByHeadquarterUuid(0),
	)
	// HQ takeout UUID → product_bom_uuid → price
	hqBomPriceMap := make(map[uint64]map[uint64]float64)
	for _, bt := range hqBomTakeouts {
		if hqBomPriceMap[bt.ProductPackageTakeoutUuid] == nil {
			hqBomPriceMap[bt.ProductPackageTakeoutUuid] = make(map[uint64]float64)
		}
		hqBomPriceMap[bt.ProductPackageTakeoutUuid][bt.ProductBomUuid] = bt.Price
	}

	// 获取子店外卖商品
	storeTakeoutRepo := repository.NewProductPackageTakeoutRepo(storeDB)
	storeTakeouts, _ := storeTakeoutRepo.GetProductPackageTakeoutList(
		commonRepo.WhereByHeadquarterUuid(hqUuid),
	)
	if len(storeTakeouts) == 0 {
		return
	}

	// 获取子店 override 状态（使用子店 UUID）
	overrideRepo := repository.NewHqFieldOverrideRepo(storeDB)
	storeUuids := make([]uint64, 0, len(storeTakeouts))
	for _, st := range storeTakeouts {
		storeUuids = append(storeUuids, st.Uuid)
	}
	overriddenMap := make(map[uint64]bool)
	if !forceOverwrite {
		overriddenMap, _ = overrideRepo.BatchCheckOverridden(storeUuids, constant.HqFieldTakeoutPrice)
	}

	// 获取子店外卖规格
	storeBomTakeoutRepo := repository.NewProductBomTakeoutRepo(storeDB)
	storeBomTakeouts, _ := storeBomTakeoutRepo.GetProductBomTakeoutList(
		commonRepo.WhereByHeadquarterUuid(hqUuid),
	)
	// 子店 takeout UUID → []*ProductBomTakeout
	storeBomByTakeout := make(map[uint64][]*model.ProductBomTakeout)
	for _, bt := range storeBomTakeouts {
		storeBomByTakeout[bt.ProductPackageTakeoutUuid] = append(storeBomByTakeout[bt.ProductPackageTakeoutUuid], bt)
	}

	for _, storeTakeout := range storeTakeouts {
		key := takeoutKey{storeTakeout.ProductPackageUuid, storeTakeout.TakeoutType}
		hqTakeout, ok := hqTakeoutMap[key]
		if !ok {
			continue
		}
		if !forceOverwrite && overriddenMap[storeTakeout.Uuid] {
			continue
		}

		if err := repository.CommonRepo.Transaction(storeDB, func(tx *gorm.DB) error {
			txTakeoutRepo := repository.NewProductPackageTakeoutRepo(tx)
			if err := txTakeoutRepo.UpdateProductPackageTakeout(
				map[string]any{"price": hqTakeout.Price},
				commonRepo.WhereByUuid(storeTakeout.Uuid),
			); err != nil {
				return err
			}
			if forceOverwrite {
				txOverrideRepo := repository.NewHqFieldOverrideRepo(tx)
				txOverrideRepo.ClearOverride(storeTakeout.Uuid, constant.HqFieldTakeoutPrice)
			}
			// 同步规格价格
			txBomTakeoutRepo := repository.NewProductBomTakeoutRepo(tx)
			hqBoms := hqBomPriceMap[hqTakeout.Uuid]
			for _, storeBom := range storeBomByTakeout[storeTakeout.Uuid] {
				if hqPrice, exists := hqBoms[storeBom.ProductBomUuid]; exists {
					if err := txBomTakeoutRepo.UpdateProductBomTakeout(
						map[string]any{"price": hqPrice},
						commonRepo.WhereByUuid(storeBom.Uuid),
					); err != nil {
						return err
					}
				}
			}
			return nil
		}); err != nil {
			logger.Logger.Error("推送外卖价格失败",
				zap.Uint64("company_uuid", hqUuid),
				zap.Uint64("store_uuid", storeUuid),
				zap.Uint64("takeout_uuid", storeTakeout.Uuid),
				zap.Error(err),
			)
		}
	}
}

// pushNegativeStockToStore 推送负库存和安全库存到单个子店
func (s *hqPushSrv) pushNegativeStockToStore(hqUuid, storeUuid uint64, forceOverwrite bool) {
	hqDB := s.dbm.GetDB(hqUuid)
	storeDB := s.dbm.GetDB(storeUuid)
	commonRepo := repository.NewCommonRepo()

	// 获取 HQ 物品列表
	hqMaterialRepo := repository.NewMaterialRepo(hqDB)
	hqMaterials := hqMaterialRepo.GetMaterialList(
		commonRepo.WhereByHeadquarterUuid(0),
	)
	if len(hqMaterials) == 0 {
		return
	}

	overrideRepo := repository.NewHqFieldOverrideRepo(storeDB)
	uuids := make([]uint64, 0, len(hqMaterials))
	hqNegStockMap := make(map[uint64]int)
	hqSafetyStockMap := make(map[uint64]*float64)
	for _, m := range hqMaterials {
		uuids = append(uuids, m.Uuid)
		hqNegStockMap[m.Uuid] = m.AllowNegativeStock
		hqSafetyStockMap[m.Uuid] = m.SafetyStock
	}

	negOverriddenMap := make(map[uint64]bool)
	if !forceOverwrite {
		negOverriddenMap, _ = overrideRepo.BatchCheckOverridden(uuids, constant.HqFieldNegativeStock)
	}
	// 安全库存无控制模式，始终使用 override 逻辑保护，不受强制推送影响
	safetyOverriddenMap, _ := overrideRepo.BatchCheckOverridden(uuids, constant.HqFieldSafetyStock)

	// 获取子店当前值
	storeMaterialRepo := repository.NewMaterialRepo(storeDB)
	storeMaterials := storeMaterialRepo.GetMaterialList(
		commonRepo.WhereByHeadquarterUuid(hqUuid),
	)
	storeMaterialSet := make(map[uint64]bool, len(storeMaterials))
	for _, sm := range storeMaterials {
		storeMaterialSet[sm.Uuid] = true
	}

	for _, uuid := range uuids {
		if !storeMaterialSet[uuid] {
			continue
		}

		updateData := make(map[string]any)

		// 负库存
		hqNegVal := hqNegStockMap[uuid]
		if forceOverwrite {
			updateData["allow_negative_stock"] = hqNegVal
		} else if !negOverriddenMap[uuid] {
			updateData["allow_negative_stock"] = hqNegVal
		}

		// 安全库存（无控制模式，始终使用 override 逻辑，不受强制推送影响）
		if !safetyOverriddenMap[uuid] {
			updateData["safety_stock"] = hqSafetyStockMap[uuid]
		}

		if len(updateData) == 0 {
			continue
		}

		materialUuid := uuid
		if err := repository.CommonRepo.Transaction(storeDB, func(tx *gorm.DB) error {
			txMaterialRepo := repository.NewMaterialRepo(tx)
			if err := txMaterialRepo.UpdateMaterialData(updateData, commonRepo.WhereByUuid(materialUuid)); err != nil {
				return err
			}
			if forceOverwrite {
				txOverrideRepo := repository.NewHqFieldOverrideRepo(tx)
				txOverrideRepo.ClearOverride(materialUuid, constant.HqFieldNegativeStock)
				// 安全库存无控制模式，不清除其 override 标记
			}
			return nil
		}); err != nil {
			logger.Logger.Error("推送负库存失败",
				zap.Uint64("company_uuid", hqUuid),
				zap.Uint64("store_uuid", storeUuid),
				zap.Uint64("material_uuid", uuid),
				zap.Error(err),
			)
		}
	}
}

// pushProductToStores 推送单个商品的变更到多个子店
func (s *hqPushSrv) pushProductToStores(hqUuid uint64, storeUuids []uint64, productUuid uint64) {
	hqDB := s.dbm.GetDB(hqUuid)
	commonRepo := repository.NewCommonRepo()
	hqProductRepo := repository.NewProductPackageRepo(hqDB)

	// 读取 HQ 商品（含关联表：BOM/属性组/属性/套餐组/套餐子项/多语言名称/卖点多语言）
	hqProduct, err := hqProductRepo.GetProductPackage(
		commonRepo.WhereByUuid(productUuid),
		commonRepo.WhereByHeadquarterUuid(0),
		hqProductRepo.WithMultiLanguageName(),
		hqProductRepo.WithDescribeMultiLanguageName(),
		hqProductRepo.WithProductBoms(),
		hqProductRepo.WithProductPackageAttributeGroups(),
		hqProductRepo.WithProductPackageAttributeGroupAttributes(),
		hqProductRepo.WithProductPackageGroups(),
		hqProductRepo.WithProductPackageGroupItems(),
		hqProductRepo.WithProductPackageGroupMultiLanguageName(),
	)
	if err != nil || hqProduct == nil {
		logger.Logger.Error("获取总部商品失败", zap.Uint64("company_uuid", hqUuid), zap.Uint64("product_uuid", productUuid), zap.Error(err))
		return
	}

	controlRepo := s.getHqControlRepo(hqUuid)

	for _, storeUuid := range storeUuids {
		su := storeUuid
		// 每个子店独立构建 updateData，避免多个 goroutine 并发写同一个 map
		storeUpdateData := s.buildProductUpdateData(hqProduct)
		utils.Go(func() {
			s.pushSingleProductToStore(hqUuid, su, productUuid, hqProduct, controlRepo, storeUpdateData)
		})
	}
}

// buildProductUpdateData 构建商品更新数据（不可覆盖字段）
// 注意：status（店内上下架）由 pushSingleProductToStore 的 override 逻辑单独处理
// 注意：printer_tag_uuid（打印档口）不推送，各店本地维护
func (s *hqPushSrv) buildProductUpdateData(hqProduct *model.ProductPackage) map[string]any {
	updateData := map[string]any{
		// 基础信息
		"name":                     hqProduct.Name,
		"multi_language_name_uuid": hqProduct.MultiLanguageNameUuid,
		"category_uuid":            hqProduct.CategoryUuid,
		"special_category_uuid":    hqProduct.SpecialCategoryUuid,
		"unit_uuid":                hqProduct.UnitUuid,
		"erp_code":                 hqProduct.ErpCode,
		"supplier_uuid":            hqProduct.SupplierUuid,
		// 价格/税率
		"price":            hqProduct.Price,
		"dine_tax_uuid":    hqProduct.DineTaxUuid,
		"takeout_tax_uuid": hqProduct.TakeoutTaxUuid,
		// 图片
		"image_file_uuid": hqProduct.ImageFileUuid,
		"image_name":      hqProduct.ImageName,
		"image_url":       hqProduct.ImageUrl,
		// 计价
		"deduct_stock_type": hqProduct.DeductStockType,
		"num_type":          hqProduct.NumType,
		// 商品类型
		"product_type":        hqProduct.ProductType,
		"is_batch":            hqProduct.IsBatch,
		"sauce_required":      hqProduct.SauceRequired,
		"sauce_min_selection": hqProduct.SauceMinSelection,
		"sauce_max_selection": hqProduct.SauceMaxSelection,
		// 显示控制
		"is_show_cashier":   hqProduct.IsShowCashier,
		"is_show_tablet":    hqProduct.IsShowTablet,
		"is_show_kitchen":   hqProduct.IsShowKitchen,
		"is_show_assistant": hqProduct.IsShowAssistant,
		"is_show_h5":        hqProduct.IsShowH5,
		"is_show_delivery":  hqProduct.IsShowDelivery,
		"is_show_kiosk":     hqProduct.IsShowKiosk,
		// 排序/限购
		"sort":      hqProduct.Sort,
		"limit_num": hqProduct.LimitNum,
		// 描述/详情
		"describe":                          hqProduct.Describe,
		"describe_multi_language_name_uuid": hqProduct.DescribeMultiLanguageNameUuid,
		"detail":                            hqProduct.Detail,
		// 折扣/标签
		"product_label_uuid":    hqProduct.ProductLabelUuid,
		"open_discount":         hqProduct.OpenDiscount,
		"open_overall_discount": hqProduct.OpenOverallDiscount,
		// 删除/更新时间
		"delete_time": hqProduct.DeleteTime,
		"update_time": hqProduct.UpdateTime,
	}
	return updateData
}

// pushSingleProductToStore 推送单个商品到单个子店
func (s *hqPushSrv) pushSingleProductToStore(hqUuid, storeUuid uint64, productUuid uint64, hqProduct *model.ProductPackage, controlRepo repository.IHqControlSettingRepo, updateData map[string]any) {
	storeDB := s.dbm.GetDB(storeUuid)
	commonRepo := repository.NewCommonRepo()
	storeProductRepo := repository.NewProductPackageRepo(storeDB)
	overrideRepo := repository.NewHqFieldOverrideRepo(storeDB)

	// 检查子店是否存在该商品
	storeProduct, err := storeProductRepo.GetProductPackage(
		commonRepo.WhereByUuid(productUuid),
		commonRepo.WhereByHeadquarterUuid(hqUuid),
	)
	if err != nil || storeProduct == nil {
		return // 子店不存在该商品
	}

	// 同步多语言名称到子店数据库（商品名称 + 卖点描述）
	multiLangRepo := repository.NewMultiLanguageNameRepo(storeDB)
	nameUuid := s.syncMultiLanguageName(multiLangRepo, storeProduct.MultiLanguageNameUuid, &hqProduct.MultiLanguageName)
	descUuid := s.syncMultiLanguageName(multiLangRepo, storeProduct.DescribeMultiLanguageNameUuid, &hqProduct.DescribeMultiLanguageName)
	updateData["name"] = hqProduct.MultiLanguageName.ToJson()
	updateData["multi_language_name_uuid"] = nameUuid
	updateData["describe_multi_language_name_uuid"] = descUuid

	// 同步图片文件记录到子店数据库
	if hqProduct.ImageFileUuid > 0 && hqProduct.ImageFileUuid != storeProduct.ImageFileUuid {
		s.syncFileToStore(hqUuid, storeUuid, hqProduct.ImageFileUuid)
	}

	// 处理店内上下架（可覆盖字段，始终参与推送）
	if controlRepo.IsUnifiedControl(hqUuid, constant.HqFieldDineShelf) {
		updateData["status"] = hqProduct.Status
		overrideRepo.ClearOverride(productUuid, constant.HqFieldDineShelf)
	} else if !overrideRepo.IsOverridden(productUuid, constant.HqFieldDineShelf) {
		updateData["status"] = hqProduct.Status
	}
	// 已 override 则不设置 status，保留子店值

	// 在事务中执行主表更新和关联表同步
	if err := commonRepo.Transaction(storeDB, func(tx *gorm.DB) error {
		txProductRepo := repository.NewProductPackageRepo(tx)
		if err := txProductRepo.UpdateProductPackage(updateData, commonRepo.WhereByUuid(productUuid)); err != nil {
			return err
		}
		return s.syncSingleProductAssociations(tx, hqUuid, productUuid, hqProduct)
	}); err != nil {
		logger.Logger.Error("推送商品到子店事务失败",
			zap.Uint64("company_uuid", hqUuid),
			zap.Uint64("store_uuid", storeUuid),
			zap.Uint64("product_uuid", productUuid),
			zap.Error(err),
		)
	}
}

// syncSingleProductAssociations 同步单个商品的关联表到子店（delete+recreate，保留子店本地字段）
// 注意：调用方需确保 db 是事务连接（tx）
func (s *hqPushSrv) syncSingleProductAssociations(db *gorm.DB, hqUuid, productUuid uint64, hqProduct *model.ProductPackage) error {
	commonRepo := repository.NewCommonRepo()

	productBomRepo := repository.NewProductBomRepo(db)
	attrGroupRepo := repository.NewProductPackageAttributeGroupRepo(db)
	attrRepo := repository.NewProductPackageAttributeRepo(db)
	pkgGroupRepo := repository.NewProductPackageGroupRepo(db)

	// 1. 读取子店现有 BOM 的本地字段（库存/销量/状态等）
	existingBoms := make(map[uint64]model.ProductBom)
	storeBomList, _ := productBomRepo.GetProductBoms(
		commonRepo.WhereByProductPackageUuid(productUuid),
	)
	for _, bom := range storeBomList {
		existingBoms[bom.Uuid] = *bom
	}

	// 2. 收集子店需删除的关联 UUID
	storeProductRepo := repository.NewProductPackageRepo(db)
	storeProduct, err := storeProductRepo.GetProductPackage(
		commonRepo.WhereByUuid(productUuid),
		commonRepo.WhereByHeadquarterUuid(hqUuid),
		storeProductRepo.WithProductBoms(),
		storeProductRepo.WithProductPackageAttributeGroups(),
		storeProductRepo.WithProductPackageAttributeGroupAttributes(),
		storeProductRepo.WithProductPackageGroups(),
		storeProductRepo.WithProductPackageGroupItems(),
	)
	if err != nil || storeProduct == nil {
		return nil
	}

	delBomUuids := make([]uint64, 0)
	delAttrGroupUuids := make([]uint64, 0)
	delAttrUuids := make([]uint64, 0)
	delPkgGroupUuids := make([]uint64, 0)
	delPkgGroupItemUuids := make([]uint64, 0)

	for _, bom := range storeProduct.ProductBoms {
		delBomUuids = append(delBomUuids, bom.Uuid)
	}
	for _, ag := range storeProduct.ProductPackageAttributeGroups {
		delAttrGroupUuids = append(delAttrGroupUuids, ag.Uuid)
		for _, attr := range ag.ProductPackageAttributes {
			delAttrUuids = append(delAttrUuids, attr.Uuid)
		}
	}
	for _, pg := range storeProduct.ProductPackageGroups {
		delPkgGroupUuids = append(delPkgGroupUuids, pg.Uuid)
		for _, item := range pg.ProductPackageGroupItems {
			delPkgGroupItemUuids = append(delPkgGroupItemUuids, item.Uuid)
		}
	}

	// 3. 删除子店关联数据
	if len(delBomUuids) > 0 {
		if err := productBomRepo.DestroyProductBom(commonRepo.WhereInUuids(delBomUuids)); err != nil {
			return err
		}
	}
	if len(delAttrGroupUuids) > 0 {
		if err := attrGroupRepo.DestroyProductPackageAttributeGroup(commonRepo.WhereInUuids(delAttrGroupUuids)); err != nil {
			return err
		}
	}
	if len(delAttrUuids) > 0 {
		if err := attrRepo.DestroyProductPackageAttribute(commonRepo.WhereInUuids(delAttrUuids)); err != nil {
			return err
		}
	}
	if len(delPkgGroupUuids) > 0 {
		if err := pkgGroupRepo.DestroyProductPackageGroup(commonRepo.WhereInUuids(delPkgGroupUuids)); err != nil {
			return err
		}
	}
	if len(delPkgGroupItemUuids) > 0 {
		if err := pkgGroupRepo.DestroyProductPackageGroupItem(commonRepo.WhereInUuids(delPkgGroupItemUuids)); err != nil {
			return err
		}
	}

	// 4. 从 HQ 数据重建 BOM（保留子店本地字段）
	for _, hqBom := range hqProduct.ProductBoms {
		var existing *model.ProductBom
		if e, ok := existingBoms[hqBom.Uuid]; ok {
			existing = &e
		}
		newBom := buildSubStoreBom(hqBom, existing)
		if _, err := productBomRepo.CreateProductBom(newBom); err != nil {
			logger.Logger.Error("重建商品BOM失败", zap.Uint64("bom_uuid", hqBom.Uuid), zap.Error(err))
		}
	}

	// 5. 重建属性组和属性
	for _, hqAttrGroup := range hqProduct.ProductPackageAttributeGroups {
		if err := attrGroupRepo.CreateProductPackageAttributeGroups([]model.ProductPackageAttributeGroup{buildSubStoreAttrGroup(hqAttrGroup)}); err != nil {
			logger.Logger.Error("重建商品属性组失败", zap.Uint64("attr_group_uuid", hqAttrGroup.Uuid), zap.Error(err))
		}
		for _, hqAttr := range hqAttrGroup.ProductPackageAttributes {
			if err := attrRepo.CreateProductPackageAttributes([]model.ProductPackageAttribute{buildSubStoreAttr(hqAttr)}); err != nil {
				logger.Logger.Error("重建商品属性失败", zap.Uint64("attr_uuid", hqAttr.Uuid), zap.Error(err))
			}
		}
	}

	// 6. 重建套餐组和套餐子项
	for _, hqPkgGroup := range hqProduct.ProductPackageGroups {
		if err := pkgGroupRepo.CreateProductPackageGroups([]model.ProductPackageGroup{buildSubStorePkgGroup(hqPkgGroup)}); err != nil {
			logger.Logger.Error("重建商品套餐组失败", zap.Uint64("pkg_group_uuid", hqPkgGroup.Uuid), zap.Error(err))
		}
		for _, hqItem := range hqPkgGroup.ProductPackageGroupItems {
			if err := pkgGroupRepo.CreateProductPackageGroupItems([]model.ProductPackageGroupItem{buildSubStorePkgGroupItem(hqItem)}); err != nil {
				logger.Logger.Error("重建商品套餐子项失败", zap.Uint64("item_uuid", hqItem.Uuid), zap.Error(err))
			}
		}
	}

	return nil
}

// pushProductTakeoutToStores 推送外卖商品变更到多个子店
func (s *hqPushSrv) pushProductTakeoutToStores(hqUuid uint64, storeUuids []uint64, takeoutUuid uint64) {
	hqDB := s.dbm.GetDB(hqUuid)
	commonRepo := repository.NewCommonRepo()

	hqTakeoutRepo := repository.NewProductPackageTakeoutRepo(hqDB)
	hqTakeout, err := hqTakeoutRepo.GetProductPackageTakeout(
		commonRepo.WhereByUuid(takeoutUuid),
		commonRepo.WhereByHeadquarterUuid(0),
		hqTakeoutRepo.WithProductBomTakeouts(),
		hqTakeoutRepo.WithProductPackageAttributeTakeouts(),
		hqTakeoutRepo.WithProductPackageGroupItemTakeouts(),
		hqTakeoutRepo.WithMultiLanguageName(),
		hqTakeoutRepo.WithDescribeMultiLanguageName(),
	)
	if err != nil || hqTakeout == nil {
		return
	}

	controlRepo := s.getHqControlRepo(hqUuid)

	for _, storeUuid := range storeUuids {
		su := storeUuid
		utils.Go(func() {
			s.pushSingleTakeoutToStore(hqUuid, su, hqTakeout, controlRepo)
		})
	}
}

// pushSingleTakeoutToStore 推送单个外卖商品到单个子店
func (s *hqPushSrv) pushSingleTakeoutToStore(hqUuid, storeUuid uint64, hqTakeout *model.ProductPackageTakeout, controlRepo repository.IHqControlSettingRepo) {
	storeDB := s.dbm.GetDB(storeUuid)
	commonRepo := repository.NewCommonRepo()
	storeTakeoutRepo := repository.NewProductPackageTakeoutRepo(storeDB)
	overrideRepo := repository.NewHqFieldOverrideRepo(storeDB)

	// 查找子店外卖商品（用 ProductPackageUuid+TakeoutType+HeadquarterUuid 匹配，与全量同步一致）
	storeTakeout, err := storeTakeoutRepo.GetProductPackageTakeout(
		storeTakeoutRepo.WhereByProductPackageUuid(hqTakeout.ProductPackageUuid),
		storeTakeoutRepo.WhereByTakeoutType(hqTakeout.TakeoutType),
		commonRepo.WhereByHeadquarterUuid(hqUuid),
		storeTakeoutRepo.WithProductBomTakeouts(),
		storeTakeoutRepo.WithProductPackageAttributeTakeouts(),
		storeTakeoutRepo.WithProductPackageGroupItemTakeouts(),
	)
	if err != nil || storeTakeout == nil {
		// 子店不存在该外卖商品 → 创建（默认下架）
		storeTakeout = s.createTakeoutInStore(storeDB, hqTakeout, hqUuid)
		if storeTakeout == nil {
			return
		}
		// 新创建的记录只需同步关联表
		if err := commonRepo.Transaction(storeDB, func(tx *gorm.DB) error {
			return syncTakeoutAssociations(tx, hqTakeout, storeTakeout, hqUuid, true)
		}); err != nil {
			logger.Logger.Error("同步新创建外卖商品关联表失败",
				zap.Uint64("company_uuid", hqUuid),
				zap.Uint64("store_uuid", storeUuid),
				zap.Uint64("takeout_uuid", storeTakeout.Uuid),
				zap.Error(err),
			)
		}
		return
	}

	// 同步多语言名称到子店数据库（更新已有记录或创建新记录）
	multiLangRepo := repository.NewMultiLanguageNameRepo(storeDB)
	nameUuid := s.syncMultiLanguageName(multiLangRepo, storeTakeout.MultiLanguageNameUuid, &hqTakeout.MultiLanguageName)
	descUuid := s.syncMultiLanguageName(multiLangRepo, storeTakeout.DescribeMultiLanguageNameUuid, &hqTakeout.DescribeMultiLanguageName)

	// 同步图片文件记录到子店数据库
	if hqTakeout.ImageFileUuid > 0 && hqTakeout.ImageFileUuid != storeTakeout.ImageFileUuid {
		s.syncFileToStore(hqUuid, storeUuid, hqTakeout.ImageFileUuid)
	}

	// 构建不可覆盖字段更新数据
	updateData := map[string]any{
		"name":                              hqTakeout.Name,
		"multi_language_name_uuid":          nameUuid,
		"category_uuid":                     hqTakeout.CategoryUuid,
		"special_category_uuid":             hqTakeout.SpecialCategoryUuid,
		"image_file_uuid":                   hqTakeout.ImageFileUuid,
		"describe":                          hqTakeout.Describe,
		"describe_multi_language_name_uuid": descUuid,
		"product_type":                      hqTakeout.ProductType,
		"takeout_type":                      hqTakeout.TakeoutType,
		"source":                            hqTakeout.Source,
		"source_product_id":                 hqTakeout.SourceProductId,
		"update_time":                       hqTakeout.UpdateTime,
		"delete_time":                       hqTakeout.DeleteTime,
	}

	// 处理外卖上下架（可覆盖字段，始终参与推送）
	if controlRepo.IsUnifiedControl(hqUuid, constant.HqFieldTakeoutShelf) {
		updateData["status"] = hqTakeout.Status
		overrideRepo.ClearOverride(storeTakeout.Uuid, constant.HqFieldTakeoutShelf)
	} else if !overrideRepo.IsOverridden(storeTakeout.Uuid, constant.HqFieldTakeoutShelf) {
		updateData["status"] = hqTakeout.Status
	}

	// 处理外卖价格（可覆盖字段，始终参与推送）
	if controlRepo.IsUnifiedControl(hqUuid, constant.HqFieldTakeoutPrice) {
		updateData["price"] = hqTakeout.Price
		overrideRepo.ClearOverride(storeTakeout.Uuid, constant.HqFieldTakeoutPrice)
	} else if !overrideRepo.IsOverridden(storeTakeout.Uuid, constant.HqFieldTakeoutPrice) {
		updateData["price"] = hqTakeout.Price
	}

	// 在事务中执行主表更新和关联表同步
	syncBomPrice := false
	if _, ok := updateData["price"]; ok {
		syncBomPrice = true
	}
	if err := commonRepo.Transaction(storeDB, func(tx *gorm.DB) error {
		txTakeoutRepo := repository.NewProductPackageTakeoutRepo(tx)
		if err := txTakeoutRepo.UpdateProductPackageTakeout(updateData, commonRepo.WhereByUuid(storeTakeout.Uuid)); err != nil {
			return err
		}
		return syncTakeoutAssociations(tx, hqTakeout, storeTakeout, hqUuid, syncBomPrice)
	}); err != nil {
		logger.Logger.Error("推送外卖商品到子店事务失败",
			zap.Uint64("company_uuid", hqUuid),
			zap.Uint64("store_uuid", storeUuid),
			zap.Uint64("takeout_uuid", hqTakeout.Uuid),
			zap.Error(err),
		)
	}
}

// createTakeoutInStore 在子店创建外卖商品主记录
// 前置条件：子店必须已有对应的 ProductPackage（店内商品），否则跳过
func (s *hqPushSrv) createTakeoutInStore(storeDB *gorm.DB, hqTakeout *model.ProductPackageTakeout, hqUuid uint64) *model.ProductPackageTakeout {
	commonRepo := repository.NewCommonRepo()

	// 检查子店是否存在对应的店内商品，不存在则跳过
	_, err := repository.NewProductPackageRepo(storeDB).GetProductPackage(
		commonRepo.WhereByUuid(hqTakeout.ProductPackageUuid),
		commonRepo.WhereByHeadquarterUuid(hqUuid),
		commonRepo.WhereBySoftDelete(),
	)
	if err != nil {
		logger.Logger.Warn("子店不存在对应店内商品，跳过创建外卖商品",
			zap.Uint64("product_package_uuid", hqTakeout.ProductPackageUuid),
			zap.Uint64("headquarter_uuid", hqUuid),
		)
		return nil
	}

	// 在子店数据库创建多语言名称记录（名称 + 卖点）
	// 注意：Name/Describe 字段可能为空，实际多语言数据存储在 MultiLanguageName 关联表中
	multiLangRepo := repository.NewMultiLanguageNameRepo(storeDB)
	var multiLanguageNameUuid uint64
	if !hqTakeout.MultiLanguageName.IsNullName() {
		mlName := hqTakeout.MultiLanguageName
		mlName.ID = 0
		mlName.Uuid = 0
		multiLanguageNameUuid, _ = multiLangRepo.CreateMultiLanguageName(mlName)
	}
	var describeMultiLanguageNameUuid uint64
	if !hqTakeout.DescribeMultiLanguageName.IsNullName() {
		mlDescribe := hqTakeout.DescribeMultiLanguageName
		mlDescribe.ID = 0
		mlDescribe.Uuid = 0
		describeMultiLanguageNameUuid, _ = multiLangRepo.CreateMultiLanguageName(mlDescribe)
	}

	newTakeout := &model.ProductPackageTakeout{
		ProductPackageUuid:            hqTakeout.ProductPackageUuid,
		MultiLanguageNameUuid:         multiLanguageNameUuid,
		HeadquarterUuid:               hqUuid,
		Name:                          hqTakeout.Name,
		Describe:                      hqTakeout.Describe,
		DescribeMultiLanguageNameUuid: describeMultiLanguageNameUuid,
		ProductType:                   hqTakeout.ProductType,
		Price:                         hqTakeout.Price,
		TakeoutType:                   hqTakeout.TakeoutType,
		Status:                        0, // 默认下架，子店需手动上架
		CategoryUuid:                  hqTakeout.CategoryUuid,
		SpecialCategoryUuid:           hqTakeout.SpecialCategoryUuid,
		ImageFileUuid:                 hqTakeout.ImageFileUuid,
		Source:                        hqTakeout.Source,
		SourceProductId:               hqTakeout.SourceProductId,
	}
	// Uuid 由 BaseModel.BeforeCreate 自动生成
	if err := repository.NewProductPackageTakeoutRepo(storeDB).CreateProductPackageTakeout(newTakeout); err != nil {
		logger.Logger.Error("在子店创建外卖商品失败",
			zap.Uint64("product_package_uuid", hqTakeout.ProductPackageUuid),
			zap.Uint64("headquarter_uuid", hqUuid),
			zap.Error(err),
		)
		return nil
	}

	return newTakeout
}

// syncMultiLanguageName 同步多语言名称到子店数据库
// 如果子店已有记录（existingUuid > 0）则更新，否则创建新记录
func (s *hqPushSrv) syncMultiLanguageName(repo repository.IMultiLanguageNameRepo, existingUuid uint64, hqLang *model.MultiLanguageName) uint64 {
	if hqLang == nil || hqLang.IsNullName() {
		return existingUuid
	}
	if existingUuid > 0 {
		_ = repo.UpdateMultiLanguageName(existingUuid, *hqLang)
		return existingUuid
	}
	newLang := *hqLang
	newLang.ID = 0
	newLang.Uuid = 0
	uuid, _ := repo.CreateMultiLanguageName(newLang)
	return uuid
}

// syncFileToStore 将 HQ 的文件记录复制到子店数据库（如果子店不存在该文件）
func (s *hqPushSrv) syncFileToStore(hqUuid, storeUuid, fileUuid uint64) {
	if fileUuid == 0 {
		return
	}
	storeDB := s.dbm.GetDB(storeUuid)
	storeFileRepo := repository.NewFileRepo(storeDB)

	// 检查子店是否已有该文件
	if _, err := storeFileRepo.GetFileByUuid(fileUuid); err == nil {
		return // 已存在，无需同步
	}

	// 从 HQ 读取文件记录并复制到子店
	hqDB := s.dbm.GetDB(hqUuid)
	hqFileRepo := repository.NewFileRepo(hqDB)
	hqFile, err := hqFileRepo.GetFileByUuid(fileUuid)
	if err != nil || hqFile == nil {
		return
	}

	newFile := *hqFile
	newFile.ID = 0
	newFile.HeadquarterUuid = hqUuid
	_ = storeFileRepo.CreateFile(newFile)
}

// pushMaterialToStores 推送物品变更到多个子店
func (s *hqPushSrv) pushMaterialToStores(hqUuid uint64, storeUuids []uint64, materialUuid uint64) {
	hqDB := s.dbm.GetDB(hqUuid)
	commonRepo := repository.NewCommonRepo()

	hqMaterialRepo := repository.NewMaterialRepo(hqDB)
	hqMaterial := hqMaterialRepo.GetMaterial(
		commonRepo.WhereByUuid(materialUuid),
		commonRepo.WhereByHeadquarterUuid(0),
		hqMaterialRepo.WithNotBaseUnitList(commonRepo.WhereBySoftDelete()),
		hqMaterialRepo.WithMultiLanguageName(),
	)
	if hqMaterial.Uuid == 0 {
		return
	}

	controlRepo := s.getHqControlRepo(hqUuid)

	sem := make(chan struct{}, 10)
	for _, storeUuid := range storeUuids {
		su := storeUuid
		utils.Go(func() {
			sem <- struct{}{}
			defer func() { <-sem }()
			s.pushSingleMaterialToStore(hqUuid, su, &hqMaterial, controlRepo)
		})
	}
}

// pushSingleMaterialToStore 推送单个物品到单个子店
func (s *hqPushSrv) pushSingleMaterialToStore(hqUuid, storeUuid uint64, hqMaterial *model.Material, controlRepo repository.IHqControlSettingRepo) {
	storeDB := s.dbm.GetDB(storeUuid)
	commonRepo := repository.NewCommonRepo()
	storeMaterialRepo := repository.NewMaterialRepo(storeDB)
	overrideRepo := repository.NewHqFieldOverrideRepo(storeDB)

	storeMaterial := storeMaterialRepo.GetMaterial(
		commonRepo.WhereByUuid(hqMaterial.Uuid),
		commonRepo.WhereByHeadquarterUuid(hqUuid),
	)
	if storeMaterial.Uuid == 0 {
		return
	}

	// 检查分类是否存在于子店，不存在则默认为 0
	categoryUuid := hqMaterial.CategoryUuid
	if categoryUuid > 0 {
		if _, err := storeMaterialRepo.GetMaterialCategoryByUuid(categoryUuid); err != nil {
			categoryUuid = 0
		}
	}

	// 检查单位 ProductUnit 是否存在于子店，过滤有效单位
	storeProductUnitRepo := repository.NewProductUnitRepo(storeDB)
	validUnits, validUnitUuidSet := resolveValidMaterialUnits(storeProductUnitRepo, hqMaterial.NotBaseUnitList)

	// 解析单位字段（参考 doUpdateMaterialByErpItem 处理不存在的引用）
	unitUuid, purchaseUnitUuid, costUnitUuid, defaultSalesUnitUuid := resolvePushMaterialUnitFields(
		hqMaterial, validUnitUuidSet,
	)

	// 不可覆盖字段（multi_language_name_uuid 保留子店自有值，仅更新内容）
	updateData := map[string]any{
		"name":                    hqMaterial.Name,
		"code":                    hqMaterial.Code,
		"category_uuid":           categoryUuid,
		"status":                   hqMaterial.Status,
		"barcode_value":            hqMaterial.BarcodeValue,
		"internal_code":            hqMaterial.InternalCode,
		"origin_country_code":      hqMaterial.OriginCountryCode,
		"unit_uuid":                unitUuid,
		"purchase_unit_uuid":       purchaseUnitUuid,
		"cost_unit_uuid":           costUnitUuid,
		"default_sales_unit_uuid":  defaultSalesUnitUuid,
		"allow_substore_visible":   hqMaterial.AllowSubstoreVisible,
		"delete_time":              hqMaterial.DeleteTime,
		"update_time":              hqMaterial.UpdateTime,
	}

	// 安全库存（可覆盖字段，无控制模式，始终使用 override 逻辑）
	if overrideRepo.IsOverridden(hqMaterial.Uuid, constant.HqFieldSafetyStock) {
		// 已 override：保留子店值
	} else {
		updateData["safety_stock"] = hqMaterial.SafetyStock
	}

	// 负库存（可覆盖字段，有控制模式）
	if controlRepo.IsUnifiedControl(hqUuid, constant.HqFieldNegativeStock) {
		updateData["allow_negative_stock"] = hqMaterial.AllowNegativeStock
		overrideRepo.ClearOverride(hqMaterial.Uuid, constant.HqFieldNegativeStock)
	} else if overrideRepo.IsOverridden(hqMaterial.Uuid, constant.HqFieldNegativeStock) {
		// 分开控制+已修改：保留子店值
	} else {
		updateData["allow_negative_stock"] = hqMaterial.AllowNegativeStock
	}

	// 在事务中执行主表更新、多语言名称同步和非基准单位同步
	if err := commonRepo.Transaction(storeDB, func(tx *gorm.DB) error {
		txMaterialRepo := repository.NewMaterialRepo(tx)
		if err := txMaterialRepo.UpdateMaterialData(updateData, commonRepo.WhereByUuid(hqMaterial.Uuid)); err != nil {
			return err
		}

		// 同步多语言名称记录到子店（使用子店自己的 multi_language_name_uuid 更新内容）
		if storeMaterial.MultiLanguageNameUuid > 0 && !hqMaterial.MultiLanguageName.IsNullName() {
			multiLanguageNameRepo := repository.NewMultiLanguageNameRepo(tx)
			if err := multiLanguageNameRepo.UpdateMultiLanguageName(storeMaterial.MultiLanguageNameUuid, hqMaterial.MultiLanguageName); err != nil {
				return err
			}
		}

		// 同步单位（delete+recreate，跳过子店不存在 ProductUnit 的单位）
		materialUnitRepo := repository.NewMaterialUnitRepo(tx)
		if err := materialUnitRepo.DestroyMaterialUnit(commonRepo.WhereByMaterialUuid(hqMaterial.Uuid)); err != nil {
			return err
		}
		if len(validUnits) > 0 {
			return materialUnitRepo.CreateMaterialUnitList(validUnits)
		}
		return nil
	}); err != nil {
		logger.Logger.Error("推送物品到子店事务失败",
			zap.Uint64("company_uuid", hqUuid),
			zap.Uint64("store_uuid", storeUuid),
			zap.Uint64("material_uuid", hqMaterial.Uuid),
			zap.Error(err),
		)
	}
}

// resolveValidMaterialUnits 检查子店 ProductUnit 存在性，过滤有效的 MaterialUnit 列表
func resolveValidMaterialUnits(productUnitRepo repository.IProductUnitRepo, hqUnits []*model.MaterialUnit) ([]model.MaterialUnit, map[uint64]bool) {
	validUnitUuidSet := make(map[uint64]bool)
	if len(hqUnits) == 0 {
		return nil, validUnitUuidSet
	}

	// 收集 HQ 单位引用的 ProductUnit UUID
	productUnitUuids := make([]uint64, 0, len(hqUnits))
	for _, u := range hqUnits {
		if u.UnitUuid > 0 {
			productUnitUuids = append(productUnitUuids, u.UnitUuid)
		}
	}

	// 批量查询子店中存在的 ProductUnit
	existingProductUnitSet := make(map[uint64]bool)
	if len(productUnitUuids) > 0 {
		existingUnits, err := productUnitRepo.GetProductUnitList(productUnitRepo.WhereByUuids(productUnitUuids))
		if err == nil {
			for _, pu := range existingUnits {
				existingProductUnitSet[pu.Uuid] = true
			}
		}
	}

	// 过滤出 ProductUnit 存在于子店的有效单位
	validUnits := make([]model.MaterialUnit, 0, len(hqUnits))
	for _, u := range hqUnits {
		if u.UnitUuid > 0 && !existingProductUnitSet[u.UnitUuid] {
			logger.Logger.Warn("推送物品-ProductUnit在子店不存在，跳过单位",
				zap.Uint64("material_unit_uuid", u.Uuid),
				zap.Uint64("product_unit_uuid", u.UnitUuid),
			)
			continue
		}
		validUnits = append(validUnits, model.MaterialUnit{
			BaseModel:      model.BaseModel{Uuid: u.Uuid, CreateTime: u.CreateTime, UpdateTime: u.UpdateTime, DeleteTime: u.DeleteTime},
			Name:           u.Name,
			UnitUuid:       u.UnitUuid,
			ConversionRate: u.ConversionRate,
			FromUnitUuid:   u.FromUnitUuid,
			IsDefault:      u.IsDefault,
			MaterialUuid:   u.MaterialUuid,
		})
		validUnitUuidSet[u.Uuid] = true
	}
	return validUnits, validUnitUuidSet
}

// resolvePushMaterialUnitFields 解析推送物品的单位字段，参考 doUpdateMaterialByErpItem 处理不存在的引用
func resolvePushMaterialUnitFields(hqMaterial *model.Material, validUnitUuidSet map[uint64]bool) (unitUuid, purchaseUnitUuid, costUnitUuid, defaultSalesUnitUuid uint64) {
	// 基准单位：不存在则清零
	unitUuid = hqMaterial.UnitUuid
	if unitUuid > 0 && !validUnitUuidSet[unitUuid] {
		unitUuid = 0
	}

	// 采购单位：不存在则清零
	purchaseUnitUuid = hqMaterial.PurchaseUnitUuid
	if purchaseUnitUuid > 0 && !validUnitUuidSet[purchaseUnitUuid] {
		purchaseUnitUuid = 0
	}

	// 成本单位：不存在则回退到基准单位（参考 handleDanglingUnitRefs）
	costUnitUuid = hqMaterial.CostUnitUuid
	if costUnitUuid > 0 && !validUnitUuidSet[costUnitUuid] {
		costUnitUuid = unitUuid
	}

	// 默认销售单位：不存在则清零
	defaultSalesUnitUuid = hqMaterial.DefaultSalesUnitUuid
	if defaultSalesUnitUuid > 0 && !validUnitUuidSet[defaultSalesUnitUuid] {
		defaultSalesUnitUuid = 0
	}
	return
}

// ========== 关联表构建 Helper（product.go 全量同步 和 hq_push.go 实时推送 共用） ==========

// buildSubStoreBom 从 HQ BOM 构建子店 BOM，保留本地字段（库存/销量/沽清/上下架等）
// existing 为子店已有的 BOM，nil 表示子店不存在该 BOM
func buildSubStoreBom(hqBom model.ProductBom, existing *model.ProductBom) model.ProductBom {
	bomStockNum := float64(0)
	bomActualSaleNum := float64(0)
	bomUseBomCardStock := 0
	bomIsSoldOut := 0
	bomIsOpenStock := 0
	bomStatus := hqBom.Status

	if existing != nil {
		bomStockNum = existing.StockNum
		bomActualSaleNum = existing.ActualSaleNum
		bomUseBomCardStock = existing.UseBomCardStock
		bomIsSoldOut = existing.IsSoldOut
		bomIsOpenStock = existing.IsOpenStock
		bomStatus = existing.Status
	}

	return model.ProductBom{
		BaseModel: model.BaseModel{
			Uuid:       hqBom.Uuid,
			CreateTime: hqBom.CreateTime,
			UpdateTime: hqBom.UpdateTime,
			DeleteTime: hqBom.DeleteTime,
		},
		PurchasePrice:      hqBom.PurchasePrice,
		Price:              hqBom.Price,
		Name:               hqBom.Name,
		ErpCode:            hqBom.ErpCode,
		StockNum:           bomStockNum,
		IsOpenStock:        bomIsOpenStock,
		BarcodeValue:       hqBom.BarcodeValue,
		InternalCode:       hqBom.InternalCode,
		IsDefaultSelect:    hqBom.IsDefaultSelect,
		Status:             bomStatus,
		IsSoldOut:          bomIsSoldOut,
		ActualSaleNum:      bomActualSaleNum,
		ProductFlavorUuid:  hqBom.ProductFlavorUuid,
		ProductSauceUuid:   hqBom.ProductSauceUuid,
		ProductPackageUuid: hqBom.ProductPackageUuid,
		ProductBomCardUuid: hqBom.ProductBomCardUuid,
		UseBomCardStock:    bomUseBomCardStock,
	}
}

// buildSubStoreAttrGroup 从 HQ 属性组构建子店属性组
func buildSubStoreAttrGroup(hq model.ProductPackageAttributeGroup) model.ProductPackageAttributeGroup {
	return model.ProductPackageAttributeGroup{
		BaseModel: model.BaseModel{
			Uuid:       hq.Uuid,
			CreateTime: hq.CreateTime,
			UpdateTime: hq.UpdateTime,
			DeleteTime: hq.DeleteTime,
		},
		IsMust:                    hq.IsMust,
		MaxSelection:              hq.MaxSelection,
		MinSelection:              hq.MinSelection,
		ProductPackageUuid:        hq.ProductPackageUuid,
		ProductAttributeGroupUuid: hq.ProductAttributeGroupUuid,
	}
}

// buildSubStoreAttr 从 HQ 属性构建子店属性
func buildSubStoreAttr(hq model.ProductPackageAttribute) model.ProductPackageAttribute {
	return model.ProductPackageAttribute{
		BaseModel: model.BaseModel{
			Uuid:       hq.Uuid,
			CreateTime: hq.CreateTime,
			UpdateTime: hq.UpdateTime,
			DeleteTime: hq.DeleteTime,
		},
		ProductPackageAttributeGroupUuid: hq.ProductPackageAttributeGroupUuid,
		AttributeUuid:                    hq.AttributeUuid,
		IsDefaultSelected:                hq.IsDefaultSelected,
	}
}

// buildSubStorePkgGroup 从 HQ 套餐组构建子店套餐组
func buildSubStorePkgGroup(hq model.ProductPackageGroup) model.ProductPackageGroup {
	return model.ProductPackageGroup{
		BaseModel: model.BaseModel{
			Uuid:       hq.Uuid,
			CreateTime: hq.CreateTime,
			UpdateTime: hq.UpdateTime,
			DeleteTime: hq.DeleteTime,
		},
		Name:                  hq.MultiLanguageName.ToJson(),
		ProductPackageUuid:    hq.ProductPackageUuid,
		MultiLanguageNameUuid: hq.MultiLanguageName.Uuid,
		GroupType:             hq.GroupType,
		OptionalCount:         hq.OptionalCount,
		OptionalMinCount:      hq.OptionalMinCount,
	}
}

// buildSubStorePkgGroupItem 从 HQ 套餐子项构建子店套餐子项
func buildSubStorePkgGroupItem(hq model.ProductPackageGroupItem) model.ProductPackageGroupItem {
	return model.ProductPackageGroupItem{
		BaseModel: model.BaseModel{
			Uuid:       hq.Uuid,
			CreateTime: hq.CreateTime,
			UpdateTime: hq.UpdateTime,
			DeleteTime: hq.DeleteTime,
		},
		ProductPackageGroupUuid: hq.ProductPackageGroupUuid,
		RelatedUuid:             hq.RelatedUuid,
		ProductBomUuid:          hq.ProductBomUuid,
		Num:                     hq.Num,
		Sort:                    hq.Sort,
		AddPrice:                hq.AddPrice,
		IsRequired:              hq.IsRequired,
		IsDefault:               hq.IsDefault,
	}
}

// syncTakeoutAssociations 同步外卖商品关联表（BomTakeout / AttributeTakeout / GroupItemTakeout）
// 使用 upsert 模式：match by key → update if exists, create if not
// syncBomPrice: 是否同步 BOM 价格（仅当主表 price 被同步时为 true）
func syncTakeoutAssociations(
	tx *gorm.DB,
	hqTakeout *model.ProductPackageTakeout,
	storeTakeout *model.ProductPackageTakeout,
	hqUuid uint64,
	syncBomPrice bool,
) error {
	commonRepo := repository.NewCommonRepo()
	bomTakeoutRepo := repository.NewProductBomTakeoutRepo(tx)
	attrTakeoutRepo := repository.NewProductPackageAttributeTakeoutRepo(tx)
	groupItemTakeoutRepo := repository.NewProductPackageGroupItemTakeoutRepo(tx)

	// --- 1. BomTakeout: match by product_bom_uuid ---
	subBomMap := make(map[uint64]*model.ProductBomTakeout)
	for i := range storeTakeout.ProductBomTakeouts {
		bom := &storeTakeout.ProductBomTakeouts[i]
		subBomMap[bom.ProductBomUuid] = bom
	}

	for _, hqBom := range hqTakeout.ProductBomTakeouts {
		if subBom, exists := subBomMap[hqBom.ProductBomUuid]; exists {
			updateData := map[string]any{
				"create_time": hqBom.CreateTime,
				"update_time": hqBom.UpdateTime,
				"delete_time": hqBom.DeleteTime,
			}
			if syncBomPrice {
				updateData["price"] = hqBom.Price
			}
			if err := bomTakeoutRepo.UpdateProductBomTakeout(updateData, commonRepo.WhereByUuid(subBom.Uuid)); err != nil {
				logger.Logger.Error("更新子店外卖规格价格失败",
					zap.Uint64("uuid", subBom.Uuid),
					zap.Uint64("product_bom_uuid", hqBom.ProductBomUuid),
					zap.Error(err))
			}
		} else {
			newBom := model.ProductBomTakeout{
				BaseModel: model.BaseModel{
					Uuid:       hqBom.Uuid,
					CreateTime: hqBom.CreateTime,
					UpdateTime: hqBom.UpdateTime,
					DeleteTime: hqBom.DeleteTime,
				},
				ProductPackageTakeoutUuid: storeTakeout.Uuid,
				ProductBomUuid:            hqBom.ProductBomUuid,
				HeadquarterUuid:           hqUuid,
				Price:                     hqBom.Price,
				GrabModifierId:            hqBom.GrabModifierId,
			}
			if err := bomTakeoutRepo.CreateProductBomTakeout(&newBom); err != nil {
				logger.Logger.Error("创建子店外卖规格价格失败",
					zap.Uint64("product_bom_uuid", hqBom.ProductBomUuid),
					zap.Error(err))
			}
		}
	}

	// --- 2. AttributeTakeout: match by product_package_attribute_uuid ---
	subAttrMap := make(map[uint64]*model.ProductPackageAttributeTakeout)
	for i := range storeTakeout.ProductPackageAttributeTakeouts {
		attr := &storeTakeout.ProductPackageAttributeTakeouts[i]
		subAttrMap[attr.ProductPackageAttributeUuid] = attr
	}

	for _, hqAttr := range hqTakeout.ProductPackageAttributeTakeouts {
		if subAttr, exists := subAttrMap[hqAttr.ProductPackageAttributeUuid]; exists {
			if err := attrTakeoutRepo.UpdateProductPackageAttributeTakeout(
				map[string]any{
					"create_time": hqAttr.CreateTime,
					"update_time": hqAttr.UpdateTime,
					"delete_time": hqAttr.DeleteTime,
				},
				commonRepo.WhereByUuid(subAttr.Uuid),
			); err != nil {
				logger.Logger.Error("更新子店外卖属性价格失败",
					zap.Uint64("uuid", subAttr.Uuid),
					zap.Uint64("product_package_attribute_uuid", hqAttr.ProductPackageAttributeUuid),
					zap.Error(err))
			}
		} else {
			newAttr := model.ProductPackageAttributeTakeout{
				BaseModel: model.BaseModel{
					Uuid:       hqAttr.Uuid,
					CreateTime: hqAttr.CreateTime,
					UpdateTime: hqAttr.UpdateTime,
					DeleteTime: hqAttr.DeleteTime,
				},
				ProductPackageTakeoutUuid:   storeTakeout.Uuid,
				ProductPackageAttributeUuid: hqAttr.ProductPackageAttributeUuid,
				HeadquarterUuid:             hqUuid,
				Price:                       hqAttr.Price,
			}
			if err := attrTakeoutRepo.CreateProductPackageAttributeTakeout(&newAttr); err != nil {
				logger.Logger.Error("创建子店外卖属性价格失败",
					zap.Uint64("product_package_attribute_uuid", hqAttr.ProductPackageAttributeUuid),
					zap.Error(err))
			}
		}
	}

	// --- 3. GroupItemTakeout: match by composite key ---
	subGroupItemMap := make(map[string]*model.ProductPackageGroupItemTakeout)
	for i := range storeTakeout.ProductPackageGroupItemTakeouts {
		gi := &storeTakeout.ProductPackageGroupItemTakeouts[i]
		key := fmt.Sprintf("%d-%d", gi.ProductPackageGroupUuid, gi.ProductPackageGroupItemUuid)
		subGroupItemMap[key] = gi
	}

	for _, hqGI := range hqTakeout.ProductPackageGroupItemTakeouts {
		key := fmt.Sprintf("%d-%d", hqGI.ProductPackageGroupUuid, hqGI.ProductPackageGroupItemUuid)
		if subGI, exists := subGroupItemMap[key]; exists {
			if err := groupItemTakeoutRepo.UpdateData(subGI.Uuid, map[string]any{
				"create_time": hqGI.CreateTime,
				"update_time": hqGI.UpdateTime,
				"delete_time": hqGI.DeleteTime,
			}); err != nil {
				logger.Logger.Error("更新子店外卖套餐子商品价格失败",
					zap.Uint64("uuid", subGI.Uuid),
					zap.Uint64("product_package_group_uuid", hqGI.ProductPackageGroupUuid),
					zap.Uint64("product_package_group_item_uuid", hqGI.ProductPackageGroupItemUuid),
					zap.Error(err))
			}
		} else {
			newGI := model.ProductPackageGroupItemTakeout{
				BaseModel: model.BaseModel{
					Uuid:       hqGI.Uuid,
					CreateTime: hqGI.CreateTime,
					UpdateTime: hqGI.UpdateTime,
					DeleteTime: hqGI.DeleteTime,
				},
				ProductPackageTakeoutUuid:   storeTakeout.Uuid,
				ProductPackageGroupItemUuid: hqGI.ProductPackageGroupItemUuid,
				ProductPackageGroupUuid:     hqGI.ProductPackageGroupUuid,
				HeadquarterUuid:             hqUuid,
				AddPrice:                    hqGI.AddPrice,
			}
			if err := groupItemTakeoutRepo.CreateProductPackageGroupItemTakeout(&newGI); err != nil {
				logger.Logger.Error("创建子店外卖套餐子商品价格失败",
					zap.Uint64("product_package_group_uuid", hqGI.ProductPackageGroupUuid),
					zap.Uint64("product_package_group_item_uuid", hqGI.ProductPackageGroupItemUuid),
					zap.Error(err))
			}
		}
	}

	return nil
}
