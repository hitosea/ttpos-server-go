# Member Service 会员端服务说明文档

## 📋 概述

会员端服务是 TTPOS 系统的核心功能之一，负责处理会员注册、登录、订单管理、优惠券管理、积分管理等会员相关功能。该服务支持手机号登录、游客登录、会员等级管理、会员卡管理、积分和余额管理等功能。

**主要服务文件**:
- `/main/app/service/member.go` - 会员核心服务（838行）
- `/main/app/service/member_service/login.go` - 会员登录服务（349行）
- `/main/app/service/member_service/base.go` - 会员基础信息服务（185行）

**接口定义**: 
- `IMemberSrv` - 会员核心服务接口
- `ILoginSrv` - 会员登录服务接口
- `IBaseSrv` - 会员基础信息服务接口

---

## 🏗️ 架构设计

### 接口定义

#### IMemberSrv - 会员核心服务

```go
type IMemberSrv interface {
    // 会员信息管理
    GetLevels(companyUuid uint64) resp.MemberLevelList                                                                                                         // 获取等级列表
    GetCardTypes(companyUuid uint64) resp.MemberCardTypeList                                                                                                   // 获取会员卡类型
    SearchMember(companyUuid uint64, keyword string) resp.SearchMemberList                                                                                     // 模糊搜索会员
    AddMember(ctx context.Context, addMemberReq req.AddMemberReq) error                                                                                        // 添加会员
    GetRechargeMember(companyUuid uint64, memberUuid uint64) resp.RechargeMember                                                                               // 获取充值会员信息
    Register(ctx context.Context, reqs member_req.MemberRegisterReq) (member_resp.LoginResp, error)                                                            // 注册
    
    // 会员优惠
    GetMemberDiscount(ctx context.Context, discountReq req.GetMemberDiscountReq) (*resp.MemberDiscountResp, error)                                             // 获取会员折扣
    CheckMemberPassword(ctx context.Context, discountReq req.CheckMemberPasswordReq) error                                                                     // 使用会员优惠验证密码
    
    // 会员升级和积分余额
    HandleMemberUpgrade(companyUuid uint64, memberUuid uint64)                                                                                                 // 处理会员升级
    HandleMemberPoints(ctx context.Context, changeReq MemberPointsChangeReq) error                                                                             // 处理会员积分
    HandleMemberBalance(ctx context.Context, changeReq MemberBalanceChangeReq) error                                                                           // 处理会员余额
    
    // 优惠券和积分记录
    GetMemberCouponList(ctx context.Context, couponListReq member_req.CouponListReq) (member_resp.CouponListWithPaginationResp, error)                         // 获取优惠券列表
    GetMemberPointsRecordList(ctx context.Context, pointsRecordListReq member_req.PointsRecordListReq) (member_resp.PointsRecordListWithPaginationResp, error) // 获取积分记录列表
    
    // 工具方法
    GenerateRandomNickname() string                                                                                                                            // 生成随机昵称
}
```

#### ILoginSrv - 会员登录服务

```go
type ILoginSrv interface {
    GetLoginInfo(ctx context.Context, req member_req.MemberLoginInfoReq) (member_resp.MemberLoginInfoResp, error) // 获取登录信息
    SendCode(ctx context.Context, req member_req.MemberSendCodeReq) error                                         // 发送验证码
    SendRegisterCode(ctx context.Context, req member_req.MemberSendCodeReq) error                                 // 发送注册验证码
    Login(ctx context.Context, req member_req.MemberLoginReq) (member_resp.LoginResp, error)                      // 登录
    VisitorLogin(ctx context.Context, loginReq req.VisitorLoginReq) (*member_resp.LoginResp, error)               // 游客登录
}
```

#### IBaseSrv - 会员基础信息服务

```go
type IBaseSrv interface {
    GetBaseInfo(ctx context.Context) (member_resp.MemberBaseInfoResp, error)                                            // 获取基础信息
    UpdateNickname(ctx context.Context, req member_req.MemberNicknameUpdateReq) (member_resp.MemberBaseInfoResp, error) // 修改会员昵称
}
```

### 依赖服务

```go
type memberSrv struct {
    dbm   *database.DBManager  // 数据库管理器
    bus   *event.SystemEventBus // 事件总线
    cache cache.Cache          // 缓存
}

type loginSrv struct {
    dbm        *database.DBManager  // 数据库管理器
    cache      cache.Cache          // 缓存
    smsSrv     service.ISmsSrv      // 短信服务
    settingSrv setting.ISrv        // 设置服务
}

type baseSrv struct {
    dbm   *database.DBManager  // 数据库管理器
    cache cache.Cache          // 缓存
}
```

---

## 🎯 核心功能

### 1. 会员登录和注册

#### 获取登录信息 (GetLoginInfo)

**功能描述**: 获取登录前的信息，包括商家信息、区号列表、语言列表等。

**处理流程**:
```
1. 获取商家信息
   ↓
2. 根据IP地址确定国家代码
   ↓
3. 根据国家代码调整区号列表顺序
   ↓
4. 获取语言列表
   ↓
5. 返回登录信息
```

**区号映射**:
- `TH` (泰国) → `+66`
- `CN` (中国) → `+86`
- `HK` (香港) → `+86`

#### 发送验证码 (SendCode / SendRegisterCode)

**功能描述**: 发送登录或注册验证码短信。

**前置条件**:
- 商家会员服务已开启
- 商家未过期且未删除
- 登录验证码：手机号必须已注册
- 注册验证码：手机号必须未注册

**处理流程**:
```
1. 验证商家状态
   ↓
2. 验证手机号状态（登录/注册）
   ↓
3. 生成验证码并缓存
   ↓
4. 发送短信验证码
```

#### 登录 (Login)

**功能描述**: 会员手机号验证码登录。

**处理流程**:
```
1. 验证验证码
   ↓
2. 验证商家状态
   ↓
3. 查询会员信息
   ↓
4. 生成JWT Token和Refresh Token
   ↓
5. 返回登录信息
```

**返回结构**:
```go
type LoginResp struct {
    Token        string `json:"token"`         // 访问令牌
    RefreshToken string `json:"refresh_token"` // 刷新令牌
}
```

#### 游客登录 (VisitorLogin)

**功能描述**: 游客登录，如果游客不存在则自动创建。

**处理流程**:
```
1. 根据设备ID查询游客
   ↓
2. 如果不存在，创建游客
   - 生成随机昵称（9位Base62编码）
   - 创建会员记录（IsVisitor=true）
   - 创建SaaS会员记录
   ↓
3. 生成JWT Token（有效期100年）
   ↓
4. 返回登录信息
```

**随机昵称生成规则**:
- 使用时间戳后8位 + 4位随机数 + 3位进程标识
- Base62编码（0-9, a-z, A-Z）
- 确保长度为9位
- 检查唯一性，如果重复则重试

#### 注册 (Register)

**功能描述**: 会员注册，支持游客转正式会员。

**处理流程**:
```
1. 验证商家状态
   ↓
2. 验证验证码
   ↓
3. 验证手机号是否已注册
   ↓
4. 验证推荐人（如果提供）
   ↓
5. 判断是否为游客
   - 是游客：更新游客信息为正式会员
   - 不是游客：创建新会员
   ↓
6. 生成JWT Token
   ↓
7. 返回登录信息
```

### 2. 会员基础信息

#### 获取基础信息 (GetBaseInfo)

**功能描述**: 获取会员基础信息，包括用户信息、会员设置、商家信息、货币设置等。

**返回结构**:
```go
type MemberBaseInfoResp struct {
    User     UserResp     `json:"user"`     // 用户信息
    Member   MemberResp   `json:"member"`   // 会员设置
    Company  CompanyResp  `json:"company"`  // 商家信息
    Currency CurrencyResp `json:"currency"` // 货币设置
}
```

**用户信息**:
- ID、UUID、昵称、手机号
- 积分、余额
- 是否游客

**会员设置**:
- 是否显示售罄商品
- 语言列表和默认语言
- 是否开启骑手配送
- 区号列表

#### 修改昵称 (UpdateNickname)

**功能描述**: 修改会员昵称。

**处理流程**:
```
1. 验证请求参数
   ↓
2. 更新会员昵称
   ↓
3. 更新上下文缓存
   ↓
4. 返回更新后的基础信息
```

### 3. 会员管理

#### 添加会员 (AddMember)

**功能描述**: 管理员添加会员，支持发卡、赠送积分和余额。

**处理流程**:
```
1. 验证手机号是否已存在
   ↓
2. 验证会员等级是否存在
   ↓
3. 验证会员卡类型（如果提供）
   ↓
4. 验证推荐人（如果提供）
   ↓
5. 处理会员密码（MD5加密）
   ↓
6. 验证卡号唯一性
   ↓
7. 创建会员记录
   ↓
8. 如果发卡，创建会员卡记录
   ↓
9. 如果开卡赠送积分/余额，处理赠送
   ↓
10. 发布事件（积分/余额变动）
```

**开卡赠送**:
- 开卡赠送积分：`OpenPoint == 1 && OpenPointNum > 0`
- 开卡赠送余额：`OpenMoney == 1 && OpenMoneyNum > 0`

#### 搜索会员 (SearchMember)

**功能描述**: 根据关键词模糊搜索会员（手机号、昵称、会员号）。

**处理流程**:
```
1. 验证关键词不为空
   ↓
2. 数据库模糊查询
   ↓
3. 返回搜索结果
```

### 4. 会员等级和升级

#### 获取等级列表 (GetLevels)

**功能描述**: 获取所有会员等级列表。

**返回结构**:
```go
type MemberLevelList struct {
    List []MemberLevel `json:"list"` // 等级列表
}
```

#### 处理会员升级 (HandleMemberUpgrade)

**功能描述**: 根据会员消费金额和积分自动升级会员等级。

**升级条件**:
- 按消费金额升级：`AccumulatedConsumptionAmount >= UpgradeMoney`
- 按积分升级：`AccumulatedConsumptionGetPoint >= UpgradePoint`
- 同时满足：两个条件都满足
- 只能升级，不能降级（优先级判断）

**处理流程**:
```
1. 获取会员当前等级
   ↓
2. 获取所有会员等级（按优先级排序）
   ↓
3. 遍历等级，找到可升级的最高等级
   ↓
4. 如果找到可升级等级，更新会员等级
   ↓
5. 记录等级变动日志
```

**等级变动日志类型**:
- `MemberLevelLogTypeAutoUpgrade` - 自动升级

### 5. 会员积分管理

#### 处理会员积分 (HandleMemberPoints)

**功能描述**: 处理会员积分变动（增加或减少）。

**参数结构**:
```go
type MemberPointsChangeReq struct {
    Uuid     uint64  // 会员UUID
    Points   float64 // 积分变动值（正数增加，负数减少）
    Scene    int     // 场景（积分日志场景）
    Describe string  // 描述
}
```

**处理流程**:
```
1. 验证会员是否存在
   ↓
2. 更新会员冻结积分（FrozenPoint）
   ↓
3. 创建积分变动日志
```

**积分场景常量**:
- `MemberPointLogSceneCashierOrAssistant` - 收银机/点餐助手
- `MemberPointLogSceneOrder` - 订单
- `MemberPointLogSceneRecharge` - 充值
- `MemberPointLogSceneRefund` - 退款

#### 获取积分记录列表 (GetMemberPointsRecordList)

**功能描述**: 获取会员积分变动记录列表。

**查询参数**:
- `Type = 1`: 只查询增加记录（正数）
- `Type = 2`: 只查询减少记录（负数）
- `Type = 0`: 查询所有记录

**返回结构**:
```go
type PointsRecordListWithPaginationResp struct {
    List       []PointsRecord `json:"list"`        // 积分记录列表
    TotalPoint float64        `json:"total_point"` // 总积分
    Meta       PageResponse   `json:"meta"`       // 分页信息
}
```

### 6. 会员余额管理

#### 处理会员余额 (HandleMemberBalance)

**功能描述**: 处理会员余额变动（充值、消费、退款等）。

**参数结构**:
```go
type MemberBalanceChangeReq struct {
    MemberUuid  uint64  // 会员UUID
    Money       float64 // 变动金额（正数增加，负数减少）
    GiftMoney   float64 // 变动赠送金额（正数增加，负数减少）
    Scene       int     // 场景
    Describe    string  // 描述
    RelatedUuid uint64  // 关联UUID（如退款单ID、订单ID）
}
```

**处理流程**:
```
1. 验证会员是否存在
   ↓
2. 更新会员冻结余额和冻结赠送余额
   ↓
3. 创建余额变动日志
```

**余额场景常量**:
- `MemberBalanceLogCashierOrAssistant` - 收银机/点餐助手
- `MemberBalanceLogOrder` - 订单
- `MemberBalanceLogRecharge` - 充值
- `MemberBalanceLogRefund` - 退款

### 7. 会员优惠

#### 获取会员折扣 (GetMemberDiscount)

**功能描述**: 获取会员在订单中的折扣金额。

**处理流程**:
```
1. 获取会员信息（包含会员卡、会员等级）
   ↓
2. 获取销售账单信息
   ↓
3. 设置会员折扣并重新计算订单金额
   ↓
4. 返回会员信息和折扣金额
```

**返回结构**:
```go
type MemberDiscountResp struct {
    Member          RechargeMember `json:"member"`           // 会员信息
    HasPassword     bool           `json:"has_password"`     // 是否有密码
    DiscountedPrice float64        `json:"discounted_price"` // 折扣后价格
}
```

#### 验证会员密码 (CheckMemberPassword)

**功能描述**: 使用会员优惠时验证会员密码。

**处理流程**:
```
1. 获取会员信息
   ↓
2. 如果会员有密码，验证密码（MD5）
   ↓
3. 获取销售账单信息
   ↓
4. 重新计算订单金额（应用会员折扣）
   ↓
5. 更新订单商品和订单记录
```

### 8. 优惠券管理

#### 获取优惠券列表 (GetMemberCouponList)

**功能描述**: 获取会员的优惠券列表（有效券或历史券）。

**查询参数**:
- `IsHistory = 0`: 查询有效优惠券
- `IsHistory = 1`: 查询历史优惠券（已使用/已过期）

**处理流程**:
```
1. 根据IsHistory查询优惠券
   ↓
2. 过滤已关闭的营销活动优惠券
   ↓
3. 计算优惠券状态
   ↓
4. 判断适用时间段
   ↓
5. 返回优惠券列表
```

**优惠券状态**:
- 未使用
- 已使用
- 已过期

### 9. 会员订单管理

会员端订单管理功能包括：
- 创建订单
- 获取订单表单信息
- 设置订单地址
- 提交支付
- 获取订单列表
- 获取订单详情
- 取消订单

详细功能请参考订单服务文档。

---

## 🔄 数据流转

### 会员注册流程

```
获取登录信息
    ↓
发送注册验证码
    ↓
输入验证码和手机号
    ↓
验证验证码
    ↓
创建会员记录
    ↓
生成Token
    ↓
返回登录信息
```

### 会员登录流程

```
获取登录信息
    ↓
发送登录验证码
    ↓
输入验证码
    ↓
验证验证码
    ↓
查询会员信息
    ↓
生成Token
    ↓
返回登录信息
```

### 会员升级流程

```
会员消费/充值
    ↓
积分/余额变动
    ↓
触发升级检查
    ↓
计算是否满足升级条件
    ↓
更新会员等级
    ↓
记录等级变动日志
```

### 会员积分/余额变动流程

```
业务触发（订单、充值等）
    ↓
调用积分/余额处理接口
    ↓
更新会员冻结积分/余额
    ↓
创建变动日志
    ↓
发布事件（异步）
    ↓
触发升级检查（积分变动）
```

---

## 🔐 权限控制

### 公开接口（无需认证）

- 获取登录信息
- 发送验证码
- 登录
- 游客登录

### 需要认证的接口

- 注册（需要游客Token）
- 发送注册验证码（需要游客Token）
- 获取基础信息
- 修改昵称
- 创建订单
- 获取优惠券列表
- 获取积分记录

### 会员状态验证

- 已注销会员不能登录
- 已注销会员不能注册
- 游客可以转为正式会员

---

## ⚠️ 错误处理

### 常见错误

| 错误信息 | 原因 | 解决方案 |
|---------|------|---------|
| "商家不存在" | 商家UUID错误或商家已删除 | 检查商家UUID |
| "商家会员服务已关闭" | 商家未开启会员功能 | 联系商家开启会员服务 |
| "无法使用该功能，请联系商家" | 商家已过期或已删除 | 联系商家续费或恢复 |
| "该手机号未在该商家进行注册" | 手机号未注册 | 先注册会员 |
| "该手机号已注册本商家会员" | 手机号已注册 | 直接登录 |
| "该会员已被注销，可联系商家处理" | 会员已被注销 | 联系商家恢复 |
| "验证码错误" | 验证码不匹配 | 重新获取验证码 |
| "密码错误" | 会员密码不正确 | 检查密码 |
| "会员等级不存在" | 等级UUID错误 | 检查等级UUID |
| "会员卡不存在" | 卡类型UUID错误 | 检查卡类型UUID |
| "卡号重复" | 卡号已存在 | 使用其他卡号 |
| "推荐人不存在" | 推荐人UUID错误 | 检查推荐人UUID |

### 错误处理机制

1. **参数验证**: 所有接口都进行严格的参数验证
2. **业务验证**: 验证商家状态、会员状态、验证码等
3. **事务处理**: 关键操作使用数据库事务保证一致性
4. **错误包装**: 使用 `errors.WithMessage` 包装错误信息
5. **日志记录**: 关键错误记录到日志系统

---

## 📊 数据模型

### Member - 会员

```go
type Member struct {
    BaseModel
    MemberNo                      string  // 会员号（5位数字）
    Nickname                      string  // 昵称
    Phone                         string  // 手机号
    Password                      string  // 密码（MD5加密）
    MemberLevelUuid              uint64  // 会员等级UUID
    Gender                        int     // 性别
    MemberCardNo                 string  // 会员卡号
    MemberCardUuid               uint64  // 会员卡UUID
    ReferrerUuid                 uint64  // 推荐人UUID
    ActivityUuid                 uint64  // 活动UUID
    IsVisitor                    bool    // 是否游客
    DeviceId                     string  // 设备ID
    FrozenBalance                float64 // 冻结余额
    FrozenGiftBalance            float64 // 冻结赠送余额
    FrozenPoint                  float64 // 冻结积分
    AccumulatedConsumptionAmount float64 // 累计消费金额
    AccumulatedConsumptionGetPoint float64 // 累计消费获得积分
}
```

### MemberLevel - 会员等级

```go
type MemberLevel struct {
    BaseModel
    Name         string  // 等级名称
    Priority     int     // 优先级（数字越大等级越高）
    IsDefault    int     // 是否默认等级
    UpgradeMoney float64 // 升级所需消费金额
    UpgradePoint float64 // 升级所需积分
    OpenMoney    int     // 是否开启按金额升级
    OpenPoint    int     // 是否开启按积分升级
}
```

### MemberCard - 会员卡

```go
type MemberCard struct {
    BaseModel
    CardTypeUuid uint64  // 卡类型UUID
    MemberUuid   uint64  // 会员UUID
    ExpireTime   int64   // 过期时间
    Discount     float64 // 折扣
}
```

### MemberCardType - 会员卡类型

```go
type MemberCardType struct {
    BaseModel
    Name          string  // 卡类型名称
    Price         float64 // 卡价格
    Discount      float64 // 折扣
    Expire        int     // 有效期（月）
    OpenPoint     int     // 是否开卡送积分
    OpenPointNum  float64 // 开卡送积分数量
    OpenMoney     int     // 是否开卡送余额
    OpenMoneyNum  float64 // 开卡送余额数量
}
```

### MemberPointLog - 积分日志

```go
type MemberPointLog struct {
    BaseModel
    MemberUuid uint64  // 会员UUID
    Scene      int     // 场景
    Value      float64 // 积分值（正数增加，负数减少）
    Describe   string  // 描述
}
```

### MemberBalanceLog - 余额日志

```go
type MemberBalanceLog struct {
    BaseModel
    MemberUuid  uint64  // 会员UUID
    Scene       int     // 场景
    Money       float64 // 金额（余额+赠送余额）
    GiftMoney   float64 // 赠送金额
    Describe    string  // 描述
    RelatedUuid uint64  // 关联UUID
}
```

### MemberCoupon - 会员优惠券

```go
type MemberCoupon struct {
    BaseModel
    MemberUuid           uint64  // 会员UUID
    MarketingCouponUuid  uint64  // 营销优惠券UUID
    DayStartTime         string  // 每日开始时间
    DayEndTime           string  // 每日结束时间
    UseTime              int64   // 使用时间
    ExpireTime           int64   // 过期时间
}
```

---

## 🚀 性能优化

### 缓存策略

1. **会员信息缓存**:
   - 会员登录后，会员信息缓存在JWT Token中
   - 减少数据库查询

2. **验证码缓存**:
   - 验证码存储在Redis中
   - 有效期5分钟
   - 防止重复发送

3. **商家信息缓存**:
   - 商家设置信息缓存
   - 减少重复查询

### 数据库优化

1. **索引优化**:
   - 手机号索引
   - 会员号索引
   - 设备ID索引（游客查询）

2. **查询优化**:
   - 使用预加载减少N+1查询
   - 分页查询优化

### 并发控制

- **会员注册**: 使用数据库唯一索引防止重复注册
- **积分/余额变动**: 使用冻结字段，避免并发问题

---

## 🧪 测试建议

### 单元测试

1. **登录注册测试**:
   - 测试验证码生成和验证
   - 测试手机号注册和登录
   - 测试游客登录和转正式会员
   - 测试错误场景（验证码错误、手机号已注册等）

2. **会员管理测试**:
   - 测试添加会员
   - 测试会员升级
   - 测试积分和余额变动
   - 测试会员搜索

3. **优惠券测试**:
   - 测试优惠券查询
   - 测试优惠券状态计算
   - 测试适用时间段判断

### 集成测试

1. **完整流程测试**:
   - 测试注册-登录-下单完整流程
   - 测试积分/余额变动-升级流程
   - 测试会员优惠使用流程

2. **并发测试**:
   - 测试并发注册
   - 测试并发积分/余额变动
   - 测试并发查询

### 性能测试

1. **查询性能**:
   - 测试会员搜索性能
   - 测试优惠券列表查询性能
   - 测试积分记录查询性能

2. **写入性能**:
   - 测试会员注册性能
   - 测试积分/余额变动性能

---

## 📝 注意事项

1. **会员密码**:
   - 密码使用MD5加密存储
   - 验证时需要对输入密码进行MD5加密后比较

2. **游客会员**:
   - 游客会员的Token有效期很长（100年）
   - 游客可以转为正式会员
   - 游客会员没有手机号

3. **会员升级**:
   - 只能升级，不能降级
   - 升级条件可以按金额、按积分或同时满足
   - 升级是自动触发的

4. **积分和余额**:
   - 使用冻结字段（FrozenPoint、FrozenBalance）避免并发问题
   - 每次变动都会记录日志
   - 积分变动会触发升级检查

5. **会员卡**:
   - 开卡可以赠送积分和余额
   - 会员卡有折扣功能
   - 会员卡有过期时间

6. **验证码**:
   - 验证码有效期5分钟
   - 验证码存储在Redis中
   - 登录和注册使用不同的验证码接口

7. **区号识别**:
   - 根据IP地址自动识别国家
   - 自动调整区号列表顺序
   - 支持泰国和中国区号

---

## 🔗 相关文档

- [订单服务文档](./order.md) - 会员订单相关功能
- [充值订单文档](./recharge_order.md) - 会员充值相关功能
- [设置服务文档](./setting.md) - 会员设置相关配置
- [认证服务文档](./auth.md) - JWT Token认证相关

---

**文档版本**: v1.0  
**最后更新**: 2025-01-27  
**维护者**: TTPOS开发团队

