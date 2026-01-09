# LINE MAN vs Grab 接口对比与 TTPOS 接入可行方案

> 基于 Lineman Partner Integration Workflow V2 和 API_SPEC 文档

---

## 一、接口数量对比总览

| 维度         | Grab | LINE MAN | 说明                 |
| ------------ | ---- | -------- | -------------------- |
| **总接口数** | 15+  | 11       | Grab 功能更全面      |
| **必选接口** | 6+   | 4        | Lineman 核心接口更少 |
| **可选接口** | 9+   | 7        | Grab 扩展能力更强    |

### LINE MAN API 列表

| 序号 | API                               | 方向         | 必选 | 端点                                                  |
| ---- | --------------------------------- | ------------ | ---- | ----------------------------------------------------- |
| 1    | Authentication                    | Partner ↔ LM | ✅   | `POST /v1/oauth/token`                                |
| 2    | Menu Sync (v1/v2)                 | Partner → LM | ✅   | `PUT /v1/partners/{partnerId}/stores/{storeId}/menus` |
| 3    | Trigger Sync Menu                 | LM → Partner | ✅   | `POST .../menus/trigger-sync` (Webhook)               |
| 4    | Place Order Notification          | LM → Partner | ✅   | `POST .../orders` (Webhook)                           |
| 5    | Menu Sync Notification            | LM → Partner | ⚪   | `POST .../menus/notification` (Webhook)               |
| 6    | Update Menu Item Status           | Partner → LM | ⚪   | `PUT .../menu/items/status`                           |
| 7    | Update Menu Property Value Status | Partner → LM | ⚪   | `PUT .../menu/property/values/status`                 |
| 8    | Order Status Update Notification  | LM → Partner | ⚪   | `POST .../order/status` (Webhook)                     |
| 9    | Order Update Notification         | LM → Partner | ⚪   | `PUT .../orders` (Webhook)                            |
| 10   | Force Close/Open Restaurant       | Partner → LM | ⚪   | `PUT .../restaurant/availability`                     |

---

## 二、详细接口映射对比

### 2.1 认证模块

| 功能          | Grab                      | LINE MAN                  | 对比结论        |
| ------------- | ------------------------- | ------------------------- | --------------- |
| **认证方式**  | OAuth2 Client Credentials | OAuth2 Client Credentials | ✅ **完全相同** |
| **IP 白名单** | 支持                      | 支持（可选）              | ✅ **完全相同** |

### 2.2 菜单模块

| 功能             | Grab 接口                   | LINE MAN 接口                      | TTPOS 已实现 | 对比结论                        |
| ---------------- | --------------------------- | ---------------------------------- | ------------ | ------------------------------- |
| **菜单推送**     | `UpdateMenuV2`              | `Menu sync` (v1/v2)                | ✅           | ✅ **可复用**，方向一致         |
| **菜单拉取**     | `GetMenu` (Grab 主动拉取)   | ❌ 不支持                          | ✅           | ⚠️ **Lineman 不支持**，只能推送 |
| **触发同步**     | ❌ 由 TTPOS 主动推送        | `Trigger sync menu` (必选)         | ❌           | 🆕 **需新增** Webhook 接收端点  |
| **同步结果通知** | `MenuSyncWebhook`           | `Menu sync notification`           | ✅           | ✅ **可复用**，逻辑相同         |
| **单品状态更新** | `UpdateMenuItem` (批量)     | `Update menu item status`          | ✅           | ⚠️ **需适配**，状态值不同       |
| **选项状态更新** | `UpdateMenuModifier` (批量) | `Update menu propertyValue status` | ✅           | ⚠️ **需适配**，状态值不同       |

**Lineman Menu Sync v2 新增字段**（支持 Delivery/Pickup 渠道差异化）：

| 字段                                 | 类型    | 说明                         |
| ------------------------------------ | ------- | ---------------------------- |
| `salesChannelsAvailability.delivery` | boolean | 是否在配送渠道可售           |
| `salesChannelsAvailability.pickup`   | boolean | 是否在自取渠道可售           |
| `salesChannelsPrice.pickup`          | double  | 自取价格（可与配送价格不同） |

**菜单状态值对比**：

| TTPOS/Grab                     | LINE MAN         | LINE MAN 属性状态 | 说明                   |
| ------------------------------ | ---------------- | ----------------- | ---------------------- |
| `availableStatus: AVAILABLE`   | `AVAILABLE`      | `1`               | 可售                   |
| `availableStatus: UNAVAILABLE` | `SOLD_OUT_TODAY` | `2`               | 售罄（次日自动恢复）   |
| ❌                             | `SUSPENDED`      | `3`               | 暂停销售（需手动恢复） |

**菜单同步注意事项**：

- ⚠️ 避免在高峰时段同步菜单：**10:00-14:00** 和 **17:00-19:00**
- 促销菜单通过 `[Promotion]` 前缀识别，普通菜品禁止使用此前缀
- 单品状态更新最大批量：**100 条/次**

### 2.3 订单模块

| 功能             | Grab 接口                     | LINE MAN 接口                       | TTPOS 已实现 | 对比结论                       |
| ---------------- | ----------------------------- | ----------------------------------- | ------------ | ------------------------------ |
| **新订单推送**   | `SubmitOrder` Webhook         | `Place order notification`          | ✅           | ✅ **可复用**，需适配数据格式  |
| **订单状态推送** | `PushOrderState` Webhook      | `Status update notification` (可选) | ✅           | ✅ **可复用**，状态值需映射    |
| **订单编辑推送** | ❌ Grab 不支持订单编辑        | `Order update notification` (可选)  | ❌           | 🆕 **Lineman 特有**，需新增    |
| **接单/拒单**    | `PrepareOrder` (TTPOS→Grab)   | ❌ 不支持（需 WMA App）             | ✅           | ⚠️ **Lineman 不支持** POS 接单 |
| **标记准备完成** | `MarkOrderReady` (TTPOS→Grab) | ❌ 可选（WMA App 操作）             | ✅           | ⚠️ **Lineman 不支持**          |
| **取消订单**     | `CancelOrder` (TTPOS→Grab)    | ❌ 不支持（需 WMA App）             | ✅           | ⚠️ **Lineman 不支持** POS 取消 |
| **检查可取消**   | `CheckOrderCancelable`        | ❌ 不支持                           | ✅           | ⚠️ **Lineman 不支持**          |

**LINE MAN 订单格式关键字段**：

| 字段                | 类型        | 说明                                                   |
| ------------------- | ----------- | ------------------------------------------------------ |
| `orderId`           | String(20)  | 格式：`LMF-yyMMdd-{number}`，如 `LMF-221031-338798091` |
| `orderShortCode`    | String(4)   | 订单后 4 位，用于店内呼叫                              |
| `restaurantRevenue` | double      | 餐厅实际收入（已扣除平台补贴折扣）                     |
| `orderAcceptedTime` | String      | ISO 8601 格式，如 `2022-11-01T13:08:06+07:00`          |
| `customerType`      | String(32)  | `DELIVERY` 或 `PICKUP`                                 |
| `memberId`          | String(255) | 顾客在商家系统的会员 ID（可选）                        |

**订单状态更新**（Lineman 仅推送 `FINISH` 或 `CANCELED`）：

| Lineman 状态 | TTPOS 映射 | 说明       |
| ------------ | ---------- | ---------- |
| `FINISH`     | 已完成(40) | 订单完成   |
| `CANCELED`   | 已取消(50) | 订单被取消 |

> ⚠️ **重要**：Lineman 平板（WMA App）仍需保留作为备份，用于处理取消、编辑等操作。

### 2.4 门店模块

| 功能             | Grab 接口                   | LINE MAN 接口                        | TTPOS 已实现 | 对比结论                    |
| ---------------- | --------------------------- | ------------------------------------ | ------------ | --------------------------- |
| **门店绑定**     | `CreateSelfServeJourney`    | ❌ 不支持                            | ✅           | ⚠️ **Lineman 绑定方式不同** |
| **集成状态回调** | `IntegrationStatus` Webhook | ❌ 不支持                            | ✅           | ⚠️ **Lineman 无此机制**     |
| **门店配置查询** | `GetShopProviderCfg`        | ❌ 不支持                            | ✅           | ⚠️ **需另外实现**           |
| **门店开关控制** | ❌ 不支持                   | `Force close/open restaurant` (可选) | ❌           | 🆕 **Lineman 特有**，可新增 |

**LINE MAN 门店上线流程**（SLA: 4 工作日）：

```
1. Partner 提供 Store ID → 2. LINE MAN 绑定配置 → 3. Partner 推送菜单 → 4. 启用集成
```

**Force Close/Open 详细逻辑**：

| 场景                   | 当前时间 | 营业时间   | 执行操作        | 结果                       |
| ---------------------- | -------- | ---------- | --------------- | -------------------------- |
| Force close (status=2) | 10:00    | 9:00-15:00 | 强制关店        | 关闭至明天 9:00            |
| Force open (status=1)  | 8:00     | 9:00-15:00 | 提前开店        | 开放至今天 15:00           |
| Force open (status=1)  | 16:00    | 9:00-15:00 | 营业后开店      | 开放至明天 15:00           |
| Force open + duration  | 6:00     | 9:00-15:00 | 临时开店 1 小时 | 开放至 7:00，9:00 正常营业 |

**请求参数**：

- `status`: 1=强制开店，2=强制关店
- `duration`: 持续时间（秒），0=持续到下次营业时间结束

---

## 三、核心差异分析

### 3.1 架构模式差异

```
┌────────────────────────────────────────────────────────────────────┐
│                        Grab 模式 (双向拉取+推送)                    │
├────────────────────────────────────────────────────────────────────┤
│                                                                    │
│   TTPOS ──────推送菜单────────> Grab                               │
│   TTPOS <─────拉取菜单───────── Grab (Grab 主动 GetMenu)           │
│   TTPOS <─────订单推送───────── Grab                               │
│   TTPOS ──────接单/拒单──────> Grab                                │
│   TTPOS ──────准备完成──────-> Grab                                │
│   TTPOS ──────取消订单──────-> Grab                                │
│                                                                    │
└────────────────────────────────────────────────────────────────────┘

┌────────────────────────────────────────────────────────────────────┐
│                      Lineman 模式 (单向推送)                        │
├────────────────────────────────────────────────────────────────────┤
│                                                                    │
│   TTPOS ──────推送菜单────────> Lineman                            │
│   TTPOS <─────触发同步───────── Lineman (Lineman 请求重新推送)     │
│   TTPOS <─────订单推送───────── Lineman                            │
│   TTPOS <─────订单状态───────── Lineman                            │
│   TTPOS ──────门店开关────────> Lineman (可选)                     │
│                                                                    │
│   ⚠️ 无接单/拒单/准备完成等反向操作能力                            │
│   ⚠️ 订单由 Lineman 平板或自动处理                                 │
│                                                                    │
└────────────────────────────────────────────────────────────────────┘
```

### 3.2 关键差异总结

| 差异点           | Grab                   | LINE MAN           | 影响                           |
| ---------------- | ---------------------- | ------------------ | ------------------------------ |
| **菜单获取方式** | 双向（推送+拉取）      | 单向（仅推送）     | Lineman 不会主动拉菜单         |
| **订单控制能力** | TTPOS 可接单/拒单/取消 | TTPOS 只能接收通知 | **订单流程简化，但灵活性降低** |
| **门店绑定**     | 自助激活链接           | 需人工配置         | **Lineman 接入流程不同**       |
| **订单编辑**     | 不支持                 | 支持               | **Lineman 特有能力**           |
| **门店营业控制** | 不支持                 | 支持               | **Lineman 特有能力**           |

---

## 四、TTPOS 已有 Grab 功能的 Lineman 支持情况

### 4.1 ✅ 可直接复用的功能

| 功能         | 复用程度 | 说明           |
| ------------ | -------- | -------------- |
| 菜单推送     | 90%      | 数据结构需适配 |
| 订单接收     | 80%      | 需新增转换器   |
| 订单状态通知 | 80%      | 状态值需映射   |
| 同步结果回调 | 90%      | 逻辑相同       |

### 4.2 ⚠️ 无法使用的 Grab 功能（Lineman 不支持）

| Grab 功能                         | 说明                     | 建议处理方式       |
| --------------------------------- | ------------------------ | ------------------ |
| **菜单拉取 (GetMenu)**            | Lineman 不会主动获取菜单 | 忽略，使用推送模式 |
| **接单/拒单 (PrepareOrder)**      | Lineman 由平板或自动处理 | 降级为只读通知     |
| **标记准备完成 (MarkOrderReady)** | Lineman 不支持           | 降级为只读通知     |
| **取消订单 (CancelOrder)**        | TTPOS 无法主动取消       | 降级为只读通知     |
| **自助门店绑定**                  | Lineman 需人工配置       | 提供配置界面       |

### 4.3 🆕 Lineman 特有功能（需新增）

| Lineman 功能                    | 优先级 | 说明                      |
| ------------------------------- | ------ | ------------------------- |
| **Trigger sync menu**           | 必选   | 接收 Lineman 触发同步请求 |
| **Order update notification**   | 可选   | 处理订单编辑              |
| **Force close/open restaurant** | 可选   | POS 控制门店营业状态      |
| **SUSPENDED 状态**              | 可选   | 菜品暂停销售状态          |

---

## 五、Lineman 接入可行方案

### 5.1 整体架构设计

```
┌────────────────────────────────────────────────────────────────────┐
│                         LINE MAN Platform                          │
└───────────────────────────────────┬────────────────────────────────┘
                                    │
        ┌───────────────────────────┼───────────────────────────┐
        │                           │                           │
        ▼                           ▼                           ▼
  ┌───────────┐             ┌───────────┐              ┌───────────┐
  │ 触发同步   │             │ 订单推送   │              │ 订单状态   │
  │ (Webhook) │             │ (Webhook) │              │ (Webhook) │
  └─────┬─────┘             └─────┬─────┘              └─────┬─────┘
        │                         │                          │
        └─────────────────────────┼──────────────────────────┘
                                  │
                                  ▼
┌────────────────────────────────────────────────────────────────────┐
│                         ttpos-bmp (中间层)                          │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │  internal/controller/lineman/   (新增 Lineman Controller)    │  │
│  │  - lineman_v1_trigger_sync.go        ← 触发菜单同步          │  │
│  │  - lineman_v1_place_order.go         ← 订单推送              │  │
│  │  - lineman_v1_status_update.go       ← 订单状态更新          │  │
│  │  - lineman_v1_order_update.go        ← 订单编辑 (可选)       │  │
│  │  - lineman_v1_menu_sync_notify.go    ← 同步结果通知          │  │
│  └──────────────────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │  internal/logic/lineman/            (新增 Lineman Logic)     │  │
│  │  - lineman.go                        ← 核心服务              │  │
│  │  - lineman_menu.go                   ← 菜单处理              │  │
│  │  - lineman_order.go                  ← 订单处理              │  │
│  │  - lineman_auth.go                   ← OAuth 认证            │  │
│  └──────────────────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │  internal/client/lineman/           (新增 Lineman Client)    │  │
│  │  - client.go                         ← HTTP 客户端           │  │
│  │  - auth.go                           ← Token 管理            │  │
│  └──────────────────────────────────────────────────────────────┘  │
└───────────────────────────────────┬────────────────────────────────┘
                                    │ RocketMQ
                                    ▼
┌────────────────────────────────────────────────────────────────────┐
│                         ttpos-server-go (Main)                     │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │  app/modules/takeout/infrastructure/adapter/lineman/         │  │
│  │  - lineman_menu_converter.go         ← 菜单格式转换          │  │
│  │  - lineman_order_converter.go        ← 订单格式转换          │  │
│  └──────────────────────────────────────────────────────────────┘  │
└────────────────────────────────────────────────────────────────────┘
```

### 5.2 开发任务清单

#### Phase 1: 必选功能（MVP）

| 任务                      | 模块       | 工作量 | 优先级 |
| ------------------------- | ---------- | ------ | ------ |
| 1. Lineman OAuth2 认证    | BMP        | 2d     | P0     |
| 2. 菜单推送 (Menu sync)   | BMP + Main | 3d     | P0     |
| 3. 触发同步 Webhook       | BMP        | 1d     | P0     |
| 4. 订单接收 (Place order) | BMP + Main | 3d     | P0     |
| 5. Lineman 订单格式转换器 | Main       | 2d     | P0     |
| 6. Lineman 菜单格式转换器 | Main       | 2d     | P0     |

**Phase 1 预估**: **13 人天**

#### Phase 2: 可选功能

| 任务                | 模块       | 工作量 | 优先级 |
| ------------------- | ---------- | ------ | ------ |
| 7. 订单状态通知     | BMP + Main | 1d     | P1     |
| 8. 菜单同步结果通知 | BMP        | 0.5d   | P1     |
| 9. 单品状态更新     | BMP + Main | 1d     | P2     |
| 10. 选项状态更新    | BMP + Main | 1d     | P2     |
| 11. 订单编辑通知    | BMP + Main | 1.5d   | P2     |
| 12. 门店开关控制    | BMP + Main | 1d     | P3     |

**Phase 2 预估**: **6 人天**

### 5.3 代码复用策略

```go
// Main 模块：复用现有架构，新增 Lineman 转换器
// main/app/modules/takeout/infrastructure/adapter/lineman/

// lineman_menu_converter.go
type LinemanMenuConverter struct {
    dbm      *database.DBManager
    menuRepo repository.IMenuDataRepository
}

func NewLinemanMenuConverter(dbm *database.DBManager) *LinemanMenuConverter {
    return &LinemanMenuConverter{dbm: dbm}
}

// 实现 IPlatformConverter 接口
func (c *LinemanMenuConverter) GetPlatformName() string {
    return "lineman"
}

func (c *LinemanMenuConverter) ConvertToTTPOS(ctx context.Context, platformData interface{}) (interface{}, error) {
    // Lineman 菜单格式 → TTPOS 格式
}

func (c *LinemanMenuConverter) LoadMenuFromDatabase(ctx context.Context, companyUuid uint64, currencyUnit string) (interface{}, error) {
    // TTPOS 商品 → Lineman 菜单格式
}
```

```go
// 注册转换器（复用现有机制）
// main/app/modules/takeout/application/takeout_app_service.go

func NewTakeoutAppService(dbm *database.DBManager, ...) ITakeoutAppService {
    converters := make(map[string]service.IPlatformConverter)

    // 现有 Grab 转换器
    converters["grab"] = grab.NewGrabConverter(dbm, nil)

    // 新增 Lineman 转换器
    converters["lineman"] = lineman.NewLinemanMenuConverter(dbm)

    // ...
}
```

### 5.4 数据格式适配要点

#### 菜单格式对比

```json
// Grab 菜单格式
{
  "merchantID": "xxx",
  "partnerMerchantID": "shop_uuid",
  "currency": {"code": "THB", "exponent": 2},
  "sellingTimes": [...],
  "categories": [
    {
      "id": "CAT-001",
      "name": "饮品",
      "items": [
        {
          "id": "ITEM-001",
          "name": "珍珠奶茶",
          "price": 5000,  // 最小单位（分）
          "modifierGroups": [...]
        }
      ]
    }
  ]
}

// Lineman 菜单格式（v1）
{
  "menuGroups": [
    {
      "id": "GROUP-001",                    // 分类ID (max 30 chars)
      "name": {
        "thai": "เครื่องดื่ม",               // 必填：泰文名
        "english": "Beverages"               // 可选：英文名
      },
      "useSellingTime": true,               // 是否启用销售时段
      "startSellingTime": 600,              // 分钟数：10:00 = 600
      "endSellingTime": 1320,               // 分钟数：22:00 = 1320
      "menuItems": [
        {
          "id": "ITEM-001",                 // 菜品ID (max 30 chars)
          "name": {
            "thai": "ชานมไข่มุก",
            "english": "Bubble Milk Tea"
          },
          "description": {
            "thai": "ชานมไข่มุกสูตรพิเศษ",
            "english": "Special bubble milk tea"
          },
          "price": 50.00,                   // 泰铢（元，非分）
          "photoUrl": "https://...",        // 图片URL
          "menuStatus": "AVAILABLE",        // AVAILABLE | SOLD_OUT_TODAY | SUSPENDED
          "properties": [                   // 规格/选项组
            {
              "id": "SIZE",
              "name": {"thai": "ขนาด", "english": "Size"},
              "min": 1,                     // 最少选择数
              "max": 1,                     // 最多选择数
              "type": "1",                  // 1=单选, 2=多选
              "values": [
                {
                  "id": "SIZE-M",
                  "name": {"thai": "กลาง", "english": "Medium"},
                  "price": 0.00,            // 加价（默认0）
                  "status": 1               // 1=可售, 2=售罄, 3=暂停
                }
              ]
            }
          ]
        }
      ]
    }
  ]
}

// Lineman 菜单格式（v2 新增字段）
{
  "menuGroups": [{
    "menuItems": [{
      "salesChannelsAvailability": {      // v2 新增：渠道可用性
        "delivery": true,
        "pickup": true
      },
      "salesChannelsPrice": {             // v2 新增：渠道差异价格
        "pickup": 45.00                   // 自取价格
      },
      "properties": [{
        "values": [{
          "salesChannelsPrice": {         // 选项也支持渠道差异价格
            "pickup": 0.00
          }
        }]
      }]
    }]
  }]
}
```

**Grab vs Lineman 菜单字段映射**：

| Grab 字段         | Lineman 字段                          | 差异说明           |
| ----------------- | ------------------------------------- | ------------------ |
| `categories`      | `menuGroups`                          | 命名不同           |
| `items`           | `menuItems`                           | 命名不同           |
| `name` (string)   | `name.thai` + `name.english`          | Lineman 支持多语言 |
| `price` (int/分)  | `price` (double/元)                   | **单位不同**       |
| `modifierGroups`  | `properties`                          | 命名不同           |
| `modifiers`       | `values`                              | 命名不同           |
| `availableStatus` | `menuStatus` / `status`               | 状态值枚举不同     |
| `sellingTimes`    | `startSellingTime` + `endSellingTime` | 结构不同（分钟数） |
| -                 | `salesChannelsAvailability`           | Lineman v2 特有    |

#### 订单状态映射

| TTPOS 内部状态 | Grab 状态      | Lineman 状态 | 说明                   |
| -------------- | -------------- | ------------ | ---------------------- |
| 待接单 (0)     | PENDING        | （无推送）   | Lineman 下单即接单     |
| 已接单 (10)    | ACCEPTED       | （无推送）   | Lineman 不推送中间状态 |
| 配餐中 (15)    | PREPARING      | （无推送）   | Lineman 不推送中间状态 |
| 待骑手 (20)    | DRIVER_PENDING | （无推送）   | Lineman 不推送中间状态 |
| 配送中 (30)    | DELIVERING     | （无推送）   | Lineman 不推送中间状态 |
| 已完成 (40)    | COMPLETED      | `FINISH`     | Lineman 仅推送最终状态 |
| 已取消 (50)    | CANCELLED      | `CANCELED`   | Lineman 仅推送取消状态 |

> ⚠️ **注意**：Lineman `Status update notification` 仅推送 `FINISH` 和 `CANCELED` 两种最终状态。

---

## 六、风险与建议

### 6.1 主要风险

| 风险                  | 影响                            | 缓解措施                  |
| --------------------- | ------------------------------- | ------------------------- |
| **订单无法接单/拒单** | 店员无法通过 POS 拒绝订单       | 保留 WMA App 作为兜底     |
| **订单无法取消**      | 店员无法通过 POS 取消订单       | 需在 WMA App 操作         |
| **订单编辑需同步**    | WMA 上编辑订单需同步到 POS      | 实现 Order Update Webhook |
| **菜单格式差异**      | 价格单位、多语言、字段结构不同  | 开发 LinemanConverter     |
| **缺少 SDK**          | Grab 有官方 SDK，Lineman 无 SDK | 需自行封装 HTTP 客户端    |
| **门店上线周期**      | LINE MAN 配置需 4 个工作日      | 提前规划上线时间          |

### 6.2 建议

1. **保留 WMA App（Wongnai Merchant App）**：

   - 订单操作能力有限，WMA 作为兜底设备
   - 订单取消、编辑仍需在 WMA 操作
   - 系统故障时订单会自动回退到 WMA

2. **菜单同步策略**：

   - 避开高峰时段（10:00-14:00 / 17:00-19:00）
   - 实现 Trigger Sync Menu Webhook 以响应 Lineman 的同步请求
   - 促销菜单使用 `[Promotion]` 前缀

3. **订单处理策略**：

   - Lineman 订单采用"自动接单"模式
   - 仅监听最终状态（FINISH/CANCELED）
   - 实现 Order Update Notification 处理订单编辑

4. **渐进式实现**：
   - MVP 先实现必选 4 个 API
   - 后续根据需求补充可选功能

---

## 七、总结

| 维度         | 结论                               |
| ------------ | ---------------------------------- |
| **可行性**   | ✅ 可行，但功能比 Grab 精简        |
| **复用度**   | 约 60-70% 架构和代码可复用         |
| **开发量**   | MVP 约 13 人天，完整功能约 19 人天 |
| **主要限制** | 无订单操作能力（接单/拒单/取消）   |
| **特有功能** | 门店开关控制、订单编辑通知         |
