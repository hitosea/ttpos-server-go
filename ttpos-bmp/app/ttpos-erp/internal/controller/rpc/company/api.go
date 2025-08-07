package company

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/api/company"
	"ttpos-bmp/app/ttpos-erp/internal/service"

	"github.com/gogf/gf/contrib/rpc/grpcx/v2"
)

type Controller struct {
	company.UnimplementedCompanyServiceServer
}

func Register(s *grpcx.GrpcServer) {
	company.RegisterCompanyServiceServer(s.Server, &Controller{})
}

func (*Controller) GetCompanyList(ctx context.Context, req *company.GetCompanyListReq) (res *company.GetCompanyListResp, err error) {
	// 调用服务层, 这里不转换, 直接返回服务层的结果
	res, err = service.Company().GetCompanyList(ctx, req)
	return res, err
}
