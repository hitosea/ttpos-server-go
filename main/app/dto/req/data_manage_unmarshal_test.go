package req

import (
	"encoding/json"
	"testing"
)

func TestGetDataManageOrderSelectStatsReq_Unmarshal_Filters(t *testing.T) {
	payload := []byte(`{"filters":[{"select_all":false,"selected_uuids":[1,2],"date_type":3,"query_start_date":"2026-03-01","query_end_date":"2026-03-31","bill_type":-1},{"select_all":false,"selected_uuids":[3,4],"date_type":3,"query_start_date":"2026-02-01","query_end_date":"2026-02-28","bill_type":-1}]}`)
	var r GetDataManageOrderSelectStatsReq
	if err := json.Unmarshal(payload, &r); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if len(r.Filters) != 2 {
		t.Fatalf("expected 2 filters, got %d", len(r.Filters))
	}
	if len(r.Filters[0].SelectedUuids) != 2 {
		t.Fatalf("expected 2 selected_uuids in filter[0], got %d", len(r.Filters[0].SelectedUuids))
	}
	if len(r.Filters[1].SelectedUuids) != 2 {
		t.Fatalf("expected 2 selected_uuids in filter[1], got %d", len(r.Filters[1].SelectedUuids))
	}
}

func TestGetDataManageOrderSelectStatsReq_Unmarshal_EmptyFilters(t *testing.T) {
	payload := []byte(`{}`)
	var r GetDataManageOrderSelectStatsReq
	if err := json.Unmarshal(payload, &r); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if len(r.Filters) != 0 {
		t.Fatalf("expected 0 filters, got %d", len(r.Filters))
	}
}
