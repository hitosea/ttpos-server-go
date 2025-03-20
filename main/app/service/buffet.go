package service

import (
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/repository/base"
	"ttpos-server-go/pkg/database"
)

// IBuffetSrv 定义收银服务接口
type IBuffetSrv interface {
	GetBuffetList(dbId uint64) (resp.BuffetListPaginationResp, error) // 获取自助餐列表
	GetBuffetDelayList(dbId uint64) (resp.BuffetDelayListResp, error) // 获取自助餐延迟列表
}

// buffetSrv 收银服务结构体
type buffetSrv struct {
	dbm *database.DBManager // 数据库管理器
}

// NewBuffetSrv 创建新的收银产品类别服务
func NewBuffetSrv(dbm *database.DBManager) IBuffetSrv {
	return NewBuffetSrvImpl(dbm)
}

// NewBuffetSrvImpl 创建新的收银服务实现
func NewBuffetSrvImpl(dbm *database.DBManager) IBuffetSrv {
	return &buffetSrv{dbm: dbm}
}

// GetBuffetList 获取列表
func (s *buffetSrv) GetBuffetList(dbId uint64) (resp.BuffetListPaginationResp, error) {
	// 获取列表
	buffets, total, err := repository.NewBuffetRepo(s.dbm.GetDB(dbId)).GetBuffetListInDeskOpen(
		1,
		1000,
	)
	if err != nil {
		return resp.BuffetListPaginationResp{}, errors.WithMessage(err)
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
		respBuffet := resp.Buffet{
			Uuid:                buffet.Uuid,
			Price:               buffet.GetMinPrice(),
			IsLimitTime:         buffet.IsLimitTime == 1,
			CanCombined:         buffet.CanCombined == 1,
			NonOrderingTime:     buffet.NonOrderingTime * 60,   // 分转为秒
			ReminderOrderTime:   buffet.ReminderOrderTime * 60, // 分转为秒
			LocaleName:          buffet.MultiLanguageName.GetNames(),
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

// GetBuffetDelayList 获取延迟列表
func (s *buffetSrv) GetBuffetDelayList(dbId uint64) (resp.BuffetDelayListResp, error) {
	// 获取列表
	delayList, err := base.NewBuffetDelayRepo(s.dbm.GetDB(dbId)).GetBuffetDelayList()
	if err != nil {
		return resp.BuffetDelayListResp{}, errors.WithMessage(err)
	}

	// 转换为响应对象
	respBuffets := make([]resp.BuffetDelay, 0, len(delayList))
	for _, delay := range delayList {
		respBuffet := resp.BuffetDelay{
			Uuid:  delay.Uuid,
			Name:  delay.Name,
			Price: delay.Price,
		}
		respBuffets = append(respBuffets, respBuffet)
	}

	// 返回响应对象
	return resp.BuffetDelayListResp{
		List: respBuffets,
	}, nil
}
