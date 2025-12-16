package service

import (
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/modules/inventory/domain/entity"
	"ttpos-server-go/app/modules/inventory/domain/repository"
	"ttpos-server-go/app/modules/inventory/domain/valueobject"
	"ttpos-server-go/pkg/context"
)

// IWarehouseDomainService 仓库领域服务接口
type IWarehouseDomainService interface {
	// CreateWarehouse 创建仓库（含业务规则验证）
	CreateWarehouse(ctx context.Context, req CreateWarehouseRequest) (*entity.Warehouse, error)

	// ValidateForUpdate 验证更新操作
	ValidateForUpdate(ctx context.Context, warehouse *entity.Warehouse, req UpdateWarehouseRequest) error

	// ValidateForDelete 验证删除操作
	ValidateForDelete(ctx context.Context, warehouse *entity.Warehouse, currentCompanyUuid uint64) error

	// SetDefaultWarehouse 设置默认仓库（确保只有一个默认仓库）
	SetDefaultWarehouse(ctx context.Context, uuid uint64) error
}

// CreateWarehouseRequest 创建仓库请求
type CreateWarehouseRequest struct {
	Name        valueobject.MultiLanguageName
	Type        valueobject.WarehouseType
	Code        valueobject.WarehouseCode
	Status      valueobject.WarehouseStatus
	ContactInfo valueobject.ContactInfo
}

// UpdateWarehouseRequest 更新仓库请求
type UpdateWarehouseRequest struct {
	Name        valueobject.MultiLanguageName
	Type        valueobject.WarehouseType
	Code        valueobject.WarehouseCode
	Status      valueobject.WarehouseStatus
	ContactInfo valueobject.ContactInfo
}

// warehouseDomainService 仓库领域服务实现
type warehouseDomainService struct {
	warehouseRepo repository.IWarehouseRepository
}

// NewWarehouseDomainService 创建仓库领域服务
func NewWarehouseDomainService(warehouseRepo repository.IWarehouseRepository) IWarehouseDomainService {
	return &warehouseDomainService{
		warehouseRepo: warehouseRepo,
	}
}

// CreateWarehouse 创建仓库（含业务规则验证）
func (s *warehouseDomainService) CreateWarehouse(ctx context.Context, req CreateWarehouseRequest) (*entity.Warehouse, error) {
	// 业务规则：检查编码是否已存在
	exists, err := s.warehouseRepo.ExistsCode(ctx, req.Code, 0)
	if err != nil {
		return nil, errors.WithMessage(err, "检查仓库编码失败")
	}
	if exists {
		return nil, errors.New("仓库编码已存在")
	}

	// 创建仓库实体
	warehouse := entity.NewWarehouse(
		req.Name,
		req.Type,
		req.Code,
		req.Status,
		req.ContactInfo,
	)

	return warehouse, nil
}

// ValidateForUpdate 验证更新操作
func (s *warehouseDomainService) ValidateForUpdate(ctx context.Context, warehouse *entity.Warehouse, req UpdateWarehouseRequest) error {
	// 业务规则：检查编码是否已被其他仓库使用
	if !req.Code.Equals(warehouse.Code()) {
		exists, err := s.warehouseRepo.ExistsCode(ctx, req.Code, warehouse.Uuid())
		if err != nil {
			return errors.WithMessage(err, "检查仓库编码失败")
		}
		if exists {
			return errors.New("仓库编码已存在")
		}
	}

	// 业务规则：默认仓库或在途仓库不可禁用
	if req.Status.IsDisabled() {
		if warehouse.IsDefault() {
			return errors.New("默认仓库不可禁用")
		}
		if warehouse.IsTransit() {
			return errors.New("在途仓库不可禁用")
		}
	}

	return nil
}

// ValidateForDelete 验证删除操作
func (s *warehouseDomainService) ValidateForDelete(ctx context.Context, warehouse *entity.Warehouse, currentCompanyUuid uint64) error {
	// 业务规则：检查是否可以删除
	if err := warehouse.CanBeDeleted(); err != nil {
		return err
	}

	// 业务规则：检查是否可以编辑（总部仓库在子店不可删除）
	if err := warehouse.CanBeEdited(currentCompanyUuid); err != nil {
		return errors.New("仓库不可删除")
	}

	return nil
}

// SetDefaultWarehouse 设置默认仓库
func (s *warehouseDomainService) SetDefaultWarehouse(ctx context.Context, uuid uint64) error {
	// 查找要设置为默认的仓库
	warehouse, err := s.warehouseRepo.FindByUuid(ctx, uuid)
	if err != nil {
		return errors.WithMessage(err, "查询仓库失败")
	}
	if warehouse == nil {
		return errors.New("仓库不存在")
	}

	// 业务规则：在途仓库不能设置为默认仓库
	if err := warehouse.SetAsDefault(); err != nil {
		return err
	}

	// 使用Repository的原子操作设置默认仓库
	return s.warehouseRepo.SetAsDefault(ctx, uuid)
}
