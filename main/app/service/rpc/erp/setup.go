package erp

import (
	"context"
	"strconv"
	"ttpos-bmp/app/ttpos-erp/api/setup"
	"ttpos-server-go/app/cloud"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"

	"google.golang.org/grpc"
)

func NewErpSetupClient() (setup.SetupServiceClient, *grpc.ClientConn, error) {
	conn, err := cloud.GetRpcConnWithName(cloud.ErpServiceName)
	if err != nil {
		return nil, nil, err
	}
	return setup.NewSetupServiceClient(conn), conn, nil
}

func (s *erpSrv) InitShop(ctx context.Context, initShopReq req.InitShopReq) (resp.InitShopResp, error) {
	var initShopResp resp.InitShopResp
	client, conn, err := NewErpSetupClient()
	if err != nil {
		return initShopResp, err
	}
	defer conn.Close()
	req := &setup.InitShopReq{
		ShopName:    initShopReq.ShopName,
		CompanyAbbr: initShopReq.CompanyAbbr,
		ShopUuid:    strconv.FormatUint(initShopReq.ShopUuid, 10),
		AdminUuid:   strconv.FormatUint(initShopReq.AdminUuid, 10),
	}
	result, err := client.InitShop(WithSiteCode(ctx, initShopReq.SiteCode), req)
	if err != nil {
		return initShopResp, err
	}
	response := &setup.InitShopResp{}
	if err := result.Data.UnmarshalTo(response); err != nil {
		return resp.InitShopResp{}, err
	}
	initShopResp.BranchName = response.BranchName
	return initShopResp, nil
}
