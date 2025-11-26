# 并台、转台、转菜操作使用多个订单锁 需求提案

> 本文档用于需求发起阶段，经团队评审后创建正式 Spec。

---

## 📋 提案信息

| 项目       | 内容     |
| ---------- | -------- |
| **提案人** | xiezhihuan   |
| **日期**   | 2025-11-26   |
| **目标版本** | {版本号} |
| **状态**   | 进行中   |
| **关联任务** | - |
| **关联 Spec** | [story-main-table-multi-order-lock](../../../shared/specs/active/story-main-table-multi-order-lock/)      |
| **关联提案** | `table-merge-transfer-lock-uuid.md`（已放弃公司锁方案） |

---

## 🎯 背景和动机

### 问题描述

经过综合考虑，决定放弃使用公司级别锁的方案，改为使用订单级别的锁。当操作涉及多个订单时，需要获取所有相关订单的锁。

**当前问题**：

1. **并台操作**：同时使用 `SaleBillUuid` 和 `companyUuid` 两个锁
   - 问题：锁粒度不一致，双重锁可能导致死锁风险
   - 实际情况：并台操作涉及多个订单（主订单 + 所有被合并的订单），需要锁定所有涉及的订单
   - 位置：`main/app/service/desk.go:806-807`

2. **转台操作**：使用 `SaleBillUuid` 作为锁 UUID
   - 问题：
     - 转台操作只锁定源订单，但会修改目标桌台的状态（`desk.SetOpenDesk(reqs.SaleBillUuid)`）
     - 开台操作锁定目标桌台（`DeskUuid`），两者锁粒度不一致，存在并发问题
     - **并发场景**：转台操作检查新桌台是否空闲时，开台操作也在检查同一桌台，两者都认为桌台空闲，导致数据不一致
   - 位置：`main/app/service/desk.go:698`
   - 相关操作：开台操作使用 `DeskUuid` 锁（`main/app/service/order_base.go:99`）

3. **转菜操作**：使用 `SaleBillUuid` 作为锁 UUID
   - 问题：转菜操作涉及源订单和目标订单两个订单，当前只锁定了源订单，可能导致数据不一致
   - 位置：`main/app/service/order_product.go:1063`

### 业务价值

- **提高并发性能**：使用订单级别的锁，不同订单的操作可以并发执行，提高系统吞吐量
- **保证数据一致性**：当操作涉及多个订单时，锁定所有相关订单，避免数据竞争
- **避免死锁**：通过统一的锁获取顺序（按订单 UUID 排序），避免死锁风险
- **符合业务逻辑**：并台、转台、转菜操作本质上是订单级别的操作，使用订单锁更符合业务语义

### 目标用户

- [x] 收银员
- [ ] 商户管理员
- [ ] 厨房人员
- [ ] 顾客
- [ ] 其他: ________

---

## 💡 解决方案概述

### 方案描述

统一并台、转台和转菜操作的锁机制，使用订单级别的锁（`SaleBillUuid`），当操作涉及多个订单时，需要获取所有相关订单的锁，并按照统一的顺序获取锁以避免死锁。

**核心原则**：
1. **并台操作**：锁定主订单 + 所有被合并的订单（按订单 UUID 排序后依次获取）
2. **转台操作**：锁定源订单 + 目标桌台（需要锁定目标桌台，避免与开台操作并发冲突）
   - **并发问题**：开台操作锁定桌台UUID，转台操作需要锁定目标桌台，避免两者同时操作同一桌台
3. **转菜操作**：锁定源订单和目标订单（按订单 UUID 排序后依次获取）

**修改点**：

1. **并台操作（MergeDesk）**
   - 当前：同时锁定 `req.SaleBillUuid` 和 `companyUuid`
   - 修改为：锁定主订单 + 所有被合并的订单（按 UUID 排序）

2. **转台操作（ChangeDesk）**
   - 当前：`lock.NewSystemLock().LockUuid(reqs.SaleBillUuid)`
   - 修改为：锁定源订单 + 目标桌台（按 UUID 排序）
   - **原因**：转台操作会修改目标桌台状态（`desk.SetOpenDesk`），需要锁定目标桌台，避免与开台操作并发冲突

3. **转菜操作（InstantOrderCartProductChangeDesk）**
   - 当前：只锁定源订单 `req.SaleBillUuid`
   - 修改为：锁定源订单和目标订单（按 UUID 排序）

### 核心功能点

1. 修改并台操作的锁机制，移除 `companyUuid` 锁，改为锁定所有涉及的订单
2. 修改转菜操作的锁机制，增加目标订单的锁
3. 实现统一的锁获取顺序（按订单 UUID 排序），避免死锁
4. 确保锁的获取和释放顺序正确（按相反顺序释放）
5. 保持现有业务逻辑不变，仅调整锁的粒度

### 影响范围

**涉及终端**：
- [x] POS 收银端
- [ ] Shop 商家管理端
- [ ] KDS 厨显端
- [ ] QDS 排号叫号端
- [x] Assistant 助手端
- [ ] Tablet 平板端
- [ ] Mobile 扫码端
- [ ] Menu 电子菜单端
- [ ] Member 会员端

**涉及模块**：
- [ ] UI 组件
- [ ] API 接口
- [ ] 数据模型
- [x] 业务逻辑
- [ ] 第三方集成
- [ ] 其他: 并发控制模块

---

## 📊 初步评估

### 技术复杂度

- [ ] **低**：纯 UI 调整，无业务逻辑变更
- [x] **中**：需要前后端联调，基础业务逻辑
- [ ] **高**：涉及架构调整、第三方集成、复杂算法

### 工作量预估

[粗略估算，具体 SP 在技术方案评审时确定]

- **预计天数**: 2-3 天
- **预估 SP**: 3-5 SP（待技术评审确认）

### 风险识别

**潜在风险**：
1. **死锁风险**：多个订单锁可能导致死锁
   - **缓解措施**：统一按订单 UUID 排序后获取锁，确保所有操作都按相同顺序获取锁

2. **锁获取顺序错误**：如果不同操作获取锁的顺序不一致，可能导致死锁
   - **缓解措施**：实现统一的锁获取工具函数，确保所有操作都使用相同的排序规则

3. **性能影响**：并台操作需要锁定多个订单，可能影响并发性能
   - **缓解措施**：并台操作相对较少，性能影响可接受；不同订单的操作可以并发执行

4. **代码复杂度增加**：需要管理多个锁的获取和释放
   - **缓解措施**：封装锁管理工具函数，简化代码

5. **转台操作与开台操作的并发冲突**：
   - **问题**：转台操作需要修改目标桌台状态，如果只锁定订单而不锁定目标桌台，可能与开台操作并发冲突
   - **场景**：转台操作检查新桌台是否空闲时，开台操作也在检查同一桌台，两者都认为桌台空闲，导致数据不一致
   - **缓解措施**：转台操作需要锁定目标桌台（`DeskUuid`），与开台操作使用相同的锁，确保串行执行

**缓解措施**：
1. **实现统一的锁管理工具**：封装多订单锁的获取和释放逻辑
2. **代码审查**：确保所有操作都使用统一的锁获取顺序
3. **单元测试**：测试多订单锁的获取和释放逻辑
4. **集成测试**：测试并台、转台、转菜操作的并发场景
   - **重点测试**：转台操作与开台操作同时操作同一桌台的并发场景
   - 验证转台操作锁定目标桌台后，开台操作会被正确阻塞
5. **性能测试**：在高并发场景下测试锁的性能影响
6. **锁顺序验证**：确保转台操作锁定源订单和目标桌台时，按 UUID 排序获取锁

---

## 🔗 相关资源

### 参考需求

- 相关提案: `table-merge-transfer-lock-uuid.md`（公司锁方案，已放弃）
- 类似功能: 商品导入使用 `company_uuid` 作为锁（`main/app/service/product.go:4767`）

### 相关文档

- 并发锁设计文档: `docs/human/architecture/concurrency_lock_design.md`
- 代码位置:
  - 转台: `main/app/service/desk.go:694-794`
  - 并台: `main/app/service/desk.go:799-1010`
  - 转菜: `main/app/service/order_product.go:1060-1220`
  - 锁实现: `main/pkg/lock/lock_redsync.go`

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

- [x] 创建 Spec：`story-main-table-multi-order-lock`
- [ ] 分配负责人：xiezhihuan
- [ ] 目标 Sprint：Sprint {N}

---

## 📝 附录

### User Story（初稿）

**作为** 收银员  
**我想** 在并台、转台和转菜时使用订单级别的锁  
**以便于** 提高并发性能，同时保证数据一致性

### AC 验收标准（初稿）

1. **WHEN** 执行并台操作 **THEN** 系统 **SHALL** 锁定主订单和所有被合并的订单（按 UUID 排序）
2. **WHEN** 执行转台操作 **THEN** 系统 **SHALL** 锁定源订单和目标桌台（按 UUID 排序）
3. **WHEN** 执行转菜操作 **THEN** 系统 **SHALL** 锁定源订单和目标订单（按 UUID 排序）
4. **IF** 转台操作和开台操作同时涉及同一桌台 **THEN** 系统 **SHALL** 串行执行，避免并发冲突
4. **IF** 多个操作同时涉及相同订单 **THEN** 系统 **SHALL** 串行执行，避免并发冲突
5. **WHEN** 锁获取失败 **THEN** 系统 **SHALL** 返回适当的错误提示
6. **WHEN** 所有操作都按订单 UUID 排序获取锁 **THEN** 系统 **SHALL** 避免死锁

### 代码修改示例

**并台操作修改**：

```go
// 当前实现（main/app/service/desk.go:804-813）
if ctx.NoLock() {
    systemLock.LockUuid(req.SaleBillUuid)
    systemLock.LockUuid(companyUuid)
    defer func() {
        systemLock.UnlockUuid(companyUuid)
        systemLock.UnlockUuid(req.SaleBillUuid)
    }()
    ctx.AddLock()
}

// 修改后：锁定所有涉及的订单（主订单 + 被合并的订单）
if ctx.NoLock() {
    // 收集所有需要锁定的订单 UUID
    orderUuids := []uint64{req.SaleBillUuid}
    for _, deskUuid := range req.DeskUuids {
        if deskUuid != saleBill.DeskUuid {
            desk, _ := repository.NewDeskRepo(db).GetDeskRecord(deskUuid)
            if desk != nil && desk.SaleBillUuid != 0 {
                orderUuids = append(orderUuids, desk.SaleBillUuid)
            }
        }
    }
    // 按 UUID 排序，确保锁获取顺序一致
    sort.Slice(orderUuids, func(i, j int) bool {
        return orderUuids[i] < orderUuids[j]
    })
    // 依次获取锁
    for _, uuid := range orderUuids {
        systemLock.LockUuid(uuid)
    }
    // 按相反顺序释放锁
    defer func() {
        for i := len(orderUuids) - 1; i >= 0; i-- {
            systemLock.UnlockUuid(orderUuids[i])
        }
    }()
    ctx.AddLock()
}
```

**转台操作修改**：

```go
// 当前实现（main/app/service/desk.go:697-701）
if ctx.NoLock() {
    lock.NewSystemLock().LockUuid(reqs.SaleBillUuid)
    defer lock.NewSystemLock().UnlockUuid(reqs.SaleBillUuid)
    ctx.AddLock()
}

// 修改后：锁定源订单和目标桌台（避免与开台操作并发冲突）
if ctx.NoLock() {
    systemLock := lock.NewSystemLock()
    
    // 收集需要锁定的资源：源订单 + 目标桌台
    lockUuids := []uint64{reqs.SaleBillUuid, reqs.DeskUuid}
    
    // 按 UUID 排序，确保锁获取顺序一致
    sort.Slice(lockUuids, func(i, j int) bool {
        return lockUuids[i] < lockUuids[j]
    })
    
    // 依次获取锁
    for _, uuid := range lockUuids {
        systemLock.LockUuid(uuid)
    }
    
    // 按相反顺序释放锁
    defer func() {
        for i := len(lockUuids) - 1; i >= 0; i-- {
            systemLock.UnlockUuid(lockUuids[i])
        }
    }()
    ctx.AddLock()
}
```

**转菜操作修改**：

```go
// 当前实现（main/app/service/order_product.go:1062-1066）
if ctx.NoLock() {
    s.lock.LockUuid(req.SaleBillUuid)
    defer s.lock.UnlockUuid(req.SaleBillUuid)
    ctx.AddLock()
}

// 修改后：锁定源订单和目标订单
if ctx.NoLock() {
    // 获取目标订单 UUID
    targetDesk, _ := repository.NewDeskRepo(db).GetDeskAndSaleBillByDeskUuid(req.DeskUuid)
    targetOrderUuid := targetDesk.SaleBillUuid
    
    // 按 UUID 排序，确保锁获取顺序一致
    orderUuids := []uint64{req.SaleBillUuid, targetOrderUuid}
    sort.Slice(orderUuids, func(i, j int) bool {
        return orderUuids[i] < orderUuids[j]
    })
    
    // 依次获取锁
    for _, uuid := range orderUuids {
        s.lock.LockUuid(uuid)
    }
    
    // 按相反顺序释放锁
    defer func() {
        for i := len(orderUuids) - 1; i >= 0; i-- {
            s.lock.UnlockUuid(orderUuids[i])
        }
    }()
    ctx.AddLock()
}
```

### 锁管理工具函数（建议）

```go
// LockMultipleOrders 锁定多个订单（按 UUID 排序）
func LockMultipleOrders(lock Lock, orderUuids []uint64) {
    // 去重并排序
    uniqueUuids := make([]uint64, 0)
    seen := make(map[uint64]bool)
    for _, uuid := range orderUuids {
        if uuid != 0 && !seen[uuid] {
            uniqueUuids = append(uniqueUuids, uuid)
            seen[uuid] = true
        }
    }
    sort.Slice(uniqueUuids, func(i, j int) bool {
        return uniqueUuids[i] < uniqueUuids[j]
    })
    
    // 依次获取锁
    for _, uuid := range uniqueUuids {
        lock.LockUuid(uuid)
    }
}

// UnlockMultipleOrders 释放多个订单锁（按相反顺序）
func UnlockMultipleOrders(lock Lock, orderUuids []uint64) {
    // 按相反顺序释放
    for i := len(orderUuids) - 1; i >= 0; i-- {
        lock.UnlockUuid(orderUuids[i])
    }
}
```

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
**创建日期**: 2025-11-26  
**维护者**: 产品组 + Scrum Master  
**相关规范**: `.cursor/rules/scrum_story_point.mdc`, `.cursor/rules/specs.mdc`

