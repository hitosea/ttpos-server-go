package service

import (
	"errors"
	"go.uber.org/zap"
	"slices"
	setting2 "ttpos-server-go/app/dto/resp/setting"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/i18n"
	"ttpos-server-go/pkg/logger"

	"github.com/gin-gonic/gin"

	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	apperrors "ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/utils"
)

type IBindRecordSrv interface {
	Add(addReq req.AddBindRecordReq, cc *gin.Context) error                            // 添加绑定记录
	Unbind(companyUuid uint64, source string, deviceId string, staffUuid uint64) error // 解绑
	GetRemark(companyUuid uint64, source string, deviceId string) string               // 获取设备绑定备注
	IsDeviceBind(companyUuid uint64, source string, deviceId string) bool              // 设备是否绑定
}

func NewBindRecordSrv(settingSrv setting.ISrv, dbm *database.DBManager) IBindRecordSrv {
	return NewBindRecordSrvImpl(settingSrv, dbm)
}

type BindRecordSrv struct {
	settingSrv setting.ISrv
	dbm        *database.DBManager
}

func NewBindRecordSrvImpl(settingSrv setting.ISrv, dbm *database.DBManager) *BindRecordSrv {
	return &BindRecordSrv{
		dbm:        dbm,
		settingSrv: settingSrv,
	}
}

func (s *BindRecordSrv) Add(addReq req.AddBindRecordReq, cc *gin.Context) error {
	if !slices.Contains([]string{constant.SourceCashier, constant.SourceTablet, constant.SourceKitchen, constant.SourceAssistant}, addReq.Source) ||
		addReq.CompanyUuid == 0 || addReq.DeviceId == "" {
		return errors.New("来源设备错误")
	}

	platform := utils.GetPlatform(addReq.UserAgent)

	// 获取绑定
	bindRecordRepo := repository.NewBindRecordRepo(s.dbm.GetDB(addReq.CompanyUuid))
	existsBindRecord := bindRecordRepo.GetRecordBySourceAndDeviceId(addReq.Source, addReq.DeviceId)
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
			return errors.New("更新绑定信息失败")
		}
		return nil
	}

	// 获取 company setting
	companySettingRepo := repository.NewCompanySettingRepo(s.dbm.GetDB(addReq.CompanyUuid))
	companySetting := companySettingRepo.GetByCompanyIdFromCompanyDB()
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
	for sourceKey, source := range sources {
		if sourceKey != addReq.Source {
			continue
		}
		count := bindRecordRepo.GetBindCount(sourceKey)
		if count >= source.Limit { // 超过绑定上线
			return apperrors.New(source.Name + "登录设备已达上限，请在其他设备上退出登录或联系销售代表")
		}
	}

	if slices.Contains(constant.BrandsPrints, addReq.Brand) {
		printerSetting, err := s.settingSrv.GetPrinterSetting(addReq.CompanyUuid, i18n.GetAcceptLanguage(cc), cc)
		if err != nil {
			return err
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
			if err = s.settingSrv.Updates(addReq.CompanyUuid, constant.SettingPrinter, printerSetting); err != nil {
				return errors.New("设置默认打印机失败")
			}
		}
	}

	uuid, err := database.GetID()
	if err != nil {
		logger.Logger.Error("生成雪花ID失败", zap.Error(err))
		return apperrors.NewWithCode(constant.CodeFail, "系统内部错误")
	}
	return bindRecordRepo.Create(model.BindRecord{
		Uuid:             uuid,
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
}

func (s *BindRecordSrv) Unbind(companyUuid uint64, source string, deviceId string, staffUuid uint64) error {
	bindRecordRepo := repository.NewBindRecordRepo(s.dbm.GetDB(companyUuid))
	return bindRecordRepo.Unbind(source, deviceId, staffUuid)
}

func (s *BindRecordSrv) GetRemark(companyUuid uint64, source string, deviceId string) string {
	bindRecordRepo := repository.NewBindRecordRepo(s.dbm.GetDB(companyUuid))
	return bindRecordRepo.GetRemark(source, deviceId)
}

func (s *BindRecordSrv) IsDeviceBind(companyUuid uint64, source string, deviceId string) bool {
	bindRecordRepo := repository.NewBindRecordRepo(s.dbm.GetDB(companyUuid))
	return bindRecordRepo.GetBindRecordUuid(source, deviceId) > 0
}
