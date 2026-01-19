# 新管理端-调拨单-优化 任务分解

> 本文档定义调拨单管理界面优化功能的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 17  
**已完成**: 0  
**进行中**: -  
**完成率**: 0%

---

## Phase 1: 数据库优化（1 天）

- [ ] 1.1 创建索引迁移文件

  - File: `admin/database/migrations/{YYYYMMDDHHMMSS}_add_transfer_order_role_type_indexes.php`
  - Purpose: 优化调拨单列表查询性能
  - Requirements: Requirement 2, 3, 4（我发出、我接收、我审核视图）
  - Leverage: 现有迁移文件: `admin/database/migrations/`，参考: `docs/shared/specs/active/story-admin-transfer-order-optimization/design.md`
  - Prompt: Role: Database Engineer | Task: 创建索引迁移文件，优化调拨单列表查询性能 | Context: 为 transfer_out_shop_uuid, transfer_in_shop_uuid, current_approval_node_shop_uuid 字段添加复合索引（包含 status 字段） | Restrictions: 遵循 .cursor/rules/database.mdc，迁移前检查索引是否存在 | Success: 迁移文件创建成功，索引定义正确

- [ ] 1.2 执行数据库迁移

  - File: -
  - Purpose: 在数据库中创建索引
  - Requirements: Requirement 2, 3, 4
  - Leverage: Task 1.1 的迁移文件
  - Command: `cd admin && php think migrate:run`
  - Success: 迁移执行成功，索引已创建

---

## Phase 2: 后端实现（2 天）

### DTO 层

- [ ] 2.1 扩展 TransferOrderListReq DTO

  - File: `main/app/dto/req/transfer_order.go`
  - Purpose: 新增 RoleType 参数，支持角色视角筛选
  - Requirements: Requirement 1（类型筛选器）
  - Leverage: 现有 DTO: `main/app/dto/req/transfer_order.go`
  - Prompt: Role: Go Developer | Task: 在 TransferOrderListReq 结构体中新增 RoleType 字段 | Context: `RoleType string json:"role_type"` // 角色类型（sender/receiver/approver） | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: DTO 扩展成功，字段定义正确

### Repository 层

- [ ] 2.2 新增 Repository 选项方法

  - File: `main/app/repository/transfer_order.go`
  - Purpose: 新增审核节点筛选选项方法
  - Requirements: Requirement 4（我审核视图）
  - Leverage: 现有 Repository: `main/app/repository/transfer_order.go`，使用选项模式
  - Prompt: Role: Go Developer with GORM expertise | Task: 在 TransferOrderRepo 中新增 WhereCurrentApprovalNodeShopUuid() 选项方法 | Context: 筛选当前审核节点为指定门店的单据，使用选项模式返回 DBOption | Restrictions: 遵循 .cursor/rules/go-main.mdc，使用 GORM | Success: 选项方法新增成功，逻辑正确

- [ ] 2.3 编写 Repository 单元测试

  - File: `main/app/repository/transfer_order_test.go`
  - Purpose: 测试新增的选项方法
  - Requirements: Requirement 4
  - Leverage: 现有测试: `main/app/repository/*_repo_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 WhereCurrentApprovalNodeShopUuid() 方法编写单元测试 | Context: 测试筛选逻辑是否正确 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试编写完成，测试通过

### Service 层

- [ ] 2.4 新增角色视角列表查询方法

  - File: `main/app/service/transfer_order/transfer_order.go`
  - Purpose: 实现根据角色类型筛选调拨单的业务逻辑
  - Requirements: Requirement 1, 2, 3, 4（所有角色视图）
  - Leverage: 现有 Service: `main/app/service/transfer_order/transfer_order.go`，Task 2.2 的 Repository 选项方法
  - Prompt: Role: Go Developer with business logic expertise | Task: 新增 GetListByRoleType() 方法，根据 role_type 参数添加不同的筛选条件 | Context: sender=调出方筛选, receiver=调入方筛选, approver=审核方筛选+强制待审核状态 | Restrictions: 遵循 .cursor/rules/go-main.mdc，不使用 panic，返回 error | Success: 方法实现完整，业务逻辑正确

- [ ] 2.5 编写 Service 单元测试

  - File: `main/app/service/transfer_order/transfer_order_test.go`
  - Purpose: 测试角色视角列表查询逻辑
  - Requirements: Requirement 1, 2, 3, 4
  - Leverage: 现有测试: `main/app/service/transfer_order/*_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 GetListByRoleType() 方法编写单元测试，覆盖三种角色视角 | Context: 测试 sender, receiver, approver 三种场景，测试边界情况 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 ≥ 70%，所有测试通过

### API 层

- [ ] 2.6 扩展列表接口

  - File: `main/app/api/v1/shop/shop_transfer.go`
  - Purpose: 在现有列表接口中支持 role_type 参数
  - Requirements: Requirement 1
  - Leverage: 现有 API: `main/app/api/v1/shop/shop_transfer.go`，Task 2.4 的 Service 方法
  - Prompt: Role: Go Developer with Gin framework expertise | Task: 修改 List() 方法，支持 role_type 参数，调用 GetListByRoleType() 方法 | Context: URL 使用 snake_case，使用 helper.SuccessWithData() 返回响应 | Restrictions: 遵循 .cursor/rules/api.mdc | Success: 接口扩展成功，响应格式正确

- [ ] 2.7 编写 API 集成测试

  - File: `main/app/api/v1/shop/shop_transfer_test.go`
  - Purpose: 测试扩展后的列表接口
  - Requirements: Requirement 1, 2, 3, 4
  - Leverage: 现有测试: `main/app/api/v1/shop/*_test.go`
  - Prompt: Role: QA Engineer specializing in API testing | Task: 为 List() 接口编写集成测试，测试 role_type 参数 | Context: 测试三种角色视角，测试参数验证，测试响应格式 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 所有 API 测试通过

---

## Phase 3: 前端实现（2-3 天）

### API 封装

- [ ] 3.1 扩展调拨单 API 封装

  - File: `admin/views/shop/api/transfer-order.ts`
  - Purpose: 在 API 调用中支持 role_type 参数
  - Requirements: Requirement 1
  - Leverage: 现有 API: `admin/views/shop/api/transfer-order.ts`
  - Prompt: Role: Frontend Developer with TypeScript expertise | Task: 在 getTransferOrderList() 方法中新增 roleType 参数 | Context: 使用 axios，定义 TypeScript 类型 | Restrictions: 遵循 .cursor/rules/vue.mdc | Success: API 封装完成，类型定义正确

### 组件开发

- [ ] 3.2 创建类型筛选器组件

  - File: `admin/views/shop/components/transfer-order/RoleTypeFilter.vue`
  - Purpose: 实现"我发出/我接收/我审核"筛选器
  - Requirements: Requirement 1（类型筛选器）
  - Leverage: 现有组件: `admin/views/shop/components/`，Element Plus Radio Group
  - Prompt: Role: Frontend Developer with Vue 3 expertise | Task: 创建 RoleTypeFilter 组件，实现单选筛选器 | Context: 使用 Element Plus Radio Group，默认选中"我接收"，支持取消选择 | Restrictions: 遵循 .cursor/rules/vue.mdc，使用 Composition API | Success: 组件创建成功，交互正常

- [ ] 3.3 扩展列表页面

  - File: `admin/views/shop/pages/transfer-order/index.vue`
  - Purpose: 集成类型筛选器组件，实现列表筛选功能
  - Requirements: Requirement 1, 2, 3, 4
  - Leverage: 现有页面: `admin/views/shop/pages/transfer-order/index.vue`，Task 3.2 的筛选器组件
  - Prompt: Role: Frontend Developer with Vue 3 expertise | Task: 在调拨单列表页面中集成 RoleTypeFilter 组件，实现筛选功能 | Context: 筛选器变化时调用 API 刷新列表，重置分页 | Restrictions: 遵循 .cursor/rules/vue.mdc | Success: 页面功能完整，筛选逻辑正确

- [ ] 3.4 实现状态筛选器联动

  - File: `admin/views/shop/pages/transfer-order/index.vue`
  - Purpose: 角色筛选切换时，状态筛选器自动联动
  - Requirements: Requirement 5（状态筛选联动）
  - Leverage: Task 3.3 的列表页面
  - Prompt: Role: Frontend Developer with Vue 3 expertise | Task: 实现角色筛选与状态筛选的联动逻辑 | Context: 选择"我审核"时，状态筛选器固定为"待审核"；其他角色时，状态筛选器可选全部状态 | Restrictions: 遵循 .cursor/rules/vue.mdc | Success: 联动逻辑正确，用户体验良好

- [ ] 3.5 前端单元测试

  - File: `admin/views/shop/components/transfer-order/__tests__/RoleTypeFilter.spec.ts`
  - Purpose: 测试筛选器组件
  - Requirements: Requirement 1
  - Leverage: 现有测试: `admin/views/shop/components/__tests__/`
  - Prompt: Role: QA Engineer with Vue testing expertise | Task: 为 RoleTypeFilter 组件编写单元测试 | Context: 测试默认选中，测试切换选择，测试取消选择 | Restrictions: 使用 Vitest + Vue Test Utils | Success: 测试通过

---

## Phase 4: 集成测试和优化（1 天）

- [ ] 4.1 端到端集成测试

  - File: -
  - Purpose: 测试前后端联调功能
  - Requirements: 所有功能需求
  - Leverage: 浏览器手动测试
  - Test Cases:
    1. 测试"我发出"视图，验证列表数据正确
    2. 测试"我接收"视图（默认），验证列表数据正确
    3. 测试"我审核"视图，验证四种审核场景
    4. 测试取消选择，显示全部单据
    5. 测试状态筛选联动，验证联动逻辑正确
    6. 测试边界场景（无数据、网络异常等）
  - Success: 所有测试用例通过

- [ ] 4.2 性能测试

  - File: -
  - Purpose: 验证查询性能
  - Requirements: 性能要求（< 200ms）
  - Leverage: 浏览器 Network 面板，或性能测试工具
  - Success: 列表查询响应时间 < 200ms

- [ ] 4.3 文档更新

  - File: `docs/shared/api/shop_api.md`, `CHANGELOG.md`
  - Purpose: 更新 API 文档和变更日志
  - Requirements: 文档要求
  - Leverage: 现有文档
  - Success: 文档已更新，CHANGELOG 已记录

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] Vue 代码通过 ESLint 检查
- [ ] 测试覆盖率达标
  - Service: ≥ 70%
  - Repository: ≥ 80%
- [ ] 所有测试通过

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成

### 文档同步

- [ ] API 文档已更新（如有新接口）
- [ ] CHANGELOG.md 已更新
- [ ] 用户指南已更新（如需要）

### 规范遵循

- [ ] 遵循 `.cursor/rules/go-main.mdc`
- [ ] 遵循 `.cursor/rules/vue.mdc`
- [ ] 遵循 `.cursor/rules/api.mdc`
- [ ] 遵循 `.cursor/rules/database.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-admin-transfer-order-optimization/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-admin-transfer-order-optimization/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-admin-transfer-order-optimization/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/story-admin-transfer-order-optimization/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/story-admin-transfer-order-optimization/tasks.md)" | bc
```

### 执行流程

1. **选择任务**: 选择下一个未完成任务
2. **阅读需求**: 查看 requirements.md 中的关联需求
3. **查看复用**: 检查 Leverage 中的可复用代码
4. **使用 AI**: 复制 Prompt 模板，让 AI 生成代码
5. **实现代码**: 按照规范实现功能
6. **运行检查**: `go fmt`, `go vet`, `go test`
7. **标记完成**: 将 `[ ]` 改为 `[x]`
8. **提交代码**: Git commit（参考 `.cursor/rules/version.mdc`）

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/weifashi/2026-01/2026-01-06.md`
- 在执行任务过程中若总结出经验或规避策略，请记录 Episode，并在 tasks.md 尾部更新 `Related Episode`。

---

**版本**: v1.0.0  
**创建日期**: 2026-01-06  
**作者**: weifashi  
**最后更新**: 2026-01-06

