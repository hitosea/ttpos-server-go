# ERPNext POS Invoice (Grab外卖) 技术设计文档

> Grab 外卖订单接单后直接同步到 ERP

## 📋 基本信息

| 项目 | 内容 |
|------|------|
| **关联需求** | [requirements.md](./requirements.md) |
| **创建日期** | 2025-12-29 |
| **负责人** | weifashi |

---

## 🎯 设计目标

1. **核心**：实现 Grab 订单接单后直接从 TakeoutOrder 同步到 ERP
2. **架构**：尊重 TakeoutOrder 独立架构，不创建 SaleBill
3. **兼容**：不影响现有店内订单和外送订单的 ERP 同步
4. **扩展**：支持未来接入其他外卖平台（LINE MAN、Foodpanda）

**实施范围**：本设计只涉及 **main 模块**实现，ERPNext 配置和 BMP 模块已由其他同事完成。

---

## 🏗️ 架构概览

```
Grab Platform
  ↓ webhook
TTPOS BMP 接收订单 → 创建 TakeoutOrder
  ↓
商家接单 → OrderAcceptedEvent
  ↓
TakeoutOrderAcceptEventHandler
  ├─ 现有逻辑（出库、送厨、打印）
  └─ ✨ 新增：同步到 ERP
       ↓
       查询 TakeoutOrder + OrderSource
       ↓
       构建 SavePosInvoiceReq
       ↓
       调用 BMP gRPC
       ↓
       BMP 调用 ERPNext API
       ↓
       创建 POS Invoice
```

---

## 📊 数据模型

### TakeoutOrder Model（现有）

```go
// main/app/modules/takeout/domain/model/takeout_order.go
type TakeoutOrder struct {
    Uuid             uint64
    TakeoutOrderUuid string  // Grab 平台订单UUID
    Platform         string  // "grab"
    PlatformOrderId  string  // Grab 订单ID
    ShortOrderNumber string  // 短订单号
    DeliveryFee      float64 // 配送费
    EaterPayment     float64 // 实付金额
    Tax              float64 // 税费
    PaymentType      string  // "CASH" / "ONLINE"
    AcceptedTime     int64   // 接单时间
    
    TakeoutOrderItems []TakeoutOrderItem
}
```

### Protobuf 扩展（已完成）

✅ BMP 模块的 Protobuf 已扩展（其他同事已完成）：

```protobuf
// ttpos-bmp/app/ttpos-erp/manifest/protobuf/selling/selling.proto
message SavePosInvoiceReq {
    // ... 现有字段 ...
    
    // 已添加的新字段
    optional string order_source_name = 17;    // "Grab"
    optional string related_order_no = 18;     // PlatformOrderId
    optional string related_order_type = 19;   // "grab"
    optional bool is_takeout_order = 20;       // true
}
```

---

## 🔧 核心实现

### 1. Event Handler 扩展

**文件**：`main/app/event/takeout/takeout_order_accept_event_handler.go`

```go
func (s *takeoutOrderAcceptEventSubscriber) Handle(domainEvent event.DomainEvent) error {
    orderAcceptedEvent, ok := domainEvent.(event.OrderAcceptedEvent)
    if !ok {
        return nil
    }
    
    // ... 现有逻辑（出库、送厨、打印）...
    
    // ✨ 新增：异步同步到 ERP
    utils.Go(func() {
        if err := syncTakeoutOrderToERP(ctx, orderAcceptedEvent); err != nil {
            logger.Logger.Error("同步 Grab 订单到 ERP 失败",
                zap.Uint64("orderUuid", orderAcceptedEvent.OrderUuid),
                zap.Error(err))
        }
    })
    
    return nil
}
```

### 2. 同步方法实现

```go
func syncTakeoutOrderToERP(ctx *appContext.Context, event event.OrderAcceptedEvent) error {
    db := ctx.GetDB()
    
    // 1. 检查 ERP 同步条件
    company, _ := repository.NewCompanyRepo(db).GetCompanyByUuid(event.CompanyUuid)
    companySetting, _ := repository.NewCompanySettingRepo(db).GetSettingByCompanyUuid(event.CompanyUuid)
    
    if !company.IsOpenErpPhase3() || companySetting.ErpnextSiteCode == "" {
        return nil // 未开启 ERP，直接返回
    }
    
    // 2. 查询 TakeoutOrder 完整信息
    takeoutOrderRepo := persistence.NewTakeoutOrderRepo(db)
    takeoutOrder, err := takeoutOrderRepo.GetByUuid(
        event.OrderUuid,
        takeoutOrderRepo.WithPreload(persistence.DBOption(func(db *gorm.DB) *gorm.DB {
            return db.Preload("TakeoutOrderItems").
                Preload("TakeoutOrderItems.TakeoutOrderItemModifiers")
        })),
    )
    if err != nil {
        return errors.WithMessage(err, "查询 TakeoutOrder 失败")
    }
    
    // 3. 查询 OrderSource（Grab）
    orderSource, err := repository.NewOrderSourceRepo(db).GetOrderSourceByName("Grab")
    if err != nil {
        logger.Logger.Warn("查询 OrderSource 失败，使用默认值", zap.Error(err))
    }
    
    // 4. 构建请求并调用 ERP Service
    erpSrv := rpc.NewErpService()
    req := buildPosInvoiceReqFromTakeoutOrder(takeoutOrder, orderSource, companySetting)
    
    res, err := erpSrv.SavePosInvoice(ctx, req)
    if err != nil {
        return errors.WithMessage(err, "调用 ERP Service 失败")
    }
    
    logger.Logger.Info("Grab 订单同步到 ERP 成功",
        zap.Uint64("orderUuid", takeoutOrder.Uuid),
        zap.String("platformOrderId", takeoutOrder.PlatformOrderId),
        zap.String("erpInvoiceName", res.ProductsInvoiceName))
    
    return nil
}
```

### 3. 构建 POS Invoice 请求

```go
func buildPosInvoiceReqFromTakeoutOrder(
    order *model.TakeoutOrder,
    orderSource *model.OrderSource,
    companySetting *model.CompanySetting,
) *dto.SavePosInvoiceReq {
    
    // 1. 构建商品列表
    items := make([]*dto.PosInvoiceItem, 0, len(order.TakeoutOrderItems)+1)
    
    for _, item := range order.TakeoutOrderItems {
        items = append(items, &dto.PosInvoiceItem{
            ItemCode:  item.ItemCode,
            ItemName:  item.ItemName,
            Qty:       float64(item.Quantity),
            Rate:      item.Price,
            Amount:    item.TotalPrice,
        })
    }
    
    // 2. 添加配送费商品项（如果有）
    if order.DeliveryFee > 0 {
        items = append(items, &dto.PosInvoiceItem{
            ItemCode: "DELIVERY_FEE",
            ItemName: "配送费",
            Qty:      1,
            Rate:     order.DeliveryFee,
            Amount:   order.DeliveryFee,
        })
    }
    
    // 3. 构建税费
    taxes := make([]*dto.PosInvoiceTax, 0)
    if order.Tax > 0 {
        taxes = append(taxes, &dto.PosInvoiceTax{
            ChargeType:  "On Net Total",
            AccountHead: "VAT - " + companySetting.ErpnextCompanyAbbr,
            Rate:        7.0,
            TaxAmount:   order.Tax,
        })
    }
    
    // 4. 构建支付方式
    payments := []*dto.PosInvoicePayment{
        {
            Mode:   mapPaymentType(order.PaymentType),
            Amount: order.EaterPayment,
        },
    }
    
    // 5. 订单来源名称
    orderSourceName := "Grab"
    if orderSource != nil {
        orderSourceName = orderSource.MultiLanguageName.GetName("zh-CN")
    }
    
    return &dto.SavePosInvoiceReq{
        OrderNo:           order.PlatformOrderId,
        CompanyAbbr:       companySetting.ErpnextCompanyAbbr,
        PostingDatetime:   time.Unix(order.AcceptedTime, 0).Format("2006-01-02 15:04:05"),
        Currency:          "THB",
        Branch:            companySetting.ErpnextBranchName,
        Items:             items,
        Taxes:             taxes,
        Payments:          payments,
        Remark:            fmt.Sprintf("Grab 外卖订单: %s", order.ShortOrderNumber),
        
        // 新增字段
        OrderSourceName:   orderSourceName,
        RelatedOrderNo:    order.PlatformOrderId,
        RelatedOrderType:  order.Platform,
        IsTakeoutOrder:    true,
    }
}

func mapPaymentType(paymentType string) string {
    switch paymentType {
    case "CASH":
        return "现金"
    case "ONLINE":
        return "Grab 在线支付"
    default:
        return "Grab 支付"
    }
}
```

---

## 🔧 BMP Module 扩展（已完成）

✅ BMP 模块的 `SavePosInvoice` 方法已扩展（其他同事已完成）

**文件**：`ttpos-bmp/app/ttpos-erp/internal/logic/selling/selling.go`

```go
func (s *sSelling) SavePosInvoice(ctx context.Context, req *selling.SavePosInvoiceReq) (*selling.SavePosInvoiceRes, error) {
    // ... 现有逻辑 ...
    
    // ✅ 已实现：处理外卖平台订单的自定义字段
    if req.IsTakeoutOrder {
        if req.OrderSourceName != "" {
            invoiceData["custom_order_source_name"] = req.OrderSourceName
        }
        if req.RelatedOrderNo != "" {
            invoiceData["custom_related_order_no"] = req.RelatedOrderNo
            invoiceData["custom_related_order_type"] = req.RelatedOrderType
        }
    }
    
    // ... 调用 ERPNext API ...
}
```

**我们只需要在 main 模块中正确调用此方法即可。**

---

## 🚨 异常处理

### 1. ERP 同步失败不阻塞流程

```go
func syncTakeoutOrderToERP(ctx *appContext.Context, event event.OrderAcceptedEvent) error {
    // ... 同步逻辑 ...
    
    if err := erpSrv.SavePosInvoice(ctx, req); err != nil {
        logger.Logger.Error("同步失败", zap.Error(err))
        // ⚠️ 不返回错误，不阻塞订单流程
        return nil
    }
    
    return nil
}
```

### 2. 超时控制

```go
func syncTakeoutOrderToERP(ctx *appContext.Context, event event.OrderAcceptedEvent) error {
    ctxWithTimeout, cancel := context.WithTimeout(ctx.GetContext(), 30*time.Second)
    defer cancel()
    
    ctx = ctx.WithContext(ctxWithTimeout)
    // ... 同步逻辑 ...
}
```

### 3. 数据验证

```go
func validateTakeoutOrderForERP(order *model.TakeoutOrder) error {
    if order.Platform != "grab" {
        return errors.New("只支持 Grab 平台")
    }
    if order.PlatformOrderId == "" {
        return errors.New("缺少 PlatformOrderId")
    }
    if len(order.TakeoutOrderItems) == 0 {
        return errors.New("商品列表为空")
    }
    return nil
}
```

---

## 🧪 测试策略

### 单元测试

```go
func TestBuildPosInvoiceReqFromTakeoutOrder(t *testing.T) {
    order := &model.TakeoutOrder{
        Platform:        "grab",
        PlatformOrderId: "T-TEST123",
        DeliveryFee:     25.00,
        TakeoutOrderItems: []model.TakeoutOrderItem{
            {ItemName: "商品A", Quantity: 2, Price: 50.00},
        },
    }
    
    req := buildPosInvoiceReqFromTakeoutOrder(order, nil, nil)
    
    assert.Equal(t, "T-TEST123", req.OrderNo)
    assert.Equal(t, 2, len(req.Items)) // 商品 + 配送费
    assert.Equal(t, "配送费", req.Items[1].ItemName)
}
```

### 集成测试场景

| 场景 | 前置条件 | 预期结果 |
|-----|---------|---------|
| **正常同步** | 公司已开启 ERP | ERP 创建 POS Invoice，包含订单来源、Grab 订单号、配送费 |
| **配送费为 0** | 订单无配送费 | ERP 创建 POS Invoice，不包含配送费商品项 |
| **未开启 ERP** | 公司未开启 ERP | 不触发同步，其他流程正常 |
| **ERP 异常** | ERP 服务不可用 | 记录日志，其他流程不受影响 |

---

## 📈 性能优化

1. **异步处理**：ERP 同步在独立 goroutine 中执行
2. **超时控制**：30 秒超时
3. **查询优化**：使用 Preload 避免 N+1 查询

---

## 🔒 安全性

1. **数据验证**：验证 Platform、PlatformOrderId、金额
2. **权限控制**：只有开启 ERP Phase 3 的公司才能同步
3. **敏感信息**：日志不记录顾客手机号等敏感信息

---

## 📝 相关文档

- **需求文档**：[requirements.md](./requirements.md)
- **任务拆解**：[tasks.md](./tasks.md)
- **Go Main 规范**：`.cursor/rules/go-main.mdc`
- **Protobuf 规范**：`ttpos-bmp/.cursor/rules/proto-rules.mdc`

---

**文档版本**: v2.0  
**最后更新**: 2025-12-29  
**维护者**: weifashi
