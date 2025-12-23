# KBTG Link POS Specification v1.5.0

> **当前版本**: 1.5.0 (25 March 2025)  
> **文档概述**: 本文档提供 POS 与 KEDC 设备连接的消息规格，基于 (1) Hypercom 和 (2) POSNET 标准。

---

## 文档版本控制

| 版本  | 日期            | 修改人                                        | 说明                                                                 |
| ----- | --------------- | --------------------------------------------- | -------------------------------------------------------------------- |
| 1.0.0 | 10 November 2023| Seri Namsri                                   | 首次发布                                                             |
| 1.1.0 | 16 March 2024   | Supoj Ploytaveechai                           | 新增 POSNET 规格                                                     |
| 1.1.1 | 28 October 2024 | Chitsanupong Norasing                         | 新增按 REF1 检索交易规格                                             |
| 1.2.0 | 04 August 2024  | Chitsanupong Norasing, Kriangkrai Khuabsoongnern | 新增 CombineQR 规格, Field A1 新增"QR PAYMENT", 新增 E-TAX 规格    |
| 1.3.0 | 11 February 2025| Chitsanupong Norasing                         | 更新按 REF1 检索交易规格                                             |
| 1.3.1 | 13 February 2025| Chitsanupong Norasing                         | 新增 QR Payment 请求和响应示例                                       |
| 1.4.0 | 24 March 2025   | Tanparit Hiranpijit                           | 新增同意接受交易、公民信息请求交易、QR Payment 查询和取消            |
| 1.5.0 | 25 March 2025   | Parkphoom Didsayamarn                         | **新增 Kiosk 规格**                                                  |

---

# Part 1: HYPERCOM 消息规格

## 消息结构

POS 和终端之间传输的消息采用以下结构：

```
+-----+------+-------------+-----+-----+
| STX | LLLL | MESSAGE DATA| ETX | LRC |
+-----+------+-------------+-----+-----+
```

| 字段         | 字节 | 值        | 说明                                                                                                                                     |
| ------------ | ---- | --------- | ---------------------------------------------------------------------------------------------------------------------------------------- |
| STX          | 1    | 02h 或 82 | **Start of Text**: 帧起始标识<br>- 普通交易: 02<br>- **E-TAX 交易**: 82                                                                  |
| LLLL         | 2    |           | MESSAGE DATA 长度，BCD 格式。例如: 256 字节 = 02h 56h<br>**E-TAX**: 长度需转换为 16 进制 (256 = 0100)                                    |
| MESSAGE DATA | 可变 |           | 消息数据，包含 Transport Header、Presentation Header 和 Field Data                                                                       |
| ETX          | 1    | 03h       | **End of Text**: 消息结束标识                                                                                                            |
| LRC          | 1    |           | **Longitudinal Redundancy Character**: 从 STX 到 ETX 的所有字符异或计算                                                                  |

---

## MESSAGE DATA 结构

MESSAGE DATA 由三部分组成:

```
+------------------+---------------------+------------+
| Transport Header | Presentation Header | Field Data |
+------------------+---------------------+------------+
```

### Transport Header

| 字段                  | 字节 | 值   | 用途                                                                                 |
| --------------------- | ---- | ---- | ------------------------------------------------------------------------------------ |
| Transport Header Type | 2    | 60   | 应用类型，'60' 表示应用消息                                                          |
| Transport Destination | 4    | 0000 | 路由目的地。'0000' 时由终端根据内部表决定路由                                        |
| Transport Source      | 4    | 0000 | 前两位必须是 '00'。后两位可由 POS 用于标识连接的多个设备                             |

### Presentation Header

| 字段                        | 字节 | 值  | 用途                                                                                                                          |
| --------------------------- | ---- | --- | ----------------------------------------------------------------------------------------------------------------------------- |
| Format Version              | 1    | 1   | 支持的消息格式版本                                                                                                            |
| Request-Response Indicator  | 1    |     | 消息类型:<br>`'0'` 需要响应的请求<br>`'1'` 响应消息<br>`'2'` 不需要响应的请求                                                 |
| Transaction Code            | 2    |     | 交易类型标识，见"交易代码"章节                                                                                                |
| Response Code               | 2    |     | 交易结果。请求消息设为 `'00'`                                                                                                 |
| More Indicator              | 1    |     | `'0'` 最后一条消息<br>`'1'` 还有后续消息                                                                                      |
| Field Separator             | 1    | 1Ch | Presentation Header 结束标识                                                                                                  |

---

## Field Data 结构

Field Data 由 0 个或多个 Field Element 组成：

```
+------------+------+------+-----------------+
| Field Type | LLLL | Data | Field Separator |
+------------+------+------+-----------------+
```

| 字段            | 字节 | 值  | 用途                                                      |
| --------------- | ---- | --- | --------------------------------------------------------- |
| Field Type      | 2    |     | 数据类型标识 (0-9, A-Z)                                   |
| LLLL            | 2    |     | 数据长度 (BCD 格式)，例如 256 字节 = 02h 56h              |
| Data            | LLLL |     | 字段数据                                                  |
| Field Separator | 1    | 1Ch | 字段结束标识                                              |

---

## 交易代码 (Transaction Codes)

### 基础交易

| 交易代码 | 描述                                           |
| -------- | ---------------------------------------------- |
| 20       | Sale (刷卡消费)                                |
| 70       | Sale Alipay/Wechat/QR Payment/Other wallet     |
| 26       | Void (撤销)                                    |
| 27       | Refund (退款)                                  |
| 34       | Balance Inquiry (余额查询)                     |
| 41       | Installment (分期付款)                         |
| 43       | Redemption (积分兑换)                          |
| 50       | Settlement (结算)                              |
| 71       | Void transaction by trace number               |
| 74       | Inquiry transaction by Reference 1             |

### 公民信息交易

| 交易代码 | 描述                                           |
| -------- | ---------------------------------------------- |
| 97       | Request for citizen information transaction    |
| CS       | Accept agreement transaction                   |

### E-TAX 交易

| 交易代码 | 描述                            |
| -------- | ------------------------------- |
| EA       | Sale + E-TAX                    |
| EB       | Sale Wallet + E-TAX             |
| ED       | Installment + E-TAX             |
| EF       | Void + E-TAX                    |
| EH       | Create E-TAX                    |
| EI       | Void E-TAX Only                 |
| EJ       | E-TAX Inquiry (All Hosts)       |
| EL       | E-TAX Inquiry (Single transaction)|
| EK       | Settlement E-TAX                |

### QR Payment 查询/取消

| 交易代码 | 描述                                           |
| -------- | ---------------------------------------------- |
| IQ       | Inquiry Alipay/Wechat/QR Payment/Other wallet  |
| CQ       | Cancel Alipay/Wechat/QR Payment/Other wallet   |

### Kiosk 专用交易 (v1.5.0 新增) ⭐

| 交易代码 | 描述                    | 方向       | 说明                                     |
| -------- | ----------------------- | ---------- | ---------------------------------------- |
| C0       | Waiting card insert     | EDC → POS  | 等待插卡通知                             |
| C1       | Chip card inserted      | EDC → POS  | 芯片卡已插入                             |
| C2       | Waiting card remove     | EDC → POS  | 等待拔卡通知                             |
| C3       | Chip card removed       | EDC → POS  | 芯片卡已移除                             |
| P1       | Waiting PIN entry       | EDC → POS  | 等待 PIN 输入                            |
| P2       | PIN entry finished      | EDC → POS  | PIN 输入完成                             |
| P3       | PIN timeout             | EDC → POS  | PIN 输入超时                             |
| P4       | PIN cancel              | EDC → POS  | PIN 输入取消                             |
| D0       | Communication test      | 双向       | 通讯测试                                 |
| D1       | Communication cancel    | EDC → POS  | 通讯取消                                 |
| I1       | Card Information Inquiry|            | 卡信息查询                               |

> **注意**: Kiosk 交易代码仅在启用 KIOSK 模式时生效，普通 POS 不受影响。均为单向通讯，无需响应。

---

## 字段类型 (Field Types)

### 字段属性

| 属性 | 描述                                                         |
| ---- | ------------------------------------------------------------ |
| ANS  | Alpha, Numeric and Special characters                        |
| AN   | Alpha and Numeric character                                  |
| N    | Numeric characters                                           |
| Z    | Numeric + '=' (用于 Track 2)                                 |
| B    | Binary Data                                                  |
| T    | Telephone Number (0-9, *, #, space, -, comma, P)             |

### 通用字段定义

| Field Type | 属性 | 长度   | 数据                        |
| ---------- | ---- | ------ | --------------------------- |
| 01         | ANS  | 6      | Approval Code               |
| 02         | ANS  | 40     | Response Text (2行×20字符)  |
| 03         | N    | 6      | Transaction Date (YYMMDD)   |
| 04         | N    | 6      | Transaction Time (HHMMSS)   |
| 16         | N    | 8      | Terminal ID                 |
| 30         | N    | ..40   | Primary Account Number      |
| 31         | N    | 4      | Expiration Date (YYMM)      |
| 40         | N    | 12     | Amount (交易金额，单位：分) |
| 50         | N    | 6      | Batch No.                   |
| 65         | ANS  | 6      | Invoice No. (Trace No.)     |
| D0         | ANS  | 96     | Merchant Name and Address   |
| D1         | ANS  | 15     | Merchant ID                 |
| D2         | ANS  | ..20   | Card Issuer Name            |
| D3         | ANS  | 12     | Reference No.               |
| D5         | ANS  | 26     | Card Holder Name            |
| R1         | ANS  | 20     | Reference 1                 |
| R2         | ANS  | 20     | Reference 2                 |

### Kiosk 专用字段 (v1.5.0 新增) ⭐

| Field Type | 属性 | 长度 | 数据                           |
| ---------- | ---- | ---- | ------------------------------ |
| FH         | ANS  | 7    | AID (Kiosk 可选)               |
| FI         | ANS  | 5    | TVR (Kiosk 可选)               |
| FJ         | ANS  | 7    | TSI (Kiosk 可选)               |

### ALIPAY / WECHAT 字段定义

| Field Type | 属性 | 长度  | 数据             |
| ---------- | ---- | ----- | ---------------- |
| A1         | AN   | ..20  | Host Name        |
| A3         | ANS  | ..25  | QRCODE 数据      |
| A4         | ANS  | ..40  | Transaction ID   |
| A5         | ANS  | ..2   | Bank ID          |
| A6         | ANS  | 2     | Payment Indicator|

---

## 响应代码 (Response Codes)

| 响应代码 | 含义                                                    |
| -------- | ------------------------------------------------------- |
| 00       | **Approved** - 交易被 ECR/主机批准                      |
| ND       | **Declined** - 交易被拒绝，详情见 Response Text 字段    |
| NB       | **No Batch** - 结算时批次为空                           |

---

## 交易消息详情

### 1. Sale (Transaction Code = 20)

**POS → EDC 请求:**

| Field Type | 属性 | 长度 | 数据               | 必填 | 示例                    |
| ---------- | ---- | ---- | ------------------ | ---- | ----------------------- |
| 40         | ANS  | 12   | Amount Transaction | M    | 000000001000 = 10.00    |
| R1         | ANS  | 20   | Reference 1        | O    |                         |
| R2         | ANS  | 20   | Reference 2        | O    |                         |

**请求示例:**
```
02003536303030303030303030313032303030311C343000123030303030303030303130301C0315
```

**EDC → POS 响应:**

| Field Type | 属性 | 长度  | 数据                       |
| ---------- | ---- | ----- | -------------------------- |
| 02         | ANS  | 40    | Response Text              |
| D0         | ANS  | 96    | Merchant Name and Address  |
| 03         | N    | 6     | Transaction Date (YYMMDD)  |
| 04         | N    | 6     | Transaction Time (HHMMSS)  |
| 01         | ANS  | 6     | Approval Code              |
| 65         | ANS  | 6     | Invoice No. (Trace No.)    |
| 16         | N    | 8     | Terminal ID                |
| D1         | ANS  | 15    | Merchant ID                |
| D2         | ANS  | ..20  | Card Issuer Name           |
| 30         | N    | ..40  | Primary Account Number     |
| 31         | N    | 4     | Card Expiry Date (YYMM)    |
| 50         | N    | 6     | Batch No.                  |
| D3         | ANS  | 12    | Reference No.              |
| 40         | ANS  | 12    | Transaction Amount         |
| D5         | ANS  | 26    | Card Holder Name           |
| R1         | ANS  | 20    | Reference 1 (Optional)     |
| R2         | ANS  | 20    | Reference 2 (Optional)     |
| FH         | ANS  | 7     | AID (Kiosk Optional) ⭐    |
| FI         | ANS  | 5     | TVR (Kiosk Optional) ⭐    |
| FJ         | ANS  | 7     | TSI (Kiosk Optional) ⭐    |

**成功响应示例:**
```
02042536303030303030303030313132303030311C30320040535543434553532020...
```

**失败响应示例:**
```
02006336303030303030303030313132304E44311C30320040486F73742052656A65...
```

---

### 2. Void (Transaction Code = 26)

撤销已完成的交易。

### 3. Settlement (Transaction Code = 50)

#### 单主机结算
- 结算单个主机的批次

#### 所有主机结算  
- 结算所有配置主机的批次

### 5. Sale Wallet (Transaction Code = 70)

支持 Alipay/Wechat/QR Payment/Other wallet 交易。

### 7. Installment (Transaction Code = 41)

分期付款交易，支持:
- Merchant Pay (商户承担利息)
- Customer Pay (客户承担利息)
- Supplier Pay (供应商承担利息)
- Special Interest (特殊利率)

### 8-12. Redemption (Transaction Code = 43)

积分兑换，支持:
- E-Voucher
- E-Voucher + Credit
- Product
- Discount % Fix Point
- Discount % Var Point

### 15. Refund (Transaction Code = 27)

退款交易。

### 16. Inquiry by Reference 1 (Transaction Code = 74)

按 Reference 1 查询交易。

---

## E-TAX 交易

E-TAX 交易需要将 STX 设为 `82` (而非普通的 `02`)。

### E-TAX 专用字段

| Field Type | 属性 | 长度  | 数据                      |
| ---------- | ---- | ----- | ------------------------- |
| T0         | ANS  | 32    | Tax ID                    |
| T1         | ANS  | 32    | Branch ID                 |
| T2         | ANS  | 32    | POS ID                    |
| T3         | ANS  | 32    | Invoice No.               |
| T4         | ANS  | 12    | Total Tax Amount          |
| T5         | ANS  | ..100 | E-TAX QR String           |
| T6         | ANS  | 10    | E-TAX Date                |
| T7         | ANS  | 8     | E-TAX Time                |

---

## QR Payment 查询与取消 (v1.4.0+)

### Inquiry QR Payment (Transaction Code = IQ)

查询 Alipay/Wechat/QR Payment 交易状态。

### Cancel QR Payment (Transaction Code = CQ)

取消进行中的 QR Payment 交易。

---

## Kiosk 规格 (v1.5.0 新增) ⭐

### 概述

Kiosk 模式专为自助服务终端设计，提供实时状态通知：

1. **卡片状态通知** (C0-C3): 插卡/拔卡状态
2. **PIN 状态通知** (P1-P4): PIN 输入进度
3. **通讯控制** (D0-D1): 连接测试和取消

### Kiosk 交易详情

#### 30. Waiting card insert (C0)

```
方向: EDC → POS (单向通讯，无需响应)
触发: 收到 Transaction Code 20 后
适用: 仅 KIOSK 模式
示例: 02001836303030303030303030313143303030311C0343
```

#### 31. Chip card inserted (C1)

```
方向: EDC → POS (单向)
触发: 芯片卡插入
示例: 02001836303030303030303030313143313030311C0342
```

#### 32. Waiting card remove (C2)

```
方向: EDC → POS (单向)
触发: 等待拔卡
示例: 02001836303030303030303030313143323030311C0341
```

#### 33. Chip card removed (C3)

```
方向: EDC → POS (单向)
触发: 卡已移除
示例: 02001836303030303030303030313143333030311C0340
```

#### 34. Waiting PIN entry (P1)

```
方向: EDC → POS (单向)
触发: 等待 PIN 输入
示例: 02001836303030303030303030313150313030311C0351
```

#### 35. PIN entry finished (P2)

```
方向: EDC → POS (单向)
触发: PIN 输入完成
示例: 02001836303030303030303030313150323030311C0352
```

#### 36. PIN timeout (P3)

```
方向: EDC → POS (单向)
触发: PIN 输入超时
示例: 02001836303030303030303030313150333030311C0353
```

#### 37. PIN cancel (P4)

```
方向: EDC → POS (单向)
触发: PIN 输入被取消
示例: 02001836303030303030303030313150343030311C0354
```

#### 38. Communication test (D0)

```
方向: 双向
功能: 通讯测试，EDC 会响应相同的 Transaction Code
示例: 02001836303030303030303030313144303030311C0344
```

#### 39. Communication cancel (D1)

```
方向: EDC → POS
功能: 取消通讯
响应: 0x0000 表示成功取消；如果交易已完成则返回其他值
示例: 02001836303030303030303030313144313030311C0345
```

---

# Part 2: POSNET 消息规格

## 概述

POSNET 是另一种 POS-EDC 通讯标准，与 HYPERCOM 并行支持。

## POSNET 交易类型

| 交易类型 | 描述                        |
| -------- | --------------------------- |
| SALE     | 消费                        |
| VOID     | 撤销                        |
| SETT     | 结算                        |
| REFU     | 退款                        |
| INST     | 分期 (Merchant/Customer/Supplier/Special) |
| REDE     | 积分兑换                    |
| REDI     | 积分查询                    |
| ETAX     | E-TAX 相关交易              |

## POSNET 字段定义

| Field Type | 属性 | 长度  | 数据                    |
| ---------- | ---- | ----- | ----------------------- |
| 01         | ANS  | 6     | Approval Code           |
| 02         | ANS  | 40    | Response Text           |
| 03         | N    | 6     | Transaction Date        |
| 04         | N    | 6     | Transaction Time        |
| 16         | N    | 8     | Terminal ID             |
| 30         | N    | ..40  | PAN (卡号)              |
| 40         | N    | 12    | Amount                  |

### POSNET QR 字段

| Field Type | 属性 | 长度  | 数据               |
| ---------- | ---- | ----- | ------------------ |
| A1         | AN   | ..20  | Host Name          |
| A3         | ANS  | ..25  | QRCODE             |
| A4         | ANS  | ..40  | Transaction ID     |

### POSNET E-TAX 字段

| Field Type | 属性 | 长度  | 数据               |
| ---------- | ---- | ----- | ------------------ |
| T0         | ANS  | 32    | Tax ID             |
| T1         | ANS  | 32    | Branch ID          |
| T2         | ANS  | 32    | POS ID             |
| T3         | ANS  | 32    | Invoice No.        |
| T4         | ANS  | 12    | Total Tax Amount   |
| T5         | ANS  | ..100 | E-TAX QR String    |

---

## POSNET 交易消息

### 1. Sale (SALE)

标准刷卡消费交易。

### 2. Sale - QR (SALE)

QR 支付消费交易，支持 Alipay/Wechat/Thai QR。

### 3. Void (VOID)

撤销交易。

### 4. Settlement (SETT)

批次结算。

### 5. Refund (REFU)

退款交易。

### 6-9. Installment (INST)

分期付款，支持多种模式:
- Merchant Pay
- Customer Pay  
- Supplier Pay
- Special Interest

### 10-15. Redemption (REDE/REDI)

积分相关:
- E-Voucher
- E-Voucher + Credit
- Product
- Discount % Fix
- Discount % Var
- Balance Point Inquiry

### 16-24. E-TAX (ETAX)

电子税务相关交易。

---

## 附录: Bank ID 参考

| Bank ID | 银行名称                    |
| ------- | --------------------------- |
| 01      | BBL                         |
| 02      | KBank                       |
| 03      | SCB                         |
| 04      | TMB                         |
| 05      | KTB                         |
| 06      | BAY                         |
| 07      | CITI                        |
| 08      | TBANK                       |
| ...     | (更多请参考原文档)          |

## 附录: Payment Indicator (PI)

| PI | 含义              |
| -- | ----------------- |
| 01 | Contact Chip      |
| 02 | Contactless       |
| 03 | Swipe             |
| 04 | Manual Entry      |
| 05 | QR Payment        |

---

## 关键差异: v1.5.0 vs 旧版本

| 特性             | 旧版本     | v1.5.0              |
| ---------------- | ---------- | ------------------- |
| Kiosk 支持       | ❌         | ✅                  |
| 卡片状态通知     | ❌         | ✅ (C0-C3)          |
| PIN 状态通知     | ❌         | ✅ (P1-P4)          |
| 通讯测试/取消    | ❌         | ✅ (D0-D1)          |
| Sale 响应新增字段| -          | FH (AID), FI (TVR), FJ (TSI) |
| QR 查询/取消     | v1.4.0+    | ✅ (IQ, CQ)         |
| 公民信息交易     | v1.4.0+    | ✅ (97, CS)         |
| E-TAX            | v1.2.0+    | ✅                  |

---

## 完整性说明

本 Markdown 文档涵盖:

- ✅ 文档版本历史
- ✅ HYPERCOM 消息规格 (完整)
- ✅ POSNET 消息规格 (概要)
- ✅ 所有交易代码 (39 种)
- ✅ 字段类型定义
- ✅ 响应代码
- ✅ **Kiosk 规格 (v1.5.0 新增)**
- ✅ E-TAX 规格
- ✅ QR Payment 规格

> **原始 HTML 文件**: 约 17,778 行，3.8MB  
> **本 Markdown 精简版**: 保留核心规格和结构

---

*文档转换日期: 2025-12-23*

