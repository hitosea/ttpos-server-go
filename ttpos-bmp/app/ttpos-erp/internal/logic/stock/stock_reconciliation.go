package stock

import (
	"context"
	"fmt"
	"math"
	"strings"
	itemApi "ttpos-bmp/app/ttpos-erp/api/item"
	"ttpos-bmp/app/ttpos-erp/api/stock"
	"ttpos-bmp/app/ttpos-erp/internal/consts"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto/erp"
	"ttpos-bmp/app/ttpos-erp/internal/service"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// SaveStockReconciliation 保存库存盘点
// 参数：ctx 上下文，req 保存库存盘点请求
// 返回：保存库存盘点响应，错误信息
func (s *sStock) SaveStockReconciliation(ctx context.Context, req *stock.SaveStockReconciliationReq) (res *stock.SaveStockReconciliationResp, err error) {
	// 获取公司信息
	company, err := service.Company().GetCompanyWithAbbr(ctx, req.CompanyAbbr)
	if err != nil {
		return nil, gerror.Wrapf(err, "获取公司信息失败")
	}

	// 设置仓库
	warehouseName := req.Warehouse
	if len(warehouseName) == 0 && len(req.Branch) > 0 {
		// 如果没有指定仓库但指定了分支，获取默认仓库
		warehouse, err := service.Warehouse().GetDefaultWarehouse(ctx, company.CompanyName, req.Branch)
		if err != nil {
			return nil, gerror.Wrapf(err, "获取默认仓库失败")
		}
		warehouseName = warehouse.Name
	}

	// 构建库存盘点数据
	data := &erp.StockReconciliation{
		NamingSeries: erp.DefaultStockReconciliationSeries,
		Company:      company.CompanyName,
		PostingDate:  req.PostingDate,
		DocType:      erp.DocTypeStockReconciliation,
	}

	// 设置过账时间
	if len(req.PostingTime) > 0 {
		data.PostingTime = req.PostingTime
		data.SetPostingTime = 1
	} else {
		data.PostingTime = service.Setup().MustGetLocalDateTime(ctx, gtime.Now()).Format("H:i:s")
	}

	// 设置用途
	if len(req.Purpose) > 0 {
		data.Purpose = req.Purpose
	} else {
		data.Purpose = erp.StockReconciliationPurposeStockReconciliation
	}

	// 设置仓库
	if len(warehouseName) > 0 {
		data.SetWarehouse = warehouseName
	}

	// 构建明细项目
	itemList := make([]erp.StockReconciliationItem, 0)
	for _, item := range req.Items {
		//判断当前仓库已有库存是否与盘点数量一致，如果一致就不要继续新增盘点记录了
		// 查询当前物品在当前仓库的库存数量
		stockQtyResp, err := service.Item().GetItemStock(ctx, &itemApi.GetItemStockReq{
			ItemCode:    item.ItemCode,
			Warehouse:   warehouseName,
			CompanyAbbr: req.CompanyAbbr,
		})
		if err != nil {
			return nil, gerror.Wrapf(err, "查询物品%s在仓库%s的库存数量失败", item.ItemCode, warehouseName)
		}
		// 如果当前物品在当前仓库的库存数量与盘点数量不一致，就不要继续新增盘点记录了; 如果当前物品在当前仓库没有库存，也不要新增盘点记录
		{
			if len(stockQtyResp.ItemStockList) > 0 && math.Abs(stockQtyResp.ItemStockList[0].ActualQty-item.Qty) < consts.DefaultDecimalPrecision {
				g.Log().Warningf(ctx, "物品[%s]在仓库[%s]的库存数量为[%.2f]，与盘点数量[%.2f]一致,无需新增盘点记录", item.ItemCode, warehouseName, stockQtyResp.ItemStockList[0].ActualQty, item.Qty)
				continue
			}
			if len(stockQtyResp.ItemStockList) == 0 && item.Qty == 0 {
				g.Log().Warningf(ctx, "物品[%s]在仓库[%s]的库存无库存且盘点数量为[0]无需新增盘点记录", item.ItemCode, warehouseName)
				continue
			}
		}

		itemData := erp.StockReconciliationItem{
			ItemCode: item.ItemCode,
			ItemName: item.ItemName,
			Qty:      item.Qty,
			DocType:  erp.DocTypeStockReconciliationItem,
		}

		// 如果明细项目指定了仓库，使用明细仓库
		if len(item.Warehouse) > 0 {
			itemData.Warehouse = item.Warehouse
		} else if len(warehouseName) > 0 {
			itemData.Warehouse = warehouseName
		}

		// 设置估值价格
		if item.ValuationRate > 0 {
			// 用户提供了估值率，直接使用
			itemData.ValuationRate = item.ValuationRate
			g.Log().Infof(ctx, "使用用户提供的估值率: item_code=%s, valuation_rate=%.2f",
				item.ItemCode, item.ValuationRate)
		} else {
			// 估值率为 0，从 Bin 表查询
			binResp, err := s.GetBin(ctx, &stock.GetBinReq{
				ItemCode:  item.ItemCode,
				Warehouse: itemData.Warehouse,
			})

			if err == nil && binResp != nil && len(binResp.Items) > 0 {
				// 使用 Bin 表中的估值率（因为指定了 item_code 和 warehouse，应该只有一条记录）
				itemData.ValuationRate = binResp.Items[0].ValuationRate
				g.Log().Infof(ctx, "从 Bin 表获取估值率: item_code=%s, warehouse=%s, valuation_rate=%.2f",
					item.ItemCode, itemData.Warehouse, binResp.Items[0].ValuationRate)
			} else {
				// Bin 表中没有估值率，返回异常
				return nil, gerror.New("缺少对应仓库的入库记录,无估值率!")
			}
		}

		itemList = append(itemList, itemData)
	}
	data.Items = itemList

	//如果盘点的物品为0，则返回默认的盘点单号
	if len(itemList) == 0 {
		return &stock.SaveStockReconciliationResp{
			StockReconciliationName: consts.DefaultStockReconciliationName,
		}, nil
	}

	// 创建库存盘点单据
	resp, err := service.Document().Create(ctx, erp.DocTypeStockReconciliation, data)
	if err != nil {
		return nil, gerror.Wrapf(err, "创建库存盘点单据失败")
	}

	j := resp
	if !j.Contains("data") {
		return nil, gerror.New("创建库存盘点单据失败：响应数据为空")
	}

	return &stock.SaveStockReconciliationResp{
		StockReconciliationName: j.Get("data.name").String(),
	}, nil
}

// GetStockReconciliationList 获取库存盘点列表
// 参数：ctx 上下文，req 获取库存盘点列表请求
// 返回：获取库存盘点列表响应，错误信息
func (s *sStock) GetStockReconciliationList(ctx context.Context, req *stock.GetStockReconciliationListReq) (res *stock.GetStockReconciliationListResp, err error) {
	// 获取公司信息
	company, err := service.Company().GetCompanyWithAbbr(ctx, req.CompanyAbbr)
	if err != nil {
		return nil, gerror.Wrapf(err, "获取公司信息失败")
	}

	// 构建查询过滤器
	filters := [][]string{
		{"company", "=", company.CompanyName},
	}

	// 按仓库过滤
	if len(req.Warehouse) > 0 {
		filters = append(filters, []string{"set_warehouse", "like", "%" + req.Warehouse + "%"})
	} else if len(req.Branch) > 0 {
		// 如果没有指定仓库但指定了分支，获取默认仓库
		warehouse, err := service.Warehouse().GetDefaultWarehouse(ctx, company.CompanyName, req.Branch)
		if err == nil && len(warehouse.Name) > 0 {
			filters = append(filters, []string{"set_warehouse", "like", "%" + warehouse.Name + "%"})
		}
	}

	// 按用途过滤
	if len(req.Purpose) > 0 {
		filters = append(filters, []string{"purpose", "=", req.Purpose})
	}

	// 按日期范围过滤
	if len(req.FromDate) > 0 {
		filters = append(filters, []string{"posting_date", ">=", req.FromDate})
	}
	if len(req.ToDate) > 0 {
		filters = append(filters, []string{"posting_date", "<=", req.ToDate})
	}

	// 设置查询限制
	limit := consts.Limit100
	if req.Limit > 0 && req.Limit <= 1000 {
		limit = int(req.Limit)
	}

	// 执行查询
	resp, err := service.Document().List(ctx, &erp.ErpReq{
		DocType: erp.DocTypeStockReconciliation,
	}, &erp.RequestParams{
		Fields: g.ArrayStr{
			"name", "company", "purpose", "posting_date", "posting_time",
			"set_warehouse", "difference_amount", "expense_account", "cost_center", "docstatus",
		},
		Filters: filters,
		Limit:   limit,
	})

	if err != nil {
		return nil, gerror.Wrapf(err, "查询库存盘点列表失败")
	}

	// 解析响应数据
	j := resp
	res = &stock.GetStockReconciliationListResp{
		StockReconciliationList: make([]*stock.StockReconciliation, 0),
	}

	dataArray := j.GetJsons("data")
	for _, data := range dataArray {
		reconciliation := &stock.StockReconciliation{
			Name:         data.Get("name").String(),
			Company:      data.Get("company").String(),
			Purpose:      data.Get("purpose").String(),
			PostingDate:  data.Get("posting_date").String(),
			PostingTime:  data.Get("posting_time").String(),
			SetWarehouse: data.Get("set_warehouse").String(),
		}

		// 获取明细项目
		if len(reconciliation.Name) > 0 {
			items, _ := s.getStockReconciliationItems(ctx, reconciliation.Name)
			reconciliation.Items = items
		}

		res.StockReconciliationList = append(res.StockReconciliationList, reconciliation)
	}

	return res, nil
}

// getStockReconciliationItems 获取库存盘点明细
func (s *sStock) getStockReconciliationItems(ctx context.Context, stockReconciliationName string) ([]*stock.StockReconciliationItem, error) {
	// 查询库存盘点单据详情
	resp, err := service.Document().Get(ctx, &erp.ErpReq{
		DocType: erp.DocTypeStockReconciliation,
		Name:    stockReconciliationName,
	}, nil)

	if err != nil {
		return nil, gerror.Wrapf(err, "查询库存盘点单据详情失败")
	}

	// 解析响应数据
	j := resp
	items := make([]*stock.StockReconciliationItem, 0)

	itemsArray := j.GetJsons("data.items")
	for _, itemData := range itemsArray {
		item := &stock.StockReconciliationItem{
			ItemCode:      itemData.Get("item_code").String(),
			ItemName:      itemData.Get("item_name").String(),
			Qty:           itemData.Get("qty").Float64(),
			Warehouse:     itemData.Get("warehouse").String(),
			ValuationRate: itemData.Get("valuation_rate").Float64(),
		}
		items = append(items, item)
	}

	return items, nil
}

// SubmitStockReconciliation 提交库存盘点
// 参数：ctx 上下文，req 提交库存盘点请求
// 返回：提交库存盘点响应，错误信息
func (s *sStock) SubmitStockReconciliation(ctx context.Context, req *stock.SubmitStockReconciliationReq) (res *stock.SubmitStockReconciliationResp, err error) {
	// 参数验证
	if len(req.StockReconciliationName) == 0 {
		return nil, gerror.New("库存盘点单号不能为空")
	}
	if req.StockReconciliationName == consts.DefaultStockReconciliationName {
		return &stock.SubmitStockReconciliationResp{
			Message: "库存盘点使用默认单据号提交成功",
		}, nil
	}

	// 查询盘点单详情，验证估值率
	resp, err := service.Document().Get(ctx, &erp.ErpReq{
		DocType: erp.DocTypeStockReconciliation,
		Name:    req.StockReconciliationName,
	}, nil)
	if err != nil {
		return nil, gerror.Wrapf(err, "查询盘点单详情失败")
	}

	// 验证估值率
	j := resp
	itemsArray := j.GetJsons("data.items")
	invalidItems := make([]string, 0)

	for _, itemData := range itemsArray {
		itemCode := itemData.Get("item_code").String()
		warehouse := itemData.Get("warehouse").String()
		valuationRate := itemData.Get("valuation_rate").Float64()

		// 检查估值率是否为 0 或 1（保底值）
		if valuationRate == 0 || valuationRate == 1 {
			invalidItems = append(invalidItems,
				fmt.Sprintf("物品 [%s] 在仓库 [%s] 的估值率为空（%.2f）",
					itemCode, warehouse, valuationRate))
		}
	}

	if len(invalidItems) > 0 {
		errMsg := fmt.Sprintf("盘点单中有物品估值率为空，无法提交。请先通过采购入库建立库存。\n详情：\n%s",
			strings.Join(invalidItems, "\n"))
		g.Log().Warning(ctx, errMsg)
		return nil, gerror.New(errMsg)
	}

	g.Log().Infof(ctx, "估值率验证通过，提交盘点单: %s", req.StockReconciliationName)

	// 提交库存盘点单据
	_, err = service.Document().ChangeDocStatus(ctx, erp.DocTypeStockReconciliation, req.StockReconciliationName, erp.DocstatusSubmitted)
	if err != nil {
		return nil, gerror.Wrapf(err, "提交库存盘点单据失败")
	}

	return &stock.SubmitStockReconciliationResp{
		Message: "库存盘点单据提交成功",
	}, nil
}

// CancelStockReconciliation 取消库存盘点
// 参数：ctx 上下文，req 取消库存盘点请求
// 返回：取消库存盘点响应，错误信息
func (s *sStock) CancelStockReconciliation(ctx context.Context, req *stock.CancelStockReconciliationReq) (res *stock.CancelStockReconciliationResp, err error) {
	// 参数验证
	if len(req.StockReconciliationName) == 0 {
		return nil, gerror.New("库存盘点单号不能为空")
	}

	// 取消库存盘点单据
	_, err = service.Document().ChangeDocStatus(ctx, erp.DocTypeStockReconciliation, req.StockReconciliationName, erp.DocstatusCancelled)
	if err != nil {
		return nil, gerror.Wrapf(err, "取消库存盘点单据失败")
	}

	return &stock.CancelStockReconciliationResp{
		Message: "库存盘点单据取消成功",
	}, nil
}
