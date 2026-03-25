package tasks

import (
	"errors"
	"os"
	"testing"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service/purchase_order"
	"ttpos-server-go/app/service/rpc/erp"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/logger"

	"go.uber.org/zap"
	"ttpos-bmp/app/ttpos-erp/api/delivery_note"
)

func TestMain(m *testing.M) {
	logger.Logger = zap.NewNop()
	os.Exit(m.Run())
}

// ---------------------------------------------------------------------------
// shouldAutoReceipt
// ---------------------------------------------------------------------------

func TestShouldAutoReceipt(t *testing.T) {
	t.Parallel()
	loc := time.FixedZone("UTC+8", 8*3600)
	// 用一个固定的"今天"来构造测试用例，避免时间漂移
	now := time.Now().In(loc)
	today := now.Format("2006-01-02")
	yesterday := now.AddDate(0, 0, -1).Format("2006-01-02")
	twoDaysAgo := now.AddDate(0, 0, -2).Format("2006-01-02")
	threeDaysAgo := now.AddDate(0, 0, -3).Format("2006-01-02")
	tomorrow := now.AddDate(0, 0, 1).Format("2006-01-02")

	tests := []struct {
		name          string
		dnPostingDate string
		delayDays     int
		want          bool
	}{
		// deadline = postingDate + delayDays + 1 天的 0:00
		// 今天过账+0天延迟 → deadline=明天0:00 → now < deadline → false
		{"今天过账-0天延迟-不收货", today, 0, false},
		// 昨天过账+0天延迟 → deadline=今天0:00 → now >= deadline → true
		{"昨天过账-0天延迟-收货", yesterday, 0, true},
		// 两天前过账+1天延迟 → deadline=昨天0:00+1天=今天0:00 → now >= deadline → true
		{"两天前过账-1天延迟-收货", twoDaysAgo, 1, true},
		// 昨天过账+1天延迟 → deadline=明天0:00 → now < deadline → false
		{"昨天过账-1天延迟-不收货", yesterday, 1, false},
		// 三天前过账+2天延迟 → deadline=今天0:00 → now >= deadline → true
		{"三天前过账-2天延迟-收货", threeDaysAgo, 2, true},
		// 明天过账+0天延迟 → deadline=后天0:00 → false
		{"明天过账-0天延迟-不收货", tomorrow, 0, false},
		// 非法日期格式 → false
		{"非法日期格式", "not-a-date", 0, false},
		// 空字符串 → false
		{"空字符串日期", "", 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := shouldAutoReceipt(tt.dnPostingDate, tt.delayDays, loc)
			if got != tt.want {
				t.Errorf("shouldAutoReceipt(%q, %d) = %v, want %v", tt.dnPostingDate, tt.delayDays, got, tt.want)
			}
		})
	}
}

func TestShouldAutoReceipt_DifferentTimezones(t *testing.T) {
	t.Parallel()
	// 构造一个刚好在 UTC+8 到了但 UTC+12 还没到的场景
	// 用远期日期确保确定性
	farPast := "2020-01-01"
	locPlus8 := time.FixedZone("UTC+8", 8*3600)
	locPlus12 := time.FixedZone("UTC+12", 12*3600)

	// 远期过账，两个时区都应该返回 true
	if !shouldAutoReceipt(farPast, 0, locPlus8) {
		t.Error("UTC+8 远期日期应该返回 true")
	}
	if !shouldAutoReceipt(farPast, 0, locPlus12) {
		t.Error("UTC+12 远期日期应该返回 true")
	}
}

// ---------------------------------------------------------------------------
// findMatchingRule
// ---------------------------------------------------------------------------

func TestFindMatchingRule(t *testing.T) {
	t.Parallel()
	rules := []shopRuleInfo{
		{RuleUuid: 1, WarehouseErpCode: "WH-001"},
		{RuleUuid: 2, WarehouseErpCode: "WH-002"},
		{RuleUuid: 3, WarehouseErpCode: "WH-003"},
	}

	tests := []struct {
		name        string
		dnWarehouse string
		wantUuid    uint64
		wantNil     bool
	}{
		{"匹配第一个", "WH-001", 1, false},
		{"匹配最后一个", "WH-003", 3, false},
		{"不匹配", "WH-999", 0, true},
		{"空字符串", "", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := findMatchingRule(tt.dnWarehouse, rules)
			if tt.wantNil {
				if got != nil {
					t.Errorf("expected nil, got rule uuid=%d", got.RuleUuid)
				}
				return
			}
			if got == nil {
				t.Fatal("expected non-nil result")
			}
			if got.RuleUuid != tt.wantUuid {
				t.Errorf("got RuleUuid=%d, want %d", got.RuleUuid, tt.wantUuid)
			}
		})
	}
}

func TestFindMatchingRule_EmptyRules(t *testing.T) {
	t.Parallel()
	got := findMatchingRule("WH-001", nil)
	if got != nil {
		t.Errorf("空规则列表应返回nil, got %+v", got)
	}
}

// ---------------------------------------------------------------------------
// buildArrivedQtyMap
// ---------------------------------------------------------------------------

func TestBuildArrivedQtyMap(t *testing.T) {
	t.Parallel()
	receipts := []model.PurchaseReceiptOrder{
		{
			DeliveryNoteNo: "DN-001",
			Items: []model.PurchaseReceiptOrderItem{
				{MaterialCode: "MAT-A", ErpnextUom: "KG", Num: 10},
				{
					MaterialCode: "MAT-B",
					Units: []model.PurchaseReceiptOrderItemUnit{
						{ErpnextUom: "PC", Num: 5},
						{ErpnextUom: "BOX", Num: 2},
					},
				},
			},
		},
		{
			DeliveryNoteNo: "DN-002", // 不同 DN，不应计入
			Items: []model.PurchaseReceiptOrderItem{
				{MaterialCode: "MAT-A", ErpnextUom: "KG", Num: 99},
			},
		},
		{
			DeliveryNoteNo: "DN-001", // 同 DN，应累加
			Items: []model.PurchaseReceiptOrderItem{
				{MaterialCode: "MAT-A", ErpnextUom: "KG", Num: 3},
			},
		},
	}

	got := buildArrivedQtyMap(receipts, "DN-001")

	tests := []struct {
		key  string
		want float64
	}{
		{"MAT-A:KG", 13}, // 10 + 3
		{"MAT-B:PC", 5},  // 来自多单位
		{"MAT-B:BOX", 2}, // 来自多单位
		{"MAT-A:PC", 0},  // 不存在
		{"MAT-C:KG", 0},  // 不存在
	}
	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			t.Parallel()
			if got[tt.key] != tt.want {
				t.Errorf("arrivedQtyMap[%q] = %v, want %v", tt.key, got[tt.key], tt.want)
			}
		})
	}
}

func TestBuildArrivedQtyMap_Empty(t *testing.T) {
	t.Parallel()
	got := buildArrivedQtyMap(nil, "DN-001")
	if len(got) != 0 {
		t.Errorf("空列表应返回空map, got len=%d", len(got))
	}
}

func TestBuildArrivedQtyMap_NoMatchingDN(t *testing.T) {
	t.Parallel()
	receipts := []model.PurchaseReceiptOrder{
		{
			DeliveryNoteNo: "DN-OTHER",
			Items: []model.PurchaseReceiptOrderItem{
				{MaterialCode: "MAT-A", ErpnextUom: "KG", Num: 10},
			},
		},
	}
	got := buildArrivedQtyMap(receipts, "DN-001")
	if len(got) != 0 {
		t.Errorf("无匹配DN应返回空map, got len=%d", len(got))
	}
}

// ---------------------------------------------------------------------------
// findMatchingUnitUuid
// ---------------------------------------------------------------------------

func TestFindMatchingUnitUuid(t *testing.T) {
	t.Parallel()
	orderItem := &model.PurchaseOrderItem{
		ErpnextUom: "KG",
		UnitUuid:   100,
		Units: []model.PurchaseOrderItemUnit{
			{ErpnextUom: "PC", UnitUuid: 201},
			{ErpnextUom: "BOX", UnitUuid: 202},
		},
	}

	tests := []struct {
		name      string
		uom       string
		wantUuid  uint64
		wantFound bool
	}{
		{"匹配多单位-PC", "PC", 201, true},
		{"匹配多单位-BOX", "BOX", 202, true},
		{"回退到明细单位-KG", "KG", 100, true},
		{"不匹配", "LITER", 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			uuid, found := findMatchingUnitUuid(orderItem, tt.uom)
			if found != tt.wantFound {
				t.Errorf("found=%v, want %v", found, tt.wantFound)
			}
			if uuid != tt.wantUuid {
				t.Errorf("uuid=%d, want %d", uuid, tt.wantUuid)
			}
		})
	}
}

func TestFindMatchingUnitUuid_NoUnits(t *testing.T) {
	t.Parallel()
	orderItem := &model.PurchaseOrderItem{
		ErpnextUom: "KG",
		UnitUuid:   100,
	}
	uuid, found := findMatchingUnitUuid(orderItem, "KG")
	if !found || uuid != 100 {
		t.Errorf("无多单位应回退到明细单位: found=%v, uuid=%d", found, uuid)
	}
}

func TestFindMatchingUnitUuid_ZeroUnitUuid(t *testing.T) {
	t.Parallel()
	orderItem := &model.PurchaseOrderItem{
		ErpnextUom: "KG",
		UnitUuid:   0, // 无效
	}
	_, found := findMatchingUnitUuid(orderItem, "KG")
	if found {
		t.Error("UnitUuid=0 时不应匹配")
	}
}

// ---------------------------------------------------------------------------
// calculatePendingItems
// ---------------------------------------------------------------------------

func TestCalculatePendingItems(t *testing.T) {
	t.Parallel()
	dn := &delivery_note.DeliveryNote{
		Name: "DN-001",
		Items: []*delivery_note.DeliveryNoteItem{
			{ItemCode: "MAT-A", Qty: 10, Uom: "KG"},
			{ItemCode: "MAT-B", Qty: 5, Uom: "PC"},
			{ItemCode: "MAT-C", Qty: 3, Uom: "BOX"}, // 不在采购订单中
		},
	}
	orderItems := []model.PurchaseOrderItem{
		{
			BaseModel:    model.BaseModel{Uuid: 1001},
			MaterialCode: "MAT-A",
			ErpnextUom:   "KG",
			UnitUuid:     100,
		},
		{
			BaseModel:    model.BaseModel{Uuid: 1002},
			MaterialCode: "MAT-B",
			Units: []model.PurchaseOrderItemUnit{
				{ErpnextUom: "PC", UnitUuid: 201},
			},
		},
	}
	arrivedQtyMap := map[string]float64{
		"MAT-A:KG": 3, // 已收3，待收7
		"MAT-B:PC": 5, // 已收5，待收0
	}

	items := calculatePendingItems(dn, orderItems, arrivedQtyMap)

	// MAT-A: 10-3=7 > 0 → 应该有
	// MAT-B: 5-5=0 → 不应该有
	// MAT-C: 不在采购订单中 → 跳过
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].PurchaseOrderItemUuid != 1001 {
		t.Errorf("PurchaseOrderItemUuid = %d, want 1001", items[0].PurchaseOrderItemUuid)
	}
	if len(items[0].UnitList) != 1 {
		t.Fatalf("expected 1 unit, got %d", len(items[0].UnitList))
	}
	if items[0].UnitList[0].Num != 7 {
		t.Errorf("pending qty = %v, want 7", items[0].UnitList[0].Num)
	}
	if items[0].UnitList[0].Uuid != 100 {
		t.Errorf("unit uuid = %d, want 100", items[0].UnitList[0].Uuid)
	}
}

func TestCalculatePendingItems_AllFullyReceived(t *testing.T) {
	t.Parallel()
	dn := &delivery_note.DeliveryNote{
		Name: "DN-001",
		Items: []*delivery_note.DeliveryNoteItem{
			{ItemCode: "MAT-A", Qty: 10, Uom: "KG"},
		},
	}
	orderItems := []model.PurchaseOrderItem{
		{
			BaseModel:    model.BaseModel{Uuid: 1001},
			MaterialCode: "MAT-A",
			ErpnextUom:   "KG",
			UnitUuid:     100,
		},
	}
	arrivedQtyMap := map[string]float64{
		"MAT-A:KG": 10,
	}

	items := calculatePendingItems(dn, orderItems, arrivedQtyMap)
	if len(items) != 0 {
		t.Errorf("全部已收，expected 0 items, got %d", len(items))
	}
}

func TestCalculatePendingItems_OverReceived(t *testing.T) {
	t.Parallel()
	dn := &delivery_note.DeliveryNote{
		Name: "DN-001",
		Items: []*delivery_note.DeliveryNoteItem{
			{ItemCode: "MAT-A", Qty: 10, Uom: "KG"},
		},
	}
	orderItems := []model.PurchaseOrderItem{
		{
			BaseModel:    model.BaseModel{Uuid: 1001},
			MaterialCode: "MAT-A",
			ErpnextUom:   "KG",
			UnitUuid:     100,
		},
	}
	arrivedQtyMap := map[string]float64{
		"MAT-A:KG": 15, // 超收
	}

	items := calculatePendingItems(dn, orderItems, arrivedQtyMap)
	if len(items) != 0 {
		t.Errorf("超收场景应返回0项, got %d", len(items))
	}
}

func TestCalculatePendingItems_EmptyDN(t *testing.T) {
	t.Parallel()
	dn := &delivery_note.DeliveryNote{Name: "DN-001"}
	items := calculatePendingItems(dn, nil, nil)
	if len(items) != 0 {
		t.Errorf("空DN应返回空, got %d", len(items))
	}
}

func TestCalculatePendingItems_MultiUnit(t *testing.T) {
	t.Parallel()
	dn := &delivery_note.DeliveryNote{
		Name: "DN-001",
		Items: []*delivery_note.DeliveryNoteItem{
			{ItemCode: "MAT-A", Qty: 10, Uom: "PC"},
		},
	}
	orderItems := []model.PurchaseOrderItem{
		{
			BaseModel:    model.BaseModel{Uuid: 1001},
			MaterialCode: "MAT-A",
			Units: []model.PurchaseOrderItemUnit{
				{ErpnextUom: "PC", UnitUuid: 201},
				{ErpnextUom: "BOX", UnitUuid: 202},
			},
		},
	}

	items := calculatePendingItems(dn, orderItems, map[string]float64{})
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].UnitList[0].Uuid != 201 {
		t.Errorf("should match PC unit uuid=201, got %d", items[0].UnitList[0].Uuid)
	}
	if items[0].UnitList[0].Num != 10 {
		t.Errorf("pending qty = %v, want 10", items[0].UnitList[0].Num)
	}
}

// ===========================================================================
// Mock types for processOrder and executeAutoReceipt tests
// ===========================================================================

// mockErpSrv 仅实现 GetDeliveryNoteList，其他方法通过嵌入接口（调用则 panic）
type mockErpSrv struct {
	erp.IErpSrv
	getDNListFn func(ctx context.Context, req *delivery_note.GetDeliveryNoteListReq) (*delivery_note.GetDeliveryNoteListResp, error)
}

func (m *mockErpSrv) GetDeliveryNoteList(ctx context.Context, req *delivery_note.GetDeliveryNoteListReq) (*delivery_note.GetDeliveryNoteListResp, error) {
	return m.getDNListFn(ctx, req)
}

// mockPurchaseOrderSrv 仅实现 CreatePurchaseReceiptOrder
type mockPurchaseOrderSrv struct {
	purchase_order.IPurchaseOrderSrv
	createReceiptFn func(ctx context.Context, req req.PurchaseReceiptCreateReq) (resp.PurchaseReceiptOrderCreateResp, error)
}

func (m *mockPurchaseOrderSrv) CreatePurchaseReceiptOrder(ctx context.Context, r req.PurchaseReceiptCreateReq) (resp.PurchaseReceiptOrderCreateResp, error) {
	return m.createReceiptFn(ctx, r)
}

// mockLogRepo 实现 IAutoReceiptLogRepo
type mockLogRepo struct {
	createFn func(log model.AutoReceiptLog) (model.AutoReceiptLog, error)
}

func (m *mockLogRepo) Create(log model.AutoReceiptLog) (model.AutoReceiptLog, error) {
	if m.createFn != nil {
		return m.createFn(log)
	}
	return log, nil
}

func (m *mockLogRepo) GetByUuid(uuid uint64, headquarterUuid uint64) (model.AutoReceiptLog, error) {
	return model.AutoReceiptLog{}, nil
}

func (m *mockLogRepo) GetList(headquarterUuid uint64, shopCompanyUuid uint64, startTime int64, endTime int64, pageNo int, pageSize int) ([]model.AutoReceiptLog, int64, error) {
	return nil, 0, nil
}

// newTestContext 创建测试用 context
func newTestContext(companyUuid uint64, headquarterAbbr string) context.Context {
	companySetting := model.CompanySetting{
		ErpnextHeadquarterAbbr: headquarterAbbr,
	}
	company := model.Company{
		CompanySetting: &companySetting,
	}
	return context.NewContext(
		context.WithCompanyUuid(companyUuid),
		context.WithCompany(company),
		context.WithCompanySetting(companySetting),
	)
}

// ---------------------------------------------------------------------------
// processOrder
// ---------------------------------------------------------------------------

func TestProcessOrder_GetDNListError(t *testing.T) {
	t.Parallel()
	task := &AutoReceiptTask{}
	ctx := newTestContext(100, "HQ")
	order := &model.PurchaseOrder{ErpSaleOrderNo: "SO-001"}
	rules := []shopRuleInfo{{RuleUuid: 1, WarehouseErpCode: "WH-001"}}
	loc := time.FixedZone("UTC+8", 8*3600)

	erpMock := &mockErpSrv{
		getDNListFn: func(_ context.Context, _ *delivery_note.GetDeliveryNoteListReq) (*delivery_note.GetDeliveryNoteListResp, error) {
			return nil, errors.New("rpc error")
		},
	}

	// Should not panic — just log and return
	task.processOrder(ctx, order, rules, loc, 100, orderProcessDeps{erpSrv: erpMock}, nil)
}

func TestProcessOrder_EmptyDNList(t *testing.T) {
	t.Parallel()
	task := &AutoReceiptTask{}
	ctx := newTestContext(100, "HQ")
	order := &model.PurchaseOrder{ErpSaleOrderNo: "SO-001"}
	rules := []shopRuleInfo{{RuleUuid: 1, WarehouseErpCode: "WH-001"}}
	loc := time.FixedZone("UTC+8", 8*3600)

	erpMock := &mockErpSrv{
		getDNListFn: func(_ context.Context, _ *delivery_note.GetDeliveryNoteListReq) (*delivery_note.GetDeliveryNoteListResp, error) {
			return &delivery_note.GetDeliveryNoteListResp{DeliveryNoteList: nil}, nil
		},
	}

	task.processOrder(ctx, order, rules, loc, 100, orderProcessDeps{erpSrv: erpMock}, nil)
}

func TestProcessOrder_NilDNListResp(t *testing.T) {
	t.Parallel()
	task := &AutoReceiptTask{}
	ctx := newTestContext(100, "HQ")
	order := &model.PurchaseOrder{ErpSaleOrderNo: "SO-001"}
	rules := []shopRuleInfo{{RuleUuid: 1, WarehouseErpCode: "WH-001"}}
	loc := time.FixedZone("UTC+8", 8*3600)

	erpMock := &mockErpSrv{
		getDNListFn: func(_ context.Context, _ *delivery_note.GetDeliveryNoteListReq) (*delivery_note.GetDeliveryNoteListResp, error) {
			return nil, nil
		},
	}

	task.processOrder(ctx, order, rules, loc, 100, orderProcessDeps{erpSrv: erpMock}, nil)
}

func TestProcessOrder_DNStatusFilter(t *testing.T) {
	t.Parallel()
	task := &AutoReceiptTask{}
	ctx := newTestContext(100, "HQ")
	order := &model.PurchaseOrder{
		BaseModel:      model.BaseModel{Uuid: 2001},
		ErpSaleOrderNo: "SO-001",
	}
	rules := []shopRuleInfo{{RuleUuid: 1, WarehouseErpCode: "WH-001", DelayDays: 0}}
	loc := time.FixedZone("UTC+8", 8*3600)

	erpMock := &mockErpSrv{
		getDNListFn: func(_ context.Context, _ *delivery_note.GetDeliveryNoteListReq) (*delivery_note.GetDeliveryNoteListResp, error) {
			return &delivery_note.GetDeliveryNoteListResp{
				DeliveryNoteList: []*delivery_note.DeliveryNote{
					{Name: "DN-DRAFT", Status: "Draft", Docstatus: 0, SetWarehouse: "WH-001"},
					{Name: "DN-CANCELLED", Status: "Cancelled", Docstatus: 2, SetWarehouse: "WH-001"},
				},
			}, nil
		},
	}

	// None of the DNs should be processed (wrong status) — no purchaseOrderSrv call
	task.processOrder(ctx, order, rules, loc, 100, orderProcessDeps{erpSrv: erpMock}, nil)
}

func TestProcessOrder_NoMatchingWarehouse(t *testing.T) {
	t.Parallel()
	task := &AutoReceiptTask{}
	ctx := newTestContext(100, "HQ")
	farPast := time.Now().AddDate(0, 0, -10).Format("2006-01-02")
	order := &model.PurchaseOrder{
		BaseModel:      model.BaseModel{Uuid: 2001},
		ErpSaleOrderNo: "SO-001",
	}
	rules := []shopRuleInfo{{RuleUuid: 1, WarehouseErpCode: "WH-001", DelayDays: 0}}
	loc := time.FixedZone("UTC+8", 8*3600)

	erpMock := &mockErpSrv{
		getDNListFn: func(_ context.Context, _ *delivery_note.GetDeliveryNoteListReq) (*delivery_note.GetDeliveryNoteListResp, error) {
			return &delivery_note.GetDeliveryNoteListResp{
				DeliveryNoteList: []*delivery_note.DeliveryNote{
					{Name: "DN-001", Status: "To Bill", Docstatus: 1, SetWarehouse: "WH-999", PostingDate: farPast},
				},
			}, nil
		},
	}

	task.processOrder(ctx, order, rules, loc, 100, orderProcessDeps{erpSrv: erpMock}, nil)
}

func TestProcessOrder_NotYetDue(t *testing.T) {
	t.Parallel()
	task := &AutoReceiptTask{}
	ctx := newTestContext(100, "HQ")
	tomorrow := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	order := &model.PurchaseOrder{
		BaseModel:      model.BaseModel{Uuid: 2001},
		ErpSaleOrderNo: "SO-001",
	}
	rules := []shopRuleInfo{{RuleUuid: 1, WarehouseErpCode: "WH-001", DelayDays: 0}}
	loc := time.FixedZone("UTC+8", 8*3600)

	erpMock := &mockErpSrv{
		getDNListFn: func(_ context.Context, _ *delivery_note.GetDeliveryNoteListReq) (*delivery_note.GetDeliveryNoteListResp, error) {
			return &delivery_note.GetDeliveryNoteListResp{
				DeliveryNoteList: []*delivery_note.DeliveryNote{
					{Name: "DN-001", Status: "To Bill", Docstatus: 1, SetWarehouse: "WH-001", PostingDate: tomorrow},
				},
			}, nil
		},
	}

	task.processOrder(ctx, order, rules, loc, 100, orderProcessDeps{erpSrv: erpMock}, nil)
}

func TestProcessOrder_AllItemsFullyReceived(t *testing.T) {
	t.Parallel()
	task := &AutoReceiptTask{}
	ctx := newTestContext(100, "HQ")
	farPast := time.Now().AddDate(0, 0, -10).Format("2006-01-02")
	order := &model.PurchaseOrder{
		BaseModel:      model.BaseModel{Uuid: 2001},
		ErpSaleOrderNo: "SO-001",
		Items: []model.PurchaseOrderItem{
			{BaseModel: model.BaseModel{Uuid: 3001}, MaterialCode: "MAT-A", ErpnextUom: "KG", UnitUuid: 100},
		},
	}
	rules := []shopRuleInfo{{RuleUuid: 1, WarehouseErpCode: "WH-001", DelayDays: 0}}
	loc := time.FixedZone("UTC+8", 8*3600)

	erpMock := &mockErpSrv{
		getDNListFn: func(_ context.Context, _ *delivery_note.GetDeliveryNoteListReq) (*delivery_note.GetDeliveryNoteListResp, error) {
			return &delivery_note.GetDeliveryNoteListResp{
				DeliveryNoteList: []*delivery_note.DeliveryNote{
					{
						Name: "DN-001", Status: "To Bill", Docstatus: 1, SetWarehouse: "WH-001", PostingDate: farPast,
						Items: []*delivery_note.DeliveryNoteItem{
							{ItemCode: "MAT-A", Qty: 10, Uom: "KG"},
						},
					},
				},
			}, nil
		},
	}

	// Already fully received
	confirmedReceipts := []model.PurchaseReceiptOrder{
		{
			DeliveryNoteNo: "DN-001",
			Items: []model.PurchaseReceiptOrderItem{
				{MaterialCode: "MAT-A", ErpnextUom: "KG", Num: 10},
			},
		},
	}

	// purchaseOrderSrv should NOT be called
	task.processOrder(ctx, order, rules, loc, 100, orderProcessDeps{erpSrv: erpMock}, confirmedReceipts)
}

func TestProcessOrder_HappyPath_CallsExecuteAutoReceipt(t *testing.T) {
	t.Parallel()
	task := &AutoReceiptTask{}
	ctx := newTestContext(100, "HQ")
	farPast := time.Now().AddDate(0, 0, -10).Format("2006-01-02")
	order := &model.PurchaseOrder{
		BaseModel:      model.BaseModel{Uuid: 2001},
		ErpSaleOrderNo: "SO-001",
		Items: []model.PurchaseOrderItem{
			{BaseModel: model.BaseModel{Uuid: 3001}, MaterialCode: "MAT-A", ErpnextUom: "KG", UnitUuid: 100},
		},
	}
	rules := []shopRuleInfo{{RuleUuid: 1, WarehouseErpCode: "WH-001", DelayDays: 0, HeadquarterUuid: 999}}
	loc := time.FixedZone("UTC+8", 8*3600)

	erpMock := &mockErpSrv{
		getDNListFn: func(_ context.Context, _ *delivery_note.GetDeliveryNoteListReq) (*delivery_note.GetDeliveryNoteListResp, error) {
			return &delivery_note.GetDeliveryNoteListResp{
				DeliveryNoteList: []*delivery_note.DeliveryNote{
					{
						Name: "DN-001", Status: "To Bill", Docstatus: 1, SetWarehouse: "WH-001", PostingDate: farPast,
						Items: []*delivery_note.DeliveryNoteItem{
							{ItemCode: "MAT-A", Qty: 10, Uom: "KG"},
						},
					},
				},
			}, nil
		},
	}

	var receiptCreated bool
	poMock := &mockPurchaseOrderSrv{
		createReceiptFn: func(_ context.Context, r req.PurchaseReceiptCreateReq) (resp.PurchaseReceiptOrderCreateResp, error) {
			receiptCreated = true
			if r.PurchaseOrderUuid != 2001 {
				t.Errorf("PurchaseOrderUuid = %d, want 2001", r.PurchaseOrderUuid)
			}
			if !r.IsAutoReceipt {
				t.Error("IsAutoReceipt should be true")
			}
			if r.DeliveryNoteNo != "DN-001" {
				t.Errorf("DeliveryNoteNo = %s, want DN-001", r.DeliveryNoteNo)
			}
			if len(r.Items) != 1 {
				t.Fatalf("expected 1 item, got %d", len(r.Items))
			}
			if r.Items[0].UnitList[0].Num != 10 {
				t.Errorf("pending qty = %v, want 10", r.Items[0].UnitList[0].Num)
			}
			return resp.PurchaseReceiptOrderCreateResp{Uuid: 5001, OrderNo: "RO-001"}, nil
		},
	}

	var logCreated bool
	logMock := &mockLogRepo{
		createFn: func(log model.AutoReceiptLog) (model.AutoReceiptLog, error) {
			logCreated = true
			if log.HeadquarterCompanyUuid != 999 {
				t.Errorf("HeadquarterCompanyUuid = %d, want 999", log.HeadquarterCompanyUuid)
			}
			if log.ReceiptOrderUuid != 5001 {
				t.Errorf("ReceiptOrderUuid = %d, want 5001", log.ReceiptOrderUuid)
			}
			if log.ReceiptOrderNo != "RO-001" {
				t.Errorf("ReceiptOrderNo = %s, want RO-001", log.ReceiptOrderNo)
			}
			if log.ReceiptErpOrderNo != "DN-001" {
				t.Errorf("ReceiptErpOrderNo = %s, want DN-001", log.ReceiptErpOrderNo)
			}
			return log, nil
		},
	}

	// No confirmed receipts → all 10 KG pending
	task.processOrder(ctx, order, rules, loc, 100, orderProcessDeps{erpSrv: erpMock, purchaseOrderSrv: poMock, logRepo: logMock}, nil)

	if !receiptCreated {
		t.Error("CreatePurchaseReceiptOrder was not called")
	}
	if !logCreated {
		t.Error("logRepo.Create was not called")
	}
}

func TestProcessOrder_MultipleDNs_MixedFiltering(t *testing.T) {
	t.Parallel()
	task := &AutoReceiptTask{}
	ctx := newTestContext(100, "HQ")
	farPast := time.Now().AddDate(0, 0, -10).Format("2006-01-02")
	tomorrow := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	order := &model.PurchaseOrder{
		BaseModel:      model.BaseModel{Uuid: 2001},
		ErpSaleOrderNo: "SO-001",
		Items: []model.PurchaseOrderItem{
			{BaseModel: model.BaseModel{Uuid: 3001}, MaterialCode: "MAT-A", ErpnextUom: "KG", UnitUuid: 100},
		},
	}
	rules := []shopRuleInfo{
		{RuleUuid: 1, WarehouseErpCode: "WH-001", DelayDays: 0, HeadquarterUuid: 999},
		{RuleUuid: 2, WarehouseErpCode: "WH-002", DelayDays: 0, HeadquarterUuid: 999},
	}
	loc := time.FixedZone("UTC+8", 8*3600)

	erpMock := &mockErpSrv{
		getDNListFn: func(_ context.Context, _ *delivery_note.GetDeliveryNoteListReq) (*delivery_note.GetDeliveryNoteListResp, error) {
			return &delivery_note.GetDeliveryNoteListResp{
				DeliveryNoteList: []*delivery_note.DeliveryNote{
					// DN1: wrong status → skip
					{Name: "DN-DRAFT", Status: "Draft", Docstatus: 0, SetWarehouse: "WH-001"},
					// DN2: no matching warehouse → skip
					{Name: "DN-NOMATCH", Status: "To Bill", Docstatus: 1, SetWarehouse: "WH-999", PostingDate: farPast},
					// DN3: not yet due → skip
					{Name: "DN-FUTURE", Status: "To Bill", Docstatus: 1, SetWarehouse: "WH-001", PostingDate: tomorrow},
					// DN4: valid, should be processed
					{
						Name: "DN-GOOD", Status: "To Bill", Docstatus: 1, SetWarehouse: "WH-002", PostingDate: farPast,
						Items: []*delivery_note.DeliveryNoteItem{
							{ItemCode: "MAT-A", Qty: 5, Uom: "KG"},
						},
					},
				},
			}, nil
		},
	}

	var createdCount int
	poMock := &mockPurchaseOrderSrv{
		createReceiptFn: func(_ context.Context, r req.PurchaseReceiptCreateReq) (resp.PurchaseReceiptOrderCreateResp, error) {
			createdCount++
			if r.DeliveryNoteNo != "DN-GOOD" {
				t.Errorf("expected DN-GOOD, got %s", r.DeliveryNoteNo)
			}
			return resp.PurchaseReceiptOrderCreateResp{Uuid: 5001, OrderNo: "RO-001"}, nil
		},
	}
	logMock := &mockLogRepo{}

	task.processOrder(ctx, order, rules, loc, 100, orderProcessDeps{erpSrv: erpMock, purchaseOrderSrv: poMock, logRepo: logMock}, nil)

	if createdCount != 1 {
		t.Errorf("expected exactly 1 receipt created, got %d", createdCount)
	}
}

// ---------------------------------------------------------------------------
// executeAutoReceipt
// ---------------------------------------------------------------------------

func TestExecuteAutoReceipt_Success(t *testing.T) {
	t.Parallel()
	task := &AutoReceiptTask{}
	ctx := newTestContext(100, "HQ")
	dn := &delivery_note.DeliveryNote{Name: "DN-001"}
	order := &model.PurchaseOrder{BaseModel: model.BaseModel{Uuid: 2001}}
	pendingItems := []req.PurchaseReceiptItemCreateReq{
		{PurchaseOrderItemUuid: 3001, UnitList: []req.PurchaseReceiptItemMaterialUnitReq{{Uuid: 100, Num: 7}}},
	}
	rule := &shopRuleInfo{RuleUuid: 1, HeadquarterUuid: 999}

	var receiptCalled, logCalled bool
	poMock := &mockPurchaseOrderSrv{
		createReceiptFn: func(_ context.Context, r req.PurchaseReceiptCreateReq) (resp.PurchaseReceiptOrderCreateResp, error) {
			receiptCalled = true
			if r.ReceiptType != constant.ReceiptTypeInternal {
				t.Errorf("ReceiptType = %d, want %d", r.ReceiptType, constant.ReceiptTypeInternal)
			}
			if !r.IsConfirm {
				t.Error("IsConfirm should be true")
			}
			return resp.PurchaseReceiptOrderCreateResp{Uuid: 5001, OrderNo: "RO-001"}, nil
		},
	}
	logMock := &mockLogRepo{
		createFn: func(log model.AutoReceiptLog) (model.AutoReceiptLog, error) {
			logCalled = true
			if log.ShopCompanyUuid != 100 {
				t.Errorf("ShopCompanyUuid = %d, want 100", log.ShopCompanyUuid)
			}
			return log, nil
		},
	}

	task.executeAutoReceipt(ctx, autoReceiptInput{
		order: order, dn: dn, pendingItems: pendingItems,
		rule: rule, shopUuid: 100,
	}, orderProcessDeps{purchaseOrderSrv: poMock, logRepo: logMock})

	if !receiptCalled {
		t.Error("CreatePurchaseReceiptOrder was not called")
	}
	if !logCalled {
		t.Error("logRepo.Create was not called")
	}
}

func TestExecuteAutoReceipt_CreateReceiptFails(t *testing.T) {
	t.Parallel()
	task := &AutoReceiptTask{}
	ctx := newTestContext(100, "HQ")
	dn := &delivery_note.DeliveryNote{Name: "DN-001"}
	order := &model.PurchaseOrder{BaseModel: model.BaseModel{Uuid: 2001}}
	pendingItems := []req.PurchaseReceiptItemCreateReq{
		{PurchaseOrderItemUuid: 3001, UnitList: []req.PurchaseReceiptItemMaterialUnitReq{{Uuid: 100, Num: 7}}},
	}
	rule := &shopRuleInfo{RuleUuid: 1, HeadquarterUuid: 999}

	poMock := &mockPurchaseOrderSrv{
		createReceiptFn: func(_ context.Context, _ req.PurchaseReceiptCreateReq) (resp.PurchaseReceiptOrderCreateResp, error) {
			return resp.PurchaseReceiptOrderCreateResp{}, errors.New("create failed")
		},
	}

	var logCalled bool
	logMock := &mockLogRepo{
		createFn: func(log model.AutoReceiptLog) (model.AutoReceiptLog, error) {
			logCalled = true
			return log, nil
		},
	}

	// Should return early without calling logRepo.Create
	task.executeAutoReceipt(ctx, autoReceiptInput{
		order: order, dn: dn, pendingItems: pendingItems,
		rule: rule, shopUuid: 100,
	}, orderProcessDeps{purchaseOrderSrv: poMock, logRepo: logMock})

	if logCalled {
		t.Error("logRepo.Create should NOT be called when receipt creation fails")
	}
}

func TestExecuteAutoReceipt_LogCreateFails(t *testing.T) {
	t.Parallel()
	task := &AutoReceiptTask{}
	ctx := newTestContext(100, "HQ")
	dn := &delivery_note.DeliveryNote{Name: "DN-001"}
	order := &model.PurchaseOrder{BaseModel: model.BaseModel{Uuid: 2001}}
	pendingItems := []req.PurchaseReceiptItemCreateReq{
		{PurchaseOrderItemUuid: 3001, UnitList: []req.PurchaseReceiptItemMaterialUnitReq{{Uuid: 100, Num: 7}}},
	}
	rule := &shopRuleInfo{RuleUuid: 1, HeadquarterUuid: 999}

	poMock := &mockPurchaseOrderSrv{
		createReceiptFn: func(_ context.Context, _ req.PurchaseReceiptCreateReq) (resp.PurchaseReceiptOrderCreateResp, error) {
			return resp.PurchaseReceiptOrderCreateResp{Uuid: 5001, OrderNo: "RO-001"}, nil
		},
	}
	logMock := &mockLogRepo{
		createFn: func(log model.AutoReceiptLog) (model.AutoReceiptLog, error) {
			return log, errors.New("log create failed")
		},
	}

	// Should not panic — just log the error
	task.executeAutoReceipt(ctx, autoReceiptInput{
		order: order, dn: dn, pendingItems: pendingItems,
		rule: rule, shopUuid: 100,
	}, orderProcessDeps{purchaseOrderSrv: poMock, logRepo: logMock})
}

// ---------------------------------------------------------------------------
// NewAutoReceiptTask
// ---------------------------------------------------------------------------

func TestNewAutoReceiptTask(t *testing.T) {
	t.Parallel()
	task := NewAutoReceiptTask(nil, nil)
	if task == nil {
		t.Fatal("NewAutoReceiptTask returned nil")
	}
}

// ---------------------------------------------------------------------------
// calculatePendingItems — UOM mismatch branch
// ---------------------------------------------------------------------------

func TestCalculatePendingItems_UomMismatch(t *testing.T) {
	t.Parallel()
	dn := &delivery_note.DeliveryNote{
		Name: "DN-001",
		Items: []*delivery_note.DeliveryNoteItem{
			{ItemCode: "MAT-A", Qty: 10, Uom: "LITER"}, // UOM 不匹配任何单位
		},
	}
	orderItems := []model.PurchaseOrderItem{
		{
			BaseModel:    model.BaseModel{Uuid: 1001},
			MaterialCode: "MAT-A",
			ErpnextUom:   "KG",
			UnitUuid:     100,
			Units: []model.PurchaseOrderItemUnit{
				{ErpnextUom: "PC", UnitUuid: 201},
			},
		},
	}

	items := calculatePendingItems(dn, orderItems, map[string]float64{})
	if len(items) != 0 {
		t.Errorf("UOM不匹配应跳过, expected 0 items, got %d", len(items))
	}
}

// ---------------------------------------------------------------------------
// groupRulesByShop
// ---------------------------------------------------------------------------

func TestGroupRulesByShop(t *testing.T) {
	t.Parallel()
	ruleShops := []repository.RuleShopJoinResult{
		{RuleUuid: 1, WarehouseErpCode: "WH-001", DelayDays: 0, ShopCompanyUuid: 100, HeadquarterUuid: 999},
		{RuleUuid: 2, WarehouseErpCode: "WH-002", DelayDays: 1, ShopCompanyUuid: 100, HeadquarterUuid: 999},
		{RuleUuid: 1, WarehouseErpCode: "WH-001", DelayDays: 0, ShopCompanyUuid: 200, HeadquarterUuid: 999},
	}

	result := groupRulesByShop(ruleShops)

	if len(result) != 2 {
		t.Fatalf("expected 2 shops, got %d", len(result))
	}
	if len(result[100]) != 2 {
		t.Errorf("shop 100 should have 2 rules, got %d", len(result[100]))
	}
	if len(result[200]) != 1 {
		t.Errorf("shop 200 should have 1 rule, got %d", len(result[200]))
	}
	// 验证字段映射正确
	rule := result[100][0]
	if rule.RuleUuid != 1 || rule.WarehouseErpCode != "WH-001" || rule.DelayDays != 0 || rule.HeadquarterUuid != 999 {
		t.Errorf("rule field mapping incorrect: %+v", rule)
	}
}

func TestGroupRulesByShop_Empty(t *testing.T) {
	t.Parallel()
	result := groupRulesByShop(nil)
	if len(result) != 0 {
		t.Errorf("empty input should return empty map, got %d", len(result))
	}
}

// Verify unused import suppression
var _ repository.IAutoReceiptLogRepo = (*mockLogRepo)(nil)
