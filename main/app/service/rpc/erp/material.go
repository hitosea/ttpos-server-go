package erp

import (
	"ttpos-bmp/app/ttpos-erp/api/item"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/pkg/context"
)

type AddMaterialParams struct {
	LocaleName   dto.LocaleResponse `json:"locale_name" binding:"required"`      // 物品名称
	CategoryUuid uint64             `json:"category_uuid" binding:"required"`    // 分类UUID
	Status       int                `json:"status" binding:"required"`           // 状态，1-启用 2-停用
	Valuation    float64            `json:"valuation" binding:"required,min=0"`  // 估值率
	InitStock    float64            `json:"init_stock" binding:"required,min=0"` // 期初库存
}

func (s *erpSrv) AddMaterial(ctx context.Context, params req.MaterialAddErpReq) (*item.ItemInfo, error) {
	company := ctx.GetCompany()
	companySetting := company.CompanySetting

	client, conn, err := NewErpItemClient()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	unitList := make([]*item.UomDetail, 0)
	for _, uom := range params.Uoms {
		unitList = append(unitList, &item.UomDetail{
			Uom:              uom.Uom,
			ConversionFactor: uom.ConversionRate,
		})
	}
	result, err := client.SaveItem(WithSiteCode(ctx.GetContext(), companySetting.ErpnextSiteCode), &item.ItemInfo{
		ItemName:      params.ItemName,
		ItemGroup:     item.ItemGroup_RawMaterial,
		StockUom:      params.StockUom,
		ValuationRate: params.ValuationRate,
		OpeningStock:  params.OpeningStock,
		IsStockItem:   true,
		Branch:        companySetting.ErpnextBranchName,
		CompanyAbbr:   companySetting.ErpnextCompanyAbbr,
		Uoms:          unitList,
	})
	if err != nil {
		return nil, err
	}
	response := &item.ItemInfo{}
	if err := result.Data.UnmarshalTo(response); err != nil {
		return nil, err
	}

	return response, nil
}
