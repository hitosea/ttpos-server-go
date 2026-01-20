# ERPNEXT 采购订单审批工作流配置指南（按金额等级审核）

> 📖 **用途**: 在 ERPNEXT 采购订单中配置按金额等级的多级审批流程

---

## 一、方案概述

ERPNEXT 支持通过**工作流（Workflow）**功能为采购订单添加多级审批流程。通过配置条件表达式，可以实现**按金额等级自动分配不同的审批人员**。

### ⚠️ 重要说明

**ERPNEXT 工作流的核心机制**：
- **不是**通过 `workflow_state` 来控制是否可以创建后续单据
- **而是**通过 `docstatus`（文档状态）来控制
- 只有 `docstatus = 1`（已提交）的采购订单才能创建收货单、发票等后续单据
- 工作流的作用是：**控制何时将 `docstatus` 从 0 更新为 1**

**按金额等级审核的实现方式**：
1. 采购订单提交后，`docstatus` 保持为 `0`（草稿），进入"待审批"状态
2. 根据订单金额（`grand_total`），自动进入不同的审批级别
3. 每个审批级别对应不同的审批人员角色
4. 所有级别审批通过后，工作流动作将 `docstatus` 更新为 `1`（已提交）
5. 只有 `docstatus = 1` 时，才能创建后续单据

### 金额等级示例

假设设置以下金额等级：

| 金额范围 | 审批级别 | 审批角色 | 说明 |
|---------|---------|---------|------|
| ≤ 10,000 | 一级审批 | 采购经理 | 小额采购，只需一级审批 |
| 10,001 - 50,000 | 二级审批 | 采购经理 → 财务经理 | 中额采购，需要两级审批 |
| 50,001 - 100,000 | 三级审批 | 采购经理 → 财务经理 → 总经理 | 大额采购，需要三级审批 |
| > 100,000 | 四级审批 | 采购经理 → 财务经理 → 总经理 → 董事长 | 超大额采购，需要四级审批 |

---

## 二、ERPNEXT 工作流配置步骤

### 2.1 创建工作流

#### 步骤 1：进入工作流管理页面

1. **登录 ERPNEXT 系统**
2. **导航路径**：
   ```
   主页 → 设置（Settings） → 工作流（Workflow） → 新建（New）
   ```
   或者直接访问：`/app/workflow/new`

#### 步骤 2：填写基本信息

在"工作流"表单中填写以下信息：

| 字段 | 值 | 说明 |
|------|-----|------|
| **工作流名称** | `Purchase Order Approval Workflow` | 工作流的显示名称 |
| **文档类型** | `Purchase Order` | 选择"采购订单" |
| **工作流状态字段** | `workflow_state` ⭐ **关键** | 存储工作流状态的字段名 |
| **工作流状态** | `Active` | 必须选择"启用"才能生效 |
| **是否系统工作流** | `否` | 用户自定义工作流 |
| **发送电子邮件提醒** | `是`（推荐） | 审批时发送邮件通知 |

**⚠️ 关键配置：工作流状态字段**

- **字段名**：`workflow_state`（默认值）
- **说明**：这是存储工作流状态的字段名，必须与采购订单中的字段名一致
- **检查方法**：
  1. 打开一个采购订单
  2. 查看是否有 `workflow_state` 字段
  3. 如果没有，需要在采购订单的 Customize Form 中添加此字段

#### 步骤 3：保存工作流

1. 点击右上角"保存"按钮
2. 系统会提示"工作流已保存"
3. **重要**：此时工作流还未生效，需要完成后续配置

### 2.2 定义工作流状态（Workflow States）

#### 步骤 1：进入状态配置页面

在工作流编辑页面，点击"工作流状态"（Workflow States）标签页。

#### 步骤 2：添加工作流状态

点击"添加行"（Add Row）按钮，依次添加以下状态：

#### ⚠️ States 界面配置说明

在 States 界面中，点击"添加行"（Add Row）按钮，然后填写以下字段：

**字段说明**：
- **State***（状态名称，必填）：输入状态的标识符，例如：`draft`、`Pending Level 1 Approval`、`Approved` 等
- **Doc Status**（文档状态，必填）：选择 `0`（Saved）、`1`（Submitted）或 `2`（Cancelled）
  - ⚠️ **关键**：只有 `Approved` 状态设置为 `1`，其他审批状态都设置为 `0`
- **Update Field**（更新字段，可选）：通常留空，由系统自动处理
- **Update Value**（更新值，可选）：通常留空，由系统自动处理
- **Only Allow Edit For***（仅允许编辑的角色，必填）：输入角色名称，例如：`Purchase Manager`、`All` 等
  - ⚠️ **注意**：此字段为必填项，不能留空

**界面配置表格**（直接在 States 界面中填写）：

| No. | State* | Doc Status | Update Field | Update Value | Only Allow Edit For* |
|-----|--------|------------|--------------|--------------|---------------------|
| 1 | `draft` | `0` | （留空） | （留空） | `All` |
| 2 | `Submitted` | `0` | （留空） | （留空） | `Purchase User` |
| 3 | `Pending Level 1 Approval` | `0` | （留空） | （留空） | `Purchase Manager` |
| 4 | `Pending Level 2 Approval` | `0` | （留空） | （留空） | `Accounts Manager` |
| 5 | `Pending Level 3 Approval` | `0` | （留空） | （留空） | `General Manager` |
| 6 | `Pending Level 4 Approval` | `0` | （留空） | （留空） | `Chairman` |
| 7 | `Approved` | `1` ⭐ | （留空） | （留空） | `System Manager` ⚠️ |
| 8 | `Rejected` | `0` | （留空） | （留空） | `Purchase User` |

**详细说明表格**：

| 序号 | 状态名称 | 状态标识 | 颜色 | 说明 | 是否允许编辑 | Docstatus |
|------|---------|---------|------|------|------------|-----------|
| 1 | Draft | `Draft` | 灰色（Gray） | 草稿状态，订单可编辑 | ✅ 是 | 0 |
| 2 | Submitted | `Submitted` | 蓝色（Blue） | 已提交，等待审批 | ❌ 否 | 0 |
| 3 | Pending Level 1 Approval | `Pending Level 1 Approval` | 橙色（Orange） | 待一级审批（采购经理） | ❌ 否 | 0 |
| 4 | Pending Level 2 Approval | `Pending Level 2 Approval` | 橙色（Orange） | 待二级审批（财务经理） | ❌ 否 | 0 |
| 5 | Pending Level 3 Approval | `Pending Level 3 Approval` | 橙色（Orange） | 待三级审批（总经理） | ❌ 否 | 0 |
| 6 | Pending Level 4 Approval | `Pending Level 4 Approval` | 橙色（Orange） | 待四级审批（董事长） | ❌ 否 | 0 |
| 7 | Approved | `Approved` | 绿色（Green） | **已批准，可以继续后续流程** | ❌ 否 | 1 ⭐ |
| 8 | Rejected | `Rejected` | 红色（Red） | 已拒绝，需要修改后重新提交 | ✅ 是 | 0 |

**⚠️ 关键配置点**：

1. **Doc Status 设置**：
   - 所有审批状态（Pending Level X Approval）都设置为 `0`
   - 只有 `Approved` 状态设置为 `1` ⭐ **最关键**
   - `Rejected` 状态设置为 `0`

2. **Only Allow Edit For 设置**：
   - `draft`：设置为 `All`（所有人都可以编辑）
   - `Submitted`：设置为 `Purchase User`（采购用户可以编辑）
   - `Pending Level 1 Approval`：设置为 `Purchase Manager`（只有采购经理可以编辑）
   - `Pending Level 2 Approval`：设置为 `Accounts Manager`（只有财务经理可以编辑）
   - `Pending Level 3 Approval`：设置为 `General Manager`（只有总经理可以编辑）
   - `Pending Level 4 Approval`：设置为 `Chairman`（只有董事长可以编辑）
   - `Approved`：设置为 `System Manager` ⚠️ **必填项处理**
     - 由于此字段为必填项，不能留空
     - 设置为 `System Manager`（系统管理员角色），这样普通用户无法编辑
     - 如果系统中没有 `System Manager` 角色，可以使用 `Administrator` 或其他只有管理员才有的角色
   - `Rejected`：设置为 `Purchase User`（采购用户可以修改后重新提交）

3. **Update Field 和 Update Value**：
   - 通常留空，由工作流转换（Transitions）自动处理
   - 如果需要特殊处理，可以在这里设置

**详细配置说明**：

**状态 1：Draft（草稿）**
- **状态标识**：`Draft`（必须与 ERPNEXT 默认状态一致）
- **颜色**：灰色
- **允许编辑**：✅ 是（用户可以修改订单）
- **Docstatus**：`0`
- **说明**：订单创建后的初始状态

**状态 2：Submitted（已提交）**
- **状态标识**：`Submitted`
- **颜色**：蓝色
- **允许编辑**：❌ 否（提交后不可编辑）
- **Docstatus**：`0`
- **说明**：订单已提交，但未触发审批流程

**状态 3-6：多级审批状态**
- **状态标识**：`Pending Level 1/2/3/4 Approval`
- **颜色**：橙色（用于提醒）
- **允许编辑**：❌ 否（审批中不可编辑）
- **Docstatus**：`0` ⚠️ **关键：保持为 0，不能创建后续单据**
- **说明**：订单已进入对应级别的审批流程，等待审批人员处理

**状态 7：Approved（已批准）⭐ 关键状态**
- **状态标识**：`Approved`（自定义状态）
- **颜色**：绿色（表示通过）
- **允许编辑**：❌ 否（已批准后不可编辑）
- **Docstatus**：`1` ⭐ **关键配置**
- **说明**：**订单已批准，可以继续后续流程**（创建收货单、发票等）
- **重要**：此状态下 `docstatus = 1`，**可以**创建后续单据

**状态 8：Rejected（已拒绝）**
- **状态标识**：`Rejected`（自定义状态）
- **颜色**：红色（表示拒绝）
- **允许编辑**：✅ 是（可以修改后重新提交）
- **Docstatus**：`0`
- **说明**：订单被拒绝，需要修改后重新提交审批

#### 步骤 3：保存状态配置

1. 添加完所有状态后，点击"保存"按钮
2. 系统会验证状态配置是否正确
3. 如有错误，会提示修改

### 2.3 配置工作流转换（Workflow Transitions）

#### 步骤 1：进入转换配置页面

在工作流编辑页面，点击"工作流转换"（Workflow Transitions）或"Transition Rules"标签页。

#### ⚠️ Transition Rules 界面配置说明

在 Transition Rules 界面中，点击"添加行"（Add Row）按钮，然后填写以下字段：

**字段说明**：
- **State***：从下拉列表选择当前状态（如：`draft`、`Submitted`、`Pending Level 1 Approval`）
- **Action***：输入动作名称（如：`Submit`、`Approve Level 1`、`Reject`）
- **Next State***：从下拉列表选择下一状态（如：`Pending Level 1 Approval`、`Approved`）
- **Allowed***：输入允许执行此动作的角色（如：`Purchase Manager`，多个角色用逗号分隔）

**注意**：如果界面有 Condition（条件）字段，可以输入 Python 表达式，如：`doc.grand_total <= 10000`

#### 步骤 2：添加状态转换规则

点击"添加行"（Add Row）按钮，依次添加以下转换（在界面表格中填写）：

**界面配置表格**（直接在 Transition Rules 界面中填写）：

| No. | State* | Action* | Next State* | Allowed* | Condition（如果有此字段） |
|-----|--------|---------|-------------|----------|------------------------|
| 1 | `draft` | `Submit` | `Submitted` | `Purchase User, Purchase Manager` | （留空） |
| 2 | `Submitted` | `Submit for Level 1 Approval` | `Pending Level 1 Approval` | `Purchase User` | `doc.grand_total <= 10000` |
| 3 | `Submitted` | `Submit for Level 2 Approval` | `Pending Level 2 Approval` | `Purchase User` | `doc.grand_total > 10000 and doc.grand_total <= 50000` |
| 4 | `Submitted` | `Submit for Level 3 Approval` | `Pending Level 3 Approval` | `Purchase User` | `doc.grand_total > 50000 and doc.grand_total <= 100000` |
| 5 | `Submitted` | `Submit for Level 4 Approval` | `Pending Level 4 Approval` | `Purchase User` | `doc.grand_total > 100000` |
| 6 | `Pending Level 1 Approval` | `Approve Level 1` | `Approved` | `Purchase Manager` | `doc.grand_total <= 10000` |
| 7 | `Pending Level 1 Approval` | `Approve Level 1 → Level 2` | `Pending Level 2 Approval` | `Purchase Manager` | `doc.grand_total > 10000` |
| 8 | `Pending Level 2 Approval` | `Approve Level 2 → Level 3` | `Pending Level 3 Approval` | `Accounts Manager` | `doc.grand_total > 50000` |
| 9 | `Pending Level 2 Approval` | `Approve Level 2` | `Approved` | `Accounts Manager` | `doc.grand_total <= 50000` |
| 10 | `Pending Level 3 Approval` | `Approve Level 3 → Level 4` | `Pending Level 4 Approval` | `General Manager` | `doc.grand_total > 100000` |
| 11 | `Pending Level 3 Approval` | `Approve Level 3` | `Approved` | `General Manager` | `doc.grand_total <= 100000` |
| 12 | `Pending Level 4 Approval` | `Approve Level 4` | `Approved` | `Chairman` | （留空） |
| 13 | `Pending Level 1 Approval` | `Reject` | `Rejected` | `Purchase Manager` | （留空） |
| 14 | `Pending Level 2 Approval` | `Reject` | `Rejected` | `Accounts Manager` | （留空） |
| 15 | `Pending Level 3 Approval` | `Reject` | `Rejected` | `General Manager` | （留空） |
| 16 | `Pending Level 4 Approval` | `Reject` | `Rejected` | `Chairman` | （留空） |
| 17 | `Rejected` | `Re-submit` | `Submitted` | `Purchase User` | （留空） |

**详细说明表格**：

| 序号 | 转换名称 | 当前状态 | 下一状态 | 条件（金额等级） | 允许的角色 | 说明 |
|------|---------|---------|---------|----------------|-----------|------|
| 1 | Submit | Draft | Submitted | - | Purchase User, Purchase Manager | 提交订单 |
| 2 | Submit for Level 1 Approval | Submitted | Pending Level 1 Approval | `doc.grand_total <= 10000` | Purchase User | 金额 ≤ 10,000，进入一级审批 |
| 3 | Submit for Level 2 Approval | Submitted | Pending Level 2 Approval | `doc.grand_total > 10000 and doc.grand_total <= 50000` | Purchase User | 金额 10,001-50,000，进入二级审批 |
| 4 | Submit for Level 3 Approval | Submitted | Pending Level 3 Approval | `doc.grand_total > 50000 and doc.grand_total <= 100000` | Purchase User | 金额 50,001-100,000，进入三级审批 |
| 5 | Submit for Level 4 Approval | Submitted | Pending Level 4 Approval | `doc.grand_total > 100000` | Purchase User | 金额 > 100,000，进入四级审批 |
| 6 | Approve Level 1 | Pending Level 1 Approval | Approved | `doc.grand_total <= 10000` | Purchase Manager | 一级审批通过（小额） |
| 7 | Approve Level 1 → Level 2 | Pending Level 1 Approval | Pending Level 2 Approval | `doc.grand_total > 10000` | Purchase Manager | 一级审批通过，进入二级审批 |
| 8 | Approve Level 2 → Level 3 | Pending Level 2 Approval | Pending Level 3 Approval | `doc.grand_total > 50000` | Accounts Manager | 二级审批通过，进入三级审批 |
| 9 | Approve Level 2 | Pending Level 2 Approval | Approved | `doc.grand_total <= 50000` | Accounts Manager | 二级审批通过（中额） |
| 10 | Approve Level 3 → Level 4 | Pending Level 3 Approval | Pending Level 4 Approval | `doc.grand_total > 100000` | General Manager | 三级审批通过，进入四级审批 |
| 11 | Approve Level 3 | Pending Level 3 Approval | Approved | `doc.grand_total <= 100000` | General Manager | 三级审批通过（大额） |
| 12 | Approve Level 4 ⭐ | Pending Level 4 Approval | Approved | - | Chairman | 四级审批通过（超大额） |
| 13-16 | Reject | Pending Level X Approval | Rejected | - | 对应审批角色 | 审批拒绝（每个级别都需要配置） |
| 17 | Re-submit | Rejected | Submitted | - | Purchase User | 重新提交审批 |

#### 步骤 3：详细配置每个转换

**转换 1：Submit（提交订单）**

```
动作（Action）：Submit
当前状态（Current State）：Draft
下一状态（Next State）：Submitted
条件（Condition）：留空（无条件）
允许的角色（Allowed Roles）：
  - Purchase User
  - Purchase Manager
动作（Action）：留空（使用默认动作）
```

**转换 2-5：按金额等级提交审批 ⭐ 关键转换**

这些转换根据订单金额自动选择审批级别：

**转换 2：金额 ≤ 10,000（一级审批）**

```
动作（Action）：Submit for Level 1 Approval
当前状态（Current State）：Submitted
下一状态（Next State）：Pending Level 1 Approval
条件（Condition）：doc.grand_total <= 10000
允许的角色（Allowed Roles）：
  - Purchase User
动作（Action）：留空（使用默认动作）
```

**转换 3：金额 10,001 - 50,000（二级审批）**

```
动作（Action）：Submit for Level 2 Approval
当前状态（Current State）：Submitted
下一状态（Next State）：Pending Level 2 Approval
条件（Condition）：doc.grand_total > 10000 and doc.grand_total <= 50000
允许的角色（Allowed Roles）：
  - Purchase User
动作（Action）：留空（使用默认动作）
```

**转换 4：金额 50,001 - 100,000（三级审批）**

```
动作（Action）：Submit for Level 3 Approval
当前状态（Current State）：Submitted
下一状态（Next State）：Pending Level 3 Approval
条件（Condition）：doc.grand_total > 50000 and doc.grand_total <= 100000
允许的角色（Allowed Roles）：
  - Purchase User
动作（Action）：留空（使用默认动作）
```

**转换 5：金额 > 100,000（四级审批）**

```
动作（Action）：Submit for Level 4 Approval
当前状态（Current State）：Submitted
下一状态（Next State）：Pending Level 4 Approval
条件（Condition）：doc.grand_total > 100000
允许的角色（Allowed Roles）：
  - Purchase User
动作（Action）：留空（使用默认动作）
```

**转换 6-12：多级审批流转 ⭐ 关键转换**

**转换 6：一级审批通过（小额，直接完成）**

```
动作（Action）：Approve Level 1
当前状态（Current State）：Pending Level 1 Approval
下一状态（Next State）：Approved
条件（Condition）：doc.grand_total <= 10000
允许的角色（Allowed Roles）：
  - Purchase Manager
动作（Action）：留空（使用默认动作）

说明：金额 ≤ 10,000 的订单，一级审批通过后直接完成
```

**转换 7：一级审批通过，进入二级审批**

```
动作（Action）：Approve Level 1 → Level 2
当前状态（Current State）：Pending Level 1 Approval
下一状态（Next State）：Pending Level 2 Approval
条件（Condition）：doc.grand_total > 10000
允许的角色（Allowed Roles）：
  - Purchase Manager
动作（Action）：留空（使用默认动作）

说明：金额 > 10,000 的订单，一级审批通过后进入二级审批
```

**转换 8：二级审批通过，进入三级审批**

```
动作（Action）：Approve Level 2 → Level 3
当前状态（Current State）：Pending Level 2 Approval
下一状态（Next State）：Pending Level 3 Approval
条件（Condition）：doc.grand_total > 50000
允许的角色（Allowed Roles）：
  - Accounts Manager
动作（Action）：留空（使用默认动作）

说明：金额 > 50,000 的订单，二级审批通过后进入三级审批
```

**转换 9：二级审批通过（中额，直接完成）**

```
动作（Action）：Approve Level 2
当前状态（Current State）：Pending Level 2 Approval
下一状态（Next State）：Approved
条件（Condition）：doc.grand_total <= 50000
允许的角色（Allowed Roles）：
  - Accounts Manager
动作（Action）：留空（使用默认动作）

说明：金额 ≤ 50,000 的订单，二级审批通过后直接完成
```

**转换 10：三级审批通过，进入四级审批**

```
动作（Action）：Approve Level 3 → Level 4
当前状态（Current State）：Pending Level 3 Approval
下一状态（Next State）：Pending Level 4 Approval
条件（Condition）：doc.grand_total > 100000
允许的角色（Allowed Roles）：
  - General Manager
动作（Action）：留空（使用默认动作）

说明：金额 > 100,000 的订单，三级审批通过后进入四级审批
```

**转换 11：三级审批通过（大额，直接完成）**

```
动作（Action）：Approve Level 3
当前状态（Current State）：Pending Level 3 Approval
下一状态（Next State）：Approved
条件（Condition）：doc.grand_total <= 100000
允许的角色（Allowed Roles）：
  - General Manager
动作（Action）：留空（使用默认动作）

说明：金额 ≤ 100,000 的订单，三级审批通过后直接完成
```

**转换 12：四级审批通过 ⭐ 最关键转换**

```
动作（Action）：Approve Level 4
当前状态（Current State）：Pending Level 4 Approval
下一状态（Next State）：Approved
条件（Condition）：留空（无条件）
允许的角色（Allowed Roles）：
  - Chairman
动作（Action）：留空（使用默认动作）

说明：超大额订单，四级审批通过后完成
```

**转换 13：Reject（审批拒绝）**

```
动作（Action）：Reject
当前状态（Current State）：Pending Level 1/2/3/4 Approval
下一状态（Next State）：Rejected
条件（Condition）：留空（无条件）
允许的角色（Allowed Roles）：
  - Purchase Manager（一级审批）
  - Accounts Manager（二级审批）
  - General Manager（三级审批）
  - Chairman（四级审批）
动作（Action）：留空（使用默认动作）

说明：任何级别的审批人员都可以拒绝订单
```

**转换 14：Re-submit（重新提交审批）**

```
动作（Action）：Re-submit
当前状态（Current State）：Rejected
下一状态（Next State）：Submitted
条件（Condition）：留空（无条件）
允许的角色（Allowed Roles）：
  - Purchase User
动作（Action）：留空（使用默认动作）

说明：被拒绝的订单修改后重新提交，会根据金额重新进入对应审批级别
```

#### 步骤 4：保存转换配置

1. 添加完所有转换后，点击"保存"按钮
2. 系统会验证转换配置是否正确
3. 确保每个转换的"当前状态"和"下一状态"都在工作流状态中已定义

### 2.4 配置工作流动作（Workflow Actions）

工作流动作定义了状态转换时执行的具体操作。ERPNEXT 会自动处理大部分动作，但我们需要配置关键的 `docstatus` 更新。

#### 步骤 1：进入动作配置页面

在工作流编辑页面，点击"工作流动作"（Workflow Actions）标签页。

#### 步骤 2：配置关键动作

**动作 1：所有级别的审批通过动作 ⭐ 最关键动作**

需要为每个"Approved"转换配置更新 `docstatus`：

**配置示例：Approve Level 1（一级审批通过）**

```
动作名称：Approve Level 1
触发时机：Pending Level 1 Approval → Approved

动作类型 1：Update Field（更新字段）- workflow_state
字段：workflow_state
值：Approved
说明：更新工作流状态为"已批准"

动作类型 2：Update Field（更新字段）- docstatus ⭐ 关键动作
字段：docstatus
值：1
说明：将文档状态更新为"已提交"，这是允许创建后续单据的关键

动作类型 3：Email Notification（邮件通知）
收件人：订单创建人（doc.owner）
主题：采购订单已批准：{{ doc.name }}
内容：采购订单 {{ doc.name }} 已通过审批，可以继续后续流程。
```

**同样需要为以下转换配置相同的动作**：
- Approve Level 2（二级审批通过）
- Approve Level 3（三级审批通过）
- Approve Level 4（四级审批通过）

**⚠️ 关键说明**：
✅ 所有审批通过动作都必须同时更新两个字段：
   1. `workflow_state = "Approved"`（工作流状态）
   2. `docstatus = 1`（文档状态）⭐ 最关键
✅ 只有 `docstatus = 1` 时，才能创建后续单据：
   - 创建收货单（Purchase Receipt）
   - 创建发票（Purchase Invoice）
   - 创建其他相关单据

**动作 2：提交审批时的通知（可选）**

```
动作名称：Submit for Level X Approval
触发时机：Submitted → Pending Level X Approval
动作类型：Email Notification（邮件通知）
收件人：对应级别的审批人员（从角色中获取）
主题：采购订单待审批：{{ doc.name }}（金额：{{ doc.grand_total }}）
内容：采购订单 {{ doc.name }} 已提交审批，金额为 {{ doc.grand_total }}，请及时处理。
```

**动作 3：审批拒绝时的通知（可选）**

```
动作名称：Reject
触发时机：Pending Level X Approval → Rejected
动作类型：Email Notification（邮件通知）
收件人：订单创建人（doc.owner）
主题：采购订单已拒绝：{{ doc.name }}
内容：采购订单 {{ doc.name }} 已被拒绝，请修改后重新提交。
```

#### 步骤 3：保存动作配置

1. 配置完所有动作后，点击"保存"按钮
2. 系统会验证动作配置是否正确

### 2.5 配置权限 ⭐ 关键步骤

权限配置是确保审批流程正常工作的关键。需要确保相关角色具有执行工作流动作的权限。

#### 步骤 1：进入权限管理页面

导航路径：
```
主页 → 设置（Settings） → 用户和权限（Users and Permissions） → 角色权限管理器（Role Permissions Manager）
```
或者直接访问：`/app/role-permissions-manager`

#### 步骤 2：配置角色权限

**角色 1：Purchase User（采购用户）**

```
角色名称：Purchase User
文档类型：Purchase Order

权限配置：
✅ 读取（Read）：是
✅ 创建（Create）：是
✅ 写入（Write）：是（仅限 Draft 状态）
✅ 提交（Submit）：是
✅ 取消（Cancel）：是（仅限 Draft 状态）

工作流动作权限：
✅ Submit for Level 1/2/3/4 Approval：是（可以提交审批）
✅ Re-submit：是（可以重新提交被拒绝的订单）
❌ Approve：否（不能审批）
❌ Reject：否（不能拒绝）
```

**角色 2：Purchase Manager（采购经理）⭐ 一级审批角色**

```
角色名称：Purchase Manager
文档类型：Purchase Order

权限配置：
✅ 读取（Read）：是
✅ 创建（Create）：是
✅ 写入（Write）：是（仅限 Draft 和 Rejected 状态）
✅ 提交（Submit）：是
✅ 取消（Cancel）：是（仅限 Draft 状态）

工作流动作权限：
✅ Submit for Level 1/2/3/4 Approval：是（可以提交审批）
✅ Approve Level 1：是（可以一级审批通过）⭐ 关键权限
✅ Approve Level 1 → Level 2：是（可以一级审批通过并流转到二级）
✅ Reject：是（可以审批拒绝）⭐ 关键权限
✅ Re-submit：是（可以重新提交）

后续流程权限：
✅ 创建收货单（Purchase Receipt）：是（审批通过后）
✅ 创建发票（Purchase Invoice）：是（审批通过后）
✅ 查看订单状态：是
```

**角色 3：Accounts Manager（财务经理）⭐ 二级审批角色**

```
角色名称：Accounts Manager
文档类型：Purchase Order

权限配置：
✅ 读取（Read）：是
✅ 创建（Create）：否（通常不创建订单）
✅ 写入（Write）：否（不能修改订单）
✅ 提交（Submit）：否
✅ 取消（Cancel）：否

工作流动作权限：
❌ Submit for Approval：否（不能提交审批）
✅ Approve Level 2：是（可以二级审批通过）⭐ 关键权限
✅ Approve Level 2 → Level 3：是（可以二级审批通过并流转到三级）
✅ Reject：是（可以审批拒绝）⭐ 关键权限
❌ Re-submit：否（不能重新提交）

后续流程权限：
✅ 创建发票（Purchase Invoice）：是（审批通过后）
✅ 查看订单状态：是
```

**角色 4：General Manager（总经理）⭐ 三级审批角色**

```
角色名称：General Manager
文档类型：Purchase Order

权限配置：
✅ 读取（Read）：是
✅ 创建（Create）：否
✅ 写入（Write）：否
✅ 提交（Submit）：否
✅ 取消（Cancel）：否

工作流动作权限：
✅ Approve Level 3：是（可以三级审批通过）⭐ 关键权限
✅ Approve Level 3 → Level 4：是（可以三级审批通过并流转到四级）
✅ Reject：是（可以审批拒绝）⭐ 关键权限

后续流程权限：
✅ 查看订单状态：是
```

**角色 5：Chairman（董事长）⭐ 四级审批角色**

```
角色名称：Chairman
文档类型：Purchase Order

权限配置：
✅ 读取（Read）：是
✅ 创建（Create）：否
✅ 写入（Write）：否
✅ 提交（Submit）：否
✅ 取消（Cancel）：否

工作流动作权限：
✅ Approve Level 4：是（可以四级审批通过）⭐ 关键权限
✅ Reject：是（可以审批拒绝）⭐ 关键权限

后续流程权限：
✅ 查看订单状态：是
```

#### 步骤 3：配置工作流状态权限

**重要**：需要确保不同状态下的权限设置正确：

**状态：Pending Level 1 Approval（待一级审批）**
- Purchase User：❌ 不能编辑
- Purchase Manager：✅ 可以审批（Approve Level 1 / Reject）

**状态：Pending Level 2 Approval（待二级审批）**
- Purchase User：❌ 不能编辑
- Accounts Manager：✅ 可以审批（Approve Level 2 / Reject）

**状态：Pending Level 3 Approval（待三级审批）**
- Purchase User：❌ 不能编辑
- General Manager：✅ 可以审批（Approve Level 3 / Reject）

**状态：Pending Level 4 Approval（待四级审批）**
- Purchase User：❌ 不能编辑
- Chairman：✅ 可以审批（Approve Level 4 / Reject）

**状态：Approved（已批准）⭐ 关键状态**
- Purchase User：✅ 可以创建收货单、发票等后续单据
- Purchase Manager：✅ 可以创建收货单、发票等后续单据
- Accounts Manager：✅ 可以创建发票等后续单据
- **所有角色**：❌ 不能编辑订单本身（已批准后不可修改）

#### 步骤 4：验证权限配置

1. 使用不同角色的用户登录系统
2. 测试是否可以执行相应的工作流动作
3. 确保审批通过后可以继续后续流程

---

## 三、金额等级配置说明

### 3.1 金额等级设置

在配置工作流转换时，需要根据实际业务需求设置金额等级。以下是配置示例：

**示例配置 1：简单三级审批**

| 金额范围 | 审批级别 | 审批角色 |
|---------|---------|---------|
| ≤ 10,000 | 一级审批 | 采购经理 |
| 10,001 - 50,000 | 二级审批 | 采购经理 → 财务经理 |
| > 50,000 | 三级审批 | 采购经理 → 财务经理 → 总经理 |

**对应的条件表达式**：

```python
# 一级审批（金额 ≤ 10,000）
doc.grand_total <= 10000

# 二级审批（金额 10,001 - 50,000）
doc.grand_total > 10000 and doc.grand_total <= 50000

# 三级审批（金额 > 50,000）
doc.grand_total > 50000
```

**示例配置 2：四级审批（如本文档示例）**

| 金额范围 | 审批级别 | 审批角色 |
|---------|---------|---------|
| ≤ 10,000 | 一级审批 | 采购经理 |
| 10,001 - 50,000 | 二级审批 | 采购经理 → 财务经理 |
| 50,001 - 100,000 | 三级审批 | 采购经理 → 财务经理 → 总经理 |
| > 100,000 | 四级审批 | 采购经理 → 财务经理 → 总经理 → 董事长 |

**对应的条件表达式**：

```python
# 一级审批（金额 ≤ 10,000）
doc.grand_total <= 10000

# 二级审批（金额 10,001 - 50,000）
doc.grand_total > 10000 and doc.grand_total <= 50000

# 三级审批（金额 50,001 - 100,000）
doc.grand_total > 50000 and doc.grand_total <= 100000

# 四级审批（金额 > 100,000）
doc.grand_total > 100000
```

### 3.2 条件表达式语法

ERPNEXT 工作流支持 Python 表达式，常用语法：

| 条件类型 | Python 表达式 | 说明 |
|---------|--------------|------|
| 金额等于 | `doc.grand_total == 10000` | 金额等于 10,000 |
| 金额大于 | `doc.grand_total > 10000` | 金额大于 10,000 |
| 金额大于等于 | `doc.grand_total >= 10000` | 金额大于等于 10,000 |
| 金额小于 | `doc.grand_total < 10000` | 金额小于 10,000 |
| 金额小于等于 | `doc.grand_total <= 10000` | 金额小于等于 10,000 |
| 金额范围 | `doc.grand_total > 10000 and doc.grand_total <= 50000` | 金额在 10,001 - 50,000 之间 |
| 组合条件（AND） | `doc.grand_total > 10000 and doc.supplier == "Supplier A"` | 金额大于 10,000 且供应商为 Supplier A |
| 组合条件（OR） | `doc.grand_total > 10000 or doc.supplier == "VIP Supplier"` | 金额大于 10,000 或供应商为 VIP Supplier |

### 3.3 其他条件示例

除了金额，还可以根据其他条件设置审批流程：

**示例 1：特定供应商需要特殊审批**

```python
# 特定供应商的订单需要三级审批
doc.supplier == "VIP Supplier" and doc.grand_total > 50000
```

**示例 2：特定物品类别需要特殊审批**

```python
# 特定物品类别的订单需要二级审批
doc.items[0].item_group == "Capital Goods" and doc.grand_total > 10000
```

**示例 3：组合条件**

```python
# 金额大于 50,000 或特定供应商的订单需要三级审批
doc.grand_total > 50000 or doc.supplier == "VIP Supplier"
```

---

## 四、工作流状态流转图

### 4.1 完整状态流转图

```
工作流状态（workflow_state）流转：

┌─────────────────────────────────────────────────────────┐
│ Draft（草稿）                                             │
│   workflow_state: Draft                                   │
│   docstatus: 0                                            │
└───────────────┬───────────────────────────────────────────┘
                │ [提交]
                ▼
┌─────────────────────────────────────────────────────────┐
│ Submitted（已提交）                                       │
│   workflow_state: Submitted                               │
│   docstatus: 0  ⚠️ 关键：保持为 0，不提交                │
└───────────────┬───────────────────────────────────────────┘
                │ [根据金额自动选择审批级别]
                ├─→ [金额 ≤ 10,000]
                │         ▼
                │   ┌─────────────────────────────────────┐
                │   │ Pending Level 1 Approval            │
                │   │   workflow_state: Pending Level 1  │
                │   │   docstatus: 0                      │
                │   └───────────┬─────────────────────────┘
                │               │ [一级审批通过]
                │               ├─→ [金额 ≤ 10,000] → Approved
                │               └─→ [金额 > 10,000] → Pending Level 2
                │
                ├─→ [金额 10,001 - 50,000]
                │         ▼
                │   ┌─────────────────────────────────────┐
                │   │ Pending Level 1 Approval            │
                │   └───────────┬─────────────────────────┘
                │               │ [一级审批通过]
                │               ▼
                │   ┌─────────────────────────────────────┐
                │   │ Pending Level 2 Approval            │
                │   │   workflow_state: Pending Level 2  │
                │   │   docstatus: 0                      │
                │   └───────────┬─────────────────────────┘
                │               │ [二级审批通过]
                │               ├─→ [金额 ≤ 50,000] → Approved
                │               └─→ [金额 > 50,000] → Pending Level 3
                │
                ├─→ [金额 50,001 - 100,000]
                │         ▼
                │   ┌─────────────────────────────────────┐
                │   │ Pending Level 1 Approval            │
                │   └───────────┬─────────────────────────┘
                │               │ [一级审批通过]
                │               ▼
                │   ┌─────────────────────────────────────┐
                │   │ Pending Level 2 Approval            │
                │   └───────────┬─────────────────────────┘
                │               │ [二级审批通过]
                │               ▼
                │   ┌─────────────────────────────────────┐
                │   │ Pending Level 3 Approval            │
                │   │   workflow_state: Pending Level 3  │
                │   │   docstatus: 0                      │
                │   └───────────┬─────────────────────────┘
                │               │ [三级审批通过]
                │               ├─→ [金额 ≤ 100,000] → Approved
                │               └─→ [金额 > 100,000] → Pending Level 4
                │
                └─→ [金额 > 100,000]
                          ▼
                    ┌─────────────────────────────────────┐
                    │ Pending Level 1 Approval            │
                    └───────────┬─────────────────────────┘
                                │ [一级审批通过]
                                ▼
                    ┌─────────────────────────────────────┐
                    │ Pending Level 2 Approval            │
                    └───────────┬─────────────────────────┘
                                │ [二级审批通过]
                                ▼
                    ┌─────────────────────────────────────┐
                    │ Pending Level 3 Approval            │
                    └───────────┬─────────────────────────┘
                                │ [三级审批通过]
                                ▼
                    ┌─────────────────────────────────────┐
                    │ Pending Level 4 Approval            │
                    │   workflow_state: Pending Level 4  │
                    │   docstatus: 0                      │
                    └───────────┬─────────────────────────┘
                                │ [四级审批通过]
                                ▼
┌─────────────────────────────────────────────────────────┐
│ Approved（已批准）                                        │
│   workflow_state: Approved                                │
│   docstatus: 1  ⚠️ 关键：更新为 1，可以创建后续单据      │
│   ✅ 可以创建收货单                                       │
│   ✅ 可以创建发票                                         │
└─────────────────────────────────────────────────────────┘
```

### 4.2 审批拒绝流程

```
任何审批级别都可以拒绝：

Pending Level X Approval
        │
        │ [拒绝]
        ▼
┌─────────────────────────────────────────────────────────┐
│ Rejected（已拒绝）                                        │
│   workflow_state: Rejected                                │
│   docstatus: 0                                            │
└───────────────┬───────────────────────────────────────────┘
                │ [重新提交]
                ▼
┌─────────────────────────────────────────────────────────┐
│ Submitted（已提交）                                       │
│   重新根据金额进入对应审批级别                           │
└─────────────────────────────────────────────────────────┘
```

---

## 五、测试验证

### 5.1 测试场景

#### 场景 1：小额订单（金额 ≤ 10,000）

**测试步骤**：
1. 创建采购订单，金额设置为 8,000
2. 提交订单
3. **验证**：订单进入 `Pending Level 1 Approval` 状态
4. 使用采购经理账号登录
5. 审批通过
6. **验证**：订单直接进入 `Approved` 状态，`docstatus = 1`
7. **验证**：可以创建收货单和发票

#### 场景 2：中额订单（金额 10,001 - 50,000）

**测试步骤**：
1. 创建采购订单，金额设置为 30,000
2. 提交订单
3. **验证**：订单进入 `Pending Level 1 Approval` 状态
4. 使用采购经理账号登录，审批通过
5. **验证**：订单进入 `Pending Level 2 Approval` 状态
6. 使用财务经理账号登录，审批通过
7. **验证**：订单进入 `Approved` 状态，`docstatus = 1`
8. **验证**：可以创建收货单和发票

#### 场景 3：大额订单（金额 50,001 - 100,000）

**测试步骤**：
1. 创建采购订单，金额设置为 80,000
2. 提交订单
3. **验证**：订单依次经过 Level 1 → Level 2 → Level 3 审批
4. 每个级别审批通过后，进入下一级别
5. 三级审批通过后，订单进入 `Approved` 状态
6. **验证**：`docstatus = 1`，可以创建后续单据

#### 场景 4：超大额订单（金额 > 100,000）

**测试步骤**：
1. 创建采购订单，金额设置为 150,000
2. 提交订单
3. **验证**：订单依次经过 Level 1 → Level 2 → Level 3 → Level 4 审批
4. 四级审批通过后，订单进入 `Approved` 状态
5. **验证**：`docstatus = 1`，可以创建后续单据

#### 场景 5：审批拒绝

**测试步骤**：
1. 创建采购订单，金额设置为 30,000
2. 提交订单，进入 `Pending Level 1 Approval`
3. 采购经理拒绝订单
4. **验证**：订单进入 `Rejected` 状态
5. 采购用户修改订单后重新提交
6. **验证**：订单重新根据金额进入对应审批级别

### 5.2 验证检查清单

**工作流配置检查**：
- [ ] 工作流状态为 Active（启用）
- [ ] 已定义所有必需的状态（Draft、Submitted、Pending Level 1/2/3/4 Approval、Approved、Rejected）
- [ ] 已配置所有必需的转换（按金额等级提交、多级审批流转）
- [ ] 所有 Approve 转换的下一状态为 `Approved`（最终审批）或下一级别（中间审批）

**权限配置检查**：
- [ ] Purchase User 可以提交审批
- [ ] Purchase Manager 可以一级审批
- [ ] Accounts Manager 可以二级审批
- [ ] General Manager 可以三级审批
- [ ] Chairman 可以四级审批
- [ ] 审批通过后，用户可以创建收货单和发票

**功能验证检查**：
- [ ] 提交订单后根据金额自动进入对应审批级别
- [ ] 小额订单（≤ 10,000）只需一级审批
- [ ] 中额订单（10,001 - 50,000）需要两级审批
- [ ] 大额订单（50,001 - 100,000）需要三级审批
- [ ] 超大额订单（> 100,000）需要四级审批
- [ ] 审批通过后 `docstatus = 1`
- [ ] 审批通过后可以创建收货单和发票
- [ ] 审批拒绝后可以重新提交

---

## 六、常见问题排查

### 6.1 订单提交后没有进入审批状态

**症状**：
- 提交订单后，`workflow_state` 仍然是 `Submitted`，没有进入审批状态

**排查步骤**：
1. 检查工作流是否启用（状态是否为 Active）
2. 检查工作流转换条件是否正确
   - 检查金额条件表达式是否正确
   - 检查订单金额是否满足条件
3. 检查用户角色权限
   - 用户是否有 "Submit for Level X Approval" 权限

**解决方案**：
- 确保工作流状态为 Active
- 检查转换条件，确保金额条件表达式正确
- 确保用户角色有相应权限

### 6.2 审批通过后无法创建收货单/发票

**症状**：
- 审批通过后，点击"创建"按钮，无法创建收货单或发票

**排查步骤**：
1. **⭐ 最关键：检查 docstatus**
   - 打开订单详情页
   - 查看 `docstatus` 字段
   - ✅ 应该为 `1`（已提交）
   - ❌ 如果是 `0`（草稿），**这是问题所在**
   - **解决方案**：检查工作流动作是否配置了更新 `docstatus` 为 `1`

2. **检查工作流动作配置**
   - 打开工作流配置页面
   - 找到所有 "Approve Level X" 动作
   - 检查是否配置了更新 `docstatus` 字段
   - ✅ 应该有一个动作：`docstatus = 1`
   - ❌ 如果没有，需要添加此动作

3. **检查工作流状态**
   - 打开订单详情页
   - 查看 `workflow_state` 字段
   - ✅ 应该为 `Approved`
   - ❌ 如果不是 `Approved`，说明审批未成功

**解决方案**：
- ⭐ **最关键**：确保所有审批通过动作都更新了 `docstatus = 1`
- 确保 `workflow_state = Approved`
- 确保工作流动作配置了更新 `docstatus` 字段

### 6.3 金额等级判断不正确

**症状**：
- 订单金额为 15,000，但进入了错误的审批级别

**排查步骤**：
1. 检查工作流转换条件表达式
2. 检查金额范围设置是否正确
3. 检查条件表达式的逻辑（AND/OR）

**解决方案**：
- 确保条件表达式正确，例如：
  - `doc.grand_total <= 10000`（一级审批）
  - `doc.grand_total > 10000 and doc.grand_total <= 50000`（二级审批）
- 确保金额范围不重叠且覆盖所有情况

---

## 七、邮件提醒配置详解

### 7.1 前置条件：配置邮件服务器

在配置工作流邮件提醒之前，需要先配置 ERPNext 的邮件服务器。

#### 步骤 1：进入邮件设置页面

导航路径：
```
主页 → 设置（Settings） → 集成（Integrations） → 邮箱账户（Email Account） → 新建（New）
```
或者直接访问：`/app/email-account/new`

#### 步骤 2：配置邮件账户

**基本配置**：

| 字段 | 值 | 说明 |
|------|-----|------|
| **邮箱账户名称** | `Default Email Account` | 邮件账户的显示名称 |
| **邮箱ID** | `noreply@yourcompany.com` | 发件人邮箱地址 |
| **启用** | ✅ 是 | 必须启用才能发送邮件 |
| **使用TLS** | ✅ 是（推荐） | 使用加密连接 |
| **SMTP服务器** | `smtp.gmail.com`（示例） | 根据邮件服务商配置 |
| **SMTP端口** | `587`（TLS）或 `465`（SSL） | 根据邮件服务商配置 |
| **登录凭证** | 邮箱用户名和密码 | 或使用应用专用密码 |

**常见邮件服务商配置**：

**Gmail**：
```
SMTP服务器：smtp.gmail.com
SMTP端口：587（TLS）或 465（SSL）
需要启用"允许不够安全的应用访问"或使用应用专用密码
```

**Outlook/Office 365**：
```
SMTP服务器：smtp.office365.com
SMTP端口：587（TLS）
```

**企业邮箱（如腾讯企业邮箱）**：
```
SMTP服务器：smtp.exmail.qq.com
SMTP端口：465（SSL）或 587（TLS）
```

#### 步骤 3：测试邮件发送

1. 配置完成后，点击"测试连接"按钮
2. 系统会发送测试邮件到配置的邮箱地址
3. 确认收到测试邮件后，保存配置

### 7.2 在工作流中启用邮件提醒

#### 步骤 1：在工作流基本信息中启用

在工作流创建/编辑页面：

| 字段 | 值 | 说明 |
|------|-----|------|
| **发送电子邮件提醒** | ✅ `是` | 启用邮件提醒功能 |

**⚠️ 重要**：
- 此选项启用后，工作流状态转换时会自动发送邮件
- 但需要在工作流动作中配置具体的邮件内容

### 7.3 配置工作流动作中的邮件通知

#### 步骤 1：进入工作流动作配置页面

在工作流编辑页面，点击"工作流动作"（Workflow Actions）标签页。

#### 步骤 2：添加邮件通知动作

点击"添加行"（Add Row）按钮，为每个需要发送邮件的转换添加邮件通知动作。

**动作配置界面字段说明**：

| 字段 | 说明 | 示例 |
|------|------|------|
| **动作名称** | 动作的标识名称 | `Send Approval Email` |
| **状态** | 触发此动作的状态 | `Pending Level 1 Approval` |
| **动作** | 动作类型 | `Email Notification` |
| **收件人** | 邮件收件人（支持多种方式） | 见下方详细说明 |
| **主题** | 邮件主题（支持变量） | `采购订单待审批：{{ doc.name }}` |
| **消息** | 邮件内容（支持变量和HTML） | 见下方详细说明 |

#### 步骤 3：配置收件人

**收件人配置方式**（按优先级）：

**方式 1：通过角色获取用户** ⭐ **推荐**

```
收件人类型：Role（角色）
角色名称：Purchase Manager
说明：系统会自动获取该角色的所有用户，发送邮件给所有审批人员
```

**方式 2：通过文档字段获取**

```
收件人类型：Document Field（文档字段）
字段名：owner
说明：发送给订单创建人（doc.owner）
```

**方式 3：通过用户列表**

```
收件人类型：User List（用户列表）
用户列表：user1@example.com, user2@example.com
说明：直接指定收件人邮箱地址（多个用逗号分隔）
```

**方式 4：通过 Python 表达式** ⭐ **灵活**

```
收件人类型：Python Expression（Python 表达式）
表达式：get_users_with_role("Purchase Manager")
说明：使用 Python 代码动态获取收件人
```

**实际配置示例**：

**示例 1：提交审批时通知审批人员**

```
动作名称：Submit for Level 1 Approval Email
状态：Pending Level 1 Approval
动作：Email Notification
收件人：Purchase Manager（角色）
主题：采购订单待审批：{{ doc.name }}（金额：{{ doc.grand_total }}）
消息：
  采购订单 {{ doc.name }} 已提交审批，请及时处理。
  
  订单信息：
  - 订单号：{{ doc.name }}
  - 供应商：{{ doc.supplier }}
  - 订单金额：{{ doc.grand_total }}
  - 提交人：{{ doc.owner }}
  
  请点击以下链接查看详情：
  {{ doc.get_url() }}
```

**示例 2：审批通过时通知创建人**

```
动作名称：Approve Level 1 Email
状态：Approved（从 Pending Level 1 Approval 转换）
动作：Email Notification
收件人：{{ doc.owner }}（文档字段）
主题：采购订单已批准：{{ doc.name }}
消息：
  恭喜！采购订单 {{ doc.name }} 已通过一级审批。
  
  订单信息：
  - 订单号：{{ doc.name }}
  - 供应商：{{ doc.supplier }}
  - 订单金额：{{ doc.grand_total }}
  - 审批人：{{ frappe.session.user }}
  
  现在可以继续后续流程（创建收货单、发票等）。
  
  查看订单：{{ doc.get_url() }}
```

**示例 3：审批拒绝时通知创建人**

```
动作名称：Reject Email
状态：Rejected（从 Pending Level 1 Approval 转换）
动作：Email Notification
收件人：{{ doc.owner }}（文档字段）
主题：采购订单已拒绝：{{ doc.name }}
消息：
  采购订单 {{ doc.name }} 已被拒绝，请修改后重新提交。
  
  订单信息：
  - 订单号：{{ doc.name }}
  - 供应商：{{ doc.supplier }}
  - 订单金额：{{ doc.grand_total }}
  - 拒绝原因：{{ doc.reason_for_rejection or "未提供" }}
  
  请修改订单后重新提交审批。
  
  查看订单：{{ doc.get_url() }}
```

**示例 4：多级审批流转通知**

```
动作名称：Approve Level 1 → Level 2 Email
状态：Pending Level 2 Approval（从 Pending Level 1 Approval 转换）
动作：Email Notification
收件人：Accounts Manager（角色）
主题：采购订单待二级审批：{{ doc.name }}
消息：
  采购订单 {{ doc.name }} 已通过一级审批，现进入二级审批流程。
  
  订单信息：
  - 订单号：{{ doc.name }}
  - 供应商：{{ doc.supplier }}
  - 订单金额：{{ doc.grand_total }}
  - 一级审批人：{{ doc.approver_level_1 }}
  
  请及时处理二级审批。
  
  查看订单：{{ doc.get_url() }}
```

### 7.4 邮件模板变量

ERPNext 工作流邮件支持使用文档变量，常用变量如下：

**文档字段变量**：

| 变量 | 说明 | 示例 |
|------|------|------|
| `{{ doc.name }}` | 文档名称（订单号） | `PO-00001` |
| `{{ doc.supplier }}` | 供应商名称 | `ABC Company` |
| `{{ doc.grand_total }}` | 订单总金额 | `50000.00` |
| `{{ doc.owner }}` | 订单创建人 | `user@example.com` |
| `{{ doc.creation }}` | 创建时间 | `2025-01-17 10:30:00` |
| `{{ doc.modified }}` | 修改时间 | `2025-01-17 14:20:00` |
| `{{ doc.workflow_state }}` | 工作流状态 | `Pending Level 1 Approval` |

**系统变量**：

| 变量 | 说明 | 示例 |
|------|------|------|
| `{{ frappe.session.user }}` | 当前操作用户 | `approver@example.com` |
| `{{ doc.get_url() }}` | 文档访问链接 | `https://erp.example.com/app/purchase-order/PO-00001` |

**Python 表达式变量**：

```python
# 获取审批人信息
{{ doc.approver_level_1 or "待审批" }}

# 格式化金额
{{ frappe.utils.fmt_money(doc.grand_total, currency=doc.currency) }}

# 格式化日期
{{ frappe.utils.format_date(doc.creation) }}

# 条件判断
{% if doc.grand_total > 50000 %}
  此订单金额较大，需要三级审批。
{% else %}
  此订单金额较小，只需一级审批。
{% endif %}
```

### 7.5 邮件内容格式

**纯文本格式**：

```
采购订单 {{ doc.name }} 已提交审批，请及时处理。

订单信息：
- 订单号：{{ doc.name }}
- 供应商：{{ doc.supplier }}
- 订单金额：{{ doc.grand_total }}

查看订单：{{ doc.get_url() }}
```

**HTML 格式**（推荐，更美观）：

```html
<div style="font-family: Arial, sans-serif;">
  <h2 style="color: #2e7d32;">采购订单待审批</h2>
  
  <p>采购订单 <strong>{{ doc.name }}</strong> 已提交审批，请及时处理。</p>
  
  <table style="border-collapse: collapse; width: 100%; margin: 20px 0;">
    <tr>
      <td style="padding: 8px; border: 1px solid #ddd;"><strong>订单号</strong></td>
      <td style="padding: 8px; border: 1px solid #ddd;">{{ doc.name }}</td>
    </tr>
    <tr>
      <td style="padding: 8px; border: 1px solid #ddd;"><strong>供应商</strong></td>
      <td style="padding: 8px; border: 1px solid #ddd;">{{ doc.supplier }}</td>
    </tr>
    <tr>
      <td style="padding: 8px; border: 1px solid #ddd;"><strong>订单金额</strong></td>
      <td style="padding: 8px; border: 1px solid #ddd;">{{ frappe.utils.fmt_money(doc.grand_total, currency=doc.currency) }}</td>
    </tr>
  </table>
  
  <p>
    <a href="{{ doc.get_url() }}" style="background-color: #1976d2; color: white; padding: 10px 20px; text-decoration: none; border-radius: 4px;">
      查看订单详情
    </a>
  </p>
</div>
```

### 7.6 完整邮件通知配置示例

**场景：三级审批流程的完整邮件通知配置**

#### 配置 1：提交审批时通知一级审批人员

```
动作名称：Submit for Level 1 Approval Email
状态：Pending Level 1 Approval
动作：Email Notification
收件人：Purchase Manager（角色）
主题：采购订单待一级审批：{{ doc.name }}（金额：{{ frappe.utils.fmt_money(doc.grand_total, currency=doc.currency) }}）
消息：
  <div style="font-family: Arial, sans-serif;">
    <h2 style="color: #ff9800;">采购订单待审批</h2>
    <p>采购订单 <strong>{{ doc.name }}</strong> 已提交，等待一级审批。</p>
    <table style="border-collapse: collapse; width: 100%; margin: 20px 0;">
      <tr><td style="padding: 8px; border: 1px solid #ddd;"><strong>订单号</strong></td><td style="padding: 8px; border: 1px solid #ddd;">{{ doc.name }}</td></tr>
      <tr><td style="padding: 8px; border: 1px solid #ddd;"><strong>供应商</strong></td><td style="padding: 8px; border: 1px solid #ddd;">{{ doc.supplier }}</td></tr>
      <tr><td style="padding: 8px; border: 1px solid #ddd;"><strong>订单金额</strong></td><td style="padding: 8px; border: 1px solid #ddd;">{{ frappe.utils.fmt_money(doc.grand_total, currency=doc.currency) }}</td></tr>
      <tr><td style="padding: 8px; border: 1px solid #ddd;"><strong>提交人</strong></td><td style="padding: 8px; border: 1px solid #ddd;">{{ doc.owner }}</td></tr>
    </table>
    <p><a href="{{ doc.get_url() }}" style="background-color: #1976d2; color: white; padding: 10px 20px; text-decoration: none; border-radius: 4px;">查看订单详情</a></p>
  </div>
```

#### 配置 2：一级审批通过，进入二级审批时通知

```
动作名称：Approve Level 1 → Level 2 Email
状态：Pending Level 2 Approval
动作：Email Notification
收件人：Accounts Manager（角色）
主题：采购订单待二级审批：{{ doc.name }}
消息：
  <div style="font-family: Arial, sans-serif;">
    <h2 style="color: #ff9800;">采购订单待二级审批</h2>
    <p>采购订单 <strong>{{ doc.name }}</strong> 已通过一级审批，现进入二级审批流程。</p>
    <p style="color: #2e7d32;">✓ 一级审批已通过（审批人：{{ frappe.session.user }}）</p>
    <p><a href="{{ doc.get_url() }}" style="background-color: #1976d2; color: white; padding: 10px 20px; text-decoration: none; border-radius: 4px;">查看订单详情</a></p>
  </div>
```

#### 配置 3：二级审批通过，进入三级审批时通知

```
动作名称：Approve Level 2 → Level 3 Email
状态：Pending Level 3 Approval
动作：Email Notification
收件人：General Manager（角色）
主题：采购订单待三级审批：{{ doc.name }}
消息：
  <div style="font-family: Arial, sans-serif;">
    <h2 style="color: #ff9800;">采购订单待三级审批</h2>
    <p>采购订单 <strong>{{ doc.name }}</strong> 已通过二级审批，现进入三级审批流程。</p>
    <p style="color: #2e7d32;">✓ 一级审批已通过</p>
    <p style="color: #2e7d32;">✓ 二级审批已通过（审批人：{{ frappe.session.user }}）</p>
    <p><a href="{{ doc.get_url() }}" style="background-color: #1976d2; color: white; padding: 10px 20px; text-decoration: none; border-radius: 4px;">查看订单详情</a></p>
  </div>
```

#### 配置 4：三级审批通过，通知创建人

```
动作名称：Approve Level 3 Email
状态：Approved
动作：Email Notification
收件人：{{ doc.owner }}（文档字段）
主题：采购订单已批准：{{ doc.name }}
消息：
  <div style="font-family: Arial, sans-serif;">
    <h2 style="color: #2e7d32;">✓ 采购订单已批准</h2>
    <p>恭喜！采购订单 <strong>{{ doc.name }}</strong> 已通过所有审批，可以继续后续流程。</p>
    <p style="color: #2e7d32;">✓ 一级审批已通过</p>
    <p style="color: #2e7d32;">✓ 二级审批已通过</p>
    <p style="color: #2e7d32;">✓ 三级审批已通过（审批人：{{ frappe.session.user }}）</p>
    <p><strong>下一步操作：</strong>可以创建收货单（Purchase Receipt）和发票（Purchase Invoice）。</p>
    <p><a href="{{ doc.get_url() }}" style="background-color: #2e7d32; color: white; padding: 10px 20px; text-decoration: none; border-radius: 4px;">查看订单详情</a></p>
  </div>
```

#### 配置 5：审批拒绝时通知创建人

```
动作名称：Reject Email
状态：Rejected
动作：Email Notification
收件人：{{ doc.owner }}（文档字段）
主题：采购订单已拒绝：{{ doc.name }}
消息：
  <div style="font-family: Arial, sans-serif;">
    <h2 style="color: #d32f2f;">✗ 采购订单已拒绝</h2>
    <p>采购订单 <strong>{{ doc.name }}</strong> 已被拒绝，请修改后重新提交。</p>
    <p><strong>拒绝原因：</strong>{{ doc.reason_for_rejection or "未提供" }}</p>
    <p><strong>拒绝人：</strong>{{ frappe.session.user }}</p>
    <p><a href="{{ doc.get_url() }}" style="background-color: #d32f2f; color: white; padding: 10px 20px; text-decoration: none; border-radius: 4px;">修改订单</a></p>
  </div>
```

### 7.7 测试邮件通知

#### 测试步骤

1. **创建测试订单**
   - 创建一个采购订单，金额设置为 80,000（触发三级审批）
   - 提交订单

2. **验证一级审批邮件**
   - 使用采购经理账号登录
   - 检查是否收到"待一级审批"邮件
   - 验证邮件内容是否正确

3. **验证二级审批邮件**
   - 一级审批通过后
   - 使用财务经理账号登录
   - 检查是否收到"待二级审批"邮件

4. **验证三级审批邮件**
   - 二级审批通过后
   - 使用总经理账号登录
   - 检查是否收到"待三级审批"邮件

5. **验证审批通过邮件**
   - 三级审批通过后
   - 使用订单创建人账号登录
   - 检查是否收到"已批准"邮件

6. **验证拒绝邮件**
   - 在任意级别拒绝订单
   - 使用订单创建人账号登录
   - 检查是否收到"已拒绝"邮件

#### 常见问题排查

**问题 1：邮件未发送** ⭐ **最常见问题**

**系统性排查步骤**（按优先级）：

**步骤 1：检查邮件服务器配置** ⭐ **最高优先级**

1. **检查邮件账户是否启用**
   - 导航：`设置 → 集成 → 邮箱账户`
   - 确认邮件账户状态为"启用"（Enabled = ✅）
   - 如果未启用，启用后保存

2. **测试邮件服务器连接**
   - 在邮件账户页面，点击"测试连接"按钮
   - 如果测试失败，检查：
     - SMTP 服务器地址是否正确
     - SMTP 端口是否正确（587/465）
     - 用户名和密码是否正确
     - TLS/SSL 配置是否正确
     - 防火墙是否允许出站连接

3. **检查邮件账户是否为默认账户**
   - 确保至少有一个邮件账户标记为"默认"（Default = ✅）
   - 如果有多个邮件账户，确保默认账户配置正确

**步骤 2：检查工作流配置**

1. **检查工作流是否启用**
   - 导航：`设置 → 工作流 → [你的工作流名称]`
   - 确认工作流状态为"启用"（Enabled = ✅）
   - 确认工作流文档类型匹配（如：Purchase Order）

2. **检查工作流是否启用邮件提醒**
   - 在工作流基本信息中，确认"发送电子邮件提醒" = ✅ `是`
   - ⚠️ **关键**：即使启用了邮件提醒，也必须配置工作流动作中的邮件通知

**步骤 3：检查工作流动作配置** ⭐ **关键步骤**

1. **确认工作流动作存在**
   - 在工作流编辑页面，点击"工作流动作"标签页
   - 确认有邮件通知动作（Action = `Email Notification`）
   - 确认动作的状态（Status）与工作流状态匹配

2. **检查动作触发时机**
   - 确认动作的状态字段与当前文档的 `workflow_state` 匹配
   - 例如：如果文档状态是 `Pending Level 1 Approval`，动作的状态应该是 `Pending Level 1 Approval`

3. **检查动作配置完整性**
   - 动作名称：已填写
   - 状态：已选择正确的工作流状态
   - 动作：已选择 `Email Notification`
   - 收件人：已配置（不能为空）
   - 主题：已填写
   - 消息：已填写

**步骤 4：检查收件人配置** ⭐ **常见问题点**

1. **如果使用角色配置收件人**：
   ```
   检查项：
   - 角色名称是否正确（区分大小写）
   - 角色是否存在
   - 角色中是否有用户
   - 用户是否启用
   - 用户是否有邮箱地址
   ```

2. **如果使用文档字段配置收件人**：
   ```
   检查项：
   - 字段名是否正确（如：owner）
   - 字段值是否为有效的邮箱地址
   - 字段值是否为空
   ```

3. **如果使用用户列表配置收件人**：
   ```
   检查项：
   - 邮箱地址格式是否正确
   - 多个邮箱是否用逗号分隔
   - 邮箱地址是否存在拼写错误
   ```

**步骤 5：检查邮件队列状态** ⭐ **诊断关键**

1. **查看邮件队列**
   - 导航：`设置 → 邮箱 → 邮件队列` 或直接访问 `/app/email-queue`
   - 查看是否有待发送的邮件
   - 查看邮件状态：
     - `Queued`：已排队，等待发送
     - `Sent`：已发送
     - `Error`：发送失败（查看错误信息）
     - `Not Sent`：未发送（查看原因）

2. **检查邮件队列错误**
   - 如果状态为 `Error`，点击查看错误详情
   - 常见错误：
     - `SMTP Authentication failed`：认证失败，检查用户名密码
     - `Connection timeout`：连接超时，检查网络和防火墙
     - `Invalid recipient`：收件人无效，检查邮箱地址
     - `Email account not found`：邮件账户未找到，检查默认邮件账户

3. **手动重试发送**
   - 在邮件队列中，找到失败的邮件
   - 点击"重试"按钮
   - 观察是否成功发送

**步骤 6：检查工作流状态转换**

1. **确认工作流状态已转换**
   - 打开触发邮件的文档（如：Purchase Order）
   - 查看 `workflow_state` 字段
   - 确认状态已从初始状态转换到目标状态
   - 例如：从 `Submitted` 转换到 `Pending Level 1 Approval`

2. **检查状态转换是否成功**
   - 如果状态未转换，邮件不会触发
   - 检查工作流转换配置：
     - 转换条件是否满足
     - 用户是否有权限执行转换
     - 转换是否被其他条件阻止

**步骤 7：检查系统日志**

1. **查看 ERPNext 日志**
   - 导航：`设置 → 日志查看器` 或直接访问 `/app/log-viewer`
   - 搜索关键词：`workflow`、`email`、`notification`
   - 查看是否有错误信息

2. **查看服务器日志**
   - 如果 ERPNext 运行在服务器上，检查服务器日志
   - 查看是否有 SMTP 连接错误
   - 查看是否有 Python 异常

**步骤 8：验证邮件发送权限**

1. **检查用户权限**
   - 确认执行工作流操作的用户有发送邮件的权限
   - 检查角色权限设置

2. **检查系统设置**
   - 导航：`设置 → 系统设置`
   - 确认邮件发送功能未被禁用

**快速诊断命令**（适用于有服务器访问权限的情况）：

```bash
# 检查 ERPNext 邮件队列（通过 bench 命令）
bench --site [site-name] console

# 在 Python 控制台中执行
import frappe
from frappe.email.queue import flush

# 查看待发送的邮件
emails = frappe.get_all("Email Queue", filters={"status": "Queued"}, limit=10)
for email in emails:
    print(frappe.get_doc("Email Queue", email.name).as_dict())

# 手动刷新邮件队列
flush(from_test=True)
```

**问题 2：邮件内容显示变量名而非实际值**

**症状**：
- 邮件中显示 `{{ doc.name }}` 而不是实际的订单号

**排查步骤**：
1. 检查邮件模板中的变量语法是否正确
2. 检查变量名是否正确（区分大小写）

**解决方案**：
- 确保变量使用双大括号：`{{ doc.name }}`
- 确保变量名正确，区分大小写
- 使用 `doc.get_url()` 而不是 `doc.url`
- 注意：变量名区分大小写，`{{ doc.name }}` 和 `{{ doc.Name }}` 是不同的

**问题 3：收件人未收到邮件**

**排查步骤**：
1. 检查收件人配置方式是否正确
2. 检查角色中是否有用户
3. 检查用户邮箱地址是否正确
4. 检查邮件是否被垃圾邮件过滤器拦截

**解决方案**：
- 使用角色配置时，确保角色中有用户
- 使用文档字段配置时，确保字段值是正确的邮箱地址
- 检查垃圾邮件文件夹
- 在 ERPNext 中查看邮件队列（Email Queue）状态
- 检查邮件是否被邮件服务商拦截（查看邮件队列错误信息）

**问题 4：邮件发送延迟**

**症状**：
- 邮件发送成功，但收件人收到时间延迟

**排查步骤**：
1. 检查邮件队列处理频率
2. 检查服务器性能
3. 检查邮件服务商限制

**解决方案**：
- ERPNext 默认通过后台任务发送邮件，可能有延迟
- 可以手动刷新邮件队列：`设置 → 邮箱 → 邮件队列 → 刷新`
- 检查后台任务（Scheduler）是否正常运行

---

## 八、相关文档

- [ERPNEXT 工作流官方文档](https://docs.erpnext.com/docs/user/manual/en/setting-up/workflows)
- [销售订单审批工作流配置指南](./sales-order-approval-workflow.md)
- [销售采购审批工作流](../business/workflows/sales-purchase-approval-flow.md)
- [ERPNEXT API 文档](./erpnext-api.md)

---

## 九、两级审批工作流配置示例（按金额分配审批人）

### 9.1 业务需求

**审批规则**：
- 订单金额 **≤ 2,000,000**：由**采购经理**审批
- 订单金额 **> 2,000,000**：由**财务经理**审批

### 9.2 工作流状态配置（Workflow States）

在工作流编辑页面的"States"标签页，添加以下状态：

| No. | State* | Doc Status | Update Field | Update Value | Only Allow Edit For* |
|-----|--------|-----------|--------------|--------------|---------------------|
| 1 | `draft` | `0` | （留空） | （留空） | `All` |
| 2 | `Submitted` | `0` | （留空） | （留空） | `Purchase User` |
| 3 | `Pending Purchase Manager Approval` | `0` | （留空） | （留空） | `Purchase Manager` |
| 4 | `Pending Finance Manager Approval` | `0` | （留空） | （留空） | `Accounts Manager` |
| 5 | `Approved` | `1` ⭐ | （留空） | （留空） | `System Manager` |
| 6 | `Rejected` | `0` | （留空） | （留空） | `Purchase User` |

**状态说明**：

| 状态标识 | 说明 | Doc Status | 允许编辑角色 |
|---------|------|-----------|------------|
| `draft` | 草稿状态 | 0 | All |
| `Submitted` | 已提交，等待分配审批 | 0 | Purchase User |
| `Pending Purchase Manager Approval` | 待采购经理审批（金额 ≤ 200万） | 0 | Purchase Manager |
| `Pending Finance Manager Approval` | 待财务经理审批（金额 > 200万） | 0 | Accounts Manager |
| `Approved` | 已批准 ⭐ | 1 | System Manager |
| `Rejected` | 已拒绝 | 0 | Purchase User |

### 9.3 工作流转换配置（Workflow Transitions）

在工作流编辑页面的"Transition Rules"标签页，添加以下转换：

#### 转换规则表格

| No. | State* | Action* | Next State* | Allowed* | Condition |
|-----|--------|---------|-------------|----------|-----------|
| 1 | `draft` | `Submit` | `Submitted` | `Purchase User, Purchase Manager` | （留空） |
| 2 | `Submitted` | `Submit for Purchase Manager Approval` | `Pending Purchase Manager Approval` | `Purchase User` | `doc.grand_total <= 2000000` |
| 3 | `Submitted` | `Submit for Finance Manager Approval` | `Pending Finance Manager Approval` | `Purchase User` | `doc.grand_total > 2000000` |
| 4 | `Pending Purchase Manager Approval` | `Approve Purchase Manager` | `Approved` | `Purchase Manager` | （留空） |
| 5 | `Pending Finance Manager Approval` | `Approve Finance Manager` | `Approved` | `Accounts Manager` | （留空） |
| 6 | `Pending Purchase Manager Approval` | `Reject` | `Rejected` | `Purchase Manager` | （留空） |
| 7 | `Pending Finance Manager Approval` | `Reject` | `Rejected` | `Accounts Manager` | （留空） |
| 8 | `Rejected` | `Re-submit` | `Submitted` | `Purchase User` | （留空） |

#### 详细转换说明

**转换 1：提交订单**

```
动作（Action）：Submit
当前状态（Current State）：draft
下一状态（Next State）：Submitted
条件（Condition）：留空（无条件）
允许的角色（Allowed Roles）：Purchase User, Purchase Manager
```

**转换 2：金额 ≤ 200万，提交到采购经理审批** ⭐

```
动作（Action）：Submit for Purchase Manager Approval
当前状态（Current State）：Submitted
下一状态（Next State）：Pending Purchase Manager Approval
条件（Condition）：doc.grand_total <= 2000000
允许的角色（Allowed Roles）：Purchase User

说明：
- 当订单金额 ≤ 2,000,000 时，自动进入采购经理审批流程
- 系统会根据条件表达式自动选择此转换
```

**转换 3：金额 > 200万，提交到财务经理审批** ⭐

```
动作（Action）：Submit for Finance Manager Approval
当前状态（Current State）：Submitted
下一状态（Next State）：Pending Finance Manager Approval
条件（Condition）：doc.grand_total > 2000000
允许的角色（Allowed Roles）：Purchase User

说明：
- 当订单金额 > 2,000,000 时，自动进入财务经理审批流程
- 系统会根据条件表达式自动选择此转换
```

**转换 4：采购经理审批通过**

```
动作（Action）：Approve Purchase Manager
当前状态（Current State）：Pending Purchase Manager Approval
下一状态（Next State）：Approved
条件（Condition）：留空（无条件）
允许的角色（Allowed Roles）：Purchase Manager

说明：
- 采购经理审批通过后，订单直接进入 Approved 状态
- 无需其他审批
```

**转换 5：财务经理审批通过**

```
动作（Action）：Approve Finance Manager
当前状态（Current State）：Pending Finance Manager Approval
下一状态（Next State）：Approved
条件（Condition）：留空（无条件）
允许的角色（Allowed Roles）：Accounts Manager

说明：
- 财务经理审批通过后，订单直接进入 Approved 状态
- 无需其他审批
```

**转换 6：采购经理审批拒绝**

```
动作（Action）：Reject
当前状态（Current State）：Pending Purchase Manager Approval
下一状态（Next State）：Rejected
条件（Condition）：留空（无条件）
允许的角色（Allowed Roles）：Purchase Manager
```

**转换 7：财务经理审批拒绝**

```
动作（Action）：Reject
当前状态（Current State）：Pending Finance Manager Approval
下一状态（Next State）：Rejected
条件（Condition）：留空（无条件）
允许的角色（Allowed Roles）：Accounts Manager
```

**转换 8：重新提交审批**

```
动作（Action）：Re-submit
当前状态（Current State）：Rejected
下一状态（Next State）：Submitted
条件（Condition）：留空（无条件）
允许的角色（Allowed Roles）：Purchase User

说明：
- 被拒绝的订单修改后重新提交
- 系统会根据新的订单金额重新分配审批流程
```

### 9.4 工作流动作配置（Workflow Actions）

在工作流编辑页面的"Workflow Actions"标签页，添加以下动作：

#### 动作 1：采购经理审批通过（更新 docstatus）⭐

```
动作名称：Approve Purchase Manager
状态：Approved（从 Pending Purchase Manager Approval 转换）

动作类型 1：Update Field（更新字段）- workflow_state
字段：workflow_state
值：Approved
说明：更新工作流状态为"已批准"

动作类型 2：Update Field（更新字段）- docstatus ⭐ 关键动作
字段：docstatus
值：1
说明：将文档状态更新为"已提交"，允许创建后续单据

动作类型 3：Email Notification（邮件通知）
收件人：{{ doc.owner }}（文档字段）
主题：采购订单已批准：{{ doc.name }}
内容：
  采购订单 {{ doc.name }} 已通过采购经理审批。
  
  订单信息：
  - 订单号：{{ doc.name }}
  - 供应商：{{ doc.supplier }}
  - 订单金额：{{ doc.grand_total }}
  - 审批人：{{ frappe.session.user }}
  
  现在可以继续后续流程（创建收货单、发票等）。
  
  查看订单：{{ doc.get_url() }}
```

#### 动作 2：财务经理审批通过（更新 docstatus）⭐

```
动作名称：Approve Finance Manager
状态：Approved（从 Pending Finance Manager Approval 转换）

动作类型 1：Update Field（更新字段）- workflow_state
字段：workflow_state
值：Approved
说明：更新工作流状态为"已批准"

动作类型 2：Update Field（更新字段）- docstatus ⭐ 关键动作
字段：docstatus
值：1
说明：将文档状态更新为"已提交"，允许创建后续单据

动作类型 3：Email Notification（邮件通知）
收件人：{{ doc.owner }}（文档字段）
主题：采购订单已批准：{{ doc.name }}
内容：
  采购订单 {{ doc.name }} 已通过财务经理审批。
  
  订单信息：
  - 订单号：{{ doc.name }}
  - 供应商：{{ doc.supplier }}
  - 订单金额：{{ doc.grand_total }}
  - 审批人：{{ frappe.session.user }}
  
  现在可以继续后续流程（创建收货单、发票等）。
  
  查看订单：{{ doc.get_url() }}
```

#### 动作 3：提交到采购经理审批时发送邮件通知

```
动作名称：Submit for Purchase Manager Approval Email
状态：Pending Purchase Manager Approval
动作：Email Notification
收件人：Purchase Manager（角色）
主题：采购订单待审批：{{ doc.name }}（金额：{{ doc.grand_total }}）
消息：
  采购订单 {{ doc.name }} 已提交采购经理审批，请及时处理。
  
  订单信息：
  - 订单号：{{ doc.name }}
  - 供应商：{{ doc.supplier }}
  - 订单金额：{{ doc.grand_total }}
  - 提交人：{{ doc.owner }}
  
  请点击以下链接查看详情：
  {{ doc.get_url() }}
```

#### 动作 4：提交到财务经理审批时发送邮件通知

```
动作名称：Submit for Finance Manager Approval Email
状态：Pending Finance Manager Approval
动作：Email Notification
收件人：Accounts Manager（角色）
主题：采购订单待财务经理审批：{{ doc.name }}（金额：{{ doc.grand_total }}）
消息：
  采购订单 {{ doc.name }} 已提交财务经理审批，请及时处理。
  
  订单信息：
  - 订单号：{{ doc.name }}
  - 供应商：{{ doc.supplier }}
  - 订单金额：{{ doc.grand_total }}
  - 提交人：{{ doc.owner }}
  
  请点击以下链接查看详情：
  {{ doc.get_url() }}
```

#### 动作 5：审批拒绝时发送邮件通知

```
动作名称：Reject Email
状态：Rejected（从 Pending Purchase Manager Approval 或 Pending Finance Manager Approval 转换）
动作：Email Notification
收件人：{{ doc.owner }}（文档字段）
主题：采购订单已拒绝：{{ doc.name }}
消息：
  采购订单 {{ doc.name }} 已被拒绝，请修改后重新提交。
  
  订单信息：
  - 订单号：{{ doc.name }}
  - 供应商：{{ doc.supplier }}
  - 订单金额：{{ doc.grand_total }}
  - 拒绝原因：{{ doc.reason_for_rejection or "未提供" }}
  - 拒绝人：{{ frappe.session.user }}
  
  请修改订单后重新提交审批。
  
  查看订单：{{ doc.get_url() }}
```

### 9.5 工作流流程图

```
订单创建（Draft）
    ↓
提交订单（Submit）
    ↓
已提交（Submitted）
    ↓
    ├─ 金额 ≤ 200万 → 采购经理审批（Pending Purchase Manager Approval）
    │                      ↓
    │                  审批通过 → 已批准（Approved，docstatus = 1）
    │                      ↓
    │                  审批拒绝 → 已拒绝（Rejected）
    │
    └─ 金额 > 200万 → 财务经理审批（Pending Finance Manager Approval）
                           ↓
                       审批通过 → 已批准（Approved，docstatus = 1）
                           ↓
                       审批拒绝 → 已拒绝（Rejected）
```

### 9.6 配置要点总结

**关键配置项**：

| 配置项 | 值 | 说明 |
|--------|-----|------|
| **金额条件 1** | `doc.grand_total <= 2000000` | 金额 ≤ 200万，采购经理审批 |
| **金额条件 2** | `doc.grand_total > 2000000` | 金额 > 200万，财务经理审批 |
| **采购经理角色** | `Purchase Manager` | 审批金额 ≤ 200万的订单 |
| **财务经理角色** | `Accounts Manager` | 审批金额 > 200万的订单 |
| **Approved 状态** | `docstatus = 1` | ⭐ 关键：允许创建后续单据 |

**注意事项**：

1. **金额条件不重叠**：`<= 2000000` 和 `> 2000000` 覆盖所有情况
2. **角色必须存在**：确保 `Purchase Manager` 和 `Accounts Manager` 角色存在且有用户
3. **docstatus 更新**：审批通过动作必须更新 `docstatus = 1`
4. **邮件通知**：确保已配置邮件服务器，且工作流启用了邮件提醒

### 9.7 测试验证

**测试场景 1：金额 ≤ 200万**

1. 创建采购订单，金额设置为 1,500,000
2. 提交订单
3. 验证：订单状态应为 `Pending Purchase Manager Approval`
4. 使用采购经理账号登录
5. 验证：采购经理可以看到待审批订单
6. 审批通过
7. 验证：订单状态应为 `Approved`，`docstatus = 1`
8. 验证：可以创建收货单、发票等后续单据

**测试场景 2：金额 > 200万**

1. 创建采购订单，金额设置为 2,500,000
2. 提交订单
3. 验证：订单状态应为 `Pending Finance Manager Approval`
4. 使用财务经理账号登录
5. 验证：财务经理可以看到待审批订单
6. 审批通过
7. 验证：订单状态应为 `Approved`，`docstatus = 1`
8. 验证：可以创建收货单、发票等后续单据

**测试场景 3：边界值测试**

1. 创建采购订单，金额设置为 2,000,000（等于边界值）
2. 提交订单
3. 验证：订单状态应为 `Pending Purchase Manager Approval`（≤ 200万）

4. 创建采购订单，金额设置为 2,000,001（大于边界值）
5. 提交订单
6. 验证：订单状态应为 `Pending Finance Manager Approval`（> 200万）

---

## 十、优化版两级审批工作流配置（按金额自动路由）

### 10.1 业务需求

**审批规则**：
- 订单金额 **< 100,000**：由**采购经理（Purchase Manager）**审批
- 订单金额 **≥ 100,000**：由**VP**审批

**优化说明**：
- ✅ **移除中间状态**：不再需要"Pending review"状态，提交后直接根据金额进入对应审批状态
- ✅ **自动判断**：提交时系统根据金额自动选择审批路径，无需用户手动选择"Submit for PM Approval"或"Submit for VP Approval"
- ✅ **简化流程**：减少用户操作步骤，提升效率

### 10.2 工作流状态配置（Workflow States）

在工作流编辑页面的"States"标签页，添加以下状态：

| No. | State* | Doc Status | Update Field | Update Value | Only Allow Edit For* |
|-----|--------|-----------|--------------|--------------|---------------------|
| 1 | `Draft` | `0` | （留空） | （留空） | `All` |
| 2 | `Pending PMA` | `0` | （留空） | （留空） | `Purchase Manager` |
| 3 | `Pending VP` | `0` | （留空） | （留空） | `VP` |
| 4 | `Approved` | `1` ⭐ | （留空） | （留空） | `System Manager` |
| 5 | `Rejected` | `0` | （留空） | （留空） | `Purchase User` |

**状态说明**：

| 状态标识 | 说明 | Doc Status | 允许编辑角色 |
|---------|------|-----------|------------|
| `Draft` | 草稿状态 | 0 | All |
| `Pending PMA` | 待采购经理审批（金额 < 100,000） | 0 | Purchase Manager |
| `Pending VP` | 待VP审批（金额 ≥ 100,000） | 0 | VP |
| `Approved` | 已批准 ⭐ | 1 | System Manager |
| `Rejected` | 已拒绝 | 0 | Purchase User |

**⚠️ 关键配置点**：
- `Approved` 状态的 `Doc Status` 必须设置为 `1` ⭐ **最关键**
- 其他审批状态的 `Doc Status` 都设置为 `0`
- `Approved` 状态的 `Only Allow Edit For` 设置为 `System Manager`（必填项处理）

### 10.3 工作流转换配置（Workflow Transitions）

在工作流编辑页面的"Transition Rules"标签页，添加以下转换：

#### 转换规则表格

| No. | State* | Action* | Next State* | Allowed* | Condition |
|-----|--------|---------|-------------|----------|-----------|
| 1 | `Draft` | `Submit for Approval` | `Pending PMA` | `Purchase User` | `doc.grand_total < 100000` |
| 2 | `Draft` | `Submit for Approval` | `Pending VP` | `Purchase User` | `doc.grand_total >= 100000` |
| 3 | `Pending PMA` | `Approve` | `Approved` | `Purchase Manager` | （留空） |
| 4 | `Pending PMA` | `Reject` | `Rejected` | `Purchase Manager` | （留空） |
| 5 | `Pending VP` | `Approve` | `Approved` | `VP` | （留空） |
| 6 | `Pending VP` | `Reject` | `Rejected` | `VP` | （留空） |
| 7 | `Rejected` | `Submit for Approval` | `Pending PMA` | `Purchase User` | `doc.grand_total < 100000` |
| 8 | `Rejected` | `Submit for Approval` | `Pending VP` | `Purchase User` | `doc.grand_total >= 100000` |

#### 详细转换说明

**转换 1：金额 < 100,000，提交到采购经理审批** ⭐

```
动作（Action）：Submit for Approval
当前状态（Current State）：Draft
下一状态（Next State）：Pending PMA
条件（Condition）：doc.grand_total < 100000
允许的角色（Allowed Roles）：Purchase User

说明：
- 当订单金额 < 100,000 时，自动进入采购经理审批流程
- 系统会根据条件表达式自动选择此转换
- 用户只需点击"Submit for Approval"，无需手动选择审批路径
```

**转换 2：金额 ≥ 100,000，提交到VP审批** ⭐

```
动作（Action）：Submit for Approval
当前状态（Current State）：Draft
下一状态（Next State）：Pending VP
条件（Condition）：doc.grand_total >= 100000
允许的角色（Allowed Roles）：Purchase User

说明：
- 当订单金额 ≥ 100,000 时，自动进入VP审批流程
- 系统会根据条件表达式自动选择此转换
- 用户只需点击"Submit for Approval"，无需手动选择审批路径
```

**转换 3：采购经理审批通过**

```
动作（Action）：Approve
当前状态（Current State）：Pending PMA
下一状态（Next State）：Approved
条件（Condition）：留空（无条件）
允许的角色（Allowed Roles）：Purchase Manager

说明：
- 采购经理审批通过后，订单直接进入 Approved 状态
- 无需其他审批
```

**转换 4：采购经理审批拒绝**

```
动作（Action）：Reject
当前状态（Current State）：Pending PMA
下一状态（Next State）：Rejected
条件（Condition）：留空（无条件）
允许的角色（Allowed Roles）：Purchase Manager
```

**转换 5：VP审批通过**

```
动作（Action）：Approve
当前状态（Current State）：Pending VP
下一状态（Next State）：Approved
条件（Condition）：留空（无条件）
允许的角色（Allowed Roles）：VP

说明：
- VP审批通过后，订单直接进入 Approved 状态
- 无需其他审批
```

**转换 6：VP审批拒绝**

```
动作（Action）：Reject
当前状态（Current State）：Pending VP
下一状态（Next State）：Rejected
条件（Condition）：留空（无条件）
允许的角色（Allowed Roles）：VP
```

**转换 7-8：重新提交审批**

```
动作（Action）：Submit for Approval
当前状态（Current State）：Rejected
下一状态（Next State）：Pending PMA 或 Pending VP（根据金额判断）
条件（Condition）：
  - doc.grand_total < 100000 → Pending PMA
  - doc.grand_total >= 100000 → Pending VP
允许的角色（Allowed Roles）：Purchase User

说明：
- 被拒绝的订单修改后重新提交
- 系统会根据新的订单金额重新分配审批流程
```

### 10.4 工作流动作配置（Workflow Actions）

在工作流编辑页面的"Workflow Actions"标签页，添加以下动作：

#### 动作 1：采购经理审批通过（更新 docstatus）⭐

```
动作名称：Approve Purchase Manager
状态：Approved（从 Pending PMA 转换）

动作类型 1：Update Field（更新字段）- workflow_state
字段：workflow_state
值：Approved
说明：更新工作流状态为"已批准"

动作类型 2：Update Field（更新字段）- docstatus ⭐ 关键动作
字段：docstatus
值：1
说明：将文档状态更新为"已提交"，允许创建后续单据

动作类型 3：Email Notification（邮件通知）
收件人：{{ doc.owner }}（文档字段）
主题：采购订单已批准：{{ doc.name }}
内容：
  采购订单 {{ doc.name }} 已通过采购经理审批。
  
  订单信息：
  - 订单号：{{ doc.name }}
  - 供应商：{{ doc.supplier }}
  - 订单金额：{{ doc.grand_total }}
  - 审批人：{{ frappe.session.user }}
  
  现在可以继续后续流程（创建收货单、发票等）。
  
  查看订单：{{ doc.get_url() }}
```

#### 动作 2：VP审批通过（更新 docstatus）⭐

```
动作名称：Approve VP
状态：Approved（从 Pending VP 转换）

动作类型 1：Update Field（更新字段）- workflow_state
字段：workflow_state
值：Approved
说明：更新工作流状态为"已批准"

动作类型 2：Update Field（更新字段）- docstatus ⭐ 关键动作
字段：docstatus
值：1
说明：将文档状态更新为"已提交"，允许创建后续单据

动作类型 3：Email Notification（邮件通知）- 通知创建人
收件人：{{ doc.owner }}（文档字段）
主题：采购订单已批准：{{ doc.name }}
内容：
  采购订单 {{ doc.name }} 已通过VP审批。
  
  订单信息：
  - 订单号：{{ doc.name }}
  - 供应商：{{ doc.supplier }}
  - 订单金额：{{ doc.grand_total }}
  - 审批人：{{ frappe.session.user }}
  
  现在可以继续后续流程（创建收货单、发票等）。
  
  查看订单：{{ doc.get_url() }}

动作类型 4：Email Notification（邮件通知）- 通知采购经理 ⭐
收件人：Purchase Manager（角色）
主题：采购订单已通过VP审批：{{ doc.name }}
内容：
  采购订单 {{ doc.name }} 已通过VP审批，请知悉。
  
  订单信息：
  - 订单号：{{ doc.name }}
  - 供应商：{{ doc.supplier }}
  - 订单金额：{{ doc.grand_total }}
  - 审批人：{{ frappe.session.user }}
  - 审批时间：{{ frappe.utils.format_datetime(frappe.utils.now(), "yyyy-MM-dd HH:mm:ss") }}
  
  订单已批准，可以继续后续流程（创建收货单、发票等）。
  
  查看订单：{{ doc.get_url() }}
```

#### 动作 3：提交到采购经理审批时发送邮件通知

```
动作名称：Submit for Approval Email (PM)
状态：Pending PMA
动作：Email Notification
收件人：Purchase Manager（角色）
主题：采购订单待审批：{{ doc.name }}（金额：{{ doc.grand_total }}）
消息：
  采购订单 {{ doc.name }} 已提交采购经理审批，请及时处理。
  
  订单信息：
  - 订单号：{{ doc.name }}
  - 供应商：{{ doc.supplier }}
  - 订单金额：{{ doc.grand_total }}
  - 提交人：{{ doc.owner }}
  
  请点击以下链接查看详情：
  {{ doc.get_url() }}
```

#### 动作 4：提交到VP审批时发送邮件通知

```
动作名称：Submit for Approval Email (VP)
状态：Pending VP
动作：Email Notification
收件人：VP（角色）
主题：采购订单待VP审批：{{ doc.name }}（金额：{{ doc.grand_total }}）
消息：
  采购订单 {{ doc.name }} 已提交VP审批，请及时处理。
  
  订单信息：
  - 订单号：{{ doc.name }}
  - 供应商：{{ doc.supplier }}
  - 订单金额：{{ doc.grand_total }}
  - 提交人：{{ doc.owner }}
  
  请点击以下链接查看详情：
  {{ doc.get_url() }}
```

#### 动作 5：审批拒绝时发送邮件通知

```
动作名称：Reject Email
状态：Rejected（从 Pending PMA 或 Pending VP 转换）
动作：Email Notification
收件人：{{ doc.owner }}（文档字段）
主题：采购订单已拒绝：{{ doc.name }}
消息：
  采购订单 {{ doc.name }} 已被拒绝，请修改后重新提交。
  
  订单信息：
  - 订单号：{{ doc.name }}
  - 供应商：{{ doc.supplier }}
  - 订单金额：{{ doc.grand_total }}
  - 拒绝原因：{{ doc.reason_for_rejection or "未提供" }}
  - 拒绝人：{{ frappe.session.user }}
  
  请修改订单后重新提交审批。
  
  查看订单：{{ doc.get_url() }}
```

### 10.5 工作流状态流转图

```
订单创建（Draft）
    ↓
提交审批（Submit for Approval）
    ↓
    ├─ 金额 < 100,000 → 采购经理审批（Pending PMA）
    │                      ↓
    │                  审批通过 → 已批准（Approved，docstatus = 1）
    │                      ↓
    │                  审批拒绝 → 已拒绝（Rejected）
    │                      ↓
    │                  重新提交 → 根据金额重新分配审批路径
    │
    └─ 金额 ≥ 100,000 → VP审批（Pending VP）
                           ↓
                       审批通过 → 已批准（Approved，docstatus = 1）
                           ↓
                       审批拒绝 → 已拒绝（Rejected）
                           ↓
                       重新提交 → 根据金额重新分配审批路径
```

### 10.6 配置要点总结

**关键配置项**：

| 配置项 | 值 | 说明 |
|--------|-----|------|
| **金额条件 1** | `doc.grand_total < 100000` | 金额 < 100,000，采购经理审批 |
| **金额条件 2** | `doc.grand_total >= 100000` | 金额 ≥ 100,000，VP审批 |
| **采购经理角色** | `Purchase Manager` | 审批金额 < 100,000 的订单 |
| **VP角色** | `VP` | 审批金额 ≥ 100,000 的订单 |
| **Approved 状态** | `docstatus = 1` | ⭐ 关键：允许创建后续单据 |

**优化点**：

1. ✅ **移除中间状态**：不再需要"Pending review"状态
2. ✅ **移除手动选择动作**：不再需要"Submit for PM Approval"和"Submit for VP Approval"动作
3. ✅ **自动路由**：提交时系统根据金额自动选择审批路径
4. ✅ **简化操作**：用户只需点击"Submit for Approval"，系统自动判断

**注意事项**：

1. **金额条件不重叠**：`< 100000` 和 `>= 100000` 覆盖所有情况
2. **角色必须存在**：确保 `Purchase Manager` 和 `VP` 角色存在且有用户
3. **docstatus 更新**：审批通过动作必须更新 `docstatus = 1`
4. **邮件通知**：确保已配置邮件服务器，且工作流启用了邮件提醒

### 10.7 与原配置对比

**原配置流程**：
```
Draft → Submit for Approval → Pending review
                              ↓
                    Submit for PM Approval → Pending PMA
                    Submit for VP Approval → Pending VP
```

**优化后流程**：
```
Draft → Submit for Approval → Pending PMA（金额 < 100,000）
                            → Pending VP（金额 ≥ 100,000）
```

**优化效果**：
- ✅ 减少 1 个中间状态（Pending review）
- ✅ 减少 2 个手动选择动作（Submit for PM Approval、Submit for VP Approval）
- ✅ 用户操作步骤从 2 步减少到 1 步
- ✅ 系统自动判断，减少人为错误

### 10.8 测试验证

**测试场景 1：金额 < 100,000**

1. 创建采购订单，金额设置为 80,000
2. 点击"Submit for Approval"
3. **验证**：订单状态应为 `Pending PMA`（无需手动选择）
4. 使用采购经理账号登录
5. **验证**：采购经理可以看到待审批订单
6. 审批通过
7. **验证**：订单状态应为 `Approved`，`docstatus = 1`
8. **验证**：可以创建收货单、发票等后续单据

**测试场景 2：金额 ≥ 100,000**

1. 创建采购订单，金额设置为 150,000
2. 点击"Submit for Approval"
3. **验证**：订单状态应为 `Pending VP`（无需手动选择）
4. 使用VP账号登录
5. **验证**：VP可以看到待审批订单
6. 审批通过
7. **验证**：订单状态应为 `Approved`，`docstatus = 1`
8. **验证**：可以创建收货单、发票等后续单据

**测试场景 3：边界值测试**

1. 创建采购订单，金额设置为 99,999（小于边界值）
2. 点击"Submit for Approval"
3. **验证**：订单状态应为 `Pending PMA`（< 100,000）

4. 创建采购订单，金额设置为 100,000（等于边界值）
5. 点击"Submit for Approval"
6. **验证**：订单状态应为 `Pending VP`（≥ 100,000）

7. 创建采购订单，金额设置为 100,001（大于边界值）
8. 点击"Submit for Approval"
9. **验证**：订单状态应为 `Pending VP`（≥ 100,000）

**测试场景 4：拒绝后重新提交**

1. 创建采购订单，金额设置为 80,000
2. 提交审批，进入 `Pending PMA`
3. 采购经理拒绝订单
4. **验证**：订单状态应为 `Rejected`
5. 修改订单金额为 150,000
6. 重新提交审批
7. **验证**：订单状态应为 `Pending VP`（根据新金额重新分配）

---

**文档版本**：v1.2  
**创建时间**：2025-01-17  
**更新时间**：2025-01-17  
**维护者**：TTPOS Team

