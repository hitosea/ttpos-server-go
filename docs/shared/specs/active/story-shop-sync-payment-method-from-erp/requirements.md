# 需求文档：散户/总店同步 ERP 支付方式

**创建时间**: 2025-12-22  
**需求层级**: story  
**关联任务**: DooTask 37829  
**影响模块**: main/app/service/payment_method.go  
**影响终端**: shop, pos

---

## 一、需求概述

### 1.1 业务背景

- 当前系统支持子店从总店同步支付方式
- 散户和总店需要能够从 ERP 系统同步支付方式到 TTPOS
- ERP 系统（ERPNext）中维护的支付方式应该能够自动同步到 TTPOS 系统

### 1.2 用户角色

- **散户**：独立运营的商户（`IsTtposSite() == true`）
- **总店**：连锁总部（`IsHeadquarter() == true`）

### 1.3 核心目标

实现散户和总店能够从 ERP 系统同步支付方式数据到 TTPOS，支持首次同步和后续增量同步。

---

## 二、功能需求

### 2.1 同步范围（Given）

**适用场景**：
- 散户（`IsTtposSite() == true`）
- 总店（`IsHeadquarter() == true`）

**前置条件**：
- 商户已配置 ERP 集成
- ERP 系统中已配置支付方式（Mode of Payment）

### 2.2 首次同步规则（When: ERP 新增支付方式）

#### 2.2.1 数据映射规则

**ERP → TTPOS 字段映射**：

| ERP 字段（Mode of Payment） | TTPOS 字段（PaymentMethod） | 映射规则 |
|---------------------------|---------------------------|---------|
| mode_of_payment（名称） | payment_name | 直接赋值 |
| mode_of_payment（名称） | name | 直接赋值（中文名称） |
| enabled（启用状态） | status | 1=启用, 0=禁用 |
| - | code | 使用 generatePaymentCode 生成（20000 起，递增 100） |
| - | source | 固定为 1（PaymentSourceDefault，手动添加） |
| - | logo_file_uuid | 0 |
| - | qrcode_file_uuid | 0 |
| - | fee_percent | 0.0000 |
| - | default_img | /image/pay/ja_pay.png |
| - | erpnext_payment | ERP 支付方式名称（用于关联） |

#### 2.2.2 显示配置（Then）

**结账显示配置**：
```go
is_show_cashier = 1        // 收银机结账显示
is_show_assistant = 1      // 点餐助手结账显示
is_show_kiosk = 0          // 自助点餐机不显示
```

**充值显示配置**：
```go
is_show_member_recharge = 1 // 收银机充值显示
```

#### 2.2.3 Code 生成规则

**使用现有方法 `generatePaymentCode`**：
- 从 20000 开始
- 每次递增 100
- 与手动添加的支付方式使用相同规则
- 避免与系统保留 code 冲突（10=余额, 40=现金, 90xxx=连连支付）

### 2.3 来源标识（Source）

**来源字段值**：
```go
source = 1  // PaymentSourceDefault（手动添加）
```

**说明**：
- ERP 同步的支付方式使用 `source=1`（手动添加），与店铺手动添加的支付方式保持一致
- 区别于 `source=0`（系统默认，如现金、余额）和 `source=2`（连连支付）

### 2.4 默认图标

**图标配置**：
```go
default_img = "/image/pay/ja_pay.png"
logo_file_uuid = 0
```

### 2.5 后续同步规则（When: 后续同步）

#### 2.5.1 同步策略

**仅同步状态字段**：
```go
status = ERP.enabled ? 1 : 0
```

**不同步的字段**：
- payment_name（保持本地修改）
- name（保持本地修改）
- code（不变）
- source（不变）
- logo_file_uuid（保持本地配置）
- fee_percent（保持本地配置）
- is_show_* 系列字段（保持本地配置）
- default_img（不变）

#### 2.5.2 关联匹配规则

**识别已存在支付方式**：
```sql
WHERE erpnext_payment = ? AND delete_time = 0
```

---

## 三、非功能需求

### 3.1 性能要求

- 同步操作应在 5 秒内完成
- 支持批量同步（一次最多 100 条）

### 3.2 数据一致性

- 使用数据库事务保证原子性
- 同步失败时完整回滚
- 记录同步日志供追溯

### 3.3 容错处理

- ERP 服务不可用时返回友好错误
- 数据格式异常时跳过该条记录并记录日志
- 部分失败时继续处理其他记录

---

## 四、业务规则

### 4.1 支付方式唯一性

- 一个 ERP 支付方式（erpnext_payment）只能对应一个 TTPOS 支付方式
- 通过 erpnext_payment 字段关联

### 4.2 删除处理

- TTPOS 不主动删除支付方式
- ERP 中删除的支付方式，在 TTPOS 中仅禁用（status=0）

### 4.3 字段保护

**本地可编辑字段**（后续同步不覆盖）：
- payment_name（支付名称）
- name（中文名称）
- logo_file_uuid（图标）
- fee_percent（手续费率）
- is_show_cashier（收银机显示）
- is_show_assistant（点餐助手显示）
- is_show_kiosk（自助机显示）
- is_show_member_recharge（充值显示）

---

## 五、边界条件

### 5.1 不支持的场景

- ❌ 子店不能直接从 ERP 同步（子店从总店同步）
- ❌ 不支持双向同步（TTPOS → ERP）
- ❌ 不同步连连支付相关支付方式（code 90xxx）

### 5.2 特殊处理

**排除的支付方式**：
- Code = 10（余额支付，系统内置）
- Code = 40（现金支付，系统内置）
- Code = 90111/90222/90333（连连支付，独立管理）

### 5.3 Code 范围说明

- **0-19999**: 系统保留
  - `10`: 余额
  - `40`: 现金
  - `90111/90222/90333`: 连连支付
- **20000+**: 手动添加和 ERP 同步共用范围
  - 每次递增 100
  - ERP 同步使用 `generatePaymentCode` 方法

---

## 六、验收标准

### 6.1 功能验收

✅ 散户可以同步 ERP 支付方式  
✅ 总店可以同步 ERP 支付方式  
✅ 首次同步时正确创建支付方式  
✅ 首次同步时自动配置显示字段  
✅ 后续同步时仅更新状态字段  
✅ 后续同步时保留本地配置  
✅ code 自动生成且不重复  
✅ erpnext_payment 字段正确关联  

### 6.2 数据验证

✅ 支付方式名称正确  
✅ 状态同步正确（启用/禁用）  
✅ 默认图标正确  
✅ 手续费为 0  
✅ 显示配置符合需求  

### 6.3 异常处理

✅ ERP 服务异常时不影响系统运行  
✅ 同步失败时有清晰的错误提示  
✅ 数据异常时记录详细日志  

---

## 七、影响范围

### 7.1 代码影响

- `main/app/service/payment_method.go` - SyncPaymentMethod 方法
- `main/app/service/rpc/erp/setup.go` - 可能需要添加获取支付方式接口

### 7.2 数据库影响

- 表：`ttpos_payment_method`
- 字段：已有字段，无需新增

### 7.3 API 影响

- 现有同步接口：`POST /api/v1/shop/sync`
- 无需新增 API

### 7.4 终端影响

| 终端 | 影响程度 | 说明 |
|-----|---------|------|
| shop | 高 | 管理支付方式 |
| pos | 中 | 使用支付方式结账 |
| assistant | 中 | 使用支付方式结账 |
| kiosk | 低 | 首次同步不显示 |

---

## 八、参考资料

### 8.1 相关代码

- `main/app/service/payment_method.go` 第 744-752 行
- `main/app/model/payment_order.go` - PaymentMethod 模型
- `main/app/service/sync.go` - 同步服务

### 8.2 相关文档

- `docs/shared/specs/active/shop-headquarters-branch-granular-sync-backend/PAYMENT_METHOD_SYNC_RULES.md`
- `docs/shared/specs/active/story-ttpos-erp-payment-mode-save/`

### 8.3 ERP 集成

- `ttpos-bmp/app/ttpos-erp/` - ERP 微服务
- ERPNext Mode of Payment 文档

---

**需求确认人**：待确认  
**预估工作量**：5 Story Points  
**优先级**：中
