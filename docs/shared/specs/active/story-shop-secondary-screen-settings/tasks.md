# 副屏设置接口开发 任务分解

> 本文档定义副屏设置接口开发的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 10  
**已完成**: 8  
**进行中**: -  
**完成率**: 80%

---

## Phase 1: 数据结构扩展

- [x] 1.1 扩展 CashierResp 结构体

  - File: `main/app/dto/resp/setting/cashier_setting.go`
  - Purpose: 添加新字段到收银机设置响应结构体
  - Requirements: Requirement 1, Requirement 2, Requirement 3
  - Leverage: 现有结构体: `main/app/dto/resp/setting/cashier_setting.go`
  - Prompt: Role: Go Developer | Task: 在 CashierResp 结构体中添加三个新字段：NoOrderCarouselInterval string, OrderDisplayMode string, OrderCarouselInterval string | Context: 使用 json 标签，字段名使用 snake_case | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 结构体扩展成功，字段定义正确

- [x] 1.2 扩展 PHP 设置默认值

  - File: `admin/app/common/model/settings/Setting.php`
  - Purpose: 在收银机设置默认值中添加新字段
  - Requirements: Requirement 1, Requirement 2, Requirement 3
  - Leverage: 现有默认值: `admin/app/common/model/settings/Setting.php` 中的 `SettingEnum::CASHIER`
  - Prompt: Role: PHP Developer | Task: 在 cashier 设置的默认值数组中添加三个新字段：no_order_carousel_interval => '10', order_display_mode => 'carousel', order_carousel_interval => '10' | Context: 添加到 values 数组中 | Restrictions: 遵循 .cursor/rules/php.mdc | Success: 默认值扩展成功

---

## Phase 2: Go Service 扩展

- [x] 2.1 扩展 GetCashierSetting 方法

  - File: `main/app/service/setting/setting.go`
  - Purpose: 在获取收银机设置时解析并返回新字段
  - Requirements: Requirement 4
  - Leverage: 现有方法: `main/app/service/setting/setting.go` - `GetCashierSetting()`
  - Prompt: Role: Go Developer | Task: 在 GetCashierSetting 方法中，解析 JSON 时读取新字段（no_order_carousel_interval, order_display_mode, order_carousel_interval），如果字段不存在、为空字符串或为"0"则设置默认值（字符串类型："10" 或 "carousel"） | Context: 在 json.Unmarshal 之后，设置默认值逻辑之前添加 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 方法扩展成功，新字段正确解析和返回

- [ ] 2.2 编写 Service 单元测试

  - File: `main/app/service/setting/setting_test.go`
  - Purpose: 测试扩展后的 GetCashierSetting 方法
  - Requirements: Requirement 4
  - Leverage: 现有测试: `main/app/service/setting/setting_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为扩展后的 GetCashierSetting 编写测试，测试新字段的解析和默认值设置 | Context: 测试包含新字段的 JSON，测试不包含新字段的 JSON（应返回默认值） | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试通过，覆盖率达标

---

## Phase 3: Go API 层扩展（新管理端）

- [x] 3.1 创建 SaveCashierSetting API Handler

  - File: `main/app/api/v1/shop/shop_setting.go`
  - Purpose: 在 main 模块中创建收银机设置保存接口
  - Requirements: Requirement 0, Requirement 1, Requirement 2, Requirement 3
  - Leverage: 现有 API: `main/app/api/v1/shop/shop_setting.go`，Service: `main/app/service/setting/setting.go`
  - Prompt: Role: Go Developer with Gin framework expertise | Task: 在 shop_setting.go 中创建 SaveCashierSetting 方法，接收 SaveCashierSettingReq DTO，调用 DTO 的 Validate() 方法进行参数验证，调用 Service 层的 EditCashierSetting 方法保存 | Context: 参数验证通过 DTO 的 Validate() 方法完成，包括 no_order_carousel_interval 和 order_carousel_interval 范围10-120，order_display_mode 枚举值 carousel/order/order_carousel，轮播内容（carousel）数量最多15个 | Restrictions: 遵循 .cursor/rules/go-main.mdc，URL 使用 snake_case，data 必须是对象 | Success: API 创建成功，调用 DTO 验证方法，调用 Service 保存方法

- [x] 3.3 创建 Request DTO

  - File: `main/app/dto/req/cashier_setting.go`
  - Purpose: 定义收银机设置保存请求 DTO，包含 Validate() 方法进行参数验证
  - Requirements: Requirement 0
  - Leverage: 现有 DTO: `main/app/dto/req/`
  - Prompt: Role: Go Developer | Task: 创建 SaveCashierSettingReq 结构体，包含 carousel、no_order_carousel_interval（string类型）、order_display_mode、order_carousel_interval（string类型）字段，并实现 Validate() 方法进行参数验证 | Context: 使用 json 标签，字段名使用 snake_case，Validate() 方法处理轮播间隔字段时：如果为空字符串或"0"则设置为默认值"10"，否则将字符串转换为 int 后验证范围（10-120），验证展示模式枚举值（carousel/order/order_carousel），验证轮播内容数量限制（最多15个） | Restrictions: 遵循 .cursor/rules/go-main.mdc，使用 errors.WithMessage 返回错误，使用 strconv.Atoi 转换字符串 | Success: DTO 创建成功，Validate() 方法实现正确

- [x] 3.4 注册 API 路由

  - File: `main/app/api/v1/shop/shop_setting.go` - `RegisterSettingHandlers` 函数
  - Purpose: 注册收银机设置保存接口路由
  - Requirements: Requirement 0
  - Leverage: 现有路由注册: `main/app/api/v1/shop/shop_setting.go`
  - Success: 路由注册成功，路径为 `/api/v1/shop/setting/cashier/carousel/upload` 和 `/api/v1/shop/setting/cashier`

- [ ] 3.5 编写 API 单元测试

  - File: `main/app/api/v1/shop/shop_setting_test.go`
  - Purpose: 测试 SaveCashierSetting API
  - Requirements: Requirement 0, Requirement 1, Requirement 2, Requirement 3
  - Leverage: 现有测试: `main/app/api/v1/shop/*_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 SaveCashierSetting API 编写测试，测试 DTO 的 Validate() 方法验证逻辑和保存逻辑 | Context: 测试正常保存，测试 Validate() 方法参数验证失败场景（轮播内容数量限制、轮播间隔范围、展示模式枚举值），测试 Service 层保存逻辑 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试通过

---

## Phase 4: 测试和文档

- [ ] 4.1 集成测试

  - File: `test/integration/cashier_setting_test.go`
  - Purpose: 测试端到端功能
  - Requirements: 所有功能需求
  - Leverage: 现有集成测试
  - Prompt: Role: QA Automation Engineer | Task: 实现端到端集成测试 | Context: 测试保存新字段，测试查询返回新字段，测试默认值，测试向后兼容 | Restrictions: 测试真实用户场景 | Success: 集成测试通过

- [ ] 4.2 向后兼容测试

  - File: `test/integration/cashier_setting_backward_compatibility_test.go`
  - Purpose: 确保现有功能不受影响
  - Requirements: 所有功能需求
  - Leverage: 现有测试
  - Prompt: Role: QA Engineer | Task: 测试向后兼容性 | Context: 测试不包含新字段的旧数据，测试现有接口返回正常，测试现有功能不受影响 | Restrictions: 确保现有功能正常 | Success: 向后兼容测试通过

- [ ] 4.3 文档更新

  - File: `docs/shared/api/shop_api.md`, `CHANGELOG.md`
  - Purpose: 确保文档与代码同步
  - Requirements: 文档要求
  - Leverage: `docs/agent/templates/api-doc-template.md`
  - Prompt: Role: Technical Writer | Task: 更新相关文档 | Context: API 文档说明新增字段, CHANGELOG 记录变更 | Restrictions: 文档准确完整 | Success: 所有文档已更新

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] PHP 代码通过 PSR-2 格式化
- [ ] 测试覆盖率达标
  - Service: ≥ 70%
- [ ] 所有测试通过

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成
- [ ] 向后兼容：现有功能不受影响

### 文档同步

- [ ] API 文档已更新（说明新增字段）
- [ ] CHANGELOG.md 已更新

### 规范遵循

- [ ] 遵循 `.cursor/rules/go-main.mdc`
- [ ] 遵循 `.cursor/rules/php.mdc`
- [ ] 遵循 `.cursor/rules/api.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-shop-secondary-screen-settings/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-shop-secondary-screen-settings/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-shop-secondary-screen-settings/tasks.md
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
**最后更新**: 2025-11-20  
**维护者**: 后端开发组

