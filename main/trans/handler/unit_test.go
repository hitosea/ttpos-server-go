//go:build manual

package handler

import "testing"

func TestAll(t *testing.T) {
	if err := testConvertProduct(); err != nil {
		t.Errorf("testConvertProduct failed: %v", err)
	}
	if err := testConvertProductSauce(); err != nil {
		t.Errorf("testConvertProductSauce failed: %v", err)
	}
}

func TestTransProduct(t *testing.T) {
	if err := testConvertProduct(); err != nil {
		t.Errorf("testConvertProduct failed: %v", err)
	}
}

func TestTransProductSauce(t *testing.T) {
	if err := testConvertProductSauce(); err != nil {
		t.Errorf("testConvertProductSauce failed: %v", err)
	}
}
