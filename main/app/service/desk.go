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

	"gorm.io/gorm"
)

// IDeskSrv 定义收银服务接口
type IDeskSrv interface {
	GetDeskList(ctx context.Context, dbId uint64, req req.DeskListReq) (resp.DeskListWithPaginationResp, error)         // 获取桌台列表
	GetDeskRegionAndTypeList(ctx context.Context) (resp.DeskRegionAndTypeListWithPaginationResp, error)                 // 获取桌台区域和类型列表
	GetDeskInfo(dbId uint64, deskUuid uint64) (resp.Desk, error)                                                        // 获取桌台详情
	GetDeskPing(ctx context.Context, deskUuid uint64, shopCart *resp.ShopCart) (resp.DeskPing, error)                   // 获取桌台详情-用于定时轮询
	GetH5DeskPing(ctx context.Context, deskUuid uint64, shopCart *resp.ShopCart) (resp.H5DeskPing, error)               // 获取桌台详情-用于h5定时轮询
	CreateDeskOrder(ctx context.Context, req req.DeskOrderCreateReq) (resp.CreateDeskOrderResp, error)                  // 创建桌台订单
	CloseDesk(ctx context.Context, req req.DeskCloseReq) error                                                          // 关闭桌台
	CompleteDesk(ctx context.Context, req req.DeskJsonUuidReq) error                                                    // 完成桌台
	ChangeDesk(ctx context.Context, req req.ChangeDeskReq) (*resp.ShopCart, error)                                      // 切换桌台
	MergeDesk(ctx context.Context, req req.MergeDeskReq) (*resp.DeskMergeShopCartResp, *resp.DeskMergeCheckResp, error) // 合并桌台
	IsCellCloseDesk(ctx context.Context, deskUuid uint64) (model.Desk, *resp.CartProductList, error)                    // 判断桌台是否可以关闭
	IsCellCloseInstant(ctx context.Context, saleBillUuid uint64) (*resp.CartProductList, error)                         // 判断订单是否可以关闭
	BindDesk(ctx context.Context, bindDeskReq req.BindDeskReq) (resp.Desk, error)                                       // 平板端绑定桌台
	ChangeBindDesk(ctx context.Context, changeBindDeskReq req.EditSettingReq) (resp.Desk, error)                        // 平板端换绑定桌台
}

// deskSrv 收银服务结构体
type deskSrv struct {
	bus         *event.SystemEventBus
	dbm         *database.DBManager // 数据库管理器
	localeSrv   ILocaleSrv          // 多语言名称服务
	orderSrv    IOrderSrv           // 订单服务
	settingSrv  setting.ISrv        // 设置服务
	deviceSrv   IDeviceSrv          // 设备服务
	mustPlanSrv IMustPlanSrv        // 必点方案服务
}

// NewDeskSrv 创建新的收银产品类别服务
func NewDeskSrv(dbm *database.DBManager, localeSrv ILocaleSrv, orderSrv IOrderSrv, settingSrv setting.ISrv, deviceSrv IDeviceSrv, mustPlanSrv IMustPlanSrv) IDeskSrv {
	return NewDeskSrvImpl(dbm, localeSrv, orderSrv, settingSrv, deviceSrv, mustPlanSrv)
}

// NewDeskSrvImpl 创建新的收银服务实现
func NewDeskSrvImpl(dbm *database.DBManager, localeSrv ILocaleSrv, orderSrv IOrderSrv, settingSrv setting.ISrv, deviceSrv IDeviceSrv, mustPlanSrv IMustPlanSrv) IDeskSrv {
	return &deskSrv{
		bus:         event.NewSystemBus(),
		dbm:         dbm,
		localeSrv:   localeSrv,
		orderSrv:    orderSrv,
		settingSrv:  settingSrv,
		deviceSrv:   deviceSrv,
		mustPlanSrv: mustPlanSrv,
	}
}

// GetDeskRegionAndTypeList 获取收银机点餐页面产品类别列表
func (s *deskSrv) GetDeskRegionAndTypeList(ctx context.Context) (resp.DeskRegionAndTypeListWithPaginationResp, error) {
	db := ctx.GetDB()
	companySetting := ctx.GetCompanySetting()

	// 获取列表
	repo := repository.NewDeskRegionRepo(db)
	regions, _ := repo.GetDeskRegionList(repo.WithDeskMapLayout())

	// 转换为响应对象
	deskRegionResp := make([]resp.DeskRegion, len(regions))
	for i, region := range regions {
		deskRegionResp[i] = resp.DeskRegion{
			Uuid: region.Uuid,
			Name: region.Name,
			IsOpenMap: func() bool {
				if companySetting.IsOpenTableMap() && region.DeskMapLayout != nil {
					return true
				}
				return false
			}(),
		}
	}

	// 获取桌台类型列表
	types, _ := repository.NewDeskTypeRepo(db).GetDeskTypeList()
	deskTypeResp := make([]resp.DeskType, len(types))
	for i, type_ := range types {
		deskTypeResp[i] = resp.DeskType{
			Uuid: type_.Uuid,
			Name: type_.Name,
		}
	}

	// 返回响应对象
	return resp.DeskRegionAndTypeListWithPaginationResp{
		UpdateTime: time.Now().Unix(),
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
	desks, total, err := repository.NewDeskRepo(s.dbm.GetDB(dbId)).GetClientDeskList(
		ctx.GetSource(),
		req.Status,
		req.IsBuffet,
		req.PageNo,
		req.PageSize,
	)
	if err != nil {
		return resp.DeskListWithPaginationResp{}, errors.WithMessage(err)
	}

	// 初始化额外信息
	// 转换为响应对象
	deskResp := make([]resp.Desk, len(desks))
	for i, desk := range desks {
		deskResp[i] = desk.GetDeskResp()
	}

	// 获取门店业务设置
	businessSetting, err := s.settingSrv.GetBusinessSetting(ctx)
	if err != nil {
		return resp.DeskListWithPaginationResp{}, errors.WithMessage(err)
	}

	var batchTagResp []resp.BatchTagRes
	if businessSetting.OpenIsBatch() {
		// 统计各个桌台的分批类型数量
		batchTagCountMap := make(map[uint64]uint)
		for _, desk := range desks {
			if desk.SaleBill != nil && desk.SaleBill.BatchTag != nil {
				batchTagCountMap[desk.SaleBill.BatchTag.Uuid]++
			}
		}

		// 获取分批类型列表
		batchTags, err := repository.NewBatchTagRepo(s.dbm.GetDB(ctx.GetDbId())).GetBatchTagList()
		if err != nil {
			return resp.DeskListWithPaginationResp{}, errors.WithMessage(err)
		}
		batchTagList := make([]resp.BatchTagRes, len(batchTags))
		for i, batchTag := range batchTags {
			batchTagList[i] = resp.BatchTagRes{
				Uuid:       batchTag.Uuid,
				LocaleName: batchTag.MultiLanguageName.GetNames(),
				Color:      batchTag.Color,
				Sort:       batchTag.Sort,
				Count:      batchTagCountMap[batchTag.Uuid],
			}
		}
		batchTagResp = batchTagList
	}
	// 返回响应对象
	return resp.DeskListWithPaginationResp{
		List: deskResp,
		Extra: resp.DeskExtra{
			UpdateTime: time.Now().Unix(),
			BatchTags:  batchTagResp,
		},
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

// GetDeskPing 获取桌台详情-用于定时轮询
func (s *deskSrv) GetDeskPing(ctx context.Context, deskUuid uint64, shopCart *resp.ShopCart) (resp.DeskPing, error) {
	res := resp.DeskPing{
		SentKitchen: resp.SentKitchen{
			Groups: resp.GroupList{
				List: make([]resp.SentKitchenProductGroup, 0),
			},
		},
		SentKitchenProducts: resp.SentKitchenProductList{
			List: make([]resp.SentKitchenProduct, 0),
		},
		MustPlans: resp.ProductMustPlanList{
			List: make([]resp.InstantProductMustPlan, 0),
		},
		SaleOrderList: make([]resp.SaleOrder, 0),
		UpdateTime:    time.Now().Unix(),
		OrderRemark: func() *resp.OrderRemarkRes {
			if shopCart != nil {
				return shopCart.OrderRemark
			}
			return nil
		}(),
	}
	// 获取桌台详情
	desk, err := repository.NewDeskRepo(ctx.GetDB()).GetDeskRecord(deskUuid)
	if err != nil {
		return res, errors.WithMessage(errors.New("桌台不存在"), "获取桌台详情失败")
	}
	// 获取销售账单信息并计算未送厨商品总金额
	saleBill, err := repository.NewOrderRepo(ctx.GetDB()).GetSaleBillAllInfo(desk.SaleBillUuid)
	if err != nil {
		return res, errors.WithMessage(errors.New("获取销售账单所有信息"), err.Error())
	}
	desk.SaleBill = saleBill
	res.DeskInfo = desk.GetDeskResp()

	// 如果没有销售账单,直接返回
	if desk.SaleBill == nil {
		return res, nil
	}
	// 设置国籍UUID
	res.NationalityUuid = desk.SaleBill.NationalityUuid
	// 获取账单信息，合计未送厨商品数量、合计已送厨商品列表
	if shopCart == nil {
		var opts []repository.OrderCartInfoOptionFunc
		if ctx.GetSource() == constant.SourceTablet { // 平板端查询购物车必点信息时，不自动加购
			opts = append(opts, repository.WithNoAutoAdd())
		}
		shopCart, err = s.orderSrv.GetOrderCartInfo(ctx, desk.SaleBillUuid, opts...)
		if err != nil {
			return res, errors.WithMessage(errors.New("订单不存在"), fmt.Sprintf("获取销售账单信息失败,SaleBillUuid: %d", desk.SaleBillUuid))
		}
	}
	// 未送厨商品信息
	res.UnsentKitchen, _ = s.orderSrv.GetUnsentKitchen(ctx, desk.SaleBillUuid, shopCart, saleBill)
	// 已送厨商品信息
	res.SentKitchen, _ = s.orderSrv.GetSentKitchen(ctx, desk.SaleBill.Uuid, shopCart, saleBill)
	// 自助餐信息
	if shopCart.Buffet != nil {
		res.Buffet = *shopCart.Buffet
	}
	// 必点方案列表
	if shopCart.MustPlans != nil {
		res.MustPlans = *shopCart.MustPlans
	}
	// 订单列表，拆单时，会有多个
	res.SaleOrderList = shopCart.SaleOrderList

	// 合计送厨商品数量和完成数量，退菜的商品不计算
	productPackageUuidMap := make(map[uint64]resp.SentKitchenProduct)
	for _, saleOrder := range shopCart.SaleOrderList {
		for _, product := range saleOrder.ProductList {
			// 合计已送厨商品列表
			if product.Status == constant.SaleOrderProductStatusCooking && !product.IsCancel && !(product.AboutBuffet.IsCustomer || product.AboutBuffet.IsDelay) && product.IsShowKitchen == 1 {
				var sentKitchenNum, finishedNum float64
				if existsProduct, exits := productPackageUuidMap[product.ProductPackageUuid]; exits {
					sentKitchenNum = existsProduct.SentKitchenNum + product.Num
					finishedNum = existsProduct.FinishedNum + product.FinishedNum
				} else {
					sentKitchenNum = product.Num
					finishedNum = product.FinishedNum
				}
				productPackageUuidMap[product.ProductPackageUuid] = resp.SentKitchenProduct{
					ProductPackageUuid: product.ProductPackageUuid,
					SentKitchenNum:     sentKitchenNum,
					FinishedNum:        finishedNum,
				}
			}
		}
	}

	// 转换成切片
	sentKitchenProducts := make([]resp.SentKitchenProduct, 0, len(productPackageUuidMap))
	for _, product := range productPackageUuidMap {
		sentKitchenProducts = append(sentKitchenProducts, product)
	}

	res.SentKitchenProducts = resp.SentKitchenProductList{
		List: sentKitchenProducts,
	}
	res.OrderRemark = shopCart.OrderRemark
	return res, nil
}

// GetH5DeskPing 获取桌台详情-用于h5定时轮询
func (s *deskSrv) GetH5DeskPing(ctx context.Context, deskUuid uint64, shopCart *resp.ShopCart) (resp.H5DeskPing, error) {
	res := resp.H5DeskPing{
		SentKitchen: resp.H5CartSendProduct{
			Groups: resp.H5GroupList{
				List: make([]resp.H5Group, 0),
			},
		},
		MustPlans: resp.ProductMustPlanList{
			List: make([]resp.InstantProductMustPlan, 0),
		},
		UpdateTime: time.Now().Unix(),
	}
	// 获取桌台详情
	desk, err := repository.NewDeskRepo(ctx.GetDB()).GetDeskInfo(deskUuid)
	if err != nil {
		return res, errors.WithMessage(errors.New("桌台不存在"), "获取桌台详情失败")
	}
	res.DeskInfo = desk.GetDeskResp()

	// 如果没有销售账单,直接返回
	if desk.SaleBill == nil {
		return res, nil
	}
	// 获取账单信息，合计未送厨商品数量
	if shopCart == nil {
		shopCart, err = s.orderSrv.GetOrderCartInfo(ctx, desk.SaleBillUuid, repository.WithUnorderedH5Product(), repository.WithH5AutoAdd())
		if err != nil {
			return res, errors.WithMessage(errors.New("订单不存在"), "获取销售账单信息失败")
		}
	}
	// 未送厨商品信息
	unsentKitchen, err := s.orderSrv.GetUnOrderedH5ProductList(ctx, desk.SaleBillUuid, shopCart, repository.WithUnorderedH5Product())
	if err != nil {
		return res, errors.WithMessage(errors.New("订单不存在"), "获取未送厨商品信息失败")
	}
	res.UnsentKitchen = *unsentKitchen

	// 自助餐信息
	if shopCart.Buffet != nil {
		res.Buffet = *shopCart.Buffet
	}
	// 必点方案列表
	if desk.SaleBill.IsShowMustPlan() {
		// 查询到购物车信息
		shopCartInfo, err := repository.NewOrderRepo(ctx.GetDB()).GetOrderCartInfo(shopCart.SaleBillUuid, repository.WithNotDeleted())
		if err != nil {
			return resp.H5DeskPing{}, errors.WithMessage(err)
		}
		if shopCartInfo.SaleBill.IsShowMustPlan() {
			var deskMustPlanList []resp.InstantProductMustPlan
			var err error
			if shopCartInfo.SaleBill.IsBuffetSaleBill() {
				deskMustPlanList, err = s.mustPlanSrv.GetDeskMustPlanList(ctx, shopCartInfo.SaleBill.MealNum, shopCartInfo.GetMustPlanProductInfo(), deskUuid, WithSaleBillUuid(desk.SaleBillUuid))
			} else {
				deskMustPlanList, err = s.mustPlanSrv.GetDeskMustPlanList(ctx, shopCartInfo.SaleBill.MealNum, shopCartInfo.GetMustPlanProductInfo(), deskUuid)
			}
			if err != nil {
				return resp.H5DeskPing{}, errors.WithMessage(err)
			}
			res.MustPlans = resp.ProductMustPlanList{
				List: deskMustPlanList,
			}
		}
	}
	// 必点商品列表
	mustProducts, err := s.mustPlanSrv.GetDeskMustPlanProductPackageList(ctx, deskUuid)
	if err != nil {
		return res, errors.WithMessage(err)
	}
	if len(mustProducts) > 0 {
		mustProductList := make([]resp.BuffetProduct, 0, len(mustProducts))
		for _, mustProduct := range mustProducts {
			mustProductList = append(mustProductList, resp.BuffetProduct{
				Uuid: mustProduct.Uuid,
				Name: mustProduct.Name,
			})
		}
		res.MustProducts = resp.BuffetProductList{
			List: mustProductList,
		}
	}
	res.OrderRemark = shopCart.OrderRemark
	return res, nil
}

// CreateDeskOrder 创建桌台订单
func (s *deskSrv) CreateDeskOrder(ctx context.Context, req req.DeskOrderCreateReq) (resp.CreateDeskOrderResp, error) {
	dbId := ctx.GetDbId()
	companySetting := ctx.GetCompanySetting()
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
	if desk.IsDisabled {
		return resp.CreateDeskOrderResp{}, errors.New("桌台未开启")
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
		if companySetting.IsOpenH5 != 1 {
			return resp.CreateDeskOrderResp{}, errors.New("当前未开启扫码点餐功能，请联系销售代表")
		}
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
		if companySetting.IsOpenBuffet != 1 {
			return resp.CreateDeskOrderResp{}, errors.New("当前尚未开启自助餐功能，如有需要，请联系销售代表")
		}
		buffetSetting, _ := s.settingSrv.GetBuffetSetting(ctx, companySetting)
		if buffetSetting.IsOpen != "1" {
			return resp.CreateDeskOrderResp{}, errors.New("未开启自助餐")
		}
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
	utils.Go(func() {
		s.bus.PublishOpenDeskEvent(event.OpenDeskPayload{
			BasePayload: event.BasePayload{ // 开台
				Ctx:           ctx,
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
	})
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
		if buffetCustomerType.MealNum == 0 {
			// 如果顾客类型没有指定用餐人数，则跳过
			continue
		}
		_, err := repository.NewBuffetRepo(s.dbm.GetDB(dbId)).GetBuffetCustomerTypeInfo(repository.NewCommonRepo().WhereByUuid(buffetCustomerType.Uuid))
		if err != nil {
			return resp.CreateDeskOrderResp{}, errors.WithMessage(err, "自助餐顾客类型不存在")
		}
	}
	// 验证自助餐套餐是否选择了自助餐顾客类型
	{
		buffetUuids := make([]uint64, 0)
		for _, buffetUuid := range req.BuffetUuids {
			selectedBuffet := false // 是否选择这个自助餐套餐，默认是不选择，当选择的顾客类型中任意一个是该自助餐套餐的顾客类型时，则选择该自助餐套餐
			for _, buffetCustomerType := range req.BuffetCustomerTypes {
				if buffetCustomerType.MealNum == 0 {
					// 如果顾客类型没有指定用餐人数，则跳过
					continue
				}
				_, exist, err := repository.NewBuffetRepo(s.dbm.GetDB(dbId)).GetBuffetCustomerTypePriceInfo(buffetUuid, buffetCustomerType.Uuid)
				if err != nil {
					return resp.CreateDeskOrderResp{}, errors.WithMessage(err, "获取自助餐顾客类型价格信息失败")
				}
				// 如果存在，则选择该自助餐套餐
				if exist {
					selectedBuffet = true
				}
			}
			// 如果选择了该自助餐套餐，则添加到列表中
			if selectedBuffet {
				buffetUuids = append(buffetUuids, buffetUuid)
			}
		}
		req.BuffetUuids = buffetUuids
	}

	return s.orderSrv.CreateDeskOrder(ctx, req)
}

// IsCellCloseDesk 判断桌台是否可关闭
func (s *deskSrv) IsCellCloseDesk(ctx context.Context, deskUuid uint64) (model.Desk, *resp.CartProductList, error) {
	if ctx.NoLock() {
		lock.NewSystemLock().LockUuid(deskUuid)
		defer lock.NewSystemLock().UnlockUuid(deskUuid)
		ctx.AddLock()
	}
	dbId := ctx.GetDbId()
	db := s.dbm.GetDB(dbId)
	desk, err := repository.NewDeskRepo(db).GetDeskInfo(deskUuid)
	if err != nil {
		return model.Desk{}, nil, errors.WithMessage(err, "桌台不存在")
	}
	if desk.Status == 0 {
		return model.Desk{}, nil, errors.New("桌台已关闭")
	}
	if desk.SaleBillUuid == 0 {
		return model.Desk{}, nil, errors.New("桌台已关闭")
	}
	// 判断是否可立即关闭
	productList, err := s.IsCellCloseInstant(ctx, desk.SaleBillUuid)
	if err != nil {
		return model.Desk{}, productList, errors.WithMessage(err)
	}
	//
	return desk, nil, nil
}

// IsCellCloseInstant 判断桌台是否可关闭
func (s *deskSrv) IsCellCloseInstant(ctx context.Context, saleBillUuid uint64) (*resp.CartProductList, error) {
	billInfo, err := s.orderSrv.IsCellCancelOrder(ctx, saleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	if billInfo.IsPartialPay() {
		return nil, errors.New("当前订单已被部分支付，不支持取消")
	}
	// 有商品已送厨
	if productCooking := billInfo.GetSaleOrderProductCooking(); len(productCooking) > 0 {
		productList := make([]resp.Product, 0, len(productCooking))
		for _, product := range productCooking {
			if product.IsDelete() || product.IsCancelProduct() {
				// 如果商品已经删除，则跳过
				// 如果商品已经退菜，则跳过
				continue
			}

			productList = append(productList, resp.Product{
				Uuid:          product.Uuid,
				LocaleName:    product.MultiLanguageName.GetNames(),
				Num:           product.Num,
				SalePrice:     product.SalePrice,
				DiscountPrice: product.Price,
				Status:        int(product.Status),
				Remark:        product.Remark,
				IsMust:        product.IsMustProduct(),
				IsGift:        product.IsGiftProduct(),
				IsBuffet:      product.IsBuffet == 1,
				IsCancel:      product.IsCancelProduct(),
			})
		}
		if len(productList) == 0 {
			return nil, nil
		}
		return &resp.CartProductList{
			List: productList,
		}, errors.New("此单有商品已送厨，是否取消此笔交易？")
	}
	return nil, nil
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
	db := s.dbm.GetDB(ctx.GetCompanyUuid())

	// 禁止并发操作：在方法开头就加锁，确保获取目标桌台的订单UUID时使用的是最新数据
	if ctx.NoLock() {
		systemLock := lock.NewSystemLock()

		// 收集需要锁定的资源：源订单 + 目标资源（目标桌台或目标订单）
		lockUuids := []uint64{reqs.SaleBillUuid}

		// 获取目标桌台的订单UUID（在锁内获取，确保数据一致性）
		// 如果目标桌台有订单，则锁定目标订单；否则锁定目标桌台
		targetSaleBillUuid, err := repository.NewDeskRepo(db).GetSaleBillUuidByDeskUuid(reqs.DeskUuid)
		if err == nil && targetSaleBillUuid != 0 {
			// 目标桌台有订单，锁定目标订单
			lockUuids = append(lockUuids, targetSaleBillUuid)
		} else {
			// 目标桌台没有订单，锁定目标桌台
			lockUuids = append(lockUuids, reqs.DeskUuid)
		}

		// 锁定源订单和目标资源（按 UUID 排序）
		// LockMultipleUuids 会自动去重和排序，返回排序后的 UUID 列表
		lockedUuids := lock.LockMultipleUuids(systemLock, lockUuids)

		// 按相反顺序释放锁（UnlockMultipleUuids 内部会使用相同的排序策略）
		defer func() {
			lock.UnlockMultipleUuids(systemLock, lockedUuids)
		}()

		ctx.AddLock()
	}

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

	// 获取桌台信息（在锁内获取，确保数据一致性）
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
		// 将该saleBill的所有h5订单都进行转台
		if err := repository.NewH5OrderRepo(tx).Update(
			map[string]interface{}{
				"desk_uuid": reqs.DeskUuid,
				"desk_no":   desk.DeskNo,
			},
			repository.NewCommonRepo().WhereBySaleBillUuid(reqs.SaleBillUuid),
		); err != nil {
			return errors.WithMessage(err)
		}
		return nil
	}); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 发布转台事件
	utils.Go(func() {
		ctx.Log().Info("发布转台事件")
		s.bus.PublishChangeDeskEvent(event.ChangeDeskPayload{
			BasePayload: event.BasePayload{ // 转台
				Ctx:           ctx,
				CompanyUuid:   ctx.GetCompanyUuid(),
				Source:        ctx.GetSource(),
				SaleBillUuid:  reqs.SaleBillUuid,
				SaleOrderUuid: reqs.SaleOrderUuid,
				OperatorUuid:  int64(ctx.GetStaffUuid()),
			},
			Old: event.Table{TableId: oldDesk.Uuid, TableNo: oldDesk.DeskNo},
			New: event.Table{TableId: reqs.DeskUuid, TableNo: desk.DeskNo},
		})
	})

	// 返回购物车信息
	info, err := s.orderSrv.GetOrderCartInfo(ctx, reqs.SaleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err, "获取订单信息失败")
	}

	return info, nil
}

// MergeDesk 合并桌台
func (s *deskSrv) MergeDesk(ctx context.Context, req req.MergeDeskReq) (*resp.DeskMergeShopCartResp, *resp.DeskMergeCheckResp, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, errors.WithMessage(err)
	}

	db := s.dbm.GetDB(ctx.GetCompanyUuid())

	// 禁止并发操作：在方法开头就加锁，确保获取被合并桌台的订单UUID时使用的是最新数据
	if ctx.NoLock() {
		systemLock := lock.NewSystemLock()
		// 收集所有需要锁定的订单 UUID（主订单 + 所有被合并的订单）
		orderUuids := []uint64{req.SaleBillUuid}

		// 收集被合并桌台的订单 UUID（在锁内获取，确保数据一致性）
		for _, deskUuid := range req.DeskUuids {
			// 通过桌台UUID获取订单UUID（在锁内获取）
			targetSaleBillUuid, err := repository.NewDeskRepo(db).GetSaleBillUuidByDeskUuid(deskUuid)
			if err == nil && targetSaleBillUuid != 0 {
				orderUuids = append(orderUuids, targetSaleBillUuid)
			}
		}

		// 锁定所有涉及的订单（按 UUID 排序）
		// LockMultipleUuids 会自动去重和排序，返回排序后的 UUID 列表
		lockedUuids := lock.LockMultipleUuids(systemLock, orderUuids)

		// 按相反顺序释放锁（UnlockMultipleUuids 内部会使用相同的排序策略）
		defer func() {
			lock.UnlockMultipleUuids(systemLock, lockedUuids)
		}()
		ctx.AddLock()
	}

	// 获取销售账单信息（用于后续业务逻辑）
	saleBill, errSaleBill := repository.NewOrderRepo(db).GetSaleBillAllInfo(req.SaleBillUuid)
	if errSaleBill != nil {
		return nil, nil, errors.WithMessage(errSaleBill, "获取销售账单信息失败")
	}
	if saleBill.IsCanceled() {
		return nil, nil, errors.New("当前订单已取消，不支持合并桌台")
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
		// 跳过当前桌台，避免关闭当前桌台
		if deskUuid == saleBill.DeskUuid {
			continue
		}
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
		// 更新销售账单的首次送厨时间
		if saleBill.ProductionTime == 0 || saleBill.ProductionTime > deskSaleBill.ProductionTime {
			if deskSaleBill.ProductionTime > 0 {
				saleBill.ProductionTime = deskSaleBill.ProductionTime
			}
		}
	}
	if len(deskMergeCheckRes.List) > 0 {
		return nil, &deskMergeCheckRes, errors.New(errDeskMsg)
	}

	// 如果没有需要合并的桌台，直接返回
	if len(saleBillList) == 0 {
		return nil, nil, errors.New("没有可合并的桌台")
	}

	resp := &resp.DeskMergeShopCartResp{}

	// 更新桌台信息
	if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		mealNum := uint(0)
		// 更新销售订单
		for _, deskSaleBill := range saleBillList {
			deskSaleOrder := deskSaleBill.GetFirstSaleOrder()
			saleBillOpt := repository.NewCommonRepo().WhereBySaleBillUuid(deskSaleBill.Uuid)
			saleOrderOpt := repository.NewCommonRepo().WhereBySaleOrderUuid(deskSaleOrder.Uuid)

			// 更新销售订单商品
			if err := repository.NewSaleOrderProductRepo(tx).Update(
				map[string]interface{}{
					"desk_uuid":       saleBill.DeskUuid,
					"sale_bill_uuid":  saleBill.Uuid,
					"sale_order_uuid": saleOrder.Uuid,
					"create_time":     time.Now().Unix(),
				},
				saleBillOpt,
				saleOrderOpt,
			); err != nil {
				return errors.WithMessage(err)
			}

			// 更新送厨单和商品
			productionRepo := repository.NewProductionRepo(tx)
			if err := productionRepo.UpdateProduct([]repository.DBOption{saleBillOpt}, map[string]any{
				"sale_bill_uuid": saleBill.Uuid,
			}); err != nil {
				return errors.WithMessage(err)
			}
			if err := productionRepo.UpdateOrder([]repository.DBOption{saleBillOpt, saleOrderOpt}, map[string]any{
				"desk_uuid":       saleBill.DeskUuid,
				"sale_bill_uuid":  saleBill.Uuid,
				"sale_order_uuid": saleOrder.Uuid,
			}); err != nil {
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
					"desk_uuid":       saleBill.DeskUuid,
					"desk_no":         saleBill.Desk.DeskNo,
					"sale_bill_uuid":  saleBill.Uuid,  // 待接单记录改成当前销售账单
					"sale_order_uuid": saleOrder.Uuid, // 待接单记录改成当前销售订单
				},
				repository.NewCommonRepo().WhereByDeskUuid(deskSaleBill.DeskUuid),
			); err != nil {
				return errors.WithMessage(err)
			}

			// 关闭桌台 - 取消订单
			if err := repository.NewDeskRepo(tx).CloseDesk(ctx, deskSaleBill.DeskUuid, deskSaleBill.Uuid, "合并桌台"); err != nil {
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

		// 重新设置会员折扣
		if saleOrder.Member != nil {
			saleOrder.SetMemberDiscount(*saleOrder.Member)
		} else {
			saleOrder.SetMemberDiscountCancel()
		}

		// 计算并保存销售账单
		if err := s.orderSrv.CalcAndSaveSaleBill(ctx, tx, saleBill); err != nil {
			return errors.WithMessage(err)
		}

		return nil
	}); err != nil {
		return nil, nil, errors.WithMessage(err)
	}
	// 发布“并台”操作事件
	utils.Go(func() {
		s.bus.PublishMergeDeskEvent(event.MergeDeskPayload{
			BasePayload: event.BasePayload{ // 并台
				Ctx:          ctx,
				CompanyUuid:  ctx.GetCompanyUuid(),
				Source:       ctx.GetSource(),
				SaleBillUuid: saleBill.Uuid,
				OperatorUuid: int64(ctx.GetStaffUuid()),
			},
			DeskNos: deskNos,
		})
	})
	// 返回购物车信息
	info, err := s.orderSrv.GetOrderCartInfo(ctx, req.SaleBillUuid)
	if err != nil {
		return nil, nil, errors.WithMessage(err, "获取订单信息失败")
	}
	resp.ShopCart = info

	//
	return resp, &deskMergeCheckRes, nil
}

// BindDesk 平板绑定桌台
func (s *deskSrv) BindDesk(ctx context.Context, bindDeskReq req.BindDeskReq) (resp.Desk, error) {
	companyUuid := ctx.GetCompanyUuid()
	deviceUuid, err := s.deviceSrv.AddDevice(ctx, req.AddDeviceReq{
		DeviceId:         ctx.GetDeviceSn(),
		Source:           constant.SourceTablet,
		FinallyLoginUuid: ctx.GetStaffUuid(),
		CompanyUuid:      companyUuid,
	})
	if err != nil {
		return resp.Desk{}, errors.WithMessage(err)
	}
	db := s.dbm.GetDB(companyUuid)
	// 当前设备已经绑定桌台
	deskRepo := repository.NewDeskRepo(db)
	desk, _ := deskRepo.GetDesk(deskRepo.WhereDeviceUuid(deviceUuid), deskRepo.WhereIsDisable(constant.DeskEnable))
	if desk.Uuid != 0 {
		return resp.Desk{}, errors.WithMessage(errors.New("已绑定桌台"))
	}
	// 桌台已被占用
	desk, err = deskRepo.GetDesk(deskRepo.WhereUuid(bindDeskReq.DeskUuid), deskRepo.WhereIsDisable(constant.DeskEnable))
	if err != nil || desk.Uuid == 0 {
		return resp.Desk{}, errors.WithMessage(errors.New("桌台不存在"))
	}
	if desk.DeviceUuid > 0 && desk.DeviceUuid != deviceUuid {
		return resp.Desk{}, errors.New("桌台已被占用")
	}
	// 绑定桌台
	if err := deskRepo.UpdateDeskByMap(desk.Uuid, map[string]any{"device_uuid": deviceUuid}); err != nil {
		return resp.Desk{}, errors.WithMessage(errors.New("绑定桌台失败"))
	}

	return s.GetDeskInfo(companyUuid, bindDeskReq.DeskUuid)
}

// ChangeBindDesk 平板换绑定桌台
func (s *deskSrv) ChangeBindDesk(ctx context.Context, changeBindDeskReq req.EditSettingReq) (resp.Desk, error) {
	dbId := ctx.GetDbId()
	db := s.dbm.GetDB(dbId)
	deviceUuid := ctx.GetDeviceUuid()
	// 当前设备未绑定桌台，不能换绑
	deskRepo := repository.NewDeskRepo(db)
	oldDesk, _ := deskRepo.GetDesk(deskRepo.WhereDeviceUuid(deviceUuid), deskRepo.WhereIsDisable(constant.DeskEnable))
	if oldDesk.Uuid == 0 {
		return resp.Desk{}, errors.WithMessage(errors.New("未绑定桌台"))
	}
	// 已绑定的桌台和传递桌台一样，不做操作
	if oldDesk.Uuid == changeBindDeskReq.DeskUuid {
		if err := s.deviceSrv.UpdateRemark(ctx, req.EditDeviceRemarkReq{Remark: changeBindDeskReq.Remark}); err != nil {
			return resp.Desk{}, errors.WithMessage(errors.New("修改机器备注失败"), err.Error())
		}
		return s.GetDeskInfo(dbId, changeBindDeskReq.DeskUuid)
	}
	// 新桌台已被占用
	newDesk, err := deskRepo.GetDesk(deskRepo.WhereUuid(changeBindDeskReq.DeskUuid), deskRepo.WhereIsDisable(constant.DeskEnable))
	if err != nil || newDesk.Uuid == 0 {
		return resp.Desk{}, errors.WithMessage(errors.New("桌台不存在"))
	}
	if newDesk.DeviceUuid > 0 {
		return resp.Desk{}, errors.New("桌台已被占用")
	}
	err = db.Transaction(func(tx *gorm.DB) error {
		ctx.SetDB(tx)
		if err := s.deviceSrv.UpdateRemark(ctx, req.EditDeviceRemarkReq{Remark: changeBindDeskReq.Remark}); err != nil {
			return err
		}
		deskRepo = repository.NewDeskRepo(tx)
		// 绑定桌台
		if err := deskRepo.UpdateDeskByMap(newDesk.Uuid, map[string]any{"device_uuid": deviceUuid}); err != nil {
			return err
		}
		// 解绑桌台
		if err := deskRepo.UpdateDeskByMap(oldDesk.Uuid, map[string]any{"device_uuid": 0}); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return resp.Desk{}, errors.WithMessage(errors.ErrInternal, err.Error())
	}
	return s.GetDeskInfo(dbId, changeBindDeskReq.DeskUuid)
}
