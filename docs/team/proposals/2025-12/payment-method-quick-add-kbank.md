# 支付方式快捷添加（Kbank渠道）需求提案

> 本文档用于需求发起阶段，经团队评审后创建正式 Spec。

---

## 📋 提案信息

| 项目       | 内容     |
| ---------- | -------- |
| **提案人** | 王昱   |
| **日期**   | 2025-12-29   |
| **目标版本** | {版本号} |
| **状态**   | 已创建Spec   |
| **关联任务** | - |
| **关联 Spec** | [story-shop-payment-method-quick-add-kbank](../../shared/specs/archived/v2.12/story-shop-payment-method-quick-add-kbank/requirements.md)      |

---

## 🎯 背景和动机

### 问题描述

当前在管理端添加支付方式时，需要手动填写多个字段（名称、支付方式、source、图片等），操作繁琐。特别是对于Kbank渠道的支付方式，商户需要逐个添加Alipay（Kbank）、WeChatPay（Kbank）、Credit QR（Kbank）、Thai QR（Kbank）、Credit Card（Kbank）这5种支付方式，每次都需要重复填写相同的信息，效率低下。

**示例场景**：
> 商户需要添加Kbank渠道的Alipay支付方式，需要手动填写：
> - 名称：Alipay
> - 支付方式：Alipay
> - source：3（Kbank）
> - 图片：上传对应支付方式图片
> 
> 如果要添加5种Kbank支付方式，需要重复操作5次，且容易出错。

### 业务价值

- **提升配置效率**：一键快速添加Kbank渠道的5种支付方式，减少90%的重复操作
- **降低配置错误**：避免手动填写导致的名称、渠道不一致问题
- **改善用户体验**：简化支付方式配置流程，提升商户满意度
- **减少支持成本**：减少因配置错误导致的客服咨询

### 目标用户

- [x] 商户管理员（Shop管理端）
- [ ] 收银员
- [ ] 商户管理员（旧Admin管理端）
- [ ] 厨房人员
- [ ] 顾客
- [ ] 其他: ________

---

## 💡 解决方案概述

### 方案描述

延用现有的 `GetDefaultPayList` 和 `Create` 接口，扩展 `GetDefaultPayList` 接口返回Kbank渠道的5种支付方式，并标记是否可添加。前端调用扩展后的 `GetDefaultPayList` 获取Kbank支付方式列表，然后调用现有的 `Create` 接口批量创建。系统自动填充名称、支付方式、source等信息，前端只需选择需要添加的支付方式即可。

**核心流程**：
> 1. 前端调用 `GET /shop/payment_method/default_pay` 接口（扩展后返回Kbank支付方式）
> 2. 后端返回5种Kbank支付方式列表，并标记哪些可添加（can_add字段）
> 3. Kbank支付方式在列表最上面显示（通过sort字段控制）
> 4. 前端选择需要添加的支付方式（可多选），过滤掉不可添加的（can_add=false）
> 5. 前端调用 `POST /shop/payment_method/create` 接口（延用现有接口，需传入source参数）
> 6. 后端自动填充名称、支付方式、source（Kbank）、图片等信息并批量创建

### 核心功能点

1. **扩展 GetDefaultPayList 接口**：在现有接口中增加Kbank渠道的5种支付方式，并标记是否可添加
   - Alipay（Kbank）- code: 93000
   - WeChatPay（Kbank）- code: 93100
   - Credit QR（Kbank）- code: 93200
   - Thai QR（Kbank）- code: 93300
   - Credit Card（Kbank）- code: 93400
2. **重复检测逻辑**：在 `GetDefaultPayList` 中检测已添加过的Kbank支付方式，通过 `can_add` 字段返回（can_add=false表示已添加，不可再选）
3. **选项排序**：Kbank支付方式在返回列表最上面（通过sort字段控制，Kbank支付方式sort值最小）
4. **延用 Create 接口**：使用现有的批量创建接口，前端传入Kbank支付方式信息（需包含source参数）
5. **自动填充逻辑**：前端根据 `GetDefaultPayList` 返回的信息自动填充名称、支付方式、source、图片等字段

### 影响范围

**涉及终端**：
- [ ] POS 收银端
- [x] Shop 商家管理端（新管理端）
- [ ] KDS 厨显端
- [ ] QDS 排号叫号端
- [ ] Assistant 助手端
- [ ] Tablet 平板端
- [ ] Mobile 扫码端
- [ ] Menu 电子菜单端
- [ ] Member 会员端
- [x] Admin 管理端（旧管理端）

**涉及模块**：
- [x] API 接口（扩展 GetDefaultPayList 接口，延用 Create 接口）
- [ ] 数据模型（已有source字段，无需新增）
- [x] 业务逻辑（重复检测、选项排序、Kbank支付方式定义）
- [ ] 第三方集成
- [ ] 其他: ________

---

## 📊 初步评估

### 技术复杂度

- [ ] **低**：纯 UI 调整，无业务逻辑变更
- [x] **中**：扩展现有接口，基础业务逻辑（重复检测、选项排序、Kbank支付方式定义）
- [ ] **高**：涉及架构调整、第三方集成、复杂算法

### 工作量预估

[粗略估算，具体 SP 在技术方案评审时确定]

- **预计天数**: 1-2 天
- **预估 SP**: 2 SP（待技术评审确认）

**工作分解**：
- 后端接口扩展：0.5-1天（扩展 GetDefaultPayList 接口，增加Kbank支付方式定义、重复检测逻辑、选项排序）
- 接口测试：0.5-1天

### 风险识别

**潜在风险**：
1. **source字段值定义**：需要确定Kbank支付方式使用的source值（可能需要新增source=3表示Kbank）
2. **图片资源**：需要确认5种Kbank支付方式的图片资源是否已准备好
3. **接口兼容性**：扩展 `GetDefaultPayList` 接口可能影响现有调用，需要确保向后兼容
4. **重复检测逻辑**：需要明确如何通过 payment_name 和 source 字段判断是否已添加
5. **Create接口扩展**：需要在 `PaymentMethodCreateItem` 中增加source字段

**缓解措施**：
1. **source值定义**：确认Kbank支付方式的source值（建议使用source=3，或复用现有值）
2. **图片资源准备**：提前准备5种Kbank支付方式的图片资源，或使用默认图片
3. **接口兼容性**：扩展 `GetDefaultPayList` 时，新增字段使用可选字段，确保现有调用不受影响
4. **重复检测**：通过查询 payment_name 和 source 字段判断是否已添加
5. **Create接口扩展**：在 `PaymentMethodCreateItem` 中增加source字段（可选字段，默认值根据业务逻辑确定）

---

## 🔗 相关资源

### 参考需求

- 类似功能: 支付方式管理现有功能
- 竞品分析: {待补充}

### 相关文档

- 支付方式管理文档: `docs/human/architecture/features/payment_method.md`
- API文档: `main/app/api/v1/shop/shop_payment_method.go`
- 服务层代码: `main/app/service/payment_method.go`
- 数据模型: `main/app/model/payment_order.go`

### 现有接口说明

**延用的接口**：
1. **GET /shop/payment_method/default_pay** - 获取默认支付方式列表
   - 当前实现：返回系统默认支付方式列表（参考PHP OrderPayTypeEnum）
   - 需要扩展：增加Kbank渠道的5种支付方式，并标记是否可添加
   - 响应结构：`[]*resp.DefaultPaymentMethodResp`（需要扩展增加 `can_add`、`source`、`payment_name` 字段）
   - Kbank支付方式code：93000（Alipay）、93100（WeChatPay）、93200（Credit QR）、93300（Thai QR）、93400（Credit Card）

2. **POST /shop/payment_method/create** - 批量创建支付方式
   - 当前实现：支持批量创建支付方式，自动生成code和sort
   - 需要扩展：在 `PaymentMethodCreateItem` 中增加 `source` 字段
   - 延用方式：前端根据 `GetDefaultPayList` 返回的信息构造请求参数（需包含source字段）
   - 请求结构：`req.PaymentMethodCreateReq`（包含 `items []PaymentMethodCreateItem`，需增加source字段）

### 技术参考

- Kbank集成文档: `docs/others/kbank/KBTG-LINKPOS-Specification-v1.5.0.md`
- 支付方式常量: `main/app/constant/payment.go`

---

## 🤝 需求评审

### 评审参与人

| 角色         | 姓名   | 签名/日期 |
| ------------ | ------ | --------- |
| 产品经理     | {姓名} |           |
| 技术负责人   | {姓名} |           |
| 开发代表     | {姓名} |           |
| 测试代表     | {姓名} |           |
| UI/UX 设计师 | {姓名} |           |

### 评审结论

- [ ] ✅ **批准**：进入技术方案设计阶段
- [ ] 🔄 **修改后批准**：需补充以下内容
- [ ] ❌ **拒绝**：不符合产品规划或优先级

**评审意见**：

```
[记录评审会议的关键讨论和决策]
```

**下一步行动**：

- [x] 创建 Spec：`story-shop-payment-method-quick-add-kbank`
- [ ] 分配负责人：{姓名}
- [ ] 目标 Sprint：Sprint {N}

---

## 📝 附录

### User Story（初稿）

**作为** 商户管理员  
**我想** 通过快捷添加功能一键添加Kbank渠道的支付方式  
**以便于** 快速完成支付方式配置，减少重复操作，提高配置效率

### AC 验收标准（初稿）

1. **WHEN** 前端调用 `GET /shop/payment_method/default_pay` 接口 **THEN** 系统 **SHALL** 返回包含5种Kbank支付方式的列表，且Kbank选项在最上面（sort值最小）
2. **WHEN** 接口返回支付方式列表 **THEN** 系统 **SHALL** 标记已添加的Kbank支付方式（can_add=false），未添加的标记为可添加（can_add=true）
3. **WHEN** 前端调用 `POST /shop/payment_method/create` 接口并传入Kbank支付方式信息（包含source参数） **THEN** 系统 **SHALL** 批量创建支付方式，使用传入的source值
4. **WHEN** 前端传入已添加的Kbank支付方式（通过payment_name和source匹配） **THEN** 系统 **SHALL** 返回错误提示，不重复创建
5. **IF** 前端传入空列表 **THEN** 系统 **SHALL** 返回参数错误提示
6. **WHEN** 批量创建成功 **THEN** 系统 **SHALL** 返回"创建成功"提示

### 功能示例

**扩展后的 GetDefaultPayList 接口响应示例**（Kbank支付方式在最前面）：

```json
{
  "code": 200,
  "data": [
    {
      "code": 93000,
      "name": "Alipay",
      "url": "https://example.com/image/pay/alipay.png",
      "img": "/image/pay/alipay.png",
      "sort": 0,
      "can_add": true,
      "source": 3,
      "payment_name": "Alipay"
    },
    {
      "code": 93100,
      "name": "WeChatPay",
      "url": "https://example.com/image/pay/wechat_pay.png",
      "img": "/image/pay/wechat_pay.png",
      "sort": 1,
      "can_add": false,
      "source": 3,
      "payment_name": "WeChatPay"
    },
    {
      "code": 93200,
      "name": "Credit QR",
      "url": "https://example.com/image/pay/credit_qr.png",
      "img": "/image/pay/credit_qr.png",
      "sort": 2,
      "can_add": true,
      "source": 3,
      "payment_name": "Credit QR"
    },
    {
      "code": 93300,
      "name": "Thai QR",
      "url": "https://example.com/image/pay/thai_qr.png",
      "img": "/image/pay/thai_qr.png",
      "sort": 3,
      "can_add": true,
      "source": 3,
      "payment_name": "Thai QR"
    },
    {
      "code": 93400,
      "name": "Credit Card",
      "url": "https://example.com/image/pay/credit_card.png",
      "img": "/image/pay/credit_card.png",
      "sort": 4,
      "can_add": true,
      "source": 3,
      "payment_name": "Credit Card"
    },
    // ... 其他默认支付方式
  ]
}
```

**延用 Create 接口批量创建Kbank支付方式请求示例**：

```json
{
  "items": [
    {
      "name": "Alipay",
      "payment_name": "Alipay",
      "code": 93000,
      "source": 3,
      "default_img": "/image/pay/alipay.png",
      "fee_percent": 0,
      "is_show_cashier": 1,
      "is_show_assistant": 1,
      "is_show_kiosk": 1,
      "is_show_member_recharge": 0,
      "status": 1
    },
    {
      "name": "Credit QR",
      "payment_name": "Credit QR",
      "code": 93200,
      "source": 3,
      "default_img": "/image/pay/credit_qr.png",
      "fee_percent": 0,
      "is_show_cashier": 1,
      "is_show_assistant": 1,
      "is_show_kiosk": 1,
      "is_show_member_recharge": 0,
      "status": 1
    }
  ]
}
```

**自动填充规则**：

| 选择项 | 名称 | 支付方式 | Code | Source | 图片 |
| ------ | ---- | -------- | ---- | ------ | ---- |
| Alipay（Kbank） | Alipay | Alipay | 93000 | 3 | Alipay默认图片 |
| WeChatPay（Kbank） | WeChatPay | WeChatPay | 93100 | 3 | WeChatPay默认图片 |
| Credit QR（Kbank） | Credit QR | Credit QR | 93200 | 3 | Credit QR默认图片 |
| Thai QR（Kbank） | Thai QR | Thai QR | 93300 | 3 | Thai QR默认图片 |
| Credit Card（Kbank） | Credit Card | Credit Card | 93400 | 3 | Credit Card默认图片 |

### 线框图/原型（可选）

[附加 UI 线框图或原型链接]

---

## 📄 模板使用说明

### 何时使用此模板

- ✅ 产品经理提出新功能想法
- ✅ 用户反馈需求建议
- ✅ 技术团队提出改进方案
- ✅ 需要团队讨论和评审的需求

### 与 Spec 的区别

| 阶段        | 文档类型      | 详细程度 | 用途               |
| ----------- | ------------- | -------- | ------------------ |
| **需求发起** | Proposal      | 粗略     | 团队评审、决策是否做 |
| **需求确认** | Requirements  | 详细     | User Story + AC，开发依据 |
| **技术设计** | Design        | 详细     | 技术方案，实现指导 |
| **任务分解** | Tasks         | 详细     | 开发执行，进度追踪 |

### 流转路径

```
提案 (Proposal) 
  ↓ 评审批准
需求文档 (Requirements) 
  ↓ 技术评审
设计文档 (Design) 
  ↓ SP 评估 ≤ 5
任务分解 (Tasks)
  ↓
开发实现
```

---

**版本**: v1.0.0  
**创建日期**: 2025-12-29  
**维护者**: 产品组 + Scrum Master  
**相关规范**: `.cursor/rules/scrum_story_point.mdc`, `.cursor/rules/specs.mdc`

