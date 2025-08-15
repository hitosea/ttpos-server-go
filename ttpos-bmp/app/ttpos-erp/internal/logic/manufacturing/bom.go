package manufacturing

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/api/manufacturing"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto/erp"
	"ttpos-bmp/app/ttpos-erp/internal/service"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
)

var (
	Bom = new(sBom)
)

type sBom struct {
}

func init() {
	service.RegisterBom(Bom)
}

// GetBomList 获取BOM列表
// 参数：ctx 上下文，req 获取BOM列表请求
// 返回：BOM列表响应，错误信息
func (s *sBom) GetBomList(ctx context.Context, req *manufacturing.GetBomListReq) (res *manufacturing.GetBomListResp, err error) {
	// 构建查询过滤条件
	filters, err := s.buildBomListFilters(ctx, req)
	if err != nil {
		return nil, gerror.Wrapf(err, "构建BOM查询过滤条件失败")
	}

	// 调用ERP接口获取BOM列表
	resp, err := service.Document().List(ctx, &erp.ErpReq{
		DocType: "BOM",
	}, &erp.RequestParams{
		Fields:  []string{"name", "item", "uom", "quantity", "is_active", "is_default"},
		Filters: filters,
	})
	if err != nil {
		return nil, gerror.Wrapf(err, "获取BOM列表失败")
	}

	// 解析响应数据并构建BOM列表
	bomList, err := s.parseBomListResponse(ctx, resp.Bytes())
	if err != nil {
		return nil, gerror.Wrapf(err, "解析BOM列表响应失败")
	}

	return &manufacturing.GetBomListResp{
		BomList: bomList,
	}, nil
}

// buildBomListFilters 构建BOM列表查询过滤条件
// 参数：ctx 上下文，req 获取BOM列表请求
// 返回：过滤条件数组，错误信息
func (s *sBom) buildBomListFilters(ctx context.Context, req *manufacturing.GetBomListReq) ([][]string, error) {
	filters := make([][]string, 0)

	// 根据公司简称构建公司过滤条件
	if len(req.CompanyAbbr) > 0 {
		company, err := service.Company().GetCompanyWithAbbr(ctx, req.CompanyAbbr)
		if err != nil {
			return nil, gerror.Wrapf(err, "根据公司缩写查询公司失败")
		}
		filters = append(filters, []string{"company", "=", company.CompanyName})
	}

	// 根据分支机构构建过滤条件
	if len(req.Branch) > 0 {
		filters = append(filters, []string{"branch", "=", req.Branch})
	}

	// 根据商品编码构建过滤条件
	if len(req.ItemCode) > 0 {
		filters = append(filters, []string{"item_code", "=", req.ItemCode})
	}

	// 根据是否激活状态构建过滤条件
	if req.GetIsActive() {
		filters = append(filters, []string{"is_active", "=", "1"})
	}

	// 根据是否默认状态构建过滤条件
	if req.GetIsDefault() {
		filters = append(filters, []string{"is_default", "=", "1"})
	}

	return filters, nil
}

// parseBomListResponse 解析BOM列表响应数据
// 参数：ctx 上下文，responseData 响应数据字节数组
// 返回：BOM信息列表，错误信息
func (s *sBom) parseBomListResponse(ctx context.Context, responseData []byte) ([]*manufacturing.BomInfo, error) {
	// 解析JSON响应
	j, err := gjson.DecodeToJson(responseData)
	if err != nil {
		return nil, gerror.Wrapf(err, "解析JSON响应失败")
	}

	// 提取数据数组
	dataArray := j.GetJsons("data")
	bomList := make([]*manufacturing.BomInfo, 0, len(dataArray))

	// 遍历数据数组，构建BOM信息列表
	for _, item := range dataArray {
		bomInfo := &manufacturing.BomInfo{
			ItemCode:  item.Get("item").String(),
			BomName:   item.Get("name").String(),
			Uom:       item.Get("uom").String(),
			Quantity:  item.Get("quantity").Float32(),
			IsActive:  item.Get("is_active").Bool(),
			IsDefault: item.Get("is_default").Bool(),
		}
		bomList = append(bomList, bomInfo)
	}

	return bomList, nil
}

// SaveBom 保存BOM信息
// 参数：ctx 上下文，req 保存BOM请求
// 返回：保存BOM响应，错误信息
func (s *sBom) SaveBom(ctx context.Context, req *manufacturing.SaveBomReq) (res *manufacturing.SaveBomResp, err error) {
	// 根据公司简称获取公司名称
	companyName, err := service.Company().GetCompanyNameWithAbbr(ctx, req.CompanyAbbr)
	if err != nil {
		return nil, gerror.Wrapf(err, "查询公司名称失败")
	}

	// 构建BOM数据结构
	bomInfo := s.buildBomData(req, companyName)

	// 调用ERP接口创建BOM
	resp, err := service.Document().Create(ctx, "BOM", bomInfo)
	if err != nil {
		return nil, gerror.Wrapf(err, "创建BOM失败")
	}

	// 解析创建BOM的响应结果
	bomName, err := s.parseSaveBomResponse(ctx, resp.Bytes())
	if err != nil {
		return nil, gerror.Wrapf(err, "解析保存BOM响应失败")
	}

	//创建后是草稿状态，提交BOM
	_, err = service.Document().ChangeDocStatus(ctx, "BOM", bomName, 1)
	if err != nil {
		return nil, gerror.Wrapf(err, "提交BOM失败")
	}

	return &manufacturing.SaveBomResp{
		BomName: bomName,
	}, nil
}

// buildBomData 构建BOM数据结构
// 参数：req 保存BOM请求，companyName 公司名称
// 返回：BOM数据结构
func (s *sBom) buildBomData(req *manufacturing.SaveBomReq, companyName string) *erp.Bom {
	bomInfo := &erp.Bom{
		Item:      req.ItemCode,
		Company:   companyName,
		Uom:       req.Uom,
		Quantity:  req.Quantity,
		IsActive:  req.IsActive,
		IsDefault: req.IsDefault,
		Items:     make([]erp.BomItem, 0, len(req.Items)), // 预分配容量以提高性能
	}

	// 构建BOM明细项目列表
	for _, item := range req.Items {
		bomInfo.Items = append(bomInfo.Items, erp.BomItem{
			ItemCode: item.ItemCode,
			Rate:     item.Rate,
			Qty:      item.Qty,
			Uom:      item.Uom,
		})
	}

	return bomInfo
}

// parseSaveBomResponse 解析保存BOM响应数据
// 参数：ctx 上下文，responseData 响应数据字节数组
// 返回：BOM名称，错误信息
func (s *sBom) parseSaveBomResponse(ctx context.Context, responseData []byte) (string, error) {
	// 解析JSON响应
	j, err := gjson.DecodeToJson(responseData)
	if err != nil {
		return "", gerror.Wrapf(err, "解析JSON响应失败")
	}

	// 提取BOM名称
	bomName := j.Get("data.name").String()
	if bomName == "" {
		return "", gerror.New("响应数据中未找到BOM名称")
	}

	return bomName, nil
}
