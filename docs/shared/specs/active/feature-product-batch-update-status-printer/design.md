# 修改总部商品（状态和打印档口）- 设计文档

## 基本信息

- **Spec ID**: feature-product-batch-update-status-printer
- **关联需求**: [requirements.md](./requirements.md)
- **设计日期**: 2026-01-12
- **设计版本**: v1.0

---

## 设计概述

本功能新增一个API接口"修改总部商品上下架和打印档口"，用于子店修改总部来源商品的上下架状态和打印档口。设计遵循现有的商品管理模块架构，复用现有的数据验证和业务逻辑。

### 设计目标

1. **简洁性**：状态为必填参数，打印档口为可选参数
2. **一致性**：复用现有的验证逻辑和事务处理，与 ProductShopStatus 方法共享 updateProductStatus 内部方法
3. **兼容性**：不影响现有接口和功能
4. **可维护性**：代码结构清晰，易于理解和维护
5. **智能同步**：只同步子店自己的商品到 ERP，保护子店对总部商品的个性化配置

---

## 技术架构

### 技术栈

- **语言**: Go 1.23+
- **框架**: Gin
- **数据库**: MySQL 8.0+
- **ORM**: GORM

### 分层设计

```
API 层 (shop_product.go)
    ↓
Service 层 (product.go)
    ↓
Repository 层 (product.go)
    ↓
Model 层 (product_package.go, product_printer.go)
    ↓
Database
```

---

## 详细设计

### 1. API 层设计

#### 新增接口

**文件位置**: `main/app/api/v1/shop/shop_product.go`

**接口定义**:

```go
// UpdateHeadquartersProduct 修改总部商品（状态和打印档口）
// @Summary 修改总部商品
// @Description 修改总部商品的上下架状态和打印档口，支持单独或同时修改
// @Tags 商家端.商品
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.UpdateHeadquartersProductReq true "修改总部商品请求"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/product/update_headquarters_product [post]
func (h *ProductHandler) UpdateHeadquartersProduct(c *gin.Context) {
    ctx := helper.GetContext(c)
    updateReq := req.UpdateHeadquartersProductReq{}
    
    if err := c.ShouldBindJSON(&updateReq); err != nil {
        helper.HandleValidationError(c, err, updateReq, dto.PageReqMessage)
        return
    }
    
    // 调用服务层
    err := h.productSrv.UpdateHeadquartersProduct(ctx, updateReq)
    if err != nil {
        helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
        return
    }
    
    helper.Success(c, nil, "修改成功")
}
```

**路由注册**:

在 `RegisterProductRouter` 函数末尾添加：

```go
privateApi.POST("/product/update_headquarters_product", wrapper.UpdateHeadquartersProduct) // 修改总部商品上下架和打印档口
```

**注意**：路由放在文件末尾，与其他外卖相关接口放在一起

---

### 2. DTO 层设计

#### 新增请求结构体

**文件位置**: `main/app/dto/req/product.go`

```go
// UpdateHeadquartersProductReq 修改总部商品请求
type UpdateHeadquartersProductReq struct {
    Uuid                uint64   `json:"uuid" binding:"required"`             // 商品UUID（必填）
    Status              *int     `json:"status" binding:"required,oneof=0 1"` // 商品状态 0-下架 1-上架（必填）
    ProductPrinterUuids []uint64 `json:"product_printer_uuids"`               // 商品打印档口UUID列表（可选）
}
```

**注意**：
- `Status` 字段使用 `binding:"required"` 标签，框架会自动验证
- 不需要额外的 `Validate()` 方法
- 打印档口仍然是可选参数

---

### 3. Service 层设计

#### 新增服务接口

**文件位置**: `main/app/service/product.go`

**接口定义**:

```go
type IProductSrv interface {
    // ... 现有方法
    
    // UpdateHeadquartersProduct 修改总部商品（状态和打印档口）
    UpdateHeadquartersProduct(ctx context.Context, req req.UpdateHeadquartersProductReq) error
}
```

**实现逻辑**:

```go
// UpdateHeadquartersProduct 修改总部商品（状态和打印档口）
func (s *productSrv) UpdateHeadquartersProduct(ctx context.Context, req req.UpdateHeadquartersProductReq) error {
    // 1. 参数验证
    if req.Status == nil {
        return errors.New("商品状态不能为空")
    }
    
    db := s.dbm.GetDB(ctx.GetDbId())
    commonRepo := repository.NewCommonRepo()
    productRepo := repository.NewProductRepo(db)
    
    // 2. 查询商品是否存在
    product, err := productRepo.GetProduct(
        commonRepo.WhereBySoftDelete(),
        productRepo.WhereUuid(req.Uuid),
        productRepo.WithProductBoms(commonRepo.WhereBySoftDelete()),
    )
    if err != nil {
        return errors.WithMessage(err, "获取商品失败")
    }
    if product.ID == 0 {
        return errors.New("商品不存在")
    }
    
    // 3. 验证打印档口
    if req.ProductPrinterUuids != nil {
        productCheckSrv := NewProductCheckSrv(s.dbm, s.localeSrv, s.settingSrv)
        err = productCheckSrv.CheckProductPrinter(ctx, db, req.Uuid, req.ProductPrinterUuids)
        if err != nil {
            return errors.WithMessage(err, "验证打印档口失败")
        }
    }
    
    // 4. 开启事务执行更新
    err = db.Transaction(func(tx *gorm.DB) error {
        // 4.1 更新商品状态
        if err := s.updateProductStatus(tx, &product, req.Status); err != nil {
            return err
        }
        
        // 4.2 更新打印档口
        if req.ProductPrinterUuids != nil {
            productPrinterRepo := repository.NewProductPrinterRepo(tx)
            err = productPrinterRepo.CreateProductPackagePrinter(product.Uuid, req.ProductPrinterUuids)
            if err != nil {
                return errors.WithMessage(err, "更新打印档口关联失败")
            }
        }
        
        return nil
    })
    
    if err != nil {
        return errors.WithMessage(err, "修改总部商品失败")
    }
    
    return nil
}

// updateProductStatus 更新商品状态（内部方法）
// 此方法被 ProductShopStatus 和 UpdateHeadquartersProduct 共同使用
func (s *productSrv) updateProductStatus(tx *gorm.DB, productPackage *model.ProductPackage, status *int) error {
    commonRepo := repository.NewCommonRepo()
    productPackageGroupRepo := repository.NewProductPackageGroupRepo(tx)

    // 更新商品状态
    err := tx.Model(&model.ProductPackage{}).
        Select("status").
        Where("uuid = ?", productPackage.Uuid).
        Updates(map[string]any{"status": status}).Error
    if err != nil {
        return errors.WithMessage(err, "修改商品状态失败")
    }

    // 更新商品规格状态
    err = tx.Model(&model.ProductBom{}).
        Select("status").
        Where("product_package_uuid = ?", productPackage.Uuid).
        Updates(map[string]any{"status": status}).Error
    if err != nil {
        return errors.WithMessage(err, "修改商品规格状态失败")
    }

    // 如果是下架商品，需要下架引用该商品的套餐
    if productPackage.ProductType == constant.ProductTypeProduct && *status == 0 {
        productPackageGroupItems, err := productPackageGroupRepo.GetProductPackageGroupItems(
            commonRepo.WhereBySoftDelete(),
            commonRepo.WhereByRelatedUuid(productPackage.Uuid),
            productPackageGroupRepo.WithProductPackageGroup(commonRepo.WhereBySoftDelete()),
            productPackageGroupRepo.WithProductPackageGroupProduct(commonRepo.WhereBySoftDelete()),
        )
        if err != nil {
            return errors.WithMessage(err, "获取商品套餐组商品失败")
        }

        for _, item := range productPackageGroupItems {
            if item.ProductPackageGroup != nil && item.ProductPackageGroup.ProductPackage != nil {
                // 下架套餐
                err = tx.Model(&model.ProductPackage{}).
                    Select("status").
                    Where("uuid = ?", item.ProductPackageGroup.ProductPackage.Uuid).
                    Updates(map[string]any{"status": 0}).Error
                if err != nil {
                    return errors.WithMessage(err, "修改商品套餐状态失败")
                }

                // 下架套餐规格
                err = tx.Model(&model.ProductBom{}).
                    Select("status").
                    Where("product_package_uuid = ?", item.ProductPackageGroup.ProductPackage.Uuid).
                    Updates(map[string]any{"status": 0}).Error
                if err != nil {
                    return errors.WithMessage(err, "修改商品套餐组商品状态失败")
                }
            }
        }
    }

    return nil
}
```

**注意**：
- 移除了 `updateProductPrinters` 方法，直接使用 Repository 层的 `CreateProductPackagePrinter` 方法
- `updateProductStatus` 方法被 `ProductShopStatus` 和 `UpdateHeadquartersProduct` 共同复用
- ProductShopStatus 方法也进行了重构，使用相同的 `updateProductStatus` 方法

---

### 4. Repository 层设计

**复用现有 Repository**，不需要新增方法：

- `ProductRepo.GetProduct()` - 查询商品
- `ProductPackageGroupRepo.GetProductPackageGroupItems()` - 查询套餐关联
- GORM 原生方法 - 执行更新和删除

---

### 5. Model 层设计

**复用现有 Model**，不需要修改：

- `ProductPackage` - 商品表
- `ProductBom` - 商品规格表
- `ProductPackageProductPrinter` - 商品打印档口关联表
- `ProductPrinter` - 打印档口表

---

## 数据库操作

### 查询操作

```sql
-- 查询商品
SELECT * FROM product_package 
WHERE uuid = ? AND deleted_at IS NULL;

-- 查询打印档口（验证）
SELECT * FROM product_printer 
WHERE uuid IN (?) AND deleted_at IS NULL;

-- 查询套餐关联（如果下架商品）
SELECT * FROM product_package_group_item 
WHERE related_uuid = ? AND deleted_at IS NULL;
```

### 更新操作

```sql
-- 更新商品状态
UPDATE product_package 
SET status = ? 
WHERE uuid = ?;

-- 更新商品规格状态
UPDATE product_bom 
SET status = ? 
WHERE product_package_uuid = ?;

-- 删除原有打印档口关联
DELETE FROM product_package_product_printer 
WHERE product_package_uuid = ?;

-- 插入新的打印档口关联
INSERT INTO product_package_product_printer 
(product_package_uuid, product_printer_uuid) 
VALUES (?, ?);
```

---

## 错误处理

### 错误类型

| 错误场景 | HTTP 状态码 | 错误码 | 错误消息 |
|---|---|---|---|
| Status 缺失 | 400 | -1 | 商品状态不能为空 |
| Status 值非法 | 400 | -1 | (框架自动验证) |
| 商品不存在 | 400 | -1 | 商品不存在 |
| 打印档口不存在 | 400 | -1 | 商品打印档口不存在 |
| 数据库操作失败 | 400 | -1 | 修改总部商品失败 |

### 错误处理流程

```go
1. Status 参数验证 → 返回"商品状态不能为空"
2. 商品查询 → 返回"商品不存在"
3. 打印档口验证 → 返回"商品打印档口不存在"
4. 事务执行 → 自动回滚，返回操作失败
```

---

## 性能优化

### 优化策略

1. **事务优化**：
   - 所有数据库操作在一个事务中完成
   - 失败时自动回滚，保证数据一致性

2. **查询优化**：
   - 使用索引字段（uuid）进行查询
   - 批量查询打印档口，避免N+1问题

3. **代码复用**：
   - 与 ProductShopStatus 共享 `updateProductStatus` 内部方法
   - 使用 Repository 层已有的 `CreateProductPackagePrinter` 方法

4. **智能同步**：
   - 在 ProductShopStatus 中添加 ERP 同步判断逻辑
   - 只同步子店自己创建的商品（通过 `isEditable` 函数判断 HeadquarterUuid）
   - 保护子店对总部商品的个性化配置

5. **同步保护**：
   - 在同步功能中，如果商品已存在，保留子店已修改的状态
   - 避免总部商品同步时覆盖子店的个性化状态

---

## 安全设计

### 权限控制

- 接口使用 JWT 鉴权
- 通过 `ctx.GetCompany()` 获取当前门店信息
- 打印档口验证时，确保只能选择当前门店的打印档口

### 数据验证

1. **商品UUID验证**：必须存在且未删除
2. **状态值验证**：必须为 0 或 1
3. **打印档口验证**：必须存在、未删除且属于当前门店

### SQL注入防护

- 使用 GORM 参数化查询
- 不拼接 SQL 字符串

---

## 测试方案

### 单元测试

**测试文件**: `main/app/service/product_test.go`

测试用例：

1. `TestUpdateHeadquartersProduct_UpdateStatusOnly` - 仅更新状态（不提供打印档口）
2. `TestUpdateHeadquartersProduct_UpdateBoth` - 同时更新状态和打印档口
3. `TestUpdateHeadquartersProduct_ProductNotFound` - 商品不存在
4. `TestUpdateHeadquartersProduct_InvalidPrinter` - 打印档口不存在
5. `TestUpdateHeadquartersProduct_MissingStatus` - 缺少 Status 参数
6. `TestUpdateHeadquartersProduct_OfflineProduct` - 下架商品验证关联套餐
7. `TestUpdateHeadquartersProduct_NoERPSync` - 验证总部商品不同步到 ERP

### 集成测试

**测试文件**: `main/test/api/shop/product_test.go`

测试场景：

1. 成功场景：调用接口并验证数据库变更
2. 失败场景：验证错误响应和数据未变更
3. 并发场景：并发修改同一商品

### 手动测试

使用 Postman 或 Swagger 测试：

1. 调用接口修改状态（不提供打印档口）
2. 调用接口同时修改状态和打印档口
3. 验证商品详情接口返回的数据
4. 验证 ERP 同步逻辑（总部商品不同步，子店商品同步）

---

## 部署方案

### 部署步骤

1. **代码部署**：
   - 合并代码到 dev 分支
   - 执行自动化测试
   - 部署到测试环境

2. **数据库迁移**：
   - 无需数据库迁移（复用现有表）

3. **配置更新**：
   - 无需配置更新

4. **验证**：
   - 执行冒烟测试
   - 验证新接口和原有接口均正常

### 回滚方案

- 如果出现问题，回滚代码到上一个稳定版本
- 新接口不影响现有功能，可以安全回滚

---

## 监控与告警

### 监控指标

1. **接口调用量**：`/shop/product/batch_update` 的 QPS
2. **接口响应时间**：P50、P95、P99
3. **错误率**：4xx 和 5xx 错误比例
4. **数据库慢查询**：超过 500ms 的查询

### 告警规则

1. 接口错误率 > 5% → 发送告警
2. 接口响应时间 P95 > 1s → 发送告警
3. 数据库慢查询 > 10 次/分钟 → 发送告警

---

## 时间估算

| 任务 | 工作量（小时） | 说明 |
|---|---|---|
| DTO 层开发 | 0.3 | 新增请求结构体（无需验证方法） |
| Service 层开发 | 2.0 | 实现业务逻辑、事务处理和代码复用优化 |
| API 层开发 | 0.3 | 新增接口和路由注册 |
| 代码重构 | 0.5 | 重构 ProductShopStatus 方法，提取公共逻辑 |
| 单元测试 | 1.5 | 编写服务层单元测试 |
| 集成测试 | 1.0 | 编写API集成测试 |
| 文档编写 | 1.0 | 更新API文档 |
| 联调测试 | 1.0 | 与前端联调 |
| **总计** | **7.6** | 约 1 个工作日 |

---

## 相关文档

- [需求文档](./requirements.md)
- [任务清单](./tasks.md)
- [Go Main 开发规范](../../../../.cursor/rules/go-main.mdc)
- [API 设计规范](../../../../.cursor/rules/api.mdc)

---

## 变更历史

| 日期 | 版本 | 变更内容 | 变更人 |
|---|---|---|---|
| 2026-01-12 | v1.0 | 初始版本 | AI Agent |
| 2026-01-12 | v1.1 | 更新文档以反映代码修改：Status 改为必填、删除 Validate 方法、优化代码复用、增加 ERP 同步判断、优化同步功能 | AI Agent |

---
