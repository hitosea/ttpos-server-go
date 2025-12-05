package adapter

import (
	"context"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
)

// ILegacyOrderAdapter defines the interface for adapting legacy order service calls to the new Order Core module.
type ILegacyOrderAdapter interface {
	// CreateDeskOrder creates a desk order.
	CreateDeskOrder(ctx context.Context, req *req.DeskOrderCreateReq) (*resp.CreateDeskOrderResp, error)

	// CreateInstantOrder creates an instant order.
	CreateInstantOrder(ctx context.Context) (*resp.CreateInstantOrderResp, error)

	// PayOrder processes the payment for an order.
	PayOrder(ctx context.Context, req *req.InstantOrderPaymentCreateReq) (*resp.InstantOrderPaymentInfoResp, error)

	// CancelOrder cancels an order.
	CancelOrder(ctx context.Context, req *req.OrderCancelReq) error
}

