// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package lineman

import (
	"context"

	"ttpos-bmp/app/ttpos-takeout/api/lineman/v1"
)

type ILinemanV1 interface {
	MenuSyncNotification(ctx context.Context, req *v1.MenuSyncNotificationReq) (res *v1.MenuSyncNotificationRes, err error)
	TriggerSyncMenu(ctx context.Context, req *v1.TriggerSyncMenuReq) (res *v1.TriggerSyncMenuRes, err error)
	OAuthToken(ctx context.Context, req *v1.OAuthTokenReq) (res *v1.OAuthTokenRes, err error)
	PlaceOrder(ctx context.Context, req *v1.PlaceOrderReq) (res *v1.PlaceOrderRes, err error)
	OrderStatusUpdate(ctx context.Context, req *v1.OrderStatusUpdateReq) (res *v1.OrderStatusUpdateRes, err error)
	OrderUpdate(ctx context.Context, req *v1.OrderUpdateReq) (res *v1.OrderUpdateRes, err error)
}
