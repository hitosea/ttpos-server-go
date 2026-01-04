> ⚠️ **已归档** - 此 Spec 已随 v2.12 发布。
>
> - 归档时间: 2026-01-04
> - 归档人: weifashi

# Grab 提交订单 API 重构 需求文档

> 本文档定义 Grab 提交订单 API 重构 的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-12/v2.12.0-grab-submit-order-api-refactor.md](../../../../team/proposals/2025-12/v2.12.0-grab-submit-order-api-refactor.md) |
| **创建日期**      | 2025-12-19                                                                                                 |
| **负责人**        | rikugun                                                                                                       |
| **目标 Sprint**   | Sprint 当前                                                                                                   |
| **涉及技术栈**    | [ ] Go (main/) [x] Go (ttpos-bmp/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                   |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 待审核 / 已通过 / 需修改 |
| **审核人**   | {审核人姓名}             |
| **审核日期** | {YYYY-MM-DD}             |
| **审核意见** | {审核意见，如有}         |

---

## 📋 概述

重构 Grab 提交订单的 API 结构，将空的 `SubmitOrderReq` 结构体修改为嵌入 `*grabfood.SubmitOrderRequest`，提升代码类型安全性和可维护性，并优化 `shopUuid` 获取逻辑。

该功能主要提升代码质量，不涉及用户界面变更，对现有 Grab 订单功能保持完全兼容。

## 🎯 产品对齐

该功能支持产品技术债务清理的目标，提升代码质量和可维护性，为后续 Grab 集成功能扩展提供更好的技术基础。

## 📝 用户故事

**作为** 开发人员  
**我想** 重构 Grab 提交订单 API 结构  
**以便于** 提升代码类型安全和可维护性

---

## 功能需求

### Requirement 1: API 结构体重构

**用户故事**: 作为开发人员，我想将 SubmitOrderReq 结构体嵌入 *grabfood.SubmitOrderRequest，以便于提升类型安全性和代码可维护性

#### 验收标准

1. **WHEN** API 接收 Grab 订单提交请求 **THEN** 系统 **SHALL** 使用类型化的 `SubmitOrderReq` 结构体
2. **WHEN** 结构体重构完成后 **THEN** 编译 **SHALL** 通过且无类型错误

#### 具体要求

- [x] 1.1 修改 `ttpos-bmp/app/ttpos-takeout/api/grab/v1/submit_order.go` 中的 `SubmitOrderReq` 结构体
- [x] 1.2 将结构体嵌入 `*grabfood.SubmitOrderRequest` 类型
- [x] 1.3 保留原有的 Meta 信息配置

---

### Requirement 2: Logic 层调整

**用户故事**: 作为开发人员，我想调整 Logic 层直接使用 API 层的类型化请求对象，以便于简化数据流和提升代码清晰度

#### 验收标准

1. **WHEN** Logic 层接收请求 **THEN** 系统 **SHALL** 直接使用类型化的请求对象而非重新解析 JSON
2. **WHEN** 重构完成后 **THEN** 现有业务逻辑 **SHALL** 保持不变

#### 具体要求

- [x] 2.1 修改 `HandleSubmitOrder` 方法签名，接收 `*grabfood.SubmitOrderRequest` 参数
- [x] 2.2 移除重复的 JSON 解析逻辑
- [x] 2.3 调整 `saveOrderFromSDK` 方法调用方式

---

### Requirement 3: ShopUuid 获取优化

**用户故事**: 作为开发人员，我想优化 shopUuid 获取逻辑优先从请求对象的 partnerMerchantID 字段获取，以便于提升查询效率

#### 验收标准

1. **IF** 请求包含有效的 `partnerMerchantID` **THEN** 系统 **SHALL** 优先使用该字段获取 `shopUuid`
2. **IF** `partnerMerchantID` 为空或无效 **THEN** 系统 **SHALL** 回退到原有配置查询逻辑

#### 具体要求

- [x] 3.1 在 `saveOrderFromSDK` 方法中优先使用 `req.GetPartnerMerchantID()` 获取不到时使用 `req.GetMerchantID()` 查询 `ShopProviderCfg`
- [x] 3.2 保留原有配置查询逻辑作为 fallback
- [x] 3.3 添加相应的日志记录用于调试

---

### Requirement 4: 向后兼容性保证

**用户故事**: 作为开发人员，我想确保重构后保持完全的向后兼容性，以便于现有功能不受影响

#### 验收标准

1. **WHEN** 重构完成后 **THEN** 现有 Grab 订单功能 **SHALL** 保持完全兼容
2. **WHEN** 运行现有测试 **THEN** 所有测试 **SHALL** 通过

#### 具体要求

- [x] 4.1 确保 API 接口签名保持不变（仍接收 []byte 参数）
- [x] 4.2 保留原有的错误处理逻辑
- [x] 4.3 运行并通过现有单元测试

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

- [x] URL 使用 snake_case 命名（如：`/partner/orders`）
- [x] data 字段必须是对象，不能是 null 或数组
- [x] 分页信息统一放在 meta 中（本功能不涉及分页）
- [x] 响应格式：`{code, message, data{}}`
- [x] 参考: `.cursor/rules/api.mdc` - API 设计规范

### 数据库设计要求

- [x] 本功能不涉及数据库变更
- [x] 参考: `.cursor/rules/database.mdc` - 数据库开发规范

### 性能要求

- [x] 本地响应时间 < 200ms（重构不影响性能）
- [x] 数据库查询优化（使用索引）- 优化 shopUuid 查询逻辑
- [x] 缓存策略（Redis）- 如有需要
- [x] 并发处理（使用 UUID 锁）- 订单处理已考虑并发

### 浏览器兼容性（管理后台）

- [x] 本功能为后端 API，不涉及前端兼容性

### 测试要求

- [x] Service 层测试覆盖率 ≥ 70%
- [x] Repository 层测试覆盖率 ≥ 80%（如有变更）
- [x] **Payment/Order 相关模块测试覆盖率 100%**（高风险）- Grab 订单处理为高风险模块
- [x] 集成测试覆盖核心流程
- [x] API 测试覆盖所有接口
- [x] 参考: `.cursor/rules/go-bmp.mdc` - 测试规范

### 国际化要求

- [x] 本功能为后端 API，不涉及国际化

### 安全要求

- [x] 所有 API 需要身份验证（Webhook 已由中间件处理）
- [x] 敏感数据加密存储
- [x] SQL 注入防护（使用参数化查询）
- [x] XSS 防护（前端输入校验）- 不涉及前端
- [x] CSRF 防护（Token 验证）- Webhook 不涉及
- [x] 参考: `.cursor/rules/security.mdc` - 安全开发规范

### 可靠性要求

- [x] 网络异常时优雅降级
- [x] 事务管理（保证数据一致性）- 已有事务处理
- [x] 错误日志记录（使用 Logger）- 已使用 g.Log()
- [x] 故障恢复机制

---

## 验收标准

### 功能验收

1. **API 结构体重构**: SubmitOrderReq 成功嵌入 *grabfood.SubmitOrderRequest，编译通过
2. **Logic 层调整**: HandleSubmitOrder 方法成功使用类型化对象，业务逻辑保持不变
3. **ShopUuid 优化**: partnerMerchantID 优先级高于配置查询，fallback 逻辑正常工作
4. **向后兼容性**: 现有 Grab 订单功能完全兼容，所有测试通过

### 测试验收

1. **单元测试**: Service 层测试覆盖率 ≥ 70%，核心逻辑测试覆盖率 100%
2. **API 测试**: Grab 提交订单接口测试通过
3. **集成测试**: 端到端订单提交流程测试通过
4. **手动测试**: 浏览器兼容性测试通过（不涉及前端）

### 文档验收

1. **技术文档**: design.md 完整且准确
2. **API 文档**: API 接口文档完整（如有变更）
3. **数据库文档**: 无数据库变更
4. **测试文档**: tasks.md 中的测试任务完成

---

## 约束条件

### 技术约束

#### Go BMP 模块

- [x] 必须使用 GoFrame 2.x
- [x] 禁止修改 dao/entity/do/ 目录（自动生成）
- [x] gRPC 服务必须注册到 Nacos（不涉及）
- [x] 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`

### 业务约束

- [x] 必须保持现有 Grab 订单功能的完整性
- [x] 重构仅涉及代码结构优化，不改变业务逻辑
- [x] ShopUuid 获取优化不能影响现有商户配置

### 资源约束

- 开发时间: 1 天
- Story Point: 2 (必须 ≤ 5)

---

## 依赖关系

### 技术依赖

- `github.com/gogf/gf/v2` - GoFrame 2.x 框架
- `github.com/grab/grabfood-api-sdk-go` - Grab SDK（现有依赖）
- `ttpos-bmp/app/ttpos-takeout/internal/service` - 内部服务

### 服务依赖

- **BMP 内部服务**: 商户配置服务（ShopProviderCfg）

### 业务依赖

- [x] Grab SDK 集成（现有）
- [x] 商户配置系统（现有）
- [x] 订单处理流程（现有）

---

## 风险和缓解

### 风险 1: SDK 类型变更导致兼容性问题

**影响**: 中  
**概率**: 低  
**缓解措施**:

- 充分测试 SDK 类型的兼容性
- 在测试环境中验证所有 Grab 订单流程
- 准备回滚方案

### 风险 2: ShopUuid 获取逻辑变更影响现有配置

**影响**: 中  
**概率**: 低  
**缓解措施**:

- 保留原有配置查询逻辑作为 fallback
- 添加详细日志记录变更行为
- 在测试环境中验证各种配置场景

---

## 时间表

- **Phase 1 - API 层重构**: 2 小时
- **Phase 2 - Logic 层调整**: 3 小时
- **Phase 3 - 测试验证**: 3 小时
- **总计**: 1 天（SP = 2）

---

## 参考资料

### 核心规范

- `.cursor/rules/go-bmp.mdc` - Go BMP 开发规范
- `.cursor/rules/api.mdc` - API 设计规范
- `.cursor/rules/security.mdc` - 安全开发规范

### 架构文档

- `docs/human/architecture/go-bmp-architecture.md` - Go BMP 架构

### 开发指南

- `docs/human/guides/go-bmp-development.md` - Go BMP 开发指南
- `docs/human/guides/api-design-guide.md` - API 设计指南

### 外部参考

- Grab Food API SDK 文档

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/2025-12/2025-12-19.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-12-19  
**作者**: rikugun  
**审核者**: {审核者}
