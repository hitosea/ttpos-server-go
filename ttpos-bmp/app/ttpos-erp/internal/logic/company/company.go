package company

import (
	"context"
	"fmt"
	"strings"
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
		if j.Contains("error") {
			return nil, gerror.Newf("获取公司列表失败, err: %v", j.Get("error"))
		}
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

// CreateBranch 创建分店
// 参数：店铺名称和公司缩写编码
// 返回：ERP用户名和创建结果
func (s *sCompany) CreateBranch(ctx context.Context, req *company.CreateBranchReq) (res *company.CreateBranchResp, err error) {
	// 参数验证
	if strings.TrimSpace(req.ShopUuid) == "" {
		g.Log().Error(ctx, "店铺UUID不能为空")
		return nil, gerror.New("店铺UUID不能为空")
	}

	if strings.TrimSpace(req.ShopName) == "" {
		g.Log().Error(ctx, "店铺名称不能为空")
		return nil, gerror.New("店铺名称不能为空")
	}

	if strings.TrimSpace(req.CompanyAbbr) == "" {
		g.Log().Error(ctx, "公司缩写编码不能为空")
		return nil, gerror.New("公司缩写编码不能为空")
	}

	// 调用 erpnext document 服务创建分店
	// 这里需要根据实际的ERP系统API来调用
	// 暂时返回模拟的成功结果
	res = &company.CreateBranchResp{
		ErpUserEmail: fmt.Sprintf("shop%s@ttpos-user.com", req.ShopUuid),
	}

	return res, nil
}
