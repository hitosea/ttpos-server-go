# 需求文档：新管理端-商品管理（属性、加料限制）

## 文档信息

| 项目       | 内容                                                          |
| ---------- | ------------------------------------------------------------- |
| 需求名称   | 新管理端-商品管理（属性、加料限制）                           |
| DooTask ID | 37946                                                         |
| 创建时间   | 2025-12-22                                                    |
| 版本       | v2.12.0                                                       |
| 关键字     | 商品管理, 属性设置, 加料设置, 套餐分组, 可选范围             |

---

## 1. 背景和目标

### 1.1 业务背景

当前商品管理系统在以下方面存在不足：

1. **属性/加料设置不灵活**：当前只能设置"必选"和"最大可选"，无法设置"最小可选"，不能表达"至少选1个，最多选3个"的业务需求
2. **套餐分组设置单一**：可选分组只能设置"可选数量"，无法设置可选范围
3. **总部数据编辑权限不清晰**：来源总部的数据编辑权限需要明确界定

### 1.2 目标

1. **优化属性/加料设置**：支持可选范围（最小-最大）设置，替代原有的"必选+最大可选"模式
2. **优化套餐分组设置**：支持可选范围（最小-最大）设置
3. **明确编辑权限**：总部数据允许修改外卖渠道的价格
4. **数据兼容**：旧数据能够正确转换为新的可选范围模式

---

## 2. 用户故事

### 2.1 总部数据编辑权限

**作为**店铺管理员  
**我想**能够修改来源总部数据的外卖渠道价格  
**以便于**根据本店实际情况调整价格策略

**验收标准：**

- AC1：来源总部的商品数据，允许修改外卖渠道的价格
- AC2：来源总部的套餐数据，允许修改外卖渠道的套餐价格和分组商品价格
- AC3：来源总部的规格数据，允许修改外卖渠道的规格价格

### 2.2 套餐分组可选范围设置

**作为**商品配置人员  
**我想**为可选套餐分组设置可选范围（最小-最大）  
**以便于**实现"至少选2个，最多选5个"等业务规则

**验收标准：**

- AC1：可选分组类型显示"本组可选范围"设置，默认值为 1-1
- AC2：可选范围前后两个数字，后者必须大于或等于前者
- AC3：数字可以为 0（表示可不选）
- AC4：当前者大于后者时，提示："分组{N}最大可选不可小于最小可选"
- AC5：分组数量上限为 100
- AC6：旧数据转换规则：
  - 旧数据中可选分组：新增 `optional_min_count=1`，`optional_count` 保持不变作为最大可选数量
  - 固定分组：`optional_min_count` 设置为分组商品数量，`optional_count` 保持不变

### 2.3 属性设置可选范围

**作为**商品配置人员  
**我想**为属性组设置可选范围（最小-最大）  
**以便于**实现"至少选1个，最多选3个"等业务规则

**验收标准：**

- AC1：属性设置显示"可选范围"，默认值为 0-属性值数量
- AC2：可选范围前后两个数字，后者必须大于或等于前者
- AC3：当前者大于后者时，提示："最大可选不可小于最小可选"
- AC4：属性组数量上限为 100
- AC5：旧数据转换规则：
  - 不开启必选，不设置最大可选 = 可选 0-属性值数量
  - 开启必选 = 最小值为 1
  - 开启最大可选 = 最大值为具体的值

### 2.4 加料设置可选范围

**作为**商品配置人员  
**我想**为加料设置可选范围（最小-最大）  
**以便于**实现"至少选1个小料，最多选3个"等业务规则

**验收标准：**

- AC1：加料设置显示"可选范围"，默认值为 0-加料值数量
- AC2：可选范围前后两个数字，后者必须大于或等于前者
- AC3：当前者大于后者时，提示："最大可选不可小于最小可选"
- AC4：加料值数量上限为 100
- AC5：旧数据转换规则：
  - 不开启必选，不设置最大可选 = 可选 0-加料值数量
  - 开启必选 = 最小值为 1
  - 开启最大可选 = 最大值为具体的值

---

## 3. 功能需求

### 3.1 总部数据编辑权限

#### 3.2.1 权限规则

- 字段标识：`headquarter_uuid != 0` 表示来源总部
- 允许编辑的内容：
  - 外卖渠道价格（单规格、多规格）
  - 套餐外卖渠道价格
  - 套餐分组商品外卖渠道价格

### 3.2 套餐分组可选范围

#### 3.3.1 数据库字段变更

需要在 `ttpos_product_package_group` 表中新增字段和修改注释：

```sql
-- 新增字段
`optional_min_count` INT NOT NULL DEFAULT 0 COMMENT '最小可选数量'

-- 修改原有字段注释（字段名保持不变）
-- optional_count 字段注释修改为：最大可选数量
COMMENT ON COLUMN ttpos_product_package_group.optional_count IS '最大可选数量，表示本组商品中最多可以选择多少个商品';
```

**说明：**
- `optional_count` 字段名保持不变，只修改注释，现在表示"最大可选数量"
- `optional_min_count` 为新增字段，表示"最小可选数量"

#### 3.3.2 验证规则

- `optional_min_count >= 0`
- `optional_count >= optional_min_count` （optional_count 作为最大可选数量）
- `optional_count <= 分组商品数量`
- 分组数量上限：100

#### 3.3.3 旧数据兼容

旧数据转换逻辑：

```
IF group_type == 1 (可选分组) THEN
  optional_min_count = 1 (默认至少选1个)
  optional_count 保持不变 (继续作为最大可选数量使用)
ELSE (固定分组)
  optional_min_count = 分组商品数量
  optional_count 保持不变 (等于分组商品数量)
END IF
```

### 3.3 属性设置可选范围

#### 3.4.1 数据库字段变更

需要在 `ttpos_product_package_attribute_group` 表中新增或修改字段：

```sql
-- 新增字段
`min_selection` INT NOT NULL DEFAULT 0 COMMENT '最小选择数量'
-- 将原有 is_must 废弃，max_selection 保留
`max_selection` INT NOT NULL DEFAULT 0 COMMENT '最大选择数量'
```

#### 3.4.2 验证规则

- `min_selection >= 0`
- `max_selection >= min_selection`
- `max_selection <= 属性值数量`
- 属性组数量上限：100

#### 3.4.3 旧数据兼容

旧数据转换逻辑：

```
IF is_must == 1 THEN
  min_selection = 1
ELSE
  min_selection = 0
END IF

IF max_selection > 0 THEN
  max_selection = max_selection (保留原值)
ELSE
  max_selection = 属性值数量
END IF
```

### 3.4 加料设置可选范围

#### 3.5.1 数据库字段变更

需要在 `ttpos_product_package` 表中新增或修改字段：

```sql
-- 新增字段
`sauce_min_selection` INT NOT NULL DEFAULT 0 COMMENT '小料最小选择数量'
-- 将原有 sauce_required 废弃，sauce_max_selection 保留
`sauce_max_selection` INT NOT NULL DEFAULT 0 COMMENT '小料最大选择数量'
```

#### 3.5.2 验证规则

- `sauce_min_selection >= 0`
- `sauce_max_selection >= sauce_min_selection`
- `sauce_max_selection <= 加料值数量`
- 加料值数量上限：100

#### 3.5.3 旧数据兼容

旧数据转换逻辑：

```
IF sauce_required == 1 THEN
  sauce_min_selection = 1
ELSE
  sauce_min_selection = 0
END IF

IF sauce_max_selection > 0 THEN
  sauce_max_selection = sauce_max_selection (保留原值)
ELSE
  sauce_max_selection = 加料值数量
END IF
```

---

## 4. 非功能需求

### 4.1 性能要求

- 删除操作检查查询时间 < 200ms
- API响应时间 < 500ms

### 4.2 兼容性要求

#### 4.2.1 版本兼容性

**当前开发版本：** v2.12.0  
**需要兼容版本：** v2.11.x

| 场景 | 要求 | 说明 |
|------|------|------|
| v2.11客户端 → v2.12后端 | 必须兼容 | 旧客户端不传新字段时，后端自动设置默认值 |
| v2.12客户端 → v2.12后端 | 必须兼容 | 新客户端传新字段，后端优先使用新字段 |
| 查询接口 | 同时返回新旧字段 | 确保新旧客户端都能正常工作 |

#### 4.2.2 API兼容性

**新旧字段对照：**

| 功能 | 旧字段 | 新字段 | 兼容策略 |
|------|--------|--------|---------|
| 属性是否必选 | `is_must` (0/1) | `min_selection` (≥0) | 后端自动转换，查询同时返回 |
| 小料是否必选 | `sauce_required` (0/1) | `sauce_min_selection` (≥0) | 后端自动转换，查询同时返回 |
| 套餐分组最小可选 | 无 | `optional_min_count` (≥0) | 新增字段，旧客户端使用默认值 |
| 套餐分组最大可选 | `optional_count` | `optional_count` | 字段名不变，语义更新 |

**兼容性处理：**

1. **添加/编辑商品接口**
   - 接受新旧字段同时传递
   - 优先使用新字段
   - 旧字段自动转换为新字段
   - v2.11客户端不传新字段时，自动设置合理默认值

2. **查询商品接口**
   - 响应同时包含新旧字段
   - 旧字段从新字段自动计算：
     - `is_must = (min_selection > 0) ? 1 : 0`
     - `sauce_required = (sauce_min_selection > 0) ? 1 : 0`

3. **字段废弃时间表**
   - v2.12: 废弃 `is_must` 和 `sauce_required`，但保留并响应
   - v2.14: 正式移除废弃字段（保留2个大版本缓冲期）

#### 4.2.3 数据兼容性

- 旧数据必须能够正确转换为新的可选范围模式
- 数据迁移脚本必须提供回滚方案
- 迁移后数据验证通过率 > 99.9%

#### 4.2.4 接口响应兼容性

**商品详情接口（GetProductDetail）响应字段：**

查询商品详情时，响应需要同时包含新旧字段，确保新旧客户端都能正常工作：

1. **小料列表（Sauces）**
   - 旧字段：`is_must` (boolean)
   - 新字段：`min_select` (int), `max_select` (int)
   - 转换规则：`is_must = (min_select > 0)`

2. **属性组列表（AttributeGroups）**
   - 旧字段：`is_must` (boolean)
   - 新字段：`min_select` (uint), `max_select` (uint)
   - 转换规则：`is_must = (min_select > 0)`

3. **套餐分组列表（PackageSubProductGroups）**
   - 新字段：`optional_min_count` (int), `optional_count` (int)
   - 注意：`optional_count` 字段名保持不变，但语义变为"最大可选数量"

**响应示例：**

```json
{
  "sauces": {
    "list": [...],
    "is_must": false,      // 旧字段：兼容v2.11
    "min_select": 0,       // 新字段
    "max_select": 3        // 新字段
  },
  "attribute_groups": {
    "list": [
      {
        "uuid": 123,
        "locale_name": {...},
        "is_must": true,    // 旧字段：兼容v2.11
        "min_select": 1,    // 新字段
        "max_select": 3,    // 新字段
        "attributes": {...}
      }
    ]
  },
  "package_sub_product_groups": {
    "list": [
      {
        "uuid": 456,
        "locale_name": {...},
        "group_type": 1,
        "optional_min_count": 2,  // 新字段：最小可选
        "optional_count": 5,      // 字段名不变，语义为最大可选
        "products": {...}
      }
    ]
  }
}
```

详细的版本兼容性设计请参考：[VERSION_COMPATIBILITY.md](./VERSION_COMPATIBILITY.md)

### 4.3 安全性要求

- 删除操作需要二次确认
- 总部数据编辑权限需要严格校验

---

## 5. 约束和假设

### 5.1 约束

1. 数据库字段变更需要通过迁移脚本完成
2. 旧数据转换需要在迁移脚本中完成
3. API需要保持向后兼容

### 5.2 假设

1. 外卖订单状态定义已明确
2. 总部数据标识（headquarter_uuid）准确可靠
3. API字段定义清晰，便于后续前端对接
4. **v2.11客户端将在v2.14之前逐步升级到v2.12+**
5. **废弃字段保留2个大版本后可安全移除**

---

## 6. 相关文档

- DooTask 任务：#37946
- 数据库规范：`.cursor/rules/database.mdc`
- API 设计规范：`.cursor/rules/api.mdc`
- Go Main 规范：`.cursor/rules/go-main.mdc`
- **版本兼容性设计：** [VERSION_COMPATIBILITY.md](./VERSION_COMPATIBILITY.md)

---

## 7. 变更历史

| 版本 | 日期       | 变更人 | 变更内容                                                      |
| ---- | ---------- | ------ | ------------------------------------------------------------- |
| 1.0  | 2025-12-22 | 曾振华 | 创建需求文档                                                   |
| 1.1  | 2025-12-22 | 曾振华 | 修改套餐分组字段方案：保持 optional_count 字段名不变，只修改注释 |
| 1.2  | 2025-12-22 | 曾振华 | 移除前端相关需求，聚焦后端功能需求                              |
| 1.3  | 2025-12-22 | 曾振华 | 新增详细的版本兼容性要求（v2.11 ↔ v2.12）                      |
| 1.4  | 2025-12-23 | AI     | 补充商品详情接口响应字段兼容性说明                             |
| 1.5  | 2025-12-23 | AI     | 修复代码中的上限验证：属性组/加料/套餐分组上限统一为100个       |
| 1.6  | 2025-12-24 | AI     | 移除商品删除限制需求，允许直接删除商品和规格                    |
| 1.7  | 2025-12-24 | 曾振华 | 补充子店同步总店商品字段（子任务）                              |

