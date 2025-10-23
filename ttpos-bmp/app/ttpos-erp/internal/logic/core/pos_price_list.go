package core

import (
	"context"

	"ttpos-bmp/app/ttpos-erp/internal/model/dto/erp"
	"ttpos-bmp/app/ttpos-erp/internal/service"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
)

type sPosPriceList struct{}

var (
	PosPriceList = new(sPosPriceList)
)

func init() {
	service.RegisterPosPriceList(PosPriceList)
}

// CreatePosPriceList 创建 POS 价格表规则
// 参数：
//   - ctx: 上下文对象，用于传递请求范围的元数据
//   - req: 创建价格表规则请求参数
//
// 返回：
//   - res: 创建的价格表规则数据
//   - err: 操作过程中产生的错误（若有）
func (s *sPosPriceList) CreatePosPriceList(ctx context.Context, req *erp.PosPriceList) (*erp.PosPriceList, error) {
	// 参数验证
	if err := s.validateCreateReq(req); err != nil {
		return nil, gerror.Wrapf(err, "创建价格表规则参数验证失败")
	}

	// 构建创建数据
	data := s.buildPosPriceListData(ctx, req)

	// 调用ERP接口创建价格表规则
	resp, err := service.Document().Create(ctx, erp.DocTypePosPriceList, data)
	if err != nil {
		g.Log().Error(ctx, "创建价格表规则失败", g.Map{
			"rule_code": req.RuleCode,
			"company":   req.Company,
			"error":     err,
		})
		return nil, gerror.Wrapf(err, "创建价格表规则失败")
	}

	// 解析响应数据
	result, err := s.parsePosPriceListResponse(resp.MustToJson())
	if err != nil {
		return nil, gerror.Wrapf(err, "解析创建价格表规则响应失败")
	}

	g.Log().Info(ctx, "价格表规则创建成功", g.Map{
		"rule_code": req.RuleCode,
		"name":      result.Name,
	})

	return result, nil
}

// GetPosPriceList 获取 POS 价格表规则详情
// 参数：
//   - ctx: 上下文对象，用于传递请求范围的元数据
//   - name: 价格表规则名称
//
// 返回：
//   - res: 价格表规则详细信息
//   - err: 操作过程中产生的错误（若有）
func (s *sPosPriceList) GetPosPriceList(ctx context.Context, name string) (*erp.PosPriceList, error) {
	// 参数验证
	if name == "" {
		return nil, gerror.New("价格表规则名称不能为空")
	}

	// 调用ERP接口获取价格表规则详情
	resp, err := service.Document().Get(ctx, &erp.ErpReq{
		DocType: erp.DocTypePosPriceList,
		Name:    name,
	}, &erp.RequestParams{})
	if err != nil {
		g.Log().Error(ctx, "获取价格表规则详情失败", g.Map{
			"name":  name,
			"error": err,
		})
		return nil, gerror.Wrapf(err, "获取价格表规则详情失败")
	}

	// 解析响应数据
	result, err := s.parsePosPriceListResponse(resp.MustToJson())
	if err != nil {
		return nil, gerror.Wrapf(err, "解析获取价格表规则响应失败")
	}

	g.Log().Debug(ctx, "获取价格表规则详情成功", g.Map{
		"name": name,
	})

	return result, nil
}

// UpdatePosPriceList 更新 POS 价格表规则
// 参数：
//   - ctx: 上下文对象，用于传递请求范围的元数据
//   - req: 更新价格表规则请求参数
//
// 返回：
//   - res: 更新后的价格表规则数据
//   - err: 操作过程中产生的错误（若有）
func (s *sPosPriceList) UpdatePosPriceList(ctx context.Context, req *erp.PosPriceList) (*erp.PosPriceList, error) {
	// 参数验证
	if err := s.validateUpdateReq(req); err != nil {
		return nil, gerror.Wrapf(err, "更新价格表规则参数验证失败")
	}

	// 构建更新数据
	data := s.buildPosPriceListData(ctx, req)

	// 调用ERP接口更新价格表规则
	resp, err := service.Document().Update(ctx, &erp.ErpReq{
		DocType: erp.DocTypePosPriceList,
		Name:    req.Name,
	}, data)
	if err != nil {
		g.Log().Error(ctx, "更新价格表规则失败", g.Map{
			"name":  req.Name,
			"error": err,
		})
		return nil, gerror.Wrapf(err, "更新价格表规则失败")
	}

	// 解析响应数据
	result, err := s.parsePosPriceListResponse(resp.MustToJson())
	if err != nil {
		return nil, gerror.Wrapf(err, "解析更新价格表规则响应失败")
	}

	g.Log().Info(ctx, "价格表规则更新成功", g.Map{
		"name": req.Name,
	})

	return result, nil
}

// DeletePosPriceList 删除 POS 价格表规则
// 参数：
//   - ctx: 上下文对象，用于传递请求范围的元数据
//   - name: 价格表规则名称
//
// 返回：
//   - err: 操作过程中产生的错误（若有）
func (s *sPosPriceList) DeletePosPriceList(ctx context.Context, name string) error {
	// 参数验证
	if name == "" {
		return gerror.New("价格表规则名称不能为空")
	}

	// 调用ERP接口删除价格表规则
	_, err := service.Document().Delete(ctx, &erp.ErpReq{
		DocType: erp.DocTypePosPriceList,
		Name:    name,
	})
	if err != nil {
		g.Log().Error(ctx, "删除价格表规则失败", g.Map{
			"name":  name,
			"error": err,
		})
		return gerror.Wrapf(err, "删除价格表规则失败")
	}

	g.Log().Info(ctx, "价格表规则删除成功", g.Map{
		"name": name,
	})

	return nil
}

// ListPosPriceLists 获取 POS 价格表规则列表
// 参数：
//   - ctx: 上下文对象，用于传递请求范围的元数据
//   - req: 查询价格表规则列表请求参数
//
// 返回：
//   - list: 价格表规则列表
//   - err: 操作过程中产生的错误（若有）
func (s *sPosPriceList) ListPosPriceLists(ctx context.Context, req *erp.ListPosPriceListsReq) ([]*erp.PosPriceList, error) {
	// 参数验证
	if req.PageNo < 1 {
		req.PageNo = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}

	// 构建过滤条件
	filterList := s.buildFilters(ctx, req)

	// 构建请求参数
	params := &erp.RequestParams{
		Fields:  g.ArrayStr{"name", "rule_code", "company", "buying_price_list", "selling_price_list", "disabled"},
		Filters: filterList,
	}

	// 添加分页参数
	if req.PageSize > 0 {
		params.Limit = req.PageSize
		if req.PageNo > 0 {
			params.LimitStart = (req.PageNo - 1) * req.PageSize
		}
	}

	// 调用ERP接口获取价格表规则列表
	resp, err := service.Document().List(ctx, &erp.ErpReq{
		DocType: erp.DocTypePosPriceList,
	}, params)
	if err != nil {
		g.Log().Error(ctx, "获取价格表规则列表失败", g.Map{
			"company_abbr": req.CompanyAbbr,
			"company":      req.Company,
			"rule_code":    req.RuleCode,
			"error":        err,
		})
		return nil, gerror.Wrapf(err, "获取价格表规则列表失败")
	}

	// 解析响应数据
	list, err := s.parsePosPriceListListResponse(resp.MustToJson())
	if err != nil {
		return nil, gerror.Wrapf(err, "解析价格表规则列表响应失败")
	}

	g.Log().Debug(ctx, "获取价格表规则列表成功", g.Map{
		"count":        len(list),
		"company_abbr": req.CompanyAbbr,
	})

	return list, nil
}

// GetDefaultPosPriceList 从配置文件获取默认价格表
// 参数：
//   - ctx: 上下文对象，用于传递请求范围的元数据
//
// 返回：
//   - res: 默认价格表配置
//   - err: 操作过程中产生的错误（若有）
func (s *sPosPriceList) GetDefaultPosPriceList(ctx context.Context) (*erp.PosPriceList, error) {
	// 从配置文件读取默认价格表
	buyingPriceList := g.Cfg().MustGet(ctx, "app.erpnext.core.pos_price_list.default.buying_price_list", "Buying - External").String()
	sellingPriceList := g.Cfg().MustGet(ctx, "app.erpnext.core.pos_price_list.default.selling_price_list", "Selling - Internal").String()

	// 验证配置是否存在
	if buyingPriceList == "" {
		err := gerror.New("配置文件中未设置默认采购价格表 (app.erpnext.core.pos_price_list.default.buying_price_list)")
		g.Log().Error(ctx, "获取默认采购价格表配置失败", err)
		return nil, err
	}

	if sellingPriceList == "" {
		err := gerror.New("配置文件中未设置默认销售价格表 (app.erpnext.core.pos_price_list.default.selling_price_list)")
		g.Log().Error(ctx, "获取默认销售价格表配置失败", err)
		return nil, err
	}

	// 构建返回对象
	result := &erp.PosPriceList{
		RuleCode:         "DEFAULT",
		Company:          "Default",
		BuyingPriceList:  buyingPriceList,
		SellingPriceList: sellingPriceList,
		Disabled:         0,
	}

	g.Log().Debug(ctx, "获取默认价格表成功", g.Map{
		"buying_price_list":  buyingPriceList,
		"selling_price_list": sellingPriceList,
	})

	return result, nil
}

// GetPosPriceListByCompany 根据公司获取默认价格表规则
// 参数：
//   - ctx: 上下文对象，用于传递请求范围的元数据
//   - company: 公司名称
//
// 返回：
//   - res: 价格表规则数据
//   - err: 操作过程中产生的错误（若有）
func (s *sPosPriceList) GetPosPriceListByCompany(ctx context.Context, company string) (*erp.PosPriceList, error) {
	// 参数验证
	if company == "" {
		return nil, gerror.New("公司名称不能为空")
	}

	// 构建过滤条件
	filters := [][]string{
		{"company", "=", company},
		{"disabled", "=", "0"},
	}

	// 调用ERP接口获取价格表规则列表
	resp, err := service.Document().List(ctx, &erp.ErpReq{
		DocType: erp.DocTypePosPriceList,
	}, &erp.RequestParams{
		Fields:  g.ArrayStr{"name", "rule_code", "company", "buying_price_list", "selling_price_list", "disabled"},
		Filters: filters,
		Limit:   1,
	})
	if err != nil {
		g.Log().Error(ctx, "获取价格表规则失败", g.Map{
			"company": company,
			"error":   err,
		})
		return nil, gerror.Wrapf(err, "获取价格表规则失败")
	}

	// 解析响应数据
	list, err := s.parsePosPriceListListResponse(resp.MustToJson())
	if err != nil {
		return nil, gerror.Wrapf(err, "解析价格表规则响应失败")
	}

	if len(list) == 0 {
		return nil, gerror.Newf("公司 %s 未配置价格表规则", company)
	}

	g.Log().Debug(ctx, "获取价格表规则成功", g.Map{
		"company": company,
		"name":    list[0].Name,
	})

	return list[0], nil
}

// validateCreateReq 验证创建请求参数
func (s *sPosPriceList) validateCreateReq(req *erp.PosPriceList) error {
	if req == nil {
		return gerror.New("请求参数不能为空")
	}
	if req.RuleCode == "" {
		return gerror.New("规则代码不能为空")
	}
	if req.Company == "" {
		return gerror.New("公司不能为空")
	}
	if req.BuyingPriceList == "" {
		return gerror.New("采购价格表不能为空")
	}
	if req.SellingPriceList == "" {
		return gerror.New("销售价格表不能为空")
	}
	return nil
}

// validateUpdateReq 验证更新请求参数
func (s *sPosPriceList) validateUpdateReq(req *erp.PosPriceList) error {
	if req == nil {
		return gerror.New("请求参数不能为空")
	}
	if req.Name == "" {
		return gerror.New("价格表规则名称不能为空")
	}
	return nil
}

// buildPosPriceListData 构建价格表规则数据
func (s *sPosPriceList) buildPosPriceListData(ctx context.Context, req *erp.PosPriceList) erp.PosPriceList {
	return erp.PosPriceList{
		Name:             req.Name,
		RuleCode:         req.RuleCode,
		Company:          req.Company,
		BuyingPriceList:  req.BuyingPriceList,
		SellingPriceList: req.SellingPriceList,
		Disabled:         req.Disabled,
	}
}

// buildFilters 构建过滤条件
func (s *sPosPriceList) buildFilters(ctx context.Context, req *erp.ListPosPriceListsReq) [][]string {
	var filterList [][]string

	// 按规则代码过滤
	if req.RuleCode != "" {
		filterList = append(filterList, []string{"rule_code", "=", req.RuleCode})
	}

	// 按公司缩写过滤（需要先查询公司名称）
	if req.CompanyAbbr != "" {
		companyName, err := service.Company().GetCompanyNameWithAbbr(ctx, req.CompanyAbbr)
		if err != nil {
			g.Log().Error(ctx, "查询公司名称失败", g.Map{
				"company_abbr": req.CompanyAbbr,
				"error":        err,
			})
		} else {
			filterList = append(filterList, []string{"company", "=", companyName})
		}
	} else if req.Company != "" {
		// 按公司名称过滤
		filterList = append(filterList, []string{"company", "=", req.Company})
	}

	// 按禁用状态过滤
	if req.Disabled >= 0 {
		filterList = append(filterList, []string{"disabled", "=", gconv.String(req.Disabled)})
	}

	return filterList
}

// parsePosPriceListResponse 解析单个价格表规则响应
func (s *sPosPriceList) parsePosPriceListResponse(data []byte) (*erp.PosPriceList, error) {
	jsonData := gjson.New(data)

	// 解析价格表规则数据
	priceListData := jsonData.Get("data")
	if priceListData.IsNil() {
		return nil, gerror.New("响应数据为空")
	}

	var priceList erp.PosPriceList
	if err := gconv.Struct(priceListData, &priceList); err != nil {
		return nil, gerror.Wrapf(err, "解析价格表规则数据失败")
	}

	return &priceList, nil
}

// parsePosPriceListListResponse 解析价格表规则列表响应
func (s *sPosPriceList) parsePosPriceListListResponse(data []byte) ([]*erp.PosPriceList, error) {
	jsonData := gjson.New(data)

	// 解析价格表规则列表数据
	listData := jsonData.Get("data")
	if listData.IsNil() {
		return []*erp.PosPriceList{}, nil
	}

	var list []*erp.PosPriceList
	dataArray := listData.Array()

	for _, item := range dataArray {
		itemValue := gconv.Map(item)
		priceList := &erp.PosPriceList{
			Name:             gconv.String(itemValue["name"]),
			RuleCode:         gconv.String(itemValue["rule_code"]),
			Company:          gconv.String(itemValue["company"]),
			BuyingPriceList:  gconv.String(itemValue["buying_price_list"]),
			SellingPriceList: gconv.String(itemValue["selling_price_list"]),
			Disabled:         gconv.Int(itemValue["disabled"]),
		}
		list = append(list, priceList)
	}

	return list, nil
}
