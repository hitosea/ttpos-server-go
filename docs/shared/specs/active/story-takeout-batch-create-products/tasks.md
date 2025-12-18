# 批量创建外卖商品 任务分解

> 本文档定义批量创建外卖商品功能的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时(SP ≤ 1)
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 任务统计

**总任务数**: 10  
**已完成**: 0  
**进行中**: -  
**完成率**: 0%

---

## Phase 1: DTO 层

- [ ] 1.1 创建 Request DTO

  - File: `main/app/dto/req/takeout_batch_req.go`
  - Purpose: 定义批量操作请求参数
  - Requirements: Requirement 1-4
  - Leverage: 现有 DTO: `main/app/dto/req/product_takeout_req.go`
  - Prompt: Role: Go Developer | Task: 创建批量操作 Request DTO,包含 TakeoutBatchCreateReq, TakeoutBatchOnlineReq, TakeoutBatchOfflineReq, TakeoutBatchDeleteReq 结构体 | Context: 包含 Platform string, ProductUuids []uint64 字段,使用 binding 标签验证参数 | Restrictions: ProductUuids 最多100个,Platform 必须是 grab/lineman | Success: DTO 创建成功,validation 正确

- [ ] 1.2 创建 Response DTO

  - File: `main/app/dto/resp/takeout_batch_resp.go`
  - Purpose: 定义批量操作响应数据
  - Requirements: Requirement 1-4
  - Leverage: 现有 DTO: `main/app/dto/resp/product_resp/product_takeout_resp.go`
  - Prompt: Role: Go Developer | Task: 创建批量操作 Response DTO,包含 TakeoutBatchResp 结构体 | Context: 包含 Total, Success, Failed, FailedProducts 字段 | Restrictions: data 必须是对象,不能是 null | Success: DTO 创建成功,响应格式正确

---

## Phase 2: Service 层

- [ ] 2.1 扩展 IProductTakeoutSrv 接口

  - File: `main/app/service/product_takeout.go`
  - Purpose: 添加批量操作方法到 Service 接口
  - Requirements: Requirement 1-4
  - Leverage: 现有 Service 接口: `IProductTakeoutSrv`
  - Prompt: Role: Go Developer | Task: 在 IProductTakeoutSrv 接口中添加批量操作方法 | Context: 添加 BatchCreateProducts, BatchOnlineProducts, BatchOfflineProducts, BatchDeleteProducts 方法 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 接口方法添加成功,方法签名正确

- [ ] 2.2 实现批量创建商品

  - File: `main/app/service/product_takeout.go`
  - Purpose: 实现批量创建商品业务逻辑
  - Requirements: Requirement 1
  - Leverage: 现有 productTakeoutSrv 实现,复用单商品推送逻辑
  - Prompt: Role: Go Developer | Task: 实现 BatchCreateProducts 方法,支持并发批量创建 | Context: 查询商品列表,使用 Goroutine 并发处理,实现限流控制(每秒10个),失败重试3次,使用 sync.WaitGroup 等待完成,直接返回结果 | Restrictions: 不使用 panic,返回 error,遵循 .cursor/rules/go-main.mdc | Success: 批量创建逻辑实现完整,并发处理正常工作,限流和重试机制生效

- [ ] 2.3 实现批量上架商品

  - File: `main/app/service/product_takeout.go`
  - Purpose: 实现批量上架商品业务逻辑
  - Requirements: Requirement 2
  - Leverage: Task 2.2 的实现模式,复用外卖平台API调用
  - Prompt: Role: Go Developer | Task: 实现 BatchOnlineProducts 方法,支持并发批量上架 | Context: 检查商品是否已同步到平台,调用平台上架API,更新商品状态,使用并发处理 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 批量上架逻辑实现完整,状态更新正确

- [ ] 2.4 实现批量下架商品

  - File: `main/app/service/product_takeout.go`
  - Purpose: 实现批量下架商品业务逻辑
  - Requirements: Requirement 3
  - Leverage: Task 2.2, 2.3 的实现模式
  - Prompt: Role: Go Developer | Task: 实现 BatchOfflineProducts 方法,支持并发批量下架 | Context: 检查商品是否已同步到平台,调用平台下架API,更新商品状态,使用并发处理 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 批量下架逻辑实现完整,状态更新正确

- [ ] 2.5 实现批量删除商品

  - File: `main/app/service/product_takeout.go`
  - Purpose: 实现批量删除商品业务逻辑
  - Requirements: Requirement 4
  - Leverage: Task 2.2, 2.3, 2.4 的实现模式
  - Prompt: Role: Go Developer | Task: 实现 BatchDeleteProducts 方法,支持并发批量删除 | Context: 检查商品是否已同步到平台,调用平台删除API,软删除 ttpos_product_takeout 记录,使用并发处理 | Restrictions: 遵循 .cursor/rules/go-main.mdc,使用软删除 | Success: 批量删除逻辑实现完整,软删除正确

---

## Phase 3: API 层

- [ ] 3.1 添加批量操作路由

  - File: `main/app/api/v1/shop/shop_takeout.go`
  - Purpose: 添加批量操作 HTTP 接口
  - Requirements: Requirement 1-4
  - Leverage: 现有路由: `shop_takeout.go`,Task 2.1-2.5 的 Service 实现
  - Prompt: Role: Go Developer | Task: 添加批量操作 Handler 方法和路由注册 | Context: 实现 BatchCreateProducts, BatchOnlineProducts, BatchOfflineProducts, BatchDeleteProducts Handler,添加 Swagger 注释,注册路由 | Restrictions: 遵循 .cursor/rules/go-main.mdc,URL 使用 snake_case,添加 middleware.Auth | Success: Handler 实现完整,Swagger 注释完整,路由注册成功

---

## Phase 4: 测试

- [ ] 4.1 编写 Service 单元测试

  - File: `main/app/service/product_takeout_batch_test.go`
  - Purpose: 确保 Service 批量操作逻辑正确
  - Requirements: Requirement 1-4
  - Leverage: 现有测试: `main/app/service/*_test.go`
  - Prompt: Role: QA Engineer | Task: 为批量操作 Service 编写单元测试,覆盖率 ≥ 70% | Context: 测试批量创建、上架、下架、删除方法,测试限流机制,测试重试机制,测试并发处理 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 ≥ 70%,所有测试通过

- [ ] 4.2 编写 API 集成测试

  - File: `docs/shared/specs/active/story-takeout-batch-create-products/API-TEST.md`
  - Purpose: 提供 API 测试用例和测试脚本
  - Requirements: Requirement 1-4
  - Leverage: 参考: `docs/shared/specs/active/story-takeout-grab-products-list/API-TEST.md`
  - Prompt: Role: QA Engineer | Task: 创建 API 测试文档,包含所有批量操作接口的测试用例 | Context: 包含正常场景、异常场景、边界场景测试,提供 curl 测试脚本 | Restrictions: - | Success: 测试文档完整,测试用例覆盖全面,测试脚本可执行

- [ ] 4.3 手动功能测试

  - File: -
  - Purpose: 验证功能完整性和用户体验
  - Requirements: Requirement 1-4
  - Leverage: Task 4.2 的测试用例
  - Test Cases:
    - 批量创建100个商品到Grab平台
    - 批量上架50个商品到LINE MAN平台
    - 批量下架50个商品
    - 批量删除30个商品
    - 验证返回的成功和失败数量
    - 验证限流机制(观察API调用频率)
    - 验证失败重试(模拟平台API失败)
  - Success: 所有测试用例通过,功能符合需求

---

## Phase 5: 文档和发布

- [ ] 5.1 完善 API 文档

  - File: `main/app/api/v1/shop/shop_takeout.go`
  - Purpose: 完善 Swagger 注释
  - Requirements: Requirement 1-4
  - Leverage: Task 3.1 的 Handler 实现
  - Prompt: Role: Technical Writer | Task: 完善批量操作接口的 Swagger 注释 | Context: 包含接口描述、参数说明、响应格式、错误码说明 | Restrictions: - | Success: Swagger 文档完整,描述清晰

---

## 任务执行建议

### 顺序执行

按 Phase 顺序执行,确保依赖关系正确:
1. Phase 1: DTO 层(定义数据结构)
2. Phase 2: Service 层(核心业务逻辑)
3. Phase 3: API 层(对外接口)
4. Phase 4: 测试(确保质量)
5. Phase 5: 文档(交付准备)

### 并行执行建议

部分任务可以并行:
- Task 1.1 和 1.2 可以并行
- Task 2.3, 2.4, 2.5 实现模式相似,可以批量完成
- Task 4.1 和 4.2 可以并行

### 检查点

- Phase 1 完成后,验证 DTO 结构正确
- Phase 2 完成后,进行单元测试
- Phase 3 完成后,进行 API 测试
- Phase 4 完成后,进行功能验收

---

## 风险提示

### 高风险任务

- Task 2.2: 批量创建逻辑复杂,需要仔细处理并发、限流、重试

### 常见问题

1. **Goroutine 内存泄漏**: 确保并发任务正确结束,使用 WaitGroup
2. **限流失效**: 验证 Ticker 正确工作,API 调用频率不超标
3. **重试循环**: 确保重试次数限制生效,避免无限重试
4. **数据一致性**: 操作完成后及时更新数据库状态
5. **并发安全**: 使用 sync.Mutex 保护共享数据

---

## 参考资料

### 核心规范

- `.cursor/rules/go-main.mdc` - Go Main 核心约束
- `.cursor/rules/api.mdc` - API 设计规范
- `.cursor/rules/database.mdc` - 数据库开发规范

### 相关代码

- `main/app/service/product_takeout.go` - 外卖商品服务
- `main/app/modules/takeout/application/takeout_app_service.go` - 外卖平台API
- `main/app/api/v1/shop/shop_takeout.go` - 外卖路由

### 示例参考

- `docs/shared/specs/active/story-takeout-grab-products-list/` - 类似功能的完整实现

---

**版本**: v2.0.0  
**创建日期**: 2025-12-18  
**作者**: weifashi  
**审核者**: 待定
