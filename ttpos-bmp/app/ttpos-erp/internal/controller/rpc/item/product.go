package item

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/api"
	"ttpos-bmp/app/ttpos-erp/api/item"
	"ttpos-bmp/app/ttpos-erp/internal/controller/rpc"
	"ttpos-bmp/app/ttpos-erp/internal/service"

	"github.com/gogf/gf/v2/errors/gerror"
)

type ProductController struct {
	item.UnimplementedProductServiceServer
}

func (*ProductController) UpdateProduct(ctx context.Context, req *item.UpdateProductReq) (*api.ResponseInfo, error) {
	// 参数验证
	if req == nil {
		return nil, gerror.New("请求参数不能为空")
	}
	if req.ItemCode == "" {
		return nil, gerror.New("物品编码不能为空")
	}
	resp, err := service.Product().UpdateProduct(ctx, req)
	if err != nil {
		return nil, gerror.Wrapf(err, "更新商品信息失败")
	}

	return rpc.ApiSuccessWithData("更新商品信息成功", resp), nil
}
