package repository

import (
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IStockLossRepo 报损单仓库接口
type IStockLossRepo interface {
	// 报损单主表操作
	CreateStockLoss(stockLoss *model.StockLoss) error
	UpdateStockLoss(stockLoss *model.StockLoss) error
	UpdateData(data map[string]any, opts ...DBOption) error
	DeleteStockLoss(uuid uint64) error
	GetStockLossByUuid(uuid uint64) (*model.StockLoss, error)
	GetStockLoss(opts ...DBOption) (*model.StockLoss, error)
	GetList(opts ...DBOption) ([]*model.StockLoss, error)
	GetStockLossListWithPagination(pageNo, pageSize int, opts ...DBOption) ([]*model.StockLoss, int64, error)
	GetMaxCodeByPrefix(prefix string) (string, error)

	// 明细操作
	CreateItem(item *model.StockLossItem) error
	CreateItems(items []*model.StockLossItem) error
	DeleteItemsByStockLossUuid(stockLossUuid uint64) error
	DeleteItemsByUuids(uuids []uint64) error
	GetItemsByStockLossUuid(stockLossUuid uint64) ([]*model.StockLossItem, error)
	GetStockLossItemCountByUuids(stockLossUuids []uint64) (map[uint64]int, error)

	// 物品单位明细操作
	CreateItemUnits(itemUnits []*model.StockLossItemUnit) error
	DeleteItemUnitsByUuids(uuids []uint64) error
	DeleteItemUnitsByStockLossItemUuids(stockLossItemUuids []uint64) error
	DeleteItemUnitsByStockLossUuid(stockLossUuid uint64) error
	GetItemUnitsByStockLossItemUuids(stockLossItemUuids []uint64) ([]*model.StockLossItemUnit, error)

	// 批注操作
	CreateAnnotation(annotation *model.StockLossAnnotation) error
	GetAnnotationsByStockLossUuid(stockLossUuid uint64) ([]*model.StockLossAnnotation, error)

	// 附件操作
	CreateFile(file *model.StockLossFile) error
	CreateFiles(files []*model.StockLossFile) error
	DeleteFilesByStockLossUuid(stockLossUuid uint64) error
	GetFilesByStockLossUuid(stockLossUuid uint64) ([]*model.StockLossFile, error)

	// 查询选项
	WhereUuid(uuid uint64) DBOption
	WhereUuids(uuids []uint64) DBOption
	WhereCode(code string) DBOption
	WhereWarehouseUuid(warehouseUuid uint64) DBOption
	WhereWarehouseUuids(warehouseUuids []uint64) DBOption
	WhereKeyword(keyword string) DBOption
	WhereCreateTimeRange(startTime, endTime int) DBOption
	WhereLossType(lossType int) DBOption
	WhereStatus(status int) DBOption
	WhereStatusIn(statuses []int) DBOption
	WhereStockLossUuid(stockLossUuid uint64) DBOption
	OrderByCreateTimeDesc() DBOption

	// 预加载选项
	WithWarehouse() DBOption
	WithItems() DBOption
	WithItemsMaterial() DBOption
	WithItemsMaterialNotBaseUnitList() DBOption
	WithItemsItemUnits() DBOption
	WithItemsItemUnitsMaterialUnit() DBOption
	WithFiles() DBOption
	WithFilesFile() DBOption
	WithAnnotations() DBOption
}

// NewStockLossRepo 创建新的报损单仓库
func NewStockLossRepo(db *gorm.DB) IStockLossRepo {
	return &StockLossRepoImpl{db: db}
}

// StockLossRepoImpl 报损单仓库实现
type StockLossRepoImpl struct {
	db *gorm.DB
}

// CreateStockLoss 创建报损单
func (r *StockLossRepoImpl) CreateStockLoss(stockLoss *model.StockLoss) error {
	if err := r.db.Create(stockLoss).Error; err != nil {
		return errors.WithMessage(err, "创建报损单失败")
	}
	return nil
}

// UpdateStockLoss 更新报损单
func (r *StockLossRepoImpl) UpdateStockLoss(stockLoss *model.StockLoss) error {
	if err := r.db.Save(stockLoss).Error; err != nil {
		return errors.WithMessage(err, "更新报损单失败")
	}
	return nil
}

// UpdateData 更新报损单数据
func (r *StockLossRepoImpl) UpdateData(data map[string]any, opts ...DBOption) error {
	query := r.db.Model(&model.StockLoss{})
	for _, opt := range opts {
		query = opt(query)
	}
	if err := query.Updates(data).Error; err != nil {
		return errors.WithMessage(err, "更新报损单数据失败")
	}
	return nil
}

// DeleteStockLoss 删除报损单及其明细、单位明细、附件（软删除）
func (r *StockLossRepoImpl) DeleteStockLoss(uuid uint64) error {
	if err := r.db.Model(&model.StockLoss{}).Where("uuid = ?", uuid).Update("delete_time", gorm.Expr("UNIX_TIMESTAMP()")).Error; err != nil {
		return errors.WithMessage(err, "删除报损单失败")
	}
	// 删除物品单位明细（使用子查询）
	subQuery := r.db.Model(&model.StockLossItem{}).
		Select("uuid").
		Where("stock_loss_uuid = ?", uuid)
	if err := r.db.Model(&model.StockLossItemUnit{}).
		Where("stock_loss_item_uuid IN (?)", subQuery).
		Update("delete_time", gorm.Expr("UNIX_TIMESTAMP()")).Error; err != nil {
		return errors.WithMessage(err, "删除报损单物品单位明细失败")
	}
	// 删除物品明细
	if err := r.db.Model(&model.StockLossItem{}).Where("stock_loss_uuid = ?", uuid).Update("delete_time", gorm.Expr("UNIX_TIMESTAMP()")).Error; err != nil {
		return errors.WithMessage(err, "删除报损单明细失败")
	}
	if err := r.db.Model(&model.StockLossFile{}).Where("stock_loss_uuid = ?", uuid).Update("delete_time", gorm.Expr("UNIX_TIMESTAMP()")).Error; err != nil {
		return errors.WithMessage(err, "删除报损单附件关联失败")
	}
	return nil
}

// GetStockLossByUuid 根据UUID获取报损单
func (r *StockLossRepoImpl) GetStockLossByUuid(uuid uint64) (*model.StockLoss, error) {
	return r.GetStockLoss(r.WhereUuid(uuid))
}

// GetStockLoss 获取单个报损单
func (r *StockLossRepoImpl) GetStockLoss(opts ...DBOption) (*model.StockLoss, error) {
	var stockLoss model.StockLoss
	query := r.db.Scopes(NotDeleted)
	for _, opt := range opts {
		query = opt(query)
	}
	if err := query.First(&stockLoss).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, errors.WithMessage(err, "获取报损单失败")
	}
	return &stockLoss, nil
}

// GetList 获取报损单列表
func (r *StockLossRepoImpl) GetList(opts ...DBOption) ([]*model.StockLoss, error) {
	var list []*model.StockLoss
	query := r.db.Scopes(NotDeleted)
	for _, opt := range opts {
		query = opt(query)
	}
	if err := query.Find(&list).Error; err != nil {
		return nil, errors.WithMessage(err, "获取报损单列表失败")
	}
	return list, nil
}

// GetStockLossListWithPagination 获取报损单列表（分页）
func (r *StockLossRepoImpl) GetStockLossListWithPagination(pageNo, pageSize int, opts ...DBOption) ([]*model.StockLoss, int64, error) {
	var list []*model.StockLoss
	var total int64

	query := r.db.Model(&model.StockLoss{}).Scopes(NotDeleted)
	for _, opt := range opts {
		query = opt(query)
	}

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errors.WithMessage(err, "计算报损单总数失败")
	}

	// 分页查询
	offset := (pageNo - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, errors.WithMessage(err, "查询报损单列表失败")
	}

	return list, total, nil
}

// GetMaxCodeByPrefix 根据前缀获取最大的单据编号
func (r *StockLossRepoImpl) GetMaxCodeByPrefix(prefix string) (string, error) {
	var maxCode string
	err := r.db.Model(&model.StockLoss{}).
		Where("code LIKE ?", prefix+"%").
		Order("code DESC").
		Limit(1).
		Pluck("code", &maxCode).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", nil
		}
		return "", errors.WithMessage(err, "查询最大单据编号失败")
	}

	return maxCode, nil
}

// CreateItem 创建单个报损单明细
func (r *StockLossRepoImpl) CreateItem(item *model.StockLossItem) error {
	if err := r.db.Create(item).Error; err != nil {
		return errors.WithMessage(err, "创建报损单明细失败")
	}
	return nil
}

// CreateItems 批量创建报损单明细
func (r *StockLossRepoImpl) CreateItems(items []*model.StockLossItem) error {
	if len(items) == 0 {
		return nil
	}
	if err := r.db.Create(&items).Error; err != nil {
		return errors.WithMessage(err, "批量创建报损单明细失败")
	}
	return nil
}

// DeleteItemsByStockLossUuid 根据报损单UUID删除明细（软删除）
func (r *StockLossRepoImpl) DeleteItemsByStockLossUuid(stockLossUuid uint64) error {
	if err := r.db.Model(&model.StockLossItem{}).
		Where("stock_loss_uuid = ?", stockLossUuid).
		Scopes(NotDeleted).
		Update("delete_time", gorm.Expr("UNIX_TIMESTAMP()")).Error; err != nil {
		return errors.WithMessage(err, "删除报损单明细失败")
	}
	return nil
}

// DeleteItemsByUuids 根据UUID列表删除物品明细（软删除）
func (r *StockLossRepoImpl) DeleteItemsByUuids(uuids []uint64) error {
	if len(uuids) == 0 {
		return nil
	}
	if err := r.db.Model(&model.StockLossItem{}).
		Where("uuid IN ?", uuids).
		Scopes(NotDeleted).
		Update("delete_time", gorm.Expr("UNIX_TIMESTAMP()")).Error; err != nil {
		return errors.WithMessage(err, "删除报损单物品明细失败")
	}
	return nil
}

// GetItemsByStockLossUuid 根据报损单UUID获取明细列表
func (r *StockLossRepoImpl) GetItemsByStockLossUuid(stockLossUuid uint64) ([]*model.StockLossItem, error) {
	var list []*model.StockLossItem
	if err := r.db.Scopes(NotDeleted).
		Where("stock_loss_uuid = ?", stockLossUuid).
		Find(&list).Error; err != nil {
		return nil, errors.WithMessage(err, "获取报损单明细列表失败")
	}
	return list, nil
}

// GetStockLossItemCountByUuids 根据报损单UUID列表获取每个报损单的物品数量
func (r *StockLossRepoImpl) GetStockLossItemCountByUuids(stockLossUuids []uint64) (map[uint64]int, error) {
	if len(stockLossUuids) == 0 {
		return make(map[uint64]int), nil
	}

	type CountResult struct {
		StockLossUuid uint64 `gorm:"column:stock_loss_uuid"`
		Count         int    `gorm:"column:count"`
	}

	var results []CountResult
	err := r.db.Model(&model.StockLossItem{}).
		Select("stock_loss_uuid, COUNT(*) as count").
		Where("stock_loss_uuid IN ?", stockLossUuids).
		Where("delete_time = ?", 0).
		Group("stock_loss_uuid").
		Scan(&results).Error

	if err != nil {
		return nil, errors.WithMessage(err, "查询报损单物品数量失败")
	}

	countMap := make(map[uint64]int)
	for _, result := range results {
		countMap[result.StockLossUuid] = result.Count
	}

	for _, uuid := range stockLossUuids {
		if _, exists := countMap[uuid]; !exists {
			countMap[uuid] = 0
		}
	}

	return countMap, nil
}

// CreateItemUnits 批量创建报损单物品单位明细
func (r *StockLossRepoImpl) CreateItemUnits(itemUnits []*model.StockLossItemUnit) error {
	if len(itemUnits) == 0 {
		return nil
	}
	if err := r.db.Create(&itemUnits).Error; err != nil {
		return errors.WithMessage(err, "批量创建报损单物品单位明细失败")
	}
	return nil
}

// DeleteItemUnitsByUuids 根据UUID列表删除单位明细（软删除）
func (r *StockLossRepoImpl) DeleteItemUnitsByUuids(uuids []uint64) error {
	if len(uuids) == 0 {
		return nil
	}
	if err := r.db.Model(&model.StockLossItemUnit{}).
		Where("uuid IN ?", uuids).
		Scopes(NotDeleted).
		Update("delete_time", gorm.Expr("UNIX_TIMESTAMP()")).Error; err != nil {
		return errors.WithMessage(err, "删除报损单物品单位明细失败")
	}
	return nil
}

// DeleteItemUnitsByStockLossItemUuids 根据报损单物品明细UUID列表删除单位明细（软删除）
func (r *StockLossRepoImpl) DeleteItemUnitsByStockLossItemUuids(stockLossItemUuids []uint64) error {
	if len(stockLossItemUuids) == 0 {
		return nil
	}
	if err := r.db.Model(&model.StockLossItemUnit{}).
		Where("stock_loss_item_uuid IN ?", stockLossItemUuids).
		Scopes(NotDeleted).
		Update("delete_time", gorm.Expr("UNIX_TIMESTAMP()")).Error; err != nil {
		return errors.WithMessage(err, "删除报损单物品单位明细失败")
	}
	return nil
}

// DeleteItemUnitsByStockLossUuid 根据报损单UUID删除物品单位明细（软删除）
func (r *StockLossRepoImpl) DeleteItemUnitsByStockLossUuid(stockLossUuid uint64) error {
	// 使用子查询删除关联的物品单位明细
	subQuery := r.db.Model(&model.StockLossItem{}).
		Select("uuid").
		Where("stock_loss_uuid = ?", stockLossUuid).
		Where("delete_time = ?", 0)

	if err := r.db.Model(&model.StockLossItemUnit{}).
		Where("stock_loss_item_uuid IN (?)", subQuery).
		Scopes(NotDeleted).
		Update("delete_time", gorm.Expr("UNIX_TIMESTAMP()")).Error; err != nil {
		return errors.WithMessage(err, "删除报损单物品单位明细失败")
	}
	return nil
}

// GetItemUnitsByStockLossItemUuids 根据报损单物品明细UUID列表获取单位明细
func (r *StockLossRepoImpl) GetItemUnitsByStockLossItemUuids(stockLossItemUuids []uint64) ([]*model.StockLossItemUnit, error) {
	if len(stockLossItemUuids) == 0 {
		return make([]*model.StockLossItemUnit, 0), nil
	}
	var list []*model.StockLossItemUnit
	if err := r.db.Scopes(NotDeleted).
		Where("stock_loss_item_uuid IN ?", stockLossItemUuids).
		Find(&list).Error; err != nil {
		return nil, errors.WithMessage(err, "获取报损单物品单位明细列表失败")
	}
	return list, nil
}

// CreateAnnotation 创建批注
func (r *StockLossRepoImpl) CreateAnnotation(annotation *model.StockLossAnnotation) error {
	if err := r.db.Create(annotation).Error; err != nil {
		return errors.WithMessage(err, "创建报损单批注失败")
	}
	return nil
}

// GetAnnotationsByStockLossUuid 根据报损单UUID获取批注列表（按创建时间倒序）
func (r *StockLossRepoImpl) GetAnnotationsByStockLossUuid(stockLossUuid uint64) ([]*model.StockLossAnnotation, error) {
	var list []*model.StockLossAnnotation
	if err := r.db.Scopes(NotDeleted).
		Where("stock_loss_uuid = ?", stockLossUuid).
		Order("create_time DESC").
		Find(&list).Error; err != nil {
		return nil, errors.WithMessage(err, "获取报损单批注列表失败")
	}
	return list, nil
}

// CreateFile 创建单个附件关联
func (r *StockLossRepoImpl) CreateFile(file *model.StockLossFile) error {
	if err := r.db.Create(file).Error; err != nil {
		return errors.WithMessage(err, "创建报损单附件关联失败")
	}
	return nil
}

// CreateFiles 批量创建附件关联
func (r *StockLossRepoImpl) CreateFiles(files []*model.StockLossFile) error {
	if len(files) == 0 {
		return nil
	}
	if err := r.db.Create(&files).Error; err != nil {
		return errors.WithMessage(err, "批量创建报损单附件关联失败")
	}
	return nil
}

// DeleteFilesByStockLossUuid 根据报损单UUID删除附件关联（硬删除）
func (r *StockLossRepoImpl) DeleteFilesByStockLossUuid(stockLossUuid uint64) error {
	if err := r.db.Where("stock_loss_uuid = ?", stockLossUuid).
		Delete(&model.StockLossFile{}).Error; err != nil {
		return errors.WithMessage(err, "删除报损单附件关联失败")
	}
	return nil
}

// GetFilesByStockLossUuid 根据报损单UUID获取附件关联列表
func (r *StockLossRepoImpl) GetFilesByStockLossUuid(stockLossUuid uint64) ([]*model.StockLossFile, error) {
	var list []*model.StockLossFile
	if err := r.db.Scopes(NotDeleted).
		Where("stock_loss_uuid = ?", stockLossUuid).
		Order("sort_order ASC").
		Find(&list).Error; err != nil {
		return nil, errors.WithMessage(err, "获取报损单附件列表失败")
	}
	return list, nil
}

// WhereUuid UUID查询选项
func (r *StockLossRepoImpl) WhereUuid(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("uuid = ?", uuid)
	}
}

// WhereUuids UUID列表查询选项
func (r *StockLossRepoImpl) WhereUuids(uuids []uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("uuid IN ?", uuids)
	}
}

// WhereCode 单据编号查询选项
func (r *StockLossRepoImpl) WhereCode(code string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("code = ?", code)
	}
}

// WhereWarehouseUuid 仓库UUID查询选项
func (r *StockLossRepoImpl) WhereWarehouseUuid(warehouseUuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("warehouse_uuid = ?", warehouseUuid)
	}
}

// WhereWarehouseUuids 仓库UUID列表查询选项
func (r *StockLossRepoImpl) WhereWarehouseUuids(warehouseUuids []uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		if len(warehouseUuids) > 0 {
			return db.Where("warehouse_uuid IN ?", warehouseUuids)
		}
		return db
	}
}

// WhereKeyword 关键字查询选项（搜索单据编号和ERP单据编号）
func (r *StockLossRepoImpl) WhereKeyword(keyword string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		if keyword != "" {
			return db.Where("code LIKE ? OR erp_code LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
		}
		return db
	}
}

// WhereCreateTimeRange 创建时间范围查询选项
func (r *StockLossRepoImpl) WhereCreateTimeRange(startTime, endTime int) DBOption {
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

// WhereLossType 报损类型查询选项
func (r *StockLossRepoImpl) WhereLossType(lossType int) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("loss_type = ?", lossType)
	}
}

// WhereStatus 状态查询选项
func (r *StockLossRepoImpl) WhereStatus(status int) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("status = ?", status)
	}
}

// WhereStatusIn 状态列表查询选项
func (r *StockLossRepoImpl) WhereStatusIn(statuses []int) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		if len(statuses) > 0 {
			return db.Where("status IN ?", statuses)
		}
		return db
	}
}

// WhereStockLossUuid 报损单UUID查询选项
func (r *StockLossRepoImpl) WhereStockLossUuid(stockLossUuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("stock_loss_uuid = ?", stockLossUuid)
	}
}

// OrderByCreateTimeDesc 按创建时间倒序
func (r *StockLossRepoImpl) OrderByCreateTimeDesc() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Order("create_time DESC")
	}
}

// WithWarehouse 预加载仓库（包含多语言名称）
func (r *StockLossRepoImpl) WithWarehouse() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("Warehouse.MultiLanguageName")
	}
}

// WithItems 预加载明细
func (r *StockLossRepoImpl) WithItems() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("StockLossItems", "delete_time = 0")
	}
}

// WithItemsMaterial 预加载明细及物料
func (r *StockLossRepoImpl) WithItemsMaterial() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("StockLossItems.Material", "delete_time = 0")
	}
}

// WithItemsMaterialNotBaseUnitList 预加载明细物料的非基准单位列表（包含单位多语言名称）
func (r *StockLossRepoImpl) WithItemsMaterialNotBaseUnitList() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.
			Preload("StockLossItems.Material.NotBaseUnitList", "delete_time = 0").
			Preload("StockLossItems.Material.NotBaseUnitList.Unit.MultiLanguageName")
	}
}

// WithItemsItemUnits 预加载明细的单位明细
func (r *StockLossRepoImpl) WithItemsItemUnits() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("StockLossItems.StockLossItemUnits", "delete_time = 0")
	}
}

// WithItemsItemUnitsMaterialUnit 预加载明细的单位明细及物料单位
func (r *StockLossRepoImpl) WithItemsItemUnitsMaterialUnit() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("StockLossItems.StockLossItemUnits.MaterialUnit", "delete_time = 0")
	}
}

// WithFiles 预加载附件关联
func (r *StockLossRepoImpl) WithFiles() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("StockLossFiles", "delete_time = 0")
	}
}

// WithFilesFile 预加载附件关联及文件
func (r *StockLossRepoImpl) WithFilesFile() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("StockLossFiles.File", "delete_time = 0")
	}
}

// WithAnnotations 预加载批注
func (r *StockLossRepoImpl) WithAnnotations() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("StockLossAnnotations", "delete_time = 0")
	}
}
