# T7: Shop端手机点餐设置 - 新增 IsOrderFirstPayLater 配置字段

## 变更概述

在门店手机点餐配置（store_scan_order）中新增 `is_order_first_pay_later` 字段，支持"先下单后付"模式。

## 变更文件

### 1. `main/app/dto/req/store_scan_order_setting.go`
- `SaveStoreScanOrderSettingReq` 新增字段 `IsOrderFirstPayLater int`
  - json tag: `is_order_first_pay_later`
  - form tag: `is_order_first_pay_later`
  - 0 = 先付后下单（默认），1 = 先下单后付

### 2. `main/app/dto/resp/setting/store_scan_order_setting.go`
- `StoreScanOrderSetting`（数据库存储结构）新增字段 `IsOrderFirstPayLater int`
- `StoreScanOrderSettingResp`（API 响应结构）新增字段 `IsOrderFirstPayLater int`
- 两者 json tag 均为 `is_order_first_pay_later`

### 3. `main/app/service/setting/setting.go`
- `GetStoreScanOrderSetting`: 响应组装中传递 `IsOrderFirstPayLater` 字段
- `SaveStoreScanOrderSetting`: 保存逻辑中传递 `IsOrderFirstPayLater` 字段
- 无需额外校验逻辑，字段默认值 0 即为先付后下单模式

### 4. `main/app/constant/order.go`
- 新增常量组：
  - `OrderFirstPayLaterNo = 0`  // 先付后下单（默认）
  - `OrderFirstPayLaterYes = 1` // 先下单后付

## 存储机制

该配置存储在 `ttpos_setting` 表中，key 为 `store_scan_order`，value 为整个 `StoreScanOrderSetting` 结构体的 JSON 序列化。新增字段自动参与 JSON 序列化/反序列化，已有数据中不包含此字段时默认为 0（先付后下单）。

## 验证

- `go fmt ./...` 通过
- `go vet ./...` 通过
