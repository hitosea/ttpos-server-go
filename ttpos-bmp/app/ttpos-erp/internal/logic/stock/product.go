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

func (s *sProduct) SetProductForSale(ctx context.Context, req *item.SetProductForSaleReq) (*item.SetProductForSaleResp, error) {
	_, err := service.Document().Update(ctx, &erp.ErpReq{
		DocType: erp.DocTypeItem,
		Name:    req.ItemCode,
	}, &erp.Item{
		CustomNotForSale: req.NotForSale,
	})
	if err != nil {
		return nil, gerror.Wrapf(err, "更新物品禁售状态失败")
	}
	return &item.SetProductForSaleResp{
		ItemCode:   req.ItemCode,
		NotForSale: req.NotForSale,
	}, nil
}
