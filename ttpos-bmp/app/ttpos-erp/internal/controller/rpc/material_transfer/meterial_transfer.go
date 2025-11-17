package material_transfer

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/api"
	"ttpos-bmp/app/ttpos-erp/api/material_transfer"
	"ttpos-bmp/app/ttpos-erp/internal/controller/rpc"
	"ttpos-bmp/app/ttpos-erp/internal/service"

	"github.com/gogf/gf/contrib/rpc/grpcx/v2"
)

type Controller struct {
	material_transfer.UnimplementedMaterialTransferServiceServer
}

// Register 注册材料调拨服务到gRPC服务器
func Register(s *grpcx.GrpcServer) {
	material_transfer.RegisterMaterialTransferServiceServer(s.Server, &Controller{})
}

func (*Controller) MaterialTransfer(ctx context.Context, req *material_transfer.MaterialTransferReq) (*api.ResponseInfo, error) {
	//输入参数校验

	resp, err := service.MaterialTransfer().MaterialTransfer(ctx, req)
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}
	return rpc.ApiSuccessWithData("材料调拨成功", resp), nil
}
