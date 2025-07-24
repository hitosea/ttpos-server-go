// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package item

import (
	"context"

	"ttpos-bmp/app/ttpos-erp/api/item/v1"
)

type IItemV1 interface {
	ItemSearch(ctx context.Context, req *v1.ItemSearchReq) (res *v1.ItemSearchRes, err error)
}
