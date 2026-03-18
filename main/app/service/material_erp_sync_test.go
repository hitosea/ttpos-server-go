package service

import (
	"testing"

	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/model"
)

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
			t.Errorf("expected En=%q, got %q", "Kg", result.EN)
		}
	})
}

func TestGetFromUnitUuid(t *testing.T) {
	material := &model.Material{
		NotBaseUnitList: []*model.MaterialUnit{
			{
				BaseModel: model.BaseModel{Uuid: 100},
				UnitUuid:  200,
			},
			{
				BaseModel: model.BaseModel{Uuid: 101},
				UnitUuid:  201,
			},
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
		result := deleteTimeFromNotForSale(true)
		if result == 0 {
			t.Error("expected non-zero delete time for not-for-sale item")
		}
	})

	t.Run("for sale returns zero", func(t *testing.T) {
		result := deleteTimeFromNotForSale(false)
		if result != 0 {
			t.Error("expected zero delete time for for-sale item")
		}
	})
}

func TestHandleDanglingUnitRefs(t *testing.T) {
	t.Run("cost unit not in saved list, fallback to base unit", func(t *testing.T) {
		material := &model.Material{
			Unit: &model.MaterialUnit{
				BaseModel: model.BaseModel{Uuid: 10},
			},
		}
		material.CostUnitUuid = 99
		material.PurchaseUnitUuid = 10
		updateData := map[string]any{}
		saveUnitUuids := []uint64{10, 20}

		handleDanglingUnitRefs(material, saveUnitUuids, updateData)

		if updateData["cost_unit_uuid"] != uint64(10) {
			t.Errorf("expected cost_unit_uuid=10, got %v", updateData["cost_unit_uuid"])
		}
	})

	t.Run("cost unit in saved list, no change", func(t *testing.T) {
		material := &model.Material{
			Unit: &model.MaterialUnit{
				BaseModel: model.BaseModel{Uuid: 10},
			},
		}
		material.CostUnitUuid = 20
		material.PurchaseUnitUuid = 10
		updateData := map[string]any{}
		saveUnitUuids := []uint64{10, 20}

		handleDanglingUnitRefs(material, saveUnitUuids, updateData)

		if _, ok := updateData["cost_unit_uuid"]; ok {
			t.Error("cost_unit_uuid should not be set when cost unit is in saved list")
		}
	})

	t.Run("purchase unit not in saved list and not set, clear to 0", func(t *testing.T) {
		material := &model.Material{
			Unit: &model.MaterialUnit{
				BaseModel: model.BaseModel{Uuid: 10},
			},
		}
		material.CostUnitUuid = 10
		material.PurchaseUnitUuid = 99
		updateData := map[string]any{}
		saveUnitUuids := []uint64{10, 20}

		handleDanglingUnitRefs(material, saveUnitUuids, updateData)

		if updateData["purchase_unit_uuid"] != 0 {
			t.Errorf("expected purchase_unit_uuid=0, got %v", updateData["purchase_unit_uuid"])
		}
	})

	t.Run("purchase unit already set in updateData, no overwrite", func(t *testing.T) {
		material := &model.Material{
			Unit: &model.MaterialUnit{
				BaseModel: model.BaseModel{Uuid: 10},
			},
		}
		material.CostUnitUuid = 10
		material.PurchaseUnitUuid = 99
		updateData := map[string]any{"purchase_unit_uuid": uint64(20)}
		saveUnitUuids := []uint64{10, 20}

		handleDanglingUnitRefs(material, saveUnitUuids, updateData)

		if updateData["purchase_unit_uuid"] != uint64(20) {
			t.Error("purchase_unit_uuid should not be overwritten when already set")
		}
	})

	t.Run("nil base unit, cost fallback to 0", func(t *testing.T) {
		material := &model.Material{}
		material.CostUnitUuid = 99
		material.PurchaseUnitUuid = 10
		updateData := map[string]any{}
		saveUnitUuids := []uint64{10}

		handleDanglingUnitRefs(material, saveUnitUuids, updateData)

		if updateData["cost_unit_uuid"] != 0 {
			t.Errorf("expected cost_unit_uuid=0 when base unit is nil, got %v", updateData["cost_unit_uuid"])
		}
	})
}

func TestSetDeleteTime(t *testing.T) {
	t.Run("not for sale with zero delete time, sets timestamp", func(t *testing.T) {
		material := &model.Material{}
		material.DeleteTime = 0
		updateData := map[string]any{}
		setDeleteTime(material, true, updateData)
		if _, ok := updateData["delete_time"]; !ok {
			t.Error("expected delete_time to be set")
		}
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
		v, ok := updateData["delete_time"]
		if !ok {
			t.Fatal("expected delete_time to be set")
		}
		if v != 0 {
			t.Errorf("expected delete_time=0, got %v", v)
		}
	})
}

func TestBuildErpUpdateData(t *testing.T) {
	allowNeg := true
	request := req.MaterialEditErpReq{
		BarcodeValue:        "12345",
		InternalCode:        "IC001",
		DeliveredBySupplier: 1,
		SupplierErpCode:     "SUP001",
		Specification:       "500g",
		Disabled:            true,
		AllowNegativeStock:  &allowNeg,
	}

	data := buildErpUpdateData(request)

	if data["barcode_value"] != "12345" {
		t.Error("barcode_value mismatch")
	}
	if data["status"] != 0 {
		t.Error("disabled=true should set status=0")
	}
	if data["allow_negative_stock"] != 1 {
		t.Error("allow_negative_stock should be 1")
	}
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

}
