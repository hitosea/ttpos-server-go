package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseStockEntryErrorItemCodes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		errMsg   string
		expected []string
	}{
		{
			name:     "empty message",
			errMsg:   "",
			expected: nil,
		},
		{
			name:     "insufficient stock - Item Code strong tag",
			errMsg:   `Item Code: <strong>ITEM-001</strong> in warehouse`,
			expected: []string{"ITEM-001"},
		},
		{
			name:     "multiple insufficient stock items",
			errMsg:   `Item Code: <strong>ITEM-001</strong> and Item Code: <strong>ITEM-002</strong>`,
			expected: []string{"ITEM-001", "ITEM-002"},
		},
		{
			name:     "duplicate item codes deduplicated",
			errMsg:   `Item Code: <strong>ITEM-001</strong> foo Item Code: <strong>ITEM-001</strong>`,
			expected: []string{"ITEM-001"},
		},
		{
			name:     "NegativeStockError format",
			errMsg:   `<a href="/app/item/ITEM-003">Item ITEM-003: Some Name</a> needed in warehouse`,
			expected: []string{"ITEM-003"},
		},
		{
			name:     "NegativeStockError without colon",
			errMsg:   `<a href="/app/item/ITEM-004">Item ITEM-004</a> needed in warehouse`,
			expected: []string{"ITEM-004"},
		},
		{
			name:     "not a stock item",
			errMsg:   `ITEM-005 is not a stock Item`,
			expected: []string{"ITEM-005"},
		},
		{
			name:     "item disabled",
			errMsg:   `Item ITEM-006 is disabled`,
			expected: []string{"ITEM-006"},
		},
		{
			name:     "item not active",
			errMsg:   `Item ITEM-007 is not active or end of life has been reached`,
			expected: []string{"ITEM-007"},
		},
		{
			name:     "valuation rate missing",
			errMsg:   `Valuation Rate for the Item <a href="/app/item/ITEM-008">ITEM-008</a>`,
			expected: []string{"ITEM-008"},
		},
		{
			name:     "BMP wrapped error - strips prefix",
			errMsg:   `调用erp接口返回异常:ValidationError;Item Code: <strong>RAW-100</strong> insufficient`,
			expected: []string{"RAW-100"},
		},
		{
			name:     "BMP wrapped NegativeStockError",
			errMsg:   `调用erp接口返回异常:NegativeStockError;<a href="/app/item/RAW-200">Item RAW-200: Raw Material</a> needed in Warehouse`,
			expected: []string{"RAW-200"},
		},
		{
			name:     "mixed error types",
			errMsg:   `Item Code: <strong>A1</strong> and ITEM-B2 is not a stock Item and Item ITEM-C3 is disabled`,
			expected: []string{"A1", "ITEM-B2", "ITEM-C3"},
		},
		{
			name:     "no matching patterns",
			errMsg:   `some random error message without item codes`,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := parseStockEntryErrorItemCodes(tt.errMsg)
			assert.Equal(t, tt.expected, result)
		})
	}
}
