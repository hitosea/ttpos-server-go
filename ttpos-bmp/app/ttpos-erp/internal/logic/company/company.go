package company

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/api/company"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto"
	"ttpos-bmp/app/ttpos-erp/internal/service"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

var Company = new(sCompany)

type sCompany struct {
}

func init() {
	service.RegisterCompany(Company)
}

func (s *sCompany) GetCompanyList(ctx context.Context, req *company.GetCompanyListReq) (res *company.GetCompanyListResp, err error) {
	var filters = make([][]string, 0)
	// 调用 erpnext 下 document 服务，从外部获取 company 信息
	if len(req.CompanyName) > 0 {
		filters = append(filters, g.ArrayStr{"name", "like", "%" + req.CompanyName + "%"})
	}
	if len(req.CompanyAbbr) > 0 {
		filters = append(filters, g.ArrayStr{"abbr", "like", "%" + req.CompanyAbbr + "%"})
	}
	if len(req.ParentCompany) > 0 {
		filters = append(filters, g.ArrayStr{"parent_company", "like", "%" + req.ParentCompany + "%"})
	}
	// 1. 调用 erpnext document 服务获取公司信息
	erpCompanyList, err := service.Document().List(ctx, &dto.ErpReq{
		DocType: "Company",
	}, &dto.RequestParams{
		Fields:  g.ArrayStr{"name", "abbr"},
		Filters: filters,
	})
	if err != nil {
		// 错误处理，返回错误信息
		return nil, err
	}
	if j, err := gjson.DecodeToJson(erpCompanyList.Bytes()); err == nil {
		// 遍历j.Get("data") 返回的数组字段，设置到 CompanyList 中
		companyList := make([]*company.CompanyInfo, 0)
		dataArray := j.GetJsons("data")
		for _, item := range dataArray {
			companyInfo := &company.CompanyInfo{
				CompanyName: item.Get("name").String(),
				CompanyAbbr: item.Get("abbr").String(),
			}
			companyList = append(companyList, companyInfo)
		}
		res = &company.GetCompanyListResp{
			CompanyList: companyList,
		}

	} else {
		g.Log().Error(ctx, "解析公司列表失败", err)
	}
	return
}

func (s *sCompany) GetCompanyWithAbbr(ctx context.Context, abbr string) (res *company.CompanyInfo, err error) {

	var companyList *company.GetCompanyListResp
	if companyList, err = service.Company().GetCompanyList(ctx, &company.GetCompanyListReq{
		CompanyAbbr: abbr,
	}); err != nil {
		g.Log().Error(ctx, "获取公司列表失败", err)
		return nil, gerror.New("获取公司列表失败")
	}
	if len(companyList.CompanyList) > 0 {
		return companyList.CompanyList[0], nil
	}
	return nil, gerror.New("公司不存在")
}
