// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto/erp"

	"github.com/gogf/gf/v2/encoding/gjson"
)

type (
	IDoctype interface {
		Meta(ctx context.Context, req *erp.ErpReq) (rst *gjson.Json, err error)
		Count(ctx context.Context, req *erp.ErpReq, params *erp.RequestParams) (int, error)
	}
	IDocument interface {
		List(ctx context.Context, req *erp.ErpReq, params *erp.RequestParams) (rst *gjson.Json, err error)
		Get(ctx context.Context, req *erp.ErpReq, params *erp.RequestParams) (rst *gjson.Json, err error)
		Create(ctx context.Context, docType string, data interface{}) (rst *gjson.Json, err error)
		Update(ctx context.Context, req *erp.ErpReq, data interface{}) (rst *gjson.Json, err error)
		Delete(ctx context.Context, req *erp.ErpReq) (rst *gjson.Json, err error)
		Copy(ctx context.Context, req *erp.ErpReq) (rst *gjson.Json, err error)
		Execute(ctx context.Context, req *erp.ErpReq, params interface{}) (rst *gjson.Json, err error)
		ChangeDocStatus(ctx context.Context, doctype string, name string, docstatus string) (rst *gjson.Json, err error)
	}
	IRpc interface {
		Execute(ctx context.Context, req *erp.ErpReq, params interface{}) (rst *gjson.Json, err error)
		// GetSiteCode 获取站点编码
		GetSiteCode(ctx context.Context) string
		// GetAndProcessSiteApiSecret 获取并处理 api_secret
		// 如果 api_secret 不是 base64 编码，则先使用 gaes.Encrypt 加密，再用 gbase64.Encode 编码后保存并返回
		GetAndProcessSiteAuthorization(ctx context.Context, siteCode string) (*erp.SiteAuthorization, error)
		// GetAndProcessCashierApiSecret 获取并处理收银员的 api_secret
		// 如果 api_secret 不是 base64 编码，则先使用 gaes.Encrypt 加密，再用 gbase64.Encode 编码后保存并返回
		GetAndProcessCashierAuthorization(ctx context.Context, cashierEmail string) (string, error)
	}
	IPrintFormat interface {
		// Meta 获取 Print Format 元数据
		// 参数：
		//   - ctx: 上下文对象
		//   - req: ERP 请求参数
		//
		// 返回：
		//   - *gjson.Json: Print Format 元数据 JSON
		//   - error: 错误信息
		Meta(ctx context.Context, req *erp.ErpReq) (*gjson.Json, error)
		// List 查询 Print Format 列表
		// 参数：
		//   - ctx: 上下文对象
		//   - req: Print Format 列表查询请求
		//
		// 返回：
		//   - []*erp.PrintFormatDetailResp: Print Format 列表
		//   - error: 错误信息
		List(ctx context.Context, req *erp.PrintFormatListReq) ([]*erp.PrintFormatDetailResp, error)
		// Get 查询 Print Format 详情
		// 参数：
		//   - ctx: 上下文对象
		//   - name: Print Format 名称
		//
		// 返回：
		//   - *erp.PrintFormatDetailResp: Print Format 详细信息
		//   - error: 错误信息
		Get(ctx context.Context, name string) (*erp.PrintFormatDetailResp, error)
		// Create 创建 Print Format
		// 参数：
		//   - ctx: 上下文对象
		//   - req: Print Format 创建请求
		//
		// 返回：
		//   - *erp.PrintFormatDetailResp: 创建后的 Print Format 信息
		//   - error: 错误信息
		Create(ctx context.Context, req *erp.PrintFormatCreateUpdateReq) (*erp.PrintFormatDetailResp, error)
		// Update 更新 Print Format
		// 参数：
		//   - ctx: 上下文对象
		//   - name: Print Format 名称
		//   - req: Print Format 更新请求
		//
		// 返回：
		//   - *erp.PrintFormatDetailResp: 更新后的 Print Format 信息
		//   - error: 错误信息
		Update(ctx context.Context, name string, req *erp.PrintFormatCreateUpdateReq) (*erp.PrintFormatDetailResp, error)
		// Delete 删除 Print Format
		// 参数：
		//   - ctx: 上下文对象
		//   - name: Print Format 名称
		//
		// 返回：
		//   - error: 错误信息
		Delete(ctx context.Context, name string) error
		// Count 统计 Print Format 数量
		// 参数：
		//   - ctx: 上下文对象
		//   - req: Print Format 列表查询请求（用于过滤条件）
		//
		// 返回：
		//   - int: Print Format 数量
		//   - error: 错误信息
		Count(ctx context.Context, req *erp.PrintFormatListReq) (int, error)
	}
	IReport interface {
		Run(ctx context.Context, params *erp.ReportParams) (rst *gjson.Json, err error)
	}
	IResource interface {
		List(ctx context.Context, docType string, params *erp.RequestParams) (rst *gjson.Json, err error)
		Get(ctx context.Context, docType string, name string, params *erp.RequestParams) (rst *gjson.Json, err error)
		Post(ctx context.Context, docType string, params *erp.RequestParams) (rst *gjson.Json, err error)
		Put(ctx context.Context, docType string, params *erp.RequestParams) (rst *gjson.Json, err error)
		Delete(ctx context.Context, docType string, params *erp.RequestParams) (rst *gjson.Json, err error)
	}
	IFile interface {
		// Upload 上传文件到 ERPNext（使用 multipart/form-data）
		// 参数：
		//   - ctx: 上下文对象
		//   - req: 文件上传请求参数
		//
		// 返回：
		//   - *erp.FileUploadResp: 上传成功后的文件信息
		//   - error: 错误信息
		Upload(ctx context.Context, req *erp.FileUploadReq) (*erp.FileUploadResp, error)
		// UploadBase64 上传 Base64 编码的文件到 ERPNext
		// 参数：
		//   - ctx: 上下文对象
		//   - req: Base64 编码的文件上传请求参数
		//
		// 返回：
		//   - *erp.FileUploadResp: 上传成功后的文件信息
		//   - error: 错误信息
		UploadBase64(ctx context.Context, req *erp.FileUploadBase64Req) (*erp.FileUploadResp, error)
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
		UploadFromBytes(ctx context.Context, fileName string, content []byte, docType, docName string, isPrivate int) (*erp.FileUploadResp, error)
		// UploadFromUrl 通过远程 URL 上传文件到 ERPNext
		// ERPNext 会自动下载远程文件并存储，适用于文件已在 CDN 或云存储的场景
		// 参数：
		//   - ctx: 上下文对象
		//   - req: URL 上传请求参数
		//
		// 返回：
		//   - *erp.FileUploadResp: 上传成功后的文件信息
		//   - error: 错误信息
		UploadFromUrl(ctx context.Context, req *erp.FileUploadUrlReq) (*erp.FileUploadResp, error)
		// List 查询文件列表
		// 参数：
		//   - ctx: 上下文对象
		//   - req: 文件列表查询请求
		//
		// 返回：
		//   - []*erp.FileUploadResp: 文件列表
		//   - error: 错误信息
		List(ctx context.Context, req *erp.FileListReq) ([]*erp.FileUploadResp, error)
		// Get 获取单个文件信息
		// 参数：
		//   - ctx: 上下文对象
		//   - name: 文件文档名称
		//
		// 返回：
		//   - *erp.FileUploadResp: 文件信息
		//   - error: 错误信息
		Get(ctx context.Context, name string) (*erp.FileUploadResp, error)
		// Delete 删除文件
		// 参数：
		//   - ctx: 上下文对象
		//   - name: 文件文档名称
		//
		// 返回：
		//   - error: 错误信息
		Delete(ctx context.Context, name string) error
	}
)

var (
	localDoctype     IDoctype
	localDocument    IDocument
	localRpc         IRpc
	localPrintFormat IPrintFormat
	localReport      IReport
	localResource    IResource
	localFile        IFile
)

func Doctype() IDoctype {
	if localDoctype == nil {
		panic("implement not found for interface IDoctype, forgot register?")
	}
	return localDoctype
}

func RegisterDoctype(i IDoctype) {
	localDoctype = i
}

func Document() IDocument {
	if localDocument == nil {
		panic("implement not found for interface IDocument, forgot register?")
	}
	return localDocument
}

func RegisterDocument(i IDocument) {
	localDocument = i
}

func Rpc() IRpc {
	if localRpc == nil {
		panic("implement not found for interface IRpc, forgot register?")
	}
	return localRpc
}

func RegisterRpc(i IRpc) {
	localRpc = i
}

func PrintFormat() IPrintFormat {
	if localPrintFormat == nil {
		panic("implement not found for interface IPrintFormat, forgot register?")
	}
	return localPrintFormat
}

func RegisterPrintFormat(i IPrintFormat) {
	localPrintFormat = i
}

func Report() IReport {
	if localReport == nil {
		panic("implement not found for interface IReport, forgot register?")
	}
	return localReport
}

func RegisterReport(i IReport) {
	localReport = i
}

func Resource() IResource {
	if localResource == nil {
		panic("implement not found for interface IResource, forgot register?")
	}
	return localResource
}

func RegisterResource(i IResource) {
	localResource = i
}

func File() IFile {
	if localFile == nil {
		panic("implement not found for interface IFile, forgot register?")
	}
	return localFile
}

func RegisterFile(i IFile) {
	localFile = i
}
