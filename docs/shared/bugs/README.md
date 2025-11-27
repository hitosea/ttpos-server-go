# Bug 管理体系

> 本目录用于管理项目中的 Bug 报告、跟踪和归档

---

## 📁 目录结构

```
docs/shared/bugs/
├── active/                          # 🔴 未解决的 Bug
│   └── bug-{YYMMDD}-{序号}-{module}-{brief}/
│       ├── bug.md                   # Bug 描述（问题详情）
│       ├── solution.md              # 修复方案（/bug-spec 后创建）
│       └── tasks.md                 # 任务分解（/bug-spec 后创建）
└── resolved/                        # 🟢 已解决的 Bug（按版本归档）
    └── {version}/                   # 版本目录（如 v2.10, v2.11）
        └── bug-{YYMMDD}-{序号}-{module}-{brief}/
            ├── bug.md               # Bug 描述（含归档标记）
            ├── solution.md          # 修复方案（含经验总结）
            └── tasks.md             # 任务分解（所有任务已完成）
```

---

## 🚀 快速开始

### 1. 创建 Bug 报告

```bash
/bug-create order-payment-timeout
/bug-create member-login-error --severity critical
```

详见：[`/bug-create` 指令](../../.cursor/commands/bug-create.md)

### 2. 创建修复方案和任务

```bash
/bug-spec @bug-251127-001
/bug-spec order-payment-timeout
```

详见：[`/bug-spec` 指令](../../.cursor/commands/bug-spec.md)

### 3. 归档已修复的 Bug

```bash
/bug-archive @bug-251127-001
/bug-archive order-payment-timeout --version v2.11
```

详见：[`/bug-archive` 指令](../../.cursor/commands/bug-archive.md)

---

## 📝 Bug ID 规则

### 格式

```
bug-{YYMMDD}-{序号}
```

### 示例

- `bug-251127-001` - 2025年11月27日发现的第1个 Bug
- `bug-251127-002` - 2025年11月27日发现的第2个 Bug
- `bug-251128-001` - 2025年11月28日发现的第1个 Bug

### 序号规则

- 每日从 `001` 开始
- 自动递增
- 不跨天累计

---

## 🏷️ Bug 状态

| 状态       | 图标 | 说明                     | 文件数量          | 下一步          |
| ---------- | ---- | ------------------------ | ----------------- | --------------- |
| **待分析** | 🔴   | Bug 刚创建，等待技术分析 | `bug.md`          | 调查分析        |
| **规划中** | 🟡   | 正在制定修复方案和任务   | `bug.md` + `solution.md` + `tasks.md` | 实施修复 |
| **已修复** | 🟢   | Bug 已修复并验证通过     | 完整文档集        | `/bug-archive`  |
| **已关闭** | ⚪   | 不需要修复或重复问题     | `bug.md`          | 直接归档        |

---

## 🎯 严重程度分级

| 级别       | 说明                           | 响应时间 | 示例                     |
| ---------- | ------------------------------ | -------- | ------------------------ |
| **critical** | 系统崩溃、数据丢失、严重安全问题 | 立即     | 订单数据丢失、支付失败   |
| **high**     | 核心功能无法使用               | 1 天内   | 收银系统无法结账         |
| **medium**   | 功能部分异常但有备用方案       | 3 天内   | 报表导出慢、界面显示异常 |
| **low**      | UI 问题、次要功能异常          | 1 周内   | 按钮样式错误、提示文案   |

---

## 📊 Bug 统计

### 按版本统计

```bash
# 查看某个版本修复的 Bug 数量
ls docs/shared/bugs/resolved/v2.10/ | wc -l

# 查看活跃 Bug 数量
ls docs/shared/bugs/active/ | wc -l
```

### 按模块统计

```bash
# 查看订单模块的 Bug
find docs/shared/bugs/ -name "*-order-*"

# 查看支付相关的 Bug
grep -r "payment" docs/shared/bugs/active/*/bug.md
```

### 按严重程度统计

```bash
# 查看高严重级别的活跃 Bug
grep -l "严重程度.*critical\|high" docs/shared/bugs/active/*/bug.md
```

---

## 🔗 Bug 生命周期

```
┌─────────────┐
│ 发现问题     │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ /bug-create │  创建 Bug 报告（bug.md）
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ 🔴 待分析    │  active/ 目录，只有 bug.md
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ 技术分析     │  调查定位、填写 bug.md
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ /bug-spec   │  创建修复方案（solution.md + tasks.md）
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ 🟡 规划中    │  active/ 目录，有完整文档
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ 实施修复     │  按 tasks.md 执行任务
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ 测试验证     │  单元测试、集成测试、手动测试
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ /bug-archive│  归档到 resolved/{version}/
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ 🟢 已修复    │  resolved/{version}/ 目录
└─────────────┘
       │
       ▼
┌─────────────┐
│ Graphiti    │  记录经验到知识图谱
└─────────────┘
```

---

## 🔍 查询 Bug

### 查询活跃 Bug

```bash
# 查看所有活跃 Bug
ls docs/shared/bugs/active/

# 搜索特定关键词
grep -r "payment" docs/shared/bugs/active/*/bug.md
```

### 查询已解决 Bug

```bash
# 查看 v2.10 版本修复的 Bug
ls docs/shared/bugs/resolved/v2.10/

# 查看所有已解决的 Bug
find docs/shared/bugs/resolved/ -name "bug.md"
```

### 在 Graphiti 中查询

```bash
# 使用 MCP 工具搜索
mcp_Graphiti_search_memory_facts --query "支付超时问题"
```

---

## 🔄 与 Spec 体系的对应关系

Bug 管理体系设计与 Spec 管理体系保持一致，采用三阶段管理：

| 阶段     | Bug 管理        | Spec 管理       | 输出产物              | 说明                 |
| -------- | --------------- | --------------- | --------------------- | -------------------- |
| **阶段 1** | `/bug-create`   | `/create-spec`  | `bug.md`              | 问题描述 vs 需求定义 |
| **阶段 2** | `/bug-spec`     | `/design-spec`  | `solution.md` + `tasks.md` | 修复方案 vs 技术设计 |
| **阶段 3** | `/bug-archive`  | `/archive-spec` | 归档到 `resolved/{version}/` | 归档已修复 vs 归档已完成 |

### 文档对应关系

| Bug 文档          | Spec 文档         | 作用                     |
| ----------------- | ----------------- | ------------------------ |
| `bug.md`          | `requirements.md` | 描述问题 vs 描述需求     |
| `solution.md`     | `design.md`       | 修复方案 vs 技术设计     |
| `tasks.md`        | `tasks.md`        | 任务分解（格式一致）     |
| `active/`         | `active/`         | 进行中的 Bug vs Spec     |
| `resolved/{ver}/` | `archived/{ver}/` | 已解决的 Bug vs 已完成的 Spec |

### 状态流转对应

| Bug 状态     | Spec 状态     | 说明                     |
| ------------ | ------------- | ------------------------ |
| 🔴 待分析    | 🔴 待审核     | 刚创建，等待审核         |
| 🟡 规划中    | 🟡 设计中     | 制定方案和任务           |
| 🟢 已修复    | 🟢 已完成     | 完成并验证通过           |
| 归档         | 归档          | 归档到版本目录           |

---

## 📚 Bug 文档模板

每个 Bug 包含以下核心信息：

### 基本信息

- Bug ID（唯一标识）
- 模块名称
- 严重程度
- 发现版本
- 状态

### 问题描述

- 现象
- 复现步骤
- 预期行为
- 实际行为

### 环境信息

- 部署环境
- 数据库版本
- 相关服务
- 客户端类型

### 影响范围

- 受影响的模块（Main/Admin/BMP）
- 受影响的终端（pos/shop/kds/kiosk/mobile）

### 技术分析

- 相关代码
- 错误日志
- 可能原因

### 解决方案

- 修复方案
- 验证记录
- 提交记录

---

## 🔗 相关资源

### 指令文档

- [`/bug-create`](../../.cursor/commands/bug-create.md) - 创建 Bug 报告（第一阶段）
- [`/bug-spec`](../../.cursor/commands/bug-spec.md) - 创建修复方案和任务（第二阶段）
- [`/bug-archive`](../../.cursor/commands/bug-archive.md) - 归档已修复的 Bug（第三阶段）

### 工作流文档

- [Bug 修复工作流](../agent/workflows/bug-fixing.md)
- [问题排查指南](./troubleshooting/)

### 相关规范

- [Git 提交规范](../../.cursor/rules/version.mdc)
- [知识管理规范](../../.cursor/rules/knowledge_management.mdc)

---

## 💡 最佳实践

### 1. 第一阶段：及时记录（/bug-create）

- 发现 Bug 后立即创建报告
- 记录详细的复现步骤
- 填写完整的环境信息
- 标注影响范围和严重程度

### 2. 第二阶段：充分规划（/bug-spec）

- 深入分析根本原因
- 对比多个修复方案
- 制定详细的测试计划
- 分解清晰的任务清单

### 3. 第三阶段：验证充分（/bug-archive）

- 编写单元测试覆盖修复逻辑
- 进行集成测试验证
- 手动测试复现步骤
- 确保所有任务完成

### 4. 持续改进

- 归档时填写经验总结
- 记录到 Graphiti 便于查询
- 更新相关文档避免重复
- 分析 Bug 模式，预防类似问题

### 5. 关联管理

- 关联相关的 Spec 和 Proposal
- 提交代码时引用 Bug ID
- 更新相关文档
- 通知相关人员

---

## 📈 版本管理

### 版本号格式

- 使用 `vX.X` 格式（major.minor）
- 例如：`v2.10`, `v2.11`
- 不包含 patch 版本号

### 归档策略

- Bug 修复后归档到解决版本目录
- 同一版本的 Bug 放在同一目录下
- 保持目录名称一致性

### 查询历史

```bash
# 查看 v2.10 版本修复了哪些 Bug
ls docs/shared/bugs/resolved/v2.10/

# 统计各版本修复的 Bug 数量
for v in docs/shared/bugs/resolved/*/; do
  echo "$(basename $v): $(ls $v | wc -l)"
done
```

---

## 🛠️ 维护指南

### 定期清理

- 每月审查活跃 Bug
- 关闭无效或重复的 Bug
- 整理归档的 Bug 文档

### 统计分析

- 分析高频 Bug 模块
- 识别系统性问题
- 改进开发流程

### 知识沉淀

- 典型 Bug 记录到 Graphiti
- 更新故障排查指南
- 编写预防措施文档

---

## 🆘 需要帮助？

1. **创建 Bug** - 参考 [`/create-bug` 指令文档](../../.cursor/commands/create-bug.md)
2. **解决 Bug** - 参考 [`/resolve-bug` 指令文档](../../.cursor/commands/resolve-bug.md)
3. **查询经验** - 使用 Graphiti 搜索相似问题
4. **工作流程** - 查看 [Bug 修复工作流](../agent/workflows/bug-fixing.md)

---

**最后更新**: 2025-11-27  
**维护者**: TTPOS Team  
**版本**: v1.0.0

