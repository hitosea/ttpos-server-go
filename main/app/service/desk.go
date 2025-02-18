package service

import (
	"errors"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
)

// IDeskSrv 定义收银服务接口
type IDeskSrv interface {
	GetDeskList(dbId uint64, req req.DeskListReq) (resp.DeskListWithPaginationResp, error)             // 获取桌台列表
	GetDeskRegionAndTypeList(dbId uint64) (resp.DeskRegionAndTypeListWithPaginationResp, error)        // 获取桌台区域和类型列表
	GetDeskInfo(dbId uint64, deskUuid uint64) (resp.DeskInfoResp, error)                               // 获取桌台详情
	CreateDeskOrder(ctx context.Context, req req.DeskOrderCreateReq) (resp.CreateDeskOrderResp, error) // 创建桌台订单
	CloseDesk(ctx context.Context, req req.DeskCloseReq) error                                         // 关闭桌台
	IsCellCloseDesk(ctx context.Context, deskUuid uint64) (model.Desk, error)                          // 判断桌台是否可以关闭
}

// deskSrv 收银服务结构体
type deskSrv struct {
	dbm        *database.DBManager // 数据库管理器
	localeSrv  ILocaleSrv          // 多语言名称服务
	orderSrv   IOrderSrv           // 订单服务
	settingSrv setting.ISrv        // 设置服务
}

// NewDeskSrv 创建新的收银产品类别服务
func NewDeskSrv(dbm *database.DBManager, localeSrv ILocaleSrv, orderSrv IOrderSrv, settingSrv setting.ISrv) IDeskSrv {
	return NewDeskSrvImpl(dbm, localeSrv, orderSrv, settingSrv)
}

// NewDeskSrvImpl 创建新的收银服务实现
func NewDeskSrvImpl(dbm *database.DBManager, localeSrv ILocaleSrv, orderSrv IOrderSrv, settingSrv setting.ISrv) IDeskSrv {
	return &deskSrv{
		dbm:        dbm,
		localeSrv:  localeSrv,
		orderSrv:   orderSrv,
		settingSrv: settingSrv,
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
func (s *deskSrv) GetDeskList(dbId uint64, req req.DeskListReq) (resp.DeskListWithPaginationResp, error) {
	// 获取列表
	desks, total, err := repository.NewDeskRepo(s.dbm.GetDB(dbId)).GetClientDeskList(req.PageNo, req.PageSize)
	if err != nil {
		return resp.DeskListWithPaginationResp{}, err
	}

	// 初始化额外信息
	var extra = resp.DeskExtra{
		AvailableNum:       0,
		LockNum:            0,
		OccupyBuffetNum:    0,
		OccupyNotBuffetNum: 0,
		OccupyWaitNum:      0,
		TotalNum:           0,
	}

	// 转换为响应对象
	deskResp := make([]resp.Desk, len(desks))
	for i, desk := range desks {
		// 桌台状态	0:空闲 1:非自助餐 2:自助餐 3:待清台 4:锁单
		var deskStatus uint
		//var elapsedTime uint
		//if desk.SaleBill.ID == 0 {
		//	if desk.Status == 1 {
		//		deskStatus = constant.DeskStatusWait
		//		extra.OccupyWaitNum++
		//	} else {
		//		deskStatus = constant.DeskStatusAvailable
		//		extra.AvailableNum++
		//	}
		//	elapsedTime = constant.DeskStatusAvailable
		//} else {
		//	extra.OccupyWaitNum++
		//	//
		//	if desk.SaleBill.IsLock == 1 {
		//		deskStatus = constant.DeskStatusLock
		//		extra.LockNum++
		//	} else if desk.SaleBill.IsBuffet == 1 {
		//		deskStatus = constant.DeskStatusBuffet
		//	} else {
		//		deskStatus = constant.DeskStatusNotBuffet
		//	}
		//	//
		//	if desk.SaleBill.IsBuffet == 1 {
		//		extra.OccupyBuffetNum++
		//	} else {
		//		extra.OccupyNotBuffetNum++
		//	}
		//	// 如果是自助餐，计算剩余时间; 非自助餐，显示已用时间
		//	passedTime := time.Now().Unix() - desk.SaleBill.CreateTime
		//	if desk.SaleBill.IsBuffet == 1 {
		//		if uint(passedTime) >= desk.SaleBill.BuffetDuration {
		//			elapsedTime = 0
		//		} else {
		//			elapsedTime = desk.SaleBill.BuffetDuration - uint(passedTime)
		//		}
		//	} else {
		//		elapsedTime = uint(passedTime)
		//	}
		//}
		//
		extra.TotalNum++
		//
		// todo  desk.SaleBill.PaymentAmount 需要等后面业务缓存中取
		deskResp[i] = resp.Desk{
			Uuid:       desk.Uuid,
			DeskNo:     desk.DeskNo,
			TypeUuid:   desk.TypeUuid,
			RegionUuid: desk.RegionUuid,
			Status:     deskStatus,
			//CustomerCount: desk.SaleBill.MealNum,
			//IsLock:        desk.SaleBill.IsLock == 1,
			//IsBuffet:      desk.SaleBill.IsBuffet == 1,
			//Remark:        desk.SaleBill.Remark,
			//Time:          elapsedTime,
			//Price:         desk.SaleBill.PaymentAmount,
		}
	}

	// 返回响应对象
	return resp.DeskListWithPaginationResp{
		Extra: extra,
		List:  deskResp,
		Meta: dto.PageResponse{
			PageNo:   req.PageNo,
			PageSize: req.PageSize,
			Total:    total,
		},
	}, nil
}

// GetDeskInfo 获取桌台详情
func (s *deskSrv) GetDeskInfo(dbId uint64, deskUuid uint64) (resp.DeskInfoResp, error) {
	// 获取列表
	desk, err := repository.NewDeskRepo(s.dbm.GetDB(dbId)).GetDeskInfo(deskUuid)
	if err != nil {
		return resp.DeskInfoResp{}, err
	}
	// 转换为响应对象
	//var deskStatus uint
	//var elapsedTime uint
	//if desk.SaleBill.ID == 0 {
	//	if desk.Status == 1 {
	//		deskStatus = constant.DeskStatusWait
	//	} else {
	//		deskStatus = constant.DeskStatusAvailable
	//	}
	//	elapsedTime = constant.DeskStatusAvailable
	//} else {
	//	//
	//	if desk.SaleBill.IsLock == 1 {
	//		deskStatus = constant.DeskStatusLock
	//	} else if desk.SaleBill.IsBuffet == 1 {
	//		deskStatus = constant.DeskStatusBuffet
	//	} else {
	//		deskStatus = constant.DeskStatusNotBuffet
	//	}
	//	// 如果是自助餐，计算剩余时间; 非自助餐，显示已用时间
	//	passedTime := time.Now().Unix() - desk.SaleBill.CreateTime
	//	if desk.SaleBill.IsBuffet == 1 {
	//		if uint(passedTime) >= desk.SaleBill.BuffetDuration {
	//			elapsedTime = 0
	//		} else {
	//			elapsedTime = desk.SaleBill.BuffetDuration - uint(passedTime)
	//		}
	//	} else {
	//		elapsedTime = uint(passedTime)
	//	}
	//}
	//
	return resp.DeskInfoResp{
		Uuid: desk.Uuid,
		//SaleBillUuid: desk.SaleBill.Uuid,
		//DeskNo:       desk.DeskNo,
		//TypeUuid:     desk.TypeUuid,
		//RegionUuid:   desk.RegionUuid,
		//Status:       deskStatus,
		//IsLock:       desk.SaleBill.IsLock == 1,
		//IsBuffet:     desk.SaleBill.IsBuffet == 1,
		//Remark:       desk.SaleBill.Remark,
		//Time:         elapsedTime,
	}, nil
}

// CreateDeskOrder 创建桌台订单
func (s *deskSrv) CreateDeskOrder(ctx context.Context, req req.DeskOrderCreateReq) (resp.CreateDeskOrderResp, error) {
	dbId := ctx.GetDbId()
	// 验证请求参数
	err := s.validateCreateDeskOrderReq(req)
	if err != nil {
		return resp.CreateDeskOrderResp{}, err
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

	// 判断是否自助餐订单
	if !*req.IsBuffet {
		if req.MealNum == nil || *req.MealNum == 0 {
			return resp.CreateDeskOrderResp{}, errors.New("就餐人数不能小于1")
		}
	} else {
		return s.createDeskBuffetOrder(ctx, req)
	}

	// 创建桌台-非自助餐订单
	return s.orderSrv.CreateDeskOrder(ctx, req)
}

// validateCreateDeskOrderReq 验证请求参数
func (s *deskSrv) validateCreateDeskOrderReq(req req.DeskOrderCreateReq) error {
	if req.DeskUuid == 0 {
		return errors.New("桌台uuid不能为0")
	}
	if req.IsBuffet == nil {
		return errors.New("是否是自助餐不能为空")
	}
	if req.MealNum == nil {
		return errors.New("就餐人数不能为空")
	}
	if !*req.IsBuffet {
		if *req.MealNum < 1 || *req.MealNum > 999 {
			return errors.New("就餐人数不能小于1或大于999")
		}
	}
	if *req.IsBuffet {
		if len(req.BuffetUuids) < 1 || len(req.BuffetUuids) > 2 {
			return errors.New("自助餐uuid列表不能小于1或大于2")
		}
		if len(req.BuffetCustomerTypes) == 0 {
			return errors.New("自助餐顾客类型列表不能为空")
		}
	}

	return nil
}

// createDeskBuffetOrder 创建桌台自助餐订单
func (s *deskSrv) createDeskBuffetOrder(ctx context.Context, req req.DeskOrderCreateReq) (resp.CreateDeskOrderResp, error) {
	dbId := ctx.GetDbId()
	// 验证自助餐是否开启
	companySetting, err := s.settingSrv.GetCompanySetting(ctx)
	if err != nil {
		return resp.CreateDeskOrderResp{}, err
	}
	buffetSetting, err := s.settingSrv.GetBuffetSetting(ctx, companySetting)
	if err != nil {
		return resp.CreateDeskOrderResp{}, err
	}
	if buffetSetting.IsOpen == "0" {
		return resp.CreateDeskOrderResp{}, errors.New("自助餐未开启")
	}

	// 验证自助餐套餐是否存在
	for _, buffetUuid := range req.BuffetUuids {
		_, err := repository.NewBuffetRepo(s.dbm.GetDB(dbId)).GetBuffetInfo(repository.NewCommonRepo().WhereByUuid(buffetUuid))
		if err != nil {
			return resp.CreateDeskOrderResp{}, errors.New("自助餐套餐不存在")
		}
	}

	// 验证自助餐顾客类型是否存在
	for _, buffetCustomerType := range req.BuffetCustomerTypes {
		_, err := repository.NewBuffetRepo(s.dbm.GetDB(dbId)).GetBuffetCustomerTypeInfo(repository.NewCommonRepo().WhereByUuid(buffetCustomerType.Uuid))
		if err != nil {
			return resp.CreateDeskOrderResp{}, errors.New("自助餐顾客类型不存在")
		}
	}

	return s.orderSrv.CreateDeskOrder(ctx, req)
}

// 判断桌台是否可关闭
func (s *deskSrv) IsCellCloseDesk(ctx context.Context, deskUuid uint64) (model.Desk, error) {
	dbId := ctx.GetDbId()
	db := s.dbm.GetDB(dbId)
	desk, err := repository.NewDeskRepo(db).GetDeskInfo(deskUuid)
	if err != nil {
		return model.Desk{}, errors.New("桌台不存在")
	}
	if desk.Status == 0 {
		return model.Desk{}, errors.New("桌台已关闭")
	}
	//if desk.SaleBill.ID == 0 {
	//	return model.Desk{}, nil
	//}
	//if _, err := NewOrderSrv(s.dbm, nil, nil).IsCellCancelOrder(ctx, desk.SaleBill.Uuid); err != nil {
	//	return model.Desk{}, err
	//}
	return desk, nil
}

// 关闭桌台
func (s *deskSrv) CloseDesk(ctx context.Context, reqs req.DeskCloseReq) error {
	//dbId := ctx.GetDbId()
	//db := s.dbm.GetDB(dbId)
	//desk, err := repository.NewDeskRepo(db).GetDeskInfo(reqs.Uuid)
	//if err != nil {
	//	return err
	//}
	//if desk.SaleBill.ID == 0 {
	//	return nil
	//}
	//// 取消订单
	//if err := NewOrderSrv(s.dbm, nil, nil).CancelOrder(ctx, req.OrderCancelReq{
	//	SaleBillUuid: desk.SaleBill.Uuid,
	//	CancelReason: reqs.Reason,
	//	Password:     reqs.Password,
	//}); err != nil {
	//	return err
	//}
	//
	return nil
}
