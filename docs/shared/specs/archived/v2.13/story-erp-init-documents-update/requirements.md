# ERP 文档初始化支持更新模式 需求文档

> ⚠️ **已归档** - 此 Spec 已随 v2.13 发布（强制归档）。
>
> - 归档时间: 2026-01-12
> - 归档人: weifashi
> - 完成状态: 部分完成或未完成（强制归档）


> 本文档定义 ERP 文档初始化支持更新模式 的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2026-01/erp-init-documents-update-support.md](../../../../team/proposals/2026-01/erp-init-documents-update-support.md) |
| **创建日期**      | 2026-01-04                                                                                                 |
| **负责人**        | rikugun                                                                                                       |
| **目标 Sprint**   | Sprint N                                                                                                   |
| **涉及技术栈**    | [x] Go (ttpos-bmp/) [ ] Go (main/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                   |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 已通过 |
| **审核人**   | rikugun             |
| **审核日期** | 2026-01-04             |
| **审核意见** | 技术方案清晰，工作量评估合理，向后兼容，批准进入设计阶段         |

---

## 📋 概述

当前 `initDocumentsFromDir` 方法只支持创建（Create）操作，导致重复初始化或配置更新时无法处理已存在的文档。本需求旨在增强该方法，使其能够根据 JSON 数据中的 `name` 字段智能判断并执行创建或更新操作，实现配置管理的幂等性。

**核心价值**：
- 提升运维效率：支持安全地重复执行初始化流程
- 简化配置管理：通过 JSON 文件统一管理文档配置
- 降低错误率：减少手动操作 ERPNext 后台的需求
- 加速版本升级：自动更新文档配置，无需手动逐个修改

## 🎯 产品对齐

该功能属于基础设施优化，支持以下产品目标：
1. **提升系统可维护性**：通过代码化配置管理，降低运维成本
2. **加速迭代速度**：简化版本升级流程，支持快速部署
3. **降低风险**：减少人工操作，避免配置错误

## 📝 用户故事

**作为** 运维人员/开发人员  
**我想** 能够通过 JSON 文件更新已有的 ERPNext 文档配置  
**以便于** 实现配置的版本控制和批量更新，提升运维效率

---

## 功能需求

### Requirement 1: 智能判断创建或更新

**用户故事**: 作为运维人员，我想系统能够自动判断是创建新文档还是更新已有文档，以便于可以安全地重复执行初始化流程

#### 验收标准

1. **WHEN** JSON 文件中的 `name` 字段不为空 **THEN** 系统 **SHALL** 调用 `service.Document().Update()` 更新文档
2. **WHEN** JSON 文件中的 `name` 字段为空 **THEN** 系统 **SHALL** 调用 `service.Document().Create()` 创建文档
3. **WHEN** `name` 字段存在但值为空字符串 **THEN** 系统 **SHALL** 调用 `service.Document().Create()` 创建文档
4. **WHEN** 更新或创建操作成功 **THEN** 系统 **SHALL** 记录详细的成功日志，包含文件路径

#### 具体要求

- [x] 1.1 从 `docData` 中读取 `name` 字段并进行类型断言
- [x] 1.2 判断 `name` 字段是否存在且不为空
- [x] 1.3 根据判断结果调用对应的 service 方法（Create 或 Update）
- [x] 1.4 保持现有的错误处理机制

---

### Requirement 2: 错误处理和日志记录

**用户故事**: 作为开发人员，我想系统能够详细记录操作日志和错误信息，以便于快速定位和解决问题

#### 验收标准

1. **WHEN** 更新操作失败 **THEN** 系统 **SHALL** 记录详细的错误日志，包含 DocType、文件路径和错误信息
2. **WHEN** 创建操作失败 **THEN** 系统 **SHALL** 记录详细的错误日志，包含 DocType、文件路径和错误信息
3. **WHEN** 操作成功 **THEN** 系统 **SHALL** 记录成功日志，明确标注是"创建"还是"更新"
4. **IF** 操作失败 **THEN** 系统 **SHALL** 继续处理其他文件，不中断整个初始化流程

#### 具体要求

- [x] 2.1 更新操作失败时记录错误日志（保持现有格式）
- [x] 2.2 创建操作失败时记录错误日志（保持现有格式）
- [x] 2.3 更新成功时记录 Info 级别日志，格式：`{ItemName}更新成功: {path}`
- [x] 2.4 创建成功时记录 Info 级别日志，格式：`{ItemName}创建成功: {path}`

---

### Requirement 3: 向后兼容性

**用户故事**: 作为开发人员，我想新功能不影响现有的初始化流程，以便于平滑升级

#### 验收标准

1. **WHEN** 现有的 JSON 文件（不包含 `name` 字段）被处理 **THEN** 系统 **SHALL** 正常执行创建操作
2. **WHEN** 新的 JSON 文件（包含 `name` 字段）被处理 **THEN** 系统 **SHALL** 执行更新操作
3. **WHEN** 重复执行初始化流程 **THEN** 第一次创建，第二次更新，都应成功

#### 具体要求

- [x] 3.1 不修改 JSON 文件格式要求
- [x] 3.2 不修改 `initDocumentsFromDir` 方法的函数签名
- [x] 3.3 不修改调用 `initDocumentsFromDir` 的其他代码

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 遵循 GoFrame 框架的 Controller → Service → Logic 分层
- **单一职责原则**: `initDocumentsFromDir` 方法专注于文档初始化逻辑
- **模块化设计**: Service 和 Logic 应独立且可复用
- **遵循规范**:
  - `.cursor/rules/go-bmp.mdc` - Go BMP 微服务规范
  - `ttpos-bmp/.cursor/rules/go-rules.mdc` - GoFrame 开发规范
  - `.cursor/rules/structs.mdc` - 项目结构规范

### API 设计要求

本需求不涉及对外 API，为内部方法优化。

### 数据库设计要求

本需求不涉及数据库变更。

### 性能要求

- [x] 文件读取和 JSON 解析时间保持不变
- [x] 单个文档的创建/更新操作时间保持不变
- [x] 不增加额外的网络调用或数据库查询

### 测试要求

- [x] Logic 层测试覆盖率 ≥ 80%
- [x] 测试用例覆盖创建、更新、错误处理三种场景
- [x] 集成测试覆盖重复执行初始化的场景

### 安全要求

- [x] JSON 文件读取使用安全的文件路径
- [x] 防止路径遍历攻击
- [x] 敏感信息不记录到日志中

### 可靠性要求

- [x] 单个文件处理失败不影响其他文件
- [x] 错误日志记录完整，便于排查问题
- [x] 保持现有的故障恢复机制

---

## 验收标准

### 功能验收

1. **创建新文档**：JSON 文件中 `name` 字段为空，系统成功创建文档并记录日志
2. **更新已有文档**：JSON 文件中 `name` 字段不为空，系统成功更新文档并记录日志
3. **重复执行初始化**：第一次创建，第二次更新，都成功且无错误
4. **错误处理**：单个文件失败不影响其他文件的处理

### 测试验收

1. **单元测试**: 覆盖率达到 80% 以上
2. **集成测试**: 重复执行初始化流程的端到端测试通过
3. **手动测试**: 在开发环境验证创建和更新两种场景

### 文档验收

1. **技术文档**: design.md 完整且准确（产品审核通过后创建）
2. **代码注释**: 修改的方法添加详细的中文注释
3. **测试文档**: tasks.md 中的测试任务完成（产品审核通过后创建）

---

## 约束条件

### 技术约束

#### Go BMP 模块

- 必须使用 GoFrame 2.x
- 禁止修改 dao/entity/do/ 目录（自动生成）
- 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`
- 所有注释使用中文
- 不使用 panic，返回 error

### 业务约束

- 不改变现有 JSON 文件的格式和结构
- 不影响其他调用 `initDocumentsFromDir` 的功能
- 保持方法的向后兼容性

### 资源约束

- 开发时间: 0.5 天
- Story Point: 1 (必须 ≤ 5)

---

## 依赖关系

### 技术依赖

- `github.com/gogf/gf/v2` - GoFrame 框架
- `ttpos-bmp/app/ttpos-erp/internal/service` - Document 服务

### 服务依赖

- **ERPNext API**: 依赖 `service.Document().Create()` 和 `service.Document().Update()` 方法

### 业务依赖

- 无前置功能依赖
- 无其他模块依赖

---

## 风险和缓解

### 风险 1: Update 方法行为不确定

**影响**: 中  
**概率**: 低  
**缓解措施**:

- 在开发前先调研 `service.Document().Update()` 的 API 文档和使用示例
- 在测试环境充分测试 Update 方法的各种场景
- 记录详细的错误日志，便于问题排查

### 风险 2: JSON 数据格式兼容性

**影响**: 低  
**概率**: 低  
**缓解措施**:

- 编写测试用例覆盖不同格式的 JSON 文件
- 在代码中进行类型断言和空值检查
- 保持向后兼容，不强制要求 `name` 字段

---

## 时间表

- **Phase 1 - 代码实现**: 0.5 小时
  - 修改 `initDocumentsFromDir` 方法逻辑
  - 添加代码注释
- **Phase 2 - 测试验证**: 1 小时
  - 编写单元测试
  - 在开发环境手动测试
- **Phase 3 - 文档更新**: 0.5 小时
  - 更新方法注释
  - 编写技术文档（如需要）
- **Phase 4 - 代码审查**: 0.5 小时
  - 代码审查
  - 修改反馈意见
- **总计**: 0.5 天（SP = 1）

---

## 参考资料

### 核心规范

- `.cursor/rules/go-bmp.mdc` - Go BMP 开发规范
- `ttpos-bmp/.cursor/rules/go-rules.mdc` - GoFrame 开发规范
- `.cursor/rules/structs.mdc` - 项目结构规范

### 架构文档

- `docs/human/architecture/go-bmp-architecture.md` - Go BMP 架构（如有）
- ERPNext API 文档: https://frappeframework.com/docs/user/en/api

### 开发指南

- `ttpos-bmp/README.md` - BMP 项目说明
- `ttpos-bmp/MIGRATION_QUICK_START.md` - 迁移快速入门

### 外部参考

- Frappe Framework API: https://frappeframework.com/docs/user/en/api
- GoFrame 官方文档: https://goframe.org.cn

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/2026-01/2026-01-04.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2026-01-04  
**作者**: rikugun  
**审核者**: 待审核

