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
	itemInfo := &erp.Item{
		CustomNotForSale: req.NotForSale,
		Disabled:         req.Disabled,
	}

	if len(req.Attributes) > 0 {
		attributes := make([]erp.ItemVariantAttribute, 0)
		for _, attr := range req.Attributes {
			attributes = append(attributes, erp.ItemVariantAttribute{
				Attribute:      attr.Attribute,
				AttributeValue: attr.AttributeValue,
			})
		}
		itemInfo.Attributes = attributes
	}

	if len(req.InternalCode) > 0 {
		itemInfo.CustomInternalCode = req.InternalCode
	}

	_, err := service.Document().Update(ctx, &erp.ErpReq{
		DocType: erp.DocTypeItem,
		Name:    req.ItemCode,
	}, itemInfo)
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
