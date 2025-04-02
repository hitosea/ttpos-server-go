package service

import (
	"slices"
	"ttpos-server-go/app/dto"
	setting2 "ttpos-server-go/app/dto/resp/setting"
	"ttpos-server-go/app/errors"
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

type IDeviceSrv interface {
	AddDevice(ctx context.Context, addReq req.AddDeviceReq) (uint64, error) // 添加绑定记录
	GetRemark(companyUuid uint64, source string, deviceId string) string    // 获取设备绑定备注
	IsDeviceBind(companyUuid uint64, source string, deviceId string) bool   // 设备是否绑定
}

func NewDeviceSrv(settingSrv setting.ISrv, dbm *database.DBManager) IDeviceSrv {
	return NewDeviceSrvImpl(settingSrv, dbm)
}

type deviceSrv struct {
	settingSrv setting.ISrv
	dbm        *database.DBManager
}

func NewDeviceSrvImpl(settingSrv setting.ISrv, dbm *database.DBManager) IDeviceSrv {
	return &deviceSrv{
		dbm:        dbm,
		settingSrv: settingSrv,
	}
}

func (s *deviceSrv) AddDevice(ctx context.Context, addReq req.AddDeviceReq) (uint64, error) {
	if !slices.Contains([]string{constant.SourceCashier, constant.SourceAssistant, constant.SourceTablet, constant.SourceKitchen}, addReq.Source) ||
		addReq.CompanyUuid == 0 || addReq.DeviceId == "" {
		return 0, errors.New("来源设备错误")
	}

	// 记录 ua 和 平台
	userAgent := ctx.GetGin().GetHeader("User-Agent") + ";" + ctx.GetGin().GetHeader("platform") // 记录平台
	platform := utils.GetPlatform(userAgent)

	db := s.dbm.GetDB(addReq.CompanyUuid)
	// 获取绑定
	deviceRepo := repository.NewDeviceRepo(db)
	existsDevice, _ := deviceRepo.GetDevice(deviceRepo.WhereSource(addReq.Source), deviceRepo.WhereSn(addReq.DeviceId))
	if existsDevice.ID != 0 {
		productPrinterUuid := addReq.ProductPrinterUuid
		if productPrinterUuid == 0 {
			productPrinterUuid = existsDevice.ProductPrinterUuid
		}
		remark := addReq.Remark
		if remark == "" {
			remark = existsDevice.Remark
		}
		finallyLoginTime := addReq.FinallyLoginTime
		if finallyLoginTime == 0 {
			finallyLoginTime = existsDevice.FinallyLoginTime
		}
		// 更新绑定
		err := deviceRepo.UpdateDevice(existsDevice.Uuid, map[string]any{
			"product_printer_uuid": productPrinterUuid,
			"remark":               remark,
			"brand":                addReq.Brand,
			"platform":             platform,
			"user_agent":           userAgent,
			"finally_login_uuid":   addReq.FinallyLoginUuid,
			"finally_login_time":   finallyLoginTime,
		})
		if err != nil {
			return 0, errors.WithMessage(err, "更新绑定信息失败")
		}
		return existsDevice.Uuid, nil
	}

	// 判断设备绑定上限
	companySetting := repository.NewCompanySettingRepo(db).Get()
	type Source struct {
		Name      string
		Limit     uint
		ErrorCode int
	}
	sources := map[string]Source{
		constant.SourceCashier:   {"收银机", uint(companySetting.CashLimit), constant.CashierLoginLimit},
		constant.SourceAssistant: {"点餐助手", uint(companySetting.AssistantLimit), constant.AssistantLoginLimit},
		constant.SourceKitchen:   {"厨显", uint(companySetting.KitchenLimit), constant.KitchenLoginLimit},
		constant.SourceTablet:    {"平板", uint(companySetting.TabletLimit), constant.TabletLoginLimit},
	}
	for sourceName, source := range sources {
		if sourceName != addReq.Source {
			continue
		}
		bindCount := deviceRepo.GetBindCountBySource(sourceName)
		if bindCount >= source.Limit { // 超过绑定上限
			return 0, apperrors.NewWithCode(source.ErrorCode, source.Name+"登录设备已达上限，请在其他设备上退出登录或联系销售代表")
		}
	}

	// 绑定品牌，如果自带打印，默认更新收银打印配置
	if addReq.Source == constant.SourceCashier && slices.Contains(constant.BrandsPrints, addReq.Brand) {
		printerSetting, err := s.settingSrv.GetPrinterSetting(ctx, []dto.LanguageItem{})
		if err != nil {
			return 0, errors.WithMessage(err)
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
				return 0, errors.WithMessage(err, "设置默认打印机失败")
			}
		}
	}

	device, err := deviceRepo.CreateDevice(model.Device{
		FinallyLoginUuid: addReq.FinallyLoginUuid,
		FinallyLoginTime: addReq.FinallyLoginTime,
		Source:           addReq.Source,
		DeviceId:         addReq.DeviceId,
		Remark:           addReq.Remark,
		Brand:            addReq.Brand,
		Platform:         platform,
		UserAgent:        userAgent,
	})
	if err != nil {
		return 0, errors.WithMessage(err)
	}
	return device.Uuid, nil
}

func (s *deviceSrv) GetRemark(companyUuid uint64, source string, deviceId string) string {
	deviceRepo := repository.NewDeviceRepo(s.dbm.GetDB(companyUuid))
	device, _ := deviceRepo.GetDevice(deviceRepo.WhereSource(source), deviceRepo.WhereSn(deviceId))
	return device.Remark
}

func (s *deviceSrv) IsDeviceBind(companyUuid uint64, source string, deviceId string) bool {
	deviceRepo := repository.NewDeviceRepo(s.dbm.GetDB(companyUuid))
	device, _ := deviceRepo.GetDevice(deviceRepo.WhereSource(source), deviceRepo.WhereSn(deviceId))
	return device.Uuid > 0
}
