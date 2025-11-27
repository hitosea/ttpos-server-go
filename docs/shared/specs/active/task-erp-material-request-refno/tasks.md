# SaveMaterialRequestReq 增加 RefNo 字段 任务分解

> 本文档定义 ttpos-erp stock 模块 SaveMaterialRequestReq 新增 ref_no 字段的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 3  
**已完成**: 3  
**进行中**: -  
**完成率**: 100% ✅

---

## Phase 1: Protobuf 修改

- [x] 1.1 修改 stock.proto 新增 ref_no 字段

  - File: `ttpos-bmp/app/ttpos-erp/manifest/protobuf/stock/stock.proto`
  - Purpose: 在 SaveMaterialRequestReq 消息中新增 ref_no 字段，用于跟踪 ttpos 原始订单号
  - Requirements: 1.1, 1.2, 1.3
  - Leverage: 现有 protobuf 定义
  - Change:
    ```protobuf
    # 在 SaveMaterialRequestReq 消息末尾（items 字段后）新增：
    string ref_no = 10;  // 来源单据号，可选，用于跟踪 ttpos 原始订单号
    ```
  - Success: 字段定义正确，注释清晰

- [x] 1.2 执行 gf gen pb 重新生成 Go 代码

  - File: -
  - Purpose: 根据修改后的 protobuf 文件重新生成 Go 代码
  - Requirements: 1.4
  - Leverage: GoFrame 代码生成工具
  - Command:
    ```bash
    cd ttpos-bmp/app/ttpos-erp && gf gen pb
    ```
  - Success: 
    - `api/stock/stock.pb.go` 包含 `RefNo` 字段
    - `GetRefNo()` 方法已生成

---

## Phase 2: 验证测试

- [x] 2.1 验证字段定义和兼容性

  - File: -
  - Purpose: 验证新字段能正确传递和接收，且向后兼容
  - Requirements: 所有验收标准
  - Leverage: gRPC 客户端测试
  - Test Cases:
    1. 传入 ref_no 时，字段能正确接收
    2. 不传 ref_no 时，接口正常工作，字段为空字符串
  - Success: 所有测试通过

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] protobuf 语法正确
- [ ] 生成的 Go 代码无错误

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成

### 规范遵循

- [ ] 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`
- [ ] 遵循 `ttpos-bmp/.cursor/rules/proto-rules.mdc`

---

## 进度追踪

### 执行流程

1. **选择任务**: 选择下一个未完成任务
2. **执行修改**: 修改 protobuf 文件
3. **生成代码**: 执行 `gf gen pb`
4. **验证测试**: 验证功能正常
5. **标记完成**: 将 `[ ]` 改为 `[x]`
6. **提交代码**: Git commit（参考 `.cursor/rules/version.mdc`）

### 预计完成时间

- **Phase 1**: 15 分钟
- **Phase 2**: 15 分钟
- **总计**: 0.5 天

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/rikugun/2025-11/2025-11-27.md`

---

**模板版本**: v1.0.0  
**最后更新**: 2025-11-27  
**维护者**: rikugun

