package service

import (
	"errors"
	"slices"
	"time"
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant/jwt"
	"ttpos-server-go/app/dto/resp/cashier_resp"
	setting2 "ttpos-server-go/app/dto/resp/setting"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/i18n"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	apperrors "ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/auth"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"
)

type IAuthSrv interface {
	Login(source string, loginReq req.CashierLoginRequest, captchaId, captchaCode string, cc *gin.Context) (string, error)
	Logout(cc *gin.Context) error
	Base(cc *gin.Context) (cashier_resp.Base, error)
	AuthenticateStaff(source, deviceId string, companyUuid, staffUuid uint64, url string) (model.Company, model.CompanySetting, model.Staff, error)
}

func NewAuthSrv(
	dbm *database.DBManager,
	captchaSrv ICaptchaSrv,
	roleAccessSrv IRoleAccessSrv,
	bindRecordSrv IBindRecordSrv,
	staffShiftSrv IStaffShiftSrv,
	settingSrv setting.ISrv,
) IAuthSrv {
	return NewAuthSrvImpl(dbm, captchaSrv, roleAccessSrv, bindRecordSrv, staffShiftSrv, settingSrv)
}

type AuthSrv struct {
	dbm           *database.DBManager
	captchaSrv    ICaptchaSrv
	roleAccessSrv IRoleAccessSrv
	bindRecordSrv IBindRecordSrv
	shiftSrv      IStaffShiftSrv
	settingSrv    setting.ISrv

	cashierOpenStatusActions []string
}

func NewAuthSrvImpl(
	dbm *database.DBManager,
	captchaSrv ICaptchaSrv,
	roleAccessSrv IRoleAccessSrv,
	bindRecordSrv IBindRecordSrv,
	staffShiftSrv IStaffShiftSrv,
	settingSrv setting.ISrv,
) *AuthSrv {
	return &AuthSrv{
		dbm:           dbm,
		captchaSrv:    captchaSrv,
		roleAccessSrv: roleAccessSrv,
		bindRecordSrv: bindRecordSrv,
		shiftSrv:      staffShiftSrv,
		settingSrv:    settingSrv,

		cashierOpenStatusActions: []string{ // ToDo 待完善
			//"/order/cart/add",
			//"/order/cart/delProduct",
			//"/order/cart/stay",
			//"/order/cart/pick",
			//"/order/cart/delStay",
			//"/order/cart/sendKitchen",
			//"/order/cart/moveProduct",
			//"/order/cart/changeMoney",
			//"/order/order/buy",
			//"/index/getAllProductImg",
		},
	}
}

// Login 登录
func (s *AuthSrv) Login(source string, loginReq req.CashierLoginRequest, captchaId, captchaCode string, cc *gin.Context) (string, error) {
	companyStaffRepo := repository.NewCompanyStaffRepo(s.dbm.GetDB(constant.DefaultDB))
	var token string
	// 验证验证码
	if !s.captchaSrv.Verify(captchaId, captchaCode) {
		return token, apperrors.New("验证码错误")
	}

	var staff model.Staff
	if config.Server.DeployMode == "cloud" { // 云上版本
		companyStaff := companyStaffRepo.GetByUsername(loginReq.Username, companyStaffRepo.WithCompany())
		if companyStaff.Uuid == 0 {
			return token, errors.New("账号不存在")
		}
		if companyStaff.CompanyUuid == 0 {
			return token, errors.New("未找到绑定的商家，请确认登录信息")
		}
		staffRepo := repository.NewStaffRepo(s.dbm.GetDB(companyStaff.CompanyUuid))
		staff = staffRepo.GetByUuid(companyStaff.Uuid, staffRepo.WithCompany())
	} else { // 离线版本
		staffRepo := repository.NewStaffRepo(s.dbm.GetDB(constant.DefaultDB))
		staff = staffRepo.OfflineGetByUsername(loginReq.Username, staffRepo.WithCompany())
	}

	if staff.Uuid == 0 {
		return token, errors.New("用户不存在")
	}

	if utils.EncryptPassword(loginReq.Password) != staff.Password {
		return token, errors.New("密码错误")
	}
	if staff.DeleteTime != 0 {
		return token, apperrors.NewWithReplace("账号 %s 被删除，请联系管理员", []string{staff.Username})
	}
	if staff.IsDisable == 1 {
		return token, errors.New("账号被禁用，请联系管理员")
	}

	// 判断权限
	permissions, err := s.roleAccessSrv.GetPermission(constant.CashierRouteName, staff.Uuid, staff.CompanyUuid)
	if err != nil {
		return token, err
	}
	if len(permissions) == 0 {
		return token, errors.New("当前无权限，请联系管理员")
	}

	// 检查是否有未交班的收银员
	staffRepo := repository.NewStaffRepo(s.dbm.GetDB(staff.CompanyUuid))
	currentStaff := staffRepo.GetCurrentCashier(loginReq.DeviceId)
	if currentStaff.Uuid != 0 && currentStaff.Uuid != staff.Uuid {
		return token, apperrors.NewWithReplace("当前收银机上有未交班的账号，请联系 %s 完成交班后再登录", []string{currentStaff.RealName})
	}

	// 是否已在其他收银机登录
	if staff.CashierOnline == 1 && loginReq.DeviceId != staff.BindKey {
		cashierName := staff.RealName
		if cashierName == "" {
			cashierName = staff.Username
		}
		return token, apperrors.NewWithReplace("收银员 %s 已在其他收银机登录未交班，请先完成交班操作", []string{cashierName})
	}

	// 添加绑定记录
	err = s.bindRecordSrv.Add(req.AddBindRecordReq{
		DeviceId:         loginReq.DeviceId,
		Brand:            loginReq.Brand,
		Source:           constant.SourceCashier,
		FinallyLoginUuid: staff.Uuid,
		FinallyLoginTime: int(time.Now().Unix()),
		CompanyUuid:      staff.CompanyUuid,
	}, cc)
	if err != nil {
		logger.Logger.Error("绑定失败", zap.Error(err))
		return token, errors.New("绑定失败")
	}
	// 更新员工信息
	updates := map[string]any{
		"cashier_online": 1,
		"bind_key":       loginReq.DeviceId,
	}
	// 创建当班日志
	if staff.CashierLoginTime == 0 || staff.CashierOnline == 0 {
		shiftLog, err := s.shiftSrv.CreateWorkingLog(staff)
		if err != nil {
			return "", apperrors.ErrInternal
		}
		updates["cashier_login_time"] = shiftLog.ShiftStartTime
		updates["duty_no"] = shiftLog.ShiftNo
	}
	staffRepo = repository.NewStaffRepo(s.dbm.GetDB(staff.CompanyUuid))
	err = staffRepo.Update(staff.Uuid, updates)
	if err != nil {
		return token, errors.New("更新信息失败")
	}

	// 生成 JWT token
	token, err = auth.GenerateToken(constant.SourceCashier, loginReq.DeviceId, staff.CompanyUuid, staff.Uuid, config.JWT.Secret, config.JWT.Expire)
	if err != nil {
		return token, errors.New("生成token失败")
	}
	return token, nil
}

// Logout 退出登录
func (s *AuthSrv) Logout(cc *gin.Context) error {
	companyUuid := cc.GetUint64(jwt.CompanyUuid)
	staffRepo := repository.NewStaffRepo(s.dbm.GetDB(companyUuid))
	staff := staffRepo.GetByUuid(cc.GetUint64(jwt.StaffUuid))
	return s.bindRecordSrv.Unbind(companyUuid, constant.SourceCashier, staff.BindKey, staff.Uuid)
}

// Base 获取基本信息
func (s *AuthSrv) Base(cc *gin.Context) (cashier_resp.Base, error) {
	company := helper.GetCompany(cc)
	staff := helper.GetStaff(cc)
	var (
		source   = cc.GetString(jwt.Source)
		deviceId = cc.GetString(jwt.DeviceId)
	)
	deviceRemark := s.bindRecordSrv.GetRemark(company.Uuid, source, deviceId)
	// 判断权限
	permissions, err := s.roleAccessSrv.GetPermission(constant.CashierRouteName, staff.Uuid, staff.CompanyUuid)
	if err != nil {
		return cashier_resp.Base{}, err
	}
	if len(permissions) == 0 {
		return cashier_resp.Base{}, errors.New("当前无权限，请联系管理员")
	}
	languageList, _ := s.settingSrv.GetStoreLanguageList(company.Uuid, i18n.GetAcceptLanguage(cc), cc)
	allSettings, _ := s.settingSrv.GetAll(company.Uuid, i18n.GetAcceptLanguage(cc), languageList, cc)
	return cashier_resp.Base{
		Username:     staff.Username,
		CashierUuid:  staff.Uuid,
		DeviceId:     deviceId,
		DeviceRemark: deviceRemark,
		Cashier:      allSettings[constant.SettingCashier].(setting2.Cashier),
		Business:     allSettings[constant.SettingBusiness].(setting2.Business),
		Buffet:       allSettings[constant.SettingBuffet].(setting2.Buffet),
		Currency:     allSettings[constant.SettingCurrency].(setting2.Currency),
		Permissions:  permissions,
		Company: cashier_resp.Company{
			Uuid: company.Uuid,
			Name: company.Name,
		},
	}, nil
}

// AuthenticateStaff 认证登录员工
func (s *AuthSrv) AuthenticateStaff(source, deviceId string, companyUuid, staffUuid uint64, urlPath string) (model.Company, model.CompanySetting, model.Staff, error) {
	var (
		company        model.Company
		companySetting model.CompanySetting
		staff          model.Staff
	)

	staffRepo := repository.NewStaffRepo(s.dbm.GetDB(companyUuid))
	staff = staffRepo.GetByUuid(staffUuid, staffRepo.WithCompany(), staffRepo.WithCompanySetting())
	if staff.Uuid == 0 {
		return company, companySetting, staff, errors.New("用户不存在")
	}

	if staff.DeleteTime != 0 {
		return company, companySetting, staff, errors.New("用户被删除")
	}
	if staff.Company != nil {
		company = *staff.Company
		if staff.Company.CompanySetting != nil {
			companySetting = *staff.Company.CompanySetting
		}
	}

	if company.Uuid == 0 || companySetting.ID == 0 {
		return company, companySetting, staff, errors.New("商家不存在")
	}
	if company.Status == 0 {
		return company, companySetting, staff, errors.New("商家被禁用")
	}

	// 判断权限
	sourceMap := map[string]constant.RouteName{
		constant.SourceCashier:   constant.CashierRouteName,
		constant.SourceAssistant: constant.AssistantRouteName,
	}
	if _, exists := sourceMap[source]; exists {
		// 判断权限
		_, err := s.roleAccessSrv.GetApiPermission(staff.Uuid, companyUuid)
		if err != nil {
			return company, companySetting, staff, errors.New("当前无权限，请联系管理员")
		}
		// ToDo 记得开放
		//if !slices.Contains(permissions, urlPath) {
		//	return company, companySetting, staff, errors.New("当前无权限，请联系管理员")
		//}
	}
	// 验证设备是否绑定
	if !s.bindRecordSrv.IsDeviceBind(companyUuid, source, deviceId) {
		return company, companySetting, staff, apperrors.NewWithCode(constant.CodeUnbindError, "设备已解绑，请重新绑定")
	}

	if source == constant.SourceCashier {
		// 检查收银是否开启
		cashierSetting, err := s.settingSrv.GetCashierSetting(companyUuid, "", nil)
		if err != nil {
			return company, companySetting, staff, err
		}
		if cashierSetting.OrderMethod.IsCashierOrder == "0" && slices.Contains(s.cashierOpenStatusActions, urlPath) {
			return company, companySetting, staff, apperrors.NewWithCode(constant.CodeTableError, "收银用餐已关闭，请选择其他用餐方式")
		}
	}

	return company, companySetting, staff, nil
}
