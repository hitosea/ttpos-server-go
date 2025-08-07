package company

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/api"
	"ttpos-bmp/app/ttpos-erp/api/company"
	"ttpos-bmp/app/ttpos-erp/internal/service"

	"github.com/gogf/gf/contrib/rpc/grpcx/v2"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
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

	res = &api.ResponseInfo{
		Code:    "0",
		Message: "success",
	}
	res.Data, _ = anypb.New(resp)
	return res, err
}

func (*Controller) CreateBranch(ctx context.Context, req *company.CreateBranchReq) (res *api.ResponseInfo, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}
