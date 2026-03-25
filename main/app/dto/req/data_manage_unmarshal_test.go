package req

import (
	"encoding/json"
	"testing"
)

func TestGetDataManageOrderSelectStatsReq_Unmarshal_WrappedFilter(t *testing.T) {
	payload := []byte(`{"filter":{"select_all":false,"date_type":-1,"bill_type":-1}}`)
	var r GetDataManageOrderSelectStatsReq
	if err := json.Unmarshal(payload, &r); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if r.Filter == nil {
		t.Fatal("expected filter not nil")
	}
	if r.Filter.DateType != -1 {
		t.Fatalf("expected date_type=-1, got %d", r.Filter.DateType)
	}
}

func TestGetDataManageOrderSelectStatsReq_Unmarshal_LegacyRootFilter(t *testing.T) {
	payload := []byte(`{"select_all":false,"date_type":-1,"bill_type":-1}`)
	var r GetDataManageOrderSelectStatsReq
	if err := json.Unmarshal(payload, &r); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if r.Filter == nil {
		t.Fatal("expected filter not nil")
	}
	if r.Filter.BillType != -1 {
		t.Fatalf("expected bill_type=-1, got %d", r.Filter.BillType)
	}
}
