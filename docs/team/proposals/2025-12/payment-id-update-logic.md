# 支付方式更新逻辑优化 需求提案

> 本文档用于需求发起阶段，经团队评审后创建正式 Spec。

---

## 📋 提案信息

| 项目       | 内容     |
| ---------- | -------- |
| **提案人** | rikugun   |
| **日期**   | 2025-12-23   |
| **目标版本** | v2.12.0 |
| **状态**   | ✅ 已批准   |
| **关联任务** | - |
| **关联 Spec** | [task-bmp-payment-id-update-logic](../../shared/specs/archived/v2.12/task-bmp-payment-id-update-logic/requirements.md)      |

---

## 🎯 背景和动机

### 问题描述

当前支付方式（Mode of Payment）的更新逻辑仅支持通过 `name` 字段判断是否为更新操作。但在实际业务场景中，客户端可能只知道 `payment_id`（自定义支付方式 ID），而不知道 ERP 系统中的 `name`。

**现状**：
- `SaveModeOfPayment` 接口仅在 `req.Name` 不为空时执行更新操作
- 当客户端只有 `payment_id` 时，无法执行更新操作，只能创建新的支付方式
- 导致可能出现重复的支付方式记录

**示例场景**：
> 第三方支付系统（如 Grab）同步支付方式时，只提供 `payment_id`（如 "PID123456789"），但不知道 ERP 中的 `name`（如 "Grab-CreditCard-0001 - SG"）。此时无法更新已存在的支付方式，只能创建新记录，造成数据冗余。

### 业务价值

- **避免重复数据**：防止相同 `payment_id` 的支付方式被重复创建
- **提升集成灵活性**：第三方系统只需要维护 `payment_id`，无需关心 ERP 内部的 `name` 命名规则
- **简化客户端逻辑**：客户端无需先查询 `name` 再调用更新接口，一次调用即可完成更新
- **数据一致性**：确保 `payment_id` 与支付方式的一对一关系

### 目标用户

- [x] 后端开发人员（BMP 模块）
- [x] 第三方集成系统（Grab、Foodpanda 等）
- [x] 前端开发人员（POS 端、Shop 端）

---

## 💡 解决方案概述

### 方案描述

优化支付方式更新逻辑，支持通过 `payment_id` 识别更新操作：

1. **Controller 层验证逻辑调整**：
   - 当 `req.Name` 不为空时，视为更新操作（保持现有逻辑）
   - **新增**：当 `req.PaymentId` 不为空时，也视为更新操作

2. **Logic 层更新逻辑调整**：
   - `SaveModeOfPayment` 方法根据 `req.Name` 或 `req.PaymentId` 判断是否为更新操作
   - `updateModeOfPayment` 方法优先使用 `PaymentId` 查询支付方式是否存在

3. **查询优先级**：
   - **优先使用 `PaymentId` 查询**（如果提供）
   - 如果未提供 `PaymentId` 但提供了 `Name`，使用 `Name` 查询
   - 统一使用 `List` 接口进行查询（支持 Filter）
   - 查询到记录后，执行更新操作

### 核心功能点

1. **Controller 层验证增强**：
   - `validateSaveModeOfPaymentReq` 方法识别 `PaymentId` 不为空时为更新操作
   - 更新操作时，`channel` 和 `pay_type` 不再强制必填

2. **Logic 层更新判断增强**：
   - `SaveModeOfPayment` 方法支持 `PaymentId` 判断更新操作
   - `updateModeOfPayment` 方法优先使用 `PaymentId` 查询
   - 统一使用 `List` 接口（支持按 `name` 或 `payment_id` 过滤）

3. **向后兼容**：
   - 保持现有 `Name` 更新逻辑不变
   - `PaymentId` 查询优先级高于 `Name`
   - 统一查询接口，提高代码一致性

### 影响范围

**涉及终端**：
- [x] POS 收银端
- [x] Shop 商家管理端
- [ ] KDS 厨显端
- [ ] QDS 排号叫号端
- [ ] Assistant 助手端
- [ ] Tablet 平板端
- [ ] Mobile 扫码端
- [ ] Menu 电子菜单端
- [ ] Member 会员端

**涉及模块**：
- [ ] UI 组件
- [x] API 接口（gRPC）
- [ ] 数据模型
- [x] 业务逻辑（Logic 层）
- [x] 第三方集成（Grab、Foodpanda 等）

---

## 📊 初步评估

### 技术复杂度

- [x] **低**：纯逻辑调整，无数据库结构变更
- [ ] **中**：需要前后端联调，基础业务逻辑
- [ ] **高**：涉及架构调整、第三方集成、复杂算法

### 工作量预估

- **预计天数**: 0.5 天
- **预估 SP**: 1 SP（待技术评审确认）

**工作内容**：
- Controller 层验证逻辑调整（10 分钟）
- Logic 层更新判断逻辑调整（20 分钟）
- Logic 层查询逻辑调整（20 分钟）
- 单元测试编写（30 分钟）
- 集成测试验证（30 分钟）

### 风险识别

**潜在风险**：
1. **并发冲突**：多个请求同时使用不同的 `payment_id` 更新同一支付方式
2. **数据不一致**：`Name` 和 `PaymentId` 同时提供但指向不同记录
3. **性能影响**：统一使用 `List` 查询，性能略低于 `Get` 主键查询

**缓解措施**：
1. **并发控制**：
   - 使用数据库唯一索引约束 `custom_payment_id` 字段
   - 更新操作前先查询，确认记录存在且属于当前公司
2. **参数优先级**：
   - 明确 `PaymentId` 优先级高于 `Name`（`PaymentId` 是业务主键）
   - 在文档中说明参数使用规则
3. **性能优化**：
   - 在 `custom_payment_id` 和 `name` 字段上创建索引（如未创建）
   - 查询时使用 `Limit: 1` 减少数据传输
   - 记录查询性能日志，监控慢查询

---

## 🔗 相关资源

### 参考需求

- 类似功能: 无
- 竞品分析: 无

### 相关文档

- 代码文件:
  - `ttpos-bmp/app/ttpos-erp/internal/controller/rpc/selling/selling.go`
  - `ttpos-bmp/app/ttpos-erp/internal/logic/selling/selling.go`
- 相关规范:
  - `.cursor/rules/go-bmp.mdc`
  - `.cursor/rules/go-ttpos-erp.mdc`

---

## 🤝 需求评审

### 评审参与人

| 角色         | 姓名   | 签名/日期 |
| ------------ | ------ | --------- |
| 产品经理     | -      |           |
| 技术负责人   | rikugun |           |
| 开发代表     | -      |           |
| 测试代表     | -      |           |
| UI/UX 设计师 | -      |           |

### 评审结论

- [ ] ✅ **批准**：进入技术方案设计阶段
- [ ] 🔄 **修改后批准**：需补充以下内容
- [ ] ❌ **拒绝**：不符合产品规划或优先级

**评审意见**：

```
[记录评审会议的关键讨论和决策]
```

**下一步行动**：

- [x] 创建 Spec：`task-bmp-payment-id-update-logic` ✅ 已完成 (2025-12-24)
- [x] 分配负责人：rikugun
- [x] 目标 Sprint：Sprint 当前

---

## 📝 附录

### User Story（初稿）

**作为** 第三方支付系统集成开发者  
**我想** 使用 `payment_id` 更新已存在的支付方式  
**以便于** 无需维护 ERP 内部的 `name` 字段，简化集成逻辑

### AC 验收标准（初稿）

1. **WHEN** 调用 `SaveModeOfPayment` 接口且 `req.PaymentId` 不为空 **THEN** 系统 **SHALL** 识别为更新操作
2. **WHEN** 调用 `SaveModeOfPayment` 接口且 `req.Name` 为空但 `req.PaymentId` 不为空 **THEN** 系统 **SHALL** 使用 `PaymentId` 查询支付方式并更新
3. **WHEN** 调用 `SaveModeOfPayment` 接口且 `req.Name` 和 `req.PaymentId` 都不为空 **THEN** 系统 **SHALL** 优先使用 `PaymentId` 查询并更新
4. **WHEN** 使用 `PaymentId` 查询支付方式不存在 **THEN** 系统 **SHALL** 返回错误 "支付方式不存在"
5. **WHEN** 使用 `PaymentId` 查询到的支付方式不属于当前公司 **THEN** 系统 **SHALL** 返回错误 "无权限修改此支付方式"

### 技术实现要点

#### 1. Controller 层调整

**文件**: `ttpos-bmp/app/ttpos-erp/internal/controller/rpc/selling/selling.go`

**方法**: `validateSaveModeOfPaymentReq`

**调整内容**：
```go
// 判断是更新操作还是创建操作
isUpdate := (req.Name != nil && strings.TrimSpace(*req.Name) != "") || 
            (req.PaymentId != "" && strings.TrimSpace(req.PaymentId) != "")

// 创建操作时，channel 和 pay_type 必填
// 更新操作时，channel 和 pay_type 不是必填
if !isUpdate {
    if strings.TrimSpace(req.PayType) == "" {
        return gerror.New("支付类型不能为空")
    }
}
```

#### 2. Logic 层 SaveModeOfPayment 调整

**文件**: `ttpos-bmp/app/ttpos-erp/internal/logic/selling/selling.go`

**方法**: `SaveModeOfPayment`

**调整内容**：
```go
func (s *sSelling) SaveModeOfPayment(ctx context.Context, req *selling.SaveModeOfPaymentReq) (*selling.SaveModeOfPaymentResp, error) {
    // 判断是更新操作还是创建操作
    // 如果传入了 name 或 payment_id，则执行更新操作
    if (req.Name != nil && strings.TrimSpace(*req.Name) != "") || 
       (req.PaymentId != "" && strings.TrimSpace(req.PaymentId) != "") {
        return s.updateModeOfPayment(ctx, req)
    }

    // 否则执行创建操作（保持现有逻辑）
    return s.createModeOfPayment(ctx, req)
}
```

#### 3. Logic 层 updateModeOfPayment 调整

**文件**: `ttpos-bmp/app/ttpos-erp/internal/logic/selling/selling.go`

**方法**: `updateModeOfPayment`

**调整内容**：
```go
func (s *sSelling) updateModeOfPayment(ctx context.Context, req *selling.SaveModeOfPaymentReq) (*selling.SaveModeOfPaymentResp, error) {
    var resp *gjson.Json
    var err error
    var name string
    var queryKey string

    // 构建查询过滤器
    var filters [][]string
    
    // 1. 优先使用 PaymentId 查询（业务主键）
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

    // 3. 统一使用 List 接口查询
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

---

## 📄 模板使用说明

### 何时使用此模板

- ✅ 技术团队提出改进方案
- ✅ 需要团队讨论和评审的需求
- ✅ 优化现有功能逻辑

### 与 Spec 的区别

| 阶段        | 文档类型      | 详细程度 | 用途               |
| ----------- | ------------- | -------- | ------------------ |
| **需求发起** | Proposal      | 粗略     | 团队评审、决策是否做 |
| **需求确认** | Requirements  | 详细     | User Story + AC，开发依据 |
| **技术设计** | Design        | 详细     | 技术方案，实现指导 |
| **任务分解** | Tasks         | 详细     | 开发执行，进度追踪 |

### 流转路径

```
提案 (Proposal) 
  ↓ 评审批准
需求文档 (Requirements) 
  ↓ 技术评审
设计文档 (Design) 
  ↓ SP 评估 ≤ 5
任务分解 (Tasks)
  ↓
开发实现
```

---

**版本**: v1.0.0  
**创建日期**: 2025-12-23  
**维护者**: rikugun  
**相关规范**: `.cursor/rules/go-bmp.mdc`, `.cursor/rules/go-ttpos-erp.mdc`

