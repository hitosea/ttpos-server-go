package erp

import (
	"context"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"

	"google.golang.org/grpc/metadata"

	cc "ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
)

type IErpSrv interface {
	GetCompanyList(ctx context.Context, erpnextSiteCompanyReq req.ErpnextSiteCompanyReq) (resp.ErpnextSiteCompanyResp, error)
	InitShop(ctx cc.Context, initShopReq req.InitShopReq) (resp.InitShopResp, error)
	GetUomList(ctx context.Context, getUomListReq req.GetUomListReq) (resp.GetUomListResp, error)
	GetAttributeList(ctx context.Context, getAttributeListReq req.GetAttributeListReq) (resp.GetAttributeListResp, error)
	SyncUomAndAttribute(ctx cc.Context, syncUomAndAttributeReq req.SyncUomAndAttributeReq) error
	GetPosProfileList(ctx context.Context, getPosProfileListReq req.GetPosProfileListReq) (resp.GetPosProfileListResp, error)
	AddLianPayment(ctx cc.Context, addLianPaymentReq req.ErpnextSiteAddLianLianPaymentReq) error
}
type erpSrv struct {
	dbm *database.DBManager
}

func NewIErpSrv(dbm *database.DBManager) IErpSrv {
	return &erpSrv{
		dbm: dbm,
	}
}

func WithSiteCode(ctx context.Context, siteCode string) context.Context {
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	}
	md.Set("erp_site_code", siteCode)
	return metadata.NewOutgoingContext(ctx, md)
}
