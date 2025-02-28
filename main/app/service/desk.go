package service

import (
	"errors"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/utils"

	"github.com/jinzhu/copier"
	"gorm.io/gorm"

	"go.uber.org/zap"
)

// IDeskSrv 定义收银服务接口
type IDeskSrv interface {
	GetDeskList(ctx context.Context, dbId uint64, req req.DeskListReq) (resp.DeskListWithPaginationResp, error) // 获取桌台列表
	GetDeskRegionAndTypeList(dbId uint64) (resp.DeskRegionAndTypeListWithPaginationResp, error)                 // 获取桌台区域和类型列表
	GetDeskInfo(dbId uint64, deskUuid uint64) (resp.DeskInfoResp, error)                                        // 获取桌台详情
	CreateDeskOrder(ctx context.Context, req req.DeskOrderCreateReq) (resp.CreateDeskOrderResp, error)          // 创建桌台订单
	CloseDesk(ctx context.Context, req req.DeskCloseReq) error                                                  // 关闭桌台
	IsCellCloseDesk(ctx context.Context, deskUuid uint64) (model.Desk, error)                                   // 判断桌台是否可以关闭
	GetTabletDeskList(ctx context.Context) (resp.TabletDeskList, error)                                         // 平板获取桌台列表
	BindDesk(ctx context.Context, bindDeskReq req.BindDeskReq) error                                            // 平板端绑定桌台
}

// deskSrv 收银服务结构体
type deskSrv struct {
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
		ctx.Log().Debug("查询桌台列表", zap.Any("desk uuid", desk.Uuid))
		var elapsedTime uint
		if desk.SaleBill != nil && desk.SaleBill.ID == 0 {
			if desk.Status == 1 {
				extra.OccupyWaitNum++
			} else {
				extra.AvailableNum++
			}
			elapsedTime = constant.DeskStatusAvailable
		} else {
			extra.OccupyWaitNum++
			//
			if desk.SaleBill != nil && desk.SaleBill.IsLock == 1 {
				extra.LockNum++
			}
			//
			if desk.SaleBill != nil && desk.SaleBill.IsBuffet == 1 {
				extra.OccupyBuffetNum++
			} else {
				extra.OccupyNotBuffetNum++
			}
			// 如果是自助餐，计算剩余时间; 非自助餐，显示已用时间
			passedTime := int64(0)
			if desk.SaleBill != nil {
				passedTime = time.Now().Unix() - desk.SaleBill.CreateTime
			}
			if desk.SaleBill != nil && desk.SaleBill.IsBuffet == 1 {
				if uint(passedTime) >= desk.SaleBill.BuffetDuration {
					elapsedTime = 0
				} else {
					elapsedTime = desk.SaleBill.BuffetDuration - uint(passedTime)
				}
			} else {
				elapsedTime = uint(passedTime)
			}
		}

		extra.TotalNum++
		// todo  desk.SaleBill.PaymentAmount 需要等后面业务缓存中取
		if desk.SaleBill != nil {
			deskRes := resp.Desk{
				Uuid:          desk.Uuid,
				DeskNo:        desk.DeskNo,
				TypeUuid:      desk.TypeUuid,
				RegionUuid:    desk.RegionUuid,
				CustomerCount: desk.SaleBill.MealNum,
				Status:        desk.Status,
				IsLock:        desk.SaleBill.IsLock == 1,
				IsBuffet:      desk.SaleBill.IsBuffet == 1,
				IsWait:        desk.SaleBill.ID == 0 && desk.Status == 1,
				Remark:        desk.SaleBill.Remark,
				Time:          elapsedTime,
				Price:         desk.SaleBill.PaymentAmount,
			}
			deskResp[i] = deskRes
			jsonString := utils.ToJsonString(deskRes)
			ctx.Log().Debug("查询桌台列表", zap.Any("desk jsonString", jsonString))
		} else {
			deskRes := resp.Desk{
				Uuid:          desk.Uuid,
				DeskNo:        desk.DeskNo,
				TypeUuid:      desk.TypeUuid,
				RegionUuid:    desk.RegionUuid,
				Status:        desk.Status,
				CustomerCount: 0,
				IsLock:        false,
				IsBuffet:      false,
				IsWait:        false,
				Remark:        "",
				Time:          elapsedTime,
				Price:         0,
			}
			deskResp[i] = deskRes
			jsonString := utils.ToJsonString(deskRes)
			ctx.Log().Debug("查询桌台列表", zap.Any("desk jsonString", jsonString))
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
	//转换为响应对象
	var deskStatus uint
	var elapsedTime uint
	var lock bool
	var buffet bool
	if desk.SaleBill == nil {
		if desk.Status == 1 {
			deskStatus = constant.DeskStatusWait
		} else {
			deskStatus = constant.DeskStatusAvailable
		}
		elapsedTime = constant.DeskStatusAvailable
	} else {
		buffet = desk.SaleBill.IsBuffet == 1
		lock = desk.SaleBill.IsLock == 1
		if lock {
			deskStatus = constant.DeskStatusLock
		} else if buffet {
			deskStatus = constant.DeskStatusBuffet
		} else {
			deskStatus = constant.DeskStatusNotBuffet
		}
		// 如果是自助餐，计算剩余时间; 非自助餐，显示已用时间
		passedTime := time.Now().Unix() - desk.SaleBill.CreateTime
		if buffet {
			if uint(passedTime) >= desk.SaleBill.BuffetDuration {
				elapsedTime = 0
			} else {
				elapsedTime = desk.SaleBill.BuffetDuration - uint(passedTime)
			}
		} else {
			elapsedTime = uint(passedTime)
		}
	}

	var saleBillUuid uint64
	if desk.SaleBill != nil {
		saleBillUuid = desk.SaleBill.Uuid
	}
	var remark string
	if desk.SaleBill != nil {
		remark = desk.SaleBill.Remark
	}
	return resp.DeskInfoResp{
		Uuid:         desk.Uuid,
		SaleBillUuid: saleBillUuid,
		DeskNo:       desk.DeskNo,
		TypeUuid:     desk.TypeUuid,
		RegionUuid:   desk.RegionUuid,
		Status:       deskStatus,
		IsLock:       lock,
		IsBuffet:     buffet,
		Remark:       remark,
		Time:         elapsedTime,
	}, nil
}

// CreateDeskOrder 创建桌台订单
func (s *deskSrv) CreateDeskOrder(ctx context.Context, req req.DeskOrderCreateReq) (resp.CreateDeskOrderResp, error) {
	dbId := ctx.GetDbId()
	// 验证请求参数
	err := req.ValidateCreateDeskOrderReq()
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

// IsCellCloseDesk 判断桌台是否可关闭
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

// CloseDesk 关闭桌台
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

// GetTabletDeskList 平板获取桌台列表
func (s *deskSrv) GetTabletDeskList(ctx context.Context) (resp.TabletDeskList, error) {
	deskRepo := repository.NewDeskRepo(s.dbm.GetDB(ctx.GetCompanyUuid()))
	desks, err := deskRepo.GetDesks(deskRepo.WhereIsNotDisable(), deskRepo.WhereIsBind())

	if err != nil {
		return resp.TabletDeskList{}, errors.New("获取桌台列表失败")
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
		return err
	}

	if bindDeskReq.DeskUuid != bindDeskReq.OldDeskUuid {
		db := s.dbm.GetDB(ctx.GetCompanyUuid())
		deskRepo := repository.NewDeskRepo(db)
		desk, err := deskRepo.GetDesk(deskRepo.WhereUuid(bindDeskReq.DeskUuid))
		if err != nil || desk.Uuid == 0 {
			return errors.New("桌台不存在")
		}
		if desk.Uuid > 0 && desk.DeviceUuid != deviceUuid {
			return errors.New("桌台已被占用")
		}
		err = db.Transaction(func(tx *gorm.DB) error {
			deskRepo = repository.NewDeskRepo(tx)
			// 解绑旧的桌台（解绑当前设备绑定了非选定的其他桌台）
			if err := deskRepo.UnbindDesk(bindDeskReq.DeskUuid, deviceUuid); err != nil {
				return err
			}
			// 绑定新的桌台
			if err := deskRepo.UpdateDeskByMap(desk.Uuid, map[string]any{"device_uuid": deviceUuid}); err != nil {
				return err
			}
			if bindDeskReq.OldDeskUuid != 0 { // 解绑旧桌台
				if err := deskRepo.UpdateDeskByMap(bindDeskReq.OldDeskUuid, map[string]any{"device_uuid": 0}); err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			return errors.New("绑定桌台失败")
		}
	}
	return nil
}
