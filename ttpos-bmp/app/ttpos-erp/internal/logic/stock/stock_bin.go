package stock

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/api/stock"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto/erp"
	"ttpos-bmp/app/ttpos-erp/internal/service"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

// GetBin 查询物品在指定仓库的 Bin 记录
// 用于获取物品的真实估值率
// 参数：
//   - ctx: 上下文对象，用于传递请求范围的元数据
//   - req: GetBin 请求参数，包含仓库名称（必填）和可选的物品代码
//
// 返回：
//   - res: Bin 数据列表，包含实际库存数量、估值率、库存价值
//   - err: 操作过程中产生的错误（若有）
func (s *sStock) GetBin(ctx context.Context, req *stock.GetBinReq) (*stock.GetBinResp, error) {
	// 参数验证
	if err := s.validateGetBinReq(req); err != nil {
		g.Log().Warning(ctx, "GetBin 参数验证失败", err)
		return nil, err
	}

	// 构建查询过滤器
	filters := [][]string{
		{"warehouse", "=", req.Warehouse},
	}
	// 如果指定了物品代码，添加到过滤器中
	if req.ItemCode != "" {
		filters = append(filters, []string{"item_code", "=", req.ItemCode})
	}

	// 查询 Bin 表，获取所有需要的字段
	binResp, err := service.Document().List(ctx, &erp.ErpReq{
		DocType: erp.DocTypeBin,
	}, &erp.RequestParams{
		Filters: filters,
		Fields: []string{
			"item_code",
			"warehouse",
			"actual_qty",
			"projected_qty",
			"reserved_qty",
			"stock_uom",
			"valuation_rate",
			"stock_value",
		},
		Limit: 99999, // 设置最大记录数
	})

	if err != nil {
		g.Log().Errorf(ctx, "GetBin 查询 Bin 表失败: warehouse=%s, item_code=%s, err=%v",
			req.Warehouse, req.ItemCode, err)
		return nil, gerror.Wrapf(err, "查询 Bin 记录失败")
	}

	// 解析响应
	data := binResp.GetJsons("data")
	if len(data) == 0 {
		// Bin 表中无记录，返回空列表
		g.Log().Infof(ctx, "GetBin Bin 表中无记录: warehouse=%s, item_code=%s",
			req.Warehouse, req.ItemCode)
		return &stock.GetBinResp{
			Items: []*stock.ItemStockBin{},
		}, nil
	}

	// 构建返回数据列表
	var result []*stock.ItemStockBin
	for _, binData := range data {
		item := &stock.ItemStockBin{
			ItemCode:      binData.Get("item_code").String(),
			Warehouse:     binData.Get("warehouse").String(),
			ActualQty:     binData.Get("actual_qty").Float64(),
			ProjectedQty:  binData.Get("projected_qty").Float64(),
			ReservedQty:   binData.Get("reserved_qty").Float64(),
			StockUom:      binData.Get("stock_uom").String(),
			ValuationRate: binData.Get("valuation_rate").Float64(),
			StockValue:    binData.Get("stock_value").Float64(),
		}
		result = append(result, item)
	}

	g.Log().Infof(ctx, "GetBin 查询成功: warehouse=%s, item_code=%s, result_count=%d",
		req.Warehouse, req.ItemCode, len(result))

	return &stock.GetBinResp{
		Items: result,
	}, nil
}

// validateGetBinReq 验证 GetBin 请求参数
func (s *sStock) validateGetBinReq(req *stock.GetBinReq) error {
	if req == nil {
		return gerror.New("请求参数不能为空")
	}
	if req.Warehouse == "" {
		return gerror.New("仓库名称不能为空")
	}
	return nil
}
