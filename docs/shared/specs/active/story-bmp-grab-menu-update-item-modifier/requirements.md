# GrabFood 菜单项和修饰符更新功能 需求文档

> 本文档定义 GrabFood 菜单项和修饰符更新功能 的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-12/story-bmp-grab-menu-update-item-modifier.md](../../../../team/proposals/2025-12/story-bmp-grab-menu-update-item-modifier.md) |
| **创建日期**      | 2025-12-15                                                                                                 |
| **负责人**        | rikugun                                                                                                       |
| **目标 Sprint**   | Sprint 25                                                                                                   |
| **涉及技术栈**    | [ ] Go (main/) [x] Go (ttpos-bmp/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                   |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 已通过                   |
| **审核人**   | rikugun                  |
| **审核日期** | 2025-12-15               |
| **审核意见** | 审核通过，开始技术设计  |

---

## 📋 概述

在现有的 `ttpos-bmp/app/ttpos-takeout/internal/logic/grab_menu/grab_menu.go` 文件中增加对接 GrabFood API `update-menu-record` 接口的实现，将商品更新和修饰符更新分为两个独立方法，实现与 GrabFood 平台的实时菜单同步，确保外卖业务的运营效率和数据一致性。

## 🎯 产品对齐

该功能支持 TTPOS 作为现代化餐饮收银系统的愿景，提升外卖业务的集成深度和运营效率，增强与 GrabFood 平台的合作关系，为商户提供更优质的数字化服务体验。

## 📝 用户故事

**作为** 商户管理员
**我想** 在 TTPOS 中更新商品或修饰符信息时，自动同步到 GrabFood 平台
**以便于** 确保外卖菜单信息的实时一致性

---

## 功能需求

### Requirement 1: 商品信息更新功能

**用户故事**: 作为商户管理员，我想更新商品的价格、状态、库存等信息，以便于这些变更能实时同步到 GrabFood 平台

#### 验收标准

1. **WHEN** 商户在 TTPOS 中修改商品价格 **THEN** 系统 **SHALL** 调用 GrabFood API 更新对应商品价格
2. **WHEN** 商户在 TTPOS 中修改商品可用状态 **THEN** 系统 **SHALL** 调用 GrabFood API 更新对应商品状态
3. **WHEN** 商户在 TTPOS 中修改商品库存 **THEN** 系统 **SHALL** 调用 GrabFood API 更新对应商品库存
4. **IF** GrabFood API 调用失败 **THEN** 系统 **SHALL** 记录错误日志并支持重试机制

#### 具体要求

- [ ] 1.1 支持商品价格更新（包含税前/税后价格）
- [ ] 1.2 支持商品可用状态变更（可用/不可用）
- [ ] 1.3 支持商品库存数量更新
- [ ] 1.4 支持商品描述信息更新
- [ ] 1.5 支持商品图片更新
- [ ] 1.6 实现异步更新机制，不阻塞主业务流程

---

### Requirement 2: 修饰符选项更新功能

**用户故事**: 作为商户管理员，我想更新修饰符的价格、可用状态等信息，以便于这些变更能实时同步到 GrabFood 平台

#### 验收标准

1. **WHEN** 商户在 TTPOS 中修改修饰符价格 **THEN** 系统 **SHALL** 调用 GrabFood API 更新对应修饰符价格
2. **WHEN** 商户在 TTPOS 中修改修饰符可用状态 **THEN** 系统 **SHALL** 调用 GrabFood API 更新对应修饰符状态
3. **IF** GrabFood API 调用失败 **THEN** 系统 **SHALL** 记录错误日志并支持重试机制

#### 具体要求

- [ ] 2.1 支持修饰符价格更新
- [ ] 2.2 支持修饰符可用状态变更
- [ ] 2.3 支持修饰符选项名称更新
- [ ] 2.4 支持修饰符选项描述更新
- [ ] 2.5 实现异步更新机制，不阻塞主业务流程

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 Controller → Service → Repository 分层
- **单一职责原则**: 每个文件应有单一、明确的目的
- **模块化设计**: Service 和 Repository 应独立且可复用
- **依赖管理**: Service 只能依赖其他 Service 接口，不能直接依赖 Repository
- **遵循规范**:
  - `.cursor/rules/go-bmp.mdc` - Go BMP 微服务规范
  - `ttpos-bmp/.cursor/rules/go-rules.mdc` - Go BMP 开发规范
  - `.cursor/rules/api.mdc` - API 设计规范
  - `.cursor/rules/database.mdc` - 数据库开发规范
  - `.cursor/rules/security.mdc` - 安全开发规范

### API 设计要求

- [ ] URL 使用 snake_case 命名（如：`/api/v1/order_info`）
- [ ] data 字段必须是对象，不能是 null 或数组
- [ ] 分页信息统一放在 meta 中
- [ ] 响应格式：`{code, message, data{}}`
- [ ] 参考: `.cursor/rules/api.mdc` - API 设计规范

### 数据库设计要求

- [ ] 必须包含: `id`, `uuid`, `create_time`, `update_time`, `delete_time`
- [ ] 时间字段使用 int 类型，\_time 结尾，默认值 0
- [ ] 金额字段使用 decimal(20,8)
- [ ] UUID 字段使用 bigint unsigned
- [ ] 表名使用 ttpos\_ 前缀
- [ ] 字段名使用 snake_case
- [ ] 参考: `.cursor/rules/database.mdc` - 数据库开发规范

### 性能要求

- [ ] 本地响应时间 < 200ms
- [ ] 数据库查询优化（使用索引）
- [ ] 缓存策略（Redis）
- [ ] 并发处理（使用 UUID 锁）

### 测试要求

- [ ] Service 层测试覆盖率 ≥ 70%
- [ ] Repository 层测试覆盖率 ≥ 80%
- [ ] **Payment/Order 相关模块测试覆盖率 100%**（高风险）
- [ ] 集成测试覆盖核心流程
- [ ] API 测试覆盖所有接口
- [ ] 参考: `.cursor/rules/go-bmp.mdc` - 测试规范

### 安全要求

- [ ] 所有 API 需要身份验证
- [ ] 敏感数据加密存储
- [ ] SQL 注入防护（使用参数化查询）
- [ ] 参考: `.cursor/rules/security.mdc` - 安全开发规范

### 可靠性要求

- [ ] 网络异常时优雅降级
- [ ] 事务管理（保证数据一致性）
- [ ] 错误日志记录（使用 Logger）
- [ ] 故障恢复机制

---

## 验收标准

### 功能验收

1. **商品更新功能**: 能够成功调用 GrabFood API 更新商品信息并返回正确响应
2. **修饰符更新功能**: 能够成功调用 GrabFood API 更新修饰符信息并返回正确响应
3. **错误处理**: API 调用失败时能正确记录日志并支持重试机制
4. **数据一致性**: 更新操作不会影响现有菜单数据的完整性

### 测试验收

1. **单元测试**: 覆盖率达标
2. **API 测试**: 所有接口测试通过
3. **集成测试**: 端到端流程测试通过

### 文档验收

1. **技术文档**: design.md 完整且准确
2. **API 文档**: API 接口文档完整（如有）
3. **数据库文档**: 迁移脚本和表结构文档完整
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

- 更新操作必须基于 GrabFood API v1.1.3 的 update-menu-record 接口规范
- 必须区分商品更新和修饰符更新两种场景
- 更新操作不应影响 GrabFood 平台的正常运营

### 资源约束

- 开发时间: 5 天
- Story Point: 8 (必须 ≤ 5) - **注意**: 当前 SP 超过 5，建议拆分为多个 Story

---

## 依赖关系

### 技术依赖

- `github.com/grab/grabfood-api-sdk-go` - GrabFood API SDK
- `github.com/gogf/gf/v2` - GoFrame 2.x 框架
- `ttpos-bmp/app/ttpos-takeout/internal/service` - 内部服务依赖

### 服务依赖

- **BMP → GrabFood**: HTTP API 调用 update-menu-record 接口

### 业务依赖

- 现有的菜单获取和同步功能
- GrabFood 商户认证和授权机制

---

## 风险和缓解

### 风险 1: GrabFood API 接口变更

**影响**: 高
**概率**: 中
**缓解措施**:

- 基于官方 SDK 开发，确保 API 兼容性
- 实施监控和告警机制，及时发现接口变更
- 保持与 GrabFood 技术支持的沟通渠道

### 风险 2: 网络超时导致同步失败

**影响**: 中
**概率**: 高
**缓解措施**:

- 实现重试机制和指数退避算法
- 添加超时控制和熔断机制
- 实现异步队列处理，降低对主业务的影响

### 风险 3: 并发更新导致数据不一致

**影响**: 中
**概率**: 中
**缓解措施**:

- 使用分布式锁机制确保操作的原子性
- 实现乐观锁或版本控制机制
- 添加数据校验和回滚机制

---

## 时间表

- **Phase 1 - 需求分析和技术设计**: 1 天
- **Phase 2 - 核心功能开发**: 3 天
- **Phase 3 - 测试和文档**: 1 天
- **总计**: 5 天（SP = 8）

---

## 参考资料

### 核心规范

- `.cursor/rules/go-bmp.mdc` - Go BMP 开发规范
- `ttpos-bmp/.cursor/rules/go-rules.mdc` - Go BMP 开发规范
- `.cursor/rules/api.mdc` - API 设计规范
- `.cursor/rules/database.mdc` - 数据库开发规范
- `.cursor/rules/security.mdc` - 安全开发规范

### 架构文档

- `docs/human/architecture/go-bmp-architecture.md` - Go BMP 架构

### 开发指南

- `docs/human/guides/go-bmp-development.md` - Go BMP 开发指南
- `docs/human/guides/api-design-guide.md` - API 设计指南

### 外部参考

- https://developer.grab.com/docs/grabfood/api/v1-1-3/#tag/update-menu-record/operation/update-menu

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0
**创建日期**: 2025-12-15
**作者**: rikugun
**审核者**: {审核者}
