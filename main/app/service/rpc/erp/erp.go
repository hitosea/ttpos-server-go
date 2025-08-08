package erp

import (
	"context"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"

	"google.golang.org/grpc/metadata"
)

type IErpSrv interface {
	GetCompanyList(ctx context.Context, erpnextSiteCompanyReq req.ErpnextSiteCompanyReq) (resp.ErpnextSiteCompanyResp, error)
}
type erpSrv struct {
}

func NewIErpSrv() IErpSrv {
	return &erpSrv{}
}

func WithSiteCode(ctx context.Context, siteCode string) context.Context {
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	}
	md.Set("erp_site_code", siteCode)
	return metadata.NewOutgoingContext(ctx, md)
}
