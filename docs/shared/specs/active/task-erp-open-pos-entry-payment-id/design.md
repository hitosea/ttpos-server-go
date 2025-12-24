# 开账接口支持 PaymentID 设计文档

> 本文档定义开账接口支持 PaymentID 的技术设计和实现方案。

## 📋 概述

本功能为 `OpenPosEntry` gRPC 接口的 `OpenPosEntryDetail` 消息增加 `payment_id` 可选参数支持，使调用方可以直接使用 PaymentID 而无需提前查询 `mode_of_payment`。系统在收到 `payment_id` 后自动调用 `GetModeOfPayment` 服务查询对应的支付方式名称，简化调用流程并与其他接口保持设计一致。

**关键特性**：
- 向后兼容：保持原有 `mode_of_payment` 字段可用
- 自动查询：`payment_id` 不为空时自动查询对应的 `mode_of_payment`
- 参数校验：两个字段不能同时为空
- 错误处理：查询失败返回明确的错误信息

---

## 🎯 规范对齐

### Go BMP 规范 (go-bmp.mdc)

本功能严格遵循 GoFrame 开发规范：

- **禁止修改自动生成代码**：`dao/`、`entity/`、`do/` 目录的代码由 `gf gen dao` 自动生成，不手动修改
- **Protobuf 规范**：遵循 `ttpos-bmp/.cursor/rules/proto-rules.mdc`，字段命名使用 snake_case，消息命名以 Req/Resp 结尾
- **Logic 层设计**：
  - 只实现业务逻辑，不返回 `erp.ResponseInfo` 类型
  - 使用 `gerror` 包处理错误
  - 通过 Service 接口调用其他服务
- **Controller 层设计**：
  - 负责将业务数据包装为 `erp.ResponseInfo`
  - 处理 gRPC 请求和响应

### Protobuf 规范 (proto-rules.mdc)

- **字段命名**：使用 snake_case（如：`payment_id`、`mode_of_payment`）
- **字段类型**：使用 `optional string` 表示可选字段
- **字段编号**：新增字段使用递增编号（避免与现有字段冲突）
- **注释规范**：使用中文注释说明字段用途和约束

### API 设计规范 (api.mdc)

- **gRPC 响应格式**：通过 `erp.ResponseInfo` 统一包装
- **错误信息**：使用中文，便于运维和调试
- **参数验证**：在 Controller 层进行参数校验，Logic 层进行业务验证

---

## 🔄 代码复用分析

### 可复用的现有组件

1. **GetModeOfPayment 服务**
   - 路径：`ttpos-bmp/app/ttpos-erp/internal/logic/selling/selling.go`
   - 方法：`GetModeOfPayment(ctx context.Context, req *selling.GetModeOfPaymentReq) (*selling.ModeOfPayment, error)`
   - 用途：通过 `payment_id` 查询支付方式信息
   - 复用方式：在 `OpenPosEntry` Logic 中直接调用

2. **OpenPosEntry 现有逻辑**
   - 路径：`ttpos-bmp/app/ttpos-erp/internal/logic/selling/selling.go`
   - 方法：`OpenPosEntry(ctx context.Context, req *selling.OpenPosEntryReq) (*selling.OpenPosEntryResp, error)`
   - 复用方式：在现有逻辑基础上增加参数校验和自动查询

3. **ClosePosEntry PaymentID 支持实现**
   - 路径：`ttpos-bmp/app/ttpos-erp/internal/logic/selling/selling.go`
   - 参考：`buildClosingEntryDetails` 方法中的 PaymentID 处理逻辑
   - 复用方式：类似的参数校验和自动查询逻辑

4. **错误处理**
   - 使用：`github.com/gogf/gf/v2/errors/gerror`
   - 方法：`gerror.New()`, `gerror.Wrapf()`

### 集成点

- **现有 API**：扩展 `OpenPosEntry` 接口，不新增接口
- **内部服务**：集成 `GetModeOfPayment` 服务进行支付方式查询
- **数据流**：`payment_id` → `GetModeOfPayment` → `mode_of_payment` → 原有开账逻辑

---

## 🏗️ 架构设计

### 分层设计原则

**Go BMP 微服务架构**：

```
gRPC Controller 层
  ↓ 调用
Logic 层（业务逻辑）
  ↓ 调用（如需要）
Service 层（服务调用）
  ↓
DAO 层（数据访问）
  ↓
Database
```

**本功能的层次调用**：

```
OpenPosEntry gRPC Controller
  ↓
OpenPosEntry Logic
  ↓ (if payment_id not empty)
GetModeOfPayment Service
  ↓
原有开账逻辑
  ↓
ERPNext API
```

### 组件职责

| 组件 | 职责 | 输入 | 输出 |
|------|------|------|------|
| **Controller** | 参数校验、响应包装 | `OpenPosEntryReq` | `ResponseInfo` |
| **Logic** | 业务逻辑、自动查询 | `OpenPosEntryReq` | `OpenPosEntryResp` 或 error |
| **GetModeOfPayment Service** | 查询支付方式 | `GetModeOfPaymentReq` | `ModeOfPayment` 或 error |

---

## 📊 数据模型

### Go DTO 定义

本功能不需要新增 DTO，使用 Protobuf 生成的类型：

**OpenPosEntryDetail（更新后）**：
```go
type OpenPosEntryDetail struct {
    ModeOfPayment *string `protobuf:"bytes,1,opt,name=mode_of_payment,json=modeOfPayment,proto3,oneof"`
    OpeningAmount float64 `protobuf:"fixed64,2,opt,name=opening_amount,json=openingAmount,proto3"`
    PaymentId     *string `protobuf:"bytes,3,opt,name=payment_id,json=paymentId,proto3,oneof"`
}
```

**GetModeOfPaymentReq**：
```go
type GetModeOfPaymentReq struct {
    Name      *string `protobuf:"bytes,1,opt,name=name,proto3,oneof"`
    PaymentId *string `protobuf:"bytes,2,opt,name=payment_id,json=paymentId,proto3,oneof"`
}
```

### 数据转换流程

```
OpenPosEntryDetail.PaymentId
  ↓ (非空)
GetModeOfPaymentReq{PaymentId: detail.PaymentId}
  ↓ 调用 GetModeOfPayment
GetModeOfPaymentResp.Name
  ↓ 提取
modeOfPayment (string)
  ↓ 用于
OpenPosEntry 原有逻辑
```

---

## 🔌 API 设计

### gRPC API

#### Protobuf 定义变更

**文件**: `ttpos-bmp/app/ttpos-erp/manifest/protobuf/selling/selling.proto`

**变更前**:
```protobuf
message OpenPosEntryDetail {
  string mode_of_payment = 1; // 支付方式
  double opening_amount = 2; // 开帐金额
}
```

**变更后**:
```protobuf
message OpenPosEntryDetail {
  optional string mode_of_payment = 1; // 支付方式，与 payment_id 二选一（必填其中之一）
  double opening_amount = 2; // 开帐金额,必填
  optional string payment_id = 3; // 支付方式唯一标识（PaymentID），与 mode_of_payment 二选一
  // 注意：当 payment_id 不为空时，系统自动调用 GetModeOfPayment 查询 mode_of_payment 值
}
```

#### 代码生成命令

```bash
cd ttpos-bmp/app/ttpos-erp
gf gen pb
```

#### 请求示例

**场景 1：使用 payment_id**
```json
{
  "pos_profile_name": "Main Counter",
  "cashier_email": "cashier@example.com",
  "company_abbr": "TTPOS",
  "period_start_date": 1703980800,
  "open_pos_entry_detail": [
    {
      "payment_id": "PID1234567890123456",
      "opening_amount": 1000.00
    }
  ]
}
```

**场景 2：使用 mode_of_payment（向后兼容）**
```json
{
  "pos_profile_name": "Main Counter",
  "cashier_email": "cashier@example.com",
  "company_abbr": "TTPOS",
  "period_start_date": 1703980800,
  "open_pos_entry_detail": [
    {
      "mode_of_payment": "Cash",
      "opening_amount": 1000.00
    }
  ]
}
```

---

## ⚡ 核心实现

### Controller 层实现

**文件**: `ttpos-bmp/app/ttpos-erp/internal/controller/rpc/selling/selling.go`

**实现要点**：

```go
func (*Controller) OpenPosEntry(ctx context.Context, req *selling.OpenPosEntryReq) (*api.ResponseInfo, error) {
    // 参数校验
    for i, detail := range req.OpenPosEntryDetail {
        // 校验 payment_id 和 mode_of_payment 不能同时为空
        if (detail.PaymentId == nil || *detail.PaymentId == "") &&
           (detail.ModeOfPayment == nil || *detail.ModeOfPayment == "") {
            return rpc.ApiError(fmt.Sprintf("open_pos_entry_detail[%d]: payment_id 和 mode_of_payment 不能同时为空", i)), nil
        }
    }

    // 调用 Logic 层
    resp, err := service.Selling().OpenPosEntry(ctx, req)
    if err != nil {
        return rpc.ApiError(err.Error()), nil
    }

    return rpc.ApiSuccess("success", resp), nil
}
```

### Logic 层实现

**文件**: `ttpos-bmp/app/ttpos-erp/internal/logic/selling/selling.go`

**核心方法**：

1. **处理 OpenPosEntryDetail 的辅助方法**：

```go
// buildOpeningEntryDetails 构建开账明细（支持 payment_id 自动查询）
func (s *sSelling) buildOpeningEntryDetails(ctx context.Context, details []*selling.OpenPosEntryDetail) ([]erp.POSOpeningEntryDetail, error) {
    openDetails := make([]erp.POSOpeningEntryDetail, 0)
    
    for i, detail := range details {
        var modeOfPayment string

        // 如果提供了 payment_id，自动查询 mode_of_payment
        if detail.PaymentId != nil && *detail.PaymentId != "" {
            getModeResp, err := service.Selling().GetModeOfPayment(ctx, &selling.GetModeOfPaymentReq{
                PaymentId: detail.PaymentId,
            })
            if err != nil {
                g.Log().Error(ctx, "查询支付方式失败",
                    g.Map{"payment_id": *detail.PaymentId, "error": err.Error()})
                return nil, gerror.Wrapf(err, "查询支付方式失败，payment_id: %s", *detail.PaymentId)
            }

            if getModeResp == nil || getModeResp.Name == "" {
                return nil, gerror.Newf("支付方式不存在或未启用，payment_id: %s", *detail.PaymentId)
            }

            modeOfPayment = getModeResp.Name

            g.Log().Info(ctx, "开账详情: 通过 payment_id 查询到 mode_of_payment",
                g.Map{"index": i, "payment_id": *detail.PaymentId, "mode_of_payment": modeOfPayment})
        } else if detail.ModeOfPayment != nil {
            // 直接使用 mode_of_payment（向后兼容）
            modeOfPayment = *detail.ModeOfPayment
        }

        openDetails = append(openDetails, erp.POSOpeningEntryDetail{
            ModeOfPayment: modeOfPayment,
            OpeningAmount: detail.OpeningAmount,
        })
    }
    
    return openDetails, nil
}
```

2. **更新 OpenPosEntry 方法**：

```go
func (s *sSelling) OpenPosEntry(ctx context.Context, req *selling.OpenPosEntryReq) (*selling.OpenPosEntryResp, error) {
    // 构建开账明细（支持 payment_id 自动查询 mode_of_payment）
    openDetails, err := s.buildOpeningEntryDetails(ctx, req.OpenPosEntryDetail)
    if err != nil {
        return nil, err
    }

    // 原有开账逻辑...
    // 使用 openDetails 继续处理
}
```

---

## 🚨 错误处理

### 错误场景和处理策略

| 错误场景 | 错误码 | 错误信息 | HTTP 状态码 | 处理方式 |
|---------|--------|---------|------------|---------|
| 两个参数都为空 | INVALID_PARAM | `open_pos_entry_detail[{index}]: payment_id 和 mode_of_payment 不能同时为空` | 400 | Controller 层校验，立即返回 |
| payment_id 查询失败 | QUERY_FAILED | `查询支付方式失败，payment_id: {payment_id}` | 500 | Logic 层处理，使用 `gerror.Wrapf` 包装原始错误 |
| 支付方式不存在 | NOT_FOUND | `支付方式不存在或未启用，payment_id: {payment_id}` | 404 | Logic 层检查，返回明确错误 |
| 网络超时 | TIMEOUT | `查询支付方式超时，payment_id: {payment_id}` | 504 | 由 GetModeOfPayment 内部处理 |

### 错误处理实现

```go
// Controller 层：参数校验错误
if bothEmpty {
    return rpc.ApiError(fmt.Sprintf("open_pos_entry_detail[%d]: payment_id 和 mode_of_payment 不能同时为空", i)), nil
}

// Logic 层：查询失败错误
if err != nil {
    g.Log().Error(ctx, "查询支付方式失败",
        g.Map{"payment_id": *detail.PaymentId, "error": err.Error()})
    return nil, gerror.Wrapf(err, "查询支付方式失败，payment_id: %s", *detail.PaymentId)
}

// Logic 层：支付方式不存在错误
if getModeResp == nil || getModeResp.Name == "" {
    return nil, gerror.Newf("支付方式不存在或未启用，payment_id: %s", *detail.PaymentId)
}
```

### 日志记录

```go
// 成功查询日志
g.Log().Info(ctx, "开账详情: 通过 payment_id 查询到 mode_of_payment",
    g.Map{"index": i, "payment_id": *detail.PaymentId, "mode_of_payment": modeOfPayment})

// 查询失败日志
g.Log().Error(ctx, "查询支付方式失败",
    g.Map{"payment_id": *detail.PaymentId, "error": err.Error()})
```

---

## 🧪 测试策略

### 单元测试

**文件**: `ttpos-bmp/app/ttpos-erp/internal/logic/selling/selling_test.go`

**测试用例**：

1. **参数校验测试**
   ```go
   func Test_OpenPosEntryDetail_ValidationLogic(t *testing.T)
   ```
   - 两个参数都为空
   - 两个参数都为空字符串
   - 只有 payment_id 不为空
   - 只有 mode_of_payment 不为空
   - 两个参数都不为空

2. **自动查询测试**（需要 Mock）
   - payment_id 查询成功
   - payment_id 查询失败
   - 支付方式未启用

3. **向后兼容测试**
   - 仅使用 mode_of_payment

### 集成测试

**测试场景**：

1. 使用 payment_id 开账成功
2. 使用 mode_of_payment 开账成功（向后兼容）
3. 参数都为空时返回错误
4. payment_id 无效时返回错误
5. 混合使用多个 detail

**测试工具**：grpcurl 或 BloomRPC

**参考**: `docs/shared/specs/active/task-erp-close-pos-entry-payment-id/integration-testing-guide.md`

---

## 📈 性能考虑

### 性能指标

- **GetModeOfPayment 查询时间**: < 100ms（有缓存）
- **整体接口响应时间**: 不显著增加（< 50ms）
- **缓存命中率**: > 90%（GetModeOfPayment 内部缓存）

### 优化策略

1. **利用现有缓存**
   - GetModeOfPayment 服务内部已有缓存机制
   - 无需额外缓存层

2. **错误快速返回**
   - 参数校验在 Controller 层立即返回
   - 避免不必要的服务调用

3. **并发处理**
   - 如果有多个 detail，顺序处理即可
   - 查询量通常较小（1-3 个）

---

## 🔐 安全考虑

### 安全措施

1. **参数校验**
   - 防止空值攻击
   - 防止注入攻击（Protobuf 自动处理）

2. **错误信息**
   - 不泄露敏感信息
   - 不暴露内部实现细节

3. **身份验证**
   - gRPC 接口需要身份验证（已有机制）
   - 使用现有的认证中间件

---

## 📝 文档要求

### Protobuf 注释

```protobuf
message OpenPosEntryDetail {
  optional string mode_of_payment = 1; // 支付方式，与 payment_id 二选一（必填其中之一）
  double opening_amount = 2; // 开帐金额,必填
  optional string payment_id = 3; // 支付方式唯一标识（PaymentID），与 mode_of_payment 二选一
  // 注意：当 payment_id 不为空时，系统自动调用 GetModeOfPayment 查询 mode_of_payment 值
}
```

### 代码注释

```go
// buildOpeningEntryDetails 构建开账明细
// 参数：
//   - ctx: 上下文
//   - details: 开账明细列表
//
// 返回：
//   - []erp.POSOpeningEntryDetail: 开账明细列表
//   - error: 错误信息
//
// 注意：
//   - 如果 payment_id 不为空，自动调用 GetModeOfPayment 查询 mode_of_payment
//   - 如果 payment_id 为空，直接使用 mode_of_payment（向后兼容）
func (s *sSelling) buildOpeningEntryDetails(ctx context.Context, details []*selling.OpenPosEntryDetail) ([]erp.POSOpeningEntryDetail, error)
```

---

## 🎯 验收标准

### 功能验收

- [x] Protobuf 定义包含 `payment_id` 字段
- [x] `mode_of_payment` 改为 optional
- [ ] 参数校验逻辑正确（两个参数不能同时为空）
- [ ] 自动查询功能正常（payment_id 不为空时）
- [ ] 查询失败时返回明确错误
- [ ] 向后兼容（仅 mode_of_payment 时正常工作）

### 测试验收

- [ ] 单元测试覆盖率 ≥ 80%
- [ ] 所有测试用例通过
- [ ] 手动集成测试通过

### 文档验收

- [x] Protobuf 注释完整
- [ ] 代码注释清晰
- [ ] CHANGELOG.md 已更新
- [x] 集成测试指南已创建

---

## 📚 参考资料

### 相关规范

- `ttpos-bmp/.cursor/rules/go-rules.mdc` - Go BMP 开发规范
- `ttpos-bmp/.cursor/rules/proto-rules.mdc` - Protobuf 开发规范
- `.cursor/rules/go-bmp.mdc` - Go BMP 微服务规范
- `.cursor/rules/api.mdc` - API 设计规范

### 相关实现

- `docs/shared/specs/active/task-erp-close-pos-entry-payment-id/` - ClosePosEntry PaymentID 支持（参考实现）
- `ttpos-bmp/app/ttpos-erp/internal/logic/selling/selling.go` - 现有实现

### 外部文档

- [GoFrame 官方文档](https://goframe.org)
- [Protobuf 官方文档](https://developers.google.com/protocol-buffers)

---

**版本**: v1.0.0  
**创建日期**: 2025-12-24  
**作者**: rikugun  
**审核者**: -


