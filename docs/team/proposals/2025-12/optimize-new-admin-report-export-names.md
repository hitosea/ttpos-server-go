# 优化新管理端导出报表名称 需求提案

> 本文档用于需求发起阶段，经团队评审后创建正式 Spec。

---

## 📋 提案信息

| 项目       | 内容     |
| ---------- | -------- |
| **提案人** | 王昱     |
| **日期**   | 2025-12-01 |
| **目标版本** | - |
| **状态**   | 已创建 Spec   |
| **关联任务** | - |
| **关联 Spec** | [story-shop-report-export-optimize](../../../shared/specs/active/story-shop-report-export-optimize/requirements.md)      |

---

## 🎯 背景和动机

### 问题描述

当前新管理端导出报表时，文件名使用时间戳格式（如 `时段营业统计_1733012345.xlsx`），存在以下问题：

1. **文件名不直观**：时间戳格式不便于用户识别导出日期，需要打开文件或查看文件属性才能知道导出时间
2. **同一天多次导出时文件名冲突**：使用时间戳虽然避免了冲突，但文件名不统一，难以管理
3. **子表名称不统一**：当前使用多语言的"报表"/"Report"等名称，不够标准化

**示例场景**：
> 商户管理员在同一天多次导出"时段营业统计"报表时，文件名分别为：
> - `时段营业统计_1733012345.xlsx`
> - `时段营业统计_1733015678.xlsx`
> - `时段营业统计_1733019012.xlsx`
> 
> 用户无法直观看出这些文件是同一天导出的，也不便于按日期整理文件。

### 业务价值

解决这个问题能带来什么业务价值？

- **提升用户体验**：文件名包含日期，用户一眼就能识别导出时间
- **便于文件管理**：同一天导出的报表文件名统一，便于按日期分类整理
- **标准化命名**：统一的命名规则，提升系统专业性
- **减少用户困惑**：避免时间戳格式带来的理解成本

### 目标用户

谁会使用这个功能？

- [x] 商户管理员
- [ ] 收银员
- [ ] 厨房人员
- [ ] 顾客
- [x] 其他: 财务人员、区域经理

---

## 💡 解决方案概述

### 方案描述

优化新管理端报表导出功能的文件命名和子表命名规则：

1. **文件名优化**：
   - 格式：`报表名YYYY-MM-DD.xlsx`（如：`时段营业统计2025-10-10.xlsx`）
   - 同一天多次导出同名报表时，自动添加序号：`报表名YYYY-MM-DD（1）.xlsx`、`报表名YYYY-MM-DD（2）.xlsx`，以此类推
   - 日期使用商户时区，而非服务器时区

2. **子表名称优化**：
   - 统一将子表名称从多语言的"报表"/"Report"等改为标准的 `Sheet1`
   - **例外**：用户分析报表保持原有子表名称逻辑不变（因为可能有多个子表）

### 核心功能点

1. **文件名格式优化**：从 `报表名_时间戳.xlsx` 改为 `报表名YYYY-MM-DD.xlsx`
2. **同名文件自动编号**：同一天导出同名报表时，自动添加序号避免冲突
3. **子表名称标准化**：统一改为 `Sheet1`（用户分析除外）
4. **时区处理**：文件名中的日期使用商户时区

### 影响范围

**涉及终端**：
- [ ] POS 收银端
- [x] Shop 商家管理端（新管理端）
- [ ] KDS 厨显端
- [ ] QDS 排号叫号端
- [ ] Assistant 助手端
- [ ] Tablet 平板端
- [ ] Mobile 扫码端
- [ ] Menu 电子菜单端
- [ ] Member 会员端

**涉及模块**：
- [ ] UI 组件
- [ ] API 接口
- [ ] 数据模型
- [x] 业务逻辑（导出服务）
- [ ] 第三方集成
- [ ] 其他: Excel 文件生成逻辑

**涉及的报表类型**：
1. 时段营业统计（`ExportTypeBusinessData`）
2. 综合运营统计（`ExportTypeBusinessDataSummary`）
3. 营业收款统计（`ExportTypeBusinessDataPaymentMethod`）
4. 渠道营业统计（`ExportTypeChannelSales`）
5. 商品销售统计（`ExportTypeProductSales`）
6. 用户分析（`ExportTypeUserAnalysis`）- 仅优化文件名，不修改子表名称
7. 后厨菜品出品明细（`ExportTypeKitchenProductionDetail`）
8. 后厨效率分析（`ExportTypeKitchenEfficiencyAnalysis`）

---

## 📊 初步评估

### 技术复杂度

- [ ] **低**：纯 UI 调整，无业务逻辑变更
- [x] **中**：需要前后端联调，基础业务逻辑
- [ ] **高**：涉及架构调整、第三方集成、复杂算法

**说明**：
- 需要修改导出服务的文件名生成逻辑
- 需要查询同一天已导出的同名报表，实现自动编号
- 需要修改 Excel 子表名称设置
- 涉及时区处理，需要确保日期计算正确

### 工作量预估

[粗略估算，具体 SP 在技术方案评审时确定]

- **预计天数**: 2-3 天
- **预估 SP**: 3-5 SP（待技术评审确认）

**分解**：
- 文件名生成逻辑优化：1 天
- 同名文件检测和编号逻辑：0.5 天
- 子表名称修改：0.5 天
- 测试和验证：1 天

### 风险识别

**潜在风险**：
1. **时区处理错误**：文件名中的日期可能因时区计算错误导致不准确
2. **并发导出冲突**：同一天多次快速导出时，可能出现文件名冲突
3. **历史数据兼容**：不影响历史导出记录，但需要考虑导出记录表的查询逻辑

**缓解措施**：
1. **时区处理**：使用商户时区工具类 `utils.SetTimezone()`，确保日期计算正确
2. **并发控制**：在创建导出记录时使用数据库事务和唯一索引，确保文件名唯一性
3. **历史兼容**：仅影响新导出的文件，历史记录不受影响

---

## 🔗 相关资源

### 参考需求

- 类似功能: 无
- 竞品分析: 无

### 相关文档

- 导出功能实现: `main/app/service/business.go`
- 导出记录模型: `main/app/model/export_record.go`
- 相关 API: `main/app/api/v1/shop/shop_statistics.go`

### 代码位置

**文件名生成相关代码**：
- `main/app/service/business.go:2263` - 时段营业统计导出
- `main/app/service/business.go:2467` - 综合运营统计导出
- `main/app/service/business.go:2676` - 营业收款统计导出
- `main/app/service/business.go:2953` - 渠道营业统计导出
- `main/app/service/business.go:789` - 商品销售统计导出
- `main/app/service/business.go:3554` - 用户分析导出
- `main/app/service/business.go:1522` - 后厨效率分析导出
- `main/app/service/business.go:1652` - 后厨菜品出品明细导出

**子表名称相关代码**：
- `main/app/service/business.go:908-920` - 商品销售统计子表名称
- `main/app/service/business.go:2017-2030` - 后厨菜品出品明细子表名称
- `main/app/service/business.go:2150-2165` - 后厨效率分析子表名称
- 其他报表的子表名称设置位置待确认

---

## 🤝 需求评审

### 评审参与人

| 角色         | 姓名   | 签名/日期 |
| ------------ | ------ | --------- |
| 产品经理     | {姓名} |           |
| 技术负责人   | {姓名} |           |
| 开发代表     | {姓名} |           |
| 测试代表     | {姓名} |           |
| UI/UX 设计师 | {姓名} |           |

### 评审结论

- [ ] ✅ **批准**：进入技术方案设计阶段
- [ ] 🔄 **修改后批准**：需补充以下内容
- [ ] ❌ **拒绝**：不符合产品规划或优先级

**评审意见**：

```
[记录评审会议的关键讨论和决策]
```

**下一步行动**：

- [x] 创建 Spec：`story-shop-report-export-optimize` ✅
- [ ] 产品审核：待审核
- [ ] 分配负责人：{姓名}
- [ ] 目标 Sprint：Sprint {N}

---

## 📝 附录

### User Story（初稿）

**作为** 商户管理员  
**我想** 导出报表时文件名包含日期，且同一天导出的同名报表自动编号  
**以便于** 快速识别导出时间，便于文件管理和整理

### AC 验收标准（初稿）

1. **WHEN** 用户导出报表 **THEN** 文件名格式为 `报表名YYYY-MM-DD.xlsx` **SHALL** 日期使用商户时区
2. **WHEN** 同一天多次导出同名报表 **THEN** 文件名自动添加序号 `报表名YYYY-MM-DD（1）.xlsx`、`报表名YYYY-MM-DD（2）.xlsx` **SHALL** 序号从1开始递增
3. **WHEN** 导出报表（除用户分析外） **THEN** Excel 子表名称 **SHALL** 为 `Sheet1`
4. **WHEN** 导出用户分析报表 **THEN** Excel 子表名称 **SHALL** 保持原有逻辑不变
5. **IF** 文件名已存在 **THEN** 系统 **SHALL** 自动检测并添加序号，避免覆盖

### 实现细节

#### 文件名生成逻辑

```go
// 伪代码示例
func generateFileName(reportName string, exportType uint8, ctx context.Context) string {
    // 1. 获取商户时区
    timezone := ctx.GetCompanySetting().Timezone
    timezoneUtils := utils.SetTimezone(timezone)
    dateString := timezoneUtils.FormatUnixTime(time.Now().Unix(), "2006-01-02")
    
    // 2. 生成基础文件名
    baseFileName := fmt.Sprintf("%s%s.xlsx", reportName, dateString)
    
    // 3. 查询同一天已导出的同名报表
    db := ctx.GetDB()
    existingFiles := repository.NewExportRecordRepo(db).GetByDateAndType(
        dateString, 
        exportType,
        ctx.GetCompanyUuid(),
    )
    
    // 4. 计算序号
    suffix := ""
    if len(existingFiles) > 0 {
        suffix = fmt.Sprintf("（%d）", len(existingFiles))
    }
    
    return fmt.Sprintf("%s%s%s", reportName, dateString, suffix)
}
```

#### 子表名称修改

```go
// 修改前
sheetNameMul := model.MultiLanguageName{
    EnName:   "Report",
    ZhName:   "报表",
    // ...
}
sheetName := sheetNameMul.GetNameByLang(ctx.GetLanguage())

// 修改后（除用户分析外）
sheetName := "Sheet1"
```

---

## 📄 模板使用说明

### 何时使用此模板

- ✅ 产品经理提出新功能想法
- ✅ 用户反馈需求建议
- ✅ 技术团队提出改进方案
- ✅ 需要团队讨论和评审的需求

### 与 Spec 的区别

| 阶段        | 文档类型      | 详细程度 | 用途               |
| ----------- | ------------- | -------- | ------------------ |
| **需求发起** | Proposal      | 粗略     | 团队评审、决策是否做 |
| **需求确认** | Requirements  | 详细     | User Story + AC，开发依据 |
| **技术设计** | Design        | 详细     | 技术方案，实现指导 |
| **任务分解** | Tasks         | 详细     | 开发执行，进度追踪 |

### 流转路径

```
提案 (Proposal) 
  ↓ 评审批准
需求文档 (Requirements) 
  ↓ 技术评审
设计文档 (Design) 
  ↓ SP 评估 ≤ 5
任务分解 (Tasks)
  ↓
开发实现
```

---

**版本**: v1.0.0  
**创建日期**: 2025-12-01  
**维护者**: 产品组 + Scrum Master  
**相关规范**: `.cursor/rules/scrum_story_point.mdc`, `.cursor/rules/specs.mdc`

