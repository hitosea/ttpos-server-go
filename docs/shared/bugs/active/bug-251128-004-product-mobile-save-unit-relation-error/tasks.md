# Bug-251128-004 修复任务清单

> **当前状态**: 🟡 规划中  
> **开始时间**: 2025-11-28  
> **预计完成**: 待定

---

## 📋 任务列表

### Phase 1: 前置条件 - Main 模块 RPC 层扩展

- [x] **1.1 扩展 UpdateProductReq 结构体** `main/app/service/rpc/erp/product.go`
  - **需求**: 在 `UpdateProductReq` 结构体中添加 `StockUom` 字段
  - **预计时间**: 0.5 小时
  - **负责人**: 
  - **代码位置**: 第 292-298 行
  - **变更内容**:
    ```go
    type UpdateProductReq struct {
        ItemCode     string                `json:"item_code"`
        NotForSale   bool                  `json:"not_for_sale"`
        InternalCode string                `json:"internal_code"`
        Disabled     bool                  `json:"disabled"`
        Attributes   []UpdateProductFlavor `json:"attributes"`
        StockUom     string                `json:"stock_uom"` // ✅ 新增字段
    }
    ```

- [x] **1.2 更新 UpdateProduct 方法** `main/app/service/rpc/erp/product.go`
  - **需求**: 在 `UpdateProduct` 方法中传递 `StockUom` 字段到 gRPC 调用
  - **预计时间**: 0.5 小时
  - **负责人**: 
  - **代码位置**: 第 306-339 行
  - **变更内容**: 在第 324-330 行的 gRPC 调用中添加 `StockUom: params.StockUom`

### Phase 2: 核心修复 - 单位关联逻辑调整

- [x] **2.1 修复 EditProductUnit 方法 - 普通商品** `main/app/service/product.go`
  - **需求**: 将普通商品的 `AddProduct` 调用改为 `UpdateProduct`，更新模板单位
  - **预计时间**: 1 小时
  - **负责人**: 
  - **代码位置**: 第 2274-2343 行
  - **变更内容**:
    - 替换第 94-115 行的普通商品处理逻辑
    - 使用 `productPackage.ErpCode`（模板 Item Code）调用 `UpdateProduct`
    - 添加详细的错误日志记录

- [x] **2.2 修复 EditProductUnit 方法 - 套餐** `main/app/service/product.go`
  - **需求**: 将套餐的 `AddPackage` 调用改为 `UpdateProduct`，更新套餐单位
  - **预计时间**: 1 小时
  - **负责人**: 
  - **代码位置**: 第 2274-2343 行
  - **变更内容**:
    - 替换第 78-93 行的套餐处理逻辑
    - 使用 `productBom.ErpCode`（套餐 Item Code）调用 `UpdateProduct`
    - 添加详细的错误日志记录

- [x] **2.3 修复 AddProductUnit 方法** `main/app/service/product.go`
  - **需求**: 将新增单位时关联商品的 `AddProduct` 调用改为 `UpdateProduct`
  - **预计时间**: 1 小时
  - **负责人**: 
  - **代码位置**: 第 2108-2147 行
  - **变更内容**:
    - 替换第 2124-2145 行的商品同步逻辑
    - 判断商品类型（普通商品 vs 套餐）
    - 普通商品使用 `productPackage.ErpCode` 调用 `UpdateProduct`
    - 套餐使用 `productBom.ErpCode` 调用 `UpdateProduct`
    - 添加详细的错误日志记录

### Phase 3: 相关场景同步修复

- [x] **3.1 检查 AddProduct 方法** `main/app/service/product.go`
  - **需求**: 检查新增商品时指定单位的逻辑，如需要则修复
  - **预计时间**: 0.5 小时
  - **负责人**: 
  - **代码位置**: 需要搜索 `AddProduct` 方法
  - **变更内容**: 如果新增商品时指定了单位，使用 `UpdateProduct` 同步到 ERP
  - **检查结果**: ✅ 新增商品时单位通过 `AddProductUnit` 方法关联，已在 Phase 2 修复

- [x] **3.2 检查 EditProduct 方法** `main/app/service/product.go`
  - **需求**: 检查编辑商品时修改单位的逻辑，如需要则修复
  - **预计时间**: 0.5 小时
  - **负责人**: 
  - **代码位置**: 需要搜索 `EditProduct` 方法
  - **变更内容**: 如果编辑商品时修改了单位，使用 `UpdateProduct` 同步更新 ERP
  - **检查结果**: ✅ 编辑商品时修改单位应通过 `EditProductUnit` 方法完成，已在 Phase 2 修复

### Phase 4: 错误处理和日志增强

- [x] **4.1 统一错误日志格式** `main/app/service/product.go`
  - **需求**: 在所有 ERP 同步调用处添加统一的错误日志格式
  - **预计时间**: 1 小时
  - **负责人**: 
  - **涉及方法**: `EditProductUnit`, `AddProductUnit`
  - **日志格式**: ✅ 已在 Phase 2 中添加到 `EditProductUnit` 和 `AddProductUnit` 方法

- [x] **4.2 改进错误提示信息** `main/app/service/product.go`
  - **需求**: 根据 ERP 错误类型返回用户友好的错误提示
  - **预计时间**: 1 小时
  - **负责人**: 
  - **变更内容**: ✅ 已在 Phase 2 中添加详细的错误日志，错误信息通过 `errors.WithMessage` 返回给用户
  - **说明**: ERP 错误类型识别需要在实际运行时根据错误信息判断，当前已记录详细日志便于排查

### Phase 5: 测试验证

- [ ] **5.1 编写单元测试 - UpdateProductReq** `main/app/service/rpc/erp/product_test.go`
  - **需求**: 测试 `UpdateProductReq` 结构体的 `StockUom` 字段序列化
  - **预计时间**: 1 小时
  - **负责人**: 
  - **测试用例**:
    - 验证 `StockUom` 字段正确序列化为 JSON
    - 验证字段为空时不影响现有逻辑

- [ ] **5.2 编写单元测试 - UpdateProduct RPC 调用** `main/app/service/rpc/erp/product_test.go`
  - **需求**: 测试 `UpdateProduct` 方法正确传递 `StockUom` 字段
  - **预计时间**: 1.5 小时
  - **负责人**: 
  - **测试用例**:
    - Mock gRPC 客户端，验证 `StockUom` 字段传递
    - 测试错误处理逻辑

- [ ] **5.3 编写集成测试 - 普通商品单位更新** `main/app/service/product_test.go`
  - **需求**: 测试普通商品关联单位时，ERP 模板单位正确更新
  - **预计时间**: 2 小时
  - **负责人**: 
  - **测试用例**:
    - 创建普通商品并关联单位
    - 验证 ERP 中模板 Item 的 `stock_uom` 字段正确更新
    - 验证变体 Item 的单位不受影响

- [ ] **5.4 编写集成测试 - 套餐单位更新** `main/app/service/product_test.go`
  - **需求**: 测试套餐关联单位时，ERP 套餐单位正确更新
  - **预计时间**: 2 小时
  - **负责人**: 
  - **测试用例**:
    - 创建套餐并关联单位
    - 验证 ERP 中套餐 Item 的 `stock_uom` 字段正确更新

- [ ] **5.5 编写错误场景测试** `main/app/service/product_test.go`
  - **需求**: 测试各种错误场景的处理
  - **预计时间**: 2 小时
  - **负责人**: 
  - **测试用例**:
    - 测试已销售商品单位更新（应返回错误）
    - 测试 ERP 服务不可用场景
    - 验证错误日志正确记录

- [ ] **5.6 手动测试 - 手机端单位编辑** 
  - **需求**: 在测试环境手动验证手机端单位编辑功能
  - **预计时间**: 1 小时
  - **负责人**: 
  - **测试步骤**:
    1. 登录手机端商家管理后台
    2. 编辑商品单位并关联商品
    3. 验证保存成功
    4. 验证 ERP 中单位正确更新

- [ ] **5.7 手动测试 - 多场景覆盖** 
  - **需求**: 测试所有相关业务场景
  - **预计时间**: 2 小时
  - **负责人**: 
  - **测试场景**:
    - 新增单位关联商品
    - 编辑单位关联商品
    - 新增商品指定单位
    - 编辑商品修改单位

### Phase 6: 代码审查和文档

- [ ] **6.1 代码审查**
  - **需求**: 通过 Code Review
  - **预计时间**: 1 小时
  - **负责人**: 
  - **审查要点**:
    - 代码逻辑正确性
    - 错误处理完整性
    - 日志记录规范性

- [ ] **6.2 更新 API 文档**（如需要）
  - **需求**: 如果涉及 API 变更，更新接口文档
  - **预计时间**: 0.5 小时
  - **负责人**: 
  - **文档位置**: `docs/shared/api/`

- [ ] **6.3 更新故障排查指南**
  - **需求**: 记录单位更新失败的处理流程
  - **预计时间**: 1 小时
  - **负责人**: 
  - **文档位置**: `docs/shared/troubleshooting/`

### Phase 7: 部署上线

- [ ] **7.1 发布到测试环境**
  - **需求**: 部署到测试环境并验证
  - **预计时间**: 1 小时
  - **负责人**: 
  - **验证内容**:
    - 功能正常
    - 错误处理正常
    - 日志记录正常

- [ ] **7.2 生产环境发布**
  - **需求**: 发布到生产环境并监控
  - **预计时间**: 1 小时
  - **负责人**: 
  - **监控指标**:
    - `UpdateProduct` 调用失败率
    - 单位关联操作失败率
    - ERP 同步错误日志

---

## 📊 任务统计

- **总任务数**: 20
- **已完成**: 8
- **进行中**: 0
- **未开始**: 12
- **完成率**: 40%

### 工作量估算

- **Phase 1**: 1 小时（前置条件）
- **Phase 2**: 3 小时（核心修复）
- **Phase 3**: 1 小时（相关场景）
- **Phase 4**: 2 小时（错误处理）
- **Phase 5**: 11.5 小时（测试验证）
- **Phase 6**: 2.5 小时（代码审查和文档）
- **Phase 7**: 2 小时（部署上线）

**总计**: 约 23 小时（约 3 个工作日）

---

## 🔗 相关链接

- **Bug 报告**: [bug.md](./bug.md)
- **修复方案**: [solution.md](./solution.md)
- **关联 Spec**: [task-erp-update-product-uom](../../shared/specs/active/task-erp-update-product-uom/requirements.md)
- **相关提案**: [ERP UpdateProduct 增加 UOM 字段更新支持](../../team/proposals/2025-11/erp-update-product-uom.md)

---

## 📝 备注

### 前置依赖

- ✅ BMP 模块的 `UpdateProduct` gRPC 接口已支持 `StockUom` 字段（已完成）
- ⏳ Main 模块的 RPC 层需要扩展支持 `StockUom` 字段（Phase 1）

### 风险提示

1. **ERP 业务规则限制**：
   - 已销售商品不允许修改单位，需要妥善处理错误提示
   - 普通商品只能修改模板单位，不能修改变体单位

2. **数据一致性**：
   - 修复后需要确保所有相关场景都使用新逻辑
   - 建议在测试环境充分验证后再发布生产

3. **回滚方案**：
   - 如果修复后出现问题，可以回滚代码
   - 保持数据库结构不变，无需数据迁移

---

**版本**: v1.0.0  
**创建日期**: 2025-11-28  
**维护者**: TTPOS Team

