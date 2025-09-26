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
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/utils"
)

type IDeviceSrv interface {
	AddDevice(ctx context.Context, addReq req.AddDeviceReq) (uint64, error)         // 添加绑定记录
	GetRemark(companyUuid uint64, source string, deviceId string) string            // 获取设备绑定备注
	IsDeviceBind(companyUuid uint64, source string, deviceId string) bool           // 设备是否绑定
	UpdateRemark(ctx context.Context, editSettingReq req.EditDeviceRemarkReq) error // 更新备注
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
	if !slices.Contains([]string{
		constant.SourceCashier,
		constant.SourceAssistant,
		constant.SourceTablet,
		constant.SourceKitchen,
		constant.SourceShop,
	}, addReq.Source) ||
		addReq.CompanyUuid == 0 || addReq.DeviceId == "" {
		return 0, errors.New("来源设备错误")
	}
	// 判断厨显模式
	if addReq.Source == constant.SourceKitchen && addReq.KdsMode != nil {
		if !slices.Contains([]uint{constant.KdsModeDefault, constant.KdsModeMake, constant.KdsModeMakeAndSend}, *addReq.KdsMode) {
			return 0, errors.New("厨显工作模式错误")
		}
	}
	// 记录 ua 和 平台
	userAgent := ctx.GetGin().GetHeader("User-Agent") + ";" + ctx.GetGin().GetHeader("platform") // 记录平台
	platform := utils.GetPlatform(userAgent)

	db := s.dbm.GetDB(addReq.CompanyUuid)
	deviceRepo := repository.NewDeviceRepo(db)
	// 判断设备绑定上限
	companySetting := repository.NewCompanySettingRepo(db).Get()
	if err := s.reachBindLimit(deviceRepo, companySetting, addReq.Source); err != nil {
		return 0, err
	}
	// 获取绑定
	existsDevice, _ := deviceRepo.GetDeviceAll(deviceRepo.WhereSource(addReq.Source), deviceRepo.WhereSn(addReq.DeviceId))
	if existsDevice.ID != 0 {
		productPrinterUuid := addReq.ProductPrinterUuid
		if productPrinterUuid == 0 {
			productPrinterUuid = existsDevice.ProductPrinterUuid
		}
		kdsMode := existsDevice.KdsMode
		if addReq.KdsMode != nil {
			kdsMode = *addReq.KdsMode
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
			"delete_time":          0,
			"product_printer_uuid": productPrinterUuid,
			"remark":               remark,
			"brand":                addReq.Brand,
			"platform":             platform,
			"user_agent":           userAgent,
			"finally_login_uuid":   addReq.FinallyLoginUuid,
			"finally_login_time":   finallyLoginTime,
			"kds_mode":             kdsMode,
			"is_main": func() int {
				if addReq.Source == constant.SourceCashier {
					if !deviceRepo.IsExistCashierMain(constant.SourceCashier) {
						return 1
					}
				}
				return existsDevice.IsMain
			}(),
		})
		if err != nil {
			return 0, errors.WithMessage(err, "更新绑定信息失败")
		}
		// 已软删除收银机重新登录，如果自带打印，默认更新收银打印配置
		if existsDevice.DeleteTime != 0 &&
			addReq.Source == constant.SourceCashier && slices.Contains(constant.BrandsPrints, addReq.Brand) {
			if err := s.bindPrinter(ctx, addReq.DeviceId); err != nil {
				return 0, errors.WithMessage(err, "设置默认打印机失败")
			}
		}
		return existsDevice.Uuid, nil
	}

	// 绑定品牌，如果自带打印，默认更新收银打印配置
	if addReq.Source == constant.SourceCashier && slices.Contains(constant.BrandsPrints, addReq.Brand) {
		if err := s.bindPrinter(ctx, addReq.DeviceId); err != nil {
			return 0, errors.WithMessage(err, "设置默认打印机失败")
		}
	}

	kdsMode := uint(0)
	if addReq.KdsMode != nil {
		kdsMode = *addReq.KdsMode
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
		KdsMode:          kdsMode,
		IsMain: func() int {
			if addReq.Source == constant.SourceCashier {
				if !deviceRepo.IsExistCashierMain(constant.SourceCashier) {
					return 1
				}
			}
			return 0
		}(),
	})
	if err != nil {
		return 0, errors.WithMessage(err)
	}
	return device.Uuid, nil
}

// 检查绑定上限
func (s *deviceSrv) reachBindLimit(deviceRepo repository.IDeviceRepo, companySetting model.CompanySetting, reqSource string) error {
	type Source struct {
		Name      string
		Limit     uint
		ErrorCode int
	}
	sources := map[string]Source{
		constant.SourceCashier:   {"收银机", uint(companySetting.CashLimit), constant.CodeCashierLoginLimit},         // "收银机登录设备已达上限，请在其他设备上退出登录或联系销售代表"
		constant.SourceAssistant: {"点餐助手", uint(companySetting.AssistantLimit), constant.CodeAssistantLoginLimit}, // "点餐助手登录设备已达上限，请在其他设备上退出登录或联系销售代表"
		constant.SourceKitchen:   {"厨显", uint(companySetting.KitchenLimit), constant.CodeKitchenLoginLimit},       // "厨显登录设备已达上限，请在其他设备上退出登录或联系销售代表"
		constant.SourceTablet:    {"平板", uint(companySetting.TabletLimit), constant.CodeTabletLoginLimit},         // "平板登录设备已达上限，请在其他设备上退出登录或联系销售代表"
	}
	for sourceName, source := range sources {
		if sourceName != reqSource {
			continue
		}
		if deviceRepo.GetBindCountBySource(sourceName) >= source.Limit {
			return errors.NewWithCode(source.ErrorCode, source.Name+"登录设备已达上限，请在其他设备上退出登录或联系销售代表")
		}
	}
	return nil
}

// 收银机绑定自带打印机
func (s *deviceSrv) bindPrinter(ctx context.Context, deviceId string) error {
	printerSetting, err := s.settingSrv.GetPrinterSetting(ctx, []dto.LanguageItem{})
	if err != nil {
		return errors.WithMessage(err)
	}
	if printerSetting.CashierOpen != "1" {
		return nil
	}
	var added bool
	for _, item := range printerSetting.CashierPrinter {
		if item.Key == deviceId {
			added = true
		}
	}
	if !added {
		printerSetting.CashierPrinter = append(printerSetting.CashierPrinter, setting2.CashierPrinterItem{
			Key:       deviceId,
			PrinterId: deviceId,
		})
	}
	// 设置默认打印机
	return s.settingSrv.UpdateSetting(ctx, constant.SettingPrinter, printerSetting)
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

func (s *deviceSrv) UpdateRemark(ctx context.Context, editSettingReq req.EditDeviceRemarkReq) error {
	return ctx.GetDB().Model(&model.Device{}).Where("uuid = ?", ctx.GetDeviceUuid()).Updates(map[string]any{
		"remark": editSettingReq.Remark,
	}).Error
}
