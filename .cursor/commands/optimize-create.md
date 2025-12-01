---
name: optimize-create
description: 创建优化需求记录（优化提案阶段）
---

# /optimize-create - 创建优化需求

## 使用场景

快速创建优化需求文档，记录性能瓶颈、体验问题或技术债务。

> **注意**: 此命令只创建 `optimize.md`（需求描述）。分析完成后，使用 `/optimize-spec` 创建优化方案和任务分解。

## 使用方式

```bash
/optimize-create order-query-performance
/optimize-create api-response-slow --priority high
/optimize-create ui-loading-experience --category ux
```

## 参数

- `optimize_brief`: 必填，优化简述（kebab-case）
  - 格式: `{module}-{brief-description}`
  - module: 业务模块（order, member, product, shop, admin, bmp, kds, pos...）
  - brief-description: 简短优化描述
- `--priority`: 可选，优先级（critical, high, medium, low），默认为 medium
- `--category`: 可选，优化类型（performance, ux, security, maintainability, scalability），默认为 performance
- `--version`: 可选，当前版本号（默认从 `main/version/version.go` 读取）

## 优化 ID 生成规则

自动生成唯一优化 ID：

```
opt-{YYMMDD}-{序号}
例如: opt-251201-001
```

- 序号从 001 开始，当日内递增
- 自动检查当日已有优化数量

## 功能特点

- ✅ 自动生成唯一优化 ID
- ✅ 创建 optimize.md（需求详情）
- ✅ 自动填充基本信息（模块、优先级、类型、版本、日期）
- ✅ 记录提出者信息（从 git config 读取）
- ✅ 初始化状态为「待评估」
- ✅ **搜索 Graphiti**（查找相似优化经验）
- ✅ **关联 Spec**（如果优化与某个功能相关）
- ✅ 提供下一步指引

## 输出产物

```
docs/shared/optimizations/active/opt-{id}-{module}-{brief}/
└── optimize.md  # 优化需求（状态: 待评估）
```

## 优化文档结构

```markdown
# Opt-{ID}: {简短描述}

## 基本信息

| 字段       | 值                    |
| ---------- | --------------------- |
| 优化 ID    | opt-{id}              |
| 模块       | {module}              |
| 优化类型   | {category}            |
| 优先级     | {priority}            |
| 当前版本   | v{version}            |
| 提出日期   | {YYYY-MM-DD}          |
| 提出者     | {name}                |
| 状态       | 🟡 待评估             |

## 优化需求

### 当前问题

描述现状和存在的问题

### 性能指标（如适用）

- **当前性能**: 
- **目标性能**: 
- **提升目标**: 

### 影响面

- **影响终端**: 
- **影响用户**: 
- **业务价值**: 

## 触发原因

为什么需要这次优化？（用户反馈、监控数据、技术债务）

## 初步分析

### 可能原因

### 优化方向

### 预估收益

## 相关链接
```

## 优化类型说明

| 类型              | 说明                       | 典型场景                  |
| ----------------- | -------------------------- | ------------------------- |
| **performance**   | 性能优化                   | 响应慢、查询慢、加载慢    |
| **ux**            | 用户体验优化               | 交互不便、流程繁琐        |
| **security**      | 安全加固                   | 漏洞修复、权限完善        |
| **maintainability** | 可维护性优化             | 代码重构、架构优化        |
| **scalability**   | 可扩展性优化               | 集群支持、容量规划        |

## 状态说明

| 状态         | 说明                     | 下一步            |
| ------------ | ------------------------ | ----------------- |
| **🟡 待评估** | 优化刚创建，等待技术评估 | 分析收益和成本    |
| **🟢 规划中** | 正在制定优化方案         | `/optimize-spec`  |
| **🔵 已完成** | 优化已实施并验证通过     | `/optimize-archive` |
| **⚪ 已取消** | 收益不足或优先级降低     | 直接归档          |

## 工作流位置

```
发现问题 → /optimize-create → 收益评估 → /optimize-spec → 实施优化 → /optimize-archive
              ↑                            ↑                            ↑
          当前命令                       下一步                       最终归档
```

## 智能功能

### 1. 搜索 Graphiti 相似优化

- 使用优化关键词搜索 Graphiti
- 查找是否有相似优化的实施经验
- 自动填充「相关链接」

### 2. 检测重复优化

- 搜索 `docs/shared/optimizations/active/` 中的相似描述
- 如发现可能重复，提示用户确认

### 3. 关联现有 Spec

- 根据模块名称搜索相关的活跃 Spec
- 自动填充「相关链接」中的关联 Spec

### 4. 自动关联监控数据

- 如果是性能优化，提示添加监控链接
- 如果是 UX 优化，提示添加用户反馈链接

## 执行流程

### Step 1: 生成优化 ID

```yaml
获取当前日期: YYMMDD
扫描: docs/shared/optimizations/active/opt-{YYMMDD}-*/
计算序号: 最大序号 + 1
生成: opt-{YYMMDD}-{序号:03d}
```

### Step 2: 搜索 Graphiti

```yaml
搜索关键词: {module} + {brief} + "优化"
IF 找到相似优化 THEN
  提示用户参考
  填充相关链接
```

### Step 3: 检测重复

```yaml
搜索: docs/shared/optimizations/active/*/*-{brief}*/
IF 找到相似优化 THEN
  提示用户
  询问是否继续
```

### Step 4: 创建目录和文件

```yaml
创建目录: docs/shared/optimizations/active/opt-{id}-{module}-{brief}/
创建文件: optimize.md
填充模板: 基本信息 + 需求描述
```

### Step 5: 关联资源

```yaml
搜索 Graphiti: 相关优化记录
搜索 Specs: 相关功能规格
填充链接: 在 optimize.md 中
```

### Step 6: 记录活动日志

按 `activity_log.mdc` 规范记录：

```
| HH:mm | /optimize-create | opt-{id}-{brief} | ✅ | 创建优化需求 |
```

## 后端特定适配

- ✅ 支持三模块（Main: Go + Gin, Admin: PHP + ThinkPHP, BMP: Go + GoFrame）
- ✅ 自动识别技术栈
- ✅ 自动填充模块影响范围
- ✅ 记录相关服务和客户端信息
- ✅ 提供性能监控指标模板

## 错误处理

| 错误类型       | 处理方式                     |
| -------------- | ---------------------------- |
| 参数格式错误   | 显示正确格式和示例           |
| 优化已存在     | 询问是否覆盖或创建新版本     |
| 版本号读取失败 | 提示手动输入版本号           |

## 相关命令

| 命令                | 用途                       |
| ------------------- | -------------------------- |
| `/optimize-create`  | 创建优化需求（当前命令）   |
| `/optimize-spec`    | 创建优化方案和任务         |
| `/optimize-archive` | 归档已完成的优化           |

## 与 Bug 体系的区别

| 优化管理         | Bug 管理      | 区别                       |
| ---------------- | ------------- | -------------------------- |
| 主动发现问题     | 被动响应问题  | 优化是主动改进             |
| 关注收益和成本   | 关注修复影响  | 优化需要评估 ROI           |
| 可以延迟实施     | 必须尽快修复  | 优化可根据优先级排期       |
| 性能/体验指标    | 功能正确性    | 衡量标准不同               |

---

**版本**: v1.0.0  
**创建日期**: 2025-12-01  
**维护者**: 知识管理组  
**状态**: ✅ MVP

