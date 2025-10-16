package stock

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/api/item"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto/erp"
	"ttpos-bmp/app/ttpos-erp/internal/service"

	"github.com/gogf/gf/v2/errors/gerror"
)

var (
	Product = new(sProduct)
)

type sProduct struct {
}

func init() {
	service.RegisterProduct(Product)
}

func (s *sProduct) UpdateProduct(ctx context.Context, req *item.UpdateProductReq) (*item.UpdateProductResp, error) {
	_, err := service.Document().Update(ctx, &erp.ErpReq{
		DocType: erp.DocTypeItem,
		Name:    req.ItemCode,
	}, &erp.Item{
		CustomNotForSale:   req.NotForSale,
		CustomInternalCode: req.InternalCode,
		Disabled:           req.Disabled,
	})
	if err != nil {
		return nil, gerror.Wrapf(err, "更新商品信息失败")
	}
	return &item.UpdateProductResp{
		ItemCode:     req.ItemCode,
		NotForSale:   req.NotForSale,
		InternalCode: req.InternalCode,
		Disabled:     req.Disabled,
	}, nil
}
