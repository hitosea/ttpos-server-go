package stock

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/api/company"
	"ttpos-bmp/app/ttpos-erp/api/item"
	"ttpos-bmp/app/ttpos-erp/internal/consts"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto/erp"
	"ttpos-bmp/app/ttpos-erp/internal/service"
	"ttpos-bmp/app/ttpos-erp/utility"

	"github.com/gogf/gf/v2/container/gmap"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
)

var (
	ItemGroup = new(sItemGroup)
)

type sItemGroup struct{}

func init() {
	service.RegisterItemGroup(ItemGroup)
}

// GetItemGroupList 获取物品分组列表
// 根据查询条件过滤并返回物品分组信息列表
func (s *sItemGroup) GetItemGroupList(ctx context.Context, req *item.GetItemGroupListReq) (res *item.GetItemGroupListResp, err error) {
	// 构建查询过滤器
	filters := s.buildItemGroupListFilters(ctx, req)

	// 查询物品分组列表
	itemGroupList, err := s.queryItemGroupList(ctx, filters, req)
	if err != nil {
		return nil, gerror.Wrapf(err, "查询物品分组列表失败")
	}

	return &item.GetItemGroupListResp{
		ItemGroupList: itemGroupList,
	}, nil
}

// buildItemGroupListFilters 构建物品分组列表查询过滤器
func (s *sItemGroup) buildItemGroupListFilters(ctx context.Context, req *item.GetItemGroupListReq) [][]string {
	filters := make([][]string, 0, 4) // 预分配容量，提高性能

	// 按分支机构过滤
	if len(req.Branch) > 0 {
		filters = append(filters, g.ArrayStr{"custom_branch", "=", req.Branch})
	}

	// 按父级分组过滤
	if len(req.ParentItemGroup) > 0 {
		filters = append(filters, g.ArrayStr{"parent_item_group", "=", req.ParentItemGroup})
	}

	// 按公司简称过滤
	if len(req.CompanyAbbr) > 0 {
		companyName, err := service.Company().GetCompanyNameWithAbbr(ctx, req.CompanyAbbr)
		if err == nil && len(companyName) > 0 {
			filters = append(filters, []string{"custom_company", "=", companyName})
		}
	}

	return filters
}

// queryItemGroupList 执行物品分组列表查询
func (s *sItemGroup) queryItemGroupList(ctx context.Context, filters [][]string, req *item.GetItemGroupListReq) ([]*item.ItemGroupInfo, error) {
	resp, err := service.Document().List(ctx, &erp.ErpReq{
		DocType: erp.DocTypeItemGroup,
	}, &erp.RequestParams{
		Fields:  g.ArrayStr{"item_group_name", "parent_item_group", "is_group", "custom_branch", "custom_company", "custom_aliasname"},
		Filters: filters,
		Limit:   consts.Limit9999,
	})

	if err != nil {
		return nil, err
	}

	// 解析响应数据
	j := resp

	// 转换为物品分组信息列表
	dataArray := j.GetJsons("data")
	itemGroupList := make([]*item.ItemGroupInfo, 0, len(dataArray))

	for _, data := range dataArray {
		itemGroupList = append(itemGroupList, &item.ItemGroupInfo{
			ItemGroupName:   data.Get("item_group_name").String(),
			ParentItemGroup: data.Get("parent_item_group").String(),
			IsGroup:         data.Get("is_group").Bool(),
			Branch:          data.Get("custom_branch").String(),
			CompanyAbbr:     req.CompanyAbbr,
			AliasName:       data.Get("custom_aliasname").String(),
		})
	}

	return itemGroupList, nil
}

// GetItemGroup 根据分组代码获取单个物品分组信息
// 参数：ctx 上下文，req 包含分组代码的请求
// 返回：物品分组详细信息，错误信息
func (s *sItemGroup) GetItemGroup(ctx context.Context, req *item.GetItemGroupReq) (res *erp.ItemGroupInfo, err error) {
	// 参数验证
	if len(req.GroupCode) == 0 {
		return nil, gerror.New("分组代码不能为空")
	}

	// 查询物品分组信息
	resp, err := service.Document().Get(ctx, &erp.ErpReq{
		DocType: erp.DocTypeItemGroup,
		Name:    req.GroupCode,
	}, nil)

	if err != nil {
		return nil, gerror.Wrapf(err, "查询物品分组信息失败")
	}

	// 解析响应数据
	j := resp

	itemGroupInfo := &erp.ItemGroupInfo{}
	if err := gconv.Structs(j.GetJson("data"), &itemGroupInfo); err != nil {
		return nil, gerror.Wrapf(err, "转换物品分组信息失败")
	}

	return itemGroupInfo, nil
}

// SaveItemGroup 保存物品分组信息
// 如果物品分组已存在则更新，否则创建新物品分组
func (s *sItemGroup) SaveItemGroup(ctx context.Context, req *item.SaveItemGroupReq) (res *erp.ItemGroupInfo, err error) {
	// 参数验证
	if req.ItemGroupInfo == nil {
		return nil, gerror.New("物品分组信息不能为空")
	}
	if len(req.ItemGroupInfo.ItemGroupName) == 0 {
		return nil, gerror.New("分组名称不能为空")
	}

	// 复制请求参数，避免修改原始数据
	itemGroupInfo := &erp.ItemGroupInfo{}
	if err := gconv.Struct(req.ItemGroupInfo, itemGroupInfo); err != nil {
		return nil, gerror.Wrapf(err, "复制请求参数失败")
	}

	// 检查物品分组是否已存在
	exists, err := s.checkItemGroupExists(ctx, req.ItemGroupInfo.ItemGroupName)
	if err != nil {
		return nil, err
	}

	if exists {
		// 更新现有物品分组
		return s.updateExistingItemGroup(ctx, itemGroupInfo)
	} else {
		// 创建新物品分组
		return s.createNewItemGroup(ctx, itemGroupInfo, req.ItemGroupInfo)
	}
}

// checkItemGroupExists 检查物品分组是否已存在
func (s *sItemGroup) checkItemGroupExists(ctx context.Context, groupName string) (bool, error) {
	if len(groupName) == 0 {
		return false, nil
	}

	filters := [][]string{{"item_group_name", "=", groupName}}

	count, err := service.Doctype().Count(ctx, &erp.ErpReq{
		DocType: erp.DocTypeItemGroup,
	}, &erp.RequestParams{
		Filters: filters,
	})

	if err != nil {
		return false, gerror.Wrapf(err, "查询现有物品分组失败")
	}

	return count > 0, nil
}

// updateExistingItemGroup 更新现有物品分组
func (s *sItemGroup) updateExistingItemGroup(ctx context.Context, itemGroupInfo *erp.ItemGroupInfo) (*erp.ItemGroupInfo, error) {
	// 构建更新数据
	itemGroupForUpdate := s.buildUpdateItemGroupData(itemGroupInfo)

	// 执行更新操作
	_, err := service.Document().Update(ctx, &erp.ErpReq{
		DocType: erp.DocTypeItemGroup,
		Name:    itemGroupInfo.ItemGroupName,
	}, &itemGroupForUpdate)

	if err != nil {
		return nil, gerror.Wrapf(err, "更新物品分组信息失败")
	}

	// 返回更新后的物品分组信息
	return itemGroupInfo, nil
}

// buildUpdateItemGroupData 构建更新物品分组数据
func (s *sItemGroup) buildUpdateItemGroupData(itemGroupInfo *erp.ItemGroupInfo) g.Map {
	itemGroupForUpdate := g.Map{}

	// 基本信息更新
	if len(itemGroupInfo.ItemGroupName) > 0 {
		itemGroupForUpdate["item_group_name"] = itemGroupInfo.ItemGroupName
	}
	if len(itemGroupInfo.ParentItemGroup) > 0 {
		itemGroupForUpdate["parent_item_group"] = itemGroupInfo.ParentItemGroup
	}
	if len(itemGroupInfo.Branch) > 0 {
		itemGroupForUpdate["custom_branch"] = itemGroupInfo.Branch
	}
	if len(itemGroupInfo.Company) > 0 {
		itemGroupForUpdate["custom_company"] = itemGroupInfo.Company
	}

	// 是否为分组
	itemGroupForUpdate["is_group"] = itemGroupInfo.IsGroup

	itemGroupForUpdate["custom_aliasname"] = itemGroupInfo.AliasName

	return itemGroupForUpdate
}

// createNewItemGroup 创建新物品分组
func (s *sItemGroup) createNewItemGroup(ctx context.Context, itemGroupInfo *erp.ItemGroupInfo, reqInfo *item.ItemGroupInfo) (*erp.ItemGroupInfo, error) {
	// 获取公司信息
	var company *company.CompanyInfo
	var err error
	if len(reqInfo.CompanyAbbr) > 0 {
		company, err = s.getCompanyInfo(ctx, reqInfo.CompanyAbbr)
		if err != nil {
			return nil, err
		}
	}

	// 构建新物品分组数据
	newItemGroup, err := s.buildNewItemGroupData(ctx, itemGroupInfo, reqInfo, company)
	if err != nil {
		return nil, err
	}

	// 创建物品分组
	_, err = service.Document().Create(ctx, erp.DocTypeItemGroup, &newItemGroup)
	if err != nil {
		return nil, gerror.Wrapf(err, "创建物品分组失败")
	}

	// 转换并返回结果
	return s.buildCreateItemGroupResponse(itemGroupInfo, reqInfo, company, newItemGroup), nil
}

// getCompanyInfo 获取公司信息
func (s *sItemGroup) getCompanyInfo(ctx context.Context, companyAbbr string) (*company.CompanyInfo, error) {
	company, err := service.Company().GetCompanyWithAbbr(ctx, companyAbbr)
	if err != nil {
		return nil, gerror.Wrapf(err, "获取公司信息失败")
	}
	return company, nil
}

// buildNewItemGroupData 构建新物品分组数据
func (s *sItemGroup) buildNewItemGroupData(ctx context.Context, itemGroupInfo *erp.ItemGroupInfo, reqInfo *item.ItemGroupInfo, company *company.CompanyInfo) (g.Map, error) {
	// 基础数据
	newItemGroup := g.Map{
		"item_group_name": itemGroupInfo.ItemGroupName,
		"is_group":        itemGroupInfo.IsGroup,
	}

	// 设置父级分组
	if len(itemGroupInfo.ParentItemGroup) > 0 {
		newItemGroup["parent_item_group"] = itemGroupInfo.ParentItemGroup
	}

	// 设置分支
	if len(reqInfo.Branch) > 0 {
		newItemGroup["custom_branch"] = reqInfo.Branch
	}

	// 设置公司信息
	if company != nil {
		newItemGroup["custom_company"] = company.CompanyName
	}
	//设置别名
	newItemGroup["custom_aliasname"] = reqInfo.AliasName

	return newItemGroup, nil
}

// buildCreateItemGroupResponse 构建创建物品分组响应
func (s *sItemGroup) buildCreateItemGroupResponse(itemGroupInfo *erp.ItemGroupInfo, reqInfo *item.ItemGroupInfo, company *company.CompanyInfo, newItemGroup g.Map) *erp.ItemGroupInfo {
	res := &erp.ItemGroupInfo{
		ItemGroupName:   itemGroupInfo.ItemGroupName,
		ParentItemGroup: itemGroupInfo.ParentItemGroup,
		IsGroup:         itemGroupInfo.IsGroup,
		Branch:          reqInfo.Branch,
	}

	// 设置公司信息
	if company != nil {
		res.Company = company.CompanyName
	}

	// 尝试扫描新物品分组数据到响应结构
	if err := gconv.Scan(newItemGroup, res); err != nil {
		// 如果扫描失败，至少返回基本信息
		if groupName, ok := newItemGroup["item_group_name"].(string); ok {
			res.ItemGroupName = groupName
		}
		if isGroup, ok := newItemGroup["is_group"].(bool); ok {
			res.IsGroup = isGroup
		}
	}

	return res
}

// DeleteItemGroup 删除物品分组
// 参数：ctx 上下文，req 删除物品分组请求
// 返回：错误信息
func (s *sItemGroup) DeleteItemGroup(ctx context.Context, req *item.DeleteItemGroupReq) error {
	// 参数验证
	if len(req.ItemGroupName) == 0 {
		return gerror.New("分组名称不能为空")
	}

	// 检查物品分组是否存在
	exists, err := s.checkItemGroupExists(ctx, req.ItemGroupName)
	if err != nil {
		return gerror.Wrapf(err, "检查物品分组是否存在失败")
	}

	if !exists {
		return gerror.New("物品分组不存在")
	}

	// 检查是否有子分组
	hasChildren, err := s.checkItemGroupHasChildren(ctx, req.ItemGroupName)
	if err != nil {
		return gerror.Wrapf(err, "检查子分组失败")
	}

	if hasChildren {
		return gerror.New("该分组下存在子分组，无法删除")
	}

	// 检查是否有关联的物品
	hasItems, err := s.checkItemGroupHasItems(ctx, req.ItemGroupName)
	if err != nil {
		return gerror.Wrapf(err, "检查关联物品失败")
	}

	if hasItems {
		return gerror.New("该分组下存在物品，无法删除")
	}

	// 执行删除操作
	_, err = service.Document().Delete(ctx, &erp.ErpReq{
		DocType: erp.DocTypeItemGroup,
		Name:    req.ItemGroupName,
	})

	if err != nil {
		return gerror.Wrapf(err, "删除物品分组失败")
	}

	return nil
}

// checkItemGroupHasChildren 检查物品分组是否有子分组
func (s *sItemGroup) checkItemGroupHasChildren(ctx context.Context, groupName string) (bool, error) {
	filters := [][]string{{"parent_item_group", "=", groupName}}

	count, err := service.Doctype().Count(ctx, &erp.ErpReq{
		DocType: erp.DocTypeItemGroup,
	}, &erp.RequestParams{
		Filters: filters,
	})

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// checkItemGroupHasItems 检查物品分组是否有关联的物品
func (s *sItemGroup) checkItemGroupHasItems(ctx context.Context, groupName string) (bool, error) {
	filters := [][]string{{"item_group", "=", groupName}}

	count, err := service.Doctype().Count(ctx, &erp.ErpReq{
		DocType: erp.DocTypeItem,
	}, &erp.RequestParams{
		Filters: filters,
	})

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// SaveAttributeGroup 保存物品属性分组
// 参数：ctx 上下文，req 保存物品属性分组请求
// 返回：物品属性分组响应，错误信息
func (s *sItemGroup) SaveAttributeGroup(ctx context.Context, req *item.SaveAttributeGroupReq) (*item.SaveAttributeGroupResp, error) {
	var (
		itemGroupInfo *erp.ItemGroupInfo
		err           error
		attrItemList  []*item.AttributeItemInfo
	)
	attrItemList = make([]*item.AttributeItemInfo, 0)

	//创建分组
	if req.AttributeGroupInfo.GroupName == "" {
		itemGroupInfo, err = s.SaveItemGroup(ctx, &item.SaveItemGroupReq{
			ItemGroupInfo: &item.ItemGroupInfo{
				ItemGroupName:   utility.GenItemCode(consts.ItemGroupPrefixPosAttributeGroup),
				ParentItemGroup: string(consts.ItemGroupPosAttribute),
				Branch:          req.AttributeGroupInfo.Branch,
				AliasName:       req.AttributeGroupInfo.AliasName,
				CompanyAbbr:     req.AttributeGroupInfo.CompanyAbbr,
			},
		})
		if err != nil {
			return nil, gerror.Wrapf(err, "创建物品属性分组失败")
		}
	} else {
		//更新 属性组信息
		itemGroupInfo, err = s.SaveItemGroup(ctx, &item.SaveItemGroupReq{
			ItemGroupInfo: &item.ItemGroupInfo{
				ItemGroupName:   req.AttributeGroupInfo.GroupName,
				ParentItemGroup: string(consts.ItemGroupPosAttribute),
				Branch:          req.AttributeGroupInfo.Branch,
				AliasName:       req.AttributeGroupInfo.AliasName,
				CompanyAbbr:     req.AttributeGroupInfo.CompanyAbbr,
			},
		})
		if err != nil {
			return nil, gerror.Wrapf(err, "更新物品属性分组失败")
		}
		// 将req.AttributeGroupInfo.AttributeItemList  转换成map
		attributeItemMap := gmap.NewHashMap()
		for _, attributeItem := range req.AttributeGroupInfo.AttributeItemList {
			attributeItemMap.Set(attributeItem.ItemCode, attributeItem)
		}

		//查询属性组下的所有属性值
		itemListResp, err := service.Item().GetItemList(ctx, &item.GetItemListReq{
			ItemGroup:     item.ItemGroup_PosAttribute,
			ItemGroupName: req.AttributeGroupInfo.GroupName,
		})
		if err != nil {
			return nil, gerror.Wrapf(err, "查询物品属性值失败")
		}
		for _, itemInfo := range itemListResp.ItemList {
			if !attributeItemMap.Contains(itemInfo.ItemCode) {
				//删除
				_, err := service.Document().Delete(ctx, &erp.ErpReq{
					DocType: erp.DocTypeItem,
					Name:    itemInfo.ItemCode,
				})
				if err != nil {
					return nil, gerror.Wrapf(err, "删除物品属性值失败")
				}
			}
		}
	}
	//更新属性值
	for _, attrItemInfo := range req.AttributeGroupInfo.AttributeItemList {
		//更新
		respItem, err := service.Item().SavePosAttribute(ctx, &item.SavePosAttributeReq{
			Item: &item.PosSpecItem{
				ItemCode: attrItemInfo.ItemCode,
				ItemName: attrItemInfo.AliasName,
			},
		})
		if err != nil {
			return nil, gerror.Wrapf(err, "更新物品属性值失败")
		}
		//获取生成的 item code
		attrItemList = append(attrItemList, &item.AttributeItemInfo{
			AliasName: attrItemInfo.AliasName,
			ItemCode:  respItem.ItemCode,
		})
	}

	return &item.SaveAttributeGroupResp{
		AttributeGroupInfo: &item.AttributeGroupInfo{
			AliasName:         req.AttributeGroupInfo.AliasName,
			CompanyAbbr:       req.AttributeGroupInfo.CompanyAbbr,
			Branch:            req.AttributeGroupInfo.Branch,
			GroupName:         itemGroupInfo.ItemGroupName,
			AttributeItemList: attrItemList,
		},
	}, nil
}

func (s *sItemGroup) DeleteAttributeGroup(ctx context.Context, req *item.DeleteAttributeGroupReq) (*item.DeleteAttributeGroupReq, error) {
	//删除组下所有商品
	//查询属性组下的所有属性值
	itemListResp, err := service.Item().GetItemList(ctx, &item.GetItemListReq{
		ItemGroup:     item.ItemGroup_PosAttribute,
		ItemGroupName: req.GroupName,
	})
	if err != nil {
		return nil, gerror.Wrapf(err, "查询物品属性值失败")
	}
	for _, itemInfo := range itemListResp.ItemList {
		_, err := service.Document().Delete(ctx, &erp.ErpReq{
			DocType: erp.DocTypeItem,
			Name:    itemInfo.ItemCode,
		})
		if err != nil {
			return nil, gerror.Wrapf(err, "删除物品属性值失败")
		}
	}
	_, err = service.Document().Delete(ctx, &erp.ErpReq{
		DocType: erp.DocTypeItemGroup,
		Name:    req.GroupName,
	})
	if err != nil {
		return nil, gerror.Wrapf(err, "删除物品属性分组失败")
	}
	return &item.DeleteAttributeGroupReq{
		GroupName: req.GroupName,
	}, nil
}

// SaveAddonGroup 保存加料组
// 关联门店时，每个门店都会自动创建一个加料组,
func (s *sItemGroup) SaveAddonGroup(ctx context.Context, req *item.SaveAddonGroupReq) (*item.SaveAddonGroupResp, error) {
	var (
		addonList     = make([]*item.AddonItemInfo, 0)
		itemGroupInfo = &erp.ItemGroupInfo{}
	)
	//获取当前公司，分支加料组
	itemGroupInfoList, err := service.ItemGroup().GetItemGroupList(ctx, &item.GetItemGroupListReq{
		Branch:          req.AddonGroupInfo.Branch,
		CompanyAbbr:     req.AddonGroupInfo.CompanyAbbr,
		ParentItemGroup: string(consts.ItemGroupPosAddon),
	})
	if err != nil {
		return nil, gerror.Wrapf(err, "查询加料分组失败")
	}
	if len(itemGroupInfoList.ItemGroupList) == 0 {
		//return nil, gerror.Wrapf(err, "当前门店加料分组不存在")
		g.Log().Warning(ctx, "当前门店加料分组不存在,自动创建")
		req.AddonGroupInfo.GroupName = utility.GenItemCode(consts.ItemGroupPrefixPosAddonGroup)
		itemGroupInfo, err = service.ItemGroup().SaveItemGroup(ctx, &item.SaveItemGroupReq{
			ItemGroupInfo: &item.ItemGroupInfo{
				ItemGroupName:   req.AddonGroupInfo.GroupName,
				ParentItemGroup: string(consts.ItemGroupPosAddon),
				Branch:          req.AddonGroupInfo.Branch,
				CompanyAbbr:     req.AddonGroupInfo.CompanyAbbr,
			},
		})
		if err != nil {
			return nil, gerror.Wrapf(err, "创建默认加料分组失败")
		}
	} else {
		// 调用服务层保存数据
		itemGroupInfo, err = service.ItemGroup().SaveItemGroup(ctx, &item.SaveItemGroupReq{
			ItemGroupInfo: &item.ItemGroupInfo{
				ItemGroupName:   req.AddonGroupInfo.GroupName,
				ParentItemGroup: string(consts.ItemGroupPosAddon),
				Branch:          req.AddonGroupInfo.Branch,
				CompanyAbbr:     req.AddonGroupInfo.CompanyAbbr,
			},
		})
		if err != nil {
			return nil, gerror.Wrapf(err, "修改加料分组失败")
		}

		// 将req.AttributeGroupInfo.AttributeItemList  转换成map
		addonItemMap := gmap.NewHashMap()
		for _, attributeItem := range req.AddonGroupInfo.AddonItemList {
			addonItemMap.Set(attributeItem.ItemCode, attributeItem)
		}
		//查询加料组下的所有加料值
		itemListResp, err := service.Item().GetItemList(ctx, &item.GetItemListReq{
			ItemGroup:     item.ItemGroup_PosAddon,
			ItemGroupName: req.AddonGroupInfo.GroupName,
		})
		if err != nil {
			return nil, gerror.Wrapf(err, "查询物品属性值失败")
		}
		//删除不要加料
		for _, itemInfo := range itemListResp.ItemList {
			if !addonItemMap.Contains(itemInfo.ItemCode) {
				//删除
				_, err := service.Document().Delete(ctx, &erp.ErpReq{
					DocType: erp.DocTypeItem,
					Name:    itemInfo.ItemCode,
				})
				if err != nil {
					return nil, gerror.Wrapf(err, "删除物品属性值失败")
				}
			}
		}
	}

	for _, addonInfo := range req.AddonGroupInfo.AddonItemList {
		//更新保存还要的
		respAddon, err := service.Item().SavePosAddon(ctx, &item.SavePosAddonReq{
			Item: &item.PosSpecItem{
				ItemName:      addonInfo.AliasName,
				ItemCode:      addonInfo.ItemCode,
				Branch:        itemGroupInfo.Branch,
				CompanyAbbr:   req.AddonGroupInfo.CompanyAbbr,
				ItemGroupName: itemGroupInfo.ItemGroupName,
			},
		})
		if err != nil {
			return nil, gerror.Wrapf(err, "更新物品属性值失败")
		}
		//获取生成的 item code
		addonList = append(addonList, &item.AddonItemInfo{
			AliasName: addonInfo.AliasName,
			ItemCode:  respAddon.ItemCode,
		})
	}

	return &item.SaveAddonGroupResp{
		AddonGroupInfo: &item.AddonGroupInfo{
			CompanyAbbr:   req.AddonGroupInfo.CompanyAbbr,
			Branch:        req.AddonGroupInfo.Branch,
			GroupName:     itemGroupInfo.ItemGroupName,
			AddonItemList: addonList,
		},
	}, nil
}
