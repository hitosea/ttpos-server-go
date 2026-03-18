package service

import (
	"fmt"
	"testing"

	"ttpos-server-go/app/dto/resp"

	"github.com/shopspring/decimal"
)

// ==================== compareStoreCode ====================

func TestCompareStoreCode(t *testing.T) {
	tests := []struct {
		name     string
		a, b     string
		expected bool
	}{
		{"both empty", "", "", false},
		{"a empty b not", "", "001", true},
		{"a not empty b empty", "001", "", false},
		{"digit before letter", "1abc", "abc1", true},
		{"letter after digit", "abc1", "1abc", false},
		{"same type digit compare", "123", "456", true},
		{"same type digit reverse", "456", "123", false},
		{"same type letter compare", "abc", "def", true},
		{"case insensitive", "ABC", "abc", false}, // same after lowering
		{"shorter before longer", "ab", "abc", true},
		{"longer after shorter", "abc", "ab", false},
		{"equal strings", "abc", "abc", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := compareStoreCode(tt.a, tt.b)
			if got != tt.expected {
				t.Errorf("compareStoreCode(%q, %q) = %v, want %v", tt.a, tt.b, got, tt.expected)
			}
		})
	}
}

// ==================== filterCompanyList ====================

func TestFilterCompanyList(t *testing.T) {
	allList := []*resp.CompanySummaryItem{
		{CompanyUuid: 1, CompanyName: "Store A"},
		{CompanyUuid: 2, CompanyName: "Store B"},
		{CompanyUuid: 3, CompanyName: "Store C"},
	}

	t.Run("empty uuids returns all", func(t *testing.T) {
		result := filterCompanyList(allList, nil)
		if len(result) != 3 {
			t.Errorf("expected 3, got %d", len(result))
		}
	})

	t.Run("filter specific uuids", func(t *testing.T) {
		result := filterCompanyList(allList, []uint64{1, 3})
		if len(result) != 2 {
			t.Errorf("expected 2, got %d", len(result))
		}
		if result[0].CompanyUuid != 1 || result[1].CompanyUuid != 3 {
			t.Errorf("unexpected result: %+v", result)
		}
	})

	t.Run("filter non-existent uuid", func(t *testing.T) {
		result := filterCompanyList(allList, []uint64{999})
		if len(result) != 0 {
			t.Errorf("expected 0, got %d", len(result))
		}
	})
}

// ==================== businessItemsHaveData ====================

func TestBusinessItemsHaveData(t *testing.T) {
	t.Run("no data", func(t *testing.T) {
		items := []resp.CompanyBusinessSummaryItem{
			{OrderNum: 0},
			{OrderNum: 0},
		}
		if businessItemsHaveData(items) {
			t.Error("expected false")
		}
	})

	t.Run("has data", func(t *testing.T) {
		items := []resp.CompanyBusinessSummaryItem{
			{OrderNum: 0},
			{OrderNum: 5},
		}
		if !businessItemsHaveData(items) {
			t.Error("expected true")
		}
	})

	t.Run("empty slice", func(t *testing.T) {
		if businessItemsHaveData(nil) {
			t.Error("expected false for nil")
		}
	})
}

// ==================== joinSortedNames ====================

func TestJoinSortedNames(t *testing.T) {
	t.Run("multiple names sorted", func(t *testing.T) {
		m := map[string]bool{"Charlie": true, "Alice": true, "Bob": true}
		got := joinSortedNames(m)
		if got != "Alice、Bob、Charlie" {
			t.Errorf("got %q", got)
		}
	})

	t.Run("single name", func(t *testing.T) {
		m := map[string]bool{"Only": true}
		got := joinSortedNames(m)
		if got != "Only" {
			t.Errorf("got %q", got)
		}
	})

	t.Run("empty map", func(t *testing.T) {
		m := map[string]bool{}
		got := joinSortedNames(m)
		if got != "" {
			t.Errorf("got %q", got)
		}
	})
}

// ==================== collectAllCompanyNames ====================

func TestCollectAllCompanyNames(t *testing.T) {
	dateMap := map[string]map[string]bool{
		"2024-01-01": {"Store A": true, "Store B": true},
		"2024-01-02": {"Store B": true, "Store C": true},
	}
	got := collectAllCompanyNames(dateMap)
	if got != "Store A、Store B、Store C" {
		t.Errorf("got %q", got)
	}
}

// ==================== addToDateCompanyNames ====================

func TestAddToDateCompanyNames(t *testing.T) {
	t.Run("add to new date", func(t *testing.T) {
		m := make(map[string]map[string]bool)
		addToDateCompanyNames(m, "2024-01-01", "Store A")
		if !m["2024-01-01"]["Store A"] {
			t.Error("expected Store A to be added")
		}
	})

	t.Run("add to existing date", func(t *testing.T) {
		m := map[string]map[string]bool{
			"2024-01-01": {"Store A": true},
		}
		addToDateCompanyNames(m, "2024-01-01", "Store B")
		if len(m["2024-01-01"]) != 2 {
			t.Errorf("expected 2 names, got %d", len(m["2024-01-01"]))
		}
	})
}

// ==================== buildCashStatsByDate ====================

func TestBuildCashStatsByDate(t *testing.T) {
	t.Run("aggregate by date", func(t *testing.T) {
		items := []StatisticsPaymentMethodListResp{
			{Date: "2024-01-01", PaymentNum: 3, PaymentAmount: 100.00},
			{Date: "2024-01-01", PaymentNum: 2, PaymentAmount: 50.00},
			{Date: "2024-01-02", PaymentNum: 5, PaymentAmount: 200.00},
		}
		result := buildCashStatsByDate(items)

		if result["2024-01-01"].CashTC != 5 {
			t.Errorf("expected CashTC 5, got %d", result["2024-01-01"].CashTC)
		}
		if result["2024-01-01"].CashAmount.InexactFloat64() != 150.00 {
			t.Errorf("expected CashAmount 150, got %v", result["2024-01-01"].CashAmount)
		}
		if result["2024-01-02"].CashTC != 5 {
			t.Errorf("expected CashTC 5, got %d", result["2024-01-02"].CashTC)
		}
	})

	t.Run("empty input", func(t *testing.T) {
		result := buildCashStatsByDate(nil)
		if len(result) != 0 {
			t.Errorf("expected empty map, got %d items", len(result))
		}
	})
}

// ==================== buildBusinessItems ====================

func TestBuildBusinessItems(t *testing.T) {
	statList := []StatisticsSummaryItem{
		{
			Date:                                "2024-01-01",
			OrderAmount:                         1000,
			PayAmount:                           900,
			OrderNum:                            10,
			MealNum:                             20,
			DeskNum:                             5,
			DeskOrderAmountEffective:            500,
			InstantOrderAmountEffective:         200,
			TakeoutOrderAmountEffective:         100,
			InstantOrderTakeawayAmountEffective: 50,
			ExternalTakeoutAmountEffective:      50,
			DeskNumEffective:                    4,
			InstantOrderNumEffective:            3,
			TakeoutOrderNumEffective:            2,
			InstantOrderTakeawayNumEffective:    1,
			ExternalTakeoutNumEffective:         1,
			PayAmountMealAvg:                    45,
			OrderAmountMealAvg:                  50,
			OrderAmountAvg:                      100,
			PayAmountAvg:                        90,
			InstantOrderAmount:                  200,
			DeskOrderAmount:                     500,
			TakeoutOrderAmount:                  100,
		},
	}
	cashStats := map[string]cashStatsEntry{
		"2024-01-01": {CashTC: 3, CashAmount: decimal.NewFromFloat(300)},
	}
	ssc := &shopSummaryCtx{
		companyName: "Test Store",
		storeCode:   "001",
		createTime:  1000,
	}

	items := buildBusinessItems(statList, cashStats, ssc)

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	item := items[0]
	if item.CompanyName != "Test Store" {
		t.Errorf("CompanyName: got %q", item.CompanyName)
	}
	if item.StoreCode != "001" {
		t.Errorf("StoreCode: got %q", item.StoreCode)
	}
	if item.CashTC != 3 {
		t.Errorf("CashTC: got %d", item.CashTC)
	}
	if item.CashAmount != 300 {
		t.Errorf("CashAmount: got %f", item.CashAmount)
	}
	if item.CashAC != 100 {
		t.Errorf("CashAC: expected 100, got %f", item.CashAC)
	}
	if item.InStoreAmount != 700 {
		t.Errorf("InStoreAmount: expected 700, got %f", item.InStoreAmount)
	}
	if item.InStoreOrderNum != 7 {
		t.Errorf("InStoreOrderNum: expected 7, got %d", item.InStoreOrderNum)
	}
	if item.DeliveryAmount != 200 {
		t.Errorf("DeliveryAmount: expected 200, got %f", item.DeliveryAmount)
	}
	if item.DeliveryOrderNum != 4 {
		t.Errorf("DeliveryOrderNum: expected 4, got %d", item.DeliveryOrderNum)
	}
}

func TestBuildBusinessItems_ZeroCashTC(t *testing.T) {
	statList := []StatisticsSummaryItem{
		{Date: "2024-01-01", OrderAmount: 100, OrderNum: 1},
	}
	cashStats := map[string]cashStatsEntry{
		"2024-01-01": {CashTC: 0, CashAmount: decimal.Zero},
	}
	ssc := &shopSummaryCtx{companyName: "Test"}

	items := buildBusinessItems(statList, cashStats, ssc)
	if items[0].CashAC != 0 {
		t.Errorf("CashAC should be 0 when CashTC is 0, got %f", items[0].CashAC)
	}
}

// ==================== paginateBusinessItems ====================

func TestPaginateBusinessItems(t *testing.T) {
	items := make([]resp.CompanyBusinessSummaryItem, 25)
	for i := range items {
		items[i] = resp.CompanyBusinessSummaryItem{OrderNum: int64(i)}
	}

	tests := []struct {
		name             string
		pageNo, pageSize int
		expectedLen      int
		firstOrderNum    int64
	}{
		{"page 1", 1, 10, 10, 0},
		{"page 2", 2, 10, 10, 10},
		{"page 3 partial", 3, 10, 5, 20},
		{"page beyond", 4, 10, 0, 0},
		{"page 1 size 25", 1, 25, 25, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := paginateBusinessItems(items, tt.pageNo, tt.pageSize)
			if len(result) != tt.expectedLen {
				t.Errorf("expected %d items, got %d", tt.expectedLen, len(result))
			}
			if tt.expectedLen > 0 && result[0].OrderNum != tt.firstOrderNum {
				t.Errorf("first item OrderNum: expected %d, got %d", tt.firstOrderNum, result[0].OrderNum)
			}
		})
	}
}

// ==================== paginatePaymentMethodItems ====================

func TestPaginatePaymentMethodItems(t *testing.T) {
	items := make([]resp.CompanyPaymentMethodSummaryItem, 5)
	for i := range items {
		items[i] = resp.CompanyPaymentMethodSummaryItem{PaymentNum: int64(i)}
	}

	result := paginatePaymentMethodItems(items, 1, 3)
	if len(result) != 3 {
		t.Errorf("expected 3 items, got %d", len(result))
	}

	result = paginatePaymentMethodItems(items, 2, 3)
	if len(result) != 2 {
		t.Errorf("expected 2 items, got %d", len(result))
	}

	result = paginatePaymentMethodItems(items, 3, 3)
	if len(result) != 0 {
		t.Errorf("expected 0 items, got %d", len(result))
	}
}

// ==================== paginateRefundItems ====================

func TestPaginateRefundItems(t *testing.T) {
	items := make([]resp.CompanyRefundSummaryItem, 5)
	for i := range items {
		items[i] = resp.CompanyRefundSummaryItem{RefundNum: int64(i)}
	}

	result := paginateRefundItems(items, 1, 3)
	if len(result) != 3 {
		t.Errorf("expected 3 items, got %d", len(result))
	}

	result = paginateRefundItems(items, 2, 3)
	if len(result) != 2 {
		t.Errorf("expected 2 items, got %d", len(result))
	}

	result = paginateRefundItems(items, 3, 3)
	if len(result) != 0 {
		t.Errorf("expected 0 items, got %d", len(result))
	}
}

// ==================== sortBusinessDetailItems ====================

func TestSortBusinessDetailItems(t *testing.T) {
	items := []resp.CompanyBusinessSummaryItem{
		{Date: "2024-01-02", StoreCode: "001", CreateTime: 100},
		{Date: "2024-01-01", StoreCode: "002", CreateTime: 200},
		{Date: "2024-01-01", StoreCode: "001", CreateTime: 300},
		{Date: "2024-01-01", StoreCode: "001", CreateTime: 100},
	}
	sortBusinessDetailItems(items)

	// Expected order: 2024-01-01/001/100, 2024-01-01/001/300, 2024-01-01/002/200, 2024-01-02/001/100
	if items[0].Date != "2024-01-01" || items[0].StoreCode != "001" || items[0].CreateTime != 100 {
		t.Errorf("item 0: %+v", items[0])
	}
	if items[1].Date != "2024-01-01" || items[1].StoreCode != "001" || items[1].CreateTime != 300 {
		t.Errorf("item 1: %+v", items[1])
	}
	if items[2].Date != "2024-01-01" || items[2].StoreCode != "002" {
		t.Errorf("item 2: %+v", items[2])
	}
	if items[3].Date != "2024-01-02" {
		t.Errorf("item 3: %+v", items[3])
	}
}

// ==================== sortPaymentMethodDetailItems ====================

func TestSortPaymentMethodDetailItems(t *testing.T) {
	items := []resp.CompanyPaymentMethodSummaryItem{
		{Date: "2024-01-01", StoreCode: "001", CreateTime: 100, PaymentName: "Cash"},
		{Date: "2024-01-01", StoreCode: "001", CreateTime: 100, PaymentName: "Alipay"},
		{Date: "2024-01-02", StoreCode: "001", CreateTime: 100, PaymentName: "Cash"},
	}
	sortPaymentMethodDetailItems(items)

	if items[0].PaymentName != "Alipay" {
		t.Errorf("expected Alipay first, got %q", items[0].PaymentName)
	}
	if items[1].PaymentName != "Cash" {
		t.Errorf("expected Cash second, got %q", items[1].PaymentName)
	}
	if items[2].Date != "2024-01-02" {
		t.Errorf("expected 2024-01-02 last, got %q", items[2].Date)
	}
}

// ==================== sortRefundDetailItems ====================

func TestSortRefundDetailItems(t *testing.T) {
	items := []resp.CompanyRefundSummaryItem{
		{Date: "2024-01-02", StoreCode: "001", CreateTime: 100},
		{Date: "2024-01-01", StoreCode: "", CreateTime: 200},
		{Date: "2024-01-01", StoreCode: "001", CreateTime: 100},
	}
	sortRefundDetailItems(items)

	if items[0].StoreCode != "" {
		t.Errorf("empty store code should come first, got %q", items[0].StoreCode)
	}
	if items[1].StoreCode != "001" || items[1].Date != "2024-01-01" {
		t.Errorf("item 1: %+v", items[1])
	}
	if items[2].Date != "2024-01-02" {
		t.Errorf("item 2: %+v", items[2])
	}
}

// ==================== collectBusinessResults ====================

func TestCollectBusinessResults(t *testing.T) {
	results := []businessShopResult{
		{items: []resp.CompanyBusinessSummaryItem{{OrderNum: 1}}, err: nil},
		{items: nil, err: fmt.Errorf("error")},
		{items: []resp.CompanyBusinessSummaryItem{{OrderNum: 2}, {OrderNum: 3}}, err: nil},
	}
	allItems := collectBusinessResults(results)
	if len(allItems) != 3 {
		t.Errorf("expected 3 items, got %d", len(allItems))
	}
}

// ==================== collectPaymentMethodResults ====================

func TestCollectPaymentMethodResults(t *testing.T) {
	results := []paymentMethodShopResult{
		{items: []resp.CompanyPaymentMethodSummaryItem{{PaymentNum: 1}}, err: nil},
		{items: nil, err: fmt.Errorf("error")},
		{items: []resp.CompanyPaymentMethodSummaryItem{{PaymentNum: 2}}, err: nil},
	}
	allItems := collectPaymentMethodResults(results)
	if len(allItems) != 2 {
		t.Errorf("expected 2 items, got %d", len(allItems))
	}
}

// ==================== collectRefundResults ====================

func TestCollectRefundResults(t *testing.T) {
	results := []refundShopResult{
		{
			items:     []resp.CompanyRefundSummaryItem{{RefundNum: 5}},
			orderNums: map[string]int64{"2024-01-01": 10},
			err:       nil,
		},
		{items: nil, err: fmt.Errorf("error")},
		{
			items:     []resp.CompanyRefundSummaryItem{{RefundNum: 3}},
			orderNums: map[string]int64{"2024-01-01": 20, "2024-01-02": 15},
			err:       nil,
		},
	}
	allItems, orderNums := collectRefundResults(results)

	if len(allItems) != 2 {
		t.Errorf("expected 2 items, got %d", len(allItems))
	}
	if orderNums["2024-01-01"] != 30 {
		t.Errorf("expected 30 orders for 2024-01-01, got %d", orderNums["2024-01-01"])
	}
	if orderNums["2024-01-02"] != 15 {
		t.Errorf("expected 15 orders for 2024-01-02, got %d", orderNums["2024-01-02"])
	}
}

// ==================== accumulateBusinessDateDecimal ====================

func TestAccumulateBusinessDateDecimal(t *testing.T) {
	dateMap := make(map[string]*businessDateSummaryDecimal)

	item1 := resp.CompanyBusinessSummaryItem{
		Date: "2024-01-01", OrderAmount: 100, PayAmount: 90,
		OrderNum: 5, MealNum: 10, DeskNum: 3, CashTC: 2,
		CashAmount: 50, InStoreAmount: 60, DeliveryAmount: 30,
		InStoreOrderNum: 3, DeliveryOrderNum: 2,
		InstantOrderAmount: 20, DeskOrderAmount: 40, TakeoutOrderAmount: 10,
	}
	item2 := resp.CompanyBusinessSummaryItem{
		Date: "2024-01-01", OrderAmount: 200, PayAmount: 180,
		OrderNum: 10, MealNum: 20, DeskNum: 5, CashTC: 3,
		CashAmount: 100, InStoreAmount: 120, DeliveryAmount: 60,
		InStoreOrderNum: 6, DeliveryOrderNum: 4,
		InstantOrderAmount: 40, DeskOrderAmount: 80, TakeoutOrderAmount: 20,
	}

	accumulateBusinessDateDecimal(dateMap, item1)
	accumulateBusinessDateDecimal(dateMap, item2)

	d := dateMap["2024-01-01"]
	if d.OrderNum != 15 {
		t.Errorf("OrderNum: expected 15, got %d", d.OrderNum)
	}
	if d.MealNum != 30 {
		t.Errorf("MealNum: expected 30, got %d", d.MealNum)
	}
	if d.OrderAmount.InexactFloat64() != 300 {
		t.Errorf("OrderAmount: expected 300, got %v", d.OrderAmount)
	}
	if d.CashTC != 5 {
		t.Errorf("CashTC: expected 5, got %d", d.CashTC)
	}
}

// ==================== buildBusinessDateItem ====================

func TestBuildBusinessDateItem(t *testing.T) {
	t.Run("with meal and order", func(t *testing.T) {
		d := &businessDateSummaryDecimal{
			OrderAmount: decimal.NewFromFloat(1000), PayAmount: decimal.NewFromFloat(900),
			OrderNum: 10, MealNum: 20, DeskNum: 5, CashTC: 3,
			CashAmount:      decimal.NewFromFloat(300),
			InStoreAmount:   decimal.NewFromFloat(600),
			DeliveryAmount:  decimal.NewFromFloat(300),
			InStoreOrderNum: 6, DeliveryOrderNum: 4,
			InstantOrderAmount: decimal.NewFromFloat(200),
			DeskOrderAmount:    decimal.NewFromFloat(400),
			TakeoutOrderAmount: decimal.NewFromFloat(100),
		}
		item := buildBusinessDateItem("2024-01-01", d)
		if item.Date != "2024-01-01" {
			t.Errorf("Date: got %q", item.Date)
		}
		if item.AvgCustomerPrice != 45 { // 900/20
			t.Errorf("AvgCustomerPrice: expected 45, got %f", item.AvgCustomerPrice)
		}
		if item.OrderAmountAvg != 100 { // 1000/10
			t.Errorf("OrderAmountAvg: expected 100, got %f", item.OrderAmountAvg)
		}
		if item.CashAC != 100 { // 300/3
			t.Errorf("CashAC: expected 100, got %f", item.CashAC)
		}
	})

	t.Run("zero meal and order", func(t *testing.T) {
		d := &businessDateSummaryDecimal{
			OrderAmount: decimal.NewFromFloat(0), PayAmount: decimal.NewFromFloat(0),
			CashAmount: decimal.Zero, InStoreAmount: decimal.Zero, DeliveryAmount: decimal.Zero,
			InstantOrderAmount: decimal.Zero, DeskOrderAmount: decimal.Zero, TakeoutOrderAmount: decimal.Zero,
		}
		item := buildBusinessDateItem("2024-01-01", d)
		if item.AvgCustomerPrice != 0 {
			t.Errorf("AvgCustomerPrice should be 0, got %f", item.AvgCustomerPrice)
		}
		if item.OrderAmountAvg != 0 {
			t.Errorf("OrderAmountAvg should be 0, got %f", item.OrderAmountAvg)
		}
	})
}

// ==================== aggregateBusinessByDate ====================

func TestAggregateBusinessByDate(t *testing.T) {
	items := []resp.CompanyBusinessSummaryItem{
		{Date: "2024-01-01", CompanyName: "Store A", OrderAmount: 100, PayAmount: 90, OrderNum: 5, MealNum: 10},
		{Date: "2024-01-01", CompanyName: "Store B", OrderAmount: 200, PayAmount: 180, OrderNum: 10, MealNum: 20},
		{Date: "2024-01-02", CompanyName: "Store A", OrderAmount: 150, PayAmount: 140, OrderNum: 8, MealNum: 15},
	}

	finalList, summaryRow := aggregateBusinessByDate(items, "汇总")

	if len(finalList) != 2 {
		t.Fatalf("expected 2 dates, got %d", len(finalList))
	}

	// Check first date (2024-01-01)
	if finalList[0].Date != "2024-01-01" {
		t.Errorf("first date: got %q", finalList[0].Date)
	}
	if finalList[0].OrderNum != 15 {
		t.Errorf("first date OrderNum: expected 15, got %d", finalList[0].OrderNum)
	}

	// Check summary row
	if summaryRow.Date != "汇总" {
		t.Errorf("summary Date: got %q", summaryRow.Date)
	}
	if summaryRow.OrderNum != 23 {
		t.Errorf("summary OrderNum: expected 23, got %d", summaryRow.OrderNum)
	}
}

// ==================== buildBusinessSummaryRow ====================

func TestBuildBusinessSummaryRow(t *testing.T) {
	finalList := []resp.CompanyBusinessSummaryItem{
		{Date: "2024-01-01", OrderAmount: 100, PayAmount: 90, OrderNum: 5, CashTC: 2, CashAmount: 50,
			InStoreAmount: 60, DeliveryAmount: 30, InStoreOrderNum: 3, DeliveryOrderNum: 2,
			InstantOrderAmount: 20, DeskOrderAmount: 40, TakeoutOrderAmount: 10, MealNum: 10, DeskNum: 3},
		{Date: "2024-01-02", OrderAmount: 200, PayAmount: 180, OrderNum: 10, CashTC: 3, CashAmount: 100,
			InStoreAmount: 120, DeliveryAmount: 60, InStoreOrderNum: 6, DeliveryOrderNum: 4,
			InstantOrderAmount: 40, DeskOrderAmount: 80, TakeoutOrderAmount: 20, MealNum: 20, DeskNum: 5},
	}
	dateCompanyNamesMap := map[string]map[string]bool{
		"2024-01-01": {"Store A": true},
		"2024-01-02": {"Store A": true, "Store B": true},
	}

	row := buildBusinessSummaryRow(finalList, dateCompanyNamesMap, "汇总")

	if row.Date != "汇总" {
		t.Errorf("Date: got %q", row.Date)
	}
	if row.OrderNum != 15 {
		t.Errorf("OrderNum: expected 15, got %d", row.OrderNum)
	}
	if row.CompanyName != "Store A、Store B" {
		t.Errorf("CompanyName: got %q", row.CompanyName)
	}
}

// ==================== calculatePaymentRatios ====================

func TestCalculatePaymentRatios(t *testing.T) {
	items := []resp.CompanyPaymentMethodSummaryItem{
		{Date: "2024-01-01", CompanyName: "Store A", PaymentName: "Cash", PaymentAmount: 300},
		{Date: "2024-01-01", CompanyName: "Store A", PaymentName: "Card", PaymentAmount: 700},
		{Date: "2024-01-01", CompanyName: "Store B", PaymentName: "Cash", PaymentAmount: 500},
	}
	calculatePaymentRatios(items)

	// Store A Cash: 300/1000 * 100 = 30%
	if items[0].PaymentRatio != 30 {
		t.Errorf("Store A Cash ratio: expected 30, got %f", items[0].PaymentRatio)
	}
	// Store A Card: 700/1000 * 100 = 70%
	if items[1].PaymentRatio != 70 {
		t.Errorf("Store A Card ratio: expected 70, got %f", items[1].PaymentRatio)
	}
	// Store B Cash: 500/500 * 100 = 100%
	if items[2].PaymentRatio != 100 {
		t.Errorf("Store B Cash ratio: expected 100, got %f", items[2].PaymentRatio)
	}
}

func TestCalculatePaymentRatios_ZeroAmount(t *testing.T) {
	items := []resp.CompanyPaymentMethodSummaryItem{
		{Date: "2024-01-01", CompanyName: "Store A", PaymentName: "Cash", PaymentAmount: 0},
	}
	calculatePaymentRatios(items)
	if items[0].PaymentRatio != 0 {
		t.Errorf("expected 0 ratio for zero amount, got %f", items[0].PaymentRatio)
	}
}

// ==================== accumulatePaymentDateDecimal ====================

func TestAccumulatePaymentDateDecimal(t *testing.T) {
	m := make(map[datePaymentKey]*datePaymentDecimal)
	key := datePaymentKey{Date: "2024-01-01", PaymentName: "Cash"}

	accumulatePaymentDateDecimal(m, key, 100.50, 5)
	accumulatePaymentDateDecimal(m, key, 200.25, 10)

	if m[key].PaymentNum != 15 {
		t.Errorf("expected 15, got %d", m[key].PaymentNum)
	}
	if m[key].PaymentAmount.InexactFloat64() != 300.75 {
		t.Errorf("expected 300.75, got %v", m[key].PaymentAmount)
	}
}

// ==================== accumulatePaymentSummary ====================

func TestAccumulatePaymentSummary(t *testing.T) {
	psMap := make(map[string]*paymentSummaryDecimal)

	item1 := resp.CompanyPaymentMethodSummaryItem{
		PaymentName: "Cash", PaymentAmount: 100, PaymentNum: 5, CompanyName: "Store A、Store B",
	}
	item2 := resp.CompanyPaymentMethodSummaryItem{
		PaymentName: "Cash", PaymentAmount: 200, PaymentNum: 10, CompanyName: "Store B、Store C",
	}

	accumulatePaymentSummary(psMap, item1)
	accumulatePaymentSummary(psMap, item2)

	ps := psMap["Cash"]
	if ps.PaymentNum != 15 {
		t.Errorf("expected 15, got %d", ps.PaymentNum)
	}
	if len(ps.CompanyNames) != 3 {
		t.Errorf("expected 3 unique companies, got %d", len(ps.CompanyNames))
	}
}

// ==================== aggregatePaymentMethodByDate ====================

func TestAggregatePaymentMethodByDate(t *testing.T) {
	items := []resp.CompanyPaymentMethodSummaryItem{
		{Date: "2024-01-01", CompanyName: "Store A", PaymentName: "Cash", PaymentAmount: 100, PaymentNum: 5},
		{Date: "2024-01-01", CompanyName: "Store B", PaymentName: "Cash", PaymentAmount: 200, PaymentNum: 10},
		{Date: "2024-01-01", CompanyName: "Store A", PaymentName: "Card", PaymentAmount: 300, PaymentNum: 8},
	}

	finalList, summaryRow := aggregatePaymentMethodByDate(items, "汇总")

	// Should have 2 entries: Cash and Card for 2024-01-01
	if len(finalList) != 2 {
		t.Fatalf("expected 2 items, got %d", len(finalList))
	}

	if len(summaryRow) != 2 {
		t.Fatalf("expected 2 summary items, got %d", len(summaryRow))
	}
}

// ==================== buildRefundDateItem ====================

func TestBuildRefundDateItem(t *testing.T) {
	t.Run("with refunds and orders", func(t *testing.T) {
		d := &refundDateSummaryDecimal{
			RefundAmount:        decimal.NewFromFloat(100),
			RefundNum:           5,
			PartialRefundAmount: decimal.NewFromFloat(60),
			PartialRefundNum:    3,
			FullRefundAmount:    decimal.NewFromFloat(40),
			FullRefundNum:       2,
			OrderNum:            50,
		}
		item := buildRefundDateItem("2024-01-01", d)
		if item.RefundRate != 10 { // 5/50*100
			t.Errorf("RefundRate: expected 10, got %f", item.RefundRate)
		}
		if item.AvgRefundAmount != 20 { // 100/5
			t.Errorf("AvgRefundAmount: expected 20, got %f", item.AvgRefundAmount)
		}
	})

	t.Run("zero orders and refunds", func(t *testing.T) {
		d := &refundDateSummaryDecimal{
			RefundAmount:        decimal.Zero,
			PartialRefundAmount: decimal.Zero,
			FullRefundAmount:    decimal.Zero,
		}
		item := buildRefundDateItem("2024-01-01", d)
		if item.RefundRate != 0 {
			t.Errorf("RefundRate should be 0, got %f", item.RefundRate)
		}
		if item.AvgRefundAmount != 0 {
			t.Errorf("AvgRefundAmount should be 0, got %f", item.AvgRefundAmount)
		}
	})
}

// ==================== aggregateRefundByDate ====================

func TestAggregateRefundByDate(t *testing.T) {
	items := []resp.CompanyRefundSummaryItem{
		{Date: "2024-01-01", CompanyName: "Store A", RefundAmount: 50, RefundNum: 2, PartialRefundAmount: 30, PartialRefundNum: 1, FullRefundAmount: 20, FullRefundNum: 1},
		{Date: "2024-01-01", CompanyName: "Store B", RefundAmount: 100, RefundNum: 5, PartialRefundAmount: 60, PartialRefundNum: 3, FullRefundAmount: 40, FullRefundNum: 2},
		{Date: "2024-01-02", CompanyName: "Store A", RefundAmount: 30, RefundNum: 1, PartialRefundAmount: 30, PartialRefundNum: 1},
	}
	dateOrderNums := map[string]int64{"2024-01-01": 100, "2024-01-02": 50}

	finalList, summaryRow := aggregateRefundByDate(items, dateOrderNums, "汇总")

	if len(finalList) != 2 {
		t.Fatalf("expected 2 dates, got %d", len(finalList))
	}

	// 2024-01-01: refund=150, num=7
	if finalList[0].RefundNum != 7 {
		t.Errorf("2024-01-01 RefundNum: expected 7, got %d", finalList[0].RefundNum)
	}

	// Summary: total refund=180, total num=8
	if summaryRow.RefundNum != 8 {
		t.Errorf("summary RefundNum: expected 8, got %d", summaryRow.RefundNum)
	}
	if summaryRow.Date != "汇总" {
		t.Errorf("summary Date: got %q", summaryRow.Date)
	}
}

// ==================== buildRefundSummaryRow ====================

func TestBuildRefundSummaryRow(t *testing.T) {
	finalList := []resp.CompanyRefundSummaryItem{
		{Date: "2024-01-01", RefundAmount: 100, RefundNum: 5, PartialRefundAmount: 60, PartialRefundNum: 3, FullRefundAmount: 40, FullRefundNum: 2},
		{Date: "2024-01-02", RefundAmount: 50, RefundNum: 2, PartialRefundAmount: 30, PartialRefundNum: 1, FullRefundAmount: 20, FullRefundNum: 1},
	}
	dateOrderNums := map[string]int64{"2024-01-01": 100, "2024-01-02": 50}
	dateCompanyNamesMap := map[string]map[string]bool{
		"2024-01-01": {"Store A": true},
		"2024-01-02": {"Store A": true, "Store B": true},
	}

	row := buildRefundSummaryRow(finalList, dateOrderNums, dateCompanyNamesMap, "汇总")

	if row.RefundNum != 7 {
		t.Errorf("RefundNum: expected 7, got %d", row.RefundNum)
	}
	if row.CompanyName != "Store A、Store B" {
		t.Errorf("CompanyName: got %q", row.CompanyName)
	}
}

// ==================== buildPaymentMethodSummaryRow ====================

func TestBuildPaymentMethodSummaryRow(t *testing.T) {
	finalList := []resp.CompanyPaymentMethodSummaryItem{
		{Date: "2024-01-01", CompanyName: "Store A", PaymentName: "Cash", PaymentAmount: 100, PaymentNum: 5},
		{Date: "2024-01-01", CompanyName: "Store B", PaymentName: "Cash", PaymentAmount: 200, PaymentNum: 10},
		{Date: "2024-01-01", CompanyName: "Store A", PaymentName: "Card", PaymentAmount: 300, PaymentNum: 8},
	}

	summaryRow := buildPaymentMethodSummaryRow(finalList, "汇总")

	if len(summaryRow) != 2 {
		t.Fatalf("expected 2 summary items, got %d", len(summaryRow))
	}

	// Find Cash summary
	var cashSummary, cardSummary *resp.CompanyPaymentMethodSummaryItem
	for i := range summaryRow {
		if summaryRow[i].PaymentName == "Cash" {
			cashSummary = &summaryRow[i]
		}
		if summaryRow[i].PaymentName == "Card" {
			cardSummary = &summaryRow[i]
		}
	}

	if cashSummary == nil || cardSummary == nil {
		t.Fatal("missing Cash or Card summary")
	}

	if cashSummary.PaymentNum != 15 {
		t.Errorf("Cash PaymentNum: expected 15, got %d", cashSummary.PaymentNum)
	}
	if cashSummary.Date != "汇总" {
		t.Errorf("Cash Date: got %q", cashSummary.Date)
	}
}

// ==================== buildRefundDateList ====================

func TestBuildRefundDateList(t *testing.T) {
	dateDecimalMap := map[string]*refundDateSummaryDecimal{
		"2024-01-02": {RefundAmount: decimal.NewFromFloat(50), RefundNum: 2, PartialRefundAmount: decimal.NewFromFloat(30), PartialRefundNum: 1, FullRefundAmount: decimal.NewFromFloat(20), FullRefundNum: 1, OrderNum: 50},
		"2024-01-01": {RefundAmount: decimal.NewFromFloat(100), RefundNum: 5, PartialRefundAmount: decimal.NewFromFloat(60), PartialRefundNum: 3, FullRefundAmount: decimal.NewFromFloat(40), FullRefundNum: 2, OrderNum: 100},
	}
	dateCompanyNamesMap := map[string]map[string]bool{
		"2024-01-01": {"Store A": true},
		"2024-01-02": {"Store B": true},
	}

	list := buildRefundDateList(dateDecimalMap, dateCompanyNamesMap)

	if len(list) != 2 {
		t.Fatalf("expected 2, got %d", len(list))
	}
	// Should be sorted by date
	if list[0].Date != "2024-01-01" {
		t.Errorf("first date: got %q", list[0].Date)
	}
	if list[0].CompanyName != "Store A" {
		t.Errorf("first CompanyName: got %q", list[0].CompanyName)
	}
}

// ==================== buildBusinessDateList ====================

func TestBuildBusinessDateList(t *testing.T) {
	dateDecimalMap := map[string]*businessDateSummaryDecimal{
		"2024-01-02": {
			OrderAmount: decimal.NewFromFloat(200), PayAmount: decimal.NewFromFloat(180),
			OrderNum: 10, MealNum: 20, CashAmount: decimal.Zero, InStoreAmount: decimal.Zero,
			DeliveryAmount: decimal.Zero, InstantOrderAmount: decimal.Zero,
			DeskOrderAmount: decimal.Zero, TakeoutOrderAmount: decimal.Zero,
		},
		"2024-01-01": {
			OrderAmount: decimal.NewFromFloat(100), PayAmount: decimal.NewFromFloat(90),
			OrderNum: 5, MealNum: 10, CashAmount: decimal.Zero, InStoreAmount: decimal.Zero,
			DeliveryAmount: decimal.Zero, InstantOrderAmount: decimal.Zero,
			DeskOrderAmount: decimal.Zero, TakeoutOrderAmount: decimal.Zero,
		},
	}
	dateCompanyNamesMap := map[string]map[string]bool{
		"2024-01-01": {"Store A": true},
		"2024-01-02": {"Store B": true},
	}

	list := buildBusinessDateList(dateDecimalMap, dateCompanyNamesMap)

	if len(list) != 2 {
		t.Fatalf("expected 2, got %d", len(list))
	}
	if list[0].Date != "2024-01-01" {
		t.Errorf("first date: got %q", list[0].Date)
	}
	if list[0].CompanyName != "Store A" {
		t.Errorf("first CompanyName: got %q", list[0].CompanyName)
	}
	if list[1].OrderNum != 10 {
		t.Errorf("second OrderNum: expected 10, got %d", list[1].OrderNum)
	}
}

// ==================== accumulateRefundDateDecimal ====================

func TestAccumulateRefundDateDecimal(t *testing.T) {
	dateMap := make(map[string]*refundDateSummaryDecimal)
	dateOrderNumMap := map[string]int64{"2024-01-01": 100}

	item1 := resp.CompanyRefundSummaryItem{
		Date: "2024-01-01", RefundAmount: 50, RefundNum: 2,
		PartialRefundAmount: 30, PartialRefundNum: 1,
		FullRefundAmount: 20, FullRefundNum: 1,
	}
	item2 := resp.CompanyRefundSummaryItem{
		Date: "2024-01-01", RefundAmount: 100, RefundNum: 5,
		PartialRefundAmount: 60, PartialRefundNum: 3,
		FullRefundAmount: 40, FullRefundNum: 2,
	}

	accumulateRefundDateDecimal(dateMap, dateOrderNumMap, item1)
	accumulateRefundDateDecimal(dateMap, dateOrderNumMap, item2)

	d := dateMap["2024-01-01"]
	if d.RefundNum != 7 {
		t.Errorf("RefundNum: expected 7, got %d", d.RefundNum)
	}
	if d.PartialRefundNum != 4 {
		t.Errorf("PartialRefundNum: expected 4, got %d", d.PartialRefundNum)
	}
	if d.OrderNum != 100 {
		t.Errorf("OrderNum: expected 100, got %d", d.OrderNum)
	}
}

// ==================== buildPaymentMethodDateList ====================

func TestBuildPaymentMethodDateList(t *testing.T) {
	datePaymentDecimalMap := map[datePaymentKey]*datePaymentDecimal{
		{Date: "2024-01-01", PaymentName: "Cash"}: {PaymentAmount: decimal.NewFromFloat(300), PaymentNum: 10},
		{Date: "2024-01-01", PaymentName: "Card"}: {PaymentAmount: decimal.NewFromFloat(700), PaymentNum: 20},
	}
	datePaymentCompanyNamesMap := map[datePaymentKey]map[string]bool{
		{Date: "2024-01-01", PaymentName: "Cash"}: {"Store A": true},
		{Date: "2024-01-01", PaymentName: "Card"}: {"Store B": true},
	}

	list := buildPaymentMethodDateList(datePaymentDecimalMap, datePaymentCompanyNamesMap)

	if len(list) != 2 {
		t.Fatalf("expected 2, got %d", len(list))
	}

	// Check ratios - Cash: 300/1000*100=30%, Card: 700/1000*100=70%
	var cashItem, cardItem *resp.CompanyPaymentMethodSummaryItem
	for i := range list {
		if list[i].PaymentName == "Cash" {
			cashItem = &list[i]
		}
		if list[i].PaymentName == "Card" {
			cardItem = &list[i]
		}
	}
	if cashItem == nil || cardItem == nil {
		t.Fatal("missing Cash or Card item")
	}
	if cashItem.PaymentRatio != 30 {
		t.Errorf("Cash ratio: expected 30, got %f", cashItem.PaymentRatio)
	}
	if cardItem.PaymentRatio != 70 {
		t.Errorf("Card ratio: expected 70, got %f", cardItem.PaymentRatio)
	}
}
