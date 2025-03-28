package handler

import "testing"

func TestTransProduct(t *testing.T) {
	if err := testConvertProduct(); err != nil {
		t.Errorf("testConvertProduct failed: %v", err)
	}
}
