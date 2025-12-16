package service

import (
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/modules/inventory/domain/repository"
	"ttpos-server-go/pkg/context"
)

// IWarehouseItemDomainService 库存物品领域服务接口
type IWarehouseItemDomainService interface {
	// GetMaterialStock 获取物品总库存（普通仓库）
	GetMaterialStock(ctx context.Context, materialUuid uint64) (float64, error)

	// GetMaterialStockBatch 批量获取物品总库存（普通仓库）
	GetMaterialStockBatch(ctx context.Context, materialUuids []uint64) (map[uint64]float64, error)

	// GetMaterialStockByWarehouse 获取物品在各仓库的库存详情
	GetMaterialStockByWarehouse(ctx context.Context, materialUuid uint64) ([]repository.WarehouseStockInfo, error)

	// TransferStock 仓库间调拨库存
	TransferStock(ctx context.Context, fromWarehouseUuid, toWarehouseUuid, materialUuid uint64, quantity float64) error

	// ReserveStock 预留库存
	ReserveStock(ctx context.Context, warehouseUuid, materialUuid uint64, quantity float64) error

	// ReleaseReservedStock 释放预留库存
	ReleaseReservedStock(ctx context.Context, warehouseUuid, materialUuid uint64, quantity float64) error

	// ConsumeStock 消耗库存（出库）
	ConsumeStock(ctx context.Context, warehouseUuid, materialUuid uint64, quantity float64) error

	// AddStock 增加库存（入库）
	AddStock(ctx context.Context, warehouseUuid, materialUuid uint64, materialCode string, quantity float64, valuation float64) error
}

// warehouseItemDomainService 库存物品领域服务实现
type warehouseItemDomainService struct {
	warehouseItemRepo repository.IWarehouseItemRepository
}

// NewWarehouseItemDomainService 创建库存物品领域服务
func NewWarehouseItemDomainService(warehouseItemRepo repository.IWarehouseItemRepository) IWarehouseItemDomainService {
	return &warehouseItemDomainService{
		warehouseItemRepo: warehouseItemRepo,
	}
}

// GetMaterialStock 获取物品总库存（普通仓库）
func (s *warehouseItemDomainService) GetMaterialStock(ctx context.Context, materialUuid uint64) (float64, error) {
	stockMap, err := s.warehouseItemRepo.GetMaterialStockInNormalWarehouses(ctx, []uint64{materialUuid})
	if err != nil {
		return 0, errors.WithMessage(err, "获取物品库存失败")
	}

	if stock, ok := stockMap[materialUuid]; ok {
		return stock, nil
	}

	return 0, nil
}

// GetMaterialStockBatch 批量获取物品总库存（普通仓库）
func (s *warehouseItemDomainService) GetMaterialStockBatch(ctx context.Context, materialUuids []uint64) (map[uint64]float64, error) {
	if len(materialUuids) == 0 {
		return make(map[uint64]float64), nil
	}

	stockMap, err := s.warehouseItemRepo.GetMaterialStockInNormalWarehouses(ctx, materialUuids)
	if err != nil {
		return nil, errors.WithMessage(err, "批量获取物品库存失败")
	}

	return stockMap, nil
}

// GetMaterialStockByWarehouse 获取物品在各仓库的库存详情
func (s *warehouseItemDomainService) GetMaterialStockByWarehouse(ctx context.Context, materialUuid uint64) ([]repository.WarehouseStockInfo, error) {
	stockList, err := s.warehouseItemRepo.GetMaterialStockByWarehouse(ctx, materialUuid)
	if err != nil {
		return nil, errors.WithMessage(err, "获取物品仓库库存详情失败")
	}

	return stockList, nil
}

// TransferStock 仓库间调拨库存
func (s *warehouseItemDomainService) TransferStock(ctx context.Context, fromWarehouseUuid, toWarehouseUuid, materialUuid uint64, quantity float64) error {
	if quantity <= 0 {
		return errors.New("调拨数量必须大于0")
	}

	// 获取源仓库库存
	fromItem, err := s.warehouseItemRepo.FindByWarehouseAndMaterial(ctx, fromWarehouseUuid, materialUuid)
	if err != nil {
		return errors.WithMessage(err, "获取源仓库库存失败")
	}
	if fromItem == nil {
		return errors.New("源仓库没有该物品库存")
	}

	// 检查可用库存是否足够
	if fromItem.AvailableStock().Value() < quantity {
		return errors.New("源仓库可用库存不足")
	}

	// 减少源仓库库存
	if err := fromItem.ReduceStock(quantity); err != nil {
		return err
	}
	if err := s.warehouseItemRepo.Save(ctx, fromItem); err != nil {
		return errors.WithMessage(err, "更新源仓库库存失败")
	}

	// 获取或创建目标仓库库存记录
	toItem, err := s.warehouseItemRepo.FindOrCreate(ctx, toWarehouseUuid, materialUuid, fromItem.MaterialCode(), fromItem.Valuation())
	if err != nil {
		return errors.WithMessage(err, "获取目标仓库库存失败")
	}

	// 增加目标仓库库存
	if err := toItem.AddStock(quantity); err != nil {
		return err
	}
	if err := s.warehouseItemRepo.Save(ctx, toItem); err != nil {
		return errors.WithMessage(err, "更新目标仓库库存失败")
	}

	return nil
}

// ReserveStock 预留库存
func (s *warehouseItemDomainService) ReserveStock(ctx context.Context, warehouseUuid, materialUuid uint64, quantity float64) error {
	if quantity <= 0 {
		return errors.New("预留数量必须大于0")
	}

	item, err := s.warehouseItemRepo.FindByWarehouseAndMaterial(ctx, warehouseUuid, materialUuid)
	if err != nil {
		return errors.WithMessage(err, "获取库存失败")
	}
	if item == nil {
		return errors.New("该仓库没有该物品库存")
	}

	// 预留库存
	if err := item.ReserveStock(quantity); err != nil {
		return err
	}

	if err := s.warehouseItemRepo.Save(ctx, item); err != nil {
		return errors.WithMessage(err, "更新预留库存失败")
	}

	return nil
}

// ReleaseReservedStock 释放预留库存
func (s *warehouseItemDomainService) ReleaseReservedStock(ctx context.Context, warehouseUuid, materialUuid uint64, quantity float64) error {
	if quantity <= 0 {
		return errors.New("释放数量必须大于0")
	}

	item, err := s.warehouseItemRepo.FindByWarehouseAndMaterial(ctx, warehouseUuid, materialUuid)
	if err != nil {
		return errors.WithMessage(err, "获取库存失败")
	}
	if item == nil {
		return errors.New("该仓库没有该物品库存")
	}

	// 释放预留库存
	if err := item.ReleaseReservedStock(quantity); err != nil {
		return err
	}

	if err := s.warehouseItemRepo.Save(ctx, item); err != nil {
		return errors.WithMessage(err, "更新预留库存失败")
	}

	return nil
}

// ConsumeStock 消耗库存（出库）
func (s *warehouseItemDomainService) ConsumeStock(ctx context.Context, warehouseUuid, materialUuid uint64, quantity float64) error {
	if quantity <= 0 {
		return errors.New("消耗数量必须大于0")
	}

	item, err := s.warehouseItemRepo.FindByWarehouseAndMaterial(ctx, warehouseUuid, materialUuid)
	if err != nil {
		return errors.WithMessage(err, "获取库存失败")
	}
	if item == nil {
		return errors.New("该仓库没有该物品库存")
	}

	// 减少库存
	if err := item.ReduceStock(quantity); err != nil {
		return err
	}

	if err := s.warehouseItemRepo.Save(ctx, item); err != nil {
		return errors.WithMessage(err, "更新库存失败")
	}

	return nil
}

// AddStock 增加库存（入库）
func (s *warehouseItemDomainService) AddStock(ctx context.Context, warehouseUuid, materialUuid uint64, materialCode string, quantity float64, valuation float64) error {
	if quantity <= 0 {
		return errors.New("入库数量必须大于0")
	}

	// 获取或创建库存记录
	item, err := s.warehouseItemRepo.FindOrCreate(ctx, warehouseUuid, materialUuid, materialCode, valuation)
	if err != nil {
		return errors.WithMessage(err, "获取库存记录失败")
	}

	// 增加库存
	if err := item.AddStock(quantity); err != nil {
		return err
	}

	if err := s.warehouseItemRepo.Save(ctx, item); err != nil {
		return errors.WithMessage(err, "更新库存失败")
	}

	return nil
}
