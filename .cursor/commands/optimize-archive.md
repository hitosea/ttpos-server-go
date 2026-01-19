---
name: optimize-archive
description: 归档已完成的优化到指定版本目录
---

# /optimize-archive - 归档优化

## 使用场景

将已完成并验证通过的优化归档到对应版本目录。

> **前置条件**: 优化必须已完成所有任务并通过验证，达到预期收益。

## 使用方式

```bash
/optimize-archive @opt-251201-001                        # 自动检测版本号
/optimize-archive @opt-251201-001 --version v2.12        # 指定发布版本号
/optimize-archive order-query-performance                 # 支持用简述搜索
```

## 参数

- `opt_id_or_brief`: 必填，优化 ID（如 `opt-251201-001`）或优化简述（如 `order-query-performance`）
  - 支持 `@` 前缀
  - 支持部分匹配搜索
- `--version`: 可选，发布版本号（格式: `vX.X`，只到 minor）

## 版本号获取优先级

1. 命令参数 `--version` 显式指定
2. 从 `main/version/version.go` 的 `Version` 变量提取（只取 major.minor，如 `2.12.5` → `v2.12`）
3. 交互询问用户

## 执行流程

### Step 1: 查找优化

```yaml
IF 参数是完整优化 ID THEN
  查找: docs/shared/opts/active/opt-{id}-*/
ELSE IF 参数是简述 THEN
  搜索: docs/shared/opts/active/*/*-{brief}*/
  IF 找到多个 THEN
    显示列表让用户选择
  END IF
END IF

IF 未找到 THEN
  报错并退出
END IF
```

### Step 2: 验证必填信息

检查优化目录是否包含：

- ✅ `optimize.md` - 优化需求
- ✅ `solution.md` - 优化方案（必填）
- ✅ `tasks.md` - 任务清单

**检查 solution.md 完整性**：
- ✅ 问题分析已填写
- ✅ 优化方案已填写
- ✅ 收益评估已填写
- ✅ 测试计划已填写

如果缺失，**阻止归档**。

### Step 3: 检查任务完成度

- 读取 `tasks.md`，统计任务完成率
- **如果有未完成任务，阻止归档**
- 输出任务完成统计

### Step 4: 验证收益达标

```yaml
IF 优化类型是 performance THEN
  检查 solution.md 中的性能指标:
    - 响应时间是否达标
    - 吞吐量是否达标
    - 资源占用是否达标
  IF 未达标 THEN
    警告并询问是否继续
  END IF
END IF

IF 优化类型是 ux THEN
  检查用户反馈或体验指标
END IF
```

### Step 5: 确定发布版本号

- 按优先级获取版本号
- 版本号格式校验（必须为 `vX.X`）

### Step 6: 更新 optimize.md

在原文件中更新：

```markdown
| 状态       | 🔵 已完成           |
| 发布版本   | v{version}          |
| 完成日期   | {YYYY-MM-DD}        |
| 完成者     | {git user.name}     |
```

### Step 7: 移动到已完成目录

```
docs/shared/opts/active/opt-{id}-{module}-{brief}/
    ↓
docs/shared/opts/completed/{version}/opt-{id}-{module}-{brief}/
```

- 如果版本目录不存在，自动创建
- 保持目录名称不变
- 移动整个目录（包含所有文件）

### Step 8: 添加归档标记

在 `optimize.md` 头部添加：

```markdown
> ✅ **已完成** - 此优化已在 {version} 中发布。
>
> - 完成时间: {YYYY-MM-DD}
> - 完成者: {git config user.name}
> - 验证状态: ✅ 已验证
> - 收益达成: ✅ 达到预期
```

### Step 9: 生成收益报告

从 `solution.md` 中提取收益数据，在 `optimize.md` 末尾添加：

```markdown
## 收益总结

**优化类型**: {category}
**实施周期**: {开始日期} ~ {完成日期} ({X}天)

### 性能提升（如适用）

| 指标       | 优化前 | 优化后 | 提升   |
| ---------- | ------ | ------ | ------ |
| 响应时间   | {val}  | {val}  | {X}%   |
| 吞吐量     | {val}  | {val}  | {X}%   |
| 资源占用   | {val}  | {val}  | -{X}%  |

### 体验改善（如适用）

- **操作步骤**: 从 {X} 步减少到 {Y} 步
- **用户满意度**: 从 {X} 提升到 {Y}
- **错误率**: 从 {X}% 降低到 {Y}%

### 成本节约（如适用）

- **服务器成本**: 节约 {X}元/月
- **维护成本**: 减少 {X}小时/月

## 经验总结

**优化方法**: {一句话总结}
**关键技术**: {技术要点}
**注意事项**: {实施要点}
**适用场景**: {何时应用此优化}
**参考资料**: {相关文档}
```

### Step 10: 创建 Graphiti Episode

自动创建 Graphiti 记录，便于未来查询：

```json
{
  "name": "Opt-{id}: {简短描述}",
  "episode_body": {
    "opt_id": "opt-251201-001",
    "module": "order",
    "category": "performance",
    "priority": "high",
    "start_version": "v2.11",
    "completed_version": "v2.12",
    "description": "...",
    "problem_analysis": "...",
    "solution": "...",
    "benefits": {
      "response_time_improvement": "50%",
      "throughput_improvement": "80%",
      "resource_reduction": "30%"
    },
    "related_files": [...],
    "lessons_learned": "...",
    "applicable_scenarios": "..."
  },
  "source": "json",
  "source_description": "性能优化记录"
}
```

### Step 11: 更新关联资源

```yaml
IF optimize.md 中有关联 Spec THEN
  在 Spec 的 tasks.md 中更新相关任务状态
END IF

IF optimize.md 中有关联 Proposal THEN
  在 Proposal 中记录优化完成情况
END IF
```

### Step 12: 生成版本报告条目

在 `docs/shared/opts/completed/{version}/README.md` 中添加条目：

```markdown
## {version} 版本优化汇总

### Opt-{id}: {简短描述}

- **模块**: {module}
- **类型**: {category}
- **收益**: {关键指标提升}
- **详情**: [查看详情](./opt-{id}-{module}-{brief}/)
```

### Step 13: 记录活动日志

按 `activity_log.mdc` 规范记录：

```
| HH:mm | /optimize-archive | opt-{id}-{brief} | ✅ | 归档优化到v{version} |
```

## 输出示例

```
✅ 优化已完成并归档

🚀 opt-251201-001: order-query-performance
   从: docs/shared/opts/active/opt-251201-001-order-query-performance/
   到: docs/shared/opts/completed/v2.12/opt-251201-001-order-query-performance/

📊 收益达成:
   - 响应时间: 500ms → 200ms (↓60%)
   - 吞吐量: 1000 QPS → 1800 QPS (↑80%)
   - CPU 占用: 70% → 45% (↓36%)

📋 任务完成: 15/15 (100%)

💰 成本节约:
   - 服务器成本: 节约 ¥2000/月
   - 维护成本: 减少 10 小时/月

🔗 已更新资源:
   - Spec: docs/shared/specs/active/story-order-manage/
   - Graphiti: ✅ 已记录经验
   - 版本报告: ✅ 已添加条目

📝 下一步:
   1. 提交代码并关联优化 ID
   2. 通知相关人员
   3. 持续监控效果
   4. 考虑推广到其他模块
```

## 错误处理

| 错误类型             | 处理方式                                     |
| -------------------- | -------------------------------------------- |
| 优化不存在           | 报错：优化不存在于 active/ 目录              |
| solution.md 不存在   | **阻止归档**：必须先使用 `/optimize-spec` 创建 |
| solution.md 不完整   | **阻止归档**：必须填写完整优化方案           |
| 任务未完成           | **阻止归档**：显示未完成任务列表             |
| 收益未达标           | **警告**：询问是否继续归档                   |
| 版本号格式错误       | 报错：版本号必须为 vX.X 格式                 |
| 找到多个匹配优化     | 显示列表，让用户选择                         |

## 前置条件

- 优化必须在 `active/` 目录中
- `solution.md` 必须存在且完整
- `tasks.md` 中所有任务必须完成（`[x]`）
- 收益指标建议达到预期（可豁免但需确认）
- 建议已提交代码并通过代码审查

## 后置操作建议

1. **代码提交**

```bash
git commit -m "perf(order): 优化订单查询性能 (#opt-251201-001)"
```

2. **更新相关文档**

- 如果是架构优化，更新架构文档
- 如果是性能优化，更新性能优化指南
- 更新最佳实践文档

3. **推广优化经验**

- 团队分享会
- 编写博客文章
- 更新开发规范

4. **持续监控**

- 关注相关监控指标
- 观察是否有副作用
- 收集用户反馈

5. **考虑推广**

- 评估是否适用其他模块
- 制定推广计划
- 记录推广效果

## 智能功能

### 1. 自动生成收益报告

从优化文档中提取关键信息，生成结构化的收益报告：

- 性能指标对比
- 成本节约计算
- 用户体验改善
- 业务价值评估

### 2. 检查相似优化

在 Graphiti 中搜索相似的已完成优化，对比效果。

### 3. 生成版本汇总

自动生成本版本的优化汇总报告：

- 本版本优化数量
- 按模块统计
- 按类型统计
- 总体收益统计

### 4. 推广建议

根据优化效果和适用性，智能推荐推广范围：

- 如果效果显著且通用 → 建议推广到所有模块
- 如果适用特定场景 → 建议在类似场景应用
- 如果有技术风险 → 建议谨慎推广

## 工作流位置

```
发现问题 → /optimize-create → 收益评估 → /optimize-spec → 实施优化 → /optimize-archive
                                                                         ↑
                                                                     当前命令
```

## 相关命令

| 命令                | 用途                           |
| ------------------- | ------------------------------ |
| `/optimize-create`  | 创建优化需求                   |
| `/optimize-spec`    | 创建优化方案和任务             |
| `/optimize-archive` | 归档已完成的优化（当前命令）   |

## 状态流转

```
🟡 待评估 → 🟢 规划中 → 🔵 已完成 → 归档到 completed/{version}/
                             ↓
                         ⚪ 已取消 (收益不足/优先级降低)
```

## 版本管理策略

### 版本号规则

- 使用 `vX.X` 格式（只到 minor 版本）
- 例如：`v2.11`, `v2.12`
- 不包含 patch 版本号

### 归档结构

```
docs/shared/opts/
├── active/                          # 进行中的优化
│   ├── opt-251201-001-order-query/
│   │   ├── optimize.md
│   │   ├── solution.md
│   │   └── tasks.md
│   └── opt-251201-002-ui-loading/
│       ├── optimize.md
│       ├── solution.md
│       └── tasks.md
└── completed/                       # 已完成的优化（按版本）
    ├── v2.11/
    │   ├── README.md                # 版本汇总
    │   ├── opt-251120-001-db-index/
    │   │   ├── optimize.md
    │   │   ├── solution.md
    │   │   └── tasks.md
    │   └── opt-251125-001-cache/
    │       ├── optimize.md
    │       ├── solution.md
    │       └── tasks.md
    └── v2.12/
        ├── README.md
        └── opt-251201-001-order-query/
            ├── optimize.md
            ├── solution.md
            └── tasks.md
```

### 查询历史优化

```bash
# 查看某个版本的所有优化
ls docs/shared/opts/completed/v2.12/

# 搜索特定模块的优化
find docs/shared/opts/completed/ -name "*-order-*"

# 在 Graphiti 中查询
# 使用 MCP 工具搜索优化相关经验
```

## 集成 Graphiti

### Episode 内容结构

```json
{
  "name": "Opt-{id}: {简短描述}",
  "episode_body": {
    "opt_id": "opt-251201-001",
    "module": "order",
    "category": "performance",
    "priority": "high",
    "start_version": "v2.11",
    "completed_version": "v2.12",
    "description": "订单查询性能优化",
    "problem_analysis": "订单表数据量大，缺少索引，查询慢",
    "solution": "添加复合索引，优化查询语句，引入 Redis 缓存",
    "benefits": {
      "response_time": {
        "before": "500ms",
        "after": "200ms",
        "improvement": "60%"
      },
      "throughput": {
        "before": "1000 QPS",
        "after": "1800 QPS",
        "improvement": "80%"
      },
      "resource": {
        "cpu_before": "70%",
        "cpu_after": "45%",
        "reduction": "36%"
      }
    },
    "related_files": [
      "main/app/service/order_manage.go",
      "admin/database/migrations/20251201_add_order_index.php"
    ],
    "lessons_learned": "对于大表查询，索引优化 + 缓存是最有效的组合",
    "applicable_scenarios": "适用于所有大表查询优化场景"
  },
  "source": "json",
  "source_description": "性能优化记录"
}
```

### 标签建议

- `optimization` - 优化记录
- `{category}` - 优化类型（performance, ux, security, maintainability, scalability）
- `{module}` - 模块名称（order, member, product...）
- 技术标签：
  - `database-optimization` - 数据库优化
  - `cache-optimization` - 缓存优化
  - `api-optimization` - API 优化
  - `frontend-optimization` - 前端优化
  - `architecture-refactor` - 架构重构
  - `code-refactor` - 代码重构

## 与 Bug/Spec 体系的对应关系

| 优化归档             | Bug 归档          | Spec 归档         | 说明           |
| -------------------- | ----------------- | ----------------- | -------------- |
| 检查 tasks.md        | 检查 tasks.md     | 检查 tasks.md     | 确保任务完成   |
| 检查 solution.md     | 检查 solution.md  | 检查 design.md    | 确保方案完整   |
| **验证收益达标**     | 验证功能正确      | 验证功能实现      | 优化特有检查   |
| 移动到 completed/    | 移动到 resolved/  | 移动到 archived/  | 归档到版本目录 |
| 添加归档标记         | 添加归档标记      | 添加归档标记      | 标注归档信息   |
| **生成收益报告**     | 生成经验总结      | 生成实施总结      | 优化特有输出   |
| 创建 Graphiti        | 创建 Graphiti     | 创建 Graphiti     | 记录经验       |
| 更新关联资源         | 更新关联资源      | 更新关联资源      | 更新 Spec/Proposal |

---

**版本**: v1.0.0  
**创建日期**: 2025-12-01  
**维护者**: 知识管理组  
**状态**: ✅ MVP

