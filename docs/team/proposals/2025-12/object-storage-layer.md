# 对象存储层 需求提案

> 本文档用于需求发起阶段，经团队评审后创建正式 Spec。

---

## 📋 提案信息

| 项目       | 内容     |
| ---------- | -------- |
| **提案人** | xiezhihuan   |
| **日期**   | 2025-12-24   |
| **目标版本** | - |
| **状态**   | 进行中   |
| **关联任务** | - |
| **关联 Spec** | [story-main-object-storage-layer](../../../shared/specs/active/story-main-object-storage-layer/)      |

---

## 🎯 背景和动机

### 问题描述

当前项目中存在大量通过 key 获取模型对象的场景，这些场景通常需要：
1. 先从缓存中查找
2. 缓存未命中时从数据库查询
3. 查询结果写入缓存
4. 管理对象的生命周期（过期、失效、更新）

目前这些逻辑分散在各个 Service 和 Repository 层，存在以下问题：
- **代码重复**：每个模块都需要实现类似的缓存查询逻辑
- **生命周期管理混乱**：缓存失效、更新策略不统一
- **缺乏统一抽象**：没有统一的对象存储接口，难以统一管理和优化

**示例场景**：
> 在订单服务中，需要频繁获取订单对象。当前实现中，每个方法都需要手动处理缓存逻辑：
> ```go
> // 当前实现：代码重复
> cacheKey := fmt.Sprintf("order:%d", orderID)
> if cached, exists := cache.Get(cacheKey); exists {
>     return cached.(*Order), nil
> }
> order, err := repo.GetOrder(orderID)
> cache.Set(cacheKey, order, ttl)
> return order, err
> ```

### 业务价值

- **提升开发效率**：减少重复代码，统一缓存访问模式
- **降低维护成本**：集中管理对象生命周期，便于统一优化和调试
- **提高代码质量**：统一抽象层，减少缓存使用错误
- **便于性能优化**：集中管理便于后续优化缓存策略、批量操作等

### 目标用户

- [ ] 收银员
- [ ] 商户管理员
- [ ] 厨房人员
- [ ] 顾客
- [x] **后端开发工程师**：使用对象存储层简化开发
- [ ] 其他: ________

---

## 💡 解决方案概述

### 方案描述

设计一个**对象存储层（Object Storage Layer）**，提供统一的接口通过 key 获取模型对象。该层基于三级缓存基础包，负责维护对象的完整生命周期。

**核心设计**：
> 对象存储层提供 `Get(key, queryFunc)` 方法，自动处理：
> 1. 从三级缓存（本地缓存 → Redis → 数据库）逐级查找
> 2. 缓存未命中时调用查询函数获取数据
> 3. 自动写入缓存
> 4. 统一管理对象的生命周期（过期、失效、更新）

**多租户设计约束**：
> TTPOS 为多个 company 提供服务，每个 company 有独立的数据空间。对象存储层必须：
> 1. Key 设计必须包含 company UUID，格式：`{company_uuid}:{object_type}:{object_uuid}`
> 2. 确保不同 company 的数据完全隔离，避免跨租户数据泄露
> 3. 支持按 company 粒度管理缓存（批量失效、批量更新等），便于维护和排查问题

**三级缓存基础包接口**（假定已实现）：
```go
// 三级缓存基础包提供的接口
type CacheLayer interface {
    // GET 方法：从缓存获取，未命中时调用 query 函数查询并写入缓存
    GET(key string, query func() (any, error)) (any, error)
}
```

### 核心功能点

1. **统一对象获取接口**
   - 提供 `Get(key, queryFunc)` 方法，自动处理缓存查询和数据库回填
   - 支持泛型，返回类型安全的模型对象
   - 支持批量获取（BatchGet）

2. **对象生命周期管理**
   - 自动设置缓存过期时间（TTL）
   - 提供缓存失效接口（Invalidate）
   - 支持缓存更新接口（Update）
   - 支持缓存预热（Warmup）

3. **灵活的配置策略**
   - 支持不同模型对象配置不同的缓存策略
   - 支持动态调整缓存过期时间
   - 支持禁用缓存（用于调试或特殊场景）

4. **扩展能力**
   - 支持缓存事件监听（命中、未命中、失效等）
   - 支持缓存统计（命中率、访问频率等）
   - 预留批量操作、分布式锁等扩展接口

5. **多租户支持**
   - Key 设计必须包含 company UUID，确保数据隔离
   - 支持从 context 自动提取 company UUID，简化 key 构建
   - 支持按 company 粒度批量失效、批量更新缓存
   - 支持按 company + object_type 粒度批量失效

### 影响范围

**涉及终端**：
- [ ] POS 收银端
- [ ] Shop 商家管理端
- [ ] KDS 厨显端
- [ ] QDS 排号叫号端
- [ ] Assistant 助手端
- [ ] Tablet 平板端
- [ ] Mobile 扫码端
- [ ] Menu 电子菜单端
- [ ] Member 会员端

**涉及模块**：
- [ ] UI 组件
- [ ] API 接口
- [ ] 数据模型
- [x] **业务逻辑**：Service 和 Repository 层将使用对象存储层
- [ ] 第三方集成
- [x] **基础设施**：新增对象存储层包（`main/pkg/storage/` 或类似）

---

## 📊 初步评估

### 技术复杂度

- [ ] **低**：纯 UI 调整，无业务逻辑变更
- [x] **中**：需要前后端联调，基础业务逻辑
- [ ] **高**：涉及架构调整、第三方集成、复杂算法

**说明**：
- 需要设计统一的接口抽象
- 需要与现有的三级缓存基础包集成
- 需要重构部分现有代码以使用新层
- 不涉及复杂的分布式算法或第三方集成

### 工作量预估

[粗略估算，具体 SP 在技术方案评审时确定]

- **预计天数**: 5-7 天
- **预估 SP**: 5 SP（待技术评审确认）

**分解**：
- 接口设计和抽象层实现：2 天
- 生命周期管理实现：2 天
- 集成测试和文档：1-2 天
- 重构现有代码（可选，可分阶段）：1-2 天

### 风险识别

**潜在风险**：
1. **三级缓存基础包接口变更**：如果基础包接口未稳定，可能导致对象存储层需要调整
2. **性能影响**：新增抽象层可能带来轻微性能开销，需要评估
3. **迁移成本**：现有代码迁移到新层需要时间，可能影响开发进度
4. **多租户数据隔离**：如果 Key 设计不当，可能导致跨租户数据泄露

**缓解措施**：
1. **接口稳定性**：确保三级缓存基础包接口稳定后再实现对象存储层
2. **性能测试**：实现后进行性能基准测试，确保开销可接受
3. **渐进式迁移**：采用渐进式迁移策略，新功能使用新层，旧代码逐步迁移
4. **多租户隔离**：
   - Key 设计强制包含 company UUID，确保不同租户数据隔离
   - 提供辅助方法自动从 context 提取 company UUID，避免手动构建 key 时遗漏
   - 单元测试覆盖多租户场景，确保不会出现跨租户数据泄露

---

## 🔗 相关资源

### 参考需求

- 类似功能: 项目中的 `main/pkg/cache/` 包（现有缓存实现）
- 竞品分析: 参考其他框架的对象存储抽象（如 Spring Cache、Laravel Cache）

### 相关文档

- 产品需求文档 (PRD): -
- 用户调研报告: -
- 技术预研文档: 需要补充三级缓存基础包的设计文档

---

## 🤝 需求评审

### 评审参与人

| 角色         | 姓名   | 签名/日期 |
| ------------ | ------ | --------- |
| 产品经理     | {姓名} |           |
| 技术负责人   | {姓名} |           |
| 开发代表     | {姓名} |           |
| 测试代表     | {姓名} |           |
| UI/UX 设计师 | {姓名} |           |

### 评审结论

- [ ] ✅ **批准**：进入技术方案设计阶段
- [ ] 🔄 **修改后批准**：需补充以下内容
- [ ] ❌ **拒绝**：不符合产品规划或优先级

**评审意见**：

```
[记录评审会议的关键讨论和决策]
```

**下一步行动**：

- [x] 创建 Spec：`story-main-object-storage-layer`
- [ ] 分配负责人：{姓名}
- [ ] 目标 Sprint：Sprint {N}

---

## 📝 附录

### User Story（初稿）

**作为** 后端开发工程师  
**我想** 通过统一的接口获取模型对象，无需关心缓存细节  
**以便于** 提高开发效率，减少重复代码，统一管理对象生命周期

### AC 验收标准（初稿）

1. **WHEN** 调用 `Get(key, queryFunc)` **THEN** 系统 **SHALL** 自动从三级缓存查找，未命中时调用查询函数并写入缓存
2. **IF** 缓存命中 **THEN** 系统 **SHALL** 直接返回缓存对象，不调用查询函数
3. **WHEN** 调用 `Invalidate(key)` **THEN** 系统 **SHALL** 使指定 key 的缓存失效
4. **WHEN** 调用 `Update(key, value)` **THEN** 系统 **SHALL** 更新缓存中的对象
5. **WHEN** 配置了 TTL **THEN** 系统 **SHALL** 在指定时间后自动使缓存过期
6. **IF** 批量获取对象 **THEN** 系统 **SHALL** 批量查询缓存，减少网络开销
7. **WHEN** Key 设计 **THEN** 系统 **SHALL** 必须包含 company UUID，格式：`{company_uuid}:{object_type}:{object_uuid}`
8. **IF** 不同 company 的数据 **THEN** 系统 **SHALL** 完全隔离，避免跨租户数据泄露
9. **WHEN** 调用 `InvalidateByCompany(companyUuid)` **THEN** 系统 **SHALL** 使指定 company 的所有缓存失效
10. **WHEN** 调用 `InvalidateByCompanyAndType(companyUuid, objectType)` **THEN** 系统 **SHALL** 使指定 company 和 object_type 的缓存失效
11. **WHEN** 调用 `PreloadWithConfig(ctx, obj, associations)` **THEN** 系统 **SHALL** 根据配置自动识别关联字段并注入对象
12. **IF** 配置了 `BatchQueryFunc` **THEN** 系统 **SHALL** 自动收集同一层级的 UUID 批量查询，减少查询次数
13. **WHEN** 配置了嵌套路径（如 "SaleOrders.SaleOrderProducts.ProductPackage"） **THEN** 系统 **SHALL** 递归处理嵌套关联

### 技术设计要点（初稿）

#### 接口设计

```go
// ObjectStorage 对象存储接口
type ObjectStorage[T any] interface {
    // Get 获取对象，自动处理缓存查询和回填
    // key 格式：{company_uuid}:{object_type}:{object_uuid}
    Get(ctx context.Context, key string, query func() (T, error)) (T, error)
    
    // BatchGet 批量获取对象
    // keys 格式：{company_uuid}:{object_type}:{object_uuid}
    BatchGet(ctx context.Context, keys []string, query func([]string) (map[string]T, error)) (map[string]T, error)
    
    // Invalidate 使缓存失效
    Invalidate(ctx context.Context, key string) error
    
    // Update 更新缓存
    Update(ctx context.Context, key string, value T) error
    
    // Warmup 预热缓存
    Warmup(ctx context.Context, keys []string, query func([]string) (map[string]T, error)) error
    
    // InvalidateByCompany 按 company 粒度批量失效缓存
    InvalidateByCompany(ctx context.Context, companyUuid uint64) error
    
    // InvalidateByCompanyAndType 按 company + object_type 粒度批量失效缓存
    InvalidateByCompanyAndType(ctx context.Context, companyUuid uint64, objectType string) error
    
    // UpdateByCompany 按 company 粒度批量更新缓存
    UpdateByCompany(ctx context.Context, companyUuid uint64, objectType string, values map[string]T) error
    
    // PreloadWithConfig 配置映射自动关联注入（推荐方式）
    PreloadWithConfig(ctx context.Context, obj interface{}, associations []Association) error
}

// Association 关联配置
type Association struct {
    // Path 关联路径，支持嵌套，如 "SaleBillSetting"、"SaleOrders.SaleOrderProducts.ProductPackage"
    Path string
    
    // ObjectType 对象类型，用于构建缓存 key
    ObjectType string
    
    // GetUUID 从对象中提取 UUID 的函数
    GetUUID func(obj interface{}) uint64
    
    // QueryFunc 单个对象查询函数
    QueryFunc func(ctx context.Context, uuid uint64) (interface{}, error)
    
    // BatchQueryFunc 批量查询函数（可选，用于性能优化）
    // 返回 map[uint64]interface{}，key 为 UUID，value 为对象
    BatchQueryFunc func(ctx context.Context, uuids []uint64) (map[uint64]interface{}, error)
}

// Config 配置选项
type Config struct {
    TTL           time.Duration  // 缓存过期时间
    DisableCache  bool           // 是否禁用缓存（用于调试）
    KeyPrefix     string         // Key 前缀（自动包含 company UUID）
    // ... 其他配置
}

// BuildKey 构建 key 的辅助方法（自动从 context 提取 company UUID）
func BuildKey(ctx context.Context, objectType string, objectUuid uint64) string {
    companyUuid := ctx.GetCompanyUuid()
    return fmt.Sprintf("%d:%s:%d", companyUuid, objectType, objectUuid)
}
```

#### 配置映射自动注入示例

```go
// 推荐方式：使用配置映射自动注入关联对象
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

#### 实现要点

1. **依赖注入**：对象存储层依赖三级缓存基础包
2. **泛型支持**：使用 Go 1.18+ 泛型，提供类型安全
3. **上下文传递**：支持 context.Context，便于链路追踪和超时控制
4. **错误处理**：统一的错误处理策略
5. **并发安全**：确保并发访问的安全性
6. **多租户隔离**：
   - Key 设计强制包含 company UUID，格式：`{company_uuid}:{object_type}:{object_uuid}`
   - 提供 `BuildKey` 辅助方法，自动从 context 提取 company UUID
   - 确保不同 company 的数据完全隔离，避免跨租户数据泄露
7. **按 company 粒度管理**：
   - 支持按 company 粒度批量失效缓存
   - 支持按 company + object_type 粒度批量失效缓存
   - 支持按 company 粒度批量更新缓存
8. **配置映射自动注入**：
   - 支持 `PreloadWithConfig` 方法，通过配置自动注入关联对象
   - 使用反射解析嵌套路径（如 "SaleOrders.SaleOrderProducts.ProductPackage"）
   - 自动批量优化：收集同一层级的 UUID，调用 `BatchQueryFunc` 批量查询
   - 支持一对一、一对多、多对多关系的自动注入
   - 错误处理：单个关联查询失败不影响其他关联的注入

### 线框图/原型（可选）

[技术架构图或代码示例]

---

## 📄 模板使用说明

### 何时使用此模板

- ✅ 产品经理提出新功能想法
- ✅ 用户反馈需求建议
- ✅ 技术团队提出改进方案
- ✅ 需要团队讨论和评审的需求

### 与 Spec 的区别

| 阶段        | 文档类型      | 详细程度 | 用途               |
| ----------- | ------------- | -------- | ------------------ |
| **需求发起** | Proposal      | 粗略     | 团队评审、决策是否做 |
| **需求确认** | Requirements  | 详细     | User Story + AC，开发依据 |
| **技术设计** | Design        | 详细     | 技术方案，实现指导 |
| **任务分解** | Tasks         | 详细     | 开发执行，进度追踪 |

### 流转路径

```
提案 (Proposal) 
  ↓ 评审批准
需求文档 (Requirements) 
  ↓ 技术评审
设计文档 (Design) 
  ↓ SP 评估 ≤ 5
任务分解 (Tasks)
  ↓
开发实现
```

---

**版本**: v1.0.0  
**创建日期**: 2025-12-24  
**维护者**: 产品组 + Scrum Master  
**相关规范**: `.cursor/rules/scrum_story_point.mdc`, `.cursor/rules/specs.mdc`

