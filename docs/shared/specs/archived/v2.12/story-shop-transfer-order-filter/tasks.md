# 调拨单-本店提交与其他门店提交分开管理 任务分解

> 本文档定义后端 API 扩展的详细执行任务清单（前端实现在 ttpos-flutter 仓库）。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 11  
**已完成**: 6  
**进行中**: 1  
**完成率**: 55%

---

## Phase 1: 数据库准备

- [x] 1.1 确认调拨单表索引

  - File: `admin/database/migrations/` 或直接查询数据库
  - Purpose: 确认 `from_store_id`, `to_store_id` 索引存在，避免查询性能问题
  - Requirements: Requirement 2（API 查询支持）
  - Leverage: 现有迁移文件，查看 `ttpos_transfer_order` 表定义
  - Command: 
    ```sql
    SHOW INDEX FROM ttpos_transfer_order WHERE Key_name LIKE '%store%';
    ```
  - Success: 确认 `idx_from_store_id` 和 `idx_to_store_id` 索引存在；若不存在则补充迁移脚本

---

## Phase 2: Go Main 后端实现

### DTO 层

- [x] 2.1 在 Request DTO 增加 SubmitSide 字段

  - File: `main/app/dto/req/transfer_order_req.go`
  - Purpose: 支持前端传递提交方筛选参数
  - Requirements: Requirement 2.1
  - Leverage: 现有 DTO: `main/app/dto/req/`，查看 `TransferOrderListReq` 结构体
  - Prompt: 
    ```
    Role: Go Developer
    Task: 在 TransferOrderListReq 结构体中增加 SubmitSide 字段
    Context: 
    - 字段名: SubmitSide
    - 类型: string
    - JSON tag: "submit_side"
    - 不需要 binding:"required"（允许为空，默认 "all"）
    - 注释: "提交方筛选：all|self|other，默认 all"
    Restrictions: 遵循 .cursor/rules/go-main.mdc
    Success: 字段添加成功，DTO 定义清晰
    ```
  - Success: `SubmitSide string json:"submit_side"` 已添加

### Repository 层

- [x] 2.2 在 Repository 接口增加 WhereSubmitSide 方法

  - File: `main/app/repository/i_transfer_order_repo.go`
  - Purpose: 定义提交方筛选的选项方法接口
  - Requirements: Requirement 2.2
  - Leverage: 现有 Repository 接口: `main/app/repository/i_*_repo.go`，查看选项方法定义模式
  - Prompt:
    ```
    Role: Go Developer specializing in Repository Pattern
    Task: 在 ITransferOrderRepo 接口中增加 WhereSubmitSide 方法
    Context:
    - 方法签名: WhereSubmitSide(currentStoreId uint64, submitSide string) DBOption
    - 返回类型: DBOption（选项模式）
    - 注释: "提交方筛选选项"
    Restrictions: 遵循 .cursor/rules/go-main.mdc，使用选项模式
    Success: 接口方法定义成功
    ```
  - Success: 接口方法已定义

- [x] 2.3 实现 Repository 的 WhereSubmitSide 方法

  - File: `main/app/repository/transfer_order_repo.go`
  - Purpose: 实现按提交方筛选的查询逻辑
  - Requirements: Requirement 2.3
  - Leverage: 现有 Repository 实现: `main/app/repository/*_repo.go`，参考其他 Where 方法的实现
  - Prompt:
    ```
    Role: Go Developer with GORM expertise
    Task: 实现 WhereSubmitSide 方法，根据 submitSide 参数组装查询条件
    Context:
    - submitSide="self": WHERE from_store_id = currentStoreId
    - submitSide="other": WHERE to_store_id = currentStoreId AND from_store_id != currentStoreId
    - submitSide="all" (默认): WHERE from_store_id = currentStoreId OR to_store_id = currentStoreId
    - 使用选项模式返回 func(db *gorm.DB) *gorm.DB
    Restrictions: 
    - 使用 GORM 查询构建
    - 使用 switch-case 判断 submitSide
    - 遵循 .cursor/rules/go-main.mdc
    Success: WhereSubmitSide 方法实现正确，三种筛选值逻辑正确
    ```
  - Success: 方法实现完成，SQL 逻辑正确

- [ ] 2.4 编写 Repository 单元测试

  - File: `main/app/repository/transfer_order_repo_test.go`
  - Purpose: 测试 WhereSubmitSide 方法的 SQL 生成正确性
  - Requirements: Requirement 2（测试要求）
  - Leverage: 现有测试: `main/app/repository/*_repo_test.go`
  - Prompt:
    ```
    Role: QA Engineer with Go testing expertise
    Task: 为 WhereSubmitSide 方法编写单元测试，覆盖三种筛选值
    Context:
    - 测试 submitSide="self" 的查询条件
    - 测试 submitSide="other" 的查询条件
    - 测试 submitSide="all" 的查询条件
    - 使用 mock DB 或内存数据库
    Restrictions: 遵循 .cursor/rules/go-main.mdc
    Success: 测试覆盖率 ≥ 80%，三种场景测试通过
    ```
  - Success: 单元测试通过

### Service 层

- [x] 2.5 修改 Service 的 GetList 方法支持 SubmitSide

  - File: `main/app/service/transfer_order_srv.go`
  - Purpose: 在业务逻辑层增加提交方筛选支持
  - Requirements: Requirement 2.4
  - Leverage: 现有 Service 实现: `main/app/service/transfer_order_srv.go`，查看 GetList 方法
  - Prompt:
    ```
    Role: Go Developer with business logic expertise
    Task: 修改 TransferOrderSrv.GetList 方法，增加 submit_side 参数处理
    Context:
    - 从 req.SubmitSide 获取参数，默认为 "all"
    - 参数校验：仅允许 "all"|"self"|"other"，非法值返回 error
    - 获取当前门店 ID: ctx.GetStoreId()
    - 调用 repo.WhereSubmitSide(currentStoreId, submitSide)
    - 与其他筛选条件（status, date 等）组合使用
    Restrictions:
    - 不使用 panic，返回 error
    - 遵循 .cursor/rules/go-main.mdc
    - Service 通过 DBManager.GetDB(ctx) 获取 db 后创建 Repository
    Success: GetList 方法支持 submit_side，参数校验正确，与其他筛选条件组合正确
    ```
  - Success: Service 方法修改完成，逻辑正确

- [ ] 2.6 编写 Service 单元测试

  - File: `main/app/service/transfer_order_srv_test.go`
  - Purpose: 测试 Service 层的参数校验和逻辑组装
  - Requirements: Requirement 2（测试要求）
  - Leverage: 现有测试: `main/app/service/*_srv_test.go`
  - Prompt:
    ```
    Role: QA Engineer with Go testing expertise
    Task: 为 TransferOrderSrv.GetList 编写单元测试，覆盖 submit_side 参数场景
    Context:
    - 测试 submit_side="all" 的正常查询
    - 测试 submit_side="self" 的正常查询
    - 测试 submit_side="other" 的正常查询
    - 测试 submit_side 非法值返回错误
    - 测试与 status、date 等条件组合
    Restrictions: 遵循 .cursor/rules/go-main.mdc
    Success: 测试覆盖率 ≥ 70%，所有场景测试通过
    ```
  - Success: 单元测试通过

### API 层

- [x] 2.7 修改 API Controller 支持 SubmitSide 参数

  - File: `main/app/api/transfer_order_api.go`
  - Purpose: API 层解析并传递 submit_side 参数
  - Requirements: Requirement 2.5
  - Leverage: 现有 API: `main/app/api/transfer_order_api.go`，查看 List 方法
  - Prompt:
    ```
    Role: Go Developer with Gin framework expertise
    Task: 修改 TransferOrderAPI.List 方法，确保 submit_side 参数正确解析和传递
    Context:
    - 使用 c.ShouldBindJSON(&req) 解析参数
    - req.SubmitSide 会自动解析，无需额外处理
    - Service 层会处理参数校验和默认值
    - 错误处理：使用 helper.ErrorWithDetail
    - 成功响应：使用 helper.Success，data 必须是对象
    Restrictions:
    - 遵循 .cursor/rules/api.mdc
    - 不直接使用 c.JSON()
    - URL 使用 snake_case
    Success: API 方法支持 submit_side，参数解析正确，响应格式正确
    ```
  - Success: API 方法修改完成

- [ ] 2.8 编写 API 集成测试

  - File: `main/app/api/transfer_order_api_test.go`
  - Purpose: 测试 API 接口的参数解析和响应格式
  - Requirements: Requirement 2（测试要求）
  - Leverage: 现有测试: `main/app/api/*_api_test.go`
  - Prompt:
    ```
    Role: QA Engineer specializing in API testing
    Task: 为 TransferOrderAPI.List 编写集成测试，覆盖 submit_side 参数
    Context:
    - 测试 GET /api/v1/transfer_order/list?submit_side=all
    - 测试 submit_side=self
    - 测试 submit_side=other
    - 测试 submit_side 非法值返回 400 错误
    - 测试响应格式 {code, message, data{list, meta}}
    - 测试与其他参数组合（status, page_no 等）
    Restrictions: 遵循 .cursor/rules/go-main.mdc
    Success: 所有 API 测试通过，响应格式正确
    ```
  - Success: API 集成测试通过

---

## Phase 3: 文档和验证

- [ ] 3.1 更新 API 文档

  - File: `docs/shared/api/transfer_order_api.md`（如存在）或创建
  - Purpose: 补充 submit_side 参数说明
  - Requirements: Requirement 2（文档要求）
  - Leverage: `docs/agent/templates/api-doc-template.md`
  - Prompt:
    ```
    Role: Technical Writer
    Task: 在调拨单 API 文档中补充 submit_side 参数说明
    Context:
    - 参数名: submit_side
    - 类型: string
    - 可选值: all|self|other
    - 默认值: all
    - 说明: 
      - all: 全部相关调拨单
      - self: 仅本店提交的调拨单
      - other: 仅其他门店提交的调拨单
    - 示例请求和响应
    Restrictions: 遵循 .cursor/rules/documentation.mdc
    Success: API 文档已更新，参数说明清晰
    ```
  - Success: API 文档已更新

- [ ] 3.2 手动测试与前端联调

  - File: -
  - Purpose: 确保后端 API 与 Flutter 前端对接正常
  - Requirements: 所有功能需求
  - Leverage: Postman 或 curl 测试工具
  - Steps:
    1. 启动本地开发环境
    2. 使用 Postman 测试三种 submit_side 值
    3. 验证响应数据正确性
    4. 与前端开发确认接口对接
    5. 测试与其他筛选条件组合
  - Success: 三种筛选值返回数据正确，与前端联调成功

- [ ] 3.3 更新 CHANGELOG

  - File: `CHANGELOG.md`
  - Purpose: 记录功能变更
  - Requirements: 文档要求
  - Content:
    ```markdown
    ## [v2.11.0] - 2025-12-05
    
    ### Added
    - 调拨单列表 API 新增 `submit_side` 查询参数，支持按提交方筛选（全部/本店提交/他店提交）
    ```
  - Success: CHANGELOG 已更新

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] 测试覆盖率达标
  - Repository: ≥ 80%
  - Service: ≥ 70%
- [ ] 所有测试通过（`go test ./...`）

### 功能完整性

- [ ] requirements.md 中的所有后端需求已满足
- [ ] design.md 中的设计已实现
- [ ] API 支持三种 submit_side 值
- [ ] 参数校验正确
- [ ] 与其他筛选条件组合正常

### 文档同步

- [ ] API 文档已更新
- [ ] CHANGELOG.md 已更新
- [ ] 与前端开发确认接口对接

### 规范遵循

- [ ] 遵循 `.cursor/rules/go-main.mdc`
- [ ] 遵循 `.cursor/rules/api.mdc`
- [ ] 遵循 `.cursor/rules/database.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-shop-transfer-order-filter/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-shop-transfer-order-filter/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-shop-transfer-order-filter/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/story-shop-transfer-order-filter/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/story-shop-transfer-order-filter/tasks.md)" | bc
```

### 执行流程

1. **选择任务**: 按 Phase 顺序执行
2. **阅读需求**: 查看 requirements.md 中的关联需求
3. **查看复用**: 检查 Leverage 中的可复用代码
4. **使用 AI**: 复制 Prompt 模板，让 AI 生成代码
5. **实现代码**: 按照规范实现功能
6. **运行检查**: `go fmt`, `go vet`, `go test`
7. **标记完成**: 将 `[ ]` 改为 `[x]`
8. **提交代码**: Git commit（参考 `.cursor/rules/version.mdc`）

---

## 附录：前端协作说明

### 前端实现范围（ttpos-flutter 仓库）

前端需实现以下功能（不在本 Spec 范围）：

1. **筛选按钮组 UI**:
   - 三个按钮：全部 / 本店提交 / 他店提交
   - 选中状态高亮
   - 与现有筛选样式一致

2. **API 调用**:
   - 调用后端接口时传递 `submit_side` 参数
   - 参数值：`all` / `self` / `other`

3. **状态管理**:
   - 筛选状态持久化（路由参数或本地存储）
   - 页面刷新后恢复上次选择

4. **交互反馈**:
   - 加载状态提示
   - 空数据提示与快捷重置
   - 错误提示

### 前后端对接检查点

- [ ] API 接口地址确认：`/api/v1/transfer_order/list`
- [ ] 参数格式确认：`submit_side` 取值 `all|self|other`
- [ ] 响应格式确认：`{code, message, data{list, meta}}`
- [ ] 错误码确认：非法参数返回 code=0
- [ ] 联调测试通过

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 在执行任务过程中若总结出经验或规避策略，请记录 Episode。

---

**模板版本**: v1.0.0  
**最后更新**: 2025-12-05  
**维护者**: 后端开发组

