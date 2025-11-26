package buying

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/api/buying"
	"ttpos-bmp/app/ttpos-erp/api/item"
	dto "ttpos-bmp/app/ttpos-erp/internal/model/dto/buying"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto/erp"
	"ttpos-bmp/app/ttpos-erp/internal/service"

	"github.com/gogf/gf/v2/container/garray"
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
	j := resp

	purchaseOrder := &erp.PurchaseOrder{}
	j.GetJson("data").Scan(purchaseOrder)
	//修改货币类型
	purchaseOrder.Currency = purchaseOrder.PriceListCurrency
	purchaseOrder.Supplier = req.Supplier
	purchaseOrder.ScheduleDate = req.RequiredBy

	//设置采购价格
	if req.BuyingPriceList != "" {
		purchaseOrder.BuyingPriceList = req.BuyingPriceList
	} else {
		//获取默认采购价格表
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

	//创建采购订单
	resp, err = service.Document().Create(ctx, erp.DocTypePurchaseOrder, purchaseOrder)
	if err != nil {
		return nil, gerror.Wrapf(err, "创建采购订单失败")
	}

	// 解析响应数据
	j = resp
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
// 参数：
//   - ctx: 上下文对象
//   - req: 创建内部销售订单请求参数
//
// 返回：
//   - res: 创建后的销售订单信息
//   - err: 错误信息
func (s *sBuying) CreateInnerSaleOrderFromPurchaseOrder(ctx context.Context, req *dto.CreateInnerSaleOrderFromPurchaseOrderReq) (res *erp.SaleOrder, err error) {
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
	j := resp
	salesOrder := &erp.SaleOrder{}
	j.GetJson("data").Scan(&salesOrder)
	// 发货时间
	salesOrder.DeliveryDate = req.DeliveryDate
	for _, item := range salesOrder.Items {
		item.DeliveryDate = req.DeliveryDate
	}

	// 处理dropship商品的供应商交付逻辑
	if err := s.processDripShopItems(ctx, salesOrder.Items); err != nil {
		g.Log().Warningf(ctx, "处理dropship商品失败: %v", err)
		// 不中断流程，继续创建订单
	}

	//设置来源仓库
	if req.SourceWarehouse != "" {
		salesOrder.SetWarehouse = req.SourceWarehouse
	} else {
		warehouse, err := service.Warehouse().GetDefaultWarehouse(ctx, salesOrder.Company, "")
		if err != nil {
			return nil, gerror.Wrapf(err, "查询默认仓库失败")
		}
		salesOrder.SetWarehouse = warehouse.Name
	}

	//设置采购价格
	if req.SellingPriceList != "" {
		salesOrder.SellingPriceList = req.SellingPriceList
	} else {
		//获取默认销售价格表
		defaultPriceList, err := service.PosPriceList().GetPosPriceListByCompany(ctx, salesOrder.Company)
		if err != nil {
			g.Log().Warningf(ctx, "获取销售价格表失败，company: %s", salesOrder.Company)
			defaultPriceList, err = service.PosPriceList().GetDefaultPosPriceList(ctx)
			if err != nil {
				return nil, gerror.Wrapf(err, "获取默认销售价格表失败")
			}
		}
		salesOrder.SellingPriceList = defaultPriceList.SellingPriceList
	}

	//创建内部销售订单
	resp, err = service.Document().Create(ctx, erp.DocTypeSaleOrder, salesOrder)
	if err != nil {
		return nil, gerror.Wrapf(err, "创建内部销售订单失败")
	}

	// 解析响应数据
	j = resp
	j.GetJson("data").Scan(&salesOrder)

	// 提交订单
	_, err = service.Document().ChangeDocStatus(ctx, erp.DocTypeSaleOrder, salesOrder.Name, erp.DocstatusSubmitted)
	if err != nil {
		return nil, gerror.Wrapf(err, "提交内部销售订单失败")
	}
	return salesOrder, nil
}

func (*sBuying) CreateDeliveryNoteFromInnerSaleOrder(ctx context.Context, req *dto.CreateDeliveryNoteFromInnerSaleOrderReq) (res *erp.DeliveryNote, err error) {
	resp, err := service.Rpc().Execute(ctx, &erp.ErpReq{
		Method: erp.ApiMethodMakeMappedDoc,
	}, g.MapStrStr{
		"method":      "erpnext.selling.doctype.sales_order.sales_order.make_delivery_note",
		"source_name": req.SourceName,
	})
	if err != nil {
		return nil, gerror.Wrapf(err, "创建内部销售订单失败")
	}

	// 解析响应数据
	j := resp

	deliveryNote := &erp.DeliveryNote{}
	j.GetJson("data").Scan(&deliveryNote)

	deliveryNote.SetWarehouse = req.SourceWarehouse
	deliveryNote.SetTargetWarehouse = req.TargetWarehouse
	deliveryNote.Docstatus = gconv.Int(erp.DocstatusSubmitted)
	//创建发货单
	resp, err = service.Document().Create(ctx, erp.DocTypeDeliveryNote, deliveryNote)
	if err != nil {
		return nil, gerror.Wrapf(err, "创建发货单失败")
	}
	j.GetJson("data").Scan(&deliveryNote)

	// 提交发货单
	//_, err = service.Document().ChangeDocStatus(ctx, erp.DocTypeDeliveryNote, deliveryNote.Name, erp.DocstatusSubmitted)
	//if err != nil {
	//	return nil, gerror.Wrapf(err, "提交发货单失败")
	//}
	return nil, nil
}

// GetPurchaseOrder 获取采购订单
func (*sBuying) GetPurchaseOrder(ctx context.Context, req *buying.GetPurchaseOrderReq) (*erp.PurchaseOrder, error) {

	resp, err := service.Document().Get(ctx, &erp.ErpReq{
		DocType: erp.DocTypePurchaseOrder,
		Name:    req.PurchaseOrderName,
	}, nil)
	if err != nil {
		return nil, gerror.Wrapf(err, "查询采购订单失败")
	}
	purchaseOrder := &erp.PurchaseOrder{}
	// 解析响应数据
	j := resp
	j.GetJson("data").Scan(purchaseOrder)
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
	j := resp

	receipt := &erp.PurchaseReceipt{}
	j.GetJson("data").Scan(&receipt)

	//根据入参调整item
	if req.Items != nil && len(req.Items) > 0 {
		receiptItems := make([]*erp.PurchaseReceiptItem, 0)
		//获取采购单中所有物品编码
		purchaseItemCodeList := make([]string, len(receipt.Items))
		_warehouse := receipt.SetWarehouse
		for _, item := range receipt.Items {
			purchaseItemCodeList = append(purchaseItemCodeList, item.ItemCode)
			if len(item.Warehouse) > 0 {
				_warehouse = item.Warehouse
			}
		}
		gPurchaseItemCodeList := garray.NewStrArrayFrom(purchaseItemCodeList)
		for _, itemReq := range req.Items {
			if gPurchaseItemCodeList.Contains(itemReq.ItemCode) {
				if len(itemReq.Warehouse) > 0 {
					_warehouse = itemReq.Warehouse
				}
				receiptItems = append(receiptItems, &erp.PurchaseReceiptItem{
					ItemCode:      itemReq.ItemCode,
					Qty:           itemReq.Qty,
					Uom:           itemReq.Uom,
					Warehouse:     _warehouse,
					PurchaseOrder: req.PurchaseOrderName,
				})
			}
		}
		receipt.Items = receiptItems
	}

	//创建采购收货订单
	resp, err = service.Document().Create(ctx, erp.DocTypePurchaseReceipt, receipt)
	if err != nil {
		return nil, gerror.Wrapf(err, "创建采购收货订单失败")
	}

	// 解析响应数据
	j = resp
	j.GetJson("data").Scan(&receipt)

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
	j := resp
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

// processDripShopItems 处理dropship商品的供应商交付逻辑
// 遍历销售订单商品，检查商品的delivered_by_supplier属性，
// 如果为true，则设置订单商品的DeliveredBySupplier为true
// 参数：
//   - ctx: 上下文对象
//   - items: 销售订单商品列表
//
// 返回：
//   - error: 错误信息
func (s *sBuying) processDripShopItems(ctx context.Context, items []*erp.SaleOrderItem) error {
	for _, orderItem := range items {
		// 获取商品信息
		itemDetail, err := service.Item().GetItem(ctx, &item.GetItemReq{
			ItemCode: orderItem.ItemCode,
		})
		if err != nil {
			// 获取商品信息失败，记录日志但不中断流程
			g.Log().Warningf(ctx, "获取商品信息失败，跳过dropship处理: %s, err: %v", orderItem.ItemCode, err)
			continue
		}

		// 检查是否为dropship商品
		if !s.isDripShopItem(itemDetail) {
			continue // 非dropship商品，跳过处理
		}

		// 设置供应商交付标识
		orderItem.DeliveredBySupplier = true
		g.Log().Debugf(ctx, "商品 %s 为dropship类型，已设置DeliveredBySupplier=true", orderItem.ItemCode)

		// 从商品的供应商列表中选择第一个供应商（如需要）
		supplier, err := s.selectFirstSupplier(ctx, itemDetail)
		if err != nil {
			g.Log().Warningf(ctx, "商品 %s 获取供应商失败: %v", orderItem.ItemCode, err)
			// 供应商获取失败不中断流程，DeliveredBySupplier 已设置
		} else if supplier != "" {
			g.Log().Debugf(ctx, "商品 %s 的第一个供应商为: %s", orderItem.ItemCode, supplier)
			orderItem.Supplier = supplier
		}
	}

	return nil
}

// isDripShopItem 判断商品是否为dropship商品
// 检查商品的delivered_by_supplier字段，1表示该商品由供应商直接交付
// 参数：
//   - itemDetail: 商品详细信息
//
// 返回：
//   - bool: 是否为dropship商品
func (s *sBuying) isDripShopItem(itemDetail *erp.Item) bool {
	if itemDetail == nil {
		return false
	}
	// 检查 Item 的 delivered_by_supplier 字段
	// 1 表示该商品是 dropship 类型，由供应商直接交付
	return itemDetail.DeliveredBySupplier == 1
}

// selectFirstSupplier 从商品中选择第一个供应商
// 解析商品的SupplierItems字段，获取第一个供应商名称
// 参数：
//   - ctx: 上下文对象
//   - itemDetail: 商品详细信息
//
// 返回：
//   - string: 供应商名称
//   - error: 错误信息
func (s *sBuying) selectFirstSupplier(ctx context.Context, itemDetail *erp.Item) (string, error) {
	if itemDetail == nil {
		return "", gerror.New("商品信息为空")
	}

	// 检查供应商列表是否为空
	if itemDetail.SupplierItems == nil || len(itemDetail.SupplierItems) == 0 {
		return "", gerror.Newf("商品 %s 没有供应商信息", itemDetail.ItemCode)
	}

	// 解析第一个供应商
	// SupplierItems 是 []interface{} 类型，使用 gjson 解析
	firstSupplierJson := gjson.New(itemDetail.SupplierItems[0])
	supplierName := firstSupplierJson.Get("supplier").String()

	if supplierName == "" {
		return "", gerror.Newf("商品 %s 的供应商列表中第一个供应商信息无效", itemDetail.ItemCode)
	}

	return supplierName, nil
}

// MustSelectFirstSupplier 从商品中选择第一个供应商，获取不到时返回默认供应商
// 解析商品的SupplierItems字段，获取第一个供应商名称
// 如果获取不到供应商信息，则返回指定的默认供应商
// 参数：
//   - ctx: 上下文对象
//   - itemDetail: 商品详细信息
//   - defaultSupplier: 默认供应商名称，当获取不到供应商时返回此值
//
// 返回：
//   - string: 供应商名称
func (s *sBuying) MustSelectFirstSupplier(ctx context.Context, itemDetail *erp.Item, defaultSupplier string) string {
	supplier, err := s.selectFirstSupplier(ctx, itemDetail)
	if err != nil {
		g.Log().Debugf(ctx, "获取供应商失败，使用默认供应商: %s, err: %v", defaultSupplier, err)
		return defaultSupplier
	}
	return supplier
}
