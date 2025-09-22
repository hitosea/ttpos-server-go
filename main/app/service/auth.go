package service

import (
	"regexp"
	"slices"
	"time"
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/constant/jwt"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/dto/resp/setting"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	settingSrv "ttpos-server-go/app/service/setting"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/auth"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"github.com/jinzhu/copier"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type IAuthSrv interface {
	Login(ctx context.Context, loginReq req.LoginReq) (resp.LoginResp, error)                                              // 登录
	Logout(ctx context.Context) error                                                                                      // 退出登录
	CashierBase(ctx context.Context) (resp.CashierBase, error)                                                             // 收银端基本信息
	AssistantBase(ctx context.Context) (resp.AssistantBase, error)                                                         // 点餐助手端基本信息
	TabletBase(ctx context.Context) (resp.TabletBase, error)                                                               // 平板端基本信息
	KitchenBase(ctx context.Context) (resp.KitchenBase, error)                                                             // 厨显端基本信息
	Auth(ctx context.Context, auth req.Authenticate) (model.Company, model.CompanySetting, model.Staff, model.Desk, error) // 鉴权
	AuthDesk(ctx context.Context, qrcodeToken string) (*model.Company, error)                                              // 鉴权桌台
	AuthMenu(ctx context.Context, qrcodeToken string) (*model.Company, error)                                              // 鉴权点子菜单
	BindCashier(ctx context.Context, bindReq req.BindCashierReq) (string, string, error)                                   // 点餐助手绑定收银机
	GetOnlineCashiers(companyUuid uint64) resp.OnlineCashierList                                                           // 获取在线收银机
	RefreshToken(ctx context.Context) (resp.LoginResp, error)                                                              // 刷新token
	ShopBase(ctx context.Context) (resp.ShopBase, error)                                                                   // 移动管理端基本信息
	ChangePassword(ctx context.Context, changePasswordReq req.ChangePasswordReq) error                                     // 移动管理端-修改密码
}

func NewAuthSrv(
	dbm *database.DBManager,
	captchaSrv ICaptchaSrv,
	roleAccessSrv IRoleAccessSrv,
	deviceSrv IDeviceSrv,
	staffShiftSrv IStaffShiftSrv,
	settingSrv settingSrv.ISrv,
) IAuthSrv {
	return NewAuthSrvImpl(dbm, captchaSrv, roleAccessSrv, deviceSrv, staffShiftSrv, settingSrv)
}

type authSrv struct {
	dbm           *database.DBManager
	captchaSrv    ICaptchaSrv
	roleAccessSrv IRoleAccessSrv
	deviceSrv     IDeviceSrv
	shiftSrv      IStaffShiftSrv
	settingSrv    settingSrv.ISrv

	assistantRoutes             []string
	tabletRoutes                []string
	memberFunctionRoutes        []string
	h5AcceptOrderFunctionRoutes []string
}

func NewAuthSrvImpl(
	dbm *database.DBManager,
	captchaSrv ICaptchaSrv,
	roleAccessSrv IRoleAccessSrv,
	deviceSrv IDeviceSrv,
	staffShiftSrv IStaffShiftSrv,
	settingSrv settingSrv.ISrv,
) IAuthSrv {
	return &authSrv{
		dbm:           dbm,
		captchaSrv:    captchaSrv,
		roleAccessSrv: roleAccessSrv,
		deviceSrv:     deviceSrv,
		shiftSrv:      staffShiftSrv,
		settingSrv:    settingSrv,

		assistantRoutes: []string{ // 已登录点餐助手，但未绑定收银机可以不判断收银机状态的接口
			"/api/v1/assistant/online_cashiers",
			"/api/v1/assistant/bind_cashier",
			"/api/v1/assistant/verify_lock_password",
		},
		tabletRoutes: []string{
			"/api/v1/tablet/desk/list",
			"/api/v1/tablet/desk/bind",
			"/api/v1/tablet/base",
			"/api/v1/tablet/logout",
			"/api/v1/tablet/check_update",
			"/api/v1/tablet/verify_advanced_password",
			"/api/v1/tablet/product/category/list",
			"/api/v1/tablet/product/list",
			"/api/v1/tablet/buffet/list",
		},
		memberFunctionRoutes: []string{
			"/api/v1/cashier/instant/order/member/confirm",  // 收银端使用会员优惠
			"/api/v1/assistant/member/add",                  // 助手端添加会员
			"/api/v1/cashier/member/add",                    // 收银端添加会员
			"/api/v1/cashier/member/create_recharge_order",  // 收银端创建会员充值单
			"/api/v1/cashier/member/confirm_recharge_order", // 收银端确认会员充值订单
			"/api/v1/assistant/desk/order/member/confirm",   // 助手端使用会员优惠
			"/api/v1/cashier/desk/order/member/confirm",     // 收银端使用会员优惠
			"/api/v1/cashier/recharge_order/list",           // 收银端订单列表
		},
		h5AcceptOrderFunctionRoutes: []string{
			"/api/v1/cashier/h5_order/list",   // h5订单列表
			"/api/v1/cashier/h5_order/detail", // h5订单详情
			"/api/v1/cashier/h5_order/reject", // h5拒单
			"/api/v1/cashier/h5_order/accept", // h5接单
		},
	}
}

// Login 登录
func (s *authSrv) Login(ctx context.Context, loginReq req.LoginReq) (resp.LoginResp, error) {
	var loginResp resp.LoginResp
	var token, refreshToken string
	// 验证验证码
	if !s.captchaSrv.Verify(ctx.GetGin().GetHeader("X-SIGN"), loginReq.Code) && viper.GetString("GENERAL_VERIFY_CODE") != loginReq.Code {
		return loginResp, errors.New("验证码错误")
	}
	var staff model.Staff
	if config.Server.DeployMode == "cloud" { // 云上版本
		companyStaffRepo := repository.NewCompanyStaffRepo(s.dbm.GetDB(constant.DefaultDB))
		companyStaff := companyStaffRepo.GetCompanyStaff(companyStaffRepo.WhereUsername(loginReq.Username))
		if companyStaff.CompanyUuid == 0 {
			return loginResp, errors.New("未找到绑定的商家，请确认登录信息")
		}
		if companyStaff.Uuid == 0 {
			return loginResp, errors.New("账号或密码错误")
		}
		staffRepo := repository.NewStaffRepo(s.dbm.GetDB(companyStaff.CompanyUuid))
		staff, _ = staffRepo.GetStaff(staffRepo.WhereUuid(companyStaff.Uuid), staffRepo.WithCompany())
	} else { // 离线版本
		staffRepo := repository.NewStaffRepo(s.dbm.GetDB(constant.DefaultDB))
		staff, _ = staffRepo.GetStaff(staffRepo.WhereUsername(loginReq.Username), staffRepo.WithCompany())
	}
	// 商家状态
	if staff.Company == nil {
		return loginResp, errors.New("未找到绑定的商家，请确认登录信息")
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
	if staff.Company.IsExpired() {
		return loginResp, errors.NewWithCode(constant.CodeCompanyLicenceExpired, "店铺状态已到期，如需继续使用，请联系销售代表")
	}
	if staff.Company.IsException() {
		return loginResp, errors.New("店铺状态异常，如需继续使用，请联系销售代表")
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

		// 获取员工仓库
		staffRepo := repository.NewStaffRepo(s.dbm.GetDB(staff.CompanyUuid))

		// 检查是否有未交班的收银员
		if viper.GetString("CHECK_SHIFT_HANDOVER") != "false" {
			currentStaff, _ := staffRepo.GetStaff(staffRepo.WhereDeviceId(loginReq.DeviceId), staffRepo.WhereCashierOnline())
			if currentStaff.Uuid != 0 && currentStaff.Uuid != staff.Uuid {
				return loginResp, errors.NewWithReplace("当前收银机上有未交班的账号，请联系 %s 完成交班后再登录", []string{currentStaff.GetUserName()})
			}
			// 是否已在其他收银机登录
			if staff.CashierOnline == 1 && loginReq.DeviceId != staff.BindKey {
				return loginResp, errors.NewWithReplace("收银员 %s 已在其他收银机登录未交班，请先完成交班操作", []string{staff.GetUserName()})
			}
		}

		// 更新员工信息
		updates := map[string]any{
			"cashier_online": 1,
			"bind_key":       loginReq.DeviceId,
		}
		// 创建当班日志
		if staff.CashierLoginTime == 0 || staff.CashierOnline == 0 {
			companySetting := repository.NewCompanySettingRepo(s.dbm.GetDB(staff.CompanyUuid)).Get()
			ctx.SetCompany(*staff.Company)
			ctx.SetCompanySetting(companySetting)
			shiftLog, err := s.shiftSrv.CreateWorkingLog(ctx, staff)
			if err != nil {
				return loginResp, errors.WithMessage(err, "创建当班日志失败")
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
		kitchenSetting, _ := s.settingSrv.GetKitchenSetting(ctx, companySetting, []dto.LanguageItem{})
		if kitchenSetting.IsOpen != "1" || companySetting.IsOpenKitchenKds != 1 {
			return loginResp, errors.New("当前尚未开启厨显功能，如有需要，请联系销售代表")
		}
	case constant.SourceTablet: // 平板端
		companySetting := repository.NewCompanySettingRepo(s.dbm.GetDB(staff.CompanyUuid)).Get()
		if companySetting.IsOpenTablet != 1 {
			return loginResp, errors.New("当前尚未开启平板点餐功能，如有需要，请联系销售代表")
		}
	case constant.SourceShop: // 移动管理端
	default:
		return loginResp, errors.New("登录来源错误")
	}

	// 登录时没有商家ID，补上
	ctx.SetCompanyUuid(staff.CompanyUuid)
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

	claims := auth.Claims{
		Source:      loginReq.Source,
		CompanyUuid: staff.CompanyUuid,
		StaffUuid:   staff.Uuid,
		DeviceUuid:  deviceUuid,
		DeviceId:    loginReq.DeviceId,
		Assistant:   auth.Assistant{},
	}
	// 生成 JWT token，refresh_token
	token, err = auth.GenerateToken(claims, config.JWT.Secret, config.JWT.Expire, false)
	if err != nil {
		return loginResp, errors.New("生成token失败")
	}
	refreshToken, err = auth.GenerateToken(claims, config.JWT.Secret, config.JWT.RefreshExpire, true)
	if err != nil {
		return loginResp, errors.New("生成refresh_token失败")
	}
	return resp.LoginResp{
		Token:               token,
		RefreshToken:        refreshToken,
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
		return errors.ErrInternal
	}
	if ctx.GetSource() == constant.SourceTablet {
		// 解绑桌台
		err = repository.NewDeskRepo(db).UnbindDesk(ctx.GetDeskUuid(), ctx.GetDeviceUuid())
		if err != nil {
			return errors.ErrInternal
		}
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
		source             = ctx.GetSource()
		deviceId           = ctx.GetGin().GetString(jwt.DeviceId)
		cashierSettingResp setting.CashierResp
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
	copier.Copy(&cashierSettingResp, cashierSetting)
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
	cloudBasicSetting, err := s.settingSrv.GetCloudBasicSetting(ctx)
	if err != nil {
		return cashierBase, errors.WithMessage(err)
	}
	printerSetting, err := s.settingSrv.GetPrinterSetting(ctx, nil)
	if err != nil {
		return cashierBase, errors.WithMessage(err)
	}
	return resp.CashierBase{
		Username:     staff.Username,
		CashierUuid:  staff.Uuid,
		DeviceId:     deviceId,
		DeviceRemark: deviceRemark,
		Cashier:      cashierSettingResp,
		Business:     businessSetting,
		Buffet:       buffetSetting,
		Currency:     currencySetting,
		Permissions:  permissions,
		Company: resp.Company{ // cashier
			Uuid:           company.Uuid,
			Name:           company.Name,
			Logo:           utils.AddImageDomain(company.Logo, utils.GetBaseURL(ctx.GetGin().Request), true),
			TimeZone:       companySetting.Timezone,
			ExpireTime:     company.ExpireTime,
			IsOpenMember:   companySetting.IsOpenMember,
			IsOpenBuffet:   companySetting.IsOpenBuffet,
			IsOpenH5Order:  companySetting.IsOpenH5Order,
			IsOpenRider:    companySetting.IsOpenRider(),
			IsOpenOldOrder: utils.IfInt(company.OldCompanyId > 0, 1, 0),
			IsEnableErp:    company.IsOpenErp(),
		},
		CloudBasic: cloudBasicSetting,
		Printer:    printerSetting,
		UpdateTime: time.Now().Unix(),
	}, nil
}

// AssistantBase 获取助手端基本信息
func (s *authSrv) AssistantBase(ctx context.Context) (resp.AssistantBase, error) {
	db := s.dbm.GetDB(ctx.GetCompanyUuid())
	assistantBase := resp.AssistantBase{}
	staff := helper.GetStaff(ctx.GetGin())
	staffRepo := repository.NewStaffRepo(db)
	assistantStaffUuid := ctx.GetGin().GetUint64(jwt.AssistantStaffUuid)
	assistantStaff, _ := staffRepo.GetStaff(staffRepo.WhereUuid(assistantStaffUuid))
	perms, err := s.roleAccessSrv.GetPermission(constant.AssistantRouteName, assistantStaff.Uuid, assistantStaff.CompanyUuid)
	if err != nil {
		return assistantBase, errors.WithMessage(err)
	}
	permissions := make([]string, 0, len(perms))
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
	var assistantSettingResp setting.AssistantResp
	copier.CopyWithOption(&assistantSettingResp, assistantSetting, copier.Option{IgnoreEmpty: true, DeepCopy: true})
	cloudBasicSetting, err := s.settingSrv.GetCloudBasicSetting(ctx)
	if err != nil {
		return assistantBase, errors.WithMessage(err)
	}
	printerSetting, err := s.settingSrv.GetPrinterSetting(ctx, nil)
	if err != nil {
		return assistantBase, errors.WithMessage(err)
	}
	var kitchenSettingResp setting.KitchenResp
	kitchenSetting, err := s.settingSrv.GetKitchenSetting(ctx, companySetting, []dto.LanguageItem{})
	if err != nil {
		return assistantBase, errors.WithMessage(err)
	}
	copier.CopyWithOption(&kitchenSettingResp, kitchenSetting, copier.Option{IgnoreEmpty: true, DeepCopy: true})
	clientVersion := ctx.GetGin().GetHeader("Version-Name")
	if clientVersion == "" {
		clientVersion = "0.0.0"
	}
	return resp.AssistantBase{
		Permissions: permissions,
		CashierStaff: resp.CashierStaff{
			RealName:     staff.RealName,
			Username:     staff.Username,
			DeviceId:     staff.BindKey,
			DeviceRemark: device.Remark,
		},
		AssistantStaff: resp.AssistantStaff{
			Uuid:     assistantStaff.Uuid,
			RealName: assistantStaff.RealName,
			Phone:    assistantStaff.Phone,
			DeviceId: assistantStaff.BindKey,
		},
		Buffet:     buffetSetting,
		CloudBasic: cloudBasicSetting,
		Company: resp.Company{ // assistant
			Uuid:           company.Uuid,
			Name:           company.Name,
			Logo:           utils.AddImageDomain(company.Logo, utils.GetBaseURL(ctx.GetGin().Request), true),
			TimeZone:       companySetting.Timezone,
			ExpireTime:     company.ExpireTime,
			IsOpenMember:   companySetting.IsOpenMember,
			IsOpenBuffet:   companySetting.IsOpenBuffet,
			IsOpenH5Order:  companySetting.IsOpenH5Order,
			IsOpenOldOrder: utils.IfInt(company.OldCompanyId > 0, 1, 0),
			IsOpenRider:    companySetting.IsOpenRider(),
			IsEnableErp:    company.IsOpenErp(),
		},
		Currency:      currencySetting,
		Business:      businessSetting,
		Assistant:     assistantSettingResp,
		Printer:       printerSetting,
		Kitchen:       kitchenSettingResp,
		ClientVersion: clientVersion,
		ServerVersion: utils.GetVersion(),
	}, nil
}

// TabletBase 平板端基本信息
func (s *authSrv) TabletBase(ctx context.Context) (resp.TabletBase, error) {
	company := helper.GetCompany(ctx.GetGin())
	companySetting := helper.GetCompanySetting(ctx.GetGin())

	var tabletBase resp.TabletBase
	buffetSetting, err := s.settingSrv.GetBuffetSetting(ctx, companySetting)
	if err != nil {
		return tabletBase, errors.WithMessage(err)
	}
	currencySetting, err := s.settingSrv.GetCurrencySetting(ctx)
	if err != nil {
		return tabletBase, errors.WithMessage(err)
	}
	var (
		tabletSettingResp  setting.TabletResp
		kitchenSettingResp setting.KitchenResp
	)
	tabletSetting, err := s.settingSrv.GetTabletSetting(ctx, nil)
	if err != nil {
		return tabletBase, errors.WithMessage(err)
	}
	copier.Copy(&tabletSettingResp, tabletSetting)
	kitchenSetting, err := s.settingSrv.GetKitchenSetting(ctx, companySetting, nil)
	if err != nil {
		return tabletBase, errors.WithMessage(err)
	}
	copier.CopyWithOption(&kitchenSettingResp, kitchenSetting, copier.Option{IgnoreEmpty: true, DeepCopy: true})
	cloudBasicSetting, err := s.settingSrv.GetCloudBasicSetting(ctx)
	if err != nil {
		return tabletBase, errors.WithMessage(err)
	}
	businessSetting, err := s.settingSrv.GetBusinessSetting(ctx)
	if err != nil {
		return tabletBase, errors.WithMessage(err)
	}
	clientVersion := ctx.GetGin().GetHeader("Version-Name")
	if clientVersion == "" {
		clientVersion = "0.0.0"
	}
	return resp.TabletBase{
		RealName:      ctx.GetStaff().RealName,
		ServerVersion: utils.GetVersion(),
		ClientVersion: clientVersion,
		Buffet:        buffetSetting,
		CloudBasic:    cloudBasicSetting,
		Company: resp.Company{ // tablet
			Uuid:           company.Uuid,
			Name:           company.Name,
			Logo:           utils.AddImageDomain(company.Logo, utils.GetBaseURL(ctx.GetGin().Request), true),
			TimeZone:       companySetting.Timezone,
			ExpireTime:     company.ExpireTime,
			IsOpenMember:   companySetting.IsOpenMember,
			IsOpenBuffet:   companySetting.IsOpenBuffet,
			IsOpenH5Order:  companySetting.IsOpenH5Order,
			IsOpenOldOrder: utils.IfInt(company.OldCompanyId > 0, 1, 0),
			IsOpenRider:    companySetting.IsOpenRider(),
			IsEnableErp:    company.IsOpenErp(),
		},
		Currency: currencySetting,
		Business: businessSetting,
		Tablet:   tabletSettingResp,
		Kitchen:  kitchenSettingResp,
	}, nil
}

// KitchenBase 获取厨显端基本信息
func (s *authSrv) KitchenBase(ctx context.Context) (resp.KitchenBase, error) {
	var (
		kitchenBase        resp.KitchenBase
		kitchenSettingResp setting.KitchenResp
	)
	company := helper.GetCompany(ctx.GetGin())
	companySetting := helper.GetCompanySetting(ctx.GetGin())
	kitchenSetting, err := s.settingSrv.GetKitchenSetting(ctx, companySetting, nil)
	if err != nil {
		return kitchenBase, errors.WithMessage(err)
	}
	copier.CopyWithOption(&kitchenSettingResp, kitchenSetting, copier.Option{IgnoreEmpty: true, DeepCopy: true})
	cloudBasicSetting, err := s.settingSrv.GetCloudBasicSetting(ctx)
	if err != nil {
		return kitchenBase, errors.WithMessage(err)
	}
	businessSetting, err := s.settingSrv.GetBusinessSetting(ctx)
	if err != nil {
		return kitchenBase, errors.WithMessage(err)
	}
	buffetSetting, err := s.settingSrv.GetBuffetSetting(ctx, companySetting)
	if err != nil {
		return kitchenBase, errors.WithMessage(err)
	}
	currencySetting, err := s.settingSrv.GetCurrencySetting(ctx)
	if err != nil {
		return kitchenBase, errors.WithMessage(err)
	}

	clientVersion := ctx.GetGin().GetHeader("Version-Name")
	if clientVersion == "" {
		clientVersion = "0.0.0"
	}
	return resp.KitchenBase{
		RealName:   ctx.GetStaff().RealName,
		Buffet:     buffetSetting,
		CloudBasic: cloudBasicSetting,
		Company: resp.Company{ // kitchen
			Uuid:           company.Uuid,
			Name:           company.Name,
			Logo:           utils.AddImageDomain(company.Logo, utils.GetBaseURL(ctx.GetGin().Request), true),
			TimeZone:       companySetting.Timezone,
			ExpireTime:     company.ExpireTime,
			IsOpenMember:   companySetting.IsOpenMember,
			IsOpenBuffet:   companySetting.IsOpenBuffet,
			IsOpenH5Order:  companySetting.IsOpenH5Order,
			IsOpenOldOrder: utils.IfInt(company.OldCompanyId > 0, 1, 0),
			IsOpenRider:    companySetting.IsOpenRider(),
			IsEnableErp:    company.IsOpenErp(),
		},
		Currency:      currencySetting,
		Business:      businessSetting,
		Kitchen:       kitchenSettingResp,
		ServerVersion: utils.GetVersion(),
		ClientVersion: clientVersion,
	}, nil
}

// Auth 鉴权
func (s *authSrv) Auth(ctx context.Context, auth req.Authenticate) (model.Company, model.CompanySetting, model.Staff, model.Desk, error) {
	var (
		company        model.Company
		companySetting model.CompanySetting
		staff          model.Staff
		desk           model.Desk
		db             = s.dbm.GetDB(auth.CompanyUuid)
	)
	// 如果使用旧的token，但是商家已被删除，优化提示
	if db == nil {
		return company, companySetting, staff, desk, errors.New("未找到绑定的商家，请确认登录信息")
	}
	staffRepo := repository.NewStaffRepo(db)
	staff, err := staffRepo.GetStaff(staffRepo.WhereUuid(auth.StaffUuid), staffRepo.WithCompany(), staffRepo.WithCompanySetting())
	if err != nil {
		logger.Logger.Error("获取员工信息失败", zap.Error(err))
		return company, companySetting, staff, desk, errors.NewWithCode(constant.CodeTokenInvalid, "没有找到用户信息")
	}
	if staff.Uuid == 0 {
		return company, companySetting, staff, desk, errors.New("用户不存在")
	}
	// 修改密码后，token失效
	if staff.PasswordChangeTime > auth.TokenIssuedAt {
		return company, companySetting, staff, desk, errors.NewWithCode(constant.CodeTokenInvalid, "登录失效，请重新登录")
	}
	if staff.DeleteTime != 0 {
		return company, companySetting, staff, desk, errors.New("用户被删除")
	}
	if staff.Company != nil {
		company = *staff.Company
		if staff.Company.CompanySetting != nil {
			companySetting = *staff.Company.CompanySetting
		}
	}
	if company.Uuid == 0 || companySetting.Uuid == 0 {
		return company, companySetting, staff, desk, errors.New("未找到绑定的商家，请确认登录信息")
	}
	if company.IsExpired() {
		return company, companySetting, staff, desk, errors.NewWithCode(constant.CodeCompanyLicenceExpired, "店铺状态已到期，如需继续使用，请联系销售代表")
	}
	if company.IsException() {
		return company, companySetting, staff, desk, errors.New("店铺状态异常，如需继续使用，请联系销售代表")
	}
	// 验证设备是否绑定
	deviceId := auth.DeviceId
	if auth.Source == constant.SourceAssistant && auth.Assistant.DeviceId != "" { // 登录了点餐助手，且绑定了收银机
		deviceId = auth.Assistant.DeviceId
	}
	if auth.Source != constant.SourceShop && !s.deviceSrv.IsDeviceBind(auth.CompanyUuid, auth.Source, deviceId) {
		if printerData := repository.NewPrinterLogRepo(db).GetShiftPrinterData(deviceId); printerData != nil {
			return company, companySetting, staff, desk, errors.NewWithCodeAndData(constant.CodeTokenInvalid, map[string]interface{}{
				"printer_data": printerData,
			}, "设备已解绑，请重新绑定")
		}
		return company, companySetting, staff, desk, errors.NewWithCode(constant.CodeTokenInvalid, "设备已解绑，请重新绑定")
	}
	if slices.Contains(s.memberFunctionRoutes, auth.UrlPath) && companySetting.IsOpenMember != 1 {
		return company, companySetting, staff, desk, errors.New("当前尚未开启会员功能，如有需要，请联系销售代表")
	}
	if slices.Contains(s.h5AcceptOrderFunctionRoutes, auth.UrlPath) && companySetting.IsOpenH5Order != 1 {
		return company, companySetting, staff, desk, errors.New("当前尚未开启扫码点餐接单功能，如有需要，请联系销售代表")
	}
	switch auth.Source {
	case constant.SourceCashier: // 收银端
		{
			cashierSetting, err := s.GetCashierSetting(ctx)
			if err != nil {
				return company, companySetting, staff, desk, errors.NewWithCode(constant.CodeSystemError, "系统错误，请稍候再试")
			}
			// 检查收银机设置-收银用餐是否开启
			if !s.isCashierOpen(cashierSetting, auth.UrlPath) {
				return company, companySetting, staff, desk, errors.NewWithCode(constant.CodeCashierOrderMethodNotOpen, "收银用餐已关闭，请选择其他用餐方式")
			}
			// 检查收银机设置-桌台用餐是否开启
			if !s.isTableOpen(ctx, cashierSetting, auth.UrlPath) {
				return company, companySetting, staff, desk, errors.NewWithCode(constant.CodeCashierOrderMethodNotOpen, "桌台用餐已关闭，请选择其他用餐方式")
			}
			// 判断权限
			permissions, err := s.roleAccessSrv.GetApiPermission(staff.Uuid, auth.CompanyUuid)
			if err != nil {
				return company, companySetting, staff, desk, errors.NewWithCode(constant.CodeUnauthorized, "当前无权限，请联系管理员")
			}
			permission := constant.CashierPermissions[auth.UrlPath]
			if permission != "" && !slices.Contains(permissions, permission) {
				return company, companySetting, staff, desk, errors.NewWithCode(constant.CodeUnauthorized, "当前无权限，请联系管理员")
			}
			// 已交班，只能打印交班单、退出登录、获取基本信息、轮询打印数据，轮询未处理消息、获取广告
			if staff.DutyNo == "" && !slices.Contains([]string{"/api/v1/cashier/shift/printer", "/api/v1/cashier/logout"}, auth.UrlPath) {
				// 判断客户端版本，低于 2.3 的就返回 -101，直接退出
				// 高于等于的就返-108，能弹出来交班弹窗
				if ctx.Version(context.GTE, "2.3.0") { // 高于等于 2.3.0
					// 获取缓存
					if cachedSubmitShift, err := s.shiftSrv.GetCachedSubmitShift(ctx); err != nil || cachedSubmitShift == nil {
						return company, companySetting, staff, desk, errors.NewWithCode(constant.CodeTokenExpired, "当前班次不存在")
					} else {
						return company, companySetting, staff, desk, errors.NewWithCodeAndData(constant.CodeCashierHandedOver, *cachedSubmitShift, "已交班")
					}
				} else { // 低于 2.3.0
					return company, companySetting, staff, desk, errors.NewWithCode(constant.CodeTokenExpired, "当前班次不存在")
				}
			}
		}
	case constant.SourceAssistant: // 点餐助手端
		{
			cashierSetting, err := s.GetCashierSetting(ctx)
			if err != nil {
				return company, companySetting, staff, desk, errors.NewWithCode(constant.CodeSystemError, "系统错误，请稍候再试")
			}
			if companySetting.IsOpenAssistant != 1 {
				return company, companySetting, staff, desk, errors.NewWithCode(constant.CodeFunctionDisabled, "当前尚未开启点餐助手功能，如有需要，请联系销售代表")
			}
			if !slices.Contains(s.assistantRoutes, auth.UrlPath) { // 除了这些接口外，其他都需要判断收银机状态
				deviceRepo := repository.NewDeviceRepo(db)
				cashierDevice, _ := deviceRepo.GetDevice(deviceRepo.WhereSource(constant.SourceCashier), deviceRepo.WhereSn(auth.DeviceId))
				if cashierDevice.Uuid == 0 {
					return company, companySetting, staff, desk, errors.NewWithCode(constant.CodeCashierNotLogin, "收银员设备已解绑，请重新绑定")
				}
				if staff.DutyNo == "" || auth.Assistant.StaffUuid == 0 {
					return company, companySetting, staff, desk, errors.NewWithCode(constant.CodeCashierNotLogin, "收银员登录信息错误，请重新登录")
				}
			}
			// 检查收银机设置-桌台用餐是否开启
			if !s.isTableOpen(ctx, cashierSetting, auth.UrlPath) {
				return company, companySetting, staff, desk, errors.NewWithCode(constant.CodeCashierOrderMethodNotOpen, "桌台用餐已关闭，请选择其他用餐方式")
			}
		}
	case constant.SourceTablet: // 平板端
		{
			cashierSetting, err := s.GetCashierSetting(ctx)
			if err != nil {
				return company, companySetting, staff, desk, errors.NewWithCode(constant.CodeSystemError, "系统错误，请稍候再试")
			}
			if companySetting.IsOpenTablet != 1 {
				return company, companySetting, staff, desk, errors.NewWithCode(constant.CodeFunctionDisabled, "当前尚未开启平板点餐功能，如有需要，请联系销售代表")
			}
			if !slices.Contains(s.tabletRoutes, auth.UrlPath) { // 除了这些接口外，其他都需要判断是否绑定了桌台
				deskRepo := repository.NewDeskRepo(db)
				var err error
				desk, err = deskRepo.GetDesk(deskRepo.WhereDeviceUuid(ctx.GetDeviceUuid()))
				if err != nil {
					return company, companySetting, staff, desk, errors.NewWithCode(constant.CodeTabletNotBindDesk, "桌台未绑定")
				}
			}
			// 检查收银机设置-桌台用餐是否开启
			if !s.isTableOpen(ctx, cashierSetting, auth.UrlPath) {
				return company, companySetting, staff, desk, errors.NewWithCode(constant.CodeCashierOrderMethodNotOpen, "桌台用餐已关闭，请选择其他用餐方式")
			}
		}
	case constant.SourceKitchen: // 厨显端
		{
			kitchenSetting, _ := s.settingSrv.GetKitchenSetting(ctx, companySetting, []dto.LanguageItem{})
			if kitchenSetting.IsOpen != "1" || companySetting.IsOpenKitchenKds != 1 {
				return company, companySetting, staff, desk, errors.NewWithCode(constant.CodeFunctionDisabled, "当前尚未开启厨显功能，如有需要，请联系销售代表")
			}
		}
	}
	return company, companySetting, staff, desk, nil
}

func (s *authSrv) AuthDesk(ctx context.Context, qrcodeToken string) (*model.Company, error) {
	companyUuid := ctx.GetCompanyUuid()
	deskUuid := ctx.GetDeskUuid()
	db := s.dbm.GetDB(companyUuid)
	company, err := repository.NewCompanyRepo(db).GetCompanyInfoByUuid(companyUuid)
	if err != nil || company.IsExpired() || company.IsDelete() {
		return nil, errors.NewWithCode(constant.CodeTokenInvalid, "二维码已失效，请联系商家")
	}

	companySetting := repository.NewCompanySettingRepo(db).Get()
	if companySetting.IsOpenH5 != 1 {
		return nil, errors.NewWithCode(constant.CodeTokenInvalid, "二维码已失效，请联系商家")
	}

	deskInfo, err := repository.NewDeskRepo(db).GetDeskInfo(deskUuid)
	if err != nil || deskInfo.IsDisableDesk() || deskInfo.IsDelete() || deskInfo.QrcodeToken != qrcodeToken {
		return nil, errors.NewWithCode(constant.CodeTokenInvalid, "二维码已失效，请联系商家")
	}

	cashierSetting, _ := s.settingSrv.GetCashierSetting(ctx, []dto.LanguageItem{})
	if cashierSetting.OrderMethod.IsTableOrder != "1" {
		return nil, errors.NewWithCode(constant.CodeCashierOrderMethodNotOpen, "桌台用餐已关闭，请选择其他用餐方式")
	}

	return company, nil
}

func (s *authSrv) AuthMenu(ctx context.Context, qrcodeToken string) (*model.Company, error) {
	companyUuid := ctx.GetCompanyUuid()
	db := s.dbm.GetDB(companyUuid)
	company, err := repository.NewCompanyRepo(db).GetCompanyInfoByUuid(companyUuid)
	if err != nil || company.IsExpired() || company.IsDelete() {
		return nil, errors.NewWithCode(constant.CodeTokenInvalid, "二维码已失效，请联系商家")
	}

	businessSetting, err := s.settingSrv.GetBusinessSetting(ctx)
	if err != nil || businessSetting.QrCode != qrcodeToken {
		return nil, errors.NewWithCode(constant.CodeTokenInvalid, "二维码已失效，请联系商家")
	}

	cashierSetting, _ := s.settingSrv.GetCashierSetting(ctx, []dto.LanguageItem{})
	if cashierSetting.OrderMethod.IsTableOrder != "1" {
		return nil, errors.NewWithCode(constant.CodeCashierOrderMethodNotOpen, "桌台用餐已关闭，请选择其他用餐方式")
	}
	return company, nil
}

func (s *authSrv) GetCashierSetting(ctx context.Context) (*setting.Cashier, error) {
	cashierSetting, err := s.settingSrv.GetCashierSetting(ctx, []dto.LanguageItem{})
	if err != nil {
		return nil, err
	}
	return &cashierSetting, nil
}

// 检查收银机设置-收银用餐是否开启
func (s *authSrv) isCashierOpen(cashierSetting *setting.Cashier, pathUrl string) bool {
	// 检查收银用餐是否开启
	if cashierSetting.OrderMethod.IsCashierOrder != "1" && regexp.MustCompile(`^/api/v\d+/cashier/instant/`).Match([]byte(pathUrl)) {
		return false
	}
	return true
}

// 检查收银机设置-桌台用餐是否开启
func (s *authSrv) isTableOpen(ctx context.Context, cashierSetting *setting.Cashier, pathUrl string) bool {
	if cashierSetting.OrderMethod.IsTableOrder != "1" &&
		(slices.Contains([]string{constant.SourceAssistant, constant.SourceTablet}, ctx.GetSource()) ||
			(ctx.GetSource() == constant.SourceCashier && regexp.MustCompile(`^/api/v\d+/cashier/desk/`).Match([]byte(pathUrl)))) {
		return false
	}
	return true
}

// BindCashier 绑定收银机
func (s *authSrv) BindCashier(ctx context.Context, bindReq req.BindCashierReq) (string, string, error) {
	var newToken, refreshToken string
	companyUuid := helper.GetCompanyUuid(ctx.GetGin())
	if helper.GetSource(ctx.GetGin()) != constant.SourceAssistant {
		return newToken, refreshToken, errors.New("用户信息错误")
	}
	staffRepo := repository.NewStaffRepo(s.dbm.GetDB(companyUuid))
	staff, _ := staffRepo.GetStaff(staffRepo.WhereUuid(bindReq.CashierUuid), staffRepo.WhereDeviceId(bindReq.DeviceId), staffRepo.WithCompany(), staffRepo.WithCompanySetting())
	// 检查传递的收银设备
	if staff.Uuid == 0 {
		return newToken, refreshToken, errors.New("用户不存在")
	}
	if staff.CashierOnline != 1 {
		return newToken, refreshToken, errors.New("收银设备不在线")
	}
	// 判断收银员状态
	if staff.DeleteTime != 0 {
		return newToken, refreshToken, errors.New("账号被删除，请联系管理员")
	}
	if staff.IsDisable == 1 {
		return newToken, refreshToken, errors.New("账号被禁用，请联系管理员")
	}
	if staff.Company == nil || staff.Company.Uuid == 0 || staff.Company.DeleteTime != 0 {
		return newToken, refreshToken, errors.New("未找到绑定的商家，请确认登录信息")
	}
	// 是否已开启点餐助手功能
	if staff.Company.CompanySetting == nil || staff.Company.CompanySetting.IsOpenAssistant != 1 {
		return newToken, refreshToken, errors.New("当前尚未开启点餐助手功能，如有需要，请联系销售代表")
	}

	claims := auth.Claims{
		Source:      constant.SourceAssistant,
		CompanyUuid: companyUuid,
		StaffUuid:   bindReq.CashierUuid,
		DeviceUuid:  ctx.GetDeviceUuid(),
		DeviceId:    bindReq.DeviceId,
		Assistant: auth.Assistant{
			DeviceId:  ctx.GetGin().GetString(jwt.DeviceId),
			StaffUuid: ctx.GetGin().GetUint64(jwt.StaffUuid),
		},
	}
	// 生成新的 JWT token，将点餐助手设备ID和员工Uuid单独存放
	newToken, err := auth.GenerateToken(claims, config.JWT.Secret, config.JWT.Expire, false)
	if err != nil {
		return newToken, refreshToken, errors.New("生成token失败")
	}
	newRefreshToken, err := auth.GenerateToken(claims, config.JWT.Secret, config.JWT.RefreshExpire, true)
	if err != nil {
		return newToken, refreshToken, errors.New("生成refresh_token失败")
	}
	return newToken, newRefreshToken, nil
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

// RefreshToken 刷新token
func (s *authSrv) RefreshToken(ctx context.Context) (resp.LoginResp, error) {
	var (
		newToken, newRefreshToken string
		loginResp                 resp.LoginResp
		err                       error
	)
	claims := auth.Claims{
		Source:      ctx.GetSource(),
		CompanyUuid: ctx.GetCompanyUuid(),
		StaffUuid:   ctx.GetStaffUuid(),
		DeviceUuid:  ctx.GetDeviceUuid(),
		DeviceId:    ctx.GetDeviceSn(),
		Assistant: auth.Assistant{
			DeviceId:  ctx.GetGin().GetString(jwt.AssistantDeviceId),
			StaffUuid: ctx.GetGin().GetUint64(jwt.AssistantStaffUuid),
		},
	}
	// 生成 JWT new token，new refresh token
	newToken, err = auth.GenerateToken(claims, config.JWT.Secret, config.JWT.Expire, false)
	if err != nil {
		return loginResp, errors.New("生成token失败")
	}
	newRefreshToken, err = auth.GenerateToken(claims, config.JWT.Secret, config.JWT.RefreshExpire, true)
	if err != nil {
		return loginResp, errors.New("生成refresh_token失败")
	}
	return resp.LoginResp{
		Token:               newToken,
		RefreshToken:        newRefreshToken,
		CashierIsFirstLogin: false,
	}, nil
}

func (s *authSrv) ShopBase(ctx context.Context) (resp.ShopBase, error) {
	var shopBase resp.ShopBase
	company := ctx.GetCompany()
	companySetting := ctx.GetCompanySetting()
	staff := ctx.GetStaff()
	var (
		source   = ctx.GetSource()
		deviceId = ctx.GetGin().GetString(jwt.DeviceId)
	)
	deviceRemark := s.deviceSrv.GetRemark(company.Uuid, source, deviceId)
	// 判断权限
	permissions := []*resp.Permission{}

	businessSetting, err := s.settingSrv.GetBusinessSetting(ctx)
	if err != nil {
		return shopBase, errors.WithMessage(err)
	}
	buffetSetting, err := s.settingSrv.GetBuffetSetting(ctx, companySetting)
	if err != nil {
		return shopBase, errors.WithMessage(err)
	}
	currencySetting, err := s.settingSrv.GetCurrencySetting(ctx)
	if err != nil {
		return shopBase, errors.WithMessage(err)
	}
	cloudBasicSetting, err := s.settingSrv.GetCloudBasicSetting(ctx)
	if err != nil {
		return shopBase, errors.WithMessage(err)
	}
	storeSetting, err := s.settingSrv.GetStoreSetting(ctx)
	if err != nil {
		return shopBase, errors.WithMessage(err)
	}
	taxSetting, err := s.settingSrv.GetTaxRateSetting(ctx)
	if err != nil {
		return shopBase, errors.WithMessage(err)
	}
	return resp.ShopBase{
		Username:     staff.Username,
		ProfileUuid:  staff.Uuid,
		DeviceId:     deviceId,
		DeviceRemark: deviceRemark,
		Permissions:  permissions,
		Phone:        staff.Phone,

		Business: businessSetting,
		Buffet:   buffetSetting,
		Currency: currencySetting,
		Company: resp.Company{ // shop
			Uuid:           company.Uuid,
			Name:           company.Name,
			Logo:           utils.AddImageDomain(company.Logo, utils.GetBaseURL(ctx.GetGin().Request), true),
			TimeZone:       companySetting.Timezone,
			ExpireTime:     company.ExpireTime,
			IsOpenMember:   companySetting.IsOpenMember,
			IsOpenBuffet:   companySetting.IsOpenBuffet,
			IsOpenH5Order:  companySetting.IsOpenH5Order,
			IsOpenOldOrder: utils.IfInt(company.OldCompanyId > 0, 1, 0),
			IsOpenRider:    companySetting.IsOpenRider(),
			IsEnableErp:    company.IsOpenErp(),
		},
		CloudBasic: cloudBasicSetting,
		Profile: resp.ShopProfile{
			Address:         storeSetting.Address,
			Coordinates:     storeSetting.Coordinates,
			IpWhiteList:     storeSetting.IPWhiteList,
			Phone:           storeSetting.Phone,
			TaxNumber:       storeSetting.TaxNumber,
			TimeZoneList:    storeSetting.TimeZoneList,
			DefaultLanguage: companySetting.GetDefaultLanguage(),
			LanguageList:    storeSetting.Language,
			Language:        companySetting.GetLanguages(),
			CompanyName:     storeSetting.Company,
		},
		IsTtposSite:   companySetting.IsTtposSite(),
		IsHeadquarter: companySetting.IsHeadquarter(),
		UpdateTime:    time.Now().Unix(),
		ServerVersion: utils.GetVersion(),
		IsOpenTax:     taxSetting.IsOpen == "1",
	}, nil
}

// 修改密码
func (s *authSrv) ChangePassword(ctx context.Context, changePasswordReq req.ChangePasswordReq) error {
	if ctx.GetSource() != constant.SourceShop {
		return errors.New("当前无权限，请联系管理员")
	}
	staff := ctx.GetStaff()
	staffRepo := repository.NewStaffRepo(s.dbm.GetDB(staff.CompanyUuid))
	staff, err := staffRepo.GetStaff(staffRepo.WhereUuid(staff.Uuid), staffRepo.WithCompany(), staffRepo.WithCompanySetting())
	if err != nil {
		return errors.WithMessage(err)
	}
	if staff.Password != utils.EncryptPassword(changePasswordReq.OldPassword) {
		return errors.New("旧密码错误")
	}
	staff.Password = utils.EncryptPassword(changePasswordReq.NewPassword)
	return staffRepo.Update(staff.Uuid, map[string]any{
		"password": staff.Password,
	})
}
