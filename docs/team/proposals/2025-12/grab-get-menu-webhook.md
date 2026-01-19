# Grab Get Menu Webhook 需求提案

> 本文档用于需求发起阶段，经团队评审后创建正式 Spec。

---

## 📋 提案信息

| 项目       | 内容     |
| ---------- | -------- |
| **提案人** | rikugun   |
| **日期**   | 2025-12-09   |
| **目标版本** | v2.11.0 |
| **状态**   | 待评审   |
| **关联任务** | - |
| **关联 Spec** | [task-takeout-grab-get-menu-webhook](../../../shared/specs/archived/v2.12/task-takeout-grab-get-menu-webhook/requirements.md)      |

---

## 🎯 背景和动机

### 问题描述

当前 `grab_v1_get_menu.go` Controller 中的 `GetMenu` 方法仅返回 `CodeNotImplemented`，未实现实际功能。根据 Grab API 文档要求，当 Grab 平台主动拉取商户菜单数据时，需要：

1. 从数据库中读取对应商户的菜单快照数据
2. 将数据转换为符合 Grab API 格式的响应结构
3. 返回标准的 Grab Menu 格式数据

**当前状态**：
- Controller 方法已定义但未实现
- 已有 `ChannelMenu` 服务可以读取菜单快照
- 已有 `GetMenuResponse` DTO 结构定义
- `MenuService.HandleGetMenu` 仅返回示例数据

### 业务价值

- ✅ **完成 Grab 集成闭环**：实现菜单拉取功能，使 Grab 平台能够主动获取最新菜单数据
- ✅ **支持菜单同步流程**：配合菜单推送和菜单更新通知，形成完整的菜单同步机制
- ✅ **提升集成稳定性**：确保 Grab 平台能够及时获取菜单变更，避免菜单数据不一致
- ✅ **符合 API 规范**：满足 Grab Food Partner API v1.1.3 的 Get Menu Webhook 要求

### 目标用户

- [ ] 收银员
- [ ] 商户管理员
- [x] **Grab 平台系统**（主要用户）
- [ ] 厨房人员
- [ ] 顾客
- [ ] 其他: ________

---

## 💡 解决方案概述

### 方案描述

实现 `GetMenu` Controller 方法，调用 `ChannelMenu` 服务从数据库读取菜单快照，将 JSON 数据解析并转换为 Grab 标准的 `GetMenuResponse` 格式返回。

**实现要点**：
1. 在 Controller 中调用 `service.Grab().HandleGetMenu()` 或直接调用 `ChannelMenu` 服务
2. 根据 `partnerMerchantID` 查询对应的 `shopUUID`
3. 从 `channel_menu_snapshot` 表读取菜单快照 JSON 数据
4. 解析 JSON 并转换为 `grabDto.GetMenuResponse` 结构
5. 验证数据格式符合 Grab API 规范
6. 返回标准响应

### 核心功能点

1. **菜单数据读取**：从 `channel_menu_snapshot` 表读取指定商户的菜单快照
2. **数据格式转换**：将存储的 JSON 数据转换为 Grab API 标准格式
3. **商户映射**：根据 `partnerMerchantID` 映射到内部 `shopUUID`
4. **错误处理**：处理菜单不存在、数据格式错误等异常情况
5. **响应验证**：确保返回的数据结构符合 Grab API 文档要求

### 影响范围

**涉及终端**：
- [ ] POS 收银端
- [ ] Shop 商家管理端
- [ ] KDS 厨显端
- [ ] QDS 排号叫号端
- [ ] Assistant 助手端
- [ ] Tablet 平板端
- [ ] Mobile 扫码端
- [ ] Menu 电子菜单端
- [ ] Member 会员端

**涉及模块**：
- [ ] UI 组件
- [x] **API 接口**（Grab Webhook）
- [x] **数据模型**（菜单快照）
- [x] **业务逻辑**（菜单读取与转换）
- [ ] 第三方集成（Grab API）
- [ ] 其他: ________

---

## 📊 初步评估

### 技术复杂度

- [ ] **低**：纯 UI 调整，无业务逻辑变更
- [x] **中**：需要前后端联调，基础业务逻辑
- [ ] **高**：涉及架构调整、第三方集成、复杂算法

### 工作量预估

[粗略估算，具体 SP 在技术方案评审时确定]

- **预计天数**: 2-3 天
- **预估 SP**: 3（待技术评审确认）

**任务分解**：
1. 实现 Controller 方法（0.5 天）
2. 实现菜单数据读取与转换逻辑（1 天）
3. 商户 ID 映射逻辑（0.5 天）
4. 单元测试与集成测试（1 天）

### 风险识别

**潜在风险**：
1. **菜单数据格式不一致**：存储的 JSON 格式可能与 Grab API 格式不完全匹配
2. **商户映射缺失**：`partnerMerchantID` 可能无法映射到 `shopUUID`
3. **性能问题**：菜单数据量大时，JSON 解析可能影响响应时间
4. **数据完整性**：菜单快照可能不完整或过期

**缓解措施**：
1. 在 `PushGrabMenu` 时确保数据格式正确，统一使用 Grab SDK 结构
2. 建立商户映射表，确保 `partnerMerchantID` 与 `shopUUID` 的映射关系
3. 使用缓存机制，减少数据库查询和 JSON 解析开销
4. 添加数据校验逻辑，确保返回的菜单数据完整且有效

---

## 🔗 相关资源

### 参考需求

- 类似功能: `push-grab-menu-webhook.md`（菜单推送）
- Grab API 文档: https://developer.grab.com/docs/grabfood/api/v1-1-3/#tag/get-menu-webhook

### 相关文档

- Grab Food Partner API 文档: https://developer.grab.com/docs/grabfood/api/v1-1-3/
- 菜单快照存储方案: `takeout-channel-menu-storage.md`
- Grab 集成总览: `takeout-grab-integration.md`

### 相关代码

- Controller: `ttpos-bmp/app/ttpos-takeout/internal/controller/grab/grab_v1_get_menu.go`
- Service: `ttpos-bmp/app/ttpos-takeout/internal/service/channel_menu.go`
- Logic: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab/menu_service.go`
- DTO: `ttpos-bmp/app/ttpos-takeout/internal/model/dto/grab/menu.go`

---

## 🤝 需求评审

### 评审参与人

| 角色         | 姓名   | 签名/日期 |
| ------------ | ------ | --------- |
| 产品经理     | {姓名} |           |
| 技术负责人   | {姓名} |           |
| 开发代表     | {姓名} |           |
| 测试代表     | {姓名} |           |
| UI/UX 设计师 | {姓名} |           |

### 评审结论

- [ ] ✅ **批准**：进入技术方案设计阶段
- [ ] 🔄 **修改后批准**：需补充以下内容
- [ ] ❌ **拒绝**：不符合产品规划或优先级

**评审意见**：

```
[记录评审会议的关键讨论和决策]
```

**下一步行动**：

- [ ] 创建 Spec：`task-takeout-grab-get-menu-webhook`
- [ ] 分配负责人：{姓名}
- [ ] 目标 Sprint：Sprint {N}

---

## 📝 附录

### User Story（初稿）

**作为** Grab 平台系统  
**我想** 通过 Get Menu Webhook 主动拉取商户的最新菜单数据  
**以便于** 确保菜单数据与 POS 系统保持同步，提供准确的菜单信息给用户

### AC 验收标准（初稿）

1. **WHEN** Grab 平台调用 Get Menu Webhook **THEN** 系统 **SHALL** 从数据库读取对应商户的菜单快照并返回
2. **IF** 菜单快照不存在 **THEN** 系统 **SHALL** 返回适当的错误响应（404 或空菜单）
3. **IF** 菜单数据格式不正确 **THEN** 系统 **SHALL** 返回格式错误响应（400）
4. **WHEN** 返回菜单数据 **THEN** 响应结构 **SHALL** 符合 Grab API v1.1.3 规范
5. **WHEN** 查询菜单数据 **THEN** 系统 **SHALL** 根据 `partnerMerchantID` 正确映射到 `shopUUID`

### 技术实现要点

1. **数据读取**：
   - 使用 `service.ChannelMenu().GetChannelMenu(ctx, shopUUID, "grab")` 读取菜单快照
   - 菜单数据以 JSON 字符串形式存储在 `channel_menu_snapshot` 表

2. **数据转换**：
   - 将 JSON 字符串解析为 `grabDto.GetMenuResponse` 结构
   - 确保字段映射正确（merchantID, partnerMerchantID, currency, sellingTimes, categories）

3. **商户映射**：
   - 需要建立 `partnerMerchantID` 到 `shopUUID` 的映射关系
   - 可能需要查询商户配置表或使用缓存

4. **错误处理**：
   - 菜单不存在：返回空菜单或 404
   - JSON 解析失败：返回 500 错误
   - 数据格式验证失败：返回 400 错误

5. **性能优化**：
   - 考虑使用 Redis 缓存菜单数据
   - 菜单数据较大时，考虑分页或压缩

---

**版本**: v1.0.0  
**创建日期**: 2025-12-09  
**维护者**: 产品组 + Scrum Master  
**相关规范**: `.cursor/rules/scrum_story_point.mdc`, `.cursor/rules/specs.mdc`
