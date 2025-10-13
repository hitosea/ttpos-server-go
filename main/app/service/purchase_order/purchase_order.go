package purchase_order

import (
	"fmt"
	"strings"
	"time"
	"ttpos-bmp/app/ttpos-erp/api/buying"
	"ttpos-bmp/app/ttpos-erp/api/stock"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/repository/base"
	"ttpos-server-go/app/service/rpc/erp"
	"ttpos-server-go/i18n"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/language"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"github.com/jinzhu/copier"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// IPurchaseOrderSrv 采购申请服务接口
type IPurchaseOrderSrv interface {
	// 采购申请管理
	GetPurchaseOrderList(ctx context.Context, req req.PurchaseOrderListReq) (resp.PurchaseOrderListResp, error)       // 获取采购申请列表
	GetPurchaseOrderDetail(ctx context.Context, req req.PurchaseOrderDetailReq) (resp.PurchaseOrderDetailResp, error) // 获取采购申请详情
	CreatePurchaseOrder(ctx context.Context, req req.PurchaseOrderCreateReq) (resp.PurchaseOrderCreateResp, error)    // 创建采购申请
	UpdatePurchaseOrder(ctx context.Context, req req.PurchaseOrderUpdateReq) error                                    // 更新采购申请
	DeletePurchaseOrder(ctx context.Context, req req.PurchaseOrderDeleteReq) error                                    // 删除采购申请
	SubmitPurchaseOrder(ctx context.Context, req req.PurchaseOrderSubmitReq) error                                    // 提交采购申请
	ApprovePurchaseOrder(ctx context.Context, req req.PurchaseOrderApproveReq) error                                  // 审核采购申请

	// 收货管理
	CreatePurchaseReceiptOrder(ctx context.Context, req req.PurchaseReceiptCreateReq) (resp.PurchaseReceiptOrderCreateResp, error)         // 创建收货单
	GetPurchaseReceiptOrderList(ctx context.Context, req req.PurchaseReceiptOrderListReq) (resp.PurchaseReceiptOrderListResp, error)       // 获取收货单列表
	GetPurchaseReceiptOrderDetail(ctx context.Context, req req.PurchaseReceiptOrderDetailReq) (resp.PurchaseReceiptOrderDetailResp, error) // 获取收货单详情
	UpdatePurchaseReceiptOrder(ctx context.Context, req req.PurchaseReceiptOrderUpdateReq) error                                           // 更新收货单
	CancelPurchaseReceiptOrder(ctx context.Context, req req.PurchaseReceiptOrderCancelReq) error                                           // 取消收货单
}

// purchaseOrderSrv 采购申请服务实现
type purchaseOrderSrv struct {
	dbm        *database.DBManager
	validator  *purchaseOrderValidator
	helper     *purchaseOrderHelper
	receiptSrv *purchaseReceiptOrderSrv
}

// NewPurchaseOrderSrv 创建采购申请服务
func NewPurchaseOrderSrv(dbm *database.DBManager) IPurchaseOrderSrv {
	return NewPurchaseOrderSrvImpl(dbm)
}

// NewPurchaseOrderSrvImpl 创建采购申请服务实现
func NewPurchaseOrderSrvImpl(dbm *database.DBManager) IPurchaseOrderSrv {
	return &purchaseOrderSrv{
		dbm:        dbm,
		validator:  &purchaseOrderValidator{},
		helper:     &purchaseOrderHelper{},
		receiptSrv: newPurchaseReceiptOrderSrv(dbm),
	}
}

// GetPurchaseOrderList 获取采购申请列表
func (s *purchaseOrderSrv) GetPurchaseOrderList(
	ctx context.Context,
	req req.PurchaseOrderListReq,
) (resp.PurchaseOrderListResp, error) {
	db := ctx.GetDB()
	purchaseOrderRepo := repository.NewPurchaseOrderRepo(db)

	// 构建查询选项
	var opts []repository.DBOption
	if req.OrderNo != "" {
		opts = append(opts, purchaseOrderRepo.WhereOrderNo(req.OrderNo))
	}
	if len(req.StatusIn) > 0 {
		opts = append(opts, purchaseOrderRepo.WhereStatusIn(req.StatusIn))
	}
	if req.SupplierName != "" {
		opts = append(opts, purchaseOrderRepo.WhereSupplierName(req.SupplierName))
	}
	if req.WarehouseErpCode != "" {
		opts = append(opts, purchaseOrderRepo.WhereWarehouseErpCode(req.WarehouseErpCode))
	}
	if req.CompanyUuid > 0 {
		opts = append(opts, purchaseOrderRepo.WhereCompanyUuid(req.CompanyUuid))
	}
	if req.CreateTimeStart > 0 || req.CreateTimeEnd > 0 {
		opts = append(opts, purchaseOrderRepo.WhereCreateTimeRange(req.CreateTimeStart, req.CreateTimeEnd))
	}
	if req.OrderTimeStart > 0 || req.OrderTimeEnd > 0 {
		opts = append(opts, purchaseOrderRepo.WhereOrderTimeRange(req.OrderTimeStart, req.OrderTimeEnd))
	}
	if req.ExpectArrivalTimeStart > 0 || req.ExpectArrivalTimeEnd > 0 {
		opts = append(opts, purchaseOrderRepo.WhereExpectArrivalTimeRange(req.ExpectArrivalTimeStart, req.ExpectArrivalTimeEnd))
	}
	if req.ReceiveTimeStart > 0 || req.ReceiveTimeEnd > 0 {
		opts = append(opts, purchaseOrderRepo.WhereReceiveTimeRange(req.ReceiveTimeStart, req.ReceiveTimeEnd))
	}

	// 采购类型
	purchaseType := utils.IfInt(req.PurchaseType == 2, 2, 1)
	opts = append(opts, purchaseOrderRepo.WherePurchaseType(purchaseType))

	// 排序和预加载
	opts = append(opts, purchaseOrderRepo.OrderByOrderTime(true))
	opts = append(opts, purchaseOrderRepo.WithItems())

	// 查询数据
	purchaseOrders, total, err := purchaseOrderRepo.GetListWithPagination(req.PageNo, req.PageSize, opts...)
	if err != nil {
		return resp.PurchaseOrderListResp{}, errors.WithMessage(errors.New("查询采购申请列表失败"), err.Error())
	}

	// 转换响应数据
	listResp := make([]*resp.PurchaseOrderInfo, 0, len(purchaseOrders))
	for _, po := range purchaseOrders {
		poInfo := &resp.PurchaseOrderInfo{}
		if err := copier.Copy(poInfo, &po); err != nil {
			continue
		}
		poInfo.ReceiptProgress = fmt.Sprintf("%.0f%%", po.GetReceiptProgress())
		listResp = append(listResp, poInfo)
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
func (s *purchaseOrderSrv) GetPurchaseOrderDetail(
	ctx context.Context,
	req req.PurchaseOrderDetailReq,
) (resp.PurchaseOrderDetailResp, error) {
	db := ctx.GetDB()
	purchaseOrderRepo := repository.NewPurchaseOrderRepo(db)

	// 查询采购申请详情
	purchaseOrder, err := purchaseOrderRepo.GetByUuid(
		req.Uuid,
		purchaseOrderRepo.WithItems(),
		purchaseOrderRepo.WithWarehouse(),
	)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return resp.PurchaseOrderDetailResp{}, errors.New("采购申请不存在")
		}
		return resp.PurchaseOrderDetailResp{}, errors.WithMessage(errors.New("查询采购申请详情失败"), err.Error())
	}

	// 转换响应数据
	var detailResp resp.PurchaseOrderDetailResp
	if err = copier.Copy(&detailResp, purchaseOrder); err != nil {
		return resp.PurchaseOrderDetailResp{}, errors.WithMessage(errors.New("数据转换失败"), err.Error())
	}

	// 转换仓库名称
	if purchaseOrder.Warehouse != nil {
		detailResp.WarehouseName = *language.JsonToLocaleResponse(purchaseOrder.Warehouse.Name)
	}

	// 初始化数组字段
	detailResp.Items = make([]resp.PurchaseOrderItemInfo, 0, len(purchaseOrder.Items))
	detailResp.ReceiptProgress = fmt.Sprintf("%.0f%%", purchaseOrder.GetReceiptProgress())

	// 转换明细数据
	for _, item := range purchaseOrder.Items {
		itemInfo := resp.PurchaseOrderItemInfo{}
		copier.Copy(&itemInfo, &item)
		itemInfo.LocaleName = *language.JsonToLocaleResponse(item.MaterialName)
		itemInfo.LocaleUnitName = *language.JsonToLocaleResponse(item.UnitName)
		itemInfo.LocaleBaseUnitName = *language.JsonToLocaleResponse(item.BaseUnitName)
		detailResp.Items = append(detailResp.Items, itemInfo)
	}

	return detailResp, nil
}

// CreatePurchaseOrder 创建采购申请
func (s *purchaseOrderSrv) CreatePurchaseOrder(
	ctx context.Context,
	req req.PurchaseOrderCreateReq,
) (resp.PurchaseOrderCreateResp, error) {
	if err := req.Validate(); err != nil {
		return resp.PurchaseOrderCreateResp{}, err
	}

	// 版本检查
	if ctx.Version(context.GTE, "2.6.0") {
		if req.SupplierErpCode == "" {
			return resp.PurchaseOrderCreateResp{}, errors.New("供应商编码不能为空")
		}
		if req.PurchaseType == 2 && req.WarehouseErpCode == "" {
			return resp.PurchaseOrderCreateResp{}, errors.New("仓库编码不能为空")
		}
	}

	db := ctx.GetDB()
	var result resp.PurchaseOrderCreateResp

	err := db.Transaction(func(tx *gorm.DB) error {
		purchaseOrderRepo := repository.NewPurchaseOrderRepo(tx)
		purchaseOrderItemRepo := repository.NewPurchaseOrderItemRepo(tx)

		// 获取默认仓库
		defaultWarehouse, err := repository.NewWarehouseRepo(tx).GetDefaultWarehouse()
		if err != nil {
			return errors.WithMessage(errors.New("获取默认仓库失败"), err.Error())
		}

		// 生成订单编号
		prefix := utils.IfString(req.PurchaseType == 2, "TPHY", "CSSQ")
		orderNo := s.helper.generateOrderNo(tx, prefix)

		// 获取仓库名称
		warehouseName := ""
		if req.WarehouseErpCode != "" {
			warehouse, err := repository.NewWarehouseRepo(tx).GetByErpCode(req.WarehouseErpCode)
			if err == nil {
				warehouseName = warehouse.Name
			}
		}

		// 创建采购申请
		purchaseOrder := &model.PurchaseOrder{
			OrderNo:           orderNo,
			SupplierName:      req.SupplierName,
			SupplierErpCode:   utils.IfString(req.SupplierErpCode != "", req.SupplierErpCode, req.SupplierName),
			Status:            constant.PurchaseOrderStatusDraft,
			Num:               float64(len(req.Items)),
			OrderTime:         req.OrderTime,
			ExpectArrivalTime: req.ExpectedDeliveryTime,
			ApplicantUuid:     ctx.GetStaffUuid(),
			ApplicantName:     ctx.GetStaff().RealName,
			PurchaseType:      utils.IfInt(req.PurchaseType == 2, 2, 1),
			WarehouseErpCode:  req.WarehouseErpCode,
			WarehouseName:     warehouseName,
		}

		// 设置默认仓库信息
		if defaultWarehouse != nil {
			purchaseOrder.DefaultWarehouseErpCode = defaultWarehouse.ErpCode
			purchaseOrder.DefaultWarehouseName = defaultWarehouse.Name
		}

		err = purchaseOrderRepo.Create(purchaseOrder)
		if err != nil {
			return errors.WithMessage(errors.New("创建采购申请失败"), err.Error())
		}

		// 构建采购申请明细
		itemReqs := make([]PurchaseOrderItemReq, 0, len(req.Items))
		for _, item := range req.Items {
			itemReqs = append(itemReqs, PurchaseOrderItemReq{
				MaterialUuid: item.MaterialUuid,
				Num:          item.Num,
			})
		}
		items, err := s.validator.buildPurchaseOrderItems(tx, purchaseOrder.Uuid, itemReqs)
		if err != nil {
			return err
		}

		err = purchaseOrderItemRepo.CreateBatch(items)
		if err != nil {
			return errors.WithMessage(errors.New("创建采购申请明细失败"), err.Error())
		}

		// 记录操作日志
		err = s.helper.createPurchaseOrderLog(
			tx,
			purchaseOrder.Uuid,
			ctx,
			"create",
			"创建采购申请",
			0,
			constant.PurchaseOrderStatusPending,
			"",
		)
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
func (s *purchaseOrderSrv) UpdatePurchaseOrder(
	ctx context.Context,
	req req.PurchaseOrderUpdateReq,
) error {
	if err := req.Validate(); err != nil {
		return err
	}

	db := ctx.GetDB()

	return db.Transaction(func(tx *gorm.DB) error {
		purchaseOrderRepo := repository.NewPurchaseOrderRepo(tx)
		purchaseOrderItemRepo := repository.NewPurchaseOrderItemRepo(tx)

		// 查询现有采购申请
		purchaseOrder, err := purchaseOrderRepo.GetByUuid(req.Uuid)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return errors.New("采购申请不存在")
			}
			return errors.WithMessage(errors.New("查询采购申请失败"), err.Error())
		}

		// 检查是否可编辑
		if !purchaseOrder.IsEditable() {
			return errors.New("当前状态不允许编辑")
		}

		// 版本检查
		if ctx.Version(context.GTE, "2.6.0") {
			if req.SupplierErpCode == "" {
				return errors.New("供应商编码不能为空")
			}
			if purchaseOrder.IsHeadquarterPurchase() && req.WarehouseErpCode == "" {
				return errors.New("仓库编码不能为空")
			}
		}

		// 获取仓库名称
		warehouseName := ""
		if req.WarehouseErpCode != "" {
			warehouse, err := repository.NewWarehouseRepo(tx).GetByErpCode(req.WarehouseErpCode)
			if err == nil {
				warehouseName = warehouse.Name
			}
		}

		// 更新采购申请基本信息
		purchaseOrder.Num = float64(len(req.Items))
		purchaseOrder.SupplierName = req.SupplierName
		purchaseOrder.SupplierErpCode = req.SupplierErpCode
		purchaseOrder.ExpectArrivalTime = req.ExpectedDeliveryTime
		purchaseOrder.WarehouseErpCode = req.WarehouseErpCode
		purchaseOrder.WarehouseName = warehouseName

		err = purchaseOrderRepo.Update(purchaseOrder)
		if err != nil {
			return errors.WithMessage(errors.New("更新采购申请失败"), err.Error())
		}

		// 先删除所有现有明细项
		err = purchaseOrderItemRepo.DeleteByPurchaseOrderUuid(req.Uuid)
		if err != nil {
			return errors.WithMessage(errors.New("删除采购申请明细失败"), err.Error())
		}

		// 构建并批量创建采购申请明细
		itemReqs := make([]PurchaseOrderItemReq, 0, len(req.Items))
		for _, item := range req.Items {
			itemReqs = append(itemReqs, PurchaseOrderItemReq{
				MaterialUuid: item.MaterialUuid,
				Num:          item.Num,
			})
		}
		items, err := s.validator.buildPurchaseOrderItems(tx, purchaseOrder.Uuid, itemReqs)
		if err != nil {
			return err
		}

		err = purchaseOrderItemRepo.CreateBatch(items)
		if err != nil {
			return errors.WithMessage(errors.New("创建采购申请明细失败"), err.Error())
		}

		// 记录操作日志
		err = s.helper.createPurchaseOrderLog(
			tx,
			req.Uuid,
			ctx,
			"update",
			"更新采购申请",
			purchaseOrder.Status,
			purchaseOrder.Status,
			"",
		)
		if err != nil {
			return err
		}

		return nil
	})
}

// DeletePurchaseOrder 删除采购申请
func (s *purchaseOrderSrv) DeletePurchaseOrder(
	ctx context.Context,
	req req.PurchaseOrderDeleteReq,
) error {
	db := s.dbm.GetDB(ctx.GetDbId())
	purchaseOrderRepo := repository.NewPurchaseOrderRepo(db)

	// 查询采购申请
	purchaseOrder, err := purchaseOrderRepo.GetByUuid(req.Uuid)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("采购申请不存在")
		}
		return errors.WithMessage(errors.New("查询采购申请失败"), err.Error())
	}

	// 检查是否可删除（只有待提交和已驳回状态可以删除）
	if !purchaseOrder.IsEditable() {
		return errors.New("当前状态不允许删除")
	}

	// 删除采购申请（软删除）
	err = purchaseOrderRepo.Delete(req.Uuid)
	if err != nil {
		return errors.WithMessage(errors.New("删除采购申请失败"), err.Error())
	}

	return nil
}

// SubmitPurchaseOrder 提交采购申请
func (s *purchaseOrderSrv) SubmitPurchaseOrder(
	ctx context.Context,
	req req.PurchaseOrderSubmitReq,
) error {
	db := ctx.GetDB()

	return db.Transaction(func(tx *gorm.DB) error {
		purchaseOrderRepo := repository.NewPurchaseOrderRepo(tx)
		purchaseOrderItemRepo := repository.NewPurchaseOrderItemRepo(tx)

		// 查询采购申请
		purchaseOrder, err := purchaseOrderRepo.GetByUuid(req.Uuid, purchaseOrderRepo.WithItems())
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return errors.New("采购申请不存在")
			}
			return errors.WithMessage(errors.New("查询采购申请失败"), err.Error())
		}

		// 检查供应商状态
		if err := s.validator.validateSupplierStatus(tx, purchaseOrder.SupplierErpCode); err != nil {
			return err
		}

		// 检查物品状态
		if !req.IsConfirm {
			disabledMaterials, err := s.validator.validateMaterialStatus(ctx, tx, purchaseOrder.Items, true)
			if err != nil {
				return err
			}

			// 如果有禁用的物品，返回提示信息
			if len(disabledMaterials) > 0 {
				materialNames := s.validator.joinMaterialNames(disabledMaterials)
				return errors.NewWithCodeAndData(
					constant.CodeMaterialDisabled,
					disabledMaterials,
					fmt.Sprintf(
						i18n.Translate(ctx.GetLanguage(), "物品 %s 的状态已关闭。\n\n提交后将移除该物品，是否继续提交？"),
						materialNames,
					),
				)
			}
		}

		// 删除物品为0的数据
		err = purchaseOrderItemRepo.DeleteByPurchaseOrderUuidAndNumIsZero(req.Uuid)
		if err != nil {
			return errors.WithMessage(errors.New("删除采购申请明细失败"), err.Error())
		}

		// 如果用户确认提交，删除禁用的物品
		if req.IsConfirm {
			materialRepo := repository.NewMaterialRepo(tx)
			var disabledMaterialUuids []uint64

			for _, item := range purchaseOrder.Items {
				if item.Num <= 0 {
					continue
				}
				material, err := materialRepo.GetMaterialByUuid(item.MaterialUuid)
				if err != nil || !material.Status {
					disabledMaterialUuids = append(disabledMaterialUuids, item.MaterialUuid)
				}
			}

			// 删除禁用的物品记录
			if len(disabledMaterialUuids) > 0 {
				err = purchaseOrderItemRepo.DeleteByPurchaseOrderUuidAndMaterialUuids(req.Uuid, disabledMaterialUuids)
				if err != nil {
					return errors.WithMessage(errors.New("删除禁用物品失败"), err.Error())
				}
			}
		}

		// 重新查询采购申请以获取最新的物品列表
		purchaseOrder, err = purchaseOrderRepo.GetByUuid(req.Uuid, purchaseOrderRepo.WithItems())
		if err != nil {
			return errors.WithMessage(errors.New("重新查询采购申请失败"), err.Error())
		}

		// 检查采购申请明细
		if len(purchaseOrder.Items) == 0 {
			return errors.New("申请物品数量不能为0")
		}

		// 过滤掉数量为0的项目后重新计算数量
		validItemCount := 0
		for _, item := range purchaseOrder.Items {
			if item.Num > 0 {
				validItemCount++
			}
		}

		oldStatus := purchaseOrder.Status
		purchaseOrder.Status = constant.PurchaseOrderStatusPending
		purchaseOrder.OrderTime = time.Now().Unix()
		purchaseOrder.ApplicantUuid = ctx.GetStaffUuid()
		purchaseOrder.ApplicantName = ctx.GetStaff().RealName
		purchaseOrder.Num = float64(validItemCount)

		err = purchaseOrderRepo.Update(purchaseOrder)
		if err != nil {
			return errors.WithMessage(errors.New("更新采购申请状态失败"), err.Error())
		}

		// 记录操作日志
		statusText := purchaseOrder.GetStatusText()
		err = s.helper.createPurchaseOrderLog(
			tx,
			req.Uuid,
			ctx,
			"status_update",
			"更新状态为"+statusText,
			oldStatus,
			purchaseOrder.Status,
			"",
		)
		if err != nil {
			return err
		}

		return nil
	})
}

// ApprovePurchaseOrder 审核采购申请
func (s *purchaseOrderSrv) ApprovePurchaseOrder(
	ctx context.Context,
	req req.PurchaseOrderApproveReq,
) error {
	db := ctx.GetDB()
	companySetting := ctx.GetCompanySetting()

	return db.Transaction(func(tx *gorm.DB) error {
		purchaseOrderRepo := repository.NewPurchaseOrderRepo(tx)

		// 查询采购申请
		purchaseOrder, err := purchaseOrderRepo.GetByUuid(req.Uuid, purchaseOrderRepo.WithItems())
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return errors.New("采购申请不存在")
			}
			return errors.WithMessage(err, "查询采购申请失败")
		}

		// 检查是否可审核
		if !purchaseOrder.CanApprove() && (companySetting.IsSubShop() || companySetting.IsHeadquarter()) {
			return errors.New("当前状态不允许审核")
		}

		// 检查供应商状态
		if err := s.validator.validateSupplierStatus(tx, purchaseOrder.SupplierErpCode); err != nil {
			return err
		}

		// 审核通过时检查物品状态
		if req.Action == "approve" {
			_, err := s.validator.validateMaterialStatus(ctx, tx, purchaseOrder.Items, false)
			if err != nil {
				return err
			}
		}

		// 处理审核逻辑
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

		// 更新状态
		purchaseOrder.Status = newStatus
		if purchaseOrder.IsHeadquarterPurchase() && (companySetting.IsHeadquarter() || newStatus != constant.PurchaseOrderStatusRejected) {
			purchaseOrder.HeadquarterStatus = newStatus
		}

		// 子店内部采购审核通过时，状态改为待总部审核
		if companySetting.IsSubShop() && purchaseOrder.IsHeadquarterPurchase() && newStatus == constant.PurchaseOrderStatusApproved {
			purchaseOrder.Status = constant.PurchaseOrderStatusHeadquarterPending
			purchaseOrder.HeadquarterStatus = constant.HeadquarterStatusPending
		}

		// 更新状态
		err = purchaseOrderRepo.Update(purchaseOrder)
		if err != nil {
			return errors.WithMessage(errors.New("更新采购申请状态失败"), err.Error())
		}

		// 记录操作日志
		err = s.helper.createPurchaseOrderLog(tx, req.Uuid, ctx, req.Action, actionDesc, oldStatus, newStatus, "")
		if err != nil {
			return errors.WithMessage(errors.New("记录操作日志失败"), err.Error())
		}

		// 处理总部驳回场景
		if companySetting.IsHeadquarter() && purchaseOrder.IsHeadquarterPurchase() && purchaseOrder.Status == constant.PurchaseOrderStatusRejected {
			return s.handleHeadquarterReject(purchaseOrder)
		}

		// 处理子店审核通过场景
		if companySetting.IsSubShop() && purchaseOrder.IsHeadquarterPurchase() && purchaseOrder.Status == constant.PurchaseOrderStatusHeadquarterPending {
			return s.handleSubShopApproval(ctx, purchaseOrder)
		}

		// 调用ERP接口
		if ctx.GetCompany().IsOpenErp() && purchaseOrder.Status == constant.PurchaseOrderStatusApproved {
			return s.handleErpApproval(ctx, tx, &companySetting, purchaseOrder)
		}

		return nil
	})
}

// handleHeadquarterReject 处理总部驳回场景
func (s *purchaseOrderSrv) handleHeadquarterReject(purchaseOrder *model.PurchaseOrder) error {
	// 获取子店数据库
	subDb := s.dbm.GetDB(purchaseOrder.CompanyUuid)
	if subDb == nil {
		return errors.New("获取子店数据库失败")
	}

	subPurchaseOrder, err := repository.NewPurchaseOrderRepo(subDb).GetByUuid(purchaseOrder.SubUuid)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("采购申请不存在")
		}
		return errors.WithMessage(errors.New("查询采购申请失败"), err.Error())
	}

	subPurchaseOrder.Status = constant.PurchaseOrderStatusRejected
	subPurchaseOrder.HeadquarterStatus = constant.HeadquarterStatusRejected
	subPurchaseOrder.RejectTime = purchaseOrder.RejectTime

	err = repository.NewPurchaseOrderRepo(subDb).Update(subPurchaseOrder)
	if err != nil {
		return errors.WithMessage(errors.New("更新子店采购申请失败"), err.Error())
	}

	return nil
}

// handleSubShopApproval 处理子店审核通过场景
func (s *purchaseOrderSrv) handleSubShopApproval(
	ctx context.Context,
	purchaseOrder *model.PurchaseOrder,
) error {
	companySetting := ctx.GetCompanySetting()

	// 获取总部数据库
	headquarterDb := s.dbm.GetDB(companySetting.HeadquarterUuid)
	if headquarterDb == nil {
		return errors.New("获取总部数据库失败")
	}

	return headquarterDb.Transaction(func(hqTx *gorm.DB) error {
		// 整单复制到总部
		subUuid := purchaseOrder.Uuid
		headquarterPurchaseOrder := model.PurchaseOrder{}

		err := copier.Copy(&headquarterPurchaseOrder, purchaseOrder)
		if err != nil {
			return errors.WithMessage(errors.New("复制总部采购订单失败"), err.Error())
		}

		// 重置主键字段
		headquarterPurchaseOrder.BaseModel.ID = 0
		headquarterPurchaseOrder.BaseModel.Uuid = func() uint64 {
			uuid, _ := utils.GetID()
			return uuid
		}()
		headquarterPurchaseOrder.CompanyUuid = ctx.GetCompanyUuid()
		headquarterPurchaseOrder.CompanyName = ctx.GetCompany().Name
		headquarterPurchaseOrder.SubUuid = subUuid
		headquarterPurchaseOrder.Status = constant.PurchaseOrderStatusPending
		headquarterPurchaseOrder.HeadquarterStatus = constant.HeadquarterStatusPending

		err = repository.NewPurchaseOrderRepo(hqTx).Create(&headquarterPurchaseOrder)
		if err != nil {
			logger.Logger.Error("创建总部采购申请失败", zap.Error(err))
			return errors.WithMessage(errors.New("创建总部采购申请失败"), err.Error())
		}

		// 创建总部采购申请明细
		var headquarterItems []model.PurchaseOrderItem
		for _, item := range purchaseOrder.Items {
			headquarterItem := model.PurchaseOrderItem{}
			err = copier.Copy(&headquarterItem, item)
			if err != nil {
				logger.Logger.Error("复制总部物品明细失败", zap.Error(err), zap.String("物料编码", item.MaterialCode))
				return errors.WithMessage(errors.New("复制总部物品明细失败"), err.Error())
			}

			material, err := repository.NewMaterialRepo(hqTx).GetMaterialByErpCode(item.MaterialCode)
			if err != nil {
				logger.Logger.Error("总部不存在该物料", zap.String("物料编码", item.MaterialCode), zap.Error(err))
				return errors.New(fmt.Sprintf(
					i18n.Translate(ctx.GetLanguage(), "总部不存在该物料，物料编码：%s"),
					item.MaterialCode,
				))
			}

			// 重置主键字段
			headquarterItem.BaseModel.ID = 0
			headquarterItem.BaseModel.Uuid = 0
			headquarterItem.PurchaseOrderUuid = headquarterPurchaseOrder.Uuid
			headquarterItem.MaterialUuid = material.Uuid
			headquarterItem.Material = nil
			headquarterItems = append(headquarterItems, headquarterItem)
		}

		err = repository.NewPurchaseOrderItemRepo(hqTx).CreateBatch(headquarterItems)
		if err != nil {
			logger.Logger.Error("创建总部采购申请明细失败", zap.Error(err))
			return errors.WithMessage(errors.New("创建总部采购申请明细失败"), err.Error())
		}

		return nil
	})
}

// handleErpApproval 处理ERP审核
func (s *purchaseOrderSrv) handleErpApproval(
	ctx context.Context,
	tx *gorm.DB,
	companySetting *model.CompanySetting,
	purchaseOrder *model.PurchaseOrder,
) error {
	purchaseOrderRepo := repository.NewPurchaseOrderRepo(tx)

	var erpOrderNo string
	var err error

	if purchaseOrder.IsHeadquarterPurchase() {
		// 内部采购
		erpOrderNo, err = s.handleInternalPurchaseErp(ctx, tx, purchaseOrder)
		if err != nil {
			return err
		}
	} else {
		// 外部采购
		erpOrderNo, err = s.handleExternalPurchaseErp(ctx, tx, companySetting, purchaseOrder)
		if err != nil {
			return err
		}
	}

	// 更新采购申请单号
	purchaseOrder.ErpOrderNo = erpOrderNo
	err = purchaseOrderRepo.Update(purchaseOrder)
	if err != nil {
		return errors.WithMessage(errors.New("更新采购申请单号失败"), err.Error())
	}

	// 同步状态到子商户采购申请
	if purchaseOrder.IsHeadquarterPurchase() {
		return s.syncToSubShop(purchaseOrder)
	}

	return nil
}

// handleInternalPurchaseErp 处理内部采购ERP
func (s *purchaseOrderSrv) handleInternalPurchaseErp(
	ctx context.Context,
	tx *gorm.DB,
	purchaseOrder *model.PurchaseOrder,
) (string, error) {
	// 构建物料请求项
	stockItems := make([]*stock.MaterialRequestItem, 0, len(purchaseOrder.Items))
	for _, item := range purchaseOrder.Items {
		erpnextUom := item.ErpnextUom
		if erpnextUom == "" {
			materialUnit, err := repository.NewMaterialUnitRepo(tx).GetMaterialUnitsByUuid(item.UnitUuid)
			if err != nil {
				return "", errors.WithMessage(errors.New("查询物品单位失败"), err.Error())
			}
			if materialUnit.Unit == nil {
				return "", errors.New("查询物品原始单位失败")
			}
			erpnextUom = materialUnit.Unit.ErpnextUom
		}
		stockItems = append(stockItems, &stock.MaterialRequestItem{
			ItemCode:     item.MaterialCode,
			Qty:          item.Num,
			ScheduleDate: purchaseOrder.ExpectArrivalTime,
			Uom:          erpnextUom,
		})
	}

	// 获取子公司数据库
	subDb := s.dbm.GetDB(purchaseOrder.CompanyUuid)
	if subDb == nil {
		return "", errors.New("获取子店数据库失败")
	}

	// 减总部库存并记录出入库日志
	{
		// 添加物料库存
		materialRepo := repository.NewMaterialRepo(tx)
		for _, item := range purchaseOrder.Items {
			// 获取物料信息
			material, err := materialRepo.GetMaterialByUuid(item.MaterialUuid)
			if err != nil {
				return "", errors.WithMessage(errors.New("获取物品信息失败"), err.Error())
			}
			//
			stockNum := material.GetStockNum()
			conversionRateNum := item.GetConversionRateNum()
			// 更新库存数量
			latestStockNum := decimal.NewFromFloat(material.GetStockNum()).
				Sub(decimal.NewFromFloat(conversionRateNum)).
				InexactFloat64()
			if latestStockNum < 0 {
				return "", errors.New(fmt.Sprintf(
					i18n.Translate(ctx.GetLanguage(), "物品库存不足，物品编码：%s"),
					item.MaterialCode,
				) + fmt.Sprintf("  (%v / %v)", stockNum, conversionRateNum))
			}
			err = base.NewMaterialRepo(tx).UpdateMaterialsStockNum(material.Uuid, material.WarehouseUuid, latestStockNum)
			if err != nil {
				return "", errors.WithMessage(errors.New("更新物品库存失败"), err.Error())
			}

			// 更新规格/加料关联材料库存
			relatedMaterialUuids := make([]uint64, 0)
			for _, relatedMaterial := range material.RelatedMaterialList {
				if relatedMaterial.IsDelete() || relatedMaterial.IsUsed == 0 {
					continue
				}
				relatedMaterialUuids = append(relatedMaterialUuids, relatedMaterial.Uuid)
			}
			err = s.helper.updateRelatedMaterialStock(tx, relatedMaterialUuids)
			if err != nil {
				return "", errors.WithMessage(errors.New("更新规格/加料关联材料库存失败"), err.Error())
			}
		}

		// 减总部库存并记录出入库日志
		err := s.helper.reduceHeadquarterStockAndLog(subDb, tx, purchaseOrder)
		if err != nil {
			return "", errors.WithMessage(errors.New("处理总部库存失败"), err.Error())
		}
	}

	// 获取子公司设置
	subCompanySetting := repository.NewCompanySettingRepo(subDb).Get()

	// 调用erp接口
	stockResp, err := erp.NewIErpSrv(s.dbm).SaveMaterialRequest(ctx, subCompanySetting, &stock.SaveMaterialRequestReq{
		TransactionDate: purchaseOrder.OrderTime,
		RequiredBy:      purchaseOrder.ExpectArrivalTime,
		Supplier: func() string {
			if purchaseOrder.SupplierErpCode != "" {
				return purchaseOrder.SupplierErpCode
			}
			return purchaseOrder.SupplierName
		}(),
		SourceWarehouse: purchaseOrder.WarehouseErpCode,
		TargetWarehouse: purchaseOrder.DefaultWarehouseErpCode,
		Items:           stockItems,
	})
	if err != nil {
		return "", s.handleErpError(err)
	}

	return stockResp.PurchaseOrder, nil
}

// handleExternalPurchaseErp 处理外部采购ERP
func (s *purchaseOrderSrv) handleExternalPurchaseErp(
	ctx context.Context,
	tx *gorm.DB,
	companySetting *model.CompanySetting,
	purchaseOrder *model.PurchaseOrder,
) (string, error) {
	// 构建采购订单项
	stockItems := make([]*buying.PurchaseOrderItemInput, 0, len(purchaseOrder.Items))
	for _, item := range purchaseOrder.Items {
		erpnextUom := item.ErpnextUom
		if erpnextUom == "" {
			materialUnit, err := repository.NewMaterialUnitRepo(tx).GetMaterialUnitsByUuid(item.UnitUuid)
			if err != nil {
				return "", errors.WithMessage(errors.New("查询物品单位失败"), err.Error())
			}
			if materialUnit.Unit == nil {
				return "", errors.New("查询物品原始单位失败")
			}
			erpnextUom = materialUnit.Unit.ErpnextUom
		}
		stockItems = append(stockItems, &buying.PurchaseOrderItemInput{
			ItemCode: item.MaterialCode,
			Qty:      item.Num,
			Uom:      erpnextUom,
		})
	}

	// 将时间戳转换为 Y-m-d 格式
	scheduleDate := ""
	if purchaseOrder.ExpectArrivalTime > 0 {
		scheduleDate = time.Unix(purchaseOrder.ExpectArrivalTime, 0).Format("2006-01-02")
	}

	// 调用ERP接口
	erpResp, err := erp.NewIErpSrv(s.dbm).CreatePurchaseOrder(ctx, &buying.CreatePurchaseOrderReq{
		Supplier: func() string {
			if purchaseOrder.SupplierErpCode != "" {
				return purchaseOrder.SupplierErpCode
			}
			return purchaseOrder.SupplierName
		}(),
		CompanyAbbr:     companySetting.ErpnextCompanyAbbr,
		ScheduleDate:    scheduleDate,
		TargetWarehouse: purchaseOrder.DefaultWarehouseErpCode,
		Items:           stockItems,
	})
	if err != nil {
		return "", s.handleErpError(err)
	}

	return erpResp.Name, nil
}

// syncToSubShop 同步状态到子商户
func (s *purchaseOrderSrv) syncToSubShop(purchaseOrder *model.PurchaseOrder) error {
	// 获取子店数据库
	subDb := s.dbm.GetDB(purchaseOrder.CompanyUuid)
	if subDb == nil {
		return errors.New("获取子店数据库失败")
	}

	subPurchaseOrder, err := repository.NewPurchaseOrderRepo(subDb).GetByUuid(purchaseOrder.SubUuid)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("采购申请不存在")
		}
		return errors.WithMessage(errors.New("查询采购申请失败"), err.Error())
	}

	subPurchaseOrder.ErpOrderNo = purchaseOrder.ErpOrderNo
	subPurchaseOrder.Status = constant.PurchaseOrderStatusApproved
	subPurchaseOrder.HeadquarterStatus = constant.HeadquarterStatusApproved
	subPurchaseOrder.PassTime = purchaseOrder.PassTime

	err = repository.NewPurchaseOrderRepo(subDb).Update(subPurchaseOrder)
	if err != nil {
		return errors.WithMessage(errors.New("更新子店采购申请失败"), err.Error())
	}

	return nil
}

// handleErpError 处理ERP错误
func (s *purchaseOrderSrv) handleErpError(err error) error {
	// 检查供应商状态
	if strings.Contains(err.Error(), "Supplier") && strings.Contains(err.Error(), "is disabled") {
		return errors.NewWithCode(constant.CodePurchaseOrderSupplierDisabled, "供应商已禁用，请修改供应商状态")
	}
	// 检查物品状态
	if strings.Contains(err.Error(), "Item") && strings.Contains(err.Error(), "is disabled") {
		return errors.NewWithCode(constant.CodeMaterialDisabled, "物品已禁用，请修改物品状态")
	}
	return errors.WithMessage(errors.New("调用erp接口失败"), err.Error())
}

// 收货单相关方法委托给receiptSrv
func (s *purchaseOrderSrv) CreatePurchaseReceiptOrder(
	ctx context.Context,
	req req.PurchaseReceiptCreateReq,
) (resp.PurchaseReceiptOrderCreateResp, error) {
	return s.receiptSrv.CreatePurchaseReceiptOrder(ctx, req)
}

func (s *purchaseOrderSrv) UpdatePurchaseReceiptOrder(
	ctx context.Context,
	req req.PurchaseReceiptOrderUpdateReq,
) error {
	return s.receiptSrv.UpdatePurchaseReceiptOrder(ctx, req)
}

func (s *purchaseOrderSrv) GetPurchaseReceiptOrderList(
	ctx context.Context,
	req req.PurchaseReceiptOrderListReq,
) (resp.PurchaseReceiptOrderListResp, error) {
	return s.receiptSrv.GetPurchaseReceiptOrderList(ctx, req)
}

func (s *purchaseOrderSrv) GetPurchaseReceiptOrderDetail(
	ctx context.Context,
	req req.PurchaseReceiptOrderDetailReq,
) (resp.PurchaseReceiptOrderDetailResp, error) {
	return s.receiptSrv.GetPurchaseReceiptOrderDetail(ctx, req)
}

func (s *purchaseOrderSrv) CancelPurchaseReceiptOrder(
	ctx context.Context,
	req req.PurchaseReceiptOrderCancelReq,
) error {
	return s.receiptSrv.CancelPurchaseReceiptOrder(ctx, req)
}
