package repository

import (
	"time"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IWarehouseItemRepo 仓库商品库存Repository接口
type IWarehouseItemRepo interface {
	// 基础操作
	Create(warehouseItem *model.WarehouseItem) error
	CreateBatch(warehouseItems []*model.WarehouseItem) error
	Update(warehouseItem *model.WarehouseItem) error
	Delete(uuid uint64) error
	GetByUuid(uuid uint64, opts ...DBOption) (*model.WarehouseItem, error)

	// 查询操作
	GetListWithPagination(pageNo, pageSize int, opts ...DBOption) ([]model.WarehouseItem, int64, error)
	GetByWarehouseUuid(warehouseUuid uint64, opts ...DBOption) ([]model.WarehouseItem, error)
	GetByMaterialUuid(materialUuid uint64, opts ...DBOption) ([]model.WarehouseItem, error)
	GetByWarehouseAndMaterial(warehouseUuid, materialUuid uint64, opts ...DBOption) (*model.WarehouseItem, error)
	GetByMaterialCode(materialCode string, opts ...DBOption) ([]model.WarehouseItem, error)

	// 库存操作
	UpdateStock(uuid uint64, stock, reservedStock float64) error
	AddStock(uuid uint64, stock float64) error
	ReduceStock(uuid uint64, stock float64) error
	AddReservedStock(uuid uint64, reservedStock float64) error
	ReduceReservedStock(uuid uint64, reservedStock float64) error
	UpdateStockBatch(items []*model.WarehouseItem) error

	// 条件查询选项
	WhereWarehouseUuid(warehouseUuid uint64) DBOption
	WhereMaterialUuid(materialUuid uint64) DBOption
	WhereMaterialCode(materialCode string) DBOption
	WhereStockGreaterThan(stock float64) DBOption
	WhereReservedStockGreaterThan(reservedStock float64) DBOption
	OrderByCreateTime(desc bool) DBOption
	OrderByStock(desc bool) DBOption
}

// WarehouseItemRepoImpl 仓库商品库存Repository实现
type WarehouseItemRepoImpl struct {
	db *gorm.DB
}

// NewWarehouseItemRepo 创建仓库商品库存Repository
func NewWarehouseItemRepo(db *gorm.DB) IWarehouseItemRepo {
	return &WarehouseItemRepoImpl{db: db}
}

// Create 创建仓库商品库存
func (r *WarehouseItemRepoImpl) Create(warehouseItem *model.WarehouseItem) error {
	return r.db.Create(warehouseItem).Error
}

// CreateBatch 批量创建仓库商品库存
func (r *WarehouseItemRepoImpl) CreateBatch(warehouseItems []*model.WarehouseItem) error {
	if len(warehouseItems) == 0 {
		return nil
	}
	return r.db.CreateInBatches(warehouseItems, 100).Error
}

// Update 更新仓库商品库存
func (r *WarehouseItemRepoImpl) Update(warehouseItem *model.WarehouseItem) error {
	return r.db.Save(warehouseItem).Error
}

// Delete 删除仓库商品库存（软删除）
func (r *WarehouseItemRepoImpl) Delete(uuid uint64) error {
	return r.db.Model(&model.WarehouseItem{}).Where("uuid = ?", uuid).Update("delete_time", time.Now().Unix()).Error
}

// GetByUuid 根据UUID获取仓库商品库存
func (r *WarehouseItemRepoImpl) GetByUuid(uuid uint64, opts ...DBOption) (*model.WarehouseItem, error) {
	var warehouseItem model.WarehouseItem
	query := r.db.Where("uuid = ?", uuid).Scopes(NotDeleted)
	// 应用查询选项
	for _, opt := range opts {
		query = opt(query)
	}
	err := query.First(&warehouseItem).Error
	if err != nil {
		return nil, err
	}
	return &warehouseItem, nil
}

// GetListWithPagination 分页获取仓库商品库存列表
func (r *WarehouseItemRepoImpl) GetListWithPagination(pageNo, pageSize int, opts ...DBOption) ([]model.WarehouseItem, int64, error) {
	var warehouseItems []model.WarehouseItem
	var total int64
	query := r.db.Model(&model.WarehouseItem{}).Scopes(NotDeleted)
	// 应用查询选项
	for _, opt := range opts {
		query = opt(query)
	}
	// 获取总数
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	// 分页查询
	offset := (pageNo - 1) * pageSize
	err = query.Offset(offset).Limit(pageSize).Find(&warehouseItems).Error
	return warehouseItems, total, err
}

// GetByWarehouseUuid 根据仓库UUID获取商品库存列表
func (r *WarehouseItemRepoImpl) GetByWarehouseUuid(warehouseUuid uint64, opts ...DBOption) ([]model.WarehouseItem, error) {
	var warehouseItems []model.WarehouseItem
	query := r.db.Where("warehouse_uuid = ?", warehouseUuid).Scopes(NotDeleted)
	// 应用查询选项
	for _, opt := range opts {
		query = opt(query)
	}
	err := query.Find(&warehouseItems).Error
	return warehouseItems, err
}

// GetByMaterialUuid 根据商品UUID获取库存列表
func (r *WarehouseItemRepoImpl) GetByMaterialUuid(materialUuid uint64, opts ...DBOption) ([]model.WarehouseItem, error) {
	var warehouseItems []model.WarehouseItem
	query := r.db.Where("material_uuid = ?", materialUuid).Scopes(NotDeleted)
	// 应用查询选项
	for _, opt := range opts {
		query = opt(query)
	}
	err := query.Find(&warehouseItems).Error
	return warehouseItems, err
}

// GetByWarehouseAndMaterial 根据仓库UUID和商品UUID获取库存
func (r *WarehouseItemRepoImpl) GetByWarehouseAndMaterial(warehouseUuid, materialUuid uint64, opts ...DBOption) (*model.WarehouseItem, error) {
	var warehouseItem model.WarehouseItem
	query := r.db.Where("warehouse_uuid = ? AND material_uuid = ?", warehouseUuid, materialUuid).Scopes(NotDeleted)
	// 应用查询选项
	for _, opt := range opts {
		query = opt(query)
	}
	err := query.First(&warehouseItem).Error
	if err != nil {
		return nil, err
	}
	return &warehouseItem, nil
}

// GetByMaterialCode 根据商品编码获取库存列表
func (r *WarehouseItemRepoImpl) GetByMaterialCode(materialCode string, opts ...DBOption) ([]model.WarehouseItem, error) {
	var warehouseItems []model.WarehouseItem
	query := r.db.Where("material_code = ?", materialCode).Scopes(NotDeleted)
	// 应用查询选项
	for _, opt := range opts {
		query = opt(query)
	}
	err := query.Find(&warehouseItems).Error
	return warehouseItems, err
}

// UpdateStock 更新库存
func (r *WarehouseItemRepoImpl) UpdateStock(uuid uint64, stock, reservedStock float64) error {
	return r.db.Model(&model.WarehouseItem{}).Where("uuid = ?", uuid).Updates(map[string]interface{}{
		"stock":          stock,
		"reserved_stock": reservedStock,
		"update_time":    time.Now().Unix(),
	}).Error
}

// AddStock 增加库存
func (r *WarehouseItemRepoImpl) AddStock(uuid uint64, stock float64) error {
	return r.db.Model(&model.WarehouseItem{}).Where("uuid = ?", uuid).UpdateColumn("stock", gorm.Expr("stock + ?", stock)).Error
}

// ReduceStock 减少库存
func (r *WarehouseItemRepoImpl) ReduceStock(uuid uint64, stock float64) error {
	return r.db.Model(&model.WarehouseItem{}).Where("uuid = ?", uuid).UpdateColumn("stock", gorm.Expr("stock - ?", stock)).Error
}

// AddReservedStock 增加预留库存
func (r *WarehouseItemRepoImpl) AddReservedStock(uuid uint64, reservedStock float64) error {
	return r.db.Model(&model.WarehouseItem{}).Where("uuid = ?", uuid).UpdateColumn("reserved_stock", gorm.Expr("reserved_stock + ?", reservedStock)).Error
}

// ReduceReservedStock 减少预留库存
func (r *WarehouseItemRepoImpl) ReduceReservedStock(uuid uint64, reservedStock float64) error {
	return r.db.Model(&model.WarehouseItem{}).Where("uuid = ?", uuid).UpdateColumn("reserved_stock", gorm.Expr("reserved_stock - ?", reservedStock)).Error
}

// UpdateStockBatch 批量更新库存
func (r *WarehouseItemRepoImpl) UpdateStockBatch(items []*model.WarehouseItem) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, item := range items {
			if err := tx.Save(item).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// 以下是查询选项方法

// WhereWarehouseUuid 仓库UUID条件
func (r *WarehouseItemRepoImpl) WhereWarehouseUuid(warehouseUuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("warehouse_uuid = ?", warehouseUuid)
	}
}

// WhereMaterialUuid 商品UUID条件
func (r *WarehouseItemRepoImpl) WhereMaterialUuid(materialUuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("material_uuid = ?", materialUuid)
	}
}

// WhereMaterialCode 商品编码条件
func (r *WarehouseItemRepoImpl) WhereMaterialCode(materialCode string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("material_code = ?", materialCode)
	}
}

// WhereStockGreaterThan 库存大于指定值条件
func (r *WarehouseItemRepoImpl) WhereStockGreaterThan(stock float64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("stock > ?", stock)
	}
}

// WhereReservedStockGreaterThan 预留库存大于指定值条件
func (r *WarehouseItemRepoImpl) WhereReservedStockGreaterThan(reservedStock float64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("reserved_stock > ?", reservedStock)
	}
}

// OrderByCreateTime 按创建时间排序
func (r *WarehouseItemRepoImpl) OrderByCreateTime(desc bool) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		if desc {
			return db.Order("create_time DESC")
		}
		return db.Order("create_time ASC")
	}
}

// OrderByStock 按库存排序
func (r *WarehouseItemRepoImpl) OrderByStock(desc bool) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		if desc {
			return db.Order("stock DESC")
		}
		return db.Order("stock ASC")
	}
}
