# UpdateProduct 增加 UOM 字段支持 任务分解

> 本文档定义 ERP UpdateProduct 接口增加 UOM 字段更新支持的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 4  
**已完成**: 4  
**进行中**: -  
**完成率**: 100%

---

## Phase 1: Protobuf 定义

- [x] 1.1 修改 UpdateProductReq 消息

  - File: `ttpos-bmp/app/ttpos-erp/manifest/protobuf/item/product.proto`
  - Purpose: 在请求消息中增加 stock_uom 字段
  - Requirements: 1.1
  - Leverage: 现有字段定义格式
  - Change:
    ```protobuf
    message UpdateProductReq {
      // ... 现有字段
      string stock_uom = 6; // 库存单位，可选
    }
    ```

- [x] 1.2 修改 UpdateProductResp 消息

  - File: `ttpos-bmp/app/ttpos-erp/manifest/protobuf/item/product.proto`
  - Purpose: 在响应消息中增加 stock_uom 字段
  - Requirements: 1.2
  - Leverage: 现有字段定义格式
  - Change:
    ```protobuf
    message UpdateProductResp {
      // ... 现有字段
      string stock_uom = 6; // 库存单位
    }
    ```

- [x] 1.3 重新生成 gRPC 代码

  - File: `ttpos-bmp/app/ttpos-erp/api/item/product.pb.go`, `product_grpc.pb.go`
  - Purpose: 生成包含新字段的 Go 代码
  - Requirements: 1.3
  - Command: `cd ttpos-bmp/app/ttpos-erp && gf gen pb`
  - Success: 生成的 `UpdateProductReq` 和 `UpdateProductResp` 结构体包含 `StockUom` 字段

---

## Phase 2: 业务逻辑实现

- [x] 2.1 修改 UpdateProduct Logic

  - File: `ttpos-bmp/app/ttpos-erp/internal/logic/stock/product.go`
  - Purpose: 处理 stock_uom 字段的更新逻辑
  - Requirements: 2.1, 2.2, 2.3
  - Leverage: 现有 `UpdateProduct` 方法实现
  - Change:
    ```go
    func (s *sProduct) UpdateProduct(ctx context.Context, req *item.UpdateProductReq) (*item.UpdateProductResp, error) {
        itemInfo := &erp.Item{
            CustomNotForSale: req.NotForSale,
            Disabled:         req.Disabled,
        }

        // 新增：处理 stock_uom
        if len(req.StockUom) > 0 {
            itemInfo.StockUom = req.StockUom
        }

        // ... 其余现有逻辑不变

        return &item.UpdateProductResp{
            ItemCode:     req.ItemCode,
            NotForSale:   req.NotForSale,
            InternalCode: req.InternalCode,
            Disabled:     req.Disabled,
            StockUom:     req.StockUom, // 新增
        }, nil
    }
    ```
  - Prompt: Role: Go Developer with GoFrame expertise | Task: 修改 sProduct.UpdateProduct 方法，增加 stock_uom 字段处理 | Context: 当 req.StockUom 非空时设置 itemInfo.StockUom，并在响应中返回 | Restrictions: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc | Success: stock_uom 字段能正确更新到 ERPNext

---

## Phase 3: 测试验证

- [ ] 3.1 手动测试接口

  - File: -
  - Purpose: 验证接口功能正确
  - Requirements: 所有功能需求
  - Test Cases:
    1. 传入有效 `stock_uom` 值（如 "个"），验证 ERPNext 中商品单位已更新
    2. 不传入 `stock_uom`，验证其他字段正常更新，单位不变
    3. 传入空字符串 `stock_uom: ""`，验证单位不变
  - Success: 所有测试场景通过

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] Protobuf 注释完整

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成

### 文档同步

- [ ] Protobuf 字段有中文注释

### 规范遵循

- [ ] 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`
- [ ] 遵循 `ttpos-bmp/.cursor/rules/proto-rules.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/task-erp-update-product-uom/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/task-erp-update-product-uom/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/task-erp-update-product-uom/tasks.md
```

### 执行流程

1. **选择任务**: 按顺序执行 1.1 → 1.2 → 1.3 → 2.1 → 3.1
2. **阅读需求**: 查看 requirements.md 中的关联需求
3. **实现代码**: 按照 Change 部分修改代码
4. **运行检查**: `go fmt`, `go vet`
5. **标记完成**: 将 `[ ]` 改为 `[x]`
6. **提交代码**: Git commit

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 活动日志：`docs/team/activities/rikugun/2025-11/2025-11-27.md`

---

**模板版本**: v1.0.0  
**最后更新**: 2025-11-27  
**维护者**: rikugun

