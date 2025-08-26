package erp

import (
	"ttpos-bmp/app/ttpos-erp/api/item"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/pkg/context"
)

// 添加商品到erp
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
	if result.GetCode() != "0" {
		return nil, errors.WithMessage(errors.New(result.GetMessage()), "同步商品到erp失败")
	}
	response := &item.ItemInfo{}
	if err := result.Data.UnmarshalTo(response); err != nil {
		return nil, err
	}

	return response, nil
}

// 更新套餐到erp
func (s *erpSrv) AddPackage(ctx context.Context, params req.PackageAddErpReq) (*item.ItemInfo, error) {
	company := ctx.GetCompany()
	companySetting := company.CompanySetting

	client, conn, err := NewErpItemClient()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	param := &item.ItemInfo{
		ItemName:    params.ItemName,
		ItemGroup:   item.ItemGroup_Package,
		StockUom:    params.StockUom,
		ItemCode:    params.ItemCode,
		Branch:      companySetting.ErpnextBranchName,
		CompanyAbbr: companySetting.ErpnextCompanyAbbr,
		Uoms:        []*item.UomDetail{},
	}
	result, err := client.SaveItem(WithSiteCode(ctx.GetContext(), companySetting.ErpnextSiteCode), param)
	if err != nil {
		return nil, err
	}
	if result.GetCode() != "0" {
		return nil, errors.WithMessage(errors.New(result.GetMessage()), "同步套餐到erp失败")
	}
	response := &item.ItemInfo{}
	if err := result.Data.UnmarshalTo(response); err != nil {
		return nil, err
	}

	return response, nil
}
