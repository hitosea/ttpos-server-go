package adapter

import (
	"context"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/modules/order_core/service"
)

type LegacyOrderAdapter struct {
	coreSrv service.ICoreOrderService
}

func NewLegacyOrderAdapter(coreSrv service.ICoreOrderService) ILegacyOrderAdapter {
	return &LegacyOrderAdapter{coreSrv: coreSrv}
}

func (s *LegacyOrderAdapter) CreateDeskOrder(ctx context.Context, req *req.DeskOrderCreateReq) (*resp.CreateDeskOrderResp, error) {
	// Convert request
	coreReq := ToCoreCreateDeskOrderReq(req)

	// Call Core
	coreResp, err := s.coreSrv.CreateOrder(ctx, coreReq)
	if err != nil {
		return nil, err
	}

	var saleOrderUuid uint64
	if len(coreResp.OrderUuids) > 0 {
		saleOrderUuid = coreResp.OrderUuids[0]
	}

	return &resp.CreateDeskOrderResp{
		SaleBillUuid:  coreResp.BillUuid,
		SaleOrderUuid: saleOrderUuid,
	}, nil
}

func (s *LegacyOrderAdapter) CreateInstantOrder(ctx context.Context) (*resp.CreateInstantOrderResp, error) {
	// Convert request (no input req for instant order in legacy interface)
	coreReq := ToCoreCreateInstantOrderReq()

	// Call Core
	coreResp, err := s.coreSrv.CreateOrder(ctx, coreReq)
	if err != nil {
		return nil, err
	}

	var saleOrderUuid uint64
	if len(coreResp.OrderUuids) > 0 {
		saleOrderUuid = coreResp.OrderUuids[0]
	}

	return &resp.CreateInstantOrderResp{
		SaleBillUuid:  coreResp.BillUuid,
		SaleOrderUuid: saleOrderUuid,
	}, nil
}

func (s *LegacyOrderAdapter) PayOrder(ctx context.Context, req *req.InstantOrderPaymentCreateReq) (*resp.InstantOrderPaymentInfoResp, error) {
	// TODO: PayOrder implementation (Task 1.3)
	return nil, nil
}

func (s *LegacyOrderAdapter) CancelOrder(ctx context.Context, req *req.OrderCancelReq) error {
	// TODO: CancelOrder implementation (Task 1.3)
	return nil
}

