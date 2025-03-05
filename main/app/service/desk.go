package service

import (
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
	"ttpos-server-go/pkg/lock"

	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

// IDeskSrv 定义收银服务接口
type IDeskSrv interface {
	GetDeskList(ctx context.Context, dbId uint64, req req.DeskListReq) (resp.DeskListWithPaginationResp, error) // 获取桌台列表
	GetDeskRegionAndTypeList(dbId uint64) (resp.DeskRegionAndTypeListWithPaginationResp, error)                 // 获取桌台区域和类型列表
	GetDeskInfo(dbId uint64, deskUuid uint64) (resp.Desk, error)                                                // 获取桌台详情
	CreateDeskOrder(ctx context.Context, req req.DeskOrderCreateReq) (resp.CreateDeskOrderResp, error)          // 创建桌台订单
	CloseDesk(ctx context.Context, req req.DeskCloseReq) error                                                  // 关闭桌台
	CompleteDesk(ctx context.Context, req req.DeskJsonUuidReq) error                                            // 完成桌台
	ChangeDesk(ctx context.Context, req req.ChangeDeskReq) error                                                // 切换桌台
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
	if desk.SaleBill.ID == 0 {
		return model.Desk{}, nil
	}
	if _, err := NewOrderSrv(s.dbm, nil, nil, nil).IsCellCancelOrder(ctx, desk.SaleBill.Uuid); err != nil {
		return model.Desk{}, errors.WithMessage(err)
	}
	return desk, nil
}

// CloseDesk 关闭桌台
func (s *deskSrv) CloseDesk(ctx context.Context, reqs req.DeskCloseReq) error {
	dbId := ctx.GetDbId()
	db := s.dbm.GetDB(dbId)
	desk, err := repository.NewDeskRepo(db).GetDeskInfo(reqs.Uuid)
	if err != nil {
		return errors.WithMessage(err)
	}
	if desk.SaleBill.ID == 0 {
		return nil
	}
	// 取消订单
	if err := NewOrderSrv(s.dbm, s.localeSrv, s.settingSrv, nil).CancelOrder(ctx, req.OrderCancelReq{
		SaleBillUuid: desk.SaleBill.Uuid,
		CancelReason: reqs.Reason,
		Password:     reqs.Password,
	}); err != nil {
		return errors.WithMessage(err)
	}
	//
	return nil
}

// CompleteDesk 完成桌台
func (s *deskSrv) CompleteDesk(ctx context.Context, reqs req.DeskJsonUuidReq) error {
	dbId := ctx.GetDbId()
	db := s.dbm.GetDB(dbId)
	_, err := repository.NewDeskRepo(db).GetDeskInfo(reqs.Uuid)
	if err != nil {
		return errors.WithMessage(err)
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
func (s *deskSrv) ChangeDesk(ctx context.Context, reqs req.ChangeDeskReq) error {
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
		return errSaleBill
	}
	// 获取桌台信息
	desk, errDesk := repository.NewDeskRepo(db).GetDeskRecord(reqs.DeskUuid)
	if errDesk != nil {
		return errDesk
	}

	// 设置新桌台的开台信息
	desk.SetOpenDesk(reqs.SaleBillUuid)
	// 更新销售账单的桌台uuid,修改旧桌台的状态
	saleBill.ChangeDesk(reqs.DeskUuid)

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
		return errors.WithMessage(err)
	}
	return nil
}

// GetTabletDeskList 平板获取桌台列表
func (s *deskSrv) GetTabletDeskList(ctx context.Context) (resp.TabletDeskList, error) {
	deskRepo := repository.NewDeskRepo(s.dbm.GetDB(ctx.GetCompanyUuid()))
	desks, err := deskRepo.GetDesks(deskRepo.WhereIsNotDisable(), deskRepo.WhereUnBind())

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
		return errors.WithMessage(err)
	}

	if bindDeskReq.DeskUuid != bindDeskReq.OldDeskUuid {
		db := s.dbm.GetDB(ctx.GetCompanyUuid())
		deskRepo := repository.NewDeskRepo(db)
		desk, err := deskRepo.GetDesk(deskRepo.WhereUuid(bindDeskReq.DeskUuid))
		if err != nil || desk.Uuid == 0 {
			return errors.New("桌台不存在")
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
			return errors.New("绑定桌台失败")
		}
	}
	return nil
}
