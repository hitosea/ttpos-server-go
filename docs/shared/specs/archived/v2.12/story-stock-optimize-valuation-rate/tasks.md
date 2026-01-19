# 优化库存盘点估值率逻辑 任务分解

> 本文档定义 优化库存盘点估值率逻辑 的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 12  
**已完成**: 8 (核心功能已完成)  
**待完成**: Phase 4 测试 (需单独执行), Phase 5.2-5.3 (性能测试和文档更新)  
**完成率**: 67%

---

## Phase 1: Protobuf 定义和代码生成

### 任务说明

本阶段定义 GetBin gRPC 接口的 Protobuf 消息，并生成 Go 代码。

---

- [x] 1.1 定义 Protobuf 消息 - GetBin 接口

  - File: `ttpos-bmp/app/ttpos-erp/manifest/protobuf/stock/stock.proto`
  - Purpose: 定义 GetBin gRPC 接口的请求和响应消息
  - Requirements: Requirement 1.1
  - Leverage: 现有 Stock Protobuf 定义: `ttpos-bmp/app/ttpos-erp/manifest/protobuf/stock/stock.proto`
  - Prompt: 
    ```
    Role: gRPC Developer with Protobuf expertise
    
    Task: 在 stock.proto 中新增 GetBin 接口定义
    
    Context:
    - File: ttpos-bmp/app/ttpos-erp/manifest/protobuf/stock/stock.proto
    - Leverage: 现有 Stock service 定义
    - Requirements: Requirement 1.1 - 定义 Protobuf 消息
    
    Interface to Add:
    ```protobuf
    service Stock {
      // 现有方法...
      
      // 查询物品在指定仓库的 Bin 记录
      rpc GetBin (GetBinReq) returns (GetBinResp);
    }
    
    message GetBinReq {
      string item_code = 1;      // 物品代码
      string warehouse = 2;      // 仓库名称
      string company_abbr = 3;   // 公司简称
    }
    
    message GetBinResp {
      int32 code = 1;
      string message = 2;
      BinData data = 3;
    }
    
    message BinData {
      string item_code = 1;
      string warehouse = 2;
      double actual_qty = 3;
      double valuation_rate = 4;
      double stock_value = 5;
    }
    ```
    
    Restrictions:
    - 使用 proto3 语法
    - 遵循现有命名规范
    - 字段编号不与现有消息冲突
    
    Success Criteria:
    - Protobuf 定义完整
    - 字段命名规范
    - 编译无错误
    ```

- [x] 1.2 生成 gRPC Go 代码

  - File: -
  - Purpose: 根据 Protobuf 定义生成 Go 代码
  - Requirements: Requirement 1.2
  - Leverage: Task 1.1 的 Protobuf 定义
  - Command: 
    ```bash
    cd ttpos-bmp/app/ttpos-erp
    gf gen pb
    ```
  - Success: 代码生成成功，`api/stock/stock.pb.go` 已更新

- [x] 1.3 验证生成的代码

  - File: `ttpos-bmp/app/ttpos-erp/api/stock/stock.pb.go`
  - Purpose: 确认生成的 Go 代码正确
  - Requirements: Requirement 1.2
  - Leverage: Task 1.2 生成的代码
  - Success: 
    - GetBinReq, GetBinResp, BinData 结构体已生成
    - GetBin 方法签名正确
    - 编译无错误

---

## Phase 2: Logic 层实现

### 任务说明

本阶段实现 Logic 层的业务逻辑，包括 GetBin 方法和现有方法的修改。

---

- [x] 2.1 创建 stock_bin.go 文件并实现 GetBin Logic 方法

  - File: `ttpos-bmp/app/ttpos-erp/internal/logic/stock/stock_bin.go`（新建文件）
  - Purpose: 独立 Bin 表查询逻辑，实现 GetBin 方法
  - Requirements: Requirement 1.3, 1.4, 1.5, 1.6
  - Leverage: 
    - 现有 Stock Logic 结构: `ttpos-bmp/app/ttpos-erp/internal/logic/stock/stock.go`（参考文件结构）
    - Document Service: `ttpos-bmp/app/ttpos-erp/internal/service/document.go`
    - Company Service: `ttpos-bmp/app/ttpos-erp/internal/service/company.go`
  - Prompt:
    ```
    Role: Go Developer with GoFrame and ERPNext expertise
    
    Task: 创建独立的 stock_bin.go 文件，实现 GetBin 方法查询 ERPNext Bin 表
    
    Context:
    - File: ttpos-bmp/app/ttpos-erp/internal/logic/stock/stock_bin.go（新建文件）
    - Leverage: 
      - stock.go 中的 sStock 结构体（使用相同的接收者）
      - service.Document().List() 查询 ERPNext
      - service.Company().GetCompanyWithAbbr() 获取公司
    - Requirements: Requirement 1 - 新增 Bin 查询服务
    
    File Structure:
    ```go
    package stock
    
    import (
        "context"
        "ttpos-bmp/app/ttpos-erp/api/stock"
        "ttpos-bmp/app/ttpos-erp/internal/model/dto/erp"
        "ttpos-bmp/app/ttpos-erp/internal/service"
        
        "github.com/gogf/gf/v2/errors/gerror"
        "github.com/gogf/gf/v2/frame/g"
    )
    
    // GetBin 查询物品在指定仓库的 Bin 记录
    func (s *sStock) GetBin(ctx context.Context, req *stock.GetBinReq) (*stock.GetBinResp, error) {
        // 实现逻辑...
    }
    ```
    
    Implementation:
    1. 调用 service.Company().GetCompanyWithAbbr() 获取公司信息
    2. 调用 service.Document().List() 查询 Bin DocType
       - Filters: item_code, warehouse
       - Fields: item_code, warehouse, actual_qty, valuation_rate, stock_value
       - Limit: 1
    3. 解析响应，构建 BinData
    4. 若无数据，返回估值率为 0 的空数据
    5. 记录日志（查询成功、失败、无数据）
    
    Restrictions:
    - 使用 gerror.Wrapf() 包装错误
    - 使用 g.Log() 记录日志
    - 遵循 .cursor/rules/go-bmp.mdc
    
    Success Criteria:
    - GetBin 方法实现完整
    - 错误处理正确
    - 日志记录清晰
    - 无编译错误
    ```

- [x] 2.2 修改 SaveStockReconciliation 方法

  - File: `ttpos-bmp/app/ttpos-erp/internal/logic/stock/stock_reconciliation.go`
  - Purpose: 集成 GetBin 查询，移除强制赋值估值率为 1 的逻辑
  - Requirements: Requirement 2.1, 2.2, 2.3, 2.4, 2.5, 2.6
  - Leverage: 
    - Task 2.1 的 GetBin 方法
    - 现有 SaveStockReconciliation 方法
  - Prompt:
    ```
    Role: Go Developer with business logic expertise
    
    Task: 修改 SaveStockReconciliation 方法，集成 GetBin 查询估值率
    
    Context:
    - File: ttpos-bmp/app/ttpos-erp/internal/logic/stock/stock_reconciliation.go
    - Leverage: 
      - Task 2.1 的 GetBin 方法
      - 现有 SaveStockReconciliation 方法（第 17-138 行）
    - Requirements: Requirement 2 - 调整盘点保存逻辑
    
    Modification:
    在构建 itemList 的循环中（第 68-114 行），修改估值率处理逻辑：
    
    ```go
    // 当前代码（第 106-111 行）
    if item.ValuationRate > 0 {
        itemData.ValuationRate = item.ValuationRate
    } else {
        //默认估值率1
        itemData.ValuationRate = consts.DefaultValuationRate
    }
    
    // 修改为：
    if item.ValuationRate > 0 {
        // 用户提供了估值率，直接使用
        itemData.ValuationRate = item.ValuationRate
        g.Log().Infof(ctx, "使用用户提供的估值率: item_code=%s, valuation_rate=%.2f", 
            item.ItemCode, item.ValuationRate)
    } else {
        // 估值率为 0，从 Bin 表查询
        binResp, err := s.GetBin(ctx, &stock.GetBinReq{
            ItemCode:    item.ItemCode,
            Warehouse:   itemData.Warehouse,
            CompanyAbbr: req.CompanyAbbr,
        })
    
        if err == nil && binResp.Code == 1 && binResp.Data.ValuationRate > 0 {
            // 使用 Bin 表中的估值率
            itemData.ValuationRate = binResp.Data.ValuationRate
            g.Log().Infof(ctx, "从 Bin 表获取估值率: item_code=%s, warehouse=%s, valuation_rate=%.2f", 
                item.ItemCode, itemData.Warehouse, binResp.Data.ValuationRate)
        } else {
            // Bin 表中没有估值率，使用保底值 1
            itemData.ValuationRate = consts.DefaultValuationRate
            g.Log().Warningf(ctx, "Bin 表中无估值率，使用保底值 1: item_code=%s, warehouse=%s", 
                item.ItemCode, itemData.Warehouse)
        }
    }
    ```
    
    Restrictions:
    - 保持现有逻辑不变（库存一致性检查等）
    - 使用 g.Log() 记录估值率来源
    - 遵循 .cursor/rules/go-bmp.mdc
    
    Success Criteria:
    - 估值率查询逻辑正确
    - 日志记录清晰
    - 保底逻辑保留
    - 无编译错误
    ```

- [x] 2.3 修改 SubmitStockReconciliation 方法

  - File: `ttpos-bmp/app/ttpos-erp/internal/logic/stock/stock_reconciliation.go`
  - Purpose: 增加估值率验证逻辑，阻止估值率为空时提交
  - Requirements: Requirement 3.1, 3.2, 3.3, 3.4, 3.5
  - Leverage: 现有 SubmitStockReconciliation 方法
  - Prompt:
    ```
    Role: Go Developer with validation expertise
    
    Task: 修改 SubmitStockReconciliation 方法，增加估值率验证
    
    Context:
    - File: ttpos-bmp/app/ttpos-erp/internal/logic/stock/stock_reconciliation.go
    - Leverage: 现有 SubmitStockReconciliation 方法（第 261-283 行）
    - Requirements: Requirement 3 - 盘点提交时验证估值率
    
    Modification:
    在提交前（第 274 行 service.Document().ChangeDocStatus() 之前）增加验证逻辑：
    
    ```go
    // 查询盘点单详情
    resp, err := service.Document().Get(ctx, &erp.ErpReq{
        DocType: erp.DocTypeStockReconciliation,
        Name:    req.StockReconciliationName,
    }, nil)
    if err != nil {
        return nil, gerror.Wrapf(err, "查询盘点单详情失败")
    }
    
    // 验证估值率
    j := resp
    itemsArray := j.GetJsons("data.items")
    invalidItems := make([]string, 0)
    
    for _, itemData := range itemsArray {
        itemCode := itemData.Get("item_code").String()
        warehouse := itemData.Get("warehouse").String()
        valuationRate := itemData.Get("valuation_rate").Float64()
    
        // 检查估值率是否为 0 或 1（保底值）
        if valuationRate == 0 || valuationRate == 1 {
            invalidItems = append(invalidItems, 
                fmt.Sprintf("物品 [%s] 在仓库 [%s] 的估值率为空（%.2f）", 
                    itemCode, warehouse, valuationRate))
        }
    }
    
    if len(invalidItems) > 0 {
        errMsg := fmt.Sprintf("盘点单中有物品估值率为空，无法提交。请先通过采购入库建立库存。\n详情：\n%s", 
            strings.Join(invalidItems, "\n"))
        g.Log().Warning(ctx, errMsg)
        return nil, gerror.New(errMsg)
    }
    
    g.Log().Infof(ctx, "估值率验证通过，提交盘点单: %s", req.StockReconciliationName)
    ```
    
    Restrictions:
    - 在提交前验证
    - 错误信息清晰
    - 遵循 .cursor/rules/go-bmp.mdc
    
    Success Criteria:
    - 验证逻辑正确
    - 错误提示清晰
    - 日志记录完整
    - 无编译错误
    ```

---

## Phase 3: Controller 层实现

### 任务说明

本阶段实现 RPC Controller 层，暴露 GetBin gRPC 接口。

---

- [x] 3.1 实现 GetBin RPC Controller

  - File: `ttpos-bmp/app/ttpos-erp/internal/controller/rpc/stock/stock.go`
  - Purpose: 实现 GetBin gRPC 接口，调用 Logic 层
  - Requirements: Requirement 1
  - Leverage: 
    - Task 2.1 的 GetBin Logic 方法
    - 现有 Stock RPC Controller
  - Prompt:
    ```
    Role: Go Developer with gRPC expertise
    
    Task: 在 Stock RPC Controller 中实现 GetBin 方法
    
    Context:
    - File: ttpos-bmp/app/ttpos-erp/internal/controller/rpc/stock/stock.go
    - Leverage: 
      - Task 2.1 的 Logic GetBin 方法
      - 现有 Stock Controller 方法
    - Requirements: Requirement 1 - 新增 Bin 查询服务
    
    Implementation:
    ```go
    // GetBin 查询 Bin 记录
    func (c *Controller) GetBin(ctx context.Context, req *stock.GetBinReq) (*stock.GetBinResp, error) {
        // 参数验证
        if req.ItemCode == "" || req.Warehouse == "" || req.CompanyAbbr == "" {
            return &stock.GetBinResp{
                Code:    0,
                Message: "参数错误：item_code, warehouse, company_abbr 为必填项",
                Data:    &stock.BinData{},
            }, nil
        }
    
        // 调用 Logic 层
        bin, err := service.Stock().GetBin(ctx, req)
        if err != nil {
            g.Log().Error(ctx, "GetBin 失败", err)
            return &stock.GetBinResp{
                Code:    0,
                Message: fmt.Sprintf("查询 Bin 记录失败: %v", err),
                Data:    &stock.BinData{},
            }, nil
        }
    
        return bin, nil
    }
    ```
    
    Restrictions:
    - 参数验证在 Controller 层
    - 使用 service.Stock() 调用 Logic
    - 错误统一返回 code=0
    - 遵循 .cursor/rules/go-bmp.mdc
    
    Success Criteria:
    - GetBin 方法实现完整
    - 参数验证正确
    - 错误处理统一
    - 无编译错误
    ```

---

## Phase 4: 测试实现

### 任务说明

本阶段编写单元测试和集成测试，确保功能正确性。

---

- [ ] 4.1 编写 GetBin Logic 单元测试

  - File: `ttpos-bmp/app/ttpos-erp/internal/logic/stock/stock_bin_test.go`（新建文件）
  - Purpose: 测试 GetBin Logic 方法的各种场景
  - Requirements: Requirement 1
  - Leverage: Task 2.1 的 GetBin 实现
  - Prompt:
    ```
    Role: QA Engineer with Go testing expertise
    
    Task: 为 GetBin Logic 方法编写单元测试
    
    Context:
    - File: ttpos-bmp/app/ttpos-erp/internal/logic/stock/stock_bin_test.go（新建文件）
    - Target: Task 2.1 的 GetBin 方法（stock_bin.go）
    - Coverage target: ≥ 70%
    
    Test Cases:
    1. TestGetBin_Success - 成功查询 Bin 记录
    2. TestGetBin_NoRecord - Bin 表中无记录
    3. TestGetBin_ERPNextError - ERPNext API 调用失败
    4. TestGetBin_InvalidParams - 参数验证失败
    
    Restrictions:
    - 使用 Mock Document Service
    - 测试所有分支
    - 遵循 .cursor/rules/go-bmp.mdc
    
    Success Criteria:
    - 测试覆盖率 ≥ 70%
    - 所有测试通过
    - 边界情况已覆盖
    ```

- [ ] 4.2 编写 SaveStockReconciliation 修改测试

  - File: `ttpos-bmp/app/ttpos-erp/internal/logic/stock/stock_reconciliation_test.go`
  - Purpose: 测试 SaveStockReconciliation 修改后的估值率查询逻辑
  - Requirements: Requirement 2
  - Leverage: Task 2.2 的修改
  - Prompt:
    ```
    Role: QA Engineer with Go testing expertise
    
    Task: 为 SaveStockReconciliation 修改编写测试
    
    Context:
    - File: ttpos-bmp/app/ttpos-erp/internal/logic/stock/stock_reconciliation_test.go
    - Target: Task 2.2 的 SaveStockReconciliation 修改
    - Coverage target: 100%（Stock 相关高风险）
    
    Test Cases:
    1. TestSaveStockReconciliation_WithUserProvidedValuationRate - 使用用户提供的估值率
    2. TestSaveStockReconciliation_WithBinValuationRate - 从 Bin 表获取估值率
    3. TestSaveStockReconciliation_WithDefaultValuationRate - 使用保底估值率 1
    4. TestSaveStockReconciliation_BinQueryError - Bin 查询失败，使用保底值
    
    Restrictions:
    - 使用 Mock GetBin 方法
    - 测试所有分支
    - 遵循 .cursor/rules/go-bmp.mdc
    
    Success Criteria:
    - 测试覆盖率 100%
    - 所有测试通过
    - 日志验证正确
    ```

- [ ] 4.3 编写 SubmitStockReconciliation 修改测试

  - File: `ttpos-bmp/app/ttpos-erp/internal/logic/stock/stock_reconciliation_test.go`
  - Purpose: 测试 SubmitStockReconciliation 修改后的估值率验证逻辑
  - Requirements: Requirement 3
  - Leverage: Task 2.3 的修改
  - Prompt:
    ```
    Role: QA Engineer with Go testing expertise
    
    Task: 为 SubmitStockReconciliation 修改编写测试
    
    Context:
    - File: ttpos-bmp/app/ttpos-erp/internal/logic/stock/stock_reconciliation_test.go
    - Target: Task 2.3 的 SubmitStockReconciliation 修改
    - Coverage target: 100%（Stock 相关高风险）
    
    Test Cases:
    1. TestSubmitStockReconciliation_ValidValuationRate - 估值率有效，正常提交
    2. TestSubmitStockReconciliation_ValuationRateZero - 估值率为 0，阻止提交
    3. TestSubmitStockReconciliation_ValuationRateOne - 估值率为 1（保底值），阻止提交
    4. TestSubmitStockReconciliation_MixedValuationRates - 部分物品估值率为空
    5. TestSubmitStockReconciliation_QueryError - 查询盘点单详情失败
    
    Restrictions:
    - 使用 Mock Document Service
    - 测试所有分支
    - 验证错误信息格式
    - 遵循 .cursor/rules/go-bmp.mdc
    
    Success Criteria:
    - 测试覆盖率 100%
    - 所有测试通过
    - 错误提示验证正确
    ```

- [ ] 4.4 执行集成测试

  - File: -
  - Purpose: 测试端到端盘点流程
  - Requirements: 所有功能需求
  - Leverage: Task 2.1, 2.2, 2.3, 3.1 的实现
  - Test Scenarios:
    1. **正常流程**:
       - 调用 GetBin 查询估值率 → 成功返回
       - 调用 SaveStockReconciliation 保存盘点单 → 使用 Bin 估值率
       - 调用 SubmitStockReconciliation 提交盘点单 → 验证通过，提交成功
    
    2. **Bin 无数据流程**:
       - 调用 GetBin 查询估值率 → 返回估值率 0
       - 调用 SaveStockReconciliation 保存盘点单 → 使用保底值 1
       - 调用 SubmitStockReconciliation 提交盘点单 → 验证失败，阻止提交
    
    3. **ERPNext 异常流程**:
       - 调用 GetBin 查询估值率 → API 调用失败
       - 调用 SaveStockReconciliation 保存盘点单 → 使用保底值 1
       - 调用 SubmitStockReconciliation 提交盘点单 → 验证失败，阻止提交
  
  - Tools: grpcurl, Postman
  - Success: 所有场景测试通过

---

## Phase 5: 代码质量和优化

### 任务说明

本阶段进行代码格式化、静态检查、性能测试和文档更新。

---

- [x] 5.1 代码格式化和静态检查

  - File: `ttpos-bmp/app/ttpos-erp/internal/`
  - Purpose: 确保代码质量
  - Requirements: 代码质量要求
  - Commands:
    ```bash
    # Go 格式化
    cd ttpos-bmp/app/ttpos-erp
    go fmt ./...
    
    # Go 静态检查
    go vet ./...
    
    # 编译检查
    go build
    ```
  - Success: 所有检查通过，无错误和警告

- [ ] 5.2 性能测试

  - File: -
  - Purpose: 确保性能达标
  - Requirements: 性能要求
  - Test Items:
    - GetBin 响应时间 < 500ms
    - SaveStockReconciliation 响应时间 < 2秒
    - SubmitStockReconciliation 响应时间 < 1秒
  - Tools: grpcurl + time 命令
  - Success: 所有性能指标达标

- [ ] 5.3 更新文档

  - File: 
    - `docs/shared/api/erp_api.md` (如存在)
    - `CHANGELOG.md`
  - Purpose: 确保文档与代码同步
  - Requirements: 文档要求
  - Updates:
    - 在 API 文档中新增 GetBin 接口说明
    - 更新 CHANGELOG.md，记录本次优化
  - Success: 所有文档已更新

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] 测试覆盖率达标
  - Logic 层: ≥ 70%
  - Stock 相关: 100%
- [ ] 所有测试通过

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
  - [x] Requirement 1: 新增 Bin 查询服务
  - [x] Requirement 2: 调整盘点保存逻辑
  - [x] Requirement 3: 盘点提交时验证估值率
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成

### 文档同步

- [ ] API 文档已更新（如有新接口）
- [ ] CHANGELOG.md 已更新
- [ ] design.md 和 tasks.md 已完成

### 规范遵循

- [ ] 遵循 `.cursor/rules/go-bmp.mdc`
- [ ] 遵循 `.cursor/rules/database.mdc`
- [ ] 遵循 `.cursor/rules/security.mdc`
- [ ] Protobuf 遵循 `ttpos-bmp/.cursor/rules/proto-rules.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-stock-optimize-valuation-rate/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-stock-optimize-valuation-rate/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-stock-optimize-valuation-rate/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/story-stock-optimize-valuation-rate/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/story-stock-optimize-valuation-rate/tasks.md)" | bc
```

### 执行流程

1. **选择任务**: 选择下一个未完成任务
2. **阅读需求**: 查看 requirements.md 中的关联需求
3. **查看设计**: 查看 design.md 中的设计方案
4. **查看复用**: 检查 Leverage 中的可复用代码
5. **使用 AI**: 复制 Prompt 模板，让 AI 生成代码
6. **实现代码**: 按照规范实现功能
7. **运行检查**: `go fmt`, `go vet`, `go test`
8. **标记完成**: 将 `[ ]` 改为 `[x]`
9. **提交代码**: Git commit（参考 `.cursor/rules/version.mdc`）

---

## 附录：标准 Prompt 模板

### Go BMP 开发

```
Role: Go Developer specializing in GoFrame and gRPC

Task: {具体任务描述，引用 Requirements}

Context:
- Current file: {文件路径}
- Leverage code: {可复用代码路径}
- Requirements: {需求编号和内容}
- Project specs: 遵循 .cursor/rules/go-bmp.mdc

Restrictions:
- 使用 GoFrame 2.x 框架
- Logic 层只依赖 Service 接口
- 使用 gerror.Wrapf() 包装错误
- 使用 g.Log() 记录日志
- 不使用 panic，返回 error

Success Criteria:
- {成功标准1}
- 代码通过 go fmt 和 go vet
- 测试覆盖率 ≥ 70% (Logic) 或 100% (Stock 相关)
```

### gRPC 开发

```
Role: gRPC Developer with Protobuf expertise

Task: {具体任务描述}

Context:
- Current file: {Protobuf 文件路径}
- Leverage code: {可复用 Protobuf 定义}
- Requirements: {需求编号和内容}
- Project specs: 遵循 ttpos-bmp/.cursor/rules/proto-rules.mdc

Restrictions:
- 使用 proto3 语法
- 遵循命名规范
- 字段编号不冲突
- 使用标准类型

Success Criteria:
- {成功标准1}
- Protobuf 编译成功
- 生成的 Go 代码无错误
```

### 测试工程师

```
Role: QA Engineer with Go testing expertise

Task: {测试任务描述}

Context:
- Target file: {测试目标文件}
- Test file: {测试文件路径}
- Coverage target: ≥ 70% (Logic) 或 100% (Stock 相关)

Test Cases Required:
- 正常场景测试
- 异常场景测试
- 边界条件测试
- 错误处理测试

Restrictions:
- 使用 Mock 依赖
- 遵循 .cursor/rules/go-bmp.mdc
- 必须包含边界情况测试

Success Criteria:
- 测试覆盖率达标
- 所有测试通过
- 边界情况已覆盖
```

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 在执行任务过程中若总结出经验或规避策略，请记录 Episode，并在 tasks.md 尾部更新 `Related Episode`。

---

**模板版本**: v1.0.0  
**创建日期**: 2025-12-23  
**维护者**: rikugun

