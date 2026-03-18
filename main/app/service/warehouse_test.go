package service

import (
	"testing"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/model"

	"github.com/stretchr/testify/assert"
)

// ─── buildWarehouseResp tests ───

func TestBuildWarehouseResp_BasicFields(t *testing.T) {
	t.Parallel()
	srv := &warehouseSrv{}
	warehouse := model.Warehouse{
		Type:    "normal",
		Code:    "WH-001",
		Status:  1,
		Contact: "John",
		Phone:   "12345",
		Address: "Test Address",
		MultiLanguageName: &model.MultiLanguageName{
			ZhName: "测试仓库",
			EnName: "Test Warehouse",
		},
	}
	warehouse.Uuid = 100

	result := srv.buildWarehouseResp(nil, warehouse, false)
	assert.Equal(t, uint64(100), result.Uuid)
	assert.Equal(t, "normal", result.Type)
	assert.Equal(t, "WH-001", result.Code)
	assert.Equal(t, 1, result.Status)
	assert.Equal(t, "John", result.Contact)
	assert.Equal(t, "12345", result.Phone)
	assert.Equal(t, "Test Address", result.Address)
	assert.True(t, result.IsEditable, "HeadquarterUuid=0 should be editable")
	assert.False(t, result.HasItem)
}

func TestBuildWarehouseResp_HeadquarterUsesErpCode(t *testing.T) {
	t.Parallel()
	srv := &warehouseSrv{}
	warehouse := model.Warehouse{
		Code:              "WH-LOCAL",
		ErpCode:           "ERP-HQ-001",
		HeadquarterUuid:   999,
		MultiLanguageName: &model.MultiLanguageName{},
	}
	warehouse.Uuid = 200

	result := srv.buildWarehouseResp(nil, warehouse, true)
	assert.Equal(t, "ERP-HQ-001", result.Code, "Headquarter should use ErpCode")
	assert.False(t, result.IsEditable, "HeadquarterUuid>0 should NOT be editable")
}

func TestBuildWarehouseResp_NilMultiLanguageName(t *testing.T) {
	t.Parallel()
	srv := &warehouseSrv{}
	warehouse := model.Warehouse{MultiLanguageName: nil}

	result := srv.buildWarehouseResp(nil, warehouse, false)
	assert.Equal(t, dto.LocaleResponse{}, result.LocalName, "Nil MultiLanguageName should give empty locale")
}

func TestBuildWarehouseResp_HasItems(t *testing.T) {
	t.Parallel()
	srv := &warehouseSrv{}
	warehouse := model.Warehouse{
		MultiLanguageName: &model.MultiLanguageName{},
		Items:             []*model.WarehouseItem{{}, {}},
	}

	result := srv.buildWarehouseResp(nil, warehouse, false)
	assert.True(t, result.HasItem, "Should detect items")
}

// ─── buildWarehouseInOutResp tests ───

func TestBuildWarehouseInOutResp_WithMaterial(t *testing.T) {
	t.Parallel()
	srv := &warehouseSrv{}
	log := model.WarehouseInOutLog{
		LogType:      constant.WarehouseInOutLogLogTypeIn,
		Scene:        constant.WarehouseInOutLogScenePurchase,
		MaterialUuid: 5001,
		Num:          10.0,
		Amount:       100.0,
		OrderNo:      "PO-001",
		Material: &model.Material{
			Code:         "MAT-001",
			BarcodeValue: "BC-001",
			CategoryUuid: 300,
			MultiLanguageName: model.MultiLanguageName{
				ZhName: "物品A",
				EnName: "ItemA",
			},
		},
	}
	log.Uuid = 1
	log.CreateTime = 1700000000

	result := srv.buildWarehouseInOutResp(log)
	assert.Equal(t, uint64(1), result.Uuid)
	assert.Equal(t, "PO-001", result.OrderNo)
	assert.Equal(t, "purchase", result.Type)
	assert.Equal(t, 10.0, result.Num)
	assert.Equal(t, 100.0, result.Amount)
	assert.Equal(t, "MAT-001", result.MaterialCode)
	assert.Equal(t, "BC-001", result.MaterialBarcode)
	assert.Equal(t, uint64(300), result.MaterialCategoryUuid)
}

func TestBuildWarehouseInOutResp_NilMaterial(t *testing.T) {
	t.Parallel()
	srv := &warehouseSrv{}
	log := model.WarehouseInOutLog{
		Scene: constant.WarehouseInOutLogSceneProfitIn,
	}

	result := srv.buildWarehouseInOutResp(log)
	assert.Equal(t, "profit_in", result.Type)
	assert.Equal(t, "", result.MaterialCode)
}

func TestBuildWarehouseInOutResp_WithSupplier(t *testing.T) {
	t.Parallel()
	srv := &warehouseSrv{}
	log := model.WarehouseInOutLog{
		Scene:        constant.WarehouseInOutLogScenePurchase,
		SupplierUuid: 7001,
		Supplier: &model.Supplier{
			Name: "供应商A",
		},
	}

	result := srv.buildWarehouseInOutResp(log)
	assert.Equal(t, uint64(7001), result.SupplierUuid)
	assert.Equal(t, "供应商A", result.SupplierName.ZH)
}

func TestBuildWarehouseInOutResp_SupplierFallback(t *testing.T) {
	t.Parallel()
	srv := &warehouseSrv{}
	log := model.WarehouseInOutLog{
		Scene:        constant.WarehouseInOutLogScenePurchase,
		SupplierName: "FallbackName",
	}

	result := srv.buildWarehouseInOutResp(log)
	assert.Equal(t, "FallbackName", result.SupplierName.ZH, "Should use SupplierName field as fallback")
	assert.Equal(t, "FallbackName", result.SupplierName.EN)
}

func TestBuildWarehouseInOutResp_WithWarehouse(t *testing.T) {
	t.Parallel()
	srv := &warehouseSrv{}
	log := model.WarehouseInOutLog{
		Scene:         constant.WarehouseInOutLogSceneSale,
		WarehouseUuid: 100,
		Warehouse: &model.Warehouse{
			MultiLanguageName: &model.MultiLanguageName{
				ZhName: "主仓库",
				EnName: "Main Warehouse",
			},
		},
	}

	result := srv.buildWarehouseInOutResp(log)
	assert.Equal(t, uint64(100), result.WarehouseUuid)
	assert.NotEmpty(t, result.WarehouseName.ZH)
}

func TestBuildWarehouseInOutResp_DateFormatting(t *testing.T) {
	t.Parallel()
	srv := &warehouseSrv{}
	log := model.WarehouseInOutLog{
		Scene: constant.WarehouseInOutLogSceneLossOut,
	}
	log.CreateTime = 0

	result := srv.buildWarehouseInOutResp(log)
	assert.Equal(t, "", result.Date, "Zero CreateTime should give empty date")
	assert.Equal(t, "loss_out", result.Type)

	log2 := model.WarehouseInOutLog{
		Scene: constant.WarehouseInOutLogSceneTransferIn,
	}
	log2.CreateTime = 1700000000

	result2 := srv.buildWarehouseInOutResp(log2)
	assert.NotEmpty(t, result2.Date, "Non-zero CreateTime should give a date")
	assert.Equal(t, "transfer_in", result2.Type)
}
