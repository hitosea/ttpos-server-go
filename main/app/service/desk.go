package service

import (
	"time"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req/cashier_req"
	"ttpos-server-go/app/dto/resp/cashier_resp"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/database"
)

// IDeskSrv 定义收银服务接口
type IDeskSrv interface {
	GetDeskList(dbId uint64, req cashier_req.DeskListReq) (cashier_resp.DeskListWithPaginationResp, error) // 获取桌台列表
	GetDeskRegionAndTypeList(dbId uint64) (cashier_resp.DeskRegionAndTypeListWithPaginationResp, error)    // 获取桌台区域和类型列表
	GetDeskInfo(dbId uint64, deskUuid uint64) (cashier_resp.DeskInfoResp, error)                           // 获取桌台详情
}

// deskSrv 收银服务结构体
type deskSrv struct {
	dbm       *database.DBManager // 数据库管理器
	localeSrv ILocaleSrv          // 多语言名称服务
}

// NewProductSrv 创建新的收银产品类别服务
func NewDeskSrv(dbm *database.DBManager, localeSrv ILocaleSrv) IDeskSrv {
	return NewDeskSrvImpl(dbm, localeSrv)
}

// NewDeskSrvImpl 创建新的收银服务实现
func NewDeskSrvImpl(dbm *database.DBManager, localeSrv ILocaleSrv) IDeskSrv {
	return &deskSrv{
		dbm:       dbm,
		localeSrv: localeSrv,
	}
}

// GetProductList 获取收银机点餐页面产品类别列表
func (s *deskSrv) GetDeskRegionAndTypeList(dbId uint64) (cashier_resp.DeskRegionAndTypeListWithPaginationResp, error) {
	// 获取列表
	regions, _ := repository.NewDeskRegionRepo(s.dbm.GetDB(dbId)).GetDeskRegionList()
	types, _ := repository.NewDeskTypeRepo(s.dbm.GetDB(dbId)).GetDeskTypeList()

	// 转换为响应对象
	deskRegionResp := make([]cashier_resp.DeskRegion, len(regions))
	for i, region := range regions {
		deskRegionResp[i] = cashier_resp.DeskRegion{
			Uuid: region.Uuid,
			Name: region.Name,
		}
	}

	deskTypeResp := make([]cashier_resp.DeskType, len(types))
	for i, type_ := range types {
		deskTypeResp[i] = cashier_resp.DeskType{
			Uuid: type_.Uuid,
			Name: type_.Name,
		}
	}

	// 返回响应对象
	return cashier_resp.DeskRegionAndTypeListWithPaginationResp{
		Region: struct {
			List []cashier_resp.DeskRegion `json:"list"`
		}{List: deskRegionResp},
		Type: struct {
			List []cashier_resp.DeskType `json:"list"`
		}{List: deskTypeResp},
	}, nil
}

// GetProductList 获取收银机点餐页面产品类别列表
func (s *deskSrv) GetDeskList(dbId uint64, req cashier_req.DeskListReq) (cashier_resp.DeskListWithPaginationResp, error) {
	// 获取列表
	desks, total, err := repository.NewDeskRepo(s.dbm.GetDB(dbId)).GetClientDeskList(req.PageNo, req.PageSize)
	if err != nil {
		return cashier_resp.DeskListWithPaginationResp{}, err
	}

	// 初始化额外信息
	var extra = cashier_resp.DeskExtra{
		AvailableNum:       0,
		LockNum:            0,
		OccupyBuffetNum:    0,
		OccupyNotBuffetNum: 0,
		OccupyWaitNum:      0,
		TotalNum:           0,
	}

	// 转换为响应对象
	deskResp := make([]cashier_resp.Desk, len(desks))
	for i, desk := range desks {
		// 桌台状态	0:空闲 1:非自助餐 2:自助餐 3:待清台 4:锁单
		var deskStatus uint
		var elapsedTime uint
		if desk.SaleBill.ID == 0 {
			if desk.Status == 1 {
				deskStatus = 3
				extra.OccupyWaitNum++
			} else {
				deskStatus = 0
				extra.AvailableNum++
			}
			elapsedTime = 0
		} else {
			extra.OccupyWaitNum++
			//
			if desk.SaleBill.IsLock == 1 {
				deskStatus = 4
				extra.LockNum++
			} else if desk.SaleBill.IsBuffet == 1 {
				deskStatus = 2
			} else {
				deskStatus = 1
			}
			//
			if desk.SaleBill.IsBuffet == 1 {
				extra.OccupyBuffetNum++
			} else {
				extra.OccupyNotBuffetNum++
			}
			// 如果是自助餐，计算剩余时间; 非自助餐，显示已用时间
			passedTime := uint(time.Now().Unix()) - desk.SaleBill.CreateTime
			if desk.SaleBill.IsBuffet == 1 {
				if passedTime >= desk.SaleBill.BuffetDuration {
					elapsedTime = 0
				} else {
					elapsedTime = desk.SaleBill.BuffetDuration - passedTime
				}
			} else {
				elapsedTime = passedTime
			}
		}
		//
		extra.TotalNum++
		//
		// todo  desk.SaleBill.PaymentAmount 需要等后面业务缓存中取
		deskResp[i] = cashier_resp.Desk{
			Uuid:          desk.Uuid,
			DeskNo:        desk.DeskNo,
			TypeUuid:      desk.TypeUuid,
			RegionUuid:    desk.RegionUuid,
			Status:        deskStatus,
			CustomerCount: desk.SaleBill.MealNum,
			IsLock:        desk.SaleBill.IsLock,
			IsBuffet:      desk.SaleBill.IsBuffet,
			Remark:        desk.SaleBill.Remark,
			Time:          elapsedTime,
			Price:         desk.SaleBill.PaymentAmount,
		}
	}

	// 返回响应对象
	return cashier_resp.DeskListWithPaginationResp{
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
func (s *deskSrv) GetDeskInfo(dbId uint64, deskUuid uint64) (cashier_resp.DeskInfoResp, error) {
	// 获取列表
	desk, err := repository.NewDeskRepo(s.dbm.GetDB(dbId)).GetDeskInfo(deskUuid)
	if err != nil {
		return cashier_resp.DeskInfoResp{}, err
	}
	// 转换为响应对象
	// 桌台状态	0:空闲 1:非自助餐 2:自助餐 3:待清台 4:锁单
	var deskStatus uint
	var elapsedTime uint
	if desk.SaleBill.ID == 0 {
		if desk.Status == 1 {
			deskStatus = 3
		} else {
			deskStatus = 0
		}
		elapsedTime = 0
	} else {
		//
		if desk.SaleBill.IsLock == 1 {
			deskStatus = 4
		} else if desk.SaleBill.IsBuffet == 1 {
			deskStatus = 2
		} else {
			deskStatus = 1
		}
		// 如果是自助餐，计算剩余时间; 非自助餐，显示已用时间
		passedTime := uint(time.Now().Unix()) - desk.SaleBill.CreateTime
		if desk.SaleBill.IsBuffet == 1 {
			if passedTime >= desk.SaleBill.BuffetDuration {
				elapsedTime = 0
			} else {
				elapsedTime = desk.SaleBill.BuffetDuration - passedTime
			}
		} else {
			elapsedTime = passedTime
		}
	}
	//
	return cashier_resp.DeskInfoResp{
		Uuid:         desk.Uuid,
		SaleBillUuid: desk.SaleBill.Uuid,
		DeskNo:       desk.DeskNo,
		TypeUuid:     desk.TypeUuid,
		RegionUuid:   desk.RegionUuid,
		Status:       deskStatus,
		IsLock:       desk.SaleBill.IsLock,
		IsBuffet:     desk.SaleBill.IsBuffet,
		Remark:       desk.SaleBill.Remark,
		Time:         elapsedTime,
	}, nil
}
