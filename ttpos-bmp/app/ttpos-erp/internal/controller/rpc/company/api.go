package company

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/api"
	"ttpos-bmp/app/ttpos-erp/api/company"
	"ttpos-bmp/app/ttpos-erp/internal/service"

	"github.com/gogf/gf/contrib/rpc/grpcx/v2"
	"google.golang.org/protobuf/types/known/anypb"
)

type Controller struct {
	company.UnimplementedCompanyServiceServer
}

func Register(s *grpcx.GrpcServer) {
	company.RegisterCompanyServiceServer(s.Server, &Controller{})
}

func (*Controller) GetCompanyList(ctx context.Context, req *company.GetCompanyListReq) (res *api.ResponseInfo, err error) {
	// 调用服务层, 这里不转换, 直接返回服务层的结果
	resp, err := service.Company().GetCompanyList(ctx, req)
	if err != nil {
		return &api.ResponseInfo{
			Code:    "1",
			Message: err.Error(),
		}, err
	}
	res = &api.ResponseInfo{
		Code:    "0",
		Message: "success",
	}
	res.Data, _ = anypb.New(resp)
	return res, err
}
