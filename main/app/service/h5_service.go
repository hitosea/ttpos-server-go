package service

import (
	"fmt"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"

	"github.com/jinzhu/copier"
)

// IH5Srv 定义H5服务接口
type IH5Srv interface {
	GetCompanyInfo(ctx context.Context, deskUuid uint64) (*resp.GetBaseInfoResponse, error)
	GetBuffetList(ctx context.Context, deskUuid uint64) (resp.H5BuffetList, error)
	OpenH5Desk(ctx context.Context, deskUuid uint64, request req.OpenDeskRequest) error
	GetCategoryList(ctx context.Context) (resp.H5CategoryList, error)
}

// h5Srv 订单服务结构
type h5Srv struct {
	dbm        *database.DBManager // 数据库管理器
	settingSrv setting.ISrv
	deskSrv    IDeskSrv
	orderSrv   IOrderSrv
	buffetSrv  IBuffetSrv
}

// NewH5Srv 创建H5服务实例
func NewH5Srv(dbm *database.DBManager, localeSrv ILocaleSrv, cache setting.ISrv) IH5Srv {
	return NewH5SrvImpl(dbm, localeSrv, cache)
}

// NewH5SrvImpl 创建H5服务实例实现
func NewH5SrvImpl(dbm *database.DBManager, localeSrv ILocaleSrv, settingSrv setting.ISrv) IH5Srv {
	return &h5Srv{
		dbm:        dbm,
		settingSrv: settingSrv,
	}
}

func (s *h5Srv) GetCompanyInfo(ctx context.Context, deskUuid uint64) (*resp.GetBaseInfoResponse, error) {
	companyRepo := repository.NewCompanyRepo(s.dbm.GetDB(ctx.GetDbId()))
	companySetting, err := s.settingSrv.GetCompanySetting(ctx.GetCompanyUuid())
	if err != nil {
		return nil, err
	}
	currencySetting, err := s.settingSrv.GetCurrencySetting(ctx.GetCompanyUuid())
	if err != nil {
		return nil, err
	}
	h5Setting, err := s.settingSrv.GetH5Setting(ctx, ctx.GetCompanyUuid(), ctx.GetLanguage(), nil)
	if err != nil {
		return nil, err
	}
	buffetSetting, err := s.settingSrv.GetBuffetSetting(ctx.GetCompanyUuid(), companySetting)
	if err != nil {
		return nil, err
	}
	companyInfo, err := companyRepo.GetCompanyInfo(ctx.GetCompanyUuid())
	if err != nil {
		return nil, err
	}
	deskInfo, err := s.deskSrv.GetDeskInfo(ctx.GetDbId(), deskUuid)
	if err != nil {
		return nil, err
	}
	shop := resp.Shop{
		CompanyUuid:       companySetting.CompanyUuid,
		Name:              companyInfo.Name,
		RealName:          companySetting.RealName,
		LinkName:          companySetting.LinkName,
		LinkPhone:         companySetting.LinkPhone,
		Logo:              companyInfo.Logo,
		SaleStock:         companySetting.SaleStock,
		IsOpenMember:      companySetting.IsOpenMember,
		IsOpenTablet:      companySetting.IsOpenTablet,
		IsOpenH5:          companySetting.IsOpenH5,
		IsOpenAssistant:   companySetting.IsOpenAssistant,
		IsOpenKitchenKds:  companySetting.IsOpenKitchenKds,
		IsOpenBuffet:      companySetting.IsOpenBuffet,
		IsAcceptScanOrder: companySetting.IsOpenH5Order,
		IsOpenLocalPrint:  companySetting.IsOpenLocalPrint,
		CashLimit:         companySetting.CashLimit,
		KitchenLimit:      companySetting.KitchenLimit,
		TabletLimit:       companySetting.TabletLimit,
		AssistantLimit:    companySetting.AssistantLimit,
		TableLimit:        companySetting.TableLimit,
		PrinterLimit:      companySetting.PrinterLimit,
		Timezone:          companySetting.Timezone,
		Languages:         companySetting.Languages,
		Address:           companySetting.Address,
	}
	currency := resp.Currency{
		Unit:         currencySetting.Unit,
		IsOpen:       currencySetting.IsOpen,
		UnitPosition: currencySetting.UnitPosition,
		Vices: resp.Vices{
			ViceUnit:         currencySetting.ViceUnit,
			ViceUnitPosition: currencySetting.ViceUnitPosition,
			UnitRate:         currencySetting.UnitRate,
		},
	}
	languages := []resp.Language{}
	for _, language := range h5Setting.LanguageList {
		languages = append(languages, resp.Language{
			Key:   language.Key,
			Value: language.Value,
			I:     language.Name,
			Index: language.Name,
		})
	}
	h5 := resp.H5{
		IsCallService:      h5Setting.IsCallService,
		IsCustomerOrder:    h5Setting.IsCustomerOrder,
		IsVoiceRemind:      h5Setting.IsVoiceRemind,
		IsShowSoldOut:      fmt.Sprintf("%d", h5Setting.IsShowScanSoldOut),
		IsBuffetOrderLimit: h5Setting.IsBuffetOrderLimit,
		BuffetOrderLimit: struct {
			IsLimitTime string `json:"is_limit_time"`
			LimitTime   string `json:"limit_time"`
			IsLimitNum  string `json:"is_limit_num"`
			LimitNum    string `json:"limit_num"`
		}{
			IsLimitTime: h5Setting.BuffetOrderLimit.IsLimitTime,
			LimitTime:   h5Setting.BuffetOrderLimit.LimitTime,
			IsLimitNum:  h5Setting.BuffetOrderLimit.IsLimitNum,
			LimitNum:    h5Setting.BuffetOrderLimit.LimitNum,
		},
		IsOrderLimit: h5Setting.IsOrderLimit,
		OrderLimit: struct {
			IsLimitTime string `json:"is_limit_time"`
			LimitTime   string `json:"limit_time"`
			IsLimitNum  string `json:"is_limit_num"`
			LimitNum    string `json:"limit_num"`
		}{
			IsLimitTime: h5Setting.OrderLimit.IsLimitTime,
			LimitTime:   h5Setting.OrderLimit.LimitTime,
			IsLimitNum:  h5Setting.OrderLimit.IsLimitNum,
			LimitNum:    h5Setting.OrderLimit.LimitNum,
		},
		Language:          h5Setting.Language,
		DefaultLanguage:   h5Setting.DefaultLanguage,
		IsShowScanSoldOut: h5Setting.IsShowScanSoldOut,
		LanguageList:      languages,
	}
	h5BuffetResponse := resp.H5BuffetResponse{
		IsOpen:                   buffetSetting.IsOpen,
		TabletEndTime:            buffetSetting.TabletEndTime,
		IsRemainContinue:         buffetSetting.IsRemainContinue,
		RemainContinueTime:       buffetSetting.RemainContinueTime,
		RemainContinueNoticeTime: buffetSetting.RemainContinueNoticeTime,
		IsBuyContinue:            buffetSetting.IsBuyContinue,
		IsAddClock:               buffetSetting.IsAddClock,
		IsBuffetDiscount:         buffetSetting.IsBuffetDiscount,
		AddClock:                 []string{},
	}

	desk := resp.DeskInfo{
		TableID:        deskInfo.Uuid,
		TableNo:        deskInfo.DeskNo,
		Sort:           0,
		AreaID:         deskInfo.RegionUuid,
		TypeID:         deskInfo.TypeUuid,
		Status:         deskInfo.Status,
		SwitchStatus:   deskInfo.Status,
		AreaName:       "",
		TypeName:       "",
		ShopSupplierID: companyInfo.Uuid,
		MinNum:         0,
		MaxNum:         0,
		AppID:          companyInfo.Uuid,
		BindInfo:       "",
		QrcodeValue:    "",
		IsBind:         0,
	}
	res := &resp.GetBaseInfoResponse{}
	res.BaseInfo.Name = companyInfo.Name
	res.BaseInfo.Logo = companyInfo.Logo
	res.BaseInfo.IsAcceptScanOrder = companySetting.IsOpenH5Order
	res.OrderInfo.IsBuffet = func() int {
		if deskInfo.IsBuffet {
			return 1
		}
		return 0
	}()
	err = copier.CopyWithOption(res.BaseInfo.Shop, shop, copier.Option{IgnoreEmpty: true})
	if err != nil {
		return nil, err
	}
	err = copier.CopyWithOption(res.BaseInfo.Currency, currency, copier.Option{IgnoreEmpty: true})
	if err != nil {
		return nil, err
	}
	err = copier.CopyWithOption(res.BaseInfo.H5, h5, copier.Option{IgnoreEmpty: true})
	if err != nil {
		return nil, err
	}
	err = copier.CopyWithOption(res.BaseInfo.Buffet, h5BuffetResponse, copier.Option{IgnoreEmpty: true})
	if err != nil {
		return nil, err
	}
	err = copier.CopyWithOption(res.TableInfo, desk, copier.Option{IgnoreEmpty: true})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *h5Srv) GetBuffetList(ctx context.Context, deskUuid uint64) (resp.H5BuffetList, error) {
	h5BuffetList := []resp.H5Buffet{}
	res, err := s.buffetSrv.GetBuffetList(deskUuid)
	if err != nil {
		return nil, err
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
			Price:          fmt.Sprintf("%d", buffet.Price),
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
	iTrue := true
	if request.IsBuffet == 1 {
		param = req.DeskOrderCreateReq{
			DeskUuid:    deskUuid,
			IsBuffet:    &iTrue,
			BuffetUuids: request.BuffetIds,
			BuffetCustomerTypes: func() (list []req.DeskBuffetCustomerType) {
				for _, customerType := range request.BuffetCustomerTypeList {
					list = append(list, req.DeskBuffetCustomerType{
						Uuid:    customerType.CustomerTypeId,
						MealNum: &customerType.Num,
					})
				}
				return
			}(),
		}
	} else {
		param = req.DeskOrderCreateReq{
			DeskUuid: deskUuid,
			MealNum:  &request.MealNum,
		}
	}
	_, err := s.deskSrv.CreateDeskOrder(ctx.GetDbId(), param)
	if err != nil {
		return err
	}

	return nil
}

func (s *h5Srv) GetCategoryList(ctx context.Context) (resp.H5CategoryList, error) {
	return resp.H5CategoryList{}, nil
}
