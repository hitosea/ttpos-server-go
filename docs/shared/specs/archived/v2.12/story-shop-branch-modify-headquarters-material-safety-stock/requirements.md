> ⚠️ **已归档** - 此 Spec 已随 v2.12 发布。
>
> - 归档时间: 2026-01-04
> - 归档人: weifashi

# 子店可修改总店同步物品安全库存 需求文档

> 本文档定义子店可修改总店同步物品安全库存功能的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-12/branch-modify-headquarters-material-safety-stock.md](../../../../team/proposals/2025-12/branch-modify-headquarters-material-safety-stock.md) |
| **来源任务**      | DooTask #37482                                                                                               |
| **创建日期**      | 2025-12-08                                                                                                   |
| **目标版本**      | v2.11.0                                                                                                      |
| **负责人**        | 待分配                                                                                                       |
| **目标 Sprint**   | Sprint TBD                                                                                                   |
| **涉及技术栈**    | [x] Go (main/) [ ] Go (ttpos-bmp/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                   |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 待审核                   |
| **审核人**   | {审核人姓名}             |
| **审核日期** | {YYYY-MM-DD}             |
| **审核意见** | {审核意见，如有}         |

---

## 📋 概述

允许子店修改总店同步下来的物品的安全库存，并优化同步逻辑，避免同步时覆盖子店已调整的库存数据。

**核心功能**：
1. **新增接口**：提供子店修改物品安全库存的独立接口，允许子店修改自己物品和总店同步物品的安全库存值。
2. **同步逻辑优化**：修改 `SyncMaterial` 方法，当子店已有该物品时（通过 `uuid` 匹配），保留子店的安全库存，不覆盖为总店的安全库存。

**业务价值**：
- 提升子店运营灵活性，允许子店根据实际情况调整安全库存
- 避免同步时覆盖子店已调整的库存数据，保护子店的本地化配置
- 提高库存预警的准确性，减少误报
- 支持子店独立管理库存策略

---

## 📝 用户故事

**作为** 子店管理员  
**我想** 修改总店同步下来的物品的安全库存  
**以便于** 根据本店实际情况设置合适的库存预警阈值

**作为** 子店管理员  
**我想** 在同步总店物品时，保留本店已调整的库存数据  
**以便于** 避免同步操作覆盖本店的本地化配置

---

## 功能需求

### Requirement 1: 子店修改物品安全库存接口

**用户故事**: 作为子店管理员，我想修改物品的安全库存（包括自己创建的物品和总店同步下来的物品），以便于根据本店实际情况设置合适的库存预警阈值

#### 验收标准

1. **WHEN** 子店管理员调用修改安全库存接口 **THEN** 系统 **SHALL** 允许修改子店中所有物品的安全库存（包括自己创建的物品和总店同步的物品）
2. **WHEN** 非子店账号（`headquarter_uuid = 0`）调用修改安全库存接口 **THEN** 系统 **SHALL** 返回错误提示 "非子店账号无法修改"
3. **WHEN** 物品不存在或已删除 **THEN** 系统 **SHALL** 返回错误提示 "物品不存在"
4. **WHEN** 修改成功 **THEN** 系统 **SHALL** 更新物品的 `SafetyStock` 字段并返回成功响应

#### 具体要求

- [x] 1.1 创建 API 接口 `POST /api/v1/shop/material/update_safety_stock`
- [x] 1.2 请求参数：`uuid`（物品UUID，必填）、`safety_stock`（安全库存值，必填，可为 null）
- [x] 1.3 权限校验：只有子店账号（`headquarter_uuid > 0`）才能调用
- [x] 1.4 业务校验：子店可以修改自己物品和总店同步物品的安全库存
- [ ] 1.5 实现 Service 方法 `UpdateMaterialSafetyStock(ctx, uuid, safetyStock) error`
- [ ] 1.6 使用事务保护更新操作
- [ ] 1.7 返回标准响应格式：`{code, message, data{}}`

---

### Requirement 2: 同步逻辑优化 - 保护子店库存数据

**用户故事**: 作为子店管理员，我想在同步总店物品时，保留本店已调整的库存数据，以便于避免同步操作覆盖本店的本地化配置

#### 验收标准

1. **WHEN** 子店同步总店物品时，子店已有该物品（通过 `uuid` 匹配）且安全库存不为 nil **THEN** 系统 **SHALL** 保留子店的安全库存，不覆盖为总店的安全库存
2. **WHEN** 子店同步总店物品时，子店已有该物品（通过 `uuid` 匹配）但安全库存为 nil **THEN** 系统 **SHALL** 保留 nil，不覆盖为总店的安全库存
3. **WHEN** 子店同步总店物品时，子店没有该物品 **THEN** 系统 **SHALL** 创建新记录并同步所有字段（包括总店的安全库存）
4. **WHEN** 同步过程中发生错误 **THEN** 系统 **SHALL** 回滚事务并返回错误信息

#### 具体要求

- [x] 2.1 修改 `SyncMaterial` 方法（`main/app/service/material.go` 第3014行）
- [x] 2.2 同步前获取子店中已存在的总部物品的安全库存，构建 `uuid -> *safety_stock` 映射（包括 nil 值）
- [x] 2.3 在创建物品时，如果子店已有该物品（通过 `uuid` 匹配），则保留子店的安全库存（包括 nil）；否则使用总店的安全库存
- [x] 2.4 统一删除后重建，在重建时保留子店已调整的安全库存（包括 nil）
- [x] 2.5 使用事务保护同步操作
- [x] 2.6 添加详细的错误日志记录

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 Controller → Service → Repository 分层
- **单一职责原则**: 每个文件应有单一、明确的目的
- **模块化设计**: Service 和 Repository 应独立且可复用
- **依赖管理**: Service 只能依赖其他 Service 接口，不能直接依赖 Repository
- **遵循规范**:
  - `.cursor/rules/go-main.mdc` - Go Main 开发规范
  - `.cursor/rules/api.mdc` - API 设计规范
  - `.cursor/rules/structs.mdc` - 项目结构规范

### API 设计要求

- [x] URL 使用 snake_case 命名（`/api/v1/shop/material/update_safety_stock`）
- [x] data 字段必须是对象，不能是 null 或数组
- [ ] 分页信息统一放在 meta 中（本功能不需要分页）
- [x] 响应格式：`{code, message, data{}}`
- [ ] 参考: `.cursor/rules/api.mdc` - API 设计规范

### 数据库设计要求

- [ ] 不需要新增表或字段（使用现有 `ttpos_material` 表的 `safety_stock` 字段）
- [ ] 时间字段使用 int 类型，\_time 结尾，默认值 0
- [ ] UUID 字段使用 bigint unsigned
- [ ] 参考: `.cursor/rules/database.mdc` - 数据库开发规范

### 性能要求

- [ ] 接口响应时间 < 200ms
- [ ] 数据库查询优化（使用索引：`idx_code`、`idx_headquarter_uuid`）
- [ ] 同步操作使用事务，确保数据一致性

### 测试要求

- [ ] Service 层测试覆盖率 ≥ 70%
- [ ] Repository 层测试覆盖率 ≥ 80%
- [ ] 集成测试覆盖核心流程（修改安全库存、同步逻辑）
- [ ] API 测试覆盖所有接口
- [ ] 参考: `.cursor/rules/go-main.mdc` - 测试规范

### 安全要求

- [x] 所有 API 需要身份验证
- [ ] 权限校验：只有子店账号才能调用
- [ ] 业务校验：只能修改总店同步的物品
- [ ] SQL 注入防护（使用参数化查询）
- [ ] 参考: `.cursor/rules/security.mdc` - 安全开发规范

### 可靠性要求

- [ ] 网络异常时优雅降级
- [x] 事务管理（保证数据一致性）
- [ ] 错误日志记录（使用 Logger）
- [ ] 故障恢复机制

---

## 验收标准

### 功能验收

1. **修改安全库存接口**: 子店可以成功修改物品的安全库存（包括自己创建的物品和总店同步的物品）
2. **权限校验**: 非子店账号无法调用接口
3. **业务校验**: 子店可以修改自己物品和总店同步物品的安全库存
4. **同步逻辑优化**: 同步时如果子店已有该物品（通过 uuid 匹配），则保留子店的安全库存（包括 nil）
5. **新物品同步**: 新物品同步时包含所有字段（包括总店的安全库存）

### 测试验收

1. **单元测试**: Service 层核心方法测试通过
2. **API 测试**: 接口调用成功，权限和业务校验正确
3. **集成测试**: 端到端流程测试通过（修改安全库存、同步逻辑）
4. **边界测试**: 测试各种边界情况（物品不存在、非子店账号等）

### 文档验收

1. **技术文档**: design.md 完整且准确（待创建）
2. **API 文档**: API 接口文档完整
3. **测试文档**: tasks.md 中的测试任务完成（待创建）

---

## 约束条件

### 技术约束

#### Go Main 模块

- 必须使用 Gin 框架
- 接口以 `I` 开头，实现以 `Impl` 结尾
- Service 只能依赖其他 Service 接口
- Repository 只能持有 db 实例，不能持有 DBManager
- 不使用 panic，返回 error
- URL 使用 snake_case 命名
- data 字段必须是对象

### 业务约束

- 只有子店账号（`headquarter_uuid > 0`）才能修改安全库存
- 子店可以修改自己物品和总店同步物品的安全库存
- 同步时通过 `uuid` 匹配物品，如果子店已有该物品，则保留子店的安全库存
- 同步时只保护 `SafetyStock` 字段（如果子店已存在该物品）

### 资源约束

- 开发时间: 2-3 天
- Story Point: 3 SP（待技术评审确认）

---

## 依赖关系

### 技术依赖

- `main/app/service/material.go` - 物品服务（`SyncMaterial` 方法）
- `main/app/model/material.go` - 物品模型
- `main/app/repository/material.go` - 物品数据访问层
- `main/pkg/database` - 数据库管理器

### 业务依赖

- 连锁店模式支持（总部-分店关系）
- 现有同步机制（`SyncTask`, `SyncTaskItem`）
- 权限系统（账号类型区分）

---

## 风险和缓解

### 风险 1: 同步逻辑修改可能影响现有同步流程

**影响**: 高  
**概率**: 中  
**缓解措施**:

- 充分测试同步逻辑，确保不影响新物品的同步
- 添加单元测试和集成测试，确保逻辑正确
- 代码 review 确保修改正确
- 在测试环境充分验证后再上线

### 风险 2: 子店修改安全库存后，如果再次同步可能会被覆盖

**影响**: 中  
**概率**: 低（已通过同步逻辑优化解决）  
**缓解措施**:

- 明确同步策略：子店已有物品时，不覆盖库存字段；新物品时，同步所有字段
- 在需求文档中明确说明同步行为
- 在用户文档中说明同步规则

---

## 时间表

- **Phase 1 - API 接口开发**: 1 天
  - 创建 API 接口
  - 实现 Service 方法
  - 权限和业务校验
- **Phase 2 - 同步逻辑优化**: 1 天
  - 修改 `SyncMaterial` 方法
  - 实现库存字段保护逻辑
  - 测试验证
- **Phase 3 - 测试和文档**: 0.5-1 天
  - 单元测试
  - 集成测试
  - API 测试
  - 文档完善
- **总计**: 2.5-3 天（SP = 3）

---

## 参考资料

### 核心规范

- `.cursor/rules/go-main.mdc` - Go Main 核心约束
- `.cursor/rules/api.mdc` - API 设计规范
- `.cursor/rules/database.mdc` - 数据库开发规范
- `.cursor/rules/security.mdc` - 安全开发规范

### 相关文档

- **现有同步逻辑**: `main/app/service/material.go` - `SyncMaterial` 方法（第2903行）
- **物品模型**: `main/app/model/material.go`
- **类似功能**: `shop-headquarters-branch-granular-sync-backend`（总部-分店颗粒化同步）

### 开发指南

- `docs/human/guides/go-main-development.md` - Go Main 开发指南
- `docs/human/guides/api-design-guide.md` - API 设计指南

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-12-08  
**作者**: 曾振华  
**审核者**: 待分配  
**关联任务**: DooTask #37482
