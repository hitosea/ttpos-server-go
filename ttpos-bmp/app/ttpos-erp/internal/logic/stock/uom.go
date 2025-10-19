package stock

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/api/item"
	"ttpos-bmp/app/ttpos-erp/internal/consts"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto/erp"
	"ttpos-bmp/app/ttpos-erp/internal/service"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
)

var (
	Uom = new(sUom)
)

type sUom struct {
}

func init() {
	service.RegisterUom(Uom)
}
func (s *sUom) DeleteUom(ctx context.Context, req *item.DeleteUomReq) (*item.DeleteUomResp, error) {
	_, err := service.Document().Delete(ctx, &erp.ErpReq{
		DocType: erp.DocTypeUom,
		Name:    req.Uom,
	})
	if err != nil {
		return nil, gerror.Wrapf(err, "删除单位失败")
	}
	return &item.DeleteUomResp{Uom: req.Uom}, nil
}

// SaveUom 保存单位信息
// 如果单位已存在则更新，否则创建新单位
func (s *sUom) SaveUom(ctx context.Context, req *item.UomInfo) error {
	// 获取公司名称
	companyName, err := service.Company().GetCompanyNameWithAbbr(ctx, req.CompanyAbbr)
	if err != nil {
		return err
	}

	// 检查单位是否已存在
	exists, err := s.checkUomExists(ctx, req.UomName)
	if err != nil {
		return err
	}

	if exists {
		// 更新现有单位
		return s.updateExistingUom(ctx, req, companyName)
	} else {
		// 创建新单位
		return s.createNewUom(ctx, req, companyName)
	}
}

// checkUomExists 检查单位是否已存在
func (s *sUom) checkUomExists(ctx context.Context, uomName string) (bool, error) {
	if len(uomName) == 0 {
		return false, nil
	}

	filters := [][]string{{"uom_name", "=", uomName}}

	count, err := service.Doctype().Count(ctx, &erp.ErpReq{
		DocType: erp.DocTypeUom,
	}, &erp.RequestParams{
		Filters: filters,
	})

	if err != nil {
		return false, gerror.Wrapf(err, "查询现有单位失败")
	}

	return count > 0, nil
}

// updateExistingUom 更新现有单位
func (s *sUom) updateExistingUom(ctx context.Context, req *item.UomInfo, companyName string) error {
	_, err := service.Document().Update(ctx, &erp.ErpReq{
		DocType: erp.DocTypeUom,
		Name:    req.UomName,
	}, &g.Map{
		"custom_alias":         req.AliasName,
		"custom_company":       companyName,
		"custom_branch":        req.Branch,
		"must_be_whole_number": req.MustBeWholeNumber,
	})

	if err != nil {
		return gerror.Wrapf(err, "更新单位信息失败")
	}

	return nil
}

// createNewUom 创建新单位
func (s *sUom) createNewUom(ctx context.Context, req *item.UomInfo, companyName string) error {
	_, err := service.Document().Create(ctx, "UOM", &g.Map{
		"uom_name":             req.UomName,
		"custom_alias":         req.AliasName,
		"custom_company":       companyName,
		"custom_branch":        req.Branch,
		"must_be_whole_number": req.MustBeWholeNumber,
	})

	if err != nil {
		return gerror.Wrapf(err, "创建单位失败")
	}

	return nil
}

// GetUomList 获取单位列表
// 根据查询条件过滤并返回单位信息列表
func (s *sUom) GetUomList(ctx context.Context, req *item.GetUomListReq) (res *item.GetUomListResp, err error) {
	// 构建查询过滤器
	filters := s.buildUomListFilters(ctx, req)

	// 查询单位列表
	uomList, err := s.queryUomList(ctx, filters, req)
	if err != nil {
		return nil, gerror.Wrapf(err, "查询单位列表失败")
	}

	return &item.GetUomListResp{
		UomList: uomList,
	}, nil
}

// buildUomListFilters 构建单位列表查询过滤器
func (s *sUom) buildUomListFilters(ctx context.Context, req *item.GetUomListReq) [][]string {
	filters := make([][]string, 0)

	// 按分支机构过滤
	if len(req.Branch) > 0 {
		filters = append(filters, g.ArrayStr{"custom_branch", "like", "%" + req.Branch + "%"})
	}

	// 按单位名称过滤
	if len(req.UomName) > 0 {
		filters = append(filters, g.ArrayStr{"name", "like", "%" + req.UomName + "%"})
	}

	// 按别名过滤
	if len(req.AliasName) > 0 {
		filters = append(filters, g.ArrayStr{"custom_alias", "like", "%" + req.AliasName + "%"})
	}

	// 按公司简称过滤
	if len(req.CompanyAbbr) > 0 {
		companyName, _ := service.Company().GetCompanyNameWithAbbr(ctx, req.CompanyAbbr)
		if len(companyName) > 0 {
			filters = append(filters, []string{"custom_company", "like", companyName})
		}
	}

	// 只查询启用的单位
	filters = append(filters, []string{"enabled", "=", "1"})

	return filters
}

// queryUomList 执行单位列表查询
func (s *sUom) queryUomList(ctx context.Context, filters [][]string, req *item.GetUomListReq) ([]*item.UomInfo, error) {
	resp, err := service.Document().List(ctx, &erp.ErpReq{
		DocType: erp.DocTypeUom,
	}, &erp.RequestParams{
		Fields:  g.ArrayStr{"name", "must_be_whole_number", "custom_alias", "custom_company", "custom_branch"},
		Filters: filters,
		Limit:   consts.Limit999,
	})

	if err != nil {
		return nil, err
	}

	// 解析响应数据
	j := resp

	// 转换为单位信息列表
	uomList := make([]*item.UomInfo, 0)
	dataArray := j.GetJsons("data")

	subCompanyName := ""
	if len(req.SubCompanyAbbr) > 0 {
		subCompanyName, err = service.Company().GetCompanyNameWithAbbr(ctx, req.SubCompanyAbbr)
		if err != nil {
			return nil, gerror.Wrapf(err, "获取子公司名称失败,companyAbbr:%s", req.SubCompanyAbbr)
		}
	}
	for _, uom := range dataArray {
		//获取当前item 的所有 Item Permission List 判断是否有权限使用, 列表查询目前不支持 表格字段查询，只能每次遍历物品查询 对应的 custom_permission_rule
		if len(req.SubCompanyAbbr) > 0 {
			uomInfo, err := s.GetUom(ctx, &item.GetUomReq{
				UomName: uom.Get("name").String(),
			})
			if err != nil {
				g.Log().Errorf(ctx, "获取单位信息失败,uomName:%s,err:%v", uom.Get("name").String(), err)
				continue // 获取单位信息失败，跳过该单位
			}
			hasPermission, err := service.Permission().CheckPermission(ctx, uomInfo.CustomPermissionRule, subCompanyName)
			if err != nil {
				g.Log().Errorf(ctx, "检查单位权限失败,uomName:%s,err:%v", uom.Get("name").String(), err)
				continue // 检查权限失败，跳过该单位
			}
			if !hasPermission {
				continue // 当前子公司无权限，跳过该单位
			}
		}
		uomInfo := &item.UomInfo{
			UomName:           uom.Get("name").String(),
			AliasName:         uom.Get("custom_alias").String(),
			Company:           uom.Get("custom_company").String(),
			Branch:            uom.Get("custom_branch").String(),
			MustBeWholeNumber: uom.Get("must_be_whole_number").Bool(),
		}
		uomList = append(uomList, uomInfo)
	}

	return uomList, nil
}

// GetUom 根据单位名称获取单个单位详细信息
// 参数：ctx 上下文，req 包含单位名称
// 返回：单位详细信息，错误信息
func (s *sUom) GetUom(ctx context.Context, req *item.GetUomReq) (res *erp.UOM, err error) {
	// 参数验证
	if len(req.UomName) == 0 {
		return nil, gerror.New("单位名称不能为空")
	}

	// 查询单位信息
	resp, err := service.Document().Get(ctx, &erp.ErpReq{
		DocType: erp.DocTypeUom,
		Name:    req.UomName,
	}, nil)

	if err != nil {
		return nil, gerror.Wrapf(err, "查询单位信息失败")
	}

	// 解析响应数据
	j := resp

	// 转换为单位信息结构体
	uomInfo := &erp.UOM{}
	gconv.Structs(j.GetJson("data"), &uomInfo)

	return uomInfo, nil
}
