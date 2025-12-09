// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package grab

import (
	"context"

	"ttpos-bmp/app/ttpos-takeout/api/grab/v1"
)

type IGrabV1 interface {
	GetMenu(ctx context.Context, req *v1.GetMenuReq) (res *v1.GetMenuRes, err error)
	IntegrationStatus(ctx context.Context, req *v1.IntegrationStatusReq) (res *v1.IntegrationStatusRes, err error)
	MenuSyncState(ctx context.Context, req *v1.MenuSyncStateReq) (res *v1.MenuSyncStateRes, err error)
	OAuthPartnerWebhook(ctx context.Context, req *v1.OAuthPartnerWebhookReq) (res *v1.OAuthPartnerWebhookRes, err error)
	PushGrabMenuWebhook(ctx context.Context, req *v1.PushGrabMenuWebhookReq) (res *v1.PushGrabMenuWebhookRes, err error)
	PushOrderState(ctx context.Context, req *v1.PushOrderStateReq) (res *v1.PushOrderStateRes, err error)
	SubmitOrder(ctx context.Context, req *v1.SubmitOrderReq) (res *v1.SubmitOrderRes, err error)
}
