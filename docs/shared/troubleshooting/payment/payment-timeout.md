---
Status: Deprecated
---

# 支付接口超时（Go 主服务）

> 适用范围：`main/app/service/payment`、`main/app/api/v1/payment`、POS 快捷支付/收银通道

## 问题现象

- POS 或第三方调用 `/api/v1/order/quick_payment`、`/api/v1/payment/create` 时超过 **200ms** SLA。
- Gin 日志/Prometheus 显示 `context deadline exceeded` 或 `payment timeout`.
- RocketMQ/事件总线出现重复的支付完成事件，或订单状态先成功又回滚。

## 问题原因

1. **SystemLock 未释放**：同一订单或桌台 UUID 反复加锁，导致下一次支付无法进入 Service 层。
2. **外部网关延迟**：第三方支付（微信/支付宝）超过默认 2s 超时时间，调用卡死。
3. **库存/积分串行流程过长**：`OrderSrv`、`PointSrv` 等串行调用导致临界路径 >200ms。
4. **数据库争用**：`ttpos_order`/`ttpos_payment` 行级锁（事务未提交）阻塞。

## 解决方案

1. **复位并监控锁**

   ```go
   // main/app/service/payment/payment_srv.go
   s.systemLock.LockUuid(orderUuid)
   defer s.systemLock.UnlockUuid(orderUuid)
   ```

   - 确保 `defer UnlockUuid` 在所有返回路径执行。
   - 检查 `pkg/lock/system_lock.go` 是否配置合理的过期时间（默认 5s，可视情况调小）。

2. **拆分外部网关调用**

   ```go
   ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
   defer cancel()
   result, err := gateway.Pay(ctx, req)
   ```

   - 为第三方调用单独设置 Context Timeout，并在失败时返回明确错误码，不阻塞主流程。

3. **异步化非关键步骤**

   - 订单状态更新、支付记录写入保留同步。
   - 库存扣减、积分累计、打印等放入事件：
     ```go
     go event.NewSystemBus().PublishOrderPaidEvent(...)
     ```

4. **数据库排查**
   ```sql
   -- 查看锁等待
   SELECT * FROM information_schema.innodb_lock_waits;
   -- 针对大事务执行 explain，确保有索引 (uuid/status)
   EXPLAIN SELECT * FROM ttpos_order WHERE uuid = ? FOR UPDATE;
   ```
   - 使用 `php think migrate:status` 确认迁移一致，避免字段缺失导致回滚。

## 预防措施

- `main/app/service/payment/payment_srv.go` 中新增 Prometheus 计时器，超时报警。
- 对关键 Service（Payment/Order）启用 `go test -run TestPaymentSrv -cover`，保证 100% 覆盖率，避免遗漏 `defer Unlock`.
- 配置 `pkg/lock` 的 `defaultExpire` ≤ 3s，并在调用前检查锁状态，必要时降级为读取缓存价格。
- 在 `docs/shared/specs` 的支付相关 Spec 中明确 SLA 与并发策略。

## 相关资源

- `.cursor/rules/go-main.mdc`
- `.cursor/rules/api.mdc`
- `.cursor/rules/knowledge_management.mdc`
- `main/app/service/payment/payment_srv.go`
- `main/app/event/order_paid_event.go`
- `pkg/lock/system_lock.go`

Related Episode: `[待补充 by @责任人]`
