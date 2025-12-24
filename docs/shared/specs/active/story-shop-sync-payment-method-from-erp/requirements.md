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
- **子店**：连锁分店（`IsBranch() == true`）

### 1.3 核心目标

实现所有商户类型（散户、总店、子店）都能够从 ERP 系统同步支付方式数据到 TTPOS，支持首次同步和后续增量同步。

**子店特殊处理**：
- 子店会先从 ERP 同步支付方式
- 然后根据 `syncHeadquarterData` 参数决定是否从总店同步支付方式

**细粒度同步行为**：
- 支付方式同步始终执行（不受 `paymentDataChecked` 控制）
- `paymentDataChecked` 参数仅控制子店是否从总店同步支付方式

---

## 二、功能需求

### 2.1 同步流程（重要）

**两阶段同步策略**：

#### 第一阶段：从 ERP 同步（所有商户）
- **适用对象**：散户、总店、子店
- **同步内容**：ERP 系统中的所有支付方式
- **匹配规则**：优先 PaymentID，降级 erpnext_payment
- **触发时机**：始终执行

#### 第二阶段：从总店同步（仅子店）
- **适用对象**：仅子店
- **同步内容**：总店的支付方式
- **匹配规则**：通过 payment_name 匹配
- **触发时机**：根据 `syncHeadquarterData` 参数决定
  - 普通同步：`syncHeadquarterData=true`，总是执行
  - 细粒度同步：`syncHeadquarterData=paymentDataChecked`，根据用户勾选决定

### 2.2 同步范围（Given）

**适用场景**：
- 散户（`IsTtposSite() == true`）- 仅第一阶段
- 总店（`IsHeadquarter() == true`）- 仅第一阶段
- 子店（`IsBranch() == true`）- 第一阶段 + 第二阶段（按需）

**前置条件**：
- 商户已配置 ERP 集成
- ERP 系统中已配置支付方式（Mode of Payment）
- 子店：总店必须存在且可访问

### 2.3 首次同步规则（When: ERP 新增支付方式）

#### 2.3.1 数据映射规则

**ERP → TTPOS 字段映射**：

| ERP 字段（Mode of Payment） | TTPOS 字段（PaymentMethod） | 映射规则 |
|---------------------------|---------------------------|---------|
| payment_id（PaymentID） | erpnext_payment_id | 直接赋值（优先使用） |
| mode_of_payment（名称） | payment_name | 直接赋值 |
| mode_of_payment（名称） | name | 直接赋值（中文名称） |
| enabled（启用状态） | status | 1=启用, 0=禁用 |
| - | code | 使用 generatePaymentCode 生成（20000 起，递增 100） |
| - | source | 固定为 1（PaymentSourceDefault，手动添加） |
| - | logo_file_uuid | 0 |
| - | qrcode_file_uuid | 0 |
| - | fee_percent | 0.0000 |
| - | default_img | /image/pay/ja_pay.png |

#### 2.3.2 显示配置（Then）

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

#### 2.3.3 Code 生成规则

**使用现有方法 `generatePaymentCode`**：
- 从 20000 开始
- 每次递增 100
- 与手动添加的支付方式使用相同规则
- 避免与系统保留 code 冲突（10=余额, 40=现金, 90xxx=连连支付）

### 2.4 来源标识（Source）

**来源字段值**：
```go
source = 1  // PaymentSourceDefault（手动添加）
```

**说明**：
- ERP 同步的支付方式使用 `source=1`（手动添加），与店铺手动添加的支付方式保持一致
- 区别于 `source=0`（系统默认，如现金、余额）和 `source=2`（连连支付）

### 2.5 默认图标

**图标配置**：
```go
default_img = "/image/pay/ja_pay.png"
logo_file_uuid = 0
```

### 2.6 后续同步规则（When: 后续同步）

#### 2.6.1 同步策略

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
- erpnext_payment_id（不变）

#### 2.6.2 关联匹配规则

**识别已存在支付方式的优先级**：

1. **优先使用 PaymentID 匹配**（如果 ERP 返回了 payment_id）：
```sql
WHERE erpnext_payment_id = ? AND delete_time = 0
```

2. **降级使用 erpnext_payment 匹配**（如果 ERP 未返回 payment_id）：
```sql
WHERE erpnext_payment = ? AND delete_time = 0
```

**匹配逻辑**：
- 如果 ERP 返回的 `payment_id` 不为空，优先使用 `erpnext_payment_id` 进行匹配
- 如果 ERP 返回的 `payment_id` 为空，则使用 ERP 的 `name` 去匹配 TTPOS 的 `erpnext_payment` 字段
- 匹配到记录则更新，未匹配到则创建新记录

### 2.7 子店从总店同步规则（第二阶段）

#### 2.7.1 同步范围

**仅适用于子店**：
- 在完成从 ERP 同步后执行
- 同步总店的所有支付方式（排除 code=10 和 code=40）

#### 2.7.2 匹配规则

**通过 payment_name 和 source 匹配**：
```sql
WHERE payment_name = ? AND source = 1 AND delete_time = 0
```

#### 2.7.3 同步行为

**如果子店已有同名支付方式**：
- 特殊 code（90111/90222/90333）：仅更新 `headquarter_uuid`
- 普通 code：跳过，不修改

**如果子店没有该支付方式**：
- 创建新记录
- 生成新的 code（从 20000 开始递增）
- 设置 `headquarter_uuid` 为总店 UUID
- 保持 `source=1`（手动添加）

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

- 一个 ERP 支付方式可以通过两种方式关联：
  - 优先：通过 `erpnext_payment_id` 字段（ERP PaymentID）
  - 降级：通过 `name` 字段（支付方式名称）
- 当 ERP 返回 `payment_id` 时，使用 `payment_id` 作为唯一标识
- 当 ERP 未返回 `payment_id` 时，使用 `name` 作为唯一标识

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
- erpnext_payment_id（ERP PaymentID，首次设置后不变）

---

## 五、边界条件

### 5.1 不支持的场景

- ❌ 不支持双向同步（TTPOS → ERP）

### 5.2 支持的场景变更

以下场景现已支持：
- ✅ 子店可以直接从 ERP 同步（原来不支持）
- ✅ 所有支付方式都可以同步（原来排除系统保留支付方式）

### 5.3 Code 范围说明

- **0-19999**: 系统保留
  - `10`: 余额
  - `40`: 现金
  - `90111/90222/90333`: 连连支付
- **20000+**: 手动添加和 ERP 同步共用范围
  - 每次递增 100
  - ERP 同步使用 `generatePaymentCode` 方法

**注意**：虽然系统保留的支付方式也会从 ERP 同步，但它们的 code 不会被修改，保持原有的保留 code 值。

---

## 六、验收标准

### 6.1 功能验收

✅ 散户可以同步 ERP 支付方式  
✅ 总店可以同步 ERP 支付方式  
✅ 子店可以同步 ERP 支付方式  
✅ 首次同步时正确创建支付方式  
✅ 首次同步时自动配置显示字段  
✅ 后续同步时仅更新状态字段  
✅ 后续同步时保留本地配置  
✅ code 自动生成且不重复  
✅ 优先使用 payment_id 进行匹配  
✅ payment_id 为空时降级使用 name 匹配  
✅ 所有 ERP 支付方式都可以同步（不排除保留支付方式）  

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
