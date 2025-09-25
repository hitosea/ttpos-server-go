# 结账单模板 (Statement Order Template) 详细说明

## 概述

`statement_order_tmp.json` 是一个用于生成餐饮收银结账单的图片打印模板配置文件。该模板支持多语言显示，包含完整的订单信息、商品详情、金额计算、支付信息等，适用于餐饮行业的收银系统。

## 模板基本信息

```json
{
    "metadata": {
        "name": "statement_order",           // 模板名称
        "description": "结账单模板",          // 模板描述
        "paper_width": 80,                  // 纸张宽度（毫米），可选项：58/80/0，为0时使用后台设置
        "version": "1.0",                   // 模板版本
        "thousandth": true                  // 启用千分位格式化
    }
}
```

### 纸张宽度说明

`paper_width` 字段用于指定打印纸张的宽度，支持以下选项：

- **58**: 58毫米纸张宽度，适用于小票打印机
- **80**: 80毫米纸张宽度，适用于标准热敏打印机（推荐）
- **0**: 使用后台系统设置，当设置为0时，系统会使用后台配置的默认纸张宽度

**选择建议：**
- 对于餐饮收银小票，推荐使用 **80毫米** 纸张宽度
- 对于便携式小票打印机，可选择 **58毫米** 纸张宽度
- 如果需要动态调整，可设置为 **0** 使用后台配置

## 全局字段说明

### 模板元数据字段 (Metadata Fields)

| 字段名 | 类型 | 必填 | 说明 | 可选项 |
|--------|------|------|------|--------|
| `name` | string | 是 | 模板名称，用于标识模板 | 任意字符串 |
| `description` | string | 否 | 模板描述，说明模板用途 | 任意字符串 |
| `paper_width` | int | 是 | 纸张宽度（毫米） | 58, 80, 0 |
| `version` | string | 否 | 模板版本号 | 任意版本号格式 |
| `thousandth` | boolean | 否 | 是否启用千分位格式化 | true, false |

### 块基础字段 (Block Base Fields)

| 字段名 | 类型 | 必填 | 说明 | 可选项 |
|--------|------|------|------|--------|
| `block_id` | string | 是 | 块唯一标识符，用于数据绑定 | 任意字符串，建议使用点号分隔 |
| `block_type` | string | 是 | 块类型，决定块的显示方式 | value, label, label:value, array, img, qrcode, barcode, blank_line, column |
| `block_label` | object/string | 否 | 块标签，支持多语言 | 多语言对象或字符串 |
| `block_after_label` | object/string | 否 | 块后标签 | 多语言对象或字符串 |
| `block_expand_labels` | array | 否 | 块扩展标签，支持条件显示 | 扩展标签对象数组 |
| `block_attr` | object | 否 | 块属性，控制显示样式 | 属性对象 |
| `rows` | array | 否 | 嵌套行，用于复杂布局 | 行数组 |
| `conditions` | array | 否 | 条件显示规则 | 条件对象数组 |

### 块类型说明 (Block Types)

| 类型 | 说明 | 用途 | 数据绑定 |
|------|------|------|----------|
| `value` | 值类型 | 直接显示数据源中的值 | 通过block_id绑定 |
| `label` | 标签类型 | 显示固定的标签文本 | 不绑定数据 |
| `label:value` | 标签值类型 | 显示标签和对应的值 | 通过block_id绑定 |
| `label:auto:value` | 自动标签值类型 | 自动显示标签和对应的值 | 通过block_id绑定 |
| `array` | 数组类型 | 循环显示数组中的每个元素 | 绑定数组字段 |
| `img` | 图片类型 | 显示图片内容 | 绑定图片路径字段 |
| `qrcode` | 二维码类型 | 显示二维码 | 绑定二维码内容字段 |
| `barcode` | 条形码类型 | 显示条形码 | 绑定条形码内容字段 |
| `blank_line` | 空行类型 | 添加空行 | 不绑定数据 |
| `column` | 列类型 | 列式布局显示 | 绑定列数据 |

### 块属性字段 (Block Attributes)

| 字段名 | 类型 | 必填 | 说明 | 可选项 |
|--------|------|------|------|--------|
| `font_size` | int | 否 | 字体大小（像素） | 12-32，推荐值：12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32 |
| `align` | string | 否 | 对齐方式 | left, center, right |
| `font_bold` | boolean | 否 | 是否粗体 | true, false |
| `dividing_line` | boolean | 否 | 是否显示分割线 | true, false |
| `leading_blank_lines` | int | 否 | 前导空行数 | 0-10 |
| `trailing_blank_lines` | int | 否 | 后置空行数 | 0-10 |
| `width` | int/object | 否 | 宽度百分比或多语言宽度 | 0-100 或 多语言宽度对象 |
| `line_height` | int | 否 | 行高（像素） | 20-50 |
| `show_currency_unit` | boolean | 否 | 是否显示货币单位 | true, false |

### 条件操作符 (Condition Operators)

| 操作符 | 说明 | 适用数据类型 | 示例 |
|--------|------|-------------|------|
| `==` | 等于 | 所有类型 | `"field": "status", "operator": "==", "value": "active"` |
| `!=` | 不等于 | 所有类型 | `"field": "status", "operator": "!=", "value": "inactive"` |
| `>` | 大于 | 数字类型 | `"field": "amount", "operator": ">", "value": 0` |
| `>=` | 大于等于 | 数字类型 | `"field": "amount", "operator": ">=", "value": 100` |
| `<` | 小于 | 数字类型 | `"field": "amount", "operator": "<", "value": 1000` |
| `<=` | 小于等于 | 数字类型 | `"field": "amount", "operator": "<=", "value": 500` |
| `contains` | 包含 | 字符串类型 | `"field": "name", "operator": "contains", "value": "test"` |
| `not_contains` | 不包含 | 字符串类型 | `"field": "name", "operator": "not_contains", "value": "test"` |
| `empty` | 为空 | 所有类型 | `"field": "description", "operator": "empty", "value": 0` |
| `not_empty` | 不为空 | 所有类型 | `"field": "description", "operator": "not_empty", "value": 0` |

### 多语言支持 (Multi-language Support)

| 语言代码 | 语言名称 | 说明 |
|----------|----------|------|
| `zh` | 中文 | 简体中文 |
| `en` | 英文 | English |
| `ko` | 韩文 | 한국어 |
| `my` | 缅甸文 | မြန်မာ |
| `tr` | 土耳其文 | Türkçe |
| `de` | 德文 | Deutsch |
| `sv` | 瑞典文 | Svenska |
| `ja` | 日文 | 日本語 |
| `zhtw` | 繁体中文 | 繁體中文 |
| `th` | 泰文 | ไทย |

### 特殊属性说明 (Special Attributes)

#### 图片块属性
- `font_size`: 在图片块中用于控制图片的缩放比例
- `width`: 控制图片的显示宽度（像素）

#### 二维码块属性
- `font_size`: 控制二维码的缩放比例
- `width`: 控制二维码的显示宽度（像素）

#### 条形码块属性
- `font_size`: 控制条形码的缩放比例
- `width`: 控制条形码的显示宽度（像素）

#### 多语言宽度属性
```json
"width": {
    "zh": 52.0,    // 中文宽度
    "en": 45.0,    // 英文宽度
    "ko": 48.0,    // 韩文宽度
    // ... 其他语言
}
```

### 数据绑定规则 (Data Binding Rules)

1. **字段路径**: 使用点号分隔的路径访问嵌套数据
   - 示例: `order.products.0.name` 访问第一个商品的名称
   - 示例: `store.company` 访问店铺公司名称

2. **数组索引**: 使用数字索引访问数组元素
   - 示例: `products.0` 访问第一个商品
   - 示例: `products.1.price` 访问第二个商品的价格

3. **条件字段**: 条件中的字段路径遵循相同规则
   - 示例: `order.discount_fee` 检查订单优惠金额

4. **变量替换**: 支持 `{variable}` 格式的变量替换
   - 示例: `{brand_name}` 会被替换为实际的品牌名称

## 模板结构分析

### 1. 店铺信息区域 (Store Information Section)

#### 1.1 店铺名称
```json
{
    "block_id": "store.name",              // 块唯一标识，用于数据绑定
    "block_type": "value",                 // 块类型：值类型，直接显示数据源中的值
    "block_label": {},                     // 块标签：空对象，表示不显示标签
    "block_attr": {
        "font_size": 26,                   // 字体大小：26像素，用于标题显示
        "align": "center",                 // 对齐方式：居中对齐
        "font_bold": true,                 // 字体样式：粗体显示
        "width": 100                       // 宽度：100%，占满整行
    },
    "conditions": [
        {
            "field": "store.name",         // 条件字段：检查店铺名称字段
            "operator": "not_empty",       // 操作符：不为空
            "value": 0                     // 比较值：0（无意义，仅用于not_empty操作）
        }
    ]
}
```

**字段说明：**
- `block_id`: 块的唯一标识符，用于从数据源中获取对应的值
- `block_type`: 块类型，`value` 表示直接显示数据值，不添加标签
- `block_label`: 块标签，空对象表示不显示标签文本
- `font_size`: 字体大小，26像素适合作为店铺名称的标题显示
- `align`: 对齐方式，`center` 表示居中对齐
- `font_bold`: 字体加粗，`true` 表示使用粗体显示
- `width`: 宽度百分比，100% 表示占满整行
- `conditions`: 显示条件，只有当 `store.name` 字段不为空时才显示此块

#### 1.2 店铺Logo
```json
{
    "block_id": "store.logo",              // 块唯一标识，绑定店铺Logo图片路径
    "block_type": "img",                   // 块类型：图片类型，用于显示图片
    "block_attr": {
        "font_size": 22,                   // 字体大小：22像素（图片块中此参数用于控制图片缩放）
        "align": "center",                 // 对齐方式：居中对齐
        "width": 180                       // 宽度：180像素，控制图片显示宽度
    },
    "conditions": [
        {
            "field": "store.logo",         // 条件字段：检查店铺Logo字段
            "operator": "not_empty",       // 操作符：不为空
            "value": 0                     // 比较值：0（无意义，仅用于not_empty操作）
        }
    ]
}
```

**字段说明：**
- `block_id`: 块的唯一标识符，用于从数据源中获取Logo图片路径
- `block_type`: 块类型，`img` 表示显示图片内容
- `font_size`: 在图片块中用于控制图片的缩放比例
- `align`: 对齐方式，`center` 表示图片居中显示
- `width`: 图片显示宽度，180像素适合80mm纸张的Logo显示
- `conditions`: 显示条件，只有当 `store.logo` 字段不为空时才显示此块

#### 1.3 结账单标题
```json
{
    "block_id": "title",                   // 块唯一标识，固定为"title"
    "block_type": "label",                 // 块类型：标签类型，显示固定文本
    "block_label": {
        "zh": "结账单",                    // 中文标题
        "en": "INVOICE / RECEIPT",         // 英文标题
        "ko": "점검",                      // 韩文标题
        "my": "ထွက်ခွာသည်",              // 缅甸文标题
        "tr": "Çıkış yapmak",              // 土耳其文标题
        "de": "Rechnung",                  // 德文标题
        "sv": "Slutnota",                  // 瑞典文标题
        "ja": "レシート",                  // 日文标题
        "zhtw": "结账单",                  // 繁体中文标题
        "th": "ใบเสร็จรับเงิน"             // 泰文标题
    },
    "block_attr": {
        "font_size": 32,                   // 字体大小：32像素，大标题显示
        "align": "center",                 // 对齐方式：居中对齐
        "font_bold": true,                 // 字体样式：粗体显示
        "trailing_blank_lines": 2          // 后置空行：2行，增加标题与内容的间距
    }
}
```

**字段说明：**
- `block_id`: 固定标识符 "title"，用于结账单标题
- `block_type`: 块类型，`label` 表示显示固定的标签文本
- `block_label`: 多语言标签文本，根据系统语言设置显示对应文本
- `font_size`: 字体大小，32像素适合作为主标题显示
- `align`: 对齐方式，`center` 表示标题居中显示
- `font_bold`: 字体加粗，`true` 表示使用粗体突出显示
- `trailing_blank_lines`: 后置空行数，2行用于增加标题与后续内容的间距

**支持的语言：**
- 中文 (zh): "结账单"
- 英文 (en): "INVOICE / RECEIPT"
- 韩文 (ko): "점검"
- 缅甸文 (my): "ထွက်ခွာသည်"
- 土耳其文 (tr): "Çıkış yapmak"
- 德文 (de): "Rechnung"
- 瑞典文 (sv): "Slutnota"
- 日文 (ja): "レシート"
- 繁体中文 (zhtw): "结账单"
- 泰文 (th): "ใบเสร็จรับเงิน"

#### 1.4 店铺详细信息
包含以下字段的条件显示，每个字段都使用 `label:value` 类型：

**公司名称**
```json
{
    "block_id": "store.company",           // 块唯一标识，绑定公司名称字段
    "block_type": "label:value",           // 块类型：标签+值类型，显示标签和对应的值
    "block_label": {
        "zh": "公司名称:",
        "en": "Company Name:",
        // ... 其他语言
    },
    "block_attr": {
        "font_size": 20,                   // 字体大小：20像素
        "align": "center",                 // 对齐方式：居中对齐
        "font_bold": true,                 // 字体样式：粗体显示
        "width": 100                       // 宽度：100%，占满整行
    },
    "conditions": [
        {
            "field": "store.company",      // 条件字段：检查公司名称字段
            "operator": "not_empty",       // 操作符：不为空
            "value": 0                     // 比较值：0（无意义，仅用于not_empty操作）
        }
    ]
}
```

**字段说明：**
- `block_id`: 块的唯一标识符，用于从数据源中获取对应的值
- `block_type`: 块类型，`label:value` 表示显示标签文本和对应的数据值
- `block_label`: 多语言标签文本，显示在值前面的标签
- `font_size`: 字体大小，20像素适合普通文本显示
- `align`: 对齐方式，`center` 表示居中对齐
- `font_bold`: 字体加粗，`true` 表示使用粗体显示
- `width`: 宽度百分比，100% 表示占满整行
- `conditions`: 显示条件，只有当对应字段不为空时才显示此块

**包含的字段：**
- 公司名称 (`store.company`) - 显示公司全称
- 公司地址 (`store.company_addr`) - 显示公司详细地址
- 公司电话 (`store.company_phone`) - 显示公司联系电话
- 税务登记号 (`store.company_tax_number`) - 显示税务登记号码
- 收银机序列号 (`store.cashier_sn`) - 显示收银机设备序列号
- 打印机序列号 (`store.printer_sn`) - 显示打印机设备序列号

### 2. 订单信息区域 (Order Information Section)

#### 2.1 订单基本信息
```json
{
    "block_id": "order.serial_no",         // 块唯一标识，绑定订单序列号字段
    "block_type": "label:value",           // 块类型：标签+值类型，显示标签和对应的值
    "block_label": {
        "zh": "订单号:",                   // 中文标签
        "en": "Order No:",                 // 英文标签
        "ko": "주문 번호:",                // 韩文标签
        "my": "အမှာစာ နံပါတ်:",          // 缅甸文标签
        "tr": "Sipariş No:",               // 土耳其文标签
        "de": "Bestell-Nr:",               // 德文标签
        "sv": "Beställningsnr:",           // 瑞典文标签
        "ja": "注文番号:",                 // 日文标签
        "zhtw": "訂單號:",                 // 繁体中文标签
        "th": "หมายเลขคำสั่งซื้อ:"         // 泰文标签
    },
    "block_attr": {
        "font_size": 20,                   // 字体大小：20像素
        "align": "left",                   // 对齐方式：左对齐
        "width": 100                       // 宽度：100%，占满整行
    }
}
```

**字段说明：**
- `block_id`: 块的唯一标识符，用于从数据源中获取订单序列号
- `block_type`: 块类型，`label:value` 表示显示标签文本和对应的数据值
- `block_label`: 多语言标签文本，根据系统语言设置显示对应标签
- `font_size`: 字体大小，20像素适合普通文本显示
- `align`: 对齐方式，`left` 表示左对齐显示
- `width`: 宽度百分比，100% 表示占满整行

**订单基本信息包含：**
- 订单序列号 (`order.serial_no`) - 显示桌号、人数等信息
- 订单编号 (`order.order_no`) - 显示系统生成的订单编号
- 订单备注 (`order.remark`) - 显示订单的特殊备注信息
- 收银员姓名 (`order.cashier_name`) - 显示处理订单的收银员
- 创建时间 (`order.create_time`) - 显示订单创建时间
- 更新时间 (`order.update_time`) - 显示订单最后更新时间
- 支付时间 (`order.pay_time`) - 显示订单支付完成时间

#### 2.2 商品列表标题
使用列式布局显示商品信息表头，包含三个列：

**商品名称列**
```json
{
    "block_id": "title",                   // 块唯一标识，固定为"title"
    "block_type": "label",                 // 块类型：标签类型，显示固定文本
    "block_label": {
        "zh": "商品",                      // 中文标题
        "en": "Products",                  // 英文标题
        "ko": "상품",                      // 韩文标题
        "my": "ကုန်ပစ္စည်းများ",          // 缅甸文标题
        "tr": "Ürünler",                   // 土耳其文标题
        "de": "Produkte",                  // 德文标题
        "sv": "Produkter",                 // 瑞典文标题
        "ja": "商品",                      // 日文标题
        "zhtw": "商品",                    // 繁体中文标题
        "th": "สินค้า"                     // 泰文标题
    },
    "block_attr": {
        "font_size": 18,                   // 字体大小：18像素
        "align": "left",                   // 对齐方式：左对齐
        "font_bold": true,                 // 字体样式：粗体显示
        "line_height": 40,                 // 行高：40像素
        "dividing_line": true,             // 分割线：显示分割线
        "width": {                         // 多语言宽度自适应
            "zh": 52.0,                    // 中文宽度：52%
            "my": 35.0,                    // 缅甸文宽度：35%
            "tr": 40.0,                    // 土耳其文宽度：40%
            "de": 28.0                     // 德文宽度：28%
        }
    }
}
```

**单价|数量列**
```json
{
    "block_id": "title",                   // 块唯一标识，固定为"title"
    "block_type": "label",                 // 块类型：标签类型
    "block_label": {
        "zh": "单价|数量",                 // 中文标题
        "en": "Price|Qty",                 // 英文标题
        "ko": "단가|수량",                 // 韩文标题
        "my": "တစ်ခုဈေး|အရေအတွက်",      // 缅甸文标题
        "tr": "Birim Fiyat|Miktar",        // 土耳其文标题
        "de": "Einzelpreis|Menge",         // 德文标题
        "sv": "Enhetspris|Antal",          // 瑞典文标题
        "ja": "単価|数量",                 // 日文标题
        "zhtw": "單價|數量",               // 繁体中文标题
        "th": "ราคาต่อหน่วย|จำนวน"         // 泰文标题
    },
    "block_attr": {
        "font_size": 18,                   // 字体大小：18像素
        "align": "center",                 // 对齐方式：居中对齐
        "font_bold": true,                 // 字体样式：粗体显示
        "width": {                         // 多语言宽度自适应
            "zh": 25.0,                    // 中文宽度：25%
            "my": 45.0,                    // 缅甸文宽度：45%
            "tr": 35.0,                    // 土耳其文宽度：35%
            "de": 35.0                     // 德文宽度：35%
        }
    }
}
```

**小计列**
```json
{
    "block_id": "title",                   // 块唯一标识，固定为"title"
    "block_type": "label",                 // 块类型：标签类型
    "block_label": {
        "zh": "小计",                      // 中文标题
        "en": "Subtotal",                  // 英文标题
        "ko": "소계",                      // 韩文标题
        "my": "အပေါင်း",                  // 缅甸文标题
        "tr": "Ara Toplam",                // 土耳其文标题
        "de": "Zwischensumme",             // 德文标题
        "sv": "Delsumma",                  // 瑞典文标题
        "ja": "小計",                      // 日文标题
        "zhtw": "小計",                    // 繁体中文标题
        "th": "ยอดรวมย่อย"                 // 泰文标题
    },
    "block_attr": {
        "font_size": 18,                   // 字体大小：18像素
        "align": "right",                  // 对齐方式：右对齐
        "font_bold": true,                 // 字体样式：粗体显示
        "width": 23                        // 宽度：23%
    }
}
```

**字段说明：**
- `block_id`: 固定标识符 "title"，用于表头显示
- `block_type`: 块类型，`label` 表示显示固定的标签文本
- `block_label`: 多语言标签文本，根据系统语言设置显示对应文本
- `font_size`: 字体大小，18像素适合表头显示
- `align`: 对齐方式，`left`/`center`/`right` 分别表示左对齐/居中/右对齐
- `font_bold`: 字体加粗，`true` 表示使用粗体突出显示
- `line_height`: 行高，40像素用于增加表头的视觉高度
- `dividing_line`: 分割线，`true` 表示在表头下方显示分割线
- `width`: 多语言宽度自适应，根据语言特点设置不同的列宽比例

#### 2.3 商品详情列表
```json
{
    "block_id": "order.products",          // 块唯一标识，绑定商品数组字段
    "block_type": "array",                 // 块类型：数组类型，循环显示数组中的每个元素
    "block_attr": {
        "font_size": 20                    // 字体大小：20像素
    },
    "rows": [
        [
            {
                "block_id": "name",        // 商品名称字段
                "block_type": "value",     // 块类型：值类型，直接显示数据值
                "block_attr": {
                    "font_size": 20,       // 字体大小：20像素
                    "align": "left",       // 对齐方式：左对齐
                    "width": 52            // 宽度：52%，与表头保持一致
                }
            },
            {
                "block_id": "price_num",   // 价格和数量字段
                "block_type": "value",     // 块类型：值类型
                "block_attr": {
                    "font_size": 20,       // 字体大小：20像素
                    "align": "center",     // 对齐方式：居中对齐
                    "width": 24            // 宽度：24%，与表头保持一致
                }
            },
            {
                "block_id": "subtotal",    // 小计字段
                "block_type": "value",     // 块类型：值类型
                "block_attr": {
                    "font_size": 20,       // 字体大小：20像素
                    "align": "right",      // 对齐方式：右对齐
                    "show_currency_unit": true,  // 显示货币单位：true
                    "width": 0             // 宽度：0，自动计算剩余宽度
                }
            }
        ],
        [
            {
                "block_id": "attrs",       // 商品属性字段
                "block_type": "label",     // 块类型：标签类型
                "block_label": "({attrs})", // 标签文本：显示属性值并用括号包围
                "block_attr": {
                    "font_size": 18,       // 字体大小：18像素，比商品名称稍小
                    "align": "left",       // 对齐方式：左对齐
                    "line_height": 35,     // 行高：35像素
                    "width": 52            // 宽度：52%，与商品名称列对齐
                },
                "conditions": [
                    {
                        "field": "attrs",  // 条件字段：检查商品属性字段
                        "operator": "not_empty", // 操作符：不为空
                        "value": 0         // 比较值：0（无意义，仅用于not_empty操作）
                    }
                ]
            }
        ]
    ]
}
```

**字段说明：**
- `block_id`: 块的唯一标识符，`order.products` 绑定商品数组字段
- `block_type`: 块类型，`array` 表示循环处理数组中的每个元素
- `font_size`: 字体大小，20像素适合商品信息显示
- `rows`: 嵌套行配置，定义每个商品项的显示结构

**商品信息行（第一行）：**
- `name`: 商品名称，左对齐显示，宽度52%
- `price_num`: 价格和数量，居中显示，宽度24%
- `subtotal`: 小计金额，右对齐显示，自动宽度，显示货币单位

**商品属性行（第二行）：**
- `attrs`: 商品属性，左对齐显示，宽度52%，只有属性不为空时才显示
- 显示格式：`(属性1;属性2;属性3)`
- 字体稍小（18像素），行高35像素

**商品数据结构要求：**
```json
{
    "name": "商品名称",                    // 商品名称
    "price_num": "1*1",                   // 价格*数量格式
    "price": "商品单价",                  // 商品单价
    "num": 12,                            // 数量
    "subtotal": "小计金额",               // 小计金额
    "attrs": "份;冰;香菜;洋葱",            // 商品属性，用分号分隔
    "attr": "主要属性",                   // 主要属性
    "attr_list": [                        // 属性列表
        {
            "name": "冰",
            "text": "冰"
        }
    ],
    "flavor_name": "份",                  // 口味名称
    "sauce_list": [                       // 调料列表
        {
            "name": "香菜",
            "text": "香菜"
        }
    ],
    "sauce_names": "香菜;洋葱",           // 调料名称
    "is_delay": false,                    // 是否延迟
    "is_buffet": false,                   // 是否自助餐
    "is_buffet_product": false,           // 是否自助餐商品
    "is_gift": false,                     // 是否赠品
    "is_package": false,                  // 是否套餐
    "is_sub_product": false               // 是否子商品
}
```

### 3. 金额汇总区域 (Amount Summary Section)

#### 3.1 商品统计
```json
{
    "block_id": "order.product_num",       // 块唯一标识，绑定商品总数量字段
    "block_type": "label:value",           // 块类型：标签+值类型，显示标签和对应的值
    "block_label": {
        "zh": "商品数量:",                 // 中文标签
        "en": "Total text:",               // 英文标签
        "ko": "상품 수량:",                // 韩文标签
        "my": "ကုန်ပစ္စည်းအရေအတွက်:",    // 缅甸文标签
        "tr": "Toplam Miktar:",            // 土耳其文标签
        "de": "Gesamtmenge:",              // 德文标签
        "sv": "Total kvantitet:",          // 瑞典文标签
        "ja": "商品数量:",                 // 日文标签
        "zhtw": "商品數量:",               // 繁体中文标签
        "th": "จำนวนสินค้า:"              // 泰文标签
    },
    "block_attr": {
        "font_size": 20,                   // 字体大小：20像素
        "align": "left",                   // 对齐方式：左对齐
        "width": 100                       // 宽度：100%，占满整行
    }
}
```

**字段说明：**
- `block_id`: 块的唯一标识符，用于从数据源中获取商品总数量
- `block_type`: 块类型，`label:value` 表示显示标签文本和对应的数据值
- `block_label`: 多语言标签文本，根据系统语言设置显示对应标签
- `font_size`: 字体大小，20像素适合普通文本显示
- `align`: 对齐方式，`left` 表示左对齐显示
- `width`: 宽度百分比，100% 表示占满整行

**显示内容：**
- 显示订单中所有商品的总数量
- 格式：`商品数量: 15`（其中15是商品总数量）

#### 3.2 金额明细
包含以下金额字段的条件显示：

**商品总金额**
```json
{
    "block_id": "order.product_amount",    // 块唯一标识，绑定商品总金额字段
    "block_type": "label:value",           // 块类型：标签+值类型
    "block_label": {
        "zh": "商品总金额:",               // 中文标签
        "en": "Product Amount:",           // 英文标签
        "ko": "상품 총액:",                // 韩文标签
        "my": "ကုန်ပစ္စည်းစုစုပေါင်း:",  // 缅甸文标签
        "tr": "Ürün Tutarı:",              // 土耳其文标签
        "de": "Produktbetrag:",            // 德文标签
        "sv": "Produktbelopp:",            // 瑞典文标签
        "ja": "商品総額:",                 // 日文标签
        "zhtw": "商品總金額:",             // 繁体中文标签
        "th": "ยอดรวมสินค้า:"              // 泰文标签
    },
    "block_attr": {
        "font_size": 20,                   // 字体大小：20像素
        "align": "right",                  // 对齐方式：右对齐
        "show_currency_unit": true,        // 显示货币单位：true
        "width": 100                       // 宽度：100%，占满整行
    }
}
```

**字段说明：**
- `block_id`: 块的唯一标识符，用于从数据源中获取商品总金额
- `block_type`: 块类型，`label:value` 表示显示标签文本和对应的数据值
- `block_label`: 多语言标签文本，根据系统语言设置显示对应标签
- `font_size`: 字体大小，20像素适合普通文本显示
- `align`: 对齐方式，`right` 表示右对齐显示，便于金额对比
- `show_currency_unit`: 显示货币单位，`true` 表示在金额后显示货币符号
- `width`: 宽度百分比，100% 表示占满整行

**显示内容：**
- 显示订单中所有商品的总金额
- 格式：`商品总金额: ¥1,234.56`（包含千分位格式化和货币单位）

**服务费**
```json
{
    "block_id": "order.service_fee",
    "block_type": "label:value",
    "block_label": {
        "zh": "服务费:",
        "en": "Service Fee:",
        // ... 其他语言
    },
    "conditions": [
        {
            "field": "order.service_fee",
            "operator": "not_empty",
            "value": 0
        }
    ]
}
```

**税费**
```json
{
    "block_id": "order.tax_fee",
    "block_type": "label:value",
    "block_label": {
        "zh": "税费:",
        "en": "Tax Fee:",
        // ... 其他语言
    },
    "conditions": [
        {
            "field": "order.tax_fee",
            "operator": "not_empty",
            "value": 0
        }
    ]
}
```

**优惠折扣**
```json
{
    "block_id": "order.discount_fee",
    "block_type": "label:value",
    "block_label": {
        "zh": "优惠折扣:",
        "en": "Discount:",
        // ... 其他语言
    },
    "block_after_label": {
        "zh": "({order.discount_rate}% OFF)",
        "en": "({order.discount_rate}% OFF)",
        // ... 其他语言
    },
    "conditions": [
        {
            "field": "order.discount_fee",
            "operator": "not_empty",
            "value": 0
        }
    ]
}
```

**会员优惠**
```json
{
    "block_id": "order.member_discount_fee",
    "block_type": "label:value",
    "block_label": {
        "zh": "会员优惠:",
        "en": "Member Discount:",
        // ... 其他语言
    },
    "conditions": [
        {
            "field": "order.member_points_discount",
            "operator": "not_empty",
            "value": 0
        }
    ]
}
```

**会员折扣率**
```json
{
    "block_id": "order.member_discount_rate",
    "block_type": "label:value",
    "block_label": {
        "zh": "会员折扣:",
        "en": "Member Rate:",
        // ... 其他语言
    },
    "block_after_label": {
        "zh": "({order.member_discount_rate}% OFF)",
        "en": "({order.member_discount_rate}% OFF)",
        // ... 其他语言
    },
    "conditions": [
        {
            "field": "order.member_discount_rate",
            "operator": "not_empty",
            "value": 0
        }
    ]
}
```

**优惠券兑换金额**
```json
{
    "block_id": "order.coupon_exchange_amount",
    "block_type": "label:value",
    "block_label": {
        "zh": "优惠券兑换:",
        "en": "Coupon Exchange:",
        // ... 其他语言
    },
    "conditions": [
        {
            "field": "order.coupon_exchange_amount",
            "operator": "not_empty",
            "value": 0
        }
    ]
}
```

**退款金额**
```json
{
    "block_id": "order.return_amount",
    "block_type": "label:value",
    "block_label": {
        "zh": "退款金额:",
        "en": "Return Amount:",
        // ... 其他语言
    },
    "conditions": [
        {
            "field": "order.return_amount",
            "operator": "not_empty",
            "value": 0
        }
    ]
}
```

**支付手续费**
```json
{
    "block_id": "order.payment_commission_fee",
    "block_type": "label:value",
    "block_label": {
        "zh": "支付手续费:",
        "en": "Payment Commission:",
        // ... 其他语言
    },
    "conditions": [
        {
            "field": "order.payment_commission_fee",
            "operator": "not_empty",
            "value": 0
        }
    ]
}
```

**免费金额**
```json
{
    "block_id": "order.free_amount",
    "block_type": "label:value",
    "block_label": {
        "zh": "免费金额:",
        "en": "Free Amount:",
        // ... 其他语言
    },
    "conditions": [
        {
            "field": "order.free_amount",
            "operator": "not_empty",
            "value": 0
        }
    ]
}
```

**实际收款金额**
```json
{
    "block_id": "order.actual_receive_price", // 块唯一标识，绑定实际收款金额字段
    "block_type": "label:value",               // 块类型：标签+值类型
    "block_label": {
        "zh": "实际收款:",                     // 中文标签
        "en": "Actual Receive:",               // 英文标签
        "ko": "실제 수령:",                    // 韩文标签
        "my": "အမှန်တကယ်လက်ခံ:",            // 缅甸文标签
        "tr": "Gerçek Alınan:",                // 土耳其文标签
        "de": "Tatsächlich erhalten:",         // 德文标签
        "sv": "Faktiskt mottaget:",            // 瑞典文标签
        "ja": "実際の受取:",                   // 日文标签
        "zhtw": "實際收款:",                   // 繁体中文标签
        "th": "เงินที่ได้รับจริง:"              // 泰文标签
    },
    "block_attr": {
        "font_size": 22,                       // 字体大小：22像素，比普通文本稍大
        "align": "right",                      // 对齐方式：右对齐
        "font_bold": true,                     // 字体样式：粗体显示，突出重要性
        "show_currency_unit": true,            // 显示货币单位：true
        "width": 100                           // 宽度：100%，占满整行
    }
}
```

**字段说明：**
- `block_id`: 块的唯一标识符，用于从数据源中获取实际收款金额
- `block_type`: 块类型，`label:value` 表示显示标签文本和对应的数据值
- `block_label`: 多语言标签文本，根据系统语言设置显示对应标签
- `font_size`: 字体大小，22像素比普通文本稍大，突出重要性
- `align`: 对齐方式，`right` 表示右对齐显示，便于金额对比
- `font_bold`: 字体加粗，`true` 表示使用粗体显示，突出实际收款金额的重要性
- `show_currency_unit`: 显示货币单位，`true` 表示在金额后显示货币符号
- `width`: 宽度百分比，100% 表示占满整行

**显示内容：**
- 显示订单的实际收款金额（扣除所有优惠、折扣后的最终金额）
- 格式：`实际收款: ¥1,123.45`（包含千分位格式化和货币单位）
- 使用粗体显示，突出这是最重要的金额信息

### 4. 支付信息区域 (Payment Information Section)

#### 4.1 支付方式
```json
{
    "block_id": "order.payment_name",      // 块唯一标识，绑定支付方式字段
    "block_type": "label:value",           // 块类型：标签+值类型
    "block_label": {
        "zh": "支付方式:",                 // 中文标签
        "en": "Payment Method:",           // 英文标签
        "ko": "결제 방법:",                // 韩文标签
        "my": "ငွေပေးချေမှုနည်းလမ်း:",    // 缅甸文标签
        "tr": "Ödeme Yöntemi:",            // 土耳其文标签
        "de": "Zahlungsmethode:",          // 德文标签
        "sv": "Betalningsmetod:",          // 瑞典文标签
        "ja": "支払い方法:",               // 日文标签
        "zhtw": "支付方式:",               // 繁体中文标签
        "th": "วิธีการชำระเงิน:"            // 泰文标签
    },
    "block_attr": {
        "font_size": 20,                   // 字体大小：20像素
        "align": "left",                   // 对齐方式：左对齐
        "width": 100                       // 宽度：100%，占满整行
    }
}
```

**字段说明：**
- `block_id`: 块的唯一标识符，用于从数据源中获取支付方式
- `block_type`: 块类型，`label:value` 表示显示标签文本和对应的数据值
- `block_label`: 多语言标签文本，根据系统语言设置显示对应标签
- `font_size`: 字体大小，20像素适合普通文本显示
- `align`: 对齐方式，`left` 表示左对齐显示
- `width`: 宽度百分比，100% 表示占满整行

**显示内容：**
- 显示订单的支付方式
- 格式：`支付方式: 微信支付`（显示具体的支付方式名称）

#### 4.2 支付二维码
```json
{
    "block_id": "order.payment_qrcode",    // 块唯一标识，绑定支付二维码字段
    "block_type": "qrcode",                // 块类型：二维码类型，用于显示二维码
    "block_attr": {
        "font_size": 22,                   // 字体大小：22像素（二维码块中用于控制二维码大小）
        "align": "center",                 // 对齐方式：居中对齐
        "width": 500                       // 宽度：500像素，控制二维码显示宽度
    },
    "conditions": [
        {
            "field": "order.payment_qrcode", // 条件字段：检查支付二维码字段
            "operator": "not_empty",         // 操作符：不为空
            "value": 0                       // 比较值：0（无意义，仅用于not_empty操作）
        }
    ]
}
```

**字段说明：**
- `block_id`: 块的唯一标识符，用于从数据源中获取支付二维码内容
- `block_type`: 块类型，`qrcode` 表示显示二维码
- `font_size`: 在二维码块中用于控制二维码的缩放比例
- `align`: 对齐方式，`center` 表示二维码居中显示
- `width`: 二维码显示宽度，500像素适合80mm纸张的二维码显示
- `conditions`: 显示条件，只有当 `order.payment_qrcode` 字段不为空时才显示此块

**显示内容：**
- 显示支付相关的二维码（如微信支付二维码、支付宝二维码等）
- 二维码内容来自 `order.payment_qrcode` 字段

#### 4.3 订单条形码
```json
{
    "block_id": "order.barcode",           // 块唯一标识，绑定订单条形码字段
    "block_type": "barcode",               // 块类型：条形码类型，用于显示条形码
    "block_attr": {
        "font_size": 22,                   // 字体大小：22像素（条形码块中用于控制条形码大小）
        "align": "center",                 // 对齐方式：居中对齐
        "width": 500                       // 宽度：500像素，控制条形码显示宽度
    },
    "conditions": [
        {
            "field": "order.barcode",      // 条件字段：检查订单条形码字段
            "operator": "not_empty",       // 操作符：不为空
            "value": 0                     // 比较值：0（无意义，仅用于not_empty操作）
        }
    ]
}
```

**字段说明：**
- `block_id`: 块的唯一标识符，用于从数据源中获取订单条形码内容
- `block_type`: 块类型，`barcode` 表示显示条形码
- `font_size`: 在条形码块中用于控制条形码的缩放比例
- `align`: 对齐方式，`center` 表示条形码居中显示
- `width`: 条形码显示宽度，500像素适合80mm纸张的条形码显示
- `conditions`: 显示条件，只有当 `order.barcode` 字段不为空时才显示此块

**显示内容：**
- 显示订单的条形码（通常为订单编号的条形码形式）
- 条形码内容来自 `order.barcode` 字段
- 可用于订单查询、验证等用途

### 5. 页脚信息区域 (Footer Section)

#### 5.1 感谢信息
```json
{
    "block_id": "title",                   // 块唯一标识，固定为"title"
    "block_type": "label",                 // 块类型：标签类型，显示固定文本
    "block_label": {
        "zh": "感谢您的光临！本店由 {brand_name} 系统提供支持。",  // 中文感谢信息
        "en": "Thank you for your visit! This store is powered by  {brand_name}  system.",  // 英文感谢信息
        "ko": "방문해 주셔서 감사합니다! 이 매장은  {brand_name}  시스템으로 지원됩니다.",  // 韩文感谢信息
        "my": "လာရောက်ခြင်းအတွက် ကျေးဇူးတင်ပါတယ်! ဤဆိုင်ကို  {brand_name}  စနစ်ဖြင့် ပံ့ပိုးထားပါတယ်။",  // 缅甸文感谢信息
        "tr": "Ziyaretiniz için teşekkür ederiz! Bu mağaza tarafından:  {brand_name}  Sistem destek sağlar.",  // 土耳其文感谢信息
        "de": "Vielen Dank für Ihren Besuch! Dieser Laden wird vom  {brand_name}  System unterstützt.",  // 德文感谢信息
        "sv": "Tack för ditt besök! Denna butik drivs av  {brand_name}  systemet.",  // 瑞典文感谢信息
        "ja": "ご来店ありがとうございます！この店舗は {brand_name} システムでサポートされています。",  // 日文感谢信息
        "zhtw": "感謝您的光臨！本店由 {brand_name} 系統提供支持。",  // 繁体中文感谢信息
        "th": "ขอบคุณที่แวะมาหากัน!สนับสนุนโดย  {brand_name} "  // 泰文感谢信息
    },
    "block_attr": {
        "font_size": 20,                   // 字体大小：20像素
        "align": "center",                 // 对齐方式：居中对齐
        "font_bold": true,                 // 字体样式：粗体显示
        "trailing_blank_lines": 6,         // 后置空行：6行，增加页脚与切纸位置的间距
        "width": 100                       // 宽度：100%，占满整行
    }
}
```

**字段说明：**
- `block_id`: 固定标识符 "title"，用于页脚感谢信息
- `block_type`: 块类型，`label` 表示显示固定的标签文本
- `block_label`: 多语言感谢信息文本，根据系统语言设置显示对应文本
- `font_size`: 字体大小，20像素适合页脚文本显示
- `align`: 对齐方式，`center` 表示感谢信息居中显示
- `font_bold`: 字体加粗，`true` 表示使用粗体显示，突出感谢信息
- `trailing_blank_lines`: 后置空行数，6行用于增加页脚与切纸位置的间距
- `width`: 宽度百分比，100% 表示占满整行

**显示内容：**
- 显示感谢顾客光临的信息
- 包含品牌名称变量 `{brand_name}`，会替换为实际的品牌名称
- 格式：`感谢您的光临！本店由 TTPOS 系统提供支持。`
- 使用粗体显示，突出感谢信息的重要性
- 后置6行空行，为切纸操作预留空间

**支持的语言：**
- 中文 (zh): "感谢您的光临！本店由 {brand_name} 系统提供支持。"
- 英文 (en): "Thank you for your visit! This store is powered by {brand_name} system."
- 韩文 (ko): "방문해 주셔서 감사합니다! 이 매장은 {brand_name} 시스템으로 지원됩니다."
- 缅甸文 (my): "လာရောက်ခြင်းအတွက် ကျေးဇူးတင်ပါတယ်! ဤဆိုင်ကို {brand_name} စနစ်ဖြင့် ပံ့ပိုးထားပါတယ်။"
- 土耳其文 (tr): "Ziyaretiniz için teşekkür ederiz! Bu mağaza tarafından: {brand_name} Sistem destek sağlar."
- 德文 (de): "Vielen Dank für Ihren Besuch! Dieser Laden wird vom {brand_name} System unterstützt."
- 瑞典文 (sv): "Tack för ditt besök! Denna butik drivs av {brand_name} systemet."
- 日文 (ja): "ご来店ありがとうございます！この店舗は {brand_name} システムでサポートされています。"
- 繁体中文 (zhtw): "感謝您的光臨！本店由 {brand_name} 系統提供支持。"
- 泰文 (th): "ขอบคุณที่แวะมาหากัน!สนับสนุนโดย {brand_name}"

## 数据结构要求

### 店铺信息 (store)
```json
{
    "store": {
        "name": "店铺名称",
        "address": "店铺地址",
        "phone": "店铺电话",
        "logo": "店铺Logo图片路径",
        "company": "公司名称",
        "company_addr": "公司地址",
        "company_phone": "公司电话",
        "company_tax_number": "税务登记号",
        "cashier_sn": "收银机序列号",
        "printer_sn": "打印机序列号"
    }
}
```

### 订单信息 (order)
```json
{
    "order": {
        "serial_no": "桌号: A5 (1人)",
        "order_no": "202509252185102915",
        "remark": "订单备注",
        "cashier_name": "收银员姓名",
        "create_time": "2025/09/25 18:28:56",
        "update_time": "2025/09/25 18:28:56",
        "pay_time": "支付时间",
        "products": [
            {
                "name": "商品名称",
                "price_num": "1*1",           // 价格*数量格式
                "price": "商品单价",
                "num": 12,                    // 数量
                "subtotal": "小计金额",
                "remark": "商品备注",
                "attrs": "份;冰;香菜;洋葱",    // 商品属性
                "attr": "主要属性",
                "attr_list": [                // 属性列表
                    {
                        "name": "冰",
                        "text": "冰"
                    }
                ],
                "flavor_name": "份",          // 口味名称
                "sauce_list": [               // 调料列表
                    {
                        "name": "香菜",
                        "text": "香菜"
                    }
                ],
                "sauce_names": "香菜;洋葱",   // 调料名称
                "is_delay": false,            // 是否延迟
                "is_buffet": false,           // 是否自助餐
                "is_buffet_product": false,   // 是否自助餐商品
                "is_gift": false,             // 是否赠品
                "is_package": false,          // 是否套餐
                "is_sub_product": false       // 是否子商品
            }
        ],
        "product_num": 0,                     // 商品总数量
        "product_amount": "29,297.01",        // 商品总金额
        "service_fee": "0",                   // 服务费
        "tax_rate": 0,                        // 税率
        "tax_fee_type": 2,                    // 税费类型
        "is_contain_tax": 2,                  // 是否含税
        "discount_fee": "0.01",               // 优惠金额
        "discount_rate": "0",                 // 优惠折扣率
        "member_discount_fee": "0",           // 会员优惠金额
        "member_discount_rate": 0,            // 会员折扣率
        "member_card_discount_rate": 0,       // 会员卡折扣率
        "member_points_discount": 0,          // 会员积分折扣
        "coupon_exchange_amount": 0,          // 优惠券兑换金额
        "check_out_zero_fee": "0",            // 零元结账费用
        "return_amount": "0",                 // 退款金额
        "payment_commission_fee": 0,          // 支付手续费
        "free_amount": "0",                   // 免费金额
        "actual_receive_price": "29,297",     // 实际收款金额
        "payment_methods": [],                // 支付方式列表
        "percentage_lists": [],               // 百分比列表
        "is_free": false,                     // 是否免费
        "is_member": false,                   // 是否会员
        "member_remaining_balance": "0",      // 会员剩余余额
        "member_points": 0,                   // 会员积分
        "payment_name": "",                   // 支付方式名称
        "payment_qrcode": "",                 // 支付二维码
        "barcode": "202509252185102915"       // 订单条形码
    }
}
```

### 品牌信息 (brand_name)
```json
{
    "brand_name": "TTPOS"                    // 品牌名称
}
```

## 模板特性

### 1. 多语言支持
- 支持10种语言：中文、英文、韩文、缅甸文、土耳其文、德文、瑞典文、日文、繁体中文、泰文
- 每种语言都有对应的标签文本
- 支持语言特定的宽度自适应

### 2. 条件显示
- 所有字段都支持条件显示
- 支持 `not_empty`、`empty`、`>`、`<`、`==`、`!=` 等操作符
- 只有满足条件的字段才会显示

### 3. 金额格式化
- 启用千分位格式化 (`thousandth: true`)
- 自动显示货币单位 (`show_currency_unit: true`)
- 支持多种金额类型：商品金额、服务费、税费、优惠、退款等

### 4. 布局特性
- 使用列式布局显示商品信息
- 支持多行显示（商品名称、属性分别显示）
- 自适应宽度，根据语言调整列宽
- 支持图片、二维码、条形码显示

### 5. 样式控制
- 支持字体大小控制 (12-32)
- 支持对齐方式：左对齐、居中、右对齐
- 支持粗体、下划线、反色等样式
- 支持行高控制
- 支持前导和后置空行

## 使用示例

### 基本使用
```go
// 读取模板文件
templateJSON, err := os.ReadFile("statement_order_tmp.json")
if err != nil {
    log.Fatal(err)
}

// 准备数据
data := map[string]interface{}{
    "brand_name": "TTPOS",
    "store": map[string]interface{}{
        "name": "重庆高老九火锅曼谷一号店",
        "address": "1167,17-20 Ratchadaphisek Rd,Din Daeng,Din Daeng,  Bangkok 10400",
        "phone": "025508999212",
    },
    "order": map[string]interface{}{
        "serial_no": "桌号: A5 (1人)",
        "order_no": "202509252185102915",
        "create_time": "2025/09/25 18:28:56",
        "products": []map[string]interface{}{
            {
                "name": "十三香小龙虾（1kg）",
                "price_num": "1*1",
                "price": "1",
                "num": 12,
                "subtotal": "1",
                "attrs": "份;冰;香菜;洋葱",
            },
        },
        "product_amount": "29,297.01",
        "actual_receive_price": "29,297",
        "barcode": "202509252185102915",
    },
}

// 创建解析器
parser, err := pkg.NewImgTemplateParser(pkg.ImgBaseData{
    Language:             "zh",
    CurrencyUnit:         "¥",
    CurrencyUnitPosition: 1,
}, string(templateJSON), data)
if err != nil {
    log.Fatal(err)
}

// 解析模板
img, err := parser.Parse()
if err != nil {
    log.Fatal(err)
}

// 保存图片
img.Save("statement_order.png", false, 0)
```

## 注意事项

1. **数据完整性**: 确保所有必需的数据字段都已提供
2. **语言设置**: 根据目标用户设置正确的语言代码
3. **图片路径**: Logo图片路径必须是有效的文件路径
4. **金额格式**: 金额字段建议使用字符串格式，避免精度问题
5. **条件显示**: 只有满足条件的字段才会显示，确保数据正确性
6. **多语言文本**: 确保所有多语言标签都有对应的翻译
7. **模板版本**: 注意模板版本兼容性
8. **纸张宽度**: 根据实际打印机设置正确的纸张宽度

## 扩展说明

### 添加新语言
1. 在 `block_label` 中添加新的语言代码和对应文本
2. 在 `width` 配置中添加该语言的宽度设置
3. 确保数据源中包含该语言的相关字段

### 添加新字段
1. 在模板中添加新的 `block` 配置
2. 在数据结构中添加对应的字段
3. 设置适当的显示条件和样式

### 自定义样式
1. 修改 `block_attr` 中的样式属性
2. 调整字体大小、对齐方式、颜色等
3. 根据需要添加或修改条件显示逻辑

这个模板为餐饮收银系统提供了完整的结账单打印功能，支持多语言、多货币、条件显示等高级特性，能够满足各种复杂的业务需求。
