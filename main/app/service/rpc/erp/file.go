package erp

import (
	"context"
	"errors"

	"ttpos-bmp/app/ttpos-erp/api/file"
	"ttpos-server-go/app/cloud"
	pkgCtx "ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/logger"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// NewErpFileClient 创建 ERP 文件服务客户端
func NewErpFileClient() (file.FileServiceClient, *grpc.ClientConn, error) {
	conn, err := cloud.GetRpcConnWithName(cloud.ErpServiceName)
	if err != nil {
		return nil, nil, err
	}
	return file.NewFileServiceClient(conn), conn, nil
}

// UploadFileUrl 通过 URL 上传文件到 ERPNext
func (s *erpSrv) UploadFileUrl(ctx pkgCtx.Context, req *file.UploadFileUrlReq) (*file.UploadFileResp, error) {
	client, conn, err := NewErpFileClient()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	companySetting := ctx.GetCompany().CompanySetting

	result, err := client.UploadFileUrl(WithSiteCode(context.Background(), companySetting.ErpnextSiteCode), req)
	if err != nil {
		logger.Logger.Error("UploadFileUrl-UploadFileUrl", zap.Any("err", err))
		return nil, err
	}
	if result.Code != "0" {
		logger.Logger.Error("UploadFileUrl-UploadFileUrl", zap.String("message", result.GetMessage()))
		return nil, errors.New("调用erp接口失败-3001-" + result.GetMessage())
	}
	if result.Data != nil {
		var resp file.UploadFileResp
		if err := result.Data.UnmarshalTo(&resp); err != nil {
			logger.Logger.Error("UploadFileUrl-UnmarshalTo", zap.Any("err", err))
			return nil, err
		}
		return &resp, nil
	}
	return &file.UploadFileResp{}, nil
}
