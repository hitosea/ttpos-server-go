package item

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/internal/service"

	"ttpos-bmp/app/ttpos-erp/api/item/v1"
)

func (c *ControllerV1) ItemSearch(ctx context.Context, req *v1.ItemSearchReq) (res *v1.ItemSearchRes, err error) {
	service.Item().SyncDelay()
	return nil, nil
}
