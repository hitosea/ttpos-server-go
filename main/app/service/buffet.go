package service

import (
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/database"
)

// IBuffetSrv 定义收银服务接口
type IBuffetSrv interface {
	GetBuffetList(dbId uint64) (resp.BuffetListPaginationResp, error) // 获取自助餐列表
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
func (s *buffetSrv) GetBuffetList(dbId uint64) (resp.BuffetListPaginationResp, error) {
	// 获取列表
	buffets, total, err := repository.NewBuffetRepo(s.dbm.GetDB(dbId)).GetBuffetList(
		1,
		1000,
		repository.NewCommonRepo().Preload(repository.WithPreload{
			Query: "MultiLanguageName",
		}),
		repository.NewCommonRepo().Preload(repository.WithPreload{
			Query: "BuffetCustomerTypePrices",
		}),
		repository.NewCommonRepo().Preload(repository.WithPreload{
			Query: "BuffetCustomerTypePrices.BuffetCustomerType",
		}),
	)
	if err != nil {
		return resp.BuffetListPaginationResp{}, err
	}

	// 转换为响应对象
	respBuffets := make([]resp.Buffet, 0, len(buffets))
	for _, buffet := range buffets {
		buffetCustomerTypes := make([]resp.BuffetCustomerType, 0, len(buffet.BuffetCustomerTypePrices))
		for _, buffetCustomerTypePrice := range buffet.BuffetCustomerTypePrices {
			buffetCustomerTypes = append(buffetCustomerTypes, resp.BuffetCustomerType{
				Uuid: buffetCustomerTypePrice.CustomerTypeUuid,
				Name: buffetCustomerTypePrice.BuffetCustomerType.Name,
			})
		}
		//
		price := float64(0)
		if len(buffet.BuffetCustomerTypePrices) > 0 {
			price = buffet.BuffetCustomerTypePrices[0].Price
		}
		//
		respBuffet := resp.Buffet{
			Uuid:                buffet.Uuid,
			Price:               price,
			IsLimitTime:         buffet.IsLimitTime == 1,
			CanCombined:         buffet.CanCombined == 1,
			NonOrderingTime:     buffet.NonOrderingTime,
			ReminderOrderTime:   buffet.ReminderOrderTime,
			LocaleName:          s.localeSrv.GetLocaleNames(buffet.MultiLanguageName),
			BuffetCustomerTypes: resp.BuffetCustomerTypeList{List: buffetCustomerTypes},
		}
		respBuffets = append(respBuffets, respBuffet)
	}

	// 返回响应对象
	return resp.BuffetListPaginationResp{
		List: respBuffets,
		Meta: dto.PageResponse{
			PageNo:   1,
			PageSize: 1000,
			Total:    total,
		},
	}, nil
}
