package buying

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/api/buying"
	dto "ttpos-bmp/app/ttpos-erp/internal/model/dto/buying"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto/erp"
	"ttpos-bmp/app/ttpos-erp/internal/service"

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
	j := resp

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
	j := resp

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
func (s *sBuying) buildCreatePurchaseOrderData(ctx context.Context, req *buying.CreatePurchaseOrderReq, companyName string) (*erp.PurchaseOrder, error) {
	// 基础数据
	purchaseOrderData := &erp.PurchaseOrder{
		Supplier: req.Supplier,
		Company:  companyName,
	}

	// 可选字段
	if len(req.ScheduleDate) > 0 {
		purchaseOrderData.ScheduleDate = req.ScheduleDate
	}
	if len(req.Currency) > 0 {
		purchaseOrderData.Currency = req.Currency
	} else {
		purchaseOrderData.Currency = "THB" // 默认泰铢
	}
	// 注意：PurchaseOrder 结构体没有 Remarks 字段，如果需要备注请使用其他字段
	if len(req.TargetWarehouse) > 0 {
		purchaseOrderData.SetWarehouse = req.TargetWarehouse
	}

	// 构建项目列表
	items := make([]*erp.PurchaseOrderItem, 0, len(req.Items))
	for _, item := range req.Items {
		itemData := &erp.PurchaseOrderItem{
			ItemCode: item.ItemCode,
			Qty:      item.Qty,
		}

		if item.Rate > 0 {
			itemData.Rate = item.Rate
		}
		if len(item.Uom) > 0 {
			itemData.Uom = item.Uom
		}

		if len(purchaseOrderData.SetWarehouse) > 0 {
			itemData.Warehouse = purchaseOrderData.SetWarehouse
		}

		items = append(items, itemData)
	}
	purchaseOrderData.Items = items

	//设置采购价格
	if req.BuyingPriceList != "" {
		purchaseOrderData.BuyingPriceList = req.BuyingPriceList
	} else {
		//子店、总店向外部采购的采购单使用“Standard Buying”
		purchaseOrderData.BuyingPriceList = "Standard Buying"
		//
		////获取默认采购价格表
		//defaultPriceList, err := service.PosPriceList().GetPosPriceListByCompany(ctx, purchaseOrderData.Company)
		//if err != nil {
		//	g.Log().Warningf(ctx, "获取采购价格表失败，company: %s", purchaseOrderData.Company)
		//	defaultPriceList, err = service.PosPriceList().GetDefaultPosPriceList(ctx)
		//	if err != nil {
		//		return nil, gerror.Wrapf(err, "获取默认采购价格表失败")
		//	}
		//}
		//purchaseOrderData.BuyingPriceList = defaultPriceList.BuyingPriceList
	}

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

// CreatePurchaseOrderFromSalesOrder 从销售订单创建采购订单
// 参数：
//   - ctx: 上下文对象
//   - req: 创建采购订单请求参数
//
// 返回：
//   - res: 采购订单信息
//   - err: 错误信息
func (s *sBuying) CreatePurchaseOrderFromSalesOrder(ctx context.Context, req *dto.CreatePurchaseOrderFromSalesOrderReq) (res *erp.PurchaseOrder, err error) {
	// 调用 ERPNext 的 make_mapped_doc 方法，从销售订单生成采购订单
	resp, err := service.Rpc().Execute(ctx, &erp.ErpReq{
		Method: erp.ApiMethodMakeMappedDoc,
	}, g.MapStrStr{
		"method":      erp.ApiMethodCreateInnerPurchaseOrder,
		"source_name": req.SourceName,
	})
	if err != nil {
		return nil, gerror.Wrapf(err, "创建内部采购订单失败")
	}

	// 解析响应数据
	j := resp
	purchaseOrder := &erp.PurchaseOrder{}
	j.GetJson("data").Scan(&purchaseOrder)

	// 设置预计交付日期
	if req.ScheduleDate != "" {
		purchaseOrder.ScheduleDate = req.ScheduleDate
		for _, item := range purchaseOrder.Items {
			item.ScheduleDate = req.ScheduleDate
		}
	}

	// 设置目标仓库
	if req.TargetWarehouse != "" {
		purchaseOrder.SetWarehouse = req.TargetWarehouse
	} else {
		// 获取默认仓库
		warehouse, err := service.Warehouse().GetDefaultWarehouse(ctx, purchaseOrder.Company, "")
		if err != nil {
			return nil, gerror.Wrapf(err, "查询默认仓库失败")
		}
		purchaseOrder.SetWarehouse = warehouse.Name
	}
	for _, item := range purchaseOrder.Items {
		if len(item.Warehouse) == 0 {
			item.Warehouse = purchaseOrder.SetWarehouse
		}
	}

	// 设置采购价格表
	if req.BuyingPriceList != "" {
		purchaseOrder.BuyingPriceList = req.BuyingPriceList
	} else {
		// 获取默认采购价格表
		defaultPriceList, err := service.PosPriceList().GetPosPriceListByCompany(ctx, purchaseOrder.Company)
		if err != nil {
			g.Log().Warningf(ctx, "获取采购价格表失败，company: %s", purchaseOrder.Company)
			defaultPriceList, err = service.PosPriceList().GetDefaultPosPriceList(ctx)
			if err != nil {
				return nil, gerror.Wrapf(err, "获取默认采购价格表失败")
			}
		}
		purchaseOrder.BuyingPriceList = defaultPriceList.BuyingPriceList
	}

	// 设置供应商（如果提供）
	if req.Supplier != "" {
		purchaseOrder.Supplier = req.Supplier
	}

	// 创建采购订单
	resp, err = service.Document().Create(ctx, erp.DocTypePurchaseOrder, purchaseOrder)
	if err != nil {
		return nil, gerror.Wrapf(err, "创建采购订单失败")
	}

	// 解析响应数据
	j = resp
	j.GetJson("data").Scan(&purchaseOrder)

	// 提交订单
	_, err = service.Document().ChangeDocStatus(ctx, erp.DocTypePurchaseOrder, purchaseOrder.Name, erp.DocstatusSubmitted)
	if err != nil {
		return nil, gerror.Wrapf(err, "提交采购订单失败")
	}

	return purchaseOrder, nil
}
