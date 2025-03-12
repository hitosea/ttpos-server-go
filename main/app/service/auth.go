package service

import (
	"fmt"
	"slices"
	"time"
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant/jwt"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/service/setting"
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
)

type IAuthSrv interface {
	Login(ctx context.Context, loginReq req.LoginReq) (resp.LoginResp, error)                                     // 登录
	Logout(ctx context.Context) error                                                                             // 退出登录
	CashierBase(ctx context.Context) (resp.CashierBase, error)                                                    // 收银端基本信息
	AssistantBase(ctx context.Context) (resp.AssistantBase, error)                                                // 点餐助手端基本信息
	TabletBase(ctx context.Context) (resp.TabletBase, error)                                                      // 平板端基本信息
	KitchenBase(ctx context.Context) (resp.KitchenBase, error)                                                    // 厨显端基本信息
	Auth(ctx context.Context, authReq req.Authenticate) (model.Company, model.CompanySetting, model.Staff, error) // 鉴权
	AuthDesk(ctx context.Context, qrcodeToken string) error                                                       // 鉴权桌台
	BindCashier(ctx context.Context, cashierReq req.BindCashierReq) (string, error)                               // 点餐助手绑定收银机
	GetOnlineCashiers(companyUuid uint64) resp.OnlineCashierList                                                  // 获取在线收银机
}

func NewAuthSrv(
	dbm *database.DBManager,
	captchaSrv ICaptchaSrv,
	roleAccessSrv IRoleAccessSrv,
	deviceSrv IDeviceSrv,
	staffShiftSrv IStaffShiftSrv,
	settingSrv setting.ISrv,
) IAuthSrv {
	return NewAuthSrvImpl(dbm, captchaSrv, roleAccessSrv, deviceSrv, staffShiftSrv, settingSrv)
}

type authSrv struct {
	dbm           *database.DBManager
	captchaSrv    ICaptchaSrv
	roleAccessSrv IRoleAccessSrv
	deviceSrv     IDeviceSrv
	shiftSrv      IStaffShiftSrv
	settingSrv    setting.ISrv

	cashierOpenStatusActions []string
	assistantRoutes          []string
	tabletRoutes             []string
}

func NewAuthSrvImpl(
	dbm *database.DBManager,
	captchaSrv ICaptchaSrv,
	roleAccessSrv IRoleAccessSrv,
	deviceSrv IDeviceSrv,
	staffShiftSrv IStaffShiftSrv,
	settingSrv setting.ISrv,
) IAuthSrv {
	return &authSrv{
		dbm:           dbm,
		captchaSrv:    captchaSrv,
		roleAccessSrv: roleAccessSrv,
		deviceSrv:     deviceSrv,
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
			"/api/v1/assistant/online_cashiers",
			"/api/v1/assistant/bind_cashier",
			//            '/index/getOnlineCashierList',
			//            '/call/call/unprocessed',
			//            '/store/table/table',
			//            '/index/tablePing',
			//            '/index/getAllProductImg'
		},
		tabletRoutes: []string{
			"/api/v1/tablet/desk/list",
			"/api/v1/tablet/bind_desk",
			//'/passport/login',
			//'/passport/captcha',
			//'/passport/logout',
			//'/base/base/bind',
			//'/base/base/getNewVersion',
			//'/base/base/getAllProductImg',
			//'/base/base/getInfo',
			//'/table/table/openPing',
			//'/table/table/bind',
			//'/table/table/index',
			//'/table/table/getInfo',
			//'/table/table/unbind',
			//'/order/order/tableBuy',
			//'/base/base/verifyPassword',
			//'/call/call/call',
			//'/order/order/buffetList',
			//'/product/category/index', // 分类基础列表-缓存完完全外
			//'/product/product/getBaseList', // 商品基础列表-缓存完完全外
		},
	}
}

// Login 登录
func (s *authSrv) Login(ctx context.Context, loginReq req.LoginReq) (resp.LoginResp, error) {
	var loginResp resp.LoginResp
	var token string
	// 验证验证码
	if !s.captchaSrv.Verify(ctx.GetGin().GetHeader("X-Sign"), loginReq.Code) {
		return loginResp, errors.New("验证码错误")
	}
	var staff model.Staff
	if config.Server.DeployMode == "cloud" { // 云上版本
		companyStaffRepo := repository.NewCompanyStaffRepo(s.dbm.GetDB(constant.DefaultDB))
		companyStaff := companyStaffRepo.GetCompanyStaff(companyStaffRepo.WhereUsername(loginReq.Username))
		if companyStaff.Uuid == 0 {
			return loginResp, errors.New("账号或密码错误")
		}
		if companyStaff.CompanyUuid == 0 {
			return loginResp, errors.New("未找到绑定的商家，请确认登录信息")
		}
		staffRepo := repository.NewStaffRepo(s.dbm.GetDB(companyStaff.CompanyUuid))
		staff = staffRepo.GetStaff(staffRepo.WhereUuid(companyStaff.Uuid), staffRepo.WithCompany())
	} else { // 离线版本
		staffRepo := repository.NewStaffRepo(s.dbm.GetDB(constant.DefaultDB))
		staff = staffRepo.GetStaff(staffRepo.WhereUsername(loginReq.Username), staffRepo.WithCompany())
	}
	if staff.Uuid == 0 || utils.EncryptPassword(loginReq.Password) != staff.Password {
		return loginResp, errors.New("账号或密码错误")
	}
	// 检查员工状态
	if staff.DeleteTime != 0 {
		return loginResp, errors.New("账号被删除，请联系管理员")
	}
	if staff.IsDisable == 1 {
		return loginResp, errors.New("账号被禁用，请联系管理员")
	}
	// 商家状态
	if staff.Company == nil || staff.Company.Uuid == 0 || staff.Company.DeleteTime != 0 {
		return loginResp, errors.New("未找到绑定的商家，请确认登录信息")
	}

	isFirstLogin := staff.CashierOnline == 0

	switch loginReq.Source {
	case constant.SourceCashier: // 收银端登录
		// 判断权限
		permissions, err := s.roleAccessSrv.GetPermission(constant.CashierRouteName, staff.Uuid, staff.CompanyUuid)
		if err != nil {
			return loginResp, errors.WithMessage(err)
		}
		if len(permissions) == 0 {
			return loginResp, errors.New("当前无权限，请联系管理员")
		}
		// ToDo 记得开启
		//// 检查是否有未交班的收银员
		staffRepo := repository.NewStaffRepo(s.dbm.GetDB(staff.CompanyUuid))
		//currentStaff := staffRepo.GetStaff(staffRepo.WhereSn(loginReq.DeviceId), staffRepo.WhereCashierOnline())
		//if currentStaff.Uuid != 0 && currentStaff.Uuid != staff.Uuid {
		//	return loginResp, apperrors.NewWithReplace("当前收银机上有未交班的账号，请联系 %s 完成交班后再登录", []string{currentStaff.RealName})
		//}
		//// 是否已在其他收银机登录
		//if staff.CashierOnline == 1 && loginReq.DeviceId != staff.BindKey {
		//	cashierName := staff.RealName
		//	if cashierName == "" {
		//		cashierName = staff.Username
		//	}
		//	return loginResp, apperrors.NewWithReplace("收银员 %s 已在其他收银机登录未交班，请先完成交班操作", []string{cashierName})
		//}

		// 更新员工信息
		updates := map[string]any{
			"cashier_online": 1,
			"bind_key":       loginReq.DeviceId,
		}
		// 创建当班日志
		if staff.CashierLoginTime == 0 || staff.CashierOnline == 0 {
			shiftLog, err := s.shiftSrv.CreateWorkingLog(staff)
			if err != nil {
				return loginResp, apperrors.ErrInternal
			}
			updates["cashier_login_time"] = shiftLog.ShiftStartTime
			updates["duty_no"] = shiftLog.ShiftNo
		}
		err = staffRepo.Update(staff.Uuid, updates)
		if err != nil {
			return loginResp, errors.WithMessage(err, "更新信息失败")
		}
	case constant.SourceAssistant: // 点餐助手登录
		companySetting := repository.NewCompanySettingRepo(s.dbm.GetDB(staff.CompanyUuid)).Get()
		if companySetting.IsOpenAssistant != 1 {
			return loginResp, errors.New("当前尚未开启点餐助手功能，如有需要，请联系销售代表")
		}
	case constant.SourceKitchen: // 厨显端
		companySetting := repository.NewCompanySettingRepo(s.dbm.GetDB(staff.CompanyUuid)).Get()
		kitchenSetting, err := s.settingSrv.GetKitchenSetting(ctx, companySetting, []dto.LanguageItem{})
		if err != nil {
			return loginResp, errors.WithMessage(err)
		}
		if kitchenSetting.IsOpen == "0" || companySetting.IsOpenKitchenKds == 0 {
			return loginResp, errors.New("当前尚未开启厨显功能，如有需要，请联系销售代表")
		}
	case constant.SourceTablet: // 平板端
		companySetting := repository.NewCompanySettingRepo(s.dbm.GetDB(staff.CompanyUuid)).Get()
		if companySetting.IsOpenTablet == 0 {
			return loginResp, errors.New("当前尚未开启平板点餐功能，如有需要，请联系销售代表")
		}
	default:
		return loginResp, errors.New("登录来源错误")
	}

	// 添加绑定记录
	deviceUuid, err := s.deviceSrv.AddDevice(ctx, req.AddDeviceReq{
		DeviceId:         loginReq.DeviceId,
		Brand:            loginReq.Brand,
		Source:           loginReq.Source,
		FinallyLoginUuid: staff.Uuid,
		FinallyLoginTime: time.Now().Unix(),
		CompanyUuid:      staff.CompanyUuid,
	})
	if err != nil {
		return loginResp, errors.WithMessage(err)
	}

	// 生成 JWT token
	token, err = auth.GenerateToken(loginReq.Source, loginReq.DeviceId, staff.CompanyUuid, staff.Uuid, deviceUuid, config.JWT.Secret, config.JWT.Expire, auth.Assistant{})
	if err != nil {
		return loginResp, errors.New("生成token失败")
	}
	fmt.Println("login token:", token)
	return resp.LoginResp{
		Token:               token,
		CashierIsFirstLogin: isFirstLogin,
	}, nil
}

// Logout 退出登录
func (s *authSrv) Logout(ctx context.Context) error {
	db := s.dbm.GetDB(ctx.GetCompanyUuid())
	// 解绑设备
	err := repository.NewDeviceRepo(db).UpdateDevice(ctx.GetDeviceUuid(), map[string]any{
		"finally_login_uuid": 0,
	})
	if err != nil {
		return apperrors.ErrInternal
	}
	return nil
}

// CashierBase 获取收银端基本信息
func (s *authSrv) CashierBase(ctx context.Context) (resp.CashierBase, error) {
	var cashierBase resp.CashierBase
	company := ctx.GetCompany()
	companySetting := ctx.GetCompanySetting()
	staff := ctx.GetStaff()
	var (
		source   = ctx.GetSource()
		deviceId = ctx.GetGin().GetString(jwt.DeviceId)
	)
	deviceRemark := s.deviceSrv.GetRemark(company.Uuid, source, deviceId)
	// 判断权限
	permissions, err := s.roleAccessSrv.GetPermission(constant.CashierRouteName, staff.Uuid, staff.CompanyUuid)
	if err != nil {
		return cashierBase, errors.WithMessage(err)
	}
	if len(permissions) == 0 {
		return cashierBase, errors.New("当前无权限，请联系管理员")
	}
	cashierSetting, err := s.settingSrv.GetCashierSetting(ctx, nil)
	if err != nil {
		return cashierBase, errors.WithMessage(err)
	}
	businessSetting, err := s.settingSrv.GetBusinessSetting(ctx)
	if err != nil {
		return cashierBase, errors.WithMessage(err)
	}
	buffetSetting, err := s.settingSrv.GetBuffetSetting(ctx, companySetting)
	if err != nil {
		return cashierBase, errors.WithMessage(err)
	}
	currencySetting, err := s.settingSrv.GetCurrencySetting(ctx)
	if err != nil {
		return cashierBase, errors.WithMessage(err)
	}
	tabletSetting, err := s.settingSrv.GetTabletSetting(ctx, nil)
	if err != nil {
		return cashierBase, errors.WithMessage(err)
	}
	paymentSetting, err := s.settingSrv.GetPaymentSetting(ctx, companySetting)
	if err != nil {
		return cashierBase, errors.WithMessage(err)
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
			Uuid:     company.Uuid,
			Name:     company.Name,
			TimeZone: companySetting.Timezone,
		},
		Tablet:  tabletSetting,
		Payment: paymentSetting,
	}, nil
}

// AssistantBase 获取收银端基本信息
func (s *authSrv) AssistantBase(ctx context.Context) (resp.AssistantBase, error) {
	db := s.dbm.GetDB(ctx.GetCompanyUuid())
	assistantBase := resp.AssistantBase{}
	staff := helper.GetStaff(ctx.GetGin())
	staffRepo := repository.NewStaffRepo(db)
	assistantStaffUuid := ctx.GetGin().GetUint64(jwt.AssistantStaffUuid)
	assistantStaff := staffRepo.GetStaff(staffRepo.WhereUuid(assistantStaffUuid))
	perms, err := s.roleAccessSrv.GetPermission(constant.AssistantRouteName, assistantStaff.Uuid, assistantStaff.CompanyUuid)
	if err != nil {
		return assistantBase, errors.WithMessage(err)
	}
	permissions := make([]string, len(perms))
	for _, perm := range perms {
		permissions = append(permissions, perm.Path)
	}

	deviceRepo := repository.NewDeviceRepo(db)
	device, err := deviceRepo.GetDevice(deviceRepo.WhereSn(ctx.GetDeviceSn()))
	if err != nil {
		return assistantBase, errors.WithMessage(err)
	}
	company := helper.GetCompany(ctx.GetGin())
	companySetting := helper.GetCompanySetting(ctx.GetGin())
	businessSetting, err := s.settingSrv.GetBusinessSetting(ctx)
	if err != nil {
		return assistantBase, errors.WithMessage(err)
	}
	buffetSetting, err := s.settingSrv.GetBuffetSetting(ctx, companySetting)
	if err != nil {
		return assistantBase, errors.WithMessage(err)
	}
	currencySetting, err := s.settingSrv.GetCurrencySetting(ctx)
	if err != nil {
		return assistantBase, errors.WithMessage(err)
	}
	assistantSetting, err := s.settingSrv.GetAssistantSetting(ctx, nil)
	if err != nil {
		return assistantBase, errors.WithMessage(err)
	}
	paymentSetting, err := s.settingSrv.GetPaymentSetting(ctx, companySetting)
	if err != nil {
		return assistantBase, errors.WithMessage(err)
	}
	kitchenSetting, err := s.settingSrv.GetKitchenSetting(ctx, companySetting, nil)
	if err != nil {
		return assistantBase, errors.WithMessage(err)
	}
	return resp.AssistantBase{
		CashierStaff: resp.CashierStaff{
			RealName:     staff.RealName,
			Username:     staff.Username,
			DeviceId:     staff.BindKey,
			DeviceRemark: device.Remark,
		},
		AssistantStaff: resp.AssistantStaff{
			Uuid:       assistantStaff.Uuid,
			RealName:   assistantStaff.RealName,
			Phone:      assistantStaff.Phone,
			DeviceId:   assistantStaff.BindKey,
			Permission: permissions,
		},
		Company: resp.Company{
			Uuid:     company.Uuid,
			Name:     company.Name,
			TimeZone: companySetting.Timezone,
		},
		Assistant: assistantSetting,
		Buffet:    buffetSetting,
		Payment:   paymentSetting,
		Business:  businessSetting,
		Kitchen:   kitchenSetting,
		Currency:  currencySetting,
	}, nil
}

// TabletBase 平板端基本信息
func (s *authSrv) TabletBase(ctx context.Context) (resp.TabletBase, error) {
	company := helper.GetCompany(ctx.GetGin())
	companySetting := helper.GetCompanySetting(ctx.GetGin())

	var tabletBase resp.TabletBase
	cashierSetting, err := s.settingSrv.GetCashierSetting(ctx, nil)
	if err != nil {
		return tabletBase, errors.WithMessage(err)
	}
	buffetSetting, err := s.settingSrv.GetBuffetSetting(ctx, companySetting)
	if err != nil {
		return tabletBase, errors.WithMessage(err)
	}
	currencySetting, err := s.settingSrv.GetCurrencySetting(ctx)
	if err != nil {
		return tabletBase, errors.WithMessage(err)
	}
	tabletSetting, err := s.settingSrv.GetTabletSetting(ctx, nil)
	if err != nil {
		return tabletBase, errors.WithMessage(err)
	}
	kitchenSetting, err := s.settingSrv.GetKitchenSetting(ctx, companySetting, nil)
	if err != nil {
		return tabletBase, errors.WithMessage(err)
	}
	storeSetting, err := s.settingSrv.GetStoreSetting(ctx)
	if err != nil {
		return tabletBase, errors.WithMessage(err)
	}
	return resp.TabletBase{
		Company: resp.Company{
			Uuid:     company.Uuid,
			Name:     company.Name,
			TimeZone: companySetting.Timezone,
		},

		Cashier:  cashierSetting,
		Buffet:   buffetSetting,
		Currency: currencySetting,
		Tablet:   tabletSetting,
		Kitchen:  kitchenSetting,
		Store:    storeSetting,
	}, nil
}

// KitchenBase 获取收银端基本信息
func (s *authSrv) KitchenBase(ctx context.Context) (resp.KitchenBase, error) {
	var kitchenBase resp.KitchenBase
	company := helper.GetCompany(ctx.GetGin())
	companySetting := helper.GetCompanySetting(ctx.GetGin())
	kitchenSetting, err := s.settingSrv.GetKitchenSetting(ctx, companySetting, nil)
	if err != nil {
		return kitchenBase, errors.WithMessage(err)
	}
	return resp.KitchenBase{
		Kitchen: kitchenSetting,
		Company: resp.Company{
			Uuid:     company.Uuid,
			Name:     company.Name,
			TimeZone: companySetting.Timezone,
		},
	}, nil
}

// Auth 鉴权
func (s *authSrv) Auth(ctx context.Context, auth req.Authenticate) (model.Company, model.CompanySetting, model.Staff, error) {
	var (
		company        model.Company
		companySetting model.CompanySetting
		staff          model.Staff
		db             = s.dbm.GetDB(auth.CompanyUuid)
	)

	staffRepo := repository.NewStaffRepo(db)
	staff = staffRepo.GetStaff(staffRepo.WhereUuid(auth.StaffUuid), staffRepo.WithCompany(), staffRepo.WithCompanySetting())
	if staff.Uuid == 0 {
		return company, companySetting, staff, errors.New("用户不存在")
	}
	// 修改密码后，token失效
	if staff.PasswordChangeTime > auth.TokenIssuedAt {
		return company, companySetting, staff, errors.New("无效的token")
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
	if auth.Source != constant.SourceShop && !s.deviceSrv.IsDeviceBind(auth.CompanyUuid, auth.Source, deviceId) {
		return company, companySetting, staff, apperrors.NewWithCode(constant.CodeTokenInvalid, "设备已解绑，请重新绑定")
	}

	switch auth.Source {
	case constant.SourceCashier: // 收银端
		{
			// 检查收银是否开启
			if !s.isCashierOpen(ctx, auth.UrlPath) {
				return company, companySetting, staff, errors.New("收银用餐已关闭，请选择其他用餐方式")
			}
			// 判断权限
			_, err := s.roleAccessSrv.GetApiPermission(staff.Uuid, auth.CompanyUuid)
			if err != nil {
				return company, companySetting, staff, apperrors.NewWithCode(constant.CodeUnauthorized, "当前无权限，请联系管理员")
			}
			// ToDo 记得开放
			//if !slices.Contains(permissions, urlPath) {
			//	return company, companySetting, staff, apperrors.NewWithCode(constant.CodeUnauthorized, "当前无权限，请联系管理员")
			//}
		}
	case constant.SourceAssistant: // 点餐助手端
		{
			if !slices.Contains(s.assistantRoutes, auth.UrlPath) { // 除了这些接口外，其他都需要判断收银机状态
				deviceRepo := repository.NewDeviceRepo(db)
				cashierDevice, _ := deviceRepo.GetDevice(deviceRepo.WhereSource(constant.SourceCashier), deviceRepo.WhereSn(auth.DeviceId))
				if cashierDevice.Uuid == 0 {
					return company, companySetting, staff, errors.New("收银员设备已解绑，请重新绑定")
				}
				if cashierDevice.FinallyLoginUuid == 0 || auth.Assistant.StaffUuid == 0 {
					return company, companySetting, staff, apperrors.NewWithCode(constant.CodeCashierNotLogin, "收银员登录信息错误，请重新登录")
				}
			}
			// 检查桌台功能是否开启
			if !s.isTableOpen(ctx) {
				return company, companySetting, staff, errors.New("桌台用餐已关闭，请选择其他用餐方式")
			}
		}
	case constant.SourceTablet: // 平板端
		if !slices.Contains(s.tabletRoutes, auth.UrlPath) { // 除了这些接口外，其他都需要判断是否绑定了桌台
			deskRepo := repository.NewDeskRepo(db)
			_, err := deskRepo.GetDesk(deskRepo.WhereUuid(auth.DeskUuid), deskRepo.WhereDeviceUuid(auth.DeviceUuid))
			if err != nil {
				return company, companySetting, staff, errors.New("桌台未绑定")
			}
		}
		if companySetting.IsOpenTablet != 1 {
			return company, companySetting, staff, errors.New("当前未开启平板点餐功能，请联系销售代表")
		}
	}

	return company, companySetting, staff, nil
}

func (s *authSrv) AuthDesk(ctx context.Context, qrcodeToken string) error {
	companyUuid := ctx.GetCompanyUuid()
	deskUuid := ctx.GetDeskUuid()
	db := s.dbm.GetDB(companyUuid)
	company, err := repository.NewCompanyRepo(db).GetCompanyInfo(ctx)
	if err != nil {
		return errors.WithMessage(err)
	}
	if company.IsDelete() {
		return errors.New("商家已经删除")
	}

	deskInfo, err := repository.NewDeskRepo(db).GetDeskInfo(deskUuid)
	if err != nil {
		return errors.WithMessage(err)
	}
	if deskInfo.IsDelete() {
		return errors.New("桌台已经删除")
	}
	if deskInfo.QrcodeToken != qrcodeToken {
		return errors.New("二维码已失效，请联系商家")
	}
	return nil
}

// 检查收银是否开启
func (s *authSrv) isCashierOpen(ctx context.Context, pathUrl string) bool {
	cashierSetting, err := s.settingSrv.GetCashierSetting(ctx, []dto.LanguageItem{})
	if err != nil {
		return false
	}
	if cashierSetting.OrderMethod.IsCashierOrder == "0" && slices.Contains(s.cashierOpenStatusActions, pathUrl) {
		return false
	}
	return true
}

// 检查桌台功能是否开启
func (s *authSrv) isTableOpen(ctx context.Context) bool {
	cashierSetting, err := s.settingSrv.GetCashierSetting(ctx, []dto.LanguageItem{})
	if err != nil {
		return false
	}
	return cashierSetting.OrderMethod.IsTableOrder != "0"
}

// BindCashier 绑定收银机
func (s *authSrv) BindCashier(ctx context.Context, bindReq req.BindCashierReq) (string, error) {
	var newToken string
	companyUuid := helper.GetCompanyUuid(ctx.GetGin())
	if helper.GetSource(ctx.GetGin()) != constant.SourceAssistant {
		return newToken, errors.New("用户信息错误")
	}
	staffRepo := repository.NewStaffRepo(s.dbm.GetDB(companyUuid))
	staff := staffRepo.GetStaff(staffRepo.WhereUuid(bindReq.CashierUuid), staffRepo.WhereDeviceId(bindReq.DeviceId), staffRepo.WithCompany(), staffRepo.WithCompanySetting())
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
	newToken, err := auth.GenerateToken(constant.SourceAssistant, bindReq.DeviceId, companyUuid, bindReq.CashierUuid, ctx.GetDeviceUuid(), config.JWT.Secret, config.JWT.Expire, auth.Assistant{
		DeviceId:  ctx.GetGin().GetString(jwt.DeviceId),
		StaffUuid: ctx.GetGin().GetUint64(jwt.StaffUuid),
	})
	if err != nil {
		return newToken, errors.New("生成token失败")
	}
	return newToken, nil
}

// GetOnlineCashiers 获取在线收银机
func (s *authSrv) GetOnlineCashiers(companyUuid uint64) resp.OnlineCashierList {
	staffRepo := repository.NewStaffRepo(s.dbm.GetDB(companyUuid))
	staffs := staffRepo.GetStaffs(staffRepo.WhereCashierOnline(), staffRepo.WithDevice(constant.SettingCashier))

	cashiers := make([]resp.OnlineCashier, 0, len(staffs))
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
