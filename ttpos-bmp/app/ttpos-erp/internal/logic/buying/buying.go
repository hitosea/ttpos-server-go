package buying

import (
	"context"
	"strings"
	"ttpos-bmp/app/ttpos-erp/api/buying"
	"ttpos-bmp/app/ttpos-erp/api/item"
	"ttpos-bmp/app/ttpos-erp/api/stock"
	"ttpos-bmp/app/ttpos-erp/api/warehouse"
	dto "ttpos-bmp/app/ttpos-erp/internal/model/dto/buying"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto/erp"
	"ttpos-bmp/app/ttpos-erp/internal/service"

	"github.com/gogf/gf/v2/container/gvar"
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

	// 补写 PO Item 自定义字段（ERPNext make_purchase_order 不会从 MR 携带自定义字段）
	s.fillPurchaseOrderItemCustomFields(ctx, purchaseOrder, req)

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

	// 设置税费参数
	var taxTemplateName string
	if req.TaxesAndCharges != "" {
		// 优先使用传入的税费模板
		taxTemplateName = req.TaxesAndCharges
	} else {
		// 未传入时，查询公司默认采购税费模板
		taxTemplateName = service.Accounts().GetDefaultPurchaseTaxTemplate(ctx, purchaseOrder.Company)
	}

	// 如果有税费模板，获取模板详情并复制 taxes 明细
	if taxTemplateName != "" {
		purchaseOrder.TaxesAndCharges = taxTemplateName
		// 获取模板的 taxes 子表并复制到 PO
		taxes := service.Accounts().GetPurchaseTaxTemplateDetails(ctx, taxTemplateName)
		if len(taxes) > 0 {
			purchaseOrder.Taxes = s.convertPurchaseTaxesToInterface(taxes)
		}
	}

	// 设置税类别
	if req.TaxCategory != "" {
		purchaseOrder.TaxCategory = req.TaxCategory
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

// fillPurchaseOrderItemCustomFields 补写 PO Item 自定义字段
// ERPNext 的 make_purchase_order（MR→PO）不会携带自定义字段，需在 PO 创建前手动设置
func (s *sBuying) fillPurchaseOrderItemCustomFields(ctx context.Context, purchaseOrder *erp.PurchaseOrder, req *dto.CreatePurchaseFromMqReq) {
	if len(purchaseOrder.Items) == 0 {
		return
	}

	// 收集 PO 物品编码
	itemCodes := make([]string, 0, len(purchaseOrder.Items))
	for _, poItem := range purchaseOrder.Items {
		if poItem.ItemCode != "" {
			itemCodes = append(itemCodes, poItem.ItemCode)
		}
	}
	if len(itemCodes) == 0 {
		return
	}

	availableQtyMap := make(map[string]float64) // item_code -> 总店默认仓库库存
	storeQtyMap := make(map[string]float64)     // item_code -> 门店所有仓库库存

	// 查询总店物品默认仓库库存（custom_qty_available_for_purchase）
	if req.SourceCompanyAbbr != "" {
		itemWarehouses, err := service.Item().GetItemDefaultWarehouses(ctx, itemCodes, req.SourceCompanyAbbr, false)
		if err != nil {
			g.Log().Warningf(ctx, "fillPurchaseOrderItemCustomFields-查询总店物品默认仓库失败: %v", err)
		} else {
			// 按仓库分组，相同仓库的物品一次查询
			warehouseItemCodes := make(map[string][]string)
			for _, itemCode := range itemCodes {
				whName, ok := itemWarehouses[itemCode]
				if !ok || whName == "" {
					continue
				}
				warehouseItemCodes[whName] = append(warehouseItemCodes[whName], itemCode)
			}
			for whName, codes := range warehouseItemCodes {
				binResp, err := service.Stock().GetBin(ctx, &stock.GetBinReq{
					Warehouse: whName,
				})
				if err != nil {
					g.Log().Warningf(ctx, "fillPurchaseOrderItemCustomFields-查询仓库Bin失败: warehouse=%s, err=%v", whName, err)
					continue
				}
				binMap := make(map[string]float64)
				for _, bin := range binResp.Items {
					binMap[bin.ItemCode] = bin.ActualQty
				}
				for _, itemCode := range codes {
					if qty, ok := binMap[itemCode]; ok {
						availableQtyMap[itemCode] = qty
					}
				}
			}
		}
	}

	// 查询门店所有仓库库存（custom_store_qty）
	if req.CompanyAbbr != "" {
		warehouseResp, err := service.Warehouse().GetWarehouseList(ctx, &warehouse.GetWarehouseListReq{
			CompanyAbbr: req.CompanyAbbr,
		})
		if err != nil {
			g.Log().Warningf(ctx, "fillPurchaseOrderItemCustomFields-查询门店仓库列表失败: %v", err)
		} else if len(warehouseResp.WarehouseList) > 0 {
			for _, wh := range warehouseResp.WarehouseList {
				binResp, err := service.Stock().GetBin(ctx, &stock.GetBinReq{
					Warehouse: wh.Name,
				})
				if err != nil {
					continue
				}
				for _, bin := range binResp.Items {
					storeQtyMap[bin.ItemCode] += bin.ActualQty
				}
			}
		}
	}

	// 设置自定义字段到 PO Items（按 PO Item 的 UOM 换算）
	// Bin 的 ActualQty 是 stock_uom，PO Item 的 qty 是 uom
	// conversion_factor = stock_uom / uom，即 display_qty = stock_qty / conversion_factor
	for _, poItem := range purchaseOrder.Items {
		cf := poItem.ConversionFactor
		if cf <= 0 {
			cf = 1
		}
		if qty, ok := req.ItemLastPurchaseQty[poItem.ItemCode]; ok && qty > 0 {
			poItem.CustomLastPurchaseQty = qty
		}
		if qty, ok := availableQtyMap[poItem.ItemCode]; ok {
			poItem.CustomQtyAvailableForPurchase = qty / cf
		}
		if qty, ok := storeQtyMap[poItem.ItemCode]; ok {
			poItem.CustomStoreQty = qty / cf
		}
	}
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

	// 处理dropship商品的供应商交付逻辑，同时判断是否全部为直送商品
	allDropShip, err := s.processDripShopItems(ctx, salesOrder.Items)
	if err != nil {
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

	// 设置税费参数
	var taxTemplateName string
	if req.TaxesAndCharges != "" {
		// 优先使用传入的税费模板
		taxTemplateName = req.TaxesAndCharges
	} else {
		// 未传入时，查询公司默认销售税费模板
		taxTemplateName = service.Accounts().GetDefaultSalesTaxTemplate(ctx, salesOrder.Company)
	}

	// 如果有税费模板，获取模板详情并复制 taxes 明细
	if taxTemplateName != "" {
		salesOrder.TaxesAndCharges = taxTemplateName
		// 获取模板的 taxes 子表并复制到 SO
		taxes := service.Accounts().GetSalesTaxTemplateDetails(ctx, taxTemplateName)
		if len(taxes) > 0 {
			salesOrder.Taxes = s.convertSalesTaxesToInterface(taxes)
		}
	}

	// 设置税类别
	if req.TaxCategory != "" {
		salesOrder.TaxCategory = req.TaxCategory
	}

	//创建内部销售订单
	resp, err = service.Document().Create(ctx, erp.DocTypeSaleOrder, salesOrder)
	if err != nil {
		return nil, gerror.Wrapf(err, "创建内部销售订单失败")
	}

	// 解析响应数据
	j = resp
	j.GetJson("data").Scan(&salesOrder)

	//20251226 任务 38171 ERP-品采销售订单审批需求，在品采流程中，把销售订单状态变为草稿，结合工作审批流。
	//20260305 任务 40167 如果全部物品都是直送(dropship)，SO 直接提交，跳过审批流。
	if allDropShip {
		_, err = service.Document().ChangeDocStatus(ctx, erp.DocTypeSaleOrder, salesOrder.Name, erp.DocstatusSubmitted)
		if err != nil {
			return nil, gerror.Wrapf(err, "提交内部销售订单失败")
		}
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

	// 获取采购订单信息（用于超收场景和物品匹配）
	purchaseOrder, err := service.Buying().GetPurchaseOrder(ctx, &buying.GetPurchaseOrderReq{
		PurchaseOrderName: req.PurchaseOrderName,
	})
	if err != nil {
		return nil, gerror.Wrapf(err, "获取采购订单失败")
	}

	//根据入参调整item
	if req.Items != nil && len(req.Items) > 0 {
		receiptItems := make([]*erp.PurchaseReceiptItem, 0)

		for _, itemReq := range req.Items {
			var receiptItem *erp.PurchaseReceiptItem

			// 首先尝试从待收货列表中匹配（正常收货场景）
			for _, item := range receipt.Items {
				if item.ItemCode == itemReq.ItemCode && item.Uom == itemReq.Uom {
					receiptItem = &erp.PurchaseReceiptItem{}
					gvar.New(item).Clone().Scan(receiptItem)
					receiptItem.Qty = itemReq.Qty
					break
				}
			}

			// 如果待收货列表为空或未找到匹配项，从采购订单物品中构建（超收场景）
			if receiptItem == nil {
				for _, poItem := range purchaseOrder.Items {
					if poItem.ItemCode == itemReq.ItemCode && poItem.Uom == itemReq.Uom {
						receiptItem = &erp.PurchaseReceiptItem{
							ItemCode:          poItem.ItemCode,
							ItemName:          poItem.ItemName,
							ItemGroup:         poItem.ItemGroup,
							Description:       poItem.Description,
							Qty:               itemReq.Qty,
							Uom:               poItem.Uom,
							StockUom:          poItem.StockUom,
							ConversionFactor:  poItem.ConversionFactor,
							Rate:              poItem.Rate,
							BaseRate:          poItem.BaseRate,
							PriceListRate:     poItem.PriceListRate,
							BasePriceListRate: poItem.BasePriceListRate,
							Warehouse:         poItem.Warehouse,
							PurchaseOrder:     req.PurchaseOrderName,
							PurchaseOrderItem: poItem.Name,
							CostCenter:        poItem.CostCenter,
							ExpenseAccount:    poItem.ExpenseAccount,
						}
						break
					}
				}
			}

			// 校验：请求的物品必须在采购单中存在
			if receiptItem == nil {
				g.Log().Warningf(ctx, "物品 %s (单位: %s) 不在采购单中，跳过", itemReq.ItemCode, itemReq.Uom)
				continue
			}

			receiptItems = append(receiptItems, receiptItem)
		}

		// 校验处理后的物品列表不为空
		if len(receiptItems) == 0 {
			return nil, gerror.Newf("没有有效的收货物品，请检查请求的物品是否在采购订单 %s 中", req.PurchaseOrderName)
		}

		receipt.Items = receiptItems
	} else if len(receipt.Items) == 0 {
		// 如果未指定收货物品且待收货列表为空，返回错误
		return nil, gerror.Newf("采购订单 %s 没有待收货的物品，请指定要收货的物品列表", req.PurchaseOrderName)
	}

	// 设置跨公司订单引用
	if req.InterCompanyReference != "" {
		receipt.IsInternalSupplier = true
		receipt.InterCompanyReference = req.InterCompanyReference
	}

	// 将结构体转换为 map，以便手动设置计算字段
	// 解决 omitempty 导致 0 值字段被省略，ERPNext 收到 None 引发 TypeError
	receiptMap := gconv.Map(receipt)

	// 禁用舍入总额，让 ERPNext 使用 grand_total 而不是 rounded_total
	receiptMap["disable_rounded_total"] = true

	// 清除所有计算字段（设置为 0），让 ERPNext 根据新的 items 列表重新计算
	// 由于使用 map，这些字段会被保留而不会因为 omitempty 被省略
	receiptMap["total"] = 0
	receiptMap["net_total"] = 0
	receiptMap["base_total"] = 0
	receiptMap["base_net_total"] = 0
	receiptMap["grand_total"] = 0
	receiptMap["base_grand_total"] = 0
	receiptMap["rounded_total"] = 0
	receiptMap["base_rounded_total"] = 0
	receiptMap["rounding_adjustment"] = 0
	receiptMap["base_rounding_adjustment"] = 0
	receiptMap["total_qty"] = 0
	receiptMap["total_net_weight"] = 0
	receiptMap["total_taxes_and_charges"] = 0
	receiptMap["base_total_taxes_and_charges"] = 0
	receiptMap["tax_withholding_net_total"] = 0
	receiptMap["base_tax_withholding_net_total"] = 0
	receiptMap["taxes_and_charges_added"] = 0
	receiptMap["taxes_and_charges_deducted"] = 0
	receiptMap["base_taxes_and_charges_added"] = 0
	receiptMap["base_taxes_and_charges_deducted"] = 0
	receiptMap["in_words"] = ""
	receiptMap["base_in_words"] = ""

	//创建采购收货订单
	resp, err = service.Document().Create(ctx, erp.DocTypePurchaseReceipt, receiptMap)
	if err != nil {
		return nil, gerror.Wrapf(err, "创建采购收货订单失败")
	}

	// 解析响应数据
	j = resp
	j.GetJson("data").Scan(&receipt)

	// 提交订单
	_, err = service.Document().ChangeDocStatus(ctx, erp.DocTypePurchaseReceipt, receipt.Name, erp.DocstatusSubmitted)
	if err != nil {
		// 如果是超收错误，删除已创建的收货单草稿
		if strings.Contains(err.Error(), "OverAllowanceError") {
			if _, deleteErr := service.Document().Delete(ctx, &erp.ErpReq{
				DocType: erp.DocTypePurchaseReceipt,
				Name:    receipt.Name,
			}); deleteErr != nil {
				g.Log().Warningf(ctx, "删除超收失败的收货单草稿 %s 失败: %v", receipt.Name, deleteErr)
			} else {
				g.Log().Infof(ctx, "已删除超收失败的收货单草稿: %s", receipt.Name)
			}
		}
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
		Name:        req.Name,
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

	// 按采购订单名称过滤（支持 IN 查询）
	if len(req.Name) > 0 {
		if strings.Contains(req.Name, ",") {
			// 多个值使用 IN 查询
			filters = append(filters, g.ArrayStr{"name", "in", req.Name})
		} else {
			// 单个值使用等值查询
			filters = append(filters, g.ArrayStr{"name", "=", req.Name})
		}
	}

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

	// 按采购订单名称过滤（支持 IN 查询）
	if len(req.Name) > 0 {
		if strings.Contains(req.Name, ",") {
			// 多个值使用 IN 查询
			filters = append(filters, g.ArrayStr{"name", "in", req.Name})
		} else {
			// 单个值使用等值查询
			filters = append(filters, g.ArrayStr{"name", "=", req.Name})
		}
	}

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
//   - allDropShip: 是否所有商品都是dropship类型
//   - error: 错误信息
func (s *sBuying) processDripShopItems(ctx context.Context, items []*erp.SaleOrderItem) (allDropShip bool, err error) {
	if len(items) == 0 {
		return false, nil
	}
	allDropShip = true
	for _, orderItem := range items {
		// 获取商品信息
		itemDetail, err := service.Item().GetItem(ctx, &item.GetItemReq{
			ItemCode: orderItem.ItemCode,
		})
		if err != nil {
			// 获取商品信息失败，记录日志但不中断流程，保守视为非dropship
			g.Log().Warningf(ctx, "获取商品信息失败，跳过dropship处理: %s, err: %v", orderItem.ItemCode, err)
			allDropShip = false
			continue
		}

		// 检查是否为dropship商品
		if !s.isDripShopItem(itemDetail) {
			allDropShip = false
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

	return allDropShip, nil
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

// convertPurchaseTaxesToInterface 将采购税费明细转换为 interface 数组
// 复制配置字段，清除计算字段（由 ERPNext 自动计算）
// 参数：
//   - taxes: 采购税费明细列表
//
// 返回：
//   - []interface{}: 转换后的税费明细数组
func (s *sBuying) convertPurchaseTaxesToInterface(taxes []*erp.PurchaseTaxesAndCharges) []interface{} {
	result := make([]interface{}, len(taxes))
	for i, tax := range taxes {
		// 复制税费明细，清除计算字段
		result[i] = &erp.PurchaseTaxesAndCharges{
			Category:            tax.Category,
			AddDeductTax:        tax.AddDeductTax,
			ChargeType:          tax.ChargeType,
			RowId:               tax.RowId,
			AccountHead:         tax.AccountHead,
			Description:         tax.Description,
			CostCenter:          tax.CostCenter,
			Rate:                tax.Rate,
			IncludedInPrintRate: tax.IncludedInPrintRate,
			// 不复制计算字段（TaxAmount, Total 等），由 ERPNext 自动计算
		}
	}
	return result
}

// convertSalesTaxesToInterface 将销售税费明细转换为 interface 数组
// 复制配置字段，清除计算字段（由 ERPNext 自动计算）
// 参数：
//   - taxes: 销售税费明细列表
//
// 返回：
//   - []interface{}: 转换后的税费明细数组
func (s *sBuying) convertSalesTaxesToInterface(taxes []*erp.SalesTaxesAndCharges) []interface{} {
	result := make([]interface{}, len(taxes))
	for i, tax := range taxes {
		// 复制税费明细，清除计算字段
		result[i] = &erp.SalesTaxesAndCharges{
			ChargeType:          tax.ChargeType,
			RowId:               tax.RowId,
			AccountHead:         tax.AccountHead,
			Description:         tax.Description,
			CostCenter:          tax.CostCenter,
			Rate:                tax.Rate,
			IncludedInPrintRate: tax.IncludedInPrintRate,
			// 不复制计算字段（TaxAmount, Total 等），由 ERPNext 自动计算
		}
	}
	return result
}

// CreateSplitSaleOrdersFromPurchaseOrder 从采购订单创建按仓库拆分的内部销售订单
// 按物品默认仓库分组，为每个仓库创建独立的 SO
// 直采物品（delivered_by_supplier）单独生成一张 SO 并自动提交
func (s *sBuying) CreateSplitSaleOrdersFromPurchaseOrder(ctx context.Context, req *dto.CreateInnerSaleOrderFromPurchaseOrderReq) ([]*erp.SaleOrder, error) {
	// 1. 调用 ERPNext API 获取 SO 模板（包含 PO 的所有物品）
	resp, err := service.Rpc().Execute(ctx, &erp.ErpReq{
		Method: erp.ApiMethodMakeMappedDoc,
	}, g.MapStrStr{
		"method":      "erpnext.buying.doctype.purchase_order.purchase_order.make_inter_company_sales_order",
		"source_name": req.SourceName,
	})
	if err != nil {
		return nil, gerror.Wrapf(err, "创建内部销售订单模板失败")
	}

	soTemplate := &erp.SaleOrder{}
	resp.GetJson("data").Scan(&soTemplate)
	soTemplate.DeliveryDate = req.DeliveryDate
	for _, itm := range soTemplate.Items {
		itm.DeliveryDate = req.DeliveryDate
	}

	// 2. 标记直采物品
	_, err = s.processDripShopItems(ctx, soTemplate.Items)
	if err != nil {
		g.Log().Warningf(ctx, "处理dropship商品失败: %v", err)
	}

	// 3. 查询 SO 所属公司（总部）的 abbr，用于获取物品默认仓库
	companyAbbr := ""
	if erpCompany, err := service.Company().GetCompany(ctx, soTemplate.Company); err == nil && erpCompany != nil {
		companyAbbr = erpCompany.Abbr
	} else {
		g.Log().Warningf(ctx, "获取SO公司简称失败, company: %s, err: %v", soTemplate.Company, err)
	}

	// 4. 收集物品编码，查询默认仓库
	itemCodes := make([]string, 0, len(soTemplate.Items))
	for _, itm := range soTemplate.Items {
		if itm.ItemCode != "" {
			itemCodes = append(itemCodes, itm.ItemCode)
		}
	}
	itemWarehouses := make(map[string]string)
	if companyAbbr != "" && len(itemCodes) > 0 {
		itemWarehouses, err = service.Item().GetItemDefaultWarehouses(ctx, itemCodes, companyAbbr, false)
		if err != nil {
			g.Log().Warningf(ctx, "查询物品默认仓库失败: %v", err)
			itemWarehouses = make(map[string]string)
		}
	}

	// 5. 分组：直采组 + 按仓库的集采组
	var dropshipItems []*erp.SaleOrderItem
	warehouseGroups := make(map[string][]*erp.SaleOrderItem)

	for _, itm := range soTemplate.Items {
		if itm.DeliveredBySupplier {
			dropshipItems = append(dropshipItems, itm)
		} else {
			warehouse := itemWarehouses[itm.ItemCode]
			if warehouse == "" {
				// 回退到公司默认仓库
				wh, whErr := service.Warehouse().GetDefaultWarehouse(ctx, soTemplate.Company, "")
				if whErr == nil && wh != nil {
					warehouse = wh.Name
				}
			}
			warehouseGroups[warehouse] = append(warehouseGroups[warehouse], itm)
		}
	}

	// 6. 获取公共配置（价格表、税费）
	sellingPriceList := s.resolveSellingPriceList(ctx, req.SellingPriceList, soTemplate.Company)
	taxTemplateName, taxes := s.resolveSalesTax(ctx, req.TaxesAndCharges, soTemplate.Company)

	// 7. 为每个仓库组创建 SO
	var createdOrders []*erp.SaleOrder

	for warehouse, items := range warehouseGroups {
		so, err := s.createSingleSaleOrder(ctx, soTemplate, items, warehouse, sellingPriceList, taxTemplateName, taxes, req.TaxCategory, false)
		if err != nil {
			return createdOrders, gerror.Wrapf(err, "创建仓库 %s 的销售订单失败", warehouse)
		}
		createdOrders = append(createdOrders, so)
	}

	// 8. 直采物品单独 SO（自动提交）
	if len(dropshipItems) > 0 {
		so, err := s.createSingleSaleOrder(ctx, soTemplate, dropshipItems, "", sellingPriceList, taxTemplateName, taxes, req.TaxCategory, true)
		if err != nil {
			return createdOrders, gerror.Wrapf(err, "创建直采销售订单失败")
		}
		createdOrders = append(createdOrders, so)
	}

	return createdOrders, nil
}

// createSingleSaleOrder 根据 SO 模板创建单个销售订单
func (s *sBuying) createSingleSaleOrder(
	ctx context.Context,
	template *erp.SaleOrder,
	items []*erp.SaleOrderItem,
	setWarehouse string,
	sellingPriceList string,
	taxTemplateName string,
	taxes []interface{},
	taxCategory string,
	autoSubmit bool,
) (*erp.SaleOrder, error) {
	so := &erp.SaleOrder{
		Doctype:                    template.Doctype,
		Customer:                   template.Customer,
		CustomerName:               template.CustomerName,
		OrderType:                  template.OrderType,
		Company:                    template.Company,
		Currency:                   template.Currency,
		ConversionRate:             template.ConversionRate,
		PriceListCurrency:          template.PriceListCurrency,
		PlcConversionRate:          template.PlcConversionRate,
		DeliveryDate:               template.DeliveryDate,
		IsInternalCustomer:         template.IsInternalCustomer,
		RepresentsCompany:          template.RepresentsCompany,
		InterCompanyOrderReference: template.InterCompanyOrderReference,
		IgnorePricingRule:          template.IgnorePricingRule,
		Items:                      items,
		SetWarehouse:               setWarehouse,
		SellingPriceList:           sellingPriceList,
		TaxCategory:                taxCategory,
	}

	if taxTemplateName != "" {
		so.TaxesAndCharges = taxTemplateName
		so.Taxes = taxes
	}

	resp, err := service.Document().Create(ctx, erp.DocTypeSaleOrder, so)
	if err != nil {
		return nil, gerror.Wrapf(err, "创建销售订单失败")
	}

	resultSO := &erp.SaleOrder{}
	resp.GetJson("data").Scan(&resultSO)

	if autoSubmit {
		_, err = service.Document().ChangeDocStatus(ctx, erp.DocTypeSaleOrder, resultSO.Name, erp.DocstatusSubmitted)
		if err != nil {
			return nil, gerror.Wrapf(err, "提交销售订单失败")
		}
	}

	return resultSO, nil
}

// resolveSellingPriceList 获取销售价格表
func (s *sBuying) resolveSellingPriceList(ctx context.Context, reqPriceList string, company string) string {
	if reqPriceList != "" {
		return reqPriceList
	}
	defaultPriceList, err := service.PosPriceList().GetPosPriceListByCompany(ctx, company)
	if err != nil {
		g.Log().Warningf(ctx, "获取销售价格表失败，company: %s", company)
		defaultPriceList, err = service.PosPriceList().GetDefaultPosPriceList(ctx)
		if err != nil {
			return ""
		}
	}
	return defaultPriceList.SellingPriceList
}

// resolveSalesTax 获取销售税费模板和明细
func (s *sBuying) resolveSalesTax(ctx context.Context, reqTax string, company string) (string, []interface{}) {
	taxTemplateName := reqTax
	if taxTemplateName == "" {
		taxTemplateName = service.Accounts().GetDefaultSalesTaxTemplate(ctx, company)
	}
	if taxTemplateName == "" {
		return "", nil
	}
	taxDetails := service.Accounts().GetSalesTaxTemplateDetails(ctx, taxTemplateName)
	if len(taxDetails) > 0 {
		return taxTemplateName, s.convertSalesTaxesToInterface(taxDetails)
	}
	return taxTemplateName, nil
}
