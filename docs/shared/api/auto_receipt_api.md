# 品采自动收货配置 API 文档

> 📚 **受众**: 前端开发者、API 对接人员
> 📖 **用途**: 品采自动收货规则管理、记录查询相关 API 接口文档

---

## 元信息

| 字段      | 内容                                      |
| --------- | ----------------------------------------- |
| 模块      | `shop_setting / auto_receipt`             |
| 版本      | `v1.0.0`                                 |
| 更新时间  | `2026-03-13`                              |
| 关联 Spec | `story-shop-auto-receipt-config`          |
| 关联任务  | DooTask #40189                            |

---

## 1. 概述

品采自动收货配置模块提供总部级别的自动收货规则管理功能：

- **规则管理**：按仓库创建/更新/删除自动收货规则，配置延迟天数和关联门店
- **门店查询**：查询可配置门店列表，已配置的门店标记为 disabled
- **收货记录**：查询自动收货执行日志及收货单详情

**业务规则**：
- 同一仓库下，每个门店只能被一条规则关联
- 定时任务每小时执行，按规则的延迟天数判断是否触发自动收货
- 仅处理品采（`purchase_type = 2`）已审核通过的采购单
- 自动创建的收货单标记 `is_auto_receipt = true`

**终端支持**：Shop（商家端后台）

**接口数量**：新增 7 个，修改 3 个

**认证方式**：JWT Token

---

## 2. 快速索引

### 2.1 新增接口

| 级别 | 接口                      | 方法 | 路径                                         | 描述               |
| ---- | ------------------------- | ---- | -------------------------------------------- | ------------------ |
| P0   | `CreateAutoReceiptRule`   | POST | `/shop/setting/auto_receipt/rule/create`     | 创建自动收货规则   |
| P0   | `UpdateAutoReceiptRule`   | POST | `/shop/setting/auto_receipt/rule/update`     | 更新自动收货规则   |
| P0   | `DeleteAutoReceiptRule`   | POST | `/shop/setting/auto_receipt/rule/delete`     | 删除自动收货规则   |
| P0   | `GetAutoReceiptRuleList`  | GET  | `/shop/setting/auto_receipt/rule/list`       | 获取规则列表       |
| P1   | `GetAutoReceiptShopList`  | GET  | `/shop/setting/auto_receipt/shop_list`       | 获取可选门店列表   |
| P1   | `GetAutoReceiptLogList`   | GET  | `/shop/setting/auto_receipt/log/list`        | 获取自动收货记录   |
| P1   | `GetAutoReceiptLogDetail` | GET  | `/shop/setting/auto_receipt/log/detail`      | 获取收货记录详情   |

### 2.2 修改接口

| 接口 | 变更说明 |
| ---- | -------- |
| 收货单列表 | 响应新增 `is_auto_receipt` 字段 |
| 收货单详情 | 响应新增 `is_auto_receipt` 字段 |
| 采购单详情收货清单 | `receipt_orders[]` 新增 `is_auto_receipt` 字段 |

---

## 3. 接口详情

### 3.1 创建自动收货规则

- **Method & URL**
  ```http
  POST /shop/setting/auto_receipt/rule/create
  ```

- **Headers**
  | 名称          | 必填 | 示例             | 说明      |
  | ------------- | ---- | ---------------- | --------- |
  | Authorization | 是   | `Bearer {token}` | JWT Token |
  | Content-Type  | 是   | `application/json` | 内容类型 |

- **Request**

  ```json
  {
    "warehouse_erp_code": "WH001",
    "warehouse_name": "中央仓库",
    "shop_uuids": [8267304538112001, 8267304538112002],
    "delay_days": 3
  }
  ```

  | 字段                | 类型     | 必填 | 说明                             |
  | ------------------- | -------- | ---- | -------------------------------- |
  | `warehouse_erp_code`| string   | 是   | 仓库 ERP 编码                    |
  | `warehouse_name`    | string   | 是   | 仓库名称（用于展示）             |
  | `shop_uuids`        | uint64[] | 是   | 关联门店 UUID 列表，至少 1 个    |
  | `delay_days`        | int      | 否   | 延迟天数，默认 0（DN 到达当天收货）|

- **Response**

  ```json
  {
    "code": 1,
    "message": "success",
    "data": null
  }
  ```

- **错误场景**
  | 场景 | 错误信息 |
  | ---- | -------- |
  | 门店已被同仓库其他规则配置 | `门店 XXX 已在该仓库的其他规则中配置` |

- **Notes**
  - 同一仓库下每个门店只能关联一条规则，跨仓库不限制
  - 创建时会在事务内校验门店冲突，避免并发竞态

---

### 3.2 更新自动收货规则

- **Method & URL**
  ```http
  POST /shop/setting/auto_receipt/rule/update
  ```

- **Headers**
  | 名称          | 必填 | 示例             | 说明      |
  | ------------- | ---- | ---------------- | --------- |
  | Authorization | 是   | `Bearer {token}` | JWT Token |
  | Content-Type  | 是   | `application/json` | 内容类型 |

- **Request**

  ```json
  {
    "uuid": 100001,
    "delay_days": 5,
    "status": 1,
    "shop_uuids": [8267304538112001, 8267304538112003]
  }
  ```

  | 字段         | 类型     | 必填 | 说明                                              |
  | ------------ | -------- | ---- | ------------------------------------------------- |
  | `uuid`       | uint64   | 是   | 规则 UUID                                         |
  | `delay_days` | *int     | 否   | 延迟天数（≥0），不传则不修改                       |
  | `status`     | *int     | 否   | 状态（0=停用, 1=启用），不传则不修改               |
  | `shop_uuids` | uint64[] | 否   | 门店列表，**全量替换语义**（见下方说明）           |

- **Response**

  ```json
  {
    "code": 1,
    "message": "success",
    "data": null
  }
  ```

- **`shop_uuids` 语义说明**
  | 传参方式           | 行为                                          |
  | ------------------ | --------------------------------------------- |
  | 不传（`null`/不含）| 仅更新 `delay_days` / `status`，不变更门店    |
  | 传空数组 `[]`      | 删除整条规则（级联删除门店）                   |
  | 传非空数组         | **全量替换**：diff 当前与新列表，执行增删      |

- **错误场景**
  | 场景 | 错误信息 |
  | ---- | -------- |
  | 规则不存在 | `规则不存在` |
  | 新增门店已被其他规则配置 | `门店 XXX 已在该仓库的其他规则中配置` |

- **Notes**
  - 门店更新采用 diff 策略：只添加新门店、删除移除的门店，不影响未变更的门店
  - 传空数组等同于删除操作，会级联删除规则和关联门店

---

### 3.3 删除自动收货规则

- **Method & URL**
  ```http
  POST /shop/setting/auto_receipt/rule/delete
  ```

- **Headers**
  | 名称          | 必填 | 示例             | 说明      |
  | ------------- | ---- | ---------------- | --------- |
  | Authorization | 是   | `Bearer {token}` | JWT Token |
  | Content-Type  | 是   | `application/json` | 内容类型 |

- **Request**

  ```json
  {
    "uuids": [100001, 100002]
  }
  ```

  | 字段    | 类型     | 必填 | 说明                       |
  | ------- | -------- | ---- | -------------------------- |
  | `uuids` | uint64[] | 是   | 规则 UUID 列表，至少 1 个  |

- **Response**

  ```json
  {
    "code": 1,
    "message": "success",
    "data": null
  }
  ```

- **Notes**
  - 支持批量删除
  - 级联软删除规则关联的门店记录

---

### 3.4 获取规则列表

- **Method & URL**
  ```http
  GET /shop/setting/auto_receipt/rule/list
  ```

- **Headers**
  | 名称          | 必填 | 示例             | 说明      |
  | ------------- | ---- | ---------------- | --------- |
  | Authorization | 是   | `Bearer {token}` | JWT Token |

- **Request**
  | 字段                 | 类型   | 必填 | 来源  | 说明                    |
  | -------------------- | ------ | ---- | ----- | ----------------------- |
  | `warehouse_erp_code` | string | 否   | Query | 按仓库编码筛选          |

  ```
  GET /shop/setting/auto_receipt/rule/list?warehouse_erp_code=WH001
  ```

- **Response**

  ```json
  {
    "code": 1,
    "message": "success",
    "data": {
      "list": [
        {
          "uuid": 100001,
          "warehouse_erp_code": "WH001",
          "locale_warehouse_name": {
            "zh": "中央仓库",
            "en": "Central Warehouse",
            "th": "",
            "zh_tw": ""
          },
          "delay_days": 3,
          "status": 1,
          "shop_count": 2,
          "shops": [
            {
              "uuid": 200001,
              "shop_uuid": 8267304538112001,
              "shop_code": "S001",
              "shop_name": "门店A"
            },
            {
              "uuid": 200002,
              "shop_uuid": 8267304538112002,
              "shop_code": "S002",
              "shop_name": "门店B"
            }
          ]
        }
      ]
    }
  }
  ```

  **`list[]` 字段说明**

  | 字段                    | 类型            | 说明                     |
  | ----------------------- | --------------- | ------------------------ |
  | `uuid`                  | uint64          | 规则 UUID                |
  | `warehouse_erp_code`    | string          | 仓库 ERP 编码            |
  | `locale_warehouse_name` | LocaleResponse  | 仓库名称（多语言）       |
  | `delay_days`            | int             | 延迟天数                 |
  | `status`                | int             | 状态：0=停用, 1=启用     |
  | `shop_count`            | int             | 关联门店数量             |
  | `shops`                 | array           | 门店列表                 |

  **`shops[]` 字段说明**

  | 字段        | 类型   | 说明         |
  | ----------- | ------ | ------------ |
  | `uuid`      | uint64 | 关联记录UUID |
  | `shop_uuid` | uint64 | 门店 UUID    |
  | `shop_code` | string | 门店编码     |
  | `shop_name` | string | 门店名称     |

---

### 3.5 获取可选门店列表

- **Method & URL**
  ```http
  GET /shop/setting/auto_receipt/shop_list
  ```

- **Headers**
  | 名称          | 必填 | 示例             | 说明      |
  | ------------- | ---- | ---------------- | --------- |
  | Authorization | 是   | `Bearer {token}` | JWT Token |

- **Request**
  | 字段                 | 类型   | 必填 | 来源  | 说明                    |
  | -------------------- | ------ | ---- | ----- | ----------------------- |
  | `warehouse_erp_code` | string | 是   | Query | 仓库 ERP 编码           |

  ```
  GET /shop/setting/auto_receipt/shop_list?warehouse_erp_code=WH001
  ```

- **Response**

  ```json
  {
    "code": 1,
    "message": "success",
    "data": {
      "list": [
        {
          "uuid": 8267304538112001,
          "name": "门店A",
          "store_code": "S001",
          "status": 1,
          "disabled": false,
          "disabled_reason": ""
        },
        {
          "uuid": 8267304538112002,
          "name": "门店B",
          "store_code": "S002",
          "status": 1,
          "disabled": true,
          "disabled_reason": "该门店已在当前仓库的其他规则中配置"
        }
      ]
    }
  }
  ```

  **`list[]` 字段说明**

  | 字段              | 类型   | 说明                                    |
  | ----------------- | ------ | --------------------------------------- |
  | `uuid`            | uint64 | 门店 UUID                               |
  | `name`            | string | 门店名称                                |
  | `store_code`      | string | 门店编码                                |
  | `status`          | int    | 门店状态                                |
  | `disabled`        | bool   | 是否禁用选择                            |
  | `disabled_reason` | string | 禁用原因（已在其他规则中配置时显示）     |

- **Notes**
  - 返回当前总部下所有子门店（排除总部自身）
  - 已在同仓库其他规则中配置的门店标记为 `disabled: true`

---

### 3.6 获取自动收货记录列表

- **Method & URL**
  ```http
  GET /shop/setting/auto_receipt/log/list
  ```

- **Headers**
  | 名称          | 必填 | 示例             | 说明      |
  | ------------- | ---- | ---------------- | --------- |
  | Authorization | 是   | `Bearer {token}` | JWT Token |

- **Request**
  | 字段           | 类型  | 必填 | 来源  | 说明                                    |
  | -------------- | ----- | ---- | ----- | --------------------------------------- |
  | `page_no`      | int   | 否   | Query | 页码，默认 1                            |
  | `page_size`    | int   | 否   | Query | 每页大小，默认 20，最大 1100            |
  | `receipt_time` | int64 | 否   | Query | 按收货时间筛选（Unix 时间戳，精确到天） |

  ```
  GET /shop/setting/auto_receipt/log/list?page_no=1&page_size=20&receipt_time=1710288000
  ```

- **Response**

  ```json
  {
    "code": 1,
    "message": "success",
    "data": {
      "list": [
        {
          "uuid": 300001,
          "shop_company_uuid": 8267304538112001,
          "shop_name": "门店A",
          "receipt_order_uuid": 400001,
          "receipt_order_no": "RO20260313001",
          "receipt_erp_order_no": "ERP-RO-001",
          "receipt_time": 1710288000
        }
      ],
      "meta": {
        "page_no": 1,
        "page_size": 20,
        "total": 1
      }
    }
  }
  ```

  **`list[]` 字段说明**

  | 字段                   | 类型   | 说明              |
  | ---------------------- | ------ | ----------------- |
  | `uuid`                 | uint64 | 记录 UUID         |
  | `shop_company_uuid`    | uint64 | 门店 UUID         |
  | `shop_name`            | string | 门店名称          |
  | `receipt_order_uuid`   | uint64 | 收货单 UUID       |
  | `receipt_order_no`     | string | 收货单号          |
  | `receipt_erp_order_no` | string | ERP 收货单号      |
  | `receipt_time`         | int64  | 收货时间（时间戳）|

---

### 3.7 获取自动收货记录详情

- **Method & URL**
  ```http
  GET /shop/setting/auto_receipt/log/detail
  ```

- **Headers**
  | 名称          | 必填 | 示例             | 说明      |
  | ------------- | ---- | ---------------- | --------- |
  | Authorization | 是   | `Bearer {token}` | JWT Token |

- **Request**
  | 字段   | 类型   | 必填 | 来源  | 说明      |
  | ------ | ------ | ---- | ----- | --------- |
  | `uuid` | uint64 | 是   | Query | 记录 UUID |

  ```
  GET /shop/setting/auto_receipt/log/detail?uuid=300001
  ```

- **Response**

  返回对应门店的收货单详情，结构与现有收货单详情接口一致（`PurchaseReceiptOrderDetailResp`）：

  ```json
  {
    "code": 1,
    "message": "success",
    "data": {
      "uuid": 400001,
      "status": 1,
      "order_no": "RO20260313001",
      "erp_order_no": "ERP-RO-001",
      "purchase_order_no": "PO20260310001",
      "purchase_order_uuid": 500001,
      "num": 5,
      "expect_arrival_time": 1710201600,
      "purchase_time": 1710115200,
      "receive_time": 1710288000,
      "create_time": 1710288000,
      "supplier_name": "供应商A",
      "locale_warehouse_name": { "zh": "中央仓库", "en": "Central Warehouse", "th": "", "zh_tw": "" },
      "is_from_delivery_note": true,
      "is_auto_receipt": true,
      "items": [],
      "files": []
    }
  }
  ```

- **Notes**
  - 该接口会跨门店查询收货单（从 saas 主库日志中获取门店信息，再到对应门店库查询收货单详情）
  - 返回的收货单 `is_auto_receipt` 字段为 `true`

---

## 4. 修改接口说明

以下现有接口新增了 `is_auto_receipt` 字段：

### 4.1 收货单列表

**接口路径**：`GET /shop/purchase/receipt_order/list`

**变更**：`list[]` 中每条收货单记录新增字段：

| 字段              | 类型 | 说明                        |
| ----------------- | ---- | --------------------------- |
| `is_auto_receipt` | bool | 是否自动收货（`true`/`false`）|

### 4.2 收货单详情

**接口路径**：`GET /shop/purchase/receipt_order/detail`

**变更**：响应体新增字段：

| 字段              | 类型 | 说明                        |
| ----------------- | ---- | --------------------------- |
| `is_auto_receipt` | bool | 是否自动收货（`true`/`false`）|

### 4.3 采购单详情 — 收货清单

**接口路径**：`GET /shop/purchase/order/detail`（`with_receipt_list=true`）

**变更**：`receipt_list[].receipt_orders[]` 中每条收货单新增字段：

| 字段              | 类型 | 说明                        |
| ----------------- | ---- | --------------------------- |
| `is_auto_receipt` | bool | 是否自动收货（`true`/`false`）|

---

## 5. 数据模型

### 5.1 ttpos_auto_receipt_rule（saas 主库）

| 字段                       | 类型          | 说明         |
| -------------------------- | ------------- | ------------ |
| `uuid`                     | bigint        | 主键         |
| `headquarter_company_uuid` | bigint        | 总部 UUID    |
| `warehouse_erp_code`       | varchar(100)  | 仓库 ERP 编码|
| `warehouse_name`           | varchar(1000) | 仓库名称     |
| `delay_days`               | int           | 延迟天数     |
| `status`                   | tinyint(1)    | 0=停用 1=启用|

### 5.2 ttpos_auto_receipt_rule_shop（saas 主库）

| 字段                       | 类型         | 说明          |
| -------------------------- | ------------ | ------------- |
| `uuid`                     | bigint       | 主键          |
| `rule_uuid`                | bigint       | 规则 UUID     |
| `shop_company_uuid`        | bigint       | 门店 UUID     |
| `headquarter_company_uuid` | bigint       | 总部 UUID     |

### 5.3 ttpos_auto_receipt_log（saas 主库）

| 字段                       | 类型         | 说明          |
| -------------------------- | ------------ | ------------- |
| `uuid`                     | bigint       | 主键          |
| `headquarter_company_uuid` | bigint       | 总部 UUID     |
| `shop_company_uuid`        | bigint       | 门店 UUID     |
| `receipt_order_uuid`       | bigint       | 收货单 UUID   |
| `receipt_order_no`         | varchar(100) | 收货单号      |
| `receipt_erp_order_no`     | varchar(100) | ERP 收货单号  |
| `receipt_time`             | int          | 收货时间      |

### 5.4 purchase_receipt_order 新增字段（门店库）

| 字段              | 类型       | 说明                     |
| ----------------- | ---------- | ------------------------ |
| `is_auto_receipt` | tinyint(1) | 0=手动收货, 1=自动收货   |
