package repository

import (
	"ttpos-server-go/app/modules/inventory/domain/entity"
	"ttpos-server-go/pkg/context"
)

// IWarehouseItemRepository 库存物品仓储接口（领域层定义）
type IWarehouseItemRepository interface {
	// Save 保存库存物品（创建或更新）
	Save(ctx context.Context, item *entity.WarehouseItem) error

	// FindByUuid 根据UUID查找库存物品
	FindByUuid(ctx context.Context, uuid uint64) (*entity.WarehouseItem, error)

	// FindByWarehouseAndMaterial 根据仓库UUID和物品UUID查找库存
	FindByWarehouseAndMaterial(ctx context.Context, warehouseUuid, materialUuid uint64) (*entity.WarehouseItem, error)

	// FindByMaterialUuid 根据物品UUID查找所有仓库的库存
	FindByMaterialUuid(ctx context.Context, materialUuid uint64) ([]*entity.WarehouseItem, error)

	// FindByWarehouseUuid 根据仓库UUID查找所有库存物品
	FindByWarehouseUuid(ctx context.Context, warehouseUuid uint64) ([]*entity.WarehouseItem, error)

	// FindOrCreate 根据仓库和物品查找或创建库存记录
	FindOrCreate(ctx context.Context, warehouseUuid, materialUuid uint64, materialCode string, valuation float64) (*entity.WarehouseItem, error)

	// Remove 删除库存物品（软删除）
	Remove(ctx context.Context, uuid uint64) error

	// GetMaterialStockInNormalWarehouses 获取物品在普通仓库（非在途）的总库存
	// 返回：map[物品UUID]总库存
	GetMaterialStockInNormalWarehouses(ctx context.Context, materialUuids []uint64) (map[uint64]float64, error)

	// GetMaterialStockByWarehouse 获取物品在各仓库的库存详情
	GetMaterialStockByWarehouse(ctx context.Context, materialUuid uint64) ([]WarehouseStockInfo, error)

	// FindWithPagination 分页查询库存物品
	FindWithPagination(ctx context.Context, spec *WarehouseItemQuerySpec, pageNo, pageSize int) ([]*entity.WarehouseItem, int64, error)

	// BatchUpdateStock 批量更新库存
	BatchUpdateStock(ctx context.Context, items []*entity.WarehouseItem) error

	// AddStock 原子增加库存
	AddStock(ctx context.Context, uuid uint64, quantity float64) error

	// ReduceStock 原子减少库存
	ReduceStock(ctx context.Context, uuid uint64, quantity float64) error

	// AddReservedStock 原子增加预留库存
	AddReservedStock(ctx context.Context, uuid uint64, quantity float64) error

	// ReduceReservedStock 原子减少预留库存
	ReduceReservedStock(ctx context.Context, uuid uint64, quantity float64) error
}

// MaterialStockInfo 物品库存信息
type MaterialStockInfo struct {
	MaterialUuid uint64  // 物品UUID
	TotalStock   float64 // 总库存
}

// WarehouseStockInfo 仓库库存信息
type WarehouseStockInfo struct {
	WarehouseUuid uint64  // 仓库UUID
	WarehouseName string  // 仓库名称
	WarehouseCode string  // 仓库编码
	WarehouseType string  // 仓库类型
	Stock         float64 // 库存数量
	ReservedStock float64 // 预留库存
}

// WarehouseItemQuerySpec 库存物品查询规格
type WarehouseItemQuerySpec struct {
	WarehouseUuid *uint64 // 仓库UUID
	MaterialUuid  *uint64 // 物品UUID
	MaterialCode  *string // 物品编码
	Keyword       *string // 关键字搜索
	HasStock      *bool   // 是否有库存
}

// NewWarehouseItemQuerySpec 创建查询规格
func NewWarehouseItemQuerySpec() *WarehouseItemQuerySpec {
	return &WarehouseItemQuerySpec{}
}

// WithWarehouseUuid 设置仓库UUID
func (s *WarehouseItemQuerySpec) WithWarehouseUuid(warehouseUuid uint64) *WarehouseItemQuerySpec {
	s.WarehouseUuid = &warehouseUuid
	return s
}

// WithMaterialUuid 设置物品UUID
func (s *WarehouseItemQuerySpec) WithMaterialUuid(materialUuid uint64) *WarehouseItemQuerySpec {
	s.MaterialUuid = &materialUuid
	return s
}

// WithMaterialCode 设置物品编码
func (s *WarehouseItemQuerySpec) WithMaterialCode(materialCode string) *WarehouseItemQuerySpec {
	s.MaterialCode = &materialCode
	return s
}

// WithKeyword 设置关键字
func (s *WarehouseItemQuerySpec) WithKeyword(keyword string) *WarehouseItemQuerySpec {
	s.Keyword = &keyword
	return s
}

// WithHasStock 设置是否有库存
func (s *WarehouseItemQuerySpec) WithHasStock(hasStock bool) *WarehouseItemQuerySpec {
	s.HasStock = &hasStock
	return s
}
