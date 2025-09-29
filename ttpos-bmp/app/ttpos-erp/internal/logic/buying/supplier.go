package buying

import (
	"context"
	"strings"
	dto "ttpos-bmp/app/ttpos-erp/internal/model/dto/buying"

	"ttpos-bmp/app/ttpos-erp/api/buying"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto/erp"
	"ttpos-bmp/app/ttpos-erp/internal/service"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
)

type sSupplier struct{}

var (
	Supplier = new(sSupplier)
)

func init() {
	service.RegisterSupplier(Supplier)
}

// GetInnerSupplierList 获取内部供应商列表
// 参数：
//   - ctx: 上下文对象，用于传递请求范围的元数据
//   - req: 获取供应商列表请求参数，包含公司缩写编码
//
// 返回：
//   - res: 获取供应商列表响应，包含供应商基本信息列表
//   - err: 操作过程中产生的错误（若有）
func (s *sSupplier) GetInnerSupplierList(ctx context.Context, req *buying.GetSupplierListReq) (*buying.GetSupplierListResp, error) {
	// 参数验证
	if req.GetCompanyAbbr() == "" {
		return nil, gerror.New("公司缩写编码不能为空")
	}

	// 根据公司简称获取公司名称
	companyName, err := service.Company().GetCompanyNameWithAbbr(ctx, req.GetCompanyAbbr())
	if err != nil {
		g.Log().Error(ctx, "查询公司名称失败", g.Map{
			"company_abbr": req.GetCompanyAbbr(),
			"error":        err,
		})
		return nil, gerror.Wrapf(err, "查询公司名称失败")
	}

	// 构建过滤条件
	filters := [][]string{
		{"Allowed To Transact With", "company", "=", companyName},
	}

	// 调用ERP接口获取供应商列表
	resp, err := service.Document().List(ctx, &erp.ErpReq{
		DocType: erp.DocTypeSupplier,
	}, &erp.RequestParams{
		Fields:  g.ArrayStr{"name", "represents_company", "supplier_name"},
		Filters: filters,
	})
	if err != nil {
		g.Log().Error(ctx, "查询供应商列表失败", g.Map{
			"company_name": companyName,
			"error":        err,
		})
		return nil, gerror.Wrapf(err, "查询供应商列表失败")
	}

	// 解析供应商列表
	supplierList, err := s.parseInnerSupplierListResponse(resp.Bytes())
	if err != nil {
		return nil, gerror.Wrapf(err, "解析供应商列表响应失败")
	}

	g.Log().Debug(ctx, "获取内部供应商列表成功", g.Map{
		"company_name": companyName,
		"count":        len(supplierList),
	})

	return &buying.GetSupplierListResp{
		SupplierList: supplierList,
	}, nil
}

// parseInnerSupplierListResponse 解析内部供应商列表响应
func (s *sSupplier) parseInnerSupplierListResponse(data []byte) ([]*buying.SupplierInfo, error) {
	jsonData := gjson.New(data)

	// 解析供应商列表数据
	supplierListData := jsonData.Get("data")
	if supplierListData.IsNil() {
		return []*buying.SupplierInfo{}, nil
	}

	var supplierList []*buying.SupplierInfo
	if err := gconv.Structs(supplierListData, &supplierList); err != nil {
		return nil, gerror.Wrapf(err, "解析供应商列表失败")
	}

	// 手动构建供应商信息列表，确保字段映射正确
	result := make([]*buying.SupplierInfo, 0)
	dataArray := supplierListData.Array()

	for _, item := range dataArray {
		itemValue := gconv.Map(item)
		supplierInfo := &buying.SupplierInfo{
			SupplierName:      gconv.String(itemValue["supplier_name"]),
			Name:              gconv.String(itemValue["name"]),
			RepresentsCompany: gconv.String(itemValue["represents_company"]),
		}
		result = append(result, supplierInfo)
	}

	return result, nil
}

// AddSupplerTransactCompany 为供应商添加允许交易的公司
// 参数：
//   - ctx: 上下文对象，用于传递请求范围的元数据
//   - req: 添加供应商交易公司请求参数，包含供应商名称和公司缩写
//
// 返回：
//   - err: 操作过程中产生的错误（若有）
func (s *sSupplier) AddSupplerTransactCompany(ctx context.Context, req *dto.AddSupplerTransactCompanyReq) error {
	// 参数验证
	if req.Supplier == "" {
		return gerror.New("供应商名称不能为空")
	}
	if req.WithCompanyAbbr == "" {
		return gerror.New("公司缩写编码不能为空")
	}

	// 根据公司缩写获取公司名称
	companyName, err := service.Company().GetCompanyNameWithAbbr(ctx, req.WithCompanyAbbr)
	if err != nil {
		g.Log().Error(ctx, "查询公司名称失败", g.Map{
			"company_abbr": req.WithCompanyAbbr,
			"error":        err,
		})
		return gerror.Wrapf(err, "查询公司名称失败")
	}

	countCustomer := 0
	customerName := companyName + " - " + req.WithCompanyAbbr
	if countCustomer, err = service.Selling().CountCustomer(ctx, &erp.Customer{
		RepresentsCompany: companyName,
		CustomerType:      "Company",
	}); err != nil {
		return gerror.Wrapf(err, "查询内部客户失败")
	}

	//创建默认内部销售客户
	if countCustomer == 0 {
		if _, err = service.Selling().CreateCustomer(ctx, &erp.Customer{
			Name:               customerName,
			CustomerType:       "Company",
			IsInternalCustomer: 1,
			RepresentsCompany:  companyName,
		}); err != nil {
			return gerror.Wrapf(err, "创建内部客户失败")
		}
	}

	// 获取供应商信息
	supplier, err := s.GetSupplier(ctx, &buying.GetSupplierReq{
		Name: req.Supplier,
	})
	if err != nil {
		g.Log().Error(ctx, "查询供应商失败", g.Map{
			"supplier": req.Supplier,
			"error":    err,
		})
		return gerror.Wrapf(err, "查询供应商失败")
	}

	// 验证是否为内部供应商
	if !supplier.IsInternalSupplier {
		return gerror.New("非内部供应商，不能添加关联公司")
	}

	// 检查公司是否已存在
	for _, company := range supplier.Companies {
		if company.Company == companyName {
			g.Log().Warning(ctx, "公司已存在于供应商关联列表中", g.Map{
				"supplier":     req.Supplier,
				"company_name": companyName,
			})
			return nil
		}
	}

	// 构建新的公司关联列表
	companies := append(supplier.Companies, erp.AllowedToTransactWith{
		Company: companyName,
	})

	// 调用ERP接口更新供应商
	_, err = service.Document().Update(ctx, &erp.ErpReq{
		DocType: erp.DocTypeSupplier,
		Name:    supplier.Name,
	}, &erp.Supplier{
		Companies: companies,
	})
	if err != nil {
		g.Log().Error(ctx, "更新供应商失败", g.Map{
			"supplier":     req.Supplier,
			"company_name": companyName,
			"error":        err,
		})
		return gerror.Wrapf(err, "更新供应商失败")
	}

	g.Log().Info(ctx, "供应商关联公司添加成功", g.Map{
		"supplier":     req.Supplier,
		"company_name": companyName,
	})

	return nil
}

// CreateSupplier 创建供应商
// 参数：
//   - ctx: 上下文对象，用于传递请求范围的元数据
//   - req: 创建供应商请求参数，包含供应商详细信息
//
// 返回：
//   - res: 创建供应商响应，包含创建结果
//   - err: 操作过程中产生的错误（若有）
func (s *sSupplier) CreateSupplier(ctx context.Context, req *buying.CreateSupplierReq) (*buying.CreateSupplierResp, error) {
	// 参数验证
	if err := s.validateCreateSupplierReq(req); err != nil {
		return nil, gerror.Wrapf(err, "创建供应商参数验证失败")
	}

	// 构建供应商数据
	supplierData := s.buildSupplierData(ctx, req.Supplier)

	if supplierData.CustomAliasName == "" {
		supplierData.CustomAliasName = req.Supplier.SupplierName
	}
	//创建时，使用别名来组装供应商名称
	supplierData.Name = supplierData.CustomAliasName + " - " + req.Supplier.CompanyAbbr
	supplierData.SupplierName = supplierData.CustomAliasName + " - " + req.Supplier.CompanyAbbr

	// 调用ERP接口创建供应商
	resp, err := service.Document().Create(ctx, erp.DocTypeSupplier, supplierData)
	if err != nil {
		g.Log().Error(ctx, "创建供应商失败", g.Map{
			"supplier_name": req.Supplier.SupplierName,
			"error":         err,
		})
		return nil, gerror.Wrapf(err, "创建供应商失败")
	}

	// 解析响应数据
	result, err := s.parseCreateSupplierResponse(resp.Bytes())
	if err != nil {
		return nil, gerror.Wrapf(err, "解析创建供应商响应失败")
	}
	//
	//links := make([]*crm.DynamicLinkInfo, 0)
	//links = append(links, &crm.DynamicLinkInfo{
	//	LinkDoctype: erp.DocTypeSupplier,
	//	LinkName:    req.Supplier.SupplierName,
	//})
	//
	////如果有联系人、联系地址、联系号码
	//if req.Supplier.Address != "" {
	//
	//	_, err := service.Crm().CreateAddress(ctx, &crm.CreateAddressReq{
	//		AddressTitle: req.Supplier.ContactName,
	//		AddressType:  "Current",
	//		AddressLine1: req.Supplier.Address,
	//		Country:      req.Supplier.Country,
	//		City:         req.Supplier.Country, //暂时用这个字段
	//		Links:        links,
	//	})
	//	if err != nil {
	//		return nil, gerror.Wrapf(err, "创建供应商地址失败")
	//	}
	//}
	//
	//if req.Supplier.ContactName != "" || req.Supplier.ContactPhone != "" {
	//	_, err := service.Crm().CreateContact(ctx, &crm.CreateContactReq{
	//		FirstName: req.Supplier.ContactName,
	//		Phone:     req.Supplier.ContactPhone,
	//		Links:     links,
	//	})
	//	if err != nil {
	//		return nil, gerror.Wrapf(err, "创建供应商地址失败")
	//	}
	//}

	g.Log().Info(ctx, "供应商创建成功", g.Map{
		"supplier_name": req.Supplier.SupplierName,
		"name":          result.Name,
	})

	return &buying.CreateSupplierResp{
		Supplier: result,
	}, nil
}

// GetSupplier 获取供应商详情
// 参数：
//   - ctx: 上下文对象，用于传递请求范围的元数据
//   - req: 获取供应商请求参数，包含供应商名称或ID
//
// 返回：
//   - res: 获取供应商响应，包含供应商详细信息
//   - err: 操作过程中产生的错误（若有）
func (s *sSupplier) GetSupplier(ctx context.Context, req *buying.GetSupplierReq) (*erp.Supplier, error) {

	// 调用ERP接口获取供应商详情
	resp, err := service.Document().Get(ctx, &erp.ErpReq{
		DocType: "Supplier",
		Name:    req.GetName(),
	}, &erp.RequestParams{})
	if err != nil {
		g.Log().Error(ctx, "获取供应商详情失败", g.Map{
			"name":  req.GetName(),
			"error": err,
		})
		return nil, gerror.Wrapf(err, "获取供应商详情失败")
	}

	// 解析响应数据
	supplier, err := s.parseGetSupplierResponse(resp.Bytes())
	if err != nil {
		return nil, gerror.Wrapf(err, "解析获取供应商响应失败")
	}

	g.Log().Debug(ctx, "获取供应商详情成功", g.Map{
		"name": req.GetName(),
	})

	return supplier, nil
}

// UpdateSupplier 更新供应商
// 参数：
//   - ctx: 上下文对象，用于传递请求范围的元数据
//   - req: 更新供应商请求参数，包含供应商更新信息
//
// 返回：
//   - res: 更新供应商响应，包含更新结果
//   - err: 操作过程中产生的错误（若有）
func (s *sSupplier) UpdateSupplier(ctx context.Context, req *buying.UpdateSupplierReq) (*buying.UpdateSupplierResp, error) {
	// 参数验证
	if err := s.validateUpdateSupplierReq(req); err != nil {
		return nil, gerror.Wrapf(err, "更新供应商参数验证失败")
	}

	// 构建更新数据
	supplierData := s.buildSupplierData(ctx, req.Supplier)

	// 调用ERP接口更新供应商
	resp, err := service.Document().Update(ctx, &erp.ErpReq{
		DocType: erp.DocTypeSupplier,
		Name:    req.Supplier.Name,
	}, supplierData)
	if err != nil {
		g.Log().Error(ctx, "更新供应商失败", g.Map{
			"supplier_name": req.Supplier.Name,
			"error":         err,
		})
		return nil, gerror.Wrapf(err, "更新供应商失败")
	}

	// 解析响应数据
	result, err := s.parseUpdateSupplierResponse(resp.Bytes())
	if err != nil {
		return nil, gerror.Wrapf(err, "解析更新供应商响应失败")
	}

	g.Log().Info(ctx, "供应商更新成功", g.Map{
		"name": req.Supplier.Name,
	})

	return &buying.UpdateSupplierResp{
		Supplier: result,
	}, nil
}

// DeleteSupplier 删除供应商
// 参数：
//   - ctx: 上下文对象，用于传递请求范围的元数据
//   - req: 删除供应商请求参数，包含供应商名称或ID
//
// 返回：
//   - res: 删除供应商响应，包含删除结果
//   - err: 操作过程中产生的错误（若有）
func (s *sSupplier) DeleteSupplier(ctx context.Context, req *buying.DeleteSupplierReq) (*buying.DeleteSupplierResp, error) {
	// 参数验证
	if req.GetName() == "" {
		return nil, gerror.New("供应商名称不能为空")
	}

	// 调用ERP接口删除供应商
	_, err := service.Document().Delete(ctx, &erp.ErpReq{
		DocType: erp.DocTypeSupplier,
		Name:    req.GetName(),
	})
	if err != nil {
		g.Log().Error(ctx, "删除供应商失败", g.Map{
			"name":  req.GetName(),
			"error": err,
		})
		return nil, gerror.Wrapf(err, "删除供应商失败")
	}

	g.Log().Info(ctx, "供应商删除成功", g.Map{
		"name": req.GetName(),
	})

	return &buying.DeleteSupplierResp{
		Success: true,
		Message: "供应商删除成功",
	}, nil
}

// ListSuppliers 获取供应商列表
// 参数：
//   - ctx: 上下文对象，用于传递请求范围的元数据
//   - req: 获取供应商列表请求参数，包含分页和过滤条件
//
// 返回：
//   - res: 获取供应商列表响应，包含供应商列表和分页信息
//   - err: 操作过程中产生的错误（若有）
func (s *sSupplier) ListSuppliers(ctx context.Context, req *buying.ListSuppliersReq) (*buying.ListSuppliersResp, error) {
	// 构建过滤条件
	filters := s.buildSupplierFilters(ctx, req)

	// 构建请求参数
	params := &erp.RequestParams{
		Fields:  g.ArrayStr{"name", "supplier_name", "country", "supplier_type", "disabled", "represents_company", "is_transporter", "is_internal_supplier", "custom_company", "custom_branch", "custom_aliasname"},
		Filters: filters,
	}

	// 添加分页参数
	if req.PageSize > 0 {
		params.Limit = int(req.PageSize)
		if req.PageNo > 0 {
			params.LimitStart = int((req.PageNo - 1) * req.PageSize)
		}
	}

	// 调用ERP接口获取供应商列表
	resp, err := service.Document().List(ctx, &erp.ErpReq{
		DocType: erp.DocTypeSupplier,
	}, params)
	if err != nil {
		g.Log().Error(ctx, "获取供应商列表失败", g.Map{
			"filters": filters,
			"error":   err,
		})
		return nil, gerror.Wrapf(err, "获取供应商列表失败")
	}

	// 解析响应数据
	j, err := gjson.DecodeToJson(resp.Bytes())
	if err != nil {
		return nil, gerror.Wrapf(err, "解析物品列表响应失败")
	}

	// 解析供应商列表
	supplierListData := j.GetJsons("data")

	subCompanyName := ""
	if len(req.SubCompanyAbbr) > 0 {
		subCompanyName, err = service.Company().GetCompanyNameWithAbbr(ctx, req.SubCompanyAbbr)
		if err != nil {
			return nil, gerror.Wrapf(err, "获取子公司名称失败,companyAbbr:%s", req.SubCompanyAbbr)
		}
	}
	var supplierList = make([]*buying.SupplierData, 0, len(supplierListData))

	for _, data := range supplierListData {
		//获取当前item 的所有 Item Permission List 判断是否有权限使用, 列表查询目前不支持 表格字段查询，只能每次遍历物品查询 对应的 custom_permission_rule
		if len(req.SubCompanyAbbr) > 0 {
			supplierInfo, err := s.GetSupplier(ctx, &buying.GetSupplierReq{
				Name: data.Get("name").String(),
			})
			if err != nil {
				g.Log().Errorf(ctx, "获取供应商信息失败,name:%s,err:%v", data.Get("name").String(), err)
				continue // 获取供应商信息失败，跳过该供应商
			}
			hasPermission, err := service.Permission().CheckPermission(ctx, supplierInfo.CustomPermissionRule, subCompanyName)
			if err != nil {
				g.Log().Errorf(ctx, "检查供应商权限失败,name:%s,err:%v", data.Get("name").String(), err)
				continue // 检查权限失败，跳过该供应商
			}
			if !hasPermission {
				continue // 当前公司无权限，跳过该供应商
			}
		}
		supplierList = append(supplierList, &buying.SupplierData{
			Name:               data.Get("name").String(),
			SupplierName:       data.Get("supplier_name").String(),
			Company:            data.Get("custom_company").String(),
			Branch:             data.Get("custom_branch").String(),
			Country:            data.Get("country").String(),
			SupplierType:       data.Get("supplier_type").String(),
			RepresentsCompany:  data.Get("represents_company").String(),
			IsTransporter:      data.Get("is_transporter").Bool(),
			IsInternalSupplier: data.Get("is_internal_supplier").Bool(),
			Disabled:           data.Get("disabled").Bool(),
			AliasName:          data.Get("custom_aliasname").String(),
		})
	}
	return &buying.ListSuppliersResp{
		Suppliers: supplierList,
	}, nil
}

// validateCreateSupplierReq 验证创建供应商请求参数
func (s *sSupplier) validateCreateSupplierReq(req *buying.CreateSupplierReq) error {
	if req == nil {
		return gerror.New("请求参数不能为空")
	}
	if req.GetSupplier() == nil {
		return gerror.New("供应商信息不能为空")
	}
	if req.GetSupplier().GetSupplierName() == "" {
		return gerror.New("供应商名称不能为空")
	}
	return nil
}

// validateUpdateSupplierReq 验证更新供应商请求参数
func (s *sSupplier) validateUpdateSupplierReq(req *buying.UpdateSupplierReq) error {
	if req == nil {
		return gerror.New("请求参数不能为空")
	}
	if req.GetSupplier() == nil {
		return gerror.New("供应商信息不能为空")
	}
	if req.GetSupplier().GetName() == "" {
		return gerror.New("供应商名称不能为空")
	}
	return nil
}

// buildSupplierData 构建供应商数据
func (s *sSupplier) buildSupplierData(ctx context.Context, supplier *buying.SupplierData) erp.Supplier {
	companyName := supplier.Company
	if supplier.CompanyAbbr != "" {
		var err error
		companyName, err = service.Company().GetCompanyNameWithAbbr(ctx, supplier.CompanyAbbr)
		if err != nil {
			g.Log().Error(ctx, "查询公司名称失败", g.Map{
				"company_abbr": supplier.CompanyAbbr,
				"error":        err,
			})
			companyName = supplier.Company
		}
	}

	return erp.Supplier{
		// 基础字段
		Name:              supplier.GetName(),
		SupplierName:      supplier.GetSupplierName(),
		Country:           supplier.GetCountry(),
		SupplierType:      supplier.GetSupplierType(),
		RepresentsCompany: supplier.GetRepresentsCompany(),

		// 状态标识
		IsTransporter:      supplier.GetIsTransporter(),
		IsInternalSupplier: supplier.GetIsInternalSupplier(),
		Disabled:           supplier.GetDisabled(),

		Company: companyName,
		Branch:  supplier.GetBranch(),

		CustomAliasName: supplier.GetAliasName(),
	}
}

// buildSupplierFilters 构建供应商过滤条件
func (s *sSupplier) buildSupplierFilters(ctx context.Context, req *buying.ListSuppliersReq) [][]string {
	var filters [][]string

	// 按供应商主键过滤
	if req.GetName() != "" {
		filters = append(filters, []string{"name", "=", req.GetName()})
	}
	// 按供应商名称过滤
	if strings.Contains(req.GetSupplierName(), "%") {
		filters = append(filters, []string{"supplier_name", "like", req.GetSupplierName()})
	} else if req.GetSupplierName() != "" {
		filters = append(filters, []string{"supplier_name", "=", req.GetSupplierName()})
	}

	// 按国家过滤
	if req.GetCountry() != "" {
		filters = append(filters, []string{"country", "=", req.GetCountry()})
	}

	// 按供应商类型过滤
	if req.GetSupplierType() != "" {
		filters = append(filters, []string{"supplier_type", "=", req.GetSupplierType()})
	}

	// 按分支过滤
	if req.GetBranch() != "" {
		filters = append(filters, []string{"custom_branch", "=", req.GetBranch()})
	}

	//按公司过滤
	if req.GetCompany() != "" {
		filters = append(filters, []string{"custom_company", "=", req.GetCompany()})
	}

	if req.GetCompanyAbbr() != "" {
		companyName, err := service.Company().GetCompanyNameWithAbbr(ctx, req.GetCompanyAbbr())
		if err != nil {
			g.Log().Error(ctx, "查询公司名称失败", g.Map{
				"company_abbr": req.GetCompanyAbbr(),
				"error":        err,
			})
		}
		filters = append(filters, []string{"custom_company", "=", companyName})
	}

	return filters
}

// parseCreateSupplierResponse 解析创建供应商响应
func (s *sSupplier) parseCreateSupplierResponse(data []byte) (*buying.SupplierData, error) {
	jsonData := gjson.New(data)

	// 解析供应商数据
	supplierData := jsonData.GetJson("data")
	if supplierData.IsNil() {
		return nil, gerror.New("响应数据为空")
	}

	var supplier buying.SupplierData
	if err := gconv.Struct(supplierData, &supplier); err != nil {
		return nil, gerror.Wrapf(err, "解析供应商数据失败")
	}

	supplier.AliasName = supplierData.Get("custom_aliasname").String()
	return &supplier, nil
}

// parseGetSupplierResponse 解析获取供应商响应
func (s *sSupplier) parseGetSupplierResponse(data []byte) (*erp.Supplier, error) {
	jsonData := gjson.New(data)

	// 解析供应商数据
	supplierData := jsonData.Get("data")
	if supplierData.IsNil() {
		return nil, gerror.New("供应商不存在")
	}

	var supplier erp.Supplier
	if err := gconv.Struct(supplierData, &supplier); err != nil {
		return nil, gerror.Wrapf(err, "解析供应商数据失败")
	}

	return &supplier, nil
}

// parseUpdateSupplierResponse 解析更新供应商响应
func (s *sSupplier) parseUpdateSupplierResponse(data []byte) (*buying.SupplierData, error) {
	jsonData := gjson.New(data)

	// 解析供应商数据
	supplierData := jsonData.Get("data")
	if supplierData.IsNil() {
		return nil, gerror.New("响应数据为空")
	}

	var supplier buying.SupplierData
	if err := gconv.Struct(supplierData, &supplier); err != nil {
		return nil, gerror.Wrapf(err, "解析供应商数据失败")
	}

	return &supplier, nil
}
