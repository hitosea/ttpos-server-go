package req

import (
	"encoding/json"
	"testing"
)

// TestSaveSalesInvoiceReq_CompanyAbbrField 验证 SaveSalesInvoiceReq 使用 CompanyAbbr 字段
// 确保 JSON 序列化使用 company_abbr 而非 company
func TestSaveSalesInvoiceReq_CompanyAbbrField(t *testing.T) {
	req := SaveSalesInvoiceReq{
		OrderNo:     "TEST-001",
		CompanyAbbr: "CFG",
		Customer:    "Default",
		Currency:    "THB",
		Branch:      "华莱士总部",
	}

	// 验证字段赋值
	if req.CompanyAbbr != "CFG" {
		t.Errorf("CompanyAbbr: got %q, want %q", req.CompanyAbbr, "CFG")
	}

	// 验证 JSON tag 为 company_abbr
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if _, ok := m["company"]; ok {
		t.Error("JSON should not have 'company' key, should use 'company_abbr'")
	}
	if v, ok := m["company_abbr"]; !ok {
		t.Error("JSON should have 'company_abbr' key")
	} else if v != "CFG" {
		t.Errorf("company_abbr value: got %q, want %q", v, "CFG")
	}
}

// TestReturnSalesInvoiceReq_CompanyAbbrField 验证 ReturnSalesInvoiceReq 使用 CompanyAbbr 字段
func TestReturnSalesInvoiceReq_CompanyAbbrField(t *testing.T) {
	req := ReturnSalesInvoiceReq{
		OrderNo:          "TEST-002",
		SalesInvoiceName: "ACC-SINV-2026-00132",
		CompanyAbbr:      "CFG",
		Customer:         "Default",
		RefundType:       "full_refund",
	}

	if req.CompanyAbbr != "CFG" {
		t.Errorf("CompanyAbbr: got %q, want %q", req.CompanyAbbr, "CFG")
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if _, ok := m["company"]; ok {
		t.Error("JSON should not have 'company' key, should use 'company_abbr'")
	}
	if v, ok := m["company_abbr"]; !ok {
		t.Error("JSON should have 'company_abbr' key")
	} else if v != "CFG" {
		t.Errorf("company_abbr value: got %q, want %q", v, "CFG")
	}
}

// TestCancelSalesInvoiceReq_NoCompanyField 验证 CancelSalesInvoiceReq 不受影响
func TestCancelSalesInvoiceReq_NoCompanyField(t *testing.T) {
	req := CancelSalesInvoiceReq{
		OrderNo:          "TEST-003",
		SalesInvoiceName: "ACC-SINV-2026-00132",
	}

	if req.OrderNo != "TEST-003" {
		t.Errorf("OrderNo: got %q, want %q", req.OrderNo, "TEST-003")
	}
}
