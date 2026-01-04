# ERPNext POS Invoice (Grab外卖) 任务分解

> Grab 外卖订单同步 ERPNext 的执行任务清单（main 模块）

## 📊 进度总览

**总任务数**: 5  
**已完成**: 0  
**完成率**: 0%

**说明**：ERPNext 配置、BMP 模块（Protobuf、SavePosInvoice）已由其他同事完成，本任务只涉及 main 模块实现。

---

## Phase 1: Main 模块实现

### Task 1.1: 实现 syncTakeoutOrderToERP 方法

- **File**: `main/app/event/takeout/takeout_order_accept_event_handler.go`（新增方法）
- **Purpose**: 核心同步逻辑
- **Implementation**:
  ```go
  func syncTakeoutOrderToERP(ctx *appContext.Context, event event.OrderAcceptedEvent) error {
      // 1. 检查 ERP 条件
      // 2. 查询 TakeoutOrder
      // 3. 查询 OrderSource
      // 4. 构建请求
      // 5. 调用 ERP Service
  }
  ```
- **Success**: 方法实现完成，编译通过
- **Status**: - [ ]

### Task 1.2: 实现 buildPosInvoiceReqFromTakeoutOrder 方法

- **File**: `main/app/event/takeout/takeout_order_accept_event_handler.go`（新增方法）
- **Purpose**: 从 TakeoutOrder 构建 POS Invoice 请求
- **Key Logic**:
  - 转换 TakeoutOrderItem → PosInvoiceItem
  - 添加配送费商品项（如果 DeliveryFee > 0）
  - 构建税费和支付方式
- **Success**: 方法实现完成，编译通过
- **Status**: - [ ]

### Task 1.3: 在 Event Handler 中调用同步方法

- **File**: `main/app/event/takeout/takeout_order_accept_event_handler.go`
- **Purpose**: 在接单事件处理中触发 ERP 同步
- **Changes**:
  ```go
  func (s *takeoutOrderAcceptEventSubscriber) Handle(domainEvent event.DomainEvent) error {
      // ... 现有逻辑 ...
      
      // 新增：异步同步到 ERP
      utils.Go(func() {
          if err := syncTakeoutOrderToERP(ctx, orderAcceptedEvent); err != nil {
              logger.Logger.Error("同步失败", zap.Error(err))
          }
      })
      
      return nil
  }
  ```
- **Success**: 调用逻辑添加完成
- **Status**: - [ ]

---

## Phase 2: 测试验证

### Task 2.1: 单元测试

- **File**: `main/app/event/takeout/takeout_order_accept_event_handler_test.go`（新建）
- **Purpose**: 测试核心方法
- **Test Cases**:
  - TestBuildPosInvoiceReqFromTakeoutOrder
  - TestBuildPosInvoiceReq_WithDeliveryFee
  - TestBuildPosInvoiceReq_NoDeliveryFee
- **Success**: 所有测试通过
- **Status**: - [ ]

### Task 2.2: 集成测试 - 正常同步

- **Scenario**: Grab 订单接单 → ERP 创建 POS Invoice
- **Steps**:
  1. 创建 Grab 订单（DeliveryFee = 25.00）
  2. 商家接单
  3. 查看 ERPNext POS Invoice
- **Verify**:
  - custom_order_source_name = "Grab"
  - custom_related_order_no = PlatformOrderId
  - Items 包含配送费
- **Status**: - [ ]

### Task 2.3: 集成测试 - 配送费为 0

- **Scenario**: 配送费为 0 时不创建商品项
- **Steps**: 同 Task 5.2，但 DeliveryFee = 0
- **Verify**: Items 不包含配送费
- **Status**: - [ ]

---

## 提交清单

### 代码质量
- [ ] 所有任务完成
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] Protobuf 代码生成成功
- [ ] 编译通过

### 功能完整性
- [ ] requirements.md 需求已满足
- [ ] design.md 设计已实现
- [ ] 验收标准已达成

### 规范遵循
- [ ] 遵循 `.cursor/rules/go-main.mdc`
- [ ] 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`
- [ ] 遵循 `ttpos-bmp/.cursor/rules/proto-rules.mdc`

---

## 进度追踪

```bash
# 查看总任务数（只计算 main 模块任务）
grep -c "^### Task [12]\." docs/shared/specs/active/story-erp-grab-invoice-sync/tasks.md

# 查看已完成任务
grep "Status.*\[x\]" docs/shared/specs/active/story-erp-grab-invoice-sync/tasks.md | wc -l

# 完成率
echo "已完成: $(grep "Status.*\[x\]" docs/shared/specs/active/story-erp-grab-invoice-sync/tasks.md | wc -l) / 5"
```

---

**文档版本**: v2.0  
**最后更新**: 2025-12-29  
**维护者**: weifashi
