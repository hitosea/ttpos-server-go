// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package callback

import (
	"context"

	"ttpos-bmp/app/ttpos-erp/api/callback/v1"
)

type ICallbackV1 interface {
	DocChange(ctx context.Context, req *v1.DocChangeReq) (res *v1.DocChangeRes, err error)
}
