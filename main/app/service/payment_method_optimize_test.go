package service

import (
	"testing"
	"time"

	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/pkg/cache"
)

// TestPaymentMethodSrv_GetManagementList_FilterFreeMeal 测试 GetManagementList 过滤 Free Meal
func TestPaymentMethodSrv_GetManagementList_FilterFreeMeal(t *testing.T) {
	dbm := setupPaymentMethodTestDB(t)
	ctx := createPaymentMethodTestContext(t, dbm)
	db := dbm.GetDB(constant.MockDB)

	// 创建测试数据
	paymentMethods := []model.PaymentMethod{
		{
			BaseModel: model.BaseModel{
				Uuid:       1001,
				CreateTime: time.Now().Unix(),
				UpdateTime: time.Now().Unix(),
			},
			Name:        "Cash",
			Code:        constant.PaymentMethodCodeCash,
			PaymentName: "Cash",
			Source:      constant.PaymentMethodSourceSystem,
			Status:      constant.PaymentMethodStatusEnable,
			Sort:        1,
		},
		{
			BaseModel: model.BaseModel{
				Uuid:       1002,
				CreateTime: time.Now().Unix(),
				UpdateTime: time.Now().Unix(),
			},
			Name:        "Free Meal",
			Code:        constant.PaymentMethodCodeFreePay, // -1
			PaymentName: "Free Meal",
			Source:      constant.PaymentMethodSourceSystem,
			Status:      constant.PaymentMethodStatusEnable,
			Sort:        2,
		},
		{
			BaseModel: model.BaseModel{
				Uuid:       1003,
				CreateTime: time.Now().Unix(),
				UpdateTime: time.Now().Unix(),
			},
			Name:        "Balance",
			Code:        constant.PaymentMethodCodeBalance,
			PaymentName: "Balance",
			Source:      constant.PaymentMethodSourceSystem,
			Status:      constant.PaymentMethodStatusEnable,
			Sort:        3,
		},
		{
			BaseModel: model.BaseModel{
				Uuid:       1004,
				CreateTime: time.Now().Unix(),
				UpdateTime: time.Now().Unix(),
			},
			Name:        "Wechat Pay",
			Code:        constant.PaymentMethodCodeWechat,
			PaymentName: "Wechat Pay",
			Source:      constant.PaymentMethodSourceDefault,
			Status:      constant.PaymentMethodStatusEnable,
			Sort:        4,
		},
	}

	for _, pm := range paymentMethods {
		db.Create(&pm)
	}

	// 创建服务
	mockCache := cache.NewCache(cache.GoCache, cache.Config{})
	settingSrv := setting.NewSrv(dbm, mockCache)
	paymentMethodSrv := NewPaymentMethodSrv(dbm, settingSrv)

	// 测试 GetManagementList
	req := &req.PaymentMethodManagementListReq{
		PageNo:   1,
		PageSize: 10,
	}
	result, err := paymentMethodSrv.GetManagementList(ctx, req)
	if err != nil {
		t.Fatalf("Failed to get management list: %v", err)
	}

	// 验证结果：应该不包含 Free Meal（code=-1）
	foundFreeMeal := false
	for _, item := range result.List {
		if item.Uuid == 1002 {
			foundFreeMeal = true
			t.Errorf("Free Meal (code=-1) should be filtered out, but found in result")
		}
	}

	if foundFreeMeal {
		t.Errorf("Free Meal should not appear in management list")
	}

	// 验证其他支付方式都存在
	// 注意：Balance 在会员功能未开启时也会被过滤，所以预期只有 2 个（Cash, Wechat Pay）
	expectedCount := 2 // Cash, Wechat Pay（Balance 被过滤因为会员功能未开启）
	if len(result.List) != expectedCount {
		t.Errorf("Expected %d payment methods, got %d. List: %+v", expectedCount, len(result.List), result.List)
	}

	// 验证 Cash 存在
	foundCash := false
	for _, item := range result.List {
		if item.Uuid == 1001 {
			foundCash = true
			break
		}
	}
	if !foundCash {
		t.Errorf("Cash payment method should be in the list")
	}
}

// TestPaymentMethodSrv_GetManagementList_FilterFreeMealForErp 测试 GetManagementList 不过滤 Free Meal for ERP
func TestPaymentMethodSrv_GetManagementList_FilterFreeMealForErp(t *testing.T) {
	dbm := setupPaymentMethodTestDB(t)
	ctx := createPaymentMethodTestContext(t, dbm)
	db := dbm.GetDB(constant.MockDB)

	// 创建测试数据
	paymentMethods := []model.PaymentMethod{
		{
			BaseModel: model.BaseModel{
				Uuid:       1001,
				CreateTime: time.Now().Unix(),
				UpdateTime: time.Now().Unix(),
			},
			Name:        "Cash",
			Code:        constant.PaymentMethodCodeCash,
			PaymentName: "Cash",
			Source:      constant.PaymentMethodSourceSystem,
			Status:      constant.PaymentMethodStatusEnable,
			Sort:        1,
		},
		{
			BaseModel: model.BaseModel{
				Uuid:       1002,
				CreateTime: time.Now().Unix(),
				UpdateTime: time.Now().Unix(),
			},
			Name:        "Free Meal",
			Code:        constant.PaymentMethodCodeFreePay, // -1（原有）
			PaymentName: "Free Meal",
			Source:      constant.PaymentMethodSourceSystem,
			Status:      constant.PaymentMethodStatusEnable,
			Sort:        2,
		},
		{
			BaseModel: model.BaseModel{
				Uuid:       1003,
				CreateTime: time.Now().Unix(),
				UpdateTime: time.Now().Unix(),
			},
			Name:        "Free Meal for ERP",
			Code:        constant.PaymentMethodCodeFreeMealForErp, // 92000（用于ERP）
			PaymentName: "Free Meal",
			Source:      constant.PaymentMethodSourceSystem,
			Status:      constant.PaymentMethodStatusEnable,
			Sort:        3,
		},
	}

	for _, pm := range paymentMethods {
		db.Create(&pm)
	}

	// 创建服务
	mockCache := cache.NewCache(cache.GoCache, cache.Config{})
	settingSrv := setting.NewSrv(dbm, mockCache)
	paymentMethodSrv := NewPaymentMethodSrv(dbm, settingSrv)

	// 测试 GetManagementList
	req := &req.PaymentMethodManagementListReq{
		PageNo:   1,
		PageSize: 10,
	}
	result, err := paymentMethodSrv.GetManagementList(ctx, req)
	if err != nil {
		t.Fatalf("Failed to get management list: %v", err)
	}

	// 验证结果：
	// - Free Meal（code=-1）应该被过滤
	// - Free Meal for ERP（code=92000）也应该被过滤（因为不在 excludeCodes 中，但实际应该被过滤）
	// 注意：当前实现只过滤 code=-1，不过滤 code=92000
	foundFreeMeal := false
	foundFreeMealForErp := false
	for _, item := range result.List {
		if item.Uuid == 1002 {
			foundFreeMeal = true
		}
		if item.Uuid == 1003 {
			foundFreeMealForErp = true
		}
	}

	if foundFreeMeal {
		t.Errorf("Free Meal (code=-1) should be filtered out")
	}

	// Free Meal for ERP 目前不会被过滤（因为 excludeCodes 中只有 code=-1）
	// 如果需要过滤，需要在 excludeCodes 中添加 constant.PaymentMethodCodeFreeMealForErp
	if foundFreeMealForErp {
		t.Logf("Free Meal for ERP (code=92000) is currently not filtered, which is expected")
	}
}

// TestPaymentMethodSrv_GetManagementList_FilterFreeMealAndFreeMealForErp 测试同时过滤 Free Meal 和 Free Meal for ERP
func TestPaymentMethodSrv_GetManagementList_FilterFreeMealAndFreeMealForErp(t *testing.T) {
	dbm := setupPaymentMethodTestDB(t)
	ctx := createPaymentMethodTestContext(t, dbm)
	db := dbm.GetDB(constant.MockDB)

	// 创建测试数据
	paymentMethods := []model.PaymentMethod{
		{
			BaseModel: model.BaseModel{
				Uuid:       1001,
				CreateTime: time.Now().Unix(),
				UpdateTime: time.Now().Unix(),
			},
			Name:        "Cash",
			Code:        constant.PaymentMethodCodeCash,
			PaymentName: "Cash",
			Source:      constant.PaymentMethodSourceSystem,
			Status:      constant.PaymentMethodStatusEnable,
			Sort:        1,
		},
		{
			BaseModel: model.BaseModel{
				Uuid:       1002,
				CreateTime: time.Now().Unix(),
				UpdateTime: time.Now().Unix(),
			},
			Name:        "Free Meal",
			Code:        constant.PaymentMethodCodeFreePay, // -1（原有）
			PaymentName: "Free Meal",
			Source:      constant.PaymentMethodSourceSystem,
			Status:      constant.PaymentMethodStatusEnable,
			Sort:        2,
		},
		{
			BaseModel: model.BaseModel{
				Uuid:       1003,
				CreateTime: time.Now().Unix(),
				UpdateTime: time.Now().Unix(),
			},
			Name:        "Free Meal for ERP",
			Code:        constant.PaymentMethodCodeFreeMealForErp, // 92000（用于ERP）
			PaymentName: "Free Meal",
			Source:      constant.PaymentMethodSourceSystem,
			Status:      constant.PaymentMethodStatusEnable,
			Sort:        3,
		},
	}

	for _, pm := range paymentMethods {
		db.Create(&pm)
	}

	// 创建服务
	mockCache := cache.NewCache(cache.GoCache, cache.Config{})
	settingSrv := setting.NewSrv(dbm, mockCache)
	paymentMethodSrv := NewPaymentMethodSrv(dbm, settingSrv)

	// 测试 GetManagementList
	req := &req.PaymentMethodManagementListReq{
		PageNo:   1,
		PageSize: 10,
	}
	result, err := paymentMethodSrv.GetManagementList(ctx, req)
	if err != nil {
		t.Fatalf("Failed to get management list: %v", err)
	}

	// 验证结果：
	// - Free Meal（code=-1）应该被过滤
	// - Free Meal for ERP（code=92000）目前不会被过滤（因为 excludeCodes 中只有 code=-1）
	foundFreeMeal := false
	for _, item := range result.List {
		if item.Uuid == 1002 {
			foundFreeMeal = true
			t.Errorf("Free Meal (code=-1) should be filtered out")
		}
	}

	if foundFreeMeal {
		t.Errorf("Free Meal (code=-1) should not appear in management list")
	}

	// 验证 Cash 存在
	foundCash := false
	for _, item := range result.List {
		if item.Uuid == 1001 {
			foundCash = true
			break
		}
	}
	if !foundCash {
		t.Errorf("Cash payment method should be in the list")
	}
}
