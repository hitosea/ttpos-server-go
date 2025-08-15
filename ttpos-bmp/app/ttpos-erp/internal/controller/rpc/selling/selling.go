package selling

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/api"
	"ttpos-bmp/app/ttpos-erp/api/selling"
	"ttpos-bmp/app/ttpos-erp/internal/consts"
	"ttpos-bmp/app/ttpos-erp/internal/controller/rpc"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto/erp"
	"ttpos-bmp/app/ttpos-erp/internal/service"

	"github.com/gogf/gf/contrib/rpc/grpcx/v2"
)

// Controller 销售服务控制器
type Controller struct {
	selling.UnsafeSellingServiceServer
}

// Register 注册销售服务到gRPC服务器
func Register(s *grpcx.GrpcServer) {
	selling.RegisterSellingServiceServer(s.Server, &Controller{})
}

func (*Controller) GetPosProfileList(ctx context.Context, req *selling.PosProfileReq) (*api.ResponseInfo, error) {
	if len(req.CompanyAbbr) == 0 {
		return rpc.ApiError("公司简称不能为空"), nil
	}

	// 调用服务层获取数据
	dataList, err := service.Selling().GetPosProfileList(ctx, req)
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	// 返回成功响应
	return rpc.ApiSuccessWithData("获取POS配置文件列表成功", dataList), nil
}

// CreatePaymentAccount 创建支付账户
// 参数：ctx 上下文，req 创建支付账户请求
// 返回：响应信息和错误
func (c *Controller) CreatePaymentAccount(ctx context.Context, req *selling.CreatePaymentAccountReq) (*api.ResponseInfo, error) {
	// 参数验证
	if len(req.CompanyAbbr) == 0 {
		return rpc.ApiError("公司简称不能为空"), nil
	}
	if len(req.PaymentType) == 0 {
		return rpc.ApiError("支付类型不能为空"), nil
	}

	// 处理连连支付前缀
	paymentType := req.PaymentType
	if req.PaymentSource == "2" {
		paymentType = consts.LianlianPayPrefix + paymentType
	}

	// 调用服务层创建支付账户
	err := service.Selling().CreateDefaultModePaymentAccount(ctx, &erp.CreateModePaymentAccountInp{
		CompanyAbbr: req.CompanyAbbr,
		PaymentType: paymentType,
	})
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	// 返回成功响应
	return rpc.ApiSuccess("创建支付账户成功"), nil
}
