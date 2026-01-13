---
description: 创建需求提案
argument-hint: [feature-description]
allowed-tools: Read, Write, Bash(TZ=Asia/Shanghai date:*), Bash(git config user.name), AskUserQuestion
---

# 上下文

- 当前月份: !`TZ=Asia/Shanghai date +%Y-%m`
- 当前日期: !`TZ=Asia/Shanghai date +%Y-%m-%d`
- 提案人: !`git config user.name`
- 提案目录: !`ls docs/team/proposals/ 2>/dev/null | tail -3`

# 任务

创建一个新的需求提案（Proposal）。通过 3 轮采访收集必要信息，填充模板并生成提案文件。

## 强制规则

1. **必须完成 3 轮采访**，不得跳过或提前结束
2. **采访结果用于填充模板**，禁止保留占位符
3. 每轮必须等待用户回答后，才能进入下一轮
4. **必须使用 AskUserQuestion 工具**，不能用文本模拟提问

## 触发条件

- 用户说"有个想法"、"想做个功能"但未提供完整信息
- `/proposal:create` 参数不完整或缺失
- 用户不确定功能应归属哪个终端

---

## 采访流程

### Round 1: What & Where (做什么？在哪？)

使用 `AskUserQuestion` 询问：

```yaml
question: |
  关于这个功能，请回答：
  1. 主要在哪个终端使用？
  2. 目前遇到什么问题？
options:
  终端: (根据 $1 描述动态推断 2-3 个最可能的选项)
    - pos: 收银端
    - shop: 商家管理端
    - kds: 厨显端
    - assistant: 助手端
    - mobile: 扫码点餐端
    - ...
  问题: (根据终端特性提供常见痛点)
    - 选项1
    - 选项2
    - 其他（请描述）
```

**采集结果填充模板：**
- Q1 终端 → 影响范围-涉及终端
- Q2 问题 → 背景和动机-问题描述

**⛔ STOP - 等待用户回答**

---

### Round 2: Why & Who (为什么？给谁？)

使用 `AskUserQuestion` 询问：

```yaml
question: |
  1. 解决这个问题能带来什么价值？
  2. 这个功能主要服务谁？
options:
  价值:
    - 效率提升
    - 成本降低
    - 体验改善
    - 合规需求
  用户: (根据终端推断，参考下方映射表)
    - 收银员（主要）
    - 店长（次要）
    - ...
```

**终端→用户映射表：**
| 终端 | 主要用户 | 次要用户 |
|------|---------|---------|
| pos | 收银员 | 店长 |
| shop | 店长、商户管理员 | 运营人员 |
| kds | 厨师、后厨员工 | - |
| qds | 顾客（通过前台下单） | - |
| assistant | 店员 | 收银员 |
| tablet | 店员、顾客 | - |
| mobile | 顾客 | - |
| menu | 顾客 | - |
| member | 会员顾客 | - |
| kiosk | 顾客 | - |

**采集结果填充模板：**
- Q3 价值 → 背景和动机-业务价值
- Q4 用户 → 背景和动机-目标用户

**⛔ STOP - 等待用户回答**

---

### Round 3: How (怎么做？)

使用 `AskUserQuestion` 询问：

```yaml
question: "你期望怎么解决？"
options: (根据问题推断可能的方向)
  - 方案1
  - 方案2
  - 方案3
  - 其他
```

**采集结果填充模板：**
- Q5 方案 → 解决方案概述 + 核心功能点

**自动推断：**
- 影响范围-涉及模块（根据方案推断）
- 初步评估-技术复杂度（根据方案推断）
- 附录-User Story：`作为{用户}，我想{方案}，以便{价值}`

**⛔ STOP - 等待用户回答**

---

## 生成提案

### Step 1: 生成命名

**格式:** `{app}-{feature-name}`

- **app**: pos | shop | kds | qds | assistant | tablet | mobile | menu | member | kiosk | all
- **feature-name**: 从用户描述中提炼英文 kebab-case（2-4 个单词）

**示例:**
- `pos-quick-payment`
- `shop-report-export`
- `all-dark-mode`

### Step 2: 确认命名

使用 `AskUserQuestion` 确认：
```yaml
question: "提案命名为 {name}，确认吗？"
options:
  - 确认
  - 修改（请输入新名称）
```

### Step 3: 填充模板

读取模板: `.claude/skills/guiding-proposals/template.md`

用采访结果填充模板：

| 采访维度 | 填充到模板章节 |
|---------|--------------|
| Q1 终端 | 影响范围-涉及终端 |
| Q2 问题 | 🎯 背景和动机-问题描述 |
| Q3 价值 | 🎯 背景和动机-业务价值 |
| Q4 用户 | 🎯 背景和动机-目标用户 |
| Q5 方案 | 💡 解决方案概述 |
| 推断-模块 | 影响范围-涉及模块 |
| 推断-复杂度 | 📊 初步评估-技术复杂度 |
| 推断-User Story | 📝 附录-User Story |

**禁止保留占位符！**

### Step 4: 创建文件

```bash
# 创建目录（如不存在）
mkdir -p docs/team/proposals/{YYYY-MM}

# 创建提案文件
docs/team/proposals/{YYYY-MM}/{name}.md
```

### Step 5: 更新索引

更新 `docs/team/proposals/README.md`（如存在）

### Step 6: 输出结果

```markdown
✅ 提案已创建

| 项目 | 内容 |
|------|------|
| 提案名称 | {name} |
| 文件路径 | docs/team/proposals/{YYYY-MM}/{name}.md |
| 状态 | 待评审 |
| 下一步 | 提交评审会议，通过后执行 `/spec:create story-{app}-{feature}` |
```

---

## 示例

**用户输入:** `/proposal:create 收银时快速选择支付方式`

**Round 1:**
```yaml
question: |
  关于这个功能，请回答：
  1. 主要在哪个终端使用？
  2. 目前遇到什么问题？
options:
  终端:
    - pos: 收银端结账流程
    - assistant: 助手端辅助收银
  问题:
    - 支付方式太多，选择慢
    - 常用支付没有快捷入口
    - 其他（请描述）
```

**Round 2:**
```yaml
question: |
  1. 解决这个问题能带来什么价值？
  2. 这个功能主要服务谁？
options:
  价值:
    - 提升结账效率
    - 减少顾客等待
    - 降低操作错误
  用户:
    - 收银员（主要）
    - 店长（次要）
```

**Round 3:**
```yaml
question: "你期望怎么解决？"
options:
  - 添加快捷支付按钮
  - 记住上次支付方式
  - 自定义支付方式排序
  - 其他
```

**生成命名:** `pos-quick-payment`

**输出文件:** `docs/team/proposals/2026-01/pos-quick-payment.md`
