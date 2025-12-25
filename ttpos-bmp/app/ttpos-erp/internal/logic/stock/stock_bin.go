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
//   - req: GetBin 请求参数，包含物品代码、仓库名称
//
// 返回：
//   - res: Bin 数据，包含实际库存数量、估值率、库存价值
//   - err: 操作过程中产生的错误（若有）
func (s *sStock) GetBin(ctx context.Context, req *stock.GetBinReq) (*stock.BinData, error) {
	// 参数验证
	if err := s.validateGetBinReq(req); err != nil {
		g.Log().Warning(ctx, "GetBin 参数验证失败", err)
		return nil, err
	}

	// 构建查询过滤器
	filters := [][]string{
		{"item_code", "=", req.ItemCode},
		{"warehouse", "=", req.Warehouse},
	}

	// 查询 Bin 表
	binResp, err := service.Document().List(ctx, &erp.ErpReq{
		DocType: erp.DocTypeBin,
	}, &erp.RequestParams{
		Filters: filters,
		Fields:  []string{"item_code", "warehouse", "actual_qty", "valuation_rate", "stock_value"},
		Limit:   1,
	})

	if err != nil {
		g.Log().Errorf(ctx, "GetBin 查询 Bin 表失败: item_code=%s, warehouse=%s, err=%v",
			req.ItemCode, req.Warehouse, err)
		return nil, gerror.Wrapf(err, "查询 Bin 记录失败")
	}

	// 解析响应
	data := binResp.GetJsons("data")
	if len(data) == 0 {
		// Bin 表中无记录，返回估值率为 0 的空数据
		g.Log().Infof(ctx, "GetBin Bin 表中无记录: item_code=%s, warehouse=%s",
			req.ItemCode, req.Warehouse)

		return &stock.BinData{
			ItemCode:      req.ItemCode,
			Warehouse:     req.Warehouse,
			ActualQty:     0,
			ValuationRate: 0,
			StockValue:    0,
		}, nil
	}

	// 构建返回数据
	binData := data[0]
	resultData := &stock.BinData{
		ItemCode:      binData.Get("item_code").String(),
		Warehouse:     binData.Get("warehouse").String(),
		ActualQty:     binData.Get("actual_qty").Float64(),
		ValuationRate: binData.Get("valuation_rate").Float64(),
		StockValue:    binData.Get("stock_value").Float64(),
	}

	g.Log().Infof(ctx, "GetBin 查询成功: item_code=%s, warehouse=%s, valuation_rate=%.2f",
		req.ItemCode, req.Warehouse, resultData.ValuationRate)

	return resultData, nil
}

// validateGetBinReq 验证 GetBin 请求参数
func (s *sStock) validateGetBinReq(req *stock.GetBinReq) error {
	if req == nil {
		return gerror.New("请求参数不能为空")
	}
	if req.ItemCode == "" {
		return gerror.New("物品代码不能为空")
	}
	if req.Warehouse == "" {
		return gerror.New("仓库名称不能为空")
	}
	return nil
}
