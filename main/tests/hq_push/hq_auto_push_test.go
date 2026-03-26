//go:build integration

package hq_push_test

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"ttpos-server-go/tests/fixture"
)

// --- Constants ---

const (
	pathProductStatus = "/api/v1/shop/product/status"
)

// --- P0: Critical Path (auto-push triggered by HQ status change) ---

// Test_P0_Shop_HqAutoPush_NameSync verifies that product name (multi_language_name) is synced to sub-store
// when HQ modifies a product. The fix ensures pushSingleProductToStore calls syncMultiLanguageName.
// Route: POST /api/v1/shop/product/status (triggers OnHqProductChanged → pushSingleProductToStore)
func Test_P0_Shop_HqAutoPush_NameSync(t *testing.T) {
	env := setupHqEnv(t)

	productUuid := fixture.GenerateSnowflakeID()

	// Seed multi_language_name in HQ DB (new name: "川味宫保鸡丁")
	hqNameUuid := fixture.GenerateSnowflakeID()
	seedMultiLanguageName(t, env.hqDB, hqNameUuid, "川味宫保鸡丁", "Sichuan Kung Pao Chicken")

	// Seed HQ product with multi_language_name_uuid
	fixture.SeedProductPackage(t, env.hqDB, fixture.WithProductPackageUUID(productUuid), fixture.WithProductPackageStatus(0))
	updateProductField(t, env.hqDB, productUuid, "multi_language_name_uuid", hqNameUuid)
	updateProductField(t, env.hqDB, productUuid, "name", `{"zh_name":"川味宫保鸡丁","en_name":"Sichuan Kung Pao Chicken"}`)

	// Seed multi_language_name in sub-store DB (old name: "宫保鸡丁")
	storeNameUuid := fixture.GenerateSnowflakeID()
	seedMultiLanguageName(t, env.storeDB1, storeNameUuid, "宫保鸡丁", "Kung Pao Chicken")

	// Seed sub-store product with old multi_language_name_uuid
	fixture.SeedProductPackage(t, env.storeDB1,
		fixture.WithProductPackageUUID(productUuid),
		fixture.WithProductPackageStatus(0),
		fixture.WithProductPackageHeadquarterUuid(env.hqUuid),
	)
	updateProductField(t, env.storeDB1, productUuid, "multi_language_name_uuid", storeNameUuid)

	// HQ changes product status → triggers auto-push of ALL fields
	resp := env.client.Post(t, pathProductStatus, map[string]any{
		"uuid":   productUuid,
		"status": 1,
	})
	resp.AssertOK(t).AssertSuccess(t)

	// Verify: sub-store's multi_language_name record should be updated
	waitForAsync(t, 5*time.Second, func() bool {
		zhName := queryStringField(t, env.storeDB1, "ttpos_multi_language_name", "zh_name", storeNameUuid)
		return zhName == "川味宫保鸡丁"
	})

	// Also verify en_name
	enName := queryStringField(t, env.storeDB1, "ttpos_multi_language_name", "en_name", storeNameUuid)
	if enName != "Sichuan Kung Pao Chicken" {
		t.Fatalf("expected en_name 'Sichuan Kung Pao Chicken', got '%s'", enName)
	}
}

// Test_P0_Shop_HqAutoPush_ImageFileSync verifies that product image file record is copied to sub-store
// when HQ changes the product image. The fix adds syncFileToStore in pushSingleProductToStore.
// Route: POST /api/v1/shop/product/status (triggers OnHqProductChanged → pushSingleProductToStore)
func Test_P0_Shop_HqAutoPush_ImageFileSync(t *testing.T) {
	env := setupHqEnv(t)

	productUuid := fixture.GenerateSnowflakeID()
	newImageFileUuid := fixture.GenerateSnowflakeID()

	// Seed new image file in HQ DB only
	seedFile(t, env.hqDB, newImageFileUuid, "local", "product/new_image.jpg")

	// Seed HQ product with new image
	fixture.SeedProductPackage(t, env.hqDB, fixture.WithProductPackageUUID(productUuid), fixture.WithProductPackageStatus(0))
	updateProductField(t, env.hqDB, productUuid, "image_file_uuid", newImageFileUuid)

	// Seed sub-store product with old/no image (image_file_uuid=0)
	fixture.SeedProductPackage(t, env.storeDB1,
		fixture.WithProductPackageUUID(productUuid),
		fixture.WithProductPackageStatus(0),
		fixture.WithProductPackageHeadquarterUuid(env.hqUuid),
	)

	// HQ changes status → triggers auto-push
	resp := env.client.Post(t, pathProductStatus, map[string]any{
		"uuid":   productUuid,
		"status": 1,
	})
	resp.AssertOK(t).AssertSuccess(t)

	// Verify: new file record should exist in sub-store DB
	waitForAsync(t, 5*time.Second, func() bool {
		return queryFileExists(t, env.storeDB1, newImageFileUuid)
	})

	// Verify: sub-store product's image_file_uuid should be updated
	imgUuid := queryIntField(t, env.storeDB1, "ttpos_product_package", "image_file_uuid", productUuid)
	if int64(imgUuid) != newImageFileUuid {
		t.Fatalf("expected image_file_uuid %d, got %d", newImageFileUuid, imgUuid)
	}
}

// Test_P0_Shop_HqAutoPush_SellingPointSync verifies that selling point (describe_multi_language_name)
// is synced to sub-store. The fix ensures syncMultiLanguageName is called for describe field.
// Route: POST /api/v1/shop/product/status (triggers OnHqProductChanged → pushSingleProductToStore)
func Test_P0_Shop_HqAutoPush_SellingPointSync(t *testing.T) {
	env := setupHqEnv(t)

	productUuid := fixture.GenerateSnowflakeID()

	// Seed describe multi_language_name in HQ DB
	hqDescUuid := fixture.GenerateSnowflakeID()
	seedMultiLanguageName(t, env.hqDB, hqDescUuid, "香辣爽口，经典川味", "Spicy & savory Sichuan classic")

	// Seed HQ product with describe
	fixture.SeedProductPackage(t, env.hqDB, fixture.WithProductPackageUUID(productUuid), fixture.WithProductPackageStatus(0))
	updateProductField(t, env.hqDB, productUuid, "describe_multi_language_name_uuid", hqDescUuid)
	updateProductField(t, env.hqDB, productUuid, "describe", `{"zh_name":"香辣爽口，经典川味","en_name":"Spicy & savory Sichuan classic"}`)

	// Seed sub-store product with empty describe
	storeDescUuid := fixture.GenerateSnowflakeID()
	seedMultiLanguageName(t, env.storeDB1, storeDescUuid, "", "")

	fixture.SeedProductPackage(t, env.storeDB1,
		fixture.WithProductPackageUUID(productUuid),
		fixture.WithProductPackageStatus(0),
		fixture.WithProductPackageHeadquarterUuid(env.hqUuid),
	)
	updateProductField(t, env.storeDB1, productUuid, "describe_multi_language_name_uuid", storeDescUuid)

	// HQ changes status → triggers auto-push
	resp := env.client.Post(t, pathProductStatus, map[string]any{
		"uuid":   productUuid,
		"status": 1,
	})
	resp.AssertOK(t).AssertSuccess(t)

	// Verify: sub-store's describe multi_language_name record should be updated
	waitForAsync(t, 5*time.Second, func() bool {
		zhDesc := queryStringField(t, env.storeDB1, "ttpos_multi_language_name", "zh_name", storeDescUuid)
		return zhDesc == "香辣爽口，经典川味"
	})
}

// Test_P0_Shop_HqAutoPush_OverrideRespected verifies that sub-store's shelf status override is respected
// during auto-push. The fix ensures each goroutine gets its own updateData map copy.
// Route: POST /api/v1/shop/product/status (triggers OnHqProductChanged → pushSingleProductToStore)
func Test_P0_Shop_HqAutoPush_OverrideRespected(t *testing.T) {
	env := setupHqEnv(t)

	productUuid := fixture.GenerateSnowflakeID()

	// HQ product: status=1 (上架)
	fixture.SeedProductPackage(t, env.hqDB, fixture.WithProductPackageUUID(productUuid), fixture.WithProductPackageStatus(1))

	// Sub-store product: status=0 (下架) — sub-store explicitly set this
	fixture.SeedProductPackage(t, env.storeDB1,
		fixture.WithProductPackageUUID(productUuid),
		fixture.WithProductPackageStatus(0),
		fixture.WithProductPackageHeadquarterUuid(env.hqUuid),
	)

	// Mark override: sub-store has overridden dine_shelf
	seedHqFieldOverride(t, env.storeDB1, productUuid, "product", "dine_shelf")

	// HQ changes status (0→1 or stays 1) → triggers auto-push
	resp := env.client.Post(t, pathProductStatus, map[string]any{
		"uuid":   productUuid,
		"status": 1,
	})
	resp.AssertOK(t).AssertSuccess(t)

	// Wait for async push to complete (other fields may be synced)
	time.Sleep(2 * time.Second)

	// Verify: sub-store status should remain 0 (下架), NOT synced to HQ's 1
	storeStatus := queryIntField(t, env.storeDB1, "ttpos_product_package", "status", productUuid)
	if storeStatus != 0 {
		t.Fatalf("expected sub-store status=0 (override should be respected), got %d", storeStatus)
	}

	// Verify: override record should still exist (not cleared)
	if !queryOverrideExists(t, env.storeDB1, productUuid, "dine_shelf") {
		t.Fatal("expected override record to still exist")
	}
}

// --- P2: Data Validation ---

// Test_P2_Shop_HqAutoPush_ConcurrentMapSafe verifies that auto-push to multiple stores
// does not panic due to concurrent map writes. The fix gives each store its own updateData map.
// Route: POST /api/v1/shop/product/status (triggers OnHqProductChanged → pushProductToStores)
func Test_P2_Shop_HqAutoPush_ConcurrentMapSafe(t *testing.T) {
	env := setupHqEnv(t)

	productUuid := fixture.GenerateSnowflakeID()

	// HQ product: status=1
	fixture.SeedProductPackage(t, env.hqDB, fixture.WithProductPackageUUID(productUuid), fixture.WithProductPackageStatus(0))

	// Both stores have the product
	fixture.SeedProductPackage(t, env.storeDB1,
		fixture.WithProductPackageUUID(productUuid),
		fixture.WithProductPackageStatus(0),
		fixture.WithProductPackageHeadquarterUuid(env.hqUuid),
	)
	fixture.SeedProductPackage(t, env.storeDB2,
		fixture.WithProductPackageUUID(productUuid),
		fixture.WithProductPackageStatus(0),
		fixture.WithProductPackageHeadquarterUuid(env.hqUuid),
	)

	// Store1 has override, Store2 does not — this creates the race condition scenario
	seedHqFieldOverride(t, env.storeDB1, productUuid, "product", "dine_shelf")

	// HQ changes status → triggers concurrent push to both stores
	resp := env.client.Post(t, pathProductStatus, map[string]any{
		"uuid":   productUuid,
		"status": 1,
	})
	resp.AssertOK(t).AssertSuccess(t)

	// Verify: store2 (no override) should be synced
	waitForAsync(t, 5*time.Second, func() bool {
		return queryIntField(t, env.storeDB2, "ttpos_product_package", "status", productUuid) == 1
	})

	// Verify: store1 (with override) should keep its own status
	store1Status := queryIntField(t, env.storeDB1, "ttpos_product_package", "status", productUuid)
	if store1Status != 0 {
		t.Fatalf("store1 has override, expected status=0, got %d", store1Status)
	}
}

// --- Helpers (auto-push specific) ---

// seedMultiLanguageName inserts a multi_language_name record.
func seedMultiLanguageName(t *testing.T, db *sql.DB, uuid int64, zhName, enName string) {
	t.Helper()
	now := time.Now().Unix()
	_, err := db.Exec(`
		INSERT INTO ttpos_multi_language_name (uuid, zh_name, en_name, th_name, zh_tw_name, ja_name, ko_name, my_name, tr_name, sv_name, create_time, update_time, delete_time)
		VALUES (?, ?, ?, '', '', '', '', '', '', '', ?, ?, 0)
	`, uuid, zhName, enName, now, now)
	if err != nil {
		t.Fatalf("failed to seed multi_language_name: %v", err)
	}
}

// seedFile inserts a file record in ttpos_file.
func seedFile(t *testing.T, db *sql.DB, uuid int64, storage, saveName string) {
	t.Helper()
	now := time.Now().Unix()
	_, err := db.Exec(`
		INSERT INTO ttpos_file (uuid, storage, file_url, save_name, file_name, file_size, file_type, extension, headquarter_uuid, create_time, update_time, delete_time)
		VALUES (?, ?, '', ?, ?, 1024, 'image', '.jpg', 0, ?, ?, 0)
	`, uuid, storage, saveName, saveName, now, now)
	if err != nil {
		t.Fatalf("failed to seed file: %v", err)
	}
}

// updateProductField updates a single field on ttpos_product_package by uuid.
func updateProductField(t *testing.T, db *sql.DB, uuid int64, field string, value any) {
	t.Helper()
	_, err := db.Exec(fmt.Sprintf("UPDATE ttpos_product_package SET `%s` = ? WHERE uuid = ?", field), value, uuid)
	if err != nil {
		t.Fatalf("failed to update product field %s: %v", field, err)
	}
}

// queryStringField queries a single string field from a table by uuid.
func queryStringField(t *testing.T, db *sql.DB, table, field string, uuid int64) string {
	t.Helper()
	var val string
	err := db.QueryRow(fmt.Sprintf("SELECT `%s` FROM `%s` WHERE uuid = ? AND delete_time = 0", field, table), uuid).Scan(&val)
	if err != nil {
		return ""
	}
	return val
}

// queryFileExists checks if a file record exists in ttpos_file.
func queryFileExists(t *testing.T, db *sql.DB, uuid int64) bool {
	t.Helper()
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM ttpos_file WHERE uuid = ? AND delete_time = 0", uuid).Scan(&count)
	if err != nil {
		return false
	}
	return count > 0
}
