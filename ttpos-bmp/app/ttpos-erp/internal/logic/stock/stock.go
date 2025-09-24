package stock

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/api/item"
	"ttpos-bmp/app/ttpos-erp/api/stock"
	"ttpos-bmp/app/ttpos-erp/internal/consts"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto/erp"
	"ttpos-bmp/app/ttpos-erp/internal/service"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"
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
	uomList, err := s.queryUomList(ctx, filters, req)
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
func (s *sStock) queryUomList(ctx context.Context, filters [][]string, req *item.GetUomListReq) ([]*item.UomInfo, error) {
	resp, err := service.Document().List(ctx, &erp.ErpReq{
		DocType: "UOM",
	}, &erp.RequestParams{
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

// SaveUom 保存单位信息
// 如果单位已存在则更新，否则创建新单位
func (s *sStock) SaveUom(ctx context.Context, req *item.UomInfo) error {
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
func (s *sStock) checkUomExists(ctx context.Context, uomName string) (bool, error) {
	if len(uomName) == 0 {
		return false, nil
	}

	filters := [][]string{{"uom_name", "=", uomName}}

	count, err := service.Doctype().Count(ctx, &erp.ErpReq{
		DocType: "UOM",
	}, &erp.RequestParams{
		Filters: filters,
	})

	if err != nil {
		return false, gerror.Wrapf(err, "查询现有单位失败")
	}

	return count > 0, nil
}

// updateExistingUom 更新现有单位
func (s *sStock) updateExistingUom(ctx context.Context, req *item.UomInfo, companyName string) error {
	_, err := service.Document().Update(ctx, &erp.ErpReq{
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
	attributeList, err := s.queryAttributeList(ctx, filters, req)
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
func (s *sStock) queryAttributeList(ctx context.Context, filters [][]string, req *item.GetAttributeListReq) ([]*item.AttributeInfo, error) {
	resp, err := service.Document().List(ctx, &erp.ErpReq{
		DocType: erp.DocTypeItemAttribute,
	}, &erp.RequestParams{
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

	subCompanyName := ""
	if len(req.SubCompanyAbbr) > 0 {
		subCompanyName, err = service.Company().GetCompanyNameWithAbbr(ctx, req.SubCompanyAbbr)
		if err != nil {
			return nil, gerror.Wrapf(err, "获取子公司名称失败,companyAbbr:%s", req.SubCompanyAbbr)
		}
	}

	for _, attribute := range dataArray {
		// 查询attribute中的 permission_rule
		if len(req.SubCompanyAbbr) > 0 {
			attr, err := s.GetItemAttribute(ctx, attribute.Get("name").String())
			if err != nil {
				g.Log().Errorf(ctx, "获取物品属性失败,attribute:%s,err:%v", attribute.Get("name").String(), err)
				continue // 获取物品属性失败，跳过
			}
			hasPermission, err := service.Permission().CheckPermission(ctx, attr.CustomPermissionRule, subCompanyName)
			if err != nil {
				g.Log().Errorf(ctx, "检查物品权限失败,attribute:%s,err:%v", attribute.Get("name").String(), err)
				continue // 检查权限失败，跳过
			}
			if !hasPermission {
				continue // 当前公司无权限，跳过该
			}
		}
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

	resp, err := service.Document().Get(ctx, &erp.ErpReq{
		DocType: erp.DocTypeItemAttribute,
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
	companyName, err := service.Company().GetCompanyNameWithAbbr(ctx, req.CompanyAbbr)
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

	count, err := service.Doctype().Count(ctx, &erp.ErpReq{
		DocType: erp.DocTypeItemAttribute,
	}, &erp.RequestParams{
		Filters: filters,
	})

	if err != nil {
		return false, gerror.Wrapf(err, "查询现有属性失败")
	}

	return count > 0, nil
}

// updateExistingAttribute 更新现有属性
func (s *sStock) updateExistingAttribute(ctx context.Context, req *item.AttributeInfo, companyName string) error {

	itemAttribute := &erp.ItemAttribute{
		CustomAlias:         req.AliasName,
		CustomCompany:       companyName,
		CustomBranch:        req.Branch,
		ItemAttributeValues: make([]erp.ItemAttributeValue, 0),
	}
	for _, value := range req.AttributeValueList {
		itemAttribute.ItemAttributeValues = append(itemAttribute.ItemAttributeValues, erp.ItemAttributeValue{
			AttributeValue: value.AttributeValue,
			Abbr:           value.Abbr,
		})
	}
	_, err := service.Document().Update(ctx, &erp.ErpReq{
		DocType: erp.DocTypeItemAttribute,
		Name:    req.AttributeName,
	}, itemAttribute)

	if err != nil {
		return gerror.Wrapf(err, "更新属性信息失败")
	}

	return nil
}

// createNewAttribute 创建新属性
func (s *sStock) createNewAttribute(ctx context.Context, req *item.AttributeInfo, companyName string) error {
	itemAttribute := &erp.ItemAttribute{
		AttributeName:       req.AttributeName,
		CustomAlias:         req.AliasName,
		CustomCompany:       companyName,
		CustomBranch:        req.Branch,
		ItemAttributeValues: make([]erp.ItemAttributeValue, 0),
	}
	for _, value := range req.AttributeValueList {
		itemAttribute.ItemAttributeValues = append(itemAttribute.ItemAttributeValues, erp.ItemAttributeValue{
			AttributeValue: value.AttributeValue,
			Abbr:           value.Abbr,
		})
	}
	_, err := service.Document().Create(ctx, erp.DocTypeItemAttribute, itemAttribute)

	if err != nil {
		return gerror.Wrapf(err, "创建属性失败")
	}

	return nil
}

func (s *sStock) CreateMaterialRequest(ctx context.Context, req *stock.SaveMaterialRequestReq) (res *stock.SaveMaterialRequestResp, err error) {
	// 获取公司名称
	companyName, err := service.Company().GetCompanyWithAbbr(ctx, req.CompanyAbbr)
	if err != nil {
		return nil, err
	}
	//设置默认到货仓
	var (
		targetWarehouseName string
	)
	if req.TargetWarehouse == "" {
		warehouse, err := service.Warehouse().GetDefaultWarehouse(ctx, companyName.CompanyName, req.Branch)
		if err != nil {
			return nil, err
		}
		targetWarehouseName = warehouse.Name
	} else {
		targetWarehouseName = req.TargetWarehouse
	}

	data := g.MapStrAny{
		"naming_series":    erp.DefaultMaterialRequestSeries,
		"transaction_date": service.Setup().MustGetLocalDateTime(ctx, gtime.New(req.TransactionDate)).Format("Y-m-d"),
		"company":          companyName.CompanyName,
	}

	if len(req.Purpose) > 0 {
		data["material_request_type"] = req.Purpose
	} else {
		data["material_request_type"] = erp.StockEntryTypePurchase
	}

	itemList := make([]g.MapStrAny, 0)
	for _, item := range req.Items {
		itemList = append(itemList, g.MapStrAny{
			"item_code":     item.ItemCode,
			"qty":           item.Qty,
			"uom":           item.Uom,
			"schedule_date": service.Setup().MustGetLocalDateTime(ctx, gtime.New(req.RequiredBy)).Format("Y-m-d"),
			"warehouse":     targetWarehouseName,
		})
	}
	data["items"] = itemList
	resp, err := service.Document().Create(ctx, erp.DocTypeMaterialRequest, &data)
	if err != nil {
		return nil, gerror.Wrapf(err, "创建物料请求失败")
	}
	j, err := gjson.DecodeToJson(resp.Bytes())
	if err != nil {
		return nil, err
	}
	if j.Contains("data") {
		//创建后提交
		//提交采购订单
		_, err = service.Document().ChangeDocStatus(ctx, erp.DocTypeMaterialRequest, j.Get("data.name").String(), erp.DocstatusSubmitted)
		if err != nil {
			return nil, gerror.Wrapf(err, "提交物料请求失败")
		}

		res = &stock.SaveMaterialRequestResp{
			MaterialRequestName: j.Get("data.name").String(),
		}
	}
	return
}

// GetMaterialRequestList 获取物料请求列表
func (s *sStock) GetMaterialRequestList(ctx context.Context, req *stock.GetMaterialRequestListReq) (res *stock.GetMaterialRequestListResp, err error) {
	// 获取公司名称
	companyName, err := service.Company().GetCompanyWithAbbr(ctx, req.CompanyAbbr)
	if err != nil {
		return nil, err
	}

	filters := make([][]string, 0)
	if len(req.Branch) > 0 {
		filters = append(filters, []string{"branch", "=", req.Branch})
	}
	filters = append(filters, []string{"company", "=", companyName.CompanyName})
	filters = append(filters, []string{"material_request_type", "=", erp.StockEntryTypePurchase})

	resp, err := service.Document().List(ctx, &erp.ErpReq{
		DocType: "Material Request",
	}, &erp.RequestParams{
		Fields:  []string{"name", "transaction_date", "schedule_date", "material_request_type", "status", "company"},
		Filters: filters,
		Limit:   consts.Limit999,
	})
	if err != nil {
		return nil, gerror.Wrapf(err, "查询物料请求列表失败")
	}
	j, err := gjson.DecodeToJson(resp.Bytes())
	if err != nil {
		return nil, err
	}
	res = &stock.GetMaterialRequestListResp{
		MaterialRequestList: make([]*stock.MaterialRequest, 0),
	}
	for _, item := range j.GetJsons("data") {
		res.MaterialRequestList = append(res.MaterialRequestList, &stock.MaterialRequest{
			MaterialRequestName: item.Get("name").String(),
			Status:              item.Get("status").String(),
			TransactionDate:     gtime.New(item.Get("transaction_date").String()).Unix(),
			RequiredBy:          gtime.New(item.Get("schedule_date").String()).Unix(),
			Company:             item.Get("company").String(),
		})
	}
	return
}

// GetUom 根据单位名称获取单个单位详细信息
// 参数：ctx 上下文，req 包含单位名称
// 返回：单位详细信息，错误信息
func (s *sStock) GetUom(ctx context.Context, req *item.GetUomReq) (res *erp.UOM, err error) {
	// 参数验证
	if len(req.UomName) == 0 {
		return nil, gerror.New("单位名称不能为空")
	}

	// 查询单位信息
	resp, err := service.Document().Get(ctx, &erp.ErpReq{
		DocType: "UOM",
		Name:    req.UomName,
	}, nil)

	if err != nil {
		return nil, gerror.Wrapf(err, "查询单位信息失败")
	}

	// 解析响应数据
	j, err := gjson.DecodeToJson(resp.Bytes())
	if err != nil {
		return nil, gerror.Wrapf(err, "解析单位信息响应失败")
	}

	// 转换为单位信息结构体
	uomInfo := &erp.UOM{}
	gconv.Structs(j.GetJson("data"), &uomInfo)

	return uomInfo, nil
}

// GetItemAttribute 根据属性名称获取单个属性详细信息
// 参数：ctx 上下文，attributeName 属性名称
// 返回：属性详细信息，错误信息
func (s *sStock) GetItemAttribute(ctx context.Context, attributeName string) (res *erp.ItemAttribute, err error) {
	// 参数验证
	if len(attributeName) == 0 {
		return nil, gerror.New("属性名称不能为空")
	}

	// 查询属性信息
	resp, err := service.Document().Get(ctx, &erp.ErpReq{
		DocType: erp.DocTypeItemAttribute,
		Name:    attributeName,
	}, nil)

	if err != nil {
		return nil, gerror.Wrapf(err, "查询属性信息失败")
	}

	// 解析响应数据
	j, err := gjson.DecodeToJson(resp.Bytes())
	if err != nil {
		return nil, gerror.Wrapf(err, "解析属性信息响应失败")
	}

	// 转换为属性信息结构体
	itemAttribute := &erp.ItemAttribute{}
	gconv.Structs(j.GetJson("data"), &itemAttribute)

	return itemAttribute, nil
}
