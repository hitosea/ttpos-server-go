# Repository 层示例

> Go Main 模块中 Repository 层开发的正确和错误示例

---

## 规范说明

- ✅ **只能持有 db 实例**
- ❌ **不能持有 DBManager**
- ✅ **使用选项模式**
- ✅ **使用 map 更新零值字段**

---

## 📋 目录

1. [结构体定义规范](#结构体定义规范)
2. [零值字段更新最佳实践](#零值字段更新最佳实践)
3. [相关文档](#相关文档)

---

## 结构体定义规范

### ✅ 正确示例

```go
type OrderRepoImpl struct {
    db  *gorm.DB           // ✅ 正确
    dbm *database.DBManager // ❌ 错误
}
```

### ❌ 错误示例

```go
type OrderRepoImpl struct {
    dbm *database.DBManager // ❌ 错误：Repository 不能持有 DBManager
}
```

**问题：**
- Repository 层只能持有 db 实例
- 持有 DBManager 违反了分层架构原则

---

## 零值字段更新最佳实践

### 问题背景

GORM 的 `Updates` 方法默认会忽略零值字段（zero values）。当需要将字段从非零值更新为零值时（如 `IsAllDay` 从 `1` 更新为 `0`），使用 struct 直接更新会失败。

**问题示例：**

```go
// ❌ 错误：IsAllDay=0 不会被更新
activity.IsAllDay = 0
db.Where("uuid = ?", activity.Uuid).Updates(activity)
```

### ✅ 推荐方案：封装 map 更新方法

**接口定义：**

```go
type IFullReductionActivityRepo interface {
    // 通用更新方法（不推荐，零值字段不会更新）
    Update(activity *model.FullReductionActivity, options ...DBOption) error
    
    // 灵活的 map 更新方法
    UpdateByMap(uuid uint64, data map[string]interface{}) error
    
    // 业务专用更新方法（推荐）
    UpdateActivity(activity *model.FullReductionActivity) error
    UpdateDisabled(activity *model.FullReductionActivity) error
}
```

**Repository 层实现：**

```go
// Update 更新满减活动（不推荐，建议使用业务专用方法）
func (r *FullReductionActivityRepoImpl) Update(activity *model.FullReductionActivity, options ...DBOption) error {
    db := r.db.Model(&model.FullReductionActivity{})
    for _, option := range options {
        db = option(db)
    }
    // 使用 Updates 更新非零值字段
    // 注意：零值字段（如 IsAllDay=0）不会被更新
    return errors.WithMessage(db.Where("uuid = ?", activity.Uuid).Updates(activity).Error)
}

// UpdateByMap 使用 map 更新（灵活但不够语义化）
func (r *FullReductionActivityRepoImpl) UpdateByMap(uuid uint64, data map[string]interface{}) error {
    return errors.WithMessage(
        r.db.Model(&model.FullReductionActivity{}).
            Where("uuid = ?", uuid).
            Updates(data).Error,
    )
}

// UpdateActivity 更新活动基本信息（推荐）
// 封装业务逻辑，使用 map 确保零值字段也能更新
func (r *FullReductionActivityRepoImpl) UpdateActivity(activity *model.FullReductionActivity) error {
    return errors.WithMessage(
        r.db.Model(&model.FullReductionActivity{}).
            Where("uuid = ?", activity.Uuid).
            Updates(map[string]interface{}{
                "name":           activity.Name,
                "start_date":     activity.StartDate,
                "end_date":       activity.EndDate,
                "start_time":     activity.StartTime,
                "end_time":       activity.EndTime,
                "is_all_day":     activity.IsAllDay,      // ✅ 零值也能更新
                "reduction_type": activity.ReductionType,
                "update_time":    activity.UpdateTime,
            }).Error,
    )
}

// UpdateDisabled 更新失效状态（推荐）
func (r *FullReductionActivityRepoImpl) UpdateDisabled(activity *model.FullReductionActivity) error {
    return errors.WithMessage(
        r.db.Model(&model.FullReductionActivity{}).
            Where("uuid = ?", activity.Uuid).
            Updates(map[string]interface{}{
                "is_disabled": activity.IsDisabled, // ✅ 零值也能更新
                "update_time": activity.UpdateTime,
            }).Error,
    )
}
```

**Service 层调用：**

```go
// ✅ 正确：使用业务专用方法
func (s *fullReductionActivitySrv) Update(ctx context.Context, req *req.FullReductionActivityUpdateReq) error {
    // ... 省略查询和验证逻辑 ...
    
    // 设置要更新的字段
    activity.Name = req.LocaleName.ToJson()
    activity.StartDate = req.StartDate
    activity.EndDate = req.EndDate
    activity.StartTime = req.StartTime
    activity.EndTime = req.EndTime
    activity.IsAllDay = req.IsAllDay  // 可能是 0
    activity.ReductionType = req.ReductionType
    activity.UpdateTime = currentTime
    
    // 调用业务专用更新方法
    activityRepo := repository.NewFullReductionActivityRepo(tx)
    if err := activityRepo.UpdateActivity(activity); err != nil {
        return errors.WithMessage(err, "更新活动失败")
    }
    
    return nil
}

func (s *fullReductionActivitySrv) Disable(ctx context.Context, uuid uint64) error {
    // ... 省略查询和验证逻辑 ...
    
    // 设置失效状态
    activity.IsDisabled = constant.Yes
    activity.UpdateTime = currentTime
    
    // 调用业务专用更新方法
    if err := activityRepo.UpdateDisabled(activity); err != nil {
        return errors.WithMessage(err, "失效活动失败")
    }
    
    return nil
}
```

### ❌ 不推荐的方案

**方案 1: 使用 Select("*")**

```go
// ❌ 不推荐：会更新所有字段，包括不应该更新的字段
db.Select("*").Where("uuid = ?", activity.Uuid).Updates(activity)
```

**问题：**
- 可能更新 BaseModel 中的敏感字段（`ID`、`Uuid`、`CreateTime`、`DeleteTime`）
- 可能更新关联对象字段
- 不够安全

**方案 2: 使用 Select 列举字段**

```go
// ❌ 不推荐：冗长且容易遗漏字段
db.Select(
    "name", "start_date", "end_date", "start_time", "end_time",
    "is_all_day", "reduction_type", "is_disabled", "update_time",
).Where("uuid = ?", activity.Uuid).Updates(activity)
```

**问题：**
- 代码重复，每个更新方法都要写一遍字段列表
- 新增字段时容易遗漏
- 不够优雅

**方案 3: Service 层直接构建 map**

```go
// ❌ 不推荐：违反分层原则
// Service 层
if err := activityRepo.UpdateByMap(req.Uuid, map[string]interface{}{
    "name":           req.LocaleName.ToJson(),
    "start_date":     req.StartDate,
    "end_date":       req.EndDate,
    "start_time":     req.StartTime,
    "end_time":       req.EndTime,
    "is_all_day":     req.IsAllDay,
    "reduction_type": req.ReductionType,
    "update_time":    currentTime,
}); err != nil {
    return errors.WithMessage(err, "更新活动失败")
}
```

**问题：**
- Repository 层只提供通用方法，业务逻辑泄漏到 Service 层
- Service 层需要知道数据库字段名
- 代码复用性差，多处调用需要重复构建 map
- 违反职责分离原则

### 方案对比总结

| 方案 | 优点 | 缺点 | 推荐度 |
|------|------|------|--------|
| **业务专用方法** | 职责清晰、类型安全、易维护、零值安全 | 需要为每个业务场景编写方法 | ⭐⭐⭐⭐⭐ |
| UpdateByMap | 灵活 | 不够语义化、Service 层需要知道字段名 | ⭐⭐⭐ |
| Select("*") | 简单 | 不安全，可能误更新敏感字段 | ⭐ |
| Select 列举字段 | 相对安全 | 冗长、容易遗漏 | ⭐⭐ |
| struct Updates | 简洁 | 不支持零值更新 | ❌ |

### 项目中的类似实践

项目中已有多处采用此模式：

1. **warehouse_item.go - 库存更新：**
```go
func (r *WarehouseItemRepoImpl) UpdateStock(uuid uint64, stock, reservedStock float64) error {
    return r.db.Model(&model.WarehouseItem{}).Where("uuid = ?", uuid).Updates(map[string]interface{}{
        "stock":          stock,
        "reserved_stock": reservedStock,
        "update_time":    time.Now().Unix(),
    }).Error
}
```

2. **desk_map_repository.go - 布局更新：**
```go
func (r *DeskMapLayoutRepoImpl) UpdateLayout(regionUuid uint64, layout model.DeskMapLayout) error {
    err := r.db.Model(&model.DeskMapLayout{}).
        Where("region_uuid = ? AND delete_time = ?", regionUuid, 0).
        Updates(map[string]interface{}{
            "layout_json": layout.LayoutJson,
            "update_time": layout.UpdateTime,
        }).Error
    return errors.WithMessage(err)
}
```

3. **sale_order.go - 订单字段更新：**
```go
func (r *saleOrderRepo) UpdateSaleOrderCashier(ctx context.Context, saleOrderUuid uint64, cashierUuid uint64, cashierName string) error {
    return r.db.Model(&model.SaleOrder{}).Where("uuid = ?", saleOrderUuid).Updates(map[string]interface{}{
        "cashier_uuid": cashierUuid,
        "cashier_name": cashierName,
    }).Error
}
```

### 最佳实践总结

1. ✅ **Repository 层封装业务专用更新方法**
   - 使用语义化的方法名（如 `UpdateActivity`、`UpdateDisabled`）
   - 内部使用 `map[string]interface{}` 确保零值字段能更新
   - 接收 model 对象，保持接口一致性

2. ✅ **Service 层调用业务专用方法**
   - 设置 model 对象的字段值
   - 调用 Repository 的业务专用方法
   - 保持代码简洁和语义清晰

3. ✅ **保留通用方法**
   - 保留 `Update` 方法用于不涉及零值的场景
   - 保留 `UpdateByMap` 方法用于特殊场景
   - 在注释中说明各方法的适用场景

4. ✅ **明确字段列表**
   - 在 Repository 方法中明确列出要更新的字段
   - 排除 BaseModel 敏感字段（`ID`、`Uuid`、`CreateTime`、`DeleteTime`）
   - 增加字段时只需修改 Repository 层

---

## 相关文档

- [Go Main 核心约束](../../../../.cursor/rules/go-main.mdc) - Repository 层规范

---

**最后更新**: 2025-11-25  
**维护者**: TTPOS Team
