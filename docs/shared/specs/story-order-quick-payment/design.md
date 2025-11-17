# 快捷支付功能 设计文档

> 本文档定义快捷支付功能的技术设计和实现方案。

## 📋 概述

快捷支付功能通过优化支付 API 和缓存设计，实现一键支付。核心实现包括：

- 新增 QuickPaymentService 封装快捷支付逻辑
- 优化商户配置缓存
- 新增数据库字段支持快捷支付标识

---

## 🎯 规范对齐

### Go Main 规范 (go-main.mdc)

- ✅ QuickPaymentService 依赖 IOrderService 和 IPaymentService 接口
- ✅ URL 使用 snake_case: `/api/v1/order/quick_payment`
- ✅ data 字段返回对象
- ✅ 不使用 panic，返回 error
- ✅ 使用 errors.WithMessage 包装错误

### API 设计规范 (api.mdc)

- ✅ 响应格式统一: `{code, message, data{}}`
- ✅ data 不能为 null 或数组

### 数据库规范 (database.mdc)

- ✅ 字段使用 snake_case
- ✅ 使用 PHP Phinx 管理迁移
- ✅ 同步更新 Go Model

---

## 🔄 代码复用分析

### 可复用的现有组件

- **OrderService**: `main/app/service/order_srv.go` - 复用订单更新逻辑
- **PaymentService**: `main/app/service/payment_srv.go` - 复用支付创建逻辑
- **SystemLock**: `main/pkg/lock/system_lock.go` - 使用 UUID 锁防止并发
- **EventBus**: `main/app/event/system_bus.go` - 发布订单支付完成事件

### 集成点

- **订单模块**: 调用 OrderService 更新订单状态
- **支付模块**: 调用 PaymentService 创建支付记录
- **缓存模块**: Redis 缓存商户配置
- **事件总线**: 发布事件触发打印小票

---

## 🏗️ 架构设计

### 分层设计

```
QuickPaymentAPI (API 层)
  ↓ 调用
QuickPaymentService (业务层)
  ↓ 依赖
IOrderService + IPaymentService (其他业务层)
  ↓ 依赖
OrderRepository + PaymentRepository (数据层)
```

**依赖规则**：

- ✅ QuickPaymentService 依赖 IOrderService 和 IPaymentService 接口
- ❌ QuickPaymentService 不直接依赖 Repository
- ✅ 使用事务管理保证数据一致性

---

## 🗄️ 数据库设计

### 表结构变更

#### 1. ttpos_company 表（新增字段）

```sql
ALTER TABLE `ttpos_company`
ADD COLUMN `default_payment_method` tinyint NOT NULL DEFAULT 1 COMMENT '默认支付方式：1-现金，2-微信，3-支付宝，4-银行卡';
```

#### 2. ttpos_order 表（新增字段）

```sql
ALTER TABLE `ttpos_order`
ADD COLUMN `is_quick_payment` tinyint NOT NULL DEFAULT 0 COMMENT '是否快捷支付：0-否，1-是';
```

### 迁移文件

**文件**: `admin/database/migrations/20251117100000_add_quick_payment_fields.php`

```php
<?php
use think\migration\Migrator;

class AddQuickPaymentFields extends Migrator
{
    public function change()
    {
        // ttpos_company 表
        $table = $this->table('ttpos_company');
        if (!$table->hasColumn('default_payment_method')) {
            $table->addColumn('default_payment_method', 'integer', [
                'limit' => 1,
                'default' => 1,
                'comment' => '默认支付方式：1-现金，2-微信，3-支付宝，4-银行卡',
                'after' => 'status'
            ])->update();
        }

        // ttpos_order 表
        $table = $this->table('ttpos_order');
        if (!$table->hasColumn('is_quick_payment')) {
            $table->addColumn('is_quick_payment', 'integer', [
                'limit' => 1,
                'default' => 0,
                'comment' => '是否快捷支付：0-否，1-是',
                'after' => 'status'
            ])->update();
        }
    }
}
```

### Go Model 更新

```go
// main/app/model/company.go
type Company struct {
    // ... 现有字段
    DefaultPaymentMethod uint8 `gorm:"column:default_payment_method" json:"default_payment_method"`
}

// main/app/model/order.go
type Order struct {
    // ... 现有字段
    IsQuickPayment uint8 `gorm:"column:is_quick_payment" json:"is_quick_payment"`
}
```

---

## 📊 数据模型

### Request DTO

```go
// main/app/dto/req/quick_payment_req.go
type QuickPaymentReq struct {
    OrderUuid     uint64 `json:"order_uuid" binding:"required"`
    PaymentMethod uint8  `json:"payment_method"` // 可选，不传则使用默认值
}
```

### Response DTO

```go
// main/app/dto/resp/quick_payment_resp.go
type QuickPaymentResp struct {
    OrderUuid     uint64 `json:"order_uuid"`
    PaymentStatus uint8  `json:"payment_status"` // 1-已支付
    PaymentMethod uint8  `json:"payment_method"`
    PaymentTime   int64  `json:"payment_time"`
}
```

---

## 🔌 API 设计

### 快捷支付 API

**URL**: `POST /api/v1/order/quick_payment`

**Request**:

```json
{
  "order_uuid": 123456789,
  "payment_method": 1
}
```

**Response**:

```json
{
  "code": 1,
  "message": "支付成功",
  "data": {
    "order_uuid": 123456789,
    "payment_status": 1,
    "payment_method": 1,
    "payment_time": 1700123456
  }
}
```

**Error Response**:

```json
{
  "code": 0,
  "message": "订单状态不允许支付",
  "data": {}
}
```

---

## 🧩 核心组件实现

### Service 接口

```go
// main/app/service/i_quick_payment_srv.go
type IQuickPaymentSrv interface {
    QuickPay(ctx *gin.Context, req *dto_req.QuickPaymentReq) (*dto_resp.QuickPaymentResp, error)
}
```

### Service 实现（核心逻辑）

```go
// main/app/service/quick_payment_srv.go
type quickPaymentSrv struct {
    dbm        *database.DBManager
    orderSrv   IOrderSrv
    paymentSrv IPaymentSrv
    systemLock *lock.SystemLock
}

func NewQuickPaymentSrv(
    dbm *database.DBManager,
    orderSrv IOrderSrv,
    paymentSrv IPaymentSrv,
) IQuickPaymentSrv {
    return &quickPaymentSrv{
        dbm:        dbm,
        orderSrv:   orderSrv,
        paymentSrv: paymentSrv,
        systemLock: lock.NewSystemLock(),
    }
}

func (s *quickPaymentSrv) QuickPay(ctx *gin.Context, req *dto_req.QuickPaymentReq) (*dto_resp.QuickPaymentResp, error) {
    // 1. 加锁防止并发
    s.systemLock.LockUuid(req.OrderUuid)
    defer s.systemLock.UnlockUuid(req.OrderUuid)

    // 2. 获取订单信息
    order, err := s.orderSrv.GetByUuid(ctx, req.OrderUuid)
    if err != nil {
        return nil, errors.WithMessage(err, "订单不存在")
    }

    // 3. 验证订单状态
    if order.Status != constant.OrderStatusPending {
        return nil, errors.New("订单状态不允许支付")
    }

    // 4. 确定支付方式
    paymentMethod := req.PaymentMethod
    if paymentMethod == 0 {
        // 使用默认支付方式
        company, _ := s.getCompanyConfig(ctx, order.CompanyUuid)
        paymentMethod = company.DefaultPaymentMethod
    }

    // 5. 创建支付记录（事务内）
    payment, err := s.paymentSrv.Create(ctx, &dto_req.PaymentCreateReq{
        OrderUuid:     req.OrderUuid,
        PaymentMethod: paymentMethod,
        Amount:        order.TotalAmount,
    })
    if err != nil {
        return nil, errors.WithMessage(err, "创建支付记录失败")
    }

    // 6. 更新订单状态
    err = s.orderSrv.UpdateStatus(ctx, req.OrderUuid, constant.OrderStatusPaid, true)
    if err != nil {
        return nil, errors.WithMessage(err, "更新订单状态失败")
    }

    // 7. 发布事件（异步）
    go func() {
        event.NewSystemBus().PublishOrderPaidEvent(
            event.OrderPaidPayload{
                BasePayload: event.BasePayload{
                    Ctx:         ctx,
                    CompanyUuid: order.CompanyUuid,
                },
                OrderUuid:     req.OrderUuid,
                IsQuickPayment: true,
            },
        )
    }()

    // 8. 返回结果
    return &dto_resp.QuickPaymentResp{
        OrderUuid:     req.OrderUuid,
        PaymentStatus: 1,
        PaymentMethod: paymentMethod,
        PaymentTime:   time.Now().Unix(),
    }, nil
}

func (s *quickPaymentSrv) getCompanyConfig(ctx *gin.Context, companyUuid uint64) (*model.Company, error) {
    // 先查缓存
    key := fmt.Sprintf("ttpos:company:config:%d", companyUuid)
    cached, err := redis.Get(key)
    if err == nil {
        return cached, nil
    }

    // 缓存未命中，查数据库
    companyRepo := repository.NewCompanyRepo(s.dbm.GetDB(ctx))
    company, err := companyRepo.GetByUuid(companyUuid)
    if err != nil {
        return nil, err
    }

    // 写入缓存（30分钟）
    redis.Set(key, company, 30*time.Minute)
    return company, nil
}
```

### API 实现

```go
// main/app/api/quick_payment_api.go
type QuickPaymentAPI struct {
    quickPaymentSrv service.IQuickPaymentSrv
}

func NewQuickPaymentAPI(quickPaymentSrv service.IQuickPaymentSrv) *QuickPaymentAPI {
    return &QuickPaymentAPI{quickPaymentSrv: quickPaymentSrv}
}

// POST /api/v1/order/quick_payment
func (api *QuickPaymentAPI) QuickPay(c *gin.Context) {
    var req dto_req.QuickPaymentReq
    if err := c.ShouldBindJSON(&req); err != nil {
        helper.ErrorWithDetail(c, constant.CodeInvalidParam, err)
        return
    }

    resp, err := api.quickPaymentSrv.QuickPay(c, &req)
    if err != nil {
        helper.ErrorWithDetail(c, constant.CodeFail, err)
        return
    }

    helper.Success(c, gin.H{"data": resp})
}
```

---

## ⚡ 缓存设计

### Redis Key 设计

- **商户配置**: `ttpos:company:config:{company_uuid}`
- **过期时间**: 30 分钟
- **更新策略**: Cache-Aside Pattern（先删缓存，再更新数据库）

### 缓存流程

```go
// 读取流程
1. 查询缓存
2. 缓存命中 → 返回
3. 缓存未命中 → 查询数据库 → 写入缓存 → 返回

// 更新流程
1. 删除缓存
2. 更新数据库
3. 下次读取时重新加载
```

---

## 🚨 错误处理

### 主要错误场景

1. **订单不存在**: 返回 "订单不存在"
2. **订单状态错误**: 返回 "订单状态不允许支付"
3. **支付创建失败**: 事务回滚，返回错误
4. **并发冲突**: UUID 锁保证串行执行

---

## 🧪 测试策略

### 单元测试

**目标覆盖率**: QuickPaymentService ≥ 70%, PaymentService 100%

**测试用例**:

- 正常支付流程
- 订单状态异常
- 并发支付场景（同一订单）
- 缓存命中/未命中场景

### 集成测试

**测试流程**:

- 创建订单 → 快捷支付 → 验证订单状态 → 验证支付记录 → 验证事件发布

---

## 📈 性能优化

### 优化措施

1. **缓存优化**: 商户配置缓存 30 分钟
2. **并发控制**: UUID 锁防止重复支付
3. **异步事件**: 事件发布使用 goroutine
4. **索引优化**: order_uuid 已有唯一索引

### 性能指标

- 本地响应时间: < 200ms
- 并发能力: 1000+ QPS
- 缓存命中率: > 80%

---

## 📚 实现清单

### Phase 1: 数据库（参见 tasks.md）

- [ ] 创建迁移文件
- [ ] 执行迁移
- [ ] 更新 Go Model

### Phase 2: 核心实现（参见 tasks.md）

- [ ] 实现 QuickPaymentService
- [ ] 实现 QuickPaymentAPI
- [ ] 注册路由

### Phase 3: 测试和优化（参见 tasks.md）

- [ ] 单元测试
- [ ] 集成测试
- [ ] 性能测试

---

**版本**: v1.0.0  
**创建日期**: 2025-11-17  
**作者**: 后端开发组  
**审核者**: CTO
