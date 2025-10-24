package repository

import (
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IStockReconciliationRepo 盘点单仓库接口
type IStockReconciliationRepo interface {
	// 盘点单主表操作
	CreateStockReconciliation(stockReconciliation *model.StockReconciliation) error
	UpdateStockReconciliation(stockReconciliation *model.StockReconciliation) error
	UpdateStockReconciliationData(data map[string]any, opts ...DBOption) error
	DeleteStockReconciliation(uuid uint64) error
	GetStockReconciliation(opts ...DBOption) (*model.StockReconciliation, error)
	GetStockReconciliationList(opts ...DBOption) ([]*model.StockReconciliation, error)
	GetStockReconciliationListWithPagination(pageNo, pageSize int, opts ...DBOption) ([]*model.StockReconciliation, int64, error)
	GetStockReconciliationByUuid(uuid uint64) (*model.StockReconciliation, error)
	GetStockReconciliationByOrderNo(orderNo string) (*model.StockReconciliation, error)
	GetMaxOrderNoByPrefix(prefix string) (string, error)
	WithMaterialUnitUnitMultiLanguageName() DBOption
	WithWarehouseMultiLanguageName() DBOption
	WithWarehouse() DBOption
	WithStockReconciliationItems() DBOption
	WithStockReconciliationItemsMultiLanguageName() DBOption
	WithStockReconciliationItemsMaterialBaseUnit() DBOption
	WithStockReconciliationItemMaterialUnits() DBOption

	// 盘点单物品明细操作
	CreateStockReconciliationItem(item *model.StockReconciliationItem) error
	CreateStockReconciliationItemBatch(items []*model.StockReconciliationItem) error
	UpdateStockReconciliationItem(item *model.StockReconciliationItem) error
	DeleteStockReconciliationItem(uuid uint64) error
	DeleteStockReconciliationItemByReconciliationUuid(reconciliationUuid uint64) error
	DeleteStockReconciliationItemUnitByReconciliationUuid(reconciliationUuid uint64) error
	GetStockReconciliationItem(opts ...DBOption) (*model.StockReconciliationItem, error)
	GetStockReconciliationItemList(opts ...DBOption) ([]*model.StockReconciliationItem, error)
	GetStockReconciliationItemByUuid(uuid uint64) (*model.StockReconciliationItem, error)
	GetStockReconciliationItemListByReconciliationUuid(reconciliationUuid uint64) ([]*model.StockReconciliationItem, error)
	GetStockReconciliationItemCountListByReconciliationUuidList(reconciliationUuids []uint64) (map[uint64]int, error)

	// 盘点单物品单位明细操作
	CreateStockReconciliationItemUnit(unit *model.StockReconciliationItemUnit) error
	CreateStockReconciliationItemUnitBatch(units []*model.StockReconciliationItemUnit) error
	UpdateStockReconciliationItemUnit(unit *model.StockReconciliationItemUnit) error
	DeleteStockReconciliationItemUnit(uuid uint64) error
	DeleteStockReconciliationItemUnitByItemUuid(itemUuid uint64) error
	GetStockReconciliationItemUnit(opts ...DBOption) (*model.StockReconciliationItemUnit, error)
	GetStockReconciliationItemUnitList(opts ...DBOption) ([]*model.StockReconciliationItemUnit, error)
	GetStockReconciliationItemUnitByUuid(uuid uint64) (*model.StockReconciliationItemUnit, error)
	GetStockReconciliationItemUnitListByItemUuid(itemUuid uint64) ([]*model.StockReconciliationItemUnit, error)

	// 查询选项
	WhereUuid(uuid uint64) DBOption
	WhereUuids(uuids []uint64) DBOption
	WhereOrderNo(orderNo string) DBOption
	WhereWarehouseUuid(warehouseUuid uint64) DBOption
	WhereWarehouseUuids(warehouseUuids []uint64) DBOption
	WhereKeyword(keyword string) DBOption
	WhereCreateTimeRange(startTime, endTime int) DBOption
	WhereType(reconciliationType int) DBOption
	WherePurpose(purpose int) DBOption
	WhereStatus(status int) DBOption
	WhereStatusIn(statuses []int) DBOption
	WhereReconciliationUuid(reconciliationUuid uint64) DBOption
	WhereItemUuid(itemUuid uint64) DBOption
	WhereMaterialUuid(materialUuid uint64) DBOption
	WhereMaterialUnitUuid(materialUnitUuid uint64) DBOption
}

// NewStockReconciliationRepo 创建新的盘点单仓库
func NewStockReconciliationRepo(db *gorm.DB) IStockReconciliationRepo {
	return NewStockReconciliationRepoImpl(db)
}

// NewStockReconciliationRepoImpl 创建新的盘点单仓库实现
func NewStockReconciliationRepoImpl(db *gorm.DB) *StockReconciliationRepoImpl {
	return &StockReconciliationRepoImpl{db: db}
}

// StockReconciliationRepoImpl 盘点单仓库实现
type StockReconciliationRepoImpl struct {
	db *gorm.DB // 数据库连接
}

// CreateStockReconciliation 创建盘点单
func (r *StockReconciliationRepoImpl) CreateStockReconciliation(stockReconciliation *model.StockReconciliation) error {
	if err := r.db.Create(stockReconciliation).Error; err != nil {
		return errors.WithMessage(err, "创建盘点单失败")
	}
	return nil
}

// UpdateStockReconciliation 更新盘点单
func (r *StockReconciliationRepoImpl) UpdateStockReconciliation(stockReconciliation *model.StockReconciliation) error {
	if err := r.db.Save(stockReconciliation).Error; err != nil {
		return errors.WithMessage(err, "更新盘点单失败")
	}
	return nil
}

// UpdateStockReconciliationData 更新盘点单数据
func (r *StockReconciliationRepoImpl) UpdateStockReconciliationData(data map[string]any, opts ...DBOption) error {
	query := r.db.Model(&model.StockReconciliation{})
	for _, opt := range opts {
		query = opt(query)
	}
	if err := query.Updates(data).Error; err != nil {
		return errors.WithMessage(err, "更新盘点单数据失败")
	}
	return nil
}

// DeleteStockReconciliation 删除盘点单以及物品明细和单位明细（软删除）
func (r *StockReconciliationRepoImpl) DeleteStockReconciliation(uuid uint64) error {
	if err := r.db.Model(&model.StockReconciliation{}).Where("uuid = ?", uuid).Update("delete_time", gorm.Expr("UNIX_TIMESTAMP()")).Error; err != nil {
		return errors.WithMessage(err, "删除盘点单失败")
	}
	if err := r.db.Model(&model.StockReconciliationItem{}).Where("stock_reconciliation_uuid = ?", uuid).Update("delete_time", gorm.Expr("UNIX_TIMESTAMP()")).Error; err != nil {
		return errors.WithMessage(err, "删除盘点单物品明细失败")
	}
	if err := r.db.Model(&model.StockReconciliationItemUnit{}).Where("stock_reconciliation_item_uuid = ?", r.db.Model(&model.StockReconciliationItem{}).Where("stock_reconciliation_uuid = ?", uuid).Select("uuid")).Update("delete_time", gorm.Expr("UNIX_TIMESTAMP()")).Error; err != nil {
		return errors.WithMessage(err, "删除盘点单物品单位明细失败")
	}
	return nil
}

// GetStockReconciliation 获取单个盘点单
func (r *StockReconciliationRepoImpl) GetStockReconciliation(opts ...DBOption) (*model.StockReconciliation, error) {
	var stockReconciliation model.StockReconciliation
	query := r.db.Scopes(NotDeleted)
	for _, opt := range opts {
		query = opt(query)
	}
	if err := query.First(&stockReconciliation).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, errors.WithMessage(err, "获取盘点单失败")
	}
	return &stockReconciliation, nil
}

// GetStockReconciliationList 获取盘点单列表
func (r *StockReconciliationRepoImpl) GetStockReconciliationList(opts ...DBOption) ([]*model.StockReconciliation, error) {
	var list []*model.StockReconciliation
	query := r.db.Scopes(NotDeleted)
	for _, opt := range opts {
		query = opt(query)
	}
	if err := query.Order("create_time DESC").Find(&list).Error; err != nil {
		return nil, errors.WithMessage(err, "获取盘点单列表失败")
	}
	return list, nil
}

// GetStockReconciliationListWithPagination 获取盘点单列表（分页）
func (r *StockReconciliationRepoImpl) GetStockReconciliationListWithPagination(pageNo, pageSize int, opts ...DBOption) ([]*model.StockReconciliation, int64, error) {
	var list []*model.StockReconciliation
	var total int64

	query := r.db.Model(&model.StockReconciliation{}).Scopes(NotDeleted)
	for _, opt := range opts {
		query = opt(query)
	}

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errors.WithMessage(err, "计算盘点单总数失败")
	}

	// 分页查询
	offset := (pageNo - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("create_time DESC").Find(&list).Error; err != nil {
		return nil, 0, errors.WithMessage(err, "查询盘点单列表失败")
	}

	return list, total, nil
}

// GetStockReconciliationByUuid 根据UUID获取盘点单
func (r *StockReconciliationRepoImpl) GetStockReconciliationByUuid(uuid uint64) (*model.StockReconciliation, error) {
	return r.GetStockReconciliation(r.WhereUuid(uuid))
}

// GetStockReconciliationByOrderNo 根据单据编号获取盘点单
func (r *StockReconciliationRepoImpl) GetStockReconciliationByOrderNo(orderNo string) (*model.StockReconciliation, error) {
	return r.GetStockReconciliation(r.WhereOrderNo(orderNo))
}

// GetMaxOrderNoByPrefix 根据前缀获取最大的单据编号
func (r *StockReconciliationRepoImpl) GetMaxOrderNoByPrefix(prefix string) (string, error) {
	var maxOrderNo string
	err := r.db.Model(&model.StockReconciliation{}).
		Where("order_no LIKE ?", prefix+"%").
		Order("order_no DESC").
		Limit(1).
		Pluck("order_no", &maxOrderNo).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", nil
		}
		return "", errors.WithMessage(err, "查询最大单据编号失败")
	}

	return maxOrderNo, nil
}

func (r *StockReconciliationRepoImpl) WithMaterialUnitUnitMultiLanguageName() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("MaterialUnit.Unit.MultiLanguageName")
	}
}

func (r *StockReconciliationRepoImpl) WithWarehouseMultiLanguageName() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("Warehouse.MultiLanguageName")
	}
}

func (r *StockReconciliationRepoImpl) WithWarehouse() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("Warehouse")
	}
}

func (r *StockReconciliationRepoImpl) WithStockReconciliationItems() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("StockReconciliationItems.Material")
	}
}

func (r *StockReconciliationRepoImpl) WithStockReconciliationItemsMultiLanguageName() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("StockReconciliationItems.Material.MultiLanguageName")
	}
}

func (r *StockReconciliationRepoImpl) WithStockReconciliationItemsMaterialBaseUnit() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("StockReconciliationItems.Material.Unit")
	}
}

func (r *StockReconciliationRepoImpl) WithStockReconciliationItemMaterialUnits() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("StockReconciliationItems.Material.NotBaseUnitList.Unit.MultiLanguageName")
	}
}

// CreateStockReconciliationItem 创建盘点单物品明细
func (r *StockReconciliationRepoImpl) CreateStockReconciliationItem(item *model.StockReconciliationItem) error {
	if err := r.db.Create(item).Error; err != nil {
		return errors.WithMessage(err, "创建盘点单物品明细失败")
	}
	return nil
}

// CreateStockReconciliationItemBatch 批量创建盘点单物品明细
func (r *StockReconciliationRepoImpl) CreateStockReconciliationItemBatch(items []*model.StockReconciliationItem) error {
	if len(items) == 0 {
		return nil
	}
	if err := r.db.Create(&items).Error; err != nil {
		return errors.WithMessage(err, "批量创建盘点单物品明细失败")
	}
	return nil
}

// UpdateStockReconciliationItem 更新盘点单物品明细
func (r *StockReconciliationRepoImpl) UpdateStockReconciliationItem(item *model.StockReconciliationItem) error {
	if err := r.db.Save(item).Error; err != nil {
		return errors.WithMessage(err, "更新盘点单物品明细失败")
	}
	return nil
}

// DeleteStockReconciliationItem 删除盘点单物品明细（软删除）
func (r *StockReconciliationRepoImpl) DeleteStockReconciliationItem(uuid uint64) error {
	if err := r.db.Model(&model.StockReconciliationItem{}).Where("uuid = ?", uuid).Update("delete_time", gorm.Expr("UNIX_TIMESTAMP()")).Error; err != nil {
		return errors.WithMessage(err, "删除盘点单物品明细失败")
	}
	return nil
}

// DeleteStockReconciliationItemByReconciliationUuid 根据盘点单UUID删除盘点单物品明细
func (r *StockReconciliationRepoImpl) DeleteStockReconciliationItemByReconciliationUuid(reconciliationUuid uint64) error {
	if err := r.db.Model(&model.StockReconciliationItem{}).Where("stock_reconciliation_uuid = ?", reconciliationUuid).Update("delete_time", gorm.Expr("UNIX_TIMESTAMP()")).Error; err != nil {
		return errors.WithMessage(err, "删除盘点单物品明细失败")
	}
	return nil
}

// ByReconciliationUuid 根据盘点单物品明细UUID删除单位明细
func (r *StockReconciliationRepoImpl) DeleteStockReconciliationItemUnitByReconciliationUuid(reconciliationUuid uint64) error {
	if err := r.db.Model(&model.StockReconciliationItemUnit{}).
		Where("stock_reconciliation_item_uuid IN (?)", r.db.Model(&model.StockReconciliationItem{}).
			Where("stock_reconciliation_uuid = ?", reconciliationUuid).Select("uuid")).
		Update("delete_time", gorm.Expr("UNIX_TIMESTAMP()")).Error; err != nil {
		return errors.WithMessage(err, "删除盘点单物品单位明细失败")
	}
	return nil
}

// GetStockReconciliationItem 获取单个盘点单物品明细
func (r *StockReconciliationRepoImpl) GetStockReconciliationItem(opts ...DBOption) (*model.StockReconciliationItem, error) {
	var item model.StockReconciliationItem
	query := r.db.Scopes(NotDeleted)
	for _, opt := range opts {
		query = opt(query)
	}
	if err := query.First(&item).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, errors.WithMessage(err, "获取盘点单物品明细失败")
	}
	return &item, nil
}

// GetStockReconciliationItemList 获取盘点单物品明细列表
func (r *StockReconciliationRepoImpl) GetStockReconciliationItemList(opts ...DBOption) ([]*model.StockReconciliationItem, error) {
	var list []*model.StockReconciliationItem
	query := r.db.Scopes(NotDeleted)
	for _, opt := range opts {
		query = opt(query)
	}
	if err := query.Find(&list).Error; err != nil {
		return nil, errors.WithMessage(err, "获取盘点单物品明细列表失败")
	}
	return list, nil
}

// GetStockReconciliationItemByUuid 根据UUID获取盘点单物品明细
func (r *StockReconciliationRepoImpl) GetStockReconciliationItemByUuid(uuid uint64) (*model.StockReconciliationItem, error) {
	return r.GetStockReconciliationItem(r.WhereUuid(uuid))
}

// GetStockReconciliationItemListByReconciliationUuid 根据盘点单UUID获取物品明细列表
func (r *StockReconciliationRepoImpl) GetStockReconciliationItemListByReconciliationUuid(reconciliationUuid uint64) ([]*model.StockReconciliationItem, error) {
	return r.GetStockReconciliationItemList(r.WhereReconciliationUuid(reconciliationUuid))
}

// GetStockReconciliationItemCountListByReconciliationUuidList 根据盘点单UUID列表获取每个盘点单的物品数量
func (r *StockReconciliationRepoImpl) GetStockReconciliationItemCountListByReconciliationUuidList(reconciliationUuids []uint64) (map[uint64]int, error) {
	// 如果列表为空，直接返回空map
	if len(reconciliationUuids) == 0 {
		return make(map[uint64]int), nil
	}

	// 定义查询结果结构
	type CountResult struct {
		StockReconciliationUuid uint64 `gorm:"column:stock_reconciliation_uuid"`
		Count                   int    `gorm:"column:count"`
	}

	// 执行分组统计查询
	var results []CountResult
	err := r.db.Model(&model.StockReconciliationItem{}).
		Select("stock_reconciliation_uuid, COUNT(*) as count").
		Where("stock_reconciliation_uuid IN ?", reconciliationUuids).
		Where("delete_time = ?", 0).
		Group("stock_reconciliation_uuid").
		Scan(&results).Error

	if err != nil {
		return nil, errors.WithMessage(err, "查询盘点单物品数量失败")
	}

	// 转换为map
	countMap := make(map[uint64]int)
	for _, result := range results {
		countMap[result.StockReconciliationUuid] = result.Count
	}

	// 为没有物品的盘点单补充0值
	for _, uuid := range reconciliationUuids {
		if _, exists := countMap[uuid]; !exists {
			countMap[uuid] = 0
		}
	}

	return countMap, nil
}

// CreateStockReconciliationItemUnit 创建盘点单物品单位明细
func (r *StockReconciliationRepoImpl) CreateStockReconciliationItemUnit(unit *model.StockReconciliationItemUnit) error {
	if err := r.db.Create(unit).Error; err != nil {
		return errors.WithMessage(err, "创建盘点单物品单位明细失败")
	}
	return nil
}

// CreateStockReconciliationItemUnitBatch 批量创建盘点单物品单位明细
func (r *StockReconciliationRepoImpl) CreateStockReconciliationItemUnitBatch(units []*model.StockReconciliationItemUnit) error {
	if len(units) == 0 {
		return nil
	}
	if err := r.db.Create(&units).Error; err != nil {
		return errors.WithMessage(err, "批量创建盘点单物品单位明细失败")
	}
	return nil
}

// UpdateStockReconciliationItemUnit 更新盘点单物品单位明细
func (r *StockReconciliationRepoImpl) UpdateStockReconciliationItemUnit(unit *model.StockReconciliationItemUnit) error {
	if err := r.db.Save(unit).Error; err != nil {
		return errors.WithMessage(err, "更新盘点单物品单位明细失败")
	}
	return nil
}

// DeleteStockReconciliationItemUnit 删除盘点单物品单位明细（软删除）
func (r *StockReconciliationRepoImpl) DeleteStockReconciliationItemUnit(uuid uint64) error {
	if err := r.db.Model(&model.StockReconciliationItemUnit{}).Where("uuid = ?", uuid).Update("delete_time", gorm.Expr("UNIX_TIMESTAMP()")).Error; err != nil {
		return errors.WithMessage(err, "删除盘点单物品单位明细失败")
	}
	return nil
}

// DeleteStockReconciliationItemUnitByItemUuid 根据盘点单物品明细UUID删除单位明细
func (r *StockReconciliationRepoImpl) DeleteStockReconciliationItemUnitByItemUuid(itemUuid uint64) error {
	if err := r.db.Model(&model.StockReconciliationItemUnit{}).Where("stock_reconciliation_item_uuid = ?", itemUuid).Update("delete_time", gorm.Expr("UNIX_TIMESTAMP()")).Error; err != nil {
		return errors.WithMessage(err, "删除盘点单物品单位明细失败")
	}
	return nil
}

// GetStockReconciliationItemUnit 获取单个盘点单物品单位明细
func (r *StockReconciliationRepoImpl) GetStockReconciliationItemUnit(opts ...DBOption) (*model.StockReconciliationItemUnit, error) {
	var unit model.StockReconciliationItemUnit
	query := r.db.Scopes(NotDeleted)
	for _, opt := range opts {
		query = opt(query)
	}
	if err := query.First(&unit).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, errors.WithMessage(err, "获取盘点单物品单位明细失败")
	}
	return &unit, nil
}

// GetStockReconciliationItemUnitList 获取盘点单物品单位明细列表
func (r *StockReconciliationRepoImpl) GetStockReconciliationItemUnitList(opts ...DBOption) ([]*model.StockReconciliationItemUnit, error) {
	var list []*model.StockReconciliationItemUnit
	query := r.db.Scopes(NotDeleted)
	for _, opt := range opts {
		query = opt(query)
	}
	if err := query.Find(&list).Error; err != nil {
		return nil, errors.WithMessage(err, "获取盘点单物品单位明细列表失败")
	}
	return list, nil
}

// GetStockReconciliationItemUnitByUuid 根据UUID获取盘点单物品单位明细
func (r *StockReconciliationRepoImpl) GetStockReconciliationItemUnitByUuid(uuid uint64) (*model.StockReconciliationItemUnit, error) {
	return r.GetStockReconciliationItemUnit(r.WhereUuid(uuid))
}

// GetStockReconciliationItemUnitListByItemUuid 根据盘点单物品明细UUID获取单位明细列表
func (r *StockReconciliationRepoImpl) GetStockReconciliationItemUnitListByItemUuid(itemUuid uint64) ([]*model.StockReconciliationItemUnit, error) {
	return r.GetStockReconciliationItemUnitList(r.WhereItemUuid(itemUuid), r.WithMaterialUnitUnitMultiLanguageName())
}

// WhereUuid UUID查询选项
func (r *StockReconciliationRepoImpl) WhereUuid(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("uuid = ?", uuid)
	}
}

// WhereUuids UUID列表查询选项
func (r *StockReconciliationRepoImpl) WhereUuids(uuids []uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("uuid IN ?", uuids)
	}
}

// WhereOrderNo 单据编号查询选项
func (r *StockReconciliationRepoImpl) WhereOrderNo(orderNo string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("order_no = ?", orderNo)
	}
}

// WhereWarehouseUuid 仓库UUID查询选项
func (r *StockReconciliationRepoImpl) WhereWarehouseUuid(warehouseUuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("warehouse_uuid = ?", warehouseUuid)
	}
}

// WhereWarehouseUuids 仓库UUID列表查询选项
func (r *StockReconciliationRepoImpl) WhereWarehouseUuids(warehouseUuids []uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		if len(warehouseUuids) > 0 {
			return db.Where("warehouse_uuid IN ?", warehouseUuids)
		}
		return db
	}
}

// WhereKeyword 关键字查询选项（搜索单据编号和ERP盘点单号）
func (r *StockReconciliationRepoImpl) WhereKeyword(keyword string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		if keyword != "" {
			return db.Where("order_no LIKE ? OR erp_code LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
		}
		return db
	}
}

// WhereCreateTimeRange 创建时间范围查询选项
func (r *StockReconciliationRepoImpl) WhereCreateTimeRange(startTime, endTime int) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		if startTime > 0 && endTime > 0 {
			return db.Where("create_time >= ? AND create_time <= ?", startTime, endTime)
		} else if startTime > 0 {
			return db.Where("create_time >= ?", startTime)
		} else if endTime > 0 {
			return db.Where("create_time <= ?", endTime)
		}
		return db
	}
}

// WhereType 盘点类型查询选项
func (r *StockReconciliationRepoImpl) WhereType(reconciliationType int) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("type = ?", reconciliationType)
	}
}

// WherePurpose 盘点目的查询选项
func (r *StockReconciliationRepoImpl) WherePurpose(purpose int) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("purpose = ?", purpose)
	}
}

// WhereStatus 状态查询选项
func (r *StockReconciliationRepoImpl) WhereStatus(status int) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("status = ?", status)
	}
}

// WhereStatusIn 状态列表查询选项
func (r *StockReconciliationRepoImpl) WhereStatusIn(statuses []int) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		if len(statuses) > 0 {
			return db.Where("status IN ?", statuses)
		}
		return db
	}
}

// WhereReconciliationUuid 盘点单UUID查询选项
func (r *StockReconciliationRepoImpl) WhereReconciliationUuid(reconciliationUuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("stock_reconciliation_uuid = ?", reconciliationUuid)
	}
}

// WhereItemUuid 盘点单物品明细UUID查询选项
func (r *StockReconciliationRepoImpl) WhereItemUuid(itemUuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("stock_reconciliation_item_uuid = ?", itemUuid)
	}
}

// WhereMaterialUuid 物品UUID查询选项
func (r *StockReconciliationRepoImpl) WhereMaterialUuid(materialUuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("material_uuid = ?", materialUuid)
	}
}

// WhereMaterialUnitUuid 物品单位UUID查询选项
func (r *StockReconciliationRepoImpl) WhereMaterialUnitUuid(materialUnitUuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("material_unit_uuid = ?", materialUnitUuid)
	}
}
