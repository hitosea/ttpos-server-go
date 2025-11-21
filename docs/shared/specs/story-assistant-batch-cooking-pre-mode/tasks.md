# 点餐助手端分批送厨前置模式 任务分解

> 本文档定义点餐助手端分批送厨前置模式的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 12  
**已完成**: 10  
**进行中**: -  
**完成率**: 83.3%

---

## Phase 1: 业务设置接口调整

- [x] 1.1 在 Business 结构中增加 batch_cooking_mode 字段

  - File: `main/app/dto/resp/setting/business_setting.go`
  - Purpose: 在业务设置响应中增加分批送厨模式字段
  - Requirements: 1.1
  - Leverage: 现有 Business 结构，design.md 中的结构定义
  - Prompt: Role: Go Developer | Task: 在 Business 结构体中添加 BatchCookingMode 字段 | Context: 类型为 string，默认值为 "post"，JSON 标签为 "batch_cooking_mode" | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 字段添加成功，类型和标签正确

- [x] 1.2 在 SettingService 中读取并返回 batch_cooking_mode

  - File: `main/app/service/setting/setting.go`
  - Purpose: 从业务设置中读取分批送厨模式并返回
  - Requirements: 1.2, 1.3
  - Leverage: 现有 GetBusinessSetting 方法，design.md 中的逻辑说明
  - Prompt: Role: Go Developer | Task: 在 GetBusinessSetting 方法中读取 batch_cooking_mode 字段 | Context: 从数据库或配置中读取，默认值为 "post" | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 方法正确读取并返回 batch_cooking_mode

- [x] 1.3 验证 /assistant/base 接口返回 batch_cooking_mode

  - File: `main/app/api/v1/assistant/assistant_base.go`
  - Purpose: 确保基础信息接口返回分批送厨模式
  - Requirements: 1.1
  - Leverage: 现有 GetBase 方法，AssistantBase 响应结构
  - Prompt: Role: Go Developer | Task: 验证 AssistantBase 响应中包含 Business 对象，Business 对象中包含 batch_cooking_mode 字段 | Context: 检查响应结构是否正确 | Restrictions: 遵循 .cursor/rules/api.mdc | Success: 接口返回正确的 batch_cooking_mode 字段

---

## Phase 2: 分批类型列表接口

- [x] 2.1 创建 BatchTagListResp 响应结构

  - File: `main/app/dto/resp/shop_cart.go` 或新建文件
  - Purpose: 创建分批类型列表响应结构
  - Requirements: 2.2
  - Leverage: 现有响应结构，design.md 中的结构定义
  - Prompt: Role: Go Developer | Task: 创建 BatchTagListResp 和 BatchTagItem 结构体 | Context: 包含 uuid、locale_name、color、sort、abbreviation 字段 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 结构体创建成功，字段类型正确

- [x] 2.2 创建 GetBatchTagList 方法

  - File: `main/app/api/v1/assistant/assistant_base.go`
  - Purpose: 创建分批类型列表接口处理方法
  - Requirements: 2.1, 2.3
  - Leverage: 现有 BaseHandler，ProductService.GetBatchTagList 方法（与收银端保持一致）
  - Prompt: Role: Go Developer | Task: 创建 GetBatchTagList 方法，通过 ProductService 调用获取分批类型列表 | Context: 在 BaseHandler 中注入 productSrv，调用 productSrv.GetBatchTagList(ctx, req.BatchTagListReq{}) | Restrictions: 遵循 .cursor/rules/go-main.mdc，遵循分层架构（API → Service → Repository） | Success: 方法创建成功，通过 Service 层调用，与收银端实现一致

- [x] 2.3 注册分批类型列表路由

  - File: `main/app/api/v1/assistant/assistant_base.go`
  - Purpose: 注册 `/assistant/batch_tag/list` 路由
  - Requirements: 2.1
  - Leverage: 现有 RegisterBaseHandlers 方法，参考收银端实现
  - Prompt: Role: Go Developer | Task: 在 RegisterBaseHandlers 中注册 GET /assistant/batch_tag/list 路由，初始化 ProductService 并注入到 BaseHandler | Context: 需要认证，调用 GetBatchTagList 方法，初始化 localeSrv、translateSrv、productSrv | Restrictions: 遵循 .cursor/rules/api.mdc | Success: 路由注册成功，ProductService 初始化正确，接口可访问

---

## Phase 3: 购物车签名计算优化

- [x] 3.1 创建商品签名计算函数（支持 batch_tag_uuid）

  - File: `main/app/service/order_product.go`
  - Purpose: 创建或修改商品签名计算函数，支持包含 batch_tag_uuid
  - Requirements: 3.1, 3.2
  - Leverage: 现有商品签名计算逻辑，design.md 中的签名计算代码
  - Prompt: Role: Go Developer | Task: 创建 calculateProductSignature 函数，支持前置模式下包含 batch_tag_uuid | Context: 函数接收 product、batch_tag_uuid、isPreMode 参数，返回签名字符串 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 函数创建成功，签名计算逻辑正确

- [x] 3.2 修改购物车商品合并逻辑（支持基于签名合并）

  - File: `main/app/service/order_product.go`
  - Purpose: 修改购物车商品合并逻辑，支持基于包含 batch_tag_uuid 的签名合并
  - Requirements: 3.2, 3.3
  - Leverage: 现有购物车商品合并逻辑
  - Prompt: Role: Go Developer | Task: 修改购物车商品合并逻辑，使用新的签名计算函数 | Context: 相同签名的商品合并数量，不同签名的商品分开显示 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 合并逻辑正确，相同商品不同分批类型能分开显示

- [ ] 3.3 编写购物车签名计算单元测试

  - File: `main/app/service/order_product_test.go`
  - Purpose: 测试购物车签名计算逻辑
  - Requirements: 3.1, 3.2
  - Leverage: 现有测试文件，design.md 中的测试策略
  - Prompt: Role: QA Engineer | Task: 为购物车签名计算函数编写单元测试，覆盖率 100% | Context: 测试相同商品不同分批类型、相同商品相同分批类型、后置模式等场景 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率达标，所有测试通过

---

## Phase 4: 加购接口调整

- [x] 4.1 在 InstantOrderAddProductReq 中增加 batch_tag_uuid 字段

  - File: `main/app/dto/req/instant.go`
  - Purpose: 在加购请求结构中增加分批类型UUID字段
  - Requirements: 4.1
  - Leverage: 现有 InstantOrderAddProductReq 结构，design.md 中的结构定义
  - Prompt: Role: Go Developer | Task: 在 InstantOrderAddProductReq 结构体中添加 BatchTagUuid 字段 | Context: 类型为 uint64，JSON 标签为 "batch_tag_uuid"，可选字段 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 字段添加成功，类型和标签正确

- [x] 4.2 修改 InstantOrderCartProductAdd 方法处理 batch_tag_uuid

  - File: `main/app/service/order_product.go`
  - Purpose: 在加购方法中处理 batch_tag_uuid 参数
  - Requirements: 4.2, 4.3, 4.4, 4.6
  - Leverage: 现有 InstantOrderCartProductAdd 方法，design.md 中的逻辑流程
  - Prompt: Role: Go Developer | Task: 修改 InstantOrderCartProductAdd 方法，支持前置模式下的 batch_tag_uuid 参数 | Context: 获取业务设置判断模式，验证 batch_tag_uuid，计算签名，合并商品，如果未提供则使用默认类型 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 方法正确处理 batch_tag_uuid，逻辑完整

- [x] 4.3 验证 batch_tag_uuid 的有效性

  - File: `main/app/service/order_product.go`
  - Purpose: 验证传入的 batch_tag_uuid 是否存在于分批类型列表中
  - Requirements: 4.5
  - Leverage: BatchTagRepository，design.md 中的验证逻辑
  - Prompt: Role: Go Developer | Task: 在加购方法中验证 batch_tag_uuid 的有效性 | Context: 调用 BatchTagRepository 查询，如果不存在则返回错误 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 验证逻辑正确，无效的 batch_tag_uuid 返回错误

- [ ] 4.4 编写加购接口单元测试

  - File: `main/app/service/order_product_test.go`
  - Purpose: 测试加购接口的前置模式逻辑
  - Requirements: 4.1-4.6
  - Leverage: 现有测试文件，design.md 中的测试策略
  - Prompt: Role: QA Engineer | Task: 为加购接口编写单元测试，覆盖前置模式场景 | Context: 测试加购商品关联分批类型、未提供 batch_tag_uuid、无效 batch_tag_uuid 等场景 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率达标，所有测试通过

---

## Phase 5: 智能送厨逻辑

- [x] 5.1 创建自动送厨方法

  - File: `main/app/service/order_action.go`
  - Purpose: 创建自动按优先级送厨的方法
  - Requirements: 5.2, 5.3
  - Leverage: 现有送厨逻辑，design.md 中的智能送厨逻辑
  - Prompt: Role: Go Developer | Task: 创建 autoSendCookingByPriority 方法，实现自动按优先级送厨 | Context: 获取预送厨商品，按分批类型分组，找出优先级最高的类型，送厨该类型商品 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 方法创建成功，逻辑正确

- [x] 5.2 修改下单逻辑调用自动送厨

  - File: `main/app/service/order_action.go`
  - Purpose: 在下单成功后调用自动送厨逻辑
  - Requirements: 5.1, 5.4
  - Leverage: 现有 ActionSubmit 方法
  - Prompt: Role: Go Developer | Task: 修改 ActionSubmit 方法，下单成功后调用自动送厨逻辑 | Context: 判断是否为前置模式，如果是则调用 autoSendCookingByPriority | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 下单逻辑正确调用自动送厨

- [ ] 5.3 编写智能送厨单元测试

  - File: `main/app/service/order_action_test.go`
  - Purpose: 测试智能送厨逻辑
  - Requirements: 5.1-5.4
  - Leverage: 现有测试文件，design.md 中的测试策略
  - Prompt: Role: QA Engineer | Task: 为智能送厨逻辑编写单元测试 | Context: 测试按优先级正确送厨、每次点击下单都送优先级最高的类型、没有可送厨商品时正常完成下单 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率达标，所有测试通过

---

## Phase 6: 集成测试和文档

- [ ] 6.1 编写端到端集成测试

  - File: `main/app/service/order_integration_test.go` 或新建文件
  - Purpose: 测试完整的加购和送厨流程
  - Requirements: 所有功能需求
  - Leverage: 现有测试框架
  - Prompt: Role: QA Engineer | Task: 编写端到端集成测试，覆盖完整加购和送厨流程 | Context: 测试前置模式下加购多个商品、下单后自动按优先级送厨等场景 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 集成测试通过，覆盖主要场景

- [ ] 6.2 更新 API 文档

  - File: `docs/shared/api/assistant.md` 或新建文件
  - Purpose: 更新助手端 API 文档
  - Requirements: 所有 API 需求
  - Leverage: 现有 API 文档格式
  - Prompt: Role: Technical Writer | Task: 更新助手端 API 文档，添加新增和修改的接口说明 | Context: 包括 /assistant/base、/assistant/batch_tag/list、/assistant/desk/order/cart/product/add 等接口 | Restrictions: 遵循 .cursor/rules/documentation.mdc | Success: API 文档更新完成，接口说明清晰

---

**版本**: v1.0.0  
**创建日期**: 2025-11-20  
**作者**: 后端开发组  
**审核者**: 待定

