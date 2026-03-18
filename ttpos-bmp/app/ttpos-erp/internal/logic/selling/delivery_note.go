package selling

import (
	"context"
	"strings"
	"ttpos-bmp/app/ttpos-erp/internal/consts"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto/delivery_note"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto/erp"
	"ttpos-bmp/app/ttpos-erp/internal/service"
	"ttpos-bmp/app/ttpos-erp/utility"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

var (
	DeliveryNote = &sDeliveryNote{}
)

type sDeliveryNote struct {
}

func init() {
	service.RegisterDeliveryNote(DeliveryNote)
}

// CreateDeliveryNote 创建送货单
// 参数：
//   - ctx: 上下文对象
//   - req: 创建送货单请求
//
// 返回：
//   - res: 创建送货单响应
//   - err: 错误信息
func (s *sDeliveryNote) CreateDeliveryNote(ctx context.Context, req *delivery_note.CreateDeliveryNoteReq) (res *delivery_note.CreateDeliveryNoteResp, err error) {
	// 获取公司信息
	company, err := service.Company().GetCompanyWithAbbr(ctx, req.CompanyAbbr)
	if err != nil {
		return nil, gerror.Wrapf(err, "获取公司信息失败")
	}

	// 设置仓库
	warehouseName := req.SetWarehouse
	if len(warehouseName) == 0 && len(req.Branch) > 0 {
		// 如果没有指定仓库但指定了分支，获取默认仓库
		warehouse, err := service.Warehouse().GetDefaultWarehouse(ctx, company.CompanyName, req.Branch)
		if err != nil {
			return nil, gerror.Wrapf(err, "获取默认仓库失败")
		}
		warehouseName = warehouse.Name
	}

	// 构建送货单数据
	data := &erp.DeliveryNote{
		NamingSeries: erp.DefaultDeliveryNoteSeries,
		Company:      company.CompanyName,
		Customer:     req.Customer,
		CustomerName: req.CustomerName,
		PostingDate:  req.PostingDate,
		SetWarehouse: warehouseName,
	}

	// 设置过账时间
	if len(req.PostingTime) > 0 {
		data.PostingTime = req.PostingTime
		data.SetPostingTime = 1
	} else {
		data.PostingTime = service.Setup().MustGetLocalDateTime(ctx, gtime.Now()).Format("H:i:s")
	}

	// 设置价格表
	if len(req.SellingPriceList) > 0 {
		data.SellingPriceList = req.SellingPriceList
	} else {
		data.SellingPriceList = consts.DefaultSellingPriceList
	}

	// 构建明细项目
	itemList := make([]*erp.DeliveryNoteItem, 0)
	for _, item := range req.Items {
		itemData := &erp.DeliveryNoteItem{
			ItemCode: item.ItemCode,
			Qty:      item.Qty,
			Uom:      item.Uom,
			Rate:     item.Rate,
		}

		// 如果明细项目指定了仓库，使用明细仓库
		if len(item.Warehouse) > 0 {
			itemData.Warehouse = item.Warehouse
		} else if len(warehouseName) > 0 {
			itemData.Warehouse = warehouseName
		}

		// 设置关联销售订单
		if len(item.AgainstSalesOrder) > 0 {
			itemData.AgainstSalesOrder = item.AgainstSalesOrder
			itemData.SoDetail = item.SoDetail
		}

		itemList = append(itemList, itemData)
	}
	data.Items = itemList

	// 如果送货项目为空，返回错误
	if len(itemList) == 0 {
		return nil, gerror.New("送货单明细项不能为空")
	}

	// 创建送货单据
	resp, err := service.Document().Create(ctx, erp.DocTypeDeliveryNote, data)
	if err != nil {
		return nil, gerror.Wrapf(err, "创建送货单据失败")
	}

	j := resp
	if !j.Contains("data") {
		return nil, gerror.New("创建送货单据失败：响应数据为空")
	}

	return &delivery_note.CreateDeliveryNoteResp{
		DeliveryNoteName: j.Get("data.name").String(),
	}, nil
}

// GetDeliveryNote 获取送货单详情
// 参数：
//   - ctx: 上下文对象
//   - req: 获取送货单详情请求
//
// 返回：
//   - res: 获取送货单详情响应
//   - err: 错误信息
func (s *sDeliveryNote) GetDeliveryNote(ctx context.Context, req *delivery_note.GetDeliveryNoteReq) (res *delivery_note.GetDeliveryNoteResp, err error) {
	// 参数验证
	if len(req.DeliveryNoteName) == 0 {
		return nil, gerror.New("送货单号不能为空")
	}

	// 查询送货单详情
	resp, err := service.Document().Get(ctx, &erp.ErpReq{
		DocType: erp.DocTypeDeliveryNote,
		Name:    req.DeliveryNoteName,
	}, nil)

	if err != nil {
		return nil, gerror.Wrapf(err, "查询送货单详情失败")
	}

	// 解析响应数据
	j := resp
	if !j.Contains("data") {
		return nil, gerror.New("查询送货单详情失败：响应数据为空")
	}

	deliveryNote := &delivery_note.DeliveryNoteData{}
	if err = j.GetJson("data").Scan(&deliveryNote); err != nil {
		return nil, gerror.Wrapf(err, "解析送货单详情失败")
	}

	return &delivery_note.GetDeliveryNoteResp{
		DeliveryNote: deliveryNote,
	}, nil
}

// GetDeliveryNoteList 获取送货单列表
// 参数：
//   - ctx: 上下文对象
//   - req: 获取送货单列表请求
//
// 返回：
//   - res: 获取送货单列表响应
//   - err: 错误信息
func (s *sDeliveryNote) GetDeliveryNoteList(ctx context.Context, req *delivery_note.GetDeliveryNoteListReq) (res *delivery_note.GetDeliveryNoteListResp, err error) {
	// 获取公司信息
	company, err := service.Company().GetCompanyWithAbbr(ctx, req.CompanyAbbr)
	if err != nil {
		return nil, gerror.Wrapf(err, "获取公司信息失败")
	}

	// 构建查询过滤器
	filters := [][]string{
		{"company", "=", company.CompanyName},
	}

	// 按客户过滤
	if len(req.Customer) > 0 {
		filters = append(filters, []string{"customer", utility.GetFilterOperator(req.Customer), req.Customer})
	}

	// 按仓库过滤
	if len(req.Warehouse) > 0 {
		filters = append(filters, []string{"set_warehouse", utility.GetFilterOperator(req.Warehouse), req.Warehouse})
	} else if len(req.Branch) > 0 {
		// 如果没有指定仓库但指定了分支，获取默认仓库
		warehouse, err := service.Warehouse().GetDefaultWarehouse(ctx, company.CompanyName, req.Branch)
		if err == nil && len(warehouse.Name) > 0 {
			filters = append(filters, []string{"set_warehouse", utility.GetFilterOperator(warehouse.Name), warehouse.Name})
		}
	}

	// 按状态过滤
	if len(req.Status) > 0 {
		filters = append(filters, []string{"status", "=", req.Status})
	}

	// 按日期范围过滤
	if len(req.FromDate) > 0 {
		filters = append(filters, []string{"posting_date", ">=", req.FromDate})
	}
	if len(req.ToDate) > 0 {
		filters = append(filters, []string{"posting_date", "<=", req.ToDate})
	}

	// 按采购订单号过滤（通过关联销售订单反查）
	if len(req.PoNo) > 0 {
		// 通过 po_no 查询关联的销售订单
		salesOrderNames, err := s.getSalesOrdersByPoNo(ctx, company.CompanyName, req.PoNo)
		if err != nil {
			g.Log().Warningf(ctx, "通过采购订单号查询销售订单失败: %v", err)
		}
		if len(salesOrderNames) > 0 {
			// 添加销售订单过滤条件（使用 in 查询）
			// 由于 ERPNext API 过滤器格式限制，这里使用 like 匹配第一个销售订单
			// 如果需要精确匹配多个销售订单，可以在应用层过滤
			filters = append(filters, []string{"name", "in", s.buildSalesOrderFilter(ctx, salesOrderNames)})
		} else {
			// 如果没有找到关联的销售订单，返回空结果
			return &delivery_note.GetDeliveryNoteListResp{
				DeliveryNoteList: make([]*delivery_note.DeliveryNoteData, 0),
			}, nil
		}
	}

	// 按销售订单号过滤（通过 Delivery Note Item.against_sales_order 过滤）
	// 支持逗号分隔的多个 SO 号（如 "SO-001,SO-002"）
	if len(req.SoNo) > 0 {
		if strings.Contains(req.SoNo, ",") {
			filters = append(filters, []string{"Delivery Note Item", "against_sales_order", "in", req.SoNo})
		} else {
			filters = append(filters, []string{"Delivery Note Item", "against_sales_order", "=", req.SoNo})
		}
	}

	// 设置查询限制
	limit := consts.Limit100
	if req.Limit > 0 && req.Limit <= 1000 {
		limit = int(req.Limit)
	}

	// 执行查询
	resp, err := service.Document().List(ctx, &erp.ErpReq{
		DocType: erp.DocTypeDeliveryNote,
	}, &erp.RequestParams{
		Fields: g.ArrayStr{
			"name", "company", "customer", "customer_name", "posting_date", "posting_time",
			"set_warehouse", "selling_price_list", "total_qty", "grand_total", "status", "docstatus",
		},
		Filters: filters,
		Limit:   limit,
	})

	if err != nil {
		return nil, gerror.Wrapf(err, "查询送货单列表失败")
	}

	// 解析响应数据
	j := resp
	var deliveryNoteList []*delivery_note.DeliveryNoteData
	if err = j.GetJson("data").Scan(&deliveryNoteList); err != nil {
		return nil, gerror.Wrapf(err, "解析送货单列表失败")
	}

	// 可选：获取明细项目
	if req.IncludeItems {
		for _, deliveryNote := range deliveryNoteList {
			items, _ := s.getDeliveryNoteItems(ctx, deliveryNote.Name)
			deliveryNote.Items = items
		}
	}

	// 如果有 po_no 过滤，在应用层进行精确过滤
	if len(req.PoNo) > 0 {
		deliveryNoteList = s.filterByPoNo(ctx, deliveryNoteList, req.PoNo)
	}

	res = &delivery_note.GetDeliveryNoteListResp{
		DeliveryNoteList: deliveryNoteList,
	}

	return res, nil
}

// getSalesOrdersByPoNo 通过采购订单号查询关联的销售订单
func (s *sDeliveryNote) getSalesOrdersByPoNo(ctx context.Context, company string, poNo string) ([]string, error) {
	// 查询销售订单，通过 po_no 字段匹配
	resp, err := service.Document().List(ctx, &erp.ErpReq{
		DocType: erp.DocTypeSaleOrder,
	}, &erp.RequestParams{
		Fields: g.ArrayStr{"name"},
		Filters: [][]string{
			{"company", "=", company},
			{"po_no", "=", poNo},
		},
		Limit: 100,
	})
	if err != nil {
		return nil, err
	}

	j := resp
	var salesOrders []struct {
		Name string `json:"name"`
	}
	if err = j.GetJson("data").Scan(&salesOrders); err != nil {
		return nil, err
	}

	names := make([]string, 0, len(salesOrders))
	for _, so := range salesOrders {
		names = append(names, so.Name)
	}
	return names, nil
}

// buildSalesOrderFilter 构建销售订单过滤条件
func (s *sDeliveryNote) buildSalesOrderFilter(ctx context.Context, salesOrderNames []string) string {
	// 由于需要查询 Delivery Note Item 中的 against_sales_order 字段
	// 这里返回第一个销售订单名称用于简单过滤
	// 更精确的过滤在 filterByPoNo 方法中进行
	if len(salesOrderNames) > 0 {
		return salesOrderNames[0]
	}
	return ""
}

// filterByPoNo 按采购订单号过滤送货单
func (s *sDeliveryNote) filterByPoNo(ctx context.Context, deliveryNotes []*delivery_note.DeliveryNoteData, poNo string) []*delivery_note.DeliveryNoteData {
	if len(poNo) == 0 || len(deliveryNotes) == 0 {
		return deliveryNotes
	}

	// 获取关联的销售订单
	salesOrderNames, err := s.getSalesOrdersByPoNo(ctx, "", poNo)
	if err != nil || len(salesOrderNames) == 0 {
		return make([]*delivery_note.DeliveryNoteData, 0)
	}

	// 构建销售订单名称集合
	soNameSet := make(map[string]bool)
	for _, name := range salesOrderNames {
		soNameSet[name] = true
	}

	// 过滤送货单
	result := make([]*delivery_note.DeliveryNoteData, 0)
	for _, dn := range deliveryNotes {
		// 获取送货单明细，检查是否关联到指定的销售订单
		items, err := s.getDeliveryNoteItems(ctx, dn.Name)
		if err != nil {
			continue
		}
		for _, item := range items {
			if soNameSet[item.AgainstSalesOrder] {
				dn.Items = items
				result = append(result, dn)
				break
			}
		}
	}

	return result
}

// getDeliveryNoteItems 获取送货单明细
// 参数：
//   - ctx: 上下文对象
//   - deliveryNoteName: 送货单名称
//
// 返回：
//   - items: 送货单明细项列表
//   - err: 错误信息
func (s *sDeliveryNote) getDeliveryNoteItems(ctx context.Context, deliveryNoteName string) ([]*delivery_note.DeliveryNoteItemData, error) {
	// 查询送货单详情
	resp, err := service.Document().Get(ctx, &erp.ErpReq{
		DocType: erp.DocTypeDeliveryNote,
		Name:    deliveryNoteName,
	}, nil)

	if err != nil {
		return nil, gerror.Wrapf(err, "查询送货单详情失败")
	}

	// 解析响应数据
	j := resp
	var items []*delivery_note.DeliveryNoteItemData
	if err = j.GetJson("data.items").Scan(&items); err != nil {
		return nil, gerror.Wrapf(err, "解析送货单明细失败")
	}

	return items, nil
}

// UpdateDeliveryNote 更新送货单
// 参数：
//   - ctx: 上下文对象
//   - req: 更新送货单请求
//
// 返回：
//   - res: 更新送货单响应
//   - err: 错误信息
func (s *sDeliveryNote) UpdateDeliveryNote(ctx context.Context, req *delivery_note.UpdateDeliveryNoteReq) (res *delivery_note.UpdateDeliveryNoteResp, err error) {
	// 参数验证
	if len(req.DeliveryNoteName) == 0 {
		return nil, gerror.New("送货单号不能为空")
	}

	// 构建更新数据
	data := make(map[string]interface{})

	// 可更新字段
	if len(req.Customer) > 0 {
		data["customer"] = req.Customer
	}
	if len(req.CustomerName) > 0 {
		data["customer_name"] = req.CustomerName
	}
	if len(req.PostingDate) > 0 {
		data["posting_date"] = req.PostingDate
	}
	if len(req.SetWarehouse) > 0 {
		data["set_warehouse"] = req.SetWarehouse
	}
	if len(req.Items) > 0 {
		data["items"] = req.Items
	}

	// 更新送货单据
	_, err = service.Document().Update(ctx, &erp.ErpReq{
		DocType: erp.DocTypeDeliveryNote,
		Name:    req.DeliveryNoteName,
	}, data)
	if err != nil {
		return nil, gerror.Wrapf(err, "更新送货单据失败")
	}

	return &delivery_note.UpdateDeliveryNoteResp{
		Message: "送货单更新成功",
	}, nil
}

// SubmitDeliveryNote 提交送货单
// 参数：
//   - ctx: 上下文对象
//   - deliveryNoteName: 送货单号
//
// 返回：
//   - err: 错误信息
func (s *sDeliveryNote) SubmitDeliveryNote(ctx context.Context, deliveryNoteName string) error {
	// 参数验证
	if len(deliveryNoteName) == 0 {
		return gerror.New("送货单号不能为空")
	}

	// 提交送货单据
	_, err := service.Document().ChangeDocStatus(ctx, erp.DocTypeDeliveryNote, deliveryNoteName, erp.DocstatusSubmitted)
	if err != nil {
		return gerror.Wrapf(err, "提交送货单据失败")
	}

	return nil
}

// CancelDeliveryNote 取消送货单
// 参数：
//   - ctx: 上下文对象
//   - deliveryNoteName: 送货单号
//
// 返回：
//   - err: 错误信息
func (s *sDeliveryNote) CancelDeliveryNote(ctx context.Context, deliveryNoteName string) error {
	// 参数验证
	if len(deliveryNoteName) == 0 {
		return gerror.New("送货单号不能为空")
	}

	// 取消送货单据
	_, err := service.Document().ChangeDocStatus(ctx, erp.DocTypeDeliveryNote, deliveryNoteName, erp.DocstatusCancelled)
	if err != nil {
		return gerror.Wrapf(err, "取消送货单据失败")
	}

	return nil
}

// DeleteDeliveryNote 删除送货单
// 参数：
//   - ctx: 上下文对象
//   - deliveryNoteName: 送货单号
//
// 返回：
//   - err: 错误信息
func (s *sDeliveryNote) DeleteDeliveryNote(ctx context.Context, deliveryNoteName string) error {
	// 参数验证
	if len(deliveryNoteName) == 0 {
		return gerror.New("送货单号不能为空")
	}

	// 删除送货单据
	_, err := service.Document().Delete(ctx, &erp.ErpReq{
		DocType: erp.DocTypeDeliveryNote,
		Name:    deliveryNoteName,
	})
	if err != nil {
		return gerror.Wrapf(err, "删除送货单据失败")
	}

	return nil
}

// CreateDeliveryNoteFromSaleOrder 从销售订单创建送货单
// 参数：
//   - ctx: 上下文对象
//   - req: 从销售订单创建送货单请求
//
// 返回：
//   - res: 送货单信息
//   - err: 错误信息
func (s *sDeliveryNote) CreateDeliveryNoteFromSaleOrder(ctx context.Context, req *delivery_note.CreateDeliveryNoteFromSaleOrderReq) (res *delivery_note.CreateDeliveryNoteFromSaleOrderResp, err error) {
	// 参数验证
	if len(req.SourceName) == 0 {
		return nil, gerror.New("销售订单名称不能为空")
	}

	// 调用 ERPNext 的 make_mapped_doc 方法，从销售订单生成送货单
	resp, err := service.Rpc().Execute(ctx, &erp.ErpReq{
		Method: erp.ApiMethodMakeMappedDoc,
	}, g.MapStrStr{
		"method":      erp.ApiMethodCreateDeliveryNote,
		"source_name": req.SourceName,
	})
	if err != nil {
		return nil, gerror.Wrapf(err, "从销售订单创建送货单失败")
	}

	// 解析响应数据
	j := resp
	deliveryNoteData := &erp.DeliveryNote{}
	if err := j.GetJson("data").Scan(&deliveryNoteData); err != nil {
		return nil, gerror.Wrapf(err, "解析送货单数据失败")
	}

	// 设置过账日期
	if req.PostingDate != "" {
		deliveryNoteData.PostingDate = req.PostingDate
	}

	// 设置源仓库
	if req.SourceWarehouse != "" {
		for i := range deliveryNoteData.Items {
			deliveryNoteData.Items[i].Warehouse = req.SourceWarehouse
		}
	}

	// 设置目标仓库
	if req.TargetWarehouse != "" {
		deliveryNoteData.SetWarehouse = req.TargetWarehouse
	} else if req.SourceWarehouse != "" {
		// 如果没有指定目标仓库但指定了源仓库，使用源仓库作为目标仓库
		deliveryNoteData.SetWarehouse = req.SourceWarehouse
	} else if deliveryNoteData.SetWarehouse == "" {
		// 获取默认仓库
		company, err := service.Company().GetCompany(ctx, deliveryNoteData.Company)
		if err == nil && company != nil {
			warehouse, err := service.Warehouse().GetDefaultWarehouse(ctx, company.CompanyName, "")
			if err == nil && len(warehouse.Name) > 0 {
				deliveryNoteData.SetWarehouse = warehouse.Name
			}
		}
	}

	// 设置销售价格表
	if req.SellingPriceList != "" {
		deliveryNoteData.SellingPriceList = req.SellingPriceList
	} else {
		// 使用默认销售价格表
		deliveryNoteData.SellingPriceList = consts.DefaultSellingPriceList
	}

	// 创建送货单
	resp, err = service.Document().Create(ctx, erp.DocTypeDeliveryNote, deliveryNoteData)
	if err != nil {
		return nil, gerror.Wrapf(err, "创建送货单失败")
	}

	// 解析响应数据
	j = resp
	if err := j.GetJson("data").Scan(&deliveryNoteData); err != nil {
		return nil, gerror.Wrapf(err, "解析创建后的送货单数据失败")
	}

	// 提交送货单
	_, err = service.Document().ChangeDocStatus(ctx, erp.DocTypeDeliveryNote, deliveryNoteData.Name, erp.DocstatusSubmitted)
	if err != nil {
		return nil, gerror.Wrapf(err, "提交送货单失败")
	}

	return &delivery_note.CreateDeliveryNoteFromSaleOrderResp{
		Name:       deliveryNoteData.Name,
		Status:     "To Bill",
		GrandTotal: deliveryNoteData.GrandTotal,
	}, nil
}
