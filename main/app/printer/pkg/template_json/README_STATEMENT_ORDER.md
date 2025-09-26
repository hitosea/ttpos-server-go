# 结账单模板 (Statement Order Template) 使用指南

## 📋 模板简介

`statement_order_tmp.json` 是餐饮收银系统的结账单打印模板，支持多语言显示、条件显示、金额格式化等功能，适用于各种餐饮场景。

### ✨ 主要特性
- 🌍 **多语言支持**: 支持中文、英文、韩文等10种语言
- 💰 **智能金额格式化**: 自动千分位分隔、货币单位显示
- 📱 **条件显示**: 根据数据内容智能显示/隐藏字段
- 🖼️ **多媒体支持**: 支持图片、二维码、条形码显示
- 📏 **自适应布局**: 根据纸张宽度和语言自动调整布局

## ⚙️ 基本配置

### 模板元数据配置
```json
{
    "metadata": {
        "name": "statement_order",           // 模板名称（必填）
        "description": "结账单模板",          // 模板描述（可选）
        "paper_width": 80,                  // 纸张宽度：58/80/0（必填）
        "version": "1.0",                   // 模板版本（可选）
        "thousandth": true                  // 千分位格式化（可选）
    }
}
```

### 配置参数说明

| 参数 | 类型 | 必填 | 说明 | 可选值 | 示例 |
|------|------|------|------|--------|------|
| `name` | string | ✅ | 模板唯一标识 | 任意字符串 | `"statement_order"` |
| `description` | string | ❌ | 模板描述 | 任意字符串 | `"餐饮收银小票模板"` |
| `paper_width` | int | ✅ | 纸张宽度（毫米） | 58, 80, 0 | `80` |
| `version` | string | ❌ | 模板版本号 | 语义化版本 | `"1.0"` |
| `thousandth` | boolean | ❌ | 千分位格式化 | true, false | `true` |

### 纸张宽度选择指南

| 宽度 | 适用场景 | 推荐用途 |
|------|----------|----------|
| `58` | 便携式小票打印机 | 外卖小票、便携设备 |
| `80` | 标准热敏打印机 | 餐厅收银、正式收据（推荐） |
| `0` | 动态配置 | 多设备环境、灵活配置 |

### 配置示例

**标准餐厅配置**
```json
{
    "metadata": {
        "name": "restaurant_receipt",
        "description": "标准餐厅收银小票",
        "paper_width": 80,
        "version": "1.0",
        "thousandth": true
    }
}
```

**便携设备配置**
```json
{
    "metadata": {
        "name": "portable_receipt", 
        "description": "便携式小票",
        "paper_width": 58,
        "version": "1.0",
        "thousandth": false
    }
}
```

## 🧩 模板结构说明

### 完整模板示例
```json
{
    "metadata": {
        "name": "statement_order",
        "description": "结账单模板",
        "paper_width": 80,
        "version": "1.0",
        "thousandth": true
    },
    "rows": [
        [
            {
                "block_id": "store.name",
                "block_type": "value",
                "block_attr": {
                    "font_size": 26,
                    "align": "center",
                    "font_bold": true,
                    "width": 100
                },
                "conditions": [
                    {
                        "field": "store.name",
                        "operator": "not_empty",
                        "value": 0
                    }
                ]
            }
        ],
        [
            {
                "block_id": "store.logo",
                "block_type": "img",
                "block_attr": {
                    "font_size": 22,
                    "align": "center",
                    "width": 180
                },
                "conditions": [
                    {
                        "field": "store.logo",
                        "operator": "not_empty",
                        "value": 0
                    }
                ]
            }
        ],
        [
            {
                "block_id": "title",
                "block_type": "label",
                "block_label": {
                    "zh": "结账单",
                    "en": "INVOICE / RECEIPT"
                },
                "block_attr": {
                    "font_size": 32,
                    "align": "center",
                    "font_bold": true,
                    "trailing_blank_lines": 2
                }
            }
        ],
        [
            {
                "block_id": "order.serial_no",
                "block_type": "label:value",
                "block_label": {
                    "zh": "订单号:",
                    "en": "Order No:"
                },
                "block_attr": {
                    "font_size": 20,
                    "align": "left",
                    "width": 100
                }
            }
        ],
        [
            {
                "block_id": "order.products",
                "block_type": "array",
                "block_attr": {
                    "font_size": 20
                },
                "rows": [
                    [
                        {
                            "block_id": "name",
                            "block_type": "value",
                            "block_attr": {
                                "font_size": 20,
                                "align": "left",
                                "width": 52
                            }
                        },
                        {
                            "block_id": "price_num",
                            "block_type": "value",
                            "block_attr": {
                                "font_size": 20,
                                "align": "center",
                                "width": 24
                            }
                        },
                        {
                            "block_id": "subtotal",
                            "block_type": "value",
                            "block_attr": {
                                "font_size": 20,
                                "align": "right",
                                "show_currency_unit": true,
                                "width": 0
                            }
                        }
                    ]
                ]
            }
        ],
        [
            {
                "block_id": "order.actual_receive_price",
                "block_type": "label:value",
                "block_label": {
                    "zh": "实际收款:",
                    "en": "Actual Receive:"
                },
                "block_attr": {
                    "font_size": 22,
                    "align": "right",
                    "font_bold": true,
                    "show_currency_unit": true,
                    "width": 100
                }
            }
        ],
        [
            {
                "block_id": "order.payment_name",
                "block_type": "label:value",
                "block_label": {
                    "zh": "支付方式:",
                    "en": "Payment Method:"
                },
                "block_attr": {
                    "font_size": 20,
                    "align": "left",
                    "width": 100
                }
            }
        ],
        [
            {
                "block_id": "order.barcode",
                "block_type": "barcode",
                "block_attr": {
                    "font_size": 22,
                    "align": "center",
                    "width": 500
                },
                "conditions": [
                    {
                        "field": "order.barcode",
                        "operator": "not_empty",
                        "value": 0
                    }
                ]
            }
        ],
        [
            {
                "block_id": "title",
                "block_type": "label",
                "block_label": {
                    "zh": "感谢您的光临！本店由 {brand_name} 系统提供支持。",
                    "en": "Thank you for your visit! This store is powered by {brand_name} system."
                },
                "block_attr": {
                    "font_size": 20,
                    "align": "center",
                    "font_bold": true,
                    "trailing_blank_lines": 6,
                    "width": 100
                }
            }
        ]
    ]
}
```

### 块类型 (Block Types)

| 类型 | 说明 | 用途 | 示例 |
|------|------|------|------|
| `value` | 值类型 | 直接显示数据值 | 商品名称、金额 |
| `label` | 标签类型 | 显示固定文本 | 标题、说明文字 |
| `label:value` | 标签值类型 | 显示标签+数据值 | "商品金额: ¥100" |
| `label:auto:value` | 自动标签值类型 | 自动显示标签+值 | 智能标签显示 |
| `array` | 数组类型 | 循环显示数组元素 | 商品列表 |
| `img` | 图片类型 | 显示图片 | Logo、商品图片 |
| `qrcode` | 二维码类型 | 显示二维码 | 支付二维码 |
| `barcode` | 条形码类型 | 显示条形码 | 订单条形码 |
| `blank_line` | 空行类型 | 添加空行 | 分隔内容 |
| `column` | 列类型 | 列式布局 | 表格显示 |

### 块属性 (Block Attributes)

| 属性 | 类型 | 说明 | 可选值 | 示例 |
|------|------|------|--------|------|
| `font_size` | int | 字体大小（像素） | 12-32 | `20` |
| `align` | string | 对齐方式 | left, center, right | `"center"` |
| `font_bold` | boolean | 是否粗体 | true, false | `true` |
| `width` | int/object | 宽度百分比 | 0-100 或多语言对象 | `100` 或 `{"zh": 52.0}` |
| `line_height` | int | 行高（像素） | 20-50 | `40` |
| `show_currency_unit` | boolean | 显示货币单位 | true, false | `true` |
| `dividing_line` | boolean | 显示分割线 | true, false | `true` |
| `leading_blank_lines` | int | 前导空行数 | 0-10 | `2` |
| `trailing_blank_lines` | int | 后置空行数 | 0-10 | `1` |

### 条件操作符 (Condition Operators)

| 操作符 | 说明 | 适用类型 | 示例 |
|--------|------|----------|------|
| `eq` | 等于 | 所有类型 | `"field": "status", "operator": "eq", "value": "active"` |
| `ne` | 不等于 | 所有类型 | `"field": "status", "operator": "ne", "value": "inactive"` |
| `gt` | 大于 | 数字类型 | `"field": "amount", "operator": "gt", "value": 0` |
| `gte` | 大于等于 | 数字类型 | `"field": "amount", "operator": "gte", "value": 100` |
| `lt` | 小于 | 数字类型 | `"field": "amount", "operator": "lt", "value": 1000` |
| `lte` | 小于等于 | 数字类型 | `"field": "amount", "operator": "lte", "value": 500` |
| `contains` | 包含 | 字符串类型 | `"field": "name", "operator": "contains", "value": "test"` |
| `not_contains` | 不包含 | 字符串类型 | `"field": "name", "operator": "not_contains", "value": "test"` |
| `in` | 在列表中 | 所有类型 | `"field": "status", "operator": "in", "value": ["active", "pending"]` |
| `not_in` | 不在列表中 | 所有类型 | `"field": "status", "operator": "not_in", "value": ["inactive", "cancelled"]` |
| `empty` | 为空 | 所有类型 | `"field": "description", "operator": "empty", "value": 0` |
| `not_empty` | 不为空 | 所有类型 | `"field": "description", "operator": "not_empty", "value": 0` |

### 多语言支持

| 语言代码 | 语言名称 | 本地名称 |
|----------|----------|----------|
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

### 数据绑定规则

1. **字段路径**: 使用点号分隔访问嵌套数据
   - `order.products.0.name` → 第一个商品名称
   - `store.company` → 店铺公司名称

2. **数组索引**: 使用数字索引访问数组元素
   - `products.0` → 第一个商品
   - `products.1.price` → 第二个商品价格

3. **变量替换**: 使用 `{variable}` 格式
   - `{brand_name}` → 品牌名称
   - `{order.serial_no}` → 订单号

## 📄 模板结构详解

### 1. 店铺信息区域

#### 店铺名称
```json
{
    "block_id": "store.name",
    "block_type": "value",
    "block_attr": {
        "font_size": 26,
        "align": "center",
        "font_bold": true,
        "width": 100
    },
    "conditions": [
        {
            "field": "store.name",
            "operator": "not_empty",
            "value": 0
        }
    ]
}
```
**说明**: 显示店铺名称，居中粗体显示，只有店铺名称不为空时才显示。

#### 店铺Logo
```json
{
    "block_id": "store.logo",
    "block_type": "img",
    "block_attr": {
        "font_size": 22,
        "align": "center",
        "width": 180
    },
    "conditions": [
        {
            "field": "store.logo",
            "operator": "not_empty",
            "value": 0
        }
    ]
}
```
**说明**: 显示店铺Logo图片，居中显示，宽度180像素，只有Logo路径不为空时才显示。

#### 结账单标题
```json
{
    "block_id": "title",
    "block_type": "label",
    "block_label": {
        "zh": "结账单",
        "en": "INVOICE / RECEIPT",
        "ko": "점검",
        "my": "ထွက်ခွာသည်",
        "tr": "Çıkış yapmak",
        "de": "Rechnung",
        "sv": "Slutnota",
        "ja": "レシート",
        "zhtw": "结账单",
        "th": "ใบเสร็จรับเงิน"
    },
    "block_attr": {
        "font_size": 32,
        "align": "center",
        "font_bold": true,
        "trailing_blank_lines": 2
    }
}
```
**说明**: 多语言标题显示，根据系统语言自动切换，32像素粗体居中显示。

#### 店铺详细信息
包含以下字段的条件显示（每个字段都使用 `label:value` 类型）：

| 字段 | 说明 | 显示条件 |
|------|------|----------|
| `store.company` | 公司名称 | 不为空时显示 |
| `store.company_addr` | 公司地址 | 不为空时显示 |
| `store.company_phone` | 公司电话 | 不为空时显示 |
| `store.company_tax_number` | 税务登记号 | 不为空时显示 |
| `store.cashier_sn` | 收银机序列号 | 不为空时显示 |
| `store.printer_sn` | 打印机序列号 | 不为空时显示 |

**示例配置**:
```json
{
    "block_id": "store.company",
    "block_type": "label:value",
    "block_label": {
        "zh": "公司名称:",
        "en": "Company Name:",
        "ko": "회사명:",
        "my": "ကုမ္ပဏီအမည်:",
        "tr": "Şirket Adı:",
        "de": "Firmenname:",
        "sv": "Företagsnamn:",
        "ja": "会社名:",
        "zhtw": "公司名稱:",
        "th": "ชื่อบริษัท:"
    },
    "block_attr": {
        "font_size": 20,
        "align": "center",
        "font_bold": true,
        "width": 100
    },
    "conditions": [
        {
            "field": "store.company",
            "operator": "not_empty",
            "value": 0
        }
    ]
}
```

### 2. 订单信息区域

#### 订单基本信息
包含以下订单字段的条件显示：

| 字段 | 说明 | 显示条件 |
|------|------|----------|
| `order.serial_no` | 订单序列号（桌号、人数等） | 不为空时显示 |
| `order.order_no` | 订单编号 | 不为空时显示 |
| `order.remark` | 订单备注 | 不为空时显示 |
| `order.cashier_name` | 收银员姓名 | 不为空时显示 |
| `order.create_time` | 创建时间 | 不为空时显示 |
| `order.update_time` | 更新时间 | 不为空时显示 |
| `order.pay_time` | 支付时间 | 不为空时显示 |

**示例配置**:
```json
{
    "block_id": "order.serial_no",
    "block_type": "label:value",
    "block_label": {
        "zh": "订单号:",
        "en": "Order No:",
        "ko": "주문 번호:",
        "my": "အမှာစာ နံပါတ်:",
        "tr": "Sipariş No:",
        "de": "Bestell-Nr:",
        "sv": "Beställningsnr:",
        "ja": "注文番号:",
        "zhtw": "訂單號:",
        "th": "หมายเลขคำสั่งซื้อ:"
    },
    "block_attr": {
        "font_size": 20,
        "align": "left",
        "width": 100
    }
}
```

#### 商品列表标题
使用列式布局显示商品信息表头，包含三个列：

| 列名 | 对齐方式 | 宽度 | 说明 |
|------|----------|------|------|
| 商品名称 | 左对齐 | 52% | 显示商品名称 |
| 单价\|数量 | 居中 | 25% | 显示价格和数量 |
| 小计 | 右对齐 | 23% | 显示小计金额 |

**示例配置**:
```json
{
    "block_id": "title",
    "block_type": "label",
    "block_label": {
        "zh": "商品",
        "en": "Products",
        "ko": "상품",
        "my": "ကုန်ပစ္စည်းများ",
        "tr": "Ürünler",
        "de": "Produkte",
        "sv": "Produkter",
        "ja": "商品",
        "zhtw": "商品",
        "th": "สินค้า"
    },
    "block_attr": {
        "font_size": 18,
        "align": "left",
        "font_bold": true,
        "line_height": 40,
        "dividing_line": true,
        "width": {
            "zh": 52.0,
            "my": 35.0,
            "tr": 40.0,
            "de": 28.0
        }
    }
}
```

#### 商品详情列表
使用数组类型循环显示商品信息，包含商品名称、价格数量、小计和属性。

**商品信息行（第一行）**:
- 商品名称：左对齐，宽度52%
- 价格|数量：居中，宽度24%
- 小计：右对齐，自动宽度，显示货币单位

**商品属性行（第二行）**:
- 商品属性：左对齐，宽度52%，只有属性不为空时才显示
- 显示格式：`(属性1;属性2;属性3)`

**示例配置**:
```json
{
    "block_id": "order.products",
    "block_type": "array",
    "block_attr": {
        "font_size": 20
    },
    "rows": [
        [
            {
                "block_id": "name",
                "block_type": "value",
                "block_attr": {
                    "font_size": 20,
                    "align": "left",
                    "width": 52
                }
            },
            {
                "block_id": "price_num",
                "block_type": "value",
                "block_attr": {
                    "font_size": 20,
                    "align": "center",
                    "width": 24
                }
            },
            {
                "block_id": "subtotal",
                "block_type": "value",
                "block_attr": {
                    "font_size": 20,
                    "align": "right",
                    "show_currency_unit": true,
                    "width": 0
                }
            }
        ],
        [
            {
                "block_id": "attrs",
                "block_type": "label",
                "block_label": "({attrs})",
                "block_attr": {
                    "font_size": 18,
                    "align": "left",
                    "line_height": 35,
                    "width": 52
                },
                "conditions": [
                    {
                        "field": "attrs",
                        "operator": "not_empty",
                        "value": 0
                    }
                ]
            }
        ]
    ]
}
```

**商品数据结构**:
```json
{
    "name": "商品名称",
    "price_num": "1*1",
    "price": "商品单价",
    "num": 12,
    "subtotal": "小计金额",
    "attrs": "份;冰;香菜;洋葱",
    "attr": "主要属性",
    "attr_list": [
        {
            "name": "冰",
            "text": "冰"
        }
    ],
    "flavor_name": "份",
    "sauce_list": [
        {
            "name": "香菜",
            "text": "香菜"
        }
    ],
    "sauce_names": "香菜;洋葱",
    "is_delay": false,
    "is_buffet": false,
    "is_buffet_product": false,
    "is_gift": false,
    "is_package": false,
    "is_sub_product": false
}
```

### 3. 金额汇总区域

#### 商品统计
显示订单中所有商品的总数量。

**示例配置**:
```json
{
    "block_id": "order.product_num",
    "block_type": "label:value",
    "block_label": {
        "zh": "商品数量:",
        "en": "Total text:",
        "ko": "상품 수량:",
        "my": "ကုန်ပစ္စည်းအရေအတွက်:",
        "tr": "Toplam Miktar:",
        "de": "Gesamtmenge:",
        "sv": "Total kvantitet:",
        "ja": "商品数量:",
        "zhtw": "商品數量:",
        "th": "จำนวนสินค้า:"
    },
    "block_attr": {
        "font_size": 20,
        "align": "left",
        "width": 100
    }
}
```

#### 金额明细
包含以下金额字段的条件显示：

| 字段 | 说明 | 显示条件 | 对齐方式 |
|------|------|----------|----------|
| `order.product_amount` | 商品总金额 | 不为空时显示 | 右对齐 |
| `order.service_fee` | 服务费 | 不为空时显示 | 右对齐 |
| `order.tax_fee` | 税费 | 不为空时显示 | 右对齐 |
| `order.discount_fee` | 优惠折扣 | 不为空时显示 | 右对齐 |
| `order.member_discount_fee` | 会员优惠 | 不为空时显示 | 右对齐 |
| `order.member_discount_rate` | 会员折扣率 | 不为空时显示 | 右对齐 |
| `order.coupon_exchange_amount` | 优惠券兑换 | 不为空时显示 | 右对齐 |
| `order.return_amount` | 退款金额 | 不为空时显示 | 右对齐 |
| `order.payment_commission_fee` | 支付手续费 | 不为空时显示 | 右对齐 |
| `order.free_amount` | 免费金额 | 不为空时显示 | 右对齐 |
| `order.actual_receive_price` | 实际收款金额 | 不为空时显示 | 右对齐（粗体） |

**示例配置**:
```json
{
    "block_id": "order.product_amount",
    "block_type": "label:value",
    "block_label": {
        "zh": "商品总金额:",
        "en": "Product Amount:",
        "ko": "상품 총액:",
        "my": "ကုန်ပစ္စည်းစုစုပေါင်း:",
        "tr": "Ürün Tutarı:",
        "de": "Produktbetrag:",
        "sv": "Produktbelopp:",
        "ja": "商品総額:",
        "zhtw": "商品總金額:",
        "th": "ยอดรวมสินค้า:"
    },
    "block_attr": {
        "font_size": 20,
        "align": "right",
        "show_currency_unit": true,
        "width": 100
    }
}
```

**实际收款金额（重点）**:
```json
{
    "block_id": "order.actual_receive_price",
    "block_type": "label:value",
    "block_label": {
        "zh": "实际收款:",
        "en": "Actual Receive:",
        "ko": "실제 수령:",
        "my": "အမှန်တကယ်လက်ခံ:",
        "tr": "Gerçek Alınan:",
        "de": "Tatsächlich erhalten:",
        "sv": "Faktiskt mottaget:",
        "ja": "実際の受取:",
        "zhtw": "實際收款:",
        "th": "เงินที่ได้รับจริง:"
    },
    "block_attr": {
        "font_size": 22,
        "align": "right",
        "font_bold": true,
        "show_currency_unit": true,
        "width": 100
    }
}
```

### 4. 支付信息区域

#### 支付方式
显示订单的支付方式信息。

**示例配置**:
```json
{
    "block_id": "order.payment_name",
    "block_type": "label:value",
    "block_label": {
        "zh": "支付方式:",
        "en": "Payment Method:",
        "ko": "결제 방법:",
        "my": "ငွေပေးချေမှုနည်းလမ်း:",
        "tr": "Ödeme Yöntemi:",
        "de": "Zahlungsmethode:",
        "sv": "Betalningsmetod:",
        "ja": "支払い方法:",
        "zhtw": "支付方式:",
        "th": "วิธีการชำระเงิน:"
    },
    "block_attr": {
        "font_size": 20,
        "align": "left",
        "width": 100
    }
}
```

#### 支付二维码
显示支付相关的二维码（如微信支付、支付宝等）。

**示例配置**:
```json
{
    "block_id": "order.payment_qrcode",
    "block_type": "qrcode",
    "block_attr": {
        "font_size": 22,
        "align": "center",
        "width": 500
    },
    "conditions": [
        {
            "field": "order.payment_qrcode",
            "operator": "not_empty",
            "value": 0
        }
    ]
}
```

#### 订单条形码
显示订单的条形码（通常为订单编号的条形码形式）。

**示例配置**:
```json
{
    "block_id": "order.barcode",
    "block_type": "barcode",
    "block_attr": {
        "font_size": 22,
        "align": "center",
        "width": 500
    },
    "conditions": [
        {
            "field": "order.barcode",
            "operator": "not_empty",
            "value": 0
        }
    ]
}
```

### 5. 页脚信息区域

#### 感谢信息
显示感谢顾客光临的信息，包含品牌名称变量。

**示例配置**:
```json
{
    "block_id": "title",
    "block_type": "label",
    "block_label": {
        "zh": "感谢您的光临！本店由 {brand_name} 系统提供支持。",
        "en": "Thank you for your visit! This store is powered by {brand_name} system.",
        "ko": "방문해 주셔서 감사합니다! 이 매장은 {brand_name} 시스템으로 지원됩니다.",
        "my": "လာရောက်ခြင်းအတွက် ကျေးဇူးတင်ပါတယ်! ဤဆိုင်ကို {brand_name} စနစ်ဖြင့် ပံ့ပိုးထားပါတယ်။",
        "tr": "Ziyaretiniz için teşekkür ederiz! Bu mağaza tarafından: {brand_name} Sistem destek sağlar.",
        "de": "Vielen Dank für Ihren Besuch! Dieser Laden wird vom {brand_name} System unterstützt.",
        "sv": "Tack för ditt besök! Denna butik drivs av {brand_name} systemet.",
        "ja": "ご来店ありがとうございます！この店舗は {brand_name} システムでサポートされています。",
        "zhtw": "感謝您的光臨！本店由 {brand_name} 系統提供支持。",
        "th": "ขอบคุณที่แวะมาหากัน!สนับสนุนโดย {brand_name}"
    },
    "block_attr": {
        "font_size": 20,
        "align": "center",
        "font_bold": true,
        "trailing_blank_lines": 6,
        "width": 100
    }
}
```

## 📊 数据结构要求

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
                "price_num": "1*1",
                "price": "商品单价",
                "num": 12,
                "subtotal": "小计金额",
                "remark": "商品备注",
                "attrs": "份;冰;香菜;洋葱",
                "attr": "主要属性",
                "attr_list": [
                    {
                        "name": "冰",
                        "text": "冰"
                    }
                ],
                "flavor_name": "份",
                "sauce_list": [
                    {
                        "name": "香菜",
                        "text": "香菜"
                    }
                ],
                "sauce_names": "香菜;洋葱",
                "is_delay": false,
                "is_buffet": false,
                "is_buffet_product": false,
                "is_gift": false,
                "is_package": false,
                "is_sub_product": false
            }
        ],
        "product_num": 0,
        "product_amount": "29,297.01",
        "service_fee": "0",
        "tax_rate": 0,
        "tax_fee_type": 2,
        "is_contain_tax": 2,
        "discount_fee": "0.01",
        "discount_rate": "0",
        "member_discount_fee": "0",
        "member_discount_rate": 0,
        "member_card_discount_rate": 0,
        "member_points_discount": 0,
        "coupon_exchange_amount": 0,
        "check_out_zero_fee": "0",
        "return_amount": "0",
        "payment_commission_fee": 0,
        "free_amount": "0",
        "actual_receive_price": "29,297",
        "payment_methods": [],
        "percentage_lists": [],
        "is_free": false,
        "is_member": false,
        "member_remaining_balance": "0",
        "member_points": 0,
        "payment_name": "",
        "payment_qrcode": "",
        "barcode": "202509252185102915"
    }
}
```

### 品牌信息 (brand_name)
```json
{
    "brand_name": "TTPOS"
}
```

## 🚀 使用示例

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

## ⚠️ 注意事项

1. **数据完整性**: 确保所有必需的数据字段都已提供
2. **语言设置**: 根据目标用户设置正确的语言代码
3. **图片路径**: Logo图片路径必须是有效的文件路径
4. **金额格式**: 金额字段建议使用字符串格式，避免精度问题
5. **条件显示**: 只有满足条件的字段才会显示，确保数据正确性
6. **多语言文本**: 确保所有多语言标签都有对应的翻译
7. **模板版本**: 注意模板版本兼容性
8. **纸张宽度**: 根据实际打印机设置正确的纸张宽度

## 🔧 扩展说明

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

---

这个模板为餐饮收银系统提供了完整的结账单打印功能，支持多语言、多货币、条件显示等高级特性，能够满足各种复杂的业务需求。
