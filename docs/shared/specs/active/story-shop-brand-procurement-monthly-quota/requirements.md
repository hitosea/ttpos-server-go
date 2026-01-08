# 品牌采购限额控制 需求文档

> 本文档定义品牌采购限额控制功能的详细需求和验收标准（包含申请次数限制、物品数量限制、月度限购）。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-12/brand-purchase-quota.md](../../../../team/proposals/2025-12/brand-purchase-quota.md) |
| **创建日期**      | 2026-01-07                                                                                                 |
| **负责人**        | BenDaye                                                                                                       |
| **目标 Sprint**   | Sprint 2.14.0                                                                                                   |
| **涉及技术栈**    | [x] Go (main/) [ ] Go (ttpos-bmp/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                   |
| **版本类型**      | MVP（最小可行产品）                                                                                                   |
| **关联任务**      | DooTask #38158, DooTask #38537                                                                                                   |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 已通过 |
| **审核人**   | 产品团队             |
| **审核日期** | 2026-01-07             |
| **审核意见** | MVP 范围合理，可进入技术设计阶段         |

---

## 📋 概述

在品牌连锁经营模式下，需要对品牌采购申请进行多维度的限额控制：

1. **申请次数限制**：限制门店每天提交品牌采购申请的次数，避免频繁提交
2. **单次数量限制**：限制单次申请中物品的总数量，避免单次申请过多
3. **物品月度限购**：限制特定物品每月的采购总量，避免过度采购

当子店提交品牌采购申请时，系统实时校验以上三个维度，任一维度超限则拒绝提交。

**MVP 范围**：

- 核心限购校验逻辑（后端）
- 限购配置数据模型
- 暂不包含前端配置页面（通过数据库直接维护）

## 🎯 产品对齐

- 控制门店每天采购申请的次数，避免频繁申请，提升采购部门效率
- 控制单次申请的物品数量，避免过度采购，优化库存周转
- 控制子店每月采购的物品数量，避免长期过度采购
- 优化总部库存周转，减少积压
- 建立规范的采购管控机制
- 为未来精细化运营打下基础

## 📝 用户故事

**作为** 门店采购员  
**我想** 在发起品牌采购时受到每月物品限额约束  
**以便于** 避免过度采购，规范采购行为

---

## 功能需求

### Requirement 1: 申请次数限制

**用户故事**: 作为采购管理员，我想限制门店每天提交品牌采购申请的次数，以便于减少审批工作量，规范采购行为。

#### 验收标准

1. **WHEN** 门店当天已提交 2 次品牌采购申请（状态为"已提交"）**THEN** 系统 **SHALL** 拒绝第 3 次提交并提示"今日申请次数已达上限（2次），请明天再试"
2. **WHEN** 用户保存草稿（未提交）**THEN** 系统 **SHALL** 不计入申请次数，草稿可以随时编辑和提交
3. **WHEN** 跨天后（店铺时区 00:00:00）**THEN** 系统 **SHALL** 重置申请次数计数
4. **WHEN** 申请被驳回 **THEN** 系统 **SHALL** 仍然计入当天申请次数（不释放）

#### 具体要求

- [x] 1.1 统计门店当天已提交的品牌采购申请次数
- [x] 1.2 只统计状态为"已提交"（非草稿）的申请
- [x] 1.3 使用店铺时区计算"当天"范围（00:00:00 ~ 23:59:59）
- [x] 1.4 限制每天最多 2 次申请（可配置）
- [x] 1.5 超出次数时返回明确错误提示

---

### Requirement 2: 单次申请数量限制

**用户故事**: 作为采购管理员，我想限制单次品牌采购申请的物品总数量，以便于控制采购规模，优化库存管理。

#### 验收标准

1. **WHEN** 单次申请中所有物品的总数量 > 配置上限 **THEN** 系统 **SHALL** 拒绝提交并提示"本次申请物品数量已超限（最多 {limit} 件），请减少物品数量后重试"
2. **WHEN** 单次申请中物品总数量 ≤ 配置上限 **THEN** 系统 **SHALL** 继续后续校验
3. **IF** 未配置数量上限 **THEN** 系统 **SHALL** 跳过数量限制校验

#### 具体要求

- [x] 2.1 统计当前申请中所有物品的总数量（sum of all item.num）
- [x] 2.2 读取单次申请数量上限配置
- [x] 2.3 超出数量时返回明确错误提示（包含当前数量和上限）

---

### Requirement 3: 限购配置数据模型

**用户故事**: 作为系统管理员，我想配置物品的月度采购限额并指定应用的门店，以便于系统按门店进行校验。

#### 验收标准

1. **WHEN** 新增限购配置记录 **THEN** 系统 **SHALL** 存储物品、单位、限额、应用门店等信息
2. **WHEN** 选择"应用到全部店铺" **THEN** 系统 **SHALL** 对所有门店生效（包括未来新增的门店）
3. **WHEN** 选择特定门店 **THEN** 系统 **SHALL** 只对这些门店生效
4. **WHEN** 查询限购配置 **THEN** 系统 **SHALL** 只返回启用状态且适用于当前门店的配置

#### 具体要求

- [x] 3.1 新增 `ttpos_purchase_quota_config` 限购配置主表
- [x] 3.2 新增 `ttpos_purchase_quota_config_shop` 门店关联表
- [x] 3.3 支持物品级别的限购配置（material_uuid + unit_uuid）
- [x] 3.4 支持"应用到全部店铺"选项（apply_to_all_shops 字段）
- [x] 3.5 支持多门店选择（通过关联表存储门店列表）
- [x] 3.6 预留扩展字段：周期类型、超限策略、配置来源
- [x] 3.7 使用雪花算法生成业务 UUID
- [x] 3.8 查询时支持 JOIN 关联表过滤门店

---

### Requirement 4: 提交时限购校验

**用户故事**: 作为采购员，当我提交品牌采购申请时，系统自动校验是否超过当月申请限额。

#### 验收标准

1. **WHEN** 子店提交品牌采购申请 **AND** 物品有限购配置 **AND** 本次采购数量 > 限额 **THEN** 系统 **SHALL** 拒绝提交并提示超限信息
2. **WHEN** 子店提交品牌采购申请 **AND** 物品有限购配置 **AND** 采购单位与配置单位不一致 **THEN** 系统 **SHALL** 拒绝提交并提示使用正确单位
3. **IF** 物品无限购配置 **THEN** 系统 **SHALL** 允许正常提交，不做限额校验

#### 具体要求

- [x] 4.1 在 `SubmitPurchaseOrder` 方法中增加限购校验逻辑
- [x] 4.2 仅对品牌采购（PurchaseType=2）进行校验
- [x] 4.3 限购为每次的申请单
- [x] 4.4 错误信息支持国际化

#### 设计说明：实时查询统计模式

**为什么"驳回后无需额外处理"？**

采用**实时查询统计**而非**预扣模式**：

| 方案         | 提交时                     | 驳回时           | 优点         | 缺点             |
| ------------ | -------------------------- | ---------------- | ------------ | ---------------- |
| 预扣模式     | 写入 `used_quota` 字段     | 需要回滚字段     | 查询快       | 需维护额外状态   |
| **实时查询** ✅ | 实时统计符合条件的订单 | 自动排除（无需处理） | 无状态、简单 | 每次需要查询统计 |

**统计 SQL 逻辑**：

```sql
SELECT COALESCE(SUM(poi.num), 0) as used_qty
FROM ttpos_purchase_order po
JOIN ttpos_purchase_order_item poi ON po.uuid = poi.purchase_order_uuid
WHERE po.purchase_type = 2                    -- 品牌采购
  AND poi.material_uuid = ?                   -- 物品UUID
  AND poi.unit_uuid = ?                       -- 限购单位UUID
  AND po.status IN (1, 2, 4, 5)               -- ⚠️ 关键：不含已驳回
  AND po.uuid != ?                            -- 排除当前单据
  AND FROM_UNIXTIME(po.create_time, '%Y-%m') = ?  -- 当前月份
  AND po.delete_time = 0
```

**订单状态说明**：

| 状态码 | 状态名     | 是否计入限额 |
| ------ | ---------- | ------------ |
| 0      | 草稿       | ❌ 不计入     |
| 1      | 待审核     | ✅ 计入       |
| 2      | 已通过     | ✅ 计入       |
| 3      | 已驳回     | ❌ 不计入     |
| 4      | 部分收货   | ✅ 计入       |
| 5      | 待总部审核 | ✅ 计入       |
| 6      | 全部收货   | ❌ 不计入     |

**驳回后的流程**：

```
订单被驳回 → status 变为 3
  ↓
下次限额校验 → 查询 status IN (1,2,4,5)
  ↓
已驳回订单不在查询范围 → 限额自动"释放"
```

---

### Requirement 5: 采购单位约束

**用户故事**: 作为采购员，当物品有限购配置时，我只能使用配置的单位进行采购。

#### 验收标准

1. **WHEN** 物品有限购配置 **AND** 采购单位与配置单位不一致 **THEN** 系统 **SHALL** 拒绝提交
2. **WHEN** 物品有限购配置 **AND** 采购单位与配置单位一致 **THEN** 系统 **SHALL** 继续执行限额校验

#### 具体要求

- [x] 5.1 校验采购明细中的单位 UUID 与配置单位 UUID 是否一致
- [x] 5.2 不一致时返回明确的错误提示，包含正确单位名称

---

### Requirement 6: 全局配置管理

**用户故事**: 作为系统管理员，我想配置全局的申请次数上限和单次数量上限，以便于统一管理采购限额规则。

#### 验收标准

1. **WHEN** 系统启动时 **THEN** 系统 **SHALL** 加载全局配置（申请次数上限、单次数量上限）
2. **IF** 配置不存在 **THEN** 系统 **SHALL** 使用默认值（每天2次，单次100件）
3. **WHEN** 配置更新 **THEN** 系统 **SHALL** 立即生效（无需重启）

#### 具体要求

- [x] 6.1 添加全局配置项：`purchase.brand.daily_limit`（每天申请次数上限，默认2）
- [x] 6.2 添加全局配置项：`purchase.brand.single_qty_limit`（单次数量上限，默认100）
- [x] 6.3 支持动态读取配置（不需要重启服务）

---

### Requirement 7: 门店配置界面

**用户故事**: 作为管理员，我想在门店管理页面配置每日品类申请数限制，以便于灵活管理不同门店的采购规则。

#### 验收标准

1. **WHEN** 点击门店管理列表中的"门店配置"按钮 **THEN** 系统 **SHALL** 打开门店配置页面
2. **WHEN** 在配置页面输入"每日品类申请数"并保存 **THEN** 系统 **SHALL** 更新该门店的配置
3. **IF** 门店未配置 **THEN** 系统 **SHALL** 使用全局默认值

#### 具体要求

- [x] 7.1 在门店列表增加"门店配置"入口
- [x] 7.2 门店配置页面包含"每日采购申请次数"输入框
- [x] 7.3 支持保存和重置操作
- [x] 7.4 显示当前配置值或默认值

#### 字段说明

- **每日品类申请数**：限制门店每天提交品牌采购申请单的次数（如配置为2次，门店今天最多提交2个品牌采购申请单）
- **存储字段**：`purchase_daily_limit`（数据库字段保持不变）

---

### Requirement 8: 物品限购配置界面

**用户故事**: 作为采购管理员，我想在物品详情页面配置申请限额并选择应用的门店，以便于精细化管理采购限制。

#### 验收标准

1. **WHEN** 在物品详情页面点击"申请限额设置" **THEN** 系统 **SHALL** 打开限额配置弹窗
2. **WHEN** 选择"应用到全部店铺" **THEN** 系统 **SHALL** 对所有门店生效
3. **WHEN** 选择特定门店 **THEN** 系统 **SHALL** 显示门店选择器，支持多选
4. **WHEN** 输入限购数量并确定 **THEN** 系统 **SHALL** 保存配置并返回
5. **WHEN** 物品已有限购配置 **THEN** 系统 **SHALL** 显示现有配置信息

#### 具体要求

- [x] 8.1 在物品详情页面增加"申请限额设置"入口
- [x] 8.2 配置弹窗包含：应用门店选择、限购数量输入
- [x] 8.3 支持"应用到全部店铺"选项（默认选中）
- [x] 8.4 门店选择器支持搜索和多选
- [x] 8.5 限购数量支持数字键盘输入（0.01-999999）
- [x] 8.6 保存时校验必填项和数据格式
- [x] 8.7 显示已选门店列表和数量统计

---

### Requirement 9: 限购配置管理 API

**用户故事**: 作为前端开发者，我需要API来支持限购配置的增删改查操作。

#### 验收标准

1. **WHEN** 调用创建限购配置API **THEN** 系统 **SHALL** 保存配置并返回成功
2. **WHEN** 调用查询限购配置API **THEN** 系统 **SHALL** 返回物品的限购配置详情
3. **WHEN** 调用更新限购配置API **THEN** 系统 **SHALL** 更新配置并返回成功
4. **WHEN** 调用删除限购配置API **THEN** 系统 **SHALL** 软删除配置

#### 具体要求

- [x] 9.1 新增 POST `/api/v1/shop/purchase/quota/config` - 创建/更新限购配置
- [x] 9.2 新增 GET `/api/v1/shop/purchase/quota/config/{material_uuid}` - 查询物品限购配置
- [x] 9.3 新增 DELETE `/api/v1/shop/purchase/quota/config/{uuid}` - 删除限购配置
- [x] 9.4 所有API支持国际化错误提示

---

### Requirement 10: 门店配置管理 API

**用户故事**: 作为前端开发者，我需要API来支持门店级配置的读写操作。

#### 验收标准

1. **WHEN** 调用获取门店配置API **THEN** 系统 **SHALL** 返回门店的配置项
2. **WHEN** 调用更新门店配置API **THEN** 系统 **SHALL** 保存配置并返回成功

#### 具体要求

- [x] 10.1 新增 GET `/api/v1/shop/config/{shop_uuid}` - 获取门店配置
- [x] 10.2 新增 POST `/api/v1/shop/config/{shop_uuid}` - 更新门店配置
- [x] 10.3 配置项包含：purchase_daily_limit（每日采购申请次数）

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 Controller → Service → Repository 分层
- **单一职责原则**: 限购校验逻辑集成在 PurchaseOrderService 中
- **模块化设计**: Repository 独立，可复用
- **依赖管理**: Service 只能依赖其他 Service 接口
- **遵循规范**:
  - `.cursor/rules/go-main.mdc` - Go Main 开发规范
  - `.cursor/rules/database.mdc` - 数据库开发规范
  - `.cursor/rules/structs.mdc` - 项目结构规范

### API 设计要求

- [x] MVP 阶段无新增 API（校验逻辑在现有提交接口中执行）
- [x] 错误响应格式：`{code, message, data{}}`

### 数据库设计要求

- [x] 必须包含: `id`, `uuid`, `create_time`, `update_time`, `delete_time`
- [x] 时间字段使用 int 类型，\_time 结尾，默认值 0
- [x] UUID 字段使用 bigint unsigned
- [x] 表名使用 ttpos\_ 前缀
- [x] 字段名使用 snake_case
- [x] 金额字段使用 decimal(10,2)（限额数量）

### 性能要求

- [x] 限购校验响应时间 < 100ms
- [x] 使用索引优化物品限购配置查询
- [x] 使用分布式锁防止并发超限

### 测试要求

- [ ] Service 层测试覆盖率 ≥ 70%
- [ ] Repository 层测试覆盖率 ≥ 80%
- [ ] 集成测试覆盖限购校验核心流程
- [ ] 并发场景测试

### 国际化要求

- [x] 超限错误提示使用多语言实现
- [x] 单位不匹配错误提示使用多语言实现

#### 错误文案清单

| Key | 中文 (zh) | 英文 (en) |
| ---- | ---- | ---- |
| `purchase.daily_limit_exceeded` | 今日申请次数已达上限（%d次），请明天再试 | Daily request limit reached (%d times), please try again tomorrow |
| `purchase.single_qty_exceeded` | 本次申请物品数量已超限（最多%d件），请减少物品数量后重试 | Single request quantity exceeded (max %d items), please reduce quantity |
| `purchase.quota.exceeded` | 品牌采购物品[%s]本月限购%v%s，已使用%v，本次申请%v，超出限额 | Brand purchase item [%s] monthly quota is %v%s, used %v, this request %v, exceeds limit |
| `purchase.quota.unit_mismatch` | 品牌采购物品[%s]限购单位为[%s]，请使用指定单位 | Brand purchase item [%s] quota unit is [%s], please use the specified unit |
| `purchase.quota.config_not_found` | 获取品牌采购限购配置失败 | Failed to get brand purchase quota config |
| `purchase.quota.used_query_failed` | 查询品牌采购已使用额度失败 | Failed to query brand purchase used quota |

**其他语言文案**（按现有 i18n 目录结构添加）：

| 语言 | 文件 | 每日次数超限 | 单次数量超限 | 物品限额超限 | 单位不匹配 |
| ---- | ---- | ---- | ---- | ---- | ---- |
| 日语 (ja) | `ja.json` | 本日の申請回数が上限（%d回）に達しました。明日再試行してください | 今回の申請商品数量が上限（最大%d個）を超えました。数量を減らしてください | ブランド仕入れ商品[%s]の今月の購入制限は%v%s、使用済み%v、今回の申請%v、制限超過 | ブランド仕入れ商品[%s]の制限単位は[%s]です、指定された単位を使用してください |
| 韩语 (ko) | `ko.json` | 오늘 신청 횟수가 한도(%d회)에 도달했습니다. 내일 다시 시도하세요 | 이번 신청 품목 수량이 한도(최대%d개)를 초과했습니다. 수량을 줄여주세요 | 브랜드 구매 품목 [%s] 월 한도 %v%s, 사용량 %v, 이번 신청 %v, 한도 초과 | 브랜드 구매 품목 [%s] 한도 단위는 [%s]입니다, 지정된 단위를 사용하세요 |
| 泰语 (th) | `th.json` | จำนวนการร้องขอวันนี้ถึงขีดจำกัด(%d ครั้ง) โปรดลองอีกครั้งในวันพรุ่งนี้ | จำนวนสินค้าในการร้องขอนี้เกินขีดจำกัด(สูงสุด%d รายการ) โปรดลดจำนวน | สินค้าสั่งซื้อแบรนด์ [%s] จำกัดต่อเดือน %v%s, ใช้แล้ว %v, คำขอนี้ %v, เกินโควตา | สินค้าสั่งซื้อแบรนด์ [%s] หน่วยจำกัดคือ [%s] กรุณาใช้หน่วยที่กำหนด |
| 德语 (de) | `de.json` | Tägliche Antragslimit erreicht (%d Mal), bitte versuchen Sie es morgen erneut | Menge dieser Anfrage überschreitet Limit (max %d Stück), bitte reduzieren | Markenartikel [%s] monatliches Limit %v%s, verwendet %v, diese Anfrage %v, Limit überschritten | Markenartikel [%s] Limiteinheit ist [%s], bitte verwenden Sie die angegebene Einheit |
| 瑞典语 (sv) | `sv.json` | Dagens ansökningsgräns uppnådd (%d gånger), försök igen imorgon | Denna begärans kvantitet överskrider gränsen (max %d artiklar), vänligen minska | Varumärkesköp artikel [%s] månatlig kvot %v%s, använd %v, denna begäran %v, överskrider gräns | Varumärkesköp artikel [%s] kvotenheten är [%s], använd den angivna enheten |
| 土耳其语 (tr) | `tr.json` | Günlük başvuru limiti ulaşıldı (%d kez), lütfen yarın tekrar deneyin | Bu talebin miktarı limiti aşıyor (maks %d adet), lütfen azaltın | Marka satın alma ürünü [%s] aylık kota %v%s, kullanılan %v, bu talep %v, limit aşıldı | Marka satın alma ürünü [%s] kota birimi [%s], lütfen belirtilen birimi kullanın |
| 缅甸语 (my) | `my.json` | ယနေ့ လျှောက်ထားမှု အကန့်အသတ် ရောက်ပြီ (%d ကြိမ်), မနက်ဖြန် ထပ်လျှောက်ပါ | ဤတောင်းဆိုမှု ပမာဏ ကန့်သတ်ချက် ကျော်လွန် (အများဆုံး %d ခု), ကျေးဇူးပြု၍ လျှော့ပါ | Brand ဝယ်ယူမှု ပစ္စည်း [%s] လစဉ် ခွင့်ပြုချက် %v%s, သုံးပြီး %v, ဤတောင်းဆိုမှု %v, ကန့်သတ်ချက် ကျော်လွန် | Brand ဝယ်ယူမှု ပစ္စည်း [%s] ကန့်သတ်ယူနစ်မှာ [%s] ဖြစ်သည်, သတ်မှတ်ထားသော ယူနစ်ကို အသုံးပြုပါ |
| 繁体中文 (zhtw) | `zhtw.json` | 今日申請次數已達上限（%d次），請明天再試 | 本次申請物品數量已超限（最多%d件），請減少物品數量後重試 | 品牌採購物品[%s]本月限購%v%s，已使用%v，本次申請%v，超出限額 | 品牌採購物品[%s]限購單位為[%s]，請使用指定單位 |

### 安全要求

- [x] 所有 API 需要身份验证（复用现有认证）
- [x] SQL 注入防护（使用参数化查询）

### 可靠性要求

- [x] 事务管理（保证数据一致性）
- [x] 错误日志记录

---

## 验收标准

### 功能验收

1. **限购配置**: 配置表可正确存储物品限购规则
2. **提交校验**: 超限时拒绝提交，返回明确错误信息
3. **单位约束**: 单位不匹配时拒绝提交
4. **无配置放行**: 无限购配置的物品正常提交

### 测试验收

1. **单元测试**: 覆盖率达标
2. **集成测试**: 端到端流程测试通过
3. **并发测试**: 多人同时提交不超限

### 文档验收

1. **技术文档**: design.md 完整且准确
2. **数据库文档**: 迁移脚本和表结构文档完整
3. **测试文档**: tasks.md 中的测试任务完成

---

## 约束条件

### 技术约束

#### Go Main 模块

- 必须使用 Gin 框架
- 接口以 `I` 开头，实现以 `Impl` 结尾
- Service 只能依赖其他 Service 接口
- Repository 只能持有 db 实例，不能持有 DBManager
- 不使用 panic，返回 error

### 业务约束

- MVP 阶段限购配置通过数据库直接维护
- 限购周期仅支持月度
- 超限策略仅支持严格拒绝

### 资源约束

- 开发时间: 1-2 天（纯后端）
- Story Point: 2（待技术评审确认）

---

## 依赖关系

### 技术依赖

- `main/app/service/purchase_order/purchase_order.go` - 品牌采购服务
- `main/app/model/purchase_order.go` - 采购订单模型
- `main/pkg/utils/snowflake.go` - 雪花算法工具

### 服务依赖

- 无外部服务依赖

### 业务依赖

- 品牌采购流程已存在且稳定运行

---

## 风险和缓解

### 风险 1: 限购配置暂无前端管理页面

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 数据模型预留所有扩展字段
- 后续迭代增加前端配置页面
- MVP 阶段提供 SQL 脚本示例供运营使用

### 风险 2: 并发提交导致超限

**影响**: 中  
**概率**: 低  
**缓解措施**:

- 使用分布式锁（对采购单 UUID 加锁）
- 查询统计时排除当前单据
- 原子性事务保证

---

## 时间表

- **Phase 1 - 产品审核**: 0.5 天
- **Phase 2 - 技术设计**: 0.5 天
- **Phase 3 - 开发实现**: 1.0 天
- **总计**: 2.0 天（SP = 2）

---

## 参考资料

### 核心规范

- `.cursor/rules/go-main.mdc` - Go Main 核心约束
- `.cursor/rules/api.mdc` - API 设计规范
- `.cursor/rules/database.mdc` - 数据库开发规范
- `.cursor/rules/security.mdc` - 安全开发规范

### 架构文档

- `docs/human/architecture/go-main-architecture.md` - Go Main 架构
- `docs/human/architecture/database-design.md` - 数据库设计

### 开发指南

- `docs/human/guides/go-main-development.md` - Go Main 开发指南
- `docs/human/guides/database-guide.md` - 数据库开发指南

### 外部参考

- 现有品牌采购流程：`main/app/service/purchase_order/purchase_order.go`

---

## 数据模型设计

### 表结构

#### 表1: 品牌采购限购配置表

```sql
-- 品牌采购限购配置表（主表）
CREATE TABLE ttpos_purchase_quota_config (
  id INT(11) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '自增ID',
  uuid BIGINT(20) UNSIGNED NOT NULL DEFAULT 0 COMMENT '绑定记录ID（雪花算法生成）',
  material_uuid BIGINT(20) UNSIGNED NOT NULL COMMENT '物品UUID',
  material_code VARCHAR(100) NOT NULL DEFAULT '' COMMENT '物品编码（冗余，便于查询）',
  unit_uuid BIGINT(20) UNSIGNED NOT NULL COMMENT '限购单位UUID',
  unit_name VARCHAR(50) NOT NULL DEFAULT '' COMMENT '限购单位名称（冗余）',
  quota_limit DECIMAL(10,2) NOT NULL DEFAULT 0.00 COMMENT '限购数量',
  
  -- 门店范围控制
  apply_to_all_shops TINYINT(4) NOT NULL DEFAULT 1 COMMENT '是否应用到全部店铺: 1=是 0=否',
  
  -- 扩展字段（预留）
  period_type TINYINT(4) NOT NULL DEFAULT 0 COMMENT '周期类型: 0=按天(默认) 1=月度',
  strict_mode TINYINT(4) NOT NULL DEFAULT 1 COMMENT '超限策略: 1=严格拒绝',
  config_source TINYINT(4) NOT NULL DEFAULT 1 COMMENT '配置来源: 1=门店 2=总部',
  
  -- 状态字段
  status TINYINT(4) NOT NULL DEFAULT 1 COMMENT '状态: 1=启用 0=禁用',
  create_time INT(10) NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
  update_time INT(10) NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
  delete_time INT(10) NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
  
  PRIMARY KEY (id),
  UNIQUE KEY uk_uuid (uuid),
  KEY idx_material (material_uuid),
  KEY idx_status (status),
  KEY idx_delete_time (delete_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='品牌采购限购配置';
```

#### 表2: 限购配置门店关联表

```sql
-- 品牌采购限购配置门店关联表
CREATE TABLE ttpos_purchase_quota_config_shop (
  id INT(11) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '自增ID',
  config_uuid BIGINT(20) UNSIGNED NOT NULL COMMENT '限购配置UUID',
  shop_uuid BIGINT(20) UNSIGNED NOT NULL COMMENT '门店UUID',
  
  -- 状态字段
  create_time INT(10) NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
  delete_time INT(10) NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
  
  PRIMARY KEY (id),
  UNIQUE KEY uk_config_shop (config_uuid, shop_uuid),
  KEY idx_config (config_uuid),
  KEY idx_shop (shop_uuid),
  KEY idx_delete_time (delete_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='品牌采购限购配置门店关联';
```

**设计说明**：

- **apply_to_all_shops=1**：应用到全部店铺，关联表无记录
- **apply_to_all_shops=0**：应用到指定店铺，关联表存储门店列表
- **查询逻辑**：先查主表，若 `apply_to_all_shops=0` 则 JOIN 关联表过滤门店
- **数据一致性**：删除配置时需同步删除关联表记录（级联或手动）

### 状态常量

```go
// main/app/constant/purchase_quota.go

package constant

// 限购配置状态
const (
	PurchaseQuotaConfigStatusDisabled = 0 // 禁用
	PurchaseQuotaConfigStatusEnabled  = 1 // 启用
)

// 限购周期类型
const (
	PurchaseQuotaPeriodTypeDaily     = 0 // 按天（默认）
	PurchaseQuotaPeriodTypeMonthly   = 1 // 月度
	PurchaseQuotaPeriodTypeQuarterly = 2 // 季度（预留）
	PurchaseQuotaPeriodTypeYearly    = 3 // 年度（预留）
)

// 限购超限策略
const (
	PurchaseQuotaStrictModeStrict = 1 // 严格拒绝
	PurchaseQuotaStrictModeSoft   = 2 // 柔性提醒（预留）
)

// 限购配置来源
const (
	PurchaseQuotaConfigSourceShop         = 1 // 门店自配
	PurchaseQuotaConfigSourceHeadquarter  = 2 // 总部下发
)
```

---

## 暂缓实现清单（MVP 后续迭代）

| 功能             | 说明                           | 预留字段/设计                              | 优先级 |
| ---------------- | ------------------------------ | ------------------------------------------ | ------ |
| **前端配置页面** | Shop 端限购配置 CRUD 界面      | Model/Repository 已就绪                    | P1     |
| **总部统一配置** | 总部配置规则，下发到所有子店   | `config_source` 字段 | P1     |
| **已用限额查询** | 返回物品本月已用/剩余额度      | 实时查询统计实现                           | P1     |
| **柔性超限策略** | 超限时允许提交但标记           | `strict_mode` 字段                         | P2     |
| **多周期支持**   | 支持季度、年度限购周期         | `period_type` 字段（默认按天） | P2     |

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`

---

**版本**: v1.0.0-MVP  
**创建日期**: 2026-01-07  
**作者**: BenDaye  
**审核者**: 待指定

