# 开账接口支持 PaymentID 设计文档

> 本文档定义开账接口支持 PaymentID 的技术设计和实现方案。

## 📋 概述

在 `OpenPosEntry` 接口中增加 `payment_id` 参数支持，当调用方提供 `payment_id` 时，系统自动查询对应的 `mode_of_payment` 完成开账操作。该功能与已完成的 `ClosePosEntry` 接口 PaymentID 支持功能保持完全一致的技术实现方案。

**系统位置**: ttpos-bmp (Go BMP 微服务) - ERP 销售模块

---

## 🎯 规范对齐

### Go BMP 规范 (go-bmp.mdc)

该设计严格遵循 Go BMP 开发规范：

- ✅ **分层架构**: Controller → Logic → Service → Repository
- ✅ **依赖注入**: 通过 Service 接口调用依赖服务
- ✅ **gRPC 服务**: 注册到 Nacos，遵循微服务规范
- ✅ **错误处理**: 使用 `gerror` 包，错误信息使用中文
- ✅ **日志记录**: 使用 GoFrame Logger，记录关键操作
- ✅ **代码生成**: 禁止修改 `dao/entity/do/` 目录

### API 设计规范 (api.mdc)

- ✅ **gRPC 接口**: 遵循 Protobuf 规范，字段命名使用 snake_case
- ✅ **响应包装**: Controller 层负责将业务数据包装为 `erp.ResponseInfo`
- ✅ **错误信息**: 使用中文，便于运维和调试
- ✅ **字段验证**: 在 Controller 层进行参数校验

### 数据库规范 (database.mdc)

- ✅ **无变更**: 本功能不涉及数据库表结构变更
- ✅ **复用表**: 使用现有的 `Mode of Payment` 相关表
- ✅ **缓存优化**: 依赖 `GetModeOfPayment` 服务的缓存机制

### 安全规范 (security.mdc)

- ✅ **身份验证**: gRPC 接口已有身份验证机制
- ✅ **参数校验**: 防止无效输入，校验 `payment_id` 和 `mode_of_payment` 的有效性
- ✅ **错误信息**: 不泄露敏感数据，只返回必要的错误上下文

---

## 🏗️ 架构设计

### 系统架构

```
调用方 (Main/Admin)
    ↓ (gRPC)
ttpos-bmp Controller
    ↓ (参数校验)
ttpos-bmp Logic
    ↓ (业务逻辑 + 自动查询)
GetModeOfPayment Service (缓存)
    ↓
ERPNext API
    ↓
数据库 (Mode of Payment)
```

### 组件关系

| 组件 | 职责 | 依赖关系 |
|------|------|----------|
| **Controller** | 参数校验、响应包装 | Logic 层 |
| **Logic** | 业务逻辑、自动查询 | Service 层 |
| **Service** | 支付方式查询服务 | ERPNext API |
| **Repository** | 数据访问 | 数据库 |

### 数据流

1. **请求到达**: Controller 接收 gRPC 请求
2. **参数校验**: Controller 校验 `payment_id` 和 `mode_of_payment` 参数
3. **业务处理**: Logic 处理开账业务逻辑
4. **自动查询**: 如提供 `payment_id`，调用 `GetModeOfPayment` 服务
5. **ERP 操作**: 使用查询到的 `mode_of_payment` 调用 ERPNext API
6. **响应返回**: Controller 包装响应返回给调用方

---

## 💾 数据模型

### Protobuf 定义

```protobuf
// OpenPosEntryDetail 开账明细
message OpenPosEntryDetail {
  optional string mode_of_payment = 1; // 支付方式，与 payment_id 二选一（必填其中之一）
  double opening_amount = 2; // 开帐金额,必填
  optional string payment_id = 3; // 支付方式唯一标识（PaymentID），与 mode_of_payment 二选一
  // 注意：当 payment_id 不为空时，系统自动调用 GetModeOfPayment 查询 mode_of_payment 值
}
```

### Go DTO 定义

```go
// OpenPosEntryDetail 开账明细 DTO
type OpenPosEntryDetail struct {
    ModeOfPayment *string  `json:"mode_of_payment,omitempty"` // 支付方式
    OpeningAmount float64  `json:"opening_amount"`            // 开账金额
    PaymentId     *string  `json:"payment_id,omitempty"`      // 支付方式 ID
}
```

### 关键变更

| 字段 | 原类型 | 新类型 | 说明 |
|------|--------|--------|------|
| `mode_of_payment` | `string` | `optional string` | 支持可选，提供向后兼容性 |
| `payment_id` | 不存在 | `optional string` (字段编号 3) | 新增字段，支持 PaymentID 查询 |

---

## 🔌 API 设计

### gRPC 接口

**接口名称**: `OpenPosEntry`

**请求消息**: `OpenPosEntryReq`
```protobuf
message OpenPosEntryReq {
  string pos_profile_name = 1; // Pos Profile名称,必填
  string cashier_email = 2;    // 收银员邮箱,必填
  string company_abbr = 3;     // 公司缩写,必填
  int64 period_start_date = 5; // 周期开始时间,必填
  repeated OpenPosEntryDetail open_pos_entry_detail = 6; // 开帐详情,必填
  string branch = 7; // 分店名称,可选
}
```

**响应消息**: `OpenPosEntryResp`
```protobuf
message OpenPosEntryResp {
  string open_pos_entry_name = 1; // 开帐的Pos Profile名称
}
```

### 参数校验规则

1. **必填参数**:
   - `pos_profile_name`: POS Profile 名称
   - `cashier_email`: 收银员邮箱
   - `company_abbr`: 公司缩写
   - `period_start_date`: 周期开始时间
   - `open_pos_entry_detail`: 开账明细列表（至少一个）

2. **OpenPosEntryDetail 参数校验**:
   - `payment_id` 和 `mode_of_payment` 不能同时为空
   - 允许同时提供两个参数（优先使用 `payment_id`）
   - `opening_amount` 必须 ≥ 0

### 错误处理

| 错误场景 | 错误码 | 错误信息示例 |
|----------|--------|--------------|
| 两个参数都为空 | PARAMETER_ERROR | `open_pos_entry_detail[0]: payment_id 和 mode_of_payment 不能同时为空` |
| payment_id 查询失败 | SERVICE_ERROR | `查询支付方式失败，payment_id: PID123456` |
| 支付方式不存在 | NOT_FOUND | `支付方式不存在或未启用，payment_id: PID123456` |
| ERP 调用失败 | ERP_ERROR | `创建开账记录失败: {ERP错误信息}` |

---

## 🔧 实现方案

### Phase 1: Protobuf 定义调整

**文件**: `ttpos-bmp/app/ttpos-erp/manifest/protobuf/selling/selling.proto`

**变更内容**:
```protobuf
// 修改前
message OpenPosEntryDetail {
  string mode_of_payment = 1; // 支付方式
  double opening_amount = 2; // 开帐金额
}

// 修改后
message OpenPosEntryDetail {
  optional string mode_of_payment = 1; // 支付方式，与 payment_id 二选一（必填其中之一）
  double opening_amount = 2; // 开帐金额,必填
  optional string payment_id = 3; // 支付方式唯一标识（PaymentID），与 mode_of_payment 二选一
  // 注意：当 payment_id 不为空时，系统自动调用 GetModeOfPayment 查询 mode_of_payment 值
}
```

**命令**: `cd ttpos-bmp/app/ttpos-erp && gf gen pb`

### Phase 2: Controller 层参数校验

**文件**: `ttpos-bmp/app/ttpos-erp/internal/controller/rpc/selling/selling.go`

**实现逻辑**:
```go
// OpenPosEntry 开账
func (*Controller) OpenPosEntry(ctx context.Context, req *selling.OpenPosEntryReq) (*api.ResponseInfo, error) {
    // 验证开账详情
    for i, detail := range req.OpenPosEntryDetail {
        // 参数校验：payment_id 和 mode_of_payment 不能同时为空
        if (detail.PaymentId == nil || strings.TrimSpace(*detail.PaymentId) == "") &&
            (detail.ModeOfPayment == nil || strings.TrimSpace(*detail.ModeOfPayment) == "") {
            return rpc.ApiError(fmt.Sprintf("open_pos_entry_detail[%d]: payment_id 和 mode_of_payment 不能同时为空", i)), nil
        }
        // 其他校验逻辑...
    }

    // 调用 Logic 层
    resp, err := service.Selling().OpenPosEntry(ctx, req)
    // ...
}
```

### Phase 3: Logic 层自动查询

**文件**: `ttpos-bmp/app/ttpos-erp/internal/logic/selling/selling.go`

**核心逻辑**: 参考已实现的 `buildClosingEntryDetails` 方法，创建类似的 `buildOpeningEntryDetails` 方法。

```go
// buildOpeningEntryDetails 构建开账明细
func (s *sSelling) buildOpeningEntryDetails(ctx context.Context, details []*selling.OpenPosEntryDetail) ([]erp.POSOpeningEntryDetail, error) {
    var openingDetails []erp.POSOpeningEntryDetail
    
    for i, detail := range details {
        var modeOfPayment string
        
        // 如果提供了 payment_id，自动查询 mode_of_payment
        if detail.PaymentId != nil && *detail.PaymentId != "" {
            getModeResp, err := service.Selling().GetModeOfPayment(ctx, &selling.GetModeOfPaymentReq{
                PaymentId: detail.PaymentId,
            })
            if err != nil {
                g.Log().Error(ctx, "查询支付方式失败", g.Map{"payment_id": *detail.PaymentId, "error": err.Error()})
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
        
        openingDetails = append(openingDetails, erp.POSOpeningEntryDetail{
            ModeOfPayment: modeOfPayment,
            OpeningAmount: detail.OpeningAmount,
        })
    }
    
    return openingDetails, nil
}
```

### Phase 4: 集成现有逻辑

**文件**: `ttpos-bmp/app/ttpos-erp/internal/logic/selling/selling.go`

**OpenPosEntry 方法更新**:
```go
func (s *sSelling) OpenPosEntry(ctx context.Context, req *selling.OpenPosEntryReq) (*selling.OpenPosEntryResp, error) {
    // 构建开账明细（支持 payment_id 自动查询 mode_of_payment）
    openingDetails, err := s.buildOpeningEntryDetails(ctx, req.OpenPosEntryDetail)
    if err != nil {
        return nil, err
    }

    // 构建开账请求（使用查询到的 mode_of_payment）
    reqInfo := s.buildOpeningEntryRequest(req, openingDetails)
    
    // 后续逻辑保持不变...
}
```

---

## 🧪 测试策略

### 单元测试

**文件**: `ttpos-bmp/app/ttpos-erp/internal/logic/selling/selling_test.go`

**测试场景**:
1. **参数校验测试**: 验证两个参数同时为空时的错误处理
2. **自动查询测试**: Mock `GetModeOfPayment` 服务，测试成功和失败场景
3. **向后兼容测试**: 仅提供 `mode_of_payment` 的正常处理
4. **优先级测试**: 同时提供两个参数时的处理逻辑

**覆盖率目标**: Logic 层 ≥ 80%

### 集成测试

**工具**: grpcurl / BloomRPC

**测试用例**:
1. 使用 `payment_id` 开账（成功）
2. 使用 `mode_of_payment` 开账（向后兼容）
3. 两个参数都为空（错误）
4. `payment_id` 不存在（错误）
5. 同时提供两个参数（优先使用 `payment_id`）

### 验收标准

1. **功能测试**: 所有测试场景通过
2. **性能测试**: 接口响应时间 < 500ms
3. **兼容性测试**: 现有调用方式不受影响

---

## 📊 性能评估

### 性能影响分析

| 场景 | 影响 | 说明 |
|------|------|------|
| 仅使用 `mode_of_payment` | 无 | 保持原有性能 |
| 使用 `payment_id` | +50-100ms | 单次查询开销 |
| 批量开账 | 线性增加 | 按 detail 数量线性增加 |

### 优化措施

1. **缓存机制**: 依赖 `GetModeOfPayment` 服务的内部缓存
2. **并发查询**: 如有性能需求，可考虑并发查询多个 `payment_id`
3. **批量查询**: 未来可扩展为批量查询 API

### 监控指标

- 接口响应时间 (P95 < 500ms)
- 查询成功率 (> 99%)
- 缓存命中率 (> 90%)

---

## 🔄 部署方案

### 部署步骤

1. **代码部署**: 部署 ttpos-bmp 服务新版本
2. **服务重启**: 重启 ttpos-bmp 服务
3. **Nacos 更新**: 服务注册信息自动更新
4. **调用方更新**: 可逐步更新调用方代码（向后兼容）

### 回滚方案

1. **快速回滚**: 部署上一版本
2. **配置开关**: 可考虑添加功能开关控制 `payment_id` 支持
3. **监控告警**: 部署后 24 小时内重点监控

### 验证检查

- [ ] 服务启动成功
- [ ] Nacos 服务注册正常
- [ ] 现有调用方式正常工作
- [ ] 新功能调用正常工作

---

## 📈 扩展性考虑

### 未来扩展

1. **批量查询**: 实现 `GetModeOfPayments` 批量查询 API
2. **缓存增强**: 增加应用层缓存
3. **配置开关**: 支持动态开启/关闭 `payment_id` 功能

### 技术债务

- **代码复用**: `buildOpeningEntryDetails` 与 `buildClosingEntryDetails` 逻辑相似，可考虑提取公共方法
- **错误处理**: 统一错误码定义
- **监控指标**: 完善性能监控

---

## 🔗 相关文档

### 参考实现

- `docs/shared/specs/active/task-erp-close-pos-entry-payment-id/design.md` - 关账接口 PaymentID 支持设计（已完成，可直接参考）

### 核心规范

- `ttpos-bmp/.cursor/rules/go-rules.mdc` - Go BMP 开发规范
- `ttpos-bmp/.cursor/rules/proto-rules.mdc` - Protobuf 开发规范
- `.cursor/rules/go-bmp.mdc` - Go BMP 微服务规范

### 相关代码

- `ttpos-bmp/app/ttpos-erp/manifest/protobuf/selling/selling.proto` - Protobuf 定义
- `ttpos-bmp/app/ttpos-erp/internal/logic/selling/selling.go` - Logic 实现
- `ttpos-bmp/app/ttpos-erp/internal/controller/rpc/selling/selling.go` - Controller 实现

---

**模板版本**: v1.0.0  
**创建日期**: 2025-12-24  
**作者**: rikugun  
**审核者**: -
