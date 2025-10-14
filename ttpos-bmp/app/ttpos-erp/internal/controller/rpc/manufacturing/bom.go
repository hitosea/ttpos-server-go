package manufacturing

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/api"
	"ttpos-bmp/app/ttpos-erp/api/manufacturing"
	"ttpos-bmp/app/ttpos-erp/internal/controller/rpc"
	"ttpos-bmp/app/ttpos-erp/internal/service"

	"github.com/gogf/gf/contrib/rpc/grpcx/v2"
	"github.com/gogf/gf/v2/errors/gerror"
)

// Controller BOM服务控制器
type Controller struct {
	manufacturing.UnimplementedBomServiceServer
}

// Register 注册BOM服务到gRPC服务器
func Register(s *grpcx.GrpcServer) {
	manufacturing.RegisterBomServiceServer(s.Server, &Controller{})
}

func (s *Controller) GetBomList(ctx context.Context, req *manufacturing.GetBomListReq) (*api.ResponseInfo, error) {
	// 参数校验
	if err := s.validateGetBomListReq(req); err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	res, err := service.Bom().GetBomList(ctx, req)
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}
	// 返回成功响应
	return rpc.ApiSuccessWithData("获取BOM列表成功", res), nil
}

// GetBom 根据BOM名称获取单个BOM信息
// 参数：ctx 上下文，req 包含BOM名称
// 返回：响应信息和错误
func (s *Controller) GetBom(ctx context.Context, req *manufacturing.GetBomReq) (*api.ResponseInfo, error) {
	if req == nil {
		return rpc.ApiError("请求参数不能为空"), nil
	}

	if len(req.BomName) == 0 {
		return rpc.ApiError("BOM名称不能为空"), nil
	}

	bomInfo, err := service.Bom().GetBom(ctx, req)
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	// 转换为 protobuf 响应格式
	resp := &manufacturing.GetBomResp{
		BomInfo: &manufacturing.BomInfo{
			ItemCode:  bomInfo.Item,
			BomName:   bomInfo.Name,
			Uom:       bomInfo.Uom,
			Company:   bomInfo.Company,
			Quantity:  bomInfo.Quantity,
			IsActive:  bomInfo.IsActive,
			IsDefault: bomInfo.IsDefault,
			Items:     make([]*manufacturing.BomItem, 0),
		},
	}

	// 转换BOM明细项目
	for _, item := range bomInfo.Items {
		resp.BomInfo.Items = append(resp.BomInfo.Items, &manufacturing.BomItem{
			ItemCode: item.ItemCode,
			Rate:     item.Rate,
			Qty:      item.Qty,
			Uom:      item.Uom,
		})
	}

	return rpc.ApiSuccessWithData("获取BOM信息成功", resp), nil
}

func (s *Controller) SaveBom(ctx context.Context, req *manufacturing.SaveBomReq) (*api.ResponseInfo, error) {
	// 参数校验
	if err := s.validateSaveBomReq(req); err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	res, err := service.Bom().SaveBom(ctx, req)
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}
	// 返回成功响应
	return rpc.ApiSuccessWithData("创建BOM成功", res), nil
}

// validateGetBomListReq 验证获取BOM列表请求参数
// 参数：req 获取BOM列表请求
// 返回：错误信息
func (s *Controller) validateGetBomListReq(req *manufacturing.GetBomListReq) error {
	if req == nil {
		return gerror.New("请求参数不能为空")
	}

	// 公司简称为必填参数
	if len(req.CompanyAbbr) == 0 {
		return gerror.New("公司简称不能为空")
	}

	return nil
}

// validateSaveBomReq 验证保存BOM请求参数
// 参数：req 保存BOM请求
// 返回：错误信息
func (s *Controller) validateSaveBomReq(req *manufacturing.SaveBomReq) error {
	if req == nil {
		return gerror.New("请求参数不能为空")
	}

	// 商品编码为必填参数
	if len(req.ItemCode) == 0 {
		return gerror.New("商品编码不能为空")
	}

	// 公司简称为必填参数
	if len(req.CompanyAbbr) == 0 {
		return gerror.New("公司简称不能为空")
	}

	// 数量必须大于0
	if req.Quantity <= 0 {
		return gerror.New("数量必须大于0")
	}

	// 物品列表为必填参数
	if len(req.Items) == 0 {
		return gerror.New("BOM物品列表不能为空")
	}

	// 验证BOM物品明细
	for i, item := range req.Items {
		if err := s.validateBomItem(item, i+1); err != nil {
			return err
		}
	}

	return nil
}

// validateBomItem 验证BOM物品明细
// 参数：item BOM物品明细，index 物品索引（用于错误提示）
// 返回：错误信息
func (s *Controller) validateBomItem(item *manufacturing.BomItem, index int) error {
	if item == nil {
		return gerror.Newf("第%d个BOM物品明细不能为空", index)
	}

	// 商品编码为必填
	if len(item.ItemCode) == 0 {
		return gerror.Newf("第%d个BOM物品的商品编码不能为空", index)
	}

	// 比率必须大于0
	//if item.Rate <= 0 {
	//	return gerror.Newf("第%d个BOM物品的比率必须大于0", index)
	//}

	// 数量必须大于0
	if item.Qty <= 0 {
		return gerror.Newf("第%d个BOM物品的数量必须大于0", index)
	}

	// 单位不能为空
	if len(item.Uom) == 0 {
		return gerror.Newf("第%d个BOM物品的单位不能为空", index)
	}

	return nil
}
