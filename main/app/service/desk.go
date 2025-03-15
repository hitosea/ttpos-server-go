package service

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
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/eventbus/event"
	"ttpos-server-go/pkg/lock"
	"ttpos-server-go/pkg/utils"

	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

// IDeskSrv 定义收银服务接口
type IDeskSrv interface {
	GetDeskList(ctx context.Context, dbId uint64, req req.DeskListReq) (resp.DeskListWithPaginationResp, error)         // 获取桌台列表
	GetDeskRegionAndTypeList(dbId uint64) (resp.DeskRegionAndTypeListWithPaginationResp, error)                         // 获取桌台区域和类型列表
	GetDeskInfo(dbId uint64, deskUuid uint64) (resp.Desk, error)                                                        // 获取桌台详情
	CreateDeskOrder(ctx context.Context, req req.DeskOrderCreateReq) (resp.CreateDeskOrderResp, error)                  // 创建桌台订单
	CloseDesk(ctx context.Context, req req.DeskCloseReq) error                                                          // 关闭桌台
	CompleteDesk(ctx context.Context, req req.DeskJsonUuidReq) error                                                    // 完成桌台
	ChangeDesk(ctx context.Context, req req.ChangeDeskReq) (*resp.ShopCart, error)                                      // 切换桌台
	MergeDesk(ctx context.Context, req req.MergeDeskReq) (*resp.DeskMergeShopCartResp, *resp.DeskMergeCheckResp, error) // 合并桌台
	IsCellCloseDesk(ctx context.Context, deskUuid uint64) (model.Desk, error)                                           // 判断桌台是否可以关闭
	GetTabletDeskList(ctx context.Context) (resp.TabletDeskList, error)                                                 // 平板获取桌台列表
	BindDesk(ctx context.Context, bindDeskReq req.BindDeskReq) error                                                    // 平板端绑定桌台
}

// deskSrv 收银服务结构体
type deskSrv struct {
	bus        *event.SystemEventBus
	dbm        *database.DBManager // 数据库管理器
	localeSrv  ILocaleSrv          // 多语言名称服务
	orderSrv   IOrderSrv           // 订单服务
	settingSrv setting.ISrv        // 设置服务
	deviceSrv  IDeviceSrv          // 设备服务
}

// NewDeskSrv 创建新的收银产品类别服务
func NewDeskSrv(dbm *database.DBManager, localeSrv ILocaleSrv, orderSrv IOrderSrv, settingSrv setting.ISrv, deviceSrv IDeviceSrv) IDeskSrv {
	return NewDeskSrvImpl(dbm, localeSrv, orderSrv, settingSrv, deviceSrv)
}

// NewDeskSrvImpl 创建新的收银服务实现
func NewDeskSrvImpl(dbm *database.DBManager, localeSrv ILocaleSrv, orderSrv IOrderSrv, settingSrv setting.ISrv, deviceSrv IDeviceSrv) IDeskSrv {
	return &deskSrv{
		bus:        event.NewSystemBus(),
		dbm:        dbm,
		localeSrv:  localeSrv,
		orderSrv:   orderSrv,
		settingSrv: settingSrv,
		deviceSrv:  deviceSrv,
	}
}

// GetDeskRegionAndTypeList 获取收银机点餐页面产品类别列表
func (s *deskSrv) GetDeskRegionAndTypeList(dbId uint64) (resp.DeskRegionAndTypeListWithPaginationResp, error) {
	// 获取列表
	regions, _ := repository.NewDeskRegionRepo(s.dbm.GetDB(dbId)).GetDeskRegionList()
	types, _ := repository.NewDeskTypeRepo(s.dbm.GetDB(dbId)).GetDeskTypeList()

	// 转换为响应对象
	deskRegionResp := make([]resp.DeskRegion, len(regions))
	for i, region := range regions {
		deskRegionResp[i] = resp.DeskRegion{
			Uuid: region.Uuid,
			Name: region.Name,
		}
	}

	deskTypeResp := make([]resp.DeskType, len(types))
	for i, type_ := range types {
		deskTypeResp[i] = resp.DeskType{
			Uuid: type_.Uuid,
			Name: type_.Name,
		}
	}

	// 返回响应对象
	return resp.DeskRegionAndTypeListWithPaginationResp{
		Region: struct {
			List []resp.DeskRegion `json:"list"`
		}{List: deskRegionResp},
		Type: struct {
			List []resp.DeskType `json:"list"`
		}{List: deskTypeResp},
	}, nil
}

// GetDeskList 获取收银机点餐页面产品类别列表
func (s *deskSrv) GetDeskList(ctx context.Context, dbId uint64, req req.DeskListReq) (resp.DeskListWithPaginationResp, error) {
	// 获取列表
	desks, total, err := repository.NewDeskRepo(s.dbm.GetDB(dbId)).GetClientDeskList(req.Status, req.IsBuffet, req.PageNo, req.PageSize)
	if err != nil {
		return resp.DeskListWithPaginationResp{}, errors.WithMessage(err)
	}

	// 初始化额外信息
	// 转换为响应对象
	deskResp := make([]resp.Desk, len(desks))
	for i, desk := range desks {
		deskResp[i] = desk.GetDeskResp()
	}

	// 返回响应对象
	return resp.DeskListWithPaginationResp{
		List: deskResp,
		Meta: dto.PageResponse{
			PageNo:   req.PageNo,
			PageSize: req.PageSize,
			Total:    total,
		},
	}, nil
}

// GetDeskInfo 获取桌台详情
func (s *deskSrv) GetDeskInfo(dbId uint64, deskUuid uint64) (resp.Desk, error) {
	// 获取桌台详情
	desk, err := repository.NewDeskRepo(s.dbm.GetDB(dbId)).GetDeskInfo(deskUuid)
	if err != nil {
		return resp.Desk{}, errors.WithMessage(err)
	}
	return desk.GetDeskResp(), nil
}

// CreateDeskOrder 创建桌台订单
func (s *deskSrv) CreateDeskOrder(ctx context.Context, req req.DeskOrderCreateReq) (resp.CreateDeskOrderResp, error) {
	dbId := ctx.GetDbId()
	// 验证请求参数
	err := req.ValidateCreateDeskOrderReq()
	if err != nil {
		return resp.CreateDeskOrderResp{}, errors.WithMessage(err)
	}

	// 判断桌台是否存在
	desk, _ := s.GetDeskInfo(dbId, req.DeskUuid)
	if desk.Uuid == 0 {
		return resp.CreateDeskOrderResp{}, errors.New("桌台不存在")
	}

	// 判断桌台是否空闲
	if desk.Status != constant.DeskStatusAvailable {
		return resp.CreateDeskOrderResp{}, errors.New("桌台不空闲")
	}

	// 判断收银机设置
	if ctx.GetSource() == constant.SourceCashier {
		cashierSetting, err := s.settingSrv.GetCashierSetting(ctx, nil)
		if err != nil {
			return resp.CreateDeskOrderResp{}, errors.WithMessage(err)
		}
		if cashierSetting.OrderMethod.IsTableOrder != "1" {
			return resp.CreateDeskOrderResp{}, errors.New("桌台用餐已关闭，请选择其他用餐方式")
		}
	}
	if ctx.GetSource() == constant.SourceTablet {
		tabletSetting, err := s.settingSrv.GetTabletSetting(ctx, nil)
		if err != nil {
			return resp.CreateDeskOrderResp{}, errors.WithMessage(err)
		}
		if tabletSetting.IsCustomerOrder != "1" {
			return resp.CreateDeskOrderResp{}, errors.New("未开启顾客开桌")
		}
	}
	if ctx.GetSource() == constant.SourceH5 {
		tabletSetting, err := s.settingSrv.GetH5Setting(ctx, nil)
		if err != nil {
			return resp.CreateDeskOrderResp{}, errors.WithMessage(err)
		}
		if tabletSetting.IsCustomerOrder != "1" {
			return resp.CreateDeskOrderResp{}, errors.New("未开启顾客开桌")
		}
	}

	// 判断是否自助餐订单
	var result resp.CreateDeskOrderResp
	if req.IsBuffet() {
		deskBuffetOrder, err := s.createDeskBuffetOrder(ctx, req)
		if err != nil {
			return resp.CreateDeskOrderResp{}, errors.WithMessage(err, "s.createDeskBuffetOrder failed, request:", utils.ToJson(req))
		}
		result = deskBuffetOrder
	} else {
		// 创建桌台-非自助餐订单
		deskOrder, err := s.orderSrv.CreateDeskOrder(ctx, req)
		if err != nil {
			return resp.CreateDeskOrderResp{}, errors.WithMessage(err, "s.orderSrv.CreateDeskOrder failed, request:", utils.ToJson(req))
		}
		result = deskOrder
	}

	// 发布“开台”操作事件
	go func() {
		s.bus.PublishOpenDeskEvent(event.OpenDeskPayload{
			BasePayload: event.BasePayload{
				CompanyUuid:   dbId,
				Source:        ctx.GetSource(),
				SaleBillUuid:  result.SaleBillUuid,
				SaleOrderUuid: result.SaleOrderUuid,
				OperatorUuid:  int64(ctx.GetStaffUuid()),
			},
			MealNum:  req.GetMealNum(),
			IsBuffet: req.IsBuffet(),
			TableId:  req.DeskUuid,
			TableNo:  desk.DeskNo,
		})
	}()
	return result, nil
}

// createDeskBuffetOrder 创建桌台自助餐订单
func (s *deskSrv) createDeskBuffetOrder(ctx context.Context, req req.DeskOrderCreateReq) (resp.CreateDeskOrderResp, error) {
	dbId := ctx.GetDbId()
	// 验证自助餐是否开启
	companySetting, err := s.settingSrv.GetCompanySetting(ctx)
	if err != nil {
		return resp.CreateDeskOrderResp{}, errors.WithMessage(err)
	}
	buffetSetting, err := s.settingSrv.GetBuffetSetting(ctx, companySetting)
	if err != nil {
		return resp.CreateDeskOrderResp{}, errors.WithMessage(err)
	}
	if buffetSetting.IsOpen == "0" {
		return resp.CreateDeskOrderResp{}, errors.New("自助餐未开启")
	}

	// 验证自助餐套餐是否存在
	for _, buffetUuid := range req.BuffetUuids {
		_, err := repository.NewBuffetRepo(s.dbm.GetDB(dbId)).GetBuffetInfo(repository.NewCommonRepo().WhereByUuid(buffetUuid))
		if err != nil {
			return resp.CreateDeskOrderResp{}, errors.WithMessage(err, "自助餐套餐不存在")
		}
	}

	// 验证自助餐顾客类型是否存在
	for _, buffetCustomerType := range req.BuffetCustomerTypes {
		_, err := repository.NewBuffetRepo(s.dbm.GetDB(dbId)).GetBuffetCustomerTypeInfo(repository.NewCommonRepo().WhereByUuid(buffetCustomerType.Uuid))
		if err != nil {
			return resp.CreateDeskOrderResp{}, errors.WithMessage(err, "自助餐顾客类型不存在")
		}
	}

	return s.orderSrv.CreateDeskOrder(ctx, req)
}

// IsCellCloseDesk 判断桌台是否可关闭
func (s *deskSrv) IsCellCloseDesk(ctx context.Context, deskUuid uint64) (model.Desk, error) {
	dbId := ctx.GetDbId()
	db := s.dbm.GetDB(dbId)
	desk, err := repository.NewDeskRepo(db).GetDeskInfo(deskUuid)
	if err != nil {
		return model.Desk{}, errors.WithMessage(err, "桌台不存在")
	}
	if desk.Status == 0 {
		return model.Desk{}, errors.New("桌台已关闭")
	}
	if desk.SaleBillUuid == 0 {
		return model.Desk{}, nil
	}
	if _, err := NewOrderSrv(s.dbm, nil, nil, nil).IsCellCancelOrder(ctx, desk.SaleBillUuid); err != nil {
		return model.Desk{}, errors.WithMessage(err)
	}
	return desk, nil
}

// CloseDesk 关闭桌台
func (s *deskSrv) CloseDesk(ctx context.Context, reqs req.DeskCloseReq) error {
	// 禁止并发操作
	if ctx.NoLock() {
		lock.NewSystemLock().LockUuid(reqs.Uuid)
		defer lock.NewSystemLock().UnlockUuid(reqs.Uuid)
		ctx.AddLock()
	}
	dbId := ctx.GetDbId()
	db := s.dbm.GetDB(dbId)
	desk, err := repository.NewDeskRepo(db).GetDeskInfo(reqs.Uuid)
	if err != nil {
		return errors.WithMessage(err)
	}
	// 如果桌台空闲，则直接返回关闭桌台成功
	if desk.IsAvailableDesk() {
		return nil
	}
	// 如果桌台非空闲，但没有关联账单，则更新桌台状态为关闭
	if !desk.IsAvailableDesk() && desk.SaleBillUuid == 0 {
		desk.SetCloseDesk()
		if err := repository.NewDeskRepo(db).UpdateDeskRecord(desk); err != nil {
			return errors.WithMessage(err, "NewDeskRepo(db).UpdateDeskRecord(desk) failed, params:", utils.ToJson(desk.GetRecord()))
		}
		return nil
	}
	// 如果桌台非空闲，且账单已完成或已取消时，直接关闭桌台
	if !desk.IsAvailableDesk() && desk.SaleBillUuid != 0 {
		ctx.Log().Info("请求关闭桌台。桌台非空闲，但销售账单已完成或已取消")
		// desk可能没有预加载SaleBill
		if desk.SaleBill == nil {
			saleBill, err := repository.NewSaleBillRepo(db).GetSaleBillRecord(desk.SaleBillUuid)
			if err != nil {
				return errors.WithMessage(err, "NewSaleBillRepo(db).GetSaleBillRecord failed, param:", fmt.Sprintf("%d", desk.SaleBillUuid))
			}
			desk.SaleBill = saleBill
		}
		if desk.SaleBill != nil && desk.SaleBill.IsEndStatus() {
			desk.SetCloseDesk()
			if err := repository.NewDeskRepo(db).UpdateDeskRecord(desk); err != nil {
				return errors.WithMessage(err, "NewDeskRepo(db).UpdateDeskRecord(desk) failed, params:", utils.ToJson(desk.GetRecord()))
			}
			return nil
		}
	}

	// 如果桌台非空闲，且关联有账单时
	// 取消订单
	if err := s.orderSrv.CancelOrder(ctx, req.OrderCancelReq{
		SaleBillUuid: desk.SaleBillUuid,
		CancelReason: reqs.Reason,
		Password:     reqs.Password,
	}); err != nil {
		return errors.WithMessage(err)
	}
	return nil
}

// CompleteDesk 完成桌台
func (s *deskSrv) CompleteDesk(ctx context.Context, reqs req.DeskJsonUuidReq) error {
	dbId := ctx.GetDbId()
	db := s.dbm.GetDB(dbId)
	desk, err := repository.NewDeskRepo(db).GetDeskInfo(reqs.Uuid)
	if err != nil {
		return errors.WithMessage(err)
	}
	if desk.SaleBill != nil && ctx.GetSource() == constant.SourceAssistant && desk.SaleBill.IsSplit() {
		return errors.WithMessage(errors.New("当前订单已拆单，请前去收银机操作"))
	}
	err = repository.NewDeskRepo(db).UpdateDeskByMap(reqs.Uuid, map[string]any{
		"sale_bill_uuid": 0,
		"status":         constant.DeskStatusClose,
	})
	if err != nil {
		return errors.WithMessage(err)
	}
	//
	return nil
}

// ChangeDesk 切换桌台
func (s *deskSrv) ChangeDesk(ctx context.Context, reqs req.ChangeDeskReq) (*resp.ShopCart, error) {
	// 禁止并发操作
	if ctx.NoLock() {
		lock.NewSystemLock().LockUuid(reqs.SaleBillUuid)
		defer lock.NewSystemLock().UnlockUuid(reqs.SaleBillUuid)
		ctx.AddLock()
	}

	db := s.dbm.GetDB(ctx.GetCompanyUuid())
	// 获取销售账单信息
	saleBill, errSaleBill := repository.NewOrderRepo(db).GetSaleBillAllInfo(reqs.SaleBillUuid)
	if errSaleBill != nil {
		return nil, errors.WithMessage(errSaleBill, "获取销售账单信息失败")
	}

	// 获取旧桌台信息
	oldDesk, errOldDesk := repository.NewDeskRepo(db).GetDeskRecord(saleBill.Desk.Uuid)
	if errOldDesk != nil {
		return nil, errors.WithMessage(errOldDesk, "获取旧桌台信息失败，Desk.Uuid:", fmt.Sprintf("%d", saleBill.Desk.Uuid))
	}
	// 检查旧桌台是否支持转台
	// 旧桌台已关闭,不允许转台
	// 旧桌台待清台,不允许转台
	if oldDesk.Status == constant.DeskStatusClose {
		return nil, errors.New("旧桌台已关闭，不允许转台")
	}
	if oldDesk.GetIsWaitClearStatus() {
		return nil, errors.New("旧桌台待清台，不允许转台")
	}

	// 获取桌台信息
	desk, errDesk := repository.NewDeskRepo(db).GetDeskRecord(reqs.DeskUuid)
	if errDesk != nil {
		return nil, errors.WithMessage(errDesk, "获取桌台信息失败")
	}
	// 检查新桌台是否支持转台
	// 新桌台已禁用,不允许转台
	// 新桌台非空闲,不允许转台
	if desk.IsDisableDesk() {
		return nil, errors.New("新桌台已禁用，不允许转台")
	}
	if !desk.IsAvailableDesk() {
		return nil, errors.New("新桌台非空闲，不允许转台")
	}

	// 设置新桌台的开台信息
	desk.SetOpenDesk(reqs.SaleBillUuid)
	// 更新销售账单的桌台uuid,修改旧桌台的状态
	saleBill.ChangeDesk(reqs.DeskUuid, desk.DeskNo)

	if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		// 更新新桌台的记录
		if err := repository.NewDeskRepo(tx).UpdateDeskRecord(*desk); err != nil {
			return errors.WithMessage(err)
		}
		// 更新旧桌台的记录
		if err := repository.NewDeskRepo(tx).UpdateDeskRecord(*saleBill.Desk); err != nil {
			return errors.WithMessage(err)
		}
		// 更新销售账单的记录
		if err := repository.NewSaleBillRepo(tx).UpdateSaleBillRecord(*saleBill); err != nil {
			return errors.WithMessage(err)
		}
		return nil
	}); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 发布转台事件
	go func() {
		ctx.Log().Info("发布转台事件")
		s.bus.PublishChangeDeskEvent(event.ChangeDeskPayload{
			BasePayload: event.BasePayload{
				CompanyUuid:   ctx.GetCompanyUuid(),
				Source:        ctx.GetSource(),
				SaleBillUuid:  reqs.SaleBillUuid,
				SaleOrderUuid: reqs.SaleOrderUuid,
				OperatorUuid:  int64(ctx.GetStaffUuid()),
			},
			Old: event.Old{TableId: oldDesk.Uuid, TableNo: oldDesk.DeskNo},
			New: event.New{TableId: reqs.DeskUuid, TableNo: desk.DeskNo},
		})
	}()

	// 返回购物车信息
	info, err := s.orderSrv.GetOrderCartInfo(ctx, reqs.SaleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err, "获取订单信息失败")
	}

	return info, nil
}

// MergeDesk 合并桌台
func (s *deskSrv) MergeDesk(ctx context.Context, req req.MergeDeskReq) (*resp.DeskMergeShopCartResp, *resp.DeskMergeCheckResp, error) {
	// 禁止并发操作
	if ctx.NoLock() {
		lock.NewSystemLock().LockUuid(req.SaleBillUuid)
		defer lock.NewSystemLock().UnlockUuid(req.SaleBillUuid)
		ctx.AddLock()
	}

	if err := req.Validate(); err != nil {
		return nil, nil, errors.WithMessage(err)
	}

	db := s.dbm.GetDB(ctx.GetCompanyUuid())

	// 获取销售账单信息
	saleBill, errSaleBill := repository.NewOrderRepo(db).GetSaleBillAllInfo(req.SaleBillUuid)
	if errSaleBill != nil {
		return nil, nil, errors.WithMessage(errSaleBill, "获取销售账单信息失败")
	}

	// 点餐助手，拆单后不可以修改人数
	if ctx.GetSource() == constant.SourceAssistant && saleBill.IsSplit() {
		return nil, nil, errors.New("当前订单已拆单，请前去收银机操作")
	}

	// 判断销售账单是否拆单
	if saleBill.IsSplit() {
		return nil, nil, errors.New("当前桌台已拆单，不支持合并桌台")
	}

	// 获取销售订单
	saleOrder := saleBill.GetFirstSaleOrder()

	// 判断桌台是否符合合并条件
	errDeskMsg := ""
	deskMergeCheckRes := resp.DeskMergeCheckResp{}
	saleBillList := []model.SaleBill{}
	deskNos := []string{}
	for _, deskUuid := range req.DeskUuids {
		desk, errDesk := repository.NewDeskRepo(db).GetDeskRecord(deskUuid)
		if errDesk != nil {
			return nil, nil, errors.WithMessage(errDesk, "获取桌台信息失败")
		}
		deskSaleBill, errSaleBill := repository.NewOrderRepo(db).GetSaleBillAllInfo(desk.SaleBillUuid)
		if errSaleBill != nil {
			deskMergeCheckRes.List = append(deskMergeCheckRes.List, resp.DeskNo{DeskNo: desk.DeskNo})
			return nil, nil, errors.New("桌台有变动，请重新选择")
		}
		// 以下桌台是自助餐，不支持合并桌台
		if deskSaleBill.IsBuffetSaleBill() {
			deskMergeCheckRes.List = append(deskMergeCheckRes.List, resp.DeskNo{DeskNo: desk.DeskNo})
			return nil, &deskMergeCheckRes, errors.New("自助餐桌台不符合并台")
		}
		// 以下桌台已被拆单，不支持合并桌台
		if deskSaleBill.IsSplit() {
			deskMergeCheckRes.List = append(deskMergeCheckRes.List, resp.DeskNo{DeskNo: desk.DeskNo})
			errDeskMsg = "以下桌台已被拆单，不支持合并桌台"
			continue
		}
		// 以下桌台已被部分支付，不支持合并桌台
		if deskSaleBill.IsPartialPay() {
			deskMergeCheckRes.List = append(deskMergeCheckRes.List, resp.DeskNo{DeskNo: desk.DeskNo})
			errDeskMsg = "以下桌台已被部分支付，不支持合并桌台"
			continue
		}
		// 将销售账单添加到销售账单列表
		saleBillList = append(saleBillList, *deskSaleBill)
		// 将桌台编号添加到桌台编号列表
		deskNos = append(deskNos, desk.DeskNo)
	}
	if len(deskMergeCheckRes.List) > 0 {
		return nil, &deskMergeCheckRes, errors.New(errDeskMsg)
	}

	resp := &resp.DeskMergeShopCartResp{}

	// 更新桌台信息
	if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		mealNum := uint(0)
		// 更新销售订单
		for _, deskSaleBill := range saleBillList {
			deskSaleOrder := deskSaleBill.GetFirstSaleOrder()

			// 更新销售订单商品
			if err := repository.NewOrderProductRepo(tx).Update(
				map[string]interface{}{
					"desk_uuid":       saleBill.DeskUuid,
					"sale_bill_uuid":  saleBill.Uuid,
					"sale_order_uuid": saleOrder.Uuid,
					"create_time":     time.Now().Unix(),
				},
				repository.NewCommonRepo().WhereBySaleBillUuid(deskSaleBill.Uuid),
				repository.NewCommonRepo().WhereBySaleOrderUuid(deskSaleOrder.Uuid),
			); err != nil {
				return errors.WithMessage(err)
			}

			// 静态追加到原订单
			for _, deskSaleOrderProduct := range deskSaleOrder.SaleOrderProducts {
				deskSaleOrderProduct.DeskUuid = saleBill.DeskUuid
				deskSaleOrderProduct.SaleBillUuid = saleBill.Uuid
				deskSaleOrderProduct.SaleOrderUuid = saleOrder.Uuid
				deskSaleOrderProduct.CreateTime = time.Now().Unix()
				saleOrder.SaleOrderProducts = append(saleOrder.SaleOrderProducts, deskSaleOrderProduct)
			}

			// 被并桌待接单记录变更为本桌
			if err := repository.NewH5OrderRepo(tx).Update(
				map[string]interface{}{
					"desk_uuid":   saleBill.DeskUuid,
					"desk_no":     saleBill.Desk.DeskNo,
					"create_time": time.Now().Unix(),
				},
				repository.NewCommonRepo().WhereByDeskUuid(deskSaleBill.DeskUuid),
			); err != nil {
				return errors.WithMessage(err)
			}

			// 关闭桌台 - 取消订单
			if err := repository.NewDeskRepo(tx).CloseDesk(deskSaleBill.DeskUuid, "合并桌台"); err != nil {
				return errors.WithMessage(err)
			}

			mealNum += deskSaleBill.MealNum
		}

		// 更新购买人数
		saleBill.MealNum = saleBill.MealNum + mealNum
		// 取消整单折扣
		if saleBill.SetAllDiscountCancel() {
			resp.IsResetDiscount = true
		}
		// 计算并保存销售账单
		if err := s.orderSrv.CalcAndSaveSaleBill(ctx, tx, saleBill); err != nil {
			return errors.WithMessage(err)
		}

		return nil
	}); err != nil {
		return nil, nil, errors.WithMessage(err)
	}

	// todo： 操作记录
	// deskNos
	// OrderOperationLog::createLog($this->order_id, OrderOperationLog::ACTION_MERGE_TABLE, $tableNos, '并台');

	// 返回购物车信息
	info, err := s.orderSrv.GetOrderCartInfo(ctx, req.SaleBillUuid)
	if err != nil {
		return nil, nil, errors.WithMessage(err, "获取订单信息失败")
	}
	resp.ShopCart = info

	//
	return resp, &deskMergeCheckRes, nil
}

// GetTabletDeskList 平板获取桌台列表
func (s *deskSrv) GetTabletDeskList(ctx context.Context) (resp.TabletDeskList, error) {
	deskRepo := repository.NewDeskRepo(s.dbm.GetDB(ctx.GetCompanyUuid()))
	desks, err := deskRepo.GetDesks(deskRepo.WhereIsNotDisable(), deskRepo.WhereUnBind())

	if err != nil {
		return resp.TabletDeskList{}, errors.WithMessage(err, "获取桌台列表失败")
	}
	list := make([]resp.TabletDeskItem, 0, len(desks))
	for _, desk := range desks {
		var item resp.TabletDeskItem
		copier.Copy(&item, desk)
		list = append(list, item)
	}
	return resp.TabletDeskList{
		List: list,
	}, nil
}

// BindDesk 平板获取桌台列表
func (s *deskSrv) BindDesk(ctx context.Context, bindDeskReq req.BindDeskReq) error {
	deviceUuid, err := s.deviceSrv.AddDevice(ctx, req.AddDeviceReq{
		DeviceId:         bindDeskReq.DeviceId,
		Brand:            bindDeskReq.Brand,
		Source:           constant.SourceTablet,
		FinallyLoginUuid: ctx.GetStaffUuid(),
		CompanyUuid:      ctx.GetCompanyUuid(),
		Remark:           bindDeskReq.Remark,
	})
	if err != nil {
		return errors.WithMessage(err)
	}

	if bindDeskReq.DeskUuid != bindDeskReq.OldDeskUuid {
		db := s.dbm.GetDB(ctx.GetCompanyUuid())
		deskRepo := repository.NewDeskRepo(db)
		desk, err := deskRepo.GetDesk(deskRepo.WhereUuid(bindDeskReq.DeskUuid))
		if desk.Uuid == 0 {
			return errors.New("桌台不存在")
		}
		if err != nil {
			return errors.WithMessage(err, "桌台不存在")
		}
		if desk.DeviceUuid > 0 && desk.DeviceUuid != deviceUuid {
			return errors.New("桌台已被占用")
		}
		err = db.Transaction(func(tx *gorm.DB) error {
			deskRepo = repository.NewDeskRepo(tx)
			// 解绑旧的桌台（解绑当前设备绑定了非选定的其他桌台）
			if err := deskRepo.UnbindDesk(bindDeskReq.DeskUuid, deviceUuid); err != nil {
				return errors.WithMessage(err)
			}
			// 绑定新的桌台
			if err := deskRepo.UpdateDeskByMap(desk.Uuid, map[string]any{"device_uuid": deviceUuid}); err != nil {
				return errors.WithMessage(err)
			}
			if bindDeskReq.OldDeskUuid != 0 { // 解绑旧桌台
				if err := deskRepo.UpdateDeskByMap(bindDeskReq.OldDeskUuid, map[string]any{"device_uuid": 0}); err != nil {
					return errors.WithMessage(err)
				}
			}
			return nil
		})
		if err != nil {
			return errors.WithMessage(err, "绑定桌台失败")
		}
	}
	return nil
}
