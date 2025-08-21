package erp

import (
	"ttpos-bmp/app/ttpos-erp/api/item"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/pkg/context"
)

func (s *erpSrv) AddProduct(ctx context.Context, params req.ProductAddErpReq) (*item.ItemInfo, error) {
	company := ctx.GetCompany()
	companySetting := company.CompanySetting

	client, conn, err := NewErpItemClient()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	result, err := client.SaveItem(WithSiteCode(ctx.GetContext(), companySetting.ErpnextSiteCode), &item.ItemInfo{
		ItemName:          params.ItemName,
		ItemGroup:         item.ItemGroup_Products,
		StockUom:          params.StockUom,
		ItemCode:          params.ItemCode,
		Branch:            companySetting.ErpnextBranchName,
		CompanyAbbr:       companySetting.ErpnextCompanyAbbr,
		TemplateItemCode:  params.TemplateItemCode,
		ItemSpecification: params.ItemSpecification,
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
