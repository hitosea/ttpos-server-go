# KDS 等待时长颜色配置 任务分解

> 本文档定义 KDS 等待时长颜色配置功能的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 19  
**已完成**: 14  
**进行中**: -  
**完成率**: 74%

---

## Phase 1: 数据模型和 DTO

### 任务模板说明

每个任务包含以下信息：

- **File**: 需要修改的文件路径
- **Purpose**: 任务目的（一句话说明为什么要做）
- **Requirements**: 关联的需求编号（如: 1.1, 2.3）
- **Leverage**: 可复用的现有代码路径
- **Prompt**: AI 执行提示模板（可选）

---

- [x] 1.1 定义 WaitTimeColorRange 结构体

  - File: `main/app/dto/resp/setting/kitchen_setting.go`
  - Purpose: 定义等待时长颜色区间的数据结构
  - Requirements: 2.1
  - Leverage: 现有 DTO: `main/app/dto/resp/setting/kitchen_setting.go`
  - Prompt: Role: Go Developer | Task: 在 kitchen_setting.go 中定义 WaitTimeColorRange 结构体，包含 Minute (string) 和 Color (string) 字段 | Context: 用于表示等待时长颜色区间配置，Minute 使用字符串类型以兼容 PHP | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 结构体定义正确，JSON 标签正确

- [x] 1.2 更新 KitchenResp 结构体

  - File: `main/app/dto/resp/setting/kitchen_setting.go`
  - Purpose: 在 KitchenResp 中新增 WaitTimeColorRanges 字段，保留 WaitColor 字段
  - Requirements: 1.4, 2.2
  - Leverage: 现有 KitchenResp: `main/app/dto/resp/setting/kitchen_setting.go`
  - Prompt: Role: Go Developer | Task: 在 KitchenResp 结构体中新增 WaitTimeColorRanges ([]WaitTimeColorRange) 字段，保留 WaitColor ([]string) 字段 | Context: 保持向后兼容，同时支持新旧格式 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 字段定义正确，JSON 标签正确

- [x] 1.3 创建 Request DTO

  - File: `main/app/dto/req/setting/kitchen_setting.go`（新建文件）
  - Purpose: 定义保存厨显设置的请求参数结构体
  - Requirements: 1.3
  - Leverage: 现有 Request DTO: `main/app/dto/req/setting/`
  - Prompt: Role: Go Developer | Task: 创建 SaveKitchenSettingReq 结构体，包含 IsWaitColor 和 WaitTimeColorRanges 字段，添加 binding 验证标签 | Context: IsWaitColor 必须是 0 或 1，WaitTimeColorRanges 必须包含 3 个区间，Color 字段统一使用 RGB 格式（#xxxxxx），限定颜色值：黑色 #100A05，黄色 #FFBE00，红色 #E50028 | Restrictions: 遵循 .cursor/rules/go-main.mdc，使用 binding 标签验证 | Success: DTO 创建成功，validation 正确

---

## Phase 2: Service 层实现

- [x] 2.1 实现参数验证逻辑

  - File: `main/app/service/setting/setting.go`
  - Purpose: 验证等待时长颜色区间的合法性（时间区间、颜色格式）
  - Requirements: 1.5
  - Leverage: 现有 Service: `main/app/service/setting/setting.go`
  - Prompt: Role: Go Developer with validation expertise | Task: 实现 validateWaitTimeColorRanges 方法，验证：1) 必须3个区间，2) 第一区间必须为0分钟，3) 第二区间1-60分钟，4) 第三区间必须大于第二区间且≤60分钟，5) 颜色格式验证（RGB 格式 #xxxxxx，限定颜色值：黑色 #100A05，黄色 #FFBE00，红色 #E50028） | Context: 严格验证参数，返回明确的错误信息 | Restrictions: 遵循 .cursor/rules/go-main.mdc，不使用 panic，返回 error | Success: 验证逻辑完整，错误信息明确

- [x] 2.2 实现新旧格式转换逻辑

  - File: `main/app/service/setting/setting.go`
  - Purpose: 实现新旧格式的相互转换，保持向后兼容
  - Requirements: 2.3
  - Leverage: 现有 Service: `main/app/service/setting/setting.go`
  - Prompt: Role: Go Developer | Task: 实现 convertToOldFormat 和 convertFromOldFormat 方法，实现新旧格式转换 | Context: 旧格式: ["red", "yellow"]（第一个元素对应第二区间，第二个元素对应第三区间），新格式: [{"minute": "0", "color": "#100A05"}, {"minute": "10", "color": "#FFBE00"}, {"minute": "20", "color": "#E50028"}]（注意：minute 为字符串类型以兼容 PHP），转换规则：red ↔ #E50028，yellow ↔ #FFBE00，黑色固定 #100A05 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 转换逻辑正确，处理边界情况

- [x] 2.3 实现 SaveKitchenSetting 方法

  - File: `main/app/service/setting/setting.go`
  - Purpose: 实现保存厨显设置的核心业务逻辑
  - Requirements: 1.5
  - Leverage: 现有 Service: `main/app/service/setting/setting.go`，参考 `EditStoreSetting` 方法
  - Prompt: Role: Go Developer with business logic expertise | Task: 实现 SaveKitchenSetting 方法，包含：1) 参数验证，2) 获取当前配置，3) 更新配置，4) 格式转换，5) 保存到数据库，6) WebSocket 推送 | Context: 使用 UpdateSetting 保存配置，使用 websocket.PushClient 推送更新 | Restrictions: 遵循 .cursor/rules/go-main.mdc，不使用 panic，返回 error | Success: 方法实现完整，业务逻辑正确

- [x] 2.4 更新 GetKitchenSetting 方法

  - File: `main/app/service/setting/setting.go`
  - Purpose: 更新 GetKitchenSetting 方法，支持返回新格式数据
  - Requirements: 1.6
  - Leverage: 现有 GetKitchenSetting: `main/app/service/setting/setting.go` (line 1144)
  - Prompt: Role: Go Developer | Task: 更新 GetKitchenSetting 方法，如果只有旧格式数据，自动转换为新格式返回 | Context: 保持向后兼容，同时返回新旧格式数据 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 方法更新正确，格式转换正确

- [x] 2.5 更新默认配置逻辑

  - File: `main/app/service/setting/default.go`
  - Purpose: 更新默认厨显设置，包含默认等待时长颜色配置
  - Requirements: 2.4
  - Leverage: 现有 getDefaultKitchen: `main/app/service/setting/default.go` (line 170)
  - Prompt: Role: Go Developer | Task: 更新 getDefaultKitchen 方法，设置默认 WaitTimeColorRanges: [0分钟黑色, 10分钟黄色, 20分钟红色] | Context: 第一区间固定0分钟黑色 #100A05，第二、三区间从旧后台配置读取（wait_color: ["red", "yellow"]），如无则使用默认值 #FFBE00 和 #E50028 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 默认配置正确

- [ ] 2.6 编写 Service 单元测试

  - File: `main/app/service/setting/setting_test.go`
  - Purpose: 确保 Service 业务逻辑正确
  - Requirements: 测试要求
  - Leverage: 现有测试: `main/app/service/setting/`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 SaveKitchenSetting 和格式转换方法编写单元测试，覆盖率 ≥ 70% | Context: 测试参数验证、格式转换、配置保存 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 ≥ 70%，所有测试通过

---

## Phase 3: API 层实现

- [x] 3.1 实现 GetKitchenSetting API

  - File: `main/app/api/v1/shop/shop_setting.go`
  - Purpose: 实现获取厨显设置的 API 接口
  - Requirements: 1.1
  - Leverage: 现有 API: `main/app/api/v1/shop/shop_setting.go`，参考 `GetKioskSetting` 方法 (line 913)
  - Prompt: Role: Go Developer with Gin framework expertise | Task: 实现 GetKitchenSetting API，调用 Service 获取配置，返回 KitchenResp | Context: URL: GET /api/v1/shop/setting/kitchen，使用 helper.Success 返回响应 | Restrictions: 遵循 .cursor/rules/api.mdc，data 必须是对象 | Success: API 创建成功，响应格式正确

- [x] 3.2 实现 SaveKitchenSetting API

  - File: `main/app/api/v1/shop/shop_setting.go`
  - Purpose: 实现保存厨显设置的 API 接口
  - Requirements: 1.2
  - Leverage: 现有 API: `main/app/api/v1/shop/shop_setting.go`，参考 `SaveKioskSetting` 方法 (line 942)
  - Prompt: Role: Go Developer with Gin framework expertise | Task: 实现 SaveKitchenSetting API，接收请求参数，调用 Service 保存配置 | Context: URL: POST /api/v1/shop/setting/kitchen，使用 helper.Success 返回响应 | Restrictions: 遵循 .cursor/rules/api.mdc，data 必须是对象 | Success: API 创建成功，参数验证正确

- [x] 3.3 注册 API 路由

  - File: `main/router/router.go`
  - Purpose: 注册厨显设置 API 路由
  - Requirements: 1.1, 1.2
  - Leverage: 现有路由: `main/router/router.go`
  - Prompt: Role: Go Developer | Task: 在 router.go 中注册 GetKitchenSetting 和 SaveKitchenSetting 路由 | Context: 路由组: /api/v1/shop/setting，需要认证中间件 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 路由注册成功

- [ ] 3.4 编写 API 集成测试

  - File: `main/app/api/v1/shop/shop_setting_test.go`
  - Purpose: 测试 API 接口
  - Requirements: 测试要求
  - Leverage: 现有测试: `main/app/api/v1/shop/`
  - Prompt: Role: QA Engineer specializing in API testing | Task: 为 GetKitchenSetting 和 SaveKitchenSetting API 编写集成测试 | Context: 测试所有 API 接口，测试参数验证，测试响应格式 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 所有 API 测试通过

---

## Phase 4: 权限管理

- [ ] 4.1 执行权限迁移

  - File: `admin/database/migrations/20251209101829_add_kitchen_wait_time_color_access.php`（已创建）
  - Purpose: 执行权限迁移，创建「厨显设置」权限项并赋给所有角色
  - Requirements: 3.1, 3.2
  - Leverage: 现有权限迁移: `admin/database/migrations/20251124014502_init_management_app_access.php`
  - Command: `cd admin && php think migrate:run`
  - Success: 权限迁移执行成功，权限项已创建，所有角色已赋权限

---

## Phase 5: PHP Admin 模块更新

- [x] 5.1 更新 Terminal.php 注释文档和代码

  - File: `admin/app/shop/controller/setting/Terminal.php`
  - Purpose: 在 PHP Admin 模块的厨显设置接口中添加 `wait_time_color_ranges` 字段支持
  - Requirements: 1.7
  - Leverage: 现有 Terminal.php: `admin/app/shop/controller/setting/Terminal.php` (line 485-536)
  - Prompt: Role: PHP Developer | Task: 在 Terminal.php 的 kitchen() 方法中：1) 添加 wait_time_color_ranges 字段的 @Apidoc\Param 注释（包含子参数 minute 和 color），2) 在 $arr 数组中添加 wait_time_color_ranges 字段处理，3) 更新 wait_color 注释说明为旧格式 | Context: 保持向后兼容，同时支持新旧格式 | Restrictions: 遵循 .cursor/rules/php.mdc | Success: 注释文档和代码都已更新，字段处理正确

---

## Phase 6: 数据迁移

- [x] 6.1 创建数据迁移文件

  - File: `admin/database/migrations/{YYYYMMDDHHMMSS}_add_kitchen_wait_time_color_config.php`
  - Purpose: 为所有门店初始化默认等待时长颜色配置
  - Requirements: 5.1（原编号，实际对应 Requirement 5）
  - Leverage: 现有迁移文件: `admin/database/migrations/`，参考 `20251124014502_init_management_app_access.php`
  - Prompt: Role: PHP Developer with ThinkPHP Migration expertise | Task: 创建数据迁移文件，查询所有门店的厨显设置，如无 wait_time_color_ranges 配置则初始化默认配置 | Context: 默认配置：第一区间0分钟黑色 #100A05，第二区间10分钟（从旧配置 wait_color[0] 读取，如 "red" 转 #E50028，"yellow" 转 #FFBE00，如无则使用 #FFBE00），第三区间20分钟（从旧配置 wait_color[1] 读取，如无则使用 #E50028） | Restrictions: 遵循 .cursor/rules/php.mdc，实现幂等性 | Success: 迁移文件创建成功，逻辑正确

- [ ] 6.2 执行数据迁移

  - File: -
  - Purpose: 在数据库中初始化默认配置
  - Requirements: 5.2（原编号，实际对应 Requirement 5）
  - Leverage: Task 6.1 的迁移文件
  - Command: `cd admin && php think migrate:run`
  - Success: 迁移执行成功，所有门店默认配置已初始化

---

## Phase 7: WebSocket 推送

- [x] 7.1 实现 WebSocket 推送逻辑

  - File: `main/app/service/setting/setting.go`
  - Purpose: 在保存配置后推送 WebSocket 配置更新事件
  - Requirements: 4.1, 4.2, 4.3
  - Leverage: 现有 WebSocket 推送: `main/app/service/setting/setting.go`，参考 `EditStoreSetting` 方法 (line 2067)
  - Prompt: Role: Go Developer | Task: 在 SaveKitchenSetting 方法中实现 WebSocket 推送，使用 websocket.PushClient，事件类型 UPDATE_CONFIG，目标 SourceKitchen | Context: 异步推送，不阻塞主流程，推送数据包含 update_time、config_type、config_data | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: WebSocket 推送实现正确

---

## Phase 8: 测试和优化

- [ ] 8.1 集成测试

  - File: `test/integration/kitchen_wait_time_color_test.go`
  - Purpose: 测试端到端功能
  - Requirements: 所有功能需求
  - Leverage: 现有集成测试
  - Prompt: Role: QA Automation Engineer | Task: 实现端到端集成测试，测试配置保存 → WebSocket 推送 → KDS 终端接收流程 | Context: 测试用户完整流程，测试数据一致性 | Restrictions: 测试真实用户场景 | Success: 集成测试通过

- [ ] 8.2 性能测试

  - File: -
  - Purpose: 确保性能达标
  - Requirements: 性能要求
  - Leverage: 性能测试工具（如：wrk, ab）
  - Success: 配置保存响应时间 < 2 秒，WebSocket 推送延迟 < 5 秒

- [ ] 8.3 文档更新

  - File: `docs/shared/api/shop_setting_api.md`, `CHANGELOG.md`
  - Purpose: 确保文档与代码同步
  - Requirements: 文档要求
  - Leverage: `docs/agent/templates/api-doc-template.md`
  - Prompt: Role: Technical Writer | Task: 更新相关文档 | Context: API 文档, 数据库文档, CHANGELOG | Restrictions: 文档准确完整 | Success: 所有文档已更新

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] PHP 代码通过 PSR-2 格式化
- [ ] 测试覆盖率达标
  - Service: ≥ 70%
  - API: ≥ 80%
- [ ] 所有测试通过

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成

### 文档同步

- [ ] API 文档已更新（如有新接口）
- [ ] 数据库文档已更新（如有新表）
- [ ] CHANGELOG.md 已更新

### 规范遵循

- [ ] 遵循 `.cursor/rules/go-main.mdc`
- [ ] 遵循 `.cursor/rules/php.mdc`
- [ ] 遵循 `.cursor/rules/api.mdc`
- [ ] 遵循 `.cursor/rules/database.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-shop-kds-wait-time-color-config/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-shop-kds-wait-time-color-config/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-shop-kds-wait-time-color-config/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/story-shop-kds-wait-time-color-config/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/story-shop-kds-wait-time-color-config/tasks.md)" | bc
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
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 在执行任务过程中若总结出经验或规避策略，请记录 Episode，并在 tasks.md 尾部更新 `Related Episode`。

---

**模板版本**: v1.0.0  
**最后更新**: 2025-12-09  
**维护者**: 后端开发组
