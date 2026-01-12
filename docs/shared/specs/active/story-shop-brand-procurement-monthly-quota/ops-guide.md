# 品牌采购限购功能 - 运维配置手册

> **目标受众**：运维工程师、DBA、系统管理员  
> **用途**：生产环境部署、配置管理、故障排查、性能优化

## 功能概述

### 二维限购控制架构

```
┌─────────────────────────────────────────────────────────────────┐
│                    门店提交品牌采购申请                          │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         ▼
         ┌────────────────────────────────────────┐
         │  维度1：每日申请次数检查               │
         │  配置来源：ttpos_setting               │
         │  purchase_brand_daily_limit            │
         │  ⚙️  值=-1 表示不限制，直接通过        │
         └───────────┬────────────────────────────┘
                     │
         ┌───────────┴────────────┐
         │                        │
      超限 ▼                   通过 ▼
    ┌──────────┐      ┌────────────────────────────────────────┐
    │ ❌ 拒绝  │      │  维度2：物品限额检查                   │
    │   提交   │      │  配置来源：purchase_quota_config       │
    └──────────┘      │  按物品+单位精确控制                   │
                      │  支持按天/月度周期                     │
                      │  ⚙️  无配置/禁用 表示不限制，直接通过  │
                      └───────────┬────────────────────────────┘
                                  │
                      ┌───────────┴────────────┐
                      │                        │
                   超限 ▼                   通过 ▼
                 ┌──────────┐          ┌──────────┐
                 │ ❌ 拒绝  │          │ ✅ 提交  │
                 │   提交   │          │   成功   │
                 └──────────┘          └──────────┘
```

## 数据库表结构

### 表关系图

```
┌─────────────────────────┐
│   ttpos_setting         │  ◄──── 全局配置表
│─────────────────────────│
│ • key (配置键)          │
│ • values (JSON配置值)   │
│ • describe (说明)       │
└─────────┬───────────────┘
          │ 读取配置
          ▼
    [每日次数/单次数量]


┌──────────────────────────────┐
│ purchase_quota_config        │  ◄──── 物品限购配置主表
│──────────────────────────────│
│ • uuid (配置ID)              │
│ • material_code (物品编码-erp的)   │
│ • unit_code (单位编码-erp的)       │
│ • quota_limit (限购数量)     │
│ • apply_to_all_shops (全店)  │
│ • period_type (周期类型)     │
│ • status (启用状态)          │
└──────┬───────────────────────┘
       │ 1对多
       ├──────────────────┐
       │                  │ 读取配置
       ▼                  ▼
┌────────────────────────┐   [月度限额]
│ purchase_quota_config_ │
│ shop                   │  ◄──── 门店关联表
│────────────────────────│
│ • config_uuid          │
│ • company_uuid         │
└────────────────────────┘


┌──────────────────────────────┐
│ ttpos_purchase_order         │  ◄──── 采购订单表
│──────────────────────────────│
│ • uuid (订单ID)              │
│ • purchase_type (采购类型)   │
│ • status (订单状态)          │
│ • order_time (提交时间)      │
│ • company_uuid (门店UUID)    │
└──────┬───────────────────────┘
       │ 1对多
       │
       ▼
┌──────────────────────────────┐
│ ttpos_purchase_order_item    │  ◄──── 订单明细表
│──────────────────────────────│
│ • purchase_order_uuid        │
│ • material_code (物品编码)   │
│ • erpnext_uom (单位编码)     │
│ • num (数量)                 │
└──────┬───────────────────────┘
       │ 统计查询
       ▼
   [已使用额度]
```

### 核心表详情

#### 1. ttpos_setting - 全局配置表

```sql
-- 查看配置
SELECT `key`, `values`, `describe`
FROM ttpos_setting
WHERE `key` IN ('purchase_brand_daily_limit', 'purchase_brand_single_qty_limit');
```

| 字段 | 类型 | 说明 | 示例值 |
|------|------|------|--------|
| `key` | varchar(255) | 配置键名 | `purchase_brand_daily_limit` |
| `values` | text | JSON 配置值 | `{"limit": -1}` |
| `describe` | varchar(255) | 配置说明 | `每日申请次数上限` |

**关键配置项：**

- `purchase_brand_daily_limit`：每日申请次数上限，`-1` 表示不限制

#### 2. purchase_quota_config - 物品限购配置主表

```sql
-- 查看所有限购配置
SELECT 
    uuid,
    material_code,
    unit_code,
    quota_limit,
    apply_to_all_shops,
    period_type,
    status
FROM purchase_quota_config
WHERE delete_time = 0;
```

| 字段 | 类型 | 说明 | 示例值 | 索引 |
|------|------|------|--------|------|
| `uuid` | bigint(20) | 配置唯一ID | 雪花算法生成 | UNIQUE |
| `material_code` | varchar(100) | 物品编码 | `ITEM001` | INDEX |
| `unit_code` | varchar(50) | 限购单位编码 | `kg` | - |
| `quota_limit` | decimal(10,2) | 限购数量 | `100.00` | - |
| `apply_to_all_shops` | int(4) | 全店铺标识 | `1`=是, `0`=否 | - |
| `period_type` | int(4) | 周期类型 | `0`=按天, `1`=月度 | - |
| `status` | int(4) | 启用状态 | `1`=启用, `0`=禁用 | INDEX |

#### 3. purchase_quota_config_shop - 门店关联表

```sql
-- 查看门店关联配置
SELECT 
    pqcs.config_uuid,
    pqc.material_code,
    pqcs.company_uuid
FROM purchase_quota_config_shop pqcs
JOIN purchase_quota_config pqc ON pqcs.config_uuid = pqc.uuid
WHERE pqcs.delete_time = 0;
```

### 索引策略

| 表 | 索引名 | 字段 | 类型 | 说明 |
|----|--------|------|------|------|
| `purchase_quota_config` | `uk_uuid` | `uuid` | UNIQUE | 主键唯一索引 |
| `purchase_quota_config` | `idx_material` | `material_code` | INDEX | 物品查询优化 |
| `purchase_quota_config` | `idx_status` | `status` | INDEX | 状态过滤 |
| `purchase_quota_config` | `idx_delete_time` | `delete_time` | INDEX | 软删除过滤 |
| `purchase_quota_config_shop` | `uk_config_company` | `config_uuid, company_uuid` | UNIQUE | 防重复关联 |
| `purchase_quota_config_shop` | `idx_config` | `config_uuid` | INDEX | 配置查询 |
| `purchase_quota_config_shop` | `idx_company` | `company_uuid` | INDEX | 门店查询 |

---

## 配置管理

### 场景 1：初始化全局配置

**目的**：设置每日申请次数上限

**操作步骤：**

```sql
-- 在总部数据库执行
-- 设置每日申请次数上限为 5 次（根据业务需求调整）
UPDATE ttpos_setting 
SET `values` = '{"limit": 5}' 
WHERE `key` = 'purchase_brand_daily_limit';

-- 验证配置
SELECT `key`, `values`, `describe`
FROM ttpos_setting
WHERE `key` = 'purchase_brand_daily_limit';
```

**配置说明：**

| 配置值 | 含义 | 适用场景 |
|--------|------|----------|
| `-1` | 不限制 | 测试环境、初期试运行 |
| `3` | 每天3次 | 中小型门店 |
| `5` | 每天5次 | 大型门店 |
| `10` | 每天10次 | 总部或特殊门店 |

**回滚方案：**

```sql
-- 恢复为不限制
UPDATE ttpos_setting 
SET `values` = '{"limit": -1}' 
WHERE `key` = 'purchase_brand_daily_limit';
```

---

### 场景 2：添加物品限购配置（全店铺）

**目的**：为特定物品设置月度采购限额，适用于所有门店

**操作步骤：**

```sql
-- 在总部数据库执行
INSERT INTO purchase_quota_config (
    uuid,
    material_code,
    unit_code,
    quota_limit,
    apply_to_all_shops,
    period_type,
    strict_mode,
    config_source,
    status,
    create_time,
    update_time
) VALUES (
    -- ⚠️ 注意：uuid 需要使用雪花算法生成，这里仅为示例
    123456789012345678,
    'ITEM001',         -- 物品编码（从 ERP 系统获取）
    'kg',              -- 限购单位
    100.00,            -- 月度限额 100kg
    1,                 -- 应用到所有门店
    1,                 -- 月度限购
    1,                 -- 严格拒绝
    2,                 -- 总部下发
    1,                 -- 启用
    UNIX_TIMESTAMP(),
    UNIX_TIMESTAMP()
);

-- 验证配置
SELECT * FROM purchase_quota_config WHERE material_code = 'ITEM001';
```

**配置参数说明：**

| 参数 | 可选值 | 说明 |
|------|--------|------|
| `period_type` | `0` | 按天限购 |
|  | `1` | 月度限购 |
| `apply_to_all_shops` | `1` | 全店铺生效 |
|  | `0` | 仅指定门店 |
| `status` | `1` | 启用 |
|  | `0` | 禁用 |

**UUID 生成方法：**

```go
// 在 Go 代码中生成
import "ttpos-server-go/pkg/utils"
uuid := utils.GenUUID()
```

或使用在线工具生成雪花ID。

---

### 场景 3：添加物品限购配置（指定门店）

**目的**：为特定物品设置限额，仅适用于指定门店

**操作步骤：**

```sql
-- 步骤1：创建限购配置
INSERT INTO purchase_quota_config (
    uuid,
    material_code,
    unit_code,
    quota_limit,
    apply_to_all_shops,  -- ⚠️ 设置为 0
    period_type,
    strict_mode,
    config_source,
    status,
    create_time,
    update_time
) VALUES (
    234567890123456789,
    'ITEM002',
    'box',
    50.00,
    0,                   -- 不应用到所有门店
    0,                   -- 按天限购
    1,
    2,
    1,
    UNIX_TIMESTAMP(),
    UNIX_TIMESTAMP()
);

-- 步骤2：关联到指定门店
INSERT INTO purchase_quota_config_shop (
    config_uuid,
    company_uuid,
    create_time
) VALUES 
(234567890123456789, 8609817471094784, UNIX_TIMESTAMP()),  -- 门店A
(234567890123456789, 8609817471094785, UNIX_TIMESTAMP());  -- 门店B

-- 验证配置
SELECT 
    pqc.material_code,
    pqc.quota_limit,
    pqcs.company_uuid
FROM purchase_quota_config pqc
JOIN purchase_quota_config_shop pqcs ON pqc.uuid = pqcs.config_uuid
WHERE pqc.material_code = 'ITEM002';
```
