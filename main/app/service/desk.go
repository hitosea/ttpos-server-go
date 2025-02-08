package service

import (
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req/cashier_req"
	"ttpos-server-go/app/dto/resp/cashier_resp"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/database"
)

// IDeskSrv 定义收银服务接口
type IDeskSrv interface {
	GetDeskList(dbId uint, req cashier_req.DeskListReq) (cashier_resp.DeskListWithPaginationResp, error) // 获取桌台列表
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
func (s *deskSrv) GetDeskList(dbId uint, req cashier_req.DeskListReq) (cashier_resp.DeskListWithPaginationResp, error) {
	// 获取列表
	desks, total, err := repository.NewDeskRepo(s.dbm.GetDB(dbId)).GetClientDeskList(req.PageNo, req.PageSize)
	if err != nil {
		return cashier_resp.DeskListWithPaginationResp{}, err
	}

	// doto wfs 未完成
	// 转换为响应对象
	deskResp := make([]cashier_resp.Desk, len(desks))
	for i, desk := range desks {
		deskResp[i] = cashier_resp.Desk{
			Uuid:          desk.Uuid,
			TableNo:       desk.TableNo,
			TypeUuid:      desk.TypeUuid,
			RegionUuid:    desk.RegionUuid,
			Status:        desk.Status,
			CustomerCount: 1,
			IsBuffet:      1,
			Time:          1,
			Price:         1,
		}
	}

	// 返回响应对象
	return cashier_resp.DeskListWithPaginationResp{
		Extra: cashier_resp.DeskExtra{
			AvailableNum:       12,
			LockNum:            1,
			OccupyBuffetNum:    1,
			OccupyNotBuffetNum: 1,
			OccupyWaitNum:      1,
			TotalNum:           1,
		},
		List: deskResp,
		Meta: dto.PageResponse{
			PageNo:   req.PageNo,
			PageSize: req.PageSize,
			Total:    total,
		},
	}, nil
}
