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
	bindRecordRepo *repository.BindRecordRepository
	supplierRepo   *repository.SupplierRepository
	settingSrv     *SettingService
}

func NewBindRecordService(bindRecordRepo *repository.BindRecordRepository, supplierRepo *repository.SupplierRepository, settingSrv *SettingService) *BindRecordService {
	return &BindRecordService{
		bindRecordRepo: bindRecordRepo,
		supplierRepo:   supplierRepo,
		settingSrv:     settingSrv,
	}
}

func (s *BindRecordService) Add(addReq req.AddBindRecordReq) error {
	if slices.Contains([]string{constant.SOURCE_CASHIER, constant.SOURCE_TABLET, constant.SOURCE_KITCHEN, constant.SOURCE_ASSISTANT}, addReq.Source) ||
		addReq.ShopSupplierId == 0 || addReq.DeviceID == "" {
		return errors.New("来源设备错误")
	}

	platform := utils.GetPlatform(addReq.UserAgent)

	// 获取绑定
	existsBindRecord := s.bindRecordRepo.GetRecordBySourceAndKey(addReq.Source, addReq.DeviceID)
	if existsBindRecord.Id != 0 {
		printPortId := addReq.PrintPortId
		remark := addReq.Remark
		if printPortId == 0 {
			printPortId = existsBindRecord.PrintPortId
		}
		if remark == "" {
			remark = existsBindRecord.Remark
		}
		if addReq.FinallyLoginTime == 0 {
			addReq.FinallyLoginTime = existsBindRecord.FinallyLoginTime
		}
		err := s.bindRecordRepo.Update(existsBindRecord.Id, map[string]interface{}{
			"print_port_id":      printPortId,
			"remark":             remark,
			"brand":              addReq.Remark,
			"platform":           platform,
			"user_agent":         addReq.UserAgent,
			"device_ip":          addReq.DeviceIP,
			"app_id":             addReq.AppId,
			"shop_supplier_id":   addReq.ShopSupplierId,
			"finally_login_id":   addReq.FinallyLoginId,
			"finally_login_time": addReq.FinallyLoginTime,
		})
		if err != nil {
			return errors.New("更新绑定信息失败")
		}
		return nil
	}

	// 获取supplier
	supplier := s.supplierRepo.GetById(addReq.ShopSupplierId)
	type Source struct {
		Name  string
		Limit uint
	}
	sources := map[string]Source{
		constant.SOURCE_CASHIER:   {"收银机", uint(supplier.CashLimit)},
		constant.SOURCE_TABLET:    {"平板", uint(supplier.TabletLimit)},
		constant.SOURCE_KITCHEN:   {"厨显", uint(supplier.KitchenLimit)},
		constant.SOURCE_ASSISTANT: {"点餐助手", uint(supplier.AssistantLimit)},
	}
	for sourceKey, source := range sources {
		if sourceKey != addReq.Source {
			continue
		}
		count := s.bindRecordRepo.GetBindCount(sourceKey)
		if count >= source.Limit { // 超过绑定上线
			return apperrors.New(constant.CodeBindLimit, source.Name+"登录设备已达上限，请在其他设备上退出登录或联系销售代表")
		}
	}

	//// 绑定品牌，如果自带打印，默认更新收银打印配置
	//if slices.Contains(constant.BRANDS_PRINTS, addReq.Brand) {
	//	printerSettings := s.settingSrv.GetSupplierItem(constant.PRINTER, addReq.ShopSupplierId, 0)
	//}

	return nil
}

func (s *BindRecordService) Unbind(appId uint, source string, key string, shopUserId uint) error {
	return s.bindRecordRepo.Unbind(appId, source, key, shopUserId)
}
