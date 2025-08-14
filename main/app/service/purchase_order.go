package service

import (
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/repository/base"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/utils"

	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

// IPurchaseOrderSrv 采购申请服务接口
type IPurchaseOrderSrv interface {
	// 采购申请管理
	GetPurchaseOrderList(ctx context.Context, req req.PurchaseOrderListReq) (resp.PurchaseOrderListResp, error)       // 获取采购申请列表
	GetPurchaseOrderDetail(ctx context.Context, req req.PurchaseOrderDetailReq) (resp.PurchaseOrderDetailResp, error) // 获取采购申请详情
	CreatePurchaseOrder(ctx context.Context, req req.PurchaseOrderCreateReq) (resp.PurchaseOrderCreateResp, error)    // 创建采购申请
	UpdatePurchaseOrder(ctx context.Context, req req.PurchaseOrderUpdateReq) (resp.PurchaseOrderUpdateResp, error)    // 更新采购申请
	DeletePurchaseOrder(ctx context.Context, req req.PurchaseOrderDeleteReq) (resp.PurchaseOrderDeleteResp, error)    // 删除采购申请
	SubmitPurchaseOrder(ctx context.Context, req req.PurchaseOrderSubmitReq) error                                    // 提交采购申请
	ApprovePurchaseOrder(ctx context.Context, req req.PurchaseOrderApproveReq) (resp.PurchaseOrderApproveResp, error) // 审核采购申请

	// 收货管理
	CreatePurchaseReceiptOrder(ctx context.Context, req req.PurchaseReceiptCreateReq) (resp.PurchaseReceiptOrderCreateResp, error)
	GetPurchaseReceiptOrderList(ctx context.Context, req req.PurchaseReceiptOrderListReq) (resp.PurchaseReceiptOrderListResp, error)
	GetPurchaseReceiptOrderDetail(ctx context.Context, req req.PurchaseReceiptOrderDetailReq) (resp.PurchaseReceiptOrderDetailResp, error)

	// 统计分析
	GetPurchaseOrderStatistics(ctx context.Context, req req.PurchaseOrderStatisticsReq) (resp.PurchaseOrderStatisticsResp, error)
}

// purchaseOrderSrv 采购申请服务实现
type purchaseOrderSrv struct {
	dbm *database.DBManager // 数据库管理器
}

// NewPurchaseOrderSrv 创建采购申请服务
func NewPurchaseOrderSrv(dbm *database.DBManager) IPurchaseOrderSrv {
	return NewPurchaseOrderSrvImpl(dbm)
}

// NewPurchaseOrderSrvImpl 创建采购申请服务实现
func NewPurchaseOrderSrvImpl(dbm *database.DBManager) IPurchaseOrderSrv {
	return &purchaseOrderSrv{
		dbm: dbm,
	}
}

// GetPurchaseOrderList 获取采购申请列表
func (s *purchaseOrderSrv) GetPurchaseOrderList(ctx context.Context, req req.PurchaseOrderListReq) (resp.PurchaseOrderListResp, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	purchaseOrderRepo := repository.NewPurchaseOrderRepo(db)

	// 构建查询选项
	var opts []repository.DBOption
	if req.OrderNo != "" {
		opts = append(opts, purchaseOrderRepo.WhereOrderNo(req.OrderNo))
	}
	if req.Status != nil && *req.Status != -1 {
		opts = append(opts, purchaseOrderRepo.WhereStatus(*req.Status))
	}

	// 排序
	opts = append(opts, purchaseOrderRepo.OrderByCreateTime(true))

	// 查询数据
	purchaseOrders, total, err := purchaseOrderRepo.GetListWithPagination(req.PageNo, req.PageSize, opts...)
	if err != nil {
		return resp.PurchaseOrderListResp{}, errors.WithMessage(err, "查询采购申请列表失败")
	}

	// 转换响应数据
	listResp := make([]*resp.PurchaseOrderInfo, 0)
	for _, po := range purchaseOrders {
		poInfo := resp.PurchaseOrderInfo{}
		err := copier.Copy(&poInfo, &po)
		if err != nil {
			continue
		}
		// 计算明细统计信息
		listResp = append(listResp, &poInfo)
	}

	return resp.PurchaseOrderListResp{
		List: listResp,
		Meta: dto.PageResponse{
			PageNo:   req.PageNo,
			PageSize: req.PageSize,
			Total:    total,
		},
	}, nil
}

// GetPurchaseOrderDetail 获取采购申请详情
func (s *purchaseOrderSrv) GetPurchaseOrderDetail(ctx context.Context, req req.PurchaseOrderDetailReq) (resp.PurchaseOrderDetailResp, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	purchaseOrderRepo := repository.NewPurchaseOrderRepo(db)

	// 查询采购申请详情
	purchaseOrder, err := purchaseOrderRepo.GetByUuid(req.Uuid, purchaseOrderRepo.WithItems())
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return resp.PurchaseOrderDetailResp{}, errors.New("采购申请不存在")
		}
		return resp.PurchaseOrderDetailResp{}, errors.WithMessage(err, "查询采购申请详情失败")
	}

	// 转换响应数据
	var detailResp resp.PurchaseOrderDetailResp
	err = copier.Copy(&detailResp, purchaseOrder)
	if err != nil {
		return resp.PurchaseOrderDetailResp{}, errors.WithMessage(err, "数据转换失败")
	}

	// 初始化数组字段，确保返回空数组而不是null
	detailResp.Items = make([]resp.PurchaseOrderItemInfo, 0)

	// 转换明细数据
	for _, item := range purchaseOrder.Items {
		itemInfo := resp.PurchaseOrderItemInfo{}
		copier.Copy(&itemInfo, &item)
		detailResp.Items = append(detailResp.Items, itemInfo)
	}

	return detailResp, nil
}

// CreatePurchaseOrder 创建采购申请
func (s *purchaseOrderSrv) CreatePurchaseOrder(ctx context.Context, req req.PurchaseOrderCreateReq) (resp.PurchaseOrderCreateResp, error) {
	db := s.dbm.GetDB(ctx.GetDbId())

	var result resp.PurchaseOrderCreateResp

	err := db.Transaction(func(tx *gorm.DB) error {
		purchaseOrderRepo := repository.NewPurchaseOrderRepo(tx)
		purchaseOrderItemRepo := repository.NewPurchaseOrderItemRepo(tx)
		// 创建采购申请
		purchaseOrder := &model.PurchaseOrder{
			OrderNo:           s.generateOrderNo(db),
			OrderType:         req.OrderType,
			Status:            constant.PurchaseOrderStatusDraft, // 待提交状态
			Num:               float64(len(req.Items)),
			OrderTime:         time.Now().Unix(),
			ExpectArrivalTime: req.ExpectedDeliveryTime,
			ApplicantUuid:     ctx.GetStaffUuid(),
			ApplicantName:     ctx.GetStaff().RealName,
		}
		err := purchaseOrderRepo.Create(purchaseOrder)
		if err != nil {
			return errors.WithMessage(err, "创建采购申请失败")
		}

		materialUuids := make([]uint64, 0)
		for _, itemReq := range req.Items {
			materialUuids = append(materialUuids, itemReq.MaterialUuid)
		}

		materialRepo := base.NewMaterialRepo(db)
		materials, err := materialRepo.GetMaterialByUuids(
			materialUuids,
			materialRepo.WithPreload("Unit"),
			materialRepo.WithPreload("PurchaseUnit"),
		)
		if err != nil {
			return errors.WithMessage(err, "查询物品失败")
		}
		if len(materials) != len(req.Items) {
			return errors.New("查询物品失败-数量不一致")
		}

		// 创建采购申请明细
		var items []model.PurchaseOrderItem
		for _, itemReq := range req.Items {
			material := materials[itemReq.MaterialUuid]
			item := model.PurchaseOrderItem{
				PurchaseOrderUuid:  purchaseOrder.Uuid,
				MaterialUuid:       itemReq.MaterialUuid,
				Num:                itemReq.Num,
				MaterialCode:       material.Code,
				MaterialName:       material.Name,
				UnitUuid:           material.PurchaseUnitUuid,
				UnitName:           material.PurchaseUnit.Name,
				UnitConversionRate: material.PurchaseUnit.ConversionRate,
				BaseUnitUuid:       material.Unit.UnitUuid,
				BaseUnitName:       material.Unit.Name,
			}
			items = append(items, item)
		}

		err = purchaseOrderItemRepo.CreateBatch(items)
		if err != nil {
			return errors.WithMessage(err, "创建采购申请明细失败")
		}

		// 记录操作日志
		err = s.createPurchaseOrderLog(tx, purchaseOrder.Uuid, ctx, "create", "创建采购申请", 0, constant.PurchaseOrderStatusPending, "")
		if err != nil {
			return err
		}

		result.Uuid = purchaseOrder.Uuid
		result.OrderNo = purchaseOrder.OrderNo

		return nil
	})

	if err != nil {
		return resp.PurchaseOrderCreateResp{}, err
	}

	return result, nil
}

// UpdatePurchaseOrder 更新采购申请
func (s *purchaseOrderSrv) UpdatePurchaseOrder(ctx context.Context, req req.PurchaseOrderUpdateReq) (resp.PurchaseOrderUpdateResp, error) {
	db := s.dbm.GetDB(ctx.GetDbId())

	err := db.Transaction(func(tx *gorm.DB) error {
		purchaseOrderRepo := repository.NewPurchaseOrderRepo(tx)
		purchaseOrderItemRepo := repository.NewPurchaseOrderItemRepo(tx)

		// 查询现有采购申请
		purchaseOrder, err := purchaseOrderRepo.GetByUuid(req.Uuid)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return errors.New("采购申请不存在")
			}
			return errors.WithMessage(err, "查询采购申请失败")
		}

		// 检查是否可编辑
		if !purchaseOrder.IsEditable() {
			return errors.New("当前状态不允许编辑")
		}

		// 更新采购申请基本信息
		purchaseOrder.Num = float64(len(req.Items))
		err = purchaseOrderRepo.Update(purchaseOrder)
		if err != nil {
			return errors.WithMessage(err, "更新采购申请失败")
		}

		// 先删除所有现有明细项
		err = purchaseOrderItemRepo.DeleteByPurchaseOrderUuid(req.Uuid)
		if err != nil {
			return errors.WithMessage(err, "删除采购申请明细失败")
		}

		materialUuids := make([]uint64, 0)
		for _, itemReq := range req.Items {
			materialUuids = append(materialUuids, itemReq.MaterialUuid)
		}
		materialRepo := base.NewMaterialRepo(db)
		materials, err := materialRepo.GetMaterialByUuids(
			materialUuids,
			materialRepo.WithPreload("Unit"),
			materialRepo.WithPreload("PurchaseUnit"),
		)
		if err != nil {
			return errors.WithMessage(err, "查询物品失败")
		}
		if len(materials) != len(req.Items) {
			return errors.New("查询物品失败-数量不一致")
		}

		// 创建采购申请明细
		var items []model.PurchaseOrderItem
		for _, itemReq := range req.Items {
			material := materials[itemReq.MaterialUuid]
			item := model.PurchaseOrderItem{
				PurchaseOrderUuid:  purchaseOrder.Uuid,
				MaterialUuid:       itemReq.MaterialUuid,
				Num:                itemReq.Num,
				MaterialCode:       material.Code,
				MaterialName:       material.Name,
				UnitUuid:           material.PurchaseUnitUuid,
				UnitName:           material.PurchaseUnit.Name,
				UnitConversionRate: material.PurchaseUnit.ConversionRate,
				BaseUnitUuid:       material.Unit.UnitUuid,
				BaseUnitName:       material.Unit.Name,
			}
			items = append(items, item)
		}

		err = purchaseOrderItemRepo.CreateBatch(items)
		if err != nil {
			return errors.WithMessage(err, "创建采购申请明细失败")
		}

		// 记录操作日志
		err = s.createPurchaseOrderLog(tx, req.Uuid, ctx, "update", "更新采购申请", purchaseOrder.Status, purchaseOrder.Status, "")
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return resp.PurchaseOrderUpdateResp{}, err
	}

	return resp.PurchaseOrderUpdateResp{Success: true}, nil
}

// DeletePurchaseOrder 删除采购申请
func (s *purchaseOrderSrv) DeletePurchaseOrder(ctx context.Context, req req.PurchaseOrderDeleteReq) (resp.PurchaseOrderDeleteResp, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	purchaseOrderRepo := repository.NewPurchaseOrderRepo(db)

	// 查询采购申请
	purchaseOrder, err := purchaseOrderRepo.GetByUuid(req.Uuid)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return resp.PurchaseOrderDeleteResp{}, errors.New("采购申请不存在")
		}
		return resp.PurchaseOrderDeleteResp{}, errors.WithMessage(err, "查询采购申请失败")
	}

	// 检查是否可删除（只有待提交和已驳回状态可以删除）
	if !purchaseOrder.IsEditable() {
		return resp.PurchaseOrderDeleteResp{}, errors.New("当前状态不允许删除")
	}

	// 删除采购申请（软删除）
	err = purchaseOrderRepo.Delete(req.Uuid)
	if err != nil {
		return resp.PurchaseOrderDeleteResp{}, errors.WithMessage(err, "删除采购申请失败")
	}

	return resp.PurchaseOrderDeleteResp{Success: true}, nil
}

// SubmitPurchaseOrder 提交采购申请
func (s *purchaseOrderSrv) SubmitPurchaseOrder(ctx context.Context, req req.PurchaseOrderSubmitReq) error {
	return s.dbm.GetDB(ctx.GetDbId()).Transaction(func(tx *gorm.DB) error {
		purchaseOrderRepo := repository.NewPurchaseOrderRepo(tx)

		// 查询采购申请
		purchaseOrder, err := purchaseOrderRepo.GetByUuid(req.Uuid)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return errors.New("采购申请不存在")
			}
			return errors.WithMessage(err, "查询采购申请失败")
		}

		oldStatus := purchaseOrder.Status
		purchaseOrder.Status = constant.PurchaseOrderStatusPending
		purchaseOrder.OrderTime = time.Now().Unix()
		purchaseOrder.ApplicantUuid = ctx.GetStaffUuid()
		purchaseOrder.ApplicantName = ctx.GetStaff().RealName

		err = purchaseOrderRepo.Update(purchaseOrder)
		if err != nil {
			return errors.WithMessage(err, "更新采购申请状态失败")
		}

		// 记录操作日志
		statusText := purchaseOrder.GetStatusText()
		err = s.createPurchaseOrderLog(tx, req.Uuid, ctx, "status_update", "更新状态为"+statusText, oldStatus, purchaseOrder.Status, "")
		if err != nil {
			return err
		}

		return nil
	})
}

// ApprovePurchaseOrder 审核采购申请
func (s *purchaseOrderSrv) ApprovePurchaseOrder(ctx context.Context, req req.PurchaseOrderApproveReq) (resp.PurchaseOrderApproveResp, error) {
	db := s.dbm.GetDB(ctx.GetDbId())

	err := db.Transaction(func(tx *gorm.DB) error {
		purchaseOrderRepo := repository.NewPurchaseOrderRepo(tx)

		// 查询采购申请
		purchaseOrder, err := purchaseOrderRepo.GetByUuid(req.Uuid)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return errors.New("采购申请不存在")
			}
			return errors.WithMessage(err, "查询采购申请失败")
		}

		// 检查是否可审核
		if !purchaseOrder.CanApprove() {
			return errors.New("当前状态不允许审核")
		}

		oldStatus := purchaseOrder.Status
		var newStatus int
		var actionDesc string

		if req.Action == "approve" {
			newStatus = constant.PurchaseOrderStatusApproved
			actionDesc = "审核通过"
			purchaseOrder.PassTime = time.Now().Unix()
		} else if req.Action == "reject" {
			newStatus = constant.PurchaseOrderStatusRejected
			actionDesc = "审核驳回"
			purchaseOrder.RejectTime = time.Now().Unix()
		} else {
			return errors.New("无效的审核动作")
		}

		purchaseOrder.Status = newStatus

		err = purchaseOrderRepo.Update(purchaseOrder)
		if err != nil {
			return errors.WithMessage(err, "更新采购申请状态失败")
		}

		// 记录操作日志
		err = s.createPurchaseOrderLog(tx, req.Uuid, ctx, req.Action, actionDesc, oldStatus, newStatus, "")
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return resp.PurchaseOrderApproveResp{}, err
	}

	return resp.PurchaseOrderApproveResp{Success: true}, nil
}

// CreateReceiptOrder 创建收货单
func (s *purchaseOrderSrv) CreatePurchaseReceiptOrder(ctx context.Context, req req.PurchaseReceiptCreateReq) (resp.PurchaseReceiptOrderCreateResp, error) {
	db := s.dbm.GetDB(ctx.GetDbId())

	var result resp.PurchaseReceiptOrderCreateResp

	err := db.Transaction(func(tx *gorm.DB) error {
		purchaseOrderRepo := repository.NewPurchaseOrderRepo(tx)
		purchaseOrderItemRepo := repository.NewPurchaseOrderItemRepo(tx)
		receiptOrderRepo := repository.NewPurchaseReceiptOrderRepo(tx)
		receiptOrderItemRepo := repository.NewPurchaseReceiptOrderItemRepo(tx)

		// 查询采购申请
		purchaseOrder, err := purchaseOrderRepo.GetByUuid(req.PurchaseOrderUuid)
		if err != nil {
			return errors.WithMessage(err, "查询采购申请失败")
		}

		// 计算收货总数量
		var totalQuantity float64
		for _, item := range req.Items {
			totalQuantity += item.ReceivedNum
		}

		// 创建收货单
		receiptOrder := &model.PurchaseReceiptOrder{
			OrderNo:           s.generateReceiptNo(tx),
			Status:            constant.ReceiptOrderStatusPending, // 待收货状态
			PurchaseOrderUuid: req.PurchaseOrderUuid,
			PurchaseOrderNo:   purchaseOrder.OrderNo,
			Num:               totalQuantity,
			ExpectArrivalTime: purchaseOrder.ExpectArrivalTime,
			ReceiveTime:       time.Now().Unix(),
		}

		err = receiptOrderRepo.Create(receiptOrder)
		if err != nil {
			return errors.WithMessage(err, "创建收货单失败")
		}

		// 创建收货明细并更新采购申请明细的到货数量
		for _, itemReq := range req.Items {
			// 查询采购申请明细
			orderItem, err := purchaseOrderItemRepo.GetByUuid(itemReq.PurchaseOrderItemUuid)
			if err != nil {
				return errors.WithMessage(err, "查询采购申请明细失败")
			}

			// 创建收货明细
			receiptItem := &model.PurchaseReceiptOrderItem{
				ReceiptOrderUuid:      receiptOrder.Uuid,
				PurchaseOrderItemUuid: itemReq.PurchaseOrderItemUuid,
				MaterialCode:          orderItem.MaterialCode,
				MaterialName:          orderItem.MaterialName,
				MaterialUuid:          orderItem.MaterialUuid,
				Num:                   itemReq.ReceivedNum,
				UnitUuid:              orderItem.UnitUuid,
				UnitName:              orderItem.UnitName,
				BaseUnitUuid:          orderItem.BaseUnitUuid,
				BaseUnitName:          orderItem.BaseUnitName,
				UnitConversionRate:    orderItem.UnitConversionRate,
			}

			err = receiptOrderItemRepo.Create(receiptItem)
			if err != nil {
				return errors.WithMessage(err, "创建收货明细失败")
			}

			// 更新采购申请明细的到货数量
			orderItem.ArrivalNum += itemReq.ReceivedNum
			err = purchaseOrderItemRepo.Update(orderItem)
			if err != nil {
				return errors.WithMessage(err, "更新采购申请明细失败")
			}
		}

		// 检查采购申请是否完成
		err = s.checkAndUpdatePurchaseOrderStatus(tx, req.PurchaseOrderUuid)
		if err != nil {
			return err
		}

		result.Uuid = receiptOrder.Uuid

		return nil
	})

	if err != nil {
		return resp.PurchaseReceiptOrderCreateResp{}, err
	}

	return result, nil
}

// GetReceiptOrderList 获取收货单列表
func (s *purchaseOrderSrv) GetPurchaseReceiptOrderList(ctx context.Context, req req.PurchaseReceiptOrderListReq) (resp.PurchaseReceiptOrderListResp, error) {
	// 设置默认分页参数
	if req.PageNo <= 0 {
		req.PageNo = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	db := s.dbm.GetDB(ctx.GetDbId())
	receiptOrderRepo := repository.NewPurchaseReceiptOrderRepo(db)

	// 构建查询选项
	var opts []repository.DBOption
	if req.PurchaseOrderUuid > 0 {
		opts = append(opts, receiptOrderRepo.WherePurchaseOrderUuid(req.PurchaseOrderUuid))
	}
	if req.ReceiptNo != "" {
		opts = append(opts, receiptOrderRepo.WhereReceiptNo(req.ReceiptNo))
	}
	if req.ReceiptTimeStart > 0 || req.ReceiptTimeEnd > 0 {
		opts = append(opts, receiptOrderRepo.WhereReceiptTimeRange(req.ReceiptTimeStart, req.ReceiptTimeEnd))
	}

	// 排序
	opts = append(opts, receiptOrderRepo.OrderByReceiptTime(true))

	// 查询数据
	receipts, total, err := receiptOrderRepo.GetListWithPagination(req.PageNo, req.PageSize, opts...)
	if err != nil {
		return resp.PurchaseReceiptOrderListResp{}, errors.WithMessage(err, "查询收货单列表失败")
	}

	// 转换响应数据
	var listResp []resp.PurchaseReceiptInfo
	for _, receipt := range receipts {
		receiptInfo := resp.PurchaseReceiptInfo{}
		err := copier.Copy(&receiptInfo, &receipt)
		if err != nil {
			continue
		}

		// receiptInfo.StatusText = receipt.GetStatusText()
		listResp = append(listResp, receiptInfo)
	}

	return resp.PurchaseReceiptOrderListResp{
		List: listResp,
		Meta: dto.PageResponse{
			PageNo:   req.PageNo,
			PageSize: req.PageSize,
			Total:    total,
		},
	}, nil
}

// GetReceiptOrderDetail 获取收货单详情
func (s *purchaseOrderSrv) GetPurchaseReceiptOrderDetail(ctx context.Context, req req.PurchaseReceiptOrderDetailReq) (resp.PurchaseReceiptOrderDetailResp, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	receiptOrderRepo := repository.NewPurchaseReceiptOrderRepo(db)

	// 查询收货单详情
	receipt, err := receiptOrderRepo.GetByUuid(req.Uuid, receiptOrderRepo.WithItems())
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return resp.PurchaseReceiptOrderDetailResp{}, errors.New("收货单不存在")
		}
		return resp.PurchaseReceiptOrderDetailResp{}, errors.WithMessage(err, "查询收货单详情失败")
	}

	// 转换响应数据
	var detailResp resp.PurchaseReceiptOrderDetailResp
	err = copier.Copy(&detailResp, receipt)
	if err != nil {
		return resp.PurchaseReceiptOrderDetailResp{}, errors.WithMessage(err, "数据转换失败")
	}

	detailResp.StatusText = receipt.GetStatusText()

	// 转换收货明细数据
	for _, item := range receipt.Items {
		itemInfo := resp.PurchaseReceiptItemInfo{}
		copier.Copy(&itemInfo, &item)
		detailResp.Items = append(detailResp.Items, itemInfo)
	}

	return detailResp, nil
}

// GetPurchaseOrderStatistics 获取采购申请统计
func (s *purchaseOrderSrv) GetPurchaseOrderStatistics(ctx context.Context, req req.PurchaseOrderStatisticsReq) (resp.PurchaseOrderStatisticsResp, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	purchaseOrderRepo := repository.NewPurchaseOrderRepo(db)

	// 构建查询选项
	var opts []repository.DBOption
	if req.TimeStart > 0 || req.TimeEnd > 0 {
		opts = append(opts, purchaseOrderRepo.WhereCreateTimeRange(req.TimeStart, req.TimeEnd))
	}

	// 状态统计
	statusStats, err := purchaseOrderRepo.GetStatusStats(opts...)
	if err != nil {
		return resp.PurchaseOrderStatisticsResp{}, errors.WithMessage(err, "查询状态统计失败")
	}

	// 订单类型统计
	// orderTypeStats, err := purchaseOrderRepo.GetOrderTypeStats(opts...)
	// if err != nil {
	// 	return resp.PurchaseOrderStatisticsResp{}, errors.WithMessage(err, "查询订单类型统计失败")
	// }

	// 转换响应数据
	var statsResp resp.PurchaseOrderStatisticsResp

	// 状态统计
	for status, count := range statusStats {
		po := &model.PurchaseOrder{Status: status}
		statsResp.StatusStats = append(statsResp.StatusStats, resp.StatusStatItem{
			Status:     status,
			StatusText: po.GetStatusText(),
			Count:      int(count),
		})
	}

	// 订单类型统计
	// for orderType, count := range orderTypeStats {
	// 	statsResp.OrderTypeStats = append(statsResp.OrderTypeStats, resp.OrderTypeStatItem{
	// 		OrderType: orderType,
	// 		Count:     int(count),
	// 	})
	// }

	return statsResp, nil
}

// generateOrderNo 生成订单编号
func (s *purchaseOrderSrv) generateOrderNo(db *gorm.DB) string {
	var orderNo string
	// 前八位是年月日
	datePart := time.Now().Format("20060102")
	// 第九位是订单来源
	// 如果订单编号存在, 则重新生成, 重试10次, 否则退出
	for i := 0; i < 10; i++ {
		// 后九位是随机生成
		n := utils.RandomNumber(9)
		// 订单编号
		orderNo = datePart + n
		// 检查订单编号是否存在
		noExist, err := repository.NewPurchaseOrderRepo(db).IsOrderNoExists(orderNo)
		if err != nil {
			return ""
		}
		// 如果订单编号存在，则重新生成
		if !noExist {
			orderNo = ""
			continue
		}
		// 如果订单编号不存在，则退出，本次生成的订单编号可用
		break
	}
	if orderNo == "" {
		return ""
	}
	return orderNo
}

// generateReceiptNo 生成收货单号
func (s *purchaseOrderSrv) generateReceiptNo(db *gorm.DB) string {
	var receiptNo string
	// 前八位是年月日
	datePart := time.Now().Format("20060102")
	// 第九位是订单来源
	// 如果订单编号存在, 则重新生成, 重试10次, 否则退出
	for i := 0; i < 10; i++ {
		// 后九位是随机生成
		n := utils.RandomNumber(9)
		// 订单编号
		receiptNo = datePart + n
		// 检查订单编号是否存在
		noExist, err := repository.NewPurchaseReceiptOrderRepo(db).IsOrderNoExists(receiptNo)
		if err != nil {
			return ""
		}
		// 如果订单编号存在，则重新生成
		if !noExist {
			receiptNo = ""
			continue
		}
		// 如果订单编号不存在，则退出，本次生成的订单编号可用
		break
	}
	if receiptNo == "" {
		return ""
	}
	return receiptNo
}

// createPurchaseOrderLog 创建采购订单操作日志
func (s *purchaseOrderSrv) createPurchaseOrderLog(db *gorm.DB, purchaseOrderUuid uint64, ctx context.Context, action, actionDesc string, oldStatus, newStatus int, remark string) error {
	logRepo := repository.NewPurchaseOrderLogRepo(db)

	log := &model.PurchaseOrderLog{
		PurchaseOrderUuid: purchaseOrderUuid,
		OperatorUuid:      ctx.GetStaffUuid(),
		OperatorName:      ctx.GetStaff().RealName,
		Action:            action,
		ActionDesc:        actionDesc,
		OldStatus:         oldStatus,
		NewStatus:         newStatus,
		Remark:            remark,
	}

	return logRepo.Create(log)
}

// checkAndUpdatePurchaseOrderStatus 检查并更新采购申请状态
func (s *purchaseOrderSrv) checkAndUpdatePurchaseOrderStatus(db *gorm.DB, purchaseOrderUuid uint64) error {
	purchaseOrderRepo := repository.NewPurchaseOrderRepo(db)
	purchaseOrderItemRepo := repository.NewPurchaseOrderItemRepo(db)

	// 获取所有明细
	items, err := purchaseOrderItemRepo.GetByPurchaseOrderUuid(purchaseOrderUuid)
	if err != nil {
		return err
	}

	// 检查是否全部到货完成
	allCompleted := true
	partialReceived := false

	for _, item := range items {
		if item.ArrivalNum < item.Num {
			allCompleted = false
		}
		if item.ArrivalNum > 0 {
			partialReceived = true
		}
	}

	// 更新采购申请状态
	purchaseOrder, err := purchaseOrderRepo.GetByUuid(purchaseOrderUuid)
	if err != nil {
		return err
	}

	oldStatus := purchaseOrder.Status

	if allCompleted {
		purchaseOrder.Status = constant.PurchaseOrderStatusCompleted
		purchaseOrder.FinalReceiveTime = time.Now().Unix()
	} else if partialReceived && purchaseOrder.Status != constant.PurchaseOrderStatusPartialReceived {
		purchaseOrder.Status = constant.PurchaseOrderStatusPartialReceived
		if purchaseOrder.FirstReceiveTime == 0 {
			purchaseOrder.FirstReceiveTime = time.Now().Unix()
		}
	}

	if oldStatus != purchaseOrder.Status {
		err = purchaseOrderRepo.Update(purchaseOrder)
		if err != nil {
			return err
		}
	}

	return nil
}
