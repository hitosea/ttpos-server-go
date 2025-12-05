# 订单下单入口梳理

> **任务**: Phase 3 - Task 3.1  
> **目的**: 梳理所有创建/更新 SaleBill 的位置，识别需要保存国籍快照的代码点

---

## 📊 核心发现

### 国籍信息设置方式

国籍信息在 SaleBill 中通过以下两种方式设置：

1. **手动设置**：用户在订单中选择国籍（主要方式）
2. **创建时设置**：在创建订单时直接指定（较少见）

---

## 🔍 代码入口分析

### 1. 手动设置国籍（主要场景）⭐

**位置**: `main/app/service/order_base.go:252-273`

```go
// SetNationality 设置销售账单国籍
func (s *orderSrv) SetNationality(ctx context.Context, saleBillUuid uint64, nationalityUuid uint64) (resp.ShopCart, error) {
    // ...
    err := repository.NewSaleBillRepo(db).UpdateNationality(saleBillUuid, nationalityUuid)
    // ...
}
```

**Repository 层**: `main/app/repository/sale_bill.go:27`

```go
UpdateNationality(saleBillUuid uint64, nationalityUuid uint64) error
```

**✅ 修改点**：在 `repository.UpdateNationality()` 方法中添加快照保存逻辑

---

### 2. 创建订单时设置国籍（次要场景）

#### 2.1 即时订单（Instant Order）

**位置**: `main/app/service/order_base.go:42-90`

```go
// CreateInstantOrder 创建即时订单
saleBill, err := repository.NewOrderRepo(tx).CreateSaleBill(model.SaleBill{
    OrderNo:       orderNo,
    SerialNo:      serialNo,
    BillType:      constant.OrderSourceMapToBillType[constant.OrderSourceInstant],
    DiningMethod:  constant.SaleBillDiningMethodDineIn,
    DeviceUuid:    ctx.GetDeviceUuid(),
    Source:        constant.MapJwtSourceToSaleBillSource(ctx.GetSource()),
    ClientVersion: constant.NormalizeClientVersion(ctx.GetVersion()),
    // ⚠️ 目前未设置 NationalityUuid
})
```

**特点**：
- 创建时未设置 NationalityUuid
- 国籍由用户后续通过 `SetNationality()` 手动设置

---

#### 2.2 桌台订单（Desk Order）

**位置**: `main/app/service/order_base.go:98-227`

```go
// CreateDeskOrder 创建桌台订单
// 创建销售账单
if _, errCreateSaleBill := repository.NewOrderRepo(tx).CreateSaleBill(*saleBill); errCreateSaleBill != nil {
    return errCreateSaleBill
}
```

**特点**：
- saleBill 对象在方法参数中传入
- 创建时通常未设置 NationalityUuid
- 国籍由用户后续通过 `SetNationality()` 手动设置

---

#### 2.3 会员端订单（Member Order）

**位置**: `main/app/service/order.go:608-634`

```go
// createMemberOrder 创建会员订单
saleBill, err := repository.NewOrderRepo(db).CreateSaleBill(model.SaleBill{
    OrderNo:             orderNo,
    SerialNo:            "", 
    BillType:            constant.OrderSourceMapToBillType[constant.OrderSourceMember],
    DiningMethod:        constant.SaleBillDiningMethodTakeout,
    DeviceUuid:          ctx.GetDeviceUuid(),
    MemberSaleOrderUuid: memberSaleOrderUuid,
    Source:              constant.MapJwtSourceToSaleBillSource(ctx.GetSource()),
    ClientVersion:       constant.NormalizeClientVersion(ctx.GetVersion()),
    // ⚠️ 目前未设置 NationalityUuid
})
```

**特点**：
- 创建时未设置 NationalityUuid
- 国籍可能在会员信息中预设

---

#### 2.4 导入订单（Import Order）

**位置**: `main/app/service/order_import_service.go:497-531`

```go
// importOrderRecord 导入订单
saleBill := model.SaleBill{
    BaseModel: model.BaseModel{
        Uuid:       saleBillUuid,
        CreateTime: orderBasic.CreateTime,
        UpdateTime: orderBasic.CreateTime,
    },
    OrderNo:       orderBasic.OrderNo,
    // ...
    // ⚠️ 导入的历史订单可能包含 NationalityUuid，但不包含快照
}
orderRepo := repository.NewOrderRepo(db)
_, err := orderRepo.CreateSaleBill(saleBill)
```

**特点**：
- 导入历史订单，可能包含 NationalityUuid
- **⚠️ 特殊处理**：需要从关联表查询并补全快照

---

## 📋 修改计划

### Phase 3.2: 修改下单逻辑 - 保存国籍名称快照

#### 优先级 1️⃣：修改 Repository.UpdateNationality()

**文件**: `main/app/repository/sale_bill.go`

**原因**：
- 覆盖 90% 以上的场景（用户手动设置国籍）
- 统一入口，修改一处即可生效

**修改内容**：
1. 在更新 NationalityUuid 时：
   - 查询 Nationality 表获取 MultiLanguageName
   - 调用 `SaleBill.SetNationalityNameSnapshot()` 序列化为 JSON
   - 同时更新 `nationality_uuid` 和 `nationality_name` 字段

---

#### 优先级 2️⃣：修改 Repository.CreateSaleBill()

**文件**: `main/app/repository/order.go:88-96`

**原因**：
- 处理创建时已设置 NationalityUuid 的场景（较少见）
- 确保数据完整性

**修改内容**：
1. 在创建 SaleBill 前：
   - 检查是否设置了 NationalityUuid
   - 如果设置了，查询并保存快照

---

#### 优先级 3️⃣：修改导入逻辑（可选）

**文件**: `main/app/service/order_import_service.go`

**原因**：
- 处理历史订单导入
- 补全历史订单的快照数据

**修改内容**：
1. 导入时如果有 NationalityUuid：
   - 尝试从 Nationality 表（包括已删除）查询
   - 保存快照

---

## 🎯 下一步

继续执行 **Phase 3 - Task 3.2**: 修改下单逻辑 - 保存国籍名称快照

**按优先级顺序修改**：
1. `repository.UpdateNationality()` - 覆盖手动设置场景
2. `repository.CreateSaleBill()` - 覆盖创建时设置场景
3. 导入逻辑（可选） - 处理历史数据

---

**最后更新**: 2025-12-02  
**任务**: story-main-nationality-snapshot-fix (JSON 方案)

