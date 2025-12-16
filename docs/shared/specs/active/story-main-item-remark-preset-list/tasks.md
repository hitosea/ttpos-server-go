# 获取单品备注预设选项列表 API 任务分解

> 本文档定义获取单品备注预设选项列表 API 的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-2 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 10  
**已完成**: 10  
**进行中**: -  
**完成率**: 100%

---

## Phase 1: API 实现

### H5 端

- [x] 1.1 实现 H5 端获取单品备注列表 API Handler

  - File: `main/app/api/v1/h5/h5_handler.go`
  - Purpose: 添加 `OrderItemRemarkList` 方法，处理 H5 端获取单品备注列表请求
  - Requirements: 1.3, 1.5-1.10
  - Leverage: 
    - 参考实现: `main/app/api/v1/h5/h5_handler.go` - `OrderRemarkList()` (line 274-293)
    - Service 方法: `main/app/service/other.go` - `GetOrderItemRemarkList()` (line 372-392)
  - Prompt: Role: Go Developer with Gin framework expertise | Task: 在 H5 Handler 中添加 OrderItemRemarkList 方法，参考 OrderRemarkList 的实现方式 | Context: 调用 otherSrv.GetOrderItemRemarkList(ctx)，使用 helper.Success() 返回响应，使用 helper.ErrorWithDetail() 处理错误 | Restrictions: 遵循 .cursor/rules/api.mdc，URL 使用 snake_case，data 必须是对象 | Success: Handler 方法创建成功，响应格式正确，错误处理正确

- [x] 1.2 注册 H5 端路由

  - File: `main/app/api/v1/h5/h5_handler.go`
  - Purpose: 在路由注册处添加新路由
  - Requirements: 1.3
  - Leverage: 现有路由注册: `main/app/api/v1/h5/h5_handler.go` (line ~832)
  - Prompt: Role: Go Developer | Task: 在 H5 路由注册处添加 GET /order/item/remark/list 路由 | Context: 参考整单备注路由注册方式，放在整单备注路由附近 | Restrictions: 遵循现有路由注册规范 | Success: 路由注册成功，路径正确

### 收银机点餐端

- [x] 2.1 实现收银机点餐端获取单品备注列表 API Handler

  - File: `main/app/api/v1/cashier/cashier_instant.go`
  - Purpose: 添加 `OrderItemRemarkList` 方法，处理收银机点餐端获取单品备注列表请求
  - Requirements: 1.1, 1.5-1.10
  - Leverage: 
    - 参考实现: `main/app/api/v1/cashier/cashier_instant.go` - `OrderRemarkList()` (line 470-489)
    - Service 方法: `main/app/service/other.go` - `GetOrderItemRemarkList()` (line 372-392)
  - Prompt: Role: Go Developer with Gin framework expertise | Task: 在 Cashier Instant Handler 中添加 OrderItemRemarkList 方法，参考 OrderRemarkList 的实现方式 | Context: 调用 otherSrv.GetOrderItemRemarkList(ctx)，使用 helper.Success() 返回响应 | Restrictions: 遵循 .cursor/rules/api.mdc | Success: Handler 方法创建成功，响应格式正确

- [x] 2.2 注册收银机点餐端路由

  - File: `main/app/api/v1/cashier/cashier_instant.go`
  - Purpose: 在路由注册处添加新路由
  - Requirements: 1.1
  - Leverage: 现有路由注册: `main/app/api/v1/cashier/cashier_instant.go` (line ~1733)
  - Prompt: Role: Go Developer | Task: 在收银机点餐端路由注册处添加 GET /instant/order/item/remark/list 路由 | Context: 参考整单备注路由注册方式 | Restrictions: 遵循现有路由注册规范 | Success: 路由注册成功，路径正确

### 点餐助手端

- [x] 3.1 实现点餐助手端获取单品备注列表 API Handler

  - File: `main/app/api/v1/assistant/assistant_desk.go`
  - Purpose: 添加 `OrderItemRemarkList` 方法，处理点餐助手端获取单品备注列表请求
  - Requirements: 1.2, 1.5-1.10
  - Leverage: 
    - 参考实现: `main/app/api/v1/assistant/assistant_desk.go` - `OrderRemarkList()` (line 276-295)
    - Service 方法: `main/app/service/other.go` - `GetOrderItemRemarkList()` (line 372-392)
  - Prompt: Role: Go Developer with Gin framework expertise | Task: 在 Assistant Desk Handler 中添加 OrderItemRemarkList 方法，参考 OrderRemarkList 的实现方式 | Context: 调用 otherSrv.GetOrderItemRemarkList(ctx)，使用 helper.Success() 返回响应 | Restrictions: 遵循 .cursor/rules/api.mdc | Success: Handler 方法创建成功，响应格式正确

- [x] 3.2 注册点餐助手端路由

  - File: `main/app/api/v1/assistant/assistant_desk.go`
  - Purpose: 在路由注册处添加新路由
  - Requirements: 1.2
  - Leverage: 现有路由注册: `main/app/api/v1/assistant/assistant_desk.go` (line ~2041)
  - Prompt: Role: Go Developer | Task: 在点餐助手端路由注册处添加 GET /desk/order/item/remark/list 路由 | Context: 参考整单备注路由注册方式 | Restrictions: 遵循现有路由注册规范 | Success: 路由注册成功，路径正确

### 平板端

- [x] 4.1 实现平板端获取单品备注列表 API Handler

  - File: `main/app/api/v1/tablet/tablet_desk.go`
  - Purpose: 添加 `OrderItemRemarkList` 方法，处理平板端获取单品备注列表请求
  - Requirements: 1.4, 1.5-1.10
  - Leverage: 
    - 参考实现: `main/app/api/v1/tablet/tablet_desk.go` - `OrderRemarkList()` (line 320-339)
    - Service 方法: `main/app/service/other.go` - `GetOrderItemRemarkList()` (line 372-392)
  - Prompt: Role: Go Developer with Gin framework expertise | Task: 在 Tablet Desk Handler 中添加 OrderItemRemarkList 方法，参考 OrderRemarkList 的实现方式 | Context: 调用 otherSrv.GetOrderItemRemarkList(ctx)，使用 helper.Success() 返回响应 | Restrictions: 遵循 .cursor/rules/api.mdc | Success: Handler 方法创建成功，响应格式正确

- [x] 4.2 注册平板端路由

  - File: `main/app/api/v1/tablet/tablet_desk.go`
  - Purpose: 在路由注册处添加新路由
  - Requirements: 1.4
  - Leverage: 现有路由注册: `main/app/api/v1/tablet/tablet_desk.go` (line ~414)
  - Prompt: Role: Go Developer | Task: 在平板端路由注册处添加 GET /desk/order/item/remark/list 路由 | Context: 参考整单备注路由注册方式 | Restrictions: 遵循现有路由注册规范 | Success: 路由注册成功，路径正确

---

## Phase 2: 测试和文档

- [ ] 5.1 API 集成测试

  - File: `main/app/api/v1/*/*_test.go` (可选)
  - Purpose: 测试 4 个 API 接口是否正常工作
  - Requirements: 所有功能需求
  - Leverage: 现有测试文件
  - Prompt: Role: QA Engineer specializing in API testing | Task: 为 4 个单品备注列表 API 编写集成测试 | Context: 测试正常场景、空列表场景、错误场景 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 所有 API 测试通过
  - Status: 可选任务，暂不执行

- [x] 5.2 更新 Swagger 文档

  - File: `main/docs/swagger.yaml` (自动生成)
  - Purpose: 确保 Swagger 文档包含新的 API 接口
  - Requirements: 文档要求
  - Leverage: 现有 Swagger 文档
  - Command: `swag init` (如果使用 swag 工具)
  - Success: Swagger 文档已更新，包含 4 个新接口
  - Completed: 2025-12-05

- [x] 5.3 手动测试验证

  - File: -
  - Purpose: 手动测试 4 个 API 接口
  - Requirements: 所有功能需求
  - Test Cases:
    1. H5 端调用 `/h5/order/item/remark/list`，验证返回数据 ✅
    2. 收银机点餐端调用 `/cashier/instant/order/item/remark/list`，验证返回数据 ✅
    3. 点餐助手端调用 `/assistant/desk/order/item/remark/list`，验证返回数据 ✅
    4. 平板端调用 `/tablet/desk/order/item/remark/list`，验证返回数据 ✅
    5. 验证空列表场景（无预设选项时）✅
    6. 验证多语言支持 ✅
    7. 验证已删除的预设选项不出现在列表中 ✅
  - Success: 所有测试用例通过
  - Completed: 2025-12-05

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [x] 所有任务标记为 `[x]`
- [x] Go 代码通过 `go fmt` 和 `go vet`
- [x] 代码风格与现有代码一致
- [x] 所有测试通过（手动测试）

### 功能完整性

- [x] requirements.md 中的所有需求已满足
- [x] design.md 中的设计已实现
- [x] 4 个 API 接口都已实现并注册路由
- [x] 验收标准已达成

### 文档同步

- [x] Swagger 文档已更新（如有新接口）
- [x] 代码注释完整（Swagger 注释）

### 规范遵循

- [x] 遵循 `.cursor/rules/go-main.mdc`
- [x] 遵循 `.cursor/rules/api.mdc`
- [x] URL 使用 snake_case
- [x] data 字段是对象

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-main-item-remark-preset-list/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-main-item-remark-preset-list/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-main-item-remark-preset-list/tasks.md
```

### 执行流程

1. **选择任务**: 按顺序选择下一个未完成任务
2. **阅读需求**: 查看 requirements.md 中的关联需求
3. **查看复用**: 检查 Leverage 中的可复用代码
4. **使用 AI**: 复制 Prompt 模板，让 AI 生成代码
5. **实现代码**: 按照规范实现功能
6. **运行检查**: `go fmt`, `go vet`
7. **标记完成**: 将 `[ ]` 改为 `[x]`
8. **提交代码**: Git commit（参考 `.cursor/rules/version.mdc`）

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`

---

**模板版本**: v1.0.0  
**最后更新**: 2025-12-05  
**维护者**: 后端开发组

