package selling

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto/erp"
	dtoSelling "ttpos-bmp/app/ttpos-erp/internal/model/dto/selling"
	"ttpos-bmp/app/ttpos-erp/internal/service"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
)

// CreateSalesOrder 创建销售订单
// 参数：
//   - ctx: 上下文对象
//   - req: 销售订单信息
//
// 返回：
//   - *dtoSelling.SalesOrder: 创建后的销售订单信息
//   - error: 错误信息
func (s *sSelling) CreateSalesOrder(ctx context.Context, req *dtoSelling.SalesOrder) (*dtoSelling.SalesOrder, error) {
	// 校验销售订单信息
	if err := s.validateSalesOrder(req); err != nil {
		return nil, err
	}

	// 创建销售订单
	resp, err := service.Document().Create(ctx, erp.DocTypeSaleOrder, req)
	if err != nil {
		return nil, gerror.Wrapf(err, "创建销售订单失败")
	}

	// 解析响应数据
	salesOrder, err := s.parseSalesOrderResponse(resp)
	if err != nil {
		return nil, err
	}

	return salesOrder, nil
}

// UpdateSalesOrder 更新销售订单
// 参数：
//   - ctx: 上下文对象
//   - name: 销售订单名称
//   - req: 更新的销售订单信息
//
// 返回：
//   - *dtoSelling.SalesOrder: 更新后的销售订单信息
//   - error: 错误信息
func (s *sSelling) UpdateSalesOrder(ctx context.Context, name string, req *dtoSelling.SalesOrder) (*dtoSelling.SalesOrder, error) {
	if name == "" {
		return nil, gerror.New("销售订单名称不能为空")
	}

	// 校验销售订单信息
	if err := s.validateSalesOrder(req); err != nil {
		return nil, err
	}

	// 更新销售订单
	_, err := service.Document().Update(ctx, &erp.ErpReq{
		DocType: erp.DocTypeSaleOrder,
		Name:    name,
	}, req)
	if err != nil {
		return nil, gerror.Wrapf(err, "更新销售订单失败")
	}

	// 获取更新后的销售订单信息
	return s.GetSalesOrder(ctx, name)
}

// GetSalesOrder 获取销售订单信息
// 参数：
//   - ctx: 上下文对象
//   - name: 销售订单名称
//
// 返回：
//   - *dtoSelling.SalesOrder: 销售订单信息
//   - error: 错误信息
func (s *sSelling) GetSalesOrder(ctx context.Context, name string) (*dtoSelling.SalesOrder, error) {
	if name == "" {
		return nil, gerror.New("销售订单名称不能为空")
	}

	// 查询销售订单信息
	resp, err := service.Document().Get(ctx, &erp.ErpReq{
		DocType: erp.DocTypeSaleOrder,
		Name:    name,
	}, nil)
	if err != nil {
		return nil, gerror.Wrapf(err, "查询销售订单失败")
	}

	// 解析响应数据
	return s.parseSalesOrderResponse(resp)
}

// GetSalesOrderList 获取销售订单列表
// 参数：
//   - ctx: 上下文对象
//   - req: 查询参数
//
// 返回：
//   - []*dtoSelling.SalesOrder: 销售订单列表
//   - error: 错误信息
func (s *sSelling) GetSalesOrderList(ctx context.Context, req *dtoSelling.SalesOrderListReq) ([]*dtoSelling.SalesOrder, error) {
	// 构建查询过滤器
	filters := s.buildSalesOrderListFilters(ctx, req)

	// 查询销售订单列表
	resp, err := service.Document().List(ctx, &erp.ErpReq{
		DocType: erp.DocTypeSaleOrder,
	}, &erp.RequestParams{
		Fields: g.ArrayStr{
			"name", "customer", "customer_name", "transaction_date",
			"delivery_date", "grand_total", "status", "delivery_status",
			"billing_status", "per_delivered", "per_billed", "company",
		},
		Filters:    filters,
		Limit:      req.Limit,
		LimitStart: req.LimitStart,
		OrderBy: erp.OrderBy{
			Field: "modified",
			Order: "desc",
		},
	})
	if err != nil {
		return nil, gerror.Wrapf(err, "查询销售订单列表失败")
	}

	// 解析响应数据
	return s.parseSalesOrderListResponse(resp)
}

// CountSalesOrder 统计销售订单数量
// 参数：
//   - ctx: 上下文对象
//   - req: 查询参数
//
// 返回：
//   - int: 销售订单数量
//   - error: 错误信息
func (s *sSelling) CountSalesOrder(ctx context.Context, req *dtoSelling.SalesOrderListReq) (int, error) {
	// 构建查询过滤器
	filters := s.buildSalesOrderListFilters(ctx, req)

	// 统计销售订单数量
	count, err := service.Doctype().Count(ctx, &erp.ErpReq{
		DocType: erp.DocTypeSaleOrder,
	}, &erp.RequestParams{
		Filters: filters,
	})
	if err != nil {
		return 0, gerror.Wrapf(err, "统计销售订单数量失败")
	}

	return count, nil
}

// CancelSalesOrder 删除销售订单（取消订单）
// 参数：
//   - ctx: 上下文对象
//   - name: 销售订单名称
//
// 返回：
//   - error: 错误信息
func (s *sSelling) CancelSalesOrder(ctx context.Context, name string) error {
	if name == "" {
		return gerror.New("销售订单名称不能为空")
	}

	// 取消销售订单（将状态改为已取消）
	_, err := service.Document().ChangeDocStatus(ctx, erp.DocTypeSaleOrder, name, erp.DocstatusCancelled)
	if err != nil {
		return gerror.Wrapf(err, "取消销售订单失败")
	}

	return nil
}

// SubmitSalesOrder 提交销售订单
// 参数：
//   - ctx: 上下文对象
//   - name: 销售订单名称
//
// 返回：
//   - *dtoSelling.SalesOrder: 提交后的销售订单信息
//   - error: 错误信息
func (s *sSelling) SubmitSalesOrder(ctx context.Context, name string) (*dtoSelling.SalesOrder, error) {
	if name == "" {
		return nil, gerror.New("销售订单名称不能为空")
	}

	// 提交销售订单
	_, err := service.Document().ChangeDocStatus(ctx, erp.DocTypeSaleOrder, name, erp.DocstatusSubmitted)
	if err != nil {
		return nil, gerror.Wrapf(err, "提交销售订单失败")
	}

	// 获取提交后的订单信息
	return s.GetSalesOrder(ctx, name)
}

// validateSalesOrder 校验销售订单信息
// 参数：
//   - req: 销售订单信息
//
// 返回：
//   - error: 错误信息
func (s *sSelling) validateSalesOrder(req *dtoSelling.SalesOrder) error {
	if req == nil {
		return gerror.New("销售订单信息不能为空")
	}
	if req.Customer == "" {
		return gerror.New("客户信息不能为空")
	}
	if req.Company == "" {
		return gerror.New("公司信息不能为空")
	}
	//if req.TransactionDate == "" {
	//	return gerror.New("交易日期不能为空")
	//}
	if req.DeliveryDate == "" {
		return gerror.New("交付日期不能为空")
	}
	if len(req.Items) == 0 {
		return gerror.New("订单商品不能为空")
	}
	return nil
}

// parseSalesOrderResponse 解析销售订单响应数据
// 参数：
//   - resp: 响应数据
//
// 返回：
//   - *dtoSelling.SalesOrder: 解析后的销售订单信息
//   - error: 错误信息
func (s *sSelling) parseSalesOrderResponse(resp *gjson.Json) (*dtoSelling.SalesOrder, error) {
	salesOrder := &dtoSelling.SalesOrder{}

	// 解析响应数据
	if err := gconv.Scan(resp.GetJson("data"), salesOrder); err != nil {
		return nil, gerror.Wrapf(err, "解析销售订单响应失败")
	}

	return salesOrder, nil
}

// parseSalesOrderListResponse 解析销售订单列表响应数据
// 参数：
//   - resp: 响应数据
//
// 返回：
//   - []*dtoSelling.SalesOrder: 解析后的销售订单列表
//   - error: 错误信息
func (s *sSelling) parseSalesOrderListResponse(resp *gjson.Json) ([]*dtoSelling.SalesOrder, error) {
	salesOrders := make([]*dtoSelling.SalesOrder, 0)

	// 解析响应数据
	dataArray := resp.GetJsons("data")
	for _, data := range dataArray {
		salesOrder := &dtoSelling.SalesOrder{}
		if err := gconv.Scan(data, salesOrder); err != nil {
			g.Log().Warning(context.Background(), "解析销售订单数据失败", "error", err)
			continue // 跳过解析失败的数据
		}
		salesOrders = append(salesOrders, salesOrder)
	}

	return salesOrders, nil
}

// buildSalesOrderListFilters 构建销售订单列表查询过滤器
// 参数：
//   - ctx: 上下文对象
//   - req: 查询参数
//
// 返回：
//   - [][]string: 过滤条件数组
func (s *sSelling) buildSalesOrderListFilters(ctx context.Context, req *dtoSelling.SalesOrderListReq) [][]string {
	filters := make([][]string, 0, 8) // 预分配容量，提高性能

	// 按销售订单名称过滤
	if req.Name != "" {
		filters = append(filters, []string{"name", "like", "%" + req.Name + "%"})
	}

	// 按客户过滤
	if req.Customer != "" {
		filters = append(filters, []string{"customer", "=", req.Customer})
	}

	// 按客户名称过滤
	if req.CustomerName != "" {
		filters = append(filters, []string{"customer_name", "like", "%" + req.CustomerName + "%"})
	}

	// 按公司过滤
	if req.Company != "" {
		filters = append(filters, []string{"company", "=", req.Company})
	}

	// 按公司缩写过滤
	if req.CompanyAbbr != "" {
		companyName, err := service.Company().GetCompanyNameWithAbbr(ctx, req.CompanyAbbr)
		if err == nil && companyName != "" {
			filters = append(filters, []string{"company", "=", companyName})
		}
	}

	// 按订单类型过滤
	if req.OrderType != "" {
		filters = append(filters, []string{"order_type", "=", req.OrderType})
	}

	// 按订单状态过滤
	if req.Status != "" {
		filters = append(filters, []string{"status", "=", req.Status})
	}

	// 按交付状态过滤
	if req.DeliveryStatus != "" {
		filters = append(filters, []string{"delivery_status", "=", req.DeliveryStatus})
	}

	// 按开票状态过滤
	if req.BillingStatus != "" {
		filters = append(filters, []string{"billing_status", "=", req.BillingStatus})
	}

	// 按交易日期范围过滤
	if req.TransactionDateStart != "" {
		filters = append(filters, []string{"transaction_date", ">=", req.TransactionDateStart})
	}
	if req.TransactionDateEnd != "" {
		filters = append(filters, []string{"transaction_date", "<=", req.TransactionDateEnd})
	}

	// 按文档状态过滤
	if req.Docstatus >= 0 {
		filters = append(filters, []string{"docstatus", "=", gconv.String(req.Docstatus)})
	}

	// 按是否内部客户过滤
	if req.IsInternalCustomer > 0 {
		filters = append(filters, []string{"is_internal_customer", "=", gconv.String(req.IsInternalCustomer)})
	}

	return filters
}
