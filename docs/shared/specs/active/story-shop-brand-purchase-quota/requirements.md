# 品牌采购月度限购 需求文档

> 本文档定义品牌采购月度限购功能的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                              |
| ----------------- | ----------------------------------------------------------------------------------------------------------------- |
| **来源 Proposal** | [docs/team/proposals/2025-12/brand-purchase-quota.md](../../../../team/proposals/2025-12/brand-purchase-quota.md) |
| **创建日期**      | 2025-12-25                                                                                                        |
| **负责人**        | BenDaye                                                                                                           |
| **目标 Sprint**   | 待定                                                                                                              |
| **涉及技术栈**    | [x] Go (main/) [ ] Go (ttpos-bmp/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                        |
| **版本类型**      | MVP（最小可行产品）                                                                                               |

## 📋 审核状态

| 项目         | 内容   |
| ------------ | ------ |
| **审核状态** | 待审核 |
| **审核人**   | 待指定 |
| **审核日期** | -      |
| **审核意见** | -      |

---

## 📋 概述

在品牌连锁经营模式下，子店向总部发起品牌采购时，某些物品存在过度采购的情况。本功能在子店提交品牌采购申请时，自动校验物品本月已采购数量是否超过限额，超出则拒绝提交。

**MVP 范围**：

- 核心限购校验逻辑（后端）
- 限购配置数据模型
- 暂不包含前端配置页面（通过数据库直接维护）

## 🎯 产品对齐

- 控制子店采购频次和数量，避免过度采购
- 优化总部库存周转，减少积压
- 建立规范的采购管控机制
- 为未来精细化运营打下基础

## 📝 用户故事

**作为** 门店采购员  
**我想** 在发起品牌采购时受到月度限额约束  
**以便于** 避免过度采购，规范采购行为

---

## 功能需求

### Requirement 1: 限购配置数据模型

**用户故事**: 作为系统管理员，我想配置物品的月度采购限额，以便于系统进行校验。

#### 验收标准

1. **WHEN** 新增限购配置记录 **THEN** 系统 **SHALL** 存储物品、单位、限额等信息
2. **IF** 同一物品已存在限购配置 **THEN** 系统 **SHALL** 拒绝重复创建（唯一约束）
3. **WHEN** 查询限购配置 **THEN** 系统 **SHALL** 只返回启用状态的配置

#### 具体要求

- [x] 1.1 新增 `ttpos_purchase_quota_config` 限购配置表
- [x] 1.2 支持物品级别的限购配置（material_uuid + unit_uuid）
- [x] 1.3 预留扩展字段：周期类型、超限策略、配置来源、总部配置 UUID
- [x] 1.4 使用雪花算法生成业务 UUID

---

### Requirement 2: 提交时限购校验

**用户故事**: 作为采购员，当我提交品牌采购申请时，系统自动校验是否超过月度限额。

#### 验收标准

1. **WHEN** 子店提交品牌采购申请 **AND** 物品有限购配置 **AND** 本月已采购数量+本次数量 > 限额 **THEN** 系统 **SHALL** 拒绝提交并提示超限信息
2. **WHEN** 子店提交品牌采购申请 **AND** 物品有限购配置 **AND** 采购单位与配置单位不一致 **THEN** 系统 **SHALL** 拒绝提交并提示使用正确单位
3. **WHEN** 品牌采购申请被驳回 **THEN** 系统 **SHALL** 自动释放已占用的限额（通过查询统计实现，无需额外处理）
4. **IF** 物品无限购配置 **THEN** 系统 **SHALL** 允许正常提交，不做限额校验

#### 具体要求

- [x] 2.1 在 `SubmitPurchaseOrder` 方法中增加限购校验逻辑
- [x] 2.2 仅对品牌采购（PurchaseType=2）进行校验
- [x] 2.3 统计口径：待审核(1)+已通过(2)+待总部审核(5)+部分收货(4)（排除草稿、已驳回、全部收货）
- [x] 2.4 限购周期为自然月（每月 1 号重置）
- [x] 2.5 错误信息支持国际化

#### 设计说明：实时查询统计模式

**为什么"驳回后无需额外处理"？**

采用**实时查询统计**而非**预扣模式**：

| 方案            | 提交时                 | 驳回时               | 优点         | 缺点             |
| --------------- | ---------------------- | -------------------- | ------------ | ---------------- |
| 预扣模式        | 写入 `used_quota` 字段 | 需要回滚字段         | 查询快       | 需维护额外状态   |
| **实时查询** ✅ | 实时统计符合条件的订单 | 自动排除（无需处理） | 无状态、简单 | 每次需要查询统计 |

**统计 SQL 逻辑**：

```sql
SELECT COALESCE(SUM(poi.num), 0) as used_qty
FROM ttpos_purchase_order po
JOIN ttpos_purchase_order_item poi ON po.uuid = poi.purchase_order_uuid
WHERE po.purchase_type = 2                          -- 品牌采购
  AND poi.material_uuid = ?                         -- 物品UUID
  AND poi.unit_uuid = ?                             -- 限购单位UUID
  AND po.status IN (1, 2, 4, 5)                     -- ⚠️ 关键：不含已驳回
  AND po.uuid != ?                                  -- 排除当前单据
  AND FROM_UNIXTIME(po.create_time, '%Y-%m') = ?    -- 当前月份
  AND po.delete_time = 0
```

**订单状态说明**：

| 状态码 | 状态名     | 是否计入限额 |
| ------ | ---------- | ------------ |
| 0      | 草稿       | ❌ 不计入    |
| 1      | 待审核     | ✅ 计入      |
| 2      | 已通过     | ✅ 计入      |
| 3      | 已驳回     | ❌ 不计入    |
| 4      | 部分收货   | ✅ 计入      |
| 5      | 待总部审核 | ✅ 计入      |
| 6      | 全部收货   | ❌ 不计入    |

**驳回后的流程**：

```
订单被驳回 → status 变为 3
     ↓
下次限额校验 → 查询 status IN (1,2,4,5)
     ↓
已驳回订单不在查询范围 → 限额自动"释放"
```

---

### Requirement 3: 采购单位约束

**用户故事**: 作为采购员，当物品有限购配置时，我只能使用配置的单位进行采购。

#### 验收标准

1. **WHEN** 物品有限购配置 **AND** 采购单位与配置单位不一致 **THEN** 系统 **SHALL** 拒绝提交
2. **WHEN** 物品有限购配置 **AND** 采购单位与配置单位一致 **THEN** 系统 **SHALL** 继续执行限额校验

#### 具体要求

- [x] 3.1 校验采购明细中的单位 UUID 与配置单位 UUID 是否一致
- [x] 3.2 不一致时返回明确的错误提示，包含正确单位名称

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

| Key                                                            | 中文 (zh)                                                    | 英文 (en)                                                                               |
| -------------------------------------------------------------- | ------------------------------------------------------------ | --------------------------------------------------------------------------------------- |
| `品牌采购物品[%s]本月限购%v%s，已使用%v，本次申请%v，超出限额` | 品牌采购物品[%s]本月限购%v%s，已使用%v，本次申请%v，超出限额 | Brand purchase item [%s] monthly quota is %v%s, used %v, this request %v, exceeds limit |
| `品牌采购物品[%s]限购单位为[%s]，请使用指定单位`               | 品牌采购物品[%s]限购单位为[%s]，请使用指定单位               | Brand purchase item [%s] quota unit is [%s], please use the specified unit              |
| `获取品牌采购限购配置失败`                                     | 获取品牌采购限购配置失败                                     | Failed to get brand purchase quota config                                               |
| `查询品牌采购已使用额度失败`                                   | 查询品牌采购已使用额度失败                                   | Failed to query brand purchase used quota                                               |

**其他语言文案**（按现有 i18n 目录结构添加）：

| 语言            | 文件        | 超限错误                                                                                                 | 单位不匹配错误                                                                              |
| --------------- | ----------- | -------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------- |
| 日语 (ja)       | `ja.json`   | ブランド仕入れ商品[%s]の今月の購入制限は%v%s、使用済み%v、今回の申請%v、制限超過                         | ブランド仕入れ商品[%s]の制限単位は[%s]です、指定された単位を使用してください                |
| 韩语 (ko)       | `ko.json`   | 브랜드 구매 품목 [%s] 월 한도 %v%s, 사용량 %v, 이번 신청 %v, 한도 초과                                   | 브랜드 구매 품목 [%s] 한도 단위는 [%s]입니다, 지정된 단위를 사용하세요                      |
| 泰语 (th)       | `th.json`   | สินค้าสั่งซื้อแบรนด์ [%s] จำกัดต่อเดือน %v%s, ใช้แล้ว %v, คำขอนี้ %v, เกินโควตา                          | สินค้าสั่งซื้อแบรนด์ [%s] หน่วยจำกัดคือ [%s] กรุณาใช้หน่วยที่กำหนด                          |
| 德语 (de)       | `de.json`   | Markenartikel [%s] monatliches Limit %v%s, verwendet %v, diese Anfrage %v, Limit überschritten           | Markenartikel [%s] Limiteinheit ist [%s], bitte verwenden Sie die angegebene Einheit        |
| 瑞典语 (sv)     | `sv.json`   | Varumärkesköp artikel [%s] månatlig kvot %v%s, använd %v, denna begäran %v, överskrider gräns            | Varumärkesköp artikel [%s] kvotenheten är [%s], använd den angivna enheten                  |
| 土耳其语 (tr)   | `tr.json`   | Marka satın alma ürünü [%s] aylık kota %v%s, kullanılan %v, bu talep %v, limit aşıldı                    | Marka satın alma ürünü [%s] kota birimi [%s], lütfen belirtilen birimi kullanın             |
| 缅甸语 (my)     | `my.json`   | Brand ဝယ်ယူမှု ပစ္စည်း [%s] လစဉ် ခွင့်ပြုချက် %v%s, သုံးပြီး %v, ဤတောင်းဆိုမှု %v, ကန့်သတ်ချက် ကျော်လွန် | Brand ဝယ်ယူမှု ပစ္စည်း [%s] ကန့်သတ်ယူနစ်မှာ [%s] ဖြစ်သည်, သတ်မှတ်ထားသော ယူနစ်ကို အသုံးပြုပါ |
| 繁体中文 (zhtw) | `zhtw.json` | 品牌採購物品[%s]本月限購%v%s，已使用%v，本次申請%v，超出限額                                             | 品牌採購物品[%s]限購單位為[%s]，請使用指定單位                                              |

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

```sql
-- 品牌采购限购配置表（门店级）
CREATE TABLE ttpos_purchase_quota_config (
    id                INT(11) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '自增ID',
    uuid              BIGINT(20) UNSIGNED NOT NULL DEFAULT 0 COMMENT '绑定记录ID（雪花算法生成）',
    material_uuid     BIGINT(20) UNSIGNED NOT NULL COMMENT '物品UUID',
    material_code     VARCHAR(100) NOT NULL COMMENT '物品编码（冗余，便于查询）',
    unit_uuid         BIGINT(20) UNSIGNED NOT NULL COMMENT '限购单位UUID',
    unit_name         VARCHAR(50) NOT NULL COMMENT '限购单位名称（冗余）',
    quota_limit       DECIMAL(10,2) NOT NULL COMMENT '限购数量',

    -- 扩展字段（预留）
    period_type       TINYINT(4) NOT NULL DEFAULT 1 COMMENT '周期类型: 1=月度',
    strict_mode       TINYINT(4) NOT NULL DEFAULT 1 COMMENT '超限策略: 1=严格拒绝',
    config_source     TINYINT(4) NOT NULL DEFAULT 1 COMMENT '配置来源: 1=门店 2=总部',
    headquarter_config_uuid BIGINT(20) UNSIGNED DEFAULT NULL COMMENT '总部配置UUID',

    -- 状态字段
    status            TINYINT(4) NOT NULL DEFAULT 1 COMMENT '状态: 1=启用 0=禁用',
    create_time       INT(10) NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    update_time       INT(10) NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    delete_time       INT(10) NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',

    PRIMARY KEY (id),
    UNIQUE KEY uk_material (material_uuid),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='品牌采购限购配置';
```

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
    PurchaseQuotaConfigSourceShop        = 1 // 门店自配
    PurchaseQuotaConfigSourceHeadquarter = 2 // 总部下发
)
```

---

## 暂缓实现清单（MVP 后续迭代）

| 功能             | 说明                         | 预留字段/设计                              | 优先级 |
| ---------------- | ---------------------------- | ------------------------------------------ | ------ |
| **前端配置页面** | Shop 端限购配置 CRUD 界面    | Model/Repository 已就绪                    | P1     |
| **总部统一配置** | 总部配置规则，下发到所有子店 | `config_source`、`headquarter_config_uuid` | P1     |
| **已用限额查询** | 返回物品本月已用/剩余额度    | 实时查询统计实现                           | P1     |
| **柔性超限策略** | 超限时允许提交但标记         | `strict_mode` 字段                         | P2     |
| **多周期支持**   | 支持季度、年度限购周期       | `period_type` 字段                         | P2     |

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`

---

**版本**: v1.0.0-MVP  
**创建日期**: 2025-12-25  
**作者**: BenDaye  
**审核者**: 待指定
