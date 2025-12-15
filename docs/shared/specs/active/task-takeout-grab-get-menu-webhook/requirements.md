# Grab Get Menu Webhook 需求文档

> 本文档定义 Grab Get Menu Webhook 的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-12/grab-get-menu-webhook.md](../../../../team/proposals/2025-12/grab-get-menu-webhook.md) |
| **创建日期**      | 2025-12-09                                                                                                 |
| **负责人**        | rikugun                                                                                                       |
| **目标 Sprint**   | Sprint {N}                                                                                                   |
| **涉及技术栈**    | [ ] Go (main/) [x] Go (ttpos-bmp/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                   |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 已通过 |
| **审核人**   | {审核人姓名}             |
| **审核日期** | {YYYY-MM-DD}             |
| **审核意见** | {审核意见，如有}         |

---

## 📋 概述

当前 `grab_v1_get_menu.go` Controller 中的 `GetMenu` 方法仅返回 `CodeNotImplemented`，未实现实际功能。根据 Grab API 文档要求，当 Grab 平台主动拉取商户菜单数据时，系统需要从数据库中读取对应商户的菜单快照数据，将数据转换为符合 Grab API 格式的响应结构，并返回标准的 Grab Menu 格式数据。

本功能将实现完整的菜单拉取流程，确保 Grab 平台能够主动获取最新菜单数据，配合菜单推送和菜单更新通知，形成完整的菜单同步机制。

## 🎯 产品对齐

- **完成 Grab 集成闭环**：实现菜单拉取功能，使 Grab 平台能够主动获取最新菜单数据
- **提升集成稳定性**：确保 Grab 平台能够及时获取菜单变更，避免菜单数据不一致
- **符合 API 规范**：满足 Grab Food Partner API v1.1.3 的 Get Menu Webhook 要求

## 📝 用户故事

**作为** Grab 平台系统  
**我想** 通过 Get Menu Webhook 主动拉取商户的最新菜单数据  
**以便于** 确保菜单数据与 POS 系统保持同步，提供准确的菜单信息给用户

---

## 功能需求

### Requirement 1: 实现 Get Menu Webhook

**用户故事**: 作为 Grab 平台系统，我想通过 Get Menu Webhook 主动拉取商户的最新菜单数据，以便于确保菜单数据与 POS 系统保持同步。

#### 验收标准

1. **WHEN** Grab 平台调用 Get Menu Webhook **THEN** 系统 **SHALL** 从数据库读取对应商户的菜单快照并返回
2. **IF** 菜单快照不存在 **THEN** 系统 **SHALL** 返回适当的错误响应（404 或空菜单，需确认 Grab 规范）
3. **IF** 菜单数据格式不正确 **THEN** 系统 **SHALL** 返回格式错误响应（400）
4. **WHEN** 返回菜单数据 **THEN** 响应结构 **SHALL** 符合 Grab API v1.1.3 规范
5. **WHEN** 查询菜单数据 **THEN** 系统 **SHALL** 根据 `partnerMerchantID` 正确映射到 `shopUUID`

#### 具体要求

- [ ] 1.1 在 Controller 中调用 `service.Grab().HandleGetMenu()` 或直接调用 `ChannelMenu` 服务
- [ ] 1.2 根据 `partnerMerchantID` 查询对应的 `shopUUID`（需实现商户映射逻辑）
- [ ] 1.3 使用 `service.ChannelMenu().GetChannelMenu()` 从 `channel_menu_snapshot` 表读取菜单快照 JSON 数据
- [ ] 1.4 解析 JSON 并转换为 `grabDto.GetMenuResponse` 结构
- [ ] 1.5 验证数据格式符合 Grab API 规范
- [ ] 1.6 错误处理：处理菜单不存在、JSON 解析失败等异常情况

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 Controller → Service → Repository 分层
- **单一职责原则**: 每个文件应有单一、明确的目的
- **遵循规范**:
  - `.cursor/rules/go-bmp.mdc` - Go BMP 微服务规范
  - `.cursor/rules/structs.mdc` - 项目结构规范

### API 设计要求

- [ ] 响应格式必须严格符合 Grab API v1.1.3 文档要求
- [ ] 确保所有必填字段都包含在响应中
- [ ] 参考: `.cursor/rules/api.mdc` - API 设计规范

### 数据库设计要求

- [ ] 复用现有的 `channel_menu_snapshot` 表，无需新增表
- [ ] 读取时应使用索引（`shop_uuid`, `provider_name`）
- [ ] 参考: `.cursor/rules/database.mdc` - 数据库开发规范

### 性能要求

- [ ] 接口响应时间 < 500ms
- [ ] 菜单数据较大时，JSON 解析性能需关注

### 安全要求

- [ ] 必须验证 Grab 签名 (X-Grab-Signature)
- [ ] 验证请求的时间戳有效性 (X-Grab-Timestamp)
- [ ] 参考: `.cursor/rules/security.mdc` - 安全开发规范

---

## 验收标准

### 功能验收

1. **正常拉取**: 发送带有有效签名的 Get Menu 请求，能够返回正确的菜单数据
2. **商户映射**: 使用正确的 `partnerMerchantID` 能够找到对应的菜单快照
3. **数据格式**: 返回的 JSON 数据结构完全符合 Grab 文档定义
4. **异常处理**: 当菜单不存在时，返回符合 Grab 预期的响应

### 测试验收

1. **单元测试**: 覆盖 Controller 和 Service 逻辑
2. **集成测试**: 模拟 Grab 请求，验证端到端流程

### 文档验收

1. **技术文档**: design.md 完整且准确
2. **API 文档**: 接口文档更新（如需）

---

## 约束条件

### 技术约束

#### Go BMP 模块

- 必须使用 GoFrame 2.x
- 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`

### 业务约束

- 必须支持 Grab 规定的菜单结构（Selling Times, Categories, Items, Modifiers）

### 资源约束

- 开发时间: 2-3 天
- Story Point: 3

---

## 依赖关系

### 技术依赖

- `grabfood-api-sdk-go` - 用于数据结构定义

### 服务依赖

- **Channel Menu Service**: 依赖 `channel_menu_snapshot` 数据的存储逻辑

### 业务依赖

- 前置条件：必须先有菜单推送（Push Menu）逻辑将菜单快照保存到数据库，否则 Get Menu 读取不到数据

---

## 风险和缓解

### 风险 1: 菜单数据格式不一致

**影响**: 高  
**概率**: 中  
**缓解措施**:
- 在 `PushGrabMenu` 时确保数据格式正确，统一使用 Grab SDK 结构
- 在 `GetMenu` 时增加数据校验逻辑

### 风险 2: 商户映射缺失

**影响**: 高  
**概率**: 低  
**缓解措施**:
- 确保 `partnerMerchantID` 与 `shopUUID` 的映射关系可靠
- 如果映射失败，返回明确的错误信息

---

## 时间表

- **Phase 1 - 核心逻辑**: 实现 Controller 和 Service 逻辑 (1.5 天)
- **Phase 2 - 测试与优化**: 单元测试、集成测试、文档 (1 天)
- **总计**: 2.5 天（SP = 3）

---

## 参考资料

### 核心规范

- `.cursor/rules/go-bmp.mdc` - Go BMP 开发规范
- `.cursor/rules/api.mdc` - API 设计规范

### 开发指南

- `docs/human/guides/go-bmp-development.md` - Go BMP 开发指南

### 外部参考

- [Grab Get Menu Webhook 文档](https://developer.grab.com/docs/grabfood/api/v1-1-3/#tag/get-menu-webhook)

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/rikugun/2025-12/2025-12-09.md`

---

**版本**: v1.0.0  
**创建日期**: 2025-12-09  
**作者**: rikugun  
**审核者**: {审核者}
