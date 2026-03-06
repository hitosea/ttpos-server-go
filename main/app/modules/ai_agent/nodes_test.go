package ai_agent

import (
	"testing"
)

// --- compareStock tests ---

func TestCompareStock_EmptyForecasts(t *testing.T) {
	state := &ProcurementState{
		Forecasts: nil,
		StepLog:   make([]string, 0),
	}
	compareStock(state)

	if state.NeedsPurchase {
		t.Error("expected NeedsPurchase=false with no forecasts")
	}
	if len(state.StepLog) == 0 {
		t.Error("expected step log entry")
	}
}

func TestCompareStock_ValidatesAndRecalculates(t *testing.T) {
	state := &ProcurementState{
		Materials: []MaterialInfo{
			{MaterialUuid: 1, BookedQuantity: 10, SafetyStock: 5},
			{MaterialUuid: 2, BookedQuantity: 100, SafetyStock: 10},
			{MaterialUuid: 3, BookedQuantity: 0, SafetyStock: 0},
		},
		Forecasts: []ForecastItem{
			{MaterialUuid: 1, PredictedDemand: 20},  // shortage: 20 - 10 + 5 = 15
			{MaterialUuid: 2, PredictedDemand: 50},  // shortage: 50 - 100 + 10 = -40 (skip)
			{MaterialUuid: 3, PredictedDemand: 10},  // shortage: 10 - 0 + 2 (20% of 10) = 12
			{MaterialUuid: 999, PredictedDemand: 5}, // unknown material (skip)
		},
		StepLog: make([]string, 0),
	}
	compareStock(state)

	if !state.NeedsPurchase {
		t.Error("expected NeedsPurchase=true")
	}
	if len(state.Forecasts) != 2 {
		t.Fatalf("expected 2 validated forecasts, got %d", len(state.Forecasts))
	}

	// Material 1: shortage = 20 - 10 + 5 = 15
	if state.Forecasts[0].MaterialUuid != 1 {
		t.Errorf("expected first forecast material_uuid=1, got %d", state.Forecasts[0].MaterialUuid)
	}
	if state.Forecasts[0].Shortage != 15 {
		t.Errorf("expected shortage=15, got %.1f", state.Forecasts[0].Shortage)
	}
	if state.Forecasts[0].CurrentStock != 10 {
		t.Errorf("expected current_stock=10, got %.1f", state.Forecasts[0].CurrentStock)
	}

	// Material 3: shortage = 10 - 0 + 2 = 12 (safety_stock=0, buffer=20% of demand)
	if state.Forecasts[1].MaterialUuid != 3 {
		t.Errorf("expected second forecast material_uuid=3, got %d", state.Forecasts[1].MaterialUuid)
	}
	if state.Forecasts[1].Shortage != 12 {
		t.Errorf("expected shortage=12, got %.1f", state.Forecasts[1].Shortage)
	}
}

func TestCompareStock_AllSufficient(t *testing.T) {
	state := &ProcurementState{
		Materials: []MaterialInfo{
			{MaterialUuid: 1, BookedQuantity: 100, SafetyStock: 5},
		},
		Forecasts: []ForecastItem{
			{MaterialUuid: 1, PredictedDemand: 10}, // shortage: 10 - 100 + 5 = -85 (skip)
		},
		StepLog: make([]string, 0),
	}
	compareStock(state)

	if state.NeedsPurchase {
		t.Error("expected NeedsPurchase=false when stock is sufficient")
	}
	if len(state.Forecasts) != 0 {
		t.Errorf("expected 0 forecasts, got %d", len(state.Forecasts))
	}
}

// --- matchSupplier tests ---

func TestMatchSupplier_MatchesKnownSuppliers(t *testing.T) {
	state := &ProcurementState{
		Materials: []MaterialInfo{
			{MaterialUuid: 1, SupplierUuid: 100},
			{MaterialUuid: 2, SupplierUuid: 200},
			{MaterialUuid: 3, SupplierUuid: 0}, // no supplier
		},
		Suppliers: []SupplierInfo{
			{Uuid: 100, Name: "供应商A", ErpCode: "SA001"},
			{Uuid: 200, Name: "供应商B", ErpCode: "SB001"},
		},
		Forecasts: []ForecastItem{
			{MaterialUuid: 1},
			{MaterialUuid: 2},
			{MaterialUuid: 3},
		},
		StepLog: make([]string, 0),
	}
	matchSupplier(state)

	if state.Forecasts[0].SupplierName != "供应商A" {
		t.Errorf("expected supplier A, got %s", state.Forecasts[0].SupplierName)
	}
	if state.Forecasts[1].SupplierName != "供应商B" {
		t.Errorf("expected supplier B, got %s", state.Forecasts[1].SupplierName)
	}
	// Material 3 has no supplier, should get default (first supplier)
	if state.Forecasts[2].SupplierName != "供应商A" {
		t.Errorf("expected default supplier A, got %s", state.Forecasts[2].SupplierName)
	}
}

func TestMatchSupplier_NoSuppliers(t *testing.T) {
	state := &ProcurementState{
		Materials: []MaterialInfo{
			{MaterialUuid: 1, SupplierUuid: 100},
		},
		Suppliers: nil,
		Forecasts: []ForecastItem{
			{MaterialUuid: 1},
		},
		StepLog: make([]string, 0),
	}
	matchSupplier(state)

	if state.Forecasts[0].SupplierName != "" {
		t.Errorf("expected empty supplier name, got %s", state.Forecasts[0].SupplierName)
	}
}

// --- generateProposal tests ---

func TestGenerateProposal_GroupsBySupplier(t *testing.T) {
	state := &ProcurementState{
		Forecasts: []ForecastItem{
			{MaterialUuid: 1, MaterialCode: "M001", SupplierName: "供应商A", SupplierErpCode: "SA", OrderQuantity: 10},
			{MaterialUuid: 2, MaterialCode: "M002", SupplierName: "供应商A", SupplierErpCode: "SA", OrderQuantity: 20},
			{MaterialUuid: 3, MaterialCode: "M003", SupplierName: "供应商B", SupplierErpCode: "SB", OrderQuantity: 5},
		},
		StepLog: make([]string, 0),
	}
	generateProposal(state)

	if len(state.Proposals) != 2 {
		t.Fatalf("expected 2 proposals, got %d", len(state.Proposals))
	}

	// Find proposal for supplier A
	var propA, propB *PurchaseProposal
	for i := range state.Proposals {
		if state.Proposals[i].SupplierName == "供应商A" {
			propA = &state.Proposals[i]
		} else if state.Proposals[i].SupplierName == "供应商B" {
			propB = &state.Proposals[i]
		}
	}

	if propA == nil || propB == nil {
		t.Fatal("missing expected supplier proposals")
	}
	if len(propA.Items) != 2 {
		t.Errorf("expected 2 items for supplier A, got %d", len(propA.Items))
	}
	if propA.TotalQuantity != 30 {
		t.Errorf("expected total 30 for supplier A, got %.1f", propA.TotalQuantity)
	}
	if len(propB.Items) != 1 {
		t.Errorf("expected 1 item for supplier B, got %d", len(propB.Items))
	}
	if propB.TotalQuantity != 5 {
		t.Errorf("expected total 5 for supplier B, got %.1f", propB.TotalQuantity)
	}
}

func TestGenerateProposal_EmptyForecasts(t *testing.T) {
	state := &ProcurementState{
		Forecasts: nil,
		StepLog:   make([]string, 0),
	}
	generateProposal(state)

	if state.Proposals == nil {
		t.Error("expected non-nil empty proposals slice")
	}
	if len(state.Proposals) != 0 {
		t.Errorf("expected 0 proposals, got %d", len(state.Proposals))
	}
}

func TestGenerateProposal_UnknownSupplier(t *testing.T) {
	state := &ProcurementState{
		Forecasts: []ForecastItem{
			{MaterialUuid: 1, SupplierName: "", OrderQuantity: 10},
		},
		StepLog: make([]string, 0),
	}
	generateProposal(state)

	if len(state.Proposals) != 1 {
		t.Fatalf("expected 1 proposal, got %d", len(state.Proposals))
	}
}

// --- humanReview tests ---

func TestHumanReview_SetsAwaitingReview(t *testing.T) {
	state := &ProcurementState{
		Status:  "running",
		StepLog: make([]string, 0),
	}
	humanReview(state)

	if state.Status != "awaiting_review" {
		t.Errorf("expected status=awaiting_review, got %s", state.Status)
	}
}

// --- detectAnomalies tests ---

func TestDetectAnomalies_ZeroStock(t *testing.T) {
	state := &ProcurementState{
		Materials: []MaterialInfo{
			{MaterialUuid: 1, MaterialNameZH: "鸡蛋", BookedQuantity: 0, SafetyStock: 10},
		},
		StepLog: make([]string, 0),
	}
	detectAnomalies(state)

	if len(state.Anomalies) != 1 {
		t.Fatalf("expected 1 anomaly, got %d", len(state.Anomalies))
	}
	if state.Anomalies[0].AnomalyType != "zero_stock" {
		t.Errorf("expected zero_stock, got %s", state.Anomalies[0].AnomalyType)
	}
	if state.Anomalies[0].Severity != "high" {
		t.Errorf("expected high severity, got %s", state.Anomalies[0].Severity)
	}
}

func TestDetectAnomalies_BelowSafetyStock(t *testing.T) {
	state := &ProcurementState{
		Materials: []MaterialInfo{
			{MaterialUuid: 1, MaterialNameZH: "面粉", BookedQuantity: 3, SafetyStock: 10},
		},
		StepLog: make([]string, 0),
	}
	detectAnomalies(state)

	if len(state.Anomalies) != 1 {
		t.Fatalf("expected 1 anomaly, got %d", len(state.Anomalies))
	}
	if state.Anomalies[0].AnomalyType != "below_safety_stock" {
		t.Errorf("expected below_safety_stock, got %s", state.Anomalies[0].AnomalyType)
	}
	if state.Anomalies[0].Severity != "medium" {
		t.Errorf("expected medium severity, got %s", state.Anomalies[0].Severity)
	}
}

func TestDetectAnomalies_NoAnomalies(t *testing.T) {
	state := &ProcurementState{
		Materials: []MaterialInfo{
			{MaterialUuid: 1, BookedQuantity: 50, SafetyStock: 10},   // above safety
			{MaterialUuid: 2, BookedQuantity: 0, SafetyStock: 0},     // zero stock but no safety set
			{MaterialUuid: 3, BookedQuantity: 100, SafetyStock: 100}, // exactly at safety
		},
		StepLog: make([]string, 0),
	}
	detectAnomalies(state)

	if len(state.Anomalies) != 0 {
		t.Errorf("expected 0 anomalies, got %d", len(state.Anomalies))
	}
}

func TestDetectAnomalies_MixedAnomalies(t *testing.T) {
	state := &ProcurementState{
		Materials: []MaterialInfo{
			{MaterialUuid: 1, MaterialNameZH: "鸡蛋", BookedQuantity: 0, SafetyStock: 10},   // zero_stock
			{MaterialUuid: 2, MaterialNameZH: "面粉", BookedQuantity: 5, SafetyStock: 20},   // below_safety
			{MaterialUuid: 3, MaterialNameZH: "牛奶", BookedQuantity: 100, SafetyStock: 50}, // ok
		},
		StepLog: make([]string, 0),
	}
	detectAnomalies(state)

	if len(state.Anomalies) != 2 {
		t.Fatalf("expected 2 anomalies, got %d", len(state.Anomalies))
	}
}
