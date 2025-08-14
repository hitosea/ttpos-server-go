package stock

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/api"
	"ttpos-bmp/app/ttpos-erp/api/stock"
	"ttpos-bmp/app/ttpos-erp/internal/controller/rpc"
	"ttpos-bmp/app/ttpos-erp/internal/service"

	"github.com/gogf/gf/contrib/rpc/grpcx/v2"
	"google.golang.org/protobuf/types/known/anypb"
)

type Controller struct {
	stock.UnimplementedStockServiceServer
}

func Register(s *grpcx.GrpcServer) {
	stock.RegisterStockServiceServer(s.Server, &Controller{})
}

func (*Controller) SaveMaterialRequest(ctx context.Context, req *stock.SaveMaterialRequestReq) (res *api.ResponseInfo, err error) {
	if len(req.CompanyAbbr) == 0 {
		return rpc.ApiError("公司简称不能为空"), nil
	}
	if len(req.Branch) == 0 {
		return rpc.ApiError("分公司不能为空"), nil
	}
	if len(req.Items) == 0 {
		return rpc.ApiError("物品列表不能为空"), nil
	}
	if req.RequiredBy == 0 {
		return rpc.ApiError("采购日期不能为空"), nil
	}
	if req.TransactionDate == 0 {
		return rpc.ApiError("单据日期不能为空"), nil
	}

	resp, err := service.Stock().CreateMaterialRequest(ctx, req)
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}
	res = rpc.ApiSuccess("保存物品申请单成功")
	res.Data, _ = anypb.New(resp)
	return
}

func (*Controller) GetMaterialRequestList(ctx context.Context, req *stock.GetMaterialRequestListReq) (*api.ResponseInfo, error) {
	if len(req.CompanyAbbr) == 0 {
		return rpc.ApiError("公司简称不能为空"), nil
	}
	resp, err := service.Stock().GetMaterialRequestList(ctx, req)
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}
	res := rpc.ApiSuccess("获取物品申请单列表成功")
	res.Data, _ = anypb.New(resp)
	return res, nil
}
