# Grab 外卖导入商品优化 任务分解

> 本文档定义 Grab 外卖导入商品优化功能的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时(SP ≤ 1)
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 29  
**已完成**: 0  
**进行中**: -  
**完成率**: 0%

---

## Phase 1: 数据库设计和迁移

- [ ] 1.1 创建 product_unit 表扩展迁移文件

  - File: `admin/database/migrations/20251211000001_add_source_to_product_unit.php`
  - Purpose: 为 product_unit 表添加 source 和 source_id 字段
  - Requirements: 5.1
  - Leverage: 现有迁移文件: `admin/database/migrations/`
  - Prompt: Role: Database Engineer | Task: 创建迁移文件,为 ttpos_product_unit 表添加 source(varchar 50)和 source_id(varchar 100)字段,并添加索引 | Context: 使用 ThinkPHP 迁移语法 | Restrictions: 遵循 .cursor/rules/database.mdc,迁移前检查字段是否存在 | Success: 迁移文件创建成功,字段定义正确

- [ ] 1.2 创建 tax 表扩展迁移文件

  - File: `admin/database/migrations/20251211000002_add_source_to_tax.php`
  - Purpose: 为 tax 表添加 source 和 source_id 字段
  - Requirements: 5.2
  - Leverage: 现有迁移文件: `admin/database/migrations/`
  - Success: 迁移文件创建成功

- [ ] 1.3 创建 product_flavor 表扩展迁移文件

  - File: `admin/database/migrations/20251211000003_add_source_to_product_flavor.php`
  - Purpose: 为 product_flavor 表添加 source 和 source_id 字段
  - Requirements: 5.3
  - Leverage: 现有迁移文件: `admin/database/migrations/`
  - Success: 迁移文件创建成功

- [ ] 1.4 执行数据库迁移

  - File: -
  - Purpose: 在数据库中添加字段
  - Requirements: 5.1, 5.2, 5.3
  - Leverage: Task 1.1-1.3 的迁移文件
  - Command: `cd admin && php think migrate:run`
  - Success: 迁移执行成功,字段已添加

- [ ] 1.5 更新 ProductUnit Go Model

  - File: `main/app/model/product_unit.go`
  - Purpose: 在 Go Model 中添加 Source 和 SourceId 字段
  - Requirements: 5.1
  - Leverage: 现有 Model: `main/app/model/product_unit.go`
  - Prompt: Role: Go Developer | Task: 在 ProductUnit 结构体中添加 Source 和 SourceId 字段 | Context: 使用 gorm 标签,类型为 string | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: Model 更新成功,字段映射正确

- [ ] 1.6 更新 Tax Go Model

  - File: `main/app/model/tax.go`
  - Purpose: 在 Go Model 中添加 Source 和 SourceId 字段
  - Requirements: 5.2
  - Leverage: 现有 Model: `main/app/model/tax.go`
  - Success: Model 更新成功

- [ ] 1.7 更新 ProductFlavor Go Model

  - File: `main/app/model/product_flavor.go`
  - Purpose: 在 Go Model 中添加 Source 和 SourceId 字段
  - Requirements: 5.3
  - Leverage: 现有 Model: `main/app/model/product_flavor.go`
  - Success: Model 更新成功

---

## Phase 2: 翻译服务实现

- [ ] 2.1 创建翻译服务接口

  - File: `main/pkg/translation/i_translation.go`
  - Purpose: 定义翻译服务接口
  - Requirements: 1.1, 1.2
  - Leverage: 无(新组件)
  - Prompt: Role: Go Developer | Task: 创建 ITranslationService 接口,定义 Translate 和 BatchTranslate 方法 | Context: Translate(text, sourceLang, targetLang) 返回翻译结果 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 接口定义完整,方法签名正确

- [ ] 2.2 实现翻译服务

  - File: `main/pkg/translation/translation.go`
  - Purpose: 实现翻译服务逻辑,集成第三方翻译 API
  - Requirements: 1.1, 1.2, 1.4
  - Leverage: Task 2.1 的接口
  - Prompt: Role: Go Developer with API integration expertise | Task: 实现 translationService,集成 Google Translate API 或 Microsoft Translator API | Context: 支持缓存(Redis),超时 5s,失败降级 | Restrictions: 遵循 .cursor/rules/go-main.mdc,不使用 panic | Success: 翻译服务实现完成,API 集成正确,缓存生效

- [ ] 2.3 编写翻译服务单元测试

  - File: `main/pkg/translation/translation_test.go`
  - Purpose: 测试翻译服务逻辑
  - Requirements: 1.1, 1.2, 1.4
  - Leverage: Task 2.2 的实现
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 TranslationService 编写单元测试,覆盖率 ≥ 80% | Context: 测试正常翻译,测试缓存命中,测试失败降级,测试超时处理 | Restrictions: 使用 mock 翻译 API | Success: 测试覆盖率 ≥ 80%,所有测试通过

---

## Phase 3: 语言映射配置

- [ ] 3.1 创建语言映射配置

  - File: `main/config/language_mapping.go`
  - Purpose: 定义 Grab 和 TTPOS 语言代码映射表
  - Requirements: 1.1
  - Leverage: 无(新配置)
  - Prompt: Role: Go Developer | Task: 创建语言映射配置,定义 LanguageMapping map 和 MapGrabLanguageToTTPOS 函数 | Context: 映射 Grab 语言代码(如 en-US)到 TTPOS 代码(如 en) | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 配置创建成功,映射完整

- [ ] 3.2 编写语言映射单元测试

  - File: `main/config/language_mapping_test.go`
  - Purpose: 测试语言映射逻辑
  - Requirements: 1.1
  - Leverage: Task 3.1 的配置
  - Success: 测试覆盖率 100%,所有映射正确

---

## Phase 4: TakeoutService 扩展

- [x] 4.1 在 TakeoutService 中注入 TranslationService

  - File: `main/app/service/takeout.go`
  - Purpose: 添加翻译服务依赖
  - Requirements: 1.1, 1.2
  - Leverage: 现有 TakeoutService: `main/app/service/takeout.go`,Task 2.1-2.2 的翻译服务
  - Status: ✅ 已完成 - 添加了 translationSrv 字段并在 NewTakeoutSrv 中注入

- [x] 4.2 实现 translateMultiLanguageName 方法

  - File: `main/app/service/takeout.go`
  - Purpose: 实现多语言名称翻译逻辑
  - Requirements: 1.1, 1.2, 1.3, 1.4
  - Leverage: Task 4.1 的翻译服务,Task 3.1 的语言映射
  - Status: ✅ 已完成 - 实现了完整的翻译逻辑和降级策略

- [x] 4.3 修改 syncCategories 方法(使用翻译)

  - File: `main/app/service/takeout.go`
  - Purpose: 在同步分类时使用翻译服务
  - Requirements: 1.1, 1.2
  - Leverage: Task 4.2 的 translateMultiLanguageName 方法,现有 syncCategories 方法
  - Status: ✅ 已完成 - createCategory 和 updateCategory 都已使用翻译服务

- [ ] 4.4 修改 syncProducts 方法(使用翻译)

  - File: `main/app/service/takeout.go`
  - Purpose: 在同步商品时使用翻译服务
  - Requirements: 1.1, 1.2
  - Leverage: Task 4.2 的 translateMultiLanguageName 方法
  - Success: 商品同步使用翻译服务

- [ ] 4.5 修改 syncProductUnit 方法(添加 source 支持)

  - File: `main/app/service/takeout.go`
  - Purpose: 在创建单位时添加 source 和 source_id 字段
  - Requirements: 5.1
  - Leverage: 现有 syncProductUnit 方法,Task 1.5 的 Model
  - Prompt: Role: Go Developer | Task: 修改 syncProductUnit 方法,在 ProductUnitAddReq 中添加 Source 和 SourceId 字段 | Context: Source 为平台名(如 "grab"),SourceId 为 "standard" | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 单位创建时包含 source 信息

- [ ] 4.6 修改 syncProductFlavor 方法(添加 source 支持)

  - File: `main/app/service/takeout.go`
  - Purpose: 在创建规格时添加 source 和 source_id 字段
  - Requirements: 5.3
  - Leverage: 现有 syncProductFlavor 方法,Task 1.7 的 Model
  - Success: 规格创建时包含 source 信息

- [ ] 4.7 实现税率自动创建逻辑(syncProductTax)

  - File: `main/app/service/takeout.go`
  - Purpose: 创建 syncProductTax 方法,自动创建 "0%" / "Grab" 税率
  - Requirements: 5.2
  - Leverage: ProductService 的税率创建方法,Task 1.6 的 Model
  - Prompt: Role: Go Developer | Task: 实现 syncProductTax 方法,查询或创建税率 | Context: 税率为 0%,名称为 "Grab",Source 为平台名,SourceId 为 "standard" | Restrictions: 先查询是否存在,不存在才创建 | Success: 税率自动创建逻辑完成

- [ ] 4.8 移除价格汇率换算逻辑

  - File: `main/app/service/takeout.go`
  - Purpose: 确保价格传输时不进行汇率换算
  - Requirements: 4.1, 4.2
  - Leverage: 现有价格处理逻辑
  - Prompt: Role: Go Developer | Task: 检查并移除价格处理中的汇率换算逻辑 | Context: 价格应直接使用原始数字,单位为分 | Restrictions: 不改变价格存储格式 | Success: 价格不再换算,测试通过

- [ ] 4.9 优化属性组选择范围映射

  - File: `main/app/service/takeout.go`
  - Purpose: 正确映射 Grab SelectionRangeMin/Max 到 TTPOS 属性组
  - Requirements: 5.5, 5.6
  - Leverage: 现有 syncModifierGroups 方法
  - Prompt: Role: Go Developer | Task: 修改 syncModifierGroups 方法,正确映射选择范围 | Context: SelectionRangeMin > 0 → IsMust=1,SelectionRangeMax > 1 → IsOpenInput=true | Restrictions: 增加边界值校验 | Success: 属性组选择范围映射正确

- [ ] 4.10 编写 TakeoutService 单元测试

  - File: `main/app/service/takeout_test.go`
  - Purpose: 测试外卖服务扩展功能
  - Requirements: 所有功能需求
  - Leverage: 现有测试
  - Prompt: Role: QA Engineer | Task: 为 TakeoutService 扩展功能编写单元测试,覆盖率 ≥ 70% | Context: 测试翻译集成,测试 source 字段,测试属性组映射 | Restrictions: 使用 mock 翻译服务 | Success: 测试覆盖率 ≥ 70%,所有测试通过

---

## Phase 5: ProductService 扩展

- [ ] 5.1 扩展 AddProductUnit 方法(支持 Source)

  - File: `main/app/service/product.go`
  - Purpose: 在创建单位时支持 Source 和 SourceId 字段
  - Requirements: 5.1
  - Leverage: 现有 AddProductUnit 方法
  - Prompt: Role: Go Developer | Task: 修改 AddProductUnit 方法,支持 req 中的 Source 和 SourceId 字段 | Context: 在保存 ProductUnit 时设置这两个字段 | Restrictions: 向后兼容,Source 可为空 | Success: AddProductUnit 支持 Source 字段

- [ ] 5.2 扩展 AddProductFlavor 方法(支持 Source)

  - File: `main/app/service/product.go`
  - Purpose: 在创建规格时支持 Source 和 SourceId 字段
  - Requirements: 5.3
  - Leverage: 现有 AddProductFlavor 方法
  - Success: AddProductFlavor 支持 Source 字段

- [ ] 5.3 更新 ProductUnitAddReq DTO

  - File: `main/app/dto/req/product_req.go`
  - Purpose: 在 DTO 中添加 Source 和 SourceId 字段
  - Requirements: 5.1
  - Leverage: 现有 DTO
  - Success: DTO 更新成功

- [ ] 5.4 更新 ProductFlavorAddReq DTO

  - File: `main/app/dto/req/product_req.go`
  - Purpose: 在 DTO 中添加 Source 和 SourceId 字段
  - Requirements: 5.3
  - Leverage: 现有 DTO
  - Success: DTO 更新成功

---

## Phase 6: 外卖开关逻辑优化(前端)

- [ ] 6.1 修改外卖配置页面(移除强关联)

  - File: `admin/views/shop/pages/takeout/config.vue`
  - Purpose: 移除外卖开关与 Grab 配置的强关联
  - Requirements: 2.1, 2.2, 2.3, 2.4
  - Leverage: 现有配置页面
  - Prompt: Role: Frontend Developer with Vue 3 expertise | Task: 修改外卖配置页面,允许未配置 Grab 也能开启外卖 | Context: 移除 Grab 配置的前置检查,增加提示信息 | Restrictions: 遵循 .cursor/rules/vue.mdc | Success: 外卖开关可独立使用

- [ ] 6.2 创建 Grab 配置跳转组件

  - File: `admin/views/shop/components/takeout/GrabLinkButton.vue`
  - Purpose: 创建跳转到 Grab 配置页面的按钮组件
  - Requirements: 3.1, 3.2
  - Leverage: Element Plus Button 组件
  - Prompt: Role: Frontend Developer | Task: 创建 GrabLinkButton 组件,点击跳转到 Grab 配置页面 | Context: 生成带参数的 Grab URL,在新窗口打开 | Restrictions: 遵循 .cursor/rules/vue.mdc | Success: 跳转组件创建成功,链接正确

- [ ] 6.3 在配置页面中集成跳转组件

  - File: `admin/views/shop/pages/takeout/config.vue`
  - Purpose: 在外卖配置页面中显示 Grab 跳转按钮
  - Requirements: 3.1, 3.2
  - Leverage: Task 6.2 的组件
  - Success: 跳转按钮显示正确

- [ ] 6.4 添加配置状态显示

  - File: `admin/views/shop/pages/takeout/config.vue`
  - Purpose: 显示 Grab 配置状态(已配置/未配置/配置中)
  - Requirements: 3.5
  - Leverage: Element Plus Tag 组件
  - Success: 配置状态显示正确

---

## Phase 7: 测试和优化

- [ ] 7.1 集成测试(完整商品导入流程)

  - File: `test/integration/takeout_import_test.go`
  - Purpose: 测试从 Grab 获取数据到 TTPOS 保存的完整流程
  - Requirements: 所有功能需求
  - Leverage: 现有集成测试框架
  - Prompt: Role: QA Automation Engineer | Task: 实现端到端集成测试 | Context: 模拟 Grab API 响应,测试完整导入流程 | Restrictions: 测试真实场景 | Success: 集成测试通过

- [ ] 7.2 翻译服务降级测试

  - File: `main/pkg/translation/translation_test.go`
  - Purpose: 测试翻译失败时的降级逻辑
  - Requirements: 1.4
  - Leverage: Task 2.3 的测试
  - Success: 降级测试通过,使用英文作为降级值

- [ ] 7.3 并发测试(多商户同时导入)

  - File: `test/integration/takeout_concurrent_test.go`
  - Purpose: 测试并发导入场景
  - Requirements: 可靠性要求
  - Leverage: Go goroutines
  - Success: 并发测试通过,无数据冲突

- [ ] 7.4 性能测试

  - File: -
  - Purpose: 测试导入性能是否达标
  - Requirements: 性能要求
  - Leverage: 性能测试工具
  - Command: `go test -bench=. -benchmem`
  - Success: 单商品 < 500ms,批量 100 个 < 10s

- [ ] 7.5 文档更新

  - File: `docs/shared/api/takeout_api.md`, `CHANGELOG.md`
  - Purpose: 更新相关文档
  - Requirements: 文档要求
  - Leverage: 现有文档
  - Success: 文档已更新,API 说明完整

---

## 提交清单

完成所有任务后,请检查:

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] PHP 代码通过 PSR-2 格式化(迁移文件)
- [ ] Vue 代码通过 ESLint 检查
- [ ] 测试覆盖率达标
  - TakeoutService: ≥ 70%
  - TranslationService: ≥ 80%
  - LanguageMapper: 100%
- [ ] 所有测试通过

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成
  - 语言映射与翻译
  - 外卖开关逻辑
  - 配置流程优化
  - 价格处理规则
  - 商品属性映射

### 文档同步

- [ ] API 文档已更新
- [ ] 数据库文档已更新(迁移脚本)
- [ ] CHANGELOG.md 已更新
- [ ] design.md 已更新(如有调整)

### 规范遵循

- [ ] 遵循 `.cursor/rules/go-main.mdc`
- [ ] 遵循 `.cursor/rules/api.mdc`
- [ ] 遵循 `.cursor/rules/database.mdc`
- [ ] 遵循 `.cursor/rules/vue.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-takeout-grab-menu-import-update/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-takeout-grab-menu-import-update/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-takeout-grab-menu-import-update/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/story-takeout-grab-menu-import-update/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/story-takeout-grab-menu-import-update/tasks.md)" | bc
```

### 执行流程

1. **选择任务**: 选择下一个未完成任务
2. **阅读需求**: 查看 requirements.md 中的关联需求
3. **查看复用**: 检查 Leverage 中的可复用代码
4. **使用 AI**: 复制 Prompt 模板,让 AI 生成代码
5. **实现代码**: 按照规范实现功能
6. **运行检查**: `go fmt`, `go vet`, `go test`
7. **标记完成**: 将 `[ ]` 改为 `[x]`
8. **提交代码**: Git commit(参考 `.cursor/rules/version.mdc`)

---

## 附录: 标准 Prompt 模板

### Go 后端开发

```
Role: Go Developer specializing in {具体领域}

Task: {具体任务描述,引用 Requirements}

Context:
- Current file: {文件路径}
- Leverage code: {可复用代码路径}
- Requirements: {需求编号和内容}
- Project specs: 遵循 .cursor/rules/go-main.mdc, .cursor/rules/api.mdc

Restrictions:
- Service 只依赖其他 Service 接口
- Repository 只持有 db 实例
- 不使用 panic,返回 error
- 使用 errors.WithMessage 包装错误

Success Criteria:
- {成功标准1}
- 代码通过 go fmt 和 go vet
- 测试覆盖率达标
```

### Vue 前端开发

```
Role: Frontend Developer with Vue 3 + TypeScript expertise

Task: {具体任务描述}

Context:
- Current file: {文件路径}
- Leverage code: {可复用代码路径}
- Requirements: {需求编号和内容}
- Project specs: 遵循 .cursor/rules/vue.mdc

Restrictions:
- 使用 Vue 3 Composition API
- 使用 TypeScript
- 使用 Element Plus 组件库
- 遵循命名规范

Success Criteria:
- {成功标准1}
- 代码通过 ESLint 检查
```

### 测试工程师

```
Role: QA Engineer with Go testing expertise

Task: {测试任务描述}

Context:
- Target file: {测试目标文件}
- Test file: {测试文件路径}
- Coverage target: ≥ 70% (Service) 或 ≥ 80% (Repository)

Test Cases Required:
- 正常场景测试
- 异常场景测试
- 边界条件测试
- 降级策略测试(翻译失败)

Restrictions:
- 遵循 .cursor/rules/go-main.mdc
- 必须包含边界情况测试

Success Criteria:
- 测试覆盖率达标
- 所有测试通过
- 边界情况已覆盖
```

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板: `docs/agent/templates/graphiti-episode.md`
- 活动日志: `docs/team/activities/weifashi/2025-12/2025-12-11.md`
- 在执行任务过程中若总结出经验或规避策略,请记录 Episode,并在 tasks.md 尾部更新 `Related Episode`。

---

**模板版本**: v1.0.0  
**创建日期**: 2025-12-11  
**作者**: weifashi  
**维护者**: 后端开发组

