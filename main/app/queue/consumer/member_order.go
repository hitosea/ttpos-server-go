package consumer

import (
	"encoding/json"
	"fmt"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"

	"go.uber.org/zap"
)

// MemberOrderCancelParams 会员订单取消队列参数
type MemberOrderCancelParams struct {
	MemberSaleOrderUuid uint64 `json:"member_sale_order_uuid"` // 会员订单UUID
	CompanyUuid         uint64 `json:"company_uuid"`           // 公司UUID
	Reason              string `json:"reason"`                 // 取消原因
}

// ProcessMemberOrderCancel 处理会员订单自动取消
func ProcessMemberOrderCancel(paramsJson string) bool {
	var params MemberOrderCancelParams
	if err := json.Unmarshal([]byte(paramsJson), &params); err != nil {
		logger.Logger.Error("ProcessMemberOrderCancel 处理会员订单自动取消失败", zap.String("paramsJson", paramsJson), zap.Error(err))
		return true
	}

	// 获取DB
	db := database.GetDBManager(config.DatabaseConf{}).GetDB(params.CompanyUuid)

	// 1. 获取会员端销售订单
	memberSaleOrder, err := repository.NewMemberSaleOrderRepo(db).GetMemberSaleOrderRecord(params.MemberSaleOrderUuid)
	if err != nil {
		fmt.Println("获取会员端销售订单失败", zap.Uint64("memberSaleOrderUuid", params.MemberSaleOrderUuid), zap.Error(err))
		logger.Logger.Error("获取会员端销售订单失败", zap.Uint64("memberSaleOrderUuid", params.MemberSaleOrderUuid), zap.Error(err))
		return false
	}

	// 2. 检查订单状态是否可以取消 - 如果订单状态不是选购中，则不进行取消
	if memberSaleOrder.Status != constant.MemberSaleOrderStatusPendingPayment {
		return true
	}

	// 3. 执行取消操作
	memberSaleOrder.SetCancel(params.Reason)
	if err := repository.NewMemberSaleOrderRepo(db).UpdateMemberSaleOrder(*memberSaleOrder); err != nil {
		fmt.Println("更新会员订单状态失败", zap.Uint64("memberSaleOrderUuid", params.MemberSaleOrderUuid), zap.Error(err))
		logger.Logger.Error("更新会员订单状态失败", zap.Uint64("memberSaleOrderUuid", params.MemberSaleOrderUuid), zap.Error(err))
		return false
	}

	// 4. 取消订单
	repository.NewOrderRepo(db).CancelOrder(context.NewContext(
		context.WithCompanyUuid(params.CompanyUuid),
		context.WithLogger(logger.Logger),
	), memberSaleOrder.SaleBillUuid, 0, params.Reason)

	// 5. 发送通知

	return true
}
