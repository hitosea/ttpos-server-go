# 支付方式更新逻辑优化 设计文档

> 本文档定义支付方式更新逻辑优化的技术设计和实现方案。

## 📋 概述

优化 ttpos-erp 模块中支付方式（Mode of Payment）的更新逻辑，支持通过 `payment_id` 识别和更新已存在的支付方式。这是一个纯逻辑优化任务，不涉及数据库结构变更和新增 API，只需要调整现有 Controller 和 Logic 层的验证和查询逻辑。

**影响范围**：
- Controller 层：`ttpos-bmp/app/ttpos-erp/internal/controller/rpc/selling/selling.go`
- Logic 层：`ttpos-bmp/app/ttpos-erp/internal/logic/selling/selling.go`

---

## 🎯 规范对齐

### Go BMP 规范 (go-bmp.mdc)

本设计遵循 GoFrame 2.x 开发规范：

- ✅ 使用 GoFrame 2.x 框架
- ✅ 使用 gerror 包进行错误处理
- ✅ 错误信息使用中文
- ✅ Controller 层响应通过 `erp.ApiResponse` 包装
- ✅ Logic 层返回具体业务数据类型（不返回 `ApiResponse`）
- ✅ 使用 g.Log() 记录审计日志

### ttpos-erp 子模块规范 (go-ttpos-erp.mdc)

- ✅ Logic 层复用已有逻辑，避免重复实现
- ✅ 返回参数类型不能是 `erp.ApiResponse`
- ✅ 与 erpnext 交互使用 service.Document() 服务

---

## 🔄 代码复用分析

### 可复用的现有组件

- **Document Service**: `ttpos-bmp/app/ttpos-erp/internal/service` - 用于查询和更新 Mode of Payment
  - `service.Document().List()` - 查询支付方式
  - `service.Document().Update()` - 更新支付方式
  
- **Company Service**: `ttpos-bmp/app/ttpos-erp/internal/service` - 用于查询公司信息
  - `service.Company().GetCompanyNameWithAbbr()` - 根据公司缩写查询公司名称

### 集成点

- **现有 API**: `SaveModeOfPayment` gRPC 接口
  - 新功能直接集成到现有接口中
  - 保持向后兼容，不破坏现有调用方

- **现有逻辑**: `createModeOfPayment` 和 `updateModeOfPayment` 方法
  - 复用现有创建逻辑
  - 扩展更新逻辑，支持 `PaymentId` 查询

---

## 🏗️ 架构设计

### 分层设计原则

**GoFrame 架构**:

```
Controller 层 (RPC)
  ↓ 参数验证
Logic 层
  ↓ 业务逻辑
Service 层 (Document/Company)
  ↓ gRPC调用
ERPNext 系统
```

**依赖规则**:

- ✅ Controller 调用 Logic
- ✅ Logic 调用 Service
- ✅ Service 调用 ERPNext API

### 模块划分

#### Go BMP 模块

- **RPC Controller**: `ttpos-bmp/app/ttpos-erp/internal/controller/rpc/selling/selling.go`
  - 方法：`SaveModeOfPayment` (已存在，调整验证逻辑)
  - 方法：`validateSaveModeOfPaymentReq` (已存在，扩展验证规则)

- **Logic 层**: `ttpos-bmp/app/ttpos-erp/internal/logic/selling/selling.go`
  - 方法：`SaveModeOfPayment` (已存在，调整路由逻辑)
  - 方法：`updateModeOfPayment` (已存在，优化查询逻辑)
  - 方法：`createModeOfPayment` (已存在，保持不变)

---

## 🔌 逻辑设计

### Controller 层验证逻辑

#### 调整前（现有逻辑）

```go
func (c *Controller) validateSaveModeOfPaymentReq(req *selling.SaveModeOfPaymentReq) error {
    // 判断是更新操作还是创建操作
    isUpdate := req.Name != nil && strings.TrimSpace(*req.Name) != ""

    // 创建操作时，channel 和 pay_type 必填
    if !isUpdate {
        if strings.TrimSpace(req.PayType) == "" {
            return gerror.New("支付类型不能为空")
        }
    }
    
    return nil
}
```

#### 调整后（新逻辑）

```go
func (c *Controller) validateSaveModeOfPaymentReq(req *selling.SaveModeOfPaymentReq) error {
    // 判断是更新操作还是创建操作
    // ✅ 新增：支持通过 PaymentId 识别更新操作
    isUpdate := (req.Name != nil && strings.TrimSpace(*req.Name) != "") || 
                (req.PaymentId != "" && strings.TrimSpace(req.PaymentId) != "")

    // 创建操作时，channel 和 pay_type 必填
    // 更新操作时，channel 和 pay_type 不是必填
    if !isUpdate {
        if strings.TrimSpace(req.PayType) == "" {
            return gerror.New("支付类型不能为空")
        }
    }
    
    return nil
}
```

**变更说明**：
- 新增 `PaymentId` 判断条件，当 `PaymentId` 不为空时也识别为更新操作
- 更新操作时，`channel` 和 `pay_type` 不再强制必填

---

### Logic 层路由逻辑

#### 调整前（现有逻辑）

```go
func (s *sSelling) SaveModeOfPayment(ctx context.Context, req *selling.SaveModeOfPaymentReq) (*selling.SaveModeOfPaymentResp, error) {
    // 判断是更新操作还是创建操作
    // 如果传入了 name，则执行更新操作
    if req.Name != nil && strings.TrimSpace(*req.Name) != "" {
        return s.updateModeOfPayment(ctx, req)
    }

    // 否则执行创建操作（保持现有逻辑）
    return s.createModeOfPayment(ctx, req)
}
```

#### 调整后（新逻辑）

```go
func (s *sSelling) SaveModeOfPayment(ctx context.Context, req *selling.SaveModeOfPaymentReq) (*selling.SaveModeOfPaymentResp, error) {
    // 判断是更新操作还是创建操作
    // ✅ 新增：如果传入了 name 或 payment_id，则执行更新操作
    if (req.Name != nil && strings.TrimSpace(*req.Name) != "") || 
       (req.PaymentId != "" && strings.TrimSpace(req.PaymentId) != "") {
        return s.updateModeOfPayment(ctx, req)
    }

    // 否则执行创建操作（保持现有逻辑）
    return s.createModeOfPayment(ctx, req)
}
```

**变更说明**：
- 新增 `PaymentId` 判断条件，支持通过 `PaymentId` 路由到更新操作
- 保持向后兼容，现有 `Name` 更新逻辑不变

---

### Logic 层查询逻辑

#### 调整前（现有逻辑）

```go
func (s *sSelling) updateModeOfPayment(ctx context.Context, req *selling.SaveModeOfPaymentReq) (*selling.SaveModeOfPaymentResp, error) {
    name := strings.TrimSpace(*req.Name)

    // 1. 查询支付方式是否存在
    resp, err := service.Document().Get(ctx, &erp.ErpReq{
        DocType: erp.DocTypeModeOfPayment,
        Name:    name,
    }, &erp.RequestParams{
        Fields: []string{"name", "custom_company", "custom_branch", "enabled", "custom_payment_id"},
    })
    if err != nil {
        if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
            return nil, gerror.Newf("支付方式 [%s] 不存在", name)
        }
        return nil, gerror.Wrapf(err, "查询支付方式失败")
    }

    // 2. 权限校验：确认支付方式属于当前公司
    companyName, err := service.Company().GetCompanyNameWithAbbr(ctx, req.CompanyAbbr)
    if err != nil {
        return nil, gerror.Wrapf(err, "根据公司缩写[%s]查询公司失败", req.CompanyAbbr)
    }

    erpCompany := resp.Get("data.custom_company").String()
    if erpCompany != companyName {
        g.Log().Warningf(ctx, "尝试越权修改支付方式：name=%s, 请求公司=%s, ERP公司=%s",
            name, companyName, erpCompany)
        return nil, gerror.New("无权限修改此支付方式")
    }

    // 3. 构建更新数据
    updateData := g.Map{}
    if req.Enabled != nil {
        if req.GetEnabled() {
            updateData["enabled"] = 1
        } else {
            updateData["enabled"] = 0
        }
    }
    if req.PaymentId != "" {
        updateData["custom_payment_id"] = req.PaymentId
    }

    // 4. 执行更新
    if len(updateData) > 0 {
        _, err = service.Document().Update(ctx, &erp.ErpReq{
            DocType: erp.DocTypeModeOfPayment,
            Name:    name,
        }, updateData)
        if err != nil {
            return nil, gerror.Wrapf(err, "更新支付方式失败")
        }

        g.Log().Infof(ctx, "更新支付方式成功：name=%s, company=%s, branch=%s, updateData=%v",
            name, req.CompanyAbbr, req.Branch, updateData)
    } else {
        g.Log().Infof(ctx, "更新支付方式：未传入任何可更新字段，跳过更新：name=%s", name)
    }

    finalPaymentID := resp.Get("data.custom_payment_id").String()
    if req.PaymentId != "" {
        finalPaymentID = req.PaymentId
    }

    return &selling.SaveModeOfPaymentResp{
        Name:      name,
        PaymentId: finalPaymentID,
    }, nil
}
```

#### 调整后（新逻辑）

```go
func (s *sSelling) updateModeOfPayment(ctx context.Context, req *selling.SaveModeOfPaymentReq) (*selling.SaveModeOfPaymentResp, error) {
    var resp *gjson.Json
    var err error
    var name string
    var queryKey string

    // 构建查询过滤器
    var filters [][]string
    
    // 1. ✅ 优先使用 PaymentId 查询（业务主键）
    if req.PaymentId != "" && strings.TrimSpace(req.PaymentId) != "" {
        paymentId := strings.TrimSpace(req.PaymentId)
        queryKey = fmt.Sprintf("payment_id=%s", paymentId)
        filters = [][]string{{"custom_payment_id", "=", paymentId}}
        g.Log().Infof(ctx, "[updateModeOfPayment] 通过 payment_id 查询支付方式: %s", queryKey)
    } else if req.Name != nil && strings.TrimSpace(*req.Name) != "" {
        // 2. 使用 Name 查询（ERP 主键）
        name = strings.TrimSpace(*req.Name)
        queryKey = fmt.Sprintf("name=%s", name)
        filters = [][]string{{"name", "=", name}}
        g.Log().Infof(ctx, "[updateModeOfPayment] 通过 name 查询支付方式: %s", queryKey)
    } else {
        return nil, gerror.New("name 或 payment_id 至少提供一个")
    }

    // 3. ✅ 统一使用 List 接口查询
    resp, err = service.Document().List(ctx, &erp.ErpReq{
        DocType: erp.DocTypeModeOfPayment,
    }, &erp.RequestParams{
        Fields:  []string{"name", "custom_company", "custom_branch", "enabled", "custom_payment_id"},
        Filters: filters,
        Limit:   1,
    })
    if err != nil {
        g.Log().Errorf(ctx, "[updateModeOfPayment] 查询支付方式失败: %s, err=%v", queryKey, err)
        return nil, gerror.Wrapf(err, "查询支付方式失败")
    }

    // 4. 检查查询结果
    dataArray := resp.GetJsons("data")
    if len(dataArray) == 0 {
        g.Log().Warningf(ctx, "[updateModeOfPayment] 支付方式不存在: %s", queryKey)
        return nil, gerror.Newf("支付方式不存在: %s", queryKey)
    }

    // 5. 获取查询到的支付方式信息
    data := dataArray[0]
    name = data.Get("name").String()
    erpCompany := data.Get("custom_company").String()

    // 6. 权限校验：确认支付方式属于当前公司
    companyName, err := service.Company().GetCompanyNameWithAbbr(ctx, req.CompanyAbbr)
    if err != nil {
        return nil, gerror.Wrapf(err, "根据公司缩写[%s]查询公司失败", req.CompanyAbbr)
    }

    if erpCompany != companyName {
        g.Log().Warningf(ctx, "[updateModeOfPayment] 尝试越权修改支付方式: name=%s, 请求公司=%s, ERP公司=%s",
            name, companyName, erpCompany)
        return nil, gerror.New("无权限修改此支付方式")
    }

    // 7. 构建更新数据
    updateData := g.Map{}

    // 仅在明确传入 enabled 时才更新
    if req.Enabled != nil {
        if req.GetEnabled() {
            updateData["enabled"] = 1
        } else {
            updateData["enabled"] = 0
        }
    }

    // 仅在明确传入 payment_id 时才更新
    if req.PaymentId != "" {
        updateData["custom_payment_id"] = req.PaymentId
    }

    // 8. 如果有字段需要更新，则调用 ERP 更新接口
    if len(updateData) > 0 {
        _, err = service.Document().Update(ctx, &erp.ErpReq{
            DocType: erp.DocTypeModeOfPayment,
            Name:    name,
        }, updateData)
        if err != nil {
            return nil, gerror.Wrapf(err, "更新支付方式失败")
        }

        // 9. 记录审计日志
        g.Log().Infof(ctx, "[updateModeOfPayment] 更新成功: name=%s, company=%s, branch=%s, updateData=%v",
            name, req.CompanyAbbr, req.Branch, updateData)
    } else {
        g.Log().Infof(ctx, "[updateModeOfPayment] 未传入任何可更新字段，跳过更新: name=%s", name)
    }

    // 10. 读取更新后的 payment_id（优先使用更新值，否则使用原值）
    finalPaymentID := data.Get("custom_payment_id").String()
    if req.PaymentId != "" {
        finalPaymentID = req.PaymentId
    }

    return &selling.SaveModeOfPaymentResp{
        Name:      name,
        PaymentId: finalPaymentID,
    }, nil
}
```

**关键变更说明**：

1. **查询优先级调整**：
   - ✅ 优先使用 `PaymentId` 查询（业务主键）
   - ✅ 其次使用 `Name` 查询（ERP 主键）

2. **统一查询接口**：
   - ✅ 统一使用 `List` 接口（支持 Filter）
   - ✅ 查询时使用 `Limit: 1` 减少数据传输

3. **增强日志记录**：
   - ✅ 记录查询键值（`payment_id` 或 `name`）
   - ✅ 记录查询方式和结果
   - ✅ 记录更新操作和数据

4. **保持向后兼容**：
   - ✅ `Name` 查询逻辑仍然有效
   - ✅ 权限校验逻辑不变
   - ✅ 更新逻辑不变

---

## 🚨 错误处理

### 错误场景

#### 场景 1: 支付方式不存在

- **处理方式**: 查询结果为空时，返回错误
- **用户影响**: 收到错误提示 "支付方式不存在: payment_id=xxx"
- **代码示例**:
  ```go
  if len(dataArray) == 0 {
      g.Log().Warningf(ctx, "[updateModeOfPayment] 支付方式不存在: %s", queryKey)
      return nil, gerror.Newf("支付方式不存在: %s", queryKey)
  }
  ```

#### 场景 2: 无权限修改

- **处理方式**: 权限校验失败时，返回错误
- **用户影响**: 收到错误提示 "无权限修改此支付方式"
- **代码示例**:
  ```go
  if erpCompany != companyName {
      g.Log().Warningf(ctx, "[updateModeOfPayment] 尝试越权修改支付方式: name=%s, 请求公司=%s, ERP公司=%s",
          name, companyName, erpCompany)
      return nil, gerror.New("无权限修改此支付方式")
  }
  ```

#### 场景 3: 查询失败

- **处理方式**: ERP 接口调用失败时，记录错误日志并返回
- **用户影响**: 收到错误提示 "查询支付方式失败"
- **代码示例**:
  ```go
  if err != nil {
      g.Log().Errorf(ctx, "[updateModeOfPayment] 查询支付方式失败: %s, err=%v", queryKey, err)
      return nil, gerror.Wrapf(err, "查询支付方式失败")
  }
  ```

---

## 🔒 安全设计

### 权限控制

- **公司隔离**: 确认支付方式属于当前公司，防止越权修改
- **审计日志**: 记录所有查询和更新操作，便于追溯

### 数据安全

- **参数验证**: Controller 层严格验证参数
- **SQL 注入防护**: 使用 GoFrame 的参数化查询
- **错误信息**: 不暴露敏感的系统信息

---

## 🧪 测试策略

### 单元测试

**目标覆盖率**:

- Logic 层: 80%+

**测试内容**:

1. **Controller 层验证逻辑**:
   - 测试 `Name` 不为空时识别为更新操作
   - 测试 `PaymentId` 不为空时识别为更新操作
   - 测试创建操作时 `pay_type` 必填
   - 测试更新操作时 `pay_type` 非必填

2. **Logic 层路由逻辑**:
   - 测试 `Name` 不为空时路由到更新方法
   - 测试 `PaymentId` 不为空时路由到更新方法
   - 测试 `Name` 和 `PaymentId` 都为空时路由到创建方法

3. **Logic 层查询逻辑**:
   - 测试优先使用 `PaymentId` 查询
   - 测试使用 `Name` 查询
   - 测试查询结果为空时返回错误
   - 测试权限校验失败时返回错误
   - 测试查询成功且权限通过时执行更新

**示例**:

```go
// ttpos-bmp/app/ttpos-erp/internal/logic/selling/selling_test.go
func TestUpdateModeOfPayment_WithPaymentId(t *testing.T) {
    // 测试使用 PaymentId 查询和更新
}

func TestUpdateModeOfPayment_WithName(t *testing.T) {
    // 测试使用 Name 查询和更新
}

func TestUpdateModeOfPayment_NotFound(t *testing.T) {
    // 测试支付方式不存在
}

func TestUpdateModeOfPayment_PermissionDenied(t *testing.T) {
    // 测试权限校验失败
}
```

### 集成测试

**测试流程**:

1. **创建测试数据**: 创建测试支付方式
2. **使用 PaymentId 更新**: 调用接口，验证更新成功
3. **使用 Name 更新**: 调用接口，验证更新成功
4. **越权测试**: 尝试更新其他公司的支付方式，验证被拒绝
5. **不存在测试**: 使用不存在的 `PaymentId`，验证返回错误

---

## 📈 性能优化

### 优化策略

1. **查询优化**:
   - 统一使用 `List` 接口（支持 Filter）
   - 查询时使用 `Limit: 1` 减少数据传输
   - 确保 `custom_payment_id` 和 `name` 字段有索引

2. **日志优化**:
   - 使用结构化日志（g.Log()）
   - 记录关键操作和查询键值
   - 便于性能监控和问题排查

### 性能指标

- 查询响应时间: < 100ms
- 更新响应时间: < 200ms

---

## 📚 实现清单

### Phase 1: Controller 层调整

- [ ] 修改 `validateSaveModeOfPaymentReq` 方法
- [ ] 编写 Controller 层单元测试

### Phase 2: Logic 层调整

- [ ] 修改 `SaveModeOfPayment` 方法（路由逻辑）
- [ ] 修改 `updateModeOfPayment` 方法（查询逻辑）
- [ ] 编写 Logic 层单元测试

### Phase 3: 测试和验证

- [ ] 单元测试
- [ ] 集成测试
- [ ] 性能测试

**详细任务**: 参见 `tasks.md`

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/2025-12/2025-12-24.md`
- 在技术方案评审、关键架构决策或踩坑总结后，立即补充 Episode 并在设计文档尾部互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-12-24  
**作者**: rikugun  
**审核者**: rikugun

