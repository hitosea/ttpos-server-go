# Shop 商家管理端-套餐管理移除ERP集成 设计文档

> 本文档定义 Shop 商家管理端套餐管理移除ERP集成的技术设计和实现方案。

## 📋 概述

移除套餐商品的添加、修改、删除、同步操作中对ERP接口的调用，使套餐管理功能独立运行，不再依赖ERP系统。**关键约束：必须确保普通商品的ERP同步功能不受影响。**

通过使用 `productPackage.IsPackage()` 和 `productPackage.IsProduct()` 方法判断商品类型，只移除套餐相关的ERP调用，保留普通商品的ERP同步逻辑。

---

## 🎯 规范对齐

### Go Main 规范 (go-main.mdc)

- Service 只依赖其他 Service 接口
- Repository 只持有 db 实例
- URL 使用 snake_case
- data 字段必须是对象
- 不使用 panic，返回 error
- 遵循错误处理规范，使用 `errors.WithMessage` 包装错误

### API 设计规范 (api.mdc)

- URL 使用 snake_case
- 响应格式统一：`{code, message, data{}}`
- data 不能为 null 或数组

---

## 🔄 代码复用分析

### 可复用的现有组件

- **ProductPackage Model**: `main/app/model/product.go` - 提供 `IsPackage()` 和 `IsProduct()` 方法判断商品类型
- **ERP Service**: `main/app/service/rpc/erp/product.go` - ERP接口调用实现（保留普通商品调用）
- **Product Service**: `main/app/service/product.go` - 套餐和商品管理业务逻辑

### 集成点

- **商品类型判断**: 使用 `productPackage.IsPackage()` 判断是否为套餐
- **ERP调用逻辑**: 在 `product.go` 和 `sync_product_to_erp.go` 中移除套餐相关的ERP调用
- **数据库操作**: 保留本地数据库操作，确保套餐数据正常保存

---

## 🏗️ 架构设计

### 分层设计原则

**Go Main 三层架构**:

```
API 层 (Controller/API)
  ↓ 依赖
业务层 (Service)
  ↓ 依赖
数据层 (Repository)
```

### 修改点分析

#### 1. 添加套餐逻辑 (`product.go`)

**位置**: `main/app/service/product.go` - `AddProductFlavor` 方法

**当前逻辑**:
```go
if productPackage.IsPackage() {
    // 调用 erpSrv.AddPackage() 同步套餐到ERP
    itemInfo, errErp := erpSrv.AddPackage(ctx, params)
    // ...
}
```

**修改后逻辑**:
```go
if productPackage.IsPackage() {
    // 套餐：不再调用ERP接口，直接保存到本地数据库
    // 移除 erpSrv.AddPackage() 调用
    // 保留本地数据库操作
} else {
    // 普通商品：继续调用ERP接口（不受影响）
    if ctx.GetCompany().IsOpenErp() {
        erpSrv.AddProductBom(...)
    }
}
```

#### 2. 修改套餐逻辑 (`product.go`)

**位置**: `main/app/service/product.go` - `UpdateProductUnit` 方法

**当前逻辑**:
```go
if productPackage.IsPackage() {
    // 套餐：调用 erpSrv.UpdateProduct() 更新套餐单位
    errErp := erpSrv.UpdateProduct(ctx, erp.UpdateProductReq{...})
    // ...
} else if productPackage.IsProduct() {
    // 普通商品：调用 erpSrv.UpdateProduct() 更新模板单位
    errErp := erpSrv.UpdateProduct(ctx, erp.UpdateProductReq{...})
    // ...
}
```

**修改后逻辑**:
```go
if productPackage.IsPackage() {
    // 套餐：不再调用ERP接口，直接更新本地数据库
    // 移除 erpSrv.UpdateProduct() 调用
    // 保留本地数据库更新操作
} else if productPackage.IsProduct() {
    // 普通商品：继续调用ERP接口（不受影响）
    if ctx.GetCompany().IsOpenErp() {
        errErp := erpSrv.UpdateProduct(ctx, erp.UpdateProductReq{...})
        // ...
    }
}
```

#### 3. 删除套餐逻辑 (`product.go`)

**位置**: `main/app/service/product.go` - `DeleteProductShop` 方法

**当前逻辑**:
```go
if ctx.GetCompany().IsOpenErp() {
    erpSrv := erp.NewIErpSrv(s.dbm)
    if !product.IsPackage() {
        // 普通商品：调用 erpSrv.UpdateProduct() 设置禁售
        errErp := erpSrv.UpdateProduct(ctx, erp.UpdateProductReq{
            ItemCode:   product.ErpCode,
            NotForSale: true,
            // ...
        })
        // ...
    } else {
        // 套餐：调用 erpSrv.DeleteProduct() 删除套餐到ERP
        items := []req.DeleteProductErpItemReq{}
        items = append(items, req.DeleteProductErpItemReq{
            ItemCode: erpCode,
            ItemName: enName,
            StockUom: product.ProductUnit.ErpnextUom,
        })
        errErp := erpSrv.DeleteProduct(ctx, req.DeleteProductErpReq{
            Items: items,
        })
        if errErp != nil {
            return errors.WithMessage(errErp, "删除商品到erp失败")
        }
    }
}
```

**修改后逻辑**:
```go
if ctx.GetCompany().IsOpenErp() {
    erpSrv := erp.NewIErpSrv(s.dbm)
    if !product.IsPackage() {
        // 普通商品：继续调用ERP接口设置禁售（不受影响）
        errErp := erpSrv.UpdateProduct(ctx, erp.UpdateProductReq{
            ItemCode:   product.ErpCode,
            NotForSale: true,
            // ...
        })
        if errErp != nil {
            return errors.WithMessage(errErp, "设置商品模板禁售失败")
        }
        // ...
    } else {
        // 套餐：不再调用ERP接口，直接软删除本地数据库
        // 移除 erpSrv.DeleteProduct() 调用
        // 保留本地数据库软删除操作
    }
}
```

#### 4. 同步套餐逻辑 (`sync_product_to_erp.go`)

**位置**: `main/app/service/sync_product_to_erp.go` - `SyncProductToErp` 方法

**当前逻辑**:
```go
for _, bom := range productPackage.ProductBoms {
    if productPackage.IsPackage() {
        // 套餐：调用 erpSrv.AddPackage() 同步套餐到ERP
        itemInfo, errErp := erpSrv.AddPackage(ctx, params)
        // ...
    } else {
        // 普通商品：调用 erpSrv.AddProductBom() 同步商品到ERP
        itemInfo, errErp := erpSrv.AddProductBom(...)
        // ...
    }
}
```

**修改后逻辑**:
```go
for _, bom := range productPackage.ProductBoms {
    if productPackage.IsPackage() {
        // 套餐：跳过ERP同步，不再调用ERP接口
        // 移除 erpSrv.AddPackage() 调用
        continue // 或跳过当前循环
    } else {
        // 普通商品：继续调用ERP接口（不受影响）
        if ctx.GetCompany().IsOpenErp() {
            itemInfo, errErp := erpSrv.AddProductBom(...)
            // ...
        }
    }
}
```

---

## 🗄️ 数据库设计

### 数据表设计

**无需修改数据库表结构**，本次变更仅涉及业务逻辑层面的ERP调用移除，不涉及数据库结构变更。

### 数据一致性

- **套餐数据**: 仅在TTPOS系统内部管理，不再同步到ERP
- **普通商品数据**: 继续同步到ERP，保持现有逻辑不变

---

## 📊 数据模型

### 商品类型判断方法

**位置**: `main/app/model/product.go`

```go
// 是否套餐
func (model *ProductPackage) IsPackage() bool {
    return model.ProductType == constant.ProductTypePackage
}

// 是否商品
func (model *ProductPackage) IsProduct() bool {
    return model.ProductType == constant.ProductTypeProduct
}
```

**使用方式**:
- `productPackage.IsPackage()` - 判断是否为套餐
- `productPackage.IsProduct()` - 判断是否为普通商品

---

## 🔌 API 设计

### RESTful API

**无需新增或修改API接口**，本次变更仅涉及后端业务逻辑，API接口保持不变。

### 影响范围

- **添加套餐API**: `/api/v1/shop/product/flavor/add` - 不再调用ERP，响应更快
- **修改套餐API**: `/api/v1/shop/product/unit/update` - 不再调用ERP，响应更快
- **删除套餐API**: `/api/v1/shop/product/delete` - 不再调用ERP，响应更快
- **同步商品API**: `/api/v1/shop/product/sync_to_erp` - 套餐跳过ERP同步，普通商品正常同步

---

## 🧩 组件和接口

### Service 层修改

#### 1. Product Service - AddProductFlavor 方法

**文件**: `main/app/service/product.go`

**修改点**: 移除套餐添加时的ERP调用

```go
// 修改前
if productPackage.IsPackage() {
    itemInfo, errErp := erpSrv.AddPackage(ctx, params)
    if errErp != nil {
        return errors.WithMessage(errErp)
    }
    erpCode = itemInfo.ItemCode
}

// 修改后
if productPackage.IsPackage() {
    // 套餐：不再调用ERP接口，直接保存到本地数据库
    // erpCode 可以设置为空字符串或保留原有逻辑（如果不需要ERP编码）
    erpCode = "" // 或保留原有逻辑
} else {
    // 普通商品：继续调用ERP接口（不受影响）
    if ctx.GetCompany().IsOpenErp() {
        itemInfo, errErp := erpSrv.AddProductBom(...)
        // ...
    }
}
```

#### 2. Product Service - UpdateProductUnit 方法

**文件**: `main/app/service/product.go`

**修改点**: 移除套餐单位更新时的ERP调用

```go
// 修改前
if productPackage.IsPackage() {
    errErp := erpSrv.UpdateProduct(ctx, erp.UpdateProductReq{
        ItemCode:   productBom.ErpCode,
        StockUom:   erpUom,
        // ...
    })
    if errErp != nil {
        return errors.WithMessage(errErp, "同步套餐单位到ERP失败")
    }
} else if productPackage.IsProduct() {
    errErp := erpSrv.UpdateProduct(ctx, erp.UpdateProductReq{...})
    // ...
}

// 修改后
if productPackage.IsPackage() {
    // 套餐：不再调用ERP接口，直接更新本地数据库
    // 移除 erpSrv.UpdateProduct() 调用
    // 保留本地数据库更新操作
} else if productPackage.IsProduct() {
    // 普通商品：继续调用ERP接口（不受影响）
    if ctx.GetCompany().IsOpenErp() {
        errErp := erpSrv.UpdateProduct(ctx, erp.UpdateProductReq{...})
        if errErp != nil {
            return errors.WithMessage(errErp, "同步商品单位到ERP失败")
        }
    }
}
```

#### 3. Product Service - DeleteProductShop 方法

**文件**: `main/app/service/product.go`

**修改点**: 移除套餐删除时的ERP调用

```go
// 修改前
if ctx.GetCompany().IsOpenErp() {
    erpSrv := erp.NewIErpSrv(s.dbm)
    if !product.IsPackage() {
        // 普通商品：调用 erpSrv.UpdateProduct() 设置禁售
        errErp := erpSrv.UpdateProduct(ctx, erp.UpdateProductReq{
            ItemCode:   product.ErpCode,
            NotForSale: true,
            // ...
        })
        // ...
    } else {
        // 套餐：调用 erpSrv.DeleteProduct() 删除套餐到ERP
        items := []req.DeleteProductErpItemReq{}
        items = append(items, req.DeleteProductErpItemReq{
            ItemCode: erpCode,
            ItemName: enName,
            StockUom: product.ProductUnit.ErpnextUom,
        })
        errErp := erpSrv.DeleteProduct(ctx, req.DeleteProductErpReq{
            Items: items,
        })
        if errErp != nil {
            return errors.WithMessage(errErp, "删除商品到erp失败")
        }
    }
}

// 修改后
if ctx.GetCompany().IsOpenErp() {
    erpSrv := erp.NewIErpSrv(s.dbm)
    if !product.IsPackage() {
        // 普通商品：继续调用ERP接口设置禁售（不受影响）
        errErp := erpSrv.UpdateProduct(ctx, erp.UpdateProductReq{
            ItemCode:   product.ErpCode,
            NotForSale: true,
            // ...
        })
        if errErp != nil {
            return errors.WithMessage(errErp, "设置商品模板禁售失败")
        }
        // ...
    } else {
        // 套餐：不再调用ERP接口，直接软删除本地数据库
        // 移除 erpSrv.DeleteProduct() 调用
        // 保留本地数据库软删除操作（已在事务中执行）
    }
}
```

#### 4. Sync Product To ERP Service

**文件**: `main/app/service/sync_product_to_erp.go`

**修改点**: 移除套餐同步时的ERP调用

```go
// 修改前
for _, bom := range productPackage.ProductBoms {
    if productPackage.IsPackage() {
        itemInfo, errErp := erpSrv.AddPackage(ctx, params)
        if errErp != nil {
            return errors.WithMessage(errErp)
        }
        bomErpCode = itemInfo.ItemCode
    } else {
        itemInfo, errErp := erpSrv.AddProductBom(...)
        // ...
    }
}

// 修改后
for _, bom := range productPackage.ProductBoms {
    if productPackage.IsPackage() {
        // 套餐：跳过ERP同步，不再调用ERP接口
        continue // 跳过当前循环，不处理套餐的ERP同步
    } else {
        // 普通商品：继续调用ERP接口（不受影响）
        if ctx.GetCompany().IsOpenErp() {
            itemInfo, errErp := erpSrv.AddProductBom(...)
            if errErp != nil {
                return errors.WithMessage(errErp, "同步商品到erp失败")
            }
            bomErpCode = itemInfo.ItemCode
        }
    }
}
```

---

## ⚡ 缓存设计

**无需修改缓存策略**，本次变更不涉及缓存逻辑。

---

## 🚨 错误处理

### 错误场景

#### 场景 1: 套餐添加/修改/删除时不再处理ERP错误

- **处理方式**: 移除ERP相关的错误处理和日志记录
- **用户影响**: 套餐操作不再因ERP服务问题而失败，操作更稳定
- **代码示例**:
  ```go
  // 移除以下代码
  if errErp != nil {
      logger.Logger.Error("同步套餐到ERP失败", ...)
      return errors.WithMessage(errErp, "同步套餐到ERP失败")
  }
  
  // 删除套餐时移除以下代码
  if errErp != nil {
      logger.Logger.Error("删除商品到erp失败", ...)
      return errors.WithMessage(errErp, "删除商品到erp失败")
  }
  ```

#### 场景 2: 普通商品ERP调用失败

- **处理方式**: 保持现有错误处理逻辑不变
- **用户影响**: 普通商品的ERP同步失败时，仍会返回错误（不受影响）

---

## 🔒 安全设计

**无需修改安全设计**，本次变更不涉及安全相关逻辑。

---

## 🧪 测试策略

### 单元测试

**目标覆盖率**:
- Service 层: 70%+
- Repository 层: 80%+

**测试内容**:
1. **套餐添加测试**: 验证套餐添加时不再调用ERP接口
2. **套餐修改测试**: 验证套餐修改时不再调用ERP接口
3. **套餐删除测试**: 验证套餐删除时不再调用ERP接口
4. **套餐同步测试**: 验证套餐同步时跳过ERP调用
5. **普通商品测试**: 验证普通商品的ERP同步功能不受影响

**测试用例**:
- 添加套餐：验证不调用 `erpSrv.AddPackage()`
- 修改套餐：验证不调用 `erpSrv.UpdateProduct()`
- 删除套餐：验证不调用 `erpSrv.DeleteProduct()`
- 同步套餐：验证跳过ERP同步逻辑
- 添加普通商品：验证继续调用 `erpSrv.AddProductBom()`
- 修改普通商品：验证继续调用 `erpSrv.UpdateProduct()`
- 删除普通商品：验证继续调用 `erpSrv.UpdateProduct()` 设置禁售
- 同步普通商品：验证继续调用ERP接口

### API 测试

**测试内容**:
- API 接口调用
- 参数验证
- 响应格式
- 错误处理

### 集成测试

**测试流程**:
1. 端到端测试：添加套餐 → 验证数据保存 → 验证不调用ERP
2. 端到端测试：修改套餐 → 验证数据更新 → 验证不调用ERP
3. 端到端测试：删除套餐 → 验证数据软删除 → 验证不调用ERP
4. 端到端测试：同步商品 → 验证套餐跳过ERP → 验证普通商品正常同步

---

## 📈 性能优化

### 优化策略

1. **移除ERP调用**: 套餐操作不再等待ERP接口响应，响应时间明显缩短
2. **减少网络请求**: 套餐操作不再发起ERP网络请求，减少网络延迟
3. **提高稳定性**: 套餐操作不再依赖ERP服务状态，提高系统稳定性

### 性能指标

- **套餐操作响应时间**: < 200ms（移除ERP调用后）
- **普通商品操作响应时间**: 保持不变（仍包含ERP调用耗时）

---

## 🌐 浏览器兼容性

**无需修改前端代码**，本次变更仅涉及后端业务逻辑。

---

## 📚 实现清单

### Phase 1: 代码分析和准备

- [x] 分析现有代码结构
- [x] 识别所有套餐相关的ERP调用位置
- [x] 确认商品类型判断方法

### Phase 2: 移除ERP调用

- [ ] 修改 `product.go` - `AddProductFlavor` 方法（添加套餐）
- [ ] 修改 `product.go` - `UpdateProductUnit` 方法（修改套餐单位）
- [ ] 修改 `product.go` - `DeleteProductShop` 方法（删除套餐）
- [ ] 修改 `sync_product_to_erp.go` - `SyncProductToErp` 方法（同步套餐）
- [ ] 清理相关的错误处理和日志记录

### Phase 3: 测试和验证

- [ ] 单元测试：套餐操作不调用ERP（添加、修改、删除、同步）
- [ ] 单元测试：普通商品ERP同步不受影响
- [ ] 集成测试：端到端流程测试（包含删除套餐）
- [ ] 回归测试：验证普通商品功能正常

**详细任务**: 参见 `tasks.md`

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 在技术方案评审、关键架构决策或踩坑总结后，立即补充 Episode 并在设计文档尾部互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-12-01  
**作者**: 王昱  
**审核者**: {审核者}

