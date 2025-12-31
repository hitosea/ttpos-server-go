# 对象存储层 需求文档

> 本文档定义对象存储层的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-12/object-storage-layer.md](../../../../team/proposals/2025-12/object-storage-layer.md) |
| **创建日期**      | 2025-12-24                                                                                                 |
| **负责人**        | xiezhihuan                                                                                                       |
| **目标 Sprint**   | Sprint {N}                                                                                                   |
| **涉及技术栈**    | [x] Go (main/) [ ] Go (ttpos-bmp/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                   |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 待审核 |
| **审核人**   | {审核人姓名}             |
| **审核日期** | {YYYY-MM-DD}             |
| **审核意见** | {审核意见，如有}         |

---

## 📋 概述

对象存储层（Object Storage Layer）是一个统一的基础设施层，提供通过 key 获取模型对象的统一接口。该层基于三级缓存基础包，自动处理缓存查询、数据库回填和对象生命周期管理，旨在减少代码重复，统一缓存访问模式，提升开发效率和代码质量。

**核心价值**：
- **提升开发效率**：减少重复代码，统一缓存访问模式
- **降低维护成本**：集中管理对象生命周期，便于统一优化和调试
- **提高代码质量**：统一抽象层，减少缓存使用错误
- **便于性能优化**：集中管理便于后续优化缓存策略、批量操作等

### 实际使用场景示例

在订单查询场景中，当前代码需要大量使用 GORM 的 Preload 来加载关联对象：

```go
// 当前实现：大量 Preload，代码冗长且重复
saleBill, err := repo.GetSaleBill(
    CommonRepo.WhereByUuid(saleBillUuid),
    CommonRepo.Preload(WithPreload{Query: "SaleBillSetting"}),
    CommonRepo.Preload(WithPreload{Query: "Desk"}),
    CommonRepo.Preload(WithPreload{Query: "SaleOrders.SaleOrderProducts.ProductPackage"}),
    CommonRepo.Preload(WithPreload{Query: "SaleOrders.SaleOrderProducts.MultiLanguageName"}),
    CommonRepo.Preload(WithPreload{Query: "SaleOrders.SaleOrderProducts.ProductPackage.ProductCategory"}),
    // ... 更多 Preload
)
```

**问题**：
- 每次查询都要从数据库加载这些关联对象，即使它们很少变更
- 代码重复，每个查询方法都需要写类似的 Preload
- 难以统一管理缓存策略

**使用对象存储层后的改进**：

```go
// 改进后：通过对象存储层获取，自动缓存
saleBill, err := repo.GetSaleBill(CommonRepo.WhereByUuid(saleBillUuid))
if err != nil {
    return err
}

// 从 context 获取 company UUID（多租户隔离）
companyUuid := ctx.GetCompanyUuid()

// 通过对象存储层获取关联对象，自动处理缓存
// Key 格式：{company_uuid}:{object_type}:{object_uuid}

// 1. 获取 SaleBillSetting（一对一关系）
if saleBillSetting, err := objectStorage.Get(ctx, 
    fmt.Sprintf("%d:sale_bill_setting:%d", companyUuid, saleBill.Uuid), 
    func() (*model.SaleBillSetting, error) {
        return repo.GetSaleBillSetting(saleBill.Uuid)
    }); err == nil {
    saleBill.SaleBillSetting = saleBillSetting // 直接赋值注入
}

// 2. 获取 Desk（一对一关系）
if saleBill.DeskUuid != 0 {
    if desk, err := objectStorage.Get(ctx, 
        fmt.Sprintf("%d:desk:%d", companyUuid, saleBill.DeskUuid),
        func() (*model.Desk, error) {
            return repo.GetDesk(saleBill.DeskUuid)
        }); err == nil {
        saleBill.Desk = desk // 直接赋值注入
    }
}

// 3. 批量获取商品包（ProductPackage），用于后续注入到 SaleOrderProducts
productPackageUuids := extractProductPackageUuids(saleBill)
if len(productPackageUuids) > 0 {
    keys := make([]string, len(productPackageUuids))
    for i, uuid := range productPackageUuids {
        keys[i] = fmt.Sprintf("%d:product_package:%d", companyUuid, uuid)
    }
    productPackages, _ := objectStorage.BatchGet(ctx, keys,
        func(keys []string) (map[string]*model.ProductPackage, error) {
            // 从 keys 中提取 UUID，调用 Repository
            uuids := extractUuidsFromKeys(keys)
            return repo.GetProductPackagesByUuids(uuids)
        })
    
    // 4. 将 ProductPackage 注入到 SaleOrderProducts 中
    // 遍历 SaleOrders，为每个 SaleOrder 的 SaleOrderProducts 设置 ProductPackage
    for _, saleOrder := range saleBill.SaleOrders {
        for _, saleOrderProduct := range saleOrder.SaleOrderProducts {
            if saleOrderProduct.ProductPackageUuid != 0 {
                key := fmt.Sprintf("%d:product_package:%d", companyUuid, saleOrderProduct.ProductPackageUuid)
                if productPackage, exists := productPackages[key]; exists {
                    saleOrderProduct.ProductPackage = productPackage // 注入 ProductPackage
                }
            }
        }
    }
}

// 5. 批量获取其他关联对象（MultiLanguageName、ProductCategory 等）
// 类似的方式，获取后注入到对应的对象中
// 例如：为 SaleOrderProducts 注入 MultiLanguageName
for _, saleOrder := range saleBill.SaleOrders {
    for _, saleOrderProduct := range saleOrder.SaleOrderProducts {
        if saleOrderProduct.MultiLanguageNameUuid != 0 {
            if multiLangName, err := objectStorage.Get(ctx,
                fmt.Sprintf("%d:multi_language_name:%d", companyUuid, saleOrderProduct.MultiLanguageNameUuid),
                func() (*model.MultiLanguageName, error) {
                    return repo.GetMultiLanguageName(saleOrderProduct.MultiLanguageNameUuid)
                }); err == nil {
                saleOrderProduct.MultiLanguageName = multiLangName // 注入 MultiLanguageName
            }
        }
    }
}

// 现在 saleBill 对象已经包含了所有通过对象存储层获取的关联对象
return saleBill, nil
```

**对象注入说明**：

### 推荐方式：配置映射自动注入

对象存储层提供配置映射方式的自动注入机制，通过配置关联关系自动识别并注入对象。这种方式结合了灵活性和性能优化：

```go
// 推荐方式：使用配置映射自动注入
saleBill, err := repo.GetSaleBill(CommonRepo.WhereByUuid(saleBillUuid))
if err != nil {
    return err
}

// 定义关联配置
associations := []objectStorage.Association{
    // 一对一关系：SaleBillSetting
    {
        Path: "SaleBillSetting",
        ObjectType: "sale_bill_setting",
        GetUUID: func(obj interface{}) uint64 {
            return obj.(*model.SaleBill).Uuid
        },
        QueryFunc: func(ctx context.Context, uuid uint64) (interface{}, error) {
            return repo.GetSaleBillSetting(uuid)
        },
    },
    // 一对一关系：Desk
    {
        Path: "Desk",
        ObjectType: "desk",
        GetUUID: func(obj interface{}) uint64 {
            return obj.(*model.SaleBill).DeskUuid
        },
        QueryFunc: func(ctx context.Context, uuid uint64) (interface{}, error) {
            return repo.GetDesk(uuid)
        },
    },
    // 嵌套关联：SaleOrders.SaleOrderProducts.ProductPackage（支持批量优化）
    {
        Path: "SaleOrders.SaleOrderProducts.ProductPackage",
        ObjectType: "product_package",
        GetUUID: func(obj interface{}) uint64 {
            return obj.(*model.SaleOrderProduct).ProductPackageUuid
        },
        QueryFunc: func(ctx context.Context, uuid uint64) (interface{}, error) {
            return repo.GetProductPackage(uuid)
        },
        BatchQueryFunc: func(ctx context.Context, uuids []uint64) (map[uint64]interface{}, error) {
            // 批量查询优化
            packages, err := repo.GetProductPackagesByUuids(uuids)
            if err != nil {
                return nil, err
            }
            result := make(map[uint64]interface{})
            for _, pkg := range packages {
                result[pkg.Uuid] = pkg
            }
            return result, nil
        },
    },
    // 嵌套关联：SaleOrders.SaleOrderProducts.MultiLanguageName（支持批量优化）
    {
        Path: "SaleOrders.SaleOrderProducts.MultiLanguageName",
        ObjectType: "multi_language_name",
        GetUUID: func(obj interface{}) uint64 {
            return obj.(*model.SaleOrderProduct).MultiLanguageNameUuid
        },
        QueryFunc: func(ctx context.Context, uuid uint64) (interface{}, error) {
            return repo.GetMultiLanguageName(uuid)
        },
        BatchQueryFunc: func(ctx context.Context, uuids []uint64) (map[uint64]interface{}, error) {
            names, err := repo.GetMultiLanguageNamesByUuids(uuids)
            if err != nil {
                return nil, err
            }
            result := make(map[uint64]interface{})
            for _, name := range names {
                result[name.Uuid] = name
            }
            return result, nil
        },
    },
    // 嵌套关联：SaleOrders.SaleOrderProducts.ProductPackage.ProductCategory
    {
        Path: "SaleOrders.SaleOrderProducts.ProductPackage.ProductCategory",
        ObjectType: "product_category",
        GetUUID: func(obj interface{}) uint64 {
            return obj.(*model.ProductPackage).CategoryUuid
        },
        QueryFunc: func(ctx context.Context, uuid uint64) (interface{}, error) {
            return repo.GetProductCategory(uuid)
        },
    },
}

// 执行自动注入
err = objectStorage.PreloadWithConfig(ctx, saleBill, associations)
if err != nil {
    return nil, err
}

// 现在 saleBill 对象已经包含了所有通过对象存储层获取的关联对象
return saleBill, nil
```

**配置映射方式的优势**：
1. **灵活性**：支持自定义查询函数，适应不同的查询需求
2. **性能优化**：通过 `BatchQueryFunc` 支持批量查询，自动收集同一层级的 UUID 批量查询
3. **类型安全**：通过 `GetUUID` 和 `QueryFunc` 明确类型，避免反射带来的性能损失
4. **可维护性**：配置集中管理，易于理解和修改
5. **支持嵌套**：支持 "Parent.Child.GrandChild" 格式的嵌套路径

**工作原理**：
1. **解析路径**：将 "SaleOrders.SaleOrderProducts.ProductPackage" 拆分为路径数组
2. **反射查找字段**：通过反射找到每个路径对应的结构体字段
3. **提取 UUID**：调用 `GetUUID` 函数从对象中提取 UUID
4. **批量查询**：如果配置了 `BatchQueryFunc`，自动收集同一层级的 UUID 批量查询
5. **递归注入**：对每个路径层级递归执行上述过程，最终注入到目标字段

**注意事项**：
- 确保在注入前已经查询到主对象（如 `saleBill.SaleOrders`），否则无法注入嵌套关联
- 如果主对象是通过 Preload 加载的，可以直接注入；如果是单独查询的，需要先查询主对象再注入关联
- 建议按需注入，只注入业务逻辑需要的关联对象，避免不必要的查询和内存占用
- 对于频繁访问的关联对象，建议配置 `BatchQueryFunc` 以提升性能

**适用对象类型**（适合放入对象存储层的对象）：
- `SaleBillSetting` - 销售账单设置（按 SaleBillUuid）
- `Desk` - 桌台信息（按 DeskUuid）
- `ProductPackage` - 商品包（按 ProductPackageUuid）
- `MultiLanguageName` - 多语言名称（按 MultiLanguageNameUuid）
- `ProductCategory` - 商品分类（按 CategoryUuid）
- `ProductBoms` - 商品清单（按 ProductBomUuid）
- `ProductPackageAttributeGroups` - 商品包属性组（按 Uuid）
- `ProductAttribute` - 商品属性（按 ProductAttributeUuid）
- `ProductFlavor` - 商品口味（按 ProductFlavorUuid）
- `ProductSauce` - 商品加料（按 ProductSauceUuid）
- `BatchTag` - 批次标签（按 BatchTagUuid）

**这些对象的共同特点**：
- 变更频率低（配置类、基础数据类）
- 访问频率高（订单查询中频繁使用）
- 可以通过 UUID 或 ID 唯一标识
- 适合缓存，减少数据库查询压力

**多租户设计约束**：
- TTPOS 为多个 company 提供服务，每个 company 有独立的数据空间
- Key 设计必须包含 company UUID，格式：`{company_uuid}:{object_type}:{object_uuid}`
- 确保不同 company 的数据完全隔离，避免跨租户数据泄露
- 支持按 company 粒度管理缓存（批量失效、批量更新等），便于维护和排查问题

## 🎯 产品对齐

对象存储层作为基础设施层，不直接面向业务用户，而是为后端开发工程师提供统一的开发工具。该功能支持产品愿景中的"提升开发效率"和"降低维护成本"目标，通过统一抽象层减少开发人员的工作量，提高代码质量和系统可维护性。

## 📝 用户故事

**作为** 后端开发工程师  
**我想** 通过统一的接口获取模型对象，无需关心缓存细节  
**以便于** 提高开发效率，减少重复代码，统一管理对象生命周期

---

## 功能需求

### Requirement 1: 统一对象获取接口

**用户故事**: 作为后端开发工程师，我想通过统一的接口获取模型对象，无需关心缓存细节，以便于提高开发效率

#### 验收标准

1. **WHEN** 调用 `Get(key, queryFunc)` **THEN** 系统 **SHALL** 自动从三级缓存查找，未命中时调用查询函数并写入缓存
2. **IF** 缓存命中 **THEN** 系统 **SHALL** 直接返回缓存对象，不调用查询函数
3. **WHEN** 调用 `BatchGet(keys, queryFunc)` **THEN** 系统 **SHALL** 批量查询缓存，减少网络开销
4. **IF** 批量获取时部分缓存命中 **THEN** 系统 **SHALL** 只对未命中的 key 调用查询函数

#### 具体要求

- [ ] 1.1 提供泛型接口 `Get[T any](ctx context.Context, key string, query func() (T, error)) (T, error)`，支持类型安全
- [ ] 1.2 提供批量获取接口 `BatchGet[T any](ctx context.Context, keys []string, query func([]string) (map[string]T, error)) (map[string]T, error)`
- [ ] 1.3 自动处理三级缓存查询（本地缓存 → Redis → 数据库）
- [ ] 1.4 缓存未命中时自动调用查询函数并写入缓存
- [ ] 1.5 支持 context.Context，便于链路追踪和超时控制
- [ ] 1.6 **Key 设计必须包含 company UUID**，格式：`{company_uuid}:{object_type}:{object_uuid}`（如 `1724054084:product_package:123456`）
- [ ] 1.7 支持从 context 自动提取 company UUID，简化 key 构建（提供辅助方法）
- [ ] 1.8 批量获取时自动去重，避免重复查询
- [ ] 1.9 确保不同 company 的数据完全隔离，避免跨租户数据泄露

---

### Requirement 1.5: 配置映射自动关联注入（推荐）

**用户故事**: 作为后端开发工程师，我想通过配置映射的方式自动注入关联对象，以便于简化代码，提高开发效率，同时保持灵活性和性能优化能力

#### 验收标准

1. **WHEN** 调用 `PreloadWithConfig(ctx, obj, associations)` **THEN** 系统 **SHALL** 根据配置自动识别关联字段并注入对象
2. **IF** 配置了 `BatchQueryFunc` **THEN** 系统 **SHALL** 自动收集同一层级的 UUID 批量查询，减少查询次数
3. **WHEN** 配置了嵌套路径（如 "SaleOrders.SaleOrderProducts.ProductPackage"） **THEN** 系统 **SHALL** 递归处理嵌套关联
4. **IF** 查询函数返回错误 **THEN** 系统 **SHALL** 优雅处理错误，不影响其他关联的注入

#### 具体要求

- [ ] 1.5.1 提供 `PreloadWithConfig(ctx, obj, associations)` 方法，支持配置映射方式的自动注入
- [ ] 1.5.2 支持 `Association` 配置结构，包含：
  - `Path`: 关联路径（如 "SaleBillSetting"、"SaleOrders.SaleOrderProducts.ProductPackage"）
  - `ObjectType`: 对象类型（用于构建缓存 key）
  - `GetUUID`: 从对象中提取 UUID 的函数
  - `QueryFunc`: 单个对象查询函数
  - `BatchQueryFunc`: 批量查询函数（可选，用于性能优化）
- [ ] 1.5.3 支持嵌套路径解析，自动处理 "Parent.Child.GrandChild" 格式
- [ ] 1.5.4 自动批量优化：收集同一层级的 UUID，调用 `BatchQueryFunc` 批量查询
- [ ] 1.5.5 支持一对一、一对多、多对多关系的自动注入
- [ ] 1.5.6 使用反射设置关联字段，支持指针和值类型
- [ ] 1.5.7 错误处理：单个关联查询失败不影响其他关联的注入

---

### Requirement 2: 对象生命周期管理

**用户故事**: 作为后端开发工程师，我想统一管理对象的生命周期（过期、失效、更新），以便于降低维护成本

#### 验收标准

1. **WHEN** 调用 `Invalidate(key)` **THEN** 系统 **SHALL** 使指定 key 的缓存失效
2. **WHEN** 调用 `Update(key, value)` **THEN** 系统 **SHALL** 更新缓存中的对象
3. **WHEN** 配置了 TTL **THEN** 系统 **SHALL** 在指定时间后自动使缓存过期
4. **WHEN** 调用 `Warmup(keys, queryFunc)` **THEN** 系统 **SHALL** 预热指定 keys 的缓存

#### 具体要求

- [ ] 2.1 提供缓存失效接口 `Invalidate(ctx context.Context, key string) error`
- [ ] 2.2 提供缓存更新接口 `Update[T any](ctx context.Context, key string, value T) error`
- [ ] 2.3 支持配置 TTL（Time To Live），自动管理缓存过期
- [ ] 2.4 提供缓存预热接口 `Warmup[T any](ctx context.Context, keys []string, query func([]string) (map[string]T, error)) error`
- [ ] 2.5 支持批量失效和批量更新操作
- [ ] 2.6 **支持按 company 粒度批量失效**：`InvalidateByCompany(ctx context.Context, companyUuid uint64) error`
- [ ] 2.7 **支持按 company 粒度批量更新**：`UpdateByCompany[T any](ctx context.Context, companyUuid uint64, objectType string, values map[string]T) error`
- [ ] 2.8 **支持按 company + object_type 粒度批量失效**：`InvalidateByCompanyAndType(ctx context.Context, companyUuid uint64, objectType string) error`


---

### Requirement 3: 灵活的配置策略

**用户故事**: 作为后端开发工程师，我想为不同模型对象配置不同的缓存策略，以便于灵活应对不同场景

#### 验收标准

1. **WHEN** 配置不同的 TTL **THEN** 系统 **SHALL** 为不同对象应用不同的过期时间
2. **WHEN** 配置禁用缓存 **THEN** 系统 **SHALL** 跳过缓存，直接调用查询函数
3. **WHEN** 配置 Key 前缀 **THEN** 系统 **SHALL** 自动为所有 key 添加前缀
4. **WHEN** 不同 company 配置不同策略 **THEN** 系统 **SHALL** 支持按 company 粒度配置

#### 具体要求

- [ ] 3.1 支持为不同模型对象配置不同的缓存策略（TTL、Key 前缀等）
- [ ] 3.2 支持动态调整缓存过期时间
- [ ] 3.3 支持禁用缓存（用于调试或特殊场景）
- [ ] 3.4 **Key 前缀自动包含 company UUID**，确保多租户隔离
- [ ] 3.5 提供配置构建器，便于链式配置
- [ ] 3.6 **支持按 company 粒度配置缓存策略**（如不同 company 可以有不同的 TTL）

---

### Requirement 4: 扩展能力（可选）

**用户故事**: 作为后端开发工程师，我想监听缓存事件和统计缓存性能，以便于优化缓存策略

#### 验收标准

1. **WHEN** 配置了缓存事件监听器 **THEN** 系统 **SHALL** 在缓存命中、未命中、失效时触发事件
2. **WHEN** 调用缓存统计接口 **THEN** 系统 **SHALL** 返回命中率、访问频率等统计信息

#### 具体要求

- [ ] 4.1 支持缓存事件监听（命中、未命中、失效等）
- [ ] 4.2 支持缓存统计（命中率、访问频率等）
- [ ] 4.3 预留批量操作、分布式锁等扩展接口

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 Controller → Service → Repository 分层
- **单一职责原则**: 对象存储层作为独立的基础设施包，只负责对象存储和缓存管理
- **模块化设计**: 对象存储层应独立且可复用，不依赖具体业务逻辑
- **依赖管理**: 对象存储层依赖三级缓存基础包，不直接依赖数据库或业务 Service
- **遵循规范**:
  - `.cursor/rules/go-main.mdc` - Go Main 开发规范
  - `.cursor/rules/structs.mdc` - 项目结构规范

### API 设计要求

- [ ] 不涉及对外 API 接口（基础设施层）

### 数据库设计要求

- [ ] 不涉及数据库表结构变更（使用现有数据库查询）

### 性能要求

- [ ] 缓存命中时响应时间 < 10ms
- [ ] 缓存未命中时响应时间 < 200ms（包含数据库查询）
- [ ] 批量操作时使用批量查询，减少网络开销
- [ ] 并发安全：确保并发访问的安全性

### 测试要求

- [ ] Service 层测试覆盖率 ≥ 70%
- [ ] Repository 层测试覆盖率 ≥ 80%
- [ ] 单元测试覆盖所有核心接口（Get、BatchGet、Invalidate、Update、Warmup）
- [ ] 集成测试覆盖三级缓存查询流程
- [ ] 性能测试验证缓存命中率和响应时间
- [ ] 参考: `.cursor/rules/go-main.mdc` - 测试规范

### 安全要求

- [ ] 支持 context.Context，便于链路追踪和超时控制
- [ ] 错误处理：统一的错误处理策略，不泄露敏感信息
- [ ] 并发安全：确保并发访问的安全性
- [ ] **多租户隔离**：确保不同 company 的数据完全隔离，Key 必须包含 company UUID
- [ ] **数据泄露防护**：防止因 key 构建错误导致的跨租户数据泄露
- [ ] 参考: `.cursor/rules/security.mdc` - 安全开发规范

### 可靠性要求

- [ ] 缓存失败时优雅降级，直接调用查询函数
- [ ] 错误日志记录（使用 Logger）
- [ ] 支持超时控制（通过 context.Context）

---

## 验收标准

### 功能验收

1. **统一对象获取**: 调用 `Get` 方法能够自动处理三级缓存查询和数据库回填
2. **批量获取**: 调用 `BatchGet` 方法能够批量查询缓存，减少网络开销
3. **生命周期管理**: 调用 `Invalidate`、`Update`、`Warmup` 方法能够正确管理缓存生命周期
4. **配置策略**: 能够为不同模型对象配置不同的缓存策略
5. **多租户隔离**: Key 必须包含 company UUID，不同 company 的数据完全隔离
6. **按 company 粒度管理**: 能够按 company 粒度批量失效、批量更新缓存

### 测试验收

1. **单元测试**: 覆盖率达标，所有核心接口测试通过
2. **集成测试**: 三级缓存查询流程测试通过
3. **性能测试**: 缓存命中率和响应时间符合要求
4. **并发测试**: 并发访问安全性测试通过
5. **多租户隔离测试**: 确保不同 company 的数据完全隔离，不会出现跨租户数据泄露
6. **按 company 粒度管理测试**: 验证按 company 粒度批量失效、批量更新功能正常

### 文档验收

1. **技术文档**: design.md 完整且准确（待创建）
2. **API 文档**: 接口文档完整（如有）
3. **使用示例**: 提供使用示例和最佳实践
4. **测试文档**: tasks.md 中的测试任务完成（待创建）

---

## 约束条件

### 技术约束

#### Go Main 模块

- 必须使用 Gin 框架（上下文传递）
- 接口以 `I` 开头，实现以 `Impl` 结尾
- 使用 Go 1.18+ 泛型，提供类型安全
- 不使用 panic，返回 error
- 依赖三级缓存基础包（假定已实现）

### 业务约束

- 对象存储层作为基础设施层，不包含业务逻辑
- 需要与现有的三级缓存基础包集成
- 需要重构部分现有代码以使用新层（可选，可分阶段）
- **多租户隔离**：TTPOS 为多个 company 提供服务，必须确保：
  - Key 设计必须包含 company UUID，格式：`{company_uuid}:{object_type}:{object_uuid}`
  - 不同 company 的数据完全隔离，避免跨租户数据泄露
  - 支持按 company 粒度管理缓存（批量失效、批量更新等）
  - 维护时需要能够基于 company 粒度去管理对象存储模型

### 资源约束

- 开发时间: 5-7 天
- Story Point: 5 SP（待技术评审确认）

---

## 依赖关系

### 技术依赖

- `三级缓存基础包` - 提供三级缓存查询能力（本地缓存 → Redis → 数据库）
- `context` - Go 标准库，用于上下文传递和超时控制
- `main/pkg/cache/` - 现有缓存实现（参考）

### 服务依赖

- **Main → 三级缓存基础包**: 依赖注入

### 业务依赖

- 无业务依赖（基础设施层）

---

## 风险和缓解

### 风险 1: 三级缓存基础包接口变更

**影响**: 高  
**概率**: 中  
**缓解措施**:

- 确保三级缓存基础包接口稳定后再实现对象存储层
- 定义清晰的接口契约，减少接口变更影响

### 风险 2: 性能影响

**影响**: 中  
**概率**: 低  
**缓解措施**:

- 实现后进行性能基准测试，确保开销可接受
- 优化批量操作，减少网络开销
- 使用泛型避免类型转换开销

### 风险 3: 迁移成本

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 采用渐进式迁移策略，新功能使用新层，旧代码逐步迁移
- 提供迁移指南和示例代码
- 保持向后兼容，不强制立即迁移

### 风险 4: 多租户数据隔离

**影响**: 高  
**概率**: 低  
**缓解措施**:

- Key 设计强制包含 company UUID，确保不同租户数据隔离
- 提供辅助方法自动从 context 提取 company UUID，避免手动构建 key 时遗漏
- 单元测试覆盖多租户场景，确保不会出现跨租户数据泄露
- 提供按 company 粒度的管理接口，便于维护和排查问题

---

## 时间表

- **Phase 1 - 接口设计和抽象层实现**: 2 天
- **Phase 2 - 生命周期管理实现**: 2 天
- **Phase 3 - 集成测试和文档**: 1-2 天
- **Phase 4 - 重构现有代码（可选）**: 1-2 天
- **总计**: 5-7 天（SP = 5，待技术评审确认）

---

## 参考资料

### 核心规范

- `.cursor/rules/go-main.mdc` - Go Main 核心约束
- `.cursor/rules/structs.mdc` - 项目结构规范
- `.cursor/rules/security.mdc` - 安全开发规范

### 架构文档

- `docs/human/architecture/go-main-architecture.md` - Go Main 架构
- `main/pkg/cache/` - 现有缓存实现（参考）

### 开发指南

- `docs/human/guides/go-main-development.md` - Go Main 开发指南

### 外部参考

- Spring Cache - Java 缓存抽象框架
- Laravel Cache - PHP 缓存抽象框架

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-12-24  
**作者**: xiezhihuan  
**审核者**: {审核者}

