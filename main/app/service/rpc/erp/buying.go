package erp

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/api/buying"
	"ttpos-server-go/app/cloud"
	cc "ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/logger"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func NewErpBuyingClient() (buying.BuyingServiceClient, *grpc.ClientConn, error) {
	conn, err := cloud.GetRpcConnWithName(cloud.ErpServiceName)
	if err != nil {
		return nil, nil, err
	}
	return buying.NewBuyingServiceClient(conn), conn, nil
}

// GetMaterialRequestList 获取物品申请单列表
func (s *erpSrv) GetSupplierList(ctx cc.Context) (*buying.GetSupplierListResp, error) {
	client, conn, err := NewErpBuyingClient()
	if err != nil {
		return &buying.GetSupplierListResp{}, err
	}
	defer conn.Close()

	companySetting := ctx.GetCompany().CompanySetting
	result, err := client.GetSupplierList(WithSiteCode(context.Background(), companySetting.ErpnextSiteCode), &buying.GetSupplierListReq{
		CompanyAbbr: companySetting.ErpnextCompanyAbbr,
	})
	if err != nil {
		return &buying.GetSupplierListResp{}, err
	}
	if result.Data != nil {
		var resp buying.GetSupplierListResp
		if err := result.Data.UnmarshalTo(&resp); err != nil {
			logger.Logger.Error("GetSupplierList-UnmarshalTo", zap.Any("err", err))
			return &buying.GetSupplierListResp{}, err
		}
		return &resp, nil
	}
	return &buying.GetSupplierListResp{}, nil
}
