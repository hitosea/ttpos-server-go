package stock

import (
	"context"
	"fmt"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto/erp"
	"ttpos-bmp/app/ttpos-erp/internal/service"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// StockEntryMergeItem 合并后的 Stock Entry 行项
type StockEntryMergeItem struct {
	ItemCode  string
	ItemName  string
	Warehouse string
	Qty       float64
}

// SubmitMergedStockEntry 合并多个订单的物品明细后提交一张 Stock Entry
// items 为已按 (item_code, warehouse) 合并后的行项列表
func (s *sStock) SubmitMergedStockEntry(ctx context.Context, companyAbbr string, postingDate string, items []StockEntryMergeItem, remarks string) (string, error) {
	if len(items) == 0 {
		return "", gerror.New("合并后的 Stock Entry 行项为空")
	}

	company, err := service.Company().GetCompanyWithAbbr(ctx, companyAbbr)
	if err != nil {
		return "", gerror.Wrapf(err, "获取公司信息失败")
	}

	// 设置仓库：使用第一个 item 的仓库作为默认，后续行项使用各自的仓库
	postingTime := service.Setup().MustGetLocalDateTime(ctx, gtime.Now()).Format("H:i:s")

	data := &erp.StockEntry{
		NamingSeries:   erp.DefaultStockEntryNamingSeries,
		Company:        company.CompanyName,
		StockEntryType: StockEntryTypeMergedDeduction,
		PostingDate:    postingDate,
		PostingTime:    postingTime,
		SetPostingTime: 1,
		DocType:        erp.DocTypeStockEntry,
		Remarks:        remarks,
	}

	// 构建明细项目
	itemList := make([]erp.StockEntryDetail, 0, len(items))
	for _, item := range items {
		itemList = append(itemList, erp.StockEntryDetail{
			ItemCode:   item.ItemCode,
			ItemName:   item.ItemName,
			Qty:        item.Qty,
			SWarehouse: item.Warehouse,
			DocType:    erp.DocTypeStockEntryDetail,
		})
	}
	data.Items = itemList

	// 创建 Stock Entry
	resp, err := service.Document().Create(ctx, erp.DocTypeStockEntry, data)
	if err != nil {
		return "", gerror.Wrapf(err, "创建合并Stock Entry失败")
	}

	seName := resp.Get("data.name").String()
	if seName == "" {
		return "", gerror.New("创建合并Stock Entry失败：响应数据为空")
	}

	g.Log().Infof(ctx, "创建合并Stock Entry成功: %s, items=%d", seName, len(items))

	// 提交
	_, err = service.Document().ChangeDocStatus(ctx, erp.DocTypeStockEntry, seName, erp.DocstatusSubmitted)
	if err != nil {
		g.Log().Warningf(ctx, "提交合并Stock Entry[%s]失败: %v", seName, err)
		return seName, gerror.Wrapf(err, "提交合并Stock Entry失败")
	}

	return seName, nil
}

// SubmitReverseStockEntry 生成反向 Stock Entry（反结账时回增库存）
// 用于反结账场景：如果订单已通过 Stock Entry 扣减库存，需要生成一张 Material Receipt 回增
func (s *sStock) SubmitReverseStockEntry(ctx context.Context, companyAbbr string, postingDate string, items []StockEntryMergeItem, remarks string) (string, error) {
	if len(items) == 0 {
		return "", gerror.New("反向 Stock Entry 行项为空")
	}

	company, err := service.Company().GetCompanyWithAbbr(ctx, companyAbbr)
	if err != nil {
		return "", gerror.Wrapf(err, "获取公司信息失败")
	}

	postingTime := service.Setup().MustGetLocalDateTime(ctx, gtime.Now()).Format("H:i:s")

	data := &erp.StockEntry{
		NamingSeries:   erp.DefaultStockEntryNamingSeries,
		Company:        company.CompanyName,
		StockEntryType: erp.StockEntryTypeMaterialReceipt,
		PostingDate:    postingDate,
		PostingTime:    postingTime,
		SetPostingTime: 1,
		DocType:        erp.DocTypeStockEntry,
		Remarks:        fmt.Sprintf("反结账回增库存: %s", remarks),
	}

	itemList := make([]erp.StockEntryDetail, 0, len(items))
	for _, item := range items {
		itemList = append(itemList, erp.StockEntryDetail{
			ItemCode:   item.ItemCode,
			ItemName:   item.ItemName,
			Qty:        item.Qty,
			TWarehouse: item.Warehouse, // 目标仓库（入库）
			DocType:    erp.DocTypeStockEntryDetail,
		})
	}
	data.Items = itemList

	resp, err := service.Document().Create(ctx, erp.DocTypeStockEntry, data)
	if err != nil {
		return "", gerror.Wrapf(err, "创建反向Stock Entry失败")
	}

	seName := resp.Get("data.name").String()
	if seName == "" {
		return "", gerror.New("创建反向Stock Entry失败：响应数据为空")
	}

	_, err = service.Document().ChangeDocStatus(ctx, erp.DocTypeStockEntry, seName, erp.DocstatusSubmitted)
	if err != nil {
		return seName, gerror.Wrapf(err, "提交反向Stock Entry[%s]失败", seName)
	}

	g.Log().Infof(ctx, "创建反向Stock Entry成功: %s, items=%d", seName, len(items))
	return seName, nil
}

// MergeStockItems 按 (item_code, warehouse) 合并行项
func MergeStockItems(items []StockEntryMergeItem) []StockEntryMergeItem {
	type key struct {
		ItemCode  string
		Warehouse string
	}
	merged := make(map[key]*StockEntryMergeItem)
	order := make([]key, 0)

	for _, item := range items {
		k := key{ItemCode: item.ItemCode, Warehouse: item.Warehouse}
		if existing, ok := merged[k]; ok {
			existing.Qty += item.Qty
		} else {
			cp := item
			merged[k] = &cp
			order = append(order, k)
		}
	}

	result := make([]StockEntryMergeItem, 0, len(merged))
	for _, k := range order {
		result = append(result, *merged[k])
	}
	return result
}

// StockEntryTypeMergedDeduction 合并扣减类型
const StockEntryTypeMergedDeduction = "Material Issue"
