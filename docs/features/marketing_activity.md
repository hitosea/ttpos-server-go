# Marketing Activity Service 营销活动服务说明文档

## 📋 概述

营销活动服务是 TTPOS 系统的核心功能之一，负责处理营销活动的创建、管理、查询、二维码生成等功能。该服务支持邀请有礼、积分商城等多种活动类型，支持优惠券和积分两种奖励方式，并提供活动二维码生成和解密功能。

**主要服务文件**:
- `/main/app/service/marketing_activity.go` - 营销活动服务（368行）
- `/main/app/repository/marketing_activity.go` - 营销活动仓库（188行）
- `/main/app/model/marketing.go` - 营销活动模型（165行）

**接口定义**: 
- `IMarketingActivitySrv` - 营销活动服务接口
- `IMarketingActivityRepo` - 营销活动仓库接口

---

## 🏗️ 架构设计

### 接口定义

#### IMarketingActivitySrv - 营销活动服务

```go
type IMarketingActivitySrv interface {
    MarketingActivity(ctx context.Context, req member_req.MarketingActivityReq) (*member_resp.MemberMarketingActivityListResp, error)               // 获取营销活动（旧接口）
    MarketingActivityDetail(ctx context.Context, req member_req.MarketingActivityDetailReq) (*member_resp.MemberMarketingActivityDetailResp, error) // 获取营销活动详情
    MarketingActivityList(ctx context.Context) (*member_resp.MemberMarketingActivityListsResp, error)                                               // 获取营销活动列表
    DecryptQrCode(ctx context.Context, req req.DecryptQrCodeReq) (*resp.DecryptQrCodeResp, error)                                                   // 解密活动二维码
}
```

#### IMarketingActivityRepo - 营销活动仓库

```go
type IMarketingActivityRepo interface {
    GetActivity(uuid uint64) (*model.MarketingActivity, error)                            // 获取营销活动
    GetActivityAndPrizes(uuid uint64) (*model.MarketingActivity, error)                   // 获取营销活动与奖励
    GetMemberClientActivityList(dbOption ...DBOption) ([]*model.MarketingActivity, error) // 获取会员端营销活动列表
    GetValidActivityListByNow(dbOption ...DBOption) ([]*model.MarketingActivity, error)   // 获取正在进行中的营销活动列表
    GetValidActivity() (*model.MarketingActivity, error)                                  // 获取正在进行中的营销活动
    GetValidActivityByUuid(uuid uint64) (*model.MarketingActivity, error)                 // 根据uuid获取正在进行中的营销活动
    GenerateQrCode(params *QrCodeParams) (string, error)                                  // 生成二维码
    
    WithLanguage() DBOption  // 预加载多语言
    WithPrizes() DBOption    // 预加载奖品
}
```

### 依赖服务

```go
type marketingActivitySrv struct {
    dbm   *database.DBManager  // 数据库管理器
    cache cache.Cache          // 缓存
}
```

---

## 🎯 核心功能

### 1. 活动类型

系统支持多种营销活动类型：

| 活动类型常量 | 说明 | 特点 |
|------------|------|-----|
| `Type = 0` | 邀请有礼 | 会员邀请新用户注册，推荐人获得奖励 |
| `Type = 1` | 积分商城 | 会员使用积分兑换奖励 |

### 2. 奖励类型

系统支持两种奖励方式：

| 奖励类型常量 | 说明 | 特点 |
|------------|------|-----|
| `RewardType = 0` | 优惠券奖励 | 奖励优惠券，可以设置多个优惠券奖品 |
| `RewardType = 1` | 积分奖励 | 奖励固定积分数量 |

### 3. 活动状态

活动状态由系统自动计算：

| 状态值 | 说明 | 判断条件 |
|-------|------|---------|
| `Status = 0` | 未开始 | `当前时间 < 开始时间` |
| `Status = 1` | 进行中 | `开始时间 <= 当前时间 <= 结束时间 && 未失效` |
| `Status = 2` | 已结束/已失效 | `当前时间 > 结束时间 || 已失效` |

**状态判断逻辑**:
```go
func (m *MarketingActivity) GetStatus() int {
    if m.IsInvalid == 1 {
        return 2  // 已失效
    }
    now := time.Now().Unix()
    if int64(m.StartTime) > now {
        return 0  // 未开始
    } else if int64(m.EndTime) < now {
        return 2  // 已结束
    } else {
        return 1  // 进行中
    }
}
```

### 4. 获取营销活动列表 (MarketingActivityList)

**功能描述**: 获取会员端可见的营销活动列表，包括进行中、未开始、已结束的活动。

**处理流程**:
```
1. 查询活动列表（最近7天内的活动）
   - 开始时间 <= 当前时间
   - 结束时间 >= 7天前
   ↓
2. 按状态排序
   - 进行中(1) 排最前
   - 未开始(0) 其次
   - 已结束/无效(2) 最后
   ↓
3. 加载多语言和奖品信息
   ↓
4. 构建响应数据
   - 优惠券奖励：返回优惠券列表
   - 积分奖励：返回积分信息
   ↓
5. 返回活动列表
```

**查询条件**:
- 查询最近7天内的活动（`end_time >= 7天前`）
- 按活动状态和开始时间排序
- 预加载多语言名称和描述
- 预加载奖品信息（优惠券）

**返回结构**:
```go
type MemberMarketingActivityListsResp struct {
    List []MemberMarketingActivityInfoResp `json:"list"` // 活动列表
}

type MemberMarketingActivityInfoResp struct {
    Uuid       uint64                             `json:"uuid"`        // 活动UUID
    Type       int                                `json:"type"`        // 活动类型
    LocaleName dto.LocaleResponse                 `json:"locale_name"` // 活动名称（多语言）
    LocaleDesc dto.LocaleResponse                 `json:"locale_desc"` // 活动描述（多语言）
    StartTime  int64                              `json:"start_time"`  // 活动开始时间
    EndTime    int64                              `json:"end_time"`    // 活动结束时间
    Status     int                                `json:"status"`      // 活动状态
    Prizes     []MemberMarketingActivityPrizeResp `json:"prizes"`      // 奖品列表
}
```

### 5. 获取营销活动详情 (MarketingActivityDetail)

**功能描述**: 获取指定营销活动的详细信息，包括活动信息、会员信息、商家信息、二维码等。

**处理流程**:
```
1. 获取正在进行中的营销活动
   ↓
2. 如果活动不存在，返回默认信息
   ↓
3. 生成活动二维码
   ↓
4. 构建响应数据
   - 活动信息（多语言）
   - 会员信息
   - 商家信息
   ↓
5. 验证商家营销功能是否开启
   ↓
6. 返回活动详情
```

**返回结构**:
```go
type MemberMarketingActivityDetailResp struct {
    Activity   MemberMarketingActivityResp `json:"activity"`   // 活动信息
    MemberInfo MemberInfoResp              `json:"member_info"` // 会员信息
    Company    CompanyInfoResp             `json:"company"`    // 商家信息
}
```

**特殊情况处理**:
- 如果活动不存在，返回"暂无营销活动"的默认信息
- 如果商家营销功能未开启，返回错误但包含活动信息

### 6. 获取营销活动（旧接口）(MarketingActivity)

**功能描述**: 获取营销活动（旧接口，已废弃但保留兼容）。

**处理流程**:
```
1. 查询正在进行中的活动列表
   ↓
2. 为每个活动生成二维码
   ↓
3. 构建响应数据
   ↓
4. 如果无活动，返回默认信息
   ↓
5. 验证商家营销功能
   ↓
6. 返回活动列表
```

**与MarketingActivityList的区别**:
- 只返回正在进行中的活动
- 包含会员信息和商家信息
- 为每个活动生成二维码

### 7. 二维码生成和解密

#### 生成二维码 (GenerateQrCode)

**功能描述**: 为营销活动生成二维码，二维码包含活动信息、会员信息、商家信息。

**处理流程**:
```
1. 构建二维码参数
   - Type: 活动类型
   - CompanyUuid: 商家UUID
   - MemberUuid: 会员UUID
   - ActivityUuid: 活动UUID
   ↓
2. 将参数转换为JSON
   ↓
3. 使用AES加密JSON数据
   ↓
4. 生成二维码图片
   ↓
5. 转换为Base64编码
   ↓
6. 返回Data URL格式的二维码
```

**二维码参数结构**:
```go
type QrCodeParams struct {
    Type         uint64 `json:"t"`       // 活动类型
    CompanyUuid  uint64 `json:"c_uuid"`  // 公司UUID
    MemberUuid   uint64 `json:"m_uuid"`  // 会员UUID
    ActivityUuid uint64 `json:"mc_uuid"` // 营销活动UUID
}
```

**二维码格式**:
- 加密方式: AES加密
- 图片格式: PNG
- 返回格式: `data:image/png;base64,{base64_string}`

#### 解密二维码 (DecryptQrCode)

**功能描述**: 解密活动二维码，获取二维码中包含的活动和会员信息。

**处理流程**:
```
1. 使用AES解密二维码数据
   ↓
2. 解析JSON数据
   ↓
3. 验证活动是否存在
   ↓
4. 验证会员是否存在
   ↓
5. 返回解密后的信息
```

**返回结构**:
```go
type DecryptQrCodeResp struct {
    Uuid         uint64 `json:"uuid"`          // 会员UUID
    Nickname     string `json:"nickname"`       // 会员昵称
    Phone        string `json:"phone"`          // 会员手机号
    ActivityUuid uint64 `json:"activity_uuid"` // 活动UUID
}
```

### 8. 活动奖励机制

#### 奖励条件

活动可以设置奖励条件：

- **奖励条件金额** (`RewardConditionAmount`): 满足消费金额条件才能获得奖励
- **奖励次数限制** (`IsOpenRewardLimit`): 是否开启奖励次数限制
- **奖励次数限制值** (`RewardLimit`): 每个会员最多可获得奖励的次数

#### 奖励发放

**优惠券奖励**:
- 活动可以设置多个优惠券奖品
- 每个奖品对应一个优惠券
- 会员获得奖励后，优惠券会发放到会员账户

**积分奖励**:
- 活动设置固定的积分奖励值
- 会员获得奖励后，积分直接增加到会员账户

### 9. 活动记录管理

#### 活动记录 (MarketingActivityRecord)

**功能描述**: 记录会员参与活动的信息。

**记录字段**:
- `ActivityUuid`: 活动UUID
- `PrizeUuid`: 奖品UUID
- `MemberUuid`: 会员UUID
- `RewardCount`: 已获得奖励次数
- `RewardValue`: 奖励值
- `LastRewardTime`: 最后一次获得奖励时间

**用途**:
- 统计会员参与活动的次数
- 判断是否达到奖励次数限制
- 记录奖励发放历史

#### 消费记录 (MarketingActivityConsumption)

**功能描述**: 记录邀请有礼活动中的消费记录。

**记录字段**:
- `ActivityUuid`: 活动UUID
- `ReferrerUuid`: 推荐人UUID
- `ConsumerUuid`: 消费人UUID
- `ConsumptionAmount`: 消费金额
- `RewardAmount`: 奖励金额
- `RewardStatus`: 奖励状态（0=未发放, 1=已发放）

**用途**:
- 记录被推荐人的消费情况
- 计算推荐人应获得的奖励
- 跟踪奖励发放状态

---

## 🔄 数据流转

### 活动创建流程

```
商家创建活动
    ↓
设置活动基本信息
    ↓
设置活动时间
    ↓
设置奖励类型和奖品
    ↓
设置奖励条件
    ↓
保存活动
    ↓
活动生效（到达开始时间）
```

### 会员参与活动流程

```
会员查看活动列表
    ↓
选择活动查看详情
    ↓
获取活动二维码
    ↓
分享二维码给新用户
    ↓
新用户扫描二维码注册
    ↓
新用户消费达到条件
    ↓
系统发放奖励给推荐人
    ↓
记录活动记录和消费记录
```

### 二维码使用流程

```
会员获取活动二维码
    ↓
分享二维码
    ↓
新用户扫描二维码
    ↓
解密二维码获取活动信息
    ↓
验证活动有效性
    ↓
引导新用户注册
    ↓
记录推荐关系
```

---

## 🔐 权限控制

### 公开接口（无需认证）

- 无（所有接口都需要会员认证）

### 需要认证的接口

- 获取营销活动列表（需要会员Token）
- 获取营销活动详情（需要会员Token）
- 获取营销活动（旧接口，需要会员Token）
- 解密活动二维码（需要会员Token）

### 商家功能验证

- 商家必须开启会员功能（`IsOpenMember == 1`）
- 商家必须开启营销活动功能（`IsOpenMarketing == 1`）
- 如果功能未开启，返回错误提示

---

## ⚠️ 错误处理

### 常见错误

| 错误信息 | 原因 | 解决方案 |
|---------|------|---------|
| "营销活动已失效" | 活动已结束或商家营销功能未开启 | 检查活动状态和商家设置 |
| "活动二维码无效" | 二维码数据损坏或活动不存在 | 重新生成二维码 |
| "商家不存在" | 商家UUID错误或商家已删除 | 检查商家UUID |
| "商家会员服务已关闭" | 商家未开启会员功能 | 联系商家开启会员服务 |
| "商家营销活动已结束" | 商家营销功能未开启 | 联系商家开启营销功能 |

### 错误处理机制

1. **参数验证**: 所有接口都进行严格的参数验证
2. **业务验证**: 验证活动状态、商家状态等
3. **错误包装**: 使用 `errors.WithMessage` 包装错误信息
4. **特殊处理**: 活动不存在时返回默认信息而不是错误

---

## 📊 数据模型

### MarketingActivity - 营销活动

```go
type MarketingActivity struct {
    BaseModel
    Name                  string  // 活动名称
    Type                  int     // 活动类型（0=邀请有礼, 1=积分商城）
    MultiLanguageNameUuid uint64  // 活动名称多语言UUID
    Description           string  // 活动文案
    MultiLanguageDescUuid uint64  // 活动文案多语言UUID
    StartTime             int     // 活动开始时间
    EndTime               int     // 活动结束时间
    RewardType            int     // 奖励类型（0=优惠券, 1=积分）
    RewardValue           float64 // 奖励值
    IsSendSms             int     // 是否发送短信通知
    RewardConditionAmount float64 // 奖励条件金额
    IsOpenRewardLimit     int     // 是否开启奖励次数限制
    RewardLimit           int64   // 奖励次数限制
    IsInvalid             int     // 是否失效
    ImageBase64           string  // 活动图片Base64
    
    // 关联模型
    Prizes            []*MarketingActivityPrize  // 奖品列表
    Records           []*MarketingActivityRecord // 活动记录列表
    MultiLanguageName *MultiLanguageName         // 多语言名称
    MultiLanguageDesc *MultiLanguageName         // 多语言描述
}
```

### MarketingActivityPrize - 活动奖品

```go
type MarketingActivityPrize struct {
    BaseModel
    ActivityUuid uint64           // 活动UUID
    PrizeType    int              // 奖品类型（1=优惠券）
    PrizeUuid    uint64           // 奖品UUID
    Coupon       *MarketingCoupon // 优惠券
}
```

### MarketingActivityRecord - 活动记录

```go
type MarketingActivityRecord struct {
    BaseModel
    ActivityUuid   uint64  // 活动UUID
    PrizeUuid      uint64  // 奖品UUID
    MemberUuid     uint64  // 会员UUID
    RewardCount    int     // 已获得奖励次数
    RewardValue    float64 // 奖励值
    LastRewardTime int64   // 最后一次获得奖励时间
}
```

### MarketingActivityConsumption - 消费记录

```go
type MarketingActivityConsumption struct {
    BaseModel
    ActivityUuid      uint64  // 活动UUID
    ReferrerUuid      uint64  // 推荐人UUID
    ConsumerUuid      uint64  // 消费人UUID
    ConsumptionAmount float64 // 消费金额
    RewardAmount      float64 // 奖励金额
    RewardStatus      int     // 奖励状态（0=未发放, 1=已发放）
}
```

### MarketingCoupon - 营销优惠券

```go
type MarketingCoupon struct {
    BaseModel
    Name           string  // 优惠券名称
    Sort           int     // 排序
    Type           string  // 优惠券类型（deduction=抵扣券）
    DeductionType  string  // 抵扣类型（taxed=税后抵扣）
    Amount         float64 // 优惠券金额
    Count          int     // 优惠券数量
    DayStartTime   string  // 每日适用时段开始时间（HH:mm）
    DayEndTime     string  // 每日适用时段结束时间（HH:mm）
    Requirement    string  // 获得条件（none=都可以, marketing=营销活动）
    ValidStartTime int     // 有效开始时间（requirement=none时有效）
    ValidEndTime   int     // 有效结束时间（requirement=none时有效）
    ValidDays      int     // 领取后n天内有效（requirement=marketing时有效）
    Status         int     // 状态（0=禁用, 1=开启）
}
```

### MarketingCouponRecord - 优惠券记录

```go
type MarketingCouponRecord struct {
    BaseModel
    CouponUuid   uint64 // 优惠券UUID
    SerialNo     string // 记录编号（yyMMddhhmmssxxxx格式）
    ActivityUuid uint64 // 活动UUID
    MemberUuid   uint64 // 会员UUID
    Type         int    // 记录类型（1-6）
    Count        int    // 变动数量
    LeftCount    int    // 剩余有效张数
}
```

**记录类型常量**:
- `CouponRecordTypeCreate = 1` - 首次添加
- `CouponRecordTypeIncrease = 2` - 调整添加
- `CouponRecordTypeDecrease = 3` - 调整扣减
- `CouponRecordTypeReverseSettle = 4` - 反结账退还
- `CouponRecordTypeBonus = 5` - 奖励领取（冻结）
- `CouponRecordTypeUsed = 6` - 核销扣减

---

## 🚀 性能优化

### 查询优化

1. **预加载关联数据**:
   - 使用 `Preload` 预加载多语言信息
   - 使用 `Preload` 预加载奖品信息
   - 减少N+1查询问题

2. **时间范围查询**:
   - 只查询最近7天内的活动
   - 减少不必要的数据查询

3. **排序优化**:
   - 使用SQL CASE语句进行状态排序
   - 避免在应用层排序

### 缓存策略

1. **活动列表缓存**:
   - 可以缓存活动列表数据
   - 活动状态变化时清除缓存

2. **二维码缓存**:
   - 相同参数的二维码可以缓存
   - 减少重复生成

### 数据压缩

- **二维码数据**: 使用AES加密，减少数据大小
- **Base64编码**: 二维码图片使用Base64编码传输

---

## 🧪 测试建议

### 单元测试

1. **活动状态测试**:
   - 测试活动状态计算逻辑
   - 测试活动有效性判断
   - 测试不同时间点的状态

2. **二维码测试**:
   - 测试二维码生成
   - 测试二维码解密
   - 测试二维码数据完整性

3. **活动查询测试**:
   - 测试活动列表查询
   - 测试活动详情查询
   - 测试活动排序逻辑

### 集成测试

1. **完整流程测试**:
   - 测试活动创建-查询-参与完整流程
   - 测试二维码生成-分享-扫描流程
   - 测试奖励发放流程

2. **边界条件测试**:
   - 测试活动开始/结束时间边界
   - 测试奖励次数限制
   - 测试活动失效场景

### 性能测试

1. **查询性能**:
   - 测试活动列表查询性能
   - 测试大量活动时的查询性能

2. **二维码生成性能**:
   - 测试二维码生成速度
   - 测试并发生成二维码

---

## 📝 注意事项

1. **活动时间**:
   - 活动开始时间不能早于当前时间
   - 活动结束时间必须大于开始时间
   - 活动时间使用Unix时间戳存储

2. **活动状态**:
   - 活动状态由系统自动计算
   - 活动可以手动设置为失效（`IsInvalid = 1`）
   - 失效的活动状态为"已结束"

3. **奖励限制**:
   - 奖励次数限制只对开启限制的活动有效
   - 每个会员的奖励次数独立计算
   - 奖励次数通过活动记录表统计

4. **二维码安全**:
   - 二维码数据使用AES加密
   - 解密时需要验证活动和会员是否存在
   - 二维码包含敏感信息，需要妥善保管

5. **多语言支持**:
   - 活动名称和描述支持多语言
   - 根据用户语言返回对应语言的内容
   - 如果没有对应语言，返回默认语言

6. **商家功能开关**:
   - 商家必须开启会员功能才能使用营销活动
   - 商家必须开启营销活动功能
   - 功能未开启时，返回错误但保留数据

7. **活动查询范围**:
   - 会员端只显示最近7天内的活动
   - 包括进行中、未开始、已结束的活动
   - 按状态和开始时间排序

8. **优惠券奖励**:
   - 活动可以设置多个优惠券奖品
   - 每个奖品对应一个优惠券
   - 优惠券需要满足使用条件才能使用

9. **积分奖励**:
   - 积分奖励是固定值
   - 奖励直接增加到会员账户
   - 积分奖励会触发会员升级检查

---

## 🔗 相关文档

- [会员服务文档](./member.md) - 会员相关功能
- [优惠券服务文档](./coupon.md) - 优惠券相关功能
- [设置服务文档](./setting.md) - 商家设置相关配置

---

**文档版本**: v1.0  
**最后更新**: 2025-01-27  
**维护者**: TTPOS开发团队

