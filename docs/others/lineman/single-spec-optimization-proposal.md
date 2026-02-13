# 单选项 ModifierGroup 在 Grab/Lineman 同步时的优化方案

> 状态: 待评审
> 涉及模块: Main (takeout), BMP (ttpos-takeout)
> 影响平台: Grab, Lineman
> 创建时间: 2026-02-13
> 更新时间: 2026-02-13

---

## 1. 问题描述

### 1.1 背景

TTPOS 的商品使用 `ProductBom` + `ProductFlavor` 体系来定义**规格**(Spec)，例如 S/M/L 三种杯型。
同时还有加料(Sauce)、属性(Attribute)等修饰类型。
Grab 和 Lineman 没有"规格"概念，使用 `ModifierGroup` / `Property` 来表示可选项组。

当 TTPOS 推送菜单到 Grab/Lineman 时，会将所有修饰项（包括只有一个选项的）包装为 `ModifierGroup`：

```
ModifierGroup {
  name: "Specifications"
  selectionRangeMin: 1
  selectionRangeMax: 1
  modifiers: [ { name: "默认规格", price: 0 } ]  // 只有一个选项
}
```

### 1.2 用户体验问题

当 `selectionRangeMin == selectionRangeMax == len(modifiers) == 1` 时，用户被迫与一个**没有实质选择**的选项组交互。

| 类型 | 场景 | 用户看到 | 体验 |
|------|------|---------|------|
| 规格 | 多规格(S/M/L) | "Specifications" 选项组，3 个选项 | ✅ 合理 |
| 规格 | **单规格(默认)** | **"Specifications" 选项组，1 个选项，必选** | **❌ 多余操作** |
| 加料 | 多加料 | "Add Toppings" 选项组，N 个选项 | ✅ 合理 |
| 加料 | **单加料且 min=1** | **"Add Toppings" 选项组，1 个选项，必选** | **❌ 多余操作** |
| 属性 | 多属性 | "辣度" 选项组，3 个选项 | ✅ 合理 |
| 属性 | **单属性且 min=1** | **"辣度" 选项组，1 个选项，必选** | **❌ 多余操作** |

核心判断标准：**当 `min == max == len(options) == 1` 时，用户被强制选择唯一选项，纯粹增加操作步骤。**

### 1.3 为什么条件是 `min == max == len(options) == 1`

需要区分几种容易混淆的场景：

| 条件 | 含义 | 是否有实质选择 | 是否应跳过 |
|------|------|--------------|-----------|
| `min=1, max=1, options=1` | 必须选唯一项 | ❌ 无选择 | ✅ 跳过 |
| `min=0, max=1, options=1` | 可选唯一项 | ✅ 选或不选 | ❌ 保留 |
| `min=2, max=2, options=2` | 必须全选 2 项 | ❌ 无选择，但有信息价值（用户需看到各项名称和价格） | ❌ 保留 |
| `min=2, max=2, options=3` | 从 3 选 2 | ✅ 有选择 | ❌ 保留 |
| `min=1, max=1, options=3` | 从 3 选 1 | ✅ 有选择 | ❌ 保留 |

**为什么不泛化到 `min == max == len(options)` (N>1)？**

虽然 `min == max == len(options)` 时用户也没有取舍权，但当 N>1 时：
- 用户仍需要看到选项组的**信息展示**（各加料的名称、价格）
- 直接移除会导致用户完全不知道商品包含了什么附加项
- 而 N=1 时，唯一选项的信息可以合并到商品本身（名称、价格已在商品层面体现）

因此 `N == 1` 是唯一安全的优化边界。

### 1.4 为什么加料 `min=0`（可选）不应跳过

加料的 `SauceMinSelection` 可以为 0（可选），此时即使只有 1 个加料：

- `min=0, max=1, options=1` → 用户可以**选择加或不加**
- 如果跳过 ModifierGroup，用户**完全无法选择**该加料 → 功能缺失
- 可选的单加料 UX 是合理的：用户看到一个可展开的加料组，按需选择

只有 `min=1`（必选）时才存在"强制选唯一项"的 UX 问题。

### 1.5 为什么套餐固定组不参与优化

套餐固定组（`GroupType=0`）的 `min=max=len(items)` 是**业务设计意图**：
- 固定组要求必须包含所有子商品（如套餐必含 A+B+C）
- 即使只有 1 个子商品，用户也需要看到这个**子商品是什么**（名称、价格）
- 套餐子商品的信息无法简单合并到父商品层面

因此套餐固定组不参与此优化。

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

### 2.3 各转换函数的 min/max 逻辑

| 类型 | 函数 | min 来源 | max 来源 | 单选项时 min==max? |
|------|------|---------|---------|------------------|
| **规格** | `convertProductFlavors()` | 硬编码 `1` | 硬编码 `1` | ✅ 永远 `1==1` |
| **加料** | `convertProductSauces()` | `SauceMinSelection`（可为0） | `SauceMaxSelection`（未配置=加料总数） | 仅当 `SauceMinSelection=1` 时 |
| **属性** | `convertProductAttributeGroups()` | `MinSelection` | `MaxSelection`（未配置=属性总数） | 仅当 `MinSelection=1` 时 |
| **套餐** | `convertPackageGroups()` | 固定组:`len(items)`；可选组:`OptionalMinCount` | 固定组:`len(items)`；可选组:`OptionalCount` | 固定组永远是，但不参与优化 |

当前所有函数均**不检查选项数量**，只要存在 ≥1 个有效选项就创建 ModifierGroup。

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

### 4.1 直接移除单选项 ModifierGroup 的风险

如果仅在菜单侧移除 ModifierGroup 而不做订单侧处理：

#### 4.1.1 规格（Flavor） — 风险最高

| 影响 | 当前行为 | 移除后行为 | 风险等级 |
|------|---------|-----------|---------|
| 商品映射 | 正常 | 正常 | 无 |
| 规格映射 | modifier → BOM UUID | 无 modifier → **BOM UUID 丢失** | **严重** |
| 库存扣减 | 通过 BOM UUID 扣原料 | **无 BOM UUID，跳过原料扣减** | **严重** |
| 销量统计 | 记录 BOM 销量 | **不记录 BOM 销量** | **严重** |
| 订单显示 | 显示规格名称 | 不显示 | 可接受 |

#### 4.1.2 加料（Sauce） — 风险中等

| 影响 | 当前行为 | 移除后行为 | 风险等级 |
|------|---------|-----------|---------|
| 加料映射 | modifier → Sauce BOM UUID | 无 modifier → **加料缺失** | **中等** |
| 库存扣减 | 通过 Sauce BOM 扣原料 | **跳过加料原料扣减** | **中等** |
| 价格计算 | 加料价格计入订单 | **加料价格缺失** | **中等** |

注意：仅当 `SauceMinSelection=1` 且只有 1 个加料时才触发此优化。当 `SauceMinSelection=0`（可选）时不跳过。

#### 4.1.3 属性（Attribute） — 风险较低

| 影响 | 当前行为 | 移除后行为 | 风险等级 |
|------|---------|-----------|---------|
| 属性映射 | modifier → Attribute UUID | 无 modifier → 属性缺失 | 较低 |
| 价格计算 | 属性加价计入订单 | 属性加价缺失 | 较低 |
| 库存影响 | 属性通常不关联物料 | 无影响 | 无 |

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

### 5.1 优化判定条件

**统一跳过条件**：`min == max == len(options) == 1`

即：只有 1 个选项，且该选项为必选（min=1），只能选 1 个（max=1）。此时用户无实质选择。

适用的 Modifier 类型：

| 类型 | 跳过条件 | 适用场景 |
|------|---------|---------|
| 规格 (Flavor) | `len(flavors) == 1`（min/max 硬编码为 1，永远满足） | 单规格商品 |
| 加料 (Sauce) | `len(sauces) == 1 && SauceMinSelection == 1` | 单加料且必选 |
| 属性 (Attribute) | `len(attrs) == 1 && MinSelection == 1 && MaxSelection == 1` | 单属性且必选1个 |
| 套餐 (Package) | **不参与优化** | 见 1.5 节说明 |

### 5.2 方案概述

**两端协同修改**：菜单侧按条件跳过 + 订单侧按实时状态补全。

```
【菜单推送侧】
convertProductFlavors():
  len(flavors) == 1 → 跳过 ModifierGroup
  len(flavors) > 1  → 保持现有逻辑

convertProductSauces():
  len(sauces) == 1 && SauceMinSelection >= 1 → 跳过 ModifierGroup
  其他 → 保持现有逻辑

convertProductAttributeGroups():
  len(attrs) == 1 && MinSelection >= 1 && maxSelection == 1 → 跳过该 ModifierGroup
  其他 → 保持现有逻辑

【订单处理侧 — 以 flavor 为例】
收到订单 → 商品已映射 → 无 flavor modifier
    │
    ▼
查询当前 ProductBom (实时)
    │
    ├─ 仍然只有 1 个 flavor BOM → 自动补全该 BOM 作为 modifier
    │
    └─ 有多个 flavor BOM → 标记为异常订单（套用现有异常逻辑）

（加料和属性同理：缺失时查实时数据，仅 1 个则补全，多个则异常）
```

### 5.3 改动清单

#### 修改点 1: 菜单推送 — 跳过单选项 ModifierGroup

**文件**: `main/app/modules/takeout/infrastructure/adapter/grab/grab_menu_converter.go`

**1a. `convertProductFlavors()` — line 527 之后**

```go
if len(flavors) == 0 {
    return nil
}

// 新增：单规格优化 — 仅有一个规格时，价格已体现在商品基础价格中，
// 无需创建额外的 ModifierGroup（min=1,max=1,options=1 → 无实质选择）
if len(flavors) == 1 {
    return nil
}
```

安全性论证:
- 商品基础价格在 `convertTTPOSProduct()` line 291-306 中已使用 `minFlavorPrice`
- 单规格时 `minFlavorPrice` = 唯一规格的价格 = 商品展示价格
- 因此商品价格已正确反映唯一规格的售价

**1b. `convertProductSauces()` — line 669 之后**

```go
if len(sauces) == 0 {
    return nil
}

// 新增：单加料且必选优化 — 仅有一个加料且 SauceMinSelection >= 1 时，
// 满足 min==max==options==1 条件，无实质选择
if len(sauces) == 1 && takeoutProduct.ProductPackage.SauceMinSelection >= 1 {
    return nil
}
```

安全性论证:
- 仅在 `SauceMinSelection >= 1`（必选）时跳过
- `SauceMinSelection == 0`（可选）时保留，确保用户可以选择"加或不加"
- 跳过后加料价格需在订单侧补全

**1c. `convertProductAttributeGroups()` — 属性组循环内**

```go
// 新增：单属性且必选优化 — 属性组只有 1 个有效属性，且 min >= 1, max == 1 时
// 满足 min==max==options==1 条件，无实质选择
validAttrs := filterValidAttrs(packageAttrGroup.ProductPackageAttributes)
if len(validAttrs) == 1 && packageAttrGroup.MinSelection >= 1 && maxSelection == 1 {
    continue  // 跳过该属性组
}
```

安全性论证:
- 每个属性组独立判断，不影响其他属性组
- 仅在 `MinSelection >= 1`（必选）且 `maxSelection == 1` 时跳过
- 属性通常不关联物料，补全逻辑较简单

#### 修改点 2: 订单处理 — 缺失 modifier 自动补全

**文件**: `main/app/modules/takeout/domain/service/takeout_order_service.go`
**位置**: `CreateOrder()` 中 `enrichModifiersInfo()` 调用之后

新增方法 `autoFillSingleModifiers()`：

```
逻辑:
1. 遍历订单商品，筛选出"已映射 + 非套餐"的商品
2. 对每个商品检查缺失的 modifier 类型（flavor/sauce/attr）

【Flavor 补全】
  a. 无 flavor modifier 的商品 → 批量查询 flavor BOM
  b. 仅 1 个 flavor BOM → 自动创建 TakeoutOrderItemModifier (type="flavor")
  c. 多个 flavor BOM → 标记异常

【Sauce 补全】
  a. 无 sauce modifier 的商品 → 查询 sauce BOM + SauceMinSelection
  b. 仅 1 个 sauce BOM 且 SauceMinSelection >= 1 → 自动补全
  c. 多个 sauce BOM 且 SauceMinSelection >= 1 → 标记异常
  d. SauceMinSelection == 0 → 无需补全（可选项，用户未选择是正常的）

【Attribute 补全】
  a. 无某属性组 modifier 的商品 → 查询属性组配置
  b. 仅 1 个属性 且 MinSelection >= 1 → 自动补全
  c. 其他 → 标记异常或跳过
```

#### 修改点 3: Repository 层 — 新增查询方法

**文件**: `main/app/modules/takeout/infrastructure/persistence/takeout_bom_mapping_repo.go`

新增接口方法:
```go
// GetFlavorBomsByProductPackageUuids 批量查询商品的 flavor BOM 列表
// 返回: map[ProductPackageUuid][]ProductBom
GetFlavorBomsByProductPackageUuids(productPackageUuids []uint64) (map[uint64][]ProductBom, error)

// GetSauceBomsByProductPackageUuids 批量查询商品的 sauce BOM 列表
// 返回: map[ProductPackageUuid][]ProductBom
GetSauceBomsByProductPackageUuids(productPackageUuids []uint64) (map[uint64][]ProductBom, error)
```

实现:
```sql
-- Flavor BOMs
SELECT product_package_uuid, uuid, price, product_flavor_uuid
FROM ttpos_product_bom
WHERE product_package_uuid IN (?)
  AND product_flavor_uuid > 0
  AND delete_time = 0

-- Sauce BOMs
SELECT product_package_uuid, uuid, price, product_sauce_uuid
FROM ttpos_product_bom
WHERE product_package_uuid IN (?)
  AND product_sauce_uuid > 0
  AND delete_time = 0
```

### 5.4 不需要修改的部分

| 组件 | 原因 |
|------|------|
| BMP ttpos-takeout | Lineman 的菜单转换消费 Grab 格式快照，上游修复后自动生效 |
| Lineman 订单转换器 | 与 Grab 共用相同的 enrichModifiersInfo 逻辑 |
| ID 前缀解析器 | 不涉及 |
| 菜单快照结构 | 不需要额外字段 |

---

## 6. 场景覆盖验证

### 6.1 规格 (Flavor) 场景

| # | 推送时 | 下单时 TTPOS 状态 | 订单内容 | 处理结果 | 正确性 |
|---|--------|-----------------|---------|---------|--------|
| 1 | 单规格,无modifier | 仍 1 个 BOM | Modifiers=[] | 自动补全唯一 BOM | ✅ |
| 2 | 单规格,无modifier | 新增为多 BOM | Modifiers=[] | 标记异常 | ✅ 安全兜底 |
| 3 | 单规格,无modifier | 删除为 0 BOM | Modifiers=[] | 无需补全 | ✅ |
| 4 | 多规格,有modifier | 不变 | Modifiers有值 | 现有逻辑 | ✅ 不受影响 |
| 5 | 多规格,有modifier | 删了某规格 | Modifiers有旧ID | 现有异常逻辑 | ✅ 既有行为 |

### 6.2 加料 (Sauce) 场景

| # | 推送时 | 下单时 TTPOS 状态 | 订单内容 | 处理结果 | 正确性 |
|---|--------|-----------------|---------|---------|--------|
| 6 | 单加料必选(min=1),无modifier | 仍 1 个 sauce | Modifiers=[] | 自动补全唯一 sauce | ✅ |
| 7 | 单加料必选(min=1),无modifier | 新增为多 sauce | Modifiers=[] | 标记异常 | ✅ 安全兜底 |
| 8 | 单加料可选(min=0),有modifier | 用户选了 | Modifiers有sauce | 现有逻辑 | ✅ 不受影响 |
| 9 | 单加料可选(min=0),有modifier | 用户未选 | Modifiers=[] | 无需补全（可选） | ✅ 正常行为 |
| 10 | 多加料,有modifier | 不变 | Modifiers有值 | 现有逻辑 | ✅ 不受影响 |

### 6.3 属性 (Attribute) 场景

| # | 推送时 | 下单时 TTPOS 状态 | 订单内容 | 处理结果 | 正确性 |
|---|--------|-----------------|---------|---------|--------|
| 11 | 单属性必选(min=1,max=1),无modifier | 仍 1 个属性 | Modifiers=[] | 自动补全唯一属性 | ✅ |
| 12 | 单属性必选,无modifier | 新增为多属性 | Modifiers=[] | 标记异常 | ✅ 安全兜底 |
| 13 | 多属性,有modifier | 不变 | Modifiers有值 | 现有逻辑 | ✅ 不受影响 |

---

## 7. 测试要点

### 7.1 菜单推送测试 — 规格

- [ ] 单规格商品推送后 Grab 菜单无 Specifications ModifierGroup
- [ ] 单规格商品推送后 Lineman 菜单无 Specifications Property
- [ ] 单规格商品的展示价格 = 唯一规格的价格
- [ ] 多规格商品不受影响，仍有 ModifierGroup

### 7.2 菜单推送测试 — 加料

- [ ] 单加料 + `SauceMinSelection=1` → 无 Add Toppings ModifierGroup
- [ ] 单加料 + `SauceMinSelection=0` → **保留** Add Toppings ModifierGroup（可选，不应跳过）
- [ ] 多加料不受影响
- [ ] 加料价格在跳过后不影响商品展示价格（加料是独立加价）

### 7.3 菜单推送测试 — 属性

- [ ] 单属性 + `MinSelection=1, MaxSelection=1` → 无该属性组的 ModifierGroup
- [ ] 单属性 + `MinSelection=0`（可选）→ **保留** ModifierGroup
- [ ] 多属性不受影响

### 7.4 订单处理测试

- [ ] 单规格商品订单回传（无modifier）→ 自动补全 BOM → 库存正常扣减
- [ ] 单加料必选商品订单回传（无modifier）→ 自动补全 sauce → 价格正确
- [ ] 竞争窗口场景（推送时单项，下单时多项）→ 标记异常
- [ ] 多规格/多加料商品订单回传（有modifier）→ 现有逻辑不受影响
- [ ] 销量统计：补全后的 flavor modifier 正确计入 BOM 销量
- [ ] 原料消耗：补全后正确通过 BOM UUID 查找配方并扣减库存

### 7.5 边界场景

- [ ] 商品无规格（0 个 BOM）→ 无 modifier 也无需补全
- [ ] 商品所有规格已售罄 → 商品状态为 UNAVAILABLE，正常
- [ ] 商品唯一规格已删除 → 0 个有效 BOM，走无规格逻辑
- [ ] 加料 `SauceMinSelection=0` 且用户未选 → 无 modifier，**不应补全**（可选项未选是正常的）
- [ ] 属性组有多个属性组，其中仅 1 个符合跳过条件 → 仅跳过该组，其他不受影响

---

## 8. 影响评估

| 维度 | 评估 |
|------|------|
| 用户体验 | ✅ 消除所有"必选唯一项"的无意义交互（规格/加料/属性） |
| 数据一致性 | ✅ 订单侧按实时状态补全，无歧义时自动处理，有歧义时安全兜底 |
| 性能影响 | ⚠️ 订单处理新增轻量 DB 查询（按 product_package_uuid），可批量化 |
| 改动范围 | ✅ 菜单侧 3 个函数 + 订单侧 1 个补全方法 + Repository 层 2 个查询方法，均在 Main 模块 |
| 向后兼容 | ✅ 多选项商品完全不受影响；可选（min=0）的单选项保留展示 |

## 9. Grab/Lineman API 兼容性确认

| 平台 | 单选项 ModifierGroup | min=0（可选）| min=1 单选项 | 依据 |
|------|---------------------|-------------|-------------|------|
| **Grab** | 允许 — SDK 无 `minItems` 约束 | 允许 | 允许 | GrabFood SDK v1.1.3 `ModifierGroup` 定义 |
| **Lineman** | 允许 — API Spec 无 values 最小数量约束 | 允许 | 允许 | Lineman Partner API Spec |

两个平台都不会因 ModifierGroup 只有 1 个选项而拒绝请求。本优化是纯 UX 改进，不涉及 API 兼容性问题。
