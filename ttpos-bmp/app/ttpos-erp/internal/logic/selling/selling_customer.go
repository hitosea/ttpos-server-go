package selling

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto/erp"
	"ttpos-bmp/app/ttpos-erp/internal/service"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/util/gconv"
)

// CountCustomer 统计客户数量
// 参数：
//   - ctx: 上下文对象
//   - filter: 客户过滤条件，可选
//
// 返回：
//   - int: 客户数量
//   - error: 错误信息
func (s *sSelling) CountCustomer(ctx context.Context, filter *erp.Customer) (int, error) {
	// 构建查询过滤器
	filters := s.buildCustomerCountFilters(ctx, filter)

	count, err := service.Doctype().Count(ctx, &erp.ErpReq{
		DocType: erp.DocTypeCustomer,
	}, &erp.RequestParams{
		Filters: filters,
	})
	if err != nil {
		return 0, gerror.Wrapf(err, "统计客户数量失败")
	}
	return count, nil
}

// CreateCustomer 创建客户
// 参数：
//   - ctx: 上下文对象
//   - req: 客户信息
//
// 返回：
//   - *erp.Customer: 创建后的客户信息
//   - error: 错误信息
func (s *sSelling) CreateCustomer(ctx context.Context, req *erp.Customer) (*erp.Customer, error) {
	if err := s.validateCustomer(req); err != nil {
		return nil, err
	}

	resp, err := service.Document().Create(ctx, erp.DocTypeCustomer, req)
	if err != nil {
		return nil, gerror.Wrapf(err, "创建客户失败")
	}

	return s.parseCustomerResponse(resp.MustToJson())
}

// UpdateCustomer 更新客户
// 参数：
//   - ctx: 上下文对象
//   - name: 客户名称
//   - req: 更新的客户信息
//
// 返回：
//   - *erp.Customer: 更新后的客户信息
//   - error: 错误信息
func (s *sSelling) UpdateCustomer(ctx context.Context, name string, req *erp.Customer) (*erp.Customer, error) {
	if name == "" {
		return nil, gerror.New("客户名称不能为空")
	}
	if err := s.validateCustomer(req); err != nil {
		return nil, err
	}

	_, err := service.Document().Update(ctx, &erp.ErpReq{
		DocType: erp.DocTypeCustomer,
		Name:    name,
	}, req)
	if err != nil {
		return nil, gerror.Wrapf(err, "更新客户失败")
	}

	return s.GetCustomer(ctx, name)
}

// GetCustomer 获取客户信息
// 参数：
//   - ctx: 上下文对象
//   - name: 客户名称
//
// 返回：
//   - *erp.Customer: 客户信息
//   - error: 错误信息
func (s *sSelling) GetCustomer(ctx context.Context, name string) (*erp.Customer, error) {
	if name == "" {
		return nil, gerror.New("客户名称不能为空")
	}

	resp, err := service.Document().Get(ctx, &erp.ErpReq{
		DocType: erp.DocTypeCustomer,
		Name:    name,
	}, nil)
	if err != nil {
		return nil, gerror.Wrapf(err, "查询客户失败")
	}

	return s.parseCustomerResponse(resp.MustToJson())
}

// validateCustomer 校验客户信息
func (s *sSelling) validateCustomer(req *erp.Customer) error {
	if req == nil {
		return gerror.New("客户信息不能为空")
	}
	if req.CustomerName == "" {
		return gerror.New("客户名称不能为空")
	}
	if req.CustomerType == "" {
		return gerror.New("客户类型不能为空")
	}
	return nil
}

// parseCustomerResponse 解析客户响应数据
func (s *sSelling) parseCustomerResponse(data []byte) (*erp.Customer, error) {
	j, err := gjson.DecodeToJson(data)
	if err != nil {
		return nil, gerror.Wrapf(err, "解析客户响应失败")
	}

	customer := new(erp.Customer)
	if err := gconv.Struct(j.GetJson("data"), customer); err != nil {
		return nil, gerror.Wrapf(err, "转换客户数据失败")
	}

	return customer, nil
}

// buildCustomerCountFilters 构建客户数量查询过滤器
// 参数：
//   - ctx: 上下文对象
//   - filter: 客户过滤条件，可选
//
// 返回：
//   - [][]string: 过滤条件数组
func (s *sSelling) buildCustomerCountFilters(ctx context.Context, filter *erp.Customer) [][]string {
	filters := make([][]string, 0, 8) // 预分配容量，提高性能

	// 如果没有过滤条件，只返回基础过滤
	if filter == nil {
		filters = append(filters, []string{"disabled", "!=", "1"})
		return filters
	}

	// 按客户名称过滤
	if filter.CustomerName != "" {
		filters = append(filters, []string{"customer_name", "like", "%" + filter.CustomerName + "%"})
	}

	// 按客户类型过滤
	if filter.CustomerType != "" {
		filters = append(filters, []string{"customer_type", "=", filter.CustomerType})
	}

	// 按代表公司过滤
	if filter.RepresentsCompany != "" {
		filters = append(filters, []string{"represents_company", "=", filter.RepresentsCompany})
	}

	// 按语言过滤
	if filter.Language != "" {
		filters = append(filters, []string{"language", "=", filter.Language})
	}

	// 按内部客户标识过滤
	if filter.IsInternalCustomer > 0 {
		filters = append(filters, []string{"is_internal_customer", "=", gconv.String(filter.IsInternalCustomer)})
	}

	// 按禁用状态过滤
	if filter.Disabled > 0 {
		filters = append(filters, []string{"disabled", "=", gconv.String(filter.Disabled)})
	} else {
		// 默认只查询未禁用的客户
		filters = append(filters, []string{"disabled", "!=", "1"})
	}

	return filters
}
