package erpnext

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"mime/multipart"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto/erp"
	"ttpos-bmp/app/ttpos-erp/internal/service"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
)

const (
	// 文件上传 API 端点
	fileUploadApiUrl = "/api/method/upload_file"
	// 文件资源 API 端点（用于查询和删除）
	fileResourceApiUrl = "/api/resource/File"
)

var File = new(sFile)

type sFile struct{}

func init() {
	service.RegisterFile(File)
}

// Upload 上传文件到 ERPNext（使用 multipart/form-data）
// 参数：
//   - ctx: 上下文对象
//   - req: 文件上传请求参数
//
// 返回：
//   - *erp.FileUploadResp: 上传成功后的文件信息
//   - error: 错误信息
func (s *sFile) Upload(ctx context.Context, req *erp.FileUploadReq) (*erp.FileUploadResp, error) {
	if req == nil || len(req.FileContent) == 0 {
		return nil, gerror.New("文件内容不能为空")
	}
	if req.FileName == "" {
		return nil, gerror.New("文件名不能为空")
	}

	// 构建 multipart/form-data 请求
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// 添加文件字段
	part, err := writer.CreateFormFile("file", req.FileName)
	if err != nil {
		return nil, gerror.Wrapf(err, "创建 multipart 文件字段失败")
	}
	if _, err := io.Copy(part, bytes.NewReader(req.FileContent)); err != nil {
		return nil, gerror.Wrapf(err, "写入文件内容失败")
	}

	// 添加可选参数
	if req.DocType != "" {
		_ = writer.WriteField("doctype", req.DocType)
	}
	if req.DocName != "" {
		_ = writer.WriteField("docname", req.DocName)
	}
	if req.IsPrivate != 0 {
		_ = writer.WriteField("is_private", gconv.String(req.IsPrivate))
	}
	if req.Folder != "" {
		_ = writer.WriteField("folder", req.Folder)
	}

	// 关闭 writer（这会添加结束边界）
	if err := writer.Close(); err != nil {
		return nil, gerror.Wrapf(err, "关闭 multipart writer 失败")
	}

	// 发送请求
	client := GetClient(ctx)
	client.SetHeader("Content-Type", writer.FormDataContentType())
	resp, err := client.Post(ctx, fileUploadApiUrl, body.Bytes())
	if err != nil {
		return nil, gerror.Wrapf(err, "上传文件请求失败")
	}
	defer resp.Close()

	// 解析响应
	respBody := resp.ReadAllString()
	rst, err := detectError(g.NewVar(respBody))
	if err != nil {
		return nil, err
	}

	return s.parseUploadResp(rst)
}

// UploadBase64 上传 Base64 编码的文件到 ERPNext
// 参数：
//   - ctx: 上下文对象
//   - req: Base64 编码的文件上传请求参数
//
// 返回：
//   - *erp.FileUploadResp: 上传成功后的文件信息
//   - error: 错误信息
func (s *sFile) UploadBase64(ctx context.Context, req *erp.FileUploadBase64Req) (*erp.FileUploadResp, error) {
	if req == nil || req.FileData == "" {
		return nil, gerror.New("文件内容不能为空")
	}
	if req.FileName == "" {
		return nil, gerror.New("文件名不能为空")
	}

	// 设置默认值
	req.FromForm = 1

	// 发送请求
	resp := GetClient(ctx).PostVar(ctx, fileUploadApiUrl, g.Map{
		"cmd":        "uploadfile",
		"filename":   req.FileName,
		"filedata":   req.FileData,
		"doctype":    req.DocType,
		"docname":    req.DocName,
		"is_private": req.IsPrivate,
		"from_form":  req.FromForm,
	})

	rst, err := detectError(resp)
	if err != nil {
		return nil, err
	}

	return s.parseUploadResp(rst)
}

// UploadFromBytes 从字节数组上传文件（自动转换为 Base64）
// 参数：
//   - ctx: 上下文对象
//   - fileName: 文件名
//   - content: 文件内容字节数组
//   - docType: 关联的文档类型（可选）
//   - docName: 关联的文档名称（可选）
//   - isPrivate: 是否私有（可选）
//
// 返回：
//   - *erp.FileUploadResp: 上传成功后的文件信息
//   - error: 错误信息
func (s *sFile) UploadFromBytes(ctx context.Context, fileName string, content []byte, docType, docName string, isPrivate int) (*erp.FileUploadResp, error) {
	// 转换为 Base64
	base64Content := base64.StdEncoding.EncodeToString(content)

	return s.UploadBase64(ctx, &erp.FileUploadBase64Req{
		FileName:  fileName,
		FileData:  base64Content,
		DocType:   docType,
		DocName:   docName,
		IsPrivate: isPrivate,
	})
}

// UploadFromUrl 通过远程 URL 上传文件到 ERPNext
// ERPNext 会自动下载远程文件并存储，适用于文件已在 CDN 或云存储的场景
// 参数：
//   - ctx: 上下文对象
//   - req: URL 上传请求参数
//
// 返回：
//   - *erp.FileUploadResp: 上传成功后的文件信息
//   - error: 错误信息
func (s *sFile) UploadFromUrl(ctx context.Context, req *erp.FileUploadUrlReq) (*erp.FileUploadResp, error) {
	if req == nil || req.FileUrl == "" {
		return nil, gerror.New("文件 URL 不能为空")
	}

	// 构建请求参数
	params := g.Map{
		"file_url": req.FileUrl,
	}
	if req.FileName != "" {
		params["file_name"] = req.FileName
	}
	if req.DocType != "" {
		params["doctype"] = req.DocType
	}
	if req.DocName != "" {
		params["docname"] = req.DocName
	}
	if req.IsPrivate != 0 {
		params["is_private"] = req.IsPrivate
	}
	if req.Folder != "" {
		params["folder"] = req.Folder
	}

	// 发送请求
	resp := GetClient(ctx).PostVar(ctx, fileUploadApiUrl, params)
	rst, err := detectError(resp)
	if err != nil {
		return nil, err
	}

	return s.parseUploadResp(rst)
}

// List 查询文件列表
// 参数：
//   - ctx: 上下文对象
//   - req: 文件列表查询请求
//
// 返回：
//   - []*erp.FileUploadResp: 文件列表
//   - error: 错误信息
func (s *sFile) List(ctx context.Context, req *erp.FileListReq) ([]*erp.FileUploadResp, error) {
	// 构建过滤条件
	filters := make([][]string, 0)
	if req.DocType != "" {
		filters = append(filters, []string{"attached_to_doctype", "=", req.DocType})
	}
	if req.DocName != "" {
		filters = append(filters, []string{"attached_to_name", "=", req.DocName})
	}
	if req.IsPrivate != nil {
		filters = append(filters, []string{"is_private", "=", gconv.String(*req.IsPrivate)})
	}
	if req.Folder != "" {
		filters = append(filters, []string{"folder", "=", req.Folder})
	}

	// 设置默认分页
	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}

	// 查询文件列表
	params := &erp.RequestParams{
		Fields: []string{
			"name",
			"file_name",
			"file_url",
			"is_private",
			"attached_to_doctype",
			"attached_to_name",
			"file_size",
			"content_type",
		},
		Filters:    filters,
		LimitStart: req.LimitStart,
		Limit:      limit,
	}

	resp := GetClient(ctx).GetVar(ctx, fileResourceApiUrl, params)
	rst, err := detectError(resp)
	if err != nil {
		return nil, err
	}

	// 解析文件列表
	dataArray := rst.GetJsons("data")
	files := make([]*erp.FileUploadResp, 0, len(dataArray))
	for _, item := range dataArray {
		files = append(files, &erp.FileUploadResp{
			Name:              item.Get("name").String(),
			FileName:          item.Get("file_name").String(),
			FileUrl:           item.Get("file_url").String(),
			IsPrivate:         item.Get("is_private").Int(),
			AttachedToDocType: item.Get("attached_to_doctype").String(),
			AttachedToName:    item.Get("attached_to_name").String(),
			FileSize:          item.Get("file_size").Int64(),
			ContentType:       item.Get("content_type").String(),
		})
	}

	return files, nil
}

// Get 获取单个文件信息
// 参数：
//   - ctx: 上下文对象
//   - name: 文件文档名称
//
// 返回：
//   - *erp.FileUploadResp: 文件信息
//   - error: 错误信息
func (s *sFile) Get(ctx context.Context, name string) (*erp.FileUploadResp, error) {
	if name == "" {
		return nil, gerror.New("文件名称不能为空")
	}

	resp := GetClient(ctx).GetVar(ctx, fmt.Sprintf("%s/%s", fileResourceApiUrl, name))
	rst, err := detectError(resp)
	if err != nil {
		return nil, err
	}

	dataJson := rst.GetJson("data")
	return &erp.FileUploadResp{
		Name:              dataJson.Get("name").String(),
		FileName:          dataJson.Get("file_name").String(),
		FileUrl:           dataJson.Get("file_url").String(),
		IsPrivate:         dataJson.Get("is_private").Int(),
		AttachedToDocType: dataJson.Get("attached_to_doctype").String(),
		AttachedToName:    dataJson.Get("attached_to_name").String(),
		FileSize:          dataJson.Get("file_size").Int64(),
		ContentType:       dataJson.Get("content_type").String(),
	}, nil
}

// Delete 删除文件
// 参数：
//   - ctx: 上下文对象
//   - name: 文件文档名称
//
// 返回：
//   - error: 错误信息
func (s *sFile) Delete(ctx context.Context, name string) error {
	if name == "" {
		return gerror.New("文件名称不能为空")
	}

	resp := GetClient(ctx).DeleteVar(ctx, fmt.Sprintf("%s/%s", fileResourceApiUrl, name))
	_, err := detectError(resp)
	return err
}

// parseUploadResp 解析上传响应
func (s *sFile) parseUploadResp(rst *gjson.Json) (*erp.FileUploadResp, error) {
	// ERPNext 响应格式可能是 {"message": {...}} 或 {"data": {...}}
	var data *gjson.Json
	if rst.Contains("message") {
		data = rst.GetJson("message")
	} else if rst.Contains("data") {
		data = rst.GetJson("data")
	} else {
		data = rst
	}

	return &erp.FileUploadResp{
		Name:              data.Get("name").String(),
		FileName:          data.Get("file_name").String(),
		FileUrl:           data.Get("file_url").String(),
		IsPrivate:         data.Get("is_private").Int(),
		AttachedToDocType: data.Get("attached_to_doctype").String(),
		AttachedToName:    data.Get("attached_to_name").String(),
		FileSize:          data.Get("file_size").Int64(),
		ContentType:       data.Get("content_type").String(),
	}, nil
}
