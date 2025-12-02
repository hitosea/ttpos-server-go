# 常用模式示例

> Go Main 模块中常用开发模式的示例代码

---

## Service 构造函数

```go
func NewOrderSrv(
    dbm *database.DBManager,
    orderRepo repository.IOrderRepo,
    memberSrv IMemberSrv,
) IOrderSrv {
    return &orderSrv{
        dbm:       dbm,
        orderRepo: orderRepo,
        memberSrv: memberSrv,
    }
}
```

---

## Repository 选项模式

```go
func (r *OrderRepoImpl) WhereUuid(uuid uint64) DBOption {
    return func(db *gorm.DB) *gorm.DB {
        return db.Where("uuid = ?", uuid)
    }
}

// 使用
order, err := orderRepo.GetOrder(
    orderRepo.WhereUuid(orderUuid),
    orderRepo.WhereStatus(1),
)
```

---

## 事件发布

```go
import "ttpos-server-go/pkg/utils"

utils.Go(func() {
    event.NewSystemBus().PublishOrderCreatedEvent(
        event.OrderCreatedPayload{
            BasePayload: event.BasePayload{
                Ctx: ctx,
                CompanyUuid: ctx.GetCompanyUuid(),
            },
            OrderUuid: orderUuid,
        },
    )
})
```

---

## UUID 锁

```go
    s.systemLock.LockUuid(deskUuid)
    defer s.systemLock.UnlockUuid(deskUuid)

    // 业务逻辑
```

---

## 相关文档

- [Go Main 核心约束](../../../../.cursor/rules/go-main.mdc) - 常用模式
- [协程使用示例](./goroutine-examples.md) - 事件发布中的协程使用

---

**最后更新**: 2025-01-27  
**维护者**: TTPOS Team

