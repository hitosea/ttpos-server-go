package erp

import (
	"context"
	"errors"

	"ttpos-bmp/app/ttpos-erp/api/delivery_note"
	"ttpos-server-go/app/cloud"
	pkgCtx "ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/logger"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func NewErpDeliveryNoteClient() (delivery_note.DeliveryNoteServiceClient, *grpc.ClientConn, error) {
	conn, err := cloud.GetRpcConnWithName(cloud.ErpServiceName)
	if err != nil {
		return nil, nil, err
	}
	return delivery_note.NewDeliveryNoteServiceClient(conn), conn, nil
}

// GetDeliveryNoteList 获取送货单列表
func (s *erpSrv) GetDeliveryNoteList(ctx pkgCtx.Context, getDeliveryNoteListReq *delivery_note.GetDeliveryNoteListReq) (*delivery_note.GetDeliveryNoteListResp, error) {
	client, conn, err := NewErpDeliveryNoteClient()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	companySetting := ctx.GetCompany().CompanySetting

	result, err := client.GetDeliveryNoteList(WithSiteCode(context.Background(), companySetting.ErpnextSiteCode), getDeliveryNoteListReq)
	if err != nil {
		return nil, err
	}
	if result.Code != "0" {
		logger.Logger.Error("GetDeliveryNoteList-GetDeliveryNoteList", zap.Any("err", err), zap.String("message", result.GetMessage()))
		return nil, errors.New("调用erp接口失败-30001-" + result.GetMessage())
	}
	if result.Data != nil {
		var resp delivery_note.GetDeliveryNoteListResp
		if err := result.Data.UnmarshalTo(&resp); err != nil {
			logger.Logger.Error("GetDeliveryNoteList-UnmarshalTo", zap.Any("err", err))
			return nil, err
		}
		return &resp, nil
	}
	return &delivery_note.GetDeliveryNoteListResp{}, nil
}

// GetDeliveryNote 获取单个送货单详情
func (s *erpSrv) GetDeliveryNote(ctx pkgCtx.Context, dnName string) (*delivery_note.DeliveryNote, error) {
	client, conn, err := NewErpDeliveryNoteClient()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	companySetting := ctx.GetCompany().CompanySetting

	result, err := client.GetDeliveryNote(WithSiteCode(context.Background(), companySetting.ErpnextSiteCode), &delivery_note.GetDeliveryNoteReq{
		DeliveryNoteName: dnName,
	})
	if err != nil {
		return nil, err
	}
	if result.Code != "0" {
		logger.Logger.Error("GetDeliveryNote-GetDeliveryNote", zap.Any("err", err), zap.String("message", result.GetMessage()))
		return nil, errors.New("调用erp接口失败-30002-" + result.GetMessage())
	}
	if result.Data != nil {
		var resp delivery_note.GetDeliveryNoteResp
		if err := result.Data.UnmarshalTo(&resp); err != nil {
			logger.Logger.Error("GetDeliveryNote-UnmarshalTo", zap.Any("err", err))
			return nil, err
		}
		if resp.DeliveryNote == nil {
			return nil, errors.New("送货单不存在: " + dnName)
		}
		return resp.DeliveryNote, nil
	}
	return nil, errors.New("送货单不存在: " + dnName)
}
