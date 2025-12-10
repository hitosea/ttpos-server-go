# Grab Get Menu Webhook 任务分解

> 本文档定义 Grab Get Menu Webhook 的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **AI 友好**: 提供 Prompt 模板辅助执行

## 📊 进度总览

**总任务数**: 4
**已完成**: 2
**进行中**: -
**完成率**: 50%

---

## Phase 1: 核心实现

- [x] 1.1 实现 HandleGetMenu 逻辑
  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab/menu_service.go`
  - Purpose: 实现从数据库读取菜单快照并转换为 Grab 响应格式
  - Requirements: 1.2, 1.3, 1.4, 1.5
  - Leverage: `service.ChannelMenu` (已存在), `grabDto` (已存在)
  - Prompt: Role: Go Developer | Task: 实现 MenuService.HandleGetMenu 方法 | Context: 1. 解析 partnerMerchantID 获取 shopUUID (假设为 ID) 2. 调用 service.ChannelMenu().GetChannelMenu 获取 JSON 3. 解析 JSON 为 PushGrabMenuDTO 4. 转换为 GetMenuResponse 返回 | Restrictions: 处理错误情况(NotFound, UnmarshalError) | Success: 逻辑实现完整，数据转换正确

- [x] 1.2 更新 GetMenu Controller
  - File: `ttpos-bmp/app/ttpos-takeout/internal/controller/grab/grab_v1_get_menu.go`
  - Purpose: 调用 Logic 层处理请求
  - Requirements: 1.1, 1.6
  - Leverage: `service.Grab`
  - Prompt: Role: Go Developer | Task: 实现 GetMenu Controller 方法 | Context: 调用 service.Grab().HandleGetMenu | Restrictions: 移除 CodeNotImplemented，正确处理返回结果 | Success: Controller 能够调用 Logic 并返回结果

---

## Phase 2: 测试与优化

- [ ] 2.1 编写 Logic 单元测试
  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab/menu_service_test.go`
  - Purpose: 验证 HandleGetMenu 逻辑
  - Requirements: 测试验收 1
  - Leverage: `gtest`
  - Prompt: Role: QA Engineer | Task: 编写 HandleGetMenu 单元测试 | Context: Mock ChannelMenu Service，测试正常、NotFound、JSON Error 场景 | Restrictions: 覆盖核心路径 | Success: 测试通过

- [ ] 2.2 验证集成测试 (可选)
  - File: -
  - Purpose: 手动验证接口
  - Requirements: 测试验收 2
  - Success: 使用 Postman 或 curl 调用接口返回预期结果

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/rikugun/2025-12/2025-12-09.md`

---

**模板版本**: v1.0.0
**最后更新**: 2025-12-09
**维护者**: rikugun
