package stock

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/api/stock"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto/erp"
	"ttpos-bmp/app/ttpos-erp/internal/service"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// SubmitStockEntry 提交库存变动（物料出库）
// 参数：ctx 上下文，req 提交库存变动请求
// 返回：提交库存变动响应，错误信息
func (s *sStock) SubmitStockEntry(ctx context.Context, req *stock.SubmitStockEntryReq) (res *stock.SubmitStockEntryResp, err error) {
	// 获取公司信息
	company, err := service.Company().GetCompanyWithAbbr(ctx, req.CompanyAbbr)
	if err != nil {
		return nil, gerror.Wrapf(err, "获取公司信息失败")
	}

	// 设置仓库
	warehouseName := req.SourceWarehouse
	if len(warehouseName) == 0 && len(req.Branch) > 0 {
		// 如果没有指定仓库但指定了分支，获取默认仓库
		warehouse, err := service.Warehouse().GetDefaultWarehouse(ctx, company.CompanyName, req.Branch)
		if err != nil {
			return nil, gerror.Wrapf(err, "获取默认仓库失败")
		}
		warehouseName = warehouse.Name
	}

	// 设置库存变动类型，默认为 Material Issue
	stockEntryType := req.StockEntryType
	if len(stockEntryType) == 0 {
		stockEntryType = erp.StockEntryTypeMaterialIssue
	}

	// 构建库存变动数据
	data := &erp.StockEntry{
		NamingSeries:   erp.DefaultStockEntryNamingSeries,
		Company:        company.CompanyName,
		StockEntryType: stockEntryType,
		PostingDate:    req.PostingDate,
		DocType:        erp.DocTypeStockEntry,
	}

	if len(req.Remarks) > 0 {
		data.Remarks = req.Remarks
	}

	// 设置过账时间
	if len(req.PostingTime) > 0 {
		data.PostingTime = req.PostingTime
		data.SetPostingTime = 1
	} else {
		data.PostingTime = service.Setup().MustGetLocalDateTime(ctx, gtime.Now()).Format("H:i:s")
	}

	// 构建明细项目
	itemList := make([]erp.StockEntryDetail, 0)
	for _, item := range req.Items {
		// 明细仓库优先使用物品指定的仓库，否则使用默认仓库
		itemWarehouse := item.Warehouse
		if len(itemWarehouse) == 0 {
			itemWarehouse = warehouseName
		}

		itemData := erp.StockEntryDetail{
			ItemCode:   item.ItemCode,
			ItemName:   item.ItemName,
			Qty:        item.Qty,
			SWarehouse: itemWarehouse,
			DocType:    erp.DocTypeStockEntryDetail,
		}

		itemList = append(itemList, itemData)
	}
	data.Items = itemList

	// 如果变动的物品为空，返回错误
	if len(itemList) == 0 {
		return nil, gerror.New("库存变动明细不能为空")
	}

	// 创建库存变动单据
	resp, err := service.Document().Create(ctx, erp.DocTypeStockEntry, data)
	if err != nil {
		return nil, gerror.Wrapf(err, "创建库存变动单据失败")
	}

	j := resp
	if !j.Contains("data") {
		return nil, gerror.New("创建库存变动单据失败：响应数据为空")
	}

	stockEntryName := j.Get("data.name").String()
	g.Log().Infof(ctx, "创建库存变动单据成功: %s", stockEntryName)

	// 提交库存变动单据
	_, err = service.Document().ChangeDocStatus(ctx, erp.DocTypeStockEntry, stockEntryName, erp.DocstatusSubmitted)
	if err != nil {
		// 提交失败，尝试取消已创建的单据
		g.Log().Warningf(ctx, "提交库存变动单据失败，尝试取消: %s, err: %v", stockEntryName, err)
		return nil, gerror.Wrapf(err, "提交库存变动单据失败")
	}

	return &stock.SubmitStockEntryResp{
		StockEntryName: stockEntryName,
		Message:        "库存变动单据提交成功",
	}, nil
}
