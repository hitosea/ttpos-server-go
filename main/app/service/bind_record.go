package service

import (
	"errors"
	"slices"
	setting2 "ttpos-server-go/app/dto/resp/setting"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/i18n"

	"github.com/gin-gonic/gin"

	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	apperrors "ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/utils"
)

type IBindRecordSrv interface {
	Add(addReq req.AddBindRecordReq, cc *gin.Context) error
	Unbind(companyId uint, source string, key string, staffId uint) error
	GetRemark(companyId uint, source string, deviceId string) string
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
		addReq.CompanyId == 0 || addReq.DeviceId == "" {
		return errors.New("来源设备错误")
	}

	platform := utils.GetPlatform(addReq.UserAgent)

	// 获取绑定
	bindRecordRepo := repository.NewBindRecordRepo(s.dbm.GetDB(addReq.CompanyId))
	existsBindRecord := bindRecordRepo.GetRecordBySourceAndDeviceId(addReq.CompanyId, addReq.Source, addReq.DeviceId)
	if existsBindRecord.ID != 0 {
		printPortId := addReq.PrintPortId
		remark := addReq.Remark
		if printPortId == 0 {
			printPortId = existsBindRecord.PrintPortId
		}
		if remark == "" {
			remark = existsBindRecord.Remark
		}
		finallyLoginTime := addReq.FinallyLoginTime
		if finallyLoginTime == 0 {
			finallyLoginTime = existsBindRecord.FinallyLoginTime
		}
		// 更新绑定
		err := bindRecordRepo.Update(addReq.CompanyId, existsBindRecord.ID, map[string]interface{}{
			"print_port_id":      printPortId,
			"remark":             remark,
			"brand":              addReq.Brand,
			"platform":           platform,
			"user_agent":         addReq.UserAgent,
			"device_ip":          addReq.DeviceIP,
			"finally_login_id":   addReq.FinallyLoginId,
			"finally_login_time": finallyLoginTime,
		})
		if err != nil {
			return errors.New("更新绑定信息失败")
		}
		return nil
	}

	// 获取 company setting
	companySettingRepo := repository.NewCompanySettingRepo(s.dbm.GetDB(addReq.CompanyId))
	companySetting := companySettingRepo.GetByCompanyIdFromCompanyDB(addReq.CompanyId)
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
		count := bindRecordRepo.GetBindCount(addReq.CompanyId, sourceKey)
		if count >= source.Limit { // 超过绑定上线
			return apperrors.New(source.Name + "登录设备已达上限，请在其他设备上退出登录或联系销售代表")
		}
	}

	if slices.Contains(constant.BrandsPrints, addReq.Brand) {
		printerSetting, err := s.settingSrv.GetPrinterSetting(addReq.CompanyId, i18n.GetAcceptLanguage(cc), cc)
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
			if err = s.settingSrv.Updates(addReq.CompanyId, constant.SettingPrinter, printerSetting); err != nil {
				return errors.New("设置默认打印机失败")
			}
		}
	}

	return bindRecordRepo.Create(addReq.CompanyId, model.BindRecord{
		FinallyLoginId:   int(addReq.FinallyLoginId),
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

func (s *BindRecordSrv) Unbind(companyId uint, source string, key string, staffId uint) error {
	bindRecordRepo := repository.NewBindRecordRepo(s.dbm.GetDB(companyId))
	return bindRecordRepo.Unbind(companyId, source, key, staffId)
}

func (s *BindRecordSrv) GetRemark(companyId uint, source string, deviceId string) string {
	bindRecordRepo := repository.NewBindRecordRepo(s.dbm.GetDB(companyId))
	return bindRecordRepo.GetRemark(companyId, source, deviceId)
}
