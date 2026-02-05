# 报损单管理 需求文档

## 基本信息

| 项目              | 内容                                                                 |
| ----------------- | -------------------------------------------------------------------- |
| **Spec ID**       | story-shop-stock-loss                                                |
| **来源 Proposal** | [报废管理功能需求提案](../../../team/proposals/2026-01/2026-01-05-scrap-management.md) |
| **DooTask 任务**  | [#39057](http://t.hitosea.com/manage/project/368/task/39057)         |
| **创建日期**      | 2026-02-02                                                           |
| **负责人**        | 曾振华                                                               |
| **目标版本**      | v2.17 (2月6日)                                                       |

## 审核状态

| 项目         | 内容       |
| ------------ | ---------- |
| **审核状态** | 开发中     |
| **审核人**   | -          |
| **审核日期** | 2026-02-02 |

---

## 用户故事

### 用户故事 1: 仓库管理员

**作为** 仓库管理员
**我想** 能够通过系统管理报损单的完整流程
**以便于** 及时处理损坏、过期、变质等无法使用的物品，规范报废流程，确保库存准确性

### 用户故事 2: 店长

**作为** 店长
**我想** 能够审核和管理报废单
**以便于** 控制报废流程，确保报废操作的合理性和准确性

---

## 功能需求

### Requirement 1: 报损单 CRUD

**用户故事**: 作为仓库管理员，我想创建、编辑、保存和删除报损单，以便于管理报损物品

#### 验收标准

1. **WHEN** 用户点击「新建报损单」**THEN** 系统 **SHALL** 显示报损单创建表单，包含：报损类型、报损仓库、报损原因、物品列表、附件
2. **WHEN** 用户填写完报损单信息并点击「保存」**THEN** 系统 **SHALL** 生成单据编号（格式：SL + 时间戳 + 序列号，如 SL202504030915120001），状态设为「已保存」(status=0)
3. **WHEN** 用户编辑已保存的报损单并点击「保存」**THEN** 系统 **SHALL** 更新单据日期为当前时间，保持状态为「已保存」
4. **WHEN** 用户删除状态为「已保存」的报损单 **THEN** 系统 **SHALL** 执行软删除（设置 delete_time）
5. **IF** 报损单状态不是「已保存」(status!=0) **THEN** 系统 **SHALL** 禁止删除操作，提示「已提交/审核的报损单不允许删除」

### Requirement 2: 提交与审核流程

**用户故事**: 作为店长，我想审核报损单，以便于控制报废流程的合规性

#### 库存校验规则

- **单位换算**：通过 `material_unit_uuid` 查询 `ttpos_material_unit.conversion_rate` 获取换算系数
- **基准单位**：`ttpos_material_unit.is_default=1` 表示基准单位
- **汇总校验**：同一物料多行报损时，需按基本单位汇总后再校验
- **校验公式**：`报损数量(基本单位) = Σ(报损数量 × conversion_rate)`，库存充足条件：`当前库存(基本单位) >= 报损数量(基本单位)`

#### 验收标准

1. **WHEN** 用户点击「提交」**THEN** 系统 **SHALL** 将所有报损物料按基本单位换算并按物料汇总，校验库存是否充足（仅校验不扣减），更新状态为「已提交」(status=1)，记录 submit_time
2. **IF** 同一物料存在多行（不同单位）**THEN** 系统 **SHALL** 先汇总为基本单位数量，再与库存比较
3. **IF** 物料库存不足 **THEN** 系统 **SHALL** 提示「物料{name}库存不足，需要 {需求数量}{基本单位}，当前库存 {当前库存}{基本单位}」，阻止提交
4. **WHEN** 审核人点击「通过」并确认 **THEN** 系统 **SHALL** 再次按相同规则校验库存，先调用 ERP 创建 Stock Entry，成功后更新 erp_code 并扣减 TTPOS 侧库存，更新状态为「已审核通过」(status=2)，记录 approve_time
5. **WHEN** 审核人点击「驳回」并填写驳回原因 **THEN** 系统 **SHALL** 更新状态为「已驳回」(status=3)，记录 reject_time，驳回原因保存到批注表
6. **IF** 驳回原因未填写 **THEN** 系统 **SHALL** 提示「请填写驳回原因」，阻止驳回操作
7. **WHEN** 报损单被驳回后用户点击「重新提交」**THEN** 系统 **SHALL** 重新按相同规则校验库存，更新状态为「已提交」(status=1)，记录新的 submit_time

### Requirement 3: ERP 同步

**用户故事**: 作为系统，我需要在审核通过后同步到 ERP 系统，以便于保持数据一致性

#### 验收标准

1. **WHEN** 报损单审核通过 **THEN** 系统 **SHALL** 通过 ttpos-bmp gRPC 接口调用 ERP API，创建 Stock Entry 单据（Type: Material Issue, docstatus=1）
2. **WHEN** ERP 同步成功 **THEN** 系统 **SHALL** 记录 erp_code（ERP 返回的单据编号），更新状态为「已审核通过」
3. **IF** ERP 同步失败 **THEN** 系统 **SHALL** 直接返回错误提示，状态保持「已提交」(status=1)，不扣减库存

### Requirement 4: 库存扣减

**用户故事**: 作为系统，我需要在审核通过后扣减库存，以便于保持库存准确性

#### 验收标准

1. **WHEN** 报损单审核通过且 ERP 同步成功 **THEN** 系统 **SHALL** 扣减 TTPOS 侧库存（先 ERP 后库存策略）
2. **WHEN** ERP Stock Entry 创建成功 **THEN** ERP 侧 **SHALL** 自动扣减库存（Stock Entry 提交时自动处理）
3. **IF** ERP 同步失败 **THEN** 系统 **SHALL** 不扣减 TTPOS 侧库存，返回同步失败错误
4. **WHEN** 报损出库完成 **THEN** 出入库明细表 **SHALL** 记录类型为「报损出库」，对方机构为「-」，单据号为 TTPOS 侧单据编号

### Requirement 5: 报损单列表与筛选

**用户故事**: 作为仓库管理员，我想查看和筛选报损单列表，以便于快速找到目标单据

#### 验收标准

1. **WHEN** 用户进入报损单列表页 **THEN** 系统 **SHALL** 显示状态 Tab：全部、待提交、待审核、已驳回、已完成
2. **WHEN** 用户搜索 **THEN** 系统 **SHALL** 支持按单据编号、报损单号（ERP 单据编号）搜索
3. **WHEN** 用户筛选日期 **THEN** 系统 **SHALL** 默认显示近 30 天，支持快速筛选（今天、近7天、上月/本月、今年）和任意日期范围
4. **WHEN** 显示列表 **THEN** 系统 **SHALL** 按日期倒序排序

### Requirement 6: 批注与操作记录

**用户故事**: 作为店长，我想查看报损单的操作记录，以便于审计追溯

#### 验收标准

1. **WHEN** 报损单被提交 **THEN** 系统 **SHALL** 记录批注：提交时间、报损原因
2. **WHEN** 报损单被驳回 **THEN** 系统 **SHALL** 记录批注：驳回时间、驳回原因
3. **WHEN** 报损单被审核通过 **THEN** 系统 **SHALL** 记录批注：完成时间、完成批注（若填写）
4. **WHEN** 用户点击「批注」**THEN** 系统 **SHALL** 显示批注详情页，按时间顺序展示所有操作记录

### Requirement 7: 功能权限

**用户故事**: 作为系统管理员，我想控制报损管理功能的访问权限

#### 验收标准

1. **WHEN** 新增报损管理功能 **THEN** 系统 **SHALL** 在功能权限列表增加「报损管理」选项
2. **WHEN** 现有商户升级 **THEN** 系统 **SHALL** 为原有角色默认勾选「报损管理」权限
3. **WHEN** 新店创建 **THEN** 系统 **SHALL** 为超管和收银员角色默认勾选「报损管理」权限

---

## 非功能需求

### 测试要求

- [ ] 单元测试覆盖率 >= 80%
- [ ] Service 层测试覆盖核心业务逻辑
- [ ] ERP 同步失败回滚场景测试

### 平台兼容性

- [ ] Shop 端（商家管理端）Web 应用
- [ ] 支持主流浏览器（Chrome、Firefox、Safari、Edge）

### 国际化要求

- [ ] 所有提示文案支持多语言
- [ ] 使用 `dto.LocaleResponse` 处理多语言字段

---

## 约束条件

### 技术约束

- Go 版本: 1.23+
- 框架: Gin + GORM（Main 模块）
- 分层架构: API -> Service -> Repository -> Model
- 必须遵循 CLAUDE.md 和 .cursor/rules/go-main.mdc 规范
- ERP 同步通过 ttpos-bmp gRPC 接口调用
- 事务使用 `repository.CommonRepo.Transaction`
- 日志必须包含 `company_uuid` 字段

### 资源约束

- Story Point: 5 (预估)

---

## 数据模型

### 报损单主表 (ttpos_stock_loss)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | bigint | 主键 |
| uuid | bigint | UUID |
| code | varchar(50) | 单据编号 (SL + 时间戳 + 序列号) |
| erp_code | varchar(50) | ERP 单据编号 |
| loss_type | int | 报损类型 (1:物品损坏 2:物品报废 3:物品过期) |
| warehouse_uuid | bigint | 报损仓库 UUID |
| reason | text | 报损原因 |
| status | int | 状态 (0:已保存 1:已提交 2:已审核通过 3:已驳回) |
| submit_time | int | 提交时间 |
| approve_time | int | 审核通过时间 |
| reject_time | int | 驳回时间 |
| submitter_uuid | bigint | 提交人 UUID |
| approver_uuid | bigint | 审核人 UUID |
| create_time | int | 创建时间 |
| update_time | int | 更新时间 |
| delete_time | int | 删除时间 |

### 报损单明细表 (ttpos_stock_loss_item)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | bigint | 主键 |
| uuid | bigint | UUID |
| stock_loss_uuid | bigint | 报损单 UUID |
| material_uuid | bigint | 物料 UUID |
| material_name | text | 物料名称 |
| material_unit_uuid | bigint | 物料单位 UUID |
| material_unit_name | text | 物料单位名称 |
| quantity | decimal(14,4) | 报损数量 |
| create_time | int | 创建时间 |
| update_time | int | 更新时间 |
| delete_time | int | 删除时间 |

### 报损单附件关联表 (ttpos_stock_loss_file)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | bigint | 主键 |
| uuid | bigint | UUID |
| stock_loss_uuid | bigint | 报损单 UUID |
| file_uuid | bigint | 文件 UUID (关联 ttpos_file) |
| sort_order | int | 排序顺序 |
| create_time | int | 创建时间 |
| update_time | int | 更新时间 |
| delete_time | int | 删除时间 |

### 报损单批注表 (ttpos_stock_loss_annotation)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | bigint | 主键 |
| uuid | bigint | UUID |
| stock_loss_uuid | bigint | 报损单 UUID |
| action | varchar(20) | 操作类型 (submit/resubmit/approve/reject) |
| content | text | 批注内容（报损原因/驳回原因/完成批注） |
| operator_uuid | bigint | 操作人 UUID |
| operator_name | varchar(100) | 操作人姓名 |
| create_time | int | 创建时间 |
| update_time | int | 更新时间 |
| delete_time | int | 删除时间 |

---

## 风险和缓解

### 风险 1: ERP 同步失败导致数据不一致

**影响**: 低
**缓解措施**:
- 采用先 ERP 后库存策略（ERP 成功后再扣减 TTPOS 库存）
- ERP 调用失败时直接返回错误，无需回滚（库存尚未扣减）

### 风险 2: 审核期间库存被其他操作消耗

**影响**: 中
**缓解措施**:
- 审核通过时再次校验库存
- 使用数据库事务保证原子性

---

**版本**: v1.0.0
**创建日期**: 2026-02-02
