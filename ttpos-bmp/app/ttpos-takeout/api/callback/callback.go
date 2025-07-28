// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package callback

import (
	"context"

	"ttpos-bmp/app/ttpos-takeout/api/callback/v1"
)

type ICallbackV1 interface {
	SkootarStatus(ctx context.Context, req *v1.SkootarStatusReq) (res *v1.SkootarStatusRes, err error)
}
