package erpnext

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto/erp"
	"ttpos-bmp/app/ttpos-erp/internal/service"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
)

const printFormatDocType = "Print Format"

var PrintFormat = new(sPrintFormat)

type sPrintFormat struct {
}

// Meta 获取 Print Format 元数据
// 参数：
//   - ctx: 上下文对象
//   - req: ERP 请求参数
//
// 返回：
//   - *gjson.Json: Print Format 元数据 JSON
//   - error: 错误信息
func (s *sPrintFormat) Meta(ctx context.Context, req *erp.ErpReq) (*gjson.Json, error) {
	resp, err := service.Doctype().Meta(ctx, &erp.ErpReq{
		DocType: printFormatDocType,
	})
	if err != nil {
		return nil, gerror.Wrapf(err, "获取 Print Format 元数据失败")
	}
	return resp, nil
}

// List 查询 Print Format 列表
// 参数：
//   - ctx: 上下文对象
//   - req: Print Format 列表查询请求
//
// 返回：
//   - []*erp.PrintFormatDetailResp: Print Format 列表
//   - error: 错误信息
func (s *sPrintFormat) List(ctx context.Context, req *erp.PrintFormatListReq) ([]*erp.PrintFormatDetailResp, error) {
	filters := make([][]string, 0)

	// 按 DocType 过滤
	if req.DocType != "" {
		filters = append(filters, []string{"doc_type", "=", req.DocType})
	}

	// 查询列表
	resp, err := service.Document().List(ctx, &erp.ErpReq{
		DocType: printFormatDocType,
	}, &erp.RequestParams{
		Fields:     []string{"name", "doc_type", "standard", "print_format_type", "module"},
		Filters:    filters,
		Limit:      req.Limit,
		LimitStart: req.LimitStart,
	})
	if err != nil {
		return nil, gerror.Wrapf(err, "查询 Print Format 列表失败")
	}

	// 解析响应
	list := make([]*erp.PrintFormatDetailResp, 0)
	dataArray := resp.GetJsons("data")
	for _, data := range dataArray {
		item := &erp.PrintFormatDetailResp{}
		if err := data.Scan(item); err != nil {
			continue
		}
		list = append(list, item)
	}

	return list, nil
}

// Get 查询 Print Format 详情
// 参数：
//   - ctx: 上下文对象
//   - name: Print Format 名称
//
// 返回：
//   - *erp.PrintFormatDetailResp: Print Format 详细信息
//   - error: 错误信息
func (s *sPrintFormat) Get(ctx context.Context, name string) (*erp.PrintFormatDetailResp, error) {
	if name == "" {
		return nil, gerror.New("Print Format 名称不能为空")
	}

	resp, err := service.Document().Get(ctx, &erp.ErpReq{
		DocType: printFormatDocType,
		Name:    name,
	}, nil)
	if err != nil {
		return nil, gerror.Wrapf(err, "查询 Print Format 详情失败")
	}

	// 解析响应
	detail := &erp.PrintFormatDetailResp{}
	if err := resp.GetJson("data").Scan(detail); err != nil {
		return nil, gerror.Wrapf(err, "解析 Print Format 详情失败")
	}

	return detail, nil
}

// Create 创建 Print Format
// 参数：
//   - ctx: 上下文对象
//   - req: Print Format 创建请求
//
// 返回：
//   - *erp.PrintFormatDetailResp: 创建后的 Print Format 信息
//   - error: 错误信息
func (s *sPrintFormat) Create(ctx context.Context, req *erp.PrintFormatCreateUpdateReq) (*erp.PrintFormatDetailResp, error) {
	if req.Name == "" {
		return nil, gerror.New("Print Format 名称不能为空")
	}
	if req.DocType == "" {
		return nil, gerror.New("DocType 不能为空")
	}

	resp, err := service.Document().Create(ctx, printFormatDocType, req)
	if err != nil {
		return nil, gerror.Wrapf(err, "创建 Print Format 失败")
	}

	// 解析响应
	detail := &erp.PrintFormatDetailResp{}
	if err := resp.GetJson("data").Scan(detail); err != nil {
		return nil, gerror.Wrapf(err, "解析 Print Format 响应失败")
	}

	return detail, nil
}

// Update 更新 Print Format
// 参数：
//   - ctx: 上下文对象
//   - name: Print Format 名称
//   - req: Print Format 更新请求
//
// 返回：
//   - *erp.PrintFormatDetailResp: 更新后的 Print Format 信息
//   - error: 错误信息
func (s *sPrintFormat) Update(ctx context.Context, name string, req *erp.PrintFormatCreateUpdateReq) (*erp.PrintFormatDetailResp, error) {
	if name == "" {
		return nil, gerror.New("Print Format 名称不能为空")
	}

	_, err := service.Document().Update(ctx, &erp.ErpReq{
		DocType: printFormatDocType,
		Name:    name,
	}, req)
	if err != nil {
		return nil, gerror.Wrapf(err, "更新 Print Format 失败")
	}

	// 获取更新后的信息
	return s.Get(ctx, name)
}

// Delete 删除 Print Format
// 参数：
//   - ctx: 上下文对象
//   - name: Print Format 名称
//
// 返回：
//   - error: 错误信息
func (s *sPrintFormat) Delete(ctx context.Context, name string) error {
	if name == "" {
		return gerror.New("Print Format 名称不能为空")
	}

	_, err := service.Document().Delete(ctx, &erp.ErpReq{
		DocType: printFormatDocType,
		Name:    name,
	})
	if err != nil {
		return gerror.Wrapf(err, "删除 Print Format 失败")
	}

	return nil
}

// Count 统计 Print Format 数量
// 参数：
//   - ctx: 上下文对象
//   - req: Print Format 列表查询请求（用于过滤条件）
//
// 返回：
//   - int: Print Format 数量
//   - error: 错误信息
func (s *sPrintFormat) Count(ctx context.Context, req *erp.PrintFormatListReq) (int, error) {
	filters := make([][]string, 0)

	// 按 DocType 过滤
	if req.DocType != "" {
		filters = append(filters, []string{"doc_type", "=", req.DocType})
	}

	count, err := service.Doctype().Count(ctx, &erp.ErpReq{
		DocType: printFormatDocType,
	}, &erp.RequestParams{
		Filters: filters,
	})
	if err != nil {
		return 0, gerror.Wrapf(err, "统计 Print Format 数量失败")
	}

	return count, nil
}
