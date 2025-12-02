# 订单查询入口梳理

> **任务**: Phase 4 - Task 4.1  
> **目的**: 梳理所有返回国籍信息的查询接口，识别需要使用快照的代码点

---

## 📊 核心发现

### 国籍信息展示方式

订单查询响应中，国籍信息通过以下两种方式展示：

1. **返回 UUID + Name**：订单详情等完整信息查询
2. **仅返回 UUID**：购物车等轻量级查询（前端自行查询配置表）

---

## 🔍 代码入口分析

### 1. 订单详情查询（需修改）⭐

**位置**: `main/app/service/order_manage.go:691-696`

**方法**: `GetOrderInfos()`

**当前实现**:
```go
NationalityName: func() string {
    if saleBill.Nationality != nil {
        return saleBill.Nationality.MultiLanguageName.GetNameByLang(ctx.GetLanguage())
    }
    return ""
}(),
```

**问题**：
- ❌ 直接使用关联表 `Nationality.MultiLanguageName`
- ❌ 如果关联表数据被删除，返回空字符串
- ❌ 不使用快照数据

**响应结构**: `dto/resp/order.go:148`
```go
type OrderInfos struct {
    NationalityUuid uint64 `json:"nationality_uuid"` // 国籍UUID
    NationalityName string `json:"nationality_name"` // 国籍名称（单语言）
    // ...
}
```

**✅ 修改方案**：
```go
NationalityName: func() string {
    return saleBill.GetLocaleNationalityName().GetByLang(ctx.GetLanguage())
}(),
```

---

### 2. 购物车查询（无需修改）

#### 2.1 点餐购物车 - `GetOrderCartInfo()`

**位置**: `main/app/service/order_product.go:1817`

**当前实现**:
```go
shopCartInfo := &resp.ShopCart{
    SaleBillUuid:    saleBillUuid,
    OrderSourceUuid: shopCart.SaleBill.OrderSourceUuid,
    NationalityUuid: shopCart.SaleBill.NationalityUuid, // 仅返回 UUID
    // ...
}
```

**响应结构**: `dto/resp/shop_cart.go:158`
```go
type ShopCart struct {
    NationalityUuid uint64 `json:"nationality_uuid"` // 国籍UUID（仅UUID，无Name）
    // ...
}
```

**✅ 无需修改**：
- 仅返回 UUID，不返回名称
- 前端通过国籍配置表查询名称

---

#### 2.2 桌台轮询 - `GetDeskPing()`

**位置**: `main/app/service/desk.go:230`

**当前实现**:
```go
// 设置国籍UUID
res.NationalityUuid = desk.SaleBill.NationalityUuid
```

**响应结构**: `dto/resp/desk.go:114`
```go
type DeskPing struct {
    NationalityUuid uint64 `json:"nationality_uuid"` // 国籍UUID（仅UUID，无Name）
    // ...
}
```

**✅ 无需修改**：
- 仅返回 UUID，不返回名称
- 前端通过国籍配置表查询名称

---

### 3. 统计报表查询（特殊情况）

#### 3.1 用户分析统计 - `GetUserAnalysis()`

**位置**: `main/app/repository/statistics_user_analysis.go:56-59`

**当前实现**:
```sql
SELECT ss.nationality_uuid, 
       COALESCE(NULLIF(mln.zh_name, ''), NULLIF(mln.en_name, ''), 'Unknown') AS name,
       COUNT(DISTINCT ss.sale_bill_uuid) AS order_count
FROM statistics_sale AS ss
LEFT JOIN nationality AS n ON ss.nationality_uuid = n.uuid
LEFT JOIN multi_language_name AS mln ON n.multi_language_name_uuid = mln.uuid
WHERE ss.nationality_uuid > 0
GROUP BY ss.nationality_uuid, name
```

**特点**：
- 从 `statistics_sale` 统计表查询（非 `sale_bill`）
- 统计表没有快照字段
- 用于历史数据统计分析

**✅ 暂不修改**：
- 统计表设计独立，暂时保持现状
- 可在后续优化时添加快照字段到统计表

---

## 📋 修改计划

### Phase 4.2: 修改订单查询逻辑 - 使用 GetLocaleNationalityName()

#### 修改位置 1️⃣：订单详情查询

**文件**: `main/app/service/order_manage.go:691-696`

**方法**: `GetOrderInfos()`

**修改内容**：
```go
// 修改前
NationalityName: func() string {
    if saleBill.Nationality != nil {
        return saleBill.Nationality.MultiLanguageName.GetNameByLang(ctx.GetLanguage())
    }
    return ""
}(),

// 修改后
NationalityName: saleBill.GetLocaleNationalityName().GetByLang(ctx.GetLanguage()),
```

**优势**：
- ✅ 优先使用快照数据（JSON，包含所有语言）
- ✅ 降级使用关联表（兼容历史数据）
- ✅ 简化代码逻辑（无需判空）

---

## 🎯 下一步

继续执行 **Phase 4 - Task 4.2**: 修改订单查询逻辑 - 使用 GetLocaleNationalityName()

**修改文件**：
- `main/app/service/order_manage.go` - 订单详情查询

---

**最后更新**: 2025-12-02  
**任务**: story-main-nationality-snapshot-fix (JSON 方案)

