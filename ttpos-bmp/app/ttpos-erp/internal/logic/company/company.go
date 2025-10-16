package company

import (
	"context"
	"strings"
	"ttpos-bmp/app/ttpos-erp/api/company"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto/erp"
	"ttpos-bmp/app/ttpos-erp/internal/service"

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
		if strings.Contains(req.CompanyName, "%") {
			filters = append(filters, g.ArrayStr{"name", "like", req.CompanyName})
		} else {
			filters = append(filters, g.ArrayStr{"name", "=", req.CompanyName})
		}
	}
	if len(req.CompanyAbbr) > 0 {
		if strings.Contains(req.CompanyAbbr, "%") {
			filters = append(filters, g.ArrayStr{"abbr", "like", req.CompanyAbbr})
		} else {
			filters = append(filters, g.ArrayStr{"abbr", "=", req.CompanyAbbr})
		}
	}
	if len(req.ParentCompany) > 0 {
		if strings.Contains(req.ParentCompany, "%") {
			filters = append(filters, g.ArrayStr{"parent_company", "like", req.ParentCompany})
		} else {
			filters = append(filters, g.ArrayStr{"parent_company", "=", req.ParentCompany})
		}
	}
	// 1. 调用 erpnext document 服务获取公司信息
	erpCompanyList, err := service.Document().List(ctx, &erp.ErpReq{
		DocType: "Company",
	}, &erp.RequestParams{
		Fields:  g.ArrayStr{"name", "abbr", "parent_company"},
		Filters: filters,
		Limit:   1000,
	})
	if err != nil {
		// 错误处理，返回错误信息
		return nil, err
	}
	if j := erpCompanyList; !j.IsNil() {
		// 遍历j.Get("data") 返回的数组字段，设置到 CompanyList 中
		companyList := make([]*company.CompanyInfo, 0)
		dataArray := j.GetJsons("data")
		for _, item := range dataArray {
			companyInfo := &company.CompanyInfo{
				CompanyName:   item.Get("name").String(),
				CompanyAbbr:   item.Get("abbr").String(),
				ParentCompany: item.Get("parent_company").String(),
			}
			//companyInfo.HasChild, err = s.HasSubCompany(ctx, companyInfo.CompanyName)
			//if err != nil {
			//	g.Log().Error(ctx, "查询子公司信息失败", err)
			//}

			companyList = append(companyList, companyInfo)
		}
		res = &company.GetCompanyListResp{
			CompanyList: companyList,
		}
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

// GetCompanyNameWithAbbr  根据公司简称获取公司名称
func (s *sCompany) GetCompanyNameWithAbbr(ctx context.Context, companyAbbr string) (string, error) {
	if len(companyAbbr) == 0 {
		return "", nil
	}

	company, err := service.Company().GetCompanyWithAbbr(ctx, companyAbbr)
	if err != nil {
		return "", gerror.Wrapf(err, "查询公司信息失败")
	}

	return company.CompanyName, nil
}

// HasSubCompany 判断公司是否有子公司
// 通过 parent_company 字段关联查询，判断是否存在子公司
// 参数：ctx 上下文，companyName 公司名称
// 返回：是否有子公司，错误信息
func (s *sCompany) HasSubCompany(ctx context.Context, companyName string) (bool, error) {
	if len(companyName) == 0 {
		return false, gerror.New("公司名称不能为空")
	}

	// 构建查询过滤器，查找 parent_company 等于当前公司的记录
	filters := [][]string{
		{"parent_company", "=", companyName},
	}

	// 使用 service.Doctype().Count 查询子公司数量
	count, err := service.Doctype().Count(ctx, &erp.ErpReq{
		DocType: "Company",
	}, &erp.RequestParams{
		Filters: filters,
	})

	if err != nil {
		return false, gerror.Wrapf(err, "查询子公司信息失败")
	}

	// 如果数量大于0，说明有子公司
	return count > 0, nil
}
