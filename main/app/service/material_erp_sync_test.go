package service

import (
	"fmt"
	"testing"

	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	pkgctx "ttpos-server-go/pkg/context"

	"gorm.io/gorm"
)

// --- Mock implementations ---

// mockProductRepo implements repository.IProductRepo for unit testing
type mockProductRepo struct {
	repository.IProductRepo
	units map[string]*model.ProductUnit
}

func (m *mockProductRepo) GetProductUnitByErpnextUom(uom string) (*model.ProductUnit, error) {
	if u, ok := m.units[uom]; ok {
		return u, nil
	}
	return nil, fmt.Errorf("unit not found: %s", uom)
}

// mockProductUnitRepo implements repository.IProductUnitRepo for unit testing
type mockProductUnitRepo struct {
	repository.IProductUnitRepo
	units map[string]*model.ProductUnit
}

func (m *mockProductUnitRepo) GetProductUnitByErpnextUom(uom string) (*model.ProductUnit, error) {
	if u, ok := m.units[uom]; ok {
		return u, nil
	}
	return nil, fmt.Errorf("unit not found: %s", uom)
}

// mockMaterialUnitRepo implements repository.IMaterialUnitRepo for unit testing
type mockMaterialUnitRepo struct {
	units         map[uint64]model.MaterialUnit // unit_uuid → MaterialUnit
	unitList      []*model.MaterialUnit         // for GetMaterialUnitListByBaseUnitUuid
	unitListErr   error                         // error for GetMaterialUnitListByBaseUnitUuid
	createdUnits  []model.MaterialUnit
	updatedData   []map[string]any
	destroyCalled bool
	updateErr     error // error for UpdateMaterialUnit
	createErr     error // error for CreateMaterialUnit
	destroyErr    error // error for DestroyMaterialUnit
}

func (m *mockMaterialUnitRepo) GetMaterialUnit(opts ...repository.DBOption) model.MaterialUnit {
	// Return first matching unit (simplified)
	for _, mu := range m.units {
		return mu
	}
	return model.MaterialUnit{}
}

func (m *mockMaterialUnitRepo) GetMaterialUnitByUuid(uuid uint64, opts ...repository.DBOption) (model.MaterialUnit, error) {
	if mu, ok := m.units[uuid]; ok {
		return mu, nil
	}
	return model.MaterialUnit{}, fmt.Errorf("not found")
}

func (m *mockMaterialUnitRepo) GetMaterialUnitsByUuid(uuid uint64, opts ...repository.DBOption) (model.MaterialUnit, error) {
	return m.GetMaterialUnitByUuid(uuid, opts...)
}

func (m *mockMaterialUnitRepo) GetMaterialUnitList(opts ...repository.DBOption) ([]*model.MaterialUnit, error) {
	return nil, nil
}

func (m *mockMaterialUnitRepo) CreateMaterialUnit(mu model.MaterialUnit) (uint64, error) {
	if m.createErr != nil {
		return 0, m.createErr
	}
	m.createdUnits = append(m.createdUnits, mu)
	return 999, nil
}

func (m *mockMaterialUnitRepo) CreateMaterialUnitList(mus []model.MaterialUnit) error {
	return nil
}

func (m *mockMaterialUnitRepo) UpdateMaterialUnit(data map[string]any, opts ...repository.DBOption) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.updatedData = append(m.updatedData, data)
	return nil
}

func (m *mockMaterialUnitRepo) GetMaterialUnitListByBaseUnitUuid(baseUnitUuid uint64) ([]*model.MaterialUnit, error) {
	if m.unitListErr != nil {
		return nil, m.unitListErr
	}
	if m.unitList != nil {
		return m.unitList, nil
	}
	return nil, nil
}

func (m *mockMaterialUnitRepo) DestroyMaterialUnit(opts ...repository.DBOption) error {
	if m.destroyErr != nil {
		return m.destroyErr
	}
	m.destroyCalled = true
	return nil
}

// mockCommonRepo implements repository.ICommonRepo for unit testing (only the methods we need)
type mockCommonRepo struct {
	repository.ICommonRepo
}

func (m *mockCommonRepo) WhereByUuid(uuid uint64) repository.DBOption {
	return func(db *gorm.DB) *gorm.DB { return db }
}

func (m *mockCommonRepo) WhereByUnitUuid(unitUuid uint64) repository.DBOption {
	return func(db *gorm.DB) *gorm.DB { return db }
}

func (m *mockCommonRepo) WhereByMaterialUuid(materialUuid uint64) repository.DBOption {
	return func(db *gorm.DB) *gorm.DB { return db }
}

func (m *mockCommonRepo) WhereByIsDefault(isDefault uint) repository.DBOption {
	return func(db *gorm.DB) *gorm.DB { return db }
}

func (m *mockCommonRepo) WhereBySoftDelete() repository.DBOption {
	return func(db *gorm.DB) *gorm.DB { return db }
}

func (m *mockCommonRepo) WhereByUuidNotIn(uuids []uint64) repository.DBOption {
	return func(db *gorm.DB) *gorm.DB { return db }
}

// mockMaterialRepo implements repository.IMaterialRepo for unit testing
type mockMaterialRepo struct {
	repository.IMaterialRepo
	categories  map[string]*model.MaterialCategory
	categoryErr error // error for GetMaterialCategoryByCode
	materials   map[uint64]*model.Material
	updatedData []map[string]any
	updateErr   error // error for UpdateMaterialData
}

func (m *mockMaterialRepo) GetMaterialCategoryByCode(code string) (*model.MaterialCategory, bool, error) {
	if m.categoryErr != nil {
		return nil, false, m.categoryErr
	}
	if cat, ok := m.categories[code]; ok {
		return cat, true, nil
	}
	return nil, false, nil
}

func (m *mockMaterialRepo) GetMaterialDetailByUuid(uuid uint64) (*model.Material, error) {
	if mat, ok := m.materials[uuid]; ok {
		return mat, nil
	}
	return nil, fmt.Errorf("not found: %d", uuid)
}

func (m *mockMaterialRepo) GetMaterialDetailContainsDeletedByUuid(uuid uint64) (*model.Material, error) {
	if mat, ok := m.materials[uuid]; ok {
		return mat, nil
	}
	return nil, fmt.Errorf("not found: %d", uuid)
}

func (m *mockMaterialRepo) UpdateMaterialData(data map[string]any, opts ...repository.DBOption) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.updatedData = append(m.updatedData, data)
	return nil
}

// mockContext implements pkg/context.Context for unit testing
type mockContext struct {
	pkgctx.Context
	language       string
	companySetting model.CompanySetting
}

func (m *mockContext) GetLanguage() string                     { return m.language }
func (m *mockContext) GetCompanySetting() model.CompanySetting { return m.companySetting }

// --- Tests ---

func TestIsUomInConversion(t *testing.T) {
	uoms := []req.MaterialUomReq{
		{Uom: "A", ConversionRate: 1},
		{Uom: "B", ConversionRate: 2},
		{Uom: "C", ConversionRate: 3},
	}

	tests := []struct {
		name      string
		targetUom string
		stockUom  string
		uoms      []req.MaterialUomReq
		want      bool
	}{
		{"target equals stock uom", "X", "X", uoms, true},
		{"target in uoms list", "B", "X", uoms, true},
		{"target not in uoms and not stock", "D", "X", uoms, false},
		{"empty uoms, target equals stock", "X", "X", nil, true},
		{"empty uoms, target not stock", "D", "X", nil, false},
		{"empty target", "", "X", uoms, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isUomInConversion(tt.targetUom, tt.stockUom, tt.uoms)
			if got != tt.want {
				t.Errorf("isUomInConversion(%q, %q, ...) = %v, want %v", tt.targetUom, tt.stockUom, got, tt.want)
			}
		})
	}
}

func TestGetMaterialUnitName(t *testing.T) {
	tests := []struct {
		name string
		mu   *model.MaterialUnit
		lang string
		want string
	}{
		{"nil material unit", nil, "en", ""},
		{"nil product unit", &model.MaterialUnit{Unit: nil}, "en", ""},
		{"valid unit", &model.MaterialUnit{
			Unit: &model.ProductUnit{
				MultiLanguageName: model.MultiLanguageName{EnName: "Kilogram", ZhName: "千克"},
			},
		}, "en", "Kilogram"},
		{"valid unit zh", &model.MaterialUnit{
			Unit: &model.ProductUnit{
				MultiLanguageName: model.MultiLanguageName{EnName: "Kilogram", ZhName: "千克"},
			},
		}, "zh", "千克"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getMaterialUnitName(tt.mu, tt.lang)
			if got != tt.want {
				t.Errorf("getMaterialUnitName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetMaterialUnitLocaleName(t *testing.T) {
	t.Run("nil unit returns empty", func(t *testing.T) {
		result := getMaterialUnitLocaleName(nil)
		if result.EN != "" || result.ZH != "" {
			t.Error("expected empty LocaleResponse for nil unit")
		}
	})

	t.Run("valid unit returns locale", func(t *testing.T) {
		mu := &model.MaterialUnit{
			Name: `{"en":"Kg","zh":"千克"}`,
		}
		result := getMaterialUnitLocaleName(mu)
		if result.EN != "Kg" {
			t.Errorf("expected EN=%q, got %q", "Kg", result.EN)
		}
	})
}

func TestGetFromUnitUuid(t *testing.T) {
	material := &model.Material{
		NotBaseUnitList: []*model.MaterialUnit{
			{BaseModel: model.BaseModel{Uuid: 100}, UnitUuid: 200},
			{BaseModel: model.BaseModel{Uuid: 101}, UnitUuid: 201},
		},
	}

	tests := []struct {
		name     string
		unitUuid uint64
		want     uint64
	}{
		{"existing unit", 100, 200},
		{"another unit", 101, 201},
		{"non-existing unit", 999, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getFromUnitUuid(material, tt.unitUuid)
			if got != tt.want {
				t.Errorf("getFromUnitUuid(_, %d) = %d, want %d", tt.unitUuid, got, tt.want)
			}
		})
	}
}

func TestGetOriginCountry(t *testing.T) {
	t.Run("empty code returns nil", func(t *testing.T) {
		if got := getOriginCountry(""); got != nil {
			t.Error("expected nil for empty code")
		}
	})

	t.Run("invalid code returns nil", func(t *testing.T) {
		if got := getOriginCountry("INVALID_CODE_XYZ"); got != nil {
			t.Error("expected nil for invalid code")
		}
	})

	t.Run("valid code returns country", func(t *testing.T) {
		got := getOriginCountry("CN")
		if got == nil {
			t.Skip("CN country code not configured")
		}
		if got.Code != "CN" {
			t.Errorf("expected code CN, got %s", got.Code)
		}
	})
}

func TestBoolToStatus(t *testing.T) {
	if boolToStatus(true) != 0 {
		t.Error("disabled=true should return 0")
	}
	if boolToStatus(false) != 1 {
		t.Error("disabled=false should return 1")
	}
}

func TestDeleteTimeFromNotForSale(t *testing.T) {
	t.Run("not for sale returns non-zero", func(t *testing.T) {
		if deleteTimeFromNotForSale(true) == 0 {
			t.Error("expected non-zero delete time for not-for-sale item")
		}
	})

	t.Run("for sale returns zero", func(t *testing.T) {
		if deleteTimeFromNotForSale(false) != 0 {
			t.Error("expected zero delete time for for-sale item")
		}
	})
}

func TestHandleDanglingUnitRefs(t *testing.T) {
	t.Run("cost unit not in saved list, fallback to base unit", func(t *testing.T) {
		material := &model.Material{
			Unit: &model.MaterialUnit{BaseModel: model.BaseModel{Uuid: 10}},
		}
		material.CostUnitUuid = 99
		material.PurchaseUnitUuid = 10
		updateData := map[string]any{}
		handleDanglingUnitRefs(material, []uint64{10, 20}, updateData)
		if updateData["cost_unit_uuid"] != uint64(10) {
			t.Errorf("expected cost_unit_uuid=10, got %v", updateData["cost_unit_uuid"])
		}
	})

	t.Run("cost unit in saved list, no change", func(t *testing.T) {
		material := &model.Material{
			Unit: &model.MaterialUnit{BaseModel: model.BaseModel{Uuid: 10}},
		}
		material.CostUnitUuid = 20
		material.PurchaseUnitUuid = 10
		updateData := map[string]any{}
		handleDanglingUnitRefs(material, []uint64{10, 20}, updateData)
		if _, ok := updateData["cost_unit_uuid"]; ok {
			t.Error("cost_unit_uuid should not be set")
		}
	})

	t.Run("purchase unit dangling, clear to 0", func(t *testing.T) {
		material := &model.Material{
			Unit: &model.MaterialUnit{BaseModel: model.BaseModel{Uuid: 10}},
		}
		material.CostUnitUuid = 10
		material.PurchaseUnitUuid = 99
		updateData := map[string]any{}
		handleDanglingUnitRefs(material, []uint64{10, 20}, updateData)
		if updateData["purchase_unit_uuid"] != 0 {
			t.Errorf("expected purchase_unit_uuid=0, got %v", updateData["purchase_unit_uuid"])
		}
	})

	t.Run("purchase unit already set, no overwrite", func(t *testing.T) {
		material := &model.Material{
			Unit: &model.MaterialUnit{BaseModel: model.BaseModel{Uuid: 10}},
		}
		material.CostUnitUuid = 10
		material.PurchaseUnitUuid = 99
		updateData := map[string]any{"purchase_unit_uuid": uint64(20)}
		handleDanglingUnitRefs(material, []uint64{10, 20}, updateData)
		if updateData["purchase_unit_uuid"] != uint64(20) {
			t.Error("purchase_unit_uuid should not be overwritten")
		}
	})

	t.Run("nil base unit, cost fallback to 0", func(t *testing.T) {
		material := &model.Material{}
		material.CostUnitUuid = 99
		material.PurchaseUnitUuid = 10
		updateData := map[string]any{}
		handleDanglingUnitRefs(material, []uint64{10}, updateData)
		if updateData["cost_unit_uuid"] != 0 {
			t.Errorf("expected cost_unit_uuid=0, got %v", updateData["cost_unit_uuid"])
		}
	})
}

func TestSetDeleteTime(t *testing.T) {
	t.Run("not for sale with zero delete time, sets timestamp", func(t *testing.T) {
		material := &model.Material{}
		updateData := map[string]any{}
		setDeleteTime(material, true, updateData)
		if updateData["delete_time"].(int64) == 0 {
			t.Error("expected non-zero delete_time")
		}
	})

	t.Run("not for sale with existing delete time, no change", func(t *testing.T) {
		material := &model.Material{}
		material.DeleteTime = 12345
		updateData := map[string]any{}
		setDeleteTime(material, true, updateData)
		if _, ok := updateData["delete_time"]; ok {
			t.Error("should not set delete_time when already deleted")
		}
	})

	t.Run("for sale, resets to 0", func(t *testing.T) {
		material := &model.Material{}
		material.DeleteTime = 12345
		updateData := map[string]any{}
		setDeleteTime(material, false, updateData)
		if updateData["delete_time"] != 0 {
			t.Errorf("expected delete_time=0, got %v", updateData["delete_time"])
		}
	})
}

func TestBuildErpUpdateData(t *testing.T) {
	t.Run("with all fields", func(t *testing.T) {
		allowNeg := true
		data := buildErpUpdateData(req.MaterialEditErpReq{
			BarcodeValue:        "12345",
			InternalCode:        "IC001",
			DeliveredBySupplier: 1,
			SupplierErpCode:     "SUP001",
			Specification:       "500g",
			Disabled:            true,
			AllowNegativeStock:  &allowNeg,
		})
		if data["barcode_value"] != "12345" {
			t.Error("barcode_value mismatch")
		}
		if data["status"] != 0 {
			t.Error("disabled=true should set status=0")
		}
		if data["allow_negative_stock"] != 1 {
			t.Error("allow_negative_stock should be 1")
		}
	})

	t.Run("without allow_negative_stock", func(t *testing.T) {
		data := buildErpUpdateData(req.MaterialEditErpReq{Disabled: false})
		if data["status"] != 1 {
			t.Error("disabled=false should set status=1")
		}
		if _, ok := data["allow_negative_stock"]; ok {
			t.Error("allow_negative_stock should not be set when nil")
		}
	})

	t.Run("allow_negative_stock false", func(t *testing.T) {
		allowNeg := false
		data := buildErpUpdateData(req.MaterialEditErpReq{AllowNegativeStock: &allowNeg})
		if data["allow_negative_stock"] != 0 {
			t.Errorf("expected allow_negative_stock=0, got %v", data["allow_negative_stock"])
		}
	})
}

func TestFindMaterialUnitByProductUnit(t *testing.T) {
	t.Run("matches base unit", func(t *testing.T) {
		material := &model.Material{
			Unit: &model.MaterialUnit{
				BaseModel: model.BaseModel{Uuid: 10},
				Unit:      &model.ProductUnit{BaseModel: model.BaseModel{Uuid: 100}},
			},
		}
		material.UnitUuid = 10
		got := findMaterialUnitByProductUnit(material, nil, nil, 100)
		if got != 10 {
			t.Errorf("expected 10, got %d", got)
		}
	})

	t.Run("matches non-base unit via repo", func(t *testing.T) {
		material := &model.Material{
			Unit: &model.MaterialUnit{
				BaseModel: model.BaseModel{Uuid: 10},
				Unit:      &model.ProductUnit{BaseModel: model.BaseModel{Uuid: 100}},
			},
		}
		material.UnitUuid = 10
		material.Uuid = 1

		muRepo := &mockMaterialUnitRepo{
			units: map[uint64]model.MaterialUnit{
				200: {BaseModel: model.BaseModel{Uuid: 50}},
			},
		}

		got := findMaterialUnitByProductUnit(material, muRepo, &mockCommonRepo{}, 200)
		if got != 50 {
			t.Errorf("expected 50, got %d", got)
		}
	})

	t.Run("not found returns 0", func(t *testing.T) {
		material := &model.Material{
			Unit: &model.MaterialUnit{
				BaseModel: model.BaseModel{Uuid: 10},
				Unit:      &model.ProductUnit{BaseModel: model.BaseModel{Uuid: 100}},
			},
		}
		material.UnitUuid = 10

		muRepo := &mockMaterialUnitRepo{units: map[uint64]model.MaterialUnit{}}

		got := findMaterialUnitByProductUnit(material, muRepo, &mockCommonRepo{}, 999)
		if got != 0 {
			t.Errorf("expected 0, got %d", got)
		}
	})
}

func TestResolveUomProductUnitUuid(t *testing.T) {
	repo := &mockProductRepo{
		units: map[string]*model.ProductUnit{
			"Kg": {BaseModel: model.BaseModel{Uuid: 100}},
			"g":  {BaseModel: model.BaseModel{Uuid: 101}},
		},
	}
	uoms := []req.MaterialUomReq{{Uom: "Kg"}, {Uom: "g"}}

	t.Run("empty uom returns 0", func(t *testing.T) {
		uuid, err := resolveUomProductUnitUuid(repo, "", "IT001", "Kg", uoms, "采购单位")
		if err != nil || uuid != 0 {
			t.Errorf("expected (0, nil), got (%d, %v)", uuid, err)
		}
	})

	t.Run("uom not in conversion returns 0", func(t *testing.T) {
		uuid, err := resolveUomProductUnitUuid(repo, "lb", "IT001", "Kg", uoms, "采购单位")
		if err != nil || uuid != 0 {
			t.Errorf("expected (0, nil), got (%d, %v)", uuid, err)
		}
	})

	t.Run("uom in conversion and exists returns uuid", func(t *testing.T) {
		uuid, err := resolveUomProductUnitUuid(repo, "g", "IT001", "Kg", uoms, "采购单位")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if uuid != 101 {
			t.Errorf("expected 101, got %d", uuid)
		}
	})

	t.Run("uom equals stock uom returns uuid", func(t *testing.T) {
		uuid, err := resolveUomProductUnitUuid(repo, "Kg", "IT001", "Kg", uoms, "采购单位")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if uuid != 100 {
			t.Errorf("expected 100, got %d", uuid)
		}
	})

	t.Run("uom in conversion but not in product_unit returns error", func(t *testing.T) {
		uomsWithMissing := []req.MaterialUomReq{{Uom: "Kg"}, {Uom: "oz"}}
		_, err := resolveUomProductUnitUuid(repo, "oz", "IT001", "Kg", uomsWithMissing, "采购单位")
		if err == nil {
			t.Error("expected error for missing product unit")
		}
	})
}

func TestBuildErpUnitList(t *testing.T) {
	repo := &mockProductRepo{
		units: map[string]*model.ProductUnit{
			"Kg":  {BaseModel: model.BaseModel{Uuid: 100}},
			"g":   {BaseModel: model.BaseModel{Uuid: 101}},
			"pcs": {BaseModel: model.BaseModel{Uuid: 102}},
		},
	}

	t.Run("filters out stock uom", func(t *testing.T) {
		uoms := []req.MaterialUomReq{
			{Uom: "Kg", ConversionRate: 1},
			{Uom: "g", ConversionRate: 1000},
			{Uom: "pcs", ConversionRate: 10},
		}
		list, err := buildErpUnitList(repo, uoms, "Kg")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(list) != 2 {
			t.Fatalf("expected 2 units, got %d", len(list))
		}
		if list[0].Uuid != 101 || list[1].Uuid != 102 {
			t.Error("unexpected unit UUIDs")
		}
	})

	t.Run("missing unit returns error", func(t *testing.T) {
		uoms := []req.MaterialUomReq{{Uom: "lb", ConversionRate: 1}}
		_, err := buildErpUnitList(repo, uoms, "Kg")
		if err == nil {
			t.Error("expected error for missing unit")
		}
	})

	t.Run("empty uoms returns empty list", func(t *testing.T) {
		list, err := buildErpUnitList(repo, nil, "Kg")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(list) != 0 {
			t.Errorf("expected empty list, got %d", len(list))
		}
	})
}

func TestResolveUpdatePurchaseUnit(t *testing.T) {
	repo := &mockProductUnitRepo{
		units: map[string]*model.ProductUnit{
			"Kg": {BaseModel: model.BaseModel{Uuid: 100}},
		},
	}

	t.Run("empty purchase uom clears to 0", func(t *testing.T) {
		updateData := map[string]any{}
		uuid := resolveUpdatePurchaseUnit(repo, req.MaterialEditErpReq{PurchaseUom: ""}, updateData)
		if uuid != 0 || updateData["purchase_unit_uuid"] != 0 {
			t.Error("expected clear to 0")
		}
	})

	t.Run("not in conversion clears to 0", func(t *testing.T) {
		updateData := map[string]any{}
		uuid := resolveUpdatePurchaseUnit(repo, req.MaterialEditErpReq{
			PurchaseUom: "lb",
			StockUom:    "Kg",
			Uoms:        []req.MaterialUomReq{{Uom: "Kg"}},
			ItemCode:    "IT001",
		}, updateData)
		if uuid != 0 || updateData["purchase_unit_uuid"] != 0 {
			t.Error("expected clear to 0")
		}
	})

	t.Run("in conversion and found returns uuid", func(t *testing.T) {
		updateData := map[string]any{}
		uuid := resolveUpdatePurchaseUnit(repo, req.MaterialEditErpReq{
			PurchaseUom: "Kg",
			StockUom:    "g",
			Uoms:        []req.MaterialUomReq{{Uom: "Kg"}},
			ItemCode:    "IT001",
		}, updateData)
		if uuid != 100 {
			t.Errorf("expected 100, got %d", uuid)
		}
		if _, ok := updateData["purchase_unit_uuid"]; ok {
			t.Error("should not set purchase_unit_uuid when found")
		}
	})

	t.Run("in conversion but not found clears to 0", func(t *testing.T) {
		updateData := map[string]any{}
		uuid := resolveUpdatePurchaseUnit(repo, req.MaterialEditErpReq{
			PurchaseUom: "oz",
			StockUom:    "Kg",
			Uoms:        []req.MaterialUomReq{{Uom: "oz"}},
			ItemCode:    "IT001",
		}, updateData)
		if uuid != 0 || updateData["purchase_unit_uuid"] != 0 {
			t.Error("expected clear to 0")
		}
	})
}

func TestSyncNonBaseUnit(t *testing.T) {
	t.Run("update error returns error", func(t *testing.T) {
		muRepo := &mockMaterialUnitRepo{
			units:     map[uint64]model.MaterialUnit{200: {BaseModel: model.BaseModel{Uuid: 50}}},
			updateErr: fmt.Errorf("update failed"),
		}
		material := &model.Material{Unit: &model.MaterialUnit{BaseModel: model.BaseModel{Uuid: 10}}}
		material.Uuid = 1
		productUnit := &model.ProductUnit{BaseModel: model.BaseModel{Uuid: 200}, Name: "Gram"}
		_, err := syncNonBaseUnit(muRepo, &mockCommonRepo{}, material, productUnit, 1000)
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("create error returns error", func(t *testing.T) {
		muRepo := &mockMaterialUnitRepo{
			units:     map[uint64]model.MaterialUnit{},
			createErr: fmt.Errorf("create failed"),
		}
		material := &model.Material{Unit: &model.MaterialUnit{BaseModel: model.BaseModel{Uuid: 10}}}
		material.Uuid = 1
		productUnit := &model.ProductUnit{BaseModel: model.BaseModel{Uuid: 300}, Name: "Piece"}
		_, err := syncNonBaseUnit(muRepo, &mockCommonRepo{}, material, productUnit, 10)
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("existing unit gets updated", func(t *testing.T) {
		muRepo := &mockMaterialUnitRepo{
			units: map[uint64]model.MaterialUnit{
				200: {BaseModel: model.BaseModel{Uuid: 50}},
			},
		}
		material := &model.Material{
			Unit: &model.MaterialUnit{BaseModel: model.BaseModel{Uuid: 10}},
		}
		material.Uuid = 1
		productUnit := &model.ProductUnit{
			BaseModel: model.BaseModel{Uuid: 200},
			Name:      "Gram",
		}

		uuid, err := syncNonBaseUnit(muRepo, &mockCommonRepo{}, material, productUnit, 1000)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if uuid != 50 {
			t.Errorf("expected 50, got %d", uuid)
		}
		if len(muRepo.updatedData) != 1 {
			t.Error("expected one update call")
		}
	})

	t.Run("new unit gets created", func(t *testing.T) {
		muRepo := &mockMaterialUnitRepo{
			units: map[uint64]model.MaterialUnit{},
		}
		material := &model.Material{
			Unit: &model.MaterialUnit{BaseModel: model.BaseModel{Uuid: 10}},
		}
		material.Uuid = 1
		productUnit := &model.ProductUnit{
			BaseModel: model.BaseModel{Uuid: 300},
			Name:      "Piece",
		}

		uuid, err := syncNonBaseUnit(muRepo, &mockCommonRepo{}, material, productUnit, 10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if uuid != 999 { // mock always returns 999
			t.Errorf("expected 999, got %d", uuid)
		}
		if len(muRepo.createdUnits) != 1 {
			t.Error("expected one create call")
		}
	})
}

func TestSyncSingleMaterialUnit(t *testing.T) {
	t.Run("base unit update error returns error", func(t *testing.T) {
		muRepo := &mockMaterialUnitRepo{updateErr: fmt.Errorf("update failed")}
		material := &model.Material{Unit: &model.MaterialUnit{BaseModel: model.BaseModel{Uuid: 10}}}
		stockUnit := &model.ProductUnit{BaseModel: model.BaseModel{Uuid: 100}}
		productUnit := &model.ProductUnit{BaseModel: model.BaseModel{Uuid: 100}, Name: "Kg"}
		_, err := syncSingleMaterialUnit(muRepo, &mockCommonRepo{}, material, productUnit, stockUnit, 1)
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("base unit gets updated", func(t *testing.T) {
		muRepo := &mockMaterialUnitRepo{}
		material := &model.Material{
			Unit: &model.MaterialUnit{BaseModel: model.BaseModel{Uuid: 10}},
		}
		stockUnit := &model.ProductUnit{BaseModel: model.BaseModel{Uuid: 100}}
		productUnit := &model.ProductUnit{BaseModel: model.BaseModel{Uuid: 100}, Name: "Kg"}

		uuid, err := syncSingleMaterialUnit(muRepo, &mockCommonRepo{}, material, productUnit, stockUnit, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if uuid != 10 {
			t.Errorf("expected 10 (base unit uuid), got %d", uuid)
		}
	})

	t.Run("non-base unit delegates to syncNonBaseUnit", func(t *testing.T) {
		muRepo := &mockMaterialUnitRepo{
			units: map[uint64]model.MaterialUnit{},
		}
		material := &model.Material{
			Unit: &model.MaterialUnit{BaseModel: model.BaseModel{Uuid: 10}},
		}
		material.Uuid = 1
		stockUnit := &model.ProductUnit{BaseModel: model.BaseModel{Uuid: 100}}
		productUnit := &model.ProductUnit{BaseModel: model.BaseModel{Uuid: 200}, Name: "g"}

		uuid, err := syncSingleMaterialUnit(muRepo, &mockCommonRepo{}, material, productUnit, stockUnit, 1000)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if uuid != 999 { // created
			t.Errorf("expected 999, got %d", uuid)
		}
	})
}

func TestCleanupDeletedUnits(t *testing.T) {
	t.Run("destroy error returns error", func(t *testing.T) {
		muRepo := &mockMaterialUnitRepo{destroyErr: fmt.Errorf("destroy failed")}
		material := &model.Material{}
		material.Uuid = 1
		err := cleanupDeletedUnits(muRepo, &mockCommonRepo{}, material, []uint64{10})
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("empty save list does nothing", func(t *testing.T) {
		muRepo := &mockMaterialUnitRepo{}
		err := cleanupDeletedUnits(muRepo, &mockCommonRepo{}, &model.Material{}, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if muRepo.destroyCalled {
			t.Error("should not call destroy with empty list")
		}
	})

	t.Run("non-empty save list calls destroy", func(t *testing.T) {
		muRepo := &mockMaterialUnitRepo{}
		material := &model.Material{}
		material.Uuid = 1
		err := cleanupDeletedUnits(muRepo, &mockCommonRepo{}, material, []uint64{10, 20})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !muRepo.destroyCalled {
			t.Error("expected destroy to be called")
		}
	})
}

func TestResolveDefaultSalesUnit(t *testing.T) {
	t.Run("empty sales unit clears to 0", func(t *testing.T) {
		updateData := map[string]any{}
		resolveDefaultSalesUnit(nil, nil, nil, &model.Material{},
			req.MaterialEditErpReq{DefaultSalesUnit: ""}, updateData)
		if updateData["default_sales_unit_uuid"] != 0 {
			t.Error("expected 0")
		}
	})

	t.Run("not in conversion clears to 0", func(t *testing.T) {
		updateData := map[string]any{}
		resolveDefaultSalesUnit(nil, nil, nil, &model.Material{},
			req.MaterialEditErpReq{
				DefaultSalesUnit: "lb",
				StockUom:         "Kg",
				Uoms:             []req.MaterialUomReq{{Uom: "g"}},
			}, updateData)
		if updateData["default_sales_unit_uuid"] != 0 {
			t.Error("expected 0")
		}
	})

	t.Run("in conversion, matches base unit", func(t *testing.T) {
		puRepo := &mockProductUnitRepo{
			units: map[string]*model.ProductUnit{
				"Kg": {BaseModel: model.BaseModel{Uuid: 100}},
			},
		}
		material := &model.Material{
			Unit: &model.MaterialUnit{
				BaseModel: model.BaseModel{Uuid: 10},
				Unit:      &model.ProductUnit{BaseModel: model.BaseModel{Uuid: 100}},
			},
		}
		material.UnitUuid = 10

		updateData := map[string]any{}
		resolveDefaultSalesUnit(puRepo, nil, nil, material,
			req.MaterialEditErpReq{
				DefaultSalesUnit: "Kg",
				StockUom:         "Kg",
				Uoms:             []req.MaterialUomReq{{Uom: "Kg"}},
			}, updateData)
		if updateData["default_sales_unit_uuid"] != uint64(10) {
			t.Errorf("expected 10, got %v", updateData["default_sales_unit_uuid"])
		}
	})

	t.Run("in conversion, product unit not found clears to 0", func(t *testing.T) {
		puRepo := &mockProductUnitRepo{
			units: map[string]*model.ProductUnit{},
		}
		updateData := map[string]any{}
		resolveDefaultSalesUnit(puRepo, nil, nil, &model.Material{},
			req.MaterialEditErpReq{
				DefaultSalesUnit: "oz",
				StockUom:         "Kg",
				Uoms:             []req.MaterialUomReq{{Uom: "oz"}},
			}, updateData)
		if updateData["default_sales_unit_uuid"] != 0 {
			t.Error("expected 0 when product unit not found")
		}
	})
}

func TestResolveMaterialCategoryUuid(t *testing.T) {
	t.Run("empty code returns 1", func(t *testing.T) {
		uuid, err := resolveMaterialCategoryUuid(nil, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if uuid != 1 {
			t.Errorf("expected 1, got %d", uuid)
		}
	})

	t.Run("code found returns category uuid", func(t *testing.T) {
		repo := &mockMaterialRepo{
			categories: map[string]*model.MaterialCategory{
				"FOOD": {BaseModel: model.BaseModel{Uuid: 42}},
			},
		}
		uuid, err := resolveMaterialCategoryUuid(repo, "FOOD")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if uuid != 42 {
			t.Errorf("expected 42, got %d", uuid)
		}
	})

	t.Run("repo error returns error", func(t *testing.T) {
		repo := &mockMaterialRepo{categoryErr: fmt.Errorf("db error")}
		_, err := resolveMaterialCategoryUuid(repo, "FOOD")
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("code not found returns 1", func(t *testing.T) {
		repo := &mockMaterialRepo{
			categories: map[string]*model.MaterialCategory{},
		}
		uuid, err := resolveMaterialCategoryUuid(repo, "UNKNOWN")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if uuid != 1 {
			t.Errorf("expected 1, got %d", uuid)
		}
	})
}

func TestSyncMaterialUnits(t *testing.T) {
	t.Run("syncs all uoms", func(t *testing.T) {
		muRepo := &mockMaterialUnitRepo{
			units: map[uint64]model.MaterialUnit{},
		}
		puRepo := &mockProductUnitRepo{
			units: map[string]*model.ProductUnit{
				"Kg": {BaseModel: model.BaseModel{Uuid: 100}, Name: "Kg"},
				"g":  {BaseModel: model.BaseModel{Uuid: 101}, Name: "g"},
			},
		}
		material := &model.Material{
			Unit: &model.MaterialUnit{BaseModel: model.BaseModel{Uuid: 10}},
		}
		material.Uuid = 1
		stockUnit := &model.ProductUnit{BaseModel: model.BaseModel{Uuid: 100}}
		updateData := map[string]any{}

		uuids, err := syncMaterialUnits(syncMaterialUnitsParams{
			materialUnitRepo: muRepo,
			productUnitRepo:  puRepo,
			commonRepo:       &mockCommonRepo{},
			material:         material,
			stockUnit:        stockUnit,
			purchaseUnitUuid: 101, // g is purchase unit
			updateData:       updateData,
		}, req.MaterialEditErpReq{
			Uoms: []req.MaterialUomReq{
				{Uom: "Kg", ConversionRate: 1},
				{Uom: "g", ConversionRate: 1000},
			},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(uuids) != 2 {
			t.Fatalf("expected 2 uuids, got %d", len(uuids))
		}
		// g matched purchaseUnitUuid=101, so purchase_unit_uuid should be set
		if _, ok := updateData["purchase_unit_uuid"]; !ok {
			t.Error("expected purchase_unit_uuid to be set for matching purchase unit")
		}
	})

	t.Run("unknown uom returns error", func(t *testing.T) {
		puRepo := &mockProductUnitRepo{
			units: map[string]*model.ProductUnit{},
		}
		_, err := syncMaterialUnits(syncMaterialUnitsParams{
			productUnitRepo: puRepo,
			commonRepo:      &mockCommonRepo{},
			material:        &model.Material{},
			stockUnit:       &model.ProductUnit{},
			updateData:      map[string]any{},
		}, req.MaterialEditErpReq{
			Uoms:     []req.MaterialUomReq{{Uom: "INVALID"}},
			ItemCode: "IT001",
		})
		if err == nil {
			t.Error("expected error for unknown uom")
		}
	})

	t.Run("sync single unit error propagates", func(t *testing.T) {
		muRepo := &mockMaterialUnitRepo{
			units:     map[uint64]model.MaterialUnit{},
			updateErr: fmt.Errorf("update failed"),
		}
		puRepo := &mockProductUnitRepo{
			units: map[string]*model.ProductUnit{
				"Kg": {BaseModel: model.BaseModel{Uuid: 100}, Name: "Kg"},
			},
		}
		material := &model.Material{
			Unit: &model.MaterialUnit{BaseModel: model.BaseModel{Uuid: 10}},
		}
		material.Uuid = 1
		stockUnit := &model.ProductUnit{BaseModel: model.BaseModel{Uuid: 100}}
		_, err := syncMaterialUnits(syncMaterialUnitsParams{
			materialUnitRepo: muRepo,
			productUnitRepo:  puRepo,
			commonRepo:       &mockCommonRepo{},
			material:         material,
			stockUnit:        stockUnit,
			updateData:       map[string]any{},
		}, req.MaterialEditErpReq{
			Uoms: []req.MaterialUomReq{{Uom: "Kg", ConversionRate: 1}},
		})
		if err == nil {
			t.Error("expected error from syncSingleMaterialUnit")
		}
	})

	t.Run("empty uoms returns empty list", func(t *testing.T) {
		uuids, err := syncMaterialUnits(syncMaterialUnitsParams{
			updateData: map[string]any{},
		}, req.MaterialEditErpReq{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(uuids) != 0 {
			t.Errorf("expected 0, got %d", len(uuids))
		}
	})
}

func TestGetDefaultSalesUnitInfo(t *testing.T) {
	srv := &materialSrv{}

	t.Run("zero uuid returns empty", func(t *testing.T) {
		uuid, locale := srv.getDefaultSalesUnitInfo(nil, 0)
		if uuid != 0 {
			t.Errorf("expected 0, got %d", uuid)
		}
		if locale.EN != "" {
			t.Error("expected empty locale")
		}
	})

	t.Run("unit not found returns empty", func(t *testing.T) {
		muRepo := &mockMaterialUnitRepo{
			units: map[uint64]model.MaterialUnit{},
		}
		uuid, locale := srv.getDefaultSalesUnitInfo(muRepo, 999)
		if uuid != 0 {
			t.Errorf("expected 0, got %d", uuid)
		}
		if locale.EN != "" {
			t.Error("expected empty locale")
		}
	})

	t.Run("unit found but no product unit returns empty", func(t *testing.T) {
		muRepo := &mockMaterialUnitRepo{
			units: map[uint64]model.MaterialUnit{
				50: {BaseModel: model.BaseModel{Uuid: 50}, Unit: nil},
			},
		}
		uuid, locale := srv.getDefaultSalesUnitInfo(muRepo, 50)
		if uuid != 0 {
			t.Errorf("expected 0, got %d", uuid)
		}
		if locale.EN != "" {
			t.Error("expected empty locale")
		}
	})

	t.Run("unit found with product unit returns data", func(t *testing.T) {
		muRepo := &mockMaterialUnitRepo{
			units: map[uint64]model.MaterialUnit{
				50: {
					BaseModel: model.BaseModel{Uuid: 50},
					Unit: &model.ProductUnit{
						MultiLanguageName: model.MultiLanguageName{EnName: "Kilogram", ZhName: "千克"},
					},
				},
			},
		}
		uuid, locale := srv.getDefaultSalesUnitInfo(muRepo, 50)
		if uuid != 50 {
			t.Errorf("expected 50, got %d", uuid)
		}
		if locale.EN != "Kilogram" {
			t.Errorf("expected EN=Kilogram, got %q", locale.EN)
		}
	})
}

func TestBuildMaterialUnitList(t *testing.T) {
	srv := &materialSrv{}

	t.Run("repo error returns error", func(t *testing.T) {
		ctx := &mockContext{language: "en"}
		muRepo := &mockMaterialUnitRepo{unitListErr: fmt.Errorf("db error")}
		_, err := srv.buildMaterialUnitList(ctx, muRepo, 0)
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("empty list returns empty", func(t *testing.T) {
		ctx := &mockContext{language: "en"}
		muRepo := &mockMaterialUnitRepo{}
		list, err := srv.buildMaterialUnitList(ctx, muRepo, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(list) != 0 {
			t.Errorf("expected empty list, got %d", len(list))
		}
	})

	t.Run("returns unit list with locale names", func(t *testing.T) {
		ctx := &mockContext{language: "en"}
		muRepo := &mockMaterialUnitRepo{
			unitList: []*model.MaterialUnit{
				{
					BaseModel:      model.BaseModel{Uuid: 10},
					UnitUuid:       100,
					ConversionRate: 1,
					Unit: &model.ProductUnit{
						MultiLanguageName: model.MultiLanguageName{EnName: "Kilogram", ZhName: "千克"},
					},
				},
				{
					BaseModel:      model.BaseModel{Uuid: 20},
					UnitUuid:       101,
					ConversionRate: 1000,
					Unit: &model.ProductUnit{
						MultiLanguageName: model.MultiLanguageName{EnName: "Gram", ZhName: "克"},
					},
				},
			},
		}
		list, err := srv.buildMaterialUnitList(ctx, muRepo, 10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(list) != 2 {
			t.Fatalf("expected 2 units, got %d", len(list))
		}
		if list[0].Name != "Kilogram" {
			t.Errorf("expected name=Kilogram, got %q", list[0].Name)
		}
		if list[1].ConversionRate != 1000 {
			t.Errorf("expected conversion_rate=1000, got %v", list[1].ConversionRate)
		}
	})
}

func TestDoUpdateMaterialByErpItem(t *testing.T) {
	// helper to build a standard material with base unit
	newTestMaterial := func(uuid uint64) *model.Material {
		m := &model.Material{
			Unit: &model.MaterialUnit{
				BaseModel: model.BaseModel{Uuid: 10},
				Unit:      &model.ProductUnit{BaseModel: model.BaseModel{Uuid: 100}},
			},
		}
		m.Uuid = uuid
		m.UnitUuid = 10
		m.CostUnitUuid = 10
		m.PurchaseUnitUuid = 10
		return m
	}

	t.Run("happy path with single unit", func(t *testing.T) {
		matRepo := &mockMaterialRepo{
			materials: map[uint64]*model.Material{
				1: newTestMaterial(1),
			},
		}
		puRepo := &mockProductUnitRepo{
			units: map[string]*model.ProductUnit{
				"Kg": {BaseModel: model.BaseModel{Uuid: 100}, Name: "Kg"},
			},
		}
		muRepo := &mockMaterialUnitRepo{
			units: map[uint64]model.MaterialUnit{},
		}

		err := doUpdateMaterialByErpItem(matRepo, puRepo, muRepo, &mockCommonRepo{},
			req.MaterialEditErpReq{
				Uuid:     1,
				ItemCode: "IT001",
				StockUom: "Kg",
				Uoms:     []req.MaterialUomReq{{Uom: "Kg", ConversionRate: 1}},
			})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(matRepo.updatedData) != 1 {
			t.Fatal("expected UpdateMaterialData to be called once")
		}
		data := matRepo.updatedData[0]
		if data["category_uuid"] != uint64(1) {
			t.Errorf("expected category_uuid=1 (default), got %v", data["category_uuid"])
		}
	})

	t.Run("material not found returns error", func(t *testing.T) {
		matRepo := &mockMaterialRepo{
			materials: map[uint64]*model.Material{},
		}
		err := doUpdateMaterialByErpItem(matRepo, nil, nil, nil,
			req.MaterialEditErpReq{Uuid: 999})
		if err == nil {
			t.Error("expected error for missing material")
		}
	})

	t.Run("stock unit not found returns error", func(t *testing.T) {
		matRepo := &mockMaterialRepo{
			materials: map[uint64]*model.Material{
				1: newTestMaterial(1),
			},
		}
		puRepo := &mockProductUnitRepo{
			units: map[string]*model.ProductUnit{},
		}

		err := doUpdateMaterialByErpItem(matRepo, puRepo, nil, &mockCommonRepo{},
			req.MaterialEditErpReq{
				Uuid:     1,
				ItemCode: "IT001",
				StockUom: "INVALID",
			})
		if err == nil {
			t.Error("expected error for missing stock unit")
		}
	})

	t.Run("with classification code", func(t *testing.T) {
		matRepo := &mockMaterialRepo{
			materials: map[uint64]*model.Material{
				1: newTestMaterial(1),
			},
			categories: map[string]*model.MaterialCategory{
				"FOOD": {BaseModel: model.BaseModel{Uuid: 42}},
			},
		}
		puRepo := &mockProductUnitRepo{
			units: map[string]*model.ProductUnit{
				"Kg": {BaseModel: model.BaseModel{Uuid: 100}, Name: "Kg"},
			},
		}
		muRepo := &mockMaterialUnitRepo{
			units: map[uint64]model.MaterialUnit{},
		}

		err := doUpdateMaterialByErpItem(matRepo, puRepo, muRepo, &mockCommonRepo{},
			req.MaterialEditErpReq{
				Uuid:               1,
				ItemCode:           "IT001",
				StockUom:           "Kg",
				ClassificationCode: "FOOD",
				Uoms:               []req.MaterialUomReq{{Uom: "Kg", ConversionRate: 1}},
			})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		data := matRepo.updatedData[0]
		if data["category_uuid"] != uint64(42) {
			t.Errorf("expected category_uuid=42, got %v", data["category_uuid"])
		}
	})

	t.Run("with multiple units and purchase unit", func(t *testing.T) {
		mat := newTestMaterial(1)
		mat.PurchaseUnitUuid = 0
		matRepo := &mockMaterialRepo{
			materials: map[uint64]*model.Material{1: mat},
		}
		puRepo := &mockProductUnitRepo{
			units: map[string]*model.ProductUnit{
				"Kg": {BaseModel: model.BaseModel{Uuid: 100}, Name: "Kg"},
				"g":  {BaseModel: model.BaseModel{Uuid: 101}, Name: "g"},
			},
		}
		muRepo := &mockMaterialUnitRepo{
			units: map[uint64]model.MaterialUnit{},
		}

		err := doUpdateMaterialByErpItem(matRepo, puRepo, muRepo, &mockCommonRepo{},
			req.MaterialEditErpReq{
				Uuid:        1,
				ItemCode:    "IT001",
				StockUom:    "Kg",
				PurchaseUom: "g",
				Uoms: []req.MaterialUomReq{
					{Uom: "Kg", ConversionRate: 1},
					{Uom: "g", ConversionRate: 1000},
				},
			})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		data := matRepo.updatedData[0]
		// g was created and matched as purchase unit
		if _, ok := data["purchase_unit_uuid"]; !ok {
			t.Error("expected purchase_unit_uuid to be set")
		}
	})

	t.Run("category resolve error returns error", func(t *testing.T) {
		matRepo := &mockMaterialRepo{
			materials:   map[uint64]*model.Material{1: newTestMaterial(1)},
			categoryErr: fmt.Errorf("db error"),
		}
		puRepo := &mockProductUnitRepo{
			units: map[string]*model.ProductUnit{
				"Kg": {BaseModel: model.BaseModel{Uuid: 100}, Name: "Kg"},
			},
		}
		err := doUpdateMaterialByErpItem(matRepo, puRepo, nil, &mockCommonRepo{},
			req.MaterialEditErpReq{
				Uuid:               1,
				ItemCode:           "IT001",
				StockUom:           "Kg",
				ClassificationCode: "BAD",
			})
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("sync units error returns error", func(t *testing.T) {
		matRepo := &mockMaterialRepo{
			materials: map[uint64]*model.Material{1: newTestMaterial(1)},
		}
		puRepo := &mockProductUnitRepo{
			units: map[string]*model.ProductUnit{
				"Kg": {BaseModel: model.BaseModel{Uuid: 100}, Name: "Kg"},
			},
		}
		// Force syncMaterialUnits to fail by having an uom that doesn't exist in puRepo
		err := doUpdateMaterialByErpItem(matRepo, puRepo, &mockMaterialUnitRepo{}, &mockCommonRepo{},
			req.MaterialEditErpReq{
				Uuid:     1,
				ItemCode: "IT001",
				StockUom: "Kg",
				Uoms: []req.MaterialUomReq{
					{Uom: "Kg", ConversionRate: 1},
					{Uom: "INVALID", ConversionRate: 1},
				},
			})
		if err == nil {
			t.Error("expected error from syncMaterialUnits")
		}
	})

	t.Run("cleanup error returns error", func(t *testing.T) {
		matRepo := &mockMaterialRepo{
			materials: map[uint64]*model.Material{1: newTestMaterial(1)},
		}
		puRepo := &mockProductUnitRepo{
			units: map[string]*model.ProductUnit{
				"Kg": {BaseModel: model.BaseModel{Uuid: 100}, Name: "Kg"},
			},
		}
		muRepo := &mockMaterialUnitRepo{
			units:      map[uint64]model.MaterialUnit{},
			destroyErr: fmt.Errorf("destroy failed"),
		}
		err := doUpdateMaterialByErpItem(matRepo, puRepo, muRepo, &mockCommonRepo{},
			req.MaterialEditErpReq{
				Uuid:     1,
				ItemCode: "IT001",
				StockUom: "Kg",
				Uoms:     []req.MaterialUomReq{{Uom: "Kg", ConversionRate: 1}},
			})
		if err == nil {
			t.Error("expected error from cleanupDeletedUnits")
		}
	})

	t.Run("update material data error returns error", func(t *testing.T) {
		matRepo := &mockMaterialRepo{
			materials: map[uint64]*model.Material{1: newTestMaterial(1)},
			updateErr: fmt.Errorf("update failed"),
		}
		puRepo := &mockProductUnitRepo{
			units: map[string]*model.ProductUnit{
				"Kg": {BaseModel: model.BaseModel{Uuid: 100}, Name: "Kg"},
			},
		}
		muRepo := &mockMaterialUnitRepo{units: map[uint64]model.MaterialUnit{}}
		err := doUpdateMaterialByErpItem(matRepo, puRepo, muRepo, &mockCommonRepo{},
			req.MaterialEditErpReq{
				Uuid:     1,
				ItemCode: "IT001",
				StockUom: "Kg",
				Uoms:     []req.MaterialUomReq{{Uom: "Kg", ConversionRate: 1}},
			})
		if err == nil {
			t.Error("expected error from UpdateMaterialData")
		}
	})

	t.Run("not for sale sets delete_time", func(t *testing.T) {
		matRepo := &mockMaterialRepo{
			materials: map[uint64]*model.Material{
				1: newTestMaterial(1),
			},
		}
		puRepo := &mockProductUnitRepo{
			units: map[string]*model.ProductUnit{
				"Kg": {BaseModel: model.BaseModel{Uuid: 100}, Name: "Kg"},
			},
		}
		muRepo := &mockMaterialUnitRepo{
			units: map[uint64]model.MaterialUnit{},
		}

		err := doUpdateMaterialByErpItem(matRepo, puRepo, muRepo, &mockCommonRepo{},
			req.MaterialEditErpReq{
				Uuid:       1,
				ItemCode:   "IT001",
				StockUom:   "Kg",
				NotForSale: true,
				Uoms:       []req.MaterialUomReq{{Uom: "Kg", ConversionRate: 1}},
			})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		data := matRepo.updatedData[0]
		if dt, ok := data["delete_time"]; !ok || dt.(int64) == 0 {
			t.Error("expected non-zero delete_time for not-for-sale item")
		}
	})
}

func TestDoGetMaterialDetail(t *testing.T) {
	srv := &materialSrv{}

	t.Run("material not found returns error", func(t *testing.T) {
		matRepo := &mockMaterialRepo{materials: map[uint64]*model.Material{}}
		ctx := &mockContext{language: "en"}
		_, err := srv.doGetMaterialDetail(ctx, matRepo, nil, req.MaterialDetailReq{Uuid: 999})
		if err == nil {
			t.Error("expected error for missing material")
		}
	})

	t.Run("unit list error returns error", func(t *testing.T) {
		mat := &model.Material{
			Unit:     &model.MaterialUnit{BaseModel: model.BaseModel{Uuid: 10}},
			Category: model.MaterialCategory{},
		}
		mat.Uuid = 1
		matRepo := &mockMaterialRepo{materials: map[uint64]*model.Material{1: mat}}
		muRepo := &mockMaterialUnitRepo{unitListErr: fmt.Errorf("db error")}
		ctx := &mockContext{language: "en"}
		_, err := srv.doGetMaterialDetail(ctx, matRepo, muRepo, req.MaterialDetailReq{Uuid: 1})
		if err == nil {
			t.Error("expected error from buildMaterialUnitList")
		}
	})

	t.Run("happy path returns detail", func(t *testing.T) {
		baseUnit := &model.MaterialUnit{
			BaseModel: model.BaseModel{Uuid: 10},
			Unit: &model.ProductUnit{
				BaseModel:         model.BaseModel{Uuid: 100},
				MultiLanguageName: model.MultiLanguageName{EnName: "Kilogram", ZhName: "千克"},
			},
			Name: `{"en":"Kg","zh":"千克"}`,
		}
		mat := &model.Material{
			Unit: baseUnit,
			Category: model.MaterialCategory{
				MultiLanguageName: model.MultiLanguageName{EnName: "Food", ZhName: "食品"},
			},
			PurchaseUnit: &model.MaterialUnit{
				Unit: &model.ProductUnit{
					MultiLanguageName: model.MultiLanguageName{EnName: "Gram", ZhName: "克"},
				},
			},
			CostUnit: baseUnit,
		}
		mat.Uuid = 1
		mat.Code = "MAT001"
		mat.UnitUuid = 10
		mat.CategoryUuid = 42
		mat.BarcodeValue = "123456"
		mat.InternalCode = "IC001"
		mat.MultiLanguageName = model.MultiLanguageName{EnName: "Flour", ZhName: "面粉"}

		matRepo := &mockMaterialRepo{materials: map[uint64]*model.Material{1: mat}}
		muRepo := &mockMaterialUnitRepo{} // empty unit list
		ctx := &mockContext{language: "en"}

		resp, err := srv.doGetMaterialDetail(ctx, matRepo, muRepo, req.MaterialDetailReq{Uuid: 1})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Uuid != 1 {
			t.Errorf("expected uuid=1, got %d", resp.Uuid)
		}
		if resp.Code != "MAT001" {
			t.Errorf("expected code=MAT001, got %s", resp.Code)
		}
		if resp.CategoryUuid != 42 {
			t.Errorf("expected category_uuid=42, got %d", resp.CategoryUuid)
		}
		if resp.UnitName != "Kilogram" {
			t.Errorf("expected unit_name=Kilogram, got %s", resp.UnitName)
		}
		if resp.BarcodeValue != "123456" {
			t.Errorf("expected barcode=123456, got %s", resp.BarcodeValue)
		}
		if !resp.IsEditable {
			t.Error("expected IsEditable=true for non-headquarter item")
		}
	})

	t.Run("nil base unit", func(t *testing.T) {
		mat := &model.Material{
			Category: model.MaterialCategory{},
		}
		mat.Uuid = 2
		matRepo := &mockMaterialRepo{materials: map[uint64]*model.Material{2: mat}}
		muRepo := &mockMaterialUnitRepo{}
		ctx := &mockContext{language: "en"}

		resp, err := srv.doGetMaterialDetail(ctx, matRepo, muRepo, req.MaterialDetailReq{Uuid: 2})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.UnitName != "" {
			t.Errorf("expected empty unit name for nil unit, got %s", resp.UnitName)
		}
	})
}

// Verify unused import suppression
var _ dto.LocaleResponse
