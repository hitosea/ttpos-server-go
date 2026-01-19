package inventory

import (
	"time"

	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/modules/inventory/domain/entity"
	"ttpos-server-go/app/modules/inventory/domain/repository"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/utils"

	"gorm.io/gorm"
)

// WarehouseItemRepositoryImpl 库存物品仓储实现
type WarehouseItemRepositoryImpl struct {
	dbm *database.DBManager
}

// NewWarehouseItemRepository 创建库存物品仓储
func NewWarehouseItemRepository(dbm *database.DBManager) repository.IWarehouseItemRepository {
	return &WarehouseItemRepositoryImpl{
		dbm: dbm,
	}
}

// Save 保存库存物品
func (r *WarehouseItemRepositoryImpl) Save(ctx context.Context, item *entity.WarehouseItem) error {
	db := r.getDB(ctx)
	po := r.toPO(item)

	if item.Uuid() == 0 {
		// 生成UUID
		uuid, err := utils.GetID()
		if err != nil {
			return errors.WithMessage(err, "生成UUID失败")
		}
		po.Uuid = uuid
		item.SetUuid(uuid)

		// 创建
		if err := db.Create(po).Error; err != nil {
			return errors.WithMessage(err, "创建库存物品失败")
		}
	} else {
		// 更新
		if err := db.Model(&model.WarehouseItem{}).Where("uuid = ?", po.Uuid).Updates(map[string]any{
			"stock":          po.Stock,
			"reserved_stock": po.ReservedStock,
			"valuation":      po.Valuation,
			"update_time":    po.UpdateTime,
		}).Error; err != nil {
			return errors.WithMessage(err, "更新库存物品失败")
		}
	}

	return nil
}

// FindByUuid 根据UUID查找库存物品
func (r *WarehouseItemRepositoryImpl) FindByUuid(ctx context.Context, uuid uint64) (*entity.WarehouseItem, error) {
	db := r.getDB(ctx)

	var po model.WarehouseItem
	if err := db.Where("uuid = ? AND delete_time = 0", uuid).First(&po).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, errors.WithMessage(err, "查询库存物品失败")
	}

	return r.toDomain(&po), nil
}

// FindByWarehouseAndMaterial 根据仓库UUID和物品UUID查找库存
func (r *WarehouseItemRepositoryImpl) FindByWarehouseAndMaterial(ctx context.Context, warehouseUuid, materialUuid uint64) (*entity.WarehouseItem, error) {
	db := r.getDB(ctx)

	var po model.WarehouseItem
	if err := db.Where("warehouse_uuid = ? AND material_uuid = ? AND delete_time = 0", warehouseUuid, materialUuid).First(&po).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, errors.WithMessage(err, "查询库存物品失败")
	}

	return r.toDomain(&po), nil
}

// FindByMaterialUuid 根据物品UUID查找所有仓库的库存
func (r *WarehouseItemRepositoryImpl) FindByMaterialUuid(ctx context.Context, materialUuid uint64) ([]*entity.WarehouseItem, error) {
	db := r.getDB(ctx)

	var pos []model.WarehouseItem
	if err := db.Where("material_uuid = ? AND delete_time = 0", materialUuid).Find(&pos).Error; err != nil {
		return nil, errors.WithMessage(err, "查询库存物品失败")
	}

	items := make([]*entity.WarehouseItem, 0, len(pos))
	for _, po := range pos {
		items = append(items, r.toDomain(&po))
	}

	return items, nil
}

// FindByWarehouseUuid 根据仓库UUID查找所有库存物品
func (r *WarehouseItemRepositoryImpl) FindByWarehouseUuid(ctx context.Context, warehouseUuid uint64) ([]*entity.WarehouseItem, error) {
	db := r.getDB(ctx)

	var pos []model.WarehouseItem
	if err := db.Where("warehouse_uuid = ? AND delete_time = 0", warehouseUuid).Find(&pos).Error; err != nil {
		return nil, errors.WithMessage(err, "查询库存物品失败")
	}

	items := make([]*entity.WarehouseItem, 0, len(pos))
	for _, po := range pos {
		items = append(items, r.toDomain(&po))
	}

	return items, nil
}

// FindOrCreate 根据仓库和物品查找或创建库存记录
func (r *WarehouseItemRepositoryImpl) FindOrCreate(ctx context.Context, warehouseUuid, materialUuid uint64, materialCode string, valuation float64) (*entity.WarehouseItem, error) {
	item, err := r.FindByWarehouseAndMaterial(ctx, warehouseUuid, materialUuid)
	if err != nil {
		return nil, err
	}

	if item != nil {
		return item, nil
	}

	// 创建新记录
	newItem := entity.NewWarehouseItem(warehouseUuid, materialUuid, materialCode, valuation)
	if err := r.Save(ctx, newItem); err != nil {
		return nil, err
	}

	return newItem, nil
}

// Remove 删除库存物品（软删除）
func (r *WarehouseItemRepositoryImpl) Remove(ctx context.Context, uuid uint64) error {
	db := r.getDB(ctx)

	if err := db.Model(&model.WarehouseItem{}).Where("uuid = ?", uuid).Update("delete_time", time.Now().Unix()).Error; err != nil {
		return errors.WithMessage(err, "删除库存物品失败")
	}

	return nil
}

// GetMaterialStockInNormalWarehouses 获取物品在普通仓库（非在途）的总库存
func (r *WarehouseItemRepositoryImpl) GetMaterialStockInNormalWarehouses(ctx context.Context, materialUuids []uint64) (map[uint64]float64, error) {
	if len(materialUuids) == 0 {
		return make(map[uint64]float64), nil
	}

	db := r.getDB(ctx)
	tablePrefix := config.Database.TablePrefix

	type result struct {
		MaterialUuid uint64  `gorm:"column:material_uuid"`
		TotalStock   float64 `gorm:"column:total_stock"`
	}

	var results []result
	err := db.Table(tablePrefix+"warehouse_item").
		Select("material_uuid, SUM(stock) as total_stock").
		Where("material_uuid IN (?)", materialUuids).
		Where("delete_time = 0").
		Where("warehouse_uuid IN (?)",
			db.Table(tablePrefix+"warehouse").
				Select("uuid").
				Where("type != ?", constant.WarehouseTypeTransit).
				Where("delete_time = 0"),
		).
		Group("material_uuid").
		Scan(&results).Error

	if err != nil {
		return nil, errors.WithMessage(err, "获取物品库存失败")
	}

	stockMap := make(map[uint64]float64)
	for _, res := range results {
		stockMap[res.MaterialUuid] = res.TotalStock
	}

	return stockMap, nil
}

// GetMaterialStockByWarehouse 获取物品在各仓库的库存详情
func (r *WarehouseItemRepositoryImpl) GetMaterialStockByWarehouse(ctx context.Context, materialUuid uint64) ([]repository.WarehouseStockInfo, error) {
	db := r.getDB(ctx)
	tablePrefix := config.Database.TablePrefix

	type result struct {
		WarehouseUuid uint64  `gorm:"column:warehouse_uuid"`
		WarehouseName string  `gorm:"column:warehouse_name"`
		WarehouseCode string  `gorm:"column:warehouse_code"`
		WarehouseType string  `gorm:"column:warehouse_type"`
		Stock         float64 `gorm:"column:stock"`
		ReservedStock float64 `gorm:"column:reserved_stock"`
	}

	var results []result
	err := db.Table(tablePrefix+"warehouse_item AS wi").
		Select("wi.warehouse_uuid, w.code as warehouse_code, w.type as warehouse_type, w.name as warehouse_name, wi.stock, wi.reserved_stock").
		Joins("LEFT JOIN "+tablePrefix+"warehouse AS w ON wi.warehouse_uuid = w.uuid").
		Where("wi.material_uuid = ?", materialUuid).
		Where("wi.delete_time = 0").
		Where("w.delete_time = 0").
		Order("wi.stock DESC").
		Scan(&results).Error

	if err != nil {
		return nil, errors.WithMessage(err, "获取物品仓库库存详情失败")
	}

	stockList := make([]repository.WarehouseStockInfo, 0, len(results))
	for _, res := range results {
		stockList = append(stockList, repository.WarehouseStockInfo{
			WarehouseUuid: res.WarehouseUuid,
			WarehouseName: res.WarehouseName,
			WarehouseCode: res.WarehouseCode,
			WarehouseType: res.WarehouseType,
			Stock:         res.Stock,
			ReservedStock: res.ReservedStock,
		})
	}

	return stockList, nil
}

// FindWithPagination 分页查询库存物品
func (r *WarehouseItemRepositoryImpl) FindWithPagination(ctx context.Context, spec *repository.WarehouseItemQuerySpec, pageNo, pageSize int) ([]*entity.WarehouseItem, int64, error) {
	db := r.getDB(ctx)

	query := db.Model(&model.WarehouseItem{}).Where("delete_time = 0")

	// 应用查询条件
	if spec != nil {
		if spec.WarehouseUuid != nil {
			query = query.Where("warehouse_uuid = ?", *spec.WarehouseUuid)
		}
		if spec.MaterialUuid != nil {
			query = query.Where("material_uuid = ?", *spec.MaterialUuid)
		}
		if spec.MaterialCode != nil {
			query = query.Where("material_code = ?", *spec.MaterialCode)
		}
		if spec.Keyword != nil && *spec.Keyword != "" {
			query = query.Where("material_code LIKE ?", "%"+*spec.Keyword+"%")
		}
		if spec.HasStock != nil && *spec.HasStock {
			query = query.Where("stock > 0")
		}
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errors.WithMessage(err, "查询库存物品总数失败")
	}

	// 分页查询
	var pos []model.WarehouseItem
	offset := (pageNo - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("update_time DESC").Find(&pos).Error; err != nil {
		return nil, 0, errors.WithMessage(err, "查询库存物品失败")
	}

	items := make([]*entity.WarehouseItem, 0, len(pos))
	for _, po := range pos {
		items = append(items, r.toDomain(&po))
	}

	return items, total, nil
}

// BatchUpdateStock 批量更新库存
func (r *WarehouseItemRepositoryImpl) BatchUpdateStock(ctx context.Context, items []*entity.WarehouseItem) error {
	db := r.getDB(ctx)

	return db.Transaction(func(tx *gorm.DB) error {
		for _, item := range items {
			if err := tx.Model(&model.WarehouseItem{}).Where("uuid = ?", item.Uuid()).Updates(map[string]any{
				"stock":          item.Stock().Value(),
				"reserved_stock": item.ReservedStock().Value(),
				"update_time":    time.Now().Unix(),
			}).Error; err != nil {
				return errors.WithMessage(err, "批量更新库存失败")
			}
		}
		return nil
	})
}

// AddStock 原子增加库存
func (r *WarehouseItemRepositoryImpl) AddStock(ctx context.Context, uuid uint64, quantity float64) error {
	db := r.getDB(ctx)

	if err := db.Model(&model.WarehouseItem{}).Where("uuid = ?", uuid).
		UpdateColumn("stock", gorm.Expr("stock + ?", quantity)).Error; err != nil {
		return errors.WithMessage(err, "增加库存失败")
	}

	return nil
}

// ReduceStock 原子减少库存
func (r *WarehouseItemRepositoryImpl) ReduceStock(ctx context.Context, uuid uint64, quantity float64) error {
	db := r.getDB(ctx)

	if err := db.Model(&model.WarehouseItem{}).Where("uuid = ?", uuid).
		UpdateColumn("stock", gorm.Expr("stock - ?", quantity)).Error; err != nil {
		return errors.WithMessage(err, "减少库存失败")
	}

	return nil
}

// AddReservedStock 原子增加预留库存
func (r *WarehouseItemRepositoryImpl) AddReservedStock(ctx context.Context, uuid uint64, quantity float64) error {
	db := r.getDB(ctx)

	if err := db.Model(&model.WarehouseItem{}).Where("uuid = ?", uuid).
		UpdateColumn("reserved_stock", gorm.Expr("reserved_stock + ?", quantity)).Error; err != nil {
		return errors.WithMessage(err, "增加预留库存失败")
	}

	return nil
}

// ReduceReservedStock 原子减少预留库存
func (r *WarehouseItemRepositoryImpl) ReduceReservedStock(ctx context.Context, uuid uint64, quantity float64) error {
	db := r.getDB(ctx)

	if err := db.Model(&model.WarehouseItem{}).Where("uuid = ?", uuid).
		UpdateColumn("reserved_stock", gorm.Expr("reserved_stock - ?", quantity)).Error; err != nil {
		return errors.WithMessage(err, "减少预留库存失败")
	}

	return nil
}

// toPO 转换为持久化对象
func (r *WarehouseItemRepositoryImpl) toPO(item *entity.WarehouseItem) *model.WarehouseItem {
	return &model.WarehouseItem{
		BaseModel: model.BaseModel{
			Uuid:       item.Uuid(),
			CreateTime: item.CreateTime(),
			UpdateTime: item.UpdateTime(),
		},
		WarehouseUuid: item.WarehouseUuid(),
		MaterialUuid:  item.MaterialUuid(),
		MaterialCode:  item.MaterialCode(),
		Stock:         item.Stock().Value(),
		ReservedStock: item.ReservedStock().Value(),
		Valuation:     item.Valuation(),
	}
}

// toDomain 转换为领域对象
func (r *WarehouseItemRepositoryImpl) toDomain(po *model.WarehouseItem) *entity.WarehouseItem {
	return entity.ReconstructWarehouseItem(
		po.Uuid,
		po.WarehouseUuid,
		po.MaterialUuid,
		po.MaterialCode,
		po.Stock,
		po.ReservedStock,
		po.Valuation,
		po.CreateTime,
		po.UpdateTime,
	)
}

// getDB 获取数据库连接
func (r *WarehouseItemRepositoryImpl) getDB(ctx context.Context) *gorm.DB {
	return ctx.GetDB()
}
