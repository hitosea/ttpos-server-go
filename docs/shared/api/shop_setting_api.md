# 商家端业务设置 API 文档

> 📚 **受众**: 前端开发者、API 对接人员  
> 📖 **用途**: 商家端业务设置相关 API 接口文档

---

## 元信息

| 字段      | 内容                                    |
| --------- | --------------------------------------- |
| 模块      | `shop_setting`                          |
| 版本      | `v2.1.0`                                |
| 更新时间  | `2025-12-05`                            |
| 负责人    | `@王昱`                                 |
| 关联 Spec | `story-main-order-item-remark-reason-management` |

---

## 1. 概述

商家端业务设置模块提供门店业务配置相关的 API 接口，包括：
- **整单备注原因管理**：管理订单整单备注的预设原因
- **单品备注原因管理**：管理订单单品备注的预设原因（新增）
- **免单原因管理**：管理免单操作的预设原因
- **退菜原因管理**：管理退菜操作的预设原因

**终端支持**：Shop（商家端后台）

**接口数量**：`P0: 4 / P1: 0 / P2: 0`（单品备注原因管理）

**认证方式**：JWT Token

---

## 2. 快速索引

| 级别 | 接口                    | 方法   | 路径                                      | 描述             |
| ---- | ----------------------- | ------ | ----------------------------------------- | ---------------- |
| P0   | `GetOrderItemRemark`    | GET    | `/shop/setting/order_item_remark`        | 获取单品备注列表 |
| P0   | `AddOrderItemRemark`    | POST   | `/shop/setting/order_item_remark/add`    | 新增单品备注     |
| P0   | `EditOrderItemRemark`   | POST   | `/shop/setting/order_item_remark/edit`   | 编辑单品备注     |
| P0   | `DeleteOrderItemRemark` | DELETE | `/shop/setting/order_item_remark`        | 删除单品备注     |

---

## 3. 接口详情

### 3.1 获取单品备注原因列表

- **Method & URL**
  ```http
  GET /shop/setting/order_item_remark
  ```

- **Headers**
  | 名称           | 必填 | 示例                    | 说明         |
  | -------------- | ---- | ----------------------- | ------------ |
  | Authorization  | 是   | `Bearer {token}`        | JWT Token    |
  | Content-Type   | 是   | `application/json`       | 内容类型     |
  | Accept-Language| 否   | `zh-CN`                 | 语言偏好     |

- **Request**
  无请求参数

- **Response**

  ```json
  {
    "code": 1,
    "message": "success",
    "data": {
      "list": [
        {
          "uuid": 2001,
          "locale_name": {
            "zh": "不要香菜",
            "en": "No Cilantro",
            "th": "",
            "zh_tw": "",
            "ja": "",
            "ko": "",
            "my": "",
            "tr": "",
            "sv": ""
          }
        },
        {
          "uuid": 2002,
          "locale_name": {
            "zh": "不要葱",
            "en": "No Scallion",
            "th": "",
            "zh_tw": "",
            "ja": "",
            "ko": "",
            "my": "",
            "tr": "",
            "sv": ""
          }
        }
      ]
    }
  }
  ```

  | 字段        | 类型   | 说明                     |
  | ----------- | ------ | ------------------------ |
  | `list`      | array  | 单品备注原因列表         |
  | `uuid`      | uint64 | 备注原因 UUID            |
  | `locale_name` | object | 多语言名称对象           |
  | `locale_name.zh` | string | 中文名称               |
  | `locale_name.en` | string | 英文名称               |
  | `locale_name.th` | string | 泰文名称               |
  | `locale_name.zh_tw` | string | 繁体中文名称         |
  | `locale_name.ja` | string | 日文名称               |
  | `locale_name.ko` | string | 韩文名称               |
  | `locale_name.my` | string | 缅甸文名称             |
  | `locale_name.tr` | string | 土耳其文名称           |
  | `locale_name.sv` | string | 瑞典文名称             |

- **错误码**
  | code | message | 场景               |
  | ---- | ------- | ------------------ |
  | 0    | 获取失败 | 系统错误           |
  | 1    | success | 成功               |

- **示例**

  ```bash
  curl -X GET https://{host}/shop/setting/order_item_remark \
    -H "Authorization: Bearer <token>" \
    -H "Content-Type: application/json"
  ```

- **备注**
  - 返回列表按创建时间倒序排列
  - 只返回未删除（`delete_time = 0`）的记录
  - 多语言名称根据门店语言设置返回对应语言的内容

---

### 3.2 新增单品备注原因

- **Method & URL**
  ```http
  POST /shop/setting/order_item_remark/add
  ```

- **Headers**
  | 名称          | 必填 | 示例                  | 说明       |
  | ------------- | ---- | --------------------- | ---------- |
  | Authorization | 是   | `Bearer {token}`      | JWT Token  |
  | Content-Type  | 是   | `application/json`    | 内容类型   |

- **Request**

  ```json
  {
    "locale_name": {
      "zh": "不要香菜",
      "en": "No Cilantro"
    }
  }
  ```

  | 字段        | 类型   | 必填 | 说明                     | 校验                           |
  | ----------- | ------ | ---- | ------------------------ | ------------------------------ |
  | `locale_name` | object | 是   | 多语言名称对象           | 必填                           |
  | `locale_name.zh` | string | 是* | 中文名称               | 根据门店语言设置，必填语言必填 |
  | `locale_name.en` | string | 是* | 英文名称               | 根据门店语言设置，必填语言必填 |
  | `locale_name.th` | string | 否  | 泰文名称               | 可选                           |
  | `locale_name.zh_tw` | string | 否 | 繁体中文名称         | 可选                           |
  | `locale_name.ja` | string | 否  | 日文名称               | 可选                           |
  | `locale_name.ko` | string | 否  | 韩文名称               | 可选                           |
  | `locale_name.my` | string | 否  | 缅甸文名称             | 可选                           |
  | `locale_name.tr` | string | 否  | 土耳其文名称           | 可选                           |
  | `locale_name.sv` | string | 否  | 瑞典文名称             | 可选                           |

  **说明**：
  - `locale_name` 中必须包含门店设置的所有语言
  - 每个语言名称的字数（非字符）不能超过 100 字
  - 单品备注原因总数不能超过 100 个

- **Response**

  ```json
  {
    "code": 1,
    "message": "新增成功",
    "data": {}
  }
  ```

- **错误码**
  | code | message                    | 场景                         |
  | ---- | -------------------------- | ---------------------------- |
  | 0    | 多语言名称不完整           | 缺少门店设置中的必填语言     |
  | 0    | 字数不能超过100个字        | 某个语言名称超过 100 字      |
  | 0    | 单品备注数量不能超过100个  | 当前数量已达到 100 个        |
  | 0    | 新增失败                   | 系统错误                     |
  | 1    | 新增成功                   | 成功                         |

- **示例**

  ```bash
  curl -X POST https://{host}/shop/setting/order_item_remark/add \
    -H "Authorization: Bearer <token>" \
    -H "Content-Type: application/json" \
    -d '{
      "locale_name": {
        "zh": "不要香菜",
        "en": "No Cilantro"
      }
    }'
  ```

- **备注**
  - 新增时会自动创建多语言名称记录
  - 使用事务确保数据一致性
  - 权限验证与整单备注一致

---

### 3.3 编辑单品备注原因

- **Method & URL**
  ```http
  POST /shop/setting/order_item_remark/edit
  ```

- **Headers**
  | 名称          | 必填 | 示例                  | 说明       |
  | ------------- | ---- | --------------------- | ---------- |
  | Authorization | 是   | `Bearer {token}`      | JWT Token  |
  | Content-Type  | 是   | `application/json`    | 内容类型   |

- **Request**

  ```json
  {
    "uuid": 2001,
    "locale_name": {
      "zh": "不要葱",
      "en": "No Scallion"
    }
  }
  ```

  | 字段        | 类型   | 必填 | 说明                     | 校验                           |
  | ----------- | ------ | ---- | ------------------------ | ------------------------------ |
  | `uuid`      | uint64 | 是   | 备注原因 UUID            | 必填，必须存在                 |
  | `locale_name` | object | 是   | 多语言名称对象           | 必填                           |
  | `locale_name.zh` | string | 是* | 中文名称               | 根据门店语言设置，必填语言必填 |
  | `locale_name.en` | string | 是* | 英文名称               | 根据门店语言设置，必填语言必填 |
  | `locale_name.th` | string | 否  | 泰文名称               | 可选                           |
  | `locale_name.zh_tw` | string | 否 | 繁体中文名称         | 可选                           |
  | `locale_name.ja` | string | 否  | 日文名称               | 可选                           |
  | `locale_name.ko` | string | 否  | 韩文名称               | 可选                           |
  | `locale_name.my` | string | 否  | 缅甸文名称             | 可选                           |
  | `locale_name.tr` | string | 否  | 土耳其文名称           | 可选                           |
  | `locale_name.sv` | string | 否  | 瑞典文名称             | 可选                           |

  **说明**：
  - `uuid` 必须存在且未删除
  - `locale_name` 中必须包含门店设置的所有语言
  - 每个语言名称的字数（非字符）不能超过 100 字

- **Response**

  ```json
  {
    "code": 1,
    "message": "编辑成功",
    "data": {}
  }
  ```

- **错误码**
  | code | message                | 场景                     |
  | ---- | ---------------------- | ------------------------ |
  | 0    | 多语言名称不完整       | 缺少门店设置中的必填语言 |
  | 0    | 字数不能超过100个字    | 某个语言名称超过 100 字  |
  | 0    | 记录不存在             | UUID 不存在或已删除      |
  | 0    | 编辑失败               | 系统错误                 |
  | 1    | 编辑成功               | 成功                     |

- **示例**

  ```bash
  curl -X POST https://{host}/shop/setting/order_item_remark/edit \
    -H "Authorization: Bearer <token>" \
    -H "Content-Type: application/json" \
    -d '{
      "uuid": 2001,
      "locale_name": {
        "zh": "不要葱",
        "en": "No Scallion"
      }
    }'
  ```

- **备注**
  - 编辑时会更新多语言名称记录
  - 使用事务确保数据一致性
  - 权限验证与整单备注一致

---

### 3.4 删除单品备注原因

- **Method & URL**
  ```http
  DELETE /shop/setting/order_item_remark
  ```

- **Headers**
  | 名称          | 必填 | 示例                  | 说明       |
  | ------------- | ---- | --------------------- | ---------- |
  | Authorization | 是   | `Bearer {token}`      | JWT Token  |
  | Content-Type  | 是   | `application/json`    | 内容类型   |

- **Request**

  ```json
  {
    "uuid": 2001
  }
  ```

  | 字段 | 类型   | 必填 | 说明          | 校验           |
  | ---- | ------ | ---- | ------------- | -------------- |
  | `uuid` | uint64 | 是   | 备注原因 UUID | 必填，必须存在 |

- **Response**

  ```json
  {
    "code": 1,
    "message": "删除成功",
    "data": {}
  }
  ```

- **错误码**
  | code | message    | 场景                 |
  | ---- | ---------- | -------------------- |
  | 0    | 记录不存在 | UUID 不存在或已删除  |
  | 0    | 删除失败   | 系统错误             |
  | 1    | 删除成功   | 成功                 |

- **示例**

  ```bash
  curl -X DELETE https://{host}/shop/setting/order_item_remark \
    -H "Authorization: Bearer <token>" \
    -H "Content-Type: application/json" \
    -d '{
      "uuid": 2001
    }'
  ```

- **备注**
  - 删除采用软删除方式，设置 `delete_time` 字段
  - 同时软删除关联的多语言名称记录
  - 不影响历史订单数据
  - 权限验证与整单备注一致

---

## 4. 业务规则

### 4.1 数量限制

- 单品备注原因总数不能超过 **100 个**
- 当达到上限时，新增操作会返回错误："单品备注数量不能超过100个"

### 4.2 多语言验证

- 必须包含门店设置中的所有语言
- 如果门店设置了中文和英文，则 `locale_name` 中必须同时包含 `zh` 和 `en`
- 缺少必填语言时返回错误："多语言名称不完整"

### 4.3 字数限制

- 每个语言名称的字数（非字符）不能超过 **100 字**
- 超过限制时返回错误："字数不能超过100个字"

### 4.4 权限验证

- 权限处理逻辑与整单备注一致
- 需要相应的业务设置权限

---

## 5. 数据模型

### 5.1 OrderItemRemark（单品备注原因）

| 字段                      | 类型   | 说明                 |
| ------------------------- | ------ | -------------------- |
| `id`                      | uint   | 主键（自增）         |
| `uuid`                    | uint64 | UUID（唯一索引）     |
| `name`                    | string | 名称（中文）         |
| `multi_language_name_uuid` | uint64 | 多语言名称 UUID      |
| `create_time`             | int64  | 创建时间（时间戳）   |
| `update_time`             | int64  | 更新时间（时间戳）   |
| `delete_time`             | int64  | 删除时间（时间戳）   |

### 5.2 MultiLanguageName（多语言名称）

| 字段         | 类型   | 说明               |
| ------------ | ------ | ------------------ |
| `uuid`       | uint64 | UUID               |
| `zh_name`    | string | 中文名称           |
| `en_name`    | string | 英文名称           |
| `th_name`    | string | 泰文名称           |
| `zh_tw_name` | string | 繁体中文名称       |
| `ja_name`    | string | 日文名称           |
| `ko_name`    | string | 韩文名称           |
| `my_name`    | string | 缅甸文名称         |
| `tr_name`    | string | 土耳其文名称       |
| `sv_name`    | string | 瑞典文名称         |

---

## 6. 依赖与配置

- **数据库表**：`ttpos_order_item_remark`、`multi_language_names`
- **服务依赖**：Setting Service（获取门店语言设置）
- **缓存**：无特殊缓存要求

---

## 7. 测试与验证

- **自动化测试位置**：`main/app/api/v1/shop/shop_setting_test.go`
- **测试用例数**：10 个集成测试用例
- **测试覆盖**：
  - ✅ GET 接口（获取列表、空列表）
  - ✅ POST 接口（新增、编辑、参数验证、数量限制、多语言验证、字数限制）
  - ✅ DELETE 接口（删除、记录不存在）
  - ✅ 端到端流程测试

---

## 8. 变更记录

| 日期       | 版本   | 说明                 | 负责人 |
| ---------- | ------ | -------------------- | ------ |
| 2025-12-05 | v2.1.0 | 新增单品备注原因管理 | @王昱  |

---

## 9. 相关文档

- **需求文档**：`docs/shared/specs/active/story-main-order-item-remark-reason-management/requirements.md`
- **设计文档**：`docs/shared/specs/active/story-main-order-item-remark-reason-management/design.md`
- **任务分解**：`docs/shared/specs/active/story-main-order-item-remark-reason-management/tasks.md`
- **API 规范**：`docs/shared/api/conventions.md`

---

## 10. Graphiti & 活动日志

- Related Episode: `[待补充]`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 模板引用：`docs/agent/templates/graphiti-episode.md`

---

**最后更新**：2025-12-05  
**维护者**：TTPOS Team

