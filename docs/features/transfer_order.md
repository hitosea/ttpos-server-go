# Transfer Order Service 调拨单服务说明文档

## 📋 概述

`service/transfer_order/transfer_order.go` 是 TTPOS 系统的调拨单管理服务，负责处理门店间的物品调拨申请、审批流程、库存流转等功能。该服务支持调入和调出两种类型，支持多级审批流程，并与 ERP 系统集成。

**文件路径**: `/main/app/service/transfer_order/transfer_order.go`  
**代码行数**: 1556 行  
**接口定义**: `ITransferOrderSrv`  
**实现结构**: `transferOrderSrv`

---

## 🏗️ 架构设计

### 接口定义 (ITransferOrderSrv)

```go
type ITransferOrderSrv interface {
    // 调拨单管理
    GetTransferOrderList(ctx context.Context, req req.TransferOrderListReq) (resp.TransferOrderListResp, error)
    GetTransferOrderDetail(ctx context.Context, req req.TransferOrderDetailReq) (resp.TransferOrderDetailResp, error)
    CreateTransferOrder(ctx context.Context, req req.TransferOrderCreateReq) (resp.TransferOrderCreateResp, error)
    UpdateTransferOrder(ctx context.Context, req req.TransferOrderUpdateReq) error
    DeleteTransferOrder(ctx context.Context, req req.TransferOrderDeleteReq) error
    SubmitTransferOrder(ctx context.Context, req req.TransferOrderSubmitReq) error
    ApproveTransferOrder(ctx context.Context, req req.TransferOrderApproveReq) error
    RejectTransferOrder(ctx context.Context, req req.TransferOrderRejectReq) error
    ReceiveTransferOrder(ctx context.Context, req req.TransferOrderReceiveReq) error

    // 审批流程和日志
    GetTransferOrderApprovalList(ctx context.Context, req req.TransferOrderApprovalListReq) (resp.TransferOrderApprovalListResp, error)
    GetTransferOrderLogList(ctx context.Context, req req.TransferOrderLogListReq) (resp.TransferOrderLogListResp, error)

    // 下拉列表
    GetTransferOrderCompanyList(ctx context.Context) (resp.TransferOrderCompanyListResp, error)
    GetTransferOrderWarehouseList(ctx context.Context) (resp.TransferOrderWarehouseListResp, error)
    GetTransferOrderMaterialList(ctx context.Context, req req.TransferOrderMaterialListReq) (material_resp.MaterialListWithPaginationResp, error)
}
```

### 依赖服务

```go
type transferOrderSrv struct {
    dbm         *database.DBManager      // 数据库管理器
    materialSrv service.IMaterialSrv     // 物品服务
    lock        lock.Lock                 // 并发锁
    validator   *transferOrderValidator   // 验证器
    helper      *transferOrderHelper      // 辅助方法
}
```

---

## 🎯 核心功能

### 1. 调拨单类型

系统支持两种调拨类型：

| 类型常量 | 说明 | 特点 |
|---------|------|-----|
| `TransferTypeIn = 1` | 调入 | 本店作为收货方，选择发货门店和入库仓库 |
| `TransferTypeOut = 2` | 调出 | 本店作为发货方，选择收货门店和出库仓库 |

### 2. 调拨单状态流转

```
待提交 (Draft)
    ↓ [提交]
待审核 (Pending)
    ↓ [审批通过] / [驳回]
待收货 (Receiving) / 已驳回 (Rejected)
    ↓ [收货]
已完成 (Completed)
```

**状态常量**:
- `TransferOrderStatusDraft = 0` - 待提交
- `TransferOrderStatusPending = 1` - 待审核
- `TransferOrderStatusRejected = 2` - 已驳回
- `TransferOrderStatusReceiving = 3` - 待收货
- `TransferOrderStatusCompleted = 4` - 已完成

### 3. 创建调拨单 (CreateTransferOrder)

**功能描述**: 创建新的调拨单，支持调入和调出两种类型。

#### 创建流程

```
1. 验证请求参数
   - 调拨类型验证
   - 门店和仓库验证
   - 物品列表验证
   ↓
2. 加锁防止并发创建（使用字符串锁保护编号生成）
   ↓
3. 设置发货门店和收货门店
   - 调入：本店为收货方，需选择发货门店
   - 调出：本店为发货方，需选择收货门店
   ↓
4. 验证门店和仓库是否存在
   ↓
5. 获取物品列表并验证
   ↓
6. 生成调拨单编号（TR+12位数字，全平台唯一）
   ↓
7. 生成调拨单UUID
   ↓
8. 创建调拨单主表和明细（事务）
   ↓
9. 记录操作日志
   ↓
10. 返回调拨单UUID和编号
```

#### 编号生成规则

```go
// 格式：TR + 8位日期(YYYYMMDD) + 4位序号
// 示例：TR202501120001
orderNo := fmt.Sprintf("TR%s%04d", dateStr, seq)
```

**编号特点**:
- 全平台唯一
- 使用 Redis 自增序列保证唯一性
- 按日期重置序号（每天从1开始）

#### 响应数据

```go
type TransferOrderCreateResp struct {
    Uuid    uint64 // 调拨单UUID
    OrderNo string // 调拨单编号
}
```

---

### 4. 更新调拨单 (UpdateTransferOrder)

**功能描述**: 更新调拨单信息，仅待提交状态可更新。

#### 更新限制

- 只有 `TransferOrderStatusDraft` 状态的调拨单才能更新
- 更新时会重新创建明细（先删除旧明细，再创建新明细）

#### 更新流程

```
1. 验证请求参数
   ↓
2. 加锁（UUID锁）
   ↓
3. 查询调拨单并验证状态
   ↓
4. 验证门店和仓库
   ↓
5. 获取物品列表
   ↓
6. 更新主表和明细（事务）
   ↓
7. 记录操作日志
```

---

### 5. 删除调拨单 (DeleteTransferOrder)

**功能描述**: 删除调拨单，仅待提交状态可删除。

#### 删除限制

- 只有 `TransferOrderStatusDraft` 状态的调拨单才能删除
- 执行软删除（设置 `delete_time`）

---

### 6. 提交调拨单 (SubmitTransferOrder)

**功能描述**: 提交调拨单进入审批流程，触发库存验证和审批流程创建。

#### 提交流程

```
1. 验证请求参数
   ↓
2. 加锁（UUID锁）
   ↓
3. 查询调拨单并验证状态
   ↓
4. 验证物品列表不为空
   ↓
5. 验证单位数量是否全部为0（如未确认）
   ↓
6. 验证物品库存是否充足
   ↓
7. 验证物品状态（是否禁用、是否删除）
   ↓
8. 更新状态为待审核（事务）
   - 删除数量为0的物品
   - 更新状态和提交时间
   - 设置下一个审批门店
   ↓
9. 创建审批流程
   ↓
10. 记录操作日志
```

#### 库存验证

**调入类型**:
- 验证发货门店的普通仓库库存（不包括在途）
- 库存不足时提示："物品 %s 的可出库数量不足。\n\n请更换发货门店"

**调出类型**:
- 验证本店选择出库仓库的库存
- 库存不足时提示："物品 %s 的可调拨数量不足。\n\n请更换出库仓库"

#### 物品状态验证

提交时会检查物品状态：
- 如果物品已禁用，提示："物品 %s 的状态已关闭。\n\n提交后将移除该物品，是否继续提交？"
- 如果物品已删除，提示："物品 %s 未找到。\n\n提交后将移除该物品，是否继续提交？"
- 用户确认后，会自动移除这些物品

---

### 7. 审批流程 (ApproveTransferOrder)

**功能描述**: 审批通过调拨单，支持多级审批流程。

#### 审批类型

| 审批类型 | 常量 | 说明 |
|---------|------|-----|
| 发货门店审批 | `TransferApprovalTypeSender` | 发货门店审核 |
| 发货门店上级审批 | `TransferApprovalTypeSenderParent` | 发货门店的上级门店审批 |
| 收货门店审批 | `TransferApprovalTypeReceiver` | 收货门店审核 |
| 收货门店上级审批 | `TransferApprovalTypeReceiverParent` | 收货门店的上级门店审批 |

#### 审批流程

```
1. 验证请求参数
   ↓
2. 加锁（UUID锁）
   ↓
3. 获取调拨单数据库
   ↓
4. 查询调拨单并验证状态
   ↓
5. 验证审批权限（当前门店是否为下一个审批门店）
   ↓
6. 获取当前审批节点
   ↓
7. 发货门店审批时：
   - 验证出库仓库是否选择
   - 验证物品库存是否充足
   ↓
8. 收货门店审批时：
   - 验证入库仓库是否选择
   - 验证发货门店库存是否充足（极限操作场景）
   ↓
9. 验证物品状态（发货门店和收货门店审批时）
   ↓
10. 更新审批节点为已通过（事务）
    ↓
11. 查找下一个审批节点
    ↓
12. 如果还有下一个审批节点：
    - 更新下一个审批门店信息
    ↓
13. 如果所有审批完成：
    - 更新状态为待收货
    - 更新在途仓库存
    - 调用ERP接口（如开启ERP）
    - 复制数据到总部
    ↓
14. 记录操作日志
```

#### 审批规则

**同上级门店合并审批**:
- 如果发货门店和收货门店的上级门店是同一个，则只进行一次上级审批
- 审批类型显示为 `parent`（待上级审批）

**审批顺序**:
1. 发货门店审批（如需要）
2. 发货门店上级审批（如需要）
3. 收货门店上级审批（如需要）
4. 收货门店审批（如需要）

**审批配置**:
- 总部可配置是否开启审批或库存流转
- 上级门店可配置是否开启审批或库存流转
- 配置对修改后的新调拨单生效

---

### 8. 驳回调拨单 (RejectTransferOrder)

**功能描述**: 驳回调拨单，终止审批流程。

#### 驳回流程

```
1. 验证请求参数
   ↓
2. 加锁（UUID锁）
   ↓
3. 获取调拨单数据库
   ↓
4. 查询调拨单并验证状态
   ↓
5. 验证审批权限
   ↓
6. 获取当前审批节点
   ↓
7. 更新审批节点为已驳回（事务）
   ↓
8. 更新调拨单为已驳回状态
   ↓
9. 复制数据到总部（非第一个审批节点）
   ↓
10. 记录操作日志
```

#### 驳回信息

驳回时会记录：
- 驳回节点（发货门店驳回、发货门店上级驳回、收货门店驳回、收货门店上级驳回）
- 驳回原因
- 驳回时间
- 审批人信息

---

### 9. 收货调拨单 (ReceiveTransferOrder)

**功能描述**: 收货门店确认收货，完成调拨流程，库存从在途仓转入目标仓库。

#### 收货流程

```
1. 验证请求参数
   ↓
2. 加锁（UUID锁）
   ↓
3. 获取调拨单数据库
   ↓
4. 查询调拨单并验证状态
   ↓
5. 验证收货权限（只有收货门店可以收货）
   ↓
6. 验证入库仓库
   ↓
7. 验证物品状态（是否禁用、是否删除）
   ↓
8. 更新调拨单为已完成状态（事务）
   ↓
9. 复制数据到总部
   ↓
10. 移动在途库存到目标仓库
    ↓
11. 调用ERP接口（如开启ERP）
    ↓
12. 记录操作日志
```

#### 库存流转

**收货前**:
- 发货门店：库存从普通仓库扣减
- 收货门店：库存进入在途仓库

**收货后**:
- 收货门店：库存从在途仓库转入选择的入库仓库

---

### 10. 获取调拨单列表 (GetTransferOrderList)

**功能描述**: 分页获取调拨单列表，支持多条件筛选。

#### 查询范围

**可见性规则**:
- **总店**: 可见自己及所有子店（包括一级+二级门店）的调拨单
- **一级分店（上级门店）**: 可见自己及下属门店（二级门店）的调拨单
- **二级门店（普通子店）**: 可见自己的调拨单

**查询逻辑**:
- 从总部数据库（saas库）和门店数据库（shop库）联合查询
- 根据门店关系过滤可见的调拨单

#### 筛选条件

| 参数 | 说明 | 类型 |
|-----|------|-----|
| `order_no` | 单据编号或调拨单号搜索 | string |
| `status_in` | 状态筛选 | []int |
| `order_time_start` | 单据时间开始 | int |
| `order_time_end` | 单据时间结束 | int |
| `opposite_company_uuids` | 对方机构UUID列表 | []uint64 |
| `my_role` | 我的角色：all/sender/receiver/approver | []string |

#### 状态筛选

| 状态 | 说明 |
|-----|------|
| 全部 | 所有状态的调拨单 |
| 待提交 | 新增了保存的单据 |
| 待审核 | 提交了的单据（包括各审批节点） |
| 已驳回 | 被驳回的单据 |
| 待收货 | 全部流程审核通过，待收货门店收货 |
| 已完成 | 收货门店已收货 |

#### 排序规则

- 默认按创建时间倒序排序（`create_time DESC`）

---

### 11. 获取调拨单详情 (GetTransferOrderDetail)

**功能描述**: 根据UUID获取调拨单详细信息，包括明细、审批流程等。

#### 返回数据

```go
type TransferOrderDetailResp struct {
    TransferOrderInfo
    Items      []TransferOrderItemInfo // 调拨明细
    RejectInfo TransferOrderRejectInfo // 驳回信息
}
```

#### 特殊处理

**待提交状态**:
- 显示发货门店的可用库存数量（不包括在途）

**待审核状态**:
- 显示是否需要选择出库仓库（发货门店审批时）
- 显示是否需要选择入库仓库（收货门店审批时）

**待收货状态**:
- 入库仓库永远都可以选择

**审批进度**:
- `wait` - 待审核
- `sender` - 待发货门店审批
- `sender_parent` - 待发货门店上级审批
- `receiver` - 待收货门店审批
- `receiver_parent` - 待收货门店上级审批
- `parent` - 待上级审批（同上级门店合并）

---

### 12. 获取物品列表 (GetTransferOrderMaterialList)

**功能描述**: 获取调拨单可选的物品列表，支持调入和调出两种场景。

#### 调入场景

- 查询发货门店的普通仓库库存（不包括在途）
- 只查询总部同步下来的物品（全部物品即本店已同步的物品，不包括禁用）
- 显示可用库存数量

#### 调出场景

- 查询本店仓库当前选择出库仓库的库存
- 只查询总部同步下来的物品（本店已经同步下来的）
- 显示可用库存数量

---

## 🔄 审批流程设计

### 审批流程创建逻辑

审批流程的创建基于以下规则：

1. **总部配置优先**
   - 总部开启审批或库存，一级门店一定要审批或经过库存
   - 总部关闭审批或库存，是否经过一级门店审批或库存根据一级门店的设置来执行

2. **审批节点类型**
   - `sender` - 发货门店审批
   - `sender_parent` - 发货门店上级审批
   - `receiver_parent` - 收货门店上级审批
   - `receiver` - 收货门店审批

3. **同上级合并**
   - 如果发货门店和收货门店的上级门店是同一个，则只进行一次上级审批
   - 审批类型显示为 `parent`

4. **审批顺序**
   - 按照 `sequence` 字段排序，从1开始递增
   - 必须审批的节点（`is_required = 1`）按顺序执行

### 审批流程示例

**场景1：同上级门店调拨**
```
发货门店A（上级：总部） → 收货门店B（上级：总部）
审批流程：
1. 发货门店A审批
2. 总部审批（合并）
3. 收货门店B审批
```

**场景2：不同上级门店调拨**
```
发货门店A（上级：总部） → 收货门店B（上级：一级门店C）
审批流程：
1. 发货门店A审批
2. 总部审批（发货门店上级）
3. 一级门店C审批（收货门店上级）
4. 收货门店B审批
```

---

## 📊 库存流转机制

### 调入类型库存流转

```
1. 提交时：不扣减库存
   ↓
2. 发货门店审批通过：不扣减库存
   ↓
3. 所有审批通过：
   - 发货门店：从普通仓库扣减库存
   - 收货门店：库存进入在途仓库
   ↓
4. 收货门店收货：
   - 收货门店：从在途仓库转入选择的入库仓库
```

### 调出类型库存流转

```
1. 提交时：不扣减库存
   ↓
2. 本店审批通过：不扣减库存
   ↓
3. 所有审批通过：
   - 本店：从选择的出库仓库扣减库存
   - 收货门店：库存进入在途仓库
   ↓
4. 收货门店收货：
   - 收货门店：从在途仓库转入选择的入库仓库
```

### 在途仓库

- 在途仓库是系统自动管理的虚拟仓库
- 用于存放已审批通过但未收货的调拨物品
- 收货时从在途仓库转入目标仓库

---

## 🔐 权限控制

### 可见性控制

**总店**:
- 可见自己及所有子店（包括一级+二级门店）的调拨单
- 无论是否在流程中，都能查看下级门店的单子

**一级分店（上级门店）**:
- 可见自己及下属门店（二级门店）的调拨单
- 无论是否在流程中，都能查看下级门店的单子

**二级门店（普通子店）**:
- 只可见自己的调拨单

### 操作权限

**创建/更新/删除**:
- 只有待提交状态的调拨单才能操作
- 只有创建人所在门店可以操作

**审批**:
- 只有当前审批门店可以审批
- 验证 `next_approval_company_uuid` 是否匹配

**收货**:
- 只有收货门店可以收货
- 验证 `receiver_company_uuid` 是否匹配

---

## 🚨 错误处理

### 错误码定义

| 错误码 | 说明 | 使用场景 |
|-------|------|---------|
| `CodeErrorConfirmClose` | 需要确认关闭 | 库存不足、物品禁用等需要用户确认的场景 |
| `CodeErrorConfirmRequest` | 需要确认请求 | 提交时物品状态异常，需要用户确认是否继续 |

### 错误处理示例

```go
// 库存不足错误
return errors.NewWithCodeAndData(
    constant.CodeErrorConfirmClose,
    notEnoughItemNames,
    fmt.Sprintf("物品 %s 的可出库数量不足。\n\n请更换发货门店", names),
)

// 物品禁用错误
return errors.NewWithCode(
    constant.CodeErrorConfirmRequest,
    fmt.Sprintf("物品 %s 的状态已关闭。\n\n提交后将移除该物品，是否继续提交？", names),
)
```

---

## 🔧 配置项

### 审批配置

**总部配置**:
- `is_transfer_approval` - 是否开启审批
- `is_transfer_stock` - 是否开启库存流转

**上级门店配置**:
- `is_transfer_approval` - 是否开启审批
- `is_transfer_stock` - 是否开启库存流转

**配置规则**:
- 总部开启后，上级门店即使不开启，也需要经过审批或库存
- 总部关闭后，上级门店可关闭可开启，开启后只对自己下属门店生效
- 配置对修改后的新调拨单有效

---

## 📝 数据模型

### 主表 (ttpos_transfer_order)

| 字段 | 类型 | 说明 |
|-----|------|-----|
| `uuid` | bigint | 主键UUID |
| `company_uuid` | bigint | 所属公司UUID（发起门店） |
| `order_no` | varchar(255) | 单据编号（TR+12位数字） |
| `erp_order_no` | varchar(255) | ERP调拨单号（销售单号） |
| `transfer_type` | int | 调拨类型：1-调入 2-调出 |
| `sender_company_uuid` | bigint | 发货门店UUID |
| `receiver_company_uuid` | bigint | 收货门店UUID |
| `out_warehouse_erp_code` | varchar(255) | 出库仓库ERP编码 |
| `in_warehouse_erp_code` | varchar(255) | 入库仓库ERP编码 |
| `status` | int | 状态：0-待提交 1-待审核 2-已驳回 3-待收货 4-已完成 |
| `next_approval_company_uuid` | bigint | 下一个审批门店UUID |
| `order_time` | bigint | 单据日期（提交时间戳） |
| `submit_time` | bigint | 提交时间 |

### 明细表 (ttpos_transfer_order_item)

| 字段 | 类型 | 说明 |
|-----|------|-----|
| `uuid` | bigint | 主键UUID |
| `transfer_order_uuid` | bigint | 调拨单UUID |
| `material_uuid` | bigint | 物品UUID |
| `material_code` | varchar(255) | 物品编码 |
| `material_name` | text | 物品名称JSON |
| `valuation` | decimal(20,8) | 估值单价（基准单位） |

### 明细单位表 (ttpos_transfer_order_item_unit)

| 字段 | 类型 | 说明 |
|-----|------|-----|
| `uuid` | bigint | 主键UUID |
| `item_uuid` | bigint | 调拨单明细UUID |
| `unit_uuid` | bigint | 单位UUID |
| `unit_name` | text | 单位名称JSON |
| `unit_conversion_rate` | decimal(12,4) | 单位转换率 |
| `num` | decimal(22,4) | 调拨数量 |

### 审批流程表 (ttpos_transfer_order_approval)

| 字段 | 类型 | 说明 |
|-----|------|-----|
| `uuid` | bigint | 主键UUID |
| `transfer_order_uuid` | bigint | 调拨单UUID |
| `approval_type` | varchar(50) | 审批类型 |
| `approval_company_uuid` | bigint | 审批门店UUID |
| `sequence` | int | 审批顺序 |
| `status` | int | 审批状态：0-待审批 1-已通过 2-已驳回 |
| `is_required` | int | 是否必须审批：0-否 1-是 |

### 操作日志表 (ttpos_transfer_order_log)

| 字段 | 类型 | 说明 |
|-----|------|-----|
| `uuid` | bigint | 主键UUID |
| `transfer_order_uuid` | bigint | 调拨单UUID |
| `action` | varchar(50) | 操作动作 |
| `action_desc` | varchar(255) | 操作描述 |
| `old_status` | int | 操作前状态 |
| `new_status` | int | 操作后状态 |
| `operator_uuid` | bigint | 操作人UUID |

---

## 🔄 依赖关系

```
transferOrderSrv
  ├── dbm (database.DBManager) - 数据库管理器
  ├── materialSrv (IMaterialSrv) - 物品服务
  ├── lock (lock.Lock) - 并发锁
  ├── validator (transferOrderValidator) - 验证器
  │   ├── validateOrderItemUnitNumZero - 验证单位数量为0
  │   ├── ValidateMaterialsByCodes - 验证物品状态
  │   └── validateOrderItemStockNotEnough - 验证库存不足
  └── helper (transferOrderHelper) - 辅助方法
      ├── GenerateOrderNo - 生成调拨单编号
      ├── CreateLog - 创建操作日志
      ├── CreateApproval - 创建审批流程
      ├── GetMaterials - 获取物品列表
      ├── UpdateStockInTransit - 更新在途仓库存
      ├── MoveStockToTargetWarehouse - 移动库存到目标仓库
      ├── SaveMaterialTransfer - 调用ERP接口（调拨）
      ├── SavePurchaseReceipt - 调用ERP接口（收货）
      └── CopyDataToHeadquarter - 复制数据到总部
```

---

## 🌐 ERP集成

### 调拨接口 (SaveMaterialTransfer)

**调用时机**: 所有审批通过后，进入待收货状态时

**功能**:
- 在ERP中创建调拨单（销售单）
- 如果经过上级门店仓库，则创建多个销售单
- 返回ERP订单号

### 收货接口 (SavePurchaseReceipt)

**调用时机**: 收货门店收货时

**功能**:
- 在ERP中创建收货单（采购收货单）
- 返回ERP收货单号

### ERP订单号

- `erp_order_no` - ERP调拨单号（销售单号）
- `receipt_order_erp_code` - 收货单ERP编码

---

## 🧪 测试建议

### 单元测试覆盖

1. **创建调拨单测试**
   - 正常创建流程
   - 参数验证
   - 编号生成唯一性
   - 并发创建测试

2. **提交调拨单测试**
   - 库存验证
   - 物品状态验证
   - 审批流程创建

3. **审批流程测试**
   - 多级审批流程
   - 同上级门店合并审批
   - 审批权限验证
   - 库存流转验证

4. **收货流程测试**
   - 收货权限验证
   - 库存从在途转入目标仓库
   - ERP接口调用

5. **边界条件测试**
   - 并发操作
   - 库存不足处理
   - 物品状态变更处理
   - 审批流程异常处理

---

## 📚 相关文档

- [物品管理](../material/material_service.md)
- [仓库管理](../warehouse/warehouse_management.md)
- [ERP集成](../erp/erp_integration.md)
- [审批流程设计](../approval/approval_flow.md)

---

## 📄 更新日志

| 日期 | 版本 | 说明 |
|-----|------|-----|
| 2025-01-12 | 1.0 | 初始文档创建 |

---

## 👥 维护者

- 开发团队：Backend Team
- 文档维护：AI Assistant

---

**注意**: 本文档基于代码自动生成，如有代码变更，请及时更新文档。

