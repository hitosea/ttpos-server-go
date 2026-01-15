---
description: 检查 Spec 开发进度
argument-hint: [spec-id]
allowed-tools: Read, Glob
---

# 上下文

- 当前日期: !`TZ=Asia/Shanghai date +%Y-%m-%d`
- Active Specs: !`ls docs/shared/specs/active/ 2>/dev/null`

# 任务

检查 Spec 的开发进度，解析 tasks.md 并生成进度报告。

---

## 执行流程

### Step 1: 定位 Spec

**如果指定了 spec-id ($1):**
```
docs/shared/specs/active/{spec-id}/tasks.md
```

**如果未指定或 $1 == "all":**
```
扫描 docs/shared/specs/active/*/tasks.md
```

### Step 2: 读取文件

对每个 Spec 目录，读取：
- `requirements.md` - 获取状态信息
- `tasks.md` - 获取任务进度

### Step 3: 解析任务

**任务格式识别:**
```markdown
- [ ] 1.1 任务标题     → 未完成
- [x] 1.2 任务标题     → 已完成
```

**按 Phase 分组统计:**
```
Phase 1: 核心实现
  - 总任务: 5
  - 已完成: 3
  - 完成率: 60%

Phase 2: API 集成
  - 总任务: 3
  - 已完成: 0
  - 完成率: 0%
```

### Step 4: 生成报告

---

## 输出格式

### 单个 Spec 报告

```markdown
# 📊 {spec-id} 进度报告

## 基本信息

| 项目 | 内容 |
|------|------|
| Spec ID | {spec-id} |
| 状态 | {status} |
| 总 SP | {total_sp} |
| 检查时间 | {YYYY-MM-DD HH:mm} |

## 进度总览

| 指标 | 数值 |
|------|------|
| 总任务 | {total} |
| 已完成 | {completed} |
| 待完成 | {pending} |
| 完成率 | {percentage}% |

## Phase 进度

| Phase | 总任务 | 已完成 | 进度条 | 完成率 |
|-------|--------|--------|--------|--------|
| Phase 1: 核心实现 | {n} | {m} | {bar} | {x}% |
| Phase 2: API 集成 | {n} | {m} | {bar} | {x}% |
| Phase 3: 测试文档 | {n} | {m} | {bar} | {x}% |

## 已完成任务 ✅

- [x] 1.1 {任务标题}
- [x] 1.2 {任务标题}
- [x] 1.3 {任务标题}

## 待完成任务 ⏳

- [ ] 2.1 {任务标题}
- [ ] 2.2 {任务标题}
- [ ] 3.1 {任务标题}

## 🚀 下一步行动

1. 继续完成 Phase {N} 的剩余任务
2. 下一个任务: {next_task}
3. 预计剩余工作量: {remaining_sp} SP
```

**进度条格式:**
```
0%   ░░░░░░░░░░░░
25%  ███░░░░░░░░░
50%  ██████░░░░░░
75%  █████████░░░
100% ████████████
```

---

### 全部 Specs 报告

```markdown
# 📊 Active Specs 进度总览

检查时间: {YYYY-MM-DD HH:mm}

## 汇总统计

| 指标 | 数值 |
|------|------|
| Active Specs | {count} |
| 可归档 | {archived_count} |
| 开发中 | {in_progress_count} |

## 详细进度

| Spec ID | 状态 | 总任务 | 已完成 | 完成率 | 进度条 |
|---------|------|--------|--------|--------|--------|
| story-pos-quick-payment | 开发中 | 15 | 7 | 47% | ██████░░░░░░ |
| task-shop-export | 开发中 | 5 | 5 | 100% | ████████████ |
| bug-order-calc | 待测试 | 3 | 3 | 100% | ████████████ |

## 状态分布

- ✅ 可归档 (100%): {count} 个
- 🚧 开发中 (<100%): {count} 个
- ⏳ 待开始 (0%): {count} 个

## 建议操作

1. **可归档的 Spec:**
   - `task-shop-export` - 执行 `/spec:archive task-shop-export`

2. **需要关注的 Spec:**
   - `story-pos-quick-payment` - 进度 47%，继续开发
```

---

## 状态判断规则

| 完成率 | 状态 | 图标 | 建议操作 |
|--------|------|------|---------|
| 0% | 待开始 | ⏳ | 开始开发 |
| 1-99% | 开发中 | 🚧 | 继续开发 |
| 100% | 可归档 | ✅ | 执行归档 |

---

## 示例

### 示例 1: 检查单个 Spec

**输入:** `/spec:check story-pos-quick-payment`

**输出:**
```markdown
# 📊 story-pos-quick-payment 进度报告

## 基本信息

| 项目 | 内容 |
|------|------|
| Spec ID | story-pos-quick-payment |
| 状态 | 开发中 |
| 总 SP | 5 |
| 检查时间 | 2026-01-13 15:30 |

## 进度总览

| 指标 | 数值 |
|------|------|
| 总任务 | 15 |
| 已完成 | 7 |
| 待完成 | 8 |
| 完成率 | 47% |

## Phase 进度

| Phase | 总任务 | 已完成 | 进度条 | 完成率 |
|-------|--------|--------|--------|--------|
| Phase 1: 核心实现 | 10 | 7 | ████████░░░░ | 70% |
| Phase 2: API 集成 | 3 | 0 | ░░░░░░░░░░░░ | 0% |
| Phase 3: 测试文档 | 2 | 0 | ░░░░░░░░░░░░ | 0% |

## 已完成任务 ✅

- [x] 1.1 创建 QuickPaymentConfig Model
- [x] 1.2 创建 QuickPaymentRepo
- [x] 1.3 扩展 PaymentService 接口
- [x] 1.4 实现 GetQuickPayments 方法
- [x] 1.5 实现 SetQuickPayment 方法
- [x] 1.6 实现 RemoveQuickPayment 方法
- [x] 1.7 添加数据库迁移

## 待完成任务 ⏳

- [ ] 2.1 创建 QuickPaymentHandler
- [ ] 2.2 注册路由
- [ ] 2.3 更新 API 文档
- [ ] 3.1 编写单元测试
- [ ] 3.2 更新 shop_01.sql

## 🚀 下一步行动

1. 继续完成 Phase 2 的剩余任务
2. 下一个任务: 2.1 创建 QuickPaymentHandler
3. 预计剩余工作量: 2 SP
```

---

### 示例 2: 检查所有 Specs

**输入:** `/spec:check` 或 `/spec:check all`

**输出:**
```markdown
# 📊 Active Specs 进度总览

检查时间: 2026-01-13 15:30

## 汇总统计

| 指标 | 数值 |
|------|------|
| Active Specs | 3 |
| 可归档 | 1 |
| 开发中 | 2 |

## 详细进度

| Spec ID | 状态 | 总任务 | 已完成 | 完成率 | 进度条 |
|---------|------|--------|--------|--------|--------|
| story-pos-quick-payment | 🚧 开发中 | 15 | 7 | 47% | ██████░░░░░░ |
| task-shop-export | ✅ 可归档 | 5 | 5 | 100% | ████████████ |
| bug-order-calc | 🚧 开发中 | 3 | 2 | 67% | ████████░░░░ |

## 状态分布

- ✅ 可归档 (100%): 1 个
- 🚧 开发中 (<100%): 2 个
- ⏳ 待开始 (0%): 0 个

## 建议操作

1. **可归档的 Spec:**
   - `task-shop-export` - 考虑执行归档

2. **需要继续的 Spec:**
   - `story-pos-quick-payment` - 进度 47%
   - `bug-order-calc` - 进度 67%
```

---

## 错误处理

### Spec 不存在

```markdown
❌ Spec 不存在

未找到 Spec: {spec-id}
路径: docs/shared/specs/active/{spec-id}/

**可能原因:**
1. Spec ID 拼写错误
2. Spec 已归档或废弃
3. Spec 尚未创建

**建议操作:**
- 检查 Spec ID 拼写
- 查看已归档: `docs/shared/specs/archived/`
- 创建新 Spec: `/spec:create {spec-id}`
```

### tasks.md 不存在

```markdown
⚠️ tasks.md 不存在

Spec: {spec-id}
状态: requirements.md 存在，但 tasks.md 不存在

**说明:**
该 Spec 可能还未进入开发阶段。

**建议操作:**
1. 确认 requirements.md 审核状态
2. 如已通过，执行 `/spec:design {spec-id}` 创建设计方案
```

### 无 Active Specs

```markdown
📭 当前没有 Active Specs

目录 `docs/shared/specs/active/` 为空。

**建议操作:**
- 创建新 Spec: `/spec:create`
- 查看已归档: `docs/shared/specs/archived/`
```
