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
	filters := s.buildCustomerCountFilters(filter)

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
//   - filter: 客户过滤条件，可选
//
// 返回：
//   - [][]string: 过滤条件数组
func (s *sSelling) buildCustomerCountFilters(filter *erp.Customer) [][]string {
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

// AddCompanyToCustomer 将公司添加到客户的允许交易公司列表
// 参数：
//   - ctx: 上下文对象
//   - customer: 客户信息
//   - companyName: 要添加的公司名称
//
// 返回：
//   - error: 错误信息
func (s *sSelling) AddCompanyToCustomer(ctx context.Context, customer *erp.Customer, companyName string) error {
	// 先获取完整客户文档（ListCustomers 不返回 companies 子表，直接用会导致覆盖清空）
	fullCustomer, err := s.GetCustomer(ctx, customer.Name)
	if err != nil {
		return gerror.Wrapf(err, "获取客户完整信息失败")
	}

	// 查重：如果已包含该公司则跳过
	for _, company := range fullCustomer.Companies {
		if company.Company == companyName {
			return nil
		}
	}

	// 在完整列表上追加新公司
	companies := append(fullCustomer.Companies, &erp.AllowedToTransactWith{
		Company: companyName,
	})

	// 调用ERP接口更新客户
	_, err = service.Document().Update(ctx, &erp.ErpReq{
		DocType: erp.DocTypeCustomer,
		Name:    customer.Name,
	}, map[string]any{
		"companies": companies,
	})
	if err != nil {
		g.Log().Error(ctx, "更新客户失败", g.Map{
			"customer":     customer.Name,
			"company_name": companyName,
			"error":        err,
		})
		return gerror.Wrapf(err, "更新客户失败")
	}

	g.Log().Info(ctx, "客户关联公司添加成功", g.Map{
		"customer":     customer.Name,
		"company_name": companyName,
	})
	return nil
}

// ListCustomers 获取客户列表
// 参数：
//   - ctx: 上下文对象
//   - req: 获取客户列表请求参数，包含分页和过滤条件
//
// 返回：
//   - []*erp.Customer: 客户列表
//   - error: 错误信息
func (s *sSelling) ListCustomers(ctx context.Context, req *dtoSelling.ListCustomersReq) ([]*erp.Customer, error) {
	// 构建过滤条件
	filters := s.buildCustomerFilters(ctx, req)

	// 构建请求参数
	params := &erp.RequestParams{
		Fields: g.ArrayStr{
			"name", "customer_name", "customer_group", "customer_type",
			"is_internal_customer", "represents_company", "language",
			"disabled", "is_frozen",
		},
		Filters: filters,
	}

	// 添加分页参数
	if req.PageSize > 0 {
		params.Limit = int(req.PageSize)
		if req.PageNo > 0 {
			params.LimitStart = int((req.PageNo - 1) * req.PageSize)
		}
	}

	// 调用ERP接口获取客户列表
	resp, err := service.Document().List(ctx, &erp.ErpReq{
		DocType: erp.DocTypeCustomer,
	}, params)
	if err != nil {
		g.Log().Error(ctx, "获取客户列表失败", g.Map{
			"filters": filters,
			"error":   err,
		})
		return nil, gerror.Wrapf(err, "获取客户列表失败")
	}

	// 解析响应数据
	customerListData := resp.GetJsons("data")

	var customerList = make([]*erp.Customer, 0, len(customerListData))

	for _, data := range customerListData {
		customer := &erp.Customer{
			Name:               data.Get("name").String(),
			CustomerName:       data.Get("customer_name").String(),
			CustomerGroup:      data.Get("customer_group").String(),
			CustomerType:       data.Get("customer_type").String(),
			IsInternalCustomer: data.Get("is_internal_customer").Int(),
			RepresentsCompany:  data.Get("represents_company").String(),
			Language:           data.Get("language").String(),
			Disabled:           data.Get("disabled").Int(),
			IsFrozen:           data.Get("is_frozen").Int(),
		}
		customerList = append(customerList, customer)
	}

	g.Log().Debug(ctx, "获取客户列表成功", g.Map{
		"count": len(customerList),
	})

	return customerList, nil
}

// buildCustomerFilters 构建客户过滤条件
// 参数：
//   - ctx: 上下文对象
//   - req: 查询请求参数
//
// 返回：
//   - [][]string: 过滤条件数组
func (s *sSelling) buildCustomerFilters(ctx context.Context, req *dtoSelling.ListCustomersReq) [][]string {
	filters := make([][]string, 0, 8)

	// 按客户名称过滤
	if req.Name != "" {
		filters = append(filters, []string{"name", "=", req.Name})
	}

	// 按客户名称模糊查询
	if req.CustomerName != "" {
		filters = append(filters, []string{"customer_name", "like", "%" + req.CustomerName + "%"})
	}

	// 按客户类型过滤
	if req.CustomerType != "" {
		filters = append(filters, []string{"customer_type", "=", req.CustomerType})
	}

	// 按客户组过滤
	if req.CustomerGroup != "" {
		filters = append(filters, []string{"customer_group", "=", req.CustomerGroup})
	}

	// 按代表公司过滤
	if req.RepresentsCompany != "" {
		filters = append(filters, []string{"represents_company", "=", req.RepresentsCompany})
	}

	// 按公司缩写过滤
	if req.CompanyAbbr != "" {
		companyName, err := service.Company().GetCompanyNameWithAbbr(ctx, req.CompanyAbbr)
		if err == nil && companyName != "" {
			filters = append(filters, []string{"represents_company", "=", companyName})
		}
	}

	// 按是否内部客户过滤
	if req.IsInternalCustomer > 0 {
		filters = append(filters, []string{"is_internal_customer", "=", gconv.String(req.IsInternalCustomer)})
	}

	// 按语言过滤
	if req.Language != "" {
		filters = append(filters, []string{"language", "=", req.Language})
	}

	// 按禁用状态过滤（默认不包括禁用的）
	if req.IncludeDisabled {
		// 包含禁用的客户，不添加过滤条件
	} else {
		filters = append(filters, []string{"disabled", "!=", "1"})
	}

	// 按冻结状态过滤
	if req.IsFrozen > 0 {
		filters = append(filters, []string{"is_frozen", "=", gconv.String(req.IsFrozen)})
	}

	return filters
}
