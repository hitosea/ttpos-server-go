# 简化菜单更新接口响应结构 需求文档

> 本文档定义 简化菜单更新接口响应结构 的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-12/story-bmp-takeout-proto-simplify-response.md](../../../../team/proposals/2025-12/story-bmp-takeout-proto-simplify-response.md) |
| **创建日期**      | 2025-12-16                                                                                                 |
| **负责人**        | rikugun                                                                                                       |
| **目标 Sprint**   | 当前 Sprint                                                                                                   |
| **涉及技术栈**    | [ ] Go (main/) [x] Go (ttpos-bmp/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                   |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 已通过                   |
| **审核人**   | rikugun                  |
| **审核日期** | 2025-12-16               |
| **审核意见** | 技术重构，风险低，批准进入设计阶段 |

---

## 📋 概述

当前 `menu.proto` 中的 `UpdateMenuItemResp` 和 `UpdateMenuModifierResp` 定义了冗余的 `error_code` 和 `error_message` 字段，而 RPC 方法已经返回 `takeout.ApiResponse`，其中已包含统一的 `code` 和 `message` 字段用于错误处理。本次需求旨在移除这些冗余字段，统一使用 `ApiResponse` 进行错误信息传递，简化接口设计，减少数据冗余。

## 🎯 产品对齐

- **统一响应格式**：所有 API 使用一致的错误处理模式，提升开发体验
- **减少维护成本**：简化接口结构，降低文档和测试成本
- **提升代码质量**：遵循 DRY 原则，避免数据冗余

## 📝 用户故事

**作为** 后端开发人员  
**我想** 移除 `UpdateMenuItemResp` 和 `UpdateMenuModifierResp` 中的冗余错误字段  
**以便于** 统一使用 `ApiResponse` 进行错误处理，简化接口设计

---

## 功能需求

### Requirement 1: 移除 UpdateMenuItemResp 中的冗余错误字段

**用户故事**: 作为后端开发人员，我想移除 `UpdateMenuItemResp.error_code` 和 `UpdateMenuItemResp.error_message`，以便于统一使用 `ApiResponse` 进行错误处理

#### 验收标准

1. **WHEN** 修改 `menu.proto` 中的 `UpdateMenuItemResp` **THEN** 系统 **SHALL** 移除 `error_code` 和 `error_message` 字段
2. **WHEN** 重新生成 proto 代码 **THEN** 系统 **SHALL** 生成的 Go 代码不包含这些字段
3. **IF** 有代码依赖这些字段 **THEN** 系统 **SHALL** 更新为使用 `ApiResponse.code` 和 `ApiResponse.message`

#### 具体要求

- [ ] 1.1 修改 `ttpos-bmp/app/ttpos-takeout/manifest/protobuf/menu/menu.proto`，移除 `UpdateMenuItemResp.error_code` 和 `UpdateMenuItemResp.error_message`
- [ ] 1.2 执行 `gf gen pb` 重新生成 proto 代码
- [ ] 1.3 检查并更新相关 DTO 和逻辑代码（如有依赖）

---

### Requirement 2: 移除 UpdateMenuModifierResp 中的冗余错误字段

**用户故事**: 作为后端开发人员，我想移除 `UpdateMenuModifierResp.error_code` 和 `UpdateMenuModifierResp.error_message`，以便于统一使用 `ApiResponse` 进行错误处理

#### 验收标准

1. **WHEN** 修改 `menu.proto` 中的 `UpdateMenuModifierResp` **THEN** 系统 **SHALL** 移除 `error_code` 和 `error_message` 字段
2. **WHEN** 重新生成 proto 代码 **THEN** 系统 **SHALL** 生成的 Go 代码不包含这些字段
3. **IF** 有代码依赖这些字段 **THEN** 系统 **SHALL** 更新为使用 `ApiResponse.code` 和 `ApiResponse.message`

#### 具体要求

- [ ] 2.1 修改 `ttpos-bmp/app/ttpos-takeout/manifest/protobuf/menu/menu.proto`，移除 `UpdateMenuModifierResp.error_code` 和 `UpdateMenuModifierResp.error_message`
- [ ] 2.2 执行 `gf gen pb` 重新生成 proto 代码
- [ ] 2.3 检查并更新相关 DTO 和逻辑代码（如有依赖）

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 Controller → Service → Repository 分层
- **单一职责原则**: 每个文件应有单一、明确的目的
- **模块化设计**: Service 和 Repository 应独立且可复用
- **依赖管理**: Service 只能依赖其他 Service 接口，不能直接依赖 Repository
- **遵循规范**:
  - `.cursor/rules/go-bmp.mdc` - Go BMP 微服务规范
  - `.cursor/rules/structs.mdc` - 项目结构规范

### API 设计要求

- [x] 统一使用 `takeout.ApiResponse` 进行错误处理
- [x] 响应格式：`{code, message, data{}}`
- [x] 参考: `.cursor/rules/api.mdc` - API 设计规范

### 数据库设计要求

- [ ] 不涉及数据库变更

### 性能要求

- [ ] 不涉及性能优化

### 测试要求

- [ ] 重新生成 proto 代码后，确保编译通过
- [ ] 如有相关测试，需更新测试用例
- [ ] 参考: `.cursor/rules/go-bmp.mdc` - 测试规范

### 安全要求

- [ ] 不涉及安全变更
- [ ] 参考: `.cursor/rules/security.mdc` - 安全开发规范

### 可靠性要求

- [ ] 确保修改后接口向后兼容（如已发布）
- [ ] 错误日志记录（使用 Logger）
- [ ] 参考: `.cursor/rules/go-bmp.mdc` - 可靠性规范

---

## 验收标准

### 功能验收

1. **Proto 文件修改**: `UpdateMenuItemResp` 和 `UpdateMenuModifierResp` 中已移除 `error_code` 和 `error_message` 字段
2. **代码生成**: 执行 `gf gen pb` 后，生成的 Go 代码不包含这些字段
3. **编译通过**: 修改后项目编译通过，无编译错误
4. **代码检查**: 确认无代码依赖已移除的字段

### 测试验收

1. **编译测试**: 项目编译通过
2. **代码检查**: 使用 `gf vet` 检查代码规范

### 文档验收

1. **Proto 文档**: proto 文件注释清晰
2. **变更记录**: 如有必要，更新相关文档

---

## 约束条件

### 技术约束

#### Go BMP 模块

- 必须使用 GoFrame 2.x
- 禁止修改 dao/entity/do/ 目录（自动生成）
- gRPC 服务必须注册到 Nacos
- 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`
- Proto 代码生成使用 `gf gen pb`

### 业务约束

- 确保修改不影响现有功能（功能刚开发完成，尚未发布）

### 资源约束

- 开发时间: 0.5 天
- Story Point: 1 (必须 ≤ 5)

---

## 依赖关系

### 技术依赖

- `github.com/gogf/gf/v2` - GoFrame 框架
- `ttpos-bmp/app/ttpos-takeout/manifest/protobuf/takeout_api.proto` - ApiResponse 定义

### 服务依赖

- 无

### 业务依赖

- 关联 Spec: `story-bmp-grab-menu-update-item-modifier`（菜单更新功能）

---

## 风险和缓解

### 风险 1: 客户端已依赖 error_code/error_message 字段

**影响**: 中  
**概率**: 低  
**缓解措施**:

- 确认当前无客户端依赖这些字段（功能刚开发完成，尚未发布）
- 如有依赖，需要同步更新客户端代码

### 风险 2: 代码生成失败

**影响**: 低  
**概率**: 低  
**缓解措施**:

- 确保 proto 文件语法正确
- 执行 `gf gen pb` 前检查 proto 文件格式

---

## 时间表

- **Phase 1 - Proto 文件修改**: 0.2 天
- **Phase 2 - 代码生成和验证**: 0.2 天
- **Phase 3 - 代码检查和测试**: 0.1 天
- **总计**: 0.5 天（SP = 1）

---

## 参考资料

### 核心规范

- `.cursor/rules/go-bmp.mdc` - Go BMP 开发规范
- `.cursor/rules/api.mdc` - API 设计规范
- `.cursor/rules/structs.mdc` - 项目结构规范

### 架构文档

- `docs/human/architecture/go-bmp-architecture.md` - Go BMP 架构

### 开发指南

- `docs/human/guides/go-bmp-development.md` - Go BMP 开发指南

### 外部参考

- Protocol Buffers 官方文档
- GoFrame 代码生成文档

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-12-16  
**作者**: rikugun  
**审核者**: {审核者}

