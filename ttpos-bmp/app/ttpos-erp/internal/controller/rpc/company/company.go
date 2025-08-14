package company

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/api"
	"ttpos-bmp/app/ttpos-erp/api/company"
	"ttpos-bmp/app/ttpos-erp/internal/controller/rpc"
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
	if resp, err := service.Company().GetCompanyList(ctx, req); err != nil {
		res = rpc.ApiError(err.Error())
	} else {
		res = rpc.ApiSuccess("获取成功")
		res.Data, _ = anypb.New(resp)
	}
	return
}
