package service

import (
	"fmt"
	"strconv"
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
	"ttpos-server-go/pkg/utils"

	"github.com/jinzhu/copier"
	"github.com/shopspring/decimal"
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

	// 排序
	opts = append(opts, purchaseOrderRepo.OrderByOrderTime(true))
	if req.PurchaseType == 2 {
		opts = append(opts, purchaseOrderRepo.WherePurchaseType(req.PurchaseType))
	} else {
		opts = append(opts, purchaseOrderRepo.WherePurchaseType(1))
	}

	// WithItems()
	opts = append(opts, purchaseOrderRepo.WithItems())

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
		poInfo.ReceiptProgress = fmt.Sprintf("%.0f%%", po.GetReceiptProgress())
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
func (s *purchaseOrderSrv) CreatePurchaseOrder(ctx context.Context, req req.PurchaseOrderCreateReq) (resp.PurchaseOrderCreateResp, error) {
	if err := req.Validate(); err != nil {
		return resp.PurchaseOrderCreateResp{}, err
	}

	if ctx.Version(context.GTE, "2.6.0") {
		if req.SupplierErpCode == "" {
			return resp.PurchaseOrderCreateResp{}, errors.New("供应商编码不能为空")
		}
		if req.PurchaseType == 2 && req.WarehouseErpCode == "" {
			return resp.PurchaseOrderCreateResp{}, errors.New("仓库编码不能为空")
		}
	}

	db := s.dbm.GetDB(ctx.GetDbId())

	var result resp.PurchaseOrderCreateResp
	err := db.Transaction(func(tx *gorm.DB) error {
		purchaseOrderRepo := repository.NewPurchaseOrderRepo(tx)
		purchaseOrderItemRepo := repository.NewPurchaseOrderItemRepo(tx)
		// 创建采购申请
		purchaseOrder := &model.PurchaseOrder{
			OrderNo:           s.generateOrderNo(ctx, db),
			SupplierName:      req.SupplierName,
			SupplierErpCode:   utils.IfString(req.SupplierErpCode != "", req.SupplierErpCode, req.SupplierName),
			Status:            constant.PurchaseOrderStatusDraft, // 待提交状态
			Num:               float64(len(req.Items)),
			OrderTime:         req.OrderTime,
			ExpectArrivalTime: req.ExpectedDeliveryTime,
			ApplicantUuid:     ctx.GetStaffUuid(),
			ApplicantName:     ctx.GetStaff().RealName,
			PurchaseType: func() int {
				if req.PurchaseType == 2 {
					return 2
				}
				return 1
			}(),
			WarehouseErpCode: req.WarehouseErpCode,
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
		materialList, err := materialRepo.GetMaterialByUuids(
			materialUuids,
			materialRepo.WithPreload("Unit"),
			materialRepo.WithPreload("PurchaseUnit"),
		)
		if err != nil {
			return errors.WithMessage(err, "查询物品失败")
		}

		// 将切片转换为map，方便后续查找
		materials := make(map[uint64]*model.Material)
		for _, material := range materialList {
			materials[material.Uuid] = material
		}

		// 验证所有请求的物品都存在
		for _, itemReq := range req.Items {
			if _, exists := materials[itemReq.MaterialUuid]; !exists {
				return errors.New(fmt.Sprintf("物品UUID%d不存在", itemReq.MaterialUuid))
			}
		}

		// 创建采购申请明细
		var items []model.PurchaseOrderItem
		for _, itemReq := range req.Items {
			material := materials[itemReq.MaterialUuid]
			item := model.PurchaseOrderItem{
				PurchaseOrderUuid: purchaseOrder.Uuid,
				MaterialUuid:      itemReq.MaterialUuid,
				Num:               itemReq.Num,
				MaterialCode:      material.Code,
				MaterialName:      material.Name,
				UnitUuid:          material.PurchaseUnitUuid,
				UnitName: func() string {
					if material.PurchaseUnit != nil {
						return material.PurchaseUnit.Name
					}
					return ""
				}(),
				UnitConversionRate: func() float64 {
					if material.PurchaseUnit != nil {
						return material.PurchaseUnit.ConversionRate
					}
					return 0
				}(),
				BaseUnitUuid: func() uint64 {
					if material.Unit != nil {
						return material.Unit.UnitUuid
					}
					return 0
				}(),
				BaseUnitName: func() string {
					if material.Unit != nil {
						return material.Unit.Name
					}
					return ""
				}(),
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
func (s *purchaseOrderSrv) UpdatePurchaseOrder(ctx context.Context, req req.PurchaseOrderUpdateReq) error {
	if err := req.Validate(); err != nil {
		return err
	}

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

		if ctx.Version(context.GTE, "2.6.0") {
			if req.SupplierErpCode == "" {
				return errors.New("供应商编码不能为空")
			}
			if purchaseOrder.PurchaseType == 2 && req.WarehouseErpCode == "" {
				return errors.New("仓库编码不能为空")
			}
		}

		// 更新采购申请基本信息
		purchaseOrder.Num = float64(len(req.Items))
		purchaseOrder.SupplierName = req.SupplierName
		purchaseOrder.SupplierErpCode = req.SupplierErpCode
		purchaseOrder.ExpectArrivalTime = req.ExpectedDeliveryTime
		purchaseOrder.WarehouseErpCode = req.WarehouseErpCode
		err = purchaseOrderRepo.Update(purchaseOrder)
		if err != nil {
			return errors.WithMessage(err, "更新采购申请失败")
		}

		materialUuids := make([]uint64, 0)
		for _, itemReq := range req.Items {
			materialUuids = append(materialUuids, itemReq.MaterialUuid)
		}
		materialRepo := base.NewMaterialRepo(db)
		materialList, err := materialRepo.GetMaterialByUuids(
			materialUuids,
			materialRepo.WithPreload("Unit"),
			materialRepo.WithPreload("PurchaseUnit"),
		)
		if err != nil {
			return errors.WithMessage(err, "查询物品失败")
		}

		// 将切片转换为map，方便后续查找
		materials := make(map[uint64]*model.Material)
		for _, material := range materialList {
			materials[material.Uuid] = material
		}

		// 验证所有请求的物品都存在
		for _, itemReq := range req.Items {
			if _, exists := materials[itemReq.MaterialUuid]; !exists {
				return errors.New(fmt.Sprintf("物品UUID %d 不存在", itemReq.MaterialUuid))
			}
		}

		// 创建采购申请明细
		var items []model.PurchaseOrderItem
		for _, itemReq := range req.Items {
			material := materials[itemReq.MaterialUuid]
			item := model.PurchaseOrderItem{
				PurchaseOrderUuid: purchaseOrder.Uuid,
				MaterialUuid:      itemReq.MaterialUuid,
				Num:               itemReq.Num,
				MaterialCode:      material.Code,
				MaterialName:      material.Name,
				UnitUuid:          material.PurchaseUnitUuid,
				UnitName: func() string {
					if material.PurchaseUnit != nil {
						return material.PurchaseUnit.Name
					}
					return ""
				}(),
				UnitConversionRate: func() float64 {
					if material.PurchaseUnit != nil {
						return material.PurchaseUnit.ConversionRate
					}
					return 0
				}(),
				BaseUnitUuid: func() uint64 {
					if material.Unit != nil {
						return material.Unit.UnitUuid
					}
					return 0
				}(),
				BaseUnitName: func() string {
					if material.Unit != nil {
						return material.Unit.Name
					}
					return ""
				}(),
			}
			items = append(items, item)
		}

		// 先删除所有现有明细项
		err = purchaseOrderItemRepo.DeleteByPurchaseOrderUuid(req.Uuid)
		if err != nil {
			return errors.WithMessage(err, "删除采购申请明细失败")
		}

		// 批量创建采购申请明细
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
		return err
	}

	return nil
}

// DeletePurchaseOrder 删除采购申请
func (s *purchaseOrderSrv) DeletePurchaseOrder(ctx context.Context, req req.PurchaseOrderDeleteReq) error {
	db := s.dbm.GetDB(ctx.GetDbId())
	purchaseOrderRepo := repository.NewPurchaseOrderRepo(db)

	// 查询采购申请
	purchaseOrder, err := purchaseOrderRepo.GetByUuid(req.Uuid)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("采购申请不存在")
		}
		return errors.WithMessage(err, "查询采购申请失败")
	}

	// 检查是否可删除（只有待提交和已驳回状态可以删除）
	if !purchaseOrder.IsEditable() {
		return errors.New("当前状态不允许删除")
	}

	// 删除采购申请（软删除）
	err = purchaseOrderRepo.Delete(req.Uuid)
	if err != nil {
		return errors.WithMessage(err, "删除采购申请失败")
	}

	return nil
}

// SubmitPurchaseOrder 提交采购申请
func (s *purchaseOrderSrv) SubmitPurchaseOrder(ctx context.Context, req req.PurchaseOrderSubmitReq) error {
	return s.dbm.GetDB(ctx.GetDbId()).Transaction(func(tx *gorm.DB) error {
		purchaseOrderRepo := repository.NewPurchaseOrderRepo(tx)
		purchaseOrderItemRepo := repository.NewPurchaseOrderItemRepo(tx)

		// 查询采购申请
		purchaseOrder, err := purchaseOrderRepo.GetByUuid(req.Uuid, purchaseOrderRepo.WithItems())
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return errors.New("采购申请不存在")
			}
			return errors.WithMessage(err, "查询采购申请失败")
		}

		// 删除物品为0的数据
		err = purchaseOrderItemRepo.DeleteByPurchaseOrderUuidAndNumIsZero(req.Uuid)
		if err != nil {
			return errors.WithMessage(err, "删除采购申请明细失败")
		}

		// 过滤掉数量为0的项目后重新计算数量
		validItems := make([]model.PurchaseOrderItem, 0)
		for _, item := range purchaseOrder.Items {
			if item.Num > 0 {
				validItems = append(validItems, item)
			}
		}

		oldStatus := purchaseOrder.Status
		purchaseOrder.Status = constant.PurchaseOrderStatusPending
		purchaseOrder.OrderTime = time.Now().Unix()
		purchaseOrder.ApplicantUuid = ctx.GetStaffUuid()
		purchaseOrder.ApplicantName = ctx.GetStaff().RealName
		purchaseOrder.Num = float64(len(validItems)) // 使用过滤后的数量

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
func (s *purchaseOrderSrv) ApprovePurchaseOrder(ctx context.Context, req req.PurchaseOrderApproveReq) error {
	db := s.dbm.GetDB(ctx.GetDbId())

	err := db.Transaction(func(tx *gorm.DB) error {
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

		// 更新状态
		purchaseOrder.Status = newStatus

		// 更新总部状态
		if purchaseOrder.PurchaseType == 2 && newStatus == constant.PurchaseOrderStatusApproved {
			purchaseOrder.Status = constant.PurchaseOrderStatusHeadquarterPending
			purchaseOrder.HeadquarterStatus = constant.HeadquarterStatusPending
		}

		// 更新状态
		err = purchaseOrderRepo.Update(purchaseOrder)
		if err != nil {
			return errors.WithMessage(err, "更新采购申请状态失败")
		}

		// 记录操作日志
		err = s.createPurchaseOrderLog(tx, req.Uuid, ctx, req.Action, actionDesc, oldStatus, newStatus, "")
		if err != nil {
			return errors.WithMessage(err, "记录操作日志失败")
		}

		// 内部采购 不调用erp接口
		if purchaseOrder.PurchaseType == 2 {
			// 获取总部数据库
			db = s.dbm.GetDB(ctx.GetCompanySetting().HeadquarterUuid)
			if db == nil {
				return errors.New("获取总部数据库失败")
			}
			// 整单复制到总部采购申请
			headquarterPurchaseOrder := &model.PurchaseOrder{
				BaseModel: model.BaseModel{
					Uuid: func() uint64 {
						uuid, _ := utils.GetID()
						return uuid
					}(),
				},
				SubUuid:           purchaseOrder.Uuid,
				OrderNo:           purchaseOrder.OrderNo,
				SupplierName:      purchaseOrder.SupplierName,
				SupplierErpCode:   purchaseOrder.SupplierErpCode,
				Status:            constant.PurchaseOrderStatusPending,
				HeadquarterStatus: constant.HeadquarterStatusPending,
				OrderTime:         purchaseOrder.OrderTime,
				ExpectArrivalTime: purchaseOrder.ExpectArrivalTime,
				WarehouseErpCode:  purchaseOrder.WarehouseErpCode,
				PurchaseType:      purchaseOrder.PurchaseType,
				ApplicantUuid:     purchaseOrder.ApplicantUuid,
				ApplicantName:     purchaseOrder.ApplicantName,
				ApproverUuid:      purchaseOrder.ApproverUuid,
				ApproverName:      purchaseOrder.ApproverName,
				Num:               purchaseOrder.Num,
				OrderType:         purchaseOrder.OrderType,
				CompanyUuid:       purchaseOrder.CompanyUuid,
				CompanyName:       purchaseOrder.CompanyName,
			}
			err = repository.NewPurchaseOrderRepo(db).Create(headquarterPurchaseOrder)
			if err != nil {
				return errors.WithMessage(err, "创建总部采购申请失败")
			}

			// 创建总部采购申请明细
			var headquarterItems []model.PurchaseOrderItem
			for _, item := range purchaseOrder.Items {
				headquarterItem := model.PurchaseOrderItem{
					PurchaseOrderUuid:  headquarterPurchaseOrder.Uuid,
					MaterialCode:       item.MaterialCode,
					MaterialName:       item.MaterialName,
					MaterialUuid:       item.MaterialUuid,
					Num:                item.Num,
					ArrivalNum:         item.ArrivalNum,
					UnitUuid:           item.UnitUuid,
					UnitName:           item.UnitName,
					UnitConversionRate: item.UnitConversionRate,
					BaseUnitUuid:       item.BaseUnitUuid,
					BaseUnitName:       item.BaseUnitName,
				}
				headquarterItems = append(headquarterItems, headquarterItem)
			}
			err = repository.NewPurchaseOrderItemRepo(db).CreateBatch(headquarterItems)
			if err != nil {
				return errors.WithMessage(err, "创建总部采购申请明细失败")
			}
		}

		if ctx.GetCompany().IsOpenErp() && purchaseOrder.Status == constant.PurchaseOrderStatusApproved {
			stockItems := make([]*stock.MaterialRequestItem, 0)
			for _, item := range purchaseOrder.Items {
				materialUnit, err := repository.NewMaterialUnitRepo(tx).GetMaterialUnitsByUuid(item.UnitUuid)
				if err != nil {
					return errors.WithMessage(err, "查询物品单位失败")
				}
				if materialUnit.Unit == nil {
					return errors.New("查询物品原始单位失败")
				}
				stockItems = append(stockItems, &stock.MaterialRequestItem{
					ItemCode:     item.MaterialCode,
					Qty:          item.Num,
					ScheduleDate: purchaseOrder.ExpectArrivalTime,
					Uom:          materialUnit.Unit.ErpnextUom,
				})
			}
			resp, err := erp.NewIErpSrv(s.dbm).CreatePurchaseOrder(ctx, &stock.SaveMaterialRequestReq{
				TransactionDate: purchaseOrder.OrderTime,
				RequiredBy:      purchaseOrder.ExpectArrivalTime,
				Supplier:        purchaseOrder.SupplierName,
				Items:           stockItems,
			})
			if err != nil {
				return errors.WithMessage(err, "调用erp接口失败")
			}
			// 更新采购申请单号
			purchaseOrder.ErpOrderNo = resp.PurchaseOrder
			err = purchaseOrderRepo.Update(purchaseOrder)
			if err != nil {
				return errors.WithMessage(err, "更新采购申请单号失败")
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

// CreateReceiptOrder 创建收货单
func (s *purchaseOrderSrv) CreatePurchaseReceiptOrder(ctx context.Context, req req.PurchaseReceiptCreateReq) (resp.PurchaseReceiptOrderCreateResp, error) {
	db := s.dbm.GetDB(ctx.GetDbId())

	// 判断物品明细是否已经停用
	if req.IsConfirm {
		var disabledMaterials []string
		for _, item := range req.Items {
			purchaseOrderItem, err := repository.NewPurchaseOrderItemRepo(db).GetByUuid(item.PurchaseOrderItemUuid)
			if err != nil {
				return resp.PurchaseReceiptOrderCreateResp{}, errors.WithMessage(err, "查询采购申请明细失败")
			}
			material, err := repository.NewMaterialRepo(db).GetMaterialByUuid(purchaseOrderItem.MaterialUuid)
			if err != nil {
				return resp.PurchaseReceiptOrderCreateResp{}, errors.WithMessage(err, "查询物品明细失败")
			}
			// 判断物品是否停用
			if material.Status == false {
				materialName := language.JsonToLocaleResponse(purchaseOrderItem.MaterialName).GetLocale(ctx.GetLanguage())
				disabledMaterials = append(disabledMaterials, materialName)
			}
		}
		// 如果有停用的物品，返回相应的错误消息
		if len(disabledMaterials) > 0 {
			return resp.PurchaseReceiptOrderCreateResp{}, errors.NewWithCode(
				constant.CodeMaterialDisabled,
				fmt.Sprintf(
					i18n.Translate(ctx.GetLanguage(), "有%d项物品已停用，您可启用物品后再进行收货"),
					len(disabledMaterials),
				),
			)
		}
	}

	var result resp.PurchaseReceiptOrderCreateResp

	err := db.Transaction(func(tx *gorm.DB) error {
		purchaseOrderRepo := repository.NewPurchaseOrderRepo(tx)
		purchaseOrderItemRepo := repository.NewPurchaseOrderItemRepo(tx)
		receiptOrderRepo := repository.NewPurchaseReceiptOrderRepo(tx)
		receiptOrderItemRepo := repository.NewPurchaseReceiptOrderItemRepo(tx)

		// 查询采购申请
		purchaseOrder, err := purchaseOrderRepo.GetByUuid(req.PurchaseOrderUuid)
		if err != nil {
			return errors.WithMessage(err, "采购申请不存在")
		}
		if !purchaseOrder.CanReceive() {
			return errors.New("采购单状态不允许收货")
		}

		// 创建收货单
		receiptOrder := &model.PurchaseReceiptOrder{
			OrderNo:           s.generateReceiptNo(ctx, tx),
			Status:            utils.IfInt(req.IsConfirm, constant.ReceiptOrderStatusReceived, constant.ReceiptOrderStatusPending), // 待收货状态
			PurchaseOrderUuid: req.PurchaseOrderUuid,
			PurchaseOrderNo:   purchaseOrder.OrderNo,
			PurchaseTime:      purchaseOrder.OrderTime,
			Num:               float64(len(req.Items)),
			ExpectArrivalTime: purchaseOrder.ExpectArrivalTime,
			ReceiveTime:       req.ReceiveTime,
			PurchaseOrder:     *purchaseOrder,
		}

		err = receiptOrderRepo.Create(receiptOrder)
		if err != nil {
			return errors.WithMessage(err, "创建收货单失败")
		}

		// 创建收货明细并更新采购申请明细的到货数量
		var receiptItems []model.PurchaseReceiptOrderItem
		for _, itemReq := range req.Items {
			// 查询采购申请明细
			orderItem, err := purchaseOrderItemRepo.GetByUuid(itemReq.PurchaseOrderItemUuid)
			if err != nil {
				return errors.WithMessage(err, "查询采购申请明细失败")
			}

			// 更新采购申请明细的到货数量
			newArrivalNum := orderItem.ArrivalNum + itemReq.Num
			if newArrivalNum > orderItem.Num {
				return errors.New(
					fmt.Sprintf(
						i18n.Translate(ctx.GetLanguage(), "物品 %s 的收货数量不能超过申请数量（申请数量：%.0f，已到货：%.0f，本次收货：%.0f）"),
						language.JsonToLocaleResponse(orderItem.MaterialName).GetLocale(ctx.GetLanguage()),
						orderItem.Num,
						orderItem.ArrivalNum,
						itemReq.Num,
					),
				)
			}

			// 创建收货明细
			receiptItems = append(receiptItems, model.PurchaseReceiptOrderItem{
				ReceiptOrderUuid:      receiptOrder.Uuid,
				PurchaseOrderItemUuid: orderItem.Uuid,
				MaterialCode:          orderItem.MaterialCode,
				MaterialName:          orderItem.MaterialName,
				MaterialUuid:          orderItem.MaterialUuid,
				Num:                   itemReq.Num,
				UnitUuid:              orderItem.UnitUuid,
				UnitName:              orderItem.UnitName,
				BaseUnitUuid:          orderItem.BaseUnitUuid,
				BaseUnitName:          orderItem.BaseUnitName,
				UnitConversionRate:    orderItem.UnitConversionRate,
			})

			// 确认收货时，更新采购申请明细的到货数量
			if req.IsConfirm {
				orderItem.ArrivalNum = newArrivalNum
				err = purchaseOrderItemRepo.Update(orderItem)
				if err != nil {
					return errors.WithMessage(err, "更新采购申请明细失败")
				}
			}
		}

		// 批量创建收货明细
		err = receiptOrderItemRepo.CreateBatch(receiptItems)
		if err != nil {
			return errors.WithMessage(err, "创建收货明细失败")
		}

		// 更新收货单明细
		receiptOrder.Items = receiptItems

		// 检查收货单是否完成
		if receiptOrder.Status == constant.ReceiptOrderStatusReceived {
			err = s.checkAndUpdatePurchaseOrderStatus(ctx, tx, req.PurchaseOrderUuid)
			if err != nil {
				return err
			}
			// 添加物料库存
			err = s.addMaterialStock(ctx, tx, receiptOrder)
			if err != nil {
				return err
			}
		}

		result.Uuid = receiptOrder.Uuid
		result.OrderNo = receiptOrder.OrderNo

		return nil
	})

	if err != nil {
		return resp.PurchaseReceiptOrderCreateResp{}, err
	}

	return result, nil
}

// UpdatePurchaseReceiptOrder 更新收货单
func (s *purchaseOrderSrv) UpdatePurchaseReceiptOrder(ctx context.Context, req req.PurchaseReceiptOrderUpdateReq) error {
	db := s.dbm.GetDB(ctx.GetDbId())

	// 判断物品明细是否已经停用
	if req.IsConfirm {
		var disabledMaterials []string
		for _, item := range req.Items {
			purchaseOrderItem, err := repository.NewPurchaseOrderItemRepo(db).GetByUuid(item.PurchaseOrderItemUuid)
			if err != nil {
				return errors.WithMessage(err, "查询采购申请明细失败")
			}
			material, err := repository.NewMaterialRepo(db).GetMaterialByUuid(purchaseOrderItem.MaterialUuid)
			if err != nil {
				return errors.WithMessage(err, "查询物品明细失败")
			}
			// 判断物品是否停用
			if material.Status == false {
				materialName := language.JsonToLocaleResponse(purchaseOrderItem.MaterialName).GetLocale(ctx.GetLanguage())
				disabledMaterials = append(disabledMaterials, materialName)
			}
		}
		// 如果有停用的物品，返回相应的错误消息
		if len(disabledMaterials) > 0 {
			return errors.NewWithCode(
				constant.CodeMaterialDisabled,
				fmt.Sprintf(
					i18n.Translate(ctx.GetLanguage(), "有%d项物品已停用，您可启用物品后再进行收货"),
					len(disabledMaterials),
				),
			)
		}
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		purchaseOrderItemRepo := repository.NewPurchaseOrderItemRepo(tx)
		receiptOrderRepo := repository.NewPurchaseReceiptOrderRepo(tx)
		receiptOrderItemRepo := repository.NewPurchaseReceiptOrderItemRepo(tx)

		// 查询收货单
		receiptOrder, err := receiptOrderRepo.GetByUuid(req.Uuid)
		if err != nil {
			return errors.WithMessage(err, "收货单不存在")
		}

		// 查询采购申请
		var purchaseOrder *model.PurchaseOrder
		if req.IsConfirm {
			purchaseOrder, err = repository.NewPurchaseOrderRepo(tx).GetByUuid(receiptOrder.PurchaseOrderUuid)
			if err != nil {
				return errors.WithMessage(err, "采购申请不存在")
			}
			if !purchaseOrder.CanReceive() {
				return errors.New("采购单状态不允许收货")
			}
			receiptOrder.PurchaseOrder = *purchaseOrder
		}

		// 重新创建收货明细并更新采购申请明细的到货数量
		var receiptItems []model.PurchaseReceiptOrderItem
		for _, itemReq := range req.Items {
			// 查询收货单明细
			receiptOrderItem, err := receiptOrderItemRepo.GetByUuid(itemReq.Uuid)
			if err != nil {
				return errors.WithMessage(err, "查询收货单明细失败")
			}

			// 查询采购申请明细
			purchaseOrderItem, err := purchaseOrderItemRepo.GetByUuid(receiptOrderItem.PurchaseOrderItemUuid)
			if err != nil {
				return errors.WithMessage(err, "查询采购申请明细失败")
			}

			// 计算新的到货数量
			newArrivalNum := purchaseOrderItem.ArrivalNum + itemReq.Num
			if newArrivalNum > purchaseOrderItem.Num {
				return errors.New(
					fmt.Sprintf(
						i18n.Translate(ctx.GetLanguage(), "物品 %s 的收货数量不能超过申请数量（申请数量：%.0f，已到货：%.0f，本次收货：%.0f）"),
						language.JsonToLocaleResponse(purchaseOrderItem.MaterialName).GetLocale(ctx.GetLanguage()),
						purchaseOrderItem.Num,
						purchaseOrderItem.ArrivalNum,
						itemReq.Num,
					),
				)
			}

			// 创建收货明细
			receiptItems = append(receiptItems, model.PurchaseReceiptOrderItem{
				ReceiptOrderUuid:      receiptOrder.Uuid,
				PurchaseOrderItemUuid: purchaseOrderItem.Uuid,
				MaterialCode:          purchaseOrderItem.MaterialCode,
				MaterialName:          purchaseOrderItem.MaterialName,
				MaterialUuid:          purchaseOrderItem.MaterialUuid,
				Num:                   itemReq.Num,
				UnitUuid:              purchaseOrderItem.UnitUuid,
				UnitName:              purchaseOrderItem.UnitName,
				BaseUnitUuid:          purchaseOrderItem.BaseUnitUuid,
				BaseUnitName:          purchaseOrderItem.BaseUnitName,
				UnitConversionRate:    purchaseOrderItem.UnitConversionRate,
			})

			// 更新采购申请明细的到货数量
			if req.IsConfirm {
				purchaseOrderItem.ArrivalNum = newArrivalNum
				err = purchaseOrderItemRepo.Update(purchaseOrderItem)
				if err != nil {
					return errors.WithMessage(err, "更新采购申请明细失败")
				}
			}
		}

		// 删除所有现有收货明细
		err = receiptOrderItemRepo.DeleteByReceiptOrderUuid(receiptOrder.Uuid)
		if err != nil {
			return errors.WithMessage(err, "删除收货明细失败")
		}

		// 批量创建收货明细
		err = receiptOrderItemRepo.CreateBatch(receiptItems)
		if err != nil {
			return errors.WithMessage(err, "创建收货明细失败")
		}

		// 更新收货单状态
		receiptOrder.Status = utils.IfInt(req.IsConfirm, constant.ReceiptOrderStatusReceived, constant.ReceiptOrderStatusPending)
		receiptOrder.ReceiveTime = req.ReceiveTime
		err = receiptOrderRepo.Update(receiptOrder)
		if err != nil {
			return errors.WithMessage(err, "更新收货单状态失败")
		}
		receiptOrder.Items = receiptItems

		// 检查收货单是否完成
		if receiptOrder.Status == constant.ReceiptOrderStatusReceived {
			err = s.checkAndUpdatePurchaseOrderStatus(ctx, tx, receiptOrder.PurchaseOrderUuid)
			if err != nil {
				return err
			}
			// 添加物料库存
			err = s.addMaterialStock(ctx, tx, receiptOrder)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

// GetReceiptOrderList 获取收货单列表
func (s *purchaseOrderSrv) GetPurchaseReceiptOrderList(ctx context.Context, req req.PurchaseReceiptOrderListReq) (resp.PurchaseReceiptOrderListResp, error) {
	receiptOrderRepo := repository.NewPurchaseReceiptOrderRepo(s.dbm.GetDB(ctx.GetDbId()))

	// 构建查询选项
	var opts []repository.DBOption
	if req.OrderNo != "" {
		opts = append(opts, receiptOrderRepo.WhereReceiptNo(req.OrderNo))
	}
	if len(req.StatusIn) > 0 {
		opts = append(opts, receiptOrderRepo.WhereStatusIn(req.StatusIn))
	}

	// 排序
	opts = append(opts, receiptOrderRepo.OrderByCreateTime(true))

	// 查询数据
	receipts, total, err := receiptOrderRepo.GetListWithPagination(req.PageNo, req.PageSize, opts...)
	if err != nil {
		return resp.PurchaseReceiptOrderListResp{}, errors.WithMessage(err, "查询收货单列表失败")
	}

	// 转换响应数据
	var listResp []resp.PurchaseReceiptOrderInfo
	for _, receipt := range receipts {
		receiptInfo := resp.PurchaseReceiptOrderInfo{}
		err := copier.Copy(&receiptInfo, &receipt)
		if err != nil {
			continue
		}
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

	// 转换收货明细数据
	detailResp.Items = make([]resp.PurchaseReceiptItemInfo, 0)
	for _, item := range receipt.Items {
		itemInfo := resp.PurchaseReceiptItemInfo{}
		err = copier.Copy(&itemInfo, &item)
		if err != nil {
			return resp.PurchaseReceiptOrderDetailResp{}, errors.WithMessage(err, "数据转换失败")
		}
		itemInfo.LocaleName = *language.JsonToLocaleResponse(item.MaterialName)
		itemInfo.PurchaseNum = item.PurchaseOrderItem.Num
		itemInfo.ArrivalNum = item.Num
		itemInfo.LocaleUnitName = *language.JsonToLocaleResponse(item.UnitName)
		itemInfo.LocaleBaseUnitName = *language.JsonToLocaleResponse(item.BaseUnitName)
		//
		detailResp.Items = append(detailResp.Items, itemInfo)
	}

	return detailResp, nil
}

// CancelPurchaseReceiptOrder 取消收货单
func (s *purchaseOrderSrv) CancelPurchaseReceiptOrder(ctx context.Context, req req.PurchaseReceiptOrderCancelReq) error {
	db := s.dbm.GetDB(ctx.GetDbId())
	receiptOrderRepo := repository.NewPurchaseReceiptOrderRepo(db)

	// 查询收货单
	receiptOrder, err := receiptOrderRepo.GetByUuid(req.Uuid)
	if err != nil {
		return errors.WithMessage(err, "收货单不存在")
	}

	// 检查收货单状态
	if receiptOrder.Status != constant.ReceiptOrderStatusPending {
		return errors.New("收货单状态不允许取消")
	}

	// 取消收货单
	receiptOrder.Status = constant.ReceiptOrderStatusRejected
	receiptOrder.CancelTime = time.Now().Unix()
	err = receiptOrderRepo.Update(receiptOrder)
	if err != nil {
		return errors.WithMessage(err, "取消收货单失败")
	}

	return nil
}

// generateOrderNo 生成采购申请订单编号
// 格式：CSSQ+年月日+0000自增序列号
func (s *purchaseOrderSrv) generateOrderNo(ctx context.Context, db *gorm.DB) string {
	// 固定前缀
	prefix := "CSSQ"
	// 年月日部分
	datePart := time.Now().Format("20060102")

	// 生成自增序列号
	serialNo, err := s.generatePurchaseOrderSerialNo(ctx, db)
	if err != nil {
		return ""
	}

	// 组装订单编号：CSSH+年月日+0000自增序列号
	orderNo := prefix + datePart + serialNo

	return orderNo
}

// generatePurchaseOrderSerialNo 生成采购申请自增序列号
func (s *purchaseOrderSrv) generatePurchaseOrderSerialNo(ctx context.Context, db *gorm.DB) (string, error) {
	var serialNo string
	purchaseOrderRepo := repository.NewPurchaseOrderRepo(db)

	// 获取今天最新的采购申请
	latestOrder, err := purchaseOrderRepo.GetLatestOrderToday()
	if err != nil && err != gorm.ErrRecordNotFound {
		return "", errors.WithMessage(err, "获取最新采购申请失败")
	}

	// 如果没有查询到今天的采购申请，则设置为0001
	if latestOrder == nil {
		serialNo = "0001"
		return serialNo, nil
	}

	// 从订单编号中提取序列号部分（最后4位）
	if len(latestOrder.OrderNo) >= 4 {
		lastSerialNo := latestOrder.OrderNo[len(latestOrder.OrderNo)-4:]
		serialNoNum, err := strconv.Atoi(lastSerialNo)
		if err != nil {
			// 如果解析失败，重新从0001开始
			serialNo = "0001"
		} else {
			// 序列号加1
			newSerialNoNum := serialNoNum + 1
			serialNo = fmt.Sprintf("%04d", newSerialNoNum)
		}
	} else {
		// 如果订单编号格式不正确，重新从0001开始
		serialNo = "0001"
	}

	return serialNo, nil
}

// generateReceiptNo 生成收货单号
// 格式：CSSH+年月日+0000自增序列号
func (s *purchaseOrderSrv) generateReceiptNo(ctx context.Context, db *gorm.DB) string {
	// 固定前缀
	prefix := "SHRK"
	// 年月日部分
	datePart := time.Now().Format("20060102")

	// 生成自增序列号
	serialNo, err := s.generateReceiptOrderSerialNo(ctx, db)
	if err != nil {
		return ""
	}

	// 组装收货单号：CSSH+年月日+0000自增序列号
	receiptNo := prefix + datePart + serialNo

	return receiptNo
}

// generateReceiptOrderSerialNo 生成收货单自增序列号
func (s *purchaseOrderSrv) generateReceiptOrderSerialNo(ctx context.Context, db *gorm.DB) (string, error) {
	var serialNo string
	receiptOrderRepo := repository.NewPurchaseReceiptOrderRepo(db)

	// 获取今天最新的收货单
	latestReceipt, err := receiptOrderRepo.GetLatestReceiptToday()
	if err != nil && err != gorm.ErrRecordNotFound {
		return "", errors.WithMessage(err, "获取最新收货单失败")
	}

	// 如果没有查询到今天的收货单，则设置为0001
	if latestReceipt == nil {
		serialNo = "0001"
		return serialNo, nil
	}

	// 从收货单号中提取序列号部分（最后4位）
	if len(latestReceipt.OrderNo) >= 4 {
		lastSerialNo := latestReceipt.OrderNo[len(latestReceipt.OrderNo)-4:]
		serialNoNum, err := strconv.Atoi(lastSerialNo)
		if err != nil {
			// 如果解析失败，重新从0001开始
			serialNo = "0001"
		} else {
			// 序列号加1
			newSerialNoNum := serialNoNum + 1
			serialNo = fmt.Sprintf("%04d", newSerialNoNum)
		}
	} else {
		// 如果收货单号格式不正确，重新从0001开始
		serialNo = "0001"
	}

	return serialNo, nil
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

// checkAndUpdatePurchaseOrderStatus 检查并更新采购单状态
func (s *purchaseOrderSrv) checkAndUpdatePurchaseOrderStatus(ctx context.Context, db *gorm.DB, purchaseOrderUuid uint64) error {
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
	} else if partialReceived && purchaseOrder.Status != constant.PurchaseOrderStatusCompleted {
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

	// 创建采购单操作日志
	err = s.createPurchaseOrderLog(db, purchaseOrderUuid, ctx, "update_status", "更新采购申请状态", oldStatus, purchaseOrder.Status, "")
	if err != nil {
		return err
	}

	return nil
}

// addMaterialStock 添加物料库存
func (s *purchaseOrderSrv) addMaterialStock(ctx context.Context, db *gorm.DB, receiptOrder *model.PurchaseReceiptOrder) error {
	if receiptOrder.Status != constant.ReceiptOrderStatusReceived {
		return nil
	}

	// 添加物料库存
	materialRepo := repository.NewMaterialRepo(db)
	for _, item := range receiptOrder.Items {
		// 获取物料信息
		material, err := materialRepo.GetMaterialByUuid(item.MaterialUuid, materialRepo.WithRelatedMaterialList())
		if err != nil {
			return errors.WithMessage(err, "获取物料信息失败")
		}
		// 更新库存数量
		material.StockNum = decimal.NewFromFloat(material.StockNum).Add(
			decimal.NewFromFloat(item.Num).Mul(decimal.NewFromFloat(item.UnitConversionRate).Round(4)),
		).InexactFloat64()
		err = materialRepo.UpdateMaterial(material)
		if err != nil {
			return errors.WithMessage(err, "更新物料库存失败")
		}
		//
		relatedMaterialUuids := make([]uint64, 0)
		for _, relatedMaterial := range material.RelatedMaterialList {
			if relatedMaterial.IsDelete() {
				continue
			}
			if relatedMaterial.IsUsed == 0 {
				continue
			}
			relatedMaterialUuids = append(relatedMaterialUuids, relatedMaterial.Uuid)
		}
		err = s.updateRelatedMaterialStock(db, relatedMaterialUuids)
		if err != nil {
			return errors.WithMessage(err, "更新规格/加料关联材料库存失败")
		}
	}

	// 调用erp接口
	if ctx.GetCompany().IsOpenErp() {
		erpReq := buying.SavePurchaseReceiptReq{
			PurchaseOrderName: receiptOrder.PurchaseOrder.ErpOrderNo,
			Items:             make([]*buying.PurchaseOrderItem, 0),
		}
		for _, item := range receiptOrder.Items {
			if item.Num > 0 {
				erpReq.Items = append(erpReq.Items, &buying.PurchaseOrderItem{
					ItemCode: item.MaterialCode,
					ItemName: language.JsonToLocaleResponse(item.MaterialName).EN,
					StockUom: language.JsonToLocaleResponse(item.UnitName).EN,
					Qty:      item.Num,
				})
			}
		}
		resp, err := erp.NewIErpSrv(s.dbm).SavePurchaseReceipt(ctx, &erpReq)
		if err != nil {
			return err
		}
		receiptOrder.ErpOrderNo = resp.PurchaseReceipt.PurchaseReceiptName
		err = repository.NewPurchaseReceiptOrderRepo(db).Update(receiptOrder)
		if err != nil {
			return errors.WithMessage(err, "更新收货单号失败")
		}
	}

	return nil
}

// updateRelatedMaterialStock 更新规格/加料关联材料库存
func (s *purchaseOrderSrv) updateRelatedMaterialStock(db *gorm.DB, relatedMaterialUuids []uint64) error {
	// 如果材料UUID列表为空，直接返回
	if len(relatedMaterialUuids) == 0 {
		return nil
	}
	// 使用事务确保数据一致性
	return db.Transaction(func(tx *gorm.DB) error {
		// 构建复杂SQL查询来更新产品BOM的库存数量
		// 1. 从related_material表获取关联信息
		// 2. 连接material表获取材料库存信息
		// 3. 计算每个关联的最小库存数量（材料库存/用量）
		// 4. 更新product_bom表的stock_num字段
		sql := `
			UPDATE ttpos_product_bom AS pb 
			JOIN (
				SELECT 
					rm.related_uuid, 
					LEAST(IFNULL(
						FLOOR(
							MIN(
								m.stock_num / (rm.num * rm.base_unit_conversion_rate)
							)
						)
					, 0), 99999999) AS min_stock_num
				FROM ttpos_related_material AS rm
				JOIN ttpos_material AS m ON rm.material_uuid = m.uuid
				WHERE rm.uuid IN (?) 
				  AND rm.delete_time = 0 
				  AND rm.unit_uuid > 0
				GROUP BY rm.related_uuid
			) AS sub ON pb.product_bom_card_uuid = sub.related_uuid
			SET pb.stock_num = sub.min_stock_num
			WHERE pb.product_bom_card_uuid IN (
				SELECT 
					DISTINCT related_uuid 
				FROM ttpos_related_material 
				WHERE uuid IN (?) 
				AND delete_time = 0 
				AND unit_uuid > 0
			)
		`

		// 执行SQL更新
		err := tx.Exec(sql, relatedMaterialUuids, relatedMaterialUuids).Error
		if err != nil {
			return errors.WithMessage(err, "更新规格/加料关联材料库存失败")
		}

		return nil
	})
}
