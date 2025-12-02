# 收银机分批送厨前置关联模式 任务分解

> 本文档定义收银机分批送厨前置关联模式的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 15  
**已完成**: 0  
**进行中**: -  
**完成率**: 0%

---

## Phase 1: 业务设置接口调整

- [ ] 1.1 在 Business 结构中增加 batch_cooking_mode 字段

  - File: `main/app/dto/resp/setting/business_setting.go`
  - Purpose: 在业务设置响应中增加分批送厨模式字段
  - Requirements: 1.1
  - Leverage: 现有 Business 结构，design.md 中的结构定义
  - Prompt: Role: Go Developer | Task: 在 Business 结构体中添加 BatchCookingMode 字段 | Context: 类型为 string，默认值为 "post"，JSON 标签为 "batch_cooking_mode" | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 字段添加成功，类型和标签正确

- [ ] 1.2 在 SettingService 中读取并返回 batch_cooking_mode

  - File: `main/app/service/setting/setting.go`
  - Purpose: 从业务设置中读取分批送厨模式并返回
  - Requirements: 1.2, 1.3
  - Leverage: 现有 GetBusinessSetting 方法，design.md 中的逻辑说明
  - Prompt: Role: Go Developer | Task: 在 GetBusinessSetting 方法中读取 batch_cooking_mode 字段 | Context: 从数据库或配置中读取，默认值为 "post" | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 方法正确读取并返回 batch_cooking_mode

- [ ] 1.3 验证 /cashier/base 接口返回 batch_cooking_mode

  - File: `main/app/api/v1/cashier/cashier_base.go`
  - Purpose: 确保基础信息接口返回分批送厨模式
  - Requirements: 1.1
  - Leverage: 现有 GetBase 方法，CashierBase 响应结构
  - Prompt: Role: Go Developer | Task: 验证 CashierBase 响应中包含 Business 对象，Business 对象中包含 batch_cooking_mode 字段 | Context: 检查响应结构是否正确 | Restrictions: 遵循 .cursor/rules/api.mdc | Success: 接口返回正确的 batch_cooking_mode 字段

---

## Phase 2: 购物车签名计算优化

- [ ] 2.1 创建商品签名计算函数（支持 batch_tag_uuid）

  - File: `main/app/service/order_product.go`
  - Purpose: 创建或修改商品签名计算函数，支持包含 batch_tag_uuid
  - Requirements: 2.1, 2.2
  - Leverage: 现有商品签名计算逻辑，design.md 中的签名计算代码
  - Prompt: Role: Go Developer | Task: 创建 calculateProductSignature 函数，支持前置模式下包含 batch_tag_uuid | Context: 函数接收 product、batch_tag_uuid、isPreMode 参数，返回签名字符串 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 函数创建成功，签名计算逻辑正确

- [ ] 2.2 修改购物车商品合并逻辑（支持基于签名合并）

  - File: `main/app/service/order_product.go`
  - Purpose: 修改购物车商品合并逻辑，支持基于包含 batch_tag_uuid 的签名合并
  - Requirements: 2.2, 2.3
  - Leverage: 现有购物车商品合并逻辑
  - Prompt: Role: Go Developer | Task: 修改购物车商品合并逻辑，使用新的签名计算函数 | Context: 相同签名的商品合并数量，不同签名的商品分开显示 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 合并逻辑正确，相同商品不同分批类型能分开显示

- [ ] 2.3 编写购物车签名计算单元测试

  - File: `main/app/service/order_product_test.go`
  - Purpose: 测试购物车签名计算逻辑
  - Requirements: 2.1, 2.2
  - Leverage: 现有测试文件，design.md 中的测试策略
  - Prompt: Role: QA Engineer | Task: 为购物车签名计算函数编写单元测试，覆盖率 100% | Context: 测试相同商品不同分批类型、相同商品相同分批类型、后置模式等场景 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率达标，所有测试通过

---

## Phase 3: 加购接口调整

- [ ] 3.1 在 InstantOrderAddProductReq 中增加 batch_tag_uuid 字段

  - File: `main/app/dto/req/instant.go`
  - Purpose: 在加购请求结构中增加分批类型UUID字段
  - Requirements: 3.1
  - Leverage: 现有 InstantOrderAddProductReq 结构，design.md 中的结构定义
  - Prompt: Role: Go Developer | Task: 在 InstantOrderAddProductReq 结构体中添加 BatchTagUuid 字段 | Context: 类型为 uint64，JSON 标签为 "batch_tag_uuid"，可选字段 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 字段添加成功，类型和标签正确

- [ ] 3.2 修改 InstantOrderCartProductAdd 方法处理 batch_tag_uuid

  - File: `main/app/service/order_product.go`
  - Purpose: 在加购方法中处理 batch_tag_uuid 参数
  - Requirements: 3.2, 3.3, 3.4
  - Leverage: 现有 InstantOrderCartProductAdd 方法，design.md 中的逻辑流程
  - Prompt: Role: Go Developer | Task: 修改 InstantOrderCartProductAdd 方法，支持前置模式下的 batch_tag_uuid 参数 | Context: 获取业务设置判断模式，验证 batch_tag_uuid，计算签名，合并商品 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 方法正确处理 batch_tag_uuid，逻辑完整

- [ ] 3.3 验证 batch_tag_uuid 的有效性

  - File: `main/app/service/order_product.go`
  - Purpose: 验证传入的 batch_tag_uuid 是否存在于分批类型列表中
  - Requirements: 3.5
  - Leverage: BatchTagRepository，design.md 中的验证逻辑
  - Prompt: Role: Go Developer | Task: 在加购方法中验证 batch_tag_uuid 的有效性 | Context: 调用 BatchTagRepository 查询，如果不存在则返回错误 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 验证逻辑正确，无效的 batch_tag_uuid 返回错误

- [ ] 3.4 编写加购接口单元测试

  - File: `main/app/service/order_product_test.go`
  - Purpose: 测试加购接口的前置模式逻辑
  - Requirements: 3.1-3.5
  - Leverage: 现有测试文件，design.md 中的测试策略
  - Prompt: Role: QA Engineer | Task: 为加购接口编写单元测试，覆盖前置模式场景 | Context: 测试加购商品关联分批类型、未提供 batch_tag_uuid、无效 batch_tag_uuid 等场景 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率达标，所有测试通过

---

## Phase 4: 分批送厨弹窗增强

- [ ] 4.1 修改 GetOrderCartProductBatchCookingList 接口排序逻辑

  - File: `main/app/service/order_cooking.go`
  - Purpose: 修改分批类型列表排序，按 sort 排序，优先级高的在前
  - Requirements: 4.1
  - Leverage: 现有 GetOrderCartProductBatchCookingList 方法，design.md 中的逻辑变更
  - Prompt: Role: Go Developer | Task: 修改 GetOrderCartProductBatchCookingList 方法，分批类型列表按 sort 排序 | Context: 使用 sort 包对分批类型列表进行排序，优先级高的在前 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 排序逻辑正确，分批类型按优先级排序

- [ ] 4.2 在响应中增加类型颜色和缩写信息

  - File: `main/app/dto/resp/shop_cart.go`
  - Purpose: 在分批送厨弹窗响应中增加类型颜色和缩写信息
  - Requirements: 4.2
  - Leverage: 现有 OrderCartProductBatchCookingTag 结构，design.md 中的响应结构
  - Prompt: Role: Go Developer | Task: 在 OrderCartProductBatchCookingTag 结构体中确保包含 Color 和 Abbreviation 字段 | Context: 从 BatchTag 模型中获取颜色和缩写信息 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 响应结构包含完整的类型信息

- [ ] 4.3 编写分批送厨弹窗接口测试

  - File: `main/app/service/order_cooking_test.go`
  - Purpose: 测试分批送厨弹窗接口
  - Requirements: 4.1, 4.2
  - Leverage: 现有测试文件
  - Prompt: Role: QA Engineer | Task: 为分批送厨弹窗接口编写测试 | Context: 测试分批类型排序、类型信息返回等场景 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试通过，排序和返回信息正确

---

## Phase 5: 更换类型接口开发

- [ ] 5.1 创建 ChangeBatchTagReq 请求结构

  - File: `main/app/dto/req/instant.go`
  - Purpose: 创建更换类型请求结构
  - Requirements: 5.2
  - Leverage: 现有请求结构，design.md 中的结构定义
  - Prompt: Role: Go Developer | Task: 创建 ChangeBatchTagReq 结构体 | Context: 包含 SaleBillUuid、SaleOrderProductUuids、BatchTagUuid 字段，使用 binding 标签验证 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 结构体创建成功，验证标签正确

- [ ] 5.2 创建 ChangeBatchTag Service 方法

  - File: `main/app/service/order_product.go`
  - Purpose: 实现更换类型业务逻辑
  - Requirements: 5.3, 5.4, 5.5
  - Leverage: 现有 Service 方法，design.md 中的逻辑流程
  - Prompt: Role: Go Developer | Task: 创建 ChangeBatchTag 方法，实现更换类型逻辑 | Context: 验证商品是否已送厨，验证 batch_tag_uuid，更新商品关联，同步到点餐助手端 | Restrictions: 遵循 .cursor/rules/go-main.mdc，使用事务 | Success: 方法实现完整，逻辑正确

- [ ] 5.3 创建 ChangeBatchTag API Handler

  - File: `main/app/api/v1/cashier/cashier_desk.go`
  - Purpose: 创建更换类型API接口
  - Requirements: 5.1
  - Leverage: 现有 API Handler，design.md 中的 API 设计
  - Prompt: Role: Go Developer | Task: 创建 ChangeBatchTag Handler 方法 | Context: 路由为 POST /cashier/desk/order/cart/batch/change_tag，调用 Service 方法 | Restrictions: 遵循 .cursor/rules/api.mdc | Success: Handler 创建成功，路由注册正确

- [ ] 5.4 编写更换类型接口单元测试

  - File: `main/app/service/order_product_test.go`
  - Purpose: 测试更换类型接口
  - Requirements: 5.1-5.6
  - Leverage: 现有测试文件，design.md 中的测试策略
  - Prompt: Role: QA Engineer | Task: 为更换类型接口编写单元测试 | Context: 测试未送厨商品可以更换、已送厨商品不允许更换、无效 batch_tag_uuid 等场景 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率达标，所有测试通过

---

## Phase 6: 分批类型列表接口

- [ ] 6.1 创建或修改分批类型列表接口

  - File: `main/app/api/v1/cashier/cashier_desk.go` 或新建文件
  - Purpose: 创建分批类型列表接口，支持按 sort 排序
  - Requirements: 6.1, 6.2
  - Leverage: 现有 BatchTagRepository，design.md 中的 API 设计
  - Prompt: Role: Go Developer | Task: 创建分批类型列表接口，返回按 sort 排序的列表 | Context: 路由为 GET /cashier/desk/batch_tag/list，返回包含 uuid、locale_name、color、sort、abbreviation 的列表 | Restrictions: 遵循 .cursor/rules/api.mdc | Success: 接口创建成功，排序和返回信息正确

- [ ] 6.2 编写分批类型列表接口测试

  - File: `main/app/api/v1/cashier/cashier_desk_test.go` 或新建测试文件
  - Purpose: 测试分批类型列表接口
  - Requirements: 6.1, 6.2
  - Leverage: 现有测试文件
  - Prompt: Role: QA Engineer | Task: 为分批类型列表接口编写测试 | Context: 测试列表排序、返回信息完整性等场景 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试通过，排序和返回信息正确

---

## Phase 7: 集成测试和文档

- [ ] 7.1 编写端到端集成测试

  - File: `main/tests/integration/` 或新建测试文件
  - Purpose: 测试完整的加购和送厨流程
  - Requirements: 所有功能需求
  - Leverage: 现有集成测试，design.md 中的测试策略
  - Prompt: Role: QA Engineer | Task: 编写端到端集成测试，覆盖完整流程 | Context: 测试前置模式下加购商品、更换类型、分批送厨等完整流程 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 集成测试通过，流程正确

- [ ] 7.2 更新 API 文档

  - File: `docs/shared/api/` 或相关文档
  - Purpose: 更新 API 文档，包含新增和修改的接口
  - Requirements: 文档验收
  - Leverage: 现有 API 文档
  - Prompt: Role: Technical Writer | Task: 更新 API 文档，包含新增和修改的接口说明 | Context: 添加更换类型接口、修改加购接口、修改基础信息接口的文档 | Restrictions: 遵循 .cursor/rules/documentation.mdc | Success: API 文档更新完整，接口说明清晰

- [ ] 7.3 代码审查和优化

  - File: 所有修改的文件
  - Purpose: 代码审查，优化性能和可维护性
  - Requirements: 所有功能需求
  - Leverage: 代码审查清单
  - Prompt: Role: Senior Developer | Task: 进行代码审查，优化代码质量和性能 | Context: 检查代码规范、性能优化、错误处理、日志记录等 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 代码审查通过，优化完成

---

## 📝 任务依赖关系

```
Phase 1 (业务设置) 
  ↓
Phase 2 (签名计算)
  ↓
Phase 3 (加购接口)
  ↓
Phase 4 (弹窗增强)
  ↓
Phase 5 (更换类型)
  ↓
Phase 6 (类型列表)
  ↓
Phase 7 (集成测试)
```

---

## 🎯 关键里程碑

- **Milestone 1**: Phase 1-2 完成（业务设置和签名计算）
- **Milestone 2**: Phase 3-4 完成（加购接口和弹窗增强）
- **Milestone 3**: Phase 5-6 完成（更换类型和类型列表）
- **Milestone 4**: Phase 7 完成（集成测试和文档）

---

**版本**: v1.0.0  
**创建日期**: 2025-11-20  
**作者**: 后端开发组  
**审核者**: 待定

