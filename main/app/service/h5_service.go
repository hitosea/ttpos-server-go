package service

import (
	"fmt"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/dto/resp/setting"
	"ttpos-server-go/app/errors"
	srvSetting "ttpos-server-go/app/service/setting"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/lock"
	"ttpos-server-go/pkg/utils"

	"github.com/jinzhu/copier"
)

// IH5Srv 定义H5服务接口
type IH5Srv interface {
	GetBaseInfo(ctx context.Context, deskUuid uint64) (*resp.H5BaseInfo, error)
	GetBuffetList(ctx context.Context, deskUuid uint64) (resp.H5BuffetList, error)
	OpenH5Desk(ctx context.Context, deskUuid uint64, request req.OpenDeskRequest) error
	RemarkProduct(ctx context.Context, remark string, saleOrderProductUuid uint64) error
}

// h5Srv 订单服务结构
type h5Srv struct {
	dbm        *database.DBManager // 数据库管理器
	settingSrv srvSetting.ISrv
	deskSrv    IDeskSrv
	orderSrv   IOrderSrv
	buffetSrv  IBuffetSrv
}

// NewH5Srv 创建H5服务实例
func NewH5Srv(dbm *database.DBManager, deskSrv IDeskSrv, orderSrv IOrderSrv, buffetSrv IBuffetSrv, settingSrv srvSetting.ISrv) IH5Srv {
	return NewH5SrvImpl(dbm, deskSrv, orderSrv, buffetSrv, settingSrv)
}

// NewH5SrvImpl 创建H5服务实例实现
func NewH5SrvImpl(dbm *database.DBManager, deskSrv IDeskSrv, orderSrv IOrderSrv, buffetSrv IBuffetSrv, settingSrv srvSetting.ISrv) IH5Srv {
	return &h5Srv{
		dbm:        dbm,
		settingSrv: settingSrv,
		deskSrv:    deskSrv,
		orderSrv:   orderSrv,
		buffetSrv:  buffetSrv,
	}
}

func (s *h5Srv) GetBaseInfo(ctx context.Context, deskUuid uint64) (*resp.H5BaseInfo, error) {
	dbId := ctx.GetDbId()

	deskInfo, err := s.deskSrv.GetDeskInfo(dbId, deskUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	company := ctx.GetCompany()
	companySetting := ctx.GetCompanySetting()

	h5Setting, err := s.settingSrv.GetH5Setting(ctx, nil)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	buffetSetting, buffetErr := s.settingSrv.GetBuffetSetting(ctx, companySetting)
	if buffetErr != nil {
		return nil, errors.WithMessage(buffetErr)
	}

	currencySetting, err := s.settingSrv.GetCurrencySetting(ctx)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	cloudBasicSetting, err := s.settingSrv.GetCloudBasicSetting(ctx)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	var kitchenSettingResp setting.KitchenResp
	kitchenSetting, err := s.settingSrv.GetKitchenSetting(ctx, companySetting, []dto.LanguageItem{})
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	copier.CopyWithOption(&kitchenSettingResp, kitchenSetting, copier.Option{IgnoreEmpty: true, DeepCopy: true})
	return &resp.H5BaseInfo{
		IsOpenH5Order: companySetting.IsOpenH5Order,
		Desk:          deskInfo,
		Company: resp.Company{ // h5
			Uuid:          company.Uuid,
			Name:          company.Name,
			Logo:          utils.AddImageDomain(company.Logo, utils.GetBaseURL(ctx.GetGin().Request), true),
			TimeZone:      companySetting.Timezone,
			ExpireTime:    company.ExpireTime,
			IsOpenMember:  companySetting.IsOpenMember,
			IsOpenBuffet:  companySetting.IsOpenBuffet,
			IsOpenH5Order: companySetting.IsOpenH5Order,
			IsOpenRider:   companySetting.IsOpenRider(),
			IsEnableErp:   company.IsOpenErp(),
		},
		H5:         h5Setting,
		Buffet:     buffetSetting,
		Currency:   currencySetting,
		CloudBasic: cloudBasicSetting,
		Business: setting.Business{
			ZeroingMethodList:         make([]setting.ZeroingMethodItem, 0),
			CheckoutZeroingMethodList: make([]setting.CheckoutZeroingMethodItem, 0),
			GiftMethodList:            make([]setting.GiftMethodItem, 0),
			FreeMethodList:            make([]setting.FreeMethodItem, 0),
		},
		Kitchen: kitchenSetting,
	}, nil
}

func (s *h5Srv) GetBuffetList(ctx context.Context, deskUuid uint64) (resp.H5BuffetList, error) {
	h5BuffetList := []resp.H5Buffet{}
	res, err := s.buffetSrv.GetBuffetList(deskUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	list := res.List
	for _, buffet := range list {
		buffetCustomerType := []resp.H5BuffetCustomerType{}
		for _, customerType := range buffet.BuffetCustomerTypes.List {
			buffetCustomerType = append(buffetCustomerType, resp.H5BuffetCustomerType{
				Id:             customerType.Uuid,
				NameText:       customerType.Name,
				CustomerTypeId: customerType.Uuid,
			})
		}
		buffetItem := resp.H5Buffet{
			NameText:       buffet.LocaleName.GetLocale(ctx.GetLanguage()),
			Id:             buffet.Uuid,
			Name:           buffet.LocaleName.GetLocale(ctx.GetLanguage()),
			Price:          fmt.Sprintf("%.2f", buffet.Price),
			BuyLimitStatus: 0,
			IsComb: func() int {
				if buffet.CanCombined {
					return 1
				}
				return 0
			}(),
			SaleNum:                  0,
			TimeLimit:                100,
			Status:                   1,
			Sort:                     0,
			RemainContinueTime:       buffet.ReminderOrderTime,
			RemainContinueNoticeTime: buffet.NonOrderingTime,
			BuffetCustomerType:       buffetCustomerType,
		}
		h5BuffetList = append(h5BuffetList, buffetItem)
	}

	return h5BuffetList, nil
}

func (s *h5Srv) OpenH5Desk(ctx context.Context, deskUuid uint64, request req.OpenDeskRequest) error {
	var param req.DeskOrderCreateReq
	if request.IsBuffet == 1 {
		param = req.DeskOrderCreateReq{
			DeskUuid:    deskUuid,
			BuffetUuids: request.BuffetIds,
			BuffetCustomerTypes: func() (list []req.DeskBuffetCustomerType) {
				for _, customerType := range request.BuffetCustomerTypeList {
					list = append(list, req.DeskBuffetCustomerType{
						Uuid:    customerType.CustomerTypeId,
						MealNum: customerType.Num,
					})
				}
				return
			}(),
		}
	} else {
		param = req.DeskOrderCreateReq{
			DeskUuid: deskUuid,
			MealNum:  request.MealNum,
		}
	}
	_, err := s.deskSrv.CreateDeskOrder(ctx, param)
	if err != nil {
		return errors.WithMessage(err)
	}

	return nil
}

func (s *h5Srv) RemarkProduct(ctx context.Context, remark string, saleOrderProductUuid uint64) error {
	// 获取桌台销售账单信息
	saleBill, errSaleBill := s.orderSrv.GetSaleBillByDeskId(ctx)
	if errSaleBill != nil {
		return errSaleBill
	}

	// 禁止并发操作
	if ctx.NoLock() {
		lock.NewSystemLock().LockUuid(saleBill.Uuid)
		defer lock.NewSystemLock().UnlockUuid(saleBill.Uuid)
		ctx.AddLock()
	}

	// 修改订单商品备注
	_, err := s.orderSrv.OrderProductRemark(ctx, req.OrderProductRemarkReq{
		SaleBillUuid:     saleBill.Uuid,
		SaleOrderUuid:    constant.OptionalUuid,
		OrderProductUuid: saleOrderProductUuid,
		Remark:           remark,
	})
	if err != nil {
		return errors.WithMessage(err)
	}
	// 备注成功。如果送厨后还能修改备注的话，需要在这里发起事件通知厨房

	return nil
}
