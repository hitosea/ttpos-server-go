> ⚠️ **已归档** - 此 Spec 已随 v2.12 发布。
>
> - 归档时间: 2026-01-04
> - 归档人: weifashi

# Get Item Stock By Bin Service 需求文档

> 本文档定义 获取商品按货位分组的库存信息 的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-12/get-item-stock-by-bin-service.md](../../../../team/proposals/2025-12/get-item-stock-by-bin-service.md) |
| **创建日期**      | 2025-12-26                                                                                                 |
| **负责人**        | rikugun                                                                                                       |
| **目标 Sprint**   | Sprint {N}                                                                                                   |
| **涉及技术栈**    | [x] Go (ttpos-bmp/) [ ] Go (main/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                   |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 已通过                     |
| **审核人**   | rikugun                   |
| **审核日期** | 2025-12-26                |
| **审核意见** | 需求明确，字段定义清晰，同意进入设计阶段 |

---

## 📋 概述

新增 GetItemStockByBin 服务接口，支持根据仓库和商品代码查询货位库存信息。在现有的 stock_bin.go 文件中新增 GetItemStockByBin 方法，复用已有的 Bin 查询逻辑，实现库存按货位分组查询功能。

## 🎯 产品对齐

该功能为库存管理系统提供按货位查询库存的能力，支持仓库管理人员更精确地了解商品在不同货位上的分布情况，提高库存管理的效率和准确性。

## 📝 用户故事

**作为** 仓库管理员/库存管理人员  
**我想** 根据仓库和商品代码查询货位库存信息  
**以便于** 快速了解商品在不同货位的库存分布情况

---

## 功能需求

### Requirement 1: GetItemStockByBin 接口实现

**用户故事**: 作为仓库管理员，我想通过 GetItemStockByBin 接口查询库存信息，以便于掌握商品在货位上的分布情况

#### 验收标准

1. **WHEN** 传入有效的 Warehouse 参数 **THEN** 系统 **SHALL** 返回该仓库所有商品的库存按货位分组信息，包含 item_code, actual_qty, projected_qty, reserved_qty_for_pos, stock_uom, valuation_rate 字段
2. **IF** 同时传入 item_code 参数 **THEN** 系统 **SHALL** 只返回指定商品的库存信息，包含完整的返回值字段
3. **WHEN** Warehouse 参数为空或无效 **THEN** 系统 **SHALL** 返回参数错误信息
4. **WHEN** 调用 Bin 查询服务失败 **THEN** 系统 **SHALL** 返回相应的错误信息并记录日志

#### 具体要求

- [ ] 1.1 接口参数验证：Warehouse 必填，item_code 可选
- [ ] 1.2 数据返回格式：按货位分组，返回值包含字段 item_code, actual_qty, projected_qty, reserved_qty_for_pos, stock_uom, valuation_rate
- [ ] 1.3 错误处理：参数校验失败时返回明确的错误信息
- [ ] 1.4 服务调用：正确调用已有的 Bin 查询服务获取库存数据
- [ ] 1.5 响应格式：遵循 API 规范的统一响应格式

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 Controller → Service → Logic 分层
- **单一职责原则**: 每个文件应有单一、明确的目的
- **模块化设计**: Service 和 Logic 应独立且可复用
- **依赖管理**: Logic 只能依赖其他 Service 接口
- **遵循规范**:
  - `.cursor/rules/go-bmp.mdc` - Go BMP 微服务规范
  - `.cursor/rules/api.mdc` - API 设计规范
  - `.cursor/rules/structs.mdc` - 项目结构规范

### API 设计要求

- [x] URL 使用 snake_case 命名（如：`/api/v1/get_item_stock_by_bin`）
- [x] data 字段必须是对象，不能是 null 或数组
- [ ] 分页信息统一放在 meta 中
- [x] 响应格式：`{code, message, data{}}`
- [ ] 参考: `.cursor/rules/api.mdc` - API 设计规范

### 数据库设计要求

- [ ] 复用现有的库存和货位相关表结构
- [ ] 查询优化：使用合适的索引提升查询性能
- [ ] 参考: `.cursor/rules/database.mdc` - 数据库开发规范

### 性能要求

- [x] 本地响应时间 < 200ms
- [x] 数据库查询优化（使用索引）
- [x] 缓存策略（Redis，如有必要）
- [ ] 并发处理（使用 UUID 锁，如有必要）

### 测试要求

- [x] Service 层测试覆盖率 ≥ 70%
- [x] Logic 层测试覆盖率 ≥ 70%
- [ ] **Payment/Order 相关模块测试覆盖率 100%**（高风险）
- [x] 集成测试覆盖核心查询流程
- [x] API 测试覆盖所有接口
- [ ] 参考: `.cursor/rules/go-bmp.mdc` - 测试规范

### 安全要求

- [x] 所有 API 需要身份验证
- [ ] 敏感数据加密存储（如涉及）
- [x] SQL 注入防护（使用参数化查询）
- [ ] XSS 防护（前端输入校验，如有前端）
- [ ] CSRF 防护（Token 验证，如有前端）
- [ ] 参考: `.cursor/rules/security.mdc` - 安全开发规范

### 可靠性要求

- [x] 网络异常时优雅降级
- [ ] 事务管理（保证数据一致性，如有写操作）
- [x] 错误日志记录（使用 Logger）
- [x] 故障恢复机制

---

## 验收标准

### 功能验收

1. **接口调用测试**: 使用有效的 Warehouse 参数调用接口，返回包含 item_code, actual_qty, projected_qty, reserved_qty_for_pos, stock_uom, valuation_rate 字段的库存数据
2. **参数验证测试**: 传入空的 Warehouse 参数，接口返回参数错误
3. **可选参数测试**: 同时传入 Warehouse 和 item_code 参数，只返回指定商品的完整库存信息
4. **错误处理测试**: 当 Bin 查询服务异常时，接口返回相应的错误信息

### 测试验收

1. **单元测试**: Logic 和 Service 层测试覆盖率达标
2. **API 测试**: 所有接口参数验证和响应格式测试通过
3. **集成测试**: 端到端查询流程测试通过
4. **异常测试**: 网络异常和参数异常情况下的处理测试通过

### 文档验收

1. **技术文档**: design.md 完整且准确
2. **API 文档**: 接口文档完整
3. **数据库文档**: 查询语句和索引优化说明完整
4. **测试文档**: tasks.md 中的测试任务完成

---

## 约束条件

### 技术约束

#### Go BMP 模块

- 必须使用 GoFrame 2.x
- 禁止修改 dao/entity/do/ 目录（自动生成）
- gRPC 服务必须注册到 Nacos
- 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`

### 业务约束

- 必须在现有的 stock_bin.go 文件中实现，不新增独立的 Logic 服务
- 库存数据必须与现有库存系统保持一致
- 返回值必须包含指定字段：item_code, actual_qty, projected_qty, reserved_qty_for_pos, stock_uom, valuation_rate
- 不允许修改现有的库存数据结构

### 资源约束

- 开发时间: 2 天
- Story Point: 3 (必须 ≤ 5)

---

## 依赖关系

### 技术依赖

- `ttpos-bmp` 框架依赖
- 现有的 Bin 查询服务

### 服务依赖

- **BMP → BMP**: 调用已有的 Bin 查询服务

### 业务依赖

- 依赖现有的库存管理系统
- 依赖现有的货位管理功能
- 前置条件：Bin 查询服务正常运行

---

## 风险和缓解

### 风险 1: Bin 查询服务接口变更

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 充分测试 Bin 查询服务的兼容性
- 编写集成测试验证接口调用
- 与 Bin 查询服务开发团队确认接口稳定性

### 风险 2: 查询性能问题

**影响**: 中  
**概率**: 低  
**缓解措施**:

- 优化数据库查询语句
- 添加必要的索引
- 实施查询结果缓存策略

---

## 时间表

- **Phase 1 - 接口设计与实现**: 1 天
- **Phase 2 - 测试与优化**: 1 天
- **总计**: 2 天（SP = 3）

---

## 参考资料

### 核心规范

- `.cursor/rules/go-bmp.mdc` - Go BMP 开发规范
- `.cursor/rules/api.mdc` - API 设计规范
- `.cursor/rules/database.mdc` - 数据库开发规范
- `.cursor/rules/security.mdc` - 安全开发规范

### 架构文档

- `docs/human/architecture/go-bmp-architecture.md` - Go BMP 架构

### 开发指南

- `docs/human/guides/go-bmp-development.md` - Go BMP 开发指南
- `docs/human/guides/api-design-guide.md` - API 设计指南

### 外部参考

- [BMP 模块现有库存相关接口]

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-12-26  
**作者**: rikugun  
**审核者**: {审核者}
