# Service 层示例

> Go Main 模块中 Service 层开发的正确和错误示例

---

## 规范说明

- ✅ **接口定义（IxxxSrv）和实现（xxxSrv）必须放在同一个文件中**
- ✅ **接口第一个参数必须使用 `ctx context.Context`**
- ✅ **通过 `db := ctx.GetDB()` 获取数据库连接**
- ✅ **只能依赖其他 Service 接口**
- ✅ **持有 DBManager**（仅在极少数需要切换数据库时使用）
- ✅ **可以依赖自己领域的 Repository 接口**
- ✅ **使用带调用链跟踪的DB，通过DBManager.GetDBWithContext() 获取DB实例**
- ❌ **不能直接依赖其他领域的 Repository**
- ❌ **不能直接操作 model，必须通过 Repository**
- ❌ **禁止直接使用 `db.Model()`、`db.Create()`、`db.Update()`、`db.Delete()`、`db.Where()` 等 GORM 方法**
- ✅ **所有数据库操作必须通过 Repository 层封装后再调用**

---

## ✅ 正确示例

### 示例 1: 接口和实现在同一个文件中

```go
// ✅ 正确：接口和实现在同一个文件中
// 文件：order_srv.go
package service

type IOrderSrv interface {
    CreateOrder(ctx context.Context, req req.CreateOrderReq) error
}

type orderSrv struct {
    dbm        *database.DBManager  // ✅ 持有 DBManager（极少使用）
    orderRepo  IOrderRepo            // ✅ 自己领域的 Repository
    memberSrv  IMemberSrv            // ✅ 其他领域的 Service 接口
    memberRepo IMemberRepo           // ❌ 不能依赖其他领域的 Repository
}
```

### 示例 2: 使用 ctx.GetDB() 并通过 Repository 操作

```go
// ✅ 正确：使用 ctx.GetDB() 并通过 Repository 操作
func (s *orderSrv) CreateOrder(ctx context.Context, req req.CreateOrderReq) error {
    db := ctx.GetDB()
    orderRepo := repository.NewOrderRepo(db)
    order := model.Order{...}
    orderUuid, err := orderRepo.CreateOrder(order)  // ✅ 通过 Repository
    if err != nil {
        return errors.WithMessage(err, "创建订单失败")
    }
    return nil
}
```

### 示例 3: 所有数据库操作都通过 Repository

```go
// ✅ 正确：所有数据库操作都通过 Repository
func (s *fullReductionActivitySrv) Create(ctx context.Context, req *req.FullReductionActivityCreateReq) error {
    db := ctx.GetDB()
    
    // ✅ 正确：通过 Repository 创建多语言名称
    multiLanguageNameRepo := repository.NewMultiLanguageNameRepo(db)
    multiLanguageName := model.MultiLanguageName{...}
    if err := multiLanguageNameRepo.Create(&multiLanguageName); err != nil {
        return errors.WithMessage(err, "创建多语言名称失败")
    }
    
    // ✅ 正确：通过 Repository 更新多语言名称
    if err := multiLanguageNameRepo.UpdateByUuid(uuid, updates); err != nil {
        return errors.WithMessage(err, "更新多语言名称失败")
    }
    
    return nil
}
```

---

## ❌ 错误示例

### 示例 1: 接口和实现分开在不同文件

```go
// ❌ 错误：接口和实现分开在不同文件
// 文件：i_order_srv.go
// type IOrderSrv interface { ... }

// 文件：order_srv.go
// type orderSrv struct { ... }
```

**问题：** 不符合项目规范，接口和实现必须放在同一个文件中

### 示例 2: 没有 ctx 参数

```go
// ❌ 错误：没有 ctx 参数
func (s *orderSrv) CreateOrder(req req.CreateOrderReq) error {
    // ...
}
```

**问题：** Service 接口第一个参数必须是 `ctx context.Context`

### 示例 3: 使用 dbm.GetDB(dbId) 而不是 ctx.GetDB()

```go
// ❌ 错误：使用 dbm.GetDB(dbId) 而不是 ctx.GetDB()
func (s *orderSrv) CreateOrder(ctx context.Context, req req.CreateOrderReq) error {
    db := s.dbm.GetDB(ctx.GetDbId())  // ❌ 应该用 ctx.GetDB()
    orderRepo := repository.NewOrderRepo(db)
    // ...
}
```

**问题：** 必须使用 `ctx.GetDB()` 获取数据库连接

### 示例 4: Service 直接操作 model

```go
// ❌ 错误：Service 直接操作 model
func (s *orderSrv) CreateOrder(ctx context.Context, req req.CreateOrderReq) error {
    db := ctx.GetDB()
    order := model.Order{...}
    return db.Create(&order).Error  // ❌ 直接操作 model
}
```

**问题：** Service 不能直接操作 model，必须通过 Repository

### 示例 5: 直接使用 db.Model、db.Create 等方法

```go
// ❌ 错误：直接使用 db.Model、db.Create 等方法
func (s *fullReductionActivitySrv) Create(ctx context.Context, req *req.FullReductionActivityCreateReq) error {
    db := ctx.GetDB()
    
    // ❌ 错误：直接使用 db.Model().Create()
    if err := db.Model(&model.MultiLanguageName{}).Create(&multiLanguageName).Error; err != nil {
        return err
    }
    
    // ❌ 错误：直接使用 db.Model().Where().Updates()
    if err := db.Model(&model.MultiLanguageName{}).
        Where("uuid = ?", uuid).
        Updates(map[string]interface{}{...}).Error; err != nil {
        return err
    }
    
    return nil
}
```

**问题：** 禁止直接使用 GORM 方法，必须通过 Repository 层封装

---

## 相关文档

- [Go Main 核心约束](../../../../.cursor/rules/go-main.mdc) - Service 层规范
- [事务使用示例](./transaction-examples.md) - Service 层事务使用

---

**最后更新**: 2025-01-27  
**维护者**: TTPOS Team

