package erp

import (
	"strings"
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
		ItemName:           params.ItemName,
		ItemGroup:          item.ItemGroup_Products,
		StockUom:           params.StockUom,
		ItemCode:           params.ItemCode,
		Branch:             companySetting.ErpnextBranchName,
		CompanyAbbr:        companySetting.ErpnextCompanyAbbr,
		TemplateItemCode:   params.TemplateItemCode,
		ItemSpecification:  params.ItemSpecification,
		Classification:     params.Classification,
		ClassificationCode: params.ClassificationCode,
		InternalCode:       params.InternalCode,
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
		ItemName:           params.ItemName,
		ItemGroup:          item.ItemGroup_Package,
		StockUom:           params.StockUom,
		ItemCode:           params.ItemCode,
		Branch:             companySetting.ErpnextBranchName,
		CompanyAbbr:        companySetting.ErpnextCompanyAbbr,
		Uoms:               []*item.UomDetail{},
		InternalCode:       params.InternalCode,
		Classification:     params.Classification,
		ClassificationCode: params.ClassificationCode,
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

// 删除商品
func (s *erpSrv) DeleteProduct(ctx context.Context, params req.DeleteProductErpReq) error {
	company := ctx.GetCompany()
	companySetting := company.CompanySetting

	client, conn, err := NewErpItemClient()
	if err != nil {
		return errors.WithMessage(err, "删除商品到erp失败")
	}
	defer conn.Close()

	for _, obj := range params.Items {
		result, err := client.SaveItem(WithSiteCode(ctx.GetContext(), companySetting.ErpnextSiteCode), &item.ItemInfo{
			ItemCode: obj.ItemCode,
			Disabled: true,
			ItemName: obj.ItemName,
			StockUom: obj.StockUom,
		})
		if err != nil {
			return errors.WithMessage(err, "删除商品到erp失败")
		}
		if result.GetCode() != "0" {
			return errors.WithMessage(errors.New(result.GetMessage()), "删除商品到erp失败")
		}
	}

	return nil
}

// GetSauceList 获取加料
func (s *erpSrv) GetSauceList(ctx context.Context, sourceListReq req.GetErpSauceListReq) ([]*item.ItemInfo, error) {
	client, conn, err := NewErpItemClient()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	result, err := client.GetItemList(WithSiteCode(ctx.GetContext(), sourceListReq.SiteCode), &item.GetItemListReq{
		ItemGroup:      item.ItemGroup_Products,
		CompanyAbbr:    sourceListReq.CompanyAbbr,
		Branch:         sourceListReq.Branch,
		SubCompanyAbbr: sourceListReq.SubCompanyAbbr,
	})
	if err != nil {
		return nil, err
	}
	if result.GetCode() != "0" {
		return nil, errors.WithMessage(errors.New(result.GetMessage()), "获取加料列表失败")
	}
	response := &item.GetItemListResp{}
	if err := result.Data.UnmarshalTo(response); err != nil {
		return nil, err
	}
	var sauceList []*item.ItemInfo
	// 遍历所有item，判断是否后缀是_00的，且不存在同样前缀_0X的，则认为是加料，
	// 比如 SP3688441864912897_00，且不存在 SP3688441864912897_XX ，则认为是 SP3688441864912897_00 加料
	for _, item := range response.ItemList {
		if !strings.HasSuffix(item.ItemCode, "_00") {
			continue
		}
		itemCodeParts := strings.Split(item.ItemCode, "_")
		if len(itemCodeParts) != 2 {
			continue
		}
		itemCodePrefix := itemCodeParts[0]
		hasSamePrefix := false
		for _, item2 := range response.ItemList {
			if item2.ItemCode != item.ItemCode && strings.Contains(item2.ItemCode, itemCodePrefix) {
				hasSamePrefix = true
				break
			}
		}
		if !hasSamePrefix {
			sauceList = append(sauceList, item)
		}
	}
	return sauceList, nil
}
