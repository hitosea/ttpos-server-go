package erp

import (
	"context"
	companyApi "ttpos-bmp/app/ttpos-erp/api/company"
	"ttpos-server-go/app/cloud"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"

	"google.golang.org/grpc"
)

func NewErpCompanyClient() (companyApi.CompanyServiceClient, *grpc.ClientConn, error) {
	conn, err := cloud.GetRpcConnWithName(cloud.ErpServiceName)
	if err != nil {
		return nil, nil, err
	}
	return companyApi.NewCompanyServiceClient(conn), conn, nil
}

// GetCompanyList 获取公司列表
func (s *erpSrv) GetCompanyList(ctx context.Context, erpnextSiteCompanyReq req.ErpnextSiteCompanyReq) (resp.ErpnextSiteCompanyResp, error) {
	var companyResp resp.ErpnextSiteCompanyResp
	client, conn, err := NewErpCompanyClient()
	if err != nil {
		return companyResp, err
	}
	defer conn.Close()

	req := &companyApi.GetCompanyListReq{
		CompanyName:   erpnextSiteCompanyReq.CompanyName,
		CompanyAbbr:   erpnextSiteCompanyReq.CompanyAbbr,
		ParentCompany: erpnextSiteCompanyReq.ParentCompany,
	}
	result, err := client.GetCompanyList(WithSiteCode(ctx, erpnextSiteCompanyReq.SiteCode), req)
	if err != nil {
		return companyResp, err
	}
	// 反序列化响应数据
	response := &companyApi.GetCompanyListResp{}
	if err := result.Data.UnmarshalTo(response); err != nil {
		return resp.ErpnextSiteCompanyResp{}, err
	}
	for _, company := range response.CompanyList {
		companyResp.List = append(companyResp.List, resp.ErpnextSiteCompany{
			CompanyName: company.CompanyName,
			CompanyAbbr: company.CompanyAbbr,
		})
	}
	return companyResp, nil
}
