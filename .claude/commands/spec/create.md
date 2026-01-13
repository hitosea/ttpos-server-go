---
description: 创建 Spec 需求文档
argument-hint: [level-app-feature]
allowed-tools: Read, Write, Glob, Bash(TZ=Asia/Shanghai date:*), Bash(git config user.name), AskUserQuestion
---

# 上下文

- 当前日期: !`TZ=Asia/Shanghai date +%Y-%m-%d`
- 创建人: !`git config user.name`
- 现有 Specs: !`ls docs/shared/specs/active/ 2>/dev/null | head -5`
- 现有 Proposals: !`ls docs/team/proposals/ 2>/dev/null | tail -3`

# 任务

创建 Spec 的需求文档（requirements.md）。通过 3 轮采访 + 拆分检查收集信息。

## 强制规则

1. **必须完成 3 轮采访**，不得跳过或提前结束
2. **必须执行拆分检查**（多终端时）
3. **采访结果用于填充模板**，禁止保留占位符
4. 每轮必须等待用户回答后，才能进入下一轮
5. **必须使用 AskUserQuestion 工具**，不能用文本模拟提问

## 触发条件

| Level | 触发场景 | 前置依赖 | 关键词识别 |
|-------|---------|---------|-----------|
| story | Proposal 审核通过 | Proposal | "新功能"、"用户想要" |
| task | 技术债务/优化需求 | 无 | "重构"、"优化"、"迁移"、"升级" |
| bug | Bug 报告确认 | Bug 报告 | "修复"、"报错"、"崩溃"、"异常" |
| spike | 技术不确定性 | 无 | "调研"、"评估"、"可行性"、"探索" |

---

## 采访流程

### Round 1: What & Where (什么类型？哪个终端？)

使用 `AskUserQuestion` 询问：

```yaml
question: |
  关于这个 Spec，请确认：
  1. 这是什么类型的需求？
  2. 主要在哪个终端实现？
options:
  类型:
    - story: 用户故事（新功能）
    - task: 技术任务（重构/优化）
    - bug: 缺陷修复
    - spike: 技术调研
  终端: (根据 $1 描述推断最可能的 2-3 个)
    - pos: 收银端
    - shop: 商家管理端
    - all: 所有终端
    - ...
```

**采集结果：**
- Q1 Level → Spec ID 前缀
- Q2 终端 → Spec ID + 平台兼容性

**⛔ STOP - 等待用户回答**

---

### 拆分检查（强制）

**触发条件：** Round 1 完成后，终端为 `all` 或涉及多个终端

使用 `AskUserQuestion` 询问：

```yaml
question: |
  请确认各终端的功能需求：
  1. 各终端的**操作流程**是否一致？
  2. 各终端的**功能范围**是否相同？
  3. 是否有终端需要**特殊处理**？
options:
  - 完全一致: 允许使用 all，继续流程
  - 有差异: 需要按终端拆分
```

**判断标准：**

| 情况 | 判断 | 动作 |
|------|------|------|
| 所有终端操作流程、功能范围完全一致 | 完全一致 | 允许 `all`，继续 |
| 不同终端有不同的操作入口或流程 | 有差异 | 按终端拆分 |
| 某些终端有额外功能或限制 | 有差异 | 按终端拆分 |

**若有差异，输出拆分建议并退出：**

```markdown
⚠️ 需要拆分为多个 Spec：

1. {level}-pos-{feature} - POS 端实现
2. {level}-shop-{feature} - Shop 端实现
3. {level}-mobile-{feature} - Mobile 端实现

请逐个执行 `/spec:create` 创建。
```

**⛔ STOP - 等待用户确认**

---

### Round 2: Who & What (给谁？做什么？)

使用 `AskUserQuestion` 询问：

```yaml
question: |
  1. 这个功能主要服务谁？
  2. 核心要实现什么？
options:
  角色: (根据终端推断)
    - pos: 收银员、店长
    - shop: 店长、商户管理员、运营人员
    - kds: 厨师、后厨员工
    - ...
  功能: (根据 $1 描述提炼核心功能点)
    - 功能点1
    - 功能点2
    - 功能点3
```

**采集结果填充模板：**
- Q3 角色 → 用户故事-作为
- Q4 功能 → 用户故事-我想

**⛔ STOP - 等待用户回答**

---

### Round 3: Value & Acceptance (什么价值？怎么验收？)

使用 `AskUserQuestion` 询问：

```yaml
question: |
  1. 这个功能能带来什么价值？
  2. 怎么验证功能完成？
options:
  价值:
    - 效率提升
    - 成本降低
    - 体验改善
    - 合规需求
  验收: (根据功能推断，使用 WHEN-THEN-SHALL 格式)
    - 当 {条件} 时，系统应该 {行为}
    - 当 {条件} 时，用户应该能 {操作}
```

**采集结果填充模板：**
- Q5 价值 → 用户故事-以便于
- Q6 验收 → 功能需求-验收标准

**⛔ STOP - 等待用户回答**

---

## 生成 Spec

### Step 1: 生成命名

**格式:** `{level}-{app}-{feature}`

| Level | 说明 | 场景 | 示例 |
|-------|------|------|------|
| story | 用户故事 | 用户可感知的新功能 | story-pos-quick-payment |
| task | 技术任务 | 用户不可见的技术改进 | task-api-refactor |
| bug | 缺陷修复 | 修复已有功能的问题 | bug-order-calculation |
| spike | 技术调研 | 探索性工作，验证可行性 | spike-websocket-evaluation |

### Step 2: 确认命名

使用 `AskUserQuestion` 确认：
```yaml
question: "Spec 命名为 {name}，确认吗？"
options:
  - 确认
  - 修改（请输入新名称）
```

### Step 3: 搜索关联 Proposal（story 类型必须）

```bash
# 搜索相关 Proposal
ls docs/team/proposals/*/{app}-*.md 2>/dev/null
```

如果是 `story` 类型且没有找到关联 Proposal，询问用户：
```yaml
question: "Story 类型通常需要关联 Proposal。是否有已批准的 Proposal？"
options:
  - 有，路径是...
  - 没有，这是技术驱动的需求
```

### Step 4: 填充模板

读取模板: `.claude/skills/guiding-specs/templates/requirements.md`

用采访结果填充模板：

| 采访维度 | 填充到模板章节 |
|---------|--------------|
| Q1 Level | 基本信息-Spec ID 前缀 |
| Q2 终端 | 基本信息-Spec ID + 平台兼容性 |
| Q3 角色 | 用户故事-作为 |
| Q4 功能 | 用户故事-我想 |
| Q5 价值 | 用户故事-以便于 |
| Q6 验收 | 功能需求-验收标准 |

**Go 后端项目技术约束（自动填充）：**
```markdown
## 技术约束

- Go 版本: 1.23+
- 框架: Gin + GORM
- 分层架构: API → Service → Repository → Model
- 测试覆盖率: ≥ 80%
- 必须遵循 CLAUDE.md 和 .cursor/rules/go-main.mdc 规范
```

### Step 5: 创建目录和文件

```bash
# 创建 Spec 目录
mkdir -p docs/shared/specs/active/{level}-{app}-{feature}

# 创建 requirements.md
docs/shared/specs/active/{level}-{app}-{feature}/requirements.md
```

### Step 6: 输出结果

```markdown
✅ Spec 需求文档已创建

| 项目 | 内容 |
|------|------|
| Spec ID | {level}-{app}-{feature} |
| 文件路径 | docs/shared/specs/active/{spec-id}/requirements.md |
| 审核状态 | 待审核 |
| Level | {level} |

## 下一步

1. 提交产品组审核 requirements.md
2. 审核通过后执行 `/spec:design {spec-id}` 创建技术设计
```

---

## Level 流程差异

| Level | 阶段一触发 | 阶段二触发 | 特殊处理 |
|-------|-----------|-----------|---------|
| story | Proposal 通过后 | requirements 审核通过后 | 完整流程 |
| task | 技术需求/债务识别 | requirements 审核通过后 | 可跳过 Proposal |
| bug | Bug 报告确认 | requirements 审核通过后 | 可跳过 Proposal，可简化 design |
| spike | 技术不确定性识别 | **不触发** | 仅输出调研报告，无 design/tasks |

---

## 示例

**用户输入:** `/spec:create pos 端快捷支付`

**Round 1:**
```yaml
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
```

**Round 2:**
```yaml
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
```

**Round 3:**
```yaml
question: |
  1. 这个功能能带来什么价值？
  2. 怎么验证功能完成？
options:
  价值:
    - 提升结账效率
    - 减少顾客等待
    - 降低操作错误
  验收:
    - 当用户进入支付界面时，应显示快捷支付入口
    - 当用户点击快捷入口时，应直接发起支付
    - 当用户设置快捷方式后，应在下次生效
```

**生成:** `story-pos-quick-payment`

**输出:** `docs/shared/specs/active/story-pos-quick-payment/requirements.md`
