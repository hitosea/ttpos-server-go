package warehouse

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/api"
	"ttpos-bmp/app/ttpos-erp/api/warehouse"
	"ttpos-bmp/app/ttpos-erp/internal/controller/rpc"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto/erp"
	"ttpos-bmp/app/ttpos-erp/internal/service"

	"github.com/gogf/gf/contrib/rpc/grpcx/v2"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"google.golang.org/protobuf/types/known/anypb"
)

type Controller struct {
	warehouse.UnimplementedWarehouseServiceServer
}

func Register(s *grpcx.GrpcServer) {
	warehouse.RegisterWarehouseServiceServer(s.Server, &Controller{})
}
func (*Controller) CreateWarehouse(ctx context.Context, req *warehouse.WarehouseInfo) (*api.ResponseInfo, error) {
	service.Warehouse().CreateWarehouse(ctx, &erp.CreateWarehouseInp{
		Company:     req.Company,
		CompanyAbbr: req.CompanyAbbr,
		Branch:      req.Branch,
		AliasName:   req.AliasName,
		WhType:      req.WarehouseType,
	})
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}

func (*Controller) GetWarehouseList(ctx context.Context, req *warehouse.GetWarehouseListReq) (res *api.ResponseInfo, err error) {
	if dataList, err := service.Warehouse().GetWarehouseList(ctx, req); err != nil {
		res = rpc.ApiSuccess("获取属性列表成功")
		res.Data, _ = anypb.New(dataList)
	} else {
		res = rpc.ApiError(err.Error())
	}
	return
}
