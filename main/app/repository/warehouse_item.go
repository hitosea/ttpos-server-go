package repository

import (
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/config"

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
	GetNormalByMaterialCodes(materialCodes []string, warehouseErpCode string) ([]model.WarehouseItem, error)

	// 查询操作
	GetListWithPagination(pageNo, pageSize int, opts ...DBOption) ([]model.WarehouseItem, int64, error)
	GetByWarehouseUuid(warehouseUuid uint64, opts ...DBOption) ([]model.WarehouseItem, error)
	GetByMaterialUuid(materialUuid uint64, opts ...DBOption) ([]model.WarehouseItem, error)
	GetByWarehouseAndMaterial(warehouseUuid, materialUuid uint64, opts ...DBOption) (*model.WarehouseItem, error)
	GetByMaterialCode(materialCode string, opts ...DBOption) ([]model.WarehouseItem, error)
	GetByWarehouseErpCode(warehouseErpCode string, opts ...DBOption) ([]model.WarehouseItem, error)
	GetListWithWarehouseInfo(pageNo, pageSize int, opts ...DBOption) ([]model.WarehouseItem, int64, error)
	// 通过物品UUID获取该物品在各个仓库的库存情况
	GetWarehouseItems(opts ...DBOption) ([]*model.WarehouseItem, error)
	GetWarehouseItemsByMaterialUuid(materialUuid uint64) ([]*model.WarehouseItem, error)
	// 获取在途仓库库存
	GetTransitWarehouseItemByWarehouseAndMaterial(warehouseUuid, materialUuid uint64, materialCode string, valuation float64) (*model.WarehouseItem, error)

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
	WhereWarehouseErpCode(warehouseErpCode string) DBOption
	WhereStockGreaterThan(stock float64) DBOption
	WhereReservedStockGreaterThan(reservedStock float64) DBOption
	OrderByCreateTime(desc bool) DBOption
	OrderByStock(desc bool) DBOption

	// 获取仓库物品列表（带物品详细信息）
	GetWarehouseMaterialsWithPagination(pageNo, pageSize int, opts ...DBOption) ([]*model.WarehouseItem, int64, error)
	GetWarehouseMaterials(opts ...DBOption) ([]*model.WarehouseItem, error)

	// 获取物料在非在途仓库中的总库存
	GetMaterialStockInNormalWarehouses(materialUuids []uint64) (map[uint64]float64, error)
	// 按照物料UUID和仓库UUID分组获取物品库存
	GetMaterialStockByWarehouse(materialUuids []uint64) ([]MaterialWarehouseStockResult, error)
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

// GetNormalByMaterialCodes 根据物料编码获取仓库商品库存列表
func (r *WarehouseItemRepoImpl) GetNormalByMaterialCodes(materialCodes []string, warehouseErpCode string) ([]model.WarehouseItem, error) {
	var warehouseItems []model.WarehouseItem
	query := r.db.Model(&model.WarehouseItem{}).
		Joins("JOIN ttpos_warehouse ON ttpos_warehouse_item.warehouse_uuid = ttpos_warehouse.uuid").
		Where("ttpos_warehouse_item.material_code IN (?)", materialCodes).
		Where("ttpos_warehouse.type != ?", constant.WarehouseTypeTransit).
		Where("ttpos_warehouse.delete_time = ?", 0)
	if warehouseErpCode != "" {
		query = query.Where("ttpos_warehouse.erp_code = ?", warehouseErpCode)
	}
	err := query.Find(&warehouseItems).Error
	return warehouseItems, err
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

// GetByWarehouseErpCode 根据仓库ERP编码获取库存列表
func (r *WarehouseItemRepoImpl) GetByWarehouseErpCode(warehouseErpCode string, opts ...DBOption) ([]model.WarehouseItem, error) {
	var warehouseItems []model.WarehouseItem
	query := r.db.Joins("JOIN ttpos_warehouse ON ttpos_warehouse_item.warehouse_uuid = ttpos_warehouse.uuid").
		Where("ttpos_warehouse.erp_code = ?", warehouseErpCode).
		Where("ttpos_warehouse.delete_time = 0").
		Scopes(NotDeleted)
	// 应用查询选项
	for _, opt := range opts {
		query = opt(query)
	}
	err := query.Find(&warehouseItems).Error
	return warehouseItems, err
}

// GetListWithWarehouseInfo 分页获取仓库商品库存列表（包含仓库信息）
func (r *WarehouseItemRepoImpl) GetListWithWarehouseInfo(pageNo, pageSize int, opts ...DBOption) ([]model.WarehouseItem, int64, error) {
	var warehouseItems []model.WarehouseItem
	var total int64
	query := r.db.Model(&model.WarehouseItem{}).
		Preload("Warehouse").
		Scopes(NotDeleted)
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

// WhereWarehouseErpCode 仓库ERP编码条件
func (r *WarehouseItemRepoImpl) WhereWarehouseErpCode(warehouseErpCode string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Joins("JOIN ttpos_warehouse ON ttpos_warehouse_item.warehouse_uuid = ttpos_warehouse.uuid").
			Where("ttpos_warehouse.erp_code = ?", warehouseErpCode).
			Where("ttpos_warehouse.delete_time = 0")
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

func (r *WarehouseItemRepoImpl) GetWarehouseItems(opts ...DBOption) ([]*model.WarehouseItem, error) {
	var warehouseItems []*model.WarehouseItem
	query := r.db.Model(&model.WarehouseItem{}).Scopes(NotDeleted)
	for _, opt := range opts {
		query = opt(query)
	}
	err := query.Find(&warehouseItems).Error
	return warehouseItems, err
}

// GetWarehouseItemsByMaterialUuid 通过物品UUID获取该物品在各个仓库的库存情况
func (r *WarehouseItemRepoImpl) GetWarehouseItemsByMaterialUuid(materialUuid uint64) ([]*model.WarehouseItem, error) {
	return r.GetWarehouseItems(
		r.WhereMaterialUuid(materialUuid),
		CommonRepo.Preload(
			WithPreload{
				Query: "Warehouse.MultiLanguageName",
			},
			WithPreload{
				Query: "Material.NotBaseUnitList",
				Args: []any{
					func(db *gorm.DB) *gorm.DB {
						return db.Where("delete_time = ?", constant.NotDeleted)
					},
				},
			},
			WithPreload{
				Query: "Material.NotBaseUnitList.Unit.MultiLanguageName",
			},
		),
	)
}

// GetWarehouseMaterialsWithPagination 获取仓库物品列表（带物品详细信息）
func (r *WarehouseItemRepoImpl) GetWarehouseMaterialsWithPagination(pageNo, pageSize int, opts ...DBOption) ([]*model.WarehouseItem, int64, error) {
	var items []*model.WarehouseItem
	var total int64

	query := r.db.Model(&model.WarehouseItem{}).Scopes(NotDeleted)

	// 应用查询选项
	for _, opt := range opts {
		query = opt(query)
	}

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (pageNo - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("create_time DESC").Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

// GetWarehouseMaterialsWithPagination 获取仓库物品列表（带物品详细信息）
func (r *WarehouseItemRepoImpl) GetWarehouseMaterials(opts ...DBOption) ([]*model.WarehouseItem, error) {
	var items []*model.WarehouseItem

	query := r.db.Model(&model.WarehouseItem{}).Scopes(NotDeleted)

	// 应用查询选项
	for _, opt := range opts {
		query = opt(query)
	}

	// 分页查询
	if err := query.Order("create_time DESC").Find(&items).Error; err != nil {
		return nil, err
	}

	return items, nil
}

// MaterialStockResult 物料库存统计结果
type MaterialStockResult struct {
	MaterialUuid uint64  `gorm:"column:material_uuid"` // 物料UUID
	TotalStock   float64 `gorm:"column:total_stock"`   // 总库存
}

// MaterialWarehouseStockResult 物料在各仓库的库存统计结果
type MaterialWarehouseStockResult struct {
	MaterialUuid  uint64  `gorm:"column:material_uuid"`  // 物料UUID
	WarehouseUuid uint64  `gorm:"column:warehouse_uuid"` // 仓库UUID
	Stock         float64 `gorm:"column:stock"`          // 库存数量
}

// GetMaterialStockInNormalWarehouses 获取物料在非在途仓库中的总库存
// 返回：map[物料UUID]总库存
func (r *WarehouseItemRepoImpl) GetMaterialStockInNormalWarehouses(materialUuids []uint64) (map[uint64]float64, error) {
	if len(materialUuids) == 0 {
		return make(map[uint64]float64), nil
	}

	var results []MaterialStockResult

	// 执行SQL查询：
	// SELECT material_uuid, SUM(stock) as total_stock
	// FROM ttpos_warehouse_item
	// WHERE material_uuid IN (?)
	//   AND warehouse_uuid IN (
	//     SELECT uuid FROM ttpos_warehouse
	//     WHERE type != 'transit' AND delete_time = 0
	//   )
	// GROUP BY material_uuid
	// 获取表前缀
	tablePrefix := config.Database.TablePrefix
	err := r.db.Table(tablePrefix+"warehouse_item").
		Select("material_uuid, SUM(stock) as total_stock").
		Where("material_uuid IN (?)", materialUuids).
		Where("warehouse_uuid IN (?)",
			r.db.Table(tablePrefix+"warehouse").
				Select("uuid").
				Where("type != ?", constant.WarehouseTypeTransit).
				Where("delete_time = ?", 0),
		).
		Group("material_uuid").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	// 将结果转换为 map
	stockMap := make(map[uint64]float64)
	for _, result := range results {
		stockMap[result.MaterialUuid] = result.TotalStock
	}

	return stockMap, nil
}

// GetMaterialStockByWarehouse 按照物料UUID和仓库UUID分组获取物品库存
// 返回：每个物料在每个仓库中的库存详情列表
func (r *WarehouseItemRepoImpl) GetMaterialStockByWarehouse(materialUuids []uint64) ([]MaterialWarehouseStockResult, error) {
	if len(materialUuids) == 0 {
		return []MaterialWarehouseStockResult{}, nil
	}

	var results []MaterialWarehouseStockResult

	// 获取表前缀
	tablePrefix := config.Database.TablePrefix

	// 执行SQL查询：
	// SELECT material_uuid, warehouse_uuid, SUM(stock) as stock
	// FROM ttpos_warehouse_item
	// WHERE material_uuid IN (?)
	//   AND warehouse_uuid IN (
	//     SELECT uuid FROM ttpos_warehouse
	//     WHERE type != 'transit' AND delete_time = 0
	//   )
	// GROUP BY material_uuid, warehouse_uuid
	err := r.db.Table(tablePrefix+"warehouse_item").
		Select("material_uuid, warehouse_uuid, SUM(stock) as stock").
		Where("material_uuid IN (?)", materialUuids).
		Where("warehouse_uuid IN (?)",
			r.db.Table(tablePrefix+"warehouse").
				Select("uuid").
				Where("type != ?", constant.WarehouseTypeTransit).
				Where("delete_time = ?", 0),
		).
		Group("material_uuid, warehouse_uuid").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	return results, nil
}

// GetTransitWarehouseItemByWarehouseAndMaterial 获取在途仓库库存
func (r *WarehouseItemRepoImpl) GetTransitWarehouseItemByWarehouseAndMaterial(
	warehouseUuid, materialUuid uint64,
	materialCode string,
	valuation float64,
) (*model.WarehouseItem, error) {
	warehouseItem, err := r.GetByWarehouseAndMaterial(
		warehouseUuid,
		materialUuid,
	)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 没有找到记录时创建新记录
			newWarehouseItem := &model.WarehouseItem{
				WarehouseUuid: warehouseUuid,
				MaterialUuid:  materialUuid,
				MaterialCode:  materialCode,
				Stock:         0,
				Valuation:     valuation,
			}
			err = r.Create(newWarehouseItem)
			if err != nil {
				return nil, errors.WithMessage(errors.New("创建仓库商品库存记录失败"), err.Error())
			}
			warehouseItem = newWarehouseItem
		} else {
			return nil, errors.WithMessage(errors.New("查询在途仓库库存失败"), err.Error())
		}
	}
	return warehouseItem, nil
}
