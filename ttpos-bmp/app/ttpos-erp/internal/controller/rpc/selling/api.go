package selling

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/api"
	"ttpos-bmp/app/ttpos-erp/api/selling"
	"ttpos-bmp/app/ttpos-erp/internal/controller/rpc"
	"ttpos-bmp/app/ttpos-erp/internal/service"

	"github.com/gogf/gf/contrib/rpc/grpcx/v2"
	"google.golang.org/protobuf/types/known/anypb"
)

type Controller struct {
	selling.UnsafeSellingServiceServer
}

func Register(s *grpcx.GrpcServer) {
	selling.RegisterSellingServiceServer(s.Server, &Controller{})
}

func (*Controller) GetPosProfileList(ctx context.Context, req *selling.PosProfileReq) (res *api.ResponseInfo, err error) {
	if dataList, err := service.Selling().GetPosProfileList(ctx, req); err == nil {
		res = rpc.ApiSuccess("获取成功")
		res.Data, _ = anypb.New(dataList)
	} else {
		res = rpc.ApiError(err.Error())
	}
	return
}
