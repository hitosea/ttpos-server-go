# Kiosk 自助点餐机商品浏览模块 任务分解

> 本文档定义 Kiosk 自助点餐机商品浏览模块的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 12  
**已完成**: 7  
**进行中**: -  
**完成率**: 58%

---

## Phase 1: Service 层扩展

### Service 层

- [x] 1.1 在 Product Service 中添加 Kiosk 终端支持

  - File: `main/app/service/product.go`
  - Purpose: 在 `GetProductList()` 方法中添加 `SourceKiosk` 的支持
  - Requirements: 2.4
  - Leverage: 现有方法: `main/app/service/product.go` - `GetProductList()` 方法中的 `sourceMap`
  - Prompt: Role: Go Developer specializing in Service Layer | Task: 在 GetProductList() 方法的 sourceMap 中添加 SourceKiosk 支持，使用 commonRepo.WhereByIsShowKiosk(1) | Context: 参考 SourceTablet、SourceMember 的实现方式，添加 constant.SourceKiosk 映射 | Restrictions: 遵循 .cursor/rules/go-main.mdc，Service 只依赖其他 Service 接口 | Success: SourceKiosk 支持添加成功，商品列表能正确筛选 Kiosk 终端显示的商品

- [x] 1.2 检查并添加 Repository WhereByIsShowKiosk 方法（如需要）

  - File: `main/app/repository/common.go`
  - Purpose: 检查 CommonRepo 是否有 WhereByIsShowKiosk() 方法，如果没有则添加
  - Requirements: 2.4
  - Leverage: 现有方法: `main/app/repository/common.go` - `WhereByIsShowCashier()`, `WhereByIsShowTablet()` 等方法
  - Prompt: Role: Go Developer | Task: 检查 CommonRepo 是否有 WhereByIsShowKiosk() 方法，如果没有则添加，参考 WhereByIsShowTablet() 的实现 | Context: 方法应该返回 DBOption，筛选 is_show_kiosk=1 的商品 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: WhereByIsShowKiosk() 方法存在或已添加，功能正确

---

## Phase 2: API 层实现

### Product API

- [x] 2.1 创建 Product Handler 文件

  - File: `main/app/api/v1/kiosk/kiosk_product.go`
  - Purpose: 创建商品相关的 Handler 文件
  - Requirements: 1.1, 2.1, 3.1
  - Leverage: 现有 API: `main/app/api/v1/tablet/tablet_product.go`, `main/app/api/v1/member/member_product.go`
  - Prompt: Role: Go Developer with Gin framework expertise | Task: 创建 kiosk_product.go 文件，定义 ProductHandler 结构体和 RegisterProductHandlers() 函数 | Context: ProductHandler 包含 productSrv 字段，RegisterProductHandlers() 初始化服务并注册路由 | Restrictions: 遵循 .cursor/rules/api.mdc，URL 使用 snake_case | Success: 文件创建成功，结构体定义正确

- [x] 2.2 实现 GetProductCategoryList Handler

  - File: `main/app/api/v1/kiosk/kiosk_product.go`
  - Purpose: 实现获取商品分类列表 API 接口
  - Requirements: 1.1, 1.2, 1.3, 1.4, 1.5
  - Leverage: 现有 API: `main/app/api/v1/tablet/tablet_product.go` - `GetProductCategoryList()` 方法
  - Prompt: Role: Go Developer with Gin framework expertise | Task: 实现 GetProductCategoryList() 方法，调用 productSrv.GetProductCategoryList()，使用 helper.Success() 返回响应 | Context: 使用 ctx.GetDbId() 获取数据库ID，错误处理使用 helper.ErrorWithDetail() | Restrictions: 遵循 .cursor/rules/api.mdc，响应格式正确 | Success: API 创建成功，响应格式正确，错误处理正确

- [x] 2.3 实现 GetProductList Handler

  - File: `main/app/api/v1/kiosk/kiosk_product.go`
  - Purpose: 实现获取商品列表 API 接口
  - Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 2.6, 2.7, 2.8, 2.9
  - Leverage: 现有 API: `main/app/api/v1/tablet/tablet_product.go` - `GetProductList()` 方法
  - Prompt: Role: Go Developer with Gin framework expertise | Task: 实现 GetProductList() 方法，绑定查询参数（page_no, page_size, category_uuid），调用 productSrv.GetProductList() | Context: 使用 c.ShouldBindQuery() 绑定参数，使用 helper.HandleValidationError() 处理验证错误 | Restrictions: 遵循 .cursor/rules/api.mdc，参数验证正确 | Success: API 创建成功，参数验证正确，响应格式正确

- [x] 2.4 实现 GetProductDetail Handler

  - File: `main/app/api/v1/kiosk/kiosk_product.go`
  - Purpose: 实现获取商品详情 API 接口
  - Requirements: 3.1, 3.2, 3.3, 3.4, 3.5, 3.6, 3.7, 3.8
  - Leverage: 现有 API: `main/app/api/v1/member/member_product.go` - `GetProductDetail()` 方法
  - Prompt: Role: Go Developer with Gin framework expertise | Task: 实现 GetProductDetail() 方法，绑定查询参数（uuid），调用 productSrv.GetProductDetail() | Context: 使用 c.ShouldBindQuery() 绑定参数，uuid 为必填参数 | Restrictions: 遵循 .cursor/rules/api.mdc，参数验证正确 | Success: API 创建成功，参数验证正确，响应格式正确

- [x] 2.5 注册 Product API 路由

  - File: `main/router/router.go`
  - Purpose: 注册商品相关路由
  - Requirements: 1.1, 2.1, 3.1
  - Leverage: 现有路由: `main/router/router.go` - kioskGroup 路由组
  - Prompt: Role: Go Developer | Task: 在 router.go 的 kioskGroup 中注册 product handlers | Context: 调用 kiosk.RegisterProductHandlers(kioskGroup, dbm, cache) | Restrictions: 遵循现有路由注册模式 | Success: 路由注册成功，接口可访问

---

## Phase 3: 缓存实现

- [ ] 3.1 实现商品分类列表缓存

  - File: `main/app/service/product.go`
  - Purpose: 实现商品分类列表的 Redis 缓存，提升性能
  - Requirements: 性能要求
  - Leverage: 现有缓存实现: `main/app/service/` 中的缓存使用方式
  - Prompt: Role: Go Developer with Redis expertise | Task: 在 GetProductCategoryList() 方法中实现缓存，使用 Cache-Aside Pattern | Context: Key 格式为 `ttpos:kiosk:product:category:list:{company_uuid}`，过期时间 5 分钟 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 缓存机制实现完成，缓存命中率 > 80%

- [ ] 3.2 实现商品列表缓存

  - File: `main/app/service/product.go`
  - Purpose: 实现商品列表的 Redis 缓存，提升性能
  - Requirements: 性能要求
  - Leverage: 现有缓存实现: `main/app/service/` 中的缓存使用方式
  - Prompt: Role: Go Developer with Redis expertise | Task: 在 GetProductList() 方法中实现缓存，使用 Cache-Aside Pattern | Context: Key 格式为 `ttpos:kiosk:product:list:{company_uuid}:{category_uuid}:{page_no}:{page_size}`，过期时间 5 分钟 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 缓存机制实现完成，缓存命中率 > 80%

- [ ] 3.3 实现商品详情缓存

  - File: `main/app/service/product.go`
  - Purpose: 实现商品详情的 Redis 缓存，提升性能
  - Requirements: 性能要求
  - Leverage: 现有缓存实现: `main/app/service/` 中的缓存使用方式
  - Prompt: Role: Go Developer with Redis expertise | Task: 在 GetProductDetail() 方法中实现缓存，使用 Cache-Aside Pattern | Context: Key 格式为 `ttpos:kiosk:product:detail:{product_uuid}`，过期时间 5 分钟 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 缓存机制实现完成，缓存命中率 > 80%

---

## Phase 4: 测试和优化

- [ ] 4.1 编写 Service 单元测试

  - File: `main/app/service/product_test.go`
  - Purpose: 确保 GetProductList() 方法中 Kiosk 终端支持正确
  - Requirements: 测试要求
  - Leverage: 现有测试: `main/app/service/product_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 GetProductList() 方法编写单元测试，测试 Kiosk 终端筛选逻辑 | Context: 测试正常场景、Kiosk 终端筛选场景、分类筛选场景 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 ≥ 70%，所有测试通过

- [ ] 4.2 编写 API 集成测试

  - File: `main/app/api/v1/kiosk/kiosk_product_test.go`
  - Purpose: 测试 API 接口
  - Requirements: 测试要求
  - Leverage: 现有测试: `main/app/api/v1/tablet/tablet_product_test.go`, `main/app/api/v1/member/member_product_test.go`
  - Prompt: Role: QA Engineer specializing in API testing | Task: 为 kiosk product API 编写集成测试 | Context: 测试所有 API 接口，测试参数验证，测试响应格式，测试错误处理，测试多语言支持 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 所有 API 测试通过

- [ ] 4.3 端到端集成测试

  - File: `test/integration/kiosk_product_browse_test.go`
  - Purpose: 测试端到端功能流程
  - Requirements: 所有功能需求
  - Leverage: 现有集成测试: `test/integration/`
  - Prompt: Role: QA Automation Engineer | Task: 实现端到端集成测试 | Context: 测试用户登录后获取商品分类列表，测试商品列表浏览，测试商品详情查看，测试多语言切换，测试分页加载 | Restrictions: 测试真实用户场景 | Success: 集成测试通过

- [ ] 4.4 性能优化和缓存优化

  - File: `main/app/service/product.go`
  - Purpose: 优化数据库查询和缓存策略
  - Requirements: 性能要求
  - Leverage: 现有优化: `main/app/service/product.go` 中的查询优化方式
  - Prompt: Role: Performance Engineer | Task: 优化商品浏览相关方法的性能，减少数据库查询次数，提升缓存命中率 | Context: 批量获取数据，使用索引查询，优化缓存 Key 设计 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 性能优化完成，响应时间 < 200ms，缓存命中率 > 80%

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] 测试覆盖率达标
  - Service: ≥ 70%
  - API: ≥ 80%
- [ ] 所有测试通过

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成

### 文档同步

- [ ] API 文档已更新
- [ ] CHANGELOG.md 已更新（如有）
- [ ] 设计文档已更新（如有变更）

### 规范遵循

- [ ] 遵循 `.cursor/rules/go-main.mdc`
- [ ] 遵循 `.cursor/rules/api.mdc`
- [ ] 遵循 `.cursor/rules/database.mdc`
- [ ] 遵循 `.cursor/rules/security.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-kiosk-product-browse/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-kiosk-product-browse/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-kiosk-product-browse/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/story-kiosk-product-browse/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/story-kiosk-product-browse/tasks.md)" | bc
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
- 活动日志：`docs/team/activities/{user}/{YYYY-MM}/{YYYY-MM-DD}.md`
- 在执行任务过程中若总结出经验或规避策略，请记录 Episode，并在 tasks.md 尾部更新 `Related Episode`。

---

**模板版本**: v1.0.0  
**最后更新**: 2025-12-18  
**维护者**: 后端开发组

