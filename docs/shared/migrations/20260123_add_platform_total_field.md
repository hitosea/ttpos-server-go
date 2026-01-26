# 外卖订单表添加 platform_total 字段

## 变更概述

为 `ttpos_takeout_order` 表添加 `platform_total` 字段，用于准确记录 Grab 等外卖平台的结算总额。

## 背景说明

在 Grab 外卖订单中，平台结算总额的计算公式为：
```
platform_total = subtotal + merchant_charge_fee - merchant_discount
```

之前没有单独字段存储这个值，导致需要在查询时实时计算，影响性能和准确性。

## 变更内容

### 1. 数据库迁移文件

**文件路径**: `admin/database/migrations/20260123143000_add_platform_total_to_takeout_order.php`

- 添加 `platform_total` 字段（decimal(20,4) 类型）
- 位置：在 `merchant_charge_fee` 字段之后
- 默认值：0.0000
- 注释：平台结算总额 (subtotal + merchant_charge_fee - merchant_discount)
- 自动更新历史数据：使用公式计算并更新所有未删除的订单

### 2. 种子文件更新

**文件路径**: `admin/database/seeds/shop_01.sql`

- 在 `ttpos_takeout_order` 表结构中添加 `platform_total` 字段定义

### 3. Go Model 更新

**文件路径**: `main/app/modules/takeout/domain/model/takeout_order.go`

- 添加 `PlatformTotal` 字段到 `TakeoutOrder` 结构体
- 字段类型：float64
- JSON 标签：platform_total
- GORM 标签：column:platform_total

### 4. Grab 订单转换器更新

**文件路径**: `main/app/modules/takeout/infrastructure/adapter/grab/grab_order_converter.go`

- 在设置价格信息时，自动计算并设置 `PlatformTotal` 值
- 计算逻辑：使用 `decimal` 库确保精度
  ```go
  order.PlatformTotal = decimal.NewFromFloat(order.Subtotal).
      Add(decimal.NewFromFloat(order.MerchantChargeFee)).
      Sub(decimal.NewFromFloat(order.MerchantDiscount)).
      InexactFloat64()
  ```
- 位置：在 `MerchantChargeFee` 赋值之后

### 5. 外卖订单列表接口更新

**文件路径**: `main/app/modules/takeout/domain/service/takeout_order_service.go` - `GetList` 方法

- 修改列表响应的 `Subtotal` 字段，返回 `platform_total` 字段的值（而非原来的 `eater_payment`）
- 变更原因：列表中展示的应该是平台结算总额，而非顾客实付金额

### 6. 外卖订单详情接口更新

**文件路径**: 
- `main/app/modules/takeout/interfaces/response/takeout_order_response.go` - 响应结构
- `main/app/modules/takeout/domain/service/takeout_order_service.go` - `GetByUuid` 方法

**响应结构变更**：
- 在 `TakeoutOrderPriceResp` 中新增 `platform_total` 字段

**版本兼容处理**：
- v2.15.0 及以上版本：
  - 正常返回所有字段，包括新增的 `platform_total`
  - `eater_payment` 字段保持原值
  
- v2.15.0 之前版本：
  - `eater_payment` 字段返回 `platform_total` 的值（兼容旧版本）
  - 保证旧版本客户端能正确显示平台结算金额

**实现逻辑**：
```go
// 获取客户端版本
clientVersion := ctx.GetVersion()

// 构建价格信息
priceResp := response.TakeoutOrderPriceResp{
    // ... 其他字段
    PlatformTotal: order.PlatformTotal,
}

// 版本兼容：2.15.0 之前的版本，eater_payment 返回 platform_total 的值
if ctx.Version(context.LT, "2.15.0") {
    priceResp.EaterPayment = order.PlatformTotal
}
```

### 7. 外卖订单打印模块更新

**文件路径**: `main/app/modules/printer/template/takeout_order_img_custom.go`

**变更说明**：
- 打印单据上的金额字段改为使用 `platform_total`
- 影响字段：
  - `ActualReceivePrice`（实际收款金额）：从 `order.EaterPayment` 改为 `order.PlatformTotal`
  - `PaidAmount`（支付金额）：从 `order.EaterPayment` 改为 `order.PlatformTotal`

**变更原因**：
- 打印单据应该显示平台结算总额，而非顾客实付金额
- 与列表和详情接口保持一致

### 8. ERP 发票同步模块更新

**文件路径**: `main/app/modules/takeout/domain/service/takeout_erp_sync_service.go` - `buildPosInvoicePayments` 方法

**变更说明**：
- ERP 发票支付金额改为使用 `platform_total`
- 影响：
  - Grab 订单：`Amount` 从 `order.EaterPayment` 改为 `order.PlatformTotal`
  - Lineman 订单：`Amount` 从 `order.EaterPayment` 改为 `order.PlatformTotal`

**变更原因**：
- ERP 系统中应该记录平台结算总额，而非顾客实付金额
- 确保 ERP 发票金额与订单实际结算金额一致
- 方便财务对账和报表统计

**实现逻辑**：
```go
func buildPosInvoicePayments(takeoutOrder *takeoutModel.TakeoutOrder, existPayment *model.PaymentMethod) []*selling.PosInvoicePayment {
    payments := make([]*selling.PosInvoicePayment, 0)
    
    if takeoutOrder.IsGrabOrder() {
        payments = append(payments, &selling.PosInvoicePayment{
            ModeOfPayment: existPayment.ErpnextPayment,
            Amount:        takeoutOrder.PlatformTotal,  // 使用平台结算总额
        })
    }
    // ... Lineman 类似
    
    return payments
}
```

## 影响范围

### 直接影响

1. **数据库表结构**：`ttpos_takeout_order` 表新增一个字段
2. **数据完整性**：历史订单数据会被自动更新（基于现有字段计算）
3. **新订单创建**：所有新创建的 Grab 订单都会自动设置 `platform_total` 值
4. **API 响应变更**：
   - 订单列表接口：`subtotal` 字段返回值从 `eater_payment` 改为 `platform_total`
   - 订单详情接口：新增 `platform_total` 字段，v2.15.0 之前版本 `eater_payment` 返回 `platform_total` 值
5. **打印单据变更**：
   - 打印订单的实际收款金额和支付金额改为显示 `platform_total`
   - 确保打印单据与系统显示的金额一致
6. **ERP 发票同步变更**：
   - ERP 发票支付金额改为使用 `platform_total`
   - 确保 ERP 系统中记录的金额与平台结算金额一致
   - 方便财务对账和报表统计

### 间接影响

1. **查询性能**：不再需要实时计算，直接查询字段值
2. **报表统计**：可以直接使用 `platform_total` 字段进行统计和分析
3. **数据一致性**：避免因计算错误导致的数据不一致
4. **客户端兼容性**：
   - v2.15.0+ 客户端：使用新的 `platform_total` 字段
   - v2.15.0 之前客户端：通过 `eater_payment` 字段获取 `platform_total` 值（向后兼容）

## 部署说明

### 执行顺序

1. **备份数据库**（重要！）
2. 运行数据库迁移：
   ```bash
   cd admin
   php think migrate:run
   ```
3. 验证迁移结果：
   ```sql
   -- 检查字段是否添加成功
   DESC ttpos_takeout_order;
   
   -- 检查历史数据是否更新成功（随机抽样）
   SELECT 
       uuid,
       subtotal,
       merchant_charge_fee,
       merchant_discount,
       platform_total,
       (subtotal + merchant_charge_fee - merchant_discount) as calculated_total
   FROM ttpos_takeout_order 
   WHERE delete_time = 0 
   LIMIT 10;
   ```
4. 重启 Go Main 服务

### 验证方法

1. **验证历史数据**：
   ```sql
   -- 检查是否有计算错误的记录（应该返回 0 行）
   SELECT COUNT(*) 
   FROM ttpos_takeout_order 
   WHERE delete_time = 0 
   AND ABS(platform_total - (subtotal + merchant_charge_fee - merchant_discount)) > 0.0001;
   ```

2. **验证新订单**：
   - 创建一个新的 Grab 外卖订单
   - 检查数据库中 `platform_total` 字段是否有值
   - 验证计算公式是否正确

## 回滚方案

如果需要回滚，可以执行：

```sql
-- 删除字段（注意：会丢失该字段的所有数据）
ALTER TABLE ttpos_takeout_order DROP COLUMN platform_total;
```

同时需要回滚以下文件的修改：
- `takeout_order.go` (Model)
- `grab_order_converter.go` (转换器)
- `takeout_order_service.go` (服务层 - 列表和详情接口)
- `takeout_order_response.go` (响应结构)
- `takeout_order_img_custom.go` (打印模板)
- `takeout_erp_sync_service.go` (ERP 同步服务)
- `shop_01.sql` (种子文件)

## 相关文件清单

1. **迁移文件**：`admin/database/migrations/20260123143000_add_platform_total_to_takeout_order.php`
2. **种子文件**：`admin/database/seeds/shop_01.sql`
3. **Model 文件**：`main/app/modules/takeout/domain/model/takeout_order.go`
4. **转换器文件**：`main/app/modules/takeout/infrastructure/adapter/grab/grab_order_converter.go`
5. **服务层文件**：`main/app/modules/takeout/domain/service/takeout_order_service.go`
6. **响应结构文件**：`main/app/modules/takeout/interfaces/response/takeout_order_response.go`
7. **打印模板文件**：`main/app/modules/printer/template/takeout_order_img_custom.go`
8. **ERP 同步服务**：`main/app/modules/takeout/domain/service/takeout_erp_sync_service.go`

## 注意事项

1. **数据精度**：使用 decimal(20,4) 确保金额计算精度
2. **历史数据**：迁移脚本会自动更新所有历史订单（包括已删除的）
3. **其他平台**：目前只有 Grab 平台的转换器设置了 `platform_total`，如果将来实现 Lineman 或 Shopeefood 平台，需要在对应的转换器中添加类似逻辑
4. **单元测试**：建议添加单元测试验证计算逻辑的正确性
5. **版本兼容**：
   - 确保前端 v2.15.0 及以上版本使用新的 `platform_total` 字段
   - v2.15.0 之前的版本会通过 `eater_payment` 字段获取 `platform_total` 值
   - 服务端会自动根据客户端版本返回相应的数据
6. **API 变更影响**：
   - 订单列表的 `subtotal` 字段含义改变，前端需要相应调整
   - 详情接口新增字段，旧版本客户端不受影响

## 测试建议

### 单元测试
- 测试 Grab 订单转换器的 `platform_total` 计算逻辑
- 测试边界情况（零值、负值、小数精度）
- 测试版本兼容逻辑（v2.15.0 前后版本）

### 集成测试
- 创建完整的 Grab 订单流程
- 验证 `platform_total` 是否正确保存到数据库
- 验证查询和统计功能
- 测试不同客户端版本的 API 响应

### API 测试
- 测试列表接口：验证 `subtotal` 返回 `platform_total` 值
- 测试详情接口（v2.15.0+）：验证返回新的 `platform_total` 字段
- 测试详情接口（v2.15.0-）：验证 `eater_payment` 返回 `platform_total` 值

### 打印测试
- 打印外卖订单小票（商家联）
- 打印外卖订单小票（客户联）
- 验证打印单据上的金额是否正确显示为 `platform_total`
- 对比打印金额与订单详情接口返回的金额是否一致

### ERP 同步测试
- 创建新的 Grab 外卖订单并接单（触发 ERP 同步）
- 检查 ERP 系统中的发票支付金额是否为 `platform_total`
- 验证 ERP 发票金额与订单的 `platform_total` 字段是否一致
- 测试订单取消后的 ERP 发票取消功能
- 验证财务对账报表中的金额准确性

---

**创建时间**: 2026-01-23
**创建人**: AI Assistant
**相关 Issue**: Grab 取值问题修复
