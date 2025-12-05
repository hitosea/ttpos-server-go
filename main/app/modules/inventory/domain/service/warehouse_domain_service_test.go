package service

import (
	"testing"

	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/modules/inventory/domain/entity"
	"ttpos-server-go/app/modules/inventory/domain/repository"
	"ttpos-server-go/app/modules/inventory/domain/valueobject"
	"ttpos-server-go/pkg/context"
)

// mockWarehouseRepository 模拟仓库仓储
type mockWarehouseRepository struct {
	existsCodeFunc   func(ctx context.Context, code valueobject.WarehouseCode, excludeUuid uint64) (bool, error)
	findByUuidFunc   func(ctx context.Context, uuid uint64) (*entity.Warehouse, error)
	setAsDefaultFunc func(ctx context.Context, uuid uint64) error
}

func (m *mockWarehouseRepository) Save(ctx context.Context, warehouse *entity.Warehouse) error {
	return nil
}

func (m *mockWarehouseRepository) FindByUuid(ctx context.Context, uuid uint64) (*entity.Warehouse, error) {
	if m.findByUuidFunc != nil {
		return m.findByUuidFunc(ctx, uuid)
	}
	return nil, nil
}

func (m *mockWarehouseRepository) FindByCode(ctx context.Context, code valueobject.WarehouseCode) (*entity.Warehouse, error) {
	return nil, nil
}

func (m *mockWarehouseRepository) FindByErpCode(ctx context.Context, erpCode string) (*entity.Warehouse, error) {
	return nil, nil
}

func (m *mockWarehouseRepository) Remove(ctx context.Context, uuid uint64) error {
	return nil
}

func (m *mockWarehouseRepository) ExistsCode(ctx context.Context, code valueobject.WarehouseCode, excludeUuid uint64) (bool, error) {
	if m.existsCodeFunc != nil {
		return m.existsCodeFunc(ctx, code, excludeUuid)
	}
	return false, nil
}

func (m *mockWarehouseRepository) FindDefault(ctx context.Context) (*entity.Warehouse, error) {
	return nil, nil
}

func (m *mockWarehouseRepository) FindAllNormal(ctx context.Context) ([]*entity.Warehouse, error) {
	return nil, nil
}

func (m *mockWarehouseRepository) FindTransit(ctx context.Context) (*entity.Warehouse, error) {
	return nil, nil
}

func (m *mockWarehouseRepository) FindByHeadquarterUuid(ctx context.Context, headquarterUuid uint64) ([]*entity.Warehouse, error) {
	return nil, nil
}

func (m *mockWarehouseRepository) FindWithPagination(ctx context.Context, spec repository.WarehouseSpecification, pageNo, pageSize int) ([]*entity.Warehouse, int64, error) {
	return nil, 0, nil
}

func (m *mockWarehouseRepository) SetAsDefault(ctx context.Context, uuid uint64) error {
	if m.setAsDefaultFunc != nil {
		return m.setAsDefaultFunc(ctx, uuid)
	}
	return nil
}

// TestCreateWarehouse_Success 测试成功创建仓库
func TestCreateWarehouse_Success(t *testing.T) {
	repo := &mockWarehouseRepository{
		existsCodeFunc: func(ctx context.Context, code valueobject.WarehouseCode, excludeUuid uint64) (bool, error) {
			return false, nil // 编码不存在
		},
	}

	service := NewWarehouseDomainService(repo)

	code, _ := valueobject.NewWarehouseCode("WH001")
	req := CreateWarehouseRequest{
		Name:        valueobject.NewMultiLanguageName(dto.LocaleResponse{ZH: "测试仓库"}),
		Type:        valueobject.NewWarehouseType("normal"),
		Code:        code,
		Status:      valueobject.NewWarehouseStatus(1),
		ContactInfo: valueobject.NewContactInfo("张三", "13800138000", "北京市"),
	}

	mockCtx := NewTestContext()
	warehouse, err := service.CreateWarehouse(mockCtx, req)

	if err != nil {
		t.Errorf("CreateWarehouse 失败: %v", err)
	}
	if warehouse == nil {
		t.Error("CreateWarehouse 返回的仓库为空")
	}
	if warehouse.Code().Value() != "WH001" {
		t.Errorf("仓库编码不匹配，期望 WH001，实际 %s", warehouse.Code().Value())
	}
}

// TestCreateWarehouse_CodeExists 测试编码已存在
func TestCreateWarehouse_CodeExists(t *testing.T) {
	repo := &mockWarehouseRepository{
		existsCodeFunc: func(ctx context.Context, code valueobject.WarehouseCode, excludeUuid uint64) (bool, error) {
			return true, nil // 编码已存在
		},
	}

	service := NewWarehouseDomainService(repo)

	code, _ := valueobject.NewWarehouseCode("WH001")
	req := CreateWarehouseRequest{
		Name:        valueobject.NewMultiLanguageName(dto.LocaleResponse{ZH: "测试仓库"}),
		Type:        valueobject.NewWarehouseType("normal"),
		Code:        code,
		Status:      valueobject.NewWarehouseStatus(1),
		ContactInfo: valueobject.NewContactInfo("张三", "13800138000", "北京市"),
	}

	mockCtx := NewTestContext()
	warehouse, err := service.CreateWarehouse(mockCtx, req)

	if err == nil {
		t.Error("期望返回错误，但没有")
	}
	if warehouse != nil {
		t.Error("期望返回的仓库为空，但不为空")
	}
}

// TestValidateForUpdate_CodeChanged 测试更新时编码变更验证
func TestValidateForUpdate_CodeChanged(t *testing.T) {
	repo := &mockWarehouseRepository{
		existsCodeFunc: func(ctx context.Context, code valueobject.WarehouseCode, excludeUuid uint64) (bool, error) {
			return true, nil // 新编码已存在
		},
	}

	service := NewWarehouseDomainService(repo)

	// 创建一个现有仓库
	oldCode, _ := valueobject.NewWarehouseCode("WH001")
	warehouse := entity.NewWarehouse(
		valueobject.NewMultiLanguageName(dto.LocaleResponse{ZH: "测试仓库"}),
		valueobject.NewWarehouseType("normal"),
		oldCode,
		valueobject.NewWarehouseStatus(1),
		valueobject.NewContactInfo("张三", "13800138000", "北京市"),
	)

	// 尝试更新为新编码
	newCode, _ := valueobject.NewWarehouseCode("WH002")
	req := UpdateWarehouseRequest{
		Name:        valueobject.NewMultiLanguageName(dto.LocaleResponse{ZH: "测试仓库"}),
		Type:        valueobject.NewWarehouseType("normal"),
		Code:        newCode,
		Status:      valueobject.NewWarehouseStatus(1),
		ContactInfo: valueobject.NewContactInfo("张三", "13800138000", "北京市"),
	}

	mockCtx := NewTestContext()
	err := service.ValidateForUpdate(mockCtx, warehouse, req)

	if err == nil {
		t.Error("期望返回错误，但没有")
	}
}

// TestValidateForUpdate_DisableDefaultWarehouse 测试禁用默认仓库
func TestValidateForUpdate_DisableDefaultWarehouse(t *testing.T) {
	repo := &mockWarehouseRepository{}

	service := NewWarehouseDomainService(repo)

	// 创建一个默认仓库
	code, _ := valueobject.NewWarehouseCode("WH001")
	warehouse := entity.NewWarehouse(
		valueobject.NewMultiLanguageName(dto.LocaleResponse{ZH: "测试仓库"}),
		valueobject.NewWarehouseType("normal"),
		code,
		valueobject.NewWarehouseStatus(1),
		valueobject.NewContactInfo("张三", "13800138000", "北京市"),
	)
	warehouse.SetAsDefault()

	// 尝试禁用
	req := UpdateWarehouseRequest{
		Name:        valueobject.NewMultiLanguageName(dto.LocaleResponse{ZH: "测试仓库"}),
		Type:        valueobject.NewWarehouseType("normal"),
		Code:        code,
		Status:      valueobject.NewWarehouseStatus(0), // 禁用
		ContactInfo: valueobject.NewContactInfo("张三", "13800138000", "北京市"),
	}

	mockCtx := NewTestContext()
	err := service.ValidateForUpdate(mockCtx, warehouse, req)

	if err == nil {
		t.Error("期望返回错误，但没有")
	}
}

// TestValidateForDelete_DefaultWarehouse 测试删除默认仓库
func TestValidateForDelete_DefaultWarehouse(t *testing.T) {
	repo := &mockWarehouseRepository{}

	service := NewWarehouseDomainService(repo)

	// 创建一个默认仓库
	code, _ := valueobject.NewWarehouseCode("WH001")
	warehouse := entity.NewWarehouse(
		valueobject.NewMultiLanguageName(dto.LocaleResponse{ZH: "测试仓库"}),
		valueobject.NewWarehouseType("normal"),
		code,
		valueobject.NewWarehouseStatus(1),
		valueobject.NewContactInfo("张三", "13800138000", "北京市"),
	)
	warehouse.SetAsDefault()

	mockCtx := NewTestContext()
	err := service.ValidateForDelete(mockCtx, warehouse, 1)

	if err == nil {
		t.Error("期望返回错误，但没有")
	}
}

// TestSetDefaultWarehouse_Success 测试成功设置默认仓库
func TestSetDefaultWarehouse_Success(t *testing.T) {
	code, _ := valueobject.NewWarehouseCode("WH001")
	warehouse := entity.NewWarehouse(
		valueobject.NewMultiLanguageName(dto.LocaleResponse{ZH: "测试仓库"}),
		valueobject.NewWarehouseType("normal"),
		code,
		valueobject.NewWarehouseStatus(1),
		valueobject.NewContactInfo("张三", "13800138000", "北京市"),
	)

	repo := &mockWarehouseRepository{
		findByUuidFunc: func(ctx context.Context, uuid uint64) (*entity.Warehouse, error) {
			return warehouse, nil
		},
		setAsDefaultFunc: func(ctx context.Context, uuid uint64) error {
			return nil
		},
	}

	service := NewWarehouseDomainService(repo)

	mockCtx := NewTestContext()
	err := service.SetDefaultWarehouse(mockCtx, 1)

	if err != nil {
		t.Errorf("SetDefaultWarehouse 失败: %v", err)
	}
}

// TestSetDefaultWarehouse_TransitWarehouse 测试设置在途仓库为默认仓库
func TestSetDefaultWarehouse_TransitWarehouse(t *testing.T) {
	code, _ := valueobject.NewWarehouseCode("WH001")
	warehouse := entity.NewWarehouse(
		valueobject.NewMultiLanguageName(dto.LocaleResponse{ZH: "在途仓库"}),
		valueobject.NewWarehouseType("transit"), // 在途仓库
		code,
		valueobject.NewWarehouseStatus(1),
		valueobject.NewContactInfo("张三", "13800138000", "北京市"),
	)

	repo := &mockWarehouseRepository{
		findByUuidFunc: func(ctx context.Context, uuid uint64) (*entity.Warehouse, error) {
			return warehouse, nil
		},
	}

	service := NewWarehouseDomainService(repo)

	mockCtx := NewTestContext()
	err := service.SetDefaultWarehouse(mockCtx, 1)

	if err == nil {
		t.Error("期望返回错误，但没有")
	}
}
