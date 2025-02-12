package model

import "testing"

func TestSaleOrderProduct_GenerateProductSign(t *testing.T) {
	productPackage := &SaleOrderProduct{
		Uuid: 1,
		SaleOrderProductBoms: []SaleOrderProductBom{
			{
				Uuid:           1,
				ProductBomUuid: 1,
			},
			{
				Uuid:           2,
				ProductBomUuid: 2,
			},
		},
		SaleOrderProductAttributes: []SaleOrderProductAttribute{
			{
				Uuid:                 1,
				ProductAttributeUuid: 31,
			},
			{
				Uuid:                 2,
				ProductAttributeUuid: 21,
			},
		},
	}
	sign := productPackage.GenerateProductSign()
	t.Log(sign)
}
