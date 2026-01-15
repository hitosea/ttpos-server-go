package purchase_order

import (
	"fmt"
	"strconv"
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
	"ttpos-server-go/app/service/rpc/erp"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/i18n"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/language"
	"ttpos-server-go/pkg/lock"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"github.com/jinzhu/copier"
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
	lock       lock.Lock
	settingSrv setting.ISrv
}

// NewPurchaseOrderSrv 创建采购申请服务
func NewPurchaseOrderSrv(dbm *database.DBManager, settingSrv setting.ISrv) IPurchaseOrderSrv {
	return NewPurchaseOrderSrvImpl(dbm, settingSrv)
}

// NewPurchaseOrderSrvImpl 创建采购申请服务实现
func NewPurchaseOrderSrvImpl(dbm *database.DBManager, settingSrv setting.ISrv) IPurchaseOrderSrv {
	return &purchaseOrderSrv{
		dbm:        dbm,
		validator:  &purchaseOrderValidator{},
		helper:     &purchaseOrderHelper{},
		receiptSrv: newPurchaseReceiptOrderSrv(dbm),
		lock:       lock.NewSystemLock(),
		settingSrv: settingSrv,
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
	if len(req.UuidIn) > 0 {
		opts = append(opts, purchaseOrderRepo.WhereUuidIn(req.UuidIn))
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
		purchaseOrderRepo.WithSupplier(),
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
	detailResp.WarehouseName = *language.JsonToLocaleResponse(purchaseOrder.WarehouseName)

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
		itemInfo.InternalCode = func(item model.PurchaseOrderItem) string {
			if item.Material == nil {
				return ""
			}
			return item.Material.InternalCode
		}(item)
		itemInfo.BarcodeValue = func(item model.PurchaseOrderItem) string {
			if item.Material == nil {
				return ""
			}
			return item.Material.BarcodeValue
		}(item)
		// 采购单位列表
		itemInfo.UnitList = func(item model.PurchaseOrderItem) []resp.PurchaseOrderItemMaterialUnit {
			unitList := []resp.PurchaseOrderItemMaterialUnit{}
			if item.Material == nil {
				return unitList
			}
			for _, unit := range item.Material.NotBaseUnitList {
				unitList = append(unitList, resp.PurchaseOrderItemMaterialUnit{
					Uuid: unit.Uuid,
					LocaleName: func() dto.LocaleResponse {
						if unit.Unit == nil {
							return dto.LocaleResponse{}
						}
						if unit.Unit.MultiLanguageName == (model.MultiLanguageName{}) {
							return dto.LocaleResponse{}
						}
						return unit.Unit.MultiLanguageName.GetNames()
					}(),
				})
			}
			return unitList
		}(item)
		// 单位列表
		itemInfo.Units = func(item model.PurchaseOrderItem) []resp.PurchaseOrderItemUnit {
			unitList := []resp.PurchaseOrderItemUnit{}
			if len(item.Units) == 0 && item.BaseUnitUuid != 0 {
				unitList = append(unitList, resp.PurchaseOrderItemUnit{
					Num:         item.Num,
					PurchaseNum: item.Num,
					ArrivalNum:  item.ArrivalNum,
					UnitUuid:    item.UnitUuid,
					LocaleName:  *language.JsonToLocaleResponse(item.UnitName),
				})
			} else {
				for _, unit := range item.Units {
					unitList = append(unitList, resp.PurchaseOrderItemUnit{
						Num:         unit.Num,
						PurchaseNum: unit.Num,
						ArrivalNum:  unit.ArrivalNum,
						UnitUuid:    unit.UnitUuid,
						LocaleName: func() dto.LocaleResponse {
							return *language.JsonToLocaleResponse(unit.UnitName)
						}(),
					})
				}
			}

			return unitList
		}(item)
		detailResp.Items = append(detailResp.Items, itemInfo)
	}

	return detailResp, nil
}

// CreatePurchaseOrder 创建采购申请
func (s *purchaseOrderSrv) CreatePurchaseOrder(
	ctx context.Context,
	req req.PurchaseOrderCreateReq,
) (resp.PurchaseOrderCreateResp, error) {
	// // 加锁
	s.lock.LockUuidString(req.SupplierErpCode + strconv.FormatInt(req.OrderTime, 10))
	defer s.lock.UnlockUuidString(req.SupplierErpCode + strconv.FormatInt(req.OrderTime, 10))

	// 验证请求
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

	// 获取店铺编码
	var companyStoreCode string
	storeSetting, err := s.settingSrv.GetStoreSetting(ctx)
	if err == nil {
		companyStoreCode = storeSetting.StoreCode
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		purchaseOrderRepo := repository.NewPurchaseOrderRepo(tx)
		purchaseOrderItemRepo := repository.NewPurchaseOrderItemRepo(tx)

		// 获取默认仓库
		defaultWarehouse, err := repository.NewWarehouseRepo(tx).GetDefaultWarehouse()
		if err != nil {
			return errors.WithMessage(errors.New("获取默认仓库失败"), err.Error())
		}

		// 获取 saas 数据库连接
		saasDB := s.dbm.GetDB(constant.DefaultDB)
		if saasDB == nil {
			return errors.New("saas 数据库连接失败")
		}

		// 获取公司 UUID（使用总部 UUID 或当前公司 UUID）
		companyUuid := ctx.GetCompanySetting().HeadquarterUuid
		if companyUuid == 0 {
			companyUuid = ctx.GetCompanyUuid()
		}

		// 确定前缀和编号类型
		var prefix, numberType string
		if req.PurchaseType == 2 {
			// 品牌采购（内部）
			prefix = "TPHY"
			numberType = constant.NumberTypeBrandPurchase
		} else {
			// 采购申请（外部）
			prefix = "PR"
			numberType = constant.NumberTypePurchaseReq
		}

		// 生成订单编号
		orderNo, err := s.helper.generateOrderNo(
			saasDB,
			companyUuid,
			prefix,
			numberType,
			ctx.GetCompanySetting().Timezone,
		)
		if err != nil {
			return errors.WithMessage(err, "生成订单编号失败")
		}

		// 获取仓库名称
		warehouseName := ""
		if req.WarehouseErpCode != "" {
			warehouse, err := repository.NewWarehouseRepo(tx).GetByErpCode(req.WarehouseErpCode)
			if err == nil {
				warehouseName = warehouse.Name
			}
		}

		// 设置期望到货时间，如果为空则默认为2035-12-31
		expectArrivalTime := req.ExpectedDeliveryTime
		if expectArrivalTime == 0 {
			expectArrivalTime = 2082672000 // 2035-12-31的时间戳
		}

		// 创建采购申请
		purchaseOrder := &model.PurchaseOrder{
			OrderNo:           orderNo,
			SupplierName:      req.SupplierName,
			SupplierErpCode:   utils.IfString(req.SupplierErpCode != "", req.SupplierErpCode, req.SupplierName),
			Status:            constant.PurchaseOrderStatusDraft,
			Num:               float64(len(req.Items)),
			OrderTime:         time.Now().Unix(), // 单据日期，采购单提交的时间（时间戳）
			ExpectArrivalTime: expectArrivalTime,
			ApplicantUuid:     ctx.GetStaffUuid(),
			ApplicantName:     ctx.GetStaff().RealName,
			PurchaseType:      utils.IfInt(req.PurchaseType == 2, 2, 1),
			WarehouseErpCode:  req.WarehouseErpCode,
			WarehouseName:     warehouseName,
			CompanyStoreCode:  companyStoreCode,
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
				UnitList:     item.UnitList,
			})
		}
		items, _, err := s.validator.buildPurchaseOrderItems(tx, purchaseOrder.Uuid, itemReqs)
		if err != nil {
			return err
		}

		// 创建采购申请明细
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
	// 加锁
	s.lock.LockUuid(req.Uuid)
	defer s.lock.UnlockUuid(req.Uuid)

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
		if !purchaseOrder.IsEditable() && purchaseOrder.Status != constant.PurchaseOrderStatusPending {
			return errors.New("当前状态不允许编辑")
		}

		// 检查是否需要更新（优化：避免不必要的数据库操作）
		needUpdate, err := s.helper.checkPurchaseOrderNeedUpdate(purchaseOrder, &req, purchaseOrderItemRepo)
		if err != nil {
			return err
		}
		// 如果没有任何变动，直接返回
		if !needUpdate {
			return nil
		}

		// 更新采购申请基本信息
		if purchaseOrder.Status != constant.PurchaseOrderStatusPending && purchaseOrder.Status != constant.PurchaseOrderStatusHeadquarterPending {
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
			// 设置期望到货时间，如果为空则默认为2035-12-31
			expectArrivalTime := req.ExpectedDeliveryTime
			if expectArrivalTime == 0 {
				expectArrivalTime = 2082672000 // 2035-12-31的时间戳
			}
			purchaseOrder.SupplierName = req.SupplierName
			purchaseOrder.SupplierErpCode = req.SupplierErpCode
			purchaseOrder.ExpectArrivalTime = expectArrivalTime
			purchaseOrder.WarehouseErpCode = req.WarehouseErpCode
			purchaseOrder.WarehouseName = warehouseName
		}
		purchaseOrder.Num = float64(len(req.Items))
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
				UnitList:     item.UnitList,
			})
		}
		items, _, err := s.validator.buildPurchaseOrderItems(tx, purchaseOrder.Uuid, itemReqs)
		if err != nil {
			return err
		}

		// 创建采购申请明细
		err = purchaseOrderItemRepo.CreateBatch(items)
		if err != nil {
			return errors.WithMessage(errors.New("创建采购申请明细失败"), err.Error())
		}

		// 如果当前操作店铺不是采购单归属店铺，需要同步更新归属店铺的数据
		if purchaseOrder.IsHeadquarterPurchase() && purchaseOrder.CompanyUuid != 0 && purchaseOrder.CompanyUuid != ctx.GetCompanyUuid() {
			ctxCopy := ctx.Copy()
			ctxCopy.SetDB(tx)
			err = s.syncItemsToCompanyShop(ctxCopy, purchaseOrder, req.Items)
			if err != nil {
				return errors.WithMessage(errors.New("同步归属店铺采购明细失败"), err.Error())
			}
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
	// 加锁
	s.lock.LockUuid(req.Uuid)
	defer s.lock.UnlockUuid(req.Uuid)

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
	// 加锁
	s.lock.LockUuid(req.Uuid)
	defer s.lock.UnlockUuid(req.Uuid)

	db := ctx.GetDB()
	companyUuid := ctx.GetCompanyUuid()

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

		// 🔥 新增：品牌采购三维度限额校验
		if purchaseOrder.IsHeadquarterPurchase() {
			// ① 检查申请次数限制
			if err := s.checkDailySubmitLimit(ctx, companyUuid); err != nil {
				return err
			}
			// ② 检查物品限购
			if err := s.checkPurchaseQuota(ctx, purchaseOrder); err != nil {
				return err
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
	// 加锁
	s.lock.LockUuid(req.Uuid)
	defer s.lock.UnlockUuid(req.Uuid)

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

		// 审核通过时检查
		if req.Action == "approve" {
			// 检查供应商状态
			if err := s.validator.validateSupplierStatus(tx, purchaseOrder.SupplierErpCode); err != nil {
				return err
			}
			// 检查物品状态
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
		hqMaterialRepo := repository.NewMaterialRepo(hqTx)
		var headquarterItems []model.PurchaseOrderItem
		for _, item := range purchaseOrder.Items {
			headquarterItem := model.PurchaseOrderItem{}
			err = copier.Copy(&headquarterItem, item)
			if err != nil {
				logger.Logger.Error("复制总部物品明细失败", zap.Error(err), zap.String("物料编码", item.MaterialCode))
				return errors.WithMessage(errors.New("复制总部物品明细失败"), err.Error())
			}

			// 查询物料
			material, err := hqMaterialRepo.GetMaterialByErpCode(item.MaterialCode)
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
			headquarterItem.Units = make([]model.PurchaseOrderItemUnit, 0)
			// 复制单位列表
			if len(item.Units) > 0 {
				for _, unit := range item.Units {
					itemUnit := model.PurchaseOrderItemUnit{}
					err := copier.Copy(&itemUnit, unit)
					if err != nil {
						logger.Logger.Error("复制总部物品单位失败", zap.Error(err))
						return errors.WithMessage(errors.New("复制总部物品单位失败"), err.Error())
					}
					itemUnit.BaseModel.ID = 0
					itemUnit.PurchaseOrderUuid = headquarterPurchaseOrder.Uuid
					itemUnit.BaseUnitUuid = material.UnitUuid
					headquarterItem.Units = append(headquarterItem.Units, itemUnit)
				}
			}
			// 添加到总部采购申请明细列表
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
		if len(item.Units) == 0 && item.BaseUnitUuid != 0 {
			if item.Num <= 0 {
				continue
			}
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
		} else {
			for _, unit := range item.Units {
				if unit.Num <= 0 {
					continue
				}
				stockItems = append(stockItems, &stock.MaterialRequestItem{
					ItemCode:     item.MaterialCode,
					Qty:          unit.Num,
					ScheduleDate: purchaseOrder.ExpectArrivalTime,
					Uom:          unit.ErpnextUom,
				})
			}
		}
	}

	// 获取子公司数据库
	subDb := s.dbm.GetDB(purchaseOrder.CompanyUuid)
	if subDb == nil {
		return "", errors.New("获取子店数据库失败")
	}

	// 减总部库存并记录出入库日志
	err := s.helper.reduceHeadquarterStockAndLog(ctx, subDb, tx, purchaseOrder)
	if err != nil {
		return "", err
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
		RefNo:           purchaseOrder.OrderNo, // 来源单据号，用于跟踪ttpos原始订单号
	})
	if err != nil {
		return "", s.helper.handleErpError(ctx, err, purchaseOrder)
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
	// 获取在途仓库
	transitWarehouse, _ := repository.NewWarehouseRepo(tx).GetTransitWarehouse()
	// 获取供应商ID
	supplierUuid := func() uint64 {
		supplier, err := repository.NewSupplierRepo(tx).GetByErpCode(purchaseOrder.SupplierErpCode)
		if err != nil {
			return 0
		}
		return supplier.Uuid
	}()
	// 构建采购订单项
	stockItems := make([]*buying.PurchaseOrderItemInput, 0, len(purchaseOrder.Items))
	for _, item := range purchaseOrder.Items {
		actualNum := 0.0
		if len(item.Units) == 0 && item.BaseUnitUuid != 0 {
			// 获取实际数量
			actualNum = item.GetConversionRateNum()
			if actualNum <= 0 {
				continue
			}
			//
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
		} else {
			for _, unit := range item.Units {
				unitActualNum := unit.GetConversionRateNum()
				if unitActualNum <= 0 {
					continue
				}
				actualNum += unitActualNum
				stockItems = append(stockItems, &buying.PurchaseOrderItemInput{
					ItemCode: item.MaterialCode,
					Qty:      unit.Num,
					Uom:      unit.ErpnextUom,
				})
			}
		}

		// 添加到本店的在途仓库
		if transitWarehouse != nil {
			err := s.helper.AddToTransitWarehouse(tx, transitWarehouse, purchaseOrder, supplierUuid, &item, actualNum)
			if err != nil {
				return "", err
			}
		}
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
		return "", s.helper.handleErpError(ctx, err, purchaseOrder)
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

// syncItemsToCompanyShop 同步采购明细到归属店铺
func (s *purchaseOrderSrv) syncItemsToCompanyShop(
	ctx context.Context,
	purchaseOrder *model.PurchaseOrder,
	reqItems []req.PurchaseOrderItemUpdateReq,
) error {
	// 获取归属店铺的数据库
	companyDb := s.dbm.GetDB(purchaseOrder.CompanyUuid)
	if companyDb == nil {
		return errors.New("获取归属店铺数据库失败")
	}

	// 获取当前操作数据库的订单最新明细（用于获取 material_code 和单位信息）
	currentItemRepo := repository.NewPurchaseOrderItemRepo(ctx.GetDB())
	currentItems, err := currentItemRepo.GetByPurchaseOrderUuid(
		purchaseOrder.Uuid,
		currentItemRepo.WithPreloadUnits(),
	)
	if err != nil {
		return errors.WithMessage(errors.New("查询当前采购明细失败"), err.Error())
	}

	return companyDb.Transaction(func(tx *gorm.DB) error {
		companyOrderRepo := repository.NewPurchaseOrderRepo(tx)
		companyItemRepo := repository.NewPurchaseOrderItemRepo(tx)
		companyItemUnitRepo := repository.NewPurchaseOrderItemUnitRepo(tx)

		// 查询归属店铺的采购单
		companyOrder, err := companyOrderRepo.GetByUuid(
			purchaseOrder.SubUuid,
			companyOrderRepo.WithItems(),
		)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				// 如果归属店铺没有这个采购单，不报错，直接返回
				return nil
			}
			return errors.WithMessage(errors.New("查询归属店铺采购申请失败"), err.Error())
		}

		// 检查是否需要同步（优化：避免不必要的数据库操作）
		needSync := s.helper.checkCompanyShopNeedSync(companyOrder, currentItems, reqItems)
		if !needSync {
			return nil
		}

		// 获取归属店铺现有的明细项
		// 构建现有明细的映射：MaterialCode -> Item
		existingItemMap := make(map[string]*model.PurchaseOrderItem)
		for i := range companyOrder.Items {
			existingItemMap[companyOrder.Items[i].MaterialCode] = &companyOrder.Items[i]
		}

		// 构建当前明细的映射：MaterialUuid -> MaterialCode
		materialCodeMap := make(map[uint64]string)
		for _, item := range currentItems {
			materialCodeMap[item.MaterialUuid] = item.MaterialCode
		}

		// 构建请求中的物品映射：MaterialCode -> PurchaseOrderItemReq
		reqItemMap := make(map[string]PurchaseOrderItemReq)
		for _, item := range reqItems {
			if materialCode, ok := materialCodeMap[item.MaterialUuid]; ok {
				reqItemMap[materialCode] = PurchaseOrderItemReq{
					MaterialUuid: item.MaterialUuid,
					UnitList:     item.UnitList,
				}
			}
		}

		// 1. 找出需要删除的明细（归属店铺存在，但请求中不存在）
		itemUuidsToDelete := make([]uint64, 0)
		for materialCode, existingItem := range existingItemMap {
			if _, exists := reqItemMap[materialCode]; !exists {
				itemUuidsToDelete = append(itemUuidsToDelete, existingItem.Uuid)
			}
		}

		// 2. 找出需要新增和更新的明细
		itemReqsToCreate := make([]PurchaseOrderItemReq, 0)
		for materialCode, reqItem := range reqItemMap {
			if existingItem, exists := existingItemMap[materialCode]; exists {
				// 已存在，需要更新（删除旧的，创建新的）
				itemUuidsToDelete = append(itemUuidsToDelete, existingItem.Uuid)
			}
			// 都加入创建列表
			itemReqsToCreate = append(itemReqsToCreate, reqItem)
		}

		// 执行删除操作
		if len(itemUuidsToDelete) > 0 {
			// 先删除关联的单位
			err = companyItemUnitRepo.DeleteByItemUuids(itemUuidsToDelete)
			if err != nil {
				return errors.WithMessage(errors.New("删除归属店铺明细单位失败"), err.Error())
			}
			// 再删除明细项
			err = companyItemRepo.DeleteByUuids(itemUuidsToDelete)
			if err != nil {
				return errors.WithMessage(errors.New("删除归属店铺明细失败"), err.Error())
			}
		}

		// 执行新增操作
		if len(itemReqsToCreate) > 0 {
			// 1. 尝试使用子店铺的物料配置构建明细（优先使用子店铺配置）
			itemsFromCompany, _, err := s.validator.buildPurchaseOrderItems(tx, companyOrder.Uuid, itemReqsToCreate, true)
			if err != nil {
				return err
			}

			// 2. 找出在子店铺中不存在的物料（需要从总部复制）
			companyMaterialUuids := make(map[uint64]bool)
			for _, item := range itemsFromCompany {
				companyMaterialUuids[item.MaterialUuid] = true
			}

			// 3. 构建缺失物料的请求列表
			missingItemReqs := make([]PurchaseOrderItemReq, 0)
			for _, itemReq := range itemReqsToCreate {
				if !companyMaterialUuids[itemReq.MaterialUuid] {
					missingItemReqs = append(missingItemReqs, itemReq)
				}
			}

			// 4. 使用总部数据库构建缺失的物料明细
			var itemsFromHeadquarter []model.PurchaseOrderItem
			if len(missingItemReqs) > 0 {
				itemsFromHeadquarter, _, err = s.validator.buildPurchaseOrderItems(ctx.GetDB(), companyOrder.Uuid, missingItemReqs)
				if err != nil {
					return errors.WithMessage(errors.New("从总部构建明细失败"), err.Error())
				}
			}

			// 5. 合并子店铺和总部的明细
			allItems := append(itemsFromCompany, itemsFromHeadquarter...)

			// 批量创建明细
			if len(allItems) > 0 {
				err = companyItemRepo.CreateBatch(allItems)
				if err != nil {
					return errors.WithMessage(errors.New("创建归属店铺明细失败"), err.Error())
				}
			}
		}

		// 更新归属店铺采购单的物品数量
		companyOrder.Num = float64(len(reqItems))
		err = companyOrderRepo.Update(companyOrder)
		if err != nil {
			return errors.WithMessage(errors.New("更新归属店铺采购申请失败"), err.Error())
		}

		return nil
	})
}

// 收货单相关方法委托给receiptSrv
func (s *purchaseOrderSrv) CreatePurchaseReceiptOrder(
	ctx context.Context,
	req req.PurchaseReceiptCreateReq,
) (resp.PurchaseReceiptOrderCreateResp, error) {
	// 加锁
	s.lock.LockUuid(req.PurchaseOrderUuid)
	defer s.lock.UnlockUuid(req.PurchaseOrderUuid)
	// 查询采购申请
	purchaseOrder, err := repository.NewPurchaseOrderRepo(ctx.GetDB()).GetByUuid(req.PurchaseOrderUuid)
	if err != nil {
		return resp.PurchaseReceiptOrderCreateResp{}, errors.WithMessage(errors.New("采购申请不存在"), err.Error())
	}
	// 创建收货单
	result, err := s.receiptSrv.CreatePurchaseReceiptOrder(ctx, req)
	if err != nil {
		// 调用erp接口失败
		if purchaseOrder != nil && strings.Contains(err.Error(), "调用erp接口失败") {
			return resp.PurchaseReceiptOrderCreateResp{}, s.helper.handleErpError(ctx, err, purchaseOrder)
		}
		return resp.PurchaseReceiptOrderCreateResp{}, err
	}
	return result, nil
}

// 更新收货单
func (s *purchaseOrderSrv) UpdatePurchaseReceiptOrder(
	ctx context.Context,
	req req.PurchaseReceiptOrderUpdateReq,
) error {
	// 加锁
	s.lock.LockUuid(req.Uuid)
	defer s.lock.UnlockUuid(req.Uuid)
	// 查询收货单
	receiptOrder, err := repository.NewPurchaseReceiptOrderRepo(ctx.GetDB()).GetByUuid(req.Uuid)
	if err != nil {
		return errors.WithMessage(errors.New("收货单不存在"), err.Error())
	}
	// 查询采购申请
	purchaseOrder, err := repository.NewPurchaseOrderRepo(ctx.GetDB()).GetByUuid(receiptOrder.PurchaseOrderUuid)
	if err != nil {
		return errors.WithMessage(errors.New("采购申请不存在"), err.Error())
	}
	//
	err = s.receiptSrv.UpdatePurchaseReceiptOrder(ctx, req)
	if err != nil {
		// 调用erp接口失败
		if purchaseOrder != nil && strings.Contains(err.Error(), "调用erp接口失败") {
			return s.helper.handleErpError(ctx, err, purchaseOrder)
		}
		return err
	}
	return nil
}

// 获取收货单列表
func (s *purchaseOrderSrv) GetPurchaseReceiptOrderList(
	ctx context.Context,
	reqs req.PurchaseReceiptOrderListReq,
) (resp.PurchaseReceiptOrderListResp, error) {
	if ctx.Version(context.LT, "2.7.0") {
		return s.receiptSrv.GetPurchaseReceiptOrderList(ctx, reqs)
	}

	// 定义查询结果结构
	type QueryResult struct {
		Uuid         uint64 `json:"uuid"`
		OrderNo      string `json:"order_no"`
		ErpOrderNo   string `json:"erp_order_no"`
		CreateTime   int64  `json:"create_time"`
		PurchaseType int    `json:"purchase_type"`
		Type         int    `json:"type"` // 1-收货单 2-采购单
		ReceiveTime  int64  `json:"receive_time"`
	}

	var results []QueryResult
	var total int64

	// 构建查询条件
	orderNoPattern := "%" + reqs.OrderNo + "%"

	// 时间格式兼容处理函数：前端传毫秒时间戳，数据库存秒时间戳
	convertTimestamp := func(timestamp int64) int64 {
		if timestamp <= 0 {
			return 0
		}
		// 如果是13位毫秒时间戳，转换为10位秒时间戳
		if timestamp > 9999999999 {
			return timestamp / 1000
		}
		return timestamp
	}

	// 处理时间字段
	receiveTimeStart := convertTimestamp(reqs.ReceiptTimeStart)
	receiveTimeEnd := convertTimestamp(reqs.ReceiptTimeEnd)

	// 联合查询
	sql := `
		SELECT * FROM (
			SELECT uuid, order_no, erp_order_no, create_time, purchase_type, 0 as receive_time, 2 as type FROM ttpos_purchase_order
			WHERE status = ?
			UNION ALL
			SELECT uuid, order_no, erp_order_no, create_time, receipt_type as purchase_type, receive_time, 1 as type FROM ttpos_purchase_receipt_order
			WHERE status in (?)
		) t 
		WHERE t.purchase_type = ?
		AND (t.order_no LIKE ? OR t.erp_order_no LIKE ?)
	`
	// 添加收货时间过滤条件
	if receiveTimeStart > 0 && receiveTimeEnd > 0 {
		sql += fmt.Sprintf(" AND (t.receive_time >= %d AND t.receive_time <= %d)", receiveTimeStart, receiveTimeEnd)
	}

	// 状态筛选
	status := constant.PurchaseOrderStatusApproved
	if len(reqs.StatusIn) > 0 && reqs.StatusIn[0] != 0 {
		status = -1
	}

	// 状态筛选
	if len(reqs.StatusIn) == 0 {
		reqs.StatusIn = []int{0, 1, 2}
	}

	// 先查询总数
	countSql := fmt.Sprintf(`SELECT COUNT(*) FROM (%s) t`, sql)
	err := ctx.GetDB().Raw(countSql, status, reqs.StatusIn, reqs.ReceiptType, orderNoPattern, orderNoPattern).Scan(&total).Error
	if err != nil {
		return resp.PurchaseReceiptOrderListResp{}, errors.WithMessage(errors.New("查询总数失败"), err.Error())
	}

	// 执行分页查询
	err = ctx.GetDB().Raw(sql+" LIMIT ? OFFSET ?", status, reqs.StatusIn, reqs.ReceiptType, orderNoPattern, orderNoPattern, reqs.PageReq.PageSize, reqs.PageSize*(reqs.PageReq.PageNo-1)).Scan(&results).Error
	if err != nil {
		return resp.PurchaseReceiptOrderListResp{}, errors.WithMessage(errors.New("查询列表失败"), err.Error())
	}

	// 分别处理采购单和收货单
	var purchaseOrderUuids []uint64
	var receiptOrderUuids []uint64
	for _, result := range results {
		if result.Type == 2 { // 采购单
			purchaseOrderUuids = append(purchaseOrderUuids, result.Uuid)
		} else if result.Type == 1 { // 收货单
			receiptOrderUuids = append(receiptOrderUuids, result.Uuid)
		}
	}

	// 构建响应数据
	response := resp.PurchaseReceiptOrderListResp{
		PurchaseOrderList: []*resp.PurchaseOrderInfo{},
		List:              []*resp.PurchaseReceiptOrderInfo{},
		Meta: dto.PageResponse{
			PageNo:   reqs.PageReq.PageNo,
			PageSize: reqs.PageReq.PageSize,
			Total:    total,
		},
	}

	// 如果有采购单，获取采购单详情
	if len(purchaseOrderUuids) > 0 {
		purchaseOrderList, err := s.GetPurchaseOrderList(ctx, req.PurchaseOrderListReq{
			UuidIn:       purchaseOrderUuids,
			PurchaseType: reqs.ReceiptType,
			PageReq: dto.PageReq{
				PageNo:   1,
				PageSize: len(purchaseOrderUuids),
			},
		})
		if err != nil {
			return resp.PurchaseReceiptOrderListResp{}, errors.WithMessage(errors.New("查询采购单列表失败"), err.Error())
		}
		response.PurchaseOrderList = purchaseOrderList.List
	}

	// 如果有收货单，获取收货单详情
	if len(receiptOrderUuids) > 0 {
		receiptOrderList, err := s.receiptSrv.GetPurchaseReceiptOrderList(ctx, req.PurchaseReceiptOrderListReq{
			UuidIn:      receiptOrderUuids,
			ReceiptType: reqs.ReceiptType,
			PageReq: dto.PageReq{
				PageNo:   1,
				PageSize: len(receiptOrderUuids),
			},
		})
		if err != nil {
			return resp.PurchaseReceiptOrderListResp{}, errors.WithMessage(errors.New("查询收货单列表失败"), err.Error())
		}
		response.List = receiptOrderList.List
	}

	return response, nil
}

// 获取收货单详情
func (s *purchaseOrderSrv) GetPurchaseReceiptOrderDetail(
	ctx context.Context,
	req req.PurchaseReceiptOrderDetailReq,
) (resp.PurchaseReceiptOrderDetailResp, error) {
	return s.receiptSrv.GetPurchaseReceiptOrderDetail(ctx, req)
}

// 取消收货单
func (s *purchaseOrderSrv) CancelPurchaseReceiptOrder(
	ctx context.Context,
	req req.PurchaseReceiptOrderCancelReq,
) error {
	// 加锁
	s.lock.LockUuid(req.Uuid)
	defer s.lock.UnlockUuid(req.Uuid)

	return s.receiptSrv.CancelPurchaseReceiptOrder(ctx, req)
}
