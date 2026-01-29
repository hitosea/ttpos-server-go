# Lineman 触发菜单同步（TriggerSyncMenu）任务分解

> 本文档定义 Lineman 触发菜单同步功能的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 9  
**已完成**: 5  
**进行中**: Phase 3  
**完成率**: 56%

**预估 SP**: 2  
**技术栈**: Go BMP (ttpos-takeout)

---

## Phase 1: 核心实现

### 任务模板说明

每个任务包含以下信息：

- **File**: 需要修改的文件路径
- **Purpose**: 任务目的（一句话说明为什么要做）
- **Requirements**: 关联的需求编号（如: 1.1, 2.3）
- **Leverage**: 可复用的现有代码路径
- **Prompt**: AI 执行提示模板（可选）

---

- [x] 1.1 实现 Controller 层

  - File: `ttpos-bmp/app/ttpos-takeout/internal/controller/lineman/lineman_v1_trigger_sync_menu.go`
  - Purpose: 实现 TriggerSyncMenu HTTP 接口，接收 Lineman 触发同步请求
  - Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 1.6
  - Leverage: 
    - 现有实现: `lineman_v1_menu_sync_notification.go` - Controller 结构和错误处理模式
    - API 定义: `api/lineman/v1/menu.go` - Request/Response 结构体
  - Prompt: 
    ```
    Role: Go Developer specializing in GoFrame HTTP Controller
    
    Task: 实现 TriggerSyncMenu Controller，参考 MenuSyncNotification 的实现模式
    
    Context:
    - Current file: ttpos-bmp/app/ttpos-takeout/internal/controller/lineman/lineman_v1_trigger_sync_menu.go
    - Leverage code: lineman_v1_menu_sync_notification.go (同目录)
    - Requirements: requirements.md Requirement 1
    - Project specs: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc
    
    Restrictions:
    - 使用类型断言模式调用 service.Lineman().HandleTriggerSyncMenu()
    - 使用 gerror 处理错误（gcode.CodeInvalidParameter, gcode.CodeNotFound, gcode.CodeInternalError）
    - 返回 LinemanCommonResData 格式（status, code, message）
    - HTTP 状态码: 200/400/404/500
    - Controller 不写业务逻辑，只负责参数解析和响应
    
    Success Criteria:
    - Controller 实现完整
    - 错误码映射正确（400/404/500）
    - 响应格式符合 Lineman API 规范
    - 代码通过 go fmt 和 go vet
    ```

- [x] 1.2 创建 Logic 层

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/trigger_sync_menu.go`（新建）
  - Purpose: 实现业务编排逻辑，记录 menu_log 并调用 SyncMenu
  - Requirements: 1.1, 2.1, 2.2, 2.3, 2.4, 2.5, 3.1, 3.2, 3.3
  - Leverage:
    - 现有 Logic: `menu_sync_notification.go` - Logic 结构和参数校验
    - Service 接口: `service.ChannelMenu().LogMenuSync()` - 记录日志
    - Service 接口: `service.Lineman().SyncMenu()` - 触发同步
    - 常量定义: `internal/consts/consts.go` - ProviderLineman, MenuSyncTypeNotify
  - Prompt:
    ```
    Role: Go Developer with GoFrame and business logic expertise
    
    Task: 创建 HandleTriggerSyncMenu Logic，实现业务编排
    
    Context:
    - New file: ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/trigger_sync_menu.go
    - Leverage code: menu_sync_notification.go (同目录)
    - Requirements: requirements.md Requirement 1, 2, 3
    - Project specs: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc
    
    Implementation Steps:
    1. 校验 req 是否为空（返回 gcode.CodeInvalidParameter）
    2. 解析 req.StoreId（在 Lineman 场景中，storeId 就是 shopUUID，使用 strconv.ParseUint 转换）
    3. 调用 service.ChannelMenu().LogMenuSync() 记录到 menu_log
       - provider: consts.ProviderLineman
       - sync_type: consts.MenuSyncTypeNotify
       - sync_status: "QUEUED"
       - success: false
    4. 调用 service.Lineman().SyncMenu(ctx, shopUUID) 触发同步
       - 失败只记录日志，不影响响应（异步同步）
    5. 返回 nil 表示成功
    
    Restrictions:
    - 使用 gerror 处理错误
    - 使用 g.Log().Errorf() 记录错误日志
    - Logic 不直接操作数据库，通过 Service 调用
    - 遵循 GoFrame 项目结构
    - storeId 直接转换为 shopUUID，不需要查询数据库
    
    Success Criteria:
    - Logic 实现完整，业务流程正确
    - 参数校验完善
    - 错误处理正确（400/404/500）
    - 日志记录清晰
    - 代码通过 go fmt 和 go vet
    ```

- [x] 1.3 注册 Logic 方法到 Service 接口

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/lineman.go`
  - Purpose: 将 HandleTriggerSyncMenu 方法添加到 sLineman 结构体
  - Requirements: 1.1
  - Leverage: 
    - 现有实现: `menu_sync_notification.go` - HandleMenuSyncNotification 方法
    - Service 定义: `internal/service/lineman.go` - ILineman 接口
  - Note: 确保 sLineman 结构体包含 HandleTriggerSyncMenu 方法，以支持 Controller 的类型断言
  - Success: sLineman 实现了 triggerSyncMenuHandler 接口

- [x] 1.4 编写 Logic 层单元测试

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/trigger_sync_menu_test.go`（新建）
  - Purpose: 确保 Logic 业务逻辑正确
  - Requirements: 所有功能需求
  - Leverage: 现有测试文件（如有）
  - Prompt:
    ```
    Role: QA Engineer with Go testing expertise
    
    Task: 为 HandleTriggerSyncMenu Logic 编写单元测试，覆盖率 ≥ 70%
    
    Context:
    - Target file: trigger_sync_menu.go
    - Test file: trigger_sync_menu_test.go (新建)
    - Coverage target: ≥ 70%
    
    Test Cases Required:
    1. 正常场景: req 有效，storeId 有效（可转换为 uint64）
    2. 异常场景: req 为 nil
    3. 异常场景: storeId 无效（无法解析为 uint64）
    4. 异常场景: LogMenuSync 失败
    5. 异常场景: SyncMenu 失败（不影响响应）
    
    Restrictions:
    - 使用 GoFrame 测试框架
    - Mock service.ChannelMenu() 和 service.Lineman()
    - 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc
    
    Success Criteria:
    - 测试覆盖率 ≥ 70%
    - 所有测试通过
    - 边界情况已覆盖
    ```

---

## Phase 2: 集成测试

- [ ] 2.1 编写集成测试

  - File: `ttpos-bmp/app/ttpos-takeout/test/integration/lineman_trigger_sync_menu_test.go`（如适用）
  - Purpose: 测试 TriggerSyncMenu 端到端流程
  - Requirements: 所有功能需求
  - Leverage: 现有集成测试
  - Test Cases:
    1. 调用 TriggerSyncMenu API
    2. 验证 menu_log 记录是否正确（sync_type=NOTIFY, status=QUEUED）
    3. 验证 SyncMenu 是否被调用
  - Success: 集成测试通过

- [ ] 2.2 API 测试（Postman）

  - File: -
  - Purpose: 手动测试 API 接口
  - Requirements: 1.1, 1.4, 1.5
  - Test Cases:
    1. 正常请求: HTTP 200，status=ok
    2. 参数缺失: HTTP 400，code=400
    3. storeId 不存在: HTTP 404，code=404
    4. 内部错误: HTTP 500，code=500
  - Success: 所有场景测试通过

---

## Phase 3: 文档和优化

- [ ] 3.1 更新 API 文档

  - File: `docs/shared/api/lineman_api.md`（如适用）
  - Purpose: 确保 API 文档与代码同步
  - Requirements: 文档要求
  - Leverage: `docs/agent/templates/api-doc-template.md`
  - Content:
    - API 名称: TriggerSyncMenu
    - URL: POST /v1/partners/{partnerId}/stores/{storeId}/menus/trigger-sync
    - 请求参数: partnerId, storeId
    - 响应格式: LinemanCommonResData
    - 错误码: 200/400/404/500
  - Success: API 文档已更新

- [x] 3.2 更新 CHANGELOG.md

  - File: `ttpos-bmp/CHANGELOG.md`
  - Purpose: 记录功能变更
  - Requirements: 文档要求
  - Content:
    ```markdown
    ## [v2.14.0] - 2026-01-15
    
    ### Added
    - 新增 Lineman TriggerSyncMenu 接口，支持平台主动触发菜单同步
    - 新增 menu_log 记录，sync_type=NOTIFY，status=QUEUED
    ```
  - Success: CHANGELOG 已更新

- [ ] 3.3 性能测试（可选）

  - File: -
  - Purpose: 确保接口性能达标
  - Requirements: 性能要求
  - Test:
    - 接口响应时间 < 300ms
    - 并发 500 QPS
  - Tools: wrk, ab
  - Success: 性能测试通过

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] 测试覆盖率达标
  - Logic: ≥ 70%
- [ ] 所有测试通过

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成:
  - [ ] TriggerSyncMenu 接口正常响应 HTTP 200
  - [ ] menu_log 正常写入（sync_type=NOTIFY, status=QUEUED）
  - [ ] SyncMenu 正常调用

### 文档同步

- [ ] API 文档已更新（如有新接口）
- [ ] CHANGELOG.md 已更新
- [ ] design.md 和 tasks.md 保持一致

### 规范遵循

- [ ] 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`
- [ ] 遵循 `ttpos-bmp/.cursor/rules/proto-rules.mdc`（如适用）
- [ ] 遵循 `.cursor/rules/api.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-shop-lineman-trigger-sync-menu/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-shop-lineman-trigger-sync-menu/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-shop-lineman-trigger-sync-menu/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/story-shop-lineman-trigger-sync-menu/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/story-shop-lineman-trigger-sync-menu/tasks.md)" | bc
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

## 附录：标准 Prompt 模板

### Go BMP 开发

```
Role: Go Developer specializing in GoFrame

Task: {具体任务描述，引用 Requirements}

Context:
- Current file: {文件路径}
- Leverage code: {可复用代码路径}
- Requirements: {需求编号和内容}
- Project specs: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc

Restrictions:
- 使用 GoFrame 2.x 框架
- 禁止修改 dao/entity/do/ 目录（自动生成）
- 使用 gerror 处理错误（不用标准库 errors）
- 使用 g.Log() 记录日志
- Controller 不写业务逻辑
- Logic 不直接操作数据库

Success Criteria:
- {成功标准1}
- 代码通过 go fmt 和 go vet
- 测试覆盖率 ≥ 70%
```

### 测试工程师

```
Role: QA Engineer with GoFrame testing expertise

Task: {测试任务描述}

Context:
- Target file: {测试目标文件}
- Test file: {测试文件路径}
- Coverage target: ≥ 70%

Test Cases Required:
- 正常场景测试
- 异常场景测试
- 边界条件测试
- 并发场景测试（如适用）

Restrictions:
- 使用 GoFrame 测试框架
- Mock service 依赖
- 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc

Success Criteria:
- 测试覆盖率达标
- 所有测试通过
- 边界情况已覆盖
```

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/rikugun/2026-01/2026-01-15.md`
- 在执行任务过程中若总结出经验或规避策略，请记录 Episode，并在 tasks.md 尾部更新 `Related Episode`。

---

**模板版本**: v1.0.0  
**创建日期**: 2026-01-15  
**作者**: rikugun  
**最后更新**: 2026-01-15
