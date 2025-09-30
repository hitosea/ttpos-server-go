package buying

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/api/buying"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto/erp"
	"ttpos-bmp/app/ttpos-erp/internal/service"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

// CreatePurchaseOrder 创建采购订单
// 参数：
//   - ctx: 上下文对象
//   - req: 创建采购订单请求参数
//
// 返回：
//   - res: 创建采购订单响应
//   - err: 错误信息
func (s *sBuying) CreatePurchaseOrder(ctx context.Context, req *buying.CreatePurchaseOrderReq) (res *buying.CreatePurchaseOrderResp, err error) {
	// 参数验证
	if err := s.validateCreatePurchaseOrderReq(req); err != nil {
		return nil, err
	}

	// 获取公司信息
	companyName, err := service.Company().GetCompanyNameWithAbbr(ctx, req.CompanyAbbr)
	if err != nil {
		return nil, gerror.Wrapf(err, "获取公司信息失败")
	}

	// 构建采购订单数据
	purchaseOrderData, err := s.buildCreatePurchaseOrderData(ctx, req, companyName)
	if err != nil {
		return nil, gerror.Wrapf(err, "构建采购订单数据失败")
	}

	// 创建采购订单
	resp, err := service.Document().Create(ctx, erp.DocTypePurchaseOrder, purchaseOrderData)
	if err != nil {
		return nil, gerror.Wrapf(err, "创建采购订单失败")
	}

	// 解析响应数据
	j, err := gjson.DecodeToJson(resp.Bytes())
	if err != nil {
		return nil, gerror.Wrapf(err, "解析创建采购订单响应失败")
	}

	//提交采购订单
	_, err = service.Document().ChangeDocStatus(ctx, erp.DocTypePurchaseOrder, j.Get("data.name").String(), erp.DocstatusSubmitted)
	if err != nil {
		return nil, gerror.Wrapf(err, "提交采购订单失败")
	}

	return &buying.CreatePurchaseOrderResp{
		Name:       j.Get("data.name").String(),
		Status:     "Receive and Bill",
		GrandTotal: j.Get("data.grand_total").Float64(),
	}, nil
}

// UpdatePurchaseOrder 更新采购订单
// 参数：
//   - ctx: 上下文对象
//   - req: 更新采购订单请求参数
//
// 返回：
//   - res: 更新采购订单响应
//   - err: 错误信息
func (s *sBuying) UpdatePurchaseOrder(ctx context.Context, req *buying.UpdatePurchaseOrderReq) (res *buying.UpdatePurchaseOrderResp, err error) {
	// 参数验证
	if err := s.validateUpdatePurchaseOrderReq(req); err != nil {
		return nil, err
	}

	// 构建更新数据
	updateData, err := s.buildUpdatePurchaseOrderData(ctx, req)
	if err != nil {
		return nil, gerror.Wrapf(err, "构建更新采购订单数据失败")
	}

	// 更新采购订单
	resp, err := service.Document().Update(ctx, &erp.ErpReq{
		DocType: erp.DocTypePurchaseOrder,
		Name:    req.Name,
	}, updateData)
	if err != nil {
		return nil, gerror.Wrapf(err, "更新采购订单失败")
	}

	// 解析响应数据
	j, err := gjson.DecodeToJson(resp.Bytes())
	if err != nil {
		return nil, gerror.Wrapf(err, "解析更新采购订单响应失败")
	}

	return &buying.UpdatePurchaseOrderResp{
		Name:       j.Get("data.name").String(),
		Status:     j.Get("data.status").String(),
		GrandTotal: j.Get("data.grand_total").Float64(),
	}, nil
}

// validateCreatePurchaseOrderReq 验证创建采购订单请求参数
func (s *sBuying) validateCreatePurchaseOrderReq(req *buying.CreatePurchaseOrderReq) error {
	if len(req.Supplier) == 0 {
		return gerror.New("供应商不能为空")
	}
	if len(req.CompanyAbbr) == 0 {
		return gerror.New("公司缩写不能为空")
	}
	//if len(req.TransactionDate) == 0 {
	//	return gerror.New("交易日期不能为空")
	//}
	if len(req.Items) == 0 {
		return gerror.New("采购订单项目不能为空")
	}

	// 验证每个项目
	for i, item := range req.Items {
		if len(item.ItemCode) == 0 {
			return gerror.Newf("第%d个项目的物品编码不能为空", i+1)
		}
		if item.Qty <= 0 {
			return gerror.Newf("第%d个项目的数量必须大于0", i+1)
		}
	}

	return nil
}

// validateUpdatePurchaseOrderReq 验证更新采购订单请求参数
func (s *sBuying) validateUpdatePurchaseOrderReq(req *buying.UpdatePurchaseOrderReq) error {
	if len(req.Name) == 0 {
		return gerror.New("采购订单名称不能为空")
	}

	// 验证项目（如果有提供）
	for i, item := range req.Items {
		if len(item.ItemCode) == 0 {
			return gerror.Newf("第%d个项目的物品编码不能为空", i+1)
		}
		if item.Qty <= 0 {
			return gerror.Newf("第%d个项目的数量必须大于0", i+1)
		}
	}

	return nil
}

// buildCreatePurchaseOrderData 构建创建采购订单的数据
func (s *sBuying) buildCreatePurchaseOrderData(ctx context.Context, req *buying.CreatePurchaseOrderReq, companyName string) (g.Map, error) {
	// 基础数据
	purchaseOrderData := g.Map{
		"supplier": req.Supplier,
		"company":  companyName,
	}

	// 可选字段
	if len(req.ScheduleDate) > 0 {
		purchaseOrderData["schedule_date"] = req.ScheduleDate
	}
	if len(req.Currency) > 0 {
		purchaseOrderData["currency"] = req.Currency
	} else {
		purchaseOrderData["currency"] = "THB" // 默认泰铢
	}
	if len(req.Remarks) > 0 {
		purchaseOrderData["remarks"] = req.Remarks
	}
	if len(req.TargetWarehouse) > 0 {
		purchaseOrderData["set_warehouse"] = req.TargetWarehouse
	}

	// 构建项目列表
	items := make([]g.Map, 0, len(req.Items))
	for _, item := range req.Items {
		itemData := g.Map{
			"item_code": item.ItemCode,
			"qty":       item.Qty,
		}

		if item.Rate > 0 {
			itemData["rate"] = item.Rate
		}
		if len(item.Uom) > 0 {
			itemData["uom"] = item.Uom
		}

		items = append(items, itemData)
	}
	purchaseOrderData["items"] = items

	return purchaseOrderData, nil
}

// buildUpdatePurchaseOrderData 构建更新采购订单的数据
func (s *sBuying) buildUpdatePurchaseOrderData(ctx context.Context, req *buying.UpdatePurchaseOrderReq) (g.Map, error) {
	updateData := g.Map{}

	// 可选字段
	if len(req.Supplier) > 0 {
		updateData["supplier"] = req.Supplier
	}
	if len(req.TransactionDate) > 0 {
		updateData["transaction_date"] = req.TransactionDate
	}
	if len(req.ScheduleDate) > 0 {
		updateData["schedule_date"] = req.ScheduleDate
	}
	if len(req.Remarks) > 0 {
		updateData["remarks"] = req.Remarks
	}

	if len(req.TargetWarehouse) > 0 {
		updateData["set_warehouse"] = req.TargetWarehouse
	}
	// 更新项目列表（如果有提供）
	if len(req.Items) > 0 {
		items := make([]g.Map, 0, len(req.Items))
		for _, item := range req.Items {
			itemData := g.Map{
				"item_code": item.ItemCode,
				"qty":       item.Qty,
			}

			if item.Rate > 0 {
				itemData["rate"] = item.Rate
			}
			if len(item.Uom) > 0 {
				itemData["uom"] = item.Uom
			}
			items = append(items, itemData)
		}
		updateData["items"] = items
	}

	return updateData, nil
}
