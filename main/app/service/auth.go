package service

import (
	"errors"
	"slices"
	"time"
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant/jwt"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/i18n"
	"ttpos-server-go/pkg/context"

	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	apperrors "ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/auth"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/utils"

	"github.com/gin-gonic/gin"
)

type IAuthSrv interface {
	Login(loginReq req.LoginReq, cc *gin.Context) (string, error)                            // 登录
	Logout(cc *gin.Context) error                                                            // 退出登录
	CashierBase(cc *gin.Context) (resp.CashierBase, error)                                   // 收银端基本信息
	AssistantBase(cc *gin.Context) (resp.AssistantBase, error)                               // 点餐助手端基本信息
	Auth(authReq req.Authenticate) (model.Company, model.CompanySetting, model.Staff, error) // 鉴权
	BindCashier(cashierReq req.BindCashierReq, cc *gin.Context) (string, error)              // 点餐助手绑定收银机
	GetOnlineCashiers(companyUuid uint64) resp.OnlineCashierList                             // 获取在线收银机
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
	assistantRoutes          []string
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

		cashierOpenStatusActions: []string{ // 开启收银端才能访问的接口 ToDo 待完善
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
		assistantRoutes: []string{ // 已登录点餐助手，但未绑定收银机可以不判断收银机状态的接口 ToDo 待完善
			//            '/index/getOnlineCashierList',
			//            '/call/call/unprocessed',
			//            '/store/table/table',
			//            '/index/tablePing',
			//            '/index/getAllProductImg'
		},
	}
}

// Login 登录
func (s *AuthSrv) Login(loginReq req.LoginReq, cc *gin.Context) (string, error) {
	var token string
	// 验证验证码
	if !s.captchaSrv.Verify(cc.GetHeader("X-Sign"), loginReq.Code) {
		return token, errors.New("验证码错误")
	}
	var staff model.Staff
	if config.Server.DeployMode == "cloud" { // 云上版本
		companyStaffRepo := repository.NewCompanyStaffRepo(s.dbm.GetDB(constant.DefaultDB))
		companyStaff := companyStaffRepo.GetByUsername(loginReq.Username)
		if companyStaff.Uuid == 0 {
			return token, errors.New("账号或密码错误")
		}
		if companyStaff.CompanyUuid == 0 {
			return token, errors.New("未找到绑定的商家，请确认登录信息")
		}
		staffRepo := repository.NewStaffRepo(s.dbm.GetDB(companyStaff.CompanyUuid))
		staff = staffRepo.GetByUuid(companyStaff.Uuid, staffRepo.WithCompany())
	} else { // 离线版本
		staffRepo := repository.NewStaffRepo(s.dbm.GetDB(constant.DefaultDB))
		staff = staffRepo.GetByUsername(loginReq.Username, staffRepo.WithCompany())
	}
	if staff.Uuid == 0 || utils.EncryptPassword(loginReq.Password) != staff.Password {
		return token, errors.New("账号或密码错误")
	}
	// 检查员工状态
	if staff.DeleteTime != 0 {
		return token, errors.New("账号被删除，请联系管理员")
	}
	if staff.IsDisable == 1 {
		return token, errors.New("账号被禁用，请联系管理员")
	}
	// 商家状态
	if staff.Company == nil || staff.Company.Uuid == 0 || staff.Company.DeleteTime != 0 {
		return token, errors.New("未找到绑定的商家，请确认登录信息")
	}

	switch loginReq.Source {
	case constant.SourceCashier: // 收银端登录
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
		currentStaff := staffRepo.GetByDeviceId(loginReq.DeviceId)
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
		err = staffRepo.Update(staff.Uuid, updates)
		if err != nil {
			return token, errors.New("更新信息失败")
		}
	case constant.SourceAssistant: // 点餐助手登录
		companySetting := repository.NewCompanySettingRepo(s.dbm.GetDB(staff.CompanyUuid)).Get()
		if companySetting.IsOpenAssistant != 1 {
			return token, errors.New("当前尚未开启点餐助手功能，如有需要，请联系销售代表")
		}
	default:
		return token, errors.New("登录来源错误")
	}

	// 添加绑定记录
	err := s.bindRecordSrv.Add(req.AddBindRecordReq{
		DeviceId:         loginReq.DeviceId,
		Brand:            loginReq.Brand,
		Source:           loginReq.Source,
		FinallyLoginUuid: staff.Uuid,
		FinallyLoginTime: time.Now().Unix(),
		CompanyUuid:      staff.CompanyUuid,
	}, cc)
	if err != nil {
		return token, err
	}

	// 生成 JWT token
	token, err = auth.GenerateToken(loginReq.Source, loginReq.DeviceId, staff.CompanyUuid, staff.Uuid, config.JWT.Secret, config.JWT.Expire, auth.Assistant{})
	if err != nil {
		return token, errors.New("生成token失败")
	}
	return token, nil
}

// Logout 退出登录
func (s *AuthSrv) Logout(cc *gin.Context) error {
	companyUuid := helper.GetCompanyUuid(cc)
	source := helper.GetSource(cc)
	staffUuid := cc.GetUint64(jwt.StaffUuid)
	assistantUuid := cc.GetUint64(jwt.AssistantStaffUuid)

	staffRepo := repository.NewStaffRepo(s.dbm.GetDB(companyUuid))
	if source == constant.SourceAssistant && assistantUuid != 0 {
		staffUuid = assistantUuid
	}
	staff := staffRepo.GetByUuid(staffUuid)
	return s.bindRecordSrv.Unbind(companyUuid, source, staff.BindKey, staff.Uuid)
}

// CashierBase 获取收银端基本信息
func (s *AuthSrv) CashierBase(cc *gin.Context) (resp.CashierBase, error) {
	var cashierBase resp.CashierBase
	company := helper.GetCompany(cc)
	companySetting := helper.GetCompanySetting(cc)
	staff := helper.GetStaff(cc)
	var (
		source   = helper.GetSource(cc)
		deviceId = cc.GetString(jwt.DeviceId)
	)
	deviceRemark := s.bindRecordSrv.GetRemark(company.Uuid, source, deviceId)
	// 判断权限
	permissions, err := s.roleAccessSrv.GetPermission(constant.CashierRouteName, staff.Uuid, staff.CompanyUuid)
	if err != nil {
		return cashierBase, err
	}
	if len(permissions) == 0 {
		return cashierBase, errors.New("当前无权限，请联系管理员")
	}
	language := i18n.GetAcceptLanguage(cc)
	ctx := context.NewContext(context.WithGinContext(cc))
	cashierSetting, err := s.settingSrv.GetCashierSetting(ctx, company.Uuid, language, nil)
	if err != nil {
		return cashierBase, err
	}
	businessSetting, err := s.settingSrv.GetBusinessSetting(company.Uuid, language)
	if err != nil {
		return cashierBase, err
	}
	buffetSetting, err := s.settingSrv.GetBuffetSetting(company.Uuid, companySetting)
	if err != nil {
		return cashierBase, err
	}
	currencySetting, err := s.settingSrv.GetCurrencySetting(company.Uuid)
	if err != nil {
		return cashierBase, err
	}
	return resp.CashierBase{
		Username:     staff.Username,
		CashierUuid:  staff.Uuid,
		DeviceId:     deviceId,
		DeviceRemark: deviceRemark,
		Cashier:      cashierSetting,
		Business:     businessSetting,
		Buffet:       buffetSetting,
		Currency:     currencySetting,
		Permissions:  permissions,
		Company: resp.Company{
			Uuid: company.Uuid,
			Name: company.Name,
		},
	}, nil
}

// AssistantBase 获取收银端基本信息
func (s *AuthSrv) AssistantBase(cc *gin.Context) (resp.AssistantBase, error) {
	company := helper.GetCompany(cc)
	staff := helper.GetStaff(cc)
	var (
		source   = helper.GetSource(cc)
		deviceId = cc.GetString(jwt.DeviceId)
	)
	_ = s.bindRecordSrv.GetRemark(company.Uuid, source, deviceId)
	return resp.AssistantBase{
		Username:      staff.Username,
		AssistantUuid: staff.Uuid,
	}, nil
}

// Auth 鉴权
func (s *AuthSrv) Auth(auth req.Authenticate) (model.Company, model.CompanySetting, model.Staff, error) {
	var (
		company        model.Company
		companySetting model.CompanySetting
		staff          model.Staff
	)

	staffRepo := repository.NewStaffRepo(s.dbm.GetDB(auth.CompanyUuid))
	staff = staffRepo.GetByUuid(auth.StaffUuid, staffRepo.WithCompany(), staffRepo.WithCompanySetting())
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
	// 验证设备是否绑定
	deviceId := auth.DeviceId
	if auth.Source == constant.SourceAssistant && auth.Assistant.DeviceId != "" { // 登录了点餐助手，且绑定了收银机
		deviceId = auth.Assistant.DeviceId
	}
	if auth.Source != constant.SourceShop && !s.bindRecordSrv.IsDeviceBind(auth.CompanyUuid, auth.Source, deviceId) {
		return company, companySetting, staff, apperrors.NewWithCode(constant.CodeUnbindError, "设备已解绑，请重新绑定")
	}

	switch auth.Source {
	case constant.SourceCashier: // 收银端
		{
			// 检查收银是否开启
			if !s.isCashierOpen(auth.CompanyUuid, auth.UrlPath) {
				return company, companySetting, staff, apperrors.NewWithCode(constant.CodeTableError, "收银用餐已关闭，请选择其他用餐方式")
			}
			// 判断权限
			_, err := s.roleAccessSrv.GetApiPermission(staff.Uuid, auth.CompanyUuid)
			if err != nil {
				return company, companySetting, staff, errors.New("当前无权限，请联系管理员")
			}
			// ToDo 记得开放
			//if !slices.Contains(permissions, urlPath) {
			//	return company, companySetting, staff, errors.New("当前无权限，请联系管理员")
			//}
		}
	case constant.SourceAssistant: // 点餐助手端
		{
			if !slices.Contains(s.assistantRoutes, auth.UrlPath) { // 除了这些接口外，其他都需要判断收银机状态
				cashierBindRecord := repository.NewBindRecordRepo(s.dbm.GetDB(auth.CompanyUuid)).GetBySourceAndDeviceId(constant.SourceCashier, auth.DeviceId)
				if cashierBindRecord.Uuid == 0 {
					return company, companySetting, staff, apperrors.NewWithCode(constant.TokenError, "收银员设备已解绑，请重新绑定")
				}
				if cashierBindRecord.FinallyLoginUuid == 0 {
					return company, companySetting, staff, apperrors.NewWithCode(constant.TokenErrorNotLogin, "收银员登录信息错误，请重新登录")
				}
			}
			// 检查桌台功能是否开启
			if !s.isTableOpen(auth.CompanyUuid) {
				return company, companySetting, staff, apperrors.NewWithCode(constant.CodeTableError, "桌台用餐已关闭，请选择其他用餐方式")
			}
		}
	}

	return company, companySetting, staff, nil
}

// 检查收银是否开启
func (s *AuthSrv) isCashierOpen(companyUuid uint64, pathUrl string) bool {
	cashierSetting, err := s.settingSrv.GetCashierSetting(context.NewDefaultContext(), companyUuid, "", []dto.LanguageItem{})
	if err != nil {
		return false
	}
	if cashierSetting.OrderMethod.IsCashierOrder == "0" && slices.Contains(s.cashierOpenStatusActions, pathUrl) {
		return false
	}
	return true
}

// 检查桌台功能是否开启
func (s *AuthSrv) isTableOpen(companyUuid uint64) bool {
	cashierSetting, err := s.settingSrv.GetCashierSetting(context.NewDefaultContext(), companyUuid, "", []dto.LanguageItem{})
	if err != nil {
		return false
	}
	return cashierSetting.OrderMethod.IsTableOrder != "0"
}

// BindCashier 绑定收银机
func (s *AuthSrv) BindCashier(bindReq req.BindCashierReq, cc *gin.Context) (string, error) {
	var newToken string
	companyUuid := helper.GetCompanyUuid(cc)
	if helper.GetSource(cc) != constant.SourceAssistant {
		return newToken, errors.New("用户信息错误")
	}
	staffRepo := repository.NewStaffRepo(s.dbm.GetDB(companyUuid))
	staff := staffRepo.GetByUuidAndDeviceId(bindReq.CashierUuid, bindReq.DeviceId, staffRepo.WithCompany(), staffRepo.WithCompanySetting())
	// 检查传递的收银设备
	if staff.Uuid == 0 {
		return newToken, errors.New("用户不存在")
	}
	if staff.CashierOnline != 1 {
		return newToken, errors.New("收银设备不在线")
	}
	// 判断收银员状态
	if staff.DeleteTime != 0 {
		return newToken, errors.New("账号被删除，请联系管理员")
	}
	if staff.IsDisable == 1 {
		return newToken, errors.New("账号被禁用，请联系管理员")
	}
	if staff.Company == nil || staff.Company.Uuid == 0 || staff.Company.DeleteTime != 0 {
		return newToken, errors.New("未找到绑定的商家，请确认登录信息")
	}
	// 是否已开启点餐助手功能
	if staff.Company.CompanySetting == nil || staff.Company.CompanySetting.IsOpenAssistant != 1 {
		return newToken, errors.New("当前尚未开启点餐助手功能，如有需要，请联系销售代表")
	}
	// 生成新的 JWT token，将点餐助手设备ID和员工Uuid单独存放
	newToken, err := auth.GenerateToken(constant.SourceAssistant, bindReq.DeviceId, companyUuid, bindReq.CashierUuid, config.JWT.Secret, config.JWT.Expire, auth.Assistant{
		DeviceId:  cc.GetString(jwt.DeviceId),
		StaffUuid: cc.GetUint64(jwt.StaffUuid),
	})
	if err != nil {
		return newToken, errors.New("生成token失败")
	}
	return newToken, nil
}

// GetOnlineCashiers 获取在线收银机
func (s *AuthSrv) GetOnlineCashiers(companyUuid uint64) resp.OnlineCashierList {
	staffRepo := repository.NewStaffRepo(s.dbm.GetDB(companyUuid))
	staffs := staffRepo.GetOnlineCashiers(staffRepo.WithDevice(constant.SettingCashier))

	var cashiers []resp.OnlineCashier
	for _, staff := range staffs {
		var remark string
		if staff.Device != nil {
			remark = staff.Device.Remark
		}
		cashiers = append(cashiers, resp.OnlineCashier{
			CashierUuid: staff.Uuid,
			Username:    staff.Username,
			DeviceId:    staff.BindKey,
			Remark:      remark,
		})
	}
	return resp.OnlineCashierList{
		List: cashiers,
	}
}
