package stock

import (
	"context"
	"strings"
	"ttpos-bmp/app/ttpos-erp/api/warehouse"
	"ttpos-bmp/app/ttpos-erp/internal/consts"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto/erp"
	"ttpos-bmp/app/ttpos-erp/internal/service"

	"github.com/gogf/gf/v2/encoding/gjson"
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
func (s *sWarehouse) CreateWarehouse(ctx context.Context, req *erp.CreateWarehouseInp) (warehouseName string, err error) {
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
func (s *sWarehouse) validateCreateWarehouseReq(req *erp.CreateWarehouseInp) error {
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
func (s *sWarehouse) generateWarehouseName(req *erp.CreateWarehouseInp) string {
	return strings.Join([]string{req.Branch, req.WhType, req.AliasName}, "-")
}

// createWarehouseDocument 创建仓库文档
func (s *sWarehouse) createWarehouseDocument(ctx context.Context, warehouseName string, req *erp.CreateWarehouseInp, companyName string) error {
	// 构建仓库创建参数
	warehousePayload := g.Map{
		"warehouse_name":   warehouseName, // 仓库名称
		"warehouse_type":   req.WhType,    // 仓库类型
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
	warehouseList, err := s.queryWarehouseList(ctx, filters)
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
		WarehouseType: "Default",
	})
	if err != nil {
		return nil, err
	}
	if len(warehouseList.WarehouseList) == 0 {
		return nil, gerror.New("默认仓库不存在")
	}
	return warehouseList.WarehouseList[0], nil
}

// buildWarehouseListFilters 构建仓库列表查询过滤器
func (s *sWarehouse) buildWarehouseListFilters(ctx context.Context, req *warehouse.GetWarehouseListReq) [][]string {
	filters := make([][]string, 0)

	// 按分支机构过滤
	if len(req.Branch) > 0 {
		filters = append(filters, g.ArrayStr{"custom_branch", "like", "%" + req.Branch + "%"})
	}

	// 按仓库名称过滤
	if len(req.WarehouseName) > 0 {
		filters = append(filters, g.ArrayStr{"warehouse_name", "like", "%" + req.WarehouseName + "%"})
	}

	// 按公司过滤
	if len(req.Company) > 0 {
		filters = append(filters, g.ArrayStr{"company", "like", "%" + req.Company + "%"})
	}

	// 按仓库类型过滤
	if len(req.WarehouseType) > 0 {
		filters = append(filters, g.ArrayStr{"warehouse_type", "=", req.WarehouseType})
	}

	// 按仓库别名过滤
	if len(req.AliasName) > 0 {
		filters = append(filters, g.ArrayStr{"custom_aliasname", "like", "%" + req.AliasName + "%"})
	}

	if len(req.CompanyAbbr) > 0 {
		companyName, _ := service.Company().GetCompanyNameWithAbbr(ctx, req.CompanyAbbr)
		filters = append(filters, []string{"company", "=", companyName})
	}

	return filters
}

// queryWarehouseList 执行仓库列表查询
func (s *sWarehouse) queryWarehouseList(ctx context.Context, filters [][]string) ([]*warehouse.WarehouseInfo, error) {
	// 查询仓库列表
	resp, err := service.Document().List(ctx, &dto.ErpReq{
		DocType: "Warehouse",
	}, &dto.RequestParams{
		Fields:  g.ArrayStr{"warehouse_name", "warehouse_type", "custom_branch", "custom_aliasname", "company"},
		Filters: filters,
		Limit:   consts.Limit999,
	})

	if err != nil {
		return nil, err
	}

	// 解析响应数据
	j, err := gjson.DecodeToJson(resp.Bytes())
	if err != nil {
		return nil, gerror.Wrapf(err, "解析仓库列表响应失败")
	}

	// 转换为仓库信息列表
	warehouseList := make([]*warehouse.WarehouseInfo, 0)
	dataArray := j.GetJsons("data")

	for _, data := range dataArray {
		warehouseInfo := &warehouse.WarehouseInfo{
			Branch:        data.Get("custom_branch").String(),
			Company:       data.Get("company").String(),
			WarehouseName: data.Get("warehouse_name").String(),
			WarehouseType: data.Get("warehouse_type").String(),
			AliasName:     data.Get("custom_aliasname").String(),
		}
		warehouseList = append(warehouseList, warehouseInfo)
	}

	return warehouseList, nil
}
