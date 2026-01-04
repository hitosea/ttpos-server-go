# 品牌采购月度限购 需求提案

> 本文档用于需求发起阶段，经团队评审后创建正式 Spec。

---

## ⚠️ 版本说明

**本提案为 MVP（最小可行产品）版本**，目标是快速上线核心限购校验能力。

为加速交付，以下功能暂缓实现，在数据模型中预留扩展字段：

- 前端配置管理页面（暂通过数据库直接维护）
- 已用限额查询接口（前端暂不展示剩余额度）
- 总部统一配置下发机制
- 柔性超限策略（允许超限但标记）
- 多周期支持（季度、年度）
- 物品分类级别限购

详见 [暂缓实现清单](#-暂缓实现清单mvp-后续迭代)。

---

## 📋 提案信息

| 项目          | 内容                                                                                             |
| ------------- | ------------------------------------------------------------------------------------------------ |
| **提案人**    | BenDaye                                                                                          |
| **日期**      | 2025-12-25                                                                                       |
| **目标版本**  | 待定                                                                                             |
| **状态**      | 进行中                                                                                           |
| **版本类型**  | MVP（最小可行产品）                                                                              |
| **关联任务**  | -                                                                                                |
| **关联 Spec** | [story-shop-brand-purchase-quota](../../../shared/specs/archived/v2.12/story-shop-brand-purchase-quota/) |

---

## 🎯 背景和动机

### 问题描述

在品牌连锁经营模式下，子店向总部发起品牌采购（内部采购）时，某些物品（如清洁用品、办公耗材等）存在过度采购的情况，导致：

1. 总部库存压力增大
2. 子店囤积物品造成浪费
3. 无法有效管控子店的采购行为

目前系统缺少对品牌采购的限额管控机制，需要新增月度限购功能。

### 业务价值

- 控制子店采购频次和数量，避免过度采购
- 优化总部库存周转，减少积压
- 建立规范的采购管控机制
- 为未来精细化运营打下基础

### 目标用户

- [x] 商户管理员（配置限购规则）
- [x] 门店店长/采购员（发起采购时受限购约束）
- [ ] 收银员
- [ ] 厨房人员
- [ ] 顾客

---

## 💡 解决方案概述

### 方案描述

在子店发起品牌采购（PurchaseType=2）提交审核时，系统自动校验该物品本月已采购数量是否超过限额。超出限额则拒绝提交，提示用户已超限。

**核心规则**：

- 限购周期为自然月（每月 1 号重置）
- 限购配置精确到物品+单位
- 有限购配置的物品，采购时只能使用配置的单位
- 统计口径：待审核+待总部审核+已通过+部分收货+全部收货（排除草稿、已驳回）

### 核心功能点

1. **新增限购配置表**：存储物品限购规则（物品、单位、限额）
2. **提交时限购校验**：在 `SubmitPurchaseOrder` 中增加校验逻辑
3. **采购单位约束**：有限购配置的物品，强制使用配置单位

### 影响范围

**涉及终端**：

- [ ] POS 收银端
- [x] Shop 商家管理端（品牌采购功能）
- [ ] KDS 厨显端
- [ ] QDS 排号叫号端
- [ ] Assistant 助手端
- [ ] Tablet 平板端
- [ ] Mobile 扫码端
- [ ] Menu 电子菜单端
- [ ] Member 会员端

**涉及模块**（MVP 范围）：

- [ ] UI 组件（MVP 暂不涉及，后续迭代）
- [x] API 接口（提交校验逻辑）
- [x] 数据模型（新增配置表）
- [x] 业务逻辑（限购校验）
- [ ] 第三方集成

---

## 📊 初步评估

### 技术复杂度

- [x] **低**：纯后端逻辑，MVP 无前端改动
- [ ] **中**：需要前后端联调，基础业务逻辑
- [ ] **高**：涉及架构调整、第三方集成、复杂算法

### 工作量预估

- **预计天数**: 1-2 天（MVP 纯后端）
- **预估 SP**: 2（待技术评审确认）

### 风险识别

**潜在风险**：

1. 限购配置暂时通过数据库直接管理，运营成本较高
2. 未来总部统一配置与门店配置的同步机制需要预留

**缓解措施**：

1. 数据模型预留 `config_source`、`headquarter_config_uuid` 字段，支持未来扩展
2. 预留 `period_type`、`strict_mode` 字段，支持未来多周期和柔性策略
3. 后续迭代增加前端配置页面

---

## 🔗 相关资源

### 参考需求

- 现有品牌采购流程：`main/app/service/purchase_order/purchase_order.go`
- 采购订单模型：`main/app/model/purchase_order.go`

### 相关文档

- 品牌采购逻辑分析：本次需求讨论中整理

---

## 🤝 需求评审

### 评审参与人

| 角色       | 姓名 | 签名/日期 |
| ---------- | ---- | --------- |
| 产品经理   |      |           |
| 技术负责人 |      |           |
| 开发代表   |      |           |
| 测试代表   |      |           |

### 评审结论

- [ ] ✅ **批准**：进入技术方案设计阶段
- [ ] 🔄 **修改后批准**：需补充以下内容
- [ ] ❌ **拒绝**：不符合产品规划或优先级

**评审意见**：

```
待评审
```

**下一步行动**：

- [ ] 创建 Spec：`story-shop-brand-purchase-quota`
- [ ] 分配负责人：待定
- [ ] 目标 Sprint：待定

---

## 📝 附录

### User Story（初稿）

**作为** 门店采购员  
**我想** 在发起品牌采购时受到月度限额约束  
**以便于** 避免过度采购，规范采购行为

### AC 验收标准（初稿）

1. **WHEN** 子店提交品牌采购申请 **AND** 物品有限购配置 **AND** 本月已采购数量+本次数量 > 限额 **THEN** 系统 **SHALL** 拒绝提交并提示超限信息
2. **WHEN** 子店提交品牌采购申请 **AND** 物品有限购配置 **AND** 采购单位与配置单位不一致 **THEN** 系统 **SHALL** 拒绝提交并提示使用正确单位
3. **WHEN** 品牌采购申请被驳回 **THEN** 系统 **SHALL** 自动释放已占用的限额（通过查询统计实现，无需额外处理）
4. **IF** 物品无限购配置 **THEN** 系统 **SHALL** 允许正常提交，不做限额校验

### 数据模型设计（初稿）

#### 表结构

```sql
-- 品牌采购限购配置表（门店级）
-- 表名规范：ttpos_ 前缀
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
    status            TINYINT(4) NOT NULL DEFAULT 1 COMMENT '状态: ✅ 已完成 - 已发布 v2.12
    create_time       INT(10) NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    update_time       INT(10) NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    delete_time       INT(10) NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',

    PRIMARY KEY (id),
    UNIQUE KEY uk_material (material_uuid),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='品牌采购限购配置';
```

#### 主键生成策略

- `id`: 自增主键
- `uuid`: 使用 `utils.GetID()` 雪花算法生成，作为业务主键

#### 状态常量定义

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

### 改动点清单（MVP 范围）

| 序号 | 改动点         | 文件/目录                                           | MVP |
| ---- | -------------- | --------------------------------------------------- | --- |
| 1    | 新增限购配置表 | `database/migrations/`                              | ✅  |
| 2    | 新增状态常量   | `main/app/constant/purchase_quota.go`               | ✅  |
| 3    | 新增配置模型   | `main/app/model/purchase_quota_config.go`           | ✅  |
| 4    | 新增配置仓库   | `main/app/repository/purchase_quota_config.go`      | ✅  |
| 5    | 提交时限购校验 | `main/app/service/purchase_order/purchase_order.go` | ✅  |
| 6    | 采购单位约束   | `main/app/service/purchase_order/validator.go`      | ✅  |

### 实现建议

#### 1. 限购校验逻辑位置

**推荐方案**：在 `SubmitPurchaseOrder` 方法中集成校验逻辑

```go
// main/app/service/purchase_order/purchase_order.go
func (s *purchaseOrderSrv) SubmitPurchaseOrder(ctx context.Context, req req.PurchaseOrderSubmitReq) error {
    // ... 现有逻辑 ...

    // 校验供应商状态
    // 校验物品状态

    // ⚡ 新增：限购校验（仅品牌采购）
    if purchaseOrder.IsHeadquarterPurchase() {
        if err := s.checkPurchaseQuota(ctx, tx, purchaseOrder); err != nil {
            return err
        }
    }

    // 删除数量为0的明细
    // 更新状态
    // ...
}
```

**备选方案**：抽取为独立的 `quota` service（如后续功能复杂化）

#### 2. 并发安全考虑

限购校验涉及"读取已用额度 → 判断 → 提交"的流程，存在并发风险。

**解决方案**：复用现有的分布式锁机制

```go
// 在 checkPurchaseQuota 中使用物品级别锁
for _, item := range purchaseOrder.Items {
    config := quotaConfigs[item.MaterialUuid]
    if config == nil {
        continue
    }

    // 加锁：material_uuid + 当前月份
    lockKey := fmt.Sprintf("quota:%d:%s", item.MaterialUuid, currentMonth)
    s.lock.LockString(lockKey)
    defer s.lock.UnlockString(lockKey)

    // 校验逻辑...
}
```

**注意**：由于采购单可能包含多个物品，建议对整个采购单加锁，而非单个物品：

```go
// 更好的方案：对采购单加锁（已有实现）
s.lock.LockUuid(req.Uuid)
defer s.lock.UnlockUuid(req.Uuid)
```

#### 3. 错误信息国际化

使用 `i18n.Translate` 返回多语言错误提示：

```go
import "main/i18n"

// 超限错误
return errors.New(fmt.Sprintf(
    i18n.Translate(ctx.GetLanguage(), "物品[%s]本月限购%v%s，已使用%v，本次申请%v，超出限额"),
    item.MaterialName,
    config.QuotaLimit,
    config.UnitName,
    usedQty,
    orderQty,
))

// 单位不匹配错误
return errors.New(fmt.Sprintf(
    i18n.Translate(ctx.GetLanguage(), "物品[%s]限购单位为[%s]，请使用指定单位"),
    item.MaterialName,
    config.UnitName,
))
```

#### 4. 统计查询 SQL 示例

```sql
-- 查询某物品本月已使用限额
SELECT COALESCE(SUM(poi.num), 0) as used_qty
FROM ttpos_purchase_order po
JOIN ttpos_purchase_order_item poi ON po.uuid = poi.purchase_order_uuid
WHERE po.purchase_type = 2                          -- 品牌采购
  AND poi.material_uuid = ?                         -- 物品UUID
  AND poi.unit_uuid = ?                             -- 限购单位UUID
  AND po.status IN (1, 2, 4, 5)                     -- 待审核、已通过、部分收货、待总部审核
  AND po.uuid != ?                                  -- 排除当前单据
  AND FROM_UNIXTIME(po.create_time, '%Y-%m') = ?    -- 当前月份
  AND po.delete_time = 0
```

---

## 🔮 暂缓实现清单（MVP 后续迭代）

以下功能在 MVP 版本中暂不实现，但已在数据模型和架构设计中预留扩展能力。

### 配置管理类

| 功能             | 说明                         | 预留字段/设计                              | 优先级 |
| ---------------- | ---------------------------- | ------------------------------------------ | ------ |
| **前端配置页面** | Shop 端限购配置 CRUD 界面    | Model/Repository 已就绪                    | P1     |
| **总部统一配置** | 总部配置规则，下发到所有子店 | `config_source`、`headquarter_config_uuid` | P1     |
| **配置优先级**   | 总部配置优先级高于门店配置   | `config_source` 字段判断                   | P1     |
| **批量导入导出** | Excel 批量管理限购配置       | -                                          | P2     |

### 查询展示类

| 功能                 | 说明                           | 预留字段/设计    | 优先级 |
| -------------------- | ------------------------------ | ---------------- | ------ |
| **已用限额查询接口** | 返回物品本月已用/剩余额度      | 实时查询统计实现 | P1     |
| **采购页面额度提示** | 前端展示"本月已采购 X，剩余 Y" | 依赖查询接口     | P2     |
| **限购预警通知**     | 接近限额时推送提醒             | -                | P3     |

### 策略扩展类

| 功能               | 说明                                   | 预留字段/设计      | 优先级 |
| ------------------ | -------------------------------------- | ------------------ | ------ |
| **柔性超限策略**   | 超限时允许提交但标记，由总部决定       | `strict_mode` 字段 | P2     |
| **多周期支持**     | 支持季度、年度限购周期                 | `period_type` 字段 | P2     |
| **物品分类限购**   | 按分类设置限购（如"清洁用品"整体限购） | 需新增关联表       | P3     |
| **差异化门店限额** | 不同门店配置不同限额                   | 总部配置时支持     | P3     |

### 运营支持类

| 功能             | 说明                   | 预留字段/设计 | 优先级 |
| ---------------- | ---------------------- | ------------- | ------ |
| **限购统计报表** | 各门店限购使用情况汇总 | -             | P2     |
| **超限审批流程** | 超限时走特殊审批流程   | -             | P3     |
| **限购日志审计** | 记录限购配置变更历史   | -             | P3     |

---

### MVP 运营使用方式

在前端配置页面上线前，运营人员通过数据库直接管理限购配置：

```sql
-- 查询现有限购配置
SELECT * FROM purchase_quota_config WHERE status = 1;

-- 新增限购配置
INSERT INTO purchase_quota_config
(uuid, material_uuid, material_code, unit_uuid, unit_name, quota_limit, status)
VALUES
(/*雪花ID*/, /*物品UUID*/, '物品编码', /*单位UUID*/, '箱', 1, 1);

-- 修改限额
UPDATE purchase_quota_config
SET quota_limit = 2, updated_at = NOW()
WHERE material_uuid = /*物品UUID*/;

-- 禁用限购
UPDATE purchase_quota_config
SET status = 0, updated_at = NOW()
WHERE material_uuid = /*物品UUID*/;

-- 删除限购配置
DELETE FROM purchase_quota_config WHERE material_uuid = /*物品UUID*/;
```

---

**版本**: v1.0.0-MVP  
**创建日期**: 2025-12-25  
**维护者**: BenDaye
