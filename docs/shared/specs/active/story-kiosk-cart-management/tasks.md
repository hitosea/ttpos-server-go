# Kiosk 自助点餐机购物车管理模块 任务分解

> 本文档定义 Kiosk 自助点餐机购物车管理模块的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 18  
**已完成**: 15  
**进行中**: -  
**完成率**: 83%

---

## Phase 1: API 层实现 - 购物车信息查询

- [x] 1.1 创建 Order Handler 结构体

  - File: `main/app/api/v1/kiosk/kiosk_order.go`
  - Purpose: 创建购物车管理 API Handler
  - Requirements: 1.1
  - Leverage: 现有 Handler: `main/app/api/v1/cashier/cashier_instant.go` - `InstantHandler` 结构体
  - Prompt: Role: Go Developer with Gin framework expertise | Task: 创建 OrderHandler 结构体，包含 orderSrv 和 orderBaseSrv 依赖 | Context: 参考 InstantHandler 的结构，使用 NewOrderHandler() 构造函数 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: Handler 结构体创建成功，依赖注入正确

- [x] 1.2 实现购物车信息查询 API

  - File: `main/app/api/v1/kiosk/kiosk_order.go`
  - Purpose: 实现 GET `/api/v1/kiosk/order/cart/info` 接口
  - Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 1.6
  - Leverage: 现有 API: `main/app/api/v1/cashier/cashier_instant.go` - `OrderCartInfo()` 方法
  - Prompt: Role: Go Developer with Gin framework expertise | Task: 实现 GetCartInfo() 方法，调用 orderSrv.GetOrderCartInfo() | Context: 参考 OrderCartInfo() 的实现，支持通过 sale_bill_uuid 查询，使用 helper.Success() 返回响应 | Restrictions: 遵循 .cursor/rules/api.mdc，URL 使用 snake_case，data 必须是对象 | Success: API 创建成功，响应格式正确，参数验证正确

- [x] 1.3 注册购物车信息查询路由

  - File: `main/router/router.go`
  - Purpose: 注册 `/api/v1/kiosk/order/cart/info` 路由
  - Requirements: 1.1
  - Leverage: 现有路由: `main/router/router.go` - kioskGroup 路由组
  - Prompt: Role: Go Developer | Task: 在 router.go 的 kioskGroup 中注册 GetCartInfo 路由 | Context: 调用 kiosk.RegisterOrderHandlers(kioskGroup, dbm, cache) | Restrictions: 遵循现有路由注册模式 | Success: 路由注册成功，接口可访问

---

## Phase 2: API 层实现 - 商品添加

- [x] 2.1 实现商品添加 API

  - File: `main/app/api/v1/kiosk/kiosk_order.go`
  - Purpose: 实现 POST `/api/v1/kiosk/order/cart/product/add` 接口
  - Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 2.6, 2.7
  - Leverage: 现有 API: `main/app/api/v1/cashier/cashier_instant.go` - `OrderCartProductAdd()` 方法
  - Leverage: 现有 DTO: `main/app/dto/req/shop_cart.go` - `OrderCartProductAddReq`
  - Prompt: Role: Go Developer with Gin framework expertise | Task: 实现 AddProduct() 方法，调用 orderSrv.AddProductToCart() | Context: 参考 OrderCartProductAdd() 的实现，绑定 OrderCartProductAddReq 参数，验证商品状态和库存，使用 helper.Success() 返回响应 | Restrictions: 遵循 .cursor/rules/api.mdc，URL 使用 snake_case，data 必须是对象 | Success: API 创建成功，响应格式正确，参数验证正确，商品验证正确

- [x] 2.2 注册商品添加路由

  - File: `main/router/router.go`
  - Purpose: 注册 `/api/v1/kiosk/order/cart/product/add` 路由
  - Requirements: 2.1
  - Leverage: 现有路由: `main/router/router.go` - kioskGroup 路由组
  - Success: 路由注册成功，接口可访问

---

## Phase 3: API 层实现 - 套餐添加

- [x] 3.1 实现套餐添加 API

  - File: `main/app/api/v1/kiosk/kiosk_order.go`
  - Purpose: 实现 POST `/api/v1/kiosk/order/cart/product_package/add` 接口
  - Requirements: 3.1, 3.2, 3.3, 3.4, 3.5
  - Leverage: 现有 API: `main/app/api/v1/cashier/cashier_instant.go` - `OrderCartProductPackageAdd()` 方法
  - Leverage: 现有 DTO: `main/app/dto/req/shop_cart.go` - `OrderCartProductPackageAddReq`
  - Prompt: Role: Go Developer with Gin framework expertise | Task: 实现 AddProductPackage() 方法，调用 orderSrv.AddProductPackageToCart() | Context: 参考 OrderCartProductPackageAdd() 的实现，绑定 OrderCartProductPackageAddReq 参数，验证套餐配置完整性，使用 helper.Success() 返回响应 | Restrictions: 遵循 .cursor/rules/api.mdc，URL 使用 snake_case，data 必须是对象 | Success: API 创建成功，响应格式正确，参数验证正确，套餐验证正确

- [x] 3.2 注册套餐添加路由

  - File: `main/router/router.go`
  - Purpose: 注册 `/api/v1/kiosk/order/cart/product_package/add` 路由
  - Requirements: 3.1
  - Leverage: 现有路由: `main/router/router.go` - kioskGroup 路由组
  - Success: 路由注册成功，接口可访问

---

## Phase 4: API 层实现 - 商品数量修改

- [x] 4.1 实现商品数量修改 API

  - File: `main/app/api/v1/kiosk/kiosk_order.go`
  - Purpose: 实现 POST `/api/v1/kiosk/order/cart/product/num` 接口
  - Requirements: 4.1, 4.2, 4.3, 4.4, 4.5
  - Leverage: 现有 API: `main/app/api/v1/cashier/cashier_instant.go` - `OrderCartProductNum()` 方法
  - Leverage: 现有 DTO: `main/app/dto/req/shop_cart.go` - `OrderCartProductNumReq`
  - Prompt: Role: Go Developer with Gin framework expertise | Task: 实现 UpdateProductNum() 方法，调用 orderSrv.UpdateProductNum() | Context: 参考 OrderCartProductNum() 的实现，绑定 OrderCartProductNumReq 参数，验证数量范围，使用 helper.Success() 返回响应 | Restrictions: 遵循 .cursor/rules/api.mdc，URL 使用 snake_case，data 必须是对象 | Success: API 创建成功，响应格式正确，参数验证正确，数量验证正确

- [x] 4.2 注册商品数量修改路由

  - File: `main/router/router.go`
  - Purpose: 注册 `/api/v1/kiosk/order/cart/product/num` 路由
  - Requirements: 4.1
  - Leverage: 现有路由: `main/router/router.go` - kioskGroup 路由组
  - Success: 路由注册成功，接口可访问

---

## Phase 5: API 层实现 - 商品选购详情

- [x] 5.1 实现商品选购详情 API

  - File: `main/app/api/v1/kiosk/kiosk_order.go`
  - Purpose: 实现 GET `/api/v1/kiosk/order/product/package/detail` 接口
  - Requirements: 5.1, 5.2, 5.3, 5.4
  - Leverage: 现有 API: `main/app/api/v1/assistant/assistant_order.go` - `GetProductPackageDetail()` 方法
  - Leverage: 现有 DTO: `main/app/dto/req/shop_cart.go` - `GetProductPackageDetailReq`
  - Leverage: 现有 DTO: `main/app/dto/resp/shop_cart.go` - `ProductPackageDetailRes`
  - Prompt: Role: Go Developer with Gin framework expertise | Task: 实现 GetProductPackageDetail() 方法，调用 orderBaseSrv.GetProductPackageDetail() | Context: 参考 GetProductPackageDetail() 的实现，绑定 GetProductPackageDetailReq 参数（使用 ShouldBindQuery），验证商品是否存在于购物车，使用 helper.Success() 返回响应 | Restrictions: 遵循 .cursor/rules/api.mdc，URL 使用 snake_case，data 必须是对象 | Success: API 创建成功，响应格式正确，参数验证正确

- [x] 5.2 注册商品选购详情路由

  - File: `main/router/router.go`
  - Purpose: 注册 `/api/v1/kiosk/order/product/package/detail` 路由
  - Requirements: 5.1
  - Leverage: 现有路由: `main/router/router.go` - kioskGroup 路由组
  - Success: 路由注册成功，接口可访问

---

## Phase 6: API 层实现 - 商品删除

- [x] 6.1 实现商品删除 API

  - File: `main/app/api/v1/kiosk/kiosk_order.go`
  - Purpose: 实现 DELETE `/api/v1/kiosk/order/cart/product/delete` 接口
  - Requirements: 6.1, 6.2, 6.3, 6.4, 6.5
  - Leverage: 现有 API: `main/app/api/v1/cashier/cashier_instant.go` - 删除商品相关方法
  - Leverage: 现有 Service: `main/app/service/order.go` - `DeleteProduct()` 方法
  - Prompt: Role: Go Developer with Gin framework expertise | Task: 实现 DeleteProduct() 方法，调用 orderSrv.DeleteProduct() | Context: 参考删除商品的实现，通过 sale_order_product_uuid 删除商品，验证商品是否存在于购物车，使用 helper.Success() 返回响应 | Restrictions: 遵循 .cursor/rules/api.mdc，URL 使用 snake_case，data 必须是对象 | Success: API 创建成功，响应格式正确，参数验证正确

- [x] 6.2 注册商品删除路由

  - File: `main/router/router.go`
  - Purpose: 注册 `/api/v1/kiosk/order/cart/product/delete` 路由
  - Requirements: 6.1
  - Leverage: 现有路由: `main/router/router.go` - kioskGroup 路由组
  - Success: 路由注册成功，接口可访问

---

## Phase 7: 路由注册和 Handler 注册函数

- [x] 7.1 创建 RegisterOrderHandlers 函数

  - File: `main/app/api/v1/kiosk/kiosk_order.go`
  - Purpose: 创建路由注册函数，统一注册所有购物车相关路由
  - Requirements: 所有 API 需求
  - Leverage: 现有注册函数: `main/app/api/v1/kiosk/kiosk_base.go` - `RegisterBaseHandlers()` 函数
  - Prompt: Role: Go Developer | Task: 创建 RegisterOrderHandlers() 函数，注册所有购物车相关路由 | Context: 参考 RegisterBaseHandlers() 的实现，创建 OrderHandler 实例，注册所有路由 | Restrictions: 遵循现有路由注册模式 | Success: 注册函数创建成功，所有路由注册正确

- [x] 7.2 在 router.go 中调用 RegisterOrderHandlers

  - File: `main/router/router.go`
  - Purpose: 在路由初始化时调用 RegisterOrderHandlers
  - Requirements: 所有 API 需求
  - Leverage: 现有路由: `main/router/router.go` - kioskGroup 路由组
  - Success: 路由注册成功，所有接口可访问

---

## Phase 8: 测试

- [ ] 8.1 编写 API 集成测试

  - File: `main/app/api/v1/kiosk/kiosk_order_test.go`
  - Purpose: 测试所有购物车管理 API 接口
  - Requirements: 测试要求
  - Leverage: 现有测试: `main/app/api/v1/cashier/cashier_instant_test.go`
  - Prompt: Role: QA Engineer specializing in API testing | Task: 为所有购物车管理 API 编写集成测试 | Context: 测试所有 API 接口，测试参数验证，测试响应格式，测试错误处理 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 所有 API 测试通过

- [ ] 8.2 端到端流程测试

  - File: `test/integration/kiosk_cart_test.go`
  - Purpose: 测试完整的购物车管理流程
  - Requirements: 集成测试要求
  - Leverage: 现有集成测试: `test/integration/`
  - Prompt: Role: QA Automation Engineer | Task: 实现端到端集成测试 | Context: 测试完整流程：添加商品 → 修改数量 → 查看详情 → 删除商品 → 查看购物车 | Restrictions: 测试真实用户场景 | Success: 集成测试通过

---

## Phase 9: 性能优化和缓存

- [ ] 9.1 实现购物车信息缓存

  - File: `main/app/service/order.go`
  - Purpose: 实现购物车信息的 Redis 缓存
  - Requirements: 性能要求
  - Leverage: 现有缓存实现: `main/app/service/` 中的缓存使用方式
  - Prompt: Role: Go Developer with Redis expertise | Task: 在 GetOrderCartInfo() 方法中实现缓存，使用 Cache-Aside Pattern | Context: Key 格式为 `ttpos:kiosk:cart:{sale_bill_uuid}`，过期时间 5 分钟，缓存未命中时查询数据库并写入缓存 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 缓存机制实现完成，缓存命中率 > 80%

- [ ] 9.2 性能测试和优化

  - File: -
  - Purpose: 确保性能达标
  - Requirements: 性能要求
  - Leverage: 性能测试工具（如：wrk, ab）
  - Success: 本地响应时间 < 200ms，数据库查询 < 50ms

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] 测试覆盖率达标
  - Service: ≥ 70%
  - Repository: ≥ 80%
  - Order 相关: 100%
- [ ] 所有测试通过

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成

### 文档同步

- [ ] API 文档已更新（如有新接口）
- [ ] CHANGELOG.md 已更新

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
grep -c "^- \[" docs/shared/specs/active/story-kiosk-cart-management/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-kiosk-cart-management/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-kiosk-cart-management/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/story-kiosk-cart-management/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/story-kiosk-cart-management/tasks.md)" | bc
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
**最后更新**: 2025-12-18  
**维护者**: 后端开发组

