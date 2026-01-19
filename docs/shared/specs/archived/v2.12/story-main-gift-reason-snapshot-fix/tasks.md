# 赠菜原因快照修复 任务分解

> 本文档定义赠菜原因快照修复功能的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 10  
**已完成**: 2  
**进行中**: -  
**完成率**: 20%

---

## Phase 1: 查询逻辑修改

### 1.1 修改 GetGiftReason() 方法

- [x] 1.1 修改 GetGiftReason() 方法（JSON 方案）

  - File: `main/app/model/sale_order_product.go:1073`
  - Purpose: 实现赠菜原因获取方法，解析 JSON 快照，支持降级兼容
  - Requirements: Requirement 1（查询逻辑修改）
  - Leverage: 参考 `GetCancelReason()` 方法的实现（`main/app/model/sale_order_product.go:988`）
  - Key Logic (JSON 方案):
    1. 优先使用 `SaleOrderProductReason.Name` 快照字段（JSON）
    2. 解析 JSON 为 `dto.LocaleResponse`（包含所有语言）
    3. 快照为空或解析失败时，降级使用 `FreeReason.MultiLanguageName`
    4. 处理自定义赠菜原因（`SaleOrderProduct.GiftReason` 字段）
    5. 返回多语言格式，多个原因用"、"分隔
  - Code Reference:
    ```go
    // 参考 GetCancelReason() 的实现
    func (model *SaleOrderProduct) GetGiftReason() dto.LocaleResponse {
        // 实现逻辑（参考 design.md）
        // 1. 遍历 CancelReasons，筛选 IsGiftReason() 为 true 的原因
        // 2. 优先使用快照字段（JSON）
        // 3. 降级使用关联表数据
        // 4. 处理自定义赠菜原因
    }
    ```
  - Import: 需要添加 `encoding/json` 和 `strings` 导入（如果未导入）
  - Prompt: Role: Go Developer | Task: 修改 GetGiftReason() 方法，优先使用快照字段（JSON），降级使用关联表数据 | Context: 参考 GetCancelReason() 的实现模式，处理 JSON 解析和降级逻辑，注意赠菜原因存储在 CancelReasons 中，需要通过 IsGiftReason() 筛选 | Restrictions: 遵循 .cursor/rules/go-main.mdc，保持现有方法签名 | Success: 方法修改完成，逻辑正确，编译通过
  - Success: 方法修改完成，逻辑正确，编译通过

---

## Phase 2: 下单逻辑修改

### 2.1 修改 CreateSaleOrderProductReasons() 方法

- [x] 2.1 修改 CreateSaleOrderProductReasons() 方法 - 保存赠菜原因快照

  - File: `main/app/repository/sale_order_product.go:209`
  - Purpose: 在创建赠菜原因时，保存快照字段（JSON 格式）
  - Requirements: Requirement 2（下单逻辑修改）
  - Leverage: 参考 `NewFreeOrderReason()` 方法的实现（`main/app/model/sale_order.go:1174`）
  - Key Logic:
    1. 如果是赠菜原因（`source == constant.ProductReasonTypeGift`），加载 `FreeReason` 数据
    2. 从 `FreeReason.MultiLanguageName` 获取完整多语言数据
    3. 序列化为 JSON 字符串
    4. 保存到 `SaleOrderProductReason.Name` 字段
    5. 如果序列化失败，`Name` 字段为空（降级使用关联表）
  - Code Reference:
    ```go
    // 参考 NewFreeOrderReason() 的实现
    func (r *saleOrderProductRepo) CreateSaleOrderProductReasons(...) error {
        // 如果是赠菜原因，加载 FreeReason 数据
        if source == constant.ProductReasonTypeGift {
            // 批量加载 FreeReason 数据
            // 序列化多语言数据为 JSON
            // 保存到 Name 字段
        }
    }
    ```
  - Import: 需要添加 `encoding/json`、`ttpos-server-go/app/repository/base` 导入（如果未导入）
  - Note: 需要批量加载 `FreeReason` 数据，避免 N+1 查询问题
  - Prompt: Role: Go Developer | Task: 修改 CreateSaleOrderProductReasons() 方法，在创建赠菜原因时保存快照字段（JSON 格式） | Context: 参考 NewFreeOrderReason() 的实现模式，批量加载 FreeReason 数据，序列化多语言数据为 JSON 保存，处理序列化失败的情况 | Restrictions: 遵循 .cursor/rules/go-main.mdc，不影响现有退菜和免单原因的创建逻辑 | Success: 方法修改完成，逻辑正确，编译通过
  - Success: 方法修改完成，逻辑正确，编译通过

---

## Phase 3: 测试验证

### 3.1 单元测试

- [ ] 3.1 编写 GetGiftReason() 单元测试

  - File: `main/app/model/sale_order_product_test.go`
  - Purpose: 测试赠菜原因快照方法的正确性
  - Requirements: Requirement 1
  - Leverage: 现有测试: `main/app/model/sale_order_product_test.go`（如有）
  - Test Cases:
    - GetGiftReason() - 快照字段有值且有效 JSON
    - GetGiftReason() - 快照字段为空
    - GetGiftReason() - 快照字段无效 JSON
    - GetGiftReason() - 关联表数据为空
    - GetGiftReason() - 多个赠菜原因组合
    - GetGiftReason() - 自定义赠菜原因
  - Prompt: Role: QA Engineer | Task: 为 GetGiftReason() 编写单元测试，覆盖快照有值/无值、JSON 有效/无效、关联表有数据/无数据、多个原因组合等场景 | Context: 参考 GetCancelReason() 的测试模式 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 ≥ 80%，所有测试通过
  - Success: 测试覆盖率 ≥ 80%，所有测试通过

### 3.2 集成测试

- [ ] 3.2 编写下单集成测试

  - File: `main/app/service/order*_test.go` 或相关测试文件
  - Purpose: 测试下单时保存快照数据
  - Requirements: Requirement 2
  - Leverage: 现有测试: `main/app/service/order*_test.go`
  - Test Cases:
    - 创建订单（包含赠菜原因） → 验证 `SaleOrderProductReason.Name` 字段保存成功（JSON 格式）
    - 删除赠菜原因配置 → 查询订单仍显示快照数据
  - Prompt: Role: QA Engineer | Task: 编写下单集成测试，验证创建订单时赠菜原因快照正确保存为 JSON | Context: 测试所有下单入口（POS、扫码点餐、外卖等） | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 所有下单场景的快照数据保存正确
  - Success: 所有下单场景的快照数据保存正确

- [ ] 3.3 编写查询集成测试

  - File: `main/app/service/order*_test.go` 或相关测试文件
  - Purpose: 测试查询时使用快照数据
  - Requirements: Requirement 1
  - Leverage: 现有测试: `main/app/service/order*_test.go`
  - Test Cases:
    - 创建订单 → 查询订单 → 验证使用快照数据
    - 删除赠菜原因配置 → 查询订单 → 验证仍显示快照数据
    - 修改赠菜原因名称 → 查询订单 → 验证仍显示修改前的名称
    - 历史订单（快照为空） → 查询订单 → 验证降级逻辑正常
  - Prompt: Role: QA Engineer | Task: 编写查询集成测试，验证订单查询时优先使用快照数据，后台删除赠菜原因后，历史订单仍能正常显示 | Context: 测试订单详情、订单列表、订单打印、订单导出等场景 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 所有查询场景的快照逻辑正确
  - Success: 所有查询场景的快照逻辑正确

### 3.3 回归测试

- [ ] 3.4 回归测试 - 订单查询接口

  - File: -
  - Purpose: 确保订单查询功能不受影响
  - Requirements: Requirement 1
  - Leverage: 现有测试用例
  - Prompt: Role: QA Engineer | Task: 执行订单查询接口回归测试，确保所有订单查询接口正常工作 | Context: 测试订单详情、订单列表、订单搜索等接口 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 所有订单查询接口测试通过
  - Success: 所有订单查询接口测试通过

- [ ] 3.5 回归测试 - 订单打印/导出/报表

  - File: -
  - Purpose: 确保订单打印、导出、报表功能不受影响
  - Requirements: Requirement 1
  - Leverage: 现有测试用例
  - Prompt: Role: QA Engineer | Task: 执行订单打印、导出、报表回归测试，确保所有功能正常工作 | Context: 测试订单打印、订单导出、订单报表等功能 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 所有功能测试通过
  - Success: 所有功能测试通过

---

## Phase 4: 代码审查和优化

- [ ] 4.1 代码审查

  - File: 所有修改的文件
  - Purpose: 确保代码质量和规范符合要求
  - Requirements: 所有 Requirement
  - Leverage: 代码审查清单
  - Checklist:
    - [ ] 代码遵循 Go Main 开发规范
    - [ ] 错误处理完善
    - [ ] 日志记录合理
    - [ ] 注释清晰
    - [ ] 性能优化（批量查询）
  - Success: 代码审查通过，无重大问题

- [ ] 4.2 性能优化检查

  - File: `main/app/repository/sale_order_product.go`
  - Purpose: 确保批量查询优化，避免 N+1 问题
  - Requirements: Requirement 2
  - Key Points:
    - 批量加载 `FreeReason` 数据
    - 避免循环中查询数据库
  - Success: 性能优化检查通过，无 N+1 查询问题

---

## Phase 5: 文档更新

- [ ] 5.1 更新代码注释

  - File: `main/app/model/sale_order_product.go`、`main/app/repository/sale_order_product.go`
  - Purpose: 添加或更新方法注释，说明快照逻辑
  - Requirements: 所有 Requirement
  - Success: 代码注释更新完成，清晰说明快照逻辑

---

**版本**: v1.0.0  
**创建日期**: 2025-12-09  
**作者**: xiezhihuan  
**审核者**: {审核者}

