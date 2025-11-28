# Bug-251128-004 修复方案

## 问题概述

手机端（Shop 商家管理端）在编辑商品单位并保存关联商品时，系统提示错误，无法成功保存。

**根本原因**：
1. 当前代码使用 `AddProduct` 和 `AddPackage` 方法更新已存在商品的单位，但这些方法不适合更新场景
2. ERP 上普通商品不可修改对应变体上的单位
3. ERP 上已销售商品单位不可修改
4. ERP 错误处理不完善，错误信息不够明确

## 根本原因

### 技术层面

1. **错误的 API 调用**：
   - 当前代码在编辑单位关联商品时，对普通商品使用 `erpSrv.AddProduct` 方法
   - 对套餐使用 `erpSrv.AddPackage` 方法
   - 这些方法用于创建新商品，不适合更新已存在商品的单位

2. **缺少前置条件支持**：
   - Main 模块的 `UpdateProductReq` 结构体缺少 `StockUom` 字段
   - 虽然 BMP 模块的 gRPC 接口已支持 `StockUom`，但 Main 模块的 RPC 封装层未传递该字段

3. **错误处理不足**：
   - ERP 接口返回错误时，没有记录详细的错误日志
   - 错误信息不够友好，用户无法理解失败原因

### 业务层面

1. **ERP 业务规则限制**：
   - ERPNext 不允许修改已销售商品的单位（有历史订单记录）
   - ERPNext 不允许修改普通商品变体上的单位（只能修改模板单位）

2. **数据一致性风险**：
   - 如果 ERP 同步失败，可能导致 TTPOS 与 ERP 数据不一致

## 修复方案

### 方案选择

**选项 1: 使用 UpdateProduct 更新单位（推荐）**

- **优点**：
  - 符合 ERPNext 的业务规则（更新模板单位而非变体单位）
  - 使用正确的更新 API，避免创建重复商品
  - BMP 模块已支持 `StockUom` 字段，只需 Main 模块适配
- **缺点**：
  - 需要修改 Main 模块的 RPC 封装层
  - 需要同步修复多个业务场景（新增单位、编辑单位、新增商品、编辑商品）
- **风险**：
  - 如果商品已销售，ERP 可能拒绝更新（需要错误处理）
  - 需要确保所有相关场景都使用新逻辑

**选项 2: 保持现有逻辑，仅改进错误处理**

- **优点**：
  - 代码改动最小
  - 风险较低
- **缺点**：
  - 无法解决根本问题（使用错误的 API）
  - 可能导致 ERP 中创建重复商品
  - 不符合 ERPNext 业务规则
- **风险**：
  - 数据一致性问题可能持续存在

**✅ 最终选择: 选项 1**

**理由**：
- 选项 1 从根本上解决了问题，使用正确的 API 更新单位
- 虽然需要更多代码修改，但能确保数据一致性和业务规则合规
- BMP 模块已支持 `StockUom` 字段，Main 模块只需适配即可

### 实施步骤

#### Phase 1: 前置条件 - Main 模块 RPC 层扩展

1. **扩展 UpdateProductReq 结构体**
   - 文件：`main/app/service/rpc/erp/product.go`
   - 添加 `StockUom` 字段

2. **更新 UpdateProduct 方法**
   - 文件：`main/app/service/rpc/erp/product.go`
   - 传递 `StockUom` 字段到 gRPC 调用

#### Phase 2: 核心修复 - 单位关联逻辑调整

3. **修复 EditProductUnit 方法**
   - 文件：`main/app/service/product.go`
   - 普通商品：使用 `UpdateProduct` 更新模板单位
   - 套餐：使用 `UpdateProduct` 更新单位

4. **修复 AddProductUnit 方法**
   - 文件：`main/app/service/product.go`
   - 新增单位时关联商品，使用 `UpdateProduct` 更新模板单位

#### Phase 3: 相关场景同步修复

5. **检查并修复 AddProduct 方法**
   - 文件：`main/app/service/product.go`
   - 如果新增商品时指定了单位，使用 `UpdateProduct` 同步到 ERP

6. **检查并修复 EditProduct 方法**
   - 文件：`main/app/service/product.go`
   - 如果编辑商品时修改了单位，使用 `UpdateProduct` 同步更新 ERP

#### Phase 4: 错误处理和日志

7. **增强错误处理**
   - 记录详细的错误日志（商品信息、单位信息、错误详情）
   - 根据错误类型决定是否中断事务
   - 返回用户友好的错误提示

### 技术方案

#### 数据结构变更

**Main 模块 UpdateProductReq 结构体扩展**：

```go
// main/app/service/rpc/erp/product.go
type UpdateProductReq struct {
    ItemCode     string                `json:"item_code"`
    NotForSale   bool                  `json:"not_for_sale"`
    InternalCode string                `json:"internal_code"`
    Disabled     bool                  `json:"disabled"`
    Attributes   []UpdateProductFlavor `json:"attributes"`
    StockUom     string                `json:"stock_uom"` // ✅ 新增字段
}
```

#### 代码修改

**1. UpdateProduct 方法更新**：

```go
// main/app/service/rpc/erp/product.go
func (s *erpSrv) UpdateProduct(ctx context.Context, params UpdateProductReq) error {
    // ... 现有代码 ...
    
    result, err := client.UpdateProduct(WithSiteCode(ctx.GetContext(), companySetting.ErpnextSiteCode), &item.UpdateProductReq{
        ItemCode:     params.ItemCode,
        NotForSale:   params.NotForSale,
        InternalCode: params.InternalCode,
        Disabled:     params.Disabled,
        Attributes:   attributes,
        StockUom:     params.StockUom, // ✅ 新增字段传递
    })
    // ... 现有代码 ...
}
```

**2. EditProductUnit 方法修复**：

```go
// main/app/service/product.go
// 在 EditProductUnit 方法中，替换 AddProduct 和 AddPackage 调用

// 普通商品：更新模板单位
if productPackage.IsProduct() {
    erpSrv := erp.NewIErpSrv(s.dbm)
    errErp := erpSrv.UpdateProduct(ctx, erp.UpdateProductReq{
        ItemCode: productPackage.ErpCode, // 使用模板 Item Code
        StockUom: productUnit.ErpnextUom,
    })
    if errErp != nil {
        // 记录详细错误日志
        ctx.Log().Error("同步商品单位到ERP失败",
            zap.String("product_package_uuid", fmt.Sprintf("%d", productPackage.Uuid)),
            zap.String("product_name", productPackage.Name),
            zap.String("erp_code", productPackage.ErpCode),
            zap.String("unit_uuid", fmt.Sprintf("%d", productUnit.Uuid)),
            zap.String("unit_name", productUnit.Name),
            zap.String("erp_uom", productUnit.ErpnextUom),
            zap.Error(errErp))
        return errors.WithMessage(errErp, "同步商品单位到ERP失败")
    }
}

// 套餐：更新套餐单位
if productPackage.IsPackage() {
    erpSrv := erp.NewIErpSrv(s.dbm)
    errErp := erpSrv.UpdateProduct(ctx, erp.UpdateProductReq{
        ItemCode: productBom.ErpCode, // 使用套餐 Item Code
        StockUom: productUnit.ErpnextUom,
    })
    if errErp != nil {
        // 记录详细错误日志
        ctx.Log().Error("同步套餐单位到ERP失败",
            zap.String("product_package_uuid", fmt.Sprintf("%d", productPackage.Uuid)),
            zap.String("product_bom_uuid", fmt.Sprintf("%d", productBom.Uuid)),
            zap.String("erp_code", productBom.ErpCode),
            zap.String("unit_uuid", fmt.Sprintf("%d", productUnit.Uuid)),
            zap.String("erp_uom", productUnit.ErpnextUom),
            zap.Error(errErp))
        return errors.WithMessage(errErp, "同步套餐单位到ERP失败")
    }
}
```

**3. AddProductUnit 方法修复**：

```go
// main/app/service/product.go
// 在 AddProductUnit 方法中，替换 AddProduct 调用

// 普通商品：更新模板单位
if productPackage.IsProduct() {
    erpSrv := erp.NewIErpSrv(s.dbm)
    errErp := erpSrv.UpdateProduct(ctx, erp.UpdateProductReq{
        ItemCode: productPackage.ErpCode, // 使用模板 Item Code
        StockUom: erpUom,
    })
    if errErp != nil {
        ctx.Log().Error("同步商品单位到ERP失败",
            zap.String("product_package_uuid", fmt.Sprintf("%d", productPackageUuid)),
            zap.String("erp_code", productPackage.ErpCode),
            zap.String("erp_uom", erpUom),
            zap.Error(errErp))
        return errors.WithMessage(errErp, "同步商品单位到ERP失败")
    }
}
```

#### 配置调整

**无配置调整** - 本修复不涉及配置文件变更。

## 影响分析

### 兼容性

- **向后兼容**：`StockUom` 字段为可选字段，不影响现有调用
- **API 兼容**：Main 模块的 `UpdateProduct` 方法签名不变，只是增加可选字段

### 性能影响

- **无性能影响**：使用 `UpdateProduct` 替代 `AddProduct`，API 调用次数相同
- **可能提升性能**：避免创建重复商品，减少 ERP 数据冗余

### 安全风险

- **低风险**：修复仅涉及单位更新逻辑，不涉及权限或数据访问控制
- **数据一致性**：修复后能确保 TTPOS 与 ERP 数据一致，降低数据不一致风险

## 测试计划

### 单元测试

1. **UpdateProductReq 结构体测试**
   - 验证 `StockUom` 字段正确序列化
   - 验证字段为空时不影响现有逻辑

2. **UpdateProduct RPC 调用测试**
   - 验证 `StockUom` 字段正确传递到 gRPC 调用
   - Mock ERP 响应，验证错误处理

### 集成测试

1. **普通商品单位更新测试**
   - 创建普通商品并关联单位
   - 验证 ERP 中模板 Item 的 `stock_uom` 字段正确更新
   - 验证变体 Item 的单位不受影响

2. **套餐单位更新测试**
   - 创建套餐并关联单位
   - 验证 ERP 中套餐 Item 的 `stock_uom` 字段正确更新

3. **错误场景测试**
   - 测试已销售商品单位更新（应返回错误）
   - 测试 ERP 服务不可用场景
   - 验证错误日志正确记录

### 手动测试

1. **手机端单位编辑测试**
   - 登录手机端商家管理后台
   - 编辑商品单位并关联商品
   - 验证保存成功
   - 验证 ERP 中单位正确更新

2. **多场景覆盖测试**
   - 新增单位关联商品
   - 编辑单位关联商品
   - 新增商品指定单位
   - 编辑商品修改单位

## 上线计划

### 发布时间

- **目标版本**：v2.10.10（或下一个版本）
- **预计发布时间**：待定

### 回滚方案

1. **代码回滚**：
   - 如果修复后出现问题，可以回滚到修复前的代码版本
   - 保持数据库结构不变，无需数据迁移

2. **功能回滚**：
   - 如果 ERP 同步出现问题，可以临时禁用 ERP 同步功能
   - 通过配置开关控制是否启用 ERP 同步

### 监控指标

1. **错误率监控**：
   - 监控 `UpdateProduct` 调用失败率
   - 监控单位关联操作失败率

2. **日志监控**：
   - 监控 ERP 同步错误日志
   - 监控错误日志中的商品和单位信息

3. **数据一致性监控**：
   - 定期检查 TTPOS 与 ERP 的单位数据一致性
   - 监控单位更新操作的执行时间

## 预防措施

### 代码层面

1. **API 使用规范**：
   - 明确 `AddProduct` 和 `UpdateProduct` 的使用场景
   - 在代码注释中说明何时使用哪个方法

2. **错误处理规范**：
   - 统一错误日志格式
   - 确保所有 ERP 调用都有错误处理和日志记录

### 测试层面

1. **自动化测试**：
   - 添加单位更新相关的集成测试
   - 覆盖普通商品和套餐两种场景

2. **回归测试**：
   - 在每次发布前，验证单位关联功能
   - 验证 ERP 同步功能正常

### 文档层面

1. **开发文档**：
   - 更新 ERP 集成文档，说明单位更新的正确方式
   - 记录常见错误和解决方案

2. **故障排查指南**：
   - 记录单位更新失败的处理流程
   - 记录如何检查 ERP 数据一致性

## 相关资源

### 关联 Spec

- [task-erp-update-product-uom](../../shared/specs/active/task-erp-update-product-uom/requirements.md) - ERP UpdateProduct 增加 UOM 字段支持

### 相关提案

- [ERP UpdateProduct 增加 UOM 字段更新支持](../../team/proposals/2025-11/erp-update-product-uom.md)

### 参考代码

- `main/app/service/rpc/erp/product.go` - UpdateProduct 方法
- `main/app/service/product.go` - EditProductUnit 和 AddProductUnit 方法
- `ttpos-bmp/app/ttpos-erp/internal/logic/stock/product.go` - BMP 模块 UpdateProduct 实现

---

**版本**: v1.0.0  
**创建日期**: 2025-11-28  
**维护者**: TTPOS Team

