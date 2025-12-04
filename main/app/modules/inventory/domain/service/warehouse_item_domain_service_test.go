package service

import (
	"testing"

	"ttpos-server-go/app/modules/inventory/domain/entity"
	"ttpos-server-go/app/modules/inventory/domain/repository"
	"ttpos-server-go/pkg/context"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockWarehouseItemRepository 模拟仓储接口
type mockWarehouseItemRepository struct {
	mock.Mock
}

func (m *mockWarehouseItemRepository) Save(ctx context.Context, item *entity.WarehouseItem) error {
	args := m.Called(ctx, item)
	return args.Error(0)
}

func (m *mockWarehouseItemRepository) FindByUuid(ctx context.Context, uuid uint64) (*entity.WarehouseItem, error) {
	args := m.Called(ctx, uuid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.WarehouseItem), args.Error(1)
}

func (m *mockWarehouseItemRepository) FindByWarehouseAndMaterial(ctx context.Context, warehouseUuid, materialUuid uint64) (*entity.WarehouseItem, error) {
	args := m.Called(ctx, warehouseUuid, materialUuid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.WarehouseItem), args.Error(1)
}

func (m *mockWarehouseItemRepository) FindByMaterialUuid(ctx context.Context, materialUuid uint64) ([]*entity.WarehouseItem, error) {
	args := m.Called(ctx, materialUuid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.WarehouseItem), args.Error(1)
}

func (m *mockWarehouseItemRepository) FindByWarehouseUuid(ctx context.Context, warehouseUuid uint64) ([]*entity.WarehouseItem, error) {
	args := m.Called(ctx, warehouseUuid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.WarehouseItem), args.Error(1)
}

func (m *mockWarehouseItemRepository) FindOrCreate(ctx context.Context, warehouseUuid, materialUuid uint64, materialCode string, valuation float64) (*entity.WarehouseItem, error) {
	args := m.Called(ctx, warehouseUuid, materialUuid, materialCode, valuation)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.WarehouseItem), args.Error(1)
}

func (m *mockWarehouseItemRepository) Remove(ctx context.Context, uuid uint64) error {
	args := m.Called(ctx, uuid)
	return args.Error(0)
}

func (m *mockWarehouseItemRepository) GetMaterialStockInNormalWarehouses(ctx context.Context, materialUuids []uint64) (map[uint64]float64, error) {
	args := m.Called(ctx, materialUuids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[uint64]float64), args.Error(1)
}

func (m *mockWarehouseItemRepository) GetMaterialStockByWarehouse(ctx context.Context, materialUuid uint64) ([]repository.WarehouseStockInfo, error) {
	args := m.Called(ctx, materialUuid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]repository.WarehouseStockInfo), args.Error(1)
}

func (m *mockWarehouseItemRepository) FindWithPagination(ctx context.Context, spec *repository.WarehouseItemQuerySpec, pageNo, pageSize int) ([]*entity.WarehouseItem, int64, error) {
	args := m.Called(ctx, spec, pageNo, pageSize)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*entity.WarehouseItem), args.Get(1).(int64), args.Error(2)
}

func (m *mockWarehouseItemRepository) BatchUpdateStock(ctx context.Context, items []*entity.WarehouseItem) error {
	args := m.Called(ctx, items)
	return args.Error(0)
}

func (m *mockWarehouseItemRepository) AddStock(ctx context.Context, uuid uint64, quantity float64) error {
	args := m.Called(ctx, uuid, quantity)
	return args.Error(0)
}

func (m *mockWarehouseItemRepository) ReduceStock(ctx context.Context, uuid uint64, quantity float64) error {
	args := m.Called(ctx, uuid, quantity)
	return args.Error(0)
}

func (m *mockWarehouseItemRepository) AddReservedStock(ctx context.Context, uuid uint64, quantity float64) error {
	args := m.Called(ctx, uuid, quantity)
	return args.Error(0)
}

func (m *mockWarehouseItemRepository) ReduceReservedStock(ctx context.Context, uuid uint64, quantity float64) error {
	args := m.Called(ctx, uuid, quantity)
	return args.Error(0)
}

// ========== 测试 GetMaterialStock ==========

func TestGetMaterialStock_Success(t *testing.T) {
	repo := new(mockWarehouseItemRepository)
	service := NewWarehouseItemDomainService(repo)
	ctx := NewTestContext()

	materialUuid := uint64(100)
	stockMap := map[uint64]float64{materialUuid: 150.5}

	repo.On("GetMaterialStockInNormalWarehouses", mock.Anything, []uint64{materialUuid}).Return(stockMap, nil)

	stock, err := service.GetMaterialStock(ctx, materialUuid)

	assert.NoError(t, err)
	assert.Equal(t, 150.5, stock)
	repo.AssertExpectations(t)
}

func TestGetMaterialStock_NotFound(t *testing.T) {
	repo := new(mockWarehouseItemRepository)
	service := NewWarehouseItemDomainService(repo)
	ctx := NewTestContext()

	materialUuid := uint64(100)
	stockMap := map[uint64]float64{} // 空map

	repo.On("GetMaterialStockInNormalWarehouses", mock.Anything, []uint64{materialUuid}).Return(stockMap, nil)

	stock, err := service.GetMaterialStock(ctx, materialUuid)

	assert.NoError(t, err)
	assert.Equal(t, 0.0, stock)
	repo.AssertExpectations(t)
}

// ========== 测试 GetMaterialStockBatch ==========

func TestGetMaterialStockBatch_Success(t *testing.T) {
	repo := new(mockWarehouseItemRepository)
	service := NewWarehouseItemDomainService(repo)
	ctx := NewTestContext()

	materialUuids := []uint64{100, 101, 102}
	stockMap := map[uint64]float64{
		100: 150.5,
		101: 200.0,
		102: 75.25,
	}

	repo.On("GetMaterialStockInNormalWarehouses", mock.Anything, materialUuids).Return(stockMap, nil)

	result, err := service.GetMaterialStockBatch(ctx, materialUuids)

	assert.NoError(t, err)
	assert.Equal(t, 150.5, result[100])
	assert.Equal(t, 200.0, result[101])
	assert.Equal(t, 75.25, result[102])
	repo.AssertExpectations(t)
}

func TestGetMaterialStockBatch_EmptyInput(t *testing.T) {
	repo := new(mockWarehouseItemRepository)
	service := NewWarehouseItemDomainService(repo)
	ctx := NewTestContext()

	result, err := service.GetMaterialStockBatch(ctx, []uint64{})

	assert.NoError(t, err)
	assert.Empty(t, result)
	// 不应该调用 repository
	repo.AssertNotCalled(t, "GetMaterialStockInNormalWarehouses")
}

// ========== 测试 GetMaterialStockByWarehouse ==========

func TestGetMaterialStockByWarehouse_Success(t *testing.T) {
	repo := new(mockWarehouseItemRepository)
	service := NewWarehouseItemDomainService(repo)
	ctx := NewTestContext()

	materialUuid := uint64(100)
	stockList := []repository.WarehouseStockInfo{
		{WarehouseUuid: 1, WarehouseName: "主仓库", Stock: 100.0, ReservedStock: 20.0},
		{WarehouseUuid: 2, WarehouseName: "分仓库", Stock: 50.0, ReservedStock: 10.0},
	}

	repo.On("GetMaterialStockByWarehouse", mock.Anything, materialUuid).Return(stockList, nil)

	result, err := service.GetMaterialStockByWarehouse(ctx, materialUuid)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, uint64(1), result[0].WarehouseUuid)
	assert.Equal(t, 100.0, result[0].Stock)
	repo.AssertExpectations(t)
}

// ========== 测试 TransferStock ==========

func TestTransferStock_Success(t *testing.T) {
	repo := new(mockWarehouseItemRepository)
	service := NewWarehouseItemDomainService(repo)
	ctx := NewTestContext()

	fromWarehouseUuid := uint64(1)
	toWarehouseUuid := uint64(2)
	materialUuid := uint64(100)
	quantity := 30.0

	// 源仓库有 100 库存
	fromItem := entity.ReconstructWarehouseItem(1001, fromWarehouseUuid, materialUuid, "MAT001", 100.0, 0, 10.5, 0, 0)
	// 目标仓库新建
	toItem := entity.NewWarehouseItem(toWarehouseUuid, materialUuid, "MAT001", 10.5)

	repo.On("FindByWarehouseAndMaterial", mock.Anything, fromWarehouseUuid, materialUuid).Return(fromItem, nil)
	repo.On("Save", mock.Anything, mock.AnythingOfType("*entity.WarehouseItem")).Return(nil).Twice()
	repo.On("FindOrCreate", mock.Anything, toWarehouseUuid, materialUuid, "MAT001", 10.5).Return(toItem, nil)

	err := service.TransferStock(ctx, fromWarehouseUuid, toWarehouseUuid, materialUuid, quantity)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestTransferStock_ZeroQuantity(t *testing.T) {
	repo := new(mockWarehouseItemRepository)
	service := NewWarehouseItemDomainService(repo)
	ctx := NewTestContext()

	err := service.TransferStock(ctx, 1, 2, 100, 0)

	assert.Error(t, err)
	assert.Equal(t, "调拨数量必须大于0", err.Error())
}

func TestTransferStock_InsufficientStock(t *testing.T) {
	repo := new(mockWarehouseItemRepository)
	service := NewWarehouseItemDomainService(repo)
	ctx := NewTestContext()

	fromWarehouseUuid := uint64(1)
	materialUuid := uint64(100)

	// 源仓库只有 20 库存
	fromItem := entity.ReconstructWarehouseItem(1001, fromWarehouseUuid, materialUuid, "MAT001", 20.0, 0, 10.5, 0, 0)

	repo.On("FindByWarehouseAndMaterial", mock.Anything, fromWarehouseUuid, materialUuid).Return(fromItem, nil)

	err := service.TransferStock(ctx, fromWarehouseUuid, 2, materialUuid, 50.0)

	assert.Error(t, err)
	assert.Equal(t, "源仓库可用库存不足", err.Error())
}

func TestTransferStock_SourceNotFound(t *testing.T) {
	repo := new(mockWarehouseItemRepository)
	service := NewWarehouseItemDomainService(repo)
	ctx := NewTestContext()

	repo.On("FindByWarehouseAndMaterial", mock.Anything, uint64(1), uint64(100)).Return(nil, nil)

	err := service.TransferStock(ctx, 1, 2, 100, 30.0)

	assert.Error(t, err)
	assert.Equal(t, "源仓库没有该物品库存", err.Error())
}

// ========== 测试 ReserveStock ==========

func TestReserveStock_Success(t *testing.T) {
	repo := new(mockWarehouseItemRepository)
	service := NewWarehouseItemDomainService(repo)
	ctx := NewTestContext()

	warehouseUuid := uint64(1)
	materialUuid := uint64(100)

	item := entity.ReconstructWarehouseItem(1001, warehouseUuid, materialUuid, "MAT001", 100.0, 0, 10.5, 0, 0)

	repo.On("FindByWarehouseAndMaterial", mock.Anything, warehouseUuid, materialUuid).Return(item, nil)
	repo.On("Save", mock.Anything, mock.AnythingOfType("*entity.WarehouseItem")).Return(nil)

	err := service.ReserveStock(ctx, warehouseUuid, materialUuid, 30.0)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestReserveStock_ZeroQuantity(t *testing.T) {
	repo := new(mockWarehouseItemRepository)
	service := NewWarehouseItemDomainService(repo)
	ctx := NewTestContext()

	err := service.ReserveStock(ctx, 1, 100, 0)

	assert.Error(t, err)
	assert.Equal(t, "预留数量必须大于0", err.Error())
}

func TestReserveStock_NotFound(t *testing.T) {
	repo := new(mockWarehouseItemRepository)
	service := NewWarehouseItemDomainService(repo)
	ctx := NewTestContext()

	repo.On("FindByWarehouseAndMaterial", mock.Anything, uint64(1), uint64(100)).Return(nil, nil)

	err := service.ReserveStock(ctx, 1, 100, 30.0)

	assert.Error(t, err)
	assert.Equal(t, "该仓库没有该物品库存", err.Error())
}

// ========== 测试 ReleaseReservedStock ==========

func TestReleaseReservedStock_Success(t *testing.T) {
	repo := new(mockWarehouseItemRepository)
	service := NewWarehouseItemDomainService(repo)
	ctx := NewTestContext()

	warehouseUuid := uint64(1)
	materialUuid := uint64(100)

	// 有 50 预留库存
	item := entity.ReconstructWarehouseItem(1001, warehouseUuid, materialUuid, "MAT001", 100.0, 50.0, 10.5, 0, 0)

	repo.On("FindByWarehouseAndMaterial", mock.Anything, warehouseUuid, materialUuid).Return(item, nil)
	repo.On("Save", mock.Anything, mock.AnythingOfType("*entity.WarehouseItem")).Return(nil)

	err := service.ReleaseReservedStock(ctx, warehouseUuid, materialUuid, 30.0)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestReleaseReservedStock_ZeroQuantity(t *testing.T) {
	repo := new(mockWarehouseItemRepository)
	service := NewWarehouseItemDomainService(repo)
	ctx := NewTestContext()

	err := service.ReleaseReservedStock(ctx, 1, 100, 0)

	assert.Error(t, err)
	assert.Equal(t, "释放数量必须大于0", err.Error())
}

// ========== 测试 ConsumeStock ==========

func TestConsumeStock_Success(t *testing.T) {
	repo := new(mockWarehouseItemRepository)
	service := NewWarehouseItemDomainService(repo)
	ctx := NewTestContext()

	warehouseUuid := uint64(1)
	materialUuid := uint64(100)

	item := entity.ReconstructWarehouseItem(1001, warehouseUuid, materialUuid, "MAT001", 100.0, 0, 10.5, 0, 0)

	repo.On("FindByWarehouseAndMaterial", mock.Anything, warehouseUuid, materialUuid).Return(item, nil)
	repo.On("Save", mock.Anything, mock.AnythingOfType("*entity.WarehouseItem")).Return(nil)

	err := service.ConsumeStock(ctx, warehouseUuid, materialUuid, 30.0)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestConsumeStock_ZeroQuantity(t *testing.T) {
	repo := new(mockWarehouseItemRepository)
	service := NewWarehouseItemDomainService(repo)
	ctx := NewTestContext()

	err := service.ConsumeStock(ctx, 1, 100, 0)

	assert.Error(t, err)
	assert.Equal(t, "消耗数量必须大于0", err.Error())
}

func TestConsumeStock_Insufficient(t *testing.T) {
	repo := new(mockWarehouseItemRepository)
	service := NewWarehouseItemDomainService(repo)
	ctx := NewTestContext()

	warehouseUuid := uint64(1)
	materialUuid := uint64(100)

	// 只有 20 库存
	item := entity.ReconstructWarehouseItem(1001, warehouseUuid, materialUuid, "MAT001", 20.0, 0, 10.5, 0, 0)

	repo.On("FindByWarehouseAndMaterial", mock.Anything, warehouseUuid, materialUuid).Return(item, nil)

	err := service.ConsumeStock(ctx, warehouseUuid, materialUuid, 50.0)

	assert.Error(t, err)
	assert.Equal(t, "库存不足", err.Error())
}

// ========== 测试 AddStock ==========

func TestAddStock_Success(t *testing.T) {
	repo := new(mockWarehouseItemRepository)
	service := NewWarehouseItemDomainService(repo)
	ctx := NewTestContext()

	warehouseUuid := uint64(1)
	materialUuid := uint64(100)
	materialCode := "MAT001"
	valuation := 10.5

	item := entity.NewWarehouseItem(warehouseUuid, materialUuid, materialCode, valuation)

	repo.On("FindOrCreate", mock.Anything, warehouseUuid, materialUuid, materialCode, valuation).Return(item, nil)
	repo.On("Save", mock.Anything, mock.AnythingOfType("*entity.WarehouseItem")).Return(nil)

	err := service.AddStock(ctx, warehouseUuid, materialUuid, materialCode, 50.0, valuation)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestAddStock_ZeroQuantity(t *testing.T) {
	repo := new(mockWarehouseItemRepository)
	service := NewWarehouseItemDomainService(repo)
	ctx := NewTestContext()

	err := service.AddStock(ctx, 1, 100, "MAT001", 0, 10.5)

	assert.Error(t, err)
	assert.Equal(t, "入库数量必须大于0", err.Error())
}

func TestAddStock_NegativeQuantity(t *testing.T) {
	repo := new(mockWarehouseItemRepository)
	service := NewWarehouseItemDomainService(repo)
	ctx := NewTestContext()

	err := service.AddStock(ctx, 1, 100, "MAT001", -10.0, 10.5)

	assert.Error(t, err)
	assert.Equal(t, "入库数量必须大于0", err.Error())
}
