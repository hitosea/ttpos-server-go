---
name: guiding-specs
description: 指导 Spec 规范,包括命名规则、目录结构、状态流转、Story Point 评估。当用户创建 Spec、询问需求文档规范、或 Spec 命名结构时触发。
---

# Spec 规范指南

## Proposal → Spec 拆分

**Proposal 是 1:N 关系到 Spec**，需要多维度叠加拆分：

```
Proposal ──→ 按终端拆 ──→ 按复杂度拆 ──→ 按功能模块拆 ──→ ... ──→ Spec (SP ≤ 5)
```

### 拆分维度（按优先级）

| 优先级 | 维度       | 触发条件         |
| ------ | ---------- | ---------------- |
| 1      | 终端       | 涉及多终端       |
| 2      | 复杂度(SP) | SP > 5           |
| 3      | 功能模块   | 功能可独立交付   |
| 4      | 用户角色   | 不同角色不同功能 |
| 5      | Phase      | 需要分批上线     |
| 6      | 依赖层级   | 有前后依赖       |

> 详细拆分策略见 [rules.md](rules.md#proposal--spec-拆分策略)

---

## 两阶段 Interview 模式

Spec 创建分为两个阶段，由不同角色负责，且不同 Level 流程有差异：

| 阶段     | 负责人 | 输出文件             | 前置依赖        |
| -------- | ------ | -------------------- | --------------- |
| 需求阶段 | 产品组 | requirements.md      | 按 Level 不同   |
| 设计阶段 | 开发组 | design.md + tasks.md | requirements.md |

### Level 流程差异

| Level | 阶段一触发        | 阶段二触发              | 特殊处理                        |
| ----- | ----------------- | ----------------------- | ------------------------------- |
| story | Proposal 通过后   | requirements 审核通过后 | 完整流程                        |
| task  | 技术需求/债务识别 | requirements 审核通过后 | 可跳过 Proposal                 |
| bug   | Bug 报告确认      | requirements 审核通过后 | 可跳过 Proposal，可简化 design  |
| spike | 技术不确定性识别  | **不触发**              | 仅输出调研报告，无 design/tasks |

---

## 阶段一：需求定义（产品组）

### ⚠️ 强制规则

1. **必须完成 3 轮采访**，不得跳过或提前结束
2. **采访结果用于填充模板**，禁止直接拷贝模板占位符
3. 每轮必须等待用户回答后，才能进入下一轮

### 触发条件

**通用触发：**
- 用户说"创建 Spec"、"写需求"但未提供完整信息
- `/spec:create` 参数不完整或缺失

**按 Level 触发：**

| Level | 触发场景          | 前置依赖 | 关键词识别                       |
| ----- | ----------------- | -------- | -------------------------------- |
| story | Proposal 审核通过 | Proposal | "新功能"、"用户想要"             |
| task  | 技术债务/优化需求 | 无       | "重构"、"优化"、"迁移"、"升级"   |
| bug   | Bug 报告确认      | Bug 报告 | "修复"、"报错"、"崩溃"、"异常"   |
| spike | 技术不确定性      | 无       | "调研"、"评估"、"可行性"、"探索" |

### 采访流程（3 轮 + 拆分检查）

```
Round 1 (What & Where): 什么类型？哪个终端？
    ├─ Q1: Level（story/task/bug/spike）
    └─ Q2: 终端（根据功能推断 2-3 个选项）
           ↓
┌─────────────────────────────────────────────────────┐
│ ⚠️ 拆分检查（强制）                                  │
│ IF 终端 == "all" OR 涉及多终端:                      │
│   询问: 各终端操作流程和功能范围是否一致？            │
│   - 完全一致 → 继续，允许 all                        │
│   - 有差异 → 停止，输出拆分建议，逐个创建 Spec        │
└─────────────────────────────────────────────────────┘
           ↓
Round 2 (Who & What): 给谁？做什么？
    ├─ Q3: 目标用户/角色
    └─ Q4: 核心功能描述
           ↓
Round 3 (Value & Acceptance): 什么价值？怎么验收？
    ├─ Q5: 业务价值
    └─ Q6: 关键验收标准
```

### 拆分检查规则

**触发条件：** Round 1 完成后，终端为 `all` 或涉及多个终端

**询问内容（产品视角）：**
```
请确认各终端的功能需求：
1. 各终端的**操作流程**是否一致？
2. 各终端的**功能范围**是否相同？
3. 是否有终端需要**特殊处理**？
```

**判断标准（产品视角）：**

| 情况                               | 判断     | 动作       |
| ---------------------------------- | -------- | ---------- |
| 所有终端操作流程、功能范围完全一致 | 完全一致 | 允许 `all` |
| 不同终端有不同的操作入口或流程     | 有差异   | 按终端拆分 |
| 某些终端有额外功能或限制           | 有差异   | 按终端拆分 |

**拆分时输出：**
```
⚠️ 需要拆分为多个 Spec：
1. story-pos-{feature} - POS 端实现
2. story-shop-{feature} - Shop 端实现
3. story-mobile-{feature} - Mobile 端实现

请逐个执行 /spec:create 创建。
```

| Round | 问题     | options 策略                        | 对应模板章节      |
| ----- | -------- | ----------------------------------- | ----------------- |
| 1     | Level    | story/task/bug/spike                | 基本信息-Spec ID  |
| 1     | 终端     | 根据功能描述推断最可能的 2-3 个     | 基本信息-Spec ID  |
| 2     | 角色     | 根据终端推断（pos→收银员/店长）     | 用户故事-作为     |
| 2     | 功能     | 根据描述提炼核心功能点              | 用户故事-我想     |
| 3     | 价值     | 效率提升/成本降低/体验改善/合规需求 | 用户故事-以便于   |
| 3     | 验收标准 | 根据功能推断 WHEN-THEN-SHALL 格式   | 功能需求-验收标准 |

**示例：用户说"pos 端需要快捷支付功能"**

```yaml
Round 1: AskUserQuestion
  question: |
    关于这个 Spec，请确认：
    1. 这是什么类型的需求？
    2. 主要在哪个终端实现？
  options:
    类型:
      - story: 用户故事（新功能）
      - task: 技术任务（重构/优化）
    终端:
      - pos: 收银端（主要）
      - assistant: 助手端（可能需要同步）

Round 2: AskUserQuestion
  question: |
    1. 这个功能主要服务谁？
    2. 核心要实现什么？
  options:
    角色:
      - 收银员
      - 店长
    功能:
      - 快捷选择常用支付方式
      - 自定义支付方式排序
      - 记住上次支付方式

Round 3: AskUserQuestion
  question: |
    1. 这个功能能带来什么价值？
    2. 怎么验证功能完成？
  options:
    价值:
      - 提升结账效率
      - 减少顾客等待
      - 降低操作错误
    验收:
      - 支付方式列表显示快捷入口
      - 点击快捷入口直接发起支付
      - 支持自定义快捷支付方式
```

### 采访完成后

1. **生成命名**: `{level}-{app}-{feature}`
2. **填充模板**: 用采访结果填充 requirements.md
3. **创建目录**: `docs/shared/specs/active/{level}-{app}-{feature}/`
4. **状态设置**: 待审核

### 采访结果 → 模板映射

| 采访维度 | 填充内容             |
| -------- | -------------------- |
| Q1 Level | Spec ID 前缀         |
| Q2 终端  | Spec ID + 平台兼容性 |
| Q3 角色  | 用户故事-作为        |
| Q4 功能  | 用户故事-我想        |
| Q5 价值  | 用户故事-以便于      |
| Q6 验收  | 功能需求-验收标准    |

---

## 阶段二：技术设计（开发组）

### ⚠️ 强制规则

1. **必须完成 3 轮采访**，不得跳过或提前结束
2. **SP > 5 必须拆分**，在采访中识别并引导拆分
3. 每轮必须等待用户回答后，才能进入下一轮

### 触发条件

**通用触发：**
- requirements.md 审核通过
- 用户说"设计方案"、"技术设计"、"任务分解"
- `/spec:design` 命令

**按 Level 触发：**

| Level | 触发条件              | 输出文件             | 特殊说明                         |
| ----- | --------------------- | -------------------- | -------------------------------- |
| story | requirements 审核通过 | design.md + tasks.md | 完整流程                         |
| task  | requirements 审核通过 | design.md + tasks.md | 完整流程                         |
| bug   | requirements 审核通过 | design.md + tasks.md | 可简化，合并 design 内容到 tasks |
| spike | **不触发阶段二**      | -                    | 调研结果直接写入 requirements.md |

### 采访流程（3 轮 6 维度）

```
Round 1 (How): 怎么实现？复用什么？
    ├─ Q1: 架构方案（新建/扩展/重构）
    └─ Q2: 可复用组件/代码
           ↓
Round 2 (Interface): 接口设计
    ├─ Q3: API 设计（新增/修改/复用）
    └─ Q4: 数据模型变更
           ↓
Round 3 (Estimate): 评估与分解
    ├─ Q5: SP 评估（复杂度+风险）
    └─ Q6: 任务分解（Phase 划分）
           ↓
自动推断: 风险识别、测试策略、平台兼容性
```

| Round | 问题     | options 策略                    | 对应模板章节        |
| ----- | -------- | ------------------------------- | ------------------- |
| 1     | 架构     | 新建组件/扩展现有/重构          | design-架构设计     |
| 1     | 复用     | 根据功能推断可复用代码          | design-代码复用分析 |
| 2     | API      | 新增/修改/复用现有 API          | design-API 设计     |
| 2     | 模型     | 新增/修改/复用现有模型          | design-数据模型     |
| 3     | SP       | 根据复杂度+风险因素计算         | tasks-进度总览      |
| 3     | 任务分解 | 按 Phase 划分（核心/集成/测试） | tasks-Phase 列表    |
| 推断  | 风险     | 根据功能模块自动判断            | design-风险识别     |
| 推断  | 测试     | 根据覆盖率要求推断              | design-测试策略     |

**示例：基于 story-pos-quick-payment 的 requirements**

```yaml
Round 1: AskUserQuestion
  question: |
    基于需求文档，请确认技术方案：
    1. 采用什么架构方案？
    2. 有哪些可复用的代码？
  options:
    架构:
      - 扩展现有 PaymentController
      - 新建 QuickPaymentWidget 组件
    复用:
      - packages/ui/lib/payment/ 现有支付组件
      - PaymentMethodModel 数据模型

Round 2: AskUserQuestion
  question: |
    1. API 层面需要什么变更？
    2. 数据模型需要调整吗？
  options:
    API:
      - 复用现有支付 API
      - 新增快捷支付配置 API
    模型:
      - 扩展 PaymentMethodModel 添加 isQuick 字段
      - 新建 QuickPaymentConfig 模型

Round 3: AskUserQuestion
  question: |
    1. 评估 Story Point？
    2. 如何分解任务？
  options:
    SP:
      - SP3: 简单扩展，风险低
      - SP5: 中等复杂度，需要测试
    Phase:
      - Phase1: 核心组件实现
      - Phase2: POS 端集成
      - Phase3: 测试与文档
```

### 采访完成后

1. **SP 检查**: 若 SP > 5，引导拆分
2. **填充模板**: 用采访结果填充 design.md + tasks.md
3. **状态更新**: 已通过 → 开发中

### 采访结果 → 模板映射

| 采访维度 | 填充到    | 填充内容         |
| -------- | --------- | ---------------- |
| Q1 架构  | design.md | 架构设计章节     |
| Q2 复用  | design.md | 代码复用分析章节 |
| Q3 API   | design.md | API 设计章节     |
| Q4 模型  | design.md | 数据模型章节     |
| Q5 SP    | tasks.md  | 进度总览-总 SP   |
| Q6 任务  | tasks.md  | Phase 列表       |
| 推断风险 | design.md | 风险识别章节     |
| 推断测试 | design.md | 测试策略章节     |

---

## 通用规范

### 命名规则

**格式:** `{level}-{app}-{feature}`

| Level | 说明     | 场景                                     | 示例                       |
| ----- | -------- | ---------------------------------------- | -------------------------- |
| story | 用户故事 | 用户可感知的新功能，有明确业务价值       | story-pos-quick-payment    |
| task  | 技术任务 | 用户不可见的技术改进（重构/优化/迁移）   | task-api-refactor          |
| bug   | 缺陷修复 | 修复已有功能的问题                       | bug-order-calculation      |
| spike | 技术调研 | 探索性工作，验证技术可行性，输出调研报告 | spike-websocket-evaluation |

**Level 选择决策：**
```
需求来了
    │
    ├─ 用户能感知到吗？
    │       │
    │       ├─ 是 → 是新功能还是修复？
    │       │           ├─ 新功能 → story
    │       │           └─ 修复问题 → bug
    │       │
    │       └─ 否 → 是探索还是确定要做？
    │                   ├─ 需要先调研 → spike
    │                   └─ 确定要做 → task
```

**App:** pos, shop, kds, qds, assistant, tablet, mobile, menu, member, kiosk, all

### 目录结构

```
docs/shared/specs/
├── active/
│   ├── story-pos-quick-payment/   # story/task/bug 完整结构
│   │   ├── requirements.md        # 需求文档（产品组）
│   │   ├── design.md              # 设计文档（开发组）
│   │   └── tasks.md               # 任务清单（开发组）
│   │
│   └── spike-websocket-evaluation/ # spike 简化结构
│       └── requirements.md         # 仅调研报告
│
├── archived/{version}/        # 已归档
└── deprecated/                # 已废弃
```

### 状态流转

```
                    产品组                              开发组
                      │                                   │
Proposal ──→ [草稿] ──→ [待审核] ──→ [已通过] ──→ [开发中] ──→ [待测试] ──→ [已验证] ──→ [已完成]
              │                        │           │
              │    requirements.md     │   design.md + tasks.md
              └────────────────────────┴───────────────────────────────────────────────────────→ [已归档]
```

### Story Point 评估

| SP  | 工作量   | 复杂度 | 行动         |
| --- | -------- | ------ | ------------ |
| 1   | 0.5-1 天 | 极简单 | 可开发       |
| 3   | 1-2 天   | 简单   | 可开发       |
| 5   | 2-3 天   | 中等   | 可开发       |
| 8   | 4-6 天   | 复杂   | **必须拆分** |
| 13  | 1-2 周   | 很复杂 | **必须拆分** |

### 风险加成

| 因素       | 加分    | 触发条件                  |
| ---------- | ------- | ------------------------- |
| 高风险模块 | +1      | payment, order, cart 相关 |
| 新技术     | +1      | 首次使用技术栈            |
| 多端适配   | +0.5/端 | 需要支持多个终端          |
| 测试要求高 | +0.5    | 覆盖率 100%、复杂测试场景 |

### SP > 5 拆分策略

```yaml
拆分原则:
  - 按应用垂直拆分（优先）
  - 按功能模块拆分
  - 按 Phase 拆分

示例:
  ❌ story-pos-shop-kds-order-sync  # SP13, 跨3个应用
  ✅ story-pos-order-sync           # SP3
  ✅ story-shop-order-sync          # SP3
  ✅ story-kds-order-sync           # SP3
```

---

## 命令参考

| 命令           | 阶段   | 说明                             |
| -------------- | ------ | -------------------------------- |
| `/spec:create` | 阶段一 | 创建 Spec 目录 + requirements.md |
| `/spec:design` | 阶段二 | 补充 design.md + tasks.md        |

## 详细规范

- [完整规则](rules.md) - 命名详解、SP 评估、拆分策略
- [需求模板](templates/requirements.md)
- [设计模板](templates/design.md)
- [任务模板](templates/tasks.md)
