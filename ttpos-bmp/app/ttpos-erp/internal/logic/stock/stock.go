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

type sStock struct{}

func init() {
	service.RegisterStock(Stock)
}

// GetUomList 获取单位列表
// 根据查询条件过滤并返回单位信息列表
func (s *sStock) GetUomList(ctx context.Context, req *item.GetUomListReq) (res *item.GetUomListResp, err error) {
	// 构建查询过滤器
	filters := s.buildUomListFilters(ctx, req)

	// 查询单位列表
	uomList, err := s.queryUomList(ctx, filters)
	if err != nil {
		return nil, gerror.Wrapf(err, "查询单位列表失败")
	}

	return &item.GetUomListResp{
		UomList: uomList,
	}, nil
}

// buildUomListFilters 构建单位列表查询过滤器
func (s *sStock) buildUomListFilters(ctx context.Context, req *item.GetUomListReq) [][]string {
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
func (s *sStock) queryUomList(ctx context.Context, filters [][]string) ([]*item.UomInfo, error) {
	resp, err := service.Document().List(ctx, &dto.ErpReq{
		DocType: "UOM",
	}, &dto.RequestParams{
		Fields:  g.ArrayStr{"name", "must_be_whole_number", "custom_alias", "custom_company", "custom_branch"},
		Filters: filters,
		Limit:   consts.Limit999,
	})

	if err != nil {
		return nil, err
	}

	// 解析响应数据
	j, err := gjson.DecodeToJson(resp.Bytes())
	if err != nil {
		return nil, gerror.Wrapf(err, "解析单位列表响应失败")
	}

	// 转换为单位信息列表
	uomList := make([]*item.UomInfo, 0)
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

	return uomList, nil
}

// SaveUom 保存单位信息
// 如果单位已存在则更新，否则创建新单位
func (s *sStock) SaveUom(ctx context.Context, req *item.UomInfo) error {
	// 获取公司名称
	companyName, err := s.getCompanyName(ctx, req.CompanyAbbr)
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

// getCompanyName 根据公司简称获取公司名称
func (s *sStock) getCompanyName(ctx context.Context, companyAbbr string) (string, error) {
	if len(companyAbbr) == 0 {
		return "", nil
	}

	company, err := service.Company().GetCompanyWithAbbr(ctx, companyAbbr)
	if err != nil {
		g.Log().Error(ctx, "根据公司缩写查询公司失败", err)
		return "", gerror.Wrapf(err, "根据公司缩写查询公司失败")
	}

	return company.CompanyName, nil
}

// checkUomExists 检查单位是否已存在
func (s *sStock) checkUomExists(ctx context.Context, uomName string) (bool, error) {
	if len(uomName) == 0 {
		return false, nil
	}

	filters := [][]string{{"uom_name", "=", uomName}}

	count, err := service.Doctype().Count(ctx, &dto.ErpReq{
		DocType: "UOM",
	}, &dto.RequestParams{
		Filters: filters,
	})

	if err != nil {
		return false, gerror.Wrapf(err, "查询现有单位失败")
	}

	return count > 0, nil
}

// updateExistingUom 更新现有单位
func (s *sStock) updateExistingUom(ctx context.Context, req *item.UomInfo, companyName string) error {
	_, err := service.Document().Update(ctx, &dto.ErpReq{
		DocType: "UOM",
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
func (s *sStock) createNewUom(ctx context.Context, req *item.UomInfo, companyName string) error {
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

// GetAttributeList 获取属性列表
// 根据查询条件过滤并返回属性信息列表
func (s *sStock) GetAttributeList(ctx context.Context, req *item.GetAttributeListReq) (res *item.GetAttributeListResp, err error) {
	// 构建查询过滤器
	filters := s.buildAttributeListFilters(ctx, req)

	// 查询属性列表
	attributeList, err := s.queryAttributeList(ctx, filters)
	if err != nil {
		return nil, gerror.Wrapf(err, "查询属性列表失败")
	}

	return &item.GetAttributeListResp{
		AttributeList: attributeList,
	}, nil
}

// buildAttributeListFilters 构建属性列表查询过滤器
func (s *sStock) buildAttributeListFilters(ctx context.Context, req *item.GetAttributeListReq) [][]string {
	filters := make([][]string, 0)

	// 按分支机构过滤
	if len(req.Branch) > 0 {
		filters = append(filters, g.ArrayStr{"custom_branch", "like", "%" + req.Branch + "%"})
	}

	// 按属性名称过滤
	if len(req.AttributeName) > 0 {
		filters = append(filters, g.ArrayStr{"attribute_name", "like", "%" + req.AttributeName + "%"})
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

	return filters
}

// queryAttributeList 执行属性列表查询
func (s *sStock) queryAttributeList(ctx context.Context, filters [][]string) ([]*item.AttributeInfo, error) {
	resp, err := service.Document().List(ctx, &dto.ErpReq{
		DocType: "Item Attribute",
	}, &dto.RequestParams{
		Fields:  g.ArrayStr{"name", "attribute_name", "custom_alias", "custom_company", "custom_branch"},
		Filters: filters,
		Limit:   consts.Limit999,
	})

	if err != nil {
		return nil, err
	}

	// 解析响应数据
	j, err := gjson.DecodeToJson(resp.Bytes())
	if err != nil {
		return nil, gerror.Wrapf(err, "解析属性列表响应失败")
	}

	// 转换为属性信息列表
	attributeList := make([]*item.AttributeInfo, 0)
	dataArray := j.GetJsons("data")

	for _, attribute := range dataArray {
		attributeInfo := &item.AttributeInfo{
			AttributeName: attribute.Get("name").String(),
			AliasName:     attribute.Get("custom_alias").String(),
			Company:       attribute.Get("custom_company").String(),
			Branch:        attribute.Get("custom_branch").String(),
		}

		// 获取属性值列表
		if values, err := s.GetAttributeValuesList(ctx, attributeInfo.AttributeName); err == nil {
			attributeInfo.AttributeValueList = values
		}

		attributeList = append(attributeList, attributeInfo)
	}

	return attributeList, nil
}

// GetAttributeValuesList 获取属性值列表
// 根据属性名称查询对应的属性值列表
func (s *sStock) GetAttributeValuesList(ctx context.Context, attributeName string) ([]*item.AttributeValueInfo, error) {
	if len(attributeName) == 0 {
		return make([]*item.AttributeValueInfo, 0), nil
	}

	resp, err := service.Document().Get(ctx, &dto.ErpReq{
		DocType: "Item Attribute",
		Name:    attributeName,
	}, nil)

	if err != nil {
		return nil, gerror.Wrapf(err, "查询属性值失败: %s", attributeName)
	}

	// 解析响应数据
	j, err := gjson.DecodeToJson(resp.Bytes())
	if err != nil {
		return nil, gerror.Wrapf(err, "解析属性值响应失败")
	}

	// 转换为属性值信息列表
	attributeValueList := make([]*item.AttributeValueInfo, 0)
	dataArray := j.GetJsons("data.item_attribute_values")

	for _, attributeValue := range dataArray {
		valueInfo := &item.AttributeValueInfo{
			AttributeValue: attributeValue.Get("attribute_value").String(),
			Abbr:           attributeValue.Get("abbr").String(),
		}
		attributeValueList = append(attributeValueList, valueInfo)
	}

	return attributeValueList, nil
}

// SaveAttribute 保存属性信息
// 如果属性已存在则更新，否则创建新属性
func (s *sStock) SaveAttribute(ctx context.Context, req *item.AttributeInfo) error {
	// 获取公司名称
	companyName, err := s.getCompanyName(ctx, req.CompanyAbbr)
	if err != nil {
		return err
	}

	// 检查属性是否已存在
	exists, err := s.checkAttributeExists(ctx, req.AttributeName)
	if err != nil {
		return err
	}

	if exists {
		// 更新现有属性
		return s.updateExistingAttribute(ctx, req, companyName)
	} else {
		// 创建新属性
		return s.createNewAttribute(ctx, req, companyName)
	}
}

// checkAttributeExists 检查属性是否已存在
func (s *sStock) checkAttributeExists(ctx context.Context, attributeName string) (bool, error) {
	if len(attributeName) == 0 {
		return false, nil
	}

	filters := [][]string{{"attribute_name", "=", attributeName}}

	count, err := service.Doctype().Count(ctx, &dto.ErpReq{
		DocType: "Item Attribute",
	}, &dto.RequestParams{
		Filters: filters,
	})

	if err != nil {
		return false, gerror.Wrapf(err, "查询现有属性失败")
	}

	return count > 0, nil
}

// updateExistingAttribute 更新现有属性
func (s *sStock) updateExistingAttribute(ctx context.Context, req *item.AttributeInfo, companyName string) error {
	_, err := service.Document().Update(ctx, &dto.ErpReq{
		DocType: "Item Attribute",
		Name:    req.AttributeName,
	}, &g.Map{
		"custom_alias":   req.AliasName,
		"custom_company": companyName,
		"custom_branch":  req.Branch,
	})

	if err != nil {
		return gerror.Wrapf(err, "更新属性信息失败")
	}

	return nil
}

// createNewAttribute 创建新属性
func (s *sStock) createNewAttribute(ctx context.Context, req *item.AttributeInfo, companyName string) error {
	_, err := service.Document().Create(ctx, "Item Attribute", &g.Map{
		"attribute_name": req.AttributeName,
		"custom_alias":   req.AliasName,
		"custom_company": companyName,
		"custom_branch":  req.Branch,
	})

	if err != nil {
		return gerror.Wrapf(err, "创建属性失败")
	}

	return nil
}
