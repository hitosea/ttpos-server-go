package erp

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/api/selling"
	"ttpos-server-go/app/cloud"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/pkg/logger"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func NewErpSellingClient() (selling.SellingServiceClient, *grpc.ClientConn, error) {
	conn, err := cloud.GetRpcConnWithName(cloud.ErpServiceName)
	if err != nil {
		return nil, nil, err
	}
	return selling.NewSellingServiceClient(conn), conn, nil
}

func (s *erpSrv) GetPosProfileList(ctx context.Context, getPosProfileListReq req.GetPosProfileListReq) (resp.GetPosProfileListResp, error) {
	var getPosProfileListResp resp.GetPosProfileListResp
	client, conn, err := NewErpSellingClient()
	if err != nil {
		return getPosProfileListResp, err
	}
	defer conn.Close()
	req := &selling.PosProfileReq{
		Name:        getPosProfileListReq.PosProfileName,
		Company:     getPosProfileListReq.Company,
		CompanyAbbr: getPosProfileListReq.CompanyAbbr,
	}
	result, err := client.GetPosProfileList(WithSiteCode(ctx, getPosProfileListReq.SiteCode), req)
	if err != nil {
		return getPosProfileListResp, err
	}
	if result.Data != nil {
		var posProfileListResp selling.PosProfileListResp
		if err := result.Data.UnmarshalTo(&posProfileListResp); err != nil {
			logger.Logger.Error("GetPosProfileList-UnmarshalTo", zap.Any("err", err))
			return getPosProfileListResp, err
		}
		for _, profile := range posProfileListResp.ProfileList {
			getPosProfileListResp.ProfileList = append(getPosProfileListResp.ProfileList, resp.PosProfileInfo{
				Name:      profile.Name,
				Company:   profile.Company,
				Branch:    profile.Branch,
				Warehouse: profile.Warehouse,
			})
		}
	}
	return getPosProfileListResp, nil
}
