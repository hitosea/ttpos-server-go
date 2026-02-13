# 单规格商品在 Grab/Lineman 同步时的优化方案

> 状态: 待评审
> 涉及模块: Main (takeout), BMP (ttpos-takeout)
> 影响平台: Grab, Lineman
> 创建时间: 2026-02-13

---

## 1. 问题描述

### 1.1 背景

TTPOS 的商品使用 `ProductBom` + `ProductFlavor` 体系来定义**规格**(Spec)，例如 S/M/L 三种杯型。
Grab 和 Lineman 没有"规格"概念，使用 `ModifierGroup` / `Property` 来表示可选项组。

当 TTPOS 推送菜单到 Grab/Lineman 时，会将所有规格（包括只有一个的）包装为一个 `ModifierGroup`：

```
ModifierGroup {
  name: "Specifications"
  selectionRangeMin: 1
  selectionRangeMax: 1
  modifiers: [ { name: "默认规格", price: 0 } ]  // 只有一个选项
}
```

### 1.2 用户体验问题

| 场景 | 用户看到 | 体验 |
|------|---------|------|
| 多规格(S/M/L) | "Specifications" 选项组，3 个选项 | ✅ 合理 |
| **单规格(默认)** | **"Specifications" 选项组，1 个选项** | **❌ 多余操作** |

单规格时用户被迫打开一个只有 1 个选项的选项组，点击唯一的选项，价格差为 0——纯粹增加操作步骤。

---

## 2. 现有菜单同步架构

### 2.1 完整调用链

```
【TTPOS → Grab 推送方向】
Main: GrabConverter.LoadMenuFromDatabase()
  → loadCategoryProducts()
  → convertTTPOSProduct()
    → convertProductFlavors()     ← 规格 → ModifierGroup
    → convertProductSauces()      ← 加料 → ModifierGroup
    → convertProductAttributeGroups() ← 属性 → ModifierGroup
    → convertPackageGroups()      ← 套餐组 → ModifierGroup
  → takeoutAppSrv.PushMenu()
  → BMP RPC: SaveMenuSnapshot()
  → BMP: asyncNotifyGrabMenuUpdate()
  → Grab API: NotifyMenuUpdate

【TTPOS → Lineman 推送方向】
Main: GrabConverter.LoadMenuFromDatabase() → 生成 Grab 格式 JSON
  → BMP RPC: SaveMenuSnapshot()
  → BMP: sLineman.SyncMenu()
    → convertGrabToLinemanMenuGroups()
      → convertGrabItemToLinemanItem()
        → 遍历 ModifierGroups → convertGrabModifierGroupToLinemanProperty()
    → Lineman API: SyncMenu
```

### 2.2 关键代码位置

| 功能 | 文件路径 | 函数 |
|------|---------|------|
| 规格→ModifierGroup | `main/.../grab/grab_menu_converter.go:490` | `convertProductFlavors()` |
| 加料→ModifierGroup | `main/.../grab/grab_menu_converter.go:653` | `convertProductSauces()` |
| 属性→ModifierGroup | `main/.../grab/grab_menu_converter.go:771` | `convertProductAttributeGroups()` |
| 套餐→ModifierGroup | `main/.../grab/grab_menu_converter.go:877` | `convertPackageGroups()` |
| Grab→Lineman 转换 | `ttpos-bmp/.../lineman/menu_sync.go:220` | `convertGrabToLinemanMenuGroups()` |
| 菜单快照保存 | `ttpos-bmp/.../channel_menu/channel_menu.go:101` | `SaveMenuSnapshot()` |

### 2.3 当前 `convertProductFlavors()` 核心逻辑

```go
// grab_menu_converter.go:527-529
// 只要有 ≥1 个 flavor 就创建 ModifierGroup，不区分单/多规格
if len(flavors) == 0 {
    return nil
}

// 固定 min=1, max=1
minSelection := 1
maxSelection := 1

// 无条件追加 ModifierGroup
menuItem.SetModifierGroups(append(menuItem.GetModifierGroups(), *modifierGroup))
```

---

## 3. 订单回传识别机制

### 3.1 ID 前缀映射表

TTPOS 在推送菜单时为每个实体生成带前缀的 ID，订单回传时通过前缀解析反向映射。

| 前缀 | 含义 | UUID 来源 | 回传映射类型 |
|------|------|-----------|------------|
| `TTPOS-ITEM-{uuid}` | 普通商品 | ProductPackage.Uuid | IDTypeItem |
| `TTPOS-PACKAGE-{uuid}` | 套餐商品 | ProductPackage.Uuid | IDTypePackage |
| `TTPOS-FLAVOR-{uuid}` | **规格选项** | **ProductBom.Uuid** | IDTypeFlavor → `ModifierTypeFlavor` |
| `TTPOS-SAUCE-{uuid}` | 加料选项 | ProductBom.Uuid | IDTypeSauce → `ModifierTypeSauce` |
| `TTPOS-ATTR-{uuid}` | 属性选项 | ProductAttribute.Uuid | IDTypeAttr → `ModifierTypeAttr` |
| `TTPOS-FLAVOR-GROUP-{uuid}` | 规格组 | ProductPackage.Uuid | IDTypeFlavorGroup |
| `TTPOS-SAUCE-GROUP-{uuid}` | 加料组 | ProductPackage.Uuid | IDTypeSauceGroup |
| `TTPOS-ATTR-GROUP-{uuid}` | 属性组 | ProductPkgAttrGroup.Uuid | IDTypeAttrGroup |
| `TTPOS-PACKAGE-GROUP-{uuid}` | 套餐组 | ProductPkgGroup.Uuid | IDTypePackageGroup |
| `TTPOS-PACKAGE-ITEM-{uuid}` | 套餐子项 | ProductPkgGroupItem.Uuid | IDTypePackageItem |

### 3.2 订单回传处理链路

```
Grab/Lineman 订单 Webhook
    │
    ▼
BMP: HandleSubmitOrder() → 保存原始 JSON → 发送 RocketMQ 事件
    │
    ▼
Main: takeoutAppSrv.HandleOrderEvent()
    │
    ▼
GrabOrderConverter/LinemanOrderConverter.ConvertToTTPOS()
    │  解析:
    │  item.Id = "TTPOS-ITEM-123456"        → PlatformItemId
    │  item.Modifiers[].Id = "TTPOS-FLAVOR-789"  → PlatformModifierId
    │
    ▼
takeoutOrderSrv.CreateOrder()
    ├─ enrichItemInfo()
    │   └─ findProductMapping()
    │       └─ ParsePlatformID("TTPOS-ITEM-123456")
    │           → IDType=item, UUID=123456
    │
    └─ enrichModifiersInfo()
        └─ findModifierMapping()
            └─ ParsePlatformID("TTPOS-FLAVOR-789")
                → IDType=flavor, UUID=789 (= ProductBom.Uuid)
                → TtposModifierType = "flavor"
```

### 3.3 Flavor Modifier 在下游的使用

`modifier.TtposModifierUuid`（即 `ProductBom.Uuid`）直接用于：

| 使用场景 | 位置 | 依赖方式 |
|---------|------|---------|
| **库存扣减** | `extractAndSaveTakeoutOrderMaterialsWithMetadata()` | BOM UUID → 查配方 → 计算原料消耗 |
| **销量统计** | `CalculateTakeoutOrderSalesVolume()` | `productBoms[bomUuid] += quantity` |
| **订单显示** | `enrichModifiersInfo()` | modifier 名称展示 |

---

## 4. 风险分析

### 4.1 直接移除单规格 ModifierGroup 的风险

如果仅在菜单侧移除单规格的 ModifierGroup 而不做订单侧处理：

| 影响 | 当前行为 | 移除后行为 | 风险等级 |
|------|---------|-----------|---------|
| 商品映射 | 正常 | 正常 | 无 |
| 规格映射 | modifier → BOM UUID | 无 modifier → **BOM UUID 丢失** | **严重** |
| 库存扣减 | 通过 BOM UUID 扣原料 | **无 BOM UUID，跳过原料扣减** | **严重** |
| 销量统计 | 记录 BOM 销量 | **不记录 BOM 销量** | **严重** |
| 订单显示 | 显示规格名称 | 不显示 | 可接受 |

### 4.2 时序竞争场景

```
T1: TTPOS 商品有 1 个规格(S, BOM-100)
    → 菜单推送: 无 ModifierGroup（单规格优化后）
    → 商品价格 = S 的价格

T2: 商户在 TTPOS 增加规格 M(BOM-200)、L(BOM-300)
    → Grab/Lineman 菜单尚未重新推送

T3: 用户在 Grab/Lineman 下单（使用旧菜单）
    → item.Modifiers = []（空）

T4: TTPOS 收到订单
    → 商品映射成功
    → 无 flavor modifier
    → ❓ 此时有 3 个 BOM，无法确定应使用哪个
```

### 4.3 既有系统的类似风险（非新增）

当前系统在以下场景下也存在类似问题（与本次优化无关）：

- 商户在 TTPOS 删除某规格后未重新同步菜单，Grab 订单仍带已删除规格的 modifier ID
- 商户修改规格价格后未重新同步，Grab/Lineman 菜单展示旧价格

这些本质上都是**分布式系统最终一致性**问题，菜单同步目前是手动触发（`PushMenuToPlatform`），不可避免存在不一致窗口。

---

## 5. 解决方案

### 5.1 方案概述

**两端协同修改**：菜单侧隐藏 + 订单侧按实时状态补全。

```
【菜单推送侧】
convertProductFlavors():
  len(flavors) == 1 → 跳过 ModifierGroup 创建
  len(flavors) > 1  → 保持现有逻辑

【订单处理侧】
收到订单 → 商品已映射 → 无 flavor modifier
    │
    ▼
查询当前 ProductBom (实时)
    │
    ├─ 仍然只有 1 个 flavor BOM → 自动补全该 BOM 作为 modifier
    │
    └─ 有多个 flavor BOM → 标记为异常订单（套用现有异常逻辑）
```

### 5.2 改动清单

#### 修改点 1: 菜单推送 — 跳过单规格 ModifierGroup

**文件**: `main/app/modules/takeout/infrastructure/adapter/grab/grab_menu_converter.go`
**函数**: `convertProductFlavors()`
**位置**: line 527 之后

```go
if len(flavors) == 0 {
    return nil
}

// 新增：单规格优化 — 仅有一个规格时，价格已体现在商品基础价格中，
// 无需创建额外的 ModifierGroup
if len(flavors) == 1 {
    return nil
}
```

**安全性论证**:
- 商品基础价格在 `convertTTPOSProduct()` line 291-306 中已使用 `minFlavorPrice`
- 单规格时 `minFlavorPrice` = 唯一规格的价格 = 商品展示价格
- 因此商品价格已正确反映唯一规格的售价

#### 修改点 2: 订单处理 — 单规格自动补全

**文件**: `main/app/modules/takeout/domain/service/takeout_order_service.go`
**位置**: `CreateOrder()` 中 `enrichModifiersInfo()` 调用之后

新增方法 `autoFillSingleFlavorModifier()`：

```
逻辑:
1. 遍历订单商品，筛选出"已映射 + 无 flavor modifier + 非套餐"的商品
2. 批量查询这些商品的 flavor BOM:
   SELECT uuid FROM ttpos_product_bom
   WHERE product_package_uuid = ? AND product_flavor_uuid > 0 AND delete_time = 0
3. 对于只有 1 个 flavor BOM 的商品:
   → 自动创建一条 TakeoutOrderItemModifier (TtposModifierType="flavor")
4. 对于有多个 flavor BOM 的商品:
   → 标记为异常（现有逻辑）
```

#### 修改点 3: Repository 层 — 新增查询方法

**文件**: `main/app/modules/takeout/infrastructure/persistence/takeout_bom_mapping_repo.go`

新增接口方法:
```go
// GetFlavorBomUuidsByProductPackageUuids 批量查询商品的 flavor BOM UUID 列表
// 返回: map[ProductPackageUuid][]BomUuid
GetFlavorBomUuidsByProductPackageUuids(productPackageUuids []uint64) (map[uint64][]uint64, error)
```

实现:
```sql
SELECT product_package_uuid, uuid
FROM ttpos_product_bom
WHERE product_package_uuid IN (?)
  AND product_flavor_uuid > 0
  AND delete_time = 0
```

### 5.3 不需要修改的部分

| 组件 | 原因 |
|------|------|
| BMP ttpos-takeout | Lineman 的菜单转换消费 Grab 格式快照，上游修复后自动生效 |
| Lineman 订单转换器 | 与 Grab 共用相同的 enrichModifiersInfo 逻辑 |
| ID 前缀解析器 | 不涉及 |
| 菜单快照结构 | 不需要额外字段 |

---

## 6. 场景覆盖验证

| # | 推送时 | 下单时 TTPOS 状态 | 订单内容 | 处理结果 | 正确性 |
|---|--------|-----------------|---------|---------|--------|
| 1 | 单规格,无modifier | 仍 1 个 BOM | Modifiers=[] | 自动补全唯一 BOM | ✅ |
| 2 | 单规格,无modifier | 新增为多 BOM | Modifiers=[] | 标记异常 | ✅ 安全兜底 |
| 3 | 单规格,无modifier | 删除为 0 BOM | Modifiers=[] | 无需补全 | ✅ |
| 4 | 多规格,有modifier | 不变 | Modifiers有值 | 现有逻辑 | ✅ 不受影响 |
| 5 | 多规格,有modifier | 删了某规格 | Modifiers有旧ID | 现有异常逻辑 | ✅ 既有行为 |

---

## 7. 测试要点

### 7.1 菜单推送测试

- [ ] 单规格商品推送后 Grab 菜单无 ModifierGroup
- [ ] 单规格商品推送后 Lineman 菜单无 Property
- [ ] 单规格商品的展示价格 = 唯一规格的价格
- [ ] 多规格商品不受影响，仍有 ModifierGroup
- [ ] 加料(Sauce)、属性(Attr)不受影响

### 7.2 订单处理测试

- [ ] 单规格商品订单回传（无modifier）→ 自动补全 BOM → 库存正常扣减
- [ ] 竞争窗口场景（推送时单规格，下单时多规格）→ 标记异常
- [ ] 多规格商品订单回传（有modifier）→ 现有逻辑不受影响
- [ ] 销量统计：补全后的 flavor modifier 正确计入 BOM 销量
- [ ] 原料消耗：补全后正确通过 BOM UUID 查找配方并扣减库存

### 7.3 边界场景

- [ ] 商品无规格（0 个 BOM）→ 无 modifier 也无需补全
- [ ] 商品所有规格已售罄 → 商品状态为 UNAVAILABLE，正常
- [ ] 商品唯一规格已删除 → 0 个有效 BOM，走无规格逻辑

---

## 8. 影响评估

| 维度 | 评估 |
|------|------|
| 用户体验 | ✅ 单规格商品下单流程简化，减少 1 次无意义点击 |
| 数据一致性 | ✅ 订单侧按实时状态补全，无歧义时自动处理，有歧义时安全兜底 |
| 性能影响 | ⚠️ 订单处理新增一次轻量 DB 查询（按 product_package_uuid），可批量化 |
| 改动范围 | ✅ 3 个修改点，均在 Main 模块内，不涉及 BMP 改动 |
| 向后兼容 | ✅ 多规格商品完全不受影响 |
