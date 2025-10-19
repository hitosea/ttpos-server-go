package stock

import (
	"context"
	"strings"
	"ttpos-bmp/app/ttpos-erp/api/warehouse"
	"ttpos-bmp/app/ttpos-erp/internal/consts"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto/erp"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto/setup"
	"ttpos-bmp/app/ttpos-erp/internal/service"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

var (
	Warehouse = new(sWarehouse)
)

type sWarehouse struct{}

func init() {
	service.RegisterWarehouse(Warehouse)
}

// CreateWarehouse 创建仓库
// 参数：ctx 上下文，req 包含 shop_name、company_abbr
// 返回：仓库名称，错误信息
func (s *sWarehouse) CreateWarehouse(ctx context.Context, req *setup.CreateWarehouseInp) (warehouseName string, err error) {
	// 校验参数
	if err := s.validateCreateWarehouseReq(req); err != nil {
		return "", err
	}

	// 获取公司信息
	company, err := service.Company().GetCompanyWithAbbr(ctx, req.CompanyAbbr)
	if err != nil {
		return "", gerror.Wrapf(err, "获取公司信息失败")
	}

	// 生成仓库名称
	warehouseName = s.generateWarehouseName(req)

	// 创建仓库
	if err := s.createWarehouseDocument(ctx, warehouseName, req, company.CompanyName); err != nil {
		return "", err
	}

	return warehouseName, nil
}

// validateCreateWarehouseReq 验证创建仓库请求参数
func (s *sWarehouse) validateCreateWarehouseReq(req *setup.CreateWarehouseInp) error {
	if len(req.Branch) == 0 {
		return gerror.New("分支机构不能为空")
	}
	if len(req.WhType) == 0 {
		return gerror.New("仓库类型不能为空")
	}
	if len(req.AliasName) == 0 {
		return gerror.New("仓库别名不能为空")
	}
	if len(req.CompanyAbbr) == 0 {
		return gerror.New("公司简称不能为空")
	}
	return nil
}

// generateWarehouseName 生成仓库名称
// 仓库名称规则：[分支名]-[仓库类型]-[仓库名称]
func (s *sWarehouse) generateWarehouseName(req *setup.CreateWarehouseInp) string {
	return strings.Join([]string{req.Branch, req.WhType, req.AliasName}, "-")
}

// createWarehouseDocument 创建仓库文档
func (s *sWarehouse) createWarehouseDocument(ctx context.Context, warehouseName string, req *setup.CreateWarehouseInp, companyName string) error {
	// 构建仓库创建参数
	warehousePayload := g.Map{
		"warehouse_name":   warehouseName, // 仓库名称
		"custom_branch":    req.Branch,    // 分支机构
		"custom_aliasname": req.AliasName, // 仓库别名
		"company":          companyName,   // 公司名称
	}

	// 特殊处理中转仓库类型
	if req.WhType == "Transit" {
		warehousePayload["warehouse_type"] = "Transit"
	}

	// 创建仓库文档
	if _, err := service.Document().Create(ctx, "Warehouse", warehousePayload); err != nil {
		g.Log().Error(ctx, "创建仓库失败", err)
		return gerror.Wrapf(err, "创建仓库失败")
	}

	return nil
}

// GetWarehouseList 获取仓库列表
// 根据查询条件过滤并返回仓库信息列表
func (s *sWarehouse) GetWarehouseList(ctx context.Context, req *warehouse.GetWarehouseListReq) (res *warehouse.GetWarehouseListResp, err error) {
	// 构建查询过滤器
	filters := s.buildWarehouseListFilters(ctx, req)

	// 查询仓库列表
	warehouseList, err := s.queryWarehouseList(ctx, filters, req)
	if err != nil {
		return nil, gerror.Wrapf(err, "查询仓库列表失败")
	}

	return &warehouse.GetWarehouseListResp{
		WarehouseList: warehouseList,
	}, nil
}

func (s *sWarehouse) GetDefaultWarehouse(ctx context.Context, company string, branch string) (res *warehouse.WarehouseInfo, err error) {
	warehouseList, err := s.GetWarehouseList(ctx, &warehouse.GetWarehouseListReq{
		Company:       company,
		Branch:        branch,
		WarehouseType: "",
		AliasName:     "Default",
	})
	if err != nil {
		return nil, err
	}
	if len(warehouseList.WarehouseList) == 0 {
		//获取默认 Stores 仓库
		warehouseList, err = s.GetWarehouseList(ctx, &warehouse.GetWarehouseListReq{
			Company:       company,
			WarehouseName: "Stores",
		})
		if len(warehouseList.WarehouseList) == 0 {
			return nil, gerror.New("默认仓库不存在")
		}
	}
	return warehouseList.WarehouseList[0], nil
}

// buildWarehouseListFilters 构建仓库列表查询过滤器
func (s *sWarehouse) buildWarehouseListFilters(ctx context.Context, req *warehouse.GetWarehouseListReq) [][]string {
	filters := make([][]string, 0)

	// 按分支机构过滤
	if len(req.Branch) > 0 {
		filters = append(filters, g.ArrayStr{"custom_branch", "=", req.Branch})
	}
	// 按仓库名称过滤
	if len(req.Name) > 0 {
		filters = append(filters, g.ArrayStr{"name", "=", req.Name})
	}

	// 按仓库名称过滤
	if len(req.WarehouseName) > 0 {
		filters = append(filters, g.ArrayStr{"warehouse_name", "=", req.WarehouseName})
	}

	// 按公司过滤
	if len(req.Company) > 0 {
		filters = append(filters, g.ArrayStr{"company", "=", req.Company})
	}

	// 按仓库类型过滤
	if len(req.WarehouseType) > 0 {
		filters = append(filters, g.ArrayStr{"warehouse_type", "=", req.WarehouseType})
	}

	// 按仓库别名过滤
	if len(req.AliasName) > 0 {
		filters = append(filters, g.ArrayStr{"custom_aliasname", "=", req.AliasName})
	}

	if len(req.CompanyAbbr) > 0 {
		companyName, _ := service.Company().GetCompanyNameWithAbbr(ctx, req.CompanyAbbr)
		filters = append(filters, []string{"company", "=", companyName})
	}

	return filters
}

// queryWarehouseList 执行仓库列表查询
func (s *sWarehouse) queryWarehouseList(ctx context.Context, filters [][]string, req *warehouse.GetWarehouseListReq) ([]*warehouse.WarehouseInfo, error) {
	// 查询仓库列表
	resp, err := service.Document().List(ctx, &erp.ErpReq{
		DocType: erp.DocTypeWarehouse,
	}, &erp.RequestParams{
		Fields:  g.ArrayStr{"name", "warehouse_name", "warehouse_type", "custom_branch", "custom_aliasname", "company", "disabled"},
		Filters: filters,
		Limit:   consts.Limit999,
	})

	if err != nil {
		return nil, err
	}

	// 解析响应数据
	j := resp

	// 转换为仓库信息列表
	warehouseList := make([]*warehouse.WarehouseInfo, 0)
	dataArray := j.GetJsons("data")

	for _, data := range dataArray {
		warehouseInfo := &warehouse.WarehouseInfo{
			Name:          data.Get("name").String(),
			Branch:        data.Get("custom_branch").String(),
			Company:       data.Get("company").String(),
			WarehouseName: data.Get("warehouse_name").String(),
			WarehouseType: data.Get("warehouse_type").String(),
			AliasName:     data.Get("custom_aliasname").String(),
			Disabled:      data.Get("disabled").Bool(),
			CompanyAbbr:   req.CompanyAbbr,
		}
		warehouseList = append(warehouseList, warehouseInfo)
	}

	return warehouseList, nil
}

// GetWarehouse 获取单个仓库详情
// 根据仓库名称获取仓库详细信息
func (s *sWarehouse) GetWarehouse(ctx context.Context, req *warehouse.GetWarehouseReq) (res *warehouse.GetWarehouseResp, err error) {
	// 参数验证
	if err := s.validateGetWarehouseReq(req); err != nil {
		return nil, err
	}

	// 查询仓库详情
	warehouseInfo, err := s.queryWarehouseDetail(ctx, req.Name)
	if err != nil {
		return nil, gerror.Wrapf(err, "查询仓库详情失败")
	}

	return &warehouse.GetWarehouseResp{
		Warehouse: warehouseInfo,
	}, nil
}

// UpdateWarehouse 更新仓库信息
// 根据仓库名称更新仓库的相关信息
func (s *sWarehouse) UpdateWarehouse(ctx context.Context, req *warehouse.UpdateWarehouseReq) (res *warehouse.UpdateWarehouseResp, err error) {
	// 参数验证
	if err := s.validateUpdateWarehouseReq(req); err != nil {
		return nil, err
	}

	// 检查仓库是否存在
	if err := s.checkWarehouseExists(ctx, req.Warehouse.Name); err != nil {
		return nil, err
	}

	// 更新仓库信息
	updatedWarehouse, err := s.updateWarehouseDocument(ctx, req.Warehouse)
	if err != nil {
		return nil, err
	}

	return &warehouse.UpdateWarehouseResp{
		Warehouse: updatedWarehouse,
	}, nil
}

// DeleteWarehouse 删除仓库
// 根据仓库名称删除指定仓库
func (s *sWarehouse) DeleteWarehouse(ctx context.Context, req *warehouse.DeleteWarehouseReq) (res *warehouse.DeleteWarehouseResp, err error) {
	// 参数验证
	if err := s.validateDeleteWarehouseReq(req); err != nil {
		return nil, gerror.Wrapf(err, "参数验证失败")
	}

	// 检查仓库是否存在
	if err := s.checkWarehouseExists(ctx, req.Name); err != nil {
		return nil, err
	}

	// 检查仓库是否可以删除（如是否有库存等）
	if err := s.checkWarehouseDeletable(ctx, req.Name); err != nil {
		return nil, err
	}

	// 删除仓库文档
	if err := s.deleteWarehouseDocument(ctx, req.Name); err != nil {
		return nil, err
	}

	return &warehouse.DeleteWarehouseResp{
		Success: true,
		Message: "仓库删除成功",
	}, nil
}

// validateGetWarehouseReq 验证获取仓库详情请求参数
func (s *sWarehouse) validateGetWarehouseReq(req *warehouse.GetWarehouseReq) error {
	if req == nil {
		return gerror.New("请求参数不能为空")
	}
	if len(req.Name) == 0 {
		return gerror.New("仓库名称不能为空")
	}
	return nil
}

// validateUpdateWarehouseReq 验证更新仓库请求参数
func (s *sWarehouse) validateUpdateWarehouseReq(req *warehouse.UpdateWarehouseReq) error {
	if req == nil {
		return gerror.New("请求参数不能为空")
	}
	if req.Warehouse == nil {
		return gerror.New("仓库信息不能为空")
	}
	if len(req.Warehouse.Name) == 0 {
		return gerror.New("仓库名称不能为空")
	}
	return nil
}

// validateDeleteWarehouseReq 验证删除仓库请求参数
func (s *sWarehouse) validateDeleteWarehouseReq(req *warehouse.DeleteWarehouseReq) error {
	if req == nil {
		return gerror.New("请求参数不能为空")
	}
	if len(req.Name) == 0 {
		return gerror.New("仓库名称不能为空")
	}
	return nil
}

// queryWarehouseDetail 查询仓库详情
func (s *sWarehouse) queryWarehouseDetail(ctx context.Context, warehouseName string) (*warehouse.WarehouseInfo, error) {
	// 查询仓库详情
	resp, err := service.Document().Get(ctx, &erp.ErpReq{
		DocType: "Warehouse",
		Name:    warehouseName,
	}, &erp.RequestParams{
		Fields: g.ArrayStr{"name", "warehouse_name", "warehouse_type", "custom_branch", "custom_aliasname", "company", "disabled"},
	})

	if err != nil {
		return nil, gerror.Wrapf(err, "查询仓库文档失败")
	}

	// 解析响应数据
	j := resp

	dataJson := j.GetJson("data")
	if dataJson == nil {
		return nil, gerror.New("仓库不存在")
	}

	// 转换为仓库信息
	warehouseInfo := &warehouse.WarehouseInfo{
		Name:          dataJson.Get("name").String(),
		Branch:        dataJson.Get("custom_branch").String(),
		Company:       dataJson.Get("company").String(),
		WarehouseName: dataJson.Get("warehouse_name").String(),
		WarehouseType: dataJson.Get("warehouse_type").String(),
		AliasName:     dataJson.Get("custom_aliasname").String(),
		Disabled:      dataJson.Get("disabled").Bool(),
	}

	return warehouseInfo, nil
}

// checkWarehouseExists 检查仓库是否存在
func (s *sWarehouse) checkWarehouseExists(ctx context.Context, warehouseName string) error {
	_, err := s.queryWarehouseDetail(ctx, warehouseName)
	if err != nil {
		return gerror.Wrapf(err, "仓库不存在或查询失败")
	}
	return nil
}

// updateWarehouseDocument 更新仓库文档
func (s *sWarehouse) updateWarehouseDocument(ctx context.Context, warehouseInfo *warehouse.WarehouseInfo) (*warehouse.WarehouseInfo, error) {
	// 构建更新参数
	updateData := g.Map{}

	// 只更新非空字段
	if warehouseInfo.WarehouseName != "" {
		updateData["warehouse_name"] = warehouseInfo.WarehouseName
	}
	if warehouseInfo.Branch != "" {
		updateData["custom_branch"] = warehouseInfo.Branch
	}
	if warehouseInfo.AliasName != "" {
		updateData["custom_aliasname"] = warehouseInfo.AliasName
	}
	if warehouseInfo.Company != "" {
		updateData["company"] = warehouseInfo.Company
	}
	if warehouseInfo.WarehouseType != "" {
		updateData["warehouse_type"] = warehouseInfo.WarehouseType
	}
	// 禁用状态可以更新为 false
	updateData["disabled"] = warehouseInfo.Disabled

	// 更新仓库文档
	if _, err := service.Document().Update(ctx, &erp.ErpReq{
		DocType: erp.DocTypeWarehouse,
		Name:    warehouseInfo.Name,
	}, updateData); err != nil {
		g.Log().Error(ctx, "更新仓库失败", err)
		return nil, gerror.Wrapf(err, "更新仓库失败")
	}

	// 返回更新后的仓库信息
	return s.queryWarehouseDetail(ctx, warehouseInfo.Name)
}

// checkWarehouseDeletable 检查仓库是否可以删除
func (s *sWarehouse) checkWarehouseDeletable(ctx context.Context, warehouseName string) error {
	// 这里可以添加更多业务规则检查，比如：
	// 1. 检查仓库是否有库存
	// 2. 检查仓库是否有未完成的单据
	// 3. 检查仓库是否是默认仓库等

	// 目前只做基本检查，后续可以根据业务需求扩展
	g.Log().Info(ctx, "检查仓库是否可以删除", warehouseName)

	return nil
}

// deleteWarehouseDocument 删除仓库文档
func (s *sWarehouse) deleteWarehouseDocument(ctx context.Context, warehouseName string) error {
	// 删除仓库文档
	if _, err := service.Document().Delete(ctx, &erp.ErpReq{
		DocType: erp.DocTypeWarehouse,
		Name:    warehouseName,
	}); err != nil {
		g.Log().Error(ctx, "删除仓库失败", err)
		return gerror.Wrapf(err, "删除仓库失败")
	}

	return nil
}
