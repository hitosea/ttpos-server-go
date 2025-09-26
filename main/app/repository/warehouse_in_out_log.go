package repository

import (
	"time"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IWarehouseInOutLogRepo 仓库出入库日志Repository接口
type IWarehouseInOutLogRepo interface {
	// 基础操作
	Create(warehouseLog *model.WarehouseInOutLog) error
	Update(warehouseLog *model.WarehouseInOutLog) error
	Delete(uuid uint64) error
	GetByUuid(uuid uint64, opts ...DBOption) (*model.WarehouseInOutLog, error)

	// 查询操作
	GetListWithPagination(pageNo, pageSize int, opts ...DBOption) ([]model.WarehouseInOutLog, int64, error)
	GetWarehouseInOutLogs(opts ...DBOption) ([]*model.WarehouseInOutLog, error)
	GetByWarehouseUuid(warehouseUuid uint64, opts ...DBOption) ([]model.WarehouseInOutLog, error)
	GetByMaterialUuid(materialUuid uint64, opts ...DBOption) ([]model.WarehouseInOutLog, error)
	GetByOrderNo(orderNo string, opts ...DBOption) ([]model.WarehouseInOutLog, error)
	GetByLogType(logType int, opts ...DBOption) ([]model.WarehouseInOutLog, error)
	GetByScene(scene int, opts ...DBOption) ([]model.WarehouseInOutLog, error)

	// 条件查询选项
	WhereWarehouseUuid(warehouseUuid uint64) DBOption
	WhereMaterialUuid(materialUuid uint64) DBOption
	WhereMaterialUuids(materialUuids []uint64) DBOption
	WhereMaterialCategoryUuids(materialCategoryUuids []uint64) DBOption
	WhereSupplierUuids(supplierUuids []uint64) DBOption
	WhereOrderNo(orderNo string) DBOption
	WhereLogType(logType int) DBOption
	WhereScene(scene int) DBOption
	WhereCreateTimeBetween(startTime, endTime int) DBOption
	WhereMaterialNameLike(keyword string) DBOption // 根据物品名称模糊查询
	OrderByCreateTime(desc bool) DBOption
}

// WarehouseInOutLogRepoImpl 仓库出入库日志Repository实现
type WarehouseInOutLogRepoImpl struct {
	db *gorm.DB
}

// NewWarehouseInOutLogRepo 创建仓库出入库日志Repository
func NewWarehouseInOutLogRepo(db *gorm.DB) IWarehouseInOutLogRepo {
	return &WarehouseInOutLogRepoImpl{db: db}
}

// Create 创建仓库出入库日志
func (r *WarehouseInOutLogRepoImpl) Create(warehouseLog *model.WarehouseInOutLog) error {
	return r.db.Create(warehouseLog).Error
}

// Update 更新仓库出入库日志
func (r *WarehouseInOutLogRepoImpl) Update(warehouseLog *model.WarehouseInOutLog) error {
	return r.db.Save(warehouseLog).Error
}

// Delete 删除仓库出入库日志（软删除）
func (r *WarehouseInOutLogRepoImpl) Delete(uuid uint64) error {
	return r.db.Model(&model.WarehouseInOutLog{}).Where("uuid = ?", uuid).Update("delete_time", time.Now().Unix()).Error
}

// GetByUuid 根据UUID获取仓库出入库日志
func (r *WarehouseInOutLogRepoImpl) GetByUuid(uuid uint64, opts ...DBOption) (*model.WarehouseInOutLog, error) {
	var warehouseLog model.WarehouseInOutLog
	query := r.db.Where("uuid = ?", uuid).Scopes(NotDeleted)
	// 应用查询选项
	for _, opt := range opts {
		query = opt(query)
	}
	err := query.First(&warehouseLog).Error
	if err != nil {
		return nil, err
	}
	return &warehouseLog, nil
}

// GetListWithPagination 分页获取仓库出入库日志列表
func (r *WarehouseInOutLogRepoImpl) GetListWithPagination(pageNo, pageSize int, opts ...DBOption) ([]model.WarehouseInOutLog, int64, error) {
	var warehouseLogs []model.WarehouseInOutLog
	var total int64

	// 构建基础查询，包含联表和预加载
	query := r.db.Model(&model.WarehouseInOutLog{}).
		Scopes(NotDeleted).
		Preload("Material.MultiLanguageName").
		Preload("Supplier").
		Preload("Warehouse.MultiLanguageName")

	// 应用查询选项
	for _, opt := range opts {
		query = opt(query)
	}

	// 计算总数
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (pageNo - 1) * pageSize
	err = query.Order("create_time DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&warehouseLogs).Error

	return warehouseLogs, total, err
}

// GetByWarehouseUuid 根据仓库UUID获取出入库日志列表
func (r *WarehouseInOutLogRepoImpl) GetByWarehouseUuid(warehouseUuid uint64, opts ...DBOption) ([]model.WarehouseInOutLog, error) {
	var warehouseLogs []model.WarehouseInOutLog
	query := r.db.Where("warehouse_uuid = ?", warehouseUuid).Scopes(NotDeleted)
	// 应用查询选项
	for _, opt := range opts {
		query = opt(query)
	}
	err := query.Find(&warehouseLogs).Error
	return warehouseLogs, err
}

// GetByMaterialUuid 根据物料UUID获取出入库日志列表
func (r *WarehouseInOutLogRepoImpl) GetByMaterialUuid(materialUuid uint64, opts ...DBOption) ([]model.WarehouseInOutLog, error) {
	var warehouseLogs []model.WarehouseInOutLog
	query := r.db.Where("material_uuid = ?", materialUuid).Scopes(NotDeleted)
	// 应用查询选项
	for _, opt := range opts {
		query = opt(query)
	}
	err := query.Find(&warehouseLogs).Error
	return warehouseLogs, err
}

// GetByOrderNo 根据单据编号获取出入库日志列表
func (r *WarehouseInOutLogRepoImpl) GetByOrderNo(orderNo string, opts ...DBOption) ([]model.WarehouseInOutLog, error) {
	var warehouseLogs []model.WarehouseInOutLog
	query := r.db.Where("order_no = ?", orderNo).Scopes(NotDeleted)
	// 应用查询选项
	for _, opt := range opts {
		query = opt(query)
	}
	err := query.Find(&warehouseLogs).Error
	return warehouseLogs, err
}

// GetByLogType 根据日志类型获取出入库日志列表
func (r *WarehouseInOutLogRepoImpl) GetByLogType(logType int, opts ...DBOption) ([]model.WarehouseInOutLog, error) {
	var warehouseLogs []model.WarehouseInOutLog
	query := r.db.Where("log_type = ?", logType).Scopes(NotDeleted)
	// 应用查询选项
	for _, opt := range opts {
		query = opt(query)
	}
	err := query.Find(&warehouseLogs).Error
	return warehouseLogs, err
}

// GetByScene 根据场景获取出入库日志列表
func (r *WarehouseInOutLogRepoImpl) GetByScene(scene int, opts ...DBOption) ([]model.WarehouseInOutLog, error) {
	var warehouseLogs []model.WarehouseInOutLog
	query := r.db.Where("scene = ?", scene).Scopes(NotDeleted)
	// 应用查询选项
	for _, opt := range opts {
		query = opt(query)
	}
	err := query.Find(&warehouseLogs).Error
	return warehouseLogs, err
}

// 以下是查询选项方法

// WhereWarehouseUuid 仓库UUID条件
func (r *WarehouseInOutLogRepoImpl) WhereWarehouseUuid(warehouseUuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("warehouse_uuid = ?", warehouseUuid)
	}
}

// WhereMaterialUuid 物料UUID条件
func (r *WarehouseInOutLogRepoImpl) WhereMaterialUuid(materialUuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("material_uuid = ?", materialUuid)
	}
}

// WhereMaterialUuids 物料UUID列表条件
func (r *WarehouseInOutLogRepoImpl) WhereMaterialUuids(materialUuids []uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("material_uuid IN ?", materialUuids)
	}
}

// WhereMaterialCategoryUuids 物料分类UUID列表条件
func (r *WarehouseInOutLogRepoImpl) WhereMaterialCategoryUuids(materialCategoryUuids []uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("m.category_uuid IN ?", materialCategoryUuids)
	}
}

// WhereOrderNo 单据编号条件
func (r *WarehouseInOutLogRepoImpl) WhereOrderNo(orderNo string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("order_no = ?", orderNo)
	}
}

// WhereLogType 日志类型条件
func (r *WarehouseInOutLogRepoImpl) WhereLogType(logType int) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("log_type = ?", logType)
	}
}

// WhereScene 场景条件
func (r *WarehouseInOutLogRepoImpl) WhereScene(scene int) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("scene = ?", scene)
	}
}

// WhereCreateTimeBetween 创建时间范围条件
func (r *WarehouseInOutLogRepoImpl) WhereCreateTimeBetween(startTime, endTime int) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("create_time BETWEEN ? AND ?", startTime, endTime)
	}
}

// OrderByCreateTime 按创建时间排序
func (r *WarehouseInOutLogRepoImpl) OrderByCreateTime(desc bool) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		if desc {
			return db.Order("create_time DESC")
		}
		return db.Order("create_time ASC")
	}
}

// WhereSupplierUuids 供应商UUID列表条件
func (r *WarehouseInOutLogRepoImpl) WhereSupplierUuids(supplierUuids []uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("supplier_uuid IN ?", supplierUuids)
	}
}

// WhereMaterialNameLike 根据物品名称模糊查询
func (r *WarehouseInOutLogRepoImpl) WhereMaterialNameLike(keyword string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("(m.name LIKE ? OR m.code LIKE ? OR m.barcode_value LIKE ?)",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
}

func (r *WarehouseInOutLogRepoImpl) GetWarehouseInOutLogs(opts ...DBOption) ([]*model.WarehouseInOutLog, error) {
	var warehouseLogs []*model.WarehouseInOutLog
	query := r.db.Model(&model.WarehouseInOutLog{}).Scopes(NotDeleted)
	// 应用查询选项
	for _, opt := range opts {
		query = opt(query)
	}
	err := query.Find(&warehouseLogs).Error
	return warehouseLogs, err
}
