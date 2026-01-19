# 敏感操作记录显示授权人信息 需求提案

> 本文档用于需求发起阶段，经团队评审后创建正式 Spec。

---

## 📋 提案信息

| 项目          | 内容                                                                                                           |
| ------------- | -------------------------------------------------------------------------------------------------------------- |
| **提案人**    | xiezhihuan                                                                                                     |
| **日期**      | 2025-01-16                                                                                                     |
| **目标版本**  | v2.10.0                                                                                                        |
| **状态**      | ✅ 已实现                                                                                                      |
| **关联任务**  | -                                                                                                              |
| **关联 Spec** | [story-pos-sensitive-operation-settings](../../../shared/specs/archived/v2.12/story-pos-sensitive-operation-settings/) |

---

## 🎯 背景和动机

### 问题描述

当前在订单详情的操作记录中，所有操作都显示的是**操作人**（实际执行操作的员工）的信息（姓名和邮箱）。但对于敏感操作（折扣/退款/免单），如果使用了授权验证，操作记录中应该显示**授权人**的信息，而不是操作人的信息，这样才能准确追溯是谁授权了这次敏感操作。

**当前问题**：
- 操作记录中显示的是操作人信息（如：`张三 (zhangsan@example.com)`）
- 无法区分是操作人自己执行的，还是通过授权验证后执行的
- 审计时无法准确追溯授权人

**用户故事**：

> 作为门店经理/店长，我想在订单详情的操作记录中看到授权人的信息（而不是操作人信息），以便于准确追溯敏感操作的授权情况，进行财务审计。

### 业务价值

- ✅ **准确追溯**：操作记录中显示授权人信息，而非操作人信息，便于审计
- ✅ **责任明确**：明确显示是谁授权了这次敏感操作
- ✅ **符合审计要求**：满足财务审计对敏感操作追溯的要求
- ✅ **提升管理透明度**：门店经理可以清楚看到每次敏感操作的授权情况

### 目标用户

- [x] 门店经理/店长（查看操作记录进行审计）
- [x] 财务人员（审计敏感操作）
- [x] 系统管理员（追溯操作责任）

---

## 💡 解决方案概述

### 方案描述

在订单详情的操作记录中，当敏感操作（折扣/退款/免单）使用了授权验证时，显示**授权人**的名字和邮箱，而不是原来的操作人信息。如果敏感操作没有使用授权验证（操作人本身在授权名单中），则继续显示操作人信息。

**技术实现思路**：

1. 在操作记录的数据结构（`data` JSON 字段）中，已经存储了授权员工信息（`authorized_staff`）
2. 修改操作记录查询接口，检查是否存在授权员工信息
3. 如果存在授权员工信息，返回授权人的姓名和邮箱
4. 如果不存在授权员工信息，返回操作人的姓名和邮箱（保持原有逻辑）
5. 前端显示逻辑保持不变，直接显示接口返回的操作人信息

### 核心功能点

1. **操作记录数据结构检查**
   - 检查操作记录的 `data` 字段中是否存在 `authorized_staff` 信息
   - 如果存在，说明使用了授权验证

2. **操作人信息返回逻辑**
   - **IF** 操作记录中存在 `authorized_staff` 信息 **THEN** 返回授权人的姓名和邮箱
   - **ELSE** 返回操作人的姓名和邮箱（原有逻辑）

3. **前端显示**
   - 前端无需修改，直接显示接口返回的操作人信息
   - 显示格式：`授权人姓名 (授权人邮箱)` 或 `操作人姓名 (操作人邮箱)`

4. **适用范围**
   - 仅适用于敏感操作（折扣/退款/免单）
   - 其他操作继续显示操作人信息

### 影响范围

**涉及终端**：

- [x] **Shop 商家管理端**：订单详情页的操作记录显示
- [x] **POS 收银端**：订单详情页的操作记录显示（如有）
- [x] **Assistant 助手端**：订单详情页的操作记录显示（如有）

**涉及模块**：

- [x] API 接口（操作记录查询接口）
- [x] Service 层（操作记录处理逻辑）
- [x] 数据模型（操作记录数据结构）

**不涉及**：

- ❌ 数据库结构（已存储授权员工信息）
- ❌ 前端 UI 组件（显示逻辑不变）
- ❌ 操作记录创建逻辑（已记录授权员工信息）

**前置依赖**：

- ✅ 敏感操作权限验证功能（`story-pos-sensitive-operation-settings`）
- ✅ 操作记录中已存储授权员工信息

---

## 📊 初步评估

### 技术复杂度

- [x] **低**：主要是数据读取和判断逻辑，无复杂业务逻辑
- [ ] **中**：需要前后端联调，基础业务逻辑
- [ ] **高**：涉及架构调整、第三方集成、复杂算法

### 工作量预估

[粗略估算，具体 SP 在技术方案评审时确定]

- **预计天数**: 1-2 天
- **预估 SP**: 2 SP（待技术评审确认）

**分解**：

- 修改操作记录查询 Service 逻辑：0.5 天
- 修改操作记录响应 DTO：0.5 天
- 测试和验证：0.5-1 天

### 风险识别

**潜在风险**：

1. **历史数据兼容性**：v2.10.0 之前的历史操作记录没有授权员工信息，需要正确处理
2. **数据一致性**：确保授权员工信息在操作记录中正确存储和读取
3. **显示逻辑**：需要明确区分哪些操作需要显示授权人信息

**缓解措施**：

1. **历史数据兼容**：
   - 检查 `authorized_staff` 字段是否存在
   - 如果不存在，使用原有逻辑显示操作人信息
   - 如果存在，显示授权人信息

2. **数据一致性**：
   - 依赖现有的操作记录创建逻辑（已在 `story-pos-sensitive-operation-settings` 中实现）
   - 确保授权员工信息正确存储到 `data` JSON 字段中

3. **显示逻辑**：
   - 仅对敏感操作（折扣/退款/免单）应用新逻辑
   - 其他操作保持原有显示逻辑

---

## 🔗 相关资源

### 参考需求

- 关联 Spec: [story-pos-sensitive-operation-settings](../../../shared/specs/archived/v2.12/story-pos-sensitive-operation-settings/)
- 前置功能: 敏感操作权限验证功能（`story-pos-sensitive-operation-settings`）

### 相关文档

- 需求文档: `docs/shared/specs/archived/v2.10.0/story-pos-sensitive-operation-settings/requirements.md`
- 设计文档: `docs/shared/specs/archived/v2.10.0/story-pos-sensitive-operation-settings/design.md`
- 操作记录 Service: `main/app/service/order_operation_log.go`
- 操作记录查询: `main/app/service/order_manage.go` - `GetOrderOperationLog` 方法

---

## 🤝 需求评审

### 评审参与人

| 角色         | 姓名   | 签名/日期 |
| ------------ | ------ | --------- |
| 产品经理     | {姓名} |           |
| 技术负责人   | {姓名} |           |
| 开发代表     | {姓名} |           |
| 测试代表     | {姓名} |           |
| UI/UX 设计师 | {姓名} |           |

### 评审结论

- [ ] ✅ **批准**：进入技术方案设计阶段
- [ ] 🔄 **修改后批准**：需补充以下内容
- [ ] ❌ **拒绝**：不符合产品规划或优先级

**评审意见**：

```
[记录评审会议的关键讨论和决策]
```

**下一步行动**：

- [x] ✅ **功能已实现**：代码已实现并合并
- [ ] 创建 Spec：`story-pos-sensitive-operation-log-display`（可选）
- [ ] 分配负责人：{姓名}
- [ ] 目标 Sprint：Sprint {N}

---

## 📝 附录

### User Story（初稿）

**作为** 门店经理/店长  
**我想** 在订单详情的操作记录中看到授权人的信息（而不是操作人信息）  
**以便于** 准确追溯敏感操作的授权情况，进行财务审计

**作为** 财务人员  
**我想** 在审计时看到每次敏感操作的授权人信息  
**以便于** 明确责任归属，满足审计要求

### AC 验收标准（初稿）

1. **WHEN** 查看订单详情的操作记录，操作类型为折扣/退款/免单 **AND** 该操作使用了授权验证 **THEN** 系统 **SHALL** 显示授权人的姓名和邮箱（格式：`授权人姓名 (授权人邮箱)`）

2. **WHEN** 查看订单详情的操作记录，操作类型为折扣/退款/免单 **AND** 该操作未使用授权验证（操作人本身在授权名单中） **THEN** 系统 **SHALL** 显示操作人的姓名和邮箱（格式：`操作人姓名 (操作人邮箱)`）

3. **WHEN** 查看订单详情的操作记录，操作类型为其他操作（非敏感操作） **THEN** 系统 **SHALL** 显示操作人的姓名和邮箱（保持原有逻辑）

4. **WHEN** 查看 v2.10.0 之前的历史订单操作记录 **THEN** 系统 **SHALL** 显示操作人的姓名和邮箱（历史数据兼容）

5. **WHEN** 操作记录中存在 `authorized_staff` 信息 **THEN** 系统 **SHALL** 优先使用授权人信息，忽略操作人信息

6. **WHEN** 操作记录中不存在 `authorized_staff` 信息 **THEN** 系统 **SHALL** 使用操作人信息（原有逻辑）

**特殊说明**：

- 仅对敏感操作（折扣/退款/免单）应用新逻辑
- 其他操作保持原有显示逻辑
- 历史数据兼容：v2.10.0 之前的数据没有授权员工信息，继续显示操作人信息

### 技术实现要点

#### 1. 操作记录数据结构

操作记录的 `data` JSON 字段中已包含授权员工信息：

```json
{
  "discount": 10.0,
  "discount_type": 0,
  "authorized_staff": {
    "uuid": 123,
    "name": "张三",
    "email": "zhangsan@example.com"
  }
}
```

#### 2. 授权员工信息结构体定义

在 `main/pkg/eventbus/event/order_return_event.go` 中定义了授权员工信息结构体：

```go
// AuthorizedStaffInfo 授权员工信息
type AuthorizedStaffInfo struct {
	Uuid  uint64 `json:"uuid"`  // 授权员工UUID
	Name  string `json:"name"`  // 授权员工姓名
	Email string `json:"email"` // 授权员工邮箱
}
```

#### 3. Service 层实现

在 `main/app/service/order_manage.go` 的 `GetOrderOperationLog` 方法中实现了授权人信息显示逻辑：

```go:753:765:main/app/service/order_manage.go
realName := record.Operator.RealName
email := record.Operator.Username

// 获取授权人信息
if strings.Contains(record.Data, "authorized_staff") { // 授权操作时，记录了授权人信息
	var authorizedStaffInfo event.AuthorizedStaffInfo
	if err := utils.ExtractNestedFieldToStruct(record.Data, "authorized_staff", &authorizedStaffInfo); err == nil {
		realName = authorizedStaffInfo.Name
		email = authorizedStaffInfo.Email
	} else {
		ctx.Log().Info("解析订单操作记录时，获取授权人信息失败", zap.Any("companyUuid", ctx.GetCompanyUuid()), zap.Any("record", record), zap.Error(err))
	}
}
```

**实现说明**：
- 首先设置默认值为操作人的信息（`record.Operator.RealName` 和 `record.Operator.Username`）
- 检查操作记录的 `data` 字段中是否包含 `authorized_staff` 字符串
- 如果包含，使用 `utils.ExtractNestedFieldToStruct` 方法从 JSON 中提取 `authorized_staff` 字段并解析到 `AuthorizedStaffInfo` 结构体
- 如果解析成功，使用授权人的姓名和邮箱替换操作人信息
- 如果解析失败，记录日志但继续使用操作人信息（容错处理）

#### 4. JSON 嵌套字段提取工具方法

在 `main/pkg/utils/json.go` 中实现了 `ExtractNestedFieldToStruct` 方法：

```go:115:142:main/pkg/utils/json.go
// ExtractNestedFieldToStruct 从 JSON 字符串中提取指定的嵌套字段并解析到目标结构体
// jsonStr: JSON 字符串
// fieldName: 要提取的字段名
// target: 目标结构体指针
// 返回错误信息
func ExtractNestedFieldToStruct(jsonStr string, fieldName string, target any) error {
	var dataMap map[string]any
	if err := json.Unmarshal([]byte(jsonStr), &dataMap); err != nil {
		return fmt.Errorf("解析 JSON 失败: %w", err)
	}

	fieldValue, ok := dataMap[fieldName]
	if !ok {
		return fmt.Errorf("字段 %s 不存在", fieldName)
	}

	// 将字段值序列化为 JSON，再解析到目标结构体
	fieldBytes, err := json.Marshal(fieldValue)
	if err != nil {
		return fmt.Errorf("序列化字段 %s 失败: %w", fieldName, err)
	}

	if err := json.Unmarshal(fieldBytes, target); err != nil {
		return fmt.Errorf("解析字段 %s 到结构体失败: %w", fieldName, err)
	}

	return nil
}
```

**方法说明**：
- 该方法用于从 JSON 字符串中提取指定的嵌套字段
- 先将整个 JSON 解析为 `map[string]any`
- 提取指定字段的值
- 将字段值重新序列化为 JSON，再解析到目标结构体
- 支持嵌套字段的提取和类型转换

#### 5. 实现特点

1. **兼容性处理**：
   - 使用 `strings.Contains` 快速检查是否存在授权员工信息，避免不必要的 JSON 解析
   - 如果解析失败，记录日志但不影响主流程，继续使用操作人信息

2. **适用范围**：
   - 对所有操作记录都进行检查，如果存在 `authorized_staff` 字段就使用授权人信息
   - 不限制操作类型，因为授权验证可能应用于多种操作

3. **数据来源**：
   - 授权员工信息已经在创建操作记录时存储到 `data` JSON 字段中（由 `story-pos-sensitive-operation-settings` 实现）
   - 本功能仅负责读取和显示

#### 6. 响应数据结构

操作记录响应结构（`resp.OrderOperationLog`）中的 `RealName` 和 `Email` 字段会根据上述逻辑填充：
- 如果存在授权员工信息：`RealName` = 授权人姓名，`Email` = 授权人邮箱
- 如果不存在授权员工信息：`RealName` = 操作人姓名，`Email` = 操作人邮箱（用户名）

### 线框图/原型（可选）

参考现有的订单详情页操作记录显示，显示格式保持不变，仅改变显示的内容（授权人信息 vs 操作人信息）。

---

## 📄 模板使用说明

### 何时使用此模板

- ✅ 产品经理提出新功能想法
- ✅ 用户反馈需求建议
- ✅ 技术团队提出改进方案
- ✅ 需要团队讨论和评审的需求

### 与 Spec 的区别

| 阶段         | 文档类型     | 详细程度 | 用途                      |
| ------------ | ------------ | -------- | ------------------------- |
| **需求发起** | Proposal     | 粗略     | 团队评审、决策是否做      |
| **需求确认** | Requirements | 详细     | User Story + AC，开发依据 |
| **技术设计** | Design       | 详细     | 技术方案，实现指导        |
| **任务分解** | Tasks        | 详细     | 开发执行，进度追踪        |

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

---

## ✅ 实现记录

### 实现状态

- ✅ **代码已实现**：2025-01-16
- ✅ **实现位置**：`main/app/service/order_manage.go` - `GetOrderOperationLog` 方法
- ✅ **实现方式**：在操作记录查询时检查是否存在授权员工信息，如果存在则使用授权人信息替换操作人信息

### 实现文件

- **Service 层**：`main/app/service/order_manage.go` (行 753-765)
- **数据结构**：`main/pkg/eventbus/event/order_return_event.go` - `AuthorizedStaffInfo`
- **工具方法**：`main/pkg/utils/json.go` - `ExtractNestedFieldToStruct`

### 实现要点

1. **检查逻辑**：使用 `strings.Contains(record.Data, "authorized_staff")` 快速检查是否存在授权员工信息
2. **数据提取**：使用 `utils.ExtractNestedFieldToStruct` 方法从 JSON 中提取嵌套字段
3. **容错处理**：如果解析失败，记录日志但继续使用操作人信息，不影响主流程
4. **适用范围**：对所有操作记录都进行检查，不限制操作类型

---

**版本**: v1.1.0  
**创建日期**: 2025-01-16  
**最后更新**: 2025-01-16  
**维护者**: 产品组 + Scrum Master  
**相关规范**: `.cursor/rules/scrum_story_point.mdc`, `.cursor/rules/specs.mdc`

