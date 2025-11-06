package transfer_order

import (
	"fmt"
	"strings"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/dto/resp/material_resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service"
	"ttpos-server-go/i18n"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/language"
	"ttpos-server-go/pkg/lock"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"github.com/jinzhu/copier"
	"github.com/shopspring/decimal"
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

	// 下拉列表
	GetTransferOrderCompanyList(ctx context.Context) (resp.TransferOrderCompanyListResp, error)
	GetTransferOrderWarehouseList(ctx context.Context) (resp.TransferOrderWarehouseListResp, error)
	GetTransferOrderMaterialList(ctx context.Context, req req.TransferOrderMaterialListReq) (material_resp.MaterialListWithPaginationResp, error)
}

// transferOrderSrv 调拨单服务实现
type transferOrderSrv struct {
	dbm         *database.DBManager
	materialSrv service.IMaterialSrv
	lock        lock.Lock
	validator   *transferOrderValidator
	helper      *transferOrderHelper
}

// NewTransferOrderSrv 创建调拨单服务
func NewTransferOrderSrv(dbm *database.DBManager, materialSrv service.IMaterialSrv) ITransferOrderSrv {
	return NewTransferOrderSrvImpl(dbm, materialSrv)
}

// NewTransferOrderSrvImpl 创建调拨单服务实现
func NewTransferOrderSrvImpl(dbm *database.DBManager, materialSrv service.IMaterialSrv) ITransferOrderSrv {
	return &transferOrderSrv{
		dbm:         dbm,
		materialSrv: materialSrv,
		lock:        lock.NewSystemLock(),
		validator:   &transferOrderValidator{},
		helper:      &transferOrderHelper{},
	}
}

// GetTransferOrderList 获取调拨单列表
func (s *transferOrderSrv) GetTransferOrderList(
	ctx context.Context,
	req req.TransferOrderListReq,
) (resp.TransferOrderListResp, error) {
	// 获取总部数据库
	db := s.dbm.GetDB(0)
	transferOrderRepo := repository.NewTransferOrderRepo(db)

	// 查询数据
	transferOrders, total, err := transferOrderRepo.GetListWithPaginationFromMultiDB(
		repository.TransferOrderQueryReq{
			PageNo:              req.PageNo,
			PageSize:            req.PageSize,
			CompanyUuid:         ctx.GetCompanyUuid(),
			OrderNo:             req.OrderNo,
			StatusIn:            req.StatusIn,
			OrderTimeStart:      req.OrderTimeStart,
			OrderTimeEnd:        req.OrderTimeEnd,
			OppositeCompanyUuid: req.OppositeCompanyUuid,
			MyRole:              req.MyRole,
		},
	)
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
		// 转换状态
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
	// 获取调拨单数据库
	db, err := s.helper.GetOrderDb(ctx, s.dbm, req.Uuid)
	if err != nil {
		return resp.TransferOrderDetailResp{}, err
	}

	// 查询调拨单详情
	transferOrderRepo := repository.NewTransferOrderRepo(db)
	transferOrder, err := transferOrderRepo.GetByUuid(
		req.Uuid,
		transferOrderRepo.WithItems(),
		transferOrderRepo.WithApprovals(),
	)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return resp.TransferOrderDetailResp{}, errors.New("调拨单不存在")
		}
		logger.Logger.Error("查询调拨单详情失败", zap.Error(err))
		return resp.TransferOrderDetailResp{}, errors.WithMessage(errors.New("查询调拨单详情失败"), err.Error())
	}

	// 转换响应数据
	var detailResp resp.TransferOrderDetailResp
	if err = copier.Copy(&detailResp, transferOrder); err != nil {
		logger.Logger.Error("数据转换失败", zap.Error(err))
		return resp.TransferOrderDetailResp{}, errors.WithMessage(errors.New("数据转换失败"), err.Error())
	}

	// 转换仓库名称
	detailResp.OutWarehouseName = *language.JsonToLocaleResponse(transferOrder.OutWarehouseName)
	detailResp.InWarehouseName = *language.JsonToLocaleResponse(transferOrder.InWarehouseName)

	// 是否可审批
	detailResp.IsCanApprove = func() bool {
		if transferOrder.Status == constant.TransferOrderStatusPending {
			return transferOrder.NextApprovalCompanyUuid == ctx.GetCompanyUuid()
		}
		return false
	}()

	// 是否需要选择仓库
	detailResp.ErpOrderNo = func() string {
		if transferOrder.ErpResp != "" {
			return strings.Join(transferOrder.GetErpOrderNos(), "、")
		}
		return ""
	}()

	// 获取当前审批节点
	if transferOrder.Status == constant.TransferOrderStatusPending && (transferOrder.OutWarehouseErpCode == "" || transferOrder.InWarehouseErpCode == "") {
		currentApproval, err := repository.NewTransferOrderApprovalRepo(db).GetCurrentApproval(req.Uuid, transferOrder.NextApprovalCompanyUuid)
		if err != nil {
			if err != gorm.ErrRecordNotFound {
				return resp.TransferOrderDetailResp{}, errors.WithMessage(errors.New("查询审批节点失败"), err.Error())
			}
		}

		// 是否需要选择出库仓库
		detailResp.IsNeedSelectOutWarehouse = func() bool {
			if currentApproval != nil && currentApproval.ApprovalType == constant.TransferApprovalTypeSender {
				return transferOrder.OutWarehouseErpCode == ""
			}
			return false
		}()

		// 是否需要选择入库仓库
		detailResp.IsNeedSelectInWarehouse = func() bool {
			if currentApproval != nil && currentApproval.ApprovalType == constant.TransferApprovalTypeReceiver {
				return transferOrder.InWarehouseErpCode == ""
			}
			return false
		}()
	}

	// 待收货时 - 入库仓库永远都可以选择
	if transferOrder.Status == constant.TransferOrderStatusReceiving {
		detailResp.IsNeedSelectInWarehouse = true
	}

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
	// detailResp.Approvals = make([]resp.TransferOrderApprovalInfo, 0, len(transferOrder.Approvals))

	// 审批类型文本映射
	approvalTypeText := map[string]string{
		constant.TransferApprovalTypeSender:         "发货门店驳回",   // "发货门店驳回"
		constant.TransferApprovalTypeSenderParent:   "发货门店上级驳回", // 发货门店上级驳回
		constant.TransferApprovalTypeReceiver:       "收货门店驳回",   // 收货门店驳回
		constant.TransferApprovalTypeReceiverParent: "收货门店上级驳回", // 收货门店上级驳回
	}

	for _, approval := range transferOrder.Approvals {
		approvalInfo := resp.TransferOrderApprovalInfo{}
		copier.Copy(&approvalInfo, &approval)
		// detailResp.Approvals = append(detailResp.Approvals, approvalInfo)

		// 只保留第一个驳回信息（因为永远只能有一个驳回）
		if detailResp.RejectInfo.Title == "" && approval.Status == constant.TransferApprovalRejected && approval.RejectReason != "" {
			title := approvalTypeText[approval.ApprovalType]

			// 如果同一个上级门店，显示为"上级门店驳回"
			if approval.ApprovalCompanyName != "" {
				for _, otherApproval := range transferOrder.Approvals {
					if otherApproval.Uuid != approval.Uuid &&
						otherApproval.ApprovalCompanyUuid == approval.ApprovalCompanyUuid &&
						otherApproval.ApprovalType != approval.ApprovalType {
						title = "上级门店驳回"
						break
					}
				}
			}

			detailResp.RejectInfo = resp.TransferOrderRejectInfo{
				Title:        i18n.Translate(ctx.GetLanguage(), title),
				RejectReason: approval.RejectReason,
				RejectTime:   approval.ApproveTime,
			}
		}
	}

	return detailResp, nil
}

// GetTransferOrderMaterialList 获取调拨单物品列表
func (s *transferOrderSrv) GetTransferOrderMaterialList(
	ctx context.Context,
	listReq req.TransferOrderMaterialListReq,
) (material_resp.MaterialListWithPaginationResp, error) {
	if err := listReq.Validate(); err != nil {
		return material_resp.MaterialListWithPaginationResp{}, errors.WithMessage(errors.New("参数验证失败"), err.Error())
	}

	// 本店UUID
	companyUuid := ctx.GetCompanyUuid()

	// 获取本店的物品列表
	var materialListReq req.MaterialListReq
	if err := copier.Copy(&materialListReq, &listReq); err != nil {
		return material_resp.MaterialListWithPaginationResp{}, errors.WithMessage(errors.New("数据转换失败"), err.Error())
	}
	materialListReq.CategoryUuids = materialListReq.GetCategoryUuids() // 获取有效分类UUID列表。 /api/v1/shop/material/list?category_uuids&keyword&page_no=1&page_size=1000&status 时，CategoryUuids会有一个0值，是无效的
	materialListReq.PurchaseType = 2
	materialListReq.SupplierErpCode = ""
	materialListReq.WarehouseErpCode = ""
	res, err := s.materialSrv.GetMaterialList(ctx, materialListReq)
	// 处理错误
	if err != nil {
		return material_resp.MaterialListWithPaginationResp{}, errors.WithMessage(errors.New("查询调拨单物品列表失败"), err.Error())
	}

	// 调入单 - 发货门店
	if res.List != nil && len(res.List) > 0 && listReq.SenderCompanyUuid != 0 && listReq.SenderCompanyUuid != companyUuid {
		otherDb := s.dbm.GetDB(listReq.SenderCompanyUuid)
		if otherDb == nil {
			return material_resp.MaterialListWithPaginationResp{}, errors.WithMessage(errors.New("获取其他店数据库失败"), "其他店数据库不存在")
		}
		erpCodes := []string{}
		for _, item := range res.List {
			erpCodes = append(erpCodes, item.ErpCode)
		}
		// 当前库存
		warehouseItemRepo := repository.NewWarehouseItemRepo(otherDb)
		warehouseItems, err := warehouseItemRepo.GetNormalByMaterialCodes(erpCodes, listReq.OutWarehouseErpCode)
		if err != nil {
			return material_resp.MaterialListWithPaginationResp{}, errors.WithMessage(errors.New("获取物品库存失败"), err.Error())
		}
		materialList := []material_resp.Material{}
		for _, item := range res.List {
			availableNum := decimal.NewFromFloat(0)
			for _, warehouseItem := range warehouseItems {
				if item.ErpCode == warehouseItem.MaterialCode {
					availableNum = availableNum.Add(decimal.NewFromFloat(warehouseItem.Stock))
				}
			}
			item.AvailableNum = availableNum.InexactFloat64()
			materialList = append(materialList, item)
		}
		res.List = materialList
	}

	return res, nil
}

// 创建明细
func (s *transferOrderSrv) createItems(ctx context.Context, tx *gorm.DB, transferOrderUuid uint64, items []req.TransferOrderItemCreateReq, materials []model.Material) error {
	companyUuid := ctx.GetCompanyUuid()
	headquarterUuid := ctx.GetCompanySetting().HeadquarterUuid
	transferOrderItemRepoTx := repository.NewTransferOrderItemRepo(tx)
	transferOrderItemUnitRepoTx := repository.NewTransferOrderItemUnitRepo(tx)
	// 先删除旧明细
	if err := transferOrderItemRepoTx.DeleteByTransferOrderUuid(transferOrderUuid); err != nil {
		return errors.WithMessage(errors.New("删除旧明细失败"), err.Error())
	}
	if err := transferOrderItemUnitRepoTx.DeleteByTransferOrderUuid(transferOrderUuid); err != nil {
		return errors.WithMessage(errors.New("删除旧单位明细失败"), err.Error())
	}
	// 如果有明细更新
	if len(items) == 0 {
		return nil
	}
	// 创建明细
	for _, itemReq := range items {
		var material model.Material
		for _, m := range materials {
			if m.Uuid == itemReq.MaterialUuid {
				material = m
				break
			}
		}
		item := &model.TransferOrderItem{
			TransferOrderUuid:    transferOrderUuid,
			CompanyUuid:          companyUuid,
			HeadquarterUuid:      headquarterUuid,
			MaterialUuid:         material.Uuid,
			MaterialCode:         material.Code,
			MaterialName:         material.Name,
			MaterialInternalCode: material.InternalCode,
			Valuation:            material.GetValuation(),
		}

		if err := transferOrderItemRepoTx.Create(item); err != nil {
			return errors.WithMessage(errors.New("创建调拨单明细失败"), err.Error())
		}

		// 创建单位明细
		for _, unitReq := range itemReq.Units {
			materialUnit := &model.MaterialUnit{}
			for _, unit := range material.NotBaseUnitList {
				if unit.Uuid == unitReq.UnitUuid {
					materialUnit = unit
					break
				}
			}
			if materialUnit.Uuid == 0 {
				return errors.WithMessage(errors.New("物品单位不存在"), fmt.Sprintf("物品%s的单位%d不存在", material.Code, unitReq.UnitUuid))
			}
			unit := &model.TransferOrderItemUnit{
				ItemUuid:           item.Uuid,
				TransferOrderUuid:  transferOrderUuid,
				UnitUuid:           unitReq.UnitUuid,
				Num:                unitReq.Num,
				UnitName:           materialUnit.Name,
				UnitConversionRate: materialUnit.ConversionRate,
				ErpnextUom: func() string {
					if materialUnit.Unit != nil {
						return materialUnit.Unit.ErpnextUom
					}
					return ""
				}(),
			}

			if err := transferOrderItemUnitRepoTx.Create(unit); err != nil {
				return errors.WithMessage(errors.New("创建调拨单单位明细失败"), err.Error())
			}
		}
	}
	return nil
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
	companyUuid := ctx.GetCompanyUuid()
	headquarterUuid := ctx.GetCompanySetting().HeadquarterUuid

	// 加锁防止并发创建（使用字符串锁保护编号生成）
	lockKey := fmt.Sprintf("transfer_order_create_%d_%d", companyUuid, req.TransferType)
	s.lock.LockUuidString(lockKey)
	defer s.lock.UnlockUuidString(lockKey)

	// 设置发货门店和收货门店
	if req.TransferType == 1 {
		req.ReceiverCompanyUuid = companyUuid
	} else {
		req.SenderCompanyUuid = companyUuid
	}

	// 判断发货门店和收货门店是否存在
	senderCompany := model.Company{}
	if req.SenderCompanyUuid != 0 {
		company, err := repository.NewCompanyRepo(s.dbm.GetDB(req.SenderCompanyUuid)).GetCompany()
		if err != nil {
			return resp.TransferOrderCreateResp{}, errors.WithMessage(errors.New("查询发货门店失败"), err.Error())
		}
		senderCompany = company
	}
	receiverCompany := model.Company{}
	if req.ReceiverCompanyUuid != 0 {
		company, err := repository.NewCompanyRepo(s.dbm.GetDB(req.ReceiverCompanyUuid)).GetCompany()
		if err != nil {
			return resp.TransferOrderCreateResp{}, errors.WithMessage(errors.New("查询收货门店失败"), err.Error())
		}
		receiverCompany = company
	}

	// 判断仓库是否存在
	warehouseRepo := repository.NewWarehouseRepo(db)
	outWarehouse := &model.Warehouse{}
	if req.OutWarehouseErpCode != "" {
		warehouse, err := warehouseRepo.GetByErpCode(req.OutWarehouseErpCode)
		if err != nil {
			return resp.TransferOrderCreateResp{}, errors.WithMessage(errors.New("入库仓库不存在"), err.Error())
		}
		if warehouse == nil {
			return resp.TransferOrderCreateResp{}, errors.New("出库仓库不存在")
		}
		outWarehouse = warehouse
	}

	// 判断入库仓库是否存在
	inWarehouse := &model.Warehouse{}
	if req.InWarehouseErpCode != "" {
		warehouse, err := warehouseRepo.GetByErpCode(req.InWarehouseErpCode)
		if err != nil {
			return resp.TransferOrderCreateResp{}, errors.WithMessage(errors.New("入库仓库不存在"), err.Error())
		}
		if warehouse == nil {
			return resp.TransferOrderCreateResp{}, errors.New("入库仓库不存在")
		}
		inWarehouse = warehouse
	}

	// 验证物品状态
	materials, _, err := s.validator.validateMaterialStatus(ctx, db, req.Items, true)
	if err != nil {
		return resp.TransferOrderCreateResp{}, err
	}

	// 生成调拨单编号
	orderNo := s.helper.GenerateOrderNo(db)

	// 生成调拨单UUID
	transferOrderUuid, err := utils.GetID()
	if err != nil {
		return resp.TransferOrderCreateResp{}, errors.WithMessage(errors.New("生成调拨单UUID失败"), err.Error())
	}

	// 开始事务
	err = db.Transaction(func(tx *gorm.DB) error {
		// 创建调拨单主表
		transferOrder := &model.TransferOrder{
			BaseModel: model.BaseModel{
				Uuid: transferOrderUuid,
			},
			CompanyUuid:         companyUuid,
			HeadquarterUuid:     headquarterUuid,
			OrderNo:             orderNo,
			TransferType:        req.TransferType,
			SenderCompanyUuid:   req.SenderCompanyUuid,
			SenderCompanyName:   senderCompany.Name,
			ReceiverCompanyUuid: req.ReceiverCompanyUuid,
			ReceiverCompanyName: receiverCompany.Name,
			OutWarehouseErpCode: req.OutWarehouseErpCode,
			OutWarehouseName:    outWarehouse.Name,
			InWarehouseErpCode:  req.InWarehouseErpCode,
			InWarehouseName:     inWarehouse.Name,
			OrderTime:           req.OrderTime,
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
		if err := s.createItems(ctx, tx, transferOrder.Uuid, req.Items, materials); err != nil {
			return errors.WithMessage(errors.New("创建调拨单明细失败"), err.Error())
		}

		// 记录操作日志
		if err := s.helper.CreateLog(ctx, db, transferOrder.Uuid, constant.TransferActionCreate, "创建调拨单", 0, constant.TransferOrderStatusDraft); err != nil {
			logger.Logger.Error("记录调拨单日志失败", zap.Error(err))
		}

		return nil
	})

	if err != nil {
		return resp.TransferOrderCreateResp{}, err
	}

	return resp.TransferOrderCreateResp{
		Uuid:    transferOrderUuid,
		OrderNo: orderNo,
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

	// 加锁
	s.lock.LockUuid(req.Uuid)
	defer s.lock.UnlockUuid(req.Uuid)

	db := ctx.GetDB()
	companyUuid := ctx.GetCompanyUuid()
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

	if transferOrder.TransferType == 1 {
		if req.SenderCompanyUuid == 0 {
			return errors.New("发货门店不能为空")
		}
		if req.InWarehouseErpCode == "" {
			return errors.New("出库仓库不能为空")
		}
	} else {
		if req.ReceiverCompanyUuid == 0 {
			return errors.New("收货门店不能为空")
		}
		if req.OutWarehouseErpCode == "" {
			return errors.New("入库仓库不能为空")
		}
	}

	// 判断发货门店和收货门店是否存在
	senderCompany := model.Company{}
	if req.SenderCompanyUuid != 0 {
		company, err := repository.NewCompanyRepo(s.dbm.GetDB(req.SenderCompanyUuid)).GetCompany()
		if err != nil {
			return errors.WithMessage(errors.New("查询发货门店失败"), err.Error())
		}
		senderCompany = company
	}
	receiverCompany := model.Company{}
	if req.ReceiverCompanyUuid != 0 {
		company, err := repository.NewCompanyRepo(s.dbm.GetDB(req.ReceiverCompanyUuid)).GetCompany()
		if err != nil {
			return errors.WithMessage(errors.New("查询收货门店失败"), err.Error())
		}
		receiverCompany = company
	}

	// 判断仓库是否存在
	warehouseRepo := repository.NewWarehouseRepo(db)
	outWarehouse := &model.Warehouse{}
	if req.OutWarehouseErpCode != "" {
		warehouse, err := warehouseRepo.GetByErpCode(req.OutWarehouseErpCode)
		if err != nil {
			return errors.WithMessage(errors.New("入库仓库不存在"), err.Error())
		}
		if warehouse == nil {
			return errors.New("出库仓库不存在")
		}
		outWarehouse = warehouse
	}

	// 判断入库仓库是否存在
	inWarehouse := &model.Warehouse{}
	if req.InWarehouseErpCode != "" {
		warehouse, err := warehouseRepo.GetByErpCode(req.InWarehouseErpCode)
		if err != nil {
			return errors.WithMessage(errors.New("入库仓库不存在"), err.Error())
		}
		if warehouse == nil {
			return errors.New("入库仓库不存在")
		}
		inWarehouse = warehouse
	}

	// 验证物品状态
	materials, _, err := s.validator.validateMaterialStatus(ctx, db, req.Items, true)
	if err != nil {
		return err
	}

	// 设置发货门店和收货门店
	if transferOrder.TransferType == 1 {
		req.ReceiverCompanyUuid = companyUuid
	} else {
		req.SenderCompanyUuid = companyUuid
	}

	// 开始事务
	err = db.Transaction(func(tx *gorm.DB) error {
		// 更新主表
		transferOrder.OrderTime = req.OrderTime
		transferOrder.ItemCount = len(req.Items)
		transferOrder.SenderCompanyUuid = req.SenderCompanyUuid
		transferOrder.SenderCompanyName = senderCompany.Name
		transferOrder.ReceiverCompanyUuid = req.ReceiverCompanyUuid
		transferOrder.ReceiverCompanyName = receiverCompany.Name
		transferOrder.OutWarehouseErpCode = req.OutWarehouseErpCode
		transferOrder.OutWarehouseName = outWarehouse.Name
		transferOrder.InWarehouseErpCode = req.InWarehouseErpCode
		transferOrder.InWarehouseName = inWarehouse.Name
		transferOrder.Remark = req.Remark
		if err := transferOrderRepo.Update(transferOrder); err != nil {
			return errors.WithMessage(errors.New("更新调拨单失败"), err.Error())
		}
		// 创建调拨单明细
		if err := s.createItems(ctx, tx, transferOrder.Uuid, req.Items, materials); err != nil {
			return errors.WithMessage(errors.New("创建调拨单明细失败"), err.Error())
		}
		return nil
	})

	if err != nil {
		return err
	}

	// 记录操作日志
	if err := s.helper.CreateLog(ctx, db, req.Uuid, "update", "更新调拨单", transferOrder.Status, transferOrder.Status); err != nil {
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

	// 加锁
	s.lock.LockUuid(req.Uuid)
	defer s.lock.UnlockUuid(req.Uuid)

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

	// 删除调拨单
	if err := transferOrderRepo.Delete(req.Uuid); err != nil {
		return errors.WithMessage(errors.New("删除调拨单失败"), err.Error())
	}

	// 记录操作日志
	if err := s.helper.CreateLog(ctx, db, req.Uuid, constant.TransferActionDelete, "删除调拨单", constant.TransferOrderStatusDraft, constant.TransferOrderStatusDraft); err != nil {
		logger.Logger.Error("记录调拨单日志失败", zap.Error(err))
	}

	return nil
}

// SubmitTransferOrder 提交调拨单
func (s *transferOrderSrv) SubmitTransferOrder(
	ctx context.Context,
	req req.TransferOrderSubmitReq,
) error {
	if err := req.Validate(); err != nil {
		return err
	}

	// 加锁
	s.lock.LockUuid(req.Uuid)
	defer s.lock.UnlockUuid(req.Uuid)

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

	// 开始事务
	err = db.Transaction(func(tx *gorm.DB) error {

		// 创建审批流程
		if err := s.helper.CreateApproval(ctx, tx, s.dbm, transferOrder); err != nil {
			return errors.WithMessage(errors.New("创建审批失败"), err.Error())
		}

		// 更新状态为待审核
		updates := map[string]interface{}{
			"status":                     constant.TransferOrderStatusPending,
			"submit_time":                time.Now().Unix(),
			"next_approval_company_uuid": ctx.GetCompanyUuid(),
			"next_approval_company_name": ctx.GetCompany().Name,
		}

		if err := tx.Model(&model.TransferOrder{}).Where("uuid = ?", req.Uuid).Updates(updates).Error; err != nil {
			return errors.WithMessage(errors.New("提交调拨单失败"), err.Error())
		}

		// 记录操作日志
		if err := s.helper.CreateLog(ctx, db, req.Uuid, constant.TransferActionSubmit, "提交调拨单", constant.TransferOrderStatusDraft, constant.TransferOrderStatusPending); err != nil {
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

	// 加锁
	s.lock.LockUuid(req.Uuid)
	defer s.lock.UnlockUuid(req.Uuid)

	// 获取调拨单数据库
	db, err := s.helper.GetOrderDb(ctx, s.dbm, req.Uuid)
	if err != nil {
		return err
	}

	transferOrderRepo := repository.NewTransferOrderRepo(db)
	approvalRepo := repository.NewTransferOrderApprovalRepo(db)

	// 查询调拨单
	transferOrder, err := transferOrderRepo.GetByUuid(req.Uuid, transferOrderRepo.WithItems())
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

	if transferOrder.NextApprovalCompanyUuid != ctx.GetCompanyUuid() {
		return errors.New("无审批权限")
	}

	// 获取当前审批节点
	currentApproval, err := approvalRepo.GetCurrentApproval(req.Uuid, transferOrder.NextApprovalCompanyUuid)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("无审批权限")
		}
		return errors.WithMessage(errors.New("查询审批节点失败"), err.Error())
	}
	if currentApproval == nil {
		return errors.New("无审批权限")
	}

	// 如果入库仓库为空，且当前审批节点为发货门店上级，则无审批权限
	if transferOrder.OutWarehouseErpCode == "" && currentApproval.ApprovalType == constant.TransferApprovalTypeSender {
		return errors.New("请选择出库仓库")
	}

	// 如果入库仓库为空，且当前审批节点为收货门店上级，则无审批权限
	if transferOrder.InWarehouseErpCode == "" && currentApproval.ApprovalType == constant.TransferApprovalTypeReceiver {
		return errors.New("请选择入库仓库")
	}

	// 开始事务
	staff := ctx.GetStaff()
	err = db.Transaction(func(tx *gorm.DB) error {
		transferOrderRepoTx := repository.NewTransferOrderRepo(tx)

		// 更新当前审批节点为已通过
		approvalRepoTx := repository.NewTransferOrderApprovalRepo(tx)
		currentApproval.Status = constant.TransferApprovalApproved
		currentApproval.ApproverUuid = staff.Uuid
		currentApproval.ApproverName = staff.GetUserName()
		currentApproval.ApproveTime = time.Now().Unix()
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
		newStatus := transferOrder.Status

		if nextApproval != nil {
			transferOrder.NextApprovalCompanyUuid = nextApproval.ApprovalCompanyUuid
			transferOrder.NextApprovalCompanyName = nextApproval.ApprovalCompanyName
		} else {
			transferOrder.Status = constant.TransferOrderStatusReceiving
			transferOrder.NextApprovalCompanyUuid = 0
			transferOrder.NextApprovalCompanyName = ""
			newStatus = constant.TransferOrderStatusReceiving
		}

		if err := transferOrderRepoTx.Update(transferOrder); err != nil {
			return errors.WithMessage(errors.New("更新调拨单状态失败"), err.Error())
		}

		// 调用erp接口
		if ctx.GetCompany().IsOpenErp() && newStatus == constant.TransferOrderStatusReceiving {
			erpResp, err := s.helper.SaveMaterialTransfer(ctx, s.dbm, tx, transferOrder)
			if err != nil {
				return errors.WithMessage(errors.New("调用erp接口失败"), err.Error())
			}
			// 更新调拨单状态
			transferOrder.ErpOrderNo = erpResp.FromReceipt.SoNo
			transferOrder.ErpResp = utils.ToJson(erpResp)
			if err := transferOrderRepoTx.Update(transferOrder); err != nil {
				return errors.WithMessage(errors.New("更新调拨单ERP响应数据失败"), err.Error())
			}
		}

		// 复制数据到总部
		if err := s.helper.CopyDataToHeadquarter(ctx, s.dbm, tx, req.Uuid); err != nil {
			return errors.WithMessage(errors.New("复制数据到总部失败"), err.Error())
		}

		// 记录操作日志
		if err := s.helper.CreateLog(ctx, db, req.Uuid, constant.TransferActionApprove, fmt.Sprintf("%s审批通过", currentApproval.ApprovalCompanyName), transferOrder.Status, newStatus); err != nil {
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

	// 加锁
	s.lock.LockUuid(req.Uuid)
	defer s.lock.UnlockUuid(req.Uuid)

	// 获取调拨单数据库
	db, err := s.helper.GetOrderDb(ctx, s.dbm, req.Uuid)
	if err != nil {
		return err
	}

	// 获取调拨单仓库
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

	if transferOrder.NextApprovalCompanyUuid != ctx.GetCompanyUuid() {
		return errors.New("无审批权限")
	}

	// 获取当前审批节点
	currentApproval, err := approvalRepo.GetCurrentApproval(req.Uuid, ctx.GetCompanyUuid())
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("无审批权限")
		}
		return errors.WithMessage(errors.New("查询审批节点失败"), err.Error())
	}
	if currentApproval == nil {
		return errors.New("无审批权限")
	}

	// 开始事务
	staff := ctx.GetStaff()
	err = db.Transaction(func(tx *gorm.DB) error {
		// 更新当前审批节点为已驳回
		approvalRepoTx := repository.NewTransferOrderApprovalRepo(tx)
		currentApproval.Status = constant.TransferApprovalRejected
		currentApproval.ApproverUuid = staff.Uuid
		currentApproval.ApproverName = staff.GetUserName()
		currentApproval.ApproveTime = time.Now().Unix()
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

		// 复制数据到总部
		if err := s.helper.CopyDataToHeadquarter(ctx, s.dbm, tx, req.Uuid); err != nil {
			return errors.WithMessage(errors.New("复制数据到总部失败"), err.Error())
		}

		// 记录操作日志
		if err := s.helper.CreateLog(ctx, db, req.Uuid, constant.TransferActionReject, fmt.Sprintf("%s驳回：%s", currentApproval.ApprovalCompanyName, req.RejectReason), transferOrder.Status, constant.TransferOrderStatusRejected); err != nil {
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

	// 加锁
	s.lock.LockUuid(req.Uuid)
	defer s.lock.UnlockUuid(req.Uuid)

	// 获取调拨单数据库
	db, err := s.helper.GetOrderDb(ctx, s.dbm, req.Uuid)
	if err != nil {
		return err
	}

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

	// 开始事务
	err = db.Transaction(func(tx *gorm.DB) error {
		// 更新调拨单为已完成状态
		updates := map[string]interface{}{
			"status": constant.TransferOrderStatusCompleted,
		}

		if err := tx.Model(&model.TransferOrder{}).Where("uuid = ?", req.Uuid).Updates(updates).Error; err != nil {
			return errors.WithMessage(errors.New("更新调拨单状态失败"), err.Error())
		}

		// 复制数据到总部
		if err := s.helper.CopyDataToHeadquarter(ctx, s.dbm, tx, req.Uuid); err != nil {
			return errors.WithMessage(errors.New("复制数据到总部失败"), err.Error())
		}

		// TODO: 这里需要调用库存服务，处理库存入库逻辑

		// 记录操作日志
		if err := s.helper.CreateLog(ctx, db, req.Uuid, constant.TransferActionReceive, "收货完成", transferOrder.Status, constant.TransferOrderStatusCompleted); err != nil {
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
	// 获取调拨单数据库
	db, err := s.helper.GetOrderDb(ctx, s.dbm, req.TransferOrderUuid)
	if err != nil {
		return resp.TransferOrderApprovalListResp{}, err
	}

	// 获取审批流程列表
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
	// 获取调拨单数据库
	db, err := s.helper.GetOrderDb(ctx, s.dbm, req.TransferOrderUuid)
	if err != nil {
		return resp.TransferOrderLogListResp{}, err
	}

	// 获取操作日志列表
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

// GetTransferOrderCompanyList 获取调拨单门店列表
func (s *transferOrderSrv) GetTransferOrderCompanyList(
	ctx context.Context,
) (resp.TransferOrderCompanyListResp, error) {
	db := s.dbm.GetDB(0)
	companyRepo := repository.NewCompanyRepo(db)
	currentCompanyUuid := ctx.GetCompanyUuid()
	companySetting := ctx.GetCompanySetting()

	// 获取总部下的所有门店
	headquarterUuid := companySetting.HeadquarterUuid
	if headquarterUuid == 0 {
		headquarterUuid = currentCompanyUuid
	}

	companies, err := companyRepo.GetNoDeleteListByHeadquarterUuid(headquarterUuid)
	if err != nil {
		return resp.TransferOrderCompanyListResp{}, errors.WithMessage(errors.New("获取总部下的所有门店失败"), err.Error())
	}

	// 转换响应数据
	list := make([]resp.TransferOrderCompanyItem, 0)
	for _, company := range companies {
		if company.Status == 0 {
			continue
		}
		if company.Uuid == currentCompanyUuid {
			continue
		}
		list = append(list, resp.TransferOrderCompanyItem{
			Uuid:          company.Uuid,
			Name:          company.Name,
			IsHeadquarter: company.Uuid == headquarterUuid,
		})
	}

	return resp.TransferOrderCompanyListResp{
		List: list,
	}, nil
}

// GetTransferOrderWarehouseList 获取调拨单仓库列表
func (s *transferOrderSrv) GetTransferOrderWarehouseList(
	ctx context.Context,
) (resp.TransferOrderWarehouseListResp, error) {
	// 获取目标公司的数据库连接
	targetDb := ctx.GetDB()
	warehouseRepoTarget := repository.NewWarehouseRepo(targetDb)
	warehouses, err := warehouseRepoTarget.GetNormalWarehouse()
	if err != nil {
		return resp.TransferOrderWarehouseListResp{}, errors.WithMessage(errors.New("查询仓库列表失败"), err.Error())
	}

	// 转换响应数据
	list := make([]resp.TransferOrderWarehouseItem, 0, len(warehouses))
	for _, warehouse := range warehouses {
		item := resp.TransferOrderWarehouseItem{
			ErpCode: warehouse.ErpCode,
			Type:    warehouse.Type,
			Name: func() dto.LocaleResponse {
				var localName dto.LocaleResponse
				if warehouse.MultiLanguageName != nil {
					localName = warehouse.MultiLanguageName.GetNames()
				} else {
					localName = *language.JsonToLocaleResponse(warehouse.Name)
				}
				return localName
			}(),
		}
		list = append(list, item)
	}

	return resp.TransferOrderWarehouseListResp{
		List: list,
	}, nil
}
