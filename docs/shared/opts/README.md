# 优化管理体系

> 本目录用于管理项目中的性能优化、体验改进、技术债务等优化类任务的规划、跟踪和归档

---

## 📁 目录结构

```
docs/shared/opts/
├── active/                          # 🟡 进行中的优化
│   └── opt-{YYMMDD}-{序号}-{module}-{brief}/
│       ├── optimize.md              # 优化需求（需求详情）
│       ├── solution.md              # 优化方案（/optimize-spec 后创建）
│       └── tasks.md                 # 任务分解（/optimize-spec 后创建）
└── completed/                       # 🔵 已完成的优化（按版本归档）
    └── {version}/                   # 版本目录（如 v2.10, v2.11）
        ├── README.md                # 版本优化汇总
        └── opt-{YYMMDD}-{序号}-{module}-{brief}/
            ├── optimize.md          # 优化需求（含归档标记和收益报告）
            ├── solution.md          # 优化方案（含经验总结）
            └── tasks.md             # 任务分解（所有任务已完成）
```

---

## 🚀 快速开始

### 1. 创建优化需求

```bash
/optimize-create order-query-performance
/optimize-create api-response-slow --priority high --category performance
/optimize-create ui-loading-experience --category ux
```

详见：[`/optimize-create` 指令](../../.cursor/commands/optimize-create.md)

### 2. 创建优化方案和任务

```bash
/optimize-spec @opt-251201-001
/optimize-spec order-query-performance
```

详见：[`/optimize-spec` 指令](../../.cursor/commands/optimize-spec.md)

### 3. 归档已完成的优化

```bash
/optimize-archive @opt-251201-001
/optimize-archive order-query-performance --version v2.12
```

详见：[`/optimize-archive` 指令](../../.cursor/commands/optimize-archive.md)

---

## 📝 优化 ID 规则

### 格式

```
opt-{YYMMDD}-{序号}
```

### 示例

- `opt-251201-001` - 2025年12月1日创建的第1个优化
- `opt-251201-002` - 2025年12月1日创建的第2个优化
- `opt-251202-001` - 2025年12月2日创建的第1个优化

### 序号规则

- 每日从 `001` 开始
- 自动递增
- 不跨天累计

---

## 🏷️ 优化状态

| 状态         | 图标 | 说明                     | 文件数量          | 下一步            |
| ------------ | ---- | ------------------------ | ----------------- | ----------------- |
| **待评估**   | 🟡   | 优化刚创建，等待技术评估 | `optimize.md`     | 分析收益和成本    |
| **规划中**   | 🟢   | 正在制定优化方案和任务   | `optimize.md` + `solution.md` + `tasks.md` | 实施优化 |
| **已完成**   | 🔵   | 优化已实施并验证通过     | 完整文档集        | `/optimize-archive` |
| **已取消**   | ⚪   | 收益不足或优先级降低     | `optimize.md`     | 直接归档          |

---

## 🎯 优化类型

| 类型                | 说明             | 典型场景                     | 评估指标               |
| ------------------- | ---------------- | ---------------------------- | ---------------------- |
| **performance**     | 性能优化         | 响应慢、查询慢、加载慢       | 响应时间、吞吐量、资源占用 |
| **ux**              | 用户体验优化     | 交互不便、流程繁琐           | 操作步骤、用户满意度、错误率 |
| **security**        | 安全加固         | 漏洞修复、权限完善           | 安全评分、漏洞数量     |
| **maintainability** | 可维护性优化     | 代码重构、架构优化           | 代码质量、技术债务     |
| **scalability**     | 可扩展性优化     | 集群支持、容量规划           | 并发能力、扩展性       |

---

## 📊 优先级分级

| 级别       | 说明                           | 实施时间 | 示例                     |
| ---------- | ------------------------------ | -------- | ------------------------ |
| **critical** | 严重影响业务或用户体验         | 立即     | 系统崩溃、严重性能问题   |
| **high**     | 明显影响业务效率或用户体验     | 本周内   | API 响应慢、关键流程繁琐 |
| **medium**   | 改善用户体验或降低维护成本     | 本月内   | 查询优化、代码重构       |
| **low**      | 小优化，不影响核心功能         | 下版本   | UI 微调、代码美化        |

---

## 📈 优化统计

### 按版本统计

```bash
# 查看某个版本完成的优化数量
ls docs/shared/opts/completed/v2.12/ | wc -l

# 查看进行中的优化数量
ls docs/shared/opts/active/ | wc -l
```

### 按模块统计

```bash
# 查看订单模块的优化
find docs/shared/opts/ -name "*-order-*"

# 查看性能相关的优化
grep -r "performance" docs/shared/opts/active/*/optimize.md
```

### 按类型统计

```bash
# 查看性能优化
grep -l "优化类型.*performance" docs/shared/opts/active/*/optimize.md

# 查看用户体验优化
grep -l "优化类型.*ux" docs/shared/opts/active/*/optimize.md
```

### 收益统计

```bash
# 查看所有性能提升数据
grep -r "提升.*%" docs/shared/opts/completed/*/optimize.md

# 查看成本节约
grep -r "节约" docs/shared/opts/completed/*/optimize.md
```

---

## 🔄 优化生命周期

```
┌─────────────┐
│ 发现问题     │  监控数据、用户反馈、技术债务
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ /optimize-create │  创建优化需求（optimize.md）
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ 🟡 待评估    │  active/ 目录，只有 optimize.md
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ 收益评估     │  分析成本和收益，决定是否继续
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ /optimize-spec │  创建优化方案（solution.md + tasks.md）
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ 🟢 规划中    │  active/ 目录，有完整文档
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ 实施优化     │  按 tasks.md 执行任务
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ 测试验证     │  性能测试、功能测试、灰度发布
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ 收益验证     │  对比优化前后的性能指标
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ /optimize-archive │  归档到 completed/{version}/
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ 🔵 已完成    │  completed/{version}/ 目录
└─────────────┘
       │
       ▼
┌─────────────┐
│ Graphiti    │  记录优化经验到知识图谱
└─────────────┘
```

---

## 🔍 查询优化

### 查询进行中的优化

```bash
# 查看所有进行中的优化
ls docs/shared/opts/active/

# 搜索特定关键词
grep -r "performance" docs/shared/opts/active/*/optimize.md
```

### 查询已完成的优化

```bash
# 查看 v2.12 版本完成的优化
ls docs/shared/opts/completed/v2.12/

# 查看所有已完成的优化
find docs/shared/opts/completed/ -name "optimize.md"
```

### 在 Graphiti 中查询

```bash
# 使用 MCP 工具搜索
mcp_Graphiti_search_memory_facts --query "性能优化 查询优化"
```

---

## 🔗 与 Bug/Spec 体系的对应关系

优化管理体系与 Bug/Spec 管理体系保持一致，采用三阶段管理：

| 阶段       | 优化管理          | Bug 管理        | Spec 管理       | 输出产物              | 说明                 |
| ---------- | ----------------- | --------------- | --------------- | --------------------- | -------------------- |
| **阶段 1** | `/optimize-create` | `/bug-create`   | `/spec-create`  | `optimize.md`         | 需求定义             |
| **阶段 2** | `/optimize-spec`   | `/bug-spec`     | `/spec-design`  | `solution.md` + `tasks.md` | 方案设计和任务分解 |
| **阶段 3** | `/optimize-archive` | `/bug-archive` | `/spec-archive` | 归档到 `completed/{version}/` | 归档并记录经验 |

### 文档对应关系

| 优化文档          | Bug 文档          | Spec 文档         | 作用                     |
| ----------------- | ----------------- | ----------------- | ------------------------ |
| `optimize.md`     | `bug.md`          | `requirements.md` | 描述需求和问题           |
| `solution.md`     | `solution.md`     | `design.md`       | 优化方案和技术设计       |
| `tasks.md`        | `tasks.md`        | `tasks.md`        | 任务分解（格式一致）     |
| `active/`         | `active/`         | `active/`         | 进行中的工作             |
| `completed/{ver}` | `resolved/{ver}/` | `archived/{ver}/` | 已完成的工作（按版本）   |

### 核心区别

| 特性       | 优化管理         | Bug 管理        | Spec 管理       |
| ---------- | ---------------- | --------------- | --------------- |
| **触发方式** | 主动改进         | 被动响应问题    | 产品需求驱动    |
| **关注点**   | 收益和成本 (ROI) | 功能正确性      | 功能实现        |
| **优先级**   | 可延迟实施       | 必须尽快修复    | 按规划实施      |
| **验收标准** | 性能/体验指标    | 功能正常        | 需求满足        |
| **收益验证** | 必须验证收益     | 验证修复效果    | 验证功能        |

---

## 📚 优化文档模板

每个优化包含以下核心信息：

### 基本信息

- 优化 ID（唯一标识）
- 模块名称
- 优化类型（performance/ux/security/maintainability/scalability）
- 优先级
- 当前版本
- 状态

### 优化需求

- 当前问题
- 性能指标（如适用）
- 影响面（终端、用户、业务价值）
- 触发原因

### 初步分析

- 可能原因
- 优化方向
- 预估收益

### 优化方案（solution.md）

- 问题分析（性能瓶颈/用户痛点/技术债务）
- 方案对比（多个方案比较）
- 收益评估（性能提升/体验改善/成本节约）
- 影响分析（兼容性/风险/回滚方案）
- 测试计划（性能测试/功能测试/灰度发布）
- 上线计划（发布时间/监控指标/应急预案）

### 任务分解（tasks.md）

- 前期准备（基线测试/环境准备）
- 代码优化
- 数据库优化（如适用）
- 缓存优化（如适用）
- 测试验证（性能测试/功能回归/灰度验证）
- 文档更新
- 部署上线

### 收益报告（归档后）

- 性能提升（响应时间/吞吐量/资源占用）
- 体验改善（操作步骤/用户满意度/错误率）
- 成本节约（服务器成本/维护成本）
- 经验总结（优化方法/关键技术/注意事项）

---

## 🔗 相关资源

### 指令文档

- [`/optimize-create`](../../.cursor/commands/optimize-create.md) - 创建优化需求（第一阶段）
- [`/optimize-spec`](../../.cursor/commands/optimize-spec.md) - 创建优化方案和任务（第二阶段）
- [`/optimize-archive`](../../.cursor/commands/optimize-archive.md) - 归档已完成的优化（第三阶段）

### 工作流文档

- 优化实施工作流（待创建）
- 性能测试指南（待创建）

### 相关规范

- [Git 提交规范](../../.cursor/rules/version.mdc)
- [知识管理规范](../../.cursor/rules/knowledge_management.mdc)

---

## 💡 最佳实践

### 1. 第一阶段：充分评估（/optimize-create）

- 明确当前问题和优化目标
- 记录现有性能指标作为基线
- 评估优化的业务价值
- 标注影响范围和优先级
- 搜索 Graphiti 查找相似优化经验

### 2. 第二阶段：科学规划（/optimize-spec）

- 深入分析性能瓶颈或用户痛点
- 对比多个优化方案的成本和收益
- 制定详细的性能测试计划
- 设计灰度发布策略降低风险
- 分解清晰的任务清单

### 3. 第三阶段：验证收益（/optimize-archive）

- 进行充分的性能测试
- 对比优化前后的性能指标
- 验证收益是否达到预期
- 记录详细的收益报告
- 总结优化经验便于推广

### 4. 持续改进

- 归档时填写详细的经验总结
- 记录到 Graphiti 便于未来查询
- 评估是否可以推广到其他模块
- 更新相关文档和最佳实践
- 分享优化经验给团队

### 5. 关联管理

- 关联相关的 Spec 和 Proposal
- 提交代码时引用优化 ID
- 更新相关文档（架构文档/API 文档）
- 通知相关人员
- 持续监控优化效果

---

## 📊 收益衡量

### 性能优化

```yaml
关键指标:
  - 响应时间: 平均响应时间、P95、P99
  - 吞吐量: QPS、TPS
  - 资源占用: CPU、内存、磁盘IO、网络IO
  - 错误率: 错误数量、成功率

目标示例:
  - 响应时间降低 50%
  - 吞吐量提升 80%
  - CPU 占用降低 30%
```

### 用户体验优化

```yaml
关键指标:
  - 操作步骤: 减少点击次数、流程简化
  - 加载时间: 首屏时间、白屏时间
  - 错误率: 用户操作错误率
  - 满意度: 用户反馈、评分

目标示例:
  - 操作步骤从 5 步减少到 3 步
  - 页面加载时间从 3s 降低到 1s
  - 操作错误率降低 50%
```

### 可维护性优化

```yaml
关键指标:
  - 代码质量: 复杂度、重复率、测试覆盖率
  - 技术债务: 债务数量、债务等级
  - 维护成本: 修改时间、Bug 数量

目标示例:
  - 代码复杂度降低 40%
  - 测试覆盖率提升到 80%
  - 维护时间减少 2 小时/周
```

### 成本节约

```yaml
关键指标:
  - 服务器成本: 资源使用量、费用
  - 开发成本: 开发时间、人力成本
  - 运维成本: 运维时间、故障率

目标示例:
  - 服务器成本节约 ¥2000/月
  - 开发效率提升 20%
  - 故障率降低 50%
```

---

## 📈 版本管理

### 版本号格式

- 使用 `vX.X` 格式（major.minor）
- 例如：`v2.11`, `v2.12`
- 不包含 patch 版本号

### 归档策略

- 优化完成后归档到发布版本目录
- 同一版本的优化放在同一目录下
- 生成版本优化汇总（README.md）
- 保持目录名称一致性

### 查询历史

```bash
# 查看 v2.12 版本完成了哪些优化
ls docs/shared/opts/completed/v2.12/

# 统计各版本完成的优化数量
for v in docs/shared/opts/completed/*/; do
  echo "$(basename $v): $(ls $v | wc -l)"
done

# 查看某个版本的优化汇总
cat docs/shared/opts/completed/v2.12/README.md
```

---

## 🛠️ 维护指南

### 定期审查

- 每月审查进行中的优化
- 评估优化的优先级是否需要调整
- 关闭收益不足或优先级降低的优化
- 整理归档的优化文档

### 统计分析

- 分析高频优化模块，识别系统性问题
- 统计优化收益，评估投入产出比
- 识别优化模式，形成最佳实践
- 改进开发和运维流程

### 知识沉淀

- 典型优化记录到 Graphiti
- 更新性能优化指南
- 编写最佳实践文档
- 分享优化经验

### 推广应用

- 评估优化是否可以推广到其他模块
- 制定推广计划和实施方案
- 跟踪推广效果
- 持续改进优化方法

---

## 🆘 需要帮助？

1. **创建优化** - 参考 [`/optimize-create` 指令文档](../../.cursor/commands/optimize-create.md)
2. **规划优化** - 参考 [`/optimize-spec` 指令文档](../../.cursor/commands/optimize-spec.md)
3. **归档优化** - 参考 [`/optimize-archive` 指令文档](../../.cursor/commands/optimize-archive.md)
4. **查询经验** - 使用 Graphiti 搜索相似优化
5. **性能测试** - 查看性能测试指南（待创建）

---

**最后更新**: 2025-12-01  
**维护者**: TTPOS Team  
**版本**: v1.0.0

