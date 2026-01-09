# story-ttpos-erp-mode-of-payment-enabled / ERP 支付方式更新（SaveModeOfPayment 扩展）设计说明

## 1. 技术方案概述

- 变更文件：`ttpos-bmp/app/ttpos-erp/manifest/protobuf/selling/selling.proto`
- 核心变更：`SaveModeOfPaymentReq` 增加 `name`（可选），用于区分"创建"和"更新"操作。
- 操作语义：
  - `name` 存在 → 执行**更新**操作（根据 name 查找已有支付方式并更新指定字段）
  - `name` 不存在 → 执行**创建**操作（保持现有行为）
- 更新字段：当前支持更新 `enabled`（启用状态），预留扩展性。

## 2. 接口定义（proto 草案）

文件：`ttpos-bmp/app/ttpos-erp/manifest/protobuf/selling/selling.proto`

```protobuf
message SaveModeOfPaymentReq {
  string company_abbr = 1; // 公司简称，必填
  string branch = 2;       // 分支，必填
  string channel = 3;      // 渠道，如 LianLianPay，创建时必填
  string pay_type = 4;     // 支付类型（TTPOS 定义），创建时必填
  
  // 可选：传入时执行更新操作，未传入时执行创建操作
  optional string name = 6;
  
  // 可选：仅在明确传入时更新 ERP 启用状态
  optional bool enabled = 5;
}
```

**字段说明**：

- `name`：支付方式名称（可选）
  - **IF** 传入 **THEN** 根据 name 查找支付方式并执行更新
  - **IF** 不传 **THEN** 执行创建操作
- `enabled`：启用状态（可选）
  - **IF** 传入且 name 存在 **THEN** 更新 ERP 中该支付方式的 enabled
  - **IF** 不传或 name 不存在 **THEN** 不更新 enabled

## 3. 服务端语义与处理流程

### 3.1 SaveModeOfPayment 操作分支

```
IF req.name != nil THEN
    // 更新操作
    1. 根据 name 和 company_abbr 查找支付方式
    2. IF 支付方式不存在 THEN 返回错误"支付方式不存在"
    3. IF 支付方式不属于当前 company_abbr THEN 返回错误"无权限修改"
    4. IF req.enabled != nil THEN 更新 ERP enabled 字段
    5. 返回成功
ELSE
    // 创建操作（保持现有逻辑）
    1. 验证必填字段（channel, pay_type）
    2. 创建新支付方式
    3. 返回成功
END IF
```

### 3.2 更新操作详细流程

**步骤 1：查找支付方式**

```go
// 伪代码
func (s *sellingService) SaveModeOfPayment(ctx context.Context, req *SaveModeOfPaymentReq) error {
    if req.Name != nil && *req.Name != "" {
        // 更新操作
        return s.updateModeOfPayment(ctx, req)
    } else {
        // 创建操作（现有逻辑）
        return s.createModeOfPayment(ctx, req)
    }
}
```

**步骤 2：更新支付方式**

```go
func (s *sellingService) updateModeOfPayment(ctx context.Context, req *SaveModeOfPaymentReq) error {
    // 1. 查找支付方式
    mop, err := s.erpClient.GetModeOfPayment(ctx, req.CompanyAbbr, *req.Name)
    if err != nil {
        return errors.WithMessage(err, "支付方式不存在")
    }
    
    // 2. 权限校验：确保支付方式属于当前公司
    if !s.belongsToCompany(mop, req.CompanyAbbr) {
        return errors.New("无权限修改此支付方式")
    }
    
    // 3. 更新字段
    updateFields := make(map[string]interface{})
    if req.Enabled != nil {
        updateFields["enabled"] = *req.Enabled
    }
    
    // 4. 调用 ERP 更新接口
    if len(updateFields) > 0 {
        if err := s.erpClient.UpdateModeOfPayment(ctx, *req.Name, updateFields); err != nil {
            return errors.WithMessage(err, "更新支付方式失败")
        }
    }
    
    return nil
}
```

## 4. ERP 字段映射

- ERPNext Doctype：`Mode of Payment`
- 查找字段：`name`（支付方式名称，唯一标识）
- 更新字段：`enabled`（启用状态）
- 类型：ERP 侧可能为 int/bool（以 ERP 实际接口返回为准），服务端做必要的类型转换与校验。

## 5. 错误处理

### 5.1 更新操作错误场景

| 场景 | 错误信息 | HTTP Code |
|------|---------|-----------|
| 支付方式不存在 | "支付方式不存在" | 404 |
| 无权限修改 | "无权限修改此支付方式" | 403 |
| ERP 更新失败 | "更新支付方式失败：{原因}" | 500 |
| name 为空字符串 | "支付方式名称不能为空" | 400 |

### 5.2 创建操作错误场景（保持现有行为）

| 场景 | 错误信息 | HTTP Code |
|------|---------|-----------|
| 缺少必填字段 | "缺少必填字段：{字段名}" | 400 |
| ERP 创建失败 | "创建支付方式失败：{原因}" | 500 |

## 6. 兼容性

- **历史客户端**：不传 `name`，执行创建操作（保持原有行为）。
- **新客户端**：传入 `name`，执行更新操作。
- **部分更新**：仅传入需要更新的字段（如 `enabled`），未传入的字段不更新。

## 7. 安全性

### 7.1 权限校验

- 更新操作前必须校验支付方式是否属于当前 `company_abbr`。
- 防止越权修改其他公司的支付方式。

### 7.2 审计日志

- 记录所有更新操作：
  - 操作时间
  - 操作用户
  - 支付方式名称
  - 更新字段及新值

## 8. 测试要点

### 8.1 创建操作测试（保持现有）

- 不传 `name` → 创建新支付方式
- 验证必填字段校验
- 验证 ERP 创建成功

### 8.2 更新操作测试（新增）

- 传 `name` + `enabled=true` → ERP enabled 更新为 true
- 传 `name` + `enabled=false` → ERP enabled 更新为 false
- 传 `name` 但不传 `enabled` → ERP enabled 不变
- 传不存在的 `name` → 返回"支付方式不存在"错误
- 传其他公司的 `name` → 返回"无权限修改"错误

### 8.3 边界情况测试

- `name` 为空字符串 → 返回错误
- `name` 为 null → 执行创建操作
- 同时传 `name` 和创建必填字段 → 执行更新（忽略创建字段）

## 9. 性能优化

### 9.1 查询优化

- 更新前查询支付方式：使用 name 索引，查询时间 < 50ms
- 缓存 ERP 支付方式列表（TTL: 5分钟）

### 9.2 并发控制

- 使用分布式锁防止并发更新冲突：`lock:erp:mode_of_payment:{name}`
- 锁超时时间：10秒

## 10. 实现清单

- [ ] 更新 `selling.proto`：增加 `name` 字段（optional string）
- [ ] 生成 proto 代码并确保编译通过
- [ ] 实现 `updateModeOfPayment` 方法
- [ ] 实现权限校验逻辑
- [ ] 实现审计日志记录
- [ ] 编写单元测试（覆盖率 ≥ 80%）
- [ ] 编写集成测试（ERP 联调）
- [ ] 更新 API 文档

## 11. 扩展性预留

当前设计支持未来扩展更多可更新字段，如：

- `channel`：支付渠道
- `pay_type`：支付类型
- `custom_config`：自定义配置（JSON）

只需在 `SaveModeOfPaymentReq` 中增加对应字段，并在 `updateModeOfPayment` 中添加更新逻辑即可。
