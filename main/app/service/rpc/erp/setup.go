package erp

import (
	"context"
	"errors"
	"strconv"
	"ttpos-bmp/app/ttpos-erp/api/setup"
	"ttpos-server-go/app/cloud"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/repository"
	cc "ttpos-server-go/pkg/context"

	"google.golang.org/grpc"
)

func NewErpSetupClient() (setup.SetupServiceClient, *grpc.ClientConn, error) {
	conn, err := cloud.GetRpcConnWithName(cloud.ErpServiceName)
	if err != nil {
		return nil, nil, err
	}
	return setup.NewSetupServiceClient(conn), conn, nil
}

func (s *erpSrv) InitShop(ctx cc.Context, initShopReq req.InitShopReq) (resp.InitShopResp, error) {
	companySettingRepo := repository.NewCompanySettingRepo(s.dbm.GetDB(0))
	// 判断是否已经有店铺
	companySetting, _ := companySettingRepo.GetOne(companySettingRepo.WhereErpnextCompanyAbbr(initShopReq.CompanyAbbr))
	if companySetting.Uuid != 0 {
		return resp.InitShopResp{}, errors.New("所属erpnext公司已授权其他商家")
	}

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
	result, err := client.InitShop(WithSiteCode(context.Background(), initShopReq.SiteCode), req)
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
