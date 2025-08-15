package buying

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/api/buying"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto/erp"
	"ttpos-bmp/app/ttpos-erp/internal/service"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

type sBuying struct {
}

var (
	Buying = new(sBuying)
)

func init() {
	service.RegisterBuying(Buying)
}

func (s *sBuying) GetSupplierList(ctx context.Context, req *buying.GetSupplierListReq) (*buying.GetSupplierListResp, error) {
	var filters = make([][]string, 0)
	// 根据公司简称获取公司名称
	companyName, err := service.Company().GetCompanyNameWithAbbr(ctx, req.CompanyAbbr)
	if err != nil {
		return nil, gerror.Wrapf(err, "查询公司名称失败")
	}
	// 根据公司名称获取供应商列表
	filters = append(filters, []string{"Allowed To Transact With", "company", "=", companyName})

	resp, err := service.Document().List(ctx, &erp.ErpReq{
		DocType: "Supplier",
	}, &erp.RequestParams{
		Fields:  g.ArrayStr{"name", "represents_company"},
		Filters: filters,
	})
	if err != nil {
		return nil, gerror.Wrapf(err, "查询供应商列表失败")
	}
	// 解析供应商列表
	supplierList, err := s.parseSupplierListResponse(resp.Bytes())
	if err != nil {
		return nil, gerror.Wrapf(err, "解析供应商列表响应失败")
	}
	return &buying.GetSupplierListResp{
		SupplierList: supplierList,
	}, nil
}

func (s *sBuying) parseSupplierListResponse(bytes []byte) ([]*buying.SupplierInfo, error) {
	// 解析响应数据
	j, err := gjson.DecodeToJson(bytes)
	if err != nil {
		return nil, gerror.Wrapf(err, "解析物品列表响应失败")
	}

	// 转换为物品信息列表
	dataArray := j.GetJsons("data")
	supplierList := make([]*buying.SupplierInfo, 0, len(dataArray))

	for _, data := range dataArray {
		supplierList = append(supplierList, &buying.SupplierInfo{
			SupplierName:      data.Get("name").String(),
			RepresentsCompany: data.Get("represents_company").String(),
		})
	}
	return supplierList, nil
}
