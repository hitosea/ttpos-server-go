# 事务使用示例

> Go Main 模块中 Service 层事务使用的正确和错误示例

---

## 规范说明

- ✅ **多个数据库操作必须使用事务包裹**，确保原子性
- ✅ **涉及多个表的数据修改操作必须使用事务**
- ✅ **删除操作涉及关联数据时必须使用事务**
- ✅ **更新操作涉及多个步骤时必须使用事务**
- ✅ **优先使用 `repository.CommonRepo.Transaction()` 方法**
- ✅ **事务中必须使用 `tx` 创建 Repository 实例**
- ✅ **事务中返回 error 会自动回滚，返回 nil 会自动提交**
- ❌ **禁止在事务中手动调用 `tx.Commit()` 或 `tx.Rollback()`**

**何时需要使用事务：**

1. **创建操作**：创建主记录 + 创建关联记录（如创建活动 + 创建规则）
2. **更新操作**：更新主记录 + 更新/删除关联记录（如更新活动 + 删除旧规则 + 创建新规则）
3. **删除操作**：删除主记录 + 删除关联记录（如删除活动 + 删除规则）
4. **复杂业务逻辑**：涉及多个表的数据修改，需要保证一致性

---

## ✅ 正确示例

### 示例 1: 使用 repository.CommonRepo.Transaction

```go
// ✅ 正确：使用 repository.CommonRepo.Transaction
func (s *fullReductionActivitySrv) Create(ctx context.Context, req *req.FullReductionActivityCreateReq) error {
    db := ctx.GetDB()
    
    // 多个数据库操作，必须使用事务
    if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
        // 1. 创建多语言名称（使用 tx 创建 Repository）
        multiLanguageNameRepo := repository.NewMultiLanguageNameRepo(tx)
        multiLanguageName := model.MultiLanguageName{...}
        if err := multiLanguageNameRepo.Create(&multiLanguageName); err != nil {
            return errors.WithMessage(err, "创建多语言名称失败")
        }
        
        // 2. 创建活动（使用 tx 创建 Repository）
        activityRepo := repository.NewFullReductionActivityRepo(tx)
        activity := &model.FullReductionActivity{...}
        if err := activityRepo.Create(activity); err != nil {
            return errors.WithMessage(err, "创建活动失败")
        }
        
        // 3. 创建规则（使用 tx 创建 Repository）
        ruleRepo := repository.NewFullReductionActivityRuleRepo(tx)
        for _, ruleReq := range req.Rules {
            rule := &model.FullReductionActivityRule{...}
            if err := ruleRepo.Create(rule); err != nil {
                return errors.WithMessage(err, "创建规则失败")
            }
        }
        
        // 返回 nil 自动提交，返回 error 自动回滚
        return nil
    }); err != nil {
        return err
    }
    
    return nil
}
```

### 示例 2: 使用 db.Transaction（GORM 标准方法）

```go
// ✅ 正确：也可以使用 db.Transaction（GORM 标准方法）
func (s *orderSrv) CreateOrder(ctx context.Context, req req.CreateOrderReq) error {
    db := ctx.GetDB()
    
    if err := db.Transaction(func(tx *gorm.DB) error {
        // 在事务中执行多个操作
        orderRepo := repository.NewOrderRepo(tx)
        if err := orderRepo.CreateOrder(order); err != nil {
            return errors.WithMessage(err, "创建订单失败")
        }
        
        productRepo := repository.NewProductRepo(tx)
        if err := productRepo.UpdateStock(productUuid, -quantity); err != nil {
            return errors.WithMessage(err, "更新库存失败")
        }
        
        return nil  // 自动提交
    }); err != nil {
        return err
    }
    
    return nil
}
```

---

## ❌ 错误示例

### 示例 1: 多个数据库操作没有使用事务

```go
// ❌ 错误：多个数据库操作没有使用事务
func (s *fullReductionActivitySrv) Create(ctx context.Context, req *req.FullReductionActivityCreateReq) error {
    db := ctx.GetDB()
    
    // ❌ 错误：多个操作没有事务包裹，可能导致数据不一致
    // ❌ 同时错误：直接使用 db.Model().Create()，应该通过 Repository
    multiLanguageNameRepo := repository.NewMultiLanguageNameRepo(db)
    if err := multiLanguageNameRepo.Create(&multiLanguageName); err != nil {
        return err
    }
    
    activityRepo := repository.NewFullReductionActivityRepo(db)
    if err := activityRepo.Create(activity); err != nil {
        return err  // 如果这里失败，上面的 multiLanguageName 已经创建了，数据不一致
    }
    
    return nil
}
```

**问题：**
- 多个操作没有事务包裹，可能导致数据不一致
- 如果第二个操作失败，第一个操作已经提交，无法回滚

### 示例 2: 事务中使用了 db 而不是 tx

```go
// ❌ 错误：事务中使用了 db 而不是 tx
func (s *fullReductionActivitySrv) Create(ctx context.Context, req *req.FullReductionActivityCreateReq) error {
    db := ctx.GetDB()
    
    if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
        // ❌ 错误：在事务中使用了 db 而不是 tx
        activityRepo := repository.NewFullReductionActivityRepo(db)  // ❌ 应该用 tx
        if err := activityRepo.Create(activity); err != nil {
            return err
        }
        return nil
    }); err != nil {
        return err
    }
    
    return nil
}
```

**问题：**
- 在事务中使用了 `db` 而不是 `tx`，操作不会在同一个事务中执行
- 无法保证原子性

### 示例 3: 在事务中手动提交或回滚

```go
// ❌ 错误：在事务中手动提交或回滚
func (s *orderSrv) CreateOrder(ctx context.Context, req req.CreateOrderReq) error {
    db := ctx.GetDB()
    
    if err := db.Transaction(func(tx *gorm.DB) error {
        // ... 业务逻辑
        
        // ❌ 错误：不要手动提交，Transaction 会自动处理
        if err := tx.Commit().Error; err != nil {
            return err
        }
        
        return nil
    }); err != nil {
        return err
    }
    
    return nil
}
```

**问题：**
- Transaction 方法会自动处理提交和回滚
- 手动调用 `tx.Commit()` 或 `tx.Rollback()` 会导致错误

---

## 相关文档

- [Go Main 核心约束](../../../../.cursor/rules/go-main.mdc) - Service 层事务使用规范
- [Service 层示例](./service-layer-examples.md) - Service 层开发规范

---

**最后更新**: 2025-01-27  
**维护者**: TTPOS Team

