package stock

import (
	"context"
	"fmt"
	"time"
	"ttpos-bmp/app/ttpos-erp/api/company"
	"ttpos-bmp/app/ttpos-erp/api/item"
	"ttpos-bmp/app/ttpos-erp/internal/consts"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto/erp"
	"ttpos-bmp/app/ttpos-erp/internal/service"
	"ttpos-bmp/app/ttpos-erp/utility"
	"ttpos-bmp/internal/pkg/queue"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"
)

var (
	Item = new(sItem)
)

type sItem struct{}

func init() {
	service.RegisterItem(Item)
}

// SyncDelay 同步延迟处理
// 将物品同步任务推送到队列，并设置10秒延迟
func (s *sItem) SyncDelay() {
	const syncDelayDuration = 10 * time.Second

	queue.Push(string(consts.TopicItemSync), &erp.Item{})
	queue.DelayPush(string(consts.TopicItemSyncDelay), &erp.Item{}, syncDelayDuration)
}

// GetItemList 获取物品列表
// 根据查询条件过滤并返回物品信息列表
func (s *sItem) GetItemList(ctx context.Context, req *item.GetItemListReq) (res *item.GetItemListResp, err error) {
	// 构建查询过滤器
	filters := s.buildItemListFilters(ctx, req)

	// 查询物品列表
	itemList, err := s.queryItemList(ctx, filters, req)
	if err != nil {
		return nil, gerror.Wrapf(err, "查询物品列表失败")
	}

	return &item.GetItemListResp{
		ItemList: itemList,
	}, nil
}

// buildItemListFilters 构建物品列表查询过滤器
func (s *sItem) buildItemListFilters(ctx context.Context, req *item.GetItemListReq) [][]string {
	filters := make([][]string, 0, 8) // 预分配容量，提高性能

	// 按分支机构过滤
	if len(req.Branch) > 0 {
		filters = append(filters, g.ArrayStr{"custom_branch", "=", req.Branch})
	}

	// 按物品名称过滤
	if len(req.ItemName) > 0 {
		filters = append(filters, g.ArrayStr{"name", "=", req.ItemName})
	}

	// 按物品分组过滤
	if req.ItemGroup != item.ItemGroup_Others {
		itemGroupStr := utility.ItemGroupToString(req.ItemGroup)
		if len(itemGroupStr) > 0 {
			filters = append(filters, g.ArrayStr{"item_group", "=", itemGroupStr})
		}
		//特殊处理套餐包
		if req.ItemGroup == item.ItemGroup_Package {
			filters = append(filters, []string{"item_code", "like", consts.ItemCodePrefixPackage + "%"})
		}
	}

	// 按物品编码过滤
	if len(req.ItemCode) > 0 {
		filters = append(filters, g.ArrayStr{"item_code", "=", req.ItemCode})
	}

	// 按物品编码前缀过滤
	if len(req.ItemCodePrefix) > 0 {
		filters = append(filters, []string{"item_code", "like", req.ItemCodePrefix + "%"})
	}

	// 按公司简称过滤
	if len(req.CompanyAbbr) > 0 {
		companyName, err := service.Company().GetCompanyNameWithAbbr(ctx, req.CompanyAbbr)
		if err == nil && len(companyName) > 0 {
			filters = append(filters, []string{"custom_company", "=", companyName})
		}
	}

	// 只查询未禁用的物品
	if !req.ContainDisabled {
		filters = append(filters, []string{"disabled", "!=", "1"})
	}
	return filters
}

// queryItemList 执行物品列表查询
func (s *sItem) queryItemList(ctx context.Context, filters [][]string, req *item.GetItemListReq) ([]*item.ItemInfo, error) {
	resp, err := service.Document().List(ctx, &erp.ErpReq{
		DocType: erp.DocTypeItem,
	}, &erp.RequestParams{
		Fields:  g.ArrayStr{"item_name", "item_code", "item_group", "custom_branch", "disabled", "custom_company", "custom_specification", "stock_uom"},
		Filters: filters,
		Limit:   consts.Limit9999,
	})

	if err != nil {
		return nil, err
	}

	// 解析响应数据
	j, err := gjson.DecodeToJson(resp.Bytes())
	if err != nil {
		return nil, gerror.Wrapf(err, "解析物品列表响应失败")
	}

	// 转换为物品信息列表
	dataArray := j.GetJsons("data")
	itemList := make([]*item.ItemInfo, 0, len(dataArray))
	subCompanyName := ""
	if len(req.SubCompanyAbbr) > 0 {
		subCompanyName, err = service.Company().GetCompanyNameWithAbbr(ctx, req.SubCompanyAbbr)
		if err != nil {
			return nil, gerror.Wrapf(err, "获取子公司名称失败,companyAbbr:%s", req.SubCompanyAbbr)
		}
	}

	for _, data := range dataArray {
		//获取当前item 的所有 Item Permission List 判断是否有权限使用, 列表查询目前不支持 表格字段查询，只能每次遍历物品查询 对应的 custom_permission_rule
		if len(req.SubCompanyAbbr) > 0 {
			itemInfo, err := s.GetItem(ctx, &item.GetItemReq{
				ItemCode: data.Get("item_code").String(),
			})
			if err != nil {
				g.Log().Errorf(ctx, "获取物品信息失败,itemCode:%s,err:%v", data.Get("item_code").String(), err)
				continue // 获取物品信息失败，跳过该物品
			}
			hasPermission, err := service.Permission().CheckPermission(ctx, itemInfo.CustomPermissionRule, subCompanyName)
			if err != nil {
				g.Log().Errorf(ctx, "检查物品权限失败,itemCode:%s,err:%v", data.Get("item_code").String(), err)
				continue // 检查权限失败，跳过该物品
			}
			if !hasPermission {
				continue // 当前公司无权限，跳过该物品
			}
		}
		itemList = append(itemList, &item.ItemInfo{
			Branch:    data.Get("custom_branch").String(),
			Company:   data.Get("custom_company").String(),
			ItemName:  data.Get("item_name").String(),
			ItemCode:  data.Get("item_code").String(),
			ItemGroup: utility.ParseItemGroupFromString(data.Get("item_group").String()),
			StockUom:  data.Get("stock_uom").String(),
			Disabled:  data.Get("disabled").Bool(),
		})
	}

	return itemList, nil
}

// SaveItem 保存物品信息
// 如果物品已存在则更新，否则创建新物品
func (s *sItem) SaveItem(ctx context.Context, reqInfo *item.ItemInfo) (res *item.ItemInfo, err error) {
	// 复制请求参数，避免修改原始数据
	req := &item.ItemInfo{}
	if err := gconv.Struct(reqInfo, req); err != nil {
		return nil, gerror.Wrapf(err, "复制请求参数失败")
	}

	// 检查物品是否已存在
	exists, err := s.checkItemExists(ctx, req.ItemCode)
	if err != nil {
		return nil, err
	}

	if exists {
		// 更新现有物品
		return s.updateExistingItem(ctx, req)
	} else {
		// 创建新物品
		return s.createNewItem(ctx, req)
	}
}

// checkItemExists 检查物品是否已存在
func (s *sItem) checkItemExists(ctx context.Context, itemCode string) (bool, error) {
	if len(itemCode) == 0 {
		return false, nil
	}

	filters := [][]string{{"item_code", "=", itemCode}}

	count, err := service.Doctype().Count(ctx, &erp.ErpReq{
		DocType: "Item",
	}, &erp.RequestParams{
		Filters: filters,
	})

	if err != nil {
		return false, gerror.Wrapf(err, "查询现有物品失败")
	}

	return count > 0, nil
}

// updateExistingItem 更新现有物品
func (s *sItem) updateExistingItem(ctx context.Context, req *item.ItemInfo) (*item.ItemInfo, error) {
	// 构建更新数据
	itemForUpdate := s.buildUpdateItemData(req)

	// 执行更新操作
	_, err := service.Document().Update(ctx, &erp.ErpReq{
		DocType: "Item",
		Name:    req.ItemCode,
	}, &itemForUpdate)

	if err != nil {
		return nil, gerror.Wrapf(err, "更新物品信息失败")
	}

	// 返回更新后的物品信息
	return req, nil
}

// buildUpdateItemData 构建更新物品数据
func (s *sItem) buildUpdateItemData(req *item.ItemInfo) g.Map {
	itemForUpdate := g.Map{}

	// 基本信息更新
	if len(req.ItemName) > 0 {
		itemForUpdate["item_name"] = req.ItemName
	}
	if len(req.StockUom) > 0 {
		itemForUpdate["stock_uom"] = req.StockUom
	}
	if len(req.ItemSpecification) > 0 {
		itemForUpdate["custom_specification"] = req.ItemSpecification
	}
	//更新估值率 不更新也必须传入此字段，否erp会设置成0
	itemForUpdate["valuation_rate"] = req.ValuationRate

	// 条形码更新
	//note: ttpos 不更新条形码也必须传入此字段，否erp会删除条形码
	if len(req.Barcode) > 0 {
		itemForUpdate["barcodes"] = g.Array{
			g.Map{
				"barcode": req.Barcode,
			},
		}
	} else {
		itemForUpdate["barcodes"] = g.Array{}
	}

	// 转换单位更新
	if len(req.Uoms) > 0 {
		uoms := make([]g.Map, 0, len(req.Uoms))
		for _, uom := range req.Uoms {
			uoms = append(uoms, g.Map{
				"uom":               uom.Uom,
				"conversion_factor": uom.ConversionFactor,
			})
		}
		itemForUpdate["uoms"] = uoms
	}

	// 禁用状态更新
	if req.Disabled {
		itemForUpdate["disabled"] = 1
	} else {
		itemForUpdate["disabled"] = 0
	}
	// 更新估值率
	if req.ValuationRate != 0 {
		itemForUpdate["valuation_rate"] = req.ValuationRate
	}

	if req.Classification != "" {
		itemForUpdate["custom_classification"] = req.Classification
	}
	if req.ClassificationCode != "" {
		itemForUpdate["custom_classification_code"] = req.ClassificationCode
	}
	if req.InternalCode != "" {
		itemForUpdate["custom_internal_code"] = req.InternalCode
	}

	return itemForUpdate
}

// createNewItem 创建新物品
func (s *sItem) createNewItem(ctx context.Context, req *item.ItemInfo) (*item.ItemInfo, error) {
	// 获取公司信息
	company, err := s.getCompanyInfo(ctx, req.CompanyAbbr)
	if err != nil {
		return nil, err
	}

	// 生成物品编码
	itemCode, err := s.generateItemCode(ctx, req)
	if err != nil {
		return nil, err
	}

	// 构建新物品数据
	newItem, err := s.buildNewItemData(ctx, req, company, itemCode)
	if err != nil {
		return nil, err
	}

	//创建禁用物品时，erpnext无法直接创建禁用物品。先创建一个启用的物品，再修改为禁用状态
	if req.Disabled {
		newItem["disabled"] = 0
	}

	// 创建物品
	_, err = service.Document().Create(ctx, "Item", &newItem)
	if err != nil {
		return nil, gerror.Wrapf(err, "创建物品失败")
	}

	if req.Disabled {
		newItem["disabled"] = 1
		service.Document().Update(ctx, &erp.ErpReq{
			DocType: erp.DocTypeItem,
			Name:    itemCode,
		}, &newItem)
	}

	// 转换并返回结果
	return s.buildCreateItemResponse(req, company, newItem), nil
}

// getCompanyInfo 获取公司信息
func (s *sItem) getCompanyInfo(ctx context.Context, companyAbbr string) (*company.CompanyInfo, error) {
	company, err := service.Company().GetCompanyWithAbbr(ctx, companyAbbr)
	if err != nil {
		return nil, gerror.Wrapf(err, "获取公司信息失败")
	}
	return company, nil
}

// buildNewItemData 构建新物品数据
func (s *sItem) buildNewItemData(ctx context.Context, req *item.ItemInfo, company *company.CompanyInfo, itemCode string) (g.Map, error) {
	// 基础数据
	newItem := g.Map{
		"item_code":                  itemCode,
		"item_name":                  req.ItemName,
		"stock_uom":                  req.StockUom,
		"item_group":                 utility.ItemGroupToString(req.ItemGroup),
		"custom_branch":              req.Branch,
		"custom_company":             company.CompanyName,
		"custom_classification":      req.Classification,
		"custom_classification_code": req.ClassificationCode,
		"custom_internal_code":       req.InternalCode,
	}

	// 根据物品分组添加特定字段
	s.addItemGroupSpecificFields(ctx, req, newItem, company)

	// 添加条形码
	if len(req.Barcode) > 0 {
		newItem["barcodes"] = g.Array{
			g.Map{
				"barcode": req.Barcode,
			},
		}
	}
	// 禁用状态更新
	if req.Disabled {
		newItem["disabled"] = 1
	} else {
		//启用
		newItem["disabled"] = 0
	}

	return newItem, nil
}

// addItemGroupSpecificFields 根据物品分组添加特定字段
func (s *sItem) addItemGroupSpecificFields(ctx context.Context, req *item.ItemInfo, newItem g.Map, company *company.CompanyInfo) {
	if req.ItemGroup == item.ItemGroup_RawMaterial {
		// 原材料特定字段
		s.addRawMaterialFields(req, newItem)
	} else if req.ItemGroup == item.ItemGroup_Products {
		// 商品特定字段
		newItem["custom_specification"] = req.ItemSpecification
		newItem["is_stock_item"] = 0
	} else if req.ItemGroup == item.ItemGroup_Package {
		// 套餐特定字段
		newItem["is_stock_item"] = 0
	}

	// 设置默认仓库
	s.setDefaultWarehouse(ctx, req, company, newItem)
}

// addRawMaterialFields 添加原材料特定字段
func (s *sItem) addRawMaterialFields(req *item.ItemInfo, newItem g.Map) {
	// 期初库存设置
	if req.OpeningStock > 0 {
		newItem["is_stock_item"] = 1
		newItem["opening_stock"] = req.OpeningStock
		newItem["valuation_rate"] = req.ValuationRate
	}

	// 转换单位设置
	if len(req.Uoms) > 0 {
		uoms := make([]g.Map, 0, len(req.Uoms))
		for _, uom := range req.Uoms {
			uoms = append(uoms, g.Map{
				"uom":               uom.Uom,
				"conversion_factor": uom.ConversionFactor,
			})
		}
		newItem["uoms"] = uoms
	}
}

// setDefaultWarehouse 设置默认仓库
func (s *sItem) setDefaultWarehouse(ctx context.Context, req *item.ItemInfo, company *company.CompanyInfo, newItem g.Map) {
	warehouse, err := service.Warehouse().GetDefaultWarehouse(ctx, company.CompanyName, req.Branch)
	if err == nil {
		newItem["item_defaults"] = g.Array{
			g.Map{
				"company":           company.CompanyName,
				"default_warehouse": warehouse.Name,
			},
		}
	}
}

// buildCreateItemResponse 构建创建物品响应
func (s *sItem) buildCreateItemResponse(req *item.ItemInfo, company *company.CompanyInfo, newItem g.Map) *item.ItemInfo {
	res := &item.ItemInfo{
		Barcode:           req.Barcode,
		Branch:            req.Branch,
		Company:           company.CompanyName,
		CompanyAbbr:       company.CompanyAbbr,
		ItemSpecification: req.ItemSpecification,
	}

	// 尝试扫描新物品数据到响应结构
	if err := gconv.Scan(newItem, res); err != nil {
		// 如果扫描失败，至少返回基本信息
		if itemCode, ok := newItem["item_code"].(string); ok {
			res.ItemCode = itemCode
		}
		if itemName, ok := newItem["item_name"].(string); ok {
			res.ItemName = itemName
		}
		if stockUom, ok := newItem["stock_uom"].(string); ok {
			res.StockUom = stockUom
		}
	}

	return res
}

// generateItemCode 生成物品编码
func (s *sItem) generateItemCode(ctx context.Context, req *item.ItemInfo) (string, error) {
	// 根据物品分组生成编码前缀
	if req.ItemGroup == item.ItemGroup_RawMaterial {
		return utility.GenItemCode(consts.ItemCodePrefixRawMaterial), nil
	}

	// 套餐编码
	if req.ItemGroup == item.ItemGroup_Package {
		return utility.GenItemCode(consts.ItemCodePrefixPackage), nil
	}

	// 商品编码
	itemCode := utility.GenItemCode(consts.ItemCodePrefixProduct)

	// 处理多规格商品编码
	if len(req.ItemSpecification) > 0 {
		suffix, err := s.generateItemCodeWithTemplate(ctx, req.TemplateItemCode)
		if err != nil {
			return "", err
		}
		itemCode = suffix
	} else {
		// 无规格商品，添加默认后缀
		itemCode += "_00"
	}

	return itemCode, nil
}

// generateItemCodeWithTemplate 生成物品编码后缀
func (s *sItem) generateItemCodeWithTemplate(ctx context.Context, templateItemCode string) (string, error) {
	if len(templateItemCode) == 0 {
		return "", gerror.New("模板物品编码不能为空")
	}

	// 查找最后一个下划线位置
	lastUnderscorePos := gstr.PosR(templateItemCode, "_")
	if lastUnderscorePos == -1 {
		return "", gerror.New("模板物品编码格式错误，缺少下划线分隔符")
	}

	// 获取编码前缀
	itemCodePrefix := gstr.SubStr(templateItemCode, 0, lastUnderscorePos)

	// 查询现有物品列表
	itemList, err := s.GetItemList(ctx, &item.GetItemListReq{
		ItemCodePrefix:  itemCodePrefix,
		ContainDisabled: true,
	})
	if err != nil {
		return "", gerror.Wrapf(err, "查询无规格物品列表失败")
	}

	// 计算最大后缀编号
	maxSuffix := s.calculateMaxSuffix(itemList.ItemList)

	// 生成新的后缀编号（补零两位）
	newSuffix := fmt.Sprintf("%02d", maxSuffix+1)

	return itemCodePrefix + "_" + newSuffix, nil
}

// calculateMaxSuffix 计算最大后缀编号
func (s *sItem) calculateMaxSuffix(itemList []*item.ItemInfo) int {
	maxSuffix := 0

	for _, item := range itemList {
		code := item.ItemCode
		// 查找最后一个下划线
		idx := gstr.PosR(code, "_")
		if idx > -1 {
			suffixStr := gstr.SubStr(code, idx+1)
			num := gconv.Int(suffixStr)
			if num > maxSuffix {
				maxSuffix = num
			}
		}
	}

	return maxSuffix
}

// GetItemStock 获取物品库存信息
// 根据公司简称、分支机构和物品编码查询库存信息
func (s *sItem) GetItemStock(ctx context.Context, req *item.GetItemStockReq) (res *item.GetItemStockResp, err error) {
	// 获取公司信息
	company, err := s.getCompanyInfo(ctx, req.CompanyAbbr)
	if err != nil {
		return nil, err
	}

	// 构建查询过滤器
	filters := gjson.New(g.Map{
		"company": company.CompanyName,
	})

	// 从默认仓库查询
	warehouse, err := service.Warehouse().GetDefaultWarehouse(ctx, company.CompanyName, req.Branch)
	if err != nil {
		return nil, gerror.Wrapf(err, "获取默认仓库失败")
	}
	filters.Set("warehouse", warehouse.Name)

	// 按物品编码过滤
	if len(req.ItemCode) > 0 {
		filters.Set("item_code", req.ItemCode)
	}

	// 执行库存报表查询
	resp, err := service.Report().Run(ctx, &erp.ReportParams{
		ReportName:           "Stock Projected Qty",
		Filters:              filters.String(),
		IgnorePreparedReport: true,
	})
	if err != nil {
		return nil, gerror.Wrapf(err, "查询库存报表失败")
	}

	// 解析响应数据
	j, err := gjson.DecodeToJson(resp.Bytes())
	if err != nil {
		return nil, gerror.Wrapf(err, "解析物品库存响应失败")
	}

	// 转换为物品库存列表
	dataArray := j.GetJsons("message.result")
	stockList := make([]*item.ItemStock, 0, len(dataArray))

	for _, data := range dataArray {
		if data.Contains("item_code") {
			stockList = append(stockList, &item.ItemStock{
				ItemCode:  data.Get("item_code").String(),
				ItemName:  data.Get("item_name").String(),
				ItemGroup: utility.ParseItemGroupFromString(data.Get("item_group").String()),
				Warehouse: data.Get("warehouse").String(),
				StockUom:  data.Get("stock_uom").String(),
				ActualQty: data.Get("actual_qty").Float64(),
			})
		}
	}

	// 构建响应结果
	return &item.GetItemStockResp{
		ItemStockList: stockList,
	}, nil
}

// GetItem 根据物品编码获取单个物品信息
// 参数：ctx 上下文，req 包含物品编码和可选的公司、分支信息
// 返回：物品详细信息，错误信息
func (s *sItem) GetItem(ctx context.Context, req *item.GetItemReq) (res *erp.Item, err error) {
	// 参数验证
	if len(req.ItemCode) == 0 {
		return nil, gerror.New("物品编码不能为空")
	}

	// 查询物品信息
	resp, err := service.Document().Get(ctx, &erp.ErpReq{
		DocType: erp.DocTypeItem,
		Name:    req.ItemCode,
	}, nil)

	if err != nil {
		return nil, gerror.Wrapf(err, "查询物品信息失败")
	}

	// 解析响应数据
	j, err := gjson.DecodeToJson(resp.Bytes())
	if err != nil {
		return nil, gerror.Wrapf(err, "解析物品信息响应失败")
	}
	itemInfo := &erp.Item{}
	gconv.Structs(j.GetJson("data"), &itemInfo)

	return itemInfo, nil
}
