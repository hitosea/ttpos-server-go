package buying

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/api/buying"
	dto "ttpos-bmp/app/ttpos-erp/internal/model/dto/buying"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto/erp"
	"ttpos-bmp/app/ttpos-erp/internal/service"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
)

type sBuying struct {
}

var (
	Buying = new(sBuying)
)

func init() {
	service.RegisterBuying(Buying)
}

// CreatePurchaseFromMq 根据材料请求创建采购订单
func (s *sBuying) CreatePurchaseFromMq(ctx context.Context, req *dto.CreatePurchaseFromMqReq) (res *erp.PurchaseOrder, err error) {

	resp, err := service.Rpc().Execute(ctx, &erp.ErpReq{
		Method: erp.ApiMethodMakeMappedDoc,
	}, g.MapStrStr{
		"method":      "erpnext.stock.doctype.material_request.material_request.make_purchase_order",
		"source_name": req.SourceName,
	})
	if err != nil {
		return nil, gerror.Wrapf(err, "创建采购订单失败")
	}

	// 解析响应数据
	j, err := gjson.DecodeToJson(resp.Bytes())
	if err != nil {
		return nil, gerror.Wrapf(err, "解析采购订单响应失败")
	}
	purchaseOrder := &erp.PurchaseOrder{}
	j.Get("data").Scan(purchaseOrder)
	//修改货币类型
	purchaseOrder.Currency = purchaseOrder.PriceListCurrency
	purchaseOrder.Supplier = req.Supplier
	purchaseOrder.ScheduleDate = req.RequiredBy

	//创建采购订单
	resp, err = service.Document().Create(ctx, "Purchase Order", purchaseOrder)
	if err != nil {
		return nil, gerror.Wrapf(err, "创建采购订单失败")
	}

	// 解析响应数据
	j, err = gjson.DecodeToJson(resp.Bytes())
	if err != nil {
		return nil, gerror.Wrapf(err, "解析采购订单响应失败")
	}
	res = &erp.PurchaseOrder{}
	gconv.Scan(purchaseOrder, res)
	res.Name = j.Get("data.name").String()

	//提交采购订单
	_, err = service.Document().ChangeDocStatus(ctx, erp.DocTypePurchaseOrder, res.Name, erp.DocstatusSubmitted)
	if err != nil {
		return nil, gerror.Wrapf(err, "提交采购订单失败")
	}
	return
}

// CreateInnerSaleOrderFromPurchaseOrder 创建内部销售订单
func (*sBuying) CreateInnerSaleOrderFromPurchaseOrder(ctx context.Context, req *dto.CreateInnerSaleOrderFromPurchaseOrderReq) (res *erp.SaleOrder, err error) {
	resp, err := service.Rpc().Execute(ctx, &erp.ErpReq{
		Method: erp.ApiMethodMakeMappedDoc,
	}, g.MapStrStr{
		"method":      "erpnext.buying.doctype.purchase_order.purchase_order.make_inter_company_sales_order",
		"source_name": req.SourceName,
	})
	if err != nil {
		return nil, gerror.Wrapf(err, "创建内部销售订单失败")
	}

	// 解析响应数据
	j, err := gjson.DecodeToJson(resp.Bytes())
	if err != nil {
		return nil, gerror.Wrapf(err, "解析采购订单响应失败")
	}
	salesOrder := &erp.SaleOrder{}
	j.Get("data").Scan(salesOrder)
	// 发货时间
	salesOrder.DeliveryDate = req.DeliveryDate
	for _, item := range salesOrder.Items {
		item.DeliveryDate = req.DeliveryDate
	}

	//设置来源仓库
	warehouse, err := service.Warehouse().GetDefaultWarehouse(ctx, salesOrder.Company, "")
	if err != nil {
		return nil, gerror.Wrapf(err, "查询默认仓库失败")
	}
	salesOrder.SetWarehouse = warehouse.Name

	//创建采购订单
	resp, err = service.Document().Create(ctx, "Sales Order", salesOrder)
	if err != nil {
		return nil, gerror.Wrapf(err, "创建内部销售订单失败")
	}

	// 解析响应数据
	j, err = gjson.DecodeToJson(resp.Bytes())
	if err != nil {
		return nil, gerror.Wrapf(err, "解析内部销售订单响应失败")
	}
	j.Get("data").Scan(salesOrder)

	// 提交订单
	_, err = service.Document().ChangeDocStatus(ctx, erp.DocTypeSaleOrder, salesOrder.Name, erp.DocstatusSubmitted)
	if err != nil {
		return nil, gerror.Wrapf(err, "提交内部销售订单失败")
	}
	return salesOrder, nil
}

// GetPurchaseOrder 获取采购订单
func (*sBuying) GetPurchaseOrder(ctx context.Context, req *buying.GetPurchaseOrderReq) (*erp.PurchaseOrder, error) {

	resp, err := service.Document().Get(ctx, &erp.ErpReq{
		DocType: "Purchase Order",
		Name:    req.PurchaseOrderName,
	}, nil)
	if err != nil {
		return nil, gerror.Wrapf(err, "查询采购订单失败")
	}
	purchaseOrder := &erp.PurchaseOrder{}
	// 解析响应数据
	j, err := gjson.DecodeToJson(resp.Bytes())
	if err != nil {
		return nil, gerror.Wrapf(err, "解析采购订单响应失败")
	}
	j.Get("data").Scan(purchaseOrder)
	return purchaseOrder, nil
}

// CreatePurchaseReceiptFromOrder 创建采购收货订单
func (*sBuying) CreatePurchaseReceiptFromOrder(ctx context.Context, req *buying.SavePurchaseReceiptReq) (*erp.PurchaseReceipt, error) {
	resp, err := service.Rpc().Execute(ctx, &erp.ErpReq{
		Method: erp.ApiMethodMakeMappedDoc,
	}, g.MapStrStr{
		"method":      "erpnext.buying.doctype.purchase_order.purchase_order.make_purchase_receipt",
		"source_name": req.PurchaseOrderName,
	})
	if err != nil {
		return nil, gerror.Wrapf(err, "创建内部销售订单失败")
	}

	// 解析响应数据
	j, err := gjson.DecodeToJson(resp.Bytes())
	if err != nil {
		return nil, gerror.Wrapf(err, "解析采购订单响应失败")
	}

	receipt := &erp.PurchaseReceipt{}
	j.Get("data").Scan(receipt)

	//根据入参调整item
	receiptItems := make([]erp.PurchaseReceiptItem, 0)
	for _, item := range receipt.Items {
		for _, itemReq := range req.Items {
			if item.ItemCode == itemReq.ItemCode {
				item.Qty = itemReq.Qty
				receiptItems = append(receiptItems, item)
				break
			}
		}
	}
	receipt.Items = receiptItems

	//创建采购收货订单
	resp, err = service.Document().Create(ctx, erp.DocTypePurchaseReceipt, receipt)
	if err != nil {
		return nil, gerror.Wrapf(err, "创建采购收货订单失败")
	}

	// 解析响应数据
	j, err = gjson.DecodeToJson(resp.Bytes())
	if err != nil {
		return nil, gerror.Wrapf(err, "解析采购收货订单响应失败")
	}
	j.Get("data").Scan(receipt)

	// 提交订单
	_, err = service.Document().ChangeDocStatus(ctx, erp.DocTypePurchaseReceipt, receipt.Name, erp.DocstatusSubmitted)
	if err != nil {
		return nil, gerror.Wrapf(err, "提交采购收货订单失败")
	}
	return receipt, nil
}

// GetPurchaseOrderList 获取采购订单列表
// 根据查询条件过滤并返回采购订单信息列表
// 参数：
//   - ctx: 上下文对象
//   - req: 获取采购订单列表请求参数
//
// 返回：
//   - res: 采购订单列表响应
//   - err: 错误信息
func (s *sBuying) GetPurchaseOrderList(ctx context.Context, req *buying.GetPurchaseOrderListReq) (res *buying.GetPurchaseOrderListResp, err error) {
	// 构建查询过滤器
	filters := s.buildPurchaseOrderListFilters(ctx, req)

	// 构建分页参数
	pageNo := req.PageNo
	if pageNo <= 0 {
		pageNo = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}

	// 查询采购订单列表
	resp, err := service.Document().List(ctx, &erp.ErpReq{
		DocType: erp.DocTypePurchaseOrder,
	}, &erp.RequestParams{
		Fields: g.ArrayStr{
			"name", "supplier", "company", "grand_total", "status",
			"transaction_date", "schedule_date", "per_received", "per_billed", "currency",
		},
		Filters:    filters,
		Limit:      int(pageSize),
		LimitStart: int((pageNo - 1) * pageSize),
	})
	if err != nil {
		return nil, gerror.Wrapf(err, "查询采购订单列表失败")
	}

	// 解析响应数据
	j, err := gjson.DecodeToJson(resp.Bytes())
	if err != nil {
		return nil, gerror.Wrapf(err, "解析采购订单列表响应失败")
	}

	// 转换为采购订单列表
	purchaseOrderList := make([]*buying.PurchaseOrderListItem, 0)
	dataArray := j.GetJsons("data")

	for _, data := range dataArray {
		purchaseOrderItem := &buying.PurchaseOrderListItem{
			Name:            data.Get("name").String(),
			Supplier:        data.Get("supplier").String(),
			Company:         data.Get("company").String(),
			GrandTotal:      data.Get("grand_total").Float64(),
			Status:          data.Get("status").String(),
			TransactionDate: data.Get("transaction_date").String(),
			ScheduleDate:    data.Get("schedule_date").String(),
			PerReceived:     data.Get("per_received").Float64(),
			PerBilled:       data.Get("per_billed").Float64(),
			Currency:        data.Get("currency").String(),
		}
		purchaseOrderList = append(purchaseOrderList, purchaseOrderItem)
	}

	// 获取总数量
	totalCount, err := s.GetPurchaseOrderCount(ctx, &buying.GetPurchaseOrderCountReq{
		Supplier:    req.Supplier,
		CompanyAbbr: req.CompanyAbbr,
		FromDate:    req.FromDate,
		ToDate:      req.ToDate,
	})
	if err != nil {
		g.Log().Warning(ctx, "获取采购订单总数失败", err)
		totalCount = &buying.GetPurchaseOrderCountResp{Count: 0}
	}

	return &buying.GetPurchaseOrderListResp{
		PurchaseOrders: purchaseOrderList,
		TotalCount:     totalCount.Count,
	}, nil
}

// GetPurchaseOrderCount 获取采购订单数量
// 根据查询条件统计采购订单数量
// 参数：
//   - ctx: 上下文对象
//   - req: 获取采购订单数量请求参数
//
// 返回：
//   - res: 采购订单数量响应
//   - err: 错误信息
func (s *sBuying) GetPurchaseOrderCount(ctx context.Context, req *buying.GetPurchaseOrderCountReq) (res *buying.GetPurchaseOrderCountResp, err error) {
	// 构建查询过滤器
	filters := s.buildPurchaseOrderCountFilters(ctx, req)

	// 查询采购订单数量
	count, err := service.Doctype().Count(ctx, &erp.ErpReq{
		DocType: erp.DocTypePurchaseOrder,
	}, &erp.RequestParams{
		Filters: filters,
	})
	if err != nil {
		return nil, gerror.Wrapf(err, "查询采购订单数量失败")
	}

	return &buying.GetPurchaseOrderCountResp{
		Count: int32(count),
	}, nil
}

// buildPurchaseOrderListFilters 构建采购订单列表查询过滤器
// 参数：
//   - ctx: 上下文对象
//   - req: 获取采购订单列表请求参数
//
// 返回：
//   - [][]string: 过滤条件数组
func (s *sBuying) buildPurchaseOrderListFilters(ctx context.Context, req *buying.GetPurchaseOrderListReq) [][]string {
	filters := make([][]string, 0, 8) // 预分配容量，提高性能

	// 按供应商过滤
	if len(req.Supplier) > 0 {
		filters = append(filters, g.ArrayStr{"supplier", "=", req.Supplier})
	}

	// 按公司缩写过滤
	if len(req.CompanyAbbr) > 0 {
		companyName, err := service.Company().GetCompanyNameWithAbbr(ctx, req.CompanyAbbr)
		if err == nil && len(companyName) > 0 {
			filters = append(filters, g.ArrayStr{"company", "=", companyName})
		}
	}

	// 按日期范围过滤
	if len(req.FromDate) > 0 {
		filters = append(filters, g.ArrayStr{"transaction_date", ">=", req.FromDate})
	}
	if len(req.ToDate) > 0 {
		filters = append(filters, g.ArrayStr{"transaction_date", "<=", req.ToDate})
	}

	return filters
}

// buildPurchaseOrderCountFilters 构建采购订单数量查询过滤器
// 参数：
//   - ctx: 上下文对象
//   - req: 获取采购订单数量请求参数
//
// 返回：
//   - [][]string: 过滤条件数组
func (s *sBuying) buildPurchaseOrderCountFilters(ctx context.Context, req *buying.GetPurchaseOrderCountReq) [][]string {
	filters := make([][]string, 0, 8) // 预分配容量，提高性能

	// 按供应商过滤
	if len(req.Supplier) > 0 {
		filters = append(filters, g.ArrayStr{"supplier", "=", req.Supplier})
	}

	// 按公司缩写过滤
	if len(req.CompanyAbbr) > 0 {
		companyName, err := service.Company().GetCompanyNameWithAbbr(ctx, req.CompanyAbbr)
		if err == nil && len(companyName) > 0 {
			filters = append(filters, g.ArrayStr{"company", "=", companyName})
		}
	}

	// 按状态过滤
	if len(req.Status) > 0 {
		filters = append(filters, g.ArrayStr{"status", "=", req.Status})
	}

	// 按日期范围过滤
	if len(req.FromDate) > 0 {
		filters = append(filters, g.ArrayStr{"transaction_date", ">=", req.FromDate})
	}
	if len(req.ToDate) > 0 {
		filters = append(filters, g.ArrayStr{"transaction_date", "<=", req.ToDate})
	}

	return filters
}
