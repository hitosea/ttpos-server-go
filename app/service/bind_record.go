package service

import (
	"errors"
	"slices"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	apperrors "ttpos-server-go/app/errors"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/utils"
)

type BindRecordService struct {
	bindRecordRepo     *repository.BindRecordRepository
	companySettingRepo *repository.CompanySettingRepository
}

func NewBindRecordService(bindRecordRepo *repository.BindRecordRepository, supplierRepo *repository.CompanySettingRepository) *BindRecordService {
	return &BindRecordService{
		bindRecordRepo:     bindRecordRepo,
		companySettingRepo: supplierRepo,
	}
}

func (s *BindRecordService) Add(addReq req.AddBindRecordReq) error {
	if slices.Contains([]string{constant.SOURCE_CASHIER, constant.SOURCE_TABLET, constant.SOURCE_KITCHEN, constant.SOURCE_ASSISTANT}, addReq.Source) ||
		addReq.CompanyId == 0 || addReq.DeviceId == "" {
		return errors.New("来源设备错误")
	}

	platform := utils.GetPlatform(addReq.UserAgent)

	// 获取绑定
	existsBindRecord := s.bindRecordRepo.GetRecordBySourceAndKey(addReq.Source, addReq.DeviceId)
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
		err := s.bindRecordRepo.Update(addReq.CompanyId, existsBindRecord.ID, map[string]interface{}{
			"print_port_id":      printPortId,
			"remark":             remark,
			"brand":              addReq.Remark,
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
	companySetting := s.companySettingRepo.GetById(addReq.CompanyId)
	type Source struct {
		Name  string
		Limit uint
	}
	sources := map[string]Source{
		constant.SOURCE_CASHIER:   {"收银机", uint(companySetting.CashLimit)},
		constant.SOURCE_TABLET:    {"平板", uint(companySetting.TabletLimit)},
		constant.SOURCE_KITCHEN:   {"厨显", uint(companySetting.KitchenLimit)},
		constant.SOURCE_ASSISTANT: {"点餐助手", uint(companySetting.AssistantLimit)},
	}
	for sourceKey, source := range sources {
		if sourceKey != addReq.Source {
			continue
		}
		count := s.bindRecordRepo.GetBindCount(addReq.CompanyId, sourceKey)
		if count >= source.Limit { // 超过绑定上线
			return apperrors.New(source.Name + "登录设备已达上限，请在其他设备上退出登录或联系销售代表")
		}
	}

	// ToDo 绑定品牌，如果自带打印，默认更新收银打印配置
	if slices.Contains(constant.BRANDS_PRINTS, addReq.Brand) {

	}

	return nil
}

func (s *BindRecordService) Unbind(appId uint, source string, key string, shopUserId uint) error {
	return s.bindRecordRepo.Unbind(appId, source, key, shopUserId)
}
