package repository

import (
	"ttpos-server-go/app/modules/inventory/domain/entity"
	"ttpos-server-go/app/modules/inventory/domain/valueobject"
	"ttpos-server-go/pkg/context"
)

// IWarehouseRepository 仓库仓储接口（领域层定义）
type IWarehouseRepository interface {
	// Save 保存仓库（创建或更新）
	Save(ctx context.Context, warehouse *entity.Warehouse) error

	// FindByUuid 根据UUID查找仓库
	FindByUuid(ctx context.Context, uuid uint64) (*entity.Warehouse, error)

	// FindByCode 根据编码查找仓库
	FindByCode(ctx context.Context, code valueobject.WarehouseCode) (*entity.Warehouse, error)

	// FindByErpCode 根据ERP编码查找仓库
	FindByErpCode(ctx context.Context, erpCode string) (*entity.Warehouse, error)

	// Remove 删除仓库（软删除）
	Remove(ctx context.Context, uuid uint64) error

	// ExistsCode 检查编码是否存在
	ExistsCode(ctx context.Context, code valueobject.WarehouseCode, excludeUuid uint64) (bool, error)

	// FindDefault 查找默认仓库
	FindDefault(ctx context.Context) (*entity.Warehouse, error)

	// FindAllNormal 查找所有普通仓库
	FindAllNormal(ctx context.Context) ([]*entity.Warehouse, error)

	// FindTransit 查找在途仓库
	FindTransit(ctx context.Context) (*entity.Warehouse, error)

	// FindByHeadquarterUuid 根据总部UUID查找仓库
	FindByHeadquarterUuid(ctx context.Context, headquarterUuid uint64) ([]*entity.Warehouse, error)

	// FindWithPagination 分页查询仓库
	FindWithPagination(
		ctx context.Context,
		spec WarehouseSpecification,
		pageNo, pageSize int,
	) ([]*entity.Warehouse, int64, error)

	// SetAsDefault 设置默认仓库（原子操作：取消其他默认 + 设置新默认）
	SetAsDefault(ctx context.Context, uuid uint64) error
}

// WarehouseSpecification 仓库查询规格接口
type WarehouseSpecification interface {
	// IsSatisfiedBy 判断仓库是否满足规格
	IsSatisfiedBy(warehouse *entity.Warehouse) bool
}
