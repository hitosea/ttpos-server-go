package stock

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/api/item"
	"ttpos-bmp/app/ttpos-erp/internal/consts"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto"
	"ttpos-bmp/app/ttpos-erp/internal/service"
	"ttpos-bmp/app/ttpos-erp/utility"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
)

var (
	Stock = new(sStock)
)

type sStock struct {
}

func init() {
	service.RegisterStock(Stock)
}

func (s *sStock) GetItemList(ctx context.Context, req *item.GetItemListReq) (res *item.GetItemListResp, err error) {
	var filters = make([][]string, 0)

	if len(req.Branch) > 0 {
		filters = append(filters, g.ArrayStr{"custom_branch", "like", "%" + req.Branch + "%"})
	}
	if len(req.ItemName) > 0 {
		filters = append(filters, g.ArrayStr{"name", "like", "%" + req.ItemName + "%"})
	}
	if len(req.ItemGroup) > 0 {
		filters = append(filters, g.ArrayStr{"item_group", "=", req.ItemGroup})
	}
	if len(req.ItemCode) > 0 {
		filters = append(filters, g.ArrayStr{"item_code", "=", req.ItemCode})
	}
	if len(req.CompanyAbbr) > 0 {
		company, err := service.Company().GetCompanyWithAbbr(ctx, req.CompanyAbbr)
		if err == nil {
			filters = append(filters, []string{"custom_company", "like", company.CompanyName})
		}
	}
	filters = append(filters, []string{"disabled", "!=", "1"})

	itemList := make([]*item.ItemInfo, 0)
	if resp, err := service.Document().List(ctx, &dto.ErpReq{
		DocType: "Item",
	}, &dto.RequestParams{
		Fields:  g.ArrayStr{"name", "item_code", "item_group", "custom_branch", "custom_company", "custom_specification"},
		Filters: filters,
		Limit:   consts.Limit999,
	}); err != nil {
		return nil, gerror.Wrapf(err, "查询物品列表失败")
	} else {
		if j, err := gjson.DecodeToJson(resp.Bytes()); err == nil {
			// 遍历j.Get("data") 返回的数组字段，设置到 ItemList 中
			dataArray := j.GetJsons("data")
			for _, data := range dataArray {
				itemInfo := &item.ItemInfo{
					Branch:    data.Get("custom_branch").String(),
					Company:   data.Get("custom_company").String(),
					ItemName:  data.Get("name").String(),
					ItemCode:  data.Get("item_code").String(),
					ItemGroup: data.Get("item_group").String(),
					StockUom:  data.Get("stock_uom").String(),
				}
				itemList = append(itemList, itemInfo)
			}
		}
	}
	return &item.GetItemListResp{
		ItemList: itemList,
	}, nil
}

func (s *sStock) SaveItem(ctx context.Context, req *item.ItemInfo) (res *item.ItemInfo, err error) {
	var (
		filters     = make([][]string, 0)
		companyName string
	)
	// 物品分组是原材料，默认是库存物品
	if req.ItemGroup == string(consts.ItemGroupRawMaterial) {
		req.IsStockItem = true
	}
	if len(req.CompanyAbbr) > 0 {
		company, err := service.Company().GetCompanyWithAbbr(ctx, req.CompanyAbbr)
		if err != nil {
			return nil, gerror.Wrapf(err, "查询公司信息失败")
		}
		companyName = company.CompanyName
	}
	if len(req.ItemCode) > 0 {
		filters = append(filters, g.ArrayStr{"item_code", "=", req.ItemCode})
	}

	if count, err := service.Doctype().Count(ctx, &dto.ErpReq{
		DocType: "Item",
	}, &dto.RequestParams{
		Filters: filters,
	}); err != nil {
		return nil, gerror.Wrapf(err, "查询现有物品失败")
	} else {
		if count > 0 {
			_, err = service.Document().Update(ctx, &dto.ErpReq{
				DocType: "Item",
				Name:    req.ItemCode,
			}, &g.Map{
				"item_name":          req.ItemName,
				"stock_uom":          req.StockUom,
				"custom_company":     companyName,
				"custom_branch":      req.Branch,
				"item_specification": req.ItemSpecification,
			})
		} else {
			//新增
			if req.ItemGroup == string(consts.ItemGroupRawMaterial) {
				req.ItemCode = utility.GenItemCode(consts.ItemCodePrefixRawMaterial)
			} else {
				req.ItemCode = utility.GenItemCode(consts.ItemCodePrefixProduct)

				if len(req.ItemSpecification) == 0 {
					//00为同一个商品无规格的自增后缀
					req.ItemCode += "_00"
				} else {
					// 取存在多规格的商品 TODO

				}
			}
			_, err = service.Document().Create(ctx, "Item", &g.Map{
				"item_code":          req.ItemCode,
				"item_name":          req.ItemName,
				"stock_uom":          req.StockUom,
				"custom_company":     companyName,
				"custom_branch":      req.Branch,
				"item_specification": req.ItemSpecification,
				"is_stock_item":      req.IsStockItem,
			})
		}
		res = &item.ItemInfo{}
		gconv.Scan(req, res)
	}
	return
}

func (s *sStock) GetUomList(ctx context.Context, req *item.GetUomListReq) (res *item.GetUomListResp, err error) {
	var filters = make([][]string, 0)

	if len(req.Branch) > 0 {
		filters = append(filters, g.ArrayStr{"custom_branch", "like", "%" + req.Branch + "%"})
	}
	if len(req.UomName) > 0 {
		filters = append(filters, g.ArrayStr{"name", "like", "%" + req.UomName + "%"})
	}
	if len(req.AliasName) > 0 {
		filters = append(filters, g.ArrayStr{"custom_alias", "like", "%" + req.AliasName + "%"})
	}

	if len(req.CompanyAbbr) > 0 {
		company, err := service.Company().GetCompanyWithAbbr(ctx, req.CompanyAbbr)
		if err == nil {
			filters = append(filters, []string{"custom_company", "like", company.CompanyName})
		}
	}
	filters = append(filters, []string{"enabled", "=", "1"})

	if resp, err := service.Document().List(ctx, &dto.ErpReq{
		DocType: "UOM",
	}, &dto.RequestParams{
		Fields:  g.ArrayStr{"name", "must_be_whole_number", "custom_alias", "custom_company", "custom_branch"},
		Filters: filters,
		Limit:   consts.Limit999,
	}); err != nil {
		return nil, gerror.Wrapf(err, "查询单位列表失败")
	} else {
		if j, err := gjson.DecodeToJson(resp.Bytes()); err == nil {
			uomList := make([]*item.UomInfo, 0)
			// 遍历j.Get("data") 返回的数组字段，设置到 UomList 中
			dataArray := j.GetJsons("data")
			for _, uom := range dataArray {
				uomInfo := &item.UomInfo{
					UomName:           uom.Get("name").String(),
					AliasName:         uom.Get("custom_alias").String(),
					Company:           uom.Get("custom_company").String(),
					Branch:            uom.Get("custom_branch").String(),
					MustBeWholeNumber: uom.Get("must_be_whole_number").Bool(),
				}
				uomList = append(uomList, uomInfo)
			}
			res = &item.GetUomListResp{
				UomList: uomList,
			}
		}
	}
	return
}

func (s *sStock) SaveUom(ctx context.Context, req *item.UomInfo) (err error) {
	var (
		companyName = ""
		filters     = make([][]string, 0)
	)
	if len(req.CompanyAbbr) > 0 {
		company, err := service.Company().GetCompanyWithAbbr(ctx, req.CompanyAbbr)
		if err != nil {
			g.Log().Error(ctx, "根据公司缩写查询公司失败", err)
			return gerror.Wrapf(err, "根据公司缩写查询公司失败")
		}
		companyName = company.CompanyName
	}
	filters = append(filters, g.ArrayStr{"uom_name", "=", req.UomName})
	if count, err := service.Doctype().Count(ctx, &dto.ErpReq{
		DocType: "UOM",
	}, &dto.RequestParams{
		Filters: filters,
	}); err != nil {
		return gerror.Wrapf(err, "查询现有单位失败")
	} else {
		if count > 0 {
			_, err = service.Document().Update(ctx, &dto.ErpReq{
				DocType: "UOM",
				Name:    req.UomName,
			}, &g.Map{
				"custom_alias":         req.AliasName,
				"custom_company":       companyName,
				"custom_branch":        req.Branch,
				"must_be_whole_number": req.MustBeWholeNumber,
			})
		} else {
			_, err = service.Document().Create(ctx, "UOM", &g.Map{
				"uom_name":             req.UomName,
				"custom_alias":         req.AliasName,
				"custom_company":       companyName,
				"custom_branch":        req.Branch,
				"must_be_whole_number": req.MustBeWholeNumber,
			})
		}
	}

	return
}

func (s *sStock) GetAttributeList(ctx context.Context, req *item.GetAttributeListReq) (res *item.GetAttributeListResp, err error) {
	var filters = make([][]string, 0)
	if len(req.Branch) > 0 {
		filters = append(filters, g.ArrayStr{"custom_branch", "like", "%" + req.Branch + "%"})
	}
	if len(req.AttributeName) > 0 {
		filters = append(filters, g.ArrayStr{"attribute_name", "like", "%" + req.AttributeName + "%"})
	}
	if len(req.AliasName) > 0 {
		filters = append(filters, g.ArrayStr{"custom_alias", "like", "%" + req.AliasName + "%"})
	}
	if len(req.CompanyAbbr) > 0 {
		company, err := service.Company().GetCompanyWithAbbr(ctx, req.CompanyAbbr)
		if err == nil {
			filters = append(filters, []string{"custom_company", "like", company.CompanyName})
		}
	}
	if resp, err := service.Document().List(ctx, &dto.ErpReq{
		DocType: "Item Attribute",
	}, &dto.RequestParams{
		Fields:  g.ArrayStr{"name", "attribute_name", "custom_alias", "custom_company", "custom_branch"},
		Filters: filters,
		Limit:   consts.Limit999,
	}); err != nil {
		return nil, gerror.Wrapf(err, "查询属性列表失败")
	} else {
		if j, err := gjson.DecodeToJson(resp.Bytes()); err == nil {
			dataList := make([]*item.AttributeInfo, 0)
			// 遍历j.Get("data") 返回的数组字段，设置到 UomList 中
			dataArray := j.GetJsons("data")
			for _, it := range dataArray {
				dataInfo := &item.AttributeInfo{
					AttributeName: it.Get("name").String(),
					AliasName:     it.Get("custom_alias").String(),
					Company:       it.Get("custom_company").String(),
					Branch:        it.Get("custom_branch").String(),
				}
				if values, err := s.GetAttributeValuesList(ctx, dataInfo.AttributeName); err == nil {
					dataInfo.AttributeValueList = values
				}
				dataList = append(dataList, dataInfo)
			}
			res = &item.GetAttributeListResp{
				AttributeList: dataList,
			}
		}
	}
	return
}

func (s *sStock) GetAttributeValuesList(ctx context.Context, attributeName string) (res []*item.AttributeValueInfo, err error) {
	res = make([]*item.AttributeValueInfo, 0)
	if resp, err := service.Document().Get(ctx, &dto.ErpReq{
		DocType: "Item Attribute",
		Name:    attributeName,
	}, nil); err != nil {
		return nil, gerror.Wrapf(err, "查询属性值失败: %s", attributeName)
	} else {
		if j, err := gjson.DecodeToJson(resp.Bytes()); err == nil {
			// 遍历j.Get("data") 返回的数组字段，设置到 UomList 中
			dataArray := j.GetJsons("data.item_attribute_values")
			for _, it := range dataArray {
				dataInfo := &item.AttributeValueInfo{
					AttributeValue: it.Get("attribute_value").String(),
					Abbr:           it.Get("abbr").String(),
				}
				res = append(res, dataInfo)
			}

		}
	}
	return
}

func (s *sStock) SaveAttribute(ctx context.Context, req *item.AttributeInfo) (err error) {
	var (
		companyName = ""
		filters     = make([][]string, 0)
	)
	if len(req.CompanyAbbr) > 0 {
		company, err := service.Company().GetCompanyWithAbbr(ctx, req.CompanyAbbr)
		if err != nil {
			g.Log().Error(ctx, "根据公司缩写查询公司失败", err)
			return gerror.Wrapf(err, "根据公司缩写查询公司失败")
		}
		companyName = company.CompanyName
	}
	filters = append(filters, g.ArrayStr{"attribute_name", "=", req.AttributeName})
	if count, err := service.Doctype().Count(ctx, &dto.ErpReq{
		DocType: "Item Attribute",
	}, &dto.RequestParams{
		Filters: filters,
	}); err != nil {
		return gerror.Wrapf(err, "查询现有属性失败")
	} else {
		if count > 0 {
			_, err = service.Document().Update(ctx, &dto.ErpReq{
				DocType: "Item Attribute",
				Name:    req.AttributeName,
			}, &g.Map{
				"custom_alias":   req.AliasName,
				"custom_company": companyName,
				"custom_branch":  req.Branch,
			})
		} else {
			_, err = service.Document().Create(ctx, "UOM", &g.Map{
				"attribute_name": req.AttributeName,
				"custom_alias":   req.AliasName,
				"custom_company": companyName,
				"custom_branch":  req.Branch,
			})
		}
	}

	return
}
