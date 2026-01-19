# ERPNEXT 销售订单审批工作流配置指南

> 📖 **用途**: 在 ERPNEXT 销售订单的 Submitted（已提交）状态后增加人员审批流程

---

## 一、方案概述

ERPNEXT 支持通过**工作流（Workflow）**功能为销售订单添加审批流程。**关键机制**：工作流通过控制 `docstatus`（文档状态）来实现审批流程，只有 `docstatus = 1`（已提交）的订单才能创建后续单据（交货单、发票等）。

### ⚠️ 重要说明

**ERPNEXT 工作流的核心机制**：
- **不是**通过 `workflow_state` 来控制是否可以创建后续单据
- **而是**通过 `docstatus`（文档状态）来控制
- 只有 `docstatus = 1`（已提交）的订单才能创建交货单、发票等后续单据
- 工作流的作用是：**控制何时将 `docstatus` 从 0 更新为 1**

**正确的实现方式**：
1. 订单提交后，`docstatus` 保持为 `0`（草稿），进入"待审批"状态
2. 审批通过时，工作流动作将 `docstatus` 更新为 `1`（已提交）
3. 只有 `docstatus = 1` 时，才能创建后续单据

### 状态流转

**完整状态流转图**：

```
工作流状态（workflow_state）流转：
┌─────────────────────────────────────────────────────────┐
│ Draft（草稿）                                             │
│   workflow_state: Draft                                   │
│   docstatus: 0                                            │
│   status: Draft                                           │
└───────────────┬───────────────────────────────────────────┘
                │ [提交]
                ▼
┌─────────────────────────────────────────────────────────┐
│ Submitted（已提交）                                       │
│   workflow_state: Submitted                               │
│   docstatus: 0  ⚠️ 关键：保持为 0，不提交                │
│   status: Draft                                           │
└───────────────┬───────────────────────────────────────────┘
                │ [触发工作流]
                ▼
┌─────────────────────────────────────────────────────────┐
│ Pending Approval（待审批）                                │
│   workflow_state: Pending Approval                        │
│   docstatus: 0  ⚠️ 关键：保持为 0，不能创建后续单据      │
│   status: Draft                                           │
│   ❌ 不能创建交货单                                       │
│   ❌ 不能创建发票                                         │
└───────────────┬───────────────────────────────────────────┘
                │ [审批通过 + 更新 docstatus = 1]
                ▼
┌─────────────────────────────────────────────────────────┐
│ Approved（已批准）                                        │
│   workflow_state: Approved                                │
│   docstatus: 1  ⚠️ 关键：更新为 1，可以创建后续单据      │
│   status: Submitted（等待系统自动更新）                    │
│   ✅ 可以创建交货单                                       │
│   ✅ 可以创建发票                                         │
└───────────────┬───────────────────────────────────────────┘
                │ [系统根据交付和开票情况自动更新 status]
                ▼
┌─────────────────────────────────────────────────────────┐
│ 订单业务状态（status）自动流转：                          │
│                                                           │
│  ┌──────────────────────────────────────────┐          │
│  │ To Deliver and Bill（待交付和开票）       │          │
│  │   条件：未交付且未开票                    │          │
│  └──────────────────────────────────────────┘          │
│              │                                           │
│              ├─→ [部分交付]                              │
│              │                                           │
│              ▼                                           │
│  ┌──────────────────────────────────────────┐          │
│  │ To Deliver（待交付）                      │          │
│  │   条件：已部分或全部开票，但未完全交付    │          │
│  └──────────────────────────────────────────┘          │
│              │                                           │
│              ├─→ [部分开票]                              │
│              │                                           │
│              ▼                                           │
│  ┌──────────────────────────────────────────┐          │
│  │ To Bill（待开票）                         │          │
│  │   条件：已部分或全部交付，但未完全开票    │          │
│  └──────────────────────────────────────────┘          │
│              │                                           │
│              ├─→ [全部交付且全部开票]                    │
│              │                                           │
│              ▼                                           │
│  ┌──────────────────────────────────────────┐          │
│  │ Completed（已完成）                      │          │
│  │   条件：全部交付且全部开票                │          │
│  └──────────────────────────────────────────┘          │
└─────────────────────────────────────────────────────────┘
```

**关键说明**：
- **docstatus（文档状态）**：⭐ **最关键**，控制是否可以创建后续单据
  - `docstatus = 0`：草稿状态，**不能**创建交货单、发票等后续单据
  - `docstatus = 1`：已提交状态，**可以**创建后续单据
- **工作流状态（workflow_state）**：用于审批流程控制，包括 Draft、Submitted、Pending Approval、Approved、Rejected
- **订单状态（status）**：用于业务流转，包括 Draft、Submitted、To Deliver and Bill、To Deliver、To Bill、Completed
- **审批流程的核心**：工作流控制 `docstatus` 的更新时机
  - 待审批时：`docstatus = 0`（不能创建后续单据）
  - 审批通过时：工作流动作将 `docstatus` 更新为 `1`（可以创建后续单据）

---

## 二、ERPNEXT 工作流配置步骤（详细版）

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
| **工作流名称** | `Sales Order Approval Workflow` | 工作流的显示名称 |
| **文档类型** | `Sales Order` | 选择"销售订单" |
| **工作流状态字段** | `workflow_state` ⭐ **关键** | 存储工作流状态的字段名 |
| **工作流状态** | `Active` | 必须选择"启用"才能生效 |
| **是否系统工作流** | `否` | 用户自定义工作流 |
| **发送电子邮件提醒** | `是`（推荐） | 审批时发送邮件通知 |

**⚠️ 关键配置：工作流状态字段**

- **字段名**：`workflow_state`（默认值）
- **说明**：这是存储工作流状态的字段名，必须与销售订单中的字段名一致
- **检查方法**：
  1. 打开一个销售订单
  2. 查看是否有 `workflow_state` 字段
  3. 如果没有，需要在销售订单的 Customize Form 中添加此字段

#### 步骤 3：保存工作流

1. 点击右上角"保存"按钮
2. 系统会提示"工作流已保存"
3. **重要**：此时工作流还未生效，需要完成后续配置

### 2.2 定义工作流状态（Workflow States）

#### 步骤 1：进入状态配置页面

在工作流编辑页面，点击"工作流状态"（Workflow States）标签页。

#### 步骤 2：添加工作流状态

点击"添加行"（Add Row）按钮，依次添加以下状态：

| 序号 | 状态名称 | 状态标识 | 颜色 | 说明 | 是否允许编辑 |
|------|---------|---------|------|------|------------|
| 1 | Draft | `Draft` | 灰色（Gray） | 草稿状态，订单可编辑 | ✅ 是 |
| 2 | Submitted | `Submitted` | 蓝色（Blue） | 已提交，等待审批 | ❌ 否 |
| 3 | Pending Approval | `Pending Approval` | 橙色（Orange） | 待审批，等待审批人员处理 | ❌ 否 |
| 4 | Approved | `Approved` | 绿色（Green） | **已批准，可以继续后续流程** | ❌ 否 |
| 5 | Rejected | `Rejected` | 红色（Red） | 已拒绝，需要修改后重新提交 | ✅ 是 |

**详细配置说明**：

**状态 1：Draft（草稿）**
- **状态标识**：`Draft`（必须与 ERPNEXT 默认状态一致）
- **颜色**：灰色
- **允许编辑**：✅ 是（用户可以修改订单）
- **说明**：订单创建后的初始状态

**状态 2：Submitted（已提交）**
- **状态标识**：`Submitted`
- **颜色**：蓝色
- **允许编辑**：❌ 否（提交后不可编辑）
- **说明**：订单已提交，但未触发审批流程

**状态 3：Pending Approval（待审批）**
- **状态标识**：`Pending Approval`（自定义状态）
- **颜色**：橙色（用于提醒）
- **允许编辑**：✅ 是（可以修改订单）
- **文档状态（Docstatus）**：`0`（草稿）⭐ **关键配置**
- **说明**：订单已进入审批流程，等待审批人员处理
- **重要**：此状态下 `docstatus = 0`，**不能**创建后续单据

**状态 4：Approved（已批准）⭐ 关键状态**
- **状态标识**：`Approved`（自定义状态）
- **颜色**：绿色（表示通过）
- **允许编辑**：❌ 否（已批准后不可编辑）
- **文档状态（Docstatus）**：`1`（已提交）⭐ **关键配置**
- **说明**：**订单已批准，可以继续后续流程**（创建交货单、发票等）
- **重要**：此状态下 `docstatus = 1`，**可以**创建后续单据

**状态 5：Rejected（已拒绝）**
- **状态标识**：`Rejected`（自定义状态）
- **颜色**：红色（表示拒绝）
- **允许编辑**：✅ 是（可以修改后重新提交）
- **说明**：订单被拒绝，需要修改后重新提交审批

#### 步骤 3：保存状态配置

1. 添加完所有状态后，点击"保存"按钮
2. 系统会验证状态配置是否正确
3. 如有错误，会提示修改

### 2.3 配置工作流转换（Workflow Transitions）

#### 步骤 1：进入转换配置页面

在工作流编辑页面，点击"工作流转换"（Workflow Transitions）标签页。

#### 步骤 2：添加状态转换规则

点击"添加行"（Add Row）按钮，依次添加以下转换：

| 序号 | 转换名称 | 当前状态 | 下一状态 | 条件 | 允许的角色 | 说明 |
|------|---------|---------|---------|------|-----------|------|
| 1 | Submit | Draft | Submitted | - | Sales User, Sales Manager | 提交订单 |
| 2 | Submit for Approval | Submitted | Pending Approval | `docstatus == 1` | Sales User, Sales Manager | 提交审批 |
| 3 | Approve ⭐ | Pending Approval | Approved | - | Sales Manager, Accounts Manager | **审批通过，订单可继续后续流程** |
| 4 | Reject | Pending Approval | Rejected | - | Sales Manager, Accounts Manager | 审批拒绝 |
| 5 | Re-submit | Rejected | Pending Approval | - | Sales User | 重新提交审批 |

#### 步骤 3：详细配置每个转换

**转换 1：Submit（提交订单）**

```
动作（Action）：Submit
当前状态（Current State）：Draft
下一状态（Next State）：Submitted
条件（Condition）：留空（无条件）
允许的角色（Allowed Roles）：
  - Sales User
  - Sales Manager
动作（Action）：留空（使用默认动作）
```

**转换 2：Submit for Approval（提交审批）⭐ 关键转换**

```
动作（Action）：Submit for Approval
当前状态（Current State）：Submitted
下一状态（Next State）：Pending Approval
条件（Condition）：docstatus == 1
  （说明：只有当订单 docstatus 为 1 时才触发）
允许的角色（Allowed Roles）：
  - Sales User
  - Sales Manager
动作（Action）：留空（使用默认动作）
```

**转换 3：Approve（审批通过）⭐ 最关键转换**

```
动作（Action）：Approve
当前状态（Current State）：Pending Approval
下一状态（Next State）：Approved
条件（Condition）：留空（无条件）
允许的角色（Allowed Roles）：
  - Sales Manager
  - Accounts Manager
动作（Action）：留空（使用默认动作）

重要说明：
- 此转换将 workflow_state 更新为 "Approved"
- 审批通过后，订单可以继续后续流程：
  ✅ 可以创建交货单（Delivery Note）
  ✅ 可以创建发票（Sales Invoice）
  ✅ 可以创建其他相关单据
- 订单的 status 字段会根据交付和开票情况自动更新
```

**转换 4：Reject（审批拒绝）**

```
动作（Action）：Reject
当前状态（Current State）：Pending Approval
下一状态（Next State）：Rejected
条件（Condition）：留空（无条件）
允许的角色（Allowed Roles）：
  - Sales Manager
  - Accounts Manager
动作（Action）：留空（使用默认动作）
```

**转换 5：Re-submit（重新提交审批）**

```
动作（Action）：Re-submit
当前状态（Current State）：Rejected
下一状态（Next State）：Pending Approval
条件（Condition）：留空（无条件）
允许的角色（Allowed Roles）：
  - Sales User
动作（Action）：留空（使用默认动作）
```

#### 步骤 4：配置转换条件（可选）

对于"Submit for Approval"转换，可以添加条件：

**条件配置示例**：

```
条件类型：Python 表达式
条件表达式：docstatus == 1

说明：
- 只有当订单的 docstatus 为 1（已提交）时才允许提交审批
- 可以添加其他条件，如：doc.grand_total > 10000（只有金额大于 10000 的订单才需要审批）
```

#### 步骤 5：保存转换配置

1. 添加完所有转换后，点击"保存"按钮
2. 系统会验证转换配置是否正确
3. 确保每个转换的"当前状态"和"下一状态"都在工作流状态中已定义

### 2.4 配置工作流动作（Workflow Actions）

工作流动作定义了状态转换时执行的具体操作。ERPNEXT 会自动处理大部分动作，但我们可以配置额外的动作。

#### 步骤 1：进入动作配置页面

在工作流编辑页面，点击"工作流动作"（Workflow Actions）标签页。

#### 步骤 2：配置动作（可选）

**默认情况下**，ERPNEXT 会自动处理状态转换，但我们可以添加额外的动作：

**动作 1：提交审批时的通知**

```
动作名称：Submit for Approval
触发时机：Submitted → Pending Approval
动作类型：Email Notification（邮件通知）
收件人：审批人员（从角色中获取）
主题：销售订单待审批：{{ doc.name }}
内容：订单 {{ doc.name }} 已提交审批，请及时处理。
```

**动作 2：审批通过 ⭐ 最关键动作**

```
动作名称：Approve
触发时机：Pending Approval → Approved

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
主题：销售订单已批准：{{ doc.name }}
内容：订单 {{ doc.name }} 已通过审批，可以继续后续流程。

⚠️ 关键说明：
✅ 审批通过后，必须同时更新两个字段：
   1. workflow_state = "Approved"（工作流状态）
   2. docstatus = 1（文档状态）⭐ 最关键
✅ 只有 docstatus = 1 时，才能创建后续单据：
   - 创建交货单（Delivery Note）
   - 创建发票（Sales Invoice）
   - 创建其他相关单据
✅ 订单的 status 字段会根据交付和开票情况自动更新：
   * 未交付且未开票 → To Deliver and Bill
   * 已部分或全部开票，但未完全交付 → To Deliver
   * 已部分或全部交付，但未完全开票 → To Bill
   * 全部交付且全部开票 → Completed
```

**动作 3：审批拒绝时的通知**

```
动作名称：Reject
触发时机：Pending Approval → Rejected
动作类型：Email Notification（邮件通知）
收件人：订单创建人（doc.owner）
主题：销售订单已拒绝：{{ doc.name }}
内容：订单 {{ doc.name }} 已被拒绝，请修改后重新提交。
```

**动作 4：重新提交时的通知**

```
动作名称：Re-submit
触发时机：Rejected → Pending Approval
动作类型：Email Notification（邮件通知）
收件人：审批人员（从角色中获取）
主题：销售订单重新提交审批：{{ doc.name }}
内容：订单 {{ doc.name }} 已重新提交审批，请及时处理。
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

**角色 1：Sales User（销售用户）**

```
角色名称：Sales User
文档类型：Sales Order

权限配置：
✅ 读取（Read）：是
✅ 创建（Create）：是
✅ 写入（Write）：是（仅限 Draft 状态）
✅ 提交（Submit）：是
✅ 取消（Cancel）：是（仅限 Draft 状态）

工作流动作权限：
✅ Submit for Approval：是（可以提交审批）
✅ Re-submit：是（可以重新提交被拒绝的订单）
❌ Approve：否（不能审批）
❌ Reject：否（不能拒绝）
```

**角色 2：Sales Manager（销售经理）⭐ 审批角色**

```
角色名称：Sales Manager
文档类型：Sales Order

权限配置：
✅ 读取（Read）：是
✅ 创建（Create）：是
✅ 写入（Write）：是（仅限 Draft 和 Rejected 状态）
✅ 提交（Submit）：是
✅ 取消（Cancel）：是（仅限 Draft 状态）

工作流动作权限：
✅ Submit for Approval：是（可以提交审批）
✅ Approve：是（可以审批通过）⭐ 关键权限
✅ Reject：是（可以审批拒绝）⭐ 关键权限
✅ Re-submit：是（可以重新提交）

后续流程权限：
✅ 创建交货单（Delivery Note）：是（审批通过后）
✅ 创建发票（Sales Invoice）：是（审批通过后）
✅ 查看订单状态：是
```

**角色 3：Accounts Manager（财务经理）⭐ 审批角色**

```
角色名称：Accounts Manager
文档类型：Sales Order

权限配置：
✅ 读取（Read）：是
✅ 创建（Create）：否（通常不创建订单）
✅ 写入（Write）：否（不能修改订单）
✅ 提交（Submit）：否
✅ 取消（Cancel）：否

工作流动作权限：
❌ Submit for Approval：否（不能提交审批）
✅ Approve：是（可以审批通过）⭐ 关键权限
✅ Reject：是（可以审批拒绝）⭐ 关键权限
❌ Re-submit：否（不能重新提交）

后续流程权限：
✅ 创建发票（Sales Invoice）：是（审批通过后）
✅ 查看订单状态：是
```

#### 步骤 3：配置工作流状态权限

**重要**：需要确保不同状态下的权限设置正确：

**状态：Pending Approval（待审批）**
- Sales User：❌ 不能编辑
- Sales Manager：✅ 可以审批（Approve/Reject）
- Accounts Manager：✅ 可以审批（Approve/Reject）

**状态：Approved（已批准）⭐ 关键状态**
- Sales User：✅ 可以创建交货单、发票等后续单据
- Sales Manager：✅ 可以创建交货单、发票等后续单据
- Accounts Manager：✅ 可以创建发票等后续单据
- **所有角色**：❌ 不能编辑订单本身（已批准后不可修改）

#### 步骤 4：验证权限配置

1. 使用不同角色的用户登录系统
2. 测试是否可以执行相应的工作流动作
3. 确保审批通过后可以继续后续流程

### 2.6 确保审批后可以继续后续流程 ⭐ 关键配置

这是最重要的配置步骤，确保审批通过后订单可以继续后续流程。

#### 关键点说明

**审批通过后的状态变化**：

```
审批通过前：
  workflow_state: Pending Approval
  docstatus: 0  ⚠️ 关键：保持为 0（草稿）
  status: Draft
  ❌ 不能创建交货单（因为 docstatus = 0）
  ❌ 不能创建发票（因为 docstatus = 0）

审批通过后：
  workflow_state: Approved  ⭐ 工作流状态变化
  docstatus: 1  ⚠️ 关键：更新为 1（已提交）
  status: Submitted（等待自动更新）
  ✅ 可以创建交货单（因为 docstatus = 1）
  ✅ 可以创建发票（因为 docstatus = 1）
  ✅ 可以继续后续流程
```

**关键机制**：
- ERPNEXT 检查的是 `docstatus`，而不是 `workflow_state`
- 只有 `docstatus = 1` 的订单才能创建后续单据
- 工作流的作用是：控制何时将 `docstatus` 从 0 更新为 1

#### 配置要点

**1. 工作流状态必须设置为 "Approved"**

在工作流转换中，"Approve" 动作必须将 `workflow_state` 更新为 `Approved`：

```
转换：Approve
当前状态：Pending Approval
下一状态：Approved  ⭐ 必须是 Approved
```

**2. ⚠️ 最关键：必须更新 docstatus 为 1**

在工作流动作中，"Approve" 动作**必须**同时更新 `docstatus` 为 `1`：

```
动作：Approve
动作类型：Update Field
字段：docstatus
值：1
说明：将文档状态更新为"已提交"，这是允许创建后续单据的关键
```

**3. 权限配置必须允许后续操作**

确保以下角色有创建后续单据的权限：

```
Sales User / Sales Manager：
  ✅ 创建交货单（Delivery Note）的权限
  ✅ 创建发票（Sales Invoice）的权限
  ✅ 查看订单详情的权限

Accounts Manager：
  ✅ 创建发票（Sales Invoice）的权限
  ✅ 查看订单详情的权限
```

**4. ERPNEXT 系统检查机制**

- ERPNEXT **检查的是 `docstatus`，而不是 `workflow_state`**
- 只有 `docstatus = 1` 的订单才能创建后续单据
- 如果 `docstatus = 0`，即使 `workflow_state = Approved`，也无法创建后续单据

#### 验证配置

**测试步骤**：

1. **创建测试订单**
   - 创建一个销售订单
   - 状态：Draft

2. **提交订单**
   - 点击"提交"按钮
   - 验证：`docstatus = 1`，`workflow_state = Submitted`

3. **提交审批**
   - 点击"提交审批"按钮（如果配置了自动触发，此步骤可能自动完成）
   - 验证：`workflow_state = Pending Approval`

4. **审批通过**
   - 使用审批人员账号登录
   - 点击"批准"按钮
   - 验证：`workflow_state = Approved` ⭐ 关键验证点

5. **验证后续流程**
   - 在订单页面，点击"创建" → "交货单"
   - ✅ 应该可以创建交货单
   - 在订单页面，点击"创建" → "发票"
   - ✅ 应该可以创建发票

**如果无法创建后续单据**：

1. 检查 `workflow_state` 是否为 `Approved`
2. 检查用户角色权限是否正确配置
3. 检查工作流是否已启用（Active）
4. 检查订单的 `docstatus` 是否为 1

### 2.7 配置条件（可选）

如果需要根据订单金额或其他条件触发审批，可以添加条件。

#### 示例 1：金额大于 10000 的订单需要审批

**配置步骤**：

1. 在工作流转换中，找到"Submit for Approval"转换
2. 点击"条件"（Condition）字段
3. 输入 Python 表达式：
   ```python
   doc.grand_total > 10000
   ```
4. 保存配置

**效果**：
- 金额 ≤ 10000 的订单：提交后直接进入 `Submitted` 状态，不触发审批
- 金额 > 10000 的订单：提交后进入 `Pending Approval` 状态，需要审批

#### 示例 2：特定客户的订单需要审批

**配置步骤**：

1. 在工作流转换中，找到"Submit for Approval"转换
2. 点击"条件"字段
3. 输入 Python 表达式：
   ```python
   doc.customer in ["Customer A", "Customer B"]
   ```
4. 保存配置

#### 示例 3：组合条件

**配置步骤**：

1. 在工作流转换中，找到"Submit for Approval"转换
2. 点击"条件"字段
3. 输入 Python 表达式：
   ```python
   doc.grand_total > 10000 or doc.customer == "VIP Customer"
   ```
4. 保存配置

**常用条件表达式**：

| 条件 | Python 表达式 |
|------|--------------|
| 金额大于 10000 | `doc.grand_total > 10000` |
| 金额大于等于 10000 | `doc.grand_total >= 10000` |
| 特定客户 | `doc.customer == "Customer Name"` |
| 多个客户 | `doc.customer in ["Customer A", "Customer B"]` |
| 组合条件（AND） | `doc.grand_total > 10000 and doc.customer == "VIP"` |
| 组合条件（OR） | `doc.grand_total > 10000 or doc.customer == "VIP"` |

---

## 三、代码集成方案

### 3.1 查询订单工作流状态

在代码中查询销售订单时，可以获取工作流状态：

```go
// 文件：ttpos-bmp/app/ttpos-erp/internal/logic/selling/sale_order.go

// GetSalesOrder 获取销售订单信息
func (s *sSelling) GetSalesOrder(ctx context.Context, name string) (*dtoSelling.SalesOrder, error) {
    // ... 现有代码 ...
    
    // 解析响应数据
    salesOrder, err := s.parseSalesOrderResponse(resp)
    if err != nil {
        return nil, err
    }
    
    // 获取工作流状态（如果存在）
    workflowState := resp.Get("data.workflow_state").String()
    if workflowState != "" {
        // 可以在这里处理工作流状态
        // 例如：Pending Approval、Approved、Rejected
    }
    
    return salesOrder, nil
}
```

### 3.2 执行工作流动作

如果需要通过代码触发工作流动作，可以使用 ERPNEXT 的 RPC API：

```go
// 文件：ttpos-bmp/app/ttpos-erp/internal/logic/selling/sale_order.go

// ApproveSalesOrder 审批通过销售订单
func (s *sSelling) ApproveSalesOrder(ctx context.Context, name string) error {
    if name == "" {
        return gerror.New("销售订单名称不能为空")
    }
    
    // 调用 ERPNEXT RPC API 执行工作流动作
    resp, err := service.Rpc().Execute(ctx, &erp.ErpReq{
        DocType: erp.DocTypeSaleOrder,
        Method:  "frappe.client.submit",
    }, map[string]interface{}{
        "doctype": "Sales Order",
        "name":    name,
        "action":  "Approve", // 工作流动作名称
    })
    
    if err != nil {
        return gerror.Wrapf(err, "审批销售订单失败")
    }
    
    return nil
}

// RejectSalesOrder 审批拒绝销售订单
func (s *sSelling) RejectSalesOrder(ctx context.Context, name string, reason string) error {
    if name == "" {
        return gerror.New("销售订单名称不能为空")
    }
    
    // 调用 ERPNEXT RPC API 执行工作流动作
    resp, err := service.Rpc().Execute(ctx, &erp.ErpReq{
        DocType: erp.DocTypeSaleOrder,
        Method:  "frappe.client.submit",
    }, map[string]interface{}{
        "doctype": "Sales Order",
        "name":    name,
        "action":  "Reject", // 工作流动作名称
        "reason":  reason,  // 拒绝原因
    })
    
    if err != nil {
        return gerror.Wrapf(err, "拒绝销售订单失败")
    }
    
    return nil
}
```

### 3.3 查询待审批订单列表

```go
// GetPendingApprovalSalesOrders 获取待审批的销售订单列表
func (s *sSelling) GetPendingApprovalSalesOrders(ctx context.Context, req *dtoSelling.SalesOrderListReq) ([]*dtoSelling.SalesOrder, error) {
    // 构建查询过滤器，添加工作流状态过滤
    filters := s.buildSalesOrderListFilters(ctx, req)
    
    // 添加工作流状态过滤
    filters = append(filters, []string{"workflow_state", "=", "Pending Approval"})
    
    // 查询销售订单列表
    resp, err := service.Document().List(ctx, &erp.ErpReq{
        DocType: erp.DocTypeSaleOrder,
    }, &erp.RequestParams{
        Fields: g.ArrayStr{
            "name", "customer", "customer_name", "transaction_date",
            "delivery_date", "grand_total", "status", "workflow_state",
            "delivery_status", "billing_status",
        },
        Filters:    filters,
        Limit:      req.Limit,
        LimitStart: req.LimitStart,
        OrderBy: erp.OrderBy{
            Field: "modified",
            Order: "desc",
        },
    })
    
    if err != nil {
        return nil, gerror.Wrapf(err, "查询待审批销售订单列表失败")
    }
    
    // 解析响应数据
    return s.parseSalesOrderListResponse(resp)
}
```

---

## 四、工作流状态与订单状态的关系

### 4.1 三种状态字段的区别

ERPNEXT 销售订单有三个独立的状态字段，各自有不同的用途：

| 字段名 | 类型 | 用途 | 控制方式 |
|--------|------|------|---------|
| `docstatus` | 整数（0/1/2） | 文档基础状态 | 系统/用户操作 |
| `workflow_state` | 字符串 | 工作流审批状态 | 工作流系统 |
| `status` | 字符串 | 订单业务状态 | 系统自动计算 |

### 4.2 状态映射表

| 工作流状态 | docstatus | workflow_state | status | 说明 |
|-----------|-----------|----------------|--------|------|
| Draft | 0 | Draft | Draft | 草稿状态 |
| Submitted | 0 | Submitted | Draft | 已提交，但 docstatus = 0，未真正提交 |
| Pending Approval | 0 | Pending Approval | Draft | 待审批，docstatus = 0，**不能**创建后续单据 |
| Approved | 1 | Approved | Submitted → To Deliver and Bill → ... | 已批准，docstatus = 1，**可以**创建后续单据 |
| Rejected | 0 | Rejected | Draft | 已拒绝，docstatus = 0，**不能**创建后续单据 |

### 4.3 审批通过后的状态自动流转

**重要**：
1. 审批通过时，工作流动作必须将 `docstatus` 更新为 `1`（已提交）
2. 只有 `docstatus = 1` 的订单才能创建后续单据
3. 审批通过后，订单的 `status` 字段会根据**交付和开票情况**自动更新，无需在工作流中配置

#### 状态流转逻辑

```
审批通过后（workflow_state = Approved，docstatus = 1）：
  ↓
系统检查交付和开票情况
  ↓
根据以下规则自动更新 status：

1. 未交付且未开票
   → status = "To Deliver and Bill"

2. 已部分或全部开票，但未完全交付
   → status = "To Deliver"

3. 已部分或全部交付，但未完全开票
   → status = "To Bill"

4. 全部交付且全部开票
   → status = "Completed"
```

#### 状态更新触发时机

`status` 字段的更新由以下操作触发：

1. **创建交货单（Delivery Note）**
   - 部分交付 → `status` 可能变为 `To Deliver` 或 `To Bill`
   - 全部交付 → `status` 可能变为 `To Bill` 或 `Completed`

2. **创建发票（Sales Invoice）**
   - 部分开票 → `status` 可能变为 `To Deliver` 或 `To Bill`
   - 全部开票 → `status` 可能变为 `To Deliver` 或 `Completed`

3. **同时完成交付和开票**
   - `status` 自动变为 `Completed`

### 4.4 工作流状态与订单状态的关系图

```
┌─────────────────────────────────────────────────────────────┐
│ 工作流状态（workflow_state）                                  │
│                                                               │
│  Draft → Submitted → Pending Approval → Approved            │
│                                    ↓                          │
│                                 Rejected                      │
└─────────────────────────────────────────────────────────────┘
                            │
                            │ 审批通过后
                            ▼
┌─────────────────────────────────────────────────────────────┐
│ 订单业务状态（status）自动流转                                │
│                                                               │
│  Submitted → To Deliver and Bill → To Deliver → Completed   │
│              │                                                │
│              └→ To Bill → Completed                           │
│                                                               │
│ 说明：                                                        │
│ - status 的更新由交付和开票操作触发                          │
│ - 不需要在工作流中配置 status 的更新                         │
│ - 工作流只控制审批流程，不控制业务流转                       │
└─────────────────────────────────────────────────────────────┘
```

### 4.5 关键要点

1. **工作流状态（workflow_state）**：
   - 用于控制审批流程
   - 由工作流系统管理
   - 审批通过后变为 `Approved`，之后不再变化

2. **订单状态（status）**：
   - 用于表示订单的业务进度
   - 由 ERPNEXT 系统根据交付和开票情况**自动计算**
   - 审批通过后，会根据实际业务操作自动更新

3. **两者关系**：
   - 工作流状态控制**是否可以继续后续流程**
   - 订单状态反映**当前业务进度**
   - 审批通过（`workflow_state = Approved`）是继续后续流程的**前提条件**
   - 但订单状态的更新**不依赖工作流**，而是由业务操作触发

### 4.6 状态查询示例

在查询订单时，需要同时关注三个状态字段：

```go
type SalesOrder struct {
    // ... 其他字段 ...
    Docstatus     int    `json:"docstatus,omitempty"`      // 文档状态：0-草稿, 1-已提交, 2-已取消
    Status        string `json:"status,omitempty"`         // 订单状态：Draft、Submitted、To Deliver 等
    WorkflowState string `json:"workflow_state,omitempty"` // 工作流状态：Pending Approval、Approved、Rejected
}
```

---

## 五、Webhook 集成（可选）

如果需要在工作流状态变更时通知 TTPOS 系统，可以配置 Webhook：

### 5.1 ERPNEXT Webhook 配置

1. **创建 Webhook**
   ```
   导航至：设置 → Webhook → 新建
   
   Webhook 名称：Sales Order Workflow Notification
   文档类型：Sales Order
   事件：After Save（保存后）
   请求 URL：https://your-ttpos-server.com/api/webhook/erpnext/sales-order
   请求方法：POST
   启用：是
   ```

2. **Webhook 请求体示例**
   ```json
   {
     "event": "after_save",
     "doctype": "Sales Order",
     "name": "SO-00001",
     "doc": {
       "name": "SO-00001",
       "customer": "Customer A",
       "grand_total": 10000,
       "docstatus": 1,
       "status": "Submitted",
       "workflow_state": "Pending Approval"
     }
   }
   ```

### 5.2 TTPOS 接收 Webhook

```go
// 文件：main/app/api/webhook/erpnext.go

// HandleSalesOrderWebhook 处理 ERPNEXT 销售订单 Webhook
func HandleSalesOrderWebhook(c *gin.Context) {
    var req struct {
        Event    string                 `json:"event"`
        Doctype  string                 `json:"doctype"`
        Name     string                 `json:"name"`
        Doc      map[string]interface{} `json:"doc"`
    }
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": "Invalid request"})
        return
    }
    
    // 获取工作流状态
    workflowState := req.Doc["workflow_state"].(string)
    
    // 根据工作流状态处理业务逻辑
    switch workflowState {
    case "Pending Approval":
        // 通知审批人员
        notifyApprovers(req.Name)
    case "Approved":
        // 订单已批准，继续后续流程
        handleApprovedOrder(req.Name)
    case "Rejected":
        // 订单被拒绝，通知创建人
        notifyCreator(req.Name)
    }
    
    c.JSON(200, gin.H{"status": "ok"})
}
```

---

## 六、高级配置：在工作流中控制订单状态（可选）

### 6.1 默认行为

**默认情况下**，审批通过后订单的 `status` 字段会根据交付和开票情况自动更新，**不需要在工作流中配置**。

### 6.2 如果需要强制控制订单状态

如果业务需求要求在审批通过时强制设置订单状态，可以在工作流动作中添加字段更新：

**示例：审批通过时强制设置为 "To Deliver and Bill"**

```
动作名称：Approve
触发时机：Pending Approval → Approved
动作类型：Update Field（多个）
字段1：workflow_state
值1：Approved

字段2：status（可选）
值2：To Deliver and Bill

说明：
- 字段2（status）的更新是可选的
- 如果不设置，系统会根据交付和开票情况自动更新
- 如果设置，会覆盖系统的自动更新逻辑
```

**注意事项**：
- ⚠️ **不推荐**在工作流中强制设置 `status` 字段
- 因为 `status` 应该反映实际的业务进度（交付和开票情况）
- 强制设置可能导致状态与实际业务不符
- 建议让系统根据业务操作自动更新 `status`

### 6.3 工作流状态与订单状态的配合使用

**推荐做法**：

1. **工作流控制审批流程**
   - 使用 `workflow_state` 控制审批状态
   - 审批通过后，`workflow_state = Approved`

2. **系统自动更新业务状态**
   - 让 ERPNEXT 系统根据交付和开票情况自动更新 `status`
   - 不在工作流中强制设置 `status`

3. **查询时组合使用**
   ```go
   // 查询已批准且待交付的订单
   filters := [][]string{
       {"workflow_state", "=", "Approved"},
       {"status", "=", "To Deliver and Bill"},
   }
   
   // 查询已批准且已完成的订单
   filters := [][]string{
       {"workflow_state", "=", "Approved"},
       {"status", "=", "Completed"},
   }
   ```

---

## 七、故障排查：工作流状态一直停留在 Draft

### 7.1 问题现象

配置工作流后，创建销售订单时，`workflow_state` 字段一直显示为 `Draft`，没有按照配置的状态流转。

### 7.2 可能原因及解决方案

#### 原因 1：工作流状态字段未添加到表单 ⭐ 最常见

**症状**：
- 订单中没有 `workflow_state` 字段
- 或者字段名不匹配

**排查步骤**：
1. 打开一个销售订单
2. 查看表单中是否有 `workflow_state` 字段
3. 如果没有，说明字段未添加到表单中

**解决方案**：
1. **添加工作流状态字段到销售订单表单**：
   ```
   导航至：设置 → 自定义 → 表单 → Sales Order
   ```
2. **添加字段**：
   - 字段名：`workflow_state`
   - 字段类型：`Data` 或 `Select`
   - 标签：`工作流状态`
   - 保存表单
3. **验证**：
   - 重新打开销售订单，应该能看到 `workflow_state` 字段

#### 原因 2：Draft 状态未设置为初始状态

**症状**：
- 工作流已配置，但订单创建后 `workflow_state` 为空或不是 `Draft`

**排查步骤**：
1. 打开工作流配置页面
2. 找到 "Draft" 状态
3. 检查是否勾选了"是否为初始状态"（Is Initial State）

**解决方案**：
1. 在工作流状态配置中，找到 "Draft" 状态
2. ✅ 勾选"是否为初始状态"（Is Initial State）
3. 保存工作流

#### 原因 3：工作流未启用

**症状**：
- 工作流已创建，但状态不是 `Active`

**排查步骤**：
1. 打开工作流列表
2. 找到配置的工作流
3. 检查"工作流状态"字段是否为 `Active`

**解决方案**：
1. 打开工作流配置页面
2. 将"工作流状态"设置为 `Active`
3. 保存工作流

#### 原因 4：工作流转换条件不满足

**症状**：
- 订单创建后，`workflow_state` 是 `Draft`
- 但提交订单后，状态没有变化

**排查步骤**：
1. 检查工作流转换配置
2. 检查"Submit" 或 "Submit for Approval" 转换的条件
3. 检查转换的"当前状态"是否为 `Draft`

**解决方案**：
1. 确保有一个从 `Draft` 到其他状态的转换
2. 转换条件应该为空或满足订单条件
3. 确保用户角色有执行转换的权限

#### 原因 5：直接提交订单，未触发工作流

**症状**：
- 用户点击"提交"按钮，订单直接提交（`docstatus = 1`）
- 工作流状态没有变化

**排查步骤**：
1. 检查订单的 `docstatus` 字段
2. 如果 `docstatus = 1`，说明订单已直接提交
3. 检查工作流是否配置了阻止直接提交

**解决方案**：
1. **方案 A：配置工作流阻止直接提交**
   - 在工作流状态配置中，将 "Draft" 状态的 `docstatus` 设置为 `0`
   - 确保只有通过工作流转换才能更新 `docstatus`

2. **方案 B：修改提交按钮行为**
   - 在销售订单的自定义脚本中，拦截提交操作
   - 改为触发工作流转换，而不是直接提交

**推荐配置**：
```
工作流状态配置：
- Draft：docstatus = 0，Is Initial State = Yes
- Submitted：docstatus = 0（不提交）
- Pending Approval：docstatus = 0（不提交）
- Approved：docstatus = 1（提交）

工作流转换：
- Draft → Submitted：条件为空，允许角色：Sales User
- Submitted → Pending Approval：条件为空，允许角色：Sales User
- Pending Approval → Approved：条件为空，允许角色：Sales Manager
```

#### 原因 6：工作流状态字段名不匹配

**症状**：
- 工作流配置中使用的字段名与销售订单中的字段名不一致

**排查步骤**：
1. 检查工作流配置中的"工作流状态字段"设置
2. 检查销售订单中实际使用的字段名
3. 对比两者是否一致

**解决方案**：
1. 确保工作流配置中的"工作流状态字段"与销售订单中的字段名一致
2. 默认应该是 `workflow_state`
3. 如果使用了自定义字段名，确保两边一致

### 7.3 完整检查清单

如果工作流状态一直停留在 Draft，请按以下清单检查：

- [ ] **工作流状态字段已添加到销售订单表单**
  - 打开销售订单，检查是否有 `workflow_state` 字段
  - 如果没有，需要在 Customize Form 中添加

- [ ] **Draft 状态已设置为初始状态**
  - 在工作流状态配置中，Draft 状态的"Is Initial State"已勾选

- [ ] **工作流已启用**
  - 工作流状态为 `Active`

- [ ] **工作流转换已配置**
  - 有从 `Draft` 到其他状态的转换
  - 转换条件满足或为空

- [ ] **用户角色权限正确**
  - 用户有执行工作流转换的权限

- [ ] **工作流状态字段名匹配**
  - 工作流配置中的字段名与销售订单中的字段名一致

- [ ] **订单创建方式正确**
  - 创建订单后，检查 `workflow_state` 是否为 `Draft`
  - 如果不是，检查初始状态配置

### 7.4 快速诊断步骤

**步骤 1：检查工作流状态字段**

```
1. 打开一个销售订单
2. 查看表单中是否有 workflow_state 字段
3. 如果没有，添加字段：
   - 设置 → 自定义 → 表单 → Sales Order
   - 添加字段：workflow_state（Data 类型）
```

**步骤 2：检查初始状态配置**

```
1. 打开工作流配置页面
2. 找到 Draft 状态
3. 检查"Is Initial State"是否勾选
4. 如果没有，勾选并保存
```

**步骤 3：检查工作流是否启用**

```
1. 打开工作流列表
2. 找到配置的工作流
3. 检查状态是否为 Active
4. 如果不是，设置为 Active 并保存
```

**步骤 4：测试创建订单**

```
1. 创建一个新的销售订单
2. 保存订单（不提交）
3. 检查 workflow_state 字段
4. 应该显示为 Draft
5. 如果为空或不是 Draft，检查上述配置
```

---

## 八、故障排查：Don't Override Status 选项导致状态不更新

### 8.1 问题现象

配置工作流时勾选了 **"Don't Override Status"** 选项，执行工作流时，订单的 `status` 字段没有按照设定的状态呈现。

### 8.2 问题原因

**"Don't Override Status" 选项的作用**：

- ✅ **勾选此选项时**：ERPNext **不会自动更新**文档的 `status` 字段
- ✅ 即使工作流状态变化了，`status` 字段也不会被工作流动作覆盖
- ✅ 这意味着如果工作流动作中设置了 `status` 字段，这个设置会被忽略

**关键理解**：

1. **工作流状态（workflow_state）** 和 **订单状态（status）** 是两个独立的字段
2. `workflow_state` 由工作流系统控制，会正常更新
3. `status` 字段：
   - **未勾选 "Don't Override Status"**：工作流动作可以更新 `status` 字段
   - **勾选 "Don't Override Status"**：工作流动作**不能**更新 `status` 字段，只能由系统根据业务逻辑自动更新

### 8.3 解决方案

#### 方案一：取消勾选 "Don't Override Status"（推荐）

**适用场景**：希望工作流动作能够更新 `status` 字段

**操作步骤**：

1. 打开工作流配置页面
2. 找到工作流动作（Workflow Actions）
3. 找到需要更新 `status` 的动作（如 "Approve"）
4. **取消勾选 "Don't Override Status"** 选项
5. 在工作流动作中添加 `status` 字段更新：
   ```
   动作名称：Approve
   触发时机：Pending Approval → Approved
   动作类型：Update Field
   字段：status
   值：To Deliver and Bill（或您需要的状态值）
   ```
6. 保存工作流配置

**注意事项**：

- ⚠️ 取消勾选后，工作流动作可以更新 `status` 字段
- ⚠️ 但 `status` 字段应该反映实际的业务进度（交付和开票情况）
- ⚠️ 强制设置可能导致状态与实际业务不符
- ✅ **推荐做法**：让系统根据业务操作自动更新 `status`，而不是在工作流中强制设置

#### 方案二：保留 "Don't Override Status"，使用系统自动更新（推荐）

**适用场景**：希望 `status` 字段由系统根据业务逻辑自动更新

**操作步骤**：

1. **保持 "Don't Override Status" 选项勾选**
2. **不在工作流动作中设置 `status` 字段**
3. **让 ERPNext 系统根据业务操作自动更新 `status`**

**状态自动更新规则**：

```
审批通过后（workflow_state = Approved，docstatus = 1）：
  ↓
系统检查交付和开票情况
  ↓
根据以下规则自动更新 status：

1. 未交付且未开票
   → status = "To Deliver and Bill"

2. 已部分或全部开票，但未完全交付
   → status = "To Deliver"

3. 已部分或全部交付，但未完全开票
   → status = "To Bill"

4. 全部交付且全部开票
   → status = "Completed"
```

**优点**：

- ✅ `status` 字段始终反映实际的业务进度
- ✅ 不会出现状态与实际业务不符的情况
- ✅ 符合 ERPNext 的设计理念

#### 方案三：使用自定义脚本自动更新（高级）

**适用场景**：需要自定义 `status` 字段的更新逻辑

**操作步骤**：

1. **保持 "Don't Override Status" 选项勾选**
2. **创建服务器端脚本（Server Script）**：
   ```
   导航至：设置 → 自动化 → 服务器脚本 → 新建
   ```
3. **配置脚本**：
   ```python
   # 监听工作流状态变化
   def on_workflow_state_change(doc, method):
       # 当工作流状态变为 Approved 时
       if doc.workflow_state == "Approved":
           # 根据业务逻辑更新 status
           if not doc.delivery_status and not doc.billing_status:
               doc.status = "To Deliver and Bill"
           elif doc.delivery_status == "Not Delivered" and doc.billing_status == "Fully Billed":
               doc.status = "To Deliver"
           elif doc.delivery_status == "Fully Delivered" and doc.billing_status == "Not Billed":
               doc.status = "To Bill"
           elif doc.delivery_status == "Fully Delivered" and doc.billing_status == "Fully Billed":
               doc.status = "Completed"
   ```
4. **保存脚本**

**注意事项**：

- ⚠️ 需要熟悉 Frappe 框架和 Python 编程
- ⚠️ 脚本逻辑需要与 ERPNext 的业务逻辑保持一致
- ⚠️ 建议先测试，确保逻辑正确

### 8.4 推荐配置

**最佳实践**：

1. ✅ **勾选 "Don't Override Status" 选项**
2. ✅ **不在工作流动作中设置 `status` 字段**
3. ✅ **让系统根据业务操作自动更新 `status`**
4. ✅ **使用 `workflow_state` 字段控制审批流程**
5. ✅ **查询时组合使用 `workflow_state` 和 `status`**

**示例查询**：

```go
// 查询已批准且待交付的订单
filters := [][]string{
    {"workflow_state", "=", "Approved"},
    {"status", "=", "To Deliver and Bill"},
}

// 查询已批准且已完成的订单
filters := [][]string{
    {"workflow_state", "=", "Approved"},
    {"status", "=", "Completed"},
}
```

### 8.5 验证步骤

**验证工作流状态更新**：

1. 创建一个销售订单
2. 提交订单，进入审批流程
3. 检查 `workflow_state` 字段：
   - 应该从 `Draft` → `Submitted` → `Pending Approval`
4. 审批通过
5. 检查 `workflow_state` 字段：
   - 应该变为 `Approved`
6. 检查 `docstatus` 字段：
   - 应该变为 `1`（已提交）

**验证订单状态更新**：

1. 审批通过后，检查 `status` 字段：
   - 如果未交付且未开票，应该为 `To Deliver and Bill`
   - 如果已部分交付，应该为 `To Deliver` 或 `To Bill`
2. 创建交货单后，检查 `status` 字段：
   - 应该自动更新为 `To Bill` 或 `Completed`
3. 创建发票后，检查 `status` 字段：
   - 应该自动更新为 `To Deliver` 或 `Completed`

### 8.6 常见错误

**错误 1：勾选了 "Don't Override Status"，但在工作流动作中设置了 `status` 字段**

- ❌ **错误**：期望工作流动作更新 `status` 字段，但由于勾选了 "Don't Override Status"，更新被忽略
- ✅ **解决**：取消勾选 "Don't Override Status"，或者不在工作流动作中设置 `status` 字段

**错误 2：未勾选 "Don't Override Status"，但工作流动作中没有设置 `status` 字段**

- ❌ **错误**：期望 `status` 字段自动更新，但工作流动作中没有配置
- ✅ **解决**：在工作流动作中添加 `status` 字段更新，或者让系统自动更新

**错误 3：混淆 `workflow_state` 和 `status` 字段**

- ❌ **错误**：期望 `workflow_state` 的变化自动更新 `status` 字段
- ✅ **解决**：理解两个字段是独立的，`status` 字段需要单独配置或由系统自动更新

---

## 九、测试验证：审批后继续后续流程

### 9.1 完整测试流程

#### 步骤 1：创建测试订单

1. **登录 ERPNEXT 系统**
   - 使用 Sales User 角色账号登录

2. **创建销售订单**
   - 导航至：`销售` → `销售订单` → `新建`
   - 填写订单信息：
     - 客户：选择任意客户
     - 商品：添加商品
     - 金额：建议设置为 > 10000（如果配置了金额条件）
   - 点击"保存"按钮
   - **验证**：
     - `docstatus = 0`（草稿）
     - `workflow_state = Draft`
     - `status = Draft`

#### 步骤 2：提交订单

1. **提交订单**
   - 在订单页面，点击"提交"按钮
   - **验证**：
     - `docstatus = 1`（已提交）
     - `workflow_state = Submitted` 或 `Pending Approval`（取决于是否自动触发审批）
     - `status = Submitted`

2. **如果未自动触发审批**
   - 点击"提交审批"按钮（如果配置了此动作）
   - **验证**：
     - `workflow_state = Pending Approval`

#### 步骤 3：审批订单 ⭐ 关键步骤

1. **切换账号**
   - 退出当前账号
   - 使用 Sales Manager 或 Accounts Manager 角色账号登录

2. **查看待审批订单**
   - 导航至：`销售` → `销售订单`
   - 筛选条件：`workflow_state = Pending Approval`
   - 应该能看到待审批的订单

3. **审批通过**
   - 打开待审批订单
   - 点击"批准"（Approve）按钮
   - **验证**：
     - `workflow_state = Approved` ⭐ 工作流状态验证
     - `docstatus = 1` ⭐ **最关键验证点**（必须从 0 更新为 1）
     - `status = Submitted`（等待自动更新）

#### 步骤 4：验证后续流程 ⭐ 核心验证

**验证点 1：可以创建交货单**

1. **在订单页面操作**
   - 打开已批准的订单（`workflow_state = Approved`）
   - 点击"创建"（Create）按钮
   - 选择"交货单"（Delivery Note）

2. **验证结果**
   - ✅ 应该可以创建交货单
   - ✅ 交货单会自动填充订单信息
   - ✅ 可以保存并提交交货单

3. **如果无法创建交货单**
   - ❌ 检查 `docstatus` 是否为 `1` ⭐ **最关键检查**
   - ❌ 检查 `workflow_state` 是否为 `Approved`
   - ❌ 检查工作流动作是否更新了 `docstatus`
   - ❌ 检查用户角色权限
   - ❌ 检查工作流是否已启用

**验证点 2：可以创建发票**

1. **在订单页面操作**
   - 打开已批准的订单（`workflow_state = Approved`）
   - 点击"创建"（Create）按钮
   - 选择"发票"（Sales Invoice）

2. **验证结果**
   - ✅ 应该可以创建发票
   - ✅ 发票会自动填充订单信息
   - ✅ 可以保存并提交发票

3. **如果无法创建发票**
   - ❌ 检查 `docstatus` 是否为 `1` ⭐ **最关键检查**
   - ❌ 检查 `workflow_state` 是否为 `Approved`
   - ❌ 检查工作流动作是否更新了 `docstatus`
   - ❌ 检查用户角色权限（Accounts Manager 通常有发票权限）
   - ❌ 检查工作流是否已启用

**验证点 3：订单状态自动更新**

1. **创建交货单后**
   - 提交交货单
   - 返回销售订单页面
   - **验证**：
     - `status` 可能变为 `To Bill` 或 `To Deliver`（取决于开票情况）

2. **创建发票后**
   - 提交发票
   - 返回销售订单页面
   - **验证**：
     - `status` 可能变为 `To Deliver` 或 `Completed`（取决于交付情况）

3. **同时完成交付和开票**
   - 创建并提交交货单
   - 创建并提交发票
   - 返回销售订单页面
   - **验证**：
     - `status = Completed`
     - `workflow_state = Approved`（保持不变）

#### 步骤 5：验证审批拒绝流程

1. **审批拒绝**
   - 使用审批人员账号登录
   - 打开待审批订单
   - 点击"拒绝"（Reject）按钮
   - **验证**：
     - `workflow_state = Rejected`
     - `docstatus = 1`（保持不变）
     - `status = Submitted`（保持不变）

2. **重新提交**
   - 使用 Sales User 账号登录
   - 打开被拒绝的订单
   - ✅ 应该可以编辑订单（如果配置了 Rejected 状态允许编辑）
   - 修改订单后，点击"重新提交审批"
   - **验证**：
     - `workflow_state = Pending Approval`

### 9.2 验证检查清单

使用以下检查清单验证配置是否正确：

**工作流配置检查**：
- [ ] 工作流状态为 Active（启用）
- [ ] 已定义所有必需的状态（Draft、Submitted、Pending Approval、Approved、Rejected）
- [ ] 已配置所有必需的转换（Submit、Submit for Approval、Approve、Reject、Re-submit）
- [ ] Approve 转换的下一状态为 `Approved`

**权限配置检查**：
- [ ] Sales User 可以提交审批
- [ ] Sales Manager 可以审批通过和拒绝
- [ ] Accounts Manager 可以审批通过和拒绝
- [ ] 审批通过后，用户可以创建交货单
- [ ] 审批通过后，用户可以创建发票

**功能验证检查**：
- [ ] 提交订单后进入审批流程
- [ ] 审批人员可以看到待审批订单
- [ ] 审批通过后 `workflow_state = Approved`
- [ ] 审批通过后可以创建交货单
- [ ] 审批通过后可以创建发票
- [ ] 创建交货单后订单状态自动更新
- [ ] 创建发票后订单状态自动更新
- [ ] 审批拒绝后可以重新提交

### 9.3 常见问题排查

#### 问题 1：提交订单后没有进入审批状态

**症状**：
- 提交订单后，`workflow_state` 仍然是 `Submitted`，没有变为 `Pending Approval`

**排查步骤**：
1. 检查工作流是否启用（状态是否为 Active）
2. 检查工作流转换条件是否正确
   - "Submit for Approval" 转换的条件是否满足
   - 如果配置了金额条件，检查订单金额是否满足
3. 检查用户角色权限
   - 用户是否有 "Submit for Approval" 权限
4. 检查是否配置了自动触发
   - 如果没有配置自动触发，需要手动点击"提交审批"按钮

**解决方案**：
- 确保工作流状态为 Active
- 检查转换条件，确保满足触发条件
- 确保用户角色有相应权限

#### 问题 2：审批人员看不到待审批订单

**症状**：
- 审批人员登录后，在订单列表中看不到待审批的订单

**排查步骤**：
1. 检查订单的 `workflow_state` 是否为 `Pending Approval`
2. 检查用户角色权限
   - 用户是否有读取 Sales Order 的权限
3. 检查工作流转换的"允许的角色"设置
   - 审批人员的角色是否在 "Approve" 转换的允许角色列表中

**解决方案**：
- 确保订单 `workflow_state = Pending Approval`
- 确保审批人员角色有读取权限
- 确保审批人员角色在转换的允许角色列表中

#### 问题 3：审批通过后无法创建交货单/发票 ⭐ 关键问题

**症状**：
- 审批通过后，点击"创建"按钮，无法创建交货单或发票
- 系统提示"订单未批准"或"不允许此操作"

**排查步骤**：
1. **⭐ 最关键：检查 docstatus**
   - 打开订单详情页
   - 查看 `docstatus` 字段
   - ✅ 应该为 `1`（已提交）
   - ❌ 如果是 `0`（草稿），**这是问题所在**
   - **解决方案**：检查工作流动作是否配置了更新 `docstatus` 为 `1`

2. **检查工作流动作配置**
   - 打开工作流配置页面
   - 找到 "Approve" 动作
   - 检查是否配置了更新 `docstatus` 字段
   - ✅ 应该有一个动作：`docstatus = 1`
   - ❌ 如果没有，需要添加此动作

3. **检查工作流状态**
   - 打开订单详情页
   - 查看 `workflow_state` 字段
   - ✅ 应该为 `Approved`
   - ❌ 如果不是 `Approved`，说明审批未成功

4. **检查工作流转换配置**
   - 打开工作流配置页面
   - 检查 "Approve" 转换的"下一状态"是否为 `Approved`
   - ✅ 应该为 `Approved`
   - ❌ 如果设置错误，需要修改

5. **检查用户角色权限**
   - 检查当前用户角色是否有创建交货单/发票的权限
   - Sales User / Sales Manager：应该有创建交货单的权限
   - Accounts Manager：应该有创建发票的权限

6. **检查工作流是否启用**
   - 打开工作流配置页面
   - 检查工作流状态是否为 `Active`
   - ✅ 应该为 `Active`
   - ❌ 如果未启用，需要启用

**解决方案**：
- ⭐ **最关键**：确保工作流动作更新了 `docstatus = 1`
- 确保 `workflow_state = Approved`
- 确保工作流转换配置正确
- 确保工作流动作配置了更新 `docstatus` 字段
- 确保用户角色有相应权限
- 确保工作流已启用

#### 问题 4：审批后订单状态没有自动更新

**症状**：
- 审批通过后，创建交货单或发票，但订单的 `status` 字段没有更新

**排查步骤**：
1. 检查是否创建了交货单/发票
2. 检查交货单/发票是否已提交（`docstatus = 1`）
3. 检查订单的交付和开票情况
   - 查看 `per_delivered`（已交付百分比）
   - 查看 `per_billed`（已开票百分比）

**说明**：
- `status` 字段的更新由 ERPNEXT 系统自动计算
- 只有当交货单/发票提交后，系统才会更新订单状态
- 如果部分交付/开票，状态可能不会立即变为 `Completed`

**解决方案**：
- 确保交货单/发票已提交
- 等待系统自动更新（通常在保存后几秒内）
- 刷新订单页面查看最新状态

#### 问题 5：工作流状态更新但订单状态未更新

**症状**：
- `workflow_state` 已更新为 `Approved`
- 但 `status` 仍然是 `Submitted`，没有变为 `To Deliver and Bill`

**说明**：
- 这是**正常行为**
- `status` 字段只有在创建交货单或发票后才会更新
- 审批通过后，`status` 可能仍然是 `Submitted`，直到创建相关单据

**解决方案**：
- 这是正常行为，无需处理
- 创建交货单或发票后，`status` 会自动更新

---

## 十、审批后继续后续流程的关键要点总结

### 10.1 核心机制

**审批后继续后续流程的核心机制**：

```
1. 工作流控制 docstatus 的更新时机
   ↓
2. 待审批时：docstatus = 0（不能创建后续单据）
   ↓
3. 审批通过时：工作流动作将 docstatus 更新为 1
   ↓
4. docstatus = 1 时，ERPNEXT 允许创建后续单据
   ↓
5. 创建交货单/发票后，系统自动更新订单状态
   ↓
6. 订单状态根据交付和开票情况自动流转
```

### 10.2 关键配置点

**必须正确配置的 5 个关键点**：

1. **工作流状态定义**
   - ✅ 必须定义 `Approved` 状态
   - ✅ `Approved` 状态是允许继续后续流程的标志

2. **工作流动作配置** ⭐ **最关键**
   - ✅ "Approve" 动作必须更新 `docstatus = 1`
   - ✅ 这是最关键的配置，决定了审批后是否可以继续流程
   - ✅ 如果只更新 `workflow_state` 而不更新 `docstatus`，无法创建后续单据

3. **工作流转换配置**
   - ✅ "Approve" 转换的下一状态必须是 `Approved`
   - ✅ 这是工作流状态流转的配置

3. **权限配置**
   - ✅ 审批通过后，用户必须有创建交货单/发票的权限
   - ✅ Sales User / Sales Manager：创建交货单权限
   - ✅ Accounts Manager：创建发票权限

4. **工作流启用**
   - ✅ 工作流状态必须为 `Active`（启用）
   - ✅ 未启用的工作流不会生效

5. **订单提交状态**
   - ✅ 审批通过后，订单的 `docstatus` 必须为 `1`（已提交）
   - ✅ 只有 `docstatus = 1` 的订单才能创建后续单据
   - ✅ 工作流动作必须将 `docstatus` 从 0 更新为 1

### 10.3 状态流转完整流程

**从创建到完成的完整流程**：

```
1. 创建订单
   docstatus: 0
   workflow_state: Draft
   status: Draft
   ✅ 可以编辑

2. 提交订单（但不真正提交）
   docstatus: 0  ⚠️ 保持为 0
   workflow_state: Submitted
   status: Draft
   ❌ 不能创建后续单据（因为 docstatus = 0）

3. 提交审批（自动或手动）
   docstatus: 0  ⚠️ 保持为 0
   workflow_state: Pending Approval
   status: Draft
   ❌ 不能创建后续单据（因为 docstatus = 0）

4. 审批通过 ⭐ 关键步骤
   docstatus: 1  ⚠️ 关键变化：从 0 更新为 1
   workflow_state: Approved  ⭐ 工作流状态变化
   status: Submitted（自动更新）
   ✅ 可以创建交货单（因为 docstatus = 1）
   ✅ 可以创建发票（因为 docstatus = 1）
   ❌ 不能编辑订单

5. 创建交货单
   docstatus: 1
   workflow_state: Approved（保持不变）
   status: To Bill 或 To Deliver（自动更新）
   ✅ 可以继续创建发票

6. 创建发票
   docstatus: 1
   workflow_state: Approved（保持不变）
   status: To Deliver 或 Completed（自动更新）
   ✅ 可以继续创建交货单（如果未完全交付）

7. 完成交付和开票
   docstatus: 1
   workflow_state: Approved（保持不变）
   status: Completed（自动更新）
   ✅ 订单完成
```

### 10.4 验证清单

**配置完成后，使用以下清单验证**：

- [ ] 工作流已创建并启用（Active）
- [ ] 已定义 `Approved` 状态，且 docstatus 设置为 `1`
- [ ] 已定义 `Pending Approval` 状态，且 docstatus 设置为 `0`
- [ ] "Approve" 转换的下一状态为 `Approved`
- [ ] ⭐ **最关键**："Approve" 动作配置了更新 `docstatus = 1`
- [ ] 用户角色权限配置正确
- [ ] 提交订单后进入审批流程（docstatus = 0）
- [ ] 审批通过后 `workflow_state = Approved`
- [ ] ⭐ **最关键**：审批通过后 `docstatus = 1`
- [ ] 审批通过后可以创建交货单
- [ ] 审批通过后可以创建发票
- [ ] 创建交货单后订单状态自动更新
- [ ] 创建发票后订单状态自动更新

### 10.5 常见错误

**避免以下常见错误**：

1. ❌ **忘记启用工作流**
   - 工作流创建后必须设置为 `Active`
   - 未启用的工作流不会生效

2. ❌ **Approve 动作未更新 docstatus**
   - ⭐ **最常见错误**：只更新了 `workflow_state`，忘记更新 `docstatus`
   - 必须配置工作流动作更新 `docstatus = 1`
   - 如果 `docstatus` 仍然是 `0`，无法创建后续单据

3. ❌ **Approve 转换的下一状态设置错误**
   - 必须设置为 `Approved`
   - 如果设置为其他状态，审批后无法继续流程

3. ❌ **权限配置不完整**
   - 审批通过后，用户必须有创建后续单据的权限
   - 只有读取权限是不够的

4. ❌ **混淆 workflow_state、docstatus 和 status**
   - `docstatus`：⭐ **最关键**，控制是否可以创建后续单据（0=不能，1=可以）
   - `workflow_state`：控制审批流程状态
   - `status`：反映业务进度，由系统自动更新
   - 三者是不同的字段，有不同的用途
   - **关键**：ERPNEXT 检查的是 `docstatus`，而不是 `workflow_state`

5. ❌ **期望审批后 status 立即更新**
   - `status` 只有在创建交货单/发票后才会更新
   - 审批通过后，`status` 可能仍然是 `Submitted`，这是正常的

---

## 十一、相关文档

- [ERPNEXT 工作流官方文档](https://docs.erpnext.com/docs/user/manual/en/setting-up/workflows)
- [销售采购审批工作流](../business/workflows/sales-purchase-approval-flow.md)
- [ERPNEXT API 文档](./erpnext-api.md)

---

**文档版本**：v1.0  
**创建时间**：2025-01-17  
**维护者**：TTPOS Team

