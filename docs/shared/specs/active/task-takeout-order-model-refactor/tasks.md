# task-takeout-order-model-refactor 任务清单

## 📊 进度总览

| 项目     | 数值 |
| -------- | ---- |
| 总 SP    | 3    |
| 总任务数 | 8    |
| 已完成   | 8    |
| 完成率   | 100% |

---

## Phase 1: 重构共享结构体

### 1.1 重构 TakeoutOrder 主结构体

| 项目         | 内容                                                 |
| ------------ | ---------------------------------------------------- |
| File         | `ttpos-api/ttpos-takeout/message/takeout_order.go`   |
| Purpose      | 字段完全对齐 Grab SDK Order                          |
| Requirements | Req 1: 字段名、类型、必选/可选与 Grab SDK 一致       |
| Leverage     | `github.com/grab/grabfood-api-sdk-go@v1.0.2/model_order.go` |

**修改要点**：
- `Cutlery`: `*bool` → `bool`
- `SubmitTime`: `*string` → `*time.Time`
- `CompleteTime`: `*string` → `*time.Time`
- `Currency`: `*TakeoutCurrency` → `TakeoutCurrency` (必需)
- `FeatureFlags`: `*TakeoutFeatureFlags` → `TakeoutFeatureFlags` (必需)
- 新增 `AdditionalProperties map[string]interface{}`

- [x] 完成

### 1.2 重构 TakeoutOrderItem

| 项目         | 内容                                                 |
| ------------ | ---------------------------------------------------- |
| File         | `ttpos-api/ttpos-takeout/message/takeout_order.go`   |
| Purpose      | 字段完全对齐 Grab SDK OrderItem                      |
| Requirements | Req 2: OrderItem 字段对齐                            |
| Leverage     | `github.com/grab/grabfood-api-sdk-go@v1.0.2/model_order_item.go` |

**修改要点**：
- `GrabItemID`: `*string` → `string` (必需)
- `Quantity`: `int` → `int32`
- `Price`: `float64` → `int64`
- `Tax`: `*float64` → `*int64`
- 新增 `AdditionalProperties map[string]interface{}`

- [x] 完成

### 1.3 重构 TakeoutOrderPrice

| 项目         | 内容                                                 |
| ------------ | ---------------------------------------------------- |
| File         | `ttpos-api/ttpos-takeout/message/takeout_order.go`   |
| Purpose      | 字段完全对齐 Grab SDK OrderPrice                     |
| Requirements | Req 2: OrderPrice 字段对齐                           |
| Leverage     | `github.com/grab/grabfood-api-sdk-go@v1.0.2/model_order_price.go` |

**修改要点**：
- `Subtotal`: `float64` → `int64`
- `Tax`: `*float64` → `*int64`
- `MerchantChargeFee`: `*float64` → `*int64`
- `GrabFundPromo`: `*float64` → `*int64`
- `MerchantFundPromo`: `*float64` → `*int64`
- `DeliveryFee`: `*float64` → `*int64`
- `EaterPayment`: `*float64` → `*int64`
- 新增 `BasketPromo *int64`
- 新增 `SmallOrderFee *int64`
- 删除 `Total`（Grab SDK 无此字段）
- 新增 `AdditionalProperties map[string]interface{}`

- [x] 完成

### 1.4 创建 Nullable 类型

| 项目         | 内容                                                 |
| ------------ | ---------------------------------------------------- |
| File         | `ttpos-api/ttpos-takeout/message/takeout_order.go`   |
| Purpose      | 创建 Nullable 包装类型                               |
| Requirements | Req 1: 与 Grab SDK 类型一致                          |

**创建类型**：
- `NullableTakeoutDineIn`
- `NullableTakeoutReceiver`
- `NullableTakeoutOutOfStockInstruction`

- [x] 完成

### 1.5 重构其他关联结构体

| 项目         | 内容                                                 |
| ------------ | ---------------------------------------------------- |
| File         | `ttpos-api/ttpos-takeout/message/takeout_order.go`   |
| Purpose      | 重构 Modifier、Campaign、Promo 等结构体              |
| Requirements | Req 2: 所有关联结构体对齐                            |

**结构体清单**：
- `TakeoutModifier` (对齐 OrderItemModifier)
- `TakeoutCampaign` (对齐 OrderCampaign)
- `TakeoutPromo` (对齐 OrderPromo)
- `TakeoutCurrency` (对齐 Currency)
- `TakeoutFeatureFlags` (对齐 OrderFeatureFlags)
- `TakeoutDineIn` (对齐 DineIn)
- `TakeoutReceiver` (对齐 Receiver)
- `TakeoutOrderReadyEstimation` (对齐 OrderReadyEstimation)
- `TakeoutOutOfStockInstruction` (对齐 OutOfStockInstruction)

- [x] 完成

---

## Phase 2: 更新 BMP 转换器

### 2.1 更新 Lineman 转换器

| 项目         | 内容                                                      |
| ------------ | --------------------------------------------------------- |
| File         | `ttpos-bmp/app/ttpos-takeout/utility/takeout_converter.go`|
| Purpose      | 适配新的 TakeoutOrder 结构体                              |
| Requirements | Req 4: Lineman 订单正确转换                               |

**修改要点**：
- 价格字段从 float64 转换为 int64（泰铢→分）
- Lineman 特有字段存入 AdditionalProperties
- 更新 Nullable 类型的使用方式

- [x] 完成

---

## Phase 3: 单元测试

### 3.1 编写转换器单元测试

| 项目         | 内容                                                           |
| ------------ | -------------------------------------------------------------- |
| File         | `ttpos-bmp/app/ttpos-takeout/utility/takeout_converter_test.go`|
| Purpose      | 测试 Lineman 订单转换                                          |
| Requirements | 覆盖率 ≥ 80%                                                   |

**测试场景**：
- Lineman 完整订单转换
- Lineman 订单缺少可选字段
- Lineman 特有字段存入 AdditionalProperties
- 价格转换正确性（泰铢→分）

- [x] 完成

### 3.2 验证编译通过

| 项目         | 内容                    |
| ------------ | ----------------------- |
| Purpose      | 确保所有模块编译通过    |
| Requirements | 无编译错误              |

**验证命令**：
```bash
cd ttpos-api && go build ./...
cd ttpos-bmp && go build ./...
```

- [x] 完成

---

## 提交清单

### 代码质量
- [x] `go mod tidy` 执行
- [x] `go fmt ./...` 执行
- [x] `go vet ./...` 通过
- [x] 测试通过: `go test ./...`

### 功能完整性
- [x] TakeoutOrder 字段与 Grab SDK Order 完全一致
- [x] 所有关联结构体已重构
- [x] BMP 转换器已适配
- [x] 单元测试覆盖率 ≥ 80%

### 验收标准
- [x] Grab 订单可直接映射到 TakeoutOrder
- [x] Lineman 订单能正确转换到 TakeoutOrder
- [x] 现有 Grab 和 Lineman 订单处理流程正常工作

---

**版本**: v1.0.0
**创建日期**: 2026-01-23
**完成日期**: 2026-01-23
