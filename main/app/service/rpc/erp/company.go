package erp

import (
	"context"
	companyApi "ttpos-bmp/app/ttpos-erp/api/company"
	"ttpos-server-go/app/cloud"
	"ttpos-server-go/pkg/logger"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func NewErpCompanyClient() (companyApi.CompanyServiceClient, *grpc.ClientConn, error) {
	conn, err := cloud.GetRpcConnWithName(cloud.ErpServiceName)
	if err != nil {
		return nil, nil, err
	}
	return companyApi.NewCompanyServiceClient(conn), conn, nil
}

// GetCompanyList 获取公司列表  FIXME ,增加查询参数
func (s *erpSrv) GetCompanyList(ctx context.Context) error {
	client, conn, err := NewErpCompanyClient()
	if err != nil {
		return err
	}
	defer conn.Close()
	//获取对应的 site_code FIXME 从配置中读取
	siteCode := "1"
	req := &companyApi.GetCompanyListReq{}
	resp, err := client.GetCompanyList(WithSiteCode(ctx, siteCode), req)
	if err != nil {
		return err
	}
	logger.Logger.Info("GetCompanyList resp: %v", zap.Any("resp", resp))
	return nil
}
