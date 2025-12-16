# Shop 商家管理端-套餐管理移除ERP集成 任务分解

> 本文档定义 Shop 商家管理端套餐管理移除ERP集成的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求
- **关键约束**: 必须确保普通商品的ERP同步功能不受影响

## 📊 进度总览

**总任务数**: 11  
**已完成**: 9  
**进行中**: -  
**完成率**: 81.8%

---

## Phase 1: 代码分析和准备

- [x] 1.1 分析套餐添加逻辑中的ERP调用

  - File: `main/app/service/product.go` - `AddProductFlavor` 方法
  - Purpose: 识别套餐添加时调用ERP的具体位置和逻辑
  - Requirements: 1.1, 1.2
  - Leverage: 现有代码: `main/app/service/product.go` (约 6360-6402 行)
  - Prompt: Role: Go Developer | Task: 分析 AddProductFlavor 方法中套餐添加时的ERP调用逻辑 | Context: 查找 `productPackage.IsPackage()` 判断后的 `erpSrv.AddPackage()` 调用 | Restrictions: 注意区分套餐和普通商品的处理逻辑 | Success: 明确识别出套餐添加时的ERP调用位置

- [x] 1.2 分析套餐修改逻辑中的ERP调用

  - File: `main/app/service/product.go` - `UpdateProductUnit` 方法
  - Purpose: 识别套餐单位更新时调用ERP的具体位置和逻辑
  - Requirements: 2.1, 2.2
  - Leverage: 现有代码: `main/app/service/product.go` (约 2125-2148 行)
  - Prompt: Role: Go Developer | Task: 分析 UpdateProductUnit 方法中套餐单位更新时的ERP调用逻辑 | Context: 查找 `productPackage.IsPackage()` 判断后的 `erpSrv.UpdateProduct()` 调用 | Restrictions: 注意区分套餐和普通商品的处理逻辑 | Success: 明确识别出套餐单位更新时的ERP调用位置

- [x] 1.3 分析套餐删除逻辑中的ERP调用

  - File: `main/app/service/product.go` - `DeleteProductShop` 方法
  - Purpose: 识别套餐删除时调用ERP的具体位置和逻辑
  - Requirements: 3.1, 3.2
  - Leverage: 现有代码: `main/app/service/product.go` (约 1414-1432 行)
  - Prompt: Role: Go Developer | Task: 分析 DeleteProductShop 方法中套餐删除时的ERP调用逻辑 | Context: 查找 `product.IsPackage()` 判断后的 `erpSrv.DeleteProduct()` 调用 | Restrictions: 注意区分套餐和普通商品的处理逻辑（普通商品删除时调用 UpdateProduct 设置禁售） | Success: 明确识别出套餐删除时的ERP调用位置

- [x] 1.4 分析套餐同步逻辑中的ERP调用

  - File: `main/app/service/sync_product_to_erp.go` - `SyncProductToErp` 方法
  - Purpose: 识别套餐同步时调用ERP的具体位置和逻辑
  - Requirements: 4.1, 4.2
  - Leverage: 现有代码: `main/app/service/sync_product_to_erp.go` (约 154-195 行)
  - Prompt: Role: Go Developer | Task: 分析 SyncProductToErp 方法中套餐同步时的ERP调用逻辑 | Context: 查找 `productPackage.IsPackage()` 判断后的 `erpSrv.AddPackage()` 调用 | Restrictions: 注意区分套餐和普通商品的处理逻辑 | Success: 明确识别出套餐同步时的ERP调用位置

---

## Phase 2: 移除套餐ERP调用

- [x] 2.1 移除添加套餐的ERP调用

  - File: `main/app/service/product.go` - `AddProductFlavor` 方法
  - Purpose: 移除套餐添加时对 `erpSrv.AddPackage()` 的调用，保留本地数据库操作
  - Requirements: 1.1, 1.2, 1.3
  - Leverage: Task 1.1 的分析结果，现有代码: `main/app/service/product.go`
  - Prompt: Role: Go Developer specializing in Service Layer | Task: 移除 AddProductFlavor 方法中套餐添加时的ERP调用 | Context: 在 `productPackage.IsPackage()` 判断后，移除 `erpSrv.AddPackage()` 调用和相关错误处理，保留本地数据库操作 | Restrictions: 必须确保普通商品的ERP调用不受影响，使用 `productPackage.IsProduct()` 判断普通商品 | Success: 套餐添加不再调用ERP，普通商品ERP调用保持不变，代码通过编译和测试

- [x] 2.2 移除修改套餐的ERP调用

  - File: `main/app/service/product.go` - `UpdateProductUnit` 方法
  - Purpose: 移除套餐单位更新时对 `erpSrv.UpdateProduct()` 的调用，保留本地数据库操作
  - Requirements: 2.1, 2.2, 2.3
  - Leverage: Task 1.2 的分析结果，现有代码: `main/app/service/product.go`
  - Prompt: Role: Go Developer specializing in Service Layer | Task: 移除 UpdateProductUnit 方法中套餐单位更新时的ERP调用 | Context: 在 `productPackage.IsPackage()` 判断后，移除 `erpSrv.UpdateProduct()` 调用和相关错误处理，保留本地数据库更新操作 | Restrictions: 必须确保普通商品的ERP调用不受影响，使用 `productPackage.IsProduct()` 判断普通商品 | Success: 套餐单位更新不再调用ERP，普通商品ERP调用保持不变，代码通过编译和测试

- [x] 2.3 移除删除套餐的ERP调用

  - File: `main/app/service/product.go` - `DeleteProductShop` 方法
  - Purpose: 移除套餐删除时对 `erpSrv.DeleteProduct()` 的调用，保留本地数据库软删除操作
  - Requirements: 3.1, 3.2, 3.3, 3.4
  - Leverage: Task 1.3 的分析结果，现有代码: `main/app/service/product.go`
  - Prompt: Role: Go Developer specializing in Service Layer | Task: 移除 DeleteProductShop 方法中套餐删除时的ERP调用 | Context: 在 `product.IsPackage()` 判断后，移除 `erpSrv.DeleteProduct()` 调用和相关错误处理，保留本地数据库软删除操作（已在事务中执行） | Restrictions: 必须确保普通商品删除时的ERP调用不受影响，普通商品删除时调用 UpdateProduct 设置禁售 | Success: 套餐删除不再调用ERP，普通商品删除ERP调用保持不变，代码通过编译和测试

- [x] 2.4 移除同步套餐的ERP调用

  - File: `main/app/service/sync_product_to_erp.go` - `SyncProductToErp` 方法
  - Purpose: 移除套餐同步时对 `erpSrv.AddPackage()` 的调用，跳过套餐的ERP同步
  - Requirements: 4.1, 4.2, 4.3
  - Leverage: Task 1.4 的分析结果，现有代码: `main/app/service/sync_product_to_erp.go`
  - Prompt: Role: Go Developer specializing in Service Layer | Task: 移除 SyncProductToErp 方法中套餐同步时的ERP调用 | Context: 在 `productPackage.IsPackage()` 判断后，使用 `continue` 跳过套餐的ERP同步，移除 `erpSrv.AddPackage()` 调用 | Restrictions: 必须确保普通商品的ERP同步不受影响，使用 `productPackage.IsProduct()` 判断普通商品 | Success: 套餐同步跳过ERP调用，普通商品ERP同步保持不变，代码通过编译和测试

- [x] 2.5 清理相关的错误处理和日志记录

  - File: `main/app/service/product.go`, `main/app/service/sync_product_to_erp.go`
  - Purpose: 清理不再使用的ERP相关错误处理和日志记录代码
  - Requirements: 5.1, 5.2
  - Leverage: Task 2.1-2.4 的修改结果
  - Prompt: Role: Go Developer | Task: 清理套餐相关ERP调用的错误处理和日志记录代码 | Context: 移除 `logger.Logger.Error("同步套餐到ERP失败", ...)` 等日志记录，移除相关的错误处理逻辑 | Restrictions: 只清理套餐相关的代码，保留普通商品的错误处理和日志记录 | Success: 代码清理完成，无遗留的套餐ERP相关代码

---

## Phase 3: 测试和验证

- [ ] 3.1 编写单元测试 - 套餐操作不调用ERP

  - File: `main/app/service/product_test.go`, `main/app/service/sync_product_to_erp_test.go`
  - Purpose: 确保套餐添加、修改、删除、同步操作不再调用ERP接口
  - Requirements: 1.1, 2.1, 3.1, 4.1
  - Leverage: 现有测试: `main/app/service/*_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为套餐操作编写单元测试，验证不调用ERP接口 | Context: 使用 mock 或 spy 验证 `erpSrv.AddPackage()`、`erpSrv.UpdateProduct()` 和 `erpSrv.DeleteProduct()` 不被调用 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 ≥ 70%，所有测试通过，验证套餐操作不调用ERP

- [ ] 3.2 编写单元测试 - 普通商品ERP同步不受影响

  - File: `main/app/service/product_test.go`, `main/app/service/sync_product_to_erp_test.go`
  - Purpose: 确保普通商品的ERP同步功能正常工作，不受套餐ERP移除影响
  - Requirements: 3.5, 4.4
  - Leverage: 现有测试: `main/app/service/*_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为普通商品ERP同步编写单元测试，验证功能不受影响 | Context: 使用 mock 或 spy 验证 `erpSrv.AddProductBom()`、`erpSrv.UpdateProduct()` 正常调用，删除普通商品时验证调用 UpdateProduct 设置禁售 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 ≥ 70%，所有测试通过，验证普通商品ERP同步正常

- [ ] 3.3 集成测试 - 端到端流程测试

  - File: `test/integration/package_erp_test.go`
  - Purpose: 测试端到端流程，验证套餐操作和普通商品操作的正确性
  - Requirements: 所有功能需求
  - Leverage: 现有集成测试
  - Prompt: Role: QA Automation Engineer | Task: 实现端到端集成测试 | Context: 测试添加套餐、修改套餐、删除套餐、同步商品（套餐跳过ERP，普通商品正常同步）的完整流程 | Restrictions: 测试真实用户场景 | Success: 集成测试通过，验证套餐不调用ERP，普通商品ERP同步正常

- [ ] 3.4 回归测试 - 验证普通商品功能正常

  - File: `test/regression/product_erp_test.go`
  - Purpose: 回归测试普通商品的ERP同步功能，确保不受影响
  - Requirements: 3.5, 4.4
  - Leverage: 现有测试用例
  - Prompt: Role: QA Engineer | Task: 执行回归测试，验证普通商品ERP同步功能正常 | Context: 测试普通商品的添加、修改、删除、同步操作，验证ERP调用正常（删除时调用 UpdateProduct 设置禁售） | Restrictions: 确保所有普通商品相关测试通过 | Success: 回归测试通过，普通商品功能正常

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] 测试覆盖率达标
  - Service: ≥ 70%
  - Repository: ≥ 80%
- [ ] 所有测试通过

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成
- [ ] **关键验证**: 套餐操作不调用ERP，普通商品ERP同步不受影响

### 文档同步

- [ ] CHANGELOG.md 已更新
- [ ] 代码注释已更新，说明已移除套餐ERP集成

### 规范遵循

- [ ] 遵循 `.cursor/rules/go-main.mdc`
- [ ] 遵循 `.cursor/rules/api.mdc`
- [ ] 遵循 `.cursor/rules/database.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-shop-package-remove-erp-integration/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-shop-package-remove-erp-integration/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-shop-package-remove-erp-integration/tasks.md
```

### 执行流程

1. **选择任务**: 选择下一个未完成任务
2. **阅读需求**: 查看 requirements.md 中的关联需求
3. **查看复用**: 检查 Leverage 中的可复用代码
4. **使用 AI**: 复制 Prompt 模板，让 AI 生成代码
5. **实现代码**: 按照规范实现功能，**特别注意区分套餐和普通商品**
6. **运行检查**: `go fmt`, `go vet`, `go test`
7. **标记完成**: 将 `[ ]` 改为 `[x]`
8. **提交代码**: Git commit（参考 `.cursor/rules/version.mdc`）

---

## 关键注意事项

### ⚠️ 必须确保普通商品不受影响

1. **使用商品类型判断**: 必须使用 `productPackage.IsPackage()` 和 `productPackage.IsProduct()` 明确区分套餐和普通商品
2. **保留普通商品逻辑**: 所有普通商品的ERP调用逻辑必须保持不变
   - 普通商品删除时调用 `erpSrv.UpdateProduct()` 设置 `NotForSale: true`（禁售）
   - 套餐删除时不再调用ERP接口
3. **测试验证**: 必须编写测试验证普通商品的ERP同步功能不受影响

### 🔍 代码审查重点

1. **套餐判断**: 确认所有 `productPackage.IsPackage()` 或 `product.IsPackage()` 判断后的ERP调用已移除
   - 添加套餐：移除 `erpSrv.AddPackage()` 调用
   - 修改套餐：移除 `erpSrv.UpdateProduct()` 调用
   - 删除套餐：移除 `erpSrv.DeleteProduct()` 调用
   - 同步套餐：跳过ERP同步逻辑
2. **普通商品判断**: 确认所有 `productPackage.IsProduct()` 或 `!product.IsPackage()` 判断后的ERP调用保持不变
   - 普通商品删除时调用 `erpSrv.UpdateProduct()` 设置禁售
3. **错误处理**: 确认套餐相关的错误处理已清理，普通商品的错误处理保持不变

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 在执行任务过程中若总结出经验或规避策略，请记录 Episode，并在 tasks.md 尾部更新 `Related Episode`。

---

**模板版本**: v1.0.0  
**最后更新**: 2025-12-01  
**维护者**: 后端开发组

