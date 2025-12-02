# ERP Print Format Doctype 通用服务支持 需求文档

> 本文档定义 ERP Print Format Doctype 通用服务支持的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                                                      |
| ----------------- | ----------------------------------------------------------------------------------------------------------------------------------------- |
| **来源 Proposal** | [docs/team/proposals/2025-11/erp-print-format-doctype-service.md](../../../../team/proposals/2025-11/erp-print-format-doctype-service.md) |
| **创建日期**      | 2025-11-24                                                                                                                                |
| **负责人**        | {待分配}                                                                                                                                  |
| **目标 Sprint**   | {待分配}                                                                                                                                  |
| **涉及技术栈**    | [ ] Go (main/) [x] Go (ttpos-bmp/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                                                |

---

## 📋 概述

ERP Print Format Doctype 通用服务支持旨在为 ERP 模块提供统一的 Print Format 服务接口，便于各业务模块复用打印格式管理功能，减少代码重复和维护成本。

本功能主要涉及 ERPNext Print Format API 的封装、服务接口设计和业务逻辑实现。

## 🎯 产品对齐

该功能支持公司 2025 年 Q4 的核心目标：

- **提升开发效率**: 新业务模块可以快速接入打印功能
- **降低维护成本**: 统一管理打印格式，减少代码重复
- **标准化**: 遵循 ERPNext 的 Print Format Doctype 规范

## 📝 用户故事

**作为** 开发人员  
**我想** 使用统一的 Print Format 服务接口  
**以便于** 快速为各业务模块接入打印功能，减少重复代码

---

## 功能需求

### Requirement 1: Print Format 元数据查询

**用户故事**: 作为开发人员，我想查询 Print Format 的元数据信息，以便于了解字段定义和结构

#### 验收标准

1. **WHEN** 调用 Print Format Meta API **THEN** 系统 **SHALL** 返回 Print Format 的元数据信息
2. **IF** DocType 参数为空 **THEN** 系统 **SHALL** 返回错误提示"DocType 不能为空"
3. **WHEN** 元数据查询成功 **THEN** 系统 **SHALL** 返回字段定义、验证规则等完整信息

#### 具体要求

- [ ] 1.1 实现 Print Format Meta 查询方法
- [ ] 1.2 调用 ERPNext API: `/api/v2/doctype/Print Format/meta`
- [ ] 1.3 返回元数据 JSON 结构
- [ ] 1.4 错误处理和日志记录

---

### Requirement 2: Print Format 列表查询

**用户故事**: 作为开发人员，我想根据 DocType 查询 Print Format 列表，以便于获取该 DocType 对应的所有打印格式

#### 验收标准

1. **WHEN** 根据 DocType 查询 Print Format 列表 **THEN** 系统 **SHALL** 返回该 DocType 对应的所有 Print Format
2. **IF** DocType 参数为空 **THEN** 系统 **SHALL** 返回错误提示
3. **WHEN** 查询成功 **THEN** 系统 **SHALL** 返回 Print Format 列表（包含名称、描述等基本信息）

#### 具体要求

- [ ] 2.1 实现 Print Format 列表查询方法
- [ ] 2.2 支持按 DocType 过滤
- [ ] 2.3 支持分页查询（Limit、LimitStart）
- [ ] 2.4 调用 ERPNext API: `/api/v2/document/Print Format`
- [ ] 2.5 返回格式化的 Print Format 列表

---

### Requirement 3: Print Format 详情查询

**用户故事**: 作为开发人员，我想根据名称查询 Print Format 详情，以便于获取完整的打印格式信息（包括 HTML 模板内容）

#### 验收标准

1. **WHEN** 根据名称查询 Print Format 详情 **THEN** 系统 **SHALL** 返回完整的 Print Format 信息（包括 HTML 模板内容）
2. **IF** 名称参数为空 **THEN** 系统 **SHALL** 返回错误提示"Print Format 名称不能为空"
3. **IF** Print Format 不存在 **THEN** 系统 **SHALL** 返回错误提示"Print Format 不存在"

#### 具体要求

- [ ] 3.1 实现 Print Format 详情查询方法
- [ ] 3.2 调用 ERPNext API: `/api/v2/document/Print Format/{name}`
- [ ] 3.3 返回完整的 Print Format 信息（包括 HTML 模板、字段映射等）
- [ ] 3.4 错误处理和日志记录

---

### Requirement 4: Print Format 创建/更新

**用户故事**: 作为开发人员，我想创建或更新 Print Format，以便于管理打印格式模板

#### 验收标准

1. **IF** 创建/更新 Print Format **THEN** 系统 **SHALL** 成功保存并返回结果
2. **IF** Print Format 名称已存在且为创建操作 **THEN** 系统 **SHALL** 返回错误提示"Print Format 已存在"
3. **WHEN** 更新成功 **THEN** 系统 **SHALL** 返回更新后的 Print Format 信息

#### 具体要求

- [ ] 4.1 实现 Print Format 创建方法
- [ ] 4.2 实现 Print Format 更新方法
- [ ] 4.3 调用 ERPNext API: `/api/v2/document/Print Format` (POST/PUT)
- [ ] 4.4 支持完整的 Print Format 数据结构（包括 HTML 模板、字段映射等）
- [ ] 4.5 数据验证和错误处理

---

### Requirement 5: Print Format 删除

**用户故事**: 作为开发人员，我想删除 Print Format，以便于清理不需要的打印格式

#### 验收标准

1. **IF** 删除 Print Format **THEN** 系统 **SHALL** 执行软删除操作
2. **IF** Print Format 名称不存在 **THEN** 系统 **SHALL** 返回错误提示"Print Format 不存在"
3. **WHEN** 删除成功 **THEN** 系统 **SHALL** 返回删除确认信息

#### 具体要求

- [ ] 5.1 实现 Print Format 删除方法
- [ ] 5.2 调用 ERPNext API: `/api/v2/document/Print Format/{name}` (DELETE)
- [ ] 5.3 支持软删除（ERPNext 的 docstatus 机制）
- [ ] 5.4 错误处理和日志记录

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 Logic → Service → ERPNext Client 分层
- **单一职责原则**: PrintFormatService 专注 Print Format 逻辑，复用 ERPNext 通用服务
- **模块化设计**: Print Format 逻辑可独立测试和复用
- **依赖管理**: PrintFormatService 依赖 IDocument 和 IDoctype 接口
- **遵循规范**:
  - `.cursor/rules/go-bmp.mdc` - Go BMP 开发规范
  - `.cursor/rules/api.mdc` - API 设计规范

### API 设计要求

- [x] URL 使用 snake_case 命名
- [x] data 字段必须是对象
- [x] 响应格式：`{code: 1, message: "success", data: {}}`
- [x] 参考: `.cursor/rules/api.mdc` - API 设计规范

### 性能要求

- [x] Print Format 查询响应时间 < 500ms（P0）
- [x] 支持并发查询
- [x] 错误响应时间 < 200ms

### 测试要求

- [x] PrintFormatService 测试覆盖率 ≥ 70%
- [x] Logic 层测试覆盖率 ≥ 80%
- [x] 集成测试覆盖完整 Print Format 流程
- [x] 参考: `.cursor/rules/go-bmp.mdc` - 测试规范

### 安全要求

- [x] Print Format API 需要身份验证（ERPNext Token）
- [x] 参数验证：DocType 和名称必须验证
- [x] 参考: `.cursor/rules/security.mdc` - 安全开发规范

### 可靠性要求

- [x] ERPNext API 调用失败时返回详细错误信息
- [x] 记录详细错误日志（Logger）
- [x] 支持重试机制（可选）

---

## 验收标准

### 功能验收

1. **元数据查询**: 调用 Meta API → 返回 Print Format 元数据信息
2. **列表查询**: 根据 DocType 查询 → 返回 Print Format 列表
3. **详情查询**: 根据名称查询 → 返回完整的 Print Format 信息
4. **创建/更新**: 创建或更新 Print Format → 成功保存并返回结果
5. **删除**: 删除 Print Format → 执行软删除操作

### 测试验收

1. **单元测试**: Service 和 Logic 覆盖率达标
2. **API 测试**: 所有接口测试通过，包含异常场景
3. **集成测试**: 端到端 Print Format 流程测试通过

### 文档验收

1. **技术文档**: design.md 包含完整的 API 设计、服务设计
2. **API 文档**: docs/shared/api/erp_api.md 已更新
3. **代码注释**: 所有公共方法都有中文注释

---

## 约束条件

### 技术约束

#### Go BMP 模块

- 必须使用 GoFrame 2.x 框架
- PrintFormatService 接口以 `I` 开头：`IPrintFormatService`
- Service 依赖 IDocument 和 IDoctype 接口
- 不使用 panic，所有错误返回 error
- 使用 gerror.Wrapf 包装错误

### 业务约束

- Print Format 必须关联 DocType
- Print Format 名称必须唯一
- 删除操作使用软删除（docstatus）

### 资源约束

- 开发时间: 3-5 天
- Story Point: 5 (符合 ≤ 5 标准)

---

## 依赖关系

### 技术依赖

- `github.com/gogf/gf/v2` - GoFrame 框架
- ERPNext API v2

### 服务依赖

- **ERPNext 服务**: 依赖 ERPNext Print Format API
- **ERPNext Client**: 复用现有的 ERPNext 客户端实现

### 业务依赖

- 依赖现有的 Document 服务（IDocument）
- 依赖现有的 Doctype 服务（IDoctype）
- 依赖现有的 ERPNext 客户端（GetClient）

---

## 风险和缓解

### 风险 1: ERPNext Print Format API 兼容性

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 参考现有 `doctype.go` 的实现模式
- 充分测试各种 DocType 的 Print Format 场景
- 记录 API 调用日志便于排查

### 风险 2: 打印模板格式标准化

**影响**: 中  
**概率**: 低  
**缓解措施**:

- 先实现基础功能，再逐步完善高级特性
- 参考现有打印模板格式
- 提供模板验证机制

### 风险 3: 多语言支持复杂度

**影响**: 低  
**概率**: 低  
**缓解措施**:

- 第一阶段不实现多语言，后续迭代添加
- 使用 ERPNext 原生的多语言机制

---

## 时间表

- **Phase 1 - 服务接口设计**: 1 天
- **Phase 2 - 核心实现**: 2 天
- **Phase 3 - 测试和优化**: 1-2 天
- **总计**: 4-5 天（SP = 5）

---

## 参考资料

### 核心规范

- `.cursor/rules/go-bmp.mdc` - Go BMP 开发规范
- `.cursor/rules/api.mdc` - API 设计规范
- `.cursor/rules/security.mdc` - 安全开发规范

### 架构文档

- `ttpos-bmp/app/ttpos-erp/internal/logic/erpnext/doctype.go` - Doctype 服务实现参考
- `ttpos-bmp/app/ttpos-erp/internal/logic/selling/sale_order.go` - 业务服务实现参考
- `ttpos-bmp/app/ttpos-erp/internal/service/erpnext.go` - 服务注册机制

### 开发指南

- ERPNext Print Format API 文档
- 现有打印模板: `ttpos-bmp/app/ttpos-erp/manifest/printformat/html/`

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 活动日志：`docs/team/activities/2025-11/2025-11-24.md`
- 提醒：当 Print Format 功能产出可复用经验或重大决策时，创建 Episode 并在此占位更新。

---

**版本**: v1.0.0  
**创建日期**: 2025-11-24  
**作者**: 后端开发组  
**审核者**: {待分配}
