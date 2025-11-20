# 需求管理工作流（后端版）

## 概述

本工作流基于 **Scrum** 框架，定义了从需求提案到 Spec 创建的完整流程。

---

## 触发条件

- 产品经理提出新功能想法
- 用户反馈功能改进建议
- 技术团队提出架构优化方案

---

## 前置条件

### 人员要求

- **产品经理 (PM)**: 负责需求提案和业务价值评估
- **技术负责人 (Tech Lead)**: 负责技术可行性评估
- **开发代表 (Dev)**: 参与技术方案讨论

### 环境要求

- 熟悉 Scrum 框架和 User Story 编写规范
- 了解项目的 Story Point 评估标准
- 掌握项目的 Spec 命名规范

---

## 执行流程概览

```
想法 → /propose → 填写提案 → 需求评审 →
  ├─ 批准 → /create-spec → 填写 requirements/design/tasks → SP评估 →
  │   ├─ SP ≤ 5 → 进 Sprint → 开发
  │   └─ SP > 5 → 拆分 Spec → 重新评估
  └─ 拒绝 → 归档
```

---

## 详细步骤

### Step 1: 创建需求提案

**操作步骤**:

```bash
# 创建提案（例如：快速支付功能）
/propose quick-payment
```

**输出产物**: `docs/team/proposals/{YYYY-MM-DD}-quick-payment.md`

**检查清单**:

- [ ] 已创建提案文档
- [ ] 文件路径符合命名规范

---

### Step 2: 填写提案内容

1. **描述问题背景**

   - 当前存在什么问题？
   - 用户遇到了什么痛点？

2. **阐明业务价值**

   - 解决这个问题能带来什么收益？
   - 对业务的影响是什么？

3. **提出解决方案**

   - 简要描述解决方案
   - 列出核心功能点（3-5 个）
   - 标注影响范围（Go/PHP/Vue）

4. **初步评估**
   - 技术复杂度（低/中/高）
   - 工作量预估
   - 潜在风险识别

**参考模板**: `docs/agent/templates/proposal-template.md`

---

### Step 3: 组织需求评审会议

**会议议程**:

1. PM 陈述 (10 分钟)
2. 技术评估 (15 分钟)
3. 团队讨论 (10 分钟)
4. 投票表决 (5 分钟)

---

### Step 4: 评审决策

#### 场景 A: ✅ 批准

- 进入 Step 5（创建 Spec）

#### 场景 B: 🔄 修改后批准

- PM 修改提案，重新评审

#### 场景 C: ❌ 拒绝

- 归档提案，标注拒绝原因

---

### Step 5: 创建 Spec 文档

```bash
# 创建 Spec
/create-spec story-order-quick-payment
```

**输出产物**: `docs/shared/specs/story-order-quick-payment/`

- `requirements.md` - 需求规格
- `design.md` - 技术设计
- `tasks.md` - 任务分解

**智能关联**:

- 自动搜索对应的 Proposal
- 自动填充 Proposal 链接
- 回写 Proposal，更新状态

---

### Step 6: 编写 User Story + AC

**User Story 格式**:

```markdown
## 📝 用户故事

**作为** 收银员  
**我想** 一键完成支付并打印小票  
**以便于** 提高收银效率
```

**验收标准 (AC)**:

```markdown
1. **WHEN** 订单金额 > 0 **THEN** 系统 **SHALL** 显示"快速支付"按钮
2. **WHEN** 点击"快速支付"按钮 **THEN** 系统 **SHALL** 自动完成支付
```

---

### Step 7: 技术方案设计

**包含内容**:

1. **架构设计**

   - 组件结构
   - 数据流
   - 技术栈选择（Go/PHP/Vue）

2. **代码复用分析**

   - 可复用的现有代码
   - 需要新建的模块

3. **技术实现细节**

   - 数据模型
   - API 接口设计
   - 数据库表设计

4. **后端特有考虑**
   - [ ] 数据库迁移计划
   - [ ] API 响应格式
   - [ ] 是否需要 gRPC 服务

**参考模板**: `docs/agent/templates/design-template.md`

---

### Step 8: Story Point 评估

**评估流程**:

1. 技术复杂度评估
2. 功能复杂度评估
3. 因素加成
4. 团队评审

**决策点**:

- **SP ≤ 5**: 进入 Step 9
- **SP > 5**: 必须拆分需求

**后端评估考虑**:

- Go 开发工作量
- PHP 开发工作量
- 数据库设计工作量
- 第三方集成工作量

参考: `.cursor/rules/scrum_story_point.mdc`

---

### Step 9: 任务分解

**任务模板**:

```markdown
## Phase 1: 核心功能实现

- [ ] 1.1 创建数据库表

  - File: admin/database/migrations/{timestamp}\_add_order_table.php
  - Purpose: 创建订单表
  - Requirements: 1.1
  - Language: PHP

- [ ] 1.2 创建 Go Model

  - File: main/app/model/order.go
  - Purpose: 定义订单模型
  - Requirements: 1.1
  - Language: Go

- [ ] 1.3 实现订单 Service
  - File: main/app/service/order_service.go
  - Purpose: 实现订单业务逻辑
  - Requirements: 1.2, 1.3
  - Language: Go
```

**参考模板**: `docs/agent/templates/tasks-template.md`

---

### Step 10: 进入 Sprint Backlog

1. Sprint Planning 会议
2. 确定优先级
3. 分配责任人
4. 确定目标 Sprint
5. 进入功能开发工作流

参考: `docs/agent/workflows/feature-development.md`

---

## 常见问题

### Q: 提案被拒绝后怎么办？

**A**: 记录拒绝原因，未来可以重新评审。

### Q: 如何判断需求是否需要拆分？

**A**: SP > 5 必须拆分，通常按模块拆分（订单、支付、会员等）。

### Q: 技术方案设计需要多详细？

**A**:

- SP1-3: 简要描述即可
- SP5: 需要详细的架构和数据模型
- 必须明确技术栈（Go/PHP/Vue）

---

## 工作流输出产物

| 阶段     | 产物                             | 存放位置                                      |
| -------- | -------------------------------- | --------------------------------------------- |
| 需求提案 | `{YYYY-MM-DD}-{feature-name}.md` | `docs/team/proposals/`                        |
| 需求规格 | `requirements.md`                | `docs/shared/specs/story-{module}-{feature}/` |
| 技术设计 | `design.md`                      | `docs/shared/specs/story-{module}-{feature}/` |
| 任务分解 | `tasks.md`                       | `docs/shared/specs/story-{module}-{feature}/` |

---

## 相关资源

### Agent 规范

- `AGENTS.md` - Agent 速查表
- `.cursor/rules/specs.mdc` - Spec 命名规范
- `.cursor/rules/scrum_story_point.mdc` - SP 评估规范

### 模板文件

- `docs/agent/templates/proposal-template.md`
- `docs/agent/templates/requirements-template.md`
- `docs/agent/templates/design-template.md`
- `docs/agent/templates/tasks-template.md`

### 相关工作流

- `docs/agent/workflows/feature-development.md` - 功能开发
- `docs/agent/workflows/database-migration.md` - 数据库迁移

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 动作：需求完成评审/拆分后，如发现可复用经验或决策沉淀，立即创建 Episode，并在 Proposal/Spec 文末互链。

---

**最后更新**: 2025-11-16  
**维护者**: 产品组 + 后端开发组
