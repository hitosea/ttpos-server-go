package transfer_order

import (
	"fmt"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/language"
	"ttpos-server-go/pkg/lock"
	"ttpos-server-go/pkg/logger"

	"github.com/jinzhu/copier"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ITransferOrderSrv 调拨单服务接口
type ITransferOrderSrv interface {
	// 调拨单管理
	GetTransferOrderList(ctx context.Context, req req.TransferOrderListReq) (resp.TransferOrderListResp, error)
	GetTransferOrderDetail(ctx context.Context, req req.TransferOrderDetailReq) (resp.TransferOrderDetailResp, error)
	CreateTransferOrder(ctx context.Context, req req.TransferOrderCreateReq) (resp.TransferOrderCreateResp, error)
	UpdateTransferOrder(ctx context.Context, req req.TransferOrderUpdateReq) error
	DeleteTransferOrder(ctx context.Context, req req.TransferOrderDeleteReq) error
	SubmitTransferOrder(ctx context.Context, req req.TransferOrderSubmitReq) error
	ApproveTransferOrder(ctx context.Context, req req.TransferOrderApproveReq) error
	RejectTransferOrder(ctx context.Context, req req.TransferOrderRejectReq) error
	ReceiveTransferOrder(ctx context.Context, req req.TransferOrderReceiveReq) error

	// 审批流程和日志
	GetTransferOrderApprovalList(ctx context.Context, req req.TransferOrderApprovalListReq) (resp.TransferOrderApprovalListResp, error)
	GetTransferOrderLogList(ctx context.Context, req req.TransferOrderLogListReq) (resp.TransferOrderLogListResp, error)
}

// transferOrderSrv 调拨单服务实现
type transferOrderSrv struct {
	dbm  *database.DBManager
	lock lock.Lock
}

// NewTransferOrderSrv 创建调拨单服务
func NewTransferOrderSrv(dbm *database.DBManager) ITransferOrderSrv {
	return NewTransferOrderSrvImpl(dbm)
}

// NewTransferOrderSrvImpl 创建调拨单服务实现
func NewTransferOrderSrvImpl(dbm *database.DBManager) ITransferOrderSrv {
	return &transferOrderSrv{
		dbm:  dbm,
		lock: lock.NewSystemLock(),
	}
}

// GetTransferOrderList 获取调拨单列表
func (s *transferOrderSrv) GetTransferOrderList(
	ctx context.Context,
	req req.TransferOrderListReq,
) (resp.TransferOrderListResp, error) {
	db := ctx.GetDB()
	transferOrderRepo := repository.NewTransferOrderRepo(db)

	// 构建查询选项
	var opts []repository.DBOption
	if req.OrderNo != "" {
		opts = append(opts, transferOrderRepo.WhereOrderNo(req.OrderNo))
	}
	if len(req.StatusIn) > 0 {
		opts = append(opts, transferOrderRepo.WhereStatusIn(req.StatusIn))
	}
	if req.TransferType > 0 {
		opts = append(opts, transferOrderRepo.WhereTransferType(req.TransferType))
	}
	if req.SenderCompanyUuid > 0 {
		opts = append(opts, transferOrderRepo.WhereSenderCompanyUuid(req.SenderCompanyUuid))
	}
	if req.ReceiverCompanyUuid > 0 {
		opts = append(opts, transferOrderRepo.WhereReceiverCompanyUuid(req.ReceiverCompanyUuid))
	}
	if req.OutWarehouseErpCode != "" {
		opts = append(opts, transferOrderRepo.WhereOutWarehouseErpCode(req.OutWarehouseErpCode))
	}
	if req.InWarehouseErpCode != "" {
		opts = append(opts, transferOrderRepo.WhereInWarehouseErpCode(req.InWarehouseErpCode))
	}
	if req.CreateTimeStart > 0 || req.CreateTimeEnd > 0 {
		opts = append(opts, transferOrderRepo.WhereCreateTimeRange(req.CreateTimeStart, req.CreateTimeEnd))
	}
	if req.OrderTimeStart > 0 || req.OrderTimeEnd > 0 {
		opts = append(opts, transferOrderRepo.WhereOrderTimeRange(req.OrderTimeStart, req.OrderTimeEnd))
	}
	if req.SubmitTimeStart > 0 || req.SubmitTimeEnd > 0 {
		opts = append(opts, transferOrderRepo.WhereSubmitTimeRange(req.SubmitTimeStart, req.SubmitTimeEnd))
	}
	if len(req.UuidIn) > 0 {
		opts = append(opts, transferOrderRepo.WhereUuidIn(req.UuidIn))
	}

	// 排序
	opts = append(opts, transferOrderRepo.OrderByCreateTime(true))

	// 查询数据
	transferOrders, total, err := transferOrderRepo.GetListWithPagination(req.PageNo, req.PageSize, opts...)
	if err != nil {
		return resp.TransferOrderListResp{}, errors.WithMessage(errors.New("查询调拨单列表失败"), err.Error())
	}

	// 转换响应数据
	listResp := make([]*resp.TransferOrderInfo, 0, len(transferOrders))
	for _, to := range transferOrders {
		toInfo := &resp.TransferOrderInfo{}
		if err := copier.Copy(toInfo, &to); err != nil {
			continue
		}
		// 转换仓库名称
		toInfo.OutWarehouseName = *language.JsonToLocaleResponse(to.OutWarehouseName)
		toInfo.InWarehouseName = *language.JsonToLocaleResponse(to.InWarehouseName)
		listResp = append(listResp, toInfo)
	}

	return resp.TransferOrderListResp{
		List: listResp,
		Meta: dto.PageResponse{
			PageNo:   req.PageNo,
			PageSize: req.PageSize,
			Total:    total,
		},
	}, nil
}

// GetTransferOrderDetail 获取调拨单详情
func (s *transferOrderSrv) GetTransferOrderDetail(
	ctx context.Context,
	req req.TransferOrderDetailReq,
) (resp.TransferOrderDetailResp, error) {
	db := ctx.GetDB()
	transferOrderRepo := repository.NewTransferOrderRepo(db)

	// 查询调拨单详情
	transferOrder, err := transferOrderRepo.GetByUuid(
		req.Uuid,
		transferOrderRepo.WithItems(),
		transferOrderRepo.WithApprovals(),
	)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return resp.TransferOrderDetailResp{}, errors.New("调拨单不存在")
		}
		return resp.TransferOrderDetailResp{}, errors.WithMessage(errors.New("查询调拨单详情失败"), err.Error())
	}

	// 转换响应数据
	var detailResp resp.TransferOrderDetailResp
	if err = copier.Copy(&detailResp, transferOrder); err != nil {
		return resp.TransferOrderDetailResp{}, errors.WithMessage(errors.New("数据转换失败"), err.Error())
	}

	// 转换仓库名称
	detailResp.OutWarehouseName = *language.JsonToLocaleResponse(transferOrder.OutWarehouseName)
	detailResp.InWarehouseName = *language.JsonToLocaleResponse(transferOrder.InWarehouseName)

	// 转换明细数据
	detailResp.Items = make([]resp.TransferOrderItemInfo, 0, len(transferOrder.Items))
	for _, item := range transferOrder.Items {
		itemInfo := resp.TransferOrderItemInfo{}
		copier.Copy(&itemInfo, &item)
		itemInfo.MaterialName = *language.JsonToLocaleResponse(item.MaterialName)

		// 转换单位列表
		itemInfo.Units = make([]resp.TransferOrderItemUnitInfo, 0, len(item.Units))
		for _, unit := range item.Units {
			unitInfo := resp.TransferOrderItemUnitInfo{}
			copier.Copy(&unitInfo, &unit)
			unitInfo.UnitName = *language.JsonToLocaleResponse(unit.UnitName)
			itemInfo.Units = append(itemInfo.Units, unitInfo)
		}

		detailResp.Items = append(detailResp.Items, itemInfo)
	}

	// 转换审批流程
	detailResp.Approvals = make([]resp.TransferOrderApprovalInfo, 0, len(transferOrder.Approvals))
	for _, approval := range transferOrder.Approvals {
		approvalInfo := resp.TransferOrderApprovalInfo{}
		copier.Copy(&approvalInfo, &approval)
		detailResp.Approvals = append(detailResp.Approvals, approvalInfo)
	}

	return detailResp, nil
}

// CreateTransferOrder 创建调拨单
func (s *transferOrderSrv) CreateTransferOrder(
	ctx context.Context,
	req req.TransferOrderCreateReq,
) (resp.TransferOrderCreateResp, error) {
	// 验证请求参数
	if err := req.Validate(); err != nil {
		return resp.TransferOrderCreateResp{}, err
	}

	db := ctx.GetDB()
	transferOrderRepo := repository.NewTransferOrderRepo(db)

	// 加锁防止并发创建（使用字符串锁保护编号生成）
	lockKey := fmt.Sprintf("transfer_order_create_%d", ctx.GetCompanyUuid())
	s.lock.LockUuidString(lockKey)
	defer s.lock.UnlockUuidString(lockKey)

	// 生成调拨单编号
	orderNo, err := s.generateOrderNo(db)
	if err != nil {
		return resp.TransferOrderCreateResp{}, errors.WithMessage(errors.New("生成调拨单编号失败"), err.Error())
	}

	// 开始事务
	err = db.Transaction(func(tx *gorm.DB) error {
		// 创建调拨单主表
		transferOrder := &model.TransferOrder{
			CompanyUuid:         ctx.GetCompanyUuid(),
			HeadquarterUuid:     0, // 根据业务需要设置
			OrderNo:             orderNo,
			TransferType:        req.TransferType,
			SenderCompanyUuid:   req.SenderCompanyUuid,
			ReceiverCompanyUuid: req.ReceiverCompanyUuid,
			OutWarehouseErpCode: req.OutWarehouseErpCode,
			InWarehouseErpCode:  req.InWarehouseErpCode,
			OrderTime:           int(req.OrderTime),
			Status:              constant.TransferOrderStatusDraft,
			CreatorUuid:         ctx.GetStaffUuid(),
			CreatorName:         ctx.GetStaff().RealName,
			Remark:              req.Remark,
			ItemCount:           len(req.Items),
		}

		transferOrderRepoTx := repository.NewTransferOrderRepo(tx)
		if err := transferOrderRepoTx.Create(transferOrder); err != nil {
			return errors.WithMessage(errors.New("创建调拨单失败"), err.Error())
		}

		// 创建调拨单明细
		for _, itemReq := range req.Items {
			item := &model.TransferOrderItem{
				TransferOrderUuid: transferOrder.Uuid,
				CompanyUuid:       ctx.GetCompanyUuid(),
				HeadquarterUuid:   0,
				MaterialUuid:      itemReq.MaterialUuid,
			}

			transferOrderItemRepoTx := repository.NewTransferOrderItemRepo(tx)
			if err := transferOrderItemRepoTx.Create(item); err != nil {
				return errors.WithMessage(errors.New("创建调拨单明细失败"), err.Error())
			}

			// 创建单位明细
			for _, unitReq := range itemReq.Units {
				unit := &model.TransferOrderItemUnit{
					ItemUuid:          item.Uuid,
					TransferOrderUuid: transferOrder.Uuid,
					MaterialUuid:      itemReq.MaterialUuid,
					UnitUuid:          unitReq.UnitUuid,
					Num:               unitReq.Num,
				}

				transferOrderItemUnitRepoTx := repository.NewTransferOrderItemUnitRepo(tx)
				if err := transferOrderItemUnitRepoTx.Create(unit); err != nil {
					return errors.WithMessage(errors.New("创建调拨单单位明细失败"), err.Error())
				}
			}
		}

		// 记录操作日志
		if err := s.createLog(tx, transferOrder.Uuid, constant.TransferActionCreate, "创建调拨单", 0, constant.TransferOrderStatusDraft, ctx); err != nil {
			logger.Logger.Error("记录调拨单日志失败", zap.Error(err))
		}

		return nil
	})

	if err != nil {
		return resp.TransferOrderCreateResp{}, err
	}

	// 查询创建的调拨单
	transferOrder, err := transferOrderRepo.GetByOrderNo(orderNo)
	if err != nil {
		return resp.TransferOrderCreateResp{}, errors.WithMessage(errors.New("查询调拨单失败"), err.Error())
	}

	return resp.TransferOrderCreateResp{
		Uuid:    transferOrder.Uuid,
		OrderNo: transferOrder.OrderNo,
	}, nil
}

// UpdateTransferOrder 更新调拨单（仅待提交状态可更新）
func (s *transferOrderSrv) UpdateTransferOrder(
	ctx context.Context,
	req req.TransferOrderUpdateReq,
) error {
	// 验证请求参数
	if err := req.Validate(); err != nil {
		return err
	}

	db := ctx.GetDB()
	transferOrderRepo := repository.NewTransferOrderRepo(db)

	// 查询调拨单
	transferOrder, err := transferOrderRepo.GetByUuid(req.Uuid)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("调拨单不存在")
		}
		return errors.WithMessage(errors.New("查询调拨单失败"), err.Error())
	}

	// 验证状态
	if transferOrder.Status != constant.TransferOrderStatusDraft {
		return errors.New("只有待提交状态的调拨单才能修改")
	}

	// 加锁
	s.lock.LockUuid(req.Uuid)
	defer s.lock.UnlockUuid(req.Uuid)

	// 开始事务
	err = db.Transaction(func(tx *gorm.DB) error {
		// 更新主表
		updates := make(map[string]interface{})
		if req.OrderTime > 0 {
			updates["order_time"] = req.OrderTime
		}
		if req.SenderCompanyUuid > 0 {
			updates["sender_company_uuid"] = req.SenderCompanyUuid
		}
		if req.ReceiverCompanyUuid > 0 {
			updates["receiver_company_uuid"] = req.ReceiverCompanyUuid
		}
		if req.OutWarehouseErpCode != "" {
			updates["out_warehouse_erp_code"] = req.OutWarehouseErpCode
		}
		if req.InWarehouseErpCode != "" {
			updates["in_warehouse_erp_code"] = req.InWarehouseErpCode
		}
		if req.Remark != "" {
			updates["remark"] = req.Remark
		}

		if len(updates) > 0 {
			if err := tx.Model(&model.TransferOrder{}).Where("uuid = ?", req.Uuid).Updates(updates).Error; err != nil {
				return errors.WithMessage(errors.New("更新调拨单失败"), err.Error())
			}
		}

		// 如果有明细更新，先删除旧明细
		if len(req.Items) > 0 {
			transferOrderItemUnitRepoTx := repository.NewTransferOrderItemUnitRepo(tx)
			if err := transferOrderItemUnitRepoTx.DeleteByTransferOrderUuid(req.Uuid); err != nil {
				return errors.WithMessage(errors.New("删除旧单位明细失败"), err.Error())
			}

			transferOrderItemRepoTx := repository.NewTransferOrderItemRepo(tx)
			if err := transferOrderItemRepoTx.DeleteByTransferOrderUuid(req.Uuid); err != nil {
				return errors.WithMessage(errors.New("删除旧明细失败"), err.Error())
			}

			// 创建新明细
			for _, itemReq := range req.Items {
				item := &model.TransferOrderItem{
					TransferOrderUuid: req.Uuid,
					CompanyUuid:       ctx.GetCompanyUuid(),
					HeadquarterUuid:   0,
					MaterialUuid:      itemReq.MaterialUuid,
				}

				if err := transferOrderItemRepoTx.Create(item); err != nil {
					return errors.WithMessage(errors.New("创建调拨单明细失败"), err.Error())
				}

				// 创建单位明细
				for _, unitReq := range itemReq.Units {
					unit := &model.TransferOrderItemUnit{
						ItemUuid:          item.Uuid,
						TransferOrderUuid: req.Uuid,
						MaterialUuid:      itemReq.MaterialUuid,
						UnitUuid:          unitReq.UnitUuid,
						Num:               unitReq.Num,
					}

					transferOrderItemUnitRepoTx := repository.NewTransferOrderItemUnitRepo(tx)
					if err := transferOrderItemUnitRepoTx.Create(unit); err != nil {
						return errors.WithMessage(errors.New("创建调拨单单位明细失败"), err.Error())
					}
				}
			}

			// 更新物品数量
			if err := tx.Model(&model.TransferOrder{}).Where("uuid = ?", req.Uuid).Update("item_count", len(req.Items)).Error; err != nil {
				return errors.WithMessage(errors.New("更新物品数量失败"), err.Error())
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	// 记录操作日志
	if err := s.createLog(db, req.Uuid, "update", "更新调拨单", transferOrder.Status, transferOrder.Status, ctx); err != nil {
		logger.Logger.Error("记录调拨单日志失败", zap.Error(err))
	}

	return nil
}

// DeleteTransferOrder 删除调拨单（仅待提交状态可删除）
func (s *transferOrderSrv) DeleteTransferOrder(
	ctx context.Context,
	req req.TransferOrderDeleteReq,
) error {
	if err := req.Validate(); err != nil {
		return err
	}

	db := ctx.GetDB()
	transferOrderRepo := repository.NewTransferOrderRepo(db)

	// 查询调拨单
	transferOrder, err := transferOrderRepo.GetByUuid(req.Uuid)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("调拨单不存在")
		}
		return errors.WithMessage(errors.New("查询调拨单失败"), err.Error())
	}

	// 验证状态
	if transferOrder.Status != constant.TransferOrderStatusDraft {
		return errors.New("只有待提交状态的调拨单才能删除")
	}

	// 加锁
	s.lock.LockUuid(req.Uuid)
	defer s.lock.UnlockUuid(req.Uuid)

	// 删除调拨单
	if err := transferOrderRepo.Delete(req.Uuid); err != nil {
		return errors.WithMessage(errors.New("删除调拨单失败"), err.Error())
	}

	// 清理锁资源
	s.lock.ClearUuidLock(req.Uuid)

	return nil
}

// 生成调拨单编号
func (s *transferOrderSrv) generateOrderNo(db *gorm.DB) (string, error) {
	transferOrderRepo := repository.NewTransferOrderRepo(db)

	// 获取今天最新的调拨单
	latestOrder, err := transferOrderRepo.GetLatestOrderToday()
	if err != nil && err != gorm.ErrRecordNotFound {
		return "", err
	}

	// 生成新编号
	now := time.Now()
	dateStr := now.Format("20060102")
	sequence := 1

	if latestOrder != nil && latestOrder.OrderNo != "" {
		// 解析最后的序号
		if len(latestOrder.OrderNo) >= 12 {
			lastSeq := latestOrder.OrderNo[len(latestOrder.OrderNo)-4:]
			if seq, err := fmt.Sscanf(lastSeq, "%d", &sequence); err == nil && seq == 1 {
				sequence++
			}
		}
	}

	orderNo := fmt.Sprintf("TR%s%04d", dateStr, sequence)
	return orderNo, nil
}

// 创建操作日志
func (s *transferOrderSrv) createLog(db *gorm.DB, transferOrderUuid uint64, action, actionDesc string, oldStatus, newStatus int, ctx context.Context) error {
	logRepo := repository.NewTransferOrderLogRepo(db)

	log := &model.TransferOrderLog{
		TransferOrderUuid: transferOrderUuid,
		CompanyUuid:       ctx.GetCompanyUuid(),
		Action:            action,
		ActionDesc:        actionDesc,
		OldStatus:         oldStatus,
		NewStatus:         newStatus,
		OperatorUuid:      ctx.GetStaffUuid(),
		OperatorName:      ctx.GetStaff().RealName,
	}

	return logRepo.Create(log)
}

// SubmitTransferOrder 提交调拨单
func (s *transferOrderSrv) SubmitTransferOrder(
	ctx context.Context,
	req req.TransferOrderSubmitReq,
) error {
	if err := req.Validate(); err != nil {
		return err
	}

	db := ctx.GetDB()
	transferOrderRepo := repository.NewTransferOrderRepo(db)

	// 查询调拨单
	transferOrder, err := transferOrderRepo.GetByUuid(req.Uuid, transferOrderRepo.WithItems())
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("调拨单不存在")
		}
		return errors.WithMessage(errors.New("查询调拨单失败"), err.Error())
	}

	// 验证状态
	if transferOrder.Status != constant.TransferOrderStatusDraft {
		return errors.New("只有待提交状态的调拨单才能提交")
	}

	// 加锁
	s.lock.LockUuid(req.Uuid)
	defer s.lock.UnlockUuid(req.Uuid)

	// 开始事务
	err = db.Transaction(func(tx *gorm.DB) error {
		// TODO: 初始化审批流程（根据业务规则创建审批节点）
		// 这里简化处理，实际需要根据发货门店、收货门店、总部等关系来确定审批流程

		// 更新状态为待审核
		updates := map[string]interface{}{
			"status":      constant.TransferOrderStatusPending,
			"submit_time": time.Now().Unix(),
		}

		if err := tx.Model(&model.TransferOrder{}).Where("uuid = ?", req.Uuid).Updates(updates).Error; err != nil {
			return errors.WithMessage(errors.New("提交调拨单失败"), err.Error())
		}

		// 记录操作日志
		if err := s.createLog(tx, req.Uuid, constant.TransferActionSubmit, "提交调拨单", constant.TransferOrderStatusDraft, constant.TransferOrderStatusPending, ctx); err != nil {
			logger.Logger.Error("记录调拨单日志失败", zap.Error(err))
		}

		return nil
	})

	return err
}

// ApproveTransferOrder 审批通过调拨单
func (s *transferOrderSrv) ApproveTransferOrder(
	ctx context.Context,
	req req.TransferOrderApproveReq,
) error {
	if err := req.Validate(); err != nil {
		return err
	}

	db := ctx.GetDB()
	transferOrderRepo := repository.NewTransferOrderRepo(db)
	approvalRepo := repository.NewTransferOrderApprovalRepo(db)

	// 查询调拨单
	transferOrder, err := transferOrderRepo.GetByUuid(req.Uuid, transferOrderRepo.WithApprovals())
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("调拨单不存在")
		}
		return errors.WithMessage(errors.New("查询调拨单失败"), err.Error())
	}

	// 验证状态
	if transferOrder.Status != constant.TransferOrderStatusPending {
		return errors.New("只有待审核状态的调拨单才能审批")
	}

	// 加锁
	s.lock.LockUuid(req.Uuid)
	defer s.lock.UnlockUuid(req.Uuid)

	// 获取当前审批节点
	currentApproval, err := approvalRepo.GetCurrentApproval(req.Uuid, ctx.GetCompanyUuid())
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("无审批权限")
		}
		return errors.WithMessage(errors.New("查询审批节点失败"), err.Error())
	}

	// 开始事务
	err = db.Transaction(func(tx *gorm.DB) error {
		// 更新当前审批节点为已通过
		approvalRepoTx := repository.NewTransferOrderApprovalRepo(tx)
		currentApproval.Status = constant.TransferApprovalApproved
		currentApproval.ApproverUuid = ctx.GetStaffUuid()
		currentApproval.ApproverName = ctx.GetStaff().RealName
		currentApproval.ApproveTime = int(time.Now().Unix())
		currentApproval.Remark = req.Remark

		if err := approvalRepoTx.Update(currentApproval); err != nil {
			return errors.WithMessage(errors.New("更新审批节点失败"), err.Error())
		}

		// 查找下一个审批节点
		nextApproval, err := approvalRepoTx.GetNextApproval(req.Uuid, currentApproval.Sequence)
		if err != nil && err != gorm.ErrRecordNotFound {
			return errors.WithMessage(errors.New("查询下一审批节点失败"), err.Error())
		}

		// 更新调拨单状态
		updates := make(map[string]interface{})
		newStatus := transferOrder.Status

		if nextApproval != nil {
			// 还有下一级审批，更新下一个审批门店
			updates["next_approval_company_uuid"] = nextApproval.ApprovalCompanyUuid
			updates["next_approval_company_name"] = nextApproval.ApprovalCompanyName
		} else {
			// 所有审批已完成，更新为待收货状态
			updates["status"] = constant.TransferOrderStatusReceiving
			updates["next_approval_company_uuid"] = 0
			updates["next_approval_company_name"] = ""
			newStatus = constant.TransferOrderStatusReceiving
		}

		if err := tx.Model(&model.TransferOrder{}).Where("uuid = ?", req.Uuid).Updates(updates).Error; err != nil {
			return errors.WithMessage(errors.New("更新调拨单状态失败"), err.Error())
		}

		// 记录操作日志
		if err := s.createLog(tx, req.Uuid, constant.TransferActionApprove, fmt.Sprintf("%s审批通过", currentApproval.ApprovalCompanyName), transferOrder.Status, newStatus, ctx); err != nil {
			logger.Logger.Error("记录调拨单日志失败", zap.Error(err))
		}

		return nil
	})

	return err
}

// RejectTransferOrder 驳回调拨单
func (s *transferOrderSrv) RejectTransferOrder(
	ctx context.Context,
	req req.TransferOrderRejectReq,
) error {
	if err := req.Validate(); err != nil {
		return err
	}

	db := ctx.GetDB()
	transferOrderRepo := repository.NewTransferOrderRepo(db)
	approvalRepo := repository.NewTransferOrderApprovalRepo(db)

	// 查询调拨单
	transferOrder, err := transferOrderRepo.GetByUuid(req.Uuid)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("调拨单不存在")
		}
		return errors.WithMessage(errors.New("查询调拨单失败"), err.Error())
	}

	// 验证状态
	if transferOrder.Status != constant.TransferOrderStatusPending {
		return errors.New("只有待审核状态的调拨单才能驳回")
	}

	// 加锁
	s.lock.LockUuid(req.Uuid)
	defer s.lock.UnlockUuid(req.Uuid)

	// 获取当前审批节点
	currentApproval, err := approvalRepo.GetCurrentApproval(req.Uuid, ctx.GetCompanyUuid())
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("无审批权限")
		}
		return errors.WithMessage(errors.New("查询审批节点失败"), err.Error())
	}

	// 开始事务
	err = db.Transaction(func(tx *gorm.DB) error {
		// 更新当前审批节点为已驳回
		approvalRepoTx := repository.NewTransferOrderApprovalRepo(tx)
		currentApproval.Status = constant.TransferApprovalRejected
		currentApproval.ApproverUuid = ctx.GetStaffUuid()
		currentApproval.ApproverName = ctx.GetStaff().RealName
		currentApproval.ApproveTime = int(time.Now().Unix())
		currentApproval.RejectReason = req.RejectReason

		if err := approvalRepoTx.Update(currentApproval); err != nil {
			return errors.WithMessage(errors.New("更新审批节点失败"), err.Error())
		}

		// 更新调拨单为已驳回状态
		updates := map[string]interface{}{
			"status":                     constant.TransferOrderStatusRejected,
			"next_approval_company_uuid": 0,
			"next_approval_company_name": "",
		}

		if err := tx.Model(&model.TransferOrder{}).Where("uuid = ?", req.Uuid).Updates(updates).Error; err != nil {
			return errors.WithMessage(errors.New("更新调拨单状态失败"), err.Error())
		}

		// 记录操作日志
		if err := s.createLog(tx, req.Uuid, constant.TransferActionReject, fmt.Sprintf("%s驳回：%s", currentApproval.ApprovalCompanyName, req.RejectReason), transferOrder.Status, constant.TransferOrderStatusRejected, ctx); err != nil {
			logger.Logger.Error("记录调拨单日志失败", zap.Error(err))
		}

		return nil
	})

	return err
}

// ReceiveTransferOrder 收货调拨单
func (s *transferOrderSrv) ReceiveTransferOrder(
	ctx context.Context,
	req req.TransferOrderReceiveReq,
) error {
	if err := req.Validate(); err != nil {
		return err
	}

	db := ctx.GetDB()
	transferOrderRepo := repository.NewTransferOrderRepo(db)

	// 查询调拨单
	transferOrder, err := transferOrderRepo.GetByUuid(req.Uuid)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("调拨单不存在")
		}
		return errors.WithMessage(errors.New("查询调拨单失败"), err.Error())
	}

	// 验证状态
	if transferOrder.Status != constant.TransferOrderStatusReceiving {
		return errors.New("只有待收货状态的调拨单才能收货")
	}

	// 验证收货权限（只有收货门店可以收货）
	if transferOrder.ReceiverCompanyUuid != ctx.GetCompanyUuid() {
		return errors.New("无收货权限")
	}

	// 加锁
	s.lock.LockUuid(req.Uuid)
	defer s.lock.UnlockUuid(req.Uuid)

	// 开始事务
	err = db.Transaction(func(tx *gorm.DB) error {
		// 更新调拨单为已完成状态
		updates := map[string]interface{}{
			"status": constant.TransferOrderStatusCompleted,
		}

		if err := tx.Model(&model.TransferOrder{}).Where("uuid = ?", req.Uuid).Updates(updates).Error; err != nil {
			return errors.WithMessage(errors.New("更新调拨单状态失败"), err.Error())
		}

		// TODO: 这里需要调用库存服务，处理库存入库逻辑

		// 记录操作日志
		if err := s.createLog(tx, req.Uuid, constant.TransferActionReceive, "收货完成", transferOrder.Status, constant.TransferOrderStatusCompleted, ctx); err != nil {
			logger.Logger.Error("记录调拨单日志失败", zap.Error(err))
		}

		return nil
	})

	return err
}

// GetTransferOrderApprovalList 获取调拨单审批流程列表
func (s *transferOrderSrv) GetTransferOrderApprovalList(
	ctx context.Context,
	req req.TransferOrderApprovalListReq,
) (resp.TransferOrderApprovalListResp, error) {
	db := ctx.GetDB()
	approvalRepo := repository.NewTransferOrderApprovalRepo(db)

	// 查询审批流程列表
	approvals, err := approvalRepo.GetListByTransferOrderUuid(req.TransferOrderUuid)
	if err != nil {
		return resp.TransferOrderApprovalListResp{}, errors.WithMessage(errors.New("查询审批流程失败"), err.Error())
	}

	// 转换响应数据
	listResp := make([]*resp.TransferOrderApprovalInfo, 0, len(approvals))
	for _, approval := range approvals {
		approvalInfo := &resp.TransferOrderApprovalInfo{}
		copier.Copy(approvalInfo, &approval)
		listResp = append(listResp, approvalInfo)
	}

	return resp.TransferOrderApprovalListResp{
		List: listResp,
	}, nil
}

// GetTransferOrderLogList 获取调拨单操作日志列表
func (s *transferOrderSrv) GetTransferOrderLogList(
	ctx context.Context,
	req req.TransferOrderLogListReq,
) (resp.TransferOrderLogListResp, error) {
	db := ctx.GetDB()
	logRepo := repository.NewTransferOrderLogRepo(db)

	// 查询操作日志列表
	logs, total, err := logRepo.GetListByTransferOrderUuid(req.TransferOrderUuid, req.PageNo, req.PageSize)
	if err != nil {
		return resp.TransferOrderLogListResp{}, errors.WithMessage(errors.New("查询操作日志失败"), err.Error())
	}

	// 转换响应数据
	listResp := make([]*resp.TransferOrderLogInfo, 0, len(logs))
	for _, log := range logs {
		logInfo := &resp.TransferOrderLogInfo{}
		copier.Copy(logInfo, &log)
		listResp = append(listResp, logInfo)
	}

	return resp.TransferOrderLogListResp{
		List: listResp,
		Meta: dto.PageResponse{
			PageNo:   req.PageNo,
			PageSize: req.PageSize,
			Total:    total,
		},
	}, nil
}
