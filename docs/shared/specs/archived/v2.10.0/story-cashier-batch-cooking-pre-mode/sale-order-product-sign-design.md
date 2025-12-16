# 销售订单商品签名设计说明

## 概述

销售订单商品签名（`Sign`）是用于标识商品唯一性的字符串，主要用于**取消拆单时合并相同商品**。系统通过比较商品的签名来判断是否为相同商品，如果签名相同，则会将商品数量合并。

## 设计目标

1. **唯一性标识**：相同配置的商品生成相同签名，不同配置的商品生成不同签名
2. **合并判断**：在取消拆单、退菜等场景中，通过签名快速识别可合并的商品
3. **状态区分**：区分商品的不同业务状态（送厨、赠菜、打包、退菜等），避免错误合并

## 方法分类

系统根据商品类型使用不同的签名生成方法：

- **普通商品（商品包）**：使用 `GenerateProductSign()`
- **套餐商品**：使用 `GeneratePackageSign()`

```go
func (model *SaleOrderProduct) UpdateSign() {
    defer model.SetUpdate() // 标记要更新model
    if model.IsPackageProduct() {
        model.Sign = model.GeneratePackageSign()
    } else {
        model.Sign = model.GenerateProductSign()
    }
}
```

---

## 一、GenerateProductSign（商品包签名）

### 1.1 方法签名

```go
func (model *SaleOrderProduct) GenerateProductSign() string
```

### 1.2 签名格式

```
物料ID列表-属性ID列表-备注内容-必点方案UUID-送厨批次UUID-改价时间-赠菜时间-打包时间-退菜原因-H5订单UUID-是否接单-套餐UUID
```

**格式说明**：
- 字段之间使用 `-` 分隔
- 物料ID列表：多个物料UUID用 `,` 连接（已排序）
- 属性ID列表：多个属性UUID用 `,` 连接（已排序）
- 退菜原因：JSON格式字符串

### 1.3 实现逻辑

#### Step 1: 收集物料ID列表

```go
bomIdList := make([]string, 0)
for _, bom := range model.SaleOrderProductBoms {
    if bom.IsDelete() {
        continue
    }
    bomIdList = append(bomIdList, strconv.FormatUint(bom.ProductBomUuid, 10))
}
```

**说明**：
- 遍历所有 `SaleOrderProductBoms`（包含规格和小料）
- 跳过已删除的物料
- 收集所有物料的 `ProductBomUuid`

#### Step 2: 收集属性ID列表

```go
attributeIdList := make([]string, 0)
for _, attributeGroup := range model.SaleOrderProductAttributes {
    if attributeGroup.IsDelete() {
        continue
    }
    attributeIdList = append(attributeIdList, strconv.FormatUint(attributeGroup.ProductAttributeUuid, 10))
}
```

**说明**：
- 遍历所有 `SaleOrderProductAttributes`
- 跳过已删除的属性
- 收集所有属性的 `ProductAttributeUuid`

#### Step 3: 排序并拼接

```go
// 物料ID列表和属性ID列表排序
sort.Slice(bomIdList, func(i, j int) bool {
    return bomIdList[i] < bomIdList[j]
})
sort.Slice(attributeIdList, func(i, j int) bool {
    return attributeIdList[i] < attributeIdList[j]
})

bomIdListStr := strings.Join(bomIdList, ",")
attributeIdListStr := strings.Join(attributeIdList, ",")
```

**说明**：
- 对ID列表进行字符串排序（确保相同配置生成相同签名）
- 使用 `,` 连接多个ID

#### Step 4: 构建退菜原因

```go
type Reason struct {
    Uuids []uint64 `json:"uuids"`
    Text  string   `json:"text"`
}
reason := Reason{Uuids: make([]uint64, 0), Text: model.CancelReason}
for _, item := range model.CancelReasons {
    reason.Uuids = append(reason.Uuids, item.ReturnFoodReasonUuid)
}
reasonStr := utils.ToJson(reason)
```

**说明**：
- 构建包含退菜原因UUID列表和自定义文本的结构体
- 序列化为JSON字符串

#### Step 5: 生成最终签名

```go
return fmt.Sprintf("%s-%s-%s-%d-%d-%d-%d-%d-%s-%d-%d-%d",
    bomIdListStr,              // 物料ID列表
    attributeIdListStr,        // 属性ID列表
    model.Remark,              // 备注内容
    model.MustPlanUuid,        // 必点方案UUID
    model.ProductionOrderUuid,  // 送厨批次UUID（生产订单UUID）
    model.ChangePriceTime,      // 改价时间（时间戳）
    model.GiftTime,            // 赠菜时间（时间戳）
    model.WrapTime,            // 打包时间（时间戳）
    reasonStr,                 // 退菜原因（JSON字符串）
    model.H5OrderUuid,         // H5订单UUID
    model.IsAcceptOrder,       // 是否接单（0-否 1-是）
    model.PackageUuid,         // 套餐UUID
)
```

### 1.4 包含的字段说明

| 字段 | 类型 | 说明 | 影响合并的场景 |
|------|------|------|----------------|
| 物料ID列表 | string | 规格和小料的UUID列表（已排序） | 规格或小料不同 |
| 属性ID列表 | string | 商品属性的UUID列表（已排序） | 属性不同 |
| 备注内容 | string | 顾客备注信息 | 备注不同 |
| 必点方案UUID | uint64 | 必点方案标识 | 必点方案不同 |
| 送厨批次UUID | uint64 | 生产订单UUID，标识送厨批次 | 送厨批次不同 |
| 改价时间 | int64 | 改价时间戳，0表示未改价 | 改价时间不同 |
| 赠菜时间 | int64 | 赠菜时间戳，0表示非赠菜 | 赠菜状态不同 |
| 打包时间 | int64 | 打包时间戳，0表示未打包 | 打包状态不同 |
| 退菜原因 | string | JSON格式的退菜原因 | 退菜原因不同 |
| H5订单UUID | uint64 | 扫码订单UUID | H5订单不同 |
| 是否接单 | uint | 0-未接单 1-已接单 | 接单状态不同 |
| 套餐UUID | uint64 | 套餐标识（普通商品可能为0） | 套餐不同 |

### 1.5 示例

**场景1：普通商品（无属性、无小料）**
```
123--备注-0-0-0-0-0-{"uuids":[],"text":""}-0-0-0
```

**场景2：带规格、属性、小料的商品**
```
123,456-789,101112-备注-0-0-0-0-0-{"uuids":[],"text":""}-0-0-0
```

**场景3：已送厨的商品**
```
123--备注-0-999-0-0-0-{"uuids":[],"text":""}-0-0-0
```

---

## 二、GeneratePackageSign（套餐签名）

### 2.1 方法签名

```go
func (model *SaleOrderProduct) GeneratePackageSign() string
```

### 2.2 签名格式

```
套餐UUID-子商品参数-备注内容-必点方案UUID-送厨批次UUID-改价时间-赠菜时间-打包时间-退菜原因-H5订单UUID-是否接单
```

**格式说明**：
- 字段之间使用 `-` 分隔
- 子商品参数：JSON格式字符串，包含套餐子商品的配置信息

### 2.3 实现逻辑

#### Step 1: 获取套餐UUID

```go
packageUuid := model.ProductPackageUuid
```

#### Step 2: 构建退菜原因（与商品包相同）

```go
type Reason struct {
    Uuids []uint64 `json:"uuids"`
    Text  string   `json:"text"`
}
reason := Reason{Uuids: make([]uint64, 0), Text: model.CancelReason}
for _, item := range model.CancelReasons {
    reason.Uuids = append(reason.Uuids, item.ReturnFoodReasonUuid)
}
reasonStr := utils.ToJson(reason)
```

#### Step 3: 生成最终签名

```go
return fmt.Sprintf("%d-%s-%s-%d-%d-%d-%d-%d-%s-%d-%d",
    packageUuid,              // 套餐UUID
    model.PackageSubProductParams, // 子商品参数（JSON字符串）
    model.Remark,             // 备注内容
    model.MustPlanUuid,       // 必点方案UUID
    model.ProductionOrderUuid, // 送厨批次UUID
    model.ChangePriceTime,     // 改价时间
    model.GiftTime,           // 赠菜时间
    model.WrapTime,           // 打包时间
    reasonStr,                // 退菜原因
    model.H5OrderUuid,        // H5订单UUID
    model.IsAcceptOrder,      // 是否接单
)
```

### 2.4 包含的字段说明

| 字段 | 类型 | 说明 | 影响合并的场景 |
|------|------|------|----------------|
| 套餐UUID | uint64 | 套餐商品包UUID | 套餐不同 |
| 子商品参数 | string | JSON格式的套餐子商品配置 | 子商品配置不同 |
| 备注内容 | string | 顾客备注信息 | 备注不同 |
| 必点方案UUID | uint64 | 必点方案标识 | 必点方案不同 |
| 送厨批次UUID | uint64 | 生产订单UUID | 送厨批次不同 |
| 改价时间 | int64 | 改价时间戳 | 改价时间不同 |
| 赠菜时间 | int64 | 赠菜时间戳 | 赠菜状态不同 |
| 打包时间 | int64 | 打包时间戳 | 打包状态不同 |
| 退菜原因 | string | JSON格式的退菜原因 | 退菜原因不同 |
| H5订单UUID | uint64 | 扫码订单UUID | H5订单不同 |
| 是否接单 | uint | 0-未接单 1-已接单 | 接单状态不同 |

**注意**：套餐签名不包含 `PackageUuid` 字段（因为本身就是套餐商品）

### 2.5 子商品参数格式

`PackageSubProductParams` 是JSON字符串，包含套餐子商品的配置信息，格式如下：

```json
[
  {
    "flavor_uuid": 123,
    "attribute_uuid": [456, 789],
    "product_package_group_uuid": 101
  },
  {
    "flavor_uuid": 234,
    "attribute_uuid": [567],
    "product_package_group_uuid": 102
  }
]
```

### 2.6 示例

**场景1：简单套餐（无子商品配置）**
```
999-[]-备注-0-0-0-0-0-{"uuids":[],"text":""}-0-0
```

**场景2：带子商品配置的套餐**
```
999-[{"flavor_uuid":123,"attribute_uuid":[456],"product_package_group_uuid":101}]-备注-0-0-0-0-0-{"uuids":[],"text":""}-0-0
```

---

## 三、签名更新场景

以下场景会触发签名重新生成（调用 `UpdateSign()`）：

1. **改价**：销售订单商品价格修改后
2. **修改备注**：商品备注信息变更
3. **送厨**：商品状态变为已送厨
4. **赠菜**：商品标记为赠菜
5. **打包**：商品标记为打包
6. **退菜**：商品标记为退菜
7. **H5下单**：H5订单商品下单
8. **接单**：H5订单商品接单

### 3.1 调用位置示例

```go
// 1. 改价后更新签名
func (model *SaleOrderProduct) ChangeProductPrice(price float64) {
    model.ChangePriceTime = time.Now().Unix()
    model.SalePrice = price
    model.UpdateSign() // 重新签名
}

// 2. 送厨后更新签名
func (model *SaleOrderProduct) SetCooking(productionOrderUuid uint64) {
    model.Status = constant.SaleOrderProductStatusCooking
    model.ProductionOrderUuid = productionOrderUuid
    model.UpdateSign() // 更新签名
    model.SendKitchenTime = time.Now().Unix()
    model.SetUpdate()
}

// 3. 赠菜后更新签名
func (model *SaleOrderProduct) SetGiftProduct(giftReason string) {
    model.GiftTime = time.Now().Unix()
    model.SetUpdate()
    model.GiftReason = giftReason
    model.UpdateSign() // 更新签名
}
```

---

## 四、签名使用场景

### 4.1 取消拆单合并商品

在取消拆单时，系统通过签名判断是否为相同商品，如果签名相同则合并数量：

```go
sameSignSaleOrderProduct := saleOrder.GetSaleOrderProductBySign(newSaleOrderProduct.Sign)
if sameSignSaleOrderProduct != nil {
    // 有相同签名的商品，将两个商品合并，数量相加
    sameSignSaleOrderProduct.SetNum(sameSignSaleOrderProduct.Num + req.Num)
    returnSaleOrderProduct = sameSignSaleOrderProduct
    saleOrderProduct.SetDelete() // 标记该商品为删除
}
```

### 4.2 移动商品合并

在移动商品到其他订单时，如果目标订单中存在相同签名的商品，则合并：

```go
if IsSameSignature(saleOrderProduct.Sign, toSaleOrderProductSignMap) {
    toSaleOrderProductSignMap[saleOrderProduct.Sign].Num += moveProductNum
}
```

### 4.3 送厨商品分组

在送厨时，系统按签名对商品进行分组，相同签名的商品会被合并处理：

```go
if p, exists := signProduct[product.Sign]; exists {
    // 相同签名的商品合并
}
```

---

## 五、设计要点

### 5.1 排序的重要性

物料ID和属性ID列表必须排序，确保相同配置的商品生成相同的签名：

```go
sort.Slice(bomIdList, func(i, j int) bool {
    return bomIdList[i] < bomIdList[j]
})
```

**原因**：如果不排序，`[1,2,3]` 和 `[3,2,1]` 会生成不同的签名，但实际上它们是相同的配置。

### 5.2 状态字段的作用

签名中包含多个状态字段（送厨批次、改价时间、赠菜时间、打包时间等），用于区分商品的不同业务状态：

- **送厨批次**：不同批次送厨的商品不应合并
- **改价时间**：不同时间改价的商品不应合并
- **赠菜时间**：赠菜商品和普通商品不应合并
- **打包时间**：打包商品和堂食商品不应合并
- **退菜原因**：不同退菜原因的商品不应合并
- **是否接单**：未接单的H5商品和已接单的商品不应合并

### 5.3 特殊处理

#### 打包时间的特殊处理

```go
func (model *SaleOrderProduct) SetWrap() {
    model.WrapTime = 1 // 不需要按打包时间合并商品。打包时间=1
    model.UpdateSign()
}
```

**说明**：打包时间统一设置为 `1`，而不是实际时间戳，这样所有打包商品可以合并。

#### 取消打包不更新签名

```go
func (model *SaleOrderProduct) SetUnwrap() {
    model.WrapTime = 0
    // 注：暂时不更新商品的签名，历史遗留问题取消赠菜也没有更新签名
    // 如果更新签名的话可能会与未打包的商品签名一致需要合并商品
}
```

**说明**：取消打包时不更新签名，避免与未打包商品合并。

### 5.4 套餐与普通商品的差异

| 维度 | 普通商品 | 套餐商品 |
|------|---------|---------|
| 物料信息 | 包含规格和小料的UUID列表 | 不包含（使用子商品参数） |
| 属性信息 | 包含属性UUID列表 | 不包含（使用子商品参数） |
| 子商品配置 | 不包含 | 包含（PackageSubProductParams） |
| 套餐UUID | 包含（可能为0） | 不包含（本身就是套餐） |

---

## 六、注意事项

### 6.1 签名长度限制

数据库字段定义：
```go
Sign string `gorm:"column:sign;type:varchar(255);not null;default:'';comment:'商品签名...'"`
```

**限制**：签名最大长度为 255 字符。如果签名过长，可能导致：
- 数据库存储失败
- 签名截断，导致合并判断错误

**建议**：监控签名长度，如果超过 200 字符，考虑优化签名生成逻辑。

### 6.2 退菜原因的JSON序列化

退菜原因使用JSON格式存储，需要注意：
- JSON中的特殊字符需要转义
- 空数组和空字符串的处理
- JSON格式的一致性

### 6.3 时间戳的精度

所有时间字段使用 `int64` 类型的时间戳（秒级精度），确保：
- 时间比较的准确性
- 跨时区的一致性
- 存储效率

### 6.4 签名更新的原子性

签名更新应该与业务操作在同一事务中完成，确保：
- 签名与商品状态的一致性
- 并发场景下的数据正确性

---

## 七、扩展建议

### 7.1 签名版本化

如果未来需要修改签名生成逻辑，建议引入版本号：

```go
type SignVersion uint8
const (
    SignVersionV1 SignVersion = 1
    SignVersionV2 SignVersion = 2
)

func (model *SaleOrderProduct) GenerateProductSign() string {
    version := SignVersionV1
    // 根据版本选择不同的生成逻辑
}
```

### 7.2 签名缓存

对于频繁查询的场景，可以考虑缓存签名到商品对象的映射关系。

### 7.3 签名校验

在关键业务操作前，可以校验签名的有效性，确保商品状态的一致性。

---

## 八、相关方法

### 8.1 UpdateSign

统一入口，根据商品类型调用对应的签名生成方法：

```go
func (model *SaleOrderProduct) UpdateSign() {
    defer model.SetUpdate()
    if model.IsPackageProduct() {
        model.Sign = model.GeneratePackageSign()
    } else {
        model.Sign = model.GenerateProductSign()
    }
}
```

### 8.2 ProductKey

生成商品键（用于其他场景，不同于签名）：

```go
func (model *SaleOrderProduct) ProductKey() string {
    // 格式：规格id-属性id,属性id-加料id,加料id
    return fmt.Sprintf("%d-%s-%s", flavorUuid, attributeIds, sauceIds)
}
```

**区别**：
- `ProductKey`：仅包含规格、属性、小料，用于商品配置识别
- `Sign`：包含更多业务状态信息，用于商品合并判断

---

## 九、总结

销售订单商品签名是一个关键的设计，它：

1. **支持商品合并**：通过签名快速识别相同商品，实现数量合并
2. **区分业务状态**：通过包含状态字段，避免错误合并
3. **保证一致性**：通过排序和标准化格式，确保相同配置生成相同签名
4. **适应复杂场景**：支持普通商品和套餐商品的不同签名生成逻辑

在实际使用中，需要注意签名的更新时机、长度限制和特殊场景的处理，确保系统的稳定性和正确性。

---

**文档版本**：v1.0  
**最后更新**：2025-11-20  
**维护者**：TTPOS Team

