# Auth Service 认证授权服务说明文档

## 📋 概述

`service/auth.go` 是 TTPOS 系统的核心认证授权服务，负责处理多端登录认证、权限验证、会话管理等功能。该服务支持收银端、点餐助手端、平板端、厨显端和移动管理端等多种客户端类型。

**文件路径**: `/home/coder/workspaces/ttpos-server-go/main/app/service/auth.go`  
**代码行数**: 1047 行  
**接口定义**: `IAuthSrv`  
**实现结构**: `authSrv`

---

## 🏗️ 架构设计

### 接口定义 (IAuthSrv)

```go
type IAuthSrv interface {
    Login(ctx context.Context, loginReq req.LoginReq) (resp.LoginResp, error)
    Logout(ctx context.Context) error
    CashierBase(ctx context.Context) (resp.CashierBase, error)
    AssistantBase(ctx context.Context) (resp.AssistantBase, error)
    TabletBase(ctx context.Context) (resp.TabletBase, error)
    KitchenBase(ctx context.Context) (resp.KitchenBase, error)
    ShopBase(ctx context.Context) (resp.ShopBase, error)
    Auth(ctx context.Context, auth req.Authenticate) (model.Company, model.CompanySetting, model.Staff, model.Desk, error)
    AuthDesk(ctx context.Context, qrcodeToken string) (*model.Company, error)
    AuthMenu(ctx context.Context, qrcodeToken string) (*model.Company, error)
    BindCashier(ctx context.Context, bindReq req.BindCashierReq) (string, string, error)
    GetOnlineCashiers(companyUuid uint64) resp.OnlineCashierList
    RefreshToken(ctx context.Context) (resp.LoginResp, error)
    ChangePassword(ctx context.Context, changePasswordReq req.ChangePasswordReq) error
}
```

### 依赖服务

```go
type authSrv struct {
    dbm           *database.DBManager      // 数据库管理器
    captchaSrv    ICaptchaSrv             // 验证码服务
    roleAccessSrv IRoleAccessSrv          // 角色权限服务
    deviceSrv     IDeviceSrv              // 设备管理服务
    shiftSrv      IStaffShiftSrv          // 员工交班服务
    settingSrv    settingSrv.ISrv         // 系统设置服务
    
    // 路由白名单配置
    assistantRoutes             []string  // 助手端白名单路由
    tabletRoutes                []string  // 平板端白名单路由
    memberFunctionRoutes        []string  // 会员功能路由
    h5AcceptOrderFunctionRoutes []string  // H5接单功能路由
}
```

---

## 🎯 核心功能

### 1. 登录功能 (Login)

**功能描述**: 处理多端用户登录，生成 JWT token 和 refresh token。

#### 支持的登录来源

| 来源常量 | 说明 | 特殊验证 |
|---------|------|---------|
| `constant.SourceCashier` | 收银端 | 需验证权限、检查交班状态、创建当班日志 |
| `constant.SourceAssistant` | 点餐助手端 | 需验证助手功能是否开启 |
| `constant.SourceKitchen` | 厨显端 | 需验证厨显功能是否开启 |
| `constant.SourceTablet` | 平板端 | 需验证平板功能是否开启 |
| `constant.SourceShop` | 移动管理端 | 基础验证 |

#### 登录流程

```
1. 验证验证码 (captchaSrv.Verify)
   ↓
2. 根据部署模式查询员工信息
   - 云上版本: 通过 company_staff 表查询
   - 离线版本: 直接查询 staff 表
   ↓
3. 验证账号密码
   - 密码加密对比: utils.EncryptPassword()
   ↓
4. 检查员工状态
   - 是否删除 (delete_time != 0)
   - 是否禁用 (is_disable == 1)
   ↓
5. 检查商家状态
   - 是否过期: company.IsExpired()
   - 是否异常: company.IsException()
   ↓
6. 根据来源进行特殊处理
   - 收银端: 验证权限、检查交班、创建当班日志
   - 其他端: 验证功能开关
   ↓
7. 添加设备绑定记录 (deviceSrv.AddDevice)
   ↓
8. 生成 JWT Token 和 Refresh Token
   ↓
9. 返回登录响应
```

#### 收银端特殊处理

```go
// 1. 验证权限
permissions, err := s.roleAccessSrv.GetPermission(constant.CashierRouteName, staff.Uuid, staff.CompanyUuid)

// 2. 检查是否有未交班的收银员
currentStaff, _ := staffRepo.GetStaff(
    staffRepo.WhereDeviceId(loginReq.DeviceId), 
    staffRepo.WhereCashierOnline()
)

// 3. 创建当班日志
shiftLog, err := s.shiftSrv.CreateWorkingLog(ctx, staff)

// 4. 更新员工信息
updates := map[string]any{
    "cashier_online":     1,
    "bind_key":           loginReq.DeviceId,
    "cashier_login_time": shiftLog.ShiftStartTime,
    "duty_no":            shiftLog.ShiftNo,
}
```

#### 响应数据

```go
type LoginResp struct {
    Token               string  // JWT访问令牌
    RefreshToken        string  // 刷新令牌
    CashierIsFirstLogin bool    // 是否首次登录
    NeedChangePassword  bool    // 是否需要修改密码
}
```

---

### 2. 鉴权功能 (Auth)

**功能描述**: 验证用户身份和权限，是所有受保护接口的前置验证。

#### 鉴权流程

```
1. 获取员工信息 (包含商家和商家设置)
   ↓
2. 验证员工状态
   - 用户是否存在
   - 密码修改时间验证 (token是否失效)
   - 用户是否被删除
   ↓
3. 验证商家状态
   - 商家是否存在
   - 商家是否过期
   - 商家是否异常
   ↓
4. 验证设备绑定
   - 设备是否已绑定 (除移动管理端)
   ↓
5. 验证功能开关
   - 会员功能路由检查
   - H5接单功能路由检查
   ↓
6. 根据来源进行特定验证
   - 收银端: 检查收银设置、桌台设置、API权限、交班状态
   - 助手端: 检查助手功能、收银机状态、桌台设置
   - 平板端: 检查平板功能、桌台绑定、桌台设置
   - 厨显端: 检查厨显功能
   ↓
7. 返回商家、商家设置、员工、桌台信息
```

#### 收银端权限验证

```go
// 1. 检查收银用餐是否开启
if !s.isCashierOpen(cashierSetting, auth.UrlPath) {
    return errors.NewWithCode(constant.CodeCashierOrderMethodNotOpen, "收银用餐已关闭")
}

// 2. 检查桌台用餐是否开启
if !s.isTableOpen(ctx, cashierSetting, auth.UrlPath) {
    return errors.NewWithCode(constant.CodeCashierOrderMethodNotOpen, "桌台用餐已关闭")
}

// 3. 验证API权限
permissions, err := s.roleAccessSrv.GetApiPermission(staff.Uuid, auth.CompanyUuid)
permission := constant.CashierPermissions[auth.UrlPath]
if permission != "" && !slices.Contains(permissions, permission) {
    return errors.NewWithCode(constant.CodeUnauthorized, "当前无权限")
}

// 4. 验证交班状态
if staff.DutyNo == "" && !slices.Contains([]string{"/api/v1/cashier/shift/printer", "/api/v1/cashier/logout"}, auth.UrlPath) {
    // 根据客户端版本返回不同错误码
    if ctx.Version(context.GTE, "2.3.0") {
        return errors.NewWithCodeAndData(constant.CodeCashierHandedOver, cachedSubmitShift, "已交班")
    } else {
        return errors.NewWithCode(constant.CodeTokenExpired, "当前班次不存在")
    }
}
```

#### 助手端特殊验证

```go
// 白名单路由，无需验证收银机状态
assistantRoutes := []string{
    "/api/v1/assistant/online_cashiers",
    "/api/v1/assistant/bind_cashier",
    "/api/v1/assistant/verify_lock_password",
}

// 其他路由需要验证收银机在线状态
if !slices.Contains(s.assistantRoutes, auth.UrlPath) {
    cashierDevice, _ := deviceRepo.GetDevice(
        deviceRepo.WhereSource(constant.SourceCashier), 
        deviceRepo.WhereSn(auth.DeviceId)
    )
    if cashierDevice.Uuid == 0 {
        return errors.NewWithCode(constant.CodeCashierNotLogin, "收银员设备已解绑")
    }
}
```

#### 平板端桌台验证

```go
// 白名单路由
tabletRoutes := []string{
    "/api/v1/tablet/desk/list",
    "/api/v1/tablet/desk/bind",
    "/api/v1/tablet/base",
    "/api/v1/tablet/logout",
    "/api/v1/tablet/check_update",
    "/api/v1/tablet/verify_advanced_password",
    // ... 其他路由
}

// 其他路由需要验证桌台绑定
if !slices.Contains(s.tabletRoutes, auth.UrlPath) {
    desk, err = deskRepo.GetDesk(deskRepo.WhereDeviceUuid(ctx.GetDeviceUuid()))
    if err != nil {
        return errors.NewWithCode(constant.CodeTabletNotBindDesk, "桌台未绑定")
    }
}
```

---

### 3. 二维码鉴权

#### AuthDesk - 桌台二维码鉴权

**应用场景**: 顾客扫描桌台二维码点餐

**验证流程**:
```
1. 验证商家状态 (是否过期、是否删除)
   ↓
2. 验证 H5 功能是否开启 (companySetting.IsOpenH5)
   ↓
3. 验证桌台信息
   - 桌台是否禁用
   - 桌台是否删除
   - 二维码 token 是否匹配
   ↓
4. 验证桌台用餐是否开启
   ↓
5. 返回商家信息
```

#### AuthMenu - 菜单二维码鉴权

**应用场景**: 顾客扫描菜单二维码查看菜单

**验证流程**:
```
1. 验证商家状态
   ↓
2. 验证业务设置中的二维码 token
   ↓
3. 验证桌台用餐是否开启
   ↓
4. 返回商家信息
```

---

### 4. 获取基本信息

系统为不同端提供了专门的基本信息获取接口：

#### CashierBase - 收银端基本信息

**返回内容**:
```go
type CashierBase struct {
    Username     string              // 用户名
    CashierUuid  uint64              // 收银员UUID
    DeviceId     string              // 设备ID
    DeviceRemark string              // 设备备注
    Cashier      CashierResp         // 收银机设置
    Business     BusinessResp        // 业务设置
    Buffet       BuffetResp          // 自助餐设置
    Currency     CurrencyResp        // 货币设置
    Permissions  []*Permission       // 权限列表
    Company      Company             // 商家信息
    CloudBasic   CloudBasicResp      // 云端基础设置
    Printer      PrinterResp         // 打印机设置
    UpdateTime   int64               // 更新时间
}
```

**特殊处理**:
- 验证收银端权限
- 获取设备备注信息
- 组装多种设置信息

#### AssistantBase - 助手端基本信息

**返回内容**:
```go
type AssistantBase struct {
    Permissions    []string          // 权限路径列表
    CashierStaff   CashierStaff      // 收银员信息
    AssistantStaff AssistantStaff    // 助手员工信息
    Buffet         BuffetResp        // 自助餐设置
    CloudBasic     CloudBasicResp    // 云端基础设置
    Company        Company           // 商家信息
    Currency       CurrencyResp      // 货币设置
    Business       BusinessResp      // 业务设置
    Assistant      AssistantResp     // 助手设置
    Printer        PrinterResp       // 打印机设置
    Kitchen        KitchenResp       // 厨显设置
    ClientVersion  string            // 客户端版本
    ServerVersion  string            // 服务端版本
}
```

**特殊处理**:
- 获取助手员工权限
- 同时返回收银员和助手员工信息
- 包含厨显设置（用于厨打功能）

#### TabletBase - 平板端基本信息

**返回内容**:
```go
type TabletBase struct {
    RealName      string          // 员工真实姓名
    ServerVersion string          // 服务端版本
    ClientVersion string          // 客户端版本
    Buffet        BuffetResp      // 自助餐设置
    CloudBasic    CloudBasicResp  // 云端基础设置
    Company       Company         // 商家信息
    Currency      CurrencyResp    // 货币设置
    Business      BusinessResp    // 业务设置
    Tablet        TabletResp      // 平板设置
    Kitchen       KitchenResp     // 厨显设置
}
```

#### KitchenBase - 厨显端基本信息

**返回内容**:
```go
type KitchenBase struct {
    RealName      string          // 员工真实姓名
    Buffet        BuffetResp      // 自助餐设置
    CloudBasic    CloudBasicResp  // 云端基础设置
    Company       Company         // 商家信息
    Currency      CurrencyResp    // 货币设置
    Business      BusinessResp    // 业务设置
    Kitchen       KitchenResp     // 厨显设置
    ServerVersion string          // 服务端版本
    ClientVersion string          // 客户端版本
}
```

#### ShopBase - 移动管理端基本信息

**返回内容**:
```go
type ShopBase struct {
    Username      string          // 用户名
    ProfileUuid   uint64          // 员工UUID
    DeviceId      string          // 设备ID
    DeviceRemark  string          // 设备备注
    Permissions   []*Permission   // 权限列表
    Phone         string          // 手机号
    Business      BusinessResp    // 业务设置
    Buffet        BuffetResp      // 自助餐设置
    Currency      CurrencyResp    // 货币设置
    Company       Company         // 商家信息
    CloudBasic    CloudBasicResp  // 云端基础设置
    Profile       ShopProfile     // 店铺资料
    IsTtposSite   bool           // 是否TTPOS站点
    IsHeadquarter bool           // 是否总部
    UpdateTime    int64          // 更新时间
    ServerVersion string         // 服务端版本
    IsOpenTax     bool           // 是否开启税费
    IsSyncing     bool           // 是否正在同步
    LastSyncTime  int64          // 最后同步时间
    HasChildren   bool           // 是否有子店铺
}
```

**特殊处理**:
- 包含店铺详细资料
- 包含同步状态信息
- 包含总部/分店关系信息

---

### 5. 点餐助手绑定收银机 (BindCashier)

**功能描述**: 将点餐助手设备绑定到在线收银机，实现协同工作。

**绑定流程**:
```
1. 验证来源是否为助手端
   ↓
2. 查询收银员信息
   - 验证收银员UUID
   - 验证设备ID
   - 验证在线状态 (cashier_online == 1)
   ↓
3. 验证收银员状态
   - 是否删除
   - 是否禁用
   ↓
4. 验证商家状态
   - 商家是否存在
   - 商家是否删除
   ↓
5. 验证助手功能是否开启
   ↓
6. 生成新的 JWT Token
   - 包含收银员信息
   - 包含助手员工信息
   ↓
7. 返回新的 token 和 refresh_token
```

**Token 结构**:
```go
claims := auth.Claims{
    Source:      constant.SourceAssistant,  // 来源：助手端
    CompanyUuid: companyUuid,               // 商家UUID
    StaffUuid:   bindReq.CashierUuid,       // 收银员UUID
    DeviceUuid:  ctx.GetDeviceUuid(),       // 助手设备UUID
    DeviceId:    bindReq.DeviceId,          // 收银机设备ID
    Assistant: auth.Assistant{
        DeviceId:  ctx.GetGin().GetString(jwt.DeviceId),      // 助手设备ID
        StaffUuid: ctx.GetGin().GetUint64(jwt.StaffUuid),     // 助手员工UUID
    },
}
```

**使用场景**:
- 点餐助手登录后，需要选择并绑定一台在线的收银机
- 绑定后，助手端的订单操作会关联到对应的收银机
- 收银机交班后，助手端需要重新绑定

---

### 6. 获取在线收银机列表 (GetOnlineCashiers)

**功能描述**: 查询当前商家所有在线的收银机，供助手端选择绑定。

**查询条件**:
```go
staffs := staffRepo.GetStaffs(
    staffRepo.WhereCashierOnline(),              // cashier_online = 1
    staffRepo.WithDevice(constant.SettingCashier) // 预加载设备信息
)
```

**返回数据**:
```go
type OnlineCashier struct {
    CashierUuid uint64  // 收银员UUID
    Username    string  // 用户名
    DeviceId    string  // 设备ID
    Remark      string  // 设备备注
}
```

---

### 7. 刷新令牌 (RefreshToken)

**功能描述**: 使用 refresh token 刷新访问令牌，延长会话时间。

**刷新流程**:
```
1. 从上下文获取当前 token 信息
   ↓
2. 构造新的 Claims
   - 保持原有信息不变
   - 更新签发时间
   ↓
3. 生成新的 Token (有效期: config.JWT.Expire)
   ↓
4. 生成新的 Refresh Token (有效期: config.JWT.RefreshExpire)
   ↓
5. 返回新的令牌对
```

**Claims 结构**:
```go
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
```

---

### 8. 退出登录 (Logout)

**功能描述**: 清理登录状态，解绑设备和桌台。

**退出流程**:
```
1. 解绑设备
   - 更新 device 表的 finally_login_uuid 为 0
   ↓
2. 如果是平板端，解绑桌台
   - 调用 deskRepo.UnbindDesk
   ↓
3. 返回成功
```

**注意事项**:
- 收银端退出不会自动交班，需要先交班再退出
- 平板端退出会自动解绑桌台
- 助手端退出会解除与收银机的绑定

---

### 9. 修改密码 (ChangePassword)

**功能描述**: 移动管理端用户修改登录密码。

**修改流程**:
```
1. 验证来源必须是移动管理端
   ↓
2. 验证旧密码是否正确
   ↓
3. 加密新密码
   ↓
4. 更新数据库
   - password: 新密码
   - password_change_count: 计数+1
   - password_change_time: 当前时间戳
   ↓
5. 返回成功
```

**密码加密**:
```go
staff.Password = utils.EncryptPassword(changePasswordReq.NewPassword)
```

**密码修改后的影响**:
- 旧的 token 会立即失效（通过 password_change_time 验证）
- 用户需要重新登录

---

## 🔐 安全机制

### 1. 密码验证

```go
// 密码加密对比
if staff.Uuid == 0 || utils.EncryptPassword(loginReq.Password) != staff.Password {
    return loginResp, errors.New("账号或密码错误")
}
```

### 2. Token 失效机制

```go
// 修改密码后，旧 token 失效
if staff.PasswordChangeTime > auth.TokenIssuedAt {
    return errors.NewWithCode(constant.CodeTokenInvalid, "登录失效，请重新登录")
}
```

### 3. 设备绑定验证

```go
// 验证设备是否绑定
if !s.deviceSrv.IsDeviceBind(ctx, auth.CompanyUuid, auth.Source, deviceId) {
    return errors.NewWithCode(constant.CodeTokenInvalid, "设备已解绑，请重新绑定")
}
```

### 4. 验证码验证

```go
// 登录时验证验证码
if !s.captchaSrv.Verify(ctx.GetGin().GetHeader("X-SIGN"), loginReq.Code) && 
   viper.GetString("GENERAL_VERIFY_CODE") != loginReq.Code {
    return errors.New("验证码错误")
}
```

### 5. 防止重复登录

```go
// 收银端登录时检查
if staff.CashierOnline == 1 && loginReq.DeviceId != staff.BindKey {
    return errors.NewWithReplace("收银员 %s 已在其他收银机登录未交班", []string{staff.GetUserName()})
}
```

---

## 🌐 多端支持

### 部署模式

系统支持两种部署模式：

#### 云上版本 (cloud)
```go
if config.Server.DeployMode == "cloud" {
    // 通过 company_staff 表查询
    companyStaff := companyStaffRepo.GetCompanyStaff(
        companyStaffRepo.WhereUsername(loginReq.Username)
    )
    // 再根据 company_uuid 查询员工详情
    staff, _ = staffRepo.GetStaff(
        staffRepo.WhereUuid(companyStaff.Uuid)
    )
}
```

#### 离线版本 (offline/standalone)
```go
else {
    // 直接查询 staff 表
    staff, _ = staffRepo.GetStaff(
        staffRepo.WhereUsername(loginReq.Username)
    )
}
```

### 客户端类型

| 客户端类型 | 常量 | 特点 |
|-----------|------|-----|
| 收银端 | `SourceCashier` | 需要交班管理、权限最多 |
| 点餐助手端 | `SourceAssistant` | 需要绑定收银机、移动点餐 |
| 平板端 | `SourceTablet` | 需要绑定桌台、自助点餐 |
| 厨显端 | `SourceKitchen` | 厨房显示、菜品制作管理 |
| 移动管理端 | `SourceShop` | 店铺管理、数据查看 |

---

## 📊 数据流转

### 登录数据流

```
用户输入账号密码
    ↓
验证码验证 (captchaSrv)
    ↓
查询员工信息 (staffRepo)
    ↓
验证账号状态 & 商家状态
    ↓
根据来源进行特殊处理
    ↓
添加设备绑定 (deviceSrv)
    ↓
生成 JWT Token
    ↓
返回登录响应
```

### 鉴权数据流

```
客户端请求 (携带 token)
    ↓
JWT 中间件解析 token
    ↓
Auth 方法验证
    ↓
验证员工、商家、设备状态
    ↓
验证功能开关和权限
    ↓
根据来源进行特定验证
    ↓
将信息存入上下文
    ↓
继续处理业务请求
```

---

## 🚨 错误处理

### 错误码定义

| 错误码 | 说明 | 使用场景 |
|-------|------|---------|
| `CodeTokenInvalid` | Token无效 | 设备解绑、二维码失效 |
| `CodeTokenExpired` | Token过期 | 已交班、密码修改 |
| `CodeCompanyLicenceExpired` | 商家过期 | 商家授权到期 |
| `CodeCashierOrderMethodNotOpen` | 用餐方式未开启 | 收银/桌台用餐关闭 |
| `CodeUnauthorized` | 无权限 | 缺少操作权限 |
| `CodeCashierNotLogin` | 收银员未登录 | 助手端绑定失败 |
| `CodeTabletNotBindDesk` | 桌台未绑定 | 平板端未绑定桌台 |
| `CodeFunctionDisabled` | 功能未开启 | 助手/平板/厨显功能关闭 |
| `CodeCashierHandedOver` | 已交班 | 收银员已交班 |
| `CodeSystemError` | 系统错误 | 内部错误 |

### 错误处理示例

```go
// 带错误码的错误
return errors.NewWithCode(constant.CodeTokenInvalid, "设备已解绑，请重新绑定")

// 带数据的错误 (用于客户端处理)
return errors.NewWithCodeAndData(constant.CodeCashierHandedOver, cachedSubmitShift, "已交班")

// 带占位符替换的错误
return errors.NewWithReplace("收银员 %s 已在其他收银机登录", []string{staff.GetUserName()})

// 包装底层错误
return errors.WithMessage(err, "查询用户失败")
```

---

## 🔧 配置项

### JWT 配置

```go
// Token 生成
token, err = auth.GenerateToken(claims, config.JWT.Secret, config.JWT.Expire, false)
refreshToken, err = auth.GenerateToken(claims, config.JWT.Secret, config.JWT.RefreshExpire, true)
```

### 环境变量

| 变量名 | 说明 | 默认值 |
|-------|------|--------|
| `GENERAL_VERIFY_CODE` | 通用验证码（开发用） | - |
| `CHECK_SHIFT_HANDOVER` | 是否检查交班 | `true` |
| `DEPLOY_MODE` | 部署模式 | `cloud` / `offline` |

---

## 📝 最佳实践

### 1. Context 使用

```go
// ✅ 正确：从 context 获取信息
companyUuid := ctx.GetCompanyUuid()
staffUuid := ctx.GetStaffUuid()
source := ctx.GetSource()

// ❌ 错误：直接从 gin.Context 获取
companyUuid := ctx.GetGin().GetUint64("company_uuid")
```

### 2. 错误处理

```go
// ✅ 正确：使用自定义错误包装
if err != nil {
    return errors.WithMessage(err, "操作失败")
}

// ✅ 正确：返回业务错误
if staff.Uuid == 0 {
    return errors.New("用户不存在")
}

// ❌ 错误：使用 panic
panic("用户不存在")
```

### 3. 权限验证

```go
// ✅ 正确：通过 roleAccessSrv 验证权限
permissions, err := s.roleAccessSrv.GetPermission(constant.CashierRouteName, staff.Uuid, staff.CompanyUuid)
if len(permissions) == 0 {
    return errors.New("当前无权限")
}

// ✅ 正确：验证具体 API 权限
permission := constant.CashierPermissions[auth.UrlPath]
if permission != "" && !slices.Contains(permissions, permission) {
    return errors.NewWithCode(constant.CodeUnauthorized, "当前无权限")
}
```

### 4. 数据库操作

```go
// ✅ 正确：使用 dbm 获取数据库连接
db := s.dbm.GetDB(companyUuid)

// ✅ 正确：使用 repository 模式
staffRepo := repository.NewStaffRepo(db)
staff, err := staffRepo.GetStaff(
    staffRepo.WhereUuid(staffUuid),
    staffRepo.WithCompany(),
)
```

---

## 🔄 依赖关系

```
authSrv
  ├── dbm (database.DBManager)
  ├── captchaSrv (ICaptchaSrv) - 验证码验证
  ├── roleAccessSrv (IRoleAccessSrv) - 权限管理
  ├── deviceSrv (IDeviceSrv) - 设备管理
  ├── shiftSrv (IStaffShiftSrv) - 交班管理
  └── settingSrv (ISrv) - 系统设置
      ├── GetCashierSetting - 收银机设置
      ├── GetAssistantSetting - 助手设置
      ├── GetTabletSetting - 平板设置
      ├── GetKitchenSetting - 厨显设置
      ├── GetBusinessSetting - 业务设置
      ├── GetBuffetSetting - 自助餐设置
      ├── GetCurrencySetting - 货币设置
      ├── GetCloudBasicSetting - 云端设置
      ├── GetPrinterSetting - 打印机设置
      ├── GetStoreSetting - 店铺设置
      └── GetTaxRateSetting - 税率设置
```

---

## 🧪 测试建议

### 单元测试覆盖

1. **登录流程测试**
   - 正常登录流程
   - 错误的账号密码
   - 验证码错误
   - 用户被禁用
   - 商家过期
   - 重复登录检查

2. **鉴权测试**
   - Token 有效性验证
   - 权限验证
   - 设备绑定验证
   - 功能开关验证
   - 交班状态验证

3. **多端支持测试**
   - 各端登录流程
   - 各端基本信息获取
   - 各端特殊验证逻辑

4. **边界条件测试**
   - 并发登录
   - Token 过期处理
   - 设备解绑处理
   - 商家状态异常处理

---

## 📚 相关文档

- [JWT 认证机制](../auth/jwt.md)
- [权限管理系统](../permission/role_access.md)
- [设备管理](../device/device_management.md)
- [交班管理](../shift/staff_shift.md)
- [系统设置](../setting/setting_service.md)

---

## 📄 更新日志

| 日期 | 版本 | 说明 |
|-----|------|-----|
| 2025-11-12 | 1.0 | 初始文档创建 |

---

## 👥 维护者

- 开发团队：Backend Team
- 文档维护：AI Assistant

---

**注意**: 本文档基于代码自动生成，如有代码变更，请及时更新文档。

