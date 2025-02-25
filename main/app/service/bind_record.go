package service

import (
	"errors"
	"slices"
	"ttpos-server-go/app/dto"
	setting2 "ttpos-server-go/app/dto/resp/setting"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/pkg/context"

	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	apperrors "ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/utils"
)

type IBindRecordSrv interface {
	AddBindRecord(ctx context.Context, addReq req.AddBindRecordReq) (uint64, error)    // 添加绑定记录
	Unbind(companyUuid uint64, source string, deviceId string, staffUuid uint64) error // 解绑
	GetRemark(companyUuid uint64, source string, deviceId string) string               // 获取设备绑定备注
	IsDeviceBind(companyUuid uint64, source string, deviceId string) bool              // 设备是否绑定
}

func NewBindRecordSrv(settingSrv setting.ISrv, dbm *database.DBManager) IBindRecordSrv {
	return NewBindRecordSrvImpl(settingSrv, dbm)
}

type bindRecordSrv struct {
	settingSrv setting.ISrv
	dbm        *database.DBManager
}

func NewBindRecordSrvImpl(settingSrv setting.ISrv, dbm *database.DBManager) IBindRecordSrv {
	return &bindRecordSrv{
		dbm:        dbm,
		settingSrv: settingSrv,
	}
}

func (s *bindRecordSrv) AddBindRecord(ctx context.Context, addReq req.AddBindRecordReq) (uint64, error) {
	if !slices.Contains([]string{constant.SourceCashier, constant.SourceTablet, constant.SourceKitchen, constant.SourceAssistant}, addReq.Source) ||
		addReq.CompanyUuid == 0 || addReq.DeviceId == "" {
		return 0, errors.New("来源设备错误")
	}

	platform := utils.GetPlatform(addReq.UserAgent)

	// 获取绑定
	bindRecordRepo := repository.NewBindRecordRepo(s.dbm.GetDB(addReq.CompanyUuid))
	existsBindRecord := bindRecordRepo.GetBySourceAndDeviceId(addReq.Source, addReq.DeviceId)
	if existsBindRecord.ID != 0 {
		printPortUuid := addReq.PrintPortUuid
		remark := addReq.Remark
		if printPortUuid == 0 {
			printPortUuid = existsBindRecord.PrintPortUuid
		}
		if remark == "" {
			remark = existsBindRecord.Remark
		}
		finallyLoginTime := addReq.FinallyLoginTime
		if finallyLoginTime == 0 {
			finallyLoginTime = existsBindRecord.FinallyLoginTime
		}
		// 更新绑定
		err := bindRecordRepo.Update(existsBindRecord.Uuid, map[string]interface{}{
			"print_port_uuid":    printPortUuid,
			"remark":             remark,
			"brand":              addReq.Brand,
			"platform":           platform,
			"user_agent":         addReq.UserAgent,
			"device_ip":          addReq.DeviceIP,
			"finally_login_uuid": addReq.FinallyLoginUuid,
			"finally_login_time": finallyLoginTime,
		})
		if err != nil {
			return 0, errors.New("更新绑定信息失败")
		}
		return existsBindRecord.Uuid, nil
	}

	// 判断设备绑定上限
	companySetting := repository.NewCompanySettingRepo(s.dbm.GetDB(addReq.CompanyUuid)).Get()
	type Source struct {
		Name  string
		Limit uint
	}
	sources := map[string]Source{
		constant.SourceCashier:   {"收银机", uint(companySetting.CashLimit)},
		constant.SourceTablet:    {"平板", uint(companySetting.TabletLimit)},
		constant.SourceKitchen:   {"厨显", uint(companySetting.KitchenLimit)},
		constant.SourceAssistant: {"点餐助手", uint(companySetting.AssistantLimit)},
	}
	for sourceName, source := range sources {
		if sourceName != addReq.Source {
			continue
		}
		bindCount := bindRecordRepo.GetBindCountBySource(sourceName)
		if bindCount >= source.Limit { // 超过绑定上限
			return 0, apperrors.New(source.Name + "登录设备已达上限，请在其他设备上退出登录或联系销售代表")
		}
	}

	// 绑定品牌，如果自带打印，默认更新收银打印配置
	if addReq.Source == constant.SourceCashier && slices.Contains(constant.BrandsPrints, addReq.Brand) {
		printerSetting, err := s.settingSrv.GetPrinterSetting(ctx, []dto.LanguageItem{})
		if err != nil {
			return 0, err
		}
		if printerSetting.CashierOpen == "1" {
			var added bool
			for _, item := range printerSetting.CashierPrinter {
				if item.Key == addReq.DeviceId {
					added = true
				}
			}
			if !added {
				printerSetting.CashierPrinter = append(printerSetting.CashierPrinter, setting2.CashierPrinterItem{
					Key:       addReq.DeviceId,
					PrinterId: addReq.DeviceId,
				})
			}
			// 设置默认打印机
			if err = s.settingSrv.UpdateSetting(ctx, constant.SettingPrinter, printerSetting); err != nil {
				return 0, errors.New("设置默认打印机失败")
			}
		}
	}

	device, err := bindRecordRepo.CreateBindRecord(model.Device{
		FinallyLoginUuid: addReq.FinallyLoginUuid,
		FinallyLoginTime: addReq.FinallyLoginTime,
		Source:           addReq.Source,
		DeviceId:         addReq.DeviceId,
		Address:          addReq.Address,
		Port:             int(addReq.Port),
		DeviceIp:         addReq.DeviceId,
		Remark:           addReq.Remark,
		Brand:            addReq.Brand,
		Platform:         platform,
		UserAgent:        addReq.UserAgent,
	})
	if err != nil {
		return 0, err
	}
	return device.Uuid, nil
}

func (s *bindRecordSrv) Unbind(companyUuid uint64, source string, deviceId string, staffUuid uint64) error {
	bindRecordRepo := repository.NewBindRecordRepo(s.dbm.GetDB(companyUuid))
	return bindRecordRepo.Unbind(source, deviceId, staffUuid)
}

func (s *bindRecordSrv) GetRemark(companyUuid uint64, source string, deviceId string) string {
	bindRecordRepo := repository.NewBindRecordRepo(s.dbm.GetDB(companyUuid))
	return bindRecordRepo.GetBySourceAndDeviceId(source, deviceId).Remark
}

func (s *bindRecordSrv) IsDeviceBind(companyUuid uint64, source string, deviceId string) bool {
	bindRecordRepo := repository.NewBindRecordRepo(s.dbm.GetDB(companyUuid))
	return bindRecordRepo.GetBySourceAndDeviceId(source, deviceId).Uuid > 0
}
