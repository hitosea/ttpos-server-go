package erpnext

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto/erp"
	"ttpos-bmp/app/ttpos-erp/internal/service"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
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

// SyncToLocalResult 同步到本地的结果
type SyncToLocalResult struct {
	Total   int      // 总数量
	Success int      // 成功数量
	Failed  int      // 失败数量
	Errors  []string // 失败详情
}

// SyncToLocal 从 ERP 同步指定前缀的打印格式到本地 JSON 文件
// 参数：
//   - ctx: 上下文对象
//   - dirBase: 输出目录
//   - prefix: 打印格式名称前缀，如 "Wallace"
//
// 返回：
//   - *SyncToLocalResult: 同步结果统计
//   - error: 错误信息
func (s *sPrintFormat) SyncToLocal(ctx context.Context, dirBase string, prefix string) (*SyncToLocalResult, error) {
	result := &SyncToLocalResult{
		Errors: make([]string, 0),
	}

	// 1. 检查并创建目标目录
	if err := os.MkdirAll(dirBase, 0755); err != nil {
		return nil, gerror.Wrapf(err, "创建目录失败: %s", dirBase)
	}

	// 2. 获取所有 Print Format 列表
	g.Log().Infof(ctx, "开始获取 Print Format 列表，前缀: %s", prefix)

	// 查询列表，使用 like 过滤
	resp, err := service.Document().List(ctx, &erp.ErpReq{
		DocType: printFormatDocType,
	}, &erp.RequestParams{
		Fields:  []string{"name"},
		Filters: [][]string{{"name", "like", prefix + "%"}},
		Limit:   0, // 获取所有
	})
	if err != nil {
		return nil, gerror.Wrapf(err, "获取 Print Format 列表失败")
	}

	// 解析列表
	dataArray := resp.GetJsons("data")
	if len(dataArray) == 0 {
		g.Log().Infof(ctx, "未找到 %s 开头的 Print Format", prefix)
		return result, nil
	}

	result.Total = len(dataArray)
	g.Log().Infof(ctx, "找到 %d 个 %s 开头的 Print Format", result.Total, prefix)

	// 3. 遍历获取每个文档详情并保存
	for _, data := range dataArray {
		name := data.Get("name").String()
		if name == "" {
			continue
		}

		g.Log().Infof(ctx, "正在同步: %s", name)

		// 获取文档详情（完整 JSON）
		detailResp, err := service.Document().Get(ctx, &erp.ErpReq{
			DocType: printFormatDocType,
			Name:    name,
		}, nil)
		if err != nil {
			errMsg := fmt.Sprintf("%s: 获取详情失败 - %v", name, err)
			result.Errors = append(result.Errors, errMsg)
			result.Failed++
			g.Log().Errorf(ctx, errMsg)
			continue
		}

		// 获取 data 字段内容
		docData := detailResp.Get("data")
		if docData == nil {
			errMsg := fmt.Sprintf("%s: 响应数据为空", name)
			result.Errors = append(result.Errors, errMsg)
			result.Failed++
			g.Log().Errorf(ctx, errMsg)
			continue
		}

		// 格式化 JSON
		jsonBytes, err := json.MarshalIndent(docData, "", "  ")
		if err != nil {
			errMsg := fmt.Sprintf("%s: JSON 序列化失败 - %v", name, err)
			result.Errors = append(result.Errors, errMsg)
			result.Failed++
			g.Log().Errorf(ctx, errMsg)
			continue
		}

		// 构建文件名（清理特殊字符）
		safeFileName := strings.ReplaceAll(name, "/", "_")
		safeFileName = strings.ReplaceAll(safeFileName, "\\", "_")
		filePath := filepath.Join(dirBase, safeFileName+".json")

		// 写入文件
		if err := os.WriteFile(filePath, jsonBytes, 0644); err != nil {
			errMsg := fmt.Sprintf("%s: 写入文件失败 - %v", name, err)
			result.Errors = append(result.Errors, errMsg)
			result.Failed++
			g.Log().Errorf(ctx, errMsg)
			continue
		}

		result.Success++
		g.Log().Infof(ctx, "  - %s: 保存成功 -> %s", name, filePath)
	}

	return result, nil
}
