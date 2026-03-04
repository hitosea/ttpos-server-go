package file

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/api"
	"ttpos-bmp/app/ttpos-erp/api/file"
	"ttpos-bmp/app/ttpos-erp/internal/controller/rpc"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto/erp"
	"ttpos-bmp/app/ttpos-erp/internal/service"

	"github.com/gogf/gf/contrib/rpc/grpcx/v2"
)

// Controller 文件服务控制器
type Controller struct {
	file.UnimplementedFileServiceServer
}

// Register 注册文件服务到 gRPC 服务器
func Register(s *grpcx.GrpcServer) {
	file.RegisterFileServiceServer(s.Server, &Controller{})
}

// UploadFile 上传文件到 ERPNext（二进制方式）
func (*Controller) UploadFile(ctx context.Context, req *file.UploadFileReq) (res *api.ResponseInfo, err error) {
	// 参数校验
	if len(req.FileName) == 0 {
		return rpc.ApiError("文件名不能为空"), nil
	}
	if len(req.FileContent) == 0 {
		return rpc.ApiError("文件内容不能为空"), nil
	}

	// 调用服务层上传文件
	resp, err := service.File().Upload(ctx, &erp.FileUploadReq{
		FileName:    req.FileName,
		FileContent: req.FileContent,
		DocType:     req.DocType,
		DocName:     req.DocName,
		IsPrivate:   int(req.IsPrivate),
		Folder:      req.Folder,
	})
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	// 转换响应
	return rpc.ApiSuccessWithData("上传成功", &file.UploadFileResp{
		Name:              resp.Name,
		FileName:          resp.FileName,
		FileUrl:           resp.FileUrl,
		IsPrivate:         int32(resp.IsPrivate),
		AttachedToDoctype: resp.AttachedToDocType,
		AttachedToName:    resp.AttachedToName,
		FileSize:          resp.FileSize,
		ContentType:       resp.ContentType,
	}), nil
}

// UploadFileBase64 上传文件到 ERPNext（Base64 编码方式）
func (*Controller) UploadFileBase64(ctx context.Context, req *file.UploadFileBase64Req) (res *api.ResponseInfo, err error) {
	// 参数校验
	if len(req.FileName) == 0 {
		return rpc.ApiError("文件名不能为空"), nil
	}
	if len(req.FileData) == 0 {
		return rpc.ApiError("文件内容不能为空"), nil
	}

	// 调用服务层上传文件
	resp, err := service.File().UploadBase64(ctx, &erp.FileUploadBase64Req{
		FileName:  req.FileName,
		FileData:  req.FileData,
		DocType:   req.DocType,
		DocName:   req.DocName,
		IsPrivate: int(req.IsPrivate),
	})
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	// 转换响应
	return rpc.ApiSuccessWithData("上传成功", &file.UploadFileResp{
		Name:              resp.Name,
		FileName:          resp.FileName,
		FileUrl:           resp.FileUrl,
		IsPrivate:         int32(resp.IsPrivate),
		AttachedToDoctype: resp.AttachedToDocType,
		AttachedToName:    resp.AttachedToName,
		FileSize:          resp.FileSize,
		ContentType:       resp.ContentType,
	}), nil
}

// UploadFileUrl 上传文件到 ERPNext（URL 方式，ERPNext 自动下载）
func (*Controller) UploadFileUrl(ctx context.Context, req *file.UploadFileUrlReq) (res *api.ResponseInfo, err error) {
	// 参数校验
	if len(req.FileUrl) == 0 {
		return rpc.ApiError("文件 URL 不能为空"), nil
	}

	// 调用服务层上传文件
	resp, err := service.File().UploadFromUrl(ctx, &erp.FileUploadUrlReq{
		FileUrl:   req.FileUrl,
		FileName:  req.FileName,
		DocType:   req.DocType,
		DocName:   req.DocName,
		IsPrivate: int(req.IsPrivate),
		Folder:    req.Folder,
	})
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	// 转换响应
	return rpc.ApiSuccessWithData("上传成功", &file.UploadFileResp{
		Name:              resp.Name,
		FileName:          resp.FileName,
		FileUrl:           resp.FileUrl,
		IsPrivate:         int32(resp.IsPrivate),
		AttachedToDoctype: resp.AttachedToDocType,
		AttachedToName:    resp.AttachedToName,
		FileSize:          resp.FileSize,
		ContentType:       resp.ContentType,
	}), nil
}

// ListFiles 查询文件列表
func (*Controller) ListFiles(ctx context.Context, req *file.ListFilesReq) (res *api.ResponseInfo, err error) {
	// 构建查询请求
	listReq := &erp.FileListReq{
		DocType:    req.DocType,
		DocName:    req.DocName,
		Folder:     req.Folder,
		LimitStart: int(req.LimitStart),
		Limit:      int(req.Limit),
	}

	// 处理 IsPrivate 参数（-1 表示不限）
	if req.IsPrivate >= 0 {
		isPrivate := int(req.IsPrivate)
		listReq.IsPrivate = &isPrivate
	}

	// 调用服务层查询
	files, err := service.File().List(ctx, listReq)
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	// 转换响应
	respFiles := make([]*file.UploadFileResp, 0, len(files))
	for _, f := range files {
		respFiles = append(respFiles, &file.UploadFileResp{
			Name:              f.Name,
			FileName:          f.FileName,
			FileUrl:           f.FileUrl,
			IsPrivate:         int32(f.IsPrivate),
			AttachedToDoctype: f.AttachedToDocType,
			AttachedToName:    f.AttachedToName,
			FileSize:          f.FileSize,
			ContentType:       f.ContentType,
		})
	}

	return rpc.ApiSuccessWithData("查询成功", &file.ListFilesResp{
		Files: respFiles,
	}), nil
}

// GetFile 获取单个文件信息
func (*Controller) GetFile(ctx context.Context, req *file.GetFileReq) (res *api.ResponseInfo, err error) {
	// 参数校验
	if len(req.Name) == 0 {
		return rpc.ApiError("文件名称不能为空"), nil
	}

	// 调用服务层查询
	resp, err := service.File().Get(ctx, req.Name)
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	// 转换响应
	return rpc.ApiSuccessWithData("查询成功", &file.UploadFileResp{
		Name:              resp.Name,
		FileName:          resp.FileName,
		FileUrl:           resp.FileUrl,
		IsPrivate:         int32(resp.IsPrivate),
		AttachedToDoctype: resp.AttachedToDocType,
		AttachedToName:    resp.AttachedToName,
		FileSize:          resp.FileSize,
		ContentType:       resp.ContentType,
	}), nil
}

// DeleteFile 删除文件
func (*Controller) DeleteFile(ctx context.Context, req *file.DeleteFileReq) (res *api.ResponseInfo, err error) {
	// 参数校验
	if len(req.Name) == 0 {
		return rpc.ApiError("文件名称不能为空"), nil
	}

	// 调用服务层删除
	err = service.File().Delete(ctx, req.Name)
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	return rpc.ApiSuccessWithData("删除成功", &file.DeleteFileResp{
		Message: "文件删除成功",
	}), nil
}
