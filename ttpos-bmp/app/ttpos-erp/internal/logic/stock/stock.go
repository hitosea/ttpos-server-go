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
		filters = append(filters, g.ArrayStr{"branch", "like", "%" + req.Branch + "%"})
	}
	if len(req.UomName) > 0 {
		filters = append(filters, g.ArrayStr{"name", "like", "%" + req.UomName + "%"})
	}
	if len(req.AliasName) > 0 {
		filters = append(filters, g.ArrayStr{"custom_alias", "like", "%" + req.AliasName + "%"})
	}

	if len(req.CompanyAbbr) > 0 {
		company, err := service.Company().GetCompanyWithAbbr(ctx, req.CompanyAbbr)
		if err != nil {
			g.Log().Error(ctx, "根据公司缩写查询公司失败", err)
			return nil, gerror.Wrapf(err, "根据公司缩写查询公司失败")
		}
		filters = append(filters, []string{"company", "like", company.CompanyName})
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
					CompanyAbbr:       uom.Get("custom_company").String(),
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
			}, &g.Map{
				"uom_name":             req.UomName,
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
