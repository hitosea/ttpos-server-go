# 参考商品单位实现，来源总部的数据不可编辑 任务分解

> 本文档定义参考商品单位（ProductUnit）的实现方式，为多个模块实现总部来源数据不可编辑功能的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码（参考商品单位实现）
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 18  
**已完成**: 10  
**进行中**: -  
**完成率**: 55.6%

---

## Phase 1: 菜品标签（ProductLabel）实现

### 1.1 修改响应结构

- [x] 1.1.1 在 ProductLabelDetail 中添加 IsEditable 字段

  - File: `main/app/dto/resp/product_label.go`
  - Purpose: 添加 `is_editable` 字段到响应结构
  - Requirements: 1.1
  - Leverage: 参考 `main/app/dto/resp/product_resp/product.go` - `ProductUnitItem` (第225行)
  - Prompt: Role: Go Developer | Task: 在 ProductLabelDetail 结构体中添加 `IsEditable bool \`json:"is_editable"\`` 字段 | Context: 参考 ProductUnitItem 的实现方式 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 字段添加成功，JSON 标签正确

### 1.2 修改 Service 方法

- [x] 1.2.1 修改 GetProductLabelList 方法返回 is_editable 字段

  - File: `main/app/service/product_label.go`
  - Purpose: 在列表方法中返回 `is_editable` 字段
  - Requirements: 1.2
  - Leverage: 参考 `main/app/service/product.go` - `GetProductUnitList()` (第2006行)
  - Prompt: Role: Go Developer | Task: 在 GetProductLabelList 方法中，为每个标签添加 `IsEditable: isEditable(ctx, label.HeadquarterUuid)` 字段 | Context: 参考 GetProductUnitList 的实现方式，使用 isEditable 函数判断 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 列表方法返回 is_editable 字段

- [x] 1.2.2 修改 EditProductLabel 方法添加总部来源数据校验

  - File: `main/app/service/product_label.go`
  - Purpose: 在编辑方法中增加总部来源数据校验
  - Requirements: 1.3
  - Leverage: 参考 `main/app/service/product.go` - `EditProductUnit()` (第2260行，第2279行)
  - Prompt: Role: Go Developer | Task: 在 EditProductLabel 方法中，查询标签后添加 `if !isEditable(ctx, label.HeadquarterUuid) { return errors.New("标签不可编辑") }` 校验 | Context: 参考 EditProductUnit 的实现方式 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 编辑方法拒绝总部来源数据

- [x] 1.2.3 修改 DeleteProductLabel 方法添加总部来源数据校验

  - File: `main/app/service/product_label.go`
  - Purpose: 在删除方法中增加总部来源数据校验
  - Requirements: 1.4
  - Leverage: 参考 `main/app/service/product.go` - `DeleteProductUnit()` (第2375行，第2390行)
  - Prompt: Role: Go Developer | Task: 在 DeleteProductLabel 方法中，查询标签后添加 `if !isEditable(ctx, label.HeadquarterUuid) { return errors.New("标签不可删除") }` 校验 | Context: 参考 DeleteProductUnit 的实现方式 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 删除方法拒绝总部来源数据

### 1.3 前端 UI 实现

- [ ] 1.3.1 列表页根据 is_editable 控制编辑按钮

  - File: `admin/views/shop/src/views/product/label/index.vue` (路径待确认)
  - Purpose: 在列表页根据 `is_editable` 字段控制编辑按钮的显示/禁用
  - Requirements: 1.5
  - Leverage: 参考商品单位列表页的前端实现
  - Prompt: Role: Vue Developer | Task: 在列表页中，根据 `is_editable` 字段禁用编辑按钮，显示"来源总部，不可编辑"提示 | Context: 参考商品单位列表页的实现方式 | Restrictions: 遵循 .cursor/rules/vue.mdc | Success: 列表页正确控制编辑按钮

- [ ] 1.3.2 详情页根据 is_editable 禁用表单字段

  - File: `admin/views/shop/src/views/product/label/edit.vue` (路径待确认)
  - Purpose: 在详情页根据 `is_editable` 字段禁用表单字段
  - Requirements: 1.6
  - Leverage: 参考商品单位详情页的前端实现
  - Prompt: Role: Vue Developer | Task: 在详情页中，根据 `is_editable` 字段禁用所有表单字段，显示"来源总部，不可编辑"提示 | Context: 参考商品单位详情页的实现方式 | Restrictions: 遵循 .cursor/rules/vue.mdc | Success: 详情页正确禁用表单字段

---

## Phase 2: 满额减（FullReductionActivity）实现

### 2.1 修改响应结构

- [x] 2.1.1 在 FullReductionActivityResp 中添加 IsEditable 字段

  - File: `main/app/dto/resp/full_reduction_activity_resp.go`
  - Purpose: 添加 `is_editable` 字段到响应结构
  - Requirements: 2.1
  - Leverage: 参考 `main/app/dto/resp/product_resp/product.go` - `ProductUnitItem` (第225行)
  - Prompt: Role: Go Developer | Task: 在 FullReductionActivityResp 结构体中添加 `IsEditable bool \`json:"is_editable"\`` 字段 | Context: 参考 ProductUnitItem 的实现方式 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 字段添加成功，JSON 标签正确

### 2.2 修改 Service 方法

- [x] 2.2.1 修改列表/详情方法返回 is_editable 字段

  - File: `main/app/service/full_reduction_activity_srv.go`
  - Purpose: 在列表/详情方法中返回 `is_editable` 字段
  - Requirements: 2.2
  - Leverage: 参考 `main/app/service/product.go` - `GetProductUnitList()`, `GetProductUnit()` (第2006行，第2044行)
  - Prompt: Role: Go Developer | Task: 在满额减列表/详情方法中，为每个活动添加 `IsEditable: isEditable(ctx, activity.HeadquarterUuid)` 字段 | Context: 参考 GetProductUnitList 和 GetProductUnit 的实现方式 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 列表/详情方法返回 is_editable 字段

- [x] 2.2.2 修改编辑方法添加总部来源数据校验

  - File: `main/app/service/full_reduction_activity_srv.go`
  - Purpose: 在编辑方法中增加总部来源数据校验
  - Requirements: 2.3
  - Leverage: 参考 `main/app/service/product.go` - `EditProductUnit()` (第2260行，第2279行)
  - Prompt: Role: Go Developer | Task: 在满额减编辑方法中，查询活动后添加 `if !isEditable(ctx, activity.HeadquarterUuid) { return errors.New("活动不可编辑") }` 校验 | Context: 参考 EditProductUnit 的实现方式 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 编辑方法拒绝总部来源数据

- [x] 2.2.3 修改删除方法添加总部来源数据校验

  - File: `main/app/service/full_reduction_activity_srv.go`
  - Purpose: 在删除方法中增加总部来源数据校验
  - Requirements: 2.4
  - Leverage: 参考 `main/app/service/product.go` - `DeleteProductUnit()` (第2375行，第2390行)
  - Prompt: Role: Go Developer | Task: 在满额减删除方法中，查询活动后添加 `if !isEditable(ctx, activity.HeadquarterUuid) { return errors.New("活动不可删除") }` 校验 | Context: 参考 DeleteProductUnit 的实现方式 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 删除方法拒绝总部来源数据

### 2.3 前端 UI 实现

- [ ] 2.3.1 列表页根据 is_editable 控制编辑按钮

  - File: `admin/views/shop/src/views/marketing/full_reduction/index.vue` (路径待确认)
  - Purpose: 在列表页根据 `is_editable` 字段控制编辑按钮
  - Requirements: 2.5
  - Leverage: 参考商品单位列表页的前端实现
  - Prompt: Role: Vue Developer | Task: 在列表页中，根据 `is_editable` 字段禁用编辑按钮，显示"来源总部，不可编辑"提示 | Context: 参考商品单位列表页的实现方式 | Restrictions: 遵循 .cursor/rules/vue.mdc | Success: 列表页正确控制编辑按钮

- [ ] 2.3.2 详情页根据 is_editable 禁用表单字段

  - File: `admin/views/shop/src/views/marketing/full_reduction/edit.vue` (路径待确认)
  - Purpose: 在详情页根据 `is_editable` 字段禁用表单字段
  - Requirements: 2.6
  - Leverage: 参考商品单位详情页的前端实现
  - Prompt: Role: Vue Developer | Task: 在详情页中，根据 `is_editable` 字段禁用所有表单字段，显示"来源总部，不可编辑"提示 | Context: 参考商品单位详情页的实现方式 | Restrictions: 遵循 .cursor/rules/vue.mdc | Success: 详情页正确禁用表单字段

---

## Phase 3: 商品（ProductPackage）实现（特殊规则）

### 4.1 确认响应结构

- [x] 4.1.1 确认商品响应结构中已有 IsEditable 字段

  - File: `main/app/dto/resp/product_resp/product.go`
  - Purpose: 确认响应结构中已有 `is_editable` 字段
  - Requirements: 4.1
  - Status: ✅ 已存在
  - Note: 商品响应结构中已有 `IsEditable` 字段，无需修改

### 4.2 修改 Service 方法（特殊规则）

- [x] 4.2.1 修改编辑方法添加总部来源数据校验（特殊规则：允许修改外卖价格、上下架）

  - File: `main/app/service/product.go`
  - Purpose: 在编辑方法中增加总部来源数据校验，但允许修改外卖价格、上下架
  - Requirements: 4.2
  - Leverage: 参考 `main/app/service/product.go` - `EditProductUnit()` (第2260行，第2279行)
  - Prompt: Role: Go Developer | Task: 在商品编辑方法中，查询商品后添加总部来源数据校验，但特殊处理：如果只修改了外卖价格或上下架字段，则允许修改；否则返回错误"商品不可编辑" | Context: 参考 EditProductUnit 的实现方式，但需要特殊规则处理 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 编辑方法正确处理特殊规则

- [x] 4.2.2 修改删除方法添加总部来源数据校验

  - File: `main/app/service/product.go`
  - Purpose: 在删除方法中增加总部来源数据校验
  - Requirements: 4.3
  - Leverage: 参考 `main/app/service/product.go` - `DeleteProductUnit()` (第2375行，第2390行)
  - Prompt: Role: Go Developer | Task: 在商品删除方法中，查询商品后添加 `if !isEditable(ctx, productPackage.HeadquarterUuid) { return errors.New("商品不可删除") }` 校验 | Context: 参考 DeleteProductUnit 的实现方式 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 删除方法拒绝总部来源数据

### 4.3 前端 UI 实现（特殊规则）

- [ ] 4.3.1 列表页根据 is_editable 控制编辑按钮

  - File: `admin/views/shop/src/views/product/index.vue` (路径待确认)
  - Purpose: 在列表页根据 `is_editable` 字段控制编辑按钮
  - Requirements: 4.4
  - Leverage: 参考商品单位列表页的前端实现
  - Prompt: Role: Vue Developer | Task: 在列表页中，根据 `is_editable` 字段禁用编辑按钮，显示"来源总部，不可编辑"提示 | Context: 参考商品单位列表页的实现方式 | Restrictions: 遵循 .cursor/rules/vue.mdc | Success: 列表页正确控制编辑按钮

- [ ] 4.3.2 详情页根据 is_editable 禁用表单字段（外卖价格、上下架除外）

  - File: `admin/views/shop/src/views/product/edit.vue` (路径待确认)
  - Purpose: 在详情页根据 `is_editable` 字段禁用表单字段，但外卖价格、上下架字段除外
  - Requirements: 4.5
  - Leverage: 参考商品单位详情页的前端实现
  - Prompt: Role: Vue Developer | Task: 在详情页中，根据 `is_editable` 字段禁用所有表单字段（外卖价格、上下架字段除外），显示"来源总部，不可编辑"提示 | Context: 参考商品单位详情页的实现方式，但需要特殊处理外卖价格、上下架字段 | Restrictions: 遵循 .cursor/rules/vue.mdc | Success: 详情页正确处理特殊规则

---

## Phase 4: 测试和优化

### 6.1 单元测试

- [ ] 6.1.1 为各模块的 Service 方法编写单元测试

  - File: `main/app/service/*_test.go`
  - Purpose: 确保 Service 方法的正确性
  - Requirements: 测试要求
  - Leverage: 参考现有测试文件
  - Prompt: Role: QA Engineer | Task: 为各模块的 Service 方法编写单元测试，测试 is_editable 字段返回和校验逻辑 | Context: 测试 headquarter_uuid = 0 和 != 0 的情况 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 ≥ 70%，所有测试通过

### 6.2 API 测试

- [ ] 6.2.1 测试各模块的列表/详情接口返回 is_editable 字段

  - File: API 测试文件
  - Purpose: 确保 API 接口正确返回 `is_editable` 字段
  - Requirements: API 测试要求
  - Leverage: 参考现有 API 测试
  - Prompt: Role: QA Engineer | Task: 测试各模块的列表/详情接口，验证返回的 `is_editable` 字段正确 | Context: 测试总部来源数据和分店自建数据的情况 | Restrictions: 遵循 API 测试规范 | Success: 所有接口测试通过

- [ ] 6.2.2 测试各模块的编辑/删除接口拒绝总部来源数据

  - File: API 测试文件
  - Purpose: 确保编辑/删除接口正确拒绝总部来源数据
  - Requirements: API 测试要求
  - Leverage: 参考现有 API 测试
  - Prompt: Role: QA Engineer | Task: 测试各模块的编辑/删除接口，验证拒绝总部来源数据，返回正确的错误提示 | Context: 测试总部来源数据和分店自建数据的情况 | Restrictions: 遵循 API 测试规范 | Success: 所有接口测试通过

### 6.3 集成测试

- [ ] 6.3.1 端到端流程测试

  - File: 集成测试文件
  - Purpose: 确保整个流程正确
  - Requirements: 集成测试要求
  - Leverage: 参考现有集成测试
  - Prompt: Role: QA Engineer | Task: 进行端到端流程测试，验证同步后的数据正确标记为不可编辑 | Context: 测试同步功能与不可编辑功能的集成 | Restrictions: 遵循集成测试规范 | Success: 端到端流程测试通过

### 6.4 UI 交互测试

- [ ] 6.4.1 前端 UI 交互测试

  - File: UI 测试文件
  - Purpose: 确保前端 UI 正确控制编辑功能
  - Requirements: UI 测试要求
  - Leverage: 参考现有 UI 测试
  - Prompt: Role: QA Engineer | Task: 进行前端 UI 交互测试，验证列表页和详情页正确控制编辑按钮和表单字段 | Context: 测试总部来源数据和分店自建数据的情况 | Restrictions: 遵循 UI 测试规范 | Success: UI 交互测试通过

---

## 📝 注意事项

1. **代码一致性**：
   - 所有模块使用相同的 `isEditable()` 函数
   - 所有模块的错误提示格式一致
   - 所有模块的响应结构格式一致

2. **特殊规则**：
   - 商品：只允许修改外卖价格、上下架

3. **前端路径**：
   - 前端文件路径可能需要根据实际项目结构调整
   - 建议先确认前端文件路径再开始实现

4. **测试覆盖**：
   - 确保所有模块的测试覆盖率达标
   - 特别关注特殊规则的测试

---

**版本**: v1.0.0  
**创建日期**: 2025-12-08  
**作者**: 曾振华  
**审核者**: {审核者}
