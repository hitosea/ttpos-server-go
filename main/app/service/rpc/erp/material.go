package erp

import (
	"ttpos-bmp/app/ttpos-erp/api/item"
	"ttpos-bmp/app/ttpos-erp/api/manufacturing"
	"ttpos-server-go/app/cloud"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/context"

	"google.golang.org/grpc"
)

func NewErpBomClient() (manufacturing.BomServiceClient, *grpc.ClientConn, error) {
	conn, err := cloud.GetRpcConnWithName(cloud.ErpServiceName)
	if err != nil {
		return nil, nil, err
	}
	return manufacturing.NewBomServiceClient(conn), conn, nil
}

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

	param := &item.ItemInfo{
		ItemCode:           params.ItemCode,
		ItemName:           params.ItemName,
		ItemGroup:          item.ItemGroup_RawMaterial,
		StockUom:           params.StockUom,
		ValuationRate:      params.ValuationRate,
		OpeningStock:       params.OpeningStock,
		IsStockItem:        true,
		Disabled:           params.Disabled,
		Branch:             companySetting.ErpnextBranchName,
		CompanyAbbr:        companySetting.ErpnextCompanyAbbr,
		Uoms:               unitList,
		InternalCode:       params.InternalCode,
		Classification:     params.Classification,
		ClassificationCode: params.ClassificationCode,
		PurchaseUom:        params.PurchaseUom,
	}
	result, err := client.SaveItem(WithSiteCode(ctx.GetContext(), companySetting.ErpnextSiteCode), param)
	if err != nil {
		return nil, err
	}
	if result.GetCode() != "0" {
		return nil, errors.WithMessage(errors.New(result.GetMessage()), "同步物品到erp失败")
	}
	response := &item.ItemInfo{}
	if err := result.Data.UnmarshalTo(response); err != nil {
		return nil, err
	}

	return response, nil
}

// 获取总部的物品列表
func (s *erpSrv) GetHeadquarterMaterialList(ctx context.Context, params req.GetHeadquarterMaterialListReq) (*item.GetItemListResp, error) {
	company := ctx.GetCompany()
	companySetting := company.CompanySetting

	client, conn, err := NewErpItemClient()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// 获取总部的公司设置
	headquarterShopCompanySetting := repository.NewCompanySettingRepo(s.dbm.GetDB(companySetting.HeadquarterUuid)).Get()

	param := &item.GetItemListReq{
		ItemGroup:      item.ItemGroup_RawMaterial,
		Branch:         headquarterShopCompanySetting.ErpnextBranchName,
		CompanyAbbr:    companySetting.ErpnextHeadquarterAbbr, // 总部
		SubCompanyAbbr: companySetting.ErpnextCompanyAbbr,     // 子店
	}
	result, err := client.GetItemList(WithSiteCode(ctx.GetContext(), companySetting.ErpnextSiteCode), param)
	if err != nil {
		return nil, err
	}
	if result.GetCode() != "0" {
		return nil, errors.WithMessage(errors.New(result.GetMessage()), "获取总部物品列表失败")
	}
	response := &item.GetItemListResp{}
	if err := result.Data.UnmarshalTo(response); err != nil {
		return nil, err
	}

	return response, nil
}

// 获取总部的物品列表
func (s *erpSrv) GetSubShopMaterialList(ctx context.Context) (*item.GetItemListResp, error) {
	company := ctx.GetCompany()
	companySetting := company.CompanySetting

	client, conn, err := NewErpItemClient()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	param := &item.GetItemListReq{
		ItemGroup:   item.ItemGroup_RawMaterial,
		Branch:      companySetting.ErpnextBranchName,
		CompanyAbbr: companySetting.ErpnextCompanyAbbr, // 总部
	}
	result, err := client.GetItemList(WithSiteCode(ctx.GetContext(), companySetting.ErpnextSiteCode), param)
	if err != nil {
		return nil, err
	}
	if result.GetCode() != "0" {
		return nil, errors.WithMessage(errors.New(result.GetMessage()), "获取总部物品列表失败")
	}
	response := &item.GetItemListResp{}
	if err := result.Data.UnmarshalTo(response); err != nil {
		return nil, err
	}

	return response, nil
}

// 获取物品列表请求参数
type GetMaterialListReq struct {
	SiteCode        string `json:"site_code"`
	Branch          string `json:"branch"`
	CompanyAbbr     string `json:"company_abbr"`
	ContainDisabled bool   `json:"contain_disabled"`
}

// 获取物品列表
func (s *erpSrv) GetMaterialList(ctx context.Context, params GetMaterialListReq) (*item.GetItemListResp, error) {
	client, conn, err := NewErpItemClient()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	param := &item.GetItemListReq{
		ItemGroup:       item.ItemGroup_RawMaterial,
		Branch:          params.Branch,
		CompanyAbbr:     params.CompanyAbbr,
		ContainDisabled: params.ContainDisabled,
	}

	result, err := client.GetItemList(WithSiteCode(ctx.GetContext(), params.SiteCode), param)
	if err != nil {
		return nil, err
	}
	if result.GetCode() != "0" {
		return nil, errors.WithMessage(errors.New(result.GetMessage()), "获取物品列表失败")
	}
	response := &item.GetItemListResp{}
	if err := result.Data.UnmarshalTo(response); err != nil {
		return nil, err
	}

	return response, nil
}

type ProductBomCardAddErpReq struct {
	ItemCode string                   `json:"item_code" binding:"required"` // 商品编码
	Quantity float64                  `json:"quantity" binding:"required"`  // 数量，加工份数
	Uom      string                   `json:"uom" binding:"required"`       // 单位,商品的单位
	Items    []*manufacturing.BomItem `json:"items" binding:"required"`     // 物品列表
}

func (s *erpSrv) AddProductBomCard(ctx context.Context, params ProductBomCardAddErpReq) (*manufacturing.SaveBomResp, error) {
	company := ctx.GetCompany()
	companySetting := company.CompanySetting

	client, conn, err := NewErpBomClient()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	result, err := client.SaveBom(WithSiteCode(ctx.GetContext(), companySetting.ErpnextSiteCode), &manufacturing.SaveBomReq{
		ItemCode:    params.ItemCode,
		CompanyAbbr: companySetting.ErpnextCompanyAbbr,
		Quantity:    params.Quantity,
		Uom:         params.Uom,
		IsActive:    true,
		IsDefault:   true,
		Items:       params.Items,
	})
	if err != nil {
		return nil, err
	}
	if result.GetCode() != "0" {
		return nil, errors.WithMessage(errors.New(result.GetMessage()), "同步成本卡到erp失败")
	}
	response := &manufacturing.SaveBomResp{}
	if err := result.Data.UnmarshalTo(response); err != nil {
		return nil, err
	}

	return response, nil
}

func (s *erpSrv) GetMaterialStockNum(ctx context.Context, warehouseErpCode string) ([]*item.ItemStock, error) {
	company := ctx.GetCompany()
	companySetting := company.CompanySetting

	client, conn, err := NewErpItemClient()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	result, err := client.GetItemStock(WithSiteCode(ctx.GetContext(), companySetting.ErpnextSiteCode), &item.GetItemStockReq{
		CompanyAbbr: companySetting.ErpnextCompanyAbbr,
		Branch:      companySetting.ErpnextBranchName,
		ItemGroup:   item.ItemGroup_RawMaterial,
		Warehouse:   warehouseErpCode,
	})
	if err != nil {
		return nil, err
	}
	if result.GetCode() != "0" {
		return nil, errors.WithMessage(errors.New(result.GetMessage()), "获取物品库存数量失败")
	}
	response := &item.GetItemStockResp{}
	if err := result.Data.UnmarshalTo(response); err != nil {
		return nil, err
	}

	return response.ItemStockList, nil
}

// 获取成本卡列表
func (s *erpSrv) GetProductBomCardList(ctx context.Context) (*manufacturing.GetBomListResp, error) {
	company := ctx.GetCompany()
	companySetting := company.CompanySetting

	client, conn, err := NewErpBomClient()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	result, err := client.GetBomList(WithSiteCode(ctx.GetContext(), companySetting.ErpnextSiteCode), &manufacturing.GetBomListReq{
		CompanyAbbr:    companySetting.ErpnextHeadquarterAbbr,
		SubCompanyAbbr: companySetting.ErpnextCompanyAbbr,
		IsActive:       true,
		IsDefault:      true,
	})
	if err != nil {
		return nil, err
	}
	if result.GetCode() != "0" {
		return nil, errors.WithMessage(errors.New(result.GetMessage()), "获取成本卡列表失败")
	}
	response := &manufacturing.GetBomListResp{}
	if err := result.Data.UnmarshalTo(response); err != nil {
		return nil, err
	}

	return response, nil
}

// 获取成本卡详情
func (s *erpSrv) GetProductBomCardDetail(ctx context.Context, params req.ErpProductBomCardDetailReq) (*manufacturing.GetBomResp, error) {
	company := ctx.GetCompany()
	companySetting := company.CompanySetting

	client, conn, err := NewErpBomClient()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	result, err := client.GetBom(WithSiteCode(ctx.GetContext(), companySetting.ErpnextSiteCode), &manufacturing.GetBomReq{
		BomName: params.BomName,
	})
	if err != nil {
		return nil, err
	}
	if result.GetCode() != "0" {
		return nil, errors.WithMessage(errors.New(result.GetMessage()), "获取成本卡详情失败")
	}
	response := &manufacturing.GetBomResp{}
	if err := result.Data.UnmarshalTo(response); err != nil {
		return nil, err
	}
	return response, nil
}
