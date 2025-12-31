package rpc

import (
	"ttpos-bmp/app/ttpos-erp/api/selling"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/service/rpc/erp"
	appContext "ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"

	"go.uber.org/zap"
)

// ErpRpcAdapter ERP RPC 适配器（基础设施层）
// @version v2.12.0
// @spec story-erp-grab-invoice-sync
// @responsibility 封装 ERP RPC 调用，作为 takeout 模块与 ERP 系统的适配层
type ErpRpcAdapter struct {
	dbm *database.DBManager
}

// NewErpRpcAdapter 创建 ERP RPC 适配器
func NewErpRpcAdapter(dbm *database.DBManager) *ErpRpcAdapter {
	return &ErpRpcAdapter{
		dbm: dbm,
	}
}

// SavePosInvoice 保存 POS Invoice 到 ERP
// ctx 上下文（需包含 company、staff 等信息）
// req POS Invoice 请求参数
//
//	返回 ERP 发票信息（products_invoice_name、material_invoice_name）
func (a *ErpRpcAdapter) SavePosInvoice(
	ctx appContext.Context,
	req req.SavePosInvoiceReq,
) (*selling.SavePosInvoiceResp, error) {
	// 调用全局 ERP Service
	resp, err := erp.NewIErpSrv(a.dbm).SavePosInvoice(ctx, req)
	if err != nil {
		logger.Logger.Error("调用 ERP SavePosInvoice 失败",
			zap.Uint64("company_uuid", ctx.GetCompanyUuid()),
			zap.String("order_no", req.OrderNo),
			zap.Error(err))
		return nil, errors.WithMessage(err, "保存 POS Invoice 到 ERP 失败")
	}

	return resp, nil
}
