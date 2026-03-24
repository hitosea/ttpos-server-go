package selling

import (
	"context"
	"testing"
	"ttpos-bmp/app/ttpos-erp/api/selling"
)

func Test_buildSalesInvoice_FreeItemDiscountPercentage(t *testing.T) {
	s := &sSelling{}
	req := &selling.SaveSalesInvoiceReq{
		CompanyAbbr:       "wfs-002",
		Currency:          "THB",
		PriceListCurrency: "THB",
		OrderNo:           "TEST001",
		Items: []*selling.PosInvoiceItem{
			{ItemCode: "TC001", Qty: 1, Rate: 510, Amount: 510, IsFreeItem: false},
			{ItemCode: "SP001", Qty: 1, Rate: 0, Amount: 0, IsFreeItem: true, Description: "Sales in package:Pork"},
			{ItemCode: "SP002", Qty: 2, Rate: 0, Amount: 0, IsFreeItem: true, Description: "Sales in package:Mushroom"},
		},
		Taxes: []*selling.PosInvoiceTax{
			{TaxAmount: 51, Description: "Tax"},
		},
	}

	si := s.buildSalesInvoice(context.Background(), req, "wfs-002", "2026-03-24", "12:00:00")

	// 验证普通商品
	if si.Items[0].Rate != 510 {
		t.Errorf("普通商品 rate = %v, want 510", si.Items[0].Rate)
	}
	if si.Items[0].DiscountPercentage != 0 {
		t.Errorf("普通商品 discount_percentage = %v, want 0", si.Items[0].DiscountPercentage)
	}
	if si.Items[0].IsFreeItem != 0 {
		t.Errorf("普通商品 is_free_item = %v, want 0", si.Items[0].IsFreeItem)
	}

	// 验证 free item 1
	if si.Items[1].IsFreeItem != 1 {
		t.Errorf("free item 1 is_free_item = %v, want 1", si.Items[1].IsFreeItem)
	}
	if si.Items[1].DiscountPercentage != 100 {
		t.Errorf("free item 1 discount_percentage = %v, want 100", si.Items[1].DiscountPercentage)
	}
	if si.Items[1].Rate != 0 {
		t.Errorf("free item 1 rate = %v, want 0", si.Items[1].Rate)
	}

	// 验证 free item 2
	if si.Items[2].IsFreeItem != 1 {
		t.Errorf("free item 2 is_free_item = %v, want 1", si.Items[2].IsFreeItem)
	}
	if si.Items[2].DiscountPercentage != 100 {
		t.Errorf("free item 2 discount_percentage = %v, want 100", si.Items[2].DiscountPercentage)
	}
}
