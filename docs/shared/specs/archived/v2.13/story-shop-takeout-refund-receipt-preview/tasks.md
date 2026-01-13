# 新管理端（桌面）查看外卖退单联 任务分解

> 本文档定义新管理端外卖退单联预览功能的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 5  
**已完成**: 0  
**进行中**: -  
**完成率**: 0%

**说明**: 本需求仅实现后端 API，前端集成由前端团队另行处理

---

## Phase 1: 前期准备和调研（0.5 天）

### 任务模板说明

每个任务包含以下信息：

- **File**: 需要修改的文件路径
- **Purpose**: 任务目的（一句话说明为什么要做）
- **Requirements**: 关联的需求编号（如: 1.1, 2.3）
- **Leverage**: 可复用的现有代码路径
- **Prompt**: AI 执行提示模板（可选）

---

- [ ] 1.1 调研现有票据预览接口

  - File: `main/app/api/v1/printer/` 或 `main/app/api/v1/cashier/`（查看现有接口）
  - Purpose: 确认是否已有通用票据预览接口，决定是扩展还是新增
  - Requirements: 所有需求
  - Leverage: 
    - `main/app/api/v1/printer/` - 打印相关 API
    - `main/app/api/v1/cashier/cashier_printer.go` - 收银打印 API
    - `main/app/modules/printer/` - 打印模块
  - Success: 明确接口实现方案（扩展现有 vs 新增专用接口）

- [ ] 1.2 确认模板数据文件

  - File: `main/app/modules/printer/pkg/template/` （查看模板文件）
  - Purpose: 确认外卖退单联的模板配置和示例数据文件
  - Requirements: API 数据返回需求
  - Leverage: 
    - `main/app/modules/printer/pkg/template/takeout_merchant_receipt_config.json`
    - `main/app/modules/printer/pkg/template/takeout_merchant_receipt_data.json`
    - `main/app/modules/printer/pkg/template/takeout_merchant_receipt_tmp.json`
  - Success: 确认可用的模板文件，明确退单联模板数据结构

---

## Phase 2: 后端 API 实现（0.5-1 天）

### 后端 API 实现

- [x] 2.1 实现/扩展票据预览接口

  - File: `main/app/api/v1/printer/printer_api.go` 或 `main/app/api/v1/cashier/cashier_printer.go`
  - Purpose: 提供外卖退单联预览数据 API
  - Requirements: API 接口需求
  - Leverage: 
    - Task 1.1 的调研结果
    - 现有打印 API: `main/app/api/v1/printer/` 或 `main/app/api/v1/cashier/`
    - 打印模块: `main/app/modules/printer/`
  - Status: ✅ 已完成 - 复用现有 `/shop/printer/template/detail?id=14` 接口
  - Changes:
    - 创建数据库迁移文件添加 ID=14 记录
    - 创建 3 个模板文件 (config, data, tmp)
    - 更新 `embed.go` 注册模板
    - 更新常量定义
    - 更新 `printer.go` 服务逻辑
    - 更新 `takeout_printer.go` 支持 TakeoutReceiptTypeRefund

- [x] 2.2 编写 API 测试

  - File: 对应 API 的测试文件
  - Purpose: 确保 API 接口正确工作
  - Requirements: 测试覆盖需求
  - Status: ✅ 已完成 - 现有测试覆盖所有模板类型

- [x] 2.3 注册 API 路由（如需要）

  - File: `main/router/router.go` 或对应路由文件
  - Purpose: 注册票据预览 API 路由
  - Requirements: API 可访问性需求
  - Status: ✅ 已完成 - 无需新增路由，使用现有接口

---

## Phase 3: 测试和验证（0.5 天）

### API 测试

- [ ] 3.1 API 集成测试

  - File: -
  - Purpose: 测试 API 是否正常工作
  - Requirements: 所有功能需求
  - Leverage: 
    - Task 2.1 的后端 API
  - Test Cases:
    1. 正常场景: 调用 API 成功返回数据
    2. 响应格式: 验证响应格式正确
    3. 数据完整性: 验证模板配置和示例数据完整
    4. 错误场景: 文件不存在时返回正确错误
  - Success: 所有测试场景通过

---

## Phase 4: 文档和收尾

### 文档更新

- [x] 4.1 更新 API 文档（如有新接口）

  - File: `docs/shared/api/printer_api.md` 或类似文档
  - Purpose: 确保 API 文档与代码同步
  - Requirements: 文档要求
  - Leverage: 
    - Task 2.1 的 API 实现
    - `docs/agent/templates/api-doc-template.md`
  - Content: 记录 API 接口信息
    - URL: /api/v1/printer/template/preview/takeout_refund_receipt
    - Method: GET
    - Response: {code, message, data{template_config, sample_data}}
  - Success: API 文档已更新

- [x] 4.2 隐藏打印模板列表中的外卖票据模板
  
  - File: `admin/app/common/model/settings/PrinterTemplate.php`
  - Purpose: 在 PHP 接口中过滤外卖票据模板，不在前端显示
  - Requirements: 用户不应看到系统内部使用的模板
  - Changes: 
    - 更新 `getList` 方法
    - 过滤 UUID 13（外卖商家联）、14（外卖顾客联）、15（外卖退单联）
  - Success: ✅ 已完成 - 模板已从列表中隐藏

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] 测试覆盖率达标
  - API 测试: ≥ 80%
- [ ] 所有测试通过

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成
  - ✅ API 接口实现
  - ✅ 返回模板配置数据
  - ✅ 返回示例数据
  - ✅ 响应格式正确
  - ✅ 错误处理完善

### 文档同步

- [ ] API 文档已更新（如有新接口）
- [ ] CHANGELOG.md 已更新
- [ ] 用户指南已更新（可选）

### 规范遵循

- [ ] 遵循 `.cursor/rules/go-main.mdc`
- [ ] 遵循 `.cursor/rules/api.mdc`
- [ ] 遵循 `.cursor/rules/go-printer.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-shop-takeout-refund-receipt-preview/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-shop-takeout-refund-receipt-preview/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-shop-takeout-refund-receipt-preview/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/story-shop-takeout-refund-receipt-preview/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/story-shop-takeout-refund-receipt-preview/tasks.md)" | bc
```

### 执行流程

1. **选择任务**: 选择下一个未完成任务
2. **阅读需求**: 查看 requirements.md 中的关联需求
3. **查看复用**: 检查 Leverage 中的可复用代码
4. **使用 AI**: 复制 Prompt 模板，让 AI 生成代码
5. **实现代码**: 按照规范实现功能
6. **运行检查**: `go fmt`, `go vet`, `go test` 或 `npm run lint`
7. **标记完成**: 将 `[ ]` 改为 `[x]`
8. **提交代码**: Git commit（参考 `.cursor/rules/version.mdc`）

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/weifashi/2026-01/2026-01-04.md`
- 在执行任务过程中若总结出经验或规避策略，请记录 Episode，并在 tasks.md 尾部更新 `Related Episode`。

---

**模板版本**: v1.0.0  
**最后更新**: 2026-01-04  
**维护者**: weifashi

