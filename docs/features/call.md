# Call Service 呼叫服务说明文档

## 📋 概述

`service/call.go` 是 TTPOS 系统的呼叫服务模块，负责处理餐厅内的顾客呼叫功能，包括服务员呼叫、结账呼叫、异常打印提醒、H5订单提醒和外送订单提醒等。该服务主要用于收银端实时监控和处理各种需要人工介入的业务场景。

**文件路径**: `/home/coder/workspaces/ttpos-server-go/main/app/service/call.go`  
**代码行数**: 313 行  
**接口定义**: `ICallSrv`  
**实现结构**: `callSrv`

---

## 🏗️ 架构设计

### 接口定义 (ICallSrv)

```go
type ICallSrv interface {
    GetUnprocessedCallList(companyUuid uint64, listReq req.UnprocessedCallListReq) (resp.UnprocessedCallList, error)
    GetAbnormalPrintList(companyUuid uint64, soldOutReq req.AbnormalPrintListReq) (resp.AbnormalPrintList, error)
    Processed(companyUuid uint64, callUuid uint64) error
    DeletePrint(companyUuid uint64, printLogUuid uint64) error
    GetUnprocessed(companyUuid uint64) (resp.UnprocessedResp, error)
    GetUnprocessedNotice(ctx context.Context) (resp.UnprocessedListResp, error)
    Call(ctx context.Context, callReq req.CallReq) error
}
```

### 依赖服务

```go
type callSrv struct {
    dbm *database.DBManager  // 数据库管理器
}
```

**依赖说明**:
- 仅依赖数据库管理器，通过 Repository 层访问数据
- 使用国际化服务进行文本翻译

---

## 🎯 核心功能

### 1. 发起呼叫 (Call)

**功能描述**: 平板端顾客发起服务呼叫，请求服务员或结账。

#### 应用场景
- 顾客在平板端点击"呼叫服务员"按钮
- 顾客在平板端点击"呼叫结账"按钮

#### 实现流程

```
1. 获取桌台信息
   - 从上下文获取 deskUuid
   - 查询桌台详情
   ↓
2. 验证桌台是否存在
   ↓
3. 创建呼叫记录
   - 记录桌台UUID
   - 记录桌号
   - 记录呼叫类型
   ↓
4. 返回成功
```

#### 呼叫类型

| 常量 | 值 | 说明 |
|-----|---|------|
| `constant.CallTypeWaiter` | 1 | 呼叫服务员 |
| `constant.CallTypeCheckout` | 2 | 呼叫结账 |

#### 代码示例

```go
func (s *callSrv) Call(ctx context.Context, callReq req.CallReq) error {
    db := s.dbm.GetDB(ctx.GetCompanyUuid())
    
    // 查询桌台信息
    deskRepo := repository.NewDeskRepo(db)
    desk, err := deskRepo.GetDesk(deskRepo.WhereUuid(ctx.GetDeskUuid()))
    if err != nil {
        return errors.New("桌台不存在")
    }
    
    // 创建呼叫记录
    if err := repository.NewCallRepo(db).CreateCall(model.CustomerCall{
        DeskUuid: desk.Uuid,
        DeskNo:   desk.DeskNo,
        CallType: callReq.CallType,
    }); err != nil {
        return err
    }
    
    return nil
}
```

#### 数据模型

```go
type CustomerCall struct {
    Uuid       uint64  // 呼叫UUID
    DeskUuid   uint64  // 桌台UUID
    DeskNo     string  // 桌号
    CallType   uint8   // 呼叫类型
    Status     uint8   // 状态：0-未处理，1-已处理
    CreateTime int64   // 创建时间
}
```

---

### 2. 获取未处理呼叫列表 (GetUnprocessedCallList)

**功能描述**: 分页获取所有未处理的呼叫记录，供收银端查看和处理。

#### 应用场景
- 收银端实时查看未处理的呼叫列表
- 服务员查看需要响应的桌台呼叫

#### 查询条件

```go
// 1. 状态为未处理
callRepo.WhereC1Status(constant.CallStatusUnprocessed)

// 2. 二级呼叫为空（排除关联呼叫）
callRepo.WhereC2IsNull()
```

#### 请求参数

```go
type UnprocessedCallListReq struct {
    PageNo   int  // 页码
    PageSize int  // 每页数量
}
```

#### 响应数据

```go
type UnprocessedCallList struct {
    List []UnprocessedCallItem  // 呼叫列表
    Meta dto.PageResponse       // 分页信息
}

type UnprocessedCallItem struct {
    Uuid       uint64  // 呼叫UUID
    DeskUuid   uint64  // 桌台UUID
    DeskNo     string  // 桌号
    CallType   uint8   // 呼叫类型
    Status     uint8   // 状态
    CreateTime int64   // 创建时间
}
```

#### 代码实现

```go
func (s *callSrv) GetUnprocessedCallList(companyUuid uint64, listReq req.UnprocessedCallListReq) (resp.UnprocessedCallList, error) {
    var res resp.UnprocessedCallList
    callRepo := repository.NewCallRepo(s.dbm.GetDB(companyUuid))
    
    // 分页查询未处理的呼叫
    calls, total, err := callRepo.PaginateGet(
        listReq.PageNo, 
        listReq.PageSize,
        callRepo.WhereC1Status(constant.CallStatusUnprocessed), 
        callRepo.WhereC2IsNull(),
    )
    if err != nil {
        return res, errors.WithMessage(err, "获取呼叫列表失败")
    }
    
    // 数据转换
    callItems := make([]resp.UnprocessedCallItem, 0, len(calls))
    for _, call := range calls {
        var item resp.UnprocessedCallItem
        copier.Copy(&item, call)
        callItems = append(callItems, item)
    }
    
    return resp.UnprocessedCallList{
        List: callItems,
        Meta: dto.PageResponse{
            PageNo:   listReq.PageNo,
            PageSize: listReq.PageSize,
            Total:    total,
        },
    }, nil
}
```

---

### 3. 获取异常打印列表 (GetAbnormalPrintList)

**功能描述**: 分页获取打印失败或异常的打印记录，供收银端重新打印或删除。

#### 应用场景
- 打印机缺纸、卡纸、离线等导致打印失败
- 收银端查看打印失败记录
- 重新打印或删除失败记录

#### 查询条件

```go
// 1. 打印状态为结束（失败）
printerLogRepo.WhereStatus(constant.PrinterLogStatusEnd)

// 2. 打印类型为默认类型
printerLogRepo.WhereType(constant.PrinterLogTypeDefault)

// 3. 预加载关联数据
printerLogRepo.WithPrinter()     // 打印机信息
printerLogRepo.WithSaleOrder()   // 订单信息
printerLogRepo.WithSaleBill()    // 账单信息
```

#### 请求参数

```go
type AbnormalPrintListReq struct {
    PageNo   int  // 页码
    PageSize int  // 每页数量
}
```

#### 响应数据

```go
type AbnormalPrintList struct {
    List []AbnormalPrintItem  // 异常打印列表
    Meta dto.PageResponse     // 分页信息
}

type AbnormalPrintItem struct {
    Uuid        uint64  // 打印日志UUID
    PrinterName string  // 打印机名称
    DeskNo      string  // 桌号/订单号
    CreateTime  int64   // 创建时间
    // ... 其他打印相关字段
}
```

#### 数据处理逻辑

```go
for _, printerLog := range printerLogs {
    var printerName, deskNo string
    
    // 获取打印机名称
    if printerLog.Printer != nil {
        printerName = printerLog.Printer.Name
    }
    
    // 获取桌号/订单号
    if printerLog.SaleBill != nil || printerLog.SaleOrder != nil {
        if printerLog.SaleBill != nil {
            deskNo = printerLog.SaleBill.SerialNo  // 账单流水号
        } else if printerLog.SaleOrder != nil {
            deskNo = printerLog.SaleOrder.SaleBill.SerialNo  // 订单关联账单流水号
        }
    }
    
    // 组装响应数据
    var item resp.AbnormalPrintItem
    copier.Copy(&item, printerLog)
    item.PrinterName = printerName
    item.DeskNo = deskNo
    
    abnormalPrintItems = append(abnormalPrintItems, item)
}
```

---

### 4. 获取未处理消息数量 (GetUnprocessed)

**功能描述**: 统计所有类型的未处理消息数量，用于收银端轮询显示提醒数字。

#### 应用场景
- 收银端顶部显示未处理消息红点数字
- 定时轮询刷新未处理消息数量

#### 统计类型

| 类型 | 说明 | 查询条件 |
|-----|------|---------|
| 未处理呼叫 | 顾客呼叫服务员/结账 | `status=未处理` 且 二级呼叫为空 |
| 异常打印 | 打印失败的记录 | `status=结束` 且 `type=默认` 且 `首次执行=0` |
| 未处理H5订单 | 需要审核的扫码点餐订单 | `status=已下单` 且 `需要审核=1` |
| 待接单外送订单 | 需要商家接单的外送订单 | `status=待商家接单` |

#### 响应数据

```go
type UnprocessedResp struct {
    UnprocessedCallCount        int64  // 未处理呼叫数量
    AbnormalPrintCount          int64  // 异常打印数量
    UnprocessedH5OrderCount     int64  // 未处理H5订单数量
    UnprocessedMemberOrderCount int64  // 未处理外送订单数量
    UpdateTime                  int64  // 更新时间
}
```

#### 代码实现

```go
func (s *callSrv) GetUnprocessed(companyUuid uint64) (resp.UnprocessedResp, error) {
    var (
        res                  resp.UnprocessedResp
        unprocessedCallCount int64
        abnormalPrintCount   int64
        db                   = s.dbm.GetDB(companyUuid)
    )
    
    // 1. 统计未处理呼叫
    callRepo := repository.NewCallRepo(db)
    unprocessedCallCount, err = callRepo.GetUnprocessedCallCount(
        callRepo.WhereC1Status(constant.CallStatusUnprocessed), 
        callRepo.WhereC2IsNull(),
    )
    if err != nil {
        return res, errors.WithMessage(errors.New("获取未处理呼叫数量失败"), err.Error())
    }
    
    // 2. 统计异常打印
    printerLogRepo := repository.NewPrinterLogRepo(db)
    abnormalPrintCount, err = printerLogRepo.GetPrintLogCount(
        printerLogRepo.WhereStatus(constant.PrinterLogStatusEnd),
        printerLogRepo.WhereType(constant.PrinterLogTypeDefault), 
        printerLogRepo.WhereFirstExecution(0),
    )
    if err != nil {
        logger.Logger.Error("获取异常打印数量失败", zap.Error(err))
        return res, errors.WithMessage(errors.New("获取异常打印数量失败"), err.Error())
    }
    
    // 3. 统计未处理H5订单
    h5OrderRepo := repository.NewH5OrderRepo(db)
    unhandledH5OrderCount, err := h5OrderRepo.GetH5OrderCount(
        h5OrderRepo.WhereStatus([]uint{constant.H5OrderStatusOrder}), 
        h5OrderRepo.WhereIsNeedAudit(1),
    )
    if err != nil {
        logger.Logger.Error("获取未处理的h5订单数量失败", zap.Error(err))
        return res, errors.WithMessage(errors.New("获取未处理的h5订单数量失败"), err.Error())
    }
    
    // 4. 统计待接单外送订单
    memberSaleOrderRepo := repository.NewMemberSaleOrderRepo(db)
    unhandledMemberSaleOrderCount, err := memberSaleOrderRepo.GetOrderCount(
        memberSaleOrderRepo.WhereStatusIn([]uint{constant.MemberSaleOrderStatusPendingMerchantAccept}),
    )
    if err != nil {
        logger.Logger.Error("获取待接单外送订单数量失败", zap.Error(err))
        return res, errors.WithMessage(errors.New("获取待接单外送订单数量失败"), err.Error())
    }
    
    return resp.UnprocessedResp{
        UnprocessedCallCount:        unprocessedCallCount,
        AbnormalPrintCount:          abnormalPrintCount,
        UnprocessedH5OrderCount:     unhandledH5OrderCount,
        UnprocessedMemberOrderCount: unhandledMemberSaleOrderCount,
        UpdateTime:                  time.Now().Unix(),
    }, nil
}
```

---

### 5. 获取未处理通知详情 (GetUnprocessedNotice)

**功能描述**: 获取最近30分钟内各类未处理消息的详细列表，用于收银端弹窗提醒或详情展示。

#### 应用场景
- 收银端弹窗显示最新的未处理消息
- 语音播报提醒
- 实时通知服务员

#### 时间范围
```go
thirtyMinutesAgo := time.Now().Add(-30 * time.Minute).Unix()
```
只查询最近30分钟内创建的未处理消息，避免返回过时信息。

#### 查询内容

##### 1. 未处理呼叫 (前10条)

```go
unprocessedCalls, _, err := callRepo.PaginateGet(1, 10,
    callRepo.WhereC1Status(constant.CallStatusUnprocessed),
    callRepo.WhereC1CreateTimeGt(thirtyMinutesAgo),
    callRepo.WhereC2IsNull(),
)
```

**数据处理**:
```go
language := ctx.GetLanguage()
textMap := map[uint8]string{
    constant.CallTypeWaiter:   "呼叫服务员",
    constant.CallTypeCheckout: "呼叫结账",
}

for _, call := range unprocessedCalls {
    var item resp.UnprocessedCallItemForNotice
    copier.Copy(&item, call)
    // 国际化处理：桌位 A01 呼叫服务员
    item.CallText = i18n.Translate(language, "桌位") + " " + 
                    item.DeskNo + " " + 
                    i18n.Translate(language, textMap[item.CallType])
    res.Call.List = append(res.Call.List, item)
}
```

##### 2. 异常打印 (前10条)

```go
abnormalPrints, _, err := printerLogRepo.PaginateGet(1, 10,
    printerLogRepo.WhereStatus(constant.PrinterLogStatusEnd),
    printerLogRepo.WhereType(constant.PrinterLogTypeDefault),
    printerLogRepo.WhereCreateTimeGt(thirtyMinutesAgo),
    printerLogRepo.WithPrinter(),
    printerLogRepo.WithSaleOrder(),
    printerLogRepo.WithSaleBill(),
)
```

**数据处理**:
```go
for _, printerLog := range abnormalPrints {
    var deskNo string
    
    // 优先从账单获取桌号，其次从订单获取
    if printerLog.SaleBill != nil {
        deskNo = printerLog.SaleBill.SerialNo
    } else if printerLog.SaleOrder != nil {
        deskNo = printerLog.SaleOrder.SaleBill.SerialNo
    }
    
    var item resp.AbnormalPrintItemForNotice
    copier.Copy(&item, printerLog)
    item.DeskNo = deskNo
    abnormalPrintItems = append(abnormalPrintItems, item)
}
```

##### 3. 未处理H5订单 (前10条)

```go
orders, _, err := h5OrderRepo.PaginateGetH5Order(1, 10, 
    h5OrderRepo.WhereIsNeedAudit(1),
    h5OrderRepo.WhereUnNotified(),
    h5OrderRepo.WhereCreateTimeGt(thirtyMinutesAgo),
    repository.CommonRepo.SortWithCreateTime("desc"),
)
```

**数据处理**:
```go
for _, order := range orders {
    res.H5Order.List = append(res.H5Order.List, resp.UnprocessedH5OrderItem{
        Uuid:         order.Uuid,
        DeskNo:       order.DeskNo,
        Status:       order.Status,
        IsAutoAccept: order.IsAutoAccept == 1,
    })
}
```

##### 4. 待接单外送订单

```go
memberSaleOrders, err := memberSaleOrderRepo.GetForCall(
    memberSaleOrderRepo.WhereUpdateTimeGt(thirtyMinutesAgo),
)
```

**数据处理**:
```go
for _, memberSaleOrder := range memberSaleOrders {
    res.MemberSaleOrder.List = append(res.MemberSaleOrder.List, resp.UnprocessedMemberSaleOrderItem{
        Uuid:         memberSaleOrder.Uuid,
        Status:       memberSaleOrder.Status,
        UpdateTime:   memberSaleOrder.UpdateTime,
        IsAutoAccept: memberSaleOrder.IsAutoAccept == 1,
        CancelScene:  memberSaleOrder.CancelScene,
        SerialNumber: memberSaleOrder.SerialNumber,
    })
}
```

#### 响应数据结构

```go
type UnprocessedListResp struct {
    Call            UnprocessedCall            // 未处理呼叫
    AbnormalPrint   UnprocessedAbnormalPrint   // 异常打印
    H5Order         UnprocessedH5Order         // H5订单
    MemberSaleOrder UnprocessedMemberSaleOrder // 外送订单
}

type UnprocessedCall struct {
    List []UnprocessedCallItemForNotice  // 呼叫列表
}

type UnprocessedCallItemForNotice struct {
    Uuid       uint64  // 呼叫UUID
    DeskUuid   uint64  // 桌台UUID
    DeskNo     string  // 桌号
    CallType   uint8   // 呼叫类型
    CallText   string  // 呼叫文本（国际化）
    CreateTime int64   // 创建时间
}

type UnprocessedAbnormalPrint struct {
    List []AbnormalPrintItemForNotice  // 异常打印列表
}

type AbnormalPrintItemForNotice struct {
    Uuid       uint64  // 打印日志UUID
    DeskNo     string  // 桌号
    CreateTime int64   // 创建时间
}

type UnprocessedH5Order struct {
    List []UnprocessedH5OrderItem  // H5订单列表
}

type UnprocessedH5OrderItem struct {
    Uuid         uint64  // 订单UUID
    DeskNo       string  // 桌号
    Status       uint    // 状态
    IsAutoAccept bool    // 是否自动接单
}

type UnprocessedMemberSaleOrder struct {
    List []UnprocessedMemberSaleOrderItem  // 外送订单列表
}

type UnprocessedMemberSaleOrderItem struct {
    Uuid         uint64  // 订单UUID
    Status       uint    // 状态
    UpdateTime   int64   // 更新时间
    IsAutoAccept bool    // 是否自动接单
    CancelScene  uint8   // 取消场景
    SerialNumber string  // 订单号
}
```

---

### 6. 处理呼叫 (Processed)

**功能描述**: 标记呼叫为已处理状态，表示服务员已经响应了顾客的呼叫。

#### 应用场景
- 服务员到达桌台后，点击"已处理"按钮
- 收银端批量处理呼叫

#### 处理逻辑

```
1. 根据呼叫UUID查找对应桌台的所有未处理呼叫
   ↓
2. 更新状态为已处理
   - 更新条件：status = 未处理
   - 更新条件：同一桌台
   ↓
3. 返回成功
```

#### 代码实现

```go
func (s *callSrv) Processed(companyUuid uint64, callUuid uint64) error {
    callRepo := repository.NewCallRepo(s.dbm.GetDB(companyUuid))
    
    // 批量更新同一桌台的所有未处理呼叫
    err := callRepo.Update(
        map[string]any{"status": constant.CallStatusProcessed},
        []repository.DBOption{
            callRepo.WhereStatus(constant.CallStatusUnprocessed),
            callRepo.WhereDeskUuidByCallUuid(callUuid),
        },
    )
    
    if err != nil {
        return errors.WithMessage(err, "处理呼叫失败")
    }
    
    return nil
}
```

#### 更新逻辑说明

**为什么通过 callUuid 批量更新同桌台的所有呼叫？**

1. **场景**: 顾客可能在短时间内多次点击呼叫按钮
2. **优化**: 服务员到达后一次性处理该桌台的所有呼叫
3. **实现**: 
   - 通过 `WhereDeskUuidByCallUuid` 查找 callUuid 对应的 deskUuid
   - 更新该 deskUuid 下所有未处理的呼叫

---

### 7. 删除打印记录 (DeletePrint)

**功能描述**: 软删除异常打印记录，清除打印失败的记录。

#### 应用场景
- 打印失败且不需要重新打印
- 收银员确认已手动处理
- 清理异常打印列表

#### 删除方式

采用**软删除**机制：
```go
err := repository.NewPrinterLogRepo(s.dbm.GetDB(companyUuid)).Update(printLogUuid, map[string]any{
    "delete_time": time.Now().Unix(),
})
```

不是物理删除记录，而是设置 `delete_time` 字段，便于后续查询和统计。

#### 代码实现

```go
func (s *callSrv) DeletePrint(companyUuid uint64, printLogUuid uint64) error {
    err := repository.NewPrinterLogRepo(s.dbm.GetDB(companyUuid)).Update(printLogUuid, map[string]any{
        "delete_time": time.Now().Unix(),
    })
    if err != nil {
        return errors.WithMessage(err, "删除打印失败")
    }
    return nil
}
```

---

## 🔄 业务流程

### 顾客呼叫服务流程

```
顾客操作（平板端）
    ↓
点击"呼叫服务员"或"呼叫结账"
    ↓
Call() - 创建呼叫记录
    ↓
收银端实时监控
    ↓
GetUnprocessed() - 轮询未处理数量（显示红点）
    ↓
GetUnprocessedNotice() - 获取详情（弹窗提醒）
    ↓
GetUnprocessedCallList() - 查看呼叫列表
    ↓
服务员响应处理
    ↓
Processed() - 标记已处理
    ↓
呼叫流程结束
```

### 异常打印处理流程

```
打印机打印失败
    ↓
系统记录异常打印日志
    - status: 结束
    - type: 默认
    ↓
收银端监控
    ↓
GetUnprocessed() - 统计异常数量
    ↓
GetAbnormalPrintList() - 查看异常列表
    ↓
收银员处理
    ├── 重新打印 → 成功后自动清除
    └── 手动处理 → DeletePrint() - 删除记录
```

### H5订单提醒流程

```
顾客扫码下单
    ↓
订单状态: 已下单 + 需要审核
    ↓
收银端监控
    ↓
GetUnprocessed() - 统计未处理订单
    ↓
GetUnprocessedNotice() - 获取订单详情
    ↓
收银员处理
    ├── 接单
    └── 拒单
```

### 外送订单提醒流程

```
会员下单（外送）
    ↓
订单状态: 待商家接单
    ↓
收银端监控
    ↓
GetUnprocessed() - 统计待接单数量
    ↓
GetUnprocessedNotice() - 获取订单详情
    ↓
商家处理
    ├── 接单 → 安排配送
    └── 拒单 → 通知顾客
```

---

## 📊 数据统计

### 呼叫状态

| 状态值 | 常量 | 说明 |
|-------|------|-----|
| 0 | `CallStatusUnprocessed` | 未处理 |
| 1 | `CallStatusProcessed` | 已处理 |

### 打印日志状态

| 状态值 | 常量 | 说明 |
|-------|------|-----|
| 0 | `PrinterLogStatusPending` | 待打印 |
| 1 | `PrinterLogStatusSuccess` | 打印成功 |
| 2 | `PrinterLogStatusEnd` | 打印结束（失败） |

### H5订单状态

| 状态值 | 常量 | 说明 |
|-------|------|-----|
| 0 | `H5OrderStatusOrder` | 已下单 |
| 1 | `H5OrderStatusAccept` | 已接单 |
| 2 | `H5OrderStatusReject` | 已拒单 |

### 外送订单状态

| 状态值 | 常量 | 说明 |
|-------|------|-----|
| 1 | `MemberSaleOrderStatusPendingMerchantAccept` | 待商家接单 |
| 2 | `MemberSaleOrderStatusMerchantAccepted` | 商家已接单 |
| 3 | `MemberSaleOrderStatusMerchantRejected` | 商家已拒单 |

---

## 🎪 国际化支持

### 呼叫文本国际化

```go
language := ctx.GetLanguage()

// 呼叫类型文本映射
textMap := map[uint8]string{
    constant.CallTypeWaiter:   "呼叫服务员",
    constant.CallTypeCheckout: "呼叫结账",
}

// 组装国际化文本
item.CallText = i18n.Translate(language, "桌位") + " " + 
                item.DeskNo + " " + 
                i18n.Translate(language, textMap[item.CallType])
```

### 支持的文本

| 中文 | 英文 | 泰文 |
|-----|------|-----|
| 桌位 | Table | โต๊ะ |
| 呼叫服务员 | Call Waiter | เรียกพนักงาน |
| 呼叫结账 | Call for Bill | เรียกเก็บเงิน |

---

## 🚨 错误处理

### 错误类型

| 错误 | 说明 | 处理方式 |
|-----|------|---------|
| 桌台不存在 | 发起呼叫时桌台未找到 | 返回错误提示 |
| 获取呼叫列表失败 | 数据库查询失败 | 记录日志，返回错误 |
| 获取异常打印失败 | 数据库查询失败 | 记录日志，返回错误 |
| 处理呼叫失败 | 更新状态失败 | 记录日志，返回错误 |
| 删除打印失败 | 软删除失败 | 记录日志，返回错误 |

### 错误处理示例

```go
// ✅ 使用自定义错误包装
if err != nil {
    return errors.WithMessage(err, "获取呼叫列表失败")
}

// ✅ 记录错误日志
logger.Logger.Error("获取异常打印数量失败", zap.Error(err))

// ✅ 返回业务错误
if desk.Uuid == 0 {
    return errors.New("桌台不存在")
}
```

---

## 🔧 性能优化

### 1. 分页查询

所有列表查询都使用分页，避免一次性加载大量数据：

```go
calls, total, err := callRepo.PaginateGet(
    listReq.PageNo,    // 页码
    listReq.PageSize,  // 每页数量
    // ... 查询条件
)
```

### 2. 预加载关联数据

异常打印查询时预加载关联数据，避免 N+1 查询：

```go
printerLogRepo.PaginateGet(
    1, 10,
    printerLogRepo.WithPrinter(),      // 预加载打印机
    printerLogRepo.WithSaleOrder(),    // 预加载订单
    printerLogRepo.WithSaleBill(),     // 预加载账单
)
```

### 3. 时间范围限制

通知详情只查询最近30分钟的数据：

```go
thirtyMinutesAgo := time.Now().Add(-30 * time.Minute).Unix()
callRepo.WhereC1CreateTimeGt(thirtyMinutesAgo)
```

### 4. 数量限制

通知详情每个类型最多返回10条：

```go
callRepo.PaginateGet(1, 10, ...)
```

---

## 📝 最佳实践

### 1. 使用 Repository 模式

```go
// ✅ 正确：通过 Repository 访问数据
callRepo := repository.NewCallRepo(db)
calls, total, err := callRepo.PaginateGet(...)

// ❌ 错误：直接使用 db 查询
db.Where("status = ?", 0).Find(&calls)
```

### 2. 数据转换使用 copier

```go
// ✅ 正确：使用 copier 进行结构体复制
var item resp.UnprocessedCallItem
copier.Copy(&item, call)

// ❌ 错误：手动赋值
item.Uuid = call.Uuid
item.DeskNo = call.DeskNo
// ... 繁琐且容易遗漏
```

### 3. 错误处理

```go
// ✅ 正确：包装错误信息
if err != nil {
    return errors.WithMessage(err, "获取呼叫列表失败")
}

// ✅ 正确：记录日志
logger.Logger.Error("获取异常打印数量失败", zap.Error(err))

// ❌ 错误：直接返回原始错误
return err
```

### 4. 国际化处理

```go
// ✅ 正确：使用国际化服务
item.CallText = i18n.Translate(language, "桌位") + " " + item.DeskNo

// ❌ 错误：硬编码中文
item.CallText = "桌位 " + item.DeskNo
```

---

## 🔄 依赖关系

```
callSrv
  └── dbm (database.DBManager)
      ├── CallRepo - 呼叫记录
      ├── DeskRepo - 桌台信息
      ├── PrinterLogRepo - 打印日志
      ├── H5OrderRepo - H5订单
      └── MemberSaleOrderRepo - 外送订单
```

---

## 🧪 测试建议

### 单元测试覆盖

1. **呼叫创建测试**
   - 正常呼叫流程
   - 桌台不存在
   - 不同呼叫类型

2. **列表查询测试**
   - 分页查询
   - 空列表
   - 关联数据加载

3. **数量统计测试**
   - 各类型数量统计
   - 空数据情况
   - 并发统计

4. **处理操作测试**
   - 标记已处理
   - 批量处理
   - 删除打印记录

5. **国际化测试**
   - 中文文本
   - 英文文本
   - 泰文文本

---

## 📊 监控指标

### 关键指标

| 指标 | 说明 | 监控方式 |
|-----|------|---------|
| 平均响应时间 | 每个接口的平均响应时间 | < 200ms |
| 未处理呼叫数量 | 实时未处理呼叫数 | 告警阈值: 10 |
| 异常打印数量 | 打印失败数量 | 告警阈值: 5 |
| 轮询频率 | 收银端轮询频率 | 建议: 5-10秒 |

### 性能要求

```
GetUnprocessed() - < 100ms
GetUnprocessedCallList() - < 200ms
GetAbnormalPrintList() - < 200ms
GetUnprocessedNotice() - < 300ms
Call() - < 50ms
Processed() - < 50ms
DeletePrint() - < 50ms
```

---

## 🔗 相关接口

### API 端点

| 端点 | 方法 | 说明 |
|-----|------|-----|
| `/api/v1/cashier/call/unprocessed/list` | GET | 获取未处理呼叫列表 |
| `/api/v1/cashier/call/abnormal_print/list` | GET | 获取异常打印列表 |
| `/api/v1/cashier/call/unprocessed` | GET | 获取未处理数量 |
| `/api/v1/cashier/call/unprocessed/notice` | GET | 获取未处理通知详情 |
| `/api/v1/cashier/call/processed` | POST | 标记呼叫已处理 |
| `/api/v1/cashier/call/print/delete` | DELETE | 删除打印记录 |
| `/api/v1/tablet/call` | POST | 平板端发起呼叫 |

---

## 📚 相关文档

- [桌台管理](./desk_service.md)
- [打印服务](./printer_service.md)
- [H5订单管理](./h5_order_service.md)
- [外送订单管理](./member_sale_order_service.md)
- [国际化服务](../i18n/i18n.md)

---

## 📄 更新日志

| 日期 | 版本 | 说明 |
|-----|------|-----|
| 2025-11-12 | 1.0 | 初始文档创建 |

---

## 👥 维护者

- 开发团队：Backend Team
- 文档维护：AI Assistant

---

**注意**: 本文档基于代码自动生成，如有代码变更，请及时更新文档。

