# Grab/LINE MAN 外卖支付方式自动创建 设计文档

> 本文档定义 Grab/LINE MAN 外卖支付方式自动创建功能的技术设计和实现方案。

## 📋 概述

在配置 Grab 外卖或 LINE MAN 外卖时，系统自动创建对应的支付方式（code = 91100 和 91200），确保支付方式数据完整，并支持自动同步到 ERP 系统。同时，在获取支付方式列表时过滤掉这些系统自动创建的支付方式，不在旧后台和新管理端显示。

---

## 🎯 规范对齐

### Go Main 规范 (go-main.mdc)

- Service 只依赖其他 Service 接口
- Repository 只持有 db 实例
- URL 使用 snake_case
- data 字段必须是对象
- 不使用 panic，返回 error

### PHP 规范 (php.mdc)

- 遵循 MVC 分层
- Controller 不写业务逻辑
- 使用验证器验证参数

### API 设计规范 (api.mdc)

- URL 使用 snake_case
- 响应格式统一
- data 不能为 null 或数组

---

## 🔄 代码复用分析

### 可复用的现有组件

- **PaymentMethodService**: `main/app/service/payment_method.go` - 支付方式服务，包含 Create 方法和 ERP 同步逻辑
- **ERP Service**: `main/app/service/rpc/erp/selling.go` - ERP 同步服务，用于同步支付方式到 ERP
- **PaymentMethodRepository**: `main/app/repository/payment_method.go` - 支付方式数据访问层
- **PaymentMethod Model**: `main/app/model/payment_method.go` - 支付方式数据模型

### 集成点

- **外卖平台配置保存**: 在外卖平台配置保存时触发支付方式创建
- **支付方式列表查询**: 在 GetList 和 GetManagementList 方法中过滤 Grab/LINE MAN 支付方式
- **ERP 同步**: 复用现有的 ERP 同步逻辑

---

## 🏗️ 架构设计

### 分层设计原则

**Go Main 三层架构**:

```
API 层 (Controller/API)
  ↓ 依赖
业务层 (Service)
  ↓ 依赖
数据层 (Repository)
```

**依赖规则**:

- ✅ 上层可依赖下层
- ❌ 禁止下层依赖上层
- ❌ 禁止跨层调用
- ❌ Service 不能依赖 Repository
- ✅ Service 可以依赖其他 Service 接口

### 模块划分

#### Go Main 模块

- **Service 层**: `main/app/service/payment_method.go` - 添加 SaveGrabPaymentMethod 和 SaveLineManPaymentMethod 方法
- **Repository 层**: `main/app/repository/payment_method.go` - 复用现有 Repository，添加过滤选项方法
- **Model 层**: `main/app/model/payment_method.go` - 复用现有 Model
- **Constant 层**: `main/app/constant/payment.go` - 添加 Grab 和 LINE MAN 的 code 常量

#### PHP Admin 模块

- **Controller 层**: `admin/app/shop/controller/setting/Paytype.php` - 在 index 方法中过滤 Grab/LINE MAN 支付方式
- **Model 层**: `admin/app/common/model/store/PayType.php` - 复用现有 Model

---

## 🗄️ 数据库设计

### 数据表设计

无需新增数据表，使用现有的 `ttpos_payment_method` 表。

**表结构**:

```sql
CREATE TABLE `ttpos_payment_method` (
    `id` bigint unsigned NOT NULL AUTO_INCREMENT,
    `uuid` bigint unsigned NOT NULL DEFAULT 0,
    `name` varchar(255) NOT NULL DEFAULT '',
    `code` int NOT NULL DEFAULT 0,
    `payment_name` varchar(255) NOT NULL DEFAULT '',
    `source` tinyint NOT NULL DEFAULT 0 COMMENT '0-系统默认 1-自行添加 2-LianLianPay',
    `logo_file_uuid` bigint unsigned NOT NULL DEFAULT 0,
    `qrcode_file_uuid` bigint unsigned NOT NULL DEFAULT 0,
    `default_img` varchar(255) NOT NULL DEFAULT '',
    `fee_percent` decimal(20,8) NOT NULL DEFAULT 0.00000000,
    `is_show_cashier` tinyint NOT NULL DEFAULT 0,
    `is_show_assistant` tinyint NOT NULL DEFAULT 0,
    `is_show_kiosk` tinyint NOT NULL DEFAULT 0,
    `is_show_member_recharge` tinyint NOT NULL DEFAULT 0,
    `status` tinyint NOT NULL DEFAULT 0,
    `sort` int NOT NULL DEFAULT 0,
    `erpnext_payment` varchar(255) NOT NULL DEFAULT '',
    `headquarter_uuid` bigint unsigned NOT NULL DEFAULT 0,
    `create_time` int NOT NULL DEFAULT 0,
    `update_time` int NOT NULL DEFAULT 0,
    `delete_time` int NOT NULL DEFAULT 0,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_uuid` (`uuid`),
    KEY `idx_code` (`code`),
    KEY `idx_payment_name` (`payment_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='支付方式表';
```

**字段说明**:

| 字段 | 类型 | 说明 | 约束 |
|------|------|------|------|
| code | int | 支付方式代码 | Grab=91100, LINE MAN=91200 |
| payment_name | varchar(255) | 支付方式名称 | "Grab" 或 "LINE MAN" |
| source | tinyint | 来源标识 | 0=系统默认 |
| is_show_cashier | tinyint | 收银端显示 | 0=不显示 |
| is_show_assistant | tinyint | 助手端显示 | 0=不显示 |
| is_show_kiosk | tinyint | 自助机显示 | 0=不显示 |
| is_show_member_recharge | tinyint | 会员充值显示 | 0=不显示 |

---

## 📊 数据模型

### Go Model

复用现有的 `model.PaymentMethod` 结构体，无需修改。

### 常量定义

需要在 `main/app/constant/payment.go` 中添加：

```go
const (
    PaymentMethodCodeGrab     = 91100 // Grab 支付方式
    PaymentMethodCodeLineMan  = 91200 // LINE MAN 支付方式
)

const (
    PaymentMethodNameGrab     = "Grab"     // Grab 支付方式名称
    PaymentMethodNameLineMan  = "LINE MAN" // LINE MAN 支付方式名称
)
```

---

## 🔌 API 设计

### RESTful API

无需新增 API 接口，复用现有的支付方式管理接口。

### Service 方法设计

#### SaveGrabPaymentMethod

**方法签名**:

```go
func (s *paymentMethodSrv) SaveGrabPaymentMethod(ctx context.Context, tx *gorm.DB) error
```

**功能描述**: 保存 Grab 支付方式，如果已存在则跳过。该方法供外部服务调用，支持事务。

**实现逻辑**:

1. 检查支付方式是否已存在（通过 payment_name = constant.PaymentMethodNameGrab 或 code = constant.PaymentMethodCodeGrab）
2. 如果不存在，创建新的支付方式记录
3. 设置支付方式属性：
   - Source = 0（系统默认）
   - Code = constant.PaymentMethodCodeGrab
   - PaymentName = constant.PaymentMethodNameGrab
   - Name = constant.PaymentMethodNameGrab
   - IsShowCashier = 0
   - IsShowAssistant = 0
   - IsShowKiosk = 0
   - IsShowMemberRecharge = 0
   - Status = 1（启用）
   - DefaultImg = ""（空字符串）
4. 如果商户开启 ERP，调用 ERP 同步接口
5. 记录创建日志

**注意事项**:
- 方法接收 `tx *gorm.DB` 参数，由外部服务在事务中调用
- 所有数据库操作使用传入的 `tx`，不内部开启事务
- **ERP 同步失败会返回错误，阻塞流程**（如果开启 ERP，必须同步成功才能创建支付方式）

#### SaveLineManPaymentMethod

**方法签名**:

```go
func (s *paymentMethodSrv) SaveLineManPaymentMethod(ctx context.Context, tx *gorm.DB) error
```

**功能描述**: 保存 LINE MAN 支付方式，如果已存在则跳过。该方法供外部服务调用，支持事务。

**实现逻辑**: 同 SaveGrabPaymentMethod，但 code = constant.PaymentMethodCodeLineMan，PaymentName = constant.PaymentMethodNameLineMan，DefaultImg = ""

#### GetList 过滤逻辑

**修改位置**: `main/app/service/payment_method.go` 的 `GetList` 方法

**过滤逻辑**: 在返回结果前，过滤掉 code = constant.PaymentMethodCodeGrab 或 constant.PaymentMethodCodeLineMan 的支付方式

#### GetManagementList 过滤逻辑

**修改位置**: `main/app/service/payment_method.go` 的 `GetManagementList` 方法

**过滤逻辑**: 在查询时添加过滤条件，排除 code = constant.PaymentMethodCodeGrab 或 constant.PaymentMethodCodeLineMan 的支付方式

---

## 🧩 组件和接口

### Service 层

#### Service 接口扩展

```go
// main/app/service/i_payment_method_srv.go
type IPaymentMethodSrv interface {
    // ... 现有方法 ...
    
    // 新增方法（供外部服务调用，支持事务）
    SaveGrabPaymentMethod(ctx context.Context, tx *gorm.DB) error
    SaveLineManPaymentMethod(ctx context.Context, tx *gorm.DB) error
}
```

#### Service 实现

```go
// main/app/service/payment_method.go
import (
    "strings"
    // ... 其他导入
)

// SaveGrabPaymentMethod 保存 Grab 支付方式（供外部服务调用，支持事务）
func (s *paymentMethodSrv) SaveGrabPaymentMethod(ctx context.Context, tx *gorm.DB) error {
    paymentMethodRepo := repository.NewPaymentMethodRepo(tx)
    
    // 检查是否已存在（使用 Repository 方法）
    // 先通过 payment_name 查找
    existPayment, err := paymentMethodRepo.GetPaymentMethodError(
        paymentMethodRepo.WherePaymentName(constant.PaymentMethodNameGrab),
        repository.CommonRepo.WhereBySoftDelete(),
    )
    
    // 如果通过 payment_name 没找到，再通过 code 查找
    if err != nil {
        existPayment, err = paymentMethodRepo.GetPaymentMethodError(
            paymentMethodRepo.WhereCode(constant.PaymentMethodCodeGrab),
            repository.CommonRepo.WhereBySoftDelete(),
        )
    }
    
    // 如果找到了，说明已存在，跳过创建
    if err == nil && existPayment != nil {
        logger.Logger.Info("Grab 支付方式已存在，跳过创建",
            zap.Uint64("uuid", existPayment.Uuid),
            zap.Int("code", existPayment.Code))
        return nil
    }
    
    // 如果错误不是记录不存在，说明查询出错
    if err != nil {
        // 检查是否是记录不存在的错误（使用字符串匹配）
        if !strings.Contains(err.Error(), "record not found") {
            return errors.WithMessage(err, "查询支付方式失败")
        }
        // 记录不存在，继续创建流程
    }
    
    // 创建支付方式
    paymentMethod := &model.PaymentMethod{
        Code:                 constant.PaymentMethodCodeGrab,
        PaymentName:          constant.PaymentMethodNameGrab,
        Name:                 constant.PaymentMethodNameGrab,
        Source:               constant.PaymentMethodSourceSystem,
        IsShowCashier:        0,
        IsShowAssistant:      0,
        IsShowKiosk:          0,
        IsShowMemberRecharge: 0,
        Status:               constant.PaymentMethodStatusEnable,
        Sort:                 0,
        DefaultImg:           "", // 空字符串
    }
    
    // 使用传入的 tx 创建支付方式
    if err := paymentMethodRepo.CreatePaymentMethodReturnRow(paymentMethod); err != nil {
        logger.Logger.Error("创建 Grab 支付方式失败", zap.Error(err))
        return errors.WithMessage(err, "创建 Grab 支付方式失败")
    }
    
    // 如果开启了 ERP，同步支付方式到 ERP（ERP 同步失败会阻塞流程）
    if ctx.GetCompany().IsOpenErp() {
        erpSrv := erpService.NewIErpSrv(s.dbm)
        channel := erpService.GetChannelBySource(paymentMethod.Source)
        
        saveModeOfPaymentResp, err := erpSrv.SaveModeOfPayment(ctx, req.SaveModeOfPaymentReq{
            CompanyUuid: ctx.GetCompanyUuid(),
            Channel:     channel,
            PayType:     paymentMethod.PaymentName,
        })
        if err != nil || saveModeOfPaymentResp == nil {
            logger.Logger.Error("创建 Grab 支付方式失败：ERP 同步失败", zap.Error(err))
            return errors.WithMessage(err, "创建 Grab 支付方式失败")
        }
        
        // 更新 ERP 支付方式名称
        if saveModeOfPaymentResp.Name != "" {
            if err := paymentMethodRepo.UpdatePaymentMethod(
                map[string]any{"erpnext_payment": saveModeOfPaymentResp.Name},
                repository.CommonRepo.WhereByUuid(paymentMethod.Uuid),
            ); err != nil {
                logger.Logger.Error("创建 Grab 支付方式失败：更新 ERP 支付方式名称失败", zap.Error(err))
                return errors.WithMessage(err, "创建 Grab 支付方式失败")
            }
        }
    }
    
    return nil
}

// SaveLineManPaymentMethod 保存 LINE MAN 支付方式（供外部服务调用，支持事务）
func (s *paymentMethodSrv) SaveLineManPaymentMethod(ctx context.Context, tx *gorm.DB) error {
    paymentMethodRepo := repository.NewPaymentMethodRepo(tx)
    
    // 检查是否已存在（使用 Repository 方法）
    // 先通过 payment_name 查找
    existPayment, err := paymentMethodRepo.GetPaymentMethodError(
        paymentMethodRepo.WherePaymentName(constant.PaymentMethodNameLineMan),
        repository.CommonRepo.WhereBySoftDelete(),
    )
    
    // 如果通过 payment_name 没找到，再通过 code 查找
    if err != nil {
        existPayment, err = paymentMethodRepo.GetPaymentMethodError(
            paymentMethodRepo.WhereCode(constant.PaymentMethodCodeLineMan),
            repository.CommonRepo.WhereBySoftDelete(),
        )
    }
    
    // 如果找到了，说明已存在，跳过创建
    if err == nil && existPayment != nil {
        logger.Logger.Info("LINE MAN 支付方式已存在，跳过创建",
            zap.Uint64("uuid", existPayment.Uuid),
            zap.Int("code", existPayment.Code))
        return nil
    }
    
    // 如果错误不是记录不存在，说明查询出错
    if err != nil {
        // 检查是否是记录不存在的错误（使用字符串匹配）
        if !strings.Contains(err.Error(), "record not found") {
            return errors.WithMessage(err, "查询支付方式失败")
        }
        // 记录不存在，继续创建流程
    }
    
    // 创建支付方式
    paymentMethod := &model.PaymentMethod{
        Code:                 constant.PaymentMethodCodeLineMan,
        PaymentName:          constant.PaymentMethodNameLineMan,
        Name:                 constant.PaymentMethodNameLineMan,
        Source:               constant.PaymentMethodSourceSystem,
        IsShowCashier:        0,
        IsShowAssistant:      0,
        IsShowKiosk:          0,
        IsShowMemberRecharge: 0,
        Status:               constant.PaymentMethodStatusEnable,
        Sort:                 0,
        DefaultImg:           "", // 空字符串
    }
    
    // 使用传入的 tx 创建支付方式
    if err := paymentMethodRepo.CreatePaymentMethodReturnRow(paymentMethod); err != nil {
        logger.Logger.Error("创建 LINE MAN 支付方式失败", zap.Error(err))
        return errors.WithMessage(err, "创建 LINE MAN 支付方式失败")
    }
    
    // 如果开启了 ERP，同步支付方式到 ERP（ERP 同步失败会阻塞流程）
    if ctx.GetCompany().IsOpenErp() {
        erpSrv := erpService.NewIErpSrv(s.dbm)
        channel := erpService.GetChannelBySource(paymentMethod.Source)
        
        saveModeOfPaymentResp, err := erpSrv.SaveModeOfPayment(ctx, req.SaveModeOfPaymentReq{
            CompanyUuid: ctx.GetCompanyUuid(),
            Channel:     channel,
            PayType:     paymentMethod.PaymentName,
        })
        if err != nil || saveModeOfPaymentResp == nil {
            logger.Logger.Error("创建 LINE MAN 支付方式失败：ERP 同步失败", zap.Error(err))
            return errors.WithMessage(err, "创建 LINE MAN 支付方式失败")
        }
        
        // 更新 ERP 支付方式名称
        if saveModeOfPaymentResp.Name != "" {
            if err := paymentMethodRepo.UpdatePaymentMethod(
                map[string]any{"erpnext_payment": saveModeOfPaymentResp.Name},
                repository.CommonRepo.WhereByUuid(paymentMethod.Uuid),
            ); err != nil {
                logger.Logger.Error("创建 LINE MAN 支付方式失败：更新 ERP 支付方式名称失败", zap.Error(err))
                return errors.WithMessage(err, "创建 LINE MAN 支付方式失败")
            }
        }
    }
    
    return nil
}
```

### Repository 层

#### Repository 选项方法扩展

```go
// main/app/repository/payment_method.go

// WherePaymentName 按支付方式名称查询
func (r *paymentMethodRepo) WherePaymentName(paymentName string) DBOption {
    return func(db *gorm.DB) *gorm.DB {
        return db.Where("payment_name = ?", paymentName)
    }
}

// WhereCode 按支付方式 code 查询
func (r *paymentMethodRepo) WhereCode(code int) DBOption {
    return func(db *gorm.DB) *gorm.DB {
        return db.Where("code = ?", code)
    }
}

// WhereNotCode 排除指定 code 的支付方式
func (r *paymentMethodRepo) WhereNotCode(codes []int) DBOption {
    return func(db *gorm.DB) *gorm.DB {
        return db.Where("code NOT IN ?", codes)
    }
}
```

**注意**: `WhereNotCode` 方法在现有代码中可能已存在，需要检查确认。

### API 层

无需新增 API 接口，支付方式创建在外卖平台配置保存时自动触发。

---

## ⚡ 缓存设计

无需新增缓存逻辑。

---

## 🚨 错误处理

### 错误场景

#### 场景 1: 支付方式已存在

- **处理方式**: 记录日志，跳过创建，返回 nil
- **用户影响**: 无影响，幂等性保证

#### 场景 2: ERP 同步失败

- **处理方式**: 记录错误日志，但不影响支付方式创建
- **用户影响**: 支付方式创建成功，但 ERP 同步失败，可后续手动同步

#### 场景 3: 数据库操作失败

- **处理方式**: 返回错误，记录日志
- **用户影响**: 支付方式创建失败，不影响外卖配置保存

---

## 🔒 安全设计

### 身份验证

- 所有 API 需要 JWT Token 验证（复用现有机制）

### 权限控制

- 支付方式创建在外卖平台配置保存时自动触发，无需额外权限控制

### 数据安全

- 使用参数化查询防止 SQL 注入
- 支付方式 code 固定值，防止恶意修改

---

## 🧪 测试策略

### 单元测试

**目标覆盖率**:

- Service: 70%+
- Repository: 80%+

**测试内容**:

- SaveGrabPaymentMethod 方法测试
- SaveLineManPaymentMethod 方法测试
- 支付方式列表过滤测试
- ERP 同步测试
- 幂等性测试

### API 测试

**测试内容**:

- 支付方式列表接口过滤测试
- 管理端支付方式列表接口过滤测试

### 集成测试

**测试流程**:

- 配置 Grab 外卖 → 自动创建 Grab 支付方式 → 验证支付方式列表不显示
- 配置 LINE MAN 外卖 → 自动创建 LINE MAN 支付方式 → 验证支付方式列表不显示
- ERP 同步测试

---

## 📈 性能优化

### 优化策略

1. **数据库优化**:
   - 使用索引查询（code 和 payment_name 字段）
   - 幂等性检查使用单次查询

2. **查询优化**:
   - 列表查询时使用 WHERE NOT IN 过滤，避免内存过滤

### 性能指标

- 支付方式创建: < 100ms
- 列表查询过滤: < 50ms

---

## 📚 实现清单

### Phase 1: 常量定义

- [ ] 在 `main/app/constant/payment.go` 中添加 Grab 和 LINE MAN 的 code 常量和名称常量

### Phase 2: Service 层实现

- [ ] 实现 SaveGrabPaymentMethod 方法
- [ ] 实现 SaveLineManPaymentMethod 方法
- [ ] 修改 GetList 方法，过滤 Grab/LINE MAN 支付方式
- [ ] 修改 GetManagementList 方法，过滤 Grab/LINE MAN 支付方式

### Phase 3: Repository 层扩展

- [ ] 添加 WhereNotCode 选项方法

### Phase 4: PHP Admin 模块

- [ ] 修改 `admin/app/shop/controller/setting/Paytype.php` 的 index 方法，过滤 Grab/LINE MAN 支付方式

### Phase 5: 测试

- [ ] 单元测试
- [ ] API 测试
- [ ] 集成测试

**详细任务**: 参见 `tasks.md`

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 在技术方案评审、关键架构决策或踩坑总结后，立即补充 Episode 并在设计文档尾部互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-12-22  
**作者**: 王昱  
**审核者**: 待审核

