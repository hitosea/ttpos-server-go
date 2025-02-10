package service

import (
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/resp/cashier_resp"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/database"
)

// IBuffetSrv 定义收银服务接口
type IBuffetSrv interface {
	GetBuffetList(dbId uint64) (cashier_resp.BuffetListPaginationResp, error) // 获取桌台列表
}

// buffetSrv 收银服务结构体
type buffetSrv struct {
	dbm       *database.DBManager // 数据库管理器
	localeSrv ILocaleSrv          // 多语言名称服务
}

// NewProductSrv 创建新的收银产品类别服务
func NewBuffetSrv(dbm *database.DBManager, localeSrv ILocaleSrv) IBuffetSrv {
	return NewBuffetSrvImpl(dbm, localeSrv)
}

// NewBuffetSrvImpl 创建新的收银服务实现
func NewBuffetSrvImpl(dbm *database.DBManager, localeSrv ILocaleSrv) IBuffetSrv {
	return &buffetSrv{
		dbm:       dbm,
		localeSrv: localeSrv,
	}
}

// GetBuffetList 获取列表
func (s *buffetSrv) GetBuffetList(dbId uint64) (cashier_resp.BuffetListPaginationResp, error) {
	// 获取列表
	buffets, total, err := repository.NewBuffetRepo(s.dbm.GetDB(dbId)).GetBuffetList(1, 1000)
	if err != nil {
		return cashier_resp.BuffetListPaginationResp{}, err
	}

	// 转换为响应对象
	respBuffets := make([]cashier_resp.Buffet, 0, len(buffets))
	for _, buffet := range buffets {
		respBuffet := cashier_resp.Buffet{
			Uuid:              uint64(buffet.UUID),
			Price:             buffet.BuffetCustomerTypePrice.Price,
			IsLimitTime:       buffet.IsLimitTime,
			CanCombined:       buffet.CanCombined,
			NonOrderingTime:   buffet.NonOrderingTime,
			ReminderOrderTime: buffet.ReminderOrderTime,
			LocaleName:        s.localeSrv.GetLocaleNames(buffet.MultiLanguageName),
			BuffetCustomerType: cashier_resp.BuffetCustomerType{
				Uuid:       uint64(buffet.BuffetCustomerTypePrice.Uuid),
				Price:      buffet.BuffetCustomerTypePrice.Price,
				LocaleName: s.localeSrv.GetLocaleNames(buffet.BuffetCustomerTypePrice.BuffetCustomerType.MultiLanguageName),
			},
		}
		respBuffets = append(respBuffets, respBuffet)
	}

	// 返回响应对象
	return cashier_resp.BuffetListPaginationResp{
		List: respBuffets,
		Meta: dto.PageResponse{
			PageNo:   1,
			PageSize: 1000,
			Total:    total,
		},
	}, nil
}
