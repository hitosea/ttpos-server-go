package stock

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/api/item"
	"ttpos-bmp/app/ttpos-erp/internal/consts"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto"
	"ttpos-bmp/app/ttpos-erp/internal/service"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

var (
	Stock = new(sStock)
)

type sStock struct {
}

func init() {
	service.RegisterStock(Stock)
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
